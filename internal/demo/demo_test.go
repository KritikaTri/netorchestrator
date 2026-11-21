package demo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"netorchestrator/internal/ratelimit"
)

func server(rps float64, burst int) *http.ServeMux {
	mux := http.NewServeMux()
	New(ratelimit.NewTokenBucket(rps, burst, 0, nil)).Register(mux)
	return mux
}

func get(mux http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestIndexServedOnlyAtRoot(t *testing.T) {
	mux := server(100, 100)
	rec := get(mux, "/")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "NetOrchestrator") {
		t.Fatalf("root: %d", rec.Code)
	}
	if csp := rec.Header().Get("Content-Security-Policy"); !strings.Contains(csp, "default-src 'none'") {
		t.Fatalf("missing CSP: %q", csp)
	}
	if rec := get(mux, "/nope"); rec.Code != 404 {
		t.Fatalf("unknown path should 404, got %d", rec.Code)
	}
}

func report(t *testing.T, mux http.Handler, id string) map[string]any {
	t.Helper()
	rec := get(mux, "/demo/preflight?remove="+id)
	if rec.Code != 200 {
		t.Fatalf("%s: %d %s", id, rec.Code, rec.Body)
	}
	var out struct{ Report map[string]any }
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out.Report
}

func TestPreflightResults(t *testing.T) {
	mux := server(100, 100)
	if r := report(t, mux, "core1"); r["risk"] != "none" {
		t.Fatalf("redundant core1 should be risk none, got %v", r["risk"])
	}
	r := report(t, mux, "dist2")
	if r["risk"] == "none" {
		t.Fatalf("dist2 carries access2/access3/printer-vlan, should not be risk none: %v", r)
	}
	if cut, _ := r["cut_off"].([]any); len(cut) != 3 {
		t.Fatalf("dist2 should cut off 3 devices, got %v", r["cut_off"])
	}
}

func TestPreflightRejectsUnknownOrMissingDevice(t *testing.T) {
	mux := server(100, 100)
	for _, p := range []string{"/demo/preflight", "/demo/preflight?remove=", "/demo/preflight?remove=evil", "/demo/preflight?remove=core1%27--"} {
		if rec := get(mux, p); rec.Code != 400 {
			t.Fatalf("%s: want 400, got %d", p, rec.Code)
		}
	}
}

func TestTopologyEndpoint(t *testing.T) {
	rec := get(server(100, 100), "/demo/topology")
	var out struct{ Nodes, Links []any }
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &out) != nil || len(out.Nodes) != 8 || len(out.Links) != 9 {
		t.Fatalf("topology: %d %s", rec.Code, rec.Body)
	}
}

func TestDemoIsRateLimited(t *testing.T) {
	mux := server(0.001, 2)
	for i := 0; i < 2; i++ {
		if rec := get(mux, "/demo/topology"); rec.Code != 200 {
			t.Fatalf("request %d: %d", i, rec.Code)
		}
	}
	rec := get(mux, "/demo/preflight?remove=core1")
	if rec.Code != 429 || rec.Header().Get("Retry-After") == "" {
		t.Fatalf("want 429 with Retry-After, got %d %q", rec.Code, rec.Header().Get("Retry-After"))
	}
}
