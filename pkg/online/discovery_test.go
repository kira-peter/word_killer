package online

import (
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

// startBrowser browses fast so tests do not wait whole seconds
func startBrowser(t *testing.T, port int) *Browser {
	t.Helper()
	b, err := newBrowser(port, 50*time.Millisecond, 300*time.Millisecond)
	if err != nil {
		t.Fatalf("newBrowser: %v", err)
	}
	t.Cleanup(b.Close)
	return b
}

// waitForRoom polls the browser until cond matches a room
func waitForRoom(t *testing.T, b *Browser, what string, cond func(RoomInfo) bool) RoomInfo {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		for _, r := range b.Rooms() {
			if cond(r) {
				return r
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s, rooms = %+v", what, b.Rooms())
	return RoomInfo{}
}

func TestDiscoverLocalRoom(t *testing.T) {
	s := startTestServer(t, time.Second)
	b := startBrowser(t, s.Port())

	// A room is hidden until its host has joined
	time.Sleep(200 * time.Millisecond)
	if rooms := b.Rooms(); len(rooms) != 0 {
		t.Fatalf("room listed before the host joined: %+v", rooms)
	}

	host := joinTest(t, s, "Hosty")
	room := waitForRoom(t, b, "host's room", func(r RoomInfo) bool { return r.HostName == "Hosty" })
	if room.Port != s.Port() || room.Players != 1 || room.MaxPlayers != MaxPlayers || room.InGame || !room.Joinable() {
		t.Errorf("room = %+v", room)
	}

	// Joining through the advertised address works
	guest, err := Connect(room.Address(), "Guesty")
	if err != nil {
		t.Fatalf("Connect(%s): %v", room.Address(), err)
	}
	t.Cleanup(guest.Close)
	waitForRoom(t, b, "two players", func(r RoomInfo) bool { return r.Players == 2 })

	// The listing follows the game state
	waitForLobbySize(t, host, 2)
	host.Send(Message{Type: TypeStart})
	room = waitForRoom(t, b, "in-game room", func(r RoomInfo) bool { return r.InGame })
	if room.Joinable() {
		t.Error("a room in game must not be joinable")
	}

	s.Close()
	deadline := time.Now().Add(2 * time.Second)
	for len(b.Rooms()) > 0 {
		if time.Now().After(deadline) {
			t.Fatalf("closed room still listed: %+v", b.Rooms())
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestFullRoomNotJoinable(t *testing.T) {
	s := startTestServer(t, time.Second)
	b := startBrowser(t, s.Port())
	for i := 0; i < MaxPlayers; i++ {
		joinTest(t, s, "P"+strconv.Itoa(i))
	}
	room := waitForRoom(t, b, "full room", func(r RoomInfo) bool { return r.Players == MaxPlayers })
	if !room.Full() || room.Joinable() {
		t.Errorf("room = %+v", room)
	}
}

func TestTwoBrowsersOnOneMachine(t *testing.T) {
	s := startTestServer(t, time.Second)
	joinTest(t, s, "Hosty")

	first := startBrowser(t, s.Port())
	second := startBrowser(t, s.Port())
	waitForRoom(t, first, "room in first browser", func(r RoomInfo) bool { return r.HostName == "Hosty" })
	waitForRoom(t, second, "room in second browser", func(r RoomInfo) bool { return r.HostName == "Hosty" })
}

func TestBrowserIgnoresForeignPackets(t *testing.T) {
	b := startBrowser(t, 9) // nothing answers on the discard port
	// The browser listens on 0.0.0.0, which Windows will not deliver to; aim at loopback
	target := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: b.conn.LocalAddr().(*net.UDPAddr).Port}
	sender, err := net.DialUDP("udp4", nil, target)
	if err != nil {
		t.Fatal(err)
	}
	defer sender.Close()

	valid := Message{Type: TypeRoomInfo, Version: ProtocolVersion, RoomID: "r1", Name: "Real", Port: 7788, PlayerNum: 1, MaxPlayers: 4}
	oversized := valid
	oversized.RoomID = "r2"
	oversized.Name = string(make([]byte, maxDiscoveryPacket))
	otherVersion := valid
	otherVersion.RoomID = "r3"
	otherVersion.Version = ProtocolVersion + 1
	noPort := valid
	noPort.RoomID = "r4"
	noPort.Port = 0

	packets := [][]byte{[]byte("not json"), []byte(`{"type":"room_info"}`)}
	for _, msg := range []Message{oversized, otherVersion, noPort, valid} {
		data, err := encodeMessage(msg)
		if err != nil {
			t.Fatal(err)
		}
		packets = append(packets, data)
	}
	// Keep sending like a real host would, since rooms expire quickly in tests
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		for {
			for _, p := range packets {
				sender.Write(p)
			}
			select {
			case <-stop:
				return
			case <-time.After(50 * time.Millisecond):
			}
		}
	}()

	room := waitForRoom(t, b, "valid room", func(r RoomInfo) bool { return r.ID == "r1" })
	if room.IP != "127.0.0.1" || room.HostName != "Real" {
		t.Errorf("room = %+v", room)
	}
	time.Sleep(100 * time.Millisecond)
	if rooms := b.Rooms(); len(rooms) != 1 {
		t.Errorf("only the valid packet should be listed: %+v", rooms)
	}
}

func TestHostSurvivesOversizedPacket(t *testing.T) {
	s := startTestServer(t, time.Second)
	joinTest(t, s, "Hosty")

	// An oversized datagram errors the read on Windows; discovery must keep going
	junk, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: s.Port()})
	if err != nil {
		t.Fatal(err)
	}
	defer junk.Close()
	junk.Write(make([]byte, 4*maxDiscoveryPacket))
	time.Sleep(50 * time.Millisecond)

	b := startBrowser(t, s.Port())
	waitForRoom(t, b, "room after junk", func(r RoomInfo) bool { return r.HostName == "Hosty" })
}

func TestBrowserPrefersLANAddress(t *testing.T) {
	b := &Browser{rooms: make(map[string]RoomInfo), timeout: time.Second}
	msg := Message{Type: TypeRoomInfo, Version: ProtocolVersion, RoomID: "r1", Name: "Host", Port: 7788, PlayerNum: 2, MaxPlayers: 4}

	b.record(msg, net.ParseIP("192.168.1.20"))
	msg.PlayerNum = 3
	b.record(msg, net.ParseIP("127.0.0.1"))

	rooms := b.Rooms()
	if len(rooms) != 1 || rooms[0].IP != "192.168.1.20" || rooms[0].Players != 3 {
		t.Errorf("rooms = %+v", rooms)
	}
}

func TestBrowserSanitizesHostName(t *testing.T) {
	b := &Browser{rooms: make(map[string]RoomInfo), timeout: time.Second}
	b.record(Message{RoomID: "r1", Name: "\x1b[31mEvil", Port: 1, MaxPlayers: 4}, net.ParseIP("10.0.0.2"))
	if name := b.Rooms()[0].HostName; name != "[31mEvil" {
		t.Errorf("name = %q", name)
	}
}

func TestRoomsSortedStably(t *testing.T) {
	b := &Browser{rooms: make(map[string]RoomInfo), timeout: time.Second}
	for _, r := range []struct{ id, name string }{{"c", "Bob"}, {"b", "Alice"}, {"a", "Bob"}} {
		b.record(Message{RoomID: r.id, Name: r.name, Port: 1, MaxPlayers: 4}, net.ParseIP("10.0.0.2"))
	}
	var got []string
	for _, r := range b.Rooms() {
		got = append(got, r.ID)
	}
	if want := []string{"b", "a", "c"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("order = %v, want %v", got, want)
	}
}

func TestQueryTargetsIncludeBroadcastAndLoopback(t *testing.T) {
	targets := queryTargets()
	var hasBcast, hasLoopback bool
	for _, ip := range targets {
		hasBcast = hasBcast || ip.Equal(net.IPv4bcast)
		hasLoopback = hasLoopback || ip.Equal(net.IPv4(127, 0, 0, 1))
	}
	if !hasBcast || !hasLoopback {
		t.Errorf("targets = %v", targets)
	}
}
