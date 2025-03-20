package lib

import (
	"fmt"
	"github.com/alecthomas/chroma/quick"
	"note/cfg"
	"note/client/git"
	"note/shell"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	StorePath = "./db/" // 默认值会被覆盖
	Editor    = "vi"
	RemoteURL = ""
)

func init() {
	// 初始化配置
	StorePath = cfg.DefaultCfg.App.Db + "/"
	Editor = cfg.DefaultCfg.App.Editor
	RemoteURL = cfg.DefaultCfg.Git.RemoteURL
}

func Help() {
	quick.Highlight(os.Stdout, git.HelpStr, "go", "terminal256", "monokai")
}

func MoveFile(filePath, targetPath string) {
	// 移动文件
	err := os.Rename(StorePath+filePath, StorePath+targetPath)
	if err != nil {
		fmt.Println("移动文件失败:", err)
		return
	}
	CommitGit("移动文件: from " + filePath + " to " + targetPath)
	fmt.Println("文件移动成功！")
}

func Edit(fileName string) {
	// 如果fileName 格式为1 或者 1.1
	if isIndexString(fileName) {
		Map := shell.GetKeyMap(StorePath)
		fileName = Map[fileName]
	} else {
		fileName = StorePath + fileName
	}
	isModify := createNote(fileName)
	if isModify {
		CommitGit(filepath.Base(fileName))
	}
}

func isIndexString(fileName string) bool {
	if strings.Contains(fileName, ".") {
		arr := strings.Split(fileName, ".")
		// 判断 arr[0] 是否为数字
		i, err := strconv.Atoi(arr[0])
		if err != nil {
			return false
		}
		// 判断 i 是否为正数
		if i <= 0 {
			return false
		}
		return true
	}
	// 判断fileName 是否为数字
	for _, r := range fileName {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func createNote(path string) bool {
	// 获取原始文件状态
	var originalExists bool
	var originalModTime time.Time
	if info, err := os.Stat(path); err == nil {
		originalExists = true
		originalModTime = info.ModTime()
	} else {
		originalExists = false
	}

	// 创建命令：vim 编辑文件
	cmd := exec.Command(Editor, path) // 假设 Editor 是已定义的编辑器路径变量

	// 连接到当前终端
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 执行编辑命令
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	// 获取编辑后的文件状态
	var newExists bool
	var newModTime time.Time
	if info, err := os.Stat(path); err == nil {
		newExists = true
		newModTime = info.ModTime()
	} else {
		newExists = false
	}

	// 判断文件是否被修改
	switch {
	case originalExists != newExists: // 存在状态变化
		return true
	case !originalExists && !newExists: // 文件从未存在
		return false
	default: // 比较修改时间
		return !originalModTime.Equal(newModTime)
	}
}

func CreateDir(dirName string) {
	path := StorePath + dirName
	err := os.MkdirAll(path, 0755)
	if err != nil {
		shell.Log(err)
		return
	}
}

func ViewNote(fileName string) {
	var path string
	if isIndexString(fileName) {
		Map := shell.GetKeyMap(StorePath)
		path = Map[fileName]
	} else {
		path = StorePath + fileName
	}
	//file, err := os.Open(path)
	bytes, err := os.ReadFile(path)
	if err != nil {
		shell.Log(err)
		return
	}
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	if ext == "" {
		ext = "go"
	}
	// 根据文件扩展名设置语法高亮
	quick.Highlight(os.Stdout, string(bytes), ext, "terminal256", "monokai")

}

// 搜索本目录所有匹配的文件
func Search(keyWord string) {
	output := Exec("grep", "--color", "-irn", keyWord, StorePath)

	ret := strings.Split(output, "\n")
	//fmt.Println(output)

	Map := shell.GetValMap(StorePath)
	//fmt.Println(Map)

	for _, subString := range ret {
		s := ExtractPath(subString)
		s = strings.Replace(s, "./", "", 1)
		first := strings.Index(subString, ":")
		//left := subString[:first]
		_ = subString[first+1:]
		//match := strings.Split(subString, ":")
		fmt.Println(Map[s], subString)
	}

}

// =================== 云仓库存储 ==================

func InitGit() {
	g, err := git.NewClient(StorePath, RemoteURL, "")
	if err != nil {
		shell.Log(err)
		return
	}
	fmt.Println(g.SSHKeyPath)
}

func CommitGit(title string) {
	g, err := git.NewClient(StorePath, RemoteURL, "")
	if err != nil {
		shell.Log(err)
		return
	}
	err = g.CommitChanges(title)
	if err != nil {
		shell.Log(err)
		return
	}
}
func PullGit() {
	g, err := git.NewClient(StorePath, RemoteURL, "")
	if err != nil {
		shell.Log(err)
		return
	}
	g.Pull()
}

func SyncGit() {
	g, err := git.NewClient(StorePath, RemoteURL, "")
	if err != nil {
		shell.Log(err)
		return
	}
	conflicts, err := g.Sync()
	if err != nil {
		if len(conflicts) > 0 {
			fmt.Println("发现冲突文件:")
			for _, f := range conflicts {
				fmt.Printf(" - %s\n", f)
			}
			g.HandleConflictResolution(conflicts[0])
		} else {
			fmt.Println("sync fail:", err)
		}
	}
}
