# Phase R — Defect Report: Risk Abatement & Justification Engine

**Scope:** Read-only review of Codex-authored work in commit `bd1cc48`.
**Reviewer:** Claude Opus 4.6 (1M).
**Date:** 2026-04-14.

## Verdict

The implementation is structurally sound and feature-complete against the specification.
Wiring, types, evaluators, and UI are all in place. However, there are **blocking
infrastructure gaps** and **minor correctness issues** that must be resolved before the
feature can run end-to-end on Railway.

## Blocker (B) / Important (I) / Minor (M) defects

### B-1. `approval_workflow` table is not defined in any migration
- **Where:** `internal/abatement/service.go:470` queries `approval_workflow`.
- **Impact:** `BuildDashboard` → `buildRiskFacts` → `countApprovals` executes
  `SELECT COUNT(*) FROM approval_workflow WHERE tenant_id = $1 ...`. On Railway,
  the table does not exist. The dashboard endpoint will fail with
  `relation "approval_workflow" does not exist`.
- **Pre-existing:** `internal/audit/store.go` already references this table, so the
  underlying gap predates Codex's work — but the new feature exposes the gap on a
  customer-facing route (Pro plan).
- **Fix:** Add `approval_workflow` (tenant_id, entity_type, entity_id, status,
  prepared_by, prepared_at, reviewed_at, reviewed_by, approved_at, approved_by,
  created_at, updated_at) to migration 0003 (or a new 0004). Alternatively, guard
  the count with a `to_regclass` check so the dashboard degrades gracefully.

### B-2. `factor_snapshots` column name mismatch risk
- **Where:** `internal/abatement/service.go:457` queries
  `factor_snapshots WHERE organization_id = $1`.
- **Status:** `infra/db/migrations/000002_diamond_tier_audit.up.sql` does define
  `factor_snapshots(organization_id, ...)` — **match ✓**.
- **Risk:** Memory notes `emission_factors.value_kg_co2e_per_unit` is missing from
  Railway (column drift). If migration `000002` was only partially applied, the
  dashboard will also fail. Verify migration state on Railway before shipping.

### I-1. Evidence-path parsing fragility in handler
- **Where:** `internal/abatement/handler.go:123-141` `handleEvidence`.
- **Issue:** The handler re-parses `r.URL.Path` from scratch and takes the last
  segment as evidence ID. Works for `/api/abatement/{framework}/evidence/{id}`
  but breaks for any trailing slash or added query path. The `suffix` returned
  by `parseFrameworkPath` is already available — use it.
- **Severity:** Won't crash; will 404 on malformed input (acceptable). Left as I
  because it's brittle and inconsistent with how other endpoints dispatch.

### I-2. `handleEvidence` does not verify framework on evidence record
- **Where:** `service.GetEvidence` → `store.GetEvidence`.
- **Issue:** Query is `WHERE tenant_id = $1 AND id = $2`. If the caller is on
  `/api/abatement/sb253/evidence/<csrd-evidence-id>`, the file is returned
  regardless of framework. RLS protects cross-tenant access, but cross-framework
  leakage within a tenant is possible.
- **Fix:** Add `AND framework = $3` to the evidence query and pass the framework
  parsed from the URL.

### I-3. `handleReport` doesn't set `Cache-Control: no-store`
- **Impact:** Generated PDFs can be cached by intermediaries. Low risk for this
  app but best practice for report downloads.

### M-1. ~~`nullString` unused~~ — withdrawn, helper is used by `SaveEvaluation`.

### M-2. `evidenceTypes` unused
- **Where:** `internal/abatement/evaluators.go:351-366`. Never called. Likely a
  placeholder for stricter required-evidence-type enforcement.

### M-3. `strings.Title` is deprecated
- **Where:** `service.go:295, 297, 339`. `strings.Title` was deprecated in
  Go 1.18 in favor of `golang.org/x/text/cases`. Won't fail build but emits a
  linter warning.

### M-4. `ComplianceCheckID` field tag mismatch in SelfCertificationRequest
- **Where:** `types.go:119` — `ComplianceCheckID string \`json:"compliance_check_id,omitempty"\``.
- **Issue:** The TS client sends `complianceCheckId` (camelCase) via `api.post`,
  but the Go struct expects `compliance_check_id` (snake). `api.post` is the
  JSON client and serializes whatever object is passed as-is. Check
  `lib/api/abatement.ts:86-90` — it sends `compliance_check_id: payload.complianceCheckId`.
  ✓ matches. So this is actually fine, but the internal TS type is `complianceCheckId`.
  No bug, noted for future-reader clarity.

### M-5. Go/TS evaluator logic drift
- **Where:** `web/lib/abatement/evaluator.ts` only implements
  `evaluateScope2MarketBased`. The Go side implements 9 evaluators. The client
  mirror is a stub. The feature always calls the server for authoritative
  evaluation, so this is acceptable, but the client-side test file claims to
  verify the evaluator — it only covers one.

### M-6. `RichTextInput` → plain-text round trip
- **Where:** `web/components/abatement/RiskCard.tsx:83-91`. HTML is stripped
  before submission. Fine for evaluator keyword matching, but the stored
  `justification` loses formatting that the user typed. If the PDF report
  reads `justification` from DB, all formatting is gone — consistent but worth
  flagging in the UX copy ("we store a plain-text version").

## What works as advertised

- `catalog.go` framework definitions match the spec (SB 253, CSRD, SEC, IFRS,
  CBAM) with appropriate severity/priority and trigger conditions.
- `evaluators.go` deterministic keyword matching is conservative and returns
  one of three well-defined statuses. Unit tests cover the critical path.
- `store.go` correctly scopes every operation to `tenant_id`, opens
  transactions with `set_config('app.tenant_id', ...)` for RLS, and uses
  parameterized queries throughout — no SQL injection risk.
- Migration 0003 ships RLS policies for both tables.
- Router wiring (`internal/api/http/router.go:460-467`) is behind
  `requireProPlan`, matching the "premium feature" intent.
- Frontend uses React Query with optimistic updates and proper rollback;
  `QueryProvider` is wired in `providers.tsx`.
- Nav entry added in `web/app/(app)/layout.tsx:42`.

## Recommended fix order

1. **B-1** — add `approval_workflow` table to a new migration `0004`. Without
   this, the dashboard cannot load on Railway.
2. **I-2** — add framework scoping to evidence retrieval.
3. **M-3** — replace `strings.Title` with a local `titleCase` helper.
4. **I-3** — add `Cache-Control: no-store` to report response.
5. **M-2** — delete dead `evidenceTypes` helper (or wire it into a check).
6. Defer M-5/M-6 — accept current behavior, track in backlog.

## Applied fixes (post-report)

- **B-1:** `approval_workflow` and `factor_snapshots` added to
  `internal/db/schema.sql`; standalone migration `0004_audit_support`
  created for golang-migrate environments.
- **I-2:** `store.GetEvidence` now filters by framework; handler passes the
  parsed framework through.
- **I-3:** `Cache-Control: no-store, max-age=0` added to the report and
  evidence responses.
- **M-2:** Dead `evidenceTypes` helper and stale `slices` import removed.
- **M-3:** Replaced `strings.Title` (deprecated) with local `titleCase` helper.

## Schema drift risk (Paul, decide)

Two parallel schema definitions exist:
- `infra/db/schema.sql` (tenants, users with tenant_id)
- `infra/db/migrations/000001_initial_schema.up.sql` (organizations, users with organization_id)

The abatement feature assumes the first. If Railway is running the migration
set, tables `tenants` and `users.tenant_id` may not exist with those exact names.
Verify which is live (`SELECT table_name FROM information_schema.tables`) before
running migration 0003.
