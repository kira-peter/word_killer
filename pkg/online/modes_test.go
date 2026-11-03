package online

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSettingsDefaultsAndSteps(t *testing.T) {
	for _, mode := range Modes {
		defaults := DefaultSettings(mode)
		for _, o := range ModeOptions(mode) {
			if o.Index(defaults.Get(o.Key)) < 0 {
				t.Errorf("%s %s default %d is not a step", mode, o.Key, defaults.Get(o.Key))
			}
		}
	}

	s := DefaultSettings(ModeWordPool)
	if _, ok := s.Apply(ModeWordPool, OptionSpawnRate, 7); ok {
		t.Error("an off-step value must be rejected")
	}
	if _, ok := s.Apply(ModeWordPool, OptionCount, 10); ok {
		t.Error("a setting the mode lacks must be rejected")
	}
	s, ok := s.Apply(ModeWordPool, OptionAllowSteal, 1)
	if !ok || !s.AllowSteal {
		t.Errorf("steal on: %+v ok=%v", s, ok)
	}

	steal := ModeOptions(ModeWordPool)[2]
	if steal.Format(0) != "Off" || steal.Format(1) != "On" || ModeOptions(ModeWordDrop)[1].Format(120) != "120s" {
		t.Error("format")
	}
}

func TestSetOptionRules(t *testing.T) {
	s := startTestServer(t, time.Second)
	host, guest := joinTwo(t, s)

	lobby := setOption(t, host, OptionTimeLimit, 300, "")
	if lobby.Settings.TimeLimit != 300 || lobby.Mode != ModeWordDrop {
		t.Fatalf("lobby = %+v", lobby)
	}

	// Off-step values and guests are ignored; the next accepted change proves it
	host.Send(Message{Type: TypeSetOption, Option: OptionCount, Value: 7})
	guest.Send(Message{Type: TypeSetOption, Option: OptionCount, Value: 20})
	lobby = setOption(t, host, OptionTimeLimit, 60, "")
	if lobby.Settings.Count != DefaultSettings(ModeWordDrop).Count {
		t.Errorf("count = %d, want the default", lobby.Settings.Count)
	}

	// A new mode starts from its defaults
	lobby = setOption(t, host, OptionMode, 0, ModeWordPool)
	if *lobby.Settings != DefaultSettings(ModeWordPool) {
		t.Errorf("pool settings = %+v", lobby.Settings)
	}
	host.Send(Message{Type: TypeSetOption, Option: OptionMode, Mode: "chess"})
	lobby = setOption(t, host, OptionSpawnRate, 3, "")
	if lobby.Mode != ModeWordPool {
		t.Errorf("unknown mode applied: %s", lobby.Mode)
	}
}

func TestKick(t *testing.T) {
	s := startTestServer(t, time.Second)
	host, guest := joinTwo(t, s)
	third := joinTest(t, s, "Third")
	waitForLobbySize(t, host, 3)

	// Guests cannot kick, and the host cannot kick itself
	guest.Send(Message{Type: TypeKick, PlayerID: third.PlayerID})
	host.Send(Message{Type: TypeKick, PlayerID: host.PlayerID})

	host.Send(Message{Type: TypeKick, PlayerID: guest.PlayerID})
	waitFor(t, guest, TypeKicked)
	waitFor(t, guest, TypeDisconnected)
	waitForLobbySize(t, host, 2)

	lobby, _ := waitFor(t, third, TypeLobby)
	for lobby.Players != nil && len(lobby.Players) != 2 {
		lobby, _ = waitFor(t, third, TypeLobby)
	}
	for _, p := range lobby.Players {
		if p.ID == guest.PlayerID {
			t.Errorf("kicked player still listed: %+v", lobby.Players)
		}
	}

	// The kicked player may come back
	again := joinTest(t, s, "Guest")
	if again.PlayerID == guest.PlayerID {
		t.Error("a returning player gets a new id")
	}
	waitForLobbySize(t, host, 3)
}

func TestKickIgnoredDuringGame(t *testing.T) {
	s := startTestServer(t, 5*time.Second)
	host, guest := startTwoPlayerGame(t, s)
	waitFor(t, host, TypeWordSpawn)

	host.Send(Message{Type: TypeKick, PlayerID: guest.PlayerID})
	guest.Send(Message{Type: TypeClaim, Index: 0, Word: testWords[0]})
	if claimed, _ := waitFor(t, host, TypeWordClaimed); claimed.PlayerID != guest.PlayerID {
		t.Errorf("guest should still be playing: %+v", claimed)
	}
}

func TestTimeLimitEndsGame(t *testing.T) {
	s := startTestServer(t, 5*time.Second)
	host, _ := startTwoPlayerGame(t, s)
	waitFor(t, host, TypeWordSpawn)

	// The smallest preset is a minute; fire the limit instead of waiting
	s.limitForTest()
	over, seen := waitFor(t, host, TypeGameOver)
	if len(over.Ranking) != 2 {
		t.Errorf("ranking = %+v", over.Ranking)
	}
	for _, msg := range seen {
		if msg.Type == TypeWordExpired || msg.Type == TypeWordClaimed {
			t.Errorf("the word in flight must not score: %+v", msg)
		}
	}
}

func TestStartFailsWithoutSentences(t *testing.T) {
	s, err := NewServer(ServerConfig{
		Host:           "127.0.0.1",
		StartCountdown: 10 * time.Millisecond,
		WordSource:     func(n int) []string { return testWords[:min(n, len(testWords))] },
		SentenceSource: func(n int) []string { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	host, guest := joinTwo(t, s)
	setOption(t, host, OptionMode, 0, ModeSentenceRace)
	host.Send(Message{Type: TypeStart})

	failed, seen := waitFor(t, host, TypeStartFailed)
	if failed.Reason != "No sentences available" {
		t.Errorf("reason = %q", failed.Reason)
	}
	for _, msg := range seen {
		if msg.Type == TypeGameStart {
			t.Error("the game must not start")
		}
	}
	// Only the host is told; the room stays in the lobby
	guest.Send(Message{Type: TypeRename, Name: "Still"})
	lobby, guestSeen := waitFor(t, guest, TypeLobby)
	for lobby.Players[1].Name != "Still" {
		lobby, guestSeen = waitFor(t, guest, TypeLobby)
	}
	for _, msg := range guestSeen {
		if msg.Type == TypeStartFailed {
			t.Error("guests are not told about a failed start")
		}
	}
}

func TestBattleStripsControlCharacters(t *testing.T) {
	b := NewBattle(2)
	b.Apply(Message{Type: TypeSentence, Word: "Hi\x1b[31m there\x07"}, time.Now())
	if b.Word != "Hi[31m there" {
		t.Errorf("sentence = %q", b.Word)
	}
	b.Apply(Message{Type: TypePoolSpawn, Index: 4, Word: "go\x1b]0;x\x07"}, time.Now())
	if b.Pool[4].Word != "go]0;x" {
		t.Errorf("pool word = %q", b.Pool[4].Word)
	}
}

func TestBattleSentenceInput(t *testing.T) {
	b := NewBattle(2)
	b.Apply(Message{Type: TypeGameStart, Mode: ModeSentenceRace, Settings: &Settings{Count: 3, TimeLimit: 60}}, time.Now())
	b.Apply(Message{Type: TypeSentence, Index: 0, Word: "Go, fast!"}, time.Now())

	for _, r := range "Go, fadt!" {
		b.AddChar(r)
	}
	if b.SentenceMatch() != 6 {
		t.Errorf("match = %d, want 6", b.SentenceMatch())
	}
	if _, ok := b.Submit(); ok || b.Input == "" || b.Hint == "" {
		t.Errorf("a wrong sentence keeps the input: input=%q hint=%q", b.Input, b.Hint)
	}
	for i := 0; i < 3; i++ {
		b.Backspace()
	}
	if b.Hint != "" {
		t.Error("editing clears the hint")
	}
	for _, r := range "st!" {
		b.AddChar(r)
	}
	claim, ok := b.Submit()
	if !ok || claim.Word != "Go, fast!" || claim.Index != 0 {
		t.Errorf("claim = %+v ok=%v", claim, ok)
	}
}

func TestBattlePoolClaimable(t *testing.T) {
	b := NewBattle(2)
	b.Apply(Message{Type: TypeGameStart, Mode: ModeWordPool, Settings: &Settings{TimeLimit: 60, SpawnRate: 1}}, time.Now())
	b.Apply(Message{Type: TypePoolSpawn, Index: 3, Word: "cat"}, time.Now())
	b.Apply(Message{Type: TypePoolSpawn, Index: 7, Word: "dog"}, time.Now())
	b.Apply(Message{Type: TypePoolClaimed, Index: 7, PlayerID: 1, Points: 3}, time.Now())

	for _, tc := range []struct {
		input string
		steal bool
		slot  int
	}{
		{"cat", false, 3},
		{"dog", false, -1}, // owned by someone else, stealing off
		{"dog", true, 7},
		{"cow", true, -1},
	} {
		b.Settings.AllowSteal = tc.steal
		for _, r := range tc.input {
			b.AddChar(r)
		}
		claim, ok := b.Submit()
		if tc.slot < 0 && ok || tc.slot >= 0 && (!ok || claim.Index != tc.slot) {
			t.Errorf("%s steal=%v: claim=%+v ok=%v", tc.input, tc.steal, claim, ok)
		}
		if b.Input != "" {
			t.Errorf("word modes clear the input, got %q", b.Input)
		}
	}

	// Own words are never claimable
	b.Apply(Message{Type: TypePoolClaimed, Index: 3, PlayerID: 2, Points: 3}, time.Now())
	b.Settings.AllowSteal = true
	for _, r := range "cat" {
		b.AddChar(r)
	}
	if _, ok := b.Submit(); ok {
		t.Error("own word must not be claimable")
	}
	if b.Scores[2] != 3 || b.Scores[1] != 3 {
		t.Errorf("scores = %v", b.Scores)
	}
}

func TestBattleColorsStable(t *testing.T) {
	b := NewBattle(1)
	lobby := func(ids ...int) {
		var players []PlayerInfo
		for _, id := range ids {
			players = append(players, PlayerInfo{ID: id})
		}
		b.Apply(Message{Type: TypeLobby, Players: players, Mode: ModeWordDrop}, time.Now())
	}
	lobby(1, 2, 3)
	if b.ColorSlot(1) != 0 || b.ColorSlot(2) != 1 || b.ColorSlot(3) != 2 {
		t.Fatalf("slots = %v", b.colors)
	}
	// A leaving player frees a slot; the others keep theirs
	lobby(1, 3)
	lobby(1, 3, 4)
	if b.ColorSlot(3) != 2 || b.ColorSlot(4) != 1 {
		t.Errorf("slots = %v", b.colors)
	}
}

func TestTypingRelay(t *testing.T) {
	s := startTestServer(t, 5*time.Second)
	host, guest := startTwoPlayerGame(t, s)

	// Typing before play starts is not relayed
	host.Send(Message{Type: TypeTyping, Word: "early"})
	waitFor(t, guest, TypeWordSpawn)

	host.Send(Message{Type: TypeTyping, Word: "alp\x1b[31m"})
	typing, seen := waitFor(t, guest, TypeTyping)
	if typing.PlayerID != host.PlayerID || typing.Word != "alp[31m" {
		t.Errorf("typing = %+v", typing)
	}
	for _, msg := range seen {
		if msg.Type == TypeTyping {
			t.Errorf("typing before play was relayed: %+v", msg)
		}
	}

	// Oversized input is cut, and the sender does not get its own echo
	host.Send(Message{Type: TypeTyping, Word: strings.Repeat("a", 500)})
	long, _ := waitFor(t, guest, TypeTyping)
	if len(long.Word) != maxTypingLength {
		t.Errorf("relayed length = %d, want %d", len(long.Word), maxTypingLength)
	}
	guest.Send(Message{Type: TypeTyping, Word: "x"})
	waitFor(t, host, TypeTyping)
	host.Send(Message{Type: TypeClaim, Index: 0, Word: testWords[0]})
	_, hostSeen := waitFor(t, host, TypeWordClaimed)
	for _, msg := range hostSeen {
		if msg.Type == TypeTyping && msg.PlayerID == host.PlayerID {
			t.Errorf("sender got its own typing back: %+v", msg)
		}
	}
}

func TestBattleTypingState(t *testing.T) {
	b := NewBattle(2)
	b.Apply(Message{Type: TypeGameStart, Mode: ModeWordDrop, Settings: &Settings{Count: 5, TimeLimit: 60}}, time.Now())
	if _, ok := b.TypingUpdate(); ok {
		t.Error("nothing to report during the countdown")
	}

	b.Apply(Message{Type: TypeWordSpawn, Word: "hello", FallMs: 1000}, time.Now())
	b.Apply(Message{Type: TypeTyping, PlayerID: 1, Word: "he"}, time.Now())
	b.Apply(Message{Type: TypeTyping, PlayerID: 2, Word: "ignored"}, time.Now())
	if b.Typing[1] != "he" || b.Typing[2] != "" {
		t.Errorf("typing = %v", b.Typing)
	}

	// Each change is reported once, however many ticks pass
	b.AddChar('h')
	b.AddChar('e')
	msg, ok := b.TypingUpdate()
	if !ok || msg.Type != TypeTyping || msg.Word != "he" {
		t.Errorf("update = %+v ok=%v", msg, ok)
	}
	if _, ok := b.TypingUpdate(); ok {
		t.Error("an unchanged input must not be resent")
	}
	b.Backspace()
	if msg, ok := b.TypingUpdate(); !ok || msg.Word != "h" {
		t.Errorf("backspace should be reported: %+v ok=%v", msg, ok)
	}

	// A new word clears everyone's progress
	b.Apply(Message{Type: TypeWordClaimed, PlayerID: 1, Points: 5}, time.Now())
	if len(b.Typing) != 0 || b.Input != "" {
		t.Errorf("claim should clear inputs: typing=%v input=%q", b.Typing, b.Input)
	}
	b.Apply(Message{Type: TypeWordSpawn, Index: 1, Word: "world", FallMs: 1000}, time.Now())
	if _, ok := b.TypingUpdate(); ok {
		t.Error("an empty input after a new word needs no update")
	}
}

func TestSentenceRace(t *testing.T) {
	s := startTestServer(t, time.Second)
	host, guest := joinTwo(t, s)
	setOption(t, host, OptionMode, 0, ModeSentenceRace)
	setOption(t, host, OptionCount, 3, "")
	host.Send(Message{Type: TypeStart})

	start, _ := waitFor(t, guest, TypeGameStart)
	if start.Mode != ModeSentenceRace || start.TotalWords != 3 || start.Settings.Count != 3 {
		t.Fatalf("game_start = %+v", start)
	}

	first, _ := waitFor(t, guest, TypeSentence)
	if first.Word != testSentences[0] {
		t.Fatalf("first sentence = %+v", first)
	}
	// An almost-right sentence earns nothing; a repeated claim scores once
	guest.Send(Message{Type: TypeClaim, Index: 0, Word: "Go fast"})
	host.Send(Message{Type: TypeClaim, Index: 0, Word: first.Word})
	host.Send(Message{Type: TypeClaim, Index: 0, Word: first.Word})
	claimed, _ := waitFor(t, guest, TypeWordClaimed)
	if claimed.PlayerID != host.PlayerID || claimed.Points != 6 {
		t.Fatalf("claimed = %+v, want host with 6 letters", claimed)
	}

	for i := 1; i < len(testSentences); i++ {
		next, _ := waitFor(t, guest, TypeSentence)
		guest.Send(Message{Type: TypeClaim, Index: next.Index, Word: next.Word})
		waitFor(t, host, TypeWordClaimed)
	}

	over, seen := waitFor(t, host, TypeGameOver)
	for _, msg := range seen {
		if msg.Type == TypeWordClaimed && msg.Index == 0 && msg.PlayerID != host.PlayerID {
			t.Errorf("first sentence claimed twice: %+v", msg)
		}
	}
	// "Type it, now!" has 9 letters, "A b c." has 3; the host's 6 stand
	if over.Ranking[0].ID != guest.PlayerID || over.Ranking[0].Score != 12 || over.Ranking[0].Words != 2 ||
		over.Ranking[1].Score != 6 || over.Ranking[1].Words != 1 {
		t.Errorf("ranking = %+v", over.Ranking)
	}
}

func TestLetterCount(t *testing.T) {
	for in, want := range map[string]int{"Go fast.": 6, "Type it, now!": 9, "": 0, "a1-b": 2} {
		if got := letterCount(in); got != want {
			t.Errorf("letterCount(%q) = %d, want %d", in, got, want)
		}
	}
}

// fakeHost records what a mode does so its rules can be tested directly
type fakeHost struct {
	sent     []Message
	timer    time.Duration
	finished bool
}

func (h *fakeHost) broadcast(msg Message)        { h.sent = append(h.sent, msg) }
func (h *fakeHost) setModeTimer(d time.Duration) { h.timer = d }
func (h *fakeHost) finishGame()                  { h.finished = true }
func (h *fakeHost) award(p *serverPlayer, points, items int) {
	p.score += points
	p.words += items
}

func countingSource() func(n int) []string {
	next := 0
	return func(n int) []string {
		words := make([]string, n)
		for i := range words {
			words[i] = "w" + strconv.Itoa(next)
			next++
		}
		return words
	}
}

func TestWordPoolSpawnRate(t *testing.T) {
	h := &fakeHost{}
	pool := newWordPool(h, countingSource(), 3, false)
	pool.begin()
	if h.timer != time.Second/3 || len(h.sent) != 1 || h.sent[0].Type != TypePoolSpawn {
		t.Errorf("timer=%v sent=%+v", h.timer, h.sent)
	}
}

func TestWordPoolFullBoardReplacesOldestFree(t *testing.T) {
	h := &fakeHost{}
	pool := newWordPool(h, countingSource(), 1, false)
	alice := &serverPlayer{id: 2}
	for i := 0; i < PoolSlots; i++ {
		pool.onTimer()
	}

	// Own everything except the two oldest words
	byWord := map[string]int{}
	for i, s := range pool.slots {
		byWord[s.word] = i
	}
	for i := 2; i < PoolSlots; i++ {
		word := "w" + strconv.Itoa(i)
		pool.onClaim(alice, Message{Index: byWord[word], Word: word})
	}

	pool.onTimer()
	if _, ok := pool.onBoard["w0"]; ok || !pool.onBoard["w1"] || !pool.onBoard["w36"] {
		t.Errorf("the oldest free word w0 should be replaced: %v", pool.onBoard)
	}

	// Once every word is owned nothing new spawns
	pool.onClaim(alice, Message{Index: byWord["w1"], Word: "w1"})
	for i, s := range pool.slots {
		if s.owner == 0 {
			pool.onClaim(alice, Message{Index: i, Word: s.word})
		}
	}
	before := len(h.sent)
	pool.onTimer()
	if len(h.sent) != before {
		t.Errorf("a fully owned board must not spawn: %+v", h.sent[before:])
	}
	if alice.words != PoolSlots {
		t.Errorf("alice owns %d words, want %d", alice.words, PoolSlots)
	}
}

func TestWordPoolStealRules(t *testing.T) {
	for _, steal := range []bool{false, true} {
		h := &fakeHost{}
		pool := newWordPool(h, func(n int) []string { return []string{"alpha"} }, 1, steal)
		pool.begin()
		slot := h.sent[0].Index
		alice, bob := &serverPlayer{id: 2}, &serverPlayer{id: 3}

		pool.onClaim(alice, Message{Index: slot, Word: "wrong"})
		pool.onClaim(alice, Message{Index: slot, Word: "alpha"})
		pool.onClaim(alice, Message{Index: slot, Word: "alpha"}) // own word: no-op
		pool.onClaim(bob, Message{Index: slot, Word: "alpha"})

		claims := 0
		for _, m := range h.sent {
			if m.Type == TypePoolClaimed {
				claims++
			}
		}
		if !steal {
			if claims != 1 || pool.slots[slot].owner != alice.id || alice.score != 5 || bob.score != 0 {
				t.Errorf("no steal: claims=%d owner=%d alice=%d bob=%d", claims, pool.slots[slot].owner, alice.score, bob.score)
			}
			continue
		}
		last := h.sent[len(h.sent)-1]
		if claims != 2 || pool.slots[slot].owner != bob.id || last.From != alice.id {
			t.Errorf("steal: claims=%d owner=%d last=%+v", claims, pool.slots[slot].owner, last)
		}
		if alice.score != 0 || alice.words != 0 || bob.score != 5 || bob.words != 1 {
			t.Errorf("steal scores: alice=%d/%d bob=%d/%d", alice.score, alice.words, bob.score, bob.words)
		}
	}
}

func TestWordPoolSkipsLongWords(t *testing.T) {
	h := &fakeHost{}
	long := "abcdefghijklmnop"
	pool := newWordPool(h, func(n int) []string { return []string{long, "short"} }, 1, false)
	pool.begin()
	if len(h.sent) != 1 || h.sent[0].Word != "short" {
		t.Errorf("sent = %+v", h.sent)
	}
}

func TestWordPoolGame(t *testing.T) {
	s := startTestServer(t, time.Second)
	host, guest := joinTwo(t, s)
	setOption(t, host, OptionMode, 0, ModeWordPool)
	setOption(t, host, OptionSpawnRate, 3, "")
	setOption(t, host, OptionAllowSteal, 1, "")
	host.Send(Message{Type: TypeStart})

	start, _ := waitFor(t, guest, TypeGameStart)
	if start.Mode != ModeWordPool || !start.Settings.AllowSteal || start.TotalWords != 0 {
		t.Fatalf("game_start = %+v", start)
	}

	spawn, _ := waitFor(t, guest, TypePoolSpawn)
	guest.Send(Message{Type: TypeClaim, Index: spawn.Index, Word: spawn.Word})
	took, _ := waitFor(t, host, TypePoolClaimed)
	if took.PlayerID != guest.PlayerID || took.From != 0 || took.Points != len(spawn.Word) {
		t.Fatalf("pool_claimed = %+v", took)
	}

	host.Send(Message{Type: TypeClaim, Index: spawn.Index, Word: spawn.Word})
	// Look past the guest's own claim for the steal
	stolen, _ := waitFor(t, guest, TypePoolClaimed)
	for stolen.From == 0 {
		stolen, _ = waitFor(t, guest, TypePoolClaimed)
	}
	if stolen.Index != spawn.Index || stolen.PlayerID != host.PlayerID || stolen.From != guest.PlayerID {
		t.Errorf("steal = %+v", stolen)
	}

	s.limitForTest()
	over, _ := waitFor(t, host, TypeGameOver)
	if over.Ranking[0].ID != host.PlayerID || over.Ranking[0].Score != len(spawn.Word) || over.Ranking[1].Score != 0 {
		t.Errorf("ranking = %+v", over.Ranking)
	}
}
