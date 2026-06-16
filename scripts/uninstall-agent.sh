#!/usr/bin/env sh
set -eu

AGENT_DIR="${AGENT_DIR:-/usr/local/s-ui-agent}"
SERVICE_FILE="${SERVICE_FILE:-/etc/systemd/system/s-ui-agent.service}"
YES="false"
KEEP_DATA="false"

while [ "$#" -gt 0 ]; do
  case "$1" in
    --yes|-y)
      YES="true"
      shift
      ;;
    --keep-data)
      KEEP_DATA="true"
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

if [ "$YES" != "true" ]; then
  echo "This will stop and remove s-ui-agent service."
  if [ "$KEEP_DATA" = "true" ]; then
    echo "Agent data will be kept in $AGENT_DIR."
  else
    echo "Agent data will be removed from $AGENT_DIR."
  fi
  printf "Continue? [y/N] "
  read answer
  case "$answer" in
    y|Y|yes|YES) ;;
    *) echo "Canceled."; exit 0 ;;
  esac
fi

if command -v systemctl >/dev/null 2>&1; then
  systemctl stop s-ui-agent 2>/dev/null || true
  systemctl disable s-ui-agent 2>/dev/null || true
fi

rm -f "$SERVICE_FILE"

if command -v systemctl >/dev/null 2>&1; then
  systemctl daemon-reload || true
  systemctl reset-failed s-ui-agent 2>/dev/null || true
fi

if [ "$KEEP_DATA" != "true" ]; then
  rm -rf "$AGENT_DIR"
fi

echo "s-ui-agent uninstalled."
