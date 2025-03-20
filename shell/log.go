package shell

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

func Log(args ...interface{}) {
	// 获取调用者的信息
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	// 提取模块名（包名）
	var module string
	fn := runtime.FuncForPC(pc)
	if fn != nil {
		// 获取完整的函数名，如github.com/user/project/pkg.FuncName
		fnName := fn.Name()
		// 分割函数名，取包名
		parts := strings.Split(fnName, ".")
		if len(parts) > 1 {
			// 包名是最后一个.之前的部分
			module = parts[len(parts)-2]
		} else {
			module = fnName
		}
	} else {
		module = "unknown"
	}

	// 获取文件名的基名
	fileName := filepath.Base(file)

	// 将参数转换为字符串
	msg := fmt.Sprint(args...)

	// 格式化输出
	fmt.Printf("[%s] %s:%d - %s\n", module, fileName, line, msg)
}
