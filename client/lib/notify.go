package lib

import (
	"bufio"
	"fmt"
	"note/shell"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

var StartTime time.Time

func Start() {
	StartTime = time.Now()
	Loop()
}

func Loop() {
	var lastModTime time.Time // 记录文件最后修改时间
	// 1. 打开todolist文件
	file, err := os.OpenFile("./db/todolist", os.O_RDONLY, 0666)
	defer file.Close()
	if err != nil {
		shell.Log(err)
		return
	}
	for {
		// 2. 检查文件是否更新
		fileInfo, _ := file.Stat()
		if fileInfo.ModTime().Before(lastModTime) {
			time.Sleep(time.Second)
			continue
		}
		file.Close()
		file, err = os.OpenFile("./db/todolist", os.O_RDONLY, 0666)
		if err != nil {
			shell.Log(err)
			return
		}
		lastModTime = fileInfo.ModTime()

		// 3. 重置文件指针到开头
		file.Seek(0, 0)

		// 4. 按行读取新内容
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			t, err := parseTimeFromString(line)
			if err != nil {
				//log.Log(err)
				continue
			}
			if t.Truncate(time.Second) == time.Now().Truncate(time.Second) {
				go Notify(line, " 提醒！！！")
			}
		}
		time.Sleep(time.Second)
	}
}

func Notify(title, msg string) {
	DoNotify(title, msg)
	time.Sleep(time.Second)
	DoNotify(title, msg)
	time.Sleep(time.Second)
	DoNotify(title, msg)
}

func DoNotify(title, message string) {
	if runtime.GOOS == "darwin" {
		exec.Command("afplay", "/System/Library/Sounds/Ping.aiff").Start()
	} else {
		fmt.Print("\a")
	}
	// todo 后台运行时暂未能处理输出
	// 直接写入标准错误（通常无缓冲）
	//fmt.Fprintf(os.Stderr, "%s\n%s\n", title, message)

	// 或手动刷新标准输出
	//fmt.Printf("%s\n%s\n", title, message)
	//if f, ok := os.Stdout.(*os.File); ok {
	//	f.Sync() // 强制刷新缓冲区
	//}
	//os.Stdout.WriteString(title + "\n")
	//os.Stdout.WriteString(message + "\n")
	fmt.Printf("\u001B[33m %s (%s) \u001b[0m\n", title, message)
}

// 解析时间字符串，返回对应的时间戳
func parseTimeFromString(s string) (time.Time, error) {
	// 匹配两种时间格式：HH:MM 和 Xmin
	hhmmRegex := regexp.MustCompile(`(\d{1,2}):(\d{2})`)
	minRegex := regexp.MustCompile(`(\d+)min`)

	// 先尝试匹配 HH:MM 格式
	if hhmmMatch := hhmmRegex.FindStringSubmatch(s); len(hhmmMatch) == 3 {
		hour, _ := strconv.Atoi(hhmmMatch[1])
		minute, _ := strconv.Atoi(hhmmMatch[2])

		// 验证时间有效性
		if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
			return time.Time{}, fmt.Errorf("invalid time: %02d:%02d", hour, minute)
		}

		// 构造当天时间
		now := time.Now()
		target := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.Local)

		return target, nil
	}

	// 尝试匹配 Xmin 格式
	if minMatch := minRegex.FindStringSubmatch(s); len(minMatch) == 2 {
		minutes, _ := strconv.Atoi(minMatch[1])
		return StartTime.Add(time.Duration(minutes) * time.Minute), nil
	}

	return time.Time{}, fmt.Errorf("no valid time pattern found")
}
