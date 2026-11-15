// Command gateway is the public entry point of NetOrchestrator.
package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"netorchestrator/internal/gateway"
	"netorchestrator/internal/platform"
)

func main() {
	log := platform.Logger("gateway")
	reg := platform.NewRegistry()

	secret := platform.Env("GATEWAY_JWT_SECRET", "")
	rps, _ := strconv.ParseFloat(platform.Env("GATEWAY_RATE_RPS", "200"), 64)
	burst, _ := strconv.Atoi(platform.Env("GATEWAY_RATE_BURST", "400"))

	routes := []gateway.Route{{
		Prefix:     "/v1/analyze/",
		Upstream:   split(platform.Env("GATEWAY_TOPOLOGY_UPSTREAMS", "http://localhost:8081")),
		Idempotent: true, // pure analysis: safe to retry on another replica
	}}
	if u := platform.Env("GATEWAY_CHANGE_UPSTREAMS", ""); u != "" {
		// Stateful, non-idempotent: never retried on another replica.
		routes = append(routes,
			gateway.Route{Prefix: "/v1/changes", Upstream: split(u)},
			gateway.Route{Prefix: "/v1/audit/", Upstream: split(u)})
	}

	gw, err := gateway.New(gateway.Config{
		Routes:        routes,
		JWTSecret:     []byte(secret),
		APIKeySHA256:  split(platform.Env("GATEWAY_API_KEY_SHA256", "")),
		RatePerSecond: rps,
		Burst:         burst,
	}, reg)
	if err != nil {
		log.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/v1/", gw)
	platform.Health(mux, reg, gw.Ready)

	srv := &http.Server{
		Addr: platform.Env("ADDR", ":8080"),
		Handler: platform.Chain(mux, platform.WithRequestID(),
			platform.WithObservability("gateway", reg, log, func(r *http.Request) string {
				if r.Pattern != "" {
					return r.Pattern
				}
				return "unmatched"
			})),
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

func split(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
