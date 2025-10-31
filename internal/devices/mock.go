package devices

import (
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// MockDeviceManager provides a simple in-memory device management system
type MockDeviceManager struct {
	devices map[string]*MockDevice
	logger  *zap.Logger
}

// MockDevice represents a virtual network device with realistic responses
type MockDevice struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"` // cisco, juniper, arista
	Vendor   string `json:"vendor"`
	Host     string `json:"host"`
	Status   string `json:"status"` // online, offline
	LastSeen string `json:"last_seen"`
}

// MockDeviceInfo represents device information
type MockDeviceInfo struct {
	Hostname     string `json:"hostname"`
	Model        string `json:"model"`
	Version      string `json:"version"`
	SerialNumber string `json:"serial_number"`
	Uptime       string `json:"uptime"`
	CPUUsage     string `json:"cpu_usage"`
	MemoryUsage  string `json:"memory_usage"`
}

// MockInterface represents a network interface
type MockInterface struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	IPAddress   string `json:"ip_address"`
	Speed       string `json:"speed"`
	Duplex      string `json:"duplex"`
	MTU         int    `json:"mtu"`
}

// NewMockDeviceManager creates a new mock device manager
func NewMockDeviceManager(logger *zap.Logger) *MockDeviceManager {
	return &MockDeviceManager{
		devices: make(map[string]*MockDevice),
		logger:  logger,
	}
}

// RegisterDevice registers a new mock device
func (m *MockDeviceManager) RegisterDevice(name, deviceType, vendor, host string) *MockDevice {
	deviceID := fmt.Sprintf("device_%d", time.Now().UnixNano())

	device := &MockDevice{
		ID:       deviceID,
		Name:     name,
		Type:     deviceType,
		Vendor:   vendor,
		Host:     host,
		Status:   "online",
		LastSeen: time.Now().Format(time.RFC3339),
	}

	m.devices[deviceID] = device
	m.logger.Info("Mock device registered",
		zap.String("id", deviceID),
		zap.String("name", name),
		zap.String("type", deviceType))

	return device
}

// ListDevices returns all registered devices
func (m *MockDeviceManager) ListDevices() []*MockDevice {
	devices := make([]*MockDevice, 0, len(m.devices))
	for _, device := range m.devices {
		devices = append(devices, device)
	}
	return devices
}

// GetDevice retrieves a device by ID
func (m *MockDeviceManager) GetDevice(deviceID string) (*MockDevice, error) {
	device, exists := m.devices[deviceID]
	if !exists {
		return nil, fmt.Errorf("device not found")
	}
	return device, nil
}

// GetDeviceInfo returns realistic device information
func (m *MockDeviceManager) GetDeviceInfo(deviceID string) (*MockDeviceInfo, error) {
	device, err := m.GetDevice(deviceID)
	if err != nil {
		return nil, err
	}

	// Generate realistic device info based on vendor
	var model, version string
	switch device.Type {
	case "cisco":
		model = "Catalyst 8000V"
		version = "17.03.04a"
	case "juniper":
		model = "vMX"
		version = "20.4R3.8"
	case "arista":
		model = "vEOS"
		version = "4.27.0F"
	default:
		model = "Generic Router"
		version = "1.0.0"
	}

	// Generate random but realistic metrics
	cpuUsage := rand.Intn(80) + 5     // 5-85%
	memoryUsage := rand.Intn(60) + 20 // 20-80%
	uptimeDays := rand.Intn(200) + 1  // 1-200 days

	return &MockDeviceInfo{
		Hostname:     device.Name,
		Model:        model,
		Version:      version,
		SerialNumber: fmt.Sprintf("%s%d", device.Type[:2], rand.Intn(999999)+100000),
		Uptime: fmt.Sprintf("%d days, %d:%02d:%02d",
			uptimeDays,
			rand.Intn(24),
			rand.Intn(60),
			rand.Intn(60)),
		CPUUsage:    fmt.Sprintf("%d%%", cpuUsage),
		MemoryUsage: fmt.Sprintf("%d%%", memoryUsage),
	}, nil
}

// GetInterfaces returns mock network interfaces
func (m *MockDeviceManager) GetInterfaces(deviceID string) ([]*MockInterface, error) {
	device, err := m.GetDevice(deviceID)
	if err != nil {
		return nil, err
	}

	interfaces := []*MockInterface{
		{
			Name:        "GigabitEthernet0/0/0",
			Description: "WAN Interface",
			Status:      "up",
			IPAddress:   "192.168.1.1",
			Speed:       "1000",
			Duplex:      "full",
			MTU:         1500,
		},
		{
			Name:        "GigabitEthernet0/0/1",
			Description: "LAN Interface",
			Status:      "up",
			IPAddress:   "10.0.0.1",
			Speed:       "1000",
			Duplex:      "full",
			MTU:         1500,
		},
		{
			Name:        "Loopback0",
			Description: "Management Interface",
			Status:      "up",
			IPAddress:   "127.0.0.1",
			Speed:       "8000000",
			Duplex:      "full",
			MTU:         1514,
		},
	}

	// Add vendor-specific interfaces
	switch device.Type {
	case "juniper":
		interfaces = append(interfaces, &MockInterface{
			Name:        "xe-0/0/0",
			Description: "Juniper 10GE Interface",
			Status:      "up",
			IPAddress:   "172.16.0.1",
			Speed:       "10000",
			Duplex:      "full",
			MTU:         9000,
		})
	case "arista":
		interfaces = append(interfaces, &MockInterface{
			Name:        "Ethernet1",
			Description: "Arista Ethernet Interface",
			Status:      "up",
			IPAddress:   "203.0.113.1",
			Speed:       "25000",
			Duplex:      "full",
			MTU:         1500,
		})
	}

	return interfaces, nil
}

// TestConnectivity simulates connectivity test
func (m *MockDeviceManager) TestConnectivity(deviceID string) error {
	device, err := m.GetDevice(deviceID)
	if err != nil {
		return err
	}

	// Simulate occasional connectivity issues
	if rand.Intn(10) == 0 { // 10% chance of failure
		return fmt.Errorf("connection timeout to %s", device.Host)
	}

	// Update last seen
	device.LastSeen = time.Now().Format(time.RFC3339)
	device.Status = "online"

	m.logger.Info("Connectivity test passed", zap.String("device", device.Name))
	return nil
}

// ConfigureInterface simulates interface configuration
func (m *MockDeviceManager) ConfigureInterface(deviceID, interfaceName, description string, enabled bool) error {
	device, err := m.GetDevice(deviceID)
	if err != nil {
		return err
	}

	// Simulate configuration delay
	time.Sleep(100 * time.Millisecond)

	m.logger.Info("Interface configured",
		zap.String("device", device.Name),
		zap.String("interface", interfaceName),
		zap.String("description", description),
		zap.Bool("enabled", enabled))

	return nil
}

// DeleteDevice removes a device
func (m *MockDeviceManager) DeleteDevice(deviceID string) error {
	_, exists := m.devices[deviceID]
	if !exists {
		return fmt.Errorf("device not found")
	}

	delete(m.devices, deviceID)
	m.logger.Info("Device deleted", zap.String("device_id", deviceID))
	return nil
}

// GetDeviceCount returns the number of registered devices
func (m *MockDeviceManager) GetDeviceCount() int {
	return len(m.devices)
}

// SeedSampleDevices adds some sample devices for demo
func (m *MockDeviceManager) SeedSampleDevices() {
	m.RegisterDevice("Production Router", "cisco", "Cisco", "prod-router.company.com")
	m.RegisterDevice("Edge Switch", "juniper", "Juniper", "edge-sw01.company.com")
	m.RegisterDevice("Core Switch", "arista", "Arista", "core-sw01.company.com")
	m.RegisterDevice("Branch Router", "cisco", "Cisco", "branch-rtr.company.com")

	m.logger.Info("Sample devices seeded for demo", zap.Int("count", len(m.devices)))
}
