// Package topologysvc is the stateless topology-analysis microservice.
//
// Every request carries the topology it wants analysed, so replicas share no
// state and scale horizontally behind any load balancer. Identical requests
// are answered from a per-replica LRU keyed on a SHA-256 of the request.
package topologysvc

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"netorchestrator/internal/platform"
	"netorchestrator/internal/preflight"
	"netorchestrator/internal/topology"
)

// Config tunes the service.
type Config struct {
	MaxBodyBytes int64 // request size cap
	MaxInFlight  int   // concurrent analyses before shedding load with 429
	CacheEntries int   // LRU capacity (0 disables)
}

// DefaultConfig returns production-leaning defaults.
func DefaultConfig() Config {
	return Config{MaxBodyBytes: 8 << 20, MaxInFlight: 64, CacheEntries: 1024}
}

// Server implements the HTTP API.
type Server struct {
	cfg   Config
	sem   chan struct{}
	cache *lru
	reg   *platform.Registry
}

// New builds a Server. reg may be nil.
func New(cfg Config, reg *platform.Registry) *Server {
	if reg == nil {
		reg = platform.NewRegistry()
	}
	return &Server{cfg: cfg, sem: make(chan struct{}, cfg.MaxInFlight), cache: newLRU(cfg.CacheEntries), reg: reg}
}

// Routes registers the API on mux.
func (s *Server) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/analyze/preflight", s.handle("preflight", s.preflight))
	mux.HandleFunc("POST /v1/analyze/resilience", s.handle("resilience", s.resilience))
	mux.HandleFunc("POST /v1/analyze/path", s.handle("path", s.path))
}

type topo struct {
	Nodes []preflight.Node `json:"nodes"`
	Links []preflight.Link `json:"links"`
}

type apiError struct {
	status int
	msg    string
}

func (e *apiError) Error() string { return e.msg }

func (s *Server) handle(name string, fn func(body []byte) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Backpressure: shed load instead of queueing unboundedly.
		select {
		case s.sem <- struct{}{}:
			defer func() { <-s.sem }()
		default:
			s.reg.Inc("topology_shed_total", 1)
			w.Header().Set("Retry-After", "1")
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "server busy"})
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, s.cfg.MaxBodyBytes))
		if err != nil {
			var mbe *http.MaxBytesError
			if errors.As(err, &mbe) {
				writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request too large"})
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unreadable body"})
			return
		}
		sum := sha256.Sum256(append([]byte(name+"\x00"), body...))
		ck := hex.EncodeToString(sum[:])
		if cached, ok := s.cache.Get(ck); ok {
			s.reg.Inc("topology_cache_total", 1, "result", "hit")
			w.Header().Set("X-Cache", "HIT")
			w.Header().Set("Content-Type", "application/json")
			w.Write(cached)
			return
		}
		s.reg.Inc("topology_cache_total", 1, "result", "miss")
		out, err := fn(body)
		if err != nil {
			var ae *apiError
			if errors.As(err, &ae) {
				writeJSON(w, ae.status, map[string]string{"error": ae.msg})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "analysis failed"})
			return
		}
		enc, _ := json.Marshal(out)
		s.cache.Put(ck, enc)
		w.Header().Set("X-Cache", "MISS")
		w.Header().Set("Content-Type", "application/json")
		w.Write(enc)
	}
}

func decode(body []byte, v any) error {
	dec := json.NewDecoder(bytesReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return &apiError{http.StatusBadRequest, "invalid JSON body: " + err.Error()}
	}
	return nil
}

func mapPreflightErr(err error) error {
	switch {
	case errors.Is(err, preflight.ErrUnknownTarget), errors.Is(err, topology.ErrUnknownNode):
		return &apiError{http.StatusNotFound, err.Error()}
	case errors.Is(err, preflight.ErrInvalidChange):
		return &apiError{http.StatusBadRequest, err.Error()}
	case errors.Is(err, preflight.ErrInvalidTopology):
		return &apiError{http.StatusUnprocessableEntity, err.Error()}
	}
	return err
}

func (s *Server) preflight(body []byte) (any, error) {
	var req struct {
		topo
		Change preflight.Change `json:"change"`
	}
	if err := decode(body, &req); err != nil {
		return nil, err
	}
	rep, err := preflight.Assess(req.Nodes, req.Links, req.Change)
	if err != nil {
		return nil, mapPreflightErr(err)
	}
	return map[string]any{"report": rep}, nil
}

func (s *Server) resilience(body []byte) (any, error) {
	var req topo
	if err := decode(body, &req); err != nil {
		return nil, err
	}
	res, err := preflight.Analyze(req.Nodes, req.Links)
	if err != nil {
		return nil, mapPreflightErr(err)
	}
	return map[string]any{"resilience": res}, nil
}

func (s *Server) path(body []byte) (any, error) {
	var req struct {
		topo
		Src string `json:"src"`
		Dst string `json:"dst"`
	}
	if err := decode(body, &req); err != nil {
		return nil, err
	}
	g, err := preflight.BuildGraph(req.Nodes, req.Links)
	if err != nil {
		return nil, mapPreflightErr(err)
	}
	p, err := g.ShortestPath(topology.NodeID(req.Src), topology.NodeID(req.Dst))
	switch {
	case errors.Is(err, topology.ErrNoPath):
		return nil, &apiError{http.StatusNotFound, "no path between " + strconv.Quote(req.Src) + " and " + strconv.Quote(req.Dst)}
	case err != nil:
		return nil, mapPreflightErr(err)
	}
	return map[string]any{"path": p}, nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
