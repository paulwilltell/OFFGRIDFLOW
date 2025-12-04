package pdf

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

// ReportData contains structured data for PDF report generation
type ReportData struct {
	Title          string
	EntityName     string
	ReportingStart time.Time
	ReportingEnd   time.Time
	Sections       []Section
	Metadata       map[string]string
}

// Section represents a section in the PDF report
type Section struct {
	Title   string
	Content []string
	Tables  []Table
}

// Table represents tabular data
type Table struct {
	Headers []string
	Rows    [][]string
}

// Generator exports data into PDF format with proper formatting
type Generator struct{}

// NewGenerator creates a new PDF generator
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate renders a PDF document with proper structure and formatting
func (g *Generator) Generate(ctx context.Context, data interface{}) ([]byte, error) {
	reportData, ok := data.(*ReportData)
	if !ok {
		// Fallback for simple data
		return g.generateSimple(ctx, data)
	}

	var buf bytes.Buffer
	offsets := make([]int, 0)
	objectNum := 1

	// PDF Header
	buf.WriteString("%PDF-1.7\n")
	buf.WriteString("%âãÏÓ\n") // Binary marker

	// Catalog
	offsets = append(offsets, buf.Len())
	buf.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /Catalog /Pages 2 0 R /Metadata 3 0 R >>\nendobj\n", objectNum))
	objectNum++

	// Pages object
	offsets = append(offsets, buf.Len())
	buf.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /Pages /Count 1 /Kids [4 0 R] >>\nendobj\n", objectNum))
	objectNum++

	// Metadata
	offsets = append(offsets, buf.Len())
	metadata := g.buildMetadata(reportData)
	buf.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /Metadata /Subtype /XML /Length %d >>\nstream\n%s\nendstream\nendobj\n",
		objectNum, len(metadata), metadata))
	objectNum++

	// Page object
	offsets = append(offsets, buf.Len())
	buf.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 5 0 R /Resources << /Font << /F1 6 0 R /F2 7 0 R >> >> >>\nendobj\n", objectNum))
	objectNum++

	// Content stream
	content := g.buildContent(reportData)
	offsets = append(offsets, buf.Len())
	buf.WriteString(fmt.Sprintf("%d 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n",
		objectNum, len(content), content))
	objectNum++

	// Font objects
	offsets = append(offsets, buf.Len())
	buf.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n", objectNum))
	objectNum++

	offsets = append(offsets, buf.Len())
	buf.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>\nendobj\n", objectNum))
	objectNum++

	// Cross-reference table
	xrefPos := buf.Len()
	buf.WriteString(fmt.Sprintf("xref\n0 %d\n", objectNum))
	buf.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offset))
	}

	// Trailer
	buf.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		objectNum, xrefPos))

	return buf.Bytes(), nil
}

// buildMetadata creates XMP metadata for the PDF
func (g *Generator) buildMetadata(data *ReportData) string {
	created := time.Now().UTC().Format(time.RFC3339)
	return fmt.Sprintf(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
      <dc:title>%s</dc:title>
      <dc:creator>OffGridFlow</dc:creator>
      <dc:description>Carbon Emissions Report</dc:description>
    </rdf:Description>
    <rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/">
      <xmp:CreateDate>%s</xmp:CreateDate>
      <xmp:CreatorTool>OffGridFlow v1.0</xmp:CreatorTool>
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`, escape(data.Title), created)
}

// buildContent creates the PDF content stream
func (g *Generator) buildContent(data *ReportData) string {
	var content strings.Builder

	content.WriteString("BT\n")

	y := 750.0

	// Title
	content.WriteString("/F2 18 Tf\n")
	content.WriteString(fmt.Sprintf("72 %.2f Td\n", y))
	content.WriteString(fmt.Sprintf("(%s) Tj\n", escape(data.Title)))
	y -= 30

	// Entity and period
	content.WriteString("/F1 12 Tf\n")
	content.WriteString(fmt.Sprintf("0 %.2f Td\n", -30.0))
	content.WriteString(fmt.Sprintf("(%s) Tj\n", escape(data.EntityName)))
	y -= 20

	content.WriteString(fmt.Sprintf("0 %.2f Td\n", -20.0))
	periodText := fmt.Sprintf("Reporting Period: %s to %s",
		data.ReportingStart.Format("2006-01-02"),
		data.ReportingEnd.Format("2006-01-02"))
	content.WriteString(fmt.Sprintf("(%s) Tj\n", escape(periodText)))
	y -= 40

	// Sections
	for _, section := range data.Sections {
		// Section title
		content.WriteString("/F2 14 Tf\n")
		content.WriteString(fmt.Sprintf("0 %.2f Td\n", -20.0))
		content.WriteString(fmt.Sprintf("(%s) Tj\n", escape(section.Title)))
		y -= 25

		// Section content
		content.WriteString("/F1 11 Tf\n")
		for _, line := range section.Content {
			content.WriteString(fmt.Sprintf("0 %.2f Td\n", -15.0))
			content.WriteString(fmt.Sprintf("(%s) Tj\n", escape(line)))
			y -= 15
		}

		// Tables
		for _, table := range section.Tables {
			y -= 10
			// Table headers
			content.WriteString("/F2 10 Tf\n")
			content.WriteString(fmt.Sprintf("0 %.2f Td\n", -15.0))
			content.WriteString(fmt.Sprintf("(%s) Tj\n", escape(strings.Join(table.Headers, " | "))))
			y -= 15

			// Table rows
			content.WriteString("/F1 10 Tf\n")
			for _, row := range table.Rows {
				content.WriteString(fmt.Sprintf("0 %.2f Td\n", -12.0))
				content.WriteString(fmt.Sprintf("(%s) Tj\n", escape(strings.Join(row, " | "))))
				y -= 12
			}
			y -= 10
		}
	}

	// Footer
	content.WriteString("/F1 8 Tf\n")
	content.WriteString(fmt.Sprintf("0 %.2f Td\n", -(y - 50)))
	content.WriteString(fmt.Sprintf("(Generated by OffGridFlow on %s) Tj\n",
		escape(time.Now().Format("2006-01-02 15:04:05"))))

	content.WriteString("ET\n")

	return content.String()
}

// generateSimple creates a simple PDF for unstructured data
func (g *Generator) generateSimple(ctx context.Context, data interface{}) ([]byte, error) {
	summary := fmt.Sprintf("%v", data)
	created := time.Now().UTC().Format(time.RFC3339)

	text := fmt.Sprintf("OffGridFlow Report\\nGenerated: %s\\n\\n%s", created, summary)

	var buf bytes.Buffer
	offsets := make([]int, 6)

	buf.WriteString("%PDF-1.4\n")
	offsets[1] = buf.Len()
	buf.WriteString("1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\n")

	offsets[2] = buf.Len()
	buf.WriteString("2 0 obj << /Type /Pages /Count 1 /Kids [3 0 R] >> endobj\n")

	offsets[3] = buf.Len()
	buf.WriteString("3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >> endobj\n")

	content := fmt.Sprintf("BT /F1 12 Tf 72 720 Td (%s) Tj ET", escape(text))
	offsets[4] = buf.Len()
	buf.WriteString(fmt.Sprintf("4 0 obj << /Length %d >> stream\n%s\nendstream endobj\n", len(content), content))

	offsets[5] = buf.Len()
	buf.WriteString("5 0 obj << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> endobj\n")

	xrefPos := buf.Len()
	buf.WriteString("xref\n0 6\n")
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= 5; i++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	buf.WriteString("trailer << /Size 6 /Root 1 0 R >>\nstartxref\n")
	buf.WriteString(fmt.Sprintf("%d\n%%%%EOF", xrefPos))

	return buf.Bytes(), nil
}

// escape escapes special characters in PDF text objects
func escape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

