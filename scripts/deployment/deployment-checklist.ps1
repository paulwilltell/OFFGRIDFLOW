#!/usr/bin/env pwsh
# OffGridFlow Production Deployment Checklist
# This script validates that all prerequisites are met before deployment

param(
    [string]$EnvFile = ".env.production"
)

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "OffGridFlow Production Deployment Checklist" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

$ErrorCount = 0
$WarningCount = 0

# Helper function to check required env var
function Test-EnvVar {
    param($Name, $Value, $Pattern = $null)
    
    if (-not $Value -or $Value -eq "") {
        Write-Host "  ✗ $Name is not set" -ForegroundColor Red
        $script:ErrorCount++
        return $false
    }
    
    if ($Value -match "CHANGE|REPLACE|YOUR_|EXAMPLE|TODO|FIXME") {
        Write-Host "  ⚠ $Name needs to be updated (contains placeholder)" -ForegroundColor Yellow
        $script:WarningCount++
        return $false
    }
    
    if ($Pattern -and $Value -notmatch $Pattern) {
        Write-Host "  ⚠ $Name format may be incorrect" -ForegroundColor Yellow
        $script:WarningCount++
        return $false
    }
    
    Write-Host "  ✓ $Name is configured" -ForegroundColor Green
    return $true
}

# Load environment file
Write-Host "Checking environment configuration..." -ForegroundColor Yellow
if (Test-Path $EnvFile) {
    Get-Content $EnvFile | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            $key = $matches[1].Trim()
            $value = $matches[2].Trim()
            Set-Item -Path "env:$key" -Value $value
        }
    }
    Write-Host "  ✓ Loaded $EnvFile" -ForegroundColor Green
} else {
    Write-Host "  ✗ $EnvFile not found!" -ForegroundColor Red
    Write-Host "  Copy .env.production.template to $EnvFile and configure" -ForegroundColor Yellow
    $ErrorCount++
}
Write-Host ""

# Check critical environment variables
Write-Host "Validating critical environment variables..." -ForegroundColor Yellow

# Database
Test-EnvVar "DATABASE_URL" $env:DATABASE_URL "^postgresql://"
Test-EnvVar "DB_PASSWORD" $env:DB_PASSWORD
if ($env:ENV -eq "production" -and $env:DB_SSLMODE -ne "require") {
    Write-Host "  ⚠ DB_SSLMODE should be 'require' in production" -ForegroundColor Yellow
    $WarningCount++
}

# JWT & Security
Test-EnvVar "JWT_SECRET" $env:JWT_SECRET
Test-EnvVar "SESSION_SECRET" $env:SESSION_SECRET
Test-EnvVar "ENCRYPTION_KEY" $env:ENCRYPTION_KEY

if ($env:JWT_SECRET.Length -lt 32) {
    Write-Host "  ⚠ JWT_SECRET should be at least 32 characters" -ForegroundColor Yellow
    $WarningCount++
}

# Stripe
Test-EnvVar "STRIPE_SECRET_KEY" $env:STRIPE_SECRET_KEY "^sk_"
Test-EnvVar "STRIPE_WEBHOOK_SECRET" $env:STRIPE_WEBHOOK_SECRET "^whsec_"
if ($env:ENV -eq "production" -and $env:STRIPE_SECRET_KEY -match "^sk_test_") {
    Write-Host "  ✗ Using Stripe TEST key in production!" -ForegroundColor Red
    $ErrorCount++
}

# Email
Test-EnvVar "SMTP_HOST" $env:SMTP_HOST
Test-EnvVar "SMTP_PASSWORD" $env:SMTP_PASSWORD
Test-EnvVar "SMTP_FROM_EMAIL" $env:SMTP_FROM_EMAIL

# Redis
Test-EnvVar "REDIS_URL" $env:REDIS_URL "^redis://"

# OTEL
Test-EnvVar "OTEL_EXPORTER_OTLP_ENDPOINT" $env:OTEL_EXPORTER_OTLP_ENDPOINT

Write-Host ""

# Check for insecure defaults
Write-Host "Checking for insecure defaults..." -ForegroundColor Yellow
$insecureFound = $false

if ($env:DB_PASSWORD -eq "changeme") {
    Write-Host "  ✗ Using default database password!" -ForegroundColor Red
    $ErrorCount++
    $insecureFound = $true
}

if ($env:JWT_SECRET -match "CHANGE_THIS") {
    Write-Host "  ✗ Using default JWT secret!" -ForegroundColor Red
    $ErrorCount++
    $insecureFound = $true
}

if (-not $insecureFound) {
    Write-Host "  ✓ No obvious insecure defaults found" -ForegroundColor Green
}

Write-Host ""

# Check prerequisites
Write-Host "Checking system prerequisites..." -ForegroundColor Yellow

# Go
try {
    $goVersion = go version
    Write-Host "  ✓ Go installed: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Go not found" -ForegroundColor Red
    $ErrorCount++
}

# Docker
try {
    $dockerVersion = docker --version
    Write-Host "  ✓ Docker installed: $dockerVersion" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Docker not found" -ForegroundColor Red
    $ErrorCount++
}

# Docker Compose
try {
    $composeVersion = docker compose version
    Write-Host "  ✓ Docker Compose installed: $composeVersion" -ForegroundColor Green
} catch {
    Write-Host "  ⚠ Docker Compose not found" -ForegroundColor Yellow
    $WarningCount++
}

# PostgreSQL client
try {
    $psqlVersion = psql --version
    Write-Host "  ✓ PostgreSQL client installed: $psqlVersion" -ForegroundColor Green
} catch {
    Write-Host "  ⚠ psql not found (needed for migrations)" -ForegroundColor Yellow
    $WarningCount++
}

# kubectl (for K8s deployments)
try {
    $kubectlVersion = kubectl version --client --short 2>$null
    Write-Host "  ✓ kubectl installed: $kubectlVersion" -ForegroundColor Green
} catch {
    Write-Host "  ⚠ kubectl not found (needed for K8s deployments)" -ForegroundColor Yellow
    $WarningCount++
}

Write-Host ""

# Check database connectivity
Write-Host "Testing database connectivity..." -ForegroundColor Yellow
if ($env:DATABASE_URL) {
    try {
        $env:PGPASSWORD = $env:DB_PASSWORD
        $result = psql $env:DATABASE_URL -c "SELECT 1" 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  ✓ Database connection successful" -ForegroundColor Green
        } else {
            Write-Host "  ✗ Cannot connect to database" -ForegroundColor Red
            $ErrorCount++
        }
    } catch {
        Write-Host "  ⚠ Could not test database connection" -ForegroundColor Yellow
        $WarningCount++
    }
} else {
    Write-Host "  ⚠ DATABASE_URL not set, skipping connectivity test" -ForegroundColor Yellow
    $WarningCount++
}

Write-Host ""

# Check if migrations are up to date
Write-Host "Checking database schema..." -ForegroundColor Yellow
if (Test-Path "infra/db/schema.sql") {
    Write-Host "  ✓ Schema file exists" -ForegroundColor Green
    
    # Check for required tables
    $requiredTables = @("tenants", "users", "activities", "emissions", "jobs", "audit_logs", "api_keys")
    foreach ($table in $requiredTables) {
        $found = Select-String -Path "infra/db/schema.sql" -Pattern "CREATE TABLE.*$table" -Quiet
        if ($found) {
            Write-Host "  ✓ Table '$table' in schema" -ForegroundColor Green
        } else {
            Write-Host "  ⚠ Table '$table' not found in schema" -ForegroundColor Yellow
            $WarningCount++
        }
    }
} else {
    Write-Host "  ✗ Schema file not found!" -ForegroundColor Red
    $ErrorCount++
}

Write-Host ""

# Check build status
Write-Host "Testing build..." -ForegroundColor Yellow
$buildOutput = go build -o offgridflow-api.exe ./cmd/api 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✓ Build successful" -ForegroundColor Green
    Remove-Item offgridflow-api.exe -ErrorAction SilentlyContinue
} else {
    Write-Host "  ✗ Build failed!" -ForegroundColor Red
    Write-Host $buildOutput -ForegroundColor Red
    $ErrorCount++
}

Write-Host ""

# Check critical files
Write-Host "Checking critical files..." -ForegroundColor Yellow
$criticalFiles = @(
    "cmd/api/main.go",
    "docker-compose.yml",
    "Dockerfile",
    "go.mod",
    "infra/k8s/api-deployment.yaml",
    "scripts/migrate.ps1"
)

foreach ($file in $criticalFiles) {
    if (Test-Path $file) {
        Write-Host "  ✓ $file exists" -ForegroundColor Green
    } else {
        Write-Host "  ✗ $file missing!" -ForegroundColor Red
        $ErrorCount++
    }
}

Write-Host ""

# Summary
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "Deployment Checklist Summary" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

if ($ErrorCount -eq 0 -and $WarningCount -eq 0) {
    Write-Host "✓ ALL CHECKS PASSED!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Ready to deploy!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "  1. Run migrations: .\scripts\migrate.ps1"
    Write-Host "  2. Start services: docker-compose up -d"
    Write-Host "  3. Run integration tests: .\scripts\test-integration.ps1"
    Write-Host "  4. Deploy to staging: .\scripts\deploy-staging.ps1"
    exit 0
} elseif ($ErrorCount -eq 0) {
    Write-Host "⚠ WARNINGS FOUND: $WarningCount" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "You can proceed, but review warnings above." -ForegroundColor Yellow
    exit 0
} else {
    Write-Host "✗ ERRORS FOUND: $ErrorCount" -ForegroundColor Red
    Write-Host "⚠ WARNINGS: $WarningCount" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Please fix errors before deployment!" -ForegroundColor Red
    exit 1
}
