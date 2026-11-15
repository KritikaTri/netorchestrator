package changesvc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"netorchestrator/internal/platform"
	"netorchestrator/internal/preflight"
)

// ErrAnalyzerUnavailable means topology-svc could not be reached or is shedding load.
var ErrAnalyzerUnavailable = errors.New("topology analysis unavailable")

// HTTPAnalyzer calls topology-svc with a timeout and a circuit breaker so a
// sick dependency fails fast instead of piling up blocked requests.
type HTTPAnalyzer struct {
	BaseURL string
	Client  *http.Client
	Breaker *platform.Breaker
}

// NewHTTPAnalyzer builds an analyzer for the given topology-svc base URL.
func NewHTTPAnalyzer(baseURL string, timeout time.Duration) *HTTPAnalyzer {
	return &HTTPAnalyzer{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Client:  &http.Client{Timeout: timeout},
		Breaker: platform.NewBreaker(5, 10*time.Second, nil),
	}
}

func (a *HTTPAnalyzer) Assess(ctx context.Context, n []preflight.Node, l []preflight.Link, c preflight.Change) (*preflight.Report, error) {
	if err := a.Breaker.Allow(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAnalyzerUnavailable, err)
	}
	body, _ := json.Marshal(map[string]any{"nodes": n, "links": l, "change": c})
	req, err := http.NewRequestWithContext(ctx, "POST", a.BaseURL+"/v1/analyze/preflight", bytes.NewReader(body))
	if err != nil {
		a.Breaker.Done(true)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if id := platform.RequestID(ctx); id != "" {
		req.Header.Set(platform.HeaderRequestID, id)
	}
	resp, err := a.Client.Do(req)
	if err != nil {
		a.Breaker.Done(false)
		return nil, fmt.Errorf("%w: %v", ErrAnalyzerUnavailable, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))

	switch {
	case resp.StatusCode == http.StatusOK:
		a.Breaker.Done(true)
		var out struct {
			Report *preflight.Report `json:"report"`
		}
		if json.Unmarshal(raw, &out) != nil || out.Report == nil {
			return nil, fmt.Errorf("%w: malformed response", ErrAnalyzerUnavailable)
		}
		return out.Report, nil
	case resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests:
		a.Breaker.Done(false)
		return nil, fmt.Errorf("%w: upstream status %d", ErrAnalyzerUnavailable, resp.StatusCode)
	}
	a.Breaker.Done(true) // 4xx = our request was bad, not the dependency's fault
	msg := strings.TrimSpace(string(raw))
	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, fmt.Errorf("%w: %s", preflight.ErrUnknownTarget, msg)
	case http.StatusUnprocessableEntity:
		return nil, fmt.Errorf("%w: %s", preflight.ErrInvalidTopology, msg)
	}
	return nil, fmt.Errorf("%w: %s", preflight.ErrInvalidChange, msg)
}
