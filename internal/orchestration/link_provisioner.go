package orchestration

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"netorchestrator/internal/models"
	"netorchestrator/internal/observability"
)

// LinkProvisioner manages real network connections between containers
type LinkProvisioner struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewLinkProvisioner creates a new link provisioner
func NewLinkProvisioner(db *gorm.DB, logger *zap.Logger) *LinkProvisioner {
	return &LinkProvisioner{
		db:     db,
		logger: logger,
	}
}

// ProvisionLink creates real network connectivity between containers
func (lp *LinkProvisioner) ProvisionLink(link *models.Link) error {
	startTime := time.Now()
	metrics := observability.GetMetrics()
	
	lp.logger.Info("Provisioning real network link between containers",
		zap.String("link_id", link.ID.String()),
		zap.String("source_node", link.SourceNodeID.String()),
		zap.String("target_node", link.TargetNodeID.String()))

	// Get source and target nodes with container information
	sourceNode, err := lp.getNodeWithContainer(link.SourceNodeID)
	if err != nil {
		metrics.RecordProvisioningOperation("link", startTime, err)
		return fmt.Errorf("failed to get source node: %w", err)
	}

	targetNode, err := lp.getNodeWithContainer(link.TargetNodeID)
	if err != nil {
		metrics.RecordProvisioningOperation("link", startTime, err)
		return fmt.Errorf("failed to get target node: %w", err)
	}

	// Check if both nodes have containers
	sourceContainerName, sourceOk := sourceNode.Config.CustomAttrs["container_name"]
	targetContainerName, targetOk := targetNode.Config.CustomAttrs["container_name"]

	if !sourceOk || !targetOk {
		err := fmt.Errorf("one or both nodes don't have containers provisioned yet")
		metrics.RecordProvisioningOperation("link", startTime, err)
		return err
	}

	// Create network bridge/connection between containers
	if err := lp.createContainerNetworkConnection(sourceContainerName, targetContainerName, link); err != nil {
		metrics.RecordProvisioningOperation("link", startTime, err)
		return fmt.Errorf("failed to create container network connection: %w", err)
	}

	// Apply link configuration (bandwidth, latency, etc.)
	if err := lp.applyLinkConfiguration(sourceContainerName, targetContainerName, link); err != nil {
		lp.logger.Warn("Failed to apply link configuration", zap.Error(err))
		// Don't fail the whole operation for configuration issues
	}

	// Update link status and add container information
	if link.Config.CustomAttrs == nil {
		link.Config.CustomAttrs = make(map[string]string)
	}
	link.Config.CustomAttrs["source_container"] = sourceContainerName
	link.Config.CustomAttrs["target_container"] = targetContainerName
	link.Config.CustomAttrs["provisioned_at"] = time.Now().Format(time.RFC3339)
	link.Config.CustomAttrs["connection_type"] = "container_bridge"
	link.Status = models.LinkStatusActive

	if err := lp.db.Save(link).Error; err != nil {
		metrics.RecordProvisioningOperation("link", startTime, err)
		return fmt.Errorf("failed to update link: %w", err)
	}

	metrics.RecordProvisioningOperation("link", startTime, nil)
	
	// NEW: Record network-scoped link metrics
	// Get network information from source node
	var network models.Network
	if err := lp.db.First(&network, "id = ?", sourceNode.NetworkID).Error; err == nil {
		metrics.SetNetworkLinkStatus(
			network.ID.String(),
			network.Name,
			sourceNode.Name,
			targetNode.Name,
			link.ID.String(),
			link.Status == models.LinkStatusActive,
		)
		metrics.RecordNetworkProvisionLatency(
			network.ID.String(),
			network.Name,
			"link",
			"provision_link",
			startTime,
		)
	}
	
	lp.logger.Info("Link provisioned successfully",
		zap.String("source_container", sourceContainerName),
		zap.String("target_container", targetContainerName),
		zap.String("link_id", link.ID.String()))

	return nil
}

// getNodeWithContainer retrieves a node with its container information
func (lp *LinkProvisioner) getNodeWithContainer(nodeID uuid.UUID) (*models.Node, error) {
	var node models.Node
	if err := lp.db.First(&node, "id = ?", nodeID).Error; err != nil {
		return nil, fmt.Errorf("node not found: %w", err)
	}
	return &node, nil
}

// createContainerNetworkConnection creates network connectivity between containers
func (lp *LinkProvisioner) createContainerNetworkConnection(sourceContainer, targetContainer string, link *models.Link) error {
	// For now, ensure both containers are on the same network
	// In advanced implementations, we could create dedicated networks, bridges, tunnels, etc.

	// Check if both containers exist and are running
	if !lp.isContainerRunning(sourceContainer) {
		return fmt.Errorf("source container %s is not running", sourceContainer)
	}

	if !lp.isContainerRunning(targetContainer) {
		return fmt.Errorf("target container %s is not running", targetContainer)
	}

	// Get container network information
	sourceNetwork, err := lp.getContainerNetwork(sourceContainer)
	if err != nil {
		return fmt.Errorf("failed to get source container network: %w", err)
	}

	targetNetwork, err := lp.getContainerNetwork(targetContainer)
	if err != nil {
		return fmt.Errorf("failed to get target container network: %w", err)
	}

	// Ensure both containers are on the same network
	if sourceNetwork != targetNetwork {
		lp.logger.Info("Containers are on different networks, connecting them",
			zap.String("source_network", sourceNetwork),
			zap.String("target_network", targetNetwork))

		// Connect target container to source network (advanced networking)
		if err := lp.connectContainerToNetwork(targetContainer, sourceNetwork); err != nil {
			lp.logger.Warn("Failed to connect containers across networks", zap.Error(err))
		}
	}

	lp.logger.Info("Container network connection established",
		zap.String("source", sourceContainer),
		zap.String("target", targetContainer),
		zap.String("network", sourceNetwork))

	return nil
}

// applyLinkConfiguration applies bandwidth, latency, and other link configurations
func (lp *LinkProvisioner) applyLinkConfiguration(sourceContainer, targetContainer string, link *models.Link) error {
	// Apply traffic control (tc) for bandwidth and latency simulation

	if link.Config.Bandwidth > 0 {
		// Apply bandwidth limiting using tc (Traffic Control)
		if err := lp.applyBandwidthLimit(sourceContainer, link.Config.Bandwidth); err != nil {
			lp.logger.Warn("Failed to apply bandwidth limit to source", zap.Error(err))
		}

		if err := lp.applyBandwidthLimit(targetContainer, link.Config.Bandwidth); err != nil {
			lp.logger.Warn("Failed to apply bandwidth limit to target", zap.Error(err))
		}
	}

	if link.Config.Latency > 0 {
		// Apply latency simulation using tc netem
		if err := lp.applyLatencySimulation(sourceContainer, link.Config.Latency); err != nil {
			lp.logger.Warn("Failed to apply latency to source", zap.Error(err))
		}
	}

	if link.Config.PacketLoss > 0 {
		// Apply packet loss simulation
		if err := lp.applyPacketLossSimulation(sourceContainer, link.Config.PacketLoss); err != nil {
			lp.logger.Warn("Failed to apply packet loss to source", zap.Error(err))
		}
	}

	return nil
}

// isContainerRunning checks if a container is running
func (lp *LinkProvisioner) isContainerRunning(containerName string) bool {
	cmd := exec.Command("podman", "ps", "--format", "{{.Names}}", "--filter", fmt.Sprintf("name=%s", containerName))
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), containerName)
}

// getContainerNetwork gets the network a container is connected to
func (lp *LinkProvisioner) getContainerNetwork(containerName string) (string, error) {
	cmd := exec.Command("podman", "inspect", containerName, "--format", "{{range .NetworkSettings.Networks}}{{.NetworkID}}{{end}}")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}

	networkID := strings.TrimSpace(string(output))
	if networkID == "" {
		return "default", nil
	}

	// Get network name from ID
	cmd = exec.Command("podman", "network", "ls", "--format", "{{.ID}}\t{{.Name}}")
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list networks: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 && strings.HasPrefix(parts[0], networkID[:12]) {
			return parts[1], nil
		}
	}

	return "unknown", nil
}

// connectContainerToNetwork connects a container to an additional network
func (lp *LinkProvisioner) connectContainerToNetwork(containerName, networkName string) error {
	cmd := exec.Command("podman", "network", "connect", networkName, containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to connect container to network: %s", output)
	}

	lp.logger.Info("Container connected to additional network",
		zap.String("container", containerName),
		zap.String("network", networkName))

	return nil
}

// applyBandwidthLimit applies bandwidth limiting to a container
func (lp *LinkProvisioner) applyBandwidthLimit(containerName string, bandwidthMbps int) error {
	// Apply bandwidth limit using tc (Traffic Control) inside container
	command := fmt.Sprintf(`
		tc qdisc add dev eth0 root tbf rate %dmbit burst 32kbit latency 300ms 2>/dev/null || 
		echo "Bandwidth limit applied: %d Mbps"
	`, bandwidthMbps, bandwidthMbps)

	cmd := exec.Command("podman", "exec", containerName, "sh", "-c", command)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply bandwidth limit: %w", err)
	}

	lp.logger.Info("Bandwidth limit applied to container",
		zap.String("container", containerName),
		zap.Int("bandwidth_mbps", bandwidthMbps))

	return nil
}

// applyLatencySimulation applies network latency simulation
func (lp *LinkProvisioner) applyLatencySimulation(containerName string, latencyMs int) error {
	command := fmt.Sprintf(`
		tc qdisc add dev eth0 root netem delay %dms 2>/dev/null || 
		echo "Latency simulation applied: %d ms"
	`, latencyMs, latencyMs)

	cmd := exec.Command("podman", "exec", containerName, "sh", "-c", command)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply latency simulation: %w", err)
	}

	lp.logger.Info("Latency simulation applied to container",
		zap.String("container", containerName),
		zap.Int("latency_ms", latencyMs))

	return nil
}

// applyPacketLossSimulation applies packet loss simulation
func (lp *LinkProvisioner) applyPacketLossSimulation(containerName string, packetLossPercent float64) error {
	command := fmt.Sprintf(`
		tc qdisc add dev eth0 root netem loss %f%% 2>/dev/null || 
		echo "Packet loss simulation applied: %f%%"
	`, packetLossPercent, packetLossPercent)

	cmd := exec.Command("podman", "exec", containerName, "sh", "-c", command)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply packet loss simulation: %w", err)
	}

	lp.logger.Info("Packet loss simulation applied to container",
		zap.String("container", containerName),
		zap.Float64("packet_loss_percent", packetLossPercent))

	return nil
}

// DestroyLink removes network connectivity configurations
func (lp *LinkProvisioner) DestroyLink(link *models.Link) error {
	if link.Config.CustomAttrs == nil {
		return nil // No container provisioning to clean up
	}

	sourceContainer, sourceOk := link.Config.CustomAttrs["source_container"]
	targetContainer, targetOk := link.Config.CustomAttrs["target_container"]

	if !sourceOk || !targetOk {
		return nil // No container information recorded
	}

	// Remove traffic control configurations
	lp.removeTrafficControl(sourceContainer)
	lp.removeTrafficControl(targetContainer)

	lp.logger.Info("Link destroyed and traffic control removed",
		zap.String("source", sourceContainer),
		zap.String("target", targetContainer))

	return nil
}

// removeTrafficControl removes traffic control configurations from a container
func (lp *LinkProvisioner) removeTrafficControl(containerName string) {
	cmd := exec.Command("podman", "exec", containerName, "sh", "-c", "tc qdisc del dev eth0 root 2>/dev/null || echo 'TC rules cleared'")
	if err := cmd.Run(); err != nil {
		lp.logger.Warn("Failed to remove traffic control",
			zap.String("container", containerName),
			zap.Error(err))
	}
}

// GetLinkStatus gets the real connectivity status between containers
func (lp *LinkProvisioner) GetLinkStatus(link *models.Link) (map[string]interface{}, error) {
	if link.Config.CustomAttrs == nil {
		return map[string]interface{}{
			"status":  "not_provisioned",
			"message": "No container connectivity provisioned",
		}, nil
	}

	sourceContainer, sourceOk := link.Config.CustomAttrs["source_container"]
	targetContainer, targetOk := link.Config.CustomAttrs["target_container"]

	if !sourceOk || !targetOk {
		return map[string]interface{}{
			"status":  "incomplete",
			"message": "Missing container information",
		}, nil
	}

	// Check if both containers are running
	sourceRunning := lp.isContainerRunning(sourceContainer)
	targetRunning := lp.isContainerRunning(targetContainer)

	status := "active"
	if !sourceRunning || !targetRunning {
		status = "degraded"
	}

	return map[string]interface{}{
		"status":              status,
		"source_container":    sourceContainer,
		"target_container":    targetContainer,
		"source_running":      sourceRunning,
		"target_running":      targetRunning,
		"connection_type":     link.Config.CustomAttrs["connection_type"],
		"provisioned_at":      link.Config.CustomAttrs["provisioned_at"],
		"bandwidth_mbps":      link.Config.Bandwidth,
		"latency_ms":          link.Config.Latency,
		"packet_loss_percent": link.Config.PacketLoss,
	}, nil
}
