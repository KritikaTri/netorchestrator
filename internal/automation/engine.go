package automation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WorkflowStatus represents the status of a workflow execution
type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
)

// TaskStatus represents the status of an individual task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusSkipped   TaskStatus = "skipped"
)

// WorkflowDefinition represents a workflow template
type WorkflowDefinition struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Tasks       []TaskDefinition       `json:"tasks" gorm:"serializer:json"`
	Parameters  map[string]interface{} `json:"parameters" gorm:"serializer:json"`
	Triggers    []TriggerDefinition    `json:"triggers" gorm:"serializer:json"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TaskDefinition represents a single task in a workflow
type TaskDefinition struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"` // terraform, ansible, api_call, script, validation
	Action       string                 `json:"action"`
	Parameters   map[string]interface{} `json:"parameters"`
	Dependencies []string               `json:"dependencies"`
	Timeout      time.Duration          `json:"timeout"`
	RetryCount   int                    `json:"retry_count"`
	OnFailure    string                 `json:"on_failure"` // continue, fail, retry
	Conditions   []TaskCondition        `json:"conditions"`
}

// TaskCondition represents a condition for task execution
type TaskCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, contains
	Value    interface{} `json:"value"`
}

// TriggerDefinition represents workflow triggers
type TriggerDefinition struct {
	Type       string                 `json:"type"` // schedule, event, webhook, manual
	Parameters map[string]interface{} `json:"parameters"`
	Enabled    bool                   `json:"enabled"`
}

// WorkflowExecution represents a running workflow instance
type WorkflowExecution struct {
	ID           string                 `json:"id" gorm:"primaryKey"`
	WorkflowID   string                 `json:"workflow_id"`
	Workflow     WorkflowDefinition     `json:"workflow" gorm:"foreignKey:WorkflowID"`
	Status       WorkflowStatus         `json:"status"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	Input        map[string]interface{} `json:"input" gorm:"serializer:json"`
	Output       map[string]interface{} `json:"output" gorm:"serializer:json"`
	Context      map[string]interface{} `json:"context" gorm:"serializer:json"`
	Tasks        []TaskExecution        `json:"tasks"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	CreatedBy    string                 `json:"created_by"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// TaskExecution represents a running task instance
type TaskExecution struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	WorkflowID       string                 `json:"workflow_id"`
	TaskID           string                 `json:"task_id"`
	Name             string                 `json:"name"`
	Status           TaskStatus             `json:"status"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          *time.Time             `json:"end_time,omitempty"`
	Input            map[string]interface{} `json:"input" gorm:"serializer:json"`
	Output           map[string]interface{} `json:"output" gorm:"serializer:json"`
	Logs             []string               `json:"logs" gorm:"serializer:json"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	RetryCount       int                    `json:"retry_count"`
	ExecutionDetails map[string]interface{} `json:"execution_details" gorm:"serializer:json"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// AutomationEngine manages workflow execution and task orchestration
type AutomationEngine struct {
	db                *gorm.DB
	logger            *zap.Logger
	taskExecutors     map[string]TaskExecutor
	runningWorkflows  map[string]*WorkflowRunner
	// workflowScheduler will be added later
	mutex             sync.RWMutex
}

// TaskExecutor interface for different task types
type TaskExecutor interface {
	Execute(ctx context.Context, task *TaskExecution, definition *TaskDefinition) error
	Validate(definition *TaskDefinition) error
	GetSupportedActions() []string
}

// WorkflowRunner manages a single workflow execution
type WorkflowRunner struct {
	execution *WorkflowExecution
	engine    *AutomationEngine
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *zap.Logger
}

// NewAutomationEngine creates a new automation engine
func NewAutomationEngine(db *gorm.DB, logger *zap.Logger) *AutomationEngine {
	engine := &AutomationEngine{
		db:               db,
		logger:           logger,
		taskExecutors:    make(map[string]TaskExecutor),
		runningWorkflows: make(map[string]*WorkflowRunner),
	}

	// Register built-in task executors
	engine.RegisterTaskExecutor("terraform", NewTerraformExecutor(logger))
	engine.RegisterTaskExecutor("ansible", NewAnsibleExecutor(logger))
	engine.RegisterTaskExecutor("api_call", NewAPICallExecutor(logger))
	engine.RegisterTaskExecutor("script", NewScriptExecutor(logger))
	engine.RegisterTaskExecutor("validation", NewValidationExecutor(logger))
	engine.RegisterTaskExecutor("network_provision", NewNetworkProvisionExecutor(db, logger))

		// Scheduler will be initialized separately if needed
	
	return engine
}

// RegisterTaskExecutor registers a new task executor
func (ae *AutomationEngine) RegisterTaskExecutor(taskType string, executor TaskExecutor) {
	ae.taskExecutors[taskType] = executor
	ae.logger.Info("Registered task executor", zap.String("type", taskType))
}

// CreateWorkflow creates a new workflow definition
func (ae *AutomationEngine) CreateWorkflow(workflow *WorkflowDefinition) error {
	// Validate workflow
	if err := ae.validateWorkflow(workflow); err != nil {
		return fmt.Errorf("workflow validation failed: %w", err)
	}

	// Save to database
	if err := ae.db.Create(workflow).Error; err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	ae.logger.Info("Workflow created", zap.String("id", workflow.ID), zap.String("name", workflow.Name))
	return nil
}

// ExecuteWorkflow starts a new workflow execution
func (ae *AutomationEngine) ExecuteWorkflow(workflowID string, input map[string]interface{}, createdBy string) (*WorkflowExecution, error) {
	// Get workflow definition
	var workflow WorkflowDefinition
	if err := ae.db.First(&workflow, "id = ?", workflowID).Error; err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	// Create execution instance
	execution := &WorkflowExecution{
		ID:         generateID(),
		WorkflowID: workflowID,
		Status:     WorkflowStatusPending,
		StartTime:  time.Now(),
		Input:      input,
		Context:    make(map[string]interface{}),
		CreatedBy:  createdBy,
	}

	// Initialize task executions
	for _, taskDef := range workflow.Tasks {
		taskExec := TaskExecution{
			ID:         generateID(),
			WorkflowID: execution.ID,
			TaskID:     taskDef.ID,
			Name:       taskDef.Name,
			Status:     TaskStatusPending,
			Input:      make(map[string]interface{}),
			Output:     make(map[string]interface{}),
			Logs:       []string{},
		}
		execution.Tasks = append(execution.Tasks, taskExec)
	}

	// Save execution to database
	if err := ae.db.Create(execution).Error; err != nil {
		return nil, fmt.Errorf("failed to create workflow execution: %w", err)
	}

	// Start workflow runner
	ae.mutex.Lock()
	runner := &WorkflowRunner{
		execution: execution,
		engine:    ae,
		logger:    ae.logger.With(zap.String("workflow_id", execution.ID)),
	}
	runner.ctx, runner.cancel = context.WithCancel(context.Background())
	ae.runningWorkflows[execution.ID] = runner
	ae.mutex.Unlock()

	// Execute workflow asynchronously
	go runner.Run()

	ae.logger.Info("Workflow execution started",
		zap.String("execution_id", execution.ID),
		zap.String("workflow_id", workflowID),
		zap.String("created_by", createdBy))

	return execution, nil
}

// Run executes the workflow
func (wr *WorkflowRunner) Run() {
	defer func() {
		wr.engine.mutex.Lock()
		delete(wr.engine.runningWorkflows, wr.execution.ID)
		wr.engine.mutex.Unlock()
	}()

	wr.logger.Info("Starting workflow execution")

	// Update status to running
	wr.execution.Status = WorkflowStatusRunning
	wr.updateExecution()

	// Execute tasks in dependency order
	if err := wr.executeTasks(); err != nil {
		wr.execution.Status = WorkflowStatusFailed
		wr.execution.ErrorMessage = err.Error()
		wr.logger.Error("Workflow execution failed", zap.Error(err))
	} else {
		wr.execution.Status = WorkflowStatusCompleted
		wr.logger.Info("Workflow execution completed successfully")
	}

	// Set end time and update
	now := time.Now()
	wr.execution.EndTime = &now
	wr.updateExecution()
}

// executeTasks executes tasks in the correct dependency order
func (wr *WorkflowRunner) executeTasks() error {
	// Get workflow definition with tasks
	var workflow WorkflowDefinition
	if err := wr.engine.db.First(&workflow, "id = ?", wr.execution.WorkflowID).Error; err != nil {
		return fmt.Errorf("failed to load workflow definition: %w", err)
	}

	// Create task execution map for easy lookup
	taskExecMap := make(map[string]*TaskExecution)
	for i := range wr.execution.Tasks {
		taskExecMap[wr.execution.Tasks[i].TaskID] = &wr.execution.Tasks[i]
	}

	// Execute tasks in dependency order
	executed := make(map[string]bool)
	
	for len(executed) < len(workflow.Tasks) {
		progress := false
		
		for _, taskDef := range workflow.Tasks {
			if executed[taskDef.ID] {
				continue
			}
			
			// Check if all dependencies are completed
			canExecute := true
			for _, depID := range taskDef.Dependencies {
				if !executed[depID] {
					canExecute = false
					break
				}
			}
			
			if canExecute {
				taskExec := taskExecMap[taskDef.ID]
				
				// Check task conditions
				if !wr.evaluateTaskConditions(taskDef.Conditions) {
					taskExec.Status = TaskStatusSkipped
					taskExec.Logs = append(taskExec.Logs, "Task skipped due to unmet conditions")
					executed[taskDef.ID] = true
					progress = true
					continue
				}
				
				// Execute task
				if err := wr.executeTask(taskExec, &taskDef); err != nil {
					if taskDef.OnFailure == "continue" {
						wr.logger.Warn("Task failed but continuing", zap.String("task", taskDef.Name), zap.Error(err))
						taskExec.Status = TaskStatusFailed
						taskExec.ErrorMessage = err.Error()
					} else {
						return fmt.Errorf("task %s failed: %w", taskDef.Name, err)
					}
				}
				
				executed[taskDef.ID] = true
				progress = true
			}
		}
		
		if !progress {
			return fmt.Errorf("circular dependency detected or missing dependencies")
		}
	}
	
	return nil
}

// executeTask executes a single task
func (wr *WorkflowRunner) executeTask(taskExec *TaskExecution, taskDef *TaskDefinition) error {
	wr.logger.Info("Executing task", zap.String("task", taskDef.Name))
	
	taskExec.Status = TaskStatusRunning
	taskExec.StartTime = time.Now()
	wr.updateExecution()
	
	// Get task executor
	executor, exists := wr.engine.taskExecutors[taskDef.Type]
	if !exists {
		return fmt.Errorf("unknown task type: %s", taskDef.Type)
	}
	
	// Set up task context with timeout
	ctx := wr.ctx
	if taskDef.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(wr.ctx, taskDef.Timeout)
		defer cancel()
	}
	
	// Execute task with retry logic
	var err error
	for attempt := 0; attempt <= taskDef.RetryCount; attempt++ {
		if attempt > 0 {
			wr.logger.Info("Retrying task", zap.String("task", taskDef.Name), zap.Int("attempt", attempt))
			taskExec.RetryCount = attempt
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
		}
		
		err = executor.Execute(ctx, taskExec, taskDef)
		if err == nil {
			break
		}
		
		taskExec.Logs = append(taskExec.Logs, fmt.Sprintf("Attempt %d failed: %v", attempt+1, err))
	}
	
	// Update task status
	now := time.Now()
	taskExec.EndTime = &now
	
	if err != nil {
		taskExec.Status = TaskStatusFailed
		taskExec.ErrorMessage = err.Error()
	} else {
		taskExec.Status = TaskStatusCompleted
	}
	
	wr.updateExecution()
	return err
}

// evaluateTaskConditions evaluates task execution conditions
func (wr *WorkflowRunner) evaluateTaskConditions(conditions []TaskCondition) bool {
	for _, condition := range conditions {
		if !wr.evaluateCondition(condition) {
			return false
		}
	}
	return true
}

// evaluateCondition evaluates a single condition
func (wr *WorkflowRunner) evaluateCondition(condition TaskCondition) bool {
	// Get field value from context
	value, exists := wr.execution.Context[condition.Field]
	if !exists {
		return false
	}
	
	switch condition.Operator {
	case "eq":
		return value == condition.Value
	case "ne":
		return value != condition.Value
	case "gt":
		if v1, ok := value.(float64); ok {
			if v2, ok := condition.Value.(float64); ok {
				return v1 > v2
			}
		}
	case "lt":
		if v1, ok := value.(float64); ok {
			if v2, ok := condition.Value.(float64); ok {
				return v1 < v2
			}
		}
	case "contains":
		if v1, ok := value.(string); ok {
			if v2, ok := condition.Value.(string); ok {
				return contains(v1, v2)
			}
		}
	}
	
	return false
}

// updateExecution updates the execution in the database
func (wr *WorkflowRunner) updateExecution() {
	if err := wr.engine.db.Save(wr.execution).Error; err != nil {
		wr.logger.Error("Failed to update workflow execution", zap.Error(err))
	}
}

// validateWorkflow validates a workflow definition
func (ae *AutomationEngine) validateWorkflow(workflow *WorkflowDefinition) error {
	// Check for circular dependencies
	if err := ae.checkCircularDependencies(workflow.Tasks); err != nil {
		return err
	}
	
	// Validate each task
	for _, task := range workflow.Tasks {
		executor, exists := ae.taskExecutors[task.Type]
		if !exists {
			return fmt.Errorf("unknown task type: %s", task.Type)
		}
		
		if err := executor.Validate(&task); err != nil {
			return fmt.Errorf("task %s validation failed: %w", task.Name, err)
		}
	}
	
	return nil
}

// checkCircularDependencies checks for circular dependencies in tasks
func (ae *AutomationEngine) checkCircularDependencies(tasks []TaskDefinition) error {
	// Create adjacency list
	graph := make(map[string][]string)
	for _, task := range tasks {
		graph[task.ID] = task.Dependencies
	}
	
	// DFS to detect cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	
	for taskID := range graph {
		if !visited[taskID] {
			if ae.hasCycle(taskID, graph, visited, recStack) {
				return fmt.Errorf("circular dependency detected involving task: %s", taskID)
			}
		}
	}
	
	return nil
}

// hasCycle performs DFS to detect cycles
func (ae *AutomationEngine) hasCycle(taskID string, graph map[string][]string, visited, recStack map[string]bool) bool {
	visited[taskID] = true
	recStack[taskID] = true
	
	for _, dep := range graph[taskID] {
		if !visited[dep] {
			if ae.hasCycle(dep, graph, visited, recStack) {
				return true
			}
		} else if recStack[dep] {
			return true
		}
	}
	
	recStack[taskID] = false
	return false
}

// GetWorkflowExecution retrieves a workflow execution by ID
func (ae *AutomationEngine) GetWorkflowExecution(executionID string) (*WorkflowExecution, error) {
	var execution WorkflowExecution
	err := ae.db.Preload("Tasks").First(&execution, "id = ?", executionID).Error
	return &execution, err
}

// ListWorkflowExecutions lists workflow executions with filtering
func (ae *AutomationEngine) ListWorkflowExecutions(filters map[string]interface{}) ([]WorkflowExecution, error) {
	var executions []WorkflowExecution
	query := ae.db.Preload("Tasks")
	
	for field, value := range filters {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}
	
	err := query.Find(&executions).Error
	return executions, err
}

// CancelWorkflow cancels a running workflow
func (ae *AutomationEngine) CancelWorkflow(executionID string) error {
	ae.mutex.RLock()
	runner, exists := ae.runningWorkflows[executionID]
	ae.mutex.RUnlock()
	
	if exists {
		runner.cancel()
		runner.execution.Status = WorkflowStatusCancelled
		runner.updateExecution()
		ae.logger.Info("Workflow cancelled", zap.String("execution_id", executionID))
	}
	
	return nil
}

// Helper functions
func generateID() string {
	return fmt.Sprintf("wf_%d", time.Now().UnixNano())
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && s != substr && len(s) >= len(substr)))
}