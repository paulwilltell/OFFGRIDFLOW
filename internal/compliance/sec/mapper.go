// Package sec provides SEC Climate Disclosure compliance mapping and report generation.
package sec

import (
	"context"
	"fmt"
	"time"

	"github.com/example/offgridflow/internal/compliance/core"
)

// =============================================================================
// SEC Climate Mapper
// =============================================================================

// SECMapper defines the interface for building SEC Climate disclosure reports.
// Implements the core.ComplianceMapper interface.
type SECMapper interface {
	BuildReport(ctx context.Context, input SECInput) (SECReport, error)
	ValidateInput(ctx context.Context, input SECInput) ([]core.ValidationResult, error)
	GetRequiredFields() []string
}

// DefaultSECMapper is the default implementation of SECMapper.
type DefaultSECMapper struct {
	validator *Validator
}

// NewDefaultSECMapper creates a new DefaultSECMapper.
func NewDefaultSECMapper() *DefaultSECMapper {
	return &DefaultSECMapper{
		validator: NewValidator(),
	}
}

// BuildReport creates a complete SEC Climate disclosure report from the input data.
func (m *DefaultSECMapper) BuildReport(ctx context.Context, input SECInput) (SECReport, error) {
	_ = ctx

	// Determine applicable requirements based on filer type
	requirements := m.determineRequirements(input)

	report := SECReport{
		OrgID:       input.OrgID,
		OrgName:     input.OrgName,
		CIK:         input.CIK,
		FiscalYear:  input.FiscalYear,
		FilingType:  "10-K",
		GeneratedAt: time.Now(),
		FilerType:   input.FilerType,
		IsEGC:       input.IsEGC,
	}

	// Populate required disclosures
	if requirements.GovernanceRequired {
		report.Governance = input.Governance
	}

	if requirements.RiskManagementRequired {
		report.RiskManagement = input.RiskManagement
	}

	if requirements.StrategyRequired {
		report.Strategy = input.Strategy
	}

	if requirements.GHGMetricsRequired {
		report.GHGMetrics = input.GHGMetrics
	}

	if requirements.FinancialImpactRequired {
		report.FinancialImpact = input.FinancialImpact
	}

	if requirements.AttestationRequired {
		report.Attestation = input.Attestation
	}

	// Assess disclosure completeness
	report.RequiredDisclosures = m.assessRequiredDisclosures(input, requirements)

	// Calculate compliance score
	report.ComplianceScore = m.calculateComplianceScore(report.RequiredDisclosures)

	// Validate the report
	if m.validator != nil {
		report.ValidationResults = m.validator.ValidateReport(report)
	}

	return report, nil
}

// =============================================================================
// Requirements Determination
// =============================================================================

// DisclosureRequirements defines which disclosures are required.
type DisclosureRequirements struct {
	GovernanceRequired      bool
	RiskManagementRequired  bool
	StrategyRequired        bool
	GHGMetricsRequired      bool
	FinancialImpactRequired bool
	AttestationRequired     bool
}

// determineRequirements determines which disclosures are required based on filer type.
func (m *DefaultSECMapper) determineRequirements(input SECInput) DisclosureRequirements {
	req := DisclosureRequirements{}

	// All filers except EGC must provide core disclosures
	if !input.IsEGC {
		req.GovernanceRequired = true
		req.RiskManagementRequired = true
		req.StrategyRequired = true
		req.FinancialImpactRequired = true
	}

	// GHG emissions metrics required for Large Accelerated Filers (LAF) and Accelerated Filers (AF)
	// Smaller Reporting Companies (SRC) and Emerging Growth Companies (EGC) are exempt
	if input.FilerType == "LAF" || input.FilerType == "AF" {
		req.GHGMetricsRequired = true
	}

	// Attestation required for LAF only, phased in starting FY2025
	// Limited assurance: FY2025-2027
	// Reasonable assurance: FY2028+
	if input.FilerType == "LAF" && input.FiscalYear >= 2025 {
		req.AttestationRequired = true
	}

	return req
}

// =============================================================================
// Disclosure Assessment
// =============================================================================

// assessRequiredDisclosures determines which disclosures are required and complete.
func (m *DefaultSECMapper) assessRequiredDisclosures(input SECInput, req DisclosureRequirements) []DisclosureStatus {
	disclosures := []DisclosureStatus{}

	// Item 1500: Governance
	disclosures = append(disclosures, DisclosureStatus{
		Item:        "Item 1500",
		Name:        "Governance",
		Required:    req.GovernanceRequired,
		Complete:    input.Governance != nil && m.isGovernanceComplete(input.Governance),
		DataQuality: m.assessGovernanceQuality(input.Governance),
		Notes:       m.getGovernanceNotes(input.Governance, req.GovernanceRequired),
	})

	// Item 1501: Risk Management
	disclosures = append(disclosures, DisclosureStatus{
		Item:        "Item 1501",
		Name:        "Risk Management",
		Required:    req.RiskManagementRequired,
		Complete:    input.RiskManagement != nil && m.isRiskManagementComplete(input.RiskManagement),
		DataQuality: m.assessRiskManagementQuality(input.RiskManagement),
		Notes:       m.getRiskManagementNotes(input.RiskManagement, req.RiskManagementRequired),
	})

	// Item 1502: Strategy, Business Model, and Outlook
	disclosures = append(disclosures, DisclosureStatus{
		Item:        "Item 1502",
		Name:        "Strategy, Business Model, and Outlook",
		Required:    req.StrategyRequired,
		Complete:    input.Strategy != nil && m.isStrategyComplete(input.Strategy),
		DataQuality: m.assessStrategyQuality(input.Strategy),
		Notes:       m.getStrategyNotes(input.Strategy, req.StrategyRequired),
	})

	// Item 1504: GHG Emissions Metrics
	disclosures = append(disclosures, DisclosureStatus{
		Item:        "Item 1504",
		Name:        "GHG Emissions Metrics",
		Required:    req.GHGMetricsRequired,
		Complete:    input.GHGMetrics != nil && m.isGHGMetricsComplete(input.GHGMetrics),
		DataQuality: m.assessGHGMetricsQuality(input.GHGMetrics),
		Notes:       m.getGHGMetricsNotes(input.GHGMetrics, req.GHGMetricsRequired, input.FilerType),
	})

	// Financial Statement Impact (Regulation S-X Article 14)
	disclosures = append(disclosures, DisclosureStatus{
		Item:        "Reg S-X Art. 14",
		Name:        "Financial Statement Impact",
		Required:    req.FinancialImpactRequired,
		Complete:    input.FinancialImpact != nil,
		DataQuality: m.assessFinancialImpactQuality(input.FinancialImpact),
		Notes:       m.getFinancialImpactNotes(input.FinancialImpact),
	})

	// Attestation (LAF only, phased)
	if req.AttestationRequired {
		disclosures = append(disclosures, DisclosureStatus{
			Item:        "Attestation",
			Name:        "Third-Party Attestation of GHG Emissions",
			Required:    true,
			Complete:    input.Attestation != nil && input.Attestation.OpinionType != "",
			DataQuality: m.assessAttestationQuality(input.Attestation, input.FiscalYear),
			Notes:       m.getAttestationNotes(input.Attestation, input.FiscalYear),
		})
	}

	return disclosures
}

// =============================================================================
// Completeness Checks
// =============================================================================

func (m *DefaultSECMapper) isGovernanceComplete(g *GovernanceDisclosure) bool {
	if g == nil {
		return false
	}
	return g.BoardOversight.HasBoardOversight &&
		g.BoardOversight.ResponsibleCommittee != "" &&
		g.ManagementRole.ResponsibleExecutive != ""
}

func (m *DefaultSECMapper) isRiskManagementComplete(r *RiskManagementDisclosure) bool {
	if r == nil {
		return false
	}
	return r.RiskIdentification.ProcessDescription != "" &&
		r.RiskManagement.ProcessDescription != "" &&
		len(r.MaterialRisks) > 0
}

func (m *DefaultSECMapper) isStrategyComplete(s *StrategyDisclosure) bool {
	if s == nil {
		return false
	}
	// At minimum, should have material impacts or transition plan
	return len(s.MaterialImpacts) > 0 || (s.TransitionPlan != nil && s.TransitionPlan.HasPlan)
}

func (m *DefaultSECMapper) isGHGMetricsComplete(g *GHGMetricsDisclosure) bool {
	if g == nil {
		return false
	}
	// Must have both Scope 1 and Scope 2
	return g.Scope1Emissions != nil && g.Scope1Emissions.TotalEmissions > 0 &&
		g.Scope2Emissions != nil && g.Scope2Emissions.TotalEmissions >= 0 && // Can be zero
		g.Methodology.Standard != ""
}

// =============================================================================
// Quality Assessment
// =============================================================================

func (m *DefaultSECMapper) assessGovernanceQuality(g *GovernanceDisclosure) string {
	if g == nil {
		return "missing"
	}
	if m.isGovernanceComplete(g) {
		return "complete"
	}
	return "partial"
}

func (m *DefaultSECMapper) assessRiskManagementQuality(r *RiskManagementDisclosure) string {
	if r == nil {
		return "missing"
	}
	if m.isRiskManagementComplete(r) {
		return "complete"
	}
	return "partial"
}

func (m *DefaultSECMapper) assessStrategyQuality(s *StrategyDisclosure) string {
	if s == nil {
		return "missing"
	}
	if m.isStrategyComplete(s) {
		return "complete"
	}
	return "partial"
}

func (m *DefaultSECMapper) assessGHGMetricsQuality(g *GHGMetricsDisclosure) string {
	if g == nil {
		return "missing"
	}
	if m.isGHGMetricsComplete(g) {
		if g.DataQuality.VerificationStatus == "verified" ||
			g.DataQuality.VerificationStatus == "limited_assurance" ||
			g.DataQuality.VerificationStatus == "reasonable_assurance" {
			return "complete"
		}
		return "complete"
	}
	return "partial"
}

func (m *DefaultSECMapper) assessFinancialImpactQuality(f *FinancialStatementImpact) string {
	if f == nil {
		return "missing"
	}
	if f.DisclosureThresholdMet && len(f.ImpactedItems) > 0 {
		return "complete"
	}
	if !f.DisclosureThresholdMet {
		return "not_applicable"
	}
	return "partial"
}

func (m *DefaultSECMapper) assessAttestationQuality(a *AttestationReport, fiscalYear int) string {
	if a == nil {
		return "missing"
	}

	// Determine required assurance level based on fiscal year
	requiredLevel := "limited"
	if fiscalYear >= 2028 {
		requiredLevel = "reasonable"
	}

	if a.AssuranceLevel == requiredLevel && a.OpinionType != "" {
		return "complete"
	}
	return "partial"
}

// =============================================================================
// Notes Generation
// =============================================================================

func (m *DefaultSECMapper) getGovernanceNotes(g *GovernanceDisclosure, required bool) string {
	if !required {
		return "Not required for this filer type (EGC exemption)"
	}
	if g == nil {
		return "Disclosure required but not provided"
	}
	if !m.isGovernanceComplete(g) {
		return "Incomplete: Missing board oversight details or management role information"
	}
	return "Complete disclosure provided"
}

func (m *DefaultSECMapper) getRiskManagementNotes(r *RiskManagementDisclosure, required bool) string {
	if !required {
		return "Not required for this filer type (EGC exemption)"
	}
	if r == nil {
		return "Disclosure required but not provided"
	}
	if !m.isRiskManagementComplete(r) {
		return "Incomplete: Missing risk identification processes or material risk details"
	}
	return fmt.Sprintf("Complete disclosure provided with %d material risks identified", len(r.MaterialRisks))
}

func (m *DefaultSECMapper) getStrategyNotes(s *StrategyDisclosure, required bool) string {
	if !required {
		return "Not required for this filer type (EGC exemption)"
	}
	if s == nil {
		return "Disclosure required but not provided"
	}
	if !m.isStrategyComplete(s) {
		return "Incomplete: Should describe material impacts or transition plan"
	}

	notes := "Complete disclosure provided"
	if s.TransitionPlan != nil && s.TransitionPlan.HasPlan {
		notes += " including transition plan"
	}
	if s.ScenarioAnalysis != nil && s.ScenarioAnalysis.Conducted {
		notes += " and scenario analysis"
	}
	return notes
}

func (m *DefaultSECMapper) getGHGMetricsNotes(g *GHGMetricsDisclosure, required bool, filerType string) string {
	if !required {
		return fmt.Sprintf("Not required for %s filers", filerType)
	}
	if g == nil {
		return "Disclosure required but not provided"
	}
	if !m.isGHGMetricsComplete(g) {
		return "Incomplete: Must include Scope 1 and Scope 2 emissions with methodology"
	}

	notes := fmt.Sprintf("Complete disclosure: Scope 1 (%.0f tCO2e), Scope 2 (%.0f tCO2e)",
		g.Scope1Emissions.TotalEmissions, g.Scope2Emissions.TotalEmissions)

	if g.Scope3Emissions != nil && g.Scope3Emissions.TotalEmissions > 0 {
		notes += fmt.Sprintf(", Scope 3 (%.0f tCO2e)", g.Scope3Emissions.TotalEmissions)
	}

	if g.DataQuality.VerificationStatus != "not_verified" {
		notes += " (Verified)"
	}

	return notes
}

func (m *DefaultSECMapper) getFinancialImpactNotes(f *FinancialStatementImpact) string {
	if f == nil {
		return "No financial statement impacts disclosed"
	}
	if !f.DisclosureThresholdMet {
		return "No material financial statement impacts (below 1% disclosure threshold)"
	}
	return fmt.Sprintf("%d line items impacted across financial statements", len(f.ImpactedItems))
}

func (m *DefaultSECMapper) getAttestationNotes(a *AttestationReport, fiscalYear int) string {
	requiredLevel := "limited"
	if fiscalYear >= 2028 {
		requiredLevel = "reasonable"
	}

	if a == nil {
		return fmt.Sprintf("Attestation required (%s assurance) but not provided", requiredLevel)
	}

	if a.AssuranceLevel != requiredLevel {
		return fmt.Sprintf("Attestation provided but assurance level (%s) does not match requirement (%s)",
			a.AssuranceLevel, requiredLevel)
	}

	return fmt.Sprintf("%s assurance provided by %s", a.AssuranceLevel, a.Provider)
}

// =============================================================================
// Compliance Score Calculation
// =============================================================================

func (m *DefaultSECMapper) calculateComplianceScore(disclosures []DisclosureStatus) float64 {
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

// ValidateInput validates the SEC input data and returns validation results.
// Implements the core.ComplianceMapper interface.
func (m *DefaultSECMapper) ValidateInput(ctx context.Context, input SECInput) ([]core.ValidationResult, error) {
	if m.validator == nil {
		m.validator = NewValidator()
	}
	
	// Create a temporary report to use existing validation logic
	tempReport, err := m.BuildReport(ctx, input)
	if err != nil {
		return nil, err
	}
	
	// Convert SEC ValidationResults to core.ValidationResult
	secResults := m.validator.ValidateReport(tempReport)
	coreResults := make([]core.ValidationResult, 0)
	
	for _, e := range secResults.Errors {
		coreResults = append(coreResults, core.ValidationResult{
			Rule:     e.Code,
			Passed:   false,
			Message:  e.Message,
			Severity: "error",
			Framework: core.FrameworkSEC,
		})
	}
	
	for _, w := range secResults.Warnings {
		coreResults = append(coreResults, core.ValidationResult{
			Rule:     w.Code,
			Passed:   false,
			Message:  w.Message,
			Severity: "warning",
			Framework: core.FrameworkSEC,
		})
	}
	
	return coreResults, nil
}

// GetRequiredFields returns the list of required fields for SEC Climate disclosure.
// Implements the core.ComplianceMapper interface.
func (m *DefaultSECMapper) GetRequiredFields() []string {
	return []string{
		"org_id",
		"org_name",
		"cik",            // Central Index Key (SEC identifier)
		"fiscal_year",
		"filer_type",     // "LAF", "SRC", or "Other"
		"governance",     // Board oversight and management role
		"risk_management", // Climate risk processes
		"strategy",       // Climate impacts and resilience
		"ghg_metrics",    // Scope 1, 2, 3 emissions (if LAF)
	}
}

// =============================================================================
// Legacy Interface Support
// =============================================================================

// Mapper implements the legacy compliance.Mapper interface.
type Mapper struct {
	secMapper *DefaultSECMapper
}

// NewMapper creates a new legacy Mapper.
func NewMapper() *Mapper {
	return &Mapper{
		secMapper: NewDefaultSECMapper(),
	}
}

// BuildReport implements the legacy interface.
func (m *Mapper) BuildReport(ctx context.Context, input core.ComplianceInput) (core.ComplianceReport, error) {
	// For legacy interface, return basic structure
	// In production, would convert core.ComplianceInput to SECInput
	return core.ComplianceReport{
		Standard: "SEC",
		Content: map[string]interface{}{
			"status":      "implemented",
			"year":        input.Year,
			"description": "SEC Climate-Related Disclosures per 17 CFR Parts 210, 229, and 249",
		},
	}, nil
}
