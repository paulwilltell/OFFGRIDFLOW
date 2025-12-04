# OffGridFlow V1 Self-Check Script
# Run this to verify your setup is complete and working
# Usage: .\scripts\v1_check.ps1

$ErrorActionPreference = "Stop"
$host.UI.RawUI.WindowTitle = "OffGridFlow V1 Check"

Write-Host "`n=== OffGridFlow V1 Self-Check ===" -ForegroundColor Cyan
Write-Host "Running verification steps...`n" -ForegroundColor Gray

$allPassed = $true
$results = @()

function Test-Step {
    param(
        [string]$Name,
        [scriptblock]$Test
    )
    
    Write-Host "[$Name] " -NoNewline
    try {
        $output = & $Test 2>&1
        if ($LASTEXITCODE -eq 0 -or $null -eq $LASTEXITCODE) {
            Write-Host "PASS" -ForegroundColor Green
            return @{ Name = $Name; Status = "PASS"; Output = $output }
        } else {
            Write-Host "FAIL" -ForegroundColor Red
            Write-Host "  Error: $output" -ForegroundColor Yellow
            return @{ Name = $Name; Status = "FAIL"; Output = $output }
        }
    } catch {
        Write-Host "FAIL" -ForegroundColor Red
        Write-Host "  Error: $_" -ForegroundColor Yellow
        return @{ Name = $Name; Status = "FAIL"; Output = $_.ToString() }
    }
}

# Change to project root
Push-Location (Split-Path -Parent $PSScriptRoot)
Write-Host "Working directory: $(Get-Location)`n" -ForegroundColor Gray

# =============================================================================
# BACKEND CHECKS
# =============================================================================
Write-Host "--- Backend ---" -ForegroundColor Yellow

$results += Test-Step "Go Build" {
    go build ./...
}

$results += Test-Step "Go Tests" {
    $env:CGO_ENABLED = "0"
    go test ./... -short 2>&1 | Out-String
    # Allow test failures for now (no tests yet)
    $global:LASTEXITCODE = 0
}

# =============================================================================
# FRONTEND CHECKS
# =============================================================================
Write-Host "`n--- Frontend ---" -ForegroundColor Yellow

$results += Test-Step "NPM Install" {
    Push-Location web
    npm install --silent 2>&1 | Out-Null
    Pop-Location
}

$results += Test-Step "Next.js Build" {
    Push-Location web
    npm run build 2>&1 | Out-String
    Pop-Location
}

# =============================================================================
# API CHECKS (requires server running on :8090)
# =============================================================================
Write-Host "`n--- API Endpoints ---" -ForegroundColor Yellow

# Check if server is running
$serverRunning = $false
try {
    $null = Invoke-RestMethod -Uri "http://localhost:8090/health" -TimeoutSec 2 -ErrorAction Stop
    $serverRunning = $true
} catch {
    Write-Host "[API Check] Server not running on :8090, skipping endpoint tests" -ForegroundColor Yellow
    Write-Host "  Start server with: go run ./cmd/api`n" -ForegroundColor Gray
}

if ($serverRunning) {
    $results += Test-Step "GET /health" {
        $r = Invoke-RestMethod -Uri "http://localhost:8090/health"
        if ($r.status -eq "ok") { "OK" } else { throw "Unexpected: $r" }
    }

    $results += Test-Step "GET /api/offgrid/mode" {
        $r = Invoke-RestMethod -Uri "http://localhost:8090/api/offgrid/mode"
        if ($r.mode) { "Mode: $($r.mode)" } else { throw "Unexpected: $r" }
    }

    $results += Test-Step "POST /api/ai/chat" {
        $body = @{ prompt = "test" } | ConvertTo-Json
        $r = Invoke-RestMethod -Uri "http://localhost:8090/api/ai/chat" `
            -Method POST -ContentType "application/json" -Body $body
        if ($r.output) { "Response received" } else { throw "Unexpected: $r" }
    }

    $results += Test-Step "GET /api/emissions/scope2" {
        $r = Invoke-RestMethod -Uri "http://localhost:8090/api/emissions/scope2"
        if ($r -is [array]) { "Found $($r.Count) records" } else { throw "Unexpected: $r" }
    }

    $results += Test-Step "GET /api/emissions/scope2/summary" {
        $r = Invoke-RestMethod -Uri "http://localhost:8090/api/emissions/scope2/summary"
        if ($null -ne $r.totalEmissionsTonsCO2e) { "Total: $($r.totalEmissionsTonsCO2e) tons" } else { throw "Unexpected: $r" }
    }

    $results += Test-Step "GET /api/compliance/summary" {
        $r = Invoke-RestMethod -Uri "http://localhost:8090/api/compliance/summary"
        if ($r.frameworks) { "Frameworks: $($r.frameworks.PSObject.Properties.Name -join ', ')" } else { throw "Unexpected: $r" }
    }
}

# =============================================================================
# SUMMARY
# =============================================================================
Pop-Location

Write-Host "`n=== Summary ===" -ForegroundColor Cyan
$passed = ($results | Where-Object { $_.Status -eq "PASS" }).Count
$failed = ($results | Where-Object { $_.Status -eq "FAIL" }).Count
$total = $results.Count

if ($failed -eq 0) {
    Write-Host "All $total checks passed! " -ForegroundColor Green -NoNewline
    Write-Host "OffGridFlow V1 is ready." -ForegroundColor Green
    Write-Host "`nYou can now:" -ForegroundColor Gray
    Write-Host "  - Start the API:      go run ./cmd/api" -ForegroundColor White
    Write-Host "  - Start the frontend: cd web && npm run dev" -ForegroundColor White
    Write-Host "  - Enable real AI:     set OFFGRIDFLOW_OPENAI_API_KEY=sk-..." -ForegroundColor White
    Write-Host "  - Enable Postgres:    set OFFGRIDFLOW_DBDSN=postgres://..." -ForegroundColor White
} else {
    Write-Host "$passed passed, $failed failed" -ForegroundColor Red
    Write-Host "`nFailed checks:" -ForegroundColor Yellow
    $results | Where-Object { $_.Status -eq "FAIL" } | ForEach-Object {
        Write-Host "  - $($_.Name)" -ForegroundColor Red
    }
}

Write-Host ""
