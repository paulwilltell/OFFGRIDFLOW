// Package emissions/scope3 provides Scope 3 indirect emissions calculations.
//
// Scope 3 emissions are all indirect greenhouse gas emissions (not included in
// Scope 2) that occur in the value chain of the reporting organization.
// The GHG Protocol defines 15 categories of Scope 3 emissions:
//
// Upstream categories:
//  1. Purchased goods and services
//  2. Capital goods
//  3. Fuel- and energy-related activities (not in Scope 1 or 2)
//  4. Upstream transportation and distribution
//  5. Waste generated in operations
//  6. Business travel
//  7. Employee commuting
//  8. Upstream leased assets
//
// Downstream categories:
//  9. Downstream transportation and distribution
//  10. Processing of sold products
//  11. Use of sold products
//  12. End-of-life treatment of sold products
//  13. Downstream leased assets
//  14. Franchises
//  15. Investments
package emissions

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// =============================================================================
// Scope 3 Category Constants
// =============================================================================

// Scope3Category identifies the GHG Protocol Scope 3 category.
type Scope3Category int

const (
	CategoryUnspecified            Scope3Category = 0
	CategoryPurchasedGoods         Scope3Category = 1
	CategoryCapitalGoods           Scope3Category = 2
	CategoryFuelEnergy             Scope3Category = 3
	CategoryUpstreamTransport      Scope3Category = 4
	CategoryWaste                  Scope3Category = 5
	CategoryBusinessTravel         Scope3Category = 6
	CategoryCommuting              Scope3Category = 7
	CategoryUpstreamLeasedAssets   Scope3Category = 8
	CategoryDownstreamTransport    Scope3Category = 9
	CategoryProcessingSold         Scope3Category = 10
	CategoryUseSoldProducts        Scope3Category = 11
	CategoryEndOfLife              Scope3Category = 12
	CategoryDownstreamLeasedAssets Scope3Category = 13
	CategoryFranchises             Scope3Category = 14
	CategoryInvestments            Scope3Category = 15
)

// String returns the human-readable category name.
func (c Scope3Category) String() string {
	names := map[Scope3Category]string{
		CategoryPurchasedGoods:         "Purchased Goods and Services",
		CategoryCapitalGoods:           "Capital Goods",
		CategoryFuelEnergy:             "Fuel and Energy Activities",
		CategoryUpstreamTransport:      "Upstream Transportation",
		CategoryWaste:                  "Waste",
		CategoryBusinessTravel:         "Business Travel",
		CategoryCommuting:              "Employee Commuting",
		CategoryUpstreamLeasedAssets:   "Upstream Leased Assets",
		CategoryDownstreamTransport:    "Downstream Transportation",
		CategoryProcessingSold:         "Processing of Sold Products",
		CategoryUseSoldProducts:        "Use of Sold Products",
		CategoryEndOfLife:              "End-of-Life Treatment",
		CategoryDownstreamLeasedAssets: "Downstream Leased Assets",
		CategoryFranchises:             "Franchises",
		CategoryInvestments:            "Investments",
	}

	if name, ok := names[c]; ok {
		return name
	}
	return "Unknown Category"
}

// IsUpstream returns true for categories 1-8.
func (c Scope3Category) IsUpstream() bool {
	return c >= CategoryPurchasedGoods && c <= CategoryUpstreamLeasedAssets
}

// IsDownstream returns true for categories 9-15.
func (c Scope3Category) IsDownstream() bool {
	return c >= CategoryDownstreamTransport && c <= CategoryInvestments
}

// =============================================================================
// Scope 3 Emission Factors
// =============================================================================

// Default Scope 3 emission factors by category and subcategory.
// Values vary widely based on methodology and data availability.

// BusinessTravelFactors contains emission factors for business travel modes.
// Values in kg CO2e per passenger-km.
var BusinessTravelFactors = map[string]float64{
	"flight-domestic":    0.255, // Short-haul flights
	"flight-short":       0.156, // Medium-haul flights
	"flight-long":        0.195, // Long-haul flights
	"flight-first-class": 0.585, // First/business class multiplier
	"train":              0.041, // Average train
	"train-high-speed":   0.006, // High-speed electric rail
	"car-rental":         0.171, // Average rental car
	"car-taxi":           0.210, // Taxi/rideshare
	"bus":                0.089, // Coach/bus
	"hotel-night":        20.6,  // Per night (varies by region)
}

// CommutingFactors contains emission factors for employee commuting.
// Values in kg CO2e per passenger-km.
var CommutingFactors = map[string]float64{
	"car-petrol":     0.171,
	"car-diesel":     0.168,
	"car-hybrid":     0.120,
	"car-electric":   0.053, // Depends on grid mix
	"motorcycle":     0.114,
	"public-transit": 0.089,
	"bus":            0.089,
	"train":          0.041,
	"bicycle":        0.000,
	"walking":        0.000,
	"work-from-home": 0.800, // Per day (heating, equipment)
}

// WasteFactors contains emission factors for waste treatment.
// Values in kg CO2e per kg of waste.
var WasteFactors = map[string]float64{
	"landfill-mixed":      0.467,
	"landfill-organic":    0.623,
	"incineration":        0.989,
	"recycling-paper":     -0.139, // Credit for avoided virgin production
	"recycling-plastic":   -0.573,
	"recycling-metal":     -1.467,
	"recycling-glass":     -0.314,
	"composting":          0.010,
	"anaerobic-digestion": -0.116,
}

// SpendBasedFactors contains emission factors for spend-based calculations.
// Values in kg CO2e per USD spent (EEIO-based factors).
var SpendBasedFactors = map[string]float64{
	"electronics":              0.35,
	"office-supplies":          0.18,
	"professional-services":    0.12,
	"software":                 0.05,
	"cloud-services":           0.08,
	"furniture":                0.25,
	"food-catering":            0.55,
	"marketing":                0.15,
	"construction":             0.42,
	"manufacturing-avg":        0.45,
	"general-goods":            0.30,
	"use_sold_products":        0.40,
	"processing_sold_products": 0.50,
	"franchises":               0.80,
	"investments":              0.70,
}

// FuelEnergyFactors captures well-to-tank and T&D losses (category 3).
// Units in kg CO2e per kWh equivalent.
var FuelEnergyFactors = map[string]float64{
	"transmission_distribution": 0.05,
	"fuel_wtt":                  0.04,
}

// TransportFactors for upstream/downstream transportation (per tonne-km).
var TransportFactors = map[string]float64{
	"truck_freight":               0.089,
	"rail_freight":                0.022,
	"ship_freight":                0.012,
	"air_freight":                 0.602,
	"transportation_distribution": 0.089,
	"downstream_transport":        0.089,
}

// UsePhaseFactors for category 11 (use of sold products) in kg/kWh.
var UsePhaseFactors = map[string]float64{
	"use_sold_products": 0.45,
}

// EndOfLifeFactors for category 12 (kg CO2e per kg waste).
var EndOfLifeFactors = map[string]float64{
	"end_of_life_treatment": 0.35,
}

// LeasedAssetsFactors for categories 8 and 13 (kg CO2e per m2-year proxy).
var LeasedAssetsFactors = map[string]float64{
	"leased_assets": 12.0,
}

// =============================================================================
// Scope 3 Calculator Configuration
// =============================================================================

// Scope3Config configures the Scope 3 calculator.
type Scope3Config struct {
	// Registry provides emission factor lookup.
	Registry FactorRegistry

	// Logger for calculation events.
	Logger *slog.Logger

	// DefaultCategory is used when category cannot be determined.
	DefaultCategory Scope3Category

	// PreferSupplierData uses supplier-specific factors when available.
	PreferSupplierData bool

	// SpendCurrency is the default currency for spend-based calculations.
	SpendCurrency string
}

// DefaultScope3Config returns sensible defaults.
func DefaultScope3Config() Scope3Config {
	return Scope3Config{
		DefaultCategory: CategoryPurchasedGoods,
		SpendCurrency:   "USD",
	}
}

// =============================================================================
// Scope 3 Calculator Implementation
// =============================================================================

// Scope3Calculator computes Scope 3 value chain emissions.
//
// Scope 3 calculations often rely on:
//   - Spend-based methods (using economic input-output factors)
//   - Activity-based methods (using physical activity data)
//   - Supplier-specific data (actual supplier emissions)
//
// The calculator prioritizes methods from highest to lowest accuracy:
//  1. Supplier-specific data
//  2. Activity-based calculations
//  3. Spend-based estimates
type Scope3Calculator struct {
	registry FactorRegistry
	logger   *slog.Logger
	config   Scope3Config
}

// NewScope3Calculator creates a new Scope 3 emissions calculator.
func NewScope3Calculator(cfg Scope3Config) *Scope3Calculator {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &Scope3Calculator{
		registry: cfg.Registry,
		logger:   logger,
		config:   cfg,
	}
}

// Supports returns true if this calculator can handle the activity.
func (c *Scope3Calculator) Supports(activity Activity) bool {
	category := c.determineCategory(activity)
	return category != CategoryUnspecified
}

// Calculate computes Scope 3 emissions for the given activity.
func (c *Scope3Calculator) Calculate(ctx context.Context, activity Activity) (EmissionRecord, error) {
	if activity == nil {
		return EmissionRecord{}, ErrNilActivity
	}

	c.logger.Debug("calculating scope 3 emissions",
		"activity_id", activity.GetID(),
		"source", activity.GetSource(),
		"category", activity.GetCategory(),
		"quantity", activity.GetQuantity(),
		"unit", activity.GetUnit(),
	)

	// Determine the Scope 3 category
	category := c.determineCategory(activity)
	if category == CategoryUnspecified {
		return EmissionRecord{}, fmt.Errorf("cannot calculate category %s for activity %s: %w",
			category.String(), activity.GetID(), ErrUnsupportedUnit)
	}

	factor, err := c.findFactor(ctx, activity)
	if err != nil {
		return EmissionRecord{}, err
	}

	quantity, err := c.convertQuantity(activity.GetQuantity(), activity.GetUnit(), factor.Unit)
	if err != nil {
		return EmissionRecord{}, err
	}

	emissionsKg := quantity * factor.ValueKgCO2ePerUnit

	method := factor.Method
	if method == "" {
		if c.isSpendActivity(activity) {
			method = MethodSpendBased
		} else {
			method = MethodActivityBased
		}
	}

	record := EmissionRecord{
		ID:                  GenerateRecordID(),
		ActivityID:          activity.GetID(),
		FactorID:            factor.ID,
		Scope:               Scope3,
		EmissionsKgCO2e:     emissionsKg,
		EmissionsTonnesCO2e: KgToTonnes(emissionsKg),
		InputQuantity:       activity.GetQuantity(),
		InputUnit:           activity.GetUnit(),
		EmissionFactor:      factor.ValueKgCO2ePerUnit,
		Method:              method,
		DataQuality:         c.determineDataQuality(method),
		Region:              activity.GetLocation(),
		OrgID:               activity.GetOrgID(),
		WorkspaceID:         activity.GetWorkspaceID(),
		PeriodStart:         activity.GetPeriodStart(),
		PeriodEnd:           activity.GetPeriodEnd(),
		CalculatedAt:        time.Now().UTC(),
		Notes:               fmt.Sprintf("Category: %s", category.String()),
	}

	c.logger.Info("calculated scope 3 emissions",
		"activity_id", activity.GetID(),
		"category", category.String(),
		"method", method,
		"emissions_kg_co2e", emissionsKg,
	)

	return record, nil
}

// CalculateBatch processes multiple activities.
func (c *Scope3Calculator) CalculateBatch(ctx context.Context, activities []Activity) ([]EmissionRecord, error) {
	records := make([]EmissionRecord, 0, len(activities))

	for _, activity := range activities {
		if !c.Supports(activity) {
			continue
		}

		record, err := c.Calculate(ctx, activity)
		if err != nil {
			c.logger.Warn("failed to calculate scope 3 activity",
				"activity_id", activity.GetID(),
				"error", err,
			)
			continue
		}

		records = append(records, record)
	}

	return records, nil
}

// determineCategory infers the Scope 3 category from activity data.
func (c *Scope3Calculator) determineCategory(activity Activity) Scope3Category {
	source := strings.ToLower(activity.GetSource())
	category := strings.ToLower(activity.GetCategory())

	switch source {
	case "travel", "business_travel":
		return CategoryBusinessTravel
	case "commuting":
		return CategoryCommuting
	case "waste":
		return CategoryWaste
	case "purchases":
		switch category {
		case "capital", "equipment", "machinery", "capital_goods":
			return CategoryCapitalGoods
		case "purchased_goods_services":
			return CategoryPurchasedGoods
		default:
			return CategoryPurchasedGoods
		}
	case "upstream":
		switch category {
		case "fuel_energy_activities":
			return CategoryFuelEnergy
		case "transportation_distribution", "truck_freight", "freight", "shipping":
			return CategoryUpstreamTransport
		case "leased_assets":
			return CategoryUpstreamLeasedAssets
		default:
			return CategoryFuelEnergy
		}
	case "downstream":
		switch category {
		case "transportation_distribution":
			return CategoryDownstreamTransport
		case "processing_sold_products":
			return CategoryProcessingSold
		case "use_sold_products":
			return CategoryUseSoldProducts
		case "end_of_life_treatment":
			return CategoryEndOfLife
		case "leased_assets":
			return CategoryDownstreamLeasedAssets
		case "franchises":
			return CategoryFranchises
		case "investments":
			return CategoryInvestments
		default:
			return CategoryDownstreamTransport
		}
	case "freight", "shipping":
		return CategoryUpstreamTransport
	case "investment":
		return CategoryInvestments
	default:
		return CategoryUnspecified
	}
}

// isSpendActivity returns true if the activity appears to be spend-based.
func (c *Scope3Calculator) isSpendActivity(activity Activity) bool {
	unit := activity.GetUnit()
	return unit == "USD" || unit == "EUR" || unit == "GBP" ||
		strings.Contains(unit, "$") || strings.HasPrefix(unit, "USD")
}

// findFactor locates a factor, preferring registry lookups when available.
func (c *Scope3Calculator) findFactor(ctx context.Context, activity Activity) (EmissionFactor, error) {
	if c.registry != nil {
		query := FactorQuery{
			Scope:    Scope3,
			Region:   activity.GetLocation(),
			Source:   activity.GetSource(),
			Category: activity.GetCategory(),
			Unit:     activity.GetUnit(),
			ValidAt:  time.Now(),
		}
		return c.registry.FindFactor(ctx, query)
	}

	return c.defaultFactor(activity)
}

// defaultFactor provides a built-in factor when no registry is configured.
func (c *Scope3Calculator) defaultFactor(activity Activity) (EmissionFactor, error) {
	category := c.determineCategory(activity)
	key := strings.ToLower(activity.GetCategory())

	switch category {
	case CategoryFuelEnergy:
		factorKey := mapFuelEnergyKey(key)
		value, ok := FuelEnergyFactors[factorKey]
		if !ok {
			return EmissionFactor{}, ErrFactorNotFound
		}
		return EmissionFactor{
			ID:                 fmt.Sprintf("scope3-fuel-energy-%s", factorKey),
			Scope:              Scope3,
			Region:             activity.GetLocation(),
			Source:             activity.GetSource(),
			Category:           factorKey,
			Unit:               "kWh",
			ValueKgCO2ePerUnit: value,
			Method:             MethodActivityBased,
		}, nil
	case CategoryBusinessTravel:
		factorKey := mapBusinessTravelKey(key)
		value, ok := BusinessTravelFactors[factorKey]
		if !ok {
			return EmissionFactor{}, ErrFactorNotFound
		}
		return EmissionFactor{
			ID:                 fmt.Sprintf("scope3-travel-%s", factorKey),
			Scope:              Scope3,
			Region:             activity.GetLocation(),
			Source:             activity.GetSource(),
			Category:           factorKey,
			Unit:               "km",
			ValueKgCO2ePerUnit: value,
			Method:             MethodActivityBased,
		}, nil
	case CategoryCommuting:
		factorKey := mapCommutingKey(key)
		value, ok := CommutingFactors[factorKey]
		if !ok {
			return EmissionFactor{}, ErrFactorNotFound
		}
		unit := "km"
		if factorKey == "work-from-home" {
			unit = "days"
		}
		return EmissionFactor{
			ID:                 fmt.Sprintf("scope3-commuting-%s", factorKey),
			Scope:              Scope3,
			Region:             activity.GetLocation(),
			Source:             activity.GetSource(),
			Category:           factorKey,
			Unit:               unit,
			ValueKgCO2ePerUnit: value,
			Method:             MethodActivityBased,
		}, nil
	case CategoryWaste:
		factorKey := mapWasteKey(key)
		value, ok := WasteFactors[factorKey]
		if !ok {
			return EmissionFactor{}, ErrFactorNotFound
		}
		return EmissionFactor{
			ID:                 fmt.Sprintf("scope3-waste-%s", factorKey),
			Scope:              Scope3,
			Region:             activity.GetLocation(),
			Source:             activity.GetSource(),
			Category:           factorKey,
			Unit:               "kg",
			ValueKgCO2ePerUnit: value,
			Method:             MethodActivityBased,
		}, nil
	case CategoryUpstreamTransport, CategoryDownstreamTransport:
		factorKey := mapTransportKey(key)
		value, ok := TransportFactors[factorKey]
		if !ok {
			return EmissionFactor{}, ErrFactorNotFound
		}
		return EmissionFactor{
			ID:                 fmt.Sprintf("scope3-transport-%s", factorKey),
			Scope:              Scope3,
			Region:             activity.GetLocation(),
			Source:             activity.GetSource(),
			Category:           factorKey,
			Unit:               "tonne-km",
			ValueKgCO2ePerUnit: value,
			Method:             MethodActivityBased,
		}, nil
	case CategoryUseSoldProducts:
		factorKey := "use_sold_products"
		value, ok := UsePhaseFactors[factorKey]
		if !ok {
			return EmissionFactor{}, ErrFactorNotFound
		}
		return EmissionFactor{
			ID:                 "scope3-use-sold-products",
			Scope:              Scope3,
			Region:             activity.GetLocation(),
			Source:             activity.GetSource(),
			Category:           factorKey,
			Unit:               "kWh",
			ValueKgCO2ePerUnit: value,
			Method:             MethodActivityBased,
		}, nil
	case CategoryEndOfLife:
		factorKey := "end_of_life_treatment"
		value, ok := EndOfLifeFactors[factorKey]
		if !ok {
			return EmissionFactor{}, ErrFactorNotFound
		}
		return EmissionFactor{
			ID:                 "scope3-end-of-life",
			Scope:              Scope3,
			Region:             activity.GetLocation(),
			Source:             activity.GetSource(),
			Category:           factorKey,
			Unit:               "kg",
			ValueKgCO2ePerUnit: value,
			Method:             MethodActivityBased,
		}, nil
	case CategoryUpstreamLeasedAssets, CategoryDownstreamLeasedAssets:
		factorKey := "leased_assets"
		value, ok := LeasedAssetsFactors[factorKey]
		if !ok {
			return EmissionFactor{}, ErrFactorNotFound
		}
		return EmissionFactor{
			ID:                 "scope3-leased-assets",
			Scope:              Scope3,
			Region:             activity.GetLocation(),
			Source:             activity.GetSource(),
			Category:           factorKey,
			Unit:               "m2",
			ValueKgCO2ePerUnit: value,
			Method:             MethodActivityBased,
		}, nil
	default:
		factorKey := mapSpendKey(key)
		value, ok := SpendBasedFactors[factorKey]
		if !ok {
			return EmissionFactor{}, ErrFactorNotFound
		}
		return EmissionFactor{
			ID:                 fmt.Sprintf("scope3-spend-%s", factorKey),
			Scope:              Scope3,
			Region:             activity.GetLocation(),
			Source:             activity.GetSource(),
			Category:           factorKey,
			Unit:               activity.GetUnit(),
			ValueKgCO2ePerUnit: value,
			Method:             MethodSpendBased,
		}, nil
	}
}

// Unit conversion helpers.
func (c *Scope3Calculator) convertQuantity(quantity float64, fromUnit, toUnit string) (float64, error) {
	from := strings.ToLower(fromUnit)
	to := strings.ToLower(toUnit)

	if from == to || to == "" {
		return quantity, nil
	}

	if (from == "mile" || from == "miles") && to == "km" {
		return quantity * 1.60934, nil
	}

	if from == "km" && (to == "mile" || to == "miles") {
		return quantity / 1.60934, nil
	}

	// tonne-km vs ton-km vs tkm
	if (from == "tonne-km" || from == "ton-km" || from == "tkm") && (to == "tonne-km" || to == "ton-km" || to == "tkm") {
		return quantity, nil
	}

	// kg-km to tonne-km
	if from == "kg-km" && (to == "tonne-km" || to == "ton-km" || to == "tkm") {
		return quantity / 1000.0, nil
	}

	// If units are incompatible, signal unsupported.
	return 0, fmt.Errorf("scope3: cannot convert %s to %s: %w", fromUnit, toUnit, ErrUnsupportedUnit)
}

// Factor key mappers for defaults.
func mapBusinessTravelKey(category string) string {
	switch {
	case strings.Contains(category, "flight") && strings.Contains(category, "long"):
		return "flight-long"
	case strings.Contains(category, "flight") && strings.Contains(category, "short"):
		return "flight-short"
	case strings.Contains(category, "flight"):
		return "flight-domestic"
	case strings.Contains(category, "train"):
		return "train"
	case strings.Contains(category, "taxi"), strings.Contains(category, "rideshare"):
		return "car-taxi"
	case strings.Contains(category, "bus"):
		return "bus"
	case strings.Contains(category, "hotel"):
		return "hotel-night"
	case strings.Contains(category, "car"), strings.Contains(category, "rental"):
		return "car-rental"
	default:
		return "flight-domestic"
	}
}

func mapCommutingKey(category string) string {
	switch {
	case strings.Contains(category, "electric"):
		return "car-electric"
	case strings.Contains(category, "hybrid"):
		return "car-hybrid"
	case strings.Contains(category, "diesel"):
		return "car-diesel"
	case strings.Contains(category, "train"):
		return "train"
	case strings.Contains(category, "bus"), strings.Contains(category, "transit"):
		return "public-transit"
	case strings.Contains(category, "bike"), strings.Contains(category, "bicycle"):
		return "bicycle"
	case strings.Contains(category, "walk"):
		return "walking"
	case strings.Contains(category, "wfh"), strings.Contains(category, "remote"):
		return "work-from-home"
	default:
		return "car-petrol"
	}
}

func mapWasteKey(category string) string {
	switch {
	case strings.Contains(category, "recycle") && strings.Contains(category, "paper"):
		return "recycling-paper"
	case strings.Contains(category, "recycle") && strings.Contains(category, "plastic"):
		return "recycling-plastic"
	case strings.Contains(category, "recycle") && strings.Contains(category, "metal"):
		return "recycling-metal"
	case strings.Contains(category, "recycle"):
		return "recycling-paper"
	case strings.Contains(category, "compost"):
		return "composting"
	case strings.Contains(category, "incinerat"):
		return "incineration"
	case strings.Contains(category, "organic"):
		return "landfill-organic"
	case strings.Contains(category, "landfill"):
		return "landfill-mixed"
	default:
		return "landfill-mixed"
	}
}

func mapSpendKey(category string) string {
	switch {
	case strings.Contains(category, "capital"):
		return "capital-goods"
	case strings.Contains(category, "electronic"):
		return "electronics"
	case strings.Contains(category, "software"):
		return "software"
	case strings.Contains(category, "cloud"):
		return "cloud-services"
	case strings.Contains(category, "office"):
		return "office-supplies"
	case strings.Contains(category, "furniture"):
		return "furniture"
	case strings.Contains(category, "food"), strings.Contains(category, "catering"):
		return "food-catering"
	case strings.Contains(category, "professional"), strings.Contains(category, "consulting"):
		return "professional-services"
	case strings.Contains(category, "marketing"):
		return "marketing"
	case strings.Contains(category, "construction"):
		return "construction"
	case strings.Contains(category, "processing_sold"):
		return "processing_sold_products"
	case strings.Contains(category, "use_sold"):
		return "use_sold_products"
	case strings.Contains(category, "transportation_distribution"):
		return "transportation_distribution"
	case strings.Contains(category, "investments"):
		return "investments"
	case strings.Contains(category, "franchise"):
		return "franchises"
	default:
		return "general-goods"
	}
}

func mapFuelEnergyKey(category string) string {
	switch {
	case strings.Contains(category, "t&d"), strings.Contains(category, "transmission"), strings.Contains(category, "distribution"):
		return "transmission_distribution"
	default:
		return "fuel_wtt"
	}
}

func mapTransportKey(category string) string {
	switch {
	case strings.Contains(category, "air"):
		return "air_freight"
	case strings.Contains(category, "ship"), strings.Contains(category, "ocean"), strings.Contains(category, "sea"):
		return "ship_freight"
	case strings.Contains(category, "rail"):
		return "rail_freight"
	case strings.Contains(category, "truck"), strings.Contains(category, "road"), strings.Contains(category, "freight"):
		return "truck_freight"
	default:
		return "transportation_distribution"
	}
}

// determineDataQuality assesses data quality based on method.
func (c *Scope3Calculator) determineDataQuality(method CalculationMethod) DataQuality {
	switch method {
	case MethodSupplierSpecific:
		return DataQualityMeasured
	case MethodActivityBased:
		return DataQualityEstimated
	case MethodSpendBased:
		return DataQualityDefault
	default:
		return DataQualityDefault
	}
}

// =============================================================================
// Scope 3 Reporting Utilities
// =============================================================================

// Scope3Summary aggregates Scope 3 emissions by category.
type Scope3Summary struct {
	// ByCategory maps category to total emissions in kg CO2e.
	ByCategory map[Scope3Category]float64 `json:"by_category"`

	// UpstreamKgCO2e is total for categories 1-8.
	UpstreamKgCO2e float64 `json:"upstream_kg_co2e"`

	// DownstreamKgCO2e is total for categories 9-15.
	DownstreamKgCO2e float64 `json:"downstream_kg_co2e"`

	// TotalKgCO2e is the overall Scope 3 total.
	TotalKgCO2e float64 `json:"total_kg_co2e"`

	// RecordCount is how many records were summarized.
	RecordCount int `json:"record_count"`
}

// SummarizeScope3 aggregates Scope 3 emissions by category.
func SummarizeScope3(records []EmissionRecord) Scope3Summary {
	summary := Scope3Summary{
		ByCategory: make(map[Scope3Category]float64),
	}

	for _, r := range records {
		if r.Scope != Scope3 {
			continue
		}

		summary.RecordCount++
		summary.TotalKgCO2e += r.EmissionsKgCO2e

		// Parse category from notes if available
		// This is a simplified approach; production code would
		// store category in the record itself
		category := inferCategoryFromRecord(r)

		summary.ByCategory[category] += r.EmissionsKgCO2e

		if category.IsUpstream() {
			summary.UpstreamKgCO2e += r.EmissionsKgCO2e
		} else if category.IsDownstream() {
			summary.DownstreamKgCO2e += r.EmissionsKgCO2e
		}
	}

	return summary
}

// inferCategoryFromRecord attempts to determine category from record data.
func inferCategoryFromRecord(r EmissionRecord) Scope3Category {
	notes := strings.ToLower(r.Notes)

	switch {
	case strings.Contains(notes, "travel"):
		return CategoryBusinessTravel
	case strings.Contains(notes, "commuting"):
		return CategoryCommuting
	case strings.Contains(notes, "waste"):
		return CategoryWaste
	case strings.Contains(notes, "purchased"):
		return CategoryPurchasedGoods
	case strings.Contains(notes, "capital"):
		return CategoryCapitalGoods
	default:
		return CategoryPurchasedGoods
	}
}
