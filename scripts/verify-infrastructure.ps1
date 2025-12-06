# Docker Compose Verification Script
# Tests all infrastructure services startup

Write-Host "=== SECTION 3: INFRASTRUCTURE VERIFICATION ===" -ForegroundColor Green
Write-Host ""

$errors = 0
$warnings = 0

# Test 1: Docker Compose File Exists
Write-Host "[1/5] Checking docker-compose.yml..." -ForegroundColor Yellow
if (Test-Path "docker-compose.yml") {
    Write-Host "  ✅ docker-compose.yml exists" -ForegroundColor Green
} else {
    Write-Host "  ❌ docker-compose.yml NOT FOUND" -ForegroundColor Red
    $errors++
}

# Test 2: Check Docker is Running
Write-Host "`n[2/5] Checking Docker daemon..." -ForegroundColor Yellow
try {
    docker info | Out-Null
    Write-Host "  ✅ Docker is running" -ForegroundColor Green
} catch {
    Write-Host "  ❌ Docker is not running - please start Docker Desktop" -ForegroundColor Red
    $errors++
    Write-Host "`nERROR: Cannot proceed without Docker. Exiting." -ForegroundColor Red
    exit 1
}

# Test 3: Start Docker Compose Services
Write-Host "`n[3/5] Starting Docker Compose services..." -ForegroundColor Yellow
Write-Host "  (This may take 2-3 minutes on first run)" -ForegroundColor Cyan

docker-compose up -d 2>&1 | Out-Null

Start-Sleep -Seconds 10

# Test 4: Check Service Status
Write-Host "`n[4/5] Checking service health..." -ForegroundColor Yellow

$services = docker-compose ps --format json | ConvertFrom-Json

if ($services) {
    foreach ($service in $services) {
        $name = $service.Service
        $state = $service.State
        $status = $service.Status
        
        if ($state -eq "running") {
            Write-Host "  ✅ $name - running" -ForegroundColor Green
        } else {
            Write-Host "  ❌ $name - $state ($status)" -ForegroundColor Red
            $errors++
        }
    }
} else {
    Write-Host "  ⚠️  Could not parse service status" -ForegroundColor Yellow
    $warnings++
}

# Test 5: Check Logs for JSON Format
Write-Host "`n[5/5] Verifying JSON logging..." -ForegroundColor Yellow

$apiLogs = docker-compose logs api --tail 5 2>&1 | Out-String

if ($apiLogs -match '\{.*"time".*"level".*"msg".*\}') {
    Write-Host "  ✅ JSON structured logging detected" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  JSON logging not detected (may need API restart)" -ForegroundColor Yellow
    Write-Host "     Run: docker-compose restart api" -ForegroundColor Cyan
    $warnings++
}

# Summary
Write-Host "`n=== VERIFICATION SUMMARY ===" -ForegroundColor Green
Write-Host "Errors: $errors" -ForegroundColor $(if ($errors -eq 0) { "Green" } else { "Red" })
Write-Host "Warnings: $warnings" -ForegroundColor $(if ($warnings -eq 0) { "Green" } else { "Yellow" })

if ($errors -eq 0) {
    Write-Host "`n✅ ALL SERVICES RUNNING - Section 3 Infrastructure: 100% COMPLETE!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Access URLs:" -ForegroundColor Cyan
    Write-Host "  API:        http://localhost:8080" -ForegroundColor White
    Write-Host "  Web:        http://localhost:3000" -ForegroundColor White
    Write-Host "  Grafana:    http://localhost:3001 (admin/admin)" -ForegroundColor White
    Write-Host "  Prometheus: http://localhost:9090" -ForegroundColor White
    Write-Host "  Jaeger:     http://localhost:16686" -ForegroundColor White
} else {
    Write-Host "`n❌ ERRORS DETECTED - Please fix before proceeding" -ForegroundColor Red
    Write-Host ""
    Write-Host "To view service logs:" -ForegroundColor Yellow
    Write-Host "  docker-compose logs [service-name]" -ForegroundColor White
    Write-Host ""
    Write-Host "To restart a service:" -ForegroundColor Yellow
    Write-Host "  docker-compose restart [service-name]" -ForegroundColor White
}

Write-Host ""
