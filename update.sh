#!/usr/bin/env bash
# update.sh — pull latest code and restart the service (run on the server)
set -euo pipefail

echo "▸ pulling latest..."
git pull

echo "▸ building..."
go build -ldflags="-s -w" -o monita-collector .
sudo mv monita-collector /usr/local/bin/monita-collector

echo "▸ running migrations..."
export $(grep -v '^#' /etc/monita-collector/.env | xargs)
migrate -path internal/db/migrations -database "$DB_URL" up

echo "▸ restarting service..."
sudo systemctl restart monita-collector
sudo systemctl status monita-collector --no-pager

echo "✓ done"
