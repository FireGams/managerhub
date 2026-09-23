package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// CreateSchedule inserts a schedule definition.
func (s *Store) CreateSchedule(ctx context.Context, sc Schedule) (Schedule, error) {
	err := s.Pool.QueryRow(ctx, `
		INSERT INTO schedules (id, name, cron_expr, enabled, job_name, job_type, command, args, workdir, env, timeout_sec, node_id, selector)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING created_at`,
		sc.ID, sc.Name, sc.CronExpr, sc.Enabled, sc.JobName, sc.JobType, sc.Command,
		sc.Args, sc.WorkDir, marshalJSON(sc.Env), sc.TimeoutSec, sc.NodeID, marshalJSON(sc.Selector),
	).Scan(&sc.CreatedAt)
	return sc, err
}

// ListSchedules returns every schedule.
func (s *Store) ListSchedules(ctx context.Context) ([]Schedule, error) {
	rows, err := s.Pool.Query(ctx, schedSelect+` ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Schedule
	for rows.Next() {
		sc, err := scanSched(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

// ListEnabledSchedules returns active schedules for the cron loop.
func (s *Store) ListEnabledSchedules(ctx context.Context) ([]Schedule, error) {
	rows, err := s.Pool.Query(ctx, schedSelect+` WHERE enabled ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Schedule
	for rows.Next() {
		sc, err := scanSched(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

// SetScheduleEnabled toggles a schedule.
func (s *Store) SetScheduleEnabled(ctx context.Context, id string, enabled bool) error {
	_, err := s.Pool.Exec(ctx, `UPDATE schedules SET enabled=$2 WHERE id=$1`, id, enabled)
	return err
}

// MarkScheduleRun records a fired execution.
func (s *Store) MarkScheduleRun(ctx context.Context, schedID, jobID string) error {
	_, err := s.Pool.Exec(ctx,
		`INSERT INTO schedule_runs (schedule_id, job_id) VALUES ($1,$2)`, schedID, jobID)
	if err != nil {
		return err
	}
	_, err = s.Pool.Exec(ctx, `UPDATE schedules SET last_run_at=now() WHERE id=$1`, schedID)
	return err
}

const schedSelect = `SELECT id, name, cron_expr, enabled, job_name, job_type, command, args, workdir, env,
	timeout_sec, COALESCE(node_id::text,''), selector, last_run_at, created_at FROM schedules`

func scanSched(r rowScanner) (Schedule, error) {
	var sc Schedule
	var nodeID string
	var env, sel []byte
	err := r.Scan(&sc.ID, &sc.Name, &sc.CronExpr, &sc.Enabled, &sc.JobName, &sc.JobType,
		&sc.Command, &sc.Args, &sc.WorkDir, &env, &sc.TimeoutSec, &nodeID, &sel,
		&sc.LastRunAt, &sc.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return sc, ErrNotFound
	}
	if err != nil {
		return sc, err
	}
	if nodeID != "" {
		sc.NodeID = &nodeID
	}
	sc.Env = map[string]string{}
	sc.Selector = map[string]any{}
	_ = unmarshal(env, &sc.Env)
	_ = unmarshal(sel, &sc.Selector)
	return sc, nil
}
