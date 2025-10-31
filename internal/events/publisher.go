package events

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// EventPublisher handles publishing events to NATS
type EventPublisher struct {
	conn   *nats.Conn
	logger *zap.Logger
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(natsURL string, logger *zap.Logger) (*EventPublisher, error) {
	conn, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &EventPublisher{
		conn:   conn,
		logger: logger,
	}, nil
}

// Close closes the NATS connection
func (ep *EventPublisher) Close() {
	if ep.conn != nil {
		ep.conn.Close()
	}
}

// Publish publishes an event to NATS
func (ep *EventPublisher) Publish(subject string, event interface{}) error {
	if ep.conn == nil {
		ep.logger.Warn("NATS not connected, skipping event", zap.String("subject", subject))
		return nil
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := ep.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	ep.logger.Info("Published event", zap.String("subject", subject), zap.String("event", fmt.Sprintf("%T", event)))
	return nil
}

// Event types for network operations
type NetworkCreatedEvent struct {
	NetworkID   string `json:"network_id"`
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Timestamp   string `json:"timestamp"`
}

type NetworkProvisionedEvent struct {
	NetworkID   string `json:"network_id"`
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp"`
}

type NodeStatusChangedEvent struct {
	NodeID      string `json:"node_id"`
	NetworkID   string `json:"network_id"`
	OldStatus   string `json:"old_status"`
	NewStatus   string `json:"new_status"`
	Timestamp   string `json:"timestamp"`
}

type TopologyUpdatedEvent struct {
	NetworkID   string `json:"network_id"`
	ChangeType  string `json:"change_type"` // node_added, link_added, config_changed
	Timestamp   string `json:"timestamp"`
}

