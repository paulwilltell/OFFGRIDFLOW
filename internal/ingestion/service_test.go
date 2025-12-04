package ingestion

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubAdapter struct {
	Activities []Activity
	Err        error
}

func (s stubAdapter) Ingest(ctx context.Context) ([]Activity, error) {
	_ = ctx
	return s.Activities, s.Err
}

// TestServiceRun validates that the service aggregates, validates, and persists activities.
func TestServiceRun(t *testing.T) {
	store := NewInMemoryActivityStore()
	logs := NewInMemoryLogStore()
	now := time.Now().UTC()

	valid := Activity{
		ID:          "ok-1",
		Source:      "utility_bill",
		Location:    "US-WEST",
		PeriodStart: now.AddDate(0, 0, -30),
		PeriodEnd:   now,
		Quantity:    10,
		Unit:        "kWh",
		OrgID:       "org-1",
	}

	invalid := Activity{
		ID:          "",
		Source:      "utility_bill",
		Location:    "US-WEST",
		PeriodStart: now.AddDate(0, 0, -30),
		PeriodEnd:   now.AddDate(0, 0, -31), // invalid period
		Quantity:    -1,
		Unit:        "kWh",
		OrgID:       "org-1",
	}

	svc := Service{
		Adapters: []SourceIngestionAdapter{
			stubAdapter{Activities: []Activity{valid, invalid}},
		},
		Store: store,
		Logs:  logs,
	}

	acts, err := svc.Run(context.Background())
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
	if !errors.Is(err, ErrInvalidPeriod) && !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(acts) != 1 {
		t.Fatalf("expected 1 valid activity, got %d", len(acts))
	}

	saved, _ := store.List(context.Background())
	if len(saved) != 1 {
		t.Fatalf("expected 1 persisted activity, got %d", len(saved))
	}

	if saved[0].ID != "ok-1" {
		t.Fatalf("unexpected activity persisted: %+v", saved[0])
	}

	if len(logs.Logs) == 0 || logs.Logs[0].Processed == 0 {
		t.Fatalf("expected ingestion log entry to be recorded")
	}
}

// TestServiceRun_AdapterError ensures adapter errors bubble up.
func TestServiceRun_AdapterError(t *testing.T) {
	svc := Service{
		Adapters: []SourceIngestionAdapter{
			stubAdapter{Err: errors.New("boom")},
		},
	}

	_, err := svc.Run(context.Background())
	if err == nil {
		t.Fatalf("expected adapter error")
	}
}
