package shell

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type FileInfo struct {
	Path string
	Size int64
}

type BySize []FileInfo

func (a BySize) Len() int           { return len(a) }
func (a BySize) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySize) Less(i, j int) bool { return a[i].Size > a[j].Size }

var (
	files        []FileInfo
	dirSizes     = make(map[string]int64) // 存储目录最终大小
	dirSizeMutex sync.Mutex
)

func Find(root string) {
	entries, err := os.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}
	wg := &sync.WaitGroup{}

	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		if entry.IsDir() {
			wg.Add(1)
			go doWalkDir(path, wg)
		} else {
			setFileSize(path)
		}
	}
	wg.Wait()

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
	fmt.Println("\n" + BrightCyan + "Top 10 largest files:" + ResetAll)
	printTop10(files, root)

	fmt.Println("\n" + BrightCyan + "Top 10 largest directories:" + ResetAll)
	printTop10(dirList, root)
}

func setFileSize(path string) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Printf("获取文件信息失败: %v\n", err)
		return
	}

	// 判断是否为文件
	if !fileInfo.Mode().IsRegular() {
		fmt.Println("这不是一个普通文件")
		return
	}

	// 获取文件大小（单位：字节）
	fileSize := fileInfo.Size()
	files = append(files, FileInfo{Path: path, Size: fileSize})
}

func doWalkDir(root string, wg *sync.WaitGroup) {
	defer wg.Done()
	// 第一阶段：收集文件并构建目录树
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || path == root {
			return nil
		}

		if d.IsDir() {
			dirSizeMutex.Lock()
			dirSizes[path] = 0 // 初始化目录
			dirSizeMutex.Unlock()
			return nil
		}

		// 处理文件
		if info, err := d.Info(); err == nil {
			files = append(files, FileInfo{Path: path, Size: info.Size()})

			// 累加文件到所有父目录
			current := filepath.Dir(path)
			for {
				dirSizeMutex.Lock()
				dirSizes[current] += info.Size()
				dirSizeMutex.Unlock()
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
}

func printTop10(items []FileInfo, root string) {
	count := 10
	if len(items) < count {
		count = len(items)
	}

	fmt.Printf(BrightYellow)
	for i := 0; i < count; i++ {
		fmt.Printf("%d. %s - %s\n", i+1, formatSize(items[i].Size), strings.TrimPrefix(items[i].Path, root))
	}
	fmt.Printf(ResetAll)
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
