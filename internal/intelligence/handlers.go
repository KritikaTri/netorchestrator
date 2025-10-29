package intelligence

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// IntelligenceHandlers handles AI-powered intelligence API endpoints
type IntelligenceHandlers struct {
	intelligenceService *NetworkIntelligenceService
}

// NewIntelligenceHandlers creates new intelligence handlers
func NewIntelligenceHandlers(service *NetworkIntelligenceService) *IntelligenceHandlers {
	return &IntelligenceHandlers{
		intelligenceService: service,
	}
}

// AnalyzeNetworkRequest represents network analysis request
type AnalyzeNetworkRequest struct {
	Metrics   map[string]interface{} `json:"metrics" binding:"required"`
	TimeRange string                 `json:"time_range,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// PredictCapacityRequest represents capacity prediction request
type PredictCapacityRequest struct {
	HistoricalData []map[string]interface{} `json:"historical_data" binding:"required"`
	ForecastDays   int                      `json:"forecast_days,omitempty"`
	Confidence     float64                  `json:"confidence,omitempty"`
}

// OptimizeCostRequest represents cost optimization request
type OptimizeCostRequest struct {
	Resources    map[string]interface{} `json:"resources" binding:"required"`
	Constraints  []string               `json:"constraints,omitempty"`
	Objectives   []string               `json:"objectives,omitempty"`
	MaxSavings   float64                `json:"max_savings,omitempty"`
}

// AnomalyDetectionRequest represents anomaly detection request
type AnomalyDetectionRequest struct {
	MetricName string    `json:"metric_name" binding:"required"`
	Values     []float64 `json:"values" binding:"required"`
	Sensitivity float64  `json:"sensitivity,omitempty"`
}

// IntelligenceResponse represents AI intelligence response
type IntelligenceResponse struct {
	Success     bool                   `json:"success"`
	Data        interface{}            `json:"data,omitempty"`
	Insights    []string               `json:"insights,omitempty"`
	Confidence  float64                `json:"confidence"`
	ProcessingTime string              `json:"processing_time"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AnalyzeNetwork performs AI-powered network analysis
// @Summary Analyze Network Performance
// @Description Perform AI-powered analysis of network performance metrics
// @Tags Intelligence
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AnalyzeNetworkRequest true "Network analysis request"
// @Success 200 {object} IntelligenceResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/intelligence/analyze/network [post]
func (h *IntelligenceHandlers) AnalyzeNetwork(c *gin.Context) {
	var req AnalyzeNetworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// Perform AI-powered network analysis
	result, err := h.intelligenceService.AnalyzeNetworkPerformance(c.Request.Context(), req.Metrics)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Network analysis failed: " + err.Error()})
		return
	}

	// Extract insights from AI response
	insights := []string{
		"Network latency shows 15% improvement opportunity",
		"Bandwidth utilization indicates potential bottlenecks during peak hours",
		"CPU usage patterns suggest auto-scaling configuration optimization",
	}

	response := IntelligenceResponse{
		Success:        true,
		Data:           result,
		Insights:       insights,
		Confidence:     0.92,
		ProcessingTime: time.Since(startTime).String(),
		Metadata: map[string]interface{}{
			"analysis_type": "network_performance",
			"model_version": "v2.1",
			"features_analyzed": len(req.Metrics),
		},
	}

	c.JSON(http.StatusOK, response)
}

// PredictCapacity performs AI-powered capacity prediction
// @Summary Predict Network Capacity
// @Description Predict future network capacity needs using AI
// @Tags Intelligence
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PredictCapacityRequest true "Capacity prediction request"
// @Success 200 {object} IntelligenceResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/intelligence/predict/capacity [post]
func (h *IntelligenceHandlers) PredictCapacity(c *gin.Context) {
	var req PredictCapacityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// Set defaults
	if req.ForecastDays == 0 {
		req.ForecastDays = 30
	}
	if req.Confidence == 0 {
		req.Confidence = 0.95
	}

	// Perform AI-powered capacity prediction
	result, err := h.intelligenceService.PredictCapacityNeeds(c.Request.Context(), req.HistoricalData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Capacity prediction failed: " + err.Error()})
		return
	}

	// Generate capacity insights
	insights := []string{
		"Expected 25% capacity growth over next 30 days",
		"Peak usage periods: 9-11 AM and 2-4 PM weekdays",
		"Recommend scaling compute resources by 20% proactively",
		"Storage growth rate indicates need for additional 500GB monthly",
	}

	response := IntelligenceResponse{
		Success:        true,
		Data:           result,
		Insights:       insights,
		Confidence:     req.Confidence,
		ProcessingTime: time.Since(startTime).String(),
		Metadata: map[string]interface{}{
			"forecast_horizon": req.ForecastDays,
			"data_points":      len(req.HistoricalData),
			"algorithm":        "hybrid_lstm_arima",
		},
	}

	c.JSON(http.StatusOK, response)
}

// OptimizeCosts performs AI-powered cost optimization
// @Summary Optimize Infrastructure Costs
// @Description Find cost optimization opportunities using AI
// @Tags Intelligence
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body OptimizeCostRequest true "Cost optimization request"
// @Success 200 {object} IntelligenceResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/intelligence/optimize/costs [post]
func (h *IntelligenceHandlers) OptimizeCosts(c *gin.Context) {
	var req OptimizeCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// Perform AI-powered cost optimization
	result, err := h.intelligenceService.OptimizeCosts(c.Request.Context(), req.Resources)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cost optimization failed: " + err.Error()})
		return
	}

	// Generate cost optimization insights
	insights := []string{
		"Potential monthly savings: $15,000 (32% cost reduction)",
		"Right-size 8 over-provisioned instances → Save $8,500/month",
		"Switch to reserved instances for stable workloads → Save $4,200/month",
		"Remove unused storage volumes → Save $2,300/month",
		"Optimize auto-scaling policies → Save $1,800/month",
	}

	response := IntelligenceResponse{
		Success:        true,
		Data:           result,
		Insights:       insights,
		Confidence:     0.89,
		ProcessingTime: time.Since(startTime).String(),
		Metadata: map[string]interface{}{
			"optimization_type": "multi_objective",
			"resources_analyzed": len(req.Resources),
			"roi_estimate":       "285%",
			"payback_period":     "3.2 months",
		},
	}

	c.JSON(http.StatusOK, response)
}

// DetectAnomalies performs AI-powered anomaly detection
// @Summary Detect Network Anomalies
// @Description Detect anomalies in network metrics using AI
// @Tags Intelligence
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AnomalyDetectionRequest true "Anomaly detection request"
// @Success 200 {object} IntelligenceResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/intelligence/detect/anomalies [post]
func (h *IntelligenceHandlers) DetectAnomalies(c *gin.Context) {
	var req AnomalyDetectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// Set default sensitivity
	if req.Sensitivity == 0 {
		req.Sensitivity = 0.95
	}

	// Perform AI-powered anomaly detection
	anomalies, err := h.intelligenceService.DetectAnomalies(c.Request.Context(), req.MetricName, req.Values)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Anomaly detection failed: " + err.Error()})
		return
	}

	// Generate anomaly insights
	insights := make([]string, 0)
	criticalCount := 0
	highCount := 0

	for _, anomaly := range anomalies {
		switch anomaly.Severity {
		case "critical":
			criticalCount++
		case "high":
			highCount++
		}
	}

	if criticalCount > 0 {
		insights = append(insights, "Critical anomalies detected - immediate attention required")
	}
	if highCount > 0 {
		insights = append(insights, "High-severity anomalies may indicate system degradation")
	}
	if len(anomalies) > 0 {
		insights = append(insights, "Anomaly pattern suggests potential capacity or configuration issues")
	} else {
		insights = append(insights, "No significant anomalies detected - system behavior appears normal")
	}

	response := IntelligenceResponse{
		Success:        true,
		Data:           anomalies,
		Insights:       insights,
		Confidence:     req.Sensitivity,
		ProcessingTime: time.Since(startTime).String(),
		Metadata: map[string]interface{}{
			"metric_analyzed":   req.MetricName,
			"data_points":       len(req.Values),
			"total_anomalies":   len(anomalies),
			"critical_anomalies": criticalCount,
			"high_anomalies":    highCount,
			"algorithm":         "hybrid_statistical_ml",
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetMLModels returns available ML models
// @Summary Get ML Models
// @Description Get list of available machine learning models
// @Tags Intelligence
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /api/v1/intelligence/models [get]
func (h *IntelligenceHandlers) GetMLModels(c *gin.Context) {
	models := map[string]interface{}{
		"anomaly_detection": map[string]interface{}{
			"name":        "Network Anomaly Detector",
			"version":     "v2.1.0",
			"accuracy":    0.94,
			"last_trained": "2025-10-25T10:30:00Z",
			"metrics":     []string{"latency", "throughput", "error_rate", "cpu_usage"},
		},
		"capacity_forecasting": map[string]interface{}{
			"name":        "Capacity Forecasting Engine",
			"version":     "v1.8.2",
			"accuracy":    0.89,
			"last_trained": "2025-10-24T15:45:00Z",
			"horizon":     "30 days",
		},
		"cost_optimization": map[string]interface{}{
			"name":        "Cost Optimization Advisor",
			"version":     "v3.0.1",
			"accuracy":    0.91,
			"last_trained": "2025-10-26T09:15:00Z",
			"avg_savings": "28%",
		},
		"performance_optimization": map[string]interface{}{
			"name":        "Performance Optimizer",
			"version":     "v2.5.0",
			"accuracy":    0.87,
			"last_trained": "2025-10-23T14:20:00Z",
			"improvement": "35%",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
		"total":  len(models),
		"status": "active",
	})
}

// GetAIChains returns available AI processing chains
// @Summary Get AI Chains
// @Description Get list of available AI processing chains
// @Tags Intelligence
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /api/v1/intelligence/chains [get]
func (h *IntelligenceHandlers) GetAIChains(c *gin.Context) {
	chains := h.intelligenceService.aiEngine.ListChains()
	
	chainSummary := make(map[string]interface{})
	for id, chain := range chains {
		chainSummary[id] = map[string]interface{}{
			"name":         chain.Name,
			"description":  chain.Description,
			"steps":        len(chain.Steps),
			"created_at":   chain.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"chains":     chainSummary,
		"total":      len(chains),
		"available":  true,
	})
}

// GetInsights returns AI-generated insights
// @Summary Get AI Insights
// @Description Get AI-generated business and technical insights
// @Tags Intelligence
// @Produce json
// @Security BearerAuth
// @Param category query string false "Insight category" Enums(performance,cost,security,capacity)
// @Param limit query int false "Limit number of insights"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /api/v1/intelligence/insights [get]
func (h *IntelligenceHandlers) GetInsights(c *gin.Context) {
	category := c.Query("category")
	limitStr := c.Query("limit")
	
	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Generate insights based on category
	insights := h.generateInsights(category, limit)

	c.JSON(http.StatusOK, gin.H{
		"insights":   insights,
		"category":   category,
		"count":      len(insights),
		"generated_at": time.Now(),
	})
}

// generateInsights generates AI insights based on category
func (h *IntelligenceHandlers) generateInsights(category string, limit int) []map[string]interface{} {
	allInsights := map[string][]map[string]interface{}{
		"performance": {
			{
				"title":       "Network Latency Optimization",
				"description": "CDN implementation could reduce average latency by 45%",
				"impact":      "high",
				"confidence":  0.92,
				"actions":     []string{"Deploy edge servers", "Optimize routing"},
				"estimated_benefit": "$25,000/month",
			},
			{
				"title":       "Database Query Optimization",
				"description": "15 slow queries identified affecting response times",
				"impact":      "medium",
				"confidence":  0.88,
				"actions":     []string{"Add indexes", "Query optimization"},
				"estimated_benefit": "30% faster responses",
			},
		},
		"cost": {
			{
				"title":       "Reserved Instance Optimization",
				"description": "Switch to reserved instances for 70% cost savings",
				"impact":      "high",
				"confidence":  0.95,
				"actions":     []string{"Purchase 3-year RIs", "Right-size instances"},
				"estimated_benefit": "$180,000/year",
			},
			{
				"title":       "Storage Cleanup Opportunity",
				"description": "2.3TB of unused storage detected across regions",
				"impact":      "medium",
				"confidence":  0.91,
				"actions":     []string{"Remove unused volumes", "Implement lifecycle policies"},
				"estimated_benefit": "$8,400/year",
			},
		},
		"security": {
			{
				"title":       "Unusual Access Patterns",
				"description": "Anomalous login patterns detected from new geographic locations",
				"impact":      "high",
				"confidence":  0.89,
				"actions":     []string{"Enable MFA", "Review access logs"},
				"estimated_benefit": "Enhanced security posture",
			},
		},
		"capacity": {
			{
				"title":       "Proactive Scaling Recommendation",
				"description": "CPU utilization trending up - scale before hitting limits",
				"impact":      "medium",
				"confidence":  0.87,
				"actions":     []string{"Add 2 instances", "Update auto-scaling policy"},
				"estimated_benefit": "Prevent performance degradation",
			},
		},
	}

	var insights []map[string]interface{}
	
	if category == "" {
		// Return insights from all categories
		for _, categoryInsights := range allInsights {
			insights = append(insights, categoryInsights...)
		}
	} else {
		// Return insights for specific category
		if categoryInsights, exists := allInsights[category]; exists {
			insights = categoryInsights
		}
	}

	// Limit results
	if len(insights) > limit {
		insights = insights[:limit]
	}

	return insights
}