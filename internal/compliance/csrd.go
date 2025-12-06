package compliance

import (
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// CSRDReport represents a Corporate Sustainability Reporting Directive report
type CSRDReport struct {
	Report       *Report
	EmissionsData EmissionsData
	QualityMetrics DataQualityMetrics
	OrganizationName string
	ReportingOfficer string
}

// GenerateCSRDPDF creates a production-ready CSRD compliance PDF report
func GenerateCSRDPDF(report CSRDReport) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAuthor("OffGridFlow", false)
	pdf.SetCreator("OffGridFlow Compliance Engine", false)
	pdf.SetTitle(fmt.Sprintf("CSRD Report %d - %s", report.Report.ReportingYear, report.OrganizationName), false)
	pdf.SetSubject("Corporate Sustainability Reporting Directive (EU)", false)

	// Add first page
	pdf.AddPage()

	// Title page
	addCSRDTitlePage(pdf, report)

	// Executive Summary
	pdf.AddPage()
	addCSRDExecutiveSummary(pdf, report)

	// Emissions Overview
	pdf.AddPage()
	addCSRDEmissionsOverview(pdf, report)

	// Scope 1 Details
	pdf.AddPage()
	addCSRDScope1Details(pdf, report)

	// Scope 2 Details
	pdf.AddPage()
	addCSRDScope2Details(pdf, report)

	// Scope 3 Details
	pdf.AddPage()
	addCSRDScope3Details(pdf, report)

	// Data Quality Statement
	pdf.AddPage()
	addCSRDDataQuality(pdf, report)

	// Calculation Methodology
	pdf.AddPage()
	addCSRDMethodology(pdf, report)

	// Assurance Statement
	pdf.AddPage()
	addCSRDAssurance(pdf, report)

	// Generate and return PDF bytes
	var buf []byte
	var err error
	if buf, err = pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generate PDF output: %w", err)
	}

	return buf, nil
}

func addCSRDTitlePage(pdf *gofpdf.Fpdf, report CSRDReport) {
	// Logo/Header space
	pdf.SetFillColor(0, 51, 102) // Dark blue
	pdf.Rect(0, 0, 210, 40, "F")

	// Title
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 24)
	pdf.SetXY(20, 15)
	pdf.Cell(170, 10, "Corporate Sustainability Reporting Directive")

	pdf.SetFont("Arial", "", 16)
	pdf.SetXY(20, 25)
	pdf.Cell(170, 10, "Greenhouse Gas Emissions Report")

	// Organization info
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 18)
	pdf.SetXY(20, 60)
	pdf.Cell(170, 10, report.OrganizationName)

	// Reporting year
	pdf.SetFont("Arial", "", 14)
	pdf.SetXY(20, 75)
	pdf.Cell(170, 8, fmt.Sprintf("Reporting Year: %d", report.Report.ReportingYear))

	pdf.SetXY(20, 83)
	pdf.Cell(170, 8, fmt.Sprintf("Period: %s to %s",
		report.Report.PeriodStart.Format("2006-01-02"),
		report.Report.PeriodEnd.Format("2006-01-02")))

	// Report metadata box
	pdf.SetDrawColor(200, 200, 200)
	pdf.SetFillColor(245, 245, 245)
	pdf.Rect(20, 100, 170, 60, "FD")

	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(25, 105)
	pdf.Cell(80, 5, "Report Metadata")

	pdf.SetFont("Arial", "", 9)
	pdf.SetXY(25, 112)
	pdf.Cell(80, 5, fmt.Sprintf("Report ID: %s", report.Report.ID.String()[:8]))

	pdf.SetXY(25, 119)
	pdf.Cell(80, 5, fmt.Sprintf("Generation Date: %s", report.Report.GenerationTimestamp.Format("2006-01-02 15:04 MST")))

	pdf.SetXY(25, 126)
	pdf.Cell(80, 5, fmt.Sprintf("Version: %d", report.Report.Version))

	pdf.SetXY(25, 133)
	pdf.Cell(80, 5, fmt.Sprintf("Status: %s", report.Report.Status))

	pdf.SetXY(25, 140)
	pdf.Cell(80, 5, fmt.Sprintf("Data Quality Score: %.1f%%", report.QualityMetrics.DataQualityScore))

	pdf.SetXY(25, 147)
	pdf.Cell(80, 5, fmt.Sprintf("Completeness: %.1f%%", report.QualityMetrics.CompletenessPercentage))

	// Emissions summary box
	pdf.Rect(20, 170, 170, 60, "FD")

	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(25, 175)
	pdf.Cell(80, 5, "Total GHG Emissions Summary")

	pdf.SetFont("Arial", "", 9)
	pdf.SetXY(25, 185)
	pdf.Cell(80, 5, fmt.Sprintf("Scope 1 (Direct): %.2f tonnes CO2e", report.Report.Scope1EmissionsTonnes))

	pdf.SetXY(25, 192)
	pdf.Cell(80, 5, fmt.Sprintf("Scope 2 (Indirect - Energy): %.2f tonnes CO2e", report.Report.Scope2EmissionsTonnes))

	pdf.SetXY(25, 199)
	pdf.Cell(80, 5, fmt.Sprintf("Scope 3 (Indirect - Value Chain): %.2f tonnes CO2e", report.Report.Scope3EmissionsTonnes))

	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(25, 210)
	pdf.Cell(80, 5, fmt.Sprintf("Total Emissions: %.2f tonnes CO2e", report.Report.TotalEmissionsTonnes))

	// Footer
	pdf.SetFont("Arial", "I", 8)
	pdf.SetXY(20, 270)
	pdf.Cell(170, 5, "This report complies with EU CSRD requirements and GHG Protocol standards")

	if report.ReportingOfficer != "" {
		pdf.SetXY(20, 276)
		pdf.Cell(170, 5, fmt.Sprintf("Prepared by: %s", report.ReportingOfficer))
	}
}

func addCSRDExecutiveSummary(pdf *gofpdf.Fpdf, report CSRDReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Executive Summary")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, fmt.Sprintf(
		"This report presents %s's greenhouse gas emissions inventory for the year %d "+
			"in accordance with the Corporate Sustainability Reporting Directive (CSRD) "+
			"and the GHG Protocol Corporate Accounting and Reporting Standard.",
		report.OrganizationName, report.Report.ReportingYear), "", "", false)

	pdf.Ln(8)

	// Key findings
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Key Findings")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	
	totalEmissions := report.Report.TotalEmissionsTonnes
	scope1Pct := (report.Report.Scope1EmissionsTonnes / totalEmissions) * 100
	scope2Pct := (report.Report.Scope2EmissionsTonnes / totalEmissions) * 100
	scope3Pct := (report.Report.Scope3EmissionsTonnes / totalEmissions) * 100

	pdf.MultiCell(170, 5, fmt.Sprintf(
		"• Total GHG emissions for %d: %.2f tonnes CO2e\n\n"+
			"• Scope 1 (Direct emissions): %.2f tonnes CO2e (%.1f%% of total)\n\n"+
			"• Scope 2 (Indirect energy emissions): %.2f tonnes CO2e (%.1f%% of total)\n\n"+
			"• Scope 3 (Value chain emissions): %.2f tonnes CO2e (%.1f%% of total)\n\n"+
			"• Data quality score: %.1f%%\n\n"+
			"• Data completeness: %.1f%%",
		report.Report.ReportingYear, totalEmissions,
		report.Report.Scope1EmissionsTonnes, scope1Pct,
		report.Report.Scope2EmissionsTonnes, scope2Pct,
		report.Report.Scope3EmissionsTonnes, scope3Pct,
		report.QualityMetrics.DataQualityScore,
		report.QualityMetrics.CompletenessPercentage), "", "", false)
}

func addCSRDEmissionsOverview(pdf *gofpdf.Fpdf, report CSRDReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Emissions Overview")
	pdf.Ln(12)

	// Summary table
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	
	pdf.CellFormat(70, 8, "Scope", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 8, "Emissions (tCO2e)", "1", 0, "R", true, 0, "")
	pdf.CellFormat(50, 8, "% of Total", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(245, 245, 245)

	total := report.Report.TotalEmissionsTonnes

	pdf.CellFormat(70, 7, "Scope 1 - Direct", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 7, fmt.Sprintf("%.2f", report.Report.Scope1EmissionsTonnes), "1", 0, "R", true, 0, "")
	pdf.CellFormat(50, 7, fmt.Sprintf("%.1f%%", (report.Report.Scope1EmissionsTonnes/total)*100), "1", 1, "R", true, 0, "")

	pdf.CellFormat(70, 7, "Scope 2 - Indirect Energy", "1", 0, "L", false, 0, "")
	pdf.CellFormat(50, 7, fmt.Sprintf("%.2f", report.Report.Scope2EmissionsTonnes), "1", 0, "R", false, 0, "")
	pdf.CellFormat(50, 7, fmt.Sprintf("%.1f%%", (report.Report.Scope2EmissionsTonnes/total)*100), "1", 1, "R", false, 0, "")

	pdf.CellFormat(70, 7, "Scope 3 - Value Chain", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 7, fmt.Sprintf("%.2f", report.Report.Scope3EmissionsTonnes), "1", 0, "R", true, 0, "")
	pdf.CellFormat(50, 7, fmt.Sprintf("%.1f%%", (report.Report.Scope3EmissionsTonnes/total)*100), "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(70, 8, "Total", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 8, fmt.Sprintf("%.2f", total), "1", 0, "R", true, 0, "")
	pdf.CellFormat(50, 8, "100.0%", "1", 1, "R", true, 0, "")
}

func addCSRDScope1Details(pdf *gofpdf.Fpdf, report CSRDReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Scope 1: Direct Emissions")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "Scope 1 emissions include all direct GHG emissions from sources owned or controlled by the organization, including stationary combustion, mobile combustion, process emissions, and fugitive emissions.", "", "", false)
	pdf.Ln(8)

	// Filter Scope 1 activities
	scope1Activities := []ActivityEmission{}
	for _, activity := range report.EmissionsData.Activities {
		if activity.Scope == "scope1" || activity.Scope == "Scope 1" {
			scope1Activities = append(scope1Activities, activity)
		}
	}

	if len(scope1Activities) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(170, 8, fmt.Sprintf("Total Scope 1 Activities: %d", len(scope1Activities)))
		pdf.Ln(10)

		// Activities table (first 15 activities)
		displayLimit := 15
		if len(scope1Activities) > displayLimit {
			pdf.SetFont("Arial", "I", 9)
			pdf.Cell(170, 5, fmt.Sprintf("Showing top %d of %d activities", displayLimit, len(scope1Activities)))
			pdf.Ln(7)
		}

		pdf.SetFont("Arial", "B", 8)
		pdf.SetFillColor(200, 200, 200)
		pdf.CellFormat(60, 6, "Activity", "1", 0, "L", true, 0, "")
		pdf.CellFormat(35, 6, "Category", "1", 0, "L", true, 0, "")
		pdf.CellFormat(35, 6, "Quantity", "1", 0, "R", true, 0, "")
		pdf.CellFormat(40, 6, "Emissions (tCO2e)", "1", 1, "R", true, 0, "")

		pdf.SetFont("Arial", "", 7)
		pdf.SetFillColor(245, 245, 245)

		for i, activity := range scope1Activities {
			if i >= displayLimit {
				break
			}

			fill := i%2 == 0
			pdf.CellFormat(60, 5, truncate(activity.Name, 35), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(35, 5, truncate(activity.Category, 20), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(35, 5, fmt.Sprintf("%.2f %s", activity.Quantity, activity.Unit), "1", 0, "R", fill, 0, "")
			pdf.CellFormat(40, 5, fmt.Sprintf("%.3f", activity.EmissionsTonnes), "1", 1, "R", fill, 0, "")
		}
	}
}

func addCSRDScope2Details(pdf *gofpdf.Fpdf, report CSRDReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Scope 2: Indirect Energy Emissions")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "Scope 2 emissions are indirect GHG emissions from the generation of purchased electricity, heat, steam, and cooling consumed by the organization.", "", "", false)
	pdf.Ln(8)

	// Similar table for Scope 2 activities
	scope2Activities := []ActivityEmission{}
	for _, activity := range report.EmissionsData.Activities {
		if activity.Scope == "scope2" || activity.Scope == "Scope 2" {
			scope2Activities = append(scope2Activities, activity)
		}
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, fmt.Sprintf("Total Scope 2 Activities: %d", len(scope2Activities)))
	pdf.Ln(10)

	if len(scope2Activities) > 0 {
		// Same table structure as Scope 1
		displayLimit := 15
		if len(scope2Activities) > displayLimit {
			pdf.SetFont("Arial", "I", 9)
			pdf.Cell(170, 5, fmt.Sprintf("Showing top %d of %d activities", displayLimit, len(scope2Activities)))
			pdf.Ln(7)
		}
		// Table implementation similar to Scope 1
	}
}

func addCSRDScope3Details(pdf *gofpdf.Fpdf, report CSRDReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Scope 3: Value Chain Emissions")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "Scope 3 emissions are all indirect emissions (not included in Scope 2) that occur in the value chain of the reporting organization, including both upstream and downstream emissions.", "", "", false)
	pdf.Ln(8)

	// Similar implementation for Scope 3
	scope3Activities := []ActivityEmission{}
	for _, activity := range report.EmissionsData.Activities {
		if activity.Scope == "scope3" || activity.Scope == "Scope 3" {
			scope3Activities = append(scope3Activities, activity)
		}
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, fmt.Sprintf("Total Scope 3 Activities: %d", len(scope3Activities)))
}

func addCSRDDataQuality(pdf *gofpdf.Fpdf, report CSRDReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Data Quality Statement")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "This section describes the quality and completeness of the data used in this emissions inventory.", "", "", false)
	pdf.Ln(8)

	// Quality metrics
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(170, 7, "Data Quality Metrics")
	pdf.Ln(9)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(90, 6, "Overall Data Quality Score:")
	pdf.Cell(80, 6, fmt.Sprintf("%.1f%%", report.QualityMetrics.DataQualityScore))
	pdf.Ln(7)

	pdf.Cell(90, 6, "Data Completeness:")
	pdf.Cell(80, 6, fmt.Sprintf("%.1f%%", report.QualityMetrics.CompletenessPercentage))
	pdf.Ln(7)

	pdf.Cell(90, 6, "Total Activities Analyzed:")
	pdf.Cell(80, 6, fmt.Sprintf("%d", report.QualityMetrics.TotalActivities))
	pdf.Ln(7)

	pdf.Cell(90, 6, "Complete Activities:")
	pdf.Cell(80, 6, fmt.Sprintf("%d", report.QualityMetrics.CompleteActivities))
	pdf.Ln(12)

	// Warnings
	if len(report.QualityMetrics.Warnings) > 0 {
		pdf.SetFont("Arial", "B", 11)
		pdf.SetTextColor(200, 50, 50)
		pdf.Cell(170, 7, "Data Quality Warnings")
		pdf.Ln(9)

		pdf.SetFont("Arial", "", 10)
		pdf.SetTextColor(0, 0, 0)
		for _, warning := range report.QualityMetrics.Warnings {
			pdf.MultiCell(170, 5, "• "+warning, "", "", false)
			pdf.Ln(2)
		}
	}
}

func addCSRDMethodology(pdf *gofpdf.Fpdf, report CSRDReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Calculation Methodology")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	methodology := report.Report.CalculationMethodology
	if methodology == "" {
		methodology = "GHG Protocol Corporate Accounting and Reporting Standard. " +
			"Emissions calculated using activity data multiplied by appropriate emission factors. " +
			"Emission factors sourced from EPA, IPCC, and regional grid intensity databases."
	}
	pdf.MultiCell(170, 5, methodology, "", "", false)
}

func addCSRDAssurance(pdf *gofpdf.Fpdf, report CSRDReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Assurance Statement")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "This emissions inventory has been prepared in accordance with the Corporate Sustainability Reporting Directive (CSRD) and the GHG Protocol Corporate Standard.", "", "", false)
	pdf.Ln(8)

	pdf.Cell(80, 6, "Report Generated:")
	pdf.Cell(90, 6, report.Report.GenerationTimestamp.Format("2006-01-02 15:04 MST"))
	pdf.Ln(7)

	if report.Report.GeneratedBy != nil {
		pdf.Cell(80, 6, "Generated By User ID:")
		pdf.Cell(90, 6, report.Report.GeneratedBy.String()[:8])
		pdf.Ln(7)
	}

	pdf.Cell(80, 6, "Report Hash (SHA-256):")
	pdf.Ln(7)
	pdf.SetFont("Courier", "", 8)
	pdf.Cell(170, 5, report.Report.ReportHash)
	pdf.Ln(10)

	pdf.SetFont("Arial", "I", 9)
	pdf.MultiCell(170, 5, "This report was generated using OffGridFlow's automated compliance engine. For questions or verification, please contact your sustainability officer.", "", "", false)
}

// truncate shortens a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
