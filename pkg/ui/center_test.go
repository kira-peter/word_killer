package ui

import "testing"

func TestCenterHorizontally(t *testing.T) {
	view := "abc\n\n  de\n"

	if got := CenterHorizontally(view, screenWidth+10); got != "     abc\n\n       de\n" {
		t.Errorf("wide terminal: got %q", got)
	}
	for _, width := range []int{0, screenWidth - 10, screenWidth, screenWidth + 1} {
		if got := CenterHorizontally(view, width); got != view {
			t.Errorf("width %d should leave the view unchanged: got %q", width, got)
		}
	}
}
