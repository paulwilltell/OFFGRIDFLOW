package ingestion

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// SourceIngestionAdapter represents an ingestion connector.
type SourceIngestionAdapter interface {
	Ingest(ctx context.Context) ([]Activity, error)
}

// ActivityStore provides storage operations for activities.
type ActivityStore interface {
	Save(ctx context.Context, activity Activity) error
	SaveBatch(ctx context.Context, activities []Activity) error
	List(ctx context.Context) ([]Activity, error)
	ListBySource(ctx context.Context, source string) ([]Activity, error)
	ListByOrgAndSource(ctx context.Context, orgID, source string) ([]Activity, error)
	ListByOrg(ctx context.Context, orgID string) ([]Activity, error)
	ListRecent(ctx context.Context, since time.Time) ([]Activity, error)
}

// InMemoryActivityStore is a simple in-memory implementation of ActivityStore.
// Suitable for development and testing; not for production use.
type InMemoryActivityStore struct {
	mu         sync.RWMutex
	activities []Activity
}

// NewInMemoryActivityStore creates a new in-memory activity store.
func NewInMemoryActivityStore() *InMemoryActivityStore {
	return &InMemoryActivityStore{
		activities: make([]Activity, 0),
	}
}

// Save stores a single activity.
func (s *InMemoryActivityStore) Save(ctx context.Context, activity Activity) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if activity.CreatedAt.IsZero() {
		activity.CreatedAt = time.Now()
	}
	s.activities = append(s.activities, activity)
	return nil
}

// SaveBatch stores multiple activities.
func (s *InMemoryActivityStore) SaveBatch(ctx context.Context, activities []Activity) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for i := range activities {
		if activities[i].CreatedAt.IsZero() {
			activities[i].CreatedAt = now
		}
	}
	s.activities = append(s.activities, activities...)
	return nil
}

// List returns all stored activities.
func (s *InMemoryActivityStore) List(ctx context.Context) ([]Activity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Activity, len(s.activities))
	copy(result, s.activities)
	return result, nil
}

// ListBySource returns activities filtered by source type.
func (s *InMemoryActivityStore) ListBySource(ctx context.Context, source string) ([]Activity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Activity
	for _, act := range s.activities {
		if act.Source == source {
			result = append(result, act)
		}
	}
	return result, nil
}

// ListRecent returns activities created after the given time.
func (s *InMemoryActivityStore) ListRecent(ctx context.Context, since time.Time) ([]Activity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Activity
	for _, act := range s.activities {
		if act.CreatedAt.After(since) {
			result = append(result, act)
		}
	}
	return result, nil
}

// ListByOrgAndSource returns activities filtered by organization and source type.
func (s *InMemoryActivityStore) ListByOrgAndSource(ctx context.Context, orgID, source string) ([]Activity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Activity
	for _, act := range s.activities {
		if act.OrgID == orgID && act.Source == source {
			result = append(result, act)
		}
	}
	return result, nil
}

// ListByOrg returns activities filtered by organization.
func (s *InMemoryActivityStore) ListByOrg(ctx context.Context, orgID string) ([]Activity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Activity
	for _, act := range s.activities {
		if act.OrgID == orgID {
			result = append(result, act)
		}
	}
	return result, nil
}

// SeedDemoData adds sample utility bill activities for testing.
func (s *InMemoryActivityStore) SeedDemoData() {
	now := time.Now()
	demoActivities := []Activity{
		{
			ID:          "act-001",
			Source:      "utility_bill",
			MeterID:     "METER-001",
			Location:    "US-WEST",
			PeriodStart: now.AddDate(0, -1, 0),
			PeriodEnd:   now,
			Quantity:    12500.0,
			Unit:        "kWh",
			OrgID:       "org-demo",
			CreatedAt:   now,
		},
		{
			ID:          "act-002",
			Source:      "utility_bill",
			MeterID:     "METER-002",
			Location:    "EU-CENTRAL",
			PeriodStart: now.AddDate(0, -1, 0),
			PeriodEnd:   now,
			Quantity:    8750.0,
			Unit:        "kWh",
			OrgID:       "org-demo",
			CreatedAt:   now,
		},
		{
			ID:          "act-003",
			Source:      "utility_bill",
			MeterID:     "METER-003",
			Location:    "US-WEST",
			PeriodStart: now.AddDate(0, -2, 0),
			PeriodEnd:   now.AddDate(0, -1, 0),
			Quantity:    11200.0,
			Unit:        "kWh",
			OrgID:       "org-demo",
			CreatedAt:   now,
		},
	}
	s.mu.Lock()
	s.activities = append(s.activities, demoActivities...)
	s.mu.Unlock()
}

// Service orchestrates ingestion across multiple sources.
type Service struct {
	Adapters []SourceIngestionAdapter
	Store    ActivityStore

	// Logger is optional; defaults to slog.Default.
	Logger *slog.Logger

	// Logs optionally records ingestion runs for audit/status.
	Logs LogStore

	// ConnectorStore optionally records connector status.
	ConnectorStore ConnectorStatusStore

	// OrgID can be used for connector scoping.
	OrgID string

	// Tracer for distributed tracing (optional)
	Tracer trace.Tracer

	// Metrics for observability (optional)
	SyncDuration     metric.Float64Histogram
	SyncCount        metric.Int64Counter
	RecordsFetched   metric.Int64Counter
	RecordsProcessed metric.Int64Counter
	ErrorCount       metric.Int64Counter
}

// Run executes ingestion across registered adapters, validating and persisting
// the resulting activities. It returns all accepted activities; validation
// failures are returned alongside a partial ingest.
func (s *Service) Run(ctx context.Context) ([]Activity, error) {
	logger := s.Logger
	if logger == nil {
		logger = slog.Default()
	}

	// Start tracing span if tracer is configured
	if s.Tracer != nil {
		var span trace.Span
		ctx, span = s.Tracer.Start(ctx, "ingestion.Run",
			trace.WithAttributes(
				attribute.String("org_id", s.OrgID),
				attribute.Int("adapter_count", len(s.Adapters)),
			),
		)
		defer span.End()
	}

	startedAt := time.Now().UTC()
	logEntry := IngestionLog{
		ID:        fmt.Sprintf("ingest-%d", startedAt.UnixNano()),
		Source:    "multi",
		Status:    "running",
		StartedAt: startedAt,
		Errors:    []ImportError{},
		Processed: 0,
		Succeeded: 0,
		Failed:    0,
	}

	var (
		all         []Activity
		validation  []error
		adapterName string
	)

	for idx, adapter := range s.Adapters {
		adapterName = fmt.Sprintf("adapter_%d", idx)
		connector := connectorName(adapter)
		s.setConnectorStatus(ctx, connector, "running", "")
		logger.Info("starting ingestion", "adapter", adapterName)

		// Start adapter-specific span
		var adapterSpan trace.Span
		if s.Tracer != nil {
			ctx, adapterSpan = s.Tracer.Start(ctx, "ingestion.adapter",
				trace.WithAttributes(
					attribute.String("adapter", adapterName),
					attribute.String("connector", connector),
				),
			)
		}

		adapterErrors := make([]string, 0)
		adapterStart := time.Now()

		activities, err := adapter.Ingest(ctx)
		adapterDuration := time.Since(adapterStart)

		if err != nil {
			logEntry.Errors = append(logEntry.Errors, ImportError{Message: err.Error()})
			logEntry.Status = "failed"
			logEntry.Failed++
			s.setConnectorStatus(ctx, connector, "error", err.Error())
			if s.ConnectorStore != nil {
				_ = s.ConnectorStore.LastError(ctx, connector, s.OrgID, err)
			}
			_ = s.recordLog(ctx, logEntry)

			// Record error metrics
			if s.ErrorCount != nil {
				s.ErrorCount.Add(ctx, 1, metric.WithAttributes(
					attribute.String("connector", connector),
					attribute.String("error_type", "ingestion_failed"),
				))
			}

			// Record span error
			if adapterSpan != nil {
				adapterSpan.RecordError(err)
				adapterSpan.End()
			}

			return nil, fmt.Errorf("%s ingest failed: %w", adapterName, err)
		}

		// Record fetched records metric
		if s.RecordsFetched != nil {
			s.RecordsFetched.Add(ctx, int64(len(activities)), metric.WithAttributes(
				attribute.String("connector", connector),
			))
		}

		valid := make([]Activity, 0, len(activities))
		for i, act := range activities {
			if err := act.Validate(); err != nil {
				validation = append(validation, fmt.Errorf("%s activity[%d] invalid: %w", adapterName, i, err))
				logEntry.Errors = append(logEntry.Errors, ImportError{Row: i + 1, Message: err.Error()})
				adapterErrors = append(adapterErrors, err.Error())
				logEntry.Failed++
				continue
			}
			if act.CreatedAt.IsZero() {
				act.CreatedAt = time.Now().UTC()
			}
			valid = append(valid, act)
			logEntry.Succeeded++
		}

		// Record processed records metric
		if s.RecordsProcessed != nil {
			s.RecordsProcessed.Add(ctx, int64(len(valid)), metric.WithAttributes(
				attribute.String("connector", connector),
				attribute.String("status", "valid"),
			))
			if logEntry.Failed > 0 {
				s.RecordsProcessed.Add(ctx, int64(logEntry.Failed), metric.WithAttributes(
					attribute.String("connector", connector),
					attribute.String("status", "invalid"),
				))
			}
		}

		// Record sync metrics
		if s.SyncDuration != nil {
			s.SyncDuration.Record(ctx, adapterDuration.Seconds(), metric.WithAttributes(
				attribute.String("connector", connector),
			))
		}
		if s.SyncCount != nil {
			s.SyncCount.Add(ctx, 1, metric.WithAttributes(
				attribute.String("connector", connector),
				attribute.String("status", "completed"),
			))
		}

		all = append(all, valid...)
		logger.Info("ingestion completed", "adapter", adapterName, "received", len(activities), "accepted", len(valid))

		if len(adapterErrors) > 0 {
			s.setConnectorStatus(ctx, connector, "error", strings.Join(adapterErrors, "; "))
			if s.ConnectorStore != nil {
				_ = s.ConnectorStore.LastError(ctx, connector, s.OrgID, errors.Join(validation...))
			}
			// Record error count
			if s.ErrorCount != nil {
				s.ErrorCount.Add(ctx, int64(len(adapterErrors)), metric.WithAttributes(
					attribute.String("connector", connector),
					attribute.String("error_type", "validation_error"),
				))
			}
		} else {
			s.setConnectorStatus(ctx, connector, "connected", "")
		}

		// End adapter span
		if adapterSpan != nil {
			adapterSpan.SetAttributes(
				attribute.Int("records_fetched", len(activities)),
				attribute.Int("records_valid", len(valid)),
				attribute.Int("records_invalid", len(adapterErrors)),
			)
			adapterSpan.End()
		}
	}

	// Persist activities if store is configured
	if s.Store != nil && len(all) > 0 {
		if err := s.Store.SaveBatch(ctx, all); err != nil {
			logEntry.Errors = append(logEntry.Errors, ImportError{Message: err.Error()})
			logEntry.Status = "failed"
			logEntry.Failed++
			_ = s.recordLog(ctx, logEntry)

			// Record error metric
			if s.ErrorCount != nil {
				s.ErrorCount.Add(ctx, 1, metric.WithAttributes(
					attribute.String("connector", "storage"),
					attribute.String("error_type", "persist_failed"),
				))
			}

			return nil, fmt.Errorf("persist activities: %w", err)
		}
	}

	logEntry.Processed = logEntry.Succeeded + logEntry.Failed
	logEntry.CompletedAt = time.Now().UTC()
	if len(validation) > 0 {
		logEntry.Status = "completed_with_errors"
	} else {
		logEntry.Status = "completed"
	}
	_ = s.recordLog(ctx, logEntry)
	errSummary := joinErrors(validation)
	s.setConnectorStatus(ctx, "all", logEntry.Status, errSummary)
	if errSummary != "" && s.ConnectorStore != nil {
		_ = s.ConnectorStore.LastError(ctx, "all", s.OrgID, errors.Join(validation...))
	}

	if len(validation) > 0 {
		return all, errors.Join(validation...)
	}

	return all, nil
}

// recordLog writes to the configured log store if present.
func (s *Service) recordLog(ctx context.Context, log IngestionLog) error {
	if s.Logs == nil {
		return nil
	}
	return s.Logs.Record(ctx, log)
}

// setConnectorStatus updates connector store if configured.
func (s *Service) setConnectorStatus(ctx context.Context, name, status, errMsg string) {
	if s.ConnectorStore == nil {
		return
	}
	now := time.Now()
	_ = s.ConnectorStore.SetStatus(ctx, name, s.OrgID, status, errMsg, &now)
}

// connectorName derives a stable name from adapter type.
func connectorName(adapter SourceIngestionAdapter) string {
	t := fmt.Sprintf("%T", adapter)
	if strings.Contains(t, ".") {
		parts := strings.Split(t, ".")
		t = parts[len(parts)-1]
	}
	return strings.ToLower(strings.TrimSuffix(t, "Adapter"))
}

func joinErrors(errs []error) string {
	if len(errs) == 0 {
		return ""
	}
	var parts []string
	for _, e := range errs {
		parts = append(parts, e.Error())
	}
	return strings.Join(parts, "; ")
}
