#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SYSTEMD_DIR="/etc/systemd/system"

echo "Installing systemd units..."
sudo cp "${PROJECT_ROOT}/systemd/hashcat-backend.service" "${SYSTEMD_DIR}/hashcat-backend.service"
sudo cp "${PROJECT_ROOT}/systemd/hashcat-frontend.service" "${SYSTEMD_DIR}/hashcat-frontend.service"

echo "Reloading systemd daemon..."
sudo systemctl daemon-reload

echo "Enabling services on boot..."
sudo systemctl enable hashcat-backend.service
sudo systemctl enable hashcat-frontend.service

echo "Starting/restarting services now..."
sudo systemctl restart hashcat-backend.service
sudo systemctl restart hashcat-frontend.service

echo
echo "Service status:"
sudo systemctl --no-pager --full status hashcat-backend.service || true
sudo systemctl --no-pager --full status hashcat-frontend.service || true
