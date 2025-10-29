package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventStore implements event sourcing pattern for distributed systems
type EventStore struct {
	events    []Event
	snapshots map[string]Snapshot
	handlers  map[string][]EventHandler
	mutex     sync.RWMutex
	streams   map[string]*EventStream
}

// Event represents a domain event in the system
type Event struct {
	ID            string                 `json:"id"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	EventType     string                 `json:"event_type"`
	EventVersion  int                    `json:"event_version"`
	Data          map[string]interface{} `json:"data"`
	Metadata      map[string]interface{} `json:"metadata"`
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlation_id"`
	CausationID   string                 `json:"causation_id"`
}

// EventStream represents a stream of events for real-time processing
type EventStream struct {
	Name        string
	Partition   int
	Offset      int64
	Subscribers []chan Event
	mutex       sync.RWMutex
}

// Snapshot represents aggregate state snapshots for performance optimization
type Snapshot struct {
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int                    `json:"version"`
	Data          map[string]interface{} `json:"data"`
	Timestamp     time.Time              `json:"timestamp"`
}

// EventHandler processes events asynchronously
type EventHandler func(ctx context.Context, event Event) error

// Command represents a command in CQRS pattern
type Command struct {
	ID            string                 `json:"id"`
	CommandType   string                 `json:"command_type"`
	AggregateID   string                 `json:"aggregate_id"`
	Data          map[string]interface{} `json:"data"`
	Metadata      map[string]interface{} `json:"metadata"`
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlation_id"`
}

// Query represents a query in CQRS pattern
type Query struct {
	ID        string                 `json:"id"`
	QueryType string                 `json:"query_type"`
	Filters   map[string]interface{} `json:"filters"`
	Sort      []string               `json:"sort"`
	Limit     int                    `json:"limit"`
	Offset    int                    `json:"offset"`
}

// CommandHandler processes commands
type CommandHandler func(ctx context.Context, cmd Command) ([]Event, error)

// QueryHandler processes queries
type QueryHandler func(ctx context.Context, query Query) (interface{}, error)

// Saga represents a distributed transaction saga
type Saga struct {
	ID            string                 `json:"id"`
	SagaType      string                 `json:"saga_type"`
	State         string                 `json:"state"`
	Steps         []SagaStep             `json:"steps"`
	CurrentStep   int                    `json:"current_step"`
	Data          map[string]interface{} `json:"data"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	CorrelationID string                 `json:"correlation_id"`
}

// SagaStep represents a step in a saga
type SagaStep struct {
	ID              string                 `json:"id"`
	StepType        string                 `json:"step_type"`
	Command         Command                `json:"command"`
	CompensateCmd   *Command               `json:"compensate_command,omitempty"`
	Status          string                 `json:"status"` // pending, completed, failed, compensated
	RetryCount      int                    `json:"retry_count"`
	MaxRetries      int                    `json:"max_retries"`
	Data            map[string]interface{} `json:"data"`
	ExecutedAt      *time.Time             `json:"executed_at"`
	CompensatedAt   *time.Time             `json:"compensated_at"`
}

// EventBus manages event distribution and messaging
type EventBus struct {
	topics      map[string]*Topic
	subscribers map[string][]Subscriber
	mutex       sync.RWMutex
}

// Topic represents a message topic for pub/sub
type Topic struct {
	Name       string
	Partitions int
	Messages   []Message
	mutex      sync.RWMutex
}

// Message represents a message in the event bus
type Message struct {
	ID        string                 `json:"id"`
	Topic     string                 `json:"topic"`
	Partition int                    `json:"partition"`
	Key       string                 `json:"key"`
	Value     []byte                 `json:"value"`
	Headers   map[string]string      `json:"headers"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Subscriber represents an event subscriber
type Subscriber struct {
	ID       string
	Topic    string
	Handler  func(Message) error
	GroupID  string
	AutoAck  bool
}

// NewEventStore creates a new event store
func NewEventStore() *EventStore {
	return &EventStore{
		events:    make([]Event, 0),
		snapshots: make(map[string]Snapshot),
		handlers:  make(map[string][]EventHandler),
		streams:   make(map[string]*EventStream),
	}
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		topics:      make(map[string]*Topic),
		subscribers: make(map[string][]Subscriber),
	}
}

// AppendEvent appends an event to the event store
func (es *EventStore) AppendEvent(ctx context.Context, event Event) error {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	event.ID = uuid.New().String()
	event.Timestamp = time.Now()

	es.events = append(es.events, event)

	// Publish to event streams
	if stream, exists := es.streams[event.EventType]; exists {
		stream.Publish(event)
	}

	// Trigger event handlers
	if handlers, exists := es.handlers[event.EventType]; exists {
		for _, handler := range handlers {
			go func(h EventHandler, e Event) {
				if err := h(ctx, e); err != nil {
					// Log error - in production, this would use proper logging
					fmt.Printf("Event handler error: %v\n", err)
				}
			}(handler, event)
		}
	}

	return nil
}

// GetEvents retrieves events for an aggregate
func (es *EventStore) GetEvents(aggregateID string, fromVersion int) ([]Event, error) {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	var events []Event
	for _, event := range es.events {
		if event.AggregateID == aggregateID && event.EventVersion >= fromVersion {
			events = append(events, event)
		}
	}

	return events, nil
}

// SaveSnapshot saves a snapshot for an aggregate
func (es *EventStore) SaveSnapshot(snapshot Snapshot) error {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	snapshot.Timestamp = time.Now()
	es.snapshots[snapshot.AggregateID] = snapshot

	return nil
}

// GetSnapshot retrieves the latest snapshot for an aggregate
func (es *EventStore) GetSnapshot(aggregateID string) (*Snapshot, error) {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	if snapshot, exists := es.snapshots[aggregateID]; exists {
		return &snapshot, nil
	}

	return nil, fmt.Errorf("snapshot not found for aggregate %s", aggregateID)
}

// RegisterEventHandler registers an event handler
func (es *EventStore) RegisterEventHandler(eventType string, handler EventHandler) {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	if _, exists := es.handlers[eventType]; !exists {
		es.handlers[eventType] = make([]EventHandler, 0)
	}

	es.handlers[eventType] = append(es.handlers[eventType], handler)
}

// CreateStream creates a new event stream
func (es *EventStore) CreateStream(name string) *EventStream {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	stream := &EventStream{
		Name:        name,
		Subscribers: make([]chan Event, 0),
	}

	es.streams[name] = stream
	return stream
}

// Subscribe to an event stream
func (stream *EventStream) Subscribe() <-chan Event {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	ch := make(chan Event, 100) // Buffered channel
	stream.Subscribers = append(stream.Subscribers, ch)

	return ch
}

// Publish publishes an event to the stream
func (stream *EventStream) Publish(event Event) {
	stream.mutex.RLock()
	defer stream.mutex.RUnlock()

	for _, subscriber := range stream.Subscribers {
		select {
		case subscriber <- event:
		default:
			// Channel is full, skip (in production, handle this better)
		}
	}
}

// CreateTopic creates a new topic in the event bus
func (eb *EventBus) CreateTopic(name string, partitions int) *Topic {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	topic := &Topic{
		Name:       name,
		Partitions: partitions,
		Messages:   make([]Message, 0),
	}

	eb.topics[name] = topic
	return topic
}

// Publish publishes a message to a topic
func (eb *EventBus) Publish(topicName string, message Message) error {
	eb.mutex.RLock()
	topic, exists := eb.topics[topicName]
	eb.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("topic %s does not exist", topicName)
	}

	topic.mutex.Lock()
	defer topic.mutex.Unlock()

	message.ID = uuid.New().String()
	message.Topic = topicName
	message.Timestamp = time.Now()

	topic.Messages = append(topic.Messages, message)

	// Notify subscribers
	if subscribers, exists := eb.subscribers[topicName]; exists {
		for _, subscriber := range subscribers {
			go func(sub Subscriber, msg Message) {
				if err := sub.Handler(msg); err != nil {
					fmt.Printf("Subscriber error: %v\n", err)
				}
			}(subscriber, message)
		}
	}

	return nil
}

// Subscribe subscribes to a topic
func (eb *EventBus) Subscribe(topicName string, subscriber Subscriber) error {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if _, exists := eb.topics[topicName]; !exists {
		return fmt.Errorf("topic %s does not exist", topicName)
	}

	if _, exists := eb.subscribers[topicName]; !exists {
		eb.subscribers[topicName] = make([]Subscriber, 0)
	}

	subscriber.Topic = topicName
	eb.subscribers[topicName] = append(eb.subscribers[topicName], subscriber)

	return nil
}

// ProcessSaga processes a saga step by step
func ProcessSaga(ctx context.Context, saga *Saga) error {
	if saga.CurrentStep >= len(saga.Steps) {
		saga.State = "completed"
		return nil
	}

	step := &saga.Steps[saga.CurrentStep]
	step.Status = "executing"

	// Execute the step command
	err := executeCommand(ctx, step.Command)
	if err != nil {
		step.Status = "failed"
		step.RetryCount++

		if step.RetryCount < step.MaxRetries {
			// Retry logic
			return ProcessSaga(ctx, saga)
		} else {
			// Start compensation
			return compensateSaga(ctx, saga)
		}
	}

	step.Status = "completed"
	now := time.Now()
	step.ExecutedAt = &now
	saga.CurrentStep++
	saga.UpdatedAt = time.Now()

	// Continue to next step
	return ProcessSaga(ctx, saga)
}

// compensateSaga compensates a failed saga
func compensateSaga(ctx context.Context, saga *Saga) error {
	saga.State = "compensating"

	// Execute compensation commands in reverse order
	for i := saga.CurrentStep - 1; i >= 0; i-- {
		step := &saga.Steps[i]
		if step.Status == "completed" && step.CompensateCmd != nil {
			err := executeCommand(ctx, *step.CompensateCmd)
			if err != nil {
				// Log compensation failure
				fmt.Printf("Compensation failed for step %s: %v\n", step.ID, err)
			} else {
				step.Status = "compensated"
				now := time.Now()
				step.CompensatedAt = &now
			}
		}
	}

	saga.State = "compensated"
	saga.UpdatedAt = time.Now()
	return nil
}

// executeCommand executes a command (mock implementation)
func executeCommand(ctx context.Context, cmd Command) error {
	// Mock command execution - in production this would route to appropriate handlers
	fmt.Printf("Executing command: %s for aggregate: %s\n", cmd.CommandType, cmd.AggregateID)
	return nil
}

// ProjectionEngine handles read model projections
type ProjectionEngine struct {
	projections map[string]Projection
	mutex       sync.RWMutex
}

// Projection represents a read model projection
type Projection struct {
	Name         string
	EventTypes   []string
	Handler      func(Event) error
	State        map[string]interface{}
	LastPosition int64
}

// NewProjectionEngine creates a new projection engine
func NewProjectionEngine() *ProjectionEngine {
	return &ProjectionEngine{
		projections: make(map[string]Projection),
	}
}

// RegisterProjection registers a new projection
func (pe *ProjectionEngine) RegisterProjection(projection Projection) {
	pe.mutex.Lock()
	defer pe.mutex.Unlock()

	pe.projections[projection.Name] = projection
}

// ProcessEvent processes an event through all relevant projections
func (pe *ProjectionEngine) ProcessEvent(event Event) error {
	pe.mutex.RLock()
	defer pe.mutex.RUnlock()

	for _, projection := range pe.projections {
		for _, eventType := range projection.EventTypes {
			if eventType == event.EventType || eventType == "*" {
				if err := projection.Handler(event); err != nil {
					return fmt.Errorf("projection %s failed: %w", projection.Name, err)
				}
				break
			}
		}
	}

	return nil
}