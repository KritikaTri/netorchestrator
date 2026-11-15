# ADR-0001: Topology analysis is a stateless service; the topology travels in the request

**Status:** accepted

**Context.** Preflight/resilience are pure functions of (topology, change). The CPU cost grows with topology size (10k nodes: ~28 ms in-process, see benchmarks), so this is the part that must scale out.

**Decision.** `topology-svc` holds no topology. Callers send nodes and links with each request. Results are cached per replica keyed by SHA-256 of (endpoint + body).

**Consequences.**
- (+) Any replica serves any request; scaling is adding replicas; no cache invalidation problem because the key *is* the input.
- (+) Deterministic, easy to test with golden inputs.
- (-) Large topologies mean large request bodies (10k nodes is ~0.9 MB JSON). Capped at 8 MiB; beyond that, a topology store keyed by version and a delta API would be the next step.
- (-) Per-replica caches are not shared; a shared cache only pays off if hit rates justify the extra hop.
