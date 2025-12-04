// Package emissions provides Scope 3 (value chain emissions) calculator tests.
package emissions

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestScope3Calculator_BusinessTravel tests employee air travel emissions.
func TestScope3Calculator_BusinessTravel(t *testing.T) {
	ctx := context.Background()
	registry := NewInMemoryRegistry()

	// Short-haul flight emission factor
	factor := EmissionFactor{
		ID:                 "test-flight-short-2024",
		Scope:              Scope3,
		Region:             "GLOBAL",
		Source:             "travel",
		Category:           "flight_short_haul",
		Unit:               "km",
		ValueKgCO2ePerUnit: 0.255, // kg CO2e per passenger-km
		Method:             MethodLocationBased,
		DataSource:         "DEFRA 2024",
		ValidFrom:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:          time.Now(),
	}

	err := registry.RegisterFactor(ctx, factor)
	if err != nil {
		t.Fatalf("Failed to register factor: %v", err)
	}

	calc := NewScope3Calculator(Scope3Config{Registry: registry})

	// 500 km business flight
	activity := ActivityAdapter{
		ID:          "test-travel-1",
		Source:      "travel",
		Category:    "flight_short_haul",
		Location:    "GLOBAL",
		Quantity:    500.0,
		Unit:        "km",
		PeriodStart: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2024, 3, 15, 23, 59, 59, 0, time.UTC),
		OrgID:       "org-123",
	}

	record, err := calc.Calculate(ctx, activity)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	expectedKg := 500.0 * 0.255 // 127.5 kg CO2e
	if record.EmissionsKgCO2e != expectedKg {
		t.Errorf("Expected %f kg CO2e, got %f", expectedKg, record.EmissionsKgCO2e)
	}

	if record.Scope != Scope3 {
		t.Errorf("Expected Scope 3, got %s", record.Scope)
	}
}

// TestScope3Calculator_EmployeeCommuting tests daily commute emissions.
func TestScope3Calculator_EmployeeCommuting(t *testing.T) {
	ctx := context.Background()
	registry := NewInMemoryRegistry()

	// Car commuting emission factor
	factor := EmissionFactor{
		ID:                 "test-commute-car-2024",
		Scope:              Scope3,
		Region:             "US-AVERAGE",
		Source:             "commuting",
		Category:           "car",
		Unit:               "km",
		ValueKgCO2ePerUnit: 0.192, // kg CO2e per km
		Method:             MethodLocationBased,
		DataSource:         "EPA 2024",
		ValidFrom:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:          time.Now(),
	}

	err := registry.RegisterFactor(ctx, factor)
	if err != nil {
		t.Fatalf("Failed to register factor: %v", err)
	}

	calc := NewScope3Calculator(Scope3Config{Registry: registry})

	// 20 km daily commute for month (20 working days)
	totalKm := 20.0 * 2 * 20 // 20 km each way * 20 days = 800 km
	activity := ActivityAdapter{
		ID:          "test-commute-1",
		Source:      "commuting",
		Category:    "car",
		Location:    "US-AVERAGE",
		Quantity:    totalKm,
		Unit:        "km",
		PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		OrgID:       "org-123",
	}

	record, err := calc.Calculate(ctx, activity)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	expectedKg := totalKm * 0.192 // 153.6 kg CO2e
	if record.EmissionsKgCO2e != expectedKg {
		t.Errorf("Expected %f kg CO2e, got %f", expectedKg, record.EmissionsKgCO2e)
	}
}

// TestScope3Calculator_PurchasedGoods tests upstream supply chain emissions.
func TestScope3Calculator_PurchasedGoods(t *testing.T) {
	ctx := context.Background()
	registry := NewInMemoryRegistry()

	// Steel purchasing emission factor (spend-based)
	factor := EmissionFactor{
		ID:                 "test-steel-purchase-2024",
		Scope:              Scope3,
		Region:             "US-AVERAGE",
		Source:             "purchases",
		Category:           "steel",
		Unit:               "USD",
		ValueKgCO2ePerUnit: 2.45, // kg CO2e per USD spent
		Method:             MethodLocationBased,
		DataSource:         "EPA EEIO 2024",
		ValidFrom:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:          time.Now(),
	}

	err := registry.RegisterFactor(ctx, factor)
	if err != nil {
		t.Fatalf("Failed to register factor: %v", err)
	}

	calc := NewScope3Calculator(Scope3Config{Registry: registry})

	// $10,000 steel purchase
	activity := ActivityAdapter{
		ID:          "test-purchase-1",
		Source:      "purchases",
		Category:    "steel",
		Location:    "US-AVERAGE",
		Quantity:    10000.0,
		Unit:        "USD",
		PeriodStart: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2024, 2, 29, 23, 59, 59, 0, time.UTC),
		OrgID:       "org-123",
	}

	record, err := calc.Calculate(ctx, activity)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	expectedKg := 10000.0 * 2.45 // 24,500 kg CO2e = 24.5 tonnes
	if record.EmissionsKgCO2e != expectedKg {
		t.Errorf("Expected %f kg CO2e, got %f", expectedKg, record.EmissionsKgCO2e)
	}
}

// TestScope3Calculator_WasteDisposal tests end-of-life emissions.
func TestScope3Calculator_WasteDisposal(t *testing.T) {
	ctx := context.Background()
	registry := NewInMemoryRegistry()

	// Landfill waste emission factor
	factor := EmissionFactor{
		ID:                 "test-waste-landfill-2024",
		Scope:              Scope3,
		Region:             "US-AVERAGE",
		Source:             "waste",
		Category:           "landfill",
		Unit:               "kg",
		ValueKgCO2ePerUnit: 0.412, // kg CO2e per kg waste
		Method:             MethodLocationBased,
		DataSource:         "EPA WARM 2024",
		ValidFrom:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:          time.Now(),
	}

	err := registry.RegisterFactor(ctx, factor)
	if err != nil {
		t.Fatalf("Failed to register factor: %v", err)
	}

	calc := NewScope3Calculator(Scope3Config{Registry: registry})

	// 1000 kg waste to landfill
	activity := ActivityAdapter{
		ID:          "test-waste-1",
		Source:      "waste",
		Category:    "landfill",
		Location:    "US-AVERAGE",
		Quantity:    1000.0,
		Unit:        "kg",
		PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		OrgID:       "org-123",
	}

	record, err := calc.Calculate(ctx, activity)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	expectedKg := 1000.0 * 0.412 // 412 kg CO2e
	if record.EmissionsKgCO2e != expectedKg {
		t.Errorf("Expected %f kg CO2e, got %f", expectedKg, record.EmissionsKgCO2e)
	}
}

// TestScope3Calculator_UpstreamTransportation tests logistics emissions.
func TestScope3Calculator_UpstreamTransportation(t *testing.T) {
	ctx := context.Background()
	registry := NewInMemoryRegistry()

	// Trucking emission factor (tonne-km)
	factor := EmissionFactor{
		ID:                 "test-truck-freight-2024",
		Scope:              Scope3,
		Region:             "US-AVERAGE",
		Source:             "upstream",
		Category:           "truck_freight",
		Unit:               "tonne-km",
		ValueKgCO2ePerUnit: 0.089, // kg CO2e per tonne-km
		Method:             MethodLocationBased,
		DataSource:         "GLEC Framework 2024",
		ValidFrom:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:          time.Now(),
	}

	err := registry.RegisterFactor(ctx, factor)
	if err != nil {
		t.Fatalf("Failed to register factor: %v", err)
	}

	calc := NewScope3Calculator(Scope3Config{Registry: registry})

	// 5 tonnes transported 200 km = 1000 tonne-km
	activity := ActivityAdapter{
		ID:          "test-freight-1",
		Source:      "upstream",
		Category:    "truck_freight",
		Location:    "US-AVERAGE",
		Quantity:    1000.0,
		Unit:        "tonne-km",
		PeriodStart: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC),
		OrgID:       "org-123",
	}

	record, err := calc.Calculate(ctx, activity)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	expectedKg := 1000.0 * 0.089 // 89 kg CO2e
	if record.EmissionsKgCO2e != expectedKg {
		t.Errorf("Expected %f kg CO2e, got %f", expectedKg, record.EmissionsKgCO2e)
	}
}

// TestScope3Calculator_Supports tests activity support detection.
func TestScope3Calculator_Supports(t *testing.T) {
	registry := NewInMemoryRegistry()
	calc := NewScope3Calculator(Scope3Config{Registry: registry})

	testCases := []struct {
		name      string
		activity  Activity
		supported bool
	}{
		{
			name: "business travel - supported",
			activity: ActivityAdapter{
				Source: "travel",
				Unit:   "km",
			},
			supported: true,
		},
		{
			name: "employee commuting - supported",
			activity: ActivityAdapter{
				Source: "commuting",
				Unit:   "km",
			},
			supported: true,
		},
		{
			name: "purchased goods - supported",
			activity: ActivityAdapter{
				Source: "purchases",
				Unit:   "USD",
			},
			supported: true,
		},
		{
			name: "waste disposal - supported",
			activity: ActivityAdapter{
				Source: "waste",
				Unit:   "kg",
			},
			supported: true,
		},
		{
			name: "upstream transportation - supported",
			activity: ActivityAdapter{
				Source: "upstream",
				Unit:   "tonne-km",
			},
			supported: true,
		},
		{
			name: "downstream transportation - supported",
			activity: ActivityAdapter{
				Source: "downstream",
				Unit:   "tonne-km",
			},
			supported: true,
		},
		{
			name: "electricity - NOT supported (Scope 2)",
			activity: ActivityAdapter{
				Source: "electricity",
				Unit:   "kWh",
			},
			supported: false,
		},
		{
			name: "fleet - NOT supported (Scope 1)",
			activity: ActivityAdapter{
				Source: "fleet",
				Unit:   "L",
			},
			supported: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calc.Supports(tc.activity)
			if result != tc.supported {
				t.Errorf("Expected Supports()=%v, got %v", tc.supported, result)
			}
		})
	}
}

// TestScope3Calculator_15Categories tests GHG Protocol's 15 Scope 3 categories.
func TestScope3Calculator_15Categories(t *testing.T) {
	ctx := context.Background()
	registry := NewInMemoryRegistry()

	// Register factors for all 15 categories
	categories := []struct {
		source   string
		category string
		unit     string
		factor   float64
	}{
		{"purchases", "purchased_goods_services", "USD", 1.5},            // Cat 1
		{"purchases", "capital_goods", "USD", 2.0},                       // Cat 2
		{"upstream", "fuel_energy_activities", "kWh", 0.05},              // Cat 3
		{"upstream", "transportation_distribution", "tonne-km", 0.089},   // Cat 4
		{"waste", "waste_operations", "kg", 0.412},                       // Cat 5
		{"travel", "business_travel", "km", 0.255},                       // Cat 6
		{"commuting", "employee_commuting", "km", 0.192},                 // Cat 7
		{"upstream", "leased_assets", "m2", 12.0},                        // Cat 8
		{"downstream", "transportation_distribution", "tonne-km", 0.089}, // Cat 9
		{"downstream", "processing_sold_products", "USD", 1.2},           // Cat 10
		{"downstream", "use_sold_products", "kWh", 0.45},                 // Cat 11
		{"downstream", "end_of_life_treatment", "kg", 0.35},              // Cat 12
		{"downstream", "leased_assets", "m2", 10.0},                      // Cat 13
		{"downstream", "franchises", "USD", 0.8},                         // Cat 14
		{"downstream", "investments", "USD", 1.1},                        // Cat 15
	}

	for i, cat := range categories {
		factor := EmissionFactor{
			ID:                 fmt.Sprintf("cat-%d-2024", i+1),
			Scope:              Scope3,
			Region:             "GLOBAL",
			Source:             cat.source,
			Category:           cat.category,
			Unit:               cat.unit,
			ValueKgCO2ePerUnit: cat.factor,
			Method:             MethodLocationBased,
			DataSource:         "GHG Protocol 2024",
			ValidFrom:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		if err := registry.RegisterFactor(ctx, factor); err != nil {
			t.Fatalf("Failed to register category %d: %v", i+1, err)
		}
	}

	calc := NewScope3Calculator(Scope3Config{Registry: registry})

	// Test that calculator supports all categories
	for i, cat := range categories {
		activity := ActivityAdapter{
			ID:       fmt.Sprintf("cat-%d-test", i+1),
			Source:   cat.source,
			Category: cat.category,
			Location: "GLOBAL",
			Quantity: 100.0,
			Unit:     cat.unit,
			OrgID:    "org-123",
		}

		if !calc.Supports(activity) {
			t.Errorf("Calculator should support category %d (%s/%s)", i+1, cat.source, cat.category)
		}

		record, err := calc.Calculate(ctx, activity)
		if err != nil {
			t.Errorf("Failed to calculate category %d: %v", i+1, err)
		}

		expected := 100.0 * cat.factor
		if record.EmissionsKgCO2e != expected {
			t.Errorf("Category %d: expected %f kg CO2e, got %f", i+1, expected, record.EmissionsKgCO2e)
		}
	}
}

// TestScope3Calculator_BatchCalculation tests multiple Scope 3 activities.
func TestScope3Calculator_BatchCalculation(t *testing.T) {
	ctx := context.Background()
	registry := NewInMemoryRegistry()

	// Register multiple Scope 3 factors
	factors := []EmissionFactor{
		{
			ID:                 "travel-2024",
			Scope:              Scope3,
			Region:             "GLOBAL",
			Source:             "travel",
			Category:           "flight_short_haul",
			Unit:               "km",
			ValueKgCO2ePerUnit: 0.255,
			Method:             MethodLocationBased,
			DataSource:         "DEFRA 2024",
			ValidFrom:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:                 "commute-2024",
			Scope:              Scope3,
			Region:             "US-AVERAGE",
			Source:             "commuting",
			Category:           "car",
			Unit:               "km",
			ValueKgCO2ePerUnit: 0.192,
			Method:             MethodLocationBased,
			DataSource:         "EPA 2024",
			ValidFrom:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:                 "waste-2024",
			Scope:              Scope3,
			Region:             "US-AVERAGE",
			Source:             "waste",
			Category:           "landfill",
			Unit:               "kg",
			ValueKgCO2ePerUnit: 0.412,
			Method:             MethodLocationBased,
			DataSource:         "EPA WARM 2024",
			ValidFrom:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, f := range factors {
		if err := registry.RegisterFactor(ctx, f); err != nil {
			t.Fatalf("Failed to register factor: %v", err)
		}
	}

	calc := NewScope3Calculator(Scope3Config{Registry: registry})

	activities := []Activity{
		ActivityAdapter{
			ID:       "travel-1",
			Source:   "travel",
			Category: "flight_short_haul",
			Location: "GLOBAL",
			Quantity: 500.0,
			Unit:     "km",
			OrgID:    "org-123",
		},
		ActivityAdapter{
			ID:       "commute-1",
			Source:   "commuting",
			Category: "car",
			Location: "US-AVERAGE",
			Quantity: 800.0,
			Unit:     "km",
			OrgID:    "org-123",
		},
		ActivityAdapter{
			ID:       "waste-1",
			Source:   "waste",
			Category: "landfill",
			Location: "US-AVERAGE",
			Quantity: 1000.0,
			Unit:     "kg",
			OrgID:    "org-123",
		},
	}

	records, err := calc.CalculateBatch(ctx, activities)
	if err != nil {
		t.Fatalf("CalculateBatch failed: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(records))
	}

	// Verify total
	totalKg := 0.0
	for _, r := range records {
		totalKg += r.EmissionsKgCO2e
	}

	expected := (500.0*0.255 + 800.0*0.192 + 1000.0*0.412)
	if totalKg != expected {
		t.Errorf("Expected total %f kg CO2e, got %f", expected, totalKg)
	}
}

// TestScope3Calculator_ErrorHandling tests error scenarios.
func TestScope3Calculator_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	registry := NewInMemoryRegistry()
	calc := NewScope3Calculator(Scope3Config{Registry: registry})

	t.Run("missing emission factor", func(t *testing.T) {
		activity := ActivityAdapter{
			ID:       "test-1",
			Source:   "travel",
			Category: "flight_long_haul",
			Location: "GLOBAL",
			Quantity: 5000.0,
			Unit:     "km",
			OrgID:    "org-123",
		}

		_, err := calc.Calculate(ctx, activity)
		if err == nil {
			t.Error("Expected error for missing emission factor")
		}

		if !IsNotFoundError(err) {
			t.Errorf("Expected ErrFactorNotFound, got: %v", err)
		}
	})

	t.Run("nil activity", func(t *testing.T) {
		_, err := calc.Calculate(ctx, nil)
		if err == nil {
			t.Error("Expected error for nil activity")
		}
	})

	t.Run("unsupported source", func(t *testing.T) {
		activity := ActivityAdapter{
			ID:       "test-2",
			Source:   "fleet", // Scope 1, not Scope 3
			Category: "diesel",
			Location: "US-AVERAGE",
			Quantity: 100.0,
			Unit:     "L",
			OrgID:    "org-123",
		}

		if calc.Supports(activity) {
			t.Error("Calculator should not support Scope 1 activities")
		}
	})
}

