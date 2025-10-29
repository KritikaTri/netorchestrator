package automation

import (
	"strconv"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CompatibilityWrapper provides compatibility between different API designs
type CompatibilityWrapper struct {
	engine *AutomationEngine
	logger *zap.Logger
}

// NewCompatibilityWrapper creates a wrapper around the automation engine
func NewCompatibilityWrapper(db *gorm.DB, logger *zap.Logger) *CompatibilityWrapper {
	engine := NewAutomationEngine(db, logger)
	return &CompatibilityWrapper{
		engine: engine,
		logger: logger,
	}
}

// EnhancedWorkflowDefinition extends the basic definition with additional fields
type EnhancedWorkflowDefinition struct {
	WorkflowDefinition
	Schedule    string                 `json:"schedule"`
	Enabled     bool                   `json:"enabled"`
	Variables   map[string]interface{} `json:"variables"`
	Status      string                 `json:"status"`
	NumericID   uint                   `json:"numeric_id"`
	NumericVersion int                 `json:"numeric_version"`
}

// CreateWorkflowCompat creates a workflow with compatibility layer
func (cw *CompatibilityWrapper) CreateWorkflowCompat(req *CreateWorkflowRequest) (*EnhancedWorkflowDefinition, error) {
	// Generate UUID for ID
	workflowID := uuid.New().String()
	
	workflow := &WorkflowDefinition{
		ID:          workflowID,
		Name:        req.Name,
		Description: req.Description,
		Tasks:       req.Tasks,
		Parameters:  req.Variables,
		Version:     "1",
	}

	// Create in database
	if err := cw.engine.db.Create(workflow).Error; err != nil {
		return nil, err
	}

	// Convert to enhanced version
	enhanced := &EnhancedWorkflowDefinition{
		WorkflowDefinition: *workflow,
		Schedule:           req.Schedule,
		Enabled:            req.Enabled,
		Variables:          req.Variables,
		Status:             "active",
		NumericID:          uint(hashString(workflowID)),
		NumericVersion:     1,
	}

	return enhanced, nil
}

// ExecuteWorkflowCompat executes a workflow with compatibility
func (cw *CompatibilityWrapper) ExecuteWorkflowCompat(workflowID string, variables map[string]interface{}) (*WorkflowExecution, error) {
	return cw.engine.ExecuteWorkflow(workflowID, variables, "system")
}

// GetWorkflowCompat gets a workflow by ID
func (cw *CompatibilityWrapper) GetWorkflowCompat(id string) (*EnhancedWorkflowDefinition, error) {
	var workflow WorkflowDefinition
	if err := cw.engine.db.First(&workflow, "id = ?", id).Error; err != nil {
		return nil, err
	}

	version, _ := strconv.Atoi(workflow.Version)
	enhanced := &EnhancedWorkflowDefinition{
		WorkflowDefinition: workflow,
		Variables:          workflow.Parameters,
		Status:             "active",
		NumericID:          uint(hashString(workflow.ID)),
		NumericVersion:     version,
		Enabled:            true,
	}

	return enhanced, nil
}

// ListWorkflowsCompat lists all workflows
func (cw *CompatibilityWrapper) ListWorkflowsCompat() ([]EnhancedWorkflowDefinition, error) {
	var workflows []WorkflowDefinition
	if err := cw.engine.db.Find(&workflows).Error; err != nil {
		return nil, err
	}

	enhanced := make([]EnhancedWorkflowDefinition, len(workflows))
	for i, workflow := range workflows {
		version, _ := strconv.Atoi(workflow.Version)
		enhanced[i] = EnhancedWorkflowDefinition{
			WorkflowDefinition: workflow,
			Variables:          workflow.Parameters,
			Status:             "active",
			NumericID:          uint(hashString(workflow.ID)),
			NumericVersion:     version,
			Enabled:            true,
		}
	}

	return enhanced, nil
}

// RegisterExecutor registers a task executor
func (cw *CompatibilityWrapper) RegisterExecutor(taskType string, executor TaskExecutor) {
	cw.engine.RegisterTaskExecutor(taskType, executor)
}

// DB returns the database connection
func (cw *CompatibilityWrapper) DB() *gorm.DB {
	return cw.engine.db
}

// hashString creates a simple hash from a string to generate numeric ID
func hashString(s string) uint64 {
	var hash uint64 = 5381
	for _, c := range s {
		hash = ((hash << 5) + hash) + uint64(c)
	}
	return hash & 0xFFFFFFFF // Keep it within uint32 range
}