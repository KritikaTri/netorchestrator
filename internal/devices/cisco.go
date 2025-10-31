package devices

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// CiscoDevice represents a Cisco network device
type CiscoDevice struct {
	Host         string `json:"host"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	RESTCONFPort int    `json:"restconf_port"`
	NETCONFPort  int    `json:"netconf_port"`
	SSHPort      int    `json:"ssh_port"`
	logger       *zap.Logger
	httpClient   *http.Client
}

// DeviceInfo represents device information
type DeviceInfo struct {
	Hostname     string `json:"hostname"`
	Model        string `json:"model"`
	Version      string `json:"version"`
	SerialNumber string `json:"serial_number"`
	Uptime       string `json:"uptime"`
}

// InterfaceConfig represents interface configuration
type InterfaceConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IPAddress   string `json:"ip_address"`
	SubnetMask  string `json:"subnet_mask"`
	Enabled     bool   `json:"enabled"`
}

// NewCiscoDevice creates a new Cisco device instance
func NewCiscoDevice(host, username, password string, logger *zap.Logger) *CiscoDevice {
	// Create HTTP client with SSL verification disabled for lab environments
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &CiscoDevice{
		Host:         host,
		Username:     username,
		Password:     password,
		RESTCONFPort: 443,
		NETCONFPort:  830,
		SSHPort:      22,
		logger:       logger,
		httpClient:   &http.Client{Transport: tr, Timeout: 30 * time.Second},
	}
}

// GetDeviceInfo retrieves basic device information
func (d *CiscoDevice) GetDeviceInfo() (*DeviceInfo, error) {
	url := fmt.Sprintf("https://%s:%d/restconf/data/Cisco-IOS-XE-native:native/hostname",
		d.Host, d.RESTCONFPort)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(d.Username, d.Password)
	req.Header.Set("Accept", "application/yang-data+json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed: %s, body: %s", resp.Status, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	d.logger.Info("Retrieved device info", zap.Any("response", result))

	// Extract hostname from response
	hostname := "Unknown"
	if hostnameData, ok := result["Cisco-IOS-XE-native:hostname"]; ok {
		if hn, ok := hostnameData.(string); ok {
			hostname = hn
		}
	}

	return &DeviceInfo{
		Hostname: hostname,
		Model:    "Catalyst 8000V",
		Version:  "Unknown", // Would need additional API call
	}, nil
}

// GetInterfaces retrieves all interfaces
func (d *CiscoDevice) GetInterfaces() ([]InterfaceConfig, error) {
	url := fmt.Sprintf("https://%s:%d/restconf/data/ietf-interfaces:interfaces",
		d.Host, d.RESTCONFPort)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(d.Username, d.Password)
	req.Header.Set("Accept", "application/yang-data+json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed: %s, body: %s", resp.Status, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	d.logger.Info("Retrieved interfaces", zap.Any("response", result))

	interfaces := []InterfaceConfig{}

	// Parse interfaces from response
	if interfacesData, ok := result["ietf-interfaces:interfaces"]; ok {
		if interfaceMap, ok := interfacesData.(map[string]interface{}); ok {
			if interfaceList, ok := interfaceMap["interface"].([]interface{}); ok {
				for _, iface := range interfaceList {
					if ifaceMap, ok := iface.(map[string]interface{}); ok {
						config := InterfaceConfig{
							Name:    getString(ifaceMap, "name"),
							Enabled: getBool(ifaceMap, "enabled"),
						}
						interfaces = append(interfaces, config)
					}
				}
			}
		}
	}

	return interfaces, nil
}

// ConfigureInterface configures a network interface
func (d *CiscoDevice) ConfigureInterface(config InterfaceConfig) error {
	url := fmt.Sprintf("https://%s:%d/restconf/data/ietf-interfaces:interfaces/interface=%s",
		d.Host, d.RESTCONFPort, config.Name)

	// Build configuration payload
	payload := map[string]interface{}{
		"ietf-interfaces:interface": map[string]interface{}{
			"name":    config.Name,
			"enabled": config.Enabled,
		},
	}

	if config.Description != "" {
		payload["ietf-interfaces:interface"].(map[string]interface{})["description"] = config.Description
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(d.Username, d.Password)
	req.Header.Set("Content-Type", "application/yang-data+json")
	req.Header.Set("Accept", "application/yang-data+json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("configuration failed: %s, body: %s", resp.Status, string(body))
	}

	d.logger.Info("Interface configured successfully", zap.String("interface", config.Name))
	return nil
}

// TestConnectivity tests basic connectivity to the device
func (d *CiscoDevice) TestConnectivity() error {
	_, err := d.GetDeviceInfo()
	if err != nil {
		return fmt.Errorf("connectivity test failed: %w", err)
	}

	d.logger.Info("Device connectivity test passed", zap.String("host", d.Host))
	return nil
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}
