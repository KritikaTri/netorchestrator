package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"netorchestrator/internal/models"
)

// NetworkMetricsHandler handles network-specific metrics queries
type NetworkMetricsHandler struct {
	db            *gorm.DB
	logger        *zap.Logger
	prometheusURL string
}

// NewNetworkMetricsHandler creates a new network metrics handler
func NewNetworkMetricsHandler(db *gorm.DB, logger *zap.Logger, prometheusURL string) *NetworkMetricsHandler {
	if prometheusURL == "" {
		prometheusURL = "http://localhost:9091" // Default Prometheus URL
	}
	return &NetworkMetricsHandler{
		db:            db,
		logger:        logger,
		prometheusURL: prometheusURL,
	}
}

// NetworkMetricsResponse represents the JSON response for network-specific metrics
type NetworkMetricsResponse struct {
	NetworkID    string                  `json:"network_id"`
	NetworkName  string                  `json:"network_name"`
	Timestamp    string                  `json:"timestamp"`
	Containers   NetworkContainerMetrics `json:"containers"`
	Links        NetworkLinkMetrics      `json:"links"`
	Policies     NetworkPolicyMetrics    `json:"policies"`
	Provisioning NetworkProvisionMetrics `json:"provisioning"`
}

type NetworkContainerMetrics struct {
	TotalActive int            `json:"total_active"`
	ByNodeType  map[string]int `json:"by_node_type"`
}

type NetworkLinkMetrics struct {
	TotalLinks  int          `json:"total_links"`
	ActiveLinks int          `json:"active_links"`
	DownLinks   int          `json:"down_links"`
	LinkDetails []LinkDetail `json:"link_details"`
}

type LinkDetail struct {
	LinkID string  `json:"link_id"`
	NodeA  string  `json:"node_a"`
	NodeB  string  `json:"node_b"`
	Status float64 `json:"status"` // 1=up, 0=down
}

type NetworkPolicyMetrics struct {
	TotalFailures  int            `json:"total_failures"`
	FailuresByType map[string]int `json:"failures_by_type"`
}

type NetworkProvisionMetrics struct {
	AvgLatencySeconds float64            `json:"avg_latency_seconds"`
	LatencyByType     map[string]float64 `json:"latency_by_type"`
}

// PrometheusQueryResult represents Prometheus API query response
type PrometheusQueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// GetNetworkMetrics handles GET /api/v1/metrics/network/:network_id
func (h *NetworkMetricsHandler) GetNetworkMetrics(c *gin.Context) {
	networkID := c.Param("network_id")

	h.logger.Info("Fetching network-specific metrics",
		zap.String("network_id", networkID))

	// Verify network exists
	var network models.Network
	if err := h.db.First(&network, "id = ?", networkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Network not found"})
			return
		}
		h.logger.Error("Failed to query network", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Query Prometheus for network-specific metrics
	containerMetrics, err := h.queryNetworkContainers(networkID)
	if err != nil {
		h.logger.Warn("Failed to query container metrics", zap.Error(err))
	}

	linkMetrics, err := h.queryNetworkLinks(networkID)
	if err != nil {
		h.logger.Warn("Failed to query link metrics", zap.Error(err))
	}

	policyMetrics, err := h.queryNetworkPolicies(networkID)
	if err != nil {
		h.logger.Warn("Failed to query policy metrics", zap.Error(err))
	}

	provisionMetrics, err := h.queryNetworkProvisioning(networkID)
	if err != nil {
		h.logger.Warn("Failed to query provisioning metrics", zap.Error(err))
	}

	// Build response
	response := NetworkMetricsResponse{
		NetworkID:    network.ID.String(),
		NetworkName:  network.Name,
		Timestamp:    fmt.Sprintf("%d", time.Now().Unix()),
		Containers:   containerMetrics,
		Links:        linkMetrics,
		Policies:     policyMetrics,
		Provisioning: provisionMetrics,
	}

	c.JSON(http.StatusOK, response)
}

// queryNetworkContainers queries container count metrics from Prometheus
func (h *NetworkMetricsHandler) queryNetworkContainers(networkID string) (NetworkContainerMetrics, error) {
	// Always initialize with empty maps/slices to avoid null in JSON
	metrics := NetworkContainerMetrics{
		TotalActive: 0,
		ByNodeType:  make(map[string]int),
	}

	query := fmt.Sprintf(`netorch_network_containers_active{network_id="%s"}`, networkID)
	result, err := h.queryPrometheus(query)
	if err != nil {
		h.logger.Debug("No container metrics found, returning defaults", zap.Error(err))
		return metrics, nil // Return initialized struct instead of error
	}

	// Process results if available
	for _, r := range result.Data.Result {
		nodeType := r.Metric["node_type"]
		if len(r.Value) >= 2 {
			if value, ok := r.Value[1].(string); ok {
				var count int
				fmt.Sscanf(value, "%d", &count)
				metrics.ByNodeType[nodeType] = count
				metrics.TotalActive += count
			}
		}
	}

	return metrics, nil
}

// queryNetworkLinks queries link status metrics from Prometheus
func (h *NetworkMetricsHandler) queryNetworkLinks(networkID string) (NetworkLinkMetrics, error) {
	// Always initialize with empty slices to avoid null in JSON
	metrics := NetworkLinkMetrics{
		TotalLinks:  0,
		ActiveLinks: 0,
		DownLinks:   0,
		LinkDetails: []LinkDetail{},
	}

	query := fmt.Sprintf(`netorch_network_link_status{network_id="%s"}`, networkID)
	result, err := h.queryPrometheus(query)
	if err != nil {
		h.logger.Debug("No link metrics found, returning defaults", zap.Error(err))
		return metrics, nil // Return initialized struct instead of error
	}

	for _, r := range result.Data.Result {
		linkID := r.Metric["link_id"]
		nodeA := r.Metric["node_a"]
		nodeB := r.Metric["node_b"]

		var status float64
		if len(r.Value) >= 2 {
			if value, ok := r.Value[1].(string); ok {
				fmt.Sscanf(value, "%f", &status)
			}
		}

		metrics.LinkDetails = append(metrics.LinkDetails, LinkDetail{
			LinkID: linkID,
			NodeA:  nodeA,
			NodeB:  nodeB,
			Status: status,
		})

		metrics.TotalLinks++
		if status == 1.0 {
			metrics.ActiveLinks++
		} else {
			metrics.DownLinks++
		}
	}

	return metrics, nil
}

// queryNetworkPolicies queries policy failure metrics from Prometheus
func (h *NetworkMetricsHandler) queryNetworkPolicies(networkID string) (NetworkPolicyMetrics, error) {
	// Always initialize with empty maps to avoid null in JSON
	metrics := NetworkPolicyMetrics{
		TotalFailures:  0,
		FailuresByType: make(map[string]int),
	}

	query := fmt.Sprintf(`netorch_network_policy_failures_total{network_id="%s"}`, networkID)
	result, err := h.queryPrometheus(query)
	if err != nil {
		h.logger.Debug("No policy metrics found, returning defaults", zap.Error(err))
		return metrics, nil // Return initialized struct instead of error
	}

	for _, r := range result.Data.Result {
		policyType := r.Metric["policy_type"]
		if len(r.Value) >= 2 {
			if value, ok := r.Value[1].(string); ok {
				var count int
				fmt.Sscanf(value, "%d", &count)
				metrics.FailuresByType[policyType] = count
				metrics.TotalFailures += count
			}
		}
	}

	return metrics, nil
}

// queryNetworkProvisioning queries provisioning latency metrics from Prometheus
func (h *NetworkMetricsHandler) queryNetworkProvisioning(networkID string) (NetworkProvisionMetrics, error) {
	// Always initialize with empty maps to avoid null in JSON
	metrics := NetworkProvisionMetrics{
		AvgLatencySeconds: 0.0,
		LatencyByType:     make(map[string]float64),
	}

	// Query average latency using rate and histogram
	query := fmt.Sprintf(
		`avg(rate(netorch_network_provision_latency_seconds_sum{network_id="%s"}[5m])) / avg(rate(netorch_network_provision_latency_seconds_count{network_id="%s"}[5m]))`,
		networkID, networkID,
	)
	result, err := h.queryPrometheus(query)
	if err != nil {
		h.logger.Debug("No provisioning metrics found, returning defaults", zap.Error(err))
		return metrics, nil // Return initialized struct instead of error
	}

	if len(result.Data.Result) > 0 && len(result.Data.Result[0].Value) >= 2 {
		if value, ok := result.Data.Result[0].Value[1].(string); ok {
			fmt.Sscanf(value, "%f", &metrics.AvgLatencySeconds)
		}
	}

	// Query by node type
	queryByType := fmt.Sprintf(`netorch_network_provision_latency_seconds_sum{network_id="%s"}`, networkID)
	resultByType, err := h.queryPrometheus(queryByType)
	if err == nil {
		for _, r := range resultByType.Data.Result {
			nodeType := r.Metric["node_type"]
			if len(r.Value) >= 2 {
				if value, ok := r.Value[1].(string); ok {
					var latency float64
					fmt.Sscanf(value, "%f", &latency)
					metrics.LatencyByType[nodeType] = latency
				}
			}
		}
	}

	return metrics, nil
}

// queryPrometheus executes a PromQL query against Prometheus API
func (h *NetworkMetricsHandler) queryPrometheus(query string) (*PrometheusQueryResult, error) {
	// Build Prometheus query URL
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s", h.prometheusURL, url.QueryEscape(query))

	h.logger.Debug("Querying Prometheus",
		zap.String("url", queryURL),
		zap.String("query", query))

	// Execute HTTP request
	resp, err := http.Get(queryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query Prometheus: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Prometheus returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result PrometheusQueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Prometheus response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("Prometheus query failed with status: %s", result.Status)
	}

	return &result, nil
}
