package ui

import (
	"strings"
	"testing"
)

// TestRhythmMasterProgressBar used to panic on a zero time limit
func TestRhythmMasterProgressBar(t *testing.T) {
	words := []WordInfo{{Text: "hello"}, {Text: "world"}}
	cases := []struct {
		name      string
		remaining float64
		limit     float64
		filled    int
	}{
		{"zero limit", 0, 0, 0},
		{"zero limit with time left", 1, 0, 0},
		{"full", 2, 2, 16},
		{"half", 1, 2, 8},
		{"overtime", -1, 2, 0},
		{"more than limit", 5, 2, 16},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			view := RenderRhythmMasterGame(words, nil, "", GameStats{}, c.remaining, c.limit, 0, 0)
			if got := strings.Count(view, "█"); got != c.filled {
				t.Errorf("filled cells = %d, want %d", got, c.filled)
			}
			if got := strings.Count(view, "█") + strings.Count(view, "░"); got != 16 {
				t.Errorf("bar width = %d, want 16", got)
			}
		})
	}
}
