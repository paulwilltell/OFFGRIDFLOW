package parser

import (
	"strings"
	"testing"
)

// A single mixed file with a type discriminator must be split row-by-row into
// the correct scope sources, while per-file uploads keep working unchanged.
func TestParseMixedDump(t *testing.T) {
	csv := strings.Join([]string{
		"type,quantity,unit,location,date,fuel_type,mode,treatment,vendor",
		"electricity,120000,kwh,US-CA,2025-01-31,,,,",
		"diesel,500,gallons,US-CA,2025-01-15,diesel,,,",
		"flight,2586,miles,,2025-03-01,,air,,",
		"commute,4000,miles,US-CA,2025-01-15,,car,,",
		"waste,1200,kg,US-CA,2025-01-15,,,landfill,",
		"freight,12000,tonne_km,US-CA,2025-01-15,,truck,,",
		"cloud spend,40000,usd,,2025-01-15,,,,AWS",
	}, "\n")

	result := parse(t, csv)

	if result.Metadata["schema_type"] != "mixed" {
		t.Fatalf("expected schema_type mixed, got %q", result.Metadata["schema_type"])
	}
	if len(result.Activities) != 7 {
		t.Fatalf("expected 7 activities, got %d (errors: %v)", len(result.Activities), result.Errors)
	}

	bySource := map[string]int{}
	for _, a := range result.Activities {
		bySource[a.Source]++
	}
	for _, src := range []string{"utility_bill", "stationary_combustion", "travel", "commuting", "waste", "freight", "purchases"} {
		if bySource[src] == 0 {
			t.Errorf("expected an activity with source %q, got sources %v", src, bySource)
		}
	}

	// Spot-check unit conversions were applied per row.
	for _, a := range result.Activities {
		switch a.Source {
		case "stationary_combustion":
			if a.Unit != "L" || !approx(a.Quantity, 500*3.78541) {
				t.Errorf("diesel row not converted to liters: %+v", a)
			}
		case "utility_bill":
			if a.Unit != "kWh" || !approx(a.Quantity, 120000) {
				t.Errorf("electricity row wrong: %+v", a)
			}
		case "waste":
			if a.Unit != "kg" || !approx(a.Quantity, 1200) {
				t.Errorf("waste row wrong: %+v", a)
			}
		}
	}
}

// A mixed row whose type cannot be classified is rejected without sinking the
// rest of the file.
func TestParseMixedDumpSkipsUnknownRow(t *testing.T) {
	csv := strings.Join([]string{
		"activity_type,amount,unit",
		"electricity,1000,kwh",
		"wormholes,5,quatloos",
		"diesel,100,liters",
	}, "\n")
	result := parse(t, csv)
	if len(result.Activities) != 2 {
		t.Fatalf("expected 2 good activities, got %d (errors: %v)", len(result.Activities), result.Errors)
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error for the unknown row, got %d: %v", len(result.Errors), result.Errors)
	}
}

// The standard electricity extended format (quantity,unit,category) must NOT be
// mistaken for a mixed dump.
func TestElectricityExtendedNotMixed(t *testing.T) {
	csv := strings.Join([]string{
		"meter_id,location,period_start,period_end,quantity,unit,category",
		"M1,US-CA,2025-01-01,2025-01-31,120000,kwh,electricity",
	}, "\n")
	result := parse(t, csv)
	if result.Metadata["schema_type"] == "mixed" {
		t.Fatal("standard electricity file was wrongly classified as a mixed dump")
	}
	if len(result.Activities) != 1 {
		t.Fatalf("expected 1 electricity activity, got %d", len(result.Activities))
	}
}
