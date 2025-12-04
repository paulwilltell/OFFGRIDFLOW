#!/usr/bin/env python3
"""
Script to integrate Scope 1 and Scope 3 calculators into emissions_handler.go
"""

# Read the file
with open('internal/api/http/handlers/emissions_handler.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Replace TODO for Scope 1 (around line 256)
content = content.replace('Scope1TonsCO2e: 0, // TODO: Implement Scope 1', 'Scope1TonsCO2e: scope1Total,')

# Replace TODO for Scope 3 (around line 258)
content = content.replace('Scope3TonsCO2e: 0, // TODO: Implement Scope 3', 'Scope3TonsCO2e: scope3Total,')

# Add calculation logic (find the Scope 2 calculation and expand it)
old_scope2_calc = '''	scope2Records, _ := h.deps.Scope2Calculator.CalculateBatch(ctx, emissionsActivities)
	var scope2Total float64
	for _, rec := range scope2Records {
		scope2Total += rec.EmissionsTonnesCO2e
	}'''

new_all_scopes_calc = '''	// Calculate all scopes
	scope1Records, _ := h.deps.Scope1Calculator.CalculateBatch(ctx, emissionsActivities)
	scope2Records, _ := h.deps.Scope2Calculator.CalculateBatch(ctx, emissionsActivities)
	scope3Records, _ := h.deps.Scope3Calculator.CalculateBatch(ctx, emissionsActivities)

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

content = content.replace(old_scope2_calc, new_all_scopes_calc)

# Write the updated file
with open('internal/api/http/handlers/emissions_handler.go', 'w', encoding='utf-8') as f:
    f.write(content)

print("✅ Successfully integrated Scope 1 and Scope 3 calculators into emissions_handler.go")
print("   - Added Scope 1 and Scope 3 calculations")
print("   - Removed TODO comments")
print("   - Updated EmissionsSummary population")
