package handler

import (
	"github.com/gofiber/fiber/v2"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health returns basic health status
// @Summary Health check
// @Description Returns service health status
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "log-service",
	})
}

// Ready returns readiness status
// @Summary Readiness check
// @Description Returns service readiness status
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Success 503 {object} map[string]string
// @Router /ready [get]
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	// Add actual readiness checks here (database, redis, etc.)
	return c.JSON(fiber.Map{
		"status": "ready",
	})
}

// Live returns liveness status
// @Summary Liveness check
// @Description Returns service liveness status
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /live [get]
func (h *HealthHandler) Live(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "alive",
	})
}
