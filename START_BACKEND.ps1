# Start OffGridFlow Backend API with proper environment variables

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Starting OffGridFlow Backend" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
Write-Host "Checking for Go installation..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "✓ Found: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Go is not installed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install Go from: https://go.dev/dl/" -ForegroundColor Yellow
    Write-Host "Then run this script again." -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "Starting API server on port 8090..." -ForegroundColor Yellow
Write-Host "Press Ctrl+C to stop the server" -ForegroundColor Gray
Write-Host ""

Set-Location "C:\Users\pault\OffGridFlow"

# Set environment variables for development
$env:OFFGRIDFLOW_APP_ENV = "development"
$env:OFFGRIDFLOW_HTTP_PORT = "8090"
$env:OFFGRIDFLOW_DB_DSN = ""
$env:OFFGRIDFLOW_JWT_SECRET = "dev-jwt-secret-for-local-development-only-minimum-32-characters-required-12345"
$env:OFFGRIDFLOW_REQUIRE_AUTH = "false"
$env:OFFGRIDFLOW_ENABLE_AUDIT_LOG = "false"
$env:OFFGRIDFLOW_ENABLE_METRICS = "true"
$env:OFFGRIDFLOW_ENABLE_GRAPHQL = "true"
$env:OFFGRIDFLOW_ENABLE_OFFLINE_AI = "true"
$env:OFFGRIDFLOW_TRACING_ENABLED = "false"
$env:LOG_LEVEL = "info"

# Build and run
go run cmd/api/main.go
