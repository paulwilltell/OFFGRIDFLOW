import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/cbam-reporting-software';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'EU CBAM Reporting Software',
  description:
    'Software for EU Carbon Border Adjustment Mechanism (CBAM) reporting. Embedded emissions per product, default vs actual values, quarterly declaration drafts.',
  keyword: 'CBAM reporting software',
});

const faqs = [
  {
    question: 'What does CBAM require?',
    answer:
      'The EU Carbon Border Adjustment Mechanism requires importers of cement, iron and steel, aluminum, fertilizers, electricity, and hydrogen to report embedded direct and indirect emissions on goods imported into the EU. Quarterly declarations were required during the transitional phase (October 2023 to December 2025). Definitive phase began January 2026 with financial obligations.',
  },
  {
    question: 'Can I use default values or do I need actual emissions?',
    answer:
      'The definitive phase requires actual verified emissions from non-EU producers. Default values remain available as a fallback with a penalty — the EU publishes default values per product category per origin country. OffGridFlow supports both, and clearly marks which reporting rows use which.',
  },
  {
    question: 'How does OffGridFlow handle CBAM certificates?',
    answer:
      'OffGridFlow calculates embedded emissions per product shipment and produces the declaration draft. Certificate purchase and surrender is handled via the EU CBAM Registry — we do not transact certificates. We do produce the data that tells you how many certificates you need.',
  },
  {
    question: 'Do I still need a customs or trade compliance tool?',
    answer:
      'Yes. CBAM sits alongside your customs filing and ERP. OffGridFlow is the emissions calculation layer. Your customs broker or trade compliance tool handles the actual CBAM declaration submission.',
  },
];

export default function CbamPage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'CBAM Reporting Software', path: PATH },
      ]}
      h1="EU Carbon Border Adjustment Mechanism (CBAM) Reporting Software"
      dek="Calculate embedded emissions per product, produce quarterly declaration drafts, and keep a defensible audit trail across transitional and definitive phases."
      slug="cbam-reporting-software"
      ctaUtm="cbam_reporting"
      faqs={faqs}
      relatedPages={[
        { href: '/carbon-accounting-software', label: 'Carbon accounting software', description: 'See the wider calculation platform behind CBAM outputs.' },
        { href: '/for-procurement', label: 'For procurement teams', description: 'Coordinate supplier and customs data collection.' },
        { href: '/sap-carbon-reporting', label: 'SAP carbon reporting', description: 'Use ERP data as the CBAM source system.' },
      ]}
    >
      <p>
        CBAM is the EU&apos;s tariff on embedded carbon in imports of cement, iron and steel,
        aluminum, fertilizers, electricity, and hydrogen. Importers must measure the direct and
        indirect emissions embedded in goods that cross the EU border and either submit actual
        data or pay the default-value premium.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">What OffGridFlow does for CBAM</h2>
      <ul>
        <li>Ingests import shipment data via CSV or customs-system export</li>
        <li>Applies product-category-specific calculation methods for embedded emissions</li>
        <li>Supports both actual-value calculations (from supplier data) and default values</li>
        <li>Tags each reporting row with data source tier so reviewers see which values are primary vs default</li>
        <li>Produces quarterly declaration drafts during transitional periods</li>
        <li>Records calculations in the immutable ledger for assurance requests</li>
      </ul>

      <h2 className="mt-10 text-2xl font-bold text-white">What OffGridFlow does not do</h2>
      <p>
        OffGridFlow does not submit CBAM declarations to the EU registry and does not transact
        CBAM certificates. We produce the declaration data; your customs broker, trade
        compliance tool, or the CBAM Registry portal handles filing and certificate surrender.
      </p>

    </MoneyPageLayout>
  );
}
