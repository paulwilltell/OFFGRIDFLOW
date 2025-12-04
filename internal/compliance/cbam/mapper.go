package cbam

import (
	"context"
	"fmt"
	"time"

	"github.com/example/offgridflow/internal/compliance/core"
)

// =============================================================================
// Mapper
// =============================================================================

// Mapper handles CBAM compliance reporting and declaration generation.
// Implements the core.ComplianceMapper interface.
type Mapper struct {
	calculator *Calculator
}

// NewMapper creates a new CBAM mapper with a calculator.
func NewMapper() *Mapper {
	return &Mapper{
		calculator: NewCalculator(),
	}
}

// BuildReport implements the ComplianceMapper interface for CBAM declarations.
func (m *Mapper) BuildReport(ctx context.Context, input core.ComplianceInput) (core.ComplianceReport, error) {
	// Coerce input to CBAM-specific input
	cbamInput, err := m.coerceInput(input)
	if err != nil {
		return core.ComplianceReport{}, fmt.Errorf("invalid CBAM input: %w", err)
	}

	// Calculate embedded emissions for all goods
	if err := m.calculateAllEmissions(ctx, cbamInput); err != nil {
		return core.ComplianceReport{}, fmt.Errorf("emissions calculation failed: %w", err)
	}

	// Generate CBAM declaration
	declaration, err := m.generateDeclaration(cbamInput)
	if err != nil {
		return core.ComplianceReport{}, fmt.Errorf("declaration generation failed: %w", err)
	}

	// Validate declaration
	validation := m.validateDeclaration(declaration)
	declaration.ValidationResults = validation

	// Build CBAM report
	report := m.buildCBAMReport(cbamInput, declaration)

	// Map to core.ComplianceReport with ValidationResults
	// Convert CBAM ValidationResults to core.ValidationResult
	coreValidation := make([]core.ValidationResult, 0)
	
	for _, e := range validation.Errors {
		coreValidation = append(coreValidation, core.ValidationResult{
			Rule:     e.Code,
			Passed:   false,
			Message:  e.Message,
			Severity: e.Severity,
			Framework: core.FrameworkCBAM,
		})
	}
	
	for _, w := range validation.Warnings {
		coreValidation = append(coreValidation, core.ValidationResult{
			Rule:     w.Code,
			Passed:   false,
			Message:  w.Message,
			Severity: "warning",
			Framework: core.FrameworkCBAM,
		})
	}

	return core.ComplianceReport{
		Standard:          "EU CBAM (Carbon Border Adjustment Mechanism)",
		Framework:         core.FrameworkCBAM,
		Content: map[string]interface{}{
			"declaration": declaration,
			"report":      report,
			"summary":     m.buildSummary(cbamInput, declaration),
		},
		ValidationResults: coreValidation,
		GeneratedAt:       time.Now(),
	}, nil
}

// ValidateInput validates CBAM input data.
// Implements the core.ComplianceMapper interface.
func (m *Mapper) ValidateInput(ctx context.Context, input core.ComplianceInput) ([]core.ValidationResult, error) {
	cbamInput, err := m.coerceInput(input)
	if err != nil {
		return nil, err
	}

	// Generate declaration to trigger validation
	declaration, err := m.generateDeclaration(cbamInput)
	if err != nil {
		return nil, err
	}

	validation := m.validateDeclaration(declaration)
	
	// Convert to core.ValidationResult
	results := make([]core.ValidationResult, 0)
	
	for _, e := range validation.Errors {
		results = append(results, core.ValidationResult{
			Rule:     e.Code,
			Passed:   false,
			Message:  e.Message,
			Severity: e.Severity,
			Framework: core.FrameworkCBAM,
		})
	}
	
	for _, w := range validation.Warnings {
		results = append(results, core.ValidationResult{
			Rule:     w.Code,
			Passed:   false,
			Message:  w.Message,
			Severity: "warning",
			Framework: core.FrameworkCBAM,
		})
	}
	
	return results, nil
}

// GetRequiredFields returns required fields for CBAM reporting.
// Implements the core.ComplianceMapper interface.
func (m *Mapper) GetRequiredFields() []string {
	return []string{
		"declarant_id",
		"declarant_eori",     // EU EORI number
		"declarant_country",
		"reporting_year",
		"reporting_quarter",  // Q1-Q4
		"goods",              // List of imported goods
		"cn_codes",           // Combined Nomenclature codes
		"embedded_emissions", // Emissions per good
	}
}

// =============================================================================
// Input Coercion
// =============================================================================

// coerceInput converts core.ComplianceInput to CBAMInput.
func (m *Mapper) coerceInput(input core.ComplianceInput) (*CBAMInput, error) {
	// Try typed input first
	if cbamInput, ok := input.Data.(*CBAMInput); ok {
		return cbamInput, nil
	}

	// Try map input
	if data, ok := input.Data.(map[string]interface{}); ok {
		return m.mapToCBAMInput(data, input.Year)
	}

	return nil, fmt.Errorf("unsupported input type: %T", input.Data)
}

// mapToCBAMInput converts a map to CBAMInput.
func (m *Mapper) mapToCBAMInput(data map[string]interface{}, year int) (*CBAMInput, error) {
	cbamInput := &CBAMInput{}

	// Extract basic fields
	if v, ok := data["declarant_id"].(string); ok {
		cbamInput.DeclarantID = v
	}
	if v, ok := data["declarant_name"].(string); ok {
		cbamInput.DeclarantName = v
	}
	if v, ok := data["declarant_eori"].(string); ok {
		cbamInput.DeclarantEORI = v
	}
	if v, ok := data["declarant_country"].(string); ok {
		cbamInput.DeclarantCountry = v
	}

	// Quarter and year
	if v, ok := data["quarter"].(int); ok {
		cbamInput.Quarter = v
	} else if v, ok := data["quarter"].(float64); ok {
		cbamInput.Quarter = int(v)
	} else {
		// Determine from current date
		q := GetQuarter(time.Now())
		cbamInput.Quarter = q.Quarter
	}

	if v, ok := data["quarter_year"].(int); ok {
		cbamInput.QuarterYear = v
	} else if v, ok := data["quarter_year"].(float64); ok {
		cbamInput.QuarterYear = int(v)
	} else {
		cbamInput.QuarterYear = year
	}

	// Goods (this would need more complex mapping in practice)
	if goods, ok := data["goods"].([]interface{}); ok {
		cbamInput.Goods = make([]ImportedGood, len(goods))
		for i, g := range goods {
			if goodMap, ok := g.(map[string]interface{}); ok {
				cbamInput.Goods[i] = m.mapToImportedGood(goodMap)
			}
		}
	}

	return cbamInput, nil
}

// mapToImportedGood converts a map to ImportedGood.
func (m *Mapper) mapToImportedGood(data map[string]interface{}) ImportedGood {
	good := ImportedGood{}

	if v, ok := data["good_id"].(string); ok {
		good.GoodID = v
	}
	if v, ok := data["commodity_type"].(string); ok {
		good.CommodityType = CommodityType(v)
	}
	if v, ok := data["cn_code"].(string); ok {
		good.CNCode = v
	}
	if v, ok := data["description"].(string); ok {
		good.Description = v
	}
	if v, ok := data["quantity"].(float64); ok {
		good.Quantity = v
	}
	if v, ok := data["unit"].(string); ok {
		good.Unit = v
	}
	if v, ok := data["country_of_origin"].(string); ok {
		good.CountryOfOrigin = v
	}
	if v, ok := data["customs_value"].(float64); ok {
		good.CustomsValue = v
	}

	return good
}

// =============================================================================
// Emissions Calculation
// =============================================================================

// calculateAllEmissions calculates emissions for all goods in the input.
func (m *Mapper) calculateAllEmissions(ctx context.Context, input *CBAMInput) error {
	for i := range input.Goods {
		emissions, err := m.calculator.CalculateEmbeddedEmissions(ctx, &input.Goods[i])
		if err != nil {
			return fmt.Errorf("failed to calculate emissions for good %s: %w", input.Goods[i].GoodID, err)
		}
		input.Goods[i].EmbeddedEmissions = *emissions
	}
	return nil
}

// =============================================================================
// Declaration Generation
// =============================================================================

// generateDeclaration creates a CBAM declaration from input data.
func (m *Mapper) generateDeclaration(input *CBAMInput) (*CBAMDeclaration, error) {
	declaration := &CBAMDeclaration{
		DeclarationID:      fmt.Sprintf("CBAM-%s-Q%d-%d", input.DeclarantID, input.Quarter, input.QuarterYear),
		DeclarantID:        input.DeclarantID,
		DeclarantName:      input.DeclarantName,
		DeclarantEORI:      input.DeclarantEORI,
		Quarter:            input.Quarter,
		Year:               input.QuarterYear,
		SubmissionDate:     time.Now().UTC(),
		DeclarationStatus:  "draft",
		GoodsEntries:       make([]GoodEntry, 0, len(input.Goods)),
	}

	// Process each good
	for _, good := range input.Goods {
		entry := m.createGoodEntry(good)
		declaration.GoodsEntries = append(declaration.GoodsEntries, entry)

		// Update totals
		declaration.TotalQuantity += good.Quantity
		declaration.TotalEmbeddedEmissions += entry.TotalEmissions
		declaration.TotalCustomsValue += good.CustomsValue

		if entry.UsedDefaultValues {
			declaration.DefaultValuesUsed++
		}
		if entry.Verified {
			declaration.VerifiedEmissions += entry.TotalEmissions
		} else {
			declaration.UnverifiedEmissions += entry.TotalEmissions
		}
	}

	declaration.TotalGoods = len(declaration.GoodsEntries)

	// Calculate financial obligations (using simplified carbon price)
	declaration.EstimatedCBAMPrice = 80.0 // EUR per tCO2e (example price)
	declaration.EstimatedCBAMCost = declaration.TotalEmbeddedEmissions * declaration.EstimatedCBAMPrice
	declaration.NetCBAMObligation = declaration.EstimatedCBAMCost // No adjustment in this simplified version

	return declaration, nil
}

// createGoodEntry creates a declaration entry from an imported good.
func (m *Mapper) createGoodEntry(good ImportedGood) GoodEntry {
	totalEmissions := good.EmbeddedEmissions.TotalSpecificEmissions * good.Quantity
	directEmissions := good.EmbeddedEmissions.DirectEmissions * good.Quantity
	indirectEmissions := good.EmbeddedEmissions.IndirectEmissions * good.Quantity

	return GoodEntry{
		EntryID:            good.GoodID,
		CommodityType:      good.CommodityType,
		CNCode:             good.CNCode,
		Description:        good.Description,
		Quantity:           good.Quantity,
		Unit:               good.Unit,
		CountryOfOrigin:    good.CountryOfOrigin,
		DirectEmissions:    directEmissions,
		IndirectEmissions:  indirectEmissions,
		TotalEmissions:     totalEmissions,
		SpecificEmissions:  good.EmbeddedEmissions.TotalSpecificEmissions,
		CalculationMethod:  good.EmbeddedEmissions.CalculationMethod,
		UsedDefaultValues:  good.EmbeddedEmissions.CalculationMethod == "default_values",
		Verified:           good.EmbeddedEmissions.Verified,
		CustomsValue:       good.CustomsValue,
		EstimatedCBAMCost:  totalEmissions * 80.0, // Example carbon price
		InstallationID:     good.InstallationID,
		ImportDate:         good.ImportDate,
	}
}

// =============================================================================
// Validation
// =============================================================================

// validateDeclaration validates a CBAM declaration.
func (m *Mapper) validateDeclaration(declaration *CBAMDeclaration) *ValidationResults {
	results := &ValidationResults{
		Valid:        true,
		Errors:       make([]ValidationError, 0),
		Warnings:     make([]ValidationWarning, 0),
		InfoMessages: make([]string, 0),
	}

	// Validate declarant information
	if declaration.DeclarantID == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:    "declarant_id",
			Code:     "REQUIRED_FIELD",
			Message:  "Declarant ID is required",
			Severity: "error",
		})
		results.Valid = false
	}

	if declaration.DeclarantEORI == "" {
		results.Errors = append(results.Errors, ValidationError{
			Field:    "declarant_eori",
			Code:     "REQUIRED_FIELD",
			Message:  "EORI number is required for EU CBAM declarations",
			Severity: "error",
		})
		results.Valid = false
	}

	// Validate quarter
	if declaration.Quarter < 1 || declaration.Quarter > 4 {
		results.Errors = append(results.Errors, ValidationError{
			Field:    "quarter",
			Code:     "INVALID_VALUE",
			Message:  "Quarter must be between 1 and 4",
			Severity: "error",
		})
		results.Valid = false
	}

	// Validate goods
	if len(declaration.GoodsEntries) == 0 {
		results.Errors = append(results.Errors, ValidationError{
			Field:    "goods_entries",
			Code:     "REQUIRED_DATA",
			Message:  "At least one good must be declared",
			Severity: "error",
		})
		results.Valid = false
	}

	// Validate each good entry
	for i, entry := range declaration.GoodsEntries {
		m.validateGoodEntry(&entry, i, results)
	}

	// Warnings for data quality
	if declaration.DefaultValuesUsed > 0 {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "default_values",
			Code:    "DATA_QUALITY",
			Message: fmt.Sprintf("%d goods using default emission values - consider obtaining installation-specific data", declaration.DefaultValuesUsed),
		})
	}

	if declaration.UnverifiedEmissions > declaration.TotalEmbeddedEmissions*0.5 {
		results.Warnings = append(results.Warnings, ValidationWarning{
			Field:   "verification",
			Code:    "DATA_QUALITY",
			Message: "More than 50% of emissions are unverified - third-party verification recommended",
		})
	}

	// Info messages
	results.InfoMessages = append(results.InfoMessages,
		fmt.Sprintf("Total embedded emissions: %.2f tCO2e", declaration.TotalEmbeddedEmissions))
	results.InfoMessages = append(results.InfoMessages,
		fmt.Sprintf("Estimated CBAM cost: €%.2f", declaration.EstimatedCBAMCost))

	return results
}

// validateGoodEntry validates a single good entry.
func (m *Mapper) validateGoodEntry(entry *GoodEntry, index int, results *ValidationResults) {
	if !entry.CommodityType.IsValid() {
		results.Errors = append(results.Errors, ValidationError{
			Field:    fmt.Sprintf("goods_entries[%d].commodity_type", index),
			Code:     "INVALID_COMMODITY",
			Message:  fmt.Sprintf("Invalid commodity type: %s", entry.CommodityType),
			Severity: "error",
			EntryID:  entry.EntryID,
		})
		results.Valid = false
	}

	if entry.Quantity <= 0 {
		results.Errors = append(results.Errors, ValidationError{
			Field:    fmt.Sprintf("goods_entries[%d].quantity", index),
			Code:     "INVALID_QUANTITY",
			Message:  "Quantity must be greater than zero",
			Severity: "error",
			EntryID:  entry.EntryID,
		})
		results.Valid = false
	}

	if entry.TotalEmissions < 0 {
		results.Errors = append(results.Errors, ValidationError{
			Field:    fmt.Sprintf("goods_entries[%d].total_emissions", index),
			Code:     "INVALID_EMISSIONS",
			Message:  "Emissions cannot be negative",
			Severity: "error",
			EntryID:  entry.EntryID,
		})
		results.Valid = false
	}
}

// =============================================================================
// Report Building
// =============================================================================

// buildCBAMReport creates a complete CBAM report.
func (m *Mapper) buildCBAMReport(input *CBAMInput, declaration *CBAMDeclaration) *CBAMReport {
	quarterInfo := fmt.Sprintf("Q%d %d", input.Quarter, input.QuarterYear)

	report := &CBAMReport{
		ReportID:        fmt.Sprintf("CBAM-REPORT-%s-%s", input.DeclarantID, quarterInfo),
		GeneratedAt:     time.Now().UTC(),
		ReportingPeriod: quarterInfo,
		Declarant: DeclarantInfo{
			ID:      input.DeclarantID,
			Name:    input.DeclarantName,
			EORI:    input.DeclarantEORI,
			Country: input.DeclarantCountry,
		},
		Declaration: *declaration,
	}

	// Determine compliance status
	if declaration.ValidationResults != nil && declaration.ValidationResults.Valid {
		report.ComplianceStatus = "compliant"
		report.ComplianceScore = m.calculateComplianceScore(declaration)
	} else {
		report.ComplianceStatus = "non_compliant"
		report.ComplianceScore = 50.0
	}

	// Add recommendations
	report.Recommendations = m.generateRecommendations(declaration)

	// Add next steps
	report.NextSteps = m.generateNextSteps(declaration)

	return report
}

// calculateComplianceScore calculates a compliance score (0-100).
func (m *Mapper) calculateComplianceScore(declaration *CBAMDeclaration) float64 {
	score := 100.0

	// Deduct points for using default values
	if declaration.TotalGoods > 0 {
		defaultPct := float64(declaration.DefaultValuesUsed) / float64(declaration.TotalGoods)
		score -= defaultPct * 20 // Up to 20 points deduction
	}

	// Deduct points for unverified emissions
	if declaration.TotalEmbeddedEmissions > 0 {
		unverifiedPct := declaration.UnverifiedEmissions / declaration.TotalEmbeddedEmissions
		score -= unverifiedPct * 30 // Up to 30 points deduction
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}

	return score
}

// generateRecommendations creates recommendations based on the declaration.
func (m *Mapper) generateRecommendations(declaration *CBAMDeclaration) []string {
	recommendations := make([]string, 0)

	if declaration.DefaultValuesUsed > 0 {
		recommendations = append(recommendations,
			"Obtain installation-specific emission data to replace default values and reduce CBAM costs")
	}

	if declaration.UnverifiedEmissions > declaration.TotalEmbeddedEmissions*0.3 {
		recommendations = append(recommendations,
			"Seek third-party verification for emission calculations to improve data quality")
	}

	if declaration.TotalEmbeddedEmissions > 1000 {
		recommendations = append(recommendations,
			"Consider supplier engagement to reduce embedded emissions in imported goods")
	}

	return recommendations
}

// generateNextSteps creates next steps based on declaration status.
func (m *Mapper) generateNextSteps(declaration *CBAMDeclaration) []string {
	steps := make([]string, 0)

	if declaration.ValidationResults != nil && !declaration.ValidationResults.Valid {
		steps = append(steps, "Correct validation errors before submission")
	}

	steps = append(steps, "Review and approve declaration")
	steps = append(steps, fmt.Sprintf("Submit declaration by end of month following Q%d %d", declaration.Quarter, declaration.Year))

	if declaration.EstimatedCBAMCost > 0 {
		steps = append(steps, fmt.Sprintf("Prepare payment of €%.2f for CBAM certificates", declaration.EstimatedCBAMCost))
	}

	return steps
}

// buildSummary creates a summary map for the report.
func (m *Mapper) buildSummary(input *CBAMInput, declaration *CBAMDeclaration) map[string]interface{} {
	summary := m.calculator.GetSummary(input.Goods)

	return map[string]interface{}{
		"declarant":              declaration.DeclarantName,
		"period":                 FormatQuarter(declaration.Year, declaration.Quarter),
		"total_goods":            declaration.TotalGoods,
		"total_emissions":        declaration.TotalEmbeddedEmissions,
		"estimated_cbam_cost":    declaration.EstimatedCBAMCost,
		"using_default_values":   summary.UsingDefaultValues,
		"verified_goods":         summary.Verified,
		"compliance_status":      declaration.DeclarationStatus,
		"validation_errors":      len(declaration.ValidationResults.Errors),
		"validation_warnings":    len(declaration.ValidationResults.Warnings),
	}
}
