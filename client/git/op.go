package git

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"io/ioutil"
	"note/shell"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 同步到远程仓库

func (c *GitHubClient) Sync() ([]string, error) {
	// 拉取最新变更
	if err := c.pullChanges(); err != nil &&
		!errors.Is(err, git.NoErrAlreadyUpToDate) &&
		!errors.Is(err, git.ErrNonFastForwardUpdate) &&
		err.Error() != "remote repository is empty" {
		//if errors.Is(err, git.ErrNonFastForwardUpdate) {
		//	// 执行 git pull --rebase
		//	//err = exec.Command("git", "pull").Run()
		//	fmt.Println("有冲突请执行手动执行cd ./db/ && git pull 然后解决冲突")
		//}
		// 检测冲突
		if conflicts := c.detectConflicts(); len(conflicts) > 0 {
			return conflicts, fmt.Errorf("检测到 %d 处冲突", len(conflicts))
		}
		return nil, err
	}

	// 推送本地变更
	if err := c.pushChanges(); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			fmt.Println("已经是最新版本")
			return nil, nil
		}
		return nil, fmt.Errorf("推送失败: %v", err)
	}

	return nil, nil
}

func (c *GitHubClient) pullChanges() error {
	w, err := c.repo.Worktree()
	if err != nil {
		return err
	}

	return w.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(DefaultBranch),
		Auth:          c.auth,
	})
}

func (c *GitHubClient) pushChanges() error {
	return c.repo.Push(&git.PushOptions{
		Auth: c.auth,
	})
}

func (c *GitHubClient) pushU() error {
	cfg, err := c.repo.Config()
	if err != nil {
		return fmt.Errorf("failed to get config: %v", err)
	}

	// 配置上游分支
	cfg.Branches[DefaultBranch] = &config.Branch{
		Name:   DefaultBranch,
		Remote: "origin",
		Merge:  plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", DefaultBranch)),
	}

	// 保存配置
	if err := c.repo.Storer.SetConfig(cfg); err != nil {
		return fmt.Errorf("failed to set config: %v", err)
	}

	// 首次初始化自动推送
	_, err = c.Sync()
	if err != nil {
		shell.Log(err)
	}

	return nil
}

func getCurrentBranch(repo *git.Repository) (string, error) {
	// 获取 HEAD 引用
	ref, err := repo.Head()
	if err != nil {
		return "", err
	}

	// 解析引用名称以获取当前分支
	// 例如：refs/heads/main -> main
	branchName := ref.Name().Short()
	return branchName, nil
}

// 检测冲突文件
func (c *GitHubClient) detectConflicts() []string {
	var conflicts []string

	filepath.Walk(c.LocalPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, _ := ioutil.ReadFile(path)
		if strings.Contains(string(content), "<<<<<<<") {
			relPath, _ := filepath.Rel(c.LocalPath, path)
			conflicts = append(conflicts, relPath)
		}
		return nil
	})

	return conflicts
}

// 解决冲突

func (c *GitHubClient) ResolveConflict(filename, content string) error {
	absPath := filepath.Join(c.LocalPath, filename)
	if err := ioutil.WriteFile(absPath, []byte(content), 0644); err != nil {
		return err
	}
	return c.CommitChanges(fmt.Sprintf("解决冲突: %s", filename))
}

// 提交变更到本地仓库

func (c *GitHubClient) CommitChanges(message string) error {
	w, err := c.repo.Worktree()
	if err != nil {
		return err
	}

	err = os.Chdir(c.LocalPath)
	if err != nil {
		fmt.Println("change:", err)
	}

	// 添加所有变更
	if _, err := w.Add("."); err != nil {
		fmt.Println(err, c.LocalPath)
		return err
	}

	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Note Client",
			Email: "client@notes.com",
			When:  time.Now(),
		},
	})
	return err
}

// 获取笔记内容

func (c *GitHubClient) GetNote(filename string) (string, error) {
	notePath := filepath.Join(c.LocalPath, filename)
	content, err := ioutil.ReadFile(notePath)
	if err != nil {
		return "", fmt.Errorf("读取失败: %v", err)
	}
	return string(content), nil
}

func (c *GitHubClient) HandleConflictResolution(filename string) {
	fmt.Printf("解决冲突文件: %s\n", filename)
	fmt.Println("当前内容:")
	content, _ := c.GetNote(filename)
	fmt.Println(content)

	fmt.Println("输入新内容 (输入 :wq 结束):")
	var newContent strings.Builder
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == ":wq" {
			break
		}
		newContent.WriteString(line + "\n")
	}

	if err := c.ResolveConflict(filename, newContent.String()); err != nil {
		fmt.Printf("解决冲突失败: %v\n", err)
	} else {
		fmt.Println("冲突已解决，请重新同步")
	}
}

func (c *GitHubClient) Pull() error {
	w, err := c.repo.Worktree()
	if err != nil {
		return fmt.Errorf("获取工作树失败: %v", err)
	}

	// 获取远程引用信息
	remote, err := c.repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("获取远程仓库失败: %v", err)
	}

	// 拉取前先 fetch 最新数据
	if err := remote.Fetch(&git.FetchOptions{
		Auth:     c.auth,
		RefSpecs: []config.RefSpec{"+refs/heads/*:refs/remotes/origin/*"},
	}); err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("抓取远程更新失败: %v", err)
	}

	// 获取当前分支引用
	headRef, err := c.repo.Head()
	if err != nil {
		return fmt.Errorf("获取 HEAD 失败: %v", err)
	}

	// 执行合并操作
	pullOpts := &git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: headRef.Name(),
		Auth:          c.auth,
		Force:         true,
	}

	// 尝试合并
	if err := w.Pull(pullOpts); err != nil {
		fmt.Println("pull:", err)
	}

	return nil
}
