// Package emissions provides carbon emissions calculation and tracking capabilities.
//
// The package implements the Greenhouse Gas Protocol scopes:
//   - Scope 1: Direct emissions from owned or controlled sources
//   - Scope 2: Indirect emissions from purchased energy
//   - Scope 3: All other indirect emissions in the value chain
//
// Key types:
//   - EmissionFactor: Conversion factors for activities to CO2e emissions
//   - EmissionRecord: Calculated emission result with full traceability
//   - Calculator: Interface for scope-specific calculation logic
//   - Registry: Interface for emission factor lookup
package emissions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// =============================================================================
// Scope Constants
// =============================================================================

// Scope represents the GHG Protocol emission scope classification.
type Scope int

const (
	// ScopeUnspecified indicates no scope was set.
	ScopeUnspecified Scope = 0

	// Scope1 represents direct emissions from owned or controlled sources.
	// Examples: company vehicles, on-site fuel combustion, fugitive emissions.
	Scope1 Scope = 1

	// Scope2 represents indirect emissions from purchased energy.
	// Examples: purchased electricity, steam, heating, cooling.
	Scope2 Scope = 2

	// Scope3 represents all other indirect emissions in the value chain.
	// Examples: business travel, employee commuting, purchased goods.
	Scope3 Scope = 3
)

// String returns the human-readable scope name.
func (s Scope) String() string {
	switch s {
	case Scope1:
		return "Scope 1"
	case Scope2:
		return "Scope 2"
	case Scope3:
		return "Scope 3"
	default:
		return "Unknown"
	}
}

// IsValid returns true if the scope is a valid GHG Protocol scope.
func (s Scope) IsValid() bool {
	return s >= Scope1 && s <= Scope3
}

// =============================================================================
// Data Quality Constants
// =============================================================================

// DataQuality indicates the reliability of emission data.
type DataQuality string

const (
	// DataQualityMeasured indicates data from direct measurements.
	DataQualityMeasured DataQuality = "measured"

	// DataQualityEstimated indicates data derived from estimates.
	DataQualityEstimated DataQuality = "estimated"

	// DataQualityDefault indicates default/fallback values were used.
	DataQualityDefault DataQuality = "default"
)

// =============================================================================
// Calculation Method Constants
// =============================================================================

// CalculationMethod identifies how emissions were calculated.
type CalculationMethod string

const (
	// MethodLocationBased uses average grid emission factors.
	// This is the standard Scope 2 method using regional grid averages.
	MethodLocationBased CalculationMethod = "location-based"

	// MethodMarketBased uses supplier-specific emission factors.
	// This accounts for renewable energy purchases and contracts.
	MethodMarketBased CalculationMethod = "market-based"

	// MethodSpendBased calculates from monetary spend data.
	// Used when quantity data is unavailable.
	MethodSpendBased CalculationMethod = "spend-based"

	// MethodActivityBased uses activity-specific factors.
	// Standard method using quantity × emission factor.
	MethodActivityBased CalculationMethod = "activity-based"

	// MethodSupplierSpecific uses actual supplier-provided data.
	// Highest accuracy for Scope 3 calculations.
	MethodSupplierSpecific CalculationMethod = "supplier-specific"
)

// =============================================================================
// Sentinel Errors
// =============================================================================

var (
	// ErrFactorNotFound is returned when no emission factor exists.
	ErrFactorNotFound = errors.New("emissions: emission factor not found")

	// ErrInvalidScope is returned when an invalid scope is provided.
	ErrInvalidScope = errors.New("emissions: invalid emission scope")

	// ErrInvalidRegion is returned when the region code is empty or invalid.
	ErrInvalidRegion = errors.New("emissions: invalid or empty region code")

	// ErrInvalidSource is returned when the activity source is invalid.
	ErrInvalidSource = errors.New("emissions: invalid activity source")

	// ErrCalculationFailed is returned when emission calculation fails.
	ErrCalculationFailed = errors.New("emissions: calculation failed")

	// ErrUnsupportedUnit is returned when the activity unit is not supported.
	ErrUnsupportedUnit = errors.New("emissions: unsupported unit for calculation")

	// ErrNilActivity is returned when a nil activity is provided.
	ErrNilActivity = errors.New("emissions: activity cannot be nil")

	// ErrRegistryNotAvailable is returned when the factor registry is unavailable.
	ErrRegistryNotAvailable = errors.New("emissions: factor registry not available")
)

// =============================================================================
// EmissionFactor Model
// =============================================================================

// EmissionFactor represents a greenhouse gas emission conversion factor.
//
// Emission factors translate activity data (e.g., kWh of electricity) into
// CO2 equivalent emissions. Factors are region-specific and may be updated
// annually as grid composition changes.
//
// The GHG Protocol recommends using location-based factors for Scope 2
// unless organizations have contractual instruments for renewable energy.
type EmissionFactor struct {
	// ID is a unique identifier for this factor.
	ID string `json:"id"`

	// Scope indicates which GHG Protocol scope this factor applies to.
	Scope Scope `json:"scope"`

	// Region is the geographic region code (e.g., "US-WEST", "EU-CENTRAL").
	// Should match region codes used in activity location fields.
	Region string `json:"region"`

	// Source identifies the activity source this factor applies to.
	// Examples: "utility_bill", "fleet", "travel"
	Source string `json:"source"`

	// Category provides sub-categorization within the source.
	// Examples: "electricity", "natural_gas", "diesel"
	Category string `json:"category,omitempty"`

	// Unit is the activity unit this factor converts from.
	// Examples: "kWh", "L", "km"
	Unit string `json:"unit"`

	// ValueKgCO2ePerUnit is the emission factor in kg CO2e per activity unit.
	// This is the core conversion value.
	ValueKgCO2ePerUnit float64 `json:"value_kg_co2e_per_unit"`

	// Method indicates the calculation methodology.
	Method CalculationMethod `json:"method,omitempty"`

	// DataSource documents where this factor came from.
	// Examples: "EPA eGRID 2023", "DEFRA 2024", "IEA 2023"
	DataSource string `json:"data_source,omitempty"`

	// ValidFrom is when this factor becomes effective.
	ValidFrom time.Time `json:"valid_from,omitempty"`

	// ValidTo is when this factor expires (zero means currently valid).
	ValidTo time.Time `json:"valid_to,omitempty"`

	// UncertaintyPercent is the uncertainty range as a percentage.
	// For example, 5.0 means ±5% uncertainty.
	UncertaintyPercent float64 `json:"uncertainty_percent,omitempty"`

	// CreatedAt is when this factor was created in the system.
	CreatedAt time.Time `json:"created_at,omitempty"`

	// UpdatedAt is when this factor was last modified.
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// Notes contains additional context or caveats.
	Notes string `json:"notes,omitempty"`
}

// IsValid checks if the emission factor has required fields.
func (f EmissionFactor) IsValid() bool {
	return f.ID != "" &&
		f.Scope.IsValid() &&
		f.Region != "" &&
		f.Source != "" &&
		f.Unit != "" &&
		f.ValueKgCO2ePerUnit >= 0
}

// IsCurrentlyValid returns true if the factor is valid at the given time.
func (f EmissionFactor) IsCurrentlyValid(at time.Time) bool {
	if !f.ValidFrom.IsZero() && at.Before(f.ValidFrom) {
		return false
	}
	if !f.ValidTo.IsZero() && at.After(f.ValidTo) {
		return false
	}
	return true
}

// CalculateEmissions applies this factor to a quantity.
func (f EmissionFactor) CalculateEmissions(quantity float64) float64 {
	return quantity * f.ValueKgCO2ePerUnit
}

// String returns a human-readable representation.
func (f EmissionFactor) String() string {
	return fmt.Sprintf(
		"EmissionFactor{id=%s, scope=%s, region=%s, value=%.4f kgCO2e/%s}",
		f.ID, f.Scope, f.Region, f.ValueKgCO2ePerUnit, f.Unit,
	)
}

// =============================================================================
// EmissionRecord Model
// =============================================================================

// EmissionRecord represents a calculated emission result.
//
// Each record links back to the source activity and emission factor used,
// providing full traceability for auditing and verification purposes.
type EmissionRecord struct {
	// ID is a unique identifier for this emission record.
	ID string `json:"id"`

	// ActivityID links to the source activity.
	ActivityID string `json:"activity_id"`

	// FactorID links to the emission factor used.
	FactorID string `json:"factor_id"`

	// Scope is the GHG Protocol scope classification.
	Scope Scope `json:"scope"`

	// EmissionsKgCO2e is the calculated emissions in kilograms of CO2 equivalent.
	EmissionsKgCO2e float64 `json:"emissions_kg_co2e"`

	// EmissionsTonnesCO2e is the same value in metric tonnes.
	EmissionsTonnesCO2e float64 `json:"emissions_tonnes_co2e"`

	// InputQuantity is the activity quantity used in calculation.
	InputQuantity float64 `json:"input_quantity"`

	// InputUnit is the unit of the input quantity.
	InputUnit string `json:"input_unit"`

	// EmissionFactor is the factor value applied.
	EmissionFactor float64 `json:"emission_factor"`

	// Method indicates the calculation methodology used.
	Method CalculationMethod `json:"method"`

	// DataQuality indicates the reliability of the result.
	DataQuality DataQuality `json:"data_quality"`

	// Region is the geographic region for this emission.
	Region string `json:"region"`

	// OrgID identifies the organization.
	OrgID string `json:"org_id"`

	// WorkspaceID identifies the workspace.
	WorkspaceID string `json:"workspace_id,omitempty"`

	// PeriodStart is the beginning of the emission period.
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the emission period.
	PeriodEnd time.Time `json:"period_end"`

	// CalculatedAt is when this record was calculated.
	CalculatedAt time.Time `json:"calculated_at"`

	// Notes contains any calculation notes or warnings.
	Notes string `json:"notes,omitempty"`
}

// Validate checks that the emission record has valid values.
func (r EmissionRecord) Validate() error {
	var errs []error

	if r.ID == "" {
		errs = append(errs, errors.New("emission record ID is required"))
	}

	if r.ActivityID == "" {
		errs = append(errs, errors.New("activity ID is required"))
	}

	if !r.Scope.IsValid() {
		errs = append(errs, ErrInvalidScope)
	}

	if r.EmissionsKgCO2e < 0 {
		errs = append(errs, errors.New("emissions cannot be negative"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// String returns a human-readable representation.
func (r EmissionRecord) String() string {
	return fmt.Sprintf(
		"EmissionRecord{id=%s, scope=%s, emissions=%.4f kgCO2e}",
		r.ID, r.Scope, r.EmissionsKgCO2e,
	)
}

// =============================================================================
// Interfaces
// =============================================================================

// Activity defines the minimal interface for activity data needed for emissions
// calculations. This allows the emissions package to work with different
// activity implementations.
type Activity interface {
	// GetID returns the activity's unique identifier.
	GetID() string

	// GetSource returns the activity source (e.g., "utility_bill").
	GetSource() string

	// GetCategory returns the optional category (e.g., "electricity").
	GetCategory() string

	// GetLocation returns the geographic region code.
	GetLocation() string

	// GetQuantity returns the consumption quantity.
	GetQuantity() float64

	// GetUnit returns the unit of measurement.
	GetUnit() string

	// GetPeriodStart returns the period start time.
	GetPeriodStart() time.Time

	// GetPeriodEnd returns the period end time.
	GetPeriodEnd() time.Time

	// GetOrgID returns the organization identifier.
	GetOrgID() string

	// GetWorkspaceID returns the optional workspace identifier.
	GetWorkspaceID() string
}

// FactorRegistry provides access to emission factors.
//
// Implementations may use in-memory storage, databases, or external APIs.
// The registry should handle factor versioning and validity periods.
type FactorRegistry interface {
	// GetFactor retrieves a factor by its unique ID.
	GetFactor(ctx context.Context, id string) (EmissionFactor, error)

	// FindFactor looks up the best matching factor for an activity.
	// The matching considers scope, region, source, category, and unit.
	FindFactor(ctx context.Context, query FactorQuery) (EmissionFactor, error)

	// ListFactors returns all factors matching the given criteria.
	ListFactors(ctx context.Context, query FactorQuery) ([]EmissionFactor, error)

	// RegisterFactor adds or updates a factor in the registry.
	RegisterFactor(ctx context.Context, factor EmissionFactor) error
}

// FactorQuery specifies criteria for finding emission factors.
type FactorQuery struct {
	// Scope filters by GHG Protocol scope.
	Scope Scope `json:"scope,omitempty"`

	// Region filters by geographic region code.
	Region string `json:"region,omitempty"`

	// Source filters by activity source.
	Source string `json:"source,omitempty"`

	// Category filters by activity category.
	Category string `json:"category,omitempty"`

	// Unit filters by activity unit.
	Unit string `json:"unit,omitempty"`

	// ValidAt filters for factors valid at this time.
	ValidAt time.Time `json:"valid_at,omitempty"`
}

// Matches returns true if the factor matches the query criteria.
func (q FactorQuery) Matches(f EmissionFactor) bool {
	// Check scope if specified
	if q.Scope != ScopeUnspecified && f.Scope != q.Scope {
		return false
	}

	// Check region if specified (case-insensitive)
	if q.Region != "" && !strings.EqualFold(f.Region, q.Region) {
		return false
	}

	// Check source if specified (case-insensitive)
	if q.Source != "" && !strings.EqualFold(f.Source, q.Source) {
		return false
	}

	// Check category if specified (case-insensitive)
	if q.Category != "" && !strings.EqualFold(f.Category, q.Category) {
		return false
	}

	// Check unit if specified
	if q.Unit != "" && f.Unit != q.Unit {
		return false
	}

	// Check time validity
	if !q.ValidAt.IsZero() && !f.IsCurrentlyValid(q.ValidAt) {
		return false
	}

	return true
}

// Calculator computes emissions from activities.
//
// Different implementations handle different scopes and calculation methods.
// Calculators should be stateless and thread-safe.
type Calculator interface {
	// Calculate computes emissions for the given activity.
	// Returns an EmissionRecord with full calculation details.
	Calculate(ctx context.Context, activity Activity) (EmissionRecord, error)

	// CalculateBatch processes multiple activities.
	// Returns results for all successfully calculated activities.
	CalculateBatch(ctx context.Context, activities []Activity) ([]EmissionRecord, error)

	// Supports returns true if this calculator can handle the activity.
	Supports(activity Activity) bool
}

// =============================================================================
// Result Types
// =============================================================================

// CalculationResult wraps an EmissionRecord with any errors or warnings.
type CalculationResult struct {
	// Record is the calculated emission record (nil if error occurred).
	Record *EmissionRecord `json:"record,omitempty"`

	// ActivityID identifies the source activity.
	ActivityID string `json:"activity_id"`

	// Error contains any error that occurred during calculation.
	Error error `json:"error,omitempty"`

	// Warnings contains non-fatal issues encountered.
	Warnings []string `json:"warnings,omitempty"`
}

// IsSuccess returns true if calculation succeeded without errors.
func (r CalculationResult) IsSuccess() bool {
	return r.Error == nil && r.Record != nil
}

// BatchResult summarizes results from batch calculation.
type BatchResult struct {
	// Records contains all successfully calculated emissions.
	Records []EmissionRecord `json:"records"`

	// Results contains individual results including errors.
	Results []CalculationResult `json:"results"`

	// TotalEmissionsKgCO2e is the sum of all calculated emissions.
	TotalEmissionsKgCO2e float64 `json:"total_emissions_kg_co2e"`

	// SuccessCount is how many activities were successfully calculated.
	SuccessCount int `json:"success_count"`

	// ErrorCount is how many activities failed calculation.
	ErrorCount int `json:"error_count"`

	// ProcessedAt is when the batch was processed.
	ProcessedAt time.Time `json:"processed_at"`
}

// SuccessRate returns the percentage of successful calculations (0.0 to 1.0).
func (r BatchResult) SuccessRate() float64 {
	total := r.SuccessCount + r.ErrorCount
	if total == 0 {
		return 1.0
	}
	return float64(r.SuccessCount) / float64(total)
}
