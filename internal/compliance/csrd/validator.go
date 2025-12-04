// Package csrd provides CSRD/ESRS E1 compliance validation.
package csrd

import (
	"fmt"
	"strings"
)

// =============================================================================
// Validator Implementation
// =============================================================================

// Validator checks CSRD report completeness and compliance with ESRS E1.
type Validator struct {
	// Validation rules configuration
	RequireScope3Categories bool
	RequireTransitionPlan   bool
	RequireTargets          bool
	StrictMode              bool
}

// NewValidator creates a new CSRD validator with default settings.
func NewValidator() *Validator {
	return &Validator{
		RequireScope3Categories: true,
		RequireTransitionPlan:   true,
		RequireTargets:          true,
		StrictMode:              false,
	}
}

// Validate performs comprehensive validation on a CSRD report.
func (v *Validator) Validate(report CSRDReport) *ValidationResults {
	results := &ValidationResults{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	// Validate required fields
	v.validateRequiredFields(report, results)

	// Validate GHG emissions (E1-6)
	v.validateGHGEmissions(report, results)

	// Validate targets (E1-4)
	v.validateTargets(report, results)

	// Validate transition plan (E1-1)
	v.validateTransitionPlan(report, results)

	// Validate energy data (E1-5)
	v.validateEnergy(report, results)

	// Validate financial effects (E1-9)
	v.validateFinancialEffects(report, results)

	// Set overall validity
	results.Valid = len(results.Errors) == 0

	return results
}

// validateRequiredFields checks that all required fields are present.
func (v *Validator) validateRequiredFields(report CSRDReport, results *ValidationResults) {
	if report.OrgID == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "orgId",
			Code:    "REQUIRED_FIELD",
			Message: "Organization ID is required",
		})
	}

	if report.Year <= 0 {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "year",
			Code:    "INVALID_VALUE",
			Message: "Reporting year must be a positive integer",
		})
	}

	if report.Metrics == nil {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "metrics",
			Code:    "REQUIRED_FIELD",
			Message: "Metrics are required",
		})
	}
}

// validateGHGEmissions validates E1-6 GHG emissions disclosure.
func (v *Validator) validateGHGEmissions(report CSRDReport, results *ValidationResults) {
	ghgMetrics, ok := report.Metrics["E1-6_ghgEmissions"].(map[string]interface{})
	if !ok {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "E1-6_ghgEmissions",
			Code:    "MISSING_DISCLOSURE",
			Message: "E1-6 GHG emissions disclosure is required",
		})
		return
	}

	// Validate Scope 1
	scope1, ok := ghgMetrics["scope1"].(map[string]interface{})
	if !ok {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "E1-6_ghgEmissions.scope1",
			Code:    "MISSING_FIELD",
			Message: "Scope 1 emissions are required",
		})
	} else {
		if emissions, ok := scope1["grossEmissions"].(float64); ok && emissions < 0 {
			results.Errors = append(results.Errors, ValidationError{
				Field:   "E1-6_ghgEmissions.scope1.grossEmissions",
				Code:    "INVALID_VALUE",
				Message: "Scope 1 emissions cannot be negative",
			})
		}
	}

	// Validate Scope 2
	scope2, ok := ghgMetrics["scope2"].(map[string]interface{})
	if !ok {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "E1-6_ghgEmissions.scope2",
			Code:    "MISSING_FIELD",
			Message: "Scope 2 emissions are required",
		})
	} else {
		// Check for both location-based and market-based
		if _, hasLocation := scope2["locationBased"]; !hasLocation {
			results.Warnings = append(results.Warnings, ValidationWarning{
				Field:   "E1-6_ghgEmissions.scope2.locationBased",
				Code:    "RECOMMENDED_FIELD",
				Message: "Location-based Scope 2 emissions are recommended",
			})
		}
		if _, hasMarket := scope2["marketBased"]; !hasMarket {
			results.Warnings = append(results.Warnings, ValidationWarning{
				Field:   "E1-6_ghgEmissions.scope2.marketBased",
				Code:    "RECOMMENDED_FIELD",
				Message: "Market-based Scope 2 emissions are recommended",
			})
		}
	}

	// Validate Scope 3
	scope3, ok := ghgMetrics["scope3"].(map[string]interface{})
	if !ok {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "E1-6_ghgEmissions.scope3",
			Code:    "RECOMMENDED_FIELD",
			Message: "Scope 3 emissions are recommended for comprehensive reporting",
		})
	} else if v.RequireScope3Categories {
		if _, hasCategories := scope3["categoryBreakdown"]; !hasCategories {
			results.Warnings = append(results.Warnings, ValidationWarning{
				Field:   "E1-6_ghgEmissions.scope3.categoryBreakdown",
				Code:    "RECOMMENDED_BREAKDOWN",
				Message: "Scope 3 category breakdown is recommended per GHG Protocol",
			})
		}
	}

	// Check for consolidation approach
	if _, hasApproach := ghgMetrics["consolidationApproach"]; !hasApproach {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "E1-6_ghgEmissions.consolidationApproach",
			Code:    "RECOMMENDED_FIELD",
			Message: "Consolidation approach should be disclosed (operational_control or financial_control)",
		})
	}
}

// validateTargets validates E1-4 climate targets disclosure.
func (v *Validator) validateTargets(report CSRDReport, results *ValidationResults) {
	targetsMetrics, ok := report.Metrics["E1-4_targets"].(map[string]interface{})
	if !ok {
		if v.RequireTargets {
			results.Errors = append(results.Errors, ValidationError{
				Field:   "E1-4_targets",
				Code:    "MISSING_DISCLOSURE",
				Message: "E1-4 climate targets disclosure is required",
			})
		}
		return
	}

	hasTargets, _ := targetsMetrics["hasTargets"].(bool)
	if !hasTargets {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "E1-4_targets",
			Code:    "NO_TARGETS",
			Message: "No climate targets defined - consider setting science-based targets",
		})
		return
	}

	// Validate individual targets
	targets, ok := targetsMetrics["targets"].([]map[string]interface{})
	if ok {
		for i, target := range targets {
			v.validateTarget(target, i, results)
		}
	}

	// Check for SBTi alignment
	sbtiAligned, _ := targetsMetrics["sbtiAligned"].(bool)
	if !sbtiAligned {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "E1-4_targets.sbtiAligned",
			Code:    "SBTI_RECOMMENDED",
			Message: "Science Based Targets initiative (SBTi) validation is recommended",
		})
	}
}

// validateTarget validates a single climate target.
func (v *Validator) validateTarget(target map[string]interface{}, index int, results *ValidationResults) {
	prefix := fmt.Sprintf("E1-4_targets.targets[%d]", index)

	// Check base year
	baseYear, ok := target["baseYear"].(int)
	if !ok || baseYear <= 0 {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".baseYear",
			Code:    "INVALID_VALUE",
			Message: "Target base year is required and must be valid",
		})
	}

	// Check target year
	targetYear, ok := target["targetYear"].(int)
	if !ok || targetYear <= 0 {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".targetYear",
			Code:    "INVALID_VALUE",
			Message: "Target year is required and must be valid",
		})
	}

	// Check that target year is after base year
	if baseYear > 0 && targetYear > 0 && targetYear <= baseYear {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".targetYear",
			Code:    "INVALID_PERIOD",
			Message: "Target year must be after base year",
		})
	}

	// Check target type
	targetType, _ := target["type"].(string)
	validTypes := []string{"absolute", "intensity"}
	if !contains(validTypes, targetType) {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   prefix + ".type",
			Code:    "INVALID_TYPE",
			Message: "Target type should be 'absolute' or 'intensity'",
		})
	}

	// Check scope
	scope, _ := target["scope"].(string)
	validScopes := []string{"scope1", "scope2", "scope3", "all", "scope1+2", "scope1+2+3"}
	if !contains(validScopes, scope) {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   prefix + ".scope",
			Code:    "INVALID_SCOPE",
			Message: "Target scope should specify which emissions scopes are covered",
		})
	}
}

// validateTransitionPlan validates E1-1 transition plan disclosure.
func (v *Validator) validateTransitionPlan(report CSRDReport, results *ValidationResults) {
	tpMetrics, ok := report.Metrics["E1-1_transitionPlan"].(map[string]interface{})
	if !ok {
		if v.RequireTransitionPlan {
			results.Errors = append(results.Errors, ValidationError{
				Field:   "E1-1_transitionPlan",
				Code:    "MISSING_DISCLOSURE",
				Message: "E1-1 transition plan disclosure is required",
			})
		}
		return
	}

	hasPlan, _ := tpMetrics["hasTransitionPlan"].(bool)
	if !hasPlan {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "E1-1_transitionPlan",
			Code:    "NO_TRANSITION_PLAN",
			Message: "No climate transition plan - required for CSRD compliance",
		})
		return
	}

	// Check for board approval
	if approved, _ := tpMetrics["approvedByBoard"].(bool); !approved {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "E1-1_transitionPlan.approvedByBoard",
			Code:    "BOARD_APPROVAL",
			Message: "Transition plan should be approved by the board",
		})
	}

	// Check for 1.5°C alignment
	if aligned, _ := tpMetrics["alignedWith1_5C"].(bool); !aligned {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "E1-1_transitionPlan.alignedWith1_5C",
			Code:    "PARIS_ALIGNMENT",
			Message: "Transition plan should be aligned with 1.5°C Paris Agreement target",
		})
	}

	// Check for net zero target year
	if _, hasNetZero := tpMetrics["netZeroTargetYear"]; !hasNetZero {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "E1-1_transitionPlan.netZeroTargetYear",
			Code:    "NET_ZERO_TARGET",
			Message: "Net zero target year should be disclosed",
		})
	}
}

// validateEnergy validates E1-5 energy disclosure.
func (v *Validator) validateEnergy(report CSRDReport, results *ValidationResults) {
	energyMetrics, ok := report.Metrics["E1-5_energy"].(map[string]interface{})
	if !ok {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "E1-5_energy",
			Code:    "MISSING_DISCLOSURE",
			Message: "E1-5 energy consumption disclosure is required",
		})
		return
	}

	// Check total energy consumption
	if totalEnergy, ok := energyMetrics["totalEnergyConsumption"].(map[string]interface{}); ok {
		if value, ok := totalEnergy["value"].(float64); ok && value < 0 {
			results.Errors = append(results.Errors, ValidationError{
				Field:   "E1-5_energy.totalEnergyConsumption",
				Code:    "INVALID_VALUE",
				Message: "Total energy consumption cannot be negative",
			})
		}
	} else {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "E1-5_energy.totalEnergyConsumption",
			Code:    "MISSING_FIELD",
			Message: "Total energy consumption should be disclosed",
		})
	}

	// Check energy mix
	if _, hasEnergyMix := energyMetrics["energyMix"]; !hasEnergyMix {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "E1-5_energy.energyMix",
			Code:    "MISSING_BREAKDOWN",
			Message: "Energy mix (renewable, fossil, nuclear) breakdown is recommended",
		})
	}
}

// validateFinancialEffects validates E1-9 financial effects disclosure.
func (v *Validator) validateFinancialEffects(report CSRDReport, results *ValidationResults) {
	feMetrics, ok := report.Metrics["E1-9_financialEffects"].(map[string]interface{})
	if !ok {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "E1-9_financialEffects",
			Code:    "MISSING_DISCLOSURE",
			Message: "E1-9 anticipated financial effects disclosure is recommended",
		})
		return
	}

	status, _ := feMetrics["status"].(string)
	if status == "requires_assessment" {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "E1-9_financialEffects",
			Code:    "INCOMPLETE_ASSESSMENT",
			Message: "Climate-related financial effects assessment is incomplete",
		})
	}

	// Check for physical risks
	physicalRisks, _ := feMetrics["physicalRisks"].([]interface{})
	if len(physicalRisks) == 0 {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "E1-9_financialEffects.physicalRisks",
			Code:    "NO_PHYSICAL_RISKS",
			Message: "Physical climate risks should be assessed",
		})
	}

	// Check for transition risks
	transitionRisks, _ := feMetrics["transitionRisks"].([]interface{})
	if len(transitionRisks) == 0 {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "E1-9_financialEffects.transitionRisks",
			Code:    "NO_TRANSITION_RISKS",
			Message: "Transition risks should be assessed",
		})
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// contains checks if a string slice contains a value.
func contains(slice []string, value string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, value) {
			return true
		}
	}
	return false
}

// =============================================================================
// Validation Rule Builder
// =============================================================================

// ValidatorOption configures the validator.
type ValidatorOption func(*Validator)

// WithStrictMode enables strict validation.
func WithStrictMode() ValidatorOption {
	return func(v *Validator) {
		v.StrictMode = true
	}
}

// WithoutScope3Categories disables Scope 3 category requirement.
func WithoutScope3Categories() ValidatorOption {
	return func(v *Validator) {
		v.RequireScope3Categories = false
	}
}

// WithoutTransitionPlan disables transition plan requirement.
func WithoutTransitionPlan() ValidatorOption {
	return func(v *Validator) {
		v.RequireTransitionPlan = false
	}
}

// WithoutTargets disables targets requirement.
func WithoutTargets() ValidatorOption {
	return func(v *Validator) {
		v.RequireTargets = false
	}
}

// NewValidatorWithOptions creates a validator with custom options.
func NewValidatorWithOptions(opts ...ValidatorOption) *Validator {
	v := NewValidator()
	for _, opt := range opts {
		opt(v)
	}
	return v
}
