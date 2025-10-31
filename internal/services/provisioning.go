package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"netorchestrator/internal/events"
	"netorchestrator/internal/models"
)

// ProvisioningService handles network provisioning based on events
type ProvisioningService struct {
	db             *gorm.DB
	logger         *zap.Logger
	eventSubscriber *events.EventSubscriber
}

// NewProvisioningService creates a new provisioning service
func NewProvisioningService(db *gorm.DB, logger *zap.Logger, eventSubscriber *events.EventSubscriber) *ProvisioningService {
	return &ProvisioningService{
		db:             db,
		logger:         logger,
		eventSubscriber: eventSubscriber,
	}
}

// Start subscribes to events and handles provisioning
func (s *ProvisioningService) Start(ctx context.Context) error {
	if s.eventSubscriber == nil {
		s.logger.Warn("Event subscriber not available, provisioning service will not start")
		return nil
	}

	// Subscribe to network.created events
	if err := s.eventSubscriber.Subscribe("network.created", s.handleNetworkCreated); err != nil {
		return fmt.Errorf("failed to subscribe to network.created: %w", err)
	}

	s.logger.Info("Provisioning service started, listening for events")
	return nil
}

// handleNetworkCreated handles network.created events
func (s *ProvisioningService) handleNetworkCreated(data []byte) error {
	var event events.NetworkCreatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal network.created event: %w", err)
	}

	s.logger.Info("Processing network.created event",
		zap.String("network_id", event.NetworkID),
		zap.String("name", event.Name))

	// Parse network ID
	networkID, err := uuid.Parse(event.NetworkID)
	if err != nil {
		return fmt.Errorf("invalid network ID: %w", err)
	}

	// Get network from database
	var network models.Network
	if err := s.db.Where("id = ?", networkID).First(&network).Error; err != nil {
		s.logger.Error("Failed to get network for provisioning", zap.Error(err))
		return err
	}

	// Provision network (mock implementation)
	return s.provisionNetwork(context.Background(), &network)
}

// provisionNetwork provisions a network topology
func (s *ProvisioningService) provisionNetwork(ctx context.Context, network *models.Network) error {
	s.logger.Info("Provisioning network",
		zap.String("network_id", network.ID.String()),
		zap.String("name", network.Name))

	// Update network status to provisioning
	network.Status = models.NetworkStatusProvisioning
	if err := s.db.Save(network).Error; err != nil {
		return fmt.Errorf("failed to update network status: %w", err)
	}

	// Simulate provisioning delay
	time.Sleep(500 * time.Millisecond)

	// Update network status to active
	network.Status = models.NetworkStatusActive
	if err := s.db.Save(network).Error; err != nil {
		return fmt.Errorf("failed to activate network: %w", err)
	}

	s.logger.Info("Network provisioned successfully",
		zap.String("network_id", network.ID.String()))

	// Publish network provisioned event
	// TODO: Wire event publisher here when needed

	return nil
}

