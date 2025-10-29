package services

import (
	"context"
	"fmt"
	"net"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"netorchestrator/internal/models"
)

// ValidationService handles validation operations
type ValidationService struct {
	db     *gorm.DB
	cache  *redis.Client
	logger *zap.Logger
}

// NewValidationService creates a new validation service
func NewValidationService(db *gorm.DB, cache *redis.Client, logger *zap.Logger) *ValidationService {
	return &ValidationService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// ValidateNetwork validates a network configuration
func (s *ValidationService) ValidateNetwork(ctx context.Context, network *models.Network) error {
	// Validate network name
	if network.Name == "" {
		return fmt.Errorf("network name cannot be empty")
	}

	// Validate network configuration
	if err := s.validateNetworkConfig(&network.Config); err != nil {
		return fmt.Errorf("invalid network configuration: %w", err)
	}

	// Check if network name is unique
	var existingNetwork models.Network
	if err := s.db.WithContext(ctx).Where("name = ? AND id != ?", network.Name, network.ID).First(&existingNetwork).Error; err == nil {
		return fmt.Errorf("network name already exists")
	}

	return nil
}

// ValidateNode validates a node configuration
func (s *ValidationService) ValidateNode(ctx context.Context, node *models.Node) error {
	// Validate node name
	if node.Name == "" {
		return fmt.Errorf("node name cannot be empty")
	}

	// Validate node type
	validTypes := []models.NodeType{
		models.NodeTypeRouter,
		models.NodeTypeSwitch,
		models.NodeTypeHost,
		models.NodeTypeFirewall,
		models.NodeTypeLoadBalancer,
		models.NodeTypeGateway,
	}

	isValidType := false
	for _, validType := range validTypes {
		if node.Type == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		return fmt.Errorf("invalid node type: %s", node.Type)
	}

	// Validate IP address if provided
	if node.IPAddress != "" {
		if err := s.validateIPAddress(node.IPAddress); err != nil {
			return fmt.Errorf("invalid IP address: %w", err)
		}
	}

	// Validate MAC address if provided
	if node.MACAddress != "" {
		if err := s.validateMACAddress(node.MACAddress); err != nil {
			return fmt.Errorf("invalid MAC address: %w", err)
		}
	}

	// Validate node configuration
	if err := s.validateNodeConfig(&node.Config); err != nil {
		return fmt.Errorf("invalid node configuration: %w", err)
	}

	// Check if node name is unique within the network
	var existingNode models.Node
	if err := s.db.WithContext(ctx).Where("name = ? AND network_id = ? AND id != ?", node.Name, node.NetworkID, node.ID).First(&existingNode).Error; err == nil {
		return fmt.Errorf("node name already exists in this network")
	}

	return nil
}

// ValidateLink validates a link configuration
func (s *ValidationService) ValidateLink(ctx context.Context, link *models.Link) error {
	// Validate that source and target nodes are different
	if link.SourceNodeID == link.TargetNodeID {
		return fmt.Errorf("source and target nodes cannot be the same")
	}

	// Validate that both nodes exist and belong to the same network
	var sourceNode, targetNode models.Node
	if err := s.db.WithContext(ctx).First(&sourceNode, "id = ?", link.SourceNodeID).Error; err != nil {
		return fmt.Errorf("source node not found")
	}
	if err := s.db.WithContext(ctx).First(&targetNode, "id = ?", link.TargetNodeID).Error; err != nil {
		return fmt.Errorf("target node not found")
	}

	if sourceNode.NetworkID != targetNode.NetworkID {
		return fmt.Errorf("source and target nodes must belong to the same network")
	}

	if sourceNode.NetworkID != link.NetworkID {
		return fmt.Errorf("link network ID does not match source node network ID")
	}

	// Check if link already exists between these nodes
	var existingLink models.Link
	if err := s.db.WithContext(ctx).Where("(source_node_id = ? AND target_node_id = ?) OR (source_node_id = ? AND target_node_id = ?)", 
		link.SourceNodeID, link.TargetNodeID, link.TargetNodeID, link.SourceNodeID).First(&existingLink).Error; err == nil {
		if existingLink.ID != link.ID {
			return fmt.Errorf("link already exists between these nodes")
		}
	}

	// Validate link configuration
	if err := s.validateLinkConfig(&link.Config); err != nil {
		return fmt.Errorf("invalid link configuration: %w", err)
	}

	return nil
}

// ValidatePolicy validates a policy configuration
func (s *ValidationService) ValidatePolicy(ctx context.Context, policy *models.Policy) error {
	// Validate policy name
	if policy.Name == "" {
		return fmt.Errorf("policy name cannot be empty")
	}

	// Validate policy type
	validTypes := []models.PolicyType{
		models.PolicyTypeTraffic,
		models.PolicyTypeSecurity,
		models.PolicyTypeQoS,
		models.PolicyTypeRouting,
		models.PolicyTypeAccess,
	}

	isValidType := false
	for _, validType := range validTypes {
		if policy.Type == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		return fmt.Errorf("invalid policy type: %s", policy.Type)
	}

	// Validate policy configuration
	if err := s.validatePolicyConfig(&policy.Config); err != nil {
		return fmt.Errorf("invalid policy configuration: %w", err)
	}

	// Check if policy name is unique within the network
	var existingPolicy models.Policy
	if err := s.db.WithContext(ctx).Where("name = ? AND network_id = ? AND id != ?", policy.Name, policy.NetworkID, policy.ID).First(&existingPolicy).Error; err == nil {
		return fmt.Errorf("policy name already exists in this network")
	}

	return nil
}

// validateNetworkConfig validates network configuration
func (s *ValidationService) validateNetworkConfig(config *models.NetworkConfig) error {
	// Validate subnet if provided
	if config.Subnet != "" {
		if err := s.validateCIDR(config.Subnet); err != nil {
			return fmt.Errorf("invalid subnet: %w", err)
		}
	}

	// Validate gateway if provided
	if config.Gateway != "" {
		if err := s.validateIPAddress(config.Gateway); err != nil {
			return fmt.Errorf("invalid gateway: %w", err)
		}
	}

	// Validate MTU
	if config.MTU < 0 || config.MTU > 9000 {
		return fmt.Errorf("MTU must be between 0 and 9000")
	}

	// Validate DNS servers if provided
	for _, dns := range config.DNS {
		if err := s.validateIPAddress(dns); err != nil {
			return fmt.Errorf("invalid DNS server: %w", err)
		}
	}

	return nil
}

// validateNodeConfig validates node configuration
func (s *ValidationService) validateNodeConfig(config *models.NodeConfig) error {
	// Validate CPU
	if config.CPU < 0 {
		return fmt.Errorf("CPU cannot be negative")
	}

	// Validate memory
	if config.Memory < 0 {
		return fmt.Errorf("memory cannot be negative")
	}

	// Validate storage
	if config.Storage < 0 {
		return fmt.Errorf("storage cannot be negative")
	}

	// Validate ports
	for _, port := range config.Ports {
		if port.Number < 1 || port.Number > 65535 {
			return fmt.Errorf("invalid port number: %d", port.Number)
		}
		if port.Speed < 0 {
			return fmt.Errorf("port speed cannot be negative")
		}
	}

	// Validate services
	for _, service := range config.Services {
		if service.Port < 1 || service.Port > 65535 {
			return fmt.Errorf("invalid service port: %d", service.Port)
		}
	}

	return nil
}

// validateLinkConfig validates link configuration
func (s *ValidationService) validateLinkConfig(config *models.LinkConfig) error {
	// Validate bandwidth
	if config.Bandwidth < 0 {
		return fmt.Errorf("bandwidth cannot be negative")
	}

	// Validate latency
	if config.Latency < 0 {
		return fmt.Errorf("latency cannot be negative")
	}

	// Validate jitter
	if config.Jitter < 0 {
		return fmt.Errorf("jitter cannot be negative")
	}

	// Validate packet loss
	if config.PacketLoss < 0 || config.PacketLoss > 100 {
		return fmt.Errorf("packet loss must be between 0 and 100")
	}

	return nil
}

// validatePolicyConfig validates policy configuration
func (s *ValidationService) validatePolicyConfig(config *models.PolicyConfig) error {
	// Validate priority
	if config.Priority < 0 {
		return fmt.Errorf("priority cannot be negative")
	}

	// Validate rules
	for _, rule := range config.Rules {
		if rule.Priority < 0 {
			return fmt.Errorf("rule priority cannot be negative")
		}
	}

	return nil
}

// validateIPAddress validates an IP address
func (s *ValidationService) validateIPAddress(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address format")
	}
	return nil
}

// validateMACAddress validates a MAC address
func (s *ValidationService) validateMACAddress(mac string) error {
	_, err := net.ParseMAC(mac)
	return err
}

// validateCIDR validates a CIDR notation
func (s *ValidationService) validateCIDR(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	return err
}
