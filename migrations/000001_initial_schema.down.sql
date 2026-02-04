-- Drop triggers
DROP TRIGGER IF EXISTS update_log_alerts_updated_at ON log_alerts;
DROP TRIGGER IF EXISTS update_log_retention_policies_updated_at ON log_retention_policies;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS log_alerts;
DROP TABLE IF EXISTS log_retention_policies;
DROP TABLE IF EXISTS log_entries;
