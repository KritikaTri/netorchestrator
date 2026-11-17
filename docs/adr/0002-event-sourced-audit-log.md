# ADR-0002: Change state is a fold over an append-only, hash-chained event log

**Status:** accepted

**Context.** Network changes need an answer to "who approved what, based on which analysis, and has the record been altered?"

**Decision.** `change-svc` stores only events. Each event carries `prev_hash` and `hash = SHA-256(length-prefixed fields)`. In-memory state is rebuilt by folding the log at startup, and the same `fold` function is used at runtime, so replay and live behaviour cannot diverge. The preflight report and a hash of the exact topology are part of `change.submitted`.

**Consequences.**
- (+) Editing, deleting or reordering any event breaks every later hash; the service refuses to start on a bad log, and `/v1/audit/verify` reports it.
- (+) Length-prefixed hashing prevents field-boundary ambiguity (tested).
- (+) A write that fails on disk is not applied in memory: memory never runs ahead of the log.
- (-) A hash chain cannot detect **tail truncation** or a full rewrite by someone who can recompute all hashes. Mitigation: publish `head_hash` to a system the service cannot write (ticket, git tag, WORM bucket). Signing events with a key held outside the service is the next step.
- (-) Denied attempts (e.g. self-approval) are returned to the caller but not logged yet.
