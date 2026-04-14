# OffGridFlow Customer Journey Audit

Date: 2026-04-14
Auditor: Codex
Mode: Live-site customer simulation on `https://off-grid-flow.com`

## Verdict

Strict verdict: `DO NOT SELL`

Customer-readiness score: `22 / 100`

The product does not currently deliver the paid value proposition end-to-end for a new customer. Awareness and account creation work well enough to create trust, but the first meaningful product actions fail:

- the post-signup dashboard is broken
- CSV import is broken
- core compliance output crashes
- audit surfaces mask missing backend functionality behind empty states
- checkout never reaches Stripe

This is not a polish problem. It is a broken first-customer journey.

## Scope

I audited the site as a real customer would:

1. Land on the marketing site and evaluate clarity/trust
2. Start a free trial
3. Create an account
4. Enter the product
5. Try to upload emissions data
6. Try to review compliance and audit outputs
7. Try to reach billing and start checkout
8. Stop before any card entry or payment submission

I also checked browser console, network behavior, and corresponding source files to identify likely root causes.

## Journey Summary

### 1. Awareness / marketing

What worked:

- The homepage, pricing, methodology, trust, and evidence pages present a coherent story.
- Positioning is clear: carbon compliance for companies that do not want Big 4-style pricing.
- Primary CTA path is easy to understand.

What failed:

- Every public page emits repeated Google Ads conversion/CSP errors in production.
- This creates immediate console noise on first load and signals weak production hygiene.

Live evidence:

- Browser console repeatedly blocked requests to `googleads.g.doubleclick.net`, `google.com`, and `googleadservices.com` because CSP does not allow them.

Repo evidence:

- CSP allows only `https://www.googletagmanager.com` in `script-src` and only Stripe plus the Railway API in `connect-src`: `web/next.config.js:58`

### 2. Trial start / registration

What worked:

- `Start Free Trial` led to registration.
- Registration succeeded.
- Auto-login succeeded.

Live evidence:

- `POST /api/auth/register` returned `201`
- `POST /api/auth/login` returned `200`

Assessment:

- This is the strongest part of the journey. A customer can become a signed-in user.

### 3. First product experience

What failed:

- After signup, the main dashboard landed on `Connecting...` and did not provide a working first-use experience.
- The dashboard attempted to call a localhost backend from the live site.

Live evidence:

- Browser attempted `http://localhost:8090/api/v1/emissions?tenantId=default&timeframe=monthly`
- CSP blocked the request

Repo evidence:

- Localhost fallback is still present in the carbon API client: `web/lib/api/carbon.ts:16`

Impact:

- A new customer’s first view of the paid product is broken before they can learn anything from it.

### 4. Data import

What failed:

- The guided CSV import flow from the empty emissions state failed immediately with `Failed to fetch`.
- Metrics remained zero and no usable data appeared.

Live evidence:

- Browser attempted to upload to `https://offgridflow-api-production.up.railway.app/api/ingestion/upload/csv`
- Production CSP allows `offgridflow-api-v2-production.up.railway.app`, not the older `offgridflow-api-production.up.railway.app`

Repo evidence:

- Upload flow still falls back to the old API host: `web/app/(app)/emissions/page.tsx:526`
- Auth for the upload uses the browser token directly: `web/app/(app)/emissions/page.tsx:524`

Impact:

- This is the single biggest product failure.
- If customers cannot import data, they cannot produce emissions calculations, compliance reports, or audit artifacts.

### 5. Data sources / connectors

What failed:

- `Settings > Data Sources` crashed with a client-side exception.
- The page fetched connector endpoints successfully, but still rendered an application error.

Live evidence:

- `GET /api/connectors/list` returned `200`
- `GET /api/ingestion/logs?limit=10` returned `200`
- `GET /api/connectors/schedule` returned `200`
- Frontend crashed with `TypeError: Cannot read properties of null (reading 'length')`

Repo evidence:

- Connector/log data is assigned directly from API responses without normalization: `web/app/(app)/settings/data-sources/page.tsx:334`
- The page later renders `logs.length` directly: `web/app/(app)/settings/data-sources/page.tsx:561`

Impact:

- The cloud-connector path is not dependable for first-time onboarding either.

### 6. Compliance output

What failed:

- `Compliance > CSRD` crashed on load.
- The API returned data, but the page still failed before the customer could use it.

Live evidence:

- `GET /api/compliance/csrd?year=2026` returned `200`
- Frontend crashed with `TypeError: Cannot read properties of undefined (reading 'scope1Tons')`

Repo evidence:

- The page assumes `report.totals` exists and reads `report?.totals.scope1Tons`: `web/app/(app)/compliance/csrd/page.tsx:177`
- The safe version should guard `totals` itself, not just `report`

Impact:

- The product cannot honestly claim to deliver a high-quality compliance workflow if the flagship report crashes after data fetch succeeds.

### 7. Audit surfaces

What failed:

- `Audit > Calculation Ledger` loaded a clean empty state even though the backend returned `404`
- `Audit > Approvals` loaded a clean empty state even though the backend returned `404`
- Creating a new approval request returned `403`

Live evidence:

- `GET /api/audit/ledger` returned `404`
- `GET /api/audit/approvals` returned `404`
- `POST /api/audit/approvals` returned `403`

Repo evidence:

- Ledger swallows backend failures and silently renders an empty state: `web/app/(app)/audit/ledger/page.tsx:41`
- On error it falls back to `[]` rather than surfacing the backend issue: `web/app/(app)/audit/ledger/page.tsx:43`
- Approvals does the same: `web/app/(app)/audit/approvals/page.tsx:45`
- Create approval is wired, but currently rejected by the API: `web/app/(app)/audit/approvals/page.tsx:71`

Impact:

- This is deceptive UX.
- The interface looks complete while core audit APIs are not available.
- Customers and auditors would assume “no records yet” when the actual state is “feature unavailable.”

### 8. Billing / checkout

What failed:

- `Manage Subscription` is shown for a user with no subscription, but returns an error
- `Get Audit Prep` does not create a checkout session
- I could not reach Stripe at all, so I stopped before payment because there was no payment page to inspect

Live evidence:

- `POST /api/billing/portal` returned `403`
- `POST /api/billing/checkout` returned `403`
- The UI surfaced:
  - `Failed to open billing portal. Please try again.`
  - `Failed to start checkout. Please try again.`

Repo evidence:

- Billing page calls checkout and portal as expected, but the live backend rejects them: `web/app/(app)/settings/billing/page.tsx:92`, `web/app/(app)/settings/billing/page.tsx:108`

Impact:

- A customer cannot buy the product from inside the application.
- Stripe may be configured somewhere, but the actual purchase path is broken.

## Security And Session Findings

### 1. JWT stored in `localStorage`

Live evidence:

- Browser storage contained `offgridflow_access_token`
- No application auth cookie was present; only Google Ads cookie data was present

Repo evidence:

- Auth token is stored in `localStorage`: `web/app/login/page.tsx:78`
- API client reads the bearer token from `localStorage`: `web/lib/api.ts:46`

Why this matters:

- This increases XSS blast radius because a successful script injection can read and exfiltrate the token.
- It also contributes to the session split between client-side auth and server-side route protection.

### 2. Hard reload / deep-link auth is broken

Observed behavior:

- Every direct navigation to protected routes such as `/settings/billing`, `/compliance/csrd`, and `/audit/ledger` forced a redirect back to `/login?returnTo=...`
- After re-login, the route loaded

Why this matters:

- Bookmarked pages, emailed links, refreshed tabs, and new-tab opens are all unreliable.
- This is a severe usability defect for B2B software.

Likely cause:

- The browser has a JWT in `localStorage`, but the server-side route guard has no durable auth cookie to trust.

### 3. Production CSP and analytics are out of sync

Observed behavior:

- The new Google Ads conversion tag is live, but CSP does not permit its secondary network/script origins

Impact:

- Production emits repeated avoidable errors
- Analytics quality is questionable
- Security policy is not aligned with deployed third-party code

## Does The Product Deliver What The Customer Is Paying For?

Current answer: `No`

A customer buying OffGridFlow is buying the ability to:

- ingest emissions data
- calculate emissions
- review compliance outputs
- produce auditable records
- manage the subscription

The current live product fails on four of those five steps:

- ingestion fails
- compliance view crashes
- audit views are backed by missing endpoints
- billing cannot start checkout

The only part that works reliably in this audit was account creation and basic page rendering.

## Quality Assessment

Current quality level: `not acceptable for production selling`

Why:

- Broken first-run experience
- Broken deep-link session handling
- Broken import
- Broken billing
- Crashing compliance module
- Audit UI masking unavailable backend functionality
- Production console noise on every page

This is below “beta-quality customer trial.” It is not close to “highest quality.”

## Highest-Leverage Fixes

1. Fix auth architecture first.
   Use server-recognized session cookies or a coherent token strategy so protected routes survive reloads, deep links, and new tabs.

2. Fix CSV import immediately.
   Replace the old ingestion host fallback in `web/app/(app)/emissions/page.tsx:526` and align it with the current production API/CSP.

3. Remove all localhost fallbacks from production-facing clients.
   Start with `web/lib/api/carbon.ts:16`.

4. Make compliance pages defensive against partial payloads.
   `web/app/(app)/compliance/csrd/page.tsx:177` should not crash when `totals` is absent.

5. Stop masking backend failures as empty states.
   Ledger and approvals should surface backend errors instead of silently rendering “No records” when endpoints return `404`.

6. Fix billing before advertising plans.
   A new customer must be able to create a checkout session and reach Stripe successfully.

7. Normalize nullable API responses before rendering lengths/maps.
   The data-sources crash strongly suggests the frontend assumes arrays when the API can return null.

8. Align CSP with deployed third-party scripts, or remove the tag until it is configured correctly.

## Final Business Call

If I were a real buyer evaluating OffGridFlow today, I would not proceed to purchase.

Reason:

- The marketing promise is stronger than the working product.
- The product cannot carry a first-time customer from signup to value.
- The billing path does not even let me verify payment readiness.

The right next move is not more marketing. It is stabilizing the first-use journey until one new customer can:

1. sign up
2. stay signed in across reloads
3. import a CSV
4. see calculations populate
5. open a compliance report without crashing
6. create at least one audit artifact
7. reach Stripe checkout successfully

Until that sequence works on production, the product is not ready for serious buyer traffic.
