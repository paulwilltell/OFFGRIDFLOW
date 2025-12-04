package connectors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Connector represents connector configuration and status.
type Connector struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Config    json.RawMessage `json:"config,omitempty"`
	Status    string          `json:"status"`
	LastRunAt *time.Time      `json:"last_run_at,omitempty"`
	LastError string          `json:"last_error,omitempty"`
	OrgID     string          `json:"org_id,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Store defines persistence for connectors.
type Store interface {
	List(ctx context.Context, orgID string) ([]Connector, error)
	SetStatus(ctx context.Context, name, orgID, status, lastError string, runAt *time.Time) error
	LastError(ctx context.Context, name, orgID string, err error) error
}

// PostgresStore implements Store.
type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) List(ctx context.Context, orgID string) ([]Connector, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, config, status, last_run_at, last_error, org_id, created_at, updated_at
		FROM connectors
		WHERE org_id = $1 OR $1 = ''
		ORDER BY name
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Connector
	for rows.Next() {
		var c Connector
		var cfg json.RawMessage
		var lastRun sql.NullTime
		if err := rows.Scan(&c.ID, &c.Name, &cfg, &c.Status, &lastRun, &c.LastError, &c.OrgID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		if len(cfg) > 0 {
			c.Config = cfg
		}
		if lastRun.Valid {
			c.LastRunAt = &lastRun.Time
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *PostgresStore) SetStatus(ctx context.Context, name, orgID, status, lastError string, runAt *time.Time) error {
	now := time.Now()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO connectors (name, org_id, status, last_error, last_run_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (name, org_id) DO UPDATE SET
			status = EXCLUDED.status,
			last_error = EXCLUDED.last_error,
			last_run_at = EXCLUDED.last_run_at,
			updated_at = EXCLUDED.updated_at
	`, name, orgID, status, nullableString(lastError), runAt, now, now)
	if err != nil {
		return fmt.Errorf("set connector status: %w", err)
	}
	return nil
}

// LastError records an error for a connector without changing status.
func (s *PostgresStore) LastError(ctx context.Context, name, orgID string, err error) error {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	now := time.Now()
	_, execErr := s.db.ExecContext(ctx, `
		UPDATE connectors SET last_error = $3, updated_at = $4 WHERE name = $1 AND org_id = $2
	`, name, orgID, msg, now)
	return execErr
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
