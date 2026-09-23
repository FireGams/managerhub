package store

import "context"

// ListEnrollTokens returns recent enrollment tokens (hashed, for display).
func (s *Store) ListEnrollTokens(ctx context.Context, limit int) ([]EnrollToken, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.Pool.Query(ctx,
		`SELECT id, token_hash, label, expires_at, used_at, created_by, created_at
		 FROM enroll_tokens ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EnrollToken
	for rows.Next() {
		var t EnrollToken
		if err := rows.Scan(&t.ID, &t.TokenHash, &t.Label, &t.ExpiresAt, &t.UsedAt, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
