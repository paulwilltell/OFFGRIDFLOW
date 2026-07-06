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

// TestRawDumpMultiScope proves the extended promise: a user can dump fuel
// (Scope 1), business travel (Scope 3 cat 6), or supplier spend (Scope 3 cat 1)
// and the engine classifies each dump into the correct scope, structures it,
// and produces audit-complete emission records with the real calculators.
//
// This replaces the old TestRawDumpBoundary tripwire, which asserted these
// dumps were REJECTED. Now that Scope 1/3 ingestion has landed, the boundary
// has moved and the promise is that they are ACCEPTED and correctly scoped.
func TestRawDumpMultiScope(t *testing.T) {
	type expect struct {
		csv         string
		source      string // Source the parser must tag the activity with
		scope       Scope  // scope the record must land in
		calcFor     Scope  // which calculator to run
		minCategory string // substring the record notes/category should reflect (optional)
	}

	cases := map[string]expect{
		"fuel_receipts": {
			csv: strings.Join([]string{
				"date,vehicle,fuel_type,gallons,cost",
				"2025-01-15,TRUCK-1,diesel,45.2,180.50",
				"2025-01-20,TRUCK-2,gasoline,30.0,120.00",
			}, "\n"),
			source: "fleet",
			scope:  Scope1,
		},
		"business_travel": {
			csv: strings.Join([]string{
				"employee,origin,destination,miles,mode,date",
				"J. Smith,SFO,JFK,2586,air,2025-02-10",
				"A. Lee,SFO,LAX,338,air,2025-02-12",
			}, "\n"),
			source: "travel",
			scope:  Scope3,
		},
		"supplier_spend": {
			csv: strings.Join([]string{
				"vendor,category,amount_usd,quarter,year",
				"Acme Steel,raw_materials,45000,Q1,2025",
				"CloudCo,cloud-services,8000,Q1,2025",
			}, "\n"),
			source: "purchases",
			scope:  Scope3,
		},
	}

	p := parser.NewUtilityBillParser("test-org", "US-CA")

	scope1 := NewScope1Calculator(Scope1Config{})
	scope3 := NewScope3Calculator(DefaultScope3Config())

	fmt.Printf("\n=== RAW DUMP -> MULTI-SCOPE ===\n")
	for name, exp := range cases {
		result, err := p.Parse(context.Background(), name+".csv", strings.NewReader(exp.csv))
		if err != nil {
			t.Fatalf("%s: engine rejected a valid dump: %v", name, err)
		}
		if len(result.Activities) == 0 {
			t.Fatalf("%s: parsed 0 activities (errors: %v)", name, result.Errors)
		}

		// Every activity must be tagged with the expected source so the right
		// calculator claims it downstream.
		for _, a := range result.Activities {
			if a.Source != exp.source {
				t.Errorf("%s: expected source %q, got %q", name, exp.source, a.Source)
			}
			if a.Unit == "" || a.Quantity <= 0 {
				t.Errorf("%s: activity not fully structured: %+v", name, a)
			}
		}

		// Run the REAL calculator for this scope and assert audit-complete,
		// positive records.
		acts := make([]Activity, 0, len(result.Activities))
		for i := range result.Activities {
			acts = append(acts, &result.Activities[i])
		}

		var records []EmissionRecord
		switch exp.scope {
		case Scope1:
			records, err = scope1.CalculateBatch(context.Background(), acts)
		case Scope3:
			records, err = scope3.CalculateBatch(context.Background(), acts)
		}
		if err != nil {
			t.Fatalf("%s: calculator error: %v", name, err)
		}
		if len(records) != len(result.Activities) {
			t.Fatalf("%s: expected %d records, calculator produced %d (a row was silently dropped)",
				name, len(result.Activities), len(records))
		}

		var totalTonnes float64
		for _, rec := range records {
			if rec.Scope != exp.scope {
				t.Errorf("%s: record landed in %v, expected %v", name, rec.Scope, exp.scope)
			}
			if rec.FactorID == "" || rec.Method == "" || rec.EmissionsTonnesCO2e <= 0 || rec.CalculatedAt.IsZero() {
				t.Errorf("%s: record missing audit fields: %+v", name, rec)
			}
			totalTonnes += rec.EmissionsTonnesCO2e
		}
		fmt.Printf("[%-16s] source=%-22s scope=%d rows=%d -> %.3f tCO2e\n",
			name, exp.source, exp.scope, len(records), totalTonnes)
	}
	fmt.Printf("===============================\n")
}

// TestScope3ExpansionCategories proves waste (cat 5), commuting (cat 7) and
// freight (cat 4) dumps now parse into correctly-scoped activities and produce
// positive Scope 3 emissions through the real calculator.
func TestScope3ExpansionCategories(t *testing.T) {
	cases := map[string]struct {
		csv      string
		source   string
		category string // expected activity category substring
	}{
		"waste": {
			csv: strings.Join([]string{
				"date,treatment,weight_kg",
				"2025-01-15,landfill,1200",
				"2025-02-15,recycling paper,400",
			}, "\n"),
			source: "waste",
		},
		"commuting": {
			csv: strings.Join([]string{
				"employee,commute_mode,miles,days",
				"J. Smith,car,20,220",
				"A. Lee,public transit,12,200",
			}, "\n"),
			source: "commuting",
		},
		"freight": {
			csv: strings.Join([]string{
				"mode,tonnes,distance_km",
				"truck,5,800",
				"rail,20,1500",
			}, "\n"),
			source: "freight",
		},
	}

	p := parser.NewUtilityBillParser("test-org", "US-CA")
	calc := NewScope3Calculator(DefaultScope3Config())

	fmt.Printf("\n=== SCOPE 3 EXPANSION ===\n")
	for name, tc := range cases {
		result, err := p.Parse(context.Background(), name+".csv", strings.NewReader(tc.csv))
		if err != nil {
			t.Fatalf("%s: parse rejected valid dump: %v", name, err)
		}
		if len(result.Activities) != 2 {
			t.Fatalf("%s: expected 2 activities, got %d (errors: %v)", name, len(result.Activities), result.Errors)
		}
		for _, a := range result.Activities {
			if a.Source != tc.source {
				t.Errorf("%s: expected source %q, got %q", name, tc.source, a.Source)
			}
			if a.Quantity <= 0 || a.Unit == "" {
				t.Errorf("%s: activity not structured: %+v", name, a)
			}
		}

		acts := make([]Activity, 0, len(result.Activities))
		for i := range result.Activities {
			acts = append(acts, &result.Activities[i])
		}
		records, err := calc.CalculateBatch(context.Background(), acts)
		if err != nil {
			t.Fatalf("%s: calculator error: %v", name, err)
		}
		if len(records) != 2 {
			t.Fatalf("%s: expected 2 emission records, got %d (a row was dropped)", name, len(records))
		}
		var total float64
		for _, r := range records {
			if r.Scope != Scope3 || r.EmissionsTonnesCO2e <= 0 || r.FactorID == "" {
				t.Errorf("%s: bad record: %+v", name, r)
			}
			total += r.EmissionsTonnesCO2e
		}
		fmt.Printf("[%-10s] source=%-10s rows=%d -> %.3f tCO2e\n", name, tc.source, len(records), total)
	}
	fmt.Printf("=========================\n")
}
