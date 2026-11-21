// Package demo serves the public landing page and a read-only preflight
// playground. It never accepts a topology from the caller: the topology is
// fixed in code and the only input is which device to take down, validated
// against that fixed set. Nothing here reads or writes state.
package demo

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strconv"

	"netorchestrator/internal/preflight"
	"netorchestrator/internal/ratelimit"
)

//go:embed index.html
var indexHTML []byte

// Campus is the fixed sample network used by the playground.
var campusNodes = []preflight.Node{
	{ID: "core1", Name: "core1"}, {ID: "core2", Name: "core2"},
	{ID: "dist1", Name: "dist1"}, {ID: "dist2", Name: "dist2"},
	{ID: "access1", Name: "access1"}, {ID: "access2", Name: "access2"},
	{ID: "access3", Name: "access3"}, {ID: "printer-vlan", Name: "printer-vlan"},
}

var campusLinks = []preflight.Link{
	{ID: "l1", Source: "core1", Target: "core2", Cost: 1},
	{ID: "l2", Source: "core1", Target: "dist1", Cost: 10},
	{ID: "l3", Source: "core2", Target: "dist1", Cost: 10},
	{ID: "l4", Source: "core1", Target: "dist2", Cost: 10},
	{ID: "l5", Source: "core2", Target: "dist2", Cost: 10},
	{ID: "l6", Source: "dist1", Target: "access1", Cost: 100},
	{ID: "l7", Source: "dist2", Target: "access2", Cost: 100},
	{ID: "l8", Source: "dist2", Target: "access3", Cost: 100},
	{ID: "l9", Source: "access3", Target: "printer-vlan", Cost: 100},
}

// Handler serves GET /, GET /demo/topology and GET /demo/preflight.
type Handler struct {
	lim ratelimit.Limiter
}

// New returns a Handler. lim bounds the whole demo (one shared budget), so it
// cannot be used to burn CPU even though each analysis costs microseconds.
func New(lim ratelimit.Limiter) *Handler { return &Handler{lim: lim} }

// Register wires the routes onto mux. "GET /{$}" matches only the exact root.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", h.index)
	mux.HandleFunc("GET /demo/topology", h.limited(h.topology))
	mux.HandleFunc("GET /demo/preflight", h.limited(h.preflight))
}

func (h *Handler) index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; script-src 'unsafe-inline'; connect-src 'self'; base-uri 'none'; form-action 'none'")
	_, _ = w.Write(indexHTML)
}

func (h *Handler) limited(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if d := h.lim.Allow("demo"); !d.Allowed {
			secs := int(d.RetryAfter.Seconds()) + 1
			w.Header().Set("Retry-After", strconv.Itoa(secs))
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "demo rate limit reached, retry shortly"})
			return
		}
		next(w, r)
	}
}

func (h *Handler) topology(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"nodes": campusNodes, "links": campusLinks})
}

func (h *Handler) preflight(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("remove")
	known := false
	for _, n := range campusNodes {
		if n.ID == id {
			known = true
			break
		}
	}
	if !known {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "remove must be one of the sample devices, see /demo/topology"})
		return
	}
	rep, err := preflight.Assess(campusNodes, campusLinks, preflight.Change{Kind: preflight.RemoveNode, ID: id})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "analysis failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"report": rep})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
