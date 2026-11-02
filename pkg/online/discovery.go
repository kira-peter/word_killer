package online

import (
	"encoding/json"
	"errors"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

const (
	// DiscoveryInterval is how often a browser asks the network for rooms
	DiscoveryInterval = 1 * time.Second
	// RoomTimeout drops a room that has not answered for this long
	RoomTimeout = 3 * time.Second

	maxDiscoveryPacket = 1024
)

// RoomInfo is a room found on the local network
type RoomInfo struct {
	ID         string
	HostName   string
	IP         string
	Port       int
	Players    int
	MaxPlayers int
	InGame     bool

	lastSeen time.Time
}

// Address returns the host:port to connect to
func (r RoomInfo) Address() string {
	return net.JoinHostPort(r.IP, strconv.Itoa(r.Port))
}

// Full reports whether the room has no free seat
func (r RoomInfo) Full() bool {
	return r.Players >= r.MaxPlayers
}

// Joinable reports whether a player may try to join the room
func (r RoomInfo) Joinable() bool {
	return !r.InGame && !r.Full()
}

// Browser finds rooms on the local network by broadcasting discover queries
// and collecting the room_info replies
type Browser struct {
	conn    *net.UDPConn
	port    int
	timeout time.Duration
	query   []byte

	mu    sync.Mutex
	rooms map[string]RoomInfo

	closeOnce sync.Once
	closed    chan struct{}
}

// NewBrowser starts browsing for rooms whose hosts listen on port
func NewBrowser(port int) (*Browser, error) {
	return newBrowser(port, DiscoveryInterval, RoomTimeout)
}

func newBrowser(port int, interval, timeout time.Duration) (*Browser, error) {
	// An ephemeral port lets any number of browsers run on one machine
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, err
	}
	query, err := encodeMessage(Message{Type: TypeDiscover, Version: ProtocolVersion})
	if err != nil {
		conn.Close()
		return nil, err
	}

	b := &Browser{
		conn:    conn,
		port:    port,
		timeout: timeout,
		query:   query,
		rooms:   make(map[string]RoomInfo),
		closed:  make(chan struct{}),
	}
	go b.receiveLoop()
	go b.queryLoop(interval)
	return b, nil
}

// Rooms returns the rooms currently known, in a stable order
func (b *Browser) Rooms() []RoomInfo {
	now := time.Now()
	b.mu.Lock()
	rooms := make([]RoomInfo, 0, len(b.rooms))
	for id, r := range b.rooms {
		if now.Sub(r.lastSeen) > b.timeout {
			delete(b.rooms, id)
			continue
		}
		rooms = append(rooms, r)
	}
	b.mu.Unlock()

	sort.Slice(rooms, func(i, j int) bool {
		if rooms[i].HostName != rooms[j].HostName {
			return rooms[i].HostName < rooms[j].HostName
		}
		return rooms[i].ID < rooms[j].ID
	})
	return rooms
}

// Close stops browsing
func (b *Browser) Close() {
	b.closeOnce.Do(func() {
		close(b.closed)
		b.conn.Close()
	})
}

func (b *Browser) queryLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		b.sendQuery()
		select {
		case <-ticker.C:
		case <-b.closed:
			return
		}
	}
}

// sendQuery broadcasts one discover query. Send errors are ignored: an
// unreachable target just means no room answers from there.
func (b *Browser) sendQuery() {
	for _, ip := range queryTargets() {
		b.conn.WriteToUDP(b.query, &net.UDPAddr{IP: ip, Port: b.port})
	}
}

func (b *Browser) receiveLoop() {
	buf := make([]byte, maxDiscoveryPacket+1)
	for {
		n, from, err := b.conn.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			// Any stray packet, such as an oversized one, must not stop browsing
			continue
		}
		msg, ok := decodeDiscovery(buf[:n])
		if !ok || msg.Type != TypeRoomInfo || msg.RoomID == "" || msg.Port <= 0 || msg.MaxPlayers <= 0 {
			continue
		}
		b.record(msg, from.IP)
	}
}

// record stores a reply, keeping one entry per room. The same room may
// answer through several interfaces; a LAN address beats loopback so the
// list shows the IP other players would use.
func (b *Browser) record(msg Message, ip net.IP) {
	room := RoomInfo{
		ID:         msg.RoomID,
		HostName:   sanitizeName(msg.Name),
		IP:         ip.String(),
		Port:       msg.Port,
		Players:    msg.PlayerNum,
		MaxPlayers: msg.MaxPlayers,
		InGame:     msg.InGame,
		lastSeen:   time.Now(),
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if old, ok := b.rooms[room.ID]; ok && ip.IsLoopback() && !net.ParseIP(old.IP).IsLoopback() {
		room.IP = old.IP
	}
	b.rooms[room.ID] = room
}

// decodeDiscovery parses one discovery packet, rejecting oversized,
// malformed and other-version packets
func decodeDiscovery(data []byte) (Message, bool) {
	var msg Message
	if len(data) > maxDiscoveryPacket {
		return msg, false
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return msg, false
	}
	return msg, msg.Version == ProtocolVersion
}

// queryTargets lists where discover queries go: the limited broadcast
// address, each interface's directed broadcast address (Windows sends a
// limited broadcast out of one interface only), and loopback so rooms on
// this machine are always found
func queryTargets() []net.IP {
	targets := []net.IP{net.IPv4bcast, net.IPv4(127, 0, 0, 1)}

	ifaces, err := net.Interfaces()
	if err != nil {
		return targets
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagBroadcast == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipNet.IP.To4()
			mask := ipNet.Mask
			if ip4 == nil || len(mask) != net.IPv4len {
				continue
			}
			bcast := make(net.IP, net.IPv4len)
			for i := range ip4 {
				bcast[i] = ip4[i] | ^mask[i]
			}
			targets = append(targets, bcast)
		}
	}
	return targets
}
