package hub

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/managerhub/managerhub/shared/protocol"
)

// ErrSendBufferFull is returned when a slow agent cannot keep up.
var ErrSendBufferFull = errors.New("hub: send buffer full")

// ErrNodeOffline is returned when no live session exists for a node.
var ErrNodeOffline = errors.New("hub: node offline")

// Hub tracks live agent connections and routes envelopes to them.
type Hub struct {
	mu    sync.RWMutex
	conns map[string]*Conn
	log   *slog.Logger
}

// New creates an empty hub.
func New(log *slog.Logger) *Hub {
	return &Hub{conns: make(map[string]*Conn), log: log}
}

// Register attaches a new connection for a node, replacing any stale one.
func (h *Hub) Register(c *Conn) {
	h.mu.Lock()
	if old, ok := h.conns[c.NodeID]; ok {
		old.close()
	}
	h.conns[c.NodeID] = c
	h.mu.Unlock()
	go c.writeLoop()
}

// Unregister removes a connection if it is still the active one.
func (h *Hub) Unregister(c *Conn) {
	h.mu.Lock()
	if cur, ok := h.conns[c.NodeID]; ok && cur == c {
		delete(h.conns, c.NodeID)
	}
	h.mu.Unlock()
	c.close()
}

// Send delivers an envelope to a node.
func (h *Hub) Send(nodeID string, env protocol.Envelope) error {
	h.mu.RLock()
	c, ok := h.conns[nodeID]
	h.mu.RUnlock()
	if !ok {
		return ErrNodeOffline
	}
	return c.Send(env)
}

// Online reports whether a node has a live session.
func (h *Hub) Online(nodeID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.conns[nodeID]
	return ok
}

// OnlineIDs returns the ids of every connected node.
func (h *Hub) OnlineIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]string, 0, len(h.conns))
	for id := range h.conns {
		out = append(out, id)
	}
	return out
}

// NotifyStatus is a convenience used on connect / disconnect transitions.
type NotifyStatus func(nodeID, status string, at time.Time)
