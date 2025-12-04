// Package ingestion provides data ingestion models and interfaces for
// importing energy and activity data from various sources into OffGridFlow.
//
// The package defines the Activity type, which is the canonical representation
// of energy consumption or other emissions-related activities that flow through
// the system for calculation and reporting.
package ingestion

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// =============================================================================
// Source Type Constants
// =============================================================================

// Source identifies the origin of activity data.
type Source string

const (
	// SourceUtilityBill represents electricity or gas utility bills.
	SourceUtilityBill Source = "utility_bill"

	// SourceFleet represents vehicle fleet fuel consumption.
	SourceFleet Source = "fleet"

	// SourceTravel represents business travel (flights, rail, etc.).
	SourceTravel Source = "travel"

	// SourcePurchases represents purchased goods and services.
	SourcePurchases Source = "purchases"

	// SourceWaste represents waste disposal and treatment.
	SourceWaste Source = "waste"

	// SourceManual represents manually entered data.
	SourceManual Source = "manual"

	// SourceAPI represents data from third-party API integrations.
	SourceAPI Source = "api"
)

// String returns the string representation of the source.
func (s Source) String() string {
	return string(s)
}

// IsValid returns true if the source is a recognized value.
func (s Source) IsValid() bool {
	switch s {
	case SourceUtilityBill, SourceFleet, SourceTravel,
		SourcePurchases, SourceWaste, SourceManual, SourceAPI:
		return true
	default:
		return false
	}
}

// =============================================================================
// Unit Type Constants
// =============================================================================

// Unit represents the unit of measurement for activity quantities.
type Unit string

const (
	// Energy units
	UnitKWh   Unit = "kWh"   // Kilowatt-hours
	UnitMWh   Unit = "MWh"   // Megawatt-hours
	UnitGJ    Unit = "GJ"    // Gigajoules
	UnitTherm Unit = "therm" // Therms (natural gas)

	// Volume units
	UnitLiter  Unit = "L"      // Liters
	UnitGallon Unit = "gallon" // US Gallons
	UnitCubicM Unit = "m3"     // Cubic meters

	// Mass units
	UnitKg    Unit = "kg"    // Kilograms
	UnitTonne Unit = "tonne" // Metric tonnes

	// Distance units
	UnitKm   Unit = "km"   // Kilometers
	UnitMile Unit = "mile" // Miles

	// Currency units (for spend-based calculations)
	UnitUSD Unit = "USD"
	UnitEUR Unit = "EUR"
	UnitGBP Unit = "GBP"
)

// String returns the string representation of the unit.
func (u Unit) String() string {
	return string(u)
}

// IsEnergyUnit returns true if the unit is an energy unit.
func (u Unit) IsEnergyUnit() bool {
	switch u {
	case UnitKWh, UnitMWh, UnitGJ, UnitTherm:
		return true
	default:
		return false
	}
}

// =============================================================================
// Sentinel Errors
// =============================================================================

var (
	// ErrEmptyID is returned when an activity has an empty ID.
	ErrEmptyID = errors.New("ingestion: activity ID cannot be empty")

	// ErrEmptySource is returned when an activity has no source.
	ErrEmptySource = errors.New("ingestion: activity source cannot be empty")

	// ErrInvalidQuantity is returned when quantity is negative.
	ErrInvalidQuantity = errors.New("ingestion: quantity cannot be negative")

	// ErrEmptyUnit is returned when an activity has no unit.
	ErrEmptyUnit = errors.New("ingestion: unit cannot be empty")

	// ErrInvalidPeriod is returned when the period dates are invalid.
	ErrInvalidPeriod = errors.New("ingestion: period end must be after period start")

	// ErrEmptyOrgID is returned when organization ID is missing.
	ErrEmptyOrgID = errors.New("ingestion: organization ID cannot be empty")
)

// =============================================================================
// Activity Model
// =============================================================================

// Activity represents an ingested activity that will be translated into emissions.
// Activities are the raw data points that feed into emissions calculations.
//
// An Activity captures:
//   - What was consumed (Quantity, Unit)
//   - Where it was consumed (Location/Region)
//   - When it was consumed (PeriodStart, PeriodEnd)
//   - Where the data came from (Source)
//   - Who it belongs to (OrgID, WorkspaceID)
type Activity struct {
	// ID is a unique identifier for this activity record.
	ID string `json:"id"`

	// Source identifies where this data came from.
	// Examples: "utility_bill", "fleet", "travel", "purchases"
	Source string `json:"source"`

	// Category provides optional sub-categorization within the source.
	// Examples: "electricity", "natural_gas", "diesel", "gasoline"
	Category string `json:"category,omitempty"`

	// MeterID identifies the specific meter for utility bills.
	// This enables tracking consumption by physical meter.
	MeterID string `json:"meter_id,omitempty"`

	// Location is the geographic region code for emission factor lookup.
	// Examples: "US-WEST", "EU-CENTRAL", "ASIA-PACIFIC"
	// This should match region codes in the emission factor registry.
	Location string `json:"location"`

	// PeriodStart is the beginning of the measurement period (inclusive).
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the measurement period (exclusive).
	PeriodEnd time.Time `json:"period_end"`

	// Quantity is the amount consumed in the specified Unit.
	// Must be non-negative.
	Quantity float64 `json:"quantity"`

	// Unit is the unit of measurement for Quantity.
	// Examples: "kWh", "L", "kg", "km"
	Unit string `json:"unit"`

	// OrgID identifies the organization this activity belongs to.
	OrgID string `json:"org_id"`

	// WorkspaceID identifies the workspace for multi-tenant scenarios.
	WorkspaceID string `json:"workspace_id,omitempty"`

	// Metadata contains additional key-value data for extensibility.
	// Examples: invoice numbers, supplier info, notes
	Metadata map[string]string `json:"metadata,omitempty"`

	// CreatedAt is when this activity record was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this activity record was last modified.
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// ImportBatchID groups activities from the same import operation.
	ImportBatchID string `json:"import_batch_id,omitempty"`

	// ExternalID is an identifier from the source system for deduplication.
	ExternalID string `json:"external_id,omitempty"`

	// IdempotencyKey ensures reliable cloud ingestion - prevents duplicate processing
	// of the same activity across retries. Format: "{source}:{external_id}:{period}"
	IdempotencyKey string `json:"idempotency_key,omitempty"`

	// IngestionAttempts tracks retry count for observability
	IngestionAttempts int `json:"ingestion_attempts,omitempty"`

	// LastIngestionError captures the last error for debugging
	LastIngestionError string `json:"last_ingestion_error,omitempty"`

	// DataQuality indicates the confidence level of this data.
	// Values: "measured", "estimated", "default"
	DataQuality string `json:"data_quality,omitempty"`
}

// Validate checks that the activity has all required fields and valid values.
func (a Activity) Validate() error {
	var errs []error

	if strings.TrimSpace(a.ID) == "" {
		errs = append(errs, ErrEmptyID)
	}

	if strings.TrimSpace(a.Source) == "" {
		errs = append(errs, ErrEmptySource)
	}

	if a.Quantity < 0 {
		errs = append(errs, ErrInvalidQuantity)
	}

	if strings.TrimSpace(a.Unit) == "" {
		errs = append(errs, ErrEmptyUnit)
	}

	if !a.PeriodEnd.IsZero() && !a.PeriodStart.IsZero() {
		if a.PeriodEnd.Before(a.PeriodStart) {
			errs = append(errs, ErrInvalidPeriod)
		}
	}

	if strings.TrimSpace(a.OrgID) == "" {
		errs = append(errs, ErrEmptyOrgID)
	}

	if len(errs) > 0 {
		return fmt.Errorf("activity validation failed: %w", errors.Join(errs...))
	}

	return nil
}

// Duration returns the time span of the activity period.
func (a Activity) Duration() time.Duration {
	if a.PeriodEnd.IsZero() || a.PeriodStart.IsZero() {
		return 0
	}
	return a.PeriodEnd.Sub(a.PeriodStart)
}

// DurationDays returns the period duration in days.
func (a Activity) DurationDays() float64 {
	return a.Duration().Hours() / 24.0
}

// IsUtilityBill returns true if this is a utility bill activity.
func (a Activity) IsUtilityBill() bool {
	return a.Source == string(SourceUtilityBill)
}

// IsElectricity returns true if this is an electricity activity.
func (a Activity) IsElectricity() bool {
	return a.IsUtilityBill() && (a.Unit == string(UnitKWh) || a.Unit == string(UnitMWh))
}

// Clone creates a deep copy of the activity.
func (a Activity) Clone() Activity {
	clone := a

	// Deep copy metadata
	if a.Metadata != nil {
		clone.Metadata = make(map[string]string, len(a.Metadata))
		for k, v := range a.Metadata {
			clone.Metadata[k] = v
		}
	}

	return clone
}

// WithMetadata returns a copy with additional metadata.
func (a Activity) WithMetadata(key, value string) Activity {
	clone := a.Clone()
	if clone.Metadata == nil {
		clone.Metadata = make(map[string]string)
	}
	clone.Metadata[key] = value
	return clone
}

// JSON serializes the activity to JSON bytes.
func (a Activity) JSON() ([]byte, error) {
	return json.Marshal(a)
}

// =============================================================================
// emissions.Activity Interface Implementation
// =============================================================================
//
// These methods implement the emissions.Activity interface, allowing Activity
// to be passed directly to emissions calculators without adaptation.

// GetID returns the activity's unique identifier.
func (a Activity) GetID() string { return a.ID }

// GetSource returns the activity source (e.g., "utility_bill").
func (a Activity) GetSource() string { return a.Source }

// GetCategory returns the activity category (e.g., "electricity").
func (a Activity) GetCategory() string { return a.Category }

// GetLocation returns the geographic region code for emission factor lookup.
func (a Activity) GetLocation() string { return a.Location }

// GetQuantity returns the consumption quantity.
func (a Activity) GetQuantity() float64 { return a.Quantity }

// GetUnit returns the unit of measurement (e.g., "kWh").
func (a Activity) GetUnit() string { return a.Unit }

// GetPeriodStart returns the period start time.
func (a Activity) GetPeriodStart() time.Time { return a.PeriodStart }

// GetPeriodEnd returns the period end time.
func (a Activity) GetPeriodEnd() time.Time { return a.PeriodEnd }

// GetOrgID returns the organization identifier.
func (a Activity) GetOrgID() string { return a.OrgID }

// GetWorkspaceID returns the workspace identifier.
func (a Activity) GetWorkspaceID() string { return a.WorkspaceID }

// =============================================================================
// Activity Builder
// =============================================================================

// ActivityBuilder provides a fluent interface for constructing activities.
type ActivityBuilder struct {
	activity Activity
}

// NewActivityBuilder creates a new activity builder.
func NewActivityBuilder() *ActivityBuilder {
	return &ActivityBuilder{
		activity: Activity{
			CreatedAt: time.Now().UTC(),
			Metadata:  make(map[string]string),
		},
	}
}

// WithID sets the activity ID.
func (b *ActivityBuilder) WithID(id string) *ActivityBuilder {
	b.activity.ID = id
	return b
}

// WithSource sets the activity source.
func (b *ActivityBuilder) WithSource(source string) *ActivityBuilder {
	b.activity.Source = source
	return b
}

// WithCategory sets the activity category.
func (b *ActivityBuilder) WithCategory(category string) *ActivityBuilder {
	b.activity.Category = category
	return b
}

// WithMeterID sets the meter ID.
func (b *ActivityBuilder) WithMeterID(meterID string) *ActivityBuilder {
	b.activity.MeterID = meterID
	return b
}

// WithLocation sets the location/region.
func (b *ActivityBuilder) WithLocation(location string) *ActivityBuilder {
	b.activity.Location = location
	return b
}

// WithPeriod sets the period start and end.
func (b *ActivityBuilder) WithPeriod(start, end time.Time) *ActivityBuilder {
	b.activity.PeriodStart = start
	b.activity.PeriodEnd = end
	return b
}

// WithQuantity sets the quantity and unit.
func (b *ActivityBuilder) WithQuantity(quantity float64, unit string) *ActivityBuilder {
	b.activity.Quantity = quantity
	b.activity.Unit = unit
	return b
}

// WithOrgID sets the organization ID.
func (b *ActivityBuilder) WithOrgID(orgID string) *ActivityBuilder {
	b.activity.OrgID = orgID
	return b
}

// WithWorkspaceID sets the workspace ID.
func (b *ActivityBuilder) WithWorkspaceID(workspaceID string) *ActivityBuilder {
	b.activity.WorkspaceID = workspaceID
	return b
}

// WithMetadata adds a metadata key-value pair.
func (b *ActivityBuilder) WithMetadata(key, value string) *ActivityBuilder {
	b.activity.Metadata[key] = value
	return b
}

// Build returns the constructed activity after validation.
func (b *ActivityBuilder) Build() (Activity, error) {
	b.activity.UpdatedAt = time.Now().UTC()

	if err := b.activity.Validate(); err != nil {
		return Activity{}, err
	}

	return b.activity, nil
}

// MustBuild returns the constructed activity, panicking on validation error.
func (b *ActivityBuilder) MustBuild() Activity {
	a, err := b.Build()
	if err != nil {
		panic(err)
	}
	return a
}

// =============================================================================
// Batch Processing Types
// =============================================================================

// ImportBatch represents a group of activities imported together.
type ImportBatch struct {
	// ID is a unique identifier for this import batch.
	ID string `json:"id"`

	// Source identifies where this batch came from.
	Source string `json:"source"`

	// FileName is the original file name if imported from a file.
	FileName string `json:"file_name,omitempty"`

	// ActivityCount is the number of activities in this batch.
	ActivityCount int `json:"activity_count"`

	// SuccessCount is how many activities were successfully processed.
	SuccessCount int `json:"success_count"`

	// ErrorCount is how many activities failed processing.
	ErrorCount int `json:"error_count"`

	// Errors contains error messages for failed activities.
	Errors []ImportError `json:"errors,omitempty"`

	// CreatedAt is when this batch was created.
	CreatedAt time.Time `json:"created_at"`

	// CompletedAt is when batch processing finished.
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// OrgID is the organization this batch belongs to.
	OrgID string `json:"org_id"`

	// UserID is the user who initiated the import.
	UserID string `json:"user_id,omitempty"`
}

// ImportError records a single import error.
type ImportError struct {
	// Row is the row number (1-indexed) where the error occurred.
	Row int `json:"row,omitempty"`

	// Field is the field name that caused the error.
	Field string `json:"field,omitempty"`

	// Message describes what went wrong.
	Message string `json:"message"`

	// ExternalID is the external identifier of the failed record.
	ExternalID string `json:"external_id,omitempty"`
}

// IsComplete returns true if the batch has finished processing.
func (b ImportBatch) IsComplete() bool {
	return b.CompletedAt != nil
}

// HasErrors returns true if any activities failed to process.
func (b ImportBatch) HasErrors() bool {
	return b.ErrorCount > 0
}

// SuccessRate returns the percentage of successful imports (0.0 to 1.0).
func (b ImportBatch) SuccessRate() float64 {
	if b.ActivityCount == 0 {
		return 1.0
	}
	return float64(b.SuccessCount) / float64(b.ActivityCount)
}
