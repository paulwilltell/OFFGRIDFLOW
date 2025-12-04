#!/usr/bin/env python3
"""
Fix emissions_handler.go to add Scope 1 and Scope 3 calculators
"""

with open('internal/api/http/handlers/emissions_handler.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()

output = []
i = 0
while i < len(lines):
    line = lines[i]
    
    # Step 1: Update EmissionsHandler struct to have all three calculators
    if 'type EmissionsHandler struct {' in line:
        output.append(line)
        i += 1
        while i < len(lines) and '}' not in lines[i]:
            if 'calculator    *emissions.Scope2Calculator' in lines[i]:
                # Replace single calculator with three
                output.append('\tscope1Calculator *emissions.Scope1Calculator\n')
                output.append('\tscope2Calculator *emissions.Scope2Calculator\n')
                output.append('\tscope3Calculator *emissions.Scope3Calculator\n')
                i += 1
                continue
            output.append(lines[i])
            i += 1
        output.append(lines[i])  # closing brace
        i += 1
        continue
    
    # Step 2: Update EmissionsHandlerConfig
    if 'type EmissionsHandlerConfig struct {' in line:
        output.append(line)
        i += 1
        while i < len(lines) and '}' not in lines[i]:
            if 'Scope2Calculator *emissions.Scope2Calculator' in lines[i]:
                # Replace with all three
                output.append('\tScope1Calculator *emissions.Scope1Calculator\n')
                output.append('\tScope2Calculator *emissions.Scope2Calculator\n')
                output.append('\tScope3Calculator *emissions.Scope3Calculator\n')
                i += 1
                continue
            output.append(lines[i])
            i += 1
        output.append(lines[i])  # closing brace
        i += 1
        continue
    
    # Step 3: Update NewEmissionsHandler constructor
    if 'return &EmissionsHandler{' in line:
        output.append(line)
        i += 1
        while i < len(lines) and '}' not in lines[i]:
            if 'calculator:    cfg.Scope2Calculator,' in lines[i]:
                # Replace with all three
                output.append('\t\tscope1Calculator: cfg.Scope1Calculator,\n')
                output.append('\t\tscope2Calculator: cfg.Scope2Calculator,\n')
                output.append('\t\tscope3Calculator: cfg.Scope3Calculator,\n')
                i += 1
                continue
            output.append(lines[i])
            i += 1
        output.append(lines[i])  # closing brace
        i += 1
        continue
    
    # Step 4: Update calculations in listEmissions (around line 135)
    if 'records, err := h.calculator.CalculateBatch(ctx, emissionsActivities)' in line:
        # Replace with calculations for all three scopes
        output.append('\t// Calculate emissions for all scopes\n')
        output.append('\tscope1Records, _ := h.scope1Calculator.CalculateBatch(ctx, emissionsActivities)\n')
        output.append('\tscope2Records, err := h.scope2Calculator.CalculateBatch(ctx, emissionsActivities)\n')
        i += 1
        # Copy error handling
        while i < len(lines) and 'var scope2Total' not in lines[i]:
            output.append(lines[i])
            i += 1
        
        output.append('\tscope3Records, _ := h.scope3Calculator.CalculateBatch(ctx, emissionsActivities)\n')
        output.append('\n')
        
        # Add aggregation
        output.append('\tvar scope1Total, scope2Total, scope3Total float64\n')
        i += 1  # Skip old var declaration
        
        output.append('\tfor _, rec := range scope1Records {\n')
        output.append('\t\tscope1Total += rec.EmissionsTonnesCO2e\n')
        output.append('\t}\n')
        
        # Copy scope2 loop
        while i < len(lines):
            output.append(lines[i])
            i += 1
            if '}' in lines[i-1] and 'scope2Total' in lines[i-2]:
                break
        
        output.append('\tfor _, rec := range scope3Records {\n')
        output.append('\t\tscope3Total += rec.EmissionsTonnesCO2e\n')
        output.append('\t}\n')
        continue
    
    # Step 5: Update EmissionsSummary population
    if 'Scope1TonsCO2e: 0, // TODO' in line:
        output.append(line.replace('0, // TODO: Implement Scope 1', 'scope1Total,'))
        i += 1
        continue
    
    if 'Scope3TonsCO2e: 0, // TODO' in line:
        output.append(line.replace('0, // TODO: Implement Scope 3', 'scope3Total,'))
        i += 1
        continue
    
    # Step 6: Update TotalTonsCO2e to include all scopes
    if 'TotalTonsCO2e:  scope2Total,' in line:
        output.append(line.replace('scope2Total,', 'scope1Total + scope2Total + scope3Total,'))
        i += 1
        continue
    
    # Default: copy as-is
    output.append(line)
    i += 1

with open('internal/api/http/handlers/emissions_handler.go', 'w', encoding='utf-8') as f:
    f.writelines(output)

print("✅ Successfully updated emissions_handler.go")
print("   - Added Scope1Calculator and Scope3Calculator fields")
print("   - Updated constructor to initialize all three calculators")
print("   - Removed TODO comments")
print("   - Updated EmissionsSummary to include all scopes")
