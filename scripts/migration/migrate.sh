#!/bin/bash
# Database Migration Script for OffGridFlow

set -e

echo "OffGridFlow Database Migration Script"
echo "======================================"

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Set defaults if not provided
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-offgridflow}
DB_NAME=${DB_NAME:-offgridflow}
DB_PASSWORD=${DB_PASSWORD:-changeme}

echo ""
echo "Database Configuration:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo ""

# Check if PostgreSQL is accessible
echo "Testing database connection..."
export PGPASSWORD=$DB_PASSWORD
if ! psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c '\q' 2>/dev/null; then
    echo "Error: Cannot connect to PostgreSQL server"
    echo "Please ensure PostgreSQL is running and credentials are correct"
    exit 1
fi

echo "✓ Database connection successful"
echo ""

# Create database if it doesn't exist
echo "Checking if database exists..."
if ! psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -lqt | cut -d \| -f 1 | grep -qw $DB_NAME; then
    echo "Creating database $DB_NAME..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"
    echo "✓ Database created"
else
    echo "✓ Database already exists"
fi

echo ""

# Run migrations
echo "Running database migrations..."
echo "Applying schema from infra/db/schema.sql..."

if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f infra/db/schema.sql; then
    echo "✓ Migrations completed successfully"
else
    echo "✗ Migration failed"
    exit 1
fi

echo ""

# Verify tables
echo "Verifying tables..."
TABLE_COUNT=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")

echo "✓ Found $TABLE_COUNT tables"

# List created tables
echo ""
echo "Created tables:"
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\dt"

echo ""
echo "======================================"
echo "Migration completed successfully!"
echo "======================================"
