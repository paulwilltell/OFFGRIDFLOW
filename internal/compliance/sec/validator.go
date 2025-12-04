package sec

import (
	"fmt"
	"strings"
)

// Validator checks SEC climate rule compliance.
type Validator struct {
	StrictMode bool // Enforce stricter validation rules
}

// NewValidator creates a new SEC climate disclosure validator.
func NewValidator() *Validator {
	return &Validator{
		StrictMode: false,
	}
}

// NewStrictValidator creates a validator with strict mode enabled.
func NewStrictValidator() *Validator {
	return &Validator{
		StrictMode: true,
	}
}

// =============================================================================
// Report Validation
// =============================================================================

// ValidateReport validates a complete SEC Climate disclosure report.
func (v *Validator) ValidateReport(report SECReport) *ValidationResults {
	results := &ValidationResults{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	// Validate header information
	v.validateHeader(report, results)

	// Validate each disclosure section
	if report.Governance != nil {
		v.validateGovernance(report.Governance, results)
	}

	if report.RiskManagement != nil {
		v.validateRiskManagement(report.RiskManagement, results)
	}

	if report.Strategy != nil {
		v.validateStrategy(report.Strategy, results)
	}

	if report.GHGMetrics != nil {
		v.validateGHGMetrics(report.GHGMetrics, report.FilerType, report.FiscalYear, results)
	}

	if report.FinancialImpact != nil {
		v.validateFinancialImpact(report.FinancialImpact, results)
	}

	if report.Attestation != nil {
		v.validateAttestation(report.Attestation, report.FilerType, report.FiscalYear, results)
	}

	// Check for required disclosures based on filer type
	v.validateRequiredDisclosures(report, results)

	// Set overall validity
	results.Valid = len(results.Errors) == 0

	return results
}

// =============================================================================
// Header Validation
// =============================================================================

func (v *Validator) validateHeader(report SECReport, results *ValidationResults) {
	if report.OrgID == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "orgId",
			Code:    "REQUIRED_FIELD",
			Message: "Organization ID is required",
		})
	}

	if report.OrgName == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "orgName",
			Code:    "REQUIRED_FIELD",
			Message: "Organization name is required",
		})
	}

	if report.CIK == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "cik",
			Code:    "REQUIRED_FIELD",
			Message: "Central Index Key (CIK) is required for SEC filings",
		})
	} else if !v.isValidCIK(report.CIK) {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "cik",
			Code:    "INVALID_FORMAT",
			Message: "CIK must be a 10-digit number",
		})
	}

	if report.FiscalYear < 2024 {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "fiscalYear",
			Code:    "EARLY_ADOPTION",
			Message: "SEC Climate Rule mandatory compliance begins FY2025 for LAF, FY2026 for AF",
		})
	}

	if report.FilerType == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "filerType",
			Code:    "REQUIRED_FIELD",
			Message: "Filer type must be specified (LAF, AF, SRC, or EGC)",
		})
	} else if !v.isValidFilerType(report.FilerType) {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "filerType",
			Code:    "INVALID_VALUE",
			Message: "Filer type must be LAF, AF, SRC, or EGC",
		})
	}
}

// =============================================================================
// Governance Validation (Item 1500)
// =============================================================================

func (v *Validator) validateGovernance(g *GovernanceDisclosure, results *ValidationResults) {
	// Board oversight
	if !g.BoardOversight.HasBoardOversight {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "governance.boardOversight",
			Code:    "NO_BOARD_OVERSIGHT",
			Message: "No board-level oversight of climate-related risks disclosed",
		})
	} else {
		if g.BoardOversight.ResponsibleCommittee == "" {
			results.Errors = append(results.Errors, ValidationError{
				Field:   "governance.boardOversight.responsibleCommittee",
				Code:    "REQUIRED_FIELD",
				Message: "Must identify the responsible board committee",
			})
		}

		if g.BoardOversight.OversightFrequency == "" {
			results.Warnings = append(results.Warnings, ValidationWarning{
				Field:   "governance.boardOversight.oversightFrequency",
				Code:    "MISSING_DETAIL",
				Message: "Frequency of board oversight not disclosed",
			})
		}
	}

	// Management role
	if g.ManagementRole.ResponsibleExecutive == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "governance.managementRole.responsibleExecutive",
			Code:    "REQUIRED_FIELD",
			Message: "Must identify executive position(s) responsible for climate-related risks",
		})
	}

	if g.ManagementRole.ProcessesAndFrequency == "" {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "governance.managementRole.processesAndFrequency",
			Code:    "MISSING_DETAIL",
			Message: "Management processes and frequency for assessing climate risks not described",
		})
	}
}

// =============================================================================
// Risk Management Validation (Item 1501)
// =============================================================================

func (v *Validator) validateRiskManagement(r *RiskManagementDisclosure, results *ValidationResults) {
	// Risk identification
	if r.RiskIdentification.ProcessDescription == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "riskManagement.riskIdentification.processDescription",
			Code:    "REQUIRED_FIELD",
			Message: "Must describe processes for identifying climate-related risks",
		})
	}

	if len(r.RiskIdentification.RiskCategories) == 0 {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "riskManagement.riskIdentification.riskCategories",
			Code:    "MISSING_DETAIL",
			Message: "Should specify categories of climate risks considered (physical, transition, etc.)",
		})
	}

	// Risk management
	if r.RiskManagement.ProcessDescription == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "riskManagement.riskManagement.processDescription",
			Code:    "REQUIRED_FIELD",
			Message: "Must describe processes for managing climate-related risks",
		})
	}

	// Material risks
	if len(r.MaterialRisks) == 0 {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "riskManagement.materialRisks",
			Code:    "NO_MATERIAL_RISKS",
			Message: "No material climate-related risks identified",
		})
	} else {
		for i, risk := range r.MaterialRisks {
			v.validateMaterialRisk(risk, i, results)
		}
	}

	// ERM integration
	if !r.ERMIntegration.IsIntegrated {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "riskManagement.ermIntegration",
			Code:    "NOT_INTEGRATED",
			Message: "Climate risk management not integrated with enterprise risk management",
		})
	}
}

func (v *Validator) validateMaterialRisk(risk MaterialClimateRisk, index int, results *ValidationResults) {
	prefix := fmt.Sprintf("riskManagement.materialRisks[%d]", index)

	if risk.RiskType == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".riskType",
			Code:    "REQUIRED_FIELD",
			Message: "Risk type must be specified (physical or transition)",
		})
	} else if risk.RiskType != "physical" && risk.RiskType != "transition" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".riskType",
			Code:    "INVALID_VALUE",
			Message: "Risk type must be 'physical' or 'transition'",
		})
	}

	if risk.Description == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".description",
			Code:    "REQUIRED_FIELD",
			Message: "Risk description is required",
		})
	}

	if risk.TimeHorizon == "" {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   prefix + ".timeHorizon",
			Code:    "MISSING_DETAIL",
			Message: "Time horizon for risk realization should be specified",
		})
	}
}

// =============================================================================
// Strategy Validation (Item 1502)
// =============================================================================

func (v *Validator) validateStrategy(s *StrategyDisclosure, results *ValidationResults) {
	// Material impacts
	if len(s.MaterialImpacts) == 0 && (s.TransitionPlan == nil || !s.TransitionPlan.HasPlan) {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "strategy",
			Code:    "MINIMAL_DISCLOSURE",
			Message: "Should disclose material impacts on strategy or transition plan",
		})
	}

	// Validate material impacts
	for i, impact := range s.MaterialImpacts {
		v.validateStrategyImpact(impact, i, results)
	}

	// Validate transition plan if present
	if s.TransitionPlan != nil && s.TransitionPlan.HasPlan {
		v.validateTransitionPlan(s.TransitionPlan, results)
	}

	// Scenario analysis
	if s.ScenarioAnalysis != nil && s.ScenarioAnalysis.Conducted {
		v.validateScenarioAnalysis(s.ScenarioAnalysis, results)
	}

	// Climate targets
	for i, target := range s.ClimateTargets {
		v.validateClimateTarget(target, i, results)
	}
}

func (v *Validator) validateStrategyImpact(impact StrategyImpact, index int, results *ValidationResults) {
	prefix := fmt.Sprintf("strategy.materialImpacts[%d]", index)

	if impact.Description == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".description",
			Code:    "REQUIRED_FIELD",
			Message: "Impact description is required",
		})
	}

	if impact.TimeHorizon == "" {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   prefix + ".timeHorizon",
			Code:    "MISSING_DETAIL",
			Message: "Time horizon for impact should be specified",
		})
	}
}

func (v *Validator) validateTransitionPlan(tp *TransitionPlan, results *ValidationResults) {
	if tp.PlanDescription == "" {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "strategy.transitionPlan.planDescription",
			Code:    "MISSING_DETAIL",
			Message: "Transition plan description should be provided",
		})
	}

	if len(tp.KeyActions) == 0 {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "strategy.transitionPlan.keyActions",
			Code:    "MISSING_DETAIL",
			Message: "Key decarbonization actions should be described",
		})
	}
}

func (v *Validator) validateScenarioAnalysis(sa *ScenarioAnalysis, results *ValidationResults) {
	if len(sa.Scenarios) == 0 {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "strategy.scenarioAnalysis.scenarios",
			Code:    "MISSING_DETAIL",
			Message: "Scenarios analyzed should be described",
		})
	}

	if sa.Methodology == "" {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "strategy.scenarioAnalysis.methodology",
			Code:    "MISSING_DETAIL",
			Message: "Scenario analysis methodology should be disclosed",
		})
	}
}

func (v *Validator) validateClimateTarget(target ClimateTarget, index int, results *ValidationResults) {
	prefix := fmt.Sprintf("strategy.climateTargets[%d]", index)

	if target.Description == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".description",
			Code:    "REQUIRED_FIELD",
			Message: "Target description is required",
		})
	}

	if target.TargetYear == 0 {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".targetYear",
			Code:    "REQUIRED_FIELD",
			Message: "Target year must be specified",
		})
	}

	if target.Unit == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".unit",
			Code:    "REQUIRED_FIELD",
			Message: "Target unit of measurement is required",
		})
	}

	if target.SBTiAligned && v.StrictMode {
		// Verify target is actually science-based
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   prefix + ".sbtiAligned",
			Code:    "VERIFICATION_RECOMMENDED",
			Message: "SBTi alignment should be verified through official SBTi validation",
		})
	}
}

// =============================================================================
// GHG Metrics Validation (Item 1504)
// =============================================================================

func (v *Validator) validateGHGMetrics(g *GHGMetricsDisclosure, filerType string, fiscalYear int, results *ValidationResults) {
	// Check if GHG metrics are required for this filer type
	ghgRequired := (filerType == "LAF" || filerType == "AF")

	if g == nil {
		if ghgRequired {
			results.Errors = append(results.Errors, ValidationError{
				Field:   "ghgMetrics",
				Code:    "REQUIRED_FIELD",
				Message: "GHG metrics are required for LAF and AF filers",
			})
		}
		return
	}

	if !ghgRequired && g.IsRequired {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "ghgMetrics.isRequired",
			Code:    "VOLUNTARY_DISCLOSURE",
			Message: fmt.Sprintf("GHG metrics not required for %s filers but voluntarily disclosed", filerType),
		})
	}

	// Scope 1 emissions
	if g.Scope1Emissions == nil {
		if ghgRequired {
			results.Errors = append(results.Errors, ValidationError{
				Field:   "ghgMetrics.scope1Emissions",
				Code:    "REQUIRED_FIELD",
				Message: "Scope 1 emissions are required for LAF and AF filers",
			})
		}
	} else {
		v.validateScopeEmissions(g.Scope1Emissions, "scope1", results)
	}

	// Scope 2 emissions
	if g.Scope2Emissions == nil {
		if ghgRequired {
			results.Errors = append(results.Errors, ValidationError{
				Field:   "ghgMetrics.scope2Emissions",
				Code:    "REQUIRED_FIELD",
				Message: "Scope 2 emissions are required for LAF and AF filers",
			})
		}
	} else {
		v.validateScopeEmissions(g.Scope2Emissions, "scope2", results)

		// Scope 2 must have both location-based and market-based
		if g.Scope2Emissions.LocationBased == 0 && g.Scope2Emissions.MarketBased == 0 {
			results.Errors = append(results.Errors, ValidationError{
				Field:   "ghgMetrics.scope2Emissions",
				Code:    "INCOMPLETE_DATA",
				Message: "Scope 2 must include both location-based and market-based methods",
			})
		}
	}

	// Scope 3 emissions (if material or included in targets)
	if g.Scope3Emissions != nil {
		v.validateScope3Emissions(g.Scope3Emissions, results)
	}

	// Methodology
	v.validateMethodology(g.Methodology, results)

	// Data quality
	v.validateDataQuality(g.DataQuality, filerType, results)

	// Intensity metric (recommended but not required)
	if g.IntensityMetric == nil && v.StrictMode {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "ghgMetrics.intensityMetric",
			Code:    "BEST_PRACTICE",
			Message: "Consider including GHG emissions intensity metric for better comparability",
		})
	}
}

func (v *Validator) validateScopeEmissions(se *ScopeEmissions, scope string, results *ValidationResults) {
	prefix := fmt.Sprintf("ghgMetrics.%sEmissions", scope)

	if se.TotalEmissions < 0 {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".totalEmissions",
			Code:    "INVALID_VALUE",
			Message: "Total emissions cannot be negative",
		})
	}

	if se.ReportingYear == 0 {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".reportingYear",
			Code:    "REQUIRED_FIELD",
			Message: "Reporting year must be specified",
		})
	}

	if se.OrganizationalBoundary == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".organizationalBoundary",
			Code:    "REQUIRED_FIELD",
			Message: "Organizational boundary must be specified (e.g., operational control, equity share)",
		})
	}

	// Check for disaggregation
	if len(se.ByConstituent) == 0 && v.StrictMode {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   prefix + ".byConstituent",
			Code:    "BEST_PRACTICE",
			Message: "Consider disaggregating emissions by constituent GHGs (CO2, CH4, N2O, etc.)",
		})
	}
}

func (v *Validator) validateScope3Emissions(s3 *Scope3Emissions, results *ValidationResults) {
	if s3.TotalEmissions < 0 {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "ghgMetrics.scope3Emissions.totalEmissions",
			Code:    "INVALID_VALUE",
			Message: "Total Scope 3 emissions cannot be negative",
		})
	}

	if len(s3.Categories) == 0 {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "ghgMetrics.scope3Emissions.categories",
			Code:    "MISSING_DETAIL",
			Message: "Should disaggregate Scope 3 emissions by category",
		})
	}

	if !s3.ScreeningPerformed {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "ghgMetrics.scope3Emissions.screeningPerformed",
			Code:    "BEST_PRACTICE",
			Message: "Scope 3 screening helps identify material categories",
		})
	}

	// Validate coverage rationale
	if s3.CoverageRationale == "" {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "ghgMetrics.scope3Emissions.coverageRationale",
			Code:    "MISSING_DETAIL",
			Message: "Should explain rationale for included/excluded Scope 3 categories",
		})
	}
}

func (v *Validator) validateMethodology(m MethodologyDisclosure, results *ValidationResults) {
	if m.Standard == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "ghgMetrics.methodology.standard",
			Code:    "REQUIRED_FIELD",
			Message: "GHG accounting standard must be specified (e.g., GHG Protocol)",
		})
	}

	if m.ConsolidationApproach == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "ghgMetrics.methodology.consolidationApproach",
			Code:    "REQUIRED_FIELD",
			Message: "Consolidation approach must be specified",
		})
	}

	if m.GWPSource == "" {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "ghgMetrics.methodology.gwpSource",
			Code:    "MISSING_DETAIL",
			Message: "Global Warming Potential (GWP) source should be specified (e.g., IPCC AR5, AR6)",
		})
	}
}

func (v *Validator) validateDataQuality(dq DataQualityInfo, filerType string, results *ValidationResults) {
	if dq.VerificationStatus == "" {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "ghgMetrics.dataQuality.verificationStatus",
			Code:    "MISSING_DETAIL",
			Message: "Verification status should be disclosed",
		})
	}

	// LAF should have verified emissions
	if filerType == "LAF" && dq.VerificationStatus == "not_verified" {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "ghgMetrics.dataQuality.verificationStatus",
			Code:    "BEST_PRACTICE",
			Message: "Large Accelerated Filers should consider third-party verification of emissions data",
		})
	}

	if dq.DataCoverage < 50 && v.StrictMode {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "ghgMetrics.dataQuality.dataCoverage",
			Code:    "DATA_QUALITY",
			Message: fmt.Sprintf("Data coverage is only %.0f%% - consider improving measurement coverage", dq.DataCoverage),
		})
	}
}

// =============================================================================
// Financial Impact Validation (Regulation S-X Article 14)
// =============================================================================

func (v *Validator) validateFinancialImpact(f *FinancialStatementImpact, results *ValidationResults) {
	if f.DisclosureThresholdMet && len(f.ImpactedItems) == 0 {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "financialImpact.impactedItems",
			Code:    "INCONSISTENT_DATA",
			Message: "If disclosure threshold is met, impacted line items must be specified",
		})
	}

	// Validate line item impacts
	for i, item := range f.ImpactedItems {
		v.validateLineItemImpact(item, i, results)
	}

	// Validate severe weather losses
	if f.Expenditures != nil {
		for i, loss := range f.Expenditures.SevereWeatherLosses {
			v.validateSevereWeatherLoss(loss, i, results)
		}
	}

	// Check for 1% threshold
	if len(f.ImpactedItems) > 0 {
		hasSignificantImpact := false
		for _, item := range f.ImpactedItems {
			if item.ImpactPercentage >= 1.0 {
				hasSignificantImpact = true
				break
			}
		}

		if !hasSignificantImpact && f.DisclosureThresholdMet {
			results.Warnings = append(results.Warnings, ValidationWarning{
				Field:   "financialImpact.disclosureThresholdMet",
				Code:    "VERIFICATION_NEEDED",
				Message: "Disclosure threshold marked as met but no line item shows ≥1% impact",
			})
		}
	}
}

func (v *Validator) validateLineItemImpact(item LineItemImpact, index int, results *ValidationResults) {
	prefix := fmt.Sprintf("financialImpact.impactedItems[%d]", index)

	if item.LineItem == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".lineItem",
			Code:    "REQUIRED_FIELD",
			Message: "Line item name must be specified",
		})
	}

	if item.ImpactAmount == 0 {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   prefix + ".impactAmount",
			Code:    "ZERO_VALUE",
			Message: "Impact amount is zero - verify if correct",
		})
	}

	if item.ImpactPercentage < 0 || item.ImpactPercentage > 100 {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".impactPercentage",
			Code:    "INVALID_VALUE",
			Message: "Impact percentage must be between 0 and 100",
		})
	}

	if item.Description == "" {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   prefix + ".description",
			Code:    "MISSING_DETAIL",
			Message: "Impact description should be provided",
		})
	}
}

func (v *Validator) validateSevereWeatherLoss(loss SevereWeatherLoss, index int, results *ValidationResults) {
	prefix := fmt.Sprintf("financialImpact.expenditures.severeWeatherLosses[%d]", index)

	if loss.EventType == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   prefix + ".eventType",
			Code:    "REQUIRED_FIELD",
			Message: "Severe weather event type must be specified",
		})
	}

	if loss.TotalLoss == 0 {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   prefix + ".totalLoss",
			Code:    "ZERO_VALUE",
			Message: "Total loss is zero - verify if correct",
		})
	}

	if loss.EventDate.IsZero() {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   prefix + ".eventDate",
			Code:    "MISSING_DETAIL",
			Message: "Event date should be specified",
		})
	}

	// Check that insured + uninsured = total
	if loss.InsuredPortion+loss.UninsuredPortion > 0 {
		total := loss.InsuredPortion + loss.UninsuredPortion
		tolerance := 0.01 * loss.TotalLoss
		if total < loss.TotalLoss-tolerance || total > loss.TotalLoss+tolerance {
			results.Warnings = append(results.Warnings, ValidationWarning{
				Field:   prefix,
				Code:    "INCONSISTENT_DATA",
				Message: "Insured + uninsured portions do not sum to total loss",
			})
		}
	}
}

// =============================================================================
// Attestation Validation
// =============================================================================

func (v *Validator) validateAttestation(a *AttestationReport, filerType string, fiscalYear int, results *ValidationResults) {
	// Only LAF requires attestation
	if filerType != "LAF" {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "attestation",
			Code:    "NOT_REQUIRED",
			Message: fmt.Sprintf("Attestation not required for %s filers", filerType),
		})
		return
	}

	// Check phase-in requirements
	if fiscalYear < 2025 {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "attestation",
			Code:    "EARLY_ADOPTION",
			Message: "Attestation required starting FY2025 for LAF",
		})
	}

	// Validate assurance level based on year
	requiredLevel := "limited"
	if fiscalYear >= 2028 {
		requiredLevel = "reasonable"
	}

	if a.AssuranceLevel != requiredLevel {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "attestation.assuranceLevel",
			Code:    "INCORRECT_ASSURANCE_LEVEL",
			Message: fmt.Sprintf("FY%d requires %s assurance, but %s provided", fiscalYear, requiredLevel, a.AssuranceLevel),
		})
	}

	if a.Provider == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "attestation.provider",
			Code:    "REQUIRED_FIELD",
			Message: "Attestation provider must be identified",
		})
	}

	if a.Standard == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "attestation.standard",
			Code:    "REQUIRED_FIELD",
			Message: "Attestation standard must be specified",
		})
	}

	if a.OpinionType == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "attestation.opinionType",
			Code:    "REQUIRED_FIELD",
			Message: "Opinion type must be specified",
		})
	} else if a.OpinionType == "adverse" || a.OpinionType == "disclaimer" {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "attestation.opinionType",
			Code:    "ADVERSE_OPINION",
			Message: fmt.Sprintf("Attestation resulted in %s opinion - emissions data quality issues identified", a.OpinionType),
		})
	}

	if len(a.ScopesCovered) == 0 {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "attestation.scopesCovered",
			Code:    "REQUIRED_FIELD",
			Message: "Scopes covered by attestation must be specified",
		})
	} else {
		// Attestation must cover at least Scope 1 and 2
		hasScope1 := false
		hasScope2 := false
		for _, scope := range a.ScopesCovered {
			if strings.Contains(strings.ToLower(scope), "scope 1") || strings.Contains(scope, "1") {
				hasScope1 = true
			}
			if strings.Contains(strings.ToLower(scope), "scope 2") || strings.Contains(scope, "2") {
				hasScope2 = true
			}
		}

		if !hasScope1 || !hasScope2 {
			results.Errors = append(results.Errors, ValidationError{
				Field:   "attestation.scopesCovered",
				Code:    "INCOMPLETE_SCOPE",
				Message: "Attestation must cover both Scope 1 and Scope 2 emissions",
			})
		}
	}

	if len(a.MaterialWeaknesses) > 0 {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "attestation.materialWeaknesses",
			Code:    "MATERIAL_WEAKNESSES",
			Message: fmt.Sprintf("%d material weakness(es) identified in GHG emissions controls", len(a.MaterialWeaknesses)),
		})
	}
}

// =============================================================================
// Required Disclosures Validation
// =============================================================================

func (v *Validator) validateRequiredDisclosures(report SECReport, results *ValidationResults) {
	// Emerging Growth Companies are exempt from all climate disclosures
	if report.IsEGC {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "filerType",
			Code:    "EGC_EXEMPTION",
			Message: "Emerging Growth Company (EGC) is exempt from SEC climate disclosures",
		})
		return
	}

	// All non-EGC filers must provide governance, risk management, and strategy
	if report.Governance == nil {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "governance",
			Code:    "REQUIRED_DISCLOSURE",
			Message: "Item 1500 (Governance) disclosure is required",
		})
	}

	if report.RiskManagement == nil {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "riskManagement",
			Code:    "REQUIRED_DISCLOSURE",
			Message: "Item 1501 (Risk Management) disclosure is required",
		})
	}

	if report.Strategy == nil {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "strategy",
			Code:    "REQUIRED_DISCLOSURE",
			Message: "Item 1502 (Strategy) disclosure is required",
		})
	}

	// GHG metrics required for LAF and AF
	if (report.FilerType == "LAF" || report.FilerType == "AF") && report.GHGMetrics == nil {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "ghgMetrics",
			Code:    "REQUIRED_DISCLOSURE",
			Message: fmt.Sprintf("Item 1504 (GHG Emissions Metrics) disclosure is required for %s filers", report.FilerType),
		})
	}

	// Attestation required for LAF starting FY2025
	if report.FilerType == "LAF" && report.FiscalYear >= 2025 && report.Attestation == nil {
		results.Errors = append(results.Errors, ValidationError{
			Field:   "attestation",
			Code:    "REQUIRED_DISCLOSURE",
			Message: "Third-party attestation of GHG emissions is required for LAF starting FY2025",
		})
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func (v *Validator) isValidCIK(cik string) bool {
	// CIK should be 10 digits (can have leading zeros)
	if len(cik) != 10 {
		return false
	}
	for _, c := range cik {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func (v *Validator) isValidFilerType(filerType string) bool {
	validTypes := map[string]bool{
		"LAF": true, // Large Accelerated Filer
		"AF":  true, // Accelerated Filer
		"SRC": true, // Smaller Reporting Company
		"EGC": true, // Emerging Growth Company
	}
	return validTypes[filerType]
}
