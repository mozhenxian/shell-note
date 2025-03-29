package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
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
	enterDir     = "./"
)

func Find(root string) {
	enterDir = root
	wg := &sync.WaitGroup{}
	// 创建一个通道用于控制耗时显示的退出
	done := make(chan struct{})
	go timePass(done)

	dirSizeMutex.Lock()
	dirSizes[root] = 0 // 初始化目录
	dirSizeMutex.Unlock()
	wg.Add(1)
	go doWalkDir(root, wg)
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
	close(done)

	// 输出结果
	fmt.Println("\n" + BrightCyan + "Top 10 largest files:" + ResetAll)
	printTop10(files, root)

	fmt.Println("\n" + BrightCyan + "Top 10 largest directories:" + ResetAll)
	printTop10(dirList, root)
}

func timePass(done chan struct{}) {
	start := time.Now()
	for {
		select {
		case <-done:
			return // 收到退出信号时结束协程
		default:
			// 计算已耗时（保留1位小数）
			elapsed := time.Since(start).Seconds()
			// \r 表示覆盖当前行，%.1f 保留1位小数
			fmt.Printf("\r耗时: \033[31m%.1f seconds\033[0m", elapsed)
			time.Sleep(100 * time.Millisecond) // 刷新间隔
		}
	}
}

func doWalkDir(root string, wg *sync.WaitGroup) {
	defer wg.Done()
	fd, err := syscall.Open(root, syscall.O_RDONLY|syscall.O_DIRECTORY, 0)
	if err != nil {
		return
	}
	defer syscall.Close(fd)

	var buf [32 * 1024]byte
	for {
		n, err := syscall.ReadDirent(fd, buf[:])
		if err != nil {
			return
		}
		if n == 0 {
			break
		}

		for remain := buf[:n]; len(remain) > 0; {
			dirent := (*syscall.Dirent)(unsafe.Pointer(&remain[0]))
			name, ok := parseDirent(dirent, remain)
			if !ok {
				break
			}
			remain = remain[dirent.Reclen:]

			if name == "." || name == ".." {
				continue
			}

			path := filepath.Join(root, name)
			if dirent.Type == syscall.DT_DIR {
				dirSizeMutex.Lock()
				dirSizes[path] = 0 // 初始化目录
				dirSizeMutex.Unlock()
				wg.Add(1)
				go doWalkDir(path, wg)
				continue
			} else {
				info, err := os.Stat(path)
				if err != nil {
					return
				}

				files = append(files, FileInfo{Path: path, Size: info.Size()})

				current := root // filepath.Dir(path)
				for {
					dirSizeMutex.Lock()
					dirSizes[current] += info.Size()
					dirSizeMutex.Unlock()
					parent := filepath.Dir(current)
					//fmt.Println(current, parent)
					if current == enterDir {
						break
					}
					current = parent
				}
			}
		}
	}

}

func parseDirent(dirent *syscall.Dirent, buf []byte) (string, bool) {
	if dirent.Reclen == 0 || int(dirent.Reclen) > len(buf) {
		return "", false
	}

	nameBuf := make([]byte, 0, 256)
	for i := 0; ; i++ {
		if dirent.Name[i] == 0 {
			break
		}
		nameBuf = append(nameBuf, byte(dirent.Name[i]))
	}

	return string(nameBuf), true
}

func printTop10(items []FileInfo, root string) {
	count := 10
	if len(items) < count {
		count = len(items)
	}

	fmt.Printf(BrightYellow)
	for i := 0; i < count; i++ {
		path := strings.TrimPrefix(items[i].Path, root)
		if path == "" {
			path = root
		}
		fmt.Printf("%d. %s - %s\n", i+1, formatSize(items[i].Size), path)
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
