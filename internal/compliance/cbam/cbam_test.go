package cbam

import (
	"context"
	"testing"
	"time"

	"github.com/example/offgridflow/internal/compliance/core"
)

// =============================================================================
// Model Tests
// =============================================================================

func TestCommodityType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		commodity CommodityType
		want     bool
	}{
		{"cement valid", CommodityCement, true},
		{"electricity valid", CommodityElectricity, true},
		{"fertilizer valid", CommodityFertilizer, true},
		{"iron_steel valid", CommodityIronSteel, true},
		{"aluminum valid", CommodityAluminum, true},
		{"hydrogen valid", CommodityHydrogen, true},
		{"invalid", CommodityType("invalid"), false},
		{"empty", CommodityType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.commodity.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetQuarter(t *testing.T) {
	tests := []struct {
		name    string
		date    time.Time
		want    int
	}{
		{"Q1", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), 1},
		{"Q2", time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC), 2},
		{"Q3", time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC), 3},
		{"Q4", time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC), 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := GetQuarter(tt.date)
			if q.Quarter != tt.want {
				t.Errorf("GetQuarter() quarter = %v, want %v", q.Quarter, tt.want)
			}
			if q.Year != tt.date.Year() {
				t.Errorf("GetQuarter() year = %v, want %v", q.Year, tt.date.Year())
			}
		})
	}
}

func TestFormatQuarter(t *testing.T) {
	tests := []struct {
		year    int
		quarter int
		want    string
	}{
		{2024, 1, "Q1 2024"},
		{2024, 4, "Q4 2024"},
		{2025, 2, "Q2 2025"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := FormatQuarter(tt.year, tt.quarter); got != tt.want {
				t.Errorf("FormatQuarter() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// Calculator Tests
// =============================================================================

func TestNewCalculator(t *testing.T) {
	calc := NewCalculator()
	if calc == nil {
		t.Fatal("NewCalculator returned nil")
	}

	if calc.defaultValues == nil {
		t.Error("defaultValues not initialized")
	}

	if calc.electricityFactors == nil {
		t.Error("electricityFactors not initialized")
	}
}

func TestCalculator_CalculateEmbeddedEmissions_DefaultValues(t *testing.T) {
	calc := NewCalculator()
	ctx := context.Background()

	good := &ImportedGood{
		GoodID:          "CEMENT-001",
		CommodityType:   CommodityCement,
		Quantity:        100.0,
		Unit:            "tonnes",
		CountryOfOrigin: "CN",
	}

	emissions, err := calc.CalculateEmbeddedEmissions(ctx, good)
	if err != nil {
		t.Fatalf("CalculateEmbeddedEmissions failed: %v", err)
	}

	if emissions == nil {
		t.Fatal("emissions is nil")
	}

	if emissions.TotalSpecificEmissions <= 0 {
		t.Error("TotalSpecificEmissions should be positive")
	}

	if emissions.CalculationMethod != "default_values" {
		t.Errorf("expected calculation method 'default_values', got %s", emissions.CalculationMethod)
	}

	if emissions.DataQuality != "default" {
		t.Errorf("expected data quality 'default', got %s", emissions.DataQuality)
	}
}

func TestCalculator_CalculateEmbeddedEmissions_AllCommodities(t *testing.T) {
	calc := NewCalculator()
	ctx := context.Background()

	commodities := []CommodityType{
		CommodityCement,
		CommodityIronSteel,
		CommodityAluminum,
		CommodityFertilizer,
		CommodityHydrogen,
		CommodityElectricity,
	}

	for _, commodity := range commodities {
		t.Run(string(commodity), func(t *testing.T) {
			good := &ImportedGood{
				GoodID:          "TEST-001",
				CommodityType:   commodity,
				Quantity:        100.0,
				Unit:            "tonnes",
				CountryOfOrigin: "DE",
			}

			emissions, err := calc.CalculateEmbeddedEmissions(ctx, good)
			if err != nil {
				t.Fatalf("CalculateEmbeddedEmissions failed: %v", err)
			}

			if emissions.TotalSpecificEmissions <= 0 {
				t.Errorf("%s: emissions should be positive", commodity)
			}
		})
	}
}

func TestCalculator_GetSummary(t *testing.T) {
	calc := NewCalculator()
	ctx := context.Background()

	goods := []ImportedGood{
		{
			GoodID:          "CEMENT-001",
			CommodityType:   CommodityCement,
			Quantity:        100.0,
			Unit:            "tonnes",
			CountryOfOrigin: "CN",
		},
		{
			GoodID:          "STEEL-001",
			CommodityType:   CommodityIronSteel,
			Quantity:        50.0,
			Unit:            "tonnes",
			CountryOfOrigin: "DE",
		},
	}

	// Calculate emissions for all goods
	for i := range goods {
		emissions, err := calc.CalculateEmbeddedEmissions(ctx, &goods[i])
		if err != nil {
			t.Fatalf("CalculateEmbeddedEmissions failed: %v", err)
		}
		goods[i].EmbeddedEmissions = *emissions
	}

	summary := calc.GetSummary(goods)

	if summary.TotalGoods != 2 {
		t.Errorf("expected 2 total goods, got %d", summary.TotalGoods)
	}

	if summary.SuccessfulCalcs != 2 {
		t.Errorf("expected 2 successful calcs, got %d", summary.SuccessfulCalcs)
	}

	if summary.TotalEmissions <= 0 {
		t.Error("total emissions should be positive")
	}

	if summary.UsingDefaultValues != 2 {
		t.Errorf("expected 2 using default values, got %d", summary.UsingDefaultValues)
	}
}

// =============================================================================
// Mapper Tests
// =============================================================================

func TestNewMapper(t *testing.T) {
	mapper := NewMapper()
	if mapper == nil {
		t.Fatal("NewMapper returned nil")
	}

	if mapper.calculator == nil {
		t.Error("calculator not initialized")
	}
}

func TestMapper_BuildReport(t *testing.T) {
	mapper := NewMapper()
	ctx := context.Background()

	input := &CBAMInput{
		DeclarantID:      "IMPORTER-001",
		DeclarantName:    "Test Importer Ltd",
		DeclarantEORI:    "GB123456789000",
		DeclarantCountry: "GB",
		QuarterYear:      2024,
		Quarter:          1,
		Goods: []ImportedGood{
			{
				GoodID:          "CEMENT-001",
				CommodityType:   CommodityCement,
				CNCode:          "25231000",
				Description:     "Portland Cement",
				Quantity:        100.0,
				Unit:            "tonnes",
				CountryOfOrigin: "CN",
				CustomsValue:    5000.0,
				ImportDate:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	coreInput := core.ComplianceInput{
		Year: 2024,
		Data: input,
	}

	report, err := mapper.BuildReport(ctx, coreInput)
	if err != nil {
		t.Fatalf("BuildReport failed: %v", err)
	}

	if report.Standard != "EU CBAM (Carbon Border Adjustment Mechanism)" {
		t.Errorf("unexpected standard: %s", report.Standard)
	}

	if report.Content == nil {
		t.Fatal("Content is nil")
	}

	// Check declaration
	declData, ok := report.Content["declaration"]
	if !ok {
		t.Fatal("declaration not in content")
	}

	declaration, ok := declData.(*CBAMDeclaration)
	if !ok {
		t.Fatal("declaration is not *CBAMDeclaration")
	}

	if declaration.DeclarantID != "IMPORTER-001" {
		t.Errorf("expected declarant ID 'IMPORTER-001', got %s", declaration.DeclarantID)
	}

	if declaration.Quarter != 1 {
		t.Errorf("expected quarter 1, got %d", declaration.Quarter)
	}

	if declaration.Year != 2024 {
		t.Errorf("expected year 2024, got %d", declaration.Year)
	}

	if len(declaration.GoodsEntries) != 1 {
		t.Fatalf("expected 1 goods entry, got %d", len(declaration.GoodsEntries))
	}

	if declaration.TotalEmbeddedEmissions <= 0 {
		t.Error("total embedded emissions should be positive")
	}

	if declaration.EstimatedCBAMCost <= 0 {
		t.Error("estimated CBAM cost should be positive")
	}
}

func TestMapper_BuildReport_MultipleGoods(t *testing.T) {
	mapper := NewMapper()
	ctx := context.Background()

	input := &CBAMInput{
		DeclarantID:      "IMPORTER-001",
		DeclarantName:    "Test Importer Ltd",
		DeclarantEORI:    "GB123456789000",
		DeclarantCountry: "GB",
		QuarterYear:      2024,
		Quarter:          2,
		Goods: []ImportedGood{
			{
				GoodID:          "CEMENT-001",
				CommodityType:   CommodityCement,
				Quantity:        100.0,
				Unit:            "tonnes",
				CountryOfOrigin: "CN",
				CustomsValue:    5000.0,
			},
			{
				GoodID:          "STEEL-001",
				CommodityType:   CommodityIronSteel,
				Quantity:        50.0,
				Unit:            "tonnes",
				CountryOfOrigin: "DE",
				CustomsValue:    8000.0,
			},
			{
				GoodID:          "ALUMINUM-001",
				CommodityType:   CommodityAluminum,
				Quantity:        25.0,
				Unit:            "tonnes",
				CountryOfOrigin: "RU",
				CustomsValue:    12000.0,
			},
		},
	}

	coreInput := core.ComplianceInput{
		Year: 2024,
		Data: input,
	}

	report, err := mapper.BuildReport(ctx, coreInput)
	if err != nil {
		t.Fatalf("BuildReport failed: %v", err)
	}

	declaration, ok := report.Content["declaration"].(*CBAMDeclaration)
	if !ok {
		t.Fatal("declaration not found or wrong type")
	}

	if declaration.TotalGoods != 3 {
		t.Errorf("expected 3 goods, got %d", declaration.TotalGoods)
	}

	if len(declaration.GoodsEntries) != 3 {
		t.Errorf("expected 3 goods entries, got %d", len(declaration.GoodsEntries))
	}

	expectedCustomsValue := 5000.0 + 8000.0 + 12000.0
	if declaration.TotalCustomsValue != expectedCustomsValue {
		t.Errorf("expected total customs value %.2f, got %.2f",
			expectedCustomsValue, declaration.TotalCustomsValue)
	}
}

// =============================================================================
// Validator Tests
// =============================================================================

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Fatal("NewValidator returned nil")
	}
}

func TestValidator_ValidateInput(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		input     *CBAMInput
		wantValid bool
		wantErrors int
	}{
		{
			name: "valid input",
			input: &CBAMInput{
				DeclarantID:   "IMPORTER-001",
				DeclarantEORI: "GB123456789000",
				Quarter:       1,
				QuarterYear:   2024,
				Goods: []ImportedGood{
					{
						GoodID:        "TEST-001",
						CommodityType: CommodityCement,
						Quantity:      100.0,
					},
				},
			},
			wantValid: true,
			wantErrors: 0,
		},
		{
			name: "missing declarant ID",
			input: &CBAMInput{
				DeclarantEORI: "GB123456789000",
				Quarter:       1,
				Goods: []ImportedGood{
					{GoodID: "TEST-001", CommodityType: CommodityCement, Quantity: 100.0},
				},
			},
			wantValid: false,
			wantErrors: 1,
		},
		{
			name: "missing EORI",
			input: &CBAMInput{
				DeclarantID: "IMPORTER-001",
				Quarter:     1,
				Goods: []ImportedGood{
					{GoodID: "TEST-001", CommodityType: CommodityCement, Quantity: 100.0},
				},
			},
			wantValid: false,
			wantErrors: 1,
		},
		{
			name: "invalid quarter",
			input: &CBAMInput{
				DeclarantID:   "IMPORTER-001",
				DeclarantEORI: "GB123456789000",
				Quarter:       5, // Invalid
				Goods: []ImportedGood{
					{GoodID: "TEST-001", CommodityType: CommodityCement, Quantity: 100.0},
				},
			},
			wantValid: false,
			wantErrors: 1,
		},
		{
			name: "no goods",
			input: &CBAMInput{
				DeclarantID:   "IMPORTER-001",
				DeclarantEORI: "GB123456789000",
				Quarter:       1,
				Goods:         []ImportedGood{},
			},
			wantValid: false,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := validator.ValidateInput(tt.input)

			if results.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", results.Valid, tt.wantValid)
			}

			if len(results.Errors) != tt.wantErrors {
				t.Errorf("got %d errors, want %d", len(results.Errors), tt.wantErrors)
				for _, err := range results.Errors {
					t.Logf("  Error: %s - %s", err.Field, err.Message)
				}
			}
		})
	}
}

func TestValidator_ValidateDeclaration(t *testing.T) {
	validator := NewValidator()

	validDeclaration := &CBAMDeclaration{
		DeclarationID:      "CBAM-001-Q1-2024",
		DeclarantID:        "IMPORTER-001",
		DeclarantName:      "Test Importer",
		DeclarantEORI:      "GB123456789000",
		Quarter:            1,
		Year:               2024,
		SubmissionDate:     time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC),
		DeclarationStatus:  "draft",
		TotalGoods:         1,
		GoodsEntries: []GoodEntry{
			{
				EntryID:         "ENTRY-001",
				CommodityType:   CommodityCement,
				Quantity:        100.0,
				TotalEmissions:  80.7,
				DirectEmissions: 76.6,
				IndirectEmissions: 4.1,
			},
		},
		TotalEmbeddedEmissions: 80.7,
		EstimatedCBAMPrice:     80.0,
		EstimatedCBAMCost:      6456.0,
	}

	results := validator.ValidateDeclaration(validDeclaration)

	if !results.Valid {
		t.Error("expected valid declaration")
		for _, err := range results.Errors {
			t.Logf("  Error: %s - %s", err.Field, err.Message)
		}
	}
}

func TestValidator_ValidateDeclaration_InvalidQuarter(t *testing.T) {
	validator := NewValidator()

	declaration := &CBAMDeclaration{
		DeclarationID:  "CBAM-001",
		DeclarantID:    "IMPORTER-001",
		DeclarantEORI:  "GB123456789000",
		Quarter:        5, // Invalid
		Year:           2024,
		TotalGoods:     1,
		GoodsEntries: []GoodEntry{
			{EntryID: "001", CommodityType: CommodityCement, Quantity: 100.0},
		},
	}

	results := validator.ValidateDeclaration(declaration)

	if results.Valid {
		t.Error("expected invalid declaration")
	}

	found := false
	for _, err := range results.Errors {
		if err.Field == "quarter" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected error for invalid quarter")
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestFullWorkflow(t *testing.T) {
	// Create components
	mapper := NewMapper()
	validator := NewValidator()
	ctx := context.Background()

	// Create input
	input := &CBAMInput{
		DeclarantID:      "ACME-IMPORTS",
		DeclarantName:    "ACME Imports GmbH",
		DeclarantEORI:    "DE123456789000",
		DeclarantCountry: "DE",
		QuarterYear:      2024,
		Quarter:          1,
		Goods: []ImportedGood{
			{
				GoodID:          "CEMENT-001",
				CommodityType:   CommodityCement,
				CNCode:          "25231000",
				Description:     "Portland Cement",
				Quantity:        500.0,
				Unit:            "tonnes",
				CountryOfOrigin: "CN",
				CustomsValue:    25000.0,
				ImportDate:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			},
			{
				GoodID:          "STEEL-001",
				CommodityType:   CommodityIronSteel,
				CNCode:          "72071100",
				Description:     "Hot-rolled steel coils",
				Quantity:        200.0,
				Unit:            "tonnes",
				CountryOfOrigin: "RU",
				CustomsValue:    100000.0,
				ImportDate:      time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	// Validate input
	inputValidation := validator.ValidateInput(input)
	if !inputValidation.Valid {
		t.Fatal("input validation failed")
	}

	// Generate report
	coreInput := core.ComplianceInput{
		Year: 2024,
		Data: input,
	}

	report, err := mapper.BuildReport(ctx, coreInput)
	if err != nil {
		t.Fatalf("BuildReport failed: %v", err)
	}

	// Verify report
	declaration, ok := report.Content["declaration"].(*CBAMDeclaration)
	if !ok {
		t.Fatal("declaration not found")
	}

	// Validate declaration
	declValidation := validator.ValidateDeclaration(declaration)
	if !declValidation.Valid {
		t.Error("declaration validation failed")
		for _, err := range declValidation.Errors {
			t.Logf("  Error: %s - %s", err.Field, err.Message)
		}
	}

	// Verify key metrics
	if declaration.TotalGoods != 2 {
		t.Errorf("expected 2 goods, got %d", declaration.TotalGoods)
	}

	if declaration.TotalEmbeddedEmissions <= 0 {
		t.Error("total emissions should be positive")
	}

	if declaration.EstimatedCBAMCost <= 0 {
		t.Error("estimated cost should be positive")
	}

	// Check summary
	summary, ok := report.Content["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("summary not found")
	}

	if summary["total_goods"] != 2 {
		t.Errorf("summary: expected 2 goods, got %v", summary["total_goods"])
	}

	t.Logf("✓ Successfully processed CBAM declaration")
	t.Logf("  - Declarant: %s", declaration.DeclarantName)
	t.Logf("  - Period: Q%d %d", declaration.Quarter, declaration.Year)
	t.Logf("  - Total goods: %d", declaration.TotalGoods)
	t.Logf("  - Total emissions: %.2f tCO2e", declaration.TotalEmbeddedEmissions)
	t.Logf("  - Estimated CBAM cost: €%.2f", declaration.EstimatedCBAMCost)
}
