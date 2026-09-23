# Roadmap

## M0 — Foundations (done)
- Repository skeleton, Go module, Makefile, lint, CI
- Docker Compose (PostgreSQL + controller)
- SQL migrations, `.env.example`, docs

## M1 — Core agent + controller (done)
- Versioned wire protocol (`shared/protocol`)
- Controller: config, store, auth (JWT + RBAC), WS hub, REST API
- Agent: identity persistence, WS client with backoff, heartbeat, metrics
- Unit tests (protocol, identity, auth, jobs)

## M2 — Jobs + scheduler
- Job execution with streamed stdout/stderr, exit code, cancel, timeout
- Central cron scheduler with enable/disable and run history
- Resource-based placement (os / cores / free RAM / tags)

## M3 — Terminal + services + GitHub runners
- Interactive PTY over WebSocket (bash / PowerShell)
- systemd / Windows Services / launchd management
- GitHub Actions runner detection, start/stop/restart, logs

## M4 — Web UI
- Dashboard, Machines, Jobs, Schedules, Runners, Logs, Settings
- Machine detail: Overview / Terminal / Services / Docker / Runner / Logs
- Responsive layout, xterm.js terminal

## M5 — Hardening
- Audit log UI, retention/purge controls
- Agent packaging (deb / msi / pkg)
- Integration tests against a live controller
- Docs polish

## Phase 2
File explorer, advanced Docker, process list, alerts, notifications,
secure agent auto-update, Wake-on-LAN, SSH fallback, runner auto-install,
ephemeral runners, custom tags, node groups, draining.

## Phase 3 — Compute cluster
Queues, priorities, max concurrency, CPU/RAM/GPU reservations,
affinity / anti-affinity, retries, draining, quotas.

## Phase 4 — AI
Capability declarations (`ollama`, `vllm`, `cuda`, `model:qwen`, …),
model-aware placement, independent task distribution across machines.
