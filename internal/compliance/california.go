package compliance

import (
	"bytes"
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

// CaliforniaReport represents a California Climate Corporate Data Accountability Act report
// SB 253 requires companies with >$1B revenue operating in California to disclose Scopes 1, 2, 3
type CaliforniaReport struct {
	Report            *Report
	EmissionsData     EmissionsData
	QualityMetrics    DataQualityMetrics
	OrganizationName  string
	ReportingYear     int
	ReportingOfficer  string
	
	// California-specific fields
	AnnualRevenue     float64 // USD
	CAOperations      bool
	Scope3Categories  []Scope3Category
	Assurance         AssuranceInfo
}

// Scope3Category represents a specific Scope 3 emissions category
type Scope3Category struct {
	Number      int
	Name        string
	Emissions   float64
	Methodology string
	DataQuality string
}

// AssuranceInfo contains third-party assurance details
type AssuranceInfo struct {
	Provider      string
	Level         string // "Limited" or "Reasonable"
	Standard      string // e.g., "ISO 14064-3"
	OpinionDate   string
	Opinion       string
}

// GenerateCaliforniaPDF creates a California CCDAA compliance PDF
func GenerateCaliforniaPDF(report CaliforniaReport) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAuthor("OffGridFlow", false)
	pdf.SetCreator("OffGridFlow California Compliance Engine", false)
	pdf.SetTitle(fmt.Sprintf("California CCDAA Report %d - %s", report.ReportingYear, report.OrganizationName), false)
	pdf.SetSubject("California Climate Corporate Data Accountability Act (SB 253)", false)

	// Cover page
	pdf.AddPage()
	addCaliforniaCoverPage(pdf, report)

	// Executive Summary
	pdf.AddPage()
	addCaliforniaExecutiveSummary(pdf, report)

	// Scope 1 & 2 Emissions
	pdf.AddPage()
	addCaliforniaScope1And2(pdf, report)

	// Scope 3 Emissions (detailed - required by SB 253)
	pdf.AddPage()
	addCaliforniaScope3Overview(pdf, report)

	// Scope 3 Category Breakdown
	pdf.AddPage()
	addCaliforniaScope3Categories(pdf, report)

	// Methodology and Data Quality
	pdf.AddPage()
	addCaliforniaMethodology(pdf, report)

	// Third-Party Assurance
	if report.Assurance.Provider != "" {
		pdf.AddPage()
		addCaliforniaAssurance(pdf, report)
	}

	// Appendix: Activity Data
	pdf.AddPage()
	addCaliforniaActivityData(pdf, report)

	// Generate PDF bytes
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generate California PDF: %w", err)
	}

	return buf.Bytes(), nil
}

func addCaliforniaCoverPage(pdf *gofpdf.Fpdf, report CaliforniaReport) {
	// California state colors header
	pdf.SetFillColor(0, 84, 166) // California blue
	pdf.Rect(0, 0, 210, 50, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 20)
	pdf.SetXY(20, 15)
	pdf.Cell(170, 10, "STATE OF CALIFORNIA")

	pdf.SetFont("Arial", "", 15)
	pdf.SetXY(20, 27)
	pdf.Cell(170, 8, "Climate Corporate Data Accountability Act")

	pdf.SetFont("Arial", "", 12)
	pdf.SetXY(20, 36)
	pdf.Cell(170, 6, "Senate Bill 253 (SB 253) - Annual Disclosure")

	// Company header
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 18)
	pdf.SetXY(20, 65)
	pdf.Cell(170, 10, report.OrganizationName)

	pdf.SetFont("Arial", "", 13)
	pdf.SetXY(20, 77)
	pdf.Cell(170, 7, fmt.Sprintf("Reporting Year: %d", report.ReportingYear))

	// Compliance statement box
	pdf.SetDrawColor(0, 84, 166)
	pdf.SetLineWidth(0.5)
	pdf.SetFillColor(250, 252, 255)
	pdf.Rect(20, 95, 170, 45, "FD")

	pdf.SetFont("Arial", "B", 11)
	pdf.SetXY(25, 100)
	pdf.Cell(160, 6, "SB 253 Compliance Declaration")

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(25, 109)
	pdf.MultiCell(160, 5, fmt.Sprintf(
		"This report is submitted in compliance with California's Climate Corporate Data Accountability Act "+
			"(SB 253), which requires disclosure of Scope 1, 2, and 3 greenhouse gas emissions for fiscal year %d. "+
			"The reporting entity operates in California with annual revenue exceeding $1 billion USD.", report.ReportingYear), "", "", false)

	pdf.SetFont("Arial", "", 9)
	pdf.SetXY(25, 130)
	pdf.Cell(80, 5, fmt.Sprintf("Annual Revenue: $%.2fB", report.AnnualRevenue/1_000_000_000))

	// Emissions summary
	pdf.SetDrawColor(100, 100, 100)
	pdf.SetFillColor(245, 250, 255)
	pdf.Rect(20, 150, 170, 80, "FD")

	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(25, 155)
	pdf.Cell(160, 7, "GHG Emissions Summary (metric tonnes CO2e)")

	pdf.SetFont("Arial", "", 10)
	
	pdf.SetXY(30, 168)
	pdf.Cell(90, 6, "Scope 1 (Direct Emissions):")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(70, 6, fmt.Sprintf("%.2f", report.Report.Scope1EmissionsTonnes))

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(30, 176)
	pdf.Cell(90, 6, "Scope 2 (Purchased Electricity):")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(70, 6, fmt.Sprintf("%.2f", report.Report.Scope2EmissionsTonnes))

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(30, 184)
	pdf.Cell(90, 6, "Scope 3 (Value Chain) [SB 253 Required]:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(70, 6, fmt.Sprintf("%.2f", report.Report.Scope3EmissionsTonnes))

	pdf.SetDrawColor(0, 84, 166)
	pdf.SetLineWidth(0.3)
	pdf.Line(30, 195, 185, 195)

	pdf.SetFont("Arial", "B", 11)
	pdf.SetXY(30, 198)
	pdf.Cell(90, 7, "Total GHG Emissions:")
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(70, 7, fmt.Sprintf("%.2f", report.Report.TotalEmissionsTonnes))

	// Data quality
	pdf.SetFont("Arial", "I", 9)
	pdf.SetXY(30, 210)
	pdf.Cell(160, 5, fmt.Sprintf("Data Quality Score: %.1f%% | Completeness: %.1f%%",
		report.QualityMetrics.DataQualityScore, report.QualityMetrics.CompletenessPercentage))

	// Assurance indicator
	if report.Assurance.Provider != "" {
		pdf.SetXY(30, 218)
		pdf.Cell(160, 5, fmt.Sprintf("Third-Party Assurance: %s (%s assurance)", report.Assurance.Provider, report.Assurance.Level))
	}

	// Footer
	pdf.SetFont("Arial", "I", 8)
	pdf.SetXY(20, 280)
	pdf.MultiCell(170, 4, "Report prepared in accordance with the GHG Protocol Corporate Accounting and Reporting Standard "+
		"and California Air Resources Board guidance. Submitted to the California Climate Accountability Registry.", "", "", false)
}

func addCaliforniaExecutiveSummary(pdf *gofpdf.Fpdf, report CaliforniaReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Executive Summary")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, fmt.Sprintf(
		"%s is submitting this greenhouse gas emissions disclosure in compliance with California's "+
			"Climate Corporate Data Accountability Act (SB 253). As a company with annual revenue exceeding "+
			"$1 billion and operations in California, we are required to publicly disclose our Scope 1, 2, and 3 "+
			"emissions on an annual basis.", report.OrganizationName), "", "", false)

	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Reporting Highlights")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)

	total := report.Report.TotalEmissionsTonnes
	scope1Pct := (report.Report.Scope1EmissionsTonnes / total) * 100
	scope2Pct := (report.Report.Scope2EmissionsTonnes / total) * 100
	scope3Pct := (report.Report.Scope3EmissionsTonnes / total) * 100

	pdf.MultiCell(170, 5, fmt.Sprintf(
		"• Reporting Period: January 1 - December 31, %d\n\n"+
			"• Total GHG Emissions: %.2f metric tonnes CO2e\n\n"+
			"• Scope 1 (Direct): %.2f tCO2e (%.1f%% of total)\n\n"+
			"• Scope 2 (Indirect - Energy): %.2f tCO2e (%.1f%% of total)\n\n"+
			"• Scope 3 (Value Chain): %.2f tCO2e (%.1f%% of total)\n\n"+
			"• Scope 3 Categories Reported: %d of 15 GHG Protocol categories\n\n"+
			"• Data Completeness: %.1f%%\n\n"+
			"• Third-Party Assurance: %s",
		report.ReportingYear,
		total,
		report.Report.Scope1EmissionsTonnes, scope1Pct,
		report.Report.Scope2EmissionsTonnes, scope2Pct,
		report.Report.Scope3EmissionsTonnes, scope3Pct,
		len(report.Scope3Categories),
		report.QualityMetrics.CompletenessPercentage,
		getAssuranceStatus(report)), "", "", false)
}

func addCaliforniaScope1And2(pdf *gofpdf.Fpdf, report CaliforniaReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Scope 1 & 2 Emissions")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "Scope 1 and Scope 2 emissions represent the direct and indirect energy-related "+
		"greenhouse gas emissions from sources owned or controlled by the organization.", "", "", false)
	pdf.Ln(8)

	// Summary table
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 230, 240)
	pdf.CellFormat(85, 8, "Emission Scope", "1", 0, "L", true, 0, "")
	pdf.CellFormat(45, 8, "tCO2e", "1", 0, "R", true, 0, "")
	pdf.CellFormat(40, 8, "% of S1+S2", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(248, 250, 252)

	s1s2Total := report.Report.Scope1EmissionsTonnes + report.Report.Scope2EmissionsTonnes

	pdf.CellFormat(85, 7, "Scope 1 - Direct Emissions", "1", 0, "L", true, 0, "")
	pdf.CellFormat(45, 7, fmt.Sprintf("%.2f", report.Report.Scope1EmissionsTonnes), "1", 0, "R", true, 0, "")
	pdf.CellFormat(40, 7, fmt.Sprintf("%.1f%%", (report.Report.Scope1EmissionsTonnes/s1s2Total)*100), "1", 1, "R", true, 0, "")

	pdf.CellFormat(85, 7, "Scope 2 - Purchased Electricity (Location)", "1", 0, "L", false, 0, "")
	pdf.CellFormat(45, 7, fmt.Sprintf("%.2f", report.Report.Scope2EmissionsTonnes), "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 7, fmt.Sprintf("%.1f%%", (report.Report.Scope2EmissionsTonnes/s1s2Total)*100), "1", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(85, 8, "Total Scope 1 + 2", "1", 0, "L", true, 0, "")
	pdf.CellFormat(45, 8, fmt.Sprintf("%.2f", s1s2Total), "1", 0, "R", true, 0, "")
	pdf.CellFormat(40, 8, "100.0%", "1", 1, "R", true, 0, "")

	pdf.Ln(8)

	// Breakdown by category
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(170, 7, "Scope 1 Emission Sources")
	pdf.Ln(9)

	scope1Activities := []ActivityEmission{}
	for _, activity := range report.EmissionsData.Activities {
		if activity.Scope == "scope1" || activity.Scope == "Scope 1" {
			scope1Activities = append(scope1Activities, activity)
		}
	}

	if len(scope1Activities) > 0 {
		// Category aggregation
		categoryTotals := make(map[string]float64)
		for _, act := range scope1Activities {
			categoryTotals[act.Category] += act.EmissionsTonnes
		}

		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(240, 240, 240)
		pdf.CellFormat(120, 6, "Category", "1", 0, "L", true, 0, "")
		pdf.CellFormat(50, 6, "Emissions (tCO2e)", "1", 1, "R", true, 0, "")

		pdf.SetFont("Arial", "", 9)
		pdf.SetFillColor(250, 250, 250)

		i := 0
		for category, emissions := range categoryTotals {
			fill := i%2 == 0
			pdf.CellFormat(120, 5, category, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(50, 5, fmt.Sprintf("%.2f", emissions), "1", 1, "R", fill, 0, "")
			i++
		}
	}

	pdf.Ln(6)

	// Scope 2 breakdown
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(170, 7, "Scope 2 Emission Sources")
	pdf.Ln(9)

	scope2Activities := []ActivityEmission{}
	for _, activity := range report.EmissionsData.Activities {
		if activity.Scope == "scope2" || activity.Scope == "Scope 2" {
			scope2Activities = append(scope2Activities, activity)
		}
	}

	if len(scope2Activities) > 0 {
		categoryTotals := make(map[string]float64)
		for _, act := range scope2Activities {
			categoryTotals[act.Category] += act.EmissionsTonnes
		}

		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(240, 240, 240)
		pdf.CellFormat(120, 6, "Category", "1", 0, "L", true, 0, "")
		pdf.CellFormat(50, 6, "Emissions (tCO2e)", "1", 1, "R", true, 0, "")

		pdf.SetFont("Arial", "", 9)
		pdf.SetFillColor(250, 250, 250)

		i := 0
		for category, emissions := range categoryTotals {
			fill := i%2 == 0
			pdf.CellFormat(120, 5, category, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(50, 5, fmt.Sprintf("%.2f", emissions), "1", 1, "R", fill, 0, "")
			i++
		}
	}
}

func addCaliforniaScope3Overview(pdf *gofpdf.Fpdf, report CaliforniaReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Scope 3 Value Chain Emissions")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(0, 84, 166)
	pdf.Cell(170, 7, "SB 253 REQUIREMENT: Full Scope 3 Disclosure")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(9)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "California SB 253 requires comprehensive disclosure of Scope 3 emissions across all 15 categories "+
		"defined by the GHG Protocol. Scope 3 represents indirect emissions from the organization's value chain, "+
		"including both upstream and downstream activities.", "", "", false)
	pdf.Ln(8)

	// Total Scope 3
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(120, 7, "Total Scope 3 Emissions:")
	pdf.Cell(50, 7, fmt.Sprintf("%.2f tCO2e", report.Report.Scope3EmissionsTonnes))
	pdf.Ln(9)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(120, 6, "Number of Categories Reported:")
	pdf.Cell(50, 6, fmt.Sprintf("%d of 15", len(report.Scope3Categories)))
	pdf.Ln(7)

	total := report.Report.TotalEmissionsTonnes
	pdf.Cell(120, 6, "Scope 3 as % of Total Emissions:")
	pdf.Cell(50, 6, fmt.Sprintf("%.1f%%", (report.Report.Scope3EmissionsTonnes/total)*100))
	pdf.Ln(12)

	// Summary note
	pdf.SetFont("Arial", "I", 9)
	pdf.SetFillColor(255, 252, 230)
	pdf.Rect(20, pdf.GetY(), 170, 25, "F")
	pdf.SetXY(25, pdf.GetY()+2)
	pdf.MultiCell(160, 4, "Note: Scope 3 emission calculations involve significant estimation and rely on "+
		"secondary data sources for many categories. The organization is working to improve data quality through "+
		"enhanced supplier engagement and primary data collection. See methodology section for detailed calculation approaches.", "", "", false)
}

func addCaliforniaScope3Categories(pdf *gofpdf.Fpdf, report CaliforniaReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Scope 3 Category Breakdown")
	pdf.Ln(12)

	if len(report.Scope3Categories) == 0 {
		pdf.SetFont("Arial", "", 11)
		pdf.Cell(170, 7, "No Scope 3 categories reported.")
		return
	}

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(170, 5, "The following table details emissions by GHG Protocol Scope 3 category:", "", "", false)
	pdf.Ln(6)

	// Category table
	pdf.SetFont("Arial", "B", 8)
	pdf.SetFillColor(220, 230, 240)
	pdf.CellFormat(12, 6, "Cat.", "1", 0, "C", true, 0, "")
	pdf.CellFormat(70, 6, "Category Name", "1", 0, "L", true, 0, "")
	pdf.CellFormat(38, 6, "Emissions (tCO2e)", "1", 0, "R", true, 0, "")
	pdf.CellFormat(50, 6, "Data Quality", "1", 1, "L", true, 0, "")

	pdf.SetFont("Arial", "", 7)
	pdf.SetFillColor(248, 250, 252)

	for i, cat := range report.Scope3Categories {
		fill := i%2 == 0
		pdf.CellFormat(12, 5, fmt.Sprintf("%d", cat.Number), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(70, 5, truncate(cat.Name, 50), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(38, 5, fmt.Sprintf("%.2f", cat.Emissions), "1", 0, "R", fill, 0, "")
		pdf.CellFormat(50, 5, cat.DataQuality, "1", 1, "L", fill, 0, "")
	}

	pdf.Ln(8)

	// Methodology notes
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(170, 6, "Category Calculation Methodologies")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 8)
	for _, cat := range report.Scope3Categories {
		if cat.Methodology != "" {
			pdf.SetFont("Arial", "B", 8)
			pdf.Cell(170, 4, fmt.Sprintf("Category %d - %s:", cat.Number, cat.Name))
			pdf.Ln(5)

			pdf.SetFont("Arial", "", 8)
			pdf.SetX(25)
			pdf.MultiCell(165, 3, cat.Methodology, "", "", false)
			pdf.Ln(2)
		}
	}
}

func addCaliforniaMethodology(pdf *gofpdf.Fpdf, report CaliforniaReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Calculation Methodology & Data Quality")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Reporting Standards")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(170, 5, "This disclosure has been prepared in accordance with:\n"+
		"• GHG Protocol Corporate Accounting and Reporting Standard\n"+
		"• GHG Protocol Scope 3 Standard\n"+
		"• California Air Resources Board SB 253 guidance\n"+
		"• ISO 14064-1:2018 GHG quantification and reporting", "", "", false)

	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Data Quality Assessment")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(95, 6, "Overall Data Quality Score:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(75, 6, fmt.Sprintf("%.1f%%", report.QualityMetrics.DataQualityScore))
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(95, 6, "Data Completeness:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(75, 6, fmt.Sprintf("%.1f%%", report.QualityMetrics.CompletenessPercentage))
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(95, 6, "Activities Analyzed:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(75, 6, fmt.Sprintf("%d", report.QualityMetrics.TotalActivities))
	pdf.Ln(12)

	// Warnings
	if len(report.QualityMetrics.Warnings) > 0 {
		pdf.SetFont("Arial", "B", 11)
		pdf.SetTextColor(180, 50, 50)
		pdf.Cell(170, 7, "Data Quality Notices")
		pdf.SetTextColor(0, 0, 0)
		pdf.Ln(9)

		pdf.SetFont("Arial", "", 9)
		for _, warning := range report.QualityMetrics.Warnings {
			pdf.Cell(5, 4, "•")
			pdf.MultiCell(165, 4, warning, "", "", false)
			pdf.Ln(1)
		}
	}
}

func addCaliforniaAssurance(pdf *gofpdf.Fpdf, report CaliforniaReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Third-Party Assurance")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "In accordance with SB 253 requirements, this emissions disclosure has been subject to "+
		"independent third-party assurance.", "", "", false)
	pdf.Ln(8)

	// Assurance details table
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 245, 250)
	pdf.Rect(20, pdf.GetY(), 170, 50, "F")

	y := pdf.GetY() + 5
	pdf.SetXY(25, y)
	pdf.Cell(80, 6, "Assurance Provider:")
	pdf.Cell(85, 6, report.Assurance.Provider)

	y += 8
	pdf.SetXY(25, y)
	pdf.Cell(80, 6, "Assurance Level:")
	pdf.Cell(85, 6, report.Assurance.Level)

	y += 8
	pdf.SetXY(25, y)
	pdf.Cell(80, 6, "Standard Applied:")
	pdf.Cell(85, 6, report.Assurance.Standard)

	y += 8
	pdf.SetXY(25, y)
	pdf.Cell(80, 6, "Opinion Date:")
	pdf.Cell(85, 6, report.Assurance.OpinionDate)

	pdf.SetY(pdf.GetY() + 52)

	// Opinion statement
	if report.Assurance.Opinion != "" {
		pdf.Ln(5)
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(170, 7, "Assurance Opinion")
		pdf.Ln(9)

		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(170, 5, report.Assurance.Opinion, "", "", false)
	}
}

func addCaliforniaActivityData(pdf *gofpdf.Fpdf, report CaliforniaReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Appendix: Activity Data Summary")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(170, 5, "This appendix provides a summary of emission-generating activities by scope.", "", "", false)
	pdf.Ln(6)

	// Activity count by scope
	scope1Count := 0
	scope2Count := 0
	scope3Count := 0

	for _, act := range report.EmissionsData.Activities {
		switch act.Scope {
		case "scope1", "Scope 1":
			scope1Count++
		case "scope2", "Scope 2":
			scope2Count++
		case "scope3", "Scope 3":
			scope3Count++
		}
	}

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(230, 235, 240)
	pdf.CellFormat(60, 6, "Scope", "1", 0, "L", true, 0, "")
	pdf.CellFormat(55, 6, "Activity Count", "1", 0, "R", true, 0, "")
	pdf.CellFormat(55, 6, "Total Emissions (tCO2e)", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(248, 250, 252)

	pdf.CellFormat(60, 5, "Scope 1", "1", 0, "L", true, 0, "")
	pdf.CellFormat(55, 5, fmt.Sprintf("%d", scope1Count), "1", 0, "R", true, 0, "")
	pdf.CellFormat(55, 5, fmt.Sprintf("%.2f", report.Report.Scope1EmissionsTonnes), "1", 1, "R", true, 0, "")

	pdf.CellFormat(60, 5, "Scope 2", "1", 0, "L", false, 0, "")
	pdf.CellFormat(55, 5, fmt.Sprintf("%d", scope2Count), "1", 0, "R", false, 0, "")
	pdf.CellFormat(55, 5, fmt.Sprintf("%.2f", report.Report.Scope2EmissionsTonnes), "1", 1, "R", false, 0, "")

	pdf.CellFormat(60, 5, "Scope 3", "1", 0, "L", true, 0, "")
	pdf.CellFormat(55, 5, fmt.Sprintf("%d", scope3Count), "1", 0, "R", true, 0, "")
	pdf.CellFormat(55, 5, fmt.Sprintf("%.2f", report.Report.Scope3EmissionsTonnes), "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(60, 6, "Total", "1", 0, "L", true, 0, "")
	pdf.CellFormat(55, 6, fmt.Sprintf("%d", len(report.EmissionsData.Activities)), "1", 0, "R", true, 0, "")
	pdf.CellFormat(55, 6, fmt.Sprintf("%.2f", report.Report.TotalEmissionsTonnes), "1", 1, "R", true, 0, "")

	pdf.Ln(8)

	// Report hash
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(170, 6, "Report Integrity Verification")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 9)
	pdf.Cell(60, 5, "Report Hash (SHA-256):")
	pdf.Ln(6)

	pdf.SetFont("Courier", "", 7)
	pdf.Cell(170, 4, report.Report.ReportHash)
	pdf.Ln(7)

	pdf.SetFont("Arial", "I", 8)
	pdf.MultiCell(170, 4, "This cryptographic hash ensures the integrity of the report data and can be used to verify that "+
		"the report contents have not been altered after generation.", "", "", false)
}

func getAssuranceStatus(report CaliforniaReport) string {
	if report.Assurance.Provider != "" {
		return fmt.Sprintf("Yes (%s assurance by %s)", report.Assurance.Level, report.Assurance.Provider)
	}
	return "Not obtained for this reporting period"
}
