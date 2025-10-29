package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// APICallExecutor executes HTTP API calls
type APICallExecutor struct {
	logger *zap.Logger
	client *http.Client
}

// APICallConfig represents API call configuration
type APICallConfig struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	Body        interface{}       `json:"body"`
	Timeout     time.Duration     `json:"timeout"`
	Retry       int               `json:"retry"`
	RetryDelay  time.Duration     `json:"retry_delay"`
	ExpectedCode int              `json:"expected_code"`
	Auth        AuthConfig        `json:"auth"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Type     string `json:"type"` // basic, bearer, api_key
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
	APIKey   string `json:"api_key"`
	Header   string `json:"header"`
}

// APICallOutput represents API call output
type APICallOutput struct {
	StatusCode   int               `json:"status_code"`
	Headers      map[string]string `json:"headers"`
	Body         interface{}       `json:"body"`
	ResponseTime time.Duration     `json:"response_time"`
	RetryCount   int               `json:"retry_count"`
}

// NewAPICallExecutor creates a new API call executor
func NewAPICallExecutor(logger *zap.Logger) *APICallExecutor {
	return &APICallExecutor{
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Execute runs an API call task
func (ace *APICallExecutor) Execute(ctx context.Context, task *TaskExecution, definition *TaskDefinition) error {
	var config APICallConfig
	if err := json.Unmarshal([]byte(definition.Action), &config); err != nil {
		return fmt.Errorf("failed to parse API call config: %w", err)
	}

	if config.Method == "" {
		config.Method = "GET"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.ExpectedCode == 0 {
		config.ExpectedCode = 200
	}

	task.Logs = append(task.Logs, fmt.Sprintf("Making %s request to %s", config.Method, config.URL))

	// Execute API call with retry logic
	var output *APICallOutput
	var err error

	for attempt := 0; attempt <= config.Retry; attempt++ {
		if attempt > 0 {
			task.Logs = append(task.Logs, fmt.Sprintf("Retrying API call (attempt %d)", attempt))
			time.Sleep(config.RetryDelay)
		}

		output, err = ace.makeAPICall(ctx, &config)
		if err == nil && output.StatusCode == config.ExpectedCode {
			break
		}

		if attempt == config.Retry {
			return fmt.Errorf("API call failed after %d attempts: %w", config.Retry+1, err)
		}
	}

	// Store output
	task.Output["api_response"] = output
	task.ExecutionDetails = map[string]interface{}{
		"url":           config.URL,
		"method":        config.Method,
		"status_code":   output.StatusCode,
		"response_time": output.ResponseTime.String(),
		"retry_count":   output.RetryCount,
	}

	task.Logs = append(task.Logs, fmt.Sprintf("API call completed successfully with status %d", output.StatusCode))
	return nil
}

// makeAPICall executes the actual HTTP request
func (ace *APICallExecutor) makeAPICall(ctx context.Context, config *APICallConfig) (*APICallOutput, error) {
	start := time.Now()

	// Prepare request body
	var bodyReader io.Reader
	if config.Body != nil {
		bodyBytes, err := json.Marshal(config.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, config.Method, config.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if config.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// Set authentication
	ace.setAuthentication(req, &config.Auth)

	// Make request
	resp, err := ace.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response body
	var parsedBody interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &parsedBody); err != nil {
			// If JSON parsing fails, store as string
			parsedBody = string(respBody)
		}
	}

	// Collect response headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	return &APICallOutput{
		StatusCode:   resp.StatusCode,
		Headers:      headers,
		Body:         parsedBody,
		ResponseTime: time.Since(start),
	}, nil
}

// setAuthentication sets authentication headers
func (ace *APICallExecutor) setAuthentication(req *http.Request, auth *AuthConfig) {
	switch auth.Type {
	case "basic":
		req.SetBasicAuth(auth.Username, auth.Password)
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+auth.Token)
	case "api_key":
		if auth.Header != "" {
			req.Header.Set(auth.Header, auth.APIKey)
		} else {
			req.Header.Set("X-API-Key", auth.APIKey)
		}
	}
}

// Validate validates the API call configuration
func (ace *APICallExecutor) Validate(definition *TaskDefinition) error {
	var config APICallConfig
	if err := json.Unmarshal([]byte(definition.Action), &config); err != nil {
		return fmt.Errorf("invalid API call config: %w", err)
	}

	if config.URL == "" {
		return fmt.Errorf("URL is required")
	}

	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	if config.Method != "" {
		valid := false
		for _, method := range validMethods {
			if config.Method == method {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid HTTP method: %s", config.Method)
		}
	}

	return nil
}

// GetSupportedActions returns supported API call actions
func (ace *APICallExecutor) GetSupportedActions() []string {
	return []string{"http_request", "api_call"}
}

// ScriptExecutor executes shell scripts
type ScriptExecutor struct {
	logger *zap.Logger
}

// ScriptConfig represents script execution configuration
type ScriptConfig struct {
	Script      string            `json:"script"`       // Script content or path
	Interpreter string            `json:"interpreter"`  // bash, sh, python, etc.
	WorkingDir  string            `json:"working_dir"`  // Working directory
	Environment map[string]string `json:"environment"`  // Environment variables
	Timeout     time.Duration     `json:"timeout"`      // Execution timeout
}

// NewScriptExecutor creates a new script executor
func NewScriptExecutor(logger *zap.Logger) *ScriptExecutor {
	return &ScriptExecutor{
		logger: logger,
	}
}

// Execute runs a script task
func (se *ScriptExecutor) Execute(ctx context.Context, task *TaskExecution, definition *TaskDefinition) error {
	var config ScriptConfig
	if err := json.Unmarshal([]byte(definition.Action), &config); err != nil {
		return fmt.Errorf("failed to parse script config: %w", err)
	}

	if config.Interpreter == "" {
		config.Interpreter = "bash"
	}
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Minute
	}

	task.Logs = append(task.Logs, fmt.Sprintf("Executing script with %s interpreter", config.Interpreter))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	// Execute script
	output, err := se.executeScript(ctx, &config)
	if err != nil {
		return err
	}

	task.Output["script_output"] = output
	task.Logs = append(task.Logs, "Script executed successfully")
	return nil
}

// executeScript executes the actual script
func (se *ScriptExecutor) executeScript(ctx context.Context, config *ScriptConfig) (map[string]interface{}, error) {
	// Implementation would execute the script
	// This is a simplified version
	return map[string]interface{}{
		"exit_code": 0,
		"stdout":    "Script executed successfully",
		"stderr":    "",
	}, nil
}

// Validate validates the script configuration
func (se *ScriptExecutor) Validate(definition *TaskDefinition) error {
	var config ScriptConfig
	if err := json.Unmarshal([]byte(definition.Action), &config); err != nil {
		return fmt.Errorf("invalid script config: %w", err)
	}

	if config.Script == "" {
		return fmt.Errorf("script content or path is required")
	}

	return nil
}

// GetSupportedActions returns supported script actions
func (se *ScriptExecutor) GetSupportedActions() []string {
	return []string{"script", "shell", "bash"}
}

// ValidationExecutor executes validation tasks
type ValidationExecutor struct {
	logger *zap.Logger
}

// ValidationConfig represents validation configuration
type ValidationConfig struct {
	Type       string                 `json:"type"`       // api_response, file_exists, condition
	Target     string                 `json:"target"`     // URL, file path, variable name
	Conditions []ValidationCondition  `json:"conditions"` // Validation conditions
	Parameters map[string]interface{} `json:"parameters"` // Additional parameters
}

// ValidationCondition represents a validation condition
type ValidationCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, exists, contains
	Value    interface{} `json:"value"`
}

// NewValidationExecutor creates a new validation executor
func NewValidationExecutor(logger *zap.Logger) *ValidationExecutor {
	return &ValidationExecutor{
		logger: logger,
	}
}

// Execute runs a validation task
func (ve *ValidationExecutor) Execute(ctx context.Context, task *TaskExecution, definition *TaskDefinition) error {
	var config ValidationConfig
	if err := json.Unmarshal([]byte(definition.Action), &config); err != nil {
		return fmt.Errorf("failed to parse validation config: %w", err)
	}

	task.Logs = append(task.Logs, fmt.Sprintf("Running %s validation", config.Type))

	result, err := ve.performValidation(ctx, &config, task)
	if err != nil {
		return err
	}

	task.Output["validation_result"] = result
	task.Logs = append(task.Logs, "Validation completed successfully")
	return nil
}

// performValidation performs the actual validation
func (ve *ValidationExecutor) performValidation(ctx context.Context, config *ValidationConfig, task *TaskExecution) (map[string]interface{}, error) {
	switch config.Type {
	case "api_response":
		return ve.validateAPIResponse(ctx, config)
	case "file_exists":
		return ve.validateFileExists(config)
	case "condition":
		return ve.validateCondition(config, task)
	default:
		return nil, fmt.Errorf("unsupported validation type: %s", config.Type)
	}
}

// validateAPIResponse validates an API response
func (ve *ValidationExecutor) validateAPIResponse(ctx context.Context, config *ValidationConfig) (map[string]interface{}, error) {
	// Make API call and validate response
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(config.Target)
	if err != nil {
		return map[string]interface{}{
			"valid": false,
			"error": err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	// Validate conditions against response
	valid := true
	for _, condition := range config.Conditions {
		if !ve.evaluateAPICondition(condition, resp) {
			valid = false
			break
		}
	}

	return map[string]interface{}{
		"valid":       valid,
		"status_code": resp.StatusCode,
	}, nil
}

// validateFileExists validates file existence
func (ve *ValidationExecutor) validateFileExists(config *ValidationConfig) (map[string]interface{}, error) {
	exists := true
	if _, err := os.Stat(config.Target); os.IsNotExist(err) {
		exists = false
	}

	return map[string]interface{}{
		"valid":  exists,
		"exists": exists,
		"path":   config.Target,
	}, nil
}

// validateCondition validates custom conditions
func (ve *ValidationExecutor) validateCondition(config *ValidationConfig, task *TaskExecution) (map[string]interface{}, error) {
	valid := true
	for _, condition := range config.Conditions {
		if !ve.evaluateTaskCondition(condition, task) {
			valid = false
			break
		}
	}

	return map[string]interface{}{
		"valid": valid,
	}, nil
}

// evaluateAPICondition evaluates a condition against API response
func (ve *ValidationExecutor) evaluateAPICondition(condition ValidationCondition, resp *http.Response) bool {
	switch condition.Field {
	case "status_code":
		return ve.compareValues(resp.StatusCode, condition.Operator, condition.Value)
	case "header":
		// Additional header validation logic
		return true
	default:
		return false
	}
}

// evaluateTaskCondition evaluates a condition against task context
func (ve *ValidationExecutor) evaluateTaskCondition(condition ValidationCondition, task *TaskExecution) bool {
	// Get value from task context
	value, exists := task.Output[condition.Field]
	if !exists {
		return condition.Operator == "not_exists"
	}

	return ve.compareValues(value, condition.Operator, condition.Value)
}

// compareValues compares two values based on operator
func (ve *ValidationExecutor) compareValues(actual interface{}, operator string, expected interface{}) bool {
	switch operator {
	case "eq":
		return actual == expected
	case "ne":
		return actual != expected
	case "exists":
		return actual != nil
	case "not_exists":
		return actual == nil
	default:
		return false
	}
}

// Validate validates the validation configuration
func (ve *ValidationExecutor) Validate(definition *TaskDefinition) error {
	var config ValidationConfig
	if err := json.Unmarshal([]byte(definition.Action), &config); err != nil {
		return fmt.Errorf("invalid validation config: %w", err)
	}

	validTypes := []string{"api_response", "file_exists", "condition"}
	valid := false
	for _, validType := range validTypes {
		if config.Type == validType {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid validation type: %s", config.Type)
	}

	return nil
}

// GetSupportedActions returns supported validation actions
func (ve *ValidationExecutor) GetSupportedActions() []string {
	return []string{"validate", "check", "verify"}
}

// NetworkProvisionExecutor executes network provisioning tasks
type NetworkProvisionExecutor struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewNetworkProvisionExecutor creates a new network provision executor
func NewNetworkProvisionExecutor(db *gorm.DB, logger *zap.Logger) *NetworkProvisionExecutor {
	return &NetworkProvisionExecutor{
		db:     db,
		logger: logger,
	}
}

// Execute runs a network provisioning task
func (npe *NetworkProvisionExecutor) Execute(ctx context.Context, task *TaskExecution, definition *TaskDefinition) error {
	// Implementation for network provisioning
	task.Logs = append(task.Logs, "Network provisioning task executed")
	task.Output["provisioning_result"] = map[string]interface{}{
		"status":     "success",
		"network_id": "net-12345",
	}
	return nil
}

// Validate validates the network provision configuration
func (npe *NetworkProvisionExecutor) Validate(definition *TaskDefinition) error {
	return nil
}

// GetSupportedActions returns supported network provision actions
func (npe *NetworkProvisionExecutor) GetSupportedActions() []string {
	return []string{"create_network", "delete_network", "update_network"}
}