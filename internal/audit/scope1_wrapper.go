package audit

import (
	"context"
	"fmt"

	"github.com/example/offgridflow/internal/emissions"
)

// AuditedScope1Calculator wraps a Scope1Calculator with ledger recording.
type AuditedScope1Calculator struct {
	inner *emissions.Scope1Calculator
	store *Store
}

func NewAuditedScope1Calculator(calc *emissions.Scope1Calculator, store *Store) *AuditedScope1Calculator {
	return &AuditedScope1Calculator{inner: calc, store: store}
}

func (a *AuditedScope1Calculator) Calculate(ctx context.Context, activity emissions.Activity) (emissions.EmissionRecord, error) {
	record, err := a.inner.Calculate(ctx, activity)
	if err != nil {
		return record, err
	}

	formula := fmt.Sprintf(
		"%.4f %s × %.6f kg CO2e/%s = %.4f kg CO2e (%.6f tonnes)",
		activity.GetQuantity(), activity.GetUnit(),
		record.EmissionFactor, activity.GetUnit(),
		record.EmissionsKgCO2e, record.EmissionsTonnesCO2e,
	)

	entry := &LedgerEntry{
		TenantID:             record.OrgID,
		Scope:                "scope1",
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

func (a *AuditedScope1Calculator) Inner() *emissions.Scope1Calculator {
	return a.inner
}
