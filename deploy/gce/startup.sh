#!/usr/bin/env bash
# GCE startup script (runs as root on every boot; idempotent).
# Reads its files from instance metadata, installs Docker, writes /opt/netorch and starts the stack.
set -euo pipefail
MD=http://metadata.google.internal/computeMetadata/v1/instance/attributes
attr() { curl -fsS -H 'Metadata-Flavor: Google' "$MD/$1"; }

if ! command -v docker >/dev/null; then
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -y
  apt-get install -y docker.io docker-compose-v2 openssl curl
  systemctl enable --now docker
fi

D=/opt/netorch
mkdir -p "$D"
attr compose    > "$D/compose.yml"
attr caddyfile  > "$D/Caddyfile"
attr prometheus > "$D/prometheus.yml"
attr alerts     > "$D/alerts.yml"

# secrets are generated once on the VM and never leave it
if [ ! -f "$D/.env" ]; then
  umask 077
  {
    echo "GATEWAY_JWT_SECRET=$(openssl rand -base64 36 | tr -d '\n')"
    echo "GHCR_OWNER=$(attr ghcr_owner)"
    echo "DOMAIN=$(attr domain)"
    echo "TAG=$(attr tag 2>/dev/null || echo latest)"
  } > "$D/.env"
fi

cd "$D"
docker compose -f compose.yml pull
docker compose -f compose.yml up -d --remove-orphans
