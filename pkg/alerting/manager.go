package alerting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityError    AlertSeverity = "error"
	SeverityCritical AlertSeverity = "critical"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	StatusFiring   AlertStatus = "firing"
	StatusResolved AlertStatus = "resolved"
)

// Alert represents a system alert
type Alert struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Severity    AlertSeverity     `json:"severity"`
	Status      AlertStatus       `json:"status"`
	Source      string            `json:"source"`
	Tags        map[string]string `json:"tags"`
	Timestamp   time.Time         `json:"timestamp"`
	ResolvedAt  *time.Time        `json:"resolved_at,omitempty"`
}

// AlertRule represents a rule that can trigger alerts
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Query       string                 `json:"query"`
	Condition   string                 `json:"condition"`
	Threshold   float64                `json:"threshold"`
	Duration    time.Duration          `json:"duration"`
	Severity    AlertSeverity          `json:"severity"`
	Enabled     bool                   `json:"enabled"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	LastEval    time.Time              `json:"last_eval"`
	NextEval    time.Time              `json:"next_eval"`
}

// NotificationChannel represents a notification delivery channel
type NotificationChannel struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"` // slack, email, webhook, pagerduty
	Config   map[string]string `json:"config"`
	Enabled  bool              `json:"enabled"`
}

// AlertManager manages alerts and notifications
type AlertManager struct {
	alerts              map[string]*Alert
	rules               map[string]*AlertRule
	channels            map[string]*NotificationChannel
	logger              *zap.Logger
	httpClient          *http.Client
}

// NewAlertManager creates a new alert manager
func NewAlertManager(logger *zap.Logger) *AlertManager {
	return &AlertManager{
		alerts:   make(map[string]*Alert),
		rules:    make(map[string]*AlertRule),
		channels: make(map[string]*NotificationChannel),
		logger:   logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Initialize sets up default alert rules and notification channels
func (am *AlertManager) Initialize() {
	// Default alert rules
	am.AddAlertRule(&AlertRule{
		ID:          "high_cpu_usage",
		Name:        "High CPU Usage",
		Description: "Node CPU usage exceeds threshold",
		Query:       "netorchestrator_node_cpu_usage_percent",
		Condition:   ">",
		Threshold:   85.0,
		Duration:    5 * time.Minute,
		Severity:    SeverityWarning,
		Enabled:     true,
		Labels: map[string]string{
			"category": "performance",
		},
		Annotations: map[string]string{
			"runbook": "https://docs.netorchestrator.io/runbooks/high-cpu",
		},
	})

	am.AddAlertRule(&AlertRule{
		ID:          "high_memory_usage",
		Name:        "High Memory Usage",
		Description: "Node memory usage exceeds threshold",
		Query:       "netorchestrator_node_memory_usage_percent",
		Condition:   ">",
		Threshold:   90.0,
		Duration:    3 * time.Minute,
		Severity:    SeverityError,
		Enabled:     true,
		Labels: map[string]string{
			"category": "performance",
		},
	})

	am.AddAlertRule(&AlertRule{
		ID:          "network_down",
		Name:        "Network Down",
		Description: "Network is not responding",
		Query:       "netorchestrator_networks_active",
		Condition:   "==",
		Threshold:   0,
		Duration:    1 * time.Minute,
		Severity:    SeverityCritical,
		Enabled:     true,
		Labels: map[string]string{
			"category": "availability",
		},
	})

	am.AddAlertRule(&AlertRule{
		ID:          "websocket_connections_high",
		Name:        "High WebSocket Connections",
		Description: "Too many WebSocket connections",
		Query:       "netorchestrator_websocket_connections",
		Condition:   ">",
		Threshold:   1000,
		Duration:    2 * time.Minute,
		Severity:    SeverityWarning,
		Enabled:     true,
		Labels: map[string]string{
			"category": "capacity",
		},
	})

	am.AddAlertRule(&AlertRule{
		ID:          "sla_breach",
		Name:        "SLA Breach",
		Description: "SLA compliance below threshold",
		Query:       "netorchestrator_sla_compliance_percent",
		Condition:   "<",
		Threshold:   95.0,
		Duration:    5 * time.Minute,
		Severity:    SeverityError,
		Enabled:     true,
		Labels: map[string]string{
			"category": "business",
		},
	})

	// Default notification channels
	am.AddNotificationChannel(&NotificationChannel{
		ID:   "default_slack",
		Name: "Default Slack Channel",
		Type: "slack",
		Config: map[string]string{
			"webhook_url": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK",
			"channel":     "#netorchestrator-alerts",
			"username":    "NetOrchestrator",
		},
		Enabled: true,
	})

	am.AddNotificationChannel(&NotificationChannel{
		ID:   "ops_email",
		Name: "Operations Team Email",
		Type: "email",
		Config: map[string]string{
			"smtp_host":     "smtp.company.com",
			"smtp_port":     "587",
			"smtp_username": "alerts@company.com",
			"smtp_password": "password",
			"to":           "ops-team@company.com",
			"from":         "netorchestrator@company.com",
		},
		Enabled: true,
	})

	am.AddNotificationChannel(&NotificationChannel{
		ID:   "webhook_pagerduty",
		Name: "PagerDuty Integration",
		Type: "webhook",
		Config: map[string]string{
			"url":    "https://events.pagerduty.com/v2/enqueue",
			"token":  "YOUR_PAGERDUTY_TOKEN",
			"source": "netorchestrator",
		},
		Enabled: true,
	})

	am.logger.Info("Alert manager initialized with default rules and channels")
}

// AddAlertRule adds a new alert rule
func (am *AlertManager) AddAlertRule(rule *AlertRule) {
	am.rules[rule.ID] = rule
	am.logger.Info("Added alert rule", zap.String("id", rule.ID), zap.String("name", rule.Name))
}

// AddNotificationChannel adds a new notification channel
func (am *AlertManager) AddNotificationChannel(channel *NotificationChannel) {
	am.channels[channel.ID] = channel
	am.logger.Info("Added notification channel", zap.String("id", channel.ID), zap.String("name", channel.Name))
}

// TriggerAlert creates and sends an alert
func (am *AlertManager) TriggerAlert(title, description, source string, severity AlertSeverity, tags map[string]string) *Alert {
	alert := &Alert{
		ID:          am.generateAlertID(),
		Title:       title,
		Description: description,
		Severity:    severity,
		Status:      StatusFiring,
		Source:      source,
		Tags:        tags,
		Timestamp:   time.Now(),
	}

	am.alerts[alert.ID] = alert

	// Send notifications
	am.sendNotifications(alert)

	am.logger.Warn("Alert triggered",
		zap.String("id", alert.ID),
		zap.String("title", alert.Title),
		zap.String("severity", string(alert.Severity)),
		zap.String("source", alert.Source))

	return alert
}

// ResolveAlert marks an alert as resolved
func (am *AlertManager) ResolveAlert(alertID string) {
	if alert, exists := am.alerts[alertID]; exists {
		now := time.Now()
		alert.Status = StatusResolved
		alert.ResolvedAt = &now

		am.logger.Info("Alert resolved", zap.String("id", alertID))
	}
}

// sendNotifications sends alert notifications to all enabled channels
func (am *AlertManager) sendNotifications(alert *Alert) {
	for _, channel := range am.channels {
		if !channel.Enabled {
			continue
		}

		go func(ch *NotificationChannel) {
			if err := am.sendToChannel(alert, ch); err != nil {
				am.logger.Error("Failed to send notification",
					zap.String("channel", ch.Name),
					zap.Error(err))
			}
		}(channel)
	}
}

// sendToChannel sends an alert to a specific notification channel
func (am *AlertManager) sendToChannel(alert *Alert, channel *NotificationChannel) error {
	switch channel.Type {
	case "slack":
		return am.sendToSlack(alert, channel)
	case "email":
		return am.sendToEmail(alert, channel)
	case "webhook":
		return am.sendToWebhook(alert, channel)
	default:
		return fmt.Errorf("unsupported channel type: %s", channel.Type)
	}
}

// sendToSlack sends alert to Slack
func (am *AlertManager) sendToSlack(alert *Alert, channel *NotificationChannel) error {
	webhookURL := channel.Config["webhook_url"]
	if webhookURL == "" {
		return fmt.Errorf("missing webhook_url in slack channel config")
	}

	color := am.getSeverityColor(alert.Severity)
	
	payload := map[string]interface{}{
		"channel":   channel.Config["channel"],
		"username":  channel.Config["username"],
		"attachments": []map[string]interface{}{
			{
				"color":      color,
				"title":      fmt.Sprintf("🚨 %s", alert.Title),
				"text":       alert.Description,
				"fields": []map[string]interface{}{
					{"title": "Severity", "value": string(alert.Severity), "short": true},
					{"title": "Source", "value": alert.Source, "short": true},
					{"title": "Status", "value": string(alert.Status), "short": true},
					{"title": "Timestamp", "value": alert.Timestamp.Format(time.RFC3339), "short": true},
				},
				"footer": "NetOrchestrator Alert Manager",
				"ts":     alert.Timestamp.Unix(),
			},
		},
	}

	return am.sendHTTPRequest(webhookURL, payload)
}

// sendToEmail sends alert via email (placeholder implementation)
func (am *AlertManager) sendToEmail(alert *Alert, channel *NotificationChannel) error {
	// In a real implementation, you would use an SMTP client here
	am.logger.Info("Email notification sent (placeholder)",
		zap.String("to", channel.Config["to"]),
		zap.String("alert", alert.Title))
	return nil
}

// sendToWebhook sends alert to a generic webhook
func (am *AlertManager) sendToWebhook(alert *Alert, channel *NotificationChannel) error {
	webhookURL := channel.Config["url"]
	if webhookURL == "" {
		return fmt.Errorf("missing url in webhook channel config")
	}

	payload := map[string]interface{}{
		"alert_id":    alert.ID,
		"title":       alert.Title,
		"description": alert.Description,
		"severity":    alert.Severity,
		"status":      alert.Status,
		"source":      alert.Source,
		"tags":        alert.Tags,
		"timestamp":   alert.Timestamp.Unix(),
	}

	return am.sendHTTPRequest(webhookURL, payload)
}

// sendHTTPRequest sends an HTTP request with JSON payload
func (am *AlertManager) sendHTTPRequest(url string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := am.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	return nil
}

// getSeverityColor returns a color for Slack based on severity
func (am *AlertManager) getSeverityColor(severity AlertSeverity) string {
	switch severity {
	case SeverityInfo:
		return "#36a64f" // Green
	case SeverityWarning:
		return "#ffeb3b" // Yellow
	case SeverityError:
		return "#ff9800" // Orange
	case SeverityCritical:
		return "#f44336" // Red
	default:
		return "#9e9e9e" // Gray
	}
}

// generateAlertID generates a unique alert ID
func (am *AlertManager) generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

// GetAlerts returns all alerts
func (am *AlertManager) GetAlerts() map[string]*Alert {
	return am.alerts
}

// GetAlertRules returns all alert rules
func (am *AlertManager) GetAlertRules() map[string]*AlertRule {
	return am.rules
}

// GetNotificationChannels returns all notification channels
func (am *AlertManager) GetNotificationChannels() map[string]*NotificationChannel {
	return am.channels
}

// StartAlertEvaluation starts the alert evaluation loop
func (am *AlertManager) StartAlertEvaluation() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			am.evaluateRules()
		}
	}()
	am.logger.Info("Alert evaluation started")
}

// evaluateRules evaluates all enabled alert rules
func (am *AlertManager) evaluateRules() {
	for _, rule := range am.rules {
		if !rule.Enabled || time.Now().Before(rule.NextEval) {
			continue
		}

		// Update evaluation times
		rule.LastEval = time.Now()
		rule.NextEval = time.Now().Add(1 * time.Minute)

		// In a real implementation, you would query Prometheus here
		// For now, we'll simulate some rule evaluations
		am.simulateRuleEvaluation(rule)
	}
}

// simulateRuleEvaluation simulates rule evaluation (placeholder)
func (am *AlertManager) simulateRuleEvaluation(rule *AlertRule) {
	// This is a placeholder implementation
	// In reality, you would query your metrics store (Prometheus) here
	
	shouldAlert := false
	
	// Simple simulation based on time for demo purposes
	now := time.Now()
	if strings.Contains(rule.ID, "cpu") && now.Second()%30 < 5 {
		shouldAlert = true
	} else if strings.Contains(rule.ID, "memory") && now.Second()%45 < 3 {
		shouldAlert = true
	}

	if shouldAlert {
		am.TriggerAlert(
			rule.Name,
			fmt.Sprintf("Rule '%s' condition met: %s %s %.1f", 
				rule.Name, rule.Query, rule.Condition, rule.Threshold),
			"alert_manager",
			rule.Severity,
			rule.Labels,
		)
	}
}