package store

import "context"

// AppendAudit writes an audit entry.
func (s *Store) AppendAudit(ctx context.Context, actor, action, target, detail string) error {
	_, err := s.Pool.Exec(ctx,
		`INSERT INTO audit_log (actor, action, target, detail) VALUES ($1,$2,$3,$4)`,
		actor, action, target, detail)
	return err
}

// ListAudit returns recent audit entries.
func (s *Store) ListAudit(ctx context.Context, limit int) ([]AuditEntry, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := s.Pool.Query(ctx,
		`SELECT id, ts, actor, action, target, detail FROM audit_log ORDER BY ts DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AuditEntry
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.ID, &e.TS, &e.Actor, &e.Action, &e.Target, &e.Detail); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// AppendNodeEvent records a node lifecycle event (connect, disconnect, ...).
func (s *Store) AppendNodeEvent(ctx context.Context, nodeID, kind, detail string) error {
	_, err := s.Pool.Exec(ctx,
		`INSERT INTO node_events (node_id, kind, detail) VALUES ($1,$2,$3)`, nodeID, kind, detail)
	return err
}
