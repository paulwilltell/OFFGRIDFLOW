// Package sec provides SEC Climate Disclosure compliance mapping and report generation.
// Implements the SEC Climate-Related Disclosures Rule (17 CFR Parts 210, 229, and 249).
//
// Reference: SEC Final Rule on The Enhancement and Standardization of Climate-Related
// Disclosures for Investors (Release Nos. 33-11042; 34-97063)
package sec

import "time"

// =============================================================================
// SEC Climate Disclosure Input Types
// =============================================================================

// SECInput holds the input data for generating an SEC Climate disclosure.
type SECInput struct {
	// Organization identification
	OrgID      string `json:"org_id"`
	OrgName    string `json:"org_name"`
	FiscalYear int    `json:"fiscal_year"`
	CIK        string `json:"cik"`         // Central Index Key
	FilerType  string `json:"filer_type"`  // "LAF" (Large Accelerated Filer), "AF" (Accelerated Filer), "SRC" (Smaller Reporting Company)
	IsEGC      bool   `json:"is_egc"`      // Emerging Growth Company status

	// Subpart 1500 - Governance
	Governance *GovernanceDisclosure `json:"governance,omitempty"`

	// Subpart 1501 - Risk Management
	RiskManagement *RiskManagementDisclosure `json:"risk_management,omitempty"`

	// Subpart 1502 - Strategy, Business Model, and Outlook
	Strategy *StrategyDisclosure `json:"strategy,omitempty"`

	// Subpart 1504 - GHG Emissions Metrics (Required for LAF and AF)
	GHGMetrics *GHGMetricsDisclosure `json:"ghg_metrics,omitempty"`

	// Financial Statement Impact (Regulation S-X Article 14)
	FinancialImpact *FinancialStatementImpact `json:"financial_impact,omitempty"`

	// Attestation (LAF only, phased in 2025-2026)
	Attestation *AttestationReport `json:"attestation,omitempty"`
}

// =============================================================================
// Item 1500: Governance
// =============================================================================

// GovernanceDisclosure captures board and management oversight of climate risks.
type GovernanceDisclosure struct {
	// Board oversight
	BoardOversight BoardOversightInfo `json:"board_oversight"`

	// Management role
	ManagementRole ManagementRoleInfo `json:"management_role"`

	// Governance processes
	GovernanceProcesses []GovernanceProcess `json:"governance_processes,omitempty"`
}

// BoardOversightInfo describes board-level climate oversight.
type BoardOversightInfo struct {
	HasBoardOversight        bool     `json:"has_board_oversight"`
	ResponsibleCommittee     string   `json:"responsible_committee,omitempty"`     // e.g., "Audit Committee", "Risk Committee"
	CommitteeCharter         string   `json:"committee_charter,omitempty"`
	OversightFrequency       string   `json:"oversight_frequency,omitempty"`       // e.g., "quarterly", "annually"
	DirectorsWithExpertise   []string `json:"directors_with_expertise,omitempty"`
	IntegrationWithStrategy  bool     `json:"integration_with_strategy"`
	RiskOversightDescription string   `json:"risk_oversight_description,omitempty"`
}

// ManagementRoleInfo describes management's role in assessing and managing climate risks.
type ManagementRoleInfo struct {
	ResponsibleExecutive      string   `json:"responsible_executive,omitempty"`      // e.g., "Chief Sustainability Officer"
	ReportingStructure        string   `json:"reporting_structure,omitempty"`
	ProcessesAndFrequency     string   `json:"processes_and_frequency,omitempty"`
	InformedPositions         []string `json:"informed_positions,omitempty"`         // Titles of positions responsible for climate matters
	IntegrationWithOperations bool     `json:"integration_with_operations"`
}

// GovernanceProcess describes a governance process related to climate matters.
type GovernanceProcess struct {
	ProcessName     string `json:"process_name"`
	Description     string `json:"description"`
	Frequency       string `json:"frequency"`
	KeyParticipants string `json:"key_participants"`
}

// =============================================================================
// Item 1501: Risk Management
// =============================================================================

// RiskManagementDisclosure captures climate risk identification and management processes.
type RiskManagementDisclosure struct {
	// Risk identification and assessment
	RiskIdentification RiskIdentificationProcess `json:"risk_identification"`

	// Risk management processes
	RiskManagement RiskManagementProcess `json:"risk_management"`

	// Integration with ERM
	ERMIntegration ERMIntegrationInfo `json:"erm_integration"`

	// Material climate risks identified
	MaterialRisks []MaterialClimateRisk `json:"material_risks,omitempty"`
}

// RiskIdentificationProcess describes how climate risks are identified.
type RiskIdentificationProcess struct {
	ProcessDescription string   `json:"process_description"`
	RiskCategories     []string `json:"risk_categories"`     // e.g., "physical", "transition", "regulatory"
	TimeHorizons       []string `json:"time_horizons"`       // e.g., "short-term (0-1 years)", "medium-term (1-5 years)", "long-term (5+ years)"
	AssessmentTools    []string `json:"assessment_tools,omitempty"`
	FrequencyOfReview  string   `json:"frequency_of_review"`
}

// RiskManagementProcess describes how identified risks are managed.
type RiskManagementProcess struct {
	ProcessDescription     string                `json:"process_description"`
	MitigationStrategies   []MitigationStrategy  `json:"mitigation_strategies,omitempty"`
	MonitoringFrequency    string                `json:"monitoring_frequency"`
	EscalationProcedures   string                `json:"escalation_procedures,omitempty"`
	AdaptationStrategies   []AdaptationStrategy  `json:"adaptation_strategies,omitempty"`
}

// ERMIntegrationInfo describes integration with enterprise risk management.
type ERMIntegrationInfo struct {
	IsIntegrated        bool   `json:"is_integrated"`
	IntegrationApproach string `json:"integration_approach,omitempty"`
	ERMFramework        string `json:"erm_framework,omitempty"` // e.g., "COSO ERM", "ISO 31000"
}

// MaterialClimateRisk represents a material climate-related risk.
type MaterialClimateRisk struct {
	RiskID          string  `json:"risk_id"`
	RiskType        string  `json:"risk_type"`        // "physical" or "transition"
	RiskCategory    string  `json:"risk_category"`    // e.g., "acute", "chronic", "policy", "technology", "market", "reputation"
	Description     string  `json:"description"`
	TimeHorizon     string  `json:"time_horizon"`     // "short-term", "medium-term", "long-term"
	Likelihood      string  `json:"likelihood,omitempty"`
	PotentialImpact string  `json:"potential_impact,omitempty"`
	MitigationPlan  string  `json:"mitigation_plan,omitempty"`
	FinancialImpact float64 `json:"financial_impact,omitempty"` // USD
}

// MitigationStrategy describes a risk mitigation strategy.
type MitigationStrategy struct {
	StrategyName    string  `json:"strategy_name"`
	Description     string  `json:"description"`
	TargetRisks     []string `json:"target_risks"`
	Implementation  string  `json:"implementation"`
	ExpectedOutcome string  `json:"expected_outcome"`
}

// AdaptationStrategy describes a climate adaptation strategy.
type AdaptationStrategy struct {
	StrategyName    string `json:"strategy_name"`
	Description     string `json:"description"`
	TargetRisks     []string `json:"target_risks"`
	Implementation  string `json:"implementation"`
	ExpectedOutcome string `json:"expected_outcome"`
}

// =============================================================================
// Item 1502: Strategy, Business Model, and Outlook
// =============================================================================

// StrategyDisclosure captures climate impacts on strategy and business model.
type StrategyDisclosure struct {
	// Material climate risks and their actual/potential impacts
	MaterialImpacts []StrategyImpact `json:"material_impacts,omitempty"`

	// Transition plan (if adopted)
	TransitionPlan *TransitionPlan `json:"transition_plan,omitempty"`

	// Scenario analysis (if conducted)
	ScenarioAnalysis *ScenarioAnalysis `json:"scenario_analysis,omitempty"`

	// Internal carbon price (if used)
	InternalCarbonPrice *InternalCarbonPrice `json:"internal_carbon_price,omitempty"`

	// Climate targets (if set)
	ClimateTargets []ClimateTarget `json:"climate_targets,omitempty"`
}

// StrategyImpact describes impact of climate risks on strategy or business model.
type StrategyImpact struct {
	ImpactArea      string  `json:"impact_area"`      // e.g., "operations", "products", "supply chain"
	Description     string  `json:"description"`
	TimeHorizon     string  `json:"time_horizon"`
	Materiality     string  `json:"materiality"`      // "material" or "reasonably_likely_material"
	FinancialImpact float64 `json:"financial_impact,omitempty"` // USD
	ResponseActions string  `json:"response_actions,omitempty"`
}

// TransitionPlan describes the registrant's climate transition plan.
type TransitionPlan struct {
	HasPlan                   bool                     `json:"has_plan"`
	PlanDescription           string                   `json:"plan_description,omitempty"`
	TargetsAndMilestones      []TargetMilestone        `json:"targets_and_milestones,omitempty"`
	KeyActions                []TransitionAction       `json:"key_actions,omitempty"`
	CapitalExpenditures       *TransitionCapEx         `json:"capital_expenditures,omitempty"`
	CompetitiveImplications   string                   `json:"competitive_implications,omitempty"`
	UpdateFrequency           string                   `json:"update_frequency,omitempty"`
}

// TargetMilestone represents a transition plan milestone.
type TargetMilestone struct {
	MilestoneID     string    `json:"milestone_id"`
	Description     string    `json:"description"`
	TargetDate      time.Time `json:"target_date"`
	ProgressMetrics string    `json:"progress_metrics"`
	CurrentStatus   string    `json:"current_status"`
}

// TransitionAction describes a specific transition action.
type TransitionAction struct {
	ActionID        string    `json:"action_id"`
	ActionType      string    `json:"action_type"` // e.g., "emissions reduction", "renewable energy", "efficiency"
	Description     string    `json:"description"`
	Timeline        string    `json:"timeline"`
	ResponsibleParty string   `json:"responsible_party"`
	Budget          float64   `json:"budget,omitempty"`
}

// TransitionCapEx describes capital expenditures related to transition plan.
type TransitionCapEx struct {
	TotalAmount     float64 `json:"total_amount"`     // USD
	ReportingPeriod string  `json:"reporting_period"`
	Breakdown       map[string]float64 `json:"breakdown,omitempty"` // Category -> Amount
	FutureCommitments float64 `json:"future_commitments,omitempty"`
}

// ScenarioAnalysis describes climate scenario analysis conducted.
type ScenarioAnalysis struct {
	Conducted         bool              `json:"conducted"`
	Scenarios         []Scenario        `json:"scenarios,omitempty"`
	AnalysisFrequency string            `json:"analysis_frequency,omitempty"`
	KeyFindings       string            `json:"key_findings,omitempty"`
	Methodology       string            `json:"methodology,omitempty"`
}

// Scenario represents a climate scenario analyzed.
type Scenario struct {
	ScenarioName   string  `json:"scenario_name"`   // e.g., "IEA NZE 2050", "IPCC SSP2-4.5"
	Description    string  `json:"description"`
	TemperatureRise string `json:"temperature_rise"` // e.g., "1.5C", "2C", "3C+"
	TimeHorizon    string  `json:"time_horizon"`
	Impacts        string  `json:"impacts"`
	Resilience     string  `json:"resilience"`
}

// InternalCarbonPrice describes use of internal carbon pricing.
type InternalCarbonPrice struct {
	Used            bool    `json:"used"`
	PricePerTonCO2e float64 `json:"price_per_ton_co2e,omitempty"`
	Currency        string  `json:"currency,omitempty"`
	PriceType       string  `json:"price_type,omitempty"` // "shadow price", "internal fee", "implicit cost"
	ApplicationScope string `json:"application_scope,omitempty"`
	Rationale       string  `json:"rationale,omitempty"`
}

// ClimateTarget represents a climate-related target.
type ClimateTarget struct {
	TargetID         string    `json:"target_id"`
	TargetType       string    `json:"target_type"`       // e.g., "GHG reduction", "renewable energy", "energy efficiency"
	Description      string    `json:"description"`
	BaseYear         int       `json:"base_year,omitempty"`
	BaselineValue    float64   `json:"baseline_value,omitempty"`
	TargetYear       int       `json:"target_year"`
	TargetValue      float64   `json:"target_value"`
	Unit             string    `json:"unit"`
	Scope            string    `json:"scope"`             // e.g., "Scope 1", "Scope 1+2", "All scopes"
	ProgressToDate   float64   `json:"progress_to_date,omitempty"`
	SBTiAligned      bool      `json:"sbti_aligned"`
	TargetSetDate    time.Time `json:"target_set_date,omitempty"`
}

// =============================================================================
// Item 1504: GHG Emissions Metrics (LAF and AF only)
// =============================================================================

// GHGMetricsDisclosure captures GHG emissions metrics disclosure.
type GHGMetricsDisclosure struct {
	// Reporting requirements
	IsRequired      bool `json:"is_required"`      // Based on filer type
	AttestationReq  bool `json:"attestation_req"`  // LAF only, phased implementation

	// Scope 1 and 2 emissions (required for LAF and AF)
	Scope1Emissions *ScopeEmissions `json:"scope1_emissions,omitempty"`
	Scope2Emissions *ScopeEmissions `json:"scope2_emissions,omitempty"`

	// Scope 3 emissions (if material, or if target/goal includes Scope 3)
	Scope3Emissions *Scope3Emissions `json:"scope3_emissions,omitempty"`

	// Emissions intensity metric
	IntensityMetric *IntensityMetric `json:"intensity_metric,omitempty"`

	// Methodology and assumptions
	Methodology MethodologyDisclosure `json:"methodology"`

	// Data quality and assurance
	DataQuality DataQualityInfo `json:"data_quality"`
}

// ScopeEmissions represents Scope 1 or Scope 2 emissions.
type ScopeEmissions struct {
	TotalEmissions  float64 `json:"total_emissions"`  // Metric tons CO2e
	ReportingYear   int     `json:"reporting_year"`
	OrganizationalBoundary string `json:"organizational_boundary"` // e.g., "operational control", "equity share"
	
	// Disaggregation
	ByConstituent  map[string]float64 `json:"by_constituent,omitempty"`  // GHG type -> tons
	ByBusinessUnit map[string]float64 `json:"by_business_unit,omitempty"`
	ByGeography    map[string]float64 `json:"by_geography,omitempty"`
	
	// Scope 2 specific
	LocationBased float64 `json:"location_based,omitempty"` // For Scope 2
	MarketBased   float64 `json:"market_based,omitempty"`   // For Scope 2
	
	// Exclusions
	Exclusions []EmissionExclusion `json:"exclusions,omitempty"`
}

// Scope3Emissions represents Scope 3 emissions by category.
type Scope3Emissions struct {
	TotalEmissions float64 `json:"total_emissions"` // Metric tons CO2e
	ReportingYear  int     `json:"reporting_year"`
	
	// By category (15 categories per GHG Protocol)
	Categories map[int]CategoryEmissions `json:"categories,omitempty"` // Category number (1-15) -> emissions
	
	// Material categories
	MaterialCategories []int `json:"material_categories,omitempty"`
	
	// Explanation for included/excluded categories
	CoverageRationale string `json:"coverage_rationale,omitempty"`
	
	// Scope 3 screening performed
	ScreeningPerformed bool   `json:"screening_performed"`
	ScreeningMethod    string `json:"screening_method,omitempty"`
}

// CategoryEmissions represents emissions for a Scope 3 category.
type CategoryEmissions struct {
	CategoryNumber  int     `json:"category_number"`
	CategoryName    string  `json:"category_name"`
	Emissions       float64 `json:"emissions"`       // Metric tons CO2e
	Methodology     string  `json:"methodology"`
	DataQuality     string  `json:"data_quality"`    // "measured", "estimated", "industry_average"
	PercentOfTotal  float64 `json:"percent_of_total"`
}

// EmissionExclusion describes excluded emission sources.
type EmissionExclusion struct {
	Source      string  `json:"source"`
	Reason      string  `json:"reason"`
	Estimated   float64 `json:"estimated,omitempty"` // Estimated emissions if excluded
}

// IntensityMetric represents GHG emissions intensity.
type IntensityMetric struct {
	MetricType    string  `json:"metric_type"` // e.g., "revenue", "production", "FTE"
	Value         float64 `json:"value"`       // e.g., tCO2e per million USD revenue
	Unit          string  `json:"unit"`
	ReportingYear int     `json:"reporting_year"`
	Methodology   string  `json:"methodology"`
}

// MethodologyDisclosure describes GHG accounting methodology.
type MethodologyDisclosure struct {
	Standard               string   `json:"standard"` // e.g., "GHG Protocol Corporate Standard"
	ConsolidationApproach  string   `json:"consolidation_approach"` // "operational control", "financial control", "equity share"
	EmissionFactorsSource  string   `json:"emission_factors_source"`
	GWPSource              string   `json:"gwp_source"` // e.g., "IPCC AR5", "IPCC AR6"
	BaseYear               int      `json:"base_year,omitempty"`
	Recalculations         []Recalculation `json:"recalculations,omitempty"`
	MaterialityThreshold   float64  `json:"materiality_threshold,omitempty"`
}

// Recalculation describes a base year recalculation.
type Recalculation struct {
	Year         int     `json:"year"`
	Reason       string  `json:"reason"`
	Impact       float64 `json:"impact"` // Change in tCO2e
	Methodology  string  `json:"methodology"`
}

// DataQualityInfo describes data quality and assurance.
type DataQualityInfo struct {
	VerificationStatus   string   `json:"verification_status"`   // "verified", "limited_assurance", "reasonable_assurance", "not_verified"
	VerificationStandard string   `json:"verification_standard,omitempty"` // e.g., "ISO 14064-3", "AA1000AS"
	VerificationProvider string   `json:"verification_provider,omitempty"`
	DataCoverage         float64  `json:"data_coverage"`         // Percentage of emissions based on measured data
	EstimationMethods    []string `json:"estimation_methods,omitempty"`
	UncertaintyAnalysis  string   `json:"uncertainty_analysis,omitempty"`
}

// =============================================================================
// Financial Statement Impact (Regulation S-X Article 14)
// =============================================================================

// FinancialStatementImpact captures climate-related impacts on financial statements.
type FinancialStatementImpact struct {
	// Disclosure threshold: 1% of line item value (aggregated for all climate impacts)
	DisclosureThresholdMet bool `json:"disclosure_threshold_met"`
	
	// Impacts by financial statement line item
	ImpactedItems []LineItemImpact `json:"impacted_items,omitempty"`
	
	// Aggregate impacts by statement
	IncomeStatementImpact   *StatementImpact `json:"income_statement_impact,omitempty"`
	BalanceSheetImpact      *StatementImpact `json:"balance_sheet_impact,omitempty"`
	CashFlowStatementImpact *StatementImpact `json:"cash_flow_statement_impact,omitempty"`
	
	// Expenditure metrics
	Expenditures *ClimateExpenditures `json:"expenditures,omitempty"`
}

// LineItemImpact describes impact on a specific financial statement line item.
type LineItemImpact struct {
	StatementType    string  `json:"statement_type"`    // "income", "balance_sheet", "cash_flow"
	LineItem         string  `json:"line_item"`
	ImpactAmount     float64 `json:"impact_amount"`     // USD (absolute value if positive, negative if favorable)
	ImpactPercentage float64 `json:"impact_percentage"` // Percentage of line item
	ImpactType       string  `json:"impact_type"`       // "positive" or "negative"
	Description      string  `json:"description"`
	ClimateEvent     string  `json:"climate_event,omitempty"` // Associated severe weather event or transition activity
}

// StatementImpact aggregates impacts by financial statement.
type StatementImpact struct {
	TotalImpact  float64 `json:"total_impact"`  // USD
	LineItems    int     `json:"line_items"`    // Number of impacted line items
	MaterialItems []string `json:"material_items,omitempty"`
}

// ClimateExpenditures tracks climate-related expenditures and costs.
type ClimateExpenditures struct {
	// Carbon offsets/RECs purchased
	CarbonOffsetsExpense float64 `json:"carbon_offsets_expense,omitempty"` // USD
	
	// Capital expenditures
	CapitalExpenditures ClimateCapEx `json:"capital_expenditures"`
	
	// Losses from severe weather events
	SevereWeatherLosses []SevereWeatherLoss `json:"severe_weather_losses,omitempty"`
}

// ClimateCapEx tracks climate-related capital expenditures.
type ClimateCapEx struct {
	TotalAmount        float64 `json:"total_amount"` // USD
	PercentOfTotalCapEx float64 `json:"percent_of_total_capex,omitempty"`
	
	// Disaggregation
	ByCategory map[string]float64 `json:"by_category,omitempty"` // e.g., "renewable energy", "efficiency"
	
	// Future commitments
	FutureCommitments  float64 `json:"future_commitments,omitempty"`
}

// SevereWeatherLoss captures costs from severe weather events.
type SevereWeatherLoss struct {
	EventType        string    `json:"event_type"`        // e.g., "hurricane", "flood", "wildfire"
	EventDate        time.Time `json:"event_date"`
	Location         string    `json:"location"`
	TotalLoss        float64   `json:"total_loss"`        // USD
	InsuredPortion   float64   `json:"insured_portion,omitempty"`
	UninsuredPortion float64   `json:"uninsured_portion,omitempty"`
	BusinessImpact   string    `json:"business_impact,omitempty"`
	Disclosure       string    `json:"disclosure"`        // How disclosed in financial statements
}

// =============================================================================
// Attestation Report (LAF only, phased 2025-2026)
// =============================================================================

// AttestationReport represents third-party attestation of GHG emissions.
type AttestationReport struct {
	Required          bool      `json:"required"`           // Based on filer status and phase-in
	AssuranceLevel    string    `json:"assurance_level"`    // "limited" (2025-2027) or "reasonable" (2028+)
	Provider          string    `json:"provider"`
	Standard          string    `json:"standard"`           // e.g., "PCAOB attestation standard", "AICPA AT-C 210"
	OpinionType       string    `json:"opinion_type"`       // "unmodified", "modified", "adverse", "disclaimer"
	OpinionStatement  string    `json:"opinion_statement"`
	ScopesCovered     []string  `json:"scopes_covered"`     // "Scope 1", "Scope 2"
	ReportDate        time.Time `json:"report_date"`
	MaterialWeaknesses []string `json:"material_weaknesses,omitempty"`
}

// =============================================================================
// SEC Climate Report
// =============================================================================

// SECReport represents a complete SEC Climate disclosure report.
type SECReport struct {
	// Header
	OrgID       string    `json:"orgId"`
	OrgName     string    `json:"orgName"`
	CIK         string    `json:"cik"`
	FiscalYear  int       `json:"fiscalYear"`
	FilingType  string    `json:"filingType"` // "10-K", "20-F", "40-F"
	GeneratedAt time.Time `json:"generatedAt"`
	
	// Filer classification
	FilerType string `json:"filerType"` // "LAF", "AF", "SRC", "EGC"
	IsEGC     bool   `json:"isEGC"`
	
	// Required disclosures
	Governance     *GovernanceDisclosure     `json:"governance,omitempty"`
	RiskManagement *RiskManagementDisclosure `json:"riskManagement,omitempty"`
	Strategy       *StrategyDisclosure       `json:"strategy,omitempty"`
	GHGMetrics     *GHGMetricsDisclosure     `json:"ghgMetrics,omitempty"`
	
	// Financial statement impacts
	FinancialImpact *FinancialStatementImpact `json:"financialImpact,omitempty"`
	
	// Attestation (if required)
	Attestation *AttestationReport `json:"attestation,omitempty"`
	
	// Validation results
	ValidationResults *ValidationResults `json:"validationResults,omitempty"`
	
	// Compliance metrics
	ComplianceScore float64 `json:"complianceScore"` // 0-100
	
	// Disclosure status
	RequiredDisclosures []DisclosureStatus `json:"requiredDisclosures"`
}

// DisclosureStatus tracks completion of required disclosures.
type DisclosureStatus struct {
	Item        string `json:"item"`        // e.g., "Item 1500", "Item 1504"
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Complete    bool   `json:"complete"`
	DataQuality string `json:"dataQuality"` // "complete", "partial", "missing"
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
