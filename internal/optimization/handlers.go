package optimization

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handlers handles optimization API endpoints
type Handlers struct {
	service *OptimizationService
	logger  *zap.Logger
}

// NewHandlers creates new optimization handlers
func NewHandlers(service *OptimizationService, logger *zap.Logger) *Handlers {
	return &Handlers{
		service: service,
		logger:  logger,
	}
}

// AnalyzeNetwork handles network analysis requests
// @Summary Analyze Network
// @Description Performs comprehensive AI-powered analysis of network topology
// @Tags Optimization
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param network_id path string true "Network ID"
// @Success 200 {object} NetworkAnalysis
// @Router /api/v1/optimization/analyze/{network_id} [post]
func (h *Handlers) AnalyzeNetwork(c *gin.Context) {
	networkIDStr := c.Param("network_id")
	networkID, err := uuid.Parse(networkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	analysis, err := h.service.AnalyzeNetwork(c.Request.Context(), networkID)
	if err != nil {
		h.logger.Error("Failed to analyze network", zap.Error(err))
		if err.Error() == "network not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Network not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to analyze network",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis": analysis,
		"status":   "success",
	})
}

// RecommendTopologyChange handles topology recommendation requests
// @Summary Get Topology Recommendations
// @Description Get AI-powered topology optimization recommendations
// @Tags Optimization
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param network_id path string true "Network ID"
// @Success 200 {object} Recommendation
// @Router /api/v1/optimization/recommend/{network_id} [post]
func (h *Handlers) RecommendTopologyChange(c *gin.Context) {
	networkIDStr := c.Param("network_id")
	networkID, err := uuid.Parse(networkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	recommendation, err := h.service.RecommendTopologyChange(c.Request.Context(), networkID)
	if err != nil {
		if err.Error() == "network not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Network not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "No topology recommendations available at this time",
			"status":  "info",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendation": recommendation,
		"status":         "success",
	})
}

// ValidateConfiguration handles configuration validation requests
// @Summary Validate Network Configuration
// @Description Validates network configuration for issues and compliance
// @Tags Optimization
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param network_id path string true "Network ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/optimization/validate/{network_id} [post]
func (h *Handlers) ValidateConfiguration(c *gin.Context) {
	networkIDStr := c.Param("network_id")
	networkID, err := uuid.Parse(networkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	issues, err := h.service.ValidateConfiguration(c.Request.Context(), networkID)
	if err != nil {
		if err.Error() == "network not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Network not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to validate configuration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issues": issues,
		"count":  len(issues),
		"status": "success",
	})
}

