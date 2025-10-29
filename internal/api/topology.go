package api

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	
	"netorchestrator/internal/models"
)

// GetNetworkTopology returns network topology data for visualization
func (h *Handlers) GetNetworkTopology(c *gin.Context) {
	networkIDStr := c.Param("id")
	networkID, err := uuid.Parse(networkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	// Get network details
	network, err := h.networkService.GetNetwork(c.Request.Context(), networkID)
	if err != nil {
		h.logger.Error("Failed to get network", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get network",
		})
		return
	}

	// Get nodes
	nodes, err := h.networkService.ListNodes(c.Request.Context(), networkID)
	if err != nil {
		h.logger.Error("Failed to get nodes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get nodes",
		})
		return
	}

	// Get links
	links, err := h.networkService.ListLinks(c.Request.Context(), networkID)
	if err != nil {
		h.logger.Error("Failed to get links", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get links",
		})
		return
	}

	// Format for visualization
	topology := gin.H{
		"network": gin.H{
			"id":          network.ID,
			"name":        network.Name,
			"description": network.Description,
			"status":      network.Status,
			"subnet":      network.Config.Subnet,
			"gateway":     network.Config.Gateway,
		},
		"nodes": formatNodesForViz(nodes),
		"links": formatLinksForViz(links),
	}

	c.JSON(http.StatusOK, topology)
}

// GetNetworkStats returns comprehensive network statistics
func (h *Handlers) GetNetworkStats(c *gin.Context) {
	networkIDStr := c.Param("id")
	networkID, err := uuid.Parse(networkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	// Get basic counts
	nodes, _ := h.networkService.ListNodes(c.Request.Context(), networkID)
	links, _ := h.networkService.ListLinks(c.Request.Context(), networkID)
	policies, _ := h.networkService.ListPolicies(c.Request.Context(), networkID)

	// Calculate stats
	activeNodes := 0
	totalCPU := 0
	totalMemory := 0
	totalDisk := 0

	for _, node := range nodes {
		if node.Status == models.NodeStatusActive {
			activeNodes++
		}
		totalCPU += node.Config.CPU
		totalMemory += node.Config.Memory
		totalDisk += node.Config.Storage
	}

	stats := gin.H{
		"overview": gin.H{
			"total_nodes":    len(nodes),
			"active_nodes":   activeNodes,
			"total_links":    len(links),
			"total_policies": len(policies),
		},
		"resources": gin.H{
			"total_cpu_cores": totalCPU,
			"total_memory_mb": totalMemory,
			"total_disk_gb":   totalDisk,
		},
		"health": gin.H{
			"node_health_ratio": float64(activeNodes) / float64(len(nodes)) * 100,
			"status":           "healthy", // TODO: Calculate based on actual health checks
		},
	}

	c.JSON(http.StatusOK, stats)
}

// LiveNetworkEvents establishes WebSocket connection for real-time updates
func (h *Handlers) LiveNetworkEvents(c *gin.Context) {
	// This would be handled by the WebSocket hub
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "WebSocket endpoint - use ws:// protocol",
		"info":  "Connect to /ws/networks/:id/events for real-time updates",
	})
}

// Helper functions for data formatting
func formatNodesForViz(nodes []models.Node) []gin.H {
	var result []gin.H
	for _, node := range nodes {
		result = append(result, gin.H{
			"id":         node.ID,
			"name":       node.Name,
			"status":     node.Status,
			"type":       node.Type,
			"ip_address": node.IPAddress,
			"cpu":        node.Config.CPU,
			"memory":     node.Config.Memory,
			"storage":    node.Config.Storage,
			"position": gin.H{
				"x": node.Position.X,
				"y": node.Position.Y,
			},
		})
	}
	return result
}

func formatLinksForViz(links []models.Link) []gin.H {
	var result []gin.H
	for _, link := range links {
		result = append(result, gin.H{
			"id":        link.ID,
			"source":    link.SourceNodeID,
			"target":    link.TargetNodeID,
			"name":      link.Name,
			"status":    link.Status,
			"bandwidth": link.Config.Bandwidth,
			"latency":   link.Config.Latency,
		})
	}
	return result
}