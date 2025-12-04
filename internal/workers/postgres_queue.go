package workers

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// PostgresQueue implements Queue using PostgreSQL
type PostgresQueue struct {
	db *sql.DB
}

// NewPostgresQueue creates a new PostgreSQL-backed job queue
func NewPostgresQueue(db *sql.DB) (*PostgresQueue, error) {
	queue := &PostgresQueue{db: db}
	if err := queue.createTables(); err != nil {
		return nil, err
	}
	return queue, nil
}

func (q *PostgresQueue) createTables() error {
	schema := `
		CREATE TABLE IF NOT EXISTS jobs (
			id VARCHAR(255) PRIMARY KEY,
			type VARCHAR(100) NOT NULL,
			status VARCHAR(50) NOT NULL,
			tenant_id VARCHAR(255) NOT NULL,
			payload JSONB,
			result JSONB,
			error TEXT,
			attempts INT NOT NULL DEFAULT 0,
			max_attempts INT NOT NULL DEFAULT 3,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			scheduled_at TIMESTAMP,
			completed_at TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_jobs_type_status ON jobs(type, status, scheduled_at);
		CREATE INDEX IF NOT EXISTS idx_jobs_tenant ON jobs(tenant_id, status);
		CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at DESC);
	`

	_, err := q.db.Exec(schema)
	return err
}

// Enqueue adds a job to the queue
func (q *PostgresQueue) Enqueue(ctx context.Context, job *Job) error {
	payloadJSON, _ := json.Marshal(job.Payload)

	query := `
		INSERT INTO jobs (
			id, type, status, tenant_id, payload, attempts, max_attempts,
			created_at, updated_at, scheduled_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	scheduledAt := job.ScheduledAt
	if scheduledAt.IsZero() {
		scheduledAt = time.Now()
	}

	_, err := q.db.ExecContext(ctx, query,
		job.ID,
		job.Type,
		job.Status,
		job.TenantID,
		payloadJSON,
		job.Attempts,
		job.MaxAttempts,
		job.CreatedAt,
		job.UpdatedAt,
		scheduledAt,
	)

	return err
}

// Dequeue retrieves the next pending job of the specified type
func (q *PostgresQueue) Dequeue(ctx context.Context, jobType JobType) (*Job, error) {
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Lock and fetch next job
	query := `
		SELECT id, type, status, tenant_id, payload, result, error, attempts, max_attempts,
		       created_at, updated_at, scheduled_at, completed_at
		FROM jobs
		WHERE type = $1 
		  AND status = $2
		  AND (scheduled_at IS NULL OR scheduled_at <= NOW())
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`

	var job Job
	var payloadJSON, resultJSON sql.NullString
	var errorStr sql.NullString
	var scheduledAt, completedAt sql.NullTime

	err = tx.QueryRowContext(ctx, query, jobType, JobStatusPending).Scan(
		&job.ID,
		&job.Type,
		&job.Status,
		&job.TenantID,
		&payloadJSON,
		&resultJSON,
		&errorStr,
		&job.Attempts,
		&job.MaxAttempts,
		&job.CreatedAt,
		&job.UpdatedAt,
		&scheduledAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No jobs available
	}
	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if payloadJSON.Valid {
		json.Unmarshal([]byte(payloadJSON.String), &job.Payload)
	}
	if resultJSON.Valid {
		json.Unmarshal([]byte(resultJSON.String), &job.Result)
	}
	if errorStr.Valid {
		job.Error = errorStr.String
	}
	if scheduledAt.Valid {
		job.ScheduledAt = scheduledAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	// Update status to processing
	updateQuery := `
		UPDATE jobs
		SET status = $1, attempts = attempts + 1, updated_at = $2
		WHERE id = $3
	`

	_, err = tx.ExecContext(ctx, updateQuery, JobStatusProcessing, time.Now(), job.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	job.Attempts++
	job.Status = JobStatusProcessing
	job.UpdatedAt = time.Now()

	return &job, nil
}

// Complete marks a job as completed
func (q *PostgresQueue) Complete(ctx context.Context, jobID string, result map[string]interface{}) error {
	resultJSON, _ := json.Marshal(result)
	now := time.Now()

	query := `
		UPDATE jobs
		SET status = $1, result = $2, updated_at = $3, completed_at = $4
		WHERE id = $5
	`

	_, err := q.db.ExecContext(ctx, query, JobStatusCompleted, resultJSON, now, now, jobID)
	return err
}

// Fail marks a job as failed
func (q *PostgresQueue) Fail(ctx context.Context, jobID string, jobErr error) error {
	query := `
		UPDATE jobs
		SET status = $1, error = $2, updated_at = $3, completed_at = $4
		WHERE id = $5
	`

	now := time.Now()
	_, err := q.db.ExecContext(ctx, query, JobStatusFailed, jobErr.Error(), now, now, jobID)
	return err
}

// Retry schedules a job for retry
func (q *PostgresQueue) Retry(ctx context.Context, jobID string, delay time.Duration) error {
	scheduledAt := time.Now().Add(delay)

	query := `
		UPDATE jobs
		SET status = $1, scheduled_at = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := q.db.ExecContext(ctx, query, JobStatusPending, scheduledAt, time.Now(), jobID)
	return err
}

// GetJob retrieves a job by ID
func (q *PostgresQueue) GetJob(ctx context.Context, jobID string) (*Job, error) {
	query := `
		SELECT id, type, status, tenant_id, payload, result, error, attempts, max_attempts,
		       created_at, updated_at, scheduled_at, completed_at
		FROM jobs
		WHERE id = $1
	`

	var job Job
	var payloadJSON, resultJSON sql.NullString
	var errorStr sql.NullString
	var scheduledAt, completedAt sql.NullTime

	err := q.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID,
		&job.Type,
		&job.Status,
		&job.TenantID,
		&payloadJSON,
		&resultJSON,
		&errorStr,
		&job.Attempts,
		&job.MaxAttempts,
		&job.CreatedAt,
		&job.UpdatedAt,
		&scheduledAt,
		&completedAt,
	)

	if err != nil {
		return nil, err
	}

	if payloadJSON.Valid {
		json.Unmarshal([]byte(payloadJSON.String), &job.Payload)
	}
	if resultJSON.Valid {
		json.Unmarshal([]byte(resultJSON.String), &job.Result)
	}
	if errorStr.Valid {
		job.Error = errorStr.String
	}
	if scheduledAt.Valid {
		job.ScheduledAt = scheduledAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return &job, nil
}

// ListJobs lists jobs for a tenant
func (q *PostgresQueue) ListJobs(ctx context.Context, tenantID string, status JobStatus, limit int) ([]*Job, error) {
	query := `
		SELECT id, type, status, tenant_id, payload, result, error, attempts, max_attempts,
		       created_at, updated_at, scheduled_at, completed_at
		FROM jobs
		WHERE tenant_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := q.db.QueryContext(ctx, query, tenantID, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]*Job, 0)
	for rows.Next() {
		var job Job
		var payloadJSON, resultJSON sql.NullString
		var errorStr sql.NullString
		var scheduledAt, completedAt sql.NullTime

		err := rows.Scan(
			&job.ID,
			&job.Type,
			&job.Status,
			&job.TenantID,
			&payloadJSON,
			&resultJSON,
			&errorStr,
			&job.Attempts,
			&job.MaxAttempts,
			&job.CreatedAt,
			&job.UpdatedAt,
			&scheduledAt,
			&completedAt,
		)
		if err != nil {
			continue
		}

		if payloadJSON.Valid {
			json.Unmarshal([]byte(payloadJSON.String), &job.Payload)
		}
		if resultJSON.Valid {
			json.Unmarshal([]byte(resultJSON.String), &job.Result)
		}
		if errorStr.Valid {
			job.Error = errorStr.String
		}
		if scheduledAt.Valid {
			job.ScheduledAt = scheduledAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}

		jobs = append(jobs, &job)
	}

	return jobs, nil
}
