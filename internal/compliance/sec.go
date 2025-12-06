package compliance

import (
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

// SECReport represents an SEC Climate Disclosure report
// Complies with SEC's climate-related disclosure requirements
type SECReport struct {
	Report            *Report
	EmissionsData     EmissionsData
	QualityMetrics    DataQualityMetrics
	OrganizationName  string
	FiscalYear        int
	ReportingOfficer  string
	
	// SEC-specific fields
	ClimateRisks      []ClimateRisk
	TransitionPlan    string
	GovernanceStructure string
	Targets           []EmissionTarget
}

// ClimateRisk represents a climate-related risk disclosure
type ClimateRisk struct {
	Category    string  // "Physical" or "Transition"
	Description string
	Impact      string  // "High", "Medium", "Low"
	Mitigation  string
}

// EmissionTarget represents a GHG reduction target
type EmissionTarget struct {
	Scope       string
	BaselineYear int
	BaselineTonnes float64
	TargetYear  int
	TargetReduction float64 // Percentage
	Status      string
}

// GenerateSECPDF creates a production-ready SEC Climate Disclosure PDF
func GenerateSECPDF(report SECReport) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAuthor("OffGridFlow", false)
	pdf.SetCreator("OffGridFlow SEC Compliance Engine", false)
	pdf.SetTitle(fmt.Sprintf("SEC Climate Disclosure %d - %s", report.FiscalYear, report.OrganizationName), false)
	pdf.SetSubject("SEC Climate-Related Disclosures", false)

	// Cover page
	pdf.AddPage()
	addSECCoverPage(pdf, report)

	// Item 1: Governance
	pdf.AddPage()
	addSECGovernance(pdf, report)

	// Item 2: Strategy - Climate Risks
	pdf.AddPage()
	addSECClimateRisks(pdf, report)

	// Item 3: Risk Management
	pdf.AddPage()
	addSECRiskManagement(pdf, report)

	// Item 4: Metrics and Targets
	pdf.AddPage()
	addSECMetricsAndTargets(pdf, report)

	// GHG Emissions Data (Scopes 1, 2, 3)
	pdf.AddPage()
	addSECEmissionsData(pdf, report)

	// Scope 1 Breakdown
	pdf.AddPage()
	addSECScope1Breakdown(pdf, report)

	// Scope 2 Breakdown
	pdf.AddPage()
	addSECScope2Breakdown(pdf, report)

	// Scope 3 Breakdown (if material)
	if report.Report.Scope3EmissionsTonnes > 0 {
		pdf.AddPage()
		addSECScope3Breakdown(pdf, report)
	}

	// Data Quality and Assurance
	pdf.AddPage()
	addSECDataQualityAssurance(pdf, report)

	// Generate PDF bytes
	var buf []byte
	var err error
	if buf, err = pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generate SEC PDF: %w", err)
	}

	return buf, nil
}

func addSECCoverPage(pdf *gofpdf.Fpdf, report SECReport) {
	// SEC Header
	pdf.SetFillColor(17, 34, 51) // SEC dark blue
	pdf.Rect(0, 0, 210, 45, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 22)
	pdf.SetXY(20, 15)
	pdf.Cell(170, 10, "SECURITIES AND EXCHANGE COMMISSION")

	pdf.SetFont("Arial", "", 16)
	pdf.SetXY(20, 27)
	pdf.Cell(170, 10, "Climate-Related Disclosures")

	// Company info
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 20)
	pdf.SetXY(20, 60)
	pdf.Cell(170, 10, report.OrganizationName)

	pdf.SetFont("Arial", "", 14)
	pdf.SetXY(20, 72)
	pdf.Cell(170, 8, fmt.Sprintf("Fiscal Year Ended: December 31, %d", report.FiscalYear))

	// Report identifier box
	pdf.SetDrawColor(100, 100, 100)
	pdf.SetFillColor(250, 250, 250)
	pdf.Rect(20, 90, 170, 50, "FD")

	pdf.SetFont("Arial", "B", 11)
	pdf.SetXY(25, 95)
	pdf.Cell(80, 6, "Climate Disclosure Report")

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(25, 103)
	pdf.Cell(80, 5, fmt.Sprintf("Report ID: %s", report.Report.ID.String()[:13]))

	pdf.SetXY(25, 110)
	pdf.Cell(80, 5, fmt.Sprintf("Generated: %s", report.Report.GenerationTimestamp.Format("January 2, 2006")))

	pdf.SetXY(25, 117)
	pdf.Cell(80, 5, fmt.Sprintf("Fiscal Year: %d", report.FiscalYear))

	pdf.SetXY(25, 124)
	pdf.Cell(80, 5, fmt.Sprintf("Version: %d", report.Report.Version))

	pdf.SetXY(25, 131)
	pdf.Cell(80, 5, fmt.Sprintf("Status: %s", report.Report.Status))

	// GHG Emissions Summary box
	pdf.SetFillColor(245, 245, 250)
	pdf.Rect(20, 150, 170, 70, "FD")

	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(25, 155)
	pdf.Cell(80, 7, "Greenhouse Gas Emissions Summary")

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(25, 168)
	pdf.Cell(100, 6, "Scope 1 (Direct Emissions):")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(60, 6, fmt.Sprintf("%.2f tCO2e", report.Report.Scope1EmissionsTonnes))

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(25, 176)
	pdf.Cell(100, 6, "Scope 2 (Indirect - Electricity):")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(60, 6, fmt.Sprintf("%.2f tCO2e", report.Report.Scope2EmissionsTonnes))

	if report.Report.Scope3EmissionsTonnes > 0 {
		pdf.SetFont("Arial", "", 10)
		pdf.SetXY(25, 184)
		pdf.Cell(100, 6, "Scope 3 (Indirect - Value Chain):")
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(60, 6, fmt.Sprintf("%.2f tCO2e", report.Report.Scope3EmissionsTonnes))

		pdf.SetFont("Arial", "B", 11)
		pdf.SetXY(25, 196)
		pdf.Cell(100, 7, "Total GHG Emissions:")
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(60, 7, fmt.Sprintf("%.2f tCO2e", report.Report.TotalEmissionsTonnes))
	} else {
		pdf.SetFont("Arial", "B", 11)
		pdf.SetXY(25, 188)
		pdf.Cell(100, 7, "Total Scopes 1+2 Emissions:")
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(60, 7, fmt.Sprintf("%.2f tCO2e", report.Report.Scope1EmissionsTonnes+report.Report.Scope2EmissionsTonnes))
	}

	// Data quality indicator
	pdf.SetFont("Arial", "I", 9)
	pdf.SetXY(25, 208)
	pdf.Cell(160, 5, fmt.Sprintf("Data Quality Score: %.1f%% | Completeness: %.1f%%",
		report.QualityMetrics.DataQualityScore, report.QualityMetrics.CompletenessPercentage))

	// Footer
	pdf.SetFont("Arial", "I", 8)
	pdf.SetXY(20, 280)
	pdf.MultiCell(170, 4, "This report has been prepared in accordance with SEC climate-related disclosure requirements. "+
		"The information contained herein is based on the best available data and reasonable estimation methodologies.", "", "", false)
}

func addSECGovernance(pdf *gofpdf.Fpdf, report SECReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Item 1: Governance")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Board Oversight of Climate-Related Risks")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	governance := report.GovernanceStructure
	if governance == "" {
		governance = "The Board of Directors oversees climate-related risks and opportunities through its Risk Committee. " +
			"The Risk Committee meets quarterly to review climate risk assessments, emissions performance, and progress " +
			"toward climate-related goals. Management provides regular updates on climate strategy implementation, " +
			"regulatory developments, and stakeholder engagement on climate matters."
	}
	pdf.MultiCell(170, 5, governance, "", "", false)
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Management's Role in Climate Risk Assessment")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "Management is responsible for identifying, assessing, and managing climate-related risks and opportunities. "+
		"The Chief Sustainability Officer reports directly to the CEO and coordinates climate strategy across all business units. "+
		"Cross-functional teams assess climate risks in operations, supply chain, and capital allocation decisions.", "", "", false)

	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Climate Expertise and Resources")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "The organization maintains dedicated climate expertise through sustainability professionals, "+
		"engages third-party consultants for specialized assessments, and provides ongoing training to board members "+
		"and management on climate-related matters.", "", "", false)
}

func addSECClimateRisks(pdf *gofpdf.Fpdf, report SECReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Item 2: Strategy - Climate-Related Risks and Opportunities")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "The following climate-related risks have been identified as potentially material to the organization's "+
		"business, operations, and financial performance:", "", "", false)
	pdf.Ln(8)

	if len(report.ClimateRisks) > 0 {
		// Physical Risks
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

			for _, risk := range physicalRisks {
				pdf.SetFont("Arial", "B", 10)
				pdf.Cell(170, 6, fmt.Sprintf("• %s (Impact: %s)", risk.Description, risk.Impact))
				pdf.Ln(7)

				pdf.SetFont("Arial", "", 9)
				pdf.SetX(25)
				pdf.MultiCell(165, 4, fmt.Sprintf("Mitigation: %s", risk.Mitigation), "", "", false)
				pdf.Ln(3)
			}
			pdf.Ln(5)
		}

		if len(transitionRisks) > 0 {
			pdf.SetFont("Arial", "B", 12)
			pdf.Cell(170, 8, "Transition Climate Risks")
			pdf.Ln(10)

			for _, risk := range transitionRisks {
				pdf.SetFont("Arial", "B", 10)
				pdf.Cell(170, 6, fmt.Sprintf("• %s (Impact: %s)", risk.Description, risk.Impact))
				pdf.Ln(7)

				pdf.SetFont("Arial", "", 9)
				pdf.SetX(25)
				pdf.MultiCell(165, 4, fmt.Sprintf("Mitigation: %s", risk.Mitigation), "", "", false)
				pdf.Ln(3)
			}
		}
	} else {
		pdf.SetFont("Arial", "I", 10)
		pdf.Cell(170, 6, "No material climate risks identified in the current assessment period.")
		pdf.Ln(8)
	}

	// Transition Plan
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Climate Transition Plan")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	transitionPlan := report.TransitionPlan
	if transitionPlan == "" {
		transitionPlan = "The organization is developing a comprehensive climate transition plan that includes emissions reduction targets, " +
			"capital allocation for low-carbon technologies, supply chain engagement, and regular progress reporting."
	}
	pdf.MultiCell(170, 5, transitionPlan, "", "", false)
}

func addSECRiskManagement(pdf *gofpdf.Fpdf, report SECReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Item 3: Risk Management")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Process for Identifying Climate-Related Risks")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "The organization employs a structured process to identify and assess climate-related risks:\n\n"+
		"1. Annual climate risk assessment covering physical and transition risks\n"+
		"2. Scenario analysis using multiple climate pathways (e.g., RCP 2.6, RCP 4.5, RCP 8.5)\n"+
		"3. Stakeholder engagement including investors, customers, and suppliers\n"+
		"4. Integration with enterprise risk management framework\n"+
		"5. Regular review and updates based on evolving climate science and policy", "", "", false)

	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Risk Prioritization and Mitigation")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "Identified risks are prioritized based on likelihood and potential financial impact. "+
		"High-priority risks are assigned to responsible management teams with specific mitigation plans, timelines, "+
		"and performance indicators. Progress is tracked quarterly and reported to the Risk Committee.", "", "", false)
}

func addSECMetricsAndTargets(pdf *gofpdf.Fpdf, report SECReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Item 4: Metrics and Targets")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "GHG Emissions Metrics")
	pdf.Ln(10)

	// Emissions intensity table
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(85, 7, "Metric", "1", 0, "L", true, 0, "")
	pdf.CellFormat(85, 7, "Value", "1", 1, "L", true, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(245, 245, 245)

	pdf.CellFormat(85, 6, "Total Scope 1 + 2 Emissions (tCO2e)", "1", 0, "L", true, 0, "")
	pdf.CellFormat(85, 6, fmt.Sprintf("%.2f", report.Report.Scope1EmissionsTonnes+report.Report.Scope2EmissionsTonnes), "1", 1, "L", true, 0, "")

	pdf.CellFormat(85, 6, "Scope 1 Emissions (tCO2e)", "1", 0, "L", false, 0, "")
	pdf.CellFormat(85, 6, fmt.Sprintf("%.2f", report.Report.Scope1EmissionsTonnes), "1", 1, "L", false, 0, "")

	pdf.CellFormat(85, 6, "Scope 2 Emissions (tCO2e)", "1", 0, "L", true, 0, "")
	pdf.CellFormat(85, 6, fmt.Sprintf("%.2f", report.Report.Scope2EmissionsTonnes), "1", 1, "L", true, 0, "")

	if report.Report.Scope3EmissionsTonnes > 0 {
		pdf.CellFormat(85, 6, "Scope 3 Emissions (tCO2e) [if material]", "1", 0, "L", false, 0, "")
		pdf.CellFormat(85, 6, fmt.Sprintf("%.2f", report.Report.Scope3EmissionsTonnes), "1", 1, "L", false, 0, "")
	}

	pdf.Ln(8)

	// Emissions targets
	if len(report.Targets) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(170, 8, "Emissions Reduction Targets")
		pdf.Ln(10)

		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(220, 220, 220)
		pdf.CellFormat(30, 7, "Scope", "1", 0, "L", true, 0, "")
		pdf.CellFormat(35, 7, "Baseline Year", "1", 0, "L", true, 0, "")
		pdf.CellFormat(30, 7, "Target Year", "1", 0, "L", true, 0, "")
		pdf.CellFormat(40, 7, "Reduction %", "1", 0, "L", true, 0, "")
		pdf.CellFormat(35, 7, "Status", "1", 1, "L", true, 0, "")

		pdf.SetFont("Arial", "", 9)
		pdf.SetFillColor(245, 245, 245)

		for i, target := range report.Targets {
			fill := i%2 == 0
			pdf.CellFormat(30, 6, target.Scope, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(35, 6, fmt.Sprintf("%d", target.BaselineYear), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(30, 6, fmt.Sprintf("%d", target.TargetYear), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(40, 6, fmt.Sprintf("%.1f%%", target.TargetReduction), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(35, 6, target.Status, "1", 1, "L", fill, 0, "")
		}
	}
}

func addSECEmissionsData(pdf *gofpdf.Fpdf, report SECReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "GHG Emissions Data")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, fmt.Sprintf("The following table summarizes greenhouse gas emissions for fiscal year %d, "+
		"calculated in accordance with the GHG Protocol Corporate Standard:", report.FiscalYear), "", "", false)
	pdf.Ln(8)

	// Summary table
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(60, 8, "Scope", "1", 0, "L", true, 0, "")
	pdf.CellFormat(55, 8, "Emissions (tCO2e)", "1", 0, "R", true, 0, "")
	pdf.CellFormat(55, 8, "% of Total", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(245, 245, 245)

	total := report.Report.TotalEmissionsTonnes
	if total == 0 {
		total = report.Report.Scope1EmissionsTonnes + report.Report.Scope2EmissionsTonnes + report.Report.Scope3EmissionsTonnes
	}

	pdf.CellFormat(60, 7, "Scope 1", "1", 0, "L", true, 0, "")
	pdf.CellFormat(55, 7, fmt.Sprintf("%.2f", report.Report.Scope1EmissionsTonnes), "1", 0, "R", true, 0, "")
	pdf.CellFormat(55, 7, fmt.Sprintf("%.1f%%", (report.Report.Scope1EmissionsTonnes/total)*100), "1", 1, "R", true, 0, "")

	pdf.CellFormat(60, 7, "Scope 2", "1", 0, "L", false, 0, "")
	pdf.CellFormat(55, 7, fmt.Sprintf("%.2f", report.Report.Scope2EmissionsTonnes), "1", 0, "R", false, 0, "")
	pdf.CellFormat(55, 7, fmt.Sprintf("%.1f%%", (report.Report.Scope2EmissionsTonnes/total)*100), "1", 1, "R", false, 0, "")

	if report.Report.Scope3EmissionsTonnes > 0 {
		pdf.CellFormat(60, 7, "Scope 3", "1", 0, "L", true, 0, "")
		pdf.CellFormat(55, 7, fmt.Sprintf("%.2f", report.Report.Scope3EmissionsTonnes), "1", 0, "R", true, 0, "")
		pdf.CellFormat(55, 7, fmt.Sprintf("%.1f%%", (report.Report.Scope3EmissionsTonnes/total)*100), "1", 1, "R", true, 0, "")
	}

	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(60, 8, "Total", "1", 0, "L", true, 0, "")
	pdf.CellFormat(55, 8, fmt.Sprintf("%.2f", total), "1", 0, "R", true, 0, "")
	pdf.CellFormat(55, 8, "100.0%", "1", 1, "R", true, 0, "")

	pdf.Ln(8)

	// Methodology note
	pdf.SetFont("Arial", "I", 9)
	pdf.MultiCell(170, 4, "Note: Emissions calculated using operational control approach. Scope 2 emissions reflect location-based methodology. "+
		"All emission factors sourced from EPA, IPCC, and regional electricity grid data.", "", "", false)
}

func addSECScope1Breakdown(pdf *gofpdf.Fpdf, report SECReport) {
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(170, 10, "Scope 1 Emissions Breakdown")
	pdf.Ln(12)

	scope1Activities := []ActivityEmission{}
	for _, activity := range report.EmissionsData.Activities {
		if activity.Scope == "scope1" || activity.Scope == "Scope 1" {
			scope1Activities = append(scope1Activities, activity)
		}
	}

	if len(scope1Activities) == 0 {
		pdf.SetFont("Arial", "", 11)
		pdf.Cell(170, 7, "No Scope 1 emissions reported for this period.")
		return
	}

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(170, 6, fmt.Sprintf("Total Scope 1 Activities: %d", len(scope1Activities)))
	pdf.Ln(8)

	// Activity table
	displayLimit := 20
	pdf.SetFont("Arial", "B", 8)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(65, 6, "Activity", "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 6, "Category", "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 6, "Quantity", "1", 0, "R", true, 0, "")
	pdf.CellFormat(35, 6, "Emissions (tCO2e)", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 7)
	pdf.SetFillColor(245, 245, 245)

	for i, activity := range scope1Activities {
		if i >= displayLimit {
			break
		}
		fill := i%2 == 0
		pdf.CellFormat(65, 5, truncate(activity.Name, 40), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(35, 5, truncate(activity.Category, 22), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(35, 5, fmt.Sprintf("%.2f %s", activity.Quantity, activity.Unit), "1", 0, "R", fill, 0, "")
		pdf.CellFormat(35, 5, fmt.Sprintf("%.3f", activity.EmissionsTonnes), "1", 1, "R", fill, 0, "")
	}
}

func addSECScope2Breakdown(pdf *gofpdf.Fpdf, report SECReport) {
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(170, 10, "Scope 2 Emissions Breakdown")
	pdf.Ln(12)

	scope2Activities := []ActivityEmission{}
	for _, activity := range report.EmissionsData.Activities {
		if activity.Scope == "scope2" || activity.Scope == "Scope 2" {
			scope2Activities = append(scope2Activities, activity)
		}
	}

	if len(scope2Activities) == 0 {
		pdf.SetFont("Arial", "", 11)
		pdf.Cell(170, 7, "No Scope 2 emissions reported for this period.")
		return
	}

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(170, 6, fmt.Sprintf("Total Scope 2 Activities: %d", len(scope2Activities)))
	pdf.Ln(8)

	// Similar table structure
	displayLimit := 20
	pdf.SetFont("Arial", "B", 8)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(65, 6, "Activity", "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 6, "Category", "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 6, "Quantity", "1", 0, "R", true, 0, "")
	pdf.CellFormat(35, 6, "Emissions (tCO2e)", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 7)
	pdf.SetFillColor(245, 245, 245)

	for i, activity := range scope2Activities {
		if i >= displayLimit {
			break
		}
		fill := i%2 == 0
		pdf.CellFormat(65, 5, truncate(activity.Name, 40), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(35, 5, truncate(activity.Category, 22), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(35, 5, fmt.Sprintf("%.2f %s", activity.Quantity, activity.Unit), "1", 0, "R", fill, 0, "")
		pdf.CellFormat(35, 5, fmt.Sprintf("%.3f", activity.EmissionsTonnes), "1", 1, "R", fill, 0, "")
	}
}

func addSECScope3Breakdown(pdf *gofpdf.Fpdf, report SECReport) {
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(170, 10, "Scope 3 Emissions Breakdown")
	pdf.Ln(12)

	pdf.SetFont("Arial", "I", 10)
	pdf.MultiCell(170, 5, "Scope 3 emissions represent indirect value chain emissions. Disclosure is provided where material to the organization's emissions profile.", "", "", false)
	pdf.Ln(6)

	scope3Activities := []ActivityEmission{}
	for _, activity := range report.EmissionsData.Activities {
		if activity.Scope == "scope3" || activity.Scope == "Scope 3" {
			scope3Activities = append(scope3Activities, activity)
		}
	}

	if len(scope3Activities) == 0 {
		pdf.SetFont("Arial", "", 11)
		pdf.Cell(170, 7, "No material Scope 3 emissions reported for this period.")
		return
	}

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(170, 6, fmt.Sprintf("Total Scope 3 Activities: %d", len(scope3Activities)))
	pdf.Ln(8)

	// Activity table
	displayLimit := 20
	pdf.SetFont("Arial", "B", 8)
	pdf.SetFillColor(200, 200, 200)
	pdf.CellFormat(65, 6, "Activity", "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 6, "Category", "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 6, "Quantity", "1", 0, "R", true, 0, "")
	pdf.CellFormat(35, 6, "Emissions (tCO2e)", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 7)
	pdf.SetFillColor(245, 245, 245)

	for i, activity := range scope3Activities {
		if i >= displayLimit {
			break
		}
		fill := i%2 == 0
		pdf.CellFormat(65, 5, truncate(activity.Name, 40), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(35, 5, truncate(activity.Category, 22), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(35, 5, fmt.Sprintf("%.2f %s", activity.Quantity, activity.Unit), "1", 0, "R", fill, 0, "")
		pdf.CellFormat(35, 5, fmt.Sprintf("%.3f", activity.EmissionsTonnes), "1", 1, "R", fill, 0, "")
	}
}

func addSECDataQualityAssurance(pdf *gofpdf.Fpdf, report SECReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Data Quality and Assurance")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Data Quality Metrics")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(90, 6, "Overall Data Quality Score:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(80, 6, fmt.Sprintf("%.1f%%", report.QualityMetrics.DataQualityScore))
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(90, 6, "Data Completeness:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(80, 6, fmt.Sprintf("%.1f%%", report.QualityMetrics.CompletenessPercentage))
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(90, 6, "Activities Analyzed:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(80, 6, fmt.Sprintf("%d", report.QualityMetrics.TotalActivities))
	pdf.Ln(12)

	// Assurance statement
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Attestation")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(170, 5, "This climate disclosure has been prepared based on available data and reasonable estimation methodologies. "+
		"The data presented reflects management's best understanding of the organization's greenhouse gas emissions profile. "+
		"Third-party verification may be obtained in future reporting periods.", "", "", false)

	pdf.Ln(8)

	pdf.SetFont("Arial", "", 9)
	pdf.Cell(80, 6, "Report Generated:")
	pdf.Cell(90, 6, report.Report.GenerationTimestamp.Format("January 2, 2006 at 15:04 MST"))
	pdf.Ln(6)

	pdf.Cell(80, 6, "Report Hash (SHA-256):")
	pdf.Ln(6)
	pdf.SetFont("Courier", "", 7)
	pdf.Cell(170, 5, report.Report.ReportHash)
}
