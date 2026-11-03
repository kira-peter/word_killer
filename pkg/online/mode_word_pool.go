package online

import (
	"math/rand"
	"time"
)

const (
	// PoolSlots is the size of the Word Pool board
	PoolSlots = 36
	// PoolColumns is how many slots each board row holds
	PoolColumns = 6
	// MaxPoolWordLength keeps every word inside its board cell
	MaxPoolWordLength = 11
)

type poolSlot struct {
	word    string
	owner   int // player id, 0 while free
	spawned int // spawn sequence number, lower is older
}

// wordPool fills a board with words over time; typing a word takes it, and
// with stealing on, typing someone else's word takes it from them
type wordPool struct {
	host     modeHost
	source   func(n int) []string
	interval time.Duration
	steal    bool
	rng      *rand.Rand

	slots   [PoolSlots]poolSlot
	seq     int
	onBoard map[string]bool
	owners  map[int]*serverPlayer
}

func newWordPool(host modeHost, source func(n int) []string, spawnRate int, steal bool) *wordPool {
	return &wordPool{
		host:     host,
		source:   source,
		interval: time.Second / time.Duration(spawnRate),
		steal:    steal,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
		onBoard:  make(map[string]bool),
		owners:   make(map[int]*serverPlayer),
	}
}

func (w *wordPool) totalItems() int {
	return 0
}

func (w *wordPool) begin() {
	w.spawn()
	w.host.setModeTimer(w.interval)
}

func (w *wordPool) onTimer() {
	w.spawn()
	w.host.setModeTimer(w.interval)
}

// spawn puts one new word on the board: in a random empty slot, else over
// the oldest free word. A board where every word is owned gets nothing.
func (w *wordPool) spawn() {
	slot := w.pickSlot()
	if slot < 0 {
		return
	}
	word := w.freshWord()
	if word == "" {
		return
	}

	if old := w.slots[slot].word; old != "" {
		delete(w.onBoard, old)
	}
	w.seq++
	w.slots[slot] = poolSlot{word: word, spawned: w.seq}
	w.onBoard[word] = true
	w.host.broadcast(Message{Type: TypePoolSpawn, Index: slot, Word: word})
}

func (w *wordPool) pickSlot() int {
	var empty []int
	oldest := -1
	for i, s := range w.slots {
		switch {
		case s.word == "":
			empty = append(empty, i)
		case s.owner == 0 && (oldest < 0 || s.spawned < w.slots[oldest].spawned):
			oldest = i
		}
	}
	if len(empty) > 0 {
		return empty[w.rng.Intn(len(empty))]
	}
	return oldest
}

// freshWord returns a word that fits a cell and is not already on the board
func (w *wordPool) freshWord() string {
	for attempt := 0; attempt < 20; attempt++ {
		for _, word := range w.source(1) {
			if len(word) <= MaxPoolWordLength && !w.onBoard[word] {
				return word
			}
		}
	}
	return ""
}

// onClaim takes a free word, or another player's word when stealing is on
func (w *wordPool) onClaim(p *serverPlayer, msg Message) {
	if msg.Index < 0 || msg.Index >= PoolSlots {
		return
	}
	slot := &w.slots[msg.Index]
	if slot.word == "" || slot.word != msg.Word || slot.owner == p.id {
		return
	}
	if slot.owner != 0 && !w.steal {
		return
	}

	points := letterCount(slot.word)
	from := slot.owner
	if prev, ok := w.owners[from]; ok {
		w.host.award(prev, -points, -1)
	}
	slot.owner = p.id
	w.owners[p.id] = p
	w.host.award(p, points, 1)

	w.host.broadcast(Message{
		Type:     TypePoolClaimed,
		Index:    msg.Index,
		PlayerID: p.id,
		From:     from,
		Points:   points,
	})
}
