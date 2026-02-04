package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestID adds a unique request ID to each request
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Locals("request_id", requestID)
		c.Set("X-Request-ID", requestID)
		return c.Next()
	}
}

// TenantExtractor extracts tenant ID from headers
func TenantExtractor() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenantIDStr := c.Get("X-Tenant-ID")
		if tenantIDStr != "" {
			if tenantID, err := uuid.Parse(tenantIDStr); err == nil {
				c.Locals("tenant_id", tenantID)
			}
		}
		return c.Next()
	}
}

// SecurityHeaders adds security headers
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		return c.Next()
	}
}

// RequestLogger logs request details
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)

		// Log request (would use structured logger in production)
		_ = duration // suppress unused warning

		return err
	}
}

// ContentType ensures JSON content type
func ContentType() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Next()
	}
}
