# Deploying to a single GCE VM

Why a VM and not Cloud Run: `change-svc` trusts the gateway-asserted identity header, `topology-svc` has no auth of its own, and the audit log needs a persistent disk. On one VM they sit on a private compose network and only Caddy is public. (See `docs/adr/` and `ARCHITECTURE.md`.)

## One-time prerequisites
1. Merge to `main`; the **publish images** workflow pushes `ghcr.io/kritikatri/netorch-{gateway,topology-svc,change-svc,devtoken}`.
2. GitHub → your profile → Packages → each `netorch-*` package → Settings → **Change visibility → Public** (otherwise the VM cannot pull without credentials).
3. A GCP project with billing enabled; open Cloud Shell.

## Run
```bash
./deploy/gce/deploy.sh up       # idempotent; re-run to roll out a new image/config
./deploy/gce/deploy.sh status   # prints the URL and `docker compose ps`
./deploy/gce/deploy.sh token alice 2h
curl -s https://<ip>.sslip.io/healthz
./deploy/gce/deploy.sh destroy
```
Overrides via env: `PROJECT REGION ZONE NAME MACHINE GHCR_OWNER TAG DOMAIN` (set `DOMAIN` to your own name pointing at the static IP).

## What you get
| Piece | Choice |
|---|---|
| VM | e2-small, Ubuntu 24.04, 20 GB, static IP, ports 80/443 only |
| TLS | Caddy, automatic Let's Encrypt; `/metrics` and `/readyz` return 404 publicly |
| Secrets | JWT secret generated on the VM (`/opt/netorch/.env`, mode 600), never in git or metadata |
| Containers | read-only rootfs, all caps dropped, no-new-privileges, memory limits |
| Public limits | 20 rps/subject, but only 1 rps (burst 5) on `/v1/changes` so the audit log cannot be flooded |
| Observability | Prometheus on 127.0.0.1:9090: `gcloud compute ssh netorch-demo -- -L 9090:localhost:9090` |

## Limits (honest)
Single VM = single point of failure, no autoscaling, rate-limit state per gateway replica, audit log on one disk. Horizontal scaling is shown separately on Kubernetes (`deploy/k8s/`, HPA).
