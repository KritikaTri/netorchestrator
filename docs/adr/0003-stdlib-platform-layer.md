# ADR-0003: Cross-cutting concerns live in a stdlib-only `internal/platform`

**Status:** accepted

**Context.** The original monolith pulled in a large dependency tree (web framework, ORM, redis, kafka, tracing). Much of it was not exercised by tests, and the module could not even be built in a restricted environment.

**Decision.** Metrics (Prometheus text format), request IDs, access logs, token bucket, circuit breaker, HS256 verification and graceful shutdown are implemented on the standard library (`net/http` routing with method patterns, `log/slog`).

**Consequences.**
- (+) Services compile to small static binaries; nothing in the auth path depends on third-party code except `crypto/hmac`.
- (+) The whole layer is race-tested and runs without external services.
- (-) We own the code. JWT verification accepts **only HS256** and is tested against `alg=none`, tampering and expiry, but a production deployment with key rotation/JWKS should move to a vetted library or an identity-aware proxy.
- (-) The Prometheus exposition is hand-written; swap in `client_golang` if you need exemplars or native histograms.
