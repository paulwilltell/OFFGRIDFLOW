package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/example/offgridflow/internal/compliance"

	"github.com/google/uuid"
)

// TestDataset represents the JSON structure from testdata files
type TestDataset struct {
	Organization struct {
		Name      string `json:"name"`
		Sector    string `json:"sector"`
		Employees int    `json:"employees"`
		Locations []struct {
			City    string `json:"city"`
			Country string `json:"country"`
		} `json:"locations"`
	} `json:"organization"`
	ReportingPeriod struct {
		Year  int    `json:"year"`
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"reporting_period"`
	Activities []struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Scope       string  `json:"scope"`
		Category    string  `json:"category"`
		Amount      float64 `json:"amount"`
		Unit        string  `json:"unit"`
		Emissions   float64 `json:"emissions_tco2e"`
		Location    string  `json:"location"`
		DataQuality string  `json:"data_quality"`
	} `json:"activities"`
	Summary struct {
		Scope1      float64 `json:"scope1_tco2e"`
		Scope2      float64 `json:"scope2_tco2e"`
		Scope3      float64 `json:"scope3_tco2e"`
		Total       float64 `json:"total_tco2e"`
		Quality     float64 `json:"data_quality_score"`
		Completeness float64 `json:"completeness_percentage"`
	} `json:"summary"`
}

func main() {
	fmt.Println("🎯 Generating Example Compliance Reports...")
	fmt.Println()

	// Create output directory
	os.MkdirAll("examples/reports", 0755)

	// Generate reports from each test dataset
	datasets := []struct {
		file       string
		reportType string
		outputName string
	}{
		{
			file:       "testdata/manufacturing_company_2024.json",
			reportType: "CSRD",
			outputName: "csrd-manufacturing-2024.pdf",
		},
		{
			file:       "testdata/tech_company_2024.json",
			reportType: "SEC",
			outputName: "sec-tech-company-2024.pdf",
		},
		{
			file:       "testdata/retail_company_2024.json",
			reportType: "California",
			outputName: "california-retail-2024.pdf",
		},
		{
			file:       "testdata/manufacturing_company_2024.json",
			reportType: "CBAM",
			outputName: "cbam-manufacturing-2024.pdf",
		},
		{
			file:       "testdata/tech_company_2024.json",
			reportType: "IFRS",
			outputName: "ifrs-tech-company-2024.pdf",
		},
	}

	successCount := 0
	for _, ds := range datasets {
		fmt.Printf("📄 Generating %s from %s...\n", ds.reportType, ds.file)

		// Load test data
		data, err := os.ReadFile(ds.file)
		if err != nil {
			fmt.Printf("   ❌ Error reading file: %v\n", err)
			continue
		}

		var testData TestDataset
		if err := json.Unmarshal(data, &testData); err != nil {
			fmt.Printf("   ❌ Error parsing JSON: %v\n", err)
			continue
		}

		// Convert to compliance structures
		emissionsData := convertToEmissionsData(testData)
		qualityMetrics := convertToQualityMetrics(testData)
		baseReport := createBaseReport(testData, ds.reportType)

		var pdfBytes []byte
		var genErr error

		// Generate appropriate report type
		switch ds.reportType {
		case "CSRD":
			csrdReport := compliance.CSRDReport{
			Report:           baseReport,
			EmissionsData:    emissionsData,
			QualityMetrics:   qualityMetrics,
			OrganizationName: testData.Organization.Name,
			ReportingOfficer: "Chief Sustainability Officer",
			}
			pdfBytes, genErr = compliance.GenerateCSRDPDF(csrdReport)

		case "SEC":
			secReport := compliance.SECReport{
				Report:           baseReport,
				EmissionsData:    emissionsData,
				QualityMetrics:   qualityMetrics,
				OrganizationName: testData.Organization.Name,
				ReportingOfficer: "Chief Financial Officer",
				FiscalYear:       testData.ReportingPeriod.Year,
				ClimateRisks: []compliance.ClimateRisk{
					{
						Category:    "Physical",
						Description: "Extreme weather events affecting operations",
						Impact:      "Medium",
						Mitigation:  "Enhanced facility resilience and backup systems",
					},
					{
						Category:    "Transition",
						Description: "Carbon pricing and regulatory changes",
						Impact:      "High",
						Mitigation:  "Emissions reduction targets and renewable energy investments",
					},
				},
				TransitionPlan:       "Committed to net-zero by 2050 with interim targets",
				GovernanceStructure:  "Board-level climate committee with quarterly reviews",
				Targets: []compliance.EmissionTarget{
					{
						Scope:           "Scope 1+2",
						BaselineYear:    2020,
						BaselineTonnes:  baseReport.Scope1EmissionsTonnes + baseReport.Scope2EmissionsTonnes,
						TargetYear:      2030,
						TargetReduction: 50.0,
						Status:          "On Track",
					},
				},
			}
			pdfBytes, genErr = compliance.GenerateSECPDF(secReport)

		case "California":
			calReport := compliance.CaliforniaReport{
				Report:           baseReport,
				EmissionsData:    emissionsData,
				QualityMetrics:   qualityMetrics,
				OrganizationName: testData.Organization.Name,
				ReportingOfficer: "VP of Sustainability",
				ReportingYear:    testData.ReportingPeriod.Year,
				AnnualRevenue:    1500000000.0, // $1.5B
				CAOperations:     true,
				Scope3Categories: generateScope3Categories(emissionsData),
				Assurance: compliance.AssuranceInfo{
					Provider:    "Independent Verification Inc.",
					Level:       "Limited Assurance",
					Standard:    "ISO 14064-3",
					OpinionDate: "2024-03-15",
					Opinion:     "Reasonable",
				},
			}
			pdfBytes, genErr = compliance.GenerateCaliforniaPDF(calReport)

		case "CBAM":
			cbamReport := compliance.CBAMReport{
				Report:           baseReport,
				EmissionsData:    emissionsData,
				QualityMetrics:   qualityMetrics,
				OrganizationName: testData.Organization.Name,
				ReportingYear:    testData.ReportingPeriod.Year,
				ReportingOfficer: "CBAM Compliance Officer",
				InstallationID:   "EU-CBAM-2024-" + uuid.New().String()[:8],
				OperatorName:     testData.Organization.Name,
				TotalCarbonPrice: baseReport.TotalEmissionsTonnes * 85.0, // €85/tonne
				ImportedGoods: []compliance.ImportedGood{
					{
						CNCode:            "7208 10 00",
						Description:       "Hot-rolled steel products",
						Quantity:          1000.0,
						Unit:              "tonnes",
						OriginCountry:     "China",
						EmbeddedEmissions: 2.5,
						TotalEmissions:    2500.0,
						CBAMPrice:         212500.0,
						ProductionRoute:   "Basic Oxygen Furnace",
					},
				},
				ProductionSites: []compliance.ProductionSite{
					{
						Name:              "Primary Manufacturing Facility",
						Location:          "Shanghai",
						Country:           "China",
						Coordinates:       "31.2304° N, 121.4737° E",
						ProductionProcess: "Integrated Steel Mill",
						EmissionsFactor:   2.5,
						Certified:         true,
						CertificationBody: "China Quality Certification Centre",
					},
				},
			}
			pdfBytes, genErr = compliance.GenerateCBAMPDF(cbamReport)

		case "IFRS":
			ifrsReport := compliance.IFRSS2Report{
				Report:           baseReport,
				EmissionsData:    emissionsData,
				QualityMetrics:   qualityMetrics,
				OrganizationName: testData.Organization.Name,
				ReportingYear:    testData.ReportingPeriod.Year,
				ReportingOfficer: "Chief Sustainability Officer",
				ClimateRisks: []compliance.ClimateRisk{
					{
						Category:    "Physical",
						Description: "Increased frequency of extreme weather events",
						Impact:      "Medium - potential facility damage",
						Mitigation:  "Infrastructure hardening and insurance coverage",
					},
					{
						Category:    "Transition",
						Description: "Carbon pricing mechanisms",
						Impact:      "High - cost implications",
						Mitigation:  "Emissions reduction program and renewable energy",
					},
				},
				Opportunities: []compliance.ClimateOpportunity{
					{
						Category:    "Products/Services",
						Description: "Growing demand for low-carbon solutions",
						Potential:   "High",
						Strategy:    "Develop carbon-neutral product lines",
						Timeline:    "2024-2027",
					},
				},
				FinancialImpacts: []compliance.FinancialImpact{
					{
						Category:    "Carbon Pricing",
						Description: "Expected carbon tax costs",
						Amount:      5000000.0,
						Currency:    "USD",
						Timeframe:   "2025-2030",
						Likelihood:  "High",
						ImpactType:  "Cost",
					},
				},
				Scenarios: []compliance.ClimateScenario{
					{
						Name:            "Orderly Transition",
						Description:     "Gradual policy implementation aligned with 1.5°C",
						TemperatureGoal: "1.5°C",
						Assumptions:     "Carbon pricing reaches $150/tonne by 2030",
						Resilience:      "Strong - business model adapts well",
						KeyFindings: []string{
							"Revenue impact: -2% to -5% by 2030",
							"Capital expenditure needed: $50M renewable energy",
							"Operating cost reduction: -15% energy efficiency gains",
						},
					},
				},
				Metrics: []compliance.SustainabilityMetric{
					{
						Name:       "Energy Intensity",
						Value:      125.5,
						Unit:       "kWh per $1M revenue",
						Baseline:   150.0,
						Target:     100.0,
						TargetYear: 2030,
						Progress:   49.0,
					},
				},
				Targets: []compliance.EmissionTarget{
					{
						Scope:           "Scope 1+2",
						BaselineYear:    2020,
						BaselineTonnes:  baseReport.Scope1EmissionsTonnes + baseReport.Scope2EmissionsTonnes,
						TargetYear:      2030,
						TargetReduction: 50.0,
						Status:          "On Track",
					},
				},
				TransitionPlan: compliance.TransitionPlan{
					Overview:   "Comprehensive decarbonization strategy aligned with SBTi",
					Investment: 75000000.0,
					Currency:   "USD",
					Timeframe:  "2024-2030",
					Milestones: []compliance.Milestone{
						{Year: 2025, Description: "100% renewable electricity", Status: "In Progress"},
						{Year: 2027, Description: "Electric fleet conversion complete", Status: "Planned"},
						{Year: 2030, Description: "50% emissions reduction achieved", Status: "On Track"},
					},
				},
				GovernanceStructure: compliance.GovernanceInfo{
					BoardOversight:     "Sustainability Committee meets quarterly",
					ManagementRole:     "CSO reports directly to CEO",
					Expertise:          "Board includes climate science expert",
					Integration:        "Climate metrics in executive compensation",
					IncentiveAlignment: "20% of annual bonus tied to emissions targets",
				},
			}
			pdfBytes, genErr = compliance.GenerateIFRSS2PDF(ifrsReport)
		}

		if genErr != nil {
			fmt.Printf("   ❌ Error generating PDF: %v\n", genErr)
			continue
		}

		// Write PDF file
		outputPath := "examples/reports/" + ds.outputName
		if err := os.WriteFile(outputPath, pdfBytes, 0644); err != nil {
			fmt.Printf("   ❌ Error writing PDF: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ Generated: %s (%.1f KB)\n", outputPath, float64(len(pdfBytes))/1024)
		successCount++
	}

	fmt.Println()
	fmt.Printf("🎉 Success! Generated %d/%d example reports\n", successCount, len(datasets))
	fmt.Println()
	fmt.Println("📁 Reports saved to: examples/reports/")
}

func convertToEmissionsData(data TestDataset) compliance.EmissionsData {
	activities := make([]compliance.ActivityEmission, 0)

	for _, act := range data.Activities {
		activities = append(activities, compliance.ActivityEmission{
		ID:              uuid.New(),
		Name:            act.Name,
		Scope:           act.Scope,
		Category:        act.Category,
		Quantity:        act.Amount,
		Unit:            act.Unit,
		EmissionsTonnes: act.Emissions,
		EmissionFactor:  0.0,
		 Location:        act.Location,
				Period:          "2024",
			})
	}

	return compliance.EmissionsData{
		Scope1Tonnes: data.Summary.Scope1,
		Scope2Tonnes: data.Summary.Scope2,
		Scope3Tonnes: data.Summary.Scope3,
		TotalTonnes:  data.Summary.Total,
		Activities:   activities,
	}
}

func convertToQualityMetrics(data TestDataset) compliance.DataQualityMetrics {
	totalActivities := len(data.Activities)
	completeActivities := 0

	for _, act := range data.Activities {
		if act.DataQuality == "High" || act.DataQuality == "Medium" {
			completeActivities++
		}
	}

	warnings := []string{}
	if data.Summary.Quality < 90 {
		warnings = append(warnings, "Some emission factors may be outdated")
	}
	if data.Summary.Completeness < 95 {
		warnings = append(warnings, "Incomplete data for some activity categories")
	}

	return compliance.DataQualityMetrics{
		TotalActivities:        totalActivities,
		CompleteActivities:     completeActivities,
		CompletenessPercentage: data.Summary.Completeness,
		MissingFields:          make(map[string]int),
		LowQualityFactors:      0,
		OutdatedFactors:        0,
		DataQualityScore:       data.Summary.Quality,
		Warnings:               warnings,
	}
}

func createBaseReport(data TestDataset, reportType string) *compliance.Report {
	startTime, _ := time.Parse("2006-01-02", data.ReportingPeriod.Start)
	endTime, _ := time.Parse("2006-01-02", data.ReportingPeriod.End)

	tenantID := uuid.New()
	userID := uuid.New()

	report := &compliance.Report{
		ID:                     uuid.New(),
		TenantID:               tenantID,
		ReportType:             compliance.ReportType(reportType),
		ReportingYear:          data.ReportingPeriod.Year,
		PeriodStart:            startTime,
		PeriodEnd:              endTime,
		ReportHash:             "",
		DataQualityScore:       data.Summary.Quality,
		CompletenessPercentage: data.Summary.Completeness,
		CalculationMethodology: "GHG Protocol Corporate Standard",
		Scope1EmissionsTonnes:  data.Summary.Scope1,
		Scope2EmissionsTonnes:  data.Summary.Scope2,
		Scope3EmissionsTonnes:  data.Summary.Scope3,
		TotalEmissionsTonnes:   data.Summary.Total,
		GeneratedBy:            &userID,
		GenerationTimestamp:    time.Now(),
		Status:                 compliance.StatusDraft,
		Version:                1,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	// Calculate report hash
	hashInput := fmt.Sprintf("%s-%d-%.2f", reportType, data.ReportingPeriod.Year, data.Summary.Total)
	report.ReportHash = compliance.CalculateReportHash(report, []byte(hashInput))

	return report
}

func generateScope3Categories(emissionsData compliance.EmissionsData) []compliance.Scope3Category {
	categories := make([]compliance.Scope3Category, 0)

	categoryMap := make(map[string]float64)
	for _, act := range emissionsData.Activities {
		if act.Scope == "Scope 3" {
			categoryMap[act.Category] += act.EmissionsTonnes
		}
	}

	categoryNum := 1
	for cat, emissions := range categoryMap {
		categories = append(categories, compliance.Scope3Category{
			Number:      categoryNum,
			Name:        cat,
			Emissions:   emissions,
			Methodology: "Supplier-specific data or industry averages",
			DataQuality: "Medium to High",
		})
		categoryNum++
	}

	return categories
}
