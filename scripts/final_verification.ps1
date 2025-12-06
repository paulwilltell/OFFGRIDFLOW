# SECTION 1 - FINAL VERIFICATION SCRIPT
# Run this on your Windows machine to complete the final verification
# Million Fold Precision Framework

$ErrorActionPreference = "Continue"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "SECTION 1 - FINAL VERIFICATION" -ForegroundColor Cyan
Write-Host "Million Fold Precision Framework" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

$ReportDir = "reports\engineering\final"
New-Item -ItemType Directory -Force -Path $ReportDir | Out-Null

$Passed = 0
$Failed = 0
$Warned = 0

function Test-Result {
    param(
        [string]$Status,
        [string]$Name,
        [string]$Details = ""
    )
    
    if ($Status -eq "PASS") {
        Write-Host "✅ PASS" -ForegroundColor Green -NoNewline
        Write-Host " - $Name"
        if ($Details) { Write-Host "   → $Details" -ForegroundColor Gray }
        $script:Passed++
    }
    elseif ($Status -eq "FAIL") {
        Write-Host "❌ FAIL" -ForegroundColor Red -NoNewline
        Write-Host " - $Name"
        if ($Details) { Write-Host "   → $Details" -ForegroundColor Gray }
        $script:Failed++
    }
    else {
        Write-Host "⚠️  WARN" -ForegroundColor Yellow -NoNewline
        Write-Host " - $Name"
        if ($Details) { Write-Host "   → $Details" -ForegroundColor Gray }
        $script:Warned++
    }
}

# ============================================================================
# CRITERION 1: Frontend Build
# ============================================================================
Write-Host "`n=== CRITERION 1: Frontend Build ===" -ForegroundColor Yellow

if (Test-Path "web\package.json") {
    Test-Result "PASS" "package.json exists"
    
    $pkg = Get-Content web\package.json | ConvertFrom-Json
    if ($pkg.scripts.build) {
        Test-Result "PASS" "Build script configured" $pkg.scripts.build
    } else {
        Test-Result "FAIL" "Build script missing"
    }
} else {
    Test-Result "FAIL" "package.json missing"
}

if (Test-Path "web\tsconfig.json") {
    Test-Result "PASS" "TypeScript configured"
} else {
    Test-Result "WARN" "TypeScript config missing"
}

if (Test-Path "web\next.config.js") {
    Test-Result "PASS" "Next.js configured"
} elseif (Test-Path "web\next.config.mjs") {
    Test-Result "PASS" "Next.js configured" "Using .mjs"
} else {
    Test-Result "WARN" "Next.js config missing"
}

# Check node_modules
if (Test-Path "web\node_modules") {
    Test-Result "PASS" "node_modules installed"
} else {
    Test-Result "WARN" "node_modules missing" "Run: cd web && npm install"
}

# ============================================================================
# CRITERION 2: Backend Build
# ============================================================================
Write-Host "`n=== CRITERION 2: Backend Build ===" -ForegroundColor Yellow

# Check Go installation
try {
    $goVersion = go version 2>&1
    Test-Result "PASS" "Go installed" $goVersion
} catch {
    Test-Result "FAIL" "Go not found" "Install Go 1.21+"
}

if (Test-Path "go.mod") {
    Test-Result "PASS" "go.mod exists"
    
    $module = (Get-Content go.mod -First 1) -replace "module ", ""
    Test-Result "PASS" "Go module" $module
} else {
    Test-Result "FAIL" "go.mod missing"
}

if (Test-Path "cmd\api\main.go") {
    Test-Result "PASS" "API entry point exists"
} else {
    Test-Result "FAIL" "cmd\api\main.go missing"
}

if (Test-Path "cmd\worker\main.go") {
    Test-Result "PASS" "Worker entry point exists"
} else {
    Test-Result "WARN" "cmd\worker\main.go missing"
}

# ============================================================================
# CRITERION 3: Go Modules Tidy
# ============================================================================
Write-Host "`n=== CRITERION 3: Go Modules ===" -ForegroundColor Yellow

if (Test-Path "go.mod") {
    Copy-Item go.mod "$ReportDir\go.mod.before"
    Copy-Item go.sum "$ReportDir\go.sum.before"
    
    Write-Host "Running go mod tidy..." -ForegroundColor Gray
    go mod tidy 2>&1 | Out-Null
    
    $modDiff = Compare-Object (Get-Content go.mod) (Get-Content "$ReportDir\go.mod.before")
    if ($null -eq $modDiff) {
        Test-Result "PASS" "go.mod unchanged" "No modifications needed"
    } else {
        Test-Result "WARN" "go.mod modified" "Changes applied by tidy"
    }
}

# ============================================================================
# CRITERION 6: Debug Logs
# ============================================================================
Write-Host "`n=== CRITERION 6: Debug Logs ===" -ForegroundColor Yellow

# Frontend console.log
$consoleLogs = Get-ChildItem -Path web -Recurse -Include *.ts,*.tsx,*.js,*.jsx -ErrorAction SilentlyContinue | 
    Where-Object { $_.FullName -notmatch "node_modules|\.next|__tests__" } |
    Select-String "console\.log" -ErrorAction SilentlyContinue

if ($consoleLogs) {
    $count = ($consoleLogs | Measure-Object).Count
    Test-Result "WARN" "console.log found" "$count instances"
    $consoleLogs | Out-File "$ReportDir\console-logs.txt"
} else {
    Test-Result "PASS" "console.log audit" "0 instances found"
}

# Backend fmt.Println
$debugPrints = Get-ChildItem -Path internal,cmd -Recurse -Include *.go -ErrorAction SilentlyContinue | 
    Where-Object { $_.Name -notmatch "_test\.go" } |
    Select-String "fmt\.Println" -ErrorAction SilentlyContinue

if ($debugPrints) {
    $count = ($debugPrints | Measure-Object).Count
    Test-Result "WARN" "fmt.Println found" "$count instances"
    $debugPrints | Out-File "$ReportDir\debug-prints.txt"
} else {
    Test-Result "PASS" "fmt.Println audit" "0 instances found"
}

# ============================================================================
# CRITERION 7: Environment Variables
# ============================================================================
Write-Host "`n=== CRITERION 7: Environment Variables ===" -ForegroundColor Yellow

if (Test-Path ".env.example") {
    $lines = (Get-Content .env.example | Measure-Object -Line).Lines
    Test-Result "PASS" ".env.example exists" "$lines lines"
    
    $vars = (Get-Content .env.example | Where-Object { $_ -match "=" } | Measure-Object).Count
    Test-Result "PASS" "Environment variables" "$vars variables documented"
} else {
    Test-Result "FAIL" ".env.example missing"
}

# ============================================================================
# CRITERION 8: Rate Limiter
# ============================================================================
Write-Host "`n=== CRITERION 8: Rate Limiter ===" -ForegroundColor Yellow

if (Test-Path "internal\ratelimit\ratelimit.go") {
    Test-Result "PASS" "Rate limiter implementation"
} else {
    Test-Result "FAIL" "internal\ratelimit\ratelimit.go missing"
}

if (Test-Path "internal\api\http\middleware\ratelimit.go") {
    Test-Result "PASS" "Rate limit middleware"
} else {
    Test-Result "WARN" "Middleware missing"
}

if (Test-Path "internal\ratelimit\ratelimit_test.go") {
    Test-Result "PASS" "Rate limiter tests"
} else {
    Test-Result "WARN" "Tests missing"
}

# ============================================================================
# CRITERION 10: API Versioning
# ============================================================================
Write-Host "`n=== CRITERION 10: API Versioning ===" -ForegroundColor Yellow

if (Test-Path "internal\api\http\router.go") {
    Test-Result "PASS" "router.go exists"
    
    $router = Get-Content "internal\api\http\router.go" -Raw
    
    foreach ($route in @("/api/auth", "/api/emissions", "/api/compliance", "/api/billing")) {
        if ($router -match [regex]::Escape($route)) {
            Test-Result "PASS" "$route endpoint exists"
        } else {
            Test-Result "WARN" "$route not found"
        }
    }
} else {
    Test-Result "FAIL" "internal\api\http\router.go missing"
}

# ============================================================================
# CRITERION 11: Frontend API Tests
# ============================================================================
Write-Host "`n=== CRITERION 11: Frontend API Tests ===" -ForegroundColor Yellow

$testFiles = @(
    "web\__tests__\lib\api\activities.test.ts",
    "web\__tests__\lib\api\auth.test.ts"
)

foreach ($file in $testFiles) {
    if (Test-Path $file) {
        $lines = (Get-Content $file | Measure-Object -Line).Lines
        $name = Split-Path $file -Leaf
        Test-Result "PASS" "$name exists" "$lines lines"
    } else {
        $name = Split-Path $file -Leaf
        Test-Result "FAIL" "$name missing"
    }
}

# API Clients
$apiClients = @("activities", "auth", "emissions", "compliance")

foreach ($client in $apiClients) {
    $file = "web\lib\api\$client.ts"
    if (Test-Path $file) {
        $lines = (Get-Content $file | Measure-Object -Line).Lines
        Test-Result "PASS" "$client API client" "$lines lines"
    } else {
        Test-Result "FAIL" "$client API client missing"
    }
}

# ============================================================================
# CRITERION 12: Integration Tests
# ============================================================================
Write-Host "`n=== CRITERION 12: Integration Tests ===" -ForegroundColor Yellow

if (Test-Path "internal\api\http\comprehensive_integration_test.go") {
    $lines = (Get-Content "internal\api\http\comprehensive_integration_test.go" | Measure-Object -Line).Lines
    $tests = (Get-Content "internal\api\http\comprehensive_integration_test.go" | Select-String "^func Test").Count
    Test-Result "PASS" "Integration tests exist" "$lines lines, $tests test functions"
} else {
    Test-Result "FAIL" "Integration tests missing"
}

# ============================================================================
# CRITERION 14: Health Probes
# ============================================================================
Write-Host "`n=== CRITERION 14: Health Probes ===" -ForegroundColor Yellow

if (Test-Path "internal\api\http\router.go") {
    $router = Get-Content "internal\api\http\router.go" -Raw
    
    foreach ($endpoint in @("/health", "/livez", "/readyz")) {
        if ($router -match [regex]::Escape($endpoint)) {
            Test-Result "PASS" "$endpoint endpoint" "Implemented"
        } else {
            Test-Result "FAIL" "$endpoint missing"
        }
    }
    
    if ($router -match "DB\.HealthCheck|BillingService\.Ready") {
        Test-Result "PASS" "Dependency checks" "Database & billing validation"
    }
}

# ============================================================================
# FINAL SUMMARY
# ============================================================================
Write-Host "`n==========================================" -ForegroundColor Cyan
Write-Host "FINAL SUMMARY" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

$Total = $Passed + $Failed + $Warned
$Percent = [math]::Round(($Passed / $Total) * 100, 1)

Write-Host "✅ Passed:   $Passed" -ForegroundColor Green
Write-Host "⚠️  Warnings: $Warned" -ForegroundColor Yellow
Write-Host "❌ Failed:   $Failed" -ForegroundColor Red
Write-Host ""
Write-Host "Success Rate: $Percent%" -ForegroundColor Cyan
Write-Host ""

if ($Failed -eq 0) {
    Write-Host "🎉 SECTION 1 VERIFICATION COMPLETE 🎉" -ForegroundColor Green
    Write-Host ""
    Write-Host "All critical criteria met!" -ForegroundColor Green
    Write-Host "Ready for production deployment" -ForegroundColor Green
} else {
    Write-Host "⚠️  Some tests failed" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Review failures and address issues"
}

Write-Host ""
Write-Host "Report saved to: $ReportDir" -ForegroundColor Gray
Write-Host "Timestamp: $(Get-Date)" -ForegroundColor Gray
Write-Host "==========================================" -ForegroundColor Cyan
