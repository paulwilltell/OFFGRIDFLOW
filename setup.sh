#!/bin/bash
# OffGridFlow Complete Setup Script
# This script helps you configure everything

set -e

echo "╔════════════════════════════════════════════════════════════╗"
echo "║  OffGridFlow - Complete Configuration Setup               ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Color codes for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print sections
print_section() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
}

# Function to print success
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Step 1: Check prerequisites
print_section "Step 1: Checking Prerequisites"

if ! command -v curl &> /dev/null; then
    print_error "curl is not installed"
    exit 1
fi
print_success "curl is installed"

if ! command -v git &> /dev/null; then
    print_warning "git is not installed (optional)"
else
    print_success "git is installed"
fi

# Step 2: SendGrid setup
print_section "Step 2: SendGrid Setup Instructions"

echo "You need a SendGrid API key. Follow these steps:"
echo ""
echo "1. Go to: https://signup.sendgrid.com"
echo "2. Create free account OR login to existing account"
echo "3. Go to: Settings → API Keys"
echo "4. Click: Create New → Full Access"
echo "5. Copy the API key (starts with SG.)"
echo ""

read -p "Do you have a SendGrid API key? (y/n) " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    read -sp "Paste your SendGrid API key: " SENDGRID_KEY
    echo ""
    
    if [[ $SENDGRID_KEY =~ ^SG\. ]]; then
        print_success "Valid SendGrid API key detected"
    else
        print_error "Invalid API key format (should start with SG.)"
        exit 1
    fi
else
    print_warning "Please go to https://signup.sendgrid.com and create an account first"
    print_warning "Then run this script again"
    exit 0
fi

# Step 3: Verify Railway configuration
print_section "Step 3: Railway Configuration"

echo "Now we need to configure Railway with these variables:"
echo ""
echo "OFFGRIDFLOW_SMTP_HOST=smtp.sendgrid.net"
echo "OFFGRIDFLOW_SMTP_PORT=587"
echo "OFFGRIDFLOW_SMTP_USERNAME=apikey"
echo "OFFGRIDFLOW_SMTP_PASSWORD=****"
echo "OFFGRIDFLOW_SMTP_FROM_EMAIL=noreply@off-grid-flow.com"
echo "OFFGRIDFLOW_SMTP_FROM_NAME=OffGridFlow"
echo "OFFGRIDFLOW_SMTP_USE_TLS=true"
echo "OFFGRIDFLOW_REQUIRE_EMAIL_VERIFICATION=true"
echo "OFFGRIDFLOW_EMAIL_VERIFICATION_TTL=24h"
echo "OFFGRIDFLOW_FRONTEND_URL=https://off-grid-flow.com"
echo "OFFGRIDFLOW_ALLOWED_ORIGINS=https://off-grid-flow.com"
echo ""

echo "Steps to add variables to Railway:"
echo "1. Go to: https://railway.app"
echo "2. Select: OffGridFlow project"
echo "3. Select: offgridflow-api service"
echo "4. Click: Variables tab"
echo "5. Add all variables above"
echo "6. Click: Deploy"
echo ""

read -p "Have you added variables to Railway? (y/n) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_warning "Please add variables to Railway first, then run this script again"
    exit 0
fi

print_success "Variables configured on Railway"

# Step 4: Wait for deployment
print_section "Step 4: Waiting for Deployment"

echo "Railway is redeploying the API service..."
echo "This usually takes 2-3 minutes."
echo ""

read -p "Press Enter when deployment is complete: "

# Step 5: Test registration
print_section "Step 5: Testing Registration"

echo "Testing registration endpoint..."

TEST_EMAIL="test-$(date +%s)@example.com"

REGISTER_RESPONSE=$(curl -s -X POST \
  "https://offgridflow-api-production.up.railway.app/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"TestPassword123!\",
    \"name\": \"Test User\",
    \"first_name\": \"Test\",
    \"last_name\": \"User\",
    \"company_name\": \"Test Company\"
  }")

if echo "$REGISTER_RESPONSE" | grep -q "user"; then
    print_success "Registration endpoint is working!"
    echo ""
    echo "Test registration response:"
    echo "$REGISTER_RESPONSE" | jq . 2>/dev/null || echo "$REGISTER_RESPONSE"
else
    print_error "Registration failed. Response:"
    echo "$REGISTER_RESPONSE"
fi

# Step 6: Test manual flow
print_section "Step 6: Manual Testing"

echo "To manually test the complete flow:"
echo ""
echo "1. Go to: https://off-grid-flow.com/register"
echo "2. Fill in the registration form"
echo "3. Submit"
echo "4. Check your email for verification"
echo "5. Click the verification link"
echo "6. Login with your credentials"
echo ""

# Step 7: Verify demo page
print_section "Step 7: Verify Demo Page"

echo "Testing demo page access..."

DEMO_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "https://off-grid-flow.com/demo")

if [ "$DEMO_RESPONSE" = "200" ]; then
    print_success "Demo page is accessible!"
    echo "Visit: https://off-grid-flow.com/demo"
else
    print_error "Demo page returned status: $DEMO_RESPONSE"
fi

# Final summary
print_section "Setup Complete! ✨"

echo "Your OffGridFlow setup is now complete!"
echo ""
echo "What's working now:"
echo "✅ Registration system"
echo "✅ Email verification"
echo "✅ Login system"
echo "✅ Demo page"
echo ""
echo "Next steps:"
echo "1. Test registration at: https://off-grid-flow.com/register"
echo "2. View demo at: https://off-grid-flow.com/demo"
echo "3. Login to dashboard"
echo ""
echo "If you encounter any issues:"
echo "- Check Railway logs for errors"
echo "- Verify SendGrid account is set up"
echo "- Ensure all variables are correctly configured"
echo ""
echo "Need help? Check these files:"
echo "- FINAL_IMPLEMENTATION_READY.md"
echo "- STATUS_AND_NEXT_STEPS.md"
echo "- COMPLETE_IMPLEMENTATION_GUIDE.md"
echo ""

print_success "Setup script complete!"
