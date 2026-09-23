package store

import (
	"context"
	"time"
)

// InsertMetrics stores one resource sample.
func (s *Store) InsertMetrics(ctx context.Context, m Metrics) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO node_metrics
			(node_id, ts, cpu_pct, cpu_cores, load1, ram_used, ram_total, disk_used, disk_total, uptime_sec, gpus)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		m.NodeID, m.TS, m.CPUPercent, m.CPUCores, m.Load1, m.RAMUsed, m.RAMTotal,
		m.DiskUsed, m.DiskTotal, m.UptimeSec, []byte(orEmpty(m.GPUs)),
	)
	return err
}

// LatestMetrics returns the most recent sample for a node, if any.
func (s *Store) LatestMetrics(ctx context.Context, nodeID string) (*Metrics, error) {
	var m Metrics
	err := s.Pool.QueryRow(ctx, `
		SELECT node_id, ts, cpu_pct, cpu_cores, load1, ram_used, ram_total, disk_used, disk_total, uptime_sec, gpus
		FROM node_metrics WHERE node_id=$1 ORDER BY ts DESC LIMIT 1`, nodeID,
	).Scan(&m.NodeID, &m.TS, &m.CPUPercent, &m.CPUCores, &m.Load1, &m.RAMUsed, &m.RAMTotal,
		&m.DiskUsed, &m.DiskTotal, &m.UptimeSec, &m.GPUs)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// PurgeMetrics deletes samples older than the retention window.
func (s *Store) PurgeMetrics(ctx context.Context, olderThan time.Duration) (int64, error) {
	tag, err := s.Pool.Exec(ctx,
		`DELETE FROM node_metrics WHERE ts < now() - $1::interval`,
		olderThan.String())
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func orEmpty(s string) string {
	if s == "" {
		return "[]"
	}
	return s
}
