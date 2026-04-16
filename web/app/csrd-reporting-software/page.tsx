import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/csrd-reporting-software';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'CSRD Reporting Software (ESRS E1)',
  description:
    'Software for CSRD / ESRS E1 climate reporting. Scope 1, 2, 3 emissions calculated, materiality-aligned Scope 3 categories, XBRL tagging, audit-ready exports.',
  keyword: 'CSRD reporting software',
});

const faqs = [
  {
    question: 'What is CSRD / ESRS E1?',
    answer:
      'The Corporate Sustainability Reporting Directive (CSRD) is the EU sustainability disclosure framework, implemented through the European Sustainability Reporting Standards (ESRS). ESRS E1 is the climate change standard. It requires double materiality assessment, Scope 1/2/3 disclosure, transition plans, and assurance.',
  },
  {
    question: 'Who has to report under CSRD?',
    answer:
      'Large EU companies and non-EU companies with EU operations above revenue and employee thresholds, phased in between 2024 and 2028. Wave 1 (largest EU public interest entities) reports first, then large undertakings, then listed SMEs.',
  },
  {
    question: 'Does OffGridFlow produce XBRL output for CSRD?',
    answer:
      'Yes. CSRD requires machine-readable tagging using ESRS XBRL taxonomies. OffGridFlow exports both PDF and XBRL/iXBRL. The output is a draft; your filing provider or auditor conducts final review.',
  },
  {
    question: 'How do you support double materiality?',
    answer:
      'OffGridFlow captures category-level Scope 3 data (15 GHG Protocol categories) so you can apply your double materiality assessment and disclose the categories that are material to your impact or financial position. The tool does not perform the materiality assessment itself — that is your disclosure responsibility.',
  },
  {
    question: 'Does the disclosure require assurance?',
    answer:
      'Yes. CSRD mandates limited assurance, moving toward reasonable assurance. OffGridFlow&apos;s evidence pack (calculation ledger, factor snapshot, approval trail, export checksum) is designed for assurance provider workflows.',
  },
];

export default function CsrdPage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'CSRD Reporting Software', path: PATH },
      ]}
      h1="CSRD / ESRS E1 Reporting Software"
      dek="OffGridFlow calculates Scope 1, 2, and all 15 Scope 3 categories, locks the methodology to the reporting period, and exports CSRD-ready PDF and XBRL drafts for your auditor."
      slug="csrd-reporting-software"
      ctaUtm="csrd_reporting"
      faqs={faqs}
      showLeadForm
      leadFormFramework="csrd"
      relatedPages={[
        { href: '/csrd-vs-ifrs-s2-carbon-reporting', label: 'CSRD vs IFRS S2', description: 'Compare double materiality with financial materiality.' },
        { href: '/scope-1-2-3-reporting-software', label: 'Scope 1, 2, 3 reporting', description: 'Understand the underlying emissions workflow.' },
        { href: '/for-sustainability-managers', label: 'For sustainability managers', description: 'See the operating model for ongoing reporting cycles.' },
      ]}
    >
      <p>
        CSRD is the most comprehensive sustainability reporting regime ever imposed on EU
        companies. ESRS E1 alone requires: double materiality assessment, Scope 1/2/3 emissions
        across the full value chain, transition plan alignment with the Paris Agreement, internal
        carbon pricing disclosure where used, and assurance by an independent provider. All in
        machine-readable XBRL tagging.
      </p>

      <p>
        OffGridFlow handles the emissions calculation, factor lock, approval workflow, and XBRL
        export. You handle the narrative disclosures, materiality assessment, and final auditor
        review.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">What OffGridFlow produces for CSRD</h2>
      <ul>
        <li>Scope 1 direct emissions by source type</li>
        <li>Scope 2 location-based AND market-based (both required)</li>
        <li>Scope 3 by all 15 GHG Protocol categories with materiality flags</li>
        <li>Factor snapshot locked to the reporting period for reproducibility</li>
        <li>Draft ESRS E1 XBRL export ready for your filing provider</li>
        <li>Evidence pack for limited assurance engagement</li>
      </ul>

      <h2 className="mt-10 text-2xl font-bold text-white">The wave-by-wave rollout</h2>
      <p>
        Wave 1 (EU public interest entities with &gt;500 employees) files first. Wave 2 (large
        undertakings above two of three thresholds: &gt;250 employees, &gt;€50M revenue, &gt;€25M
        balance sheet) follows. Non-EU companies with &gt;€150M EU revenue and at least one EU
        subsidiary or branch are pulled in last. If any wave applies to you, the reporting year
        is already underway — data collection and methodology lock need to happen now, not
        post-year-end.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">Why teams use OffGridFlow for CSRD</h2>
      <ul>
        <li>Methodology is public and versioned — auditors can diff across reporting years</li>
        <li>Factor provenance traces to EPA, IEA, DEFRA, IPCC, and GHG Protocol</li>
        <li>Immutable ledger survives multi-year restatement pressure</li>
        <li>Self-serve onboarding means the assurance team sees the data before the audit starts</li>
      </ul>

    </MoneyPageLayout>
  );
}
