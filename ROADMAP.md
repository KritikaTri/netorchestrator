# Roadmap

Ordered by what most improves correctness and credibility first. Items marked ✔ are shipped and tested.

## Shipped
- ✔ Analysis core: shortest path, single points of failure, blast radius, config diff (oracle-tested)
- ✔ Preflight: change-impact report with correct handling of already-partitioned networks and tied fragments
- ✔ Services: gateway, topology-svc, change-svc on a shared stdlib platform layer
- ✔ Four-eyes approval policy scaled by risk; event-sourced, hash-chained audit log; refuse-to-start on tamper
- ✔ Security fixes in the legacy API: no forged API keys, no default secrets, demo login off by default
- ✔ Load generator, benchmarks, e2e test in CI, Dockerfile build in CI, k8s manifests with HPA/PDB/NetworkPolicy, SLOs + alerts

## Next (highest value first)
1. **Sign audit events** with a key held outside the service and publish the chain head externally (closes the tail-truncation gap, [ADR-0002](docs/adr/0002-event-sourced-audit-log.md)).
2. **Audit denied attempts** (self-approval, duplicate approval, forbidden cancel).
3. **RBAC**: roles/groups from the identity provider (OIDC/JWKS) instead of subject-only four-eyes; per-tenant isolation of changes.
4. **Idempotency keys** on `POST /v1/changes`.
5. **DNS-based upstream discovery** in the gateway so each pod gets its own breaker in Kubernetes.
6. **Topology store with versions + delta API** to avoid shipping large topologies per request.
7. **Executor**: apply approved config diffs (`cfgdiff`) through a device driver, then record `change.applied` from the executor, not by hand.
8. **Distributed load test** on separate hosts to publish real scale-out numbers.
9. Legacy platform (Gin/GORM) — add unit tests, then retire in favour of the services above; fix refresh-token handling and per-user ownership checks.
10. Clean git history of large binaries (`git filter-repo`).
