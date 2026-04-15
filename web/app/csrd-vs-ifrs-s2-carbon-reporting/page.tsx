import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/csrd-vs-ifrs-s2-carbon-reporting';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'CSRD vs IFRS S2 — Carbon Reporting Compared',
  description:
    'Side-by-side comparison of CSRD (ESRS E1) and IFRS S2 climate disclosure requirements. Scope, materiality, assurance, and machine-readable filing differences.',
  keyword: 'CSRD vs IFRS S2',
});

const faqs = [
  {
    question: 'Do I have to report under both CSRD and IFRS S2?',
    answer:
      'Often yes. A multinational with EU operations above CSRD thresholds and securities listed in an IFRS S2 jurisdiction (UK, Canada, Singapore, etc.) reports under both. The underlying Scope 1/2/3 data is the same; the packaging differs.',
  },
  {
    question: 'What is the biggest functional difference?',
    answer:
      'Materiality. CSRD requires double materiality (financial + impact on people and environment). IFRS S2 requires financial materiality only. This affects what Scope 3 categories you must disclose and how narrative context is framed.',
  },
  {
    question: 'Which is stricter on assurance?',
    answer:
      'CSRD mandates assurance explicitly, moving from limited to reasonable over time. IFRS S2 defers to the adopting jurisdiction — some require assurance, some do not yet. Expect assurance for both as adoption matures.',
  },
  {
    question: 'Does OffGridFlow produce both outputs from one dataset?',
    answer:
      'Yes. Activity data is imported once; calculations run once. CSRD export uses ESRS XBRL; IFRS S2 export uses the ISSB taxonomy where applicable. One audit trail, two drafts.',
  },
];

export default function CsrdVsIfrsPage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'CSRD vs IFRS S2', path: PATH },
      ]}
      h1="CSRD vs IFRS S2 — Carbon Reporting Compared"
      dek="Two major climate disclosure regimes, one underlying dataset. Here is how CSRD (ESRS E1) and IFRS S2 differ on scope, materiality, assurance, and filing format — and how to report under both without duplicating work."
      slug="csrd-vs-ifrs-s2"
      ctaUtm="csrd_vs_ifrs"
      faqs={faqs}
    >
      <p>
        CSRD and IFRS S2 are the two most consequential climate disclosure frameworks of the
        next five years. They overlap on Scope 1/2/3 emissions disclosure but diverge on
        materiality, narrative expectations, assurance rigor, and machine-readable filing
        format.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">Head-to-head</h2>
      <div className="not-prose my-6 overflow-x-auto rounded-xl border border-gray-800">
        <table className="w-full text-sm">
          <thead className="bg-gray-800/50 text-left text-xs text-gray-500">
            <tr>
              <th className="px-4 py-3">Dimension</th>
              <th className="px-4 py-3">CSRD / ESRS E1</th>
              <th className="px-4 py-3">IFRS S2</th>
            </tr>
          </thead>
          <tbody className="text-gray-300">
            {[
              ['Issuing body', 'European Union', 'IFRS Foundation / ISSB'],
              ['Jurisdiction', 'EU + non-EU with EU ops', 'UK, Canada, Singapore, Australia, and more adopting'],
              ['Materiality', 'Double (financial + impact)', 'Financial only'],
              ['Scope 1 & 2', 'Required', 'Required'],
              ['Scope 3', 'All material categories', 'All material categories'],
              ['Scope 2 methods', 'Location + market required', 'Location + market required'],
              ['Scenario analysis', 'Required with transition plan', 'Required with transition plan'],
              ['Assurance', 'Limited -> reasonable (phased)', 'Jurisdiction-dependent'],
              ['Filing format', 'XBRL (ESRS taxonomy)', 'ISSB taxonomy (emerging)'],
              ['First reporting wave', 'FY2024 (largest EU PIEs)', 'FY2024/2025 depending on jurisdiction'],
            ].map((row) => (
              <tr key={row[0]} className="border-t border-gray-800/50">
                {row.map((cell, i) => (
                  <td key={i} className={`px-4 py-2 ${i === 0 ? 'font-medium text-white' : ''}`}>
                    {cell}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <h2 className="mt-10 text-2xl font-bold text-white">How OffGridFlow handles both</h2>
      <ul>
        <li>Import activity data once (CSV, cloud connectors, SAP, utility APIs)</li>
        <li>Calculate Scope 1/2/3 using GHG Protocol methodology — same engine for both frameworks</li>
        <li>Lock a factor snapshot per reporting period for CSRD and IFRS S2 reproducibility</li>
        <li>Export CSRD draft (PDF + ESRS XBRL)</li>
        <li>Export IFRS S2 draft (PDF + ISSB taxonomy where applicable)</li>
        <li>Single assurance evidence pack serves both engagements</li>
      </ul>

      <p className="mt-8 text-sm text-gray-500">
        Related:{' '}
        <Link href="/csrd-reporting-software" className="text-primary-400 hover:underline">CSRD reporting software</Link>
        {' · '}
        <Link href="/ifrs-s2-reporting-software" className="text-primary-400 hover:underline">IFRS S2 reporting software</Link>
      </p>
    </MoneyPageLayout>
  );
}
