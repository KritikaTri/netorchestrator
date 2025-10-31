package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"netorchestrator/internal/models"
)

// NetworkService handles network-related operations
type NetworkService struct {
	db     *gorm.DB
	cache  *redis.Client
	logger *zap.Logger
}

// NewNetworkService creates a new network service
func NewNetworkService(db *gorm.DB, cache *redis.Client, logger *zap.Logger) *NetworkService {
	return &NetworkService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// CreateNetwork creates a new network
func (s *NetworkService) CreateNetwork(ctx context.Context, network *models.Network) error {
	if s.db == nil {
		s.logger.Warn("Database not available - running in simple mode")
		return fmt.Errorf("database not available - running in simple mode")
	}

	if err := s.db.WithContext(ctx).Create(network).Error; err != nil {
		s.logger.Error("Failed to create network", zap.Error(err))
		return fmt.Errorf("failed to create network: %w", err)
	}

	s.logger.Info("Network created successfully", zap.String("network_id", network.ID.String()))
	return nil
}

// GetNetwork retrieves a network by ID
func (s *NetworkService) GetNetwork(ctx context.Context, id uuid.UUID) (*models.Network, error) {
	var network models.Network
	if err := s.db.WithContext(ctx).Preload("Nodes").Preload("Links").Preload("Policies").First(&network, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("network not found")
		}
		s.logger.Error("Failed to get network", zap.Error(err))
		return nil, fmt.Errorf("failed to get network: %w", err)
	}

	return &network, nil
}

// ListNetworks retrieves all networks for a user
func (s *NetworkService) ListNetworks(ctx context.Context, userID uuid.UUID) ([]models.Network, error) {
	var networks []models.Network
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&networks).Error; err != nil {
		s.logger.Error("Failed to list networks", zap.Error(err))
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	return networks, nil
}

// ListAllNetworks retrieves all networks (for admin users)
func (s *NetworkService) ListAllNetworks(ctx context.Context) ([]models.Network, error) {
	var networks []models.Network
	if err := s.db.WithContext(ctx).Find(&networks).Error; err != nil {
		s.logger.Error("Failed to list all networks", zap.Error(err))
		return nil, fmt.Errorf("failed to list all networks: %w", err)
	}

	return networks, nil
}

// GetUserByUsername retrieves a user UUID by username
func (s *NetworkService) GetUserByUsername(ctx context.Context, username string) (uuid.UUID, error) {
	if s.db == nil {
		s.logger.Warn("Database not available - running in simple mode")
		return uuid.Nil, fmt.Errorf("database not available")
	}

	var user models.User
	if err := s.db.WithContext(ctx).Select("id").Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return uuid.Nil, fmt.Errorf("user not found: %s", username)
		}
		s.logger.Error("Failed to get user by username", zap.String("username", username), zap.Error(err))
		return uuid.Nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user.ID, nil
}

// UpdateNetwork updates an existing network
func (s *NetworkService) UpdateNetwork(ctx context.Context, network *models.Network) error {
	if err := s.db.WithContext(ctx).Save(network).Error; err != nil {
		s.logger.Error("Failed to update network", zap.Error(err))
		return fmt.Errorf("failed to update network: %w", err)
	}

	s.logger.Info("Network updated successfully", zap.String("network_id", network.ID.String()))

	return nil
}

// DeleteNetwork deletes a network
func (s *NetworkService) DeleteNetwork(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.Network{}, "id = ?", id).Error; err != nil {
		s.logger.Error("Failed to delete network", zap.Error(err))
		return fmt.Errorf("failed to delete network: %w", err)
	}

	s.logger.Info("Network deleted successfully", zap.String("network_id", id.String()))
	return nil
}

// CreateNode creates a new node in a network
func (s *NetworkService) CreateNode(ctx context.Context, node *models.Node) error {
	if err := s.db.WithContext(ctx).Create(node).Error; err != nil {
		s.logger.Error("Failed to create node", zap.Error(err))
		return fmt.Errorf("failed to create node: %w", err)
	}

	s.logger.Info("Node created successfully", zap.String("node_id", node.ID.String()))
	return nil
}

// GetNode retrieves a node by ID
func (s *NetworkService) GetNode(ctx context.Context, id uuid.UUID) (*models.Node, error) {
	var node models.Node
	if err := s.db.WithContext(ctx).Preload("Network").Preload("Links").First(&node, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("node not found")
		}
		s.logger.Error("Failed to get node", zap.Error(err))
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	return &node, nil
}

// ListNodes retrieves all nodes for a network
func (s *NetworkService) ListNodes(ctx context.Context, networkID uuid.UUID) ([]models.Node, error) {
	var nodes []models.Node
	if err := s.db.WithContext(ctx).Where("network_id = ?", networkID).Find(&nodes).Error; err != nil {
		s.logger.Error("Failed to list nodes", zap.Error(err))
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	return nodes, nil
}

// ListAllNodes retrieves all nodes across all networks
func (s *NetworkService) ListAllNodes(ctx context.Context) ([]models.Node, error) {
	if s.db == nil {
		s.logger.Warn("Database not available - running in simple mode")
		return nil, fmt.Errorf("database not available")
	}

	var nodes []models.Node
	if err := s.db.WithContext(ctx).Preload("Network").Find(&nodes).Error; err != nil {
		s.logger.Error("Failed to list all nodes", zap.Error(err))
		return nil, fmt.Errorf("failed to list all nodes: %w", err)
	}

	return nodes, nil
}

// UpdateNode updates an existing node
func (s *NetworkService) UpdateNode(ctx context.Context, node *models.Node) error {
	if err := s.db.WithContext(ctx).Save(node).Error; err != nil {
		s.logger.Error("Failed to update node", zap.Error(err))
		return fmt.Errorf("failed to update node: %w", err)
	}

	s.logger.Info("Node updated successfully", zap.String("node_id", node.ID.String()))
	return nil
}

// DeleteNode deletes a node
func (s *NetworkService) DeleteNode(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.Node{}, "id = ?", id).Error; err != nil {
		s.logger.Error("Failed to delete node", zap.Error(err))
		return fmt.Errorf("failed to delete node: %w", err)
	}

	s.logger.Info("Node deleted successfully", zap.String("node_id", id.String()))
	return nil
}

// CreateLink creates a new link between nodes
func (s *NetworkService) CreateLink(ctx context.Context, link *models.Link) error {
	if err := s.db.WithContext(ctx).Create(link).Error; err != nil {
		s.logger.Error("Failed to create link", zap.Error(err))
		return fmt.Errorf("failed to create link: %w", err)
	}

	s.logger.Info("Link created successfully", zap.String("link_id", link.ID.String()))
	return nil
}

// GetLink retrieves a link by ID
func (s *NetworkService) GetLink(ctx context.Context, id uuid.UUID) (*models.Link, error) {
	var link models.Link
	if err := s.db.WithContext(ctx).Preload("Network").Preload("SourceNode").Preload("TargetNode").First(&link, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("link not found")
		}
		s.logger.Error("Failed to get link", zap.Error(err))
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	return &link, nil
}

// ListLinks retrieves all links for a network
func (s *NetworkService) ListLinks(ctx context.Context, networkID uuid.UUID) ([]models.Link, error) {
	var links []models.Link
	if err := s.db.WithContext(ctx).Where("network_id = ?", networkID).Preload("SourceNode").Preload("TargetNode").Find(&links).Error; err != nil {
		s.logger.Error("Failed to list links", zap.Error(err))
		return nil, fmt.Errorf("failed to list links: %w", err)
	}

	return links, nil
}

// UpdateLink updates an existing link
func (s *NetworkService) UpdateLink(ctx context.Context, link *models.Link) error {
	if err := s.db.WithContext(ctx).Save(link).Error; err != nil {
		s.logger.Error("Failed to update link", zap.Error(err))
		return fmt.Errorf("failed to update link: %w", err)
	}

	s.logger.Info("Link updated successfully", zap.String("link_id", link.ID.String()))
	return nil
}

// DeleteLink deletes a link
func (s *NetworkService) DeleteLink(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.Link{}, "id = ?", id).Error; err != nil {
		s.logger.Error("Failed to delete link", zap.Error(err))
		return fmt.Errorf("failed to delete link: %w", err)
	}

	s.logger.Info("Link deleted successfully", zap.String("link_id", id.String()))
	return nil
}

// CreatePolicy creates a new network policy
func (s *NetworkService) CreatePolicy(ctx context.Context, policy *models.Policy) error {
	if err := s.db.WithContext(ctx).Create(policy).Error; err != nil {
		s.logger.Error("Failed to create policy", zap.Error(err))
		return fmt.Errorf("failed to create policy: %w", err)
	}

	s.logger.Info("Policy created successfully", zap.String("policy_id", policy.ID.String()))
	return nil
}

// GetPolicy retrieves a policy by ID
func (s *NetworkService) GetPolicy(ctx context.Context, id uuid.UUID) (*models.Policy, error) {
	var policy models.Policy
	if err := s.db.WithContext(ctx).Preload("Network").First(&policy, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("policy not found")
		}
		s.logger.Error("Failed to get policy", zap.Error(err))
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	return &policy, nil
}

// ListPolicies retrieves all policies for a network
func (s *NetworkService) ListPolicies(ctx context.Context, networkID uuid.UUID) ([]models.Policy, error) {
	var policies []models.Policy
	if err := s.db.WithContext(ctx).
		Select("id, network_id, name, policy_type, status, created_at, updated_at").
		Where("network_id = ?", networkID).
		Find(&policies).Error; err != nil {
		s.logger.Error("Failed to list policies", zap.Error(err))
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	return policies, nil
}

// ListAllPolicies retrieves all policies across all networks
func (s *NetworkService) ListAllPolicies(ctx context.Context) ([]models.Policy, error) {
	var policies []models.Policy
	if err := s.db.WithContext(ctx).
		Select("id, network_id, name, policy_type, status, created_at, updated_at").
		Preload("Network").
		Find(&policies).Error; err != nil {
		s.logger.Error("Failed to list all policies", zap.Error(err))
		return nil, fmt.Errorf("failed to list all policies: %w", err)
	}

	return policies, nil
}

// UpdatePolicy updates an existing policy
func (s *NetworkService) UpdatePolicy(ctx context.Context, policy *models.Policy) error {
	if err := s.db.WithContext(ctx).Save(policy).Error; err != nil {
		s.logger.Error("Failed to update policy", zap.Error(err))
		return fmt.Errorf("failed to update policy: %w", err)
	}

	s.logger.Info("Policy updated successfully", zap.String("policy_id", policy.ID.String()))
	return nil
}

// DeletePolicy deletes a policy
func (s *NetworkService) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&models.Policy{}, "id = ?", id).Error; err != nil {
		s.logger.Error("Failed to delete policy", zap.Error(err))
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	s.logger.Info("Policy deleted successfully", zap.String("policy_id", id.String()))
	return nil
}
