package compliance

import (
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// CBAMReport represents a Carbon Border Adjustment Mechanism report
// EU CBAM requires reporting of embedded emissions in imported goods
type CBAMReport struct {
	Report            *Report
	EmissionsData     EmissionsData
	QualityMetrics    DataQualityMetrics
	OrganizationName  string
	ReportingYear     int
	ReportingOfficer  string
	
	// CBAM-specific fields
	ImportedGoods     []ImportedGood
	ProductionSites   []ProductionSite
	InstallationID    string
	OperatorName      string
	TotalCarbonPrice  float64  // EUR
}

// ImportedGood represents a product subject to CBAM
type ImportedGood struct {
	CNCode            string   // Combined Nomenclature code
	Description       string
	Quantity          float64
	Unit              string
	OriginCountry     string
	EmbeddedEmissions float64  // tonnes CO2e per unit
	TotalEmissions    float64  // tonnes CO2e
	CBAMPrice         float64  // EUR
	ProductionRoute   string
}

// ProductionSite represents a manufacturing facility
type ProductionSite struct {
	Name              string
	Location          string
	Country           string
	Coordinates       string
	ProductionProcess string
	EmissionsFactor   float64
	Certified         bool
	CertificationBody string
}

// GenerateCBAMPDF creates an EU CBAM compliance PDF report
func GenerateCBAMPDF(report CBAMReport) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAuthor("OffGridFlow", false)
	pdf.SetCreator("OffGridFlow CBAM Compliance Engine", false)
	pdf.SetTitle(fmt.Sprintf("CBAM Report %d - %s", report.ReportingYear, report.OrganizationName), false)
	pdf.SetSubject("Carbon Border Adjustment Mechanism (EU Regulation 2023/956)", false)

	// Cover page
	pdf.AddPage()
	addCBAMCoverPage(pdf, report)

	// Executive Summary
	pdf.AddPage()
	addCBAMExecutiveSummary(pdf, report)

	// Imported Goods Overview
	pdf.AddPage()
	addCBAMImportedGoods(pdf, report)

	// Embedded Emissions Calculation
	pdf.AddPage()
	addCBAMEmbeddedEmissions(pdf, report)

	// Production Sites
	pdf.AddPage()
	addCBAMProductionSites(pdf, report)

	// CBAM Pricing and Obligations
	pdf.AddPage()
	addCBAMPricing(pdf, report)

	// Verification and Assurance
	pdf.AddPage()
	addCBAMVerification(pdf, report)

	// Methodology
	pdf.AddPage()
	addCBAMMethodology(pdf, report)

	// Generate PDF bytes
	var buf []byte
	var err error
	if buf, err = pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generate CBAM PDF: %w", err)
	}

	return buf, nil
}

func addCBAMCoverPage(pdf *gofpdf.Fpdf, report CBAMReport) {
	// EU flag colors header
	pdf.SetFillColor(0, 51, 153) // EU blue
	pdf.Rect(0, 0, 210, 45, "F")

	// EU stars would go here in production (simplified for now)
	pdf.SetTextColor(255, 204, 0) // EU yellow
	pdf.SetFont("Arial", "B", 24)
	pdf.SetXY(20, 12)
	pdf.Cell(170, 10, "EUROPEAN UNION")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "", 16)
	pdf.SetXY(20, 24)
	pdf.Cell(170, 8, "Carbon Border Adjustment Mechanism")

	pdf.SetFont("Arial", "", 12)
	pdf.SetXY(20, 33)
	pdf.Cell(170, 6, "Regulation (EU) 2023/956")

	// Company header
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 18)
	pdf.SetXY(20, 60)
	pdf.Cell(170, 10, report.OrganizationName)

	pdf.SetFont("Arial", "", 13)
	pdf.SetXY(20, 72)
	pdf.Cell(170, 7, fmt.Sprintf("CBAM Reporting Period: %d", report.ReportingYear))

	// Operator information box
	pdf.SetDrawColor(0, 51, 153)
	pdf.SetLineWidth(0.5)
	pdf.SetFillColor(250, 252, 255)
	pdf.Rect(20, 90, 170, 55, "FD")

	pdf.SetFont("Arial", "B", 11)
	pdf.SetXY(25, 95)
	pdf.Cell(160, 6, "CBAM Operator Declaration")

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(25, 105)
	pdf.Cell(80, 5, fmt.Sprintf("Operator Name: %s", report.OperatorName))

	pdf.SetXY(25, 112)
	pdf.Cell(80, 5, fmt.Sprintf("Installation ID: %s", report.InstallationID))

	pdf.SetXY(25, 119)
	pdf.Cell(80, 5, fmt.Sprintf("Reporting Year: %d", report.ReportingYear))

	pdf.SetXY(25, 126)
	pdf.Cell(80, 5, fmt.Sprintf("Report Generated: %s", time.Now().Format("2006-01-02")))

	pdf.SetXY(25, 133)
	pdf.Cell(80, 5, fmt.Sprintf("Number of Imported Goods: %d", len(report.ImportedGoods)))

	// CBAM obligations summary
	pdf.SetFillColor(255, 250, 240)
	pdf.Rect(20, 155, 170, 75, "FD")

	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(25, 160)
	pdf.Cell(160, 7, "CBAM Carbon Obligations Summary")

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(25, 173)
	pdf.Cell(100, 6, "Total Embedded Emissions:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(60, 6, fmt.Sprintf("%.2f tonnes CO2e", report.Report.TotalEmissionsTonnes))

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(25, 181)
	pdf.Cell(100, 6, "Direct Emissions (Scope 1):")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(60, 6, fmt.Sprintf("%.2f tonnes CO2e", report.Report.Scope1EmissionsTonnes))

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(25, 189)
	pdf.Cell(100, 6, "Indirect Emissions (Electricity):")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(60, 6, fmt.Sprintf("%.2f tonnes CO2e", report.Report.Scope2EmissionsTonnes))

	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(25, 197)
	pdf.Cell(100, 6, "Number of Production Sites:")
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(60, 6, fmt.Sprintf("%d", len(report.ProductionSites)))

	pdf.SetDrawColor(0, 51, 153)
	pdf.SetLineWidth(0.3)
	pdf.Line(25, 208, 185, 208)

	pdf.SetFont("Arial", "B", 11)
	pdf.SetXY(25, 211)
	pdf.Cell(100, 7, "Total CBAM Carbon Price:")
	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(0, 100, 0)
	pdf.Cell(60, 7, fmt.Sprintf("€%.2f", report.TotalCarbonPrice))
	pdf.SetTextColor(0, 0, 0)

	// Footer
	pdf.SetFont("Arial", "I", 8)
	pdf.SetXY(20, 280)
	pdf.MultiCell(170, 4, "This CBAM report has been prepared in accordance with EU Regulation 2023/956. "+
		"The embedded emissions are calculated based on actual production data and default values where applicable. "+
		"CBAM certificates must be surrendered quarterly.", "", "", false)
}

func addCBAMExecutiveSummary(pdf *gofpdf.Fpdf, report CBAMReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Executive Summary")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, fmt.Sprintf(
		"This Carbon Border Adjustment Mechanism (CBAM) report details the embedded emissions in goods "+
			"imported into the European Union by %s during the reporting period %d. CBAM is designed to "+
			"prevent carbon leakage by ensuring that imported goods are subject to carbon pricing equivalent "+
			"to that applied within the EU Emissions Trading System (ETS).",
		report.OrganizationName, report.ReportingYear), "", "", false)

	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "CBAM Reporting Overview")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, fmt.Sprintf(
		"• Reporting Period: January 1 - December 31, %d\n\n"+
			"• Imported Goods Categories: %d CN codes\n\n"+
			"• Total Embedded Emissions: %.2f tonnes CO2e\n\n"+
			"• Direct Emissions (Production): %.2f tonnes CO2e\n\n"+
			"• Indirect Emissions (Electricity): %.2f tonnes CO2e\n\n"+
			"• Production Sites Covered: %d facilities\n\n"+
			"• Total CBAM Liability: €%.2f\n\n"+
			"• Data Quality Score: %.1f%%",
		report.ReportingYear,
		len(report.ImportedGoods),
		report.Report.TotalEmissionsTonnes,
		report.Report.Scope1EmissionsTonnes,
		report.Report.Scope2EmissionsTonnes,
		len(report.ProductionSites),
		report.TotalCarbonPrice,
		report.QualityMetrics.DataQualityScore), "", "", false)

	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Regulatory Context")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(170, 5, "CBAM applies to imports of carbon-intensive goods including cement, iron and steel, "+
		"aluminum, fertilizers, electricity, and hydrogen. Importers must declare embedded emissions quarterly "+
		"and surrender CBAM certificates corresponding to the carbon price differential between the EU ETS and "+
		"any carbon price paid in the country of origin.", "", "", false)
}

func addCBAMImportedGoods(pdf *gofpdf.Fpdf, report CBAMReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Imported Goods Subject to CBAM")
	pdf.Ln(12)

	if len(report.ImportedGoods) == 0 {
		pdf.SetFont("Arial", "", 11)
		pdf.Cell(170, 7, "No CBAM-covered goods imported during this period.")
		return
	}

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(170, 5, fmt.Sprintf(
		"The following table lists all goods imported during %d that fall under CBAM scope, "+
			"identified by their Combined Nomenclature (CN) codes:",
		report.ReportingYear), "", "", false)
	pdf.Ln(6)

	// Goods table
	pdf.SetFont("Arial", "B", 7)
	pdf.SetFillColor(220, 230, 240)
	pdf.CellFormat(22, 6, "CN Code", "1", 0, "L", true, 0, "")
	pdf.CellFormat(48, 6, "Description", "1", 0, "L", true, 0, "")
	pdf.CellFormat(20, 6, "Quantity", "1", 0, "R", true, 0, "")
	pdf.CellFormat(28, 6, "Origin", "1", 0, "L", true, 0, "")
	pdf.CellFormat(28, 6, "Emissions (tCO2e)", "1", 0, "R", true, 0, "")
	pdf.CellFormat(24, 6, "CBAM (EUR)", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 6)
	pdf.SetFillColor(248, 250, 252)

	for i, good := range report.ImportedGoods {
		if i >= 30 {
			pdf.SetFont("Arial", "I", 8)
			pdf.Cell(170, 5, fmt.Sprintf("... and %d more items (see full report)", len(report.ImportedGoods)-30))
			break
		}

		fill := i%2 == 0
		pdf.CellFormat(22, 5, good.CNCode, "1", 0, "L", fill, 0, "")
		pdf.CellFormat(48, 5, truncate(good.Description, 32), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(20, 5, fmt.Sprintf("%.1f %s", good.Quantity, good.Unit), "1", 0, "R", fill, 0, "")
		pdf.CellFormat(28, 5, good.OriginCountry, "1", 0, "L", fill, 0, "")
		pdf.CellFormat(28, 5, fmt.Sprintf("%.2f", good.TotalEmissions), "1", 0, "R", fill, 0, "")
		pdf.CellFormat(24, 5, fmt.Sprintf("%.2f", good.CBAMPrice), "1", 1, "R", fill, 0, "")
	}

	// Totals row
	totalEmissions := 0.0
	totalPrice := 0.0
	for _, good := range report.ImportedGoods {
		totalEmissions += good.TotalEmissions
		totalPrice += good.CBAMPrice
	}

	pdf.SetFont("Arial", "B", 7)
	pdf.SetFillColor(200, 210, 220)
	pdf.CellFormat(90, 6, "TOTAL", "1", 0, "R", true, 0, "")
	pdf.CellFormat(28, 6, "", "1", 0, "L", true, 0, "")
	pdf.CellFormat(28, 6, fmt.Sprintf("%.2f", totalEmissions), "1", 0, "R", true, 0, "")
	pdf.CellFormat(24, 6, fmt.Sprintf("%.2f", totalPrice), "1", 1, "R", true, 0, "")
}

func addCBAMEmbeddedEmissions(pdf *gofpdf.Fpdf, report CBAMReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Embedded Emissions Calculation")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "Embedded emissions represent the total greenhouse gas emissions released during the "+
		"production of imported goods, including both direct emissions from production processes and indirect "+
		"emissions from electricity consumption.", "", "", false)
	pdf.Ln(8)

	// Calculation methodology
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Emission Calculation Approach")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(170, 5, "CBAM embedded emissions are calculated using one of three approaches:\n\n"+
		"1. Actual Emissions: Based on verified monitoring data from the production installation\n"+
		"2. Default Values: EU Commission default values for specific production routes\n"+
		"3. Hybrid Approach: Combination of actual and default values\n\n"+
		"This report primarily uses actual emissions data where available, supplemented by default values "+
		"for installations without verified monitoring systems.", "", "", false)

	pdf.Ln(8)

	// Emissions breakdown
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(170, 7, "Emissions by Source")
	pdf.Ln(9)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(230, 235, 240)
	pdf.CellFormat(100, 6, "Emission Source", "1", 0, "L", true, 0, "")
	pdf.CellFormat(70, 6, "Emissions (tonnes CO2e)", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(248, 250, 252)

	pdf.CellFormat(100, 5, "Direct Process Emissions (Scope 1)", "1", 0, "L", true, 0, "")
	pdf.CellFormat(70, 5, fmt.Sprintf("%.2f", report.Report.Scope1EmissionsTonnes), "1", 1, "R", true, 0, "")

	pdf.CellFormat(100, 5, "Indirect Emissions - Electricity (Scope 2)", "1", 0, "L", false, 0, "")
	pdf.CellFormat(70, 5, fmt.Sprintf("%.2f", report.Report.Scope2EmissionsTonnes), "1", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(100, 6, "Total Embedded Emissions", "1", 0, "L", true, 0, "")
	pdf.CellFormat(70, 6, fmt.Sprintf("%.2f", report.Report.TotalEmissionsTonnes), "1", 1, "R", true, 0, "")

	pdf.Ln(6)

	// Emission intensity
	if len(report.ImportedGoods) > 0 {
		totalQuantity := 0.0
		for _, good := range report.ImportedGoods {
			totalQuantity += good.Quantity
		}

		avgIntensity := report.Report.TotalEmissionsTonnes / totalQuantity

		pdf.SetFont("Arial", "", 9)
		pdf.Cell(100, 5, "Average Emission Intensity:")
		pdf.SetFont("Arial", "B", 9)
		pdf.Cell(70, 5, fmt.Sprintf("%.4f tCO2e per unit", avgIntensity))
	}
}

func addCBAMProductionSites(pdf *gofpdf.Fpdf, report CBAMReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Production Sites and Installations")
	pdf.Ln(12)

	if len(report.ProductionSites) == 0 {
		pdf.SetFont("Arial", "", 11)
		pdf.Cell(170, 7, "No production site information available.")
		return
	}

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(170, 5, "CBAM requires detailed information on production installations where covered goods are manufactured. "+
		"The following sites are included in this report:", "", "", false)
	pdf.Ln(6)

	for i, site := range report.ProductionSites {
		if i >= 10 {
			pdf.SetFont("Arial", "I", 9)
			pdf.Cell(170, 5, fmt.Sprintf("... and %d additional sites", len(report.ProductionSites)-10))
			break
		}

		// Site info box
		pdf.SetDrawColor(180, 180, 180)
		pdf.SetFillColor(250, 252, 255)
		pdf.Rect(20, pdf.GetY(), 170, 35, "FD")

		y := pdf.GetY() + 3
		pdf.SetFont("Arial", "B", 10)
		pdf.SetXY(25, y)
		pdf.Cell(160, 5, fmt.Sprintf("Site %d: %s", i+1, site.Name))

		y += 6
		pdf.SetFont("Arial", "", 9)
		pdf.SetXY(30, y)
		pdf.Cell(80, 4, fmt.Sprintf("Location: %s, %s", site.Location, site.Country))

		y += 5
		pdf.SetXY(30, y)
		pdf.Cell(80, 4, fmt.Sprintf("Production Process: %s", site.ProductionProcess))

		y += 5
		pdf.SetXY(30, y)
		pdf.Cell(80, 4, fmt.Sprintf("Emission Factor: %.3f tCO2e/unit", site.EmissionsFactor))

		y += 5
		pdf.SetXY(30, y)
		certified := "No"
		if site.Certified {
			certified = fmt.Sprintf("Yes (%s)", site.CertificationBody)
		}
		pdf.Cell(80, 4, fmt.Sprintf("Third-Party Certified: %s", certified))

		pdf.SetY(pdf.GetY() + 35)
		pdf.Ln(3)
	}
}

func addCBAMPricing(pdf *gofpdf.Fpdf, report CBAMReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "CBAM Carbon Pricing and Obligations")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "The Carbon Border Adjustment Mechanism requires importers to purchase CBAM certificates "+
		"at a price corresponding to the weekly average closing price of EU ETS allowances. The number of certificates "+
		"required equals the embedded emissions minus any carbon price already paid in the country of origin.", "", "", false)
	pdf.Ln(8)

	// Pricing summary
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(170, 7, "CBAM Certificate Obligations")
	pdf.Ln(9)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(240, 245, 250)
	pdf.CellFormat(100, 6, "Description", "1", 0, "L", true, 0, "")
	pdf.CellFormat(70, 6, "Value", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(250, 252, 255)

	pdf.CellFormat(100, 5, "Total Embedded Emissions", "1", 0, "L", true, 0, "")
	pdf.CellFormat(70, 5, fmt.Sprintf("%.2f tonnes CO2e", report.Report.TotalEmissionsTonnes), "1", 1, "R", true, 0, "")

	pdf.CellFormat(100, 5, "Carbon Price Already Paid (origin countries)", "1", 0, "L", false, 0, "")
	pdf.CellFormat(70, 5, "€0.00", "1", 1, "R", false, 0, "")

	pdf.CellFormat(100, 5, "Net CBAM Obligations", "1", 0, "L", true, 0, "")
	pdf.CellFormat(70, 5, fmt.Sprintf("%.2f tonnes CO2e", report.Report.TotalEmissionsTonnes), "1", 1, "R", true, 0, "")

	avgPrice := 85.00 // Simplified - would use actual EU ETS price
	pdf.CellFormat(100, 5, "Average CBAM Certificate Price (2024)", "1", 0, "L", false, 0, "")
	pdf.CellFormat(70, 5, fmt.Sprintf("€%.2f per tonne", avgPrice), "1", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(100, 6, "Total CBAM Liability", "1", 0, "L", true, 0, "")
	pdf.SetTextColor(0, 100, 0)
	pdf.CellFormat(70, 6, fmt.Sprintf("€%.2f", report.TotalCarbonPrice), "1", 1, "R", true, 0, "")
	pdf.SetTextColor(0, 0, 0)

	pdf.Ln(8)

	pdf.SetFont("Arial", "I", 8)
	pdf.MultiCell(170, 4, "Note: CBAM certificates must be surrendered quarterly. The certificate price is based on "+
		"the weekly average closing price of EU ETS allowances. Failure to surrender sufficient certificates results in penalties.", "", "", false)
}

func addCBAMVerification(pdf *gofpdf.Fpdf, report CBAMReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Verification and Data Assurance")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(170, 5, "CBAM regulations require that embedded emissions data be verified by an accredited verifier "+
		"for imports exceeding certain thresholds. This section details the verification approach and data quality assessment.", "", "", false)
	pdf.Ln(8)

	// Data quality
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
	pdf.Cell(95, 6, "Verified Production Sites:")
	pdf.SetFont("Arial", "B", 10)
	certifiedCount := 0
	for _, site := range report.ProductionSites {
		if site.Certified {
			certifiedCount++
		}
	}
	pdf.Cell(75, 6, fmt.Sprintf("%d of %d", certifiedCount, len(report.ProductionSites)))
	pdf.Ln(12)

	// Verification statement
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(170, 7, "Verification Statement")
	pdf.Ln(9)

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(170, 5, "This CBAM report has been prepared based on the best available data from production installations. "+
		"Where actual monitoring data is unavailable, EU Commission default values have been applied in accordance with "+
		"CBAM implementing regulations. Third-party verification will be obtained for final quarterly submissions.", "", "", false)
}

func addCBAMMethodology(pdf *gofpdf.Fpdf, report CBAMReport) {
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(170, 10, "Calculation Methodology")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Regulatory Framework")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(170, 5, "This CBAM report has been prepared in accordance with:\n"+
		"• Regulation (EU) 2023/956 establishing a carbon border adjustment mechanism\n"+
		"• Commission Implementing Regulation on CBAM reporting obligations\n"+
		"• EU ETS Monitoring and Reporting Regulation (MRR)\n"+
		"• ISO 14064-1:2018 for GHG quantification", "", "", false)

	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(170, 8, "Emission Boundaries")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(170, 5, "CBAM emissions include:\n"+
		"• Direct emissions from production processes (combustion, chemical reactions)\n"+
		"• Indirect emissions from electricity consumption at production sites\n"+
		"• Precursor emissions where applicable (e.g., electricity used in aluminum production)", "", "", false)

	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(170, 7, "Report Hash")
	pdf.Ln(9)

	pdf.SetFont("Arial", "", 9)
	pdf.Cell(170, 5, "Report Integrity Hash (SHA-256):")
	pdf.Ln(6)

	pdf.SetFont("Courier", "", 7)
	pdf.Cell(170, 4, report.Report.ReportHash)
	pdf.Ln(8)

	pdf.SetFont("Arial", "I", 8)
	pdf.MultiCell(170, 4, "This cryptographic hash ensures data integrity and can be used to verify that report contents "+
		"have not been altered after generation.", "", "", false)
}
