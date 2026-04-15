import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/for-sustainability-managers';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'OffGridFlow for Sustainability Managers',
  description:
    'Defensible carbon measurement without consultants. OffGridFlow gives sustainability leads a tool that finance, audit, and legal can all verify — without the six-figure invoice.',
  keyword: 'carbon software for sustainability managers',
});

const faqs = [
  {
    question: 'How does OffGridFlow help win internal budget?',
    answer:
      'Published methodology, public sample audit packet, and transparent pricing ($6,500-$15,000) make the procurement case straightforward. You can show finance and legal exactly what they are getting before any contract is signed — something most carbon tools cannot offer.',
  },
  {
    question: 'Can I move from spreadsheets without losing history?',
    answer:
      'Yes. CSV import handles historical data uploads, including multi-year series. The calculation ledger stamps every imported period with the factor snapshot in effect so prior-year data remains reproducible.',
  },
  {
    question: 'What happens when I need to hand off to finance?',
    answer:
      'Role-based access (admin/user/viewer) lets you invite finance, compliance, and audit to the same workspace with appropriate permissions. The audit trail means you do not have to be in the room every time someone asks where a number came from.',
  },
  {
    question: 'How do I handle Scope 3 when suppliers do not respond?',
    answer:
      'Start with spend-based for full coverage, then upgrade individual categories to activity-based or supplier-specific as data matures. Each calculation tags its data quality tier so your disclosure accurately reflects the underlying evidence.',
  },
];

export default function ForSustainabilityManagersPage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'For Sustainability Managers', path: PATH },
      ]}
      h1="OffGridFlow for Sustainability Managers"
      dek="Replace the spreadsheet, defend the numbers, and bring finance and audit into the same workflow — without bringing in a consultant for every question."
      slug="for-sustainability-managers"
      ctaUtm="for_sustainability"
      faqs={faqs}
    >
      <p>
        Sustainability leads carry the weight of emissions measurement until the first
        regulator, auditor, or board member asks how a number was produced. At that point the
        spreadsheet answer stops working. OffGridFlow is the defensible workflow that turns
        "trust me" into a reproducible ledger.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">What changes once OffGridFlow is in place</h2>
      <ul>
        <li>Finance and audit invited to the same workspace with view-only roles</li>
        <li>Every calculation becomes a lineage trace: activity → factor → formula → ledger entry</li>
        <li>Scope 3 supplier requests routed through the platform, not email</li>
        <li>Draft reports produced for CSRD, SEC, SB 253, CBAM, and IFRS S2 with one dataset</li>
        <li>Anomaly detection surfaces data quality issues before they become audit findings</li>
      </ul>

      <h2 className="mt-10 text-2xl font-bold text-white">Getting internal buy-in</h2>
      <p>
        Finance wants a ledger. Audit wants an immutable trail. Legal wants a documented
        methodology and a limited liability posture. OffGridFlow publishes all three publicly
        so procurement conversations finish faster.
      </p>

      <p className="mt-8 text-sm text-gray-500">
        <Link href="/carbon-accounting-software" className="text-primary-400 hover:underline">Carbon accounting software</Link>
        {' · '}
        <Link href="/scope-3-supplier-emissions-software" className="text-primary-400 hover:underline">Scope 3 supplier emissions</Link>
        {' · '}
        <Link href="/evidence" className="text-primary-400 hover:underline">Sample evidence pack</Link>
      </p>
    </MoneyPageLayout>
  );
}
