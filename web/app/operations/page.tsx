import Link from 'next/link';
import type { Metadata } from 'next';
import { SiteNav } from '../components/SiteNav';

export const metadata: Metadata = {
  title: 'Operations Proof | OffGridFlow',
  description:
    'Measured benchmark results, release discipline, rollback procedure, and operational controls for OffGridFlow.',
};

const benchmarkCards = [
  { metric: 'Health endpoint p95', value: '18 ms', note: 'Target < 50 ms' },
  { metric: 'Auth p95', value: '78 ms', note: 'Target < 100 ms' },
  { metric: 'Emissions calc p95', value: '156 ms', note: 'Target < 200 ms' },
  { metric: 'Report generation p95', value: '780 ms', note: 'Target < 1000 ms' },
  { metric: 'Database query p95', value: '58 ms', note: 'Target < 100 ms' },
  { metric: 'Overall success rate', value: '99.8%', note: '14,275 / 14,300 requests' },
];

const releaseDiscipline = [
  'Run backend, frontend, integration, and security checks before major release.',
  'Verify secrets, database connectivity, backups, indexes, HPA, PDB, and ingress controls.',
  'Execute migrations first, then roll out API, worker, and web services with rollout status checks.',
  'Run smoke tests and live health checks after deploy, then monitor error rate and p95 latency for two hours.',
];

const rollbackSteps = [
  'Roll back API, worker, and web deployments with rollout undo.',
  'Verify rollback health before considering database rollback.',
  'If schema change is implicated, roll back one migration step and re-check service health.',
  'Continue monitoring for 30 minutes and run a post-incident review.',
];

const accessibilityProof = [
  'Dashboard keyboard navigation coverage exists in Playwright workflow tests.',
  'Dynamic chart components expose aria-label, aria-labelledby, and aria-describedby metadata.',
  'Dialogs use role="dialog" and aria-modal="true"; error states use aria-live alert regions.',
  'Loading states, focusable controls, and dense tables are exercised in end-to-end flows.',
];

export default function OperationsPage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <SiteNav />

      <main className="mx-auto max-w-6xl px-6 py-16">
        <h1 className="text-3xl font-bold text-white">Operations Proof</h1>
        <p className="mt-3 max-w-3xl text-gray-400">
          This page pulls together the measured benchmark snapshot, release checklist, rollback path,
          and operator-facing UX controls used to run OffGridFlow. Benchmark figures below come from the
          latest documented benchmark run on <span className="text-white">December 5, 2025</span> in a
          controlled Docker Compose environment.
        </p>

        <section className="mt-10 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {benchmarkCards.map((card) => (
            <div key={card.metric} className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <div className="text-2xl font-bold text-primary-400">{card.value}</div>
              <div className="mt-1 text-sm text-white">{card.metric}</div>
              <div className="mt-1 text-xs text-gray-500">{card.note}</div>
            </div>
          ))}
        </section>

        <section className="mt-16">
          <h2 className="text-xl font-semibold text-white">1. Benchmark Snapshot</h2>
          <p className="mt-2 text-sm text-gray-500">
            Latest documented test configuration: 10 workers, 60-second duration, 100 RPS target for the main workload.
          </p>
          <div className="mt-6 overflow-hidden rounded-xl border border-gray-800">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-800 bg-gray-800/50 text-left text-xs text-gray-500">
                  <th className="px-4 py-3">Workload</th>
                  <th className="px-4 py-3">Total requests</th>
                  <th className="px-4 py-3">Success</th>
                  <th className="px-4 py-3">Avg latency</th>
                  <th className="px-4 py-3">p95</th>
                  <th className="px-4 py-3">Throughput</th>
                </tr>
              </thead>
              <tbody className="text-gray-300">
                {[
                  ['Health endpoint', '500', '100%', '12 ms', '18 ms', '50.2 RPS'],
                  ['Authentication API', '3,000', '99.9%', '45 ms', '78 ms', '99.9 RPS'],
                  ['Emissions calculation API', '6,000', '99.8%', '85 ms', '156 ms', '99.8 RPS'],
                  ['Report generation', '300', '99.3%', '450 ms', '780 ms', '9.93 RPS'],
                  ['Database query load', '4,500', '99.8%', '32 ms', '58 ms', '149.7 RPS'],
                ].map((row) => (
                  <tr key={row[0]} className="border-b border-gray-800/30">
                    {row.map((cell, index) => (
                      <td key={`${row[0]}-${index}`} className={`px-4 py-2 ${index === 0 ? 'font-medium text-white' : ''}`}>
                        {cell}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="mt-4 rounded-xl border border-gray-800 bg-gray-800/20 p-5 text-sm text-gray-300">
            Capacity model from the same benchmark set: API target <span className="text-white">1,000 RPS</span>,
            web target <span className="text-white">500 to 1,000 concurrent users</span>, and worker target{' '}
            <span className="text-white">50 batch jobs per hour</span> per baseline replica set.
          </div>
        </section>

        <section className="mt-16 grid gap-8 lg:grid-cols-2">
          <div className="rounded-2xl border border-gray-800 bg-gray-800/20 p-6">
            <h2 className="text-xl font-semibold text-white">2. Release Discipline</h2>
            <p className="mt-2 text-sm text-gray-500">
              Major releases follow a written pre-deployment, deployment, and post-deployment checklist.
            </p>
            <div className="mt-5 space-y-3">
              {releaseDiscipline.map((item) => (
                <div key={item} className="flex items-start gap-3 text-sm text-gray-300">
                  <span className="mt-0.5 text-primary-400">&#10003;</span>
                  {item}
                </div>
              ))}
            </div>
            <div className="mt-5 rounded-lg bg-gray-900 p-4 text-xs text-gray-500">
              Monitoring window after deployment: error rate &lt; 1%, p95 latency &lt; 500 ms, database
              connections &lt; 80% of max, no unexpected alerts firing.
            </div>
          </div>

          <div className="rounded-2xl border border-gray-800 bg-gray-800/20 p-6">
            <h2 className="text-xl font-semibold text-white">3. Rollback and Recovery</h2>
            <p className="mt-2 text-sm text-gray-500">
              Recovery targets are documented and rollback steps are explicit rather than ad hoc.
            </p>
            <div className="mt-5 space-y-3">
              {rollbackSteps.map((item) => (
                <div key={item} className="flex items-start gap-3 text-sm text-gray-300">
                  <span className="mt-0.5 text-primary-400">&#10003;</span>
                  {item}
                </div>
              ))}
            </div>
            <div className="mt-5 grid gap-3 sm:grid-cols-2">
              <div className="rounded-lg border border-gray-800 bg-gray-900/50 p-4">
                <div className="text-xs text-gray-500">RTO</div>
                <div className="mt-1 text-lg font-semibold text-white">4 hours</div>
              </div>
              <div className="rounded-lg border border-gray-800 bg-gray-900/50 p-4">
                <div className="text-xs text-gray-500">RPO</div>
                <div className="mt-1 text-lg font-semibold text-white">1 hour</div>
              </div>
            </div>
          </div>
        </section>

        <section className="mt-16">
          <h2 className="text-xl font-semibold text-white">4. Accessibility and Dense Workflow Evidence</h2>
          <p className="mt-2 text-sm text-gray-500">
            The operator-facing application is designed for keyboard, assistive technology, and high-density data work.
          </p>
          <div className="mt-6 grid gap-4 md:grid-cols-2">
            {accessibilityProof.map((item) => (
              <div key={item} className="rounded-xl border border-gray-800 bg-gray-800/20 p-5 text-sm text-gray-300">
                {item}
              </div>
            ))}
          </div>
        </section>

        <section className="mt-16 rounded-2xl border border-primary-600/20 bg-primary-600/5 p-6">
          <h2 className="text-xl font-semibold text-white">5. Operational Caveat</h2>
          <p className="mt-2 text-sm text-gray-300">
            The benchmark figures on this page are the latest documented internal measurements, not a public third-party attestation.
            They are published to show actual performance envelopes, targets, and release controls without overstating certification or uptime claims.
          </p>
        </section>

        <div className="mt-14 flex flex-wrap gap-4 text-sm">
          <Link href="/status" className="text-primary-400 hover:underline">Live Status</Link>
          <Link href="/evidence" className="text-primary-400 hover:underline">Evidence Pack</Link>
          <Link href="/trust" className="text-primary-400 hover:underline">Trust Center</Link>
          <Link href="/how-we-operate" className="text-primary-400 hover:underline">How We Operate</Link>
        </div>
      </main>
    </div>
  );
}
