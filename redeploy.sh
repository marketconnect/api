#!/bin/bash
set -e

echo "=== GIT PULL ==="
git pull

echo "=== BUILD ==="
go build -o /usr/local/bin/api ./app/cmd/main.go

echo "=== SYSTEMD RESTART ==="
sudo systemctl restart api.service

echo "=== SHOW JOURNAL ==="
sudo journalctl -u api.service -n 150 --no-pager
