package nlp

import (
	"fmt"
	"strings"

	"netorchestrator/internal/models"
)

// SupportedTypes defines all the supported network component types
type SupportedTypes struct {
	NodeTypes       []string
	Topologies      []string
	PolicyTypes     []string
	DeviceVendors   []string
	FirewallActions []string
}

// GetSupportedTypes returns all supported network component types
func GetSupportedTypes() *SupportedTypes {
	return &SupportedTypes{
		NodeTypes: []string{
			string(models.NodeTypeRouter),
			string(models.NodeTypeSwitch),
			string(models.NodeTypeHost),
			string(models.NodeTypeFirewall),
			string(models.NodeTypeLoadBalancer),
			string(models.NodeTypeGateway),
		},
		Topologies: []string{
			"star",
			"mesh",
			"tree",
			"custom",
		},
		PolicyTypes: []string{
			string(models.PolicyTypeTraffic),
			string(models.PolicyTypeSecurity),
			string(models.PolicyTypeFirewall),
			string(models.PolicyTypeQoS),
			string(models.PolicyTypeRouting),
			string(models.PolicyTypeAccess),
		},
		DeviceVendors: []string{
			string(models.DeviceTypeCisco),
			string(models.DeviceTypeJuniper),
			string(models.DeviceTypeArista),
		},
		FirewallActions: []string{
			"allow",
			"deny",
			"drop",
		},
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// NetworkValidator validates network configurations
type NetworkValidator struct {
	supportedTypes *SupportedTypes
}

// NewNetworkValidator creates a new network validator
func NewNetworkValidator() *NetworkValidator {
	return &NetworkValidator{
		supportedTypes: GetSupportedTypes(),
	}
}

// ValidateNodeType checks if a node type is supported
func (v *NetworkValidator) ValidateNodeType(nodeType string) *ValidationError {
	nodeType = strings.ToLower(strings.TrimSpace(nodeType))

	for _, supportedType := range v.supportedTypes.NodeTypes {
		if strings.ToLower(supportedType) == nodeType {
			return nil
		}
	}

	return &ValidationError{
		Field:   "node_type",
		Value:   nodeType,
		Message: fmt.Sprintf("Unsupported node type '%s'. Supported types: %s", nodeType, strings.Join(v.supportedTypes.NodeTypes, ", ")),
	}
}

// ValidateTopology checks if a topology is supported
func (v *NetworkValidator) ValidateTopology(topology string) *ValidationError {
	topology = strings.ToLower(strings.TrimSpace(topology))

	for _, supportedTopology := range v.supportedTypes.Topologies {
		if strings.ToLower(supportedTopology) == topology {
			return nil
		}
	}

	return &ValidationError{
		Field:   "topology",
		Value:   topology,
		Message: fmt.Sprintf("Unsupported topology '%s'. Supported topologies: %s", topology, strings.Join(v.supportedTypes.Topologies, ", ")),
	}
}

// ValidatePolicyType checks if a policy type is supported
func (v *NetworkValidator) ValidatePolicyType(policyType string) *ValidationError {
	policyType = strings.ToLower(strings.TrimSpace(policyType))

	for _, supportedType := range v.supportedTypes.PolicyTypes {
		if strings.ToLower(supportedType) == policyType {
			return nil
		}
	}

	return &ValidationError{
		Field:   "policy_type",
		Value:   policyType,
		Message: fmt.Sprintf("Unsupported policy type '%s'. Supported types: %s", policyType, strings.Join(v.supportedTypes.PolicyTypes, ", ")),
	}
}

// ValidateFirewallAction checks if a firewall action is supported
func (v *NetworkValidator) ValidateFirewallAction(action string) *ValidationError {
	action = strings.ToLower(strings.TrimSpace(action))

	for _, supportedAction := range v.supportedTypes.FirewallActions {
		if strings.ToLower(supportedAction) == action {
			return nil
		}
	}

	return &ValidationError{
		Field:   "firewall_action",
		Value:   action,
		Message: fmt.Sprintf("Unsupported firewall action '%s'. Supported actions: %s", action, strings.Join(v.supportedTypes.FirewallActions, ", ")),
	}
}

// ValidateTopologyConstraints checks topology-specific constraints
func (v *NetworkValidator) ValidateTopologyConstraints(topology string, nodeCount int) []*ValidationError {
	var errors []*ValidationError
	topology = strings.ToLower(strings.TrimSpace(topology))

	switch topology {
	case "star":
		if nodeCount < 2 {
			errors = append(errors, &ValidationError{
				Field:   "node_count",
				Value:   fmt.Sprintf("%d", nodeCount),
				Message: "Star topology requires at least 2 nodes (1 hub + 1 spoke)",
			})
		}
	case "mesh":
		if nodeCount < 3 {
			errors = append(errors, &ValidationError{
				Field:   "node_count",
				Value:   fmt.Sprintf("%d", nodeCount),
				Message: "Mesh topology requires at least 3 nodes for meaningful connectivity",
			})
		}
	case "tree":
		if nodeCount < 3 {
			errors = append(errors, &ValidationError{
				Field:   "node_count",
				Value:   fmt.Sprintf("%d", nodeCount),
				Message: "Tree topology requires at least 3 nodes (1 root + 2 children)",
			})
		}
	case "ring":
		if nodeCount < 3 {
			errors = append(errors, &ValidationError{
				Field:   "node_count",
				Value:   fmt.Sprintf("%d", nodeCount),
				Message: "Ring topology requires at least 3 nodes to form a circle",
			})
		}
	}

	// General constraints
	if nodeCount > 50 {
		errors = append(errors, &ValidationError{
			Field:   "node_count",
			Value:   fmt.Sprintf("%d", nodeCount),
			Message: "Maximum supported nodes per network is 50 for performance reasons",
		})
	}

	return errors
}

// ValidateNetworkSpec validates an entire network specification
func (v *NetworkValidator) ValidateNetworkSpec(spec *NetworkSpec) []*ValidationError {
	var errors []*ValidationError

	// Validate topology
	if err := v.ValidateTopology(spec.Topology); err != nil {
		errors = append(errors, err)
	}

	// Validate topology constraints
	topologyErrors := v.ValidateTopologyConstraints(spec.Topology, len(spec.Nodes))
	errors = append(errors, topologyErrors...)

	// Validate node types
	nodeTypeCounts := make(map[string]int)
	for _, node := range spec.Nodes {
		if err := v.ValidateNodeType(node.Type); err != nil {
			errors = append(errors, &ValidationError{
				Field:   "node_type",
				Value:   fmt.Sprintf("%s (%s)", node.Name, node.Type),
				Message: err.Message,
			})
		}
		nodeTypeCounts[strings.ToLower(node.Type)]++
	}

	// Validate node type combinations for specific topologies
	if spec.Topology == "star" {
		// Star topology should have at least one router or switch as hub
		hasHub := nodeTypeCounts["router"] > 0 || nodeTypeCounts["switch"] > 0
		if !hasHub {
			errors = append(errors, &ValidationError{
				Field:   "topology_nodes",
				Value:   "star",
				Message: "Star topology requires at least one router or switch as a hub",
			})
		}
	}

	// Validate policies if present
	for _, policy := range spec.Policies {
		if err := v.ValidatePolicyType(policy.Type); err != nil {
			errors = append(errors, &ValidationError{
				Field:   "policy_type",
				Value:   fmt.Sprintf("%s (%s)", policy.Name, policy.Type),
				Message: err.Message,
			})
		}

		// Validate firewall rules if it's a firewall policy
		if strings.ToLower(policy.Type) == "firewall" {
			for _, rule := range policy.Rules {
				if action, exists := rule["action"]; exists {
					if err := v.ValidateFirewallAction(fmt.Sprintf("%v", action)); err != nil {
						errors = append(errors, &ValidationError{
							Field:   "firewall_action",
							Value:   fmt.Sprintf("%s.%v", policy.Name, action),
							Message: err.Message,
						})
					}
				}
			}
		}
	}

	return errors
}

// SuggestCorrections suggests corrections for validation errors
func (v *NetworkValidator) SuggestCorrections(errors []*ValidationError) map[string][]string {
	suggestions := make(map[string][]string)

	for _, err := range errors {
		switch err.Field {
		case "node_type":
			suggestions["node_types"] = v.supportedTypes.NodeTypes
		case "topology":
			suggestions["topologies"] = v.supportedTypes.Topologies
		case "policy_type":
			suggestions["policy_types"] = v.supportedTypes.PolicyTypes
		case "firewall_action":
			suggestions["firewall_actions"] = v.supportedTypes.FirewallActions
		}
	}

	return suggestions
}

// GetSupportedTypesForLLM returns a formatted string of supported types for LLM prompts
func (v *NetworkValidator) GetSupportedTypesForLLM() string {
	return fmt.Sprintf(`SUPPORTED NETWORK COMPONENTS:
- Node Types: %s
- Topologies: %s  
- Policy Types: %s
- Firewall Actions: %s
- Device Vendors: %s

TOPOLOGY CONSTRAINTS:
- Star: minimum 2 nodes (1 hub + 1 spoke), hub must be router/switch
- Mesh: minimum 3 nodes for meaningful connectivity
- Tree: minimum 3 nodes (1 root + 2 children)
- Ring: minimum 3 nodes to form a circle
- Maximum: 50 nodes per network

VALIDATION RULES:
- Only use supported node types, topologies, and policy types
- Respect topology constraints
- Firewall policies must use supported actions
- Network names must be unique and descriptive`,
		strings.Join(v.supportedTypes.NodeTypes, ", "),
		strings.Join(v.supportedTypes.Topologies, ", "),
		strings.Join(v.supportedTypes.PolicyTypes, ", "),
		strings.Join(v.supportedTypes.FirewallActions, ", "),
		strings.Join(v.supportedTypes.DeviceVendors, ", "),
	)
}
