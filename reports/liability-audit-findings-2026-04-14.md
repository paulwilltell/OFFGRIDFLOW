# OffGridFlow Liability Audit — Findings Report

**Date:** 2026-04-14
**Auditor:** Senior Full-Stack Engineer + Legal-Risk Auditor
**Product:** OffGridFlow carbon compliance SaaS (~$6,500–$15,000/year)
**Scope:** Full codebase (`C:\Users\pault\OffGridFlow\web`), live site (`https://off-grid-flow.com`)
**Methodology:** Pattern-based static analysis across four risk categories — Data Integrity, Availability & Reliability, Security & Privacy, Legal & Contractual.

---

## Executive Summary

Three **Critical data-integrity bugs** in production code paths are generating fabricated emissions and uptime numbers visible to paying customers. These are direct lawsuit triggers for a compliance tool. All three were previously hidden by REST-first data flow but are executed the moment real data arrives in edge cases. Additionally, four pages in the authenticated shell lack error boundaries, and one public page uses `dangerouslySetInnerHTML` with lightly-processed content.

The session/auth layer and WebSocket layer were already hardened in a prior session. The Terms of Service and Privacy Policy exist but need limitation-of-liability tightening and explicit acceptance at registration.

---

## Risk Table

| # | Category | Severity | File / URL | Finding | Proposed Fix |
|---|---|---|---|---|---|
| 1 | Data Integrity | **CRITICAL** | `web/components/charts/EmissionChart.tsx:33-51` | `(Math.random() - 0.5) * 0.1 * baseValue` applied to chart data. Fabricates trend variance on top of real emissions data. | Replace with deterministic per-period values from the real API (activities grouped by period). Remove all randomness. |
| 2 | Data Integrity | **CRITICAL** | `web/components/charts/EmissionChartJS.tsx:93-113` | When no real historical data is present, generates 12 months of FAKE historical emissions via random variance + synthetic trend. A customer sees 12 months of fabricated history they never reported. | Remove synthetic history generation entirely. Show an empty-state "No historical data yet" message instead. |
| 3 | Data Integrity | **CRITICAL** | `web/components/StatusPage.tsx:466-477` | Generates 90 days of fake uptime history (`99.5 + Math.random() * 0.5`) and fake incidents (`Math.random() > 0.95`). Published to customers as reliability evidence. False advertising + misrepresentation. | Remove entirely. Show "Uptime history is available by request. Live health check is above." |
| 4 | Data Integrity | Medium | `web/stores/carbonStore.ts:141-150` | Mock data path exists but is guarded by `process.env.NODE_ENV === 'development'` — safe in production. | Add runtime assertion: `throw new Error()` if ever reached outside dev. Belt-and-suspenders. |
| 5 | Data Integrity | Low | `web/components/visualizations/DataGlobe.tsx:91-92` | Random lat/lng on 3D globe visualization. Purely decorative, no emissions data affected. | No change required. Add comment clarifying decorative intent. |
| 6 | Availability | **CRITICAL** | `web/app/(app)/layout.tsx` | No error boundary wrapping the authenticated app shell. A crash in any child page takes down the entire dashboard → user gets a white screen or is kicked to login. | Add a top-level `<AppErrorBoundary>` around `{children}` in the app layout with a reset button. |
| 7 | Availability | High | `web/app/(app)/dashboard/carbon/page.tsx` | Dashboard not explicitly wrapped in its own error boundary. Relies on parent boundary only. | Already has `<ErrorBoundary>` via CarbonDashboard component. Verified OK. |
| 8 | Availability | High | `web/app/(app)/compliance/*/page.tsx` | Compliance framework pages (CSRD, SEC, CBAM, California, Scope 3) have no error boundary. A failing framework export crashes the tree. | Wrap each compliance page in `<ErrorBoundary>`. |
| 9 | Availability | High | `web/app/(app)/audit/*/page.tsx` | Audit pages (ledger, approvals, alerts, data-quality, factor-snapshots) lack boundaries. | Wrap each in `<ErrorBoundary>`. |
| 10 | Availability | FIXED | `web/providers/RealTimeDataProvider.tsx` | Previously used `Math.random()` fallback to fabricate emissions data when WebSocket failed. Fixed in commit `68ec514`. | Already resolved. |
| 11 | Availability | FIXED | `web/lib/session.tsx` | Previously cleared session on any API error, kicking users to login on transient failures. Fixed in commit `68ec514`. | Already resolved. |
| 12 | Security | Medium | `web/app/usps-advocate/page.tsx:364-375` | `dangerouslySetInnerHTML` renders content with only `<strong>` and `>` replacements. `line` source comes from a static array — not user input — but pattern is risky. | Replace with proper React component rendering (no HTML injection). Or wrap content through DOMPurify if keeping. |
| 13 | Security | Low | `web/app/layout.tsx:38` | `dangerouslySetInnerHTML` for Google Ads gtag. Content is static, not user-sourced. | No change. Acceptable pattern for gtag. |
| 14 | Security | OK | API keys in client code | Only `SENDGRID_API_KEY` referenced in `web/app/api/elite-inquiry/route.ts` — server-side Next.js route, not exposed to browser. | No change. |
| 15 | Privacy | High | Registration flow | No explicit Terms & Privacy acceptance checkbox. User implicitly agrees by clicking Create Account but there is no audit trail. | Add required checkbox with link to `/terms` and `/privacy`. Log timestamp on submission. |
| 16 | Privacy | High | Cookie consent | No cookie consent banner for EU/UK users. Google Ads tag fires on page load regardless of consent. | Add minimal consent banner that gates gtag initialization. Only mandatory for EU/UK but safer to show globally. |
| 17 | Privacy | Medium | CCPA "Do Not Sell" link | Absent from footer. Required in California even if no data is sold. | Add footer link to a short CCPA statement page. |
| 18 | Legal | High | `/terms` page | Exists but limitation-of-liability clauses are thin. Needs cap on damages (e.g., fees paid in prior 12 months), "AS IS" warranty disclaimer, no uptime guarantee, data-accuracy disclaimer. | Expand Terms. Clearly mark as draft pending attorney review. |
| 19 | Legal | High | `/privacy` page | Does not explicitly enumerate subprocessors (Stripe, Railway, SendGrid, Google Analytics/Ads). GDPR requires this. | Add subprocessor table with company name, purpose, data accessed, region. |
| 20 | Legal | Medium | Pricing / homepage disclaimers | No visible disclaimer that OffGridFlow is a calculation tool, not a certified audit or legal advice. Customers may rely on it for regulatory submission. | Add short disclaimer on pricing and compliance pages: "OffGridFlow does not guarantee regulatory acceptance; consult your compliance advisor before submission." |
| 21 | Legal | Medium | Refund policy | No explicit refund policy published. Ambiguity invites chargebacks. | Publish clear refund policy: "Annual plans: 14-day refund window; after 14 days, no refunds. Monthly: no refunds, cancel anytime." |

---

## Risk Counts by Severity

- **CRITICAL: 4** — 3 data integrity, 1 availability (app shell boundary)
- **High: 7** — availability gaps + registration/privacy/legal
- **Medium: 5** — security pattern + legal disclaimers + refund policy
- **Low: 2** — decorative Math.random + Google gtag innerHTML
- **Fixed (from prior session): 2** — RealTimeProvider + session refresh

**Total open findings: 18**

---

## Remediation Priority (Phase 2 Order)

1. Remove fake emissions variance from `EmissionChart.tsx` and `EmissionChartJS.tsx`
2. Remove fake uptime history from `StatusPage.tsx`
3. Add `<AppErrorBoundary>` to `web/app/(app)/layout.tsx`
4. Wrap compliance and audit pages with `<ErrorBoundary>`
5. Add explicit Terms/Privacy acceptance checkbox to registration
6. Expand `/terms` with limitation of liability, AS IS disclaimer
7. Expand `/privacy` with subprocessor table
8. Add calculation disclaimer to pricing and compliance pages
9. Publish refund policy
10. Add cookie consent banner + CCPA footer link

---

## Liability Readiness Score (Pre-Fix Baseline)

| Category | Weight | Score | Weighted |
|---|---|---|---|
| Data Integrity | 30 | 40/100 | 12.0 |
| Availability & Reliability | 25 | 70/100 | 17.5 |
| Security & Privacy | 20 | 75/100 | 15.0 |
| Legal & Contractual | 25 | 55/100 | 13.75 |
| **Total** | **100** | | **58.25 / 100** |

**Pass threshold: 95.** Current exposure is material — three critical fake-data bugs alone could produce a single-claim case that exceeds annual ARR.

Remediation plan in Phases 2–7 is projected to raise the score to **96–98**.

---

## Notes for Legal Review

All legal text created or modified in Phase 4 will be clearly marked "DRAFT — PENDING ATTORNEY REVIEW." The Terms of Service and Privacy Policy updates are templates intended to accelerate attorney review, not substitute for it. Recommended next steps: obtain cyber liability insurance ($1M–$2M aggregate typical for SaaS at this ARR) and engage a technology attorney to finalize Terms, DPA, and SLA language.
