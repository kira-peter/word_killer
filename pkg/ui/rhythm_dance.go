package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// RhythmDanceStats 节奏舞蹈模式统计信息
type RhythmDanceStats struct {
	RemainingTime  int
	TotalScore     int
	PerfectCount   int
	NiceCount      int
	OKCount        int
	MissCount      int
	CompletedWords int
	CurrentCombo   int
	MaxCombo       int
}

// RhythmBarInfo 节奏条信息
type RhythmBarInfo struct {
	PointerPosition float64
	GoldenRatio     float64
}

// JudgmentEffectInfo 判定特效信息
type JudgmentEffectInfo struct {
	LastJudgment     string
	LastJudgmentTime time.Time
}

// RenderRhythmDanceGame 渲染节奏舞蹈模式主界面
func RenderRhythmDanceGame(
	danceFrame string,
	currentWord string,
	userInput string,
	rhythmBar RhythmBarInfo,
	stats RhythmDanceStats,
	effectInfo JudgmentEffectInfo,
) string {
	var s strings.Builder

	// === 顶部：统计信息 ===
	s.WriteString(renderRhythmStats(stats))
	s.WriteString("\n")

	// === 中部：主游戏区域（舞蹈小人 + 单词 + 节奏条）===
	mainArea := renderRhythmMainArea(danceFrame, currentWord, userInput, rhythmBar, effectInfo)
	s.WriteString(mainArea)
	s.WriteString("\n")

	// === 底部：提示 ===
	s.WriteString(hintStyle.Render("  [Space] Judge Timing  │  [ESC] Pause"))
	s.WriteString("\n")

	return s.String()
}

// renderRhythmStats 渲染统计信息栏
func renderRhythmStats(stats RhythmDanceStats) string {
	// 时间显示（小于10秒时红色警告）
	var timeStyle lipgloss.Style
	if stats.RemainingTime < 10 {
		timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Background(lipgloss.Color("52"))
	} else {
		timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)
	}

	timeDisplay := fmt.Sprintf("Time: %02d", stats.RemainingTime)
	scoreDisplay := fmt.Sprintf("Score: %d", stats.TotalScore)
	comboDisplay := fmt.Sprintf("Combo: %d", stats.CurrentCombo)

	statusLine := fmt.Sprintf("%s  │  %s  │  %s",
		timeStyle.Render(timeDisplay),
		scoreDisplay,
		comboDisplay)

	statusStyled := headerStyle.Render(statusLine)
	return lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(statusStyled)
}

// renderRhythmMainArea 渲染主游戏区域
func renderRhythmMainArea(
	danceFrame string,
	currentWord string,
	userInput string,
	rhythmBar RhythmBarInfo,
	effectInfo JudgmentEffectInfo,
) string {
	var lines []string

	// 1. 舞蹈小人区域（上部，3行）
	danceLines := renderDanceCharacter(danceFrame)
	lines = append(lines, danceLines...)

	// 空行
	lines = append(lines, "")

	// 2. 单词区域和节奏条（中间，左右布局）
	wordAndBarLines := renderWordAndRhythmBar(currentWord, userInput, rhythmBar, effectInfo)
	lines = append(lines, wordAndBarLines...)

	// 3. 判定计数区域（底部）
	lines = append(lines, "")
	judgmentLine := renderJudgmentCounts(
		effectInfo.LastJudgment,
		effectInfo.LastJudgmentTime,
	)
	lines = append(lines, judgmentLine)

	return wordBoxStyle.Render(strings.Join(lines, "\n"))
}

// renderDanceCharacter 渲染舞蹈小人
func renderDanceCharacter(danceFrame string) []string {
	var lines []string

	// 将舞蹈帧按行分割
	frameLines := strings.Split(danceFrame, "\n")

	// 居中显示
	for _, line := range frameLines {
		centeredLine := lipgloss.NewStyle().
			Width(contentWidth - 8).
			Align(lipgloss.Center).
			Render(highlightStyle.Render(line))
		lines = append(lines, centeredLine)
	}

	// 确保始终是3行
	for len(lines) < 3 {
		lines = append(lines, "")
	}

	return lines
}

// renderWordAndRhythmBar 渲染单词和节奏条（左右布局）
func renderWordAndRhythmBar(
	currentWord string,
	userInput string,
	rhythmBar RhythmBarInfo,
	effectInfo JudgmentEffectInfo,
) []string {
	var lines []string

	// 左侧：单词显示
	wordDisplay := renderRhythmDanceWordArea(currentWord, userInput)

	// 右侧：节奏条
	rhythmBarDisplay := renderRhythmBar(rhythmBar, effectInfo)

	// 合并左右两侧
	wordLines := strings.Split(wordDisplay, "\n")
	barLines := strings.Split(rhythmBarDisplay, "\n")

	maxLines := len(wordLines)
	if len(barLines) > maxLines {
		maxLines = len(barLines)
	}

	for i := 0; i < maxLines; i++ {
		var leftPart, rightPart string

		if i < len(wordLines) {
			leftPart = wordLines[i]
		} else {
			leftPart = strings.Repeat(" ", 35)
		}

		if i < len(barLines) {
			rightPart = barLines[i]
		} else {
			rightPart = ""
		}

		line := leftPart + rightPart
		lines = append(lines, line)
	}

	return lines
}

// renderRhythmDanceWordArea 渲染单词和输入（节奏舞蹈模式专用）
func renderRhythmDanceWordArea(currentWord string, userInput string) string {
	var lines []string

	// 标题
	lines = append(lines, "  "+titleStyle.Render("Word:"))
	lines = append(lines, "")

	// 显示单词（根据输入高亮）
	var renderedWord string
	if len(userInput) == 0 {
		renderedWord = wordStyle.Render(currentWord)
	} else {
		// 高亮已输入的部分
		matchLen := len(userInput)
		if matchLen > len(currentWord) {
			matchLen = len(currentWord)
		}

		// 检查是否正确
		isCorrect := true
		for i := 0; i < matchLen; i++ {
			if i >= len(currentWord) || userInput[i] != currentWord[i] {
				isCorrect = false
				break
			}
		}

		if isCorrect {
			renderedWord = highlightStyle.Render(currentWord[:matchLen]) +
				wordStyle.Render(currentWord[matchLen:])
		} else {
			// 错误显示为红色
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)
			renderedWord = errorStyle.Render(currentWord)
		}
	}

	lines = append(lines, "  "+renderedWord)
	lines = append(lines, "")

	// 输入显示
	lines = append(lines, "  "+statItemStyle.Render("Input: ")+inputStyle.Render(userInput))

	return strings.Join(lines, "\n")
}

// renderRhythmBar 渲染节奏条
func renderRhythmBar(rhythmBar RhythmBarInfo, effectInfo JudgmentEffectInfo) string {
	var lines []string

	// 节奏条宽度
	const barWidth = 40

	// 计算指针和黄金点的位置
	pointerPos := int(rhythmBar.PointerPosition * float64(barWidth))
	goldenPos := int(rhythmBar.GoldenRatio * float64(barWidth))

	if pointerPos < 0 {
		pointerPos = 0
	}
	if pointerPos >= barWidth {
		pointerPos = barWidth - 1
	}

	// 计算距离（用于颜色渐变）
	distance := math.Abs(rhythmBar.PointerPosition - rhythmBar.GoldenRatio)

	// 根据距离选择指针颜色
	var pointerStyle lipgloss.Style
	if distance <= 0.05 {
		pointerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true) // 黄色（Perfect）
	} else if distance <= 0.15 {
		pointerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true) // 青色（Nice）
	} else if distance <= 0.30 {
		pointerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")) // 白色（OK）
	} else {
		pointerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // 灰色（Miss）
	}

	// 渲染节奏条
	bar := make([]rune, barWidth)
	for i := 0; i < barWidth; i++ {
		bar[i] = '─'
	}

	// 标记黄金分割点
	bar[goldenPos] = '▼'

	// 标记指针
	bar[pointerPos] = '◆'

	// 应用颜色
	var styledBar strings.Builder
	for i, ch := range bar {
		if i == pointerPos {
			styledBar.WriteString(pointerStyle.Render(string(ch)))
		} else if i == goldenPos {
			styledBar.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")).
				Render(string(ch)))
		} else {
			// 根据距离黄金点的距离设置颜色渐变
			distFromGolden := math.Abs(float64(i-goldenPos) / float64(barWidth))
			var barCharStyle lipgloss.Style
			if distFromGolden < 0.1 {
				barCharStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // 黄色
			} else if distFromGolden < 0.2 {
				barCharStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // 浅黄
			} else {
				barCharStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // 灰色
			}
			styledBar.WriteString(barCharStyle.Render(string(ch)))
		}
	}

	lines = append(lines, "├"+styledBar.String()+"┤")

	// 判定特效显示
	if effectInfo.LastJudgment != "" {
		elapsed := time.Since(effectInfo.LastJudgmentTime)
		if elapsed < 1*time.Second {
			effectText := renderJudgmentEffect(effectInfo.LastJudgment, elapsed)
			lines = append(lines, effectText)
		}
	}

	return strings.Join(lines, "\n")
}

// renderJudgmentEffect 渲染判定特效
func renderJudgmentEffect(judgment string, elapsed time.Duration) string {
	var effectStyle lipgloss.Style
	var text string

	switch judgment {
	case "Perfect":
		effectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)
		text = "★ PERFECT ★"
	case "Nice":
		effectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("51")).
			Bold(true)
		text = "✓ Nice!"
	case "OK":
		effectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
		text = "○ OK"
	case "Miss":
		effectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
		text = "✗ Miss..."
	default:
		return ""
	}

	// 文字上浮效果（根据时间偏移）
	ms := elapsed.Milliseconds()
	offset := int(ms / 100) // 每100ms上移一行

	var padding string
	for i := 0; i < offset; i++ {
		padding += " "
	}

	// 闪烁效果（前500ms）
	if ms < 500 && ms%200 < 100 {
		return lipgloss.NewStyle().
			Width(42).
			Align(lipgloss.Center).
			Render(padding + effectStyle.Render(text))
	}

	return lipgloss.NewStyle().
		Width(42).
		Align(lipgloss.Center).
		Render(padding + effectStyle.Render(text))
}

// renderJudgmentCounts 渲染判定计数
func renderJudgmentCounts(lastJudgment string, lastJudgmentTime time.Time) string {
	// 这个函数将在完整版中显示 Perfect/Nice/OK/Miss 的计数
	// 现在先返回空，稍后实现
	return ""
}

// RenderRhythmDanceResults 渲染节奏舞蹈模式结果页面
func RenderRhythmDanceResults(stats RhythmDanceStats, selectedOption int, animFrame int) string {
	var s strings.Builder

	// === TOP: Header ===
	header := fmt.Sprintf("Total Score: %d  │  Max Combo: %d", stats.TotalScore, stats.MaxCombo)
	headerStyled := headerStyle.Render(header)
	s.WriteString(lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(headerStyled))
	s.WriteString("\n")

	// === MIDDLE: Statistics ===
	statsArea := renderRhythmDanceResultsArea(stats, selectedOption, animFrame)
	s.WriteString(statsArea)
	s.WriteString("\n")

	// === BOTTOM: Hints ===
	hints := inputBoxStyle.Render("[↑↓] Select  │  [Enter] Confirm  │  [ESC] Exit")
	s.WriteString(hints)
	s.WriteString("\n")

	return s.String()
}

// renderRhythmDanceResultsArea 渲染结果统计区域
func renderRhythmDanceResultsArea(stats RhythmDanceStats, selectedOption int, animFrame int) string {
	var content strings.Builder

	// Title
	content.WriteString(fmt.Sprintf("%60s\n", titleStyle.Render("RHYTHM DANCE RESULTS")))
	content.WriteString("    " + separatorStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━") + "\n")

	// Statistics
	content.WriteString(fmt.Sprintf("%51s\n", titleStyle.Render("Performance:")))
	content.WriteString("\n")

	// 判定计数
	totalJudgments := stats.PerfectCount + stats.NiceCount + stats.OKCount + stats.MissCount
	var accuracy float64
	if totalJudgments > 0 {
		accuracy = float64(stats.PerfectCount+stats.NiceCount+stats.OKCount) / float64(totalJudgments) * 100
	}

	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Perfect:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true).Render(fmt.Sprintf("%7d", stats.PerfectCount))))
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Nice:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true).Render(fmt.Sprintf("%7d", stats.NiceCount))))
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("OK:"),
		statValueStyle.Render(fmt.Sprintf("%7d", stats.OKCount))))
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Miss:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("%7d", stats.MissCount))))

	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Completed Words:"),
		statValueStyle.Render(fmt.Sprintf("%7d", stats.CompletedWords))))
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Total Score:"),
		statValueStyle.Render(fmt.Sprintf("%7d", stats.TotalScore))))
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Max Combo:"),
		statValueStyle.Render(fmt.Sprintf("%7d", stats.MaxCombo))))
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Accuracy:"),
		statValueStyle.Render(fmt.Sprintf("%6.2f%%", accuracy))))

	// Menu
	content.WriteString("\n")
	content.WriteString("    " + separatorStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━") + "\n")
	content.WriteString("\n")

	options := []string{"Restart", "Select Mode", "Main Menu"}
	selectedStyle := lipgloss.NewStyle().
		Foreground(getRandomMenuColor(animFrame)).
		Bold(true)

	for i, opt := range options {
		var optionDisplay string
		if i == selectedOption {
			optionDisplay = "> " + opt + " <"
		} else {
			optionDisplay = "  " + opt + "  "
		}

		var styledText string
		if i == selectedOption {
			styledText = selectedStyle.Render(optionDisplay)
		} else {
			styledText = menuNormalStyle.Render(optionDisplay)
		}

		alignedText := lipgloss.NewStyle().
			Width(contentWidth - 8).
			Align(lipgloss.Center).
			Render(styledText)

		content.WriteString("  " + alignedText + "\n")
	}

	return wordBoxStyle.Render(content.String())
}
