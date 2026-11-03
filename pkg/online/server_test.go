package online

import (
	"bufio"
	"errors"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

// testWords has as many words as the smallest Word Drop count
var testWords = []string{"alpha", "go", "keyboard", "cat", "river"}

var testSentences = []string{"Go fast.", "Type it, now!", "A b c."}

func startTestServer(t *testing.T, drop time.Duration) *Server {
	t.Helper()
	s, err := NewServer(ServerConfig{
		Host:           "127.0.0.1",
		Port:           0,
		DropDuration:   drop,
		StartCountdown: 10 * time.Millisecond,
		NextWordDelay:  10 * time.Millisecond,
		WordSource:     func(n int) []string { return testWords[:min(n, len(testWords))] },
		SentenceSource: func(n int) []string { return testSentences[:min(n, len(testSentences))] },
	})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	t.Cleanup(s.Close)
	return s
}

func testAddr(s *Server) string {
	return net.JoinHostPort("127.0.0.1", strconv.Itoa(s.Port()))
}

func joinTest(t *testing.T, s *Server, name string) *Client {
	t.Helper()
	c, err := Connect(testAddr(s), name)
	if err != nil {
		t.Fatalf("Connect(%s): %v", name, err)
	}
	t.Cleanup(c.Close)
	return c
}

// waitFor polls until a message of the given type arrives, returning it
// along with any message seen on the way
func waitFor(t *testing.T, c *Client, msgType string) (Message, []Message) {
	t.Helper()
	var seen []Message
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case msg := <-c.incoming:
			if msg.Type == msgType {
				return msg, seen
			}
			seen = append(seen, msg)
		case <-time.After(10 * time.Millisecond):
		}
	}
	t.Fatalf("timed out waiting for %s, saw %+v", msgType, seen)
	return Message{}, nil
}

func waitForLobbySize(t *testing.T, c *Client, size int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		msg, _ := waitFor(t, c, TypeLobby)
		if len(msg.Players) == size {
			return
		}
	}
	t.Fatalf("lobby never reached %d players", size)
}

// joinTwo joins a host and a guest and waits until both see the full lobby
func joinTwo(t *testing.T, s *Server) (*Client, *Client) {
	t.Helper()
	host := joinTest(t, s, "Host")
	guest := joinTest(t, s, "Guest")
	waitForLobbySize(t, host, 2)
	waitForLobbySize(t, guest, 2)
	return host, guest
}

// setOption sends one lobby setting and waits until the lobby shows it
func setOption(t *testing.T, host *Client, key string, value int, mode string) Message {
	t.Helper()
	host.Send(Message{Type: TypeSetOption, Option: key, Value: value, Mode: mode})
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		lobby, _ := waitFor(t, host, TypeLobby)
		if key == OptionMode && lobby.Mode == mode {
			return lobby
		}
		if key != OptionMode && lobby.Settings != nil && lobby.Settings.Get(key) == value {
			return lobby
		}
	}
	t.Fatalf("setting %s=%d/%s never applied", key, value, mode)
	return Message{}
}

// startTwoPlayerGame starts Word Drop with the smallest word count, which
// uses every test word
func startTwoPlayerGame(t *testing.T, s *Server) (*Client, *Client) {
	t.Helper()
	host, guest := joinTwo(t, s)
	setOption(t, host, OptionCount, len(testWords), "")
	if err := host.Send(Message{Type: TypeStart}); err != nil {
		t.Fatalf("Send start: %v", err)
	}
	return host, guest
}

func TestMessageRoundTrip(t *testing.T) {
	want := Message{Type: TypeWordSpawn, Index: 3, Word: "hello", FallMs: 8000}
	data, err := encodeMessage(want)
	if err != nil {
		t.Fatalf("encodeMessage: %v", err)
	}
	if data[len(data)-1] != '\n' {
		t.Fatal("encoded message must end with a newline")
	}
	got, err := readMessage(bufio.NewScanner(strings.NewReader(string(data))))
	if err != nil {
		t.Fatalf("readMessage: %v", err)
	}
	if got.Type != want.Type || got.Index != want.Index || got.Word != want.Word || got.FallMs != want.FallMs {
		t.Errorf("round trip = %+v, want %+v", got, want)
	}
}

func TestJoinSuccess(t *testing.T) {
	s := startTestServer(t, time.Second)
	host := joinTest(t, s, "Host")
	guest := joinTest(t, s, "Guest")

	if host.PlayerID != hostPlayerID {
		t.Errorf("host id = %d, want %d", host.PlayerID, hostPlayerID)
	}
	if guest.PlayerID == hostPlayerID {
		t.Error("guest must not get the host id")
	}
	waitForLobbySize(t, host, 2)
}

func assertRejected(t *testing.T, err error, reason string) {
	t.Helper()
	var rejected *RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != reason {
		t.Fatalf("err = %v, want rejection %q", err, reason)
	}
}

// joinWithVersion sends a raw join with the given protocol version
func joinWithVersion(t *testing.T, s *Server, version int) Message {
	t.Helper()
	conn, err := net.DialTimeout("tcp", testAddr(s), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(2 * time.Second))
	if err := writeMessage(conn, Message{Type: TypeJoin, Version: version, Name: "Old"}); err != nil {
		t.Fatal(err)
	}
	reply, err := readMessage(bufio.NewScanner(conn))
	if err != nil {
		t.Fatalf("readMessage: %v", err)
	}
	return reply
}

func TestJoinVersionMismatch(t *testing.T) {
	s := startTestServer(t, time.Second)

	// Version 0 is what 2.0.0 clients send, since they had no version field
	for _, version := range []int{0, ProtocolVersion + 1} {
		reply := joinWithVersion(t, s, version)
		if reply.Type != TypeJoinRejected || reply.Reason != ReasonVersionMismatch {
			t.Errorf("version %d: reply = %+v", version, reply)
		}
	}

	var rejected RejectedError
	rejected.Reason = ReasonVersionMismatch
	if rejected.Error() != "Version mismatch" {
		t.Errorf("message = %q", rejected.Error())
	}
}

func TestJoinRoomFull(t *testing.T) {
	s := startTestServer(t, time.Second)
	for i := 0; i < MaxPlayers; i++ {
		joinTest(t, s, "P"+strconv.Itoa(i))
	}
	_, err := Connect(testAddr(s), "Late")
	assertRejected(t, err, ReasonRoomFull)
}

func TestJoinInGame(t *testing.T) {
	s := startTestServer(t, time.Second)
	host, _ := startTwoPlayerGame(t, s)
	waitFor(t, host, TypeGameStart)

	_, err := Connect(testAddr(s), "Late")
	assertRejected(t, err, ReasonInGame)
}

func TestConnectFailure(t *testing.T) {
	// Grab a free port and release it so nothing is listening there
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close()

	_, err = Connect(addr, "Guest")
	if !errors.Is(err, ErrCannotConnect) {
		t.Fatalf("err = %v, want ErrCannotConnect", err)
	}
}

func TestStartNeedsTwoPlayers(t *testing.T) {
	s := startTestServer(t, time.Second)
	host := joinTest(t, s, "Host")
	waitForLobbySize(t, host, 1)
	host.Send(Message{Type: TypeStart})

	// A second player can still join, proving the room stayed in the lobby
	joinTest(t, s, "Guest")
	waitForLobbySize(t, host, 2)
}

func TestNonHostCannotStart(t *testing.T) {
	s := startTestServer(t, time.Second)
	host := joinTest(t, s, "Host")
	guest := joinTest(t, s, "Guest")
	waitForLobbySize(t, host, 2)
	guest.Send(Message{Type: TypeStart})

	joinTest(t, s, "Third")
	_, seen := waitFor(t, host, TypeLobby)
	for _, msg := range seen {
		if msg.Type == TypeGameStart {
			t.Fatal("non-host started the game")
		}
	}
}

// TestFullGame claims the first word with each player, lets one expire,
// then checks the ranking
func TestFullGame(t *testing.T) {
	s := startTestServer(t, 300*time.Millisecond)
	host, guest := startTwoPlayerGame(t, s)

	first, _ := waitFor(t, guest, TypeWordSpawn)
	if first.Index != 0 || first.Word != testWords[0] {
		t.Fatalf("first spawn = %+v", first)
	}
	guest.Send(Message{Type: TypeClaim, Index: 0, Word: first.Word})
	claimed, _ := waitFor(t, host, TypeWordClaimed)
	if claimed.PlayerID != guest.PlayerID || claimed.Points != len(testWords[0]) {
		t.Fatalf("claimed = %+v", claimed)
	}

	second, _ := waitFor(t, host, TypeWordSpawn)
	host.Send(Message{Type: TypeClaim, Index: second.Index, Word: second.Word})
	waitFor(t, guest, TypeWordClaimed)

	// The remaining words fall unclaimed
	for want := 2; want < len(testWords); want++ {
		expired, _ := waitFor(t, guest, TypeWordExpired)
		if expired.Index != want {
			t.Errorf("expired index = %d, want %d", expired.Index, want)
		}
	}

	over, _ := waitFor(t, host, TypeGameOver)
	if len(over.Ranking) != 2 || over.Ranking[0].ID != guest.PlayerID {
		t.Fatalf("ranking = %+v", over.Ranking)
	}
	if over.Ranking[0].Score != 5 || over.Ranking[0].Words != 1 || over.Ranking[1].Score != 2 {
		t.Errorf("ranking = %+v", over.Ranking)
	}
	if len(over.Winners) != 1 || over.Winners[0] != guest.PlayerID {
		t.Errorf("winners = %v", over.Winners)
	}

	host.Send(Message{Type: TypeBackToLobby})
	lobby, _ := waitFor(t, guest, TypeLobby)
	if len(lobby.Players) != 2 {
		t.Errorf("lobby after game = %+v", lobby.Players)
	}
}

func TestDuplicateClaimScoresOnce(t *testing.T) {
	s := startTestServer(t, 300*time.Millisecond)
	host, guest := startTwoPlayerGame(t, s)

	spawn, _ := waitFor(t, host, TypeWordSpawn)
	claim := Message{Type: TypeClaim, Index: spawn.Index, Word: spawn.Word}
	host.Send(claim)
	guest.Send(claim)

	over, seen := waitFor(t, host, TypeGameOver)
	claims := 0
	for _, msg := range seen {
		if msg.Type == TypeWordClaimed && msg.Index == spawn.Index {
			claims++
		}
	}
	if claims != 1 {
		t.Errorf("word_claimed broadcasts = %d, want 1", claims)
	}
	total := 0
	for _, r := range over.Ranking {
		total += r.Score
	}
	if total != len(spawn.Word) {
		t.Errorf("total score = %d, want %d", total, len(spawn.Word))
	}
}

func TestInvalidClaimsIgnored(t *testing.T) {
	s := startTestServer(t, 200*time.Millisecond)
	host, guest := startTwoPlayerGame(t, s)

	spawn, _ := waitFor(t, host, TypeWordSpawn)
	guest.Send(Message{Type: TypeClaim, Index: spawn.Index, Word: "wrong"})
	guest.Send(Message{Type: TypeClaim, Index: spawn.Index + 1, Word: testWords[1]})

	expired, seen := waitFor(t, host, TypeWordExpired)
	if expired.Index != spawn.Index {
		t.Errorf("expired index = %d, want %d", expired.Index, spawn.Index)
	}
	for _, msg := range seen {
		if msg.Type == TypeWordClaimed {
			t.Fatalf("invalid claim was accepted: %+v", msg)
		}
	}
}

func TestTieHasMultipleWinners(t *testing.T) {
	s := startTestServer(t, 30*time.Millisecond)
	host, _ := startTwoPlayerGame(t, s)

	over, _ := waitFor(t, host, TypeGameOver)
	if len(over.Winners) != 2 {
		t.Errorf("winners = %v, want both players on a 0-0 tie", over.Winners)
	}
}

func TestDisconnectEndsGameEarly(t *testing.T) {
	s := startTestServer(t, 5*time.Second)
	host, guest := startTwoPlayerGame(t, s)

	spawn, _ := waitFor(t, guest, TypeWordSpawn)
	guest.Send(Message{Type: TypeClaim, Index: spawn.Index, Word: spawn.Word})
	waitFor(t, host, TypeWordClaimed)
	guest.Close()

	over, _ := waitFor(t, host, TypeGameOver)
	if len(over.Ranking) != 2 || over.Ranking[0].ID != guest.PlayerID {
		t.Errorf("departed player's score should stay in the ranking: %+v", over.Ranking)
	}
}

func TestPlayerLeavesLobby(t *testing.T) {
	s := startTestServer(t, time.Second)
	host := joinTest(t, s, "Host")
	guest := joinTest(t, s, "Guest")
	waitForLobbySize(t, host, 2)

	guest.Close()
	waitForLobbySize(t, host, 1)
}

func TestHostClosedNotifiesPlayers(t *testing.T) {
	s := startTestServer(t, time.Second)
	joinTest(t, s, "Host")
	guest := joinTest(t, s, "Guest")
	waitForLobbySize(t, guest, 2)

	s.Close()
	waitFor(t, guest, TypeHostClosed)
}

func TestHostDisconnectClosesRoom(t *testing.T) {
	s := startTestServer(t, time.Second)
	host := joinTest(t, s, "Host")
	guest := joinTest(t, s, "Guest")
	waitForLobbySize(t, guest, 2)

	host.Close()
	waitFor(t, guest, TypeHostClosed)
}

func TestClientSeesDisconnect(t *testing.T) {
	s := startTestServer(t, time.Second)
	host := joinTest(t, s, "Host")
	waitForLobbySize(t, host, 1)

	s.listener.Close()
	s.Close()
	waitFor(t, host, TypeDisconnected)
}

func TestSanitizeName(t *testing.T) {
	cases := map[string]string{
		"Alice":               "Alice",
		"  Bob  ":             "Bob",
		"":                    "Player",
		"\x1b[31mEvil\x07":    "[31mEvil",
		"AVeryLongPlayerName": "AVeryLongPla",
		"\x1b\x07":            "Player",
	}
	for in, want := range cases {
		if got := sanitizeName(in); got != want {
			t.Errorf("sanitizeName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBattleSubmit(t *testing.T) {
	b := NewBattle(2)
	b.Apply(Message{Type: TypeLobby, Players: []PlayerInfo{{ID: 1, Name: "Host", IsHost: true}, {ID: 2, Name: "Guest"}}, Mode: ModeWordDrop}, time.Now())

	// Typing before a word is falling is ignored
	b.AddChar('x')
	if b.Input != "" {
		t.Errorf("input outside falling phase = %q", b.Input)
	}

	b.Apply(Message{Type: TypeWordSpawn, Index: 1, Word: "go", FallMs: 1000}, time.Now())

	b.AddChar('g')
	b.AddChar('x')
	if _, ok := b.Submit(); ok {
		t.Error("wrong input must not produce a claim")
	}
	if b.Input != "" {
		t.Errorf("wrong input should clear, got %q", b.Input)
	}

	b.AddChar('g')
	b.AddChar('x')
	b.Backspace()
	b.AddChar('o')
	claim, ok := b.Submit()
	if !ok || claim.Type != TypeClaim || claim.Index != 1 || claim.Word != "go" {
		t.Errorf("claim = %+v, ok = %v", claim, ok)
	}

	b.Apply(Message{Type: TypeWordClaimed, Index: 1, PlayerID: 1, Points: 2}, time.Now())
	if b.Scores[1] != 2 || b.Feedback != "Host +2" || b.Phase != BattleBetween {
		t.Errorf("after claim: scores=%v feedback=%q phase=%v", b.Scores, b.Feedback, b.Phase)
	}
}

func TestBattleTiming(t *testing.T) {
	b := NewBattle(1)
	start := time.Now()

	b.Apply(Message{Type: TypeGameStart, TotalWords: 10, CountdownMs: 3000}, start)
	if got := b.CountdownSeconds(start); got != 3 {
		t.Errorf("countdown = %d, want 3", got)
	}
	if got := b.CountdownSeconds(start.Add(2500 * time.Millisecond)); got != 1 {
		t.Errorf("countdown = %d, want 1", got)
	}

	b.Apply(Message{Type: TypeWordSpawn, Word: "go", FallMs: 8000}, start)
	if got := b.FallProgress(start.Add(4 * time.Second)); got != 0.5 {
		t.Errorf("progress = %v, want 0.5", got)
	}
	if got := b.FallProgress(start.Add(20 * time.Second)); got != 1 {
		t.Errorf("progress = %v, want clamped to 1", got)
	}
}

func TestRenameInLobby(t *testing.T) {
	s := startTestServer(t, time.Second)
	host := joinTest(t, s, "Host")
	guest := joinTest(t, s, "Guest")
	waitForLobbySize(t, host, 2)

	guest.Send(Message{Type: TypeRename, Name: "\x1b[31mAVeryLongPlayerName"})
	lobby, _ := waitFor(t, host, TypeLobby)
	for _, p := range lobby.Players {
		if p.ID == guest.PlayerID && p.Name != "[31mAVeryLon" {
			t.Errorf("name = %q", p.Name)
		}
	}
}

func TestRenameIgnoredDuringGame(t *testing.T) {
	s := startTestServer(t, 30*time.Millisecond)
	host, guest := startTwoPlayerGame(t, s)
	waitFor(t, host, TypeGameStart)
	guest.Send(Message{Type: TypeRename, Name: "Sneaky"})

	over, _ := waitFor(t, host, TypeGameOver)
	for _, r := range over.Ranking {
		if r.Name == "Sneaky" {
			t.Errorf("rename applied during game: %+v", over.Ranking)
		}
	}
}
