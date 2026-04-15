import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/for-procurement';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'OffGridFlow for Procurement',
  description:
    'Scope 3 supplier emissions tracking for procurement. Collect supplier data, score by PCAF/GHG-Protocol data quality, and wire into CSRD and SEC Climate disclosures.',
  keyword: 'procurement carbon software',
});

const faqs = [
  {
    question: 'Why should procurement care about Scope 3?',
    answer:
      'For most companies, Scope 3 Category 1 (purchased goods and services) is the largest single emissions category. CSRD, SEC Climate, and SB 253 require material Scope 3 disclosure. Procurement owns the supplier relationships that produce this data.',
  },
  {
    question: 'How does OffGridFlow fit into sourcing decisions?',
    answer:
      'Supplier emissions data is stored alongside other supplier attributes. When sourcing teams evaluate vendors, they can see emission intensity per dollar spent. When setting science-based targets, supplier engagement tracking shows which vendors are delivering reductions.',
  },
  {
    question: 'Can we request data from suppliers at scale?',
    answer:
      'Yes. Supplier request workflows, CDP-style response imports, and CSV templates let you collect primary data without custom email chains. Each response is tagged with data quality so your Scope 3 disclosure reflects real coverage.',
  },
  {
    question: 'What about CBAM?',
    answer:
      'Procurement teams importing cement, iron/steel, aluminum, fertilizer, electricity, or hydrogen into the EU need embedded emissions data from non-EU suppliers. OffGridFlow ingests shipment data and applies product-category calculations — actual values from suppliers or EU default values as fallback.',
  },
];

export default function ForProcurementPage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'For Procurement', path: PATH },
      ]}
      h1="OffGridFlow for Procurement"
      dek="Scope 3 Category 1 is where sourcing decisions meet climate disclosure. OffGridFlow gives procurement teams a supplier emissions system that ties directly into CSRD, SEC, SB 253, and CBAM reporting."
      slug="for-procurement"
      ctaUtm="for_procurement"
      faqs={faqs}
    >
      <p>
        Purchased goods and services is the largest emissions category for most non-financial
        companies. Procurement owns the relationships, the spend, and now — implicitly — the
        emissions disclosure risk. OffGridFlow connects supplier data collection to the
        compliance reporting layer without forcing procurement to become an emissions expert.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">What procurement teams do in OffGridFlow</h2>
      <ul>
        <li>Import supplier spend from ERP (SAP connector or CSV)</li>
        <li>Send supplier emissions data requests with pre-populated templates</li>
        <li>Track response rates and data quality tier by supplier</li>
        <li>Progress categories from spend-based to activity-based to supplier-specific</li>
        <li>Feed Scope 3 totals directly into CSRD, SEC, and SB 253 disclosure drafts</li>
        <li>Produce CBAM draft declarations for covered EU imports</li>
      </ul>

      <h2 className="mt-10 text-2xl font-bold text-white">What procurement does NOT need to do</h2>
      <p>
        Learn GHG Protocol methodology. Master emission factor databases. Build a spreadsheet
        model. Chase suppliers via email. OffGridFlow handles the methodology layer; procurement
        owns the relationships.
      </p>

      <p className="mt-8 text-sm text-gray-500">
        <Link href="/scope-3-supplier-emissions-software" className="text-primary-400 hover:underline">Scope 3 supplier emissions</Link>
        {' · '}
        <Link href="/cbam-reporting-software" className="text-primary-400 hover:underline">CBAM reporting</Link>
        {' · '}
        <Link href="/sap-carbon-reporting" className="text-primary-400 hover:underline">SAP integration</Link>
      </p>
    </MoneyPageLayout>
  );
}
