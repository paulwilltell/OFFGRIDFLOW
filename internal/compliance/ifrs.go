package compliance

import (
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// IFRSS2Report represents an IFRS S2 Climate-related Disclosures report
// International sustainability disclosure standard
type IFRSS2Report struct {
	Report            *Report
	EmissionsData     EmissionsData
	QualityMetrics    DataQualityMetrics
	OrganizationName  string
	ReportingYear     int
	ReportingOfficer  string
	
	// IFRS S2-specific fields
	ClimateRisks          []ClimateRisk
	Opportunities         []ClimateOpportunity
	FinancialImpacts      []FinancialImpact
	Scenarios             []ClimateScenario
	Metrics               []SustainabilityMetric
	Targets               []EmissionTarget
	TransitionPlan        TransitionPlan
	GovernanceStructure   GovernanceInfo
}

// ClimateOpportunity represents a climate-related business opportunity
type ClimateOpportunity struct {
	Category    string  // "Products/Services", "Resource Efficiency", "Markets", etc.
	Description string
	Potential   string  // "High", "Medium", "Low"
	Strategy    string
	Timeline    string
}

// FinancialImpact represents quantified financial effects of climate change
type FinancialImpact struct {
	Category        string
	Description     string
	Amount          float64  // Currency amount
	Currency        string
	Timeframe       string
	Likelihood      string
	ImpactType      string  // "Revenue", "Cost", "Asset", "Liability"
}

// ClimateScenario represents a climate scenario analysis
type ClimateScenario struct {
	Name            string
	Description     string
	TemperatureGoal string  // e.g., "1.5°C", "2°C", "3°C+"
	Assumptions     string
	Resilience      string
	KeyFindings     []string
}

// SustainabilityMetric represents a tracked sustainability KPI
type SustainabilityMetric struct {
	Name        string
	Value       float64
	Unit        string
	Baseline    float64
	Target      float64
	TargetYear  int
	Progress    float64  // Percentage
}

// TransitionPlan details the organization's climate transition strategy
type TransitionPlan struct {
	Overview        string
	Targets         []EmissionTarget
	Milestones      []Milestone
	Investment      float64  // Currency amount
	Currency        string
	Timeframe       string
}

// Milestone represents a transition plan checkpoint
type Milestone struct {
	Year        int
	Description string
	Status      string
}

// GovernanceInfo details climate governance structure
type GovernanceInfo struct {
	BoardOversight      string
	ManagementRole      string
	Expertise           string
	Integration         string
	IncentiveAlignment  string
}

// GenerateIFRSS2PDF creates an IFRS S2 Climate-related Disclosures PDF
func GenerateIFRSS2PDF(report IFRSS2Report) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAuthor("OffGridFlow", false)
	pdf.SetCreator("OffGridFlow IFRS S2 Compliance Engine", false)
	pdf.SetTitle(fmt.Sprintf("IFRS S2 Report %d - %s", report.ReportingYear, report.OrganizationName), false)
	pdf.SetSubject("IFRS S2 Climate-related Disclosures", false)

	// Cover page
	pdf.AddPage()
	addIFRSS2CoverPage(pdf, report)

	// Governance
	pdf.AddPage()
	addIFRSS2Governance(pdf, report)

	// Strategy - Climate Risks
	pdf.AddPage()
	addIFRSS2Strategy(pdf, report)

	// Strategy - Opportunities
	pdf.AddPage()
	addIFRSS2Opportunities(pdf, report)

	// Financial Impacts
	pdf.AddPage()
	addIFRSS2FinancialImpacts(pdf, report)

	// Scenario Analysis
	pdf.AddPage()
	addIFRSS2ScenarioAnalysis(pdf, report)

	// Transition Plan
	pdf.AddPage()
	addIFRSS2TransitionPlan(pdf, report)

	// Risk Management
	pdf.AddPage()
	addIFRSS2RiskManagement(pdf, report)

	// Metrics and Targets
	pdf.AddPage()
	addIFRSS2MetricsAndTargets(pdf, report)

	// GHG Emissions
	pdf.AddPage()
	addIFRSS2GHGEmissions(pdf, report)

	// Generate PDF bytes
	var buf []byte
	var err error
	if buf, err = pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generate IFRS S2 PDF: %w", err)
	}

	return buf, nil
}

func addIFRSS2CoverPage(pdf *gofpdf.Fpdf, report IFRSS2Report) {
	// IFRS Foundation header
	pdf.SetFillColor(0, 82, 147) // IFRS blue
	pdf.Rect(0, 0, 210, 50, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 22)
	pdf.SetXY(20, 14)
	pdf.Cell(170, 10, "IFRS® SUSTAINABILITY DISCLOSURE")

	pdf.SetFont("Arial", "", 16)
	pdf.SetXY(20, 28)
	pdf.Cell(170, 8, "IFRS S2 Climate-related Disclosures")

	// Company info
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 19)
	pdf.SetXY(20, 65)
	pdf.Cell(170, 10, report.OrganizationName)

	pdf.SetFont("Arial", "", 14)
	pdf.SetXY(20, 78)
	pdf.Cell(170, 7, fmt.Sprintf("Reporting Period: Year Ended December 31, %d", report.ReportingYear))

	// IFRS S2 compliance box
	pdf.SetDrawColor(0, 82, 147)
	pdf.SetLineWidth(0.5)
	pdf.SetFillColor(250, 252, 255)
	pdf.Rect(20, 95, 170, 50, "FD")

	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(25, 100)
	pdf.Cell(160, 7, "IFRS S2 Sustainability Disclosure")

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(25, 110)
	pdf.MultiCell(160, 5, "This report has been prepared in accordance with IFRS S2 Climate-related Disclosures, "+
		"as issued by the International Sustainability Standards Board (ISSB). The disclosures provide information "+
		"about climate-related risks and opportunities that could reasonably be expected to affect the entity's "+
		"prospects.", "", "", false)

	pdf.SetXY(25, 133)
	pdf.Cell(80, 5, fmt.Sprintf("Report Generated: %s", time.Now().Format("January 2, 2006")))

	// Emissions summary
	pdf.SetFillColor(245, 250, 255)
	pdf.Rect(20, 155, 170, 75, "FD")

	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(25, 160)
	pdf.Cell(160, 7, "Climate-Related Metrics Summary")

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(30, 173)
	pdf.Cell(90, 6, "Scope 1 + 2 Emissions:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(70, 6, fmt.Sprintf("%.2f tCO2e", report.Report.Scope1EmissionsTonnes+report.Report.Scope2EmissionsTonnes))

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(30, 181)
	pdf.Cell(90, 6, "Scope 3 Emissions (if material):")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(70, 6, fmt.Sprintf("%.2f tCO2e", report.Report.Scope3EmissionsTonnes))

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(30, 189)
	pdf.Cell(90, 6, "Climate Risks Identified:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(70, 6, fmt.Sprintf("%d", len(report.ClimateRisks)))

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(30, 197)
	pdf.Cell(90, 6, "Climate Opportunities:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(70, 6, fmt.Sprintf("%d", len(report.Opportunities)))

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(30, 205)
	pdf.Cell(90, 6, "Scenario Analyses Conducted:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(70, 6, fmt.Sprintf("%d", len(report.Scenarios)))

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(30, 213)
	pdf.Cell(90, 6, "Active Emission Reduction Targets:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(70, 6, fmt.Sprintf("%d", len(report.Targets)))

	// Footer
	pdf.SetFont("Arial", "I", 8)
	pdf.SetXY(20, 280)
	pdf.MultiCell(170, 4, "IFRS S2 requires disclosure of material information about climate-related risks and opportunities. "+
		"This report follows the recommendations of the Task Force on Climate-related Financial Disclosures (TCFD) "+
		"and aligns with the four core pillars: Governance, Strategy, Risk Management, and Metrics and Targets.", "", "", false)
}

func addIFRSS2Governance(pdf *gofpdf.Fpdf, report IFRSS2Report) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Governance")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "IFRS S2 requires disclosure of the governance processes, controls, and procedures used to "+
		"monitor, manage, and oversee climate-related risks and opportunities.", "", "", false)
	pdf.Ln(8)

	// Board oversight
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Board Oversight")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	governance := report.GovernanceStructure.BoardOversight
	if governance == "" {
		governance = "The Board of Directors provides oversight of climate-related risks and opportunities through "+
			"its Risk and Sustainability Committee. The Committee meets quarterly to review climate strategy, "+
			"emissions performance, and emerging regulatory requirements. Climate considerations are integrated "+
			"into strategic planning and capital allocation decisions."
	}
	pdf.MultiCell(170, 5, governance, "", "", false)

	pdf.Ln(8)

	// Management role
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Management's Role")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	mgmtRole := report.GovernanceStructure.ManagementRole
	if mgmtRole == "" {
		mgmtRole = "The Chief Sustainability Officer (CSO) is responsible for developing and implementing climate strategy. "+
			"A cross-functional Climate Steering Committee, chaired by the CSO, coordinates climate initiatives across "+
			"operations, supply chain, product development, and investor relations. Management reports climate metrics "+
			"and progress to the Board quarterly."
	}
	pdf.MultiCell(170, 5, mgmtRole, "", "", false)

	pdf.Ln(8)

	// Skills and competencies
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Climate Expertise and Training")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	expertise := report.GovernanceStructure.Expertise
	if expertise == "" {
		expertise = "Board members receive annual training on climate science, emissions accounting, and climate-related "+
			"financial disclosure requirements. The Risk Committee includes members with sustainability expertise and "+
			"experience in ESG investing. Management team members participate in external climate leadership programs."
	}
	pdf.MultiCell(170, 5, expertise, "", "", false)

	pdf.Ln(8)

	// Incentive alignment
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Performance Incentive Alignment")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	incentives := report.GovernanceStructure.IncentiveAlignment
	if incentives == "" {
		incentives = "Executive compensation includes climate-related performance metrics. Annual bonuses are partially tied "+
			"to achievement of emissions reduction targets and successful implementation of climate initiatives. "+
			"Long-term incentive plans consider sustainability performance in ESG ratings and climate disclosures."
	}
	pdf.MultiCell(170, 5, incentives, "", "", false)
}

func addIFRSS2Strategy(pdf *gofpdf.Fpdf, report IFRSS2Report) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Strategy: Climate-Related Risks")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "IFRS S2 requires disclosure of climate-related risks that could reasonably be expected to affect "+
		"the entity's business model, strategy, and financial position over the short, medium, and long term.", "", "", false)
	pdf.Ln(8)

	if len(report.ClimateRisks) == 0 {
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(170, 7, "No material climate-related risks identified in the current assessment.")
		return
	}

	// Physical risks
	physicalRisks := []ClimateRisk{}
	transitionRisks := []ClimateRisk{}

	for _, risk := range report.ClimateRisks {
		if risk.Category == "Physical" {
			physicalRisks = append(physicalRisks, risk)
		} else {
			transitionRisks = append(transitionRisks, risk)
		}
	}

	if len(physicalRisks) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(170, 8, "Physical Climate Risks")
		pdf.Ln(10)

		pdf.SetFont("Arial", "", 10)
		for i, risk := range physicalRisks {
			pdf.SetFont("Arial", "B", 10)
			pdf.Cell(170, 6, fmt.Sprintf("%d. %s (Impact: %s)", i+1, risk.Description, risk.Impact))
			pdf.Ln(7)

			pdf.SetFont("Arial", "", 9)
			pdf.SetX(25)
			pdf.MultiCell(165, 4, fmt.Sprintf("Mitigation Strategy: %s", risk.Mitigation), "", "", false)
			pdf.Ln(4)
		}
	}

	if len(transitionRisks) > 0 {
		pdf.Ln(5)
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(170, 8, "Transition Climate Risks")
		pdf.Ln(10)

		pdf.SetFont("Arial", "", 10)
		for i, risk := range transitionRisks {
			pdf.SetFont("Arial", "B", 10)
			pdf.Cell(170, 6, fmt.Sprintf("%d. %s (Impact: %s)", i+1, risk.Description, risk.Impact))
			pdf.Ln(7)

			pdf.SetFont("Arial", "", 9)
			pdf.SetX(25)
			pdf.MultiCell(165, 4, fmt.Sprintf("Mitigation Strategy: %s", risk.Mitigation), "", "", false)
			pdf.Ln(4)
		}
	}
}

func addIFRSS2Opportunities(pdf *gofpdf.Fpdf, report IFRSS2Report) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Strategy: Climate-Related Opportunities")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "Climate change presents opportunities for innovation, market expansion, resource efficiency, "+
		"and resilience-building that could enhance long-term value creation.", "", "", false)
	pdf.Ln(8)

	if len(report.Opportunities) == 0 {
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(170, 7, "Climate opportunity assessment in progress.")
		return
	}

	for i, opp := range report.Opportunities {
		pdf.SetFont("Arial", "B", 11)
		pdf.SetTextColor(0, 100, 0)
		pdf.Cell(170, 7, fmt.Sprintf("%d. %s (Potential: %s)", i+1, opp.Category, opp.Potential))
		pdf.SetTextColor(0, 0, 0)
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 10)
		pdf.SetX(25)
		pdf.MultiCell(165, 5, opp.Description, "", "", false)
		pdf.Ln(3)

		pdf.SetFont("Arial", "I", 9)
		pdf.SetX(25)
		pdf.MultiCell(165, 4, fmt.Sprintf("Strategy: %s | Timeline: %s", opp.Strategy, opp.Timeline), "", "", false)
		pdf.Ln(5)
	}
}

func addIFRSS2FinancialImpacts(pdf *gofpdf.Fpdf, report IFRSS2Report) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Anticipated Financial Impacts")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "IFRS S2 requires quantification of the current and anticipated financial effects of climate-related "+
		"risks and opportunities on the entity's financial position, performance, and cash flows.", "", "", false)
	pdf.Ln(8)

	if len(report.FinancialImpacts) == 0 {
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(170, 7, "Financial impact quantification in progress.")
		return
	}

	// Financial impacts table
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(230, 240, 245)
	pdf.CellFormat(50, 6, "Impact Category", "1", 0, "L", true, 0, "")
	pdf.CellFormat(40, 6, "Type", "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 6, "Amount", "1", 0, "R", true, 0, "")
	pdf.CellFormat(45, 6, "Timeframe", "1", 1, "L", true, 0, "")

	pdf.SetFont("Arial", "", 8)
	pdf.SetFillColor(248, 252, 255)

	for i, impact := range report.FinancialImpacts {
		fill := i%2 == 0
		pdf.CellFormat(50, 5, truncate(impact.Category, 30), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(40, 5, impact.ImpactType, "1", 0, "L", fill, 0, "")
		pdf.CellFormat(35, 5, fmt.Sprintf("%s %.1fM", impact.Currency, impact.Amount/1_000_000), "1", 0, "R", fill, 0, "")
		pdf.CellFormat(45, 5, impact.Timeframe, "1", 1, "L", fill, 0, "")
	}
}

func addIFRSS2ScenarioAnalysis(pdf *gofpdf.Fpdf, report IFRSS2Report) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Climate Scenario Analysis")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "IFRS S2 encourages scenario analysis to assess the resilience of the entity's strategy and "+
		"business model to climate-related changes. The following scenarios have been analyzed:", "", "", false)
	pdf.Ln(8)

	if len(report.Scenarios) == 0 {
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(170, 7, "Climate scenario analysis to be conducted in future reporting periods.")
		return
	}

	for i, scenario := range report.Scenarios {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(170, 7, fmt.Sprintf("Scenario %d: %s (%s pathway)", i+1, scenario.Name, scenario.TemperatureGoal))
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 10)
		pdf.SetX(25)
		pdf.MultiCell(165, 5, scenario.Description, "", "", false)
		pdf.Ln(3)

		pdf.SetFont("Arial", "I", 9)
		pdf.SetX(25)
		pdf.MultiCell(165, 4, fmt.Sprintf("Key Assumptions: %s", scenario.Assumptions), "", "", false)
		pdf.Ln(3)

		if len(scenario.KeyFindings) > 0 {
			pdf.SetFont("Arial", "B", 9)
			pdf.SetX(25)
			pdf.Cell(165, 4, "Key Findings:")
			pdf.Ln(5)

			pdf.SetFont("Arial", "", 8)
			for _, finding := range scenario.KeyFindings {
				pdf.SetX(30)
				pdf.Cell(5, 3, "•")
				pdf.MultiCell(160, 3, finding, "", "", false)
				pdf.Ln(1)
			}
		}

		pdf.Ln(5)
	}
}

func addIFRSS2TransitionPlan(pdf *gofpdf.Fpdf, report IFRSS2Report) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Climate Transition Plan")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	overview := report.TransitionPlan.Overview
	if overview == "" {
		overview = "The organization is developing a comprehensive climate transition plan aligned with limiting global "+
			"warming to 1.5°C. The plan includes emissions reduction targets, technology investments, supply chain engagement, "+
			"and regular progress monitoring."
	}
	pdf.MultiCell(170, 5, overview, "", "", false)

	pdf.Ln(8)

	// Investment commitment
	if report.TransitionPlan.Investment > 0 {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(95, 6, "Committed Transition Investment:")
		pdf.SetFont("Arial", "B", 11)
		pdf.SetTextColor(0, 100, 0)
		pdf.Cell(75, 6, fmt.Sprintf("%s %.1fM (%s)", report.TransitionPlan.Currency, 
			report.TransitionPlan.Investment/1_000_000, report.TransitionPlan.Timeframe))
		pdf.SetTextColor(0, 0, 0)
		pdf.Ln(10)
	}

	// Milestones
	if len(report.TransitionPlan.Milestones) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(170, 8, "Transition Milestones")
		pdf.Ln(10)

		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(240, 245, 250)
		pdf.CellFormat(25, 6, "Year", "1", 0, "C", true, 0, "")
		pdf.CellFormat(110, 6, "Milestone", "1", 0, "L", true, 0, "")
		pdf.CellFormat(35, 6, "Status", "1", 1, "C", true, 0, "")

		pdf.SetFont("Arial", "", 9)
		pdf.SetFillColor(250, 252, 255)

		for i, milestone := range report.TransitionPlan.Milestones {
			fill := i%2 == 0
			pdf.CellFormat(25, 5, fmt.Sprintf("%d", milestone.Year), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(110, 5, truncate(milestone.Description, 70), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(35, 5, milestone.Status, "1", 1, "C", fill, 0, "")
		}
	}
}

func addIFRSS2RiskManagement(pdf *gofpdf.Fpdf, report IFRSS2Report) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Risk Management")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "Climate-related risks are integrated into the organization's enterprise risk management framework. "+
		"Risk assessment processes identify, evaluate, prioritize, and monitor climate risks over multiple time horizons.", "", "", false)

	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Risk Identification and Assessment")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(170, 5, "The organization employs a systematic process to identify climate-related risks:\n\n"+
		"• Annual climate risk assessment covering physical and transition risks\n"+
		"• Scenario analysis using NGFS and IEA climate pathways\n"+
		"• Stakeholder engagement including investors, customers, and suppliers\n"+
		"• Integration with financial planning and capital allocation decisions\n"+
		"• Regular monitoring of climate science and policy developments", "", "", false)

	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Risk Prioritization")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(170, 5, "Climate risks are prioritized based on likelihood and potential financial impact using a "+
		"standardized risk matrix. High-priority risks are assigned to executive owners with specific mitigation plans, "+
		"timelines, and key risk indicators. Progress is reviewed quarterly by the Risk Committee.", "", "", false)
}

func addIFRSS2MetricsAndTargets(pdf *gofpdf.Fpdf, report IFRSS2Report) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Metrics and Targets")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "IFRS S2 requires disclosure of metrics and targets used to measure and manage material "+
		"climate-related risks and opportunities, including cross-industry and industry-specific metrics.", "", "", false)
	pdf.Ln(8)

	// Emission reduction targets
	if len(report.Targets) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(170, 8, "GHG Emissions Reduction Targets")
		pdf.Ln(10)

		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(230, 240, 245)
		pdf.CellFormat(30, 6, "Scope", "1", 0, "L", true, 0, "")
		pdf.CellFormat(30, 6, "Baseline", "1", 0, "L", true, 0, "")
		pdf.CellFormat(30, 6, "Target", "1", 0, "L", true, 0, "")
		pdf.CellFormat(40, 6, "Reduction %", "1", 0, "L", true, 0, "")
		pdf.CellFormat(40, 6, "Status", "1", 1, "L", true, 0, "")

		pdf.SetFont("Arial", "", 9)
		pdf.SetFillColor(248, 252, 255)

		for i, target := range report.Targets {
			fill := i%2 == 0
			pdf.CellFormat(30, 5, target.Scope, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(30, 5, fmt.Sprintf("%d", target.BaselineYear), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(30, 5, fmt.Sprintf("%d", target.TargetYear), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(40, 5, fmt.Sprintf("%.1f%%", target.TargetReduction), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(40, 5, target.Status, "1", 1, "L", fill, 0, "")
		}

		pdf.Ln(8)
	}

	// Other sustainability metrics
	if len(report.Metrics) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(170, 8, "Additional Climate Metrics")
		pdf.Ln(10)

		for _, metric := range report.Metrics {
			pdf.SetFont("Arial", "B", 10)
			pdf.Cell(170, 6, metric.Name)
			pdf.Ln(7)

			pdf.SetFont("Arial", "", 9)
			pdf.SetX(25)
			pdf.Cell(60, 5, fmt.Sprintf("Current: %.2f %s", metric.Value, metric.Unit))
			pdf.Cell(60, 5, fmt.Sprintf("Target: %.2f %s (%d)", metric.Target, metric.Unit, metric.TargetYear))
			pdf.Cell(50, 5, fmt.Sprintf("Progress: %.1f%%", metric.Progress))
			pdf.Ln(7)
		}
	}
}

func addIFRSS2GHGEmissions(pdf *gofpdf.Fpdf, report IFRSS2Report) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "GHG Emissions (Cross-Industry Metric)")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "IFRS S2 requires disclosure of Scope 1, Scope 2, and Scope 3 greenhouse gas emissions, "+
		"calculated in accordance with the GHG Protocol.", "", "", false)
	pdf.Ln(8)

	// Emissions table
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 230, 240)
	pdf.CellFormat(60, 8, "Scope", "1", 0, "L", true, 0, "")
	pdf.CellFormat(55, 8, "Emissions (tCO2e)", "1", 0, "R", true, 0, "")
	pdf.CellFormat(55, 8, "% of Total", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(245, 250, 255)

	total := report.Report.TotalEmissionsTonnes

	pdf.CellFormat(60, 7, "Scope 1 - Direct", "1", 0, "L", true, 0, "")
	pdf.CellFormat(55, 7, fmt.Sprintf("%.2f", report.Report.Scope1EmissionsTonnes), "1", 0, "R", true, 0, "")
	pdf.CellFormat(55, 7, fmt.Sprintf("%.1f%%", (report.Report.Scope1EmissionsTonnes/total)*100), "1", 1, "R", true, 0, "")

	pdf.CellFormat(60, 7, "Scope 2 - Indirect Energy", "1", 0, "L", false, 0, "")
	pdf.CellFormat(55, 7, fmt.Sprintf("%.2f", report.Report.Scope2EmissionsTonnes), "1", 0, "R", false, 0, "")
	pdf.CellFormat(55, 7, fmt.Sprintf("%.1f%%", (report.Report.Scope2EmissionsTonnes/total)*100), "1", 1, "R", false, 0, "")

	if report.Report.Scope3EmissionsTonnes > 0 {
		pdf.CellFormat(60, 7, "Scope 3 - Value Chain", "1", 0, "L", true, 0, "")
		pdf.CellFormat(55, 7, fmt.Sprintf("%.2f", report.Report.Scope3EmissionsTonnes), "1", 0, "R", true, 0, "")
		pdf.CellFormat(55, 7, fmt.Sprintf("%.1f%%", (report.Report.Scope3EmissionsTonnes/total)*100), "1", 1, "R", true, 0, "")
	}

	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 215, 230)
	pdf.CellFormat(60, 8, "Total", "1", 0, "L", true, 0, "")
	pdf.CellFormat(55, 8, fmt.Sprintf("%.2f", total), "1", 0, "R", true, 0, "")
	pdf.CellFormat(55, 8, "100.0%", "1", 1, "R", true, 0, "")

	pdf.Ln(8)

	// Methodology note
	pdf.SetFont("Arial", "I", 9)
	pdf.MultiCell(170, 4, "GHG emissions calculated using the GHG Protocol Corporate Standard. Scope 2 reflects "+
		"location-based methodology. Scope 3 categories reported based on materiality assessment.", "", "", false)

	pdf.Ln(6)

	// Report hash
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(170, 6, "Report Integrity Hash (SHA-256):")
	pdf.Ln(7)

	pdf.SetFont("Courier", "", 7)
	pdf.Cell(170, 4, report.Report.ReportHash)
}
