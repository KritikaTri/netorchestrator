# ADR-0006: Rate limiting is a standalone package with two algorithms, applied per subject and per route

**Status:** accepted

**Context.** The first gateway used one hand-rolled token bucket behind a single mutex, returned a bare 429, and applied the same budget to cheap reads and to state-changing calls.

**Decision.** `internal/ratelimit` exposes one `Limiter` interface returning a `Decision` (allowed, limit, remaining, retry-after, reset) with two implementations:

| | Token bucket | Sliding-window counter |
|---|---|---|
| Behaviour | burst up to `burst`, then `rate`/s | at most `limit` per trailing `window` |
| Memory per key | 2 numbers | 3 numbers |
| Weak spot | allows a full burst after idling | an *approximation*: assumes the previous window was evenly spread |
| Measured | exact (tests) | worst overshoot 2 % vs an exact sliding log on random load (test `ApproximationVsExactLog`) |

The gateway applies a gateway-wide limit per authenticated subject plus an optional stricter per-route limit (default: `/v1/changes` and `/v1/audit/` share 10 req/s, burst 20). Responses carry `RateLimit-Limit`, `RateLimit-Remaining`, `RateLimit-Reset` and, on 429, `Retry-After`; `Retry-After` is computed exactly (client that waits that long succeeds, verified for both algorithms, including across a window rollover).

**Consequences.**
- (+) Clients can pace themselves from headers instead of guessing; a noisy user cannot starve others (keys are independent) and a flood of cheap reads cannot exhaust the budget for approvals (budgets are per route).
- (+) Idle keys are evicted lazily after the time they would need to fully refill, which provably does not change behaviour (an evicted key behaves like a fresh one), so memory follows *recently active* subjects.
- (+) Injectable clock: all timing behaviour is tested deterministically, no sleeps.
- (-) State is **per gateway replica**: with N replicas a subject can use up to N x the budget. A shared store (Redis with a Lua script, or consistent-hash routing of subjects to replicas) is needed for exact global limits; not implemented.
- (-) Keys are authenticated subjects, so cardinality is bounded by user count. If keys ever become attacker-controlled (e.g. rate limiting before authentication, by IP), add a hard cap on tracked keys.
- (-) Sharding (64 independent locks) is in place for many-core hosts, but on the 2-core test machine `shards=1` and `shards=64` were within noise of each other (170-360 ns/op, 0 allocs), so **no speed-up is claimed**.
