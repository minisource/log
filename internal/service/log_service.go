package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/minisource/log/config"
	"github.com/minisource/log/internal/models"
	"github.com/minisource/log/internal/repository"
	"github.com/redis/go-redis/v9"
)

// LogService handles log business logic
type LogService struct {
	logRepo       *repository.LogRepository
	retentionRepo *repository.RetentionRepository
	alertRepo     *repository.AlertRepository
	redis         *redis.Client
	config        *config.Config
	bufferMu      sync.Mutex
	buffer        []models.LogEntry
	flushTicker   *time.Ticker
}

// NewLogService creates a new log service
func NewLogService(
	logRepo *repository.LogRepository,
	retentionRepo *repository.RetentionRepository,
	alertRepo *repository.AlertRepository,
	redisClient *redis.Client,
	cfg *config.Config,
) *LogService {
	svc := &LogService{
		logRepo:       logRepo,
		retentionRepo: retentionRepo,
		alertRepo:     alertRepo,
		redis:         redisClient,
		config:        cfg,
		buffer:        make([]models.LogEntry, 0, 1000),
	}

	// Start background flush
	svc.flushTicker = time.NewTicker(5 * time.Second)
	go svc.backgroundFlush()

	return svc
}

// IngestSingle ingests a single log entry
func (s *LogService) IngestSingle(ctx context.Context, entry *models.LogEntry) error {
	// Set defaults
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	// Check alerts asynchronously
	go s.checkAlerts(context.Background(), *entry)

	return s.logRepo.Create(ctx, entry)
}

// IngestBatch ingests multiple log entries
func (s *LogService) IngestBatch(ctx context.Context, batch *models.LogBatch) error {
	entries := batch.Entries
	now := time.Now().UTC()

	for i := range entries {
		if entries[i].ID == uuid.Nil {
			entries[i].ID = uuid.New()
		}
		if entries[i].Timestamp.IsZero() {
			entries[i].Timestamp = now
		}
	}

	// Check alerts for error/fatal logs
	go func() {
		for _, entry := range entries {
			if entry.Level == models.LogLevelError || entry.Level == models.LogLevelFatal {
				s.checkAlerts(context.Background(), entry)
			}
		}
	}()

	return s.logRepo.CreateBatch(ctx, entries)
}

// BufferLog adds a log to the buffer for batch processing
func (s *LogService) BufferLog(entry models.LogEntry) {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	s.bufferMu.Lock()
	s.buffer = append(s.buffer, entry)
	shouldFlush := len(s.buffer) >= 1000
	s.bufferMu.Unlock()

	if shouldFlush {
		go s.flushBuffer()
	}
}

// flushBuffer writes buffered logs to the database
func (s *LogService) flushBuffer() {
	s.bufferMu.Lock()
	if len(s.buffer) == 0 {
		s.bufferMu.Unlock()
		return
	}
	entries := s.buffer
	s.buffer = make([]models.LogEntry, 0, 1000)
	s.bufferMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.logRepo.CreateBatch(ctx, entries); err != nil {
		// Log error (would normally use structured logging)
		fmt.Printf("Failed to flush log buffer: %v\n", err)
	}
}

// backgroundFlush periodically flushes the buffer
func (s *LogService) backgroundFlush() {
	for range s.flushTicker.C {
		s.flushBuffer()
	}
}

// Query searches for log entries
func (s *LogService) Query(ctx context.Context, filter models.LogFilter) (*models.LogQueryResult, error) {
	// Try cache first for common queries
	cacheKey := s.buildCacheKey(filter)
	if cached, err := s.getCachedResult(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	entries, total, err := s.logRepo.Query(ctx, filter)
	if err != nil {
		return nil, err
	}

	result := &models.LogQueryResult{
		Entries:    entries,
		TotalCount: total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}

	// Cache the result
	s.cacheResult(ctx, cacheKey, result, 30*time.Second)

	return result, nil
}

// GetByID retrieves a single log entry
func (s *LogService) GetByID(ctx context.Context, id uuid.UUID) (*models.LogEntry, error) {
	return s.logRepo.FindByID(ctx, id)
}

// GetByTraceID retrieves all logs for a trace
func (s *LogService) GetByTraceID(ctx context.Context, traceID string) ([]models.LogEntry, error) {
	return s.logRepo.GetByTraceID(ctx, traceID)
}

// GetByRequestID retrieves all logs for a request
func (s *LogService) GetByRequestID(ctx context.Context, requestID string) ([]models.LogEntry, error) {
	return s.logRepo.GetByRequestID(ctx, requestID)
}

// GetStats retrieves aggregated statistics
func (s *LogService) GetStats(ctx context.Context, tenantID *uuid.UUID, startTime, endTime time.Time) (*models.LogStats, error) {
	return s.logRepo.GetStats(ctx, tenantID, startTime, endTime)
}

// Aggregate retrieves time-bucketed aggregations
func (s *LogService) Aggregate(ctx context.Context, filter models.LogFilter, interval string) ([]models.LogAggregation, error) {
	return s.logRepo.Aggregate(ctx, filter, interval)
}

// GetServices returns available service names
func (s *LogService) GetServices(ctx context.Context, tenantID *uuid.UUID) ([]string, error) {
	return s.logRepo.GetServices(ctx, tenantID)
}

// GetStorageSize returns storage usage
func (s *LogService) GetStorageSize(ctx context.Context, tenantID *uuid.UUID) (int64, error) {
	return s.logRepo.GetStorageSize(ctx, tenantID)
}

// Cleanup removes old log entries based on retention policies
func (s *LogService) Cleanup(ctx context.Context) error {
	policies, err := s.retentionRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	// Apply tenant-specific retention
	for _, policy := range policies {
		cutoff := time.Now().AddDate(0, 0, -policy.RetentionDays)
		_, err := s.logRepo.DeleteOlderThan(ctx, &policy.TenantID, cutoff)
		if err != nil {
			fmt.Printf("Failed to cleanup logs for tenant %s: %v\n", policy.TenantID, err)
		}
	}

	// Apply default retention for logs without tenant-specific policy
	defaultCutoff := time.Now().AddDate(0, 0, -s.config.Retention.RetentionDays)
	_, err = s.logRepo.DeleteOlderThan(ctx, nil, defaultCutoff)

	return err
}

// checkAlerts evaluates alerts for the given log entry
func (s *LogService) checkAlerts(ctx context.Context, entry models.LogEntry) {
	alerts, err := s.alertRepo.FindEnabled(ctx)
	if err != nil {
		return
	}

	for _, alert := range alerts {
		if s.matchesAlert(entry, alert) {
			s.triggerAlert(ctx, alert, entry)
		}
	}
}

// matchesAlert checks if a log entry matches an alert filter
func (s *LogService) matchesAlert(entry models.LogEntry, alert models.LogAlert) bool {
	// Parse the filter from JSON
	var filter models.LogFilter
	if err := json.Unmarshal(alert.Filter, &filter); err != nil {
		return false
	}

	if filter.ServiceName != "" && filter.ServiceName != entry.ServiceName {
		return false
	}

	if filter.Level != "" && filter.Level != entry.Level {
		return false
	}

	if filter.TenantID != nil && *filter.TenantID != entry.TenantID {
		return false
	}

	return true
}

// triggerAlert handles alert triggering
func (s *LogService) triggerAlert(ctx context.Context, alert models.LogAlert, entry models.LogEntry) {
	// Rate limit alerts (minimum 1 minute between triggers)
	if alert.LastTriggered != nil && time.Since(*alert.LastTriggered) < time.Minute {
		return
	}

	// Update last triggered
	s.alertRepo.UpdateLastTriggered(ctx, alert.ID)

	// Would normally send to notification channels here
	fmt.Printf("Alert triggered: %s for log: %s\n", alert.Name, entry.Message)
}

// Cache helpers
func (s *LogService) buildCacheKey(filter models.LogFilter) string {
	data, _ := json.Marshal(filter)
	return fmt.Sprintf("log_query:%x", data)
}

func (s *LogService) getCachedResult(ctx context.Context, key string) (*models.LogQueryResult, error) {
	if s.redis == nil {
		return nil, nil
	}

	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var result models.LogQueryResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *LogService) cacheResult(ctx context.Context, key string, result *models.LogQueryResult, ttl time.Duration) {
	if s.redis == nil {
		return
	}

	data, err := json.Marshal(result)
	if err != nil {
		return
	}

	s.redis.Set(ctx, key, data, ttl)
}

// Close cleans up resources
func (s *LogService) Close() {
	if s.flushTicker != nil {
		s.flushTicker.Stop()
	}
	s.flushBuffer()
}
