package ifrs

import (
	"context"
	"testing"

	"github.com/example/offgridflow/internal/compliance/core"
)

func TestMapper_BuildReport_Ready(t *testing.T) {
	mapper := &Mapper{}
	input := core.ComplianceInput{
		Year: 2025,
		Data: Input{
			OrgID:            "org-1",
			OrgName:          "IFRS Corp",
			Year:             2025,
			Scope1:           10,
			Scope2LB:         20,
			Scope2MB:         18,
			Scope3:           5,
			HasRiskProcess:   true,
			HasMetricsTargets: true,
			HasTransitionPlan: true,
			Methodology:      "GHG Protocol",
		},
	}

	report, err := mapper.BuildReport(context.Background(), input)
	if err != nil {
		t.Fatalf("BuildReport error: %v", err)
	}

	content, ok := report.Content["report"].(Report)
	if !ok {
		t.Fatalf("expected report content")
	}
	if content.Readiness != "ready" {
		t.Fatalf("expected readiness ready, got %s", content.Readiness)
	}
	if content.Total != 35 {
		t.Fatalf("expected total 35, got %.2f", content.Total)
	}
}

func TestMapper_BuildReport_Incomplete(t *testing.T) {
	mapper := &Mapper{}
	report, err := mapper.BuildReport(context.Background(), core.ComplianceInput{
		Year: 2024,
		Data: Input{},
	})
	if err != nil {
		t.Fatalf("BuildReport error: %v", err)
	}
	content := report.Content["report"].(Report)
	if content.Readiness != "incomplete" {
		t.Fatalf("expected incomplete readiness, got %s", content.Readiness)
	}
	if len(content.Gaps) == 0 {
		t.Fatalf("expected gaps for missing required fields")
	}
}
