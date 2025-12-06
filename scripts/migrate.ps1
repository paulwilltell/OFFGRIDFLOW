# Database Migration Script for OffGridFlow (PowerShell)
# Uses golang-migrate for database versioning

param(
    [Parameter(Mandatory=$true, Position=0)]
    [ValidateSet('up','down','create','version','force','drop')]
    [string]$Command,
    
    [Parameter(Position=1)]
    [string]$Arg
)

$ErrorActionPreference = "Stop"

# Database connection (from environment or defaults)
$DB_HOST = if ($env:DB_HOST) { $env:DB_HOST } else { "localhost" }
$DB_PORT = if ($env:DB_PORT) { $env:DB_PORT } else { "5432" }
$DB_USER = if ($env:DB_USER) { $env:DB_USER } else { "offgridflow" }
$DB_PASSWORD = if ($env:DB_PASSWORD) { $env:DB_PASSWORD } else { "changeme" }
$DB_NAME = if ($env:DB_NAME) { $env:DB_NAME } else { "offgridflow" }

$DB_URL = "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"
$MIGRATIONS_DIR = "internal\db\migrations"

Write-Host "=== OffGridFlow Database Migrations ===" -ForegroundColor Green
Write-Host "Database: $DB_NAME"
Write-Host "Host: ${DB_HOST}:${DB_PORT}"
Write-Host ""

# Check if migrate is installed
try {
    $null = Get-Command migrate -ErrorAction Stop
} catch {
    Write-Host "Error: golang-migrate not installed" -ForegroundColor Red
    Write-Host ""
    Write-Host "Install with:"
    Write-Host "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    Write-Host ""
    exit 1
}

# Create migrations directory if it doesn't exist
if (!(Test-Path $MIGRATIONS_DIR)) {
    New-Item -ItemType Directory -Force -Path $MIGRATIONS_DIR | Out-Null
}

switch ($Command) {
    'up' {
        if ($Arg) {
            Write-Host "Applying $Arg up migrations..." -ForegroundColor Yellow
            migrate -path $MIGRATIONS_DIR -database $DB_URL up $Arg
        } else {
            Write-Host "Applying all up migrations..." -ForegroundColor Yellow
            migrate -path $MIGRATIONS_DIR -database $DB_URL up
        }
        Write-Host "✅ Migrations applied successfully" -ForegroundColor Green
    }
    
    'down' {
        if ($Arg) {
            Write-Host "Applying $Arg down migrations..." -ForegroundColor Yellow
            migrate -path $MIGRATIONS_DIR -database $DB_URL down $Arg
        } else {
            Write-Host "Applying all down migrations..." -ForegroundColor Yellow
            migrate -path $MIGRATIONS_DIR -database $DB_URL down
        }
        Write-Host "✅ Migrations rolled back successfully" -ForegroundColor Green
    }
    
    'create' {
        if (!$Arg) {
            Write-Host "Error: Migration name required" -ForegroundColor Red
            Write-Host "Usage: .\migrate.ps1 create MIGRATION_NAME"
            exit 1
        }
        
        Write-Host "Creating migration: $Arg" -ForegroundColor Yellow
        migrate create -ext sql -dir $MIGRATIONS_DIR -seq $Arg
        Write-Host "✅ Migration files created in $MIGRATIONS_DIR" -ForegroundColor Green
    }
    
    'version' {
        Write-Host "Current migration version:" -ForegroundColor Yellow
        migrate -path $MIGRATIONS_DIR -database $DB_URL version
    }
    
    'force' {
        if (!$Arg) {
            Write-Host "Error: Version number required" -ForegroundColor Red
            Write-Host "Usage: .\migrate.ps1 force VERSION"
            exit 1
        }
        
        Write-Host "Forcing version to: $Arg" -ForegroundColor Yellow
        Write-Host "WARNING: This will not run migrations, only set version" -ForegroundColor Red
        $confirm = Read-Host "Are you sure? (yes/no)"
        
        if ($confirm -eq "yes") {
            migrate -path $MIGRATIONS_DIR -database $DB_URL force $Arg
            Write-Host "✅ Version forced to $Arg" -ForegroundColor Green
        } else {
            Write-Host "Cancelled"
        }
    }
    
    'drop' {
        Write-Host "WARNING: This will DROP ALL TABLES in $DB_NAME" -ForegroundColor Red
        Write-Host "This action CANNOT be undone!" -ForegroundColor Red
        $confirm = Read-Host "Type 'DROP ALL DATA' to confirm"
        
        if ($confirm -eq "DROP ALL DATA") {
            migrate -path $MIGRATIONS_DIR -database $DB_URL drop
            Write-Host "✅ Database dropped" -ForegroundColor Green
        } else {
            Write-Host "Cancelled"
        }
    }
}
