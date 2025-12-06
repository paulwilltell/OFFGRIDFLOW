# Generate Example Reports Script
# Runs the Go script to create sample PDF reports

Write-Host "🎯 Generating Example Compliance Reports..." -ForegroundColor Green
Write-Host ""

$scriptPath = "C:\Users\pault\OffGridFlow"
Set-Location $scriptPath

# Check if Go is installed
$goVersion = & go version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Go not found. Please install Go 1.21+ first." -ForegroundColor Red
    exit 1
}

Write-Host "✅ Using: $goVersion" -ForegroundColor Green
Write-Host ""

# Run the generator
Write-Host "📄 Running report generator..." -ForegroundColor Cyan
& go run scripts/generate-example-reports.go

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "🎉 SUCCESS! Example reports generated." -ForegroundColor Green
    Write-Host ""
    Write-Host "📁 Check: examples\reports\" -ForegroundColor Yellow
    Write-Host ""
    
    # List generated files
    if (Test-Path "examples\reports") {
        $files = Get-ChildItem "examples\reports" -Filter "*.pdf"
        foreach ($file in $files) {
            $sizeKB = [math]::Round($file.Length / 1KB, 1)
            Write-Host "   ✅ $($file.Name) ($sizeKB KB)" -ForegroundColor White
        }
    }
} else {
    Write-Host ""
    Write-Host "❌ Report generation failed." -ForegroundColor Red
    Write-Host "Check error messages above." -ForegroundColor Yellow
}

Write-Host ""
