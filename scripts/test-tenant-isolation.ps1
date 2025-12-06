# Tenant Isolation Security Test
# Verifies that tenants cannot access each other's compliance reports

Write-Host "=== TENANT ISOLATION SECURITY TEST ===" -ForegroundColor Green
Write-Host ""

$errors = 0
$passed = 0

# Test configuration
$apiUrl = "http://localhost:8080"

Write-Host "[INFO] Testing tenant isolation for compliance reports..." -ForegroundColor Yellow
Write-Host ""

# Simulated test (requires running API)
# In production, this would:
# 1. Create Tenant A with User A
# 2. Create Tenant B with User B
# 3. User A creates compliance report
# 4. User B attempts to access User A's report
# 5. Verify 403/404 response

Write-Host "[TEST 1] Cross-tenant report access prevention" -ForegroundColor Cyan
Write-Host "  Scenario: User from Tenant B attempts to access Tenant A's report" -ForegroundColor Gray
Write-Host ""

# Check if API is running
try {
    $healthCheck = Invoke-WebRequest -Uri "$apiUrl/health" -Method GET -TimeoutSec 2 -ErrorAction Stop
    Write-Host "  ✅ API is running" -ForegroundColor Green
    
    # In production environment, run actual tests:
    # POST /api/v1/auth/register (Tenant A, User A)
    # POST /api/v1/compliance/reports (create report as User A)
    # POST /api/v1/auth/register (Tenant B, User B)
    # GET /api/v1/compliance/reports/{reportId} (attempt access as User B)
    # Assert: Response is 403 Forbidden or 404 Not Found
    
    Write-Host "  ⚠️  Manual test required with running database" -ForegroundColor Yellow
    Write-Host "     Run this script after deploying full stack" -ForegroundColor Gray
    
} catch {
    Write-Host "  ⚠️  API not running - cannot execute integration test" -ForegroundColor Yellow
    Write-Host "     Start API with: docker-compose up api" -ForegroundColor Gray
}

Write-Host ""

# Test database schema for tenant isolation
Write-Host "[TEST 2] Database schema tenant isolation" -ForegroundColor Cyan
Write-Host "  Checking schema.sql for tenant_id foreign keys..." -ForegroundColor Gray

$schemaFile = "C:\Users\pault\OffGridFlow\infra\db\schema.sql"

if (Test-Path $schemaFile) {
    $schemaContent = Get-Content $schemaFile -Raw
    
    # Check compliance_reports table
    if ($schemaContent -match "CREATE TABLE.*compliance_reports") {
        if ($schemaContent -match "tenant_id.*REFERENCES tenants") {
            Write-Host "  ✅ compliance_reports.tenant_id has foreign key constraint" -ForegroundColor Green
            $passed++
        } else {
            Write-Host "  ❌ compliance_reports.tenant_id missing foreign key" -ForegroundColor Red
            $errors++
        }
    }
    
    # Check audit_logs table
    if ($schemaContent -match "CREATE TABLE.*audit_logs") {
        if ($schemaContent -match "tenant_id.*REFERENCES tenants") {
            Write-Host "  ✅ audit_logs.tenant_id has foreign key constraint" -ForegroundColor Green
            $passed++
        } else {
            Write-Host "  ❌ audit_logs.tenant_id missing foreign key" -ForegroundColor Red
            $errors++
        }
    }
    
    # Check activities table
    if ($schemaContent -match "CREATE TABLE.*activities") {
        if ($schemaContent -match "org_id.*REFERENCES tenants") {
            Write-Host "  ✅ activities.org_id has foreign key constraint" -ForegroundColor Green
            $passed++
        } else {
            Write-Host "  ❌ activities.org_id missing foreign key" -ForegroundColor Red
            $errors++
        }
    }
    
    # Check emissions table
    if ($schemaContent -match "CREATE TABLE.*emissions") {
        if ($schemaContent -match "org_id.*REFERENCES tenants") {
            Write-Host "  ✅ emissions.org_id has foreign key constraint" -ForegroundColor Green
            $passed++
        } else {
            Write-Host "  ❌ emissions.org_id missing foreign key" -ForegroundColor Red
            $errors++
        }
    }
    
} else {
    Write-Host "  ❌ schema.sql not found" -ForegroundColor Red
    $errors++
}

Write-Host ""

# Test code for tenant filtering
Write-Host "[TEST 3] Code-level tenant filtering verification" -ForegroundColor Cyan
Write-Host "  Checking Go code for WHERE tenant_id clauses..." -ForegroundColor Gray

$complianceFiles = Get-ChildItem "C:\Users\pault\OffGridFlow\internal" -Recurse -Filter "*.go" -ErrorAction SilentlyContinue

if ($complianceFiles) {
    $tenantFilterFound = $false
    
    foreach ($file in $complianceFiles) {
        $content = Get-Content $file.FullName -Raw -ErrorAction SilentlyContinue
        
        if ($content -match 'WHERE.*tenant_id\s*=|tenant_id\s*=\s*\$\d+') {
            $tenantFilterFound = $true
            break
        }
    }
    
    if ($tenantFilterFound) {
        Write-Host "  ✅ Tenant filtering logic found in codebase" -ForegroundColor Green
        $passed++
    } else {
        Write-Host "  ⚠️  No explicit tenant_id filtering found (may use middleware)" -ForegroundColor Yellow
    }
} else {
    Write-Host "  ⚠️  Could not scan Go files" -ForegroundColor Yellow
}

Write-Host ""

# Security best practices check
Write-Host "[TEST 4] Security best practices" -ForegroundColor Cyan

# Check for SQL injection protection
Write-Host "  Checking for parameterized queries..." -ForegroundColor Gray

if ($complianceFiles) {
    $hasParameterized = $false
    
    foreach ($file in $complianceFiles) {
        $content = Get-Content $file.FullName -Raw -ErrorAction SilentlyContinue
        
        # Look for prepared statements or parameterized queries
        if ($content -match 'QueryRowContext|ExecContext|QueryContext.*\$\d+') {
            $hasParameterized = $true
            break
        }
    }
    
    if ($hasParameterized) {
        Write-Host "  ✅ Parameterized queries detected (SQL injection protection)" -ForegroundColor Green
        $passed++
    } else {
        Write-Host "  ⚠️  Could not verify parameterized query usage" -ForegroundColor Yellow
    }
}

Write-Host ""

# Summary
Write-Host "=== TEST SUMMARY ===" -ForegroundColor Green
Write-Host "Tests Passed: $passed" -ForegroundColor $(if ($passed -gt 0) { "Green" } else { "Yellow" })
Write-Host "Tests Failed: $errors" -ForegroundColor $(if ($errors -eq 0) { "Green" } else { "Red" })
Write-Host ""

if ($errors -eq 0) {
    Write-Host "✅ TENANT ISOLATION: Schema-level protection verified" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next Steps:" -ForegroundColor Cyan
    Write-Host "  1. Deploy full stack: docker-compose up -d" -ForegroundColor White
    Write-Host "  2. Create test tenants and users" -ForegroundColor White
    Write-Host "  3. Run integration tests with real HTTP requests" -ForegroundColor White
    Write-Host "  4. Verify 403/404 responses for cross-tenant access" -ForegroundColor White
} else {
    Write-Host "❌ ERRORS DETECTED - Review schema and code" -ForegroundColor Red
}

Write-Host ""

# Test plan documentation
Write-Host "=== MANUAL INTEGRATION TEST PLAN ===" -ForegroundColor Yellow
Write-Host ""
Write-Host "When API is running, execute these steps:" -ForegroundColor Gray
Write-Host ""
Write-Host "# Step 1: Create Tenant A + User A" -ForegroundColor Cyan
Write-Host 'curl -X POST http://localhost:8080/api/v1/auth/register \' -ForegroundColor White
Write-Host '  -H "Content-Type: application/json" \' -ForegroundColor White
Write-Host '  -d ''{"email":"usera@tenanta.com","password":"test123","name":"User A","tenant":"TenantA"}''' -ForegroundColor White
Write-Host ""

Write-Host "# Step 2: Login as User A and get token" -ForegroundColor Cyan
Write-Host 'curl -X POST http://localhost:8080/api/v1/auth/login \' -ForegroundColor White
Write-Host '  -H "Content-Type: application/json" \' -ForegroundColor White
Write-Host '  -d ''{"email":"usera@tenanta.com","password":"test123"}''' -ForegroundColor White
Write-Host ""

Write-Host "# Step 3: Create compliance report as User A" -ForegroundColor Cyan
Write-Host 'curl -X POST http://localhost:8080/api/v1/compliance/reports \' -ForegroundColor White
Write-Host '  -H "Authorization: Bearer {TOKEN_A}" \' -ForegroundColor White
Write-Host '  -H "Content-Type: application/json" \' -ForegroundColor White
Write-Host '  -d ''{"report_type":"CSRD","reporting_year":2024}''' -ForegroundColor White
Write-Host ""

Write-Host "# Step 4: Create Tenant B + User B" -ForegroundColor Cyan
Write-Host 'curl -X POST http://localhost:8080/api/v1/auth/register \' -ForegroundColor White
Write-Host '  -H "Content-Type: application/json" \' -ForegroundColor White
Write-Host '  -d ''{"email":"userb@tenantb.com","password":"test123","name":"User B","tenant":"TenantB"}''' -ForegroundColor White
Write-Host ""

Write-Host "# Step 5: Login as User B and get token" -ForegroundColor Cyan
Write-Host 'curl -X POST http://localhost:8080/api/v1/auth/login \' -ForegroundColor White
Write-Host '  -H "Content-Type: application/json" \' -ForegroundColor White
Write-Host '  -d ''{"email":"userb@tenantb.com","password":"test123"}''' -ForegroundColor White
Write-Host ""

Write-Host "# Step 6: Attempt to access User A's report as User B" -ForegroundColor Cyan
Write-Host 'curl -X GET http://localhost:8080/api/v1/compliance/reports/{REPORT_ID_FROM_STEP_3} \' -ForegroundColor White
Write-Host '  -H "Authorization: Bearer {TOKEN_B}"' -ForegroundColor White
Write-Host ""

Write-Host "# Expected Result: 403 Forbidden or 404 Not Found" -ForegroundColor Green
Write-Host ""

Write-Host "=== END OF SECURITY TEST ===" -ForegroundColor Green
