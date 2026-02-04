-- Convert log_entries to partitioned table by month
-- Note: This migration requires PostgreSQL 11+

-- Step 1: Rename existing table
ALTER TABLE IF EXISTS log_entries RENAME TO log_entries_old;

-- Step 2: Create partitioned table
CREATE TABLE log_entries (
    id UUID DEFAULT gen_random_uuid(),
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
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Step 3: Create partitions for the next 12 months
DO $$
DECLARE
    start_date DATE := DATE_TRUNC('month', CURRENT_DATE);
    end_date DATE;
    partition_name TEXT;
BEGIN
    FOR i IN 0..11 LOOP
        end_date := start_date + INTERVAL '1 month';
        partition_name := 'log_entries_' || TO_CHAR(start_date, 'YYYY_MM');
        
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF log_entries FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date
        );
        
        start_date := end_date;
    END LOOP;
END $$;

-- Step 4: Create indexes on partitioned table
CREATE INDEX idx_logs_tenant_time ON log_entries (tenant_id, timestamp DESC);
CREATE INDEX idx_logs_service ON log_entries (service_name);
CREATE INDEX idx_logs_level ON log_entries (level);
CREATE INDEX idx_logs_trace ON log_entries (trace_id) WHERE trace_id IS NOT NULL;
CREATE INDEX idx_logs_request ON log_entries (request_id) WHERE request_id IS NOT NULL;
CREATE INDEX idx_logs_tenant_service_time ON log_entries (tenant_id, service_name, timestamp DESC);
CREATE INDEX idx_logs_metadata_gin ON log_entries USING gin (metadata jsonb_path_ops);

-- Step 5: Migrate data from old table
INSERT INTO log_entries SELECT * FROM log_entries_old;

-- Step 6: Drop old table
DROP TABLE log_entries_old;

-- Create function to auto-create monthly partitions
CREATE OR REPLACE FUNCTION create_monthly_partition()
RETURNS VOID AS $$
DECLARE
    next_month DATE := DATE_TRUNC('month', CURRENT_DATE + INTERVAL '1 month');
    end_date DATE := next_month + INTERVAL '1 month';
    partition_name TEXT := 'log_entries_' || TO_CHAR(next_month, 'YYYY_MM');
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_class WHERE relname = partition_name
    ) THEN
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF log_entries FOR VALUES FROM (%L) TO (%L)',
            partition_name, next_month, end_date
        );
    END IF;
END;
$$ LANGUAGE plpgsql;
