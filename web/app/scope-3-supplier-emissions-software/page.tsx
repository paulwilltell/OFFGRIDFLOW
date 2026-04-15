import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/scope-3-supplier-emissions-software';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'Scope 3 Supplier Emissions Software',
  description:
    'Collect, calculate, and disclose Scope 3 emissions across all 15 GHG Protocol categories. Spend-based to supplier-specific tier progression with PCAF-aligned data quality flags.',
  keyword: 'scope 3 supplier emissions software',
});

const faqs = [
  {
    question: 'How do you handle suppliers that do not report emissions data?',
    answer:
      'Start with spend-based calculations using EEIO coefficients (documented in the methodology library). As suppliers report primary data, upgrade those categories to activity-based or supplier-specific. Each calculation is tagged with its data quality tier so your disclosure matches reality.',
  },
  {
    question: 'Which Scope 3 categories are supported?',
    answer:
      'All 15 GHG Protocol Corporate Value Chain categories: purchased goods and services, capital goods, fuel and energy activities, upstream transport, waste, business travel, employee commuting, upstream leased assets, downstream transport, processing of sold products, use of sold products, end-of-life treatment, downstream leased assets, franchises, and investments.',
  },
  {
    question: 'Can OffGridFlow ingest supplier data automatically?',
    answer:
      'Yes, via CSV upload or direct SAP/ERP connector. Supplier-reported data can be mapped to specific Scope 3 categories and tied to the source record for traceability. CDP-style supplier response imports are supported.',
  },
  {
    question: 'How do you flag data quality for Scope 3?',
    answer:
      'Each calculation records its data quality (measured, estimated, supplier-specific, spend-based, default). Aggregate reports show the distribution of data quality tiers so stakeholders see the basis of the disclosure. CSRD, SEC, and SB 253 all require this transparency.',
  },
];

export default function Scope3SupplierPage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'Scope 3 Supplier Emissions', path: PATH },
      ]}
      h1="Scope 3 Supplier and Value Chain Emissions Software"
      dek="Cover all 15 Scope 3 categories. Start with spend-based, upgrade to supplier-specific. Data quality tagged on every calculation so disclosure matches the evidence."
      slug="scope-3-supplier-emissions-software"
      ctaUtm="scope_3_supplier"
      faqs={faqs}
    >
      <p>
        Scope 3 is where carbon accounting gets hard. Activity data sits in supplier emails,
        procurement spreadsheets, and invoices. The GHG Protocol Scope 3 Standard covers 15
        categories with radically different data collection patterns. SEC Climate, CSRD, and SB
        253 all require material Scope 3 disclosure by category.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">The tier progression</h2>
      <ul>
        <li>
          <strong className="text-white">Spend-based:</strong> multiply supplier spend by industry
          EEIO coefficient. Broad coverage, low accuracy. Good starting point.
        </li>
        <li>
          <strong className="text-white">Activity-based:</strong> multiply a physical quantity
          (kWh, tonne-km, kg shipped) by a published factor. Better accuracy than spend-based.
        </li>
        <li>
          <strong className="text-white">Supplier-specific:</strong> use primary data provided by
          the supplier. Highest accuracy; required for material categories under reasonable
          assurance.
        </li>
      </ul>

      <h2 className="mt-10 text-2xl font-bold text-white">What OffGridFlow gives you</h2>
      <ul>
        <li>CSV and SAP/ERP connectors to ingest procurement, logistics, and travel data</li>
        <li>Per-category calculation methods across all 15 categories</li>
        <li>PCAF tier tagging for Investment category (financed emissions)</li>
        <li>Supplier-request templates to collect primary data from your vendors</li>
        <li>Year-over-year tier progression view so you can demonstrate data quality improvement</li>
      </ul>

      <p className="mt-8 text-sm text-gray-500">
        Related reading:{' '}
        <Link href="/scope-1-2-3-reporting-software" className="text-primary-400 hover:underline">Scope 1, 2, 3 reporting</Link>
        {' · '}
        <Link href="/for-procurement" className="text-primary-400 hover:underline">For procurement teams</Link>
        {' · '}
        <Link href="/sap-carbon-reporting" className="text-primary-400 hover:underline">SAP carbon reporting</Link>
      </p>
    </MoneyPageLayout>
  );
}
