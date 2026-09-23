// Package terminal provides interactive PTY sessions over WebSocket.
package terminal

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/managerhub/managerhub/shared/protocol"
)

// Emitter sends protocol envelopes to the controller.
type Emitter interface {
	Send(env protocol.Envelope)
}

// Manager owns live PTY sessions.
type Manager struct {
	mu       sync.Mutex
	sessions map[string]*session
	emit     Emitter
	nodeID   string
	shell    string
}

type session struct {
	ptmx *os.File
	cmd  *exec.Cmd
}

// NewManager creates a terminal manager. shell may be empty for the platform default.
func NewManager(emit Emitter, nodeID, shell string) *Manager {
	return &Manager{sessions: make(map[string]*session), emit: emit, nodeID: nodeID, shell: shell}
}

// Open starts a PTY session.
func (m *Manager) Open(req protocol.TermOpen) error {
	shell := req.Shell
	if shell == "" {
		shell = m.shell
	}
	if shell == "" {
		if runtime.GOOS == "windows" {
			shell = "powershell.exe"
		} else {
			shell = "/bin/bash"
		}
	}
	cols, rows := req.Cols, req.Rows
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
	if err != nil {
		return err
	}
	s := &session{ptmx: ptmx, cmd: cmd}
	m.mu.Lock()
	m.sessions[req.SessionID] = s
	m.mu.Unlock()

	go m.readLoop(req.SessionID, s)
	return nil
}

func (m *Manager) readLoop(id string, s *session) {
	buf := make([]byte, 4096)
	for {
		n, err := s.ptmx.Read(buf)
		if n > 0 {
			data := base64.StdEncoding.EncodeToString(buf[:n])
			env, e := protocol.NewEnvelope(protocol.TypeTermOutput, id+"-out", time.Now().Unix(), m.nodeID,
				protocol.TermOutput{SessionID: id, Data: data})
			if e == nil {
				m.emit.Send(env)
			}
		}
		if err != nil {
			m.Close(id, "read error")
			return
		}
	}
}

// Input forwards keystrokes to the PTY.
// Data is base64-encoded from the UI.
func (m *Manager) Input(req protocol.TermInput) error {
	m.mu.Lock()
	s, ok := m.sessions[req.SessionID]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("terminal: unknown session %s", req.SessionID)
	}
	raw, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		// fallback: treat as raw text
		raw = []byte(req.Data)
	}
	n, err := s.ptmx.Write(raw)
	_ = n
	if err != nil {
		return fmt.Errorf("terminal: write: %w", err)
	}
	_ = n
	return nil
}

// Resize updates the PTY window size.
func (m *Manager) Resize(req protocol.TermResize) error {
	m.mu.Lock()
	s, ok := m.sessions[req.SessionID]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("terminal: unknown session")
	}
	return pty.Setsize(s.ptmx, &pty.Winsize{Cols: uint16(req.Cols), Rows: uint16(req.Rows)})
}

// Close terminates a session.
func (m *Manager) Close(id, reason string) {
	m.mu.Lock()
	s, ok := m.sessions[id]
	if ok {
		delete(m.sessions, id)
	}
	m.mu.Unlock()
	if !ok {
		return
	}
	_ = s.ptmx.Close()
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	env, _ := protocol.NewEnvelope(protocol.TypeTermClosed, id+"-closed", time.Now().Unix(), m.nodeID,
		protocol.TermClosed{SessionID: id, Reason: reason})
	if env.Type != "" {
		m.emit.Send(env)
	}
}

// CloseAll terminates every session (agent shutdown).
func (m *Manager) CloseAll() {
	m.mu.Lock()
	ids := make([]string, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	m.mu.Unlock()
	for _, id := range ids {
		m.Close(id, "agent shutdown")
	}
}
