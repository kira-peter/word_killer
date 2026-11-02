package main

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/word-killer/word-killer/pkg/config"
	"github.com/word-killer/word-killer/pkg/game"
	"github.com/word-killer/word-killer/pkg/online"
)

const testOnlinePort = 17788

func newTestModel(t *testing.T, name string) model {
	t.Helper()
	g := game.New()
	if err := g.LoadWordDictionaries(
		"../../data/google-10000-short.txt",
		"../../data/google-10000-medium.txt",
		"../../data/google-10000-long.txt",
		0.3, 0.5, 0.2,
	); err != nil {
		t.Fatalf("LoadWordDictionaries: %v", err)
	}
	cfg := config.DefaultConfig()
	cfg.OnlinePort = testOnlinePort
	cfg.OnlineDropDuration = 1
	cfg.OnlinePlayerName = name
	m := initialModel(cfg, g)
	t.Cleanup(m.online.closeConnections)
	return m
}

func pressKey(m model, key tea.KeyType) (model, tea.Cmd) {
	next, cmd := m.Update(tea.KeyMsg{Type: key})
	return next.(model), cmd
}

func typeText(m model, text string) model {
	for _, r := range text {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(model)
	}
	return m
}

func tick(m model) model {
	next, _ := m.Update(tickMsg(time.Now()))
	return next.(model)
}

// runUntil ticks both models until the condition holds
func runUntil(t *testing.T, a, b *model, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		*a = tick(*a)
		*b = tick(*b)
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

// openOnlineMenu navigates from the welcome screen to the online menu
func openOnlineMenu(t *testing.T, m model) model {
	t.Helper()
	m, _ = pressKey(m, tea.KeyDown)
	m, _ = pressKey(m, tea.KeyEnter)
	if m.online.screen != onlineMenu {
		t.Fatalf("screen = %v, want online menu", m.online.screen)
	}
	return m
}

func joinRoom(t *testing.T, m model, room string) model {
	t.Helper()
	m, _ = pressKey(m, tea.KeyDown)
	m, _ = pressKey(m, tea.KeyEnter)
	if m.online.screen != onlineJoinForm {
		t.Fatalf("screen = %v, want join form", m.online.screen)
	}
	m = typeText(m, "127.0.0.1")
	m, _ = pressKey(m, tea.KeyTab)
	m = typeText(m, room)

	m, cmd := pressKey(m, tea.KeyEnter)
	if cmd == nil {
		t.Fatalf("expected a join command, notice = %q", m.online.notice)
	}
	next, _ := m.Update(cmd())
	return next.(model)
}

// TestOnlineFlow plays a whole game through the real key handling and rendering
func TestOnlineFlow(t *testing.T) {
	host := newTestModel(t, "Hosty")
	guest := newTestModel(t, "Guesty")

	host = openOnlineMenu(t, host)
	host, _ = pressKey(host, tea.KeyEnter)
	if host.online.screen != onlineRoom || !host.online.battle.IsHost() {
		t.Fatalf("host not in room: screen=%v notice=%q", host.online.screen, host.online.notice)
	}
	room := host.online.battle.Room

	// Starting alone is refused
	host = tick(host)
	host, _ = pressKey(host, tea.KeyEnter)
	if host.online.notice != "Need at least 2 players" {
		t.Errorf("notice = %q", host.online.notice)
	}

	guest = openOnlineMenu(t, guest)
	guest = joinRoom(t, guest, room)
	if guest.online.screen != onlineRoom {
		t.Fatalf("guest not in room: %q", guest.online.notice)
	}

	runUntil(t, &host, &guest, "lobby with two players", func() bool {
		return len(host.online.battle.Players) == 2 && len(guest.online.battle.Players) == 2
	})
	for _, want := range []string{room, "Hosty", "Guesty", "Word Drop"} {
		if view := host.View(); !strings.Contains(view, want) {
			t.Errorf("lobby view missing %q:\n%s", want, view)
		}
	}
	if view := guest.View(); !strings.Contains(view, "Waiting for host") {
		t.Errorf("guest lobby view:\n%s", view)
	}

	// A guest cannot start the game
	guest, _ = pressKey(guest, tea.KeyEnter)

	host, _ = pressKey(host, tea.KeyEnter)
	runUntil(t, &host, &guest, "first word", func() bool {
		return guest.online.battle.Phase == online.BattleFalling
	})

	expected := 0
	for i := 0; i < onlineWordCount; i++ {
		runUntil(t, &host, &guest, "word to fall", func() bool {
			b := guest.online.battle
			return b.Phase == online.BattleFalling && b.WordIndex == i &&
				host.online.battle.Phase == online.BattleFalling && host.online.battle.WordIndex == i
		})

		word := guest.online.battle.Word
		if view := guest.View(); !strings.Contains(view, word) {
			t.Errorf("battle view missing word %q:\n%s", word, view)
		}

		// A wrong entry only clears the input
		guest = typeText(guest, "zzz")
		guest, _ = pressKey(guest, tea.KeyEnter)
		if guest.online.battle.Input != "" {
			t.Errorf("input not cleared: %q", guest.online.battle.Input)
		}

		guest = typeText(guest, word)
		guest, _ = pressKey(guest, tea.KeyEnter)
		expected += len(word)

		runUntil(t, &host, &guest, "claim result", func() bool {
			return host.online.battle.Phase != online.BattleFalling || host.online.battle.WordIndex != i
		})
		if i < 2 {
			if view := host.View(); !strings.Contains(view, "Guesty +") {
				t.Errorf("feedback missing:\n%s", view)
			}
		}
	}

	runUntil(t, &host, &guest, "result screen", func() bool {
		return host.online.battle.Phase == online.BattleResult && guest.online.battle.Phase == online.BattleResult
	})
	top := host.online.battle.Ranking[0]
	if top.Name != "Guesty" || top.Score != expected || top.Words != onlineWordCount {
		t.Errorf("top rank = %+v, want Guesty with %d", top, expected)
	}
	if view := guest.View(); !strings.Contains(view, "You win!") {
		t.Errorf("guest result view:\n%s", view)
	}

	host, _ = pressKey(host, tea.KeyEnter)
	runUntil(t, &host, &guest, "back to lobby", func() bool {
		return guest.online.battle.Phase == online.BattleLobby
	})

	// Closing the room sends the guest back to the menu with a notice
	host, _ = pressKey(host, tea.KeyEsc)
	if host.online.screen != onlineMenu || host.online.server != nil {
		t.Errorf("host after close: screen=%v", host.online.screen)
	}
	runUntil(t, &host, &guest, "guest notified", func() bool {
		return guest.online.screen == onlineMenu
	})
	if guest.online.notice != "Host closed the room" {
		t.Errorf("guest notice = %q", guest.online.notice)
	}
}

func TestWelcomeMenuOptions(t *testing.T) {
	m := newTestModel(t, "Solo")
	if view := m.View(); !strings.Contains(view, "Online") || !strings.Contains(view, "About") {
		t.Errorf("welcome view:\n%s", view)
	}

	// Start, Online, About, then wrap around to Start
	m, _ = pressKey(m, tea.KeyDown)
	m, _ = pressKey(m, tea.KeyDown)
	m, _ = pressKey(m, tea.KeyEnter)
	if !m.showAbout {
		t.Error("third option should open About")
	}
	m, _ = pressKey(m, tea.KeyEsc)
	m, _ = pressKey(m, tea.KeyDown)
	if m.welcomeAnimState.SelectedOption != 0 {
		t.Errorf("selection = %d, want wrap to 0", m.welcomeAnimState.SelectedOption)
	}
}

func TestJoinFormValidation(t *testing.T) {
	m := openOnlineMenu(t, newTestModel(t, "Solo"))
	m, _ = pressKey(m, tea.KeyDown)
	m, _ = pressKey(m, tea.KeyEnter)

	m, cmd := pressKey(m, tea.KeyEnter)
	if cmd != nil || m.online.notice == "" {
		t.Errorf("empty host should be rejected, notice = %q", m.online.notice)
	}

	m = typeText(m, "127.0.0.1")
	m, _ = pressKey(m, tea.KeyTab)
	m = typeText(m, "12ab34567")
	if m.online.room != "1234" {
		t.Errorf("room = %q, want digits only, at most 4", m.online.room)
	}

	// Nothing listens on the test port, so the form stays with an error
	m, cmd = pressKey(m, tea.KeyEnter)
	if cmd == nil {
		t.Fatal("expected a join command")
	}
	next, _ := m.Update(cmd())
	m = next.(model)
	if m.online.screen != onlineJoinForm || m.online.notice != "Cannot connect to host" {
		t.Errorf("screen=%v notice=%q", m.online.screen, m.online.notice)
	}
}

func TestPortInUseShowsError(t *testing.T) {
	first := openOnlineMenu(t, newTestModel(t, "First"))
	first, _ = pressKey(first, tea.KeyEnter)
	if first.online.screen != onlineRoom {
		t.Fatalf("first room failed: %q", first.online.notice)
	}

	second := openOnlineMenu(t, newTestModel(t, "Second"))
	second, _ = pressKey(second, tea.KeyEnter)
	if second.online.screen != onlineMenu || !strings.Contains(second.online.notice, "17788") {
		t.Errorf("screen=%v notice=%q", second.online.screen, second.online.notice)
	}
}

func TestBatchedKeysAreNotDropped(t *testing.T) {
	m := openOnlineMenu(t, newTestModel(t, "Solo"))
	m, _ = pressKey(m, tea.KeyDown)
	m, _ = pressKey(m, tea.KeyEnter)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("192.168.1.20")})
	m = next.(model)
	if m.online.host != "192.168.1.20" {
		t.Errorf("host = %q", m.online.host)
	}

	m.online.screen = onlineRoom
	m.online.battle = online.NewBattle(2, "1234")
	m.online.battle.Apply(online.Message{Type: online.TypeWordSpawn, Word: "hello", FallMs: 1000}, time.Now())
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("HeL1")})
	m = next.(model)
	if m.online.battle.Input != "hel" {
		t.Errorf("input = %q, want %q", m.online.battle.Input, "hel")
	}
}

func TestRenameInLobbyUI(t *testing.T) {
	host := openOnlineMenu(t, newTestModel(t, "Hosty"))
	host, _ = pressKey(host, tea.KeyEnter)
	guest := joinRoom(t, openOnlineMenu(t, newTestModel(t, "Guesty")), host.online.battle.Room)
	runUntil(t, &host, &guest, "both in lobby", func() bool {
		return len(guest.online.battle.Players) == 2 && len(host.online.battle.Players) == 2
	})

	// Letters like j and k are part of the name, and Esc only cancels the edit
	guest, _ = pressKey(guest, tea.KeyTab)
	guest = typeText(guest, "jk")
	if !strings.Contains(guest.View(), "Guestyjk_") {
		t.Errorf("editing view:\n%s", guest.View())
	}
	guest, _ = pressKey(guest, tea.KeyEsc)
	if guest.online.screen != onlineRoom || guest.online.editingName {
		t.Fatalf("esc should only cancel: screen=%v", guest.online.screen)
	}

	// An empty name keeps the old one
	guest, _ = pressKey(guest, tea.KeyTab)
	for range "Guesty" {
		guest, _ = pressKey(guest, tea.KeyBackspace)
	}
	guest, _ = pressKey(guest, tea.KeyEnter)
	if guest.cfg.OnlinePlayerName != "Guesty" {
		t.Errorf("empty name changed config: %q", guest.cfg.OnlinePlayerName)
	}

	// Save a new name with a space in it
	guest, _ = pressKey(guest, tea.KeyTab)
	for range "Guesty" {
		guest, _ = pressKey(guest, tea.KeyBackspace)
	}
	guest = typeText(guest, "New Name")
	guest, _ = pressKey(guest, tea.KeyEnter)

	runUntil(t, &host, &guest, "rename broadcast", func() bool {
		return strings.Contains(host.View(), "New Name")
	})
	if strings.Contains(host.View(), "Guesty") {
		t.Errorf("old name still shown:\n%s", host.View())
	}

	// The host can rename too
	host, _ = pressKey(host, tea.KeyTab)
	host = typeText(host, "x")
	host, _ = pressKey(host, tea.KeyEnter)
	runUntil(t, &host, &guest, "host rename broadcast", func() bool {
		return guest.online.battle.PlayerName(1) == "Hostyx"
	})

	// An unfinished edit is dropped when a game starts
	guest, _ = pressKey(guest, tea.KeyTab)
	guest = typeText(guest, "zzz")
	host, _ = pressKey(host, tea.KeyEnter)
	runUntil(t, &host, &guest, "game start", func() bool {
		return guest.online.battle.Phase != online.BattleLobby
	})
	if guest.online.editingName {
		t.Error("editing should end when the game starts")
	}
}
