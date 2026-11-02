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
	maxHostLength        = 64
	roomNumberLength     = 4
)

// onlineScreen identifies which online screen is showing
type onlineScreen int

const (
	onlineNone onlineScreen = iota
	onlineMenu
	onlineJoinForm
	onlineRoom // lobby, battle or result, depending on battle.Phase
)

// Online menu options
const (
	onlineMenuCreate = iota
	onlineMenuJoin
	onlineMenuBack
	onlineMenuCount
)

// Join form fields
const (
	joinFieldHost = iota
	joinFieldRoom
	joinFieldCount
)

// onlineState holds everything the online screens need
type onlineState struct {
	screen     onlineScreen
	menuIndex  int
	notice     string // shown on the current screen until the next action
	host       string
	room       string
	joinField  int
	connecting bool
	address    string // host:port shown in the lobby

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

// enterRoom switches to the room screen using an established connection
func (o *onlineState) enterRoom(client *online.Client, address string) {
	o.client = client
	o.battle = online.NewBattle(client.PlayerID, client.Room)
	o.address = address
	o.screen = onlineRoom
	o.notice = ""
	o.editingName = false
}

// leaveRoom releases the connection and server, then returns to the online menu
func (o *onlineState) leaveRoom(notice string) {
	o.closeConnections()
	o.battle = nil
	o.editingName = false
	o.screen = onlineMenu
	o.menuIndex = 0
	o.notice = notice
}

func (o *onlineState) closeConnections() {
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
	client, err := online.Connect(loopback, server.Room(), m.cfg.OnlinePlayerName)
	if err != nil {
		server.Close()
		m.online.notice = "Cannot open room on port " + strconv.Itoa(m.cfg.OnlinePort)
		return
	}

	m.online.server = server
	m.online.enterRoom(client, net.JoinHostPort(online.LocalIP(), strconv.Itoa(server.Port())))
}

// joinCmd connects in the background so the UI keeps rendering meanwhile
func joinCmd(address, room, name string) tea.Cmd {
	return func() tea.Msg {
		client, err := online.Connect(address, room, name)
		return joinResultMsg{client: client, err: err}
	}
}

func (m model) handleJoinResult(msg joinResultMsg) {
	m.online.connecting = false
	if msg.err != nil {
		var rejected *online.RejectedError
		if errors.As(msg.err, &rejected) {
			m.online.notice = rejected.Error()
		} else {
			m.online.notice = "Cannot connect to host"
		}
		return
	}
	address := net.JoinHostPort(m.online.host, strconv.Itoa(m.cfg.OnlinePort))
	m.online.enterRoom(msg.client, address)
}

// pollOnline applies pending server messages; called on every tick
func (m model) pollOnline() {
	o := m.online
	if o.client == nil {
		return
	}
	now := time.Now()
	for _, msg := range o.client.Poll() {
		switch msg.Type {
		case online.TypeHostClosed:
			o.leaveRoom("Host closed the room")
			return
		case online.TypeDisconnected:
			o.leaveRoom("Connection lost")
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
	case onlineJoinForm:
		return m.handleJoinFormKey(msg)
	case onlineRoom:
		return m.handleRoomKey(msg)
	}
	return m, nil
}

func (m model) handleOnlineMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	o := m.online
	switch msg.String() {
	case "up", "k":
		o.menuIndex = (o.menuIndex - 1 + onlineMenuCount) % onlineMenuCount
	case "down", "j":
		o.menuIndex = (o.menuIndex + 1) % onlineMenuCount
	case "enter":
		o.notice = ""
		switch o.menuIndex {
		case onlineMenuCreate:
			m.createRoom()
		case onlineMenuJoin:
			o.screen = onlineJoinForm
			o.joinField = joinFieldHost
			o.room = ""
		case onlineMenuBack:
			o.screen = onlineNone
		}
	case "esc":
		o.notice = ""
		o.screen = onlineNone
	}
	return m, nil
}

func (m model) handleJoinFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	o := m.online
	// A join is in flight; its result arrives as joinResultMsg
	if o.connecting {
		return m, nil
	}

	switch msg.String() {
	case "esc":
		o.notice = ""
		o.screen = onlineMenu
	case "tab", "down":
		o.joinField = (o.joinField + 1) % joinFieldCount
	case "shift+tab", "up":
		o.joinField = (o.joinField - 1 + joinFieldCount) % joinFieldCount
	case "backspace":
		if o.joinField == joinFieldHost && len(o.host) > 0 {
			o.host = o.host[:len(o.host)-1]
		} else if o.joinField == joinFieldRoom && len(o.room) > 0 {
			o.room = o.room[:len(o.room)-1]
		}
	case "enter":
		if o.host == "" {
			o.notice = "Please enter the host IP"
			o.joinField = joinFieldHost
			return m, nil
		}
		if len(o.room) != roomNumberLength {
			o.notice = "Room number must be 4 digits"
			o.joinField = joinFieldRoom
			return m, nil
		}
		o.notice = ""
		o.connecting = true
		address := net.JoinHostPort(o.host, strconv.Itoa(m.cfg.OnlinePort))
		return m, joinCmd(address, o.room, m.cfg.OnlinePlayerName)
	default:
		// msg.Runes may hold several characters when pasting or typing fast
		for _, r := range msg.Runes {
			if o.joinField == joinFieldHost {
				if r > ' ' && r <= '~' && len(o.host) < maxHostLength {
					o.host += string(r)
				}
			} else if r >= '0' && r <= '9' && len(o.room) < roomNumberLength {
				o.room += string(r)
			}
		}
	}
	return m, nil
}

func (m model) handleRoomKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	o := m.online
	key := msg.String()

	if o.editingName {
		return m.handleNameEditKey(msg)
	}

	if key == "esc" {
		o.leaveRoom("")
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
		return ui.RenderOnlineMenu(o.menuIndex, o.notice, m.animFrame)
	case onlineJoinForm:
		return ui.RenderJoinForm(o.host, o.room, o.joinField, o.notice, o.connecting)
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
