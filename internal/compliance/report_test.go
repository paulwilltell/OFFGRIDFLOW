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
