package ifrs

import (
	"context"
	"strings"

	"github.com/example/offgridflow/internal/compliance/core"
)

// Input captures a minimal set of data for IFRS S2 climate-related disclosures.
type Input struct {
	OrgID    string
	OrgName  string
	Year     int
	Scope1   float64
	Scope2LB float64 // location-based
	Scope2MB float64 // market-based
	Scope3   float64

	HasTransitionPlan bool
	HasRiskProcess    bool
	HasMetricsTargets bool
	Methodology       string
}

// Report is a simplified IFRS S2 climate disclosure output.
type Report struct {
	OrgID          string   `json:"org_id"`
	OrgName        string   `json:"org_name,omitempty"`
	Year           int      `json:"year"`
	Scope1         float64  `json:"scope1_tons"`
	Scope2Location float64  `json:"scope2_location_tons"`
	Scope2Market   float64  `json:"scope2_market_tons"`
	Scope3         float64  `json:"scope3_tons"`
	Total          float64  `json:"total_tons"`
	Readiness      string   `json:"readiness"` // ready, partial, incomplete
	Warnings       []string `json:"warnings,omitempty"`
	Gaps           []string `json:"gaps,omitempty"`
	Methodology    string   `json:"methodology,omitempty"`
}

// Mapper builds IFRS S2-aligned reports.
type Mapper struct{}

// BuildReport maps generic compliance input to an IFRS S2 report.
func (m *Mapper) BuildReport(ctx context.Context, input core.ComplianceInput) (core.ComplianceReport, error) {
	_ = ctx
	parsed := coerceInput(input)

	report := Report{
		OrgID:          parsed.OrgID,
		OrgName:        parsed.OrgName,
		Year:           parsed.Year,
		Scope1:         parsed.Scope1,
		Scope2Location: parsed.Scope2LB,
		Scope2Market:   parsed.Scope2MB,
		Scope3:         parsed.Scope3,
		Methodology:    parsed.Methodology,
	}
	report.Total = report.Scope1 + report.Scope2Location + report.Scope3

	validator := Validator{}
	valid, errs, warns := validator.ValidateReport(report, parsed)
	report.Warnings = warns
	report.Gaps = errs
	switch {
	case !valid:
		report.Readiness = "incomplete"
	case len(warns) > 0:
		report.Readiness = "partial"
	default:
		report.Readiness = "ready"
	}

	return core.ComplianceReport{
		Standard: "IFRS S2",
		Content: map[string]interface{}{
			"report":   report,
			"warnings": report.Warnings,
			"gaps":     report.Gaps,
		},
	}, nil
}

func coerceInput(in core.ComplianceInput) Input {
	if typed, ok := in.Data.(Input); ok {
		return typed
	}
	if data, ok := in.Data.(map[string]interface{}); ok {
		return Input{
			OrgID:            str(data["org_id"]),
			OrgName:          str(data["org_name"]),
			Year:             intVal(data["year"]),
			Scope1:           floatVal(data["scope1"]),
			Scope2LB:         floatVal(data["scope2_location"]),
			Scope2MB:         floatVal(data["scope2_market"]),
			Scope3:           floatVal(data["scope3"]),
			HasTransitionPlan: boolVal(data["has_transition_plan"]),
			HasRiskProcess:    boolVal(data["has_risk_process"]),
			HasMetricsTargets: boolVal(data["has_metrics_targets"]),
			Methodology:       str(data["methodology"]),
		}
	}
	return Input{OrgID: "", OrgName: "", Year: in.Year}
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
	if s, ok := v.(string); ok {
		return strings.ToLower(s) == "true"
	}
	return false
}
