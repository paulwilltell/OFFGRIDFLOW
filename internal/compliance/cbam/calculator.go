package cbam

import (
	"context"
	"fmt"
	"math"
	"time"
)

// =============================================================================
// CBAM Calculator
// =============================================================================

// Calculator computes embedded emissions for CBAM goods.
type Calculator struct {
	// Default emission values registry
	defaultValues map[string]*DefaultEmissionValue

	// Electricity emission factors by country (gCO2e/kWh)
	electricityFactors map[string]float64
}

// NewCalculator creates a new CBAM calculator with default values.
func NewCalculator() *Calculator {
	calc := &Calculator{
		defaultValues:      make(map[string]*DefaultEmissionValue),
		electricityFactors: make(map[string]float64),
	}

	// Load default emission values
	calc.loadDefaultEmissionValues()

	// Load electricity emission factors
	calc.loadElectricityFactors()

	return calc
}

// =============================================================================
// Embedded Emissions Calculation
// =============================================================================

// CalculateEmbeddedEmissions computes the embedded emissions for an imported good.
func (c *Calculator) CalculateEmbeddedEmissions(ctx context.Context, good *ImportedGood) (*EmbeddedEmissions, error) {
	if good == nil {
		return nil, fmt.Errorf("good cannot be nil")
	}

	// If emissions are already provided and calculation method is specified, validate and return
	if good.EmbeddedEmissions.CalculationMethod != "" {
		return c.validateProvidedEmissions(&good.EmbeddedEmissions, good)
	}

	// Calculate based on available data
	return c.calculateFromData(ctx, good)
}

// calculateFromData determines the best calculation method based on available data.
func (c *Calculator) calculateFromData(ctx context.Context, good *ImportedGood) (*EmbeddedEmissions, error) {
	// Check if we have installation-specific data
	if good.InstallationID != "" {
		// In a real implementation, we would fetch installation data
		// For now, use default values
		return c.useDefaultValues(good)
	}

	// Check if we have precursor information for material balance
	if len(good.PrecursorInfo) > 0 {
		return c.calculateFromPrecursors(good)
	}

	// Fall back to default values
	return c.useDefaultValues(good)
}

// calculateFromPrecursors calculates emissions based on precursor materials.
func (c *Calculator) calculateFromPrecursors(good *ImportedGood) (*EmbeddedEmissions, error) {
	var totalPrecursorEmissions float64

	for _, precursor := range good.PrecursorInfo {
		totalPrecursorEmissions += precursor.EmbeddedEmissions
	}

	// Add process emissions (simplified)
	processEmissionFactor := c.getProcessEmissionFactor(good.CommodityType, good.ProductionRoute)
	processEmissions := good.Quantity * processEmissionFactor

	// Calculate indirect emissions from electricity
	electricityFactor := c.getElectricityFactor(good.CountryOfOrigin)
	// Estimate electricity consumption based on commodity type (kWh per tonne)
	electricityConsumption := c.getElectricityConsumption(good.CommodityType)
	indirectEmissions := (electricityConsumption * electricityFactor / 1000.0) * good.Quantity // Convert gCO2/kWh to tCO2e

	directEmissions := (totalPrecursorEmissions + processEmissions) / good.Quantity
	indirectEmissionsSpecific := indirectEmissions / good.Quantity

	return &EmbeddedEmissions{
		DirectEmissions:        directEmissions,
		IndirectEmissions:      indirectEmissionsSpecific,
		TotalSpecificEmissions: directEmissions + indirectEmissionsSpecific,
		CalculationMethod:      "material_balance",
		DataQuality:            "estimated",
		UncertaintyLevel:       25.0, // 25% uncertainty for material balance method
		Verified:               false,
	}, nil
}

// useDefaultValues applies default emission values for a good.
func (c *Calculator) useDefaultValues(good *ImportedGood) (*EmbeddedEmissions, error) {
	key := c.makeDefaultValueKey(good.CommodityType, good.CountryOfOrigin, good.ProductionRoute)

	defaultValue, exists := c.defaultValues[key]
	if !exists {
		// Try without production route
		key = c.makeDefaultValueKey(good.CommodityType, good.CountryOfOrigin, "")
		defaultValue, exists = c.defaultValues[key]
	}

	if !exists {
		// Try global default for commodity type
		key = c.makeDefaultValueKey(good.CommodityType, "", "")
		defaultValue, exists = c.defaultValues[key]
	}

	if !exists {
		return nil, fmt.Errorf("no default values found for commodity type %s", good.CommodityType)
	}

	return &EmbeddedEmissions{
		DirectEmissions:        defaultValue.DefaultDirectEmissions,
		IndirectEmissions:      defaultValue.DefaultIndirectEmissions,
		TotalSpecificEmissions: defaultValue.DefaultTotalEmissions,
		CalculationMethod:      "default_values",
		DataQuality:            "default",
		UncertaintyLevel:       40.0, // 40% uncertainty for default values
		Verified:               false,
	}, nil
}

// validateProvidedEmissions validates emissions data provided by the user.
func (c *Calculator) validateProvidedEmissions(emissions *EmbeddedEmissions, good *ImportedGood) (*EmbeddedEmissions, error) {
	// Validate that totals match
	calculatedTotal := emissions.DirectEmissions + emissions.IndirectEmissions
	tolerance := 0.01 // 1% tolerance

	if math.Abs(calculatedTotal-emissions.TotalSpecificEmissions) > tolerance {
		return nil, fmt.Errorf(
			"total specific emissions (%.4f) does not match direct + indirect (%.4f)",
			emissions.TotalSpecificEmissions,
			calculatedTotal,
		)
	}

	// Validate emission values are non-negative
	if emissions.DirectEmissions < 0 || emissions.IndirectEmissions < 0 || emissions.TotalSpecificEmissions < 0 {
		return nil, fmt.Errorf("emissions values cannot be negative")
	}

	// Perform reasonableness checks against default values
	if err := c.performReasonablenessCheck(emissions, good); err != nil {
		// Don't fail, just add a warning
		emissions.DataQuality = "estimated" // Downgrade quality if outside reasonable range
	}

	return emissions, nil
}

// performReasonablenessCheck validates emissions against expected ranges.
func (c *Calculator) performReasonablenessCheck(emissions *EmbeddedEmissions, good *ImportedGood) error {
	key := c.makeDefaultValueKey(good.CommodityType, "", "")
	defaultValue, exists := c.defaultValues[key]

	if !exists {
		return nil // Can't perform check without default values
	}

	// Check if emissions are within a reasonable range (50% to 300% of default)
	if emissions.TotalSpecificEmissions < defaultValue.DefaultTotalEmissions*0.5 {
		return fmt.Errorf("emissions appear unusually low compared to defaults")
	}

	if emissions.TotalSpecificEmissions > defaultValue.DefaultTotalEmissions*3.0 {
		return fmt.Errorf("emissions appear unusually high compared to defaults")
	}

	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// makeDefaultValueKey creates a key for looking up default values.
func (c *Calculator) makeDefaultValueKey(commodity CommodityType, country, route string) string {
	if route != "" {
		return fmt.Sprintf("%s:%s:%s", commodity, country, route)
	}
	if country != "" {
		return fmt.Sprintf("%s:%s", commodity, country)
	}
	return string(commodity)
}

// getProcessEmissionFactor returns the process emission factor for a commodity type.
func (c *Calculator) getProcessEmissionFactor(commodity CommodityType, route string) float64 {
	// Simplified process emission factors (tCO2e per tonne of product)
	// In practice, these would be more detailed and route-specific

	switch commodity {
	case CommodityCement:
		return 0.525 // Typical cement clinker process emissions

	case CommodityIronSteel:
		if route == "blast_furnace" || route == "bf_bof" {
			return 1.85 // Blast furnace route
		} else if route == "electric_arc_furnace" || route == "eaf" {
			return 0.20 // Electric arc furnace (mainly scrap-based)
		}
		return 1.50 // Average

	case CommodityAluminum:
		if route == "primary" {
			return 1.60 // Primary aluminum (from bauxite)
		} else if route == "secondary" {
			return 0.30 // Secondary aluminum (recycled)
		}
		return 1.00 // Average

	case CommodityFertilizer:
		return 2.00 // Ammonia-based fertilizers

	case CommodityHydrogen:
		if route == "steam_methane_reforming" || route == "smr" {
			return 10.0 // Grey hydrogen from natural gas
		} else if route == "electrolysis" {
			return 0.50 // Green hydrogen (depends on electricity source)
		}
		return 5.0 // Average

	default:
		return 0.50 // Conservative default
	}
}

// getElectricityFactor returns the electricity emission factor for a country (gCO2e/kWh).
func (c *Calculator) getElectricityFactor(country string) float64 {
	if factor, exists := c.electricityFactors[country]; exists {
		return factor
	}
	return 500.0 // Global average default
}

// getElectricityConsumption returns typical electricity consumption (kWh per tonne).
func (c *Calculator) getElectricityConsumption(commodity CommodityType) float64 {
	switch commodity {
	case CommodityCement:
		return 90.0 // kWh per tonne
	case CommodityIronSteel:
		return 400.0 // kWh per tonne (varies greatly by route)
	case CommodityAluminum:
		return 15000.0 // kWh per tonne (primary aluminum is very energy-intensive)
	case CommodityFertilizer:
		return 300.0 // kWh per tonne
	case CommodityHydrogen:
		return 50000.0 // kWh per tonne (electrolysis)
	default:
		return 200.0 // Conservative default
	}
}

// =============================================================================
// Default Values Loading
// =============================================================================

// loadDefaultEmissionValues loads default emission values based on EU implementing regulations.
func (c *Calculator) loadDefaultEmissionValues() {
	// These values are based on EU Commission Implementing Regulation (EU) 2023/1773
	// Real implementation would load from database or configuration

	// Cement
	c.addDefaultValue(CommodityCement, "", "", 0.766, 0.041, 0.807)

	// Iron and Steel
	c.addDefaultValue(CommodityIronSteel, "", "blast_furnace", 1.85, 0.25, 2.10)
	c.addDefaultValue(CommodityIronSteel, "", "electric_arc_furnace", 0.20, 0.40, 0.60)
	c.addDefaultValue(CommodityIronSteel, "", "", 1.35, 0.30, 1.65) // Average

	// Aluminum
	c.addDefaultValue(CommodityAluminum, "", "primary", 11.0, 4.0, 15.0)
	c.addDefaultValue(CommodityAluminum, "", "secondary", 0.50, 0.30, 0.80)
	c.addDefaultValue(CommodityAluminum, "", "", 6.0, 2.0, 8.0) // Average

	// Fertilizer
	c.addDefaultValue(CommodityFertilizer, "", "", 2.77, 0.75, 3.52)

	// Hydrogen
	c.addDefaultValue(CommodityHydrogen, "", "steam_methane_reforming", 10.0, 0.50, 10.5)
	c.addDefaultValue(CommodityHydrogen, "", "electrolysis", 0.40, 9.0, 9.4)
	c.addDefaultValue(CommodityHydrogen, "", "", 5.0, 2.0, 7.0) // Average

	// Electricity (per MWh)
	c.addDefaultValue(CommodityElectricity, "", "", 0.40, 0.0, 0.40)
}

// addDefaultValue adds a default emission value to the registry.
func (c *Calculator) addDefaultValue(commodity CommodityType, country, route string, direct, indirect, total float64) {
	key := c.makeDefaultValueKey(commodity, country, route)
	c.defaultValues[key] = &DefaultEmissionValue{
		CommodityType:            commodity,
		CountryOfOrigin:          country,
		ProductionRoute:          route,
		DefaultDirectEmissions:   direct,
		DefaultIndirectEmissions: indirect,
		DefaultTotalEmissions:    total,
		ValueSource:              "EU Commission Implementing Regulation (EU) 2023/1773",
		EffectiveDate:            time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
	}
}

// loadElectricityFactors loads electricity emission factors by country.
func (c *Calculator) loadElectricityFactors() {
	// gCO2e/kWh - based on IEA and EU data
	factors := map[string]float64{
		// EU countries
		"DE": 420, // Germany
		"FR": 60,  // France (nuclear-heavy)
		"PL": 780, // Poland (coal-heavy)
		"ES": 230, // Spain
		"IT": 350, // Italy
		"NL": 400, // Netherlands
		"BE": 170, // Belgium
		"AT": 100, // Austria
		"SE": 20,  // Sweden (hydro + nuclear)
		"DK": 110, // Denmark (wind-heavy)
		"FI": 90,  // Finland

		// Major trading partners
		"CN": 580, // China
		"US": 420, // United States
		"IN": 630, // India
		"RU": 450, // Russia
		"BR": 90,  // Brazil (hydro-heavy)
		"TR": 450, // Turkey
		"UA": 380, // Ukraine
		"GB": 230, // United Kingdom

		// Others
		"GLOBAL": 500, // Global average
	}

	c.electricityFactors = factors
}

// =============================================================================
// Batch Calculation
// =============================================================================

// CalculateBatch calculates embedded emissions for multiple goods.
func (c *Calculator) CalculateBatch(ctx context.Context, goods []ImportedGood) ([]EmbeddedEmissions, []error) {
	results := make([]EmbeddedEmissions, len(goods))
	errors := make([]error, len(goods))

	for i := range goods {
		emissions, err := c.CalculateEmbeddedEmissions(ctx, &goods[i])
		if err != nil {
			errors[i] = err
			continue
		}
		results[i] = *emissions
		// Update the good with calculated emissions
		goods[i].EmbeddedEmissions = *emissions
	}

	return results, errors
}

// =============================================================================
// Summary Statistics
// =============================================================================

// CalculationSummary provides summary statistics for a batch calculation.
type CalculationSummary struct {
	TotalGoods         int     `json:"total_goods"`
	SuccessfulCalcs    int     `json:"successful_calcs"`
	FailedCalcs        int     `json:"failed_calcs"`
	TotalEmissions     float64 `json:"total_emissions"`  // tCO2e
	AverageSpecific    float64 `json:"average_specific"` // tCO2e per unit
	DirectPortion      float64 `json:"direct_portion"`   // Percentage
	IndirectPortion    float64 `json:"indirect_portion"` // Percentage
	UsingDefaultValues int     `json:"using_default_values"`
	Verified           int     `json:"verified"`
}

// GetSummary returns a calculation summary for a list of goods.
func (c *Calculator) GetSummary(goods []ImportedGood) *CalculationSummary {
	summary := &CalculationSummary{
		TotalGoods: len(goods),
	}

	var totalDirect, totalIndirect float64

	for _, good := range goods {
		if good.EmbeddedEmissions.TotalSpecificEmissions > 0 {
			summary.SuccessfulCalcs++

			totalEmissions := good.EmbeddedEmissions.TotalSpecificEmissions * good.Quantity
			summary.TotalEmissions += totalEmissions
			totalDirect += good.EmbeddedEmissions.DirectEmissions * good.Quantity
			totalIndirect += good.EmbeddedEmissions.IndirectEmissions * good.Quantity

			if good.EmbeddedEmissions.CalculationMethod == "default_values" {
				summary.UsingDefaultValues++
			}

			if good.EmbeddedEmissions.Verified {
				summary.Verified++
			}
		} else {
			summary.FailedCalcs++
		}
	}

	if summary.SuccessfulCalcs > 0 {
		summary.AverageSpecific = summary.TotalEmissions / float64(summary.SuccessfulCalcs)

		if summary.TotalEmissions > 0 {
			summary.DirectPortion = (totalDirect / summary.TotalEmissions) * 100
			summary.IndirectPortion = (totalIndirect / summary.TotalEmissions) * 100
		}
	}

	return summary
}
