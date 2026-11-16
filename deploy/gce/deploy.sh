#!/usr/bin/env bash
# One-command GCE deployment. Run it from Cloud Shell (or any gcloud-authenticated shell):
#   ./deploy/gce/deploy.sh up | status | token | destroy
# Needs: gcloud authenticated + a project with billing (e2-small + static IP, a few $/month).
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT="${PROJECT:-$(gcloud config get-value project 2>/dev/null)}"
REGION="${REGION:-asia-south1}"
ZONE="${ZONE:-asia-south1-a}"
NAME="${NAME:-netorch-demo}"
MACHINE="${MACHINE:-e2-small}"
GHCR_OWNER="${GHCR_OWNER:-kritikatri}"
TAG="${TAG:-latest}"
[ -n "$PROJECT" ] || { echo "no project: run 'gcloud config set project <id>' or set PROJECT=" >&2; exit 2; }
G=(gcloud --project "$PROJECT")

ip() { "${G[@]}" compute addresses describe "$NAME-ip" --region "$REGION" --format='value(address)'; }

cmd_up() {
  "${G[@]}" services enable compute.googleapis.com >/dev/null
  "${G[@]}" compute addresses describe "$NAME-ip" --region "$REGION" >/dev/null 2>&1 \
    || "${G[@]}" compute addresses create "$NAME-ip" --region "$REGION"
  local IP DOMAIN; IP="$(ip)"; DOMAIN="${DOMAIN:-${IP//./-}.sslip.io}"

  "${G[@]}" compute firewall-rules describe "$NAME-web" >/dev/null 2>&1 \
    || "${G[@]}" compute firewall-rules create "$NAME-web" --allow tcp:80,tcp:443 --target-tags netorch

  if "${G[@]}" compute instances describe "$NAME" --zone "$ZONE" >/dev/null 2>&1; then
    echo "instance exists; updating its metadata and re-running the startup script"
    "${G[@]}" compute instances add-metadata "$NAME" --zone "$ZONE" \
      --metadata ghcr_owner="$GHCR_OWNER",domain="$DOMAIN",tag="$TAG" \
      --metadata-from-file startup-script="$HERE/startup.sh",compose="$HERE/compose.prod.yml",caddyfile="$HERE/Caddyfile",prometheus="$HERE/prometheus.yml",alerts="$HERE/alerts.yml"
    "${G[@]}" compute ssh "$NAME" --zone "$ZONE" --command 'sudo google_metadata_script_runner startup'
  else
    "${G[@]}" compute instances create "$NAME" --zone "$ZONE" --machine-type "$MACHINE" \
      --image-family ubuntu-2404-lts-amd64 --image-project ubuntu-os-cloud \
      --boot-disk-size 20GB --address "$IP" --tags netorch \
      --metadata ghcr_owner="$GHCR_OWNER",domain="$DOMAIN",tag="$TAG" \
      --metadata-from-file startup-script="$HERE/startup.sh",compose="$HERE/compose.prod.yml",caddyfile="$HERE/Caddyfile",prometheus="$HERE/prometheus.yml",alerts="$HERE/alerts.yml"
  fi
  echo
  echo "VM is booting (first start pulls images, ~2-3 min). Then:"
  echo "  curl -s https://$DOMAIN/healthz"
  echo "  ./deploy/gce/deploy.sh token     # mint a demo JWT"
}

cmd_status() {
  local IP; IP="$(ip)"; echo "https://${DOMAIN:-${IP//./-}.sslip.io}"
  "${G[@]}" compute ssh "$NAME" --zone "$ZONE" --command 'sudo docker compose -f /opt/netorch/compose.yml --env-file /opt/netorch/.env ps'
}

cmd_token() {
  "${G[@]}" compute ssh "$NAME" --zone "$ZONE" --command "sudo bash -s -- ${1:-demo} ${2:-1h}" < "$HERE/token.sh"
}

cmd_destroy() {
  read -r -p "Delete VM, static IP and firewall rule for $NAME (audit log on its disk is lost)? [y/N] " a
  [ "$a" = y ] || exit 1
  "${G[@]}" compute instances delete "$NAME" --zone "$ZONE" --quiet || true
  "${G[@]}" compute addresses delete "$NAME-ip" --region "$REGION" --quiet || true
  "${G[@]}" compute firewall-rules delete "$NAME-web" --quiet || true
}

case "${1:-}" in
  up) cmd_up ;; status) cmd_status ;; token) shift; cmd_token "$@" ;; destroy) cmd_destroy ;;
  *) echo "usage: $0 up|status|token [subject] [ttl]|destroy" >&2; exit 2 ;;
esac
