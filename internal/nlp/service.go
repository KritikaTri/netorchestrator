package nlp

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"netorchestrator/internal/ai"

	"go.uber.org/zap"
)

// NLPService handles natural language processing for network provisioning
type NLPService struct {
	logger    *zap.Logger
	validator *NetworkValidator
	aiEngine  *ai.AIEngine
}

// NewNLPService creates a new NLP service
func NewNLPService(logger *zap.Logger, aiEngine *ai.AIEngine) *NLPService {
	return &NLPService{
		logger:    logger,
		validator: NewNetworkValidator(),
		aiEngine:  aiEngine,
	}
}

// NetworkSpec represents a parsed network topology specification
type NetworkSpec struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Nodes            []NodeSpec             `json:"nodes"`
	Topology         string                 `json:"topology"` // "star", "mesh", "tree", "custom"
	Policies         []PolicySpec           `json:"policies"`
	Config           map[string]interface{} `json:"config"`
	Valid            bool                   `json:"valid"`
	Confidence       float64                `json:"confidence"`
	ValidationErrors []string               `json:"validation_errors"`
}

// PolicySpec represents a parsed policy specification
type PolicySpec struct {
	Name  string                   `json:"name"`
	Type  string                   `json:"type"`
	Rules []map[string]interface{} `json:"rules"`
}

// NodeSpec represents a parsed node specification
type NodeSpec struct {
	Name      string `json:"name"`
	Type      string `json:"type"` // "host", "router", "switch"
	IPAddress string `json:"ip_address,omitempty"`
}

// ParseNetworkRequest parses natural language and returns a network specification
func (s *NLPService) ParseNetworkRequest(ctx context.Context, text string) (*NetworkSpec, error) {
	text = strings.TrimSpace(text)
	s.logger.Info("Parsing natural language request", zap.String("text", text))

	// Try AI-enhanced parsing first if available, fallback to rule-based
	var spec *NetworkSpec
	var err error

	if s.aiEngine != nil {
		spec, err = s.parseWithLLM(ctx, text)
		if err != nil {
			s.logger.Warn("LLM parsing failed, falling back to rule-based", zap.Error(err))
			spec = s.parseWithRules(text)
		}
	} else {
		spec = s.parseWithRules(text)
	}

	// Validate the parsed specification
	validationErrors := s.validator.ValidateNetworkSpec(spec)
	spec.Valid = len(validationErrors) == 0
	spec.ValidationErrors = make([]string, len(validationErrors))

	for i, verr := range validationErrors {
		spec.ValidationErrors[i] = verr.Message
	}

	// Log validation results
	if spec.Valid {
		s.logger.Info("Successfully parsed and validated network specification",
			zap.String("topology", spec.Topology),
			zap.Int("node_count", len(spec.Nodes)),
			zap.Float64("confidence", spec.Confidence))
	} else {
		s.logger.Warn("Network specification validation failed",
			zap.Strings("errors", spec.ValidationErrors))
	}

	return spec, nil
}

// parseWithLLM uses OpenAI to parse natural language requests
func (s *NLPService) parseWithLLM(ctx context.Context, text string) (*NetworkSpec, error) {
	// Build validation prompt
	systemPrompt := ai.BuildNLPValidationPrompt()
	userPrompt := fmt.Sprintf(`Parse this network request and return valid JSON:

"%s"

%s

Respond only with valid JSON matching the schema.`, text, s.validator.GetSupportedTypesForLLM())

	// Use AI engine to process
	aiReq := &ai.AIRequest{
		ChainID: "nlp_parsing",
		Input:   userPrompt,
		Context: map[string]interface{}{
			"original_text": text,
			"system_prompt": systemPrompt,
		},
	}

	// Create a simple LLM chain for parsing
	if _, exists := s.aiEngine.GetChain("nlp_parsing"); !exists {
		chain := &ai.Chain{
			ID:   "nlp_parsing",
			Name: "NLP Network Parsing",
			Steps: []ai.ChainStep{
				{
					ID:       "llm_parse",
					Type:     "llm",
					Template: systemPrompt + "\n\nUser request: {input}",
				},
			},
		}
		s.aiEngine.CreateChain(chain)
	}

	resp, err := s.aiEngine.ProcessRequest(ctx, aiReq)
	if err != nil {
		return nil, fmt.Errorf("AI processing failed: %w", err)
	}

	// Parse JSON response
	var spec NetworkSpec
	if err := json.Unmarshal([]byte(resp.Output), &spec); err != nil {
		// Try to extract JSON from response
		jsonStr := s.extractJSONFromResponse(resp.Output)
		if jsonStr != "" {
			if err := json.Unmarshal([]byte(jsonStr), &spec); err != nil {
				return nil, fmt.Errorf("failed to parse LLM JSON response: %w", err)
			}
		} else {
			return nil, fmt.Errorf("no valid JSON found in LLM response: %s", resp.Output)
		}
	}

	// Set confidence from LLM or estimate based on validation
	if spec.Confidence == 0 {
		spec.Confidence = 0.8 // High confidence for LLM parsing
	}

	return &spec, nil
}

// parseWithRules uses rule-based parsing as fallback
func (s *NLPService) parseWithRules(text string) *NetworkSpec {
	text = strings.ToLower(strings.TrimSpace(text))

	spec := &NetworkSpec{
		Nodes:            []NodeSpec{},
		Policies:         []PolicySpec{},
		Config:           make(map[string]interface{}),
		Topology:         "custom",
		ValidationErrors: []string{},
		Confidence:       0.6, // Lower confidence for rule-based
	}

	// Extract topology type
	if strings.Contains(text, "star") || strings.Contains(text, "hub") {
		spec.Topology = "star"
	} else if strings.Contains(text, "mesh") {
		spec.Topology = "mesh"
	} else if strings.Contains(text, "tree") {
		spec.Topology = "tree"
	} else if strings.Contains(text, "ring") {
		spec.Topology = "ring"
	}

	// Extract number of nodes
	nodeCount := s.extractNumber(text)
	if nodeCount == 0 {
		nodeCount = s.extractNodeCount(text)
	}
	if nodeCount == 0 {
		nodeCount = 3 // default
	}

	// Extract node types (only supported ones)
	nodeTypes := s.extractSupportedNodeTypes(text)
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
			Name: fmt.Sprintf("%s-%d", strings.Title(nodeType), i+1),
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

	// Extract firewall policies if mentioned
	if strings.Contains(text, "firewall") || strings.Contains(text, "security") {
		firewallPolicy := PolicySpec{
			Name: "Basic Firewall",
			Type: "firewall",
			Rules: []map[string]interface{}{
				{
					"name":     "Allow HTTP",
					"action":   "allow",
					"port":     80,
					"protocol": "tcp",
				},
			},
		}
		spec.Policies = append(spec.Policies, firewallPolicy)
	}

	return spec
}

// extractJSONFromResponse attempts to extract JSON from LLM response
func (s *NLPService) extractJSONFromResponse(response string) string {
	// Look for JSON wrapped in code blocks
	jsonRegex := regexp.MustCompile(`(?s)\{.*\}`)
	matches := jsonRegex.FindAllString(response, -1)

	for _, match := range matches {
		// Try to parse each potential JSON match
		var test map[string]interface{}
		if err := json.Unmarshal([]byte(match), &test); err == nil {
			return match
		}
	}

	return ""
}

// extractSupportedNodeTypes extracts only supported node types
func (s *NLPService) extractSupportedNodeTypes(text string) []string {
	supportedTypes := s.validator.supportedTypes.NodeTypes
	var foundTypes []string

	for _, nodeType := range supportedTypes {
		if strings.Contains(text, strings.ToLower(nodeType)) {
			foundTypes = append(foundTypes, strings.ToLower(nodeType))
		}
	}

	return foundTypes
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
