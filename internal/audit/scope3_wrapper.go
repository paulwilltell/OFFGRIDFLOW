package audit

import (
	"context"
	"fmt"

	"github.com/example/offgridflow/internal/emissions"
)

// AuditedScope3Calculator wraps a Scope3Calculator with ledger recording.
type AuditedScope3Calculator struct {
	inner *emissions.Scope3Calculator
	store *Store
}

func NewAuditedScope3Calculator(calc *emissions.Scope3Calculator, store *Store) *AuditedScope3Calculator {
	return &AuditedScope3Calculator{inner: calc, store: store}
}

func (a *AuditedScope3Calculator) Calculate(ctx context.Context, activity emissions.Activity) (emissions.EmissionRecord, error) {
	record, err := a.inner.Calculate(ctx, activity)
	if err != nil {
		return record, err
	}

	formula := fmt.Sprintf(
		"%.4f %s × %.6f kg CO2e/%s = %.4f kg CO2e (%.6f tonnes) [category: %s]",
		activity.GetQuantity(), activity.GetUnit(),
		record.EmissionFactor, activity.GetUnit(),
		record.EmissionsKgCO2e, record.EmissionsTonnesCO2e,
		activity.GetCategory(),
	)

	entry := &LedgerEntry{
		TenantID:             record.OrgID,
		Scope:                "scope3",
		Category:             activity.GetCategory(),
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

	if a.store != nil {
		_ = a.store.RecordCalculation(ctx, entry)
	}

	return record, nil
}

func (a *AuditedScope3Calculator) Inner() *emissions.Scope3Calculator {
	return a.inner
}
