import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { JsonLd, datasetSchema } from '@/components/JsonLd';
import { buildMoneyPageMetadata, SITE_URL } from '@/lib/seo';

const PATH = '/carbon-reporting-template';
const DOWNLOAD_URL = `${SITE_URL}/downloads/carbon-reporting-template.csv`;

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'Carbon Reporting CSV Template (Free Download)',
  description:
    'Free CSV template for carbon accounting uploads. Pre-populated with Scope 1, 2, and 3 example rows covering electricity, fuel, refrigerants, travel, commuting, and spend.',
  keyword: 'carbon reporting template',
});

const faqs = [
  {
    question: 'What columns does the template include?',
    answer:
      'meter_id, location, period_start, period_end, quantity, unit, scope, category, source, notes. This is the canonical format for OffGridFlow CSV uploads and is also a reasonable baseline for any carbon accounting workflow.',
  },
  {
    question: 'Are the example rows realistic?',
    answer:
      'Yes. The rows cover typical Scope 1 sources (diesel, natural gas, refrigerant), Scope 2 (electricity in CAMX and NWPP), and Scope 3 (flight, commute, spend). Quantities and units match what a real monthly utility or travel export would contain.',
  },
  {
    question: 'Can I use this with tools other than OffGridFlow?',
    answer:
      'Yes. The format is standard enough to feed into most carbon accounting platforms or a spreadsheet model. For tools with different column requirements, use this as a starting point and adjust headers.',
  },
  {
    question: 'What if I do not have a row for every scope?',
    answer:
      'Fine. Upload what you have. Scope coverage can be built up over time. Data quality tiers let you label which calculations use primary data vs estimates — recommended for CSRD and SB 253 disclosures.',
  },
];

export default function CarbonReportingTemplatePage() {
  return (
    <>
      <JsonLd
        id="ld-dataset-csv-template"
        data={datasetSchema({
          name: 'OffGridFlow Carbon Reporting CSV Template',
          description:
            'Canonical CSV format for carbon accounting uploads. Includes example rows for Scope 1 (diesel, natural gas, refrigerants), Scope 2 (electricity), and Scope 3 (travel, commuting, spend).',
          url: `${SITE_URL}${PATH}`,
          distributionUrl: DOWNLOAD_URL,
          encodingFormat: 'text/csv',
          keywords: ['carbon accounting template', 'CSV template', 'emissions upload', 'GHG Protocol'],
        })}
      />
      <MoneyPageLayout
        breadcrumbs={[
          { name: 'Home', path: '/' },
          { name: 'Carbon Reporting Template', path: PATH },
        ]}
        h1="Carbon Reporting CSV Template"
        dek="Free CSV template pre-populated with representative Scope 1, 2, and 3 example rows. Drop your data into the columns and upload directly into OffGridFlow — or use as a baseline for any carbon accounting workflow."
        slug="carbon-reporting-template"
        ctaUtm="reporting_template"
        ctaText="Upload in OffGridFlow"
        secondaryCtaText="Download template"
        secondaryCtaHref="/downloads/carbon-reporting-template.csv"
        faqs={faqs}
      >
        <div className="not-prose my-8 rounded-2xl border border-primary-600/30 bg-primary-600/5 p-6">
          <h2 className="text-lg font-semibold text-white">Download the template</h2>
          <p className="mt-2 text-sm text-gray-400">
            Canonical CSV format. Example rows cover Scope 1 (diesel, natural gas, refrigerants),
            Scope 2 (electricity in two US grid regions), and Scope 3 (business travel, commuting,
            procurement spend). Replace the example rows with your data and upload.
          </p>
          <div className="mt-4 flex flex-wrap gap-3">
            <a
              href="/downloads/carbon-reporting-template.csv"
              download
              className="rounded-lg bg-primary-600 px-5 py-2.5 text-sm font-semibold text-white hover:bg-primary-500"
            >
              Download CSV
            </a>
            <Link
              href="/csv-emissions-import"
              className="rounded-lg border border-gray-700 px-5 py-2.5 text-sm text-gray-300 hover:border-gray-500 hover:text-white"
            >
              How upload works
            </Link>
          </div>
        </div>

        <h2 className="mt-10 text-2xl font-bold text-white">Column reference</h2>
        <ul>
          <li><code className="rounded bg-gray-800 px-1 text-primary-400">meter_id</code> — stable identifier per activity source (e.g., meter, vehicle, account)</li>
          <li><code className="rounded bg-gray-800 px-1 text-primary-400">location</code> — region code (CAMX, NWPP, US-WEST, etc.) used for factor lookup</li>
          <li><code className="rounded bg-gray-800 px-1 text-primary-400">period_start / period_end</code> — ISO 8601 dates bounding the activity</li>
          <li><code className="rounded bg-gray-800 px-1 text-primary-400">quantity</code> — numeric value of consumption</li>
          <li><code className="rounded bg-gray-800 px-1 text-primary-400">unit</code> — kWh, L, m3, kg, pax-km, USD, etc.</li>
          <li><code className="rounded bg-gray-800 px-1 text-primary-400">scope</code> — 1, 2, or 3</li>
          <li><code className="rounded bg-gray-800 px-1 text-primary-400">category</code> — electricity, diesel, natural_gas, r_134a, flight_domestic, etc.</li>
          <li><code className="rounded bg-gray-800 px-1 text-primary-400">source</code> — ingestion source label (utility_bill_csv, fleet, business_travel, etc.)</li>
          <li><code className="rounded bg-gray-800 px-1 text-primary-400">notes</code> — free-form context preserved in the ledger</li>
        </ul>

        <p className="mt-8 text-sm text-gray-500">
          Related:{' '}
          <Link href="/csv-emissions-import" className="text-primary-400 hover:underline">CSV emissions import</Link>
          {' · '}
          <Link href="/scope-2-factor-library" className="text-primary-400 hover:underline">Scope 2 factor library</Link>
          {' · '}
          <Link href="/methodology" className="text-primary-400 hover:underline">Methodology</Link>
        </p>
      </MoneyPageLayout>
    </>
  );
}
