package parser

import (
	"context"
	"math"
	"strings"
	"testing"
)

func parse(t *testing.T, csv string) *ParseResult {
	t.Helper()
	p := NewUtilityBillParser("test-org", "US-CA")
	result, err := p.Parse(context.Background(), "dump.csv", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return result
}

func approx(a, b float64) bool { return math.Abs(a-b) < 0.01 }

// A fuel dump must classify as Scope 1, tag source fleet/stationary, and
// convert gallons to liters (the calculator's native unit).
func TestParseFuelDump(t *testing.T) {
	result := parse(t, strings.Join([]string{
		"date,vehicle,fuel_type,gallons,cost",
		"2025-01-15,TRUCK-1,diesel,45.2,180.50",
	}, "\n"))

	if len(result.Activities) != 1 {
		t.Fatalf("expected 1 activity, got %d (errors: %v)", len(result.Activities), result.Errors)
	}
	a := result.Activities[0]
	if a.Source != "fleet" {
		t.Errorf("expected source fleet (vehicle column present), got %q", a.Source)
	}
	if a.Category != "diesel" {
		t.Errorf("expected category diesel, got %q", a.Category)
	}
	if a.Unit != "L" {
		t.Errorf("expected unit L, got %q", a.Unit)
	}
	if !approx(a.Quantity, 45.2*3.78541) {
		t.Errorf("expected gallons converted to %.2f L, got %.2f", 45.2*3.78541, a.Quantity)
	}
	if result.Metadata["schema_type"] != "fuel" {
		t.Errorf("expected schema_type fuel, got %q", result.Metadata["schema_type"])
	}
}

// Natural gas billed in therms must infer the natural_gas fuel type and convert
// to cubic meters even without an explicit fuel_type column.
func TestParseNaturalGasTherms(t *testing.T) {
	result := parse(t, strings.Join([]string{
		"meter_id,location,period_start,period_end,therms",
		"GAS-1,US-CA,2025-01-01,2025-01-31,820",
	}, "\n"))

	if len(result.Activities) != 1 {
		t.Fatalf("expected 1 activity, got %d (errors: %v)", len(result.Activities), result.Errors)
	}
	a := result.Activities[0]
	if a.Category != "natural_gas" {
		t.Errorf("expected inferred natural_gas, got %q", a.Category)
	}
	if a.Unit != "m3" {
		t.Errorf("expected unit m3, got %q", a.Unit)
	}
	if !approx(a.Quantity, 820*2.731) {
		t.Errorf("expected therms converted to %.2f m3, got %.2f", 820*2.731, a.Quantity)
	}
}

// A travel dump must classify as Scope 3, normalize the mode, and keep miles as
// the unit (the calculator converts miles->km).
func TestParseTravelDump(t *testing.T) {
	result := parse(t, strings.Join([]string{
		"employee,origin,destination,miles,mode",
		"J. Smith,SFO,JFK,2586,air",
	}, "\n"))

	if len(result.Activities) != 1 {
		t.Fatalf("expected 1 activity, got %d (errors: %v)", len(result.Activities), result.Errors)
	}
	a := result.Activities[0]
	if a.Source != "travel" {
		t.Errorf("expected source travel, got %q", a.Source)
	}
	if a.Category != "flight" {
		t.Errorf("expected mode air normalized to flight, got %q", a.Category)
	}
	if a.Unit != "mile" || a.Quantity != 2586 {
		t.Errorf("expected 2586 mile, got %v %q", a.Quantity, a.Unit)
	}
	if a.Metadata["route"] != "SFO->JFK" {
		t.Errorf("expected route SFO->JFK, got %q", a.Metadata["route"])
	}
}

// A spend dump must classify as Scope 3 purchases, default currency USD, and
// resolve a quarter+year into a period so year-filtered reports include it.
func TestParseSpendDump(t *testing.T) {
	result := parse(t, strings.Join([]string{
		"vendor,category,amount_usd,quarter,year",
		"Acme Steel,raw_materials,45000,Q1,2025",
	}, "\n"))

	if len(result.Activities) != 1 {
		t.Fatalf("expected 1 activity, got %d (errors: %v)", len(result.Activities), result.Errors)
	}
	a := result.Activities[0]
	if a.Source != "purchases" {
		t.Errorf("expected source purchases, got %q", a.Source)
	}
	if a.Unit != "USD" || a.Quantity != 45000 {
		t.Errorf("expected 45000 USD, got %v %q", a.Quantity, a.Unit)
	}
	if a.Metadata["vendor"] != "Acme Steel" {
		t.Errorf("expected vendor Acme Steel, got %q", a.Metadata["vendor"])
	}
	if a.PeriodStart.IsZero() || a.PeriodStart.Month() != 1 || a.PeriodStart.Year() != 2025 {
		t.Errorf("expected Q1 2025 period start, got %v", a.PeriodStart)
	}
}

// An electricity dump with synonym columns must still route through the
// unchanged Scope 2 utility-bill path — no regression from multi-scope routing.
func TestParseElectricityStillRoutesToScope2(t *testing.T) {
	result := parse(t, strings.Join([]string{
		"account_number,grid_region,service_start,service_end,usage,notes",
		"ACC-1001,US-CA,2025-01-01,2025-01-31,36200,HQ",
	}, "\n"))

	if len(result.Activities) != 1 {
		t.Fatalf("expected 1 activity, got %d (errors: %v)", len(result.Activities), result.Errors)
	}
	a := result.Activities[0]
	if a.Source != "utility_bill" || a.Category != "electricity" {
		t.Errorf("electricity misrouted: source=%q category=%q", a.Source, a.Category)
	}
	if a.Unit != "kWh" || a.Quantity != 36200 {
		t.Errorf("expected 36200 kWh, got %v %q", a.Quantity, a.Unit)
	}
}
