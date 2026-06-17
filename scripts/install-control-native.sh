#!/usr/bin/env sh
set -eu

REPO="${SUI_REPO:-sellength/S-UI_rebuild}"
INSTALL_DIR="${SUI_INSTALL_DIR:-/usr/local/s-ui}"
SERVICE_FILE="${SUI_SERVICE_FILE:-/etc/systemd/system/s-ui.service}"
ENV_FILE="${SUI_ENV_FILE:-${INSTALL_DIR}/s-ui.env}"
PACKAGE_URL="${SUI_PACKAGE_URL:-}"
NON_INTERACTIVE="${SUI_NON_INTERACTIVE:-false}"

while [ "$#" -gt 0 ]; do
  case "$1" in
    --yes|-y|--non-interactive)
      NON_INTERACTIVE="true"
      ;;
    --package-url)
      if [ "$#" -lt 2 ]; then
        echo "--package-url requires a value." >&2
        exit 1
      fi
      PACKAGE_URL="${2:-}"
      shift
      ;;
    --panel-port)
      if [ "$#" -lt 2 ]; then
        echo "--panel-port requires a value." >&2
        exit 1
      fi
      SUI_PANEL_PORT="${2:-}"
      shift
      ;;
    --sub-port)
      if [ "$#" -lt 2 ]; then
        echo "--sub-port requires a value." >&2
        exit 1
      fi
      SUI_SUB_PORT="${2:-}"
      shift
      ;;
    --panel-url)
      if [ "$#" -lt 2 ]; then
        echo "--panel-url requires a value." >&2
        exit 1
      fi
      SUI_PANEL_DOMAIN="${2:-}"
      shift
      ;;
    --postgres-dsn)
      if [ "$#" -lt 2 ]; then
        echo "--postgres-dsn requires a value." >&2
        exit 1
      fi
      SUI_POSTGRES_DSN="${2:-}"
      SUI_DB_TYPE="postgres"
      shift
      ;;
    --help|-h)
      cat <<'EOF'
Usage:
  install-control-native.sh [options]

This installs S-UI Control Plane directly on the host with systemd.
It does not use Docker.

Options:
  --package-url URL       Override release package URL.
  --panel-port PORT       Panel Web UI port. Default: 2095
  --sub-port PORT         Subscription service port. Default: 2096
  --panel-url URL         Public panel URL used in install output and docs.
  --postgres-dsn DSN      Use external PostgreSQL instead of default SQLite.
  --non-interactive, -y   Use defaults or SUI_* environment variables.

Environment variables with the same names are also supported.
EOF
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
  esac
  shift
done

if [ "$(id -u)" -ne 0 ]; then
  echo "Please run as root." >&2
  exit 1
fi

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing command: $1" >&2
    exit 1
  fi
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

random_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -base64 48 | tr -d '\n'
  else
    tr -dc 'A-Za-z0-9' </dev/urandom | head -c 64
  fi
}

detect_arch() {
  machine="$(uname -m)"
  case "$machine" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    armv7l|armv7*) echo "armv7" ;;
    armv6l|armv6*) echo "armv6" ;;
    armv5l|armv5*) echo "armv5" ;;
    i386|i686) echo "386" ;;
    s390x) echo "s390x" ;;
    *)
      echo "Unsupported architecture: $machine" >&2
      exit 1
      ;;
  esac
}

prompt_value() {
  label="$1"
  default="$2"
  result="$default"
  if [ "$NON_INTERACTIVE" != "true" ] && [ -r /dev/tty ]; then
    printf "%s [%s]: " "$label" "$default" >/dev/tty
    IFS= read -r input </dev/tty || input=""
    if [ -n "$input" ]; then
      result="$input"
    fi
  fi
  printf '%s' "$result"
}

validate_port() {
  name="$1"
  value="$2"
  case "$value" in
    ''|*[!0-9]*)
      echo "$name must be a number." >&2
      exit 1
      ;;
  esac
  if [ "$value" -lt 1 ] || [ "$value" -gt 65535 ]; then
    echo "$name must be between 1 and 65535." >&2
    exit 1
  fi
}

get_env_var() {
  key="$1"
  default="$2"
  value="$(grep "^${key}=" "$ENV_FILE" 2>/dev/null | tail -n 1 | cut -d= -f2- || true)"
  if [ -z "$value" ]; then
    value="$default"
  fi
  printf '%s' "$value"
}

shell_quote() {
  printf "%s" "$1" | sed "s/'/'\\\\''/g; 1s/^/'/; \$s/\$/'/"
}

need_cmd tar
need_cmd systemctl

arch="$(detect_arch)"
if [ -z "$PACKAGE_URL" ]; then
  PACKAGE_URL="https://github.com/${REPO}/releases/latest/download/s-ui-linux-${arch}.tar.gz"
fi

server_ip="SERVER_IP"
if command -v hostname >/dev/null 2>&1; then
  detected_ip="$(hostname -I 2>/dev/null | awk '{print $1}' || true)"
  if [ -n "$detected_ip" ]; then
    server_ip="$detected_ip"
  fi
fi

created_env="false"
if [ ! -f "$ENV_FILE" ]; then
  created_env="true"
  echo
  echo "S-UI Control Plane native setup"
  echo "This mode installs directly on the host with systemd. Docker is not used."
  echo "Press Enter to keep the default value."
  panel_port="$(prompt_value "Panel Web UI port" "${SUI_PANEL_PORT:-2095}")"
  sub_port="$(prompt_value "Subscription service port" "${SUI_SUB_PORT:-2096}")"
  validate_port "Panel Web UI port" "$panel_port"
  validate_port "Subscription service port" "$sub_port"

  panel_default="http://${server_ip}:${panel_port}"
  panel_url="$(prompt_value "Public panel URL" "${SUI_PANEL_DOMAIN:-$panel_default}")"
  secret_key="${SUI_SECRET_KEY:-$(random_secret)}"
  register_token="${SUI_AGENT_REGISTER_TOKEN:-$(random_secret)}"
else
  panel_port="$(get_env_var "SUI_PANEL_PORT" "${SUI_PANEL_PORT:-2095}")"
  sub_port="$(get_env_var "SUI_SUB_PORT" "${SUI_SUB_PORT:-2096}")"
  panel_url="$(get_env_var "SUI_PANEL_DOMAIN" "${SUI_PANEL_DOMAIN:-http://${server_ip}:${panel_port}}")"
  secret_key="$(get_env_var "SUI_SECRET_KEY" "${SUI_SECRET_KEY:-$(random_secret)}")"
  register_token="$(get_env_var "SUI_AGENT_REGISTER_TOKEN" "${SUI_AGENT_REGISTER_TOKEN:-$(random_secret)}")"
  validate_port "Panel Web UI port" "$panel_port"
  validate_port "Subscription service port" "$sub_port"
  echo "Existing $ENV_FILE detected, keeping current secrets."
fi

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

echo "Downloading S-UI package:"
echo "  $PACKAGE_URL"
if ! download "$PACKAGE_URL" "$tmp_dir/s-ui.tar.gz"; then
  echo "Failed to download release package." >&2
  echo "If you have not published a GitHub Release yet, pass --package-url URL to a manually uploaded s-ui-linux-${arch}.tar.gz." >&2
  exit 1
fi

tar -xzf "$tmp_dir/s-ui.tar.gz" -C "$tmp_dir"
if [ ! -x "$tmp_dir/s-ui/sui" ]; then
  echo "Invalid package: missing s-ui/sui." >&2
  exit 1
fi

systemctl stop s-ui 2>/dev/null || true
systemctl stop sing-box 2>/dev/null || true

mkdir -p "$INSTALL_DIR"
cp -R "$tmp_dir/s-ui/." "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/sui"
if [ -f "$INSTALL_DIR/bin/sing-box" ]; then
  chmod +x "$INSTALL_DIR/bin/sing-box"
fi
if [ -f "$INSTALL_DIR/bin/runSingbox.sh" ]; then
  chmod +x "$INSTALL_DIR/bin/runSingbox.sh"
fi

mkdir -p "$INSTALL_DIR/db" "$INSTALL_DIR/bin" "$INSTALL_DIR/cert" "$INSTALL_DIR/certificates" "$INSTALL_DIR/acme" "$INSTALL_DIR/logs"
chmod 700 "$INSTALL_DIR/db" "$INSTALL_DIR/cert" "$INSTALL_DIR/certificates" "$INSTALL_DIR/acme"

if [ ! -x "$INSTALL_DIR/acme/acme.sh" ]; then
  tmp_acme="$tmp_dir/acme-src"
  mkdir -p "$tmp_acme"
  echo "Installing acme.sh into $INSTALL_DIR/acme ..."
  download "https://github.com/acmesh-official/acme.sh/archive/refs/heads/master.tar.gz" "$tmp_acme/acme.tar.gz"
  tar -xzf "$tmp_acme/acme.tar.gz" -C "$tmp_acme"
  cp -R "$tmp_acme/acme.sh-master/." "$INSTALL_DIR/acme/"
  chmod +x "$INSTALL_DIR/acme/acme.sh"
else
  echo "Existing acme.sh detected, keeping current acme directory."
fi

db_type="${SUI_DB_TYPE:-sqlite}"
postgres_dsn="${SUI_POSTGRES_DSN:-}"
if [ -n "$postgres_dsn" ]; then
  db_type="postgres"
fi

cat > "$ENV_FILE" <<EOF
TZ=${TZ:-Asia/Shanghai}
SUI_DB_TYPE=${db_type}
SUI_DB_FOLDER=${INSTALL_DIR}/db
SUI_POSTGRES_DSN=$(shell_quote "$postgres_dsn")
SUI_BIN_FOLDER=${INSTALL_DIR}/bin
SUI_ACME_SH=${INSTALL_DIR}/acme/acme.sh
SUI_ACME_HOME=${INSTALL_DIR}/acme
SUI_CERTIFICATE_WORK_DIR=${INSTALL_DIR}/certificates
SUI_SECRET_KEY=$(shell_quote "$secret_key")
SUI_AGENT_REGISTER_TOKEN=$(shell_quote "$register_token")
SUI_PANEL_PORT=${panel_port}
SUI_SUB_PORT=${sub_port}
SUI_PANEL_DOMAIN=$(shell_quote "$panel_url")
EOF
chmod 600 "$ENV_FILE"

cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=S-UI Distributed Control Plane
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
EnvironmentFile=${ENV_FILE}
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/sui
Restart=on-failure
RestartSec=10s
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF

cat > /etc/systemd/system/sing-box.service <<EOF
[Unit]
Description=sing-box service
Documentation=https://sing-box.sagernet.org
After=network.target nss-lookup.target network-online.target

[Service]
EnvironmentFile=${ENV_FILE}
CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_BIND_SERVICE CAP_SYS_PTRACE CAP_DAC_READ_SEARCH
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_BIND_SERVICE CAP_SYS_PTRACE CAP_DAC_READ_SEARCH
WorkingDirectory=${INSTALL_DIR}/bin
ExecStart=${INSTALL_DIR}/bin/runSingbox.sh
ExecReload=/bin/kill -HUP \$MAINPID
Restart=on-failure
RestartSec=10s
LimitNOFILE=infinity

[Install]
WantedBy=multi-user.target
EOF

echo "Initializing database and panel settings ..."
set -a
. "$ENV_FILE"
set +a
"$INSTALL_DIR/sui" migrate >/dev/null || true
"$INSTALL_DIR/sui" setting -port "$panel_port" -subPort "$sub_port" >/dev/null

systemctl daemon-reload
systemctl enable s-ui >/dev/null
systemctl restart s-ui

echo
echo "S-UI Control Plane native install completed."
echo "Panel URL: ${panel_url}/app/"
echo "Install dir: $INSTALL_DIR"
echo "Env file: $ENV_FILE"
if [ "$created_env" = "true" ]; then
  echo
  echo "Generated SUI_SECRET_KEY and SUI_AGENT_REGISTER_TOKEN in $ENV_FILE."
  echo "Back up this file. Losing SUI_SECRET_KEY makes encrypted credentials unreadable."
fi
echo
echo "Useful commands:"
echo "  systemctl status s-ui"
echo "  journalctl -u s-ui -f"
