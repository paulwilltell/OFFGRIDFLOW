import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/for-cfos';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'OffGridFlow for CFOs',
  description:
    'Carbon accounting built for CFO ownership. Ledger-style audit trail, framework-coverage (SEC, CSRD, SB 253, IFRS S2), reconciliation checksums, and defensible methodology.',
  keyword: 'carbon accounting for CFOs',
});

const faqs = [
  {
    question: 'Why is carbon accounting now a CFO responsibility?',
    answer:
      'SEC Climate Disclosure puts Scope 1 and 2 (and material Scope 3) in the 10-K. CSRD puts emissions in the annual report. SB 253 requires limited and eventually reasonable assurance. When climate data lives in financial filings, the CFO inherits the restatement risk and the internal controls obligation.',
  },
  {
    question: 'What controls does OffGridFlow provide?',
    answer:
      'RBAC with admin/user/viewer separation of duties, approval workflow with preparer/reviewer/approver attribution, immutable calculation ledger, change log with field-level before/after tracking, factor snapshot lock per reporting period, and export reconciliation checksum. Documented in the public Trust Center.',
  },
  {
    question: 'How does this fit into the monthly close?',
    answer:
      'Activity ingestion (CSV or SAP/ERP connector) runs on a schedule. Anomaly detection flags outliers before they hit the dashboard. Month-end: review anomalies, approve draft, lock snapshot. Year-end: publish full disclosures from locked snapshots.',
  },
  {
    question: 'What is the cost vs Big 4?',
    answer:
      'OffGridFlow starts at $6,500/year. Typical Big 4 carbon audit engagements run $50,000-$200,000. The savings fund internal controls improvements and external assurance instead of one-off measurement exercises.',
  },
];

export default function ForCfosPage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'For CFOs', path: PATH },
      ]}
      h1="OffGridFlow for CFOs"
      dek="Climate disclosure is now financial disclosure. OffGridFlow gives finance leadership the ledger-grade audit trail, framework coverage, and published methodology needed to defend the emissions line alongside the audited financials."
      slug="for-cfos"
      ctaUtm="for_cfos"
      faqs={faqs}
      showLeadForm
    >
      <p>
        When Scope 1, 2, and 3 emissions appear on the 10-K, the annual report, and the SB 253
        filing, the responsibility shifts from the sustainability team alone to the finance
        leadership. The same controls that protect the financial statements must now protect
        the emissions disclosure.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">What a CFO gets from OffGridFlow</h2>
      <ul>
        <li>Ledger-style data model mirroring the general ledger discipline</li>
        <li>Factor snapshot lock per reporting period to prevent silent restatements</li>
        <li>Approval workflow with separation-of-duty enforcement at the role layer</li>
        <li>Export reconciliation checksum matching PDF output to the ledger at export time</li>
        <li>Multi-framework draft output from one source of truth (SEC, CSRD, SB 253, CBAM, IFRS S2)</li>
        <li>Downloadable assurance pack for external auditor engagement</li>
        <li>Published methodology version — v2026.1.0 — for auditor reference</li>
      </ul>

      <h2 className="mt-10 text-2xl font-bold text-white">The carbon close, monthly</h2>
      <p>
        Run carbon close on the same rhythm as financial close. Ingest activities, flag
        anomalies, review drafts, approve, lock snapshot. Move from reactive year-end panic to
        continuous disclosure readiness.
      </p>

      <p className="mt-8 text-sm text-gray-500">
        <Link href="/carbon-accounting-software-for-finance-teams" className="text-primary-400 hover:underline">Finance-team workflow</Link>
        {' · '}
        <Link href="/audit-ready-carbon-accounting" className="text-primary-400 hover:underline">Audit-ready carbon accounting</Link>
        {' · '}
        <Link href="/architecture" className="text-primary-400 hover:underline">Data architecture</Link>
      </p>
    </MoneyPageLayout>
  );
}
