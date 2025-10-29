package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for our application
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	InfluxDB InfluxDBConfig `mapstructure:"influxdb"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	NATS     NATSConfig     `mapstructure:"nats"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Security SecurityConfig `mapstructure:"security"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Services ServicesConfig `mapstructure:"services"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// InfluxDBConfig holds InfluxDB configuration
type InfluxDBConfig struct {
	URL    string `mapstructure:"url"`
	Token  string `mapstructure:"token"`
	Org    string `mapstructure:"org"`
	Bucket string `mapstructure:"bucket"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	GroupID string   `mapstructure:"group_id"`
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL string `mapstructure:"url"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	Expiration int    `mapstructure:"expiration"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	JWTSecret         string `mapstructure:"jwt_secret"`
	RateLimitRPS      int    `mapstructure:"rate_limit_rps"`
	RateLimitBurst    int    `mapstructure:"rate_limit_burst"`
	EnableAuditLogging bool   `mapstructure:"enable_audit_logging"`
	EncryptionKey     string `mapstructure:"encryption_key"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// ServicesConfig holds microservices configuration
type ServicesConfig struct {
	NetworkService      ServiceConfig `mapstructure:"network_service"`
	MonitoringService   ServiceConfig `mapstructure:"monitoring_service"`
	OrchestrationService ServiceConfig `mapstructure:"orchestration_service"`
	ValidationService   ServiceConfig `mapstructure:"validation_service"`
}

// ServiceConfig holds individual service configuration
type ServiceConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.idle_timeout", 120)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "netorchestrator")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.dbname", "netorchestrator")
	viper.SetDefault("database.sslmode", "disable")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// InfluxDB defaults
	viper.SetDefault("influxdb.url", "http://localhost:8086")
	viper.SetDefault("influxdb.token", "")
	viper.SetDefault("influxdb.org", "netorchestrator")
	viper.SetDefault("influxdb.bucket", "network_metrics")

	// Kafka defaults
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.group_id", "netorchestrator")

	// NATS defaults
	viper.SetDefault("nats.url", "nats://localhost:4222")

	// JWT defaults
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expiration", 3600)

	// Security defaults
	viper.SetDefault("security.jwt_secret", "netorchestrator-super-secret-jwt-key-2024")
	viper.SetDefault("security.rate_limit_rps", 100)
	viper.SetDefault("security.rate_limit_burst", 10)
	viper.SetDefault("security.enable_audit_logging", true)
	viper.SetDefault("security.encryption_key", "netorchestrator-encryption-key-32-byte")

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	// Services defaults
	viper.SetDefault("services.network_service.host", "localhost")
	viper.SetDefault("services.network_service.port", 8081)
	viper.SetDefault("services.monitoring_service.host", "localhost")
	viper.SetDefault("services.monitoring_service.port", 8082)
	viper.SetDefault("services.orchestration_service.host", "localhost")
	viper.SetDefault("services.orchestration_service.port", 8083)
	viper.SetDefault("services.validation_service.host", "localhost")
	viper.SetDefault("services.validation_service.port", 8084)
}
