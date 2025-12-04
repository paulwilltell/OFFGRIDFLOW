package ifrs

// Validator checks IFRS S2 disclosure completeness.
type Validator struct{}

// ValidateReport evaluates readiness and gaps/warnings.
func (v *Validator) ValidateReport(report Report, input Input) (bool, []string, []string) {
	var errs []string
	var warns []string
	valid := true

	if report.OrgID == "" {
		errs = append(errs, "org_id is required")
		valid = false
	}
	if report.Year == 0 {
		errs = append(errs, "year is required")
		valid = false
	}
	if report.Scope1 == 0 {
		errs = append(errs, "Scope 1 emissions required for IFRS S2")
		valid = false
	}
	if report.Scope2Location == 0 {
		errs = append(errs, "Scope 2 (location-based) required for IFRS S2")
		valid = false
	}
	// Scope 2 market-based optional but recommended
	if report.Scope2Market == 0 {
		warns = append(warns, "Scope 2 (market-based) missing")
	}
	if report.Scope3 == 0 {
		warns = append(warns, "Scope 3 emissions missing (disclose or explain)")
	}

	// Governance/strategy/process expectations
	if !input.HasRiskProcess {
		warns = append(warns, "Climate risk identification/assessment process not provided")
	}
	if !input.HasMetricsTargets {
		warns = append(warns, "Metrics and targets disclosures incomplete")
	}
	if !input.HasTransitionPlan {
		warns = append(warns, "Transition plan not provided (state if not applicable)")
	}
	if report.Methodology == "" {
		warns = append(warns, "GHG accounting methodology not specified")
	}

	return valid, errs, warns
}
