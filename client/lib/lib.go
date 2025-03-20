package lib

// 通用函数模块

import (
	"path/filepath"
	"strings"
)

func ExtractPath(input string) string {
	// 1. 查找第一个冒号的位置
	colonPos := strings.Index(input, ":")

	// 2. 分割路径部分和非路径部分
	var pathSegment string
	if colonPos == -1 {
		pathSegment = input
	} else {
		pathSegment = input[:colonPos]
	}

	// 3. 规范路径格式（处理多余斜杠和相对路径）
	cleanedPath := filepath.Clean(pathSegment)

	// 4. 保留原始 ./ 前缀（可选）
	if strings.HasPrefix(pathSegment, "./") && !strings.HasPrefix(cleanedPath, "./") {
		cleanedPath = "./" + cleanedPath
	}

	return cleanedPath
}
