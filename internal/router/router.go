package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minisource/log/internal/handler"
)

// SetupRoutes configures all API routes
func SetupRoutes(
	app *fiber.App,
	logHandler *handler.LogHandler,
	retentionHandler *handler.RetentionHandler,
	alertHandler *handler.AlertHandler,
	healthHandler *handler.HealthHandler,
) {
	// Health endpoints
	app.Get("/health", healthHandler.Health)
	app.Get("/ready", healthHandler.Ready)
	app.Get("/live", healthHandler.Live)

	// API v1
	api := app.Group("/api/v1")

	// Log endpoints
	logs := api.Group("/logs")
	logs.Get("/", logHandler.List)
	logs.Post("/", logHandler.IngestSingle)
	logs.Post("/batch", logHandler.IngestBatch)
	logs.Post("/query", logHandler.Query)
	logs.Get("/stats", logHandler.GetStats)
	logs.Post("/aggregate", logHandler.Aggregate)
	logs.Get("/services", logHandler.GetServices)
	logs.Get("/storage", logHandler.GetStorage)
	logs.Get("/stream", logHandler.Stream)
	logs.Get("/trace/:trace_id", logHandler.GetByTrace)
	logs.Get("/request/:request_id", logHandler.GetByRequest)
	logs.Get("/:id", logHandler.GetByID)

	// Retention policy endpoints
	retention := api.Group("/retention")
	retention.Get("/", retentionHandler.ListPolicies)
	retention.Post("/", retentionHandler.CreatePolicy)
	retention.Get("/tenant/:tenant_id", retentionHandler.GetPolicy)
	retention.Put("/:id", retentionHandler.UpdatePolicy)
	retention.Delete("/:id", retentionHandler.DeletePolicy)

	// Alert endpoints
	alerts := api.Group("/alerts")
	alerts.Get("/", alertHandler.ListAlerts)
	alerts.Post("/", alertHandler.CreateAlert)
	alerts.Get("/:id", alertHandler.GetAlert)
	alerts.Put("/:id", alertHandler.UpdateAlert)
	alerts.Delete("/:id", alertHandler.DeleteAlert)
	alerts.Post("/:id/enable", alertHandler.EnableAlert)
	alerts.Post("/:id/disable", alertHandler.DisableAlert)
}
