// Package csrd provides CSRD/ESRS E1 report building and export functionality.
package csrd

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"time"
)

// =============================================================================
// Report Builder
// =============================================================================

// ReportBuilder assembles and serializes CSRD reports.
type ReportBuilder struct {
	Mapper    CSRDMapper
	Validator *Validator
}

// NewReportBuilder creates a new ReportBuilder with the given mapper.
func NewReportBuilder(mapper CSRDMapper) *ReportBuilder {
	return &ReportBuilder{
		Mapper:    mapper,
		Validator: NewValidator(),
	}
}

// NewReportBuilderWithValidator creates a ReportBuilder with custom validator.
func NewReportBuilderWithValidator(mapper CSRDMapper, validator *Validator) *ReportBuilder {
	return &ReportBuilder{
		Mapper:    mapper,
		Validator: validator,
	}
}

// BuildAndSerialize creates a CSRD report and returns it as JSON bytes.
func (b *ReportBuilder) BuildAndSerialize(ctx context.Context, input CSRDInput) ([]byte, error) {
	report, err := b.Mapper.BuildReport(ctx, input)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(report, "", "  ")
}

// Build creates a CSRD report from the input.
func (b *ReportBuilder) Build(ctx context.Context, input CSRDInput) (CSRDReport, error) {
	return b.Mapper.BuildReport(ctx, input)
}

// BuildAndValidate creates and validates a CSRD report.
func (b *ReportBuilder) BuildAndValidate(ctx context.Context, input CSRDInput) (CSRDReport, *ValidationResults, error) {
	report, err := b.Mapper.BuildReport(ctx, input)
	if err != nil {
		return CSRDReport{}, nil, err
	}

	var results *ValidationResults
	if b.Validator != nil {
		results = b.Validator.Validate(report)
	}

	return report, results, nil
}

// =============================================================================
// Export Formats
// =============================================================================

// ExportFormat defines the output format for reports.
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatXML  ExportFormat = "xml"
	FormatXBRL ExportFormat = "xbrl"
	FormatHTML ExportFormat = "html"
	FormatPDF  ExportFormat = "pdf"
)

// ExportOptions configures report export.
type ExportOptions struct {
	Format          ExportFormat
	IncludeMetadata bool
	PrettyPrint     bool
	Language        string // ISO 639-1 code (e.g., "en", "de", "fr")
}

// DefaultExportOptions returns default export options.
func DefaultExportOptions() ExportOptions {
	return ExportOptions{
		Format:          FormatJSON,
		IncludeMetadata: true,
		PrettyPrint:     true,
		Language:        "en",
	}
}

// Export serializes the report to the specified format.
func (b *ReportBuilder) Export(ctx context.Context, input CSRDInput, opts ExportOptions) ([]byte, error) {
	report, err := b.Mapper.BuildReport(ctx, input)
	if err != nil {
		return nil, err
	}

	switch opts.Format {
	case FormatJSON:
		return b.exportJSON(report, opts)
	case FormatXML:
		return b.exportXML(report, opts)
	case FormatXBRL:
		return b.exportXBRL(report, opts)
	case FormatHTML:
		return b.exportHTML(report, opts)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", opts.Format)
	}
}

// exportJSON exports the report as JSON.
func (b *ReportBuilder) exportJSON(report CSRDReport, opts ExportOptions) ([]byte, error) {
	if opts.PrettyPrint {
		return json.MarshalIndent(report, "", "  ")
	}
	return json.Marshal(report)
}

// exportXML exports the report as XML.
func (b *ReportBuilder) exportXML(report CSRDReport, opts ExportOptions) ([]byte, error) {
	wrapper := struct {
		XMLName     xml.Name    `xml:"CSRDReport"`
		Xmlns       string      `xml:"xmlns,attr"`
		OrgID       string      `xml:"orgId"`
		OrgName     string      `xml:"orgName"`
		Year        int         `xml:"year"`
		GeneratedAt string      `xml:"generatedAt"`
		Metrics     interface{} `xml:"metrics"`
	}{
		Xmlns:       "http://www.esrs.eu/xbrl/esrs/e1/2024",
		OrgID:       report.OrgID,
		OrgName:     report.OrgName,
		Year:        report.Year,
		GeneratedAt: report.GeneratedAt.Format(time.RFC3339),
		Metrics:     report.Metrics,
	}

	if opts.PrettyPrint {
		return xml.MarshalIndent(wrapper, "", "  ")
	}
	return xml.Marshal(wrapper)
}

// exportXBRL exports the report as XBRL (iXBRL format).
func (b *ReportBuilder) exportXBRL(report CSRDReport, opts ExportOptions) ([]byte, error) {
	// Build XBRL taxonomy references for ESRS E1
	xbrl := XBRLReport{
		SchemaRef:   "http://www.esrs.eu/xbrl/taxonomy/esrs-e1-2024.xsd",
		ContextRef:  fmt.Sprintf("ctx_%d", report.Year),
		EntityID:    report.OrgID,
		PeriodStart: fmt.Sprintf("%d-01-01", report.Year),
		PeriodEnd:   fmt.Sprintf("%d-12-31", report.Year),
		Facts:       b.buildXBRLFacts(report),
		GeneratedAt: report.GeneratedAt,
	}

	return json.MarshalIndent(xbrl, "", "  ")
}

// XBRLReport represents an XBRL report structure.
type XBRLReport struct {
	SchemaRef   string     `json:"schemaRef"`
	ContextRef  string     `json:"contextRef"`
	EntityID    string     `json:"entityId"`
	PeriodStart string     `json:"periodStart"`
	PeriodEnd   string     `json:"periodEnd"`
	Facts       []XBRLFact `json:"facts"`
	GeneratedAt time.Time  `json:"generatedAt"`
}

// XBRLFact represents a single XBRL fact.
type XBRLFact struct {
	Concept    string      `json:"concept"`
	Value      interface{} `json:"value"`
	Unit       string      `json:"unit,omitempty"`
	Decimals   int         `json:"decimals,omitempty"`
	ContextRef string      `json:"contextRef"`
}

// buildXBRLFacts converts metrics to XBRL facts.
func (b *ReportBuilder) buildXBRLFacts(report CSRDReport) []XBRLFact {
	facts := make([]XBRLFact, 0)
	contextRef := fmt.Sprintf("ctx_%d", report.Year)

	// E1-6 GHG Emissions facts
	if ghgEmissions, ok := report.Metrics["E1-6_ghgEmissions"].(map[string]interface{}); ok {
		// Scope 1
		if scope1, ok := ghgEmissions["scope1"].(map[string]interface{}); ok {
			if emissions, ok := scope1["grossEmissions"].(float64); ok {
				facts = append(facts, XBRLFact{
					Concept:    "esrs-e1:GrossScope1GHGEmissions",
					Value:      emissions,
					Unit:       "esrs-utr:tCO2e",
					Decimals:   0,
					ContextRef: contextRef,
				})
			}
		}

		// Scope 2 Location-based
		if scope2, ok := ghgEmissions["scope2"].(map[string]interface{}); ok {
			if lb, ok := scope2["locationBased"].(map[string]interface{}); ok {
				if emissions, ok := lb["grossEmissions"].(float64); ok {
					facts = append(facts, XBRLFact{
						Concept:    "esrs-e1:GrossScope2LocationBasedGHGEmissions",
						Value:      emissions,
						Unit:       "esrs-utr:tCO2e",
						Decimals:   0,
						ContextRef: contextRef,
					})
				}
			}
			// Market-based
			if mb, ok := scope2["marketBased"].(map[string]interface{}); ok {
				if emissions, ok := mb["grossEmissions"].(float64); ok {
					facts = append(facts, XBRLFact{
						Concept:    "esrs-e1:GrossScope2MarketBasedGHGEmissions",
						Value:      emissions,
						Unit:       "esrs-utr:tCO2e",
						Decimals:   0,
						ContextRef: contextRef,
					})
				}
			}
		}

		// Scope 3
		if scope3, ok := ghgEmissions["scope3"].(map[string]interface{}); ok {
			if emissions, ok := scope3["grossEmissions"].(float64); ok {
				facts = append(facts, XBRLFact{
					Concept:    "esrs-e1:GrossScope3GHGEmissions",
					Value:      emissions,
					Unit:       "esrs-utr:tCO2e",
					Decimals:   0,
					ContextRef: contextRef,
				})
			}
		}

		// Total GHG
		if total, ok := ghgEmissions["totalGHGEmissions"].(map[string]interface{}); ok {
			if value, ok := total["value"].(float64); ok {
				facts = append(facts, XBRLFact{
					Concept:    "esrs-e1:TotalGHGEmissions",
					Value:      value,
					Unit:       "esrs-utr:tCO2e",
					Decimals:   0,
					ContextRef: contextRef,
				})
			}
		}
	}

	// E1-5 Energy facts
	if energy, ok := report.Metrics["E1-5_energy"].(map[string]interface{}); ok {
		if totalEnergy, ok := energy["totalEnergyConsumption"].(map[string]interface{}); ok {
			if value, ok := totalEnergy["value"].(float64); ok {
				facts = append(facts, XBRLFact{
					Concept:    "esrs-e1:TotalEnergyConsumption",
					Value:      value,
					Unit:       "esrs-utr:MWh",
					Decimals:   0,
					ContextRef: contextRef,
				})
			}
		}
	}

	// E1-4 Targets
	if targets, ok := report.Metrics["E1-4_targets"].(map[string]interface{}); ok {
		if hasTargets, ok := targets["hasTargets"].(bool); ok {
			facts = append(facts, XBRLFact{
				Concept:    "esrs-e1:HasGHGReductionTargets",
				Value:      hasTargets,
				ContextRef: contextRef,
			})
		}
		if sbti, ok := targets["sbtiAligned"].(bool); ok {
			facts = append(facts, XBRLFact{
				Concept:    "esrs-e1:TargetsValidatedBySBTi",
				Value:      sbti,
				ContextRef: contextRef,
			})
		}
	}

	// E1-1 Transition Plan
	if tp, ok := report.Metrics["E1-1_transitionPlan"].(map[string]interface{}); ok {
		if hasPlan, ok := tp["hasTransitionPlan"].(bool); ok {
			facts = append(facts, XBRLFact{
				Concept:    "esrs-e1:HasClimateTransitionPlan",
				Value:      hasPlan,
				ContextRef: contextRef,
			})
		}
	}

	return facts
}

// exportHTML exports the report as an HTML document.
func (b *ReportBuilder) exportHTML(report CSRDReport, opts ExportOptions) ([]byte, error) {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CSRD Report - %s - %d</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 40px; color: #1a1a2e; }
        h1 { color: #16213e; border-bottom: 2px solid #0f3460; padding-bottom: 10px; }
        h2 { color: #0f3460; margin-top: 30px; }
        .metric { background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 8px; border-left: 4px solid #0f3460; }
        .metric-label { font-weight: 600; color: #16213e; }
        .metric-value { font-size: 1.2em; color: #0f3460; margin-top: 5px; }
        .disclosure { margin: 20px 0; }
        .status-complete { color: #28a745; }
        .status-incomplete { color: #ffc107; }
        .status-missing { color: #dc3545; }
        table { width: 100%%; border-collapse: collapse; margin: 15px 0; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #dee2e6; }
        th { background: #f8f9fa; font-weight: 600; }
        .footer { margin-top: 40px; padding-top: 20px; border-top: 1px solid #dee2e6; font-size: 0.9em; color: #666; }
    </style>
</head>
<body>
    <h1>CSRD/ESRS E1 Climate Disclosure Report</h1>
    <p><strong>Organization:</strong> %s (%s)</p>
    <p><strong>Reporting Year:</strong> %d</p>
    <p><strong>Generated:</strong> %s</p>
    <p><strong>Completeness Score:</strong> %.0f%%</p>

    <h2>E1-6: GHG Emissions</h2>
    %s

    <h2>E1-5: Energy Consumption</h2>
    %s

    <h2>E1-4: Climate Targets</h2>
    %s

    <h2>E1-1: Transition Plan</h2>
    %s

    <h2>Disclosure Status</h2>
    %s

    <div class="footer">
        <p>This report was generated in accordance with ESRS E1 Climate Change disclosure requirements under the Corporate Sustainability Reporting Directive (CSRD).</p>
    </div>
</body>
</html>`,
		opts.Language,
		report.OrgName,
		report.Year,
		report.OrgName,
		report.OrgID,
		report.Year,
		report.GeneratedAt.Format("January 2, 2006 at 15:04 MST"),
		report.CompletenessScore,
		b.formatGHGEmissionsHTML(report),
		b.formatEnergyHTML(report),
		b.formatTargetsHTML(report),
		b.formatTransitionPlanHTML(report),
		b.formatDisclosureStatusHTML(report),
	)

	return []byte(html), nil
}

// formatGHGEmissionsHTML formats E1-6 for HTML.
func (b *ReportBuilder) formatGHGEmissionsHTML(report CSRDReport) string {
	ghg, ok := report.Metrics["E1-6_ghgEmissions"].(map[string]interface{})
	if !ok {
		return "<p>GHG emissions data not available.</p>"
	}

	html := `<table>
        <tr><th>Scope</th><th>Emissions (tCO2e)</th></tr>`

	if scope1, ok := ghg["scope1"].(map[string]interface{}); ok {
		if emissions, ok := scope1["grossEmissions"].(float64); ok {
			html += fmt.Sprintf("<tr><td>Scope 1 (Direct)</td><td>%.2f</td></tr>", emissions)
		}
	}

	if scope2, ok := ghg["scope2"].(map[string]interface{}); ok {
		if lb, ok := scope2["locationBased"].(map[string]interface{}); ok {
			if emissions, ok := lb["grossEmissions"].(float64); ok {
				html += fmt.Sprintf("<tr><td>Scope 2 (Location-based)</td><td>%.2f</td></tr>", emissions)
			}
		}
		if mb, ok := scope2["marketBased"].(map[string]interface{}); ok {
			if emissions, ok := mb["grossEmissions"].(float64); ok {
				html += fmt.Sprintf("<tr><td>Scope 2 (Market-based)</td><td>%.2f</td></tr>", emissions)
			}
		}
	}

	if scope3, ok := ghg["scope3"].(map[string]interface{}); ok {
		if emissions, ok := scope3["grossEmissions"].(float64); ok {
			html += fmt.Sprintf("<tr><td>Scope 3 (Value Chain)</td><td>%.2f</td></tr>", emissions)
		}
	}

	if total, ok := ghg["totalGHGEmissions"].(map[string]interface{}); ok {
		if value, ok := total["value"].(float64); ok {
			html += fmt.Sprintf("<tr style=\"font-weight:bold;\"><td>Total GHG Emissions</td><td>%.2f</td></tr>", value)
		}
	}

	html += "</table>"
	return html
}

// formatEnergyHTML formats E1-5 for HTML.
func (b *ReportBuilder) formatEnergyHTML(report CSRDReport) string {
	energy, ok := report.Metrics["E1-5_energy"].(map[string]interface{})
	if !ok {
		return "<p>Energy consumption data not available.</p>"
	}

	html := ""
	if total, ok := energy["totalEnergyConsumption"].(map[string]interface{}); ok {
		if value, ok := total["value"].(float64); ok {
			html += fmt.Sprintf(`<div class="metric">
                <div class="metric-label">Total Energy Consumption</div>
                <div class="metric-value">%.2f MWh</div>
            </div>`, value)
		}
	}

	if mix, ok := energy["energyMix"].(map[string]interface{}); ok {
		if renewable, ok := mix["renewable"].(map[string]interface{}); ok {
			if pct, ok := renewable["percent"].(float64); ok {
				html += fmt.Sprintf(`<div class="metric">
                    <div class="metric-label">Renewable Energy Share</div>
                    <div class="metric-value">%.1f%%</div>
                </div>`, pct)
			}
		}
	}

	return html
}

// formatTargetsHTML formats E1-4 for HTML.
func (b *ReportBuilder) formatTargetsHTML(report CSRDReport) string {
	targets, ok := report.Metrics["E1-4_targets"].(map[string]interface{})
	if !ok {
		return "<p>No climate targets defined.</p>"
	}

	hasTargets, _ := targets["hasTargets"].(bool)
	if !hasTargets {
		return `<p class="status-incomplete">No climate targets have been set. Consider adopting Science Based Targets.</p>`
	}

	html := ""
	if sbti, ok := targets["sbtiAligned"].(bool); ok && sbti {
		html += `<p class="status-complete">✓ Targets validated by Science Based Targets initiative (SBTi)</p>`
	}

	return html
}

// formatTransitionPlanHTML formats E1-1 for HTML.
func (b *ReportBuilder) formatTransitionPlanHTML(report CSRDReport) string {
	tp, ok := report.Metrics["E1-1_transitionPlan"].(map[string]interface{})
	if !ok {
		return `<p class="status-missing">Transition plan not disclosed.</p>`
	}

	hasPlan, _ := tp["hasTransitionPlan"].(bool)
	if !hasPlan {
		return `<p class="status-incomplete">No climate transition plan adopted.</p>`
	}

	html := `<p class="status-complete">✓ Climate transition plan adopted</p>`

	if approved, ok := tp["approvedByBoard"].(bool); ok && approved {
		html += `<p class="status-complete">✓ Approved by board of directors</p>`
	}

	if aligned, ok := tp["alignedWith1_5C"].(bool); ok && aligned {
		html += `<p class="status-complete">✓ Aligned with 1.5°C Paris Agreement target</p>`
	}

	if netZero, ok := tp["netZeroTargetYear"].(int); ok {
		html += fmt.Sprintf(`<p>Net zero target year: %d</p>`, netZero)
	}

	return html
}

// formatDisclosureStatusHTML formats disclosure status table.
func (b *ReportBuilder) formatDisclosureStatusHTML(report CSRDReport) string {
	if len(report.RequiredDisclosures) == 0 {
		return "<p>Disclosure status not available.</p>"
	}

	html := `<table>
        <tr><th>Disclosure</th><th>Status</th><th>Notes</th></tr>`

	for _, d := range report.RequiredDisclosures {
		statusClass := "status-missing"
		statusText := "Missing"
		if d.Complete {
			statusClass = "status-complete"
			statusText = "Complete"
		} else if !d.Required {
			statusClass = ""
			statusText = "Optional"
		}

		html += fmt.Sprintf(`<tr>
            <td>%s: %s</td>
            <td class="%s">%s</td>
            <td>%s</td>
        </tr>`, d.ID, d.Name, statusClass, statusText, d.Notes)
	}

	html += "</table>"
	return html
}
