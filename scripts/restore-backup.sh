#!/bin/bash
# OffGridFlow Database Restore Script
# Restores PostgreSQL database from S3 backup
# Usage: ./restore-backup.sh [backup_file] [target_database]

set -euo pipefail

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
S3_BUCKET="${S3_BUCKET:-offgridflow-backups}"
AWS_REGION="${AWS_REGION:-us-east-1}"
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_USER="${POSTGRES_USER:-offgridflow}"
POSTGRES_DB="${POSTGRES_DB:-offgridflow}"

# Function to print colored output
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to list available backups
list_backups() {
    log_info "Available backups in s3://${S3_BUCKET}/backups/postgres/:"
    aws s3 ls "s3://${S3_BUCKET}/backups/postgres/" --region "${AWS_REGION}" | \
        grep "offgridflow_backup_" | \
        awk '{print $4, "(" $3 " bytes)"}'
}

# Function to download backup
download_backup() {
    local backup_file=$1
    local local_path="/tmp/${backup_file}"
    
    log_info "Downloading backup: ${backup_file}"
    aws s3 cp "s3://${S3_BUCKET}/backups/postgres/${backup_file}" \
        "${local_path}" \
        --region "${AWS_REGION}"
    
    if [ ! -f "${local_path}" ]; then
        log_error "Failed to download backup file"
        exit 1
    fi
    
    log_info "Downloaded to: ${local_path}"
    echo "${local_path}"
}

# Function to restore database
restore_database() {
    local backup_path=$1
    local target_db=$2
    
    log_warn "This will restore ${target_db} database from ${backup_path}"
    log_warn "Current database contents will be LOST!"
    read -p "Are you sure you want to continue? (yes/no): " confirm
    
    if [ "$confirm" != "yes" ]; then
        log_info "Restore cancelled"
        exit 0
    fi
    
    # Create backup of current database before restore
    log_info "Creating safety backup of current database..."
    TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    SAFETY_BACKUP="/tmp/pre_restore_backup_${TIMESTAMP}.sql.gz"
    
    PGPASSWORD="${POSTGRES_PASSWORD}" pg_dump \
        -h "${POSTGRES_HOST}" \
        -p "${POSTGRES_PORT}" \
        -U "${POSTGRES_USER}" \
        -d "${target_db}" \
        --format=custom \
        --compress=9 \
        --file="${SAFETY_BACKUP}" || log_warn "Could not create safety backup (database may not exist)"
    
    # Terminate existing connections
    log_info "Terminating existing connections..."
    PGPASSWORD="${POSTGRES_PASSWORD}" psql \
        -h "${POSTGRES_HOST}" \
        -p "${POSTGRES_PORT}" \
        -U "${POSTGRES_USER}" \
        -d postgres \
        -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${target_db}' AND pid <> pg_backend_pid();" || true
    
    # Drop and recreate database
    log_info "Dropping and recreating database..."
    PGPASSWORD="${POSTGRES_PASSWORD}" psql \
        -h "${POSTGRES_HOST}" \
        -p "${POSTGRES_PORT}" \
        -U "${POSTGRES_USER}" \
        -d postgres \
        -c "DROP DATABASE IF EXISTS ${target_db};"
    
    PGPASSWORD="${POSTGRES_PASSWORD}" psql \
        -h "${POSTGRES_HOST}" \
        -p "${POSTGRES_PORT}" \
        -U "${POSTGRES_USER}" \
        -d postgres \
        -c "CREATE DATABASE ${target_db};"
    
    # Restore from backup
    log_info "Restoring database from backup..."
    PGPASSWORD="${POSTGRES_PASSWORD}" pg_restore \
        -h "${POSTGRES_HOST}" \
        -p "${POSTGRES_PORT}" \
        -U "${POSTGRES_USER}" \
        -d "${target_db}" \
        --verbose \
        --no-owner \
        --no-acl \
        "${backup_path}"
    
    if [ $? -eq 0 ]; then
        log_info "✅ Database restored successfully!"
        log_info "Safety backup saved at: ${SAFETY_BACKUP}"
    else
        log_error "❌ Restore failed!"
        log_info "Attempting to restore from safety backup..."
        
        PGPASSWORD="${POSTGRES_PASSWORD}" pg_restore \
            -h "${POSTGRES_HOST}" \
            -p "${POSTGRES_PORT}" \
            -U "${POSTGRES_USER}" \
            -d "${target_db}" \
            --clean \
            "${SAFETY_BACKUP}" || log_error "Failed to restore safety backup!"
        
        exit 1
    fi
    
    # Verify restore
    log_info "Verifying restore..."
    TABLE_COUNT=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql \
        -h "${POSTGRES_HOST}" \
        -p "${POSTGRES_PORT}" \
        -U "${POSTGRES_USER}" \
        -d "${target_db}" \
        -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")
    
    log_info "Tables restored: ${TABLE_COUNT}"
    
    # Cleanup
    log_info "Cleaning up temporary files..."
    rm -f "${backup_path}"
}

# Main execution
main() {
    log_info "OffGridFlow Database Restore Tool"
    log_info "=================================="
    
    # Check if backup file is provided
    if [ $# -eq 0 ]; then
        log_info "No backup file specified. Listing available backups:"
        list_backups
        echo ""
        log_info "Usage: $0 <backup_file> [target_database]"
        log_info "Example: $0 offgridflow_backup_20250101_020000.sql.gz offgridflow_restore"
        exit 0
    fi
    
    BACKUP_FILE=$1
    TARGET_DB=${2:-${POSTGRES_DB}}
    
    # Check for required environment variables
    if [ -z "${POSTGRES_PASSWORD:-}" ]; then
        log_error "POSTGRES_PASSWORD environment variable is required"
        exit 1
    fi
    
    # Check if AWS CLI is installed
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI is not installed. Please install it first."
        exit 1
    fi
    
    # Check if pg_restore is installed
    if ! command -v pg_restore &> /dev/null; then
        log_error "pg_restore is not installed. Please install PostgreSQL client tools."
        exit 1
    fi
    
    # Download backup
    LOCAL_BACKUP=$(download_backup "${BACKUP_FILE}")
    
    # Restore database
    restore_database "${LOCAL_BACKUP}" "${TARGET_DB}"
    
    log_info "✅ Restore process completed successfully!"
}

main "$@"
