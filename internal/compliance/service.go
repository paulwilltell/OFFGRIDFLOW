// Package compliance provides unified compliance service for all frameworks
package compliance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/compliance/california"
	"github.com/example/offgridflow/internal/compliance/cbam"
	"github.com/example/offgridflow/internal/compliance/csrd"
	"github.com/example/offgridflow/internal/compliance/ifrs"
	"github.com/example/offgridflow/internal/compliance/sec"
	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/ingestion"
)

// Service provides methods to generate compliance reports across all frameworks
type Service struct {
	activityStore    ingestion.ActivityStore
	scope1Calculator *emissions.Scope1Calculator
	scope2Calculator *emissions.Scope2Calculator
	scope3Calculator *emissions.Scope3Calculator
	csrdMapper       *csrd.DefaultCSRDMapper
	secMapper        *sec.DefaultSECMapper
	californiaMapper *california.Mapper
	cbamMapper       *cbam.Mapper
	ifrsMapper       *ifrs.Mapper
}

// NewService creates a new compliance service
func NewService(
	activityStore ingestion.ActivityStore,
	scope1Calc *emissions.Scope1Calculator,
	scope2Calc *emissions.Scope2Calculator,
	scope3Calc *emissions.Scope3Calculator,
) *Service {
	return &Service{
		activityStore:    activityStore,
		scope1Calculator: scope1Calc,
		scope2Calculator: scope2Calc,
		scope3Calculator: scope3Calc,
		csrdMapper:       csrd.NewDefaultCSRDMapper(),
		secMapper:        sec.NewDefaultSECMapper(),
		californiaMapper: &california.Mapper{},
		cbamMapper:       &cbam.Mapper{},
		ifrsMapper:       &ifrs.Mapper{},
	}
}

// EmissionsTotals holds aggregated emissions
type EmissionsTotals struct {
	Scope1Tons float64
	Scope2Tons float64
	Scope3Tons float64
	TotalTons  float64
}

// calculateEmissions fetches activities and calculates all scope emissions
func (s *Service) calculateEmissions(ctx context.Context, orgID string, year int) (*EmissionsTotals, []ingestion.Activity, error) {
	// Fetch activities for the organization and year
	allActivities, err := s.activityStore.List(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load activities: %w", err)
	}

	// Filter activities
	var activities []ingestion.Activity
	for _, act := range allActivities {
		// Filter by org_id if provided
		if orgID != "" && orgID != "org-demo" && act.OrgID != orgID {
			continue
		}

		// Filter by year
		if year > 0 {
			startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
			endDate := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
			if act.PeriodStart.Before(startDate) || !act.PeriodStart.Before(endDate) {
				continue
			}
		}

		activities = append(activities, act)
	}

	// Convert to emissions.Activity interface
	emissionsActivities := make([]emissions.Activity, 0, len(activities))
	for i := range activities {
		emissionsActivities = append(emissionsActivities, &activities[i])
	}

	// Calculate all scopes
	scope1Records, err := s.scope1Calculator.CalculateBatch(ctx, emissionsActivities)
		if err != nil {
			return nil, nil, fmt.Errorf("scope1 calculation failed: %w", err)
		}

	scope2Records, err := s.scope2Calculator.CalculateBatch(ctx, emissionsActivities)
		if err != nil {
			return nil, nil, fmt.Errorf("scope2 calculation failed: %w", err)
		}

	scope3Records, err := s.scope3Calculator.CalculateBatch(ctx, emissionsActivities)
		if err != nil {
			return nil, nil, fmt.Errorf("scope3 calculation failed: %w", err)
		}

	// Aggregate totals
	totals := &EmissionsTotals{}
	for _, rec := range scope1Records {
		totals.Scope1Tons += rec.EmissionsTonnesCO2e
	}
	for _, rec := range scope2Records {
		totals.Scope2Tons += rec.EmissionsTonnesCO2e
	}
	for _, rec := range scope3Records {
		totals.Scope3Tons += rec.EmissionsTonnesCO2e
	}
	totals.TotalTons = totals.Scope1Tons + totals.Scope2Tons + totals.Scope3Tons

	return totals, activities, nil
}

// ComplianceStatus represents the overall compliance status
type ComplianceStatus string

const (
	StatusNotStarted  ComplianceStatus = "not_started"
	StatusPartial     ComplianceStatus = "partial"
	StatusCompliant   ComplianceStatus = "compliant"
	StatusUnknown     ComplianceStatus = "unknown"
	StatusNotRequired ComplianceStatus = "not_required"
)

// determineStatus calculates compliance status based on available data
func determineStatus(totals *EmissionsTotals, hasGovernance, hasAttestation bool) ComplianceStatus {
	if totals == nil {
		return StatusUnknown
	}

	hasScope1 := totals.Scope1Tons > 0
	hasScope2 := totals.Scope2Tons > 0
	hasScope3 := totals.Scope3Tons > 0

	// No data at all
	if !hasScope1 && !hasScope2 && !hasScope3 {
		return StatusNotStarted
	}

	// All scopes present with governance/attestation
	if hasScope1 && hasScope2 && hasScope3 && hasGovernance && hasAttestation {
		return StatusCompliant
	}

	// Some data present
	return StatusPartial
}

// GenerateCSRDReport generates a CSRD/ESRS E1 compliance report
func (s *Service) GenerateCSRDReport(ctx context.Context, orgID string, year int) (*csrd.CSRDReport, error) {
	totals, activities, err := s.calculateEmissions(ctx, orgID, year)
	if err != nil {
		return nil, err
	}

	input := s.buildCSRDInput(orgID, year, totals, activities)

	report, err := s.csrdMapper.BuildReport(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to build CSRD report: %w", err)
	}

	return &report, nil
}

// GenerateSECReport generates an SEC Climate Disclosure report
func (s *Service) GenerateSECReport(ctx context.Context, orgID, orgName, cik string, fiscalYear int) (*sec.SECReport, error) {
	totals, _, err := s.calculateEmissions(ctx, orgID, fiscalYear)
	if err != nil {
		return nil, err
	}

	input := sec.SECInput{
		OrgID:      orgID,
		OrgName:    orgName,
		CIK:        cik,
		FiscalYear: fiscalYear,
		FilerType:  "large-accelerated", // Default, should be configurable
		IsEGC:      false,
		GHGMetrics: &sec.GHGMetricsDisclosure{
			IsRequired: true,
			Scope1Emissions: &sec.ScopeEmissions{
				TotalEmissions: totals.Scope1Tons,
			},
			Scope2Emissions: &sec.ScopeEmissions{
				TotalEmissions: totals.Scope2Tons,
			},
		},
	}

	report, err := s.secMapper.BuildReport(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to build SEC report: %w", err)
	}

	return &report, nil
}

// GenerateCaliforniaReport generates a California Climate Disclosure report
func (s *Service) GenerateCaliforniaReport(ctx context.Context, orgID, orgName string, year int) (interface{}, error) {
	totals, _, err := s.calculateEmissions(ctx, orgID, year)
	if err != nil {
		return nil, err
	}

	input := california.Input{
		OrgID:       orgID,
		OrgName:     orgName,
		Year:        year,
		Scope1Tons:  totals.Scope1Tons,
		Scope2Tons:  totals.Scope2Tons,
		Scope3Tons:  totals.Scope3Tons,
		Methodology: "GHG Protocol",
	}

	// Use core.ComplianceInput wrapper
	coreInput := struct {
		OrgID string
		Year  int
		Data  interface{}
	}{
		OrgID: orgID,
		Year:  year,
		Data:  input,
	}

	// We need a core.ComplianceInput type
	// For now, return a simplified report directly from mapper
	report := california.Report{
		OrgID:       input.OrgID,
		OrgName:     input.OrgName,
		Year:        input.Year,
		Scope1Tons:  input.Scope1Tons,
		Scope2Tons:  input.Scope2Tons,
		Scope3Tons:  input.Scope3Tons,
		TotalTons:   input.Scope1Tons + input.Scope2Tons + input.Scope3Tons,
		Methodology: input.Methodology,
	}

	_ = coreInput

	return &report, nil
}

// GenerateCBAMReport generates an EU CBAM report
func (s *Service) GenerateCBAMReport(ctx context.Context, orgID string, quarter int, year int) (interface{}, error) {
	totals, _, err := s.calculateEmissions(ctx, orgID, year)
	if err != nil {
		return nil, err
	}

	// CBAM report structure - simplified
	report := map[string]interface{}{
		"org_id":          orgID,
		"quarter":         quarter,
		"year":            year,
		"total_emissions": totals.TotalTons,
		"status":          "not_applicable", // Most orgs won't need CBAM unless importing to EU
		"message":         "CBAM applies to specific goods imported into the EU",
	}

	_ = totals

	return report, nil
}

// GenerateIFRSReport generates an IFRS S2 Climate-related Disclosures report
func (s *Service) GenerateIFRSReport(ctx context.Context, orgID, orgName string, year int) (interface{}, error) {
	totals, _, err := s.calculateEmissions(ctx, orgID, year)
	if err != nil {
		return nil, err
	}

	input := ifrs.Input{
		OrgID:       orgID,
		OrgName:     orgName,
		Year:        year,
		Scope1:      totals.Scope1Tons,
		Scope2LB:    totals.Scope2Tons,
		Scope2MB:    totals.Scope2Tons,
		Scope3:      totals.Scope3Tons,
		Methodology: "GHG Protocol",
	}

	// Simplified IFRS report
	report := ifrs.Report{
		OrgID:          input.OrgID,
		OrgName:        input.OrgName,
		Year:           input.Year,
		Scope1:         input.Scope1,
		Scope2Location: input.Scope2LB,
		Scope2Market:   input.Scope2MB,
		Scope3:         input.Scope3,
		Total:          input.Scope1 + input.Scope2LB + input.Scope3,
		Methodology:    input.Methodology,
	}

	return &report, nil
}

// FrameworkSummary represents the status of a single framework
type FrameworkSummary struct {
	Name     string           `json:"name"`
	Status   ComplianceStatus `json:"status"`
	Scope1   bool             `json:"scope1_ready"`
	Scope2   bool             `json:"scope2_ready"`
	Scope3   bool             `json:"scope3_ready"`
	HasData  bool             `json:"has_data"`
	DataGaps []string         `json:"data_gaps,omitempty"`
}

// ComplianceSummary represents the overall compliance status across all frameworks
type ComplianceSummary struct {
	Frameworks map[string]FrameworkSummary `json:"frameworks"`
	Totals     EmissionsTotals             `json:"totals"`
	Timestamp  string                      `json:"timestamp"`
}

// GenerateSummary generates a compliance summary across all frameworks
func (s *Service) GenerateSummary(ctx context.Context, orgID string, year int) (*ComplianceSummary, error) {
	totals, _, err := s.calculateEmissions(ctx, orgID, year)
	if err != nil {
		return nil, err
	}

	hasScope1 := totals.Scope1Tons > 0
	hasScope2 := totals.Scope2Tons > 0
	hasScope3 := totals.Scope3Tons > 0

	// Determine status for each framework
	summary := &ComplianceSummary{
		Frameworks: make(map[string]FrameworkSummary),
		Totals:     *totals,
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	// CSRD - requires all scopes
	csrdGaps := []string{}
	if !hasScope1 {
		csrdGaps = append(csrdGaps, "Scope 1 emissions")
	}
	if !hasScope2 {
		csrdGaps = append(csrdGaps, "Scope 2 emissions")
	}
	if !hasScope3 {
		csrdGaps = append(csrdGaps, "Scope 3 emissions")
	}
	summary.Frameworks["csrd"] = FrameworkSummary{
		Name:     "CSRD/ESRS E1",
		Status:   determineStatus(totals, false, false),
		Scope1:   hasScope1,
		Scope2:   hasScope2,
		Scope3:   hasScope3,
		HasData:  totals.TotalTons > 0,
		DataGaps: csrdGaps,
	}

	// SEC - requires Scope 1 & 2, Scope 3 if material
	secGaps := []string{}
	if !hasScope1 {
		secGaps = append(secGaps, "Scope 1 emissions")
	}
	if !hasScope2 {
		secGaps = append(secGaps, "Scope 2 emissions")
	}
	secStatus := StatusNotStarted
	if hasScope1 && hasScope2 {
		if hasScope3 {
			secStatus = StatusCompliant
		} else {
			secStatus = StatusPartial
		}
	} else if hasScope1 || hasScope2 {
		secStatus = StatusPartial
	}
	summary.Frameworks["sec"] = FrameworkSummary{
		Name:     "SEC Climate Disclosure",
		Status:   secStatus,
		Scope1:   hasScope1,
		Scope2:   hasScope2,
		Scope3:   hasScope3,
		HasData:  totals.TotalTons > 0,
		DataGaps: secGaps,
	}

	// California - requires all scopes
	californiaGaps := []string{}
	if !hasScope1 {
		californiaGaps = append(californiaGaps, "Scope 1 emissions")
	}
	if !hasScope2 {
		californiaGaps = append(californiaGaps, "Scope 2 emissions")
	}
	if !hasScope3 {
		californiaGaps = append(californiaGaps, "Scope 3 emissions")
	}
	summary.Frameworks["california"] = FrameworkSummary{
		Name:     "California Climate Disclosure",
		Status:   determineStatus(totals, false, false),
		Scope1:   hasScope1,
		Scope2:   hasScope2,
		Scope3:   hasScope3,
		HasData:  totals.TotalTons > 0,
		DataGaps: californiaGaps,
	}

	// CBAM - product-specific, may not be applicable
	cbamStatus := StatusNotRequired
	if totals.TotalTons > 0 {
		cbamStatus = StatusPartial // Need product-level data
	}
	summary.Frameworks["cbam"] = FrameworkSummary{
		Name:     "EU CBAM",
		Status:   cbamStatus,
		Scope1:   hasScope1,
		Scope2:   hasScope2,
		Scope3:   hasScope3,
		HasData:  totals.TotalTons > 0,
		DataGaps: []string{"Product-level emissions data required"},
	}

	// IFRS S2 - similar to CSRD
	ifrsGaps := []string{}
	if !hasScope1 {
		ifrsGaps = append(ifrsGaps, "Scope 1 emissions")
	}
	if !hasScope2 {
		ifrsGaps = append(ifrsGaps, "Scope 2 emissions")
	}
	if !hasScope3 {
		ifrsGaps = append(ifrsGaps, "Scope 3 emissions")
	}
	summary.Frameworks["ifrs_s2"] = FrameworkSummary{
		Name:     "IFRS S2 Climate-related Disclosures",
		Status:   determineStatus(totals, false, false),
		Scope1:   hasScope1,
		Scope2:   hasScope2,
		Scope3:   hasScope3,
		HasData:  totals.TotalTons > 0,
		DataGaps: ifrsGaps,
	}

	return summary, nil
}

// buildCSRDInput assembles a CSRD input payload based on calculated totals and raw activities.
func (s *Service) buildCSRDInput(orgID string, year int, totals *EmissionsTotals, activities []ingestion.Activity) csrd.CSRDInput {
	energy := s.aggregateEnergyBreakdown(activities)
	scope3 := s.aggregateScope3ByCategory(activities, totals.Scope3Tons)

	return csrd.CSRDInput{
		OrgID:               orgID,
		OrgName:             orgID,
		Year:                year,
		TotalScope1Tons:     totals.Scope1Tons,
		TotalScope2Tons:     totals.Scope2Tons,
		TotalScope3Tons:     totals.Scope3Tons,
		TotalEnergyMWh:      energy.total,
		RenewableEnergyMWh:  energy.renewable,
		FossilFuelEnergyMWh: energy.fossil,
		NuclearEnergyMWh:    energy.nuclear,
		Scope3Categories:    scope3,
	}
}

type energyBreakdown struct {
	total     float64
	renewable float64
	fossil    float64
	nuclear   float64
}

func (s *Service) aggregateEnergyBreakdown(activities []ingestion.Activity) energyBreakdown {
	result := energyBreakdown{}
	for _, act := range activities {
		if !ingestion.Unit(act.Unit).IsEnergyUnit() {
			continue
		}
		mwh := convertToMWh(act.Quantity, ingestion.Unit(act.Unit))
		if mwh == 0 {
			continue
		}
		result.total += mwh
		switch classifyEnergySource(act) {
		case "renewable":
			result.renewable += mwh
		case "fossil":
			result.fossil += mwh
		case "nuclear":
			result.nuclear += mwh
		}
	}
	return result
}

func convertToMWh(quantity float64, unit ingestion.Unit) float64 {
	switch unit {
	case ingestion.UnitMWh:
		return quantity
	case ingestion.UnitKWh:
		return quantity / 1000
	case ingestion.UnitGJ:
		// 1 GJ ≈ 0.277778 MWh
		return quantity * 0.277778
	case ingestion.UnitTherm:
		// 1 therm ≈ 0.0293071 MWh
		return quantity * 0.0293071
	default:
		return 0
	}
}

func classifyEnergySource(act ingestion.Activity) string {
	if source, ok := act.Metadata["energy_source"]; ok && source != "" {
		return strings.ToLower(strings.TrimSpace(source))
	}
	lower := strings.ToLower(act.Category)
	switch {
	case strings.Contains(lower, "solar"), strings.Contains(lower, "wind"), strings.Contains(lower, "renewable"):
		return "renewable"
	case strings.Contains(lower, "gas"), strings.Contains(lower, "diesel"), strings.Contains(lower, "natural"):
		return "fossil"
	case strings.Contains(lower, "nuclear"):
		return "nuclear"
	default:
		return "other"
	}
}

func (s *Service) aggregateScope3ByCategory(activities []ingestion.Activity, scope3Total float64) map[string]float64 {
	counts := make(map[string]int)
	for _, act := range activities {
		if !isScope3Source(act.Source) {
			continue
		}
		key := scope3CategoryKey(act)
		if key == "" {
			key = "scope3_other"
		}
		counts[key]++
	}
	if len(counts) == 0 || scope3Total == 0 {
		return nil
	}
	totalCount := 0
	for _, v := range counts {
		totalCount += v
	}
	if totalCount == 0 {
		return nil
	}
	result := make(map[string]float64, len(counts))
	perUnit := scope3Total / float64(totalCount)
	for key, cnt := range counts {
		result[key] = perUnit * float64(cnt)
	}
	return result
}

func isScope3Source(source string) bool {
	switch ingestion.Source(source) {
	case ingestion.SourceTravel, ingestion.SourcePurchases, ingestion.SourceWaste, ingestion.SourceManual, ingestion.SourceAPI:
		return true
	default:
		return false
	}
}

func scope3CategoryKey(act ingestion.Activity) string {
	if cat, ok := act.Metadata["scope3_category"]; ok && cat != "" {
		return strings.ToLower(strings.TrimSpace(cat))
	}
	if act.Category != "" {
		return strings.ToLower(strings.TrimSpace(act.Category))
	}
	if act.Source != "" {
		return strings.ToLower(strings.TrimSpace(act.Source))
	}
	return ""
}
