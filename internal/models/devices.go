package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeviceType represents different types of network devices
type DeviceType string

const (
	DeviceTypeCisco   DeviceType = "cisco"
	DeviceTypeJuniper DeviceType = "juniper"
	DeviceTypeArista  DeviceType = "arista"
)

// DeviceStatus represents the operational status of a device
type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "online"
	DeviceStatusOffline DeviceStatus = "offline"
	DeviceStatusUnknown DeviceStatus = "unknown"
	DeviceStatusError   DeviceStatus = "error"
)

// DeviceConfig represents device configuration as JSONB
type DeviceConfig map[string]interface{}

// Scan implements the Scanner interface for database reads
func (dc *DeviceConfig) Scan(value interface{}) error {
	if value == nil {
		*dc = make(DeviceConfig)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, dc)
	case string:
		return json.Unmarshal([]byte(v), dc)
	default:
		return errors.New("cannot scan into DeviceConfig")
	}
}

// Value implements the Valuer interface for database writes
func (dc DeviceConfig) Value() (driver.Value, error) {
	if dc == nil {
		return nil, nil
	}
	return json.Marshal(dc)
}

// NetworkDevice represents a managed network device
type NetworkDevice struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string         `json:"name" gorm:"not null"`
	Type      DeviceType     `json:"type" gorm:"type:varchar(20);not null"`
	Vendor    string         `json:"vendor" gorm:"not null"`
	Model     string         `json:"model"`
	Host      string         `json:"host" gorm:"not null"`
	Username  string         `json:"username" gorm:"not null"`
	Password  string         `json:"password" gorm:"not null"` // In production, encrypt this
	Config    DeviceConfig   `json:"config" gorm:"type:jsonb"`
	Status    DeviceStatus   `json:"status" gorm:"type:varchar(20);default:'unknown'"`
	LastSeen  *time.Time     `json:"last_seen,omitempty"`
	NetworkID *uuid.UUID     `json:"network_id,omitempty" gorm:"type:uuid"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Network *Network `json:"network,omitempty" gorm:"foreignKey:NetworkID"`
}

// TableName specifies the table name for NetworkDevice
func (NetworkDevice) TableName() string {
	return "network_devices"
}

// BeforeCreate hook to set default values
func (d *NetworkDevice) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if d.Status == "" {
		d.Status = DeviceStatusUnknown
	}
	return nil
}

// IsOnline returns true if the device is currently online
func (d *NetworkDevice) IsOnline() bool {
	return d.Status == DeviceStatusOnline
}

// GetConnectionString returns the connection string for the device
func (d *NetworkDevice) GetConnectionString() string {
	return d.Host
}

// SetLastSeen updates the last seen timestamp
func (d *NetworkDevice) SetLastSeen() {
	now := time.Now()
	d.LastSeen = &now
}

// DeviceInterface represents configuration for a network interface
type DeviceInterface struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DeviceID    uuid.UUID      `json:"device_id" gorm:"type:uuid;not null"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Type        string         `json:"type"` // ethernet, loopback, tunnel, etc.
	IPAddress   string         `json:"ip_address"`
	SubnetMask  string         `json:"subnet_mask"`
	Gateway     string         `json:"gateway"`
	MTU         int            `json:"mtu"`
	Speed       string         `json:"speed"`
	Duplex      string         `json:"duplex"`
	Enabled     bool           `json:"enabled" gorm:"default:true"`
	Config      DeviceConfig   `json:"config" gorm:"type:jsonb"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Device *NetworkDevice `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
}

// TableName specifies the table name for DeviceInterface
func (DeviceInterface) TableName() string {
	return "device_interfaces"
}

// BeforeCreate hook to set default values
func (di *DeviceInterface) BeforeCreate(tx *gorm.DB) error {
	if di.ID == uuid.Nil {
		di.ID = uuid.New()
	}
	return nil
}

// DeviceMetric represents metrics collected from network devices
type DeviceMetric struct {
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DeviceID    uuid.UUID    `json:"device_id" gorm:"type:uuid;not null"`
	MetricName  string       `json:"metric_name" gorm:"not null"`
	MetricValue float64      `json:"metric_value"`
	Unit        string       `json:"unit"`
	Tags        DeviceConfig `json:"tags" gorm:"type:jsonb"`
	Timestamp   time.Time    `json:"timestamp" gorm:"not null"`
	CreatedAt   time.Time    `json:"created_at"`

	// Relationships
	Device *NetworkDevice `json:"device,omitempty" gorm:"foreignKey:DeviceID"`
}

// TableName specifies the table name for DeviceMetric
func (DeviceMetric) TableName() string {
	return "device_metrics"
}

// BeforeCreate hook to set default values
func (dm *DeviceMetric) BeforeCreate(tx *gorm.DB) error {
	if dm.ID == uuid.Nil {
		dm.ID = uuid.New()
	}
	if dm.Timestamp.IsZero() {
		dm.Timestamp = time.Now()
	}
	return nil
}
