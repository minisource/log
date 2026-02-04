package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Postgres  PostgresConfig
	Redis     RedisConfig
	Logging   LoggingConfig
	Tracing   TracingConfig
	Retention RetentionConfig
}

type ServerConfig struct {
	Port            string
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type PostgresConfig struct {
	Host               string
	Port               string
	User               string
	Password           string
	DBName             string
	SSLMode            string
	MaxOpenConns       int
	MaxIdleConns       int
	MaxLifetimeMinutes int
	LogLevel           string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type LoggingConfig struct {
	Level  string
	Format string
}

type TracingConfig struct {
	Enabled     bool
	ServiceName string
	Endpoint    string
	SampleRate  float64
}

type RetentionConfig struct {
	Days           int
	RetentionDays  int
	MaxSizeGB      int
	CleanupEnabled bool
	CleanupCron    string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "5002"),
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:     getDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Postgres: PostgresConfig{
			Host:               getEnv("DB_HOST", "localhost"),
			Port:               getEnv("DB_PORT", "5432"),
			User:               getEnv("DB_USER", "postgres"),
			Password:           getEnv("DB_PASSWORD", "postgres"),
			DBName:             getEnv("DB_NAME", "minisource_logs"),
			SSLMode:            getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:       getEnvInt("DB_MAX_OPEN_CONNS", 50),
			MaxIdleConns:       getEnvInt("DB_MAX_IDLE_CONNS", 10),
			MaxLifetimeMinutes: getEnvInt("DB_MAX_LIFETIME_MINS", 30),
			LogLevel:           getEnv("DB_LOG_LEVEL", "info"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 1),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Tracing: TracingConfig{
			Enabled:     getEnvBool("TRACING_ENABLED", true),
			ServiceName: getEnv("SERVICE_NAME", "minisource-log"),
			Endpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			SampleRate:  getEnvFloat("TRACING_SAMPLE_RATE", 1.0),
		},
		Retention: RetentionConfig{
			Days:           getEnvInt("LOG_RETENTION_DAYS", 30),
			RetentionDays:  getEnvInt("LOG_RETENTION_DAYS", 30),
			MaxSizeGB:      getEnvInt("LOG_MAX_SIZE_GB", 50),
			CleanupEnabled: getEnvBool("LOG_CLEANUP_ENABLED", true),
			CleanupCron:    getEnv("LOG_CLEANUP_CRON", "0 2 * * *"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
