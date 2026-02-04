package database

import (
	"fmt"
	"time"

	"github.com/minisource/log/config"
	"github.com/minisource/log/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB creates a new PostgreSQL connection
func NewPostgresDB(cfg config.PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	logLevel := logger.Silent
	if cfg.LogLevel == "debug" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetimeMinutes) * time.Minute)

	return db, nil
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.LogEntry{},
		&models.LogRetention{},
		&models.LogAlert{},
	)
}

// CreateIndexes creates additional database indexes
func CreateIndexes(db *gorm.DB) error {
	// Create composite indexes for common queries
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_logs_tenant_service_time 
         ON log_entries (tenant_id, service_name, timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_level_time 
         ON log_entries (level, timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_metadata_gin 
         ON log_entries USING gin (metadata jsonb_path_ops)`,
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// CreatePartitions sets up table partitioning for log_entries
func CreatePartitions(db *gorm.DB) error {
	// Check if table is already partitioned
	var count int64
	db.Raw(`
		SELECT COUNT(*) FROM pg_inherits 
		WHERE inhparent = 'log_entries'::regclass
	`).Scan(&count)

	if count > 0 {
		return nil // Already partitioned
	}

	// For a proper partitioning setup, you would need to:
	// 1. Create a new partitioned table
	// 2. Migrate data from the old table
	// 3. Drop the old table and rename
	// This is typically handled via proper migrations

	return nil
}
