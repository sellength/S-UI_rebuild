#!/usr/bin/env sh
set -eu

INSTALL_DIR="${SUI_INSTALL_DIR:-/usr/local/s-ui}"
SERVICE_FILE="${SUI_SERVICE_FILE:-/etc/systemd/system/s-ui.service}"
YES="false"
PURGE="false"

while [ "$#" -gt 0 ]; do
  case "$1" in
    --yes|-y)
      YES="true"
      ;;
    --purge)
      PURGE="true"
      ;;
    --help|-h)
      cat <<'EOF'
Usage:
  uninstall-control-native.sh [--yes] [--purge]

Stops and removes the native systemd S-UI Control Plane service.
Without --purge, data under /usr/local/s-ui is kept.
EOF
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
  shift
done

if [ "$(id -u)" -ne 0 ]; then
  echo "Please run as root." >&2
  exit 1
fi

if [ "$YES" != "true" ]; then
  echo "This will stop and remove the native S-UI Control Plane systemd service."
  if [ "$PURGE" = "true" ]; then
    echo "It will also delete $INSTALL_DIR because --purge is set."
  else
    echo "Data in $INSTALL_DIR will be kept. Use --purge to remove it."
  fi
  printf "Continue? [y/N] "
  read answer
  case "$answer" in
    y|Y|yes|YES) ;;
    *) echo "Canceled."; exit 0 ;;
  esac
fi

systemctl stop s-ui 2>/dev/null || true
systemctl disable s-ui 2>/dev/null || true
rm -f "$SERVICE_FILE"
systemctl daemon-reload
systemctl reset-failed s-ui 2>/dev/null || true

if [ "$PURGE" = "true" ]; then
  rm -rf "$INSTALL_DIR"
  echo "Removed $INSTALL_DIR."
else
  echo "Removed native service. Data remains in $INSTALL_DIR."
fi
