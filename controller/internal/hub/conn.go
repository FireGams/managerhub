// Package hub maintains the live WebSocket sessions of every agent.
package hub

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/managerhub/managerhub/shared/protocol"
)

// Conn is one authenticated agent WebSocket session.
type Conn struct {
	NodeID string
	ws     *websocket.Conn
	send   chan []byte
	dedup  *protocol.Deduper
	mu     sync.Mutex
	closed bool
}

func NewConn(nodeID string, ws *websocket.Conn) *Conn {
	return &Conn{
		NodeID: nodeID,
		ws:     ws,
		send:   make(chan []byte, 256),
		dedup:  protocol.NewDeduper(8192),
	}
}

// Send queues an envelope for delivery; drops if the buffer is full.
func (c *Conn) Send(env protocol.Envelope) error {
	raw, err := env.Encode()
	if err != nil {
		return err
	}
	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if closed {
		return ErrSendBufferFull
	}
	select {
	case c.send <- raw:
		return nil
	default:
		return ErrSendBufferFull
	}
}

// Duplicate reports whether the message id was already processed.
func (c *Conn) Duplicate(id string) bool { return c.dedup.Seen(id) }

func (c *Conn) writeLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case raw, ok := <-c.send:
			if !ok {
				return
			}
			if c.IsClosed() {
				return
			}
			_ = c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.ws.WriteMessage(websocket.TextMessage, raw); err != nil {
				return
			}
		case <-ticker.C:
			if c.IsClosed() {
				return
			}
			_ = c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Conn) close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	ws := c.ws
	c.mu.Unlock()
	_ = ws.Close()
}

// IsClosed reports whether the connection is closed.
func (c *Conn) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}
