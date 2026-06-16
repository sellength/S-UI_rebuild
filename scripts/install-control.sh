#!/usr/bin/env sh
set -eu

REPO="${SUI_REPO:-sellength/S-UI_rebuild}"
BRANCH="${SUI_BRANCH:-dev}"
INSTALL_DIR="${SUI_INSTALL_DIR:-/opt/s-ui-distributed}"
RAW_BASE="${SUI_RAW_BASE:-https://raw.githubusercontent.com/${REPO}/${BRANCH}}"

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

if [ ! -f ".env" ]; then
  download "${RAW_BASE}/.env.example" .env
  secret_key="$(random_secret)"
  register_token="$(random_secret)"
  postgres_password="$(random_secret)"
  sed -i "s|replace-with-a-long-random-postgres-password|${postgres_password}|g" .env
  sed -i "s|replace-with-a-long-random-secret-at-least-32-chars|${secret_key}|g" .env
  sed -i "s|replace-with-a-long-random-agent-register-token|${register_token}|g" .env
  chmod 600 .env
  echo "Generated .env with a new SUI_SECRET_KEY and SUI_AGENT_REGISTER_TOKEN."
else
  chmod 600 .env
  echo "Existing .env detected, keeping current secrets."
fi

$COMPOSE --env-file .env up -d

panel_url="$(grep '^SUI_PANEL_DOMAIN=' .env 2>/dev/null | cut -d= -f2- || true)"
if [ -z "$panel_url" ] || [ "$panel_url" = "https://panel.example.com" ]; then
  panel_url="http://SERVER_IP:2095"
fi

echo
echo "S-UI Distributed Control Plane is starting."
echo "Install dir: $INSTALL_DIR"
echo "Panel: ${panel_url}/app/"
echo "Subscription service: port 2096, path /sub/ unless proxied"
echo
echo "Useful commands:"
echo "  cd $INSTALL_DIR && $COMPOSE --env-file .env ps"
echo "  cd $INSTALL_DIR && $COMPOSE --env-file .env logs -f s-ui"
