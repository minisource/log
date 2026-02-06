package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/log/internal/models"
	"gorm.io/gorm"
)

// LogRepository handles log entry persistence
type LogRepository struct {
	db *gorm.DB
}

// NewLogRepository creates a new log repository
func NewLogRepository(db *gorm.DB) *LogRepository {
	return &LogRepository{db: db}
}

// Create inserts a single log entry
func (r *LogRepository) Create(ctx context.Context, entry *models.LogEntry) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

// CreateBatch inserts multiple log entries
func (r *LogRepository) CreateBatch(ctx context.Context, entries []models.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(entries, 1000).Error
}

// FindByID retrieves a log entry by ID
func (r *LogRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.LogEntry, error) {
	var entry models.LogEntry
	err := r.db.WithContext(ctx).First(&entry, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// Query finds log entries matching the filter
func (r *LogRepository) Query(ctx context.Context, filter models.LogFilter) ([]models.LogEntry, int64, error) {
	var entries []models.LogEntry
	var total int64

	query := r.buildQuery(filter)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	err := query.Order("timestamp DESC").Offset(offset).Limit(pageSize).Find(&entries).Error
	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

// buildQuery creates the GORM query from filter
func (r *LogRepository) buildQuery(filter models.LogFilter) *gorm.DB {
	query := r.db.Model(&models.LogEntry{})

	if filter.TenantID != nil {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}

	if filter.ServiceName != "" {
		query = query.Where("service_name = ?", filter.ServiceName)
	}

	if filter.Level != "" {
		query = query.Where("level = ?", filter.Level)
	}

	if filter.MinLevel != "" {
		levels := getLevelsAtOrAbove(filter.MinLevel)
		query = query.Where("level IN ?", levels)
	}

	if filter.StartTime != nil {
		query = query.Where("timestamp >= ?", filter.StartTime)
	}

	if filter.EndTime != nil {
		query = query.Where("timestamp <= ?", filter.EndTime)
	}

	if filter.TraceID != "" {
		query = query.Where("trace_id = ?", filter.TraceID)
	}

	if filter.UserID != nil {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.RequestID != "" {
		query = query.Where("request_id = ?", filter.RequestID)
	}

	if filter.Environment != "" {
		query = query.Where("environment = ?", filter.Environment)
	}

	if filter.Search != "" {
		search := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(message) LIKE ?", search)
	}

	return query
}

// getLevelsAtOrAbove returns all log levels at or above the given level
func getLevelsAtOrAbove(level models.LogLevel) []models.LogLevel {
	levels := []models.LogLevel{
		models.LogLevelDebug,
		models.LogLevelInfo,
		models.LogLevelWarn,
		models.LogLevelError,
		models.LogLevelFatal,
	}

	var result []models.LogLevel
	found := false
	for _, l := range levels {
		if l == level {
			found = true
		}
		if found {
			result = append(result, l)
		}
	}
	return result
}

// GetStats retrieves aggregated statistics
func (r *LogRepository) GetStats(ctx context.Context, tenantID *uuid.UUID, startTime, endTime time.Time) (*models.LogStats, error) {
	stats := &models.LogStats{
		LevelCounts:   make(map[models.LogLevel]int64),
		ServiceCounts: make(map[string]int64),
		TimeRange: models.TimeRange{
			Start: startTime,
			End:   endTime,
		},
	}

	query := r.db.WithContext(ctx).Model(&models.LogEntry{}).
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime)

	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// Total count
	query.Count(&stats.TotalCount)

	// Level counts
	var levelResults []struct {
		Level models.LogLevel
		Count int64
	}
	r.db.WithContext(ctx).Model(&models.LogEntry{}).
		Select("level, COUNT(*) as count").
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Group("level").Scan(&levelResults)

	for _, lr := range levelResults {
		stats.LevelCounts[lr.Level] = lr.Count
	}

	// Service counts
	var serviceResults []struct {
		ServiceName string
		Count       int64
	}
	r.db.WithContext(ctx).Model(&models.LogEntry{}).
		Select("service_name, COUNT(*) as count").
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Group("service_name").Scan(&serviceResults)

	for _, sr := range serviceResults {
		stats.ServiceCounts[sr.ServiceName] = sr.Count
	}

	return stats, nil
}

// Aggregate retrieves aggregated log counts over time
func (r *LogRepository) Aggregate(ctx context.Context, filter models.LogFilter, interval string) ([]models.LogAggregation, error) {
	var bucketExpr string
	switch interval {
	case "minute":
		bucketExpr = "date_trunc('minute', timestamp)"
	case "hour":
		bucketExpr = "date_trunc('hour', timestamp)"
	case "day":
		bucketExpr = "date_trunc('day', timestamp)"
	default:
		bucketExpr = "date_trunc('hour', timestamp)"
	}

	query := r.buildQuery(filter)

	var results []struct {
		Bucket time.Time
		Count  int64
	}

	err := query.Select(fmt.Sprintf("%s as bucket, COUNT(*) as count", bucketExpr)).
		Group("bucket").
		Order("bucket").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	aggregations := make([]models.LogAggregation, len(results))
	for i, res := range results {
		aggregations[i] = models.LogAggregation{
			Bucket: res.Bucket,
			Count:  res.Count,
		}
	}

	return aggregations, nil
}

// DeleteOlderThan removes log entries older than the specified time
func (r *LogRepository) DeleteOlderThan(ctx context.Context, tenantID *uuid.UUID, before time.Time) (int64, error) {
	query := r.db.WithContext(ctx).Where("timestamp < ?", before)

	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID)
	}

	result := query.Delete(&models.LogEntry{})
	return result.RowsAffected, result.Error
}

// GetByTraceID retrieves all log entries for a trace
func (r *LogRepository) GetByTraceID(ctx context.Context, traceID string) ([]models.LogEntry, error) {
	var entries []models.LogEntry
	err := r.db.WithContext(ctx).
		Where("trace_id = ?", traceID).
		Order("timestamp ASC").
		Find(&entries).Error
	return entries, err
}

// GetByRequestID retrieves all log entries for a request
func (r *LogRepository) GetByRequestID(ctx context.Context, requestID string) ([]models.LogEntry, error) {
	var entries []models.LogEntry
	err := r.db.WithContext(ctx).
		Where("request_id = ?", requestID).
		Order("timestamp ASC").
		Find(&entries).Error
	return entries, err
}

// GetServices returns distinct service names
func (r *LogRepository) GetServices(ctx context.Context, tenantID *uuid.UUID) ([]string, error) {
	var services []string
	query := r.db.WithContext(ctx).Model(&models.LogEntry{}).
		Distinct("service_name")

	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID)
	}

	err := query.Pluck("service_name", &services).Error
	return services, err
}

// GetStorageSize returns approximate storage size in bytes
func (r *LogRepository) GetStorageSize(ctx context.Context, tenantID *uuid.UUID) (int64, error) {
	var size int64
	query := `SELECT pg_total_relation_size('log_entries')`

	if tenantID != nil {
		// Estimate based on row count ratio
		var total, tenantTotal int64
		r.db.Model(&models.LogEntry{}).Count(&total)
		r.db.Model(&models.LogEntry{}).Where("tenant_id = ?", tenantID).Count(&tenantTotal)

		if total > 0 {
			r.db.Raw(query).Scan(&size)
			size = size * tenantTotal / total
		}
		return size, nil
	}

	err := r.db.Raw(query).Scan(&size).Error
	return size, err
}
