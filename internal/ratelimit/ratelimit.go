// Package ratelimit provides per-key rate limiters behind one interface.
//
// Two algorithms are offered because they trade off differently:
//
//   - TokenBucket: allows bursts up to `burst`, then a steady `rate`/second.
//     O(1) memory per key. Best for APIs that should tolerate short spikes.
//   - SlidingWindow: at most `limit` requests in any trailing `window`
//     (sliding-window *counter* approximation: weights the previous window by
//     how much of it still overlaps). O(1) memory per key, and it avoids the
//     classic fixed-window flaw of letting 2x the limit through around a
//     window boundary.
//
// Both are sharded (independent locks per key hash) so unrelated keys do not
// contend, take an injectable clock for deterministic tests, and evict idle
// keys lazily so memory follows the number of *recently active* keys.
package ratelimit

import (
	"math"
	"sync"
	"time"
)

// Decision is the outcome of one Allow call, with what a client needs to back off.
type Decision struct {
	Allowed bool
	// Limit is the burst (token bucket) or per-window limit (sliding window).
	Limit int
	// Remaining requests available right now (after this one, if allowed).
	Remaining int
	// RetryAfter is, when denied, how long to wait before the next attempt
	// succeeds, assuming this key sends nothing in between (rounded up to
	// 1ms). Zero when allowed.
	RetryAfter time.Duration
	// Reset is how long until the key's budget is fully restored.
	Reset time.Duration
}

// Limiter decides whether a request for key may proceed.
type Limiter interface {
	Allow(key string) Decision
	// Len is the number of keys currently tracked (for metrics/tests).
	Len() int
}

const defaultShards = 64

// --- sharded storage ------------------------------------------------------

type entry[T any] struct {
	last time.Time
	v    T
}

type shard[T any] struct {
	mu        sync.Mutex
	m         map[string]*entry[T]
	nextSweep time.Time
}

type store[T any] struct {
	shards []shard[T]
	ttl    time.Duration // idle time after which an entry is equivalent to a fresh one
	now    func() time.Time
}

func newStore[T any](shards int, ttl time.Duration, now func() time.Time) *store[T] {
	if shards < 1 {
		shards = defaultShards
	}
	if now == nil {
		now = time.Now
	}
	s := &store[T]{shards: make([]shard[T], shards), ttl: ttl, now: now}
	for i := range s.shards {
		s.shards[i].m = map[string]*entry[T]{}
	}
	return s
}

func fnv32(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

// do runs fn on key's state while holding only that key's shard lock.
func (s *store[T]) do(key string, fn func(v *T, fresh bool, now time.Time) Decision) Decision {
	now := s.now()
	sh := &s.shards[fnv32(key)%uint32(len(s.shards))]
	sh.mu.Lock()
	defer sh.mu.Unlock()
	if now.After(sh.nextSweep) {
		for k, e := range sh.m {
			if now.Sub(e.last) >= s.ttl {
				delete(sh.m, k) // idle long enough that it equals a fresh entry
			}
		}
		sh.nextSweep = now.Add(s.ttl)
	}
	e, ok := sh.m[key]
	if !ok {
		e = &entry[T]{}
		sh.m[key] = e
	}
	d := fn(&e.v, !ok, now)
	e.last = now
	return d
}

func (s *store[T]) len() int {
	n := 0
	for i := range s.shards {
		s.shards[i].mu.Lock()
		n += len(s.shards[i].m)
		s.shards[i].mu.Unlock()
	}
	return n
}

// up rounds a positive duration up to the next millisecond so that waiting
// exactly RetryAfter is always enough despite floating-point error.
func up(sec float64) time.Duration {
	if sec <= 0 {
		return 0
	}
	ms := math.Ceil(sec * 1000)
	return time.Duration(ms) * time.Millisecond
}

// --- token bucket ----------------------------------------------------------

// TokenBucket allows `burst` requests at once, refilling at `rate` per second.
type TokenBucket struct {
	rate  float64
	burst float64
	st    *store[bucket]
}

type bucket struct {
	tokens float64
	last   time.Time
}

// NewTokenBucket builds a limiter. now may be nil (time.Now); shards <= 0 uses 64.
func NewTokenBucket(rate float64, burst int, shards int, now func() time.Time) *TokenBucket {
	if rate <= 0 || burst < 1 {
		panic("ratelimit: rate must be > 0 and burst >= 1")
	}
	// An entry idle for the time-to-full is indistinguishable from a new one.
	ttl := up(float64(burst)/rate) + time.Second
	return &TokenBucket{rate: rate, burst: float64(burst), st: newStore[bucket](shards, ttl, now)}
}

func (t *TokenBucket) Allow(key string) Decision {
	return t.st.do(key, func(b *bucket, fresh bool, now time.Time) Decision {
		if fresh {
			b.tokens, b.last = t.burst, now
		}
		if el := now.Sub(b.last).Seconds(); el > 0 { // ignore clock going backwards
			b.tokens = math.Min(t.burst, b.tokens+el*t.rate)
			b.last = now
		}
		d := Decision{Limit: int(t.burst)}
		if b.tokens >= 1 {
			b.tokens--
			d.Allowed = true
		} else {
			d.RetryAfter = up((1 - b.tokens) / t.rate)
		}
		d.Remaining = int(b.tokens)
		d.Reset = up((t.burst - b.tokens) / t.rate)
		return d
	})
}

func (t *TokenBucket) Len() int { return t.st.len() }

// --- sliding window counter -------------------------------------------------

// SlidingWindow allows about `limit` requests in any trailing `window`.
//
// The estimate for the trailing window is
//
//	prevCount * (1 - elapsedFractionOfCurrentWindow) + currentCount
//
// which assumes the previous window's requests were evenly spread. It is an
// approximation (as used by e.g. Cloudflare) that costs two integers per key.
type SlidingWindow struct {
	limit  int
	window time.Duration
	st     *store[winState]
}

type winState struct {
	start     time.Time
	cur, prev int
}

// NewSlidingWindow builds a limiter. now may be nil; shards <= 0 uses 64.
func NewSlidingWindow(limit int, window time.Duration, shards int, now func() time.Time) *SlidingWindow {
	if limit < 1 || window <= 0 {
		panic("ratelimit: limit must be >= 1 and window > 0")
	}
	return &SlidingWindow{limit: limit, window: window, st: newStore[winState](shards, 2*window, now)}
}

func (w *SlidingWindow) Allow(key string) Decision {
	return w.st.do(key, func(s *winState, fresh bool, now time.Time) Decision {
		if fresh {
			s.start = now
		}
		el := now.Sub(s.start)
		if el < 0 { // clock moved backwards: treat as the start of the window
			el = 0
		}
		if el >= w.window {
			if n := el / w.window; n == 1 {
				s.prev = s.cur
			} else {
				s.prev = 0 // idle for 2+ windows: nothing from the past counts
			}
			s.cur = 0
			s.start = s.start.Add((el / w.window) * w.window)
			el = now.Sub(s.start)
		}
		frac := float64(el) / float64(w.window)
		est := float64(s.prev)*(1-frac) + float64(s.cur)
		d := Decision{Limit: w.limit}
		if est+1 <= float64(w.limit) {
			s.cur++
			est++
			d.Allowed = true
		} else if s.prev > 0 && s.cur+1 < w.limit {
			// the previous window's weight decays linearly: solve for when the estimate fits
			x := 1 - float64(w.limit-s.cur-1)/float64(s.prev)
			d.RetryAfter = up((x - frac) * w.window.Seconds())
		} else {
			// The window must roll first; then the (old) current count becomes the
			// previous window and decays from weight 1, so add the time until it fits.
			after := 0.0
			if s.cur > 0 {
				after = math.Min(1, math.Max(0, 1-float64(w.limit-1)/float64(s.cur)))
			}
			d.RetryAfter = up((1 - frac + after) * w.window.Seconds())
		}
		if rem := math.Floor(float64(w.limit) - est); rem > 0 {
			d.Remaining = int(rem)
		}
		d.Reset = up((1 - frac) * w.window.Seconds())
		return d
	})
}

func (w *SlidingWindow) Len() int { return w.st.len() }
