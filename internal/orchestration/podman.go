package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"netorchestrator/internal/models"
)

type PodmanOrchestrator struct {
	logger *zap.Logger
}

func NewPodmanOrchestrator(logger *zap.Logger) *PodmanOrchestrator {
	return &PodmanOrchestrator{
		logger: logger,
	}
}

// StartNetwork creates and starts all containers for a network
func (p *PodmanOrchestrator) StartNetwork(ctx context.Context, network *models.Network, nodes []models.Node) error {
	p.logger.Info("Starting network", zap.String("network_id", network.ID.String()))

	// Create Podman network
	if err := p.createPodmanNetwork(network); err != nil {
		return fmt.Errorf("failed to create podman network: %w", err)
	}

	// Start all nodes
	for _, node := range nodes {
		if err := p.startNode(ctx, &node, network); err != nil {
			p.logger.Error("Failed to start node", zap.String("node_id", node.ID.String()), zap.Error(err))
			continue
		}
	}

	p.logger.Info("Network started successfully", zap.String("network_id", network.ID.String()))
	return nil
}

// StopNetwork stops all containers in a network
func (p *PodmanOrchestrator) StopNetwork(ctx context.Context, networkID uuid.UUID) error {
	p.logger.Info("Stopping network", zap.String("network_id", networkID.String()))

	// Get all containers in the network
	cmd := exec.CommandContext(ctx, "podman", "ps", "--filter", fmt.Sprintf("label=network_id=%s", networkID.String()), "--format", "{{.ID}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	containerIDs := strings.Fields(string(output))
	for _, containerID := range containerIDs {
		if err := p.stopContainer(ctx, containerID); err != nil {
			p.logger.Error("Failed to stop container", zap.String("container_id", containerID), zap.Error(err))
		}
	}

	// Remove Podman network
	if err := p.removePodmanNetwork(ctx, networkID.String()); err != nil {
		p.logger.Warn("Failed to remove podman network", zap.Error(err))
	}

	p.logger.Info("Network stopped successfully", zap.String("network_id", networkID.String()))
	return nil
}

// StartNode starts a single node container
func (p *PodmanOrchestrator) StartNode(ctx context.Context, node *models.Node, network *models.Network) error {
	return p.startNode(ctx, node, network)
}

// StopNode stops a single node container
func (p *PodmanOrchestrator) StopNode(ctx context.Context, nodeID uuid.UUID) error {
	p.logger.Info("Stopping node", zap.String("node_id", nodeID.String()))

	// Find container by node ID
	cmd := exec.CommandContext(ctx, "podman", "ps", "--filter", fmt.Sprintf("label=node_id=%s", nodeID.String()), "--format", "{{.ID}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to find container: %w", err)
	}

	containerID := strings.TrimSpace(string(output))
	if containerID == "" {
		return fmt.Errorf("container not found for node")
	}

	return p.stopContainer(ctx, containerID)
}

// GetNodeStatus gets the runtime status of a node
func (p *PodmanOrchestrator) GetNodeStatus(ctx context.Context, nodeID uuid.UUID) (string, error) {
	cmd := exec.CommandContext(ctx, "podman", "ps", "--filter", fmt.Sprintf("label=node_id=%s", nodeID.String()), "--format", "{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		return "unknown", err
	}

	status := strings.TrimSpace(string(output))
	if status == "" {
		return "stopped", nil
	}

	return status, nil
}

// Private helper methods

func (p *PodmanOrchestrator) createPodmanNetwork(network *models.Network) error {
	networkName := fmt.Sprintf("netorch_%s", network.ID.String()[:8])
	
	args := []string{"network", "create"}
	
	if network.Subnet != "" {
		args = append(args, "--subnet", network.Subnet)
	}
	if network.Gateway != "" {
		args = append(args, "--gateway", network.Gateway)
	}
	
	args = append(args, networkName)

	cmd := exec.Command("podman", args...)
	if err := cmd.Run(); err != nil {
		// Network might already exist, check
		if !p.networkExists(networkName) {
			return fmt.Errorf("failed to create network: %w", err)
		}
	}

	p.logger.Info("Podman network created", zap.String("network_name", networkName))
	return nil
}

func (p *PodmanOrchestrator) networkExists(networkName string) bool {
	cmd := exec.Command("podman", "network", "exists", networkName)
	return cmd.Run() == nil
}

func (p *PodmanOrchestrator) startNode(ctx context.Context, node *models.Node, network *models.Network) error {
	p.logger.Info("Starting node", zap.String("node_id", node.ID.String()))

	containerName := fmt.Sprintf("netorch_%s_%s", network.ID.String()[:8], node.Name)
	networkName := fmt.Sprintf("netorch_%s", network.ID.String()[:8])

	args := []string{
		"run", "-d",
		"--name", containerName,
		"--network", networkName,
		"--label", fmt.Sprintf("network_id=%s", network.ID.String()),
		"--label", fmt.Sprintf("node_id=%s", node.ID.String()),
		"--label", "managed_by=netorchestrator",
	}

	// Set IP address if specified
	if node.IPAddress != "" {
		args = append(args, "--ip", node.IPAddress)
	}

	// Set resource limits
	if node.CPUCores > 0 {
		args = append(args, "--cpus", fmt.Sprintf("%d", node.CPUCores))
	}
	if node.MemoryMB > 0 {
		args = append(args, "--memory", fmt.Sprintf("%dm", node.MemoryMB))
	}

	// Add port mappings
	for portStr := range node.Ports {
		args = append(args, "-p", portStr)
	}

	// Add environment variables
	for key, value := range node.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Use specified image or default
	image := node.Image
	if image == "" {
		image = "nginx:alpine" // Default image
	}

	args = append(args, image)

	cmd := exec.CommandContext(ctx, "podman", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	p.logger.Info("Node started successfully", 
		zap.String("node_id", node.ID.String()),
		zap.String("container_name", containerName))
	return nil
}

func (p *PodmanOrchestrator) stopContainer(ctx context.Context, containerID string) error {
	cmd := exec.CommandContext(ctx, "podman", "stop", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", containerID, err)
	}

	// Remove the container
	cmd = exec.CommandContext(ctx, "podman", "rm", containerID)
	if err := cmd.Run(); err != nil {
		p.logger.Warn("Failed to remove container", zap.String("container_id", containerID), zap.Error(err))
	}

	return nil
}

func (p *PodmanOrchestrator) removePodmanNetwork(ctx context.Context, networkID string) error {
	networkName := fmt.Sprintf("netorch_%s", networkID[:8])
	cmd := exec.CommandContext(ctx, "podman", "network", "rm", networkName)
	return cmd.Run()
}

// GetContainerStats gets real-time stats for a container
func (p *PodmanOrchestrator) GetContainerStats(ctx context.Context, nodeID uuid.UUID) (map[string]interface{}, error) {
	cmd := exec.CommandContext(ctx, "podman", "stats", "--format", "json", "--no-stream",
		"--filter", fmt.Sprintf("label=node_id=%s", nodeID.String()))
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(output, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse stats: %w", err)
	}

	return stats, nil
}