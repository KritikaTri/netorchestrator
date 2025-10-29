package services

import (
	"time"

	"go.uber.org/zap"
	"netorchestrator/internal/websocket"
)

// EventService handles broadcasting events to WebSocket clients
type EventService struct {
	wsHub  *websocket.Hub
	logger *zap.Logger
}

// NewEventService creates a new event service
func NewEventService(wsHub *websocket.Hub, logger *zap.Logger) *EventService {
	return &EventService{
		wsHub:  wsHub,
		logger: logger,
	}
}

// BroadcastNetworkUpdate broadcasts network status updates
func (s *EventService) BroadcastNetworkUpdate(networkID string, status string, data interface{}) {
	event := map[string]interface{}{
		"network_id": networkID,
		"status":     status,
		"data":       data,
		"timestamp":  time.Now().Unix(),
	}

	s.wsHub.BroadcastEvent("network_update", event)
	s.logger.Info("Broadcasted network update", 
		zap.String("network_id", networkID), 
		zap.String("status", status))
}

// BroadcastNodeStatusUpdate broadcasts node status changes
func (s *EventService) BroadcastNodeStatusUpdate(nodeID string, networkID string, status string, data interface{}) {
	event := map[string]interface{}{
		"node_id":    nodeID,
		"network_id": networkID,
		"status":     status,
		"data":       data,
		"timestamp":  time.Now().Unix(),
	}

	s.wsHub.BroadcastEvent("node_status", event)
	s.logger.Info("Broadcasted node status update", 
		zap.String("node_id", nodeID), 
		zap.String("network_id", networkID), 
		zap.String("status", status))
}

// BroadcastMetricUpdate broadcasts real-time metrics
func (s *EventService) BroadcastMetricUpdate(sourceType string, sourceID string, metricName string, value float64) {
	event := map[string]interface{}{
		"source_type":  sourceType,
		"source_id":    sourceID,
		"metric_name":  metricName,
		"value":        value,
		"timestamp":    time.Now().Unix(),
	}

	s.wsHub.BroadcastEvent("metric_update", event)
	s.logger.Debug("Broadcasted metric update", 
		zap.String("source_type", sourceType), 
		zap.String("source_id", sourceID), 
		zap.String("metric_name", metricName), 
		zap.Float64("value", value))
}

// BroadcastAlert broadcasts system alerts
func (s *EventService) BroadcastAlert(severity string, message string, source string, data interface{}) {
	event := map[string]interface{}{
		"severity":  severity,
		"message":   message,
		"source":    source,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	s.wsHub.BroadcastEvent("alert", event)
	s.logger.Warn("Broadcasted alert", 
		zap.String("severity", severity), 
		zap.String("message", message), 
		zap.String("source", source))
}

// BroadcastTopologyUpdate broadcasts network topology changes
func (s *EventService) BroadcastTopologyUpdate(networkID string, changeType string, data interface{}) {
	event := map[string]interface{}{
		"network_id":  networkID,
		"change_type": changeType,
		"data":        data,
		"timestamp":   time.Now().Unix(),
	}

	s.wsHub.BroadcastEvent("topology_update", event)
	s.logger.Info("Broadcasted topology update", 
		zap.String("network_id", networkID), 
		zap.String("change_type", changeType))
}

// StartPeriodicUpdates starts sending periodic system updates
func (s *EventService) StartPeriodicUpdates() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			// Send periodic system status
			systemStatus := map[string]interface{}{
				"uptime":           time.Now().Unix(),
				"connected_clients": s.getConnectedClientCount(),
				"status":           "operational",
			}

			s.wsHub.BroadcastEvent("system_status", systemStatus)
		}
	}()
}

// getConnectedClientCount gets the number of connected WebSocket clients
func (s *EventService) getConnectedClientCount() int {
	// This would need to be implemented in the websocket.Hub
	// For now, return a placeholder
	return 0
}