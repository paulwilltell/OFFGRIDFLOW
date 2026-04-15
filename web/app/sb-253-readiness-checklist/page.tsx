import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { JsonLd, datasetSchema } from '@/components/JsonLd';
import { buildMoneyPageMetadata, SITE_URL } from '@/lib/seo';

const PATH = '/sb-253-readiness-checklist';
const DOWNLOAD_URL = `${SITE_URL}/downloads/sb-253-readiness-checklist.md`;

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'California SB 253 Readiness Checklist (Free Download)',
  description:
    'Free SB 253 readiness checklist covering scope determination, governance, methodology, data collection, QA, assurance, and filing. Downloadable Markdown/PDF.',
  keyword: 'SB 253 readiness checklist',
});

const faqs = [
  {
    question: 'Is this checklist free to use?',
    answer:
      'Yes. Download it, adapt it, and share it internally. Attribution to OffGridFlow is appreciated but not required. The checklist is not legal advice and does not substitute for consulting CARB guidance.',
  },
  {
    question: 'What format is the download?',
    answer:
      'Markdown for easy editing in your team\'s documentation system. Open it in any text editor, or convert to PDF/DOCX using standard tooling. A PDF mirror will be added as demand warrants.',
  },
  {
    question: 'How does the checklist map to OffGridFlow?',
    answer:
      'Each section references an artifact a carbon accounting tool should produce (methodology version, calculation ledger, factor snapshot, approval trail, export reconciliation). OffGridFlow produces all of them. The checklist helps you evaluate any tool, not just ours.',
  },
];

export default function SB253ChecklistPage() {
  return (
    <>
      <JsonLd
        id="ld-dataset-sb253-checklist"
        data={datasetSchema({
          name: 'California SB 253 Readiness Checklist',
          description:
            'Practical checklist for companies preparing California SB 253 (Climate Corporate Data Accountability Act) disclosure. Covers scope determination, governance, methodology, data collection, QA, assurance, and filing.',
          url: `${SITE_URL}${PATH}`,
          distributionUrl: DOWNLOAD_URL,
          encodingFormat: 'text/markdown',
          keywords: ['SB 253', 'climate disclosure', 'CARB', 'GHG Protocol', 'assurance readiness'],
        })}
      />
      <MoneyPageLayout
        breadcrumbs={[
          { name: 'Home', path: '/' },
          { name: 'SB 253 Readiness Checklist', path: PATH },
        ]}
        h1="California SB 253 Readiness Checklist"
        dek="A practical, eight-section checklist for companies preparing their first SB 253 disclosure. Free to download, adapt, and share internally."
        slug="sb-253-checklist"
        ctaUtm="sb253_checklist"
        ctaText="See SB 253 Software"
        secondaryCtaText="Download checklist"
        secondaryCtaHref="/downloads/sb-253-readiness-checklist.md"
        faqs={faqs}
      >
        <div className="not-prose my-8 rounded-2xl border border-primary-600/30 bg-primary-600/5 p-6">
          <h2 className="text-lg font-semibold text-white">Download the checklist</h2>
          <p className="mt-2 text-sm text-gray-400">
            Markdown source, ~8 sections, no email required, no signup gate. Use it to evaluate
            your current carbon reporting tooling or to audit the gaps before engaging an
            assurance provider.
          </p>
          <div className="mt-4 flex flex-wrap gap-3">
            <a
              href="/downloads/sb-253-readiness-checklist.md"
              download
              className="rounded-lg bg-primary-600 px-5 py-2.5 text-sm font-semibold text-white hover:bg-primary-500"
            >
              Download Markdown
            </a>
            <Link
              href="/sb-253-reporting-software"
              className="rounded-lg border border-gray-700 px-5 py-2.5 text-sm text-gray-300 hover:border-gray-500 hover:text-white"
            >
              See SB 253 software
            </Link>
          </div>
        </div>

        <h2 className="mt-10 text-2xl font-bold text-white">What the checklist covers</h2>
        <ol className="list-decimal list-inside space-y-1">
          <li>Scope determination (revenue, CA nexus, subsidiaries)</li>
          <li>Governance and ownership (executive sponsor, internal controls)</li>
          <li>Methodology selection (GHG Protocol alignment, Scope 2 method choice, Scope 3 tier)</li>
          <li>Data collection (Scope 1 sources, Scope 2 facilities, Scope 3 by category)</li>
          <li>Calculation infrastructure (ledger, snapshot, change log, approval)</li>
          <li>Quality assurance (anomaly detection, data quality tier tagging)</li>
          <li>Assurance readiness (provider selection, evidence pack format)</li>
          <li>Filing readiness (draft review, board sign-off, CARB submission)</li>
        </ol>

        <h2 className="mt-10 text-2xl font-bold text-white">Why this checklist exists</h2>
        <p>
          The conversation about SB 253 tends to focus on scope and timeline. The harder problem
          is what you need before the assurance provider walks in: immutable calculation ledger,
          locked factor snapshot, documented methodology, approval trail. Every bullet in the
          checklist represents something assurance providers will ask for — often in writing,
          with reconciliation expectations that mirror financial audits.
        </p>

        <p className="mt-8 text-sm text-gray-500">
          Related:{' '}
          <Link href="/sb-253-reporting-software" className="text-primary-400 hover:underline">SB 253 reporting software</Link>
          {' · '}
          <Link href="/audit-ready-carbon-accounting" className="text-primary-400 hover:underline">Audit-ready carbon accounting</Link>
          {' · '}
          <Link href="/evidence" className="text-primary-400 hover:underline">Sample evidence pack</Link>
        </p>
      </MoneyPageLayout>
    </>
  );
}
