package emissions

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/example/offgridflow/internal/ingestion/parser"
)

// TestRawDumpToAuditStructure proves the core promise: a user dumps a messy,
// real-world utility CSV (non-canonical column names, mixed case, extra
// columns) and the engine parses, structures, and calculates audit-ready
// emission records without any manual mapping.
func TestRawDumpToAuditStructure(t *testing.T) {
	// A "dumped" CSV as a real customer would export it: column names use
	// synonyms the canonical schema does NOT list verbatim (account_number,
	// usage, service_start/end), plus an unrelated trailing column.
	rawDump := strings.Join([]string{
		"account_number,grid_region,service_start,service_end,usage,notes",
		"ACC-4001,US-CA,2025-01-01,2025-01-31,21400,Placerville yard",
		"ACC-3001,US-CA,2025-01-01,2025-01-31,28700,Folsom retail",
		"ACC-2002,US-CA,2025-01-01,2025-01-31,43300,Sac warehouse HVAC",
		"ACC-1001,US-CA,2025-01-01,2025-01-31,36200,EDH HQ",
	}, "\n")

	p := parser.NewUtilityBillParser("test-org", "US-CA")
	result, err := p.Parse(context.Background(), "dump.csv", strings.NewReader(rawDump))
	if err != nil {
		t.Fatalf("engine could not parse a raw dump with synonym columns: %v", err)
	}

	if len(result.Activities) != 4 {
		t.Fatalf("expected 4 structured activities from raw dump, got %d (errors: %v)",
			len(result.Activities), result.Errors)
	}

	// Every parsed row must carry the fields a report needs: meter, location,
	// period, quantity, unit.
	for i, a := range result.Activities {
		if a.MeterID == "" || a.Location == "" || a.Quantity <= 0 || a.PeriodStart.IsZero() {
			t.Fatalf("row %d not fully structured: %+v", i, a)
		}
	}

	// Now run the REAL Scope 2 calculator over the structured activities and
	// build the aggregate an audit report consumes.
	calc := NewScope2Calculator(DefaultScope2Config())

	var totalTonnes float64
	perFactor := map[string]int{}
	for _, a := range result.Activities {
		adapter := ActivityAdapter{
			ID:          a.ID,
			Source:      "utility_bill",
			Category:    "electricity",
			Location:    a.Location,
			Quantity:    a.Quantity,
			Unit:        "kWh",
			PeriodStart: a.PeriodStart,
			PeriodEnd:   a.PeriodEnd,
			OrgID:       "test-org",
		}
		rec, err := calc.Calculate(context.Background(), adapter)
		if err != nil {
			t.Fatalf("calculator failed on structured activity %s: %v", a.MeterID, err)
		}
		// Audit-critical fields must be populated.
		if rec.FactorID == "" || rec.Method == "" || rec.EmissionsTonnesCO2e <= 0 {
			t.Fatalf("emission record missing audit fields: %+v", rec)
		}
		if rec.CalculatedAt.IsZero() {
			t.Fatalf("emission record missing calculation timestamp (audit trail)")
		}
		totalTonnes += rec.EmissionsTonnesCO2e
		perFactor[rec.FactorID]++
	}

	if totalTonnes <= 0 {
		t.Fatalf("expected a positive total footprint, got %.4f", totalTonnes)
	}

	// Emit the structured audit summary so we can eyeball it in test output.
	fmt.Printf("\n=== RAW DUMP -> AUDIT STRUCTURE ===\n")
	fmt.Printf("Input: 4 rows, non-canonical columns (account_number/usage/service_start)\n")
	fmt.Printf("Parsed & structured: %d activities\n", len(result.Activities))
	fmt.Printf("Total Scope 2 footprint: %.2f tCO2e\n", totalTonnes)
	fmt.Printf("Emission factors applied: %d distinct\n", len(perFactor))
	fmt.Printf("Schema auto-detected: %s\n", result.Metadata["schema_type"])
	fmt.Printf("===================================\n")
}

// TestRawDumpBoundary documents the ENGINE'S ACTUAL BOUNDARY: what happens
// when a user dumps data that is not tidy utility/energy data. This is an
// honest coverage probe, not a pass/fail on quality.
func TestRawDumpBoundary(t *testing.T) {
	cases := map[string]string{
		"fuel_receipts": strings.Join([]string{
			"date,vehicle,fuel_type,gallons,cost",
			"2025-01-15,TRUCK-1,diesel,45.2,180.50",
		}, "\n"),
		"business_travel": strings.Join([]string{
			"employee,origin,destination,miles,mode",
			"J. Smith,SFO,JFK,2586,air",
		}, "\n"),
		"supplier_spend": strings.Join([]string{
			"vendor,category,amount_usd,quarter",
			"Acme Steel,raw_materials,45000,Q1",
		}, "\n"),
	}

	p := parser.NewUtilityBillParser("test-org", "US-CA")
	for name, csv := range cases {
		result, err := p.Parse(context.Background(), name+".csv", strings.NewReader(csv))
		status := "PARSED"
		detail := ""
		if err != nil {
			status = "REJECTED"
			detail = err.Error()
		} else if result != nil {
			detail = fmt.Sprintf("%d activities", len(result.Activities))
		}
		fmt.Printf("[boundary] %-16s -> %s (%s)\n", name, status, detail)

		// Current engine boundary: only utility/energy (Scope 2) data is
		// accepted. Fuel (Scope 1), travel and supplier spend (Scope 3) are
		// rejected. When Scope 1/3 ingestion lands, THIS ASSERTION SHOULD FAIL
		// and be updated — it is the tripwire that keeps the product's
		// "Scope 1/2/3" promise honest.
		if err == nil {
			t.Errorf("boundary changed: %q now parses — update the Scope 1/2/3 coverage claim and this test", name)
		}
	}
}
