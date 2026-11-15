package topologysvc

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"netorchestrator/internal/platform"
)

const star = `"nodes":[{"id":"h","name":"hub"},{"id":"a"},{"id":"b"}],
"links":[{"id":"1","source":"h","target":"a"},{"id":"2","source":"h","target":"b"}]`

func newTS(t *testing.T, cfg Config) (*httptest.Server, *platform.Registry) {
	t.Helper()
	reg := platform.NewRegistry()
	mux := http.NewServeMux()
	New(cfg, reg).Routes(mux)
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts, reg
}

func post(t *testing.T, ts *httptest.Server, path, body string) (*http.Response, string) {
	t.Helper()
	resp, err := http.Post(ts.URL+path, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp, string(b)
}

func TestPreflightEndpoint(t *testing.T) {
	ts, _ := newTS(t, DefaultConfig())
	resp, body := post(t, ts, "/v1/analyze/preflight", `{`+star+`,"change":{"kind":"remove_node","id":"h"}}`)
	if resp.StatusCode != 200 {
		t.Fatalf("%d %s", resp.StatusCode, body)
	}
	var out struct {
		Report struct {
			Risk   string `json:"risk"`
			Target string `json:"target"`
		} `json:"report"`
	}
	if err := json.Unmarshal([]byte(body), &out); err != nil || out.Report.Target != "hub" || out.Report.Risk != "critical" {
		t.Fatalf("%s", body)
	}
}

func TestResilienceAndPath(t *testing.T) {
	ts, _ := newTS(t, DefaultConfig())
	if resp, body := post(t, ts, "/v1/analyze/resilience", `{`+star+`}`); resp.StatusCode != 200 || !strings.Contains(body, `"redundant":false`) {
		t.Fatalf("%d %s", resp.StatusCode, body)
	}
	resp, body := post(t, ts, "/v1/analyze/path", `{`+star+`,"src":"a","dst":"b"}`)
	if resp.StatusCode != 200 || !strings.Contains(body, `"cost":2`) {
		t.Fatalf("%d %s", resp.StatusCode, body)
	}
	if resp, _ := post(t, ts, "/v1/analyze/path", `{`+star+`,"src":"a","dst":"zzz"}`); resp.StatusCode != 404 {
		t.Errorf("unknown node: %d", resp.StatusCode)
	}
	split := `{"nodes":[{"id":"x"},{"id":"y"}],"links":[],"src":"x","dst":"y"}`
	if resp, _ := post(t, ts, "/v1/analyze/path", split); resp.StatusCode != 404 {
		t.Errorf("no path: %d", resp.StatusCode)
	}
}

func TestValidationErrors(t *testing.T) {
	ts, _ := newTS(t, DefaultConfig())
	cases := []struct {
		name, path, body string
		want             int
	}{
		{"bad json", "/v1/analyze/preflight", `nope`, 400},
		{"unknown field", "/v1/analyze/resilience", `{"nodes":[],"evil":1}`, 400},
		{"bad change kind", "/v1/analyze/preflight", `{` + star + `,"change":{"kind":"explode","id":"h"}}`, 400},
		{"unknown target", "/v1/analyze/preflight", `{` + star + `,"change":{"kind":"remove_node","id":"nope"}}`, 404},
		{"dangling link", "/v1/analyze/resilience", `{"nodes":[{"id":"a"}],"links":[{"id":"l","source":"a","target":"ghost"}]}`, 422},
	}
	for _, c := range cases {
		if resp, body := post(t, ts, c.path, c.body); resp.StatusCode != c.want {
			t.Errorf("%s: got %d want %d (%s)", c.name, resp.StatusCode, c.want, body)
		}
	}
}

func TestCacheHitIsIdenticalAndCounted(t *testing.T) {
	ts, reg := newTS(t, DefaultConfig())
	req := `{` + star + `}`
	r1, b1 := post(t, ts, "/v1/analyze/resilience", req)
	r2, b2 := post(t, ts, "/v1/analyze/resilience", req)
	if r1.Header.Get("X-Cache") != "MISS" || r2.Header.Get("X-Cache") != "HIT" || b1 != b2 {
		t.Fatalf("cache: %q %q equal=%v", r1.Header.Get("X-Cache"), r2.Header.Get("X-Cache"), b1 == b2)
	}
	// same body to a different endpoint must not collide
	r3, _ := post(t, ts, "/v1/analyze/preflight", `{`+star+`,"change":{"kind":"remove_node","id":"h"}}`)
	if r3.Header.Get("X-Cache") != "MISS" {
		t.Error("different endpoint must not share cache entries")
	}
	if reg.Counter("topology_cache_total", "result", "hit") != 1 {
		t.Error("hit not counted")
	}
}

func TestErrorsAreNotCached(t *testing.T) {
	ts, _ := newTS(t, DefaultConfig())
	bad := `{` + star + `,"change":{"kind":"remove_node","id":"nope"}}`
	post(t, ts, "/v1/analyze/preflight", bad)
	if r, _ := post(t, ts, "/v1/analyze/preflight", bad); r.Header.Get("X-Cache") == "HIT" {
		t.Fatal("error responses must not be cached")
	}
}

func TestBodyTooLarge(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxBodyBytes = 64
	ts, _ := newTS(t, cfg)
	if resp, _ := post(t, ts, "/v1/analyze/resilience", `{`+star+`}`); resp.StatusCode != 413 {
		t.Fatalf("got %d", resp.StatusCode)
	}
}

func TestLoadSheddingReturns429WhenSaturated(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxInFlight = 1
	ts, reg := newTS(t, cfg)

	// A raw TCP client promises a 1 MiB body but sends one byte, so the handler
	// holds the only slot while blocked reading. (net/http clients buffer small
	// writes, which made this nondeterministic.)
	conn, err := net.Dial("tcp", strings.TrimPrefix(ts.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	io.WriteString(conn, "POST /v1/analyze/resilience HTTP/1.1\r\nHost: x\r\nContent-Type: application/json\r\nContent-Length: 1048576\r\n\r\n{")

	waitFor := func(what string, cond func() bool) {
		t.Helper()
		for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
			if cond() {
				return
			}
		}
		t.Fatalf("timed out waiting for %s", what)
	}
	waitFor("429 while slot is held", func() bool {
		resp, _ := post(t, ts, "/v1/analyze/resilience", `{}`)
		if resp.StatusCode == 429 && resp.Header.Get("Retry-After") == "" {
			t.Error("429 must carry Retry-After")
		}
		return resp.StatusCode == 429
	})
	if reg.Counter("topology_shed_total") == 0 {
		t.Error("shed not counted")
	}

	conn.Close() // client disconnects; handler's read fails and the slot is released
	waitFor("recovery after saturation", func() bool {
		resp, _ := post(t, ts, "/v1/analyze/resilience", `{`+star+`}`)
		return resp.StatusCode == 200
	})
}

func TestLRUEvictsOldest(t *testing.T) {
	c := newLRU(2)
	c.Put("a", []byte("1"))
	c.Put("b", []byte("2"))
	c.Get("a")
	c.Put("c", []byte("3"))
	if _, ok := c.Get("b"); ok {
		t.Error("b should be evicted")
	}
	if _, ok := c.Get("a"); !ok || c.Len() != 2 {
		t.Error("a must survive")
	}
	c.Put("a", []byte("9"))
	if v, _ := c.Get("a"); string(v) != "9" {
		t.Error("update")
	}
	if _, ok := newLRU(0).Get("x"); ok {
		t.Error("zero-cap cache stores nothing")
	}
}
