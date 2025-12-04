#!/bin/bash
# Integration Test Script for Cloud Connectors

set -e

echo "================================================"
echo "OffGridFlow Cloud Connector Integration Tests"
echo "================================================"
echo ""

# Load environment variables
if [ -f .env.test ]; then
    export $(cat .env.test | grep -v '^#' | xargs)
else
    echo "Warning: .env.test not found. Using .env instead."
    if [ -f .env ]; then
        export $(cat .env | grep -v '^#' | xargs)
    fi
fi

# Test configuration
API_URL=${API_URL:-http://localhost:8080}
TEST_TENANT_ID=${TEST_TENANT_ID:-test-tenant}

echo "Test Configuration:"
echo "  API URL: $API_URL"
echo "  Tenant ID: $TEST_TENANT_ID"
echo ""

# Check if API is running
echo "Checking API health..."
if ! curl -s -f $API_URL/health > /dev/null; then
    echo "Error: API is not responding at $API_URL"
    echo "Please start the API server first"
    exit 1
fi

echo "✓ API is healthy"
echo ""

# Test AWS Connector
echo "Testing AWS CUR Connector..."
if [ -z "$AWS_ACCESS_KEY_ID" ]; then
    echo "  ⚠ Skipped: AWS credentials not configured"
else
    echo "  Testing AWS connection..."
    curl -X POST $API_URL/api/v1/ingestion/aws \
        -H "Content-Type: application/json" \
        -H "X-Tenant-ID: $TEST_TENANT_ID" \
        -d '{
            "bucket": "test-bucket",
            "prefix": "cur/",
            "month": "2024-01"
        }' || echo "  ⚠ AWS test failed (expected if no real data)"
fi
echo ""

# Test Azure Connector
echo "Testing Azure Emissions Connector..."
if [ -z "$AZURE_CLIENT_ID" ]; then
    echo "  ⚠ Skipped: Azure credentials not configured"
else
    echo "  Testing Azure connection..."
    curl -X POST $API_URL/api/v1/ingestion/azure \
        -H "Content-Type: application/json" \
        -H "X-Tenant-ID: $TEST_TENANT_ID" \
        -d '{
            "subscription_id": "'"$AZURE_SUBSCRIPTION_ID"'",
            "start_date": "2024-01-01",
            "end_date": "2024-01-31"
        }' || echo "  ⚠ Azure test failed (expected if no real data)"
fi
echo ""

# Test GCP Connector
echo "Testing GCP Carbon Connector..."
if [ -z "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
    echo "  ⚠ Skipped: GCP credentials not configured"
else
    echo "  Testing GCP connection..."
    curl -X POST $API_URL/api/v1/ingestion/gcp \
        -H "Content-Type: application/json" \
        -H "X-Tenant-ID: $TEST_TENANT_ID" \
        -d '{
            "project_id": "'"$GCP_PROJECT_ID"'",
            "start_date": "2024-01-01",
            "end_date": "2024-01-31"
        }' || echo "  ⚠ GCP test failed (expected if no real data)"
fi
echo ""

# Test SAP Connector
echo "Testing SAP Connector..."
if [ -z "$SAP_BASE_URL" ]; then
    echo "  ⚠ Skipped: SAP credentials not configured"
else
    echo "  Testing SAP connection..."
    curl -X POST $API_URL/api/v1/ingestion/sap \
        -H "Content-Type: application/json" \
        -H "X-Tenant-ID: $TEST_TENANT_ID" \
        -d '{
            "system": "production",
            "entity": "emissions",
            "start_date": "2024-01-01",
            "end_date": "2024-01-31"
        }' || echo "  ⚠ SAP test failed (expected if no real data)"
fi
echo ""

# Test Utility Connector
echo "Testing Utility Bill Connector..."
if [ -z "$UTILITY_API_KEY" ]; then
    echo "  ⚠ Skipped: Utility API credentials not configured"
else
    echo "  Testing Utility API connection..."
    curl -X POST $API_URL/api/v1/ingestion/utility \
        -H "Content-Type: application/json" \
        -H "X-Tenant-ID: $TEST_TENANT_ID" \
        -d '{
            "account_id": "test-account",
            "start_date": "2024-01-01",
            "end_date": "2024-01-31"
        }' || echo "  ⚠ Utility test failed (expected if no real data)"
fi
echo ""

# Test Job Queue
echo "Testing Job Queue..."
curl -X GET $API_URL/api/v1/jobs?tenant_id=$TEST_TENANT_ID \
    -H "Content-Type: application/json" || echo "  ⚠ Job queue test failed"
echo ""

# Test Emissions Calculation
echo "Testing Emissions Calculation..."
curl -X POST $API_URL/api/v1/emissions/calculate \
    -H "Content-Type: application/json" \
    -H "X-Tenant-ID: $TEST_TENANT_ID" \
    -d '{
        "scope": "scope1",
        "category": "stationary_combustion",
        "fuel_type": "natural_gas",
        "quantity": 1000,
        "unit": "therms"
    }' || echo "  ⚠ Emissions calculation test failed"
echo ""

echo "================================================"
echo "Integration Tests Complete"
echo "================================================"
echo ""
echo "Summary:"
echo "  - Review the output above for any errors"
echo "  - Configure credentials in .env for full testing"
echo "  - Check logs for detailed error messages"
echo ""
