// Package sec provides SEC Climate Disclosure report building for 10-K filings.
package sec

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// =============================================================================
// 10-K Report Builder
// =============================================================================

// ReportBuilder builds formatted SEC Climate disclosure reports for 10-K filings.
type ReportBuilder struct {
	mapper    *DefaultSECMapper
	validator *Validator
}

// NewReportBuilder creates a new 10-K report builder.
func NewReportBuilder() *ReportBuilder {
	return &ReportBuilder{
		mapper:    NewDefaultSECMapper(),
		validator: NewValidator(),
	}
}

// Build10KReport creates a formatted 10-K climate disclosure report.
func (rb *ReportBuilder) Build10KReport(ctx context.Context, input SECInput) (*Report10K, error) {
	// Build the structured report
	secReport, err := rb.mapper.BuildReport(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to build SEC report: %w", err)
	}

	// Generate formatted sections
	report := &Report10K{
		Header:             rb.buildHeader(secReport),
		Item1500Governance: rb.buildItem1500(secReport.Governance),
		Item1501RiskMgmt:   rb.buildItem1501(secReport.RiskManagement),
		Item1502Strategy:   rb.buildItem1502(secReport.Strategy),
		Item1504GHGMetrics: rb.buildItem1504(secReport.GHGMetrics),
		FinancialImpact:    rb.buildFinancialImpact(secReport.FinancialImpact),
		AttestationReport:  rb.buildAttestationReport(secReport.Attestation),
		ValidationSummary:  rb.buildValidationSummary(secReport.ValidationResults),
		ComplianceScore:    secReport.ComplianceScore,
		GeneratedAt:        secReport.GeneratedAt,
	}

	return report, nil
}

// Report10K represents a formatted 10-K climate disclosure report.
type Report10K struct {
	Header             string    `json:"header"`
	Item1500Governance string    `json:"item1500_governance"`
	Item1501RiskMgmt   string    `json:"item1501_risk_management"`
	Item1502Strategy   string    `json:"item1502_strategy"`
	Item1504GHGMetrics string    `json:"item1504_ghg_metrics"`
	FinancialImpact    string    `json:"financial_impact"`
	AttestationReport  string    `json:"attestation_report"`
	ValidationSummary  string    `json:"validation_summary"`
	ComplianceScore    float64   `json:"compliance_score"`
	GeneratedAt        time.Time `json:"generated_at"`
}

// =============================================================================
// Header
// =============================================================================

func (rb *ReportBuilder) buildHeader(report SECReport) string {
	var b strings.Builder

	b.WriteString("CLIMATE-RELATED DISCLOSURES\n")
	b.WriteString("Pursuant to SEC Rule 17 CFR Parts 210, 229, and 249\n\n")
	b.WriteString(fmt.Sprintf("Registrant: %s\n", report.OrgName))
	b.WriteString(fmt.Sprintf("CIK: %s\n", report.CIK))
	b.WriteString(fmt.Sprintf("Fiscal Year Ended: December 31, %d\n", report.FiscalYear))
	b.WriteString(fmt.Sprintf("Filer Status: %s\n", rb.getFilerStatusName(report.FilerType)))
	if report.IsEGC {
		b.WriteString("Emerging Growth Company: Yes\n")
	}
	b.WriteString(fmt.Sprintf("Report Generated: %s\n", report.GeneratedAt.Format("January 2, 2006")))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("=", 80))
	b.WriteString("\n\n")

	return b.String()
}

// =============================================================================
// Item 1500: Governance
// =============================================================================

func (rb *ReportBuilder) buildItem1500(g *GovernanceDisclosure) string {
	if g == nil {
		return "Item 1500: Governance\n\nNot disclosed.\n\n"
	}

	var b strings.Builder

	b.WriteString("Item 1500: Climate-Related Governance\n")
	b.WriteString(strings.Repeat("-", 80))
	b.WriteString("\n\n")

	// Board Oversight
	b.WriteString("(a) Board Oversight of Climate-Related Risks\n\n")
	if g.BoardOversight.HasBoardOversight {
		b.WriteString(fmt.Sprintf("The %s has oversight responsibility for climate-related risks. ",
			g.BoardOversight.ResponsibleCommittee))

		if g.BoardOversight.OversightFrequency != "" {
			b.WriteString(fmt.Sprintf("The committee reviews climate-related matters %s. ",
				g.BoardOversight.OversightFrequency))
		}

		if g.BoardOversight.RiskOversightDescription != "" {
			b.WriteString(fmt.Sprintf("\n\n%s\n", g.BoardOversight.RiskOversightDescription))
		}

		if len(g.BoardOversight.DirectorsWithExpertise) > 0 {
			b.WriteString(fmt.Sprintf("\n\nBoard members with climate-related expertise: %s\n",
				strings.Join(g.BoardOversight.DirectorsWithExpertise, ", ")))
		}
	} else {
		b.WriteString("The Board does not have specific oversight of climate-related risks.\n")
	}

	b.WriteString("\n")

	// Management's Role
	b.WriteString("(b) Management's Role in Assessing and Managing Climate-Related Risks\n\n")
	if g.ManagementRole.ResponsibleExecutive != "" {
		b.WriteString(fmt.Sprintf("The %s is responsible for assessing and managing climate-related risks. ",
			g.ManagementRole.ResponsibleExecutive))

		if g.ManagementRole.ReportingStructure != "" {
			b.WriteString(fmt.Sprintf("%s ", g.ManagementRole.ReportingStructure))
		}

		if g.ManagementRole.ProcessesAndFrequency != "" {
			b.WriteString(fmt.Sprintf("\n\n%s\n", g.ManagementRole.ProcessesAndFrequency))
		}

		if len(g.ManagementRole.InformedPositions) > 0 {
			b.WriteString(fmt.Sprintf("\n\nThe following positions are informed about climate-related risks: %s\n",
				strings.Join(g.ManagementRole.InformedPositions, ", ")))
		}
	}

	b.WriteString("\n\n")
	return b.String()
}

// =============================================================================
// Item 1501: Risk Management
// =============================================================================

func (rb *ReportBuilder) buildItem1501(r *RiskManagementDisclosure) string {
	if r == nil {
		return "Item 1501: Risk Management\n\nNot disclosed.\n\n"
	}

	var b strings.Builder

	b.WriteString("Item 1501: Climate-Related Risk Management\n")
	b.WriteString(strings.Repeat("-", 80))
	b.WriteString("\n\n")

	// Risk Identification
	b.WriteString("(a) Processes for Identifying Climate-Related Risks\n\n")
	b.WriteString(fmt.Sprintf("%s\n", r.RiskIdentification.ProcessDescription))

	if len(r.RiskIdentification.RiskCategories) > 0 {
		b.WriteString(fmt.Sprintf("\nRisk categories considered: %s\n",
			strings.Join(r.RiskIdentification.RiskCategories, ", ")))
	}

	if len(r.RiskIdentification.TimeHorizons) > 0 {
		b.WriteString(fmt.Sprintf("Time horizons assessed: %s\n",
			strings.Join(r.RiskIdentification.TimeHorizons, ", ")))
	}

	b.WriteString("\n")

	// Risk Management
	b.WriteString("(b) Processes for Managing Climate-Related Risks\n\n")
	b.WriteString(fmt.Sprintf("%s\n", r.RiskManagement.ProcessDescription))

	if len(r.RiskManagement.MitigationStrategies) > 0 {
		b.WriteString("\nMitigation Strategies:\n")
		for i, strategy := range r.RiskManagement.MitigationStrategies {
			b.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, strategy.StrategyName, strategy.Description))
		}
	}

	b.WriteString("\n")

	// Integration with ERM
	b.WriteString("(c) Integration with Enterprise Risk Management\n\n")
	if r.ERMIntegration.IsIntegrated {
		b.WriteString("Climate-related risk management is integrated into our overall enterprise risk management system. ")
		if r.ERMIntegration.ERMFramework != "" {
			b.WriteString(fmt.Sprintf("We use the %s framework. ", r.ERMIntegration.ERMFramework))
		}
		if r.ERMIntegration.IntegrationApproach != "" {
			b.WriteString(fmt.Sprintf("\n\n%s", r.ERMIntegration.IntegrationApproach))
		}
	} else {
		b.WriteString("Climate-related risks are managed separately from our enterprise risk management system.")
	}

	b.WriteString("\n\n")

	// Material Risks
	if len(r.MaterialRisks) > 0 {
		b.WriteString("Material Climate-Related Risks Identified:\n\n")
		for i, risk := range r.MaterialRisks {
			b.WriteString(fmt.Sprintf("%d. %s Risk - %s\n", i+1,
				cases.Title(language.English).String(risk.RiskType), risk.Description))
			b.WriteString(fmt.Sprintf("   Time Horizon: %s\n", risk.TimeHorizon))
			if risk.MitigationPlan != "" {
				b.WriteString(fmt.Sprintf("   Mitigation: %s\n", risk.MitigationPlan))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	return b.String()
}

// =============================================================================
// Item 1502: Strategy
// =============================================================================

func (rb *ReportBuilder) buildItem1502(s *StrategyDisclosure) string {
	if s == nil {
		return "Item 1502: Strategy, Business Model, and Outlook\n\nNot disclosed.\n\n"
	}

	var b strings.Builder

	b.WriteString("Item 1502: Climate-Related Strategy, Business Model, and Outlook\n")
	b.WriteString(strings.Repeat("-", 80))
	b.WriteString("\n\n")

	// Material Impacts
	if len(s.MaterialImpacts) > 0 {
		b.WriteString("(a) Material Climate-Related Impacts\n\n")
		for i, impact := range s.MaterialImpacts {
			b.WriteString(fmt.Sprintf("%d. Impact on %s (%s time horizon)\n",
				i+1, impact.ImpactArea, impact.TimeHorizon))
			b.WriteString(fmt.Sprintf("   %s\n", impact.Description))
			if impact.ResponseActions != "" {
				b.WriteString(fmt.Sprintf("   Response: %s\n", impact.ResponseActions))
			}
			b.WriteString("\n")
		}
	}

	// Transition Plan
	if s.TransitionPlan != nil && s.TransitionPlan.HasPlan {
		b.WriteString("(b) Climate Transition Plan\n\n")
		b.WriteString(fmt.Sprintf("%s\n\n", s.TransitionPlan.PlanDescription))

		if len(s.TransitionPlan.TargetsAndMilestones) > 0 {
			b.WriteString("Key Milestones:\n")
			for _, milestone := range s.TransitionPlan.TargetsAndMilestones {
				b.WriteString(fmt.Sprintf("- %s (Target: %s) - %s\n",
					milestone.Description,
					milestone.TargetDate.Format("2006"),
					milestone.CurrentStatus))
			}
			b.WriteString("\n")
		}

		if s.TransitionPlan.CapitalExpenditures != nil {
			b.WriteString(fmt.Sprintf("Transition-related capital expenditures: $%.0f million\n\n",
				s.TransitionPlan.CapitalExpenditures.TotalAmount/1000000))
		}
	}

	// Scenario Analysis
	if s.ScenarioAnalysis != nil && s.ScenarioAnalysis.Conducted {
		b.WriteString("(c) Scenario Analysis\n\n")
		b.WriteString(fmt.Sprintf("Methodology: %s\n\n", s.ScenarioAnalysis.Methodology))

		if len(s.ScenarioAnalysis.Scenarios) > 0 {
			b.WriteString("Scenarios Analyzed:\n")
			for _, scenario := range s.ScenarioAnalysis.Scenarios {
				b.WriteString(fmt.Sprintf("- %s (%s temperature rise)\n",
					scenario.ScenarioName, scenario.TemperatureRise))
				b.WriteString(fmt.Sprintf("  %s\n", scenario.Description))
			}
			b.WriteString("\n")
		}

		if s.ScenarioAnalysis.KeyFindings != "" {
			b.WriteString(fmt.Sprintf("Key Findings: %s\n\n", s.ScenarioAnalysis.KeyFindings))
		}
	}

	// Internal Carbon Price
	if s.InternalCarbonPrice != nil && s.InternalCarbonPrice.Used {
		b.WriteString("(d) Internal Carbon Price\n\n")
		b.WriteString(fmt.Sprintf("We use an internal carbon price of $%.2f per ton CO2e (%s). ",
			s.InternalCarbonPrice.PricePerTonCO2e, s.InternalCarbonPrice.PriceType))
		if s.InternalCarbonPrice.Rationale != "" {
			b.WriteString(fmt.Sprintf("\n\n%s\n\n", s.InternalCarbonPrice.Rationale))
		}
	}

	// Climate Targets
	if len(s.ClimateTargets) > 0 {
		b.WriteString("(e) Climate-Related Targets and Goals\n\n")
		for i, target := range s.ClimateTargets {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, target.Description))
			b.WriteString(fmt.Sprintf("   Target: %s by %d (%s)\n",
				rb.formatTargetValue(target.TargetValue, target.Unit),
				target.TargetYear, target.Scope))
			if target.BaseYear > 0 {
				b.WriteString(fmt.Sprintf("   Base year: %d (%.0f %s)\n",
					target.BaseYear, target.BaselineValue, target.Unit))
			}
			if target.SBTiAligned {
				b.WriteString("   Aligned with Science Based Targets initiative\n")
			}
			if target.ProgressToDate > 0 {
				b.WriteString(fmt.Sprintf("   Progress to date: %.1f%%\n", target.ProgressToDate))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	return b.String()
}

// =============================================================================
// Item 1504: GHG Emissions Metrics
// =============================================================================

func (rb *ReportBuilder) buildItem1504(g *GHGMetricsDisclosure) string {
	if g == nil {
		return "Item 1504: Greenhouse Gas Emissions Metrics\n\nNot required for this filer type.\n\n"
	}

	var b strings.Builder

	b.WriteString("Item 1504: Greenhouse Gas Emissions Metrics\n")
	b.WriteString(strings.Repeat("-", 80))
	b.WriteString("\n\n")

	// Scope 1 and 2
	b.WriteString("(a) Scope 1 and Scope 2 Emissions\n\n")

	if g.Scope1Emissions != nil {
		b.WriteString(fmt.Sprintf("Scope 1 (Direct) Emissions: %s metric tons CO2e\n",
			rb.formatNumber(g.Scope1Emissions.TotalEmissions)))
	}

	if g.Scope2Emissions != nil {
		b.WriteString("Scope 2 (Indirect) Emissions:\n")
		b.WriteString(fmt.Sprintf("  Location-based: %s metric tons CO2e\n",
			rb.formatNumber(g.Scope2Emissions.LocationBased)))
		b.WriteString(fmt.Sprintf("  Market-based: %s metric tons CO2e\n",
			rb.formatNumber(g.Scope2Emissions.MarketBased)))
	}

	totalScope12 := 0.0
	if g.Scope1Emissions != nil {
		totalScope12 += g.Scope1Emissions.TotalEmissions
	}
	if g.Scope2Emissions != nil {
		totalScope12 += g.Scope2Emissions.MarketBased
	}
	b.WriteString(fmt.Sprintf("\nTotal Scope 1 + 2 Emissions (market-based): %s metric tons CO2e\n\n",
		rb.formatNumber(totalScope12)))

	// Scope 3
	if g.Scope3Emissions != nil && g.Scope3Emissions.TotalEmissions > 0 {
		b.WriteString("(b) Scope 3 Emissions\n\n")
		b.WriteString(fmt.Sprintf("Total Scope 3 Emissions: %s metric tons CO2e\n\n",
			rb.formatNumber(g.Scope3Emissions.TotalEmissions)))

		if len(g.Scope3Emissions.Categories) > 0 {
			b.WriteString("Emissions by Category:\n")
			for catNum := 1; catNum <= 15; catNum++ {
				if cat, exists := g.Scope3Emissions.Categories[catNum]; exists {
					b.WriteString(fmt.Sprintf("  Category %d - %s: %s tCO2e (%.1f%%)\n",
						catNum, cat.CategoryName,
						rb.formatNumber(cat.Emissions),
						cat.PercentOfTotal))
				}
			}
			b.WriteString("\n")
		}

		if g.Scope3Emissions.CoverageRationale != "" {
			b.WriteString(fmt.Sprintf("Coverage Rationale: %s\n\n", g.Scope3Emissions.CoverageRationale))
		}
	}

	// Intensity Metric
	if g.IntensityMetric != nil {
		b.WriteString("(c) Emissions Intensity\n\n")
		b.WriteString(fmt.Sprintf("Emissions Intensity: %.2f tCO2e per %s\n\n",
			g.IntensityMetric.Value, g.IntensityMetric.MetricType))
	}

	// Methodology
	b.WriteString("(d) Methodology and Assumptions\n\n")
	b.WriteString(fmt.Sprintf("Standard: %s\n", g.Methodology.Standard))
	b.WriteString(fmt.Sprintf("Organizational Boundary: %s\n", g.Methodology.ConsolidationApproach))
	if g.Methodology.GWPSource != "" {
		b.WriteString(fmt.Sprintf("GWP Source: %s\n", g.Methodology.GWPSource))
	}
	if g.Methodology.BaseYear > 0 {
		b.WriteString(fmt.Sprintf("Base Year: %d\n", g.Methodology.BaseYear))
	}

	b.WriteString("\n")

	// Data Quality
	b.WriteString("(e) Data Quality and Assurance\n\n")
	b.WriteString(fmt.Sprintf("Verification Status: %s\n", g.DataQuality.VerificationStatus))
	b.WriteString(fmt.Sprintf("Data Coverage: %.1f%% based on measured data\n",
		g.DataQuality.DataCoverage))
	if g.DataQuality.VerificationProvider != "" {
		b.WriteString(fmt.Sprintf("Verified by: %s\n", g.DataQuality.VerificationProvider))
	}

	b.WriteString("\n\n")
	return b.String()
}

// =============================================================================
// Financial Statement Impact
// =============================================================================

func (rb *ReportBuilder) buildFinancialImpact(f *FinancialStatementImpact) string {
	if f == nil {
		return "Financial Statement Impact\n\nNot disclosed.\n\n"
	}

	var b strings.Builder

	b.WriteString("Financial Statement Impact (Regulation S-X Article 14)\n")
	b.WriteString(strings.Repeat("-", 80))
	b.WriteString("\n\n")

	if !f.DisclosureThresholdMet {
		b.WriteString("No climate-related impacts exceed the 1% materiality threshold for financial statement disclosure.\n\n")
		return b.String()
	}

	b.WriteString("The following line items were impacted by climate-related events and transition activities:\n\n")

	// Group by statement type
	incomeItems := []LineItemImpact{}
	balanceItems := []LineItemImpact{}
	cashFlowItems := []LineItemImpact{}

	for _, item := range f.ImpactedItems {
		switch item.StatementType {
		case "income":
			incomeItems = append(incomeItems, item)
		case "balance_sheet":
			balanceItems = append(balanceItems, item)
		case "cash_flow":
			cashFlowItems = append(cashFlowItems, item)
		}
	}

	if len(incomeItems) > 0 {
		b.WriteString("Income Statement:\n")
		for _, item := range incomeItems {
			b.WriteString(fmt.Sprintf("  %s: $%s (%.1f%%)\n",
				item.LineItem,
				rb.formatNumber(item.ImpactAmount),
				item.ImpactPercentage))
			if item.Description != "" {
				b.WriteString(fmt.Sprintf("    %s\n", item.Description))
			}
		}
		b.WriteString("\n")
	}

	if len(balanceItems) > 0 {
		b.WriteString("Balance Sheet:\n")
		for _, item := range balanceItems {
			b.WriteString(fmt.Sprintf("  %s: $%s (%.1f%%)\n",
				item.LineItem,
				rb.formatNumber(item.ImpactAmount),
				item.ImpactPercentage))
			if item.Description != "" {
				b.WriteString(fmt.Sprintf("    %s\n", item.Description))
			}
		}
		b.WriteString("\n")
	}

	if len(cashFlowItems) > 0 {
		b.WriteString("Cash Flow Statement:\n")
		for _, item := range cashFlowItems {
			b.WriteString(fmt.Sprintf("  %s: $%s (%.1f%%)\n",
				item.LineItem,
				rb.formatNumber(item.ImpactAmount),
				item.ImpactPercentage))
			if item.Description != "" {
				b.WriteString(fmt.Sprintf("    %s\n", item.Description))
			}
		}
		b.WriteString("\n")
	}

	// Severe weather losses
	if len(f.Expenditures.SevereWeatherLosses) > 0 {
		b.WriteString("Severe Weather Events:\n")
		for _, loss := range f.Expenditures.SevereWeatherLosses {
			b.WriteString(fmt.Sprintf("  %s - %s (%s)\n",
				loss.EventType,
				loss.Location,
				loss.EventDate.Format("January 2006")))
			b.WriteString(fmt.Sprintf("    Total Loss: $%s\n", rb.formatNumber(loss.TotalLoss)))
			if loss.InsuredPortion > 0 {
				b.WriteString(fmt.Sprintf("    Insured: $%s, Uninsured: $%s\n",
					rb.formatNumber(loss.InsuredPortion),
					rb.formatNumber(loss.UninsuredPortion)))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	return b.String()
}

// =============================================================================
// Attestation Report
// =============================================================================

func (rb *ReportBuilder) buildAttestationReport(a *AttestationReport) string {
	if a == nil {
		return "Attestation Report\n\nNot required for this filer type.\n\n"
	}

	var b strings.Builder

	b.WriteString("Third-Party Attestation Report\n")
	b.WriteString(strings.Repeat("-", 80))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Assurance Provider: %s\n", a.Provider))
	b.WriteString(fmt.Sprintf("Assurance Level: %s\n", cases.Title(language.English).String(a.AssuranceLevel)))
	b.WriteString(fmt.Sprintf("Standard: %s\n", a.Standard))
	b.WriteString(fmt.Sprintf("Opinion: %s\n", cases.Title(language.English).String(a.OpinionType)))
	b.WriteString(fmt.Sprintf("Scopes Covered: %s\n", strings.Join(a.ScopesCovered, ", ")))
	b.WriteString(fmt.Sprintf("Report Date: %s\n\n", a.ReportDate.Format("January 2, 2006")))

	b.WriteString(fmt.Sprintf("%s\n\n", a.OpinionStatement))

	if len(a.MaterialWeaknesses) > 0 {
		b.WriteString("Material Weaknesses Identified:\n")
		for i, weakness := range a.MaterialWeaknesses {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, weakness))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	return b.String()
}

// =============================================================================
// Validation Summary
// =============================================================================

func (rb *ReportBuilder) buildValidationSummary(vr *ValidationResults) string {
	if vr == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString("Validation Summary\n")
	b.WriteString(strings.Repeat("-", 80))
	b.WriteString("\n\n")

	if vr.Valid {
		b.WriteString("✓ Report passes all validation checks\n\n")
	} else {
		b.WriteString(fmt.Sprintf("✗ Report has %d validation error(s)\n\n", len(vr.Errors)))
	}

	if len(vr.Errors) > 0 {
		b.WriteString("Errors:\n")
		for i, err := range vr.Errors {
			b.WriteString(fmt.Sprintf("%d. [%s] %s: %s\n",
				i+1, err.Code, err.Field, err.Message))
		}
		b.WriteString("\n")
	}

	if len(vr.Warnings) > 0 {
		b.WriteString("Warnings:\n")
		for i, warn := range vr.Warnings {
			b.WriteString(fmt.Sprintf("%d. [%s] %s: %s\n",
				i+1, warn.Code, warn.Field, warn.Message))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// =============================================================================
// Helper Functions
// =============================================================================

func (rb *ReportBuilder) getFilerStatusName(filerType string) string {
	names := map[string]string{
		"LAF": "Large Accelerated Filer",
		"AF":  "Accelerated Filer",
		"SRC": "Smaller Reporting Company",
		"EGC": "Emerging Growth Company",
	}
	if name, exists := names[filerType]; exists {
		return name
	}
	return filerType
}

func (rb *ReportBuilder) formatNumber(n float64) string {
	return fmt.Sprintf("%.0f", n)
}

func (rb *ReportBuilder) formatTargetValue(value float64, unit string) string {
	if strings.Contains(unit, "%") {
		return fmt.Sprintf("%.1f%%", value)
	}
	return fmt.Sprintf("%.0f %s", value, unit)
}
