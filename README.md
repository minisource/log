# Log Microservice

Centralized logging microservice for the Minisource platform. Provides log ingestion, storage, querying, and alerting capabilities.

## Features

- **High-throughput ingestion**: Single and batch log ingestion APIs
- **Flexible querying**: Filter by service, level, time range, trace ID, etc.
- **Distributed tracing**: Correlation with OpenTelemetry trace IDs
- **Multi-tenancy**: Tenant isolation for log data
- **Retention policies**: Per-tenant configurable retention
- **Alerting**: Threshold-based alerting rules
- **Real-time streaming**: SSE-based log streaming
- **Storage optimization**: PostgreSQL with table partitioning

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Log Service (Port 5002)                  │
├─────────────────────────────────────────────────────────────┤
│  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌──────────┐ │
│  │  Ingest   │  │   Query   │  │   Stats   │  │  Alerts  │ │
│  │  Handler  │  │  Handler  │  │  Handler  │  │  Handler │ │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  └────┬─────┘ │
│        │              │              │             │        │
│  ┌─────┴──────────────┴──────────────┴─────────────┴─────┐ │
│  │                    Log Service                         │ │
│  │  • Buffered batch writes                               │ │
│  │  • Query caching (Redis)                               │ │
│  │  • Alert evaluation                                    │ │
│  │  • Retention cleanup                                   │ │
│  └─────────────────────────┬─────────────────────────────┘ │
│                            │                                │
│  ┌─────────────────────────┴─────────────────────────────┐ │
│  │                   Log Repository                       │ │
│  └─────────────────────────┬─────────────────────────────┘ │
└────────────────────────────┼────────────────────────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
       ┌──────┴──────┐ ┌─────┴─────┐ ┌──────┴──────┐
       │  PostgreSQL │ │   Redis   │ │   Jaeger    │
       │  (Storage)  │ │  (Cache)  │ │  (Tracing)  │
       └─────────────┘ └───────────┘ └─────────────┘
```

## API Endpoints

### Log Ingestion

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/logs` | Ingest single log entry |
| POST | `/api/v1/logs/batch` | Ingest batch of log entries |

### Log Querying

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/logs` | List logs with pagination |
| POST | `/api/v1/logs/query` | Advanced log query |
| GET | `/api/v1/logs/:id` | Get log by ID |
| GET | `/api/v1/logs/trace/:trace_id` | Get logs by trace ID |
| GET | `/api/v1/logs/request/:request_id` | Get logs by request ID |

### Statistics & Aggregation

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/logs/stats` | Get log statistics |
| POST | `/api/v1/logs/aggregate` | Time-bucketed aggregations |
| GET | `/api/v1/logs/services` | List available services |
| GET | `/api/v1/logs/storage` | Get storage usage |

### Real-time Streaming

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/logs/stream` | SSE log streaming |

### Retention Policies

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/retention` | List retention policies |
| POST | `/api/v1/retention` | Create retention policy |
| GET | `/api/v1/retention/tenant/:tenant_id` | Get tenant policy |
| PUT | `/api/v1/retention/:id` | Update retention policy |
| DELETE | `/api/v1/retention/:id` | Delete retention policy |

### Alerts

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/alerts` | List alerts |
| POST | `/api/v1/alerts` | Create alert |
| GET | `/api/v1/alerts/:id` | Get alert |
| PUT | `/api/v1/alerts/:id` | Update alert |
| DELETE | `/api/v1/alerts/:id` | Delete alert |
| POST | `/api/v1/alerts/:id/enable` | Enable alert |
| POST | `/api/v1/alerts/:id/disable` | Disable alert |

### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/ready` | Readiness probe |
| GET | `/live` | Liveness probe |

## Log Entry Structure

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tenant_id": "123e4567-e89b-12d3-a456-426614174000",
  "service_name": "auth-service",
  "level": "ERROR",
  "message": "Failed to authenticate user",
  "timestamp": "2024-01-15T10:30:00Z",
  "trace_id": "abc123def456",
  "span_id": "span789",
  "user_id": "user-uuid",
  "request_id": "req-123",
  "metadata": {
    "error_code": "AUTH_001",
    "ip_address": "192.168.1.1"
  },
  "source": "handler/auth.go:45",
  "host": "pod-abc123",
  "environment": "production"
}
```

## Log Levels

- `DEBUG`: Detailed debugging information
- `INFO`: General informational messages
- `WARN`: Warning messages
- `ERROR`: Error conditions
- `FATAL`: Critical errors causing service failure

## Query Filter

```json
{
  "tenant_id": "uuid",
  "service_name": "auth-service",
  "level": "ERROR",
  "min_level": "WARN",
  "start_time": "2024-01-15T00:00:00Z",
  "end_time": "2024-01-15T23:59:59Z",
  "trace_id": "abc123",
  "user_id": "uuid",
  "request_id": "req-123",
  "search": "authentication failed",
  "environment": "production",
  "page": 1,
  "page_size": 100
}
```

## Configuration

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `5002` |
| `POSTGRES_HOST` | PostgreSQL host | `localhost` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | PostgreSQL user | `log_user` |
| `POSTGRES_PASSWORD` | PostgreSQL password | - |
| `POSTGRES_DB` | PostgreSQL database | `log_db` |
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `LOG_RETENTION_DAYS` | Default log retention | `30` |
| `LOG_MAX_SIZE_GB` | Maximum storage size | `50` |

## Quick Start

### Development

```bash
# Start development stack
make dev-up

# Run locally
make run

# Run tests
make test
```

### Production

```bash
# Build Docker image
make docker-build

# Deploy with docker-compose
docker-compose -f docker-compose.prod.yml up -d
```

## Database Schema

The service uses PostgreSQL with the following tables:

- `log_entries`: Main log storage (partitioned by month)
- `log_retention_policies`: Per-tenant retention configuration
- `log_alerts`: Alert rule definitions

## Performance Considerations

1. **Batch Ingestion**: Use batch API for high-volume logging
2. **Query Caching**: Common queries are cached in Redis
3. **Partitioning**: Logs are partitioned by month for efficient cleanup
4. **Indexes**: Optimized indexes for common query patterns
5. **Connection Pooling**: Configurable database connection pool

## Integration

### From Go Services

```go
import "github.com/minisource/log/pkg/client"

logger := client.NewLogClient("http://log-service:5002")
logger.Log(ctx, client.LogEntry{
    ServiceName: "my-service",
    Level:       "INFO",
    Message:     "Operation completed",
    Metadata:    map[string]interface{}{"duration_ms": 150},
})
```

### From HTTP

```bash
curl -X POST http://localhost:5002/api/v1/logs \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: tenant-uuid" \
  -d '{
    "service_name": "my-service",
    "level": "INFO",
    "message": "Hello World"
  }'
```

## License

MIT License