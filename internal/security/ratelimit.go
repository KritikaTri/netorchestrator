package security

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	r        rate.Limit
	b        int
	cleanup  time.Duration
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
		cleanup:  time.Minute * 3,
	}

	// Start cleanup goroutine
	go rl.cleanupVisitors()

	return rl
}

func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.r, rl.b)
		rl.visitors[ip] = limiter
	}

	return limiter
}

func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for ip, limiter := range rl.visitors {
				if limiter.Allow() {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// Common rate limiting configurations
var (
	// Standard API rate limiting: 100 requests per minute
	StandardRateLimit = rate.Every(time.Minute / 100)
	
	// Strict rate limiting for sensitive endpoints: 10 requests per minute
	StrictRateLimit = rate.Every(time.Minute / 10)
	
	// Burst allowance
	StandardBurst = 10
	StrictBurst   = 3
)