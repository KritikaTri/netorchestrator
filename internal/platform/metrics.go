// Package platform holds the small, dependency-free building blocks every
// NetOrchestrator service shares: metrics, request IDs, access logs, rate
// limiting, circuit breaking, JWT verification and graceful shutdown.
//
// It is deliberately stdlib-only so each service builds to a tiny static
// binary and the whole package is unit-testable without external services.
package platform

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// DefaultBuckets are latency buckets in seconds (Prometheus-style, cumulative).
var DefaultBuckets = []float64{.0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5}

// Registry is a minimal Prometheus-compatible metric registry (text format 0.0.4).
type Registry struct {
	mu         sync.RWMutex
	counters   map[string]*atomic.Uint64 // key: name{labels}
	histograms map[string]*histogram
	gauges     map[string]*atomic.Int64
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		counters:   map[string]*atomic.Uint64{},
		histograms: map[string]*histogram{},
		gauges:     map[string]*atomic.Int64{},
	}
}

func key(name string, labels []string) string {
	if len(labels) == 0 {
		return name
	}
	var b strings.Builder
	b.WriteString(name)
	b.WriteByte('{')
	for i := 0; i+1 < len(labels); i += 2 {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, "%s=%q", labels[i], labels[i+1])
	}
	b.WriteByte('}')
	return b.String()
}

// Inc adds delta to the counter name{labels...}; labels are k,v,k,v pairs.
func (r *Registry) Inc(name string, delta uint64, labels ...string) {
	k := key(name, labels)
	r.mu.RLock()
	c, ok := r.counters[k]
	r.mu.RUnlock()
	if !ok {
		r.mu.Lock()
		if c, ok = r.counters[k]; !ok {
			c = new(atomic.Uint64)
			r.counters[k] = c
		}
		r.mu.Unlock()
	}
	c.Add(delta)
}

// Counter returns the current value of a counter (0 if absent). Used in tests.
func (r *Registry) Counter(name string, labels ...string) uint64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if c, ok := r.counters[key(name, labels)]; ok {
		return c.Load()
	}
	return 0
}

// GaugeAdd adjusts a gauge by delta (e.g. in-flight requests).
func (r *Registry) GaugeAdd(name string, delta int64, labels ...string) {
	k := key(name, labels)
	r.mu.RLock()
	g, ok := r.gauges[k]
	r.mu.RUnlock()
	if !ok {
		r.mu.Lock()
		if g, ok = r.gauges[k]; !ok {
			g = new(atomic.Int64)
			r.gauges[k] = g
		}
		r.mu.Unlock()
	}
	g.Add(delta)
}

type histogram struct {
	buckets []float64
	counts  []atomic.Uint64 // per-bucket (non-cumulative)
	sumNano atomic.Int64
	count   atomic.Uint64
}

// Observe records a duration in seconds under name{labels...}.
func (r *Registry) Observe(name string, seconds float64, labels ...string) {
	k := key(name, labels)
	r.mu.RLock()
	h, ok := r.histograms[k]
	r.mu.RUnlock()
	if !ok {
		r.mu.Lock()
		if h, ok = r.histograms[k]; !ok {
			h = &histogram{buckets: DefaultBuckets, counts: make([]atomic.Uint64, len(DefaultBuckets)+1)}
			r.histograms[k] = h
		}
		r.mu.Unlock()
	}
	i := sort.SearchFloat64s(h.buckets, seconds)
	h.counts[i].Add(1)
	h.sumNano.Add(int64(seconds * 1e9))
	h.count.Add(1)
}

// Quantile estimates the q-quantile (0..1) of a histogram from its buckets.
// It returns the upper bound of the bucket containing the quantile.
func (r *Registry) Quantile(name string, q float64, labels ...string) float64 {
	r.mu.RLock()
	h, ok := r.histograms[key(name, labels)]
	r.mu.RUnlock()
	if !ok || h.count.Load() == 0 {
		return 0
	}
	target := uint64(float64(h.count.Load())*q + 0.9999)
	var cum uint64
	for i := range h.counts {
		cum += h.counts[i].Load()
		if cum >= target {
			if i < len(h.buckets) {
				return h.buckets[i]
			}
			return h.buckets[len(h.buckets)-1]
		}
	}
	return h.buckets[len(h.buckets)-1]
}

// Expose renders the registry in Prometheus text exposition format.
func (r *Registry) Expose(w io.Writer) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var keys []string
	for k := range r.counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(w, "%s %d\n", k, r.counters[k].Load())
	}
	keys = keys[:0]
	for k := range r.gauges {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(w, "%s %d\n", k, r.gauges[k].Load())
	}
	keys = keys[:0]
	for k := range r.histograms {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		h := r.histograms[k]
		name, lbl := k, ""
		if i := strings.IndexByte(k, '{'); i >= 0 {
			name, lbl = k[:i], k[i+1:len(k)-1]
		}
		join := func(extra string) string {
			if lbl == "" {
				return "{" + extra + "}"
			}
			return "{" + lbl + "," + extra + "}"
		}
		var cum uint64
		for i, ub := range h.buckets {
			cum += h.counts[i].Load()
			fmt.Fprintf(w, "%s_bucket%s %d\n", name, join(fmt.Sprintf("le=%q", fmt.Sprint(ub))), cum)
		}
		cum += h.counts[len(h.buckets)].Load()
		fmt.Fprintf(w, "%s_bucket%s %d\n", name, join(`le="+Inf"`), cum)
		suffix := ""
		if lbl != "" {
			suffix = "{" + lbl + "}"
		}
		fmt.Fprintf(w, "%s_sum%s %g\n", name, suffix, float64(h.sumNano.Load())/1e9)
		fmt.Fprintf(w, "%s_count%s %d\n", name, suffix, h.count.Load())
	}
}
