import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/csv-emissions-import';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'CSV Emissions Import',
  description:
    'Upload your utility bills, fleet fuel, and activity data as CSV. OffGridFlow calculates Scope 1, 2, and 3 emissions from your file and produces audit-ready draft reports.',
  keyword: 'CSV emissions import',
});

const faqs = [
  {
    question: 'What columns does the CSV need?',
    answer:
      'At minimum: meter_id, location, period_start, period_end, and quantity (kWh, liters, m3, etc.). Optional columns for scope, category, supplier id, and notes improve accuracy. A template is available from the in-app help widget.',
  },
  {
    question: 'How large can the file be?',
    answer:
      'Up to 50 MB per upload. For larger files, split by reporting period. Idempotent ingestion means re-uploading the same rows does not create duplicates — the dedupe keys are meter_id + period.',
  },
  {
    question: 'What happens if a row has bad data?',
    answer:
      'Validation errors are returned row-by-row with specific reasons (missing field, unrecognized unit, unparseable date). No partial imports — either the whole batch passes or the whole batch is rejected so you never wonder which rows made it.',
  },
  {
    question: 'Do I need to know what emission factor applies?',
    answer:
      'No. OffGridFlow selects the factor based on location, period, and category. You can see the selected factor in the calculation ledger and override if needed.',
  },
  {
    question: 'How long before I see a first report?',
    answer:
      'Minutes. Upload → calculation → dashboard view typically under 10 minutes. Generating a framework-specific draft (CSRD, SEC, SB 253) takes another 30 seconds to a few minutes depending on data volume.',
  },
];

export default function CsvImportPage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'CSV Emissions Import', path: PATH },
      ]}
      h1="CSV Emissions Import"
      dek="Upload a file. Get a calculation ledger, a dashboard, and a draft compliance report — in under two hours. No IT project, no cloud connector setup required."
      slug="csv-emissions-import"
      ctaUtm="csv_import"
      faqs={faqs}
    >
      <p>
        The fastest path from spreadsheet to audit-ready disclosure is a CSV upload. OffGridFlow
        accepts utility bills, fleet fuel records, electricity consumption, and generic activity
        data in a simple format. The calculation engine does the factor lookup, applies the
        right method, and produces an immutable ledger entry per row.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">Template</h2>
      <div className="not-prose my-6 overflow-x-auto rounded-xl border border-gray-800">
        <table className="w-full text-xs">
          <thead className="bg-gray-800/50 text-left text-[11px] text-gray-500">
            <tr>
              <th className="px-3 py-2">meter_id</th>
              <th className="px-3 py-2">location</th>
              <th className="px-3 py-2">period_start</th>
              <th className="px-3 py-2">period_end</th>
              <th className="px-3 py-2">kwh</th>
              <th className="px-3 py-2">scope</th>
              <th className="px-3 py-2">category</th>
            </tr>
          </thead>
          <tbody className="text-gray-300 font-mono">
            <tr className="border-t border-gray-800/50">
              <td className="px-3 py-2">CAMX-HQ-01</td>
              <td className="px-3 py-2">CAMX</td>
              <td className="px-3 py-2">2026-01-01</td>
              <td className="px-3 py-2">2026-01-31</td>
              <td className="px-3 py-2">3112</td>
              <td className="px-3 py-2">2</td>
              <td className="px-3 py-2">electricity</td>
            </tr>
            <tr className="border-t border-gray-800/50">
              <td className="px-3 py-2">FLEET-DIESEL-03</td>
              <td className="px-3 py-2">US-WEST</td>
              <td className="px-3 py-2">2026-01-01</td>
              <td className="px-3 py-2">2026-01-31</td>
              <td className="px-3 py-2">842</td>
              <td className="px-3 py-2">1</td>
              <td className="px-3 py-2">diesel</td>
            </tr>
          </tbody>
        </table>
      </div>

      <h2 className="mt-10 text-2xl font-bold text-white">What happens after upload</h2>
      <ol className="list-decimal list-inside space-y-1">
        <li>Each row becomes an activity record</li>
        <li>Factor lookup selects the right EPA/IEA/DEFRA/IPCC factor by region and period</li>
        <li>Calculation runs, produces an immutable ledger entry with formula</li>
        <li>Dashboard updates with scope totals</li>
        <li>Anomaly detection flags outliers, duplicates, or unexpected jumps</li>
        <li>Draft compliance report generated on demand</li>
      </ol>

      <p className="mt-8 text-sm text-gray-500">
        <Link href="/aws-carbon-data" className="text-primary-400 hover:underline">AWS connector</Link>
        {' · '}
        <Link href="/sap-carbon-reporting" className="text-primary-400 hover:underline">SAP integration</Link>
        {' · '}
        <Link href="/carbon-reporting-template" className="text-primary-400 hover:underline">Download template</Link>
      </p>
    </MoneyPageLayout>
  );
}
