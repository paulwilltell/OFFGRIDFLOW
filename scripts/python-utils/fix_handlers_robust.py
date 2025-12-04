#!/usr/bin/env python3
"""
Script to integrate Scope 1 and Scope 3 calculators - ROBUST VERSION
"""

import re

def update_compliance_handler():
    with open('internal/api/http/handlers/compliance_handler.go', 'r', encoding='utf-8') as f:
        lines = f.readlines()
    
    output = []
    i = 0
    while i < len(lines):
        line = lines[i]
        
        # Step 1: Update ComplianceHandlerDeps struct
        if 'type ComplianceHandlerDeps struct {' in line:
            output.append(line)
            i += 1
            # Add all fields
            found_scope2 = False
            while i < len(lines) and '}' not in lines[i]:
                if 'Scope2Calculator' in lines[i]:
                    # Add Scope1 before Scope2
                    output.append('\tScope1Calculator *emissions.Scope1Calculator\n')
                    output.append(lines[i])
                    # Add Scope3 after Scope2
                    i += 1
                    output.append('\tScope3Calculator *emissions.Scope3Calculator\n')
                    found_scope2 = True
                    continue
                output.append(lines[i])
                i += 1
            output.append(lines[i])  # closing brace
            i += 1
            continue
        
        # Step 2: Replace Scope 2-only calculation with all three scopes
        if 'scope2Records, err := deps.Scope2Calculator.CalculateBatch' in line:
            # Replace the whole scope2 calculation block with all three scopes
            output.append('\t\t// Calculate Scope 1 (direct emissions)\n')
            output.append('\t\tscope1Records, err := deps.Scope1Calculator.CalculateBatch(ctx, emissionsActivities)\n')
            output.append('\t\tif err != nil {\n')
            output.append('\t\t\tresponders.Error(w, http.StatusInternalServerError, "scope1_calc_failed", "failed to calculate scope 1 emissions")\n')
            output.append('\t\t\treturn\n')
            output.append('\t\t}\n')
            output.append('\n')
            output.append('\t\t// Calculate Scope 2 (purchased electricity)\n')
            output.append('\t\tscope2Records, err := deps.Scope2Calculator.CalculateBatch(ctx, emissionsActivities)\n')
            # Skip the original scope2 line
            i += 1
            # Copy error handling
            while i < len(lines) and 'var scope2TotalTons' not in lines[i]:
                output.append(lines[i])
                i += 1
            
            # Add Scope 3 calculation
            output.append('\n')
            output.append('\t\t// Calculate Scope 3 (value chain)\n')
            output.append('\t\tscope3Records, err := deps.Scope3Calculator.CalculateBatch(ctx, emissionsActivities)\n')
            output.append('\t\tif err != nil {\n')
            output.append('\t\t\tresponders.Error(w, http.StatusInternalServerError, "scope3_calc_failed", "failed to calculate scope 3 emissions")\n')
            output.append('\t\t\treturn\n')
            output.append('\t\t}\n')
            output.append('\n')
            
            # Add aggregation for all three scopes
            output.append('\t\t// Aggregate emissions by scope\n')
            output.append('\t\tvar scope1TotalTons, scope2TotalTons, scope3TotalTons float64\n')
            # Skip old var line
            i += 1
            # Add aggregation loops
            output.append('\t\tfor _, rec := range scope1Records {\n')
            output.append('\t\t\tscope1TotalTons += rec.EmissionsTonnesCO2e\n')
            output.append('\t\t}\n')
            # Copy the existing scope2 loop
            while i < len(lines) and 'for _, rec := range scope2Records' not in lines[i]:
                i += 1
            while i < len(lines):
                output.append(lines[i])
                i += 1
                if '}' in lines[i-1] and 'EmissionsTonnesCO2e' in lines[i-2]:
                    break
            # Add scope3 loop
            output.append('\t\tfor _, rec := range scope3Records {\n')
            output.append('\t\t\tscope3TotalTons += rec.EmissionsTonnesCO2e\n')
            output.append('\t\t}\n')
            continue
        
        # Step 3: Remove TODOs in report struct
        if 'TotalScope1Tons: 0, // TODO' in line:
            output.append(line.replace('0, // TODO: Implement Scope 1 calculator (direct emissions)', 'scope1TotalTons,'))
            i += 1
            continue
        
        if 'TotalScope3Tons: 0, // TODO' in line:
            output.append(line.replace('0, // TODO: Implement Scope 3 calculator (value chain)', 'scope3TotalTons,'))
            i += 1
            continue
        
        # Step 4: Update compliance summary calculations
        if '"scope1Ready": false, // TODO' in line:
            output.append(line.replace('"scope1Ready": false, // TODO', '"scope1Ready": scope1Total > 0,'))
            i += 1
            continue
        
        if '"scope3Ready": false, // TODO' in line:
            output.append(line.replace('"scope3Ready": false, // TODO', '"scope3Ready": scope3Total > 0,'))
            i += 1
            continue
        
        # For the summary handler, expand scope2 to all scopes
        if i < len(lines) - 5 and 'scope2Records, _ := deps.Scope2Calculator.CalculateBatch' in line and 'var scope2Total float64' in lines[i+1]:
            # This is in the summary handler - replace with all three scopes
            output.append('\t\t// Calculate all scopes\n')
            output.append('\t\tscope1Records, _ := deps.Scope1Calculator.CalculateBatch(ctx, emissionsActivities)\n')
            output.append('\t\tscope2Records, _ := deps.Scope2Calculator.CalculateBatch(ctx, emissionsActivities)\n')
            output.append('\t\tscope3Records, _ := deps.Scope3Calculator.CalculateBatch(ctx, emissionsActivities)\n')
            output.append('\n')
            output.append('\t\tvar scope1Total, scope2Total, scope3Total float64\n')
            # Skip original lines
            i += 2
            # Add all three loops
            output.append('\t\tfor _, rec := range scope1Records {\n')
            output.append('\t\t\tscope1Total += rec.EmissionsTonnesCO2e\n')
            output.append('\t\t}\n')
            # Copy scope2 loop
            while i < len(lines):
                output.append(lines[i])
                i += 1
                if '}' in lines[i-1] and 'scope2Total' in lines[i-2]:
                    break
            output.append('\t\tfor _, rec := range scope3Records {\n')
            output.append('\t\t\tscope3Total += rec.EmissionsTonnesCO2e\n')
            output.append('\t\t}\n')
            continue
        
        # Default: copy line as-is
        output.append(line)
        i += 1
    
    with open('internal/api/http/handlers/compliance_handler.go', 'w', encoding='utf-8') as f:
        f.writelines(output)
    
    print("✅ Updated compliance_handler.go")

def update_emissions_handler():
    with open('internal/api/http/handlers/emissions_handler.go', 'r', encoding='utf-8') as f:
        lines = f.readlines()
    
    output = []
    i = 0
    while i < len(lines):
        line = lines[i]
        
        # Replace scope2-only calculation with all three scopes
        if 'scope2Records, _ := h.deps.Scope2Calculator.CalculateBatch' in line:
            output.append('\t// Calculate all scopes\n')
            output.append('\tscope1Records, _ := h.deps.Scope1Calculator.CalculateBatch(ctx, emissionsActivities)\n')
            output.append('\tscope2Records, _ := h.deps.Scope2Calculator.CalculateBatch(ctx, emissionsActivities)\n')
            output.append('\tscope3Records, _ := h.deps.Scope3Calculator.CalculateBatch(ctx, emissionsActivities)\n')
            output.append('\n')
            # Skip original line
            i += 1
            # Replace var declaration
            if 'var scope2Total float64' in lines[i]:
                output.append('\tvar scope1Total, scope2Total, scope3Total float64\n')
                i += 1
            # Add scope1 loop
            output.append('\tfor _, rec := range scope1Records {\n')
            output.append('\t\tscope1Total += rec.EmissionsTonnesCO2e\n')
            output.append('\t}\n')
            # Copy scope2 loop
            while i < len(lines):
                output.append(lines[i])
                i += 1
                if '}' in lines[i-1] and 'scope2Total' in lines[i-2]:
                    break
            # Add scope3 loop
            output.append('\tfor _, rec := range scope3Records {\n')
            output.append('\t\tscope3Total += rec.EmissionsTonnesCO2e\n')
            output.append('\t}\n')
            continue
        
        # Remove TODOs in EmissionsSummary
        if 'Scope1TonsCO2e: 0, // TODO' in line:
            output.append(line.replace('0, // TODO: Implement Scope 1', 'scope1Total,'))
            i += 1
            continue
        
        if 'Scope3TonsCO2e: 0, // TODO' in line:
            output.append(line.replace('0, // TODO: Implement Scope 3', 'scope3Total,'))
            i += 1
            continue
        
        # Default: copy line as-is
        output.append(line)
        i += 1
    
    with open('internal/api/http/handlers/emissions_handler.go', 'w', encoding='utf-8') as f:
        f.writelines(output)
    
    print("✅ Updated emissions_handler.go")

if __name__ == '__main__':
    update_compliance_handler()
    update_emissions_handler()
    print("\n🎉 All handlers updated successfully!")
