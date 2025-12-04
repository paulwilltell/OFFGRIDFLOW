# Utility Bills Ingestion Guide

## Overview

The Utility Bills ingestion module provides comprehensive capabilities for importing utility consumption data from various sources and formats into OffGridFlow. This data flows into the emissions calculation pipeline for accurate carbon accounting.

## Features

- **Multi-Format Support**: CSV, JSON, PDF (planned), Excel (planned), XML (planned)
- **Intelligent Parsing**: Auto-detection of file formats and flexible column mapping
- **Validation & Error Handling**: Comprehensive validation with detailed error reporting
- **Deduplication**: Prevents duplicate bill processing based on configurable time windows
- **Batch Processing**: Upload multiple bills concurrently with parallel processing
- **Data Enrichment**: Automatic category standardization and location mapping
- **RESTful API**: Complete HTTP API for integration with external systems

## Supported File Formats

### CSV Format

The most flexible format, supporting multiple column naming conventions:

```csv
meter_id,location,period_start,period_end,kwh,category,provider
METER-001,US-WEST,2025-01-01,2025-01-31,12500.5,electricity,PG&E
METER-002,EU-CENTRAL,2025-01-01,2025-01-31,8750.25,electricity,EDF
```

**Supported Column Names:**
- Meter ID: `meter_id`, `meter`, `meter_number`, `account_number`, `service_id`
- Location: `location`, `region`, `zone`, `area`, `grid_region`
- Start Date: `period_start`, `start_date`, `from_date`, `bill_start`
- End Date: `period_end`, `end_date`, `to_date`, `bill_end`
- Quantity: `quantity`, `kwh`, `usage`, `consumption`, `amount`
- Unit: `unit`, `uom`, `unit_of_measure` (defaults to kWh)
- Category: `category`, `type`, `utility_type`, `service_type`

**Date Formats Supported:**
- ISO 8601: `2025-01-01T00:00:00Z`
- Simple date: `2025-01-01`
- US format: `01/02/2025`
- EU format: `02/01/2025`
- Month names: `Jan 02, 2025`, `January 02, 2025`

### JSON Format

Structured format ideal for API integrations:

```json
{
  "bills": [
    {
      "meter_id": "METER-001",
      "location": "US-WEST",
      "period_start": "2025-01-01",
      "period_end": "2025-01-31",
      "quantity": 12500.5,
      "unit": "kWh",
      "category": "electricity",
      "provider": "PG&E",
      "invoice": "INV-2025-001",
      "metadata": {
        "rate_class": "commercial",
        "demand_charge": "152.00"
      }
    }
  ]
}
```

**Single Bill Format:**
```json
{
  "meter_id": "METER-001",
  "location": "US-WEST",
  "period_start": "2025-01-01",
  "period_end": "2025-01-31",
  "quantity": 12500.5,
  "unit": "kWh",
  "category": "electricity"
}
```

### PDF Format (Planned)

PDF parsing requires additional dependencies:
- `github.com/pdfcpu/pdfcpu` for basic text extraction
- `github.com/unidoc/unipdf` for advanced parsing

### Excel Format (Planned)

Excel parsing requires:
- `github.com/xuri/excelize/v2`

## API Endpoints

### Upload Single Bill

```
POST /api/ingestion/utility-bills/upload
Content-Type: multipart/form-data
```

**Form Parameters:**
- `file` (required): The utility bill file
- `org_id` (optional): Organization ID (can be from JWT)
- `strict` (optional): Enable strict validation mode (`true`/`false`)

**Response:**
```json
{
  "success": true,
  "filename": "january-2025.csv",
  "activities_count": 3,
  "errors_count": 0,
  "activities": [ /* Activity objects */ ]
}
```

**Example with curl:**
```bash
curl -X POST http://localhost:8080/api/ingestion/utility-bills/upload \
  -F "file=@bills/january-2025.csv" \
  -F "org_id=my-company" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Batch Upload

```
POST /api/ingestion/utility-bills/batch-upload
Content-Type: multipart/form-data
```

**Form Parameters:**
- `files[]` (required): Multiple utility bill files

**Response:**
```json
{
  "success": true,
  "total_files": 5,
  "success_files": 4,
  "failed_files": 1,
  "total_activities": 47,
  "total_errors": 2,
  "duration": "1.234s",
  "file_results": [
    {
      "filename": "january.csv",
      "success": true,
      "activities": 12,
      "errors": 0
    }
  ]
}
```

**Example with curl:**
```bash
curl -X POST http://localhost:8080/api/ingestion/utility-bills/batch-upload \
  -F "files=@bills/january-2025.csv" \
  -F "files=@bills/february-2025.csv" \
  -F "files=@bills/march-2025.csv" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### List Utility Bills

```
GET /api/ingestion/utility-bills
```

**Query Parameters:**
- `org_id` (required or from JWT): Organization ID
- `limit` (optional): Maximum records (default: 100)
- `source` (optional): Source filter (default: utility_bill)

**Response:**
```json
{
  "activities": [ /* Activity objects */ ],
  "count": 47,
  "source": "utility_bill",
  "org_id": "my-company"
}
```

## Programmatic Usage

### Basic Usage

```go
package main

import (
    "context"
    "os"

    "github.com/example/offgridflow/internal/ingestion"
    "github.com/example/offgridflow/internal/ingestion/sources/utility_bills"
)

func main() {
    // Create adapter with default configuration
    config := utility_bills.DefaultConfig("my-org-id")
    config.DefaultLocation = "US-WEST"
    config.EnableDeduplication = true

    adapter := utility_bills.NewAdapter(config)

    // Open a bill file
    file, _ := os.Open("january-2025.csv")
    defer file.Close()

    // Ingest the bill
    ctx := context.Background()
    activities, errors, err := adapter.IngestFile(ctx, "january-2025.csv", file)

    if err != nil {
        panic(err)
    }

    // Process results
    for _, activity := range activities {
        fmt.Printf("Meter %s: %.2f %s\n",
            activity.MeterID, activity.Quantity, activity.Unit)
    }

    // Handle import errors
    for _, importErr := range errors {
        fmt.Printf("Row %d: %s\n", importErr.Row, importErr.Message)
    }
}
```

### Advanced Configuration

```go
config := utility_bills.Config{
    DefaultOrgID:            "acme-corp",
    DefaultLocation:         "US-WEST",
    EnableDeduplication:     true,
    DeduplicationWindow:     90 * 24 * time.Hour,  // 90 days
    StrictValidation:        false,
    MaxConcurrentParsing:    4,
    MaxFileSize:             50 * 1024 * 1024,     // 50MB
    AutoEnrichLocation:      true,
    UtilityProviderMappings: map[string]string{
        "PG&E":      "US-CA-PGAE",
        "ConEd":     "US-NY-CONED",
        "EDF":       "EU-FR",
    },
}

adapter := utility_bills.NewAdapter(config)
```

### Batch Processing

```go
files := map[string]io.Reader{
    "january.csv":  reader1,
    "february.csv": reader2,
    "march.csv":    reader3,
}

result, err := adapter.IngestFiles(ctx, files)
if err != nil {
    panic(err)
}

fmt.Printf("Processed %d files in %v\n", result.TotalFiles, result.Duration())
fmt.Printf("Success: %d, Failed: %d\n", result.SuccessFiles, result.FailedFiles)
fmt.Printf("Total activities: %d\n", result.TotalActivities)
```

## Configuration Options

### Adapter Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `DefaultOrgID` | string | "" | Organization ID for bills without org specified |
| `DefaultLocation` | string | "US" | Default location/region |
| `EnableDeduplication` | bool | true | Prevent duplicate bill processing |
| `DeduplicationWindow` | duration | 90 days | How far back to check for duplicates |
| `StrictValidation` | bool | false | Fail entire import on any error |
| `MaxConcurrentParsing` | int | 4 | Parallel file processing limit |
| `MaxFileSize` | int64 | 50MB | Maximum file size |
| `AutoEnrichLocation` | bool | true | Auto-determine location from provider |
| `UtilityProviderMappings` | map | {} | Provider name to location mappings |

## Data Quality

### Activity Fields

Each ingested bill becomes an Activity with these fields:

```go
type Activity struct {
    ID          string            // Unique identifier
    Source      string            // "utility_bill"
    Category    string            // "electricity", "natural_gas", etc.
    MeterID     string            // Meter/account identifier
    Location    string            // Region code for emissions factors
    PeriodStart time.Time         // Billing period start
    PeriodEnd   time.Time         // Billing period end
    Quantity    float64           // Consumption amount
    Unit        string            // "kWh", "therm", etc.
    OrgID       string            // Organization identifier
    Metadata    map[string]string // Additional data
    DataQuality string            // "measured", "estimated", "default"
    CreatedAt   time.Time         // Import timestamp
}
```

### Data Quality Levels

- **measured**: Direct meter readings (highest quality)
- **estimated**: Estimated consumption (medium quality)
- **default**: Default/assumed values (lowest quality)

### Validation Rules

1. **Required Fields**: meter_id, period_start, period_end, quantity, org_id
2. **Quantity**: Must be non-negative
3. **Dates**: period_end must be after period_start
4. **Units**: Must be valid energy unit (kWh, MWh, GJ, therm, etc.)

## Error Handling

### Import Errors

Import errors are non-fatal and allow partial imports:

```go
type ImportError struct {
    Row        int    // Row number (1-indexed)
    Field      string // Field that caused error
    Message    string // Error description
    ExternalID string // External identifier if available
}
```

### Error Recovery

In non-strict mode, the adapter continues processing after errors:

```go
activities, errors, err := adapter.IngestFile(ctx, "bills.csv", file)

// err is only set for fatal errors (file can't be opened, etc.)
if err != nil {
    return err
}

// errors contains row-level validation failures
for _, importErr := range errors {
    log.Printf("Row %d failed: %s", importErr.Row, importErr.Message)
}

// activities contains successfully parsed bills
processActivities(activities)
```

## Best Practices

### 1. Use Batch Uploads for Efficiency

```go
// Good: Single batch upload
files := collectMonthlyBills()
result, _ := adapter.IngestFiles(ctx, files)

// Less efficient: Multiple single uploads
for _, file := range files {
    adapter.IngestFile(ctx, file.Name, file.Reader)
}
```

### 2. Enable Deduplication

```go
config := utility_bills.DefaultConfig("org")
config.EnableDeduplication = true  // Prevent re-processing bills
config.DeduplicationWindow = 90 * 24 * time.Hour
```

### 3. Configure Provider Mappings

```go
config.UtilityProviderMappings = map[string]string{
    "Pacific Gas & Electric": "US-CA-PGAE",
    "Con Edison":             "US-NY-CONED",
    "Duke Energy":            "US-NC-DUKE",
}
```

### 4. Handle Partial Failures

```go
result, err := adapter.IngestFiles(ctx, files)

// Log overall success
log.Printf("Success rate: %.1f%%", result.SuccessRate() * 100)

// Process individual file failures
for _, fr := range result.FileResults {
    if fr.Error != nil {
        log.Printf("File %s failed: %v", fr.Filename, fr.Error)
        notifyAdmin(fr.Filename, fr.Error)
    }
}
```

### 5. Implement Retry Logic

```go
const maxRetries = 3

for attempt := 0; attempt < maxRetries; attempt++ {
    _, errors, err := adapter.IngestFile(ctx, filename, file)

    if err == nil && len(errors) == 0 {
        break // Success
    }

    if attempt < maxRetries-1 {
        time.Sleep(time.Second * time.Duration(attempt+1))
        file.Seek(0, 0) // Reset file pointer
    }
}
```

## Integration Examples

### Web Upload Handler

```go
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    file, header, _ := r.FormFile("file")
    defer file.Close()

    activities, errors, err := adapter.IngestFile(
        r.Context(),
        header.Filename,
        file,
    )

    response := map[string]interface{}{
        "success":    err == nil,
        "activities": len(activities),
        "errors":     len(errors),
    }

    json.NewEncoder(w).Encode(response)
}
```

### Email Attachment Processing

```go
func processEmailAttachments(email *Email) error {
    files := make(map[string]io.Reader)

    for _, attachment := range email.Attachments {
        if isBillFile(attachment.Filename) {
            files[attachment.Filename] = attachment.Reader
        }
    }

    result, err := adapter.IngestFiles(context.Background(), files)

    if result.HasErrors() {
        notifyUser(email.From, result.Summary())
    }

    return err
}
```

### Scheduled Batch Import

```go
func scheduledImport() {
    files := scanBillsDirectory("./incoming")

    result, _ := adapter.IngestFiles(context.Background(), files)

    // Move processed files
    for _, fr := range result.FileResults {
        if fr.Error == nil {
            moveToArchive(fr.Filename)
        } else {
            moveToFailed(fr.Filename)
        }
    }

    // Send summary email
    emailReport(result.Summary())
}
```

## Troubleshooting

### Common Issues

**Issue**: "missing required column 'meter_id'"
- **Solution**: Ensure CSV has a column for meter ID (can be named: meter_id, meter, account_number, etc.)

**Issue**: "invalid date format"
- **Solution**: Use ISO 8601 format (2025-01-01) or one of the supported date formats

**Issue**: "file too large"
- **Solution**: Increase MaxFileSize in configuration or split the file

**Issue**: "validation failed in strict mode"
- **Solution**: Set StrictValidation to false to allow partial imports, or fix data quality issues

**Issue**: Duplicate bills being imported
- **Solution**: Enable deduplication: `config.EnableDeduplication = true`

## Performance Considerations

- **Batch Size**: Optimal batch size is 5-20 files for concurrent processing
- **File Size**: Files up to 50MB are handled efficiently
- **Concurrency**: Default 4 parallel parsers, increase for more cores
- **Memory**: ~10MB per file being processed in parallel

## Future Enhancements

- [ ] PDF text extraction and bill parsing
- [ ] Excel/XLSX file support
- [ ] XML/EDI format support
- [ ] Machine learning-based data extraction
- [ ] OCR for scanned bills
- [ ] Automated email fetching from utility providers
- [ ] Direct API integration with utility providers

## Support

For issues or questions:
- GitHub Issues: https://github.com/example/offgridflow/issues
- Documentation: https://docs.offgridflow.com
- Email: support@offgridflow.com
