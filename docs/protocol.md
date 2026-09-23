# Agent protocol

WebSocket, JSON text frames. Every frame is wrapped in an **envelope**:

```json
{
  "v": 1,
  "type": "heartbeat",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "ts": 1727000000,
  "node_id": "uuid",
  "payload": {}
}
```

- `v` — protocol version (current: 1). Unknown versions are rejected.
- `id` — unique message id. Receivers keep an LRU of recent ids and drop duplicates.
- `payload` — type-specific body.

## Authentication

```
GET /api/v1/agent/ws
Authorization: Bearer <node_token>
```

Enrollment (once):

```
POST /api/v1/agent/enroll
{ "enroll_token": "mhk_...", "name": "...", "hostname": "...", "os": "...", "arch": "..." }
→ 201 { "node_id": "...", "node_token": "mhk_..." }
```

## Message types

### Agent → Controller

| Type | Payload | Purpose |
|---|---|---|
| `hello` | `Hello` | Announce identity after connect |
| `heartbeat` | `Heartbeat` | Liveness + uptime |
| `metrics` | `Metrics` | CPU / RAM / disk / GPU snapshot |
| `job_output` | `JobOutput` | Streamed stdout/stderr |
| `job_result` | `JobResult` | Final status + exit code |
| `term_output` | `TermOutput` | PTY output (base64) |
| `term_closed` | `TermClosed` | Session ended |
| `svc_list_result` | `SvcListResult` | Service inventory |
| `svc_action_result` | `SvcActionResult` | Service action outcome |
| `runner_list_result` | `RunnerListResult` | GitHub runners |
| `runner_action_result` | `RunnerActionResult` | Runner action + logs |
| `pong` | `PingPong` | Keepalive reply |
| `error` | `ErrorPayload` | Generic error |

### Controller → Agent

| Type | Payload | Purpose |
|---|---|---|
| `hello_ack` | `HelloAck` | Registration confirmed |
| `ping` | `PingPong` | Keepalive probe |
| `job_assign` | `JobAssign` | Execute a job |
| `job_cancel` | `JobCancel` | Terminate a job |
| `term_open` / `term_input` / `term_resize` / `term_close` | `Term*` | Interactive shell |
| `svc_list` / `svc_action` | `Svc*` | Service management |
| `runner_list` / `runner_action` | `Runner*` | GitHub runner management |

## Versioning

The controller accepts `v` in `[1, CurrentVersion]`. New optional fields are
added inside `payload` only — envelopes stay backward compatible. Agents
report their build version in `hello.agent_version` for inventory purposes.
