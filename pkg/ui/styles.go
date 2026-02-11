package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
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

	// Completed word styles
	completedWordStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Strikethrough(true)

	// Hit effect style (bright flash)
	hitEffectStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")). // Bright yellow
		Bold(true)

	// Fading styles (for gradual transition)
	fadingStyle1 = lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")). // Bright green
		Bold(true)

	fadingStyle2 = lipgloss.NewStyle().
		Foreground(lipgloss.Color("82")) // Medium green

	fadingStyle3 = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")) // Light gray

	fadingStyle4 = lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")) // Medium gray

	fadingStyle5 = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")) // Dark gray (final)

	// New styles for professional layout
	headerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Background(lipgloss.Color("235")).
		Padding(0, 2).
		MarginBottom(1)

	statItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(false)

	statValueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true)

	inputBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86")).
		Width(contentWidth).
		Padding(0, 2).
		MarginTop(1)

	wordBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(contentWidth).
		Padding(1, 2).
		MarginTop(1).
		MarginBottom(1)

	separatorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
)

// Layout constants
const (
	// Fixed width for all content areas to ensure alignment
	contentWidth = 80
)

// GameStats game statistics
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

// WordInfo contains word display information
type WordInfo struct {
	Text        string
	Completed   bool
	CompletedAt time.Time
}

// WelcomeAnimationState tracks the welcome screen animation state
type WelcomeAnimationState struct {
	Frame              int
	SelectedOption     int  // 0 for start, 1 for about
	BulletActive       bool // whether a bullet is currently flying
	BulletX            int  // bullet column position
	BulletRow          int  // which line the bullet is on (relative to content box)
	Exploding          bool // whether tagline is exploding
	ExplosionTime      time.Time
	ExplosionTriggered bool
}

// TaglineInfo contains tagline display information for explosion animation
type TaglineInfo struct {
	Text        string
	CompletedAt time.Time
}

// RenderWelcome renders welcome screen with unified style
func RenderWelcome(state *WelcomeAnimationState, animFrame int) string {
	var s strings.Builder

	// TOP: Header
	header := headerStyle.Render("Word Killer")
	s.WriteString(lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(header))
	s.WriteString("\n")

	// MIDDLE: Content (tagline + menu + bullet animation)
	content := renderWelcomeContent(state.SelectedOption, animFrame, state)
	s.WriteString(content)
	s.WriteString("\n")

	// BOTTOM: Hints
	hints := inputBoxStyle.Render("[↑↓] Select  │  [Enter] Confirm  │  [ESC] Quit")
	s.WriteString(hints)
	s.WriteString("\n")

	return s.String()
}

// renderWelcomeContent renders the welcome screen content area
func renderWelcomeContent(selectedOption int, animFrame int, state *WelcomeAnimationState) string {
	const totalLines = 12 // Total lines in the content box
	var lines []string

	// Line 0: Version (top-left)
	topline := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Render("V1.0.0")
	versionLine := lipgloss.NewStyle().Width(contentWidth-8).
		Align(lipgloss.Left, lipgloss.Top).
		Render(topline)
	lines = append(lines, addBulletToLine(versionLine, 0, state))

	// Line 1: Empty
	lines = append(lines, addBulletToLine(strings.Repeat(" ", contentWidth-8), 1, state))

	// Lines 2-3: Menu options (Start, About)
	options := []string{"Start", "About"}
	selectedStyle := lipgloss.NewStyle().Foreground(getRandomMenuColor(animFrame)).Bold(true)

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

		lineIndex := 2 + i
		lines = append(lines, addBulletToLine("  "+alignedText, lineIndex, state))
	}

	// Lines 4-9: Empty (middle spacing)
	for i := 4; i < 10; i++ {
		lines = append(lines, addBulletToLine(strings.Repeat(" ", contentWidth-8), i, state))
	}

	// Line 10: Tagline with explosion effect (bottom-right)
	taglineText := "Low-key but never simple"
	var taglineRendered string

	if state.Exploding {
		taglineRendered = renderTaglineExplosion(TaglineInfo{
			Text:        taglineText,
			CompletedAt: state.ExplosionTime,
		})
	} else {
		taglineRendered = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Render(taglineText)
	}

	taglineLine := lipgloss.NewStyle().Width(contentWidth-8).
		Align(lipgloss.Right, lipgloss.Bottom).
		Render(taglineRendered)
	lines = append(lines, addBulletToLine(taglineLine, 10, state))

	// Line 11: Empty
	lines = append(lines, addBulletToLine(strings.Repeat(" ", contentWidth-8), 11, state))

	return wordBoxStyle.Render(strings.Join(lines, "\n"))
}

// addBulletToLine adds bullet to a line if the bullet is on that line
func addBulletToLine(line string, lineIndex int, state *WelcomeAnimationState) string {
	if !state.BulletActive || state.BulletRow != lineIndex {
		return line
	}

	// Strip ANSI codes to calculate actual width
	plainLine := stripAnsi(line)
	if state.BulletX >= 0 && state.BulletX < len(plainLine) {
		// Inject bullet character - note: this is simplified for demonstration
		bulletChar := lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true).
			Render("►")

		// For simplicity, overlay the bullet at position (not perfect but functional)
		return line[:min(state.BulletX, len(line))] + bulletChar + line[min(state.BulletX+1, len(line)):]
	}

	return line
}

// stripAnsi removes ANSI escape codes for width calculation
func stripAnsi(s string) string {
	// Simple approximation - just return input for now
	// In production, use a proper ANSI stripper
	return s
}

// RenderGame renders game screen with professional layout
func RenderGame(words []WordInfo, highlightedIndices []int, input string, stats GameStats, remainingWords int) string {
	var s strings.Builder

	// === TOP: Status Bar ===
	statusBar := renderStatusBar(stats, remainingWords)
	s.WriteString(statusBar)
	s.WriteString("\n")

	// === MIDDLE: Word List Area ===
	wordArea := renderWordArea(words, highlightedIndices, input)
	s.WriteString(wordArea)
	s.WriteString("\n")

	// === BOTTOM: Input Area ===
	inputArea := renderInputArea(input)
	s.WriteString(inputArea)
	s.WriteString("\n")

	// Hints
	s.WriteString(hintStyle.Render("  [ESC] Pause  "))
	s.WriteString("\n")

	return s.String()
}

// renderStatusBar renders the top status bar with statistics
func renderStatusBar(stats GameStats, remainingWords int) string {
	// Format each stat with fixed width for alignment
	timeStr := fmt.Sprintf("Time: %6.1fs", stats.ElapsedSeconds)
	progressStr := fmt.Sprintf("Progress: %2d/%-2d", stats.WordsCompleted, stats.WordsCompleted+remainingWords)
	speedStr := fmt.Sprintf("Speed: %5.1f l/s", stats.LettersPerSecond)
	accuracyStr := fmt.Sprintf("Accuracy: %5.1f%%", stats.AccuracyPercent)

	// Build status line with fixed spacing
	statusLine := fmt.Sprintf("%s  │  %s  │  %s  │  %s",
		timeStr, progressStr, speedStr, accuracyStr)

	// Apply style and ensure fixed width
	styled := headerStyle.Render(statusLine)
	return lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(styled)
}

// renderWordArea renders the middle word list area with fixed height
func renderWordArea(words []WordInfo, highlightedIndices []int, input string) string {
	if len(words) == 0 {
		content := statsStyle.Render("All words completed!")
		return wordBoxStyle.Render(content)
	}

	var wordLines []string
	wordLines = append(wordLines, titleStyle.Render("Words:"))
	wordLines = append(wordLines, "")

	// Fixed layout: 10 rows maximum
	const maxRows = 10
	const wordColumnWidth = 18 // Each word gets 18 chars (word + spacing)

	// Calculate optimal columns based on content width
	availableWidth := contentWidth - 8 // Reserve space for padding and borders
	wordsPerRow := availableWidth / wordColumnWidth
	if wordsPerRow < 1 {
		wordsPerRow = 1
	}

	// Calculate max words to display
	maxWordsToDisplay := maxRows * wordsPerRow

	// Render words in columns (up to maxWordsToDisplay)
	displayCount := len(words)
	if displayCount > maxWordsToDisplay {
		displayCount = maxWordsToDisplay
	}

	rowCount := 0
	for i := 0; i < displayCount; i += wordsPerRow {
		var rowWords []string
		for j := 0; j < wordsPerRow; j++ {
			idx := i + j
			if idx >= displayCount {
				// Fill empty cells with spaces
				rowWords = append(rowWords, strings.Repeat(" ", wordColumnWidth))
				continue
			}

			wordInfo := words[idx]

			// Check if highlighted (only for active words)
			isHighlighted := false
			if !wordInfo.Completed {
				for _, hIdx := range highlightedIndices {
					if hIdx == idx {
						isHighlighted = true
						break
					}
				}
			}

			var renderedWord string

			// Render completed word with animation
			if wordInfo.Completed {
				renderedWord = renderCompletedWordAnimation(wordInfo)
			} else if isHighlighted && len(input) > 0 {
				// Highlight matched part for active words
				matchLen := len(input)
				if matchLen > len(wordInfo.Text) {
					matchLen = len(wordInfo.Text)
				}
				renderedWord = highlightStyle.Render(wordInfo.Text[:matchLen]) + wordStyle.Render(wordInfo.Text[matchLen:])
			} else {
				// Normal active word
				renderedWord = wordStyle.Render(wordInfo.Text)
			}

			// Pad word to fixed width for alignment
			paddedWord := padToWidth(wordInfo.Text, renderedWord, wordColumnWidth)
			rowWords = append(rowWords, paddedWord)
		}
		wordLines = append(wordLines, "  "+strings.Join(rowWords, ""))
		rowCount++
	}

	// Fill remaining rows with empty lines to maintain fixed height
	for rowCount < maxRows {
		emptyRow := "  " + strings.Repeat(" ", wordsPerRow*wordColumnWidth)
		wordLines = append(wordLines, emptyRow)
		rowCount++
	}

	content := strings.Join(wordLines, "\n")
	return wordBoxStyle.Render(content)
}

// padToWidth pads a styled string to a specific width based on the plain text length
func padToWidth(plainText, styledText string, width int) string {
	plainLen := len(plainText)
	if plainLen >= width {
		return styledText
	}
	padding := strings.Repeat(" ", width-plainLen)
	return styledText + padding
}

// renderCompletedWordAnimation renders a completed word with hit effect and letter-by-letter fade
func renderCompletedWordAnimation(wordInfo WordInfo) string {
	timeSinceCompletion := time.Since(wordInfo.CompletedAt)
	msElapsed := timeSinceCompletion.Milliseconds()

	// Phase 1: Hit effect (0-150ms) - Bright flash
	if msElapsed < 150 {
		// Alternate between bright yellow and bright green for impact
		if msElapsed < 50 {
			return hitEffectStyle.Render(wordInfo.Text)
		} else if msElapsed < 100 {
			return fadingStyle1.Render(wordInfo.Text)
		} else {
			return hitEffectStyle.Render(wordInfo.Text)
		}
	}

	// Phase 2: Letter-by-letter fade (150ms onwards)
	// Each letter takes 80ms to fade through colors
	const letterFadeTime = 80  // ms per letter
	const animationStart = 150 // when fade animation starts

	var result strings.Builder
	wordLen := len(wordInfo.Text)

	for i, ch := range wordInfo.Text {
		letterStartTime := animationStart + int64(i*letterFadeTime)
		letterElapsed := msElapsed - letterStartTime

		var charStyle lipgloss.Style

		if letterElapsed < 0 {
			// Letter hasn't started fading yet - still bright green
			charStyle = fadingStyle1
		} else if letterElapsed < 20 {
			// Stage 1: Bright green
			charStyle = fadingStyle1
		} else if letterElapsed < 40 {
			// Stage 2: Medium green
			charStyle = fadingStyle2
		} else if letterElapsed < 60 {
			// Stage 3: Light gray
			charStyle = fadingStyle3
		} else if letterElapsed < 80 {
			// Stage 4: Medium gray
			charStyle = fadingStyle4
		} else {
			// Stage 5: Dark gray (final)
			charStyle = fadingStyle5
		}

		result.WriteString(charStyle.Render(string(ch)))
	}

	// Add strikethrough after complete fade (all letters + fade time)
	totalAnimTime := animationStart + int64(wordLen*letterFadeTime) + 80
	if msElapsed >= totalAnimTime {
		return completedWordStyle.Render(wordInfo.Text)
	}

	return result.String()
}

// renderInputArea renders the bottom input area
func renderInputArea(input string) string {
	// Input label and value
	label := statItemStyle.Render("Input: ")
	value := inputStyle.Render(input)
	if input == "" {
		value = statItemStyle.Render("_")
	}

	content := label + value
	return inputBoxStyle.Render(content)
}

// RenderPauseMenu renders pause menu with stats and animation
func RenderPauseMenu(selectedIndex int, stats GameStats, remainingWords int, animFrame int) string {
	var s strings.Builder

	// === TOP: Status Bar (same as game screen) ===
	statusBar := renderStatusBar(stats, remainingWords)
	s.WriteString(statusBar)
	s.WriteString("\n")

	// === MIDDLE: Pause Animation + Menu ===
	pauseArea := renderPauseArea(selectedIndex, animFrame)
	s.WriteString(pauseArea)
	s.WriteString("\n")

	// === BOTTOM: Hints ===
	hints := inputBoxStyle.Render("[↑↓] Select  │  [Enter] Confirm  │  [ESC] Quit Game")
	s.WriteString(hints)
	s.WriteString("\n")

	return s.String()
}

// renderPauseArea renders the pause menu area with scrolling animation and fixed height
func renderPauseArea(selectedIndex int, animFrame int) string {
	// Scrolling "GAME PAUSED" text
	pauseText := "    GAME PAUSED    "
	displayWidth := 20 // width inside the box
	totalLength := len(pauseText) + displayWidth

	// Simple scrolling: text moves from right to left
	scrollOffset := animFrame % totalLength
	extendedText := strings.Repeat(" ", displayWidth) + pauseText + strings.Repeat(" ", displayWidth)
	startPos := scrollOffset
	if startPos+displayWidth > len(extendedText) {
		startPos = len(extendedText) - displayWidth
	}
	visibleText := extendedText[startPos : startPos+displayWidth]

	// Fixed height: match word area (12 lines total)
	lines := []string{}

	// Title line
	lines = append(lines, titleStyle.Render("Pause Menu:"))
	lines = append(lines, "") // Empty line

	// Scrolling text box (5 lines)
	lines = append(lines, "                         "+titleStyle.Render("╔══════════════════════╗"))
	lines = append(lines, "                         "+titleStyle.Render("║                      ║"))
	lines = append(lines, "                         "+titleStyle.Render("║ ")+hintStyle.Render(visibleText)+titleStyle.Render(" ║"))
	lines = append(lines, "                         "+titleStyle.Render("║                      ║"))
	lines = append(lines, "                         "+titleStyle.Render("╚══════════════════════╝"))

	lines = append(lines, "") // Empty line after box

	// Menu options - 4 options now
	options := []string{"Resume Game", "Restart", "Select Mode", "Main Menu"}
	// Use random color for selected option
	selectedStyle := lipgloss.NewStyle().
		Foreground(getRandomMenuColor(animFrame)).
		Bold(true)

	for i, opt := range options {
		// Build the option text with indicator
		var optionDisplay string
		if i == selectedIndex {
			optionDisplay = "> " + opt + " <"
		} else {
			optionDisplay = "  " + opt + "  "
		}

		// Apply style
		var styledText string
		if i == selectedIndex {
			styledText = selectedStyle.Render(optionDisplay)
		} else {
			styledText = menuNormalStyle.Render(optionDisplay)
		}

		// Use lipgloss to center the text within the content width
		alignedText := lipgloss.NewStyle().
			Width(contentWidth - 8). // Reserve space for padding and borders
			Align(lipgloss.Center).
			Render(styledText)

		lines = append(lines, "  "+alignedText)
	}

	// Fill to exactly 12 lines (title + empty + 5 box + 1 empty + 4 menu = 12 lines)
	for len(lines) < 12 {
		lines = append(lines, "")
	}

	return wordBoxStyle.Render(strings.Join(lines, "\n"))
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getRandomMenuColor returns a color for menu selection based on frame
func getRandomMenuColor(frame int) lipgloss.Color {
	// Define a palette of vibrant colors
	colors := []string{
		"226", // Bright yellow
		"46",  // Bright green
		"51",  // Bright cyan
		"201", // Bright magenta
		"208", // Bright orange
		"196", // Bright red
		"93",  // Bright purple
		"87",  // Light blue
	}

	// Change color every 5 frames for smooth transition
	colorIndex := (frame / 5) % len(colors)
	return lipgloss.Color(colors[colorIndex])
}

// RenderResults renders game results with consistent layout
func RenderResults(stats GameStats, aborted bool, selectedOption int, animFrame int) string {
	var s strings.Builder

	// === TOP: Header ===
	var header string
	if aborted {
		header = "Time:   " + fmt.Sprintf("%6.1fs", stats.ElapsedSeconds) + "  │  Status: Game Over"
	} else {
		header = "Time:   " + fmt.Sprintf("%6.1fs", stats.ElapsedSeconds) + "  │  Status: Completed! 🎉"
	}
	headerStyled := headerStyle.Render(header)
	s.WriteString(lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(headerStyled))
	s.WriteString("\n")

	// === MIDDLE: Statistics Area ===
	statsArea := renderResultsArea(stats, aborted, selectedOption, animFrame)
	s.WriteString(statsArea)
	s.WriteString("\n")

	// === BOTTOM: Hints ===
	hints := inputBoxStyle.Render("[↑↓] Select  │  [Enter] Confirm  │  [ESC] Exit")
	s.WriteString(hints)
	s.WriteString("\n")

	return s.String()
}

// renderResultsArea renders the statistics area
func renderResultsArea(stats GameStats, aborted bool, selectedOption int, animFrame int) string {
	var content strings.Builder

	// Title
	if aborted {
		content.WriteString(fmt.Sprintf("%58s\n", titleStyle.Render("GAME OVER")))
	} else {
		content.WriteString(fmt.Sprintf("%70s\n", titleStyle.Render("CONGRATULATIONS")))
	}
	content.WriteString("    " + separatorStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━") + "\n")

	// Statistics section
	content.WriteString(fmt.Sprintf("%51s\n", titleStyle.Render("Performance Metrics:")))

	// Keystroke stats
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Total Keystrokes:"),
		statValueStyle.Render(fmt.Sprintf("%7d", stats.TotalKeystrokes))))
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Valid Keystrokes:"),
		statValueStyle.Render(fmt.Sprintf("%7d", stats.ValidKeystrokes))))
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Correct Chars:"),
		statValueStyle.Render(fmt.Sprintf("%7d", stats.CorrectChars))))

	// Word stats
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Completed Words:"),
		statValueStyle.Render(fmt.Sprintf("%7d", stats.WordsCompleted))))
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Total Letters:"),
		statValueStyle.Render(fmt.Sprintf("%7d", stats.TotalLetters))))

	// Speed and accuracy (highlighted)
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Letters/second:"),
		statValueStyle.Render(fmt.Sprintf("%7.2f", stats.LettersPerSecond))))
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Words/second:"),
		statValueStyle.Render(fmt.Sprintf("%7.2f", stats.WordsPerSecond))))
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Accuracy:"),
		statValueStyle.Render(fmt.Sprintf("%6.2f%%", stats.AccuracyPercent))))

	// Add separator before menu
	content.WriteString("\n")
	content.WriteString("    " + separatorStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━") + "\n")

	// Menu options - similar to pause menu
	content.WriteString("\n")
	options := []string{"Restart", "Select Mode", "Main Menu"}
	// Use random color for selected option
	selectedStyle := lipgloss.NewStyle().
		Foreground(getRandomMenuColor(animFrame)).
		Bold(true)

	for i, opt := range options {
		// Build the option text with indicator
		var optionDisplay string
		if i == selectedOption {
			optionDisplay = "> " + opt + " <"
		} else {
			optionDisplay = "  " + opt + "  "
		}

		// Apply style
		var styledText string
		if i == selectedOption {
			styledText = selectedStyle.Render(optionDisplay)
		} else {
			styledText = menuNormalStyle.Render(optionDisplay)
		}

		// Use lipgloss to center the text within the content width
		alignedText := lipgloss.NewStyle().
			Width(contentWidth - 8). // Reserve space for padding and borders
			Align(lipgloss.Center).
			Render(styledText)

		content.WriteString("  " + alignedText + "\n")
	}

	return wordBoxStyle.Render(content.String())
}

// RenderModeSelection renders the mode selection screen with unified style
func RenderModeSelection(selectedMode int, animFrame int) string {
	var s strings.Builder

	// TOP: Header
	header := headerStyle.Render("Select Game Mode")
	s.WriteString(lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(header))
	s.WriteString("\n")

	// MIDDLE: Content (mode options)
	content := renderModeSelectionContent(selectedMode, animFrame)
	s.WriteString(content)
	s.WriteString("\n")

	// BOTTOM: Hints
	hints := inputBoxStyle.Render("[↑↓] Select  │  [Enter] Confirm  │  [ESC] Back")
	s.WriteString(hints)
	s.WriteString("\n")

	return s.String()
}

// renderModeSelectionContent renders the mode selection content area
func renderModeSelectionContent(selectedMode int, animFrame int) string {
	var lines []string

	lines = append(lines, "")

	// Menu options - 现在有6个模式
	options := []string{
		"Classic Mode",
		"Sentence Mode",
		"Countdown Mode",
		"Speed Run Mode",
		"Rhythm Master",
		"Underwater Countdown",
	}
	selectedStyle := lipgloss.NewStyle().Foreground(getRandomMenuColor(animFrame)).Bold(true)

	for i, opt := range options {
		var optionDisplay string
		if i == selectedMode {
			optionDisplay = "> " + opt + " <"
		} else {
			optionDisplay = "  " + opt + "  "
		}

		var styledText string
		if i == selectedMode {
			styledText = selectedStyle.Render(optionDisplay)
		} else {
			styledText = menuNormalStyle.Render(optionDisplay)
		}

		alignedText := lipgloss.NewStyle().
			Width(contentWidth - 8).
			Align(lipgloss.Center).
			Render(styledText)

		lines = append(lines, "  "+alignedText)
	}

	// Fill to fixed height (14 lines to accommodate 5 options)
	for len(lines) < 14 {
		lines = append(lines, "")
	}

	return wordBoxStyle.Render(strings.Join(lines, "\n"))
}

// RenderSentenceGame renders the sentence typing game screen
func RenderSentenceGame(targetSentence string, userInput string, stats GameStats) string {
	var s strings.Builder

	// === TOP: Status Bar ===
	timeStr := fmt.Sprintf("Time: %6.1fs", stats.ElapsedSeconds)
	progressStr := fmt.Sprintf("Progress: %2d/%2d", len(userInput), len(targetSentence))
	accuracyStr := fmt.Sprintf("Accuracy: %5.1f%%", stats.AccuracyPercent)

	statusLine := fmt.Sprintf("%s  │  %s  │  %s", timeStr, progressStr, accuracyStr)
	statusStyled := headerStyle.Render(statusLine)
	s.WriteString(lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(statusStyled))
	s.WriteString("\n\n")

	// === MIDDLE: Sentence Display Area ===
	sentenceArea := renderSentenceArea(targetSentence, userInput)
	s.WriteString(sentenceArea)
	s.WriteString("\n")

	// === BOTTOM: Stats and Hints ===
	detailedStats := renderSentenceStats(stats, len(targetSentence))
	s.WriteString(detailedStats)
	s.WriteString("\n")

	s.WriteString(hintStyle.Render("  [ESC] Pause  "))
	s.WriteString("\n")

	return s.String()
}

// renderSentenceArea renders the target sentence and user input with color coding
func renderSentenceArea(targetSentence string, userInput string) string {
	var content strings.Builder

	content.WriteString(titleStyle.Render("Target:") + "\n")
	content.WriteString("  " + wordStyle.Render(targetSentence) + "\n\n")

	content.WriteString(titleStyle.Render("Your Input:") + "\n")
	content.WriteString("  ")

	// Render each character with color coding
	for i, targetChar := range targetSentence {
		if i < len(userInput) {
			userChar := rune(userInput[i])
			if userChar == targetChar {
				// Correct character - green
				content.WriteString(highlightStyle.Render(string(userChar)))
			} else {
				// Incorrect character - red
				errorStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("196")).
					Bold(true)
				content.WriteString(errorStyle.Render(string(userChar)))
			}
		} else {
			// Not yet typed - show target in gray
			content.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Render(string(targetChar)))
		}
	}

	return wordBoxStyle.Render(content.String())
}

// renderSentenceStats renders detailed statistics for sentence mode
func renderSentenceStats(stats GameStats, totalChars int) string {
	statsLine := fmt.Sprintf("Characters: %d/%d  │  Correct: %d  │  Speed: %.1f chars/s",
		stats.TotalKeystrokes,
		totalChars,
		stats.CorrectChars,
		stats.LettersPerSecond)

	content := statsStyle.Render(statsLine)
	return inputBoxStyle.Render(content)
}

// UpdateWelcomeAnimation updates the welcome screen animation state
func UpdateWelcomeAnimation(state *WelcomeAnimationState) {
	state.Frame++

	// Trigger bullet if not active and explosion not triggered
	if !state.BulletActive && !state.ExplosionTriggered {
		// Start bullet from bottom-left (line 10, X=0)
		state.BulletActive = true
		state.BulletX = 0
		state.BulletRow = 10 // Tagline is on line 10
	}

	// Move bullet if active
	if state.BulletActive {
		state.BulletX += 2 // Move bullet to the right

		// Calculate tagline starting position
		// Tagline "Low-key but never simple" is 24 chars, right-aligned
		taglineText := "Low-key but never simple"
		taglineStartX := (contentWidth - 8) - len(taglineText)

		// Check if bullet hits the tagline at line 10
		if state.BulletRow == 10 && state.BulletX >= taglineStartX {
			// Bullet reached tagline
			state.BulletActive = false
			state.Exploding = true
			state.ExplosionTime = time.Now()
			state.ExplosionTriggered = true
		}

		// If bullet goes past content width without hitting, reset
		if state.BulletX >= (contentWidth - 8) {
			state.BulletActive = false
			state.ExplosionTriggered = false
		}
	}

	// End explosion after animation completes
	if state.Exploding {
		elapsed := time.Since(state.ExplosionTime)
		// Total explosion animation: 24 chars * 80ms = 1920ms + extra padding
		if elapsed > 2500*time.Millisecond {
			state.Exploding = false
			// Reset for next cycle
			state.ExplosionTriggered = false
		}
	}
}

// renderTaglineExplosion renders the tagline with letter-by-letter explosion effect
func renderTaglineExplosion(taglineInfo TaglineInfo) string {
	timeSinceExplosion := time.Since(taglineInfo.CompletedAt)
	msElapsed := timeSinceExplosion.Milliseconds()

	// Phase 1: Hit effect (0-150ms) - Bright flash for strong impact
	if msElapsed < 150 {
		// Alternate between bright yellow and bright green for impact
		if msElapsed < 50 {
			return hitEffectStyle.Render(taglineInfo.Text)
		} else if msElapsed < 100 {
			return fadingStyle1.Render(taglineInfo.Text)
		} else {
			return hitEffectStyle.Render(taglineInfo.Text)
		}
	}

	// Phase 2: Letter-by-letter fade (150ms onwards)
	// Each letter takes 80ms to fade through colors
	const letterFadeTime = 80  // ms per letter
	const animationStart = 150 // when fade animation starts

	var result strings.Builder

	for i, ch := range taglineInfo.Text {
		letterStartTime := animationStart + int64(i*letterFadeTime)
		letterElapsed := msElapsed - letterStartTime

		var charStyle lipgloss.Style

		if letterElapsed < 0 {
			// Letter hasn't started fading yet - still bright green
			charStyle = fadingStyle1
		} else if letterElapsed < 20 {
			// Stage 1: Bright green
			charStyle = fadingStyle1
		} else if letterElapsed < 40 {
			// Stage 2: Medium green
			charStyle = fadingStyle2
		} else if letterElapsed < 60 {
			// Stage 3: Light gray
			charStyle = fadingStyle3
		} else if letterElapsed < 80 {
			// Stage 4: Medium gray
			charStyle = fadingStyle4
		} else {
			// Stage 5: Back to normal blue
			charStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
		}

		result.WriteString(charStyle.Render(string(ch)))
	}

	return result.String()
}

// RenderCountdownGame 渲染倒计时模式游戏界面
func RenderCountdownGame(words []WordInfo, highlightedIndices []int, input string, stats GameStats,
	timeRemaining float64, totalDuration float64) string {
	var s strings.Builder

	// === 顶部：倒计时器（大号显示）===
	timeRemainingInt := int(timeRemaining)
	var timerStyle lipgloss.Style

	// 小于10秒时显示红色警告
	if timeRemaining < 10 {
		timerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Background(lipgloss.Color("52")) // 深红背景
	} else {
		timerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)
	}

	timerDisplay := fmt.Sprintf("⏱  Time: %02d:%02d", timeRemainingInt/60, timeRemainingInt%60)
	timerRendered := timerStyle.Render(timerDisplay)

	progressStr := fmt.Sprintf("Words: %d", stats.WordsCompleted)
	speedStr := fmt.Sprintf("Speed: %.1f letters/s", stats.LettersPerSecond)

	statusLine := fmt.Sprintf("%s  │  %s  │  %s", timerRendered, progressStr, speedStr)
	statusStyled := headerStyle.Render(statusLine)
	s.WriteString(lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(statusStyled))
	s.WriteString("\n")

	// === 中部：单词区域（复用现有渲染）===
	wordArea := renderWordArea(words, highlightedIndices, input)
	s.WriteString(wordArea)
	s.WriteString("\n")

	// === 底部：输入区域 ===
	inputArea := renderInputArea(input)
	s.WriteString(inputArea)
	s.WriteString("\n")

	s.WriteString(hintStyle.Render("  [ESC] Pause  │  Eliminate as many words as possible before time runs out!"))
	s.WriteString("\n")

	return s.String()
}

// RenderSpeedRunGame 渲染极速模式游戏界面
func RenderSpeedRunGame(words []WordInfo, highlightedIndices []int, input string, stats GameStats,
	currentTime float64, bestTime float64) string {
	var s strings.Builder

	// === 顶部：毫秒级计时器 ===
	timerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true)

	// 格式：MM:SS.mmm
	minutes := int(currentTime) / 60
	seconds := int(currentTime) % 60
	milliseconds := int((currentTime - float64(int(currentTime))) * 1000)
	timerDisplay := fmt.Sprintf("⏱  %02d:%02d.%03d", minutes, seconds, milliseconds)
	timerRendered := timerStyle.Render(timerDisplay)

	// 计算剩余单词数
	remainingWords := 0
	for _, w := range words {
		if !w.Completed {
			remainingWords++
		}
	}
	progressStr := fmt.Sprintf("Words: %d/%d", stats.WordsCompleted, stats.WordsCompleted+remainingWords)

	// 最佳记录显示
	var bestDisplay string
	if bestTime > 0 {
		bestMinutes := int(bestTime) / 60
		bestSeconds := int(bestTime) % 60
		bestMillis := int((bestTime - float64(int(bestTime))) * 1000)
		bestDisplay = fmt.Sprintf("Best: %02d:%02d.%03d", bestMinutes, bestSeconds, bestMillis)
	} else {
		bestDisplay = "Best: --:--:---"
	}

	statusLine := fmt.Sprintf("%s  │  %s  │  %s", timerRendered, progressStr, bestDisplay)
	statusStyled := headerStyle.Render(statusLine)
	s.WriteString(lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(statusStyled))
	s.WriteString("\n")

	// === 中部：单词区域 ===
	wordArea := renderWordArea(words, highlightedIndices, input)
	s.WriteString(wordArea)
	s.WriteString("\n")

	// === 底部：输入 + 速度指标 ===
	inputArea := renderInputArea(input)
	s.WriteString(inputArea)
	s.WriteString("\n")

	// 当前速度指标
	speedIndicator := fmt.Sprintf("Current Speed: %.2f words/s", stats.WordsPerSecond)
	speedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
	s.WriteString(speedStyle.Render("  " + speedIndicator))
	s.WriteString("\n")

	s.WriteString(hintStyle.Render("  [ESC] Pause  │  Complete all words as fast as possible!"))
	s.WriteString("\n")

	return s.String()
}

// RenderRhythmMasterGame 渲染节奏大师模式游戏界面
func RenderRhythmMasterGame(words []WordInfo, highlightedIndices []int, input string, stats GameStats,
	wordTimeRemaining float64, wordTimeLimit float64, combo int, level int) string {
	var s strings.Builder

	// === 顶部：连击和等级显示 ===
	comboStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("201")).
		Bold(true)

	comboDisplay := fmt.Sprintf("Combo: %d", combo)
	levelDisplay := fmt.Sprintf("Level: %d", level)
	speedDisplay := fmt.Sprintf("Speed: %.1f letters/s", stats.LettersPerSecond)

	statusLine := fmt.Sprintf("%s  │  %s  │  %s", comboStyle.Render(comboDisplay), levelDisplay, speedDisplay)
	statusStyled := headerStyle.Render(statusLine)
	s.WriteString(lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(statusStyled))
	s.WriteString("\n")

	// === 中部：带进度条的单词区域 ===
	wordArea := renderRhythmWordArea(words, highlightedIndices, input, wordTimeRemaining, wordTimeLimit)
	s.WriteString(wordArea)
	s.WriteString("\n")

	// === 时间限制显示 ===
	timeLimitInfo := fmt.Sprintf("Time Limit per Word: %.1fs", wordTimeLimit)
	timeLimitStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
	s.WriteString(inputBoxStyle.Render(timeLimitStyle.Render(timeLimitInfo)))
	s.WriteString("\n")

	// === 底部：输入区域 ===
	inputArea := renderInputArea(input)
	s.WriteString(inputArea)
	s.WriteString("\n")

	s.WriteString(hintStyle.Render("  [ESC] Pause  │  Complete each word within the time limit!"))
	s.WriteString("\n")

	return s.String()
}

// renderRhythmWordArea 渲染带进度条的单词区域（节奏大师专用）
func renderRhythmWordArea(words []WordInfo, highlightedIndices []int, input string,
	wordTimeRemaining float64, wordTimeLimit float64) string {
	if len(words) == 0 {
		content := statsStyle.Render("All words completed!")
		return wordBoxStyle.Render(content)
	}

	var wordLines []string
	wordLines = append(wordLines, titleStyle.Render("Words:"))
	wordLines = append(wordLines, "")

	const maxRows = 10
	const wordColumnWidth = 18

	availableWidth := contentWidth - 8
	wordsPerRow := availableWidth / wordColumnWidth
	if wordsPerRow < 1 {
		wordsPerRow = 1
	}

	maxWordsToDisplay := maxRows * wordsPerRow
	displayCount := len(words)
	if displayCount > maxWordsToDisplay {
		displayCount = maxWordsToDisplay
	}

	// 找到第一个活动单词（用于显示进度条）
	firstActiveIdx := -1
	for i := 0; i < displayCount; i++ {
		if !words[i].Completed {
			firstActiveIdx = i
			break
		}
	}

	rowCount := 0
	for i := 0; i < displayCount; i += wordsPerRow {
		var rowWords []string
		for j := 0; j < wordsPerRow; j++ {
			idx := i + j
			if idx >= displayCount {
				rowWords = append(rowWords, strings.Repeat(" ", wordColumnWidth))
				continue
			}

			wordInfo := words[idx]
			isHighlighted := false
			if !wordInfo.Completed {
				for _, hIdx := range highlightedIndices {
					if hIdx == idx {
						isHighlighted = true
						break
					}
				}
			}

			var renderedWord string

			if wordInfo.Completed {
				// 已完成的单词
				renderedWord = renderCompletedWordAnimation(wordInfo)
			} else if idx == firstActiveIdx {
				// 当前活动单词 - 显示进度条
				if isHighlighted && len(input) > 0 {
					matchLen := len(input)
					if matchLen > len(wordInfo.Text) {
						matchLen = len(wordInfo.Text)
					}
					renderedWord = highlightStyle.Render(wordInfo.Text[:matchLen]) +
						wordStyle.Render(wordInfo.Text[matchLen:])
				} else {
					renderedWord = wordStyle.Render(wordInfo.Text)
				}

				// 添加进度条
				barWidth := 16
				progressPercent := wordTimeRemaining / wordTimeLimit
				if progressPercent < 0 {
					progressPercent = 0
				}
				filledWidth := int(float64(barWidth) * progressPercent)

				// 根据剩余时间设置颜色
				var barStyle lipgloss.Style
				if progressPercent < 0.3 {
					barStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // 红色
				} else if progressPercent < 0.6 {
					barStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // 黄色
				} else {
					barStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46")) // 绿色
				}

				bar := barStyle.Render(strings.Repeat("█", filledWidth)) +
					lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("░", barWidth-filledWidth))

				renderedWord = renderedWord + "\n  " + bar
			} else if isHighlighted && len(input) > 0 {
				// 高亮匹配的单词
				matchLen := len(input)
				if matchLen > len(wordInfo.Text) {
					matchLen = len(wordInfo.Text)
				}
				renderedWord = highlightStyle.Render(wordInfo.Text[:matchLen]) +
					wordStyle.Render(wordInfo.Text[matchLen:])
			} else {
				// 普通未完成单词
				renderedWord = wordStyle.Render(wordInfo.Text)
			}

			paddedWord := padToWidth(wordInfo.Text, renderedWord, wordColumnWidth)
			rowWords = append(rowWords, paddedWord)
		}
		wordLines = append(wordLines, "  "+strings.Join(rowWords, ""))
		rowCount++
	}

	// 填充到固定高度
	for rowCount < maxRows {
		emptyRow := "  " + strings.Repeat(" ", wordsPerRow*wordColumnWidth)
		wordLines = append(wordLines, emptyRow)
		rowCount++
	}

	content := strings.Join(wordLines, "\n")
	return wordBoxStyle.Render(content)
}

// RenderAbout renders the about page with game information
func RenderAbout() string {
	var s strings.Builder

	// TOP: Header
	header := headerStyle.Render("About Word Killer")
	s.WriteString(lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(header))
	s.WriteString("\n")

	// MIDDLE: Content
	content := renderAboutContent()
	s.WriteString(content)
	s.WriteString("\n")

	// BOTTOM: Hints
	hints := inputBoxStyle.Render("[ESC] Back to Main Menu")
	s.WriteString(hints)
	s.WriteString("\n")

	return s.String()
}

// renderAboutContent renders the about page content
func renderAboutContent() string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("  Game Modes:"))
	lines = append(lines, "")
	lines = append(lines, "    "+statsStyle.Render("• Classic Mode")+" - Type and eliminate falling words")
	lines = append(lines, "    "+statsStyle.Render("• Sentence Mode")+" - Type complete sentences accurately")
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("  Features:"))
	lines = append(lines, "")
	lines = append(lines, "    "+statsStyle.Render("• Real-time statistics")+" - Track your speed and accuracy")
	lines = append(lines, "    "+statsStyle.Render("• Multiple difficulty levels")+" - Short, medium, and long words")
	lines = append(lines, "")
	lines = append(lines, titleStyle.Render("  License:"))
	lines = append(lines, "")
	lines = append(lines, "    "+hintStyle.Render("Open Source")+" - MIT License")
	lines = append(lines, "    "+hintStyle.Render("https://github.com/word-killer/word-killer"))

	// Fill to fixed height (12 lines)
	for len(lines) < 12 {
		lines = append(lines, "")
	}

	return wordBoxStyle.Render(strings.Join(lines, "\n"))
}
