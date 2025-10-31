package containerlab

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ContainerlabHandlers handles HTTP requests for Containerlab topology management
type ContainerlabHandlers struct {
	labManager *LabManager
	logger     *zap.Logger
}

// NewContainerlabHandlers creates new Containerlab handlers
func NewContainerlabHandlers(logger *zap.Logger) *ContainerlabHandlers {
	labManager := NewLabManager(logger)

	// Load default topology
	if err := labManager.LoadTopology("netorchestrator-lab.clab.yml"); err != nil {
		logger.Warn("Failed to load topology file, using defaults", zap.Error(err))
	}

	return &ContainerlabHandlers{
		labManager: labManager,
		logger:     logger,
	}
}

// DeployLab deploys the network topology
// @Summary Deploy network lab topology
// @Description Deploy a Containerlab-style network topology using containers
// @Tags containerlab
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /lab/deploy [post]
func (h *ContainerlabHandlers) DeployLab(c *gin.Context) {
	if err := h.labManager.Deploy(); err != nil {
		h.logger.Error("Lab deployment failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Lab deployment failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":            "success",
		"message":           "Network lab deployed successfully",
		"lab_name":          h.labManager.topology.Name,
		"deployment_method": "Podman with Containerlab-style topology",
	})
}

// DestroyLab tears down the network topology
// @Summary Destroy network lab
// @Description Destroy the running network lab topology
// @Tags containerlab
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /lab/destroy [post]
func (h *ContainerlabHandlers) DestroyLab(c *gin.Context) {
	if err := h.labManager.Destroy(); err != nil {
		h.logger.Error("Lab destruction failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Lab destruction failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Network lab destroyed successfully",
	})
}

// InspectLab returns detailed information about the running lab
// @Summary Inspect network lab
// @Description Get detailed information about the running network lab
// @Tags containerlab
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /lab/inspect [get]
func (h *ContainerlabHandlers) InspectLab(c *gin.Context) {
	inspection := h.labManager.Inspect()

	c.JSON(http.StatusOK, gin.H{
		"lab_info":      inspection,
		"status":        "success",
		"message":       "Network lab inspection complete",
		"topology_type": "Containerlab-compatible",
	})
}

// ListNodes lists all nodes in the topology
// @Summary List lab nodes
// @Description Get a list of all nodes in the network lab
// @Tags containerlab
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /lab/nodes [get]
func (h *ContainerlabHandlers) ListNodes(c *gin.Context) {
	nodes := h.labManager.GetRunningNodes()

	nodeList := make([]interface{}, 0, len(nodes))
	for _, node := range nodes {
		nodeList = append(nodeList, node)
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes":    nodeList,
		"count":    len(nodes),
		"status":   "success",
		"message":  "Lab nodes retrieved from live containers",
		"lab_name": h.labManager.topology.Name,
	})
}

// GetNode retrieves detailed information about a specific node
// @Summary Get lab node details
// @Description Get detailed information about a specific node in the lab
// @Tags containerlab
// @Produce json
// @Param name path string true "Node name"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /lab/nodes/{name} [get]
func (h *ContainerlabHandlers) GetNode(c *gin.Context) {
	nodeName := c.Param("name")

	nodeInfo, err := h.labManager.GetNodeInfo(nodeName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Node not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"node_info": nodeInfo,
		"status":    "success",
		"message":   "Live node information retrieved from container",
	})
}

// ExecCommand executes a command on a specific node
// @Summary Execute command on node
// @Description Execute a network command on a specific lab node
// @Tags containerlab
// @Accept json
// @Produce json
// @Param name path string true "Node name"
// @Param request body map[string]string true "Command to execute"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /lab/nodes/{name}/exec [post]
func (h *ContainerlabHandlers) ExecCommand(c *gin.Context) {
	nodeName := c.Param("name")

	var req struct {
		Command string `json:"command" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	output, err := h.labManager.ExecInNode(nodeName, req.Command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Command execution failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"node":    nodeName,
		"command": req.Command,
		"output":  output,
		"message": "Command executed successfully on network container",
		"method":  "podman exec",
	})
}

// GetTopologyGraph returns a graph representation for visualization
// @Summary Get topology graph
// @Description Get network topology as a graph for visualization
// @Tags containerlab
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /lab/topology [get]
func (h *ContainerlabHandlers) GetTopologyGraph(c *gin.Context) {
	graph := h.labManager.GetTopologyGraph()

	c.JSON(http.StatusOK, gin.H{
		"topology_graph": graph,
		"status":         "success",
		"message":        "Network topology graph generated",
		"format":         "nodes-and-edges",
	})
}

// GetLabStatus returns the overall lab status
// @Summary Get lab status
// @Description Get the overall status of the network lab
// @Tags containerlab
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /lab/status [get]
func (h *ContainerlabHandlers) GetLabStatus(c *gin.Context) {
	nodes := h.labManager.GetRunningNodes()

	onlineCount := 0
	for _, node := range nodes {
		if node.Status == "running" {
			onlineCount++
		}
	}

	status := "healthy"
	if onlineCount == 0 {
		status = "down"
	} else if onlineCount < len(nodes) {
		status = "partial"
	}

	c.JSON(http.StatusOK, gin.H{
		"lab_status":         status,
		"total_nodes":        len(nodes),
		"online_nodes":       onlineCount,
		"lab_name":           h.labManager.topology.Name,
		"uptime":             "Active",
		"management_network": h.labManager.networkName,
		"status":             "success",
	})
}

// RegisterRoutes registers all Containerlab routes
func (h *ContainerlabHandlers) RegisterRoutes(router *gin.RouterGroup) {
	lab := router.Group("/lab")
	{
		lab.GET("/status", h.GetLabStatus)
		lab.GET("/inspect", h.InspectLab)
		lab.POST("/deploy", h.DeployLab)
		lab.POST("/destroy", h.DestroyLab)
		lab.GET("/topology", h.GetTopologyGraph)

		// Node management
		lab.GET("/nodes", h.ListNodes)
		lab.GET("/nodes/:name", h.GetNode)
		lab.POST("/nodes/:name/exec", h.ExecCommand)
	}
}
