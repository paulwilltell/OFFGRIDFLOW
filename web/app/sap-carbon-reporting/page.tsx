import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/sap-carbon-reporting';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'SAP Carbon Reporting Integration',
  description:
    'Pull energy, fleet, and spend data from SAP S/4HANA and ECC into OffGridFlow for Scope 1, 2, and 3 carbon accounting. CSRD, SEC, SB 253, and CBAM reporting from one pipeline.',
  keyword: 'SAP carbon reporting',
});

const faqs = [
  {
    question: 'Which SAP systems are supported?',
    answer:
      'SAP S/4HANA and SAP ECC (ERP Central Component) via standard OData endpoints. Energy, utility, fleet fuel, facility, procurement, and travel modules are the typical ingestion sources.',
  },
  {
    question: 'How is authentication handled?',
    answer:
      'OAuth2 client credentials flow with scoped read-only service account access. Credentials encrypted at rest. No direct database access required — all ingestion goes through OData APIs.',
  },
  {
    question: 'Does it work with SAP Sustainability Control Tower?',
    answer:
      'SAP Sustainability Control Tower is SAP\'s own sustainability reporting product. OffGridFlow is a focused alternative for teams who want published methodology, transparent pricing, and multi-framework export without the SAP module commitment. For teams deep in the SAP ecosystem, SCT may be the better fit.',
  },
  {
    question: 'Can I start with CSV if SAP integration is blocked by IT?',
    answer:
      'Yes. CSV upload is always available. Start with exported SAP reports, get to a first compliance draft, and add direct connector access once IT approves.',
  },
];

export default function SapCarbonReportingPage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'SAP Carbon Reporting', path: PATH },
      ]}
      h1="SAP Carbon Reporting Integration"
      dek="Ingest energy, fleet, facility, and spend data directly from SAP S/4HANA or ECC. Calculate Scope 1, 2, and 3 against published factor sources. Export drafts for every major framework."
      slug="sap-carbon-reporting"
      ctaUtm="sap_carbon"
      faqs={faqs}
    >
      <p>
        For companies running SAP as the system of record, moving emissions data into a
        spreadsheet is a regression. OffGridFlow connects directly to SAP via standard OData
        APIs so activity data flows from ERP to calculation ledger to compliance export without
        manual extraction.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">What the connector pulls</h2>
      <ul>
        <li>Energy consumption (electricity, natural gas, steam, heating, cooling) per facility</li>
        <li>Fleet fuel consumption by vehicle or route</li>
        <li>Facility master data (location, floor area, operational hours)</li>
        <li>Procurement spend by commodity for Scope 3 Category 1 estimation</li>
        <li>Business travel records for Scope 3 Category 6</li>
      </ul>

      <h2 className="mt-10 text-2xl font-bold text-white">Why it matters for audit</h2>
      <p>
        When emissions data traces back to the same SAP source that feeds financial statements,
        assurance providers only have to verify the mapping logic once. OffGridFlow records the
        SAP document id on every ingested activity so the chain from financial audit to
        emissions audit is explicit.
      </p>

      <p className="mt-8 text-sm text-gray-500">
        <Link href="/aws-carbon-data" className="text-primary-400 hover:underline">AWS integration</Link>
        {' · '}
        <Link href="/scope-3-supplier-emissions-software" className="text-primary-400 hover:underline">Scope 3 supplier emissions</Link>
        {' · '}
        <Link href="/for-procurement" className="text-primary-400 hover:underline">For procurement</Link>
      </p>
    </MoneyPageLayout>
  );
}
