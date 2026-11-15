package changesvc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"netorchestrator/internal/preflight"
)

var (
	starNodes = []preflight.Node{{ID: "h", Name: "hub"}, {ID: "a"}, {ID: "b"}}
	starLinks = []preflight.Link{{ID: "1", Source: "h", Target: "a"}, {ID: "2", Source: "h", Target: "b"}}
	ringNodes = []preflight.Node{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	ringLinks = []preflight.Link{{ID: "1", Source: "a", Target: "b"}, {ID: "2", Source: "b", Target: "c"}, {ID: "3", Source: "c", Target: "a"}}
)

func riskyReq() SubmitRequest { // removing the hub is critical
	return SubmitRequest{Title: "decommission hub", Change: preflight.Change{Kind: preflight.RemoveNode, ID: "h"}, Nodes: starNodes, Links: starLinks}
}
func safeReq() SubmitRequest { // removing a ring node is harmless
	return SubmitRequest{Title: "maintenance on a", Change: preflight.Change{Kind: preflight.RemoveNode, ID: "a"}, Nodes: ringNodes, Links: ringLinks}
}

func newSvc(t *testing.T) *Service {
	t.Helper()
	s, err := NewService(NewMemLog(), nil)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSubmitStoresPreflightAndPolicyScalesWithRisk(t *testing.T) {
	s := newSvc(t)
	c, err := s.Submit(context.Background(), "alice", riskyReq())
	if err != nil {
		t.Fatal(err)
	}
	if c.Risk != preflight.RiskCritical || c.Required != 2 || c.Status != PendingApproval || len(c.Report.CutOff) != 2 || c.TopologyHash == "" {
		t.Fatalf("%+v", c)
	}
	safe, _ := s.Submit(context.Background(), "alice", safeReq())
	if safe.Risk != preflight.RiskNone || safe.Required != 1 {
		t.Fatalf("%+v", safe)
	}
}

func TestFourEyesAndCriticalNeedsTwoDistinctApprovers(t *testing.T) {
	s := newSvc(t)
	c, _ := s.Submit(context.Background(), "alice", riskyReq())
	if _, err := s.Approve("alice", c.ID); !errors.Is(err, ErrSelfApproval) {
		t.Fatalf("self approval: %v", err)
	}
	if got, err := s.Approve("bob", c.ID); err != nil || got.Status != PendingApproval {
		t.Fatalf("first approval must not be enough: %v %v", err, got)
	}
	if _, err := s.Approve("bob", c.ID); !errors.Is(err, ErrDuplicateApproval) {
		t.Fatalf("same person twice: %v", err)
	}
	got, err := s.Approve("carol", c.ID)
	if err != nil || got.Status != Approved || len(got.Approvers) != 2 {
		t.Fatalf("%v %+v", err, got)
	}
	if _, err := s.Approve("dave", c.ID); !errors.Is(err, ErrInvalidState) {
		t.Errorf("approving an approved change: %v", err)
	}
}

func TestLifecycleRules(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()
	c, _ := s.Submit(ctx, "alice", safeReq())

	if _, err := s.Apply("alice", c.ID); !errors.Is(err, ErrInvalidState) {
		t.Error("cannot apply before approval")
	}
	if _, err := s.Cancel("bob", c.ID); !errors.Is(err, ErrForbidden) {
		t.Error("only requester cancels")
	}
	if _, err := s.Reject("alice", c.ID, "x"); !errors.Is(err, ErrSelfApproval) {
		t.Error("requester cannot reject own change")
	}
	s.Approve("bob", c.ID)
	if a, err := s.Apply("bob", c.ID); err != nil || a.Status != Applied {
		t.Fatalf("%v %+v", err, a)
	}
	if _, err := s.Apply("bob", c.ID); !errors.Is(err, ErrInvalidState) {
		t.Error("double apply")
	}
	if _, err := s.Cancel("alice", c.ID); !errors.Is(err, ErrInvalidState) {
		t.Error("cannot cancel after apply")
	}

	r, _ := s.Submit(ctx, "alice", safeReq())
	if got, err := s.Reject("bob", r.ID, "needs window"); err != nil || got.Status != Rejected || got.Reason != "needs window" {
		t.Fatalf("%v %+v", err, got)
	}
	if _, err := s.Approve("carol", r.ID); !errors.Is(err, ErrInvalidState) {
		t.Error("approving a rejected change")
	}
	if _, err := s.Get("nope"); !errors.Is(err, ErrNotFound) {
		t.Error("unknown id")
	}
	if len(s.List("")) != 2 || len(s.List(Applied)) != 1 || s.List("")[0].ID != r.ID {
		t.Error("list ordering/filter")
	}
}

func TestSubmitValidation(t *testing.T) {
	s := newSvc(t)
	ctx := context.Background()
	if _, err := s.Submit(ctx, "", safeReq()); !errors.Is(err, ErrInvalid) {
		t.Error("missing actor")
	}
	bad := safeReq()
	bad.Title = "  "
	if _, err := s.Submit(ctx, "a", bad); !errors.Is(err, ErrInvalid) {
		t.Error("blank title")
	}
	unk := safeReq()
	unk.Change.ID = "ghost"
	if _, err := s.Submit(ctx, "a", unk); !errors.Is(err, preflight.ErrUnknownTarget) {
		t.Errorf("unknown target: %v", err)
	}
	if len(s.Log().Events("")) != 0 {
		t.Error("failed submissions must leave no trace in the log")
	}
}

func TestConcurrentApprovalsNeverExceedOrDoubleCount(t *testing.T) {
	s := newSvc(t)
	c, _ := s.Submit(context.Background(), "alice", riskyReq())
	var wg sync.WaitGroup
	for _, who := range []string{"b", "b", "b", "c", "c", "d", "e"} {
		wg.Add(1)
		go func(w string) { defer wg.Done(); s.Approve(w, c.ID) }(who)
	}
	wg.Wait()
	got, _ := s.Get(c.ID)
	seen := map[string]bool{}
	for _, a := range got.Approvers {
		if seen[a] {
			t.Fatalf("duplicate approver recorded: %v", got.Approvers)
		}
		seen[a] = true
	}
	if got.Status != Approved || len(got.Approvers) != 2 {
		t.Fatalf("%+v", got)
	}
	if err := s.Log().Verify(); err != nil {
		t.Fatal(err)
	}
}

func TestLogDetectsTampering(t *testing.T) {
	build := func() *Log {
		l := NewMemLog()
		for i, a := range []string{"alice", "bob", "carol"} {
			l.Append(a, "t", "c1", map[string]int{"i": i})
		}
		return l
	}
	if err := build().Verify(); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(l *Log){
		"edit data":     func(l *Log) { l.events[1].Data = json.RawMessage(`{"i":99}`) },
		"edit actor":    func(l *Log) { l.events[1].Actor = "mallory" },
		"delete middle": func(l *Log) { l.events = append(l.events[:1], l.events[2:]...) },
		"reorder":       func(l *Log) { l.events[0], l.events[1] = l.events[1], l.events[0] },
		"rewrite hash":  func(l *Log) { l.events[2].Hash = strings.Repeat("0", 64) },
		"drop first":    func(l *Log) { l.events = l.events[1:] },
		"forge and rehash only the edited event": func(l *Log) {
			l.events[1].Data = json.RawMessage(`{"i":99}`)
			l.events[1].Hash = computeHash(l.events[1]) // later events still point at the old hash
		},
	}
	for name, tamper := range cases {
		l := build()
		tamper(l)
		if err := l.Verify(); !errors.Is(err, ErrTampered) {
			t.Errorf("%s: not detected (%v)", name, err)
		}
	}
}

func TestHashIsUnambiguousAcrossFieldBoundaries(t *testing.T) {
	a := Event{Seq: 1, Actor: "a|b", Type: "c"}
	b := Event{Seq: 1, Actor: "a", Type: "b|c"}
	if computeHash(a) == computeHash(b) {
		t.Fatal("field-boundary collision")
	}
}

func TestPersistenceReplayRebuildsStateAndRefusesTamperedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	l, err := OpenLog(path, true)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := NewService(l, nil)
	c, _ := s.Submit(context.Background(), "alice", riskyReq())
	s.Approve("bob", c.ID)
	s.Approve("carol", c.ID)
	s.Apply("bob", c.ID)
	want, _ := s.Get(c.ID)
	l.Close()

	l2, err := OpenLog(path, false)
	if err != nil {
		t.Fatal(err)
	}
	s2, err := NewService(l2, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s2.Get(c.ID)
	if err != nil || got.Status != Applied || len(got.Approvers) != 2 || got.TopologyHash != want.TopologyHash {
		t.Fatalf("replay mismatch: %v %+v", err, got)
	}
	// appending after restart continues the same chain
	if _, err := s2.Submit(context.Background(), "alice", safeReq()); err != nil {
		t.Fatal(err)
	}
	l2.Close()
	if _, err := OpenLog(path, false); err != nil {
		t.Fatalf("chain must survive restart + append: %v", err)
	}

	raw, _ := os.ReadFile(path)
	forged := bytes.Replace(raw, []byte(`"actor":"bob"`), []byte(`"actor":"eve"`), 1)
	os.WriteFile(path, forged, 0o600)
	if _, err := OpenLog(path, false); !errors.Is(err, ErrTampered) {
		t.Fatalf("tampered file must be refused, got %v", err)
	}
	os.WriteFile(path, []byte("not json\n"), 0o600)
	if _, err := OpenLog(path, false); !errors.Is(err, ErrTampered) {
		t.Fatalf("garbage must be refused, got %v", err)
	}
}

// ---- HTTP ----

func httpSvc(t *testing.T, a Analyzer) (*httptest.Server, *Service) {
	t.Helper()
	s, err := NewService(NewMemLog(), a)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	s.Routes(mux)
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts, s
}

func call(t *testing.T, method, url, who string, body any) (int, map[string]any) {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, url, rd)
	if who != "" {
		req.Header.Set(SubjectHeader, who)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func TestHTTPFullWorkflow(t *testing.T) {
	ts, _ := httpSvc(t, nil)
	if code, _ := call(t, "GET", ts.URL+"/v1/changes", "", nil); code != 401 {
		t.Fatalf("anonymous must be 401, got %d", code)
	}
	code, c := call(t, "POST", ts.URL+"/v1/changes", "alice", riskyReq())
	if code != 201 || c["risk"] != "critical" {
		t.Fatalf("%d %v", code, c)
	}
	id := c["id"].(string)
	base := ts.URL + "/v1/changes/" + id
	if code, _ := call(t, "POST", base+"/approve", "alice", nil); code != 403 {
		t.Errorf("self approve: %d", code)
	}
	if code, _ := call(t, "POST", base+"/approve", "bob", nil); code != 200 {
		t.Errorf("approve: %d", code)
	}
	if code, _ := call(t, "POST", base+"/approve", "bob", nil); code != 409 {
		t.Errorf("duplicate: %d", code)
	}
	if code, _ := call(t, "POST", base+"/apply", "bob", nil); code != 409 {
		t.Errorf("apply before quorum: %d", code)
	}
	call(t, "POST", base+"/approve", "carol", nil)
	if code, got := call(t, "POST", base+"/apply", "carol", nil); code != 200 || got["status"] != "applied" {
		t.Errorf("apply: %d %v", code, got)
	}
	if code, ev := call(t, "GET", ts.URL+"/v1/audit/events?change="+id, "auditor", nil); code != 200 || len(ev["events"].([]any)) != 4 {
		t.Errorf("events: %d %v", code, ev)
	}
	if code, v := call(t, "GET", ts.URL+"/v1/audit/verify", "auditor", nil); code != 200 || v["ok"] != true || v["events"].(float64) != 4 {
		t.Errorf("verify: %d %v", code, v)
	}
	if code, _ := call(t, "GET", ts.URL+"/v1/changes/nope", "alice", nil); code != 404 {
		t.Errorf("unknown: %d", code)
	}
	if code, _ := call(t, "POST", ts.URL+"/v1/changes", "alice", map[string]any{"title": "x", "bogus": 1}); code != 400 {
		t.Errorf("unknown field: %d", code)
	}
	if code, _ := call(t, "POST", ts.URL+"/v1/changes/"+id+"/reject", "bob", map[string]string{}); code != 400 {
		t.Errorf("reject without reason: %d", code)
	}
}

func TestVerifyEndpointReportsTamperingAs409(t *testing.T) {
	ts, s := httpSvc(t, nil)
	call(t, "POST", ts.URL+"/v1/changes", "alice", safeReq())
	s.log.events[0].Actor = "mallory"
	code, v := call(t, "GET", ts.URL+"/v1/audit/verify", "auditor", nil)
	if code != 409 || v["ok"] != false {
		t.Fatalf("%d %v", code, v)
	}
}

func TestHTTPAnalyzerMapsErrorsAndTripsBreaker(t *testing.T) {
	status := http.StatusOK
	calls := 0
	topo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(status)
		if status == 200 {
			io.WriteString(w, `{"report":{"change":{"kind":"remove_node","id":"a"},"target":"a","risk":"none","cut_off":[],"warnings":[]}}`)
		} else {
			io.WriteString(w, `{"error":"x"}`)
		}
	}))
	defer topo.Close()
	a := NewHTTPAnalyzer(topo.URL, time.Second)
	ctx := context.Background()
	ch := preflight.Change{Kind: preflight.RemoveNode, ID: "a"}

	if rep, err := a.Assess(ctx, ringNodes, ringLinks, ch); err != nil || rep.Risk != preflight.RiskNone {
		t.Fatalf("%v %+v", err, rep)
	}
	for st, want := range map[int]error{404: preflight.ErrUnknownTarget, 422: preflight.ErrInvalidTopology, 400: preflight.ErrInvalidChange} {
		status = st
		if _, err := a.Assess(ctx, ringNodes, ringLinks, ch); !errors.Is(err, want) {
			t.Errorf("status %d: %v", st, err)
		}
	}
	if a.Breaker.State() != 0 {
		t.Error("4xx responses must not trip the breaker")
	}
	status = 503
	before := calls
	for i := 0; i < 8; i++ {
		if _, err := a.Assess(ctx, ringNodes, ringLinks, ch); !errors.Is(err, ErrAnalyzerUnavailable) {
			t.Fatalf("503 must map to unavailable: %v", err)
		}
	}
	if calls-before != 5 {
		t.Fatalf("breaker should cap calls to a failing dependency at 5, got %d", calls-before)
	}
	// and the API turns it into a retryable 503, not a 500
	ts, _ := httpSvc(t, a)
	if code, _ := call(t, "POST", ts.URL+"/v1/changes", "alice", safeReq()); code != 503 {
		t.Errorf("dependency down => 503, got %d", code)
	}
}
