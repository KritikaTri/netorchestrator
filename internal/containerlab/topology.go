package containerlab

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

// LabTopology represents a Containerlab-style network topology
type LabTopology struct {
	Name       string             `yaml:"name"`
	Management ManagementConfig   `yaml:"mgmt"`
	Topology   TopologyDefinition `yaml:"topology"`
}

// ManagementConfig defines management network settings
type ManagementConfig struct {
	Network    string `yaml:"network"`
	IPv4Subnet string `yaml:"ipv4-subnet"`
}

// TopologyDefinition defines the network topology
type TopologyDefinition struct {
	Nodes map[string]NodeConfig `yaml:"nodes"`
	Links []LinkConfig          `yaml:"links"`
}

// NodeConfig represents a network node configuration
type NodeConfig struct {
	Kind     string            `yaml:"kind"`
	Image    string            `yaml:"image"`
	MgmtIPv4 string            `yaml:"mgmt-ipv4"`
	Ports    []string          `yaml:"ports,omitempty"`
	Env      map[string]string `yaml:"env,omitempty"`
	Binds    []string          `yaml:"binds,omitempty"`
	Exec     []string          `yaml:"exec,omitempty"`
	Labels   map[string]string `yaml:"labels,omitempty"`
}

// LinkConfig represents a network link between nodes
type LinkConfig struct {
	Endpoints []string `yaml:"endpoints"`
}

// LabManager manages containerlab-style topologies using Podman
type LabManager struct {
	logger      *zap.Logger
	topology    *LabTopology
	running     map[string]*RunningNode
	networkName string
}

// RunningNode represents an active network node
type RunningNode struct {
	Name        string            `json:"name"`
	Kind        string            `json:"kind"`
	Image       string            `json:"image"`
	ContainerID string            `json:"container_id"`
	MgmtIPv4    string            `json:"mgmt_ipv4"`
	Status      string            `json:"status"`
	Ports       map[string]string `json:"ports"`
	Labels      map[string]string `json:"labels"`
	StartedAt   time.Time         `json:"started_at"`
}

// NewLabManager creates a new Containerlab-style lab manager
func NewLabManager(logger *zap.Logger) *LabManager {
	return &LabManager{
		logger:      logger,
		running:     make(map[string]*RunningNode),
		networkName: "netorchestrator-lab",
	}
}

// LoadTopology loads a topology from a YAML file
func (lm *LabManager) LoadTopology(filePath string) error {
	// For now, create a default topology programmatically
	// In production, this would read from the YAML file

	lm.topology = &LabTopology{
		Name: "netorchestrator-lab",
		Management: ManagementConfig{
			Network:    "netorchestrator-lab",
			IPv4Subnet: "172.20.20.0/24",
		},
		Topology: TopologyDefinition{
			Nodes: map[string]NodeConfig{
				"frr-router": {
					Kind:     "linux",
					Image:    "frrouting/frr:latest",
					MgmtIPv4: "172.20.20.11",
					Ports:    []string{"22001:22", "22002:2601", "22003:2605"},
					Labels: map[string]string{
						"vendor":     "FRRouting",
						"type":       "router",
						"management": "ssh,telnet",
					},
				},
				"alpine-node1": {
					Kind:     "linux",
					Image:    "alpine:latest",
					MgmtIPv4: "172.20.20.12",
					Ports:    []string{"23001:22"},
					Labels: map[string]string{
						"vendor":     "Alpine",
						"type":       "host",
						"management": "ssh",
					},
				},
				"ubuntu-router": {
					Kind:     "linux",
					Image:    "ubuntu:20.04",
					MgmtIPv4: "172.20.20.13",
					Ports:    []string{"24001:22", "24002:80"},
					Labels: map[string]string{
						"vendor":     "Ubuntu",
						"type":       "router",
						"management": "ssh,http",
					},
				},
			},
		},
	}

	lm.logger.Info("Topology loaded", zap.String("name", lm.topology.Name))
	return nil
}

// Deploy deploys the network topology using Podman
func (lm *LabManager) Deploy() error {
	lm.logger.Info("Deploying network topology", zap.String("lab", lm.topology.Name))

	// Create network
	if err := lm.createNetwork(); err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	// Deploy each node
	for nodeName, nodeConfig := range lm.topology.Topology.Nodes {
		if err := lm.deployNode(nodeName, nodeConfig); err != nil {
			lm.logger.Error("Failed to deploy node",
				zap.String("node", nodeName),
				zap.Error(err))
			continue
		}
	}

	// Wait for containers to be ready
	time.Sleep(5 * time.Second)

	// Update running status
	lm.discoverRunningNodes()

	lm.logger.Info("Topology deployed successfully",
		zap.String("lab", lm.topology.Name),
		zap.Int("nodes", len(lm.running)))

	return nil
}

// createNetwork creates the management network
func (lm *LabManager) createNetwork() error {
	cmd := exec.Command("podman", "network", "create", lm.networkName)
	output, err := cmd.CombinedOutput()

	if err != nil && !strings.Contains(string(output), "already exists") {
		return fmt.Errorf("network creation failed: %s", output)
	}

	lm.logger.Info("Management network ready", zap.String("network", lm.networkName))
	return nil
}

// deployNode deploys a single network node
func (lm *LabManager) deployNode(nodeName string, config NodeConfig) error {
	args := []string{"run", "-d", "--name", nodeName, "--hostname", nodeName}

	// Add network
	args = append(args, "--network", lm.networkName)

	// Add ports
	for _, port := range config.Ports {
		args = append(args, "-p", port)
	}

	// Add environment variables
	for key, value := range config.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add volume binds
	for _, bind := range config.Binds {
		args = append(args, "-v", bind)
	}

	// Add image
	args = append(args, config.Image)

	// Add startup commands based on node type
	if config.Image == "alpine:latest" {
		startupCmd := `sh -c "
			apk add --no-cache openssh-server curl net-tools iproute2 && \
			ssh-keygen -A && \
			echo 'root:netorchestrator' | chpasswd && \
			sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
			echo 'PasswordAuthentication yes' >> /etc/ssh/sshd_config && \
			mkdir -p /run/sshd && \
			/usr/sbin/sshd -D
		"`
		args = append(args, startupCmd)
	} else if config.Image == "ubuntu:20.04" {
		startupCmd := `bash -c "
			apt-get update && \
			apt-get install -y openssh-server curl net-tools iproute2 bird2 nginx && \
			mkdir -p /run/sshd && \
			echo 'root:netorchestrator' | chpasswd && \
			sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
			echo 'PasswordAuthentication yes' >> /etc/ssh/sshd_config && \
			service ssh start && \
			service nginx start && \
			sleep infinity
		"`
		args = append(args, startupCmd)
	}

	// Execute podman run
	cmd := exec.Command("podman", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to deploy %s: %s", nodeName, output)
	}

	containerID := strings.TrimSpace(string(output))

	// Create running node record
	runningNode := &RunningNode{
		Name:        nodeName,
		Kind:        config.Kind,
		Image:       config.Image,
		ContainerID: containerID,
		MgmtIPv4:    config.MgmtIPv4,
		Status:      "starting",
		Labels:      config.Labels,
		StartedAt:   time.Now(),
		Ports:       make(map[string]string),
	}

	// Parse port mappings
	for _, port := range config.Ports {
		parts := strings.Split(port, ":")
		if len(parts) == 2 {
			runningNode.Ports[parts[1]] = parts[0]
		}
	}

	lm.running[nodeName] = runningNode

	lm.logger.Info("Node deployed",
		zap.String("node", nodeName),
		zap.String("image", config.Image),
		zap.String("container_id", containerID[:12]))

	return nil
}

// discoverRunningNodes updates the status of running nodes
func (lm *LabManager) discoverRunningNodes() {
	cmd := exec.Command("podman", "ps", "--format", "{{.Names}}\t{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		lm.logger.Error("Failed to discover running nodes", zap.Error(err))
		return
	}

	runningContainers := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			name := parts[0]
			status := parts[1]
			runningContainers[name] = status
		}
	}

	// Update node statuses
	for nodeName, node := range lm.running {
		if status, exists := runningContainers[nodeName]; exists {
			if strings.Contains(strings.ToLower(status), "up") {
				node.Status = "running"
			} else {
				node.Status = "stopped"
			}
		} else {
			node.Status = "not found"
		}
	}
}

// Destroy tears down the entire topology
func (lm *LabManager) Destroy() error {
	lm.logger.Info("Destroying network topology", zap.String("lab", lm.topology.Name))

	// Stop and remove all containers
	for nodeName := range lm.running {
		cmd := exec.Command("podman", "rm", "-f", nodeName)
		if err := cmd.Run(); err != nil {
			lm.logger.Warn("Failed to remove container",
				zap.String("container", nodeName),
				zap.Error(err))
		}
	}

	// Remove network
	cmd := exec.Command("podman", "network", "rm", lm.networkName)
	if err := cmd.Run(); err != nil {
		lm.logger.Warn("Failed to remove network",
			zap.String("network", lm.networkName),
			zap.Error(err))
	}

	lm.running = make(map[string]*RunningNode)
	lm.logger.Info("Topology destroyed")

	return nil
}

// Inspect returns information about the running topology
func (lm *LabManager) Inspect() map[string]interface{} {
	lm.discoverRunningNodes()

	return map[string]interface{}{
		"lab_name":      lm.topology.Name,
		"running_nodes": len(lm.running),
		"nodes":         lm.running,
		"network":       lm.networkName,
		"management":    lm.topology.Management,
		"status":        "running",
	}
}

// GetRunningNodes returns all running nodes
func (lm *LabManager) GetRunningNodes() map[string]*RunningNode {
	lm.discoverRunningNodes()
	return lm.running
}

// ExecInNode executes a command in a specific node
func (lm *LabManager) ExecInNode(nodeName, command string) (string, error) {
	_, exists := lm.running[nodeName]
	if !exists {
		return "", fmt.Errorf("node %s not found", nodeName)
	}

	cmd := exec.Command("podman", "exec", nodeName, "sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("command execution failed: %w", err)
	}

	result := strings.TrimSpace(string(output))

	lm.logger.Info("Command executed in node",
		zap.String("node", nodeName),
		zap.String("command", command),
		zap.Int("output_length", len(result)))

	return result, nil
}

// GetNodeInfo retrieves detailed information about a node
func (lm *LabManager) GetNodeInfo(nodeName string) (map[string]interface{}, error) {
	node, exists := lm.running[nodeName]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeName)
	}

	// Gather information via container exec
	hostname, _ := lm.ExecInNode(nodeName, "hostname")
	osRelease, _ := lm.ExecInNode(nodeName, "cat /etc/os-release 2>/dev/null | grep PRETTY_NAME | cut -d'\"' -f2 || echo 'Unknown OS'")
	uptime, _ := lm.ExecInNode(nodeName, "uptime")
	interfaces, _ := lm.ExecInNode(nodeName, "ip -br addr show")

	nodeInfo := map[string]interface{}{
		"name":         node.Name,
		"kind":         node.Kind,
		"image":        node.Image,
		"container_id": node.ContainerID[:12],
		"mgmt_ipv4":    node.MgmtIPv4,
		"status":       node.Status,
		"started_at":   node.StartedAt,
		"ports":        node.Ports,
		"labels":       node.Labels,
		"live_data": map[string]interface{}{
			"hostname":   hostname,
			"os":         osRelease,
			"uptime":     uptime,
			"interfaces": interfaces,
		},
	}

	return nodeInfo, nil
}

// GetTopologyGraph returns a graph representation of the topology
func (lm *LabManager) GetTopologyGraph() map[string]interface{} {
	nodes := make([]map[string]interface{}, 0)
	edges := make([]map[string]interface{}, 0)

	// Add nodes
	for nodeName, nodeData := range lm.running {
		nodeInfo := map[string]interface{}{
			"id":     nodeName,
			"label":  nodeName,
			"type":   nodeData.Kind,
			"vendor": nodeData.Labels["vendor"],
			"status": nodeData.Status,
			"image":  nodeData.Image,
		}
		nodes = append(nodes, nodeInfo)
	}

	// Add edges from topology links
	linkID := 0
	for _, link := range lm.topology.Topology.Links {
		if len(link.Endpoints) == 2 {
			// Parse endpoints like "frr-router:eth1"
			source := strings.Split(link.Endpoints[0], ":")[0]
			target := strings.Split(link.Endpoints[1], ":")[0]

			edge := map[string]interface{}{
				"id":     fmt.Sprintf("link_%d", linkID),
				"source": source,
				"target": target,
				"label":  fmt.Sprintf("%s ↔ %s", link.Endpoints[0], link.Endpoints[1]),
			}
			edges = append(edges, edge)
			linkID++
		}
	}

	return map[string]interface{}{
		"nodes":    nodes,
		"edges":    edges,
		"lab_name": lm.topology.Name,
		"network":  lm.networkName,
	}
}
