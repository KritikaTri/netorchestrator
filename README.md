# NetOrchestrator

[![CI](https://github.com/KritikaTri/netorchestrator/actions/workflows/ci.yml/badge.svg)](https://github.com/KritikaTri/netorchestrator/actions/workflows/ci.yml)

A Go toolkit and service platform for **safe network automation**: reason about a
topology and a config change *before* anything touches a device.

The project has two layers, and this README is explicit about which is which:

| Layer | What it is | Status |
|---|---|---|
| **Analysis core** — `internal/topology`, `internal/cfgdiff`, `cmd/netorchctl` | Dependency-free libraries + CLI | Tested (`-race`, oracle tests, ~90%+ coverage), required in CI |
| **Service platform** — API gateway, automation (Ansible/Terraform executors), security (JWT, API keys, rate limit, audit), event store, service registry, circuit breaker, WebSocket hub | Gin/GORM/Redis/Prometheus services | Implemented and compiling (`go build` + `go vet` are required in CI); **not yet covered by unit tests** |
| **Experimental** — `experimental/`: `microservices`, `distributed`, `observability`, `pipeline`, `enterprise`, `dashboard`, `infrastructure`, `compliance`, `business`, `sales`, `monetization`, plus the commands that wire them | Design explorations / simulations | Behind `//go:build experimental`; several do not compile yet (see [`experimental/README.md`](experimental/README.md)). `internal/ai` uses mock embeddings. Original vision: [`docs/archive/PLATFORM_VISION.md`](docs/archive/PLATFORM_VISION.md) |

## Analysis core

### `internal/topology` — graph analysis for network change planning
- **Shortest path** (Dijkstra, binary heap, deterministic tie-break)
- **Single points of failure**: articulation points *and* bridges via Tarjan low-link, O(V+E), iterative (a 200 000-node chain is part of the test suite)
- **Blast radius**: how many devices lose the main network if one fails
- Connected components, JSON loader with strict validation

Algorithms are checked against independent oracles: Floyd–Warshall for paths, and
brute-force "remove it and count components" for cut vertices/bridges, over hundreds
of random graphs. Benchmark: `make bench`.

### `internal/cfgdiff` — IOS-style config diff → minimal commands
Parses indentation-based configs into a tree and emits exactly the commands needed
to turn *running* into *desired* (`no <line>` for stale lines, whole-block negation,
context headers emitted once). A round-trip test applies the commands with a tiny
interpreter and asserts the diff converges to empty.

Known limitation (by design, documented in the package): sibling order is ignored, so
order-sensitive constructs such as ACL sequences and route-maps need extra handling.

### `cmd/netorchctl` — offline CLI
```bash
make demo
# or individually
go run ./cmd/netorchctl spof    -topology examples/campus.json
go run ./cmd/netorchctl blast   -topology examples/campus.json -node dist2
go run ./cmd/netorchctl path    -topology examples/campus.json -from access1 -to access3
go run ./cmd/netorchctl cfgdiff -desired examples/desired.cfg -running examples/running.cfg
```
Example output of the last command:
```
no interface Loopback9
interface Gi0/1
 no description legacy
 no switchport mode access
 description to-access1
 switchport mode trunk
router ospf 1
 no passive-interface default
```
Exit codes: `0` ok, `1` runtime error (e.g. no path), `2` usage error. It reads files only,
so it is safe in CI / change-review pipelines.

## Development

```bash
make test        # race-tested core packages
make fmt-check   # gofmt gate
make vet         # whole module (required in CI)
make build       # all commands into ./bin
```
Requires Go 1.24+ (see `go.mod`). `docker-compose.yml`, `helm/` and `k8s/` contain deployment
manifests for the service platform; `docs/API.md` and `docs/swagger/` describe its HTTP API.

## Roadmap
See [`ROADMAP.md`](ROADMAP.md). Next: unit tests for the security and automation packages,
and wire `topology` + `cfgdiff` into the API gateway as pre-change checks
(`POST /api/v1/changes/preflight`).
