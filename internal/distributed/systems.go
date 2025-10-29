package distributed

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// DistributedCache implements advanced caching patterns
type DistributedCache struct {
	nodes      map[string]*CacheNode
	consistent *ConsistentHash
	replication int
	mutex      sync.RWMutex
}

// CacheNode represents a cache node in the distributed system
type CacheNode struct {
	ID       string
	Address  string
	Data     map[string]*CacheEntry
	Stats    *NodeStats
	Healthy  bool
	mutex    sync.RWMutex
}

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Key        string                 `json:"key"`
	Value      interface{}            `json:"value"`
	TTL        time.Duration          `json:"ttl"`
	CreatedAt  time.Time              `json:"created_at"`
	AccessedAt time.Time              `json:"accessed_at"`
	AccessCount int64                 `json:"access_count"`
	Tags       []string               `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// NodeStats tracks cache node performance metrics
type NodeStats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	Sets        int64     `json:"sets"`
	Deletes     int64     `json:"deletes"`
	Evictions   int64     `json:"evictions"`
	Connections int64     `json:"connections"`
	Memory      int64     `json:"memory_bytes"`
	LastReset   time.Time `json:"last_reset"`
}

// ConsistentHash implements consistent hashing for distributed caching
type ConsistentHash struct {
	ring     map[uint32]string
	nodes    map[string]bool
	replicas int
	mutex    sync.RWMutex
}

// MessageQueue implements advanced message queuing patterns
type MessageQueue struct {
	queues      map[string]*Queue
	exchanges   map[string]*Exchange
	bindings    map[string][]Binding
	dlq         *Queue // Dead Letter Queue
	consumers   map[string]*Consumer
	publishers  map[string]*Publisher
	mutex       sync.RWMutex
}

// Queue represents a message queue
type Queue struct {
	Name         string           `json:"name"`
	Messages     []*QueueMessage  `json:"messages"`
	Consumers    []string         `json:"consumers"`
	MaxLength    int              `json:"max_length"`
	TTL          time.Duration    `json:"ttl"`
	DeadLetter   string           `json:"dead_letter_queue"`
	Durable      bool             `json:"durable"`
	AutoDelete   bool             `json:"auto_delete"`
	Stats        *QueueStats      `json:"stats"`
	mutex        sync.RWMutex
}

// Exchange represents a message exchange
type Exchange struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"` // direct, fanout, topic, headers
	Durable    bool              `json:"durable"`
	AutoDelete bool              `json:"auto_delete"`
	Internal   bool              `json:"internal"`
	Args       map[string]interface{} `json:"args"`
}

// Binding represents a queue-exchange binding
type Binding struct {
	Queue      string            `json:"queue"`
	Exchange   string            `json:"exchange"`
	RoutingKey string            `json:"routing_key"`
	Args       map[string]interface{} `json:"args"`
}

// QueueMessage represents a message in the queue
type QueueMessage struct {
	ID          string                 `json:"id"`
	Body        []byte                 `json:"body"`
	Headers     map[string]interface{} `json:"headers"`
	RoutingKey  string                 `json:"routing_key"`
	Priority    int                    `json:"priority"`
	Timestamp   time.Time              `json:"timestamp"`
	Expiration  *time.Time             `json:"expiration"`
	DeliveryTag uint64                 `json:"delivery_tag"`
	Redelivered bool                   `json:"redelivered"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
}

// Consumer represents a message consumer
type Consumer struct {
	ID          string                              `json:"id"`
	Queue       string                              `json:"queue"`
	Tag         string                              `json:"tag"`
	Handler     func(*QueueMessage) error           `json:"-"`
	AutoAck     bool                                `json:"auto_ack"`
	Exclusive   bool                                `json:"exclusive"`
	NoLocal     bool                                `json:"no_local"`
	NoWait      bool                                `json:"no_wait"`
	Args        map[string]interface{}              `json:"args"`
	Stats       *ConsumerStats                      `json:"stats"`
}

// Publisher represents a message publisher
type Publisher struct {
	ID          string                 `json:"id"`
	Exchange    string                 `json:"exchange"`
	Mandatory   bool                   `json:"mandatory"`
	Immediate   bool                   `json:"immediate"`
	Stats       *PublisherStats        `json:"stats"`
}

// QueueStats tracks queue performance metrics
type QueueStats struct {
	Messages      int64     `json:"messages"`
	Consumers     int64     `json:"consumers"`
	Published     int64     `json:"published"`
	Delivered     int64     `json:"delivered"`
	Acknowledged  int64     `json:"acknowledged"`
	Redelivered   int64     `json:"redelivered"`
	Rejected      int64     `json:"rejected"`
	LastActivity  time.Time `json:"last_activity"`
}

// ConsumerStats tracks consumer performance metrics
type ConsumerStats struct {
	Delivered    int64     `json:"delivered"`
	Acknowledged int64     `json:"acknowledged"`
	Rejected     int64     `json:"rejected"`
	LastDelivery time.Time `json:"last_delivery"`
}

// PublisherStats tracks publisher performance metrics
type PublisherStats struct {
	Published    int64     `json:"published"`
	Confirmed    int64     `json:"confirmed"`
	Returned     int64     `json:"returned"`
	LastPublish  time.Time `json:"last_publish"`
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	name           string
	state          string // closed, open, half-open
	failureCount   int64
	successCount   int64
	failureThreshold int64
	resetTimeout   time.Duration
	lastFailTime   time.Time
	mutex          sync.RWMutex
}

// NewDistributedCache creates a new distributed cache
func NewDistributedCache(replication int) *DistributedCache {
	return &DistributedCache{
		nodes:       make(map[string]*CacheNode),
		consistent:  NewConsistentHash(100), // 100 virtual nodes per physical node
		replication: replication,
	}
}

// NewConsistentHash creates a new consistent hash ring
func NewConsistentHash(replicas int) *ConsistentHash {
	return &ConsistentHash{
		ring:     make(map[uint32]string),
		nodes:    make(map[string]bool),
		replicas: replicas,
	}
}

// NewMessageQueue creates a new message queue system
func NewMessageQueue() *MessageQueue {
	dlq := &Queue{
		Name:      "dead-letter-queue",
		Messages:  make([]*QueueMessage, 0),
		Consumers: make([]string, 0),
		MaxLength: 10000,
		Durable:   true,
		Stats:     &QueueStats{},
	}

	return &MessageQueue{
		queues:     make(map[string]*Queue),
		exchanges:  make(map[string]*Exchange),
		bindings:   make(map[string][]Binding),
		dlq:        dlq,
		consumers:  make(map[string]*Consumer),
		publishers: make(map[string]*Publisher),
	}
}

// AddNode adds a cache node to the distributed cache
func (dc *DistributedCache) AddNode(nodeID, address string) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	node := &CacheNode{
		ID:      nodeID,
		Address: address,
		Data:    make(map[string]*CacheEntry),
		Stats:   &NodeStats{LastReset: time.Now()},
		Healthy: true,
	}

	dc.nodes[nodeID] = node
	dc.consistent.AddNode(nodeID)
}

// Get retrieves a value from the distributed cache
func (dc *DistributedCache) Get(ctx context.Context, key string) (interface{}, error) {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	nodeID := dc.consistent.GetNode(key)
	node, exists := dc.nodes[nodeID]
	if !exists || !node.Healthy {
		return nil, fmt.Errorf("node %s not available", nodeID)
	}

	node.mutex.RLock()
	defer node.mutex.RUnlock()

	entry, exists := node.Data[key]
	if !exists {
		node.Stats.Misses++
		return nil, fmt.Errorf("key %s not found", key)
	}

	// Check TTL
	if entry.TTL > 0 && time.Since(entry.CreatedAt) > entry.TTL {
		delete(node.Data, key)
		node.Stats.Evictions++
		return nil, fmt.Errorf("key %s expired", key)
	}

	entry.AccessedAt = time.Now()
	entry.AccessCount++
	node.Stats.Hits++

	return entry.Value, nil
}

// Set stores a value in the distributed cache
func (dc *DistributedCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	nodeID := dc.consistent.GetNode(key)
	node, exists := dc.nodes[nodeID]
	if !exists || !node.Healthy {
		return fmt.Errorf("node %s not available", nodeID)
	}

	node.mutex.Lock()
	defer node.mutex.Unlock()

	entry := &CacheEntry{
		Key:        key,
		Value:      value,
		TTL:        ttl,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		AccessCount: 1,
		Tags:       make([]string, 0),
		Metadata:   make(map[string]interface{}),
	}

	node.Data[key] = entry
	node.Stats.Sets++

	// Replicate to other nodes if replication > 1
	if dc.replication > 1 {
		dc.replicate(key, entry)
	}

	return nil
}

// replicate replicates cache entry to replica nodes
func (dc *DistributedCache) replicate(key string, entry *CacheEntry) {
	// Get replica nodes
	replicaNodes := dc.consistent.GetReplicas(key, dc.replication-1)
	
	for _, nodeID := range replicaNodes {
		if node, exists := dc.nodes[nodeID]; exists && node.Healthy {
			go func(n *CacheNode, k string, e *CacheEntry) {
				n.mutex.Lock()
				defer n.mutex.Unlock()
				n.Data[k] = e
			}(node, key, entry)
		}
	}
}

// AddNode adds a node to the consistent hash ring
func (ch *ConsistentHash) AddNode(nodeID string) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	ch.nodes[nodeID] = true
	for i := 0; i < ch.replicas; i++ {
		virtualKey := fmt.Sprintf("%s:%d", nodeID, i)
		hash := ch.hash(virtualKey)
		ch.ring[hash] = nodeID
	}
}

// GetNode gets the node responsible for a key
func (ch *ConsistentHash) GetNode(key string) string {
	ch.mutex.RLock()
	defer ch.mutex.RUnlock()

	if len(ch.ring) == 0 {
		return ""
	}

	hash := ch.hash(key)
	
	// Find the first node with hash >= key hash
	for h, node := range ch.ring {
		if h >= hash {
			return node
		}
	}

	// Wrap around to the first node
	var minHash uint32 = ^uint32(0)
	var minNode string
	for h, node := range ch.ring {
		if h < minHash {
			minHash = h
			minNode = node
		}
	}

	return minNode
}

// GetReplicas gets replica nodes for a key
func (ch *ConsistentHash) GetReplicas(key string, count int) []string {
	ch.mutex.RLock()
	defer ch.mutex.RUnlock()

	if count <= 0 || len(ch.nodes) <= 1 {
		return []string{}
	}

	replicas := make([]string, 0, count)
	seen := make(map[string]bool)
	hash := ch.hash(key)

	// Find nodes after the primary node
	for i := 0; i < count && len(replicas) < count; i++ {
		for h, node := range ch.ring {
			if h > hash && !seen[node] {
				replicas = append(replicas, node)
				seen[node] = true
				hash = h
				break
			}
		}
	}

	return replicas
}

// hash function for consistent hashing
func (ch *ConsistentHash) hash(key string) uint32 {
	// Simple hash function - in production use a better one like FNV
	var hash uint32
	for _, c := range key {
		hash = hash*31 + uint32(c)
	}
	return hash
}

// DeclareQueue declares a new queue
func (mq *MessageQueue) DeclareQueue(name string, durable, autoDelete bool, maxLength int, ttl time.Duration) *Queue {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	queue := &Queue{
		Name:       name,
		Messages:   make([]*QueueMessage, 0),
		Consumers:  make([]string, 0),
		MaxLength:  maxLength,
		TTL:        ttl,
		Durable:    durable,
		AutoDelete: autoDelete,
		Stats:      &QueueStats{},
	}

	mq.queues[name] = queue
	return queue
}

// DeclareExchange declares a new exchange
func (mq *MessageQueue) DeclareExchange(name, exchangeType string, durable, autoDelete, internal bool) *Exchange {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	exchange := &Exchange{
		Name:       name,
		Type:       exchangeType,
		Durable:    durable,
		AutoDelete: autoDelete,
		Internal:   internal,
		Args:       make(map[string]interface{}),
	}

	mq.exchanges[name] = exchange
	return exchange
}

// BindQueue binds a queue to an exchange
func (mq *MessageQueue) BindQueue(queueName, exchangeName, routingKey string) error {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	if _, exists := mq.queues[queueName]; !exists {
		return fmt.Errorf("queue %s does not exist", queueName)
	}

	if _, exists := mq.exchanges[exchangeName]; !exists {
		return fmt.Errorf("exchange %s does not exist", exchangeName)
	}

	binding := Binding{
		Queue:      queueName,
		Exchange:   exchangeName,
		RoutingKey: routingKey,
		Args:       make(map[string]interface{}),
	}

	if _, exists := mq.bindings[exchangeName]; !exists {
		mq.bindings[exchangeName] = make([]Binding, 0)
	}

	mq.bindings[exchangeName] = append(mq.bindings[exchangeName], binding)
	return nil
}

// Publish publishes a message to an exchange
func (mq *MessageQueue) Publish(exchangeName, routingKey string, body []byte, headers map[string]interface{}) error {
	mq.mutex.RLock()
	exchange, exists := mq.exchanges[exchangeName]
	bindings := mq.bindings[exchangeName]
	mq.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("exchange %s does not exist", exchangeName)
	}

	message := &QueueMessage{
		ID:         fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Body:       body,
		Headers:    headers,
		RoutingKey: routingKey,
		Timestamp:  time.Now(),
		MaxRetries: 3,
	}

	// Route message based on exchange type
	targetQueues := mq.routeMessage(exchange, bindings, routingKey)

	for _, queueName := range targetQueues {
		if err := mq.enqueue(queueName, message); err != nil {
			return fmt.Errorf("failed to enqueue to %s: %w", queueName, err)
		}
	}

	return nil
}

// routeMessage routes a message to appropriate queues based on exchange type
func (mq *MessageQueue) routeMessage(exchange *Exchange, bindings []Binding, routingKey string) []string {
	var targetQueues []string

	switch exchange.Type {
	case "direct":
		for _, binding := range bindings {
			if binding.RoutingKey == routingKey {
				targetQueues = append(targetQueues, binding.Queue)
			}
		}
	case "fanout":
		for _, binding := range bindings {
			targetQueues = append(targetQueues, binding.Queue)
		}
	case "topic":
		for _, binding := range bindings {
			if mq.matchTopicPattern(binding.RoutingKey, routingKey) {
				targetQueues = append(targetQueues, binding.Queue)
			}
		}
	}

	return targetQueues
}

// matchTopicPattern matches topic routing patterns
func (mq *MessageQueue) matchTopicPattern(pattern, routingKey string) bool {
	// Simple topic matching - * matches one word, # matches zero or more words
	// In production, use a more sophisticated pattern matching algorithm
	return pattern == routingKey || pattern == "#" || pattern == "*"
}

// enqueue adds a message to a queue
func (mq *MessageQueue) enqueue(queueName string, message *QueueMessage) error {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	queue, exists := mq.queues[queueName]
	if !exists {
		return fmt.Errorf("queue %s does not exist", queueName)
	}

	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	// Check queue length limit
	if queue.MaxLength > 0 && len(queue.Messages) >= queue.MaxLength {
		// Move oldest message to DLQ
		oldestMsg := queue.Messages[0]
		queue.Messages = queue.Messages[1:]
		mq.dlq.Messages = append(mq.dlq.Messages, oldestMsg)
	}

	queue.Messages = append(queue.Messages, message)
	queue.Stats.Messages++
	queue.Stats.Published++
	queue.Stats.LastActivity = time.Now()

	return nil
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, failureThreshold int64, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		state:            "closed",
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.state == "open" {
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = "half-open"
			cb.successCount = 0
		} else {
			return nil, fmt.Errorf("circuit breaker %s is open", cb.name)
		}
	}

	result, err := fn()

	if err != nil {
		cb.failureCount++
		cb.lastFailTime = time.Now()

		if cb.state == "half-open" || cb.failureCount >= cb.failureThreshold {
			cb.state = "open"
		}

		return nil, err
	}

	cb.successCount++
	if cb.state == "half-open" && cb.successCount >= 3 {
		cb.state = "closed"
		cb.failureCount = 0
	}

	return result, nil
}