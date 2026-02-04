-- Revert partitioning (data will be lost if partitions exist)
-- This is a destructive migration

DROP FUNCTION IF EXISTS create_monthly_partition();

-- Drop partitioned table
DROP TABLE IF EXISTS log_entries CASCADE;

-- Recreate non-partitioned table
CREATE TABLE IF NOT EXISTS log_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    service_name VARCHAR(100) NOT NULL,
    level VARCHAR(10) NOT NULL,
    message TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    trace_id VARCHAR(64),
    span_id VARCHAR(32),
    user_id UUID,
    request_id VARCHAR(64),
    metadata JSONB,
    source VARCHAR(255),
    host VARCHAR(255),
    environment VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_logs_tenant_time ON log_entries (tenant_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_service ON log_entries (service_name);
CREATE INDEX IF NOT EXISTS idx_logs_level ON log_entries (level);
CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON log_entries (timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_trace ON log_entries (trace_id) WHERE trace_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_logs_request ON log_entries (request_id) WHERE request_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_logs_user ON log_entries (user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_logs_env ON log_entries (environment) WHERE environment IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_logs_tenant_service_time ON log_entries (tenant_id, service_name, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_level_time ON log_entries (level, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_metadata_gin ON log_entries USING gin (metadata jsonb_path_ops);
