package resilience

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateHalfOpen
	StateOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF_OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig holds configuration for a circuit breaker
type CircuitBreakerConfig struct {
	MaxRequests         uint32        `json:"max_requests"`          // Max requests allowed in half-open state
	Interval            time.Duration `json:"interval"`              // Interval to clear the internal counts
	Timeout             time.Duration `json:"timeout"`               // Timeout to open the circuit breaker
	ReadyToTrip         func(counts Counts) bool `json:"-"`        // Function to determine when to trip
	OnStateChange       func(name string, from CircuitState, to CircuitState) `json:"-"` // Callback for state changes
	IsSuccessful        func(err error) bool `json:"-"`           // Function to determine if request was successful
}

// Counts holds the numbers of requests and their successes/failures
type Counts struct {
	Requests             uint32    `json:"requests"`
	TotalSuccesses       uint32    `json:"total_successes"`
	TotalFailures        uint32    `json:"total_failures"`
	ConsecutiveSuccesses uint32    `json:"consecutive_successes"`
	ConsecutiveFailures  uint32    `json:"consecutive_failures"`
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name          string
	maxRequests   uint32
	interval      time.Duration
	timeout       time.Duration
	readyToTrip   func(counts Counts) bool
	isSuccessful  func(err error) bool
	onStateChange func(name string, from CircuitState, to CircuitState)

	mutex      sync.Mutex
	state      CircuitState
	generation uint64
	counts     Counts
	expiry     time.Time
	logger     *zap.Logger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config CircuitBreakerConfig, logger *zap.Logger) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:          name,
		maxRequests:   config.MaxRequests,
		interval:      config.Interval,
		timeout:       config.Timeout,
		readyToTrip:   config.ReadyToTrip,
		isSuccessful:  config.IsSuccessful,
		onStateChange: config.OnStateChange,
		logger:        logger,
	}

	if cb.maxRequests == 0 {
		cb.maxRequests = 1
	}
	if cb.interval <= 0 {
		cb.interval = time.Duration(0)
	}
	if cb.timeout <= 0 {
		cb.timeout = 60 * time.Second
	}
	if cb.readyToTrip == nil {
		cb.readyToTrip = defaultReadyToTrip
	}
	if cb.isSuccessful == nil {
		cb.isSuccessful = defaultIsSuccessful
	}

	cb.toNewGeneration(time.Now())
	return cb
}

// Execute runs the given request function with circuit breaker protection
func (cb *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	defer func() {
		e := recover()
		if e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	result, err := req()
	cb.afterRequest(generation, cb.isSuccessful(err))
	return result, err
}

// ExecuteWithContext runs the given request function with context and circuit breaker protection
func (cb *CircuitBreaker) ExecuteWithContext(ctx context.Context, req func(context.Context) (interface{}, error)) (interface{}, error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	defer func() {
		e := recover()
		if e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	result, err := req(ctx)
	cb.afterRequest(generation, cb.isSuccessful(err))
	return result, err
}

// beforeRequest is called before executing a request
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, ErrOpenState
	} else if state == StateHalfOpen && cb.counts.Requests >= cb.maxRequests {
		return generation, ErrTooManyRequests
	}

	cb.counts.onRequest()
	return generation, nil
}

// afterRequest is called after executing a request
func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)
	if generation != before {
		return
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

// onSuccess handles successful requests
func (cb *CircuitBreaker) onSuccess(state CircuitState, now time.Time) {
	cb.counts.onSuccess()

	if state == StateHalfOpen {
		cb.setState(StateClosed, now)
	}
}

// onFailure handles failed requests
func (cb *CircuitBreaker) onFailure(state CircuitState, now time.Time) {
	cb.counts.onFailure()

	switch state {
	case StateClosed:
		if cb.readyToTrip(cb.counts) {
			cb.setState(StateOpen, now)
		}
	case StateHalfOpen:
		cb.setState(StateOpen, now)
	}
}

// currentState returns the current state of the circuit breaker
func (cb *CircuitBreaker) currentState(now time.Time) (CircuitState, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state, cb.generation
}

// setState changes the state of the circuit breaker
func (cb *CircuitBreaker) setState(state CircuitState, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration(now)

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}

	cb.logger.Info("Circuit breaker state changed",
		zap.String("name", cb.name),
		zap.String("from", prev.String()),
		zap.String("to", state.String()),
		zap.Uint32("requests", cb.counts.Requests),
		zap.Uint32("failures", cb.counts.TotalFailures))
}

// toNewGeneration resets the circuit breaker to a new generation
func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts.clear()

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	default: // StateHalfOpen
		cb.expiry = zero
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitState {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, _ := cb.currentState(now)
	return state
}

// Counts returns a copy of the current counts
func (cb *CircuitBreaker) Counts() Counts {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	return cb.counts
}

// Name returns the name of the circuit breaker
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// Methods for Counts
func (c *Counts) onRequest() {
	c.Requests++
}

func (c *Counts) onSuccess() {
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
}

func (c *Counts) onFailure() {
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
}

func (c *Counts) clear() {
	c.Requests = 0
	c.TotalSuccesses = 0
	c.TotalFailures = 0
	c.ConsecutiveSuccesses = 0
	c.ConsecutiveFailures = 0
}

// Default functions
func defaultReadyToTrip(counts Counts) bool {
	return counts.ConsecutiveFailures > 5
}

func defaultIsSuccessful(err error) bool {
	return err == nil
}

// Errors
var (
	ErrTooManyRequests = errors.New("circuit breaker: too many requests")
	ErrOpenState       = errors.New("circuit breaker: open state")
)

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
	logger   *zap.Logger
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(logger *zap.Logger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logger,
	}
}

// GetOrCreate gets an existing circuit breaker or creates a new one
func (cbm *CircuitBreakerManager) GetOrCreate(name string, config CircuitBreakerConfig) *CircuitBreaker {
	cbm.mutex.RLock()
	if breaker, exists := cbm.breakers[name]; exists {
		cbm.mutex.RUnlock()
		return breaker
	}
	cbm.mutex.RUnlock()

	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	// Double-check pattern
	if breaker, exists := cbm.breakers[name]; exists {
		return breaker
	}

	breaker := NewCircuitBreaker(name, config, cbm.logger)
	cbm.breakers[name] = breaker
	return breaker
}

// GetStats returns statistics for all circuit breakers
func (cbm *CircuitBreakerManager) GetStats() map[string]interface{} {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	stats := make(map[string]interface{})
	for name, breaker := range cbm.breakers {
		stats[name] = map[string]interface{}{
			"state":  breaker.State().String(),
			"counts": breaker.Counts(),
		}
	}

	return stats
}

// HealthCheckWithCircuitBreaker wraps a health check with circuit breaker
func HealthCheckWithCircuitBreaker(breaker *CircuitBreaker, healthCheck func() error) func() error {
	return func() error {
		_, err := breaker.Execute(func() (interface{}, error) {
			return nil, healthCheck()
		})
		return err
	}
}