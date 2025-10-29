package security

import (
	"encoding/json"
	"log"
	"time"
)

type AuditLogLevel string

const (
	AuditInfo     AuditLogLevel = "INFO"
	AuditWarning  AuditLogLevel = "WARNING"
	AuditError    AuditLogLevel = "ERROR"
	AuditCritical AuditLogLevel = "CRITICAL"
)

type AuditEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Level       AuditLogLevel          `json:"level"`
	UserID      string                 `json:"user_id,omitempty"`
	Username    string                 `json:"username,omitempty"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Success     bool                   `json:"success"`
	ErrorMsg    string                 `json:"error_message,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
}

type AuditLogger struct {
	// In production, this would write to a secure audit database
	// For demo purposes, we'll log to stdout with structured format
}

func NewAuditLogger() *AuditLogger {
	return &AuditLogger{}
}

// LogEvent records an audit event
func (a *AuditLogger) LogEvent(event AuditEvent) {
	event.ID = generateID()
	event.Timestamp = time.Now()

	// Convert to JSON for structured logging
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("AUDIT_ERROR: Failed to marshal audit event: %v", err)
		return
	}

	// In production, this would be sent to:
	// - Secure audit database
	// - SIEM system
	// - Compliance logging service
	log.Printf("AUDIT_LOG: %s", string(eventJSON))
}

// Pre-built audit event creators
func (a *AuditLogger) LogAuthentication(userID, username, ipAddress string, success bool, errorMsg string) {
	a.LogEvent(AuditEvent{
		Level:     AuditInfo,
		UserID:    userID,
		Username:  username,
		Action:    "authenticate",
		Resource:  "user_session",
		IPAddress: ipAddress,
		Success:   success,
		ErrorMsg:  errorMsg,
	})
}

func (a *AuditLogger) LogWorkflowExecution(userID, username, workflowID, workflowName string, success bool, duration time.Duration) {
	level := AuditInfo
	if !success {
		level = AuditError
	}

	a.LogEvent(AuditEvent{
		Level:      level,
		UserID:     userID,
		Username:   username,
		Action:     "execute_workflow",
		Resource:   "workflow",
		ResourceID: workflowID,
		Success:    success,
		Duration:   duration,
		Metadata: map[string]interface{}{
			"workflow_name": workflowName,
		},
	})
}

func (a *AuditLogger) LogAPIAccess(userID, username, method, path, ipAddress, userAgent string, statusCode int, duration time.Duration) {
	level := AuditInfo
	success := statusCode < 400

	if statusCode >= 400 && statusCode < 500 {
		level = AuditWarning
	} else if statusCode >= 500 {
		level = AuditError
	}

	a.LogEvent(AuditEvent{
		Level:     level,
		UserID:    userID,
		Username:  username,
		Action:    method,
		Resource:  path,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   success,
		Duration:  duration,
		Metadata: map[string]interface{}{
			"status_code": statusCode,
		},
	})
}

func (a *AuditLogger) LogSecurityEvent(userID, username, eventType, description, ipAddress string, level AuditLogLevel) {
	a.LogEvent(AuditEvent{
		Level:     level,
		UserID:    userID,
		Username:  username,
		Action:    "security_event",
		Resource:  "system",
		IPAddress: ipAddress,
		Success:   false, // Security events are typically failures or warnings
		ErrorMsg:  description,
		Metadata: map[string]interface{}{
			"event_type": eventType,
		},
	})
}