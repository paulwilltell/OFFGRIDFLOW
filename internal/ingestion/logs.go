package ingestion

import (
	"context"
	"time"
)

// IngestionLog captures the outcome of an ingestion run.
type IngestionLog struct {
	ID          string        `json:"id"`
	Source      string        `json:"source"`
	Status      string        `json:"status"`
	Processed   int           `json:"processed"`
	Succeeded   int           `json:"succeeded"`
	Failed      int           `json:"failed"`
	Errors      []ImportError `json:"errors,omitempty"`
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt time.Time     `json:"completed_at,omitempty"`
	OrgID       string        `json:"org_id,omitempty"`
}

// LogStore persists ingestion logs for audit and status UI.
type LogStore interface {
	Record(ctx context.Context, log IngestionLog) error
	List(ctx context.Context, limit int) ([]IngestionLog, error)
}

// InMemoryLogStore stores logs in memory (dev/test).
type InMemoryLogStore struct {
	Logs []IngestionLog
}

// NewInMemoryLogStore creates an in-memory log store.
func NewInMemoryLogStore() *InMemoryLogStore {
	return &InMemoryLogStore{Logs: make([]IngestionLog, 0)}
}

// Record appends a log entry.
func (s *InMemoryLogStore) Record(ctx context.Context, log IngestionLog) error {
	_ = ctx
	s.Logs = append(s.Logs, log)
	return nil
}

// List returns recent logs (most recent first).
func (s *InMemoryLogStore) List(ctx context.Context, limit int) ([]IngestionLog, error) {
	_ = ctx
	if limit <= 0 || limit > len(s.Logs) {
		limit = len(s.Logs)
	}
	out := make([]IngestionLog, 0, limit)
	for i := len(s.Logs) - 1; i >= 0 && len(out) < limit; i-- {
		out = append(out, s.Logs[i])
	}
	return out, nil
}
