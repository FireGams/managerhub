// Command controller runs the ManagerHub central server.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/managerhub/managerhub/controller/internal/api"
	"github.com/managerhub/managerhub/controller/internal/config"
	"github.com/managerhub/managerhub/controller/internal/db"
	"github.com/managerhub/managerhub/controller/internal/db/store"
	"github.com/managerhub/managerhub/controller/internal/hub"
	"github.com/managerhub/managerhub/controller/internal/scheduler"
)

var version = "dev"

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := run(log); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer st.Close()

	if err := db.Migrate(ctx, st.Pool); err != nil {
		return err
	}

	h := hub.New(log)
	srv := &api.Server{Cfg: cfg, Store: st, Hub: h, Log: log}
	sched := scheduler.New(st, srv, log)
	sched.Start(ctx)
	api.SetSchedulerReload(func() { sched.Reload(context.Background()) })

	go purgeLoop(ctx, st, cfg, log)

	httpSrv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		log.Info("controller listening", "addr", cfg.HTTPAddr, "version", version)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http", "err", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return httpSrv.Shutdown(shutdownCtx)
}

func purgeLoop(ctx context.Context, st *store.Store, cfg config.Config, log *slog.Logger) {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if n, err := st.PurgeMetrics(ctx, cfg.MetricsRetention); err == nil && n > 0 {
				log.Info("purged metrics", "rows", n)
			}
			if n, err := st.PurgeJobLogs(ctx, cfg.JobLogRetention); err == nil && n > 0 {
				log.Info("purged job logs", "rows", n)
			}
		}
	}
}
