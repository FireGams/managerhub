#!/usr/bin/env bash
# ManagerHub Agent installer — https://github.com/FireGams/managerhub
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/FireGams/managerhub/main/scripts/install.sh | bash -s -- \
#     --controller http://your-server:8080 --token mhk_xxx --name my-machine
set -euo pipefail

CONTROLLER=""
TOKEN=""
NAME=""
SERVICE=true
VERSION="latest"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --controller)  CONTROLLER="$2"; shift 2 ;;
    --token)       TOKEN="$2"; shift 2 ;;
    --name)        NAME="$2"; shift 2 ;;
    --version)     VERSION="$2"; shift 2 ;;
    --no-service)  SERVICE=false; shift ;;
    --help|-h)
      echo "Usage: install.sh --controller URL --token TOKEN [--name NAME] [--no-service] [--version TAG]"
      exit 0 ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

if [[ -z "$CONTROLLER" || -z "$TOKEN" ]]; then
  echo "ERROR: --controller and --token are required"
  echo "  curl -fsSL ... | bash -s -- --controller http://myserver:8080 --token mhk_xxx --name my-machine"
  exit 1
fi

# Detect OS / arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac
case "$OS" in
  linux|darwin) ;;
  msys*|mingw*|cygwin*) OS="windows" ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

EXT=""
if [[ "$OS" == "windows" ]]; then EXT=".exe"; fi
BINARY="managerhub-agent-${OS}-${ARCH}${EXT}"
REPO="${GITHUB_REPO:-FireGams/managerhub}"

if [[ "$VERSION" == "latest" ]]; then
  # Resolve the actual latest tag first
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": "\([^"]*\)".*/\1/')
  if [[ -z "$VERSION" ]]; then VERSION="v0.1.0"; fi
fi
URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}"

INSTALL_DIR="/usr/local/bin"
STATE_DIR="/var/lib/managerhub"
BIN_PATH="${INSTALL_DIR}/managerhub-agent${EXT}"

if [[ -z "$NAME" ]]; then NAME="$(hostname)"; fi

echo ">> Downloading ${BINARY} ..."
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$URL" -o "/tmp/${BINARY}"
elif command -v wget >/dev/null 2>&1; then
  wget -q "$URL" -O "/tmp/${BINARY}"
else
  echo "ERROR: curl or wget required"; exit 1
fi

echo ">> Installing to ${INSTALL_DIR} ..."
if [[ -w "$INSTALL_DIR" ]]; then
  mv "/tmp/${BINARY}" "$BIN_PATH"
  chmod +x "$BIN_PATH"
else
  sudo mv "/tmp/${BINARY}" "$BIN_PATH"
  sudo chmod +x "$BIN_PATH"
fi

mkdir -p "$STATE_DIR" 2>/dev/null || sudo mkdir -p "$STATE_DIR"

# Enroll
echo ">> Enrolling with controller ..."
export MH_CONTROLLER_URL="$CONTROLLER"
export MH_ENROLL_TOKEN="$TOKEN"
export MH_AGENT_NAME="$NAME"
export MH_AGENT_STATE="${STATE_DIR}/agent.state.json"
export MH_AUTO_UPDATE=true
"$BIN_PATH" &
AGENT_PID=$!
sleep 3
kill "$AGENT_PID" 2>/dev/null || true
echo ">> Enrolled (identity saved to ${STATE_DIR}/agent.state.json)"

if [[ "$SERVICE" == "true" ]]; then
  echo ">> Installing system service ..."
  if [[ "$OS" == "linux" ]] && command -v systemctl >/dev/null 2>&1; then
    sudo tee /etc/systemd/system/managerhub-agent.service > /dev/null << EOF
[Unit]
Description=ManagerHub Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=${BIN_PATH}
Environment=MH_CONTROLLER_URL=${CONTROLLER}
Environment=MH_AGENT_NAME="${NAME}"
Environment=MH_AGENT_STATE=${STATE_DIR}/agent.state.json
Environment=MH_AUTO_UPDATE=true
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
EOF
    sudo systemctl daemon-reload
    sudo systemctl enable --now managerhub-agent
    echo ">> Service installed and started (systemd)"

  elif [[ "$OS" == "darwin" ]]; then
    sudo tee /Library/LaunchDaemons/com.managerhub.agent.plist > /dev/null << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>com.managerhub.agent</string>
  <key>ProgramArguments</key>
  <array><string>${BIN_PATH}</string></array>
  <key>EnvironmentVariables</key>
  <dict>
    <key>MH_CONTROLLER_URL</key><string>${CONTROLLER}</string>
    <key>MH_AGENT_NAME</key><string>${NAME}</string>
    <key>MH_AGENT_STATE</key><string>${STATE_DIR}/agent.state.json</string>
    <key>MH_AUTO_UPDATE</key><string>true</string>
  </dict>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><true/>
</dict>
</plist>
EOF
    sudo launchctl load /Library/LaunchDaemons/com.managerhub.agent.plist
    echo ">> Service installed and started (launchd)"

  elif [[ "$OS" == "windows" ]]; then
    echo ">> Run as Administrator to register the Windows service:"
    echo "   sc.exe create ManagerHubAgent binPath= \"${BIN_PATH}\" start= auto"
    echo "   sc.exe start ManagerHubAgent"
  fi
else
  echo ">> Run manually: MH_AUTO_UPDATE=true MH_CONTROLLER_URL=${CONTROLLER} MH_AGENT_STATE=${STATE_DIR}/agent.state.json ${BIN_PATH}"
fi

echo ""
echo "============================================"
echo " ManagerHub Agent installed successfully!"
echo " Machine '${NAME}' will appear in the dashboard."
echo "============================================"
