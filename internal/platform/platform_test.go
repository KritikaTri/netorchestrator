package platform

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time          { return c.t }
func (c *clock) advance(d time.Duration) { c.t = c.t.Add(d) }

func TestBreakerLifecycle(t *testing.T) {
	c := &clock{time.Unix(1000, 0)}
	b := NewBreaker(3, 10*time.Second, c.now)
	for i := 0; i < 3; i++ {
		if b.Allow() != nil {
			t.Fatal("closed breaker must allow")
		}
		b.Done(false)
	}
	if b.State() != Open || b.Allow() != ErrCircuitOpen {
		t.Fatal("should be open after threshold")
	}
	c.advance(11 * time.Second)
	if b.Allow() != nil {
		t.Fatal("cooled breaker must allow one probe")
	}
	if b.Allow() != ErrCircuitOpen {
		t.Fatal("only one probe at a time")
	}
	b.Done(false) // probe fails -> re-open
	if b.State() != Open {
		t.Fatal("failed probe re-opens")
	}
	c.advance(11 * time.Second)
	b.Allow()
	b.Done(true)
	if b.State() != Closed || b.Allow() != nil {
		t.Fatal("successful probe closes")
	}
}

func TestBreakerSuccessResetsFailureStreak(t *testing.T) {
	b := NewBreaker(2, time.Second, nil)
	b.Allow()
	b.Done(false)
	b.Allow()
	b.Done(true)
	b.Allow()
	b.Done(false)
	if b.State() != Closed {
		t.Fatal("non-consecutive failures must not open")
	}
}

func TestJWT(t *testing.T) {
	sec := []byte(strings.Repeat("s", 32))
	now := time.Unix(1_000_000, 0)
	good := SignHS256(sec, Claims{Subject: "u1", Roles: []string{"admin"}, Expires: now.Add(time.Hour).Unix()})
	c, err := VerifyHS256(sec, good, now)
	if err != nil || c.Subject != "u1" || c.Roles[0] != "admin" {
		t.Fatalf("%v %+v", err, c)
	}
	if _, err := VerifyHS256(sec, good, now.Add(2*time.Hour)); err == nil {
		t.Error("expired token accepted")
	}
	if _, err := VerifyHS256([]byte("other-secret-other-secret-other-"), good, now); err == nil {
		t.Error("wrong key accepted")
	}
	// alg=none forgery
	none := b64.EncodeToString([]byte(`{"alg":"none"}`)) + "." + b64.EncodeToString([]byte(`{"sub":"x","exp":9999999999}`)) + "."
	if _, err := VerifyHS256(sec, none, now); err == nil {
		t.Error("alg none accepted")
	}
	// tampered payload keeps old signature
	parts := strings.Split(good, ".")
	forged := parts[0] + "." + b64.EncodeToString([]byte(`{"sub":"admin","exp":9999999999}`)) + "." + parts[2]
	if _, err := VerifyHS256(sec, forged, now); err == nil {
		t.Error("tampered payload accepted")
	}
	for _, junk := range []string{"", "a.b", "a.b.c.d", "!!.!!.!!"} {
		if _, err := VerifyHS256(sec, junk, now); err == nil {
			t.Errorf("junk %q accepted", junk)
		}
	}
	noExp := SignHS256(sec, Claims{Subject: "u"})
	if _, err := VerifyHS256(sec, noExp, now); err == nil {
		t.Error("token without exp accepted")
	}
}

func TestMetricsExposeAndQuantile(t *testing.T) {
	r := NewRegistry()
	r.Inc("reqs", 2, "code", "2xx")
	r.Inc("reqs", 1, "code", "5xx")
	for i := 0; i < 99; i++ {
		r.Observe("lat", 0.002, "svc", "a")
	}
	r.Observe("lat", 3, "svc", "a")
	var buf bytes.Buffer
	r.Expose(&buf)
	out := buf.String()
	for _, want := range []string{`reqs{code="2xx"} 2`, `reqs{code="5xx"} 1`, `lat_count{svc="a"} 100`, `lat_bucket{svc="a",le="+Inf"} 100`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if q := r.Quantile("lat", 0.5, "svc", "a"); q != 0.0025 {
		t.Errorf("p50=%v", q)
	}
	if q := r.Quantile("lat", 1, "svc", "a"); q < 2.5 {
		t.Errorf("p100=%v", q)
	}
	if r.Counter("reqs", "code", "2xx") != 2 || r.Counter("nope") != 0 {
		t.Error("counter readback")
	}
}

func TestMiddlewareRequestIDMetricsAndPanicRecovery(t *testing.T) {
	reg := NewRegistry()
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ok", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(RequestID(r.Context()))) })
	mux.HandleFunc("GET /boom", func(http.ResponseWriter, *http.Request) { panic("x") })
	h := Chain(mux, WithRequestID(), WithObservability("t", reg, log, func(r *http.Request) string { return r.URL.Path }))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/ok", nil))
	id := w.Header().Get(HeaderRequestID)
	if id == "" || w.Body.String() != id {
		t.Fatalf("id %q body %q", id, w.Body)
	}

	req := httptest.NewRequest("GET", "/ok", nil)
	req.Header.Set(HeaderRequestID, "abc-123")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Header().Get(HeaderRequestID) != "abc-123" {
		t.Error("inbound request id must be preserved")
	}
	req.Header.Set(HeaderRequestID, "bad id\r\ninjected")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if strings.ContainsAny(w.Header().Get(HeaderRequestID), " \r\n") {
		t.Error("unsafe request id must be replaced")
	}

	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/boom", nil))
	if w.Code != 500 || reg.Counter("http_requests_total", "service", "t", "route", "/boom", "code", "5xx") != 1 {
		t.Fatalf("panic not recovered/recorded: %d", w.Code)
	}
	if reg.Counter("http_requests_total", "service", "t", "route", "/ok", "code", "2xx") != 3 {
		t.Error("ok requests not counted")
	}
}

func TestHealthEndpoints(t *testing.T) {
	mux := http.NewServeMux()
	ready := error(nil)
	Health(mux, NewRegistry(), func() error { return ready })
	get := func(p string) int {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", p, nil))
		return w.Code
	}
	if get("/healthz") != 200 || get("/readyz") != 200 || get("/metrics") != 200 {
		t.Fatal("endpoints should be up")
	}
	ready = io.ErrClosedPipe
	if get("/readyz") != 503 || get("/healthz") != 200 {
		t.Fatal("readiness must fail independently of liveness")
	}
}
