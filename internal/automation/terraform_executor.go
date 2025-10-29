package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// TerraformExecutor executes Terraform tasks for infrastructure automation
type TerraformExecutor struct {
	logger *zap.Logger
}

// TerraformConfig represents Terraform task configuration
type TerraformConfig struct {
	Action      string            `json:"action"`      // plan, apply, destroy, validate
	WorkingDir  string            `json:"working_dir"` // Directory containing Terraform files
	Variables   map[string]string `json:"variables"`   // Terraform variables
	StateFile   string            `json:"state_file"`  // Custom state file path
	AutoApprove bool              `json:"auto_approve"`
	Targets     []string          `json:"targets"` // Specific resources to target
	Workspace   string            `json:"workspace"`
}

// TerraformOutput represents Terraform execution output
type TerraformOutput struct {
	ExitCode int                    `json:"exit_code"`
	Stdout   string                 `json:"stdout"`
	Stderr   string                 `json:"stderr"`
	Outputs  map[string]interface{} `json:"outputs"`
	Plan     *TerraformPlan         `json:"plan,omitempty"`
}

// TerraformPlan represents a Terraform plan
type TerraformPlan struct {
	ResourceChanges []TerraformResourceChange `json:"resource_changes"`
	Summary         TerraformPlanSummary      `json:"summary"`
}

// TerraformResourceChange represents a planned resource change
type TerraformResourceChange struct {
	Address      string                 `json:"address"`
	Mode         string                 `json:"mode"`
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	Change       TerraformChange        `json:"change"`
	Values       map[string]interface{} `json:"values"`
}

// TerraformChange represents the change action
type TerraformChange struct {
	Actions []string               `json:"actions"`
	Before  map[string]interface{} `json:"before"`
	After   map[string]interface{} `json:"after"`
}

// TerraformPlanSummary provides a summary of planned changes
type TerraformPlanSummary struct {
	Add    int `json:"add"`
	Change int `json:"change"`
	Delete int `json:"delete"`
}

// NewTerraformExecutor creates a new Terraform executor
func NewTerraformExecutor(logger *zap.Logger) *TerraformExecutor {
	return &TerraformExecutor{
		logger: logger,
	}
}

// Execute runs a Terraform task
func (te *TerraformExecutor) Execute(ctx context.Context, task *TaskExecution, definition *TaskDefinition) error {
	var config TerraformConfig
	if err := json.Unmarshal([]byte(definition.Action), &config); err != nil {
		// Try to parse from parameters if action is simple string
		config = TerraformConfig{
			Action:      definition.Action,
			WorkingDir:  getStringParam(definition.Parameters, "working_dir", "."),
			Variables:   getMapParam(definition.Parameters, "variables"),
			StateFile:   getStringParam(definition.Parameters, "state_file", ""),
			AutoApprove: getBoolParam(definition.Parameters, "auto_approve", false),
			Targets:     getSliceParam(definition.Parameters, "targets"),
			Workspace:   getStringParam(definition.Parameters, "workspace", "default"),
		}
	}

	task.Logs = append(task.Logs, fmt.Sprintf("Starting Terraform %s action", config.Action))

	// Ensure Terraform is installed
	if err := te.checkTerraformInstalled(); err != nil {
		return fmt.Errorf("terraform not available: %w", err)
	}

	// Create working directory if it doesn't exist
	if err := os.MkdirAll(config.WorkingDir, 0755); err != nil {
		return fmt.Errorf("failed to create working directory: %w", err)
	}

	// Initialize Terraform if needed
	if err := te.initTerraform(ctx, config.WorkingDir); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	// Select workspace if specified
	if config.Workspace != "default" {
		if err := te.selectWorkspace(ctx, config.WorkingDir, config.Workspace); err != nil {
			return fmt.Errorf("workspace selection failed: %w", err)
		}
	}

	// Execute the specified action
	output, err := te.executeAction(ctx, &config, task)
	if err != nil {
		return err
	}

	// Store output in task execution
	task.Output["terraform_output"] = output
	task.ExecutionDetails = map[string]interface{}{
		"action":       config.Action,
		"working_dir":  config.WorkingDir,
		"workspace":    config.Workspace,
		"exit_code":    output.ExitCode,
		"has_changes":  output.Plan != nil && (output.Plan.Summary.Add > 0 || output.Plan.Summary.Change > 0 || output.Plan.Summary.Delete > 0),
	}

	task.Logs = append(task.Logs, fmt.Sprintf("Terraform %s completed successfully", config.Action))
	return nil
}

// executeAction executes the specific Terraform action
func (te *TerraformExecutor) executeAction(ctx context.Context, config *TerraformConfig, task *TaskExecution) (*TerraformOutput, error) {
	switch config.Action {
	case "validate":
		return te.validate(ctx, config)
	case "plan":
		return te.plan(ctx, config)
	case "apply":
		return te.apply(ctx, config, task)
	case "destroy":
		return te.destroy(ctx, config, task)
	case "output":
		return te.getOutputs(ctx, config)
	default:
		return nil, fmt.Errorf("unsupported terraform action: %s", config.Action)
	}
}

// validate runs terraform validate
func (te *TerraformExecutor) validate(ctx context.Context, config *TerraformConfig) (*TerraformOutput, error) {
	cmd := exec.CommandContext(ctx, "terraform", "validate", "-json")
	cmd.Dir = config.WorkingDir

	stdout, stderr, exitCode := te.runCommand(cmd)
	
	return &TerraformOutput{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}, nil
}

// plan runs terraform plan
func (te *TerraformExecutor) plan(ctx context.Context, config *TerraformConfig) (*TerraformOutput, error) {
	args := []string{"plan", "-json", "-detailed-exitcode"}
	
	// Add variables
	for key, value := range config.Variables {
		args = append(args, "-var", fmt.Sprintf("%s=%s", key, value))
	}
	
	// Add targets
	for _, target := range config.Targets {
		args = append(args, "-target", target)
	}
	
	// Add state file
	if config.StateFile != "" {
		args = append(args, "-state", config.StateFile)
	}

	cmd := exec.CommandContext(ctx, "terraform", args...)
	cmd.Dir = config.WorkingDir

	stdout, stderr, exitCode := te.runCommand(cmd)
	
	output := &TerraformOutput{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}

	// Parse plan if successful
	if exitCode == 0 || exitCode == 2 { // 2 means plan has changes
		if plan, err := te.parsePlan(stdout); err == nil {
			output.Plan = plan
		}
	}

	return output, nil
}

// apply runs terraform apply
func (te *TerraformExecutor) apply(ctx context.Context, config *TerraformConfig, task *TaskExecution) (*TerraformOutput, error) {
	args := []string{"apply", "-json"}
	
	if config.AutoApprove {
		args = append(args, "-auto-approve")
	}
	
	// Add variables
	for key, value := range config.Variables {
		args = append(args, "-var", fmt.Sprintf("%s=%s", key, value))
	}
	
	// Add targets
	for _, target := range config.Targets {
		args = append(args, "-target", target)
	}
	
	// Add state file
	if config.StateFile != "" {
		args = append(args, "-state", config.StateFile)
	}

	cmd := exec.CommandContext(ctx, "terraform", args...)
	cmd.Dir = config.WorkingDir

	// Log the start of apply
	task.Logs = append(task.Logs, "Starting terraform apply...")

	stdout, stderr, exitCode := te.runCommand(cmd)
	
	output := &TerraformOutput{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}

	// Get outputs after successful apply
	if exitCode == 0 {
		if outputs, err := te.getTerraformOutputs(ctx, config.WorkingDir); err == nil {
			output.Outputs = outputs
		}
		task.Logs = append(task.Logs, "Terraform apply completed successfully")
	} else {
		task.Logs = append(task.Logs, fmt.Sprintf("Terraform apply failed with exit code %d", exitCode))
	}

	return output, nil
}

// destroy runs terraform destroy
func (te *TerraformExecutor) destroy(ctx context.Context, config *TerraformConfig, task *TaskExecution) (*TerraformOutput, error) {
	args := []string{"destroy"}
	
	if config.AutoApprove {
		args = append(args, "-auto-approve")
	}
	
	// Add variables
	for key, value := range config.Variables {
		args = append(args, "-var", fmt.Sprintf("%s=%s", key, value))
	}
	
	// Add targets
	for _, target := range config.Targets {
		args = append(args, "-target", target)
	}

	cmd := exec.CommandContext(ctx, "terraform", args...)
	cmd.Dir = config.WorkingDir

	task.Logs = append(task.Logs, "Starting terraform destroy...")

	stdout, stderr, exitCode := te.runCommand(cmd)
	
	if exitCode == 0 {
		task.Logs = append(task.Logs, "Terraform destroy completed successfully")
	} else {
		task.Logs = append(task.Logs, fmt.Sprintf("Terraform destroy failed with exit code %d", exitCode))
	}

	return &TerraformOutput{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}, nil
}

// getOutputs retrieves terraform outputs
func (te *TerraformExecutor) getOutputs(ctx context.Context, config *TerraformConfig) (*TerraformOutput, error) {
	outputs, err := te.getTerraformOutputs(ctx, config.WorkingDir)
	if err != nil {
		return nil, err
	}

	return &TerraformOutput{
		ExitCode: 0,
		Outputs:  outputs,
	}, nil
}

// initTerraform runs terraform init
func (te *TerraformExecutor) initTerraform(ctx context.Context, workingDir string) error {
	cmd := exec.CommandContext(ctx, "terraform", "init", "-input=false")
	cmd.Dir = workingDir

	_, stderr, exitCode := te.runCommand(cmd)
	
	if exitCode != 0 {
		return fmt.Errorf("terraform init failed: %s", stderr)
	}

	return nil
}

// selectWorkspace selects or creates a terraform workspace
func (te *TerraformExecutor) selectWorkspace(ctx context.Context, workingDir, workspace string) error {
	// Try to select workspace
	cmd := exec.CommandContext(ctx, "terraform", "workspace", "select", workspace)
	cmd.Dir = workingDir

	_, _, exitCode := te.runCommand(cmd)
	
	if exitCode != 0 {
		// Workspace doesn't exist, try to create it
		cmd = exec.CommandContext(ctx, "terraform", "workspace", "new", workspace)
		cmd.Dir = workingDir

		_, stderr, exitCode := te.runCommand(cmd)
		
		if exitCode != 0 {
			return fmt.Errorf("failed to create workspace %s: %s", workspace, stderr)
		}
	}

	return nil
}

// getTerraformOutputs retrieves all terraform outputs
func (te *TerraformExecutor) getTerraformOutputs(ctx context.Context, workingDir string) (map[string]interface{}, error) {
	cmd := exec.CommandContext(ctx, "terraform", "output", "-json")
	cmd.Dir = workingDir

	stdout, stderr, exitCode := te.runCommand(cmd)
	
	if exitCode != 0 {
		return nil, fmt.Errorf("terraform output failed: %s", stderr)
	}

	var outputs map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &outputs); err != nil {
		return nil, fmt.Errorf("failed to parse terraform outputs: %w", err)
	}

	// Extract just the values
	result := make(map[string]interface{})
	for key, output := range outputs {
		if outputMap, ok := output.(map[string]interface{}); ok {
			if value, exists := outputMap["value"]; exists {
				result[key] = value
			}
		}
	}

	return result, nil
}

// parsePlan parses terraform plan output
func (te *TerraformExecutor) parsePlan(planOutput string) (*TerraformPlan, error) {
	lines := strings.Split(planOutput, "\n")
	plan := &TerraformPlan{
		ResourceChanges: []TerraformResourceChange{},
		Summary:         TerraformPlanSummary{},
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var planLine map[string]interface{}
		if err := json.Unmarshal([]byte(line), &planLine); err != nil {
			continue // Skip non-JSON lines
		}

		if planLine["type"] == "resource_drift" || planLine["type"] == "planned_change" {
			if change, ok := planLine["change"].(map[string]interface{}); ok {
				resourceChange := TerraformResourceChange{
					Address: getStringFromMap(change, "resource.addr"),
					Mode:    getStringFromMap(change, "resource.resource_mode"),
					Type:    getStringFromMap(change, "resource.resource_type"),
					Name:    getStringFromMap(change, "resource.resource_name"),
				}

				if actions, ok := change["action"].([]interface{}); ok {
					for _, action := range actions {
						if actionStr, ok := action.(string); ok {
							resourceChange.Change.Actions = append(resourceChange.Change.Actions, actionStr)
							
							// Count changes
							switch actionStr {
							case "create":
								plan.Summary.Add++
							case "update":
								plan.Summary.Change++
							case "delete":
								plan.Summary.Delete++
							}
						}
					}
				}

				plan.ResourceChanges = append(plan.ResourceChanges, resourceChange)
			}
		}
	}

	return plan, nil
}

// runCommand executes a command and returns stdout, stderr, and exit code
func (te *TerraformExecutor) runCommand(cmd *exec.Cmd) (string, string, int) {
	te.logger.Debug("Executing command", zap.String("cmd", cmd.String()))

	// Set environment variables
	cmd.Env = append(os.Environ(),
		"TF_IN_AUTOMATION=true",
		"TF_INPUT=false",
	)

	stdout, err := cmd.Output()
	exitCode := 0
	stderr := ""

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
			stderr = string(exitError.Stderr)
		} else {
			exitCode = 1
			stderr = err.Error()
		}
	}

	return string(stdout), stderr, exitCode
}

// checkTerraformInstalled verifies terraform is available
func (te *TerraformExecutor) checkTerraformInstalled() error {
	cmd := exec.Command("terraform", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform is not installed or not in PATH")
	}
	return nil
}

// Validate validates the task definition
func (te *TerraformExecutor) Validate(definition *TaskDefinition) error {
	// Check if terraform is available
	if err := te.checkTerraformInstalled(); err != nil {
		return err
	}

	// Validate action
	validActions := []string{"validate", "plan", "apply", "destroy", "output"}
	action := definition.Action
	
	// If action is JSON, parse it
	var config TerraformConfig
	if err := json.Unmarshal([]byte(definition.Action), &config); err == nil {
		action = config.Action
	}

	for _, validAction := range validActions {
		if action == validAction {
			return nil
		}
	}

	return fmt.Errorf("invalid terraform action: %s. Valid actions: %v", action, validActions)
}

// GetSupportedActions returns supported terraform actions
func (te *TerraformExecutor) GetSupportedActions() []string {
	return []string{"validate", "plan", "apply", "destroy", "output"}
}

// Helper functions
func getStringFromMap(m map[string]interface{}, key string) string {
	if value, exists := m[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if value, exists := params[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if value, exists := params[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func getMapParam(params map[string]interface{}, key string) map[string]string {
	result := make(map[string]string)
	if value, exists := params[key]; exists {
		if m, ok := value.(map[string]interface{}); ok {
			for k, v := range m {
				if str, ok := v.(string); ok {
					result[k] = str
				}
			}
		}
	}
	return result
}

func getSliceParam(params map[string]interface{}, key string) []string {
	var result []string
	if value, exists := params[key]; exists {
		if slice, ok := value.([]interface{}); ok {
			for _, item := range slice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
		}
	}
	return result
}