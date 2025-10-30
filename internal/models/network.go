package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Network represents a virtual network topology
type Network struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	Description string         `json:"description"`
	Status      NetworkStatus  `json:"status" gorm:"type:varchar(20);default:'pending'"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	Config      NetworkConfig  `json:"config" gorm:"type:jsonb"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Nodes    []Node   `json:"nodes" gorm:"foreignKey:NetworkID"`
	Links    []Link   `json:"links" gorm:"foreignKey:NetworkID"`
	Policies []Policy `json:"policies" gorm:"foreignKey:NetworkID"`
	User     User     `json:"user" gorm:"foreignKey:UserID"`
}

// NetworkStatus represents the current status of a network
type NetworkStatus string

const (
	NetworkStatusPending      NetworkStatus = "pending"
	NetworkStatusProvisioning NetworkStatus = "provisioning"
	NetworkStatusActive       NetworkStatus = "active"
	NetworkStatusError        NetworkStatus = "error"
	NetworkStatusSuspended    NetworkStatus = "suspended"
	NetworkStatusDeleting     NetworkStatus = "deleting"
)

// NetworkConfig holds network configuration
type NetworkConfig struct {
	Topology    string            `json:"topology"` // star, mesh, tree, custom
	Subnet      string            `json:"subnet"`   // CIDR notation
	Gateway     string            `json:"gateway"`
	DNS         []string          `json:"dns"`
	MTU         int               `json:"mtu"`
	QoS         QoSConfig         `json:"qos"`
	Security    SecurityConfig    `json:"security"`
	Monitoring  MonitoringConfig  `json:"monitoring"`
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// Node represents a network node (router, switch, host, etc.)
type Node struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	NetworkID  uuid.UUID      `json:"network_id" gorm:"type:uuid;not null"`
	Name       string         `json:"name" gorm:"not null"`
	Type       NodeType       `json:"type" gorm:"column:node_type;type:varchar(20);not null"`
	Status     NodeStatus     `json:"status" gorm:"type:varchar(20);default:'pending'"`
	IPAddress  string         `json:"ip_address"`
	MACAddress string         `json:"mac_address"`
	Config     NodeConfig     `json:"config" gorm:"type:jsonb"`
	Position   Position       `json:"position" gorm:"type:jsonb"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Network Network `json:"network" gorm:"foreignKey:NetworkID"`
	Links   []Link  `json:"links" gorm:"foreignKey:SourceNodeID"`
}

// NodeType represents the type of network node
type NodeType string

const (
	NodeTypeRouter       NodeType = "router"
	NodeTypeSwitch       NodeType = "switch"
	NodeTypeHost         NodeType = "host"
	NodeTypeFirewall     NodeType = "firewall"
	NodeTypeLoadBalancer NodeType = "load_balancer"
	NodeTypeGateway      NodeType = "gateway"
)

// NodeStatus represents the current status of a node
type NodeStatus string

const (
	NodeStatusPending      NodeStatus = "pending"
	NodeStatusProvisioning NodeStatus = "provisioning"
	NodeStatusActive       NodeStatus = "active"
	NodeStatusError        NodeStatus = "error"
	NodeStatusSuspended    NodeStatus = "suspended"
	NodeStatusDeleting     NodeStatus = "deleting"
)

// NodeConfig holds node-specific configuration
type NodeConfig struct {
	CPU         int               `json:"cpu"`
	Memory      int               `json:"memory"`  // MB
	Storage     int               `json:"storage"` // GB
	OS          string            `json:"os"`
	Image       string            `json:"image"`
	Ports       []PortConfig      `json:"ports"`
	Services    []ServiceConfig   `json:"services"`
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// PortConfig represents a network port configuration
type PortConfig struct {
	Number        int    `json:"number"`
	Name          string `json:"name"`
	Type          string `json:"type"`   // ethernet, wifi, etc.
	Speed         int    `json:"speed"`  // Mbps
	Duplex        string `json:"duplex"` // full, half
	AutoNegotiate bool   `json:"auto_negotiate"`
}

// ServiceConfig represents a service running on a node
type ServiceConfig struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Port        int               `json:"port"`
	Protocol    string            `json:"protocol"`
	Config      map[string]string `json:"config"`
	HealthCheck HealthCheckConfig `json:"health_check"`
}

// HealthCheckConfig holds health check configuration
type HealthCheckConfig struct {
	Enabled          bool   `json:"enabled"`
	Type             string `json:"type"` // http, tcp, ping
	Path             string `json:"path"`
	Interval         int    `json:"interval"` // seconds
	Timeout          int    `json:"timeout"`  // seconds
	Retries          int    `json:"retries"`
	SuccessThreshold int    `json:"success_threshold"`
	FailureThreshold int    `json:"failure_threshold"`
}

// Link represents a connection between two nodes
type Link struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	NetworkID    uuid.UUID      `json:"network_id" gorm:"type:uuid;not null"`
	SourceNodeID uuid.UUID      `json:"source_node_id" gorm:"type:uuid;not null"`
	TargetNodeID uuid.UUID      `json:"target_node_id" gorm:"type:uuid;not null"`
	Status       LinkStatus     `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Config       LinkConfig     `json:"config" gorm:"type:jsonb"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Network    Network `json:"network" gorm:"foreignKey:NetworkID"`
	SourceNode Node    `json:"source_node" gorm:"foreignKey:SourceNodeID"`
	TargetNode Node    `json:"target_node" gorm:"foreignKey:TargetNodeID"`
}

// LinkStatus represents the current status of a link
type LinkStatus string

const (
	LinkStatusPending      LinkStatus = "pending"
	LinkStatusProvisioning LinkStatus = "provisioning"
	LinkStatusActive       LinkStatus = "active"
	LinkStatusError        LinkStatus = "error"
	LinkStatusSuspended    LinkStatus = "suspended"
	LinkStatusDeleting     LinkStatus = "deleting"
)

// LinkConfig holds link-specific configuration
type LinkConfig struct {
	Bandwidth   int               `json:"bandwidth"`   // Mbps
	Latency     int               `json:"latency"`     // ms
	Jitter      int               `json:"jitter"`      // ms
	PacketLoss  float64           `json:"packet_loss"` // percentage
	Protocol    string            `json:"protocol"`    // ethernet, wifi, etc.
	Encryption  bool              `json:"encryption"`
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// Position represents 2D coordinates for network visualization
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// QoSConfig holds Quality of Service configuration
type QoSConfig struct {
	Enabled     bool              `json:"enabled"`
	Policies    []QoSPolicy       `json:"policies"`
	Bandwidth   BandwidthConfig   `json:"bandwidth"`
	Priority    PriorityConfig    `json:"priority"`
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// QoSPolicy represents a QoS policy
type QoSPolicy struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Rules       []QoSRule         `json:"rules"`
	Priority    int               `json:"priority"`
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// QoSRule represents a QoS rule
type QoSRule struct {
	Match       map[string]string `json:"match"`
	Action      map[string]string `json:"action"`
	Priority    int               `json:"priority"`
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// BandwidthConfig holds bandwidth configuration
type BandwidthConfig struct {
	Upstream   int `json:"upstream"`   // Mbps
	Downstream int `json:"downstream"` // Mbps
	Burst      int `json:"burst"`      // Mbps
}

// PriorityConfig holds priority configuration
type PriorityConfig struct {
	Levels []PriorityLevel `json:"levels"`
}

// PriorityLevel represents a priority level
type PriorityLevel struct {
	Name        string `json:"name"`
	Value       int    `json:"value"`
	Description string `json:"description"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	Firewall    FirewallConfig    `json:"firewall"`
	Encryption  EncryptionConfig  `json:"encryption"`
	Access      AccessConfig      `json:"access"`
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// FirewallConfig holds firewall configuration
type FirewallConfig struct {
	Enabled bool           `json:"enabled"`
	Rules   []FirewallRule `json:"rules"`
}

// FirewallRule represents a firewall rule
type FirewallRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Action      string            `json:"action"` // allow, deny, drop
	Protocol    string            `json:"protocol"`
	Source      string            `json:"source"`
	Destination string            `json:"destination"`
	Port        int               `json:"port"`
	Priority    int               `json:"priority"`
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	Enabled     bool              `json:"enabled"`
	Algorithm   string            `json:"algorithm"`
	KeySize     int               `json:"key_size"`
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// AccessConfig holds access control configuration
type AccessConfig struct {
	Enabled     bool              `json:"enabled"`
	Rules       []AccessRule      `json:"rules"`
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// AccessRule represents an access control rule
type AccessRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Subject     string            `json:"subject"`
	Resource    string            `json:"resource"`
	Action      string            `json:"action"`
	Effect      string            `json:"effect"` // allow, deny
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Enabled     bool              `json:"enabled"`
	Metrics     []MetricConfig    `json:"metrics"`
	Alerts      []AlertConfig     `json:"alerts"`
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// MetricConfig represents a monitoring metric
type MetricConfig struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"` // counter, gauge, histogram
	Description string            `json:"description"`
	Unit        string            `json:"unit"`
	Interval    int               `json:"interval"` // seconds
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// AlertConfig represents an alert configuration
type AlertConfig struct {
	Name        string            `json:"name"`
	Condition   string            `json:"condition"`
	Severity    string            `json:"severity"` // info, warning, error, critical
	Action      string            `json:"action"`
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// Policy represents a network policy
type Policy struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	NetworkID uuid.UUID      `json:"network_id" gorm:"type:uuid;not null"`
	Name      string         `json:"name" gorm:"not null"`
	Type      PolicyType     `json:"type" gorm:"column:policy_type;type:varchar(20);not null"`
	Status    PolicyStatus   `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Config    PolicyConfig   `json:"config" gorm:"type:jsonb"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Network Network `json:"network" gorm:"foreignKey:NetworkID"`
}

// PolicyType represents the type of network policy
type PolicyType string

const (
	PolicyTypeTraffic  PolicyType = "traffic"
	PolicyTypeSecurity PolicyType = "security"
	PolicyTypeQoS      PolicyType = "qos"
	PolicyTypeRouting  PolicyType = "routing"
	PolicyTypeAccess   PolicyType = "access"
)

// PolicyStatus represents the current status of a policy
type PolicyStatus string

const (
	PolicyStatusPending   PolicyStatus = "pending"
	PolicyStatusActive    PolicyStatus = "active"
	PolicyStatusError     PolicyStatus = "error"
	PolicyStatusSuspended PolicyStatus = "suspended"
	PolicyStatusDeleting  PolicyStatus = "deleting"
)

// PolicyConfig holds policy-specific configuration
type PolicyConfig struct {
	Rules       []PolicyRule      `json:"rules"`
	Priority    int               `json:"priority"`
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// PolicyRule represents a policy rule
type PolicyRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Condition   map[string]string `json:"condition"`
	Action      map[string]string `json:"action"`
	Priority    int               `json:"priority"`
	CustomAttrs map[string]string `json:"custom_attrs"`
}

// AlertRule represents an alert rule configuration
type AlertRule struct {
	ID          uuid.UUID          `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string             `json:"name" gorm:"not null"`
	Description string             `json:"description"`
	Condition   AlertRuleCondition `json:"condition" gorm:"type:jsonb"`
	Actions     AlertRuleActions   `json:"actions" gorm:"type:jsonb"`
	Severity    AlertSeverity      `json:"severity" gorm:"type:varchar(20);default:'warning'"`
	Status      AlertRuleStatus    `json:"status" gorm:"type:varchar(20);default:'active'"`
	Tags        StringSlice        `json:"tags" gorm:"type:jsonb"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   gorm.DeletedAt     `json:"deleted_at" gorm:"index"`
}

// AlertRuleCondition defines when an alert should trigger
type AlertRuleCondition struct {
	Metric     string            `json:"metric"`              // cpu_usage, memory_usage, network_latency, etc.
	Operator   string            `json:"operator"`            // >, <, >=, <=, ==, !=
	Threshold  float64           `json:"threshold"`           // threshold value (renamed from Value to avoid conflict)
	Duration   string            `json:"duration"`            // how long condition must be true (e.g., "5m")
	EntityType string            `json:"entity_type"`         // network, node, link
	EntityID   *uuid.UUID        `json:"entity_id,omitempty"` // specific entity or null for all
	Labels     map[string]string `json:"labels"`              // additional label filters
}

// AlertRuleAction defines what to do when an alert triggers
type AlertRuleAction struct {
	Type       string            `json:"type"`       // email, webhook, slack, etc.
	Target     string            `json:"target"`     // email address, webhook URL, etc.
	Template   string            `json:"template"`   // message template
	Parameters map[string]string `json:"parameters"` // additional action parameters
}

// AlertRuleActions is a wrapper type for []AlertRuleAction to implement Scanner/Valuer
type AlertRuleActions []AlertRuleAction

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	AlertSeverityDebug    AlertSeverity = "debug"
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertRuleStatus represents the status of an alert rule
type AlertRuleStatus string

const (
	AlertRuleStatusActive   AlertRuleStatus = "active"
	AlertRuleStatusInactive AlertRuleStatus = "inactive"
	AlertRuleStatusPaused   AlertRuleStatus = "paused"
)

// User represents a system user
type User struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"`
	Role      UserRole       `json:"role" gorm:"type:varchar(20);default:'user'"`
	Status    UserStatus     `json:"status" gorm:"type:varchar(20);default:'active'"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Networks []Network `json:"networks" gorm:"foreignKey:UserID"`
}

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
	UserRoleGuest UserRole = "guest"
)

// UserStatus represents the current status of a user
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

// GORM JSONB Scanner/Valuer implementations for PostgreSQL compatibility

// NetworkConfig JSONB methods
func (nc NetworkConfig) Value() (driver.Value, error) {
	return json.Marshal(nc)
}

func (nc *NetworkConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan NetworkConfig from non-bytes/string")
	}

	return json.Unmarshal(bytes, nc)
}

// NodeConfig JSONB methods
func (nc NodeConfig) Value() (driver.Value, error) {
	return json.Marshal(nc)
}

func (nc *NodeConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan NodeConfig from non-bytes/string")
	}

	return json.Unmarshal(bytes, nc)
}

// LinkConfig JSONB methods
func (lc LinkConfig) Value() (driver.Value, error) {
	return json.Marshal(lc)
}

func (lc *LinkConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan LinkConfig from non-bytes/string")
	}

	return json.Unmarshal(bytes, lc)
}

// PolicyConfig JSONB methods
func (pc PolicyConfig) Value() (driver.Value, error) {
	return json.Marshal(pc)
}

func (pc *PolicyConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan PolicyConfig from non-bytes/string")
	}

	return json.Unmarshal(bytes, pc)
}

// Position JSONB methods
func (p Position) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Position) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan Position from non-bytes/string")
	}

	return json.Unmarshal(bytes, p)
}

// AlertRuleCondition JSONB methods
func (arc AlertRuleCondition) Value() (driver.Value, error) {
	return json.Marshal(arc)
}

func (arc *AlertRuleCondition) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan AlertRuleCondition from non-bytes/string")
	}

	return json.Unmarshal(bytes, arc)
}

// AlertRuleActions JSONB methods
func (ara AlertRuleActions) Value() (driver.Value, error) {
	return json.Marshal(ara)
}

func (ara *AlertRuleActions) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan AlertRuleActions from non-bytes/string")
	}

	return json.Unmarshal(bytes, ara)
}

// NotificationChannel represents a notification channel configuration
type NotificationChannel struct {
	ID          uuid.UUID                 `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string                    `json:"name" gorm:"not null"`
	Description string                    `json:"description"`
	Type        NotificationChannelType   `json:"type" gorm:"type:varchar(20);not null"`
	Config      NotificationChannelConfig `json:"config" gorm:"type:jsonb"`
	Status      NotificationChannelStatus `json:"status" gorm:"type:varchar(20);default:'active'"`
	Tags        StringSlice               `json:"tags" gorm:"type:jsonb"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
	DeletedAt   gorm.DeletedAt            `json:"deleted_at" gorm:"index"`
}

// NotificationChannelType represents the type of notification channel
type NotificationChannelType string

const (
	NotificationChannelTypeEmail     NotificationChannelType = "email"
	NotificationChannelTypeSlack     NotificationChannelType = "slack"
	NotificationChannelTypeWebhook   NotificationChannelType = "webhook"
	NotificationChannelTypePagerDuty NotificationChannelType = "pagerduty"
	NotificationChannelTypeSMS       NotificationChannelType = "sms"
	NotificationChannelTypeTeams     NotificationChannelType = "teams"
	NotificationChannelTypeDiscord   NotificationChannelType = "discord"
)

// NotificationChannelStatus represents the status of a notification channel
type NotificationChannelStatus string

const (
	NotificationChannelStatusActive   NotificationChannelStatus = "active"
	NotificationChannelStatusInactive NotificationChannelStatus = "inactive"
	NotificationChannelStatusError    NotificationChannelStatus = "error"
)

// NotificationChannelConfig holds channel-specific configuration
type NotificationChannelConfig struct {
	// Email configuration
	EmailTo      []string `json:"email_to,omitempty"`
	EmailFrom    string   `json:"email_from,omitempty"`
	SMTPServer   string   `json:"smtp_server,omitempty"`
	SMTPPort     int      `json:"smtp_port,omitempty"`
	SMTPUsername string   `json:"smtp_username,omitempty"`
	SMTPPassword string   `json:"smtp_password,omitempty"`

	// Slack configuration
	SlackWebhookURL string `json:"slack_webhook_url,omitempty"`
	SlackChannel    string `json:"slack_channel,omitempty"`
	SlackUsername   string `json:"slack_username,omitempty"`

	// Webhook configuration
	WebhookURL     string            `json:"webhook_url,omitempty"`
	WebhookHeaders map[string]string `json:"webhook_headers,omitempty"`
	WebhookMethod  string            `json:"webhook_method,omitempty"`

	// PagerDuty configuration
	PagerDutyIntegrationKey string `json:"pagerduty_integration_key,omitempty"`
	PagerDutyServiceKey     string `json:"pagerduty_service_key,omitempty"`

	// SMS configuration
	SMSProvider string   `json:"sms_provider,omitempty"`
	SMSTo       []string `json:"sms_to,omitempty"`
	SMSFrom     string   `json:"sms_from,omitempty"`

	// Teams configuration
	TeamsWebhookURL string `json:"teams_webhook_url,omitempty"`

	// Discord configuration
	DiscordWebhookURL string `json:"discord_webhook_url,omitempty"`

	// Generic configuration
	CustomAttrs map[string]string `json:"custom_attrs,omitempty"`
}

// StringSlice represents a []string with JSONB support
type StringSlice []string

// Value implements driver.Valuer for StringSlice
func (ss StringSlice) Value() (driver.Value, error) {
	return json.Marshal(ss)
}

// Scan implements sql.Scanner for StringSlice
func (ss *StringSlice) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan StringSlice from non-bytes/string")
	}

	return json.Unmarshal(bytes, ss)
}

// NotificationChannelConfig JSONB methods
func (ncc NotificationChannelConfig) Value() (driver.Value, error) {
	return json.Marshal(ncc)
}

func (ncc *NotificationChannelConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan NotificationChannelConfig from non-bytes/string")
	}

	return json.Unmarshal(bytes, ncc)
}
