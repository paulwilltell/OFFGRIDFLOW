import Link from 'next/link';
import type { Metadata } from 'next';

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
      <nav className="border-b border-gray-800/50 bg-dark-900/80 backdrop-blur-md">
        <div className="mx-auto flex max-w-4xl items-center justify-between px-6 py-4">
          <Link href="/" className="text-xl font-bold text-white">OffGridFlow</Link>
          <Link href="/" className="text-sm text-gray-400 hover:text-white">Back to Home</Link>
        </div>
      </nav>

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

        <div className="mt-10 text-center">
          <Link href="/security" className="text-sm text-primary-400 hover:underline">Security &amp; Trust Center</Link>
          <span className="mx-3 text-gray-700">|</span>
          <Link href="/about" className="text-sm text-primary-400 hover:underline">About OffGridFlow</Link>
        </div>
      </div>
    </div>
  );
}
