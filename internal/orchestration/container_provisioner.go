package orchestration

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"netorchestrator/internal/models"
	"netorchestrator/internal/observability"
)

// ContainerProvisioner bridges virtual networks with real container infrastructure
type ContainerProvisioner struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewContainerProvisioner creates a new container provisioner
func NewContainerProvisioner(db *gorm.DB, logger *zap.Logger) *ContainerProvisioner {
	return &ContainerProvisioner{
		db:     db,
		logger: logger,
	}
}

// ProvisionNetworkInfrastructure creates real containers when a network is created
func (cp *ContainerProvisioner) ProvisionNetworkInfrastructure(network *models.Network) error {
	startTime := time.Now()
	metrics := observability.GetMetrics()

	cp.logger.Info("Provisioning real infrastructure for network",
		zap.String("network_id", network.ID.String()),
		zap.String("network_name", network.Name))

	// Create network-specific container subnet
	networkSubnet := fmt.Sprintf("net_%s", network.ID.String()[:8])

	// Create dedicated network for this virtual network
	if err := cp.createPodmanNetwork(networkSubnet, network); err != nil {
		cp.logger.Error("Failed to create network", zap.Error(err))
		metrics.RecordProvisioningOperation("network", startTime, err)
		return err
	}

	// Update network status to active after successful provisioning
	network.Status = models.NetworkStatusActive
	if err := cp.db.Save(network).Error; err != nil {
		metrics.RecordProvisioningOperation("network", startTime, err)
		return fmt.Errorf("failed to update network status: %w", err)
	}

	cp.logger.Info("Network infrastructure provisioning completed successfully",
		zap.String("container_network", networkSubnet),
		zap.String("status", string(models.NetworkStatusActive)))

	metrics.RecordProvisioningOperation("network", startTime, nil)

	// NEW: Record network-level provisioning latency
	metrics.RecordNetworkProvisionLatency(
		network.ID.String(),
		network.Name,
		"network",
		"provision_network",
		startTime,
	)

	return nil
}

// ProvisionNodeContainer creates a real container when a node is added
func (cp *ContainerProvisioner) ProvisionNodeContainer(node *models.Node, network *models.Network) error {
	startTime := time.Now()
	metrics := observability.GetMetrics()

	cp.logger.Info("Provisioning real container for node",
		zap.String("node_id", node.ID.String()),
		zap.String("node_name", node.Name),
		zap.String("network_id", network.ID.String()))

	// Determine container image based on node type
	image := cp.getContainerImageForNode(node)
	networkName := fmt.Sprintf("net_%s", network.ID.String()[:8])
	containerName := fmt.Sprintf("node_%s", node.ID.String()[:8])

	// Create container and get the container ID
	containerID, err := cp.createNodeContainer(containerName, image, networkName, node)
	if err != nil {
		cp.logger.Error("Failed to create node container", zap.Error(err))
		metrics.RecordProvisioningOperation("node", startTime, err)
		metrics.RecordContainerOperation("create", err)
		return err
	}

	// Update node with container information and entity_id
	if node.Config.CustomAttrs == nil {
		node.Config.CustomAttrs = make(map[string]string)
	}
	node.Config.CustomAttrs["container_name"] = containerName
	node.Config.CustomAttrs["container_image"] = image
	node.Config.CustomAttrs["container_network"] = networkName
	node.Config.CustomAttrs["provisioned_at"] = time.Now().Format(time.RFC3339)
	node.EntityID = containerID // Set the entity_id for metrics mapping
	node.Status = models.NodeStatusActive

	if err := cp.db.Save(node).Error; err != nil {
		metrics.RecordProvisioningOperation("node", startTime, err)
		return fmt.Errorf("failed to update node: %w", err)
	}

	cp.logger.Info("Node container provisioned successfully",
		zap.String("container_name", containerName),
		zap.String("image", image))

	metrics.RecordProvisioningOperation("node", startTime, nil)
	metrics.RecordContainerOperation("create", nil)

	// NEW: Record network-scoped metrics
	metrics.RecordNetworkProvisionLatency(
		network.ID.String(),
		network.Name,
		string(node.Type),
		"provision_node",
		startTime,
	)
	metrics.IncrementNetworkContainerCount(
		network.ID.String(),
		network.Name,
		string(node.Type),
	)

	return nil
}

// getContainerImageForNode determines the best container image for a node type
func (cp *ContainerProvisioner) getContainerImageForNode(node *models.Node) string {
	// Check node type or name to determine appropriate image
	nodeType := strings.ToLower(string(node.Type))
	nodeName := strings.ToLower(node.Name)

	switch {
	case strings.Contains(nodeType, "router") || strings.Contains(nodeName, "router"):
		return "frrouting/frr:latest"
	case strings.Contains(nodeType, "switch") || strings.Contains(nodeName, "switch"):
		return "alpine:latest" // Could be enhanced with OpenSwitch
	case strings.Contains(nodeType, "firewall") || strings.Contains(nodeName, "firewall"):
		return "alpine:latest" // Could be enhanced with pfSense
	case strings.Contains(nodeType, "server") || strings.Contains(nodeName, "server"):
		return "ubuntu:20.04"
	default:
		return "alpine:latest" // Default lightweight option
	}
}

// createPodmanNetwork creates a dedicated network for the virtual network with specified subnet
func (cp *ContainerProvisioner) createPodmanNetwork(networkName string, network *models.Network) error {
	metrics := observability.GetMetrics()

	// Build command with subnet configuration if available
	args := []string{"network", "create"}

	// Add subnet configuration if specified in the network config
	if network.Config.Subnet != "" {
		args = append(args, "--subnet", network.Config.Subnet)

		// Add gateway if specified
		if network.Config.Gateway != "" {
			args = append(args, "--gateway", network.Config.Gateway)
		}

		cp.logger.Info("Creating container network with custom subnet",
			zap.String("network", networkName),
			zap.String("subnet", network.Config.Subnet),
			zap.String("gateway", network.Config.Gateway))
	} else {
		cp.logger.Info("Creating container network with default subnet",
			zap.String("network", networkName))
	}

	// Add network name as final argument
	args = append(args, networkName)

	cmd := exec.Command("podman", args...)
	output, err := cmd.CombinedOutput()

	if err != nil && !strings.Contains(string(output), "already exists") {
		metrics.RecordPodmanError("network_create")

		// Parse error message for better user feedback
		outputStr := string(output)
		if strings.Contains(outputStr, "subnet") && strings.Contains(outputStr, "already used") {
			return fmt.Errorf("subnet %s is already in use by another network. Please choose a different subnet range", network.Config.Subnet)
		} else if strings.Contains(outputStr, "gateway") {
			return fmt.Errorf("gateway %s conflicts with existing network configuration. Please choose a different gateway", network.Config.Gateway)
		} else if strings.Contains(outputStr, "invalid") && strings.Contains(outputStr, "CIDR") {
			return fmt.Errorf("invalid subnet format: %s. Please use valid CIDR notation (e.g., 192.168.1.0/24)", network.Config.Subnet)
		}

		return fmt.Errorf("container network creation failed: %s", outputStr)
	}

	cp.logger.Info("Container network created successfully",
		zap.String("network", networkName),
		zap.String("subnet", network.Config.Subnet))
	return nil
}

// createNodeContainer creates a container for a node and returns the container ID
func (cp *ContainerProvisioner) createNodeContainer(containerName, image, networkName string, node *models.Node) (string, error) {
	args := []string{
		"run", "-d",
		"--name", containerName,
		"--hostname", node.Name,
		"--network", networkName,
	}

	// Add node-specific configuration
	if image == "frrouting/frr:latest" {
		// FRRouting router - needs full network access for routing protocols and policy enforcement
		args = append(args, "--privileged")
		args = append(args, "--cap-add=NET_ADMIN")
		args = append(args, "--cap-add=SYS_ADMIN")
		args = append(args, image)
	} else if image == "ubuntu:20.04" {
		// Ubuntu server - needs network admin for policy enforcement
		args = append(args, "--cap-add=NET_ADMIN")
		args = append(args, "--cap-add=NET_RAW")
		args = append(args, image)
		args = append(args, "bash", "-c",
			"apt-get update >/dev/null 2>&1 && "+
				"apt-get install -y openssh-server curl net-tools iproute2 >/dev/null 2>&1 && "+
				"mkdir -p /run/sshd && "+
				"echo 'root:netorchestrator' | chpasswd && "+
				"sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && "+
				"service ssh start >/dev/null 2>&1 && "+
				"sleep infinity")
	} else {
		// Alpine Linux (default) - lightweight but policy-capable
		args = append(args, "--cap-add=NET_ADMIN")
		args = append(args, image)
		args = append(args, "sh", "-c",
			"apk add --no-cache openssh-server curl net-tools iproute2 && "+
				"ssh-keygen -A && "+
				"echo 'root:netorchestrator' | chpasswd && "+
				"sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && "+
				"echo 'PasswordAuthentication yes' >> /etc/ssh/sshd_config && "+
				"mkdir -p /run/sshd && "+
				"/usr/sbin/sshd -D")
	}

	cmd := exec.Command("podman", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		observability.GetMetrics().RecordPodmanError("container_create")
		return "", fmt.Errorf("container creation failed: %s", output)
	}

	containerID := strings.TrimSpace(string(output))
	cp.logger.Info("Node container created",
		zap.String("container_name", containerName),
		zap.String("container_id", containerID[:12]), // Log truncated for readability
		zap.String("full_container_id", containerID), // Log full ID for debugging
		zap.String("image", image))

	return containerID, nil // Return full container ID
}

// DestroyNodeContainer removes the container for a node
func (cp *ContainerProvisioner) DestroyNodeContainer(node *models.Node) error {
	metrics := observability.GetMetrics()

	if node.Config.CustomAttrs == nil {
		return nil // No container to destroy
	}

	containerName, exists := node.Config.CustomAttrs["container_name"]
	if !exists {
		return nil // No container name recorded
	}

	// Remove container
	cmd := exec.Command("podman", "rm", "-f", containerName)
	if err := cmd.Run(); err != nil {
		cp.logger.Warn("Failed to remove container",
			zap.String("container", containerName),
			zap.Error(err))
		metrics.RecordContainerOperation("delete", err)
		metrics.RecordPodmanError("container_delete")
		return err
	}

	cp.logger.Info("Node container destroyed",
		zap.String("container", containerName))

	metrics.RecordContainerOperation("delete", nil)
	return nil
}

// GetNodeContainerInfo retrieves information about a node's container
func (cp *ContainerProvisioner) GetNodeContainerInfo(node *models.Node) (map[string]interface{}, error) {
	if node.Config.CustomAttrs == nil {
		return nil, fmt.Errorf("no container provisioned for this node")
	}

	containerName, exists := node.Config.CustomAttrs["container_name"]
	if !exists {
		return nil, fmt.Errorf("no container name recorded for this node")
	}

	// Get container status
	cmd := exec.Command("podman", "ps", "--format", "{{.Names}}\t{{.Status}}", "--filter", fmt.Sprintf("name=%v", containerName))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to check container status: %w", err)
	}

	status := "stopped"
	if strings.Contains(string(output), containerName) {
		status = "running"
	}

	info := map[string]interface{}{
		"container_name":    containerName,
		"container_image":   node.Config.CustomAttrs["container_image"],
		"container_network": node.Config.CustomAttrs["container_network"],
		"provisioned_at":    node.Config.CustomAttrs["provisioned_at"],
		"status":            status,
		"node_id":           node.ID.String(),
		"node_name":         node.Name,
	}

	return info, nil
}
