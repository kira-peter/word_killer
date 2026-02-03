package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// 样式定义
var (
	// 颜色
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	wordStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	highlightStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true)

	inputStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true)

	statsStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("117"))

	hintStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))

	menuSelectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true)

	menuNormalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))
)

// GameStats 游戏统计数据
type GameStats struct {
	TotalKeystrokes  int
	ValidKeystrokes  int
	CorrectChars     int
	WordsCompleted   int
	TotalLetters     int
	ElapsedSeconds   float64
	LettersPerSecond float64
	WordsPerSecond   float64
	AccuracyPercent  float64
}

// RenderWelcome 渲染欢迎界面
func RenderWelcome() string {
	var s string

	s += "\n\n"
	s += titleStyle.Render("    ╔════════════════════════════╗") + "\n"
	s += titleStyle.Render("    ║                            ║") + "\n"
	s += titleStyle.Render("    ║         Word Killer        ║") + "\n"
	s += titleStyle.Render("    ║                            ║") + "\n"
	s += titleStyle.Render("    ╚════════════════════════════╝") + "\n"
	s += "\n"
	s += statsStyle.Render("低调但绝不简单") + "\n\n"
	s += highlightStyle.Render("按 [Enter] 开始") + "\n"
	s += hintStyle.Render("按 [ESC] 退出") + "\n\n"

	return s
}

// RenderGame 渲染游戏画面
func RenderGame(words []string, highlightedIndices []int, input string, stats GameStats) string {
	var s string

	// 标题
	s += titleStyle.Render("Word Killer - 经典模式") + "\n\n"

	// 单词列表
	s += statsStyle.Render("剩余单词:") + "\n"
	for i, word := range words {
		// 检查是否高亮
		isHighlighted := false
		for _, idx := range highlightedIndices {
			if idx == i {
				isHighlighted = true
				break
			}
		}

		if isHighlighted {
			// 高亮显示匹配部分
			matchLen := len(input)
			if matchLen > len(word) {
				matchLen = len(word)
			}
			s += "  " + highlightStyle.Render(word[:matchLen]) + wordStyle.Render(word[matchLen:]) + "\n"
		} else {
			s += "  " + wordStyle.Render(word) + "\n"
		}
	}

	// 当前输入
	s += "\n" + statsStyle.Render("当前输入: ") + inputStyle.Render(input) + "\n"

	// 统计信息
	s += "\n" + titleStyle.Render("--- 统计信息 ---") + "\n"
	s += fmt.Sprintf("完成单词: %d | 剩余: %d | 用时: %.1fs\n",
		stats.WordsCompleted, len(words), stats.ElapsedSeconds)
	s += fmt.Sprintf("速度: %.2f 字母/秒 | %.2f 单词/秒\n",
		stats.LettersPerSecond, stats.WordsPerSecond)
	s += fmt.Sprintf("准确率: %.2f%%\n", stats.AccuracyPercent)

	// 提示信息
	s += "\n" + hintStyle.Render("[ESC] 暂停") + "\n"

	return s
}

// RenderPauseMenu 渲染暂停菜单
func RenderPauseMenu(selectedIndex int) string {
	var s string

	s += "\n\n"
	s += titleStyle.Render("    ╔════════════════════╗") + "\n"
	s += titleStyle.Render("    ║   游 戏 已 暂 停   ║") + "\n"
	s += titleStyle.Render("    ╚════════════════════╝") + "\n"
	s += "\n\n"

	options := []string{"继续游戏", "结束游戏"}
	for i, opt := range options {
		if i == selectedIndex {
			s += "    " + menuSelectedStyle.Render("> "+opt) + "\n"
		} else {
			s += "      " + menuNormalStyle.Render(opt) + "\n"
		}
	}

	s += "\n" + hintStyle.Render("[↑↓] 选择 | [Enter] 确认 | [ESC] 退出游戏") + "\n"

	return s
}

// RenderResults 渲染游戏结果
func RenderResults(stats GameStats, aborted bool) string {
	var s string

	s += "\n\n"
	s += titleStyle.Render("    ╔═════════════════════╗") + "\n"
	if aborted {
		s += titleStyle.Render("    ║     游 戏 结 束     ║") + "\n"
	} else {
		s += titleStyle.Render("    ║     恭 喜 完 成     ║") + "\n"
	}
	s += titleStyle.Render("    ╚═════════════════════╝") + "\n"
	s += "\n\n"

	s += statsStyle.Render("=== 最终数据统计 ===") + "\n\n"
	s += fmt.Sprintf("总敲击数:     %d\n", stats.TotalKeystrokes)
	s += fmt.Sprintf("有效敲击数:   %d\n", stats.ValidKeystrokes)
	s += fmt.Sprintf("正确字符数:   %d\n", stats.CorrectChars)
	s += fmt.Sprintf("完成单词数:   %d\n", stats.WordsCompleted)
	s += fmt.Sprintf("总字母数:     %d\n", stats.TotalLetters)
	s += fmt.Sprintf("总耗时:       %.2f 秒\n", stats.ElapsedSeconds)
	s += "\n"
	s += highlightStyle.Render("速度统计:") + "\n"
	s += fmt.Sprintf("字母速度:     %.2f 字母/秒\n", stats.LettersPerSecond)
	s += fmt.Sprintf("单词速度:     %.2f 单词/秒\n", stats.WordsPerSecond)
	s += "\n"
	s += highlightStyle.Render("准确率: ") + fmt.Sprintf("%.2f%%\n", stats.AccuracyPercent)

	s += "\n" + hintStyle.Render("按任意键退出...") + "\n"

	return s
}
