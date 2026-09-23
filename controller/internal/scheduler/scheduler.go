// Package scheduler runs controller-owned cron schedules and dispatches jobs.
package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/managerhub/managerhub/controller/internal/db/store"
	"github.com/managerhub/managerhub/shared/protocol"
	"github.com/robfig/cron/v3"
)

// Dispatcher creates and sends a job to a node.
type Dispatcher interface {
	CreateAndDispatch(j store.Job) (store.Job, error)
}

// Scheduler owns the cron engine and the dispatch loop.
type Scheduler struct {
	cron *cron.Cron
	st   *store.Store
	disp Dispatcher
	log  *slog.Logger
}

// New builds a scheduler.
func New(st *store.Store, disp Dispatcher, log *slog.Logger) *Scheduler {
	return &Scheduler{cron: cron.New(), st: st, disp: disp, log: log}
}

// Start loads enabled schedules and starts the cron engine.
func (s *Scheduler) Start(ctx context.Context) {
	s.cron.Start()
	s.reload(ctx)
	go func() {
		t := time.NewTicker(time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				s.cron.Stop()
				return
			case <-t.C:
				s.reload(ctx)
			}
		}
	}()
}

// Reload rebuilds cron entries from the database.
func (s *Scheduler) Reload(ctx context.Context) { s.reload(ctx) }

func (s *Scheduler) reload(ctx context.Context) {
	list, err := s.st.ListEnabledSchedules(ctx)
	if err != nil {
		s.log.Error("scheduler: list", "err", err)
		return
	}
	for _, e := range s.cron.Entries() {
		s.cron.Remove(e.ID)
	}
	for _, sc := range list {
		sc := sc
		if _, err := s.cron.AddFunc(sc.CronExpr, func() { s.fire(sc) }); err != nil {
			s.log.Error("scheduler: bad cron", "name", sc.Name, "err", err)
		}
	}
}

func (s *Scheduler) fire(sc store.Schedule) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	j := store.Job{
		Name: sc.JobName, Type: sc.JobType, Command: sc.Command, Args: sc.Args,
		WorkDir: sc.WorkDir, Env: sc.Env, TimeoutSec: sc.TimeoutSec, NodeID: sc.NodeID,
		Status: protocol.JobQueued,
	}
	out, err := s.disp.CreateAndDispatch(j)
	if err != nil {
		s.log.Error("scheduler: dispatch", "schedule", sc.Name, "err", err)
		return
	}
	_ = s.st.MarkScheduleRun(ctx, sc.ID, out.ID)
	s.log.Info("scheduler: fired", "schedule", sc.Name, "job", out.ID)
}
