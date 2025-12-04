# SAP ERP Ingestion Adapter

This adapter enables OffGridFlow to ingest energy consumption and emissions data from SAP ERP and SAP S/4HANA systems, including the SAP Sustainability module.

## Features

- **OAuth2 Authentication**: Secure authentication with SAP APIs using client credentials flow
- **Energy Consumption Data**: Fetch electricity, gas, and other energy data from SAP ERP
- **Emissions Data**: Retrieve Scope 1, 2, and 3 emissions from SAP Sustainability module
- **Plant/Facility Filtering**: Filter data by specific SAP plants or retrieve organization-wide data
- **Automatic Token Management**: Handles OAuth token refresh automatically
- **Comprehensive Metadata**: Captures SAP-specific fields (plant, meter, cost center, etc.)
- **Data Quality Tracking**: Marks SAP data as "measured" for high-quality reporting

## Configuration

### Environment Variables

Set the following environment variables to enable SAP integration:

```bash
# Enable SAP integration
export OFFGRIDFLOW_SAP_INGEST_ENABLED=true

# SAP API endpoint
export OFFGRIDFLOW_SAP_BASE_URL=https://api.sap.yourcompany.com

# OAuth2 credentials
export OFFGRIDFLOW_SAP_CLIENT_ID=your-client-id
export OFFGRIDFLOW_SAP_CLIENT_SECRET=your-client-secret

# SAP company code
export OFFGRIDFLOW_SAP_COMPANY=1000

# Optional: Filter by specific plant/facility
export OFFGRIDFLOW_SAP_PLANT=US-TX-001

# Organization ID in OffGridFlow
export OFFGRIDFLOW_SAP_ORG_ID=org-123
```

### SAP API Requirements

The adapter expects the following SAP API endpoints to be available:

1. **OAuth Token Endpoint**: `POST /oauth/token`
   - Used for authentication
   - Requires client credentials grant type

2. **Energy Consumption Endpoint**: `GET /api/energy/consumption`
   - Query parameters: `company`, `from`, `to`, `plant` (optional)
   - Returns energy consumption records

3. **Emissions Endpoint**: `GET /api/sustainability/emissions`
   - Query parameters: `company`, `from`, `to`, `plant` (optional)
   - Returns emissions records from SAP Sustainability module

## Data Mapping

### Energy Data

SAP energy records are mapped to OffGridFlow activities as follows:

| SAP Field | OffGridFlow Field | Notes |
|-----------|------------------|-------|
| RecordID | ExternalID | Unique identifier from SAP |
| Plant | Location | Mapped to geographic region |
| Meter | MeterID | Physical meter identifier |
| Date | PeriodStart/End | Daily granularity |
| EnergyType | Category | electricity, natural_gas, diesel, etc. |
| Quantity | Quantity | Numeric consumption value |
| Unit | Unit | kWh, MWh, GJ, m3, etc. |
| CostCenter | Metadata | Stored in metadata |

### Emissions Data

SAP emissions records from the Sustainability module:

| SAP Field | OffGridFlow Field | Notes |
|-----------|------------------|-------|
| RecordID | ExternalID | Unique identifier |
| Date | PeriodStart/End | Daily granularity |
| Source | Metadata | Emission source description |
| EmissionType | Category | CO2, CH4, N2O, etc. |
| KgCO2e | Quantity | CO2 equivalent in kg |
| Scope | Metadata | Scope1, Scope2, or Scope3 |
| Plant | Location | Geographic mapping |

## Energy Type Mapping

The adapter automatically maps SAP energy types to standardized categories:

- **Electricity** → `electricity`
- **Natural Gas** → `natural_gas`
- **Diesel** → `diesel`
- **Fuel Oil** / **Heating Oil** → `fuel_oil`
- **Steam** → `steam`
- **Water** → `water`
- Other types → `energy_<type>`

## Unit Conversion

SAP units are standardized to OffGridFlow units:

| SAP Unit | OffGridFlow Unit |
|----------|------------------|
| kWh, kw | kWh |
| MWh, mw | MWh |
| GJ | GJ |
| L, liter, litre | L |
| m3, m³ | m3 |
| kg | kg |
| tonne, ton, mt | tonne |

## Plant Location Mapping

SAP plant codes are mapped to geographic regions. The adapter includes intelligent mapping:

- **Hyphen-separated codes**: `US-TX-001` → `US-TX`
- **Country prefixes**: 
  - `US*` → `US-UNKNOWN`
  - `EU*`, `DE*` → `EU-CENTRAL`
  - `CN*` → `ASIA-CHINA`
- **Unknown plants** → `GLOBAL`

**Note**: In production environments, customize the `mapPlantToLocation` function with your organization's specific plant-to-region mappings.

## Usage Example

```go
package main

import (
    "context"
    "time"
    
    "github.com/example/offgridflow/internal/ingestion/sources/sap"
)

func main() {
    // Configure the adapter
    cfg := sap.Config{
        BaseURL:      "https://api.sap.company.com",
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        Company:      "1000",
        Plant:        "US-TX-001", // Optional
        OrgID:        "org-123",
        StartDate:    time.Now().AddDate(0, -1, 0), // Last month
        EndDate:      time.Now(),
    }
    
    // Create adapter
    adapter, err := sap.NewAdapter(cfg)
    if err != nil {
        panic(err)
    }
    
    // Ingest data
    activities, err := adapter.Ingest(context.Background())
    if err != nil {
        panic(err)
    }
    
    // Process activities
    for _, activity := range activities {
        fmt.Printf("Activity: %s - %s: %.2f %s\n",
            activity.Source,
            activity.Category,
            activity.Quantity,
            activity.Unit)
    }
}
```

## Testing

Run the comprehensive test suite:

```bash
go test -v ./internal/ingestion/sources/sap
```

The tests include:
- Configuration validation
- Authentication flow
- Energy data fetching
- Emissions data fetching
- Data mapping and conversion
- Unit and location mapping
- End-to-end ingestion

## SAP API Setup Guide

### 1. Create OAuth2 Client

In SAP, create an OAuth2 client with the following scopes:
- `SAP_BASIS_ENERGY_READ` - Read energy consumption data
- `SAP_SUSTAINABILITY_EMISSIONS_READ` - Read emissions data

### 2. Configure API Access

Ensure the following OData services are exposed:
- Energy Consumption Service
- Sustainability Emissions Service

### 3. Network Configuration

- Add OffGridFlow server IP to SAP allowlist
- Configure reverse proxy if needed
- Set up SSL/TLS certificates

### 4. Data Permissions

Grant the OAuth client read access to:
- Energy master data
- Meter reading data
- Sustainability module data
- Plant and cost center data

## Troubleshooting

### Authentication Failures

**Problem**: `authentication failed` errors

**Solutions**:
- Verify client ID and secret are correct
- Check OAuth endpoint URL
- Ensure client has necessary permissions
- Check network connectivity to SAP

### Missing Data

**Problem**: Adapter returns no activities

**Solutions**:
- Verify date range includes data
- Check plant code filter (if specified)
- Confirm company code is correct
- Review SAP API logs for errors
- Ensure SAP Sustainability module is installed (for emissions)

### Data Quality Issues

**Problem**: Incorrect quantities or units

**Solutions**:
- Review unit mapping in adapter code
- Check SAP master data configuration
- Verify conversion factors in SAP
- Customize mapping functions if needed

## Performance Considerations

- **Batch Size**: The adapter fetches all data for the date range in one request
- **Rate Limiting**: Implements automatic retry with exponential backoff
- **Token Caching**: Reuses OAuth tokens until expiration (with 5-minute buffer)
- **Timeout**: Default 30-second timeout per request

For large data volumes, consider:
- Reducing date range
- Using plant-level filtering
- Implementing pagination (requires SAP API support)

## Security Best Practices

1. **Never commit credentials** - Use environment variables
2. **Rotate secrets regularly** - Update client secrets quarterly
3. **Use least privilege** - Grant only required SAP permissions
4. **Enable audit logging** - Track all API access in SAP
5. **Secure transmission** - Always use HTTPS/TLS

## Integration with OffGridFlow

The SAP adapter integrates seamlessly with OffGridFlow's ingestion pipeline:

1. **Automatic Discovery**: Enabled via `OFFGRIDFLOW_SAP_INGEST_ENABLED`
2. **Scheduled Ingestion**: Runs on configured schedule
3. **Activity Storage**: Data stored in PostgreSQL or in-memory store
4. **Emissions Calculation**: Activities flow to emissions calculators
5. **Reporting**: Data available in dashboard and reports

## Support

For issues or questions:
- Check SAP API documentation
- Review OffGridFlow ingestion logs
- Contact SAP administrator for API access issues
- File issues in OffGridFlow repository

## License

Same as OffGridFlow main project.
