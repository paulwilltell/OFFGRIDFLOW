# Diamond-Tier SaaS Gatekeeper Audit

Date: 2026-04-13
Target: `https://off-grid-flow.com`
Rubric: `C:\Users\pault\Downloads\Diamond-Tier-SaaS-Gatekeeper.pdf`
Repo context: `C:\Users\pault\OffGridFlow`
Deployment context: Railway CLI from `C:\Users\pault\OffGridFlow`

## Scope

This audit was performed strictly against the public product and trust surface of `off-grid-flow.com`, with local repo files and Railway CLI used only as objective supporting evidence where helpful. If proof was missing, the item was not counted as passed.

## Evidence Used

- Live pages observed: `/`, `/pricing`, `/demo`, `/case-study`, `/about`, `/security`, `/privacy`, `/terms`, `/status`, `/trust`, `/methodology`, `/register`, `/login`
- HTTP checks:
  - `favicon.ico` returns `404`
  - `robots.txt` returns `404`
  - `sitemap.xml` returns `404`
- Railway CLI:
  - `railway whoami` shows authenticated session
  - `railway status` in repo shows `Project: OFFGRIDFLOW`, `Environment: production`, `Service: offgridflow-api-v2`
  - `railway service status` shows latest successful deployment `28bc787a-6554-471d-85cf-c0afd5b6bdf3`
  - `railway deployment list` shows multiple removed deployments on 2026-04-13 before the current success
  - `railway domain --json -s offgridflow-api-v2` shows `https://offgridflow-api-v2-production.up.railway.app`
- DNS checks:
  - `off-grid-flow.com` resolves and has MX
  - `offgridflow.com` resolves, but no MX was returned in the audit run
- Key source files:
  - [web/app/page.tsx](C:/Users/pault/OffGridFlow/web/app/page.tsx:7)
  - [web/app/pricing/page.tsx](C:/Users/pault/OffGridFlow/web/app/pricing/page.tsx:4)
  - [web/app/demo/page.tsx](C:/Users/pault/OffGridFlow/web/app/demo/page.tsx:6)
  - [web/app/case-study/page.tsx](C:/Users/pault/OffGridFlow/web/app/case-study/page.tsx:4)
  - [web/app/about/page.tsx](C:/Users/pault/OffGridFlow/web/app/about/page.tsx:4)
  - [web/app/security/page.tsx](C:/Users/pault/OffGridFlow/web/app/security/page.tsx:4)
  - [web/app/status/page.tsx](C:/Users/pault/OffGridFlow/web/app/status/page.tsx:13)
  - [web/app/trust/page.tsx](C:/Users/pault/OffGridFlow/web/app/trust/page.tsx:4)
  - [web/app/methodology/page.tsx](C:/Users/pault/OffGridFlow/web/app/methodology/page.tsx:4)
  - [web/app/register/page.tsx](C:/Users/pault/OffGridFlow/web/app/register/page.tsx:26)

## Scoring Method

The PDF defines thresholds and criticality, but not a numeric point model per checklist line. For a strict calculable score:

- `Met = 1.0`
- `Partial = 0.5`
- `Not met = 0.0`
- `Critical (C) = weight 2`
- `High (H) = weight 1`

Panel percent = earned weighted points / available weighted points.
Overall score = panel percent multiplied by panel weight:

- Panel 1 weight = 35
- Panel 2 weight = 30
- Panel 3 weight = 35

This does not override the rubric rule that any failed critical item or triggered auto-fail still produces a `Hold`.

## Executive Verdict

OffGridFlow does **not** pass the Diamond-Tier Gatekeeper.

Strict evidence-weighted score:

- Panel 1: `39.3%` -> `13.75 / 35`
- Panel 2: `20.0%` -> `6.00 / 30`
- Panel 3: `21.4%` -> `7.50 / 35`
- Overall: `27.25 / 100`

Required by rubric:

- Overall `>= 92/100`
- Panel 1 `>= 93/100`
- Panel 2 `>= 90/100`
- Panel 3 `>= 90/100`
- Every critical item passed
- No auto-fail triggered

Actual result:

- Overall threshold missed by `64.75` points
- All three panel thresholds missed
- Multiple critical items failed
- Auto-fails triggered

## Auto-Fails Triggered

### 1. Numbers, records, or outputs cannot be traced back to source evidence

Why this triggers:

- The homepage and case study use quantified commercial claims without visible source pack:
  - homepage metadata says `Starting at $4,800/year` while hero copy says `$6,500/year` and case study impact says `$4,800/year`
  - case study uses `86% savings`, `<2 hrs`, `21.36 tonnes CO2e`, and consultant quote figures without source artifact or downloadable proof
- The demo page uses synthetic KPI statements and social proof without substantiation:
  - `↓ 62%`, `Reduce audit effort by 40%`, `97% evidence coverage`, `0 critical gaps`
  - fake analyst markers: `Gartner MQ`, `Forrester Wave`, `IDC MarketScape`
  - fake press markers: `Forbes`, `WSJ`, `TechCrunch`
  - testimonial from `VP Sustainability, Fortune 500 Manufacturing`

### 2. Product promise, website story, and sales claim do not match the actual experience

Why this triggers:

- The public site presents the product as carbon accounting and climate compliance.
- The public demo page pivots into generic enterprise governance and security theater:
  - `Policy control map`
  - `Vendor risk review`
  - `HIPAA Ready`
  - `Top performer in governance automation`
- The pricing and acquisition story is internally inconsistent:
  - homepage, demo page, and case study say `Start Free Trial`
  - pricing FAQ says `Request a demo`
  - the Global plan button says `Contact Sales` but routes to `/register?plan=global`
  - the register page ignores the `plan` query string entirely

## Highest-Risk Findings

1. The demo page is the single biggest trust destroyer. It reads like a different product category and injects fabricated enterprise proof. See [web/app/demo/page.tsx](C:/Users/pault/OffGridFlow/web/app/demo/page.tsx:6), [web/app/demo/page.tsx](C:/Users/pault/OffGridFlow/web/app/demo/page.tsx:139), and [web/app/demo/page.tsx](C:/Users/pault/OffGridFlow/web/app/demo/page.tsx:509).
2. Pricing and offer logic are broken in the funnel. The Global tier `Contact Sales` CTA routes to registration because the code checks for `Global Enterprise` while the actual tier is `Global`. See [web/app/pricing/page.tsx](C:/Users/pault/OffGridFlow/web/app/pricing/page.tsx:62) and [web/app/pricing/page.tsx](C:/Users/pault/OffGridFlow/web/app/pricing/page.tsx:159).
3. The register page does not capture or use the selected plan, so plan-specific CTAs do not produce a plan-specific onboarding path. See [web/app/register/page.tsx](C:/Users/pault/OffGridFlow/web/app/register/page.tsx:26) and [web/app/register/page.tsx](C:/Users/pault/OffGridFlow/web/app/register/page.tsx:68).
4. Contact addresses are inconsistent across domains:
  - `contact@off-grid-flow.com`
  - `paul@off-gridflow.com`
  - `sales@offgridflow.com`
  The audited DNS run found MX only on `off-grid-flow.com`, not on `offgridflow.com`. This is a direct commercial-trust risk.
5. Basic website hygiene is incomplete:
  - no `favicon.ico`
  - no `robots.txt`
  - no `sitemap.xml`

## Panel 1

### 1A. Problem integrity and product truth

- `1A.1 [PARTIAL][C]` The pain is credible, recurring, and budget-shaped, but there is no public ICP proof pack, discovery evidence, or win/loss artifact. Recommendation: publish a one-page ICP and trigger memo with buyer roles, compliance trigger events, and quantified cost of delay.
- `1A.2 [FAIL][C]` The marketed promise does not match the shipped sales experience. The homepage sells carbon compliance; the demo sells synthetic enterprise governance theater. Recommendation: rebuild `/demo` around carbon workflows only.
- `1A.3 [FAIL][H]` The product does not hold a clean canonical use case on the public surface. It oscillates between climate reporting, security governance, and generic enterprise ops. Recommendation: choose one primary use case and make every public page subordinate to it.

### 1B. Core product and data integrity

- `1B.1 [FAIL][C]` No live artifact proves a canonical data model or unified source of truth. Recommendation: publish an entity map for activity, factor, calculation, report, evidence, approver.
- `1B.2 [PARTIAL][C]` The methodology library references factor sources and ledger concepts, but there is no public version-history artifact showing prior output reconstruction. Recommendation: show factor versioning and a real before/after recalculation example.
- `1B.3 [PARTIAL][C]` The methodology page claims traceability, but no public drill-down or sample audit packet proves source lineage on an actual report. Recommendation: publish a redacted report where every output links to activity record, factor source, and calculation record.
- `1B.4 [PARTIAL][H]` Imports, exports, and integrations are described, but there is no public proof of idempotent import behavior, duplicate protection, or export parity. Recommendation: add a sample import log, duplicate handling screenshot, and matched export example.

### 1C. Security, privacy, and governance

- `1C.1 [PARTIAL][C]` There is objective code and page evidence for auth, lockout, CSRF, and API keys, but tenant isolation and MFA are asserted more than demonstrated. Recommendation: publish an RBAC matrix and one tenant-isolation test artifact.
- `1C.2 [PARTIAL][C]` Ownership, export, deletion, and retention are documented across privacy/security/trust pages, but execution proof is missing. Recommendation: add a customer-facing deletion/export workflow walkthrough.
- `1C.3 [PARTIAL][H]` Security posture is inspectable because `/security`, `/trust`, `/status`, and `/methodology` exist, but the demo page contaminates trust with unsupported badges. Recommendation: strip all unsupported certifications and analyst references.
- `1C.4 [PARTIAL][H]` Governance intent is visible through subprocessor and retention text, but there is no public data classification, minimization, or redaction standard. Recommendation: publish a short data classification and minimization policy.

### 1D. Reliability, service operations, and scale

- `1D.1 [PARTIAL][C]` `/status` publishes SLO-style targets and a live API health check, but monitoring, alerting, and recovery discipline are not evidenced end-to-end. Recommendation: add uptime history sourced from actual telemetry and incident postmortem links.
- `1D.2 [PARTIAL][C]` Railway shows a current successful production deployment, but no public rollback or release review discipline is documented. Recommendation: publish change log, rollback policy, and release quality checklist.
- `1D.3 [PARTIAL][H]` Incident communication exists on `/status`, but trust is weakened by inconsistent contact domains and hidden trust surfaces. Recommendation: unify all trust/support contact paths on `off-grid-flow.com` and link `/status` from the main nav or footer.
- `1D.4 [FAIL][H]` There is no objective load, concurrency, or scale evidence. Recommendation: publish performance benchmarks for report generation, ingestion, and dashboard latency.

### 1E. Usability and regulated-domain overlay

- `1E.1 [PARTIAL][C]` The home and pricing flow are understandable, but the demo dilutes the main job path. Recommendation: make the primary path explicit: import data -> validate evidence -> calculate -> review -> export.
- `1E.2 [PARTIAL][H]` Forms are labeled and structurally decent, but there is no accessibility proof and no published check. Recommendation: run Axe and keyboard-only audits on marketing, auth, and app flows; publish remediation.
- `1E.3 [PARTIAL][H]` The public methodology page covers factor sources, Scope 2 method split, and standards mapping, but the emissions-specific overlay is incomplete. Recommendation: publish proof for evidence vault, approval trail, factor provenance by version, and framework output mapping.

## Panel 2

### 2A. Clarity, orientation, and role fit

- `2A.1 [FAIL][C]` The surface does not consistently fit the product type. The demo is not a carbon-accounting demo. Recommendation: replace governance/security abstractions with carbon-specific screens and workflows.
- `2A.2 [PARTIAL][C]` A new user can orient on the homepage, but orientation degrades once they enter the demo or plan selection flow. Recommendation: make the first five minutes consistent across homepage, demo, pricing, and register.
- `2A.3 [FAIL][H]` Role adaptation fragments truth instead of clarifying it. Recommendation: keep one canonical product and let roles change priorities, not product identity.

### 2B. Traceability, confidence, and explainability

- `2B.1 [PARTIAL][C]` The methodology page explains factors and methods, but public metrics on the homepage and demo are not explainable in-place. Recommendation: every metric should have a visible source note or supporting artifact.
- `2B.2 [FAIL][C]` No public drill-down from summary to record-level proof is visible. Recommendation: show a real redacted drill-down path in the demo or methodology library.
- `2B.3 [FAIL][C]` Actuals, estimates, freshness, confidence, and review state are not labeled on public outputs. Recommendation: tag every visible output as actual, sample, estimate, draft, or approved.
- `2B.4 [FAIL][H]` There is no demonstrated anomaly workflow. Recommendation: publish one alert scenario with root cause, affected records, and next action.

### 2C. Actionability and workflow continuity

- `2C.1 [FAIL][C]` No public evidence shows alerts/exceptions resolving into action. Recommendation: demonstrate assign, comment, approve, reject, and remediate paths.
- `2C.2 [FAIL][C]` Ownership and approval state are not visible in the working surface evidence provided. Recommendation: show owner, reviewer, approver, and blocked status inside the main workflow.
- `2C.3 [FAIL][H]` Saved views and recurring workflows are not demonstrated. Recommendation: publish a repeatable month-close or quarter-close workflow.

### 2D. Performance, accessibility, and composure under scale

- `2D.1 [PARTIAL][C]` Marketing pages respond quickly, but realistic application-load evidence is absent. Recommendation: publish p95 targets and measured results for real app actions.
- `2D.2 [PARTIAL][H]` Public tables and forms are readable, but this is not evidence of large-scale operational composure. Recommendation: show dense tables, long reports, and large imports remaining usable.
- `2D.3 [FAIL][H]` Dynamic-view accessibility is not evidenced. Recommendation: audit charts, tabs, filters, and modals for keyboard and screen-reader behavior.

### 2E. Output quality and stakeholder communication

- `2E.1 [FAIL][C]` No public export reconciliation proves on-screen truth equals exported truth. Recommendation: publish one matched screen-to-PDF or screen-to-XBRL example.
- `2E.2 [PARTIAL][H]` Trust and methodology pages are stakeholder-oriented, but there is no true board-ready or auditor-ready output pack on display. Recommendation: add a downloadable redacted audit packet.
- `2E.3 [MET][H]` The public site uses plain language well in FAQ, methodology, and trust content. Recommendation: keep this plain-language standard, but tie it to real proof.

## Panel 3

### 3A. ICP precision, positioning, and proof of pain

- `3A.1 [FAIL][C]` No public evidence shows explicit disqualifiers, fit boundaries, or maturity gating. Recommendation: publish fit criteria by company size, data maturity, and regulatory scope.
- `3A.2 [FAIL][C]` No public asset maps champion, economic buyer, operator, and blocker. Recommendation: build a buyer-map one-pager and use it in demo routing.
- `3A.3 [PARTIAL][C]` Category, pain, and differentiator are clear; proof is not. Recommendation: replace synthetic proof with one real named or clearly anonymized evidence-backed case.

### 3B. Website, offer, and pricing clarity

- `3B.1 [MET][C]` The homepage makes the product and next step clear quickly. Recommendation: preserve this clarity while cleaning the trust layer.
- `3B.2 [FAIL][C]` Pricing and packaging are not commercially coherent. Recommendation: unify the starting price everywhere, fix the Global CTA, and make plan selection persist into registration and onboarding.
- `3B.3 [FAIL][H]` The land-prove-expand path is not clean. Recommendation: define one conversion path: discovery call, guided trial, assisted onboarding, expansion criteria.

### 3C. Sales motion, procurement trust, and proof assets

- `3C.1 [PARTIAL][C]` The demo is segmented by role, but the role content is not disciplined to real carbon pain. Recommendation: tailor by CFO, sustainability lead, controller, compliance owner, not generic security archetypes.
- `3C.2 [PARTIAL][C]` Procurement-facing materials exist, but they are buried and polluted by unsupported claims elsewhere. Recommendation: move `/trust`, `/status`, and `/methodology` into the main trust path and purge all faux proof.
- `3C.3 [FAIL][H]` The proof assets are not believable enough because key social proof is synthetic or unverifiable. Recommendation: either publish one real proof asset or label every synthetic asset explicitly as illustrative.

### 3D. Onboarding, time-to-value, and adoption design

- `3D.1 [PARTIAL][C]` A first-value story exists in the case study and homepage, but milestones are not visibly mapped or measured. Recommendation: publish first login, first import, first calculation, first report, first approval milestones.
- `3D.2 [FAIL][C]` No public evidence shows segment-specific onboarding tracks. Recommendation: create self-serve, assisted, and enterprise onboarding checklists.
- `3D.3 [FAIL][H]` No visible training or enablement library exists by role. Recommendation: publish role-based onboarding guides and sample admin/operator/executive quick starts.

### 3E. Renewal, expansion, economics, and commercial operating rhythm

- `3E.1 [FAIL][C]` No public evidence of health scoring or renewal engineering. Recommendation: define renewal health signals and sponsor review cadence.
- `3E.2 [FAIL][C]` No evidence that expansion follows proven value realization. Recommendation: document expansion triggers tied to adoption, entities, and additional frameworks.
- `3E.3 [FAIL][C]` No commercial quality metrics are visible. Recommendation: internally track win rate, payback, churn risk, onboarding duration, and services drag before claiming scale readiness.
- `3E.4 [FAIL][H]` No visible feedback loop from sales/success into product and messaging. Recommendation: create recurring win/loss, churn, and friction reviews.

## Specific Contradictions and Defects

### Pricing and offer contradictions

- Homepage metadata says `Starting at $4,800/year` while hero body says `$6,500/year`. See [web/app/page.tsx](C:/Users/pault/OffGridFlow/web/app/page.tsx:9) and [web/app/page.tsx](C:/Users/pault/OffGridFlow/web/app/page.tsx:49).
- Case study impact still says `$4,800/year`. See [web/app/case-study/page.tsx](C:/Users/pault/OffGridFlow/web/app/case-study/page.tsx:134).
- Pricing Global CTA is miswired to registration, not contact sales. See [web/app/pricing/page.tsx](C:/Users/pault/OffGridFlow/web/app/pricing/page.tsx:159).
- Register page does not read a selected plan, so the funnel drops pricing context. See [web/app/register/page.tsx](C:/Users/pault/OffGridFlow/web/app/register/page.tsx:26).

### Trust contradictions

- Demo shows `SOC 2 Type II`, `ISO 27001`, `HIPAA Ready`, analyst badges, and press badges as live trust symbols. See [web/app/demo/page.tsx](C:/Users/pault/OffGridFlow/web/app/demo/page.tsx:139) and [web/app/demo/page.tsx](C:/Users/pault/OffGridFlow/web/app/demo/page.tsx:563).
- Security page says certification is still on roadmap. See [web/app/security/page.tsx](C:/Users/pault/OffGridFlow/web/app/security/page.tsx:71).
- Trust center says `SOC 2 Type I` is in preparation and `SSO / SAML` is planned. See [web/app/trust/page.tsx](C:/Users/pault/OffGridFlow/web/app/trust/page.tsx:27).

### Contact-path contradictions

- Valid MX during audit was found on `off-grid-flow.com`.
- Public pages still use:
  - `paul@off-gridflow.com`
  - `sales@offgridflow.com`
  - `contact@off-grid-flow.com`
- This is avoidable trust damage in procurement and support flows.

### Website hygiene defects

- `favicon.ico` missing
- `robots.txt` missing
- `sitemap.xml` missing
- Status/trust/methodology exist but are not surfaced as first-class trust navigation on the homepage

## What Must Be Done

### Immediate fixes before any serious sales push

1. Delete or rewrite `/demo`.
   - No fake analyst badges
   - No fake press badges
   - No synthetic testimonial unless clearly labeled as illustrative
   - No security-governance identity drift
2. Unify pricing and offer logic.
   - One starting price everywhere
   - One trial/demo story everywhere
   - Fix Global CTA routing
   - Persist selected plan into registration and onboarding
3. Unify all contact addresses on one mail-receiving domain.
4. Add `favicon.ico`, `robots.txt`, and `sitemap.xml`.
5. Link `/trust`, `/status`, and `/methodology` from the primary nav or footer trust block.

### Must-do evidence upgrades for Diamond threshold

1. Publish one real or rigorously anonymized case study with downloadable proof artifact.
2. Publish one real redacted output pack:
   - input records
   - factor source/version
   - calculation ledger
   - report export
3. Publish a procurement pack:
   - security architecture
   - subprocessor list
   - deletion/export workflow
   - incident response policy
   - certification roadmap
4. Publish a first-value map:
   - first login
   - first import
   - first successful calculation
   - first approved report
   - median time-to-first-value
5. Publish performance and reliability proof:
   - SLO measurement source
   - uptime history source
   - ingestion latency
   - report generation latency

## Path to 101

The rubric caps at `100`, so “101” means exceeding the standard in proof quality, not literally scoring above the sheet. To get there:

### Step 1. Reach honesty-first compliance

- Strip every unsupported claim.
- Remove all faux proof.
- Replace every contradictory number with one canonical source.

### Step 2. Make the public story match the actual product

- Homepage, pricing, demo, case study, trust, and onboarding must all describe the same primary job:
  - ingest emissions data
  - validate evidence
  - calculate scopes
  - review approvals
  - export disclosure-ready outputs

### Step 3. Turn trust from decorative to inspectable

- Keep `/security`, `/trust`, `/status`, `/methodology`
- Add:
  - downloadable evidence pack
  - real audit packet sample
  - real uptime source
  - real contact ownership

### Step 4. Turn the website into a coherent commercial machine

- One ICP
- One buyer trigger model
- One first-value path
- One clean land-prove-expand motion

### Step 5. Make every critical item pass

The Diamond sheet is not a brand exercise. It is a proof exercise. OffGridFlow only becomes “101-level” when a skeptical buyer, auditor, or procurement lead can verify the product promise without trusting the marketing voice.

## Bottom Line

Current grade:

- `27.25 / 100`
- `HOLD`

Real interpretation:

- The product concept is credible.
- The methodology surface is promising.
- The trust infrastructure pages are directionally useful.
- The current public sales surface is not diamond-tier because it overclaims, fragments the product story, and substitutes synthetic proof where objective evidence is required.
