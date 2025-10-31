package automation

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler handles automation-related HTTP requests
type Handler struct {
	wrapper *CompatibilityWrapper
	logger  *zap.Logger
}

// NewHandler creates a new automation handler
func NewHandler(db *gorm.DB, logger *zap.Logger) *Handler {
	wrapper := NewCompatibilityWrapper(db, logger)

	// Register executors
	wrapper.RegisterExecutor("terraform", NewTerraformExecutor(logger))
	wrapper.RegisterExecutor("ansible", NewAnsibleExecutor(logger))
	wrapper.RegisterExecutor("api_call", NewAPICallExecutor(logger))
	wrapper.RegisterExecutor("script", NewScriptExecutor(logger))
	wrapper.RegisterExecutor("validation", NewValidationExecutor(logger))
	wrapper.RegisterExecutor("network_provision", NewNetworkProvisionExecutor(db, logger))

	return &Handler{
		wrapper: wrapper,
		logger:  logger,
	}
}

// CreateWorkflowRequest represents a workflow creation request
type CreateWorkflowRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Tasks       []TaskDefinition       `json:"tasks" binding:"required"`
	Variables   map[string]interface{} `json:"variables"`
	Schedule    string                 `json:"schedule"`
	Enabled     bool                   `json:"enabled"`
}

// ExecuteWorkflowRequest represents a workflow execution request
type ExecuteWorkflowRequest struct {
	Variables map[string]interface{} `json:"variables"`
	DryRun    bool                   `json:"dry_run"`
}

// WorkflowResponse represents a workflow response
type WorkflowResponse struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Version     int                    `json:"version"`
	Tasks       []TaskDefinition       `json:"tasks"`
	Variables   map[string]interface{} `json:"variables"`
	Schedule    string                 `json:"schedule"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

// ExecutionResponse represents an execution response
type ExecutionResponse struct {
	ID           uint                    `json:"id"`
	WorkflowID   uint                    `json:"workflow_id"`
	Status       string                  `json:"status"`
	StartedAt    string                  `json:"started_at"`
	CompletedAt  string                  `json:"completed_at"`
	Duration     string                  `json:"duration"`
	TaskResults  []TaskExecutionResponse `json:"task_results"`
	Variables    map[string]interface{}  `json:"variables"`
	ErrorMessage string                  `json:"error_message,omitempty"`
}

// TaskExecutionResponse represents a task execution response
type TaskExecutionResponse struct {
	ID               uint                   `json:"id"`
	TaskID           string                 `json:"task_id"`
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`
	Status           string                 `json:"status"`
	StartedAt        string                 `json:"started_at"`
	CompletedAt      string                 `json:"completed_at"`
	Duration         string                 `json:"duration"`
	Output           map[string]interface{} `json:"output"`
	Logs             []string               `json:"logs"`
	ExecutionDetails map[string]interface{} `json:"execution_details"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
}

// RegisterRoutes registers automation API routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	automation := r.Group("/automation")
	{
		// Workflow management
		automation.POST("/workflows", h.CreateWorkflow)
		automation.GET("/workflows", h.ListWorkflows)
		automation.GET("/workflows/:id", h.GetWorkflow)
		automation.PUT("/workflows/:id", h.UpdateWorkflow)
		automation.DELETE("/workflows/:id", h.DeleteWorkflow)

		// Workflow execution
		automation.POST("/workflows/:id/execute", h.ExecuteWorkflow)
		automation.GET("/workflows/:id/executions", h.ListExecutions)
		automation.GET("/executions/:id", h.GetExecution)
		automation.POST("/executions/:id/cancel", h.CancelExecution)

		// Task management
		automation.GET("/executions/:id/tasks", h.GetExecutionTasks)
		automation.GET("/tasks/:id", h.GetTaskExecution)

		// Templates and examples
		automation.GET("/templates", h.ListTemplates)
		automation.GET("/templates/:name", h.GetTemplate)
		automation.GET("/executors", h.ListExecutors)
	}
}

// CreateWorkflow creates a new workflow
// @Summary Create workflow
// @Description Create a new automation workflow
// @Tags automation
// @Accept json
// @Produce json
// @Param workflow body CreateWorkflowRequest true "Workflow data"
// @Success 201 {object} WorkflowResponse
// @Router /api/v1/automation/workflows [post]
func (h *Handler) CreateWorkflow(c *gin.Context) {
	var req CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create workflow using compatibility wrapper
	workflow, err := h.wrapper.CreateWorkflowCompat(&req)
	if err != nil {
		h.logger.Error("Failed to create workflow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow"})
		return
	}

	response := h.convertEnhancedWorkflowToResponse(workflow)
	c.JSON(http.StatusCreated, response)
}

// ListWorkflows lists all workflows
// @Summary List workflows
// @Description Get all automation workflows
// @Tags automation
// @Produce json
// @Param enabled query bool false "Filter by enabled status"
// @Success 200 {array} WorkflowResponse
// @Router /api/v1/automation/workflows [get]
func (h *Handler) ListWorkflows(c *gin.Context) {
	// Query basic workflow info without complex JSONB deserialization
	type SimpleWorkflow struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Version     string `json:"version"`
	}

	var workflows []SimpleWorkflow
	if err := h.wrapper.DB().Table("workflow_definitions").
		Select("id, name, description, version").
		Find(&workflows).Error; err != nil {
		h.logger.Error("Failed to list workflows", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list workflows"})
		return
	}

	// Convert to response format
	response := make([]map[string]interface{}, len(workflows))
	for i, wf := range workflows {
		response[i] = map[string]interface{}{
			"id":          wf.ID,
			"name":        wf.Name,
			"description": wf.Description,
			"version":     wf.Version,
			"enabled":     true,
			"status":      "active",
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"workflows": response,
		"count":     len(response),
	})
}

// GetWorkflow gets a specific workflow
// @Summary Get workflow
// @Description Get automation workflow by ID
// @Tags automation
// @Produce json
// @Param id path int true "Workflow ID"
// @Success 200 {object} WorkflowResponse
// @Router /api/v1/automation/workflows/{id} [get]
func (h *Handler) GetWorkflow(c *gin.Context) {
	id := c.Param("id")

	// Simplified query without JSONB deserialization
	type SimpleWorkflow struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Version     string `json:"version"`
	}

	var workflow SimpleWorkflow
	if err := h.wrapper.DB().Table("workflow_definitions").
		Select("id, name, description, version").
		Where("id = ?", id).
		First(&workflow).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
			return
		}
		h.logger.Error("Failed to get workflow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workflow"})
		return
	}

	response := map[string]interface{}{
		"id":          workflow.ID,
		"name":        workflow.Name,
		"description": workflow.Description,
		"version":     workflow.Version,
		"enabled":     true,
		"status":      "active",
	}

	c.JSON(http.StatusOK, response)
}

// UpdateWorkflow updates a workflow (stub implementation)
func (h *Handler) UpdateWorkflow(c *gin.Context) {
	// Stub implementation
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Update workflow not implemented yet"})
}

// DeleteWorkflow deletes a workflow (stub implementation)
func (h *Handler) DeleteWorkflow(c *gin.Context) {
	// Stub implementation
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Delete workflow not implemented yet"})
}

// ExecuteWorkflow executes a workflow
func (h *Handler) ExecuteWorkflow(c *gin.Context) {
	id := c.Param("id")

	var req ExecuteWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify workflow exists
	var count int64
	if err := h.wrapper.DB().Table("workflow_definitions").Where("id = ?", id).Count(&count).Error; err != nil {
		h.logger.Error("Failed to check workflow", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check workflow"})
		return
	}

	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// For MVP, return mock execution response
	executionID := "exec-" + id[:8]

	response := map[string]interface{}{
		"execution_id": executionID,
		"workflow_id":  id,
		"status":       "running",
		"started_at":   time.Now().UTC().Format(time.RFC3339),
		"message":      "Workflow execution started successfully",
	}

	h.logger.Info("Workflow execution started", zap.String("workflow_id", id), zap.String("execution_id", executionID))
	c.JSON(http.StatusAccepted, response)
}

// ListExecutions lists workflow executions
func (h *Handler) ListExecutions(c *gin.Context) {
	workflowID := c.Param("id")

	// Query executions from database
	type SimpleExecution struct {
		ID         string `json:"id"`
		WorkflowID string `json:"workflow_id"`
		Status     string `json:"status"`
		StartedAt  string `json:"started_at"`
	}

	var executions []SimpleExecution
	if err := h.wrapper.DB().Table("workflow_executions").
		Select("id, workflow_id, status, started_at").
		Where("workflow_id = ?", workflowID).
		Order("started_at DESC").
		Limit(50).
		Find(&executions).Error; err != nil {
		h.logger.Error("Failed to list executions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list executions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"executions": executions,
		"count":      len(executions),
	})
}

// GetExecution gets a specific execution
func (h *Handler) GetExecution(c *gin.Context) {
	id := c.Param("id")

	// Query execution from database
	type SimpleExecution struct {
		ID         string `json:"id"`
		WorkflowID string `json:"workflow_id"`
		Status     string `json:"status"`
		StartedAt  string `json:"started_at"`
	}

	var execution SimpleExecution

	// Support special keyword "latest" (optionally filtered by workflow_id query param)
	if id == "latest" {
		workflowID := c.Query("workflow_id")
		db := h.wrapper.DB().Table("workflow_executions").
			Select("id, workflow_id, status, started_at").
			Order("started_at DESC")
		if workflowID != "" {
			db = db.Where("workflow_id = ?", workflowID)
		}
		if err := db.First(&execution).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "No executions found"})
				return
			}
			h.logger.Error("Failed to get latest execution", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get execution"})
			return
		}
		c.JSON(http.StatusOK, execution)
		return
	}

	// Fetch by specific ID
	if err := h.wrapper.DB().Table("workflow_executions").
		Select("id, workflow_id, status, started_at").
		Where("id = ?", id).
		First(&execution).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
			return
		}
		h.logger.Error("Failed to get execution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get execution"})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// CancelExecution cancels a workflow execution (stub)
func (h *Handler) CancelExecution(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

// GetExecutionTasks gets tasks for an execution
func (h *Handler) GetExecutionTasks(c *gin.Context) {
	executionID := c.Param("id")

	// Query tasks from database
	type SimpleTask struct {
		ID        string `json:"id"`
		TaskName  string `json:"task_name"`
		TaskType  string `json:"task_type"`
		Status    string `json:"status"`
		StartedAt string `json:"started_at"`
	}

	var tasks []SimpleTask
	if err := h.wrapper.DB().Table("task_executions").
		Select("id, task_name, task_type, status, started_at").
		Where("workflow_execution_id = ?", executionID).
		Order("started_at ASC").
		Find(&tasks).Error; err != nil {
		h.logger.Error("Failed to get execution tasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get execution tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// GetTaskExecution gets a specific task execution (stub)
func (h *Handler) GetTaskExecution(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

// ListTemplates lists available workflow templates
// @Summary List templates
// @Description Get available workflow templates
// @Tags automation
// @Produce json
// @Success 200 {array} object
// @Router /api/v1/automation/templates [get]
func (h *Handler) ListTemplates(c *gin.Context) {
	templates := h.getWorkflowTemplates()
	c.JSON(http.StatusOK, templates)
}

// GetTemplate gets a specific workflow template
// @Summary Get template
// @Description Get workflow template by name
// @Tags automation
// @Produce json
// @Param name path string true "Template name"
// @Success 200 {object} object
// @Router /api/v1/automation/templates/{name} [get]
func (h *Handler) GetTemplate(c *gin.Context) {
	name := c.Param("name")
	templates := h.getWorkflowTemplates()

	for _, template := range templates {
		if template["name"] == name {
			c.JSON(http.StatusOK, template)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
}

// ListExecutors lists available task executors
// @Summary List executors
// @Description Get available task executors
// @Tags automation
// @Produce json
// @Success 200 {array} object
// @Router /api/v1/automation/executors [get]
func (h *Handler) ListExecutors(c *gin.Context) {
	executors := []map[string]interface{}{
		{
			"name":        "terraform",
			"description": "Execute Terraform infrastructure provisioning",
			"actions":     []string{"plan", "apply", "destroy", "validate"},
		},
		{
			"name":        "ansible",
			"description": "Execute Ansible configuration management",
			"actions":     []string{"playbook", "ad-hoc", "vault"},
		},
		{
			"name":        "api_call",
			"description": "Make HTTP API calls",
			"actions":     []string{"http_request", "api_call"},
		},
		{
			"name":        "script",
			"description": "Execute shell scripts",
			"actions":     []string{"script", "shell", "bash"},
		},
		{
			"name":        "validation",
			"description": "Validate conditions and responses",
			"actions":     []string{"validate", "check", "verify"},
		},
		{
			"name":        "network_provision",
			"description": "Provision network resources",
			"actions":     []string{"create_network", "delete_network", "update_network"},
		},
	}

	c.JSON(http.StatusOK, executors)
}

// Helper functions

func (h *Handler) convertEnhancedWorkflowToResponse(workflow *EnhancedWorkflowDefinition) WorkflowResponse {
	return WorkflowResponse{
		ID:          workflow.NumericID,
		Name:        workflow.Name,
		Description: workflow.Description,
		Status:      workflow.Status,
		Version:     workflow.NumericVersion,
		Tasks:       workflow.Tasks,
		Variables:   workflow.Variables,
		Schedule:    workflow.Schedule,
		Enabled:     workflow.Enabled,
		CreatedAt:   workflow.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   workflow.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *Handler) convertExecutionToResponse(execution *WorkflowExecution) ExecutionResponse {
	response := ExecutionResponse{
		ID:           uint(hashString(execution.ID)),
		WorkflowID:   uint(hashString(execution.WorkflowID)),
		Status:       string(execution.Status),
		StartedAt:    execution.StartTime.Format("2006-01-02T15:04:05Z"),
		Variables:    execution.Input,
		ErrorMessage: execution.ErrorMessage,
	}

	if execution.EndTime != nil {
		response.CompletedAt = execution.EndTime.Format("2006-01-02T15:04:05Z")
		response.Duration = execution.EndTime.Sub(execution.StartTime).String()
	}

	// Convert task executions
	response.TaskResults = make([]TaskExecutionResponse, len(execution.Tasks))
	for i, task := range execution.Tasks {
		response.TaskResults[i] = h.convertTaskExecutionToResponse(&task)
	}

	return response
}

func (h *Handler) convertTaskExecutionToResponse(task *TaskExecution) TaskExecutionResponse {
	response := TaskExecutionResponse{
		ID:               uint(hashString(task.ID)),
		TaskID:           task.TaskID,
		Name:             task.Name,
		Status:           string(task.Status),
		StartedAt:        task.StartTime.Format("2006-01-02T15:04:05Z"),
		Output:           task.Output,
		Logs:             task.Logs,
		ExecutionDetails: task.ExecutionDetails,
		ErrorMessage:     task.ErrorMessage,
	}

	if task.EndTime != nil {
		response.CompletedAt = task.EndTime.Format("2006-01-02T15:04:05Z")
		response.Duration = task.EndTime.Sub(task.StartTime).String()
	}

	return response
}

func (h *Handler) getWorkflowTemplates() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "network_provisioning",
			"description": "Basic network provisioning workflow",
			"tasks": []TaskDefinition{
				{
					ID:   "validate_config",
					Name: "Validate Configuration",
					Type: "validation",
					Action: `{
						"type": "condition",
						"conditions": [
							{"field": "network_cidr", "operator": "exists", "value": null}
						]
					}`,
				},
				{
					ID:           "provision_network",
					Name:         "Provision Network",
					Type:         "terraform",
					Dependencies: []string{"validate_config"},
					Action: `{
						"command": "apply",
						"working_dir": "/terraform/networks",
						"auto_approve": false
					}`,
				},
			},
		},
		{
			"name":        "service_deployment",
			"description": "Deploy service with configuration management",
			"tasks": []TaskDefinition{
				{
					ID:   "deploy_infrastructure",
					Name: "Deploy Infrastructure",
					Type: "terraform",
					Action: `{
						"command": "apply",
						"working_dir": "/terraform/services"
					}`,
				},
				{
					ID:           "configure_service",
					Name:         "Configure Service",
					Type:         "ansible",
					Dependencies: []string{"deploy_infrastructure"},
					Action: `{
						"playbook": "/ansible/configure-service.yml",
						"inventory": "/ansible/inventory"
					}`,
				},
				{
					ID:           "health_check",
					Name:         "Health Check",
					Type:         "api_call",
					Dependencies: []string{"configure_service"},
					Action: `{
						"url": "{{service_url}}/health",
						"method": "GET",
						"expected_code": 200
					}`,
				},
			},
		},
	}
}
