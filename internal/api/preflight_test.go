package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"netorchestrator/internal/models"
)

type fakeSource struct {
	net      *models.Network
	nodes    []models.Node
	links    []models.Link
	netErr   error
	nodesErr error
}

func (f *fakeSource) GetNetwork(context.Context, uuid.UUID) (*models.Network, error) {
	return f.net, f.netErr
}
func (f *fakeSource) ListNodes(context.Context, uuid.UUID) ([]models.Node, error) {
	return f.nodes, f.nodesErr
}
func (f *fakeSource) ListLinks(context.Context, uuid.UUID) ([]models.Link, error) {
	return f.links, nil
}

// hub-and-spoke: hub -- leafA, hub -- leafB
func fixture() (*fakeSource, map[string]uuid.UUID) {
	ids := map[string]uuid.UUID{"hub": uuid.New(), "a": uuid.New(), "b": uuid.New(), "l1": uuid.New(), "l2": uuid.New()}
	return &fakeSource{
		net: &models.Network{ID: uuid.New()},
		nodes: []models.Node{
			{ID: ids["hub"], Name: "hub"}, {ID: ids["a"], Name: "leafA"}, {ID: ids["b"], Name: "leafB"},
		},
		links: []models.Link{
			{ID: ids["l1"], SourceNodeID: ids["hub"], TargetNodeID: ids["a"]},
			{ID: ids["l2"], SourceNodeID: ids["hub"], TargetNodeID: ids["b"]},
		},
	}, ids
}

func router(src TopologySource) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewPreflightHandlers(src, zap.NewNop())
	r := gin.New()
	r.POST("/networks/:id/preflight", h.Preflight)
	r.GET("/networks/:id/resilience", h.Resilience)
	return r
}

func do(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestPreflightReportsCutOffDevices(t *testing.T) {
	src, ids := fixture()
	w := do(router(src), "POST", "/networks/"+uuid.NewString()+"/preflight",
		`{"kind":"remove_node","id":"`+ids["hub"].String()+`"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	var out struct {
		Report struct {
			Risk   string `json:"risk"`
			Target string `json:"target"`
			CutOff []struct {
				Name string `json:"name"`
			} `json:"cut_off"`
		} `json:"report"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Report.Target != "hub" || len(out.Report.CutOff) != 1 || out.Report.Risk == "none" {
		t.Fatalf("unexpected report: %s", w.Body)
	}
}

func TestPreflightErrorMapping(t *testing.T) {
	src, ids := fixture()
	r := router(src)
	nid := uuid.NewString()
	cases := []struct {
		name, body string
		want       int
	}{
		{"unknown node", `{"kind":"remove_node","id":"` + uuid.NewString() + `"}`, http.StatusNotFound},
		{"bad kind", `{"kind":"explode","id":"` + ids["hub"].String() + `"}`, http.StatusBadRequest},
		{"missing fields", `{}`, http.StatusBadRequest},
		{"not json", `nope`, http.StatusBadRequest},
	}
	for _, c := range cases {
		if w := do(r, "POST", "/networks/"+nid+"/preflight", c.body); w.Code != c.want {
			t.Errorf("%s: got %d want %d (%s)", c.name, w.Code, c.want, w.Body)
		}
	}
	if w := do(r, "POST", "/networks/not-a-uuid/preflight", `{}`); w.Code != http.StatusBadRequest {
		t.Errorf("bad network id: %d", w.Code)
	}
}

func TestNetworkNotFoundAndBackendFailure(t *testing.T) {
	src, ids := fixture()
	src.netErr = errors.New("network not found")
	body := `{"kind":"remove_node","id":"` + ids["hub"].String() + `"}`
	if w := do(router(src), "POST", "/networks/"+uuid.NewString()+"/preflight", body); w.Code != http.StatusNotFound {
		t.Errorf("not found: %d", w.Code)
	}
	src.netErr = errors.New("db exploded")
	w := do(router(src), "POST", "/networks/"+uuid.NewString()+"/preflight", body)
	if w.Code != http.StatusInternalServerError || bytes.Contains(w.Body.Bytes(), []byte("exploded")) {
		t.Errorf("must be 500 without leaking internals: %d %s", w.Code, w.Body)
	}
}

func TestResilienceListsSinglePointsOfFailure(t *testing.T) {
	src, _ := fixture()
	w := do(router(src), "GET", "/networks/"+uuid.NewString()+"/resilience", "")
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var out struct {
		Resilience struct {
			Redundant bool `json:"redundant"`
			APs       []struct {
				Name string `json:"name"`
			} `json:"articulation_points"`
		} `json:"resilience"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if out.Resilience.Redundant || len(out.Resilience.APs) != 1 || out.Resilience.APs[0].Name != "hub" {
		t.Fatalf("unexpected: %s", w.Body)
	}
}
