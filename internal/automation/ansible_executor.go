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

// AnsibleExecutor executes Ansible tasks for configuration management
type AnsibleExecutor struct {
	logger *zap.Logger
}

// AnsibleConfig represents Ansible task configuration
type AnsibleConfig struct {
	Action       string                 `json:"action"`       // playbook, ad-hoc, vault
	Playbook     string                 `json:"playbook"`     // Path to playbook
	Inventory    string                 `json:"inventory"`    // Inventory file or inline
	ExtraVars    map[string]interface{} `json:"extra_vars"`   // Extra variables
	Tags         []string               `json:"tags"`         // Tags to run
	SkipTags     []string               `json:"skip_tags"`    // Tags to skip
	Limit        string                 `json:"limit"`        // Limit to specific hosts
	Check        bool                   `json:"check"`        // Dry run mode
	Diff         bool                   `json:"diff"`         // Show diffs
	Verbose      int                    `json:"verbose"`      // Verbosity level (0-4)
	PrivateKey   string                 `json:"private_key"`  // SSH private key path
	User         string                 `json:"user"`         // Remote user
	BecomeUser   string                 `json:"become_user"`  // Become user (sudo)
	Become       bool                   `json:"become"`       // Use privilege escalation
	Module       string                 `json:"module"`       // Module for ad-hoc commands
	ModuleArgs   string                 `json:"module_args"`  // Module arguments
	VaultPassFile string                `json:"vault_pass_file"` // Vault password file
}

// AnsibleOutput represents Ansible execution output
type AnsibleOutput struct {
	ExitCode    int                    `json:"exit_code"`
	Stdout      string                 `json:"stdout"`
	Stderr      string                 `json:"stderr"`
	PlayRecap   *AnsiblePlayRecap      `json:"play_recap,omitempty"`
	TaskResults []AnsibleTaskResult    `json:"task_results"`
	Stats       map[string]interface{} `json:"stats"`
}

// AnsiblePlayRecap represents the final recap of an ansible run
type AnsiblePlayRecap struct {
	Hosts map[string]AnsibleHostStats `json:"hosts"`
}

// AnsibleHostStats represents statistics for a single host
type AnsibleHostStats struct {
	Ok          int `json:"ok"`
	Changed     int `json:"changed"`
	Unreachable int `json:"unreachable"`
	Failed      int `json:"failed"`
	Skipped     int `json:"skipped"`
	Rescued     int `json:"rescued"`
	Ignored     int `json:"ignored"`
}

// AnsibleTaskResult represents the result of a single task
type AnsibleTaskResult struct {
	Host     string                 `json:"host"`
	Task     string                 `json:"task"`
	Action   string                 `json:"action"`
	Status   string                 `json:"status"` // ok, changed, failed, skipped
	Result   map[string]interface{} `json:"result"`
	Duration float64                `json:"duration"`
}

// NewAnsibleExecutor creates a new Ansible executor
func NewAnsibleExecutor(logger *zap.Logger) *AnsibleExecutor {
	return &AnsibleExecutor{
		logger: logger,
	}
}

// Execute runs an Ansible task
func (ae *AnsibleExecutor) Execute(ctx context.Context, task *TaskExecution, definition *TaskDefinition) error {
	var config AnsibleConfig
	if err := json.Unmarshal([]byte(definition.Action), &config); err != nil {
		// Try to parse from parameters if action is simple string
		config = AnsibleConfig{
			Action:        definition.Action,
			Playbook:      getStringParam(definition.Parameters, "playbook", ""),
			Inventory:     getStringParam(definition.Parameters, "inventory", "localhost,"),
			ExtraVars:     definition.Parameters["extra_vars"].(map[string]interface{}),
			Tags:          getSliceParam(definition.Parameters, "tags"),
			SkipTags:      getSliceParam(definition.Parameters, "skip_tags"),
			Limit:         getStringParam(definition.Parameters, "limit", ""),
			Check:         getBoolParam(definition.Parameters, "check", false),
			Diff:          getBoolParam(definition.Parameters, "diff", false),
			Verbose:       getIntParam(definition.Parameters, "verbose", 0),
			PrivateKey:    getStringParam(definition.Parameters, "private_key", ""),
			User:          getStringParam(definition.Parameters, "user", ""),
			BecomeUser:    getStringParam(definition.Parameters, "become_user", ""),
			Become:        getBoolParam(definition.Parameters, "become", false),
			Module:        getStringParam(definition.Parameters, "module", ""),
			ModuleArgs:    getStringParam(definition.Parameters, "module_args", ""),
			VaultPassFile: getStringParam(definition.Parameters, "vault_pass_file", ""),
		}
	}

	task.Logs = append(task.Logs, fmt.Sprintf("Starting Ansible %s action", config.Action))

	// Ensure Ansible is installed
	if err := ae.checkAnsibleInstalled(); err != nil {
		return fmt.Errorf("ansible not available: %w", err)
	}

	// Execute the specified action
	output, err := ae.executeAction(ctx, &config, task)
	if err != nil {
		return err
	}

	// Store output in task execution
	task.Output["ansible_output"] = output
	task.ExecutionDetails = map[string]interface{}{
		"action":     config.Action,
		"playbook":   config.Playbook,
		"inventory":  config.Inventory,
		"exit_code":  output.ExitCode,
		"has_changes": ae.hasChanges(output),
	}

	task.Logs = append(task.Logs, fmt.Sprintf("Ansible %s completed successfully", config.Action))
	return nil
}

// executeAction executes the specific Ansible action
func (ae *AnsibleExecutor) executeAction(ctx context.Context, config *AnsibleConfig, task *TaskExecution) (*AnsibleOutput, error) {
	switch config.Action {
	case "playbook":
		return ae.runPlaybook(ctx, config, task)
	case "ad-hoc":
		return ae.runAdHoc(ctx, config, task)
	case "vault":
		return ae.runVault(ctx, config, task)
	case "galaxy":
		return ae.runGalaxy(ctx, config, task)
	default:
		return nil, fmt.Errorf("unsupported ansible action: %s", config.Action)
	}
}

// runPlaybook executes an ansible playbook
func (ae *AnsibleExecutor) runPlaybook(ctx context.Context, config *AnsibleConfig, task *TaskExecution) (*AnsibleOutput, error) {
	if config.Playbook == "" {
		return nil, fmt.Errorf("playbook path is required")
	}

	args := []string{"ansible-playbook"}
	
	// Add inventory
	if config.Inventory != "" {
		args = append(args, "-i", config.Inventory)
	}
	
	// Add extra vars
	if len(config.ExtraVars) > 0 {
		extraVarsJSON, _ := json.Marshal(config.ExtraVars)
		args = append(args, "--extra-vars", string(extraVarsJSON))
	}
	
	// Add tags
	if len(config.Tags) > 0 {
		args = append(args, "--tags", strings.Join(config.Tags, ","))
	}
	
	// Add skip tags
	if len(config.SkipTags) > 0 {
		args = append(args, "--skip-tags", strings.Join(config.SkipTags, ","))
	}
	
	// Add limit
	if config.Limit != "" {
		args = append(args, "--limit", config.Limit)
	}
	
	// Add check mode
	if config.Check {
		args = append(args, "--check")
	}
	
	// Add diff
	if config.Diff {
		args = append(args, "--diff")
	}
	
	// Add verbosity
	if config.Verbose > 0 {
		verbosity := strings.Repeat("v", config.Verbose)
		args = append(args, "-"+verbosity)
	}
	
	// Add private key
	if config.PrivateKey != "" {
		args = append(args, "--private-key", config.PrivateKey)
	}
	
	// Add user
	if config.User != "" {
		args = append(args, "-u", config.User)
	}
	
	// Add become
	if config.Become {
		args = append(args, "--become")
		if config.BecomeUser != "" {
			args = append(args, "--become-user", config.BecomeUser)
		}
	}
	
	// Add vault password file
	if config.VaultPassFile != "" {
		args = append(args, "--vault-password-file", config.VaultPassFile)
	}
	
	// Add playbook path
	args = append(args, config.Playbook)

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	task.Logs = append(task.Logs, fmt.Sprintf("Running: %s", strings.Join(args, " ")))

	stdout, stderr, exitCode := ae.runCommand(cmd)
	
	output := &AnsibleOutput{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}

	// Parse ansible output for structured data
	ae.parseAnsibleOutput(output, stdout)

	if exitCode == 0 {
		task.Logs = append(task.Logs, "Ansible playbook executed successfully")
	} else {
		task.Logs = append(task.Logs, fmt.Sprintf("Ansible playbook failed with exit code %d", exitCode))
	}

	return output, nil
}

// runAdHoc executes an ansible ad-hoc command
func (ae *AnsibleExecutor) runAdHoc(ctx context.Context, config *AnsibleConfig, task *TaskExecution) (*AnsibleOutput, error) {
	if config.Module == "" {
		return nil, fmt.Errorf("module is required for ad-hoc commands")
	}

	args := []string{"ansible", "all"}
	
	// Add inventory
	if config.Inventory != "" {
		args = append(args, "-i", config.Inventory)
	}
	
	// Add module
	args = append(args, "-m", config.Module)
	
	// Add module args
	if config.ModuleArgs != "" {
		args = append(args, "-a", config.ModuleArgs)
	}
	
	// Add limit
	if config.Limit != "" {
		args = []string{"ansible", config.Limit}
		args = append(args, "-i", config.Inventory, "-m", config.Module)
		if config.ModuleArgs != "" {
			args = append(args, "-a", config.ModuleArgs)
		}
	}
	
	// Add user
	if config.User != "" {
		args = append(args, "-u", config.User)
	}
	
	// Add become
	if config.Become {
		args = append(args, "--become")
		if config.BecomeUser != "" {
			args = append(args, "--become-user", config.BecomeUser)
		}
	}
	
	// Add private key
	if config.PrivateKey != "" {
		args = append(args, "--private-key", config.PrivateKey)
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	task.Logs = append(task.Logs, fmt.Sprintf("Running ad-hoc: %s", strings.Join(args, " ")))

	stdout, stderr, exitCode := ae.runCommand(cmd)
	
	output := &AnsibleOutput{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}

	if exitCode == 0 {
		task.Logs = append(task.Logs, "Ansible ad-hoc command executed successfully")
	} else {
		task.Logs = append(task.Logs, fmt.Sprintf("Ansible ad-hoc command failed with exit code %d", exitCode))
	}

	return output, nil
}

// runVault executes ansible vault operations
func (ae *AnsibleExecutor) runVault(ctx context.Context, config *AnsibleConfig, task *TaskExecution) (*AnsibleOutput, error) {
	// Implementation for ansible-vault operations
	args := []string{"ansible-vault"}
	
	// Add vault password file
	if config.VaultPassFile != "" {
		args = append(args, "--vault-password-file", config.VaultPassFile)
	}
	
	// Add specific vault operation from module args
	if config.ModuleArgs != "" {
		vaultArgs := strings.Fields(config.ModuleArgs)
		args = append(args, vaultArgs...)
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	stdout, stderr, exitCode := ae.runCommand(cmd)
	
	return &AnsibleOutput{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}, nil
}

// runGalaxy executes ansible galaxy operations
func (ae *AnsibleExecutor) runGalaxy(ctx context.Context, config *AnsibleConfig, task *TaskExecution) (*AnsibleOutput, error) {
	args := []string{"ansible-galaxy"}
	
	// Add specific galaxy operation from module args
	if config.ModuleArgs != "" {
		galaxyArgs := strings.Fields(config.ModuleArgs)
		args = append(args, galaxyArgs...)
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	stdout, stderr, exitCode := ae.runCommand(cmd)
	
	return &AnsibleOutput{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}, nil
}

// parseAnsibleOutput parses ansible output for structured data
func (ae *AnsibleExecutor) parseAnsibleOutput(output *AnsibleOutput, stdout string) {
	lines := strings.Split(stdout, "\n")
	
	// Initialize structures
	output.PlayRecap = &AnsiblePlayRecap{
		Hosts: make(map[string]AnsibleHostStats),
	}
	output.TaskResults = []AnsibleTaskResult{}
	output.Stats = make(map[string]interface{})
	
	// Parse output line by line
	inRecap := false
	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		// Look for play recap section
		if strings.Contains(line, "PLAY RECAP") {
			inRecap = true
			continue
		}
		
		if inRecap && line != "" && !strings.HasPrefix(line, "=") {
			ae.parseRecapLine(line, output.PlayRecap)
		}
		
		// Look for task results (simplified parsing)
		if strings.Contains(line, "TASK [") || strings.Contains(line, "ok:") || strings.Contains(line, "changed:") || strings.Contains(line, "failed:") {
			if taskResult := ae.parseTaskLine(line, i, lines); taskResult != nil {
				output.TaskResults = append(output.TaskResults, *taskResult)
			}
		}
	}
}

// parseRecapLine parses a single recap line
func (ae *AnsibleExecutor) parseRecapLine(line string, recap *AnsiblePlayRecap) {
	// Example: "web-01                     : ok=5    changed=2    unreachable=0    failed=0    skipped=1    rescued=0    ignored=0"
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return
	}
	
	hostname := strings.TrimSpace(parts[0])
	statsPart := strings.TrimSpace(parts[1])
	
	stats := AnsibleHostStats{}
	
	// Parse stats
	statItems := strings.Fields(statsPart)
	for _, item := range statItems {
		if strings.Contains(item, "=") {
			kv := strings.Split(item, "=")
			if len(kv) == 2 {
				key := kv[0]
				value := parseInt(kv[1])
				
				switch key {
				case "ok":
					stats.Ok = value
				case "changed":
					stats.Changed = value
				case "unreachable":
					stats.Unreachable = value
				case "failed":
					stats.Failed = value
				case "skipped":
					stats.Skipped = value
				case "rescued":
					stats.Rescued = value
				case "ignored":
					stats.Ignored = value
				}
			}
		}
	}
	
	recap.Hosts[hostname] = stats
}

// parseTaskLine parses a task execution line
func (ae *AnsibleExecutor) parseTaskLine(line string, lineNum int, allLines []string) *AnsibleTaskResult {
	// Simplified parsing - in reality this would be more complex
	result := &AnsibleTaskResult{}
	
	if strings.Contains(line, "ok:") {
		result.Status = "ok"
	} else if strings.Contains(line, "changed:") {
		result.Status = "changed"
	} else if strings.Contains(line, "failed:") {
		result.Status = "failed"
	} else if strings.Contains(line, "skipping:") {
		result.Status = "skipped"
	} else {
		return nil
	}
	
	// Extract host name (simplified)
	if idx := strings.Index(line, "["); idx != -1 {
		if endIdx := strings.Index(line[idx:], "]"); endIdx != -1 {
			result.Host = line[idx+1 : idx+endIdx]
		}
	}
	
	return result
}

// hasChanges checks if ansible execution had any changes
func (ae *AnsibleExecutor) hasChanges(output *AnsibleOutput) bool {
	if output.PlayRecap == nil {
		return false
	}
	
	for _, stats := range output.PlayRecap.Hosts {
		if stats.Changed > 0 {
			return true
		}
	}
	
	return false
}

// runCommand executes a command and returns stdout, stderr, and exit code
func (ae *AnsibleExecutor) runCommand(cmd *exec.Cmd) (string, string, int) {
	ae.logger.Debug("Executing command", zap.String("cmd", cmd.String()))

	// Set environment variables
	cmd.Env = append(os.Environ(),
		"ANSIBLE_HOST_KEY_CHECKING=False",
		"ANSIBLE_RETRY_FILES_ENABLED=False",
		"ANSIBLE_STDOUT_CALLBACK=json",
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

// checkAnsibleInstalled verifies ansible is available
func (ae *AnsibleExecutor) checkAnsibleInstalled() error {
	cmd := exec.Command("ansible", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ansible is not installed or not in PATH")
	}
	return nil
}

// Validate validates the task definition
func (ae *AnsibleExecutor) Validate(definition *TaskDefinition) error {
	// Check if ansible is available
	if err := ae.checkAnsibleInstalled(); err != nil {
		return err
	}

	validActions := []string{"playbook", "ad-hoc", "vault", "galaxy"}
	action := definition.Action
	
	// If action is JSON, parse it
	var config AnsibleConfig
	if err := json.Unmarshal([]byte(definition.Action), &config); err == nil {
		action = config.Action
	}

	for _, validAction := range validActions {
		if action == validAction {
			return nil
		}
	}

	return fmt.Errorf("invalid ansible action: %s. Valid actions: %v", action, validActions)
}

// GetSupportedActions returns supported ansible actions
func (ae *AnsibleExecutor) GetSupportedActions() []string {
	return []string{"playbook", "ad-hoc", "vault", "galaxy"}
}

// Helper functions
func getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if value, exists := params[key]; exists {
		if i, ok := value.(int); ok {
			return i
		}
		if f, ok := value.(float64); ok {
			return int(f)
		}
	}
	return defaultValue
}

func parseInt(s string) int {
	// Simple integer parsing - in reality would use strconv.Atoi
	result := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		}
	}
	return result
}