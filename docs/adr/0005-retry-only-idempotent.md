# ADR-0005: The gateway retries only routes declared idempotent

**Status:** accepted

**Decision.** Each gateway route is configured `Idempotent: true/false`. Idempotent routes (`/v1/analyze/*`, pure functions) are retried once on a different replica after a transport error or 502/503/504. State-changing routes (`/v1/changes`, `/v1/audit/`) are never retried: replaying an approval could double-apply it. 4xx responses and 429 pass through and do not count against a replica's circuit breaker (they are the caller's problem, not the replica's).

**Consequences.** Retries cannot cause duplicate side effects. The gateway's upstream list is static (comma-separated URLs), so Kubernetes deployments point it at a Service and rely on kube-proxy for connection-level balancing; DNS-based discovery with per-pod breakers is on the roadmap.
