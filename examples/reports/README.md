# Example Compliance Reports

This directory contains sample PDF reports generated from OffGridFlow test datasets.

## Available Reports

### 1. CSRD Report - Manufacturing Company
**File**: `csrd-manufacturing-2024.pdf`  
**Standard**: EU Corporate Sustainability Reporting Directive  
**Organization**: Precision Manufacturing Corp  
**Sector**: Manufacturing (Industrial)  
**Reporting Year**: 2024

**Emissions Summary**:
- Scope 1: 550.53 tCO2e (Direct emissions)
- Scope 2: 1,779.44 tCO2e (Electricity)
- Scope 3: 5,173.81 tCO2e (Value chain)
- **Total**: 7,503.78 tCO2e

**Key Features**:
- 15 tracked activities
- 2 locations (Detroit, Cleveland)
- 450 employees
- Data quality: 94.2%, Completeness: 96.7%

**Report Sections**:
1. Title Page
2. Executive Summary
3. Emissions Overview
4. Scope 1/2/3 Details
5. Data Quality Statement
6. Calculation Methodology
7. Assurance Statement

---

### 2. SEC Climate Disclosure - Tech Company
**File**: `sec-climate-disclosure-2024.pdf`  
**Standard**: US Securities & Exchange Commission Climate Disclosure  
**Organization**: CloudTech Solutions Inc  
**Sector**: Technology (SaaS/Cloud)  
**Reporting Year**: 2024

**Emissions Summary**:
- Scope 1: 27.79 tCO2e (Minimal - electric fleet)
- Scope 2: 4,839.88 tCO2e (Cloud infrastructure)
- Scope 3: 21,713.39 tCO2e (Customer usage dominant)
- **Total**: 26,581.06 tCO2e

**Key Features**:
- 20 tracked activities
- 1,250 employees
- Heavy cloud infrastructure footprint
- Customer usage = 80% of Scope 3
- Data quality: 91.5%, Completeness: 100%

**Report Sections**:
1. Cover Page
2. Governance (Board oversight)
3. Strategy (Climate risks)
4. Risk Management
5. Metrics and Targets
6. GHG Emissions Data
7. Scope Breakdowns
8. Data Quality & Assurance

---

### 3. California Climate Corporate Data Accountability Act
**File**: `california-ccdaa-2024.pdf`  
**Standard**: California SB 253 (State Climate Disclosure)  
**Organization**: Global Retail Group  
**Sector**: Retail (Multi-channel)  
**Reporting Year**: 2024

**Emissions Summary**:
- Scope 1: 4,206.69 tCO2e (Delivery fleet + facilities)
- Scope 2: 24,563.00 tCO2e (250 stores + 8 DCs)
- Scope 3: 52,065.48 tCO2e (Supply chain dominant)
- **Total**: 80,835.17 tCO2e

**Key Features**:
- 24 tracked activities
- 12,500 employees
- 250 retail locations
- 8 distribution centers
- Annual revenue: $1.5B+ (SB 253 threshold met)
- Data quality: 89.3%, Completeness: 95.8%

**Report Sections**:
1. Cover Page (California branding)
2. Executive Summary
3. Scope 1 & 2 Emissions
4. Scope 3 Overview
5. All 15 Scope 3 Categories
6. Methodology & Data Quality
7. Third-Party Assurance
8. Appendix

---

### 4. CBAM Report - Carbon Border Adjustment
**File**: `cbam-report-2024.pdf`  
**Standard**: EU Carbon Border Adjustment Mechanism  
**Organization**: Precision Manufacturing Corp (Importer)  
**Sector**: Manufacturing (with imports)  
**Reporting Year**: 2024

**Emissions Summary**:
- Direct Process Emissions: 550.53 tCO2e
- Indirect (Electricity): 1,779.44 tCO2e
- **Total Embedded**: 2,329.97 tCO2e

**Key Features**:
- Import scenario: Steel from China
- CN Code: 7208 10 00 (Hot-rolled steel)
- 1,000 tonnes imported
- Embedded emissions: 2.5 tCO2e/tonne
- CBAM carbon price: €85/tonne
- **Total CBAM liability**: €212,500

**Report Sections**:
1. Cover Page (EU branding)
2. Executive Summary
3. Imported Goods
4. Embedded Emissions Calculation
5. Production Sites
6. CBAM Pricing & Obligations
7. Verification
8. Methodology

---

### 5. IFRS S2 Climate-Related Disclosures
**File**: `ifrs-s2-tech-company-2024.pdf`  
**Standard**: IFRS Sustainability Disclosure (Global)  
**Organization**: CloudTech Solutions Inc  
**Sector**: Technology  
**Reporting Year**: 2024

**Emissions Summary**:
- Scope 1 + 2: 4,867.67 tCO2e
- Scope 3: 21,713.39 tCO2e
- **Total**: 26,581.06 tCO2e

**Key Features**:
- TCFD-aligned four-pillar structure
- Climate scenario analysis (1.5°C, 2°C pathways)
- Financial impact quantification
- Transition plan with milestones
- Governance disclosure

**Report Sections**:
1. Cover Page
2. Governance
3. Strategy: Climate Risks
4. Strategy: Opportunities
5. Financial Impacts
6. Scenario Analysis
7. Transition Plan
8. Risk Management
9. Metrics and Targets
10. GHG Emissions

---

## Report Characteristics

### Diversity Across Industries

| Report | Sector | Scope 1 | Scope 2 | Scope 3 | Dominant Source |
|--------|--------|---------|---------|---------|-----------------|
| CSRD | Manufacturing | 7% | 24% | **69%** | Aluminum procurement |
| SEC | Tech/SaaS | 0.1% | 18% | **82%** | Customer cloud usage |
| California | Retail | 5% | 30% | **64%** | Supply chain |
| CBAM | Import | 24% | 76% | N/A | Production energy |
| IFRS | Tech | 0.1% | 18% | **82%** | Value chain |

### Regulatory Coverage

- ✅ **EU**: CSRD, CBAM
- ✅ **US Federal**: SEC Climate Disclosure
- ✅ **US State**: California SB 253
- ✅ **Global**: IFRS S2 (ISSB)

### Use Cases

**For Sales/Marketing**:
- Show real output quality
- Demonstrate multi-jurisdiction support
- Highlight professional formatting

**For Compliance Review**:
- Send to external auditors
- Validate report structure
- Test regulatory alignment

**For Customer Onboarding**:
- Provide sample outputs
- Set expectations
- Show report customization

**For Developer Testing**:
- Integration test references
- PDF generation validation
- Data structure examples

---

## Generating Additional Reports

To regenerate or create new example reports:

```powershell
# Using PowerShell
cd C:\Users\pault\OffGridFlow
.\scripts\generate-reports.ps1
```

Or directly with Go:

```powershell
go run scripts/generate-example-reports.go
```

---

## Data Sources

All reports are generated from test datasets in `testdata/`:

1. `manufacturing_company_2024.json` → CSRD + CBAM reports
2. `tech_company_2024.json` → SEC + IFRS reports  
3. `retail_company_2024.json` → California report

These datasets contain realistic emissions data across:
- 59 total activities
- 3 industry sectors
- All Scope 1/2/3 categories
- Multiple geographic locations
- Varied data quality scenarios

---

## Technical Details

**PDF Generation**: 
- Library: `github.com/jung-kurt/gofpdf`
- Format: A4 size, portrait orientation
- Font: Arial (standard, bold, italic variants)
- Colors: Report-specific branding (EU blue, SEC blue-gray, California gold, etc.)

**Report Hashing**:
- Algorithm: SHA-256
- Purpose: Integrity verification
- Included in each report footer

**Data Quality Tracking**:
- Completeness percentage
- Quality score (0-100)
- Missing field identification
- Warning generation

---

## File Sizes

Typical report sizes:
- CSRD: ~50-80 KB (9 sections)
- SEC: ~70-100 KB (10 sections)
- California: ~80-110 KB (8 sections + Scope 3 detail)
- CBAM: ~60-90 KB (8 sections)
- IFRS S2: ~90-120 KB (10 sections)

---

## Compliance Notes

These are **example reports** for demonstration purposes. 

For production use:
1. Replace sample data with actual organizational data
2. Obtain third-party verification where required
3. Review with legal/compliance team
4. Submit through appropriate regulatory channels
5. Maintain audit trail of all submissions

---

**Last Updated**: December 5, 2025  
**Generated By**: OffGridFlow Compliance Engine v1.0.0  
**Status**: Example/Demo Reports
