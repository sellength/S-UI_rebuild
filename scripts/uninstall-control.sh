#!/usr/bin/env sh
set -eu

INSTALL_DIR="${SUI_INSTALL_DIR:-/opt/s-ui-distributed}"
YES="false"
PURGE="false"

while [ "$#" -gt 0 ]; do
  case "$1" in
    --yes|-y)
      YES="true"
      shift
      ;;
    --purge)
      PURGE="true"
      shift
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [ "$(id -u)" -ne 0 ]; then
  echo "Please run as root." >&2
  exit 1
fi

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

if [ ! -d "$INSTALL_DIR" ]; then
  echo "Install dir not found: $INSTALL_DIR"
  exit 0
fi

if [ "$YES" != "true" ]; then
  echo "This will stop S-UI Distributed Control Plane in:"
  echo "  $INSTALL_DIR"
  if [ "$PURGE" = "true" ]; then
    echo "It will also remove persisted data because --purge is set."
  else
    echo "Persisted data will be kept. Use --purge to remove it."
  fi
  printf "Continue? [y/N] "
  read answer
  case "$answer" in
    y|Y|yes|YES) ;;
    *) echo "Canceled."; exit 0 ;;
  esac
fi

COMPOSE="$(compose_cmd)"
cd "$INSTALL_DIR"

if [ -f docker-compose.yml ]; then
  if [ -f .env ]; then
    $COMPOSE --env-file .env down || true
  else
    $COMPOSE down || true
  fi
fi

if [ "$PURGE" = "true" ]; then
  rm -rf "$INSTALL_DIR"
  echo "Removed $INSTALL_DIR."
else
  echo "Stopped containers. Data remains in $INSTALL_DIR."
fi
