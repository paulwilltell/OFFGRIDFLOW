// Package emissions/scope1 provides Scope 1 direct emissions calculations.
//
// Scope 1 emissions are direct greenhouse gas emissions from sources that are
// owned or controlled by the organization, including:
//   - Stationary combustion (boilers, furnaces, generators)
//   - Mobile combustion (company vehicles, fleet)
//   - Process emissions (chemical/physical processes)
//   - Fugitive emissions (refrigerants, methane leaks)
package emissions

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// =============================================================================
// Scope 1 Emission Factors
// =============================================================================

// FuelType identifies different combustion fuel types for Scope 1.
type FuelType string

const (
	FuelDiesel     FuelType = "diesel"
	FuelGasoline   FuelType = "gasoline"
	FuelNaturalGas FuelType = "natural_gas"
	FuelPropane    FuelType = "propane"
	FuelFuelOil    FuelType = "fuel_oil"
	FuelFuelOil1   FuelType = "fuel_oil_1"
	FuelFuelOil2   FuelType = "fuel_oil_2"
	FuelFuelOil4   FuelType = "fuel_oil_4"
	FuelFuelOil5   FuelType = "fuel_oil_5"
	FuelFuelOil6   FuelType = "fuel_oil_6"
	FuelBiodiesel  FuelType = "biodiesel"
	FuelEthanol    FuelType = "ethanol"
	FuelJetFuel    FuelType = "jet_fuel"
	FuelCoal       FuelType = "coal"
	FuelWood       FuelType = "wood"
)

// String returns the string representation of the fuel type.
func (f FuelType) String() string {
	return string(f)
}

// DefaultScope1Factors contains standard emission factors for common fuels.
// Values are in kg CO2e per unit (L for liquids, m³ for gas, kg for solids).
// Source: EPA GHG Emission Factors Hub, IPCC Guidelines.
var DefaultScope1Factors = map[FuelType]float64{
	FuelDiesel:     2.68, // kg CO2e per liter
	FuelGasoline:   2.31, // kg CO2e per liter
	FuelNaturalGas: 1.93, // kg CO2e per cubic meter
	FuelPropane:    1.51, // kg CO2e per liter
	FuelFuelOil:    2.96, // kg CO2e per liter (No. 2)
	FuelFuelOil1:   2.72,
	FuelFuelOil2:   2.96,
	FuelFuelOil4:   3.10,
	FuelFuelOil5:   3.20,
	FuelFuelOil6:   3.25,
	FuelBiodiesel:  0.00, // Considered carbon neutral (biogenic)
	FuelEthanol:    0.00, // Considered carbon neutral (biogenic)
	FuelJetFuel:    2.52, // kg CO2e per liter
	FuelCoal:       2.42, // kg CO2e per kg
	FuelWood:       0.00, // Biogenic carbon (excluded from Scope 1)
}

// defaultCH4N2OUplift adds CH4/N2O contributions when enabled (kg CO2e/unit).
var defaultCH4N2OUplift = map[FuelType]float64{
	FuelDiesel:     0.01,
	FuelGasoline:   0.012,
	FuelNaturalGas: 0.004,
	FuelPropane:    0.003,
	FuelFuelOil:    0.012,
	FuelFuelOil1:   0.010,
	FuelFuelOil2:   0.012,
	FuelFuelOil4:   0.014,
	FuelFuelOil5:   0.014,
	FuelFuelOil6:   0.015,
	FuelJetFuel:    0.010,
	FuelCoal:       0.020,
}

// =============================================================================
// Scope 1 Calculator Configuration
// =============================================================================

// Scope1Config configures the Scope 1 calculator.
type Scope1Config struct {
	// Registry provides emission factor lookup.
	Registry FactorRegistry

	// Logger for calculation events.
	Logger *slog.Logger

	// UseBiogenicCredits includes biogenic fuels in calculations.
	// When true, biofuels may have zero or negative emission factors.
	UseBiogenicCredits bool

	// IncludeCH4AndN2O includes methane and nitrous oxide in calculations.
	// When true, uses higher GWP-adjusted emission factors.
	IncludeCH4AndN2O bool
}

// =============================================================================
// Scope 1 Calculator Implementation
// =============================================================================

// Scope1Calculator computes Scope 1 direct emissions.
//
// Scope 1 includes:
//   - Fleet vehicles (company cars, trucks, ships)
//   - Stationary combustion (generators, boilers, furnaces)
//   - Fugitive emissions (refrigerant leaks, natural gas leaks)
//   - Process emissions (from manufacturing processes)
//
// The calculator supports activity-based calculations using fuel quantity
// and unit-specific emission factors.
type Scope1Calculator struct {
	registry FactorRegistry
	logger   *slog.Logger
	config   Scope1Config
}

// NewScope1Calculator creates a new Scope 1 emissions calculator.
func NewScope1Calculator(cfg Scope1Config) *Scope1Calculator {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &Scope1Calculator{
		registry: cfg.Registry,
		logger:   logger,
		config:   cfg,
	}
}

// Supports returns true if this calculator can handle the activity.
func (c *Scope1Calculator) Supports(activity Activity) bool {
	source := activity.GetSource()

	switch source {
	case "fleet", "vehicle", "on-site", "stationary_combustion",
		"mobile_combustion", "fugitive", "refrigerants", "process":
		return true
	default:
		return false
	}
}

// Calculate computes Scope 1 emissions for the given activity.
func (c *Scope1Calculator) Calculate(ctx context.Context, activity Activity) (EmissionRecord, error) {
	if activity == nil {
		return EmissionRecord{}, ErrNilActivity
	}

	c.logger.Debug("calculating scope 1 emissions",
		"activity_id", activity.GetID(),
		"source", activity.GetSource(),
		"quantity", activity.GetQuantity(),
		"unit", activity.GetUnit(),
	)

	// Find the appropriate emission factor
	factor, err := c.findFactor(ctx, activity)
	if err != nil {
		return EmissionRecord{}, fmt.Errorf("find scope 1 factor: %w", err)
	}

	// Calculate emissions
	quantity := activity.GetQuantity()
	emissionsKg := factor.CalculateEmissions(quantity)

	record := EmissionRecord{
		ID:                  GenerateRecordID(),
		ActivityID:          activity.GetID(),
		FactorID:            factor.ID,
		Scope:               Scope1,
		EmissionsKgCO2e:     emissionsKg,
		EmissionsTonnesCO2e: KgToTonnes(emissionsKg),
		InputQuantity:       quantity,
		InputUnit:           activity.GetUnit(),
		EmissionFactor:      factor.ValueKgCO2ePerUnit,
		Method:              MethodActivityBased,
		DataQuality:         DataQualityMeasured,
		Region:              activity.GetLocation(),
		OrgID:               activity.GetOrgID(),
		WorkspaceID:         activity.GetWorkspaceID(),
		PeriodStart:         activity.GetPeriodStart(),
		PeriodEnd:           activity.GetPeriodEnd(),
		CalculatedAt:        time.Now().UTC(),
	}

	c.logger.Info("calculated scope 1 emissions",
		"activity_id", activity.GetID(),
		"factor_id", factor.ID,
		"emissions_kg_co2e", emissionsKg,
	)

	return record, nil
}

// CalculateBatch processes multiple activities.
func (c *Scope1Calculator) CalculateBatch(ctx context.Context, activities []Activity) ([]EmissionRecord, error) {
	records := make([]EmissionRecord, 0, len(activities))

	for _, activity := range activities {
		if !c.Supports(activity) {
			continue
		}

		record, err := c.Calculate(ctx, activity)
		if err != nil {
			c.logger.Warn("failed to calculate scope 1 activity",
				"activity_id", activity.GetID(),
				"error", err,
			)
			continue
		}

		records = append(records, record)
	}

	return records, nil
}

// findFactor locates the best emission factor for the activity.
func (c *Scope1Calculator) findFactor(ctx context.Context, activity Activity) (EmissionFactor, error) {
	// First try the registry
	if c.registry != nil {
		query := FactorQuery{
			Scope:    Scope1,
			Region:   activity.GetLocation(),
			Source:   activity.GetSource(),
			Category: activity.GetCategory(),
			Unit:     activity.GetUnit(),
			ValidAt:  time.Now(),
		}

		factor, err := c.registry.FindFactor(ctx, query)
		if err == nil {
			return factor, nil
		}

		// With an explicit registry, treat missing factors as errors to avoid silent fallbacks.
		return EmissionFactor{}, err
	}

	// Fall back to default factors based on category/fuel type
	return c.getDefaultFactor(activity)
}

// getDefaultFactor returns a built-in factor based on fuel type.
func (c *Scope1Calculator) getDefaultFactor(activity Activity) (EmissionFactor, error) {
	category := activity.GetCategory()
	unit := activity.GetUnit()

	// Map category to fuel type
	var fuelType FuelType
	switch category {
	case "diesel", "Diesel":
		fuelType = FuelDiesel
	case "gasoline", "petrol", "Gasoline":
		fuelType = FuelGasoline
	case "natural_gas", "gas", "Natural Gas":
		fuelType = FuelNaturalGas
	case "propane", "lpg", "LPG":
		fuelType = FuelPropane
	case "fuel_oil", "heating_oil", "fuel_oil_2":
		fuelType = FuelFuelOil
	case "fuel_oil_1":
		fuelType = FuelFuelOil1
	case "fuel_oil_4":
		fuelType = FuelFuelOil4
	case "fuel_oil_5":
		fuelType = FuelFuelOil5
	case "fuel_oil_6":
		fuelType = FuelFuelOil6
	case "jet_fuel", "aviation":
		fuelType = FuelJetFuel
	default:
		// Try to infer from source if fleet
		if activity.GetSource() == "fleet" {
			fuelType = FuelDiesel // Default fleet fuel
		} else {
			return EmissionFactor{}, fmt.Errorf(
				"unknown fuel category %q: %w",
				category, ErrFactorNotFound,
			)
		}
	}

	factorValue, ok := DefaultScope1Factors[fuelType]
	if !ok {
		return EmissionFactor{}, fmt.Errorf(
			"no default factor for fuel type %s: %w",
			fuelType, ErrFactorNotFound,
		)
	}

	if c.config.IncludeCH4AndN2O {
		if extra, ok := defaultCH4N2OUplift[fuelType]; ok {
			factorValue += extra
		}
	}

	return EmissionFactor{
		ID:                 fmt.Sprintf("default-scope1-%s", fuelType),
		Scope:              Scope1,
		Region:             activity.GetLocation(),
		Source:             activity.GetSource(),
		Category:           string(fuelType),
		Unit:               unit,
		ValueKgCO2ePerUnit: factorValue,
		Method:             MethodActivityBased,
		DataSource:         "EPA GHG Emission Factors Hub (default)",
		CreatedAt:          time.Now().UTC(),
	}, nil
}

// =============================================================================
// Scope 1 Helpers
// =============================================================================

// CalculateFleetEmissions is a convenience function for fleet calculations.
func CalculateFleetEmissions(fuelType FuelType, liters float64) float64 {
	factor, ok := DefaultScope1Factors[fuelType]
	if !ok {
		return 0
	}
	return liters * factor
}

// CalculateStationaryCombustion calculates emissions from on-site fuel use.
func CalculateStationaryCombustion(fuelType FuelType, quantity float64) float64 {
	factor, ok := DefaultScope1Factors[fuelType]
	if !ok {
		return 0
	}
	return quantity * factor
}

// EstimateFugitiveEmissions calculates refrigerant leakage emissions.
// Uses GWP values for common refrigerants.
func EstimateFugitiveEmissions(refrigerantType string, leakageKg float64) float64 {
	// Global Warming Potentials for common refrigerants (100-year GWP)
	gwpValues := map[string]float64{
		"R-134a":  1430, // HFC-134a
		"R-410A":  2088, // HFC blend
		"R-404A":  3922, // HFC blend
		"R-407C":  1774, // HFC blend
		"R-22":    1810, // HCFC-22 (being phased out)
		"R-290":   3,    // Propane
		"R-744":   1,    // CO2
		"R-717":   0,    // Ammonia
		"default": 1500, // Conservative estimate
	}

	gwp, ok := gwpValues[refrigerantType]
	if !ok {
		gwp = gwpValues["default"]
	}

	// Emissions = leakage (kg) × GWP
	return leakageKg * gwp
}
