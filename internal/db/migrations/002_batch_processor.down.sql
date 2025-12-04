-- Drop views
DROP VIEW IF EXISTS batch_statistics_view;
DROP VIEW IF EXISTS batch_job_status_view;

-- Drop indexes
DROP INDEX IF EXISTS idx_batch_progress_log_batch_id;
DROP INDEX IF EXISTS idx_batch_activity_refs_batch_id;
DROP INDEX IF EXISTS idx_batch_jobs_pending;
DROP INDEX IF EXISTS idx_batch_jobs_org_status;
DROP INDEX IF EXISTS idx_batch_jobs_created_at;
DROP INDEX IF EXISTS idx_batch_jobs_status;
DROP INDEX IF EXISTS idx_batch_jobs_workspace_id;
DROP INDEX IF EXISTS idx_batch_jobs_org_id;

-- Drop tables
DROP TABLE IF EXISTS batch_progress_log;
DROP TABLE IF EXISTS batch_activity_refs;
DROP TABLE IF EXISTS batch_jobs;
