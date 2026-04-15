import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/ifrs-s2-reporting-software';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'IFRS S2 Climate Reporting Software',
  description:
    'Software for IFRS S2 climate-related disclosures. Scope 1, 2, 3 emissions, financed emissions guidance, TCFD alignment, audit-ready evidence pack.',
  keyword: 'IFRS S2 reporting software',
});

const faqs = [
  {
    question: 'What is IFRS S2?',
    answer:
      'IFRS S2 is the International Sustainability Standards Board (ISSB) climate-related disclosure standard. It builds on TCFD and is being adopted by jurisdictions including the UK, Canada, Singapore, Australia, and many emerging markets. It requires Scope 1/2/3 disclosure, climate risk and opportunity assessment, scenario analysis, and transition plan disclosure.',
  },
  {
    question: 'How does IFRS S2 differ from CSRD?',
    answer:
      'IFRS S2 is a financial-materiality-only standard; CSRD requires double materiality (financial + impact). IFRS S2 maps more directly to financial statements, while CSRD demands broader sustainability context. Many companies must report under both.',
  },
  {
    question: 'Does OffGridFlow handle financed emissions?',
    answer:
      'OffGridFlow calculates Scope 3 Category 15 (Investments) using published PCAF methodology where activity data is available. For asset-class-specific financed emissions, we flag which tier of PCAF data quality each calculation represents so your disclosure matches your actual data granularity.',
  },
  {
    question: 'Is the output aligned with TCFD?',
    answer:
      'IFRS S2 supersedes TCFD by incorporating its governance/strategy/risk/metrics framework. OffGridFlow&apos;s metrics and targets section produces TCFD-aligned outputs automatically when you generate an IFRS S2 draft.',
  },
  {
    question: 'Can I use OffGridFlow if I also report under SEC Climate rules?',
    answer:
      'Yes. Scope 1 and 2 calculations reuse the same activity data and ledger for SEC, CSRD, SB 253, CBAM, and IFRS S2. You reconcile once, disclose many times — with each framework&apos;s specific packaging.',
  },
];

export default function IfrsS2Page() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'IFRS S2 Reporting Software', path: PATH },
      ]}
      h1="IFRS S2 Climate Reporting Software"
      dek="Produce IFRS S2-ready Scope 1, 2, and 3 disclosures with a locked methodology, traceable calculation ledger, and TCFD-aligned metrics output."
      slug="ifrs-s2-reporting-software"
      ctaUtm="ifrs_s2_reporting"
      faqs={faqs}
      relatedPages={[
        { href: '/csrd-vs-ifrs-s2-carbon-reporting', label: 'CSRD vs IFRS S2', description: 'Compare adoption patterns, assurance, and filing differences.' },
        { href: '/scope-1-2-3-reporting-software', label: 'Scope 1, 2, 3 reporting', description: 'Review the underlying emissions data model.' },
        { href: '/methodology', label: 'Methodology library', description: 'Inspect factor sources and standards alignment.' },
      ]}
    >
      <p>
        IFRS S2 is the ISSB&apos;s climate-related financial disclosure standard. Adopted or
        planned for adoption in the UK, Canada, Singapore, Australia, Brazil, Hong Kong, and many
        other jurisdictions, it is becoming the de facto global baseline for climate reporting
        alongside CSRD in the EU and SEC Climate Disclosure in the U.S.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">What IFRS S2 requires</h2>
      <ul>
        <li>Governance over climate-related risks and opportunities</li>
        <li>Strategy for managing climate risks, including scenario analysis</li>
        <li>Risk management processes</li>
        <li>Metrics and targets: Scope 1, 2, 3 emissions, financed emissions for financials, internal carbon pricing where used</li>
      </ul>

      <h2 className="mt-10 text-2xl font-bold text-white">What OffGridFlow contributes</h2>
      <p>
        OffGridFlow owns the metrics and targets layer. The calculation engine produces the
        Scope 1/2/3 figures; the calculation ledger produces the audit trail; factor snapshots
        lock the methodology for consistent year-over-year restatement handling.
      </p>

      <ul>
        <li>Scope 1 and 2 per GHG Protocol, location-based and market-based</li>
        <li>All 15 Scope 3 categories including Investments (PCAF-aligned where data permits)</li>
        <li>Factor snapshot for reproducibility across restatement periods</li>
        <li>XBRL-ready draft export for jurisdictions that require machine-readable filing</li>
        <li>Evidence pack for assurance providers</li>
      </ul>

    </MoneyPageLayout>
  );
}
