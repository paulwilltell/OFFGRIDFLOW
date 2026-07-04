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

// IngestUtilityCSV parses an uploaded activity CSV and returns the structured
// activities, saving them to the store when configured.
//
// The parser auto-detects the GHG scope from the CSV's columns, so a user can
// "just dump" any of these and have each row structured into the right scope:
//
//   - electricity / utility usage  -> Scope 2  (meter_id, usage/kwh, period)
//   - fuel / fleet / combustion     -> Scope 1  (fuel_type, gallons/liters/therms)
//   - business travel               -> Scope 3  (miles/km, mode, origin/destination)
//   - supplier spend                -> Scope 3  (vendor, amount_usd, category)
//
// Column names are matched by synonym, and dates accept RFC3339 or "2006-01-02".
func (a *Adapter) IngestUtilityCSV(ctx context.Context, r io.Reader, orgID string) ([]ingestion.Activity, error) {
	p := parser.NewUtilityBillParser(orgID, "")
	result, err := p.Parse(ctx, "upload.csv", r)
	if err != nil {
		return nil, err
	}
	if result != nil && len(result.Errors) > 0 {
		return result.Activities, fmt.Errorf("csv ingest encountered %d row errors (first: %s)",
			len(result.Errors), result.Errors[0].Message)
	}

	activities := result.Activities

	// Save to store if configured
	if a.Store != nil && len(activities) > 0 {
		if err := a.Store.SaveBatch(ctx, activities); err != nil {
			return nil, fmt.Errorf("failed to save activities: %w", err)
		}
	}

	return activities, nil
}
