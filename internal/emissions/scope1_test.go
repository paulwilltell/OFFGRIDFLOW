// Package emissions provides Scope 1 (direct emissions) calculator tests.
package emissions

import (
	"context"
	"testing"
	"time"
)

// TestScope1Calculator_VehicleEmissions tests fleet/vehicle emissions calculation.
func TestScope1Calculator_VehicleEmissions(t *testing.T) {
	ctx := context.Background()

	// Create a simple in-memory factor registry for testing
	registry := NewInMemoryRegistry()

	// Register test emission factor for diesel fuel
	factor := EmissionFactor{
		ID:                 "test-diesel-2024",
		Scope:              Scope1,
		Region:             "US-AVERAGE",
		Source:             "fleet",
		Category:           "diesel",
		Unit:               "L",
		ValueKgCO2ePerUnit: 2.68, // kg CO2e per liter of diesel
		Method:             MethodLocationBased,
		DataSource:         "EPA 2024",
		ValidFrom:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:          time.Now(),
	}

	err := registry.RegisterFactor(ctx, factor)
	if err != nil {
		t.Fatalf("Failed to register factor: %v", err)
	}

	// Create Scope 1 calculator
	calc := NewScope1Calculator(Scope1Config{
		Registry: registry,
	})

	// Create test activity: 100 liters of diesel consumed
	activity := ActivityAdapter{
		ID:          "test-activity-1",
		Source:      "fleet",
		Category:    "diesel",
		Location:    "US-AVERAGE",
		Quantity:    100.0,
		Unit:        "L",
		PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		OrgID:       "org-123",
		WorkspaceID: "ws-456",
	}

	// Calculate emissions
	record, err := calc.Calculate(ctx, activity)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// Verify results
	expectedKg := 100.0 * 2.68 // 268 kg CO2e
	if record.EmissionsKgCO2e != expectedKg {
		t.Errorf("Expected %f kg CO2e, got %f", expectedKg, record.EmissionsKgCO2e)
	}

	expectedTonnes := expectedKg / 1000.0
	if record.EmissionsTonnesCO2e != expectedTonnes {
		t.Errorf("Expected %f tonnes CO2e, got %f", expectedTonnes, record.EmissionsTonnesCO2e)
	}

	if record.Scope != Scope1 {
		t.Errorf("Expected Scope 1, got %s", record.Scope)
	}

	if record.OrgID != "org-123" {
		t.Errorf("Expected org-123, got %s", record.OrgID)
	}
}

// TestScope1Calculator_StationaryCombustion tests facility natural gas emissions.
func TestScope1Calculator_StationaryCombustion(t *testing.T) {
	ctx := context.Background()
	registry := NewInMemoryRegistry()

	// Register emission factor for natural gas
	factor := EmissionFactor{
		ID:                 "test-natgas-2024",
		Scope:              Scope1,
		Region:             "US-AVERAGE",
		Source:             "stationary_combustion",
		Category:           "natural_gas",
		Unit:               "m3",
		ValueKgCO2ePerUnit: 1.93, // kg CO2e per cubic meter
		Method:             MethodLocationBased,
		DataSource:         "EPA 2024",
		ValidFrom:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:          time.Now(),
	}

	err := registry.RegisterFactor(ctx, factor)
	if err != nil {
		t.Fatalf("Failed to register factor: %v", err)
	}

	calc := NewScope1Calculator(Scope1Config{Registry: registry})

	// 1000 cubic meters of natural gas
	activity := ActivityAdapter{
		ID:          "test-activity-2",
		Source:      "stationary_combustion",
		Category:    "natural_gas",
		Location:    "US-AVERAGE",
		Quantity:    1000.0,
		Unit:        "m3",
		PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		OrgID:       "org-123",
	}

	record, err := calc.Calculate(ctx, activity)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	expectedKg := 1000.0 * 1.93 // 1930 kg CO2e
	if record.EmissionsKgCO2e != expectedKg {
		t.Errorf("Expected %f kg CO2e, got %f", expectedKg, record.EmissionsKgCO2e)
	}
}

// TestScope1Calculator_FugitiveEmissions tests refrigerant leaks.
func TestScope1Calculator_FugitiveEmissions(t *testing.T) {
	ctx := context.Background()
	registry := NewInMemoryRegistry()

	// R-410A refrigerant has very high GWP
	factor := EmissionFactor{
		ID:                 "test-r410a-2024",
		Scope:              Scope1,
		Region:             "GLOBAL",
		Source:             "refrigerants",
		Category:           "R-410A",
		Unit:               "kg",
		ValueKgCO2ePerUnit: 2088.0, // GWP of R-410A
		Method:             MethodLocationBased,
		DataSource:         "IPCC AR6",
		ValidFrom:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:          time.Now(),
	}

	err := registry.RegisterFactor(ctx, factor)
	if err != nil {
		t.Fatalf("Failed to register factor: %v", err)
	}

	calc := NewScope1Calculator(Scope1Config{Registry: registry})

	// 0.5 kg refrigerant leak
	activity := ActivityAdapter{
		ID:          "test-activity-3",
		Source:      "refrigerants",
		Category:    "R-410A",
		Location:    "GLOBAL",
		Quantity:    0.5,
		Unit:        "kg",
		PeriodStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2024, 6, 30, 23, 59, 59, 0, time.UTC),
		OrgID:       "org-123",
	}

	record, err := calc.Calculate(ctx, activity)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	expectedKg := 0.5 * 2088.0 // 1044 kg CO2e
	if record.EmissionsKgCO2e != expectedKg {
		t.Errorf("Expected %f kg CO2e, got %f", expectedKg, record.EmissionsKgCO2e)
	}
}

// TestScope1Calculator_Supports tests the calculator's activity support detection.
func TestScope1Calculator_Supports(t *testing.T) {
	registry := NewInMemoryRegistry()
	calc := NewScope1Calculator(Scope1Config{Registry: registry})

	testCases := []struct {
		name      string
		activity  Activity
		supported bool
	}{
		{
			name: "fleet vehicle - supported",
			activity: ActivityAdapter{
				Source: "fleet",
				Unit:   "L",
			},
			supported: true,
		},
		{
			name: "stationary combustion - supported",
			activity: ActivityAdapter{
				Source: "stationary_combustion",
				Unit:   "m3",
			},
			supported: true,
		},
		{
			name: "refrigerants - supported",
			activity: ActivityAdapter{
				Source: "refrigerants",
				Unit:   "kg",
			},
			supported: true,
		},
		{
			name: "on-site combustion - supported",
			activity: ActivityAdapter{
				Source: "on-site",
				Unit:   "L",
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
			name: "travel - NOT supported (Scope 3)",
			activity: ActivityAdapter{
				Source: "travel",
				Unit:   "km",
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

// TestScope1Calculator_BatchCalculation tests multiple activities at once.
func TestScope1Calculator_BatchCalculation(t *testing.T) {
	ctx := context.Background()
	registry := NewInMemoryRegistry()

	// Register multiple factors
	factors := []EmissionFactor{
		{
			ID:                 "diesel-2024",
			Scope:              Scope1,
			Region:             "US-AVERAGE",
			Source:             "fleet",
			Category:           "diesel",
			Unit:               "L",
			ValueKgCO2ePerUnit: 2.68,
			Method:             MethodLocationBased,
			DataSource:         "EPA 2024",
			ValidFrom:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:                 "gasoline-2024",
			Scope:              Scope1,
			Region:             "US-AVERAGE",
			Source:             "fleet",
			Category:           "gasoline",
			Unit:               "L",
			ValueKgCO2ePerUnit: 2.31,
			Method:             MethodLocationBased,
			DataSource:         "EPA 2024",
			ValidFrom:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, f := range factors {
		if err := registry.RegisterFactor(ctx, f); err != nil {
			t.Fatalf("Failed to register factor: %v", err)
		}
	}

	calc := NewScope1Calculator(Scope1Config{Registry: registry})

	// Multiple vehicle fuel activities
	activities := []Activity{
		ActivityAdapter{
			ID:       "diesel-1",
			Source:   "fleet",
			Category: "diesel",
			Location: "US-AVERAGE",
			Quantity: 50.0,
			Unit:     "L",
			OrgID:    "org-123",
		},
		ActivityAdapter{
			ID:       "gasoline-1",
			Source:   "fleet",
			Category: "gasoline",
			Location: "US-AVERAGE",
			Quantity: 100.0,
			Unit:     "L",
			OrgID:    "org-123",
		},
		ActivityAdapter{
			ID:       "diesel-2",
			Source:   "fleet",
			Category: "diesel",
			Location: "US-AVERAGE",
			Quantity: 75.0,
			Unit:     "L",
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

	// Verify total emissions
	totalKg := 0.0
	for _, r := range records {
		totalKg += r.EmissionsKgCO2e
	}

	expected := (50.0*2.68 + 100.0*2.31 + 75.0*2.68) // diesel + gasoline + diesel
	if totalKg != expected {
		t.Errorf("Expected total %f kg CO2e, got %f", expected, totalKg)
	}
}

// TestScope1Calculator_ErrorHandling tests error scenarios.
func TestScope1Calculator_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	registry := NewInMemoryRegistry()
	calc := NewScope1Calculator(Scope1Config{Registry: registry})

	t.Run("missing emission factor", func(t *testing.T) {
		activity := ActivityAdapter{
			ID:       "test-1",
			Source:   "fleet",
			Category: "diesel",
			Location: "US-AVERAGE",
			Quantity: 100.0,
			Unit:     "L",
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
			Source:   "electricity", // Scope 2, not Scope 1
			Category: "grid",
			Location: "US-AVERAGE",
			Quantity: 100.0,
			Unit:     "kWh",
			OrgID:    "org-123",
		}

		if calc.Supports(activity) {
			t.Error("Calculator should not support Scope 2 activities")
		}
	})
}
