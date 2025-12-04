package california

import (
	"context"
	"testing"

	"github.com/example/offgridflow/internal/compliance/core"
)

func TestMapper_BuildReport_Complete(t *testing.T) {
	mapper := &Mapper{}
	input := core.ComplianceInput{
		Year: 2025,
		Data: Input{
			OrgID:             "org-123",
			OrgName:           "Acme Corp",
			Year:              2025,
			Scope1Tons:        10,
			Scope2Tons:        20,
			Scope3Tons:        30,
			AssuranceProvided: true,
			AssuranceLevel:    "limited",
			Methodology:       "GHG Protocol",
		},
	}

	report, err := mapper.BuildReport(context.Background(), input)
	if err != nil {
		t.Fatalf("BuildReport returned error: %v", err)
	}

	if report.Standard == "" {
		t.Fatalf("expected standard to be set")
	}

	reportContent, ok := report.Content["report"].(Report)
	if !ok {
		t.Fatalf("expected report content to be california.Report")
	}

	if reportContent.DisclosureStatus != "ready" {
		t.Fatalf("expected status ready, got %s", reportContent.DisclosureStatus)
	}
	if reportContent.TotalTons != 60 {
		t.Fatalf("expected total 60, got %.2f", reportContent.TotalTons)
	}
}

func TestMapper_BuildReport_Incomplete(t *testing.T) {
	mapper := &Mapper{}
	input := core.ComplianceInput{
		Year: 2025,
		Data: Input{
			OrgID:      "",
			Year:       0,
			Scope1Tons: 0,
			Scope2Tons: 0,
			Scope3Tons: 0,
		},
	}

	report, err := mapper.BuildReport(context.Background(), input)
	if err != nil {
		t.Fatalf("BuildReport returned error: %v", err)
	}

	reportContent, ok := report.Content["report"].(Report)
	if !ok {
		t.Fatalf("expected report content to be california.Report")
	}

	if reportContent.DisclosureStatus != "incomplete" {
		t.Fatalf("expected incomplete status, got %s", reportContent.DisclosureStatus)
	}
	if len(reportContent.Gaps) == 0 {
		t.Fatalf("expected gaps to be populated")
	}
}

func TestValidator(t *testing.T) {
	v := &Validator{}
	report := Report{
		OrgID:             "org-1",
		Year:              2026,
		Scope1Tons:        0,
		Scope2Tons:        5,
		Scope3Tons:        0,
		AssuranceRequired: true,
		AssuranceProvided: false,
		DisclosureStatus:  "partial",
	}

	valid, errs, warns := v.ValidateReport(report)
	if valid {
		t.Fatalf("expected report to be invalid due to missing scopes")
	}
	if len(errs) == 0 {
		t.Fatalf("expected at least one error for missing emissions data")
	}
	if len(warns) == 0 {
		t.Fatalf("expected warnings for missing scopes/assurance")
	}
}
