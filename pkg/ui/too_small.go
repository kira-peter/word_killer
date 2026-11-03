package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// The largest screen (the Rhythm Dance results) is 82 columns by 32 rows;
// a smaller terminal would clip or wrap it
const (
	MinTerminalWidth  = screenWidth
	MinTerminalHeight = 32
)

// TerminalTooSmall reports whether the game cannot be drawn in a terminal of
// this size. A size of zero means it is not known yet, which is not too small.
func TerminalTooSmall(width, height int) bool {
	if width == 0 || height == 0 {
		return false
	}
	return width < MinTerminalWidth || height < MinTerminalHeight
}

// RenderTooSmall asks the player to enlarge the window. It is kept narrow so
// it fits the very terminal that is too small for the game.
func RenderTooSmall(width, height int) string {
	warn := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	ok := highlightStyle
	bad := onlineErrorStyle

	sizeText := func(value, min int) string {
		if value < min {
			return bad.Render(fmt.Sprint(value))
		}
		return ok.Render(fmt.Sprint(value))
	}

	lines := []string{
		warn.Render("Window too small"),
		"",
		statItemStyle.Render(fmt.Sprintf("Need  %d x %d", MinTerminalWidth, MinTerminalHeight)),
		statItemStyle.Render("Now   ") + sizeText(width, MinTerminalWidth) + statItemStyle.Render(" x ") + sizeText(height, MinTerminalHeight),
		"",
		statsStyle.Render("Enlarge the window"),
		statsStyle.Render("or run start.bat"),
	}
	block := strings.Join(lines, "\n")
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, block)
}
