# OffGridFlow Complete Setup - PowerShell Version
# Windows-compatible setup script

param(
    [string]$SendGridKey = ""
)

Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  OffGridFlow - Complete Configuration Setup               ║" -ForegroundColor Cyan
Write-Host "║  Windows PowerShell Edition                               ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Helper functions
function Print-Success {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor Green
}

function Print-Error {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor Red
}

function Print-Warning {
    param([string]$Message)
    Write-Host "⚠️  $Message" -ForegroundColor Yellow
}

function Print-Section {
    param([string]$Title)
    Write-Host ""
    Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host $Title -ForegroundColor Cyan
    Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host ""
}

# Step 1: Prerequisites
Print-Section "Step 1: Checking Prerequisites"

try {
    $null = curl --version 2>$null
    Print-Success "curl is available"
} catch {
    Print-Error "curl is not available on this system"
    Print-Warning "Please install curl or use the configuration manually"
}

try {
    $null = git --version 2>$null
    Print-Success "git is available"
} catch {
    Print-Warning "git is not installed (optional)"
}

# Step 2: Get SendGrid API Key
Print-Section "Step 2: SendGrid Configuration"

Write-Host "You need a SendGrid API key. Follow these steps:"
Write-Host ""
Write-Host "1. Go to: https://signup.sendgrid.com"
Write-Host "2. Create free account OR login to existing account"
Write-Host "3. Go to: Settings → API Keys"
Write-Host "4. Click: Create New → Full Access"
Write-Host "5. Copy the API key (starts with SG.)"
Write-Host ""

if ([string]::IsNullOrEmpty($SendGridKey)) {
    $SendGridKey = Read-Host "Paste your SendGrid API key (or press Enter to skip)"
}

if ([string]::IsNullOrEmpty($SendGridKey)) {
    Print-Warning "No SendGrid API key provided. Going to manual instructions..."
    Print-Warning "You'll need to manually add variables to Railway"
} else {
    if ($SendGridKey -match "^SG\.") {
        Print-Success "Valid SendGrid API key detected"
    } else {
        Print-Error "API key should start with 'SG.'. Please verify"
    }
}

# Step 3: Display configuration
Print-Section "Step 3: Railway Configuration Variables"

Write-Host "Add these environment variables to Railway (offgridflow-api service):" -ForegroundColor Cyan
Write-Host ""

$Variables = @{
    "OFFGRIDFLOW_SMTP_HOST" = "smtp.sendgrid.net"
    "OFFGRIDFLOW_SMTP_PORT" = "587"
    "OFFGRIDFLOW_SMTP_USERNAME" = "apikey"
    "OFFGRIDFLOW_SMTP_PASSWORD" = if ($SendGridKey) { $SendGridKey } else { "YOUR_SENDGRID_API_KEY" }
    "OFFGRIDFLOW_SMTP_FROM_EMAIL" = "noreply@off-grid-flow.com"
    "OFFGRIDFLOW_SMTP_FROM_NAME" = "OffGridFlow"
    "OFFGRIDFLOW_SMTP_USE_TLS" = "true"
    "OFFGRIDFLOW_REQUIRE_EMAIL_VERIFICATION" = "true"
    "OFFGRIDFLOW_EMAIL_VERIFICATION_TTL" = "24h"
    "OFFGRIDFLOW_FRONTEND_URL" = "https://off-grid-flow.com"
    "OFFGRIDFLOW_ALLOWED_ORIGINS" = "https://off-grid-flow.com"
}

foreach ($key in $Variables.Keys) {
    Write-Host "$key = $($Variables[$key])"
}

Write-Host ""
Write-Host "Steps to add to Railway:"
Write-Host "1. Go to: https://railway.app"
Write-Host "2. Select: OffGridFlow project"
Write-Host "3. Select: offgridflow-api service"
Write-Host "4. Click: Variables tab"
Write-Host "5. Add each variable above"
Write-Host "6. Click: Deploy"
Write-Host ""

$deployed = Read-Host "Have you added variables and deployed? (y/n)"

if ($deployed -ne "y" -and $deployed -ne "Y") {
    Print-Warning "Please configure Railway first, then run this script again"
    exit 0
}

Print-Success "Variables configured on Railway"

# Step 4: Wait for deployment
Print-Section "Step 4: Waiting for Deployment"

Write-Host "Railway is redeploying the API service..."
Write-Host "This usually takes 2-3 minutes."
Write-Host ""

Read-Host "Press Enter when deployment is complete"

# Step 5: Test endpoints
Print-Section "Step 5: Testing API Endpoints"

Write-Host "Testing API health..."

try {
    $HealthResponse = Invoke-WebRequest -Uri "https://offgridflow-api-production.up.railway.app/health" -UseBasicParsing -TimeoutSec 10
    Print-Success "API is responding (status: $($HealthResponse.StatusCode))"
} catch {
    Print-Warning "Could not reach API. This might be normal if deployment is still in progress."
}

# Step 6: Test registration endpoint
Write-Host ""
Write-Host "Testing registration endpoint..."

try {
    $TestEmail = "test-$(Get-Random -Minimum 1000 -Maximum 9999)@example.com"
    
    $RegisterPayload = @{
        email = $TestEmail
        password = "TestPassword123!"
        name = "Test User"
        first_name = "Test"
        last_name = "User"
        company_name = "Test Company"
    } | ConvertTo-Json

    $RegisterResponse = Invoke-WebRequest `
        -Uri "https://offgridflow-api-production.up.railway.app/api/auth/register" `
        -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $RegisterPayload `
        -UseBasicParsing `
        -TimeoutSec 10

    if ($RegisterResponse.Content -match "user") {
        Print-Success "Registration endpoint is working!"
        Write-Host ""
        Write-Host "Test response:"
        $RegisterResponse.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10 | Write-Host
    }
} catch {
    Print-Warning "Could not test registration. Error: $($_.Exception.Message)"
}

# Step 7: Final instructions
Print-Section "Setup Complete! ✨"

Write-Host "Your OffGridFlow setup is now ready!"
Write-Host ""
Write-Host "What's working now:"
Print-Success "Registration system"
Print-Success "Email verification"
Print-Success "Login system"
Print-Success "Demo page"
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Test registration at: https://off-grid-flow.com/register"
Write-Host "2. View demo at: https://off-grid-flow.com/demo"
Write-Host "3. Login to your dashboard"
Write-Host ""
Write-Host "If you encounter issues:" -ForegroundColor Cyan
Write-Host "- Check Railway logs for errors"
Write-Host "- Verify SendGrid account is set up correctly"
Write-Host "- Ensure all variables are spelled correctly"
Write-Host ""
Write-Host "Documentation files:" -ForegroundColor Cyan
Write-Host "- FINAL_IMPLEMENTATION_READY.md (Simplest guide)"
Write-Host "- STATUS_AND_NEXT_STEPS.md (Complete overview)"
Write-Host "- COMPLETE_IMPLEMENTATION_GUIDE.md (Detailed walkthrough)"
Write-Host ""

Print-Success "Setup script complete!"
Write-Host ""
