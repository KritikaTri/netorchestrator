package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"netorchestrator/internal/platform"
)

var secret = []byte(strings.Repeat("g", 40))

func token(sub string) string {
	return "Bearer " + platform.SignHS256(secret, platform.Claims{Subject: sub, Expires: time.Now().Add(time.Hour).Unix()})
}

func backendCounting(hits *atomic.Int64, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		b, _ := io.ReadAll(r.Body)
		w.Header().Set("X-Seen-Subject", r.Header.Get("X-Authenticated-Subject"))
		w.WriteHeader(status)
		w.Write(b)
	}))
}

func newGW(t *testing.T, cfg Config) (*Gateway, *platform.Registry) {
	t.Helper()
	cfg.JWTSecret = secret
	if cfg.RatePerSecond == 0 {
		cfg.RatePerSecond, cfg.Burst = 1000, 1000
	}
	reg := platform.NewRegistry()
	g, err := New(cfg, reg)
	if err != nil {
		t.Fatal(err)
	}
	return g, reg
}

func do(g http.Handler, method, path, body, auth string, hdr ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	for i := 0; i+1 < len(hdr); i += 2 {
		req.Header.Set(hdr[i], hdr[i+1])
	}
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)
	return w
}

func TestRequiresValidAuthAndNeverReachesBackendOtherwise(t *testing.T) {
	var hits atomic.Int64
	be := backendCounting(&hits, 200)
	defer be.Close()
	g, reg := newGW(t, Config{Routes: []Route{{Prefix: "/v1/analyze/", Upstream: []string{be.URL}, Idempotent: true}}})

	expired := "Bearer " + platform.SignHS256(secret, platform.Claims{Subject: "u", Expires: time.Now().Add(-time.Hour).Unix()})
	for _, a := range []string{"", "Bearer junk", expired, "Basic abc"} {
		if w := do(g, "POST", "/v1/analyze/path", "{}", a); w.Code != 401 {
			t.Errorf("auth %q: got %d", a, w.Code)
		}
	}
	if w := do(g, "POST", "/v1/analyze/path", "{}", "", "X-API-Key", "forged-key"); w.Code != 401 {
		t.Errorf("forged api key: %d", w.Code)
	}
	if hits.Load() != 0 {
		t.Fatal("unauthenticated request reached a backend")
	}
	if reg.Counter("gateway_auth_failures_total") != 5 {
		t.Error("auth failures not counted")
	}
}

func TestJWTAndAPIKeyBothWork(t *testing.T) {
	var hits atomic.Int64
	be := backendCounting(&hits, 200)
	defer be.Close()
	sum := sha256.Sum256([]byte("my-key"))
	g, _ := newGW(t, Config{
		Routes:       []Route{{Prefix: "/v1/analyze/", Upstream: []string{be.URL}}},
		APIKeySHA256: []string{hex.EncodeToString(sum[:])},
	})
	w := do(g, "POST", "/v1/analyze/x", "hello", token("alice"))
	if w.Code != 200 || w.Body.String() != "hello" || w.Header().Get("X-Seen-Subject") != "user:alice" {
		t.Fatalf("jwt: %d %q %q", w.Code, w.Body, w.Header().Get("X-Seen-Subject"))
	}
	w = do(g, "POST", "/v1/analyze/x", "hi", "", "X-API-Key", "my-key")
	if w.Code != 200 || !strings.HasPrefix(w.Header().Get("X-Seen-Subject"), "key:") {
		t.Fatalf("api key: %d %q", w.Code, w.Header().Get("X-Seen-Subject"))
	}
}

func TestClientCannotSpoofIdentityHeader(t *testing.T) {
	var hits atomic.Int64
	be := backendCounting(&hits, 200)
	defer be.Close()
	g, _ := newGW(t, Config{Routes: []Route{{Prefix: "/v1/analyze/", Upstream: []string{be.URL}}}})
	w := do(g, "POST", "/v1/analyze/x", "", token("mallory"), "X-Authenticated-Subject", "user:admin")
	if got := w.Header().Get("X-Seen-Subject"); got != "user:mallory" {
		t.Fatalf("backend saw %q", got)
	}
}

func TestRoundRobinSpreadsLoad(t *testing.T) {
	var a, b, c atomic.Int64
	sa, sb, sc := backendCounting(&a, 200), backendCounting(&b, 200), backendCounting(&c, 200)
	defer sa.Close()
	defer sb.Close()
	defer sc.Close()
	g, _ := newGW(t, Config{Routes: []Route{{Prefix: "/v1/analyze/", Upstream: []string{sa.URL, sb.URL, sc.URL}, Idempotent: true}}})
	for i := 0; i < 30; i++ {
		if w := do(g, "POST", "/v1/analyze/x", "{}", token("u")); w.Code != 200 {
			t.Fatal(w.Code)
		}
	}
	if a.Load() != 10 || b.Load() != 10 || c.Load() != 10 {
		t.Fatalf("uneven: %d %d %d", a.Load(), b.Load(), c.Load())
	}
}

func TestIdempotentRequestFailsOverAndBreakerStopsHammeringDeadReplica(t *testing.T) {
	var good, bad atomic.Int64
	sg := backendCounting(&good, 200)
	defer sg.Close()
	sd := backendCounting(&bad, 503)
	defer sd.Close()
	g, reg := newGW(t, Config{
		Routes:       []Route{{Prefix: "/v1/analyze/", Upstream: []string{sd.URL, sg.URL}, Idempotent: true}},
		BreakerFails: 3, BreakerCool: time.Hour,
	})
	for i := 0; i < 20; i++ {
		if w := do(g, "POST", "/v1/analyze/x", "{}", token("u")); w.Code != 200 {
			t.Fatalf("request %d failed despite a healthy replica: %d", i, w.Code)
		}
	}
	if bad.Load() != 3 {
		t.Fatalf("dead replica hit %d times; breaker should cap it at 3", bad.Load())
	}
	if good.Load() != 20 || reg.Counter("gateway_upstream_total", "upstream", strings.TrimPrefix(sd.URL, "http://"), "result", "failure") != 3 {
		t.Fatalf("good=%d", good.Load())
	}
	if g.Ready() != nil {
		t.Error("gateway is ready while one healthy replica remains")
	}
}

func TestNonIdempotentRouteIsNeverRetried(t *testing.T) {
	var bad, good atomic.Int64
	sd, sg := backendCounting(&bad, 503), backendCounting(&good, 200)
	defer sd.Close()
	defer sg.Close()
	g, _ := newGW(t, Config{Routes: []Route{{Prefix: "/v1/changes", Upstream: []string{sd.URL, sg.URL}}}})
	w := do(g, "POST", "/v1/changes", "{}", token("u")) // first pick = the failing replica
	if w.Code != 503 || good.Load() != 0 || bad.Load() != 1 {
		t.Fatalf("code=%d good=%d bad=%d; a POST must not be replayed", w.Code, good.Load(), bad.Load())
	}
}

func TestAllReplicasDownGives503AndNotReady(t *testing.T) {
	dead := httptest.NewServer(http.NotFoundHandler())
	dead.Close() // connection refused
	g, _ := newGW(t, Config{Routes: []Route{{Prefix: "/v1/analyze/", Upstream: []string{dead.URL}, Idempotent: true}}, BreakerFails: 1, BreakerCool: time.Hour})
	w := do(g, "POST", "/v1/analyze/x", "{}", token("u"))
	if w.Code != 503 || w.Header().Get("Retry-After") == "" {
		t.Fatalf("%d", w.Code)
	}
	if g.Ready() == nil {
		t.Error("readiness must fail when every replica is circuit-open")
	}
}

func TestUpstreamClientErrorsPassThroughAndDontTripBreaker(t *testing.T) {
	var hits atomic.Int64
	be := backendCounting(&hits, 422)
	defer be.Close()
	g, _ := newGW(t, Config{Routes: []Route{{Prefix: "/v1/analyze/", Upstream: []string{be.URL}, Idempotent: true}}, BreakerFails: 2})
	for i := 0; i < 10; i++ {
		if w := do(g, "POST", "/v1/analyze/x", "{}", token("u")); w.Code != 422 {
			t.Fatalf("%d", w.Code)
		}
	}
	if g.Ready() != nil || hits.Load() != 10 {
		t.Fatal("4xx are the caller's fault and must not open the circuit")
	}
}

func TestRateLimitPerSubject(t *testing.T) {
	var hits atomic.Int64
	be := backendCounting(&hits, 200)
	defer be.Close()
	g, reg := newGW(t, Config{Routes: []Route{{Prefix: "/v1/analyze/", Upstream: []string{be.URL}}}, RatePerSecond: 0.001, Burst: 2})
	codes := []int{}
	for i := 0; i < 4; i++ {
		codes = append(codes, do(g, "POST", "/v1/analyze/x", "{}", token("noisy")).Code)
	}
	if codes[0] != 200 || codes[1] != 200 || codes[2] != 429 || codes[3] != 429 {
		t.Fatalf("%v", codes)
	}
	if do(g, "POST", "/v1/analyze/x", "{}", token("quiet")).Code != 200 {
		t.Error("one noisy client must not starve another")
	}
	if reg.Counter("gateway_rate_limited_total") != 2 {
		t.Error("limit not counted")
	}
}

func TestRoutingBodyLimitAndConfigValidation(t *testing.T) {
	var hits atomic.Int64
	be := backendCounting(&hits, 200)
	defer be.Close()
	g, _ := newGW(t, Config{Routes: []Route{{Prefix: "/v1/analyze/", Upstream: []string{be.URL}}}, MaxBodyBytes: 8})
	if do(g, "GET", "/v1/other", "", token("u")).Code != 404 {
		t.Error("unknown prefix must 404")
	}
	if do(g, "POST", "/v1/analyze/x", strings.Repeat("x", 100), token("u")).Code != 413 {
		t.Error("oversize body must 413")
	}
	for name, cfg := range map[string]Config{
		"short secret": {JWTSecret: []byte("short")},
		"no upstreams": {JWTSecret: secret, Routes: []Route{{Prefix: "/x"}}},
		"bad url":      {JWTSecret: secret, Routes: []Route{{Prefix: "/x", Upstream: []string{"not a url"}}}},
		"bad digest":   {JWTSecret: secret, APIKeySHA256: []string{"zz"}, Routes: []Route{{Prefix: "/x", Upstream: []string{"http://a:1"}}}},
	} {
		if _, err := New(cfg, nil); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}
