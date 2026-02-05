package ui

import (
	"fmt"
	"math"
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
		Bold(true).
		Underline(true)

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
	BulletActive       bool        // whether a bullet is currently flying
	BulletX            int         // bullet column position (0-45 for internal width)
	BulletY            int         // bullet row position (0=title row, 1=empty1, 2=empty2, 3=tagline)
	Exploding          bool        // whether tagline is exploding
	ExplosionTime      time.Time   // when explosion started (for renderCompletedWordAnimation logic)
	TaglineWord        TaglineInfo // tagline word info for explosion animation
	ExplosionTriggered bool        // whether explosion was triggered (to prevent re-triggering)
	SelectedOption     int         // 0 for start, 1 for quit
}

// TaglineInfo contains tagline display information for explosion animation
type TaglineInfo struct {
	Text        string
	Completed   bool
	CompletedAt time.Time
}

// RenderWelcome renders welcome screen with bullet animation
func RenderWelcome(state *WelcomeAnimationState) string {
	var s strings.Builder

	// Both Word and Killer use the same red pulsing effect
	wordText := renderPulsingKiller("Word", state.Frame)
	killerText := renderPulsingKiller("Killer", state.Frame)

	s.WriteString("\n\n")
	s.WriteString(titleStyle.Render("    ╔══════════════════════════════════════════════╗") + "\n")
	s.WriteString(titleStyle.Render("    ║                                              ║") + "\n")

	// Title row (row 0)
	titleRow := renderTitleRow(wordText, killerText, state, 0)
	s.WriteString(titleRow)

	// Empty row (row 1) - only one empty row after title
	emptyRow1 := renderEmptyRow(state, 1)
	s.WriteString(emptyRow1)

	// Tagline row (row 2) - explosion happens here (moved from row 3)
	taglineRow := renderTaglineRow(state)
	s.WriteString(taglineRow)

	s.WriteString(titleStyle.Render("    ║                                              ║") + "\n")

	// Options row 1: start
	optionRow1 := renderOptionLine(state, 0)
	s.WriteString(optionRow1)

	// Options row 2: quit
	optionRow2 := renderOptionLine(state, 1)
	s.WriteString(optionRow2)

	// Empty row after quit
	s.WriteString(titleStyle.Render("    ║                                              ║") + "\n")

	s.WriteString(titleStyle.Render("    ╚══════════════════════════════════════════════╝") + "\n")

	return s.String()
}

// renderOptionsRow renders the options row with selection
func renderOptionsRow(state *WelcomeAnimationState) string {
	// No longer used - replaced by renderOptionLine
	return ""
}

// renderOptionLine renders a single option line
func renderOptionLine(state *WelcomeAnimationState, optionIndex int) string {
	const boxWidth = 46

	options := []string{"start", "quit"}
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")). // Bright yellow
		Bold(true)
	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	// Build the option text with indicator
	var optionDisplay string
	if state.SelectedOption == optionIndex {
		optionDisplay = "> " + options[optionIndex] + " <"
	} else {
		optionDisplay = "  " + options[optionIndex] + "  "
	}

	// Apply style
	var styledText string
	if state.SelectedOption == optionIndex {
		styledText = selectedStyle.Render(optionDisplay)
	} else {
		styledText = normalStyle.Render(optionDisplay)
	}

	// Use lipgloss to center the text within the box
	alignedText := lipgloss.NewStyle().
		Width(boxWidth).
		Align(lipgloss.Center).
		Render(styledText)

	return titleStyle.Render("    ║") + alignedText + titleStyle.Render("║") + "\n"
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

	// Menu options
	options := []string{" Resume Game", " Quit Game"}
	for i, opt := range options {
		if i == selectedIndex {
			lines = append(lines, "                           "+menuSelectedStyle.Render("> "+opt))
		} else {
			lines = append(lines, "                             "+menuNormalStyle.Render(opt))
		}
	}

	// Fill to exactly 12 lines (title + empty + 10 content rows)
	// Current: 1 title + 1 empty + 5 box + 1 empty + 2 menu = 10 lines
	// Need 2 more lines
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

// RenderResults renders game results with consistent layout
func RenderResults(stats GameStats, aborted bool) string {
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
	statsArea := renderResultsArea(stats, aborted)
	s.WriteString(statsArea)
	s.WriteString("\n")

	// === BOTTOM: Exit Hint ===
	hints := inputBoxStyle.Render("Press any key to exit...")
	s.WriteString(hints)
	s.WriteString("\n")

	return s.String()
}

// renderResultsArea renders the statistics area
func renderResultsArea(stats GameStats, aborted bool) string {
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
	// Keystroke stats
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Letters/second:"),
		statValueStyle.Render(fmt.Sprintf("%7.2f", stats.LettersPerSecond))))
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Words/second:"),
		statValueStyle.Render(fmt.Sprintf("%7.2f", stats.WordsPerSecond))))
	content.WriteString(fmt.Sprintf("%50s %s\n",
		statItemStyle.Render("Correct Chars:"),
		statValueStyle.Render(fmt.Sprintf("%7d", stats.CorrectChars))))
	content.WriteString(fmt.Sprintf("%50s %s",
		statItemStyle.Render("Accuracy:"),
		statValueStyle.Render(fmt.Sprintf("%6.2f%%", stats.AccuracyPercent))))

	return wordBoxStyle.Render(content.String())
}

// renderTitleRow renders the title row with bullet if present
func renderTitleRow(wordText, killerText string, state *WelcomeAnimationState, row int) string {
	// Internal box width is 46 characters
	const boxWidth = 46

	// Combine Word + space + Killer
	titleText := wordText + " " + killerText

	// Use lipgloss to center the text within the box
	// lipgloss handles ANSI codes correctly
	centeredText := lipgloss.NewStyle().
		Width(boxWidth).
		Align(lipgloss.Center).
		Render(titleText)

	return titleStyle.Render("    ║") + centeredText + titleStyle.Render("║") + "\n"
}

// renderEmptyRow renders an empty row with bullet if present
func renderEmptyRow(state *WelcomeAnimationState, row int) string {
	const boxWidth = 46

	// Empty content (all spaces)
	contentStr := strings.Repeat(" ", boxWidth)

	// Check if bullet is on this row
	if state.BulletActive && state.BulletY == row {
		// Add bullet at current position
		bulletChar := renderBullet()
		bulletPos := state.BulletX
		// Ensure bullet position is within bounds
		if bulletPos >= 0 && bulletPos < boxWidth && bulletPos < len(contentStr) {
			// Replace character at bullet position with bullet
			contentStr = contentStr[:bulletPos] + bulletChar + contentStr[bulletPos+1:]
		}
	}

	return titleStyle.Render("    ║") + contentStr + titleStyle.Render("║") + "\n"
}

// renderTaglineRow renders the tagline row with explosion effect
func renderTaglineRow(state *WelcomeAnimationState) string {
	const boxWidth = 46
	taglineText := "Low-key but never simple"

	// Calculate tagline position (centered in the box)
	taglineLen := len(taglineText)
	taglineStartPadding := (boxWidth - taglineLen) / 2
	rightPadding := boxWidth - taglineLen - taglineStartPadding

	// Check if bullet is at row 2 (tagline row) and at tagline text position
	if state.BulletActive && state.BulletY == 2 && !state.ExplosionTriggered {
		// Check if bullet position intersects with tagline text
		if state.BulletX >= taglineStartPadding && state.BulletX < taglineStartPadding+taglineLen {
			// Trigger explosion when bullet reaches tagline text
			state.BulletActive = false
			state.Exploding = true
			state.ExplosionTime = time.Now()
			state.TaglineWord = TaglineInfo{
				Text:        taglineText,
				Completed:   true,
				CompletedAt: state.ExplosionTime,
			}
			state.ExplosionTriggered = true
		}
	}

	// Render tagline text with explosion effect if exploding
	var contentStr string
	if state.Exploding {
		// Use renderCompletedWordAnimation logic for explosion
		explosionText := renderTaglineExplosion(state.TaglineWord)
		// Pad the explosion text to maintain box width
		leftPadding := strings.Repeat(" ", taglineStartPadding)
		rightPaddingStr := strings.Repeat(" ", rightPadding)
		contentStr = leftPadding + explosionText + rightPaddingStr
	} else {
		// Normal tagline rendering (centered)
		leftPadding := strings.Repeat(" ", taglineStartPadding)
		rightPaddingStr := strings.Repeat(" ", rightPadding)
		contentStr = leftPadding + taglineText + rightPaddingStr
	}

	// Check if bullet is on this row (after explosion check, so bullet doesn't show during explosion)
	if state.BulletActive && state.BulletY == 2 {
		// Add bullet at current position
		bulletChar := renderBullet()
		bulletPos := state.BulletX
		// Ensure bullet position is within bounds
		if bulletPos >= 0 && bulletPos < boxWidth && bulletPos < len(contentStr) {
			// Replace character at bullet position with bullet
			contentStr = contentStr[:bulletPos] + bulletChar + contentStr[bulletPos+1:]
		}
	}

	return titleStyle.Render("    ║") + contentStr + titleStyle.Render("║") + "\n"
}

// renderTaglineExplosion renders tagline explosion effect (similar to renderCompletedWordAnimation)
func renderTaglineExplosion(tagline TaglineInfo) string {
	timeSinceCompletion := time.Since(tagline.CompletedAt)
	msElapsed := timeSinceCompletion.Milliseconds()

	// Phase 1: Hit effect (0-150ms) - Bright flash
	if msElapsed < 150 {
		// Alternate between bright yellow and bright green for impact
		if msElapsed < 50 {
			return hitEffectStyle.Render(tagline.Text)
		} else if msElapsed < 100 {
			return fadingStyle1.Render(tagline.Text)
		} else {
			return hitEffectStyle.Render(tagline.Text)
		}
	}

	// Phase 2: Letter-by-letter fade (150ms onwards)
	// Each letter takes 80ms to fade through colors
	const letterFadeTime = 80  // ms per letter
	const animationStart = 150 // when fade animation starts

	var result strings.Builder
	wordLen := len(tagline.Text)

	for i, ch := range tagline.Text {
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

	// After complete fade, show in dark gray
	totalAnimTime := animationStart + int64(wordLen*letterFadeTime) + 80
	if msElapsed >= totalAnimTime {
		return completedWordStyle.Render(tagline.Text)
	}

	return result.String()
}

// renderBullet renders the bullet character
func renderBullet() string {
	bulletStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")). // Bright yellow
		Bold(true)
	return bulletStyle.Render("►")
}

// renderExplodingText renders text with explosion effect (similar to renderCompletedWordAnimation)
func renderExplodingText(text string, framesIn int) string {
	// Explosion phase 1: Hit flash (0-10 frames) - bright yellow with background
	if framesIn < 10 {
		hitStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")). // Bright yellow
			Background(lipgloss.Color("52")). // Red background
			Bold(true)
		return hitStyle.Render(text)
	}

	// Explosion phase 2: Letter-by-letter fade (10+ frames)
	const letterFadeTime = 8 // ms per letter equivalent in frames
	const animationStart = 10

	var result strings.Builder

	for i, ch := range text {
		letterStartTime := animationStart + i*letterFadeTime
		letterElapsed := framesIn - letterStartTime

		var charStyle lipgloss.Style

		if letterElapsed < 0 {
			// Letter hasn't started exploding yet
			charStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Background(lipgloss.Color("52")).Bold(true)
		} else if letterElapsed < 2 {
			charStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Background(lipgloss.Color("52")).Bold(true)
		} else if letterElapsed < 4 {
			charStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Background(lipgloss.Color("52")).Bold(true) // Green on red
		} else if letterElapsed < 6 {
			charStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Background(lipgloss.Color("52")) // Medium green on red
		} else if letterElapsed < 8 {
			charStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Background(lipgloss.Color("52")) // Light gray on red
		} else {
			charStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("52")) // Dark gray on red
		}

		result.WriteString(charStyle.Render(string(ch)))
	}

	return result.String()
}

// UpdateWelcomeAnimation updates the welcome animation state
func UpdateWelcomeAnimation(state *WelcomeAnimationState) {
	state.Frame++

	const boxWidth = 46

	// Bullet firing cycle
	const bulletFireInterval = 30 // frames between shots (about 3 seconds)
	const bulletSpeed = 3         // pixels per frame (faster)

	// Check if explosion is done and reset triggered flag
	if state.Exploding {
		// Calculate time since explosion
		timeSinceExplosion := time.Since(state.ExplosionTime)
		// Total animation time = 150ms hit effect + (24 letters * 80ms) + 80ms = about 2150ms
		const explosionDuration = 2200 * time.Millisecond
		if timeSinceExplosion >= explosionDuration {
			// Explosion done, reset all explosion state
			state.Exploding = false
			state.ExplosionTriggered = false
		}
	}

	if !state.BulletActive && !state.Exploding {
		// Check if it's time to fire a new bullet
		if state.Frame%bulletFireInterval == 0 {
			// Fire bullet from row 1, first position after ║
			state.BulletX = 0 // first position in content area
			state.BulletY = 1 // row below "Word Killer"
			state.BulletActive = true
		}
	}

	if state.BulletActive {
		// Move bullet to the right
		state.BulletX += bulletSpeed

		// Check if bullet went off screen (width is 46)
		// Use boxWidth - 1 to leave room for the border
		if state.BulletX >= boxWidth-1 {
			// Wrap to next row from left side
			state.BulletY++
			state.BulletX = 1 // start from left (offset by border)

			// Don't let it continue past row 2 (tagline row)
			if state.BulletY > 2 {
				state.BulletActive = false
			}
		}
	}
}

// renderShakingWord renders "Word" with random shaking effect
// Each character independently shakes with random horizontal offset
func renderShakingWord(text string, frame int) string {
	// Change shake pattern every 5 frames for more dynamic effect
	shakePattern := frame / 5

	var result strings.Builder

	for i, ch := range text {
		// Generate different offset for each character and pattern
		// Use different multipliers to create chaotic shaking
		offset1 := (shakePattern + i*3) % 3
		offset2 := (shakePattern + i*7) % 3

		// -1, 0, or 1 spaces before the character
		beforeSpaces := offset1 - 1
		// 0, 1, or 2 spaces after the character
		afterSpaces := offset2 + 1

		// Add before spaces (can't go negative, so min 0)
		if beforeSpaces > 0 {
			result.WriteString(strings.Repeat(" ", beforeSpaces))
		}

		// Add the character
		result.WriteString(titleStyle.Render(string(ch)))

		// Add after spaces
		if i < len(text)-1 {
			result.WriteString(strings.Repeat(" ", afterSpaces))
		}
	}

	return result.String()
}

// renderPulsingKiller renders "Killer" with red breathing pulse effect
// Color cycles from dark red to bright red and back
func renderPulsingKiller(text string, frame int) string {
	// Pulse cycle: 40 frames for full cycle (about 4 seconds at 10fps)
	const pulseCycle = 40
	cyclePos := frame % pulseCycle

	// Calculate pulse intensity using sine wave for smooth breathing effect
	angle := float64(cyclePos) / float64(pulseCycle) * 2 * math.Pi
	intensity := (math.Sin(angle-math.Pi/2) + 1) / 2 // 0.0 to 1.0

	// Map intensity to ANSI color (red range: 88-196)
	// 88 = dark red, 196 = bright red
	colorCode := 88 + int(intensity*108)
	color := lipgloss.Color(fmt.Sprintf("%d", colorCode))

	// Create pulsing style with bold
	pulseStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true)

	return pulseStyle.Render(text)
}
