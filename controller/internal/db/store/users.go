package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ErrNotFound is returned when a record does not exist.
var ErrNotFound = errors.New("store: not found")

// CreateUser inserts a user and returns the stored record.
func (s *Store) CreateUser(ctx context.Context, username, passwordHash, role string) (User, error) {
	u := User{ID: uuid.NewString(), Username: username, PasswordHash: passwordHash, Role: role}
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO users (id, username, password_hash, role)
		 VALUES ($1,$2,$3,$4) RETURNING created_at`,
		u.ID, u.Username, u.PasswordHash, u.Role,
	).Scan(&u.CreatedAt)
	return u, err
}

// GetUserByUsername loads a user by its unique name.
func (s *Store) GetUserByUsername(ctx context.Context, username string) (User, error) {
	var u User
	err := s.Pool.QueryRow(ctx,
		`SELECT id, username, password_hash, role, created_at FROM users WHERE username=$1`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

// CountUsers returns the number of registered accounts.
func (s *Store) CountUsers(ctx context.Context) (int, error) {
	var n int
	err := s.Pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&n)
	return n, err
}
