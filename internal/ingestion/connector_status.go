package ingestion

import (
	"context"
	"database/sql"
	"time"
)

// ConnectorStatusStore records connector status/last run.
type ConnectorStatusStore interface {
	SetStatus(ctx context.Context, name, orgID, status, lastError string, runAt *time.Time) error
	LastError(ctx context.Context, name, orgID string, err error) error
}

// PostgresConnectorStatusStore implements ConnectorStatusStore with PostgreSQL.
type PostgresConnectorStatusStore struct {
	db *sql.DB
}

// NewPostgresConnectorStatusStore creates a new PostgreSQL-backed connector status store.
func NewPostgresConnectorStatusStore(db *sql.DB) *PostgresConnectorStatusStore {
	return &PostgresConnectorStatusStore{db: db}
}

// SetStatus records the status of a connector.
func (s *PostgresConnectorStatusStore) SetStatus(ctx context.Context, name, orgID, status, lastError string, runAt *time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO connector_status (name, org_id, status, last_error, last_run_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (name, org_id) DO UPDATE SET
			status = EXCLUDED.status,
			last_error = EXCLUDED.last_error,
			last_run_at = EXCLUDED.last_run_at,
			updated_at = NOW()
	`, name, orgID, status, lastError, runAt)
	return err
}

// LastError records an error for a connector.
func (s *PostgresConnectorStatusStore) LastError(ctx context.Context, name, orgID string, err error) error {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	now := time.Now()
	return s.SetStatus(ctx, name, orgID, "error", errMsg, &now)
}
