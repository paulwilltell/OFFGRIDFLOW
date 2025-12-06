# ========================================
# AUTOMATED SCREENSHOT HELPER
# ========================================
# Makes screenshot capture easier

Write-Host "📸 Screenshot Capture Helper" -ForegroundColor Cyan
Write-Host ""

# Check if app is running
Write-Host "Checking if OffGridFlow is running..." -ForegroundColor Yellow
$response = $null
try {
    $response = Invoke-WebRequest -Uri "http://localhost:3000" -TimeoutSec 5 -UseBasicParsing -ErrorAction SilentlyContinue
} catch {
    # App not running
}

if ($null -eq $response) {
    Write-Host "❌ App not running. Starting docker-compose..." -ForegroundColor Red
    Write-Host ""
    
    docker-compose up -d
    
    Write-Host ""
    Write-Host "⏱️  Waiting 30 seconds for app to start..." -ForegroundColor Yellow
    Start-Sleep -Seconds 30
}

Write-Host "✅ App is running" -ForegroundColor Green
Write-Host ""

# Create screenshots directory
New-Item -ItemType Directory -Force -Path "docs\screenshots" | Out-Null

# Open browser
Write-Host "🌐 Opening browser to http://localhost:3000" -ForegroundColor Cyan
Start-Process "http://localhost:3000"
Write-Host ""

Write-Host "📋 SCREENSHOT CHECKLIST:" -ForegroundColor Yellow
Write-Host ""
Write-Host "Use Win+Shift+S to capture each screen" -ForegroundColor White
Write-Host "Save to: C:\Users\pault\OffGridFlow\docs\screenshots\" -ForegroundColor Gray
Write-Host ""

$screenshots = @(
    @{Name="01-login.png"; URL="http://localhost:3000/login"; Desc="Login page"},
    @{Name="02-dashboard.png"; URL="http://localhost:3000"; Desc="Main dashboard"},
    @{Name="03-activities.png"; URL="http://localhost:3000/activities"; Desc="Activities list"},
    @{Name="04-activity-form.png"; URL="http://localhost:3000/activities/new"; Desc="Create activity form"},
    @{Name="05-emissions-summary.png"; URL="http://localhost:3000/emissions"; Desc="Emissions summary"},
    @{Name="06-compliance-reports.png"; URL="http://localhost:3000/reports"; Desc="Compliance reports"},
    @{Name="07-csrd-report.png"; URL="http://localhost:3000/reports/csrd"; Desc="Sample report"},
    @{Name="08-settings.png"; URL="http://localhost:3000/settings"; Desc="Settings page"},
    @{Name="09-api-keys.png"; URL="http://localhost:3000/settings/api"; Desc="API keys"},
    @{Name="10-audit-log.png"; URL="http://localhost:3000/audit"; Desc="Audit log"}
)

$currentDir = Get-Location
foreach ($shot in $screenshots) {
    Write-Host "[ ] $($shot.Name) - $($shot.Desc)" -ForegroundColor White
}

Write-Host ""
Write-Host "Press any key to start screenshot capture process..." -ForegroundColor Yellow
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
Write-Host ""

$capturedCount = 0
foreach ($shot in $screenshots) {
    Write-Host ""
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host "📸 Screenshot $($capturedCount + 1)/10: $($shot.Desc)" -ForegroundColor Yellow
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Opening: $($shot.URL)" -ForegroundColor Gray
    
    # Open URL
    Start-Process $shot.URL
    Start-Sleep -Seconds 2
    
    Write-Host ""
    Write-Host "1. Use Win+Shift+S to capture" -ForegroundColor White
    Write-Host "2. Select area to capture" -ForegroundColor White
    Write-Host "3. Click notification to open Snip & Sketch" -ForegroundColor White
    Write-Host "4. Click 'Save As' (Ctrl+S)" -ForegroundColor White
    Write-Host "5. Save as: $($shot.Name)" -ForegroundColor Yellow
    Write-Host "   Location: $currentDir\docs\screenshots\" -ForegroundColor Gray
    Write-Host ""
    
    Write-Host "Press any key when done capturing this screenshot..." -ForegroundColor Green
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    
    # Check if file exists
    $filePath = "docs\screenshots\$($shot.Name)"
    if (Test-Path $filePath) {
        $capturedCount++
        Write-Host "   ✅ Captured: $($shot.Name)" -ForegroundColor Green
    } else {
        Write-Host "   ⚠️  File not found - you can capture it later" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "📊 SUMMARY" -ForegroundColor Yellow
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""
Write-Host "Screenshots captured: $capturedCount/10" -ForegroundColor $(if ($capturedCount -eq 10) { "Green" } else { "Yellow" })
Write-Host ""

if ($capturedCount -eq 10) {
    Write-Host "🎉 ALL SCREENSHOTS CAPTURED!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next step: Commit to GitHub" -ForegroundColor Cyan
    Write-Host "   Run: .\GIT_COMMIT_FIXED.ps1" -ForegroundColor White
    Write-Host ""
    Write-Host "📊 Section 5 Status: 100% COMPLETE!" -ForegroundColor Green
} elseif ($capturedCount -gt 0) {
    Write-Host "✅ Good progress! You can capture the rest later." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Missing:" -ForegroundColor Yellow
    foreach ($shot in $screenshots) {
        $filePath = "docs\screenshots\$($shot.Name)"
        if (-not (Test-Path $filePath)) {
            Write-Host "   - $($shot.Name)" -ForegroundColor Gray
        }
    }
} else {
    Write-Host "⚠️  No screenshots captured yet" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "You can:" -ForegroundColor White
    Write-Host "1. Re-run this script to try again" -ForegroundColor Gray
    Write-Host "2. Capture manually using the checklist above" -ForegroundColor Gray
}

Write-Host ""
Write-Host "Press any key to exit..." -ForegroundColor Gray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
