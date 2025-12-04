package narratives

import (
	"testing"
	"time"
)

func TestReportSection_Values(t *testing.T) {
	sections := []ReportSection{
		SectionE1Overview,
		SectionE1GHGEmissions,
		SectionE1Targets,
		SectionE1TransitionPlan,
		SectionE1ClimateRisks,
		SectionE1EnergyConsumption,
		SectionE2Pollution,
		SectionE3Water,
		SectionE4Biodiversity,
		SectionE5CircularEconomy,
		SectionExecutiveSummary,
		SectionMateriality,
		SectionGovernance,
	}

	for _, s := range sections {
		if string(s) == "" {
			t.Errorf("ReportSection %v has empty string value", s)
		}
	}
}

func TestEmissionsData_Fields(t *testing.T) {
	data := EmissionsData{
		TenantID:        "tenant-1",
		ReportingPeriod: "FY2024",
		Scope1: ScopeData{
			TotalCO2e: 1000.0,
		},
		Scope2: ScopeData{
			TotalCO2e: 2000.0,
		},
		Scope3: ScopeData{
			TotalCO2e: 5000.0,
		},
		Totals: TotalEmissions{
			TotalCO2e: 8000.0,
		},
	}

	if data.TenantID == "" {
		t.Error("TenantID should not be empty")
	}
	total := data.Scope1.TotalCO2e + data.Scope2.TotalCO2e + data.Scope3.TotalCO2e
	if total != data.Totals.TotalCO2e {
		t.Logf("Note: scope totals (%.2f) != total (%.2f)", total, data.Totals.TotalCO2e)
	}
}

func TestNewGenerator(t *testing.T) {
	cfg := GeneratorConfig{}
	gen := NewGenerator(cfg)

	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}
}

func TestTrendData_Fields(t *testing.T) {
	trend := TrendData{
		Period: "2024-Q1",
		CO2e:   1500.0,
		Scope1: 300.0,
		Scope2: 700.0,
		Scope3: 500.0,
	}

	if trend.Period == "" {
		t.Error("Period should not be empty")
	}
	if trend.CO2e <= 0 {
		t.Error("CO2e should be positive")
	}
}

func TestEmissionTarget_Fields(t *testing.T) {
	target := EmissionTarget{
		Name:          "Net Zero 2050",
		TargetYear:    2050,
		BaselineYear:  2020,
		BaselineCO2e:  10000.0,
		TargetCO2e:    0.0,
		Reduction:     100.0,
		Scope:         "all",
		SBTiValidated: true,
		Progress:      15.0,
	}

	if target.Name == "" {
		t.Error("Name should not be empty")
	}
	if target.TargetYear <= target.BaselineYear {
		t.Error("TargetYear should be after BaselineYear")
	}
}

func TestCompanyProfile_Fields(t *testing.T) {
	profile := CompanyProfile{
		Name:      "Acme Corp",
		Industry:  "Technology",
		Employees: 5000,
		Revenue:   500000000.0,
		Locations: 25,
		Countries: []string{"US", "UK", "DE"},
	}

	if profile.Name == "" {
		t.Error("Name should not be empty")
	}
	if profile.Employees <= 0 {
		t.Error("Employees should be positive")
	}
}

func TestGeneratedNarrative_Fields(t *testing.T) {
	narrative := GeneratedNarrative{
		Section:     SectionE1Overview,
		Title:       "Climate Change Overview",
		Content:     "This section describes our climate impact...",
		DataSources: []string{"emissions_db", "facilities"},
		GeneratedAt: time.Now(),
		ModelUsed:   "gpt-4",
		Confidence:  0.95,
		WordCount:   150,
	}

	if narrative.Content == "" {
		t.Error("Content should not be empty")
	}
	if narrative.WordCount <= 0 {
		t.Error("WordCount should be positive")
	}
}

func TestCategoryEmission_Fields(t *testing.T) {
	cat := CategoryEmission{
		Category:   "stationary_combustion",
		CO2e:       500.0,
		Unit:       "tCO2e",
		Percentage: 25.0,
	}

	if cat.Category == "" {
		t.Error("Category should not be empty")
	}
	if cat.Percentage < 0 || cat.Percentage > 100 {
		t.Error("Percentage should be between 0 and 100")
	}
}

func TestScopeData_Fields(t *testing.T) {
	scope := ScopeData{
		TotalCO2e: 1000.0,
		Categories: []CategoryEmission{
			{Category: "fuel", CO2e: 600.0},
			{Category: "fleet", CO2e: 400.0},
		},
		YoYChange:   -5.0,
		Methodology: "GHG Protocol",
	}

	if scope.TotalCO2e <= 0 {
		t.Error("TotalCO2e should be positive")
	}
	if len(scope.Categories) == 0 {
		t.Error("Categories should not be empty")
	}
}
