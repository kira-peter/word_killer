package online

import "time"

// sentenceRace shows one sentence at a time; the first player to type it
// exactly takes it, scoring its letters
type sentenceRace struct {
	host      modeHost
	sentences []string
	pause     time.Duration
	current   int
	open      bool // false during the pause between sentences
}

func newSentenceRace(host modeHost, sentences []string, pause time.Duration) *sentenceRace {
	return &sentenceRace{host: host, sentences: sentences, pause: pause}
}

func (r *sentenceRace) totalItems() int {
	return len(r.sentences)
}

func (r *sentenceRace) begin() {
	r.show()
}

func (r *sentenceRace) show() {
	r.open = true
	r.host.broadcast(Message{Type: TypeSentence, Index: r.current, Word: r.sentences[r.current]})
}

// onTimer only ends the pause after a claimed sentence; an open sentence
// has no deadline of its own, the overall time limit covers it
func (r *sentenceRace) onTimer() {
	r.current++
	if r.current >= len(r.sentences) {
		r.host.finishGame()
		return
	}
	r.show()
}

func (r *sentenceRace) onClaim(p *serverPlayer, msg Message) {
	if !r.open || msg.Index != r.current || msg.Word != r.sentences[r.current] {
		return
	}

	points := letterCount(msg.Word)
	r.host.award(p, points, 1)
	r.host.broadcast(Message{
		Type:     TypeWordClaimed,
		Index:    r.current,
		PlayerID: p.id,
		Points:   points,
	})
	r.open = false
	r.host.setModeTimer(r.pause)
}
