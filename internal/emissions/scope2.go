// Package emissions/scope2 provides Scope 2 indirect emissions calculations.
//
// Scope 2 emissions are indirect greenhouse gas emissions from the generation
// of purchased energy consumed by the organization, including:
//   - Purchased electricity
//   - Purchased steam
//   - Purchased heating
//   - Purchased cooling
//
// The GHG Protocol requires organizations to report Scope 2 using two methods:
//   - Location-based: Uses average grid emission factors
//   - Market-based: Uses supplier-specific factors (RECs, PPAs, contracts)
package emissions

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// =============================================================================
// Scope 2 Default Factors
// =============================================================================

// DefaultScope2Factor is the fallback emission factor for electricity
// when region-specific data is unavailable.
// Value: 0.4 kg CO2e per kWh (approximate global average)
const DefaultScope2Factor = 0.4

// RegionalGridFactors contains location-based emission factors for major regions.
// Values are in kg CO2e per kWh for purchased electricity.
// Source: IEA, EPA eGRID, EEA, national grid operators.
var RegionalGridFactors = map[string]float64{
	// United States regions (EPA eGRID 2023)
	"US-AVERAGE":    0.386,
	"US-WEST":       0.298, // WECC
	"US-EAST":       0.388, // NPCC + RFC + SERC
	"US-TEXAS":      0.395, // ERCOT
	"US-MIDWEST":    0.452, // MRO
	"US-CALIFORNIA": 0.225, // CAMX subregion

	// European regions (EEA 2023)
	"EU-AVERAGE": 0.276,
	"EU-CENTRAL": 0.350, // Germany, Poland
	"EU-NORTH":   0.150, // Nordic countries
	"EU-SOUTH":   0.295, // Italy, Spain
	"EU-WEST":    0.185, // France, Belgium
	"EU-UK":      0.207,
	"EU-FRANCE":  0.052, // Low due to nuclear
	"EU-GERMANY": 0.385,
	"EU-SPAIN":   0.212,
	"EU-POLAND":  0.723,

	// Asia-Pacific regions
	"ASIA-PACIFIC":   0.550,
	"ASIA-JAPAN":     0.470,
	"ASIA-CHINA":     0.555,
	"ASIA-INDIA":     0.708,
	"ASIA-KOREA":     0.459,
	"ASIA-AUSTRALIA": 0.656,
	"ASIA-SINGAPORE": 0.408,

	// Other regions
	"LATAM-BRAZIL": 0.075, // Mostly hydro
	"LATAM-MEXICO": 0.435,
	"AFRICA-SOUTH": 0.928,
	"MIDDLE-EAST":  0.650,
	"CANADA":       0.130,
}

// =============================================================================
// Scope 2 Calculator Configuration
// =============================================================================

// Scope2Config configures the Scope 2 calculator.
type Scope2Config struct {
	// Registry provides emission factor lookup.
	Registry FactorRegistry

	// Logger for calculation events.
	Logger *slog.Logger

	// DefaultRegion is used when activity location is empty.
	DefaultRegion string

	// DefaultMethod is the calculation method when not specified.
	// Defaults to location-based per GHG Protocol.
	DefaultMethod CalculationMethod

	// PreferMarketBased prefers market-based factors when available.
	PreferMarketBased bool

	// StrictRegionMatching fails if exact region match is not found.
	// When false, falls back to parent region or default.
	StrictRegionMatching bool
}

// DefaultScope2Config returns sensible defaults for Scope 2 calculations.
func DefaultScope2Config() Scope2Config {
	return Scope2Config{
		DefaultRegion: "US-AVERAGE",
		DefaultMethod: MethodLocationBased,
	}
}

// =============================================================================
// Scope 2 Calculator Implementation
// =============================================================================

// Scope2Calculator computes Scope 2 indirect emissions from purchased energy.
//
// The calculator supports both location-based and market-based methods:
//
// Location-based (default):
//   - Uses grid average emission factors for the geographic region
//   - Required for GHG Protocol reporting
//   - Reflects actual grid composition
//
// Market-based:
//   - Uses supplier-specific factors or residual mix
//   - Accounts for renewable energy purchases
//   - Requires contractual instruments (RECs, PPAs)
//
// Example usage:
//
//	calc := emissions.NewScope2Calculator(config)
//	record, err := calc.Calculate(ctx, activity)
type Scope2Calculator struct {
	registry FactorRegistry
	logger   *slog.Logger
	config   Scope2Config
}

// NewScope2Calculator creates a new Scope 2 emissions calculator.
func NewScope2Calculator(cfg Scope2Config) *Scope2Calculator {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	// Apply defaults
	if cfg.DefaultMethod == "" {
		cfg.DefaultMethod = MethodLocationBased
	}

	return &Scope2Calculator{
		registry: cfg.Registry,
		logger:   logger,
		config:   cfg,
	}
}

// Supports returns true if this calculator can handle the activity.
func (c *Scope2Calculator) Supports(activity Activity) bool {
	source := activity.GetSource()
	unit := activity.GetUnit()

	// Check if it's an energy-related source
	switch source {
	case "utility_bill", "electricity", "steam", "heating", "cooling":
		// Check unit is energy-based
		switch unit {
		case "kWh", "MWh", "GJ", "therm", "MMBtu":
			return true
		}
	}

	return false
}

// Calculate computes Scope 2 emissions for the given activity.
func (c *Scope2Calculator) Calculate(ctx context.Context, activity Activity) (EmissionRecord, error) {
	if activity == nil {
		return EmissionRecord{}, ErrNilActivity
	}

	// Validate unit is supported
	if !c.Supports(activity) {
		return EmissionRecord{}, fmt.Errorf(
			"unsupported activity: source=%s unit=%s: %w",
			activity.GetSource(), activity.GetUnit(), ErrUnsupportedUnit,
		)
	}

	c.logger.Debug("calculating scope 2 emissions",
		"activity_id", activity.GetID(),
		"source", activity.GetSource(),
		"location", activity.GetLocation(),
		"quantity", activity.GetQuantity(),
		"unit", activity.GetUnit(),
	)

	// Normalize quantity to kWh
	quantityKWh := c.normalizeToKWh(activity.GetQuantity(), activity.GetUnit())

	// Find the appropriate emission factor
	factor, method, err := c.findFactor(ctx, activity)
	if err != nil {
		return EmissionRecord{}, fmt.Errorf("find scope 2 factor: %w", err)
	}

	// Calculate emissions: kWh × factor (kg CO2e/kWh)
	emissionsKg := quantityKWh * factor.ValueKgCO2ePerUnit

	record := EmissionRecord{
		ID:                  GenerateRecordID(),
		ActivityID:          activity.GetID(),
		FactorID:            factor.ID,
		Scope:               Scope2,
		EmissionsKgCO2e:     emissionsKg,
		EmissionsTonnesCO2e: KgToTonnes(emissionsKg),
		InputQuantity:       activity.GetQuantity(),
		InputUnit:           activity.GetUnit(),
		EmissionFactor:      factor.ValueKgCO2ePerUnit,
		Method:              method,
		DataQuality:         c.determineDataQuality(factor),
		Region:              activity.GetLocation(),
		OrgID:               activity.GetOrgID(),
		WorkspaceID:         activity.GetWorkspaceID(),
		PeriodStart:         activity.GetPeriodStart(),
		PeriodEnd:           activity.GetPeriodEnd(),
		CalculatedAt:        time.Now().UTC(),
	}

	c.logger.Info("calculated scope 2 emissions",
		"activity_id", activity.GetID(),
		"factor_id", factor.ID,
		"method", method,
		"emissions_kg_co2e", emissionsKg,
	)

	return record, nil
}

// CalculateBatch processes multiple activities.
func (c *Scope2Calculator) CalculateBatch(ctx context.Context, activities []Activity) ([]EmissionRecord, error) {
	records := make([]EmissionRecord, 0, len(activities))

	for _, activity := range activities {
		if !c.Supports(activity) {
			continue
		}

		record, err := c.Calculate(ctx, activity)
		if err != nil {
			c.logger.Warn("failed to calculate scope 2 activity",
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
func (c *Scope2Calculator) findFactor(ctx context.Context, activity Activity) (EmissionFactor, CalculationMethod, error) {
	location := activity.GetLocation()
	method := c.config.DefaultMethod

	// First try the registry for exact match
	if c.registry != nil {
		query := FactorQuery{
			Scope:   Scope2,
			Region:  location,
			Source:  activity.GetSource(),
			Unit:    "kWh",
			ValidAt: time.Now(),
		}

		// Try market-based first if preferred
		if c.config.PreferMarketBased {
			factor, err := c.registry.FindFactor(ctx, query)
			if err == nil && factor.Method == MethodMarketBased {
				return factor, MethodMarketBased, nil
			}
		}

		factor, err := c.registry.FindFactor(ctx, query)
		if err == nil {
			return factor, factor.Method, nil
		}
	}

	// Fall back to regional grid factors
	factor, err := c.getRegionalFactor(location)
	if err != nil && c.config.StrictRegionMatching {
		return EmissionFactor{}, method, err
	}

	if err != nil {
		// Use global default
		c.logger.Warn("using default emission factor",
			"location", location,
			"default_factor", DefaultScope2Factor,
		)
		factor = EmissionFactor{
			ID:                 "default-scope2-global",
			Scope:              Scope2,
			Region:             "GLOBAL",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: DefaultScope2Factor,
			Method:             MethodLocationBased,
			DataSource:         "Global average (default)",
		}
	}

	return factor, MethodLocationBased, nil
}

// getRegionalFactor looks up the built-in regional grid factor.
func (c *Scope2Calculator) getRegionalFactor(location string) (EmissionFactor, error) {
	// Normalize location
	location = strings.ToUpper(strings.TrimSpace(location))

	// Try exact match first
	if factorValue, ok := RegionalGridFactors[location]; ok {
		return EmissionFactor{
			ID:                 fmt.Sprintf("grid-%s", strings.ToLower(location)),
			Scope:              Scope2,
			Region:             location,
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: factorValue,
			Method:             MethodLocationBased,
			DataSource:         "IEA/EPA/EEA grid emission factors",
			CreatedAt:          time.Now().UTC(),
		}, nil
	}

	// Try parent region (e.g., "US-CALIFORNIA" → "US-WEST" → "US-AVERAGE")
	if strings.HasPrefix(location, "US-") {
		for _, parent := range []string{"US-WEST", "US-EAST", "US-AVERAGE"} {
			if factorValue, ok := RegionalGridFactors[parent]; ok {
				return EmissionFactor{
					ID:                 fmt.Sprintf("grid-%s", strings.ToLower(parent)),
					Scope:              Scope2,
					Region:             parent,
					Source:             "electricity",
					Unit:               "kWh",
					ValueKgCO2ePerUnit: factorValue,
					Method:             MethodLocationBased,
					DataSource:         "IEA/EPA/EEA grid emission factors (fallback)",
					Notes:              fmt.Sprintf("Fallback from %s to %s", location, parent),
				}, nil
			}
		}
	}

	if strings.HasPrefix(location, "EU-") {
		if factorValue, ok := RegionalGridFactors["EU-AVERAGE"]; ok {
			return EmissionFactor{
				ID:                 "grid-eu-average",
				Scope:              Scope2,
				Region:             "EU-AVERAGE",
				Source:             "electricity",
				Unit:               "kWh",
				ValueKgCO2ePerUnit: factorValue,
				Method:             MethodLocationBased,
				DataSource:         "EEA grid emission factors (fallback)",
			}, nil
		}
	}

	if strings.HasPrefix(location, "ASIA-") {
		if factorValue, ok := RegionalGridFactors["ASIA-PACIFIC"]; ok {
			return EmissionFactor{
				ID:                 "grid-asia-pacific",
				Scope:              Scope2,
				Region:             "ASIA-PACIFIC",
				Source:             "electricity",
				Unit:               "kWh",
				ValueKgCO2ePerUnit: factorValue,
				Method:             MethodLocationBased,
				DataSource:         "IEA grid emission factors (fallback)",
			}, nil
		}
	}

	return EmissionFactor{}, fmt.Errorf("no emission factor for region %q: %w", location, ErrFactorNotFound)
}

// normalizeToKWh converts various energy units to kWh.
func (c *Scope2Calculator) normalizeToKWh(quantity float64, unit string) float64 {
	switch unit {
	case "kWh":
		return quantity
	case "MWh":
		return quantity * 1000
	case "GJ":
		return quantity * 277.778 // 1 GJ = 277.778 kWh
	case "therm":
		return quantity * 29.3071 // 1 therm = 29.3 kWh
	case "MMBtu":
		return quantity * 293.071 // 1 MMBtu = 293 kWh
	default:
		c.logger.Warn("unknown energy unit, treating as kWh",
			"unit", unit,
		)
		return quantity
	}
}

// determineDataQuality assesses the quality of the emission factor.
func (c *Scope2Calculator) determineDataQuality(factor EmissionFactor) DataQuality {
	if factor.DataSource == "" || strings.Contains(factor.DataSource, "default") {
		return DataQualityDefault
	}
	if strings.Contains(factor.DataSource, "fallback") {
		return DataQualityEstimated
	}
	return DataQualityMeasured
}

// =============================================================================
// Scope 2 Reporting Utilities
// =============================================================================

// Scope2Summary aggregates Scope 2 emissions by method.
type Scope2Summary struct {
	// LocationBasedKgCO2e is total using grid average factors.
	LocationBasedKgCO2e float64 `json:"location_based_kg_co2e"`

	// MarketBasedKgCO2e is total using supplier-specific factors.
	MarketBasedKgCO2e float64 `json:"market_based_kg_co2e"`

	// TotalKWh is the total electricity consumption.
	TotalKWh float64 `json:"total_kwh"`

	// RecordCount is how many records were summarized.
	RecordCount int `json:"record_count"`
}

// SummarizeScope2 aggregates Scope 2 emissions by calculation method.
func SummarizeScope2(records []EmissionRecord) Scope2Summary {
	var summary Scope2Summary

	for _, r := range records {
		if r.Scope != Scope2 {
			continue
		}

		summary.RecordCount++

		switch r.Method {
		case MethodLocationBased:
			summary.LocationBasedKgCO2e += r.EmissionsKgCO2e
		case MethodMarketBased:
			summary.MarketBasedKgCO2e += r.EmissionsKgCO2e
		default:
			// Default to location-based
			summary.LocationBasedKgCO2e += r.EmissionsKgCO2e
		}

		// Accumulate kWh (normalize if needed)
		if r.InputUnit == "kWh" {
			summary.TotalKWh += r.InputQuantity
		} else if r.InputUnit == "MWh" {
			summary.TotalKWh += r.InputQuantity * 1000
		}
	}

	return summary
}

// CalculateRenewableOffset estimates emissions avoided through renewable energy.
func CalculateRenewableOffset(renewableKWh float64, region string) float64 {
	// Get the grid factor for the region
	factorValue, ok := RegionalGridFactors[strings.ToUpper(region)]
	if !ok {
		factorValue = DefaultScope2Factor
	}

	// Offset = renewable kWh × grid factor
	// This represents emissions that would have occurred without renewables
	return renewableKWh * factorValue
}
