package database

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"netorchestrator/internal/config"
)

// PostgreSQL represents a PostgreSQL database connection
type PostgreSQL struct {
	DB     *gorm.DB
	config config.DatabaseConfig
	logger *zap.Logger
}

// NewPostgreSQL creates a new PostgreSQL connection
func NewPostgreSQL(cfg config.DatabaseConfig) (*PostgreSQL, error) {
	// Build connection string
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	// Configure GORM logger
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	postgres := &PostgreSQL{
		DB:     db,
		config: cfg,
		logger: zap.NewNop(), // Will be set later if needed
	}

	return postgres, nil
}

// SetLogger sets the logger for the database connection
func (p *PostgreSQL) SetLogger(logger *zap.Logger) {
	p.logger = logger
}

// Close closes the database connection
func (p *PostgreSQL) Close() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// GetDB returns the underlying GORM database instance
func (p *PostgreSQL) GetDB() *gorm.DB {
	return p.DB
}

// Health checks the database connection health
func (p *PostgreSQL) Health() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Ping()
}
