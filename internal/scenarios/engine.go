// Package scenarios provides what-if decarbonization scenario modeling.
//
// This package enables simulation of different emissions reduction strategies
// to support strategic planning and target setting.
package scenarios

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"
)

// =============================================================================
// Core Types
// =============================================================================

// Scenario represents a decarbonization scenario.
type Scenario struct {
	ID            string             `json:"id"`
	TenantID      string             `json:"tenantId"`
	Name          string             `json:"name"`
	Description   string             `json:"description,omitempty"`
	Type          ScenarioType       `json:"type"`
	BaselineYear  int                `json:"baselineYear"`
	TargetYear    int                `json:"targetYear"`
	Baseline      Emissions          `json:"baseline"`
	Interventions []Intervention     `json:"interventions"`
	Projections   []YearlyProjection `json:"projections"`
	Summary       *ScenarioSummary   `json:"summary,omitempty"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
}

// ScenarioType categorizes the scenario.
type ScenarioType string

const (
	TypeBAU       ScenarioType = "business_as_usual"
	Type1_5Degree ScenarioType = "1.5_degree"
	Type2Degree   ScenarioType = "2_degree"
	TypeNetZero   ScenarioType = "net_zero"
	TypeCustom    ScenarioType = "custom"
)

// Emissions represents emissions at a point in time.
type Emissions struct {
	Year   int     `json:"year"`
	Scope1 float64 `json:"scope1"`
	Scope2 float64 `json:"scope2"`
	Scope3 float64 `json:"scope3"`
	Total  float64 `json:"total"`
}

// Intervention represents a decarbonization action.
type Intervention struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	Description     string               `json:"description,omitempty"`
	Category        InterventionCategory `json:"category"`
	Scope           int                  `json:"scope"` // 1, 2, or 3
	Type            InterventionType     `json:"type"`
	StartYear       int                  `json:"startYear"`
	EndYear         int                  `json:"endYear"`
	AnnualReduction float64              `json:"annualReduction"` // tCO2e per year
	TotalReduction  float64              `json:"totalReduction"`  // Total over period
	ReductionRate   float64              `json:"reductionRate"`   // Percentage
	Cost            *Cost                `json:"cost,omitempty"`
	Enabled         bool                 `json:"enabled"`
}

// InterventionCategory groups interventions.
type InterventionCategory string

const (
	CatEnergy      InterventionCategory = "energy"
	CatTransport   InterventionCategory = "transport"
	CatBuildings   InterventionCategory = "buildings"
	CatSupplyChain InterventionCategory = "supply_chain"
	CatProcess     InterventionCategory = "process"
	CatOffset      InterventionCategory = "offset"
)

// InterventionType defines how the intervention applies.
type InterventionType string

const (
	TypeAbsolute   InterventionType = "absolute"   // Fixed reduction
	TypePercentage InterventionType = "percentage" // % reduction
	TypePhaseOut   InterventionType = "phase_out"  // Gradual elimination
)

// Cost tracks intervention costs.
type Cost struct {
	CapEx        float64 `json:"capex"`   // Capital expenditure
	OpEx         float64 `json:"opex"`    // Annual operating cost
	Savings      float64 `json:"savings"` // Annual savings
	PaybackYears float64 `json:"paybackYears"`
	CarbonPrice  float64 `json:"carbonPrice"`   // $/tCO2e
	NPV          float64 `json:"npv,omitempty"` // Net present value
}

// YearlyProjection shows emissions for a year.
type YearlyProjection struct {
	Year       int                `json:"year"`
	Scope1     float64            `json:"scope1"`
	Scope2     float64            `json:"scope2"`
	Scope3     float64            `json:"scope3"`
	Total      float64            `json:"total"`
	Reductions map[string]float64 `json:"reductions"` // By intervention
	Cumulative float64            `json:"cumulative"` // Cumulative reduction
	OnTrack    bool               `json:"onTrack"`
	TargetPath float64            `json:"targetPath"` // Target for this year
}

// ScenarioSummary provides scenario metrics.
type ScenarioSummary struct {
	TotalReduction      float64 `json:"totalReduction"`
	ReductionPercentage float64 `json:"reductionPercentage"`
	Scope1Reduction     float64 `json:"scope1Reduction"`
	Scope2Reduction     float64 `json:"scope2Reduction"`
	Scope3Reduction     float64 `json:"scope3Reduction"`
	TotalCost           float64 `json:"totalCost"`
	TotalSavings        float64 `json:"totalSavings"`
	NetCost             float64 `json:"netCost"`
	AverageCostPerTon   float64 `json:"averageCostPerTon"` // $/tCO2e
	TargetAchieved      bool    `json:"targetAchieved"`
	Gap                 float64 `json:"gap"` // Remaining to target
}

// =============================================================================
// Engine
// =============================================================================

// Engine runs scenario simulations.
type Engine struct {
	scenarios map[string]*Scenario
	templates map[ScenarioType][]Intervention
	logger    *slog.Logger
}

// EngineConfig configures the scenario engine.
type EngineConfig struct {
	Logger *slog.Logger
}

// NewEngine creates a new scenario engine.
func NewEngine(cfg EngineConfig) *Engine {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	e := &Engine{
		scenarios: make(map[string]*Scenario),
		templates: make(map[ScenarioType][]Intervention),
		logger:    cfg.Logger.With("component", "scenario-engine"),
	}

	e.initTemplates()
	return e
}

// initTemplates sets up intervention templates for scenario types.
func (e *Engine) initTemplates() {
	// 1.5°C pathway template - aggressive interventions
	e.templates[Type1_5Degree] = []Intervention{
		{
			ID:            "renewable-energy",
			Name:          "100% Renewable Electricity",
			Category:      CatEnergy,
			Scope:         2,
			Type:          TypePhaseOut,
			ReductionRate: 100,
			Cost: &Cost{
				CapEx:        50000,
				OpEx:         5000,
				Savings:      20000,
				PaybackYears: 3,
			},
			Enabled: true,
		},
		{
			ID:            "fleet-electrification",
			Name:          "Full Fleet Electrification",
			Category:      CatTransport,
			Scope:         1,
			Type:          TypePhaseOut,
			ReductionRate: 100,
			Cost: &Cost{
				CapEx:        200000,
				OpEx:         10000,
				Savings:      50000,
				PaybackYears: 5,
			},
			Enabled: true,
		},
		{
			ID:            "supplier-engagement",
			Name:          "Supplier Decarbonization Program",
			Category:      CatSupplyChain,
			Scope:         3,
			Type:          TypePercentage,
			ReductionRate: 50,
			Enabled:       true,
		},
		{
			ID:            "building-efficiency",
			Name:          "Deep Building Retrofits",
			Category:      CatBuildings,
			Scope:         1,
			Type:          TypePercentage,
			ReductionRate: 40,
			Enabled:       true,
		},
	}

	// Net Zero template
	e.templates[TypeNetZero] = append(e.templates[Type1_5Degree], Intervention{
		ID:       "carbon-removal",
		Name:     "Carbon Removal Offsets",
		Category: CatOffset,
		Scope:    0, // Applies to all
		Type:     TypeAbsolute,
		Cost: &Cost{
			CarbonPrice: 100, // $/tCO2e for removal
		},
		Enabled: true,
	})
}

// CreateScenario creates a new scenario.
func (e *Engine) CreateScenario(tenantID, name string, scenarioType ScenarioType, baseline Emissions, targetYear int) (*Scenario, error) {
	id := fmt.Sprintf("scenario-%d", time.Now().UnixNano())

	scenario := &Scenario{
		ID:           id,
		TenantID:     tenantID,
		Name:         name,
		Type:         scenarioType,
		BaselineYear: baseline.Year,
		TargetYear:   targetYear,
		Baseline:     baseline,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Apply template interventions if available
	if interventions, ok := e.templates[scenarioType]; ok {
		for _, intervention := range interventions {
			intervention.StartYear = baseline.Year + 1
			intervention.EndYear = targetYear
			scenario.Interventions = append(scenario.Interventions, intervention)
		}
	}

	e.scenarios[id] = scenario
	return scenario, nil
}

// AddIntervention adds an intervention to a scenario.
func (e *Engine) AddIntervention(scenarioID string, intervention Intervention) error {
	scenario, ok := e.scenarios[scenarioID]
	if !ok {
		return fmt.Errorf("scenario not found: %s", scenarioID)
	}

	if intervention.ID == "" {
		intervention.ID = fmt.Sprintf("int-%d", time.Now().UnixNano())
	}

	scenario.Interventions = append(scenario.Interventions, intervention)
	scenario.UpdatedAt = time.Now()

	return nil
}

// RunSimulation projects emissions with interventions.
func (e *Engine) RunSimulation(ctx context.Context, scenarioID string) error {
	scenario, ok := e.scenarios[scenarioID]
	if !ok {
		return fmt.Errorf("scenario not found: %s", scenarioID)
	}

	e.logger.Info("running simulation",
		"scenarioId", scenarioID,
		"type", scenario.Type,
		"interventions", len(scenario.Interventions))

	// Calculate target path
	targetReduction := e.getTargetReduction(scenario.Type)
	years := scenario.TargetYear - scenario.BaselineYear

	// Initialize projections
	scenario.Projections = make([]YearlyProjection, 0, years+1)

	// Starting emissions
	scope1 := scenario.Baseline.Scope1
	scope2 := scenario.Baseline.Scope2
	scope3 := scenario.Baseline.Scope3

	var cumulative float64

	for year := scenario.BaselineYear; year <= scenario.TargetYear; year++ {
		projection := YearlyProjection{
			Year:       year,
			Reductions: make(map[string]float64),
		}

		// Calculate target for this year (linear interpolation)
		yearOffset := year - scenario.BaselineYear
		targetProgress := float64(yearOffset) / float64(years)
		targetTotal := scenario.Baseline.Total * (1 - targetReduction*targetProgress)
		projection.TargetPath = targetTotal

		// Apply interventions
		for _, intervention := range scenario.Interventions {
			if !intervention.Enabled {
				continue
			}
			if year < intervention.StartYear || year > intervention.EndYear {
				continue
			}

			reduction := e.calculateReduction(intervention, year, scope1, scope2, scope3)
			projection.Reductions[intervention.ID] = reduction

			switch intervention.Scope {
			case 1:
				scope1 = math.Max(0, scope1-reduction)
			case 2:
				scope2 = math.Max(0, scope2-reduction)
			case 3:
				scope3 = math.Max(0, scope3-reduction)
			case 0: // All scopes
				total := scope1 + scope2 + scope3
				if total > 0 {
					scope1 = math.Max(0, scope1-reduction*scope1/total)
					scope2 = math.Max(0, scope2-reduction*scope2/total)
					scope3 = math.Max(0, scope3-reduction*scope3/total)
				}
			}

			cumulative += reduction
		}

		projection.Scope1 = scope1
		projection.Scope2 = scope2
		projection.Scope3 = scope3
		projection.Total = scope1 + scope2 + scope3
		projection.Cumulative = cumulative
		projection.OnTrack = projection.Total <= projection.TargetPath

		scenario.Projections = append(scenario.Projections, projection)
	}

	// Calculate summary
	scenario.Summary = e.calculateSummary(scenario)
	scenario.UpdatedAt = time.Now()

	return nil
}

// calculateReduction calculates reduction for an intervention in a year.
func (e *Engine) calculateReduction(intervention Intervention, year int, scope1, scope2, scope3 float64) float64 {
	years := intervention.EndYear - intervention.StartYear + 1
	if years <= 0 {
		years = 1
	}
	_ = year - intervention.StartYear // yearIndex - used for phased interventions

	var baseEmissions float64
	switch intervention.Scope {
	case 1:
		baseEmissions = scope1
	case 2:
		baseEmissions = scope2
	case 3:
		baseEmissions = scope3
	case 0:
		baseEmissions = scope1 + scope2 + scope3
	}

	switch intervention.Type {
	case TypeAbsolute:
		return intervention.AnnualReduction

	case TypePercentage:
		// Apply percentage reduction
		return baseEmissions * (intervention.ReductionRate / 100) / float64(years)

	case TypePhaseOut:
		// Linear phase out over the period
		annualPhaseOut := intervention.ReductionRate / float64(years)
		return baseEmissions * (annualPhaseOut / 100)

	default:
		return 0
	}
}

// getTargetReduction returns the target reduction for a scenario type.
func (e *Engine) getTargetReduction(scenarioType ScenarioType) float64 {
	switch scenarioType {
	case Type1_5Degree:
		return 0.50 // 50% by 2030
	case Type2Degree:
		return 0.30 // 30% by 2030
	case TypeNetZero:
		return 0.90 // 90% (with 10% offsets)
	case TypeBAU:
		return 0.0
	default:
		return 0.30
	}
}

// calculateSummary calculates scenario summary metrics.
func (e *Engine) calculateSummary(scenario *Scenario) *ScenarioSummary {
	if len(scenario.Projections) == 0 {
		return nil
	}

	final := scenario.Projections[len(scenario.Projections)-1]
	baseline := scenario.Baseline

	totalReduction := baseline.Total - final.Total
	reductionPct := 0.0
	if baseline.Total > 0 {
		reductionPct = (totalReduction / baseline.Total) * 100
	}

	// Calculate costs
	var totalCost, totalSavings float64
	for _, intervention := range scenario.Interventions {
		if intervention.Cost != nil {
			totalCost += intervention.Cost.CapEx +
				intervention.Cost.OpEx*float64(scenario.TargetYear-scenario.BaselineYear)
			totalSavings += intervention.Cost.Savings * float64(scenario.TargetYear-scenario.BaselineYear)
		}
	}

	avgCost := 0.0
	if totalReduction > 0 {
		avgCost = (totalCost - totalSavings) / totalReduction
	}

	targetReduction := e.getTargetReduction(scenario.Type)
	targetEmissions := baseline.Total * (1 - targetReduction)
	gap := math.Max(0, final.Total-targetEmissions)

	return &ScenarioSummary{
		TotalReduction:      totalReduction,
		ReductionPercentage: reductionPct,
		Scope1Reduction:     baseline.Scope1 - final.Scope1,
		Scope2Reduction:     baseline.Scope2 - final.Scope2,
		Scope3Reduction:     baseline.Scope3 - final.Scope3,
		TotalCost:           totalCost,
		TotalSavings:        totalSavings,
		NetCost:             totalCost - totalSavings,
		AverageCostPerTon:   avgCost,
		TargetAchieved:      final.Total <= targetEmissions,
		Gap:                 gap,
	}
}

// GetScenario retrieves a scenario.
func (e *Engine) GetScenario(scenarioID string) (*Scenario, error) {
	scenario, ok := e.scenarios[scenarioID]
	if !ok {
		return nil, fmt.Errorf("scenario not found: %s", scenarioID)
	}
	return scenario, nil
}

// CompareScenarios compares multiple scenarios.
func (e *Engine) CompareScenarios(scenarioIDs []string) (*ScenarioComparison, error) {
	if len(scenarioIDs) < 2 {
		return nil, fmt.Errorf("need at least 2 scenarios to compare")
	}

	comparison := &ScenarioComparison{
		Scenarios: make([]ScenarioMetrics, len(scenarioIDs)),
	}

	for i, id := range scenarioIDs {
		scenario, ok := e.scenarios[id]
		if !ok {
			return nil, fmt.Errorf("scenario not found: %s", id)
		}

		if scenario.Summary == nil {
			return nil, fmt.Errorf("scenario not simulated: %s", id)
		}

		metrics := ScenarioMetrics{
			ScenarioID:     id,
			Name:           scenario.Name,
			Type:           scenario.Type,
			FinalEmissions: 0,
		}

		if len(scenario.Projections) > 0 {
			final := scenario.Projections[len(scenario.Projections)-1]
			metrics.FinalEmissions = final.Total
		}

		metrics.TotalReduction = scenario.Summary.TotalReduction
		metrics.ReductionPct = scenario.Summary.ReductionPercentage
		metrics.NetCost = scenario.Summary.NetCost
		metrics.CostPerTon = scenario.Summary.AverageCostPerTon
		metrics.TargetAchieved = scenario.Summary.TargetAchieved

		comparison.Scenarios[i] = metrics
	}

	// Find best scenario by reduction
	sort.Slice(comparison.Scenarios, func(i, j int) bool {
		return comparison.Scenarios[i].TotalReduction > comparison.Scenarios[j].TotalReduction
	})

	comparison.BestReduction = comparison.Scenarios[0].ScenarioID

	// Find best by cost-effectiveness
	costEffective := make([]ScenarioMetrics, len(comparison.Scenarios))
	copy(costEffective, comparison.Scenarios)
	sort.Slice(costEffective, func(i, j int) bool {
		return costEffective[i].CostPerTon < costEffective[j].CostPerTon
	})
	comparison.MostCostEffective = costEffective[0].ScenarioID

	return comparison, nil
}

// ScenarioComparison compares multiple scenarios.
type ScenarioComparison struct {
	Scenarios         []ScenarioMetrics `json:"scenarios"`
	BestReduction     string            `json:"bestReduction"`
	MostCostEffective string            `json:"mostCostEffective"`
}

// ScenarioMetrics contains scenario comparison metrics.
type ScenarioMetrics struct {
	ScenarioID     string       `json:"scenarioId"`
	Name           string       `json:"name"`
	Type           ScenarioType `json:"type"`
	FinalEmissions float64      `json:"finalEmissions"`
	TotalReduction float64      `json:"totalReduction"`
	ReductionPct   float64      `json:"reductionPct"`
	NetCost        float64      `json:"netCost"`
	CostPerTon     float64      `json:"costPerTon"`
	TargetAchieved bool         `json:"targetAchieved"`
}

// =============================================================================
// Optimization
// =============================================================================

// OptimizationGoal defines what to optimize for.
type OptimizationGoal string

const (
	GoalMinCost       OptimizationGoal = "min_cost"
	GoalMaxReduction  OptimizationGoal = "max_reduction"
	GoalCostEffective OptimizationGoal = "cost_effective"
)

// OptimizeScenario suggests optimal interventions.
func (e *Engine) OptimizeScenario(ctx context.Context, scenarioID string, goal OptimizationGoal, budget float64) ([]Intervention, error) {
	scenario, ok := e.scenarios[scenarioID]
	if !ok {
		return nil, fmt.Errorf("scenario not found: %s", scenarioID)
	}

	// Get available interventions (disabled ones that could be enabled)
	var candidates []Intervention
	for _, intervention := range scenario.Interventions {
		if !intervention.Enabled && intervention.Cost != nil {
			candidates = append(candidates, intervention)
		}
	}

	switch goal {
	case GoalMinCost:
		// Sort by cost ascending
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Cost.CapEx < candidates[j].Cost.CapEx
		})

	case GoalMaxReduction:
		// Sort by reduction descending
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].ReductionRate > candidates[j].ReductionRate
		})

	case GoalCostEffective:
		// Sort by cost per ton CO2e
		sort.Slice(candidates, func(i, j int) bool {
			costI := candidates[i].Cost.CapEx / math.Max(1, candidates[i].TotalReduction)
			costJ := candidates[j].Cost.CapEx / math.Max(1, candidates[j].TotalReduction)
			return costI < costJ
		})
	}

	// Select interventions within budget
	var selected []Intervention
	remaining := budget
	for _, intervention := range candidates {
		if intervention.Cost.CapEx <= remaining {
			selected = append(selected, intervention)
			remaining -= intervention.Cost.CapEx
		}
	}

	return selected, nil
}

// =============================================================================
// Export
// =============================================================================

// ExportJSON exports a scenario as JSON.
func (e *Engine) ExportJSON(scenarioID string) ([]byte, error) {
	scenario, ok := e.scenarios[scenarioID]
	if !ok {
		return nil, fmt.Errorf("scenario not found: %s", scenarioID)
	}
	return json.MarshalIndent(scenario, "", "  ")
}
