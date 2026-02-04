package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/log/internal/models"
	"github.com/minisource/log/internal/service"
)

// AlertHandler handles alert HTTP requests
type AlertHandler struct {
	service *service.AlertService
}

// NewAlertHandler creates a new alert handler
func NewAlertHandler(service *service.AlertService) *AlertHandler {
	return &AlertHandler{service: service}
}

// CreateAlert creates a new alert
// @Summary Create alert
// @Description Creates a new log alert rule
// @Tags alerts
// @Accept json
// @Produce json
// @Param alert body models.LogAlert true "Log Alert"
// @Success 201 {object} models.LogAlert
// @Failure 400 {object} ErrorResponse
// @Router /alerts [post]
func (h *AlertHandler) CreateAlert(c *fiber.Ctx) error {
	var alert models.LogAlert
	if err := c.BodyParser(&alert); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}

	// Set tenant from context
	if tenantID := c.Locals("tenant_id"); tenantID != nil {
		if tid, ok := tenantID.(uuid.UUID); ok {
			alert.TenantID = tid
		}
	}

	if err := h.service.CreateAlert(c.Context(), &alert); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(alert)
}

// UpdateAlert updates an alert
// @Summary Update alert
// @Description Updates an existing alert
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path string true "Alert ID"
// @Param alert body models.LogAlert true "Log Alert"
// @Success 200 {object} models.LogAlert
// @Failure 400 {object} ErrorResponse
// @Router /alerts/{id} [put]
func (h *AlertHandler) UpdateAlert(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid alert ID format",
		})
	}

	var alert models.LogAlert
	if err := c.BodyParser(&alert); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}

	alert.ID = id
	if err := h.service.UpdateAlert(c.Context(), &alert); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(alert)
}

// GetAlert retrieves an alert
// @Summary Get alert
// @Description Retrieves an alert by ID
// @Tags alerts
// @Produce json
// @Param id path string true "Alert ID"
// @Success 200 {object} models.LogAlert
// @Failure 404 {object} ErrorResponse
// @Router /alerts/{id} [get]
func (h *AlertHandler) GetAlert(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid alert ID format",
		})
	}

	alert, err := h.service.GetAlert(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error:   "not_found",
			Message: "Alert not found",
		})
	}

	return c.JSON(alert)
}

// ListAlerts lists alerts for a tenant
// @Summary List alerts
// @Description Lists all alerts for the current tenant
// @Tags alerts
// @Produce json
// @Success 200 {array} models.LogAlert
// @Router /alerts [get]
func (h *AlertHandler) ListAlerts(c *fiber.Ctx) error {
	var tenantID uuid.UUID
	if tid := c.Locals("tenant_id"); tid != nil {
		if t, ok := tid.(uuid.UUID); ok {
			tenantID = t
		}
	}

	alerts, err := h.service.GetAlertsByTenant(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(alerts)
}

// DeleteAlert deletes an alert
// @Summary Delete alert
// @Description Deletes an alert
// @Tags alerts
// @Param id path string true "Alert ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Router /alerts/{id} [delete]
func (h *AlertHandler) DeleteAlert(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid alert ID format",
		})
	}

	if err := h.service.DeleteAlert(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "deletion_failed",
			Message: err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// EnableAlert enables an alert
// @Summary Enable alert
// @Description Enables an alert
// @Tags alerts
// @Param id path string true "Alert ID"
// @Success 204
// @Router /alerts/{id}/enable [post]
func (h *AlertHandler) EnableAlert(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid alert ID format",
		})
	}

	if err := h.service.EnableAlert(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "enable_failed",
			Message: err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DisableAlert disables an alert
// @Summary Disable alert
// @Description Disables an alert
// @Tags alerts
// @Param id path string true "Alert ID"
// @Success 204
// @Router /alerts/{id}/disable [post]
func (h *AlertHandler) DisableAlert(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid alert ID format",
		})
	}

	if err := h.service.DisableAlert(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "disable_failed",
			Message: err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
