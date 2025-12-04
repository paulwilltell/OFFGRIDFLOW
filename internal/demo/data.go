package demo

import (
	"context"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// DemoData represents all demo data for investor presentations.
type DemoData struct {
	Organization DemoOrganization
	Emissions    []DemoEmission
	Compliance   DemoCompliance
	TrendData    []DemoTrendPoint
	Benchmarks   []DemoBenchmark
}

// DemoOrganization represents a demo company profile.
type DemoOrganization struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Industry string   `json:"industry"`
	Size     string   `json:"size"`
	Regions  []string `json:"regions"`
	Logo     string   `json:"logo"`
}

// DemoEmission represents a single emission record.
type DemoEmission struct {
	ID            string    `json:"id"`
	Source        string    `json:"source"`
	Scope         int       `json:"scope"`
	Category      string    `json:"category"`
	ActivityType  string    `json:"activityType"`
	QuantityValue float64   `json:"quantityValue"`
	QuantityUnit  string    `json:"quantityUnit"`
	EmissionsCO2e float64   `json:"emissionsCO2e"`
	Region        string    `json:"region"`
	Facility      string    `json:"facility"`
	Period        string    `json:"period"`
	DataQuality   string    `json:"dataQuality"`
	Verified      bool      `json:"verified"`
	CreatedAt     time.Time `json:"createdAt"`
}

// DemoCompliance represents compliance readiness status.
type DemoCompliance struct {
	CSRD       DemoFrameworkStatus `json:"csrd"`
	SEC        DemoFrameworkStatus `json:"sec"`
	CBAM       DemoFrameworkStatus `json:"cbam"`
	California DemoFrameworkStatus `json:"california"`
	Overall    float64             `json:"overall"`
}

// DemoFrameworkStatus represents status for a compliance framework.
type DemoFrameworkStatus struct {
	Name         string  `json:"name"`
	Readiness    float64 `json:"readiness"`
	Status       string  `json:"status"`
	DueDate      string  `json:"dueDate,omitempty"`
	Requirements int     `json:"requirements"`
	Completed    int     `json:"completed"`
}

// DemoTrendPoint represents a data point in trend charts.
type DemoTrendPoint struct {
	Period    string  `json:"period"`
	Scope1    float64 `json:"scope1"`
	Scope2    float64 `json:"scope2"`
	Scope3    float64 `json:"scope3"`
	Total     float64 `json:"total"`
	Intensity float64 `json:"intensity"` // tCO2e per $M revenue
}

// DemoBenchmark represents industry benchmark comparison.
type DemoBenchmark struct {
	Category    string  `json:"category"`
	YourValue   float64 `json:"yourValue"`
	IndustryAvg float64 `json:"industryAvg"`
	BestInClass float64 `json:"bestInClass"`
	Unit        string  `json:"unit"`
	PercentDiff float64 `json:"percentDiff"`
}

// DemoConfig holds demo mode configuration.
type DemoConfig struct {
	// Enabled determines if demo mode is active.
	Enabled bool `json:"enabled"`
	// CompanyName is the demo company name.
	CompanyName string `json:"companyName"`
	// Industry is the demo industry sector.
	Industry string `json:"industry"`
	// BaseYear is the starting year for trend data.
	BaseYear int `json:"baseYear"`
	// ShowAIDemo enables AI chat demo features.
	ShowAIDemo bool `json:"showAiDemo"`
}

// DefaultDemoConfig returns sensible demo defaults.
func DefaultDemoConfig() DemoConfig {
	return DemoConfig{
		Enabled:     false,
		CompanyName: "Acme Sustainability Corp",
		Industry:    "Technology",
		BaseYear:    2022,
		ShowAIDemo:  true,
	}
}

// GenerateDemoData creates realistic demo data for presentations.
func GenerateDemoData(ctx context.Context, cfg DemoConfig) *DemoData {
	// Note: rand.Seed is deprecated in Go 1.20+, global rand is automatically seeded

	data := &DemoData{
		Organization: generateDemoOrg(cfg),
		Emissions:    generateDemoEmissions(cfg),
		Compliance:   generateDemoCompliance(),
		TrendData:    generateDemoTrends(cfg),
		Benchmarks:   generateDemoBenchmarks(),
	}

	return data
}

func generateDemoOrg(cfg DemoConfig) DemoOrganization {
	return DemoOrganization{
		ID:       uuid.NewString(),
		Name:     cfg.CompanyName,
		Industry: cfg.Industry,
		Size:     "Enterprise (5000+ employees)",
		Regions:  []string{"North America", "Europe", "Asia Pacific"},
		Logo:     "/demo/logo.svg",
	}
}

func generateDemoEmissions(cfg DemoConfig) []DemoEmission {
	emissions := make([]DemoEmission, 0, 100)
	facilities := []string{"HQ San Francisco", "Data Center Virginia", "Office London", "Factory Shanghai"}

	// Scope 1 - Direct emissions
	scope1Sources := []struct {
		source   string
		category string
		unit     string
		factor   float64
	}{
		{"Natural Gas", "Stationary Combustion", "therms", 0.0053},
		{"Fleet Diesel", "Mobile Combustion", "gallons", 0.0102},
		{"Fleet Gasoline", "Mobile Combustion", "gallons", 0.0089},
		{"Refrigerants", "Fugitive Emissions", "kg", 2.088},
	}

	for _, src := range scope1Sources {
		for _, facility := range facilities {
			qty := 10000 + rand.Float64()*50000
			emissions = append(emissions, DemoEmission{
				ID:            uuid.NewString(),
				Source:        src.source,
				Scope:         1,
				Category:      src.category,
				ActivityType:  src.source,
				QuantityValue: qty,
				QuantityUnit:  src.unit,
				EmissionsCO2e: qty * src.factor,
				Region:        getRegion(facility),
				Facility:      facility,
				Period:        "2024",
				DataQuality:   "High",
				Verified:      true,
				CreatedAt:     time.Now().Add(-time.Duration(rand.Intn(90)) * 24 * time.Hour),
			})
		}
	}

	// Scope 2 - Purchased energy
	scope2Sources := []struct {
		source string
		unit   string
		factor float64
	}{
		{"Electricity", "kWh", 0.0004},
		{"District Heating", "kWh", 0.0002},
		{"District Cooling", "kWh", 0.0001},
	}

	for _, src := range scope2Sources {
		for _, facility := range facilities {
			qty := 500000 + rand.Float64()*2000000
			emissions = append(emissions, DemoEmission{
				ID:            uuid.NewString(),
				Source:        src.source,
				Scope:         2,
				Category:      "Purchased Energy",
				ActivityType:  src.source,
				QuantityValue: qty,
				QuantityUnit:  src.unit,
				EmissionsCO2e: qty * src.factor,
				Region:        getRegion(facility),
				Facility:      facility,
				Period:        "2024",
				DataQuality:   "High",
				Verified:      true,
				CreatedAt:     time.Now().Add(-time.Duration(rand.Intn(60)) * 24 * time.Hour),
			})
		}
	}

	// Scope 3 - Value chain
	scope3Categories := []struct {
		category string
		name     string
		tons     float64
	}{
		{"Category 1", "Purchased Goods & Services", 5200},
		{"Category 2", "Capital Goods", 1800},
		{"Category 3", "Fuel & Energy Related", 420},
		{"Category 4", "Upstream Transportation", 890},
		{"Category 5", "Waste", 120},
		{"Category 6", "Business Travel", 340},
		{"Category 7", "Employee Commuting", 580},
		{"Category 11", "Use of Sold Products", 2100},
		{"Category 12", "End-of-Life Treatment", 280},
	}

	for _, cat := range scope3Categories {
		emissions = append(emissions, DemoEmission{
			ID:            uuid.NewString(),
			Source:        cat.name,
			Scope:         3,
			Category:      cat.category,
			ActivityType:  cat.name,
			QuantityValue: cat.tons,
			QuantityUnit:  "tCO2e",
			EmissionsCO2e: cat.tons * (0.9 + rand.Float64()*0.2),
			Region:        "Global",
			Facility:      "All Facilities",
			Period:        "2024",
			DataQuality:   "Medium",
			Verified:      false,
			CreatedAt:     time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
		})
	}

	return emissions
}

func generateDemoCompliance() DemoCompliance {
	return DemoCompliance{
		CSRD: DemoFrameworkStatus{
			Name:         "CSRD / ESRS E1",
			Readiness:    0.78,
			Status:       "in_progress",
			DueDate:      "2025-01-01",
			Requirements: 11,
			Completed:    8,
		},
		SEC: DemoFrameworkStatus{
			Name:         "SEC Climate Disclosure",
			Readiness:    0.45,
			Status:       "in_progress",
			DueDate:      "2025-12-31",
			Requirements: 8,
			Completed:    3,
		},
		CBAM: DemoFrameworkStatus{
			Name:         "EU CBAM",
			Readiness:    0.25,
			Status:       "not_started",
			DueDate:      "2026-01-01",
			Requirements: 6,
			Completed:    1,
		},
		California: DemoFrameworkStatus{
			Name:         "California SB 253/261",
			Readiness:    0.60,
			Status:       "in_progress",
			DueDate:      "2026-01-01",
			Requirements: 5,
			Completed:    3,
		},
		Overall: 0.52,
	}
}

func generateDemoTrends(cfg DemoConfig) []DemoTrendPoint {
	trends := make([]DemoTrendPoint, 0, 12)

	// Generate 3 years of quarterly data showing improvement
	baseScope1 := 1800.0
	baseScope2 := 3200.0
	baseScope3 := 12500.0
	baseRevenue := 250.0 // $M

	for year := cfg.BaseYear; year <= cfg.BaseYear+2; year++ {
		for q := 1; q <= 4; q++ {
			// Model emissions reduction over time
			yearOffset := float64(year - cfg.BaseYear)
			quarterOffset := float64(q-1) * 0.25
			reduction := 1.0 - (yearOffset*0.08 + quarterOffset*0.02)

			scope1 := baseScope1 * reduction * (0.95 + rand.Float64()*0.1)
			scope2 := baseScope2 * reduction * (0.93 + rand.Float64()*0.1)
			scope3 := baseScope3 * reduction * (0.97 + rand.Float64()*0.06)
			total := scope1 + scope2 + scope3

			// Revenue grows while emissions decrease
			revenue := baseRevenue * (1 + yearOffset*0.15 + quarterOffset*0.02)
			intensity := total / revenue

			trends = append(trends, DemoTrendPoint{
				Period:    time.Date(year, time.Month(q*3), 1, 0, 0, 0, 0, time.UTC).Format("2006-Q1"),
				Scope1:    scope1,
				Scope2:    scope2,
				Scope3:    scope3,
				Total:     total,
				Intensity: intensity,
			})
		}
	}

	return trends
}

func generateDemoBenchmarks() []DemoBenchmark {
	return []DemoBenchmark{
		{
			Category:    "Carbon Intensity",
			YourValue:   52.3,
			IndustryAvg: 68.5,
			BestInClass: 38.2,
			Unit:        "tCO2e/$M revenue",
			PercentDiff: -23.6,
		},
		{
			Category:    "Renewable Energy %",
			YourValue:   42.0,
			IndustryAvg: 28.0,
			BestInClass: 85.0,
			Unit:        "% of electricity",
			PercentDiff: 50.0,
		},
		{
			Category:    "Scope 3 Coverage",
			YourValue:   75.0,
			IndustryAvg: 45.0,
			BestInClass: 95.0,
			Unit:        "% of categories",
			PercentDiff: 66.7,
		},
		{
			Category:    "Data Quality Score",
			YourValue:   8.2,
			IndustryAvg: 6.5,
			BestInClass: 9.5,
			Unit:        "out of 10",
			PercentDiff: 26.2,
		},
		{
			Category:    "Compliance Readiness",
			YourValue:   78.0,
			IndustryAvg: 52.0,
			BestInClass: 95.0,
			Unit:        "%",
			PercentDiff: 50.0,
		},
	}
}

func getRegion(facility string) string {
	switch {
	case facility == "HQ San Francisco" || facility == "Data Center Virginia":
		return "US-WECC"
	case facility == "Office London":
		return "UK"
	case facility == "Factory Shanghai":
		return "CN"
	default:
		return "US-WECC"
	}
}

// SummaryStats returns aggregate statistics for the demo data.
func (d *DemoData) SummaryStats() map[string]interface{} {
	var scope1, scope2, scope3 float64
	for _, e := range d.Emissions {
		switch e.Scope {
		case 1:
			scope1 += e.EmissionsCO2e
		case 2:
			scope2 += e.EmissionsCO2e
		case 3:
			scope3 += e.EmissionsCO2e
		}
	}

	return map[string]interface{}{
		"organization": d.Organization.Name,
		"industry":     d.Organization.Industry,
		"totals": map[string]float64{
			"scope1": scope1,
			"scope2": scope2,
			"scope3": scope3,
			"total":  scope1 + scope2 + scope3,
		},
		"compliance":      d.Compliance.Overall * 100,
		"emissionRecords": len(d.Emissions),
		"trendDataPoints": len(d.TrendData),
		"benchmarks":      len(d.Benchmarks),
	}
}
