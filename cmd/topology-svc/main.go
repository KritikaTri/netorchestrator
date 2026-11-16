// Command topology-svc is the stateless topology analysis service.
package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"netorchestrator/internal/platform"
	"netorchestrator/internal/topologysvc"
)

func main() {
	log := platform.Logger("topology-svc")
	reg := platform.NewRegistry()

	cfg := topologysvc.DefaultConfig()
	if v, err := strconv.Atoi(platform.Env("TOPOLOGY_MAX_INFLIGHT", "")); err == nil {
		cfg.MaxInFlight = v
	}
	if v, err := strconv.Atoi(platform.Env("TOPOLOGY_CACHE_ENTRIES", "")); err == nil {
		cfg.CacheEntries = v
	}

	mux := http.NewServeMux()
	topologysvc.New(cfg, reg).Routes(mux)
	platform.Health(mux, reg, nil)

	srv := &http.Server{
		Addr:              platform.Env("ADDR", ":8081"),
		Handler:           platform.Chain(mux, platform.WithRequestID(), platform.WithObservability("topology-svc", reg, log, routeLabel)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := platform.Serve(context.Background(), srv, 10*time.Second, log); err != nil {
		log.Error("server failed", "err", err)
		os.Exit(1)
	}
}

// routeLabel keeps metric cardinality bounded: only registered mux patterns
// are used as labels, never raw URL paths.
func routeLabel(r *http.Request) string {
	if r.Pattern != "" {
		return r.Pattern
	}
	return "unmatched"
}
