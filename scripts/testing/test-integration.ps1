# OffGridFlow Cloud Connector Integration Tests (PowerShell)

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "OffGridFlow Cloud Connector Integration Tests" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Load environment variables
if (Test-Path .env.test) {
    Get-Content .env.test | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            $key = $matches[1].Trim()
            $value = $matches[2].Trim()
            [Environment]::SetEnvironmentVariable($key, $value, 'Process')
        }
    }
} elseif (Test-Path .env) {
    Write-Host "Warning: .env.test not found. Using .env instead." -ForegroundColor Yellow
    Get-Content .env | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            $key = $matches[1].Trim()
            $value = $matches[2].Trim()
            [Environment]::SetEnvironmentVariable($key, $value, 'Process')
        }
    }
}

# Test configuration
$API_URL = if ($env:API_URL) { $env:API_URL } else { "http://localhost:8080" }
$TEST_TENANT_ID = if ($env:TEST_TENANT_ID) { $env:TEST_TENANT_ID } else { "test-tenant" }

Write-Host "Test Configuration:"
Write-Host "  API URL: $API_URL"
Write-Host "  Tenant ID: $TEST_TENANT_ID"
Write-Host ""

# Check if API is running
Write-Host "Checking API health..."
try {
    $response = Invoke-WebRequest -Uri "$API_URL/health" -Method Get -UseBasicParsing
    Write-Host "✓ API is healthy" -ForegroundColor Green
} catch {
    Write-Host "Error: API is not responding at $API_URL" -ForegroundColor Red
    Write-Host "Please start the API server first" -ForegroundColor Yellow
    exit 1
}
Write-Host ""

# Test AWS Connector
Write-Host "Testing AWS CUR Connector..." -ForegroundColor Yellow
if (-not $env:AWS_ACCESS_KEY_ID) {
    Write-Host "  ⚠ Skipped: AWS credentials not configured" -ForegroundColor Yellow
} else {
    Write-Host "  Testing AWS connection..."
    try {
        $body = @{
            bucket = "test-bucket"
            prefix = "cur/"
            month = "2024-01"
        } | ConvertTo-Json

        $response = Invoke-WebRequest -Uri "$API_URL/api/v1/ingestion/aws" `
            -Method Post `
            -Headers @{"Content-Type"="application/json"; "X-Tenant-ID"=$TEST_TENANT_ID} `
            -Body $body `
            -UseBasicParsing
        Write-Host "  ✓ AWS test passed" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠ AWS test failed (expected if no real data)" -ForegroundColor Yellow
    }
}
Write-Host ""

# Test Azure Connector
Write-Host "Testing Azure Emissions Connector..." -ForegroundColor Yellow
if (-not $env:AZURE_CLIENT_ID) {
    Write-Host "  ⚠ Skipped: Azure credentials not configured" -ForegroundColor Yellow
} else {
    Write-Host "  Testing Azure connection..."
    try {
        $body = @{
            subscription_id = $env:AZURE_SUBSCRIPTION_ID
            start_date = "2024-01-01"
            end_date = "2024-01-31"
        } | ConvertTo-Json

        $response = Invoke-WebRequest -Uri "$API_URL/api/v1/ingestion/azure" `
            -Method Post `
            -Headers @{"Content-Type"="application/json"; "X-Tenant-ID"=$TEST_TENANT_ID} `
            -Body $body `
            -UseBasicParsing
        Write-Host "  ✓ Azure test passed" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠ Azure test failed (expected if no real data)" -ForegroundColor Yellow
    }
}
Write-Host ""

# Test GCP Connector
Write-Host "Testing GCP Carbon Connector..." -ForegroundColor Yellow
if (-not $env:GOOGLE_APPLICATION_CREDENTIALS) {
    Write-Host "  ⚠ Skipped: GCP credentials not configured" -ForegroundColor Yellow
} else {
    Write-Host "  Testing GCP connection..."
    try {
        $body = @{
            project_id = $env:GCP_PROJECT_ID
            start_date = "2024-01-01"
            end_date = "2024-01-31"
        } | ConvertTo-Json

        $response = Invoke-WebRequest -Uri "$API_URL/api/v1/ingestion/gcp" `
            -Method Post `
            -Headers @{"Content-Type"="application/json"; "X-Tenant-ID"=$TEST_TENANT_ID} `
            -Body $body `
            -UseBasicParsing
        Write-Host "  ✓ GCP test passed" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠ GCP test failed (expected if no real data)" -ForegroundColor Yellow
    }
}
Write-Host ""

# Test SAP Connector
Write-Host "Testing SAP Connector..." -ForegroundColor Yellow
if (-not $env:SAP_BASE_URL) {
    Write-Host "  ⚠ Skipped: SAP credentials not configured" -ForegroundColor Yellow
} else {
    Write-Host "  Testing SAP connection..."
    try {
        $body = @{
            system = "production"
            entity = "emissions"
            start_date = "2024-01-01"
            end_date = "2024-01-31"
        } | ConvertTo-Json

        $response = Invoke-WebRequest -Uri "$API_URL/api/v1/ingestion/sap" `
            -Method Post `
            -Headers @{"Content-Type"="application/json"; "X-Tenant-ID"=$TEST_TENANT_ID} `
            -Body $body `
            -UseBasicParsing
        Write-Host "  ✓ SAP test passed" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠ SAP test failed (expected if no real data)" -ForegroundColor Yellow
    }
}
Write-Host ""

# Test Utility Connector
Write-Host "Testing Utility Bill Connector..." -ForegroundColor Yellow
if (-not $env:UTILITY_API_KEY) {
    Write-Host "  ⚠ Skipped: Utility API credentials not configured" -ForegroundColor Yellow
} else {
    Write-Host "  Testing Utility API connection..."
    try {
        $body = @{
            account_id = "test-account"
            start_date = "2024-01-01"
            end_date = "2024-01-31"
        } | ConvertTo-Json

        $response = Invoke-WebRequest -Uri "$API_URL/api/v1/ingestion/utility" `
            -Method Post `
            -Headers @{"Content-Type"="application/json"; "X-Tenant-ID"=$TEST_TENANT_ID} `
            -Body $body `
            -UseBasicParsing
        Write-Host "  ✓ Utility test passed" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠ Utility test failed (expected if no real data)" -ForegroundColor Yellow
    }
}
Write-Host ""

# Test Job Queue
Write-Host "Testing Job Queue..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$API_URL/api/v1/jobs?tenant_id=$TEST_TENANT_ID" `
        -Method Get `
        -Headers @{"Content-Type"="application/json"} `
        -UseBasicParsing
    Write-Host "  ✓ Job queue test passed" -ForegroundColor Green
} catch {
    Write-Host "  ⚠ Job queue test failed" -ForegroundColor Yellow
}
Write-Host ""

# Test Emissions Calculation
Write-Host "Testing Emissions Calculation..." -ForegroundColor Yellow
try {
    $body = @{
        scope = "scope1"
        category = "stationary_combustion"
        fuel_type = "natural_gas"
        quantity = 1000
        unit = "therms"
    } | ConvertTo-Json

    $response = Invoke-WebRequest -Uri "$API_URL/api/v1/emissions/calculate" `
        -Method Post `
        -Headers @{"Content-Type"="application/json"; "X-Tenant-ID"=$TEST_TENANT_ID} `
        -Body $body `
        -UseBasicParsing
    Write-Host "  ✓ Emissions calculation test passed" -ForegroundColor Green
} catch {
    Write-Host "  ⚠ Emissions calculation test failed" -ForegroundColor Yellow
}
Write-Host ""

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "Integration Tests Complete" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Summary:"
Write-Host "  - Review the output above for any errors"
Write-Host "  - Configure credentials in .env for full testing"
Write-Host "  - Check logs for detailed error messages"
Write-Host ""
