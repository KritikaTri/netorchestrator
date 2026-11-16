// Package gateway is the single public entry point: it authenticates callers,
// rate-limits them, and load-balances requests across service replicas with
// per-replica circuit breakers and a bounded retry for idempotent calls.
package gateway

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"netorchestrator/internal/platform"
	"netorchestrator/internal/ratelimit"
)

// Route maps a path prefix to a pool of replicas.
type Route struct {
	Prefix   string
	Upstream []string // base URLs, e.g. http://topology-1:8081
	// Idempotent routes may be retried on another replica after a transport
	// error or 502/503/504. Non-idempotent routes are never retried.
	Idempotent bool
	// Limiter, if set, is applied per subject on this route in addition to the
	// gateway-wide limit (e.g. a stricter budget for state-changing calls).
	Limiter ratelimit.Limiter
}

// Config for the gateway.
type Config struct {
	Routes       []Route
	JWTSecret    []byte
	APIKeySHA256 []string // hex SHA-256 digests of accepted API keys
	// Limiter is the gateway-wide per-subject limiter. If nil, a token bucket
	// is built from RatePerSecond/Burst (both must then be set).
	Limiter        ratelimit.Limiter
	RatePerSecond  float64
	Burst          int
	MaxBodyBytes   int64
	UpstreamTimout time.Duration
	BreakerFails   int
	BreakerCool    time.Duration
}

type backend struct {
	base    *url.URL
	breaker *platform.Breaker
}

type pool struct {
	route    Route
	backends []*backend
	next     atomic.Uint64
}

// Gateway is an http.Handler.
type Gateway struct {
	cfg     Config
	pools   []*pool
	limiter ratelimit.Limiter
	client  *http.Client
	reg     *platform.Registry
	now     func() time.Time
	keys    [][]byte
}

// New validates cfg and builds a Gateway.
func New(cfg Config, reg *platform.Registry) (*Gateway, error) {
	if len(cfg.JWTSecret) < 32 {
		return nil, errors.New("gateway: JWT secret must be at least 32 bytes")
	}
	if reg == nil {
		reg = platform.NewRegistry()
	}
	if cfg.UpstreamTimout == 0 {
		cfg.UpstreamTimout = 10 * time.Second
	}
	if cfg.MaxBodyBytes == 0 {
		cfg.MaxBodyBytes = 8 << 20
	}
	if cfg.BreakerFails == 0 {
		cfg.BreakerFails = 5
	}
	if cfg.BreakerCool == 0 {
		cfg.BreakerCool = 10 * time.Second
	}
	g := &Gateway{
		cfg:     cfg,
		limiter: cfg.Limiter,
		client: &http.Client{
			Transport: &http.Transport{MaxIdleConnsPerHost: 256, IdleConnTimeout: 90 * time.Second},
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		reg: reg,
		now: time.Now,
	}
	if g.limiter == nil {
		if cfg.RatePerSecond <= 0 || cfg.Burst < 1 {
			return nil, errors.New("gateway: configure Limiter, or RatePerSecond and Burst")
		}
		g.limiter = ratelimit.NewTokenBucket(cfg.RatePerSecond, cfg.Burst, 0, nil)
	}
	for _, d := range cfg.APIKeySHA256 {
		b, err := hex.DecodeString(strings.TrimSpace(d))
		if err != nil || len(b) != sha256.Size {
			return nil, errors.New("gateway: API key digests must be hex SHA-256")
		}
		g.keys = append(g.keys, b)
	}
	for _, r := range cfg.Routes {
		p := &pool{route: r}
		for _, raw := range r.Upstream {
			u, err := url.Parse(strings.TrimSpace(raw))
			if err != nil || u.Scheme == "" || u.Host == "" {
				return nil, errors.New("gateway: bad upstream URL " + raw)
			}
			p.backends = append(p.backends, &backend{base: u, breaker: platform.NewBreaker(cfg.BreakerFails, cfg.BreakerCool, nil)})
		}
		if len(p.backends) == 0 {
			return nil, errors.New("gateway: route " + r.Prefix + " has no upstreams")
		}
		g.pools = append(g.pools, p)
	}
	return g, nil
}

// Ready reports whether at least one backend per route is not circuit-open.
func (g *Gateway) Ready() error {
	for _, p := range g.pools {
		ok := false
		for _, b := range p.backends {
			if b.breaker.State() != platform.Open {
				ok = true
			}
		}
		if !ok {
			return errors.New("no healthy upstream for " + p.route.Prefix)
		}
	}
	return nil
}

func (g *Gateway) authenticate(r *http.Request) (subject string, ok bool) {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		c, err := platform.VerifyHS256(g.cfg.JWTSecret, strings.TrimPrefix(h, "Bearer "), g.now())
		if err != nil {
			return "", false
		}
		return "user:" + c.Subject, true
	}
	if k := r.Header.Get("X-API-Key"); k != "" {
		sum := sha256.Sum256([]byte(k))
		for _, want := range g.keys {
			if subtle.ConstantTimeCompare(sum[:], want) == 1 {
				return "key:" + hex.EncodeToString(sum[:4]), true
			}
		}
	}
	return "", false
}

// setRateHeaders sets the IETF draft RateLimit-* headers so well-behaved
// clients can pace themselves before they ever see a 429.
func setRateHeaders(w http.ResponseWriter, d ratelimit.Decision) {
	h := w.Header()
	h.Set("RateLimit-Limit", strconv.Itoa(d.Limit))
	h.Set("RateLimit-Remaining", strconv.Itoa(d.Remaining))
	h.Set("RateLimit-Reset", strconv.Itoa(ceilSeconds(d.Reset)))
}

func denyRateLimited(w http.ResponseWriter, d ratelimit.Decision) {
	setRateHeaders(w, d)
	w.Header().Set("Retry-After", strconv.Itoa(ceilSeconds(d.RetryAfter)))
	writeErr(w, http.StatusTooManyRequests, "rate limit exceeded")
}

func ceilSeconds(d time.Duration) int {
	s := int((d + time.Second - 1) / time.Second)
	if s < 1 {
		return 1
	}
	return s
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}

// ServeHTTP implements http.Handler.
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var p *pool
	for _, c := range g.pools {
		if strings.HasPrefix(r.URL.Path, c.route.Prefix) {
			p = c
			break
		}
	}
	if p == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	subject, ok := g.authenticate(r)
	if !ok {
		g.reg.Inc("gateway_auth_failures_total", 1)
		w.Header().Set("WWW-Authenticate", `Bearer realm="netorchestrator"`)
		writeErr(w, http.StatusUnauthorized, "authentication required")
		return
	}
	dec := g.limiter.Allow(subject)
	if !dec.Allowed {
		g.reg.Inc("gateway_rate_limited_total", 1)
		denyRateLimited(w, dec)
		return
	}
	if rl := p.route.Limiter; rl != nil {
		rd := rl.Allow(subject)
		if !rd.Allowed {
			g.reg.Inc("gateway_route_rate_limited_total", 1, "route", p.route.Prefix)
			denyRateLimited(w, rd)
			return
		}
		if rd.Remaining < dec.Remaining { // advertise whichever budget is closer to running out
			dec = rd
		}
	}
	setRateHeaders(w, dec)
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, g.cfg.MaxBodyBytes))
	if err != nil {
		writeErr(w, http.StatusRequestEntityTooLarge, "request too large")
		return
	}
	g.forward(w, r, p, subject, body)
}

func (g *Gateway) forward(w http.ResponseWriter, r *http.Request, p *pool, subject string, body []byte) {
	attempts := 1
	if p.route.Idempotent {
		attempts = 2
	}
	if attempts > len(p.backends) {
		attempts = len(p.backends)
	}
	start := int(p.next.Add(1) - 1)
	tried := 0
	for i := 0; i < len(p.backends) && tried < attempts; i++ {
		b := p.backends[(start+i)%len(p.backends)]
		if b.breaker.Allow() != nil {
			continue
		}
		tried++
		resp, err := g.call(r, b, subject, body)
		failed := err != nil || resp.StatusCode == http.StatusBadGateway ||
			resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusGatewayTimeout
		b.breaker.Done(!failed)
		up := b.base.Host
		if failed {
			g.reg.Inc("gateway_upstream_total", 1, "upstream", up, "result", "failure")
			if resp != nil {
				resp.Body.Close()
			}
			if r.Context().Err() != nil {
				return
			}
			continue
		}
		g.reg.Inc("gateway_upstream_total", 1, "upstream", up, "result", "ok")
		defer resp.Body.Close()
		for k, vs := range resp.Header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.Header().Set("X-Upstream", up)
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
		return
	}
	g.reg.Inc("gateway_unavailable_total", 1)
	w.Header().Set("Retry-After", "2")
	writeErr(w, http.StatusServiceUnavailable, "upstream unavailable")
}

func (g *Gateway) call(r *http.Request, b *backend, subject string, body []byte) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(r.Context(), g.cfg.UpstreamTimout)
	target := *b.base
	target.Path = strings.TrimRight(b.base.Path, "/") + r.URL.Path
	target.RawQuery = r.URL.RawQuery
	req, err := http.NewRequestWithContext(ctx, r.Method, target.String(), bytes.NewReader(body))
	if err != nil {
		cancel()
		return nil, err
	}
	if ct := r.Header.Get("Content-Type"); ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	// Identity is asserted by the gateway only; client-supplied values never pass through.
	req.Header.Set("X-Authenticated-Subject", subject)
	req.Header.Set(platform.HeaderRequestID, r.Header.Get(platform.HeaderRequestID))
	resp, err := g.client.Do(req)
	if err != nil {
		cancel()
		return nil, err
	}
	resp.Body = &cancelOnClose{resp.Body, cancel}
	return resp, nil
}

type cancelOnClose struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelOnClose) Close() error { err := c.ReadCloser.Close(); c.cancel(); return err }
