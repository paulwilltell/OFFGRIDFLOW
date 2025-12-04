// Package parser provides parsing capabilities for various utility bill formats.
//
// This parser supports multiple file formats commonly used by utility providers:
//   - PDF (text extraction with pattern recognition)
//   - Excel/XLSX (structured data)
//   - CSV (comma-separated values)
//   - JSON (structured API responses)
//
// The parser implements intelligent extraction algorithms that can handle
// variations in formatting across different utility providers.
package parser

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/ingestion"
)

// =============================================================================
// Parser Configuration and Types
// =============================================================================

// UtilityBillParser handles parsing of utility bills from various formats.
type UtilityBillParser struct {
	// DefaultOrgID is used when the file doesn't specify an organization
	DefaultOrgID string

	// DefaultLocation is used when location cannot be determined from the bill
	DefaultLocation string

	// StrictMode when true causes parsing to fail on any validation errors
	StrictMode bool

	// MaxFileSize in bytes (default 50MB)
	MaxFileSize int64
}

// NewUtilityBillParser creates a parser with sensible defaults.
func NewUtilityBillParser(defaultOrgID, defaultLocation string) *UtilityBillParser {
	return &UtilityBillParser{
		DefaultOrgID:    defaultOrgID,
		DefaultLocation: defaultLocation,
		StrictMode:      false,
		MaxFileSize:     50 * 1024 * 1024, // 50MB
	}
}

// ParseResult contains the parsing results including activities and any errors.
type ParseResult struct {
	Activities []ingestion.Activity
	Errors     []ingestion.ImportError
	Metadata   map[string]string
}

// =============================================================================
// Format Detection and Routing
// =============================================================================

// FileFormat represents supported utility bill file formats.
type FileFormat string

const (
	FormatPDF   FileFormat = "pdf"
	FormatExcel FileFormat = "xlsx"
	FormatCSV   FileFormat = "csv"
	FormatJSON  FileFormat = "json"
	FormatXLS   FileFormat = "xls"
	FormatXML   FileFormat = "xml"
)

// DetectFormat identifies the file format from filename and content.
func (p *UtilityBillParser) DetectFormat(filename string, content []byte) (FileFormat, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		return FormatPDF, nil
	case ".xlsx":
		return FormatExcel, nil
	case ".xls":
		return FormatXLS, nil
	case ".csv":
		return FormatCSV, nil
	case ".json":
		return FormatJSON, nil
	case ".xml":
		return FormatXML, nil
	}

	// Fallback: detect by content signature
	if len(content) < 4 {
		return "", fmt.Errorf("file too small to determine format")
	}

	// PDF signature
	if bytes.HasPrefix(content, []byte("%PDF")) {
		return FormatPDF, nil
	}

	// Excel signature (ZIP archive)
	if bytes.HasPrefix(content, []byte("PK\x03\x04")) {
		return FormatExcel, nil
	}

	// JSON signature
	trimmed := bytes.TrimSpace(content)
	if bytes.HasPrefix(trimmed, []byte("{")) || bytes.HasPrefix(trimmed, []byte("[")) {
		return FormatJSON, nil
	}

	// XML signature
	if bytes.HasPrefix(trimmed, []byte("<?xml")) || bytes.HasPrefix(trimmed, []byte("<")) {
		return FormatXML, nil
	}

	// Default to CSV for plain text
	return FormatCSV, nil
}

// Parse is the main entry point that auto-detects format and routes to appropriate parser.
func (p *UtilityBillParser) Parse(ctx context.Context, filename string, content io.Reader) (*ParseResult, error) {
	// Read content into buffer for format detection
	buf := &bytes.Buffer{}
	limited := io.LimitReader(content, p.MaxFileSize)
	n, err := io.Copy(buf, limited)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	if n >= p.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum of %d bytes", p.MaxFileSize)
	}

	data := buf.Bytes()

	// Detect format
	format, err := p.DetectFormat(filename, data)
	if err != nil {
		return nil, fmt.Errorf("detect format: %w", err)
	}

	// Route to appropriate parser
	switch format {
	case FormatPDF:
		return p.parsePDF(ctx, data)
	case FormatExcel, FormatXLS:
		return p.parseExcel(ctx, data)
	case FormatCSV:
		return p.parseCSV(ctx, bytes.NewReader(data))
	case FormatJSON:
		return p.parseJSON(ctx, data)
	case FormatXML:
		return p.parseXML(ctx, data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// ParseMultipart parses a multipart file upload.
func (p *UtilityBillParser) ParseMultipart(ctx context.Context, fileHeader *multipart.FileHeader) (*ParseResult, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("open multipart file: %w", err)
	}
	defer file.Close()

	return p.Parse(ctx, fileHeader.Filename, file)
}

// =============================================================================
// CSV Parser
// =============================================================================

// parseCSV handles CSV formatted utility bills.
//
// Expected formats:
//  1. Standard format with headers: meter_id,location,period_start,period_end,quantity,unit,category
//  2. Simple format: meter_id,start_date,end_date,kwh
//  3. Extended format with additional metadata columns
func (p *UtilityBillParser) parseCSV(ctx context.Context, r io.Reader) (*ParseResult, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV header: %w", err)
	}

	// Normalize and map headers
	colIndex := make(map[string]int)
	for i, col := range header {
		normalized := normalizeColumnName(col)
		colIndex[normalized] = i
	}

	// Detect CSV schema
	schema := p.detectCSVSchema(colIndex)
	// Validate required columns
	var missing []string
	if schema.MeterIDCol == "" {
		missing = append(missing, "meter_id")
	}
	if schema.StartCol == "" {
		missing = append(missing, "period_start")
	}
	if schema.EndCol == "" {
		missing = append(missing, "period_end")
	}
	if schema.QuantityCol == "" {
		missing = append(missing, "quantity/kwh")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required columns: %s", strings.Join(missing, ", "))
	}

	var (
		activities []ingestion.Activity
		errors     []ingestion.ImportError
		lineNum    = 1 // Header is line 1
	)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		lineNum++

		if err != nil {
			errors = append(errors, ingestion.ImportError{
				Row:     lineNum,
				Message: fmt.Sprintf("CSV parse error: %v", err),
			})
			continue
		}

		// Skip empty rows
		if isEmptyRecord(record) {
			continue
		}

		activity, parseErr := p.parseCSVRecord(schema, colIndex, record, lineNum)
		if parseErr != nil {
			errors = append(errors, *parseErr)
			continue
		}

		activities = append(activities, *activity)
	}

	result := &ParseResult{
		Activities: activities,
		Errors:     errors,
		Metadata: map[string]string{
			"format":      "csv",
			"total_rows":  strconv.Itoa(lineNum - 1),
			"parsed_rows": strconv.Itoa(len(activities)),
			"error_rows":  strconv.Itoa(len(errors)),
			"schema_type": schema.Name,
		},
	}

	if p.StrictMode && len(errors) > 0 {
		return result, fmt.Errorf("parsing failed with %d errors in strict mode", len(errors))
	}

	return result, nil
}

// CSVSchema defines the structure of a CSV file.
type CSVSchema struct {
	Name        string
	MeterIDCol  string
	LocationCol string
	StartCol    string
	EndCol      string
	QuantityCol string
	UnitCol     string
	CategoryCol string
	OrgIDCol    string
	AddressCol  string
	ProviderCol string
	AccountCol  string
	InvoiceCol  string
}

// detectCSVSchema infers the schema from column headers.
func (p *UtilityBillParser) detectCSVSchema(colIndex map[string]int) *CSVSchema {
	schema := &CSVSchema{Name: "standard"}

	// Map common column name variations
	columnMappings := map[string][]string{
		"meter_id": {"meter_id", "meter", "meter_number", "account_number", "service_id"},
		"location": {"location", "region", "zone", "area", "grid_region"},
		"start":    {"period_start", "start_date", "from_date", "bill_start", "service_start"},
		"end":      {"period_end", "end_date", "to_date", "bill_end", "service_end"},
		"quantity": {"quantity", "kwh", "usage", "consumption", "amount", "energy"},
		"unit":     {"unit", "uom", "unit_of_measure"},
		"category": {"category", "type", "utility_type", "service_type"},
		"org_id":   {"org_id", "organization_id", "customer_id"},
		"address":  {"address", "service_address", "location_address"},
		"provider": {"provider", "utility_provider", "supplier", "utility_company"},
		"account":  {"account", "account_number", "customer_account"},
		"invoice":  {"invoice", "invoice_number", "bill_number"},
	}

	for field, variants := range columnMappings {
		for _, variant := range variants {
			if _, exists := colIndex[variant]; exists {
				switch field {
				case "meter_id":
					schema.MeterIDCol = variant
				case "location":
					schema.LocationCol = variant
				case "start":
					schema.StartCol = variant
				case "end":
					schema.EndCol = variant
				case "quantity":
					schema.QuantityCol = variant
				case "unit":
					schema.UnitCol = variant
				case "category":
					schema.CategoryCol = variant
				case "org_id":
					schema.OrgIDCol = variant
				case "address":
					schema.AddressCol = variant
				case "provider":
					schema.ProviderCol = variant
				case "account":
					schema.AccountCol = variant
				case "invoice":
					schema.InvoiceCol = variant
				}
				break
			}
		}
	}

	return schema
}

// parseCSVRecord converts a CSV record into an Activity.
func (p *UtilityBillParser) parseCSVRecord(schema *CSVSchema, colIndex map[string]int, record []string, lineNum int) (*ingestion.Activity, *ingestion.ImportError) {
	get := func(colName string) string {
		if colName == "" {
			return ""
		}
		if idx, ok := colIndex[colName]; ok && idx < len(record) {
			return strings.TrimSpace(record[idx])
		}
		return ""
	}

	// Extract required fields
	meterID := get(schema.MeterIDCol)
	if meterID == "" {
		return nil, &ingestion.ImportError{
			Row:     lineNum,
			Field:   "meter_id",
			Message: "meter_id is required but not found",
		}
	}

	// Parse dates
	startStr := get(schema.StartCol)
	start, err := parseFlexibleDate(startStr)
	if err != nil {
		return nil, &ingestion.ImportError{
			Row:     lineNum,
			Field:   schema.StartCol,
			Message: fmt.Sprintf("invalid start date %q: %v", startStr, err),
		}
	}

	endStr := get(schema.EndCol)
	end, err := parseFlexibleDate(endStr)
	if err != nil {
		return nil, &ingestion.ImportError{
			Row:     lineNum,
			Field:   schema.EndCol,
			Message: fmt.Sprintf("invalid end date %q: %v", endStr, err),
		}
	}

	// Parse quantity
	quantityStr := get(schema.QuantityCol)
	quantity, err := parseFlexibleNumber(quantityStr)
	if err != nil {
		return nil, &ingestion.ImportError{
			Row:     lineNum,
			Field:   schema.QuantityCol,
			Message: fmt.Sprintf("invalid quantity %q: %v", quantityStr, err),
		}
	}

	// Determine unit (default to kWh for electricity)
	unit := get(schema.UnitCol)
	if unit == "" {
		unit = "kWh" // Default assumption for utility bills
	}

	// Determine category
	category := get(schema.CategoryCol)
	if category == "" {
		category = "electricity" // Default
	}

	// Get location
	location := get(schema.LocationCol)
	if location == "" {
		location = p.DefaultLocation
	}

	// Get organization
	orgID := get(schema.OrgIDCol)
	if orgID == "" {
		orgID = p.DefaultOrgID
	}

	// Build metadata
	metadata := make(map[string]string)
	if address := get(schema.AddressCol); address != "" {
		metadata["address"] = address
	}
	if provider := get(schema.ProviderCol); provider != "" {
		metadata["provider"] = provider
	}
	if account := get(schema.AccountCol); account != "" {
		metadata["account"] = account
	}
	if invoice := get(schema.InvoiceCol); invoice != "" {
		metadata["invoice"] = invoice
	}

	// Create activity
	activity := &ingestion.Activity{
		ID:          fmt.Sprintf("csv-%s-%d-%d", meterID, start.Unix(), lineNum),
		Source:      string(ingestion.SourceUtilityBill),
		Category:    category,
		MeterID:     meterID,
		Location:    location,
		PeriodStart: start,
		PeriodEnd:   end,
		Quantity:    quantity,
		Unit:        unit,
		OrgID:       orgID,
		Metadata:    metadata,
		DataQuality: "measured",
		CreatedAt:   time.Now().UTC(),
	}

	// Validate
	if err := activity.Validate(); err != nil {
		return nil, &ingestion.ImportError{
			Row:     lineNum,
			Message: fmt.Sprintf("validation failed: %v", err),
		}
	}

	return activity, nil
}

// =============================================================================
// JSON Parser
// =============================================================================

// UtilityBillJSON represents the JSON structure for utility bills.
type UtilityBillJSON struct {
	Bills []struct {
		MeterID     string                 `json:"meter_id"`
		AccountID   string                 `json:"account_id,omitempty"`
		Location    string                 `json:"location,omitempty"`
		PeriodStart string                 `json:"period_start"`
		PeriodEnd   string                 `json:"period_end"`
		Quantity    float64                `json:"quantity"`
		Unit        string                 `json:"unit"`
		Category    string                 `json:"category,omitempty"`
		Provider    string                 `json:"provider,omitempty"`
		Invoice     string                 `json:"invoice,omitempty"`
		Address     string                 `json:"address,omitempty"`
		OrgID       string                 `json:"org_id,omitempty"`
		Metadata    map[string]interface{} `json:"metadata,omitempty"`
	} `json:"bills,omitempty"`

	// Alternative single-bill format
	MeterID     string                 `json:"meter_id,omitempty"`
	AccountID   string                 `json:"account_id,omitempty"`
	Location    string                 `json:"location,omitempty"`
	PeriodStart string                 `json:"period_start,omitempty"`
	PeriodEnd   string                 `json:"period_end,omitempty"`
	Quantity    float64                `json:"quantity,omitempty"`
	Unit        string                 `json:"unit,omitempty"`
	Category    string                 `json:"category,omitempty"`
	Provider    string                 `json:"provider,omitempty"`
	Invoice     string                 `json:"invoice,omitempty"`
	Address     string                 `json:"address,omitempty"`
	OrgID       string                 `json:"org_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// parseJSON handles JSON formatted utility bills.
func (p *UtilityBillParser) parseJSON(ctx context.Context, data []byte) (*ParseResult, error) {
	var billData UtilityBillJSON
	if err := json.Unmarshal(data, &billData); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	var (
		activities []ingestion.Activity
		errors     []ingestion.ImportError
	)

	// Handle array of bills
	if len(billData.Bills) > 0 {
		for i, bill := range billData.Bills {
			activity, err := p.convertJSONBill(bill.MeterID, bill.AccountID, bill.Location, bill.PeriodStart,
				bill.PeriodEnd, bill.Quantity, bill.Unit, bill.Category, bill.Provider,
				bill.Invoice, bill.Address, bill.OrgID, bill.Metadata, i+1)
			if err != nil {
				errors = append(errors, *err)
				continue
			}
			activities = append(activities, *activity)
		}
	} else if billData.MeterID != "" {
		// Handle single bill format
		activity, err := p.convertJSONBill(billData.MeterID, billData.AccountID, billData.Location,
			billData.PeriodStart, billData.PeriodEnd, billData.Quantity, billData.Unit,
			billData.Category, billData.Provider, billData.Invoice, billData.Address,
			billData.OrgID, billData.Metadata, 1)
		if err != nil {
			errors = append(errors, *err)
		} else {
			activities = append(activities, *activity)
		}
	} else {
		return nil, fmt.Errorf("invalid JSON structure: expected 'bills' array or single bill object")
	}

	result := &ParseResult{
		Activities: activities,
		Errors:     errors,
		Metadata: map[string]string{
			"format":       "json",
			"total_bills":  strconv.Itoa(len(billData.Bills)),
			"parsed_bills": strconv.Itoa(len(activities)),
			"error_bills":  strconv.Itoa(len(errors)),
		},
	}

	if p.StrictMode && len(errors) > 0 {
		return result, fmt.Errorf("parsing failed with %d errors in strict mode", len(errors))
	}

	return result, nil
}

// convertJSONBill converts a JSON bill structure to an Activity.
func (p *UtilityBillParser) convertJSONBill(meterID, accountID, location, periodStart, periodEnd string,
	quantity float64, unit, category, provider, invoice, address, orgID string,
	meta map[string]interface{}, index int) (*ingestion.Activity, *ingestion.ImportError) {

	// Parse dates
	start, err := parseFlexibleDate(periodStart)
	if err != nil {
		return nil, &ingestion.ImportError{
			Row:     index,
			Field:   "period_start",
			Message: fmt.Sprintf("invalid start date %q: %v", periodStart, err),
		}
	}

	end, err := parseFlexibleDate(periodEnd)
	if err != nil {
		return nil, &ingestion.ImportError{
			Row:     index,
			Field:   "period_end",
			Message: fmt.Sprintf("invalid end date %q: %v", periodEnd, err),
		}
	}

	// Set defaults
	if location == "" {
		location = p.DefaultLocation
	}
	if orgID == "" {
		orgID = p.DefaultOrgID
	}
	if unit == "" {
		unit = "kWh"
	}
	if category == "" {
		category = "electricity"
	}

	// Build metadata
	metadata := make(map[string]string)
	if provider != "" {
		metadata["provider"] = provider
	}
	if invoice != "" {
		metadata["invoice"] = invoice
	}
	if address != "" {
		metadata["address"] = address
	}
	if accountID != "" {
		metadata["account_id"] = accountID
	}

	// Add custom metadata
	for k, v := range meta {
		if str, ok := v.(string); ok {
			metadata[k] = str
		} else {
			metadata[k] = fmt.Sprintf("%v", v)
		}
	}

	activity := &ingestion.Activity{
		ID:          fmt.Sprintf("json-%s-%d", meterID, start.Unix()),
		Source:      string(ingestion.SourceUtilityBill),
		Category:    category,
		MeterID:     meterID,
		Location:    location,
		PeriodStart: start,
		PeriodEnd:   end,
		Quantity:    quantity,
		Unit:        unit,
		OrgID:       orgID,
		Metadata:    metadata,
		DataQuality: "measured",
		CreatedAt:   time.Now().UTC(),
	}

	if err := activity.Validate(); err != nil {
		return nil, &ingestion.ImportError{
			Row:     index,
			Message: fmt.Sprintf("validation failed: %v", err),
		}
	}

	return activity, nil
}

// =============================================================================
// PDF Parser - Production Implementation Required
// =============================================================================

// parsePDF extracts utility bill data from PDF files.
// Production implementation requires a PDF parsing library.
// Recommended libraries:
//   - github.com/pdfcpu/pdfcpu - For basic text extraction
//   - github.com/unidoc/unipdf - For advanced parsing and table extraction
//
// To enable PDF parsing:
// 1. Add dependency: go get github.com/pdfcpu/pdfcpu/pkg/api
// 2. Import: import "github.com/pdfcpu/pdfcpu/pkg/api"
// 3. Implement text extraction and pattern matching for your utility provider
func (p *UtilityBillParser) parsePDF(ctx context.Context, data []byte) (*ParseResult, error) {
	// Example implementation pattern (requires pdfcpu):
	// 1. Extract text: text, err := api.ExtractText(bytes.NewReader(data))
	// 2. Parse extracted text for utility bill patterns
	// 3. Match account number, meter ID, usage, cost patterns
	// 4. Return structured ParseResult
	
	return nil, fmt.Errorf("pdf parsing requires external library installation - run: go get github.com/pdfcpu/pdfcpu/pkg/api - or use CSV/JSON format")
}

// =============================================================================
// Excel Parser - Production Implementation Required
// =============================================================================

// parseExcel extracts utility bill data from Excel files.
// Production implementation requires the excelize library.
// Recommended library: github.com/xuri/excelize/v2
//
// To enable Excel parsing:
// 1. Add dependency: go get github.com/xuri/excelize/v2
// 2. Import: import "github.com/xuri/excelize/v2"
// 3. Implement sheet reading and cell pattern matching
func (p *UtilityBillParser) parseExcel(ctx context.Context, data []byte) (*ParseResult, error) {
	// Example implementation pattern (requires excelize):
	// 1. Open file: f, err := excelize.OpenReader(bytes.NewReader(data))
	// 2. Read rows: rows, err := f.GetRows("Sheet1")
	// 3. Parse rows for utility bill patterns
	// 4. Extract account, meter, usage, cost data
	// 5. Return structured ParseResult
	
	return nil, fmt.Errorf("excel parsing requires external library installation - run: go get github.com/xuri/excelize/v2 - or use CSV/JSON format")
}

// =============================================================================
// XML Parser - Production Implementation Required
// =============================================================================

// parseXML extracts utility bill data from XML files.
// Production implementation uses encoding/xml for EDI and XML-based formats.
// Common formats: EDI 810 (Invoice), EDI 867 (Usage Report), Green Button XML
//
// To enable XML parsing:
// 1. Define XML structs matching your utility provider's schema
// 2. Use encoding/xml to unmarshal
// 3. Map to ParseResult structure
func (p *UtilityBillParser) parseXML(ctx context.Context, data []byte) (*ParseResult, error) {
	// Example implementation pattern:
	// 1. Define XML structs for utility bill schema
	// 2. Unmarshal: err := xml.Unmarshal(data, &bill)
	// 3. Extract required fields
	// 4. Return structured ParseResult
	//
	// For Green Button XML, see: https://www.greenbuttondata.org/
	
	return nil, fmt.Errorf("xml parsing requires schema definition for your utility provider - use CSV/JSON format or implement custom XML unmarshaling")
}

// =============================================================================
// Utility Functions
// =============================================================================

// normalizeColumnName converts column headers to a standard format.
func normalizeColumnName(col string) string {
	// Convert to lowercase, remove special chars, replace spaces with underscores
	normalized := strings.ToLower(strings.TrimSpace(col))
	normalized = regexp.MustCompile(`[^a-z0-9_]+`).ReplaceAllString(normalized, "_")
	normalized = strings.Trim(normalized, "_")
	return normalized
}

// isEmptyRecord checks if a CSV record is empty.
func isEmptyRecord(record []string) bool {
	for _, field := range record {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}

// parseFlexibleDate attempts to parse a date string in various common formats.
func parseFlexibleDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// Common date formats used in utility bills
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"2006/01/02",
		"Jan 02, 2006",
		"January 02, 2006",
		"02-Jan-2006",
		"2006-01-02 15:04:05",
		"01/02/06",
		"2006-01-02T15:04:05Z07:00",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized date format: %q", dateStr)
}

// parseFlexibleNumber parses a number that may contain commas, spaces, or currency symbols.
func parseFlexibleNumber(numStr string) (float64, error) {
	numStr = strings.TrimSpace(numStr)
	if numStr == "" {
		return 0, fmt.Errorf("empty number string")
	}

	// Remove common formatting characters
	numStr = strings.ReplaceAll(numStr, ",", "")
	numStr = strings.ReplaceAll(numStr, " ", "")
	numStr = strings.TrimPrefix(numStr, "$")
	numStr = strings.TrimPrefix(numStr, "€")
	numStr = strings.TrimPrefix(numStr, "£")

	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q: %w", numStr, err)
	}

	return val, nil
}
