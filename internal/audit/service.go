// Package audit provides the audit logging service.
package audit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Service Configuration
// =============================================================================

// ServiceConfig configures the audit service.
type ServiceConfig struct {
	// Store is the audit entry storage backend.
	Store Store

	// Logger for audit operations.
	Logger *slog.Logger

	// AsyncWrite enables asynchronous write with buffering.
	AsyncWrite bool

	// BufferSize is the async write buffer size.
	BufferSize int

	// FlushInterval is how often to flush the buffer.
	FlushInterval time.Duration

	// RetentionDays is how long to keep audit entries (0 = forever).
	RetentionDays int

	// ExcludeReadActions skips logging read-only actions.
	ExcludeReadActions bool

	// MaskSensitiveFields redacts sensitive data in changes.
	MaskSensitiveFields []string
}

// DefaultServiceConfig returns sensible defaults.
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		AsyncWrite:    true,
		BufferSize:    1000,
		FlushInterval: 5 * time.Second,
		RetentionDays: 365, // 1 year
		MaskSensitiveFields: []string{
			"password", "secret", "token", "api_key",
			"credit_card", "ssn", "bank_account",
		},
	}
}

// =============================================================================
// Store Interface
// =============================================================================

// Store defines the storage interface for audit entries.
type Store interface {
	// Write persists an audit entry.
	Write(ctx context.Context, entry AuditEntry) error

	// WriteBatch persists multiple entries.
	WriteBatch(ctx context.Context, entries []AuditEntry) error

	// Read retrieves an entry by ID.
	Read(ctx context.Context, id string) (AuditEntry, error)

	// Search finds entries matching the query.
	Search(ctx context.Context, query Query) (QueryResult, error)

	// Summarize generates aggregated statistics.
	Summarize(ctx context.Context, orgID string, from, to time.Time) (Summary, error)

	// Delete removes entries (for retention policy).
	Delete(ctx context.Context, before time.Time) (int64, error)
}

// =============================================================================
// Service Implementation
// =============================================================================

// Service provides audit logging capabilities.
//
// The service supports both synchronous and asynchronous writes. Async mode
// buffers entries and writes them in batches for better performance.
//
// Example usage:
//
//	svc := audit.NewService(config)
//	defer svc.Close()
//
//	entry := audit.NewEntryBuilder().
//	    WithEntity(audit.EntityEmission, "em_123").
//	    WithAction(audit.ActionCreate).
//	    WithActor("user_456", "user").
//	    MustBuild()
//
//	if err := svc.Record(ctx, entry); err != nil {
//	    log.Error("audit failed", "error", err)
//	}
type Service struct {
	store  Store
	logger *slog.Logger
	config ServiceConfig
	buffer chan AuditEntry
	done   chan struct{}
	wg     sync.WaitGroup
	mu     sync.RWMutex
	closed bool
}

// NewService creates a new audit service.
func NewService(cfg ServiceConfig) *Service {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	s := &Service{
		store:  cfg.Store,
		logger: logger,
		config: cfg,
		done:   make(chan struct{}),
	}

	if cfg.AsyncWrite {
		bufSize := cfg.BufferSize
		if bufSize <= 0 {
			bufSize = 1000
		}
		s.buffer = make(chan AuditEntry, bufSize)

		s.wg.Add(1)
		go s.flushLoop()
	}

	logger.Info("audit service started",
		"async_write", cfg.AsyncWrite,
		"buffer_size", cfg.BufferSize,
		"retention_days", cfg.RetentionDays,
	)

	return s
}

// Record logs an audit entry.
func (s *Service) Record(ctx context.Context, entry AuditEntry) error {
	if s == nil {
		return errors.New("audit: nil service")
	}

	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return errors.New("audit: service is closed")
	}
	s.mu.RUnlock()

	// Skip read actions if configured
	if s.config.ExcludeReadActions && entry.Action == ActionRead {
		return nil
	}

	// Generate ID if not provided
	if entry.ID == "" {
		entry.ID = GenerateAuditID()
	}

	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	// Validate entry
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("validate audit entry: %w", err)
	}

	// Mask sensitive fields
	s.maskSensitiveData(&entry)

	// Write async or sync
	if s.config.AsyncWrite && s.buffer != nil {
		select {
		case s.buffer <- entry:
			return nil
		default:
			// Buffer full, fall back to sync write
			s.logger.Warn("audit buffer full, writing synchronously")
		}
	}

	// Synchronous write
	if s.store != nil {
		if err := s.store.Write(ctx, entry); err != nil {
			s.logger.Error("failed to write audit entry",
				"entry_id", entry.ID,
				"error", err,
			)
			return fmt.Errorf("write audit entry: %w", err)
		}
	}

	s.logger.Debug("recorded audit entry",
		"id", entry.ID,
		"entity", entry.Entity,
		"entity_id", entry.EntityID,
		"action", entry.Action,
		"actor", entry.Actor.ID,
	)

	return nil
}

// RecordCreate is a convenience method for create actions.
func (s *Service) RecordCreate(ctx context.Context, entity EntityType, entityID string, actor ActorInfo, data interface{}) error {
	entry := NewEntryBuilder().
		WithEntity(entity, entityID).
		WithAction(ActionCreate).
		WithActor(actor.ID, actor.Type).
		WithActorDetails(actor.Name, actor.Email).
		WithChanges(nil, data).
		MustBuild()

	return s.Record(ctx, entry)
}

// RecordUpdate is a convenience method for update actions.
func (s *Service) RecordUpdate(ctx context.Context, entity EntityType, entityID string, actor ActorInfo, before, after interface{}) error {
	entry := NewEntryBuilder().
		WithEntity(entity, entityID).
		WithAction(ActionUpdate).
		WithActor(actor.ID, actor.Type).
		WithActorDetails(actor.Name, actor.Email).
		WithChanges(before, after).
		MustBuild()

	return s.Record(ctx, entry)
}

// RecordDelete is a convenience method for delete actions.
func (s *Service) RecordDelete(ctx context.Context, entity EntityType, entityID string, actor ActorInfo, data interface{}) error {
	entry := NewEntryBuilder().
		WithEntity(entity, entityID).
		WithAction(ActionDelete).
		WithActor(actor.ID, actor.Type).
		WithActorDetails(actor.Name, actor.Email).
		WithChanges(data, nil).
		MustBuild()

	return s.Record(ctx, entry)
}

// Search finds audit entries matching the query.
func (s *Service) Search(ctx context.Context, query Query) (QueryResult, error) {
	if s.store == nil {
		return QueryResult{}, errors.New("audit: no store configured")
	}

	// Apply default limit
	if query.Limit <= 0 {
		query.Limit = 100
	}
	if query.Limit > 1000 {
		query.Limit = 1000
	}

	return s.store.Search(ctx, query)
}

// GetEntry retrieves a single audit entry by ID.
func (s *Service) GetEntry(ctx context.Context, id string) (AuditEntry, error) {
	if s.store == nil {
		return AuditEntry{}, errors.New("audit: no store configured")
	}

	return s.store.Read(ctx, id)
}

// GetEntityHistory retrieves all audit entries for a specific entity.
func (s *Service) GetEntityHistory(ctx context.Context, entity EntityType, entityID string) ([]AuditEntry, error) {
	result, err := s.Search(ctx, Query{
		Entity:   entity,
		EntityID: entityID,
		OrderBy:  "-timestamp",
		Limit:    100,
	})
	if err != nil {
		return nil, err
	}

	return result.Entries, nil
}

// Summarize generates audit statistics.
func (s *Service) Summarize(ctx context.Context, orgID string, from, to time.Time) (Summary, error) {
	if s.store == nil {
		return Summary{}, errors.New("audit: no store configured")
	}

	return s.store.Summarize(ctx, orgID, from, to)
}

// Cleanup removes old audit entries based on retention policy.
func (s *Service) Cleanup(ctx context.Context) (int64, error) {
	if s.store == nil {
		return 0, errors.New("audit: no store configured")
	}

	if s.config.RetentionDays <= 0 {
		return 0, nil // No retention limit
	}

	before := time.Now().UTC().AddDate(0, 0, -s.config.RetentionDays)
	count, err := s.store.Delete(ctx, before)
	if err != nil {
		return 0, fmt.Errorf("delete old entries: %w", err)
	}

	if count > 0 {
		s.logger.Info("cleaned up old audit entries",
			"deleted_count", count,
			"before", before,
		)
	}

	return count, nil
}

// Close shuts down the service gracefully.
func (s *Service) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	// Signal flush loop to stop
	close(s.done)

	// Wait for flush to complete
	s.wg.Wait()

	s.logger.Info("audit service stopped")

	return nil
}

// Flush forces an immediate flush of buffered entries.
func (s *Service) Flush(ctx context.Context) error {
	if s.buffer == nil {
		return nil
	}

	entries := make([]AuditEntry, 0, len(s.buffer))

	// Drain buffer
	for {
		select {
		case entry := <-s.buffer:
			entries = append(entries, entry)
		default:
			// Buffer empty
			if len(entries) == 0 {
				return nil
			}

			if s.store != nil {
				return s.store.WriteBatch(ctx, entries)
			}
			return nil
		}
	}
}

// flushLoop runs in the background to batch write entries.
func (s *Service) flushLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.FlushInterval)
	defer ticker.Stop()

	entries := make([]AuditEntry, 0, 100)

	flush := func() {
		if len(entries) == 0 {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if s.store != nil {
			if err := s.store.WriteBatch(ctx, entries); err != nil {
				s.logger.Error("failed to flush audit entries",
					"count", len(entries),
					"error", err,
				)
			} else {
				s.logger.Debug("flushed audit entries",
					"count", len(entries),
				)
			}
		}

		entries = entries[:0]
	}

	for {
		select {
		case entry := <-s.buffer:
			entries = append(entries, entry)

			// Flush if buffer reaches threshold
			if len(entries) >= 100 {
				flush()
			}

		case <-ticker.C:
			flush()

		case <-s.done:
			// Final flush
			flush()
			return
		}
	}
}

// maskSensitiveData redacts sensitive fields in the entry.
func (s *Service) maskSensitiveData(entry *AuditEntry) {
	if entry.Changes == nil || len(s.config.MaskSensitiveFields) == 0 {
		return
	}

	for i := range entry.Changes.Fields {
		field := &entry.Changes.Fields[i]
		for _, sensitive := range s.config.MaskSensitiveFields {
			if containsIgnoreCase(field.Field, sensitive) {
				field.OldValue = "[REDACTED]"
				field.NewValue = "[REDACTED]"
				break
			}
		}
	}
}

// =============================================================================
// In-Memory Store Implementation
// =============================================================================

// MemoryStore is an in-memory audit store for testing.
type MemoryStore struct {
	entries []AuditEntry
	mu      sync.RWMutex
}

// NewMemoryStore creates a new in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		entries: make([]AuditEntry, 0, 1000),
	}
}

// Write stores an entry.
func (m *MemoryStore) Write(ctx context.Context, entry AuditEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries = append(m.entries, entry)
	return nil
}

// WriteBatch stores multiple entries.
func (m *MemoryStore) WriteBatch(ctx context.Context, entries []AuditEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries = append(m.entries, entries...)
	return nil
}

// Read retrieves an entry by ID.
func (m *MemoryStore) Read(ctx context.Context, id string) (AuditEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, e := range m.entries {
		if e.ID == id {
			return e, nil
		}
	}

	return AuditEntry{}, ErrEntryNotFound
}

// Search finds matching entries.
func (m *MemoryStore) Search(ctx context.Context, query Query) (QueryResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matches []AuditEntry

	for _, e := range m.entries {
		if m.matchesQuery(e, query) {
			matches = append(matches, e)
		}
	}

	// Sort by timestamp
	if query.OrderBy == "-timestamp" {
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Timestamp.After(matches[j].Timestamp)
		})
	} else {
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Timestamp.Before(matches[j].Timestamp)
		})
	}

	total := len(matches)

	// Apply pagination
	if query.Offset > 0 && query.Offset < len(matches) {
		matches = matches[query.Offset:]
	} else if query.Offset >= len(matches) {
		matches = nil
	}

	if query.Limit > 0 && query.Limit < len(matches) {
		matches = matches[:query.Limit]
	}

	return QueryResult{
		Entries: matches,
		Total:   total,
		Limit:   query.Limit,
		Offset:  query.Offset,
	}, nil
}

// Summarize generates statistics.
func (m *MemoryStore) Summarize(ctx context.Context, orgID string, from, to time.Time) (Summary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := Summary{
		ByAction:    make(map[Action]int64),
		ByEntity:    make(map[EntityType]int64),
		ByOutcome:   make(map[Outcome]int64),
		GeneratedAt: time.Now().UTC(),
	}

	actorCounts := make(map[string]int64)

	for _, e := range m.entries {
		// Filter by org if specified
		if orgID != "" && e.OrgID != orgID {
			continue
		}

		// Filter by time range
		if !from.IsZero() && e.Timestamp.Before(from) {
			continue
		}
		if !to.IsZero() && e.Timestamp.After(to) {
			continue
		}

		summary.TotalEntries++
		summary.ByAction[e.Action]++
		summary.ByEntity[e.Entity]++
		summary.ByOutcome[e.Outcome]++
		actorCounts[e.Actor.ID]++

		// Track time range
		if summary.TimeRange[0].IsZero() || e.Timestamp.Before(summary.TimeRange[0]) {
			summary.TimeRange[0] = e.Timestamp
		}
		if summary.TimeRange[1].IsZero() || e.Timestamp.After(summary.TimeRange[1]) {
			summary.TimeRange[1] = e.Timestamp
		}
	}

	// Get top actors
	type actorCount struct {
		id    string
		count int64
	}
	var actors []actorCount
	for id, count := range actorCounts {
		actors = append(actors, actorCount{id, count})
	}
	sort.Slice(actors, func(i, j int) bool {
		return actors[i].count > actors[j].count
	})

	for i := 0; i < len(actors) && i < 10; i++ {
		summary.TopActors = append(summary.TopActors, ActorSummary{
			ActorID:    actors[i].id,
			EntryCount: actors[i].count,
		})
	}

	return summary, nil
}

// Delete removes old entries.
func (m *MemoryStore) Delete(ctx context.Context, before time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var kept []AuditEntry
	var deleted int64

	for _, e := range m.entries {
		if e.Timestamp.Before(before) {
			deleted++
		} else {
			kept = append(kept, e)
		}
	}

	m.entries = kept
	return deleted, nil
}

// matchesQuery checks if an entry matches the query.
func (m *MemoryStore) matchesQuery(e AuditEntry, q Query) bool {
	if q.Entity != "" && e.Entity != q.Entity {
		return false
	}
	if q.EntityID != "" && e.EntityID != q.EntityID {
		return false
	}
	if q.Action != "" && e.Action != q.Action {
		return false
	}
	if q.ActorID != "" && e.Actor.ID != q.ActorID {
		return false
	}
	if q.OrgID != "" && e.OrgID != q.OrgID {
		return false
	}
	if q.Outcome != "" && e.Outcome != q.Outcome {
		return false
	}
	if q.CorrelationID != "" && e.CorrelationID != q.CorrelationID {
		return false
	}
	if !q.From.IsZero() && e.Timestamp.Before(q.From) {
		return false
	}
	if !q.To.IsZero() && e.Timestamp.After(q.To) {
		return false
	}

	return true
}

// Count returns the total number of entries.
func (m *MemoryStore) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entries)
}

// =============================================================================
// Helper Functions
// =============================================================================

// GenerateAuditID creates a unique audit entry identifier.
func GenerateAuditID() string {
	return fmt.Sprintf("audit_%s", uuid.New().String()[:12])
}

// containsIgnoreCase checks if s contains substr (case-insensitive).
func containsIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	for i := 0; i+len(substrLower) <= len(sLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

// toLower converts a string to lowercase.
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		} else {
			b[i] = c
		}
	}
	return string(b)
}
