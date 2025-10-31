package nlp

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// NLPService handles natural language processing for network provisioning
type NLPService struct {
	logger *zap.Logger
}

// NewNLPService creates a new NLP service
func NewNLPService(logger *zap.Logger) *NLPService {
	return &NLPService{
		logger: logger,
	}
}

// NetworkSpec represents a parsed network topology specification
type NetworkSpec struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Nodes       []NodeSpec             `json:"nodes"`
	Topology    string                 `json:"topology"` // "star", "mesh", "ring", "bus"
	Config      map[string]interface{} `json:"config"`
}

// NodeSpec represents a parsed node specification
type NodeSpec struct {
	Name      string `json:"name"`
	Type      string `json:"type"` // "host", "router", "switch"
	IPAddress string `json:"ip_address,omitempty"`
}

// ParseNetworkRequest parses natural language and returns a network specification
func (s *NLPService) ParseNetworkRequest(ctx context.Context, text string) (*NetworkSpec, error) {
	text = strings.ToLower(strings.TrimSpace(text))
	s.logger.Info("Parsing natural language request", zap.String("text", text))

	// Rule-based parsing
	spec := &NetworkSpec{
		Nodes:    []NodeSpec{},
		Config:   make(map[string]interface{}),
		Topology: "custom",
	}

	// Extract topology type
	if strings.Contains(text, "star") || strings.Contains(text, "hub") {
		spec.Topology = "star"
	} else if strings.Contains(text, "mesh") {
		spec.Topology = "mesh"
	} else if strings.Contains(text, "ring") {
		spec.Topology = "ring"
	} else if strings.Contains(text, "bus") || strings.Contains(text, "linear") {
		spec.Topology = "bus"
	}

	// Extract number of nodes
	nodeCount := s.extractNumber(text)
	if nodeCount == 0 {
		// Try to extract from phrases like "5 nodes", "three routers"
		nodeCount = s.extractNodeCount(text)
	}
	if nodeCount == 0 {
		nodeCount = 3 // default
	}

	// Extract node types
	nodeTypes := s.extractNodeTypes(text)
	if len(nodeTypes) == 0 {
		nodeTypes = []string{"host"} // default
	}

	// Generate network name
	spec.Name = s.extractName(text)
	if spec.Name == "" {
		spec.Name = fmt.Sprintf("Network from '%s'", truncate(text, 30))
	}

	// Generate description
	spec.Description = fmt.Sprintf("Network created from natural language: %s", text)

	// Generate nodes
	for i := 0; i < nodeCount; i++ {
		nodeType := nodeTypes[i%len(nodeTypes)]
		node := NodeSpec{
			Name: fmt.Sprintf("Node %d", i+1),
			Type: nodeType,
		}

		// Try to extract IP addresses if mentioned
		ip := s.extractIP(text, i)
		if ip != "" {
			node.IPAddress = ip
		}

		spec.Nodes = append(spec.Nodes, node)
	}

	// Extract subnet if mentioned
	subnet := s.extractSubnet(text)
	if subnet != "" {
		spec.Config["subnet"] = subnet
	}

	// Extract description
	if strings.Contains(text, "for") || strings.Contains(text, "to") {
		parts := strings.Split(text, "for")
		if len(parts) > 1 {
			spec.Description = strings.TrimSpace(parts[1])
		}
	}

	s.logger.Info("Parsed network specification",
		zap.String("topology", spec.Topology),
		zap.Int("node_count", len(spec.Nodes)),
	)

	return spec, nil
}

// extractNumber extracts a number from text
func (s *NLPService) extractNumber(text string) int {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(text, -1)
	if len(matches) > 0 {
		var num int
		fmt.Sscanf(matches[0], "%d", &num)
		return num
	}
	return 0
}

// extractNodeCount extracts node count from text (handles both numeric and word forms)
func (s *NLPService) extractNodeCount(text string) int {
	// Check for word forms
	wordNumbers := map[string]int{
		"one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
		"six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
	}

	for word, num := range wordNumbers {
		if strings.Contains(text, word) {
			return num
		}
	}

	// Check for phrases like "5 nodes", "3 routers"
	re := regexp.MustCompile(`(\d+)\s+(nodes?|routers?|switches?|hosts?)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		var num int
		fmt.Sscanf(matches[1], "%d", &num)
		return num
	}

	return 0
}

// extractNodeTypes extracts node types from text
func (s *NLPService) extractNodeTypes(text string) []string {
	types := []string{}

	if strings.Contains(text, "router") {
		types = append(types, "router")
	}
	if strings.Contains(text, "switch") {
		types = append(types, "switch")
	}
	if strings.Contains(text, "host") || strings.Contains(text, "server") || strings.Contains(text, "node") {
		types = append(types, "host")
	}

	return types
}

// extractName extracts a name from text
func (s *NLPService) extractName(text string) string {
	// Look for "called", "named", or "name"
	re := regexp.MustCompile(`(?:called|named|name)\s+(\w+)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}

	// Look for quotes
	re = regexp.MustCompile(`["']([^"']+)["']`)
	matches = re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// extractIP extracts IP address from text (if mentioned)
func (s *NLPService) extractIP(text string, index int) string {
	// Look for IP patterns like 192.168.1.10
	re := regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)
	matches := re.FindAllString(text, -1)
	if len(matches) > index {
		return matches[index]
	}
	return ""
}

// extractSubnet extracts subnet from text
func (s *NLPService) extractSubnet(text string) string {
	// Look for subnet patterns like 192.168.1.0/24
	re := regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2})\b`)
	matches := re.FindString(text)
	return matches
}

// truncate truncates a string to max length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// ValidateSpec validates a parsed network specification
func (s *NLPService) ValidateSpec(ctx context.Context, spec *NetworkSpec) error {
	if spec.Name == "" {
		return fmt.Errorf("network name is required")
	}

	if len(spec.Nodes) == 0 {
		return fmt.Errorf("at least one node is required")
	}

	// Validate topology constraints
	if spec.Topology == "star" && len(spec.Nodes) < 2 {
		return fmt.Errorf("star topology requires at least 2 nodes")
	}

	if spec.Topology == "mesh" && len(spec.Nodes) < 3 {
		return fmt.Errorf("mesh topology requires at least 3 nodes")
	}

	if spec.Topology == "ring" && len(spec.Nodes) < 3 {
		return fmt.Errorf("ring topology requires at least 3 nodes")
	}

	return nil
}
