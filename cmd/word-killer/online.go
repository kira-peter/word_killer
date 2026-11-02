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
	onlineWordCount      = 10
	onlineStartCountdown = 3 * time.Second
	onlineNextWordDelay  = 1 * time.Second
	onlineLoopbackHost   = "127.0.0.1"
)

// onlineScreen identifies which online screen is showing
type onlineScreen int

const (
	onlineNone onlineScreen = iota
	onlineMenu // Create Room plus the rooms found on the LAN
	onlineRoom // lobby, battle or result, depending on battle.Phase
)

// onlineState holds everything the online screens need
type onlineState struct {
	screen     onlineScreen
	notice     string // shown on the current screen until the next action
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

	server *online.Server // only set on the host
	client *online.Client
	battle *online.Battle
}

// joinResultMsg reports the outcome of an asynchronous join
type joinResultMsg struct {
	client *online.Client
	err    error
}

// openMenu shows the online menu and starts looking for rooms
func (o *onlineState) openMenu(port int, notice string) {
	o.screen = onlineMenu
	o.notice = notice
	o.selectedRoom = ""
	o.rooms = nil
	if o.browser == nil {
		browser, err := online.NewBrowser(port)
		if err != nil {
			o.notice = "Cannot search for rooms"
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
	o.editingName = false
}

// leaveRoom releases the connection and server, then returns to the online menu
func (m model) leaveRoom(notice string) {
	o := m.online
	o.closeConnections()
	o.battle = nil
	o.editingName = false
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
		WordCount:      onlineWordCount,
		StartCountdown: onlineStartCountdown,
		NextWordDelay:  onlineNextWordDelay,
		WordSource:     m.game.DrawWords,
	})
	if err != nil {
		m.online.notice = "Cannot open room on port " + strconv.Itoa(m.cfg.OnlinePort)
		return
	}

	loopback := net.JoinHostPort(onlineLoopbackHost, strconv.Itoa(server.Port()))
	client, err := online.Connect(loopback, m.cfg.OnlinePlayerName)
	if err != nil {
		server.Close()
		m.online.notice = "Cannot open room on port " + strconv.Itoa(m.cfg.OnlinePort)
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
			m.online.notice = rejected.Error()
		} else {
			m.online.notice = "Cannot connect to host"
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
	}
	if o.client == nil {
		return
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
		default:
			o.battle.Apply(msg, now)
			// The host may start the game while someone is still renaming
			if o.battle.Phase != online.BattleLobby {
				o.editingName = false
			}
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
			o.notice = "Game in progress"
		case room.Full():
			o.notice = "Room is full"
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

	if key == "esc" {
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
		if !o.battle.IsHost() {
			return m, nil
		}
		switch key {
		case "up", "k":
			o.client.Send(online.Message{Type: online.TypeSelectMode, Mode: stepMode(o.battle.Mode, -1)})
		case "down", "j":
			o.client.Send(online.Message{Type: online.TypeSelectMode, Mode: stepMode(o.battle.Mode, 1)})
		case "enter":
			if len(o.battle.Players) < online.MinPlayers {
				o.notice = "Need at least 2 players"
				return m, nil
			}
			o.notice = ""
			o.client.Send(online.Message{Type: online.TypeStart})
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
			// msg.Runes may hold several characters when typing fast
			for _, r := range msg.Runes {
				if r >= 'A' && r <= 'Z' {
					r += 'a' - 'A'
				}
				if r >= 'a' && r <= 'z' {
					o.battle.AddChar(r)
				}
			}
		}

	case online.BattleResult:
		if key == "enter" && o.battle.IsHost() {
			o.client.Send(online.Message{Type: online.TypeBackToLobby})
		}
	}
	return m, nil
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

// stepMode returns the mode delta positions away from the current one
func stepMode(current string, delta int) string {
	count := len(online.Modes)
	for i, mode := range online.Modes {
		if mode == current {
			return online.Modes[(i+delta+count)%count]
		}
	}
	return online.Modes[0]
}

func (m model) onlineView() string {
	o := m.online
	switch o.screen {
	case onlineMenu:
		return ui.RenderOnlineMenu(o.rooms, o.selectedIndex(), o.notice, o.connecting, m.animFrame)
	case onlineRoom:
		switch o.battle.Phase {
		case online.BattleLobby:
			return ui.RenderLobby(o.battle, o.address, o.notice, o.editingName, o.nameInput)
		case online.BattleResult:
			return ui.RenderBattleResult(o.battle)
		default:
			return ui.RenderBattle(o.battle, time.Now())
		}
	}
	return ""
}
