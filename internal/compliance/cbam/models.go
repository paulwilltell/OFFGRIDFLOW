// Package cbam provides CBAM (Carbon Border Adjustment Mechanism) compliance
// mapping, calculation, and reporting capabilities.
//
// The CBAM is the EU's Carbon Border Adjustment Mechanism that requires importers
// to declare and pay for carbon emissions embedded in imported goods. It applies
// to specific sectors with high carbon leakage risk.
//
// Reference: EU Regulation 2023/956 (CBAM Regulation)
// Covered sectors: Cement, Electricity, Fertilizer, Iron & Steel, Aluminum, Hydrogen
package cbam

import (
	"fmt"
	"time"
)

// =============================================================================
// CBAM Covered Goods
// =============================================================================

// CommodityType represents the type of goods covered by CBAM.
type CommodityType string

const (
	CommodityCement     CommodityType = "cement"
	CommodityElectricity CommodityType = "electricity"
	CommodityFertilizer CommodityType = "fertilizer"
	CommodityIronSteel  CommodityType = "iron_steel"
	CommodityAluminum   CommodityType = "aluminum"
	CommodityHydrogen   CommodityType = "hydrogen"
)

// String returns the string representation of the commodity type.
func (c CommodityType) String() string {
	return string(c)
}

// IsValid returns true if the commodity type is recognized.
func (c CommodityType) IsValid() bool {
	switch c {
	case CommodityCement, CommodityElectricity, CommodityFertilizer,
		CommodityIronSteel, CommodityAluminum, CommodityHydrogen:
		return true
	default:
		return false
	}
}

// =============================================================================
// CBAM Input Types
// =============================================================================

// CBAMInput holds the input data for generating a CBAM declaration.
type CBAMInput struct {
	// Declarant information
	DeclarantID        string `json:"declarant_id"`
	DeclarantName      string `json:"declarant_name"`
	DeclarantEORI      string `json:"declarant_eori"` // Economic Operator Registration and Identification
	DeclarantCountry   string `json:"declarant_country"`

	// Reporting period
	QuarterYear int    `json:"quarter_year"` // e.g., 2024
	Quarter     int    `json:"quarter"`      // 1, 2, 3, or 4

	// Imported goods
	Goods []ImportedGood `json:"goods"`

	// Installation information (optional, for specific emissions)
	Installations []Installation `json:"installations,omitempty"`
}

// ImportedGood represents a specific good imported into the EU.
type ImportedGood struct {
	// Good identification
	GoodID           string        `json:"good_id"`
	CommodityType    CommodityType `json:"commodity_type"`
	CNCode           string        `json:"cn_code"` // Combined Nomenclature code
	Description      string        `json:"description"`

	// Quantity
	Quantity float64 `json:"quantity"` // In tons or MWh for electricity
	Unit     string  `json:"unit"`     // "tonnes", "MWh"

	// Origin
	CountryOfOrigin  string `json:"country_of_origin"`
	InstallationID   string `json:"installation_id,omitempty"` // ID of production installation

	// Emissions data
	EmbeddedEmissions EmbeddedEmissions `json:"embedded_emissions"`

	// Customs information
	CustomsValue     float64   `json:"customs_value"` // EUR
	ImportDate       time.Time `json:"import_date"`
	CustomsOffice    string    `json:"customs_office,omitempty"`

	// Additional data
	ProductionRoute  string    `json:"production_route,omitempty"` // e.g., "blast furnace", "electric arc furnace"
	PrecursorInfo    []Precursor `json:"precursor_info,omitempty"`
}

// EmbeddedEmissions represents the embedded emissions in an imported good.
type EmbeddedEmissions struct {
	// Direct emissions (Scope 1)
	DirectEmissions float64 `json:"direct_emissions"` // tCO2e per unit

	// Indirect emissions from electricity (part of embedded emissions)
	IndirectEmissions float64 `json:"indirect_emissions"` // tCO2e per unit

	// Total specific embedded emissions
	TotalSpecificEmissions float64 `json:"total_specific_emissions"` // tCO2e per unit

	// Calculation method
	CalculationMethod string `json:"calculation_method"` // "monitoring", "default_values", "other"

	// Monitoring plan reference (if using monitoring method)
	MonitoringPlanID string `json:"monitoring_plan_id,omitempty"`

	// Data quality
	DataQuality      string  `json:"data_quality"` // "measured", "estimated", "default"
	UncertaintyLevel float64 `json:"uncertainty_level,omitempty"` // Percentage

	// Verification status
	Verified          bool   `json:"verified"`
	VerifierName      string `json:"verifier_name,omitempty"`
	VerificationDate  time.Time `json:"verification_date,omitempty"`
}

// Installation represents a production facility where goods are produced.
type Installation struct {
	InstallationID   string `json:"installation_id"`
	InstallationName string `json:"installation_name"`
	Operator         string `json:"operator"`
	Country          string `json:"country"`
	Address          string `json:"address,omitempty"`

	// Production processes
	ProductionProcesses []ProductionProcess `json:"production_processes"`

	// Economic activity
	EconomicActivity string `json:"economic_activity"` // NACE code or description

	// Permit information
	PermitID         string    `json:"permit_id,omitempty"`
	PermitAuthority  string    `json:"permit_authority,omitempty"`
	PermitDate       time.Time `json:"permit_date,omitempty"`
}

// ProductionProcess describes a specific production process at an installation.
type ProductionProcess struct {
	ProcessID     string  `json:"process_id"`
	ProcessName   string  `json:"process_name"`
	CommodityType CommodityType `json:"commodity_type"`

	// Emissions monitoring
	DirectEmissionsFactor   float64 `json:"direct_emissions_factor"`   // tCO2e per unit of product
	IndirectEmissionsFactor float64 `json:"indirect_emissions_factor"` // tCO2e per unit

	// Production data
	AnnualProduction float64 `json:"annual_production"` // tonnes/year or MWh/year
	ProductionUnit   string  `json:"production_unit"`

	// Monitoring methodology
	MonitoringMethod string `json:"monitoring_method"` // "calculation", "measurement"
	EmissionSources  []EmissionSource `json:"emission_sources,omitempty"`
}

// EmissionSource represents a source of emissions in a production process.
type EmissionSource struct {
	SourceID          string  `json:"source_id"`
	SourceName        string  `json:"source_name"`
	SourceType        string  `json:"source_type"` // "combustion", "process", "other"
	FuelType          string  `json:"fuel_type,omitempty"`
	EmissionFactor    float64 `json:"emission_factor"` // tCO2e per unit of activity
	ActivityLevel     float64 `json:"activity_level"`  // e.g., GJ of fuel consumed
	AnnualEmissions   float64 `json:"annual_emissions"` // tCO2e/year
}

// Precursor represents a material used in producing the final good.
type Precursor struct {
	MaterialID        string  `json:"material_id"`
	MaterialName      string  `json:"material_name"`
	CNCode            string  `json:"cn_code,omitempty"`
	Quantity          float64 `json:"quantity"` // tonnes
	EmbeddedEmissions float64 `json:"embedded_emissions"` // tCO2e
	SupplierCountry   string  `json:"supplier_country,omitempty"`
}

// =============================================================================
// CBAM Declaration Types
// =============================================================================

// CBAMDeclaration represents a quarterly CBAM declaration.
type CBAMDeclaration struct {
	// Declaration metadata
	DeclarationID    string    `json:"declaration_id"`
	DeclarantID      string    `json:"declarant_id"`
	DeclarantName    string    `json:"declarant_name"`
	DeclarantEORI    string    `json:"declarant_eori"`
	Quarter          int       `json:"quarter"`
	Year             int       `json:"year"`
	SubmissionDate   time.Time `json:"submission_date"`
	DeclarationStatus string   `json:"declaration_status"` // "draft", "submitted", "accepted", "rejected"

	// Goods and emissions summary
	TotalGoods           int     `json:"total_goods"`
	TotalQuantity        float64 `json:"total_quantity"` // Sum of all goods quantities
	TotalEmbeddedEmissions float64 `json:"total_embedded_emissions"` // Total tCO2e
	TotalCustomsValue    float64 `json:"total_customs_value"` // Total EUR

	// Goods details
	GoodsEntries []GoodEntry `json:"goods_entries"`

	// Compliance metrics
	DefaultValuesUsed    int     `json:"default_values_used"`
	VerifiedEmissions    float64 `json:"verified_emissions"` // tCO2e
	UnverifiedEmissions  float64 `json:"unverified_emissions"` // tCO2e

	// Financial obligations
	EstimatedCBAMPrice    float64 `json:"estimated_cbam_price"` // EUR per tCO2e
	EstimatedCBAMCost     float64 `json:"estimated_cbam_cost"` // EUR
	CarbonPriceAdjustment float64 `json:"carbon_price_adjustment,omitempty"` // EUR (if carbon price paid in origin country)
	NetCBAMObligation     float64 `json:"net_cbam_obligation"` // EUR

	// Validation results
	ValidationResults *ValidationResults `json:"validation_results,omitempty"`
}

// GoodEntry represents a single entry in a CBAM declaration.
type GoodEntry struct {
	EntryID          string        `json:"entry_id"`
	CommodityType    CommodityType `json:"commodity_type"`
	CNCode           string        `json:"cn_code"`
	Description      string        `json:"description"`
	Quantity         float64       `json:"quantity"`
	Unit             string        `json:"unit"`
	CountryOfOrigin  string        `json:"country_of_origin"`

	// Embedded emissions
	DirectEmissions      float64 `json:"direct_emissions"`      // tCO2e
	IndirectEmissions    float64 `json:"indirect_emissions"`    // tCO2e
	TotalEmissions       float64 `json:"total_emissions"`       // tCO2e
	SpecificEmissions    float64 `json:"specific_emissions"`    // tCO2e per unit

	// Calculation details
	CalculationMethod    string  `json:"calculation_method"`
	UsedDefaultValues    bool    `json:"used_default_values"`
	Verified             bool    `json:"verified"`

	// Financial
	CustomsValue         float64 `json:"customs_value"` // EUR
	EstimatedCBAMCost    float64 `json:"estimated_cbam_cost"` // EUR

	// References
	InstallationID       string  `json:"installation_id,omitempty"`
	ImportDate           time.Time `json:"import_date"`
}

// =============================================================================
// Default Emission Values
// =============================================================================

// DefaultEmissionValue represents default emission values for goods without specific data.
type DefaultEmissionValue struct {
	CommodityType     CommodityType `json:"commodity_type"`
	CNCode            string        `json:"cn_code,omitempty"`
	CountryOfOrigin   string        `json:"country_of_origin,omitempty"`
	ProductionRoute   string        `json:"production_route,omitempty"`

	// Default values
	DefaultDirectEmissions   float64 `json:"default_direct_emissions"`   // tCO2e per unit
	DefaultIndirectEmissions float64 `json:"default_indirect_emissions"` // tCO2e per unit
	DefaultTotalEmissions    float64 `json:"default_total_emissions"`    // tCO2e per unit

	// Source
	ValueSource      string    `json:"value_source"` // e.g., "EU Commission Implementing Regulation"
	EffectiveDate    time.Time `json:"effective_date"`
	ExpiryDate       time.Time `json:"expiry_date,omitempty"`
}

// =============================================================================
// Validation Types
// =============================================================================

// ValidationResults contains declaration validation outcomes.
type ValidationResults struct {
	Valid      bool                `json:"valid"`
	Errors     []ValidationError   `json:"errors,omitempty"`
	Warnings   []ValidationWarning `json:"warnings,omitempty"`
	InfoMessages []string          `json:"info_messages,omitempty"`
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field      string `json:"field"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Severity   string `json:"severity"` // "error", "critical"
	EntryID    string `json:"entry_id,omitempty"` // If related to specific good entry
}

// ValidationWarning represents a validation warning.
type ValidationWarning struct {
	Field    string `json:"field"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	EntryID  string `json:"entry_id,omitempty"`
}

// =============================================================================
// CBAM Report
// =============================================================================

// CBAMReport represents a complete CBAM compliance report.
type CBAMReport struct {
	// Report metadata
	ReportID       string    `json:"report_id"`
	GeneratedAt    time.Time `json:"generated_at"`
	ReportingPeriod string   `json:"reporting_period"` // e.g., "Q1 2024"

	// Declarant
	Declarant DeclarantInfo `json:"declarant"`

	// Declaration summary
	Declaration CBAMDeclaration `json:"declaration"`

	// Compliance status
	ComplianceStatus string  `json:"compliance_status"` // "compliant", "non_compliant", "pending"
	ComplianceScore  float64 `json:"compliance_score"`  // 0-100

	// Recommendations
	Recommendations []string `json:"recommendations,omitempty"`

	// Next steps
	NextSteps []string `json:"next_steps,omitempty"`
}

// DeclarantInfo holds declarant information.
type DeclarantInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	EORI        string `json:"eori"`
	Country     string `json:"country"`
	Address     string `json:"address,omitempty"`
	ContactName string `json:"contact_name,omitempty"`
	ContactEmail string `json:"contact_email,omitempty"`
}

// =============================================================================
// Quarter Helper Functions
// =============================================================================

// QuarterInfo represents a calendar quarter.
type QuarterInfo struct {
	Year    int
	Quarter int
	Start   time.Time
	End     time.Time
}

// GetQuarter returns the quarter information for a given date.
func GetQuarter(date time.Time) QuarterInfo {
	year := date.Year()
	month := date.Month()

	var quarter int
	var start, end time.Time

	switch {
	case month <= 3:
		quarter = 1
		start = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		end = time.Date(year, 3, 31, 23, 59, 59, 0, time.UTC)
	case month <= 6:
		quarter = 2
		start = time.Date(year, 4, 1, 0, 0, 0, 0, time.UTC)
		end = time.Date(year, 6, 30, 23, 59, 59, 0, time.UTC)
	case month <= 9:
		quarter = 3
		start = time.Date(year, 7, 1, 0, 0, 0, 0, time.UTC)
		end = time.Date(year, 9, 30, 23, 59, 59, 0, time.UTC)
	default:
		quarter = 4
		start = time.Date(year, 10, 1, 0, 0, 0, 0, time.UTC)
		end = time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)
	}

	return QuarterInfo{
		Year:    year,
		Quarter: quarter,
		Start:   start,
		End:     end,
	}
}

// FormatQuarter returns a formatted quarter string (e.g., "Q1 2024").
func FormatQuarter(year, quarter int) string {
	return fmt.Sprintf("Q%d %d", quarter, year)
}
