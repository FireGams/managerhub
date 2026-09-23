// Package identity persists the agent node identity across restarts.
package identity

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// State is the on-disk identity of an agent.
type State struct {
	NodeID    string `json:"node_id"`
	NodeToken string `json:"node_token"`
}

// Load reads the state file; returns ErrNotEnrolled if absent.
func Load(path string) (State, error) {
	var st State
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return st, ErrNotEnrolled
	}
	if err != nil {
		return st, err
	}
	if err := json.Unmarshal(b, &st); err != nil {
		return st, err
	}
	if st.NodeID == "" || st.NodeToken == "" {
		return st, ErrNotEnrolled
	}
	return st, nil
}

// Save writes the state file with restrictive permissions.
func Save(path string, st State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

// ErrNotEnrolled signals that the agent has no identity yet.
var ErrNotEnrolled = errors.New("identity: not enrolled")
