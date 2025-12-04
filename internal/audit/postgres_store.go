package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// PostgresStore implements Store using PostgreSQL
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a new PostgreSQL audit store
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// SaveEvent saves an audit event to the database
func (s *PostgresStore) SaveEvent(ctx context.Context, event *Event) error {
	query := `
		INSERT INTO audit_logs (
			id, timestamp, event_type, actor_id, actor_type, tenant_id,
			resource, action, status, ip_address, user_agent, details, error
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	var detailsJSON []byte
	var err error
	if event.Details != nil {
		detailsJSON, err = json.Marshal(event.Details)
		if err != nil {
			return fmt.Errorf("failed to marshal details: %w", err)
		}
	}

	_, err = s.db.ExecContext(ctx, query,
		event.ID,
		event.Timestamp,
		event.EventType,
		event.ActorID,
		event.ActorType,
		event.TenantID,
		event.Resource,
		event.Action,
		event.Status,
		event.IPAddress,
		event.UserAgent,
		detailsJSON,
		event.Error,
	)

	if err != nil {
		return fmt.Errorf("failed to save audit event: %w", err)
	}

	return nil
}

// QueryEvents retrieves events matching the query
func (s *PostgresStore) QueryEvents(ctx context.Context, query EventQuery) ([]*Event, error) {
	sqlQuery := `
		SELECT id, timestamp, event_type, actor_id, actor_type, tenant_id,
		       resource, action, status, ip_address, user_agent, details, error
		FROM audit_logs
		WHERE 1=1
	`
	args := make([]interface{}, 0)
	argIndex := 1

	if query.TenantID != "" {
		sqlQuery += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, query.TenantID)
		argIndex++
	}

	if query.ActorID != "" {
		sqlQuery += fmt.Sprintf(" AND actor_id = $%d", argIndex)
		args = append(args, query.ActorID)
		argIndex++
	}

	if len(query.EventTypes) > 0 {
		sqlQuery += fmt.Sprintf(" AND event_type = ANY($%d)", argIndex)
		eventTypes := make([]string, len(query.EventTypes))
		for i, et := range query.EventTypes {
			eventTypes[i] = string(et)
		}
		args = append(args, eventTypes)
		argIndex++
	}

	if !query.StartTime.IsZero() {
		sqlQuery += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, query.StartTime)
		argIndex++
	}

	if !query.EndTime.IsZero() {
		sqlQuery += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, query.EndTime)
		argIndex++
	}

	sqlQuery += " ORDER BY timestamp DESC"

	if query.Limit > 0 {
		sqlQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, query.Limit)
		argIndex++
	}

	if query.Offset > 0 {
		sqlQuery += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, query.Offset)
		argIndex++
	}

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events: %w", err)
	}
	defer rows.Close()

	events := make([]*Event, 0)
	for rows.Next() {
		event := &Event{}
		var detailsJSON []byte
		var resource, action, ipAddress, userAgent, eventError sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.Timestamp,
			&event.EventType,
			&event.ActorID,
			&event.ActorType,
			&event.TenantID,
			&resource,
			&action,
			&event.Status,
			&ipAddress,
			&userAgent,
			&detailsJSON,
			&eventError,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit event: %w", err)
		}

		if resource.Valid {
			event.Resource = resource.String
		}
		if action.Valid {
			event.Action = action.String
		}
		if ipAddress.Valid {
			event.IPAddress = ipAddress.String
		}
		if userAgent.Valid {
			event.UserAgent = userAgent.String
		}
		if eventError.Valid {
			event.Error = eventError.String
		}

		if detailsJSON != nil {
			if err := json.Unmarshal(detailsJSON, &event.Details); err != nil {
				return nil, fmt.Errorf("failed to unmarshal details: %w", err)
			}
		}

		events = append(events, event)
	}

	return events, nil
}

// GetEvent retrieves a single event by ID
func (s *PostgresStore) GetEvent(ctx context.Context, id string) (*Event, error) {
	query := `
		SELECT id, timestamp, event_type, actor_id, actor_type, tenant_id,
		       resource, action, status, ip_address, user_agent, details, error
		FROM audit_logs
		WHERE id = $1
	`

	event := &Event{}
	var detailsJSON []byte
	var resource, action, ipAddress, userAgent, eventError sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.Timestamp,
		&event.EventType,
		&event.ActorID,
		&event.ActorType,
		&event.TenantID,
		&resource,
		&action,
		&event.Status,
		&ipAddress,
		&userAgent,
		&detailsJSON,
		&eventError,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("audit event not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get audit event: %w", err)
	}

	if resource.Valid {
		event.Resource = resource.String
	}
	if action.Valid {
		event.Action = action.String
	}
	if ipAddress.Valid {
		event.IPAddress = ipAddress.String
	}
	if userAgent.Valid {
		event.UserAgent = userAgent.String
	}
	if eventError.Valid {
		event.Error = eventError.String
	}

	if detailsJSON != nil {
		if err := json.Unmarshal(detailsJSON, &event.Details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal details: %w", err)
		}
	}

	return event, nil
}

// CreateAuditLogsTable creates the audit_logs table if it doesn't exist
func CreateAuditLogsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS audit_logs (
			id VARCHAR(255) PRIMARY KEY,
			timestamp TIMESTAMP NOT NULL,
			event_type VARCHAR(100) NOT NULL,
			actor_id VARCHAR(255) NOT NULL,
			actor_type VARCHAR(50) NOT NULL,
			tenant_id VARCHAR(255),
			resource VARCHAR(255),
			action VARCHAR(100),
			status VARCHAR(50) NOT NULL,
			ip_address VARCHAR(45),
			user_agent TEXT,
			details JSONB,
			error TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_audit_tenant_timestamp 
			ON audit_logs(tenant_id, timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_audit_actor 
			ON audit_logs(actor_id, timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_audit_event_type 
			ON audit_logs(event_type, timestamp DESC);
		CREATE INDEX IF NOT EXISTS idx_audit_timestamp 
			ON audit_logs(timestamp DESC);
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create audit_logs table: %w", err)
	}

	return nil
}
