package online

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	// DialTimeout limits connecting to a host and waiting for its join reply
	DialTimeout = 3 * time.Second

	incomingBufferSize = 256
)

// ErrCannotConnect means the host could not be reached
var ErrCannotConnect = errors.New("cannot connect to host")

// RejectedError means the host refused the join request
type RejectedError struct {
	Reason string
}

func (e *RejectedError) Error() string {
	switch e.Reason {
	case ReasonVersionMismatch:
		return "Version mismatch"
	case ReasonRoomFull:
		return "Room is full"
	case ReasonInGame:
		return "Game in progress"
	}
	return "Join rejected"
}

// Client is a connection to a room server
type Client struct {
	PlayerID int

	conn     net.Conn
	incoming chan Message
	closed   chan struct{}
	once     sync.Once
}

// Connect dials a room server and completes the join handshake
func Connect(addr, name string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, DialTimeout)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCannotConnect, err)
	}

	conn.SetDeadline(time.Now().Add(DialTimeout))
	if err := writeMessage(conn, Message{Type: TypeJoin, Version: ProtocolVersion, Name: name}); err != nil {
		conn.Close()
		return nil, fmt.Errorf("%w: %v", ErrCannotConnect, err)
	}

	scanner := bufio.NewScanner(conn)
	reply, err := readMessage(scanner)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("%w: %v", ErrCannotConnect, err)
	}
	if reply.Type != TypeJoinAccepted {
		conn.Close()
		return nil, &RejectedError{Reason: reply.Reason}
	}
	conn.SetDeadline(time.Time{})

	c := &Client{
		PlayerID: reply.PlayerID,
		conn:     conn,
		incoming: make(chan Message, incomingBufferSize),
		closed:   make(chan struct{}),
	}
	go c.readLoop(scanner)
	return c, nil
}

func (c *Client) readLoop(scanner *bufio.Scanner) {
	for {
		msg, err := readMessage(scanner)
		if err != nil {
			msg = Message{Type: TypeDisconnected}
		}
		select {
		case c.incoming <- msg:
		case <-c.closed:
			return
		}
		if err != nil {
			return
		}
	}
}

// Poll returns all pending messages without blocking
func (c *Client) Poll() []Message {
	var msgs []Message
	for {
		select {
		case msg := <-c.incoming:
			msgs = append(msgs, msg)
		default:
			return msgs
		}
	}
}

// Send sends a message to the server; a dead connection is reported via Poll
func (c *Client) Send(msg Message) error {
	c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	return writeMessage(c.conn, msg)
}

// Close closes the connection
func (c *Client) Close() {
	c.once.Do(func() {
		close(c.closed)
		c.conn.Close()
	})
}
