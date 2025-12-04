package csv_upload

import (
	"context"
	"fmt"
	"io"

	"github.com/example/offgridflow/internal/ingestion"
	"github.com/example/offgridflow/internal/ingestion/parser"
)

// Adapter ingests user-uploaded CSV activity files.
type Adapter struct {
	Store ingestion.ActivityStore
}

// NewAdapter creates a new CSV adapter with the given store.
func NewAdapter(store ingestion.ActivityStore) *Adapter {
	return &Adapter{Store: store}
}

// Ingest implements SourceIngestionAdapter (returns empty for manual CSV upload use case).
func (a *Adapter) Ingest(ctx context.Context) ([]ingestion.Activity, error) {
	// This adapter is designed for manual CSV uploads via IngestUtilityCSV
	_ = ctx
	return []ingestion.Activity{}, nil
}

// IngestUtilityCSV parses a utility bill CSV and returns activities.
//
// Expected CSV format (with header row):
//
//	meter_id,location,period_start,period_end,kwh
//
// Date format: RFC3339 or "2006-01-02"
// Example row: METER-001,US-WEST,2025-01-01,2025-01-31,12500.5
func (a *Adapter) IngestUtilityCSV(ctx context.Context, r io.Reader, orgID string) ([]ingestion.Activity, error) {
	p := parser.CSVParser{}
	activities, errs, err := p.ParseUtilityBills(ctx, r, orgID)
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		return activities, fmt.Errorf("csv ingest encountered %d row errors (first: %s)", len(errs), errs[0].Message)
	}

	// Save to store if configured
	if a.Store != nil && len(activities) > 0 {
		if err := a.Store.SaveBatch(ctx, activities); err != nil {
			return nil, fmt.Errorf("failed to save activities: %w", err)
		}
	}

	return activities, nil
}
