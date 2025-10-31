package devices

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ContainerHandlers handles HTTP requests for container device management
type ContainerHandlers struct {
	containerManager *ContainerDeviceManager
	logger           *zap.Logger
}

// NewContainerHandlers creates new container device handlers
func NewContainerHandlers(logger *zap.Logger) *ContainerHandlers {
	containerManager := NewContainerDeviceManager(logger)

	return &ContainerHandlers{
		containerManager: containerManager,
		logger:           logger,
	}
}

// ListDevices lists all discovered container devices
func (h *ContainerHandlers) ListDevices(c *gin.Context) {
	devices := h.containerManager.ListDevices()

	c.JSON(http.StatusOK, gin.H{
		"devices": devices,
		"count":   len(devices),
		"status":  "success",
		"message": "Container devices discovered via Podman runtime",
		"lab_info": gin.H{
			"type":    "Container-based Network Lab",
			"runtime": "Podman",
			"network": "netorchestrator-lab",
		},
	})
}

// GetDevice retrieves a specific container device
func (h *ContainerHandlers) GetDevice(c *gin.Context) {
	deviceID := c.Param("id")

	device, err := h.containerManager.GetDevice(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"device":  device,
		"status":  "success",
		"message": "Real network container device information",
	})
}

// GetDeviceInfo retrieves detailed device information from live container
func (h *ContainerHandlers) GetDeviceInfo(c *gin.Context) {
	deviceID := c.Param("id")

	info, err := h.containerManager.GetDeviceInfo(deviceID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to get device info",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"device_info": info,
		"status":      "success",
		"message":     "Live data retrieved from network container via Podman exec",
		"data_source": "Real container runtime",
	})
}

// GetDeviceInterfaces retrieves network interfaces from live container
func (h *ContainerHandlers) GetDeviceInterfaces(c *gin.Context) {
	deviceID := c.Param("id")

	interfaces, err := h.containerManager.GetInterfaces(deviceID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to get interfaces",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"interfaces": interfaces,
		"count":      len(interfaces),
		"status":     "success",
		"message":    "Network interfaces retrieved from live container",
		"method":     "ip addr show via podman exec",
	})
}

// TestConnectivity tests connectivity to container device
func (h *ContainerHandlers) TestConnectivity(c *gin.Context) {
	deviceID := c.Param("id")

	err := h.containerManager.TestConnectivity(deviceID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "failed",
			"error":   err.Error(),
			"message": "Container connectivity test failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Container connectivity test passed - Podman exec responding",
		"method":  "podman exec connectivity test",
	})
}

// ExecuteCommand executes a command on the container device
func (h *ContainerHandlers) ExecuteCommand(c *gin.Context) {
	deviceID := c.Param("id")

	var req struct {
		Command string `json:"command" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	output, err := h.containerManager.ExecuteCommand(deviceID, req.Command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Command execution failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"command": req.Command,
		"output":  output,
		"message": "Command executed successfully on network container",
		"method":  "podman exec",
	})
}

// GetStats returns container lab statistics
func (h *ContainerHandlers) GetStats(c *gin.Context) {
	stats := h.containerManager.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"lab_stats": stats,
		"lab_type":  "Container-based Network Lab",
		"runtime":   "Podman",
		"network":   "netorchestrator-lab",
		"status":    "success",
		"message":   "Real container network statistics",
	})
}

// GetRoutes returns routing information from container (if available)
func (h *ContainerHandlers) GetRoutes(c *gin.Context) {
	deviceID := c.Param("id")

	// Execute routing table command
	output, err := h.containerManager.ExecuteCommand(deviceID, "ip route")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to get routing table",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"routing_table": output,
		"status":        "success",
		"message":       "Live routing table from network container",
		"command":       "ip route",
	})
}

// GetOSPF returns OSPF information from FRR container (if applicable)
func (h *ContainerHandlers) GetOSPF(c *gin.Context) {
	deviceID := c.Param("id")

	device, err := h.containerManager.GetDevice(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	if device.Type != "frr" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "OSPF only available on FRRouting devices",
			"device_type": device.Type,
		})
		return
	}

	// Execute OSPF neighbor command
	output, err := h.containerManager.ExecuteCommand(deviceID, "vtysh -c 'show ip ospf neighbor'")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to get OSPF information",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ospf_neighbors": output,
		"status":         "success",
		"message":        "Live OSPF data from FRRouting container",
		"command":        "vtysh -c 'show ip ospf neighbor'",
	})
}

// RefreshDevices rediscovers container devices
func (h *ContainerHandlers) RefreshDevices(c *gin.Context) {
	h.containerManager.discoverContainers()
	devices := h.containerManager.ListDevices()

	c.JSON(http.StatusOK, gin.H{
		"devices":      devices,
		"count":        len(devices),
		"status":       "success",
		"message":      "Container devices rediscovered",
		"refreshed_at": "now",
	})
}

// RegisterContainerRoutes registers all container device routes
func (h *ContainerHandlers) RegisterRoutes(router *gin.RouterGroup) {
	devices := router.Group("/devices")
	{
		devices.GET("", h.ListDevices)
		devices.GET("/refresh", h.RefreshDevices)
		devices.GET("/stats", h.GetStats)
		devices.GET("/:id", h.GetDevice)
		devices.GET("/:id/info", h.GetDeviceInfo)
		devices.GET("/:id/interfaces", h.GetDeviceInterfaces)
		devices.GET("/:id/routes", h.GetRoutes)
		devices.GET("/:id/ospf", h.GetOSPF)
		devices.POST("/:id/test", h.TestConnectivity)
		devices.POST("/:id/execute", h.ExecuteCommand)
	}
}
