package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// CreateJob inserts a queued job.
func (s *Store) CreateJob(ctx context.Context, j Job) (Job, error) {
	err := s.Pool.QueryRow(ctx, `
		INSERT INTO jobs (id, name, type, node_id, command, args, workdir, env, timeout_sec, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'queued')
		RETURNING created_at`,
		j.ID, j.Name, j.Type, j.NodeID, j.Command, j.Args, j.WorkDir,
		marshalJSON(j.Env), j.TimeoutSec,
	).Scan(&j.CreatedAt)
	j.Status = "queued"
	return j, err
}

// GetJob loads one job.
func (s *Store) GetJob(ctx context.Context, id string) (Job, error) {
	return scanJob(s.Pool.QueryRow(ctx, jobSelect+` WHERE id=$1`, id))
}

// ListJobs returns recent jobs, newest first.
func (s *Store) ListJobs(ctx context.Context, limit int) ([]Job, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.Pool.Query(ctx, jobSelect+` ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Job
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// UpdateJobStatus transitions a job and optionally stores exit code / error.
func (s *Store) UpdateJobStatus(ctx context.Context, id, status string, exitCode *int, errMsg string) error {
	_, err := s.Pool.Exec(ctx, `
		UPDATE jobs SET status=$2, exit_code=$3, error=$4,
			started_at = CASE WHEN $2='running' AND started_at IS NULL THEN now() ELSE started_at END,
			finished_at = CASE WHEN $2 IN ('success','failed','cancelled','timeout') THEN now() ELSE finished_at END
		WHERE id=$1`, id, status, exitCode, errMsg)
	return err
}

// AppendJobOutput appends streamed output to the stdout or stderr column.
func (s *Store) AppendJobOutput(ctx context.Context, id, stream, data string) error {
	col := "stdout"
	if stream == "stderr" {
		col = "stderr"
	}
	_, err := s.Pool.Exec(ctx,
		`UPDATE jobs SET `+col+` = `+col+` || $2 WHERE id=$1`, id, data)
	return err
}

// PurgeJobLogs clears output of finished jobs older than the retention window.
func (s *Store) PurgeJobLogs(ctx context.Context, olderThan time.Duration) (int64, error) {
	tag, err := s.Pool.Exec(ctx, `
		UPDATE jobs SET stdout='', stderr=''
		WHERE finished_at IS NOT NULL AND finished_at < now() - $1::interval
		  AND (stdout <> '' OR stderr <> '')`, olderThan.String())
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

const jobSelect = `SELECT id, name, type, COALESCE(node_id::text,''), command, args, workdir, env,
	timeout_sec, status, exit_code, error, stdout, stderr, created_at, started_at, finished_at FROM jobs`

func scanJob(r rowScanner) (Job, error) {
	var j Job
	var nodeID string
	var env []byte
	err := r.Scan(&j.ID, &j.Name, &j.Type, &nodeID, &j.Command, &j.Args, &j.WorkDir, &env,
		&j.TimeoutSec, &j.Status, &j.ExitCode, &j.Error, &j.Stdout, &j.Stderr,
		&j.CreatedAt, &j.StartedAt, &j.FinishedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return j, ErrNotFound
	}
	if err != nil {
		return j, err
	}
	if nodeID != "" {
		j.NodeID = &nodeID
	}
	j.Env = map[string]string{}
	_ = unmarshal(env, &j.Env)
	return j, nil
}
