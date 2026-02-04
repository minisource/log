-- Create log_entries table with proper indexes
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

-- Create indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_logs_tenant_time ON log_entries (tenant_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_service ON log_entries (service_name);
CREATE INDEX IF NOT EXISTS idx_logs_level ON log_entries (level);
CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON log_entries (timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_trace ON log_entries (trace_id) WHERE trace_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_logs_request ON log_entries (request_id) WHERE request_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_logs_user ON log_entries (user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_logs_env ON log_entries (environment) WHERE environment IS NOT NULL;

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_logs_tenant_service_time ON log_entries (tenant_id, service_name, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_level_time ON log_entries (level, timestamp DESC);

-- GIN index for JSONB metadata queries
CREATE INDEX IF NOT EXISTS idx_logs_metadata_gin ON log_entries USING gin (metadata jsonb_path_ops);

-- Create log_retention_policies table
CREATE TABLE IF NOT EXISTS log_retention_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID UNIQUE NOT NULL,
    retention_days INTEGER DEFAULT 30,
    max_size_gb INTEGER DEFAULT 10,
    archive_enabled BOOLEAN DEFAULT FALSE,
    archive_path VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create log_alerts table
CREATE TABLE IF NOT EXISTS log_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    filter JSONB NOT NULL,
    threshold INTEGER NOT NULL,
    window_mins INTEGER NOT NULL DEFAULT 5,
    severity VARCHAR(20) NOT NULL,
    channels JSONB,
    last_triggered TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alerts_tenant ON log_alerts (tenant_id);
CREATE INDEX IF NOT EXISTS idx_alerts_enabled ON log_alerts (enabled) WHERE enabled = TRUE;

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
DROP TRIGGER IF EXISTS update_log_retention_policies_updated_at ON log_retention_policies;
CREATE TRIGGER update_log_retention_policies_updated_at
    BEFORE UPDATE ON log_retention_policies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_log_alerts_updated_at ON log_alerts;
CREATE TRIGGER update_log_alerts_updated_at
    BEFORE UPDATE ON log_alerts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
