package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

// BatchStore defines the interface for batch persistence
type BatchStore interface {
	CreateBatch(ctx context.Context, job BatchJob) (string, error)
	GetBatch(ctx context.Context, batchID string) (*BatchJob, error)
	ListBatches(ctx context.Context, orgID string, filter BatchFilter) ([]BatchJob, error)
	UpdateBatchStatus(ctx context.Context, batchID string, status JobStatus, progress *BatchProgress) error
	AddActivityRef(ctx context.Context, batchID, activityID string) error
	MarkActivityComplete(ctx context.Context, batchID, activityID string, record interface{}) error
	MarkActivityFailed(ctx context.Context, batchID, activityID string, errMsg string) error
	GetPendingBatches(ctx context.Context, limit int) ([]BatchJob, error)
	AcquireBatchLock(ctx context.Context, batchID, workerID string, timeout time.Duration) (bool, error)
	ReleaseBatchLock(ctx context.Context, batchID string) error
	DeleteBatch(ctx context.Context, batchID string) error
	GetProgressLog(ctx context.Context, batchID string) ([]BatchProgressLog, error)
}

// PostgresBatchStore implements BatchStore using PostgreSQL
type PostgresBatchStore struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewPostgresBatchStore creates a new PostgreSQL batch store
func NewPostgresBatchStore(db *sql.DB, logger *slog.Logger) BatchStore {
	return &PostgresBatchStore{
		db:     db,
		logger: logger,
	}
}

// CreateBatch creates a new batch job
func (s *PostgresBatchStore) CreateBatch(ctx context.Context, job BatchJob) (string, error) {
	query := `
		INSERT INTO batch_jobs (
			id, org_id, workspace_id, status, activity_count, success_count,
			error_count, total_emissions, retry_count, max_retries, priority,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`

	var batchID string
	now := time.Now()
	job.CreatedAt = now
	job.UpdatedAt = now

	err := s.db.QueryRowContext(ctx, query,
		job.ID, job.OrgID, job.WorkspaceID, job.Status.String(),
		job.ActivityCount, job.SuccessCount, job.ErrorCount,
		job.TotalEmissions, job.RetryCount, job.MaxRetries,
		job.Priority, job.CreatedAt, job.UpdatedAt,
	).Scan(&batchID)

	if err != nil {
		s.logger.Error("failed to create batch", "error", err, "batch_id", job.ID)
		return "", err
	}

	s.logger.Debug("batch created", "batch_id", batchID, "org_id", job.OrgID)
	return batchID, nil
}

// GetBatch retrieves a batch by ID
func (s *PostgresBatchStore) GetBatch(ctx context.Context, batchID string) (*BatchJob, error) {
	query := `
		SELECT id, org_id, workspace_id, status, activity_count, success_count,
			   error_count, total_emissions, started_at, completed_at, error_message,
			   retry_count, max_retries, priority, created_at, updated_at,
			   locked_by, locked_until
		FROM batch_jobs
		WHERE id = $1
	`

	job := &BatchJob{}
	err := s.db.QueryRowContext(ctx, query, batchID).Scan(
		&job.ID, &job.OrgID, &job.WorkspaceID, &job.Status,
		&job.ActivityCount, &job.SuccessCount, &job.ErrorCount,
		&job.TotalEmissions, &job.StartedAt, &job.CompletedAt,
		&job.ErrorMessage, &job.RetryCount, &job.MaxRetries,
		&job.Priority, &job.CreatedAt, &job.UpdatedAt,
		&job.LockedBy, &job.LockedUntil,
	)

	if err == sql.ErrNoRows {
		return nil, ErrBatchNotFound
	}
	if err != nil {
		s.logger.Error("failed to get batch", "error", err, "batch_id", batchID)
		return nil, err
	}

	return job, nil
}

// ListBatches lists batches with filtering
func (s *PostgresBatchStore) ListBatches(ctx context.Context, orgID string, filter BatchFilter) ([]BatchJob, error) {
	query := `
		SELECT id, org_id, workspace_id, status, activity_count, success_count,
			   error_count, total_emissions, started_at, completed_at, error_message,
			   retry_count, max_retries, priority, created_at, updated_at,
			   locked_by, locked_until
		FROM batch_jobs
		WHERE org_id = $1
	`

	args := []interface{}{orgID}
	argNum := 2

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, filter.Status)
		argNum++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filter.Limit)
		argNum++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("failed to list batches", "error", err, "org_id", orgID)
		return nil, err
	}
	defer rows.Close()

	var batches []BatchJob
	for rows.Next() {
		job := BatchJob{}
		err := rows.Scan(
			&job.ID, &job.OrgID, &job.WorkspaceID, &job.Status,
			&job.ActivityCount, &job.SuccessCount, &job.ErrorCount,
			&job.TotalEmissions, &job.StartedAt, &job.CompletedAt,
			&job.ErrorMessage, &job.RetryCount, &job.MaxRetries,
			&job.Priority, &job.CreatedAt, &job.UpdatedAt,
			&job.LockedBy, &job.LockedUntil,
		)
		if err != nil {
			s.logger.Error("failed to scan batch", "error", err)
			continue
		}
		batches = append(batches, job)
	}

	return batches, rows.Err()
}

// UpdateBatchStatus updates the status of a batch
func (s *PostgresBatchStore) UpdateBatchStatus(ctx context.Context, batchID string, status JobStatus, progress *BatchProgress) error {
	query := `
		UPDATE batch_jobs
		SET status = $1, updated_at = $2
	`
	args := []interface{}{status.String(), time.Now()}
	argNum := 3

	if progress != nil {
		query += fmt.Sprintf(", success_count = $%d, error_count = $%d, total_emissions = $%d",
			argNum, argNum+1, argNum+2)
		args = append(args, progress.SuccessCount, progress.ErrorCount, progress.TotalEmissions)
		argNum += 3
	}

	if status == JobStatusProcessing {
		query += fmt.Sprintf(", started_at = $%d", argNum)
		args = append(args, time.Now())
		argNum++
	}

	if status == JobStatusComplete || status == JobStatusFailed {
		query += fmt.Sprintf(", completed_at = $%d", argNum)
		args = append(args, time.Now())
	}

	query += fmt.Sprintf(" WHERE id = $%d", argNum)
	args = append(args, batchID)

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("failed to update batch status", "error", err, "batch_id", batchID)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrBatchNotFound
	}

	return nil
}

// AddActivityRef adds an activity reference to a batch
func (s *PostgresBatchStore) AddActivityRef(ctx context.Context, batchID, activityID string) error {
	query := `
		INSERT INTO batch_activity_refs (batch_id, activity_id, status, created_at)
		VALUES ($1, $2, 'pending', $3)
	`

	_, err := s.db.ExecContext(ctx, query, batchID, activityID, time.Now())
	if err != nil {
		s.logger.Error("failed to add activity ref", "error", err, "batch_id", batchID)
		return err
	}

	return nil
}

// MarkActivityComplete marks an activity as successfully processed
func (s *PostgresBatchStore) MarkActivityComplete(ctx context.Context, batchID, activityID string, record interface{}) error {
	query := `
		UPDATE batch_activity_refs
		SET status = 'complete'
		WHERE batch_id = $1 AND activity_id = $2
	`

	_, err := s.db.ExecContext(ctx, query, batchID, activityID)
	if err != nil {
		s.logger.Error("failed to mark activity complete", "error", err)
		return err
	}

	// Update batch success count
	updateQuery := `
		UPDATE batch_jobs
		SET success_count = success_count + 1
		WHERE id = $1
	`
	_, err = s.db.ExecContext(ctx, updateQuery, batchID)
	return err
}

// MarkActivityFailed marks an activity as failed
func (s *PostgresBatchStore) MarkActivityFailed(ctx context.Context, batchID, activityID string, errMsg string) error {
	query := `
		UPDATE batch_activity_refs
		SET status = 'failed', error_message = $1
		WHERE batch_id = $2 AND activity_id = $3
	`

	_, err := s.db.ExecContext(ctx, query, errMsg, batchID, activityID)
	if err != nil {
		s.logger.Error("failed to mark activity failed", "error", err)
		return err
	}

	// Update batch error count
	updateQuery := `
		UPDATE batch_jobs
		SET error_count = error_count + 1
		WHERE id = $1
	`
	_, err = s.db.ExecContext(ctx, updateQuery, batchID)
	return err
}

// GetPendingBatches retrieves pending batches for processing
func (s *PostgresBatchStore) GetPendingBatches(ctx context.Context, limit int) ([]BatchJob, error) {
	query := `
		SELECT id, org_id, workspace_id, status, activity_count, success_count,
			   error_count, total_emissions, started_at, completed_at, error_message,
			   retry_count, max_retries, priority, created_at, updated_at,
			   locked_by, locked_until
		FROM batch_jobs
		WHERE status IN ('pending', 'queued')
		AND (locked_until IS NULL OR locked_until < NOW())
		ORDER BY priority DESC, created_at ASC
		LIMIT $1
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		s.logger.Error("failed to get pending batches", "error", err)
		return nil, err
	}
	defer rows.Close()

	var batches []BatchJob
	for rows.Next() {
		job := BatchJob{}
		err := rows.Scan(
			&job.ID, &job.OrgID, &job.WorkspaceID, &job.Status,
			&job.ActivityCount, &job.SuccessCount, &job.ErrorCount,
			&job.TotalEmissions, &job.StartedAt, &job.CompletedAt,
			&job.ErrorMessage, &job.RetryCount, &job.MaxRetries,
			&job.Priority, &job.CreatedAt, &job.UpdatedAt,
			&job.LockedBy, &job.LockedUntil,
		)
		if err != nil {
			s.logger.Error("failed to scan pending batch", "error", err)
			continue
		}
		batches = append(batches, job)
	}

	return batches, rows.Err()
}

// AcquireBatchLock acquires a distributed lock for a batch
func (s *PostgresBatchStore) AcquireBatchLock(ctx context.Context, batchID, workerID string, timeout time.Duration) (bool, error) {
	query := `
		UPDATE batch_jobs
		SET locked_by = $1, locked_until = $2
		WHERE id = $3
		AND (locked_by IS NULL OR locked_until < NOW())
	`

	result, err := s.db.ExecContext(ctx, query, workerID, time.Now().Add(timeout), batchID)
	if err != nil {
		s.logger.Error("failed to acquire lock", "error", err, "batch_id", batchID)
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}

// ReleaseBatchLock releases a distributed lock
func (s *PostgresBatchStore) ReleaseBatchLock(ctx context.Context, batchID string) error {
	query := `
		UPDATE batch_jobs
		SET locked_by = NULL, locked_until = NULL
		WHERE id = $1
	`

	_, err := s.db.ExecContext(ctx, query, batchID)
	if err != nil {
		s.logger.Error("failed to release lock", "error", err, "batch_id", batchID)
		return err
	}

	return nil
}

// DeleteBatch deletes a batch and all related records
func (s *PostgresBatchStore) DeleteBatch(ctx context.Context, batchID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete activity refs
	_, err = tx.ExecContext(ctx, "DELETE FROM batch_activity_refs WHERE batch_id = $1", batchID)
	if err != nil {
		s.logger.Error("failed to delete activity refs", "error", err)
		return err
	}

	// Delete progress logs
	_, err = tx.ExecContext(ctx, "DELETE FROM batch_progress_log WHERE batch_id = $1", batchID)
	if err != nil {
		s.logger.Error("failed to delete progress logs", "error", err)
		return err
	}

	// Delete batch
	_, err = tx.ExecContext(ctx, "DELETE FROM batch_jobs WHERE id = $1", batchID)
	if err != nil {
		s.logger.Error("failed to delete batch", "error", err)
		return err
	}

	return tx.Commit()
}

// GetProgressLog retrieves the progress log for a batch
func (s *PostgresBatchStore) GetProgressLog(ctx context.Context, batchID string) ([]BatchProgressLog, error) {
	query := `
		SELECT id, batch_id, event_type, processed_count, error_count, total_emissions, timestamp
		FROM batch_progress_log
		WHERE batch_id = $1
		ORDER BY timestamp ASC
	`

	rows, err := s.db.QueryContext(ctx, query, batchID)
	if err != nil {
		s.logger.Error("failed to get progress log", "error", err)
		return nil, err
	}
	defer rows.Close()

	var logs []BatchProgressLog
	for rows.Next() {
		log := BatchProgressLog{}
		err := rows.Scan(
			&log.ID, &log.BatchID, &log.EventType,
			&log.ProcessedCount, &log.ErrorCount,
			&log.TotalEmissions, &log.Timestamp,
		)
		if err != nil {
			s.logger.Error("failed to scan progress log", "error", err)
			continue
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}
