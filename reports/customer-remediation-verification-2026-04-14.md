# Customer Remediation Verification

Date: 2026-04-14

## Deployments

- API service `offgridflow-api-v2`: `2195e279-fdfa-4c79-b7d8-95ee0b9b63a9` `SUCCESS`
- Web service `offgridflow-web`: `2304f5b1-f3c7-48a2-ba46-2b78687b3ee9` `SUCCESS`

## Commits

- `0658e9c` `fix(core): restore onboarding, import, billing, and audit traceability`
- `5933158` `fix(compliance): normalize latest-year reporting views`

## Live Checks

- Created a fresh production user account:
  - `gate.customer.20260414223019@example.com`
- Verified registration redirected into the authenticated app shell without a login loop.
- Verified protected routes remained accessible after direct navigation and hard route loads.
- Imported `sample-data/offgridflow-llc-utility-data.csv` on `/emissions`.
  - 24 `POST /api/emissions/activities` requests returned `201`
  - UI confirmed `24 Activities Imported`
- Verified emissions explorer after refresh:
  - `26.44 tCO2e`
  - `88,711 kWh`
  - `24` records
  - meter/location/period fidelity present in table rows
- Verified audit ledger after import:
  - `GET /api/audit/ledger` returned `200`
  - ledger rendered populated Scope 2 calculation rows for the fresh tenant
- Verified billing handoff:
  - `POST /api/billing/checkout` returned `200`
  - browser redirected to hosted Stripe Checkout
  - stopped before any payment entry
- Verified compliance views on fresh tabs after the final web deploy:
  - `/compliance/csrd` showed real 2025 Scope 2 totals and validation output
  - `/compliance/sec` showed `In Progress` and resolved to reporting year `2025`
  - `/compliance/california` showed `In Progress` and resolved to reporting year `2025`
  - `/compliance/cbam` showed `In Progress` and resolved to reporting year `2025`
- Verified `/settings/data-sources` rendered without the previous null-length crash.
- Verified dashboard and validated pages were console-clean for the checked paths.

## Local Validation

- `npm run build --workspace web` passed after the core flow fixes.
- `npm run build --workspace web` passed again after the compliance normalization fixes.

## Constraint / Caveat

- A local Go toolchain was not available in the active PowerShell environment, so backend verification in this pass relied on:
  - successful Railway deployments
  - live API responses
  - full browser-based customer-path validation
