package compliance

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/ingestion"
)

func seedInventoryStore(t *testing.T) ingestion.ActivityStore {
	t.Helper()
	store := ingestion.NewInMemoryActivityStore()
	y := 2025
	acts := []ingestion.Activity{
		{ID: "e1", OrgID: "test-org", Source: "utility_bill", Category: "electricity",
			Location: "US-CA", Unit: "kWh", Quantity: 120000,
			PeriodStart: time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC), PeriodEnd: time.Date(y, 1, 31, 0, 0, 0, 0, time.UTC), CreatedAt: time.Now()},
		{ID: "f1", OrgID: "test-org", Source: "fleet", Category: "diesel",
			Location: "US-CA", Unit: "L", Quantity: 5000,
			PeriodStart: time.Date(y, 2, 1, 0, 0, 0, 0, time.UTC), PeriodEnd: time.Date(y, 2, 28, 0, 0, 0, 0, time.UTC), CreatedAt: time.Now()},
		{ID: "t1", OrgID: "test-org", Source: "travel", Category: "flight",
			Location: "US-CA", Unit: "mile", Quantity: 8400,
			PeriodStart: time.Date(y, 3, 1, 0, 0, 0, 0, time.UTC), PeriodEnd: time.Date(y, 3, 5, 0, 0, 0, 0, time.UTC), CreatedAt: time.Now()},
		{ID: "s1", OrgID: "test-org", Source: "purchases", Category: "cloud-services",
			Location: "US-CA", Unit: "USD", Quantity: 40000,
			PeriodStart: time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC), PeriodEnd: time.Date(y, 3, 31, 0, 0, 0, 0, time.UTC), CreatedAt: time.Now()},
	}
	if err := store.SaveBatch(context.Background(), acts); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return store
}

func testService(t *testing.T) *Service {
	t.Helper()
	return NewService(
		seedInventoryStore(t),
		emissions.NewScope1Calculator(emissions.Scope1Config{}),
		emissions.NewScope2Calculator(emissions.Scope2Config{}),
		emissions.NewScope3Calculator(emissions.DefaultScope3Config()),
	)
}

// The inventory must gather every scope's per-activity records with a full
// factor audit trail — the backbone of an assurance-ready report.
func TestGenerateInventory(t *testing.T) {
	inv, err := testService(t).GenerateInventory(context.Background(), "test-org", 2025)
	if err != nil {
		t.Fatalf("GenerateInventory: %v", err)
	}

	if inv.Scope1Tonnes <= 0 || inv.Scope2Tonnes <= 0 || inv.Scope3Tonnes <= 0 {
		t.Fatalf("expected positive emissions in all scopes, got s1=%.3f s2=%.3f s3=%.3f",
			inv.Scope1Tonnes, inv.Scope2Tonnes, inv.Scope3Tonnes)
	}
	if inv.TotalTonnes <= 0 {
		t.Fatalf("expected positive total")
	}
	if len(inv.LineItems) != 4 {
		t.Fatalf("expected 4 line items, got %d", len(inv.LineItems))
	}
	// Every line item must carry a factor + source (audit trail).
	for _, li := range inv.LineItems {
		if li.FactorID == "" || li.FactorSource == "" || li.EmissionsTonnes <= 0 {
			t.Errorf("line item missing audit fields: %+v", li)
		}
	}
	if len(inv.Factors) == 0 {
		t.Fatal("expected distinct emission factors captured")
	}
	if inv.CompletenessPct != 100 {
		t.Errorf("expected 100%% completeness for fully-specified activities, got %.1f", inv.CompletenessPct)
	}
}

// The PDF must be a real, non-trivial document.
func TestExportInventoryPDF(t *testing.T) {
	inv, err := testService(t).GenerateInventory(context.Background(), "test-org", 2025)
	if err != nil {
		t.Fatalf("inventory: %v", err)
	}
	inv.OrgName = "Acme Manufacturing Inc."

	data, err := ExportInventoryReportPDF(inv)
	if err != nil {
		t.Fatalf("pdf: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Fatalf("output is not a PDF (prefix %q)", data[:min(8, len(data))])
	}
	if len(data) < 5000 {
		t.Fatalf("PDF suspiciously small (%d bytes) — likely missing content", len(data))
	}

	// Write to disk for manual eyeballing when OGF_WRITE_REPORT is set.
	if os.Getenv("OGF_WRITE_REPORT") != "" {
		_ = os.WriteFile("sample_report.pdf", data, 0o644)
	}
}

// The CSV must be self-describing with a header and one row per activity.
func TestExportInventoryCSV(t *testing.T) {
	inv, err := testService(t).GenerateInventory(context.Background(), "test-org", 2025)
	if err != nil {
		t.Fatalf("inventory: %v", err)
	}
	data, err := ExportInventoryCSV(inv)
	if err != nil {
		t.Fatalf("csv: %v", err)
	}
	s := string(data)
	for _, col := range []string{"scope", "emission_factor", "factor_source", "emissions_tco2e"} {
		if !strings.Contains(s, col) {
			t.Errorf("CSV missing column %q", col)
		}
	}
	if !strings.Contains(s, "GHG Protocol Corporate Standard") {
		t.Error("CSV missing provenance header")
	}
	lines := strings.Count(strings.TrimSpace(s), "\n") + 1
	if lines < 4+len(inv.LineItems) {
		t.Errorf("CSV has too few lines (%d) for %d activities", lines, len(inv.LineItems))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Scope 2 must be dual-reported: location-based (grid) and market-based
// (supplier factor / renewable share). Market emissions fall below location
// when contractual instruments are supplied.
func TestScope2MarketBased(t *testing.T) {
	store := ingestion.NewInMemoryActivityStore()
	y := 2025
	acts := []ingestion.Activity{
		// 100% renewable -> market-based emissions ~0.
		{ID: "e-green", OrgID: "test-org", Source: "utility_bill", Category: "electricity",
			Location: "US-CA", Unit: "kWh", Quantity: 100000,
			Metadata:    map[string]string{"renewable_pct": "100"},
			PeriodStart: time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC), PeriodEnd: time.Date(y, 1, 31, 0, 0, 0, 0, time.UTC), CreatedAt: time.Now()},
		// Supplier-specific factor 0.05 kgCO2e/kWh -> market = 100000*0.05/1000 = 5t.
		{ID: "e-supplier", OrgID: "test-org", Source: "utility_bill", Category: "electricity",
			Location: "US-CA", Unit: "kWh", Quantity: 100000,
			Metadata:    map[string]string{"market_factor": "0.05"},
			PeriodStart: time.Date(y, 2, 1, 0, 0, 0, 0, time.UTC), PeriodEnd: time.Date(y, 2, 28, 0, 0, 0, 0, time.UTC), CreatedAt: time.Now()},
	}
	if err := store.SaveBatch(context.Background(), acts); err != nil {
		t.Fatalf("seed: %v", err)
	}
	svc := NewService(store,
		emissions.NewScope1Calculator(emissions.Scope1Config{}),
		emissions.NewScope2Calculator(emissions.Scope2Config{}),
		emissions.NewScope3Calculator(emissions.DefaultScope3Config()))

	inv, err := svc.GenerateInventory(context.Background(), "test-org", y)
	if err != nil {
		t.Fatalf("inventory: %v", err)
	}
	if !inv.HasMarketData {
		t.Fatal("expected HasMarketData true")
	}
	if inv.Scope2Tonnes <= 0 {
		t.Fatalf("expected positive location-based scope 2, got %.3f", inv.Scope2Tonnes)
	}
	// Market-based must be lower than location-based (green + low supplier factor).
	if inv.Scope2MarketTonnes >= inv.Scope2Tonnes {
		t.Fatalf("expected market-based (%.3f) < location-based (%.3f)", inv.Scope2MarketTonnes, inv.Scope2Tonnes)
	}
	// Supplier factor row: 100000 kWh * 0.05 / 1000 = 5t; green row ~0 -> market ~5t.
	if inv.Scope2MarketTonnes < 4.5 || inv.Scope2MarketTonnes > 5.5 {
		t.Errorf("expected market-based ~5 tCO2e, got %.3f", inv.Scope2MarketTonnes)
	}
}

// Scope 1 must be disaggregated into the seven Kyoto gases, with CO2/CH4/N2O
// reconciling exactly to the fossil Scope 1 total.
func TestScope1GasBreakdown(t *testing.T) {
	inv, err := testService(t).GenerateInventory(context.Background(), "test-org", 2025)
	if err != nil {
		t.Fatalf("inventory: %v", err)
	}
	g := inv.Scope1Gases
	if g.CO2 <= 0 {
		t.Fatalf("expected positive CO2, got %.3f", g.CO2)
	}
	if g.CH4 < 0 || g.N2O < 0 {
		t.Fatal("gas shares must be non-negative")
	}
	// CO2 must dominate combustion.
	if g.CO2 < inv.Scope1Tonnes*0.9 {
		t.Errorf("expected CO2 to dominate Scope 1; CO2=%.3f scope1=%.3f", g.CO2, inv.Scope1Tonnes)
	}
	// The gas split must reconcile to the Scope 1 total.
	diff := g.Total() - inv.Scope1Tonnes
	if diff < -0.001 || diff > 0.001 {
		t.Errorf("gas breakdown %.4f does not reconcile to Scope 1 total %.4f", g.Total(), inv.Scope1Tonnes)
	}
	// Industrial gases zero (no fugitive sources in this inventory).
	if g.HFCs != 0 || g.SF6 != 0 {
		t.Error("expected zero industrial gases with no fugitive sources")
	}
}
