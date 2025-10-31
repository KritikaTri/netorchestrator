package devices

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DeviceType represents different types of network devices
type DeviceType string

const (
	DeviceTypeCisco   DeviceType = "cisco"
	DeviceTypeJuniper DeviceType = "juniper"
	DeviceTypeArista  DeviceType = "arista"
)

// DeviceStatus represents the operational status of a device
type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "online"
	DeviceStatusOffline DeviceStatus = "offline"
	DeviceStatusUnknown DeviceStatus = "unknown"
	DeviceStatusError   DeviceStatus = "error"
)

// NetworkDevice represents a managed network device
type NetworkDevice struct {
	ID        string                 `json:"id" gorm:"primaryKey"`
	Name      string                 `json:"name"`
	Type      DeviceType             `json:"type"`
	Vendor    string                 `json:"vendor"`
	Model     string                 `json:"model"`
	Host      string                 `json:"host"`
	Username  string                 `json:"username"`
	Password  string                 `json:"password"` // In production, encrypt this
	Config    map[string]interface{} `json:"config" gorm:"serializer:json;type:jsonb"`
	Status    DeviceStatus           `json:"status"`
	LastSeen  time.Time              `json:"last_seen"`
	NetworkID string                 `json:"network_id,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// DeviceManager manages network devices across multiple vendors
type DeviceManager struct {
	db      *gorm.DB
	logger  *zap.Logger
	devices map[string]DeviceInterface
	mutex   sync.RWMutex
}

// DeviceInterface defines the interface for different device types
type DeviceInterface interface {
	GetDeviceInfo() (*DeviceInfo, error)
	GetInterfaces() ([]InterfaceConfig, error)
	ConfigureInterface(config InterfaceConfig) error
	TestConnectivity() error
}

// NewDeviceManager creates a new device manager
func NewDeviceManager(db *gorm.DB, logger *zap.Logger) *DeviceManager {
	return &DeviceManager{
		db:      db,
		logger:  logger,
		devices: make(map[string]DeviceInterface),
	}
}

// RegisterDevice registers a new network device
func (dm *DeviceManager) RegisterDevice(device *NetworkDevice) error {
	// Generate ID if not provided
	if device.ID == "" {
		device.ID = fmt.Sprintf("device_%d", time.Now().UnixNano())
	}

	// Create device connection based on type
	deviceInterface, err := dm.createDeviceInterface(device)
	if err != nil {
		return fmt.Errorf("failed to create device interface: %w", err)
	}

	// Test connectivity
	if err := deviceInterface.TestConnectivity(); err != nil {
		device.Status = DeviceStatusError
		dm.logger.Warn("Device connectivity test failed",
			zap.String("device_id", device.ID),
			zap.Error(err))
	} else {
		device.Status = DeviceStatusOnline
		device.LastSeen = time.Now()
	}

	// Save to database
	if err := dm.db.Create(device).Error; err != nil {
		return fmt.Errorf("failed to save device: %w", err)
	}

	// Store device interface
	dm.mutex.Lock()
	dm.devices[device.ID] = deviceInterface
	dm.mutex.Unlock()

	dm.logger.Info("Device registered successfully",
		zap.String("device_id", device.ID),
		zap.String("name", device.Name),
		zap.String("type", string(device.Type)))

	return nil
}

// GetDevice retrieves a device by ID
func (dm *DeviceManager) GetDevice(deviceID string) (*NetworkDevice, error) {
	var device NetworkDevice
	if err := dm.db.First(&device, "id = ?", deviceID).Error; err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}
	return &device, nil
}

// ListDevices lists all registered devices
func (dm *DeviceManager) ListDevices() ([]NetworkDevice, error) {
	var devices []NetworkDevice
	if err := dm.db.Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}
	return devices, nil
}

// GetDeviceInfo retrieves detailed information about a device
func (dm *DeviceManager) GetDeviceInfo(deviceID string) (*DeviceInfo, error) {
	dm.mutex.RLock()
	deviceInterface, exists := dm.devices[deviceID]
	dm.mutex.RUnlock()

	if !exists {
		// Try to load device from database and create interface
		device, err := dm.GetDevice(deviceID)
		if err != nil {
			return nil, err
		}

		deviceInterface, err = dm.createDeviceInterface(device)
		if err != nil {
			return nil, fmt.Errorf("failed to create device interface: %w", err)
		}

		dm.mutex.Lock()
		dm.devices[deviceID] = deviceInterface
		dm.mutex.Unlock()
	}

	return deviceInterface.GetDeviceInfo()
}

// GetDeviceInterfaces retrieves all interfaces for a device
func (dm *DeviceManager) GetDeviceInterfaces(deviceID string) ([]InterfaceConfig, error) {
	dm.mutex.RLock()
	deviceInterface, exists := dm.devices[deviceID]
	dm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("device interface not found")
	}

	return deviceInterface.GetInterfaces()
}

// ConfigureDeviceInterface configures an interface on a device
func (dm *DeviceManager) ConfigureDeviceInterface(deviceID string, config InterfaceConfig) error {
	dm.mutex.RLock()
	deviceInterface, exists := dm.devices[deviceID]
	dm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("device interface not found")
	}

	return deviceInterface.ConfigureInterface(config)
}

// MonitorDevices starts monitoring all registered devices
func (dm *DeviceManager) MonitorDevices(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			dm.performHealthCheck()
		}
	}
}

// performHealthCheck checks the health of all devices
func (dm *DeviceManager) performHealthCheck() {
	devices, err := dm.ListDevices()
	if err != nil {
		dm.logger.Error("Failed to list devices for health check", zap.Error(err))
		return
	}

	for _, device := range devices {
		go dm.checkDeviceHealth(&device)
	}
}

// checkDeviceHealth checks the health of a single device
func (dm *DeviceManager) checkDeviceHealth(device *NetworkDevice) {
	dm.mutex.RLock()
	deviceInterface, exists := dm.devices[device.ID]
	dm.mutex.RUnlock()

	if !exists {
		// Try to create device interface
		var err error
		deviceInterface, err = dm.createDeviceInterface(device)
		if err != nil {
			dm.logger.Error("Failed to create device interface for health check",
				zap.String("device_id", device.ID), zap.Error(err))
			return
		}

		dm.mutex.Lock()
		dm.devices[device.ID] = deviceInterface
		dm.mutex.Unlock()
	}

	// Test connectivity
	err := deviceInterface.TestConnectivity()
	previousStatus := device.Status

	if err != nil {
		device.Status = DeviceStatusOffline
		dm.logger.Warn("Device health check failed",
			zap.String("device_id", device.ID),
			zap.String("name", device.Name),
			zap.Error(err))
	} else {
		device.Status = DeviceStatusOnline
		device.LastSeen = time.Now()
	}

	// Update status in database if changed
	if device.Status != previousStatus {
		if err := dm.db.Save(device).Error; err != nil {
			dm.logger.Error("Failed to update device status",
				zap.String("device_id", device.ID), zap.Error(err))
		} else {
			dm.logger.Info("Device status updated",
				zap.String("device_id", device.ID),
				zap.String("old_status", string(previousStatus)),
				zap.String("new_status", string(device.Status)))
		}
	}
}

// createDeviceInterface creates the appropriate device interface based on type
func (dm *DeviceManager) createDeviceInterface(device *NetworkDevice) (DeviceInterface, error) {
	switch device.Type {
	case DeviceTypeCisco:
		return NewCiscoDevice(device.Host, device.Username, device.Password, dm.logger), nil
	case DeviceTypeJuniper:
		// TODO: Implement Juniper device interface
		return nil, fmt.Errorf("juniper devices not yet implemented")
	case DeviceTypeArista:
		// TODO: Implement Arista device interface
		return nil, fmt.Errorf("arista devices not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported device type: %s", device.Type)
	}
}

// DeleteDevice removes a device from management
func (dm *DeviceManager) DeleteDevice(deviceID string) error {
	// Remove from memory
	dm.mutex.Lock()
	delete(dm.devices, deviceID)
	dm.mutex.Unlock()

	// Remove from database
	if err := dm.db.Delete(&NetworkDevice{}, "id = ?", deviceID).Error; err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	dm.logger.Info("Device deleted successfully", zap.String("device_id", deviceID))
	return nil
}

// GetDevicesByNetwork retrieves all devices for a specific network
func (dm *DeviceManager) GetDevicesByNetwork(networkID string) ([]NetworkDevice, error) {
	var devices []NetworkDevice
	if err := dm.db.Where("network_id = ?", networkID).Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("failed to get devices for network: %w", err)
	}
	return devices, nil
}
