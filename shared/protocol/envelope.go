// Package protocol defines the versioned Controller <-> Agent wire format.
package protocol

import "encoding/json"

// CurrentVersion is the highest protocol version understood by this build.
const CurrentVersion = 1

// Envelope is the universal wrapper for every WebSocket frame exchanged
// between controller and agent. All payloads are JSON objects.
type Envelope struct {
	Version int             `json:"v"`
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	TS      int64           `json:"ts"`
	NodeID  string          `json:"node_id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// NewEnvelope builds an envelope with the current protocol version.
func NewEnvelope(typ, id string, ts int64, nodeID string, payload any) (Envelope, error) {
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return Envelope{}, err
		}
		raw = b
	}
	return Envelope{
		Version: CurrentVersion,
		Type:    typ,
		ID:      id,
		TS:      ts,
		NodeID:  nodeID,
		Payload: raw,
	}, nil
}

// Decode unmarshals the payload into dst.
func (e Envelope) Decode(dst any) error {
	if len(e.Payload) == 0 {
		return nil
	}
	return json.Unmarshal(e.Payload, dst)
}

// Encode serialises the envelope.
func (e Envelope) Encode() ([]byte, error) { return json.Marshal(e) }

// Parse decodes a raw frame into an Envelope and validates the version.
func Parse(data []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return Envelope{}, err
	}
	if env.Version < 1 || env.Version > CurrentVersion {
		return Envelope{}, ErrUnsupportedVersion
	}
	return env, nil
}
