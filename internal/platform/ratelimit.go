package platform

import (
	"sync"
	"time"
)

// Limiter is a per-key token bucket. Idle keys are swept lazily so memory is
// bounded by the number of recently active clients.
type Limiter struct {
	rate  float64 // tokens per second
	burst float64
	now   func() time.Time

	mu      sync.Mutex
	buckets map[string]*bucket
	sweepAt time.Time
}

type bucket struct {
	tokens float64
	last   time.Time
}

// NewLimiter allows `rate` requests/second per key with the given burst.
func NewLimiter(rate float64, burst int, now func() time.Time) *Limiter {
	if now == nil {
		now = time.Now
	}
	return &Limiter{rate: rate, burst: float64(burst), now: now, buckets: map[string]*bucket{}}
}

// Allow consumes one token for key, returning false when the key is over limit.
func (l *Limiter) Allow(key string) bool {
	t := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()
	if t.After(l.sweepAt) {
		for k, b := range l.buckets {
			if t.Sub(b.last) > time.Minute {
				delete(l.buckets, k)
			}
		}
		l.sweepAt = t.Add(time.Minute)
	}
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: t}
		l.buckets[key] = b
	}
	b.tokens += t.Sub(b.last).Seconds() * l.rate
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.last = t
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Size is the number of tracked keys (for tests/metrics).
func (l *Limiter) Size() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buckets)
}
