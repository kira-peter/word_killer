package online

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	// MaxPlayers is the room capacity
	MaxPlayers = 4
	// MinPlayers is the minimum number of players to start a game
	MinPlayers = 2

	hostPlayerID    = 1
	maxNameLength   = 12
	joinTimeout     = 5 * time.Second
	writeTimeout    = 2 * time.Second
	eventBufferSize = 64
)

// ServerConfig configures a room server
type ServerConfig struct {
	// Host is the listen address; empty means all interfaces
	Host           string
	Port           int
	DropDuration   time.Duration
	WordCount      int
	StartCountdown time.Duration
	NextWordDelay  time.Duration
	// WordSource returns n distinct words for a new game
	WordSource func(n int) []string
}

type serverPhase int

const (
	phaseLobby serverPhase = iota
	phaseCountdown
	phaseFalling
	phaseBetween
	phaseResult
)

type serverPlayer struct {
	id        int
	name      string
	conn      net.Conn
	connected bool
	score     int
	words     int
}

// serverEvent is a connection event delivered to the main loop
type serverEvent struct {
	conn net.Conn
	msg  Message
	gone bool
}

// Server hosts one room. All game state is owned by the run goroutine,
// so claims and timers are handled strictly one at a time.
type Server struct {
	cfg      ServerConfig
	room     string
	listener net.Listener
	events   chan serverEvent
	closeCh  chan struct{}
	done     chan struct{}
	once     sync.Once

	players []*serverPlayer
	nextID  int
	mode    string
	phase   serverPhase
	words   []string
	current int
	timer   *time.Timer
	timerC  <-chan time.Time
}

// NewServer starts listening and returns a running room server
func NewServer(cfg ServerConfig) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("cannot listen on port %d: %w", cfg.Port, err)
	}

	s := &Server{
		cfg:      cfg,
		room:     fmt.Sprintf("%04d", rand.Intn(10000)),
		listener: listener,
		events:   make(chan serverEvent, eventBufferSize),
		closeCh:  make(chan struct{}),
		done:     make(chan struct{}),
		nextID:   hostPlayerID,
		mode:     ModeWordDrop,
	}

	go s.acceptLoop()
	go s.run()
	return s, nil
}

// Room returns the 4-digit room number
func (s *Server) Room() string {
	return s.room
}

// Port returns the port actually being listened on
func (s *Server) Port() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}

// Close notifies all players and shuts the room down
func (s *Server) Close() {
	s.signalClose()
	<-s.done
}

func (s *Server) signalClose() {
	s.once.Do(func() { close(s.closeCh) })
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.readLoop(conn)
	}
}

// readLoop forwards every message on a connection to the main loop
func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()

	// Connections must identify themselves promptly
	conn.SetReadDeadline(time.Now().Add(joinTimeout))
	scanner := bufio.NewScanner(conn)
	for {
		msg, err := readMessage(scanner)
		if err != nil {
			s.post(serverEvent{conn: conn, gone: true})
			return
		}
		conn.SetReadDeadline(time.Time{})
		s.post(serverEvent{conn: conn, msg: msg})
	}
}

func (s *Server) post(ev serverEvent) {
	select {
	case s.events <- ev:
	case <-s.done:
	}
}

func (s *Server) run() {
	defer close(s.done)
	for {
		select {
		case ev := <-s.events:
			s.handleEvent(ev)
		case <-s.timerC:
			s.timerC = nil
			s.handleTimer()
		case <-s.closeCh:
			s.shutdown()
			return
		}
	}
}

func (s *Server) shutdown() {
	s.stopTimer()
	s.listener.Close()
	for _, p := range s.players {
		if p.connected {
			s.send(p, Message{Type: TypeHostClosed})
			p.conn.Close()
		}
	}
}

func (s *Server) handleEvent(ev serverEvent) {
	player := s.playerByConn(ev.conn)

	if ev.gone {
		if player != nil {
			s.handleDisconnect(player)
		}
		return
	}

	if player == nil {
		s.handleJoin(ev.conn, ev.msg)
		return
	}

	switch ev.msg.Type {
	case TypeSelectMode:
		if player.id == hostPlayerID && s.phase == phaseLobby && isKnownMode(ev.msg.Mode) {
			s.mode = ev.msg.Mode
			s.broadcastLobby()
		}
	case TypeRename:
		if s.phase == phaseLobby {
			player.name = sanitizeName(ev.msg.Name)
			s.broadcastLobby()
		}
	case TypeStart:
		if player.id == hostPlayerID && s.phase == phaseLobby {
			s.startGame()
		}
	case TypeClaim:
		s.handleClaim(player, ev.msg)
	case TypeBackToLobby:
		if player.id == hostPlayerID && s.phase == phaseResult {
			s.backToLobby()
		}
	}
}

func (s *Server) handleJoin(conn net.Conn, msg Message) {
	reject := func(reason string) {
		conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		writeMessage(conn, Message{Type: TypeJoinRejected, Reason: reason})
		conn.Close()
	}

	if msg.Type != TypeJoin || msg.Room != s.room {
		reject(ReasonBadRoom)
		return
	}
	if s.phase != phaseLobby {
		reject(ReasonInGame)
		return
	}
	if len(s.players) >= MaxPlayers {
		reject(ReasonRoomFull)
		return
	}

	player := &serverPlayer{
		id:        s.nextID,
		name:      sanitizeName(msg.Name),
		conn:      conn,
		connected: true,
	}
	s.nextID++
	s.players = append(s.players, player)

	s.send(player, Message{Type: TypeJoinAccepted, PlayerID: player.id, Room: s.room})
	s.broadcastLobby()
}

func (s *Server) handleDisconnect(player *serverPlayer) {
	player.connected = false

	// The host's own connection going away means the room is gone
	if player.id == hostPlayerID {
		s.signalClose()
		return
	}

	switch s.phase {
	case phaseLobby:
		s.removeDisconnected()
		s.broadcastLobby()
	case phaseCountdown, phaseFalling, phaseBetween:
		// Keep the player so their score still appears in the ranking
		if s.connectedCount() < MinPlayers {
			s.finishGame()
		}
	}
}

func (s *Server) startGame() {
	if len(s.players) < MinPlayers {
		return
	}
	words := s.cfg.WordSource(s.cfg.WordCount)
	if len(words) == 0 {
		return
	}

	s.words = words
	s.current = 0
	for _, p := range s.players {
		p.score = 0
		p.words = 0
	}

	s.phase = phaseCountdown
	s.broadcast(Message{
		Type:        TypeGameStart,
		Mode:        s.mode,
		TotalWords:  len(words),
		CountdownMs: int(s.cfg.StartCountdown.Milliseconds()),
	})
	s.setTimer(s.cfg.StartCountdown)
}

func (s *Server) handleTimer() {
	switch s.phase {
	case phaseCountdown:
		s.spawnWord()
	case phaseFalling:
		s.broadcast(Message{Type: TypeWordExpired, Index: s.current})
		s.endWord()
	case phaseBetween:
		s.current++
		if s.current >= len(s.words) {
			s.finishGame()
		} else {
			s.spawnWord()
		}
	}
}

func (s *Server) spawnWord() {
	s.phase = phaseFalling
	s.broadcast(Message{
		Type:   TypeWordSpawn,
		Index:  s.current,
		Word:   s.words[s.current],
		FallMs: int(s.cfg.DropDuration.Milliseconds()),
	})
	s.setTimer(s.cfg.DropDuration)
}

// endWord starts the pause before the next word
func (s *Server) endWord() {
	s.phase = phaseBetween
	s.setTimer(s.cfg.NextWordDelay)
}

// handleClaim awards the current word to the first valid claim to arrive.
// Later or stale claims fail the phase/index check and are ignored.
func (s *Server) handleClaim(player *serverPlayer, msg Message) {
	if s.phase != phaseFalling || msg.Index != s.current || msg.Word != s.words[s.current] {
		return
	}

	points := len([]rune(msg.Word))
	player.score += points
	player.words++

	s.broadcast(Message{
		Type:     TypeWordClaimed,
		Index:    s.current,
		PlayerID: player.id,
		Points:   points,
	})
	s.endWord()
}

func (s *Server) finishGame() {
	s.stopTimer()
	s.phase = phaseResult

	ranking := make([]RankEntry, 0, len(s.players))
	for _, p := range s.players {
		ranking = append(ranking, RankEntry{ID: p.id, Name: p.name, Score: p.score, Words: p.words})
	}
	sort.SliceStable(ranking, func(i, j int) bool {
		return ranking[i].Score > ranking[j].Score
	})

	var winners []int
	for _, r := range ranking {
		if r.Score == ranking[0].Score {
			winners = append(winners, r.ID)
		}
	}

	s.broadcast(Message{Type: TypeGameOver, Ranking: ranking, Winners: winners})
}

func (s *Server) backToLobby() {
	s.phase = phaseLobby
	s.removeDisconnected()
	s.broadcastLobby()
}

func (s *Server) removeDisconnected() {
	kept := s.players[:0]
	for _, p := range s.players {
		if p.connected {
			kept = append(kept, p)
		}
	}
	s.players = kept
}

func (s *Server) broadcastLobby() {
	players := make([]PlayerInfo, 0, len(s.players))
	for _, p := range s.players {
		players = append(players, PlayerInfo{ID: p.id, Name: p.name, IsHost: p.id == hostPlayerID})
	}
	s.broadcast(Message{Type: TypeLobby, Players: players, Mode: s.mode})
}

func (s *Server) broadcast(msg Message) {
	for _, p := range s.players {
		if p.connected {
			s.send(p, msg)
		}
	}
}

// send writes to a player; a failed write closes the connection,
// which surfaces as a disconnect event from its read loop
func (s *Server) send(p *serverPlayer, msg Message) {
	p.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err := writeMessage(p.conn, msg); err != nil {
		p.conn.Close()
	}
}

func (s *Server) setTimer(d time.Duration) {
	s.stopTimer()
	s.timer = time.NewTimer(d)
	s.timerC = s.timer.C
}

func (s *Server) stopTimer() {
	if s.timer != nil {
		s.timer.Stop()
	}
	s.timerC = nil
}

func (s *Server) playerByConn(conn net.Conn) *serverPlayer {
	for _, p := range s.players {
		if p.conn == conn {
			return p
		}
	}
	return nil
}

func (s *Server) connectedCount() int {
	count := 0
	for _, p := range s.players {
		if p.connected {
			count++
		}
	}
	return count
}

func isKnownMode(mode string) bool {
	for _, m := range Modes {
		if m == mode {
			return true
		}
	}
	return false
}

// sanitizeName strips control characters so a remote name cannot inject
// terminal escape sequences, and limits its length
func sanitizeName(name string) string {
	name = strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, name))
	if name == "" {
		return "Player"
	}
	if runes := []rune(name); len(runes) > maxNameLength {
		return string(runes[:maxNameLength])
	}
	return name
}
