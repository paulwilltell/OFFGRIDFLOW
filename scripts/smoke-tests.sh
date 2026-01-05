#!/bin/bash
# OffGridFlow Smoke Tests
# Validates critical functionality after deployment
# Usage: ./scripts/smoke-tests.sh [environment]
# Example: ./scripts/smoke-tests.sh production

set -euo pipefail

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
ENVIRONMENT=${1:-staging}
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
TEST_RESULTS_DIR="test-results"
RESULTS_FILE="${TEST_RESULTS_DIR}/smoke-test-${ENVIRONMENT}-${TIMESTAMP}.json"

# Environment URLs
case $ENVIRONMENT in
  production)
    BASE_URL="https://api.offgridflow.com"
    WEB_URL="https://app.offgridflow.com"
    ;;
  staging)
    BASE_URL="https://api.staging.offgridflow.com"
    WEB_URL="https://app.staging.offgridflow.com"
    ;;
  local)
    BASE_URL="http://localhost:8080"
    WEB_URL="http://localhost:3000"
    ;;
  *)
    echo -e "${RED}Unknown environment: ${ENVIRONMENT}${NC}"
    exit 1
    ;;
esac

# Test credentials
TEST_EMAIL="${SMOKE_TEST_EMAIL:-smoketest@offgridflow.com}"
TEST_PASSWORD="${SMOKE_TEST_PASSWORD:-SmokeTest123!}"

# Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Create results directory
mkdir -p "${TEST_RESULTS_DIR}"

# Initialize results file
cat > "${RESULTS_FILE}" <<EOF
{
  "environment": "${ENVIRONMENT}",
  "timestamp": "${TIMESTAMP}",
  "base_url": "${BASE_URL}",
  "tests": []
}
EOF

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_failure() {
    echo -e "${RED}[FAIL]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Test runner function
run_test() {
    local test_name=$1
    local test_command=$2
    local expected_status=${3:-0}
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log_info "Running: ${test_name}"
    
    local start_time=$(date +%s%3N)
    local output
    local status
    
    set +e
    output=$(eval "$test_command" 2>&1)
    status=$?
    set -e
    
    local end_time=$(date +%s%3N)
    local duration=$((end_time - start_time))
    
    if [ $status -eq $expected_status ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        log_success "${test_name} (${duration}ms)"
        
        # Add to results
        jq --arg name "$test_name" \
           --arg status "PASS" \
           --arg duration "$duration" \
           '.tests += [{"name": $name, "status": $status, "duration": $duration}]' \
           "${RESULTS_FILE}" > "${RESULTS_FILE}.tmp" && mv "${RESULTS_FILE}.tmp" "${RESULTS_FILE}"
        
        return 0
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        log_failure "${test_name}"
        log_failure "Expected status: ${expected_status}, Got: ${status}"
        log_failure "Output: ${output}"
        
        # Add to results
        jq --arg name "$test_name" \
           --arg status "FAIL" \
           --arg duration "$duration" \
           --arg error "$output" \
           '.tests += [{"name": $name, "status": $status, "duration": $duration, "error": $error}]' \
           "${RESULTS_FILE}" > "${RESULTS_FILE}.tmp" && mv "${RESULTS_FILE}.tmp" "${RESULTS_FILE}"
        
        return 1
    fi
}

# Banner
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  OffGridFlow Smoke Tests - ${ENVIRONMENT}${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "  Base URL: ${BASE_URL}"
echo -e "  Web URL:  ${WEB_URL}"
echo -e "  Time:     $(date)"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# ============================================================================
# Health & Infrastructure Tests
# ============================================================================
echo -e "${YELLOW}▶ Health & Infrastructure Tests${NC}"

run_test "API Health Check" \
    "curl -f -s ${BASE_URL}/health | jq -e '.status == \"healthy\"'"

run_test "API Liveness Probe" \
    "curl -f -s ${BASE_URL}/health/live | jq -e '.status == \"healthy\"'"

run_test "API Readiness Probe" \
    "curl -f -s ${BASE_URL}/health/ready | jq -e '.status == \"healthy\" or .status == \"degraded\"'"

run_test "Metrics Endpoint Available" \
    "curl -f -s ${BASE_URL}:8081/metrics | grep -q 'http_requests_total'"

run_test "API Response Time < 500ms" \
    "time=$(curl -o /dev/null -s -w '%{time_total}' ${BASE_URL}/health) && (( \$(echo \"\$time < 0.5\" | bc -l) ))"

# ============================================================================
# Authentication Tests
# ============================================================================
echo ""
echo -e "${YELLOW}▶ Authentication Tests${NC}"

run_test "User Registration" \
    "curl -f -s -X POST ${BASE_URL}/v1/auth/register \
        -H 'Content-Type: application/json' \
        -d '{\"email\":\"test-${TIMESTAMP}@example.com\",\"password\":\"Test123!\",\"name\":\"Test User\"}' \
        | jq -e '.user.id != null'"

# Login and get token
AUTH_TOKEN=$(curl -s -X POST "${BASE_URL}/v1/auth/login" \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"${TEST_EMAIL}\",\"password\":\"${TEST_PASSWORD}\"}" \
    | jq -r '.token')

if [ "$AUTH_TOKEN" != "null" ] && [ -n "$AUTH_TOKEN" ]; then
    log_success "User Login"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    log_failure "User Login - No token received"
    FAILED_TESTS=$((FAILED_TESTS + 1))
    exit 1
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

run_test "Token Validation" \
    "curl -f -s -H 'Authorization: Bearer ${AUTH_TOKEN}' ${BASE_URL}/v1/auth/me | jq -e '.user.email != null'"

# ============================================================================
# Core API Tests
# ============================================================================
echo ""
echo -e "${YELLOW}▶ Core API Functionality Tests${NC}"

run_test "List Activities" \
    "curl -f -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \
        ${BASE_URL}/v1/activities?limit=10 | jq -e '.activities != null'"

# Create activity
ACTIVITY_DATA=$(cat <<EOF
{
  "name": "Smoke Test Activity ${TIMESTAMP}",
  "activity_type": "electricity",
  "scope": 2,
  "quantity": 1000,
  "unit": "kWh",
  "activity_date": "$(date +%Y-%m-%d)"
}
EOF
)

ACTIVITY_ID=$(curl -s -X POST "${BASE_URL}/v1/activities" \
    -H "Authorization: Bearer ${AUTH_TOKEN}" \
    -H 'Content-Type: application/json' \
    -d "${ACTIVITY_DATA}" \
    | jq -r '.id')

if [ "$ACTIVITY_ID" != "null" ] && [ -n "$ACTIVITY_ID" ]; then
    log_success "Create Activity (ID: ${ACTIVITY_ID})"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    log_failure "Create Activity"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

run_test "Get Activity by ID" \
    "curl -f -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \
        ${BASE_URL}/v1/activities/${ACTIVITY_ID} | jq -e '.activity.id == \"${ACTIVITY_ID}\"'"

run_test "Update Activity" \
    "curl -f -s -X PUT ${BASE_URL}/v1/activities/${ACTIVITY_ID} \
        -H 'Authorization: Bearer ${AUTH_TOKEN}' \
        -H 'Content-Type: application/json' \
        -d '{\"quantity\": 1500}' | jq -e '.activity.quantity == 1500'"

run_test "Delete Activity" \
    "curl -f -s -X DELETE ${BASE_URL}/v1/activities/${ACTIVITY_ID} \
        -H 'Authorization: Bearer ${AUTH_TOKEN}' -w '%{http_code}' -o /dev/null | grep -q '204'"

# ============================================================================
# Emission Calculation Tests
# ============================================================================
echo ""
echo -e "${YELLOW}▶ Emission Calculation Tests${NC}"

run_test "Calculate Emissions" \
    "curl -f -s -X POST ${BASE_URL}/v1/emissions/calculate \
        -H 'Authorization: Bearer ${AUTH_TOKEN}' \
        -H 'Content-Type: application/json' \
        -d '{\"activity_type\":\"electricity\",\"quantity\":1000,\"unit\":\"kWh\",\"region\":\"US\"}' \
        | jq -e '.emissions_co2e != null'"

run_test "List Emission Factors" \
    "curl -f -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \
        ${BASE_URL}/v1/emission-factors?limit=10 | jq -e '.emission_factors != null'"

# ============================================================================
# Compliance & Reporting Tests
# ============================================================================
echo ""
echo -e "${YELLOW}▶ Compliance & Reporting Tests${NC}"

run_test "Get Compliance Status" \
    "curl -f -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \
        ${BASE_URL}/v1/compliance/status | jq -e '.frameworks != null'"

run_test "List Reports" \
    "curl -f -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \
        ${BASE_URL}/v1/reports?limit=10 | jq -e '.reports != null'"

# ============================================================================
# Dashboard & Analytics Tests
# ============================================================================
echo ""
echo -e "${YELLOW}▶ Dashboard & Analytics Tests${NC}"

run_test "Get Dashboard Data" \
    "curl -f -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \
        ${BASE_URL}/v1/dashboard | jq -e '.total_emissions != null'"

run_test "Get Emissions Summary" \
    "curl -f -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \
        ${BASE_URL}/v1/emissions/summary | jq -e '.scope1 != null and .scope2 != null and .scope3 != null'"

# ============================================================================
# Frontend Tests
# ============================================================================
echo ""
echo -e "${YELLOW}▶ Frontend Tests${NC}"

run_test "Web Application Accessible" \
    "curl -f -s ${WEB_URL} | grep -q 'OffGridFlow'"

run_test "Web App Response Time < 1s" \
    "time=$(curl -o /dev/null -s -w '%{time_total}' ${WEB_URL}) && (( \$(echo \"\$time < 1.0\" | bc -l) ))"

run_test "Static Assets Loading" \
    "curl -f -s ${WEB_URL}/_next/static/ -I | grep -q '200'"

# ============================================================================
# Security Tests
# ============================================================================
echo ""
echo -e "${YELLOW}▶ Security Tests${NC}"

run_test "Unauthorized Access Blocked" \
    "curl -s ${BASE_URL}/v1/activities -w '%{http_code}' -o /dev/null | grep -q '401'" 0

run_test "HTTPS Redirect (Production)" \
    "[ '${ENVIRONMENT}' != 'production' ] || curl -s -I http://api.offgridflow.com | grep -q 'HTTP/.*301'" 0

run_test "Security Headers Present" \
    "curl -s -I ${BASE_URL}/health | grep -q 'X-Content-Type-Options: nosniff'" 0

# ============================================================================
# Results Summary
# ============================================================================
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  Test Results Summary${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "  Total Tests:   ${TOTAL_TESTS}"
echo -e "  ${GREEN}Passed Tests:  ${PASSED_TESTS}${NC}"
echo -e "  ${RED}Failed Tests:  ${FAILED_TESTS}${NC}"
echo -e "  Success Rate:  $(awk "BEGIN {printf \"%.1f\", ($PASSED_TESTS/$TOTAL_TESTS)*100}")%"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "Results saved to: ${RESULTS_FILE}"
echo ""

# Update results file with summary
jq --arg total "$TOTAL_TESTS" \
   --arg passed "$PASSED_TESTS" \
   --arg failed "$FAILED_TESTS" \
   '. + {summary: {total: $total, passed: $passed, failed: $failed}}' \
   "${RESULTS_FILE}" > "${RESULTS_FILE}.tmp" && mv "${RESULTS_FILE}.tmp" "${RESULTS_FILE}"

# Exit with failure if any tests failed
if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "${RED}❌ Smoke tests FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}✅ All smoke tests PASSED${NC}"
    exit 0
fi
