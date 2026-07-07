package emissions

import (
	"context"
	"testing"
)

// emptyRegistry always misses — mirrors production, where the configured factor
// registry holds curated grid (Scope 2) factors but not the full Scope 1/3 set.
type emptyRegistry struct{}

func (emptyRegistry) GetFactor(context.Context, string) (EmissionFactor, error) {
	return EmissionFactor{}, ErrFactorNotFound
}
func (emptyRegistry) FindFactor(context.Context, FactorQuery) (EmissionFactor, error) {
	return EmissionFactor{}, ErrFactorNotFound
}
func (emptyRegistry) ListFactors(context.Context, FactorQuery) ([]EmissionFactor, error) {
	return nil, nil
}
func (emptyRegistry) RegisterFactor(context.Context, EmissionFactor) error { return nil }

// When a registry is configured but lacks a factor, Scope 1 and Scope 3 must
// fall back to the built-in published factor tables — not silently produce
// zero. This is the exact production bug the full audit caught.
func TestCalculatorsFallBackWhenRegistryMisses(t *testing.T) {
	reg := emptyRegistry{}
	ctx := context.Background()

	s1 := NewScope1Calculator(Scope1Config{Registry: reg})
	rec, err := s1.Calculate(ctx, ActivityAdapter{
		ID: "f1", Source: "fleet", Category: "diesel", Location: "US-CA",
		Quantity: 100, Unit: "L", OrgID: "o",
	})
	if err != nil {
		t.Fatalf("scope1 with empty registry should fall back, got error: %v", err)
	}
	if rec.EmissionsTonnesCO2e <= 0 {
		t.Fatalf("scope1 fell back but produced zero emissions: %+v", rec)
	}

	s3 := NewScope3Calculator(DefaultScope3Config())
	// Re-create with the empty registry.
	s3 = NewScope3Calculator(Scope3Config{Registry: reg, DefaultCategory: CategoryPurchasedGoods, SpendCurrency: "USD"})
	rec3, err := s3.Calculate(ctx, ActivityAdapter{
		ID: "t1", Source: "travel", Category: "flight", Location: "US-CA",
		Quantity: 1000, Unit: "mile", OrgID: "o",
	})
	if err != nil {
		t.Fatalf("scope3 with empty registry should fall back, got error: %v", err)
	}
	if rec3.EmissionsTonnesCO2e <= 0 {
		t.Fatalf("scope3 fell back but produced zero emissions: %+v", rec3)
	}
}
