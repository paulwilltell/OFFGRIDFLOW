import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/audit-ready-carbon-accounting';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'Audit-Ready Carbon Accounting',
  description:
    'Carbon accounting built for third-party assurance. Versioned methodology, immutable calculation ledger, factor snapshots, and approval trails survive assurance and restatement.',
  keyword: 'audit-ready carbon accounting',
});

const faqs = [
  {
    question: 'What makes a carbon calculation "audit-ready"?',
    answer:
      'An audit-ready calculation can be reconstructed by a third-party assurance provider without access to the original user session. That requires: the activity data that produced the result, the factor used (including source, region, vintage), the method applied, the formula, the user who executed, and the timestamp. OffGridFlow records all of these in the calculation ledger, immutably.',
  },
  {
    question: 'How do factor snapshots work?',
    answer:
      'When you lock a reporting period, OffGridFlow freezes the factor set in use and records the snapshot ID on every subsequent calculation. Even if EPA eGRID publishes new factors next month, your 2026 disclosure still reproduces exactly because the snapshot is immutable.',
  },
  {
    question: 'Can we export the evidence pack to our auditor?',
    answer:
      'Yes. The Data Governance page produces a full JSON export including activities, calculation ledger entries, and change log. The Evidence page shows a redacted sample pack so you can see what an assurance provider will receive.',
  },
  {
    question: 'Does OffGridFlow perform the assurance engagement?',
    answer:
      'No. OffGridFlow is calculation and reporting software. Third-party assurance is performed by an independent qualified auditor. OffGridFlow&apos;s job is to make their work fast and cheap by producing artifacts that match how assurance providers already work.',
  },
];

export default function AuditReadyPage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'Audit-Ready Carbon Accounting', path: PATH },
      ]}
      h1="Audit-Ready Carbon Accounting"
      dek="Every emission number traces to its activity, factor, formula, actor, and timestamp. Factor snapshots lock the methodology. Export reconciliation proves the PDF matches the ledger."
      slug="audit-ready-carbon-accounting"
      ctaUtm="audit_ready"
      faqs={faqs}
      relatedPages={[
        { href: '/architecture', label: 'Architecture & traceability chain', description: 'Inspect the underlying ledger and factor snapshot model.' },
        { href: '/evidence', label: 'Sample evidence pack', description: 'Review a redacted audit packet end to end.' },
        { href: '/sb-253-reporting-software', label: 'SB 253 reporting', description: 'See one of the highest-assurance disclosure use cases.' },
      ]}
    >
      <p>
        Auditor-ready carbon accounting means the same thing as auditor-ready financial
        accounting: every disclosed number has a provenance that an independent reviewer can
        reconstruct. OffGridFlow models emissions data the way a general ledger models
        transactions — append-only, time-stamped, actor-attributed, and immutable once posted.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">The primitives that make it work</h2>
      <ul>
        <li>
          <strong className="text-white">Calculation ledger:</strong> one row per calculation,
          capturing the activity id, factor id, factor value, source, region, method, formula,
          result, calculated-by user, and calculated-at timestamp.
        </li>
        <li>
          <strong className="text-white">Factor snapshot:</strong> a JSON-frozen copy of the
          factor set locked to a reporting period. Calculations tagged with a snapshot are
          reproducible even after new factor vintages are released.
        </li>
        <li>
          <strong className="text-white">Approval workflow:</strong> draft → submitted → reviewed
          → approved, each transition time-stamped and user-attributed. Approved reports are
          locked against further modification.
        </li>
        <li>
          <strong className="text-white">Change log:</strong> field-level modification history
          with before/after values and actor attribution.
        </li>
        <li>
          <strong className="text-white">Export reconciliation:</strong> SHA-256 checksum of
          scope totals at export time. Assurance providers can verify the PDF they received
          matches the ledger at the time of generation.
        </li>
      </ul>

      <h2 className="mt-10 text-2xl font-bold text-white">How an assurance engagement runs</h2>
      <ol className="list-decimal list-inside space-y-1">
        <li>Customer generates the compliance report and locks the factor snapshot</li>
        <li>Customer exports the JSON evidence pack from the Data Governance page</li>
        <li>Assurance provider walks samples from report → ledger → activity → factor source</li>
        <li>Provider spot-checks change log and approval trail for any late modifications</li>
        <li>Export reconciliation confirms final numbers match the locked ledger</li>
      </ol>

    </MoneyPageLayout>
  );
}
