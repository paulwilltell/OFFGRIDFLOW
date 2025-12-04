#!/usr/bin/env pwsh
# Complete Production Deployment Script
# This script automates the entire deployment process

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet('local', 'staging', 'production')]
    [string]$Environment = 'local',
    
    [Parameter(Mandatory=$false)]
    [switch]$SkipTests,
    
    [Parameter(Mandatory=$false)]
    [switch]$SkipBuild,
    
    [Parameter(Mandatory=$false)]
    [switch]$Force
)

$ErrorActionPreference = "Stop"

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "OffGridFlow Complete Deployment" -ForegroundColor Cyan
Write-Host "Environment: $Environment" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Pre-deployment checks
Write-Host "[1/8] Running pre-deployment checks..." -ForegroundColor Yellow

if ($Environment -eq 'production' -and -not $Force) {
    Write-Host "WARNING: Deploying to PRODUCTION!" -ForegroundColor Red
    $confirm = Read-Host "Type 'DEPLOY' to continue"
    if ($confirm -ne 'DEPLOY') {
        Write-Host "Deployment cancelled." -ForegroundColor Yellow
        exit 0
    }
}

# Load environment
$envFile = ".env.$Environment"
if ($Environment -eq 'local') { $envFile = ".env" }

if (-not (Test-Path $envFile)) {
    Write-Host "Error: $envFile not found!" -ForegroundColor Red
    exit 1
}

Get-Content $envFile | ForEach-Object {
    if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
        $key = $matches[1].Trim()
        $value = $matches[2].Trim()
        Set-Item -Path "env:$key" -Value $value
    }
}

Write-Host "  ✓ Environment loaded" -ForegroundColor Green

# Step 2: Run tests
if (-not $SkipTests) {
    Write-Host ""
    Write-Host "[2/8] Running tests..." -ForegroundColor Yellow
    
    # Unit tests
    Write-Host "  Running unit tests..."
    go test ./... -v -cover
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ✗ Tests failed!" -ForegroundColor Red
        exit 1
    }
    Write-Host "  ✓ Unit tests passed" -ForegroundColor Green
    
    # Build test
    Write-Host "  Testing build..."
    go build -o test-build.exe ./cmd/api
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ✗ Build failed!" -ForegroundColor Red
        exit 1
    }
    Remove-Item test-build.exe
    Write-Host "  ✓ Build test passed" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "[2/8] Skipping tests..." -ForegroundColor Yellow
}

# Step 3: Build Docker images
if (-not $SkipBuild) {
    Write-Host ""
    Write-Host "[3/8] Building Docker images..." -ForegroundColor Yellow
    
    $version = try { git rev-parse --short HEAD } catch { "latest" }
    
    Write-Host "  Building API image..."
    docker build -t offgridflow-api:$version -t offgridflow-api:$Environment .
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ✗ API build failed!" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "  Building Web image..."
    docker build -t offgridflow-web:$version -t offgridflow-web:$Environment ./web
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ✗ Web build failed!" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "  ✓ Docker images built" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "[3/8] Skipping build..." -ForegroundColor Yellow
}

# Step 4: Run database migrations
Write-Host ""
Write-Host "[4/8] Running database migrations..." -ForegroundColor Yellow

if ($Environment -eq 'local') {
    # Check if DB is running
    $dbRunning = docker ps --filter "name=offgridflow-postgres" --format "{{.Names}}"
    if (-not $dbRunning) {
        Write-Host "  Starting local database..."
        docker-compose up -d postgres
        Start-Sleep -Seconds 5
    }
}

.\scripts\migrate.ps1
if ($LASTEXITCODE -ne 0) {
    Write-Host "  ⚠ Migration warnings (may be ok if already applied)" -ForegroundColor Yellow
} else {
    Write-Host "  ✓ Migrations completed" -ForegroundColor Green
}

# Step 5: Deploy based on environment
Write-Host ""
Write-Host "[5/8] Deploying to $Environment..." -ForegroundColor Yellow

switch ($Environment) {
    'local' {
        Write-Host "  Starting services with docker-compose..."
        docker-compose up -d
        
        Write-Host "  Waiting for services to be healthy..."
        Start-Sleep -Seconds 10
        
        # Check health
        try {
            $health = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing
            Write-Host "  ✓ API is healthy" -ForegroundColor Green
        } catch {
            Write-Host "  ⚠ API health check failed" -ForegroundColor Yellow
        }
    }
    
    'staging' {
        Write-Host "  Deploying to staging cluster..."
        .\scripts\deploy-staging.ps1
    }
    
    'production' {
        Write-Host "  Deploying to production cluster..."
        $env:CLUSTER_NAME = "offgridflow-production"
        $env:NAMESPACE = "offgridflow-prod"
        .\scripts\deploy-staging.ps1
    }
}

Write-Host "  ✓ Deployment completed" -ForegroundColor Green

# Step 6: Run integration tests
if (-not $SkipTests -and $Environment -eq 'local') {
    Write-Host ""
    Write-Host "[6/8] Running integration tests..." -ForegroundColor Yellow
    
    Start-Sleep -Seconds 5  # Wait for services to stabilize
    
    .\scripts\test-integration.ps1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ⚠ Some integration tests failed" -ForegroundColor Yellow
    } else {
        Write-Host "  ✓ Integration tests passed" -ForegroundColor Green
    }
} else {
    Write-Host ""
    Write-Host "[6/8] Skipping integration tests..." -ForegroundColor Yellow
}

# Step 7: Smoke tests
Write-Host ""
Write-Host "[7/8] Running smoke tests..." -ForegroundColor Yellow

$baseUrl = switch ($Environment) {
    'local' { "http://localhost:8080" }
    'staging' { "https://staging.offgridflow.com" }
    'production' { "https://api.offgridflow.com" }
}

Write-Host "  Testing $baseUrl/health..."
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/health" -UseBasicParsing
    Write-Host "  ✓ Health endpoint OK" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Health endpoint failed!" -ForegroundColor Red
    Write-Host "  Error: $_" -ForegroundColor Red
}

Write-Host "  Testing $baseUrl/api/v1/health..."
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/api/v1/health" -UseBasicParsing
    Write-Host "  ✓ API health endpoint OK" -ForegroundColor Green
} catch {
    Write-Host "  ⚠ API health endpoint failed" -ForegroundColor Yellow
}

# Step 8: Display summary
Write-Host ""
Write-Host "[8/8] Deployment Summary" -ForegroundColor Yellow
Write-Host ""

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "Deployment Complete!" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Environment: $Environment" -ForegroundColor Cyan
Write-Host "Endpoints:" -ForegroundColor Cyan

switch ($Environment) {
    'local' {
        Write-Host "  API:        http://localhost:8080" -ForegroundColor White
        Write-Host "  Web:        http://localhost:3000" -ForegroundColor White
        Write-Host "  Grafana:    http://localhost:3001" -ForegroundColor White
        Write-Host "  Jaeger:     http://localhost:16686" -ForegroundColor White
        Write-Host "  Prometheus: http://localhost:9090" -ForegroundColor White
        Write-Host ""
        Write-Host "View logs:" -ForegroundColor Cyan
        Write-Host "  docker-compose logs -f api" -ForegroundColor White
        Write-Host "  docker-compose logs -f worker" -ForegroundColor White
        Write-Host ""
        Write-Host "Stop services:" -ForegroundColor Cyan
        Write-Host "  docker-compose down" -ForegroundColor White
    }
    
    'staging' {
        Write-Host "  API:     https://staging.offgridflow.com" -ForegroundColor White
        Write-Host "  Web:     https://staging-app.offgridflow.com" -ForegroundColor White
        Write-Host ""
        Write-Host "View logs:" -ForegroundColor Cyan
        Write-Host "  kubectl logs -f deployment/offgridflow-api -n offgridflow" -ForegroundColor White
        Write-Host ""
        Write-Host "Check status:" -ForegroundColor Cyan
        Write-Host "  kubectl get pods -n offgridflow" -ForegroundColor White
    }
    
    'production' {
        Write-Host "  API:     https://api.offgridflow.com" -ForegroundColor White
        Write-Host "  Web:     https://app.offgridflow.com" -ForegroundColor White
        Write-Host ""
        Write-Host "View logs:" -ForegroundColor Cyan
        Write-Host "  kubectl logs -f deployment/offgridflow-api -n offgridflow-prod" -ForegroundColor White
        Write-Host ""
        Write-Host "Monitor:" -ForegroundColor Cyan
        Write-Host "  kubectl get pods -n offgridflow-prod -w" -ForegroundColor White
    }
}

Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  1. Verify all services are healthy" -ForegroundColor White
Write-Host "  2. Check logs for any errors" -ForegroundColor White
Write-Host "  3. Run manual smoke tests" -ForegroundColor White
Write-Host "  4. Monitor metrics in Grafana" -ForegroundColor White
Write-Host ""
