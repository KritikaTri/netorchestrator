// Command change-svc is the change-management service: preflight-gated change
// requests, four-eyes approval workflow and a tamper-evident audit log.
package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"netorchestrator/internal/changesvc"
	"netorchestrator/internal/platform"
)

func main() {
	log := platform.Logger("change-svc")
	reg := platform.NewRegistry()

	el, err := changesvc.OpenLog(platform.Env("CHANGE_LOG_PATH", "./change-audit.jsonl"), platform.Env("CHANGE_LOG_FSYNC", "1") == "1")
	if err != nil {
		log.Error("audit log failed integrity check or could not be opened; refusing to start", "err", err)
		os.Exit(1)
	}
	defer el.Close()

	var analyzer changesvc.Analyzer = changesvc.LocalAnalyzer{}
	if u := platform.Env("CHANGE_TOPOLOGY_URL", ""); u != "" {
		analyzer = changesvc.NewHTTPAnalyzer(u, 5*time.Second)
	}
	svc, err := changesvc.NewService(el, analyzer)
	if err != nil {
		log.Error("could not rebuild state from audit log", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	svc.Routes(mux)
	platform.Health(mux, reg, func() error { return el.Verify() })

	srv := &http.Server{
		Addr: platform.Env("ADDR", ":8082"),
		Handler: platform.Chain(mux, platform.WithRequestID(),
			platform.WithObservability("change-svc", reg, log, func(r *http.Request) string {
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
