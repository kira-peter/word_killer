package online

import "time"

// wordDrop drops one word at a time; the first correct claim takes it
type wordDrop struct {
	host    modeHost
	words   []string
	fall    time.Duration
	pause   time.Duration
	current int
	falling bool // false during the pause between words
}

func newWordDrop(host modeHost, words []string, fall, pause time.Duration) *wordDrop {
	return &wordDrop{host: host, words: words, fall: fall, pause: pause}
}

func (w *wordDrop) totalItems() int {
	return len(w.words)
}

func (w *wordDrop) begin() {
	w.spawn()
}

func (w *wordDrop) spawn() {
	w.falling = true
	w.host.broadcast(Message{
		Type:   TypeWordSpawn,
		Index:  w.current,
		Word:   w.words[w.current],
		FallMs: int(w.fall.Milliseconds()),
	})
	w.host.setModeTimer(w.fall)
}

func (w *wordDrop) onTimer() {
	if w.falling {
		// The word reached the bottom unclaimed
		w.host.broadcast(Message{Type: TypeWordExpired, Index: w.current})
		w.endWord()
		return
	}
	w.current++
	if w.current >= len(w.words) {
		w.host.finishGame()
		return
	}
	w.spawn()
}

// endWord starts the pause before the next word
func (w *wordDrop) endWord() {
	w.falling = false
	w.host.setModeTimer(w.pause)
}

// onClaim awards the current word to the first valid claim to arrive.
// Later or stale claims fail the falling/index check and are ignored.
func (w *wordDrop) onClaim(p *serverPlayer, msg Message) {
	if !w.falling || msg.Index != w.current || msg.Word != w.words[w.current] {
		return
	}

	points := letterCount(msg.Word)
	w.host.award(p, points, 1)
	w.host.broadcast(Message{
		Type:     TypeWordClaimed,
		Index:    w.current,
		PlayerID: p.id,
		Points:   points,
	})
	w.endWord()
}
