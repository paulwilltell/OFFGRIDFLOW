package cbam

import (
	"fmt"
	"time"
)

// =============================================================================
// Validator
// =============================================================================

// Validator checks CBAM declaration and report consistency.
type Validator struct{}

// NewValidator creates a new CBAM validator.
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateDeclaration performs comprehensive validation of a CBAM declaration.
func (v *Validator) ValidateDeclaration(declaration *CBAMDeclaration) *ValidationResults {
	results := &ValidationResults{
		Valid:        true,
		Errors:       make([]ValidationError, 0),
		Warnings:     make([]ValidationWarning, 0),
		InfoMessages: make([]string, 0),
	}

	// Validate header information
	v.validateHeader(declaration, results)

	// Validate temporal data
	v.validateTemporal(declaration, results)

	// Validate goods entries
	v.validateGoodsEntries(declaration, results)

	// Validate financial calculations
	v.validateFinancials(declaration, results)

	// Cross-validation checks
	v.crossValidate(declaration, results)

	return results
}

// validateHeader validates declaration header information.
func (v *Validator) validateHeader(declaration *CBAMDeclaration, results *ValidationResults) {
	if declaration.DeclarationID == "" {
		v.addError(results, "declaration_id", "REQUIRED", "Declaration ID is required", "critical")
	}

	if declaration.DeclarantID == "" {
		v.addError(results, "declarant_id", "REQUIRED", "Declarant ID is required", "error")
	}

	if declaration.DeclarantEORI == "" {
		v.addError(results, "declarant_eori", "REQUIRED", "EORI number is required for EU CBAM", "error")
	} else if !v.isValidEORI(declaration.DeclarantEORI) {
		v.addWarning(results, "declarant_eori", "FORMAT", "EORI number format appears invalid")
	}

	if declaration.DeclarantName == "" {
		v.addWarning(results, "declarant_name", "MISSING", "Declarant name is recommended")
	}
}

// validateTemporal validates time-related fields.
func (v *Validator) validateTemporal(declaration *CBAMDeclaration, results *ValidationResults) {
	if declaration.Quarter < 1 || declaration.Quarter > 4 {
		v.addError(results, "quarter", "INVALID_RANGE", "Quarter must be between 1 and 4", "error")
	}

	currentYear := time.Now().Year()
	if declaration.Year < 2023 || declaration.Year > currentYear+1 {
		v.addWarning(results, "year", "UNUSUAL_VALUE",
			fmt.Sprintf("Year %d is outside expected range (2023-%d)", declaration.Year, currentYear+1))
	}

	// Check if submission is within deadline
	if !declaration.SubmissionDate.IsZero() {
		deadline := v.getSubmissionDeadline(declaration.Year, declaration.Quarter)
		if declaration.SubmissionDate.After(deadline) {
			v.addWarning(results, "submission_date", "LATE_SUBMISSION",
				fmt.Sprintf("Submission after deadline (%s)", deadline.Format("2006-01-02")))
		}
	}
}

// validateGoodsEntries validates all goods entries.
func (v *Validator) validateGoodsEntries(declaration *CBAMDeclaration, results *ValidationResults) {
	if len(declaration.GoodsEntries) == 0 {
		v.addError(results, "goods_entries", "REQUIRED", "At least one good must be declared", "error")
		return
	}

	for i, entry := range declaration.GoodsEntries {
		v.validateGoodEntry(&entry, i, results)
	}
}

// validateGoodEntry validates a single good entry.
func (v *Validator) validateGoodEntry(entry *GoodEntry, index int, results *ValidationResults) {
	prefix := fmt.Sprintf("goods_entries[%d]", index)

	// Commodity type
	if !entry.CommodityType.IsValid() {
		v.addErrorWithEntry(results, prefix+".commodity_type", "INVALID",
			fmt.Sprintf("Invalid commodity type: %s", entry.CommodityType), "error", entry.EntryID)
	}

	// Quantity
	if entry.Quantity <= 0 {
		v.addErrorWithEntry(results, prefix+".quantity", "INVALID",
			"Quantity must be greater than zero", "error", entry.EntryID)
	}

	// Emissions
	if entry.TotalEmissions < 0 {
		v.addErrorWithEntry(results, prefix+".total_emissions", "NEGATIVE",
			"Emissions cannot be negative", "error", entry.EntryID)
	}

	// Check emissions calculation
	calculated := entry.DirectEmissions + entry.IndirectEmissions
	tolerance := 0.01
	if entry.TotalEmissions > 0 && (calculated < entry.TotalEmissions*(1-tolerance) || calculated > entry.TotalEmissions*(1+tolerance)) {
		v.addWarningWithEntry(results, prefix+".total_emissions", "CALCULATION_MISMATCH",
			fmt.Sprintf("Total emissions (%.4f) doesn't match direct + indirect (%.4f)", entry.TotalEmissions, calculated),
			entry.EntryID)
	}

	// Country of origin
	if entry.CountryOfOrigin == "" {
		v.addWarningWithEntry(results, prefix+".country_of_origin", "MISSING",
			"Country of origin is recommended", entry.EntryID)
	}

	// CN code
	if entry.CNCode == "" {
		v.addWarningWithEntry(results, prefix+".cn_code", "MISSING",
			"CN code (Combined Nomenclature) is recommended", entry.EntryID)
	}
}

// validateFinancials validates financial calculations.
func (v *Validator) validateFinancials(declaration *CBAMDeclaration, results *ValidationResults) {
	// Validate CBAM price
	if declaration.EstimatedCBAMPrice <= 0 {
		v.addWarning(results, "estimated_cbam_price", "INVALID",
			"CBAM price should be greater than zero")
	}

	// Validate cost calculation
	expectedCost := declaration.TotalEmbeddedEmissions * declaration.EstimatedCBAMPrice
	tolerance := 0.01
	if expectedCost > 0 && (declaration.EstimatedCBAMCost < expectedCost*(1-tolerance) ||
		declaration.EstimatedCBAMCost > expectedCost*(1+tolerance)) {
		v.addWarning(results, "estimated_cbam_cost", "CALCULATION_MISMATCH",
			fmt.Sprintf("CBAM cost (%.2f) doesn't match emissions × price (%.2f)",
				declaration.EstimatedCBAMCost, expectedCost))
	}
}

// crossValidate performs cross-field validation.
func (v *Validator) crossValidate(declaration *CBAMDeclaration, results *ValidationResults) {
	// Verify total goods count
	if declaration.TotalGoods != len(declaration.GoodsEntries) {
		v.addWarning(results, "total_goods", "MISMATCH",
			fmt.Sprintf("Total goods count (%d) doesn't match entries (%d)",
				declaration.TotalGoods, len(declaration.GoodsEntries)))
	}

	// Verify emissions totals
	var calculatedTotal float64
	for _, entry := range declaration.GoodsEntries {
		calculatedTotal += entry.TotalEmissions
	}

	tolerance := 0.01
	if calculatedTotal > 0 && (declaration.TotalEmbeddedEmissions < calculatedTotal*(1-tolerance) ||
		declaration.TotalEmbeddedEmissions > calculatedTotal*(1+tolerance)) {
		v.addWarning(results, "total_embedded_emissions", "MISMATCH",
			fmt.Sprintf("Declared total (%.4f) doesn't match sum of entries (%.4f)",
				declaration.TotalEmbeddedEmissions, calculatedTotal))
	}

	// Data quality checks
	if declaration.DefaultValuesUsed > declaration.TotalGoods {
		v.addWarning(results, "default_values_used", "INCONSISTENT",
			"Default values used exceeds total goods count")
	}

	if declaration.UnverifiedEmissions > declaration.TotalEmbeddedEmissions {
		v.addWarning(results, "unverified_emissions", "INCONSISTENT",
			"Unverified emissions exceeds total emissions")
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// addError adds a validation error.
func (v *Validator) addError(results *ValidationResults, field, code, message, severity string) {
	results.Errors = append(results.Errors, ValidationError{
		Field:    field,
		Code:     code,
		Message:  message,
		Severity: severity,
	})
	results.Valid = false
}

// addErrorWithEntry adds a validation error with entry ID.
func (v *Validator) addErrorWithEntry(results *ValidationResults, field, code, message, severity, entryID string) {
	results.Errors = append(results.Errors, ValidationError{
		Field:    field,
		Code:     code,
		Message:  message,
		Severity: severity,
		EntryID:  entryID,
	})
	results.Valid = false
}

// addWarning adds a validation warning.
func (v *Validator) addWarning(results *ValidationResults, field, code, message string) {
	results.Warnings = append(results.Warnings, ValidationWarning{
		Field:   field,
		Code:    code,
		Message: message,
	})
}

// addWarningWithEntry adds a validation warning with entry ID.
func (v *Validator) addWarningWithEntry(results *ValidationResults, field, code, message, entryID string) {
	results.Warnings = append(results.Warnings, ValidationWarning{
		Field:   field,
		Code:    code,
		Message: message,
		EntryID: entryID,
	})
}

// isValidEORI performs basic EORI format validation.
func (v *Validator) isValidEORI(eori string) bool {
	// EORI format: 2-letter country code + up to 15 alphanumeric characters
	// This is a simplified check
	if len(eori) < 3 || len(eori) > 17 {
		return false
	}
	return true
}

// getSubmissionDeadline returns the submission deadline for a quarter.
func (v *Validator) getSubmissionDeadline(year, quarter int) time.Time {
	// Deadline is end of month following the quarter
	var month time.Month
	switch quarter {
	case 1:
		month = 4 // End of April
	case 2:
		month = 7 // End of July
	case 3:
		month = 10 // End of October
	case 4:
		month = 1 // End of January (next year)
		year++
	default:
		return time.Time{}
	}

	// Last day of the month
	lastDay := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
	return lastDay
}

// ValidateInput validates CBAM input data before processing.
func (v *Validator) ValidateInput(input *CBAMInput) *ValidationResults {
	results := &ValidationResults{
		Valid:        true,
		Errors:       make([]ValidationError, 0),
		Warnings:     make([]ValidationWarning, 0),
		InfoMessages: make([]string, 0),
	}

	// Validate declarant information
	if input.DeclarantID == "" {
		v.addError(results, "declarant_id", "REQUIRED", "Declarant ID is required", "error")
	}

	if input.DeclarantEORI == "" {
		v.addError(results, "declarant_eori", "REQUIRED", "EORI number is required", "error")
	}

	// Validate quarter
	if input.Quarter < 1 || input.Quarter > 4 {
		v.addError(results, "quarter", "INVALID_RANGE", "Quarter must be between 1 and 4", "error")
	}

	// Validate goods
	if len(input.Goods) == 0 {
		v.addError(results, "goods", "REQUIRED", "At least one good must be provided", "error")
	}

	for i, good := range input.Goods {
		if good.GoodID == "" {
			v.addErrorWithEntry(results, fmt.Sprintf("goods[%d].good_id", i), "REQUIRED",
				"Good ID is required", "error", "")
		}

		if !good.CommodityType.IsValid() {
			v.addErrorWithEntry(results, fmt.Sprintf("goods[%d].commodity_type", i), "INVALID",
				fmt.Sprintf("Invalid commodity type: %s", good.CommodityType), "error", good.GoodID)
		}

		if good.Quantity <= 0 {
			v.addErrorWithEntry(results, fmt.Sprintf("goods[%d].quantity", i), "INVALID",
				"Quantity must be greater than zero", "error", good.GoodID)
		}
	}

	return results
}
