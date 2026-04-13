package factors

import (
	"time"

	"github.com/example/offgridflow/internal/emissions"
)

// SeedComprehensiveFactors adds a full set of real-world emission factors to the registry.
// Sources: EPA eGRID 2023, UK DEFRA 2024, IPCC AR6, IEA 2023, GHG Protocol defaults.
// Returns the number of factors seeded.
func SeedComprehensiveFactors(r *InMemoryRegistry) int {
	now := time.Now().UTC()
	count := 0

	// =========================================================================
	// SCOPE 2: EPA eGRID Subregions (US) — 2023 data
	// Source: https://www.epa.gov/egrid
	// Unit: kg CO2e per kWh
	// =========================================================================
	eGridFactors := []struct {
		id, region, source string
		value              float64
	}{
		{"egrid-akgd", "US-AK-AKGD", "ASCC Alaska Grid", 0.437},
		{"egrid-akms", "US-AK-AKMS", "ASCC Miscellaneous", 0.228},
		{"egrid-aznm", "US-AZNM", "WECC Southwest", 0.391},
		{"egrid-camx", "US-CAMX", "WECC California", 0.225},
		{"egrid-erct", "US-ERCT", "ERCOT Texas", 0.373},
		{"egrid-frcc", "US-FRCC", "FRCC Florida", 0.381},
		{"egrid-hims", "US-HI-HIMS", "HICC Miscellaneous", 0.594},
		{"egrid-hioa", "US-HI-HIOA", "HICC Oahu", 0.664},
		{"egrid-mroe", "US-MROE", "MRO East", 0.549},
		{"egrid-mrow", "US-MROW", "MRO West", 0.423},
		{"egrid-newe", "US-NEWE", "NPCC New England", 0.218},
		{"egrid-nwpp", "US-NWPP", "WECC Northwest", 0.252},
		{"egrid-nycw", "US-NYCW", "NPCC NYC/Westchester", 0.227},
		{"egrid-nyli", "US-NYLI", "NPCC Long Island", 0.379},
		{"egrid-nyup", "US-NYUP", "NPCC Upstate NY", 0.095},
		{"egrid-prms", "US-PR", "Puerto Rico", 0.737},
		{"egrid-rfce", "US-RFCE", "RFC East", 0.295},
		{"egrid-rfcm", "US-RFCM", "RFC Michigan", 0.445},
		{"egrid-rfcw", "US-RFCW", "RFC West", 0.462},
		{"egrid-rmpa", "US-RMPA", "WECC Rockies", 0.530},
		{"egrid-spno", "US-SPNO", "SPP North", 0.459},
		{"egrid-spso", "US-SPSO", "SPP South", 0.407},
		{"egrid-srmv", "US-SRMV", "SERC Mississippi Valley", 0.371},
		{"egrid-srmw", "US-SRMW", "SERC Midwest", 0.639},
		{"egrid-srso", "US-SRSO", "SERC South", 0.374},
		{"egrid-srtv", "US-SRTV", "SERC Tennessee Valley", 0.363},
		{"egrid-srvc", "US-SRVC", "SERC Virginia/Carolina", 0.292},
	}

	for _, f := range eGridFactors {
		r.factors[f.id] = emissions.EmissionFactor{
			ID: f.id, Scope: emissions.Scope2, Region: f.region,
			Source: f.source, Unit: "kWh", ValueKgCO2ePerUnit: f.value,
			Method: emissions.MethodLocationBased, DataSource: "EPA eGRID 2023",
			CreatedAt: now,
		}
		count++
	}

	// =========================================================================
	// SCOPE 2: International Grid Factors — IEA 2023
	// =========================================================================
	intlGridFactors := []struct {
		id, region, source string
		value              float64
	}{
		// Europe
		{"grid-at", "AT", "Austria", 0.087},
		{"grid-be", "BE", "Belgium", 0.135},
		{"grid-bg", "BG", "Bulgaria", 0.380},
		{"grid-hr", "HR", "Croatia", 0.140},
		{"grid-cy", "CY", "Cyprus", 0.583},
		{"grid-cz", "CZ", "Czechia", 0.381},
		{"grid-dk", "DK", "Denmark", 0.112},
		{"grid-ee", "EE", "Estonia", 0.376},
		{"grid-fi", "FI", "Finland", 0.075},
		{"grid-de", "DE", "Germany", 0.366},
		{"grid-gr", "GR", "Greece", 0.333},
		{"grid-hu", "HU", "Hungary", 0.208},
		{"grid-ie", "IE", "Ireland", 0.296},
		{"grid-it", "IT", "Italy", 0.258},
		{"grid-lv", "LV", "Latvia", 0.082},
		{"grid-lt", "LT", "Lithuania", 0.030},
		{"grid-lu", "LU", "Luxembourg", 0.063},
		{"grid-mt", "MT", "Malta", 0.378},
		{"grid-nl", "NL", "Netherlands", 0.300},
		{"grid-no", "NO", "Norway", 0.008},
		{"grid-pl", "PL", "Poland", 0.614},
		{"grid-pt", "PT", "Portugal", 0.148},
		{"grid-ro", "RO", "Romania", 0.249},
		{"grid-sk", "SK", "Slovakia", 0.104},
		{"grid-si", "SI", "Slovenia", 0.215},
		{"grid-es", "ES", "Spain", 0.150},
		{"grid-se", "SE", "Sweden", 0.010},
		{"grid-ch", "CH", "Switzerland", 0.013},
		{"grid-uk", "GB", "United Kingdom", 0.193},
		{"grid-fr", "FR", "France", 0.051},
		// Americas
		{"grid-ca", "CA", "Canada", 0.120},
		{"grid-mx", "MX", "Mexico", 0.404},
		{"grid-br", "BR", "Brazil", 0.062},
		{"grid-ar", "AR", "Argentina", 0.311},
		{"grid-cl", "CL", "Chile", 0.333},
		{"grid-co", "CO", "Colombia", 0.126},
		// Asia Pacific
		{"grid-cn", "CN", "China", 0.555},
		{"grid-in", "IN", "India", 0.708},
		{"grid-jp", "JP", "Japan", 0.457},
		{"grid-kr", "KR", "South Korea", 0.415},
		{"grid-au", "AU", "Australia", 0.656},
		{"grid-nz", "NZ", "New Zealand", 0.082},
		{"grid-sg", "SG", "Singapore", 0.408},
		{"grid-th", "TH", "Thailand", 0.466},
		{"grid-id", "ID", "Indonesia", 0.718},
		{"grid-my", "MY", "Malaysia", 0.558},
		{"grid-ph", "PH", "Philippines", 0.503},
		{"grid-vn", "VN", "Vietnam", 0.462},
		{"grid-tw", "TW", "Taiwan", 0.495},
		// Middle East & Africa
		{"grid-za", "ZA", "South Africa", 0.928},
		{"grid-ae", "AE", "UAE", 0.404},
		{"grid-sa", "SA", "Saudi Arabia", 0.553},
		{"grid-eg", "EG", "Egypt", 0.424},
		{"grid-ng", "NG", "Nigeria", 0.380},
		{"grid-ke", "KE", "Kenya", 0.056},
		{"grid-il", "IL", "Israel", 0.437},
	}

	for _, f := range intlGridFactors {
		r.factors[f.id] = emissions.EmissionFactor{
			ID: f.id, Scope: emissions.Scope2, Region: f.region,
			Source: f.source, Unit: "kWh", ValueKgCO2ePerUnit: f.value,
			Method: emissions.MethodLocationBased, DataSource: "IEA 2023",
			CreatedAt: now,
		}
		count++
	}

	// =========================================================================
	// SCOPE 1: Stationary Combustion — DEFRA 2024 / EPA AP-42
	// Unit: kg CO2e per unit specified
	// =========================================================================
	scope1Fuels := []struct {
		id, source, unit string
		value            float64
	}{
		{"fuel-natural-gas-kwh", "Natural Gas", "kWh", 0.18293},
		{"fuel-natural-gas-therm", "Natural Gas", "therm", 5.3058},
		{"fuel-natural-gas-m3", "Natural Gas", "m3", 2.0199},
		{"fuel-natural-gas-mmbtu", "Natural Gas", "MMBtu", 53.058},
		{"fuel-diesel-liter", "Diesel", "liter", 2.6885},
		{"fuel-diesel-gallon", "Diesel", "gallon", 10.177},
		{"fuel-gasoline-liter", "Gasoline/Petrol", "liter", 2.3148},
		{"fuel-gasoline-gallon", "Gasoline/Petrol", "gallon", 8.765},
		{"fuel-lpg-liter", "LPG/Propane", "liter", 1.5570},
		{"fuel-lpg-gallon", "LPG/Propane", "gallon", 5.893},
		{"fuel-fuel-oil-liter", "Fuel Oil (#2)", "liter", 2.7551},
		{"fuel-fuel-oil-gallon", "Fuel Oil (#2)", "gallon", 10.430},
		{"fuel-coal-kg", "Coal (bituminous)", "kg", 2.4532},
		{"fuel-coal-ton", "Coal (bituminous)", "short_ton", 2226.4},
		{"fuel-kerosene-liter", "Kerosene/Jet Fuel", "liter", 2.5408},
		{"fuel-kerosene-gallon", "Kerosene/Jet Fuel", "gallon", 9.616},
		{"fuel-wood-kg", "Wood/Biomass", "kg", 0.01543},
		{"fuel-ethanol-liter", "Ethanol (E100)", "liter", 0.00151},
		{"fuel-biodiesel-liter", "Biodiesel (B100)", "liter", 0.00098},
	}

	for _, f := range scope1Fuels {
		r.factors[f.id] = emissions.EmissionFactor{
			ID: f.id, Scope: emissions.Scope1, Region: "GLOBAL",
			Source: f.source, Unit: f.unit, ValueKgCO2ePerUnit: f.value,
			Method: emissions.MethodDefault, DataSource: "UK DEFRA 2024 / EPA AP-42",
			CreatedAt: now,
		}
		count++
	}

	// =========================================================================
	// SCOPE 1: Mobile Combustion — Vehicle Fleet
	// =========================================================================
	mobileFuels := []struct {
		id, source, unit string
		value            float64
	}{
		{"mobile-car-gasoline-km", "Passenger Car (gasoline)", "km", 0.17060},
		{"mobile-car-gasoline-mi", "Passenger Car (gasoline)", "mile", 0.27452},
		{"mobile-car-diesel-km", "Passenger Car (diesel)", "km", 0.16789},
		{"mobile-car-diesel-mi", "Passenger Car (diesel)", "mile", 0.27017},
		{"mobile-car-hybrid-km", "Passenger Car (hybrid)", "km", 0.11480},
		{"mobile-car-ev-km", "Passenger Car (BEV)", "km", 0.0},
		{"mobile-van-km", "Light Commercial Van", "km", 0.24098},
		{"mobile-van-mi", "Light Commercial Van", "mile", 0.38787},
		{"mobile-hgv-km", "Heavy Goods Vehicle", "km", 0.57262},
		{"mobile-hgv-mi", "Heavy Goods Vehicle", "mile", 0.92133},
	}

	for _, f := range mobileFuels {
		r.factors[f.id] = emissions.EmissionFactor{
			ID: f.id, Scope: emissions.Scope1, Region: "GLOBAL",
			Source: f.source, Unit: f.unit, ValueKgCO2ePerUnit: f.value,
			Method: emissions.MethodDefault, DataSource: "UK DEFRA 2024",
			CreatedAt: now,
		}
		count++
	}

	// =========================================================================
	// SCOPE 1: Refrigerant Leakage (F-Gases)
	// =========================================================================
	refrigerants := []struct {
		id, source string
		gwp        float64
	}{
		{"ref-r134a", "R-134a (HFC)", 1430},
		{"ref-r410a", "R-410A (HFC blend)", 2088},
		{"ref-r404a", "R-404A (HFC blend)", 3922},
		{"ref-r407c", "R-407C (HFC blend)", 1774},
		{"ref-r32", "R-32 (HFC)", 675},
		{"ref-r290", "R-290 (Propane)", 3},
		{"ref-r744", "R-744 (CO2)", 1},
		{"ref-sf6", "SF6", 22800},
	}

	for _, f := range refrigerants {
		r.factors[f.id] = emissions.EmissionFactor{
			ID: f.id, Scope: emissions.Scope1, Region: "GLOBAL",
			Source: f.source, Unit: "kg", ValueKgCO2ePerUnit: f.gwp,
			Method: emissions.MethodDefault, DataSource: "IPCC AR6 GWP-100",
			CreatedAt: now,
		}
		count++
	}

	// =========================================================================
	// SCOPE 3: Category Factors — GHG Protocol / DEFRA 2024
	// =========================================================================
	scope3Factors := []struct {
		id, source, unit string
		value            float64
	}{
		// Category 1: Purchased Goods (spend-based EEIO)
		{"s3-purchased-goods-usd", "Purchased Goods & Services (avg)", "USD", 0.00042},
		{"s3-it-equipment-usd", "IT Equipment & Electronics", "USD", 0.00076},
		{"s3-office-supplies-usd", "Office Supplies", "USD", 0.00031},
		{"s3-food-beverage-usd", "Food & Beverage", "USD", 0.00068},
		{"s3-chemicals-usd", "Chemicals", "USD", 0.00095},
		{"s3-metals-usd", "Metals & Minerals", "USD", 0.00112},
		{"s3-textiles-usd", "Textiles & Clothing", "USD", 0.00058},
		{"s3-paper-usd", "Paper & Printing", "USD", 0.00089},
		{"s3-construction-usd", "Construction Materials", "USD", 0.00073},
		// Category 4 & 9: Transportation
		{"s3-freight-road-tkm", "Road Freight", "tonne-km", 0.10759},
		{"s3-freight-rail-tkm", "Rail Freight", "tonne-km", 0.02774},
		{"s3-freight-sea-tkm", "Sea Freight (container)", "tonne-km", 0.01604},
		{"s3-freight-air-tkm", "Air Freight", "tonne-km", 0.60230},
		// Category 5: Waste
		{"s3-waste-landfill-kg", "Waste to Landfill", "kg", 0.58670},
		{"s3-waste-recycling-kg", "Waste Recycled", "kg", 0.02140},
		{"s3-waste-incineration-kg", "Waste Incinerated", "kg", 0.91120},
		{"s3-waste-composting-kg", "Organic Waste Composted", "kg", 0.01050},
		// Category 6: Business Travel
		{"s3-flight-domestic-pkm", "Domestic Flight", "passenger-km", 0.24587},
		{"s3-flight-shorthaul-pkm", "Short-Haul Flight (<3700km)", "passenger-km", 0.15553},
		{"s3-flight-longhaul-pkm", "Long-Haul Flight (>3700km)", "passenger-km", 0.19309},
		{"s3-rail-pkm", "Rail Travel", "passenger-km", 0.03549},
		{"s3-hotel-night", "Hotel Stay", "room-night", 20.6},
		// Category 7: Employee Commuting
		{"s3-commute-car-km", "Car Commute", "km", 0.17060},
		{"s3-commute-bus-km", "Bus Commute", "km", 0.08920},
		{"s3-commute-rail-km", "Rail Commute", "km", 0.03549},
		{"s3-commute-wfh-day", "Work From Home", "day", 0.59},
		// Category 8: Upstream Leased Assets
		{"s3-office-m2-yr", "Office Space", "m2-year", 57.3},
		// Category 11: Use of Sold Products
		{"s3-electricity-sold-kwh", "Electricity (sold product use)", "kWh", 0.417},
		// Category 12: End of Life
		{"s3-eol-electronics-kg", "Electronics End-of-Life", "kg", 1.450},
		{"s3-eol-packaging-kg", "Packaging End-of-Life", "kg", 0.370},
		// Water
		{"s3-water-supply-m3", "Water Supply", "m3", 0.149},
		{"s3-water-treatment-m3", "Water Treatment", "m3", 0.272},
	}

	for _, f := range scope3Factors {
		r.factors[f.id] = emissions.EmissionFactor{
			ID: f.id, Scope: emissions.Scope3, Region: "GLOBAL",
			Source: f.source, Unit: f.unit, ValueKgCO2ePerUnit: f.value,
			Method: emissions.MethodDefault, DataSource: "UK DEFRA 2024 / GHG Protocol",
			CreatedAt: now,
		}
		count++
	}

	// =========================================================================
	// SCOPE 2: Thermal Energy
	// =========================================================================
	thermalFactors := []struct {
		id, source, unit string
		value            float64
	}{
		{"thermal-steam-kwh", "Steam", "kWh", 0.18544},
		{"thermal-steam-mmbtu", "Steam", "MMBtu", 54.378},
		{"thermal-district-heat-kwh", "District Heating", "kWh", 0.16600},
		{"thermal-district-cool-kwh", "District Cooling", "kWh", 0.10400},
		{"thermal-chilled-water-kwh", "Chilled Water", "kWh", 0.09800},
	}

	for _, f := range thermalFactors {
		r.factors[f.id] = emissions.EmissionFactor{
			ID: f.id, Scope: emissions.Scope2, Region: "GLOBAL",
			Source: f.source, Unit: f.unit, ValueKgCO2ePerUnit: f.value,
			Method: emissions.MethodLocationBased, DataSource: "EPA 2023 / IEA",
			CreatedAt: now,
		}
		count++
	}

	return count
}
