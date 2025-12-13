# Fix OffGridFlow Docker Build
# This script generates the missing package-lock.json file

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "OffGridFlow Docker Build Fix" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Navigate to web directory
Write-Host "Step 1: Navigating to web directory..." -ForegroundColor Yellow
$webDir = "C:\Users\pault\OffGridFlow\web"

if (-not (Test-Path $webDir)) {
    Write-Host "ERROR: Web directory not found at $webDir" -ForegroundColor Red
    exit 1
}

Set-Location $webDir
Write-Host "✓ In directory: $webDir" -ForegroundColor Green
Write-Host ""

# Step 2: Check if package.json exists
Write-Host "Step 2: Checking for package.json..." -ForegroundColor Yellow
if (-not (Test-Path "package.json")) {
    Write-Host "ERROR: package.json not found!" -ForegroundColor Red
    exit 1
}
Write-Host "✓ package.json found" -ForegroundColor Green
Write-Host ""

# Step 3: Generate package-lock.json
Write-Host "Step 3: Generating package-lock.json..." -ForegroundColor Yellow
Write-Host "This may take a few minutes as npm downloads dependencies..." -ForegroundColor Gray
Write-Host ""

npm install

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "ERROR: npm install failed!" -ForegroundColor Red
    Write-Host "Please check the error messages above." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "✓ package-lock.json generated successfully!" -ForegroundColor Green
Write-Host ""

# Step 4: Verify the file was created
Write-Host "Step 4: Verifying package-lock.json..." -ForegroundColor Yellow
if (Test-Path "package-lock.json") {
    $fileSize = (Get-Item "package-lock.json").Length
    Write-Host "✓ package-lock.json exists ($([math]::Round($fileSize/1KB, 2)) KB)" -ForegroundColor Green
} else {
    Write-Host "ERROR: package-lock.json was not created!" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Step 5: Go back to root and try Docker again
Write-Host "Step 5: Ready to build with Docker..." -ForegroundColor Yellow
Set-Location "C:\Users\pault\OffGridFlow"
Write-Host "✓ Back in root directory" -ForegroundColor Green
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "SUCCESS! Now run:" -ForegroundColor Green
Write-Host "  docker-compose up -d" -ForegroundColor White
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Or to run just the web app without Docker:" -ForegroundColor Gray
Write-Host "  cd web" -ForegroundColor White
Write-Host "  npm run dev" -ForegroundColor White
Write-Host ""
