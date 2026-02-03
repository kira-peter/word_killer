package ui

import (
	"fmt"

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

// RenderGame renders game screen
func RenderGame(words []string, highlightedIndices []int, input string, stats GameStats) string {
	var s string

	// Title
	s += titleStyle.Render("Word Killer - Classic Mode") + "\n\n"

	// Word list
	s += statsStyle.Render("Remaining Words:") + "\n"
	for i, word := range words {
		// Check if highlighted
		isHighlighted := false
		for _, idx := range highlightedIndices {
			if idx == i {
				isHighlighted = true
				break
			}
		}

		if isHighlighted {
			// Highlight matched part
			matchLen := len(input)
			if matchLen > len(word) {
				matchLen = len(word)
			}
			s += "  " + highlightStyle.Render(word[:matchLen]) + wordStyle.Render(word[matchLen:]) + "\n"
		} else {
			s += "  " + wordStyle.Render(word) + "\n"
		}
	}

	// Current input
	s += "\n" + statsStyle.Render("Current Input: ") + inputStyle.Render(input) + "\n"

	// Statistics
	s += "\n" + titleStyle.Render("--- Statistics ---") + "\n"
	s += fmt.Sprintf("Completed: %d | Remaining: %d | Time: %.1fs\n",
		stats.WordsCompleted, len(words), stats.ElapsedSeconds)
	s += fmt.Sprintf("Speed: %.2f letters/s | %.2f words/s\n",
		stats.LettersPerSecond, stats.WordsPerSecond)
	s += fmt.Sprintf("Accuracy: %.2f%%\n", stats.AccuracyPercent)

	// Hints
	s += "\n" + hintStyle.Render("[ESC] Pause") + "\n"

	return s
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
	s += fmt.Sprintf("Total Keystrokes:   %d\n", stats.TotalKeystrokes)
	s += fmt.Sprintf("Valid Keystrokes:   %d\n", stats.ValidKeystrokes)
	s += fmt.Sprintf("Correct Chars:      %d\n", stats.CorrectChars)
	s += fmt.Sprintf("Completed Words:    %d\n", stats.WordsCompleted)
	s += fmt.Sprintf("Total Letters:      %d\n", stats.TotalLetters)
	s += fmt.Sprintf("Total Time:         %.2f seconds\n", stats.ElapsedSeconds)
	s += "\n"
	s += highlightStyle.Render("Speed:") + "\n"
	s += fmt.Sprintf("Letters/second:     %.2f\n", stats.LettersPerSecond)
	s += fmt.Sprintf("Words/second:       %.2f\n", stats.WordsPerSecond)
	s += "\n"
	s += highlightStyle.Render("Accuracy: ") + fmt.Sprintf("%.2f%%\n", stats.AccuracyPercent)

	s += "\n" + hintStyle.Render("Press any key to exit...") + "\n"

	return s
}
