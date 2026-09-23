# Installation

## Controller (Docker Compose)

```bash
cp .env.example .env
# edit MH_JWT_SECRET to a long random value
docker compose -f deploy/docker-compose.yml up -d
```

The controller runs migrations automatically at startup.

### Bootstrap admin

```bash
curl -X POST http://localhost:8080/api/v1/auth/bootstrap \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"very-secret"}'
```

### Login

```bash
JWT=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"very-secret"}' | jq -r .token)
```

## Agent

### Build

```bash
GOOS=linux GOARCH=amd64 make agent   # or windows/amd64, darwin/arm64 …
```

### Enroll

```bash
ENROLL=$(curl -s -X POST http://localhost:8080/api/v1/enroll-tokens \
  -H "Authorization: Bearer $JWT" \
  -H 'Content-Type: application/json' \
  -d '{"label":"my-server"}' | jq -r .token)

export MH_CONTROLLER_URL=https://manager.example.com
export MH_ENROLL_TOKEN=$ENROLL
export MH_AGENT_NAME=my-server
export MH_AGENT_STATE=/var/lib/managerhub/agent.state.json
./bin/agent
```

After the first run the state file holds `node_id` + `node_token`.
`MH_ENROLL_TOKEN` can be removed.

### Linux (systemd)

```ini
# /etc/systemd/system/managerhub-agent.service
[Unit]
Description=ManagerHub Agent
After=network-online.target

[Service]
User=mh
Environment=MH_CONTROLLER_URL=https://manager.example.com
Environment=MH_AGENT_STATE=/var/lib/managerhub/agent.state.json
ExecStart=/usr/local/bin/managerhub-agent
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable --now managerhub-agent
```

### Windows (Service)

```powershell
New-Service -Name ManagerHubAgent `
  -BinaryPathName "C:\ManagerHub\agent.exe" `
  -DisplayName "ManagerHub Agent" -StartupType Automatic
Start-Service ManagerHubAgent
```

### macOS (launchd)

```xml
<!-- /Library/LaunchDaemons/com.managerhub.agent.plist -->
<plist version="1.0"><dict>
  <key>Label</key><string>com.managerhub.agent</string>
  <key>ProgramArguments</key><array><string>/usr/local/bin/managerhub-agent</string></array>
  <key>EnvironmentVariables</key><dict>
    <key>MH_CONTROLLER_URL</key><string>https://manager.example.com</string>
  </dict>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
</dict></plist>
```

## Production hardening

- Terminate TLS in front of the controller (Caddy / nginx / Traefik).
- Set `MH_JWT_SECRET` to a 32+ byte random value.
- Never commit `.env`.
- Restrict enrollment token TTL.
- Run the agent as a non-root user (`mh`).
