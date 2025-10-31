package events

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// EventSubscriber handles subscribing to events from NATS
type EventSubscriber struct {
	conn   *nats.Conn
	logger *zap.Logger
}

// NewEventSubscriber creates a new event subscriber
func NewEventSubscriber(natsURL string, logger *zap.Logger) (*EventSubscriber, error) {
	conn, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &EventSubscriber{
		conn:   conn,
		logger: logger,
	}, nil
}

// Close closes the NATS connection
func (es *EventSubscriber) Close() {
	if es.conn != nil {
		es.conn.Close()
	}
}

// Subscribe subscribes to an event subject and calls handler when events arrive
func (es *EventSubscriber) Subscribe(subject string, handler func([]byte) error) error {
	if es.conn == nil {
		return fmt.Errorf("NATS not connected")
	}

	_, err := es.conn.Subscribe(subject, func(msg *nats.Msg) {
		es.logger.Info("Received event", zap.String("subject", subject))
		if err := handler(msg.Data); err != nil {
			es.logger.Error("Error handling event", zap.Error(err), zap.String("subject", subject))
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}

	es.logger.Info("Subscribed to event", zap.String("subject", subject))
	return nil
}

// UnmarshalEvent unmarshals event data into the provided type
func UnmarshalEvent(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

