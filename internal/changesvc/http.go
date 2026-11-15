package changesvc

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"netorchestrator/internal/preflight"
)

// SubjectHeader carries the caller identity asserted by the gateway.
// change-svc must only be reachable through the gateway (network policy).
const SubjectHeader = "X-Authenticated-Subject"

// Routes registers the API on mux.
func (s *Service) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/changes", s.hSubmit)
	mux.HandleFunc("GET /v1/changes", s.hList)
	mux.HandleFunc("GET /v1/changes/{id}", s.hGet)
	mux.HandleFunc("POST /v1/changes/{id}/approve", s.act(func(a, id string, _ *http.Request) (*Change, error) { return s.Approve(a, id) }))
	mux.HandleFunc("POST /v1/changes/{id}/cancel", s.act(func(a, id string, _ *http.Request) (*Change, error) { return s.Cancel(a, id) }))
	mux.HandleFunc("POST /v1/changes/{id}/apply", s.act(func(a, id string, _ *http.Request) (*Change, error) { return s.Apply(a, id) }))
	mux.HandleFunc("POST /v1/changes/{id}/reject", s.act(func(a, id string, r *http.Request) (*Change, error) {
		var body struct {
			Reason string `json:"reason"`
		}
		_ = json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<16)).Decode(&body)
		if strings.TrimSpace(body.Reason) == "" {
			return nil, ErrInvalid
		}
		return s.Reject(a, id, body.Reason)
	}))
	mux.HandleFunc("GET /v1/audit/events", s.hEvents)
	mux.HandleFunc("GET /v1/audit/verify", s.hVerify)
}

func reply(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	msg := "internal error"
	switch {
	case errors.Is(err, ErrNotFound):
		code, msg = http.StatusNotFound, err.Error()
	case errors.Is(err, ErrInvalidState), errors.Is(err, ErrDuplicateApproval):
		code, msg = http.StatusConflict, err.Error()
	case errors.Is(err, ErrSelfApproval), errors.Is(err, ErrForbidden):
		code, msg = http.StatusForbidden, err.Error()
	case errors.Is(err, ErrInvalid), errors.Is(err, preflight.ErrInvalidChange):
		code, msg = http.StatusBadRequest, err.Error()
	case errors.Is(err, preflight.ErrUnknownTarget):
		code, msg = http.StatusNotFound, err.Error()
	case errors.Is(err, preflight.ErrInvalidTopology):
		code, msg = http.StatusUnprocessableEntity, err.Error()
	case errors.Is(err, ErrAnalyzerUnavailable):
		w.Header().Set("Retry-After", "2")
		code, msg = http.StatusServiceUnavailable, "topology analysis temporarily unavailable"
	}
	reply(w, code, map[string]string{"error": msg})
}

func actor(w http.ResponseWriter, r *http.Request) (string, bool) {
	a := strings.TrimSpace(r.Header.Get(SubjectHeader))
	if a == "" || len(a) > 128 {
		reply(w, http.StatusUnauthorized, map[string]string{"error": "missing caller identity"})
		return "", false
	}
	return a, true
}

func (s *Service) hSubmit(w http.ResponseWriter, r *http.Request) {
	a, ok := actor(w, r)
	if !ok {
		return
	}
	var req SubmitRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		fail(w, ErrInvalid)
		return
	}
	c, err := s.Submit(r.Context(), a, req)
	if err != nil {
		fail(w, err)
		return
	}
	reply(w, http.StatusCreated, c)
}

func (s *Service) hList(w http.ResponseWriter, r *http.Request) {
	if _, ok := actor(w, r); !ok {
		return
	}
	reply(w, http.StatusOK, map[string]any{"changes": s.List(Status(r.URL.Query().Get("status")))})
}

func (s *Service) hGet(w http.ResponseWriter, r *http.Request) {
	if _, ok := actor(w, r); !ok {
		return
	}
	c, err := s.Get(r.PathValue("id"))
	if err != nil {
		fail(w, err)
		return
	}
	reply(w, http.StatusOK, c)
}

func (s *Service) act(fn func(actor, id string, r *http.Request) (*Change, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a, ok := actor(w, r)
		if !ok {
			return
		}
		c, err := fn(a, r.PathValue("id"), r)
		if err != nil {
			fail(w, err)
			return
		}
		reply(w, http.StatusOK, c)
	}
}

func (s *Service) hEvents(w http.ResponseWriter, r *http.Request) {
	if _, ok := actor(w, r); !ok {
		return
	}
	reply(w, http.StatusOK, map[string]any{"events": s.log.Events(r.URL.Query().Get("change"))})
}

func (s *Service) hVerify(w http.ResponseWriter, r *http.Request) {
	if _, ok := actor(w, r); !ok {
		return
	}
	seq, head := s.log.Head()
	if err := s.log.Verify(); err != nil {
		reply(w, http.StatusConflict, map[string]any{"ok": false, "error": err.Error(), "events": seq})
		return
	}
	reply(w, http.StatusOK, map[string]any{"ok": true, "events": seq, "head_hash": head})
}
