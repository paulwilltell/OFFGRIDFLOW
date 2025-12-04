#!/usr/bin/env python3
"""
Script to integrate Scope 1 and Scope 3 calculators into compliance_handler.go
"""

import re

# Read the file
with open('internal/api/http/handlers/compliance_handler.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Step 1: Update ComplianceHandlerDeps struct to add Scope1 and Scope3 calculators
deps_pattern = r'(type ComplianceHandlerDeps struct \{\s+)Scope2Calculator \*emissions\.Scope2Calculator(\s+\})'
deps_replacement = r'\1Scope1Calculator *emissions.Scope1Calculator\n\tScope2Calculator *emissions.Scope2Calculator\n\tScope3Calculator *emissions.Scope3Calculator\2'
content = re.sub(deps_pattern, deps_replacement, content)

# Step 2: Replace TODO for Scope 1 calculation in CSRD handler (around line 109)
scope1_todo = r'TotalScope1Tons:\s+0,\s+//\s+TODO:\s+Implement Scope 1 calculator \(direct emissions\)'
scope1_fix = 'TotalScope1Tons: scope1TotalTons,'
content = content.replace('TotalScope1Tons: 0, // TODO: Implement Scope 1 calculator (direct emissions)', scope1_fix)

# Step 3: Replace TODO for Scope 3 calculation in CSRD handler (around line 111)
scope3_todo = r'TotalScope3Tons:\s+0,\s+//\s+TODO:\s+Implement Scope 3 calculator \(value chain\)'
scope3_fix = 'TotalScope3Tons: scope3TotalTons,'
content = content.replace('TotalScope3Tons: 0, // TODO: Implement Scope 3 calculator (value chain)', scope3_fix)

# Step 4: Add calculation logic before the report struct (insert before line "var scope2TotalTons")
calc_insertion = '''		// Calculate Scope 1 (direct emissions)
		scope1Records, err := deps.Scope1Calculator.CalculateBatch(ctx, emissionsActivities)
		if err != nil {
			responders.Error(w, http.StatusInternalServerError, "scope1_calc_failed", "failed to calculate scope 1 emissions")
			return
		}

		// Calculate Scope 2 (purchased electricity)
		scope2Records, err := deps.Scope2Calculator.CalculateBatch(ctx, emissionsActivities)
		if err != nil {
			responders.Error(w, http.StatusInternalServerError, "scope2_calc_failed", "failed to calculate scope 2 emissions")
			return
		}

		// Calculate Scope 3 (value chain)
		scope3Records, err := deps.Scope3Calculator.CalculateBatch(ctx, emissionsActivities)
		if err != nil {
			responders.Error(w, http.StatusInternalServerError, "scope3_calc_failed", "failed to calculate scope 3 emissions")
			return
		}

		// Aggregate emissions by scope
		var scope1TotalTons, scope2TotalTons, scope3TotalTons float64
		for _, rec := range scope1Records {
			scope1TotalTons += rec.EmissionsTonnesCO2e
		}
		for _, rec := range scope2Records {
			scope2TotalTons += rec.EmissionsTonnesCO2e
		}
		for _, rec := range scope3Records {
			scope3TotalTons += rec.EmissionsTonnesCO2e
		}'''

# Replace the old Scope 2-only calculation logic
old_calc = '''		scope2Records, err := deps.Scope2Calculator.CalculateBatch(ctx, emissionsActivities)
		if err != nil {
			responders.Error(w, http.StatusInternalServerError, "scope2_calc_failed", "failed to calculate scope 2 emissions")
			return
		}

		var scope2TotalTons float64
		for _, rec := range scope2Records {
			scope2TotalTons += rec.EmissionsTonnesCO2e
		}'''

content = content.replace(old_calc, calc_insertion)

# Step 5: Update compliance summary handler (around line 240-241)
# Replace "scope1Ready": false, // TODO
content = content.replace('"scope1Ready": false, // TODO', '"scope1Ready": scope1Total > 0,')
# Replace "scope3Ready": false, // TODO  
content = content.replace('"scope3Ready": false, // TODO', '"scope3Ready": scope3Total > 0,')

# Step 6: Add calculation in summary handler
summary_calc = '''		// Calculate all scopes
		scope1Records, _ := deps.Scope1Calculator.CalculateBatch(ctx, emissionsActivities)
		scope2Records, _ := deps.Scope2Calculator.CalculateBatch(ctx, emissionsActivities)
		scope3Records, _ := deps.Scope3Calculator.CalculateBatch(ctx, emissionsActivities)

		var scope1Total, scope2Total, scope3Total float64
		for _, rec := range scope1Records {
			scope1Total += rec.EmissionsTonnesCO2e
		}
		for _, rec := range scope2Records {
			scope2Total += rec.EmissionsTonnesCO2e
		}
		for _, rec := range scope3Records {
			scope3Total += rec.EmissionsTonnesCO2e
		}'''

old_summary_calc = '''		scope2Records, _ := deps.Scope2Calculator.CalculateBatch(ctx, emissionsActivities)
		var scope2Total float64
		for _, rec := range scope2Records {
			scope2Total += rec.EmissionsTonnesCO2e
		}'''

content = content.replace(old_summary_calc, summary_calc)

# Write the updated file
with open('internal/api/http/handlers/compliance_handler.go', 'w', encoding='utf-8') as f:
    f.write(content)

print("✅ Successfully integrated Scope 1 and Scope 3 calculators into compliance_handler.go")
print("   - Updated ComplianceHandlerDeps struct")
print("   - Added Scope 1 and Scope 3 calculations to CSRD handler")
print("   - Removed TODO comments")
print("   - Updated compliance summary handler")
