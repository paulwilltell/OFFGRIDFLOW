package emissions

import (
	"context"
	"math"
	"testing"
)

func TestScope2SupportsMMBtuVariants(t *testing.T) {
	calc := NewScope2Calculator(DefaultScope2Config())

	activity := ActivityAdapter{
		ID:       "act-steam-uppercase",
		Source:   "utility_bill",
		Category: "Purchased Heat",
		Quantity: 1,
		Unit:     "MMBTU",
	}

	if !calc.Supports(activity) {
		t.Fatalf("expected calculator to support MMBTU unit")
	}

	activity.Unit = "mmbtu"
	if !calc.Supports(activity) {
		t.Fatalf("expected calculator to support lowercase mmbtu unit")
	}
}

func TestScope2CalculateSteamUsesThermalFallback(t *testing.T) {
	calc := NewScope2Calculator(Scope2Config{})

	activity := ActivityAdapter{
		ID:       "act-steam",
		Source:   "utility_bill",
		Category: "Purchased Heat",
		Location: "Detroit, MI",
		Quantity: 45000,
		Unit:     "MMBTU",
	}

	record, err := calc.Calculate(context.Background(), activity)
	if err != nil {
		t.Fatalf("calculate returned error: %v", err)
	}

	// 45,000 MMBtu * 14.7 kgCO2e/MMBtu = 661,500 kg = 661.5 tCO2e.
	expectedTonnes := 661.5
	if math.Abs(record.EmissionsTonnesCO2e-expectedTonnes) > 0.01 {
		t.Fatalf("unexpected steam emissions: got %.4f, want %.4f", record.EmissionsTonnesCO2e, expectedTonnes)
	}

	if record.FactorID == "default-scope2-global" {
		t.Fatalf("expected steam fallback factor, got electricity global default")
	}
}

