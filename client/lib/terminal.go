package lib

import (
	"bytes"
	"fmt"
	"github.com/creack/pty"
	"io"
	"os/exec"
)

func Exec(cmdName string, Args ...string) string {
	cmd := exec.Command(cmdName, Args...)

	// 创建伪终端
	ptmx, err := pty.Start(cmd)
	if err != nil {
		panic(err)
	}
	defer ptmx.Close()

	// 实时捕获输出（含颜色）
	var coloredOutput bytes.Buffer
	go func() {
		//io.Copy(io.MultiWriter(os.Stdout, &coloredOutput), ptmx) // 同时输出到终端和缓冲区
		io.Copy(io.MultiWriter(&coloredOutput), ptmx) // 只输出到缓冲区
	}()

	// 等待命令结束
	err = cmd.Wait()
	if err != nil {
		if err.Error() == "exit status 1" {
			return ""
		}
		fmt.Printf("命令退出: %v\n", err)
	}

	// 使用带颜色的结果
	return coloredOutput.String()
}
