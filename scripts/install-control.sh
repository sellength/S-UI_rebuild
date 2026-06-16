#!/usr/bin/env sh
set -eu

REPO="${SUI_REPO:-sellength/S-UI_rebuild}"
BRANCH="${SUI_BRANCH:-dev}"
INSTALL_DIR="${SUI_INSTALL_DIR:-/opt/s-ui-distributed}"
RAW_BASE="${SUI_RAW_BASE:-https://raw.githubusercontent.com/${REPO}/${BRANCH}}"
NON_INTERACTIVE="${SUI_NON_INTERACTIVE:-false}"

while [ "$#" -gt 0 ]; do
  case "$1" in
    --yes|-y|--non-interactive)
      NON_INTERACTIVE="true"
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
    --postgres-port)
      if [ "$#" -lt 2 ]; then
        echo "--postgres-port requires a value." >&2
        exit 1
      fi
      SUI_POSTGRES_PORT="${2:-}"
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
    --help|-h)
      cat <<'EOF'
Usage:
  install-control.sh [options]

Options:
  --panel-port PORT       Panel Web UI port. Default: 2095
  --sub-port PORT         Subscription service port. Default: 2096
  --postgres-port PORT    PostgreSQL host port bound to 127.0.0.1. Default: 54329
  --panel-url URL         Public panel URL used in install output and docs.
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

compose_cmd() {
  if docker compose version >/dev/null 2>&1; then
    echo "docker compose"
  elif command -v docker-compose >/dev/null 2>&1; then
    echo "docker-compose"
  else
    echo "Docker Compose is required." >&2
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

set_env_var() {
  key="$1"
  value="$2"
  file="${3:-.env}"
  tmp_file="$(mktemp)"
  if [ -f "$file" ]; then
    awk -v k="$key" -v v="$value" '
      BEGIN { done = 0 }
      $0 ~ "^" k "=" {
        print k "=" v
        done = 1
        next
      }
      { print }
      END {
        if (done == 0) {
          print k "=" v
        }
      }
    ' "$file" > "$tmp_file"
    mv "$tmp_file" "$file"
  else
    printf '%s=%s\n' "$key" "$value" > "$file"
    rm -f "$tmp_file"
  fi
}

get_env_var() {
  key="$1"
  default="$2"
  value="$(grep "^${key}=" .env 2>/dev/null | tail -n 1 | cut -d= -f2- || true)"
  if [ -z "$value" ]; then
    value="$default"
  fi
  printf '%s' "$value"
}

wait_for_sui_container() {
  i=0
  while [ "$i" -lt 30 ]; do
    if docker inspect -f '{{.State.Running}}' s-ui 2>/dev/null | grep -q true; then
      return 0
    fi
    i=$((i + 1))
    sleep 1
  done
  echo "s-ui container did not start in time." >&2
  return 1
}

need_cmd docker
need_cmd tar
if ! docker info >/dev/null 2>&1; then
  echo "Docker daemon is not running." >&2
  exit 1
fi

COMPOSE="$(compose_cmd)"

mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

mkdir -p db cert certificates acme
chmod 700 db cert certificates acme

download "${RAW_BASE}/docker-compose.yml" docker-compose.yml

if [ ! -x "acme/acme.sh" ]; then
  tmp_acme="$(mktemp -d)"
  echo "Installing acme.sh into $INSTALL_DIR/acme ..."
  download "https://github.com/acmesh-official/acme.sh/archive/refs/heads/master.tar.gz" "$tmp_acme/acme.tar.gz"
  tar -xzf "$tmp_acme/acme.tar.gz" -C "$tmp_acme"
  cp -R "$tmp_acme/acme.sh-master/." acme/
  chmod +x acme/acme.sh
  rm -rf "$tmp_acme"
else
  echo "Existing acme.sh detected, keeping current acme directory."
fi

created_env="false"
if [ ! -f ".env" ]; then
  download "${RAW_BASE}/.env.example" .env
  secret_key="$(random_secret)"
  register_token="$(random_secret)"
  postgres_password="$(random_secret)"
  sed -i "s|replace-with-a-long-random-postgres-password|${postgres_password}|g" .env
  sed -i "s|replace-with-a-long-random-secret-at-least-32-chars|${secret_key}|g" .env
  sed -i "s|replace-with-a-long-random-agent-register-token|${register_token}|g" .env
  created_env="true"
  chmod 600 .env
  echo "Generated .env with a new SUI_SECRET_KEY and SUI_AGENT_REGISTER_TOKEN."
else
  chmod 600 .env
  echo "Existing .env detected, keeping current secrets."
fi

server_ip="SERVER_IP"
if command -v hostname >/dev/null 2>&1; then
  detected_ip="$(hostname -I 2>/dev/null | awk '{print $1}' || true)"
  if [ -n "$detected_ip" ]; then
    server_ip="$detected_ip"
  fi
fi

if [ "$created_env" = "true" ]; then
  echo
  echo "S-UI Control Plane setup"
  echo "Press Enter to keep the default value."
  panel_port="$(prompt_value "Panel Web UI port" "${SUI_PANEL_PORT:-2095}")"
  sub_port="$(prompt_value "Subscription service port" "${SUI_SUB_PORT:-2096}")"
  postgres_port="$(prompt_value "PostgreSQL host port" "${SUI_POSTGRES_PORT:-54329}")"
  validate_port "Panel Web UI port" "$panel_port"
  validate_port "Subscription service port" "$sub_port"
  validate_port "PostgreSQL host port" "$postgres_port"

  panel_default="http://${server_ip}:${panel_port}"
  panel_url="$(prompt_value "Public panel URL" "${SUI_PANEL_DOMAIN:-$panel_default}")"

  set_env_var "SUI_PANEL_PORT" "$panel_port"
  set_env_var "SUI_SUB_PORT" "$sub_port"
  set_env_var "SUI_POSTGRES_PORT" "$postgres_port"
  set_env_var "SUI_PANEL_DOMAIN" "$panel_url"
else
  panel_port="$(get_env_var "SUI_PANEL_PORT" "2095")"
  sub_port="$(get_env_var "SUI_SUB_PORT" "2096")"
  postgres_port="$(get_env_var "SUI_POSTGRES_PORT" "54329")"
  validate_port "Panel Web UI port" "$panel_port"
  validate_port "Subscription service port" "$sub_port"
  validate_port "PostgreSQL host port" "$postgres_port"
fi

$COMPOSE --env-file .env up -d postgres s-ui
wait_for_sui_container

echo "Applying panel and subscription ports ..."
if $COMPOSE --env-file .env exec -T s-ui /usr/local/s-ui/sui setting -port "$panel_port" -subPort "$sub_port" >/dev/null 2>&1; then
  $COMPOSE --env-file .env restart s-ui >/dev/null
else
  echo "Warning: failed to apply panel/subscription ports inside S-UI. Check logs after startup." >&2
fi

$COMPOSE --env-file .env up -d

panel_url="$(grep '^SUI_PANEL_DOMAIN=' .env 2>/dev/null | cut -d= -f2- || true)"
if [ -z "$panel_url" ] || [ "$panel_url" = "https://panel.example.com" ]; then
  panel_url="http://${server_ip}:${panel_port}"
fi

echo
echo "S-UI Distributed Control Plane is starting."
echo "Install dir: $INSTALL_DIR"
echo "Panel: ${panel_url}/app/"
echo "Subscription service: port ${sub_port}, path /sub/ unless proxied"
echo
echo "Useful commands:"
echo "  cd $INSTALL_DIR && $COMPOSE --env-file .env ps"
echo "  cd $INSTALL_DIR && $COMPOSE --env-file .env logs -f s-ui"
