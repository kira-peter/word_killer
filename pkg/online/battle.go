package online

import (
	"strconv"
	"strings"
	"time"
)

// BattlePhase is the client-side phase of a room
type BattlePhase int

const (
	BattleLobby BattlePhase = iota
	BattleCountdown
	BattleFalling // a Word Drop word is falling, or a sentence or the pool is open
	BattleBetween // the pause after a word or sentence is taken
	BattleResult
)

// PoolCell is one Word Pool slot as the client sees it
type PoolCell struct {
	Word  string
	Owner int // player id, 0 while free
}

// Battle is the client-side view of a room: lobby, running game and result.
// It is driven by server messages and owned by the UI goroutine.
type Battle struct {
	MyID     int
	Phase    BattlePhase
	Players  []PlayerInfo
	Mode     string
	Settings Settings

	TotalWords int
	WordIndex  int
	Word       string // falling word, or current sentence
	Input      string
	Scores     map[int]int
	Pool       [PoolSlots]PoolCell
	// Typing holds what each other player is typing right now
	Typing map[int]string

	// sentInput is the input last reported to the server, so it is only
	// sent again when it changes
	sentInput string

	// Feedback is the outcome of the last word, shown until the next one spawns
	Feedback    string
	FeedbackHit bool
	// Hint explains why a sentence was not sent, until the input changes
	Hint string

	Ranking []RankEntry
	Winners []int

	countdownEnd time.Time
	playEnd      time.Time
	spawnedAt    time.Time
	fall         time.Duration
	colors       map[int]int // player id -> color slot, kept for players who left
}

// NewBattle creates the state for a freshly joined room
func NewBattle(myID int) *Battle {
	return &Battle{
		MyID:     myID,
		Mode:     ModeWordDrop,
		Settings: DefaultSettings(ModeWordDrop),
		Scores:   make(map[int]int),
		Typing:   make(map[int]string),
		colors:   make(map[int]int),
	}
}

// IsHost reports whether the local player is the host
func (b *Battle) IsHost() bool {
	return b.MyID == hostPlayerID
}

// Apply updates the state from a server message. Text from the server is
// stripped of control characters so a hostile host cannot inject terminal
// escape sequences.
func (b *Battle) Apply(msg Message, now time.Time) {
	msg.Word = printableASCII(msg.Word)
	switch msg.Type {
	case TypeLobby:
		b.Phase = BattleLobby
		b.Players = msg.Players
		b.Mode = msg.Mode
		if msg.Settings != nil {
			b.Settings = *msg.Settings
		}
		b.assignColors()
	case TypeGameStart:
		b.Phase = BattleCountdown
		b.Mode = msg.Mode
		if msg.Settings != nil {
			b.Settings = *msg.Settings
		}
		b.TotalWords = msg.TotalWords
		b.WordIndex = 0
		b.Word = ""
		b.clearInputs()
		b.Feedback = ""
		b.Scores = make(map[int]int)
		b.Pool = [PoolSlots]PoolCell{}
		b.Ranking = nil
		b.Winners = nil
		b.countdownEnd = now.Add(time.Duration(msg.CountdownMs) * time.Millisecond)
		b.playEnd = b.countdownEnd.Add(time.Duration(b.Settings.TimeLimit) * time.Second)
	case TypeWordSpawn:
		b.Phase = BattleFalling
		b.WordIndex = msg.Index
		b.Word = msg.Word
		b.clearInputs()
		b.Feedback = ""
		b.spawnedAt = now
		b.fall = time.Duration(msg.FallMs) * time.Millisecond
	case TypeSentence:
		b.Phase = BattleFalling
		b.WordIndex = msg.Index
		b.Word = msg.Word
		b.clearInputs()
		b.Feedback = ""
	case TypeWordClaimed:
		b.Phase = BattleBetween
		b.Scores[msg.PlayerID] += msg.Points
		b.Feedback = b.PlayerName(msg.PlayerID) + " +" + strconv.Itoa(msg.Points)
		b.FeedbackHit = true
		b.Word = ""
		b.clearInputs()
	case TypeWordExpired:
		b.Phase = BattleBetween
		b.Feedback = "Missed!"
		b.FeedbackHit = false
		b.Word = ""
		b.clearInputs()
	case TypeTyping:
		if msg.PlayerID != b.MyID {
			b.Typing[msg.PlayerID] = msg.Word
		}
	case TypePoolSpawn:
		b.Phase = BattleFalling
		if msg.Index >= 0 && msg.Index < PoolSlots {
			b.Pool[msg.Index] = PoolCell{Word: msg.Word}
		}
	case TypePoolClaimed:
		if msg.Index >= 0 && msg.Index < PoolSlots {
			b.Pool[msg.Index].Owner = msg.PlayerID
			b.Scores[msg.PlayerID] += msg.Points
			if msg.From != 0 {
				b.Scores[msg.From] -= msg.Points
			}
		}
	case TypeGameOver:
		b.Phase = BattleResult
		b.Word = ""
		b.clearInputs()
		b.Ranking = msg.Ranking
		b.Winners = msg.Winners
	}
}

// clearInputs empties every player's input when the target changes. The
// word or sentence they were typing is gone, and so is their progress on it.
func (b *Battle) clearInputs() {
	b.Input = ""
	b.Hint = ""
	b.sentInput = ""
	b.Typing = make(map[int]string)
}

// TypingUpdate returns a typing message when the local input has changed
// since it was last reported. The UI calls it once per tick, which keeps the
// rate at one message per frame however fast the player types.
func (b *Battle) TypingUpdate() (Message, bool) {
	if b.Phase != BattleFalling || b.Input == b.sentInput {
		return Message{}, false
	}
	b.sentInput = b.Input
	return Message{Type: TypeTyping, Word: b.Input}, true
}

// assignColors gives each new player the lowest free color slot. Slots stay
// with a player for the whole room so colors never shift mid-game.
func (b *Battle) assignColors() {
	used := make(map[int]bool)
	present := make(map[int]bool)
	for _, p := range b.Players {
		present[p.ID] = true
	}
	for id, c := range b.colors {
		if present[id] {
			used[c] = true
		} else {
			delete(b.colors, id)
		}
	}
	for _, p := range b.Players {
		if _, ok := b.colors[p.ID]; ok {
			continue
		}
		c := 0
		for used[c] {
			c++
		}
		b.colors[p.ID] = c
		used[c] = true
	}
}

// ColorSlot returns the color index (0-3) of a player
func (b *Battle) ColorSlot(id int) int {
	return b.colors[id]
}

// AddChar appends typed text while play is open. Sentence Race keeps every
// printable character; the word modes keep letters only, lowercased.
func (b *Battle) AddChar(r rune) {
	if b.Phase != BattleFalling {
		return
	}
	if b.Mode == ModeSentenceRace {
		// Nothing past the sentence's length can be right, so cap the input there
		if r >= ' ' && r <= '~' && len(b.Input) < len(b.Word) {
			b.Input += string(r)
			b.Hint = ""
		}
		return
	}
	if r >= 'A' && r <= 'Z' {
		r += 'a' - 'A'
	}
	if r >= 'a' && r <= 'z' {
		b.Input += string(r)
	}
}

// Backspace removes the last input character
func (b *Battle) Backspace() {
	if b.Phase == BattleFalling && len(b.Input) > 0 {
		b.Input = b.Input[:len(b.Input)-1]
		b.Hint = ""
	}
}

// Submit handles Enter and returns the claim to send, if any. A wrong word
// is cleared; a wrong sentence is kept so the player can fix it.
func (b *Battle) Submit() (Message, bool) {
	if b.Phase != BattleFalling {
		return Message{}, false
	}

	switch b.Mode {
	case ModeSentenceRace:
		if b.Input != b.Word {
			b.Hint = "Not yet - fix the red part"
			return Message{}, false
		}
		return Message{Type: TypeClaim, Index: b.WordIndex, Word: b.Word}, true

	case ModeWordPool:
		input := b.Input
		b.Input = ""
		if slot := b.claimableSlot(input); slot >= 0 {
			return Message{Type: TypeClaim, Index: slot, Word: input}, true
		}
		return Message{}, false

	default:
		input := b.Input
		b.Input = ""
		if input != b.Word {
			return Message{}, false
		}
		return Message{Type: TypeClaim, Index: b.WordIndex, Word: b.Word}, true
	}
}

// claimableSlot finds the pool slot the input may take: a free word, or
// someone else's word when stealing is on. Free words are preferred.
func (b *Battle) claimableSlot(input string) int {
	if input == "" {
		return -1
	}
	stealable := -1
	for i, c := range b.Pool {
		if c.Word != input {
			continue
		}
		if c.Owner == 0 {
			return i
		}
		if c.Owner != b.MyID && b.Settings.AllowSteal && stealable < 0 {
			stealable = i
		}
	}
	return stealable
}

// SentenceMatch returns how many leading input characters match the sentence
func (b *Battle) SentenceMatch() int {
	n := 0
	for n < len(b.Input) && n < len(b.Word) && b.Input[n] == b.Word[n] {
		n++
	}
	return n
}

// MatchesPrefix reports whether the input so far starts the given word
func (b *Battle) MatchesPrefix(word string) bool {
	return b.Input != "" && strings.HasPrefix(word, b.Input)
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
	return ceilSeconds(b.countdownEnd.Sub(now))
}

// RemainingSeconds returns the whole seconds left in the game's time limit
func (b *Battle) RemainingSeconds(now time.Time) int {
	if b.Phase == BattleCountdown {
		return b.Settings.TimeLimit
	}
	return ceilSeconds(b.playEnd.Sub(now))
}

func ceilSeconds(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	return int((d + time.Second - 1) / time.Second)
}

// printableASCII keeps only printable ASCII; every word and sentence the
// game deals is plain ASCII, so nothing legitimate is lost
func printableASCII(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= ' ' && r <= '~' {
			return r
		}
		return -1
	}, s)
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
