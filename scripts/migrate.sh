#!/bin/bash
# Database Migration Script for OffGridFlow
# Uses golang-migrate for database versioning

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MIGRATIONS_DIR="$PROJECT_ROOT/internal/db/migrations"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Database connection (from environment or defaults)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-offgridflow}"
DB_PASSWORD="${DB_PASSWORD:-changeme}"
DB_NAME="${DB_NAME:-offgridflow}"
DB_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

echo -e "${GREEN}=== OffGridFlow Database Migrations ===${NC}"
echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo ""

# Check if migrate is installed
if ! command -v migrate &> /dev/null; then
    echo -e "${RED}Error: golang-migrate not installed${NC}"
    echo ""
    echo "Install with:"
    echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    echo ""
    exit 1
fi

# Create migrations directory if it doesn't exist
mkdir -p "$MIGRATIONS_DIR"

# Function to show usage
usage() {
    echo "Usage: $0 {up|down|create|version|force|drop}"
    echo ""
    echo "Commands:"
    echo "  up [N]       Apply all (or N) up migrations"
    echo "  down [N]     Apply all (or N) down migrations"
    echo "  create NAME  Create a new migration file"
    echo "  version      Show current migration version"
    echo "  force V      Force set version without running migrations"
    echo "  drop         Drop everything in the database (DANGEROUS!)"
    echo ""
    exit 1
}

# Check arguments
if [ $# -eq 0 ]; then
    usage
fi

COMMAND=$1
shift

case $COMMAND in
    up)
        if [ $# -eq 0 ]; then
            echo -e "${YELLOW}Applying all up migrations...${NC}"
            migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" up
        else
            echo -e "${YELLOW}Applying $1 up migrations...${NC}"
            migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" up $1
        fi
        echo -e "${GREEN}✅ Migrations applied successfully${NC}"
        ;;
        
    down)
        if [ $# -eq 0 ]; then
            echo -e "${YELLOW}Applying all down migrations...${NC}"
            migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" down
        else
            echo -e "${YELLOW}Applying $1 down migrations...${NC}"
            migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" down $1
        fi
        echo -e "${GREEN}✅ Migrations rolled back successfully${NC}"
        ;;
        
    create)
        if [ $# -eq 0 ]; then
            echo -e "${RED}Error: Migration name required${NC}"
            echo "Usage: $0 create MIGRATION_NAME"
            exit 1
        fi
        
        NAME=$1
        echo -e "${YELLOW}Creating migration: $NAME${NC}"
        migrate create -ext sql -dir "$MIGRATIONS_DIR" -seq "$NAME"
        echo -e "${GREEN}✅ Migration files created in $MIGRATIONS_DIR${NC}"
        ;;
        
    version)
        echo -e "${YELLOW}Current migration version:${NC}"
        migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" version
        ;;
        
    force)
        if [ $# -eq 0 ]; then
            echo -e "${RED}Error: Version number required${NC}"
            echo "Usage: $0 force VERSION"
            exit 1
        fi
        
        VERSION=$1
        echo -e "${YELLOW}Forcing version to: $VERSION${NC}"
        echo -e "${RED}WARNING: This will not run migrations, only set version${NC}"
        read -p "Are you sure? (yes/no): " CONFIRM
        
        if [ "$CONFIRM" = "yes" ]; then
            migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" force $VERSION
            echo -e "${GREEN}✅ Version forced to $VERSION${NC}"
        else
            echo "Cancelled"
        fi
        ;;
        
    drop)
        echo -e "${RED}WARNING: This will DROP ALL TABLES in $DB_NAME${NC}"
        echo -e "${RED}This action CANNOT be undone!${NC}"
        read -p "Type 'DROP ALL DATA' to confirm: " CONFIRM
        
        if [ "$CONFIRM" = "DROP ALL DATA" ]; then
            migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" drop
            echo -e "${GREEN}✅ Database dropped${NC}"
        else
            echo "Cancelled"
        fi
        ;;
        
    *)
        echo -e "${RED}Unknown command: $COMMAND${NC}"
        usage
        ;;
esac
