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

// RenderWelcome renders welcome screen
func RenderWelcome() string {
	var s string

	s += "\n\n"
	s += titleStyle.Render("    ╔════════════════════════════╗") + "\n"
	s += titleStyle.Render("    ║                            ║") + "\n"
	s += titleStyle.Render("    ║         Word Killer        ║") + "\n"
	s += titleStyle.Render("    ║                            ║") + "\n"
	s += titleStyle.Render("    ╚════════════════════════════╝") + "\n"
	s += "\n"
	s += statsStyle.Render("Low-key but never simple") + "\n\n"
	s += highlightStyle.Render("Press [Enter] to start") + "\n"
	s += hintStyle.Render("Press [ESC] to quit") + "\n\n"

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

// renderWordArea renders the middle word list area
func renderWordArea(words []WordInfo, highlightedIndices []int, input string) string {
	if len(words) == 0 {
		content := statsStyle.Render("All words completed!")
		return wordBoxStyle.Render(content)
	}

	var wordLines []string
	wordLines = append(wordLines, titleStyle.Render("Words:"))
	wordLines = append(wordLines, "")

	// Calculate optimal columns based on content width
	// Reserve space for padding and borders (about 8 chars)
	availableWidth := contentWidth - 8
	const wordColumnWidth = 18 // Each word gets 18 chars (word + spacing)
	wordsPerRow := availableWidth / wordColumnWidth
	if wordsPerRow < 1 {
		wordsPerRow = 1
	}

	// Render words in columns
	for i := 0; i < len(words); i += wordsPerRow {
		var rowWords []string
		for j := 0; j < wordsPerRow && i+j < len(words); j++ {
			idx := i + j
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

			// Pad word to fixed width for alignment (use plain text length, not styled length)
			paddedWord := padToWidth(wordInfo.Text, renderedWord, wordColumnWidth)
			rowWords = append(rowWords, paddedWord)
		}
		wordLines = append(wordLines, "  "+strings.Join(rowWords, ""))
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
	const letterFadeTime = 80 // ms per letter
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

// RenderPauseMenu renders pause menu
func RenderPauseMenu(selectedIndex int) string {
	var s string

	s += "\n\n"
	s += titleStyle.Render("    ╔════════════════════╗") + "\n"
	s += titleStyle.Render("    ║   Game Paused      ║") + "\n"
	s += titleStyle.Render("    ╚════════════════════╝") + "\n"
	s += "\n\n"

	options := []string{"Resume", "Quit"}
	for i, opt := range options {
		if i == selectedIndex {
			s += "    " + menuSelectedStyle.Render("> "+opt) + "\n"
		} else {
			s += "      " + menuNormalStyle.Render(opt) + "\n"
		}
	}

	s += "\n" + hintStyle.Render("[↑↓] Select | [Enter] Confirm | [ESC] Quit Game") + "\n"

	return s
}

// RenderResults renders game results
func RenderResults(stats GameStats, aborted bool) string {
	var s string

	s += "\n\n"
	s += titleStyle.Render("    ╔═══════════════════════╗") + "\n"
	if aborted {
		s += titleStyle.Render("    ║     Game Over         ║") + "\n"
	} else {
		s += titleStyle.Render("    ║   Congratulations!    ║") + "\n"
	}
	s += titleStyle.Render("    ╚═══════════════════════╝") + "\n"
	s += "\n\n"

	s += statsStyle.Render("=== Final Statistics ===") + "\n\n"

	// Main stats
	s += fmt.Sprintf("%s %s\n",
		statItemStyle.Render("Total Keystrokes:  "),
		statValueStyle.Render(fmt.Sprintf("%d", stats.TotalKeystrokes)))
	s += fmt.Sprintf("%s %s\n",
		statItemStyle.Render("Valid Keystrokes:  "),
		statValueStyle.Render(fmt.Sprintf("%d", stats.ValidKeystrokes)))
	s += fmt.Sprintf("%s %s\n",
		statItemStyle.Render("Correct Chars:     "),
		statValueStyle.Render(fmt.Sprintf("%d", stats.CorrectChars)))
	s += fmt.Sprintf("%s %s\n",
		statItemStyle.Render("Completed Words:   "),
		statValueStyle.Render(fmt.Sprintf("%d", stats.WordsCompleted)))
	s += fmt.Sprintf("%s %s\n",
		statItemStyle.Render("Total Letters:     "),
		statValueStyle.Render(fmt.Sprintf("%d", stats.TotalLetters)))
	s += fmt.Sprintf("%s %s\n",
		statItemStyle.Render("Total Time:        "),
		statValueStyle.Render(fmt.Sprintf("%.2f seconds", stats.ElapsedSeconds)))

	s += "\n"
	s += highlightStyle.Render("Performance:") + "\n"
	s += fmt.Sprintf("%s %s\n",
		statItemStyle.Render("Letters/second:    "),
		statValueStyle.Render(fmt.Sprintf("%.2f", stats.LettersPerSecond)))
	s += fmt.Sprintf("%s %s\n",
		statItemStyle.Render("Words/second:      "),
		statValueStyle.Render(fmt.Sprintf("%.2f", stats.WordsPerSecond)))
	s += "\n"
	s += fmt.Sprintf("%s %s\n",
		highlightStyle.Render("Accuracy:          "),
		statValueStyle.Render(fmt.Sprintf("%.2f%%", stats.AccuracyPercent)))

	s += "\n" + hintStyle.Render("Press any key to exit...") + "\n"

	return s
}
