# Database Migration Script for OffGridFlow (PowerShell)

Write-Host "OffGridFlow Database Migration Script" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# Load environment variables from .env file
if (Test-Path .env) {
    Get-Content .env | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            $key = $matches[1].Trim()
            $value = $matches[2].Trim()
            [Environment]::SetEnvironmentVariable($key, $value, 'Process')
        }
    }
}

# Set defaults if not provided
$DB_HOST = if ($env:DB_HOST) { $env:DB_HOST } else { "localhost" }
$DB_PORT = if ($env:DB_PORT) { $env:DB_PORT } else { "5432" }
$DB_USER = if ($env:DB_USER) { $env:DB_USER } else { "offgridflow" }
$DB_NAME = if ($env:DB_NAME) { $env:DB_NAME } else { "offgridflow" }
$DB_PASSWORD = if ($env:DB_PASSWORD) { $env:DB_PASSWORD } else { "changeme" }

Write-Host "Database Configuration:"
Write-Host "  Host: $DB_HOST"
Write-Host "  Port: $DB_PORT"
Write-Host "  Database: $DB_NAME"
Write-Host "  User: $DB_USER"
Write-Host ""

# Set PostgreSQL password environment variable
$env:PGPASSWORD = $DB_PASSWORD

# Check if psql is available
try {
    $null = Get-Command psql -ErrorAction Stop
} catch {
    Write-Host "Error: psql command not found. Please install PostgreSQL client tools." -ForegroundColor Red
    Write-Host "Download from: https://www.postgresql.org/download/windows/" -ForegroundColor Yellow
    exit 1
}

# Test database connection
Write-Host "Testing database connection..."
$testConn = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "\q" 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Cannot connect to PostgreSQL server" -ForegroundColor Red
    Write-Host "Please ensure PostgreSQL is running and credentials are correct" -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ Database connection successful" -ForegroundColor Green
Write-Host ""

# Check if database exists
Write-Host "Checking if database exists..."
$dbExists = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -t -c "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'" 2>&1

if ($dbExists -notmatch "1") {
    Write-Host "Creating database $DB_NAME..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Database created" -ForegroundColor Green
    } else {
        Write-Host "✗ Failed to create database" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "✓ Database already exists" -ForegroundColor Green
}

Write-Host ""

# Run migrations
Write-Host "Running database migrations..."
Write-Host "Applying schema from infra/db/schema.sql..."

psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f infra/db/schema.sql

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Migrations completed successfully" -ForegroundColor Green
} else {
    Write-Host "✗ Migration failed" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Verify tables
Write-Host "Verifying tables..."
$tableCount = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';"

Write-Host "✓ Found $($tableCount.Trim()) tables" -ForegroundColor Green

# List created tables
Write-Host ""
Write-Host "Created tables:"
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\dt"

Write-Host ""
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "Migration completed successfully!" -ForegroundColor Green
Write-Host "======================================" -ForegroundColor Cyan
