package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"netorchestrator/internal/models"
)

// MonitoringService handles monitoring-related operations
type MonitoringService struct {
	db     *gorm.DB
	cache  *redis.Client
	logger *zap.Logger
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(db *gorm.DB, cache *redis.Client, logger *zap.Logger) *MonitoringService {
	return &MonitoringService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// GetNetworkMetrics retrieves metrics for a network
func (s *MonitoringService) GetNetworkMetrics(ctx context.Context, networkID uuid.UUID) (map[string]interface{}, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("network_metrics:%s", networkID.String())
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		s.logger.Debug("Retrieved network metrics from cache", zap.String("network_id", networkID.String()))
		// In a real implementation, you would unmarshal the cached JSON
		return map[string]interface{}{
			"cached": true,
			"data":   cached,
		}, nil
	}

	// Get network and related data
	var network models.Network
	if err := s.db.WithContext(ctx).Preload("Nodes").Preload("Links").First(&network, "id = ?", networkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("network not found")
		}
		s.logger.Error("Failed to get network for metrics", zap.Error(err))
		return nil, fmt.Errorf("failed to get network: %w", err)
	}

	// Calculate basic metrics
	metrics := map[string]interface{}{
		"network_id":    networkID.String(),
		"network_name":  network.Name,
		"status":        network.Status,
		"node_count":    len(network.Nodes),
		"link_count":    len(network.Links),
		"created_at":    network.CreatedAt,
		"updated_at":    network.UpdatedAt,
		"timestamp":     time.Now(),
	}

	// Add node status breakdown
	nodeStatusCount := make(map[string]int)
	for _, node := range network.Nodes {
		nodeStatusCount[string(node.Status)]++
	}
	metrics["node_status_breakdown"] = nodeStatusCount

	// Add link status breakdown
	linkStatusCount := make(map[string]int)
	for _, link := range network.Links {
		linkStatusCount[string(link.Status)]++
	}
	metrics["link_status_breakdown"] = linkStatusCount

	// Cache the results for 5 minutes
	s.cache.Set(ctx, cacheKey, fmt.Sprintf("%+v", metrics), 5*time.Minute)

	return metrics, nil
}

// GetNetworkHealth checks the health of a network
func (s *MonitoringService) GetNetworkHealth(ctx context.Context, networkID uuid.UUID) (map[string]interface{}, error) {
	// Get network
	var network models.Network
	if err := s.db.WithContext(ctx).Preload("Nodes").Preload("Links").First(&network, "id = ?", networkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("network not found")
		}
		s.logger.Error("Failed to get network for health check", zap.Error(err))
		return nil, fmt.Errorf("failed to get network: %w", err)
	}

	// Calculate health metrics
	health := map[string]interface{}{
		"network_id": networkID.String(),
		"status":     network.Status,
		"timestamp":  time.Now(),
	}

	// Check overall network status
	if network.Status == models.NetworkStatusActive {
		health["overall_health"] = "healthy"
	} else if network.Status == models.NetworkStatusError {
		health["overall_health"] = "unhealthy"
	} else {
		health["overall_health"] = "unknown"
	}

	// Check node health
	healthyNodes := 0
	totalNodes := len(network.Nodes)
	for _, node := range network.Nodes {
		if node.Status == models.NodeStatusActive {
			healthyNodes++
		}
	}

	health["node_health"] = map[string]interface{}{
		"healthy_nodes": healthyNodes,
		"total_nodes":   totalNodes,
		"health_ratio":  float64(healthyNodes) / float64(totalNodes),
	}

	// Check link health
	healthyLinks := 0
	totalLinks := len(network.Links)
	for _, link := range network.Links {
		if link.Status == models.LinkStatusActive {
			healthyLinks++
		}
	}

	health["link_health"] = map[string]interface{}{
		"healthy_links": healthyLinks,
		"total_links":   totalLinks,
		"health_ratio":  float64(healthyLinks) / float64(totalLinks),
	}

	return health, nil
}

// GetNodeMetrics retrieves metrics for a specific node
func (s *MonitoringService) GetNodeMetrics(ctx context.Context, nodeID uuid.UUID) (map[string]interface{}, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("node_metrics:%s", nodeID.String())
	cached, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		s.logger.Debug("Retrieved node metrics from cache", zap.String("node_id", nodeID.String()))
		return map[string]interface{}{
			"cached": true,
			"data":   cached,
		}, nil
	}

	// Get node
	var node models.Node
	if err := s.db.WithContext(ctx).Preload("Network").Preload("Links").First(&node, "id = ?", nodeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("node not found")
		}
		s.logger.Error("Failed to get node for metrics", zap.Error(err))
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Calculate node metrics
	metrics := map[string]interface{}{
		"node_id":      nodeID.String(),
		"node_name":    node.Name,
		"node_type":    node.Type,
		"status":       node.Status,
		"ip_address":   node.IPAddress,
		"mac_address":  node.MACAddress,
		"network_id":   node.NetworkID.String(),
		"created_at":   node.CreatedAt,
		"updated_at":   node.UpdatedAt,
		"timestamp":    time.Now(),
		"link_count":   len(node.Links),
	}

	// Add configuration metrics
	if node.Config.CPU > 0 {
		metrics["cpu"] = node.Config.CPU
	}
	if node.Config.Memory > 0 {
		metrics["memory_mb"] = node.Config.Memory
	}
	if node.Config.Storage > 0 {
		metrics["storage_gb"] = node.Config.Storage
	}

	// Cache the results for 2 minutes
	s.cache.Set(ctx, cacheKey, fmt.Sprintf("%+v", metrics), 2*time.Minute)

	return metrics, nil
}

// GetNodeHealth checks the health of a specific node
func (s *MonitoringService) GetNodeHealth(ctx context.Context, nodeID uuid.UUID) (map[string]interface{}, error) {
	// Get node
	var node models.Node
	if err := s.db.WithContext(ctx).Preload("Network").First(&node, "id = ?", nodeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("node not found")
		}
		s.logger.Error("Failed to get node for health check", zap.Error(err))
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Calculate health
	health := map[string]interface{}{
		"node_id":   nodeID.String(),
		"status":    node.Status,
		"timestamp": time.Now(),
	}

	// Determine health based on status
	if node.Status == models.NodeStatusActive {
		health["health"] = "healthy"
	} else if node.Status == models.NodeStatusError {
		health["health"] = "unhealthy"
	} else {
		health["health"] = "unknown"
	}

	return health, nil
}

// ListAlerts retrieves all active alerts
func (s *MonitoringService) ListAlerts(ctx context.Context) ([]map[string]interface{}, error) {
	// In a real implementation, this would query an alerts table
	// For now, return empty list
	return []map[string]interface{}{}, nil
}

// AcknowledgeAlert acknowledges an alert
func (s *MonitoringService) AcknowledgeAlert(ctx context.Context, alertID uuid.UUID) error {
	// In a real implementation, this would update an alerts table
	s.logger.Info("Alert acknowledged", zap.String("alert_id", alertID.String()))
	return nil
}
