package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"netorchestrator/internal/models"
	"netorchestrator/internal/observability"
)

// MetricsHandler handles metrics aggregation endpoints
type MetricsHandler struct {
	db      *gorm.DB
	metrics *observability.Metrics
	logger  *zap.Logger
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(db *gorm.DB, metrics *observability.Metrics, logger *zap.Logger) *MetricsHandler {
	return &MetricsHandler{
		db:      db,
		metrics: metrics,
		logger:  logger,
	}
}

// MetricsSummaryResponse represents the aggregated metrics response
type MetricsSummaryResponse struct {
	Timestamp      string                `json:"timestamp"`
	Health         HealthMetrics         `json:"health"`
	Infrastructure InfrastructureMetrics `json:"infrastructure"`
	Performance    PerformanceMetrics    `json:"performance"`
	Resources      ResourceMetrics       `json:"resources"`
	Uptime         string                `json:"uptime"`
}

type HealthMetrics struct {
	Score  float64 `json:"score"`
	Status string  `json:"status"`
}

type InfrastructureMetrics struct {
	Networks   CountMetrics     `json:"networks"`
	Nodes      NodeMetrics      `json:"nodes"`
	Containers ContainerMetrics `json:"containers"`
	Links      CountMetrics     `json:"links"`
	Policies   PolicyMetrics    `json:"policies"`
}

type CountMetrics struct {
	Total  int `json:"total"`
	Active int `json:"active"`
}

type NodeMetrics struct {
	Total   int            `json:"total"`
	Active  int            `json:"active"`
	Pending int            `json:"pending"`
	ByType  map[string]int `json:"by_type"`
}

type ContainerMetrics struct {
	Total  int            `json:"total"`
	ByType map[string]int `json:"by_type"`
}

type PolicyMetrics struct {
	Total   int            `json:"total"`
	Active  int            `json:"active"`
	Pending int            `json:"pending"`
	ByType  map[string]int `json:"by_type"`
}

type PerformanceMetrics struct {
	ResponseTimeAvg float64 `json:"response_time_avg_ms"`
	ResponseTimeP95 float64 `json:"response_time_p95_ms"`
	ResponseTimeP99 float64 `json:"response_time_p99_ms"`
	Throughput      float64 `json:"throughput_req_per_sec"`
	TotalRequests   float64 `json:"total_requests"`
	ErrorRate       float64 `json:"error_rate_percent"`
	SuccessRate     float64 `json:"success_rate_percent"`
}

type ResourceMetrics struct {
	CPU     CPUMetrics     `json:"cpu"`
	Memory  MemoryMetrics  `json:"memory"`
	Storage StorageMetrics `json:"storage"`
	Network NetworkMetrics `json:"network"`
}

type CPUMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	CoresUsed    float64 `json:"cores_used"`
	CoresTotal   int     `json:"cores_total"`
}

type MemoryMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	UsedMB       float64 `json:"used_mb"`
	TotalMB      float64 `json:"total_mb"`
}

type StorageMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	UsedGB       float64 `json:"used_gb"`
	TotalGB      float64 `json:"total_gb"`
}

type NetworkMetrics struct {
	BandwidthUsagePercent float64 `json:"bandwidth_usage_percent"`
	PacketsIn             int64   `json:"packets_in"`
	PacketsOut            int64   `json:"packets_out"`
	Errors                int     `json:"errors"`
	ActiveConnections     int     `json:"active_connections"`
}

// GetMetricByEntityAndName handles GET /api/v1/metrics endpoint with entity_id and metric_name query params
// @Summary Get specific metric by entity ID and metric name
// @Description Get detailed information about a specific metric from the database
// @Tags Metrics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param entity_id query string true "Entity ID (container ID, node ID, etc.)"
// @Param metric_name query string true "Metric name (cpu_percent, memory_usage_mb, etc.)"
// @Success 200 {object} map[string]interface{} "Metric details"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /metrics [get]
func (h *MetricsHandler) GetMetricByEntityAndName(c *gin.Context) {
	// Get query parameters
	entityID := c.Query("entity_id")
	metricName := c.Query("metric_name")

	// Validate required parameters
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "entity_id parameter is required",
		})
		return
	}

	if metricName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "metric_name parameter is required",
		})
		return
	}

	// Query the database for the specific metric
	var metric models.Metric
	err := h.db.Where("entity_id = ? AND metric_name = ?", entityID, metricName).
		Order("timestamp DESC").
		First(&metric).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Metric not found",
				"details": map[string]string{
					"entity_id":   entityID,
					"metric_name": metricName,
				},
			})
			return
		}
		h.logger.Error("Failed to query metric", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve metric",
		})
		return
	}

	// Get additional historical data (last 10 records)
	var historicalMetrics []models.Metric
	h.db.Where("entity_id = ? AND metric_name = ?", entityID, metricName).
		Order("timestamp DESC").
		Limit(10).
		Find(&historicalMetrics)

	// Calculate some statistics
	var values []float64
	for _, m := range historicalMetrics {
		values = append(values, m.MetricValue)
	}

	var avgValue, minValue, maxValue float64
	if len(values) > 0 {
		sum := 0.0
		minValue = values[0]
		maxValue = values[0]

		for _, v := range values {
			sum += v
			if v < minValue {
				minValue = v
			}
			if v > maxValue {
				maxValue = v
			}
		}
		avgValue = sum / float64(len(values))
	}

	// Build response
	response := gin.H{
		"entity_id":     metric.EntityID,
		"entity_type":   metric.EntityType,
		"metric_name":   metric.MetricName,
		"current_value": metric.MetricValue,
		"unit":          metric.Unit,
		"timestamp":     metric.Timestamp,
		"metadata":      metric.Metadata,
		"statistics": gin.H{
			"count":          len(historicalMetrics),
			"average":        avgValue,
			"minimum":        minValue,
			"maximum":        maxValue,
			"last_10_values": values,
		},
		"historical_data": historicalMetrics,
		"status":          "success",
	}

	c.JSON(http.StatusOK, response)
}

// GetMetricsSummary aggregates and returns system metrics
func (h *MetricsHandler) GetMetricsSummary(c *gin.Context) {
	ctx := context.Background()

	summary := MetricsSummaryResponse{
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Get infrastructure metrics from database
	infrastructureMetrics, err := h.getInfrastructureMetrics(ctx)
	if err != nil {
		h.logger.Error("Failed to get infrastructure metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch metrics"})
		return
	}
	summary.Infrastructure = infrastructureMetrics

	// Calculate health score based on infrastructure
	summary.Health = h.calculateHealthScore(infrastructureMetrics)

	// Get performance metrics from Prometheus
	performanceMetrics, err := h.getPerformanceMetrics()
	if err != nil {
		h.logger.Warn("Failed to get performance metrics, using defaults", zap.Error(err))
		performanceMetrics = PerformanceMetrics{
			ResponseTimeAvg: 0,
			ResponseTimeP95: 0,
			ResponseTimeP99: 0,
			Throughput:      0,
			TotalRequests:   0,
			ErrorRate:       0,
			SuccessRate:     100,
		}
	}
	summary.Performance = performanceMetrics

	// Get resource metrics (simulated for now, can be enhanced with actual system metrics)
	summary.Resources = h.getResourceMetrics(infrastructureMetrics)

	// Get uptime
	summary.Uptime = h.getUptime()

	c.JSON(http.StatusOK, summary)
}

// getInfrastructureMetrics queries the database for infrastructure counts
func (h *MetricsHandler) getInfrastructureMetrics(ctx context.Context) (InfrastructureMetrics, error) {
	var metrics InfrastructureMetrics

	// Networks
	var totalNetworks, activeNetworks int64
	if err := h.db.WithContext(ctx).Model(&models.Network{}).Count(&totalNetworks).Error; err != nil {
		return metrics, err
	}
	if err := h.db.WithContext(ctx).Model(&models.Network{}).Where("status = ?", "active").Count(&activeNetworks).Error; err != nil {
		return metrics, err
	}
	metrics.Networks = CountMetrics{
		Total:  int(totalNetworks),
		Active: int(activeNetworks),
	}

	// Nodes
	var totalNodes, activeNodes, pendingNodes int64
	if err := h.db.WithContext(ctx).Model(&models.Node{}).Count(&totalNodes).Error; err != nil {
		return metrics, err
	}
	if err := h.db.WithContext(ctx).Model(&models.Node{}).Where("status = ?", "active").Count(&activeNodes).Error; err != nil {
		return metrics, err
	}
	if err := h.db.WithContext(ctx).Model(&models.Node{}).Where("status = ?", "pending").Count(&pendingNodes).Error; err != nil {
		return metrics, err
	}

	// Nodes by type
	type NodeTypeCount struct {
		Type  string
		Count int64
	}
	var nodeTypeCounts []NodeTypeCount
	if err := h.db.WithContext(ctx).Model(&models.Node{}).
		Select("node_type as type, count(*) as count").
		Where("status = ?", "active").
		Group("node_type").
		Scan(&nodeTypeCounts).Error; err != nil {
		return metrics, err
	}

	nodesByType := make(map[string]int)
	for _, ntc := range nodeTypeCounts {
		nodesByType[ntc.Type] = int(ntc.Count)
	}

	metrics.Nodes = NodeMetrics{
		Total:   int(totalNodes),
		Active:  int(activeNodes),
		Pending: int(pendingNodes),
		ByType:  nodesByType,
	}

	// Containers (from nodes that are active)
	containersByType := make(map[string]int)
	for nodeType, count := range nodesByType {
		containersByType[nodeType] = count
	}
	metrics.Containers = ContainerMetrics{
		Total:  int(activeNodes),
		ByType: containersByType,
	}

	// Links
	var totalLinks, activeLinks int64
	if err := h.db.WithContext(ctx).Model(&models.Link{}).Count(&totalLinks).Error; err != nil {
		return metrics, err
	}
	if err := h.db.WithContext(ctx).Model(&models.Link{}).Where("status = ?", "active").Count(&activeLinks).Error; err != nil {
		return metrics, err
	}
	metrics.Links = CountMetrics{
		Total:  int(totalLinks),
		Active: int(activeLinks),
	}

	// Policies
	var totalPolicies, activePolicies, pendingPolicies int64
	if err := h.db.WithContext(ctx).Model(&models.Policy{}).Count(&totalPolicies).Error; err != nil {
		return metrics, err
	}
	if err := h.db.WithContext(ctx).Model(&models.Policy{}).Where("status = ?", "active").Count(&activePolicies).Error; err != nil {
		return metrics, err
	}
	if err := h.db.WithContext(ctx).Model(&models.Policy{}).Where("status = ?", "pending").Count(&pendingPolicies).Error; err != nil {
		return metrics, err
	}

	// Policies by type
	type PolicyTypeCount struct {
		Type  string
		Count int64
	}
	var policyTypeCounts []PolicyTypeCount
	if err := h.db.WithContext(ctx).Model(&models.Policy{}).
		Select("policy_type as type, count(*) as count").
		Group("policy_type").
		Scan(&policyTypeCounts).Error; err != nil {
		return metrics, err
	}

	policiesByType := make(map[string]int)
	for _, ptc := range policyTypeCounts {
		policiesByType[ptc.Type] = int(ptc.Count)
	}

	metrics.Policies = PolicyMetrics{
		Total:   int(totalPolicies),
		Active:  int(activePolicies),
		Pending: int(pendingPolicies),
		ByType:  policiesByType,
	}

	return metrics, nil
}

// calculateHealthScore calculates overall system health based on infrastructure
func (h *MetricsHandler) calculateHealthScore(infra InfrastructureMetrics) HealthMetrics {
	score := 100.0
	status := "healthy"

	// Deduct points for inactive resources
	if infra.Networks.Total > 0 {
		networkHealth := float64(infra.Networks.Active) / float64(infra.Networks.Total)
		if networkHealth < 0.9 {
			score -= (1.0 - networkHealth) * 20
		}
	}

	if infra.Nodes.Total > 0 {
		nodeHealth := float64(infra.Nodes.Active) / float64(infra.Nodes.Total)
		if nodeHealth < 0.9 {
			score -= (1.0 - nodeHealth) * 30
		}
	}

	if infra.Links.Total > 0 {
		linkHealth := float64(infra.Links.Active) / float64(infra.Links.Total)
		if linkHealth < 0.9 {
			score -= (1.0 - linkHealth) * 15
		}
	}

	// Determine status
	if score >= 90 {
		status = "healthy"
	} else if score >= 70 {
		status = "degraded"
	} else {
		status = "unhealthy"
	}

	return HealthMetrics{
		Score:  score,
		Status: status,
	}
}

// getPerformanceMetrics extracts performance data from Prometheus metrics
func (h *MetricsHandler) getPerformanceMetrics() (PerformanceMetrics, error) {
	var metrics PerformanceMetrics

	// Get HTTP request duration histogram
	histogramMetric := h.metrics.HTTPRequestDuration
	metricChan := make(chan prometheus.Metric, 100)
	go func() {
		histogramMetric.Collect(metricChan)
		close(metricChan)
	}()

	var totalCount uint64
	var totalSum float64

	for metric := range metricChan {
		var m dto.Metric
		if err := metric.Write(&m); err != nil {
			continue
		}

		if m.Histogram != nil {
			totalCount += m.Histogram.GetSampleCount()
			totalSum += m.Histogram.GetSampleSum()
		}
	}

	// Calculate average response time (convert to milliseconds)
	if totalCount > 0 {
		metrics.ResponseTimeAvg = (totalSum / float64(totalCount)) * 1000
		metrics.TotalRequests = float64(totalCount)
	}

	// Estimate P95 and P99 (simplified calculation)
	// In production, you'd query Prometheus with quantile queries
	metrics.ResponseTimeP95 = metrics.ResponseTimeAvg * 2.5
	metrics.ResponseTimeP99 = metrics.ResponseTimeAvg * 4.0

	// Get HTTP request count
	counterMetric := h.metrics.HTTPRequestsTotal
	counterChan := make(chan prometheus.Metric, 100)
	go func() {
		counterMetric.Collect(counterChan)
		close(counterChan)
	}()

	var totalRequests float64
	var errorRequests float64

	for metric := range counterChan {
		var m dto.Metric
		if err := metric.Write(&m); err != nil {
			continue
		}

		if m.Counter != nil {
			value := m.Counter.GetValue()
			totalRequests += value

			// Check if this is an error status code
			for _, label := range m.Label {
				if label.GetName() == "status_code" && (label.GetValue()[0] == '4' || label.GetValue()[0] == '5') {
					errorRequests += value
				}
			}
		}
	}

	// Calculate rates
	if totalRequests > 0 {
		metrics.ErrorRate = (errorRequests / totalRequests) * 100
		metrics.SuccessRate = 100 - metrics.ErrorRate

		// Estimate throughput (requests per second)
		// This is a simple estimate; in production, calculate over a time window
		metrics.Throughput = totalRequests / 60 // Assume 1 minute window
	} else {
		metrics.SuccessRate = 100
		metrics.ErrorRate = 0
		metrics.Throughput = 0
	}

	return metrics, nil
}

// getResourceMetrics provides resource utilization estimates
func (h *MetricsHandler) getResourceMetrics(infra InfrastructureMetrics) ResourceMetrics {
	// These are estimates based on infrastructure load
	// In production, you'd integrate with actual system monitoring

	activeContainers := float64(infra.Containers.Total)
	activeNetworks := float64(infra.Networks.Active)

	// Estimate CPU usage based on active resources
	cpuUsage := (activeContainers * 2.5) + (activeNetworks * 1.5)
	if cpuUsage > 100 {
		cpuUsage = 95 // Cap at 95%
	}

	// Estimate memory usage
	memoryUsageMB := (activeContainers * 256) + (activeNetworks * 128)
	totalMemoryMB := 16384.0 // 16 GB
	memoryUsagePercent := (memoryUsageMB / totalMemoryMB) * 100

	// Storage usage
	storageUsedGB := 50.0 + (activeNetworks * 2) + (activeContainers * 0.5)
	totalStorageGB := 200.0
	storageUsagePercent := (storageUsedGB / totalStorageGB) * 100

	// Network metrics
	bandwidthUsage := (activeContainers * 1.5) + (activeNetworks * 2)
	if bandwidthUsage > 100 {
		bandwidthUsage = 90
	}

	return ResourceMetrics{
		CPU: CPUMetrics{
			UsagePercent: cpuUsage,
			CoresUsed:    cpuUsage / 100 * 8,
			CoresTotal:   8,
		},
		Memory: MemoryMetrics{
			UsagePercent: memoryUsagePercent,
			UsedMB:       memoryUsageMB,
			TotalMB:      totalMemoryMB,
		},
		Storage: StorageMetrics{
			UsagePercent: storageUsagePercent,
			UsedGB:       storageUsedGB,
			TotalGB:      totalStorageGB,
		},
		Network: NetworkMetrics{
			BandwidthUsagePercent: bandwidthUsage,
			PacketsIn:             int64(5000 + (activeContainers * 100)),
			PacketsOut:            int64(4000 + (activeContainers * 80)),
			Errors:                int(activeContainers / 50), // Very low error rate
			ActiveConnections:     infra.Nodes.Active,
		},
	}
}

// getUptime returns the system uptime (placeholder)
func (h *MetricsHandler) getUptime() string {
	// In production, you'd track actual start time
	// For now, return a placeholder
	return "24h 15m"
}
