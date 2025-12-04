// Package csrd provides CSRD/ESRS E1 compliance mapping and report generation.
// It implements the European Sustainability Reporting Standards (ESRS) E1
// Climate Change disclosure requirements as mandated by the Corporate
// Sustainability Reporting Directive (CSRD).
//
// Reference: ESRS E1 Climate Change (Delegated Regulation (EU) 2023/2772)
package csrd

import (
	"context"
	"fmt"
	"time"
)

// =============================================================================
// ESRS E1 Disclosure Requirements
// =============================================================================

// ESRS E1 requires the following disclosures:
// E1-1: Transition plan for climate change mitigation
// E1-2: Policies related to climate change mitigation and adaptation
// E1-3: Actions and resources in relation to climate change policies
// E1-4: Targets related to climate change mitigation and adaptation
// E1-5: Energy consumption and mix
// E1-6: Gross Scopes 1, 2, 3 and Total GHG emissions
// E1-7: GHG removals and GHG mitigation projects financed through carbon credits
// E1-8: Internal carbon pricing
// E1-9: Anticipated financial effects from material physical and transition risks

// =============================================================================
// Input and Output Types
// =============================================================================

// CSRDInput holds the input data for generating a CSRD report.
type CSRDInput struct {
	// Organization identification
	OrgID   string `json:"org_id"`
	OrgName string `json:"org_name"`
	Year    int    `json:"year"`

	// GHG Emissions data (E1-6)
	TotalScope1Tons  float64 `json:"total_scope1_tons"`
	TotalScope2Tons  float64 `json:"total_scope2_tons"`            // Location-based
	Scope2MarketTons float64 `json:"scope2_market_tons,omitempty"` // Market-based
	TotalScope3Tons  float64 `json:"total_scope3_tons"`

	// Scope 3 category breakdown
	Scope3Categories map[string]float64 `json:"scope3_categories,omitempty"`

	// Energy data (E1-5)
	TotalEnergyMWh      float64 `json:"total_energy_mwh"`
	RenewableEnergyMWh  float64 `json:"renewable_energy_mwh"`
	FossilFuelEnergyMWh float64 `json:"fossil_fuel_energy_mwh"`
	NuclearEnergyMWh    float64 `json:"nuclear_energy_mwh"`

	// Fuel consumption breakdown
	FuelConsumption map[string]FuelData `json:"fuel_consumption,omitempty"`

	// Targets (E1-4)
	Targets []ClimateTarget `json:"targets,omitempty"`

	// Carbon credits and removals (E1-7)
	CarbonCredits  float64 `json:"carbon_credits_tons,omitempty"`
	CarbonRemovals float64 `json:"carbon_removals_tons,omitempty"`

	// Internal carbon price (E1-8)
	InternalCarbonPrice *CarbonPrice `json:"internal_carbon_price,omitempty"`

	// Financial effects (E1-9)
	FinancialEffects *FinancialEffects `json:"financial_effects,omitempty"`

	// Transition plan (E1-1)
	TransitionPlan *TransitionPlan `json:"transition_plan,omitempty"`

	// Prior year data for trend analysis
	PriorYearScope1Tons *float64 `json:"prior_year_scope1_tons,omitempty"`
	PriorYearScope2Tons *float64 `json:"prior_year_scope2_tons,omitempty"`
	PriorYearScope3Tons *float64 `json:"prior_year_scope3_tons,omitempty"`

	// Base year for target comparisons
	BaseYear          int     `json:"base_year,omitempty"`
	BaseYearEmissions float64 `json:"base_year_emissions,omitempty"`
}

// FuelData represents fuel consumption data.
type FuelData struct {
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"` // e.g., "liters", "kg", "m3"
	FuelType string  `json:"fuel_type"`
}

// ClimateTarget represents a GHG reduction target.
type ClimateTarget struct {
	ID               string  `json:"id"`
	Description      string  `json:"description"`
	TargetType       string  `json:"target_type"` // "absolute" or "intensity"
	Scope            string  `json:"scope"`       // "scope1", "scope2", "scope3", "all"
	BaseYear         int     `json:"base_year"`
	BaseYearValue    float64 `json:"base_year_value"`
	TargetYear       int     `json:"target_year"`
	TargetValue      float64 `json:"target_value"`
	CurrentValue     float64 `json:"current_value"`
	ReductionPercent float64 `json:"reduction_percent"`
	SBTiValidated    bool    `json:"sbti_validated"`
	NetZeroAligned   bool    `json:"net_zero_aligned"`
}

// CarbonPrice represents internal carbon pricing.
type CarbonPrice struct {
	PricePerTon  float64 `json:"price_per_ton"`
	Currency     string  `json:"currency"`
	PriceType    string  `json:"price_type"` // "shadow", "internal_fee", "implicit"
	CoveredScope string  `json:"covered_scope"`
}

// FinancialEffects captures climate-related financial impacts.
type FinancialEffects struct {
	// Physical risks
	PhysicalRisks []RiskAssessment `json:"physical_risks"`

	// Transition risks
	TransitionRisks []RiskAssessment `json:"transition_risks"`

	// Opportunities
	Opportunities []Opportunity `json:"opportunities"`

	// Quantified amounts
	TotalRiskExposure float64 `json:"total_risk_exposure"`
	Currency          string  `json:"currency"`
}

// RiskAssessment captures a climate-related risk.
type RiskAssessment struct {
	RiskType        string  `json:"risk_type"`
	Description     string  `json:"description"`
	Likelihood      string  `json:"likelihood"`   // "low", "medium", "high", "very_high"
	TimeHorizon     string  `json:"time_horizon"` // "short", "medium", "long"
	FinancialImpact float64 `json:"financial_impact"`
	MitigationPlan  string  `json:"mitigation_plan"`
}

// Opportunity captures a climate-related opportunity.
type Opportunity struct {
	Type            string  `json:"type"`
	Description     string  `json:"description"`
	TimeHorizon     string  `json:"time_horizon"`
	FinancialImpact float64 `json:"financial_impact"`
}

// TransitionPlan captures the organization's climate transition plan.
type TransitionPlan struct {
	HasPlan            bool     `json:"has_plan"`
	ApprovedByBoard    bool     `json:"approved_by_board"`
	AlignedWith15C     bool     `json:"aligned_with_1_5c"`
	NetZeroYear        int      `json:"net_zero_year,omitempty"`
	KeyActions         []string `json:"key_actions,omitempty"`
	InvestmentAmount   float64  `json:"investment_amount,omitempty"`
	InvestmentCurrency string   `json:"investment_currency,omitempty"`
	LockedInEmissions  float64  `json:"locked_in_emissions,omitempty"`
	StrandedAssetsRisk float64  `json:"stranded_assets_risk,omitempty"`
}

// CSRDReport represents a complete CSRD/ESRS E1 compliance report.
type CSRDReport struct {
	// Header
	OrgID       string    `json:"orgId"`
	OrgName     string    `json:"orgName"`
	Year        int       `json:"year"`
	GeneratedAt time.Time `json:"generatedAt"`

	// Metrics organized by ESRS E1 disclosure requirement
	Metrics map[string]interface{} `json:"metrics"`

	// Validation results
	ValidationResults *ValidationResults `json:"validationResults,omitempty"`

	// Completeness score (0-100)
	CompletenessScore float64 `json:"completenessScore"`

	// Required disclosures status
	RequiredDisclosures []DisclosureStatus `json:"requiredDisclosures"`
}

// DisclosureStatus tracks completion of required disclosures.
type DisclosureStatus struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Complete    bool   `json:"complete"`
	DataQuality string `json:"dataQuality"` // "measured", "estimated", "missing"
	Notes       string `json:"notes,omitempty"`
}

// ValidationResults contains report validation outcomes.
type ValidationResults struct {
	Valid    bool                `json:"valid"`
	Errors   []ValidationError   `json:"errors,omitempty"`
	Warnings []ValidationWarning `json:"warnings,omitempty"`
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ValidationWarning represents a validation warning.
type ValidationWarning struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// =============================================================================
// CSRDMapper Interface
// =============================================================================

// CSRDMapper defines the interface for building CSRD reports.
// Implements the core.ComplianceMapper interface.
type CSRDMapper interface {
	BuildReport(ctx context.Context, input CSRDInput) (CSRDReport, error)
	ValidateInput(ctx context.Context, input CSRDInput) ([]ValidationError, []ValidationWarning, error)
	GetRequiredFields() []string
}

// =============================================================================
// Default CSRD Mapper Implementation
// =============================================================================

// DefaultCSRDMapper is the default implementation of CSRDMapper.
type DefaultCSRDMapper struct {
	validator *Validator
}

// NewDefaultCSRDMapper creates a new DefaultCSRDMapper.
func NewDefaultCSRDMapper() *DefaultCSRDMapper {
	return &DefaultCSRDMapper{
		validator: NewValidator(),
	}
}

// BuildReport creates a complete CSRD/ESRS E1 report from the input emissions data.
func (m *DefaultCSRDMapper) BuildReport(ctx context.Context, input CSRDInput) (CSRDReport, error) {
	_ = ctx

	// Calculate total emissions
	totalGHG := input.TotalScope1Tons + input.TotalScope2Tons + input.TotalScope3Tons

	// Build ESRS E1 metrics structure
	metrics := make(map[string]interface{})

	// E1-1: Transition plan for climate change mitigation
	metrics["E1-1_transitionPlan"] = m.buildTransitionPlanDisclosure(input)

	// E1-2: Policies related to climate change
	metrics["E1-2_policies"] = m.buildPoliciesDisclosure(input)

	// E1-3: Actions and resources
	metrics["E1-3_actions"] = m.buildActionsDisclosure(input)

	// E1-4: Targets related to climate change
	metrics["E1-4_targets"] = m.buildTargetsDisclosure(input)

	// E1-5: Energy consumption and mix
	metrics["E1-5_energy"] = m.buildEnergyDisclosure(input)

	// E1-6: Gross Scopes 1, 2, 3 and Total GHG emissions
	metrics["E1-6_ghgEmissions"] = m.buildGHGEmissionsDisclosure(input, totalGHG)

	// E1-7: GHG removals and carbon credits
	metrics["E1-7_removalsAndCredits"] = m.buildRemovalsDisclosure(input)

	// E1-8: Internal carbon pricing
	metrics["E1-8_carbonPricing"] = m.buildCarbonPricingDisclosure(input)

	// E1-9: Anticipated financial effects
	metrics["E1-9_financialEffects"] = m.buildFinancialEffectsDisclosure(input)

	// Track required disclosures
	requiredDisclosures := m.assessRequiredDisclosures(input, metrics)

	// Calculate completeness score
	completenessScore := m.calculateCompletenessScore(requiredDisclosures)

	report := CSRDReport{
		OrgID:               input.OrgID,
		OrgName:             input.OrgName,
		Year:                input.Year,
		Metrics:             metrics,
		GeneratedAt:         time.Now(),
		RequiredDisclosures: requiredDisclosures,
		CompletenessScore:   completenessScore,
	}

	// Validate the report
	if m.validator != nil {
		report.ValidationResults = m.validator.Validate(report)
	}

	return report, nil
}

// buildTransitionPlanDisclosure builds E1-1 disclosure.
func (m *DefaultCSRDMapper) buildTransitionPlanDisclosure(input CSRDInput) map[string]interface{} {
	disclosure := map[string]interface{}{
		"hasTransitionPlan": false,
		"status":            "not_disclosed",
	}

	if input.TransitionPlan != nil {
		disclosure["hasTransitionPlan"] = input.TransitionPlan.HasPlan
		disclosure["approvedByBoard"] = input.TransitionPlan.ApprovedByBoard
		disclosure["alignedWith1_5C"] = input.TransitionPlan.AlignedWith15C

		if input.TransitionPlan.NetZeroYear > 0 {
			disclosure["netZeroTargetYear"] = input.TransitionPlan.NetZeroYear
		}

		if len(input.TransitionPlan.KeyActions) > 0 {
			disclosure["keyDecarbonizationActions"] = input.TransitionPlan.KeyActions
		}

		if input.TransitionPlan.InvestmentAmount > 0 {
			disclosure["decarbonizationInvestment"] = map[string]interface{}{
				"amount":   input.TransitionPlan.InvestmentAmount,
				"currency": input.TransitionPlan.InvestmentCurrency,
			}
		}

		if input.TransitionPlan.LockedInEmissions > 0 {
			disclosure["lockedInGHGEmissions"] = input.TransitionPlan.LockedInEmissions
		}

		disclosure["status"] = "disclosed"
	}

	return disclosure
}

// buildPoliciesDisclosure builds E1-2 disclosure.
func (m *DefaultCSRDMapper) buildPoliciesDisclosure(input CSRDInput) map[string]interface{} {
	return map[string]interface{}{
		"climateChangeMitigationPolicy": map[string]interface{}{
			"adopted": input.TransitionPlan != nil && input.TransitionPlan.HasPlan,
			"scope":   "organization-wide",
		},
		"climateChangeAdaptationPolicy": map[string]interface{}{
			"adopted": input.FinancialEffects != nil && len(input.FinancialEffects.PhysicalRisks) > 0,
			"scope":   "organization-wide",
		},
		"status": "requires_review",
	}
}

// buildActionsDisclosure builds E1-3 disclosure.
func (m *DefaultCSRDMapper) buildActionsDisclosure(input CSRDInput) map[string]interface{} {
	disclosure := map[string]interface{}{
		"keyActions": []interface{}{},
		"status":     "requires_input",
	}

	if input.TransitionPlan != nil && len(input.TransitionPlan.KeyActions) > 0 {
		actions := make([]map[string]interface{}, 0, len(input.TransitionPlan.KeyActions))
		for i, action := range input.TransitionPlan.KeyActions {
			actions = append(actions, map[string]interface{}{
				"id":          fmt.Sprintf("action_%d", i+1),
				"description": action,
				"status":      "in_progress",
			})
		}
		disclosure["keyActions"] = actions
		disclosure["status"] = "disclosed"
	}

	return disclosure
}

// buildTargetsDisclosure builds E1-4 disclosure.
func (m *DefaultCSRDMapper) buildTargetsDisclosure(input CSRDInput) map[string]interface{} {
	disclosure := map[string]interface{}{
		"targets":       []interface{}{},
		"hasTargets":    false,
		"sbtiAligned":   false,
		"netZeroTarget": false,
	}

	if len(input.Targets) > 0 {
		targets := make([]map[string]interface{}, 0, len(input.Targets))
		hasSBTi := false
		hasNetZero := false

		for _, target := range input.Targets {
			progress := calculateTargetProgress(target)

			targets = append(targets, map[string]interface{}{
				"id":               target.ID,
				"description":      target.Description,
				"type":             target.TargetType,
				"scope":            target.Scope,
				"baseYear":         target.BaseYear,
				"baseYearValue":    target.BaseYearValue,
				"targetYear":       target.TargetYear,
				"targetValue":      target.TargetValue,
				"currentValue":     target.CurrentValue,
				"reductionPercent": target.ReductionPercent,
				"progressPercent":  progress,
				"sbtiValidated":    target.SBTiValidated,
				"netZeroAligned":   target.NetZeroAligned,
			})

			if target.SBTiValidated {
				hasSBTi = true
			}
			if target.NetZeroAligned {
				hasNetZero = true
			}
		}

		disclosure["targets"] = targets
		disclosure["hasTargets"] = true
		disclosure["sbtiAligned"] = hasSBTi
		disclosure["netZeroTarget"] = hasNetZero
	}

	return disclosure
}

// buildEnergyDisclosure builds E1-5 disclosure.
func (m *DefaultCSRDMapper) buildEnergyDisclosure(input CSRDInput) map[string]interface{} {
	totalEnergy := input.TotalEnergyMWh
	if totalEnergy == 0 {
		totalEnergy = input.RenewableEnergyMWh + input.FossilFuelEnergyMWh + input.NuclearEnergyMWh
	}

	renewablePercent := 0.0
	if totalEnergy > 0 {
		renewablePercent = (input.RenewableEnergyMWh / totalEnergy) * 100
	}

	disclosure := map[string]interface{}{
		"totalEnergyConsumption": map[string]interface{}{
			"value": totalEnergy,
			"unit":  "MWh",
		},
		"energyMix": map[string]interface{}{
			"renewable": map[string]interface{}{
				"value":   input.RenewableEnergyMWh,
				"unit":    "MWh",
				"percent": renewablePercent,
			},
			"fossilFuel": map[string]interface{}{
				"value": input.FossilFuelEnergyMWh,
				"unit":  "MWh",
			},
			"nuclear": map[string]interface{}{
				"value": input.NuclearEnergyMWh,
				"unit":  "MWh",
			},
		},
		"energyIntensity": map[string]interface{}{
			"status": "requires_calculation",
		},
	}

	// Add fuel consumption breakdown if available
	if len(input.FuelConsumption) > 0 {
		fuelBreakdown := make(map[string]interface{})
		for fuelType, data := range input.FuelConsumption {
			fuelBreakdown[fuelType] = map[string]interface{}{
				"quantity": data.Quantity,
				"unit":     data.Unit,
				"fuelType": data.FuelType,
			}
		}
		disclosure["fuelConsumption"] = fuelBreakdown
	}

	return disclosure
}

// buildGHGEmissionsDisclosure builds E1-6 disclosure.
func (m *DefaultCSRDMapper) buildGHGEmissionsDisclosure(input CSRDInput, totalGHG float64) map[string]interface{} {
	disclosure := map[string]interface{}{
		"scope1": map[string]interface{}{
			"grossEmissions": input.TotalScope1Tons,
			"unit":           "metric tons CO2e",
			"methodology":    "GHG Protocol Corporate Standard",
			"verification":   "requires_verification",
		},
		"scope2": map[string]interface{}{
			"locationBased": map[string]interface{}{
				"grossEmissions": input.TotalScope2Tons,
				"unit":           "metric tons CO2e",
			},
			"marketBased": map[string]interface{}{
				"grossEmissions": input.Scope2MarketTons,
				"unit":           "metric tons CO2e",
			},
			"methodology": "GHG Protocol Scope 2 Guidance",
		},
		"scope3": map[string]interface{}{
			"grossEmissions": input.TotalScope3Tons,
			"unit":           "metric tons CO2e",
			"methodology":    "GHG Protocol Scope 3 Standard",
		},
		"totalGHGEmissions": map[string]interface{}{
			"value": totalGHG,
			"unit":  "metric tons CO2e",
		},
		"reportingYear":         input.Year,
		"consolidationApproach": "operational_control",
	}

	// Add scope 3 category breakdown if available
	if len(input.Scope3Categories) > 0 {
		categories := make(map[string]interface{})
		for cat, value := range input.Scope3Categories {
			categories[cat] = map[string]interface{}{
				"emissions": value,
				"unit":      "metric tons CO2e",
			}
		}
		disclosure["scope3"].(map[string]interface{})["categoryBreakdown"] = categories
	}

	// Add year-over-year comparison if prior year data available
	if input.PriorYearScope1Tons != nil {
		priorTotal := *input.PriorYearScope1Tons
		if input.PriorYearScope2Tons != nil {
			priorTotal += *input.PriorYearScope2Tons
		}
		if input.PriorYearScope3Tons != nil {
			priorTotal += *input.PriorYearScope3Tons
		}

		changePercent := 0.0
		if priorTotal > 0 {
			changePercent = ((totalGHG - priorTotal) / priorTotal) * 100
		}

		disclosure["yearOverYearChange"] = map[string]interface{}{
			"priorYearTotal":   priorTotal,
			"currentYearTotal": totalGHG,
			"changePercent":    changePercent,
			"trend":            getTrend(changePercent),
		}
	}

	// Add emissions intensity metrics
	disclosure["emissionsIntensity"] = map[string]interface{}{
		"status": "requires_revenue_data",
		"note":   "Calculate as tCO2e per million EUR revenue",
	}

	return disclosure
}

// buildRemovalsDisclosure builds E1-7 disclosure.
func (m *DefaultCSRDMapper) buildRemovalsDisclosure(input CSRDInput) map[string]interface{} {
	disclosure := map[string]interface{}{
		"ghgRemovals": map[string]interface{}{
			"value":  input.CarbonRemovals,
			"unit":   "metric tons CO2e",
			"status": "not_applicable",
		},
		"carbonCredits": map[string]interface{}{
			"value":  input.CarbonCredits,
			"unit":   "metric tons CO2e",
			"status": "not_applicable",
		},
		"netEmissions": map[string]interface{}{
			"status": "requires_calculation",
		},
	}

	if input.CarbonRemovals > 0 {
		disclosure["ghgRemovals"].(map[string]interface{})["status"] = "disclosed"
	}

	if input.CarbonCredits > 0 {
		disclosure["carbonCredits"].(map[string]interface{})["status"] = "disclosed"
		disclosure["carbonCredits"].(map[string]interface{})["note"] = "Carbon credits should not be used to achieve GHG reduction targets per ESRS E1"
	}

	return disclosure
}

// buildCarbonPricingDisclosure builds E1-8 disclosure.
func (m *DefaultCSRDMapper) buildCarbonPricingDisclosure(input CSRDInput) map[string]interface{} {
	disclosure := map[string]interface{}{
		"hasInternalCarbonPrice": false,
		"status":                 "not_disclosed",
	}

	if input.InternalCarbonPrice != nil {
		disclosure["hasInternalCarbonPrice"] = true
		disclosure["pricePerTon"] = input.InternalCarbonPrice.PricePerTon
		disclosure["currency"] = input.InternalCarbonPrice.Currency
		disclosure["priceType"] = input.InternalCarbonPrice.PriceType
		disclosure["coveredScope"] = input.InternalCarbonPrice.CoveredScope
		disclosure["status"] = "disclosed"
	}

	return disclosure
}

// buildFinancialEffectsDisclosure builds E1-9 disclosure.
func (m *DefaultCSRDMapper) buildFinancialEffectsDisclosure(input CSRDInput) map[string]interface{} {
	disclosure := map[string]interface{}{
		"physicalRisks":   []interface{}{},
		"transitionRisks": []interface{}{},
		"opportunities":   []interface{}{},
		"status":          "requires_assessment",
	}

	if input.FinancialEffects != nil {
		// Physical risks
		if len(input.FinancialEffects.PhysicalRisks) > 0 {
			risks := make([]map[string]interface{}, 0, len(input.FinancialEffects.PhysicalRisks))
			for _, risk := range input.FinancialEffects.PhysicalRisks {
				risks = append(risks, map[string]interface{}{
					"type":            risk.RiskType,
					"description":     risk.Description,
					"likelihood":      risk.Likelihood,
					"timeHorizon":     risk.TimeHorizon,
					"financialImpact": risk.FinancialImpact,
					"mitigationPlan":  risk.MitigationPlan,
				})
			}
			disclosure["physicalRisks"] = risks
		}

		// Transition risks
		if len(input.FinancialEffects.TransitionRisks) > 0 {
			risks := make([]map[string]interface{}, 0, len(input.FinancialEffects.TransitionRisks))
			for _, risk := range input.FinancialEffects.TransitionRisks {
				risks = append(risks, map[string]interface{}{
					"type":            risk.RiskType,
					"description":     risk.Description,
					"likelihood":      risk.Likelihood,
					"timeHorizon":     risk.TimeHorizon,
					"financialImpact": risk.FinancialImpact,
					"mitigationPlan":  risk.MitigationPlan,
				})
			}
			disclosure["transitionRisks"] = risks
		}

		// Opportunities
		if len(input.FinancialEffects.Opportunities) > 0 {
			opps := make([]map[string]interface{}, 0, len(input.FinancialEffects.Opportunities))
			for _, opp := range input.FinancialEffects.Opportunities {
				opps = append(opps, map[string]interface{}{
					"type":            opp.Type,
					"description":     opp.Description,
					"timeHorizon":     opp.TimeHorizon,
					"financialImpact": opp.FinancialImpact,
				})
			}
			disclosure["opportunities"] = opps
		}

		if input.FinancialEffects.TotalRiskExposure > 0 {
			disclosure["totalRiskExposure"] = map[string]interface{}{
				"value":    input.FinancialEffects.TotalRiskExposure,
				"currency": input.FinancialEffects.Currency,
			}
		}

		disclosure["status"] = "disclosed"
	}

	return disclosure
}

// assessRequiredDisclosures determines which disclosures are required and complete.
func (m *DefaultCSRDMapper) assessRequiredDisclosures(input CSRDInput, metrics map[string]interface{}) []DisclosureStatus {
	disclosures := []DisclosureStatus{
		{
			ID:       "E1-1",
			Name:     "Transition plan for climate change mitigation",
			Required: true,
			Complete: input.TransitionPlan != nil && input.TransitionPlan.HasPlan,
		},
		{
			ID:       "E1-2",
			Name:     "Policies related to climate change",
			Required: true,
			Complete: input.TransitionPlan != nil,
		},
		{
			ID:       "E1-3",
			Name:     "Actions and resources",
			Required: true,
			Complete: input.TransitionPlan != nil && len(input.TransitionPlan.KeyActions) > 0,
		},
		{
			ID:       "E1-4",
			Name:     "Targets related to climate change",
			Required: true,
			Complete: len(input.Targets) > 0,
		},
		{
			ID:          "E1-5",
			Name:        "Energy consumption and mix",
			Required:    true,
			Complete:    input.TotalEnergyMWh > 0 || input.RenewableEnergyMWh > 0,
			DataQuality: getDataQuality(input.TotalEnergyMWh > 0),
		},
		{
			ID:          "E1-6",
			Name:        "Gross Scopes 1, 2, 3 and Total GHG emissions",
			Required:    true,
			Complete:    input.TotalScope1Tons > 0 || input.TotalScope2Tons > 0,
			DataQuality: "measured",
		},
		{
			ID:       "E1-7",
			Name:     "GHG removals and carbon credits",
			Required: false, // Only if applicable
			Complete: input.CarbonCredits > 0 || input.CarbonRemovals > 0,
		},
		{
			ID:       "E1-8",
			Name:     "Internal carbon pricing",
			Required: false, // Only if applicable
			Complete: input.InternalCarbonPrice != nil,
		},
		{
			ID:       "E1-9",
			Name:     "Anticipated financial effects",
			Required: true,
			Complete: input.FinancialEffects != nil,
		},
	}

	return disclosures
}

// calculateCompletenessScore calculates the overall report completeness.
func (m *DefaultCSRDMapper) calculateCompletenessScore(disclosures []DisclosureStatus) float64 {
	requiredCount := 0
	completeCount := 0

	for _, d := range disclosures {
		if d.Required {
			requiredCount++
			if d.Complete {
				completeCount++
			}
		}
	}

	if requiredCount == 0 {
		return 100.0
	}

	return float64(completeCount) / float64(requiredCount) * 100
}

// =============================================================================
// Helper Functions
// =============================================================================

// calculateTargetProgress calculates progress toward a target.
func calculateTargetProgress(target ClimateTarget) float64 {
	if target.BaseYearValue == target.TargetValue {
		return 100.0
	}

	totalReduction := target.BaseYearValue - target.TargetValue
	actualReduction := target.BaseYearValue - target.CurrentValue

	if totalReduction == 0 {
		return 0
	}

	progress := (actualReduction / totalReduction) * 100
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	return progress
}

// getTrend returns a trend indicator based on change percentage.
func getTrend(changePercent float64) string {
	if changePercent < -5 {
		return "decreasing_significantly"
	} else if changePercent < 0 {
		return "decreasing"
	} else if changePercent < 5 {
		return "stable"
	} else if changePercent < 15 {
		return "increasing"
	}
	return "increasing_significantly"
}

// getDataQuality returns data quality assessment.
func getDataQuality(hasData bool) string {
	if hasData {
		return "measured"
	}
	return "missing"
}

// ValidateInput validates the CSRD input data and returns validation errors/warnings.
// Implements the core.ComplianceMapper interface.
func (m *DefaultCSRDMapper) ValidateInput(ctx context.Context, input CSRDInput) ([]ValidationError, []ValidationWarning, error) {
	if m.validator == nil {
		m.validator = NewValidator()
	}
	
	// Run validation and convert to core.ValidationResult
	// The CSRD validator returns CSRDReport which has its own validation structure
	// For now, return empty results - full implementation would map CSRD validation to core types
	return nil, nil, nil
}

// GetRequiredFields returns the list of required fields for CSRD/ESRS E1 reporting.
// Implements the core.ComplianceMapper interface.
func (m *DefaultCSRDMapper) GetRequiredFields() []string {
	return []string{
		"org_id",
		"org_name",
		"year",
		"total_scope1_tons",    // E1-6 required
		"total_scope2_tons",    // E1-6 required
		"total_scope3_tons",    // E1-6 required
		"total_energy_mwh",     // E1-5 required
		"renewable_energy_mwh", // E1-5 required
	}
}

// =============================================================================
// Legacy Mapper (Backward Compatibility)
// =============================================================================

// Mapper translates emissions into ESRS E1 metrics (legacy interface).
type Mapper struct{}

// BuildReport implements the legacy interface.
func (m *Mapper) BuildReport(ctx context.Context, input CSRDInput) (CSRDReport, error) {
	mapper := NewDefaultCSRDMapper()
	return mapper.BuildReport(ctx, input)
}
