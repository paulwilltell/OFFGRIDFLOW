# SEC Climate Disclosure Compliance Module

Complete implementation of SEC Climate-Related Disclosure Rules (17 CFR Parts 210, 229, and 249) for 10-K reporting.

## Overview

This module provides comprehensive support for SEC climate-related disclosure requirements, including:

- **Item 1500**: Climate-Related Governance
- **Item 1501**: Climate-Related Risk Management
- **Item 1502**: Strategy, Business Model, and Outlook
- **Item 1504**: Greenhouse Gas Emissions Metrics
- **Regulation S-X Article 14**: Financial Statement Impact
- **Third-Party Attestation** (for applicable filers)

## Features

### ✅ Complete Implementation

- **Full SEC Rule Compliance**: All disclosure items (1500-1504) fully implemented
- **Filer Type Support**: LAF, AF, SRC, and EGC with appropriate requirements
- **Phase-in Timeline**: Proper handling of FY2025-2028 attestation phase-in
- **Validation Engine**: Comprehensive validation with 50+ validation rules
- **10-K Report Builder**: Formatted reports ready for SEC filing
- **High Test Coverage**: 100+ test cases covering all scenarios

### 📊 Supported Filer Types

| Filer Type | Governance | Risk Mgmt | Strategy | GHG Metrics | Attestation |
|------------|------------|-----------|----------|-------------|-------------|
| **LAF** (Large Accelerated) | Required | Required | Required | Required | FY2025+ (limited), FY2028+ (reasonable) |
| **AF** (Accelerated) | Required | Required | Required | Required | Not required |
| **SRC** (Smaller Reporting) | Optional | Optional | Optional | Optional | Not required |
| **EGC** (Emerging Growth) | Exempt | Exempt | Exempt | Exempt | Not required |

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/example/offgridflow/internal/compliance/sec"
    "github.com/example/offgridflow/internal/models"
)

func main() {
    // Create mapper and validator
    mapper := sec.NewDefaultSECMapper()
    validator := sec.NewValidator()
    
    // Prepare input data
    input := sec.SECInput{
        OrgID:      "org-123",
        OrgName:    "Example Corporation",
        CIK:        "0001234567",
        FiscalYear: 2025,
        FilerType:  "LAF",
        UnifiedData: &models.UnifiedESGData{
            // ... your ESG data
        },
    }
    
    // Build SEC report
    report, err := mapper.BuildReport(context.Background(), input)
    if err != nil {
        panic(err)
    }
    
    // Validate report
    results := validator.ValidateReport(*report)
    if !results.Valid {
        for _, err := range results.Errors {
            fmt.Printf("Error: %s - %s\n", err.Field, err.Message)
        }
    }
    
    fmt.Printf("Compliance Score: %.1f%%\n", report.ComplianceScore)
}
```

### Generate 10-K Formatted Report

```go
// Create report builder
builder := sec.NewReportBuilder()

// Build formatted 10-K report
report10K, err := builder.Build10KReport(context.Background(), input)
if err != nil {
    panic(err)
}

// Access formatted sections
fmt.Println(report10K.Header)
fmt.Println(report10K.Item1500Governance)
fmt.Println(report10K.Item1504GHGMetrics)
```

### Strict Validation Mode

```go
// Enable strict mode for best-practice warnings
validator := sec.NewStrictValidator()

results := validator.ValidateReport(report)
for _, warning := range results.Warnings {
    fmt.Printf("Warning: %s - %s\n", warning.Field, warning.Message)
}
```

## Disclosure Requirements

### Item 1500: Governance

**Required Elements:**
- Board oversight of climate-related risks
- Responsible board committee identification
- Oversight frequency
- Management's role in assessment and management
- Responsible executive identification
- Processes and reporting structure

**Example:**
```go
governance := &sec.GovernanceDisclosure{
    BoardOversight: sec.BoardOversightInfo{
        HasBoardOversight:    true,
        ResponsibleCommittee: "Risk Committee",
        OversightFrequency:   "quarterly",
        DirectorsWithExpertise: []string{"John Doe", "Jane Smith"},
    },
    ManagementRole: sec.ManagementRoleInfo{
        ResponsibleExecutive:  "Chief Sustainability Officer",
        ProcessesAndFrequency: "Monthly climate risk reviews",
        ReportingStructure:    "Reports to CEO",
    },
}
```

### Item 1501: Risk Management

**Required Elements:**
- Processes for identifying climate-related risks
- Processes for managing climate-related risks  
- Integration with enterprise risk management
- Material climate-related risks disclosure

**Example:**
```go
riskMgmt := &sec.RiskManagementDisclosure{
    RiskIdentification: sec.RiskIdentificationProcess{
        ProcessDescription: "Annual climate risk assessment",
        RiskCategories:     []string{"physical", "transition"},
        TimeHorizons:       []string{"short-term", "medium-term", "long-term"},
    },
    MaterialRisks: []sec.MaterialClimateRisk{
        {
            RiskType:       "physical",
            Description:    "Flooding risk at coastal facilities",
            TimeHorizon:    "long-term",
            MitigationPlan: "Facility relocation and flood defenses",
        },
    },
    ERMIntegration: sec.ERMIntegrationInfo{
        IsIntegrated:        true,
        ERMFramework:        "COSO ERM",
        IntegrationApproach: "Quarterly risk assessments",
    },
}
```

### Item 1502: Strategy

**Required Elements:**
- Material climate-related impacts
- Climate transition plan (if applicable)
- Scenario analysis (if conducted)
- Internal carbon price (if used)
- Climate-related targets and goals

**Example:**
```go
strategy := &sec.StrategyDisclosure{
    MaterialImpacts: []sec.StrategyImpact{
        {
            ImpactArea:      "operations",
            Description:     "Supply chain disruption from extreme weather",
            TimeHorizon:     "medium-term",
            ResponseActions: "Supplier diversification",
        },
    },
    ClimateTargets: []sec.ClimateTarget{
        {
            Description:   "Reduce Scope 1+2 emissions by 50%",
            TargetYear:    2030,
            BaseYear:      2020,
            BaselineValue: 100000,
            TargetValue:   50000,
            Unit:          "tCO2e",
            Scope:         "Scope 1+2",
            SBTiAligned:   true,
        },
    },
}
```

### Item 1504: GHG Emissions Metrics

**Required for LAF and AF:**
- Scope 1 emissions (direct)
- Scope 2 emissions (indirect) - both location-based and market-based
- Methodology and assumptions
- Data quality and verification

**Optional but recommended:**
- Scope 3 emissions
- Emissions intensity metric
- Historical trends

**Example:**
```go
ghgMetrics := &sec.GHGMetricsDisclosure{
    Scope1Emissions: &sec.ScopeEmissions{
        TotalEmissions:         10000.0,
        ReportingYear:          2025,
        OrganizationalBoundary: "operational control",
    },
    Scope2Emissions: &sec.ScopeEmissions{
        TotalEmissions:         5000.0,
        ReportingYear:          2025,
        OrganizationalBoundary: "operational control",
        LocationBased:          5000.0,
        MarketBased:            4500.0,
    },
    Methodology: sec.MethodologyDisclosure{
        Standard:              "GHG Protocol Corporate Standard",
        ConsolidationApproach: "operational control",
        GWPSource:             "IPCC AR6",
        BaseYear:              2020,
    },
    DataQuality: sec.DataQualityInfo{
        VerificationStatus:   "limited_assurance",
        VerificationProvider: "Independent Auditor",
        DataCoverage:         95.0,
    },
}
```

### Third-Party Attestation

**Required for Large Accelerated Filers:**
- FY2025-2027: Limited assurance on Scopes 1 & 2
- FY2028+: Reasonable assurance on Scopes 1 & 2

**Standards:**
- AT-C Section 210 (AICPA)
- ISAE 3410 (International)

**Example:**
```go
attestation := &sec.AttestationReport{
    Required:         true,
    AssuranceLevel:   "limited", // or "reasonable" for FY2028+
    Provider:         "Big Four Accounting Firm",
    Standard:         "AT-C 210",
    OpinionType:      "unmodified",
    OpinionStatement: "The GHG emissions are fairly stated...",
    ScopesCovered:    []string{"Scope 1", "Scope 2"},
    ReportDate:       time.Now(),
}
```

## Validation

The validator provides three levels of feedback:

### 1. Errors (Blocking)

Critical compliance issues that must be fixed:
- Missing required disclosures
- Invalid data formats
- Regulatory requirement violations

### 2. Warnings (Non-blocking)

Best practice recommendations and optional disclosures:
- Missing optional but recommended information
- Potential data quality concerns
- Enhancement suggestions

### 3. Info (Informational)

Additional context and guidance.

## Compliance Scoring

The module calculates an overall compliance score (0-100%) based on:

- **Required Disclosures**: Weighted heavily
- **Optional Disclosures**: Moderate weight
- **Data Quality**: Quality and completeness of provided data
- **Best Practices**: Alignment with leading practices

```go
// Compliance score calculation
score := report.ComplianceScore // 0-100%

// Breakdown by section
for _, disclosure := range report.RequiredDisclosures {
    fmt.Printf("%s: Required=%v, Complete=%v\n", 
        disclosure.Item, disclosure.Required, disclosure.Complete)
}
```

## Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./internal/compliance/sec/...

# Run with coverage
go test ./internal/compliance/sec/... -cover

# Run specific test
go test ./internal/compliance/sec/... -run TestValidator_ValidateReport
```

## Validation Rules

### CIK Validation
- Must be exactly 10 digits
- Numeric characters only
- Leading zeros allowed

### Filer Type Validation
- Must be one of: LAF, AF, SRC, EGC
- Case-sensitive

### GHG Metrics Validation
- Scope 1 required for LAF/AF
- Scope 2 required for LAF/AF with both location-based and market-based
- Methodology must include consolidation approach
- Data coverage >= 70% recommended

### Financial Impact Validation
- Impact amounts must be reasonable
- Percentages must be 0-100%
- All impacted line items must be disclosed if threshold met

## Integration Examples

### With Data Sources

```go
// Map from Sustainability Accounting
import "github.com/example/offgridflow/internal/accounting"

accounting := accounting.NewSustainabilityAccounting()
emissions := accounting.CalculateScope1Emissions(ctx, orgID, year)

// Map to SEC format
ghgMetrics := &sec.GHGMetricsDisclosure{
    Scope1Emissions: &sec.ScopeEmissions{
        TotalEmissions:         emissions.Total,
        ReportingYear:          year,
        OrganizationalBoundary: "operational control",
    },
}
```

### Export to Different Formats

```go
// JSON export
jsonData, err := json.MarshalIndent(report, "", "  ")

// XML export (XBRL)
// Custom serialization to XBRL format

// PDF generation
// Use report builder's formatted output
```

## Architecture

```
sec/
├── mapper.go              # SEC report builder
├── mapper_test.go         # Mapper tests
├── validator.go           # Validation engine
├── validator_test.go      # Validator tests  
├── report_builder.go      # 10-K format builder
├── types.go               # Data structures
└── README.md             # This file
```

## Compliance Timeline

| Fiscal Year | LAF Requirements | AF Requirements | SRC/EGC |
|-------------|------------------|-----------------|---------|
| **2024** | Voluntary | Voluntary | Exempt |
| **2025** | All + Limited Assurance | All (no attestation) | Exempt |
| **2026** | All + Limited Assurance | All + Limited Assurance | Exempt |
| **2027** | All + Limited Assurance | All + Limited Assurance | Exempt |
| **2028+** | All + Reasonable Assurance | All + Limited Assurance | Exempt |

## References

- [SEC Final Rule](https://www.sec.gov/files/rules/final/2024/33-11275.pdf)
- [GHG Protocol Corporate Standard](https://ghgprotocol.org/corporate-standard)
- [TCFD Recommendations](https://www.fsb-tcfd.org/)
- [AICPA AT-C Section 210](https://www.aicpa.org/research/standards/auditattest/downloadabledocuments/at-c-00210.pdf)

## Support

For issues or questions:
1. Check the test files for usage examples
2. Review the validation error messages
3. Consult the SEC final rule documentation

## License

Part of the OffGridFlow ESG platform.
