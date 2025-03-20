package shell

import "fmt"

// ANSI 颜色/样式常量定义
const (
	ResetAll = "\u001B[0m" // 重置所有样式

	// 基础前景色（字体颜色）
	Black   = "\u001B[30m"
	Red     = "\u001B[31m"
	Green   = "\u001B[32m" // 用户示例中的绿色
	Yellow  = "\u001B[33m"
	Blue    = "\u001B[34m"
	Magenta = "\u001B[35m"
	Cyan    = "\u001B[36m"
	White   = "\u001B[37m"

	// 亮色前景色（加粗字体颜色）
	BrightBlack   = "\u001B[90m"
	BrightRed     = "\u001B[91m"
	BrightGreen   = "\u001B[92m"
	BrightYellow  = "\u001B[93m"
	BrightBlue    = "\u001B[94m"
	BrightMagenta = "\u001B[95m"
	BrightCyan    = "\u001B[96m"
	BrightWhite   = "\u001B[97m"

	// 背景色
	BlackBg   = "\u001B[40m"
	RedBg     = "\u001B[41m"
	GreenBg   = "\u001B[42m"
	YellowBg  = "\u001B[43m"
	BlueBg    = "\u001B[44m"
	MagentaBg = "\u001B[45m"
	CyanBg    = "\u001B[46m"
	WhiteBg   = "\u001B[47m"

	// 亮色背景色
	BrightBlackBg   = "\u001B[100m"
	BrightRedBg     = "\u001B[101m"
	BrightGreenBg   = "\u001B[102m"
	BrightYellowBg  = "\u001B[103m"
	BrightBlueBg    = "\u001B[104m"
	BrightMagentaBg = "\u001B[105m"
	BrightCyanBg    = "\u001B[106m"
	BrightWhiteBg   = "\u001B[107m"

	// 样式控制
	Bold      = "\u001B[1m" // 粗体
	Underline = "\u001B[4m" // 下划线
	Italic    = "\u001B[3m" // 斜体（部分终端不支持）
	Blink     = "\u001B[5m" // 闪烁（部分终端禁用）
	Reverse   = "\u001B[7m" // 反转前景/背景色
)

func Test() {
	// 示例：绿色粗体文本
	fmt.Printf("%s%sHello, 绿色粗体文本！%s\n", Green, Bold, ResetAll)

	// 示例：黄色背景+蓝色文字
	fmt.Printf("%s%s蓝色文字+黄色背景%s\n", Blue, YellowBg, ResetAll)

	// 示例：亮青色下划线
	fmt.Printf("%s%s亮青色下划线%s\n", BrightCyan, Underline, ResetAll)

	// 组合样式：粗体+红色+白色背景
	fmt.Printf("%s%s%s粗体红字白底%s\n", Bold, Red, WhiteBg, ResetAll)
}
