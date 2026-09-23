# Architecture

## Components

### Controller
Single Go binary exposing:
- REST API (`/api/v1`) for the web UI
- WebSocket endpoint (`/api/v1/agent/ws`) for agents
- Cron scheduler owning all recurring work
- PostgreSQL persistence

### Agent
Small Go binary installed as a system service. It:
- enrolls once with a one-time token, then keeps its identity on disk
- opens an **outbound** WSS connection to the controller
- sends heartbeats and resource metrics
- executes jobs and streams stdout/stderr
- hosts interactive PTY sessions
- lists/controls host services and GitHub runners

### Shared protocol
`shared/protocol` is the single source of truth for the wire format.
Both sides import it; it contains no I/O.

## Data model (PostgreSQL)

| Table | Purpose |
|---|---|
| `users` | Web accounts (bcrypt) |
| `enroll_tokens` | One-time agent enrollment secrets |
| `nodes` | Registered machines + live status |
| `node_metrics` | Resource samples (time-series, purged) |
| `node_events` | Connect / disconnect / enroll events |
| `jobs` | Job definitions, status, stdout/stderr |
| `schedules` | Cron templates owned by the controller |
| `schedule_runs` | Fired executions |
| `audit_log` | Administrative actions |
| `terminal_sessions` | Terminal session audit |

## Placement (scheduler V1)

1. Drop offline nodes
2. Drop nodes not matching `os` / `cpu_cores` / `min_ram_free`
3. Pick the node with the lowest CPU load
4. Assign the job

## Resilience

- Agent reconnects with exponential backoff (1s → 60s)
- Message envelopes carry a unique `id`; receivers drop duplicates
- Job IDs are idempotent — replaying `job_result` is safe
- Controller restart: agents redial; jobs re-queued stay `queued` until timeout
- Metrics and job logs are purged on a retention schedule
