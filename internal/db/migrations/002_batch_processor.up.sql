-- Create batch_jobs table
CREATE TABLE IF NOT EXISTS batch_jobs (
    id VARCHAR(255) PRIMARY KEY,
    org_id VARCHAR(255) NOT NULL,
    workspace_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    activity_count INTEGER NOT NULL DEFAULT 0,
    success_count INTEGER NOT NULL DEFAULT 0,
    error_count INTEGER NOT NULL DEFAULT 0,
    total_emissions NUMERIC(15,2) NOT NULL DEFAULT 0.0,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    priority INTEGER NOT NULL DEFAULT 5,
    locked_by VARCHAR(255) NULL,
    locked_until TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_org_id CHECK (org_id != ''),
    CONSTRAINT fk_workspace_id CHECK (workspace_id != '')
);

-- Create batch_activity_refs table
CREATE TABLE IF NOT EXISTS batch_activity_refs (
    batch_id VARCHAR(255) NOT NULL REFERENCES batch_jobs(id) ON DELETE CASCADE,
    activity_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (batch_id, activity_id)
);

-- Create batch_progress_log table
CREATE TABLE IF NOT EXISTS batch_progress_log (
    id BIGSERIAL PRIMARY KEY,
    batch_id VARCHAR(255) NOT NULL REFERENCES batch_jobs(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    processed_count INTEGER NOT NULL DEFAULT 0,
    error_count INTEGER NOT NULL DEFAULT 0,
    total_emissions NUMERIC(15,2) NOT NULL DEFAULT 0.0,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_batch_jobs_org_id ON batch_jobs(org_id);
CREATE INDEX IF NOT EXISTS idx_batch_jobs_workspace_id ON batch_jobs(workspace_id);
CREATE INDEX IF NOT EXISTS idx_batch_jobs_status ON batch_jobs(status);
CREATE INDEX IF NOT EXISTS idx_batch_jobs_created_at ON batch_jobs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_batch_jobs_org_status ON batch_jobs(org_id, status);
CREATE INDEX IF NOT EXISTS idx_batch_jobs_pending ON batch_jobs(status) WHERE status IN ('pending', 'queued');
CREATE INDEX IF NOT EXISTS idx_batch_activity_refs_batch_id ON batch_activity_refs(batch_id);
CREATE INDEX IF NOT EXISTS idx_batch_progress_log_batch_id ON batch_progress_log(batch_id);

-- Create view for batch job status details
CREATE OR REPLACE VIEW batch_job_status_view AS
SELECT 
    id,
    org_id,
    workspace_id,
    status,
    activity_count,
    success_count,
    error_count,
    CASE 
        WHEN activity_count = 0 THEN 0.0
        ELSE ROUND((CAST(success_count + error_count AS FLOAT) / CAST(activity_count AS FLOAT)) * 100, 2)
    END as progress_percent,
    CASE 
        WHEN activity_count = 0 THEN 0.0
        ELSE ROUND(total_emissions / CAST(activity_count AS FLOAT), 4)
    END as avg_emissions_per_activity,
    total_emissions,
    EXTRACT(EPOCH FROM (COALESCE(completed_at, NOW()) - started_at))::INTEGER as duration_seconds,
    retry_count,
    max_retries,
    priority,
    created_at,
    updated_at,
    locked_by,
    locked_until
FROM batch_jobs;

-- Create view for batch statistics
CREATE OR REPLACE VIEW batch_statistics_view AS
SELECT 
    org_id,
    COUNT(*) as total_batches,
    SUM(activity_count) as total_activities,
    SUM(success_count) as total_success,
    SUM(error_count) as total_errors,
    SUM(total_emissions) as total_emissions,
    ROUND(AVG(total_emissions), 2) as avg_emissions_per_batch,
    SUM(CASE WHEN status = 'complete' THEN 1 ELSE 0 END) as completed_batches,
    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_batches,
    SUM(CASE WHEN status IN ('pending', 'queued') THEN 1 ELSE 0 END) as pending_batches
FROM batch_jobs
GROUP BY org_id;
