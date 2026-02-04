package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/minisource/log/config"
	"github.com/minisource/log/internal/database"
	"github.com/minisource/log/internal/handler"
	"github.com/minisource/log/internal/middleware"
	"github.com/minisource/log/internal/repository"
	"github.com/minisource/log/internal/router"
	"github.com/minisource/log/internal/service"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewPostgresDB(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create indexes
	if err := database.CreateIndexes(db); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	// Initialize Redis
	var redisClient *redis.Client
	if cfg.Redis.Host != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Printf("Warning: Redis connection failed: %v", err)
			redisClient = nil
		}
	}

	// Initialize repositories
	logRepo := repository.NewLogRepository(db)
	retentionRepo := repository.NewRetentionRepository(db)
	alertRepo := repository.NewAlertRepository(db)

	// Initialize services
	logService := service.NewLogService(logRepo, retentionRepo, alertRepo, redisClient, cfg)
	retentionService := service.NewRetentionService(retentionRepo)
	alertService := service.NewAlertService(alertRepo)

	// Initialize handlers
	logHandler := handler.NewLogHandler(logService)
	retentionHandler := handler.NewRetentionHandler(retentionService)
	alertHandler := handler.NewAlertHandler(alertService)
	healthHandler := handler.NewHealthHandler()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Log Service",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		BodyLimit:    10 * 1024 * 1024, // 10MB for batch ingestion
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(compress.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-ID,X-Tenant-ID",
	}))
	app.Use(middleware.RequestID())
	app.Use(middleware.TenantExtractor())
	app.Use(middleware.SecurityHeaders())
	app.Use(middleware.ContentType())

	// Setup routes
	router.SetupRoutes(app, logHandler, retentionHandler, alertHandler, healthHandler)

	// Start cleanup scheduler
	go startCleanupScheduler(logService, cfg)

	// Start server
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		log.Printf("Starting Log Service on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Log Service...")

	// Close services
	logService.Close()

	// Shutdown app with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	// Close Redis
	if redisClient != nil {
		redisClient.Close()
	}

	// Close database
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	log.Println("Log Service stopped")
}

// startCleanupScheduler runs periodic log cleanup
func startCleanupScheduler(logService *service.LogService, cfg *config.Config) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
			if err := logService.Cleanup(ctx); err != nil {
				log.Printf("Cleanup failed: %v", err)
			}
			cancel()
		}
	}
}
