# ADR-0004: `change-svc` is a single-writer service

**Status:** accepted (with known scaling limit)

**Context.** The audit log is one hash chain; two writers would fork it. Approval decisions must be made against the latest state.

**Decision.** One replica owns the log (StatefulSet, 1 replica, PVC). All decisions are taken and appended under one lock. Read scaling and write scaling are not separated yet.

**Consequences.**
- (+) Simple, strongly consistent, easy to reason about and test (concurrent approval test under `-race`).
- (-) Availability is that of one pod + volume; throughput is bounded by fsync'd appends. This is appropriate for change management (human-paced, tens to hundreds of changes per day), not for high-rate event ingestion.
- Path forward if it ever matters: per-tenant chains partitioned across writers, or a replicated log (Raft, or a managed log such as Kafka/Postgres with a serial append) with the same `fold`.
