# experimental/

Design explorations and simulations that are **not part of the supported build**.
Every file carries `//go:build experimental`, so `go build ./...` ignores them.
Several are known not to compile yet (types referenced but never defined).

Build them explicitly with `go build -tags experimental ./experimental/...`
once you are ready to finish one; promote it back to `internal/` when it has tests.
