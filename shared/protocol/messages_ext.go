package protocol

// TermOpen starts an interactive shell session.
type TermOpen struct {
	SessionID string `json:"session_id"`
	Shell     string `json:"shell,omitempty"`
	Cols      int    `json:"cols"`
	Rows      int    `json:"rows"`
}

// TermInput carries keystrokes to the PTY.
type TermInput struct {
	SessionID string `json:"session_id"`
	Data      string `json:"data"`
}

// TermResize updates PTY window size.
type TermResize struct {
	SessionID string `json:"session_id"`
	Cols      int    `json:"cols"`
	Rows      int    `json:"rows"`
}

// TermOutput carries PTY output to the browser.
type TermOutput struct {
	SessionID string `json:"session_id"`
	Data      string `json:"data"`
}

// TermClose / TermClosed terminate a session.
type TermClose struct {
	SessionID string `json:"session_id"`
}
type TermClosed struct {
	SessionID string `json:"session_id"`
	Reason    string `json:"reason,omitempty"`
}

// ServiceInfo describes a host service (systemd / Windows / launchd).
type ServiceInfo struct {
	Name   string `json:"name"`
	State  string `json:"state"`  // active|inactive|failed|unknown
	Sub    string `json:"sub,omitempty"`
	Enabled bool  `json:"enabled"`
}

// SvcListResult returns the service inventory.
type SvcListResult struct {
	ReqID    string        `json:"req_id"`
	Services []ServiceInfo `json:"services"`
	Error    string        `json:"error,omitempty"`
}

// SvcAction requests start/stop/restart/status on one service.
type SvcAction struct {
	ReqID   string `json:"req_id"`
	Name    string `json:"name"`
	Action  string `json:"action"` // start|stop|restart|status
}

// SvcActionResult reports the outcome.
type SvcActionResult struct {
	ReqID  string `json:"req_id"`
	Name   string `json:"name"`
	Action string `json:"action"`
	OK     bool   `json:"ok"`
	Error  string `json:"error,omitempty"`
}
