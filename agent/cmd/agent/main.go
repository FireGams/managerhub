// Command agent runs the ManagerHub node agent.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/managerhub/managerhub/agent/internal/client"
	"github.com/managerhub/managerhub/agent/internal/config"
	"github.com/managerhub/managerhub/agent/internal/identity"
	"github.com/managerhub/managerhub/agent/internal/jobs"
	"github.com/managerhub/managerhub/agent/internal/runner"
	"github.com/managerhub/managerhub/agent/internal/services"
	"github.com/managerhub/managerhub/agent/internal/sysinfo"
	"github.com/managerhub/managerhub/agent/internal/terminal"
	"github.com/managerhub/managerhub/shared/protocol"
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

	state, err := ensureIdentity(cfg, log)
	if err != nil {
		return err
	}

	cli, err := client.New(cfg.ControllerURL, state.NodeToken, log)
	if err != nil {
		return err
	}

	a := &agent{
		cfg:   cfg,
		state: state,
		cli:   cli,
		log:   log,
		jobs:  jobs.NewRunner(cli, state.NodeID),
		term:  terminal.NewManager(cli, state.NodeID, cfg.Shell),
		svc:   services.New(),
		coll:  sysinfo.NewCollector(),
	}
	a.det = runner.NewDetector(a.svc)

	cli.OnMessage = a.handle
	cli.OnConnect = func() { a.sendHello() }

	go a.heartbeatLoop(ctx)
	go a.metricsLoop(ctx)

	log.Info("agent started", "node_id", state.NodeID, "name", cfg.Name,
		"version", version, "os", runtime.GOOS, "arch", runtime.GOARCH)
	cli.Run(ctx)
	a.term.CloseAll()
	return nil
}

func ensureIdentity(cfg config.Config, log *slog.Logger) (identity.State, error) {
	st, err := identity.Load(cfg.StatePath)
	if err == nil {
		return st, nil
	}
	if err != identity.ErrNotEnrolled {
		return st, err
	}
	if cfg.EnrollToken == "" {
		return st, fmt.Errorf("agent: not enrolled and MH_ENROLL_TOKEN is empty")
	}
	st, err = enroll(cfg, log)
	if err != nil {
		return st, err
	}
	if err := identity.Save(cfg.StatePath, st); err != nil {
		return st, err
	}
	log.Info("enrolled", "node_id", st.NodeID)
	return st, nil
}

func enroll(cfg config.Config, log *slog.Logger) (identity.State, error) {
	var st identity.State
	hostname, goos, arch := sysinfo.HostInfo()
	body, _ := json.Marshal(map[string]any{
		"enroll_token":  cfg.EnrollToken,
		"name":          cfg.Name,
		"hostname":      hostname,
		"os":            goos,
		"arch":          arch,
		"agent_version": version,
		"tags":          []string{goos},
		"capabilities":  map[string]string{"shell": "true"},
	})
	url := cfg.ControllerURL + "/api/v1/agent/enroll"
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return st, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		return st, fmt.Errorf("enroll: status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		return st, err
	}
	_ = log
	return st, nil
}

type agent struct {
	cfg   config.Config
	state identity.State
	cli   *client.Client
	log   *slog.Logger
	jobs  *jobs.Runner
	term  *terminal.Manager
	svc   interface {
		List() ([]protocol.ServiceInfo, error)
		Action(name, action string) error
	}
	det  *runner.Detector
	coll *sysinfo.Collector
}

func (a *agent) sendHello() {
	hostname, goos, arch := sysinfo.HostInfo()
	env, err := protocol.NewEnvelope(protocol.TypeHello, uuid.NewString(), time.Now().Unix(), a.state.NodeID,
		protocol.Hello{
			NodeID: a.state.NodeID, Name: a.cfg.Name, Hostname: hostname,
			OS: goos, Arch: arch, AgentVer: version,
			Tags: []string{goos}, Capabilities: map[string]string{"shell": "true"},
		})
	if err == nil {
		a.cli.Send(env)
	}
}

func (a *agent) heartbeatLoop(ctx context.Context) {
	t := time.NewTicker(a.cfg.HeartbeatInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			env, _ := protocol.NewEnvelope(protocol.TypeHeartbeat, uuid.NewString(), time.Now().Unix(),
				a.state.NodeID, protocol.Heartbeat{UptimeSec: a.coll.Uptime()})
			a.cli.Send(env)
		}
	}
}

func (a *agent) metricsLoop(ctx context.Context) {
	t := time.NewTicker(a.cfg.MetricsInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			m := a.coll.Snapshot()
			env, _ := protocol.NewEnvelope(protocol.TypeMetrics, uuid.NewString(), time.Now().Unix(),
				a.state.NodeID, m)
			a.cli.Send(env)
		}
	}
}
