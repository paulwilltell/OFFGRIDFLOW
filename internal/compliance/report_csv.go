package compliance

// report_csv.go exports the inventory's activity-level data as CSV — the
// machine-readable companion to the PDF, for the customer's records and for a
// third-party assurer to re-check every line.

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
)

// ExportInventoryCSV renders the inventory line items as a CSV document.
func ExportInventoryCSV(inv *InventoryReport) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Provenance header rows (commented with #) so the file is self-describing.
	meta := [][]string{
		{"# OffGridFlow GHG Emissions Inventory — activity data export"},
		{fmt.Sprintf("# Organization: %s", nonEmpty(inv.OrgName, inv.OrgID))},
		{fmt.Sprintf("# Reporting year: %d", inv.Year)},
		{fmt.Sprintf("# Generated: %s", inv.GeneratedAt.Format("2006-01-02T15:04:05Z07:00"))},
		{"# Standard: GHG Protocol Corporate Standard (California SB 253)"},
		{fmt.Sprintf("# Totals (tCO2e): scope1=%.4f scope2_location=%.4f scope2_market=%.4f scope3=%.4f total=%.4f",
			inv.Scope1Tonnes, inv.Scope2Tonnes, inv.Scope2MarketTonnes, inv.Scope3Tonnes, inv.TotalTonnes)},
		{fmt.Sprintf("# Scope 1 gases (tCO2e): co2=%.4f ch4=%.4f n2o=%.4f biogenic=%.4f",
			inv.Scope1Gases.CO2, inv.Scope1Gases.CH4, inv.Scope1Gases.N2O, inv.Scope1Gases.Biogenic)},
	}
	if inv.IntensityPerRevenueMM() > 0 || inv.IntensityPerEmployee() > 0 {
		meta = append(meta, []string{fmt.Sprintf("# Intensity: per_$M_revenue=%.4f per_employee=%.4f (revenue=%.0f employees=%d)",
			inv.IntensityPerRevenueMM(), inv.IntensityPerEmployee(), inv.Revenue, inv.Employees)})
	}
	for _, m := range meta {
		if err := w.Write(m); err != nil {
			return nil, fmt.Errorf("csv meta: %w", err)
		}
	}

	header := []string{
		"scope", "source", "category", "location", "period_start", "period_end",
		"quantity", "unit", "emission_factor", "factor_id", "factor_source",
		"method", "data_quality", "emissions_tco2e", "market_emissions_tco2e",
	}
	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("csv header: %w", err)
	}

	for _, li := range inv.LineItems {
		row := []string{
			strconv.Itoa(li.Scope),
			li.Source,
			li.Category,
			li.Location,
			dateOrEmpty(li.PeriodStart),
			dateOrEmpty(li.PeriodEnd),
			strconv.FormatFloat(li.Quantity, 'f', -1, 64),
			li.Unit,
			strconv.FormatFloat(li.EmissionFactor, 'f', -1, 64),
			li.FactorID,
			li.FactorSource,
			li.Method,
			li.DataQuality,
			strconv.FormatFloat(li.EmissionsTonnes, 'f', 6, 64),
			strconv.FormatFloat(li.MarketEmissionsTonnes, 'f', 6, 64),
		}
		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("csv row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("csv flush: %w", err)
	}
	return buf.Bytes(), nil
}

func dateOrEmpty(t interface{ IsZero() bool }) string {
	type dateT interface {
		IsZero() bool
		Format(string) string
	}
	if d, ok := t.(dateT); ok {
		if d.IsZero() {
			return ""
		}
		return d.Format("2006-01-02")
	}
	return ""
}

func nonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
