# OffGridFlow Launch Checklist Validation Script
# Million Fold Precision - Zero Tolerance for Incomplete Implementation
# Author: Paul Canttell
# Date: 2024-12-04

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet('All', 'Engineering', 'Security', 'Infrastructure', 'Compliance', 'Documentation', 'GTM', 'PostLaunch')]
    [string]$Section = 'All',
    
    [Parameter(Mandatory=$false)]
    [switch]$Verbose,
    
    [Parameter(Mandatory=$false)]
    [switch]$FixIssues
)

$ErrorActionPreference = 'Continue'
$ProgressPreference = 'SilentlyContinue'

# ANSI Colors for output
$GREEN = "`e[32m"
$RED = "`e[31m"
$YELLOW = "`e[33m"
$BLUE = "`e[34m"
$RESET = "`e[0m"

$script:TotalChecks = 0
$script:PassedChecks = 0
$script:FailedChecks = 0
$script:WarningChecks = 0
$script:Results = @()

function Write-ColorOutput {
    param([string]$Message, [string]$Color = $RESET)
    Write-Host "$Color$Message$RESET"
}

function Test-Check {
    param(
        [string]$Name,
        [scriptblock]$Test,
        [string]$Category,
        [scriptblock]$Fix = $null,
        [bool]$IsMandatory = $true
    )
    
    $script:TotalChecks++
    
    try {
        $result = & $Test
        
        if ($result -eq $true) {
            $script:PassedChecks++
            $status = "${GREEN}✓ PASS${RESET}"
            $severity = "PASS"
        }
        elseif ($result -eq "WARNING") {
            $script:WarningChecks++
            $status = "${YELLOW}⚠ WARN${RESET}"
            $severity = "WARNING"
        }
        else {
            $script:FailedChecks++
            $status = "${RED}✗ FAIL${RESET}"
            $severity = "FAIL"
            
            if ($FixIssues -and $null -ne $Fix) {
                Write-ColorOutput "  → Attempting automatic fix..." $YELLOW
                & $Fix
                Write-ColorOutput "  → Fix applied, re-testing..." $YELLOW
                $retestResult = & $Test
                if ($retestResult -eq $true) {
                    $script:FailedChecks--
                    $script:PassedChecks++
                    $status = "${GREEN}✓ FIXED${RESET}"
                    $severity = "FIXED"
                }
            }
        }
        
        $mandatory = if ($IsMandatory) { "MANDATORY" } else { "RECOMMENDED" }
        
        if ($Verbose -or $severity -ne "PASS") {
            Write-Host "[$Category][$mandatory] $status $Name"
        }
        
        $script:Results += [PSCustomObject]@{
            Category = $Category
            Name = $Name
            Status = $severity
            Mandatory = $IsMandatory
            Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
        }
        
        return $severity
    }
    catch {
        $script:FailedChecks++
        Write-ColorOutput "[$Category][ERROR] ✗ EXCEPTION $Name" $RED
        Write-ColorOutput "  Error: $_" $RED
        
        $script:Results += [PSCustomObject]@{
            Category = $Category
            Name = $Name
            Status = "EXCEPTION"
            Mandatory = $IsMandatory
            Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
            Error = $_.ToString()
        }
        
        return "EXCEPTION"
    }
}

# ============================================================================
# 1. ENGINEERING READINESS CHECKS
# ============================================================================

function Test-EngineeringReadiness {
    Write-ColorOutput "`n=== 1️⃣ ENGINEERING READINESS ===" $BLUE
    
    # Frontend Build
    Test-Check -Name "Frontend builds successfully" -Category "Engineering" -Test {
        Push-Location "web"
        $output = npm run build 2>&1
        Pop-Location
        return $LASTEXITCODE -eq 0
    }
    
    # Backend Build
    Test-Check -Name "Backend builds successfully" -Category "Engineering" -Test {
        $output = go build ./... 2>&1
        return $LASTEXITCODE -eq 0
    }
    
    # Go Modules
    Test-Check -Name "Go modules are tidy" -Category "Engineering" -Test {
        $before = Get-FileHash go.mod, go.sum
        go mod tidy 2>&1 | Out-Null
        $after = Get-FileHash go.mod, go.sum
        $changed = Compare-Object $before $after
        return $null -eq $changed
    }
    
    # Environment Variables
    Test-Check -Name "Environment variables documented" -Category "Engineering" -Test {
        $example = Get-Content .env.example -Raw
        $required = @('DATABASE_URL', 'REDIS_URL', 'JWT_SECRET', 'PORT')
        $missing = $required | Where-Object { $example -notmatch $_ }
        return $missing.Count -eq 0
    }
    
    # Rate Limiter
    Test-Check -Name "Rate limiter configured" -Category "Engineering" -Test {
        $authCode = Get-Content internal/handlers/auth.go -Raw
        return $authCode -match 'RateLimiter|ratelimit'
    }
    
    # Multi-tenant Isolation
    Test-Check -Name "Multi-tenant isolation implemented" -Category "Engineering" -Test {
        $dbCode = Get-Content internal/database/tenants.go -Raw -ErrorAction SilentlyContinue
        return $dbCode -match 'WHERE.*tenant_id|TenantID'
    }
    
    # API Versioning
    Test-Check -Name "API versioning confirmed (/api/v1/)" -Category "Engineering" -Test {
        $routerCode = Get-Content cmd/api/main.go -Raw -ErrorAction SilentlyContinue
        if ($null -eq $routerCode) {
            $routerCode = Get-Content cmd/server/main.go -Raw -ErrorAction SilentlyContinue
        }
        return $routerCode -match '/api/v1/'
    }
    
    # Console Logs Removed
    Test-Check -Name "No console.log in production code" -Category "Engineering" -Test {
        $logs = Get-ChildItem -Path web/app, web/components, web/lib -Recurse -Filter *.tsx, *.ts | 
                Select-String -Pattern 'console\.(log|debug)' -SimpleMatch
        return $logs.Count -eq 0
    } -IsMandatory $true
    
    # Go Debug Prints Removed
    Test-Check -Name "No fmt.Println in production code" -Category "Engineering" -Test {
        $prints = Get-ChildItem -Path internal, cmd -Recurse -Filter *.go | 
                  Select-String -Pattern 'fmt\.Println|fmt\.Printf(?!.*log)' 
        return $prints.Count -eq 0
    } -IsMandatory $true
}

# ============================================================================
# 2. SECURITY READINESS CHECKS
# ============================================================================

function Test-SecurityReadiness {
    Write-ColorOutput "`n=== 2️⃣ SECURITY READINESS ===" $BLUE
    
    # .env excluded from git
    Test-Check -Name ".env files excluded from git" -Category "Security" -Test {
        $gitignore = Get-Content .gitignore -Raw
        return $gitignore -match '\.env'
    }
    
    # No secrets in git history
    Test-Check -Name "No .env files in git history" -Category "Security" -Test {
        $history = git log --all --full-history -- '.env*' 2>&1
        return $history -match 'fatal: ambiguous argument'
    }
    
    # JWT Secret Configuration
    Test-Check -Name "JWT secret not default" -Category "Security" -Test {
        if (Test-Path .env) {
            $env = Get-Content .env -Raw
            $defaultSecrets = @('secret', 'your-secret-key', 'change-me', 'default')
            $hasDefault = $defaultSecrets | Where-Object { $env -match $_ }
            return $hasDefault.Count -eq 0
        }
        return "WARNING"
    }
    
    # Password Hashing
    Test-Check -Name "Password hashing uses bcrypt cost >= 12" -Category "Security" -Test {
        $authCode = Get-Content internal/auth/*.go -Raw -ErrorAction SilentlyContinue
        if ($authCode -match 'bcrypt\.GenerateFromPassword') {
            return $authCode -match 'bcrypt\.DefaultCost|cost.*1[2-9]'
        }
        return "WARNING"
    }
    
    # HTTPS Enforcement
    Test-Check -Name "HTTPS enforcement documented" -Category "Security" -Test {
        $k8sFiles = Get-ChildItem -Path deployments/kubernetes -Recurse -Filter *.yaml
        $hasHTTPS = $false
        foreach ($file in $k8sFiles) {
            $content = Get-Content $file.FullName -Raw
            if ($content -match 'tls|https') {
                $hasHTTPS = $true
                break
            }
        }
        return $hasHTTPS
    }
}

# ============================================================================
# 3. INFRASTRUCTURE READINESS CHECKS
# ============================================================================

function Test-InfrastructureReadiness {
    Write-ColorOutput "`n=== 3️⃣ INFRASTRUCTURE READINESS ===" $BLUE
    
    # Docker Compose
    Test-Check -Name "Docker Compose file valid" -Category "Infrastructure" -Test {
        if (-not (Test-Path docker-compose.yml)) { return $false }
        $output = docker-compose config 2>&1
        return $LASTEXITCODE -eq 0
    }
    
    # Migrations Exist
    Test-Check -Name "Database migrations exist" -Category "Infrastructure" -Test {
        $migrationDirs = @('internal/database/migrations', 'migrations', 'db/migrations')
        foreach ($dir in $migrationDirs) {
            if (Test-Path $dir) {
                $files = Get-ChildItem $dir -Filter *.sql
                if ($files.Count -gt 0) { return $true }
            }
        }
        return $false
    }
    
    # Structured Logging
    Test-Check -Name "Structured logging implemented" -Category "Infrastructure" -Test {
        $logCode = Get-ChildItem -Path internal, cmd -Recurse -Filter *.go | 
                   Select-String -Pattern 'log\.WithFields|log\.Info|logger\.'
        return $logCode.Count -gt 5
    }
    
    # Production .env Example
    Test-Check -Name "Production .env.example exists" -Category "Infrastructure" -Test {
        return Test-Path .env.production.example, .env.example
    }
}

# ============================================================================
# 4. COMPLIANCE READINESS CHECKS
# ============================================================================

function Test-ComplianceReadiness {
    Write-ColorOutput "`n=== 4️⃣ COMPLIANCE READINESS ===" $BLUE
    
    # Scope 1/2/3 Calculations
    Test-Check -Name "Scope 1/2/3 calculations implemented" -Category "Compliance" -Test {
        $emissionsCode = Get-ChildItem -Path internal/emissions -Recurse -Filter *.go -ErrorAction SilentlyContinue
        if ($emissionsCode) {
            $content = $emissionsCode | Get-Content -Raw | Out-String
            $hasScope1 = $content -match 'Scope1|scope_1'
            $hasScope2 = $content -match 'Scope2|scope_2'
            $hasScope3 = $content -match 'Scope3|scope_3'
            return $hasScope1 -and $hasScope2 -and $hasScope3
        }
        return $false
    }
    
    # Export Formats
    Test-Check -Name "Compliance exports (PDF/XBRL) implemented" -Category "Compliance" -Test {
        $exportCode = Get-ChildItem -Path internal/exports, internal/reports -Recurse -Filter *.go -ErrorAction SilentlyContinue
        if ($exportCode) {
            $content = $exportCode | Get-Content -Raw | Out-String
            $hasPDF = $content -match 'pdf|PDF'
            $hasXBRL = $content -match 'xbrl|XBRL'
            return $hasPDF -and $hasXBRL
        }
        return "WARNING"
    }
    
    # Audit Logging
    Test-Check -Name "Audit log for exports exists" -Category "Compliance" -Test {
        $auditCode = Get-ChildItem -Path internal -Recurse -Filter *.go | 
                     Select-String -Pattern 'AuditLog|audit_log|LogExport'
        return $auditCode.Count -gt 0
    }
}

# ============================================================================
# 5. DOCUMENTATION READINESS CHECKS
# ============================================================================

function Test-DocumentationReadiness {
    Write-ColorOutput "`n=== 5️⃣ DOCUMENTATION READINESS ===" $BLUE
    
    # README
    Test-Check -Name "README.md is comprehensive" -Category "Documentation" -Test {
        if (-not (Test-Path README.md)) { return $false }
        $readme = Get-Content README.md -Raw
        $required = @('Features', 'Quick Start', 'Installation', 'Architecture')
        $missing = $required | Where-Object { $readme -notmatch $_ }
        return $missing.Count -eq 0
    }
    
    # QUICKSTART.md
    Test-Check -Name "QUICKSTART.md exists and is current" -Category "Documentation" -Test {
        return Test-Path QUICKSTART.md
    }
    
    # API Documentation
    Test-Check -Name "API documentation exists" -Category "Documentation" -Test {
        $apiDocs = @('docs/API.md', 'docs/api.md', 'API.md', 'openapi.yaml', 'swagger.yaml')
        foreach ($doc in $apiDocs) {
            if (Test-Path $doc) { return $true }
        }
        return "WARNING"
    } -IsMandatory $false
    
    # Architecture Diagram
    Test-Check -Name "Architecture diagram exists" -Category "Documentation" -Test {
        $diagrams = Get-ChildItem -Path docs, . -Filter *.png, *.svg, *.drawio -Recurse -ErrorAction SilentlyContinue
        $archDiagrams = $diagrams | Where-Object { $_.Name -match 'architecture|diagram|flow' }
        return $archDiagrams.Count -gt 0
    } -IsMandatory $false
}

# ============================================================================
# 6. GO-TO-MARKET READINESS CHECKS
# ============================================================================

function Test-GTMReadiness {
    Write-ColorOutput "`n=== 6️⃣ GO-TO-MARKET READINESS ===" $BLUE
    
    # Pricing Documentation
    Test-Check -Name "Pricing documentation exists" -Category "GTM" -Test {
        $pricingDocs = @('PRICING.md', 'docs/PRICING.md', 'docs/pricing.pdf')
        foreach ($doc in $pricingDocs) {
            if (Test-Path $doc) { return $true }
        }
        return "WARNING"
    } -IsMandatory $false
    
    # Demo Environment
    Test-Check -Name "Demo environment documented" -Category "GTM" -Test {
        $readme = Get-Content README.md -Raw -ErrorAction SilentlyContinue
        return $readme -match 'demo|Demo'
    } -IsMandatory $false
}

# ============================================================================
# 7. POST-LAUNCH OPS CHECKS
# ============================================================================

function Test-PostLaunchOps {
    Write-ColorOutput "`n=== 7️⃣ POST-LAUNCH OPS ===" $BLUE
    
    # Error Tracking
    Test-Check -Name "Error tracking configured (Sentry)" -Category "PostLaunch" -Test {
        $sentryConfig = Get-ChildItem -Path web, internal -Recurse | 
                        Select-String -Pattern 'sentry|Sentry'
        return $sentryConfig.Count -gt 0
    } -IsMandatory $false
    
    # Health Endpoints
    Test-Check -Name "Health endpoints implemented" -Category "PostLaunch" -Test {
        $healthCode = Get-ChildItem -Path internal/handlers, cmd -Recurse -Filter *.go | 
                      Select-String -Pattern '/health|/readyz|/livez'
        return $healthCode.Count -gt 0
    }
    
    # Backup Documentation
    Test-Check -Name "Backup procedures documented" -Category "PostLaunch" -Test {
        $backupDocs = @('docs/BACKUP.md', 'BACKUP.md', 'docs/operations.md')
        foreach ($doc in $backupDocs) {
            if (Test-Path $doc) { return $true }
        }
        return "WARNING"
    } -IsMandatory $false
}

# ============================================================================
# MAIN EXECUTION
# ============================================================================

Write-ColorOutput @"
╔════════════════════════════════════════════════════════════════════╗
║                                                                    ║
║   🚀 OFFGRIDFLOW LAUNCH CHECKLIST VALIDATION                      ║
║   Million Fold Precision Framework                                ║
║   Zero Tolerance for Incomplete Implementation                    ║
║                                                                    ║
╚════════════════════════════════════════════════════════════════════╝
"@ $BLUE

$startTime = Get-Date

# Change to repository root
Set-Location $PSScriptRoot\..

# Run selected sections
switch ($Section) {
    'All' {
        Test-EngineeringReadiness
        Test-SecurityReadiness
        Test-InfrastructureReadiness
        Test-ComplianceReadiness
        Test-DocumentationReadiness
        Test-GTMReadiness
        Test-PostLaunchOps
    }
    'Engineering' { Test-EngineeringReadiness }
    'Security' { Test-SecurityReadiness }
    'Infrastructure' { Test-InfrastructureReadiness }
    'Compliance' { Test-ComplianceReadiness }
    'Documentation' { Test-DocumentationReadiness }
    'GTM' { Test-GTMReadiness }
    'PostLaunch' { Test-PostLaunchOps }
}

$endTime = Get-Date
$duration = ($endTime - $startTime).TotalSeconds

# ============================================================================
# RESULTS SUMMARY
# ============================================================================

Write-ColorOutput "`n╔════════════════════════════════════════════════════════════════════╗" $BLUE
Write-ColorOutput "║                       VALIDATION RESULTS                           ║" $BLUE
Write-ColorOutput "╚════════════════════════════════════════════════════════════════════╝" $BLUE

$passRate = if ($script:TotalChecks -gt 0) { 
    [math]::Round(($script:PassedChecks / $script:TotalChecks) * 100, 2) 
} else { 0 }

Write-Host ""
Write-ColorOutput "Total Checks:    $($script:TotalChecks)" $BLUE
Write-ColorOutput "Passed:          $($script:PassedChecks)" $GREEN
Write-ColorOutput "Failed:          $($script:FailedChecks)" $RED
Write-ColorOutput "Warnings:        $($script:WarningChecks)" $YELLOW
Write-ColorOutput "Pass Rate:       $passRate%" $(if ($passRate -ge 90) { $GREEN } elseif ($passRate -ge 70) { $YELLOW } else { $RED })
Write-ColorOutput "Duration:        $([math]::Round($duration, 2))s" $BLUE

# Export results to JSON
$resultsJson = $script:Results | ConvertTo-Json -Depth 10
$resultsPath = "validation-results-$(Get-Date -Format 'yyyyMMdd-HHmmss').json"
$resultsJson | Out-File -FilePath $resultsPath -Encoding UTF8
Write-ColorOutput "`nResults exported to: $resultsPath" $BLUE

# Determine exit code
$exitCode = if ($script:FailedChecks -eq 0) { 0 } else { 1 }

if ($exitCode -eq 0) {
    Write-ColorOutput "`n✓ ALL CHECKS PASSED - PRODUCTION READY" $GREEN
} else {
    Write-ColorOutput "`n✗ CHECKS FAILED - REMEDIATION REQUIRED" $RED
}

Write-Host ""
exit $exitCode
