package sec

import (
	"context"
	"testing"
	"time"
)

func TestNewDefaultSECMapper(t *testing.T) {
	mapper := NewDefaultSECMapper()
	if mapper == nil {
		t.Fatal("NewDefaultSECMapper returned nil")
	}
	if mapper.validator == nil {
		t.Error("Mapper should have a validator")
	}
}

func TestDefaultSECMapper_BuildReport_BasicStructure(t *testing.T) {
	mapper := NewDefaultSECMapper()
	ctx := context.Background()

	input := SECInput{
		OrgID:      "org-test-001",
		OrgName:    "Test Corporation",
		FiscalYear: 2025,
		CIK:        "0001234567",
		FilerType:  "LAF",
		IsEGC:      false,
	}

	report, err := mapper.BuildReport(ctx, input)
	if err != nil {
		t.Fatalf("BuildReport failed: %v", err)
	}

	// Verify basic structure
	if report.OrgID != input.OrgID {
		t.Errorf("Expected OrgID %s, got %s", input.OrgID, report.OrgID)
	}

	if report.OrgName != input.OrgName {
		t.Errorf("Expected OrgName %s, got %s", input.OrgName, report.OrgName)
	}

	if report.CIK != input.CIK {
		t.Errorf("Expected CIK %s, got %s", input.CIK, report.CIK)
	}

	if report.FiscalYear != input.FiscalYear {
		t.Errorf("Expected FiscalYear %d, got %d", input.FiscalYear, report.FiscalYear)
	}

	if report.FilerType != input.FilerType {
		t.Errorf("Expected FilerType %s, got %s", input.FilerType, report.FilerType)
	}

	if report.FilingType != "10-K" {
		t.Errorf("Expected FilingType 10-K, got %s", report.FilingType)
	}

	if report.GeneratedAt.IsZero() {
		t.Error("GeneratedAt should be set")
	}
}

func TestDefaultSECMapper_BuildReport_RequiredDisclosures_LAF(t *testing.T) {
	mapper := NewDefaultSECMapper()
	ctx := context.Background()

	input := createCompleteInputLAF()

	report, err := mapper.BuildReport(ctx, input)
	if err != nil {
		t.Fatalf("BuildReport failed: %v", err)
	}

	// Verify required disclosures for LAF
	expectedDisclosures := map[string]bool{
		"Item 1500":        true, // Governance
		"Item 1501":        true, // Risk Management
		"Item 1502":        true, // Strategy
		"Item 1504":        true, // GHG Metrics
		"Reg S-X Art. 14":  true, // Financial Impact
		"Attestation":      true, // Required for LAF FY2025+
	}

	for _, disclosure := range report.RequiredDisclosures {
		if disclosure.Required {
			if _, exists := expectedDisclosures[disclosure.Item]; !exists {
				t.Errorf("Unexpected required disclosure: %s", disclosure.Item)
			}
			delete(expectedDisclosures, disclosure.Item)
		}
	}

	if len(expectedDisclosures) > 0 {
		for item := range expectedDisclosures {
			t.Errorf("Missing required disclosure: %s", item)
		}
	}
}

func TestDefaultSECMapper_BuildReport_RequiredDisclosures_SRC(t *testing.T) {
	mapper := NewDefaultSECMapper()
	ctx := context.Background()

	input := SECInput{
		OrgID:      "org-src-001",
		OrgName:    "Small Corp",
		FiscalYear: 2025,
		CIK:        "0009876543",
		FilerType:  "SRC",
		IsEGC:      false,
		Governance: &GovernanceDisclosure{
			BoardOversight: BoardOversightInfo{
				HasBoardOversight:    true,
				ResponsibleCommittee: "Audit Committee",
			},
			ManagementRole: ManagementRoleInfo{
				ResponsibleExecutive: "CFO",
			},
		},
	}

	report, err := mapper.BuildReport(ctx, input)
	if err != nil {
		t.Fatalf("BuildReport failed: %v", err)
	}

	// SRC should NOT be required to disclose GHG metrics
	for _, disclosure := range report.RequiredDisclosures {
		if disclosure.Item == "Item 1504" && disclosure.Required {
			t.Error("Item 1504 (GHG Metrics) should not be required for SRC filers")
		}
		if disclosure.Item == "Attestation" && disclosure.Required {
			t.Error("Attestation should not be required for SRC filers")
		}
	}
}

func TestDefaultSECMapper_BuildReport_EGC_Exemption(t *testing.T) {
	mapper := NewDefaultSECMapper()
	ctx := context.Background()

	input := SECInput{
		OrgID:      "org-egc-001",
		OrgName:    "Emerging Growth Company",
		FiscalYear: 2025,
		CIK:        "0001111111",
		FilerType:  "EGC",
		IsEGC:      true,
	}

	report, err := mapper.BuildReport(ctx, input)
	if err != nil {
		t.Fatalf("BuildReport failed: %v", err)
	}

	// EGC should have NO required disclosures
	for _, disclosure := range report.RequiredDisclosures {
		if disclosure.Required {
			t.Errorf("EGC should have no required disclosures, but %s is marked as required", disclosure.Item)
		}
	}
}

func TestDefaultSECMapper_CalculateComplianceScore(t *testing.T) {
	mapper := NewDefaultSECMapper()

	tests := []struct {
		name        string
		disclosures []DisclosureStatus
		expected    float64
	}{
		{
			name: "All complete",
			disclosures: []DisclosureStatus{
				{Required: true, Complete: true},
				{Required: true, Complete: true},
				{Required: true, Complete: true},
			},
			expected: 100.0,
		},
		{
			name: "Half complete",
			disclosures: []DisclosureStatus{
				{Required: true, Complete: true},
				{Required: true, Complete: false},
				{Required: true, Complete: true},
				{Required: true, Complete: false},
			},
			expected: 50.0,
		},
		{
			name: "None complete",
			disclosures: []DisclosureStatus{
				{Required: true, Complete: false},
				{Required: true, Complete: false},
			},
			expected: 0.0,
		},
		{
			name: "Only optional",
			disclosures: []DisclosureStatus{
				{Required: false, Complete: true},
				{Required: false, Complete: false},
			},
			expected: 100.0, // No required disclosures
		},
		{
			name: "Mixed required and optional",
			disclosures: []DisclosureStatus{
				{Required: true, Complete: true},
				{Required: false, Complete: true},
				{Required: true, Complete: false},
				{Required: false, Complete: false},
			},
			expected: 50.0, // 1 out of 2 required
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := mapper.calculateComplianceScore(tt.disclosures)
			if score != tt.expected {
				t.Errorf("Expected score %.1f, got %.1f", tt.expected, score)
			}
		})
	}
}

func TestDefaultSECMapper_GovernanceCompletenesss(t *testing.T) {
	mapper := NewDefaultSECMapper()

	tests := []struct {
		name       string
		governance *GovernanceDisclosure
		complete   bool
	}{
		{
			name: "Complete governance",
			governance: &GovernanceDisclosure{
				BoardOversight: BoardOversightInfo{
					HasBoardOversight:    true,
					ResponsibleCommittee: "Risk Committee",
				},
				ManagementRole: ManagementRoleInfo{
					ResponsibleExecutive: "Chief Sustainability Officer",
				},
			},
			complete: true,
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
			complete: false,
		},
		{
			name: "Missing executive",
			governance: &GovernanceDisclosure{
				BoardOversight: BoardOversightInfo{
					HasBoardOversight:    true,
					ResponsibleCommittee: "Audit Committee",
				},
			},
			complete: false,
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
			complete: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complete := mapper.isGovernanceComplete(tt.governance)
			if complete != tt.complete {
				t.Errorf("Expected complete=%v, got %v", tt.complete, complete)
			}
		})
	}
}

func TestDefaultSECMapper_GHGMetricsCompleteness(t *testing.T) {
	mapper := NewDefaultSECMapper()

	tests := []struct {
		name     string
		metrics  *GHGMetricsDisclosure
		complete bool
	}{
		{
			name: "Complete metrics",
			metrics: &GHGMetricsDisclosure{
				Scope1Emissions: &ScopeEmissions{
					TotalEmissions: 1000.0,
				},
				Scope2Emissions: &ScopeEmissions{
					TotalEmissions: 500.0,
				},
				Methodology: MethodologyDisclosure{
					Standard: "GHG Protocol Corporate Standard",
				},
			},
			complete: true,
		},
		{
			name: "Missing Scope 1",
			metrics: &GHGMetricsDisclosure{
				Scope2Emissions: &ScopeEmissions{
					TotalEmissions: 500.0,
				},
				Methodology: MethodologyDisclosure{
					Standard: "GHG Protocol",
				},
			},
			complete: false,
		},
		{
			name: "Missing Scope 2",
			metrics: &GHGMetricsDisclosure{
				Scope1Emissions: &ScopeEmissions{
					TotalEmissions: 1000.0,
				},
				Methodology: MethodologyDisclosure{
					Standard: "GHG Protocol",
				},
			},
			complete: false,
		},
		{
			name: "Missing methodology",
			metrics: &GHGMetricsDisclosure{
				Scope1Emissions: &ScopeEmissions{
					TotalEmissions: 1000.0,
				},
				Scope2Emissions: &ScopeEmissions{
					TotalEmissions: 500.0,
				},
			},
			complete: false,
		},
		{
			name: "Zero Scope 2 is valid",
			metrics: &GHGMetricsDisclosure{
				Scope1Emissions: &ScopeEmissions{
					TotalEmissions: 1000.0,
				},
				Scope2Emissions: &ScopeEmissions{
					TotalEmissions: 0.0,
				},
				Methodology: MethodologyDisclosure{
					Standard: "GHG Protocol",
				},
			},
			complete: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complete := mapper.isGHGMetricsComplete(tt.metrics)
			if complete != tt.complete {
				t.Errorf("Expected complete=%v, got %v", tt.complete, complete)
			}
		})
	}
}

func TestDefaultSECMapper_AttestationRequirements(t *testing.T) {
	mapper := NewDefaultSECMapper()
	ctx := context.Background()

	tests := []struct {
		name                 string
		filerType            string
		fiscalYear           int
		attestationRequired  bool
	}{
		{
			name:                "LAF FY2025 requires attestation",
			filerType:           "LAF",
			fiscalYear:          2025,
			attestationRequired: true,
		},
		{
			name:                "LAF FY2024 no attestation",
			filerType:           "LAF",
			fiscalYear:          2024,
			attestationRequired: false,
		},
		{
			name:                "AF no attestation",
			filerType:           "AF",
			fiscalYear:          2025,
			attestationRequired: false,
		},
		{
			name:                "SRC no attestation",
			filerType:           "SRC",
			fiscalYear:          2025,
			attestationRequired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := SECInput{
				OrgID:      "org-test",
				OrgName:    "Test Corp",
				FiscalYear: tt.fiscalYear,
				CIK:        "0001234567",
				FilerType:  tt.filerType,
			}

			report, err := mapper.BuildReport(ctx, input)
			if err != nil {
				t.Fatalf("BuildReport failed: %v", err)
			}

			hasAttestation := false
			for _, disclosure := range report.RequiredDisclosures {
				if disclosure.Item == "Attestation" && disclosure.Required {
					hasAttestation = true
					break
				}
			}

			if hasAttestation != tt.attestationRequired {
				t.Errorf("Expected attestationRequired=%v, got %v", tt.attestationRequired, hasAttestation)
			}
		})
	}
}

// Helper function to create a complete LAF input
func createCompleteInputLAF() SECInput {
	return SECInput{
		OrgID:      "org-laf-001",
		OrgName:    "Large Accelerated Filer Corp",
		FiscalYear: 2025,
		CIK:        "0001234567",
		FilerType:  "LAF",
		IsEGC:      false,
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
				ProcessDescription: "Annual climate risk assessment process",
				RiskCategories:     []string{"physical", "transition"},
				TimeHorizons:       []string{"short-term", "medium-term", "long-term"},
			},
			RiskManagement: RiskManagementProcess{
				ProcessDescription: "Integrated risk management framework",
			},
			MaterialRisks: []MaterialClimateRisk{
				{
					RiskID:       "RISK-001",
					RiskType:     "physical",
					Description:  "Increased flooding at coastal facilities",
					TimeHorizon:  "long-term",
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
					Description: "Supply chain disruption from extreme weather",
					TimeHorizon: "medium-term",
				},
			},
		},
		GHGMetrics: &GHGMetricsDisclosure{
			IsRequired: true,
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
			Provider:         "Big Four Accounting Firm",
			Standard:         "AT-C 210",
			OpinionType:      "unmodified",
			OpinionStatement: "In our opinion, the GHG emissions are fairly stated.",
			ScopesCovered:    []string{"Scope 1", "Scope 2"},
			ReportDate:       time.Now(),
		},
	}
}
