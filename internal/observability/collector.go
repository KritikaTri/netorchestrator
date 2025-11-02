package observability

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"netorchestrator/internal/models"
)

// MetricsCollector periodically collects and updates infrastructure metrics
type MetricsCollector struct {
	db      *gorm.DB
	metrics *Metrics
	logger  *zap.Logger
	ticker  *time.Ticker
	done    chan bool
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(db *gorm.DB, metrics *Metrics, logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		db:      db,
		metrics: metrics,
		logger:  logger,
		done:    make(chan bool),
	}
}

// Start begins periodic metric collection
func (c *MetricsCollector) Start(ctx context.Context, interval time.Duration) {
	c.ticker = time.NewTicker(interval)
	
	// Collect immediately on start
	c.collectMetrics(ctx)
	
	go func() {
		for {
			select {
			case <-c.ticker.C:
				c.collectMetrics(ctx)
			case <-c.done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
	
	c.logger.Info("Metrics collector started", zap.Duration("interval", interval))
}

// Stop stops the metrics collector
func (c *MetricsCollector) Stop() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
	close(c.done)
	c.logger.Info("Metrics collector stopped")
}

// collectMetrics queries database and updates all infrastructure metrics
func (c *MetricsCollector) collectMetrics(ctx context.Context) {
	// Collect network metrics
	networksByStatus := make(map[string]int)
	var networkResults []struct {
		Status string
		Count  int
	}
	if err := c.db.WithContext(ctx).
		Model(&models.Network{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&networkResults).Error; err != nil {
		c.logger.Error("Failed to collect network metrics", zap.Error(err))
	} else {
		for _, result := range networkResults {
			networksByStatus[string(result.Status)] = result.Count
		}
	}

	// Collect node metrics
	nodesByTypeAndStatus := make(map[string]map[string]int)
	var nodeResults []struct {
		Type   string
		Status string
		Count  int
	}
	if err := c.db.WithContext(ctx).
		Model(&models.Node{}).
		Select("node_type as type, status, count(*) as count").
		Group("node_type, status").
		Scan(&nodeResults).Error; err != nil {
		c.logger.Error("Failed to collect node metrics", zap.Error(err))
	} else {
		for _, result := range nodeResults {
			if nodesByTypeAndStatus[result.Type] == nil {
				nodesByTypeAndStatus[result.Type] = make(map[string]int)
			}
			nodesByTypeAndStatus[result.Type][string(result.Status)] = result.Count
		}
	}

	// Collect active container counts by querying nodes with status=active
	containersByType := make(map[string]int)
	var containerResults []struct {
		Type  string
		Count int
	}
	if err := c.db.WithContext(ctx).
		Model(&models.Node{}).
		Select("node_type as type, count(*) as count").
		Where("status = ?", models.NodeStatusActive).
		Group("node_type").
		Scan(&containerResults).Error; err != nil {
		c.logger.Error("Failed to collect container metrics", zap.Error(err))
	} else {
		for _, result := range containerResults {
			containersByType[result.Type] = result.Count
		}
	}

	// Collect link metrics
	linksByStatus := make(map[string]int)
	var linkResults []struct {
		Status string
		Count  int
	}
	if err := c.db.WithContext(ctx).
		Model(&models.Link{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&linkResults).Error; err != nil {
		c.logger.Error("Failed to collect link metrics", zap.Error(err))
	} else {
		for _, result := range linkResults {
			linksByStatus[string(result.Status)] = result.Count
		}
	}

	// Collect policy metrics
	policiesByTypeAndStatus := make(map[string]map[string]int)
	var policyResults []struct {
		Type   string
		Status string
		Count  int
	}
	if err := c.db.WithContext(ctx).
		Model(&models.Policy{}).
		Select("policy_type as type, status, count(*) as count").
		Group("policy_type, status").
		Scan(&policyResults).Error; err != nil {
		c.logger.Error("Failed to collect policy metrics", zap.Error(err))
	} else {
		for _, result := range policyResults {
			if policiesByTypeAndStatus[result.Type] == nil {
				policiesByTypeAndStatus[result.Type] = make(map[string]int)
			}
			policiesByTypeAndStatus[result.Type][string(result.Status)] = result.Count
		}
	}

	// Update all metrics
	c.metrics.UpdateInfrastructureMetrics(
		networksByStatus,
		nodesByTypeAndStatus,
		containersByType,
		linksByStatus,
		policiesByTypeAndStatus,
	)

	c.logger.Debug("Metrics collected successfully",
		zap.Int("networks", len(networksByStatus)),
		zap.Int("node_types", len(nodesByTypeAndStatus)),
		zap.Int("links", len(linksByStatus)),
		zap.Int("policies", len(policiesByTypeAndStatus)))
}
