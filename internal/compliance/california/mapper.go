package california

import (
	"context"
	"fmt"
	"strings"

	"github.com/example/offgridflow/internal/compliance/core"
)

// Input captures the minimum data required for California climate disclosure (SB 253).
type Input struct {
	OrgID             string
	OrgName           string
	Year              int
	Scope1Tons        float64
	Scope2Tons        float64
	Scope3Tons        float64
	AssuranceProvided bool
	AssuranceLevel    string // limited, reasonable, none
	Methodology       string // e.g., GHG Protocol
}

// Report represents a structured California climate disclosure output.
type Report struct {
	OrgID             string   `json:"org_id"`
	OrgName           string   `json:"org_name,omitempty"`
	Year              int      `json:"year"`
	Scope1Tons        float64  `json:"scope1_tons"`
	Scope2Tons        float64  `json:"scope2_tons"`
	Scope3Tons        float64  `json:"scope3_tons"`
	TotalTons         float64  `json:"total_tons"`
	DisclosureStatus  string   `json:"disclosure_status"` // ready, partial, incomplete
	AssuranceRequired bool     `json:"assurance_required"`
	AssuranceProvided bool     `json:"assurance_provided"`
	AssuranceLevel    string   `json:"assurance_level,omitempty"`
	Methodology       string   `json:"methodology,omitempty"`
	Gaps              []string `json:"gaps,omitempty"`
	Warnings          []string `json:"warnings,omitempty"`
	Summary           string   `json:"summary,omitempty"`
}

// Mapper handles California climate disclosure mapping.
// Implements the core.ComplianceMapper interface.
type Mapper struct{}

// BuildReport maps raw inputs into a structured California climate report.
// It tolerates core.ComplianceInput.Data being either a typed Input or a map[string]any.
func (m *Mapper) BuildReport(ctx context.Context, input core.ComplianceInput) (core.ComplianceReport, error) {
	_ = ctx

	caInput := coerceInput(input)
	report := Report{
		OrgID:             caInput.OrgID,
		OrgName:           caInput.OrgName,
		Year:              chooseYear(caInput.Year, input.Year),
		Scope1Tons:        caInput.Scope1Tons,
		Scope2Tons:        caInput.Scope2Tons,
		Scope3Tons:        caInput.Scope3Tons,
		AssuranceProvided: caInput.AssuranceProvided,
		AssuranceLevel:    strings.ToLower(caInput.AssuranceLevel),
		Methodology:       caInput.Methodology,
	}
	report.TotalTons = report.Scope1Tons + report.Scope2Tons + report.Scope3Tons

	// Assurance required once data is disclosed (simplified rule).
	report.AssuranceRequired = report.TotalTons > 0

	// Identify gaps and warnings.
	if report.OrgID == "" {
		report.Gaps = append(report.Gaps, "org_id is required")
	}
	if report.Year == 0 {
		report.Gaps = append(report.Gaps, "year is required")
	}
	if report.Scope1Tons == 0 {
		report.Warnings = append(report.Warnings, "Scope 1 emissions not provided")
	}
	if report.Scope2Tons == 0 {
		report.Warnings = append(report.Warnings, "Scope 2 emissions not provided")
	}
	if report.Scope3Tons == 0 {
		report.Warnings = append(report.Warnings, "Scope 3 emissions not provided (required by SB 253)")
	}
	if report.AssuranceRequired && !report.AssuranceProvided {
		report.Warnings = append(report.Warnings, "Third-party assurance is required for disclosed totals")
	}

	switch {
	case len(report.Gaps) > 0:
		report.DisclosureStatus = "incomplete"
	case len(report.Warnings) > 0:
		report.DisclosureStatus = "partial"
	default:
		report.DisclosureStatus = "ready"
	}

	report.Summary = buildSummary(report)

	// Convert gaps/warnings to ValidationResults
	validationResults := make([]core.ValidationResult, 0)
	for _, gap := range report.Gaps {
		validationResults = append(validationResults, core.ValidationResult{
			Rule:     "required_field",
			Passed:   false,
			Message:  gap,
			Severity: "error",
			Framework: core.FrameworkCalifornia,
		})
	}
	for _, warning := range report.Warnings {
		validationResults = append(validationResults, core.ValidationResult{
			Rule:     "data_completeness",
			Passed:   false,
			Message:  warning,
			Severity: "warning",
			Framework: core.FrameworkCalifornia,
		})
	}

	return core.ComplianceReport{
		Standard:          "California Climate (SB 253)",
		Framework:         core.FrameworkCalifornia,
		Content: map[string]interface{}{
			"report": report,
		},
	}, nil
}

func coerceInput(in core.ComplianceInput) Input {
	if typed, ok := in.Data.(Input); ok {
		return typed
	}
	if data, ok := in.Data.(map[string]interface{}); ok {
		return Input{
			OrgID:             str(data["org_id"]),
			OrgName:           str(data["org_name"]),
			Year:              intVal(data["year"]),
			Scope1Tons:        floatVal(data["scope1_tons"]),
			Scope2Tons:        floatVal(data["scope2_tons"]),
			Scope3Tons:        floatVal(data["scope3_tons"]),
			AssuranceProvided: boolVal(data["assurance_provided"]),
			AssuranceLevel:    str(data["assurance_level"]),
			Methodology:       str(data["methodology"]),
		}
	}
	// Fallback minimal input
	return Input{
		OrgID:      "",
		OrgName:    "",
		Year:       in.Year,
		Scope1Tons: 0,
		Scope2Tons: 0,
		Scope3Tons: 0,
	}
}

func chooseYear(primary, fallback int) int {
	if primary != 0 {
		return primary
	}
	return fallback
}

func str(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func floatVal(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
}

func intVal(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	default:
		return 0
	}
}

func boolVal(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func buildSummary(r Report) string {
	status := strings.ToUpper(r.DisclosureStatus)
	return fmt.Sprintf("%s: Scope1=%.2f, Scope2=%.2f, Scope3=%.2f, Total=%.2f",
		status, r.Scope1Tons, r.Scope2Tons, r.Scope3Tons, r.TotalTons)
}

// ValidateInput validates California SB 253 input data.
// Implements the core.ComplianceMapper interface.
func (m *Mapper) ValidateInput(ctx context.Context, input core.ComplianceInput) ([]core.ValidationResult, error) {
	caInput := coerceInput(input)
	results := make([]core.ValidationResult, 0)
	
	// Check required fields
	if caInput.OrgID == "" {
		results = append(results, core.ValidationResult{
			Rule:     "org_id_required",
			Passed:   false,
			Message:  "Organization ID is required",
			Severity: "error",
			Framework: core.FrameworkCalifornia,
		})
	}
	
	if caInput.Year == 0 {
		results = append(results, core.ValidationResult{
			Rule:     "year_required",
			Passed:   false,
			Message:  "Reporting year is required",
			Severity: "error",
			Framework: core.FrameworkCalifornia,
		})
	}
	
	// Check scope 3 requirement
	if caInput.Scope3Tons == 0 {
		results = append(results, core.ValidationResult{
			Rule:     "scope3_required",
			Passed:   false,
			Message:  "Scope 3 emissions required by SB 253",
			Severity: "warning",
			Framework: core.FrameworkCalifornia,
		})
	}
	
	// Check assurance requirement
	totalTons := caInput.Scope1Tons + caInput.Scope2Tons + caInput.Scope3Tons
	if totalTons > 0 && !caInput.AssuranceProvided {
		results = append(results, core.ValidationResult{
			Rule:     "assurance_required",
			Passed:   false,
			Message:  "Third-party assurance required for disclosed emissions",
			Severity: "warning",
			Framework: core.FrameworkCalifornia,
		})
	}
	
	return results, nil
}

// GetRequiredFields returns required fields for California SB 253.
// Implements the core.ComplianceMapper interface.
func (m *Mapper) GetRequiredFields() []string {
	return []string{
		"org_id",
		"year",
		"scope1_tons",
		"scope2_tons",
		"scope3_tons",          // Required by SB 253
		"assurance_provided",   // Must be true
		"assurance_level",      // "limited" or "reasonable"
		"methodology",          // e.g., "GHG Protocol"
	}
}
