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
	"strings"
	"time"

	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/ingestion"
)

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
	PeriodStart time.Time
	PeriodEnd   time.Time
	GeneratedAt time.Time

	Scope1Tonnes float64
	Scope2Tonnes float64
	Scope3Tonnes float64
	TotalTonnes  float64

	LineItems  []InventoryLineItem
	ByScope    map[int]float64
	ByCategory []CategoryTotal
	ByLocation []CategoryTotal
	Factors    []FactorRef

	TotalActivities    int
	CompleteActivities int
	CompletenessPct    float64
	Warnings           []string
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
		report.LineItems = append(report.LineItems, li)

		report.ByScope[scope] += rec.EmissionsTonnesCO2e
		switch scope {
		case 1:
			report.Scope1Tonnes += rec.EmissionsTonnesCO2e
		case 2:
			report.Scope2Tonnes += rec.EmissionsTonnesCO2e
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
