#!/bin/bash
# Execute E2E tests for OffGridFlow Web Application

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
BROWSER="chromium"
UI_MODE=false
HEADED=false
DEBUG=false
REPORT=false
PROJECT=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --browser)
            BROWSER="$2"
            shift 2
            ;;
        --ui)
            UI_MODE=true
            shift
            ;;
        --headed)
            HEADED=true
            shift
            ;;
        --debug)
            DEBUG=true
            shift
            ;;
        --report)
            REPORT=true
            shift
            ;;
        --project)
            PROJECT="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --browser <name>   Browser to run tests in (chromium, firefox, webkit, all)"
            echo "  --ui              Run tests in UI mode"
            echo "  --headed          Run tests in headed mode (show browser)"
            echo "  --debug           Run tests in debug mode"
            echo "  --report          Show test report after completion"
            echo "  --project <name>  Run specific project/browser only"
            echo "  --help            Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Run with --help for usage information"
            exit 1
            ;;
    esac
done

echo_step() {
    echo -e "${CYAN}===> $1${NC}"
}

echo_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

echo_failure() {
    echo -e "${RED}✗ $1${NC}"
}

echo_info() {
    echo -e "${CYAN}ℹ $1${NC}"
}

echo_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Change to web directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEB_DIR="$SCRIPT_DIR/../web"

if [ ! -d "$WEB_DIR" ]; then
    echo_failure "Web directory not found: $WEB_DIR"
    exit 1
fi

cd "$WEB_DIR"

echo_step "OffGridFlow E2E Test Runner"
echo ""

# Check if Playwright is installed
echo_step "Checking Playwright installation..."
if ! npm list @playwright/test > /dev/null 2>&1; then
    echo_info "Installing Playwright..."
    npm install --save-dev @playwright/test
    
    echo_info "Installing Playwright browsers..."
    npx playwright install --with-deps
else
    echo_success "Playwright is installed"
fi

# Check if Next.js dev server is running
echo_step "Checking if Next.js dev server is running..."
SERVER_RUNNING=false

if curl -s http://localhost:3000 > /dev/null 2>&1; then
    SERVER_RUNNING=true
    echo_success "Dev server is running"
else
    echo_warning "Dev server is not running"
    echo_info "Starting Next.js dev server..."
    
    # Start dev server in background
    npm run dev > /dev/null 2>&1 &
    DEV_SERVER_PID=$!
    
    echo_info "Waiting for dev server to start (PID: $DEV_SERVER_PID)..."
    MAX_WAIT=60
    WAITED=0
    
    while [ $WAITED -lt $MAX_WAIT ]; do
        sleep 2
        WAITED=$((WAITED + 2))
        
        if curl -s http://localhost:3000 > /dev/null 2>&1; then
            echo_success "Dev server is ready!"
            SERVER_RUNNING=true
            break
        fi
        
        echo -n "."
    done
    
    echo ""
    
    if [ "$SERVER_RUNNING" = false ]; then
        echo_failure "Dev server failed to start within $MAX_WAIT seconds"
        kill $DEV_SERVER_PID 2>/dev/null || true
        exit 1
    fi
fi

# Build test command
echo_step "Running E2E tests..."
TEST_COMMAND="npx playwright test"

if [ "$UI_MODE" = true ]; then
    TEST_COMMAND="$TEST_COMMAND --ui"
    echo_info "Running in UI mode"
elif [ "$DEBUG" = true ]; then
    TEST_COMMAND="$TEST_COMMAND --debug"
    echo_info "Running in debug mode"
else
    # Add browser selection
    if [ "$BROWSER" != "all" ]; then
        TEST_COMMAND="$TEST_COMMAND --project=$BROWSER"
        echo_info "Running tests in $BROWSER"
    else
        echo_info "Running tests in all browsers"
    fi
    
    if [ "$HEADED" = true ]; then
        TEST_COMMAND="$TEST_COMMAND --headed"
        echo_info "Running in headed mode"
    fi
    
    if [ -n "$PROJECT" ]; then
        TEST_COMMAND="$TEST_COMMAND --project=$PROJECT"
    fi
fi

echo ""
echo_info "Executing: $TEST_COMMAND"
echo ""

# Run tests
TEST_START_TIME=$(date +%s)
eval $TEST_COMMAND
TEST_EXIT_CODE=$?
TEST_END_TIME=$(date +%s)
TEST_DURATION=$((TEST_END_TIME - TEST_START_TIME))

echo ""
echo -e "${CYAN}========================================${NC}"
echo ""

if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo_success "All E2E tests passed!"
    echo -e "${GREEN}Duration: $(($TEST_DURATION / 60))m $(($TEST_DURATION % 60))s${NC}"
else
    echo_failure "Some E2E tests failed"
    echo -e "${RED}Duration: $(($TEST_DURATION / 60))m $(($TEST_DURATION % 60))s${NC}"
fi

echo ""

# Show report option
if [ "$REPORT" = true ] || { [ $TEST_EXIT_CODE -ne 0 ] && [ "$UI_MODE" = false ] && [ "$DEBUG" = false ]; }; then
    echo_info "Opening test report..."
    npx playwright show-report
fi

# Cleanup: Stop dev server if we started it
if [ -n "$DEV_SERVER_PID" ]; then
    echo_info "Stopping dev server..."
    kill $DEV_SERVER_PID 2>/dev/null || true
fi

echo ""
echo_info "Test results saved to: playwright-report/"
echo_info "To view report: npm run test:e2e:report"
echo ""

exit $TEST_EXIT_CODE
