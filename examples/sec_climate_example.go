//go:build ignore

// Package main demonstrates the SEC Climate compliance module usage.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/example/offgridflow/internal/compliance/sec"
	"github.com/example/offgridflow/internal/models"
)

func main() {
	fmt.Println("=== SEC Climate Disclosure Example ===\n")

	// Example 1: Basic LAF Report
	fmt.Println("Example 1: Large Accelerated Filer (LAF) - FY2025")
	lafExample()

	fmt.Println("\n---\n")

	// Example 2: SRC Filer (Smaller Reporting Company)
	fmt.Println("Example 2: Smaller Reporting Company (SRC) - FY2025")
	srcExample()

	fmt.Println("\n---\n")

	// Example 3: Validation and Error Handling
	fmt.Println("Example 3: Validation and Error Handling")
	validationExample()
}

// lafExample demonstrates a complete LAF filer disclosure
func lafExample() {
	mapper := sec.NewDefaultSECMapper()
	validator := sec.NewValidator()
	builder := sec.NewReportBuilder()

	// Prepare input data for a Large Accelerated Filer
	input := sec.SECInput{
		OrgID:      "org-example-001",
		OrgName:    "Example Energy Corporation",
		CIK:        "0001234567",
		FiscalYear: 2025,
		FilerType:  "LAF",
		IsEGC:      false,
		UnifiedData: &models.UnifiedESGData{
			Emissions: models.EmissionsData{
				Scope1: 150000.0, // tCO2e
				Scope2: models.Scope2Emissions{
					LocationBased: 75000.0,
					MarketBased:   65000.0,
				},
				Scope3: models.Scope3Emissions{
					Total: 500000.0,
					Categories: map[int]models.Scope3Category{
						1: {
							Number:    1,
							Name:      "Purchased Goods and Services",
							Emissions: 200000.0,
						},
						11: {
							Number:    11,
							Name:      "Use of Sold Products",
							Emissions: 300000.0,
						},
					},
				},
				Year:         2025,
				Methodology:  "GHG Protocol Corporate Standard",
				Verification: "limited_assurance",
			},
			Governance: models.GovernanceData{
				BoardOversight:   true,
				ClimateCommittee: "Sustainability Committee",
				ExecutiveRole:    "Chief Sustainability Officer",
			},
		},
		// Additional context
		Governance: &sec.GovernanceDisclosure{
			BoardOversight: sec.BoardOversightInfo{
				HasBoardOversight:      true,
				ResponsibleCommittee:   "Sustainability Committee",
				OversightFrequency:     "quarterly",
				DirectorsWithExpertise: []string{"Dr. Jane Climate", "Prof. John Environment"},
			},
			ManagementRole: sec.ManagementRoleInfo{
				ResponsibleExecutive:  "Chief Sustainability Officer",
				ProcessesAndFrequency: "Monthly climate risk review meetings with executive team",
				ReportingStructure:    "CSO reports directly to CEO",
			},
		},
		RiskManagement: &sec.RiskManagementDisclosure{
			RiskIdentification: sec.RiskIdentificationProcess{
				ProcessDescription: "Annual comprehensive climate risk assessment covering all major facilities",
				RiskCategories:     []string{"physical", "transition"},
				TimeHorizons:       []string{"short-term", "medium-term", "long-term"},
			},
			RiskManagement: sec.RiskManagementProcess{
				ProcessDescription: "Integrated into enterprise risk management framework",
			},
			MaterialRisks: []sec.MaterialClimateRisk{
				{
					RiskType:       "physical",
					Description:    "Increased flooding risk at coastal facilities due to sea level rise",
					TimeHorizon:    "long-term",
					MitigationPlan: "Facility hardening and long-term relocation planning",
				},
				{
					RiskType:       "transition",
					Description:    "Carbon pricing exposure in key operating jurisdictions",
					TimeHorizon:    "medium-term",
					MitigationPlan: "Energy efficiency improvements and renewable energy procurement",
				},
			},
			ERMIntegration: sec.ERMIntegrationInfo{
				IsIntegrated:        true,
				ERMFramework:        "COSO ERM Framework",
				IntegrationApproach: "Climate risks reviewed quarterly as part of enterprise risk register",
			},
		},
		Strategy: &sec.StrategyDisclosure{
			MaterialImpacts: []sec.StrategyImpact{
				{
					ImpactArea:      "operations",
					Description:     "Potential supply chain disruption from extreme weather events",
					TimeHorizon:     "medium-term",
					ResponseActions: "Supplier diversification and increased inventory buffers",
				},
			},
			ClimateTargets: []sec.ClimateTarget{
				{
					Description:    "Reduce absolute Scope 1 and 2 GHG emissions 50% by 2030",
					TargetYear:     2030,
					BaseYear:       2020,
					BaselineValue:  450000.0,
					TargetValue:    225000.0,
					Unit:           "tCO2e",
					Scope:          "Scope 1+2",
					SBTiAligned:    true,
					ProgressToDate: 12.5,
				},
			},
			InternalCarbonPrice: &sec.InternalCarbonPriceInfo{
				Used:            true,
				PricePerTonCO2e: 50.00,
				PriceType:       "shadow price",
				Rationale:       "Used for internal capital allocation decisions",
			},
		},
		Attestation: &sec.AttestationReport{
			Required:         true,
			AssuranceLevel:   "limited",
			Provider:         "Independent Assurance Services LLC",
			Standard:         "AT-C 210",
			OpinionType:      "unmodified",
			OpinionStatement: "In our opinion, the GHG emissions disclosures are fairly stated, in all material respects.",
			ScopesCovered:    []string{"Scope 1", "Scope 2"},
			ReportDate:       time.Now(),
		},
	}

	// Build SEC report
	report, err := mapper.BuildReport(context.Background(), input)
	if err != nil {
		fmt.Printf("Error building report: %v\n", err)
		return
	}

	// Validate report
	results := validator.ValidateReport(*report)
	fmt.Printf("Validation: %s\n", validationStatus(results.Valid))
	fmt.Printf("Compliance Score: %.1f%%\n", report.ComplianceScore)

	if len(results.Errors) > 0 {
		fmt.Println("\nValidation Errors:")
		for _, err := range results.Errors {
			fmt.Printf("  - [%s] %s: %s\n", err.Code, err.Field, err.Message)
		}
	}

	// Generate 10-K formatted report
	report10K, err := builder.Build10KReport(context.Background(), input)
	if err != nil {
		fmt.Printf("Error building 10-K report: %v\n", err)
		return
	}

	// Display portions of the formatted report
	fmt.Println("\n--- 10-K Report Preview ---")
	fmt.Println(report10K.Header)
	fmt.Println("\n" + report10K.Item1504GHGMetrics[:min(len(report10K.Item1504GHGMetrics), 500)] + "...\n")
}

// srcExample demonstrates an SRC filer (mostly voluntary)
func srcExample() {
	mapper := sec.NewDefaultSECMapper()

	input := sec.SECInput{
		OrgID:      "org-small-001",
		OrgName:    "Small Tech Startup Inc.",
		CIK:        "0009876543",
		FiscalYear: 2025,
		FilerType:  "SRC",
		UnifiedData: &models.UnifiedESGData{
			Emissions: models.EmissionsData{
				Scope1: 500.0,
				Scope2: models.Scope2Emissions{
					LocationBased: 1000.0,
					MarketBased:   800.0,
				},
			},
		},
	}

	report, err := mapper.BuildReport(context.Background(), input)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Organization: %s\n", report.OrgName)
	fmt.Printf("Filer Type: %s (Smaller Reporting Company)\n", report.FilerType)
	fmt.Println("\nRequired Disclosures:")
	for _, disc := range report.RequiredDisclosures {
		status := "Optional"
		if disc.Required {
			status = "Required"
		}
		complete := "❌"
		if disc.Complete {
			complete = "✓"
		}
		fmt.Printf("  %s %s - %s\n", complete, disc.Item, status)
	}
}

// validationExample demonstrates validation and error handling
func validationExample() {
	validator := sec.NewValidator()

	// Create an incomplete report (missing required fields)
	incompleteReport := sec.SECReport{
		OrgID:   "org-incomplete",
		OrgName: "Incomplete Corp",
		// CIK:        "", // Missing - should cause error
		FiscalYear: 2025,
		FilerType:  "LAF",
		Governance: &sec.GovernanceDisclosure{
			BoardOversight: sec.BoardOversightInfo{
				HasBoardOversight: true,
				// ResponsibleCommittee: "", // Missing - should cause error
			},
		},
		// GHGMetrics: nil, // Missing - should cause error for LAF
	}

	results := validator.ValidateReport(incompleteReport)

	fmt.Printf("Validation Status: %s\n", validationStatus(results.Valid))
	fmt.Printf("Errors: %d\n", len(results.Errors))
	fmt.Printf("Warnings: %d\n\n", len(results.Warnings))

	if len(results.Errors) > 0 {
		fmt.Println("Validation Errors:")
		for i, err := range results.Errors {
			fmt.Printf("%d. Field: %s\n", i+1, err.Field)
			fmt.Printf("   Code: %s\n", err.Code)
			fmt.Printf("   Message: %s\n\n", err.Message)
		}
	}

	if len(results.Warnings) > 0 {
		fmt.Println("Warnings:")
		for i, warn := range results.Warnings {
			fmt.Printf("%d. %s: %s\n", i+1, warn.Field, warn.Message)
		}
	}

	// Export validation results as JSON
	resultsJSON, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println("\nValidation Results (JSON):")
	fmt.Println(string(resultsJSON))
}

// Helper functions

func validationStatus(valid bool) string {
	if valid {
		return "✓ PASS"
	}
	return "✗ FAIL"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
