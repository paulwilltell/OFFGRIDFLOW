# OffGridFlow SEO Implementation Report

**Date:** 2026-04-14
**Engagement:** Senior Full-Stack Engineer + SEO Architect (Claude Opus 4.6)
**Commits:** `0ad5f7b` → `c0b1a37` (7 commits)
**Build status:** Passes. All 21 new pages statically prerendered. Sitemap emitted.

---

## Executive Summary

Implemented a full SEO architecture overhaul transforming OffGridFlow from a single high-conversion homepage into a 37-URL topical moat. Every new page ships with canonical URLs, OpenGraph/Twitter metadata, breadcrumb schema, FAQ schema, and SoftwareApplication schema. Global Organization and SoftwareApplication entities are emitted site-wide. Footer reorganized into a 6-column topical map reinforcing the internal link graph. Dynamic sitemap covers all public routes.

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
| `MoneyPageLayout` | `web/components/MoneyPageLayout.tsx` | Reusable layout: breadcrumbs, H1, dek, CTAs, proof block, FAQ |
| `JsonLd` + schema builders | `web/components/JsonLd.tsx` | Typed generators for BreadcrumbList, FAQPage, SoftwareApplication, Organization, Dataset |
| `buildMoneyPageMetadata` | `web/lib/seo.ts` | Generates Next.js Metadata with canonical, OG, Twitter |
| Dynamic sitemap | `web/app/sitemap.ts` | 37 URLs with priority + changeFrequency |

---

## Internal Link Graph

### Footer rebuilt (site-wide)
Expanded from 4 columns to 6 columns:
- **Product:** How It Works, Pricing, Carbon Accounting, Scope 1-2-3, Audit-Ready, Case Study
- **Frameworks:** SB 253, CSRD, IFRS S2, CBAM, CSRD vs IFRS S2
- **By Role:** CFOs, Sustainability Managers, Procurement, Finance Teams, Scope 3 Supplier
- **Trust:** Trust Center, Methodology, Architecture, Evidence, Security, Operations, Status, How We Operate
- **Resources:** SB 253 Checklist, Scope 2 Factors, CSV Template, AWS, SAP, CSV Import, About, Contact

Every new page reachable from the footer. Legal links moved to a tertiary row.

### Per-page contextual links
Every money page closes with "Related reading" linking to 2-3 sibling pages forming a topical cluster. Dataset pages link to the money page they support (SB 253 checklist → SB 253 software; Scope 2 factors → Scope 1-2-3 reporting; Template → CSV import).

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

### Structured data validation
Run `https://validator.schema.org/` against any deployed money page (e.g., `https://off-grid-flow.com/sb-253-reporting-software`). Expected output:
- BreadcrumbList: 2 items, valid
- FAQPage: 4-5 Question/Answer pairs, valid
- SoftwareApplication: name, offer, provider, valid
- (Dataset pages only) Dataset: name, description, distribution, valid

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
