package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"note/cfg"
	"note/shell"
	"os"
	"sort"
)

var DefaultBranch = "main"

var HelpStr = "使用方法:" +
	"\n	note add fileName/number // 新增/编辑文件,举例 note add ReadMe 或者 note 1" +
	"\n	note addDir dirName // 新增目录, 支持多级目录" +
	"\n	note list/l // 列出存储目录结构" +
	"\n	note view/v fileName // 查看文件内容" +
	"\n	note s <keyWord> // 搜索关键字" +
	"\n	note move srcPath targetPath //也支持重命名 note move java/a.go golang/b.go" +
	"\n	note init // 初始化仓库" +
	"\n	note push // 推送到github仓库" +
	"\n	note rm fileName // 删除目录/文件" +
	"\n	note log // 查看仓库提交日志" +
	""

type GitHubClient struct {
	LocalPath  string // 本地仓库路径
	RemoteURL  string // GitHub仓库地址
	SSHKeyPath string // SSH私钥路径
	repo       *git.Repository
	//auth       *ssh.PublicKeys
	auth *http.BasicAuth
}

func NewClient(localPath, remoteURL, sshKeyPath string) (*GitHubClient, error) {
	auth := &http.BasicAuth{
		Username: cfg.DefaultCfg.Git.User,
		Password: cfg.DefaultCfg.Git.Password,
	}

	// 初始化客户端
	c := &GitHubClient{
		LocalPath:  localPath,
		RemoteURL:  remoteURL,
		SSHKeyPath: sshKeyPath,
		auth:       auth,
	}

	// 初始化/打开仓库
	err := c.initRepo()
	if err != nil {
		return nil, err
	}

	name, err := getCurrentBranch(c.repo)
	DefaultBranch = name

	return c, nil
}

func (c *GitHubClient) initEmptyRepo() error {
	// 初始化/打开仓库
	r, err := git.PlainInit(c.LocalPath, false)
	if err != nil {
		return fmt.Errorf("仓库初始化失败: %v", err)
	}
	c.repo = r
	return nil
}

func (c *GitHubClient) initRepo() error {
	repo, err := git.PlainOpen(c.LocalPath)
	if err == git.ErrRepositoryNotExists {
		// 克隆仓库
		repo, err = git.PlainClone(c.LocalPath, false, &git.CloneOptions{
			URL:          c.RemoteURL,
			Auth:         c.auth,
			SingleBranch: true,
		})
	}

	if err != nil && err.Error() == "remote repository is empty" {
		err = c.initEmptyRepo()
		if err != nil {
			fmt.Println("仓库初始化失败", err)
			return err
		} else {

			err = createFile(c.LocalPath+"readme", HelpStr)
			if err != nil {
				fmt.Println("create file failed:", err)
			}
			err = c.CommitChanges("init")
			if err != nil {
				fmt.Println("commit failed:", err)
			}
			err = addRemote(c.repo, "origin", c.RemoteURL)
			if err != nil {
				fmt.Println("add remote failed:", err)
			}
			//c.pushChanges()
			//_, err = c.Sync()
			name, err := getCurrentBranch(c.repo)
			DefaultBranch = name

			err = c.pushU()
			if err != nil {
				fmt.Println("sync failed:", err)
			}
			fmt.Println("仓库初始化成功!!")
			return err
		}
	}

	c.repo = repo
	return nil
}

func createFile(path, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// 关联远程仓库
func addRemote(repo *git.Repository, name, url string) error {
	_, err := repo.CreateRemote(&config.RemoteConfig{
		Name: name,
		URLs: []string{url},
	})
	return err
}

type CommitNode struct {
	Commit     *object.Commit
	BranchTips map[string]bool // 记录该提交所在分支的TIP
	Parents    []*CommitNode
	Children   []*CommitNode
}

func (c *GitHubClient) ShowLog() {
	repo := c.repo

	// 获取所有分支引用
	branches, err := repo.Branches()
	if err != nil {
		panic(err)
	}

	// 构建提交图
	commitGraph := make(map[plumbing.Hash]*CommitNode)
	var allCommits []*CommitNode

	// 遍历所有分支
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		commitIter, err := repo.Log(&git.LogOptions{
			From: ref.Hash(),
		})
		if err != nil {
			return err
		}

		// 遍历分支提交
		err = commitIter.ForEach(func(c *object.Commit) error {
			if _, exists := commitGraph[c.Hash]; !exists {
				node := &CommitNode{
					Commit:     c,
					BranchTips: make(map[string]bool),
					Parents:    make([]*CommitNode, 0),
					Children:   make([]*CommitNode, 0),
				}
				commitGraph[c.Hash] = node
				allCommits = append(allCommits, node)
			}

			// 标记分支TIP
			if ref.Hash() == c.Hash {
				commitGraph[c.Hash].BranchTips[ref.Name().Short()] = true
			}

			// 构建父子关系
			for _, ph := range c.ParentHashes {
				if parentNode, exists := commitGraph[ph]; exists {
					parentNode.Children = append(parentNode.Children, commitGraph[c.Hash])
					commitGraph[c.Hash].Parents = append(commitGraph[c.Hash].Parents, parentNode)
				}
			}

			return nil
		})

		return err
	})

	// 按时间排序
	sort.Slice(allCommits, func(i, j int) bool {
		return allCommits[i].Commit.Committer.When.After(allCommits[j].Commit.Committer.When)
	})

	// 生成图形化输出
	renderGraph(allCommits)
}

func renderGraph(commits []*CommitNode) {
	lines := make([][]string, 0)
	positions := make(map[plumbing.Hash]int)

	// 初始化连接线
	for i, commit := range commits {
		positions[commit.Commit.Hash] = i
		line := generateGraphLine(commit, positions, i)
		lines = append(lines, line)
	}

	// 打印结果
	for i, line := range lines {
		commit := commits[i]

		timeStr := shell.BrightCyan + commit.Commit.Author.When.Format("2006-01-02 15:04:05") + shell.ResetAll

		// 提交信息
		message := shell.BrightYellow + firstLine(commit.Commit.Message) + shell.ResetAll
		fmt.Printf("%s %s %s %s (%s)\n",
			formatGraphLine(line),
			timeStr,
			commit.Commit.Hash.String()[:7],
			message,
			commit.Commit.Author.Name,
		)
	}
}

func generateGraphLine(commit *CommitNode, positions map[plumbing.Hash]int, idx int) []string {
	line := make([]string, 5) // 控制图形宽度

	// 合并提交处理
	if len(commit.Parents) > 1 {
		line[0] = "|\\"
	} else if len(commit.Children) > 0 {
		line[0] = "|"
	}

	// 分支连接线
	for _, child := range commit.Children {
		if pos, exists := positions[child.Commit.Hash]; exists && pos < idx {
			line[pos-idx+2] = "/"
		}
	}

	return line
}

func formatGraphLine(line []string) string {
	str := ""
	for _, s := range line {
		if s == "" {
			str += ""
		} else {
			str += s
		}
	}
	return str
}

func firstLine(s string) string {
	for i := range s {
		if s[i] == '\n' {
			return s[:i]
		}
	}
	return s
}
