package compliance

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// ExportSummaryToPDF builds a simple PDF representation of the compliance summary.
func ExportSummaryToPDF(summary *ComplianceSummary) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Compliance Summary", false)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, "Compliance Overview", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 7, fmt.Sprintf("Generated: %s", time.Now().UTC().Format(time.RFC3339)), "", 1, "R", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "Framework Status", "", 1, "", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	for _, frame := range summary.Frameworks {
		pdf.CellFormat(0, 7, fmt.Sprintf("%s (%s)", frame.Name, frame.Status), "", 1, "", false, 0, "")
		var gaps string
		if len(frame.DataGaps) > 0 {
			gaps = fmt.Sprintf("Data gaps: %s", frame.DataGaps)
			pdf.CellFormat(0, 6, gaps, "", 1, "", false, 0, "")
		}
	}

	pdf.Ln(3)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "Emissions Totals (tCO2e)", "", 1, "", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 7, fmt.Sprintf("Scope 1: %.3f", summary.Totals.Scope1Tons), "", 1, "", false, 0, "")
	pdf.CellFormat(0, 7, fmt.Sprintf("Scope 2: %.3f", summary.Totals.Scope2Tons), "", 1, "", false, 0, "")
	pdf.CellFormat(0, 7, fmt.Sprintf("Scope 3: %.3f", summary.Totals.Scope3Tons), "", 1, "", false, 0, "")
	pdf.CellFormat(0, 7, fmt.Sprintf("Total: %.3f", summary.Totals.TotalTons), "", 1, "", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf export: %w", err)
	}

	return buf.Bytes(), nil
}

type xbrlDocument struct {
	XMLName    xml.Name         `xml:"xbrl"`
	Generated  string           `xml:"generatedAt"`
	Frameworks []xbrlFramework  `xml:"framework"`
	Totals     xbrlTotalsRecord `xml:"totals"`
}

type xbrlFramework struct {
	Name   string `xml:"name"`
	Status string `xml:"status"`
	Scope1 bool   `xml:"scope1Ready"`
	Scope2 bool   `xml:"scope2Ready"`
	Scope3 bool   `xml:"scope3Ready"`
}

type xbrlTotalsRecord struct {
	Scope1 float64 `xml:"scope1Tons"`
	Scope2 float64 `xml:"scope2Tons"`
	Scope3 float64 `xml:"scope3Tons"`
	Total  float64 `xml:"totalTons"`
}

// ExportSummaryToXBRL renders a lightweight XBRL-like XML payload for compliance summaries.
func ExportSummaryToXBRL(summary *ComplianceSummary) ([]byte, error) {
	doc := xbrlDocument{
		Generated: time.Now().UTC().Format(time.RFC3339),
		Totals: xbrlTotalsRecord{
			Scope1: summary.Totals.Scope1Tons,
			Scope2: summary.Totals.Scope2Tons,
			Scope3: summary.Totals.Scope3Tons,
			Total:  summary.Totals.TotalTons,
		},
	}

	for _, frame := range summary.Frameworks {
		doc.Frameworks = append(doc.Frameworks, xbrlFramework{
			Name:   frame.Name,
			Status: string(frame.Status),
			Scope1: frame.Scope1,
			Scope2: frame.Scope2,
			Scope3: frame.Scope3,
		})
	}

	data, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("xbrl export: %w", err)
	}
	return append([]byte(xml.Header), data...), nil
}
