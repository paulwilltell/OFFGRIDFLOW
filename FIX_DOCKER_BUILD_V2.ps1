# Fix OffGridFlow Docker Build - Version 2
# This script forces generation of package-lock.json

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "OffGridFlow Docker Build Fix (v2)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Navigate to web directory
$webDir = "C:\Users\pault\OffGridFlow\web"
Set-Location $webDir
Write-Host "Working in: $webDir" -ForegroundColor Yellow
Write-Host ""

# Method 1: Try to generate lockfile only
Write-Host "Attempting to generate package-lock.json..." -ForegroundColor Yellow
npm install --package-lock-only

if (Test-Path "package-lock.json") {
    Write-Host "✓ SUCCESS! package-lock.json created" -ForegroundColor Green
    $fileSize = (Get-Item "package-lock.json").Length
    Write-Host "  File size: $([math]::Round($fileSize/1KB, 2)) KB" -ForegroundColor Gray
    Write-Host ""
    
    # Return to root
    Set-Location "C:\Users\pault\OffGridFlow"
    
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "Ready to build! Run:" -ForegroundColor Green
    Write-Host "  docker-compose up -d" -ForegroundColor White
    Write-Host "========================================" -ForegroundColor Cyan
    exit 0
}

Write-Host "⚠ Method 1 failed. Trying fresh install..." -ForegroundColor Yellow
Write-Host ""

# Method 2: Delete node_modules and reinstall
Write-Host "Removing old node_modules..." -ForegroundColor Yellow
if (Test-Path "node_modules") {
    Remove-Item -Recurse -Force "node_modules"
    Write-Host "✓ Removed node_modules" -ForegroundColor Gray
}

Write-Host ""
Write-Host "Installing fresh dependencies (this may take a few minutes)..." -ForegroundColor Yellow
npm install

if (Test-Path "package-lock.json") {
    Write-Host ""
    Write-Host "✓ SUCCESS! package-lock.json created" -ForegroundColor Green
    $fileSize = (Get-Item "package-lock.json").Length
    Write-Host "  File size: $([math]::Round($fileSize/1KB, 2)) KB" -ForegroundColor Gray
    Write-Host ""
    
    # Return to root
    Set-Location "C:\Users\pault\OffGridFlow"
    
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "Ready to build! Run:" -ForegroundColor Green
    Write-Host "  docker-compose up -d" -ForegroundColor White
    Write-Host "========================================" -ForegroundColor Cyan
} else {
    Write-Host ""
    Write-Host "ERROR: Still couldn't create package-lock.json" -ForegroundColor Red
    Write-Host "This is unusual. Please check your npm installation." -ForegroundColor Red
    exit 1
}
