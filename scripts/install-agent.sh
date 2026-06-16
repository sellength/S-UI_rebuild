#!/usr/bin/env sh
set -eu

AGENT_BIN="${AGENT_BIN:-/usr/local/s-ui-agent/s-ui-agent}"
AGENT_DIR="${AGENT_DIR:-/usr/local/s-ui-agent}"
SERVICE_FILE="${SERVICE_FILE:-/etc/systemd/system/s-ui-agent.service}"
GITHUB_REPO="${SUI_AGENT_GITHUB_REPO:-sellength/S-UI_rebuild}"
DOWNLOAD_URL="${SUI_AGENT_DOWNLOAD_URL:-}"

PANEL_AGENT_URL=""
NODE_CODE=""
AGENT_ID=""
AGENT_TOKEN=""
REGISTER_TOKEN=""
INTERVAL="${INTERVAL:-30s}"
RELOAD_COMMAND="${RELOAD_COMMAND:-}"

while [ "$#" -gt 0 ]; do
  case "$1" in
    --panel-agent-url)
      PANEL_AGENT_URL="$2"
      shift 2
      ;;
    --node-code)
      NODE_CODE="$2"
      shift 2
      ;;
    --agent-id)
      AGENT_ID="$2"
      shift 2
      ;;
    --agent-token)
      AGENT_TOKEN="$2"
      shift 2
      ;;
    --register-token)
      REGISTER_TOKEN="$2"
      shift 2
      ;;
    --interval)
      INTERVAL="$2"
      shift 2
      ;;
    --reload-command)
      RELOAD_COMMAND="$2"
      shift 2
      ;;
    --agent-download-url)
      DOWNLOAD_URL="$2"
      shift 2
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [ -z "$PANEL_AGENT_URL" ] || [ -z "$NODE_CODE" ] || [ -z "$AGENT_ID" ] || [ -z "$AGENT_TOKEN" ] || [ -z "$REGISTER_TOKEN" ]; then
  echo "Usage: $0 --panel-agent-url URL --node-code CODE --agent-id ID --agent-token TOKEN --register-token TOKEN [--interval 30s] [--agent-download-url URL]" >&2
  exit 1
fi

if [ "$(id -u)" -ne 0 ]; then
  echo "Please run as root." >&2
  exit 1
fi

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    armv7l|armv7*) echo "armv7" ;;
    armv6l|armv6*) echo "armv6" ;;
    armv5l|armv5*) echo "armv5" ;;
    i386|i686) echo "386" ;;
    s390x) echo "s390x" ;;
    *) echo "unsupported" ;;
  esac
}

download() {
  url="$1"
  output="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$output"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$output" "$url"
  else
    echo "Missing curl or wget." >&2
    exit 1
  fi
}

if [ ! -x "$AGENT_BIN" ]; then
  arch="$(detect_arch)"
  if [ "$arch" = "unsupported" ]; then
    echo "Unsupported CPU architecture: $(uname -m)" >&2
    exit 1
  fi
  if [ -z "$DOWNLOAD_URL" ]; then
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/s-ui-agent-linux-${arch}"
  fi
  mkdir -p "$(dirname "$AGENT_BIN")"
  tmp_bin="${AGENT_BIN}.download"
  echo "Agent binary not found. Downloading from:"
  echo "  $DOWNLOAD_URL"
  if ! download "$DOWNLOAD_URL" "$tmp_bin"; then
    rm -f "$tmp_bin"
    echo "Failed to download Agent binary." >&2
    echo "You can copy it manually to $AGENT_BIN, or pass --agent-download-url URL." >&2
    exit 1
  fi
  chmod +x "$tmp_bin"
  mv "$tmp_bin" "$AGENT_BIN"
fi

mkdir -p "$AGENT_DIR/configs" "$AGENT_DIR/certs" "$AGENT_DIR/logs"

systemd_env_value() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g; s/%/%%/g'
}

cat > "$AGENT_DIR/agent.env" <<EOF
SUI_AGENT_BASE_URL="$(systemd_env_value "$PANEL_AGENT_URL")"
SUI_NODE_CODE="$(systemd_env_value "$NODE_CODE")"
SUI_AGENT_ID="$(systemd_env_value "$AGENT_ID")"
SUI_AGENT_TOKEN="$(systemd_env_value "$AGENT_TOKEN")"
SUI_AGENT_REGISTER_TOKEN="$(systemd_env_value "$REGISTER_TOKEN")"
SUI_AGENT_INTERVAL="$(systemd_env_value "$INTERVAL")"
SUI_AGENT_CONFIG_DIR="$(systemd_env_value "$AGENT_DIR/configs")"
SUI_AGENT_CERT_DIR="$(systemd_env_value "$AGENT_DIR/certs")"
SUI_AGENT_STATE_PATH="$(systemd_env_value "$AGENT_DIR/state.json")"
SUI_SINGBOX_BIN="$(systemd_env_value "${SUI_SINGBOX_BIN:-sing-box}")"
SUI_AGENT_CHECK_CONFIG="$(systemd_env_value "${SUI_AGENT_CHECK_CONFIG:-true}")"
SUI_AGENT_RELOAD_COMMAND="$(systemd_env_value "$RELOAD_COMMAND")"
EOF

chmod 600 "$AGENT_DIR/agent.env"

cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=S-UI Node Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
EnvironmentFile=$AGENT_DIR/agent.env
ExecStart=$AGENT_BIN
Restart=always
RestartSec=5
WorkingDirectory=$AGENT_DIR

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable s-ui-agent
systemctl restart s-ui-agent

echo "s-ui-agent installed and started."
