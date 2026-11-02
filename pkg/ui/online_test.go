package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/word-killer/word-killer/pkg/online"
)

func testRooms(n int) []online.RoomInfo {
	rooms := make([]online.RoomInfo, n)
	for i := range rooms {
		rooms[i] = online.RoomInfo{ID: fmt.Sprint(i), HostName: fmt.Sprintf("Host%02d", i), IP: "10.0.0.1", Port: 7788, Players: 1, MaxPlayers: 4}
	}
	return rooms
}

func TestOnlineMenuKeepsStandardHeight(t *testing.T) {
	want := lipgloss.Height(RenderOnlineMenu(nil, 0, "", false, 0))
	for _, n := range []int{0, 1, maxListedRooms, maxListedRooms + 1, 20} {
		rooms := testRooms(n)
		for _, selected := range []int{0, n} {
			view := RenderOnlineMenu(rooms, selected, "Cannot connect to host", false, 0)
			if got := lipgloss.Height(view); got != want {
				t.Errorf("%d rooms, selected %d: height %d, want %d", n, selected, got, want)
			}
		}
	}
}

func TestOnlineMenuScrollsToSelection(t *testing.T) {
	rooms := testRooms(20)

	view := RenderOnlineMenu(rooms, 0, "", false, 0)
	if !strings.Contains(view, "Host00") || strings.Contains(view, "Host19") || !strings.Contains(view, "14 more room(s)") {
		t.Errorf("top of list:\n%s", view)
	}

	view = RenderOnlineMenu(rooms, 20, "", false, 0)
	if !strings.Contains(view, "> Host19") || strings.Contains(view, "Host00") {
		t.Errorf("selection at the end must be visible:\n%s", view)
	}
}
