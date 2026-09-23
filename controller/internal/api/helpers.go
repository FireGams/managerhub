package api

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/managerhub/managerhub/controller/internal/db/store"
	"github.com/managerhub/managerhub/shared/protocol"
)

func nowUnix() int64 { return time.Now().Unix() }

// PickNode selects the least-loaded online node matching a simple selector.
// Supported keys: os (string), cpu_cores (number), min_ram_free (bytes).
func (s *Server) PickNode(ctx context.Context, selector map[string]any) (string, error) {
	nodes, err := s.Store.ListNodes(ctx)
	if err != nil {
		return "", err
	}
	bestID := ""
	bestLoad := 1e18
	for _, n := range nodes {
		if !s.Hub.Online(n.ID) {
			continue
		}
		if wantOS, ok := selector["os"].(string); ok && wantOS != "" && n.OS != wantOS {
			continue
		}
		m, _ := s.Store.LatestMetrics(ctx, n.ID)
		var load float64
		var free uint64
		var cores float64
		if m != nil {
			load = m.CPUPercent
			cores = float64(m.CPUCores)
			if m.RAMTotal > m.RAMUsed {
				free = m.RAMTotal - m.RAMUsed
			}
		}
		if wantCores, ok := selector["cpu_cores"].(float64); ok && cores < wantCores {
			continue
		}
		if wantFree, ok := selector["min_ram_free"].(float64); ok && free < uint64(wantFree) {
			continue
		}
		if load < bestLoad {
			bestLoad = load
			bestID = n.ID
		}
	}
	if bestID == "" {
		return "", errors.New("no compatible online node")
	}
	return bestID, nil
}

// CreateAndDispatch persists a job and pushes it to the target node.
func (s *Server) CreateAndDispatch(j store.Job) (store.Job, error) {
	if j.NodeID == nil {
		id, err := s.PickNode(context.Background(), nil)
		if err != nil {
			return j, err
		}
		j.NodeID = &id
	}
	j.ID = uuid.NewString()
	j.Status = protocol.JobQueued
	out, err := s.Store.CreateJob(context.Background(), j)
	if err != nil {
		return out, err
	}
	s.assignJob(out)
	return out, nil
}
