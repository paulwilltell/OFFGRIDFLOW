package california

import "fmt"

// Validator checks California climate report completeness.
type Validator struct{}

// ValidateReport verifies that a California climate report contains minimum required data.
func (v *Validator) ValidateReport(report Report) (valid bool, errors []string, warnings []string) {
	valid = true

	if report.OrgID == "" {
		errors = append(errors, "org_id is required")
		valid = false
	}
	if report.Year == 0 {
		errors = append(errors, "year is required")
		valid = false
	}
	if report.Scope1Tons == 0 {
		errors = append(errors, "Scope 1 emissions are required for SB 253 disclosure")
		valid = false
	}
	if report.Scope2Tons == 0 {
		errors = append(errors, "Scope 2 emissions are required for SB 253 disclosure")
		valid = false
	}
	if report.Scope3Tons == 0 {
		errors = append(errors, "Scope 3 emissions are required for SB 253 disclosure")
		valid = false
	}
	if report.AssuranceRequired && !report.AssuranceProvided {
		warnings = append(warnings, "Third-party assurance required but not provided")
	}

	// If no emissions at all, treat as invalid for disclosure purposes.
	if report.Scope1Tons == 0 && report.Scope2Tons == 0 && report.Scope3Tons == 0 {
		errors = append(errors, "no emissions data provided")
		valid = false
	}

	// Ensure disclosure status aligns with validation outcome.
	return valid, errors, warnings
}

// ValidateInput ensures the input has required identifiers.
func (v *Validator) ValidateInput(in Input) error {
	var missing []string
	if in.OrgID == "" {
		missing = append(missing, "org_id")
	}
	if in.Year == 0 {
		missing = append(missing, "year")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %v", missing)
	}
	return nil
}
