# Service level objectives

These are **targets** chosen from load-test headroom (see README, "Measured performance"). There is no production
traffic yet, so attainment is not measured; the alert rules in [`deploy/prometheus/alerts.yml`](../deploy/prometheus/alerts.yml)
are what will measure it.

| Service | SLI | Objective | Basis |
|---|---|---|---|
| gateway | share of requests not answered `5xx` (excluding 401/429, which are client-side) | **99.9 %** over 30 days (43 min of budget) | breaker + retry keeps a single replica failure invisible to clients (0 failures in the kill-a-replica test) |
| topology-svc | p99 latency of `POST /v1/analyze/*` for topologies up to 1,000 nodes | **≤ 250 ms** | measured p99 ≈ 48 ms at 1,000 nodes with 5 concurrent clients on a 2-core box: ~5x headroom |
| topology-svc | share of requests shed with `429` | **< 1 %** | shedding is a capacity signal: add replicas or raise `TOPOLOGY_MAX_INFLIGHT` |
| change-svc | audit chain verifies (`/v1/audit/verify` → `ok:true`) | **100 %**, any failure pages | integrity is binary |

## Error-budget policy
- Budget > 50 % left: normal feature work.
- Budget < 50 % with a fast-burn alert fired: reliability work takes priority over features until the burn stops.
- Budget exhausted: freeze non-reliability releases for the rest of the window.

## Not covered yet
Per-tenant objectives, synthetic probing from outside the cluster, and latency SLOs for topologies above 1,000 nodes
(10,000-node preflight measured p50 ≈ 128 ms, p99 ≈ 224 ms at 4 clients; request decoding dominates).
