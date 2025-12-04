# OffGridFlow - One-Command Local Development Setup (Windows)
# This script sets up and runs the entire OffGridFlow stack locally

$ErrorActionPreference = "Stop"

# Colors for output
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Function to check if a command exists
function Test-Command {
    param([string]$Command)
    $null -ne (Get-Command $Command -ErrorAction SilentlyContinue)
}

# Print banner
Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "  OffGridFlow - Local Development Setup  " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Check prerequisites
Write-Info "Checking prerequisites..."

if (-not (Test-Command docker)) {
    Write-Error "Docker is not installed. Please install Docker Desktop."
    exit 1
}

if (-not (Test-Command docker-compose)) {
    Write-Error "Docker Compose is not installed. Please install Docker Compose."
    exit 1
}

Write-Success "All prerequisites satisfied"

# Check if .env exists
if (-not (Test-Path .env)) {
    Write-Warning ".env file not found. Creating from template..."
    if (Test-Path .env.production.template) {
        Copy-Item .env.production.template .env
        Write-Info "Created .env file. Please update it with your credentials."
        Write-Warning "Using default development credentials for now..."
    } else {
        Write-Error ".env.production.template not found!"
        exit 1
    }
}

# Stop any existing containers
Write-Info "Stopping any existing containers..."
docker-compose down 2>$null

# Clean up old volumes (optional)
if ($args -contains "--clean") {
    Write-Warning "Cleaning up old volumes..."
    docker-compose down -v
}

# Pull latest images
Write-Info "Pulling latest images..."
try {
    docker-compose pull
} catch {
    Write-Warning "Could not pull images, will build locally"
}

# Build images
Write-Info "Building Docker images..."
docker-compose build

# Start services
Write-Info "Starting services..."
docker-compose up -d postgres redis jaeger otel-collector prometheus

# Wait for database to be ready
Write-Info "Waiting for PostgreSQL to be ready..."
$retries = 0
$maxRetries = 30
while ($retries -lt $maxRetries) {
    try {
        docker-compose exec -T postgres pg_isready -U offgridflow 2>$null
        if ($LASTEXITCODE -eq 0) { break }
    } catch {}
    Write-Host -NoNewline "."
    Start-Sleep -Seconds 1
    $retries++
}
Write-Host ""
if ($retries -eq $maxRetries) {
    Write-Error "PostgreSQL failed to start"
    exit 1
}
Write-Success "PostgreSQL is ready"

# Wait for Redis to be ready
Write-Info "Waiting for Redis to be ready..."
$retries = 0
while ($retries -lt $maxRetries) {
    try {
        docker-compose exec -T redis redis-cli ping 2>$null
        if ($LASTEXITCODE -eq 0) { break }
    } catch {}
    Write-Host -NoNewline "."
    Start-Sleep -Seconds 1
    $retries++
}
Write-Host ""
if ($retries -eq $maxRetries) {
    Write-Error "Redis failed to start"
    exit 1
}
Write-Success "Redis is ready"

# Start API (which will run migrations)
Write-Info "Starting API server (migrations will run automatically)..."
docker-compose up -d api

# Wait for API to be healthy
Write-Info "Waiting for API to be ready..."
$retries = 0
while ($retries -lt $maxRetries) {
    try {
        $response = Invoke-WebRequest -Uri http://localhost:8080/health -UseBasicParsing -TimeoutSec 2 -ErrorAction SilentlyContinue
        if ($response.StatusCode -eq 200) { break }
    } catch {}
    Write-Host -NoNewline "."
    Start-Sleep -Seconds 2
    $retries++
}
Write-Host ""
if ($retries -eq $maxRetries) {
    Write-Error "API failed to start after $maxRetries attempts"
    docker-compose logs api
    exit 1
}
Write-Success "API is ready"

# Start worker
Write-Info "Starting worker..."
docker-compose up -d worker

# Start web
Write-Info "Starting web frontend..."
docker-compose up -d web

# Wait for web to be ready
Write-Info "Waiting for web frontend to be ready..."
$retries = 0
while ($retries -lt $maxRetries) {
    try {
        $response = Invoke-WebRequest -Uri http://localhost:3000 -UseBasicParsing -TimeoutSec 2 -ErrorAction SilentlyContinue
        if ($response.StatusCode -eq 200) { break }
    } catch {}
    Write-Host -NoNewline "."
    Start-Sleep -Seconds 2
    $retries++
}
Write-Host ""
if ($retries -eq $maxRetries) {
    Write-Warning "Web frontend took longer than expected to start"
}

# Start observability stack
Write-Info "Starting Grafana..."
docker-compose up -d grafana

# Print status
Write-Host ""
Write-Success "=========================================="
Write-Success "  OffGridFlow is now running!           "
Write-Success "=========================================="
Write-Host ""
Write-Info "Services:"
Write-Host "  🌐 Web UI:        http://localhost:3000" -ForegroundColor White
Write-Host "  🔌 API:           http://localhost:8080" -ForegroundColor White
Write-Host "  📊 API Docs:      http://localhost:8080/swagger" -ForegroundColor White
Write-Host "  📈 Grafana:       http://localhost:3001 (admin/admin)" -ForegroundColor White
Write-Host "  🔍 Jaeger:        http://localhost:16686" -ForegroundColor White
Write-Host "  📊 Prometheus:    http://localhost:9090" -ForegroundColor White
Write-Host ""
Write-Info "Database:"
Write-Host "  🐘 PostgreSQL:    localhost:5432" -ForegroundColor White
Write-Host "     Database:      offgridflow" -ForegroundColor White
Write-Host "     User:          offgridflow" -ForegroundColor White
Write-Host "     Password:      changeme" -ForegroundColor White
Write-Host ""
Write-Info "Cache:"
Write-Host "  💾 Redis:         localhost:6379" -ForegroundColor White
Write-Host ""
Write-Info "Useful commands:"
Write-Host "  View logs:        docker-compose logs -f [service]" -ForegroundColor White
Write-Host "  Stop all:         docker-compose down" -ForegroundColor White
Write-Host "  Restart service:  docker-compose restart [service]" -ForegroundColor White
Write-Host "  Run tests:        make test" -ForegroundColor White
Write-Host ""
Write-Info "To stop everything: docker-compose down"
Write-Info "To stop and clean:  docker-compose down -v"
Write-Host ""

# Optionally show logs
if ($args -contains "--logs") {
    Write-Info "Showing logs (Ctrl+C to exit)..."
    docker-compose logs -f
}
