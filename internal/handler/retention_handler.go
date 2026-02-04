package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minisource/log/internal/models"
	"github.com/minisource/log/internal/service"
)

// RetentionHandler handles retention policy HTTP requests
type RetentionHandler struct {
	service *service.RetentionService
}

// NewRetentionHandler creates a new retention handler
func NewRetentionHandler(service *service.RetentionService) *RetentionHandler {
	return &RetentionHandler{service: service}
}

// CreatePolicy creates a new retention policy
// @Summary Create retention policy
// @Description Creates a new retention policy for a tenant
// @Tags retention
// @Accept json
// @Produce json
// @Param policy body models.LogRetention true "Retention Policy"
// @Success 201 {object} models.LogRetention
// @Failure 400 {object} ErrorResponse
// @Router /retention [post]
func (h *RetentionHandler) CreatePolicy(c *fiber.Ctx) error {
	var policy models.LogRetention
	if err := c.BodyParser(&policy); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}

	if err := h.service.CreatePolicy(c.Context(), &policy); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(policy)
}

// UpdatePolicy updates a retention policy
// @Summary Update retention policy
// @Description Updates an existing retention policy
// @Tags retention
// @Accept json
// @Produce json
// @Param id path string true "Policy ID"
// @Param policy body models.LogRetention true "Retention Policy"
// @Success 200 {object} models.LogRetention
// @Failure 400 {object} ErrorResponse
// @Router /retention/{id} [put]
func (h *RetentionHandler) UpdatePolicy(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid policy ID format",
		})
	}

	var policy models.LogRetention
	if err := c.BodyParser(&policy); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
	}

	policy.ID = id
	if err := h.service.UpdatePolicy(c.Context(), &policy); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(policy)
}

// GetPolicy retrieves a retention policy
// @Summary Get retention policy
// @Description Retrieves retention policy for a tenant
// @Tags retention
// @Produce json
// @Param tenant_id path string true "Tenant ID"
// @Success 200 {object} models.LogRetention
// @Failure 404 {object} ErrorResponse
// @Router /retention/tenant/{tenant_id} [get]
func (h *RetentionHandler) GetPolicy(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenant_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_tenant_id",
			Message: "Invalid tenant ID format",
		})
	}

	policy, err := h.service.GetPolicy(c.Context(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error:   "not_found",
			Message: "Retention policy not found",
		})
	}

	return c.JSON(policy)
}

// ListPolicies lists all retention policies
// @Summary List retention policies
// @Description Lists all retention policies
// @Tags retention
// @Produce json
// @Success 200 {array} models.LogRetention
// @Router /retention [get]
func (h *RetentionHandler) ListPolicies(c *fiber.Ctx) error {
	policies, err := h.service.GetAllPolicies(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(policies)
}

// DeletePolicy deletes a retention policy
// @Summary Delete retention policy
// @Description Deletes a retention policy
// @Tags retention
// @Param id path string true "Policy ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Router /retention/{id} [delete]
func (h *RetentionHandler) DeletePolicy(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid policy ID format",
		})
	}

	if err := h.service.DeletePolicy(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "deletion_failed",
			Message: err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
