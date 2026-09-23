# ManagerHub

Centralized management of machines and compute resources: server fleet,
GitHub Actions self-hosted runners, remote jobs and cron scheduling.

## Architecture

```
                    CONTROLLER
               Web UI + REST API + WS hub
                        │
                  HTTPS / WSS
                        │
      ┌─────────────────┼─────────────────┐
      ▼                 ▼                 ▼
  Linux Agent      Windows Agent      macOS Agent
      │                 │                 │
  CPU/RAM/GPU       CPU/RAM/GPU       CPU/RAM/GPU
  Services          Services          Services
  Docker            Processes         Processes
  GitHub Runner     GitHub Runner     GitHub Runner
  Jobs              Jobs              Jobs
```

Agents open **outbound** WebSocket connections to the controller — no inbound
ports required on managed machines.

## Repository layout

| Path | Content |
|---|---|
| `shared/protocol` | Versioned Controller ↔ Agent wire format |
| `controller` | REST API, WS hub, scheduler, PostgreSQL store |
| `agent` | Node agent (metrics, jobs, terminal, services, runners) |
| `web` | React dashboard |
| `docs` | Architecture, protocol, install guides |
| `deploy` | Docker Compose + Dockerfiles |

## Quick install (one command)

Install the agent on any Linux/macOS machine:

```bash
curl -fsSL https://raw.githubusercontent.com/managerhub/managerhub/main/scripts/install.sh | bash -s -- \
  --controller http://your-server:8080 \
  --token mhk_xxxxx \
  --name my-machine
```

It auto-detects OS/arch, downloads the binary, enrolls and installs the service.

---

## Quick start (dev)

```bash
# 1. Start PostgreSQL + controller
cp .env.example .env
docker compose -f deploy/docker-compose.yml up -d

# 2. Create the first admin
curl -X POST localhost:8080/api/v1/auth/bootstrap \
  -d '{"username":"admin","password":"changeme123"}'

# 3. Create an enrollment token
TOKEN=$(curl -s -X POST localhost:8080/api/v1/enroll-tokens \
  -H "Authorization: Bearer $JWT" -d '{"label":"laptop"}' | jq -r .token)

# 4. Enroll an agent
export MH_CONTROLLER_URL=http://localhost:8080
export MH_ENROLL_TOKEN=$TOKEN
make agent && ./bin/agent
```

See [docs/install.md](docs/install.md) for production setup and
[docs/protocol.md](docs/protocol.md) for the agent protocol.

## Development

```bash
make build     # compile controller + agent
make test      # unit tests
make lint      # golangci-lint
make web       # build the dashboard
```

## Security

- HTTPS/WSS only in production
- Per-agent identity tokens (SHA-256 at rest)
- One-time enrollment tokens
- JWT sessions + RBAC (`admin`, `operator`, `viewer`)
- Audit log of every administrative action
- Job timeouts, terminal session expiry
- Agent runs unprivileged by default

## License

Open source — see LICENSE.
