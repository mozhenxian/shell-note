package shell

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func Init(root string) {
	//root := "./db" // 指定根目录
	printDir(root)
	entries := getSortedEntries(root)
	printTree(root, entries, "", make([]int, 0))
}

func printDir(dir string) {
	fmt.Printf("%s%s%s%s\n", BrightCyan, Underline, dir, ResetAll)
}
func printFile(file string) {
	fmt.Printf("%s%s%s\n", Yellow, file, ResetAll)
}

func ColorPrint(color, text string) {
	fmt.Printf("%s%s%s\n", color, text, ResetAll)
}

// 递归打印目录结构
func printTree(parentPath string, entries []fs.DirEntry, prefix string, fatherIndex []int) {
	for i, entry := range entries {
		isLast := i == len(entries)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		index := i + 1
		if len(fatherIndex) > 0 {
			fmt.Printf("%s%s%s.%d ", prefix, connector, formatIndex(fatherIndex...), index)
		} else {
			fmt.Printf("%s%s%d ", prefix, connector, index)
		}
		// 打印当前条目名称
		if entry.IsDir() {
			printDir(entry.Name())
			//fmt.Printf("%s%s\u001B[32m%s\u001B[0m\n", prefix, connector, entry.Name())
		} else {
			printFile(entry.Name())
		}

		// 如果是目录，递归打印子项
		if entry.IsDir() {
			fullPath := filepath.Join(parentPath, entry.Name())
			subEntries := getSortedEntries(fullPath)
			if len(subEntries) > 0 {
				newPrefix := prefix
				if isLast {
					newPrefix += "    "
				} else {
					newPrefix += "│   "
				}
				printTree(fullPath, subEntries, newPrefix, append(fatherIndex, index))
			}
		}
	}
}

// 获取目录下所有条目（目录和文件）并按名称排序
func getSortedEntries(path string) []fs.DirEntry {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	// 过滤掉 .git 目录
	filteredEntries := make([]os.DirEntry, 0)
	for _, entry := range entries {
		if entry.Name() != ".git" {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	// 按名称排序（目录和文件混合排序）
	sort.Slice(filteredEntries, func(i, j int) bool {
		return filteredEntries[i].Name() < filteredEntries[j].Name()
	})

	return filteredEntries
}

var mapStr = make(map[string]string)

// 获取map[下标] 路径文件

func GetKeyMap(root string) map[string]string {
	entries := getSortedEntries(root)
	return getKvMap(root, entries, "", make([]int, 0), false)
}

// 获取map[路径文件] 下标

func GetValMap(root string) map[string]string {
	entries := getSortedEntries(root)
	return getKvMap(root, entries, "", make([]int, 0), true)
}

func getKvMap(parentPath string, entries []fs.DirEntry, prefix string, fatherIndex []int, isRevert bool) map[string]string {
	for i, entry := range entries {
		index := i + 1
		var keyString string
		if len(fatherIndex) > 0 {
			keyString = fmt.Sprintf("%s.%d", formatIndex(fatherIndex...), index)
		} else {
			keyString = fmt.Sprintf("%d", index)
		}

		var valString string
		valString = filepath.Join(parentPath, entry.Name())

		if isRevert {
			mapStr[valString] = keyString
		} else {
			mapStr[keyString] = valString
		}

		// 如果是目录，递归打印子项
		if entry.IsDir() {
			fullPath := filepath.Join(parentPath, entry.Name())
			subEntries := getSortedEntries(fullPath)
			if len(subEntries) > 0 {
				newPrefix := prefix
				getKvMap(fullPath, subEntries, newPrefix, append(fatherIndex, index), isRevert)
			}
		}
	}
	return mapStr
}

func formatIndex(fatherIndex ...int) string {
	// 将每个 int 转换为字符串
	parts := make([]string, len(fatherIndex))
	for i, num := range fatherIndex {
		parts[i] = strconv.Itoa(num)
	}
	// 用 "-" 拼接字符串
	return strings.Join(parts, ".")
}
