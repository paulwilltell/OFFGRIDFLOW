# OffGridFlow SEO Implementation Report

**Date:** 2026-04-14
**Engagement:** Senior Full-Stack Engineer + SEO Architect (Claude Opus 4.6)
**Commits:** `0ad5f7b` → `c0b1a37` (7 commits) plus Codex verification/remediation patch
**Build status:** Passes. All 21 new pages statically prerendered. Sitemap emitted.

---

## Executive Summary

Implemented a full SEO architecture overhaul transforming OffGridFlow from a single high-conversion homepage into a 37-URL topical moat. Every new page ships with canonical URLs, OpenGraph/Twitter metadata, breadcrumb schema, FAQ schema, and SoftwareApplication schema. Global Organization and SoftwareApplication entities are emitted site-wide. A shared footer and related-pages component now reinforce the internal link graph across the homepage, money pages, and trust surfaces. Dynamic sitemap covers all public routes.

Zero fabricated metrics. Zero unverifiable claims. Every money page links out to the methodology library, evidence pack, architecture, and status page so search results lead directly to verifiable proof.

---

## Pages Created

### 9 Money Pages (framework + role focused)
| Path | Primary query |
|---|---|
| `/carbon-accounting-software` | carbon accounting software |
| `/scope-1-2-3-reporting-software` | scope 1 2 3 reporting software |
| `/sb-253-reporting-software` | SB 253 reporting software |
| `/csrd-reporting-software` | CSRD reporting software |
| `/ifrs-s2-reporting-software` | IFRS S2 reporting software |
| `/cbam-reporting-software` | CBAM reporting software |
| `/scope-3-supplier-emissions-software` | scope 3 supplier emissions |
| `/audit-ready-carbon-accounting` | audit-ready carbon accounting |
| `/carbon-accounting-software-for-finance-teams` | carbon accounting for finance teams |

### 9 Query Trap Pages (comparisons + alternatives + roles + integrations)
| Path | Intent |
|---|---|
| `/csrd-vs-ifrs-s2-carbon-reporting` | comparison table |
| `/persefoni-alternative` | competitor alternative |
| `/watershed-alternative` | competitor alternative |
| `/for-cfos` | role landing |
| `/for-sustainability-managers` | role landing |
| `/for-procurement` | role landing |
| `/aws-carbon-data` | integration explainer |
| `/sap-carbon-reporting` | integration explainer |
| `/csv-emissions-import` | integration explainer |

### 3 Dataset Pages with Downloadable Resources
| Path | Download | Schema |
|---|---|---|
| `/sb-253-readiness-checklist` | `/downloads/sb-253-readiness-checklist.md` | Dataset + FAQPage |
| `/scope-2-factor-library` | `/downloads/scope-2-factor-library.csv` | Dataset + FAQPage |
| `/carbon-reporting-template` | `/downloads/carbon-reporting-template.csv` | Dataset + FAQPage |

**Total new pages: 21**
**Total sitemap entries: 37**

---

## Structured Data Implemented

| Schema | Coverage |
|---|---|
| `Organization` | Site-wide via root layout |
| `SoftwareApplication` | Site-wide + per-page via MoneyPageLayout |
| `BreadcrumbList` | Every money/trap/dataset page |
| `FAQPage` | Every page with FAQ section (all 21 new pages) |
| `Dataset` | 3 dataset pages |

All schemas emitted as `<script type="application/ld+json">` tags with unique ids to avoid collisions when multiple schemas coexist.

---

## Infrastructure Components

| Component | Path | Purpose |
|---|---|---|
| `MoneyPageLayout` | `web/components/MoneyPageLayout.tsx` | Reusable layout: breadcrumbs, H1, dek, CTAs, proof block, FAQ, related pages, shared footer |
| `RelatedPages` | `web/components/RelatedPages.tsx` | Shared internal-link block for primary money pages |
| `SiteFooter` | `web/app/components/SiteFooter.tsx` | Shared six-column public footer reinforcing framework, role, trust, and resource clusters |
| `JsonLd` + schema builders | `web/components/JsonLd.tsx` | Typed generators for BreadcrumbList, FAQPage, SoftwareApplication, Organization, Dataset |
| `buildMoneyPageMetadata` | `web/lib/seo.ts` | Generates Next.js Metadata with canonical, OG, Twitter |
| Dynamic sitemap | `web/app/sitemap.ts` | 37 URLs with priority + changeFrequency |

---

## Internal Link Graph

### Shared footer and topic map
Shared public footer now mounts from one source of truth:
- homepage
- all 21 money/query-trap/dataset pages through `MoneyPageLayout`
- methodology, evidence, security, and trust pages

Expanded to 6 topical columns:
- **Product:** How It Works, Pricing, Carbon Accounting, Scope 1-2-3, Audit-Ready, Case Study
- **Compliance Frameworks:** SB 253, CSRD, IFRS S2, CBAM, CSRD vs IFRS S2
- **By Role:** CFOs, Sustainability Managers, Procurement, Finance Teams, Scope 3 Supplier
- **Trust:** Trust Center, Methodology, Architecture, Evidence, Security, Operations, Status, How We Operate
- **Resources:** SB 253 Checklist, Scope 2 Factors, CSV Template, AWS, SAP, CSV Import, About, Contact

Every primary SEO page is reachable from the shared footer. Legal links remain in the lower utility row.

### Per-page contextual links
Every primary money page now mounts a shared `RelatedPages` block linking to 2-3 sibling pages forming a topical cluster. Dataset pages still link to the money page they support (SB 253 checklist → SB 253 software; Scope 2 factors → Scope 1-2-3 reporting; Template → CSV import).

### Breadcrumbs
Every new page ships with visible breadcrumbs (Home → Page Name) plus BreadcrumbList JSON-LD.

---

## Quality Controls Applied

| Control | Evidence |
|---|---|
| No fabricated metrics | Every claim ties to platform behavior or published sources (EPA, IEA, DEFRA, IPCC, GHG Protocol) |
| No invented testimonials | Pages contain zero quoted customers |
| No unverifiable certifications | SOC 2 Type II, ISO 27001 appear only as "planned Q1 2027 / Q2 2027" roadmap items — never as present-tense claims |
| Honest competitive positioning | Persefoni and Watershed alternative pages explicitly call out when those tools are the better fit |
| Consistent CTA | Every page uses "Start Free Trial" → `/register?plan=starter` with page-specific UTM tags |
| Million Fold Precision | Every page carries the drafts disclaimer: "OffGridFlow does not guarantee regulatory acceptance; customers are responsible for verification before submission" |

---

## Build Evidence

Last build output (commit `c0b1a37`):

- 21 new routes prerendered as static
- `/sitemap.xml` emits dynamically from `app/sitemap.ts`
- All new pages 241 B client JS (shared MoneyPageLayout, minimal per-page overhead)
- First Load JS shared 103 kB
- Zero new TypeScript errors from SEO work
- Zero build warnings from new pages

---

## Validation

### Local Playwright validation (2026-04-14)
Validated against a local production build served from `http://127.0.0.1:3006`.

#### Primary money pages
| Path | Title | H1 | JSON-LD count | Schema types |
|---|---|---|---:|---|
| `/carbon-accounting-software` | `Carbon Accounting Software \| OffGridFlow` | `Carbon Accounting Software That Survives Audit` | 5 | `Organization`, `SoftwareApplication`, `BreadcrumbList`, `FAQPage`, `SoftwareApplication` |
| `/scope-1-2-3-reporting-software` | `Scope 1, 2, 3 Reporting Software \| OffGridFlow` | `Scope 1, 2, and 3 Reporting Software` | 5 | `Organization`, `SoftwareApplication`, `BreadcrumbList`, `FAQPage`, `SoftwareApplication` |
| `/sb-253-reporting-software` | `California SB 253 Reporting Software \| OffGridFlow` | `California SB 253 Reporting Software` | 5 | `Organization`, `SoftwareApplication`, `BreadcrumbList`, `FAQPage`, `SoftwareApplication` |
| `/csrd-reporting-software` | `CSRD Reporting Software (ESRS E1) \| OffGridFlow` | `CSRD / ESRS E1 Reporting Software` | 5 | `Organization`, `SoftwareApplication`, `BreadcrumbList`, `FAQPage`, `SoftwareApplication` |
| `/ifrs-s2-reporting-software` | `IFRS S2 Climate Reporting Software \| OffGridFlow` | `IFRS S2 Climate Reporting Software` | 5 | `Organization`, `SoftwareApplication`, `BreadcrumbList`, `FAQPage`, `SoftwareApplication` |
| `/cbam-reporting-software` | `EU CBAM Reporting Software \| OffGridFlow` | `EU Carbon Border Adjustment Mechanism (CBAM) Reporting Software` | 5 | `Organization`, `SoftwareApplication`, `BreadcrumbList`, `FAQPage`, `SoftwareApplication` |
| `/scope-3-supplier-emissions-software` | `Scope 3 Supplier Emissions Software \| OffGridFlow` | `Scope 3 Supplier and Value Chain Emissions Software` | 5 | `Organization`, `SoftwareApplication`, `BreadcrumbList`, `FAQPage`, `SoftwareApplication` |
| `/audit-ready-carbon-accounting` | `Audit-Ready Carbon Accounting \| OffGridFlow` | `Audit-Ready Carbon Accounting` | 5 | `Organization`, `SoftwareApplication`, `BreadcrumbList`, `FAQPage`, `SoftwareApplication` |
| `/carbon-accounting-software-for-finance-teams` | `Carbon Accounting Software for Finance Teams \| OffGridFlow` | `Carbon Accounting Software for Finance Teams` | 5 | `Organization`, `SoftwareApplication`, `BreadcrumbList`, `FAQPage`, `SoftwareApplication` |

Representative full-page screenshot captured during this validation run:
- `seo-sb-253-reporting-software.png`

#### Dataset pages
| Path | Title | H1 | Schema types |
|---|---|---|---|
| `/sb-253-readiness-checklist` | `California SB 253 Readiness Checklist (Free Download) \| OffGridFlow` | `California SB 253 Readiness Checklist` | `Organization`, `SoftwareApplication`, `Dataset`, `BreadcrumbList`, `FAQPage`, `SoftwareApplication` |
| `/scope-2-factor-library` | `Scope 2 Emission Factor Library (Free CSV Download) \| OffGridFlow` | `Scope 2 Emission Factor Library` | `Organization`, `SoftwareApplication`, `Dataset`, `BreadcrumbList`, `FAQPage`, `SoftwareApplication` |
| `/carbon-reporting-template` | `Carbon Reporting CSV Template (Free Download) \| OffGridFlow` | `Carbon Reporting CSV Template` | `Organization`, `SoftwareApplication`, `Dataset`, `BreadcrumbList`, `FAQPage`, `SoftwareApplication` |

#### Console health
- Local validation console sweep returned `0` new errors on the SEO pages after navigation.

### Structured data validation
Run `https://validator.schema.org/` against any deployed money page (for example `https://off-grid-flow.com/sb-253-reporting-software`). Expected output:
- `BreadcrumbList`: 2 items, valid
- `FAQPage`: 4-5 Question/Answer pairs, valid
- `SoftwareApplication`: name, offer, provider, valid
- Dataset pages additionally emit `Dataset` with `distribution` metadata

### Google Search Console submission
After deployment:
1. Submit `https://off-grid-flow.com/sitemap.xml` to Search Console Sitemaps
2. Request indexing for each of the 9 money pages individually to accelerate crawl
3. Monitor "Enhancements" tab for FAQPage and BreadcrumbList rich result coverage

### Paid search integration
Every money page CTA includes `utm_source=organic&utm_medium=money_page&utm_campaign=<slug>`. When Google Ads campaigns launch targeting framework + software keywords, conversion attribution will show which pages produced paid trial starts vs organic.

---

## Recommended Next Steps

### Week 1
1. Submit sitemap to Google Search Console
2. Request indexing for 9 money pages
3. Launch Google Ads campaigns targeting exact-match keywords:
   - `[sb 253 reporting software]` → `/sb-253-reporting-software`
   - `[csrd reporting software]` → `/csrd-reporting-software`
   - `[carbon accounting software]` → `/carbon-accounting-software`
   - Budget $50/day per campaign for 30-day measurement window

### Week 2-4
4. Build 2-3 more dataset pages (Scope 3 factor library, CBAM product category guide, CSRD ESRS disclosure checklist)
5. Add schema.org `Article` markup to methodology library once it's split into standalone articles
6. Monitor Search Console for query impressions, click-through rate, and rich result appearances

### Month 2-3
7. Commission 1-2 named customer case studies to replace the illustrative case study with real named evidence
8. Expand the alternative comparison set (Watershed, Persefoni + Normative, Sphera, Emitwise)
9. Run Lighthouse SEO audit on each money page, target 95+

---

## Summary Statement

SEO architecture implementation complete. The site now has 9 money pages, 9 query trap pages, 3 dataset pages with Schema.org Dataset markup, global Organization/SoftwareApplication entity schema, and a reinforced internal link graph spanning footer, homepage, and contextual related-page links. Build passes cleanly. Ready for deployment, Search Console monitoring, and paid search campaigns.
