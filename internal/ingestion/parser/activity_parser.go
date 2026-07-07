package parser

// activity_parser.go extends the CSV ingestion path beyond electricity so a
// user can "just dump" fuel, travel, or supplier-spend data and have the engine
// structure it into the right GHG Protocol scope automatically.
//
// The utility-bill (Scope 2) path is unchanged. This file adds:
//
//   - Scope 1 (direct combustion): fuel receipts / fleet / stationary burners.
//   - Scope 3 cat 6 (business travel): trips with a distance and mode.
//   - Scope 3 cat 1 (purchased goods & services): supplier spend rows.
//
// Each parser emits ingestion.Activity records tagged with the Source/Category/
// Unit the corresponding calculator (Scope1Calculator / Scope3Calculator)
// already understands, and normalizes quantities into the calculators' native
// units so no downstream conversion is required:
//
//   - liquid fuels  -> liters (US gallons x 3.78541)
//   - natural gas   -> cubic meters (therms x 2.731, ccf x 2.83168, mcf x 28.3168)
//   - travel        -> miles or km (the Scope 3 calculator converts miles->km)
//   - spend         -> USD / EUR / GBP (spend-based EEIO factors)

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/ingestion"
)

// scopeKind identifies which GHG scope a dumped CSV represents, inferred from
// its column headers.
type scopeKind string

const (
	kindElectricity scopeKind = "electricity" // Scope 2 (existing utility-bill path)
	kindFuel        scopeKind = "fuel"         // Scope 1 direct combustion
	kindTravel      scopeKind = "travel"       // Scope 3 cat 6 business travel
	kindSpend       scopeKind = "spend"        // Scope 3 cat 1 purchased goods/services
	kindWaste       scopeKind = "waste"        // Scope 3 cat 5 waste generated in operations
	kindCommuting   scopeKind = "commuting"    // Scope 3 cat 7 employee commuting
	kindFreight     scopeKind = "freight"      // Scope 3 cat 4/9 transportation & distribution
	kindUnknown     scopeKind = "unknown"      // a mixed-dump row whose type could not be classified
)

// Column signatures used to classify a dump. Names are already normalized by
// normalizeColumnName (lowercase, underscores).
var (
	fuelTypeCols  = []string{"fuel_type", "fuel", "fuel_category"}
	fuelQtyCols   = []string{"gallons", "gallon", "liters", "litres", "liter", "litre", "therms", "therm", "ccf", "mcf", "cubic_meters", "m3"}
	distanceCols  = []string{"miles", "mile", "distance", "distance_miles", "distance_km", "km", "kilometers", "kilometres"}
	travelHintCol = []string{"mode", "travel_mode", "transport_mode", "origin", "destination"}
	spendAmtCols  = []string{"amount_usd", "amount", "spend", "spend_usd", "cost", "total", "value"}
	spendHintCols = []string{"vendor", "supplier", "merchant", "category", "spend_category"}
	dateCols      = []string{"date", "period_start", "transaction_date", "invoice_date", "trip_date", "service_start"}

	// Scope 3 category signatures.
	wasteHintCols   = []string{"waste_type", "waste_stream", "treatment", "disposal", "disposal_method"}
	wasteQtyCols    = []string{"weight_kg", "kg", "kilograms", "tonnes", "tonne", "tons", "ton", "weight", "mass"}
	commuteHintCols = []string{"commute", "commuting", "commute_distance", "commute_mode"}
	freightTonneKm  = []string{"tonne_km", "tonnekm", "tonne_kilometers", "tonne_kilometres", "ton_miles", "ton_mile"}
	freightMassCols = []string{"cargo_weight", "freight_weight", "weight_tonnes", "shipment_weight", "cargo", "tonnes", "tonne", "tons", "ton"}
	freightHintCols = []string{"freight", "shipment", "shipping", "carrier"}
)

func hasAny(colIndex map[string]int, names []string) bool {
	for _, n := range names {
		if _, ok := colIndex[n]; ok {
			return true
		}
	}
	return false
}

func firstCol(colIndex map[string]int, names []string) string {
	for _, n := range names {
		if _, ok := colIndex[n]; ok {
			return n
		}
	}
	return ""
}

// classifyScope inspects the header columns and decides which scope this dump
// represents. Order matters: fuel and travel have the most specific signatures,
// spend is next, and electricity is the fallback so existing utility-bill
// behavior is preserved for anything that doesn't look like the new kinds.
func classifyScope(colIndex map[string]int) scopeKind {
	// Scope 1: an explicit fuel type, or a combustion-fuel quantity unit.
	if hasAny(colIndex, fuelTypeCols) || hasAny(colIndex, fuelQtyCols) {
		return kindFuel
	}

	// Scope 3 cat 5 — waste: a disposal/treatment signal.
	if hasAny(colIndex, wasteHintCols) {
		return kindWaste
	}

	// Scope 3 cat 4/9 — freight: cargo mass moved over distance (or tonne-km).
	// Checked before travel because freight also has distance + mode.
	if hasAny(colIndex, freightTonneKm) ||
		(hasAny(colIndex, freightMassCols) && hasAny(colIndex, distanceCols)) ||
		(hasAny(colIndex, freightHintCols) && hasAny(colIndex, distanceCols)) {
		return kindFreight
	}

	// Scope 3 cat 7 — employee commuting: an explicit commute signal.
	if hasAny(colIndex, commuteHintCols) {
		return kindCommuting
	}

	// Scope 3 cat 6 — business travel: a distance column together with a travel
	// hint (mode or an origin/destination pair).
	if hasAny(colIndex, distanceCols) && hasAny(colIndex, travelHintCol) {
		return kindTravel
	}

	// Scope 3 cat 1 — spend: a money amount together with a vendor/category hint.
	if hasAny(colIndex, spendAmtCols) && hasAny(colIndex, spendHintCols) {
		return kindSpend
	}

	return kindElectricity
}

// rowGetter builds a column accessor for a decoded CSV record.
func rowGetter(colIndex map[string]int, record []string) func(string) string {
	return func(col string) string {
		if col == "" {
			return ""
		}
		if idx, ok := colIndex[col]; ok && idx < len(record) {
			return strings.TrimSpace(record[idx])
		}
		return ""
	}
}

// parsePeriod resolves a period from a date column, or a quarter+year pair, and
// returns zero times when no temporal information is present (valid per
// Activity.Validate, which only orders non-zero periods).
func parsePeriod(get func(string) string) (time.Time, time.Time) {
	for _, c := range dateCols {
		if v := get(c); v != "" {
			if t, err := parseFlexibleDate(v); err == nil {
				// Prefer an explicit end column when the start came from a
				// service_start/period_start pair.
				end := t
				if ev := get("period_end"); ev != "" {
					if et, err := parseFlexibleDate(ev); err == nil {
						end = et
					}
				} else if ev := get("service_end"); ev != "" {
					if et, err := parseFlexibleDate(ev); err == nil {
						end = et
					}
				}
				return t, end
			}
		}
	}

	// Quarter + year (common in spend exports), e.g. "Q1" / "2025".
	if q := strings.ToUpper(get("quarter")); strings.HasPrefix(q, "Q") {
		if year := get("year"); year != "" {
			if y, err := strconv.Atoi(strings.TrimSpace(year)); err == nil {
				switch q {
				case "Q1":
					return time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(y, 3, 31, 0, 0, 0, 0, time.UTC)
				case "Q2":
					return time.Date(y, 4, 1, 0, 0, 0, 0, time.UTC), time.Date(y, 6, 30, 0, 0, 0, 0, time.UTC)
				case "Q3":
					return time.Date(y, 7, 1, 0, 0, 0, 0, time.UTC), time.Date(y, 9, 30, 0, 0, 0, 0, time.UTC)
				case "Q4":
					return time.Date(y, 10, 1, 0, 0, 0, 0, time.UTC), time.Date(y, 12, 31, 0, 0, 0, 0, time.UTC)
				}
			}
		}
	}

	return time.Time{}, time.Time{}
}

// =============================================================================
// Scope 1 — Fuel / combustion
// =============================================================================

// normalizeFuelType maps common fuel-name synonyms to the categories the
// Scope 1 calculator's default factor table recognizes.
func normalizeFuelType(raw string) string {
	f := strings.ToLower(strings.TrimSpace(raw))
	f = strings.ReplaceAll(f, " ", "_")
	switch f {
	case "petrol", "unleaded", "gas_petrol", "regular", "premium":
		return "gasoline"
	case "cng", "lng", "ng":
		return "natural_gas"
	case "heating_oil", "oil":
		return "fuel_oil"
	case "lpg":
		return "propane"
	case "jet", "aviation", "jet_a", "jet_a1":
		return "jet_fuel"
	default:
		return f
	}
}

// normalizeFuelQuantity converts a raw fuel quantity into the calculator's
// native unit (liters for liquids, cubic meters for gas) and returns the value,
// the normalized unit, and whether the fuel is gaseous.
func normalizeFuelQuantity(qtyCol string, value float64) (float64, string) {
	switch qtyCol {
	case "gallons", "gallon":
		return value * 3.78541, "L" // US gallon -> liter
	case "liters", "litres", "liter", "litre":
		return value, "L"
	case "therms", "therm":
		return value * 2.731, "m3" // therm -> m3 (HHV ~1037 BTU/ft3)
	case "ccf":
		return value * 2.83168, "m3" // 100 ft3 -> m3
	case "mcf":
		return value * 28.3168, "m3" // 1000 ft3 -> m3
	case "cubic_meters", "m3":
		return value, "m3"
	default:
		return value, "L"
	}
}

// parseFuelRows converts fuel/combustion rows into Scope 1 activities.
func (p *UtilityBillParser) parseFuelRows(reader *csv.Reader, colIndex map[string]int, startLine int) (*ParseResult, error) {
	typeCol := firstCol(colIndex, fuelTypeCols)
	qtyCol := firstCol(colIndex, fuelQtyCols)
	if qtyCol == "" {
		return nil, fmt.Errorf("fuel data detected but no quantity column (gallons/liters/therms/ccf/mcf) found")
	}
	hasVehicle := hasAny(colIndex, []string{"vehicle", "vehicle_id", "fleet", "asset"})

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

		// Resolve fuel type: explicit column, else infer natural gas from a
		// gaseous quantity unit.
		fuelType := normalizeFuelType(get(typeCol))
		if fuelType == "" {
			switch qtyCol {
			case "therms", "therm", "ccf", "mcf", "cubic_meters", "m3":
				fuelType = "natural_gas"
			default:
				errors = append(errors, ingestion.ImportError{Row: lineNum, Field: "fuel_type", Message: "fuel type is required (e.g. diesel, gasoline, natural_gas)"})
				continue
			}
		}

		rawQty, err := parseFlexibleNumber(get(qtyCol))
		if err != nil {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Field: qtyCol, Message: fmt.Sprintf("invalid quantity %q: %v", get(qtyCol), err)})
			continue
		}
		qty, unit := normalizeFuelQuantity(qtyCol, rawQty)
		start, end := parsePeriod(get)

		source := "stationary_combustion"
		if hasVehicle {
			source = "fleet"
		}

		metadata := map[string]string{"ingested_unit": qtyCol}
		if v := get("vehicle"); v != "" {
			metadata["vehicle"] = v
		}
		if v := get("location"); v != "" {
			metadata["site"] = v
		}
		if v := get("notes"); v != "" {
			metadata["notes"] = v
		}

		location := get("location")
		if location == "" {
			location = get("site")
		}
		if location == "" {
			location = p.DefaultLocation
		}

		activity := ingestion.Activity{
			ID:          fmt.Sprintf("fuel-%s-%d", fuelType, lineNum),
			Source:      source,
			Category:    fuelType,
			Location:    location,
			PeriodStart: start,
			PeriodEnd:   end,
			Quantity:    qty,
			Unit:        unit,
			OrgID:       p.DefaultOrgID,
			Metadata:    metadata,
			DataQuality: "measured",
			CreatedAt:   time.Now().UTC(),
		}
		if err := activity.Validate(); err != nil {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Message: fmt.Sprintf("validation failed: %v", err)})
			continue
		}
		activities = append(activities, activity)
	}

	return p.finishResult(activities, errors, lineNum, "fuel"), nil
}

// =============================================================================
// Scope 3 cat 6 — Business travel
// =============================================================================

// normalizeTravelMode maps a raw travel mode to a category the Scope 3
// business-travel factor table recognizes.
func normalizeTravelMode(raw string) string {
	m := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case strings.Contains(m, "air"), strings.Contains(m, "plane"), strings.Contains(m, "flight"), strings.Contains(m, "fly"):
		return "flight"
	case strings.Contains(m, "rail"), strings.Contains(m, "train"):
		return "train"
	case strings.Contains(m, "taxi"), strings.Contains(m, "uber"), strings.Contains(m, "lyft"), strings.Contains(m, "rideshare"):
		return "taxi"
	case strings.Contains(m, "bus"), strings.Contains(m, "coach"):
		return "bus"
	case strings.Contains(m, "car"), strings.Contains(m, "drive"), strings.Contains(m, "auto"), strings.Contains(m, "rental"):
		return "car"
	case m == "":
		return "car" // conservative default when mode is unstated
	default:
		return m
	}
}

// travelUnit returns "mile" or "km" for a distance column name.
func travelUnit(distCol string) string {
	switch distCol {
	case "km", "kilometers", "kilometres", "distance_km":
		return "km"
	default:
		return "mile"
	}
}

// parseTravelRows converts trip rows into Scope 3 business-travel activities.
func (p *UtilityBillParser) parseTravelRows(reader *csv.Reader, colIndex map[string]int, startLine int) (*ParseResult, error) {
	distCol := firstCol(colIndex, distanceCols)
	if distCol == "" {
		return nil, fmt.Errorf("travel data detected but no distance column (miles/km) found")
	}
	modeCol := firstCol(colIndex, []string{"mode", "travel_mode", "transport_mode", "type"})
	unit := travelUnit(distCol)

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

		distance, err := parseFlexibleNumber(get(distCol))
		if err != nil {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Field: distCol, Message: fmt.Sprintf("invalid distance %q: %v", get(distCol), err)})
			continue
		}

		// Multiply by passenger count when present (factors are per passenger-km).
		if pc := get("passengers"); pc != "" {
			if n, err := parseFlexibleNumber(pc); err == nil && n > 0 {
				distance *= n
			}
		}

		mode := normalizeTravelMode(get(modeCol))
		start, end := parsePeriod(get)

		metadata := map[string]string{}
		if v := get("employee"); v != "" {
			metadata["employee"] = v
		}
		if o, d := get("origin"), get("destination"); o != "" || d != "" {
			metadata["route"] = strings.TrimSpace(o + "->" + d)
		}

		activity := ingestion.Activity{
			ID:          fmt.Sprintf("travel-%s-%d", mode, lineNum),
			Source:      "travel",
			Category:    mode,
			Location:    p.DefaultLocation,
			PeriodStart: start,
			PeriodEnd:   end,
			Quantity:    distance,
			Unit:        unit,
			OrgID:       p.DefaultOrgID,
			Metadata:    metadata,
			DataQuality: "measured",
			CreatedAt:   time.Now().UTC(),
		}
		if err := activity.Validate(); err != nil {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Message: fmt.Sprintf("validation failed: %v", err)})
			continue
		}
		activities = append(activities, activity)
	}

	return p.finishResult(activities, errors, lineNum, "travel"), nil
}

// =============================================================================
// Scope 3 cat 1 — Purchased goods & services (spend-based)
// =============================================================================

// spendCurrency resolves the currency for a spend row, defaulting to USD.
func spendCurrency(get func(string) string, amtCol string) string {
	if c := strings.ToUpper(get("currency")); c == "USD" || c == "EUR" || c == "GBP" {
		return c
	}
	if amtCol == "amount_usd" || amtCol == "spend_usd" {
		return "USD"
	}
	return "USD"
}

// parseSpendRows converts supplier-spend rows into Scope 3 purchased-goods
// activities calculated with spend-based (EEIO) factors.
func (p *UtilityBillParser) parseSpendRows(reader *csv.Reader, colIndex map[string]int, startLine int) (*ParseResult, error) {
	amtCol := firstCol(colIndex, spendAmtCols)
	if amtCol == "" {
		return nil, fmt.Errorf("spend data detected but no amount column (amount/spend/cost) found")
	}
	catCol := firstCol(colIndex, []string{"category", "spend_category", "type", "description"})

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

		amount, err := parseFlexibleNumber(get(amtCol))
		if err != nil {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Field: amtCol, Message: fmt.Sprintf("invalid amount %q: %v", get(amtCol), err)})
			continue
		}

		category := strings.ToLower(strings.TrimSpace(get(catCol)))
		if category == "" {
			category = "general-goods"
		}
		currency := spendCurrency(get, amtCol)
		start, end := parsePeriod(get)

		metadata := map[string]string{}
		if v := get("vendor"); v != "" {
			metadata["vendor"] = v
		} else if v := get("supplier"); v != "" {
			metadata["vendor"] = v
		}

		activity := ingestion.Activity{
			ID:          fmt.Sprintf("spend-%d", lineNum),
			Source:      "purchases",
			Category:    category,
			Location:    p.DefaultLocation,
			PeriodStart: start,
			PeriodEnd:   end,
			Quantity:    amount,
			Unit:        currency,
			OrgID:       p.DefaultOrgID,
			Metadata:    metadata,
			DataQuality: "estimated", // spend-based is inherently an estimate
			CreatedAt:   time.Now().UTC(),
		}
		if err := activity.Validate(); err != nil {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Message: fmt.Sprintf("validation failed: %v", err)})
			continue
		}
		activities = append(activities, activity)
	}

	return p.finishResult(activities, errors, lineNum, "spend"), nil
}

// =============================================================================
// Scope 3 cat 5 — Waste generated in operations
// =============================================================================

// wasteWeightToKg converts a waste quantity column to kilograms (the Scope 3
// waste calculator's native unit).
func wasteWeightToKg(col string, value float64) float64 {
	switch col {
	case "tonnes", "tonne", "tons", "ton":
		return value * 1000
	default: // kg, kilograms, weight_kg, weight, mass
		return value
	}
}

func (p *UtilityBillParser) parseWasteRows(reader *csv.Reader, colIndex map[string]int, startLine int) (*ParseResult, error) {
	treatCol := firstCol(colIndex, []string{"treatment", "disposal", "disposal_method", "waste_type", "waste_stream"})
	qtyCol := firstCol(colIndex, wasteQtyCols)
	if qtyCol == "" {
		return nil, fmt.Errorf("waste data detected but no weight column (kg/tonnes) found")
	}

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

		raw, err := parseFlexibleNumber(get(qtyCol))
		if err != nil {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Field: qtyCol, Message: fmt.Sprintf("invalid weight %q: %v", get(qtyCol), err)})
			continue
		}
		category := strings.ToLower(strings.TrimSpace(get(treatCol)))
		if category == "" {
			category = "landfill"
		}
		start, end := parsePeriod(get)

		activity := ingestion.Activity{
			ID:          fmt.Sprintf("waste-%d", lineNum),
			Source:      "waste",
			Category:    category,
			Location:    p.DefaultLocation,
			PeriodStart: start,
			PeriodEnd:   end,
			Quantity:    wasteWeightToKg(qtyCol, raw),
			Unit:        "kg",
			OrgID:       p.DefaultOrgID,
			DataQuality: "measured",
			CreatedAt:   time.Now().UTC(),
		}
		if err := activity.Validate(); err != nil {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Message: fmt.Sprintf("validation failed: %v", err)})
			continue
		}
		activities = append(activities, activity)
	}
	return p.finishResult(activities, errors, lineNum, "waste"), nil
}

// =============================================================================
// Scope 3 cat 7 — Employee commuting
// =============================================================================

// normalizeCommuteMode maps a raw commute mode to a category the Scope 3
// commuting factor table recognizes.
func normalizeCommuteMode(raw string) string {
	m := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case strings.Contains(m, "electric"):
		return "car-electric"
	case strings.Contains(m, "hybrid"):
		return "car-hybrid"
	case strings.Contains(m, "diesel"):
		return "car-diesel"
	case strings.Contains(m, "train"), strings.Contains(m, "rail"):
		return "train"
	case strings.Contains(m, "bus"), strings.Contains(m, "transit"), strings.Contains(m, "subway"), strings.Contains(m, "metro"):
		return "public-transit"
	case strings.Contains(m, "bike"), strings.Contains(m, "bicycle"), strings.Contains(m, "cycle"):
		return "bicycle"
	case strings.Contains(m, "walk"):
		return "walking"
	case strings.Contains(m, "wfh"), strings.Contains(m, "remote"), strings.Contains(m, "home"):
		return "work-from-home"
	case m == "":
		return "car-petrol"
	default:
		return m
	}
}

func (p *UtilityBillParser) parseCommutingRows(reader *csv.Reader, colIndex map[string]int, startLine int) (*ParseResult, error) {
	distCol := firstCol(colIndex, distanceCols)
	if distCol == "" {
		return nil, fmt.Errorf("commuting data detected but no distance column found")
	}
	modeCol := firstCol(colIndex, []string{"commute_mode", "mode", "travel_mode", "transport_mode", "type"})
	unit := travelUnit(distCol)

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

		distance, err := parseFlexibleNumber(get(distCol))
		if err != nil {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Field: distCol, Message: fmt.Sprintf("invalid distance %q: %v", get(distCol), err)})
			continue
		}
		// Scale by commuting days and headcount when provided (per passenger-km).
		for _, mult := range []string{"days", "commute_days", "trips", "employees", "headcount"} {
			if v := get(mult); v != "" {
				if n, err := parseFlexibleNumber(v); err == nil && n > 0 {
					distance *= n
				}
			}
		}
		mode := normalizeCommuteMode(get(modeCol))
		start, end := parsePeriod(get)

		activity := ingestion.Activity{
			ID:          fmt.Sprintf("commute-%s-%d", mode, lineNum),
			Source:      "commuting",
			Category:    mode,
			Location:    p.DefaultLocation,
			PeriodStart: start,
			PeriodEnd:   end,
			Quantity:    distance,
			Unit:        unit,
			OrgID:       p.DefaultOrgID,
			DataQuality: "estimated",
			CreatedAt:   time.Now().UTC(),
		}
		if err := activity.Validate(); err != nil {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Message: fmt.Sprintf("validation failed: %v", err)})
			continue
		}
		activities = append(activities, activity)
	}
	return p.finishResult(activities, errors, lineNum, "commuting"), nil
}

// =============================================================================
// Scope 3 cat 4/9 — Transportation & distribution (freight)
// =============================================================================

// normalizeFreightMode maps a raw freight mode to a transport factor key.
func normalizeFreightMode(raw string) string {
	m := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case strings.Contains(m, "air"), strings.Contains(m, "plane"):
		return "air_freight"
	case strings.Contains(m, "ship"), strings.Contains(m, "ocean"), strings.Contains(m, "sea"), strings.Contains(m, "marine"):
		return "ship_freight"
	case strings.Contains(m, "rail"), strings.Contains(m, "train"):
		return "rail_freight"
	default:
		return "truck_freight"
	}
}

func freightMassToTonnes(col string, value float64) float64 {
	switch col {
	case "kg", "kilograms", "weight_kg":
		return value / 1000
	default: // tonnes, tonne, tons, ton, cargo_weight, freight_weight, weight_tonnes, cargo
		return value
	}
}

func (p *UtilityBillParser) parseFreightRows(reader *csv.Reader, colIndex map[string]int, startLine int) (*ParseResult, error) {
	tkmCol := firstCol(colIndex, freightTonneKm)
	massCol := firstCol(colIndex, freightMassCols)
	distCol := firstCol(colIndex, distanceCols)
	if tkmCol == "" && (massCol == "" || distCol == "") {
		return nil, fmt.Errorf("freight data detected but need tonne-km, or weight + distance")
	}
	modeCol := firstCol(colIndex, []string{"mode", "transport_mode", "freight_mode", "type"})

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

		var tonneKm float64
		if tkmCol != "" {
			v, err := parseFlexibleNumber(get(tkmCol))
			if err != nil {
				errors = append(errors, ingestion.ImportError{Row: lineNum, Field: tkmCol, Message: fmt.Sprintf("invalid tonne-km %q: %v", get(tkmCol), err)})
				continue
			}
			tonneKm = v
		} else {
			mass, err1 := parseFlexibleNumber(get(massCol))
			dist, err2 := parseFlexibleNumber(get(distCol))
			if err1 != nil || err2 != nil {
				errors = append(errors, ingestion.ImportError{Row: lineNum, Message: "invalid weight or distance for freight"})
				continue
			}
			tonnes := freightMassToTonnes(massCol, mass)
			// Normalize distance to km (transport factors are per tonne-km).
			if travelUnit(distCol) == "mile" {
				dist *= 1.60934
			}
			tonneKm = tonnes * dist
		}

		mode := normalizeFreightMode(get(modeCol))
		start, end := parsePeriod(get)

		activity := ingestion.Activity{
			ID:          fmt.Sprintf("freight-%s-%d", mode, lineNum),
			Source:      "freight",
			Category:    mode,
			Location:    p.DefaultLocation,
			PeriodStart: start,
			PeriodEnd:   end,
			Quantity:    tonneKm,
			Unit:        "tonne-km",
			OrgID:       p.DefaultOrgID,
			DataQuality: "measured",
			CreatedAt:   time.Now().UTC(),
		}
		if err := activity.Validate(); err != nil {
			errors = append(errors, ingestion.ImportError{Row: lineNum, Message: fmt.Sprintf("validation failed: %v", err)})
			continue
		}
		activities = append(activities, activity)
	}
	return p.finishResult(activities, errors, lineNum, "freight"), nil
}

// applyStrict propagates a sub-parser error and enforces StrictMode, mirroring
// the electricity path's behavior.
func (p *UtilityBillParser) applyStrict(result *ParseResult, err error) (*ParseResult, error) {
	if err != nil {
		return result, err
	}
	if p.StrictMode && result != nil && len(result.Errors) > 0 {
		return result, fmt.Errorf("parsing failed with %d errors in strict mode", len(result.Errors))
	}
	return result, nil
}

// finishResult assembles a ParseResult and honors StrictMode.
func (p *UtilityBillParser) finishResult(activities []ingestion.Activity, errors []ingestion.ImportError, lineNum int, schemaType string) *ParseResult {
	return &ParseResult{
		Activities: activities,
		Errors:     errors,
		Metadata: map[string]string{
			"format":      "csv",
			"total_rows":  strconv.Itoa(lineNum - 1),
			"parsed_rows": strconv.Itoa(len(activities)),
			"error_rows":  strconv.Itoa(len(errors)),
			"schema_type": schemaType,
		},
	}
}
