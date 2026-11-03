package online

import "time"

// modeHost is what a game mode may do to the room: send messages, award
// points and run its single timer. The server implements it; all calls
// happen on the server's run goroutine.
type modeHost interface {
	broadcast(msg Message)
	award(p *serverPlayer, points, items int)
	setModeTimer(d time.Duration)
	finishGame()
}

// gameMode is the rules of one online game. The server runs the shared
// parts (countdown, overall time limit, disconnects, ranking) and hands
// the rest to the mode.
type gameMode interface {
	// begin starts play once the countdown is over
	begin()
	// onTimer is called when the mode's timer fires
	onTimer()
	// onClaim handles a player's submission
	onClaim(p *serverPlayer, msg Message)
	// totalItems is the number of words or sentences shown as progress, 0 if none
	totalItems() int
}

// letterCount counts the letters in s, which is how words and sentences score
func letterCount(s string) int {
	n := 0
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			n++
		}
	}
	return n
}

// itemNoun names what a mode is played with, for messages
func itemNoun(mode string) string {
	if mode == ModeSentenceRace {
		return "sentences"
	}
	return "words"
}
