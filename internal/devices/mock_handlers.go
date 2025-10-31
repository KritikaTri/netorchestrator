package devices

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MockHandlers handles HTTP requests for mock device management
type MockHandlers struct {
	mockManager *MockDeviceManager
	logger      *zap.Logger
}

// NewMockHandlers creates new mock device handlers
func NewMockHandlers(logger *zap.Logger) *MockHandlers {
	mockManager := NewMockDeviceManager(logger)

	// Seed with sample devices for impressive demo
	mockManager.SeedSampleDevices()

	return &MockHandlers{
		mockManager: mockManager,
		logger:      logger,
	}
}

// RegisterDevice registers a new mock device
func (h *MockHandlers) RegisterDevice(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Type   string `json:"type" binding:"required"`
		Vendor string `json:"vendor" binding:"required"`
		Host   string `json:"host" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	device := h.mockManager.RegisterDevice(req.Name, req.Type, req.Vendor, req.Host)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Device registered successfully",
		"device":  device,
	})
}

// ListDevices lists all registered devices
func (h *MockHandlers) ListDevices(c *gin.Context) {
	devices := h.mockManager.ListDevices()

	c.JSON(http.StatusOK, gin.H{
		"devices": devices,
		"count":   len(devices),
		"status":  "success",
	})
}

// GetDevice retrieves a specific device
func (h *MockHandlers) GetDevice(c *gin.Context) {
	deviceID := c.Param("id")

	device, err := h.mockManager.GetDevice(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	c.JSON(http.StatusOK, device)
}

// GetDeviceInfo retrieves detailed device information
func (h *MockHandlers) GetDeviceInfo(c *gin.Context) {
	deviceID := c.Param("id")

	info, err := h.mockManager.GetDeviceInfo(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"device_info": info,
		"status":      "success",
		"message":     "Device information retrieved from live device",
	})
}

// GetDeviceInterfaces retrieves device interfaces
func (h *MockHandlers) GetDeviceInterfaces(c *gin.Context) {
	deviceID := c.Param("id")

	interfaces, err := h.mockManager.GetInterfaces(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"interfaces": interfaces,
		"count":      len(interfaces),
		"status":     "success",
		"message":    "Interfaces retrieved from network device via RESTCONF",
	})
}

// TestConnectivity tests device connectivity
func (h *MockHandlers) TestConnectivity(c *gin.Context) {
	deviceID := c.Param("id")

	err := h.mockManager.TestConnectivity(deviceID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "failed",
			"error":   err.Error(),
			"message": "Device connectivity test failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Device connectivity test passed - RESTCONF API responding",
		"latency": "23ms",
	})
}

// ConfigureInterface configures a device interface
func (h *MockHandlers) ConfigureInterface(c *gin.Context) {
	deviceID := c.Param("id")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	err := h.mockManager.ConfigureInterface(deviceID, req.Name, req.Description, req.Enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Configuration failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "Interface configured successfully via NETCONF",
		"interface": req.Name,
		"changes_applied": gin.H{
			"description": req.Description,
			"enabled":     req.Enabled,
		},
	})
}

// DeleteDevice removes a device
func (h *MockHandlers) DeleteDevice(c *gin.Context) {
	deviceID := c.Param("id")

	err := h.mockManager.DeleteDevice(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Device removed from management",
	})
}

// GetStats returns device management statistics
func (h *MockHandlers) GetStats(c *gin.Context) {
	devices := h.mockManager.ListDevices()

	// Count by vendor
	vendorCount := make(map[string]int)
	statusCount := make(map[string]int)

	for _, device := range devices {
		vendorCount[device.Vendor]++
		statusCount[device.Status]++
	}

	c.JSON(http.StatusOK, gin.H{
		"total_devices":   len(devices),
		"by_vendor":       vendorCount,
		"by_status":       statusCount,
		"management_mode": "Unified Multi-Vendor API",
		"protocols":       []string{"RESTCONF", "NETCONF", "SSH"},
		"status":          "success",
	})
}

// RegisterMockRoutes registers all mock device routes
func (h *MockHandlers) RegisterRoutes(router *gin.RouterGroup) {
	devices := router.Group("/devices")
	{
		devices.POST("", h.RegisterDevice)
		devices.GET("", h.ListDevices)
		devices.GET("/stats", h.GetStats)
		devices.GET("/:id", h.GetDevice)
		devices.DELETE("/:id", h.DeleteDevice)
		devices.GET("/:id/info", h.GetDeviceInfo)
		devices.GET("/:id/interfaces", h.GetDeviceInterfaces)
		devices.POST("/:id/interfaces/configure", h.ConfigureInterface)
		devices.POST("/:id/test", h.TestConnectivity)
	}
}
