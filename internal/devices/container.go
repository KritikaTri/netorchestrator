package devices

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ContainerDevice represents a network device running in a container
type ContainerDevice struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"` // frr, alpine, ubuntu, vyos
	Vendor        string `json:"vendor"`
	ContainerName string `json:"container_name"`
	Port          int    `json:"port"`
	Status        string `json:"status"`
	LastSeen      string `json:"last_seen"`
	logger        *zap.Logger
}

// ContainerDeviceManager manages real network containers
type ContainerDeviceManager struct {
	devices map[string]*ContainerDevice
	logger  *zap.Logger
}

// ContainerDeviceInfo represents detailed device information from container
type ContainerDeviceInfo struct {
	Hostname    string   `json:"hostname"`
	OS          string   `json:"os"`
	Kernel      string   `json:"kernel"`
	Uptime      string   `json:"uptime"`
	CPUUsage    string   `json:"cpu_usage"`
	MemoryUsage string   `json:"memory_usage"`
	IPAddresses []string `json:"ip_addresses"`
	Routes      []string `json:"routes"`
	Processes   []string `json:"processes"`
}

// ContainerInterface represents a network interface from container
type ContainerInterface struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	IPAddress  string `json:"ip_address"`
	MACAddress string `json:"mac_address"`
	MTU        int    `json:"mtu"`
	RXBytes    string `json:"rx_bytes"`
	TXBytes    string `json:"tx_bytes"`
	Type       string `json:"type"` // physical, loopback, bridge, etc.
}

// NewContainerDeviceManager creates a new container device manager
func NewContainerDeviceManager(logger *zap.Logger) *ContainerDeviceManager {
	manager := &ContainerDeviceManager{
		devices: make(map[string]*ContainerDevice),
		logger:  logger,
	}

	// Auto-discover running containers
	manager.discoverContainers()

	return manager
}

// discoverContainers automatically discovers running network containers
func (m *ContainerDeviceManager) discoverContainers() {
	containers := []struct {
		name, deviceType, vendor string
		port                     int
	}{
		{"frr-router", "frr", "FRRouting", 23001},
		{"alpine-node", "alpine", "Alpine Linux", 24001},
		{"ubuntu-router", "ubuntu", "Ubuntu", 25001},
		{"vyos-node", "vyos", "VyOS", 26001},
	}

	for _, container := range containers {
		if m.isContainerRunning(container.name) {
			deviceID := fmt.Sprintf("container_%s", container.name)

			device := &ContainerDevice{
				ID:            deviceID,
				Name:          container.name,
				Type:          container.deviceType,
				Vendor:        container.vendor,
				ContainerName: container.name,
				Port:          container.port,
				Status:        "online",
				LastSeen:      time.Now().Format(time.RFC3339),
				logger:        m.logger,
			}

			m.devices[deviceID] = device
			m.logger.Info("Discovered container device",
				zap.String("name", container.name),
				zap.String("type", container.deviceType))
		}
	}
}

// isContainerRunning checks if a container is running
func (m *ContainerDeviceManager) isContainerRunning(containerName string) bool {
	cmd := exec.Command("podman", "ps", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), containerName)
}

// execInContainer executes a command in a container
func (m *ContainerDeviceManager) execInContainer(containerName, command string) (string, error) {
	cmd := exec.Command("podman", "exec", containerName, "sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ListDevices returns all discovered container devices
func (m *ContainerDeviceManager) ListDevices() []*ContainerDevice {
	devices := make([]*ContainerDevice, 0, len(m.devices))
	for _, device := range m.devices {
		// Update status
		if m.isContainerRunning(device.ContainerName) {
			device.Status = "online"
			device.LastSeen = time.Now().Format(time.RFC3339)
		} else {
			device.Status = "offline"
		}
		devices = append(devices, device)
	}
	return devices
}

// GetDevice retrieves a device by ID
func (m *ContainerDeviceManager) GetDevice(deviceID string) (*ContainerDevice, error) {
	device, exists := m.devices[deviceID]
	if !exists {
		return nil, fmt.Errorf("device not found")
	}
	return device, nil
}

// GetDeviceInfo retrieves detailed information from the actual container
func (m *ContainerDeviceManager) GetDeviceInfo(deviceID string) (*ContainerDeviceInfo, error) {
	device, err := m.GetDevice(deviceID)
	if err != nil {
		return nil, err
	}

	if !m.isContainerRunning(device.ContainerName) {
		return nil, fmt.Errorf("container is not running")
	}

	// Get hostname
	hostname, _ := m.execInContainer(device.ContainerName, "hostname")

	// Get OS info
	osInfo, _ := m.execInContainer(device.ContainerName, "cat /etc/os-release | grep PRETTY_NAME | cut -d'\"' -f2")

	// Get kernel version
	kernel, _ := m.execInContainer(device.ContainerName, "uname -r")

	// Get uptime
	uptime, _ := m.execInContainer(device.ContainerName, "uptime")

	// Get IP addresses
	ipOutput, _ := m.execInContainer(device.ContainerName, "ip addr show | grep 'inet ' | awk '{print $2}'")
	ipAddresses := strings.Split(strings.TrimSpace(ipOutput), "\n")

	// Get routes
	routeOutput, _ := m.execInContainer(device.ContainerName, "ip route")
	routes := strings.Split(strings.TrimSpace(routeOutput), "\n")

	// Get processes (top 5)
	processOutput, _ := m.execInContainer(device.ContainerName, "ps aux | head -6")
	processes := strings.Split(strings.TrimSpace(processOutput), "\n")

	return &ContainerDeviceInfo{
		Hostname:    hostname,
		OS:          osInfo,
		Kernel:      kernel,
		Uptime:      uptime,
		IPAddresses: ipAddresses,
		Routes:      routes,
		Processes:   processes,
		CPUUsage:    "N/A (Container)",
		MemoryUsage: "N/A (Container)",
	}, nil
}

// GetInterfaces retrieves network interfaces from the actual container
func (m *ContainerDeviceManager) GetInterfaces(deviceID string) ([]*ContainerInterface, error) {
	device, err := m.GetDevice(deviceID)
	if err != nil {
		return nil, err
	}

	if !m.isContainerRunning(device.ContainerName) {
		return nil, fmt.Errorf("container is not running")
	}

	// Get interface information
	output, err := m.execInContainer(device.ContainerName, "ip addr show")
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	interfaces := []*ContainerInterface{}

	// Parse ip addr output
	lines := strings.Split(output, "\n")
	var currentInterface *ContainerInterface

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Interface line: "2: eth0@if9: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 65520 ..."
		if matched, _ := regexp.MatchString(`^\d+: \w+`, line); matched {
			if currentInterface != nil {
				interfaces = append(interfaces, currentInterface)
			}

			parts := strings.Fields(line)
			if len(parts) >= 2 {
				interfaceName := strings.TrimSuffix(parts[1], ":")
				// Remove @if suffix if present
				if idx := strings.Index(interfaceName, "@"); idx > 0 {
					interfaceName = interfaceName[:idx]
				}

				status := "down"
				if strings.Contains(line, "UP") {
					status = "up"
				}

				mtu := 1500
				if strings.Contains(line, "mtu") {
					for i, part := range parts {
						if part == "mtu" && i+1 < len(parts) {
							if m, err := strconv.Atoi(parts[i+1]); err == nil {
								mtu = m
							}
							break
						}
					}
				}

				currentInterface = &ContainerInterface{
					Name:   interfaceName,
					Status: status,
					MTU:    mtu,
					Type:   "ethernet",
				}
			}
		}

		// MAC address line: "link/ether 92:bf:9a:2a:f9:e4 brd ff:ff:ff:ff:ff:ff"
		if currentInterface != nil && strings.Contains(line, "link/ether") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentInterface.MACAddress = parts[1]
			}
		}

		// IP address line: "inet 10.89.0.3/24 brd 10.89.0.255 scope global eth0"
		if currentInterface != nil && strings.Contains(line, "inet ") && !strings.Contains(line, "inet6") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentInterface.IPAddress = parts[1]
			}
		}
	}

	// Add the last interface
	if currentInterface != nil {
		interfaces = append(interfaces, currentInterface)
	}

	return interfaces, nil
}

// TestConnectivity tests connectivity to the container
func (m *ContainerDeviceManager) TestConnectivity(deviceID string) error {
	device, err := m.GetDevice(deviceID)
	if err != nil {
		return err
	}

	if !m.isContainerRunning(device.ContainerName) {
		return fmt.Errorf("container %s is not running", device.ContainerName)
	}

	// Test by executing a simple command
	_, err = m.execInContainer(device.ContainerName, "echo 'connectivity test'")
	if err != nil {
		return fmt.Errorf("connectivity test failed: %w", err)
	}

	device.Status = "online"
	device.LastSeen = time.Now().Format(time.RFC3339)

	m.logger.Info("Container connectivity test passed", zap.String("container", device.ContainerName))
	return nil
}

// ExecuteCommand executes a network command on the container
func (m *ContainerDeviceManager) ExecuteCommand(deviceID, command string) (string, error) {
	device, err := m.GetDevice(deviceID)
	if err != nil {
		return "", err
	}

	if !m.isContainerRunning(device.ContainerName) {
		return "", fmt.Errorf("container is not running")
	}

	output, err := m.execInContainer(device.ContainerName, command)
	if err != nil {
		return "", fmt.Errorf("command execution failed: %w", err)
	}

	m.logger.Info("Command executed on container",
		zap.String("container", device.ContainerName),
		zap.String("command", command))

	return output, nil
}

// GetStats returns statistics about container devices
func (m *ContainerDeviceManager) GetStats() map[string]interface{} {
	devices := m.ListDevices()

	stats := map[string]interface{}{
		"total_devices":  len(devices),
		"online_devices": 0,
		"device_types":   make(map[string]int),
		"vendors":        make(map[string]int),
	}

	for _, device := range devices {
		if device.Status == "online" {
			stats["online_devices"] = stats["online_devices"].(int) + 1
		}
		stats["device_types"].(map[string]int)[device.Type]++
		stats["vendors"].(map[string]int)[device.Vendor]++
	}

	return stats
}
