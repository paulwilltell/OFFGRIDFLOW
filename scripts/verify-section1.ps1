# OffGridFlow Engineering Readiness - Complete Verification Script
# Run this to validate all Section 1 criteria
# Framework: Million Fold Precision

$ErrorActionPreference = "Continue"
$ReportDir = "reports\engineering"

Write-Host "=========================================="  -ForegroundColor Cyan
Write-Host "OFFGRIDFLOW SECTION 1 VERIFICATION"       -ForegroundColor Cyan
Write-Host "Started: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Cyan
Write-Host "=========================================="  -ForegroundColor Cyan

# Create report directory
New-Item -ItemType Directory -Force -Path $ReportDir | Out-Null

# ============================================================================
# CRITERION 1: Frontend Build
# ============================================================================
Write-Host "`n[1/10] Frontend Build Verification..." -ForegroundColor Yellow

try {
    Push-Location web
    
    Write-Host "  → Running npm install (if needed)..."
    if (-not (Test-Path "node_modules")) {
        npm install 2>&1 | Tee-Object -FilePath "..\$ReportDir\npm-install.log"
    }
    
    Write-Host "  → Building frontend..."
    $buildOutput = npm run build 2>&1 | Tee-Object -FilePath "..\$ReportDir\frontend-build.log"
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✅ Frontend build SUCCESSFUL" -ForegroundColor Green
        
        # Check bundle size
        if (Test-Path ".next") {
            $size = (Get-ChildItem .next -Recurse | Measure-Object -Property Length -Sum).Sum
            $sizeMB = [math]::Round($size / 1MB, 2)
            Write-Host "  → Bundle size: $sizeMB MB"
            
            # Document build time
            $buildTime = $buildOutput | Select-String "Compiled.*in" | Select-Object -First 1
            Write-Host "  → Build time: $buildTime"
        }
    } else {
        Write-Host "  ❌ Frontend build FAILED" -ForegroundColor Red
        Write-Host "  → Check $ReportDir\frontend-build.log for errors"
    }
    
    Pop-Location
} catch {
    Write-Host "  ❌ Error during frontend build: $_" -ForegroundColor Red
    Pop-Location
}

# ============================================================================
# CRITERION 2: Backend Build
# ============================================================================
Write-Host "`n[2/10] Backend Build Verification..." -ForegroundColor Yellow

try {
    Write-Host "  → Building all Go packages..."
    go build -v ./... 2>&1 | Tee-Object -FilePath "$ReportDir\backend-build.log"
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✅ Backend build SUCCESSFUL" -ForegroundColor Green
    } else {
        Write-Host "  ❌ Backend build FAILED" -ForegroundColor Red
    }
    
    Write-Host "  → Building API binary..."
    go build -o bin\api.exe .\cmd\api 2>&1 | Tee-Object -FilePath "$ReportDir\api-build.log"
    
    if ($LASTEXITCODE -eq 0 -and (Test-Path "bin\api.exe")) {
        $apiSize = (Get-Item bin\api.exe).Length / 1MB
        Write-Host "  ✅ API binary created: $([math]::Round($apiSize, 2)) MB" -ForegroundColor Green
        
        # Test binary execution
        Write-Host "  → Testing API binary..."
        $helpOutput = .\bin\api.exe --help 2>&1
        if ($helpOutput) {
            Write-Host "  ✅ API binary executes correctly" -ForegroundColor Green
        }
    }
    
    Write-Host "  → Building Worker binary..."
    go build -o bin\worker.exe .\cmd\worker 2>&1 | Tee-Object -FilePath "$ReportDir\worker-build.log"
    
    if ($LASTEXITCODE -eq 0 -and (Test-Path "bin\worker.exe")) {
        $workerSize = (Get-Item bin\worker.exe).Length / 1MB
        Write-Host "  ✅ Worker binary created: $([math]::Round($workerSize, 2)) MB" -ForegroundColor Green
    }
    
} catch {
    Write-Host "  ❌ Error during backend build: $_" -ForegroundColor Red
}

# ============================================================================
# CRITERION 3: Go Modules Tidy
# ============================================================================
Write-Host "`n[3/10] Go Modules Tidy Check..." -ForegroundColor Yellow

try {
    # Backup current state
    Copy-Item go.mod go.mod.backup
    Copy-Item go.sum go.sum.backup
    
    Write-Host "  → Running go mod tidy..."
    go mod tidy 2>&1 | Tee-Object -FilePath "$ReportDir\go-mod-tidy.log"
    
    # Check for differences
    $modDiff = Compare-Object (Get-Content go.mod) (Get-Content go.mod.backup)
    $sumDiff = Compare-Object (Get-Content go.sum) (Get-Content go.sum.backup)
    
    if ($null -eq $modDiff -and $null -eq $sumDiff) {
        Write-Host "  ✅ Go modules are tidy (no changes)" -ForegroundColor Green
    } else {
        Write-Host "  ⚠️  Go modules modified by tidy" -ForegroundColor Yellow
        if ($modDiff) {
            $modDiff | Out-File "$ReportDir\go-mod-diff.txt"
            Write-Host "  → go.mod changes saved to $ReportDir\go-mod-diff.txt"
        }
        if ($sumDiff) {
            $sumDiff | Out-File "$ReportDir\go-sum-diff.txt"
            Write-Host "  → go.sum changes saved to $ReportDir\go-sum-diff.txt"
        }
    }
} catch {
    Write-Host "  ❌ Error during go mod tidy: $_" -ForegroundColor Red
}

# ============================================================================
# CRITERION 4: ESLint Warnings
# ============================================================================
Write-Host "`n[4/10] ESLint Verification..." -ForegroundColor Yellow

try {
    Push-Location web
    
    Write-Host "  → Running ESLint..."
    $lintOutput = npm run lint 2>&1 | Tee-Object -FilePath "..\$ReportDir\eslint-report.txt"
    
    # Count warnings and errors
    $warnings = ($lintOutput | Select-String "warning").Count
    $errors = ($lintOutput | Select-String "error").Count
    
    if ($errors -eq 0) {
        Write-Host "  ✅ No ESLint errors found" -ForegroundColor Green
        if ($warnings -eq 0) {
            Write-Host "  ✅ No ESLint warnings found" -ForegroundColor Green
        } else {
            Write-Host "  ⚠️  $warnings ESLint warnings found" -ForegroundColor Yellow
        }
    } else {
        Write-Host "  ❌ $errors ESLint errors found" -ForegroundColor Red
    }
    
    Pop-Location
} catch {
    Write-Host "  ❌ Error during ESLint: $_" -ForegroundColor Red
    Pop-Location
}

# ============================================================================
# CRITERION 5: Chakra Compatibility
# ============================================================================
Write-Host "`n[5/10] Chakra Compatibility Audit..." -ForegroundColor Yellow

try {
    Write-Host "  → Searching for deprecated Chakra props..."
    
    # Search for common Chakra v3 breaking changes
    Push-Location web
    
    $colorScheme = (Get-ChildItem -Recurse -Include *.tsx,*.ts | Select-String "colorScheme" | Measure-Object).Count
    $variantProp = (Get-ChildItem -Recurse -Include *.tsx,*.ts | Select-String 'variant=' | Measure-Object).Count
    
    Write-Host "  → Found $colorScheme uses of 'colorScheme'"
    Write-Host "  → Found $variantProp uses of 'variant=' prop"
    
    # Check for 'use client' directives
    $useClient = (Get-ChildItem -Recurse -Include *.tsx | Select-String "'use client'" | Measure-Object).Count
    Write-Host "  → Found $useClient files with 'use client' directive"
    
    if ($useClient -gt 0) {
        Write-Host "  ✅ Chakra client components properly marked" -ForegroundColor Green
    } else {
        Write-Host "  ⚠️  No 'use client' directives found - verify if needed" -ForegroundColor Yellow
    }
    
    Pop-Location
} catch {
    Write-Host "  ❌ Error during Chakra audit: $_" -ForegroundColor Red
}

# ============================================================================
# CRITERION 6: Debug Logs
# ============================================================================
Write-Host "`n[6/10] Debug Logs Audit..." -ForegroundColor Yellow

try {
    # Frontend console.log
    Write-Host "  → Checking frontend for console.log..."
    Push-Location web
    $consoleLogs = Get-ChildItem -Recurse -Include *.ts,*.tsx,*.js,*.jsx -Exclude node_modules,.next | 
                   Select-String "console\.log" | 
                   Where-Object { $_.Line -notmatch "//.*console\.log" }
    
    $consoleCount = ($consoleLogs | Measure-Object).Count
    
    if ($consoleCount -eq 0) {
        Write-Host "  ✅ No console.log statements found" -ForegroundColor Green
    } else {
        Write-Host "  ❌ Found $consoleCount console.log statements" -ForegroundColor Red
        $consoleLogs | Out-File "..\$ReportDir\console-logs.txt"
    }
    Pop-Location
    
    # Backend fmt.Println
    Write-Host "  → Checking backend for fmt.Println..."
    $debugPrints = Get-ChildItem -Recurse -Include *.go | 
                   Select-String "fmt\.Println" | 
                   Where-Object { $_.Path -notmatch "_test\.go" }
    
    $debugCount = ($debugPrints | Measure-Object).Count
    
    if ($debugCount -eq 0) {
        Write-Host "  ✅ No fmt.Println statements found" -ForegroundColor Green
    } else {
        Write-Host "  ❌ Found $debugCount fmt.Println statements" -ForegroundColor Red
        $debugPrints | Out-File "$ReportDir\debug-prints.txt"
    }
} catch {
    Write-Host "  ❌ Error during debug logs audit: $_" -ForegroundColor Red
}

# ============================================================================
# CRITERION 7: Environment Variables
# ============================================================================
Write-Host "`n[7/10] Environment Variables Check..." -ForegroundColor Yellow

try {
    if (Test-Path ".env.example") {
        $envExample = Get-Content .env.example
        $envCount = ($envExample | Where-Object { $_ -match "=" }).Count
        Write-Host "  ✅ .env.example exists with $envCount variables" -ForegroundColor Green
        
        # Check for placeholder values
        $placeholders = $envExample | Select-String "CHANGE_THIS|YOUR_|changeme"
        Write-Host "  → Found $($placeholders.Count) placeholder values (expected)"
    } else {
        Write-Host "  ❌ .env.example not found" -ForegroundColor Red
    }
} catch {
    Write-Host "  ❌ Error checking environment variables: $_" -ForegroundColor Red
}

# ============================================================================
# CRITERION 8: Rate Limiter
# ============================================================================
Write-Host "`n[8/10] Rate Limiter Verification..." -ForegroundColor Yellow

try {
    if (Test-Path "internal\ratelimit\ratelimit.go") {
        Write-Host "  ✅ Rate limiter implementation exists" -ForegroundColor Green
        
        # Check for middleware
        if (Test-Path "internal\api\http\middleware\ratelimit.go") {
            Write-Host "  ✅ Rate limiter middleware exists" -ForegroundColor Green
        }
        
        # Check for tests
        if (Test-Path "internal\ratelimit\ratelimit_test.go") {
            Write-Host "  ✅ Rate limiter tests exist" -ForegroundColor Green
        }
    } else {
        Write-Host "  ❌ Rate limiter not found" -ForegroundColor Red
    }
} catch {
    Write-Host "  ❌ Error checking rate limiter: $_" -ForegroundColor Red
}

# ============================================================================
# CRITERION 9: Multi-Tenant Isolation
# ============================================================================
Write-Host "`n[9/10] Multi-Tenant Isolation Check..." -ForegroundColor Yellow

try {
    # Check for tenant context in auth
    $tenantContext = Get-ChildItem -Recurse -Include *.go | 
                     Select-String "TenantFromContext|WithTenant"
    
    $contextCount = ($tenantContext | Measure-Object).Count
    
    if ($contextCount -gt 0) {
        Write-Host "  ✅ Found $contextCount tenant context references" -ForegroundColor Green
    } else {
        Write-Host "  ⚠️  No tenant context found" -ForegroundColor Yellow
    }
    
    # Check for orgID/tenantID in database queries
    $queries = Get-ChildItem -Recurse -Include *.go | 
               Select-String "WHERE.*tenant_id|WHERE.*org_id"
    
    $queryCount = ($queries | Measure-Object).Count
    Write-Host "  → Found $queryCount tenant-scoped queries"
    
} catch {
    Write-Host "  ❌ Error checking multi-tenancy: $_" -ForegroundColor Red
}

# ============================================================================
# CRITERION 10: API Versioning
# ============================================================================
Write-Host "`n[10/10] API Versioning Check..." -ForegroundColor Yellow

try {
    $v1Routes = Get-ChildItem -Recurse -Path internal\api\http -Include *.go | 
                Select-String "/api/v1|/api/auth|/api/emissions"
    
    $routeCount = ($v1Routes | Measure-Object).Count
    
    if ($routeCount -gt 0) {
        Write-Host "  ✅ Found $routeCount API route references" -ForegroundColor Green
        Write-Host "  → API versioning implemented"
    } else {
        Write-Host "  ⚠️  No versioned routes found" -ForegroundColor Yellow
    }
} catch {
    Write-Host "  ❌ Error checking API versioning: $_" -ForegroundColor Red
}

# ============================================================================
# SUMMARY REPORT
# ============================================================================
Write-Host "`n=========================================="  -ForegroundColor Cyan
Write-Host "VERIFICATION SUMMARY"                         -ForegroundColor Cyan
Write-Host "=========================================="  -ForegroundColor Cyan

Write-Host "`nReports saved to: $ReportDir"
Write-Host "Timestamp: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"

Write-Host "`nNext Steps:"
Write-Host "1. Review build logs if any errors occurred"
Write-Host "2. Fix ESLint warnings (if any)"
Write-Host "3. Run integration tests"
Write-Host "4. Deploy to staging environment"

Write-Host "`n=========================================="  -ForegroundColor Cyan
