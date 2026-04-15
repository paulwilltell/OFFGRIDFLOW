import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/sb-253-reporting-software';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'California SB 253 Reporting Software',
  description:
    'Software for California SB 253 (Climate Corporate Data Accountability Act) reporting. Scope 1, 2, and 3 emissions calculated, locked, and exported for CARB-ready disclosure.',
  keyword: 'SB 253 reporting software',
});

const faqs = [
  {
    question: 'Who must comply with California SB 253?',
    answer:
      'The Climate Corporate Data Accountability Act (SB 253) applies to U.S. companies doing business in California with annual revenue above $1 billion. Scope 1 and 2 disclosure is required starting in the 2026 reporting year; Scope 3 follows in 2027. Deadlines are set by the California Air Resources Board (CARB).',
  },
  {
    question: 'How does OffGridFlow prepare a SB 253 disclosure?',
    answer:
      'OffGridFlow calculates Scope 1 and 2 using GHG Protocol-aligned methods, applies location-based factors from EPA eGRID, captures Scope 3 across all 15 categories, and produces a draft filing package that maps to CARB requirements. Every calculation is stored in an immutable ledger for third-party assurance.',
  },
  {
    question: 'Does SB 253 require third-party assurance?',
    answer:
      'Yes. SB 253 requires limited assurance for Scope 1 and 2 initially, moving to reasonable assurance later. OffGridFlow&apos;s calculation ledger, factor snapshots, and evidence pack are designed to satisfy assurance provider requests without re-running calculations.',
  },
  {
    question: 'What happens if emission factors change between reporting years?',
    answer:
      'Lock a factor snapshot to the 2026 reporting period. Even if EPA eGRID 2024 or 2025 factors are released later, your 2026 disclosure remains reproducible. This is essential for consistent year-over-year comparisons SB 253 requires.',
  },
  {
    question: 'Is OffGridFlow a substitute for a CARB consultant?',
    answer:
      'OffGridFlow is calculation and reporting software. The output is a draft disclosure you or your consultants review before submission. We do not guarantee CARB acceptance — customers remain responsible for final verification and filing.',
  },
];

export default function Sb253Page() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'SB 253 Reporting Software', path: PATH },
      ]}
      h1="California SB 253 Reporting Software"
      dek="OffGridFlow calculates Scope 1, 2, and 3 emissions, locks them to the reporting period, and produces a CARB-ready draft disclosure package — without the six-figure consultant engagement."
      slug="sb-253-reporting-software"
      ctaUtm="sb253_reporting"
      faqs={faqs}
    >
      <p>
        California Senate Bill 253 (Climate Corporate Data Accountability Act) makes annual
        greenhouse gas disclosure mandatory for companies doing business in California with over
        $1 billion in annual revenue. First Scope 1 and 2 disclosures cover fiscal year 2026.
        Scope 3 follows. Limited assurance is required initially, escalating to reasonable
        assurance.
      </p>

      <p>
        The challenge is not the spreadsheet. It is reproducibility. A SB 253 filing must be
        reconstructible by a third-party assurance provider next year, the year after, and
        whenever CARB audits. That requires locked factor versions, an immutable ledger, and an
        approval trail — exactly what OffGridFlow is built for.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">How OffGridFlow handles SB 253</h2>
      <ul>
        <li>
          <strong className="text-white">Scope 1:</strong> stationary and mobile combustion via
          activity data + EPA factors; refrigerant fugitives via IPCC AR6 GWP-100.
        </li>
        <li>
          <strong className="text-white">Scope 2:</strong> location-based using EPA eGRID 2023
          subregional factors; market-based from supplier-specific instruments.
        </li>
        <li>
          <strong className="text-white">Scope 3:</strong> all 15 GHG Protocol categories,
          starting with spend-based for rapid coverage and upgrading to supplier-specific as
          your value chain data matures.
        </li>
        <li>
          <strong className="text-white">Factor snapshot:</strong> lock the 2026 calculation
          basis so your disclosure remains reproducible after future factor releases.
        </li>
        <li>
          <strong className="text-white">Assurance pack:</strong> downloadable evidence pack with
          source activity records, factor provenance, calculation ledger, approval trail, and
          export checksum.
        </li>
      </ul>

      <h2 className="mt-10 text-2xl font-bold text-white">Why finance teams pick OffGridFlow for SB 253</h2>
      <p>
        Assurance providers expect the same rigor they see in financial audits: tie every
        disclosed number to source evidence, document the methodology applied, and prove the
        calculation has not been altered. OffGridFlow surfaces the activity → factor → formula →
        result chain for every reported figure, in seconds.
      </p>

      <p className="mt-8 text-sm text-gray-500">
        Related reading:{' '}
        <Link href="/sb-253-readiness-checklist" className="text-primary-400 hover:underline">SB 253 readiness checklist</Link>
        {' · '}
        <Link href="/audit-ready-carbon-accounting" className="text-primary-400 hover:underline">Audit-ready carbon accounting</Link>
        {' · '}
        <Link href="/carbon-accounting-software-for-finance-teams" className="text-primary-400 hover:underline">Carbon accounting for finance teams</Link>
      </p>
    </MoneyPageLayout>
  );
}
