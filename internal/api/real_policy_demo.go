package api

import (
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RealPolicyHandlers demonstrates actual policy enforcement on privileged containers
type RealPolicyHandlers struct {
	logger *zap.Logger
}

// NewRealPolicyHandlers creates handlers for real policy enforcement demo
func NewRealPolicyHandlers(logger *zap.Logger) *RealPolicyHandlers {
	return &RealPolicyHandlers{
		logger: logger,
	}
}

// ApplyFirewallPolicy applies real iptables rules to a container
// @Summary Apply real firewall policy
// @Description Apply actual iptables firewall rules to a privileged container
// @Tags Demo
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Firewall policy configuration"
// @Success 200 {object} map[string]interface{}
// @Router /demo/firewall [post]
func (rph *RealPolicyHandlers) ApplyFirewallPolicy(c *gin.Context) {
	var req struct {
		ContainerName string                   `json:"container_name" binding:"required"`
		Rules         []map[string]interface{} `json:"rules" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	rph.logger.Info("Applying real firewall policy",
		zap.String("container", req.ContainerName),
		zap.Int("rule_count", len(req.Rules)))

	// Clear existing NetOrchestrator rules
	cmd := exec.Command("podman", "exec", req.ContainerName, "sh", "-c",
		"iptables -F INPUT; iptables -P INPUT ACCEPT")
	if err := cmd.Run(); err != nil {
		rph.logger.Error("Failed to clear existing rules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear existing firewall rules",
		})
		return
	}

	appliedRules := []map[string]interface{}{}

	// Apply each rule
	for i, rule := range req.Rules {
		ruleCmd := rph.buildIptablesRule(rule)

		cmd := exec.Command("podman", "exec", req.ContainerName, "sh", "-c", ruleCmd)
		if err := cmd.Run(); err != nil {
			rph.logger.Warn("Failed to apply firewall rule",
				zap.Int("rule_index", i), zap.Error(err))
			continue
		}

		appliedRules = append(appliedRules, map[string]interface{}{
			"rule_index": i + 1,
			"command":    ruleCmd,
			"status":     "applied",
		})
	}

	// Get current iptables status
	cmd = exec.Command("podman", "exec", req.ContainerName, "iptables", "-L", "INPUT", "-n", "--line-numbers")
	output, err := cmd.Output()
	if err != nil {
		rph.logger.Error("Failed to get iptables status", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"status":             "success",
		"container":          req.ContainerName,
		"applied_rules":      appliedRules,
		"iptables_output":    string(output),
		"message":            "Real firewall policy enforced on container",
		"enforcement_method": "iptables",
	})
}

// ApplyQoSPolicy applies real traffic control rules to a container
// @Summary Apply real QoS policy
// @Description Apply actual tc (traffic control) QoS rules to a privileged container
// @Tags Demo
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "QoS policy configuration"
// @Success 200 {object} map[string]interface{}
// @Router /demo/qos [post]
func (rph *RealPolicyHandlers) ApplyQoSPolicy(c *gin.Context) {
	var req struct {
		ContainerName string                   `json:"container_name" binding:"required"`
		Classes       []map[string]interface{} `json:"classes" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	rph.logger.Info("Applying real QoS policy",
		zap.String("container", req.ContainerName),
		zap.Int("class_count", len(req.Classes)))

	// Clear existing QoS rules
	cmd := exec.Command("podman", "exec", req.ContainerName, "tc", "qdisc", "del", "dev", "eth0", "root")
	cmd.Run() // Ignore errors - qdisc might not exist

	// Create HTB root qdisc
	cmd = exec.Command("podman", "exec", req.ContainerName, "tc", "qdisc", "add", "dev", "eth0", "root", "handle", "1:", "htb", "default", "30")
	if err := cmd.Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create HTB qdisc",
			"details": err.Error(),
		})
		return
	}

	appliedClasses := []map[string]interface{}{}

	// Apply each traffic class
	for i, class := range req.Classes {
		classID := class["classid"].(string)
		rate := class["rate"].(string)

		cmd := exec.Command("podman", "exec", req.ContainerName, "tc", "class", "add", "dev", "eth0", "parent", "1:", "classid", classID, "htb", "rate", rate)
		if err := cmd.Run(); err != nil {
			rph.logger.Warn("Failed to apply QoS class",
				zap.Int("class_index", i), zap.Error(err))
			continue
		}

		appliedClasses = append(appliedClasses, map[string]interface{}{
			"class_id": classID,
			"rate":     rate,
			"status":   "applied",
		})
	}

	// Get current tc status
	cmd = exec.Command("podman", "exec", req.ContainerName, "tc", "qdisc", "show", "dev", "eth0")
	qdiscOutput, _ := cmd.Output()

	cmd = exec.Command("podman", "exec", req.ContainerName, "tc", "class", "show", "dev", "eth0")
	classOutput, _ := cmd.Output()

	c.JSON(http.StatusOK, gin.H{
		"status":             "success",
		"container":          req.ContainerName,
		"applied_classes":    appliedClasses,
		"qdisc_output":       string(qdiscOutput),
		"class_output":       string(classOutput),
		"message":            "Real QoS policy enforced on container",
		"enforcement_method": "tc (traffic control)",
	})
}

// GetContainerPolicyStatus shows real policy enforcement status
// @Summary Get container policy status
// @Description Get actual policy enforcement status from a container
// @Tags Demo
// @Produce json
// @Param container_name path string true "Container name"
// @Success 200 {object} map[string]interface{}
// @Router /demo/containers/{container_name}/policy-status [get]
func (rph *RealPolicyHandlers) GetContainerPolicyStatus(c *gin.Context) {
	containerName := c.Param("container_name")

	if containerName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Container name is required",
		})
		return
	}

	// Get iptables status
	iptablesCmd := exec.Command("podman", "exec", containerName, "iptables", "-L", "-n", "--line-numbers")
	iptablesOutput, iptablesErr := iptablesCmd.Output()

	// Get tc status
	tcQdiscCmd := exec.Command("podman", "exec", containerName, "tc", "qdisc", "show")
	tcQdiscOutput, tcQdiscErr := tcQdiscCmd.Output()

	tcClassCmd := exec.Command("podman", "exec", containerName, "tc", "class", "show")
	tcClassOutput, tcClassErr := tcClassCmd.Output()

	// Count NetOrchestrator rules
	netOrchestratorRules := strings.Count(string(iptablesOutput), "NetOrchestrator")

	status := map[string]interface{}{
		"container": containerName,
		"firewall": map[string]interface{}{
			"status":                "unknown",
			"rules_count":           0,
			"netorchestrator_rules": netOrchestratorRules,
		},
		"qos": map[string]interface{}{
			"status":        "unknown",
			"classes_count": 0,
		},
	}

	if iptablesErr == nil {
		status["firewall"].(map[string]interface{})["status"] = "available"
		status["firewall"].(map[string]interface{})["rules_count"] = strings.Count(string(iptablesOutput), "\n")
		status["firewall"].(map[string]interface{})["output"] = string(iptablesOutput)
	}

	if tcQdiscErr == nil {
		status["qos"].(map[string]interface{})["status"] = "available"
		status["qos"].(map[string]interface{})["qdisc_output"] = string(tcQdiscOutput)
	}

	if tcClassErr == nil {
		status["qos"].(map[string]interface{})["classes_count"] = strings.Count(string(tcClassOutput), "class htb")
		status["qos"].(map[string]interface{})["class_output"] = string(tcClassOutput)
	}

	c.JSON(http.StatusOK, gin.H{
		"policy_status": status,
		"message":       "Real container policy enforcement status",
		"timestamp":     "live_query_result",
	})
}

// buildIptablesRule builds an iptables command from rule configuration
func (rph *RealPolicyHandlers) buildIptablesRule(rule map[string]interface{}) string {
	var cmd strings.Builder
	cmd.WriteString("iptables -A INPUT")

	// Add protocol
	if protocol, exists := rule["protocol"]; exists {
		cmd.WriteString(" -p ")
		cmd.WriteString(protocol.(string))
	}

	// Add port
	if port, exists := rule["port"]; exists {
		cmd.WriteString(" --dport ")
		cmd.WriteString(port.(string))
	}

	// Add source
	if source, exists := rule["source"]; exists {
		cmd.WriteString(" -s ")
		cmd.WriteString(source.(string))
	}

	// Add action
	action := "ACCEPT" // Default
	if ruleAction, exists := rule["action"]; exists {
		action = strings.ToUpper(ruleAction.(string))
	}
	cmd.WriteString(" -j ")
	cmd.WriteString(action)

	// Add comment
	if name, exists := rule["name"]; exists {
		cmd.WriteString(" -m comment --comment \"NetOrchestrator: ")
		cmd.WriteString(name.(string))
		cmd.WriteString("\"")
	}

	return cmd.String()
}

// RegisterRealPolicyRoutes registers demo routes for real policy enforcement
func (rph *RealPolicyHandlers) RegisterRoutes(router *gin.RouterGroup) {
	demo := router.Group("/demo")
	{
		demo.POST("/firewall", rph.ApplyFirewallPolicy)
		demo.POST("/qos", rph.ApplyQoSPolicy)
		demo.GET("/containers/:container_name/policy-status", rph.GetContainerPolicyStatus)
	}
}
