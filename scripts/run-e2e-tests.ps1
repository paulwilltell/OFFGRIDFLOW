#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Execute E2E tests for OffGridFlow Web Application
    
.DESCRIPTION
    Runs Playwright E2E tests to validate authentication flows and integration workflows
    
.PARAMETER Browser
    Browser to run tests in (chromium, firefox, webkit, all)
    
.PARAMETER UI
    Run tests in UI mode for debugging
    
.PARAMETER Headed
    Run tests in headed mode (show browser)
    
.PARAMETER Debug
    Run tests in debug mode
    
.PARAMETER Report
    Show test report after completion
    
.PARAMETER Project
    Run specific project/browser only
    
.EXAMPLE
    .\run-e2e-tests.ps1
    
.EXAMPLE
    .\run-e2e-tests.ps1 -Browser chromium -Headed
    
.EXAMPLE
    .\run-e2e-tests.ps1 -UI
#>

param(
    [ValidateSet('chromium', 'firefox', 'webkit', 'all')]
    [string]$Browser = 'chromium',
    
    [switch]$UI,
    [switch]$Headed,
    [switch]$Debug,
    [switch]$Report,
    [string]$Project = ''
)

$ErrorActionPreference = 'Stop'

# Colors
$ColorSuccess = 'Green'
$ColorError = 'Red'
$ColorInfo = 'Cyan'
$ColorWarning = 'Yellow'

function Write-Step {
    param([string]$Message)
    Write-Host "===> $Message" -ForegroundColor $ColorInfo
}

function Write-Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor $ColorSuccess
}

function Write-Failure {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor $ColorError
}

function Write-Info {
    param([string]$Message)
    Write-Host "ℹ $Message" -ForegroundColor $ColorInfo
}

# Change to web directory
$webDir = Join-Path $PSScriptRoot ".." "web"
if (-not (Test-Path $webDir)) {
    Write-Failure "Web directory not found: $webDir"
    exit 1
}

Set-Location $webDir

Write-Step "OffGridFlow E2E Test Runner"
Write-Host ""

# Check if Playwright is installed
Write-Step "Checking Playwright installation..."
$playwrightInstalled = npm list @playwright/test 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Info "Installing Playwright..."
    npm install --save-dev @playwright/test
    
    Write-Info "Installing Playwright browsers..."
    npx playwright install --with-deps
} else {
    Write-Success "Playwright is installed"
}

# Check if Next.js dev server is running
Write-Step "Checking if Next.js dev server is running..."
$serverRunning = $false
try {
    $response = Invoke-WebRequest -Uri "http://localhost:3000" -TimeoutSec 2 -ErrorAction SilentlyContinue
    $serverRunning = $true
    Write-Success "Dev server is running"
} catch {
    Write-Warning "Dev server is not running"
    Write-Info "Starting Next.js dev server..."
    
    # Start dev server in background
    $devServerJob = Start-Job -ScriptBlock {
        Set-Location $using:webDir
        npm run dev
    }
    
    Write-Info "Waiting for dev server to start..."
    $maxWait = 60
    $waited = 0
    while ($waited -lt $maxWait) {
        Start-Sleep -Seconds 2
        $waited += 2
        
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:3000" -TimeoutSec 2 -ErrorAction SilentlyContinue
            Write-Success "Dev server is ready!"
            $serverRunning = $true
            break
        } catch {
            Write-Host "." -NoNewline
        }
    }
    
    Write-Host ""
    
    if (-not $serverRunning) {
        Write-Failure "Dev server failed to start within $maxWait seconds"
        Stop-Job $devServerJob
        Remove-Job $devServerJob
        exit 1
    }
}

# Build test command
Write-Step "Running E2E tests..."
$testCommand = "npx playwright test"

if ($UI) {
    $testCommand += " --ui"
    Write-Info "Running in UI mode"
} elseif ($Debug) {
    $testCommand += " --debug"
    Write-Info "Running in debug mode"
} else {
    # Add browser selection
    if ($Browser -ne 'all') {
        $testCommand += " --project=$Browser"
        Write-Info "Running tests in $Browser"
    } else {
        Write-Info "Running tests in all browsers"
    }
    
    if ($Headed) {
        $testCommand += " --headed"
        Write-Info "Running in headed mode"
    }
    
    if ($Project) {
        $testCommand += " --project=$Project"
    }
}

Write-Host ""
Write-Info "Executing: $testCommand"
Write-Host ""

# Run tests
$testStartTime = Get-Date
Invoke-Expression $testCommand
$testExitCode = $LASTEXITCODE
$testEndTime = Get-Date
$testDuration = $testEndTime - $testStartTime

Write-Host ""
Write-Host "========================================" -ForegroundColor $ColorInfo
Write-Host ""

if ($testExitCode -eq 0) {
    Write-Success "All E2E tests passed!"
    Write-Host "Duration: $($testDuration.ToString('mm\:ss'))" -ForegroundColor $ColorSuccess
} else {
    Write-Failure "Some E2E tests failed"
    Write-Host "Duration: $($testDuration.ToString('mm\:ss'))" -ForegroundColor $ColorError
}

Write-Host ""

# Show report option
if ($Report -or ($testExitCode -ne 0 -and -not $UI -and -not $Debug)) {
    Write-Info "Opening test report..."
    npx playwright show-report
}

# Cleanup: Stop dev server if we started it
if ($devServerJob) {
    Write-Info "Stopping dev server..."
    Stop-Job $devServerJob
    Remove-Job $devServerJob
}

Write-Host ""
Write-Info "Test results saved to: playwright-report/"
Write-Info "To view report: npm run test:e2e:report"
Write-Host ""

exit $testExitCode
