#!/bin/bash
# OffGridFlow - One-Command Local Development Setup
# This script sets up and runs the entire OffGridFlow stack locally

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored messages
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Print banner
echo "=========================================="
echo "  OffGridFlow - Local Development Setup  "
echo "=========================================="
echo ""

# Check prerequisites
info "Checking prerequisites..."

if ! command_exists docker; then
    error "Docker is not installed. Please install Docker Desktop."
    exit 1
fi

if ! command_exists docker-compose; then
    error "Docker Compose is not installed. Please install Docker Compose."
    exit 1
fi

success "All prerequisites satisfied"

# Check if .env exists
if [ ! -f .env ]; then
    warning ".env file not found. Creating from template..."
    if [ -f .env.production.template ]; then
        cp .env.production.template .env
        info "Created .env file. Please update it with your credentials."
        warning "Using default development credentials for now..."
    else
        error ".env.production.template not found!"
        exit 1
    fi
fi

# Stop any existing containers
info "Stopping any existing containers..."
docker-compose down 2>/dev/null || true

# Clean up old volumes (optional)
if [ "$1" == "--clean" ]; then
    warning "Cleaning up old volumes..."
    docker-compose down -v
fi

# Pull latest images
info "Pulling latest images..."
docker-compose pull || warning "Could not pull images, will build locally"

# Build images
info "Building Docker images..."
docker-compose build

# Start services
info "Starting services..."
docker-compose up -d postgres redis jaeger otel-collector prometheus

# Wait for database to be ready
info "Waiting for PostgreSQL to be ready..."
until docker-compose exec -T postgres pg_isready -U offgridflow >/dev/null 2>&1; do
    echo -n "."
    sleep 1
done
echo ""
success "PostgreSQL is ready"

# Wait for Redis to be ready
info "Waiting for Redis to be ready..."
until docker-compose exec -T redis redis-cli ping >/dev/null 2>&1; do
    echo -n "."
    sleep 1
done
echo ""
success "Redis is ready"

# Start API (which will run migrations)
info "Starting API server (migrations will run automatically)..."
docker-compose up -d api

# Wait for API to be healthy
info "Waiting for API to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0
until curl -f http://localhost:8080/health >/dev/null 2>&1; do
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
        error "API failed to start after $MAX_RETRIES attempts"
        docker-compose logs api
        exit 1
    fi
    echo -n "."
    sleep 2
done
echo ""
success "API is ready"

# Start worker
info "Starting worker..."
docker-compose up -d worker

# Start web
info "Starting web frontend..."
docker-compose up -d web

# Wait for web to be ready
info "Waiting for web frontend to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0
until curl -f http://localhost:3000 >/dev/null 2>&1; do
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
        warning "Web frontend took longer than expected to start"
        break
    fi
    echo -n "."
    sleep 2
done
echo ""

# Start observability stack
info "Starting Grafana..."
docker-compose up -d grafana

# Print status
echo ""
success "=========================================="
success "  OffGridFlow is now running!           "
success "=========================================="
echo ""
info "Services:"
echo "  🌐 Web UI:        http://localhost:3000"
echo "  🔌 API:           http://localhost:8080"
echo "  📊 API Docs:      http://localhost:8080/swagger"
echo "  📈 Grafana:       http://localhost:3001 (admin/admin)"
echo "  🔍 Jaeger:        http://localhost:16686"
echo "  📊 Prometheus:    http://localhost:9090"
echo ""
info "Database:"
echo "  🐘 PostgreSQL:    localhost:5432"
echo "     Database:      offgridflow"
echo "     User:          offgridflow"
echo "     Password:      changeme"
echo ""
info "Cache:"
echo "  💾 Redis:         localhost:6379"
echo ""
info "Useful commands:"
echo "  View logs:        docker-compose logs -f [service]"
echo "  Stop all:         docker-compose down"
echo "  Restart service:  docker-compose restart [service]"
echo "  Run tests:        make test"
echo ""
info "To stop everything: docker-compose down"
info "To stop and clean:  docker-compose down -v"
echo ""

# Optionally show logs
if [ "$1" == "--logs" ] || [ "$2" == "--logs" ]; then
    info "Showing logs (Ctrl+C to exit)..."
    docker-compose logs -f
fi
