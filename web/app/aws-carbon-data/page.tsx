import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/aws-carbon-data';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'AWS Carbon Data Integration',
  description:
    'Pull AWS carbon footprint data into OffGridFlow via Cost and Usage Reports (CUR) or the AWS Customer Carbon Footprint Tool API. Scope 2 and Scope 3 Category 1 coverage.',
  keyword: 'AWS carbon data integration',
});

const faqs = [
  {
    question: 'Which AWS data source does OffGridFlow use?',
    answer:
      'Both. The AWS Customer Carbon Footprint Tool (CCFT) exposes monthly emission estimates by service and region. Cost and Usage Reports (CUR) delivered to S3 provide the usage-level granularity needed for more accurate Scope 2 location-based and market-based calculations.',
  },
  {
    question: 'Does AWS\'s own carbon report replace OffGridFlow?',
    answer:
      'No. AWS reports only your AWS footprint — useful but limited. OffGridFlow combines AWS data with on-premises, utility, fleet, travel, and value-chain data to produce a full-organization Scope 1/2/3 disclosure.',
  },
  {
    question: 'What credentials do I need?',
    answer:
      'An IAM role with read access to the CUR S3 bucket (for CUR ingestion) or appropriate permissions for the Carbon Footprint Tool API. Credentials are stored encrypted at rest and never transmitted in logs. Documented in the /trust center.',
  },
  {
    question: 'How often does data sync?',
    answer:
      'AWS publishes CUR at least daily and CCFT monthly. OffGridFlow syncs on a schedule you configure with idempotent ingestion so replaying a file never creates duplicate activities.',
  },
];

export default function AwsCarbonDataPage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'AWS Carbon Data', path: PATH },
      ]}
      h1="AWS Carbon Data Integration"
      dek="Pull cloud emissions from AWS Cost and Usage Reports or the Customer Carbon Footprint Tool API directly into the OffGridFlow calculation engine — as part of your organization-wide Scope 1, 2, and 3 disclosure."
      slug="aws-carbon-data"
      ctaUtm="aws_carbon"
      faqs={faqs}
    >
      <p>
        Cloud compute is a significant Scope 2 and Scope 3 Category 1 source for most modern
        companies. AWS exposes carbon data through two channels: the Customer Carbon Footprint
        Tool (CCFT) API and Cost and Usage Reports (CUR). OffGridFlow ingests both and merges
        them with your on-premises, utility, fleet, travel, and supplier data for a complete
        organization-wide inventory.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">How the connector works</h2>
      <ul>
        <li>Provide IAM credentials (scoped read-only to CUR S3 bucket or CCFT API)</li>
        <li>Select a reporting period and schedule (daily, weekly, monthly)</li>
        <li>OffGridFlow pulls the latest report and ingests it into the activity ledger</li>
        <li>Calculations use EPA eGRID or IEA factors depending on AWS region</li>
        <li>Idempotent ingestion prevents duplicate activities on re-sync</li>
      </ul>

      <h2 className="mt-10 text-2xl font-bold text-white">Security posture</h2>
      <p>
        Credentials encrypted at rest (AES-256). Never echoed in logs. IAM role scoped to
        read-only access on the specific CUR bucket or CCFT endpoint. Full detail in the{' '}
        <Link href="/trust" className="text-primary-400 hover:underline">Trust Center</Link>.
      </p>

      <p className="mt-8 text-sm text-gray-500">
        <Link href="/sap-carbon-reporting" className="text-primary-400 hover:underline">SAP integration</Link>
        {' · '}
        <Link href="/csv-emissions-import" className="text-primary-400 hover:underline">CSV import alternative</Link>
        {' · '}
        <Link href="/architecture" className="text-primary-400 hover:underline">Architecture</Link>
      </p>
    </MoneyPageLayout>
  );
}
