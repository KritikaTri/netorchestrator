package platform

import (
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the breaker is rejecting calls.
var ErrCircuitOpen = errors.New("platform: circuit open")

// State of a circuit breaker.
type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

func (s State) String() string { return [...]string{"closed", "open", "half-open"}[s] }

// Breaker is a consecutive-failure circuit breaker with half-open probing.
// After Threshold consecutive failures it opens for Cooldown, then lets a
// single probe through; success closes it, failure re-opens it.
type Breaker struct {
	Threshold int
	Cooldown  time.Duration
	now       func() time.Time

	mu       sync.Mutex
	state    State
	fails    int
	openedAt time.Time
	probing  bool
}

// NewBreaker builds a breaker. A nil clock uses time.Now.
func NewBreaker(threshold int, cooldown time.Duration, now func() time.Time) *Breaker {
	if now == nil {
		now = time.Now
	}
	if threshold < 1 {
		threshold = 1
	}
	return &Breaker{Threshold: threshold, Cooldown: cooldown, now: now}
}

// Allow reports whether a call may proceed. Every Allow() that returns nil
// must be followed by exactly one Done().
func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.state {
	case Open:
		if b.now().Sub(b.openedAt) < b.Cooldown {
			return ErrCircuitOpen
		}
		b.state, b.probing = HalfOpen, true
		return nil
	case HalfOpen:
		if b.probing {
			return ErrCircuitOpen // one probe at a time
		}
		b.probing = true
	}
	return nil
}

// Done records the outcome of an allowed call.
func (b *Breaker) Done(success bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if success {
		b.fails, b.state, b.probing = 0, Closed, false
		return
	}
	b.fails++
	if b.state == HalfOpen || b.fails >= b.Threshold {
		b.state, b.openedAt, b.probing = Open, b.now(), false
	}
}

// State returns the current state (Open reads as HalfOpen-pending once cooled).
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
