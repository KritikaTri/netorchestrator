package devices

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DeviceHandlers handles HTTP requests for device management
type DeviceHandlers struct {
	deviceManager *DeviceManager
	logger        *zap.Logger
}

// NewDeviceHandlers creates new device handlers
func NewDeviceHandlers(deviceManager *DeviceManager, logger *zap.Logger) *DeviceHandlers {
	return &DeviceHandlers{
		deviceManager: deviceManager,
		logger:        logger,
	}
}

// RegisterDeviceRequest represents a device registration request
type RegisterDeviceRequest struct {
	Name      string                 `json:"name" binding:"required"`
	Type      string                 `json:"type" binding:"required"`
	Vendor    string                 `json:"vendor" binding:"required"`
	Host      string                 `json:"host" binding:"required"`
	Username  string                 `json:"username" binding:"required"`
	Password  string                 `json:"password" binding:"required"`
	NetworkID string                 `json:"network_id,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

// ConfigureInterfaceRequest represents an interface configuration request
type ConfigureInterfaceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
	IPAddress   string `json:"ip_address,omitempty"`
	SubnetMask  string `json:"subnet_mask,omitempty"`
	Enabled     bool   `json:"enabled"`
}

// RegisterDevice registers a new network device
// @Summary Register a new network device
// @Description Register a network device for management and monitoring
// @Tags devices
// @Accept json
// @Produce json
// @Param request body RegisterDeviceRequest true "Device registration details"
// @Success 201 {object} NetworkDevice
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /devices [post]
func (h *DeviceHandlers) RegisterDevice(c *gin.Context) {
	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate device type
	deviceType := DeviceType(req.Type)
	if deviceType != DeviceTypeCisco && deviceType != DeviceTypeJuniper && deviceType != DeviceTypeArista {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported device type. Supported types: cisco, juniper, arista"})
		return
	}

	// Create device
	device := &NetworkDevice{
		Name:      req.Name,
		Type:      deviceType,
		Vendor:    req.Vendor,
		Host:      req.Host,
		Username:  req.Username,
		Password:  req.Password, // TODO: Encrypt in production
		NetworkID: req.NetworkID,
		Config:    req.Config,
	}

	if err := h.deviceManager.RegisterDevice(device); err != nil {
		h.logger.Error("Failed to register device", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register device", "details": err.Error()})
		return
	}

	h.logger.Info("Device registered successfully", zap.String("device_id", device.ID))
	c.JSON(http.StatusCreated, device)
}

// ListDevices lists all registered devices
// @Summary List all network devices
// @Description Get a list of all registered network devices
// @Tags devices
// @Produce json
// @Success 200 {array} NetworkDevice
// @Failure 500 {object} map[string]string
// @Router /devices [get]
func (h *DeviceHandlers) ListDevices(c *gin.Context) {
	devices, err := h.deviceManager.ListDevices()
	if err != nil {
		h.logger.Error("Failed to list devices", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list devices", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"devices": devices, "count": len(devices)})
}

// GetDevice retrieves a specific device
// @Summary Get device details
// @Description Get detailed information about a specific network device
// @Tags devices
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} NetworkDevice
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /devices/{id} [get]
func (h *DeviceHandlers) GetDevice(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device ID is required"})
		return
	}

	device, err := h.deviceManager.GetDevice(deviceID)
	if err != nil {
		h.logger.Error("Failed to get device", zap.String("device_id", deviceID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, device)
}

// GetDeviceInfo retrieves detailed device information
// @Summary Get device information
// @Description Get detailed technical information about a network device
// @Tags devices
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} DeviceInfo
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /devices/{id}/info [get]
func (h *DeviceHandlers) GetDeviceInfo(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device ID is required"})
		return
	}

	info, err := h.deviceManager.GetDeviceInfo(deviceID)
	if err != nil {
		h.logger.Error("Failed to get device info", zap.String("device_id", deviceID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get device info", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// GetDeviceInterfaces retrieves all interfaces for a device
// @Summary Get device interfaces
// @Description Get a list of all network interfaces on a device
// @Tags devices
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {array} InterfaceConfig
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /devices/{id}/interfaces [get]
func (h *DeviceHandlers) GetDeviceInterfaces(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device ID is required"})
		return
	}

	interfaces, err := h.deviceManager.GetDeviceInterfaces(deviceID)
	if err != nil {
		h.logger.Error("Failed to get device interfaces", zap.String("device_id", deviceID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get device interfaces", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"interfaces": interfaces, "count": len(interfaces)})
}

// ConfigureInterface configures a network interface
// @Summary Configure device interface
// @Description Configure a network interface on a device
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Param request body ConfigureInterfaceRequest true "Interface configuration"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /devices/{id}/interfaces/configure [post]
func (h *DeviceHandlers) ConfigureInterface(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device ID is required"})
		return
	}

	var req ConfigureInterfaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	config := InterfaceConfig{
		Name:        req.Name,
		Description: req.Description,
		IPAddress:   req.IPAddress,
		SubnetMask:  req.SubnetMask,
		Enabled:     req.Enabled,
	}

	if err := h.deviceManager.ConfigureDeviceInterface(deviceID, config); err != nil {
		h.logger.Error("Failed to configure interface",
			zap.String("device_id", deviceID),
			zap.String("interface", req.Name),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to configure interface", "details": err.Error()})
		return
	}

	h.logger.Info("Interface configured successfully",
		zap.String("device_id", deviceID),
		zap.String("interface", req.Name))

	c.JSON(http.StatusOK, gin.H{"message": "Interface configured successfully"})
}

// DeleteDevice removes a device from management
// @Summary Delete device
// @Description Remove a network device from management
// @Tags devices
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /devices/{id} [delete]
func (h *DeviceHandlers) DeleteDevice(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device ID is required"})
		return
	}

	if err := h.deviceManager.DeleteDevice(deviceID); err != nil {
		h.logger.Error("Failed to delete device", zap.String("device_id", deviceID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete device", "details": err.Error()})
		return
	}

	h.logger.Info("Device deleted successfully", zap.String("device_id", deviceID))
	c.JSON(http.StatusOK, gin.H{"message": "Device deleted successfully"})
}

// TestDeviceConnectivity tests connectivity to a device
// @Summary Test device connectivity
// @Description Test network connectivity to a device
// @Tags devices
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /devices/{id}/test [post]
func (h *DeviceHandlers) TestDeviceConnectivity(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device ID is required"})
		return
	}

	// Get device from database
	device, err := h.deviceManager.GetDevice(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	// Create device interface for testing
	deviceInterface, err := h.deviceManager.createDeviceInterface(device)
	if err != nil {
		h.logger.Error("Failed to create device interface", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create device interface"})
		return
	}

	// Test connectivity
	if err := deviceInterface.TestConnectivity(); err != nil {
		h.logger.Error("Device connectivity test failed",
			zap.String("device_id", deviceID), zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "failed",
			"error":   err.Error(),
			"message": "Device is not reachable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Device connectivity test passed",
	})
}

// RegisterRoutes registers all device management routes
func (h *DeviceHandlers) RegisterRoutes(router *gin.RouterGroup) {
	devices := router.Group("/devices")
	{
		devices.POST("", h.RegisterDevice)
		devices.GET("", h.ListDevices)
		devices.GET("/:id", h.GetDevice)
		devices.DELETE("/:id", h.DeleteDevice)
		devices.GET("/:id/info", h.GetDeviceInfo)
		devices.GET("/:id/interfaces", h.GetDeviceInterfaces)
		devices.POST("/:id/interfaces/configure", h.ConfigureInterface)
		devices.POST("/:id/test", h.TestDeviceConnectivity)
	}
}
