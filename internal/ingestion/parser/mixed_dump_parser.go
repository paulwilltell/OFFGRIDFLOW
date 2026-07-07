package parser

// mixed_dump_parser.go adds support for a single "everything in one file" dump:
// a CSV that carries rows of different scopes, distinguished by an explicit
// activity-type column plus generic quantity + unit columns, e.g.
//
//	type,quantity,unit,location,date,fuel_type,mode,treatment,vendor
//	electricity,120000,kwh,US-CA,2025-01-31,,,,
//	diesel,500,gallons,US-CA,2025-01-15,diesel,,,
//	flight,2586,miles,,2025-03-01,,air,,
//	waste,1200,kg,US-CA,2025-01-15,,,landfill,
//	cloud spend,40000,usd,,2025-01-15,,,,AWS
//
// Each row is classified and routed individually, reusing the same normalize
// helpers and source/category conventions as the per-file parsers. Files that
// do not carry this discriminator fall through to the per-file classifier, so
// both upload styles are supported.

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/ingestion"
)

var (
	mixedTypeCols = []string{"type", "activity_type", "activity", "emission_type", "data_type"}
	mixedQtyCols  = []string{"quantity", "amount", "value", "qty"}
)

// detectMixedSchema reports whether the header describes a mixed/generic dump:
// an explicit activity-type column together with generic quantity and unit
// columns. It deliberately excludes "category"/"scope" as discriminators so it
// never hijacks the standard electricity extended format (which carries
// quantity, unit and category columns of its own).
func detectMixedSchema(colIndex map[string]int) (discCol, qtyCol string, ok bool) {
	discCol = firstCol(colIndex, mixedTypeCols)
	qtyCol = firstCol(colIndex, mixedQtyCols)
	_, hasUnit := colIndex["unit"]
	if discCol == "" || qtyCol == "" || !hasUnit {
		return "", "", false
	}
	return discCol, qtyCol, true
}

// anySubstr reports whether s contains any of the given substrings.
func anySubstr(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// firstNonBlank returns the first non-blank value.
func firstNonBlank(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// mixedDistanceUnit resolves a distance unit value to "km" or "mile".
func mixedDistanceUnit(unitVal string) string {
	switch strings.ToLower(strings.TrimSpace(unitVal)) {
	case "km", "kilometer", "kilometers", "kilometre", "kilometres":
		return "km"
	default:
		return "mile"
	}
}

// classifyMixedRow maps a row's declared type (with unit as a fallback signal)
// to a scope kind. The explicit type value is the primary signal. Order matters:
// freight is checked before travel (both may say "air"/"rail"), and commuting
// before travel so an explicit "commute" is not read as business travel.
func classifyMixedRow(typeVal, unitVal string) scopeKind {
	t := strings.ToLower(strings.TrimSpace(typeVal))
	u := strings.ToLower(strings.TrimSpace(unitVal))
	switch {
	case anySubstr(t, "electric", "grid", "power"), u == "kwh", u == "mwh":
		return kindElectricity
	case anySubstr(t, "commut"):
		return kindCommuting
	case anySubstr(t, "freight", "shipping", "cargo", "haul", "distribution"),
		u == "tonne_km", u == "tonnekm", u == "ton_mile", u == "ton_miles":
		return kindFreight
	case anySubstr(t, "waste", "landfill", "recycl", "compost", "incinerat", "disposal", "trash", "garbage"):
		return kindWaste
	case anySubstr(t, "flight", "air", "plane", "travel", "train", "rail", "taxi", "uber", "lyft", "bus", "car", "drive", "ferry"):
		return kindTravel
	case anySubstr(t, "diesel", "gasoline", "petrol", "natural_gas", "propane", "lpg", "fuel_oil", "jet", "fuel", "combustion", "heating"),
		anySubstr(u, "gallon", "liter", "litre", "therm", "ccf", "mcf"):
		return kindFuel
	case anySubstr(t, "spend", "purchase", "goods", "service", "cloud", "software", "procure", "supplier"),
		u == "usd", u == "eur", u == "gbp":
		return kindSpend
	default:
		return kindUnknown
	}
}

// parseMixedRows converts a mixed/generic dump into activities, classifying and
// routing each row individually.
func (p *UtilityBillParser) parseMixedRows(reader *csv.Reader, colIndex map[string]int, discCol, qtyCol string, startLine int) (*ParseResult, error) {
	var (
		activities []ingestion.Activity
		errors     []ingestion.ImportError
		lineNum    = startLine
	)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		lineNum++
		if err != nil {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Message: fmt.Sprintf("CSV parse error: %v", err)})
			continue
		}
		if isEmptyRecord(record) {
			continue
		}
		get := rowGetter(colIndex, record)

		typeVal := strings.TrimSpace(get(discCol))
		if typeVal == "" {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Field: discCol, Message: "activity type is required for a mixed dump row"})
			continue
		}
		unitVal := strings.ToLower(strings.TrimSpace(get("unit")))
		rawQty, err := parseFlexibleNumber(get(qtyCol))
		if err != nil {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Field: qtyCol, Message: fmt.Sprintf("invalid quantity %q: %v", get(qtyCol), err)})
			continue
		}

		activity, err := p.buildMixedActivity(classifyMixedRow(typeVal, unitVal), typeVal, unitVal, rawQty, get, lineNum)
		if err != nil {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Field: discCol, Message: err.Error()})
			continue
		}
		if err := activity.Validate(); err != nil {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Message: fmt.Sprintf("validation failed: %v", err)})
			continue
		}
		activities = append(activities, activity)
	}
	return p.finishResult(activities, errors, lineNum, "mixed"), nil
}

// buildMixedActivity constructs one activity from a classified mixed-dump row,
// reusing the same conversions and source/category conventions as the per-file
// parsers so a row means the same thing whichever upload style produced it.
func (p *UtilityBillParser) buildMixedActivity(kind scopeKind, typeVal, unitVal string, rawQty float64, get func(string) string, lineNum int) (ingestion.Activity, error) {
	start, end := parsePeriod(get)
	location := firstNonBlank(get("location"), get("site"), p.DefaultLocation)

	base := ingestion.Activity{
		Location:    location,
		PeriodStart: start,
		PeriodEnd:   end,
		OrgID:       p.DefaultOrgID,
		DataQuality: "measured",
		CreatedAt:   time.Now().UTC(),
		Metadata:    map[string]string{"schema_type": "mixed", "declared_type": typeVal},
	}

	switch kind {
	case kindElectricity:
		qty := rawQty
		if unitVal == "mwh" {
			qty *= 1000 // MWh -> kWh
		}
		base.ID = fmt.Sprintf("mixed-elec-%d", lineNum)
		base.Source = "utility_bill"
		base.Category = "electricity"
		base.Quantity = qty
		base.Unit = "kWh"
		if m := get("meter_id"); m != "" {
			base.MeterID = m
		}
		return base, nil

	case kindFuel:
		fuelType := normalizeFuelType(get("fuel_type"))
		if fuelType == "" || fuelType == "fuel" {
			fuelType = normalizeFuelType(typeVal)
		}
		if fuelType == "" || fuelType == "fuel" {
			return base, fmt.Errorf("fuel row needs a specific fuel type (e.g. diesel, gasoline, natural_gas)")
		}
		qty, unit := normalizeFuelQuantity(unitVal, rawQty)
		source := "stationary_combustion"
		if firstNonBlank(get("vehicle"), get("fleet")) != "" || strings.Contains(strings.ToLower(typeVal), "fleet") {
			source = "fleet"
		}
		base.ID = fmt.Sprintf("mixed-fuel-%s-%d", fuelType, lineNum)
		base.Source = source
		base.Category = fuelType
		base.Quantity = qty
		base.Unit = unit
		return base, nil

	case kindTravel:
		mode := normalizeTravelMode(firstNonBlank(get("mode"), typeVal))
		base.ID = fmt.Sprintf("mixed-travel-%s-%d", mode, lineNum)
		base.Source = "travel"
		base.Category = mode
		base.Quantity = rawQty
		base.Unit = mixedDistanceUnit(unitVal)
		return base, nil

	case kindCommuting:
		mode := normalizeCommuteMode(firstNonBlank(get("mode"), get("commute_mode"), typeVal))
		base.ID = fmt.Sprintf("mixed-commute-%s-%d", mode, lineNum)
		base.Source = "commuting"
		base.Category = mode
		base.Quantity = rawQty
		base.Unit = mixedDistanceUnit(unitVal)
		base.DataQuality = "estimated"
		return base, nil

	case kindWaste:
		cat := strings.ToLower(strings.TrimSpace(firstNonBlank(get("treatment"), get("disposal"), get("waste_type"), typeVal)))
		if cat == "" || cat == "waste" {
			cat = "landfill"
		}
		base.ID = fmt.Sprintf("mixed-waste-%d", lineNum)
		base.Source = "waste"
		base.Category = cat
		base.Quantity = wasteWeightToKg(unitVal, rawQty)
		base.Unit = "kg"
		return base, nil

	case kindFreight:
		if !(unitVal == "tonne_km" || unitVal == "tonnekm" || unitVal == "ton_mile" || unitVal == "ton_miles") {
			return base, fmt.Errorf("freight row needs unit tonne_km in a mixed dump")
		}
		tkm := rawQty
		if unitVal == "ton_mile" || unitVal == "ton_miles" {
			tkm *= 1.60934 // ton-mile -> tonne-km (approx)
		}
		mode := normalizeFreightMode(firstNonBlank(get("mode"), typeVal))
		base.ID = fmt.Sprintf("mixed-freight-%s-%d", mode, lineNum)
		base.Source = "freight"
		base.Category = mode
		base.Quantity = tkm
		base.Unit = "tonne_km"
		return base, nil

	case kindSpend:
		cat := strings.ToLower(strings.TrimSpace(firstNonBlank(get("category"), get("spend_category"), typeVal)))
		if cat == "" || cat == "spend" || cat == "purchase" {
			cat = "general-goods"
		}
		cur := "USD"
		if c := strings.ToUpper(unitVal); c == "USD" || c == "EUR" || c == "GBP" {
			cur = c
		} else if c := strings.ToUpper(get("currency")); c == "USD" || c == "EUR" || c == "GBP" {
			cur = c
		}
		if v := firstNonBlank(get("vendor"), get("supplier")); v != "" {
			base.Metadata["vendor"] = v
		}
		base.ID = fmt.Sprintf("mixed-spend-%d", lineNum)
		base.Source = "purchases"
		base.Category = cat
		base.Quantity = rawQty
		base.Unit = cur
		base.DataQuality = "estimated"
		return base, nil

	default:
		return base, fmt.Errorf("unrecognized activity type %q (expected e.g. electricity, diesel, flight, commute, waste, freight, spend)", typeVal)
	}
}
