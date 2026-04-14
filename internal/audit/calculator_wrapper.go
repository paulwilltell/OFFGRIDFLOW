package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/example/offgridflow/internal/emissions"
)

// AuditedScope2Calculator wraps a Scope2Calculator and records every calculation
// into the immutable calculation ledger with full formula transparency.
type AuditedScope2Calculator struct {
	inner *emissions.Scope2Calculator
	store *Store
}

// NewAuditedScope2Calculator wraps an existing calculator with audit logging.
func NewAuditedScope2Calculator(calc *emissions.Scope2Calculator, store *Store) *AuditedScope2Calculator {
	return &AuditedScope2Calculator{inner: calc, store: store}
}

// Calculate performs the emission calculation and records it in the ledger.
func (a *AuditedScope2Calculator) Calculate(ctx context.Context, activity emissions.Activity) (emissions.EmissionRecord, error) {
	record, err := a.inner.Calculate(ctx, activity)
	if err != nil {
		return record, err
	}

	// Build human-readable formula
	formula := fmt.Sprintf(
		"%.4f %s × %.6f kg CO2e/%s = %.4f kg CO2e (%.6f tonnes CO2e)",
		activity.GetQuantity(), activity.GetUnit(),
		record.EmissionFactor, activity.GetUnit(),
		record.EmissionsKgCO2e, record.EmissionsTonnesCO2e,
	)

	// Record in immutable ledger
	entry := &LedgerEntry{
		TenantID:             record.OrgID,
		Scope:                "scope2",
		ActivityID:           record.ActivityID,
		Quantity:             activity.GetQuantity(),
		Unit:                 activity.GetUnit(),
		EmissionFactorID:     record.FactorID,
		EmissionFactorValue:  record.EmissionFactor,
		EmissionFactorSource: string(record.Method),
		EmissionFactorRegion: record.Region,
		Method:               string(record.Method),
		ResultKgCO2e:         record.EmissionsKgCO2e,
		ResultTonnesCO2e:     record.EmissionsTonnesCO2e,
		PeriodStart:          formatDate(record.PeriodStart),
		PeriodEnd:            formatDate(record.PeriodEnd),
		CalculatedAt:         record.CalculatedAt,
		Formula:              formula,
	}

	// Best-effort ledger write — don't fail the calculation if audit logging fails
	if a.store != nil {
		_ = a.store.RecordCalculation(ctx, entry)
	}

	return record, nil
}

// CalculateBatch processes multiple activities with audit logging.
func (a *AuditedScope2Calculator) CalculateBatch(ctx context.Context, activities []emissions.Activity) ([]emissions.EmissionRecord, error) {
	records := make([]emissions.EmissionRecord, 0, len(activities))
	for _, activity := range activities {
		rec, err := a.Calculate(ctx, activity)
		if err != nil {
			continue // Skip failed calculations, same as inner behavior
		}
		records = append(records, rec)
	}
	return records, nil
}

// Inner returns the underlying calculator for direct access.
func (a *AuditedScope2Calculator) Inner() *emissions.Scope2Calculator {
	return a.inner
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}
