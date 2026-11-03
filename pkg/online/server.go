package online

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
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
	DropDuration   time.Duration // how long a Word Drop word takes to fall
	StartCountdown time.Duration
	NextWordDelay  time.Duration // pause after a word or sentence is taken
	// WordSource returns n distinct words for a new game
	WordSource func(n int) []string
	// SentenceSource returns n distinct sentences for a new game
	SentenceSource func(n int) []string
}

type serverPhase int

const (
	phaseLobby serverPhase = iota
	phaseCountdown
	phasePlaying
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
	// timeLimit ends the game as the time limit would; tests use it to avoid
	// waiting out the real limit
	timeLimit bool
}

// roomStatus is what discovery replies report; run publishes it, and the
// UDP responder reads it
type roomStatus struct {
	hostName string
	players  int
	inGame   bool
}

// Server hosts one room. All game state is owned by the run goroutine,
// so claims and timers are handled strictly one at a time.
type Server struct {
	cfg      ServerConfig
	roomID   string
	listener net.Listener
	udp      *net.UDPConn
	events   chan serverEvent
	closeCh  chan struct{}
	done     chan struct{}
	once     sync.Once

	players  []*serverPlayer
	nextID   int
	mode     string
	settings Settings
	phase    serverPhase
	game     gameMode

	// timer drives the countdown and then the mode; limit ends the game
	// when the overall time limit is reached
	timer  *time.Timer
	timerC <-chan time.Time
	limit  *time.Timer
	limitC <-chan time.Time

	statusMu sync.Mutex
	status   roomStatus
}

// NewServer starts listening and returns a running room server
func NewServer(cfg ServerConfig) (*Server, error) {
	listener, udp, err := listenPair(cfg.Host, cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("cannot listen on port %d: %w", cfg.Port, err)
	}

	s := &Server{
		cfg:      cfg,
		roomID:   newRoomID(),
		listener: listener,
		udp:      udp,
		events:   make(chan serverEvent, eventBufferSize),
		closeCh:  make(chan struct{}),
		done:     make(chan struct{}),
		nextID:   hostPlayerID,
		mode:     ModeWordDrop,
		settings: DefaultSettings(ModeWordDrop),
	}

	go s.acceptLoop()
	go s.discoveryLoop()
	go s.run()
	return s, nil
}

// listenPair opens the TCP game port and the UDP discovery port with the same
// number, since players broadcast queries to the port they would connect to.
// With port 0 the system picks the TCP port, whose UDP twin may be taken, so a
// few picks are tried; a fixed port gets exactly one attempt.
func listenPair(host string, port int) (net.Listener, *net.UDPConn, error) {
	attempts := 1
	if port == 0 {
		attempts = 10
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		listener, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			return nil, nil, err
		}
		tcpPort := listener.Addr().(*net.TCPAddr).Port
		udp, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(host), Port: tcpPort})
		if err == nil {
			return listener, udp, nil
		}
		listener.Close()
		lastErr = err
	}
	return nil, nil, lastErr
}

// newRoomID returns a random identifier used to de-duplicate discovery replies
func newRoomID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
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
		case <-s.limitC:
			s.limitC = nil
			s.finishGame()
			s.publishStatus()
		case <-s.closeCh:
			s.shutdown()
			return
		}
	}
}

// discoveryLoop answers discover queries until the UDP socket is closed
func (s *Server) discoveryLoop() {
	buf := make([]byte, maxDiscoveryPacket+1)
	for {
		n, from, err := s.udp.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			// Any stray packet, such as an oversized one, must not stop discovery
			continue
		}
		msg, ok := decodeDiscovery(buf[:n])
		if !ok || msg.Type != TypeDiscover {
			continue
		}

		s.statusMu.Lock()
		status := s.status
		s.statusMu.Unlock()
		// Not discoverable until the host has joined its own room
		if status.players == 0 {
			continue
		}

		reply, err := encodeMessage(Message{
			Type:       TypeRoomInfo,
			Version:    ProtocolVersion,
			RoomID:     s.roomID,
			Name:       status.hostName,
			Port:       s.Port(),
			PlayerNum:  status.players,
			MaxPlayers: MaxPlayers,
			InGame:     status.inGame,
		})
		if err == nil {
			s.udp.WriteToUDP(reply, from)
		}
	}
}

// publishStatus records the room state for discovery replies
func (s *Server) publishStatus() {
	status := roomStatus{players: len(s.players), inGame: s.phase != phaseLobby}
	for _, p := range s.players {
		if p.id == hostPlayerID {
			status.hostName = p.name
		}
	}
	s.statusMu.Lock()
	s.status = status
	s.statusMu.Unlock()
}

func (s *Server) shutdown() {
	s.stopTimers()
	s.listener.Close()
	s.udp.Close()
	for _, p := range s.players {
		if p.connected {
			s.send(p, Message{Type: TypeHostClosed})
			p.conn.Close()
		}
	}
}

func (s *Server) handleEvent(ev serverEvent) {
	// Every event may change the players or phase that discovery reports
	defer s.publishStatus()

	if ev.timeLimit {
		s.finishGame()
		return
	}

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
	case TypeSetOption:
		if player.id == hostPlayerID && s.phase == phaseLobby {
			s.setOption(ev.msg.Option, ev.msg.Value, ev.msg.Mode)
		}
	case TypeKick:
		if player.id == hostPlayerID && s.phase == phaseLobby {
			s.kick(ev.msg.PlayerID)
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
		if s.phase == phasePlaying {
			s.game.onClaim(player, ev.msg)
		}
	case TypeTyping:
		// Relay what a player is typing so everyone can watch the race
		if s.phase == phasePlaying {
			s.broadcastExcept(player, Message{Type: TypeTyping, PlayerID: player.id, Word: limitTyping(ev.msg.Word)})
		}
	case TypeBackToLobby:
		if player.id == hostPlayerID && s.phase == phaseResult {
			s.backToLobby()
		}
	}
}

// setOption changes the mode or one of its settings. Values outside the
// preset steps are ignored; a new mode starts from its defaults.
func (s *Server) setOption(key string, value int, mode string) {
	if key == OptionMode {
		if !isKnownMode(mode) || mode == s.mode {
			return
		}
		s.mode = mode
		s.settings = DefaultSettings(mode)
		s.broadcastLobby()
		return
	}
	settings, ok := s.settings.Apply(s.mode, key, value)
	if !ok {
		return
	}
	s.settings = settings
	s.broadcastLobby()
}

// kick tells a player they were kicked, drops their connection and removes
// them. Their read loop then reports a disconnect for a connection that no
// longer belongs to anyone, which handleEvent ignores.
func (s *Server) kick(id int) {
	if id == hostPlayerID {
		return
	}
	for _, p := range s.players {
		if p.id == id {
			s.send(p, Message{Type: TypeKicked})
			p.connected = false
			p.conn.Close()
			s.removeDisconnected()
			s.broadcastLobby()
			return
		}
	}
}

func (s *Server) handleJoin(conn net.Conn, msg Message) {
	reject := func(reason string) {
		conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		writeMessage(conn, Message{Type: TypeJoinRejected, Reason: reason})
		conn.Close()
	}

	if msg.Type != TypeJoin || msg.Version != ProtocolVersion {
		reject(ReasonVersionMismatch)
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

	s.send(player, Message{Type: TypeJoinAccepted, PlayerID: player.id})
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
	case phaseCountdown, phasePlaying:
		// Keep the player so their score still appears in the ranking
		if s.connectedCount() < MinPlayers {
			s.finishGame()
		}
	}
}

// newGame builds the selected mode with fresh words or sentences, or
// returns nil when there is nothing to play with
func (s *Server) newGame() gameMode {
	switch s.mode {
	case ModeSentenceRace:
		sentences := s.cfg.SentenceSource(s.settings.Count)
		if len(sentences) == 0 {
			return nil
		}
		return newSentenceRace(s, sentences, s.cfg.NextWordDelay)
	case ModeWordPool:
		if len(s.cfg.WordSource(1)) == 0 {
			return nil
		}
		return newWordPool(s, s.cfg.WordSource, s.settings.SpawnRate, s.settings.AllowSteal)
	default:
		words := s.cfg.WordSource(s.settings.Count)
		if len(words) == 0 {
			return nil
		}
		return newWordDrop(s, words, s.cfg.DropDuration, s.cfg.NextWordDelay)
	}
}

func (s *Server) startGame() {
	if len(s.players) < MinPlayers {
		return
	}
	game := s.newGame()
	if game == nil {
		// Tell the host instead of silently staying in the lobby
		for _, p := range s.players {
			if p.id == hostPlayerID {
				s.send(p, Message{Type: TypeStartFailed, Reason: "No " + itemNoun(s.mode) + " available"})
			}
		}
		return
	}

	s.game = game
	for _, p := range s.players {
		p.score = 0
		p.words = 0
	}

	s.phase = phaseCountdown
	s.broadcast(Message{
		Type:        TypeGameStart,
		Mode:        s.mode,
		Settings:    &s.settings,
		TotalWords:  game.totalItems(),
		CountdownMs: int(s.cfg.StartCountdown.Milliseconds()),
	})
	s.setModeTimer(s.cfg.StartCountdown)
}

func (s *Server) handleTimer() {
	defer s.publishStatus()

	switch s.phase {
	case phaseCountdown:
		// The time limit runs from the first word, not from the countdown
		s.phase = phasePlaying
		s.limit = time.NewTimer(time.Duration(s.settings.TimeLimit) * time.Second)
		s.limitC = s.limit.C
		s.game.begin()
	case phasePlaying:
		s.game.onTimer()
	}
}

// award adds a mode's points; negative values take them back after a steal
func (s *Server) award(p *serverPlayer, points, items int) {
	p.score += points
	p.words += items
}

func (s *Server) finishGame() {
	if s.phase != phaseCountdown && s.phase != phasePlaying {
		return
	}
	s.stopTimers()
	s.phase = phaseResult
	s.game = nil

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
	s.broadcast(Message{Type: TypeLobby, Players: players, Mode: s.mode, Settings: &s.settings})
}

func (s *Server) broadcast(msg Message) {
	for _, p := range s.players {
		if p.connected {
			s.send(p, msg)
		}
	}
}

// broadcastExcept sends to everyone but one player, who already knows
func (s *Server) broadcastExcept(skip *serverPlayer, msg Message) {
	for _, p := range s.players {
		if p.connected && p != skip {
			s.send(p, msg)
		}
	}
}

// maxTypingLength bounds a relayed input; the longest sentence is far shorter
const maxTypingLength = 200

// limitTyping keeps relayed input printable and bounded, since it is passed
// on to every other player's terminal
func limitTyping(text string) string {
	text = printableASCII(text)
	if len(text) > maxTypingLength {
		text = text[:maxTypingLength]
	}
	return text
}

// send writes to a player; a failed write closes the connection,
// which surfaces as a disconnect event from its read loop
func (s *Server) send(p *serverPlayer, msg Message) {
	p.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err := writeMessage(p.conn, msg); err != nil {
		p.conn.Close()
	}
}

// setModeTimer (re)arms the single timer shared by the countdown and the mode
func (s *Server) setModeTimer(d time.Duration) {
	if s.timer != nil {
		s.timer.Stop()
	}
	s.timer = time.NewTimer(d)
	s.timerC = s.timer.C
}

func (s *Server) stopTimers() {
	if s.timer != nil {
		s.timer.Stop()
	}
	if s.limit != nil {
		s.limit.Stop()
	}
	s.timerC = nil
	s.limitC = nil
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
