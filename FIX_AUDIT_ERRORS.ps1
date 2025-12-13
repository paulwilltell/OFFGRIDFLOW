# Fix Audit Package Duplicate Declarations

Write-Host "Fixing audit package compilation errors..." -ForegroundColor Yellow

Set-Location "C:\Users\pault\OffGridFlow"

# Rename the conflicting logger.go file
$source = "internal\audit\logger.go"
$backup = "internal\audit\logger.go.old"

if (Test-Path $source) {
    Write-Host "Backing up logger.go..." -ForegroundColor Gray
    Move-Item -Path $source -Destination $backup -Force
    Write-Host "✓ Moved logger.go to logger.go.old" -ForegroundColor Green
}

Write-Host ""
Write-Host "Attempting to build..." -ForegroundColor Yellow
go run cmd/api/main.go
