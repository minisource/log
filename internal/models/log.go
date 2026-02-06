package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// LogLevel represents log severity
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelFatal LogLevel = "FATAL"
)

// LogEntry represents a single log entry
type LogEntry struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID       `json:"tenant_id" gorm:"type:uuid;index:idx_logs_tenant_time"`
	ServiceName string          `json:"service_name" gorm:"type:varchar(100);index:idx_logs_service"`
	Level       LogLevel        `json:"level" gorm:"type:varchar(10);index:idx_logs_level"`
	Message     string          `json:"message" gorm:"type:text"`
	Timestamp   time.Time       `json:"timestamp" gorm:"index:idx_logs_tenant_time;index:idx_logs_timestamp"`
	TraceID     string          `json:"trace_id,omitempty" gorm:"type:varchar(64);index:idx_logs_trace"`
	SpanID      string          `json:"span_id,omitempty" gorm:"type:varchar(32)"`
	UserID      *uuid.UUID      `json:"user_id,omitempty" gorm:"type:uuid;index:idx_logs_user"`
	RequestID   string          `json:"request_id,omitempty" gorm:"type:varchar(64);index:idx_logs_request"`
	Metadata    json.RawMessage `json:"metadata,omitempty" gorm:"type:jsonb"`
	Source      string          `json:"source,omitempty" gorm:"type:varchar(255)"`
	Host        string          `json:"host,omitempty" gorm:"type:varchar(255)"`
	Environment string          `json:"environment,omitempty" gorm:"type:varchar(50);index:idx_logs_env"`
	CreatedAt   time.Time       `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the table name for GORM
func (LogEntry) TableName() string {
	return "log_entries"
}

// LogBatch represents a batch of log entries for bulk ingestion
type LogBatch struct {
	Entries []LogEntry `json:"entries"`
}

// LogFilter defines query filters for logs
type LogFilter struct {
	TenantID    *uuid.UUID `json:"tenant_id,omitempty"`
	ServiceName string     `json:"service_name,omitempty"`
	Level       LogLevel   `json:"level,omitempty"`
	MinLevel    LogLevel   `json:"min_level,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	TraceID     string     `json:"trace_id,omitempty"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	RequestID   string     `json:"request_id,omitempty"`
	Search      string     `json:"search,omitempty"`
	Environment string     `json:"environment,omitempty"`
	Page        int        `json:"page,omitempty"`
	PageSize    int        `json:"page_size,omitempty"`
}

// LogStats represents aggregated log statistics
type LogStats struct {
	TotalCount    int64              `json:"total_count"`
	LevelCounts   map[LogLevel]int64 `json:"level_counts"`
	ServiceCounts map[string]int64   `json:"service_counts"`
	TimeRange     TimeRange          `json:"time_range"`
}

// TimeRange represents a time range
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// LogAggregation represents aggregated log data
type LogAggregation struct {
	Bucket      time.Time          `json:"bucket"`
	Count       int64              `json:"count"`
	LevelCounts map[LogLevel]int64 `json:"level_counts,omitempty"`
}

// LogRetention defines retention policy
type LogRetention struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID       uuid.UUID `json:"tenant_id" gorm:"type:uuid;uniqueIndex"`
	RetentionDays  int       `json:"retention_days" gorm:"default:30"`
	MaxSizeGB      int       `json:"max_size_gb" gorm:"default:10"`
	ArchiveEnabled bool      `json:"archive_enabled" gorm:"default:false"`
	ArchivePath    string    `json:"archive_path,omitempty" gorm:"type:varchar(500)"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (LogRetention) TableName() string {
	return "log_retention_policies"
}

// LogAlert defines alerting rules for logs
type LogAlert struct {
	ID            uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID      uuid.UUID       `json:"tenant_id" gorm:"type:uuid;index"`
	Name          string          `json:"name" gorm:"type:varchar(255);not null"`
	Description   string          `json:"description,omitempty" gorm:"type:text"`
	Enabled       bool            `json:"enabled" gorm:"default:true"`
	Filter        json.RawMessage `json:"filter" gorm:"type:jsonb;not null"`
	Threshold     int             `json:"threshold" gorm:"not null"`
	WindowMins    int             `json:"window_mins" gorm:"not null;default:5"`
	Severity      string          `json:"severity" gorm:"type:varchar(20);not null"`
	Channels      json.RawMessage `json:"channels" gorm:"type:jsonb"`
	LastTriggered *time.Time      `json:"last_triggered,omitempty"`
	CreatedAt     time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (LogAlert) TableName() string {
	return "log_alerts"
}

// LogQueryResult represents paginated query results
type LogQueryResult struct {
	Entries    []LogEntry `json:"entries"`
	TotalCount int64      `json:"total_count"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	HasMore    bool       `json:"has_more"`
}
