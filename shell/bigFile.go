package shell

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

type FileInfo struct {
	Path string
	Size int64
}

type BySize []FileInfo

func (a BySize) Len() int           { return len(a) }
func (a BySize) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySize) Less(i, j int) bool { return a[i].Size > a[j].Size } // 降序排序

func Find(root string) {

	var files []FileInfo
	var dirs []FileInfo

	// 遍历目录
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			//fmt.Printf("Error accessing path %q: %v\n", path, err)
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		if !d.IsDir() {
			files = append(files, FileInfo{Path: path, Size: info.Size()})
		} else {
			// 计算目录大小
			dirSize, err := getDirSize(path)
			if err != nil {
				return nil
			}
			dirs = append(dirs, FileInfo{Path: path, Size: dirSize})
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", root, err)
		os.Exit(1)
	}

	// 排序
	sort.Sort(BySize(files))
	sort.Sort(BySize(dirs))

	// 打印结果
	fmt.Println("\nTop 10 largest files:")
	printTop10(files)

	fmt.Println("\nTop 10 largest directories:")
	printTop10(dirs)
}

func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
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
