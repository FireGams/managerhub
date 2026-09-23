package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// UpsertNode registers or refreshes a node identity (enrollment / hello).
func (s *Store) UpsertNode(ctx context.Context, n Node, tokenHash string) (Node, error) {
	err := s.Pool.QueryRow(ctx, `
		INSERT INTO nodes (id, name, hostname, os, arch, ip, agent_version, token_hash, status, tags, capabilities)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'online',$9,$10)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name, hostname=EXCLUDED.hostname, os=EXCLUDED.os,
			arch=EXCLUDED.arch, ip=EXCLUDED.ip, agent_version=EXCLUDED.agent_version,
			status='online', last_seen_at=now(),
			token_hash=CASE WHEN EXCLUDED.token_hash='' THEN nodes.token_hash ELSE EXCLUDED.token_hash END,
			tags=EXCLUDED.tags,
			capabilities=EXCLUDED.capabilities, updated_at=now()
		RETURNING created_at, updated_at, last_seen_at`,
		n.ID, n.Name, n.Hostname, n.OS, n.Arch, n.IP, n.AgentVersion, tokenHash,
		n.Tags, marshalJSON(n.Capabilities),
	).Scan(&n.CreatedAt, &n.UpdatedAt, &n.LastSeenAt)
	n.Status = "online"
	return n, err
}

// TouchNode records a heartbeat / reconnect.
func (s *Store) TouchNode(ctx context.Context, id string) error {
	_, err := s.Pool.Exec(ctx,
		`UPDATE nodes SET last_seen_at=now(), status='online', updated_at=now() WHERE id=$1`, id)
	return err
}

// SetNodeStatus flips a node between online and offline.
func (s *Store) SetNodeStatus(ctx context.Context, id, status string) error {
	_, err := s.Pool.Exec(ctx,
		`UPDATE nodes SET status=$2, updated_at=now() WHERE id=$1`, id, status)
	return err
}

// GetNode loads one node by id.
func (s *Store) GetNode(ctx context.Context, id string) (Node, error) {
	return scanNode(s.Pool.QueryRow(ctx, nodeSelect+` WHERE id=$1`, id))
}

// GetNodeByToken loads a node authenticating with its token hash.
func (s *Store) GetNodeByToken(ctx context.Context, tokenHash string) (Node, error) {
	return scanNode(s.Pool.QueryRow(ctx, nodeSelect+` WHERE token_hash=$1`, tokenHash))
}

// ListNodes returns every registered node, newest activity first.
func (s *Store) ListNodes(ctx context.Context) ([]Node, error) {
	rows, err := s.Pool.Query(ctx, nodeSelect+` ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Node
	for rows.Next() {
		n, err := scanNode(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

const nodeSelect = `SELECT id, name, hostname, os, arch, ip, agent_version, status,
	COALESCE(last_seen_at, NULL), tags, capabilities, created_at, updated_at
	FROM nodes`

func scanNode(r rowScanner) (Node, error) {
	var n Node
	var caps []byte
	err := r.Scan(&n.ID, &n.Name, &n.Hostname, &n.OS, &n.Arch, &n.IP, &n.AgentVersion,
		&n.Status, &n.LastSeenAt, &n.Tags, &caps, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return n, ErrNotFound
	}
	if err != nil {
		return n, err
	}
	n.Capabilities = map[string]string{}
	_ = unmarshal(caps, &n.Capabilities)
	return n, nil
}

func unmarshal(b []byte, v any) error {
	if len(b) == 0 {
		return nil
	}
	return jsonUnmarshal(b, v)
}

