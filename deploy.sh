#!/usr/bin/env bash
# deploy.sh — build and ship monita-collector to production
# Usage: ./deploy.sh [host] [user]
#
# Examples:
#   ./deploy.sh 1.2.3.4
#   ./deploy.sh 1.2.3.4 ubuntu
#   DEPLOY_HOST=1.2.3.4 ./deploy.sh

set -euo pipefail

# ── Config ────────────────────────────────────────────────────────────────────
DEPLOY_HOST="${1:-${DEPLOY_HOST:?'set DEPLOY_HOST or pass it as first arg'}}"
DEPLOY_USER="${2:-${DEPLOY_USER:-ubuntu}}"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

BINARY="monita-collector"
REMOTE_BIN="/usr/local/bin/${BINARY}"
REMOTE_WORKDIR="/etc/monita-collector"
SERVICE_NAME="monita-collector"

# ── Build ─────────────────────────────────────────────────────────────────────
echo "▸ building linux/amd64 binary..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o "${BINARY}" .
echo "  binary size: $(du -sh "${BINARY}" | cut -f1)"

# ── Upload ────────────────────────────────────────────────────────────────────
echo "▸ uploading binary..."
scp "${BINARY}" "${REMOTE}:/tmp/${BINARY}"

echo "▸ uploading Caddyfile..."
scp Caddyfile "${REMOTE}:/tmp/Caddyfile"

echo "▸ uploading systemd unit..."
scp deploy/monita-collector.service "${REMOTE}:/tmp/${SERVICE_NAME}.service"

echo "▸ uploading migrations..."
rsync -az --delete internal/db/migrations/ "${REMOTE}:${REMOTE_WORKDIR}/migrations/"

# ── Remote setup ──────────────────────────────────────────────────────────────
echo "▸ installing on server..."
ssh "${REMOTE}" bash -s <<EOF
set -euo pipefail

# Create service user if missing
if ! id monita &>/dev/null; then
  useradd --system --no-create-home --shell /usr/sbin/nologin monita
  echo "  created user: monita"
fi

# Working dir
mkdir -p ${REMOTE_WORKDIR}
chown monita:monita ${REMOTE_WORKDIR}

# Copy .env if it doesn't exist yet (first deploy)
if [ ! -f ${REMOTE_WORKDIR}/.env ]; then
  echo ""
  echo "  ⚠  No .env found at ${REMOTE_WORKDIR}/.env"
  echo "     Create it before the service will start:"
  echo "       DB_URL=postgres://..."
  echo "       JWT_SECRET=..."
fi

# Install binary
install -o root -g root -m 0755 /tmp/${BINARY} ${REMOTE_BIN}

# Install systemd unit
install -o root -g root -m 0644 /tmp/${SERVICE_NAME}.service /etc/systemd/system/${SERVICE_NAME}.service
systemctl daemon-reload

# Stop service before migration
if systemctl is-active --quiet ${SERVICE_NAME}; then
  systemctl stop ${SERVICE_NAME}
fi

# Run migrations
if command -v migrate &>/dev/null; then
  source ${REMOTE_WORKDIR}/.env 2>/dev/null || true
  if [ -n "\${DB_URL:-}" ]; then
    echo "  running migrations..."
    migrate -path ${REMOTE_WORKDIR}/migrations -database "\${DB_URL}" up
  else
    echo "  ⚠  DB_URL not set — skipping migrations"
  fi
else
  echo "  ⚠  migrate CLI not found — skipping migrations"
  echo "     Install: https://github.com/golang-migrate/migrate"
fi

# Start / enable service
systemctl enable --now ${SERVICE_NAME}
systemctl restart ${SERVICE_NAME}
sleep 1
systemctl status ${SERVICE_NAME} --no-pager

# Install Caddyfile and reload
install -o root -g root -m 0644 /tmp/Caddyfile /etc/caddy/Caddyfile
if systemctl is-active --quiet caddy; then
  caddy reload --config /etc/caddy/Caddyfile
  echo "  caddy reloaded"
else
  systemctl enable --now caddy
  echo "  caddy started"
fi

echo ""
echo "✓ deployed to ${DEPLOY_HOST}"
EOF

# ── Cleanup ───────────────────────────────────────────────────────────────────
rm -f "${BINARY}"
echo "✓ done"
