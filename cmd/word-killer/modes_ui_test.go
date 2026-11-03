package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/word-killer/word-killer/pkg/online"
	"github.com/word-killer/word-killer/pkg/ui"
)

// moveLobbyTo moves the host's lobby selection to row index
func moveLobbyTo(t *testing.T, m model, index int) model {
	t.Helper()
	for m.online.lobbyIndex > index {
		m, _ = pressKey(m, tea.KeyUp)
	}
	for m.online.lobbyIndex < index {
		before := m.online.lobbyIndex
		m, _ = pressKey(m, tea.KeyDown)
		if m.online.lobbyIndex == before {
			t.Fatalf("cannot reach lobby row %d", index)
		}
	}
	return m
}

// startFromLobby selects Start Game, the last lobby row, and presses Enter
func startFromLobby(t *testing.T, m model) model {
	t.Helper()
	m = moveLobbyTo(t, m, len(ui.LobbyItems(m.online.battle))-1)
	m, _ = pressKey(m, tea.KeyEnter)
	return m
}

// setLobbyOption opens a lobby row's wheel, scrolls to want and confirms
func setLobbyOption(t *testing.T, host, guest *model, row int, want string, check func() bool) {
	t.Helper()
	*host = moveLobbyTo(t, *host, row)
	*host, _ = pressKey(*host, tea.KeyEnter)
	w := host.online.wheel
	if w == nil {
		t.Fatalf("row %d did not open a wheel", row)
	}
	target := -1
	for i, s := range w.steps {
		if s == want {
			target = i
		}
	}
	if target < 0 {
		t.Fatalf("wheel %s has no step %q: %v", w.label, want, w.steps)
	}
	for w.index < target {
		*host, _ = pressKey(*host, tea.KeyDown)
	}
	for w.index > target {
		*host, _ = pressKey(*host, tea.KeyUp)
	}
	*host, _ = pressKey(*host, tea.KeyEnter)
	if host.online.wheel != nil {
		t.Fatal("Enter should close the wheel")
	}
	runUntil(t, host, guest, "setting "+want, check)
}

// lineContaining returns the index of the first view line holding text, or -1.
// The column must match too, so the line is compared up to and including text.
func lineContaining(view, text string) int {
	for i, line := range strings.Split(view, "\n") {
		if strings.Contains(line, text) {
			return i*1000 + strings.Index(line, text)
		}
	}
	return -1
}

// newRoomPair creates a room with Hosty and joins Guesty
func newRoomPair(t *testing.T) (model, model) {
	t.Helper()
	host := openOnlineMenu(t, newTestModel(t, "Hosty"))
	host, _ = pressKey(host, tea.KeyEnter)
	guest := joinRoom(t, openOnlineMenu(t, newTestModel(t, "Guesty")), "Hosty")
	runUntil(t, &host, &guest, "both in lobby", func() bool {
		return len(guest.online.battle.Players) == 2 && len(host.online.battle.Players) == 2
	})
	return host, guest
}

func TestLobbyWheel(t *testing.T) {
	host, guest := newRoomPair(t)

	view := host.View()
	for _, want := range []string{"> Mode", "Word Drop", "Words", "10", "Time limit", "120s", "Start Game"} {
		if !strings.Contains(view, want) {
			t.Errorf("lobby missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(guest.View(), "> Mode") {
		t.Error("guests see no selection marker")
	}

	// The wheel opens in place: the label keeps its exact spot, the other
	// settings are hidden, and players stay visible on the right
	host = moveLobbyTo(t, host, 1)
	before := lineContaining(host.View(), "> Words")
	host, _ = pressKey(host, tea.KeyEnter)
	view = host.View()
	after := lineContaining(view, "> Words")
	if before < 0 || after != before {
		t.Errorf("label moved from line %d to %d:\n%s", before, after, view)
	}
	for _, want := range []string{"▶ 10 ◀", "5", "15", "20", "Scroll", "Guesty"} {
		if !strings.Contains(view, want) {
			t.Errorf("wheel missing %q:\n%s", want, view)
		}
	}
	for _, hidden := range []string{"Time limit", "Start Game", "> Mode"} {
		if strings.Contains(view, hidden) {
			t.Errorf("wheel should hide %q:\n%s", hidden, view)
		}
	}

	// Scrolling stops at the ends and stays local until confirmed
	for i := 0; i < 10; i++ {
		host, _ = pressKey(host, tea.KeyDown)
	}
	if host.online.wheel.steps[host.online.wheel.index] != "30" {
		t.Errorf("wheel should stop at the last step")
	}
	host, _ = pressKey(host, tea.KeyEsc)
	if host.online.wheel != nil || host.online.screen != onlineRoom {
		t.Fatal("Esc should only close the wheel")
	}
	host = tick(host)
	if host.online.battle.Settings.Count != 10 {
		t.Errorf("Esc must discard the change, count = %d", host.online.battle.Settings.Count)
	}

	setLobbyOption(t, &host, &guest, 1, "20", func() bool { return guest.online.battle.Settings.Count == 20 })
	setLobbyOption(t, &host, &guest, 0, "Word Pool", func() bool {
		return guest.online.battle.Mode == online.ModeWordPool
	})
	for _, want := range []string{"Word Pool", "Duration", "Words/sec", "Steal", "Off"} {
		if !strings.Contains(guest.View(), want) {
			t.Errorf("guest lobby missing %q:\n%s", want, guest.View())
		}
	}

	// Guests cannot change anything
	guest = moveLobbyTo(t, guest, 0)
	guest, _ = pressKey(guest, tea.KeyEnter)
	if guest.online.wheel != nil {
		t.Error("guests must not open wheels")
	}
}

func TestKickFromLobby(t *testing.T) {
	host, guest := newRoomPair(t)

	host, _ = pressKey(host, tea.KeyRunes)
	next, _ := host.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("K")})
	host = next.(model)
	if host.online.kickIDs == nil || !strings.Contains(host.View(), "Kick which player?") || !strings.Contains(host.View(), "✕ Guesty") {
		t.Fatalf("kick picker not open:\n%s", host.View())
	}
	// Esc closes the picker without leaving the room
	host, _ = pressKey(host, tea.KeyEsc)
	if host.online.kickIDs != nil || host.online.screen != onlineRoom {
		t.Fatal("Esc should only close the picker")
	}

	next, _ = host.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("K")})
	host = next.(model)
	host, _ = pressKey(host, tea.KeyEnter)
	runUntil(t, &host, &guest, "guest kicked", func() bool {
		return guest.online.screen == onlineMenu && len(host.online.battle.Players) == 1
	})
	if guest.online.notice != "You were kicked" || guest.online.browser == nil {
		t.Errorf("guest notice=%q browsing=%v", guest.online.notice, guest.online.browser != nil)
	}

	// Alone again, there is no one to kick
	next, _ = host.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("K")})
	host = next.(model)
	if host.online.kickIDs != nil || host.online.notice != "No one to kick" {
		t.Errorf("picker=%v notice=%q", host.online.kickIDs, host.online.notice)
	}

	// The kicked player can come straight back
	guest = joinRoom(t, guest, "Hosty")
	runUntil(t, &host, &guest, "guest back", func() bool { return len(host.online.battle.Players) == 2 })
}

func TestSentenceRaceUI(t *testing.T) {
	host, guest := newRoomPair(t)
	setLobbyOption(t, &host, &guest, 0, "Sentence Race", func() bool {
		return guest.online.battle.Mode == online.ModeSentenceRace
	})
	setLobbyOption(t, &host, &guest, 1, "3", func() bool { return guest.online.battle.Settings.Count == 3 })
	host = startFromLobby(t, host)

	for i := 0; i < 3; i++ {
		runUntil(t, &host, &guest, "sentence", func() bool {
			b := guest.online.battle
			return b.Phase == online.BattleFalling && b.WordIndex == i && host.online.battle.WordIndex == i &&
				host.online.battle.Phase == online.BattleFalling
		})
		sentence := guest.online.battle.Word
		if view := guest.View(); !strings.Contains(view, strings.Fields(sentence)[0]) || !strings.Contains(view, "left") {
			t.Errorf("sentence view:\n%s", view)
		}

		// A typo is kept with a hint, then fixed with backspace
		guest = typeText(guest, sentence[:len(sentence)-1]+"#")
		guest, _ = pressKey(guest, tea.KeyEnter)
		if guest.online.battle.Hint == "" || guest.online.battle.Input == "" {
			t.Errorf("typo should keep the input and hint: %+v", guest.online.battle.Hint)
		}
		guest, _ = pressKey(guest, tea.KeyBackspace)
		guest = typeText(guest, sentence[len(sentence)-1:])
		guest, _ = pressKey(guest, tea.KeyEnter)

		runUntil(t, &host, &guest, "sentence claimed", func() bool {
			return host.online.battle.Phase != online.BattleFalling || host.online.battle.WordIndex != i
		})
	}

	runUntil(t, &host, &guest, "result", func() bool { return guest.online.battle.Phase == online.BattleResult })
	if top := host.online.battle.Ranking[0]; top.Name != "Guesty" || top.Words != 3 {
		t.Errorf("top = %+v", top)
	}
	if !strings.Contains(host.View(), "Lines") {
		t.Errorf("sentence results count lines:\n%s", host.View())
	}
}

func TestLiveTypingUI(t *testing.T) {
	host, guest := newRoomPair(t)
	host = startFromLobby(t, host)
	runUntil(t, &host, &guest, "first word", func() bool {
		return host.online.battle.Phase == online.BattleFalling && guest.online.battle.Phase == online.BattleFalling
	})

	// Each side sees the other's input appear in its live row as they type
	word := guest.online.battle.Word
	guest = typeText(guest, word[:2])
	host = typeText(host, "zz")
	runUntil(t, &host, &guest, "inputs relayed", func() bool {
		return host.online.battle.Typing[guest.online.battle.MyID] == word[:2] &&
			guest.online.battle.Typing[host.online.battle.MyID] == "zz"
	})
	if line := liveRowFor(host.View(), "Guesty"); !strings.Contains(line, word[:2]) {
		t.Errorf("host's live row for Guesty = %q", line)
	}
	if line := liveRowFor(guest.View(), "Hosty"); !strings.Contains(line, "zz") {
		t.Errorf("guest's live row for Hosty = %q", line)
	}

	// Clearing the input is relayed too
	host, _ = pressKey(host, tea.KeyBackspace)
	host, _ = pressKey(host, tea.KeyBackspace)
	runUntil(t, &host, &guest, "cleared input relayed", func() bool {
		_, ok := guest.online.battle.Typing[host.online.battle.MyID]
		return ok && guest.online.battle.Typing[host.online.battle.MyID] == ""
	})
}

// liveRowFor finds a player's live input row, below the word separator, so
// the score bar at the top is skipped
func liveRowFor(view, who string) string {
	below := false
	for _, l := range strings.Split(view, "\n") {
		if strings.Contains(l, "━━") {
			below = true
			continue
		}
		if below && strings.Contains(l, who) {
			return l
		}
	}
	return ""
}

func TestWordPoolUI(t *testing.T) {
	host, guest := newRoomPair(t)
	setLobbyOption(t, &host, &guest, 0, "Word Pool", func() bool {
		return guest.online.battle.Mode == online.ModeWordPool
	})
	setLobbyOption(t, &host, &guest, 2, "3", func() bool { return guest.online.battle.Settings.SpawnRate == 3 })
	setLobbyOption(t, &host, &guest, 3, "On", func() bool { return guest.online.battle.Settings.AllowSteal })
	host = startFromLobby(t, host)

	var word string
	runUntil(t, &host, &guest, "a word on the board", func() bool {
		for _, c := range guest.online.battle.Pool {
			if c.Word != "" {
				word = c.Word
				return true
			}
		}
		return false
	})
	if view := guest.View(); !strings.Contains(view, word) || !strings.Contains(view, "Steal ON") {
		t.Errorf("pool view:\n%s", view)
	}

	owner := func(m model) int {
		for _, c := range m.online.battle.Pool {
			if c.Word == word {
				return c.Owner
			}
		}
		return -1
	}

	guest = typeText(guest, word)
	guest, _ = pressKey(guest, tea.KeyEnter)
	runUntil(t, &host, &guest, "guest owns the word", func() bool { return owner(host) == guest.online.battle.MyID })

	// The guest's own word cannot be claimed again; the host steals it
	guest = typeText(guest, word)
	if _, ok := guest.online.battle.Submit(); ok {
		t.Error("an owned word must not be re-claimed")
	}
	host = typeText(host, word)
	host, _ = pressKey(host, tea.KeyEnter)
	runUntil(t, &host, &guest, "host steals the word", func() bool { return owner(guest) == host.online.battle.MyID })
	if guest.online.battle.Scores[guest.online.battle.MyID] != 0 || guest.online.battle.Scores[1] != len(word) {
		t.Errorf("scores after steal = %v", guest.online.battle.Scores)
	}
}
