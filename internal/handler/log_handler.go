package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/log/internal/models"
	"github.com/minisource/log/internal/service"
)

// LogHandler handles log HTTP requests
type LogHandler struct {
	logService *service.LogService
}

// NewLogHandler creates a new log handler
func NewLogHandler(logService *service.LogService) *LogHandler {
	return &LogHandler{logService: logService}
}

// IngestSingle handles single log ingestion
// @Summary Ingest a single log entry
// @Description Ingests a single log entry
// @Tags logs
// @Accept json
// @Produce json
// @Param log body models.LogEntry true "Log Entry"
// @Success 201 {object} models.LogEntry
// @Failure 400 {object} ErrorResponse
// @Router /logs [post]
func (h *LogHandler) IngestSingle(c *fiber.Ctx) error {
	var entry models.LogEntry
	if err := c.BodyParser(&entry); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}

	// Set tenant from context if available
	if tenantID := c.Locals("tenant_id"); tenantID != nil {
		if tid, ok := tenantID.(uuid.UUID); ok {
			entry.TenantID = tid
		}
	}

	if err := h.logService.IngestSingle(c.Context(), &entry); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "ingestion_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(entry)
}

// IngestBatch handles batch log ingestion
// @Summary Ingest multiple log entries
// @Description Ingests a batch of log entries
// @Tags logs
// @Accept json
// @Produce json
// @Param logs body models.LogBatch true "Log Batch"
// @Success 201 {object} map[string]int
// @Failure 400 {object} ErrorResponse
// @Router /logs/batch [post]
func (h *LogHandler) IngestBatch(c *fiber.Ctx) error {
	var batch models.LogBatch
	if err := c.BodyParser(&batch); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}

	// Set tenant from context if available
	if tenantID := c.Locals("tenant_id"); tenantID != nil {
		if tid, ok := tenantID.(uuid.UUID); ok {
			for i := range batch.Entries {
				if batch.Entries[i].TenantID == uuid.Nil {
					batch.Entries[i].TenantID = tid
				}
			}
		}
	}

	if err := h.logService.IngestBatch(c.Context(), &batch); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "ingestion_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"count": len(batch.Entries),
	})
}

// Query handles log search/filtering
// @Summary Query logs
// @Description Search and filter logs
// @Tags logs
// @Accept json
// @Produce json
// @Param filter body models.LogFilter true "Log Filter"
// @Success 200 {object} models.LogQueryResult
// @Failure 400 {object} ErrorResponse
// @Router /logs/query [post]
func (h *LogHandler) Query(c *fiber.Ctx) error {
	var filter models.LogFilter
	if err := c.BodyParser(&filter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}

	// Apply tenant from context
	if tenantID := c.Locals("tenant_id"); tenantID != nil {
		if tid, ok := tenantID.(uuid.UUID); ok {
			filter.TenantID = &tid
		}
	}

	result, err := h.logService.Query(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(result)
}

// GetByID retrieves a single log entry
// @Summary Get log by ID
// @Description Retrieves a single log entry by ID
// @Tags logs
// @Produce json
// @Param id path string true "Log ID"
// @Success 200 {object} models.LogEntry
// @Failure 404 {object} ErrorResponse
// @Router /logs/{id} [get]
func (h *LogHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid log ID format",
		})
	}

	entry, err := h.logService.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error:   "not_found",
			Message: "Log entry not found",
		})
	}

	return c.JSON(entry)
}

// GetByTrace retrieves logs by trace ID
// @Summary Get logs by trace ID
// @Description Retrieves all logs for a distributed trace
// @Tags logs
// @Produce json
// @Param trace_id path string true "Trace ID"
// @Success 200 {array} models.LogEntry
// @Router /logs/trace/{trace_id} [get]
func (h *LogHandler) GetByTrace(c *fiber.Ctx) error {
	traceID := c.Params("trace_id")
	if traceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_trace_id",
			Message: "Trace ID is required",
		})
	}

	entries, err := h.logService.GetByTraceID(c.Context(), traceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(entries)
}

// GetByRequest retrieves logs by request ID
// @Summary Get logs by request ID
// @Description Retrieves all logs for a request
// @Tags logs
// @Produce json
// @Param request_id path string true "Request ID"
// @Success 200 {array} models.LogEntry
// @Router /logs/request/{request_id} [get]
func (h *LogHandler) GetByRequest(c *fiber.Ctx) error {
	requestID := c.Params("request_id")
	if requestID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request_id",
			Message: "Request ID is required",
		})
	}

	entries, err := h.logService.GetByRequestID(c.Context(), requestID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(entries)
}

// GetStats retrieves log statistics
// @Summary Get log statistics
// @Description Retrieves aggregated log statistics
// @Tags logs
// @Produce json
// @Param start query string false "Start time (RFC3339)"
// @Param end query string false "End time (RFC3339)"
// @Success 200 {object} models.LogStats
// @Router /logs/stats [get]
func (h *LogHandler) GetStats(c *fiber.Ctx) error {
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	if s := c.Query("start"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			startTime = t
		}
	}
	if e := c.Query("end"); e != "" {
		if t, err := time.Parse(time.RFC3339, e); err == nil {
			endTime = t
		}
	}

	var tenantID *uuid.UUID
	if tid := c.Locals("tenant_id"); tid != nil {
		if t, ok := tid.(uuid.UUID); ok {
			tenantID = &t
		}
	}

	stats, err := h.logService.GetStats(c.Context(), tenantID, startTime, endTime)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "stats_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(stats)
}

// Aggregate retrieves time-bucketed aggregations
// @Summary Aggregate logs
// @Description Retrieves time-bucketed log aggregations
// @Tags logs
// @Accept json
// @Produce json
// @Param filter body models.LogFilter true "Log Filter"
// @Param interval query string false "Time interval (minute, hour, day)"
// @Success 200 {array} models.LogAggregation
// @Router /logs/aggregate [post]
func (h *LogHandler) Aggregate(c *fiber.Ctx) error {
	var filter models.LogFilter
	if err := c.BodyParser(&filter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}

	interval := c.Query("interval", "hour")

	aggregations, err := h.logService.Aggregate(c.Context(), filter, interval)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "aggregation_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(aggregations)
}

// GetServices retrieves available service names
// @Summary Get service names
// @Description Retrieves list of services that have logged entries
// @Tags logs
// @Produce json
// @Success 200 {array} string
// @Router /logs/services [get]
func (h *LogHandler) GetServices(c *fiber.Ctx) error {
	var tenantID *uuid.UUID
	if tid := c.Locals("tenant_id"); tid != nil {
		if t, ok := tid.(uuid.UUID); ok {
			tenantID = &t
		}
	}

	services, err := h.logService.GetServices(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(services)
}

// GetStorage retrieves storage usage
// @Summary Get storage usage
// @Description Retrieves storage usage statistics
// @Tags logs
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /logs/storage [get]
func (h *LogHandler) GetStorage(c *fiber.Ctx) error {
	var tenantID *uuid.UUID
	if tid := c.Locals("tenant_id"); tid != nil {
		if t, ok := tid.(uuid.UUID); ok {
			tenantID = &t
		}
	}

	size, err := h.logService.GetStorageSize(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"size_bytes": size,
		"size_mb":    float64(size) / (1024 * 1024),
		"size_gb":    float64(size) / (1024 * 1024 * 1024),
	})
}

// Stream handles real-time log streaming via SSE
// @Summary Stream logs
// @Description Stream logs in real-time using Server-Sent Events
// @Tags logs
// @Produce text/event-stream
// @Param service query string false "Filter by service"
// @Param level query string false "Filter by log level"
// @Success 200 {string} string "SSE stream"
// @Router /logs/stream [get]
func (h *LogHandler) Stream(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	service := c.Query("service")
	level := models.LogLevel(c.Query("level"))

	// Create filter
	filter := models.LogFilter{
		ServiceName: service,
		Level:       level,
	}

	// Apply tenant from context
	if tenantID := c.Locals("tenant_id"); tenantID != nil {
		if tid, ok := tenantID.(uuid.UUID); ok {
			filter.TenantID = &tid
		}
	}

	// Start streaming
	ctx := c.Context()
	lastCheck := time.Now()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Check for new logs since last check
			filter.StartTime = &lastCheck
			result, err := h.logService.Query(c.Context(), filter)
			if err == nil && len(result.Entries) > 0 {
				for _, entry := range result.Entries {
					c.Writef("data: %s\n\n", entry.Message)
				}
			}
			lastCheck = time.Now()
			time.Sleep(1 * time.Second)
		}
	}
}

// List handles simple log listing
// @Summary List logs
// @Description List logs with optional filters
// @Tags logs
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param service query string false "Filter by service"
// @Param level query string false "Filter by log level"
// @Success 200 {object} models.LogQueryResult
// @Router /logs [get]
func (h *LogHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "100"))

	filter := models.LogFilter{
		ServiceName: c.Query("service"),
		Level:       models.LogLevel(c.Query("level")),
		Page:        page,
		PageSize:    pageSize,
	}

	// Apply tenant from context
	if tenantID := c.Locals("tenant_id"); tenantID != nil {
		if tid, ok := tenantID.(uuid.UUID); ok {
			filter.TenantID = &tid
		}
	}

	result, err := h.logService.Query(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(result)
}
