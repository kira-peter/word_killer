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
	LastJudgment         string
	LastJudgmentTime     time.Time
	LastJudgmentPosition float64 // 上次判定的指针位置（用于显示箭头）
}

// RenderRhythmDanceGame 渲染节奏舞蹈模式主界面
func RenderRhythmDanceGame(
	danceFrame string,
	wordQueue []string,   // 单词队列（固定长度5）
	currentWordIndex int, // 当前单词索引（固定为2）
	userInput string,
	rhythmBar RhythmBarInfo,
	stats RhythmDanceStats,
	effectInfo JudgmentEffectInfo,
) string {
	var s strings.Builder

	// === 顶部：统计信息 ===
	s.WriteString(renderRhythmStats(stats))
	s.WriteString("\n")

	// === 中部：主游戏区域（舞蹈小人 + 单词队列 + 节奏条）===
	mainArea := renderRhythmMainArea(danceFrame, wordQueue, currentWordIndex, userInput, rhythmBar, effectInfo, stats)
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

// renderRhythmMainArea 渲染主游戏区域（新布局：上下布局，单词队列在中间）
func renderRhythmMainArea(
	danceFrame string,
	wordQueue []string,
	currentWordIndex int,
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

	// == 中上部：舞蹈小人 ==
	danceLines := renderDanceCharacter(danceFrame, effectInfo.LastJudgment)
	lines = append(lines, danceLines...)
	lines = append(lines, "") // 空行

	// == 中部：左右分栏（单词队列 ｜ 节奏条）==
	splitLines := renderWordQueueAndBar(wordQueue, currentWordIndex, userInput, rhythmBar, effectInfo)
	lines = append(lines, splitLines...)

	// 添加底部空行增加高度，使其与结算界面一致（约18-20行总高度）
	lines = append(lines, "", "")

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

	// 根据判定类型选择颜色（与判定颜色统一）
	var characterStyle lipgloss.Style
	switch lastJudgment {
	case "Perfect":
		// Perfect: 金色
		characterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)
	case "Nice":
		// Nice: 蓝色
		characterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("33"))
	case "OK":
		// OK: 绿色
		characterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46"))
	case "Miss":
		// Miss: 红色
		characterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
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

// renderWordQueueAndBar 渲染左右分栏：单词队列（左半屏，靠右对齐）｜ 节奏条（右半屏，仅显示在中间行）
func renderWordQueueAndBar(
	wordQueue []string,
	currentWordIndex int,
	userInput string,
	rhythmBar RhythmBarInfo,
	effectInfo JudgmentEffectInfo,
) []string {
	var lines []string

	// 确保单词队列长度为5
	if len(wordQueue) != 5 {
		// 如果长度不对，返回错误提示
		return []string{"Error: WordQueue length must be 5"}
	}

	// 左半屏宽度和右半屏宽度
	const leftWidth = contentWidth/2 - 4
	const rightWidth = contentWidth/2 - 4

	// 颜色配置（从上到下：深灰 → 浅灰 → 白色 → 浅灰 → 深灰）
	colors := []string{"#444444", "#888888", "#FFFFFF", "#888888", "#444444"}

	// 渲染节奏条内容（包含 Perfect 特效、节奏条、箭头、其他判定特效），共5行
	// 结构：[0-1: Perfect特效2行, 2: 节奏条主体1行, 3: 箭头1行, 4: 其他判定1行]
	rhythmBarLines := renderRhythmBarWithEffects(rhythmBar, effectInfo)

	// 首先渲染节奏条的前2行（Perfect特效），左侧空白
	for j := 0; j < 2; j++ {
		leftPart := lipgloss.NewStyle().Width(leftWidth).Render("")
		rightPart := lipgloss.NewStyle().Width(rightWidth).Align(lipgloss.Center).Render(rhythmBarLines[j])
		line := leftPart + "  │  " + rightPart
		lines = append(lines, line)
	}

	// 渲染单词队列的前3行（索引0-2），索引2是当前单词，右侧显示节奏条主体
	for i := 0; i < 3; i++ {
		word := wordQueue[i]
		color := colors[i]

		var leftContent string

		if i == currentWordIndex {
			// 当前行（索引2）：显示 "Input:[完整单词]"，带颜色编码
			leftContent = renderCurrentWordWithInput(word, userInput)
		} else {
			// 其他行：使用对齐格式显示单词，保持与 Input:[...] 相同宽度
			leftContent = renderWordAligned(word, color)
		}

		// 左半屏：单词靠右对齐
		leftPart := lipgloss.NewStyle().Width(leftWidth).Align(lipgloss.Right).Render(leftContent)

		// 右半屏：仅在当前行（索引2）显示节奏条主体（rhythmBarLines[2]）
		var rightPart string
		if i == currentWordIndex {
			// 显示节奏条主体（索引2）
			rightPart = lipgloss.NewStyle().Width(rightWidth).Align(lipgloss.Center).Render(rhythmBarLines[2])
		} else {
			// 其他行：空白
			rightPart = lipgloss.NewStyle().Width(rightWidth).Render("")
		}

		// 合并左右两侧，当前行使用 ━ 连接，其他行使用 │
		var separator string
		if i == currentWordIndex {
			// 当前行：使用深灰色 ━ 连接单词与节奏条
			separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // 深灰色
			separator = separatorStyle.Render("━━━━━")
		} else {
			separator = "  │  " // 其他行：使用 │ 分隔
		}
		line := leftPart + separator + rightPart
		lines = append(lines, line)
	}

	// 渲染单词行3 + 箭头行 - 左侧显示单词队列索引3，右侧显示箭头
	word3 := wordQueue[3]
	color3 := colors[3]
	leftContent3 := renderWordAligned(word3, color3)
	leftPart := lipgloss.NewStyle().Width(leftWidth).Align(lipgloss.Right).Render(leftContent3)
	rightPart := lipgloss.NewStyle().Width(rightWidth).Align(lipgloss.Center).Render(rhythmBarLines[3])
	line := leftPart + "  │  " + rightPart
	lines = append(lines, line)

	// 渲染单词队列的最后1行（索引4）
	word4 := wordQueue[4]
	color4 := colors[4]
	leftContent4 := renderWordAligned(word4, color4)
	leftPart = lipgloss.NewStyle().Width(leftWidth).Align(lipgloss.Right).Render(leftContent4)
	rightPart = lipgloss.NewStyle().Width(rightWidth).Render("")
	line = leftPart + "  │  " + rightPart
	lines = append(lines, line)

	// 其他判定特效行（rhythmBarLines[4]）
	leftPart = lipgloss.NewStyle().Width(leftWidth).Render("")
	rightPart = lipgloss.NewStyle().Width(rightWidth).Align(lipgloss.Center).Render(rhythmBarLines[4])
	line = leftPart + "  │  " + rightPart
	lines = append(lines, line)

	return lines
}

// renderRhythmBarWithEffects 渲染完整的节奏条内容（包含判定特效、节奏条、箭头）
// 返回固定5行内容：[上方特效2行, 节奏条1行, 箭头1行, 下方特效1行]
func renderRhythmBarWithEffects(rhythmBar RhythmBarInfo, effectInfo JudgmentEffectInfo) []string {
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

	// 根据判定类型渲染特效
	var topEffect []string  // 上方特效（2行）
	var bottomEffect string // 下方特效（1行）

	if effectInfo.LastJudgment != "" {
		elapsed := time.Since(effectInfo.LastJudgmentTime)
		if elapsed < 1*time.Second {
			topEffect, bottomEffect = renderJudgmentEffects(effectInfo.LastJudgment, elapsed)
		} else {
			topEffect = []string{"", ""}
			bottomEffect = ""
		}
	} else {
		topEffect = []string{"", ""}
		bottomEffect = ""
	}

	// 添加上方特效（2行）
	lines = append(lines, topEffect...)

	// 渲染节奏条主体（1行）
	var barChars []string
	for i := 0; i < barWidth; i++ {
		// 计算距离黄金点的字符数（绝对距离）
		charDistance := int(math.Abs(float64(i - goldenPos)))

		// 根据字符距离选择颜色，与判定颜色一致
		// 模式: --4433223344--
		//       Miss OK Nice Perfect Nice OK Miss
		var cellStyle lipgloss.Style
		if charDistance == 0 {
			// Perfect 区域（黄金点本身）：金色
			cellStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // 金色
		} else if charDistance <= 2 {
			// Nice 区域（1-2个字符距离）：蓝色
			cellStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("33")) // 蓝色
		} else if charDistance <= 4 {
			// OK 区域（3-4个字符距离）：绿色
			cellStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46")) // 绿色
		} else {
			// Miss 区域（>4个字符距离）：红色
			cellStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // 红色
		}

		// 黄金点位置显示 "|"，指针位置显示 "▲"，其他位置显示 "━"
		if i == goldenPos {
			barChars = append(barChars, lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")).
				Bold(true).
				Render("│")) // 黄金点标记为 |
		} else if i == pointerPos {
			barChars = append(barChars, lipgloss.NewStyle().
				Foreground(lipgloss.Color("201")).
				Bold(true).
				Render("▲"))
		} else {
			barChars = append(barChars, cellStyle.Render("━"))
		}
	}

	bar := strings.Join(barChars, "")
	lines = append(lines, bar)

	// 添加箭头行（指向上次判定位置，渐渐消失）- 紧贴节奏条
	arrowLine := renderJudgmentArrow(effectInfo, barWidth)
	lines = append(lines, arrowLine)

	// 添加下方特效（1行）
	lines = append(lines, bottomEffect)

	return lines
}

// renderJudgmentArrow 渲染指向上次判定位置的箭头，随时间渐渐消失
func renderJudgmentArrow(effectInfo JudgmentEffectInfo, barWidth int) string {
	if effectInfo.LastJudgment == "" {
		return "" // 没有判定记录，不显示箭头
	}

	elapsed := time.Since(effectInfo.LastJudgmentTime)
	const fadeOutDuration = 1500 * time.Millisecond // 箭头在1.5秒内消失

	if elapsed >= fadeOutDuration {
		return "" // 已经消失
	}

	// 计算上次判定位置
	lastJudgmentPos := int(effectInfo.LastJudgmentPosition * float64(barWidth))
	if lastJudgmentPos < 0 {
		lastJudgmentPos = 0
	}
	if lastJudgmentPos >= barWidth {
		lastJudgmentPos = barWidth - 1
	}

	// 构建箭头行：在判定位置显示 "^"，其他位置显示空格
	var arrowChars []string
	for i := 0; i < barWidth; i++ {
		if i == lastJudgmentPos {
			// 根据判定类型选择箭头颜色
			var arrowColor lipgloss.Color
			switch effectInfo.LastJudgment {
			case "Perfect":
				arrowColor = lipgloss.Color("226") // 金色
			case "Nice":
				arrowColor = lipgloss.Color("33") // 蓝色
			case "OK":
				arrowColor = lipgloss.Color("46") // 绿色
			case "Miss":
				arrowColor = lipgloss.Color("196") // 红色
			default:
				arrowColor = lipgloss.Color("255") // 白色
			}

			// 计算透明度（通过颜色亮度模拟淡出效果）
			// elapsed: 0 -> fadeOutDuration, 箭头从亮到暗
			fadeRatio := float64(elapsed) / float64(fadeOutDuration)

			// 根据淡出程度选择箭头样式
			style := lipgloss.NewStyle().Foreground(arrowColor)
			if fadeRatio < 0.5 {
				style = style.Bold(true) // 前半段加粗
			}

			arrowChars = append(arrowChars, style.Render("^"))
		} else {
			arrowChars = append(arrowChars, " ")
		}
	}

	return strings.Join(arrowChars, "")
}

// renderCurrentWordWithInput 渲染当前单词行，显示 "Input:[    word]" 固定宽度20字符，带颜色编码
func renderCurrentWordWithInput(targetWord string, userInput string) string {
	const wordFieldWidth = 20 // 单词显示区域固定宽度

	// 前缀 "Input:[" - 深灰色
	prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // 深灰色
	prefix := prefixStyle.Render("Input:[")

	// 渲染目标单词的每个字符，根据用户输入进行颜色编码
	var wordChars []string

	for i, ch := range targetWord {
		charStr := string(ch)

		if i < len(userInput) {
			// 有对应的输入字符
			if userInput[i] == byte(ch) {
				// 正确：绿色
				wordChars = append(wordChars, lipgloss.NewStyle().
					Foreground(lipgloss.Color("46")).
					Bold(true).
					Render(charStr))
			} else {
				// 错误：红色
				wordChars = append(wordChars, lipgloss.NewStyle().
					Foreground(lipgloss.Color("196")).
					Bold(true).
					Render(charStr))
			}
		} else {
			// 未输入的字符：白色
			wordChars = append(wordChars, lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Render(charStr))
		}
	}

	// 将渲染后的单词字符拼接
	renderedWord := strings.Join(wordChars, "")

	// 计算需要补齐的空格数（单词左侧填充空格，使总宽度为 wordFieldWidth）
	paddingCount := wordFieldWidth - len(targetWord)
	if paddingCount < 0 {
		paddingCount = 0 // 如果单词超长，不补齐
	}
	padding := strings.Repeat(" ", paddingCount)

	// 后缀 "]" - 深灰色
	suffix := prefixStyle.Render("]")

	return prefix + padding + renderedWord + suffix
}

// renderWordAligned 渲染对齐的单词（非当前行），保持与 Input:[    word] 相同的格式宽度
func renderWordAligned(word string, color string) string {
	const wordFieldWidth = 20 // 单词显示区域固定宽度
	const prefixWidth = 7     // "Input:[" 的长度
	const suffixWidth = 1     // "]" 的长度

	if word == "" {
		// 空字符串：返回空格填充，保持格式宽度
		return strings.Repeat(" ", prefixWidth+wordFieldWidth+suffixWidth)
	}

	// 前缀：7 个空格（对应 "Input:[" 的位置）
	prefix := strings.Repeat(" ", prefixWidth)

	// 计算单词左侧填充（右对齐到 wordFieldWidth）
	paddingCount := wordFieldWidth - len(word)
	if paddingCount < 0 {
		paddingCount = 0
	}
	padding := strings.Repeat(" ", paddingCount)

	// 渲染单词（带颜色）
	wordStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	renderedWord := wordStyle.Render(word)

	// 后缀：1 个空格（对应 "]" 的位置）
	suffix := strings.Repeat(" ", suffixWidth)

	return prefix + padding + renderedWord + suffix
}

// renderJudgmentEffects 渲染判定特效（上下对称设计）
// 返回：上方特效（2行）和下方特效（1行）
func renderJudgmentEffects(judgment string, elapsed time.Duration) ([]string, string) {
	ms := elapsed.Milliseconds()

	switch judgment {
	case "Perfect":
		return renderPerfectEffect(ms)
	case "Nice":
		return renderNiceEffect(ms)
	case "OK":
		return renderOKEffect(ms)
	case "Miss":
		return renderMissEffect(ms)
	default:
		return []string{"", ""}, ""
	}
}

// renderPerfectEffect 渲染 Perfect 特效（金色星星，上下闪烁）
func renderPerfectEffect(ms int64) ([]string, string) {
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
		}, style.Render("    ✧    ✧    ✧    ")
	} else if ms < 600 {
		// 第2阶段：火花四溅
		return []string{
			style.Render(" ✧ ★ PERFECT! ★ ✧ "),
			style.Render("  ✦  ✧  ✦  ✧  ✦  "),
		}, style.Render("  ✦  ✧  ✦  ✧  ✦  ")
	} else {
		// 第3阶段：逐渐消失
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
		return []string{
			dimStyle.Render("   ★ PERFECT! ★   "),
			dimStyle.Render("      ✧  ✧      "),
		}, dimStyle.Render("      ✧  ✧      ")
	}
}

// renderNiceEffect 渲染 Nice 特效（蓝色波纹）
func renderNiceEffect(ms int64) ([]string, string) {
	// 使用统一的蓝色（33）
	color := lipgloss.Color("33")
	style := lipgloss.NewStyle().Foreground(color).Bold(true)

	if ms < 300 {
		return []string{
			style.Render("     ～ Nice! ～     "),
			style.Render("    ≈  ≈  ≈  ≈    "),
		}, style.Render("    ≈  ≈  ≈  ≈    ")
	} else if ms < 600 {
		return []string{
			style.Render("      Nice!      "),
			style.Render("     ≈  ≈  ≈     "),
		}, style.Render("     ≈  ≈  ≈     ")
	} else {
		dimStyle := lipgloss.NewStyle().Foreground(color)
		return []string{
			dimStyle.Render("      Nice!      "),
			dimStyle.Render("       ≈  ≈       "),
		}, dimStyle.Render("       ≈  ≈       ")
	}
}

// renderOKEffect 渲染 OK 特效（绿色勾号）
func renderOKEffect(ms int64) ([]string, string) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))

	if ms < 400 {
		return []string{
			style.Render("       ✓ OK ✓       "),
			style.Render("                   "),
		}, style.Render("                   ")
	} else {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		return []string{
			dimStyle.Render("       ✓ OK ✓       "),
			dimStyle.Render("                   "),
		}, dimStyle.Render("                   ")
	}
}

// renderMissEffect 渲染 Miss 特效（红色警告，上下闪烁）
func renderMissEffect(ms int64) ([]string, string) {
	// 闪烁效果
	frame := int(ms / 150)
	var color lipgloss.Color
	if frame%2 == 0 {
		color = lipgloss.Color("196") // 亮红
	} else {
		color = lipgloss.Color("160") // 深红
	}

	style := lipgloss.NewStyle().Foreground(color).Bold(true)

	if ms < 300 {
		return []string{
			style.Render("    ✗✗ MISS! ✗✗    "),
			style.Render("    ！  ！  ！    "),
		}, style.Render("    ！  ！  ！    ")
	} else if ms < 600 {
		return []string{
			style.Render("     ✗ MISS! ✗     "),
			style.Render("      ！  ！      "),
		}, style.Render("      ！  ！      ")
	} else {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("160"))
		return []string{
			dimStyle.Render("       MISS!       "),
			dimStyle.Render("                   "),
		}, dimStyle.Render("                   ")
	}
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
		lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true).Render(fmt.Sprintf("%7d", stats.NiceCount))))
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
