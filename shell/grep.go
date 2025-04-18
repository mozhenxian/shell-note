package shell

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

const (
	redColor   = "\033[31m"
	resetColor = "\033[0m"
)

var (
	dirPath    = flag.String("d", ".", "Search directory")
	keyword    = flag.String("k", "", "Keywords (| for OR, & for AND with order)")
	workers    = flag.Int("w", 10, "Worker goroutines")
	fileMatch  = flag.String("f", ".*", "Filename pattern")
	contextLen = flag.Int("l", 40, "内容长度")
)

type searchPattern struct {
	regex      *regexp.Regexp
	keywords   []string
	isAndMode  bool
	colorRegex *regexp.Regexp
}

type task struct {
	path string
	info os.FileInfo
}

func Search() {
	os.Args = os.Args[1:]
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	if *keyword == "" {
		fmt.Println("必须指定关键字 (-k)")
		os.Exit(1)
	}

	// 解析关键词模式
	pattern, err := parseSearchPattern(*keyword)
	if err != nil {
		fmt.Printf("正则表达式错误: %v\n", err)
		os.Exit(1)
	}

	// 创建任务和结果通道
	taskChan := make(chan task, 100)
	resultChan := make(chan string, 100)

	// 启动工作池
	var wg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		//fmt.Println("start:", i)
		go worker(taskChan, resultChan, pattern, &wg, i)
	}

	// 结果收集
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 生成任务
	go generateTasks(taskChan)

	// 输出结果
	for res := range resultChan {
		fmt.Println(res)
	}
}

func parseSearchPattern(input string) (*searchPattern, error) {
	var pattern searchPattern

	if strings.Contains(input, "|") {
		// OR 模式
		keywords := strings.Split(input, "|")
		escaped := make([]string, len(keywords))
		for i, k := range keywords {
			escaped[i] = regexp.QuoteMeta(k)
		}
		regexStr := "(" + strings.Join(escaped, "|") + ")"
		re, err := regexp.Compile(regexStr)
		if err != nil {
			return nil, err
		}
		pattern.regex = re
		pattern.keywords = keywords
		pattern.colorRegex = re
	} else if strings.Contains(input, "&") {
		// AND 顺序模式优化
		keywords := strings.Split(input, "&")
		escaped := make([]string, len(keywords))

		// 转义特殊字符并添加单词边界
		for i, k := range keywords {
			escaped[i] = regexp.QuoteMeta(k)
		}

		// 生成高效正则表达式
		regexStr := `(?i)` + strings.Join(escaped, `.*?`) // 使用 .*? 连接关键字，按顺序匹配
		re, err := regexp.Compile(regexStr)
		if err != nil {
			return nil, err
		}

		// 创建颜色高亮正则
		colorRegexStr := `(` + strings.Join(escaped, `|`) + `)`
		colorRe, err := regexp.Compile(colorRegexStr)
		if err != nil {
			return nil, err
		}

		pattern.regex = re
		pattern.keywords = keywords
		pattern.isAndMode = true
		pattern.colorRegex = colorRe
	} else {
		// 单个关键词
		re, err := regexp.Compile(`(` + regexp.QuoteMeta(input) + `)`)
		if err != nil {
			return nil, err
		}
		pattern.regex = re
		pattern.keywords = []string{input}
		pattern.colorRegex = re
	}

	return &pattern, nil
}

func worker(tasks <-chan task, results chan<- string, pattern *searchPattern, wg *sync.WaitGroup, i int) {
	defer func() {
		//fmt.Println("down:", i)
		wg.Done()
	}()

	for t := range tasks {
		processFile(t.path, pattern, results)
	}
}

func processFile(path string, pattern *searchPattern, results chan<- string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	lineNum := 0

	for {
		line, err := reader.ReadString('\n')
		//fmt.Println("line:", line, err)
		lineNum++
		if err != nil && err == io.EOF {
			break
		}

		// 检查AND模式顺序
		if pattern.isAndMode {
			indexes := make([]int, len(pattern.keywords))
			lastIndex := 0
			valid := true

			for i, kw := range pattern.keywords {
				pos := strings.Index(line[lastIndex:], kw)
				if pos == -1 {
					valid = false
					break
				}
				indexes[i] = lastIndex + pos
				lastIndex += pos + len(kw)
			}

			if !valid {
				continue
			}
		}

		// 正则匹配
		matches := pattern.regex.FindAllStringIndex(line, -1)
		if len(matches) == 0 {
			if err == io.EOF {
				break
			}
			continue
		}

		// 处理匹配结果
		coloredLine := pattern.colorRegex.ReplaceAllStringFunc(line, func(m string) string {
			return redColor + m + resetColor
		})

		// 截取上下文
		for _, match := range matches {
			start := match[0]
			end := match[1]

			ctxStart := max(0, start-*contextLen)
			ctxEnd := min(len(coloredLine), end+*contextLen)
			context := coloredLine[ctxStart:ctxEnd]

			// 去除可能的颜色代码截断
			context = fixColorCodes(context)

			results <- fmt.Sprintf("%s:%d: %s", path, lineNum, strings.TrimSpace(context))
		}

		if err == io.EOF {
			break
		}
	}
}

// 修复颜色代码截断问题
func fixColorCodes(s string) string {
	re := regexp.MustCompile(`\033\[[0-9;]*m?`)
	matches := re.FindAllStringIndex(s, -1)

	if len(matches) == 0 {
		return s
	}

	// 检查最后一个颜色代码是否完整
	lastMatch := matches[len(matches)-1]
	if lastMatch[1] != len(s) {
		return s + resetColor
	}
	return s
}

func generateTasks(taskChan chan<- task) {
	fileRegex := regexp.MustCompile(*fileMatch)
	filepath.Walk(*dirPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && fileRegex.MatchString(info.Name()) {
			taskChan <- task{path, info}
		}
		return nil
	})
	close(taskChan)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
