//go:build ignore

// generate-example-reports.go produces a reference copy of the flagship GHG
// Emissions Inventory report (the same PDF + CSV customers pay for) from a
// realistic, deliberately messy multi-scope data dump. It drives the real
// engine end-to-end -- ingestion classifier -> scope calculators -> inventory
// assembly -> PDF/CSV export -- so it doubles as a smoke test that arbitrary
// dumped data is sorted, calculated and organized correctly.
//
// Run: go run scripts/generate-example-reports.go
// Output: examples/reports/ghg-inventory-sample.pdf and .csv

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/example/offgridflow/internal/compliance"
	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/ingestion"
)

func main() {
	fmt.Println("Generating flagship GHG inventory sample from a mixed data dump...")

	ctx := context.Background()
	store := ingestion.NewInMemoryActivityStore()

	// A messy, mixed data dump: multiple scopes, units and categories jumbled
	// together the way a real customer would export them. The engine must sort
	// each into the right scope and apply the right factor.
	y := 2025
	dump := []ingestion.Activity{
		{ID: "elec-hq-jan", OrgID: "sample-org", Source: "utility_bill", Category: "electricity", Location: "US-CA", Unit: "kWh", Quantity: 128000, PeriodStart: date(y, 1), PeriodEnd: date(y, 1)},
		{ID: "elec-hq-feb", OrgID: "sample-org", Source: "utility_bill", Category: "electricity", Location: "US-CA", Unit: "kWh", Quantity: 119500, PeriodStart: date(y, 2), PeriodEnd: date(y, 2)},
		{ID: "fleet-diesel-q1", OrgID: "sample-org", Source: "fleet", Category: "diesel", Location: "US-CA", Unit: "L", Quantity: 9200, PeriodStart: date(y, 1), PeriodEnd: date(y, 3)},
		{ID: "natgas-heating", OrgID: "sample-org", Source: "stationary_combustion", Category: "natural_gas", Location: "US-CA", Unit: "therm", Quantity: 4300, PeriodStart: date(y, 1), PeriodEnd: date(y, 3)},
		{ID: "biodiesel-generator", OrgID: "sample-org", Source: "fleet", Category: "biodiesel", Location: "US-CA", Unit: "L", Quantity: 800, PeriodStart: date(y, 2), PeriodEnd: date(y, 2)},
		{ID: "travel-sfo-jfk", OrgID: "sample-org", Source: "travel", Category: "flight_long_haul", Location: "GLOBAL", Unit: "mile", Quantity: 21000, PeriodStart: date(y, 3), PeriodEnd: date(y, 3)},
		{ID: "commute-staff", OrgID: "sample-org", Source: "commuting", Category: "car", Location: "US-CA", Unit: "mile", Quantity: 46000, PeriodStart: date(y, 1), PeriodEnd: date(y, 3)},
		{ID: "waste-landfill", OrgID: "sample-org", Source: "waste", Category: "landfill", Location: "US-CA", Unit: "kg", Quantity: 5400, PeriodStart: date(y, 1), PeriodEnd: date(y, 3)},
		{ID: "freight-inbound", OrgID: "sample-org", Source: "freight", Category: "truck", Location: "US-CA", Unit: "tonne-km", Quantity: 12000, PeriodStart: date(y, 1), PeriodEnd: date(y, 3)},
		{ID: "cloud-spend", OrgID: "sample-org", Source: "purchases", Category: "cloud-services", Location: "US-CA", Unit: "USD", Quantity: 88000, PeriodStart: date(y, 1), PeriodEnd: date(y, 3)},
	}
	if err := store.SaveBatch(ctx, dump); err != nil {
		fail("seed dump", err)
	}

	svc := compliance.NewService(
		store,
		emissions.NewScope1Calculator(emissions.Scope1Config{}),
		emissions.NewScope2Calculator(emissions.Scope2Config{}),
		emissions.NewScope3Calculator(emissions.DefaultScope3Config()),
	)

	inv, err := svc.GenerateInventory(ctx, "sample-org", y)
	if err != nil {
		fail("GenerateInventory", err)
	}
	inv.OrgName = "Northwind Manufacturing, Inc."
	inv.Revenue = 1_200_000_000 // $1.2B (SB 253 threshold band)
	inv.Employees = 850

	fmt.Printf("  Sorted %d dumped rows -> S1=%.2f  S2=%.2f  S3=%.2f  total=%.2f tCO2e\n",
		len(dump), inv.Scope1Tonnes, inv.Scope2Tonnes, inv.Scope3Tonnes, inv.TotalTonnes)
	fmt.Printf("  Intensity: %.2f tCO2e/$M revenue, %.3f tCO2e/FTE\n",
		inv.IntensityPerRevenueMM(), inv.IntensityPerEmployee())

	if err := os.MkdirAll("examples/reports", 0o755); err != nil {
		fail("mkdir", err)
	}
	pdf, err := compliance.ExportInventoryReportPDF(inv)
	if err != nil {
		fail("pdf", err)
	}
	if err := os.WriteFile("examples/reports/ghg-inventory-sample.pdf", pdf, 0o644); err != nil {
		fail("write pdf", err)
	}
	csv, err := compliance.ExportInventoryCSV(inv)
	if err != nil {
		fail("csv", err)
	}
	if err := os.WriteFile("examples/reports/ghg-inventory-sample.csv", csv, 0o644); err != nil {
		fail("write csv", err)
	}

	fmt.Printf("  Wrote examples/reports/ghg-inventory-sample.pdf (%.1f KB) and .csv\n", float64(len(pdf))/1024)
	fmt.Println("Done.")
}

func date(year, month int) time.Time {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
}

func fail(stage string, err error) {
	fmt.Fprintf(os.Stderr, "  ERROR (%s): %v\n", stage, err)
	os.Exit(1)
}
