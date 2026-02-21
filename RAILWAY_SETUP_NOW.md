# OffGridFlow — Railway Environment Setup
**Do this once. Platform runs itself after.**

---

## STEP 1 — Generate Secrets (do this first, takes 2 min)

### JWT Secret (64 chars)
Go to: https://generate-secret.vercel.app/64
Copy the result. Paste as `OFFGRIDFLOW_JWT_SECRET` below.

### NextAuth Secret (32 chars)
Go to: https://generate-secret.vercel.app/32
Copy the result. Paste as `NEXTAUTH_SECRET` below.

---

## STEP 2 — SendGrid (email verification, takes 5 min)

1. Go to: https://signup.sendgrid.com
2. Create free account → verify your email
3. Settings → API Keys → Create API Key → Full Access
4. Copy key (starts with `SG.`)
5. Paste as `OFFGRIDFLOW_SMTP_PASSWORD` below

---

## STEP 3 — Stripe (subscriptions + payments, takes 10 min)

1. Go to: https://dashboard.stripe.com/register
2. Create account → activate (add business details)
3. **Create Products** (Catalog → Products → Add Product):

   | Product Name | Price      | Billing |
   |-------------|------------|---------|
   | Basic       | $299/month | Monthly |
   | Pro         | $999/month | Monthly |
   | Enterprise  | $2999/month| Monthly |

4. After creating each product, click the price to get its ID (format: `price_xxxxxxxxx`)
5. Go to Developers → API Keys → copy **Secret key** (starts with `sk_live_`)
6. Go to Developers → Webhooks → Add endpoint:
   - URL: `https://offgridflow-api-production.up.railway.app/api/billing/webhook`
   - Events: `checkout.session.completed`, `customer.subscription.updated`, `customer.subscription.deleted`, `invoice.payment_failed`
   - Copy the **Signing secret** (starts with `whsec_`)

---

## STEP 4 — Add Variables to Railway

### Service: `offgridflow-api`
Go to Railway → your project → `offgridflow-api` service → Variables tab

Add ALL of these:

```
# ── CRITICAL: Generate these ──────────────────────────────────────
OFFGRIDFLOW_JWT_SECRET           = [your 64-char random string from Step 1]

# ── CRITICAL: SendGrid ────────────────────────────────────────────
OFFGRIDFLOW_SMTP_HOST            = smtp.sendgrid.net
OFFGRIDFLOW_SMTP_PORT            = 587
OFFGRIDFLOW_SMTP_USERNAME        = apikey
OFFGRIDFLOW_SMTP_PASSWORD        = [your SG.xxxxx key from Step 2]
OFFGRIDFLOW_SMTP_FROM_EMAIL      = noreply@off-grid-flow.com
OFFGRIDFLOW_SMTP_FROM_NAME       = OffGridFlow
OFFGRIDFLOW_SMTP_USE_TLS         = true

# ── CRITICAL: Stripe ──────────────────────────────────────────────
STRIPE_SECRET_KEY                = [your sk_live_xxxxx from Step 3]
STRIPE_WEBHOOK_SECRET            = [your whsec_xxxxx from Step 3]
STRIPE_PRICE_FREE                = (leave blank or omit)
STRIPE_PRICE_BASIC               = [price_xxxxx for Basic from Step 3]
STRIPE_PRICE_PRO                 = [price_xxxxx for Pro from Step 3]
STRIPE_PRICE_ENTERPRISE          = [price_xxxxx for Enterprise from Step 3]

# ── Already configured (verify these exist) ───────────────────────
OFFGRIDFLOW_APP_ENV              = production
OFFGRIDFLOW_HTTP_PORT            = 8090
PORT                             = 8090
OFFGRIDFLOW_DB_DSN               = ${{Postgres.DATABASE_URL}}
DATABASE_URL                     = ${{Postgres.DATABASE_URL}}
OFFGRIDFLOW_FRONTEND_URL         = https://off-grid-flow.com
OFFGRIDFLOW_ALLOWED_ORIGINS      = https://off-grid-flow.com
OFFGRIDFLOW_REQUIRE_EMAIL_VERIFICATION = true
OFFGRIDFLOW_EMAIL_VERIFICATION_TTL     = 24h
OFFGRIDFLOW_COOKIE_SECURE        = true
OFFGRIDFLOW_COOKIE_DOMAIN        = .off-grid-flow.com
OFFGRIDFLOW_REQUIRE_AUTH         = true
OFFGRIDFLOW_ENABLE_AUDIT_LOG     = true
OFFGRIDFLOW_ENABLE_METRICS       = true
OFFGRIDFLOW_ENABLE_GRAPHQL       = true
```

---

### Service: `offgridflow-web`
Go to Railway → `offgridflow-web` service → Variables tab

Add these:

```
NEXTAUTH_SECRET                  = [your 32-char random string from Step 1]
NEXTAUTH_URL                     = https://off-grid-flow.com
NEXT_PUBLIC_OFFGRIDFLOW_API_URL  = https://offgridflow-api-production.up.railway.app
NEXT_PUBLIC_API_URL              = https://offgridflow-api-production.up.railway.app
NODE_ENV                         = production
PORT                             = 3000
```

---

## STEP 5 — Deploy

Click **Deploy** on both services. Wait ~3 minutes.

### Verify it works:
1. Go to `https://off-grid-flow.com/register`
2. Fill out the form → submit
3. Check your email for verification link
4. Click link → login
5. Go to Settings → Billing → should show pricing plans
6. Go to Settings → Data Sources → should show AWS/Azure/GCP/SAP/Utility connectors

---

## What Happens Automatically After This

| Event | What the platform does |
|-------|----------------------|
| Company registers | Account created, verification email sent |
| Email verified | Login unlocked |
| Company subscribes | Stripe checkout → payment → subscription activated |
| Company connects AWS/Azure/GCP/SAP | Credentials saved, emissions pulled every 30 min |
| Invoice paid | Email receipt sent automatically |
| Payment fails | Email alert sent automatically |
| Subscription cancelled | Downgraded to free tier automatically |

**You never touch any of this. It runs itself.**

---

## Your LinkedIn Outreach Flow
`LinkedIn message` → `off-grid-flow.com` → `Register` → `Billing` → `Data Sources` → **done**

No human intervention needed at any step.
