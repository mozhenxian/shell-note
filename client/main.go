package main

import (
	"note/client/lib"
	"note/shell"
	"os"
	"os/exec"
)

func main() {
	// 获取命令行参数
	args := os.Args[1:]
	if len(args) == 0 {
		lib.Help()
		return
	}
	action := args[0] // 动作
	// 如果参数不存在
	var parma string
	if len(args) >= 2 {
		parma = args[1] // 对应动作的参数
	}

	switch action {
	case "add":
		// 创建笔记
		lib.Edit(parma)
	case "addDir":
		// 创建文件夹
		lib.CreateDir(parma)
	case "v", "view":
		lib.ViewNote(parma) // 读取笔记
	case "s":
		lib.Search(parma)
	case "l", "list":
		shell.Init(lib.StorePath)
	case "start":
		exec.Command("note", "server").Start()
	case "server":
		lib.Start()
	case "move":
		lib.MoveFile(parma, args[2])
	case "h", "-h", "--help", "help":
		lib.Help()
	case "init":
		lib.InitGit()
	case "commit", "ci":
		lib.CommitGit(parma)
	case "push":
		lib.SyncGit()
	case "pull":
		lib.PullGit()
	case "rm":
		lib.RemoveFile(parma)
	case "log":
		lib.ShowLog()
	case "lz":
		shell.Find(parma)
	case "grep":
		shell.Search()
	default:
		lib.Edit(action)
	}
}
