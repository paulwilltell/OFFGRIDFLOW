package sec

import (
	"testing"
	"time"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	if v == nil {
		t.Fatal("NewValidator returned nil")
	}
	if v.StrictMode {
		t.Error("Default validator should not be in strict mode")
	}
}

func TestNewStrictValidator(t *testing.T) {
	v := NewStrictValidator()
	if v == nil {
		t.Fatal("NewStrictValidator returned nil")
	}
	if !v.StrictMode {
		t.Error("Strict validator should be in strict mode")
	}
}

func TestValidator_ValidateReport_Complete(t *testing.T) {
	v := NewValidator()
	report := createValidLAFReport()

	results := v.ValidateReport(report)
	
	if !results.Valid {
		t.Errorf("Expected valid report, but got %d errors", len(results.Errors))
		for _, err := range results.Errors {
			t.Logf("Error: %s - %s", err.Field, err.Message)
		}
	}
}

func TestValidator_ValidateReport_MissingCIK(t *testing.T) {
	v := NewValidator()
	report := createValidLAFReport()
	report.CIK = ""

	results := v.ValidateReport(report)

	if results.Valid {
		t.Error("Expected invalid report due to missing CIK")
	}

	foundError := false
	for _, err := range results.Errors {
		if err.Field == "cik" && err.Code == "REQUIRED_FIELD" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected CIK required field error")
	}
}

func TestValidator_ValidateReport_InvalidCIK(t *testing.T) {
	v := NewValidator()
	report := createValidLAFReport()
	report.CIK = "123" // Too short

	results := v.ValidateReport(report)

	if results.Valid {
		t.Error("Expected invalid report due to invalid CIK format")
	}

	foundError := false
	for _, err := range results.Errors {
		if err.Field == "cik" && err.Code == "INVALID_FORMAT" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected CIK invalid format error")
	}
}

func TestValidator_ValidateReport_InvalidFilerType(t *testing.T) {
	v := NewValidator()
	report := createValidLAFReport()
	report.FilerType = "INVALID"

	results := v.ValidateReport(report)

	if results.Valid {
		t.Error("Expected invalid report due to invalid filer type")
	}

	foundError := false
	for _, err := range results.Errors {
		if err.Field == "filerType" && err.Code == "INVALID_VALUE" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Error("Expected filer type invalid value error")
	}
}

func TestValidator_ValidateGovernance(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name          string
		governance    *GovernanceDisclosure
		expectErrors  int
		expectWarnings int
	}{
		{
			name: "Complete governance",
			governance: &GovernanceDisclosure{
				BoardOversight: BoardOversightInfo{
					HasBoardOversight:    true,
					ResponsibleCommittee: "Risk Committee",
					OversightFrequency:   "quarterly",
				},
				ManagementRole: ManagementRoleInfo{
					ResponsibleExecutive:  "CSO",
					ProcessesAndFrequency: "Monthly reviews",
				},
			},
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name: "Missing committee",
			governance: &GovernanceDisclosure{
				BoardOversight: BoardOversightInfo{
					HasBoardOversight: true,
				},
				ManagementRole: ManagementRoleInfo{
					ResponsibleExecutive: "CSO",
				},
			},
			expectErrors:   1,
			expectWarnings: 2,
		},
		{
			name: "No board oversight",
			governance: &GovernanceDisclosure{
				BoardOversight: BoardOversightInfo{
					HasBoardOversight: false,
				},
				ManagementRole: ManagementRoleInfo{
					ResponsibleExecutive: "CFO",
				},
			},
			expectErrors:   0,
			expectWarnings: 2, // NO_BOARD_OVERSIGHT + missing ProcessesAndFrequency
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := &ValidationResults{
				Errors:   []ValidationError{},
				Warnings: []ValidationWarning{},
			}

			v.validateGovernance(tt.governance, results)

			if len(results.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(results.Errors))
			}

			if len(results.Warnings) != tt.expectWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectWarnings, len(results.Warnings))
				for i, w := range results.Warnings {
					t.Logf("Warning %d: [%s] %s - %s", i+1, w.Code, w.Field, w.Message)
				}
			}
		})
	}
}

func TestValidator_ValidateGHGMetrics_LAF(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name         string
		metrics      *GHGMetricsDisclosure
		filerType    string
		expectErrors int
	}{
		{
			name: "Complete GHG metrics for LAF",
			metrics: &GHGMetricsDisclosure{
				Scope1Emissions: &ScopeEmissions{
					TotalEmissions:         10000.0,
					ReportingYear:          2025,
					OrganizationalBoundary: "operational control",
				},
				Scope2Emissions: &ScopeEmissions{
					TotalEmissions:         5000.0,
					ReportingYear:          2025,
					OrganizationalBoundary: "operational control",
					LocationBased:          5000.0,
					MarketBased:            4500.0,
				},
				Methodology: MethodologyDisclosure{
					Standard:              "GHG Protocol",
					ConsolidationApproach: "operational control",
				},
				DataQuality: DataQualityInfo{
					VerificationStatus: "limited_assurance",
				},
			},
			filerType:    "LAF",
			expectErrors: 0,
		},
		{
			name:         "Missing GHG metrics for LAF",
			metrics:      nil,
			filerType:    "LAF",
			expectErrors: 1, // Missing GHG metrics entirely
		},
		{
			name: "Missing Scope 2",
			metrics: &GHGMetricsDisclosure{
				Scope1Emissions: &ScopeEmissions{
					TotalEmissions:         10000.0,
					ReportingYear:          2025,
					OrganizationalBoundary: "operational control",
				},
				Methodology: MethodologyDisclosure{
					Standard:              "GHG Protocol",
					ConsolidationApproach: "operational control",
				},
				DataQuality: DataQualityInfo{},
			},
			filerType:    "LAF",
			expectErrors: 1, // Missing Scope 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := &ValidationResults{
				Errors:   []ValidationError{},
				Warnings: []ValidationWarning{},
			}

			v.validateGHGMetrics(tt.metrics, tt.filerType, 2025, results)

			if len(results.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(results.Errors))
				for _, err := range results.Errors {
					t.Logf("Error: %s - %s", err.Field, err.Message)
				}
			}
		})
	}
}

func TestValidator_ValidateAttestation_LAF(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name         string
		attestation  *AttestationReport
		fiscalYear   int
		expectErrors int
	}{
		{
			name: "Valid limited assurance FY2025",
			attestation: &AttestationReport{
				AssuranceLevel:   "limited",
				Provider:         "Big Four",
				Standard:         "AT-C 210",
				OpinionType:      "unmodified",
				OpinionStatement: "Opinion statement",
				ScopesCovered:    []string{"Scope 1", "Scope 2"},
			},
			fiscalYear:   2025,
			expectErrors: 0,
		},
		{
			name: "Wrong assurance level FY2028",
			attestation: &AttestationReport{
				AssuranceLevel:   "limited",
				Provider:         "Big Four",
				Standard:         "AT-C 210",
				OpinionType:      "unmodified",
				OpinionStatement: "Opinion",
				ScopesCovered:    []string{"Scope 1", "Scope 2"},
			},
			fiscalYear:   2028,
			expectErrors: 1, // Should be reasonable assurance
		},
		{
			name: "Missing Scope 2",
			attestation: &AttestationReport{
				AssuranceLevel:   "limited",
				Provider:         "Big Four",
				Standard:         "AT-C 210",
				OpinionType:      "unmodified",
				OpinionStatement: "Opinion",
				ScopesCovered:    []string{"Scope 1"},
			},
			fiscalYear:   2025,
			expectErrors: 1,
		},
		{
			name: "Adverse opinion",
			attestation: &AttestationReport{
				AssuranceLevel:   "limited",
				Provider:         "Big Four",
				Standard:         "AT-C 210",
				OpinionType:      "adverse",
				OpinionStatement: "Adverse opinion",
				ScopesCovered:    []string{"Scope 1", "Scope 2"},
			},
			fiscalYear:   2025,
			expectErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := &ValidationResults{
				Errors:   []ValidationError{},
				Warnings: []ValidationWarning{},
			}

			v.validateAttestation(tt.attestation, "LAF", tt.fiscalYear, results)

			if len(results.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(results.Errors))
				for _, err := range results.Errors {
					t.Logf("Error: %s - %s", err.Field, err.Message)
				}
			}
		})
	}
}

func TestValidator_ValidateFinancialImpact(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name         string
		impact       *FinancialStatementImpact
		expectErrors int
	}{
		{
			name: "Valid impact disclosure",
			impact: &FinancialStatementImpact{
				DisclosureThresholdMet: true,
				ImpactedItems: []LineItemImpact{
					{
						LineItem:         "Cost of Goods Sold",
						ImpactAmount:     100000,
						ImpactPercentage: 1.5,
						Description:      "Carbon tax impact",
					},
				},
			},
			expectErrors: 0,
		},
		{
			name: "Threshold met but no items",
			impact: &FinancialStatementImpact{
				DisclosureThresholdMet: true,
				ImpactedItems:          []LineItemImpact{},
			},
			expectErrors: 1,
		},
		{
			name: "Invalid percentage",
			impact: &FinancialStatementImpact{
				DisclosureThresholdMet: true,
				ImpactedItems: []LineItemImpact{
					{
						LineItem:         "Revenue",
						ImpactAmount:     100000,
						ImpactPercentage: 150.0, // Invalid
					},
				},
			},
			expectErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := &ValidationResults{
				Errors:   []ValidationError{},
				Warnings: []ValidationWarning{},
			}

			v.validateFinancialImpact(tt.impact, results)

			if len(results.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(results.Errors))
			}
		})
	}
}

func TestValidator_ValidateMaterialRisk(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name         string
		risk         MaterialClimateRisk
		expectErrors int
	}{
		{
			name: "Valid physical risk",
			risk: MaterialClimateRisk{
				RiskType:    "physical",
				Description: "Flooding risk",
				TimeHorizon: "long-term",
			},
			expectErrors: 0,
		},
		{
			name: "Valid transition risk",
			risk: MaterialClimateRisk{
				RiskType:    "transition",
				Description: "Carbon pricing risk",
				TimeHorizon: "medium-term",
			},
			expectErrors: 0,
		},
		{
			name: "Invalid risk type",
			risk: MaterialClimateRisk{
				RiskType:    "unknown",
				Description: "Some risk",
			},
			expectErrors: 1,
		},
		{
			name: "Missing description",
			risk: MaterialClimateRisk{
				RiskType: "physical",
			},
			expectErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := &ValidationResults{
				Errors:   []ValidationError{},
				Warnings: []ValidationWarning{},
			}

			v.validateMaterialRisk(tt.risk, 0, results)

			if len(results.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(results.Errors))
			}
		})
	}
}

func TestValidator_ValidateClimateTarget(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name         string
		target       ClimateTarget
		expectErrors int
	}{
		{
			name: "Valid target",
			target: ClimateTarget{
				Description: "Reduce Scope 1+2 by 50%",
				TargetYear:  2030,
				Unit:        "tCO2e",
				Scope:       "Scope 1+2",
			},
			expectErrors: 0,
		},
		{
			name: "Missing target year",
			target: ClimateTarget{
				Description: "Net zero",
				Unit:        "tCO2e",
			},
			expectErrors: 1,
		},
		{
			name: "Missing unit",
			target: ClimateTarget{
				Description: "Reduce emissions",
				TargetYear:  2030,
			},
			expectErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := &ValidationResults{
				Errors:   []ValidationError{},
				Warnings: []ValidationWarning{},
			}

			v.validateClimateTarget(tt.target, 0, results)

			if len(results.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(results.Errors))
			}
		})
	}
}

func TestValidator_IsValidCIK(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		cik   string
		valid bool
	}{
		{"0001234567", true},
		{"0000000001", true},
		{"1234567890", true},
		{"123456789", false},  // Too short
		{"12345678901", false}, // Too long
		{"000123456A", false},  // Non-numeric
		{"", false},            // Empty
	}

	for _, tt := range tests {
		t.Run(tt.cik, func(t *testing.T) {
			valid := v.isValidCIK(tt.cik)
			if valid != tt.valid {
				t.Errorf("CIK %s: expected valid=%v, got %v", tt.cik, tt.valid, valid)
			}
		})
	}
}

func TestValidator_IsValidFilerType(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		filerType string
		valid     bool
	}{
		{"LAF", true},
		{"AF", true},
		{"SRC", true},
		{"EGC", true},
		{"INVALID", false},
		{"", false},
		{"laf", false}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.filerType, func(t *testing.T) {
			valid := v.isValidFilerType(tt.filerType)
			if valid != tt.valid {
				t.Errorf("FilerType %s: expected valid=%v, got %v", tt.filerType, tt.valid, valid)
			}
		})
	}
}

// Helper function to create a valid LAF report for testing
func createValidLAFReport() SECReport {
	return SECReport{
		OrgID:      "org-test-001",
		OrgName:    "Test Corporation",
		CIK:        "0001234567",
		FiscalYear: 2025,
		FilingType: "10-K",
		FilerType:  "LAF",
		IsEGC:      false,
		GeneratedAt: time.Now(),
		Governance: &GovernanceDisclosure{
			BoardOversight: BoardOversightInfo{
				HasBoardOversight:    true,
				ResponsibleCommittee: "Risk Committee",
				OversightFrequency:   "quarterly",
			},
			ManagementRole: ManagementRoleInfo{
				ResponsibleExecutive:  "Chief Sustainability Officer",
				ProcessesAndFrequency: "Monthly climate risk reviews",
			},
		},
		RiskManagement: &RiskManagementDisclosure{
			RiskIdentification: RiskIdentificationProcess{
				ProcessDescription: "Annual climate risk assessment",
				RiskCategories:     []string{"physical", "transition"},
				TimeHorizons:       []string{"short-term", "medium-term", "long-term"},
			},
			RiskManagement: RiskManagementProcess{
				ProcessDescription: "Integrated risk management",
			},
			MaterialRisks: []MaterialClimateRisk{
				{
					RiskType:    "physical",
					Description: "Flooding risk at coastal facilities",
					TimeHorizon: "long-term",
				},
			},
			ERMIntegration: ERMIntegrationInfo{
				IsIntegrated: true,
			},
		},
		Strategy: &StrategyDisclosure{
			MaterialImpacts: []StrategyImpact{
				{
					ImpactArea:  "operations",
					Description: "Supply chain disruption",
					TimeHorizon: "medium-term",
				},
			},
		},
		GHGMetrics: &GHGMetricsDisclosure{
			Scope1Emissions: &ScopeEmissions{
				TotalEmissions:         10000.0,
				ReportingYear:          2025,
				OrganizationalBoundary: "operational control",
			},
			Scope2Emissions: &ScopeEmissions{
				TotalEmissions:         5000.0,
				ReportingYear:          2025,
				OrganizationalBoundary: "operational control",
				LocationBased:          5000.0,
				MarketBased:            4500.0,
			},
			Methodology: MethodologyDisclosure{
				Standard:              "GHG Protocol Corporate Standard",
				ConsolidationApproach: "operational control",
				GWPSource:             "IPCC AR6",
			},
			DataQuality: DataQualityInfo{
				VerificationStatus: "limited_assurance",
				DataCoverage:       95.0,
			},
		},
		FinancialImpact: &FinancialStatementImpact{
			DisclosureThresholdMet: false,
		},
		Attestation: &AttestationReport{
			Required:         true,
			AssuranceLevel:   "limited",
			Provider:         "Independent Auditor",
			Standard:         "AT-C 210",
			OpinionType:      "unmodified",
			OpinionStatement: "The GHG emissions are fairly stated",
			ScopesCovered:    []string{"Scope 1", "Scope 2"},
			ReportDate:       time.Now(),
		},
		RequiredDisclosures: []DisclosureStatus{
			{Item: "Item 1500", Required: true, Complete: true},
			{Item: "Item 1501", Required: true, Complete: true},
			{Item: "Item 1502", Required: true, Complete: true},
			{Item: "Item 1504", Required: true, Complete: true},
			{Item: "Attestation", Required: true, Complete: true},
		},
		ComplianceScore: 100.0,
	}
}
