package microservices

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// ServiceMesh implements a comprehensive microservices mesh with service discovery
type ServiceMesh struct {
	services         map[string]*ServiceGroup
	serviceDiscovery *ServiceDiscovery
	loadBalancer     *LoadBalancer
	healthChecker    *HealthChecker
	circuitBreaker   *CircuitBreakerManager
	rateLimiter      *DistributedRateLimiter
	logger           *zap.Logger
	mu               sync.RWMutex
}

// ServiceGroup represents a group of service instances
type ServiceGroup struct {
	Name      string                    `json:"name"`
	Instances []*ServiceInstance        `json:"instances"`
	Config    *ServiceConfiguration     `json:"config"`
	Metadata  map[string]interface{}    `json:"metadata"`
	Health    *ServiceHealthStatus      `json:"health"`
	LoadBalancer *LoadBalancingStrategy `json:"load_balancer"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
}

// ServiceInstance represents a single service instance
type ServiceInstance struct {
	ID          string                 `json:"id"`
	Address     string                 `json:"address"`
	Port        int                    `json:"port"`
	Protocol    string                 `json:"protocol"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Metadata    map[string]interface{} `json:"metadata"`
	Health      *InstanceHealth        `json:"health"`
	Metrics     *InstanceMetrics       `json:"metrics"`
	Weight      int                    `json:"weight"`
	Tags        []string               `json:"tags"`
	TLS         *TLSConfig             `json:"tls,omitempty"`
	RegisteredAt time.Time             `json:"registered_at"`
	LastSeen     time.Time             `json:"last_seen"`
}

// ServiceConfiguration holds service-specific configuration
type ServiceConfiguration struct {
	MaxInstances     int                    `json:"max_instances"`
	MinInstances     int                    `json:"min_instances"`
	HealthCheckPath  string                 `json:"health_check_path"`
	HealthTimeout    time.Duration          `json:"health_timeout"`
	LoadBalancing    string                 `json:"load_balancing"` // round_robin, weighted, least_conn, consistent_hash
	CircuitBreaker   *CircuitBreakerConfig  `json:"circuit_breaker"`
	RateLimit        *RateLimitConfig       `json:"rate_limit"`
	Retry            *RetryConfig           `json:"retry"`
	Timeout          *TimeoutConfig         `json:"timeout"`
	Security         *SecurityConfig        `json:"security"`
	Caching          *CachingConfig         `json:"caching"`
}

// ServiceHealthStatus represents overall health of a service group
type ServiceHealthStatus struct {
	Status       string    `json:"status"` // healthy, degraded, unhealthy
	HealthyCount int       `json:"healthy_count"`
	TotalCount   int       `json:"total_count"`
	LastCheck    time.Time `json:"last_check"`
	Issues       []string  `json:"issues,omitempty"`
}

// InstanceHealth represents health of a single instance
type InstanceHealth struct {
	Status         string    `json:"status"` // healthy, unhealthy, unknown
	ResponseTime   int64     `json:"response_time_ms"`
	LastCheck      time.Time `json:"last_check"`
	ConsecutiveFails int     `json:"consecutive_fails"`
	Uptime         int64     `json:"uptime_seconds"`
	ErrorRate      float64   `json:"error_rate"`
}

// InstanceMetrics contains performance metrics for a service instance
type InstanceMetrics struct {
	RequestCount    int64   `json:"request_count"`
	ErrorCount      int64   `json:"error_count"`
	AvgResponseTime float64 `json:"avg_response_time_ms"`
	CurrentLoad     float64 `json:"current_load"`
	MemoryUsage     float64 `json:"memory_usage_percent"`
	CPUUsage        float64 `json:"cpu_usage_percent"`
	ConnectionCount int     `json:"connection_count"`
	LastUpdated     time.Time `json:"last_updated"`
}

// ServiceDiscovery handles service registration and discovery
type ServiceDiscovery struct {
	registry    map[string]*ServiceGroup
	watchers    map[string][]ServiceWatcher
	hashRing    *ConsistentHashRing
	cache       *DistributedCache
	logger      *zap.Logger
	mu          sync.RWMutex
}

// ServiceWatcher interface for service change notifications
type ServiceWatcher interface {
	OnServiceRegistered(service *ServiceInstance)
	OnServiceDeregistered(service *ServiceInstance)
	OnServiceHealthChanged(service *ServiceInstance, health *InstanceHealth)
}

// LoadBalancer implements various load balancing algorithms
type LoadBalancer struct {
	strategies map[string]LoadBalancingStrategy
	metrics    *LoadBalancerMetrics
	logger     *zap.Logger
}

// LoadBalancingStrategy interface for different algorithms
type LoadBalancingStrategy interface {
	SelectInstance(instances []*ServiceInstance, context *RequestContext) *ServiceInstance
	UpdateMetrics(instance *ServiceInstance, metrics *RequestMetrics)
	GetName() string
}

// RequestContext contains information about the incoming request
type RequestContext struct {
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	Headers     map[string]string      `json:"headers"`
	ClientIP    string                 `json:"client_ip"`
	UserAgent   string                 `json:"user_agent"`
	SessionID   string                 `json:"session_id"`
	RequestID   string                 `json:"request_id"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// RequestMetrics contains metrics from a completed request
type RequestMetrics struct {
	Duration    time.Duration `json:"duration"`
	StatusCode  int          `json:"status_code"`
	BytesSent   int64        `json:"bytes_sent"`
	BytesRecv   int64        `json:"bytes_received"`
	Error       error        `json:"error,omitempty"`
	Timestamp   time.Time    `json:"timestamp"`
}

// RoundRobinStrategy implements round-robin load balancing
type RoundRobinStrategy struct {
	counter uint64
}

// WeightedStrategy implements weighted round-robin load balancing
type WeightedStrategy struct {
	weights map[string]int
	current map[string]int
	mu      sync.Mutex
}

// LeastConnectionsStrategy implements least connections load balancing
type LeastConnectionsStrategy struct {
	connections map[string]int64
	mu          sync.RWMutex
}

// ConsistentHashStrategy implements consistent hashing load balancing
type ConsistentHashStrategy struct {
	hashRing *ConsistentHashRing
}

// ConsistentHashRing is a simple consistent hash ring implementation
type ConsistentHashRing struct {
	nodes map[uint32]string
	keys  []uint32
}

// HealthChecker performs health checks on service instances
type HealthChecker struct {
	checkInterval time.Duration
	timeout       time.Duration
	httpClient    *http.Client
	checks        map[string]*HealthCheckConfig
	results       map[string]*HealthCheckResult
	logger        *zap.Logger
	mu            sync.RWMutex
	stopChan      chan struct{}
}

// HealthCheckConfig defines how to check service health
type HealthCheckConfig struct {
	Type        string        `json:"type"` // http, tcp, grpc
	Path        string        `json:"path"`
	Method      string        `json:"method"`
	Headers     map[string]string `json:"headers"`
	Body        string        `json:"body"`
	Interval    time.Duration `json:"interval"`
	Timeout     time.Duration `json:"timeout"`
	Retries     int           `json:"retries"`
	SuccessCode int           `json:"success_code"`
}

// HealthCheckResult contains the result of a health check
type HealthCheckResult struct {
	Status       string    `json:"status"`
	ResponseTime int64     `json:"response_time_ms"`
	Error        string    `json:"error,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	Attempts     int       `json:"attempts"`
}

// CircuitBreakerManager manages circuit breakers for all services
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	config   *CircuitBreakerConfig
	logger   *zap.Logger
	mu       sync.RWMutex
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name           string
	state          CircuitBreakerState
	failureCount   int64
	successCount   int64
	requestCount   int64
	lastFailTime   time.Time
	halfOpenStart  time.Time
	config         *CircuitBreakerConfig
	mu             sync.RWMutex
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreakerConfig defines circuit breaker behavior
type CircuitBreakerConfig struct {
	FailureThreshold   int           `json:"failure_threshold"`
	SuccessThreshold   int           `json:"success_threshold"`
	Timeout            time.Duration `json:"timeout"`
	RecoveryTimeout    time.Duration `json:"recovery_timeout"`
	MaxRequests        int           `json:"max_requests"`
	FailureRate        float64       `json:"failure_rate"`
}

// DistributedRateLimiter implements distributed rate limiting
type DistributedRateLimiter struct {
	limiters map[string]*TokenBucket
	storage  *DistributedStorage
	logger   *zap.Logger
	mu       sync.RWMutex
}

// TokenBucket implements token bucket rate limiting
type TokenBucket struct {
	capacity     int64
	tokens       int64
	refillRate   int64
	lastRefill   time.Time
	mu           sync.Mutex
}

// RateLimitConfig defines rate limiting parameters
type RateLimitConfig struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	Window            time.Duration `json:"window"`
	Strategy          string        `json:"strategy"` // token_bucket, sliding_window, fixed_window
}

// DistributedCache implements a sophisticated caching layer
type DistributedCache struct {
	nodes       []*CacheNode
	hashRing    *ConsistentHashRing
	replication int
	consistency string // eventual, strong, weak
	metrics     *CacheMetrics
	logger      *zap.Logger
	mu          sync.RWMutex
}

// CacheNode represents a cache node in the distributed cache
type CacheNode struct {
	ID       string            `json:"id"`
	Address  string            `json:"address"`
	Status   string            `json:"status"`
	Storage  map[string]*CacheEntry `json:"storage"`
	Metrics  *NodeMetrics      `json:"metrics"`
	mu       sync.RWMutex
}

// CacheEntry represents a cached item
type CacheEntry struct {
	Key        string                 `json:"key"`
	Value      interface{}            `json:"value"`
	TTL        time.Duration          `json:"ttl"`
	ExpiresAt  time.Time              `json:"expires_at"`
	Version    int64                  `json:"version"`
	Tags       []string               `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata"`
	AccessCount int64                 `json:"access_count"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// CacheMetrics tracks cache performance
type CacheMetrics struct {
	HitCount    int64 `json:"hit_count"`
	MissCount   int64 `json:"miss_count"`
	SetCount    int64 `json:"set_count"`
	DeleteCount int64 `json:"delete_count"`
	EvictCount  int64 `json:"evict_count"`
	TotalSize   int64 `json:"total_size"`
	NodeCount   int   `json:"node_count"`
}

// NodeMetrics tracks individual node performance
type NodeMetrics struct {
	RequestCount  int64   `json:"request_count"`
	HitRate       float64 `json:"hit_rate"`
	MemoryUsage   int64   `json:"memory_usage"`
	DiskUsage     int64   `json:"disk_usage"`
	NetworkIO     int64   `json:"network_io"`
	LastUpdated   time.Time `json:"last_updated"`
}

// Configuration structs for various components
type RetryConfig struct {
	MaxAttempts    int           `json:"max_attempts"`
	InitialDelay   time.Duration `json:"initial_delay"`
	MaxDelay       time.Duration `json:"max_delay"`
	BackoffFactor  float64       `json:"backoff_factor"`
	RetryableErrors []string     `json:"retryable_errors"`
}

type TimeoutConfig struct {
	ConnectionTimeout time.Duration `json:"connection_timeout"`
	RequestTimeout    time.Duration `json:"request_timeout"`
	KeepAliveTimeout  time.Duration `json:"keep_alive_timeout"`
}

type SecurityConfig struct {
	EnableTLS       bool      `json:"enable_tls"`
	TLSConfig       *TLSConfig `json:"tls_config"`
	RequireAuth     bool      `json:"require_auth"`
	AllowedOrigins  []string  `json:"allowed_origins"`
	RateLimit       *RateLimitConfig `json:"rate_limit"`
}

type TLSConfig struct {
	CertFile     string `json:"cert_file"`
	KeyFile      string `json:"key_file"`
	CAFile       string `json:"ca_file"`
	ServerName   string `json:"server_name"`
	InsecureSkip bool   `json:"insecure_skip"`
}

type CachingConfig struct {
	Enabled    bool          `json:"enabled"`
	TTL        time.Duration `json:"ttl"`
	MaxSize    int64         `json:"max_size"`
	Strategy   string        `json:"strategy"` // lru, lfu, ttl
	Compress   bool          `json:"compress"`
}

type LoadBalancerMetrics struct {
	RequestCount    int64                  `json:"request_count"`
	ErrorCount      int64                  `json:"error_count"`
	AvgResponseTime float64                `json:"avg_response_time"`
	InstanceMetrics map[string]*RequestMetrics `json:"instance_metrics"`
	LastUpdated     time.Time              `json:"last_updated"`
}

type DistributedStorage interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Exists(key string) bool
	Increment(key string, delta int64) (int64, error)
	GetMultiple(keys []string) (map[string]interface{}, error)
	SetMultiple(items map[string]interface{}, ttl time.Duration) error
}

// NewServiceMesh creates a new service mesh instance
func NewServiceMesh(logger *zap.Logger) *ServiceMesh {
	return &ServiceMesh{
		services:         make(map[string]*ServiceGroup),
		serviceDiscovery: NewServiceDiscovery(logger),
		loadBalancer:     NewLoadBalancer(logger),
		healthChecker:    NewHealthChecker(logger),
		circuitBreaker:   NewCircuitBreakerManager(logger),
		rateLimiter:      NewDistributedRateLimiter(logger),
		logger:           logger,
	}
}

// NewServiceDiscovery creates a new service discovery instance
func NewServiceDiscovery(logger *zap.Logger) *ServiceDiscovery {
	return &ServiceDiscovery{
		registry: make(map[string]*ServiceGroup),
		watchers: make(map[string][]ServiceWatcher),
		hashRing: NewConsistentHashRing(),
		logger:   logger,
	}
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(logger *zap.Logger) *LoadBalancer {
	lb := &LoadBalancer{
		strategies: make(map[string]LoadBalancingStrategy),
		metrics:    &LoadBalancerMetrics{InstanceMetrics: make(map[string]*RequestMetrics)},
		logger:     logger,
	}
	
	// Register built-in strategies
	lb.RegisterStrategy("round_robin", &RoundRobinStrategy{})
	lb.RegisterStrategy("weighted", &WeightedStrategy{
		weights: make(map[string]int),
		current: make(map[string]int),
	})
	lb.RegisterStrategy("least_connections", &LeastConnectionsStrategy{
		connections: make(map[string]int64),
	})
	lb.RegisterStrategy("consistent_hash", &ConsistentHashStrategy{
		hashRing: NewConsistentHashRing(),
	})
	
	return lb
}

// RegisterStrategy registers a new load balancing strategy
func (lb *LoadBalancer) RegisterStrategy(name string, strategy LoadBalancingStrategy) {
	lb.strategies[name] = strategy
}

// SelectInstance selects an instance using the specified strategy
func (lb *LoadBalancer) SelectInstance(serviceName string, strategy string, instances []*ServiceInstance, context *RequestContext) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	
	// Filter healthy instances
	healthyInstances := make([]*ServiceInstance, 0)
	for _, instance := range instances {
		if instance.Health != nil && instance.Health.Status == "healthy" {
			healthyInstances = append(healthyInstances, instance)
		}
	}
	
	if len(healthyInstances) == 0 {
		// Fallback to all instances if none are healthy
		healthyInstances = instances
	}
	
	strategyImpl, exists := lb.strategies[strategy]
	if !exists {
		strategyImpl = lb.strategies["round_robin"] // Default fallback
	}
	
	return strategyImpl.SelectInstance(healthyInstances, context)
}

// Round Robin Strategy Implementation
func (rr *RoundRobinStrategy) SelectInstance(instances []*ServiceInstance, context *RequestContext) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	
	index := atomic.AddUint64(&rr.counter, 1) % uint64(len(instances))
	return instances[index]
}

func (rr *RoundRobinStrategy) UpdateMetrics(instance *ServiceInstance, metrics *RequestMetrics) {
	// Update instance metrics
	if instance.Metrics != nil {
		atomic.AddInt64(&instance.Metrics.RequestCount, 1)
		if metrics.Error != nil {
			atomic.AddInt64(&instance.Metrics.ErrorCount, 1)
		}
		instance.Metrics.LastUpdated = time.Now()
	}
}

func (rr *RoundRobinStrategy) GetName() string {
	return "round_robin"
}

// Weighted Strategy Implementation
func (w *WeightedStrategy) SelectInstance(instances []*ServiceInstance, context *RequestContext) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// Calculate total weight
	totalWeight := 0
	for _, instance := range instances {
		totalWeight += instance.Weight
	}
	
	if totalWeight == 0 {
		// Fallback to round robin if no weights
		return instances[rand.Intn(len(instances))]
	}
	
	// Select based on weight
	target := rand.Intn(totalWeight)
	current := 0
	
	for _, instance := range instances {
		current += instance.Weight
		if current >= target {
			return instance
		}
	}
	
	return instances[0] // Fallback
}

func (w *WeightedStrategy) UpdateMetrics(instance *ServiceInstance, metrics *RequestMetrics) {
	// Update weights based on performance
	if metrics.Error != nil {
		// Decrease weight for failing instances
		if instance.Weight > 1 {
			instance.Weight--
		}
	} else if metrics.Duration < 100*time.Millisecond {
		// Increase weight for fast instances
		if instance.Weight < 10 {
			instance.Weight++
		}
	}
}

func (w *WeightedStrategy) GetName() string {
	return "weighted"
}

// Least Connections Strategy Implementation
func (lc *LeastConnectionsStrategy) SelectInstance(instances []*ServiceInstance, context *RequestContext) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	
	var selectedInstance *ServiceInstance
	minConnections := int64(^uint64(0) >> 1) // Max int64
	
	for _, instance := range instances {
		connections := lc.connections[instance.ID]
		if connections < minConnections {
			minConnections = connections
			selectedInstance = instance
		}
	}
	
	if selectedInstance != nil {
		lc.mu.RUnlock()
		lc.mu.Lock()
		lc.connections[selectedInstance.ID]++
		lc.mu.Unlock()
		lc.mu.RLock()
	}
	
	return selectedInstance
}

func (lc *LeastConnectionsStrategy) UpdateMetrics(instance *ServiceInstance, metrics *RequestMetrics) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	
	// Decrement connection count when request completes
	if lc.connections[instance.ID] > 0 {
		lc.connections[instance.ID]--
	}
}

func (lc *LeastConnectionsStrategy) GetName() string {
	return "least_connections"
}

// Consistent Hash Strategy Implementation
func (ch *ConsistentHashStrategy) SelectInstance(instances []*ServiceInstance, context *RequestContext) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	
	ch.mu.Lock()
	defer ch.mu.Unlock()
	
	// Update hash ring with current instances
	ch.hashRing = NewConsistentHashRing()
	for _, instance := range instances {
		ch.hashRing.Add(instance.ID)
	}
	
	// Use session ID or client IP for consistent hashing
	hashKey := context.SessionID
	if hashKey == "" {
		hashKey = context.ClientIP
	}
	if hashKey == "" {
		hashKey = context.RequestID
	}
	
	selectedID, err := ch.hashRing.Get(hashKey)
	if err != nil {
		return instances[0] // Fallback
	}
	
	// Find instance by ID
	for _, instance := range instances {
		if instance.ID == selectedID {
			return instance
		}
	}
	
	return instances[0] // Fallback
}

func (ch *ConsistentHashStrategy) UpdateMetrics(instance *ServiceInstance, metrics *RequestMetrics) {
	// Consistent hash doesn't need to update selection based on metrics
	// But we can still track instance performance
	if instance.Metrics != nil {
		atomic.AddInt64(&instance.Metrics.RequestCount, 1)
		if metrics.Error != nil {
			atomic.AddInt64(&instance.Metrics.ErrorCount, 1)
		}
	}
}

func (ch *ConsistentHashStrategy) GetName() string {
	return "consistent_hash"
}

// NewConsistentHashRing creates a new consistent hash ring
func NewConsistentHashRing() *ConsistentHashRing {
	return &ConsistentHashRing{
		nodes: make(map[uint32]string),
		keys:  make([]uint32, 0),
	}
}

// Add adds a node to the hash ring
func (chr *ConsistentHashRing) Add(node string) {
	// Create multiple virtual nodes for better distribution
	for i := 0; i < 3; i++ {
		key := chr.hash(fmt.Sprintf("%s:%d", node, i))
		chr.nodes[key] = node
		chr.keys = append(chr.keys, key)
	}
	
	// Sort keys for binary search
	for i := 0; i < len(chr.keys); i++ {
		for j := i + 1; j < len(chr.keys); j++ {
			if chr.keys[i] > chr.keys[j] {
				chr.keys[i], chr.keys[j] = chr.keys[j], chr.keys[i]
			}
		}
	}
}

// Get gets the node for a given key
func (chr *ConsistentHashRing) Get(key string) (string, error) {
	if len(chr.keys) == 0 {
		return "", fmt.Errorf("no nodes available")
	}
	
	hash := chr.hash(key)
	
	// Find the first node with hash >= hash
	idx := 0
	for i, k := range chr.keys {
		if k >= hash {
			idx = i
			break
		}
	}
	
	return chr.nodes[chr.keys[idx]], nil
}

// Remove removes a node from the hash ring
func (chr *ConsistentHashRing) Remove(node string) {
	for i := 0; i < 3; i++ {
		key := chr.hash(fmt.Sprintf("%s:%d", node, i))
		delete(chr.nodes, key)
		
		// Remove from keys slice
		for j, k := range chr.keys {
			if k == key {
				chr.keys = append(chr.keys[:j], chr.keys[j+1:]...)
				break
			}
		}
	}
}

// hash creates a hash for the given key
func (chr *ConsistentHashRing) hash(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(logger *zap.Logger) *HealthChecker {
	return &HealthChecker{
		checkInterval: 30 * time.Second,
		timeout:       5 * time.Second,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		checks:   make(map[string]*HealthCheckConfig),
		results:  make(map[string]*HealthCheckResult),
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// StartHealthChecks begins periodic health checking
func (hc *HealthChecker) StartHealthChecks() {
	go func() {
		ticker := time.NewTicker(hc.checkInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				hc.performHealthChecks()
			case <-hc.stopChan:
				return
			}
		}
	}()
}

// performHealthChecks checks all registered services
func (hc *HealthChecker) performHealthChecks() {
	hc.mu.RLock()
	checks := make(map[string]*HealthCheckConfig)
	for k, v := range hc.checks {
		checks[k] = v
	}
	hc.mu.RUnlock()
	
	for instanceID, config := range checks {
		go hc.checkInstance(instanceID, config)
	}
}

// checkInstance performs health check on a single instance
func (hc *HealthChecker) checkInstance(instanceID string, config *HealthCheckConfig) {
	result := &HealthCheckResult{
		Timestamp: time.Now(),
		Attempts:  1,
	}
	
	start := time.Now()
	
	switch config.Type {
	case "http":
		err := hc.httpHealthCheck(instanceID, config, result)
		result.Error = ""
		if err != nil {
			result.Error = err.Error()
			result.Status = "unhealthy"
		} else {
			result.Status = "healthy"
		}
	case "tcp":
		err := hc.tcpHealthCheck(instanceID, config, result)
		if err != nil {
			result.Error = err.Error()
			result.Status = "unhealthy"
		} else {
			result.Status = "healthy"
		}
	default:
		result.Status = "unknown"
		result.Error = "unsupported health check type"
	}
	
	result.ResponseTime = time.Since(start).Milliseconds()
	
	hc.mu.Lock()
	hc.results[instanceID] = result
	hc.mu.Unlock()
	
	hc.logger.Debug("Health check completed",
		zap.String("instance", instanceID),
		zap.String("status", result.Status),
		zap.Int64("response_time", result.ResponseTime),
	)
}

// httpHealthCheck performs HTTP health check
func (hc *HealthChecker) httpHealthCheck(instanceID string, config *HealthCheckConfig, result *HealthCheckResult) error {
	url := fmt.Sprintf("http://%s%s", instanceID, config.Path)
	
	req, err := http.NewRequest(config.Method, url, nil)
	if err != nil {
		return err
	}
	
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}
	
	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != config.SuccessCode {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}
	
	return nil
}

// tcpHealthCheck performs TCP health check
func (hc *HealthChecker) tcpHealthCheck(instanceID string, config *HealthCheckConfig, result *HealthCheckResult) error {
	conn, err := net.DialTimeout("tcp", instanceID, config.Timeout)
	if err != nil {
		return err
	}
	defer conn.Close()
	
	return nil
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(logger *zap.Logger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		config: &CircuitBreakerConfig{
			FailureThreshold: 5,
			SuccessThreshold: 3,
			Timeout:          60 * time.Second,
			RecoveryTimeout:  30 * time.Second,
			MaxRequests:      10,
			FailureRate:      0.5,
		},
		logger: logger,
	}
}

// NewDistributedRateLimiter creates a new distributed rate limiter
func NewDistributedRateLimiter(logger *zap.Logger) *DistributedRateLimiter {
	return &DistributedRateLimiter{
		limiters: make(map[string]*TokenBucket),
		logger:   logger,
	}
}

// RegisterService registers a new service group
func (sm *ServiceMesh) RegisterService(service *ServiceGroup) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	service.CreatedAt = time.Now()
	service.UpdatedAt = time.Now()
	
	sm.services[service.Name] = service
	
	// Register with service discovery
	for _, instance := range service.Instances {
		err := sm.serviceDiscovery.RegisterInstance(service.Name, instance)
		if err != nil {
			sm.logger.Error("Failed to register service instance",
				zap.String("service", service.Name),
				zap.String("instance", instance.ID),
				zap.Error(err),
			)
		}
	}
	
	sm.logger.Info("Service registered",
		zap.String("service", service.Name),
		zap.Int("instances", len(service.Instances)),
	)
	
	return nil
}

// GetService retrieves a service group by name
func (sm *ServiceMesh) GetService(name string) (*ServiceGroup, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	service, exists := sm.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}
	
	return service, nil
}

// RouteRequest routes a request to an appropriate service instance
func (sm *ServiceMesh) RouteRequest(serviceName string, context *RequestContext) (*ServiceInstance, error) {
	service, err := sm.GetService(serviceName)
	if err != nil {
		return nil, err
	}
	
	// Check circuit breaker
	breaker := sm.circuitBreaker.GetBreaker(serviceName)
	if breaker != nil && breaker.State() == StateOpen {
		return nil, fmt.Errorf("circuit breaker is open for service %s", serviceName)
	}
	
	// Check rate limiting
	allowed, err := sm.rateLimiter.Allow(serviceName, context.ClientIP)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("rate limit exceeded for service %s", serviceName)
	}
	
	// Select instance using load balancer
	strategy := "round_robin"
	if service.Config != nil && service.Config.LoadBalancing != "" {
		strategy = service.Config.LoadBalancing
	}
	
	instance := sm.loadBalancer.SelectInstance(serviceName, strategy, service.Instances, context)
	if instance == nil {
		return nil, fmt.Errorf("no healthy instances available for service %s", serviceName)
	}
	
	return instance, nil
}

// Allow checks if a request is allowed by the rate limiter
func (drl *DistributedRateLimiter) Allow(service, clientID string) (bool, error) {
	key := fmt.Sprintf("%s:%s", service, clientID)
	
	drl.mu.RLock()
	bucket, exists := drl.limiters[key]
	drl.mu.RUnlock()
	
	if !exists {
		drl.mu.Lock()
		bucket = &TokenBucket{
			capacity:   100,
			tokens:     100,
			refillRate: 10,
			lastRefill: time.Now(),
		}
		drl.limiters[key] = bucket
		drl.mu.Unlock()
	}
	
	return bucket.Allow(), nil
}

// Allow checks if a token is available in the bucket
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	
	// Refill tokens
	tokensToAdd := int64(elapsed * float64(tb.refillRate))
	tb.tokens = minInt64(tb.capacity, tb.tokens+tokensToAdd)
	tb.lastRefill = now
	
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	
	return false
}

// Utility functions
func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// RegisterInstance registers a service instance with discovery
func (sd *ServiceDiscovery) RegisterInstance(serviceName string, instance *ServiceInstance) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	
	service, exists := sd.registry[serviceName]
	if !exists {
		service = &ServiceGroup{
			Name:      serviceName,
			Instances: make([]*ServiceInstance, 0),
			CreatedAt: time.Now(),
		}
		sd.registry[serviceName] = service
	}
	
	instance.RegisteredAt = time.Now()
	instance.LastSeen = time.Now()
	
	service.Instances = append(service.Instances, instance)
	service.UpdatedAt = time.Now()
	
	// Add to hash ring for consistent hashing
	sd.hashRing.Add(instance.ID)
	
	// Notify watchers
	for _, watcher := range sd.watchers[serviceName] {
		watcher.OnServiceRegistered(instance)
	}
	
	sd.logger.Info("Service instance registered",
		zap.String("service", serviceName),
		zap.String("instance", instance.ID),
		zap.String("address", fmt.Sprintf("%s:%d", instance.Address, instance.Port)),
	)
	
	return nil
}

// GetBreaker gets or creates a circuit breaker for a service
func (cbm *CircuitBreakerManager) GetBreaker(serviceName string) *CircuitBreaker {
	cbm.mu.RLock()
	breaker, exists := cbm.breakers[serviceName]
	cbm.mu.RUnlock()
	
	if !exists {
		cbm.mu.Lock()
		breaker = &CircuitBreaker{
			name:   serviceName,
			state:  StateClosed,
			config: cbm.config,
		}
		cbm.breakers[serviceName] = breaker
		cbm.mu.Unlock()
	}
	
	return breaker
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns comprehensive metrics for the service mesh
func (sm *ServiceMesh) GetMetrics() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	metrics := map[string]interface{}{
		"total_services":  len(sm.services),
		"total_instances": 0,
		"healthy_instances": 0,
		"services": make(map[string]interface{}),
		"load_balancer": sm.loadBalancer.metrics,
		"timestamp": time.Now(),
	}
	
	totalInstances := 0
	healthyInstances := 0
	
	services := make(map[string]interface{})
	for name, service := range sm.services {
		serviceMetrics := map[string]interface{}{
			"instance_count": len(service.Instances),
			"healthy_count":  0,
			"instances":     make([]interface{}, 0),
		}
		
		healthyCount := 0
		instances := make([]interface{}, 0)
		
		for _, instance := range service.Instances {
			totalInstances++
			
			instanceData := map[string]interface{}{
				"id":      instance.ID,
				"address": fmt.Sprintf("%s:%d", instance.Address, instance.Port),
				"health":  instance.Health,
				"metrics": instance.Metrics,
			}
			
			if instance.Health != nil && instance.Health.Status == "healthy" {
				healthyCount++
				healthyInstances++
			}
			
			instances = append(instances, instanceData)
		}
		
		serviceMetrics["healthy_count"] = healthyCount
		serviceMetrics["instances"] = instances
		services[name] = serviceMetrics
	}
	
	metrics["total_instances"] = totalInstances
	metrics["healthy_instances"] = healthyInstances
	metrics["services"] = services
	
	return metrics
}