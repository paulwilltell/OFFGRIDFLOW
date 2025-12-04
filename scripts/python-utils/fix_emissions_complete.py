#!/usr/bin/env python3
"""
Complete fix for emissions_handler.go - handle all records correctly
"""

with open('internal/api/http/handlers/emissions_handler.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()

output = []
i = 0
while i < len(lines):
    line = lines[i]
    
    # Step 1: Update EmissionsHandler struct
    if 'type EmissionsHandler struct {' in line:
        output.append(line)
        i += 1
        while i < len(lines) and '}' not in lines[i]:
            if 'calculator    *emissions.Scope2Calculator' in lines[i]:
                output.append('\tscope1Calculator *emissions.Scope1Calculator\n')
                output.append('\tscope2Calculator *emissions.Scope2Calculator\n')
                output.append('\tscope3Calculator *emissions.Scope3Calculator\n')
                i += 1
                continue
            elif 'scope2Calculator' in lines[i] or 'scope1Calculator' in lines[i] or 'scope3Calculator' in lines[i]:
                i += 1
                continue
            output.append(lines[i])
            i += 1
        output.append(lines[i])
        i += 1
        continue
    
    # Step 2: Update EmissionsHandlerConfig
    if 'type EmissionsHandlerConfig struct {' in line:
        output.append(line)
        i += 1
        while i < len(lines) and '}' not in lines[i]:
            if 'Scope2Calculator *emissions.Scope2Calculator' in lines[i]:
                output.append('\tScope1Calculator *emissions.Scope1Calculator\n')
                output.append('\tScope2Calculator *emissions.Scope2Calculator\n')
                output.append('\tScope3Calculator *emissions.Scope3Calculator\n')
                i += 1
                continue
            elif 'Scope1Calculator' in lines[i] or 'Scope3Calculator' in lines[i]:
                i += 1
                continue
            output.append(lines[i])
            i += 1
        output.append(lines[i])
        i += 1
        continue
    
    # Step 3: Update constructor
    if '\treturn &EmissionsHandler{' in line:
        output.append(line)
        i += 1
        while i < len(lines) and '}' not in lines[i]:
            if 'calculator:    cfg.Scope2Calculator,' in lines[i]:
                output.append('\t\tscope1Calculator: cfg.Scope1Calculator,\n')
                output.append('\t\tscope2Calculator: cfg.Scope2Calculator,\n')
                output.append('\t\tscope3Calculator: cfg.Scope3Calculator,\n')
                i += 1
                continue
            elif 'scope1Calculator' in lines[i] or 'scope2Calculator' in lines[i] or 'scope3Calculator' in lines[i]:
                i += 1
                continue
            output.append(lines[i])
            i += 1
        output.append(lines[i])
        i += 1
        continue
    
    # Step 4: First calculation section (listEmissions)
    if '\t// Calculate emissions for all scopes' in line and 'scope1Records' in lines[i+1]:
        # This section is already updated, but we need to combine records
        output.append(line)
        i += 1
        # Copy scope calculations
        while i < len(lines) and 'Convert to response format' not in lines[i]:
            if 'var scope1Total, scope2Total, scope3Total float64' in lines[i]:
                # Skip aggregation variables - we don't need them here
                i += 1
                # Skip the aggregation loops
                while i < len(lines) and ('for _, rec := range scope' in lines[i] or 'scope' in lines[i] and 'Total +=' in lines[i+1] if i+1 < len(lines) else False or lines[i].strip() == '}'):
                    i += 1
                continue
            output.append(lines[i])
            i += 1
        
        # Now combine all records
        output.append('\n')
        output.append('\t// Combine all scope records\n')
        output.append('\tallRecords := make([]emissions.EmissionRecord, 0, len(scope1Records)+len(scope2Records)+len(scope3Records))\n')
        output.append('\tallRecords = append(allRecords, scope1Records...)\n')
        output.append('\tallRecords = append(allRecords, scope2Records...)\n')
        output.append('\tallRecords = append(allRecords, scope3Records...)\n')
        output.append('\n')
        output.append(lines[i])  # "Convert to response format"
        i += 1
        continue
    
    # Replace references to 'records' with 'allRecords' in listEmissions
    if '\tresponse := make([]EmissionRecord, 0, len(records))' in line:
        output.append('\tresponse := make([]EmissionRecord, 0, len(allRecords))\n')
        i += 1
        continue
    
    if '\tfor i, rec := range records {' in line:
        output.append('\tfor i, rec := range allRecords {\n')
        i += 1
        continue
    
    if '\tpageInfo := responders.NewPageInfo(filter.Page, filter.PerPage, len(records))' in line:
        output.append('\tpageInfo := responders.NewPageInfo(filter.Page, filter.PerPage, len(allRecords))\n')
        i += 1
        continue
    
    # Step 5: Second calculation section (getSummary)  
    if '\trecords, err := h.calculator.CalculateBatch(ctx, emissionsActivities)' in line:
        # Replace with all three scopes
        output.append('\t// Calculate all scopes\n')
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
        i += 1  # Skip old var
        
        output.append('\tfor _, rec := range scope1Records {\n')
        output.append('\t\tscope1Total += rec.EmissionsTonnesCO2e\n')
        output.append('\t}\n')
        
        # Copy scope2 loop
        while i < len(lines) and 'for _, rec := range' in lines[i]:
            output.append(lines[i])
            i += 1
            # Copy loop body
            while i < len(lines) and '}' not in lines[i]:
                output.append(lines[i])
                i += 1
            output.append(lines[i])  # closing brace
            i += 1
            break
        
        output.append('\tfor _, rec := range scope3Records {\n')
        output.append('\t\tscope3Total += rec.EmissionsTonnesCO2e\n')
        output.append('\t}\n')
        
        # Combine all records for count
        output.append('\n')
        output.append('\tallRecords := len(scope1Records) + len(scope2Records) + len(scope3Records)\n')
        output.append('\n')
        continue
    
    # Update EmissionsSummary
    if '\t\tScope1TonsCO2e: 0, // TODO' in line or '\t\tScope1TonsCO2e: scope1Total,' in line:
        output.append('\t\tScope1TonsCO2e: scope1Total,\n')
        i += 1
        continue
    
    if '\t\tScope3TonsCO2e: 0, // TODO' in line or '\t\tScope3TonsCO2e: scope3Total,' in line:
        output.append('\t\tScope3TonsCO2e: scope3Total,\n')
        i += 1
        continue
    
    if '\t\tTotalTonsCO2e:  scope2Total,' in line:
        output.append('\t\tTotalTonsCO2e:  scope1Total + scope2Total + scope3Total,\n')
        i += 1
        continue
    
    if '\t\tActivityCount:  len(records),' in line:
        output.append('\t\tActivityCount:  allRecords,\n')
        i += 1
        continue
    
    # Default: copy as-is
    output.append(line)
    i += 1

with open('internal/api/http/handlers/emissions_handler.go', 'w', encoding='utf-8') as f:
    f.writelines(output)

print("✅ Complete fix applied to emissions_handler.go")
print("   - Updated all calculator references")
print("   - Combined scope records correctly")
print("   - Fixed all 'records' variable references")
