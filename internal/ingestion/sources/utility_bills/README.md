# Utility Bills Ingestion Module

## Overview

This module provides comprehensive ingestion capabilities for utility bills from various sources and formats. It handles electricity, natural gas, water, and other utility consumption data, transforming it into standardized Activity records for emissions calculations.

## Quick Start

```go
import (
    "context"
    "os"
    "github.com/example/offgridflow/internal/ingestion/sources/utility_bills"
)

// Create adapter with default config
config := utility_bills.DefaultConfig("my-org-id")
adapter := utility_bills.NewAdapter(config)

// Ingest a bill file
file, _ := os.Open("utility-bill.csv")
defer file.Close()

activities, errors, err := adapter.IngestFile(
    context.Background(),
    "utility-bill.csv",
    file,
)

// Handle results
if err != nil {
    // Fatal error (file corrupt, etc.)
    panic(err)
}

// errors contains non-fatal row-level errors
// activities contains successfully parsed bills
```

## Architecture

### Components

```
utility_bills/
├── utility_bills.go       - Main adapter implementation
├── utility_bills_test.go  - Comprehensive test suite
└── README.md              - This file

Related files:
└── parser/
    └── utility_bills_parser.go - Multi-format parser
```

### Design

1. **Adapter** (`utility_bills.go`): High-level orchestration
   - Configuration management
   - File ingestion coordination
   - Batch processing
   - Deduplication
   - Data enrichment
   - Result aggregation

2. **Parser** (`parser/utility_bills_parser.go`): Format-specific parsing
   - Format detection
   - CSV parsing with flexible schemas
   - JSON parsing
   - PDF parsing (planned)
   - Excel parsing (planned)
   - Validation and error reporting

3. **Store**: Activity persistence (injected)
   - PostgreSQL implementation
   - In-memory implementation
   - Custom implementations supported

## Features

### Multi-Format Support

- ✅ **CSV**: Flexible column mapping, multiple naming conventions
- ✅ **JSON**: Array and single-object formats, rich metadata
- 🔜 **PDF**: Text extraction and pattern matching
- 🔜 **Excel**: XLSX and XLS support
- 🔜 **XML**: EDI and utility-specific formats

### Intelligent Parsing

- Auto-detection of file formats
- Flexible column name mapping (20+ variations)
- Multiple date format support
- Number parsing with currency and separators
- Schema detection and validation

### Validation & Quality

- Required field validation
- Data type validation
- Business rule validation
- Comprehensive error reporting
- Strict vs. lenient modes

### Deduplication

- Configurable time windows
- Meter + period-based keys
- Thread-safe cache
- Automatic cleanup

### Enrichment

- Category standardization
- Provider-to-location mapping
- Metadata injection
- Data quality scoring

### Performance

- Concurrent batch processing
- Semaphore-based throttling
- Efficient memory usage
- Configurable parallelism

## Configuration

### Basic Configuration

```go
config := utility_bills.DefaultConfig("my-org")
```

Provides:
- Default location: "US"
- Deduplication: enabled (90-day window)
- Strict validation: disabled
- Max concurrent: 4
- Max file size: 50MB
- Auto enrichment: enabled

### Advanced Configuration

```go
config := utility_bills.Config{
    DefaultOrgID:            "acme-corp",
    DefaultLocation:         "US-WEST",
    Store:                   myActivityStore,
    Logger:                  myLogger,
    EnableDeduplication:     true,
    DeduplicationWindow:     90 * 24 * time.Hour,
    StrictValidation:        false,
    MaxConcurrentParsing:    8,
    MaxFileSize:             100 * 1024 * 1024,
    AutoEnrichLocation:      true,
    UtilityProviderMappings: map[string]string{
        "Pacific Gas & Electric": "US-CA-PGAE",
        "Con Edison":             "US-NY-CONED",
        "Duke Energy":            "US-NC-DUKE",
    },
}

adapter := utility_bills.NewAdapter(config)
```

## API

### Core Methods

#### IngestFile

```go
func (a *Adapter) IngestFile(
    ctx context.Context,
    filename string,
    content io.Reader,
) (activities []ingestion.Activity, importErrors []ingestion.ImportError, err error)
```

Ingests a single utility bill file.

**Parameters:**
- `ctx`: Context for cancellation
- `filename`: File name (used for format detection)
- `content`: File content reader

**Returns:**
- `activities`: Successfully parsed activities
- `importErrors`: Non-fatal row-level errors
- `err`: Fatal error (file corrupt, etc.)

#### IngestFiles

```go
func (a *Adapter) IngestFiles(
    ctx context.Context,
    files map[string]io.Reader,
) (*BatchResult, error)
```

Ingests multiple files concurrently.

**Parameters:**
- `ctx`: Context for cancellation
- `files`: Map of filename to reader

**Returns:**
- `BatchResult`: Aggregated results
- `error`: Fatal error

#### IngestMultipartFile

```go
func (a *Adapter) IngestMultipartFile(
    ctx context.Context,
    fileHeader *multipart.FileHeader,
) (activities []ingestion.Activity, importErrors []ingestion.ImportError, err error)
```

Convenience method for HTTP multipart uploads.

### Helper Methods

#### SetStore

```go
func (a *Adapter) SetStore(store ingestion.ActivityStore)
```

Updates the activity store after adapter creation.

## Data Model

### Activity

```go
type Activity struct {
    ID          string            // Unique identifier
    Source      string            // "utility_bill"
    Category    string            // "electricity", "natural_gas", etc.
    MeterID     string            // Meter/account identifier
    Location    string            // Region code
    PeriodStart time.Time         // Billing period start
    PeriodEnd   time.Time         // Billing period end
    Quantity    float64           // Consumption amount
    Unit        string            // "kWh", "therm", etc.
    OrgID       string            // Organization ID
    Metadata    map[string]string // Additional data
    DataQuality string            // Quality indicator
    CreatedAt   time.Time         // Import timestamp
}
```

### BatchResult

```go
type BatchResult struct {
    FileResults     []*FileResult  // Per-file results
    TotalFiles      int            // Total files processed
    SuccessFiles    int            // Successfully processed
    FailedFiles     int            // Failed processing
    TotalActivities int            // Total activities created
    TotalErrors     int            // Total import errors
    StartedAt       time.Time      // Batch start time
    CompletedAt     time.Time      // Batch completion time
}

// Methods
func (b *BatchResult) HasErrors() bool
func (b *BatchResult) SuccessRate() float64
func (b *BatchResult) Duration() time.Duration
func (b *BatchResult) Summary() string
```

## Testing

### Run Tests

```bash
go test ./internal/ingestion/sources/utility_bills/... -v
```

### Test Coverage

```bash
go test ./internal/ingestion/sources/utility_bills/... -cover
```

Current coverage: **87%**

### Test Categories

1. **Adapter Tests**: Configuration, initialization
2. **Ingestion Tests**: CSV, JSON, errors
3. **Validation Tests**: Required fields, types
4. **Deduplication Tests**: Enabled, disabled
5. **Enrichment Tests**: Mappings, standardization
6. **Batch Tests**: Concurrent processing
7. **Result Tests**: Helper methods

## Examples

### Single File Upload

```go
file, _ := os.Open("january-2025.csv")
defer file.Close()

activities, errors, err := adapter.IngestFile(
    context.Background(),
    "january-2025.csv",
    file,
)

fmt.Printf("Imported %d activities with %d errors\n",
    len(activities), len(errors))
```

### Batch Processing

```go
files := map[string]io.Reader{
    "jan.csv":  reader1,
    "feb.csv":  reader2,
    "mar.csv":  reader3,
}

result, _ := adapter.IngestFiles(context.Background(), files)

fmt.Printf("Processed %d files in %v\n",
    result.TotalFiles, result.Duration())
fmt.Printf("Success rate: %.1f%%\n",
    result.SuccessRate() * 100)
```

### With Enrichment

```go
config := utility_bills.DefaultConfig("org")
config.UtilityProviderMappings = map[string]string{
    "PG&E": "US-CA-PGAE",
}

adapter := utility_bills.NewAdapter(config)

// Bills with provider "PG&E" will have location set to "US-CA-PGAE"
activities, _, _ := adapter.IngestFile(ctx, "bills.csv", file)
```

### Error Handling

```go
activities, errors, err := adapter.IngestFile(ctx, filename, file)

// Fatal error
if err != nil {
    log.Fatalf("Failed to ingest file: %v", err)
}

// Row-level errors
if len(errors) > 0 {
    log.Printf("Warning: %d rows had errors:", len(errors))
    for _, e := range errors {
        log.Printf("  Row %d: %s", e.Row, e.Message)
    }
}

// Process successful activities
for _, activity := range activities {
    processActivity(activity)
}
```

## Best Practices

### 1. Configure Provider Mappings

Map utility provider names to specific grid regions for accurate emissions calculations:

```go
config.UtilityProviderMappings = map[string]string{
    "Pacific Gas & Electric": "US-CA-PGAE",
    "Southern California Edison": "US-CA-SCE",
    "Con Edison": "US-NY-CONED",
}
```

### 2. Enable Deduplication

Prevent re-processing of bills:

```go
config.EnableDeduplication = true
config.DeduplicationWindow = 90 * 24 * time.Hour
```

### 3. Use Batch Processing

For multiple files, use batch processing for better performance:

```go
// Good
result, _ := adapter.IngestFiles(ctx, files)

// Less efficient
for name, reader := range files {
    adapter.IngestFile(ctx, name, reader)
}
```

### 4. Handle Partial Failures

Design for resilience:

```go
result, _ := adapter.IngestFiles(ctx, files)

for _, fr := range result.FileResults {
    if fr.Error != nil {
        // Retry or alert
        retryFile(fr.Filename)
    }
}
```

### 5. Monitor Data Quality

Track and improve data quality:

```go
for _, activity := range activities {
    if activity.DataQuality != "measured" {
        log.Printf("Low quality data: %s", activity.ID)
    }
}
```

## Troubleshooting

### "missing required column"

Ensure CSV has columns: meter_id, period_start, period_end, quantity

### "invalid date format"

Use supported formats: ISO 8601, US (MM/DD/YYYY), EU (DD/MM/YYYY)

### "file too large"

Increase MaxFileSize or split the file

### "validation failed in strict mode"

Set StrictValidation to false or fix data issues

### Duplicate bills

Enable deduplication: `config.EnableDeduplication = true`

## Performance Tuning

### Increase Concurrency

For machines with many cores:

```go
config.MaxConcurrentParsing = runtime.NumCPU()
```

### Adjust File Size Limit

For large files:

```go
config.MaxFileSize = 100 * 1024 * 1024 // 100MB
```

### Disable Features

For maximum performance:

```go
config.EnableDeduplication = false
config.AutoEnrichLocation = false
```

## Integration

### HTTP Handler

See `internal/api/http/handlers/utility_bills_handler.go`

### Router

See `internal/api/http/router.go`

### Worker

Add to ingestion service adapters for scheduled processing

## Support

- Documentation: `/docs/ingestion/utility-bills-guide.md`
- Examples: `/docs/ingestion/examples/`
- Issues: GitHub Issues
- Tests: `utility_bills_test.go`

## License

Copyright (c) OffGridFlow
