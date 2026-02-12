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
	RemainingTime   int
	TotalScore      int
	PerfectCount    int
	NiceCount       int
	OKCount         int
	MissCount       int
	CompletedWords  int
	CurrentCombo    int
	MaxCombo        int
	JudgmentHistory []string // 判定历史记录
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
	mainArea := renderRhythmMainArea(danceFrame, currentWord, userInput, rhythmBar, effectInfo, stats)
	s.WriteString(mainArea)
	s.WriteString("\n")

	// === 底部：提示 ===
	s.WriteString(hintStyle.Render("  [Space/Enter] Judge Timing  │  [ESC] Pause"))
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

// renderRhythmMainArea 渲染主游戏区域（新布局：左右分栏）
func renderRhythmMainArea(
	danceFrame string,
	currentWord string,
	userInput string,
	rhythmBar RhythmBarInfo,
	effectInfo JudgmentEffectInfo,
	stats RhythmDanceStats,
) string {
	var lines []string

	// 添加顶部空行增加高度
	lines = append(lines, "")

	// == 上部：判定历史记录（两行）==
	historyLines := renderJudgmentHistory(stats)
	lines = append(lines, historyLines...)
	lines = append(lines, "") // 空行

	// == 中部：舞蹈小人 ==
	danceLines := renderDanceCharacter(danceFrame, effectInfo.LastJudgment)
	lines = append(lines, danceLines...)
	lines = append(lines, "") // 空行

	// == 下部：左右分栏（单词+输入 ｜ 节奏条）==
	splitLines := renderSplitWordAndBar(currentWord, userInput, rhythmBar, effectInfo)
	lines = append(lines, splitLines...)

	// 添加底部空行增加高度，使其与结算界面一致（约18-20行总高度）
	lines = append(lines, "", "", "", "", "")

	return wordBoxStyle.Render(strings.Join(lines, "\n"))
}

// renderJudgmentHistory 渲染判定历史记录（两行）
// 使用彩色方块显示，按等级统计后显示
func renderJudgmentHistory(stats RhythmDanceStats) []string {
	var lines []string

	// 第一行：颜色说明
	perfectStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true) // 黄色
	niceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33"))                // 蓝色
	okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))                  // 绿色
	missStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))               // 红色

	legend := perfectStyle.Render("■ Perfect") + "  " +
		niceStyle.Render("■ Nice") + "  " +
		okStyle.Render("■ OK") + "  " +
		missStyle.Render("■ Miss")

	lines = append(lines, lipgloss.NewStyle().
		Width(contentWidth-8).
		Align(lipgloss.Center).
		Render(legend))

	// 第二行：统计各等级数量后显示
	if len(stats.JudgmentHistory) == 0 {
		lines = append(lines, lipgloss.NewStyle().
			Width(contentWidth-8).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("240")).
			Render("(empty)"))
		return lines
	}

	// 统计各等级的数量
	perfectCount := 0
	niceCount := 0
	okCount := 0
	missCount := 0

	for _, judgment := range stats.JudgmentHistory {
		switch judgment {
		case "Perfect":
			perfectCount++
		case "Nice":
			niceCount++
		case "OK":
			okCount++
		case "Miss":
			missCount++
		}
	}

	// 按顺序渲染：Perfect -> Nice -> OK -> Miss
	var barParts []string

	if perfectCount > 0 {
		barParts = append(barParts, perfectStyle.Render(strings.Repeat("■", perfectCount)))
	}
	if niceCount > 0 {
		barParts = append(barParts, niceStyle.Render(strings.Repeat("■", niceCount)))
	}
	if okCount > 0 {
		barParts = append(barParts, okStyle.Render(strings.Repeat("■", okCount)))
	}
	if missCount > 0 {
		barParts = append(barParts, missStyle.Render(strings.Repeat("■", missCount)))
	}

	historyBar := strings.Join(barParts, "")
	lines = append(lines, lipgloss.NewStyle().
		Width(contentWidth-8).
		Align(lipgloss.Center).
		Render(historyBar))

	return lines
}

// renderDanceCharacter 渲染舞蹈小人（带颜色效果）
func renderDanceCharacter(danceFrame string, lastJudgment string) []string {
	var lines []string

	// 将舞蹈帧按行分割
	frameLines := strings.Split(danceFrame, "\n")

	// 根据判定类型选择颜色
	var characterStyle lipgloss.Style
	switch lastJudgment {
	case "Perfect":
		// Perfect: 亮黄色闪烁
		characterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)
	case "Nice":
		// Nice: 绿色
		characterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46"))
	case "OK":
		// OK: 灰色
		characterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("247"))
	case "Miss":
		// Miss: 暗灰色
		characterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
	default:
		// 默认: 白色
		characterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))
	}

	// 居中显示
	for _, line := range frameLines {
		centeredLine := lipgloss.NewStyle().
			Width(contentWidth - 8).
			Align(lipgloss.Center).
			Render(characterStyle.Render(line))
		lines = append(lines, centeredLine)
	}

	// 确保始终是3行
	for len(lines) < 3 {
		lines = append(lines, "")
	}

	return lines
}

// renderSplitWordAndBar 渲染左右分栏：单词+输入（左半屏）｜节奏条（右半屏）
func renderSplitWordAndBar(
	currentWord string,
	userInput string,
	rhythmBar RhythmBarInfo,
	effectInfo JudgmentEffectInfo,
) []string {
	// 左半屏：单词和输入（2行）
	leftContent := renderCompactWordInput(currentWord, userInput)
	leftLines := strings.Split(leftContent, "\n")

	// 右半屏：节奏条（现在是多行，包含Perfect特效区）
	rightContent := renderCompactRhythmBar(rhythmBar, effectInfo)
	rightLines := strings.Split(rightContent, "\n")

	// 确保两边行数一致，补齐空行
	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	for len(leftLines) < maxLines {
		leftLines = append(leftLines, "")
	}
	for len(rightLines) < maxLines {
		rightLines = append(rightLines, "")
	}

	// 逐行合并左右内容
	var lines []string
	leftStyle := lipgloss.NewStyle().Width(contentWidth/2 - 4).Align(lipgloss.Center)
	rightStyle := lipgloss.NewStyle().Width(contentWidth/2 - 4).Align(lipgloss.Center)

	for i := 0; i < maxLines; i++ {
		leftPart := leftStyle.Render(leftLines[i])
		rightPart := rightStyle.Render(rightLines[i])
		line := leftPart + "  │  " + rightPart
		lines = append(lines, line)
	}

	return lines
}

// renderCompactWordInput 渲染紧凑的单词和输入（左半屏）
func renderCompactWordInput(currentWord string, userInput string) string {
	// 显示单词（根据输入高亮）
	var wordChars []string

	for i, ch := range currentWord {
		charStr := string(ch)

		if i < len(userInput) {
			// 已输入的部分
			if userInput[i] == byte(ch) {
				// 正确：绿色高亮
				wordChars = append(wordChars, highlightStyle.Render(charStr))
			} else {
				// 错误：红色
				errorStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("196")).
					Bold(true)
				wordChars = append(wordChars, errorStyle.Render(charStr))
			}
		} else {
			// 未输入的部分：普通样式
			wordChars = append(wordChars, wordStyle.Render(charStr))
		}
	}

	renderedWord := strings.Join(wordChars, "")

	// 输入提示（与单词对齐）
	var inputChars []string
	for i := 0; i < len(currentWord); i++ {
		if i < len(userInput) {
			// 已输入的字符
			charStr := string(userInput[i])
			if i < len(currentWord) && userInput[i] == currentWord[i] {
				// 正确：绿色
				inputChars = append(inputChars, highlightStyle.Render(charStr))
			} else {
				// 错误：红色
				errorStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("196")).
					Bold(true)
				inputChars = append(inputChars, errorStyle.Render(charStr))
			}
		} else {
			// 未输入的位置：显示下划线占位符
			inputChars = append(inputChars, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("_"))
		}
	}

	inputDisplay := strings.Join(inputChars, "")

	return renderedWord + "\n" + inputDisplay
}

// renderCompactRhythmBar 渲染紧凑的节奏条（右半屏）
// Perfect特效显示在节奏条上方，保持固定高度
func renderCompactRhythmBar(rhythmBar RhythmBarInfo, effectInfo JudgmentEffectInfo) string {
	var lines []string

	// 节奏条宽度
	const barWidth = 35

	// 计算指针和黄金点的位置
	pointerPos := int(rhythmBar.PointerPosition * float64(barWidth))
	goldenPos := int(rhythmBar.GoldenRatio * float64(barWidth))

	if pointerPos < 0 {
		pointerPos = 0
	}
	if pointerPos >= barWidth {
		pointerPos = barWidth - 1
	}

	// 第一部分：Perfect特效区域（固定高度2行）
	// DEBUG: 直接显示特效测试
	if effectInfo.LastJudgment == "Perfect" {
		elapsed := time.Since(effectInfo.LastJudgmentTime)
		if elapsed < 1*time.Second {
			perfectEffect := renderPerfectExplosion(elapsed)
			lines = append(lines, perfectEffect...)
		} else {
			// 特效结束，显示空行保持高度
			lines = append(lines, "", "")
		}
	} else if effectInfo.LastJudgment != "" {
		// DEBUG: 显示当前判定类型
		debugStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		lines = append(lines, debugStyle.Render("Last: "+effectInfo.LastJudgment), "")
	} else {
		// 非Perfect判定，显示空行保持高度
		lines = append(lines, "", "")
	}

	// 第二部分：节奏条
	var barChars []string
	for i := 0; i < barWidth; i++ {
		// 计算距离黄金点的距离
		distance := math.Abs(float64(i)/float64(barWidth) - rhythmBar.GoldenRatio)

		// 根据距离选择颜色
		var cellStyle lipgloss.Style
		if distance <= 0.05 {
			cellStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // 亮黄
		} else if distance <= 0.15 {
			cellStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46")) // 绿色
		} else if distance <= 0.30 {
			cellStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("147")) // 浅紫
		} else {
			cellStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // 暗灰
		}

		// 指针位置显示 "▲"，其他位置显示 "━"
		if i == pointerPos {
			barChars = append(barChars, lipgloss.NewStyle().
				Foreground(lipgloss.Color("201")).
				Bold(true).
				Render("▲"))
		} else if i == goldenPos {
			barChars = append(barChars, cellStyle.Render("◆")) // 黄金点标记
		} else {
			barChars = append(barChars, cellStyle.Render("━"))
		}
	}

	bar := strings.Join(barChars, "")
	lines = append(lines, bar)

	// 第三部分：其他判定特效（Nice/OK/Miss）（固定高度1行）
	if effectInfo.LastJudgment != "" && effectInfo.LastJudgment != "Perfect" {
		elapsed := time.Since(effectInfo.LastJudgmentTime)
		if elapsed < 1*time.Second {
			effect := renderJudgmentEffect(effectInfo.LastJudgment, elapsed)
			lines = append(lines, effect)
		} else {
			lines = append(lines, "")
		}
	} else {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderPerfectExplosion 渲染Perfect爆炸闪烁特效（2行）
func renderPerfectExplosion(elapsed time.Duration) []string {
	ms := elapsed.Milliseconds()

	// 闪烁效果：每100ms切换颜色
	frame := int(ms / 100)
	var color lipgloss.Color
	if frame%2 == 0 {
		color = lipgloss.Color("226") // 亮黄
	} else {
		color = lipgloss.Color("220") // 浅黄
	}

	style := lipgloss.NewStyle().Foreground(color).Bold(true)

	// 爆炸动画帧
	if ms < 300 {
		// 第1阶段：爆炸扩散
		return []string{
			style.Render("   ✦ ★ PERFECT! ★ ✦   "),
			style.Render("    ✧    ✧    ✧    "),
		}
	} else if ms < 600 {
		// 第2阶段：火花四溅
		return []string{
			style.Render(" ✧ ★ PERFECT! ★ ✧ "),
			style.Render("  ✦  ✧  ✦  ✧  ✦  "),
		}
	} else {
		// 第3阶段：逐渐消失
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
		return []string{
			dimStyle.Render("   ★ PERFECT! ★   "),
			dimStyle.Render("                   "),
		}
	}
}

// renderWordAndRhythmBar 渲染单词和节奏条（旧版，保留以兼容）
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
