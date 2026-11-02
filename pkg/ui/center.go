package ui

import "strings"

// screenWidth is the outer width of the standard layout: content plus borders.
// A fixed value keeps the left edge from shifting between screens.
const screenWidth = contentWidth + 2

// CenterHorizontally shifts a rendered screen to the middle of the terminal.
// Every line gets the same padding so the layout inside stays aligned.
func CenterHorizontally(view string, terminalWidth int) string {
	padding := (terminalWidth - screenWidth) / 2
	if padding <= 0 {
		return view
	}

	prefix := strings.Repeat(" ", padding)
	lines := strings.Split(view, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}
