package main

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/word-killer/word-killer/pkg/config"
	"github.com/word-killer/word-killer/pkg/online"
	"github.com/word-killer/word-killer/pkg/ui"
)

const (
	onlineStartCountdown = 3 * time.Second
	onlineNextWordDelay  = 1 * time.Second
	onlineLoopbackHost   = "127.0.0.1"
)

// onlineScreen identifies which online screen is showing
type onlineScreen int

const (
	onlineNone onlineScreen = iota
	onlineMenu              // Create Room plus the rooms found on the LAN
	onlineRoom              // lobby, battle or result, depending on battle.Phase
)

// onlineState holds everything the online screens need
type onlineState struct {
	screen     onlineScreen
	notice     string    // shown on the current screen until the next action
	noticeAt   time.Time // when the menu notice appeared; it fades after menuNoticeTTL
	connecting bool
	address    string // host:port shown in the lobby

	// The menu lists Create Room, then rooms. The selection is tracked by room
	// ID so it stays on the same room as the list refreshes; "" is Create Room.
	browser      *online.Browser
	rooms        []online.RoomInfo
	selectedRoom string
	joiningRoom  online.RoomInfo

	editingName bool // the local player is renaming in the lobby
	nameInput   string

	// Host-only lobby controls: the highlighted menu row, the setting wheel
	// being adjusted, and the kick picker
	lobbyIndex int
	wheel      *lobbyWheel
	kickIDs    []int // nil while the picker is closed
	kickIndex  int

	server *online.Server // only set on the host
	client *online.Client
	battle *online.Battle
}

// lobbyWheel is a setting being adjusted. The value only changes locally
// until the host confirms; Esc throws it away.
type lobbyWheel struct {
	key    string   // online.OptionMode or a setting key
	label  string   //
	values []int    // setting steps; for the mode wheel, indexes into online.Modes
	steps  []string // display text for each value
	index  int
}

// joinResultMsg reports the outcome of an asynchronous join
type joinResultMsg struct {
	client *online.Client
	err    error
}

// menuNoticeTTL is how long a notice stays on the online menu
const menuNoticeTTL = 3 * time.Second

// setNotice shows a notice and starts its timer
func (o *onlineState) setNotice(notice string) {
	o.notice = notice
	o.noticeAt = time.Now()
}

// openMenu shows the online menu and starts looking for rooms
func (o *onlineState) openMenu(port int, notice string) {
	o.screen = onlineMenu
	o.setNotice(notice)
	o.selectedRoom = ""
	o.rooms = nil
	if o.browser == nil {
		browser, err := online.NewBrowser(port)
		if err != nil {
			o.setNotice("Cannot search for rooms")
			return
		}
		o.browser = browser
	}
}

// stopBrowsing stops looking for rooms
func (o *onlineState) stopBrowsing() {
	if o.browser != nil {
		o.browser.Close()
		o.browser = nil
	}
	o.rooms = nil
}

// closeMenu leaves the online screens for the welcome screen
func (o *onlineState) closeMenu() {
	o.stopBrowsing()
	o.notice = ""
	o.screen = onlineNone
}

// enterRoom switches to the room screen using an established connection
func (o *onlineState) enterRoom(client *online.Client, address string) {
	o.stopBrowsing()
	o.client = client
	o.battle = online.NewBattle(client.PlayerID)
	o.address = address
	o.screen = onlineRoom
	o.notice = ""
	o.resetLobbyControls()
}

// resetLobbyControls closes any open lobby overlay and returns to the top row
func (o *onlineState) resetLobbyControls() {
	o.editingName = false
	o.lobbyIndex = 0
	o.wheel = nil
	o.kickIDs = nil
	o.kickIndex = 0
}

// leaveRoom releases the connection and server, then returns to the online menu
func (m model) leaveRoom(notice string) {
	o := m.online
	o.closeConnections()
	o.battle = nil
	o.resetLobbyControls()
	o.openMenu(m.cfg.OnlinePort, notice)
}

// selectedIndex returns the menu position of the selection: 0 is Create Room
func (o *onlineState) selectedIndex() int {
	for i, r := range o.rooms {
		if r.ID == o.selectedRoom {
			return i + 1
		}
	}
	return 0
}

// refreshRooms takes a new room snapshot. A selected room that disappeared
// falls back to Create Room.
func (o *onlineState) refreshRooms() {
	if o.browser == nil {
		return
	}
	o.rooms = o.browser.Rooms()
	if o.selectedRoom != "" && o.selectedIndex() == 0 {
		o.selectedRoom = ""
	}
}

func (o *onlineState) closeConnections() {
	o.stopBrowsing()
	// Close the server first so other players are told the room is gone
	if o.server != nil {
		o.server.Close()
		o.server = nil
	}
	if o.client != nil {
		o.client.Close()
		o.client = nil
	}
}

// createRoom starts a server on this machine and joins it as the host
func (m model) createRoom() {
	server, err := online.NewServer(online.ServerConfig{
		Port:           m.cfg.OnlinePort,
		DropDuration:   time.Duration(m.cfg.OnlineDropDuration) * time.Second,
		StartCountdown: onlineStartCountdown,
		NextWordDelay:  onlineNextWordDelay,
		WordSource:     m.game.DrawWords,
		SentenceSource: m.game.DrawSentences,
	})
	if err != nil {
		m.online.setNotice("Cannot open room on port " + strconv.Itoa(m.cfg.OnlinePort))
		return
	}

	loopback := net.JoinHostPort(onlineLoopbackHost, strconv.Itoa(server.Port()))
	client, err := online.Connect(loopback, m.cfg.OnlinePlayerName)
	if err != nil {
		server.Close()
		m.online.setNotice("Cannot open room on port " + strconv.Itoa(m.cfg.OnlinePort))
		return
	}

	m.online.server = server
	m.online.enterRoom(client, net.JoinHostPort(online.LocalIP(), strconv.Itoa(server.Port())))
}

// joinCmd connects in the background so the UI keeps rendering meanwhile
func joinCmd(address, name string) tea.Cmd {
	return func() tea.Msg {
		client, err := online.Connect(address, name)
		return joinResultMsg{client: client, err: err}
	}
}

func (m model) handleJoinResult(msg joinResultMsg) {
	m.online.connecting = false
	// The player may have left the menu while connecting
	if m.online.screen != onlineMenu {
		if msg.client != nil {
			msg.client.Close()
		}
		return
	}
	if msg.err != nil {
		var rejected *online.RejectedError
		if errors.As(msg.err, &rejected) {
			m.online.setNotice(rejected.Error())
		} else {
			m.online.setNotice("Cannot connect to host")
		}
		return
	}
	m.online.enterRoom(msg.client, m.online.joiningRoom.Address())
}

// pollOnline applies pending server messages; called on every tick
func (m model) pollOnline() {
	o := m.online
	if o.screen == onlineMenu {
		o.refreshRooms()
		if o.notice != "" && time.Since(o.noticeAt) >= menuNoticeTTL {
			o.notice = ""
		}
	}
	if o.client == nil {
		return
	}
	// Share the local input with the other players, at most once per tick
	if msg, ok := o.battle.TypingUpdate(); ok {
		o.client.Send(msg)
	}
	now := time.Now()
	for _, msg := range o.client.Poll() {
		switch msg.Type {
		case online.TypeHostClosed:
			m.leaveRoom("Host closed the room")
			return
		case online.TypeDisconnected:
			m.leaveRoom("Connection lost")
			return
		case online.TypeKicked:
			m.leaveRoom("You were kicked")
			return
		case online.TypeStartFailed:
			o.notice = msg.Reason
		default:
			o.battle.Apply(msg, now)
			// The host may start the game while someone is still renaming
			if o.battle.Phase != online.BattleLobby {
				o.resetLobbyControls()
			}
			o.keepLobbyControlsValid()
		}
	}
}

func (m model) handleOnlineKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		m.online.closeConnections()
		return m, tea.Quit
	}

	switch m.online.screen {
	case onlineMenu:
		return m.handleOnlineMenuKey(msg)
	case onlineRoom:
		return m.handleRoomKey(msg)
	}
	return m, nil
}

func (m model) handleOnlineMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	o := m.online
	// A join is in flight; its result arrives as joinResultMsg
	if o.connecting {
		return m, nil
	}

	// Positions: 0 is Create Room, 1..len(rooms) are the rooms
	count := len(o.rooms) + 1
	switch msg.String() {
	case "up", "k":
		o.selectRoomAt((o.selectedIndex() - 1 + count) % count)
	case "down", "j":
		o.selectRoomAt((o.selectedIndex() + 1) % count)
	case "enter":
		o.notice = ""
		index := o.selectedIndex()
		if index == 0 {
			m.createRoom()
			return m, nil
		}
		room := o.rooms[index-1]
		switch {
		case room.InGame:
			o.setNotice("Game in progress")
		case room.Full():
			o.setNotice("Room is full")
		default:
			o.connecting = true
			o.joiningRoom = room
			return m, joinCmd(room.Address(), m.cfg.OnlinePlayerName)
		}
	case "esc":
		o.closeMenu()
	}
	return m, nil
}

// selectRoomAt moves the selection to a menu position: 0 is Create Room
func (o *onlineState) selectRoomAt(index int) {
	if index == 0 {
		o.selectedRoom = ""
	} else {
		o.selectedRoom = o.rooms[index-1].ID
	}
}

func (m model) handleRoomKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	o := m.online
	key := msg.String()

	if o.editingName {
		return m.handleNameEditKey(msg)
	}

	// Esc first closes an open wheel or kick picker; only then does it leave
	if key == "esc" && o.wheel == nil && o.kickIDs == nil {
		m.leaveRoom("")
		return m, nil
	}

	switch o.battle.Phase {
	case online.BattleLobby:
		if key == "tab" {
			o.editingName = true
			o.nameInput = o.battle.PlayerName(o.battle.MyID)
			return m, nil
		}
		if o.battle.IsHost() {
			m.handleHostLobbyKey(key)
		}

	case online.BattleFalling:
		switch key {
		case "enter":
			if claim, ok := o.battle.Submit(); ok {
				o.client.Send(claim)
			}
		case "backspace":
			o.battle.Backspace()
		default:
			// msg.Runes may hold several characters when typing fast; a space
			// arrives as KeySpace with Runes holding ' '
			for _, r := range msg.Runes {
				o.battle.AddChar(r)
			}
		}

	case online.BattleResult:
		if key == "enter" && o.battle.IsHost() {
			o.client.Send(online.Message{Type: online.TypeBackToLobby})
		}
	}
	return m, nil
}

// handleHostLobbyKey drives the host's lobby menu, setting wheel and kick picker
func (m model) handleHostLobbyKey(key string) {
	o := m.online
	switch {
	case o.wheel != nil:
		m.handleWheelKey(key)
	case o.kickIDs != nil:
		m.handleKickKey(key)
	default:
		m.handleLobbyMenuKey(key)
	}
}

func (m model) handleLobbyMenuKey(key string) {
	o := m.online
	items := len(ui.LobbyItems(o.battle))
	switch key {
	case "up", "k":
		if o.lobbyIndex > 0 {
			o.lobbyIndex--
		}
	case "down", "j":
		if o.lobbyIndex < items-1 {
			o.lobbyIndex++
		}
	case "K", "x":
		o.openKickPicker()
	case "enter":
		o.notice = ""
		if o.lobbyIndex == items-1 {
			if len(o.battle.Players) < online.MinPlayers {
				o.notice = "Need at least 2 players"
				return
			}
			o.client.Send(online.Message{Type: online.TypeStart})
			return
		}
		o.openWheel(o.lobbyIndex)
	}
}

// openWheel starts adjusting menu row i: row 0 is the mode, then settings
func (o *onlineState) openWheel(i int) {
	if i == 0 {
		w := &lobbyWheel{key: online.OptionMode, label: "Mode"}
		for idx, mode := range online.Modes {
			w.values = append(w.values, idx)
			w.steps = append(w.steps, online.ModeName(mode))
			if mode == o.battle.Mode {
				w.index = idx
			}
		}
		o.wheel = w
		return
	}

	options := online.ModeOptions(o.battle.Mode)
	if i-1 >= len(options) {
		return
	}
	opt := options[i-1]
	w := &lobbyWheel{key: opt.Key, label: opt.Label, values: opt.Values}
	for _, v := range opt.Values {
		w.steps = append(w.steps, opt.Format(v))
	}
	if idx := opt.Index(o.battle.Settings.Get(opt.Key)); idx >= 0 {
		w.index = idx
	}
	o.wheel = w
}

func (m model) handleWheelKey(key string) {
	o := m.online
	w := o.wheel
	switch key {
	case "up", "k":
		if w.index > 0 {
			w.index--
		}
	case "down", "j":
		if w.index < len(w.values)-1 {
			w.index++
		}
	case "enter":
		msg := online.Message{Type: online.TypeSetOption, Option: w.key, Value: w.values[w.index]}
		if w.key == online.OptionMode {
			msg.Mode = online.Modes[w.values[w.index]]
		}
		o.client.Send(msg)
		o.wheel = nil
	case "esc":
		o.wheel = nil
	}
}

// openKickPicker offers every other player for kicking
func (o *onlineState) openKickPicker() {
	var ids []int
	for _, p := range o.battle.Players {
		if p.ID != o.battle.MyID {
			ids = append(ids, p.ID)
		}
	}
	if len(ids) == 0 {
		o.notice = "No one to kick"
		return
	}
	o.notice = ""
	o.kickIDs = ids
	o.kickIndex = 0
}

func (m model) handleKickKey(key string) {
	o := m.online
	switch key {
	case "up", "k":
		if o.kickIndex > 0 {
			o.kickIndex--
		}
	case "down", "j":
		if o.kickIndex < len(o.kickIDs)-1 {
			o.kickIndex++
		}
	case "enter":
		o.client.Send(online.Message{Type: online.TypeKick, PlayerID: o.kickIDs[o.kickIndex]})
		o.kickIDs = nil
	case "esc":
		o.kickIDs = nil
	}
}

// keepLobbyControlsValid fixes the lobby controls after a lobby update: the
// menu may have shrunk with a new mode, and a player offered for kicking may
// have left
func (o *onlineState) keepLobbyControlsValid() {
	if o.battle.Phase != online.BattleLobby {
		return
	}
	if last := len(ui.LobbyItems(o.battle)) - 1; o.lobbyIndex > last {
		o.lobbyIndex = last
	}
	if o.kickIDs == nil {
		return
	}
	present := make(map[int]bool)
	for _, p := range o.battle.Players {
		present[p.ID] = true
	}
	ids := o.kickIDs[:0]
	for _, id := range o.kickIDs {
		if present[id] {
			ids = append(ids, id)
		}
	}
	switch {
	case len(ids) == 0:
		o.kickIDs = nil
	default:
		o.kickIDs = ids
		o.kickIndex = min(o.kickIndex, len(ids)-1)
	}
}

// handleNameEditKey edits the local player's name in the lobby
func (m model) handleNameEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	o := m.online
	switch msg.String() {
	case "esc":
		o.editingName = false
	case "enter":
		o.editingName = false
		// An empty name keeps the current one
		if name := strings.TrimSpace(o.nameInput); name != "" {
			m.cfg.OnlinePlayerName = name // reuse it for later rooms this session
			o.client.Send(online.Message{Type: online.TypeRename, Name: name})
		}
	case "backspace":
		if runes := []rune(o.nameInput); len(runes) > 0 {
			o.nameInput = string(runes[:len(runes)-1])
		}
	default:
		for _, r := range msg.Runes {
			if unicode.IsPrint(r) && len([]rune(o.nameInput)) < config.MaxPlayerNameLength {
				o.nameInput += string(r)
			}
		}
	}
	return m, nil
}

func (m model) onlineView() string {
	o := m.online
	switch o.screen {
	case onlineMenu:
		return ui.RenderOnlineMenu(o.rooms, o.selectedIndex(), o.notice, o.connecting, m.animFrame)
	case onlineRoom:
		switch o.battle.Phase {
		case online.BattleLobby:
			view := ui.LobbyView{
				Selected:    o.lobbyIndex,
				Notice:      o.notice,
				EditingName: o.editingName,
				NameInput:   o.nameInput,
				KickPicker:  o.kickIDs,
				KickIndex:   o.kickIndex,
				Address:     o.address,
			}
			if w := o.wheel; w != nil {
				view.Wheel = &ui.Wheel{Label: w.label, Steps: w.steps, Index: w.index}
			}
			return ui.RenderLobby(o.battle, view)
		case online.BattleResult:
			return ui.RenderBattleResult(o.battle)
		default:
			return ui.RenderBattle(o.battle, time.Now())
		}
	}
	return ""
}
