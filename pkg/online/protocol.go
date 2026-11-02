// Package online implements the room server, client and battle state
// for online multiplayer modes.
package online

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

// Message types exchanged between client and server
const (
	TypeJoin         = "join"
	TypeJoinAccepted = "join_accepted"
	TypeJoinRejected = "join_rejected"
	TypeLobby        = "lobby"
	TypeSelectMode   = "select_mode"
	TypeRename       = "rename"
	TypeStart        = "start"
	TypeGameStart    = "game_start"
	TypeWordSpawn    = "word_spawn"
	TypeClaim        = "claim"
	TypeWordClaimed  = "word_claimed"
	TypeWordExpired  = "word_expired"
	TypeGameOver     = "game_over"
	TypeBackToLobby  = "back_to_lobby"
	TypeHostClosed   = "host_closed"

	// TypeDisconnected is synthesized locally by Client when the connection drops
	TypeDisconnected = "disconnected"
)

// Reasons carried by join_rejected
const (
	ReasonBadRoom  = "bad_room"
	ReasonRoomFull = "room_full"
	ReasonInGame   = "in_game"
)

// Online game modes
const (
	ModeWordDrop = "word_drop"
)

// Modes lists the modes a host can choose from
var Modes = []string{ModeWordDrop}

// ModeName returns the display name of a mode
func ModeName(mode string) string {
	switch mode {
	case ModeWordDrop:
		return "Word Drop"
	}
	return mode
}

// PlayerInfo describes a player in the lobby
type PlayerInfo struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	IsHost bool   `json:"is_host,omitempty"`
}

// RankEntry is one line of the final ranking
type RankEntry struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
	Words int    `json:"words"`
}

// Message is the single wire format; only the fields relevant to Type are set
type Message struct {
	Type        string       `json:"type"`
	Room        string       `json:"room,omitempty"`
	Name        string       `json:"name,omitempty"`
	PlayerID    int          `json:"player_id,omitempty"`
	Reason      string       `json:"reason,omitempty"`
	Players     []PlayerInfo `json:"players,omitempty"`
	Mode        string       `json:"mode,omitempty"`
	TotalWords  int          `json:"total_words,omitempty"`
	CountdownMs int          `json:"countdown_ms,omitempty"`
	Index       int          `json:"index,omitempty"`
	Word        string       `json:"word,omitempty"`
	FallMs      int          `json:"fall_ms,omitempty"`
	Points      int          `json:"points,omitempty"`
	Ranking     []RankEntry  `json:"ranking,omitempty"`
	Winners     []int        `json:"winners,omitempty"`
}

// encodeMessage returns the newline-delimited JSON form of a message
func encodeMessage(msg Message) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to encode message: %w", err)
	}
	return append(data, '\n'), nil
}

// writeMessage writes one newline-delimited JSON message
func writeMessage(conn net.Conn, msg Message) error {
	data, err := encodeMessage(msg)
	if err != nil {
		return err
	}
	_, err = conn.Write(data)
	return err
}

// readMessage reads the next message from a line scanner
func readMessage(scanner *bufio.Scanner) (Message, error) {
	var msg Message
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return msg, err
		}
		return msg, fmt.Errorf("connection closed")
	}
	if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
		return msg, fmt.Errorf("failed to decode message: %w", err)
	}
	return msg, nil
}

// LocalIP returns this machine's LAN IP, or 127.0.0.1 if it cannot be determined
func LocalIP() string {
	// Connecting a UDP socket sends no packets; it only selects the outbound interface
	if conn, err := net.Dial("udp4", "8.8.8.8:80"); err == nil {
		defer conn.Close()
		if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			return addr.IP.String()
		}
	}
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
