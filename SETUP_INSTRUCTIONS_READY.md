# Complete Setup Instructions - Ready to Execute

## STEP 1: CREATE SENDGRID ACCOUNT & GET API KEY

### If you DON'T have SendGrid yet:
```
1. Go to: https://signup.sendgrid.com
2. Sign up with your email
3. Verify email address  
4. Complete the quick setup
5. Go to Settings → API Keys
6. Create New → Full Access API Key
7. Copy the key (format: SG.xxxxxxxxxxxxx)
8. SAVE IT SAFELY
```

### Sender Email Setup:
If you OWN the domain `off-grid-flow.com`:
- Go to Sender Authentication
- Add Domain Authentication
- Follow SendGrid's DNS setup
- Use: noreply@off-grid-flow.com

If you DON'T own the domain yet:
- Use your personal email as sender temporarily
- Update later when domain is ready

---

## STEP 2: PROVIDED - Railway Environment Variables Configuration

Once you have SendGrid API key, these are the EXACT variables to set on Railway:

```
# SMTP Configuration for SendGrid
OFFGRIDFLOW_SMTP_HOST=smtp.sendgrid.net
OFFGRIDFLOW_SMTP_PORT=587
OFFGRIDFLOW_SMTP_USERNAME=apikey
OFFGRIDFLOW_SMTP_PASSWORD=SG.YOUR_SENDGRID_API_KEY_HERE
OFFGRIDFLOW_SMTP_FROM_EMAIL=noreply@off-grid-flow.com
OFFGRIDFLOW_SMTP_FROM_NAME=OffGridFlow
OFFGRIDFLOW_SMTP_USE_TLS=true

# Email Verification Settings
OFFGRIDFLOW_REQUIRE_EMAIL_VERIFICATION=true
OFFGRIDFLOW_EMAIL_VERIFICATION_TTL=24h

# Frontend Configuration
OFFGRIDFLOW_FRONTEND_URL=https://off-grid-flow.com
OFFGRIDFLOW_ALLOWED_ORIGINS=https://off-grid-flow.com
```

### How to set on Railway:
1. Go to Railway Dashboard
2. Select OffGridFlow Project
3. Select offgridflow-api service
4. Go to "Variables" tab
5. Add each variable above
6. Click "Deploy" when done

---

## STEP 3: WHAT I'M IMPLEMENTING FOR YOU

### A. Email Verification System
- ✅ Already built into the codebase
- ✅ Just needs SMTP configuration (done above)
- ✅ Automatically sends verification emails
- ✅ Email templates are ready to go
- ✅ 24-hour expiration on links

### B. Complete Demo Page  
I will implement:
- Live demo page at `/demo`
- Pre-loaded sample ESG data
- Interactive dashboard preview
- Feature showcase with real data
- "Try It Now" button to register
- No login required to view demo

### C. Demo Tenant & User
- Pre-created demo organization
- Sample emissions activities
- Sample compliance reports
- Demo user credentials for testing

---

## STEP 4: COMPLETE FLOW AFTER SETUP

```
User visits: https://off-grid-flow.com/register
         ↓
User fills form with email
         ↓
POST /api/auth/register
         ↓
API creates user & tenant
         ↓
API sends verification email via SendGrid ✅
         ↓
User gets email with verification link
         ↓
User clicks link
         ↓
Email verified
         ↓
User can login
         ↓
Sees dashboard with onboarding
```

And for demo:
```
User visits: https://off-grid-flow.com/demo
         ↓
Sees interactive demo (NO LOGIN NEEDED)
         ↓
Explores dashboard, reports, features
         ↓
Clicks "Try It Now" → Goes to register
```

---

## WHAT HAPPENS NEXT

1. You provide SendGrid API key
2. I implement demo page with sample data
3. You set Railway environment variables
4. Railway rebuilds and deploys
5. Registration works end-to-end
6. Demo page is live
7. Everything fully tested

---

## DO YOU HAVE SENDGRID API KEY?

**Option 1**: You already have one
- Just provide the key: `SG.xxxxx`

**Option 2**: You need to create account
- Go to https://signup.sendgrid.com
- Takes 3-5 minutes
- Come back with the API key

**Option 3**: Use different email service
- Resend.com (modern alternative)
- AWS SES (if you use AWS)
- Mailgun
- I can implement whichever you prefer

---

## READY?

Let me know:
1. ✅ SendGrid API Key (or will create account)
2. ✅ Email address to verify with
3. ✅ Confirmation to proceed with demo implementation

Once you confirm, I'll:
- Build complete demo page
- Configure all settings
- Test everything
- Get you fully working
