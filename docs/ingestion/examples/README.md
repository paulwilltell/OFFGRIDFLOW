# Utility Bills Ingestion Examples

This directory contains example files for utility bill ingestion.

## Files

### utility-bill-example.csv
Standard CSV format with multiple meters and utility types (electricity and natural gas).

**Columns:**
- `meter_id`: Unique meter identifier
- `location`: Geographic region code
- `period_start`: Billing period start date
- `period_end`: Billing period end date
- `kwh`: Consumption quantity (or therms for gas)
- `category`: Utility type (electricity, natural_gas, etc.)
- `provider`: Utility company name
- `account`: Account number
- `invoice`: Invoice number

### utility-bill-example.json
JSON format with rich metadata, suitable for API integrations.

**Structure:**
- `bills`: Array of bill objects
- Each bill contains required fields plus optional metadata
- Supports nested metadata objects for additional context

## Usage

### Upload via HTTP API

**Single file:**
```bash
curl -X POST http://localhost:8080/api/ingestion/utility-bills/upload \
  -F "file=@utility-bill-example.csv" \
  -F "org_id=my-company" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Batch upload:**
```bash
curl -X POST http://localhost:8080/api/ingestion/utility-bills/batch-upload \
  -F "files=@utility-bill-example.csv" \
  -F "files=@utility-bill-example.json" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Programmatic Usage

```go
package main

import (
    "context"
    "os"
    "github.com/example/offgridflow/internal/ingestion/sources/utility_bills"
)

func main() {
    config := utility_bills.DefaultConfig("my-org")
    adapter := utility_bills.NewAdapter(config)

    file, _ := os.Open("utility-bill-example.csv")
    defer file.Close()

    activities, errors, err := adapter.IngestFile(
        context.Background(),
        "utility-bill-example.csv",
        file,
    )

    // Process results...
}
```

## Data Quality Notes

- All dates are in ISO 8601 format (YYYY-MM-DD)
- Quantities are positive numbers
- Location codes should match your emission factor registry
- Provider names can be mapped to specific locations via configuration

## Customization

You can customize these examples for your specific utility providers:

1. Update location codes to match your regions
2. Add provider-specific metadata fields
3. Adjust column names to match your bill format
4. Include additional categories (water, steam, etc.)
