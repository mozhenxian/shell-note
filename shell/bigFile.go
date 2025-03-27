package shell

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

type FileInfo struct {
	Path string
	Size int64
}

type BySize []FileInfo

func (a BySize) Len() int           { return len(a) }
func (a BySize) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySize) Less(i, j int) bool { return a[i].Size > a[j].Size }

func Find(root string) {
	var (
		files    []FileInfo
		dirSizes = make(map[string]int64) // 存储目录最终大小
	)

	// 第一阶段：收集文件并构建目录树
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || path == root {
			return nil
		}

		if d.IsDir() {
			dirSizes[path] = 0 // 初始化目录
			return nil
		}

		// 处理文件
		if info, err := d.Info(); err == nil {
			files = append(files, FileInfo{Path: path, Size: info.Size()})

			// 累加文件到所有父目录
			current := filepath.Dir(path)
			for {
				dirSizes[current] += info.Size()
				parent := filepath.Dir(current)
				if parent == current || !strings.HasPrefix(current, root) {
					break
				}
				current = parent
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// 第二阶段：后序遍历计算目录大小
	var dirList []FileInfo
	for dir := range dirSizes {
		dirList = append(dirList, FileInfo{Path: dir, Size: dirSizes[dir]})
	}
	sort.Sort(BySize(dirList))

	// 取前10的目录（如果不足则全取）
	if len(dirList) > 10 {
		dirList = dirList[:10]
	}

	// 排序文件
	sort.Sort(BySize(files))

	// 输出结果
	fmt.Println("\nTop 10 largest files:")
	printTop10(files)

	fmt.Println("\nTop 10 largest directories:")
	printTop10(dirList)
}

func printTop10(items []FileInfo) {
	count := 10
	if len(items) < count {
		count = len(items)
	}

	for i := 0; i < count; i++ {
		fmt.Printf("%d. %s - %s\n", i+1, formatSize(items[i].Size), items[i].Path)
	}
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
