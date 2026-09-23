// Package db provides PostgreSQL connectivity and schema migrations.
package db

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Migrate applies every pending "Up" migration in lexical order.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("migrate: init: %w", err)
	}

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("migrate: read: %w", err)
	}
	var ups []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") && !strings.HasSuffix(e.Name(), ".down.sql") {
			ups = append(ups, e.Name())
		}
	}
	sort.Strings(ups)

	for _, name := range ups {
		var exists bool
		if err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, name,
		).Scan(&exists); err != nil {
			return fmt.Errorf("migrate: check %s: %w", name, err)
		}
		if exists {
			continue
		}
		sql, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("migrate: apply %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`, name,
		); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}

// RollbackDown applies the matching .down.sql of the latest applied migration.
func RollbackDown(ctx context.Context, pool *pgxpool.Pool) error {
	var version string
	err := pool.QueryRow(ctx,
		`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`,
	).Scan(&version)
	if err != nil {
		return fmt.Errorf("migrate: nothing to roll back: %w", err)
	}
	down := strings.TrimSuffix(version, ".sql") + ".down.sql"
	sql, err := migrationFS.ReadFile("migrations/" + down)
	if err != nil {
		return fmt.Errorf("migrate: missing %s: %w", down, err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, string(sql)); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM schema_migrations WHERE version=$1`, version); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
