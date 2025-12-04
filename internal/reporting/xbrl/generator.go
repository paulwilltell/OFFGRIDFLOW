package xbrl

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"time"
)

// XBRL namespace constants
const (
	XBRLNamespace     = "http://www.xbrl.org/2003/instance"
	XLinkNamespace    = "http://www.w3.org/1999/xlink"
	ISONamespace      = "http://www.xbrl.org/2003/iso4217"
	GHGNamespace      = "http://xbrl.org/ghg/2023"
	CSRDNamespace     = "http://www.esma.europa.eu/taxonomy/2023-12-31/esef_cor"
)

// Document represents an XBRL instance document for sustainability reporting
type Document struct {
	XMLName   xml.Name  `xml:"xbrl"`
	XBRLns    string    `xml:"xmlns:xbrl,attr"`
	XLinkns   string    `xml:"xmlns:xlink,attr"`
	ISOns     string    `xml:"xmlns:iso4217,attr"`
	GHGns     string    `xml:"xmlns:ghg,attr"`
	CSRDns    string    `xml:"xmlns:csrd,attr"`
	SchemaRef SchemaRef `xml:"schemaRef"`
	Context   Context   `xml:"context"`
	Unit      []Unit    `xml:"unit"`
	Facts     []Fact    `xml:",any"`
}

// SchemaRef references the taxonomy schema
type SchemaRef struct {
	XMLName xml.Name `xml:"schemaRef"`
	Type    string   `xml:"xlink:type,attr"`
	Href    string   `xml:"xlink:href,attr"`
}

// Context defines the reporting context
type Context struct {
	XMLName    xml.Name `xml:"context"`
	ID         string   `xml:"id,attr"`
	Entity     Entity   `xml:"entity"`
	Period     Period   `xml:"period"`
}

// Entity identifies the reporting entity
type Entity struct {
	Identifier Identifier `xml:"identifier"`
}

// Identifier contains entity identification
type Identifier struct {
	Scheme string `xml:"scheme,attr"`
	Value  string `xml:",chardata"`
}

// Period defines the reporting period
type Period struct {
	StartDate string `xml:"startDate,omitempty"`
	EndDate   string `xml:"endDate,omitempty"`
	Instant   string `xml:"instant,omitempty"`
}

// Unit defines measurement units
type Unit struct {
	XMLName xml.Name `xml:"unit"`
	ID      string   `xml:"id,attr"`
	Measure string   `xml:"measure"`
}

// Fact represents an XBRL fact
type Fact struct {
	XMLName   xml.Name
	ContextRef string  `xml:"contextRef,attr"`
	UnitRef   string  `xml:"unitRef,attr,omitempty"`
	Decimals  string  `xml:"decimals,attr,omitempty"`
	Value     string  `xml:",chardata"`
}

// EmissionsData contains emissions data for XBRL export
type EmissionsData struct {
	EntityName     string
	EntityID       string
	ReportingStart time.Time
	ReportingEnd   time.Time
	Scope1Total    float64
	Scope2Total    float64
	Scope3Total    float64
	TotalEmissions float64
	Currency       string
	RevenueIntensity float64
}

// Generator exports emissions data into XBRL format
type Generator struct{}

// NewGenerator creates a new XBRL generator
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates an XBRL instance document from emissions data
func (g *Generator) Generate(ctx context.Context, data interface{}) ([]byte, error) {
	emissionsData, ok := data.(*EmissionsData)
	if !ok {
		return nil, fmt.Errorf("invalid data type for XBRL generation")
	}

	doc := &Document{
		XBRLns:  XBRLNamespace,
		XLinkns: XLinkNamespace,
		ISOns:   ISONamespace,
		GHGns:   GHGNamespace,
		CSRDns:  CSRDNamespace,
		SchemaRef: SchemaRef{
			Type: "simple",
			Href: "http://xbrl.org/ghg/2023/ghg-2023-12-31.xsd",
		},
		Context: Context{
			ID: "Current_AsOf_" + emissionsData.ReportingEnd.Format("2006-01-02"),
			Entity: Entity{
				Identifier: Identifier{
					Scheme: "http://www.sec.gov/CIK",
					Value:  emissionsData.EntityID,
				},
			},
			Period: Period{
				StartDate: emissionsData.ReportingStart.Format("2006-01-02"),
				EndDate:   emissionsData.ReportingEnd.Format("2006-01-02"),
			},
		},
		Unit: []Unit{
			{
				ID:      "tCO2e",
				Measure: "ghg:MetricTonsCO2Equivalent",
			},
			{
				ID:      "usd",
				Measure: "iso4217:USD",
			},
		},
	}

	// Add emission facts
	contextRef := doc.Context.ID
	doc.Facts = []Fact{
		{
			XMLName:   xml.Name{Space: GHGNamespace, Local: "Scope1Emissions"},
			ContextRef: contextRef,
			UnitRef:   "tCO2e",
			Decimals:  "2",
			Value:     fmt.Sprintf("%.2f", emissionsData.Scope1Total/1000), // Convert kg to metric tons
		},
		{
			XMLName:   xml.Name{Space: GHGNamespace, Local: "Scope2LocationBasedEmissions"},
			ContextRef: contextRef,
			UnitRef:   "tCO2e",
			Decimals:  "2",
			Value:     fmt.Sprintf("%.2f", emissionsData.Scope2Total/1000),
		},
		{
			XMLName:   xml.Name{Space: GHGNamespace, Local: "Scope3Emissions"},
			ContextRef: contextRef,
			UnitRef:   "tCO2e",
			Decimals:  "2",
			Value:     fmt.Sprintf("%.2f", emissionsData.Scope3Total/1000),
		},
		{
			XMLName:   xml.Name{Space: GHGNamespace, Local: "TotalGHGEmissions"},
			ContextRef: contextRef,
			UnitRef:   "tCO2e",
			Decimals:  "2",
			Value:     fmt.Sprintf("%.2f", emissionsData.TotalEmissions/1000),
		},
		{
			XMLName:   xml.Name{Space: CSRDNamespace, Local: "EmissionsIntensityRevenue"},
			ContextRef: contextRef,
			UnitRef:   "tCO2e",
			Decimals:  "4",
			Value:     fmt.Sprintf("%.4f", emissionsData.RevenueIntensity),
		},
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("failed to encode XBRL document: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateFromMap creates XBRL from a generic map (backward compatibility)
func (g *Generator) GenerateFromMap(ctx context.Context, data map[string]interface{}) ([]byte, error) {
	// Convert map to EmissionsData
	emissionsData := &EmissionsData{
		EntityName:     getStringOrDefault(data, "entity_name", "Unknown Entity"),
		EntityID:       getStringOrDefault(data, "entity_id", "0000000000"),
		Scope1Total:    getFloatOrDefault(data, "scope1_total", 0),
		Scope2Total:    getFloatOrDefault(data, "scope2_total", 0),
		Scope3Total:    getFloatOrDefault(data, "scope3_total", 0),
		TotalEmissions: getFloatOrDefault(data, "total_emissions", 0),
		Currency:       getStringOrDefault(data, "currency", "USD"),
		RevenueIntensity: getFloatOrDefault(data, "revenue_intensity", 0),
	}

	if startDate, ok := data["reporting_start"].(time.Time); ok {
		emissionsData.ReportingStart = startDate
	} else {
		emissionsData.ReportingStart = time.Now().AddDate(-1, 0, 0)
	}

	if endDate, ok := data["reporting_end"].(time.Time); ok {
		emissionsData.ReportingEnd = endDate
	} else {
		emissionsData.ReportingEnd = time.Now()
	}

	return g.Generate(ctx, emissionsData)
}

func getStringOrDefault(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultValue
}

func getFloatOrDefault(m map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	if val, ok := m[key].(int); ok {
		return float64(val)
	}
	return defaultValue
}

