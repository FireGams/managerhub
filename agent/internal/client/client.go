// Package client maintains the outbound WebSocket session to the controller.
package client

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/managerhub/managerhub/shared/protocol"
)

// Handler processes an inbound envelope from the controller.
type Handler func(env protocol.Envelope)

// Client is a reconnecting WebSocket client.
type Client struct {
	rawURL string
	token  string
	log    *slog.Logger
	dedup  *protocol.Deduper
	send   chan protocol.Envelope
	mu     sync.Mutex
	ws     *websocket.Conn

	OnMessage Handler
	OnConnect func()
	OnClose   func()
}

// New creates a client targeting controllerURL with the given node token.
func New(controllerURL, token string, log *slog.Logger) (*Client, error) {
	u, err := url.Parse(controllerURL)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	}
	u.Path = strings.TrimSuffix(u.Path, "/") + "/api/v1/agent/ws"
	return &Client{
		rawURL: u.String(),
		token:  token,
		log:    log,
		dedup:  protocol.NewDeduper(8192),
		send:   make(chan protocol.Envelope, 256),
	}, nil
}

// Send queues an envelope for the controller.
func (c *Client) Send(env protocol.Envelope) {
	select {
	case c.send <- env:
	default:
		c.log.Warn("client: send buffer full, dropping", "type", env.Type)
	}
}

// Run blocks, reconnecting forever until ctx is cancelled.
func (c *Client) Run(ctx context.Context) {
	backoff := time.Second
	for {
		if ctx.Err() != nil {
			return
		}
		err := c.session(ctx)
		if ctx.Err() != nil {
			return
		}
		c.log.Warn("client: session ended", "err", err, "retry_in", backoff)
		if c.OnClose != nil {
			c.OnClose()
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff *= 2
		if backoff > 30*time.Second {
			backoff = 30 * time.Second
		}
	}
}

func (c *Client) session(ctx context.Context) error {
	hdr := http.Header{"Authorization": []string{"Bearer " + c.token}}
	ws, _, err := websocket.DefaultDialer.DialContext(ctx, c.rawURL, hdr)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.ws = ws
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		c.ws = nil
		c.mu.Unlock()
		_ = ws.Close()
	}()

	if c.OnConnect != nil {
		c.OnConnect()
	}

	writeErr := make(chan error, 1)
	go c.writeLoop(ctx, ws, writeErr)

	readErr := make(chan error, 1)
	go func() { readErr <- c.readLoop(ws) }()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-readErr:
		return err
	case err := <-writeErr:
		return err
	}
}

func (c *Client) writeLoop(ctx context.Context, ws *websocket.Conn, out chan<- error) {
	ping := time.NewTicker(30 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-ctx.Done():
			out <- ctx.Err()
			return
		case <-ping.C:
			ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				out <- err
				return
			}
		case env := <-c.send:
			raw, err := env.Encode()
			if err != nil {
				continue
			}
			ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ws.WriteMessage(websocket.TextMessage, raw); err != nil {
				out <- err
				return
			}
		}
	}
}

func (c *Client) readLoop(ws *websocket.Conn) error {
	ws.SetReadLimit(1 << 20)
	_ = ws.SetReadDeadline(time.Now().Add(120 * time.Second))
	ws.SetPongHandler(func(string) error {
		_ = ws.SetReadDeadline(time.Now().Add(120 * time.Second))
		return nil
	})
	for {
		_, raw, err := ws.ReadMessage()
		if err != nil {
			return err
		}
		_ = ws.SetReadDeadline(time.Now().Add(120 * time.Second))
		env, err := protocol.Parse(raw)
		if err != nil {
			c.log.Warn("client: bad envelope", "err", err)
			continue
		}
		if c.dedup.Seen(env.ID) {
			continue
		}
		if c.OnMessage != nil {
			c.OnMessage(env)
		}
	}
}

// NewMsgID returns a fresh unique message id.
func NewMsgID() string { return uuid.NewString() }
