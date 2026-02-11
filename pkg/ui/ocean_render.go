package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/word-killer/word-killer/pkg/game"
)

const (
	oceanContentWidth = 72 // 内容区域宽度（不含边框）
	oceanHeight       = 20 // 海洋场景高度（增加到20行）
)

// RenderUnderwaterGame 渲染海底世界游戏界面
func RenderUnderwaterGame(g *game.Game) string {
	if g.UnderwaterState == nil {
		return "海底世界初始化中..."
	}

	var sections []string

	// 1. 顶部状态栏（倒计时、统计）
	sections = append(sections, renderUnderwaterStatus(g))

	// 2. 海洋场景（主要游戏区域）
	sections = append(sections, renderOceanScene(g.UnderwaterState, g.InputBuffer))

	// 3. 输入提示
	sections = append(sections, renderUnderwaterInput(g.InputBuffer))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderUnderwaterStatus 渲染状态栏
func renderUnderwaterStatus(g *game.Game) string {
	remaining := g.GetRemainingTime()
	minutes := remaining / 60
	seconds := remaining % 60

	// Countdown display
	timeStr := fmt.Sprintf("Time: %02d:%02d", minutes, seconds)
	var timeStyle lipgloss.Style
	if remaining <= 10 {
		// Last 10 seconds - red warning
		timeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	} else if remaining <= 30 {
		// Under 30 seconds - yellow alert
		timeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	} else {
		// Normal display
		timeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBubbleCyan)).Bold(true)
	}

	// Statistics
	stats := g.Stats
	statsStr := fmt.Sprintf("Fish: %d  Keys: %d  Accuracy: %.1f%%",
		stats.WordsCompleted,
		stats.TotalKeystrokes,
		stats.GetAccuracyPercent(),
	)
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorLightBlue))

	// Combine status bar
	statusLeft := timeStyle.Render(timeStr)
	statusRight := statsStyle.Render(statsStr)

	// Calculate spacing
	leftWidth := lipgloss.Width(statusLeft)
	rightWidth := lipgloss.Width(statusRight)
	spacing := contentWidth - leftWidth - rightWidth - 4 // Subtract margins

	if spacing < 1 {
		spacing = 1
	}

	statusLine := "  " + statusLeft + strings.Repeat(" ", spacing) + statusRight

	return statusLine + "\n"
}

// renderOceanScene 渲染海洋场景（核心渲染函数）
func renderOceanScene(state *game.UnderwaterState, input string) string {
	// 创建字符网格 (10行 × 72列)
	grid := make([][]rune, oceanHeight)
	for i := range grid {
		grid[i] = make([]rune, oceanContentWidth)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// 分层渲染（从后往前）
	renderBackgroundElements(grid, state) // 1. 背景（海藻、珊瑚）
	renderBubbles(grid, state)            // 2. 气泡
	renderFishes(grid, state, input)      // 3. 小鱼和单词

	// Convert to colored strings
	lines := renderGridWithColors(grid, state, input)

	return wordBoxStyle.Render(strings.Join(lines, "\n"))
}

// renderBackgroundElements 渲染背景元素（珊瑚、海草、贝壳、海星）
func renderBackgroundElements(grid [][]rune, state *game.UnderwaterState) {
	gridHeight := len(grid)
	gridWidth := len(grid[0])
	bottomRow := gridHeight - 1

	// === 珊瑚礁（多行组合）===
	// 珊瑚礁1 - 左侧 (位置 10-14)
	if bottomRow >= 0 && bottomRow-2 >= 0 {
		grid[bottomRow][12] = '※'   // 底部中心
		grid[bottomRow][11] = '◊'   // 左侧
		grid[bottomRow][13] = '◇'   // 右侧
		grid[bottomRow-1][12] = 'Ψ' // 上层
		grid[bottomRow-1][11] = '°' // 气泡
	}

	// 珊瑚礁2 - 中央 (位置 35-40)
	if bottomRow >= 0 && bottomRow-3 >= 0 {
		grid[bottomRow][37] = '※'   // 底部中心
		grid[bottomRow][36] = '◈'   // 左
		grid[bottomRow][38] = '⟡'   // 右
		grid[bottomRow-1][37] = '✿' // 中层
		grid[bottomRow-1][36] = 'ω' // 左中
		grid[bottomRow-1][38] = 'Ψ' // 右中
		grid[bottomRow-2][37] = '°' // 顶部气泡
	}

	// 珊瑚礁3 - 右侧 (位置 58-62)
	if bottomRow >= 0 && bottomRow-2 >= 0 {
		grid[bottomRow][60] = '◆'   // 底部
		grid[bottomRow][59] = '◊'   // 左
		grid[bottomRow][61] = '◇'   // 右
		grid[bottomRow-1][60] = '※' // 上层
		grid[bottomRow-1][61] = '°' // 气泡
	}

	// === 海草（摇曳的感觉）===
	seaGrassPositions := []int{8, 18, 28, 48, 68}
	for i, x := range seaGrassPositions {
		if x >= gridWidth {
			continue
		}
		// 不同高度的海草
		height := 2 + (i % 2) // 2或3行高
		for h := 0; h < height; h++ {
			row := bottomRow - h
			if row >= 0 {
				if h == 0 {
					grid[row][x] = '|' // 底部
				} else if h == height-1 {
					grid[row][x] = ')' // 顶部
				} else {
					grid[row][x] = '|' // 中间
				}
			}
		}
	}

	// === 贝壳和海星（散落在海底）===
	// 贝壳
	shellDecorations := []struct {
		x    int
		char rune
	}{
		{5, '◊'},
		{22, '◇'},
		{44, '◈'},
		{52, '◊'},
		{65, '◇'},
	}
	for _, shell := range shellDecorations {
		if shell.x < gridWidth && bottomRow >= 0 {
			grid[bottomRow][shell.x] = shell.char
		}
	}

	// 海星
	starPositions := []int{15, 33, 55}
	for _, x := range starPositions {
		if x < gridWidth && bottomRow >= 0 {
			grid[bottomRow][x] = '✦'
		}
	}

	// === Top waves (dynamic waves with periodic speed) ===
	for y := 0; y < 2; y++ {
		for x := 0; x < gridWidth; x++ {
			// Use sine wave to create periodic speed variation
			// Speed varies dramatically between slow and fast over time
			speedFactor := math.Sin(float64(state.BackgroundFrame)/30.0)*1.2 + 1.3 // Range: 0.1 to 2.5
			effectiveFrame := int(float64(state.BackgroundFrame) * speedFactor)

			offset := (effectiveFrame / 3) % 6
			pattern := (x + offset) % 6

			if y == 0 {
				// First row
				if pattern == 0 || pattern == 1 {
					grid[y][x] = '~'
				} else if pattern == 3 || pattern == 4 {
					grid[y][x] = '≈'
				}
			} else {
				// Second row, reverse pattern
				if pattern == 2 || pattern == 3 {
					grid[y][x] = '≈'
				} else if pattern == 5 || pattern == 0 {
					grid[y][x] = '~'
				}
			}
		}
	}
}

// renderBubbles 渲染气泡
func renderBubbles(grid [][]rune, state *game.UnderwaterState) {
	bubbleChars := []rune{BubbleSmall, BubbleMedium, BubbleLarge}

	for i, bubble := range state.BubbleStreams {
		if !bubble.Active {
			continue
		}

		x := bubble.X
		y := int(bubble.Y)

		if x >= 0 && x < len(grid[0]) && y >= 0 && y < len(grid) {
			// 根据气泡流编号选择不同大小的气泡
			char := bubbleChars[i%len(bubbleChars)]
			grid[y][x] = char
		}
	}
}

// renderFishes 渲染小鱼和单词
func renderFishes(grid [][]rune, state *game.UnderwaterState, input string) {
	// 检测重叠并标记需要隐藏的小鱼
	hiddenFishIndices := make(map[int]bool)
	for i := range state.Fishes {
		fish1 := &state.Fishes[i]
		if fish1.Completed && !fish1.Glowing {
			continue
		}

		for j := i + 1; j < len(state.Fishes); j++ {
			fish2 := &state.Fishes[j]
			if fish2.Completed && !fish2.Glowing {
				continue
			}

			// Check if on the same row
			if fish1.Y == fish2.Y {
				// Calculate X ranges for both fish
				// Both directions now use 4 extra chars: left=◀□word-◁, right=▷-word□▶
				x1Start := int(fish1.X * float64(len(grid[0])))
				x1End := x1Start + len(fish1.Word) + 4

				x2Start := int(fish2.X * float64(len(grid[0])))
				x2End := x2Start + len(fish2.Word) + 4

				// Check if ranges overlap
				if x1Start < x2End && x2Start < x1End {
					// Overlap detected, hide the second fish
					hiddenFishIndices[j] = true
				}
			}
		}
	}

	// 渲染小鱼
	for i, fish := range state.Fishes {
		// 跳过已完成且发光动画结束的小鱼
		if fish.Completed && !fish.Glowing {
			if time.Since(fish.CompletedAt).Milliseconds() > 800 {
				continue
			}
		}

		// 跳过被隐藏的小鱼
		if hiddenFishIndices[i] {
			continue
		}

		// 计算屏幕位置
		xPos := int(fish.X * float64(len(grid[0])))
		yPos := fish.Y

		// Render fish based on direction
		if yPos >= 0 && yPos < len(grid) {
			if fish.Direction == 1 {
				// Right: ▷-word□▶
				renderFishPart(grid, yPos, xPos, "▷-")
				renderFishPart(grid, yPos, xPos+2, fish.Word)
				renderFishPart(grid, yPos, xPos+2+len(fish.Word), "□▶")
			} else {
				// Left: ◀□word-◁
				renderFishPart(grid, yPos, xPos, "◀□")
				renderFishPart(grid, yPos, xPos+2, fish.Word)
				renderFishPart(grid, yPos, xPos+2+len(fish.Word), "-◁")
			}
		}
	}
}

// renderFishPart renders a part of the fish (head/body/tail)
func renderFishPart(grid [][]rune, y, xStart int, part string) {
	runes := []rune(part)
	for i, ch := range runes {
		x := xStart + i
		if x >= 0 && x < len(grid[0]) && y >= 0 && y < len(grid) {
			grid[y][x] = ch
		}
	}
}

// renderGridWithColors converts grid to colored strings with match highlighting
func renderGridWithColors(grid [][]rune, state *game.UnderwaterState, input string) []string {
	lines := make([]string, len(grid))

	// Predefined styles
	grassStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSeaweedGreen))
	coralStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorCoralPink))
	shellStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("222")) // Light pink
	starStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))  // Orange-yellow
	bubbleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBubbleCyan))
	waveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorLightBlue))
	wordStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Bold(true)       // Bright white
	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true) // Yellow highlight for matched chars

	// Three shades of blue for fish
	fishStyles := []lipgloss.Style{
		lipgloss.NewStyle().Foreground(lipgloss.Color(ColorFishLightBlue)),
		lipgloss.NewStyle().Foreground(lipgloss.Color(ColorFishBlue)),
		lipgloss.NewStyle().Foreground(lipgloss.Color(ColorFishDarkBlue)),
	}

	for y, row := range grid {
		var sb strings.Builder
		sb.WriteString("  ") // 左边距

		for x, ch := range row {
			if ch == ' ' {
				sb.WriteRune(ch)
				continue
			}

			// 优先检查是否是小鱼中的内容（单词、身体）
			var styled string
			inFishWord := false
			inFishBody := false
			var fishIdx int
			var targetFish game.Fish

			// Check if within any fish range
			for idx, fish := range state.Fishes {
				fishX := int(fish.X * float64(len(grid[0])))
				fishY := fish.Y
				fishWidth := len(fish.Word) + 4 // Updated for new fish design

				if y == fishY && x >= fishX && x < fishX+fishWidth {
					fishIdx = idx
					targetFish = fish

					// Determine if it's a word or body character
					var wordStart int
					if fish.Direction == 1 {
						// Right: ▷-word□▶
						wordStart = fishX + 2
					} else {
						// Left: ◀□word-◁
						wordStart = fishX + 2
					}
					wordEnd := wordStart + len(fish.Word)
					if x >= wordStart && x < wordEnd {
						inFishWord = true
					} else {
						inFishBody = true
					}
					break
				}
			}

			// If in fish word part
			if inFishWord {
				if targetFish.Completed && targetFish.Glowing {
					styled = renderFishGlowChar(targetFish, string(ch))
				} else {
					// Calculate character position in word
					var wordStart int
					if targetFish.Direction == 1 {
						wordStart = int(targetFish.X * float64(len(grid[0]))) + 2
					} else {
						wordStart = int(targetFish.X * float64(len(grid[0]))) + 2
					}
					charPos := x - wordStart

					// Check if this character should be highlighted (matches input)
					if !targetFish.Completed && len(input) > 0 &&
					   strings.HasPrefix(targetFish.Word, input) &&
					   charPos < len(input) {
						// Matched character - use highlight style
						styled = highlightStyle.Render(string(ch))
					} else {
						// Normal word character
						styled = wordStyle.Render(string(ch))
					}
				}
			} else if inFishBody {
				// 在小鱼的身体部分
				fishStyle := fishStyles[fishIdx%len(fishStyles)]
				if targetFish.Completed && targetFish.Glowing {
					styled = renderFishGlowChar(targetFish, string(ch))
				} else {
					styled = fishStyle.Render(string(ch))
				}
			} else {
				// 不在小鱼中，根据字符类型判断
				switch ch {
				case '|', '\\', ')':
					// 海草
					styled = grassStyle.Render(string(ch))
				case '※', 'Ψ', 'ω', '✿', '◆':
					// 珊瑚
					styled = coralStyle.Render(string(ch))
				case '◊', '◇', '◈', '⟡':
					// 贝壳
					styled = shellStyle.Render(string(ch))
				case '✦':
					// 海星
					styled = starStyle.Render(string(ch))
				case '°', 'o', 'O', '∘':
					// 气泡
					styled = bubbleStyle.Render(string(ch))
				case '~', '≈':
					// 海浪
					styled = waveStyle.Render(string(ch))
				default:
					// 其他字符
					styled = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(string(ch))
				}
			}

			sb.WriteString(styled)
		}

		lines[y] = sb.String()
	}

	return lines
}

// renderFishGlowChar 渲染发光小鱼字符
func renderFishGlowChar(fish game.Fish, char string) string {
	msElapsed := time.Since(fish.CompletedAt).Milliseconds()

	// 阶段1：亮光脉冲 (0-100ms)
	if msElapsed < 50 {
		// 亮青闪烁
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorGlowCyan)).
			Bold(true).Render(char)
	} else if msElapsed < 100 {
		// 白色闪烁
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("231")).
			Bold(true).Render(char)
	}

	// 阶段2：逐渐变暗
	if msElapsed < 200 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorGlowCyan)).
			Bold(true).Render(char)
	} else if msElapsed < 400 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorFishLightBlue)).
			Render(char)
	} else if msElapsed < 600 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorLightBlue)).
			Render(char)
	}

	// 淡出
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(char)
}

// renderUnderwaterInput 渲染输入提示
func renderUnderwaterInput(input string) string {
	prompt := "Type to catch fish: "
	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorBubbleCyan)).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("231")).
		Bold(true)

	return "\n" + promptStyle.Render(prompt) + inputStyle.Render(input+"_")
}
