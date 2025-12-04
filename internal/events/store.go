package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// EventStore provides persistent storage for events.
// It enables event sourcing, audit trails, and event replay.
type EventStore interface {
	// Append adds an event to the store
	Append(ctx context.Context, event Event) error

	// AppendBatch adds multiple events atomically
	AppendBatch(ctx context.Context, events []Event) error

	// Load retrieves events matching the criteria
	Load(ctx context.Context, criteria EventCriteria) ([]Event, error)

	// LoadStream retrieves events for a specific aggregate/stream
	LoadStream(ctx context.Context, streamID string, fromVersion int) ([]Event, error)

	// Subscribe to new events (real-time)
	Subscribe(ctx context.Context, criteria EventCriteria, handler Handler) error
}

// EventCriteria defines filters for loading events.
type EventCriteria struct {
	// EventTypes filters by event type (empty = all)
	EventTypes []string

	// StreamID filters by aggregate/stream ID
	StreamID string

	// Since filters events after this time
	Since time.Time

	// Until filters events before this time
	Until time.Time

	// CorrelationID filters by correlation
	CorrelationID string

	// Limit limits the number of results
	Limit int

	// Offset for pagination
	Offset int

	// OrderBy specifies sort order ("asc" or "desc")
	OrderBy string
}

// PostgresEventStore implements EventStore using PostgreSQL.
type PostgresEventStore struct {
	db  *sql.DB
	bus Bus // Optional: publish events in real-time
}

// NewPostgresEventStore creates a new PostgreSQL-backed event store.
func NewPostgresEventStore(db *sql.DB, bus Bus) (*PostgresEventStore, error) {
	store := &PostgresEventStore{
		db:  db,
		bus: bus,
	}

	if err := store.ensureSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("ensure schema: %w", err)
	}

	return store, nil
}

// ensureSchema creates the necessary database schema.
func (s *PostgresEventStore) ensureSchema(ctx context.Context) error {
	schema := `
		CREATE TABLE IF NOT EXISTS event_store (
			id BIGSERIAL PRIMARY KEY,
			event_id TEXT NOT NULL UNIQUE,
			event_type TEXT NOT NULL,
			stream_id TEXT,
			version INTEGER,
			payload JSONB NOT NULL,
			metadata JSONB,
			timestamp TIMESTAMPTZ NOT NULL,
			source TEXT,
			correlation_id TEXT,
			causation_id TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(stream_id, version)
		);

		CREATE INDEX IF NOT EXISTS idx_event_store_type ON event_store(event_type);
		CREATE INDEX IF NOT EXISTS idx_event_store_stream ON event_store(stream_id);
		CREATE INDEX IF NOT EXISTS idx_event_store_timestamp ON event_store(timestamp);
		CREATE INDEX IF NOT EXISTS idx_event_store_correlation ON event_store(correlation_id);
		CREATE INDEX IF NOT EXISTS idx_event_store_created ON event_store(created_at);
	`

	_, err := s.db.ExecContext(ctx, schema)
	return err
}

// Append adds an event to the store.
func (s *PostgresEventStore) Append(ctx context.Context, event Event) error {
	if err := event.Validate(); err != nil {
		return err
	}

	// Ensure event has ID and timestamp
	if event.ID == "" {
		return fmt.Errorf("event ID is required")
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		INSERT INTO event_store (
			event_id, event_type, stream_id, version, payload, metadata,
			timestamp, source, correlation_id, causation_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = s.db.ExecContext(ctx, query,
		event.ID,
		event.Type,
		sql.NullString{String: event.CorrelationID, Valid: event.CorrelationID != ""},
		sql.NullInt64{Int64: int64(event.Version), Valid: event.Version > 0},
		payload,
		metadata,
		event.Timestamp,
		sql.NullString{String: event.Source, Valid: event.Source != ""},
		sql.NullString{String: event.CorrelationID, Valid: event.CorrelationID != ""},
		sql.NullString{String: event.CausationID, Valid: event.CausationID != ""},
	)

	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}

	// Publish to bus if available
	if s.bus != nil {
		s.bus.Publish(ctx, event)
	}

	return nil
}

// AppendBatch adds multiple events atomically.
func (s *PostgresEventStore) AppendBatch(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO event_store (
			event_id, event_type, stream_id, version, payload, metadata,
			timestamp, source, correlation_id, causation_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, event := range events {
		if err := event.Validate(); err != nil {
			return err
		}

		if event.ID == "" {
			return fmt.Errorf("event ID is required")
		}
		if event.Timestamp.IsZero() {
			event.Timestamp = time.Now().UTC()
		}

		payload, err := json.Marshal(event.Payload)
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}

		metadata, err := json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			event.ID,
			event.Type,
			sql.NullString{String: event.CorrelationID, Valid: event.CorrelationID != ""},
			sql.NullInt64{Int64: int64(event.Version), Valid: event.Version > 0},
			payload,
			metadata,
			event.Timestamp,
			sql.NullString{String: event.Source, Valid: event.Source != ""},
			sql.NullString{String: event.CorrelationID, Valid: event.CorrelationID != ""},
			sql.NullString{String: event.CausationID, Valid: event.CausationID != ""},
		)
		if err != nil {
			return fmt.Errorf("insert event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// Publish to bus if available
	if s.bus != nil {
		for _, event := range events {
			s.bus.Publish(ctx, event)
		}
	}

	return nil
}

// Load retrieves events matching the criteria.
func (s *PostgresEventStore) Load(ctx context.Context, criteria EventCriteria) ([]Event, error) {
	query := `
		SELECT event_id, event_type, stream_id, version, payload, metadata,
			   timestamp, source, correlation_id, causation_id
		FROM event_store
		WHERE 1=1
	`
	args := make([]interface{}, 0)
	argNum := 1

	if len(criteria.EventTypes) > 0 {
		query += fmt.Sprintf(" AND event_type = ANY($%d)", argNum)
		args = append(args, pq.Array(criteria.EventTypes))
		argNum++
	}

	if criteria.StreamID != "" {
		query += fmt.Sprintf(" AND stream_id = $%d", argNum)
		args = append(args, criteria.StreamID)
		argNum++
	}

	if !criteria.Since.IsZero() {
		query += fmt.Sprintf(" AND timestamp >= $%d", argNum)
		args = append(args, criteria.Since)
		argNum++
	}

	if !criteria.Until.IsZero() {
		query += fmt.Sprintf(" AND timestamp <= $%d", argNum)
		args = append(args, criteria.Until)
		argNum++
	}

	if criteria.CorrelationID != "" {
		query += fmt.Sprintf(" AND correlation_id = $%d", argNum)
		args = append(args, criteria.CorrelationID)
		argNum++
	}

	// Order by
	if criteria.OrderBy == "asc" {
		query += " ORDER BY timestamp ASC, id ASC"
	} else {
		query += " ORDER BY timestamp DESC, id DESC"
	}

	// Limit and offset
	if criteria.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, criteria.Limit)
		argNum++
	}

	if criteria.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, criteria.Offset)
		argNum++
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		var streamID, source, correlationID, causationID sql.NullString
		var version sql.NullInt64
		var payloadJSON, metadataJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.Type,
			&streamID,
			&version,
			&payloadJSON,
			&metadataJSON,
			&event.Timestamp,
			&source,
			&correlationID,
			&causationID,
		)
		if err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}

		if version.Valid {
			event.Version = int(version.Int64)
		}
		if source.Valid {
			event.Source = source.String
		}
		if correlationID.Valid {
			event.CorrelationID = correlationID.String
		}
		if causationID.Valid {
			event.CausationID = causationID.String
		}

		if err := json.Unmarshal(payloadJSON, &event.Payload); err != nil {
			return nil, fmt.Errorf("unmarshal payload: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshal metadata: %w", err)
			}
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// LoadStream retrieves events for a specific aggregate/stream.
func (s *PostgresEventStore) LoadStream(ctx context.Context, streamID string, fromVersion int) ([]Event, error) {
	return s.Load(ctx, EventCriteria{
		StreamID: streamID,
		OrderBy:  "asc",
	})
}

// Subscribe to new events in real-time using PostgreSQL LISTEN/NOTIFY.
func (s *PostgresEventStore) Subscribe(ctx context.Context, criteria EventCriteria, handler Handler) error {
	// This is a simplified implementation
	// In production, you'd use PostgreSQL LISTEN/NOTIFY or polling
	if s.bus != nil {
		// Delegate to the bus if available
		topic := "*"
		if len(criteria.EventTypes) == 1 {
			topic = criteria.EventTypes[0]
		}
		return s.bus.Subscribe(ctx, topic, handler)
	}

	return fmt.Errorf("real-time subscription not available without event bus")
}

// Count returns the total number of events matching the criteria.
func (s *PostgresEventStore) Count(ctx context.Context, criteria EventCriteria) (int64, error) {
	query := "SELECT COUNT(*) FROM event_store WHERE 1=1"
	args := make([]interface{}, 0)
	argNum := 1

	if len(criteria.EventTypes) > 0 {
		query += fmt.Sprintf(" AND event_type = ANY($%d)", argNum)
		args = append(args, pq.Array(criteria.EventTypes))
		argNum++
	}

	if criteria.StreamID != "" {
		query += fmt.Sprintf(" AND stream_id = $%d", argNum)
		args = append(args, criteria.StreamID)
		argNum++
	}

	if !criteria.Since.IsZero() {
		query += fmt.Sprintf(" AND timestamp >= $%d", argNum)
		args = append(args, criteria.Since)
		argNum++
	}

	if !criteria.Until.IsZero() {
		query += fmt.Sprintf(" AND timestamp <= $%d", argNum)
		args = append(args, criteria.Until)
		argNum++
	}

	var count int64
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// Compile-time interface check
var _ EventStore = (*PostgresEventStore)(nil)
