package online

import (
	"strconv"
	"time"
)

// BattlePhase is the client-side phase of a room
type BattlePhase int

const (
	BattleLobby BattlePhase = iota
	BattleCountdown
	BattleFalling
	BattleBetween
	BattleResult
)

// Battle is the client-side view of a room: lobby, running game and result.
// It is driven by server messages and owned by the UI goroutine.
type Battle struct {
	MyID    int
	Phase   BattlePhase
	Players []PlayerInfo
	Mode    string

	TotalWords int
	WordIndex  int
	Word       string
	Input      string
	Scores     map[int]int

	// Feedback is the outcome of the last word, shown until the next one spawns
	Feedback    string
	FeedbackHit bool

	Ranking []RankEntry
	Winners []int

	countdownEnd time.Time
	spawnedAt    time.Time
	fall         time.Duration
}

// NewBattle creates the state for a freshly joined room
func NewBattle(myID int) *Battle {
	return &Battle{MyID: myID, Mode: ModeWordDrop, Scores: make(map[int]int)}
}

// IsHost reports whether the local player is the host
func (b *Battle) IsHost() bool {
	return b.MyID == hostPlayerID
}

// Apply updates the state from a server message
func (b *Battle) Apply(msg Message, now time.Time) {
	switch msg.Type {
	case TypeLobby:
		b.Phase = BattleLobby
		b.Players = msg.Players
		b.Mode = msg.Mode
	case TypeGameStart:
		b.Phase = BattleCountdown
		b.Mode = msg.Mode
		b.TotalWords = msg.TotalWords
		b.WordIndex = 0
		b.Word = ""
		b.Input = ""
		b.Feedback = ""
		b.Scores = make(map[int]int)
		b.Ranking = nil
		b.Winners = nil
		b.countdownEnd = now.Add(time.Duration(msg.CountdownMs) * time.Millisecond)
	case TypeWordSpawn:
		b.Phase = BattleFalling
		b.WordIndex = msg.Index
		b.Word = msg.Word
		b.Input = ""
		b.Feedback = ""
		b.spawnedAt = now
		b.fall = time.Duration(msg.FallMs) * time.Millisecond
	case TypeWordClaimed:
		b.Phase = BattleBetween
		b.Scores[msg.PlayerID] += msg.Points
		b.Feedback = b.PlayerName(msg.PlayerID) + " +" + strconv.Itoa(msg.Points)
		b.FeedbackHit = true
		b.Word = ""
		b.Input = ""
	case TypeWordExpired:
		b.Phase = BattleBetween
		b.Feedback = "Missed!"
		b.FeedbackHit = false
		b.Word = ""
		b.Input = ""
	case TypeGameOver:
		b.Phase = BattleResult
		b.Word = ""
		b.Input = ""
		b.Ranking = msg.Ranking
		b.Winners = msg.Winners
	}
}

// AddChar appends a letter to the input while a word is falling
func (b *Battle) AddChar(r rune) {
	if b.Phase == BattleFalling {
		b.Input += string(r)
	}
}

// Backspace removes the last input character
func (b *Battle) Backspace() {
	if b.Phase == BattleFalling && len(b.Input) > 0 {
		b.Input = b.Input[:len(b.Input)-1]
	}
}

// Submit handles Enter. It returns the claim to send when the input is correct;
// a wrong input is simply cleared and nothing is sent.
func (b *Battle) Submit() (Message, bool) {
	if b.Phase != BattleFalling {
		return Message{}, false
	}
	input := b.Input
	b.Input = ""
	if input != b.Word {
		return Message{}, false
	}
	return Message{Type: TypeClaim, Index: b.WordIndex, Word: b.Word}, true
}

// FallProgress returns how far the current word has fallen, in [0, 1]
func (b *Battle) FallProgress(now time.Time) float64 {
	if b.fall <= 0 {
		return 0
	}
	progress := float64(now.Sub(b.spawnedAt)) / float64(b.fall)
	if progress < 0 {
		return 0
	}
	if progress > 1 {
		return 1
	}
	return progress
}

// CountdownSeconds returns the whole seconds left before the first word
func (b *Battle) CountdownSeconds(now time.Time) int {
	remaining := b.countdownEnd.Sub(now)
	if remaining <= 0 {
		return 0
	}
	return int((remaining + time.Second - 1) / time.Second)
}

// IsWinner reports whether the given player is among the winners
func (b *Battle) IsWinner(id int) bool {
	for _, w := range b.Winners {
		if w == id {
			return true
		}
	}
	return false
}

// PlayerName returns the name of the given player
func (b *Battle) PlayerName(id int) string {
	for _, p := range b.Players {
		if p.ID == id {
			return p.Name
		}
	}
	return "Player"
}
