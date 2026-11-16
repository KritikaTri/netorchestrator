package ratelimit

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type clock struct {
	mu sync.Mutex
	t  time.Time
}

func newClock() *clock                   { return &clock{t: time.Unix(1_000_000, 0)} }
func (c *clock) now() time.Time          { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *clock) advance(d time.Duration) { c.mu.Lock(); c.t = c.t.Add(d); c.mu.Unlock() }

// ---- token bucket ----

func TestTokenBucketBurstThenSteadyRate(t *testing.T) {
	c := newClock()
	l := NewTokenBucket(10, 5, 0, c.now) // 10/s, burst 5
	for i := 0; i < 5; i++ {
		d := l.Allow("k")
		if !d.Allowed || d.Remaining != 4-i || d.Limit != 5 {
			t.Fatalf("burst %d: %+v", i, d)
		}
	}
	if d := l.Allow("k"); d.Allowed || d.Remaining != 0 || d.RetryAfter == 0 {
		t.Fatalf("over burst: %+v", d)
	}
	// steady state: exactly 10 per second afterwards
	allowed := 0
	for i := 0; i < 1000; i++ { // 10 s in 10ms steps
		c.advance(10 * time.Millisecond)
		if l.Allow("k").Allowed {
			allowed++
		}
	}
	if allowed < 99 || allowed > 101 {
		t.Fatalf("allowed %d in 10s at 10/s", allowed)
	}
}

func TestTokenBucketRetryAfterIsSufficientAndTight(t *testing.T) {
	c := newClock()
	l := NewTokenBucket(4, 1, 0, c.now)
	l.Allow("k")
	d := l.Allow("k")
	if d.Allowed || d.RetryAfter != 250*time.Millisecond {
		t.Fatalf("%+v", d)
	}
	c.advance(d.RetryAfter - time.Millisecond)
	if l.Allow("k").Allowed {
		t.Fatal("must not be allowed before RetryAfter")
	}
	c.advance(time.Millisecond)
	if !l.Allow("k").Allowed {
		t.Fatal("must be allowed once RetryAfter has elapsed")
	}
}

func TestTokenBucketKeysIndependentAndNoOverfill(t *testing.T) {
	c := newClock()
	l := NewTokenBucket(1, 2, 0, c.now)
	l.Allow("a")
	l.Allow("a")
	if l.Allow("a").Allowed || !l.Allow("b").Allowed {
		t.Fatal("keys must be isolated")
	}
	c.advance(time.Hour) // refill must cap at burst, not accumulate
	n := 0
	for l.Allow("a").Allowed {
		n++
		if n > 10 {
			break
		}
	}
	if n != 2 {
		t.Fatalf("after long idle got %d tokens, want burst=2", n)
	}
}

func TestTokenBucketClockGoingBackwardsDoesNotMintTokens(t *testing.T) {
	c := newClock()
	l := NewTokenBucket(1, 1, 0, c.now)
	l.Allow("k")
	c.advance(-time.Hour)
	if l.Allow("k").Allowed {
		t.Fatal("backwards clock must not grant tokens")
	}
}

func TestTokenBucketConcurrentNeverExceedsBurst(t *testing.T) {
	l := NewTokenBucket(0.0001, 100, 0, nil)
	var ok atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow("shared").Allowed {
				ok.Add(1)
			}
		}()
	}
	wg.Wait()
	if ok.Load() != 100 {
		t.Fatalf("allowed %d, want exactly 100", ok.Load())
	}
}

func TestIdleKeysAreEvictedWithoutChangingBehaviour(t *testing.T) {
	c := newClock()
	l := NewTokenBucket(10, 10, 4, c.now)
	for i := 0; i < 1000; i++ {
		l.Allow(fmt.Sprintf("user-%d", i))
	}
	if l.Len() != 1000 {
		t.Fatalf("len %d", l.Len())
	}
	c.advance(10 * time.Minute)
	for i := 0; i < 50; i++ { // activity triggers the lazy sweep in each shard
		l.Allow(fmt.Sprintf("new-%d", i))
	}
	if l.Len() > 60 {
		t.Fatalf("idle keys not evicted: %d tracked", l.Len())
	}
	if d := l.Allow("user-1"); !d.Allowed || d.Remaining != 9 {
		t.Fatalf("an evicted key must behave like a fresh full bucket: %+v", d)
	}
}

// ---- sliding window ----

func TestSlidingWindowBasicLimit(t *testing.T) {
	c := newClock()
	l := NewSlidingWindow(5, 10*time.Second, 0, c.now)
	for i := 0; i < 5; i++ {
		if d := l.Allow("k"); !d.Allowed || d.Remaining != 4-i {
			t.Fatalf("%d: %+v", i, d)
		}
	}
	d := l.Allow("k")
	// window rolls in 10s, then the old 5 decay from full weight: the estimate
	// 5*(1-f) must drop to 4 before one more fits, i.e. f = 0.2 => +2s
	if d.Allowed || d.RetryAfter != 12*time.Second {
		t.Fatalf("%+v", d)
	}
	c.advance(d.RetryAfter - time.Millisecond)
	if l.Allow("k").Allowed {
		t.Fatal("allowed before RetryAfter")
	}
	c.advance(time.Millisecond)
	if !l.Allow("k").Allowed {
		t.Fatal("must be allowed once RetryAfter has elapsed")
	}
}

// The classic fixed-window bug: send `limit` just before a boundary and `limit`
// just after, getting 2x through in a moment. A sliding window must not.
func TestSlidingWindowHasNoBoundaryBurst(t *testing.T) {
	c := newClock()
	l := NewSlidingWindow(10, 10*time.Second, 0, c.now)
	c.advance(9900 * time.Millisecond)
	for i := 0; i < 10; i++ {
		if !l.Allow("k").Allowed {
			t.Fatal("first burst should pass")
		}
	}
	c.advance(200 * time.Millisecond) // 0.2s later, now in the next window
	n := 0
	for i := 0; i < 10; i++ {
		if l.Allow("k").Allowed {
			n++
		}
	}
	if n > 1 {
		t.Fatalf("%d more requests allowed 0.2s after a full burst; fixed windows would allow 10", n)
	}
}

func TestSlidingWindowRetryAfterAcrossRolloverIsExact(t *testing.T) {
	c := newClock()
	l := NewSlidingWindow(10, 10*time.Second, 0, c.now)
	c.advance(9 * time.Second)
	for i := 0; i < 10; i++ {
		l.Allow("k")
	}
	c.advance(2 * time.Second) // 1s into the next window; previous weight 0.9
	d := l.Allow("k")
	if d.Allowed || d.RetryAfter != 9*time.Second { // 8s to the roll + 1s of decay
		t.Fatalf("%+v", d)
	}
	c.advance(d.RetryAfter - time.Millisecond)
	if l.Allow("k").Allowed {
		t.Fatal("allowed before RetryAfter")
	}
	c.advance(time.Millisecond)
	if !l.Allow("k").Allowed {
		t.Fatal("must be allowed once RetryAfter has elapsed")
	}
}

func TestSlidingWindowRetryAfterDecayInsideWindowIsExact(t *testing.T) {
	c := newClock()
	l := NewSlidingWindow(10, 10*time.Second, 0, c.now)
	for i := 0; i < 10; i++ {
		l.Allow("k") // window starts now with a full budget used
	}
	c.advance(10500 * time.Millisecond) // next window, 5% in: estimate 9.5, request would make 10.5
	d := l.Allow("k")
	if d.Allowed || d.RetryAfter != 500*time.Millisecond {
		t.Fatalf("want denied with RetryAfter=500ms, got %+v", d)
	}
	c.advance(d.RetryAfter - time.Millisecond)
	if l.Allow("k").Allowed {
		t.Fatal("allowed before RetryAfter")
	}
	c.advance(time.Millisecond)
	if !l.Allow("k").Allowed {
		t.Fatal("must be allowed once RetryAfter has elapsed")
	}
}

func TestSlidingWindowRecoversAfterIdleAndKeysIsolated(t *testing.T) {
	c := newClock()
	l := NewSlidingWindow(3, time.Second, 0, c.now)
	for i := 0; i < 3; i++ {
		l.Allow("a")
	}
	if l.Allow("a").Allowed || !l.Allow("b").Allowed {
		t.Fatal("isolation")
	}
	c.advance(5 * time.Second)
	for i := 0; i < 3; i++ {
		if !l.Allow("a").Allowed {
			t.Fatalf("full budget expected after long idle (%d)", i)
		}
	}
}

// Compare against an exact sliding-log oracle under random load: long-run
// throughput matches, and the worst overshoot of the approximation is bounded.
func TestSlidingWindowApproximationVsExactLog(t *testing.T) {
	const limit, window = 50, time.Second
	c := newClock()
	l := NewSlidingWindow(limit, window, 0, c.now)
	rng := rand.New(rand.NewSource(7))
	var allowedAt []time.Time
	start := c.now()
	for i := 0; i < 20000; i++ {
		c.advance(time.Duration(rng.Intn(20)+1) * time.Millisecond) // ~100 req/s offered, limit 50/s
		if l.Allow("k").Allowed {
			allowedAt = append(allowedAt, c.now())
		}
	}
	total := c.now().Sub(start).Seconds()
	rate := float64(len(allowedAt)) / total
	if rate < 0.9*limit || rate > 1.1*limit {
		t.Fatalf("long-run rate %.1f/s, want ~%d", rate, limit)
	}
	// exact count in every trailing window ending at an allowed request
	worst, j := 0, 0
	for i := range allowedAt {
		for allowedAt[i].Sub(allowedAt[j]) >= window {
			j++
		}
		if n := i - j + 1; n > worst {
			worst = n
		}
	}
	t.Logf("exact worst trailing-window count = %d (limit %d, %.0f%% over)", worst, limit, 100*float64(worst-limit)/limit)
	if worst > limit+limit/2 {
		t.Fatalf("approximation overshoots too far: %d vs limit %d", worst, limit)
	}
}

func TestConstructorsRejectNonsense(t *testing.T) {
	for name, f := range map[string]func(){
		"zero rate":   func() { NewTokenBucket(0, 1, 0, nil) },
		"zero burst":  func() { NewTokenBucket(1, 0, 0, nil) },
		"zero limit":  func() { NewSlidingWindow(0, time.Second, 0, nil) },
		"zero window": func() { NewSlidingWindow(1, 0, 0, nil) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s accepted", name)
				}
			}()
			f()
		}()
	}
}

// ---- benchmarks ----

func benchParallel(b *testing.B, l Limiter, keys int) {
	names := make([]string, keys)
	for i := range names {
		names[i] = fmt.Sprintf("user-%d", i)
	}
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(rand.Int63()))
		for pb.Next() {
			l.Allow(names[r.Intn(keys)])
		}
	})
}

func BenchmarkAllow(b *testing.B) {
	for _, keys := range []int{1, 10000} {
		for _, shards := range []int{1, 64} {
			b.Run(fmt.Sprintf("token_bucket/keys=%d/shards=%d", keys, shards), func(b *testing.B) {
				benchParallel(b, NewTokenBucket(1e9, 1<<30, shards, nil), keys)
			})
		}
		b.Run(fmt.Sprintf("sliding_window/keys=%d/shards=64", keys), func(b *testing.B) {
			benchParallel(b, NewSlidingWindow(1<<30, time.Second, 64, nil), keys)
		})
	}
}
