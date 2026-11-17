#!/usr/bin/env bash
# Runs ON the VM: mints a short-lived JWT for the demo API.  Usage: sudo ./token.sh [subject] [ttl]
set -euo pipefail
cd /opt/netorch
set -a; . ./.env; set +a
docker run --rm -e GATEWAY_JWT_SECRET \
  "ghcr.io/${GHCR_OWNER}/netorch-devtoken:${TAG:-latest}" -sub "${1:-demo}" -ttl "${2:-1h}"
