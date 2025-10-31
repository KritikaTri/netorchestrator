package optimization

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"netorchestrator/internal/models"
)

// OptimizationService provides AI-powered network optimization
type OptimizationService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewOptimizationService creates a new optimization service
func NewOptimizationService(db *gorm.DB, logger *zap.Logger) *OptimizationService {
	return &OptimizationService{
		db:     db,
		logger: logger,
	}
}

// AnalyzeNetwork performs comprehensive network analysis
type NetworkAnalysis struct {
	NetworkID           string                 `json:"network_id"`
	Issues              []NetworkIssue         `json:"issues"`
	Recommendations     []Recommendation       `json:"recommendations"`
	HealthScore         float64                `json:"health_score"` // 0-100
	EfficiencyScore     float64                `json:"efficiency_score"`
	ComplianceScore     float64                `json:"compliance_score"`
}

type NetworkIssue struct {
	Type        string `json:"type"` // conflict, inefficiency, security, compliance
	Severity    string `json:"severity"` // critical, warning, info
	Description string `json:"description"`
	ResourceID  string `json:"resource_id,omitempty"`
}

type Recommendation struct {
	Type        string                 `json:"type"` // topology_change, ip_reallocation, routing_optimization
	Priority    string                 `json:"priority"` // high, medium, low
	Description string                 `json:"description"`
	Action      map[string]interface{} `json:"action,omitempty"`
	Impact      string                 `json:"impact"` // expected improvement
}

// AnalyzeNetwork analyzes a network and returns recommendations
func (s *OptimizationService) AnalyzeNetwork(ctx context.Context, networkID uuid.UUID) (*NetworkAnalysis, error) {
	// Get network with all relationships
	var network models.Network
	if err := s.db.WithContext(ctx).
		Preload("Nodes").
		Preload("Links").
		Preload("Policies").
		First(&network, "id = ?", networkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("network not found")
		}
		return nil, fmt.Errorf("failed to get network: %w", err)
	}

	analysis := &NetworkAnalysis{
		NetworkID: networkID.String(),
		Issues:    []NetworkIssue{},
		Recommendations: []Recommendation{},
	}

	// Detect subnet conflicts
	issues, recommendations := s.detectSubnetConflicts(&network)
	analysis.Issues = append(analysis.Issues, issues...)
	analysis.Recommendations = append(analysis.Recommendations, recommendations...)

	// Check routing efficiency
	routeIssues, routeRecs := s.analyzeRoutingEfficiency(&network)
	analysis.Issues = append(analysis.Issues, routeIssues...)
	analysis.Recommendations = append(analysis.Recommendations, routeRecs...)

	// Validate configuration compliance
	compIssues, compRecs := s.validateConfigurationCompliance(&network)
	analysis.Issues = append(analysis.Issues, compIssues...)
	analysis.Recommendations = append(analysis.Recommendations, compRecs...)

	// Calculate scores
	analysis.HealthScore = s.calculateHealthScore(&network, analysis.Issues)
	analysis.EfficiencyScore = s.calculateEfficiencyScore(&network)
	analysis.ComplianceScore = s.calculateComplianceScore(compIssues)

	return analysis, nil
}

// RecommendTopologyChange suggests topology improvements
func (s *OptimizationService) RecommendTopologyChange(ctx context.Context, networkID uuid.UUID) (*Recommendation, error) {
	analysis, err := s.AnalyzeNetwork(ctx, networkID)
	if err != nil {
		return nil, err
	}

	// Find highest priority topology recommendation
	for _, rec := range analysis.Recommendations {
		if rec.Type == "topology_change" && rec.Priority == "high" {
			return &rec, nil
		}
	}

	return nil, fmt.Errorf("no topology recommendations available")
}

// ValidateConfiguration validates network configuration
func (s *OptimizationService) ValidateConfiguration(ctx context.Context, networkID uuid.UUID) ([]NetworkIssue, error) {
	analysis, err := s.AnalyzeNetwork(ctx, networkID)
	if err != nil {
		return nil, err
	}

	return analysis.Issues, nil
}

// Helper methods

func (s *OptimizationService) detectSubnetConflicts(network *models.Network) ([]NetworkIssue, []Recommendation) {
	issues := []NetworkIssue{}
	recommendations := []Recommendation{}

	// Check for duplicate IP addresses
	ipMap := make(map[string]string) // IP -> NodeID
	for _, node := range network.Nodes {
		if node.IPAddress != "" {
			if _, exists := ipMap[node.IPAddress]; exists {
				issues = append(issues, NetworkIssue{
					Type:        "conflict",
					Severity:    "critical",
					Description: fmt.Sprintf("IP address conflict: %s is used by multiple nodes", node.IPAddress),
					ResourceID:  node.ID.String(),
				})
				recommendations = append(recommendations, Recommendation{
					Type:        "ip_reallocation",
					Priority:    "high",
					Description: fmt.Sprintf("Reallocate IP address %s to resolve conflict between nodes", node.IPAddress),
					Impact:      "Resolves IP conflict, prevents connectivity issues",
				})
			} else {
				ipMap[node.IPAddress] = node.ID.String()
			}
		}
	}

	return issues, recommendations
}

func (s *OptimizationService) analyzeRoutingEfficiency(network *models.Network) ([]NetworkIssue, []Recommendation) {
	issues := []NetworkIssue{}
	recommendations := []Recommendation{}

	// Check for isolated nodes (no links)
	for _, node := range network.Nodes {
		hasLinks := false
		for _, link := range network.Links {
			if link.SourceNodeID == node.ID || link.TargetNodeID == node.ID {
				hasLinks = true
				break
			}
		}
		if !hasLinks && len(network.Nodes) > 1 {
			issues = append(issues, NetworkIssue{
				Type:        "inefficiency",
				Severity:    "warning",
				Description: fmt.Sprintf("Node %s is isolated (no links)", node.Name),
				ResourceID:  node.ID.String(),
			})
			recommendations = append(recommendations, Recommendation{
				Type:        "topology_change",
				Priority:    "medium",
				Description: fmt.Sprintf("Connect isolated node %s to network topology", node.Name),
				Impact:      "Improves network connectivity and redundancy",
			})
		}
	}

	// Check for star topology efficiency
	if len(network.Nodes) > 3 {
		// Suggest mesh for better redundancy if nodes support it
		recommendations = append(recommendations, Recommendation{
			Type:        "topology_change",
			Priority:    "low",
			Description: "Consider mesh topology for improved redundancy",
			Impact:      "Better fault tolerance and load distribution",
		})
	}

	return issues, recommendations
}

func (s *OptimizationService) validateConfigurationCompliance(network *models.Network) ([]NetworkIssue, []Recommendation) {
	issues := []NetworkIssue{}
	recommendations := []Recommendation{}

	// Check for missing security policies
	if len(network.Policies) == 0 {
		issues = append(issues, NetworkIssue{
			Type:        "security",
			Severity:    "warning",
			Description: "Network has no security policies configured",
		})
		recommendations = append(recommendations, Recommendation{
			Type:        "topology_change",
			Priority:    "medium",
			Description: "Add firewall or security policies to protect network",
			Impact:      "Improves network security posture",
		})
	}

	// Validate subnet configuration
	if network.Config.Subnet == "" {
		issues = append(issues, NetworkIssue{
			Type:        "compliance",
			Severity:    "warning",
			Description: "Network subnet not configured",
		})
	}

	return issues, recommendations
}

func (s *OptimizationService) calculateHealthScore(network *models.Network, issues []NetworkIssue) float64 {
	baseScore := 100.0
	
	for _, issue := range issues {
		switch issue.Severity {
		case "critical":
			baseScore -= 20
		case "warning":
			baseScore -= 10
		case "info":
			baseScore -= 5
		}
	}

	if baseScore < 0 {
		baseScore = 0
	}
	return baseScore
}

func (s *OptimizationService) calculateEfficiencyScore(network *models.Network) float64 {
	if len(network.Nodes) == 0 {
		return 0
	}

	// Simple metric: ratio of linked nodes
	linkedNodes := make(map[uuid.UUID]bool)
	for _, link := range network.Links {
		linkedNodes[link.SourceNodeID] = true
		linkedNodes[link.TargetNodeID] = true
	}

	if len(linkedNodes) == 0 {
		return 0
	}

	efficiency := float64(len(linkedNodes)) / float64(len(network.Nodes)) * 100
	return efficiency
}

func (s *OptimizationService) calculateComplianceScore(compIssues []NetworkIssue) float64 {
	if len(compIssues) == 0 {
		return 100.0
	}

	deductions := float64(len(compIssues)) * 15
	score := 100.0 - deductions
	if score < 0 {
		score = 0
	}
	return score
}

