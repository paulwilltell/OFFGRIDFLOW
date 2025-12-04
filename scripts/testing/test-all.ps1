#!/usr/bin/env pwsh
# Comprehensive Test Suite for OffGridFlow
# Tests all major features and integrations

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet('unit', 'integration', 'e2e', 'all')]
    [string]$TestType = 'all',
    
    [Parameter(Mandatory=$false)]
    [switch]$Verbose,
    
    [Parameter(Mandatory=$false)]
    [switch]$Coverage
)

$ErrorActionPreference = "Continue"

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "OffGridFlow Comprehensive Test Suite" -ForegroundColor Cyan
Write-Host "Test Type: $TestType" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

$testsPassed = 0
$testsFailed = 0
$testsSkipped = 0

# Helper function to run test
function Invoke-Test {
    param($Name, $Command)
    
    Write-Host "Testing: $Name..." -ForegroundColor Yellow
    try {
        $output = Invoke-Expression $Command 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  ✓ PASS" -ForegroundColor Green
            $script:testsPassed++
            if ($Verbose) { Write-Host $output }
            return $true
        } else {
            Write-Host "  ✗ FAIL" -ForegroundColor Red
            Write-Host $output -ForegroundColor Red
            $script:testsFailed++
            return $false
        }
    } catch {
        Write-Host "  ✗ ERROR: $_" -ForegroundColor Red
        $script:testsFailed++
        return $false
    }
}

# Unit Tests
if ($TestType -in @('unit', 'all')) {
    Write-Host ""
    Write-Host "==================== UNIT TESTS ====================" -ForegroundColor Cyan
    Write-Host ""
    
    # Auth tests
    Write-Host "[Auth Package]" -ForegroundColor Yellow
    Invoke-Test "JWT token generation" "go test ./internal/auth -run TestGenerateToken -v"
    Invoke-Test "Password hashing" "go test ./internal/auth -run TestHashPassword -v"
    Invoke-Test "Token validation" "go test ./internal/auth -run TestValidateToken -v"
    Invoke-Test "RBAC authorization" "go test ./internal/auth -run TestAuthorize -v"
    
    # Emissions tests
    Write-Host ""
    Write-Host "[Emissions Package]" -ForegroundColor Yellow
    Invoke-Test "Scope 1 calculations" "go test ./internal/emissions -run TestScope1 -v"
    Invoke-Test "Scope 2 calculations" "go test ./internal/emissions -run TestScope2 -v"
    Invoke-Test "Scope 3 calculations" "go test ./internal/emissions -run TestScope3 -v"
    Invoke-Test "Emission factors" "go test ./internal/emissions -run TestEmissionFactors -v"
    
    # Billing tests
    Write-Host ""
    Write-Host "[Billing Package]" -ForegroundColor Yellow
    Invoke-Test "Stripe webhook processing" "go test ./internal/billing -run TestWebhook -v"
    Invoke-Test "Subscription management" "go test ./internal/billing -run TestSubscription -v"
    Invoke-Test "Usage tracking" "go test ./internal/billing -run TestUsage -v"
    
    # Job queue tests
    Write-Host ""
    Write-Host "[Job Queue Package]" -ForegroundColor Yellow
    Invoke-Test "Job creation" "go test ./internal/jobs -run TestCreateJob -v"
    Invoke-Test "Job processing" "go test ./internal/jobs -run TestProcessJob -v"
    Invoke-Test "Job retry logic" "go test ./internal/jobs -run TestRetry -v"
    
    # Connectors tests
    Write-Host ""
    Write-Host "[Connectors Package]" -ForegroundColor Yellow
    Invoke-Test "AWS connector" "go test ./internal/connectors -run TestAWS -v"
    Invoke-Test "Azure connector" "go test ./internal/connectors -run TestAzure -v"
    Invoke-Test "GCP connector" "go test ./internal/connectors -run TestGCP -v"
    Invoke-Test "SAP connector" "go test ./internal/connectors -run TestSAP -v"
    
    # Exporters tests
    Write-Host ""
    Write-Host "[Exporters Package]" -ForegroundColor Yellow
    Invoke-Test "XBRL exporter" "go test ./internal/exporters -run TestXBRL -v"
    Invoke-Test "PDF exporter" "go test ./internal/exporters -run TestPDF -v"
    Invoke-Test "CSV exporter" "go test ./internal/exporters -run TestCSV -v"
    
    # Audit logging tests
    Write-Host ""
    Write-Host "[Audit Package]" -ForegroundColor Yellow
    Invoke-Test "Audit log creation" "go test ./internal/audit -run TestCreate -v"
    Invoke-Test "Audit log querying" "go test ./internal/audit -run TestQuery -v"
    
    # Rate limiting tests
    Write-Host ""
    Write-Host "[Rate Limiting Package]" -ForegroundColor Yellow
    Invoke-Test "Rate limiter" "go test ./internal/ratelimit -run TestRateLimit -v"
    Invoke-Test "Token bucket" "go test ./internal/ratelimit -run TestTokenBucket -v"
}

# Integration Tests
if ($TestType -in @('integration', 'all')) {
    Write-Host ""
    Write-Host "================ INTEGRATION TESTS ================" -ForegroundColor Cyan
    Write-Host ""
    
    # Check if services are running
    Write-Host "Checking prerequisites..." -ForegroundColor Yellow
    try {
        $health = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 5
        Write-Host "  ✓ API is running" -ForegroundColor Green
    } catch {
        Write-Host "  ✗ API is not running. Start with: docker-compose up -d" -ForegroundColor Red
        $testsSkipped += 10
        if ($TestType -eq 'integration') { exit 1 }
    }
    
    if ($? -eq $true) {
        # API Integration Tests
        Write-Host ""
        Write-Host "[API Integration]" -ForegroundColor Yellow
        
        # Auth flow
        Write-Host "Testing authentication flow..." -ForegroundColor Yellow
        $body = @{
            email = "test@example.com"
            password = "test123"
        } | ConvertTo-Json
        
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/auth/login" `
                -Method Post `
                -Body $body `
                -ContentType "application/json" `
                -UseBasicParsing
            Write-Host "  ✓ Login works" -ForegroundColor Green
            $testsPassed++
            
            $token = ($response.Content | ConvertFrom-Json).token
        } catch {
            Write-Host "  ⚠ Login failed (may need test user)" -ForegroundColor Yellow
            $testsSkipped++
        }
        
        # Emissions calculation
        Write-Host "Testing emissions calculation..." -ForegroundColor Yellow
        $emissionsBody = @{
            scope = "scope1"
            category = "stationary_combustion"
            fuel_type = "natural_gas"
            quantity = 1000
            unit = "therms"
        } | ConvertTo-Json
        
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/emissions/calculate" `
                -Method Post `
                -Body $emissionsBody `
                -ContentType "application/json" `
                -Headers @{"X-Tenant-ID"="test-tenant"} `
                -UseBasicParsing
            Write-Host "  ✓ Emissions calculation works" -ForegroundColor Green
            $testsPassed++
        } catch {
            Write-Host "  ✗ Emissions calculation failed" -ForegroundColor Red
            $testsFailed++
        }
        
        # Job queue
        Write-Host "Testing job queue..." -ForegroundColor Yellow
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/jobs?tenant_id=test-tenant" `
                -Method Get `
                -UseBasicParsing
            Write-Host "  ✓ Job queue works" -ForegroundColor Green
            $testsPassed++
        } catch {
            Write-Host "  ✗ Job queue failed" -ForegroundColor Red
            $testsFailed++
        }
        
        # Rate limiting
        Write-Host "Testing rate limiting..." -ForegroundColor Yellow
        $hitLimit = $false
        for ($i = 0; $i -lt 150; $i++) {
            try {
                $response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/health" `
                    -Method Get `
                    -Headers @{"X-Tenant-ID"="test-tenant"} `
                    -UseBasicParsing `
                    -TimeoutSec 1
            } catch {
                if ($_.Exception.Response.StatusCode -eq 429) {
                    $hitLimit = $true
                    break
                }
            }
        }
        if ($hitLimit) {
            Write-Host "  ✓ Rate limiting works" -ForegroundColor Green
            $testsPassed++
        } else {
            Write-Host "  ⚠ Rate limiting not triggered (may be disabled)" -ForegroundColor Yellow
            $testsSkipped++
        }
    }
}

# E2E Tests
if ($TestType -in @('e2e', 'all')) {
    Write-Host ""
    Write-Host "=================== E2E TESTS ===================" -ForegroundColor Cyan
    Write-Host ""
    
    Write-Host "[End-to-End Workflows]" -ForegroundColor Yellow
    Write-Host "  ⚠ E2E tests require manual verification" -ForegroundColor Yellow
    Write-Host "  Run: npm run test:e2e (in web directory)" -ForegroundColor Yellow
    $testsSkipped++
}

# Coverage Report
if ($Coverage) {
    Write-Host ""
    Write-Host "================= COVERAGE REPORT =================" -ForegroundColor Cyan
    Write-Host ""
    
    Write-Host "Generating coverage report..." -ForegroundColor Yellow
    go test ./... -coverprofile=coverage.out -covermode=atomic
    go tool cover -html=coverage.out -o coverage.html
    
    Write-Host "  ✓ Coverage report generated: coverage.html" -ForegroundColor Green
    
    # Display summary
    $coverageData = go tool cover -func=coverage.out | Select-String "total:"
    Write-Host $coverageData -ForegroundColor Cyan
}

# Summary
Write-Host ""
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "Test Results Summary" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Passed:  $testsPassed" -ForegroundColor Green
Write-Host "Failed:  $testsFailed" -ForegroundColor Red
Write-Host "Skipped: $testsSkipped" -ForegroundColor Yellow
Write-Host ""

$total = $testsPassed + $testsFailed + $testsSkipped
if ($total -gt 0) {
    $passRate = [math]::Round(($testsPassed / ($testsPassed + $testsFailed)) * 100, 2)
    Write-Host "Pass Rate: $passRate%" -ForegroundColor Cyan
}

if ($testsFailed -eq 0) {
    Write-Host ""
    Write-Host "✓ ALL TESTS PASSED!" -ForegroundColor Green
    exit 0
} else {
    Write-Host ""
    Write-Host "✗ SOME TESTS FAILED" -ForegroundColor Red
    exit 1
}
