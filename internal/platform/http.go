package platform

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type ctxKey int

const requestIDKey ctxKey = 1

// HeaderRequestID is propagated across every service hop for tracing.
const HeaderRequestID = "X-Request-ID"

// RequestID returns the request ID stored in ctx ("" if none).
func RequestID(ctx context.Context) string {
	s, _ := ctx.Value(requestIDKey).(string)
	return s
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Middleware wraps a handler.
type Middleware func(http.Handler) http.Handler

// Chain applies middleware so the first listed is outermost.
func Chain(h http.Handler, m ...Middleware) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

// WithRequestID accepts an inbound ID (sanitised) or mints one.
func WithRequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(HeaderRequestID)
			if len(id) == 0 || len(id) > 64 || strings.ContainsAny(id, " \r\n\t") {
				id = newID()
			}
			w.Header().Set(HeaderRequestID, id)
			r.Header.Set(HeaderRequestID, id)
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusWriter) WriteHeader(c int) { s.status = c; s.ResponseWriter.WriteHeader(c) }
func (s *statusWriter) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (s *statusWriter) Unwrap() http.ResponseWriter { return s.ResponseWriter }

// WithObservability records request count, latency histogram, in-flight gauge
// and a structured access log line. route is a low-cardinality label function.
func WithObservability(service string, reg *Registry, log *slog.Logger, route func(*http.Request) string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reg.GaugeAdd("http_in_flight", 1, "service", service)
			defer reg.GaugeAdd("http_in_flight", -1, "service", service)
			sw := &statusWriter{ResponseWriter: w}
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic", "request_id", RequestID(r.Context()), "panic", rec)
					if sw.status == 0 {
						http.Error(sw, `{"error":"internal error"}`, http.StatusInternalServerError)
					}
				}
				if sw.status == 0 {
					sw.status = http.StatusOK
				}
				d := time.Since(start)
				rt := route(r)
				reg.Inc("http_requests_total", 1, "service", service, "route", rt, "code", classOf(sw.status))
				reg.Observe("http_request_duration_seconds", d.Seconds(), "service", service, "route", rt)
				log.Info("request", "request_id", RequestID(r.Context()), "method", r.Method, "path", r.URL.Path,
					"status", sw.status, "bytes", sw.bytes, "ms", float64(d.Microseconds())/1000, "remote", clientIP(r))
			}()
			next.ServeHTTP(sw, r)
		})
	}
}

func classOf(code int) string { return string(rune('0'+code/100)) + "xx" }

func clientIP(r *http.Request) string {
	h, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return h
}

// ClientIP is exported for rate-limit keys.
func ClientIP(r *http.Request) string { return clientIP(r) }

// Health registers /healthz (liveness) and /readyz (readiness) plus /metrics.
func Health(mux *http.ServeMux, reg *Registry, ready func() error) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		if ready != nil {
			if err := ready(); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}
		w.Write([]byte("ready"))
	})
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		reg.Expose(w)
	})
}

// Logger builds a JSON slog logger tagged with the service name.
func Logger(service string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, nil)).With("service", service)
}

// Serve runs srv until SIGINT/SIGTERM (or ctx is cancelled), then drains
// in-flight requests for up to grace before returning.
func Serve(ctx context.Context, srv *http.Server, grace time.Duration, log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	log.Info("listening", "addr", srv.Addr)
	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	log.Info("shutting down", "grace", grace.String())
	sctx, cancel := context.WithTimeout(context.Background(), grace)
	defer cancel()
	if err := srv.Shutdown(sctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Env returns the environment variable or a default.
func Env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
