package ingestion

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// PostgresLogStore persists ingestion logs to the ingestion_logs table.
type PostgresLogStore struct {
	db *sql.DB
}

// NewPostgresLogStore creates a new Postgres-backed log store.
func NewPostgresLogStore(db *sql.DB) *PostgresLogStore {
	return &PostgresLogStore{db: db}
}

// Record inserts a log entry.
func (s *PostgresLogStore) Record(ctx context.Context, log IngestionLog) error {
	if log.ID == "" {
		log.ID = time.Now().UTC().Format("20060102T150405.000000000")
	}
	errorsJSON, _ := json.Marshal(log.Errors)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO ingestion_logs (
			id, source, status, processed, succeeded, failed, errors, started_at, completed_at, org_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			processed = EXCLUDED.processed,
			succeeded = EXCLUDED.succeeded,
			failed = EXCLUDED.failed,
			errors = EXCLUDED.errors,
			started_at = EXCLUDED.started_at,
			completed_at = EXCLUDED.completed_at,
			org_id = EXCLUDED.org_id
	`, log.ID, log.Source, log.Status, log.Processed, log.Succeeded, log.Failed, errorsJSON, log.StartedAt, log.CompletedAt, log.OrgID)
	return err
}

// List returns recent ingestion logs.
func (s *PostgresLogStore) List(ctx context.Context, limit int) ([]IngestionLog, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, source, status, processed, succeeded, failed, errors, started_at, completed_at, org_id
		FROM ingestion_logs
		ORDER BY started_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []IngestionLog
	for rows.Next() {
		var (
			log      IngestionLog
			errorsJSON []byte
			completedAt sql.NullTime
			orgID sql.NullString
		)
		if err := rows.Scan(&log.ID, &log.Source, &log.Status, &log.Processed, &log.Succeeded, &log.Failed, &errorsJSON, &log.StartedAt, &completedAt, &orgID); err != nil {
			return nil, err
		}
		if completedAt.Valid {
			log.CompletedAt = completedAt.Time
		}
		if orgID.Valid {
			log.OrgID = orgID.String
		}
		if len(errorsJSON) > 0 {
			_ = json.Unmarshal(errorsJSON, &log.Errors)
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}
