# Diamond-Tier Reaudit Delta 3

Date: 2026-04-13
Project: OffGridFlow
Domain audited: https://off-grid-flow.com
Repo audited: `C:\Users\pault\OffGridFlow`

## Verdict

Strict live verdict: `HOLD`

Strict live score: `~35.78 / 100`

Reason: several claimed fixes are either only partially true in source, or not yet reflected on the public site. The live site is the thing being audited.

## Railway verification

PowerShell + Railway CLI output confirms the current CLI context is still:

- Project: `OFFGRIDFLOW`
- Environment: `production`
- Service: `offgridflow-api-v2`
- Latest successful deployment visible from this context: `28bc787a-6554-471d-85cf-c0afd5b6bdf3`

This is an API service context, not confirmed frontend hosting context. That means frontend marketing pages can remain stale even while Railway reports a healthy production deployment.

## Claim-by-claim delta

### 1. Unsupported proof language

Status: `NOT CLEARED`

Live site still contains unsupported or over-assertive language:

- Home still says `SB 253 reporting deadline approaching — 5,300+ companies must comply`.
- Home still says `the same GHG Protocol-compliant calculations and audit-ready reports`.
- About still says `the same GHG Protocol-compliant calculations and audit-ready reports`.
- Pricing FAQ still says `GHG Protocol-compliant calculation methodologies`.

Source is improved, but not fully clean:

- [web/app/about/page.tsx](C:/Users/pault/OffGridFlow/web/app/about/page.tsx:40) now uses `using GHG Protocol methodology`.
- [web/app/page.tsx](C:/Users/pault/OffGridFlow/web/app/page.tsx:172) still says `GHG Protocol-compliant engine`.
- [web/app/pricing/page.tsx](C:/Users/pault/OffGridFlow/web/app/pricing/page.tsx:90) still says `GHG Protocol-compliant calculation methodologies`.
- [web/app/terms/page.tsx](C:/Users/pault/OffGridFlow/web/app/terms/page.tsx:73) still says `GHG Protocol-compliant calculation methodologies`.

Audit impact:

- Auto-fail proof blocker remains active.

### 2. Favicon 404

Status: `REPO FIXED, LIVE NOT CLEARED`

Verified in repo:

- [web/public/favicon.ico](C:/Users/pault/OffGridFlow/web/public/favicon.ico)
- [web/app/layout.tsx](C:/Users/pault/OffGridFlow/web/app/layout.tsx:23)

Live result:

- `https://off-grid-flow.com/favicon.ico` still returns `404`.

Audit impact:

- Public polish/trust issue remains open until the live asset resolves.

### 3. Contact inconsistency

Status: `NOT CLEARED`

Verified fix:

- [web/app/api/elite-inquiry/route.ts](C:/Users/pault/OffGridFlow/web/app/api/elite-inquiry/route.ts:54) now routes to `contact@off-grid-flow.com`.

Remaining inconsistency:

- [web/components/EliteInquiryModal.tsx](C:/Users/pault/OffGridFlow/web/components/EliteInquiryModal.tsx:68) still sends `to: 'paul@offgridflow.com'`
- [web/components/EliteInquiryModal.tsx](C:/Users/pault/OffGridFlow/web/components/EliteInquiryModal.tsx:91) still tells users to email `paul@offgridflow.com`
- [web/components/EliteInquiryModal.tsx](C:/Users/pault/OffGridFlow/web/components/EliteInquiryModal.tsx:158) still displays `paul@offgridflow.com`

Audit impact:

- Contact-path consistency blocker remains active.

### 4. Offer / CTA consistency

Status: `NOT CLEARED LIVE`

Repo state is improved:

- [web/app/page.tsx](C:/Users/pault/OffGridFlow/web/app/page.tsx:68)
- [web/app/about/page.tsx](C:/Users/pault/OffGridFlow/web/app/about/page.tsx:21)
- [web/app/pricing/page.tsx](C:/Users/pault/OffGridFlow/web/app/pricing/page.tsx:114)
- [web/app/demo/page.tsx](C:/Users/pault/OffGridFlow/web/app/demo/page.tsx:84)

Live site still shows mixed CTA language:

- Home top nav still shows `Start Demo`.
- About still shows `Start Demo` and `Request a Demo`.
- Pricing still shows `Start Demo`; FAQ still says `Request a demo`.
- Demo top nav still shows `Start Trial`; bottom CTA says `Start Free Trial`.

Audit impact:

- Commercial path remains mixed, so this blocker is not closed.

### 5. Evidence-page discoverability

Status: `PARTIAL`

Repo state:

- [web/app/components/SiteNav.tsx](C:/Users/pault/OffGridFlow/web/app/components/SiteNav.tsx:20)
- [web/app/components/SiteNav.tsx](C:/Users/pault/OffGridFlow/web/app/components/SiteNav.tsx:26)
- [web/app/components/SiteNav.tsx](C:/Users/pault/OffGridFlow/web/app/components/SiteNav.tsx:29)

Live state:

- Home footer exposes `Methodology` and `Trust Center`.
- Demo page exposes `Trust Center`, `Security`, `Privacy Policy`, `Methodology`, and `Status`.
- Home top nav still does not expose `Methodology` or `Trust`.

Audit impact:

- This moves from `not met` to `partial`, but it does not satisfy the specific claim that those pages are reachable from the main navigation bar on every page.

## Net assessment

What is genuinely better on the live site:

- `/demo` is substantially more credible and product-aligned than the earlier fake-proof version.
- `/trust` and `/methodology` are live and useful.
- Pricing numbers are more internally coherent than the initial audit state.

What still blocks a meaningful score jump:

- Unsupported proof language is still present live and partly present in source.
- CTA language is still inconsistent live.
- Contact routing is still inconsistent in source.
- Favicon is still broken live.
- Railway CLI evidence still does not prove that the frontend deployment carrying these changes is live.

## What must happen before the next score increase

1. Deploy the current frontend build that contains the nav and CTA cleanup.
2. Remove every remaining `GHG Protocol-compliant` claim unless backed by explicit public proof.
3. Remove the stale `5,300+ companies` claim from the live home page.
4. Replace `paul@offgridflow.com` in `EliteInquiryModal.tsx`.
5. Make `/favicon.ico` return `200` on the public domain.
6. Re-run the audit only after the public site reflects those exact changes.

## Practical conclusion

This session shows real progress, but not enough to move the strict public audit score yet. The repo and the live site are out of sync, and at least one claimed blocker is still not actually fixed in source.
