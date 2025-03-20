package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"note/cfg"
	"os"
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
