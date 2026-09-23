package api

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// broadcaster sends events to one UI WebSocket session.
type broadcaster struct {
	ws        *websocket.Conn
	mu        sync.Mutex
	terminals map[string]string // session_id → node_id
	jobs      map[string]bool   // job_id → watching
	closed    bool
}

func newBroadcaster(ws *websocket.Conn) *broadcaster {
	return &broadcaster{
		ws:        ws,
		terminals: make(map[string]string),
		jobs:      make(map[string]bool),
	}
}

func (b *broadcaster) send(event string, payload any) {
	msg := map[string]any{"event": event, "data": payload, "ts": time.Now().Unix()}
	raw, _ := json.Marshal(msg)
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.closed {
		_ = b.ws.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_ = b.ws.WriteMessage(websocket.TextMessage, raw)
	}
}

func (b *broadcaster) trackTerminal(sessID, nodeID string) {
	b.mu.Lock()
	b.terminals[sessID] = nodeID
	b.mu.Unlock()
}

func (b *broadcaster) watchJob(jobID string) {
	b.mu.Lock()
	b.jobs[jobID] = true
	b.mu.Unlock()
}

func (b *broadcaster) isWatchingJob(jobID string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.jobs[jobID]
}

func (b *broadcaster) close() {
	b.mu.Lock()
	b.closed = true
	b.mu.Unlock()
}

func (s *Server) addUIClient(bc *broadcaster) {
	s.uiMu.Lock()
	if s.uiClients == nil {
		s.uiClients = make(map[*broadcaster]bool)
	}
	s.uiClients[bc] = true
	s.uiMu.Unlock()
}

func (s *Server) removeUIClient(bc *broadcaster) {
	s.uiMu.Lock()
	delete(s.uiClients, bc)
	s.uiMu.Unlock()
	bc.close()
}

func (s *Server) broadcastUI(event string, payload any) {
	s.uiMu.RLock()
	clients := make([]*broadcaster, 0, len(s.uiClients))
	for bc := range s.uiClients {
		clients = append(clients, bc)
	}
	s.uiMu.RUnlock()
	for _, bc := range clients {
		bc.send(event, payload)
	}
}

func (s *Server) broadcastJobEvent(jobID, event string, payload any) {
	s.uiMu.RLock()
	clients := make([]*broadcaster, 0, len(s.uiClients))
	for bc := range s.uiClients {
		if bc.isWatchingJob(jobID) {
			clients = append(clients, bc)
		}
	}
	s.uiMu.RUnlock()
	for _, bc := range clients {
		bc.send(event, payload)
	}
}
