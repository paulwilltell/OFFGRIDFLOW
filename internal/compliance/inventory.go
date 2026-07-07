package compliance

// inventory.go assembles a full, traceable greenhouse-gas inventory from the
// per-activity emission records — not just scope totals. This is the data
// backbone for an audit-ready, assurance-grade report (California SB 253 /
// GHG Protocol Corporate Standard): every reported number can be traced back
// to an activity, an emission factor, and its published source.

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/ingestion"
)

// metaFloat parses a numeric value from activity metadata, returning -1 when the
// key is absent or unparseable.
func metaFloat(meta map[string]string, key string) float64 {
	if meta == nil {
		return -1
	}
	v, ok := meta[key]
	if !ok {
		return -1
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil {
		return -1
	}
	return f
}

// InventoryLineItem is one activity's contribution to the inventory, carrying
// the full calculation trail (activity data -> factor -> emissions).
type InventoryLineItem struct {
	ActivityID      string
	Source          string
	Category        string
	Scope           int
	ScopeLabel      string
	Location        string
	PeriodStart     time.Time
	PeriodEnd       time.Time
	Quantity        float64
	Unit            string
	EmissionFactor  float64
	FactorID        string
	FactorSource    string
	Method          string
	DataQuality     string
	EmissionsTonnes float64
	// Biogenic marks combustion of biofuels (biodiesel, ethanol, wood, etc.).
	// Its CO2 is a separate memo item, excluded from the fossil scope totals.
	Biogenic bool

	// Scope 2 market-based (dual reporting). MarketEmissionsTonnes equals the
	// location-based figure unless a supplier/market factor or renewable share
	// was supplied for this account.
	MarketFactor          float64
	MarketEmissionsTonnes float64
}

// FactorRef is a distinct emission factor used in the inventory, for the
// factor-source audit trail section.
type FactorRef struct {
	FactorID string
	Category string
	Scope    int
	Value    float64
	Unit     string
	Source   string
	Uses     int
}

// InventoryReport is the complete, structured GHG inventory for one org-year.
type InventoryReport struct {
	OrgID       string
	OrgName     string
	Year        int
	BaseYear    int
	PeriodStart time.Time
	PeriodEnd   time.Time
	GeneratedAt time.Time

	Scope1Tonnes float64
	Scope2Tonnes float64 // Scope 2 location-based (the primary reported figure)
	Scope3Tonnes float64
	TotalTonnes  float64

	// Scope2MarketTonnes is the Scope 2 market-based total (dual reporting).
	// HasMarketData is true when at least one account supplied a market factor
	// or renewable share.
	Scope2MarketTonnes float64
	HasMarketData      bool

	// Scope1Gases disaggregates Scope 1 into the seven Kyoto gases.
	Scope1Gases GasBreakdown

	LineItems     []InventoryLineItem
	ByScope       map[int]float64
	ByCategory    []CategoryTotal
	ByLocation    []CategoryTotal
	Factors       []FactorRef
	Scope3Coverage []Scope3CategoryStatus

	TotalActivities    int
	CompleteActivities int
	CompletenessPct    float64
	Warnings           []string

	// Normalization denominators for emissions-intensity metrics. Zero means
	// "not supplied" (intensity is then omitted rather than divided by zero).
	Revenue   float64 // annual revenue for the reporting period, in USD
	Employees int     // average full-time-equivalent headcount
}

// IntensityPerRevenueMM returns emissions intensity in tCO2e per $1M of revenue,
// or 0 when revenue was not supplied. This is the primary comparability metric
// under the GHG Protocol and the basis on which SB 253 applicability is assessed.
func (r *InventoryReport) IntensityPerRevenueMM() float64 {
	if r.Revenue <= 0 {
		return 0
	}
	return r.TotalTonnes / (r.Revenue / 1_000_000)
}

// IntensityPerEmployee returns emissions intensity in tCO2e per full-time
// equivalent employee, or 0 when headcount was not supplied.
func (r *InventoryReport) IntensityPerEmployee() float64 {
	if r.Employees <= 0 {
		return 0
	}
	return r.TotalTonnes / float64(r.Employees)
}

// GasBreakdown disaggregates emissions into the seven Kyoto greenhouse gases
// (all values tCO2e). Biogenic CO2 is reported separately from the totals.
type GasBreakdown struct {
	CO2      float64
	CH4      float64
	N2O      float64
	HFCs     float64
	PFCs     float64
	SF6      float64
	NF3      float64
	Biogenic float64
}

// Total returns the sum of the seven Kyoto gases (excludes biogenic).
func (g GasBreakdown) Total() float64 {
	return g.CO2 + g.CH4 + g.N2O + g.HFCs + g.PFCs + g.SF6 + g.NF3
}

// combustionGasShares returns the representative share of a combustion fuel's
// CO2e attributable to CO2, CH4 and N2O. Combustion is CO2-dominant; CH4/N2O
// are minor but non-zero. Shares sum to 1 so the gas split reconciles exactly
// to the fuel's total. For assurance, fuel/technology-specific gas factors
// (EPA GHG Emission Factors Hub / IPCC) should replace these.
func combustionGasShares(fuelCategory string) (co2, ch4, n2o float64) {
	f := strings.ToLower(fuelCategory)
	switch {
	case strings.Contains(f, "natural_gas"), strings.Contains(f, "propane"), strings.Contains(f, "lpg"), strings.Contains(f, "methane"), strings.Contains(f, "cng"), strings.Contains(f, "lng"):
		return 0.995, 0.003, 0.002 // gaseous fuels
	case strings.Contains(f, "coal"), strings.Contains(f, "wood"), strings.Contains(f, "biomass"):
		return 0.985, 0.006, 0.009 // solid fuels
	default:
		return 0.985, 0.003, 0.012 // liquid fuels (diesel, gasoline, fuel oil, jet)
	}
}

// isBiogenicFuel reports whether a fuel's combustion CO2 is biogenic (reported
// separately from the fossil Scope 1 total per the GHG Protocol).
func isBiogenicFuel(fuelCategory string) bool {
	f := strings.ToLower(fuelCategory)
	return strings.Contains(f, "biodiesel") || strings.Contains(f, "ethanol") ||
		strings.Contains(f, "biomass") || strings.Contains(f, "biogas") ||
		strings.Contains(f, "wood") || strings.Contains(f, "biofuel")
}

// Scope3CategoryStatus reports whether each of the 15 GHG Protocol Scope 3
// categories is included, as required (excluded categories must be justified).
type Scope3CategoryStatus struct {
	Number   int
	Name     string
	Reported bool
	Tonnes   float64
}

// scope3CategoryNames are the 15 GHG Protocol Scope 3 categories in order.
var scope3CategoryNames = [15]string{
	"Purchased goods and services",
	"Capital goods",
	"Fuel- and energy-related activities",
	"Upstream transportation and distribution",
	"Waste generated in operations",
	"Business travel",
	"Employee commuting",
	"Upstream leased assets",
	"Downstream transportation and distribution",
	"Processing of sold products",
	"Use of sold products",
	"End-of-life treatment of sold products",
	"Downstream leased assets",
	"Franchises",
	"Investments",
}

// scope3CategoryOf maps a Scope 3 activity's source/category to its GHG Protocol
// category number (0 if not classifiable).
func scope3CategoryOf(source, category string) int {
	s := strings.ToLower(source)
	c := strings.ToLower(category)
	switch {
	case strings.Contains(s, "travel"):
		return 6
	case strings.Contains(s, "commut"):
		return 7
	case strings.Contains(s, "waste"):
		return 5
	case strings.Contains(s, "freight"), strings.Contains(s, "upstream_transport"), strings.Contains(s, "shipping"):
		return 4
	case strings.Contains(s, "purchas"), strings.Contains(s, "spend"):
		if strings.Contains(c, "capital") || strings.Contains(c, "equipment") || strings.Contains(c, "machinery") {
			return 2 // Capital goods
		}
		return 1 // Purchased goods and services
	default:
		return 1
	}
}

// CategoryTotal is a labeled emissions subtotal (by category or location).
type CategoryTotal struct {
	Label   string
	Scope   int
	Tonnes  float64
	Percent float64
}

func scopeLabel(scope int) string {
	switch scope {
	case 1:
		return "Scope 1 (Direct)"
	case 2:
		return "Scope 2 (Energy indirect)"
	case 3:
		return "Scope 3 (Value chain)"
	default:
		return "Unclassified"
	}
}

// factorSourceLabel derives the published source of an emission factor from its
// ID and scope, so the report can cite where each factor came from. These map
// to the factor tables the calculators use.
func factorSourceLabel(scope int, factorID string) string {
	id := strings.ToLower(factorID)
	switch {
	case strings.Contains(id, "grid"), strings.Contains(id, "egrid"):
		return "US EPA eGRID (location-based)"
	case scope == 1:
		return "US EPA GHG Emission Factors Hub"
	case strings.Contains(id, "travel"), strings.Contains(id, "commuting"):
		return "UK DEFRA / GHG Protocol (passenger transport)"
	case strings.Contains(id, "transport"), strings.Contains(id, "freight"):
		return "UK DEFRA / GHG Protocol (freight)"
	case strings.Contains(id, "spend"), strings.Contains(id, "eeio"):
		return "US EPA EEIO (spend-based)"
	case strings.Contains(id, "waste"):
		return "UK DEFRA (waste treatment)"
	case scope == 3:
		return "GHG Protocol Scope 3 factors"
	default:
		return "OffGridFlow default factor library"
	}
}

// GenerateInventory builds the full traceable inventory for an org and year by
// running each scope calculator and preserving every per-activity record.
func (s *Service) GenerateInventory(ctx context.Context, orgID string, year int) (*InventoryReport, error) {
	_, activities, err := s.calculateEmissions(ctx, orgID, year)
	if err != nil {
		return nil, err
	}

	// Index activities by ID so records can be joined back to their source.
	actByID := make(map[string]*ingestion.Activity, len(activities))
	emActivities := make([]emissions.Activity, 0, len(activities))
	for i := range activities {
		actByID[activities[i].ID] = &activities[i]
		emActivities = append(emActivities, &activities[i])
	}

	// Collect per-activity records from every scope calculator.
	var records []emissions.EmissionRecord
	if s.scope1Calculator != nil {
		if r, err := s.scope1Calculator.CalculateBatch(ctx, emActivities); err == nil {
			records = append(records, r...)
		}
	}
	if s.scope2Calculator != nil {
		if r, err := s.scope2Calculator.CalculateBatch(ctx, emActivities); err == nil {
			records = append(records, r...)
		}
	}
	if s.scope3Calculator != nil {
		if r, err := s.scope3Calculator.CalculateBatch(ctx, emActivities); err == nil {
			records = append(records, r...)
		}
	}

	report := &InventoryReport{
		OrgID:       orgID,
		Year:        year,
		PeriodStart: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC),
		GeneratedAt: time.Now().UTC(),
		ByScope:     map[int]float64{1: 0, 2: 0, 3: 0},
	}

	catTotals := map[string]*CategoryTotal{}
	locTotals := map[string]*CategoryTotal{}
	factorByID := map[string]*FactorRef{}

	for _, rec := range records {
		scope := int(rec.Scope)
		act := actByID[rec.ActivityID]

		li := InventoryLineItem{
			ActivityID:      rec.ActivityID,
			Scope:           scope,
			ScopeLabel:      scopeLabel(scope),
			PeriodStart:     rec.PeriodStart,
			PeriodEnd:       rec.PeriodEnd,
			Quantity:        rec.InputQuantity,
			Unit:            rec.InputUnit,
			EmissionFactor:  rec.EmissionFactor,
			FactorID:        rec.FactorID,
			FactorSource:    factorSourceLabel(scope, rec.FactorID),
			Method:          string(rec.Method),
			DataQuality:     string(rec.DataQuality),
			EmissionsTonnes: rec.EmissionsTonnesCO2e,
			Location:        rec.Region,
		}
		if act != nil {
			li.Source = act.Source
			li.Category = act.Category
			if act.Location != "" {
				li.Location = act.Location
			}
		}
		if li.Category == "" {
			li.Category = "uncategorized"
		}
		if li.Location == "" {
			li.Location = "unspecified"
		}

		// Market-based (dual reporting) figure. Defaults to the location-based
		// value unless the account supplied a supplier/market emission factor or
		// a renewable share (contractual instruments per GHG Protocol Scope 2).
		li.MarketEmissionsTonnes = rec.EmissionsTonnesCO2e
		if scope == 2 && act != nil {
			if mf := metaFloat(act.Metadata, "market_factor"); mf >= 0 {
				li.MarketFactor = mf
				li.MarketEmissionsTonnes = rec.InputQuantity * mf / 1000.0
				report.HasMarketData = true
			} else if rp := metaFloat(act.Metadata, "renewable_pct"); rp >= 0 {
				share := rp
				if share > 1 {
					share /= 100 // accept 0-100 or 0-1
				}
				if share < 0 {
					share = 0
				} else if share > 1 {
					share = 1
				}
				li.MarketEmissionsTonnes = rec.EmissionsTonnesCO2e * (1 - share)
				report.HasMarketData = true
			}
		}

		li.Biogenic = scope == 1 && isBiogenicFuel(li.Category)
		report.LineItems = append(report.LineItems, li)

		// Record the emission factor in the audit trail regardless of origin.
		if factorByID[rec.FactorID] == nil {
			factorByID[rec.FactorID] = &FactorRef{
				FactorID: rec.FactorID,
				Category: li.Category,
				Scope:    scope,
				Value:    rec.EmissionFactor,
				Unit:     rec.InputUnit,
				Source:   li.FactorSource,
			}
		}
		factorByID[rec.FactorID].Uses++

		// Biogenic CO2 (from biofuels) is a separate memo item per the GHG
		// Protocol: reported on its own and excluded from every fossil scope
		// aggregate, so the reported totals reconcile exactly.
		if li.Biogenic {
			report.Scope1Gases.Biogenic += rec.EmissionsTonnesCO2e
			continue
		}

		report.ByScope[scope] += rec.EmissionsTonnesCO2e
		switch scope {
		case 1:
			report.Scope1Tonnes += rec.EmissionsTonnesCO2e
			// Disaggregate this fuel's CO2e into CO2/CH4/N2O.
			co2s, ch4s, n2os := combustionGasShares(li.Category)
			report.Scope1Gases.CO2 += rec.EmissionsTonnesCO2e * co2s
			report.Scope1Gases.CH4 += rec.EmissionsTonnesCO2e * ch4s
			report.Scope1Gases.N2O += rec.EmissionsTonnesCO2e * n2os
		case 2:
			report.Scope2Tonnes += rec.EmissionsTonnesCO2e
			report.Scope2MarketTonnes += li.MarketEmissionsTonnes
		case 3:
			report.Scope3Tonnes += rec.EmissionsTonnesCO2e
		}

		catKey := fmt.Sprintf("%d|%s", scope, li.Category)
		if catTotals[catKey] == nil {
			catTotals[catKey] = &CategoryTotal{Label: li.Category, Scope: scope}
		}
		catTotals[catKey].Tonnes += rec.EmissionsTonnesCO2e

		if locTotals[li.Location] == nil {
			locTotals[li.Location] = &CategoryTotal{Label: li.Location}
		}
		locTotals[li.Location].Tonnes += rec.EmissionsTonnesCO2e
	}

	report.TotalTonnes = report.Scope1Tonnes + report.Scope2Tonnes + report.Scope3Tonnes

	// Sort line items by scope then emissions desc for a readable report.
	sort.SliceStable(report.LineItems, func(i, j int) bool {
		if report.LineItems[i].Scope != report.LineItems[j].Scope {
			return report.LineItems[i].Scope < report.LineItems[j].Scope
		}
		return report.LineItems[i].EmissionsTonnes > report.LineItems[j].EmissionsTonnes
	})

	report.ByCategory = sortedTotals(catTotals, report.TotalTonnes)
	report.ByLocation = sortedTotals(locTotals, report.TotalTonnes)

	for _, f := range factorByID {
		report.Factors = append(report.Factors, *f)
	}
	sort.SliceStable(report.Factors, func(i, j int) bool {
		if report.Factors[i].Scope != report.Factors[j].Scope {
			return report.Factors[i].Scope < report.Factors[j].Scope
		}
		return report.Factors[i].Category < report.Factors[j].Category
	})

	// Data quality: completeness = share of activities with the fields a report
	// needs (quantity, unit, period, location).
	report.TotalActivities = len(activities)
	for i := range activities {
		a := activities[i]
		if a.Quantity > 0 && a.Unit != "" && !a.PeriodStart.IsZero() && a.Location != "" {
			report.CompleteActivities++
		}
	}
	if report.TotalActivities > 0 {
		report.CompletenessPct = float64(report.CompleteActivities) / float64(report.TotalActivities) * 100
	}

	// Base year: absent a prior inventory, the reporting year is established as
	// the base year for future GHG Protocol tracking.
	report.BaseYear = year

	// Scope 3 category coverage — which of the 15 categories are included.
	s3Tonnes := map[int]float64{}
	for _, li := range report.LineItems {
		if li.Scope == 3 {
			s3Tonnes[scope3CategoryOf(li.Source, li.Category)] += li.EmissionsTonnes
		}
	}
	for i, name := range scope3CategoryNames {
		num := i + 1
		report.Scope3Coverage = append(report.Scope3Coverage, Scope3CategoryStatus{
			Number:   num,
			Name:     name,
			Reported: s3Tonnes[num] > 0,
			Tonnes:   s3Tonnes[num],
		})
	}

	if report.TotalActivities == 0 {
		report.Warnings = append(report.Warnings, "No activity data found for the reporting period.")
	}
	if report.Scope3Tonnes == 0 {
		report.Warnings = append(report.Warnings, "No Scope 3 emissions reported. SB 253 requires Scope 3 disclosure (from 2027 reporting).")
	}
	if report.CompletenessPct < 100 && report.TotalActivities > 0 {
		report.Warnings = append(report.Warnings, fmt.Sprintf("%.0f%% of activities are missing one or more recommended fields (location/period).", 100-report.CompletenessPct))
	}

	return report, nil
}

// sortedTotals converts a total map into a percentage-ranked slice.
func sortedTotals(m map[string]*CategoryTotal, grand float64) []CategoryTotal {
	out := make([]CategoryTotal, 0, len(m))
	for _, v := range m {
		if grand > 0 {
			v.Percent = v.Tonnes / grand * 100
		}
		out = append(out, *v)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Tonnes > out[j].Tonnes })
	return out
}
