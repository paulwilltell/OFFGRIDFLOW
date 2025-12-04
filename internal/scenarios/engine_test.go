package scenarios

import (
	"testing"
	"time"
)

func TestScenarioType_Values(t *testing.T) {
	types := []ScenarioType{
		TypeBAU,
		Type1_5Degree,
		Type2Degree,
		TypeNetZero,
		TypeCustom,
	}

	for _, st := range types {
		if string(st) == "" {
			t.Errorf("ScenarioType %v has empty string value", st)
		}
	}
}

func TestInterventionCategory_Values(t *testing.T) {
	categories := []InterventionCategory{
		CatEnergy,
		CatTransport,
		CatBuildings,
		CatSupplyChain,
		CatProcess,
		CatOffset,
	}

	for _, c := range categories {
		if string(c) == "" {
			t.Errorf("InterventionCategory %v has empty string value", c)
		}
	}
}

func TestInterventionType_Values(t *testing.T) {
	types := []InterventionType{
		TypeAbsolute,
		TypePercentage,
		TypePhaseOut,
	}

	for _, it := range types {
		if string(it) == "" {
			t.Errorf("InterventionType %v has empty string value", it)
		}
	}
}

func TestNewEngine(t *testing.T) {
	cfg := EngineConfig{}
	engine := NewEngine(cfg)

	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}
}

func TestEngine_CreateScenario(t *testing.T) {
	cfg := EngineConfig{}
	engine := NewEngine(cfg)

	baseline := Emissions{
		Year:   2024,
		Scope1: 1000.0,
		Scope2: 2000.0,
		Scope3: 5000.0,
		Total:  8000.0,
	}

	scenario, err := engine.CreateScenario("tenant-1", "Net Zero 2050", TypeNetZero, baseline, 2050)
	if err != nil {
		t.Fatalf("CreateScenario failed: %v", err)
	}

	if scenario == nil {
		t.Fatal("CreateScenario returned nil")
	}
	if scenario.ID == "" {
		t.Error("Scenario ID should not be empty")
	}
	if scenario.Name != "Net Zero 2050" {
		t.Errorf("Expected name 'Net Zero 2050', got '%s'", scenario.Name)
	}
}

func TestEngine_AddIntervention(t *testing.T) {
	cfg := EngineConfig{}
	engine := NewEngine(cfg)

	baseline := Emissions{
		Year:   2024,
		Scope1: 1000.0,
		Scope2: 2000.0,
		Scope3: 5000.0,
		Total:  8000.0,
	}

	scenario, _ := engine.CreateScenario("tenant-1", "Test", TypeCustom, baseline, 2030)

	intervention := Intervention{
		Name:            "Solar Installation",
		Category:        CatEnergy,
		Scope:           2,
		Type:            TypeAbsolute,
		StartYear:       2025,
		EndYear:         2030,
		AnnualReduction: 500.0,
		Enabled:         true,
	}

	err := engine.AddIntervention(scenario.ID, intervention)
	if err != nil {
		t.Fatalf("AddIntervention failed: %v", err)
	}
}

func TestEmissions_Fields(t *testing.T) {
	e := Emissions{
		Year:   2024,
		Scope1: 1000.0,
		Scope2: 2000.0,
		Scope3: 5000.0,
		Total:  8000.0,
	}

	sum := e.Scope1 + e.Scope2 + e.Scope3
	if sum != e.Total {
		t.Errorf("Expected sum %.2f to equal total %.2f", sum, e.Total)
	}
}

func TestIntervention_Fields(t *testing.T) {
	i := Intervention{
		ID:              "int-001",
		Name:            "Renewable Energy",
		Category:        CatEnergy,
		Scope:           2,
		Type:            TypePercentage,
		StartYear:       2025,
		EndYear:         2030,
		AnnualReduction: 200.0,
		ReductionRate:   50.0,
		Cost: &Cost{
			CapEx:        100000,
			OpEx:         10000,
			Savings:      30000,
			PaybackYears: 4.0,
		},
		Enabled: true,
	}

	if i.Name == "" {
		t.Error("Name should not be empty")
	}
	if i.EndYear <= i.StartYear {
		t.Error("EndYear should be after StartYear")
	}
}

func TestCost_Fields(t *testing.T) {
	cost := Cost{
		CapEx:        500000,
		OpEx:         50000,
		Savings:      100000,
		PaybackYears: 5.0,
		CarbonPrice:  50.0,
		NPV:          250000,
	}

	if cost.PaybackYears <= 0 {
		t.Error("PaybackYears should be positive")
	}
	netAnnual := cost.Savings - cost.OpEx
	if netAnnual <= 0 {
		t.Log("Net annual should be positive for ROI")
	}
}

func TestYearlyProjection_Fields(t *testing.T) {
	proj := YearlyProjection{
		Year:       2025,
		Scope1:     900.0,
		Scope2:     1800.0,
		Scope3:     4500.0,
		Total:      7200.0,
		Reductions: map[string]float64{"solar": 200, "efficiency": 600},
		Cumulative: 800.0,
		OnTrack:    true,
		TargetPath: 7500.0,
	}

	if proj.Year <= 0 {
		t.Error("Year should be positive")
	}
	if proj.Total < proj.TargetPath && !proj.OnTrack {
		t.Log("Should be on track if below target")
	}
}

func TestScenario_Fields(t *testing.T) {
	now := time.Now()
	scenario := Scenario{
		ID:           "scenario-001",
		TenantID:     "tenant-1",
		Name:         "Net Zero 2050",
		Type:         TypeNetZero,
		BaselineYear: 2024,
		TargetYear:   2050,
		Baseline: Emissions{
			Year:  2024,
			Total: 10000.0,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if scenario.ID == "" {
		t.Error("ID should not be empty")
	}
	if scenario.TargetYear <= scenario.BaselineYear {
		t.Error("TargetYear should be after BaselineYear")
	}
}

func TestScenarioSummary_Fields(t *testing.T) {
	summary := ScenarioSummary{
		TotalReduction:      5000.0,
		ReductionPercentage: 50.0,
		Scope1Reduction:     1000.0,
		Scope2Reduction:     2000.0,
		Scope3Reduction:     2000.0,
		TotalCost:           1000000.0,
		TotalSavings:        500000.0,
		NetCost:             500000.0,
		AverageCostPerTon:   100.0,
		TargetAchieved:      true,
	}

	if summary.ReductionPercentage < 0 || summary.ReductionPercentage > 100 {
		t.Error("ReductionPercentage should be between 0 and 100")
	}
}
