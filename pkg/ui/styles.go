package ui

import (
	"fmt"
	"strings"

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
			Padding(0, 2).
			MarginTop(1)

	wordBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
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
func RenderGame(words []string, highlightedIndices []int, input string, stats GameStats) string {
	var s strings.Builder

	// === TOP: Status Bar ===
	statusBar := renderStatusBar(stats, len(words))
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
	var items []string

	// Time
	items = append(items, fmt.Sprintf("%s %s",
		statItemStyle.Render("Time:"),
		statValueStyle.Render(fmt.Sprintf("%.1fs", stats.ElapsedSeconds))))

	// Progress
	items = append(items, fmt.Sprintf("%s %s",
		statItemStyle.Render("Progress:"),
		statValueStyle.Render(fmt.Sprintf("%d/%d", stats.WordsCompleted, stats.WordsCompleted+remainingWords))))

	// Speed
	items = append(items, fmt.Sprintf("%s %s",
		statItemStyle.Render("Speed:"),
		statValueStyle.Render(fmt.Sprintf("%.1f l/s", stats.LettersPerSecond))))

	// Accuracy
	items = append(items, fmt.Sprintf("%s %s",
		statItemStyle.Render("Accuracy:"),
		statValueStyle.Render(fmt.Sprintf("%.1f%%", stats.AccuracyPercent))))

	statusLine := strings.Join(items, "  │  ")
	return headerStyle.Render(statusLine)
}

// renderWordArea renders the middle word list area
func renderWordArea(words []string, highlightedIndices []int, input string) string {
	if len(words) == 0 {
		return wordBoxStyle.Render(statsStyle.Render("All words completed! Press Enter to finish..."))
	}

	var wordLines []string
	wordLines = append(wordLines, titleStyle.Render("Words:"))
	wordLines = append(wordLines, "")

	// Render words in columns for better space usage
	const wordsPerRow = 3
	for i := 0; i < len(words); i += wordsPerRow {
		var rowWords []string
		for j := 0; j < wordsPerRow && i+j < len(words); j++ {
			idx := i + j
			word := words[idx]

			// Check if highlighted
			isHighlighted := false
			for _, hIdx := range highlightedIndices {
				if hIdx == idx {
					isHighlighted = true
					break
				}
			}

			var renderedWord string
			if isHighlighted && len(input) > 0 {
				// Highlight matched part
				matchLen := len(input)
				if matchLen > len(word) {
					matchLen = len(word)
				}
				renderedWord = highlightStyle.Render(word[:matchLen]) + wordStyle.Render(word[matchLen:])
			} else {
				renderedWord = wordStyle.Render(word)
			}

			// Pad word to fixed width for alignment
			renderedWord = fmt.Sprintf("%-20s", renderedWord)
			rowWords = append(rowWords, renderedWord)
		}
		wordLines = append(wordLines, "  "+strings.Join(rowWords, "  "))
	}

	content := strings.Join(wordLines, "\n")
	return wordBoxStyle.Render(content)
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
