package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestTerminalTooSmall(t *testing.T) {
	cases := []struct {
		w, h int
		want bool
	}{
		{0, 0, false}, // size not known yet
		{MinTerminalWidth, MinTerminalHeight, false},
		{200, 60, false},
		{MinTerminalWidth - 1, MinTerminalHeight, true},
		{MinTerminalWidth, MinTerminalHeight - 1, true},
		{80, 24, true},
	}
	for _, c := range cases {
		if got := TerminalTooSmall(c.w, c.h); got != c.want {
			t.Errorf("TerminalTooSmall(%d, %d) = %v, want %v", c.w, c.h, got, c.want)
		}
	}
}

func TestRenderTooSmallFits(t *testing.T) {
	// The warning must fit the small window it is shown in
	for _, size := range [][2]int{{80, 24}, {40, 12}, {82, 31}} {
		view := RenderTooSmall(size[0], size[1])
		if w, h := lipgloss.Width(view), lipgloss.Height(view); w > size[0] || h > size[1] {
			t.Errorf("%dx%d: warning is %dx%d", size[0], size[1], w, h)
		}
		for _, want := range []string{"Window too small", "82 x 32", "start.bat"} {
			if !strings.Contains(view, want) {
				t.Errorf("%dx%d: missing %q:\n%s", size[0], size[1], want, view)
			}
		}
	}
}
