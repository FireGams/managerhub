package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// EnrollToken is a one-time credential used to register a new agent.
type EnrollToken struct {
	ID        string     `json:"id"`
	TokenHash string     `json:"-"`
	Label     string     `json:"label"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedBy string     `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
}

// CreateEnrollToken stores a hashed one-time enrollment token.
func (s *Store) CreateEnrollToken(ctx context.Context, tokenHash, label, createdBy string, ttl time.Duration) (EnrollToken, error) {
	t := EnrollToken{
		ID: uuid.NewString(), TokenHash: tokenHash, Label: label,
		ExpiresAt: time.Now().Add(ttl), CreatedBy: createdBy,
	}
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO enroll_tokens (id, token_hash, label, expires_at, created_by)
		 VALUES ($1,$2,$3,$4,$5) RETURNING created_at`,
		t.ID, t.TokenHash, t.Label, t.ExpiresAt, t.CreatedBy,
	).Scan(&t.CreatedAt)
	return t, err
}

// ConsumeEnrollToken atomically marks a token as used if still valid.
func (s *Store) ConsumeEnrollToken(ctx context.Context, tokenHash string) (EnrollToken, error) {
	var t EnrollToken
	err := s.Pool.QueryRow(ctx, `
		UPDATE enroll_tokens SET used_at=now()
		WHERE token_hash=$1 AND used_at IS NULL AND expires_at > now()
		RETURNING id, token_hash, label, expires_at, used_at, created_by, created_at`,
		tokenHash,
	).Scan(&t.ID, &t.TokenHash, &t.Label, &t.ExpiresAt, &t.UsedAt, &t.CreatedBy, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return t, ErrNotFound
	}
	return t, err
}
