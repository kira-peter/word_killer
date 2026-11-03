package main

import (
	"strconv"
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
	if err := g.LoadSentences("../../data/sentences.txt"); err != nil {
		t.Fatalf("LoadSentences: %v", err)
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

// waitForListedRoom ticks the model until a room hosted by hostName is listed
func waitForListedRoom(t *testing.T, m model, hostName string, cond func(online.RoomInfo) bool) (model, online.RoomInfo) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		m = tick(m)
		for _, r := range m.online.rooms {
			if r.HostName == hostName && cond(r) {
				return m, r
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("room of %s never listed, rooms = %+v", hostName, m.online.rooms)
	return m, online.RoomInfo{}
}

// selectListedRoom moves the selection down until it reaches the room
func selectListedRoom(t *testing.T, m model, roomID string) model {
	t.Helper()
	for i := 0; i <= len(m.online.rooms); i++ {
		if m.online.selectedRoom == roomID {
			return m
		}
		m, _ = pressKey(m, tea.KeyDown)
	}
	t.Fatalf("could not select room %s", roomID)
	return m
}

// joinRoom joins the room hosted by hostName from the online menu's list
func joinRoom(t *testing.T, m model, hostName string) model {
	t.Helper()
	m, room := waitForListedRoom(t, m, hostName, online.RoomInfo.Joinable)
	m = selectListedRoom(t, m, room.ID)

	m, cmd := pressKey(m, tea.KeyEnter)
	if cmd == nil || !m.online.connecting {
		t.Fatalf("expected a join command, notice = %q", m.online.notice)
	}
	if !strings.Contains(m.View(), "Connecting...") {
		t.Errorf("connecting view:\n%s", m.View())
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
	// Starting alone is refused
	host = tick(host)
	host = startFromLobby(t, host)
	if host.online.notice != "Need at least 2 players" {
		t.Errorf("notice = %q", host.online.notice)
	}

	guest = openOnlineMenu(t, guest)
	guest, _ = waitForListedRoom(t, guest, "Hosty", online.RoomInfo.Joinable)
	if view := guest.View(); !strings.Contains(view, "Hosty") || !strings.Contains(view, "Waiting") {
		t.Errorf("menu view missing the room:\n%s", view)
	}
	guest = joinRoom(t, guest, "Hosty")
	if guest.online.screen != onlineRoom {
		t.Fatalf("guest not in room: %q", guest.online.notice)
	}
	if guest.online.browser != nil {
		t.Error("browsing should stop after joining")
	}

	runUntil(t, &host, &guest, "lobby with two players", func() bool {
		return len(host.online.battle.Players) == 2 && len(guest.online.battle.Players) == 2
	})
	for _, want := range []string{strconv.Itoa(testOnlinePort), "Hosty", "Guesty", "Word Drop"} {
		if view := host.View(); !strings.Contains(view, want) {
			t.Errorf("lobby view missing %q:\n%s", want, view)
		}
	}
	if view := guest.View(); !strings.Contains(view, "Waiting for host") {
		t.Errorf("guest lobby view:\n%s", view)
	}

	// A guest cannot start the game
	guest, _ = pressKey(guest, tea.KeyEnter)

	host = startFromLobby(t, host)
	runUntil(t, &host, &guest, "first word", func() bool {
		return guest.online.battle.Phase == online.BattleFalling
	})

	wordCount := online.DefaultSettings(online.ModeWordDrop).Count
	expected := 0
	for i := 0; i < wordCount; i++ {
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
	if top.Name != "Guesty" || top.Score != expected || top.Words != wordCount {
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

func TestOnlineMenuListsRooms(t *testing.T) {
	m := openOnlineMenu(t, newTestModel(t, "Solo"))
	if view := m.View(); !strings.Contains(view, "Searching for rooms") || strings.Contains(view, "Join") {
		t.Errorf("empty menu view:\n%s", view)
	}

	// Rooms are fed in directly so the list can be checked without a network
	m.online.rooms = []online.RoomInfo{
		{ID: "a", HostName: "Alice", IP: "192.168.1.20", Port: 7788, Players: 2, MaxPlayers: 4},
		{ID: "b", HostName: "Bob", IP: "192.168.1.21", Port: 7788, Players: 4, MaxPlayers: 4},
		{ID: "c", HostName: "Carol", IP: "192.168.1.35", Port: 7788, Players: 3, MaxPlayers: 4, InGame: true},
	}
	m.online.browser.Close()
	m.online.browser = nil // keep the fake list from being refreshed away

	view := m.View()
	for _, want := range []string{"Alice", "2/4", "Waiting", "192.168.1.20", "Full", "In game"} {
		if !strings.Contains(view, want) {
			t.Errorf("menu view missing %q:\n%s", want, view)
		}
	}

	// Full and in-game rooms only explain why they cannot be joined
	m, _ = pressKey(m, tea.KeyDown)
	m, _ = pressKey(m, tea.KeyDown)
	m, cmd := pressKey(m, tea.KeyEnter)
	if cmd != nil || m.online.notice != "Room is full" {
		t.Errorf("full room: notice=%q", m.online.notice)
	}
	m, _ = pressKey(m, tea.KeyDown)
	m, cmd = pressKey(m, tea.KeyEnter)
	if cmd != nil || m.online.notice != "Game in progress" {
		t.Errorf("in-game room: notice=%q", m.online.notice)
	}

	// The selection wraps around to Create Room
	m, _ = pressKey(m, tea.KeyDown)
	if m.online.selectedIndex() != 0 {
		t.Errorf("selection = %d, want Create Room", m.online.selectedIndex())
	}

	// A joinable room that is no longer there reports the failure in the menu
	m, _ = pressKey(m, tea.KeyDown)
	m.online.rooms[0].Port = 9 // nothing listens on the discard port
	m, cmd = pressKey(m, tea.KeyEnter)
	if cmd == nil {
		t.Fatal("expected a join command")
	}
	m, _ = pressKey(m, tea.KeyUp) // ignored while connecting
	if m.online.selectedRoom != "a" {
		t.Errorf("keys should be ignored while connecting, selection = %q", m.online.selectedRoom)
	}
	next, _ := m.Update(cmd())
	m = next.(model)
	if m.online.screen != onlineMenu || m.online.notice != "Cannot connect to host" || m.online.connecting {
		t.Errorf("screen=%v notice=%q connecting=%v", m.online.screen, m.online.notice, m.online.connecting)
	}

	m, _ = pressKey(m, tea.KeyEsc)
	if m.online.screen != onlineNone {
		t.Errorf("esc should leave the online menu, screen=%v", m.online.screen)
	}
}

func TestMenuNoticeFades(t *testing.T) {
	m := openOnlineMenu(t, newTestModel(t, "Solo"))
	m.online.setNotice("Host closed the room")
	m = tick(m)
	if m.online.notice == "" || !strings.Contains(m.View(), "Host closed the room") {
		t.Fatal("a fresh notice should show")
	}

	// Pretend it appeared three seconds ago
	m.online.noticeAt = time.Now().Add(-menuNoticeTTL)
	m = tick(m)
	if m.online.notice != "" || strings.Contains(m.View(), "Host closed the room") {
		t.Errorf("notice should be gone after %v", menuNoticeTTL)
	}
}

func TestTooSmallWindow(t *testing.T) {
	m := newTestModel(t, "Solo")
	resize := func(m model, w, h int) model {
		next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
		return next.(model)
	}

	m = resize(m, 80, 24)
	if view := m.View(); !strings.Contains(view, "Window too small") || strings.Contains(view, "Start") {
		t.Errorf("a small window should only show the warning:\n%s", view)
	}

	// Keys still work underneath, and the game comes back on a bigger window
	m, _ = pressKey(m, tea.KeyDown)
	m = resize(m, 120, 40)
	if view := m.View(); strings.Contains(view, "Window too small") || !strings.Contains(view, "> Online <") {
		t.Errorf("an enlarged window should show the game again:\n%s", view)
	}
}

func TestSelectionFollowsRoom(t *testing.T) {
	o := &onlineState{screen: onlineMenu}
	o.rooms = []online.RoomInfo{{ID: "a"}, {ID: "b"}}
	o.selectRoomAt(2)

	// A new room sorted ahead of the selection must not move it
	o.rooms = []online.RoomInfo{{ID: "0"}, {ID: "a"}, {ID: "b"}}
	if o.selectedIndex() != 3 || o.selectedRoom != "b" {
		t.Errorf("selection moved: index=%d room=%q", o.selectedIndex(), o.selectedRoom)
	}

	// When the selected room disappears the selection returns to Create Room
	o.browser = nil
	o.rooms = []online.RoomInfo{{ID: "0"}, {ID: "a"}}
	if o.selectedIndex() != 0 {
		t.Errorf("index = %d, want Create Room", o.selectedIndex())
	}
}

func TestLeftMenuWhileConnecting(t *testing.T) {
	host := openOnlineMenu(t, newTestModel(t, "Hosty"))
	host, _ = pressKey(host, tea.KeyEnter)

	guest := openOnlineMenu(t, newTestModel(t, "Guesty"))
	guest, room := waitForListedRoom(t, guest, "Hosty", online.RoomInfo.Joinable)
	guest = selectListedRoom(t, guest, room.ID)
	guest, cmd := pressKey(guest, tea.KeyEnter)
	if cmd == nil {
		t.Fatal("expected a join command")
	}

	// Keys are ignored while connecting, so clear the flag to simulate the
	// player leaving before a slow join reply lands; that reply must be dropped
	guest.online.connecting = false
	guest, _ = pressKey(guest, tea.KeyEsc)
	next, _ := guest.Update(cmd())
	guest = next.(model)
	if guest.online.screen != onlineNone || guest.online.client != nil {
		t.Errorf("late join should be dropped: screen=%v", guest.online.screen)
	}
	runUntil(t, &host, &guest, "host sees the dropped guest leave", func() bool {
		return len(host.online.battle.Players) == 1
	})
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
	m := newTestModel(t, "Solo")
	m.online.screen = onlineRoom
	m.online.battle = online.NewBattle(2)
	m.online.battle.Apply(online.Message{Type: online.TypeWordSpawn, Word: "hello", FallMs: 1000}, time.Now())
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("HeL1")})
	m = next.(model)
	if m.online.battle.Input != "hel" {
		t.Errorf("input = %q, want %q", m.online.battle.Input, "hel")
	}
}

func TestRenameInLobbyUI(t *testing.T) {
	host := openOnlineMenu(t, newTestModel(t, "Hosty"))
	host, _ = pressKey(host, tea.KeyEnter)
	guest := joinRoom(t, openOnlineMenu(t, newTestModel(t, "Guesty")), "Hosty")
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
	host = startFromLobby(t, host)
	runUntil(t, &host, &guest, "game start", func() bool {
		return guest.online.battle.Phase != online.BattleLobby
	})
	if guest.online.editingName {
		t.Error("editing should end when the game starts")
	}
}
