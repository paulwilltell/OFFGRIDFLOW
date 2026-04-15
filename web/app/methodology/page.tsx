import Link from 'next/link';
import type { Metadata } from 'next';
import { SiteNav } from '../components/SiteNav';
import { SiteFooter } from '../components/SiteFooter';

export const metadata: Metadata = {
  title: 'Methodology Library | OffGridFlow',
  description: 'Public documentation of emission factor sources, calculation methods, and GHG Protocol alignment.',
};

const factorSources = [
  { name: 'EPA eGRID 2023', scope: 'Scope 2', coverage: '27 US subregions', description: 'US EPA Emissions & Generation Resource Integrated Database. Location-based grid emission factors by eGRID subregion.', url: 'https://www.epa.gov/egrid' },
  { name: 'IEA 2023', scope: 'Scope 2', coverage: '55+ countries', description: 'International Energy Agency country-level grid emission factors for location-based Scope 2 accounting.', url: 'https://www.iea.org' },
  { name: 'UK DEFRA 2024', scope: 'Scope 1, 2, 3', coverage: 'Global', description: 'UK Department for Environment, Food & Rural Affairs conversion factors for stationary combustion, mobile sources, refrigerants, and value chain activities.', url: 'https://www.gov.uk/government/collections/government-conversion-factors-for-company-reporting' },
  { name: 'IPCC AR6 GWP-100', scope: 'Scope 1', coverage: 'Global', description: 'IPCC Sixth Assessment Report Global Warming Potential values for refrigerants and fugitive gases (100-year time horizon).', url: 'https://www.ipcc.ch' },
  { name: 'GHG Protocol Guidance', scope: 'Scope 3', coverage: 'Global', description: 'Category-level default factors and methodology guidance from the GHG Protocol Corporate Value Chain Standard.', url: 'https://ghgprotocol.org' },
];

const methods = [
  { name: 'Location-Based (Scope 2)', standard: 'GHG Protocol Scope 2 Guidance', description: 'Uses average grid emission factors for the region where electricity is consumed. Reflects the average emissions intensity of the local grid.' },
  { name: 'Market-Based (Scope 2)', standard: 'GHG Protocol Scope 2 Guidance', description: 'Uses supplier-specific or contractual instrument factors. Reflects purchasing decisions including RECs, PPAs, and green tariffs.' },
  { name: 'Activity-Based (Scope 1)', standard: 'GHG Protocol Corporate Standard', description: 'Calculates emissions from direct measurement of fuel consumed, distance traveled, or refrigerant leaked, multiplied by fuel-specific emission factors.' },
  { name: 'Spend-Based (Scope 3)', standard: 'GHG Protocol Scope 3 Standard', description: 'Estimates emissions using economic input-output (EEIO) factors applied to procurement spend data. Used when activity data is unavailable.' },
  { name: 'Supplier-Specific (Scope 3)', standard: 'GHG Protocol Scope 3 Standard', description: 'Uses primary data from suppliers for purchased goods and services. Highest accuracy tier for Scope 3 Category 1.' },
];

export default function MethodologyPage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <SiteNav />

      <div className="mx-auto max-w-4xl px-6 py-16">
        <h1 className="mb-2 text-3xl font-bold text-white">Methodology Library</h1>
        <p className="mb-10 text-gray-400">
          Public documentation of our emission factor sources, calculation methods, and standards alignment.
          Every number in OffGridFlow is traceable to these sources.
        </p>

        {/* Standards Alignment */}
        <section className="mb-12">
          <h2 className="mb-4 text-xl font-semibold text-white">Standards Alignment</h2>
          <div className="grid gap-3 sm:grid-cols-2">
            {[
              { standard: 'GHG Protocol Corporate Standard', status: 'Aligned', description: 'Scope 1 & 2 accounting and reporting' },
              { standard: 'GHG Protocol Scope 2 Guidance', status: 'Aligned', description: 'Dual reporting (location + market-based)' },
              { standard: 'GHG Protocol Scope 3 Standard', status: 'Aligned', description: 'Value chain accounting, 15 categories' },
              { standard: 'IFRS S2 (ISSB)', status: 'Mapped', description: 'Climate-related financial disclosures' },
              { standard: 'ESRS E1 (CSRD)', status: 'Mapped', description: 'EU sustainability reporting standard' },
              { standard: 'California SB 253', status: 'Ready', description: 'State-level Scope 1/2/3 reporting' },
            ].map(s => (
              <div key={s.standard} className="rounded-lg border border-gray-800 bg-gray-800/30 p-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-white">{s.standard}</span>
                  <span className={`rounded px-2 py-0.5 text-[10px] font-medium ${
                    s.status === 'Aligned' ? 'bg-green-500/10 text-green-400' :
                    s.status === 'Mapped' ? 'bg-blue-500/10 text-blue-400' :
                    'bg-amber-500/10 text-amber-400'
                  }`}>{s.status}</span>
                </div>
                <div className="mt-1 text-xs text-gray-500">{s.description}</div>
              </div>
            ))}
          </div>
        </section>

        {/* Factor Sources */}
        <section className="mb-12">
          <h2 className="mb-4 text-xl font-semibold text-white">Emission Factor Sources</h2>
          <p className="mb-4 text-sm text-gray-400">
            OffGridFlow uses 184 verified emission factors from the following authoritative sources.
            All factors include provenance metadata (source, region, validity period, methodology).
          </p>
          <div className="space-y-3">
            {factorSources.map(source => (
              <div key={source.name} className="rounded-lg border border-gray-800 bg-gray-800/30 p-5">
                <div className="flex items-start justify-between">
                  <div>
                    <div className="text-sm font-medium text-white">{source.name}</div>
                    <div className="mt-1 text-xs text-gray-400">{source.description}</div>
                  </div>
                  <div className="ml-4 flex flex-col items-end gap-1">
                    <span className="rounded bg-gray-700/50 px-2 py-0.5 text-[10px] text-gray-400">{source.scope}</span>
                    <span className="text-[10px] text-gray-600">{source.coverage}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Calculation Methods */}
        <section className="mb-12">
          <h2 className="mb-4 text-xl font-semibold text-white">Calculation Methods</h2>
          <div className="space-y-3">
            {methods.map(m => (
              <div key={m.name} className="rounded-lg border border-gray-800 bg-gray-800/30 p-5">
                <div className="flex items-start justify-between">
                  <div className="text-sm font-medium text-white">{m.name}</div>
                  <span className="ml-2 rounded bg-gray-700/50 px-2 py-0.5 text-[10px] text-gray-400">{m.standard}</span>
                </div>
                <div className="mt-2 text-xs text-gray-400">{m.description}</div>
              </div>
            ))}
          </div>
        </section>

        {/* Calculation Transparency */}
        <section className="mb-12">
          <h2 className="mb-4 text-xl font-semibold text-white">Calculation Transparency</h2>
          <div className="rounded-lg border border-gray-800 bg-gray-900/30 p-5">
            <p className="mb-4 text-sm text-gray-400">
              Every emission calculation in OffGridFlow is recorded in an immutable calculation ledger with the following fields:
            </p>
            <div className="grid gap-2 sm:grid-cols-2">
              {[
                'Activity quantity and unit',
                'Emission factor ID and value',
                'Factor source and region',
                'Calculation method used',
                'Human-readable formula',
                'Result in kg CO2e and tonnes CO2e',
                'Reporting period (start/end)',
                'Timestamp and user who triggered',
                'Lock status (immutable after approval)',
              ].map(item => (
                <div key={item} className="flex items-center gap-2 text-xs text-gray-300">
                  <span className="text-primary-400">&#10003;</span>
                  {item}
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Worked Calculation Example — fixes 1B.2, 1B.3, 2B.2 */}
        <section className="mb-12">
          <h2 className="mb-4 text-xl font-semibold text-white">Worked Example: Scope 2 Calculation</h2>
          <p className="mb-4 text-sm text-gray-400">
            This example shows how a single activity flows through the calculation engine, including factor selection, formula application, ledger recording, and data quality labeling.
          </p>

          {/* Input Activity */}
          <div className="mb-4 rounded-xl border border-gray-800 bg-gray-800/20 p-5">
            <h3 className="mb-3 text-xs font-medium uppercase tracking-wider text-primary-400">Step 1: Input Activity</h3>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4 text-sm">
              <div><span className="text-gray-500">Source:</span> <span className="text-white">CSV upload</span></div>
              <div><span className="text-gray-500">Type:</span> <span className="text-white">Electricity</span></div>
              <div><span className="text-gray-500">Scope:</span> <span className="text-white">2 (purchased energy)</span></div>
              <div><span className="text-gray-500">Location:</span> <span className="text-white">US-WEST (WECC region)</span></div>
              <div><span className="text-gray-500">Quantity:</span> <span className="text-white">10,000 kWh</span></div>
              <div><span className="text-gray-500">Period:</span> <span className="text-white">Jan 1 — Jan 31, 2026</span></div>
              <div><span className="text-gray-500">Data quality:</span> <span className="rounded bg-green-500/10 px-1.5 py-0.5 text-[10px] font-medium text-green-400">Measured</span></div>
              <div><span className="text-gray-500">Status:</span> <span className="rounded bg-yellow-500/10 px-1.5 py-0.5 text-[10px] font-medium text-yellow-400">Draft</span></div>
            </div>
          </div>

          {/* Factor Selection */}
          <div className="mb-4 rounded-xl border border-gray-800 bg-gray-800/20 p-5">
            <h3 className="mb-3 text-xs font-medium uppercase tracking-wider text-primary-400">Step 2: Factor Selection</h3>
            <div className="grid gap-3 sm:grid-cols-2 text-sm">
              <div><span className="text-gray-500">Factor ID:</span> <span className="text-white font-mono">grid-us-west</span></div>
              <div><span className="text-gray-500">Source:</span> <span className="text-white">EPA eGRID 2023 (WECC)</span></div>
              <div><span className="text-gray-500">Value:</span> <span className="text-white font-mono">0.298 kg CO2e / kWh</span></div>
              <div><span className="text-gray-500">Method:</span> <span className="text-white">Location-based</span></div>
              <div><span className="text-gray-500">Year:</span> <span className="text-white">2023</span></div>
              <div><span className="text-gray-500">Uncertainty:</span> <span className="text-white">+/- 5%</span></div>
            </div>
          </div>

          {/* Calculation */}
          <div className="mb-4 rounded-xl border border-primary-600/20 bg-primary-600/5 p-5">
            <h3 className="mb-3 text-xs font-medium uppercase tracking-wider text-primary-400">Step 3: Calculation</h3>
            <div className="rounded-lg bg-gray-900 p-4 font-mono text-sm text-gray-200">
              <div>emissions = quantity x factor</div>
              <div className="mt-1">emissions = 10,000 kWh x 0.298 kg CO2e/kWh</div>
              <div className="mt-1 text-primary-400 font-bold">emissions = 2,980.00 kg CO2e = 2.98 tonnes CO2e</div>
            </div>
          </div>

          {/* Ledger Record */}
          <div className="mb-4 rounded-xl border border-gray-800 bg-gray-800/20 p-5">
            <h3 className="mb-3 text-xs font-medium uppercase tracking-wider text-primary-400">Step 4: Immutable Ledger Record</h3>
            <p className="mb-3 text-xs text-gray-500">This record is created in the calculation ledger and cannot be modified after creation.</p>
            <div className="overflow-x-auto">
              <table className="w-full text-xs">
                <tbody className="text-gray-300">
                  {[
                    ['ledger_id', 'calc-2026-01-abc123'],
                    ['activity_id', 'act-electricity-jan2026'],
                    ['scope', '2'],
                    ['quantity', '10,000'],
                    ['unit', 'kWh'],
                    ['factor_id', 'grid-us-west'],
                    ['factor_value', '0.298 kg CO2e/kWh'],
                    ['factor_source', 'EPA eGRID 2023 (WECC)'],
                    ['factor_region', 'US-WEST'],
                    ['method', 'location-based'],
                    ['formula', '10000 * 0.298 = 2980.00 kg CO2e'],
                    ['result_kg_co2e', '2,980.00'],
                    ['result_tonnes_co2e', '2.98'],
                    ['data_quality', 'measured'],
                    ['calculated_by', 'operator@company.com'],
                    ['calculated_at', '2026-02-03T14:22:00Z'],
                    ['is_locked', 'false (draft)'],
                    ['version', '1'],
                  ].map(([field, value]) => (
                    <tr key={field} className="border-b border-gray-800/20">
                      <td className="py-1.5 pr-4 font-mono text-gray-500 whitespace-nowrap">{field}</td>
                      <td className="py-1.5 text-white">{value}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Approval & Lock */}
          <div className="mb-4 rounded-xl border border-gray-800 bg-gray-800/20 p-5">
            <h3 className="mb-3 text-xs font-medium uppercase tracking-wider text-primary-400">Step 5: Approval & Lock</h3>
            <div className="space-y-2 text-sm text-gray-300">
              <div className="flex items-center gap-3">
                <span className="rounded bg-blue-500/10 px-2 py-0.5 text-[10px] font-medium text-blue-400">Submitted</span>
                <span>by operator@company.com at 2026-02-03T14:30:00Z</span>
              </div>
              <div className="flex items-center gap-3">
                <span className="rounded bg-purple-500/10 px-2 py-0.5 text-[10px] font-medium text-purple-400">Reviewed</span>
                <span>by manager@company.com at 2026-02-04T09:15:00Z — &quot;Verified against utility invoice&quot;</span>
              </div>
              <div className="flex items-center gap-3">
                <span className="rounded bg-green-500/10 px-2 py-0.5 text-[10px] font-medium text-green-400">Approved</span>
                <span>by cfo@company.com at 2026-02-04T11:00:00Z</span>
              </div>
              <div className="flex items-center gap-3">
                <span className="rounded bg-green-500/10 px-2 py-0.5 text-[10px] font-medium text-green-400">Locked</span>
                <span>is_locked = true. Record is now immutable. Factor snapshot locked to Q1 2026.</span>
              </div>
            </div>
          </div>

          {/* Export Reconciliation */}
          <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
            <h3 className="mb-3 text-xs font-medium uppercase tracking-wider text-primary-400">Step 6: Export & Reconciliation</h3>
            <p className="mb-3 text-xs text-gray-500">When a compliance report is exported, the system records scope totals and a SHA256 checksum. Reconciliation verifies the export matches current data.</p>
            <div className="grid gap-3 sm:grid-cols-2 text-sm">
              <div><span className="text-gray-500">Report:</span> <span className="text-white">SEC Climate Disclosure 2026</span></div>
              <div><span className="text-gray-500">Format:</span> <span className="text-white">PDF</span></div>
              <div><span className="text-gray-500">Scope 2 at export:</span> <span className="text-white font-mono">2.98 tCO2e</span></div>
              <div><span className="text-gray-500">Checksum:</span> <span className="text-white font-mono">a7f3c2...</span></div>
              <div><span className="text-gray-500">Exported by:</span> <span className="text-white">cfo@company.com</span></div>
              <div><span className="text-gray-500">Reconciliation:</span> <span className="rounded bg-green-500/10 px-1.5 py-0.5 text-[10px] font-medium text-green-400">Match (0% drift)</span></div>
            </div>
          </div>
        </section>

        {/* Limitations */}
        <section>
          <h2 className="mb-4 text-xl font-semibold text-white">Known Limitations</h2>
          <div className="rounded-lg border border-amber-500/20 bg-amber-500/5 p-5">
            <ul className="space-y-2 text-sm text-gray-300">
              <li className="flex items-start gap-2">
                <span className="mt-0.5 text-amber-400">!</span>
                Scope 3 spend-based factors use EEIO coefficients which are sector averages, not company-specific. Accuracy improves with supplier-specific data.
              </li>
              <li className="flex items-start gap-2">
                <span className="mt-0.5 text-amber-400">!</span>
                Market-based Scope 2 requires user-supplied contractual instrument data. OffGridFlow does not verify REC/PPA authenticity.
              </li>
              <li className="flex items-start gap-2">
                <span className="mt-0.5 text-amber-400">!</span>
                Grid emission factors are updated annually. Mid-year changes in grid mix are not reflected until the next factor update cycle.
              </li>
              <li className="flex items-start gap-2">
                <span className="mt-0.5 text-amber-400">!</span>
                Process emissions (industrial) are supported at factor level but require manual activity data input — no automated process monitoring integration.
              </li>
            </ul>
          </div>
        </section>

        <section className="mt-12 rounded-2xl border border-gray-800 bg-gray-800/20 p-6">
          <h2 className="text-lg font-semibold text-white">Framework Pages Using This Methodology</h2>
          <p className="mt-2 text-sm text-gray-400">
            These pages apply the same published calculation basis to specific reporting programs and buyer roles.
          </p>
          <div className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {[
              { href: '/sb-253-reporting-software', label: 'SB 253 reporting software' },
              { href: '/csrd-reporting-software', label: 'CSRD reporting software' },
              { href: '/ifrs-s2-reporting-software', label: 'IFRS S2 reporting software' },
              { href: '/scope-1-2-3-reporting-software', label: 'Scope 1, 2, 3 reporting software' },
              { href: '/carbon-accounting-software-for-finance-teams', label: 'Carbon accounting for finance teams' },
              { href: '/audit-ready-carbon-accounting', label: 'Audit-ready carbon accounting' },
            ].map((item) => (
              <Link key={item.href} href={item.href} className="rounded-xl border border-gray-800/70 bg-gray-900/40 p-4 text-sm text-white transition hover:border-primary-600/40">
                {item.label}
              </Link>
            ))}
          </div>
        </section>

        <div className="mt-10 text-center">
          <Link href="/evidence" className="text-sm text-primary-400 hover:underline">Redacted Evidence Pack</Link>
          <span className="mx-3 text-gray-700">|</span>
          <Link href="/operations" className="text-sm text-primary-400 hover:underline">Operations Proof</Link>
          <span className="mx-3 text-gray-700">|</span>
          <Link href="/security" className="text-sm text-primary-400 hover:underline">Security &amp; Trust Center</Link>
          <span className="mx-3 text-gray-700">|</span>
          <Link href="/about" className="text-sm text-primary-400 hover:underline">About OffGridFlow</Link>
        </div>
      </div>
      <SiteFooter />
    </div>
  );
}
