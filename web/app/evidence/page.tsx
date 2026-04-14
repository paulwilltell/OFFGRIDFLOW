import Link from 'next/link';
import type { Metadata } from 'next';
import { SiteNav } from '../components/SiteNav';

export const metadata: Metadata = {
  title: 'Evidence Pack | OffGridFlow',
  description:
    'Redacted sample audit packet showing source activity rows, factor provenance, immutable ledger entries, approval trail, and export reconciliation.',
};

const activityExcerpt = [
  { record: 'act-office-2026-01', source: 'utility_bill_csv', meter: 'CAMX-HQ-01', period: '2026-01', region: 'CAMX', quantity: '3,112 kWh', quality: 'measured' },
  { record: 'act-office-2026-02', source: 'utility_bill_csv', meter: 'CAMX-HQ-01', period: '2026-02', region: 'CAMX', quantity: '2,984 kWh', quality: 'measured' },
  { record: 'act-cloud-2026-01', source: 'aws_cost_and_usage', meter: 'NWPP-AWS-01', period: '2026-01', region: 'NWPP', quantity: '4,406 kWh', quality: 'measured' },
  { record: 'act-cloud-2026-02', source: 'aws_cost_and_usage', meter: 'NWPP-AWS-01', period: '2026-02', region: 'NWPP', quantity: '4,212 kWh', quality: 'measured' },
];

const importRuns = [
  { batch: 'imp-2026-02-03-001', run: 'Initial import', rowsReceived: '24', accepted: '24', duplicates: '0', rejected: '0', idempotency: 'utility_bill:CAMX-HQ-01:2026-01' },
  { batch: 'imp-2026-02-03-001-replay', run: 'Replay of same file', rowsReceived: '24', accepted: '0', duplicates: '24', rejected: '0', idempotency: 'same keys reused; no new ledger rows created' },
];

const factors = [
  { factorId: 'grid-camx-2023', source: 'EPA eGRID 2023', region: 'CAMX', value: '0.225 kg CO2e/kWh', valid: '2023-01-01 to 2026-12-31', snapshot: 'snap-q1-2026-001' },
  { factorId: 'grid-nwpp-2023', source: 'EPA eGRID 2023', region: 'NWPP', value: '0.252 kg CO2e/kWh', valid: '2023-01-01 to 2026-12-31', snapshot: 'snap-q1-2026-001' },
];

const anomalyTrail = [
  { label: 'Detected', value: 'dq-2026-02-014 · sudden_change · Cloud electricity +61.4% vs trailing 3-month average' },
  { label: 'Assigned', value: 'owner = operator@company.com · status = in_progress' },
  { label: 'Commented', value: 'AWS billing export confirms one-time model training workload in September 2026' },
  { label: 'Reviewed', value: 'manager@company.com accepted explanation and kept the record in scope' },
  { label: 'Outcome', value: 'Marked valid, alert closed, totals unchanged, audit log retained' },
];

const ledgerRows = [
  { ledgerId: 'calc-camx-annual-2026', activity: '36,556 kWh office electricity', formula: '36,556 × 0.225', result: '8.22 tCO2e', actor: 'operator@company.com', timestamp: '2026-02-03T14:22:00Z', status: 'draft' },
  { ledgerId: 'calc-nwpp-annual-2026', activity: '52,155 kWh cloud electricity', formula: '52,155 × 0.252', result: '13.14 tCO2e', actor: 'operator@company.com', timestamp: '2026-02-03T14:23:00Z', status: 'draft' },
];

const approvalTrail = [
  { state: 'Prepared', actor: 'operator@company.com', when: '2026-02-03T14:30:00Z', note: 'Imported 24 monthly activity records and reviewed factor selection.' },
  { state: 'Reviewed', actor: 'manager@company.com', when: '2026-02-04T09:15:00Z', note: 'Confirmed line items against utility invoices and cloud billing export.' },
  { state: 'Approved', actor: 'cfo@company.com', when: '2026-02-04T11:00:00Z', note: 'Approved for stakeholder review and locked factor snapshot snap-q1-2026-001.' },
  { state: 'Exported', actor: 'cfo@company.com', when: '2026-02-04T11:06:00Z', note: 'Generated stakeholder PDF and checksum for reconciliation.' },
];

const governanceResponses = [
  {
    title: 'GET /api/governance/export',
    copy: 'Returns exported_at, tenant_id, tenant_name, users, activities, calculation_ledger, and change_log in one JSON package.',
    sample: ['"exported_at": "2026-02-04T11:08:00Z"', '"tenant_id": "tenant-demo-001"', '"tenant_name": "Redacted SaaS Co."', '"activities": 24 line items', '"calculation_ledger": 24 immutable entries'],
  },
  {
    title: 'POST /api/governance/delete-request',
    copy: 'Admin-only. Starts a 30-day retention window and records the request in the change log before permanent deletion.',
    sample: ['"status": "deletion_requested"', '"retention_days": 30', '"deletion_date": "2026-03-06"', '"message": "Export your data before deletion via GET /api/governance/export."'],
  },
];

export default function EvidencePage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <SiteNav />

      <main className="mx-auto max-w-6xl px-6 py-16">
        <h1 className="text-3xl font-bold text-white">Redacted Evidence Pack</h1>
        <p className="mt-3 max-w-3xl text-gray-400">
          This packet is the drill-down behind the public case study on{' '}
          <Link href="/case-study" className="text-primary-400 hover:underline">/case-study</Link>.
          It shows how the same totals move from source activity rows to factor provenance, immutable ledger entries,
          approval states, and export reconciliation. Values are redacted or summarized where customer-sensitive detail would normally appear.
        </p>
        <div className="mt-6 flex flex-wrap gap-4 text-sm">
          <Link
            href="/redacted-audit-packet.pdf"
            className="rounded-lg bg-primary-600 px-5 py-2.5 font-medium text-white transition hover:bg-primary-500"
          >
            Download Sample PDF Packet
          </Link>
          <Link
            href="/case-study"
            className="rounded-lg border border-gray-700 px-5 py-2.5 font-medium text-gray-300 transition hover:border-gray-500 hover:text-white"
          >
            Review Case Study
          </Link>
        </div>

        <div className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {[
            { label: 'Activity rows in batch', value: '24' },
            { label: 'Factor regions locked', value: '2' },
            { label: 'Approval states recorded', value: '4' },
            { label: 'Export drift', value: '0.00%' },
          ].map((card) => (
            <div key={card.label} className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <div className="text-2xl font-bold text-primary-400">{card.value}</div>
              <div className="mt-1 text-xs text-gray-500">{card.label}</div>
            </div>
          ))}
        </div>

        <section className="mt-12 rounded-2xl border border-gray-800 bg-gray-800/20 p-6">
          <h2 className="text-xs font-medium uppercase tracking-widest text-primary-400">What This Proves</h2>
          <div className="mt-4 grid gap-3 sm:grid-cols-2">
            {[
              'Source rows can be tied to exact meters, regions, and reporting periods.',
              'Replay of the same import batch does not create duplicate records.',
              'Emission factors are versioned, region-specific, and locked into a factor snapshot.',
              'Calculated totals reconcile to exported totals through a deterministic checksum.',
              'Approval, export, and deletion flows produce durable machine-readable records.',
              'The public case study numbers can be traced back to activity-level evidence.',
            ].map((item) => (
              <div key={item} className="flex items-start gap-2 text-sm text-gray-300">
                <span className="mt-0.5 text-primary-400">&#10003;</span>
                {item}
              </div>
            ))}
          </div>
        </section>

        <section className="mt-16">
          <h2 className="text-xl font-semibold text-white">1. Source Activity Register</h2>
          <p className="mt-2 text-sm text-gray-500">Excerpt from the 24-row activity batch used for the public case study totals.</p>
          <div className="mt-6 overflow-hidden rounded-xl border border-gray-800">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-800 bg-gray-800/50 text-left text-xs text-gray-500">
                  <th className="px-4 py-3">Record</th>
                  <th className="px-4 py-3">Source</th>
                  <th className="px-4 py-3">Meter / Connector</th>
                  <th className="px-4 py-3">Period</th>
                  <th className="px-4 py-3">Region</th>
                  <th className="px-4 py-3">Quantity</th>
                  <th className="px-4 py-3">Quality</th>
                </tr>
              </thead>
              <tbody className="text-gray-300">
                {activityExcerpt.map((row) => (
                  <tr key={row.record} className="border-b border-gray-800/30">
                    <td className="px-4 py-2 font-mono text-xs text-white">{row.record}</td>
                    <td className="px-4 py-2">{row.source}</td>
                    <td className="px-4 py-2">{row.meter}</td>
                    <td className="px-4 py-2">{row.period}</td>
                    <td className="px-4 py-2">{row.region}</td>
                    <td className="px-4 py-2">{row.quantity}</td>
                    <td className="px-4 py-2">
                      <span className="rounded bg-green-500/10 px-2 py-0.5 text-[10px] font-medium text-green-400">{row.quality}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="mt-4 grid gap-4 md:grid-cols-2">
            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <div className="text-sm font-semibold text-white">Annual Office Rollup</div>
              <div className="mt-2 text-xs text-gray-400">36,556 kWh x 0.225 kg CO2e/kWh = <span className="text-white">8.22 tCO2e</span></div>
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <div className="text-sm font-semibold text-white">Annual Cloud Rollup</div>
              <div className="mt-2 text-xs text-gray-400">52,155 kWh x 0.252 kg CO2e/kWh = <span className="text-white">13.14 tCO2e</span></div>
            </div>
          </div>
        </section>

        <section className="mt-16">
          <h2 className="text-xl font-semibold text-white">2. Import Replay and Duplicate Protection</h2>
          <p className="mt-2 text-sm text-gray-500">Import batches use stable idempotency keys so retries do not create duplicate activities or duplicate ledger rows.</p>
          <div className="mt-6 overflow-hidden rounded-xl border border-gray-800">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-800 bg-gray-800/50 text-left text-xs text-gray-500">
                  <th className="px-4 py-3">Batch</th>
                  <th className="px-4 py-3">Run</th>
                  <th className="px-4 py-3">Rows received</th>
                  <th className="px-4 py-3">Accepted</th>
                  <th className="px-4 py-3">Duplicates</th>
                  <th className="px-4 py-3">Rejected</th>
                  <th className="px-4 py-3">Idempotency proof</th>
                </tr>
              </thead>
              <tbody className="text-gray-300">
                {importRuns.map((run) => (
                  <tr key={run.batch} className="border-b border-gray-800/30">
                    <td className="px-4 py-2 font-mono text-xs text-white">{run.batch}</td>
                    <td className="px-4 py-2">{run.run}</td>
                    <td className="px-4 py-2">{run.rowsReceived}</td>
                    <td className="px-4 py-2">{run.accepted}</td>
                    <td className="px-4 py-2">{run.duplicates}</td>
                    <td className="px-4 py-2">{run.rejected}</td>
                    <td className="px-4 py-2 text-xs text-gray-400">{run.idempotency}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>

        <section className="mt-16">
          <h2 className="text-xl font-semibold text-white">3. Anomaly-to-Action Example</h2>
          <p className="mt-2 text-sm text-gray-500">
            One alert from the same reporting workflow, shown here to demonstrate assignment, comments, review, and closure.
          </p>
          <div className="mt-6 rounded-2xl border border-gray-800 bg-gray-800/20 p-6">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="rounded-xl border border-gray-800 bg-gray-900/40 p-4">
                <div className="text-xs font-medium uppercase tracking-widest text-primary-400">Alert snapshot</div>
                <div className="mt-3 space-y-2 text-sm text-gray-300">
                  <div className="flex items-center justify-between gap-4">
                    <span className="text-gray-500">Alert ID</span>
                    <span className="font-mono text-xs">dq-2026-02-014</span>
                  </div>
                  <div className="flex items-center justify-between gap-4">
                    <span className="text-gray-500">Category</span>
                    <span>sudden_change</span>
                  </div>
                  <div className="flex items-center justify-between gap-4">
                    <span className="text-gray-500">Severity</span>
                    <span>high</span>
                  </div>
                  <div className="flex items-center justify-between gap-4">
                    <span className="text-gray-500">Source record</span>
                    <span className="font-mono text-xs">act-cloud-2026-09</span>
                  </div>
                  <div className="flex items-center justify-between gap-4">
                    <span className="text-gray-500">Expected</span>
                    <span>3,210 kWh</span>
                  </div>
                  <div className="flex items-center justify-between gap-4">
                    <span className="text-gray-500">Actual</span>
                    <span>5,182 kWh</span>
                  </div>
                  <div className="flex items-center justify-between gap-4">
                    <span className="text-gray-500">Deviation</span>
                    <span className="text-amber-300">+61.4%</span>
                  </div>
                </div>
              </div>

              <div className="rounded-xl border border-gray-800 bg-gray-900/40 p-4">
                <div className="text-xs font-medium uppercase tracking-widest text-primary-400">Resolution trail</div>
                <div className="mt-3 space-y-3">
                  {anomalyTrail.map((item) => (
                    <div key={item.label}>
                      <div className="text-xs font-semibold text-white">{item.label}</div>
                      <div className="mt-1 text-xs text-gray-400">{item.value}</div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="mt-16">
          <h2 className="text-xl font-semibold text-white">4. Factor Provenance and Snapshot Lock</h2>
          <p className="mt-2 text-sm text-gray-500">Factors are stored with source metadata and then frozen into a period-specific snapshot for reproducibility.</p>
          <div className="mt-6 overflow-hidden rounded-xl border border-gray-800">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-800 bg-gray-800/50 text-left text-xs text-gray-500">
                  <th className="px-4 py-3">Factor ID</th>
                  <th className="px-4 py-3">Source</th>
                  <th className="px-4 py-3">Region</th>
                  <th className="px-4 py-3">Value</th>
                  <th className="px-4 py-3">Validity</th>
                  <th className="px-4 py-3">Locked snapshot</th>
                </tr>
              </thead>
              <tbody className="text-gray-300">
                {factors.map((factor) => (
                  <tr key={factor.factorId} className="border-b border-gray-800/30">
                    <td className="px-4 py-2 font-mono text-xs text-white">{factor.factorId}</td>
                    <td className="px-4 py-2">{factor.source}</td>
                    <td className="px-4 py-2">{factor.region}</td>
                    <td className="px-4 py-2">{factor.value}</td>
                    <td className="px-4 py-2">{factor.valid}</td>
                    <td className="px-4 py-2 font-mono text-xs text-gray-400">{factor.snapshot}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="mt-4 rounded-xl border border-gray-800 bg-gray-800/20 p-5 text-sm text-gray-300">
            Snapshot <code className="rounded bg-gray-800 px-1.5 py-0.5 text-xs text-primary-400">snap-q1-2026-001</code> was locked by <span className="text-white">cfo@company.com</span> on <span className="text-white">2026-02-04T11:00:00Z</span>. After this point, the report remained reproducible even if a newer factor release became available.
          </div>
        </section>

        <section className="mt-16">
          <h2 className="text-xl font-semibold text-white">5. Immutable Calculation Ledger</h2>
          <p className="mt-2 text-sm text-gray-500">Each calculation stores the activity reference, factor used, formula, result, actor, and timestamp.</p>
          <div className="mt-6 overflow-hidden rounded-xl border border-gray-800">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-800 bg-gray-800/50 text-left text-xs text-gray-500">
                  <th className="px-4 py-3">Ledger ID</th>
                  <th className="px-4 py-3">Activity</th>
                  <th className="px-4 py-3">Formula</th>
                  <th className="px-4 py-3">Result</th>
                  <th className="px-4 py-3">Actor</th>
                  <th className="px-4 py-3">Timestamp</th>
                  <th className="px-4 py-3">State</th>
                </tr>
              </thead>
              <tbody className="text-gray-300">
                {ledgerRows.map((row) => (
                  <tr key={row.ledgerId} className="border-b border-gray-800/30">
                    <td className="px-4 py-2 font-mono text-xs text-white">{row.ledgerId}</td>
                    <td className="px-4 py-2">{row.activity}</td>
                    <td className="px-4 py-2 font-mono text-xs text-gray-400">{row.formula}</td>
                    <td className="px-4 py-2 text-white">{row.result}</td>
                    <td className="px-4 py-2">{row.actor}</td>
                    <td className="px-4 py-2 font-mono text-xs text-gray-400">{row.timestamp}</td>
                    <td className="px-4 py-2">{row.status}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>

        <section className="mt-16 grid gap-8 lg:grid-cols-[1.2fr,0.8fr]">
          <div>
            <h2 className="text-xl font-semibold text-white">6. Approval Trail and Export Reconciliation</h2>
            <p className="mt-2 text-sm text-gray-500">Ownership and status changes stay visible through preparation, review, approval, and export.</p>
            <div className="mt-6 space-y-3">
              {approvalTrail.map((step, index) => (
                <div key={step.state} className="flex items-start gap-4">
                  <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary-600/10 text-xs font-bold text-primary-400">{index + 1}</div>
                  <div className="flex-1 rounded-lg border border-gray-800 bg-gray-800/20 px-4 py-3">
                    <div className="text-sm font-semibold text-white">{step.state}</div>
                    <div className="mt-1 text-xs text-gray-400">{step.note}</div>
                    <div className="mt-1 text-[10px] font-mono text-gray-600">{step.actor} · {step.when}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="rounded-2xl border border-primary-600/20 bg-primary-600/5 p-6">
            <h3 className="text-sm font-semibold text-white">Export Record</h3>
            <div className="mt-4 space-y-2 text-sm text-gray-300">
              {[
                ['Report type', 'california'],
                ['Purpose', 'stakeholder_review'],
                ['Scope 1 at export', '0.00 tCO2e'],
                ['Scope 2 at export', '21.368160 tCO2e'],
                ['Scope 3 at export', '0.00 tCO2e'],
                ['Total at export', '21.368160 tCO2e'],
                ['Checksum', 'b7fd1f798bcfe0b67d43496696a2e52e'],
              ].map(([label, value]) => (
                <div key={label} className="flex items-center justify-between gap-4">
                  <span className="text-gray-500">{label}</span>
                  <span className={label === 'Checksum' ? 'font-mono text-xs' : ''}>{value}</span>
                </div>
              ))}
              <div className="mt-4 rounded-lg border border-green-500/20 bg-green-500/5 px-3 py-2 text-xs text-green-300">
                Reconciliation result: checksum match = true, drift = 0.00%, on-screen totals equal exported totals.
              </div>
            </div>
          </div>
        </section>

        <section className="mt-16">
          <h2 className="text-xl font-semibold text-white">7. Governance Execution Proof</h2>
          <p className="mt-2 text-sm text-gray-500">The same governance APIs documented in the trust center return machine-readable export and deletion metadata.</p>
          <div className="mt-6 grid gap-4 lg:grid-cols-2">
            {governanceResponses.map((item) => (
              <div key={item.title} className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
                <h3 className="text-sm font-semibold text-white">{item.title}</h3>
                <p className="mt-2 text-xs leading-relaxed text-gray-400">{item.copy}</p>
                <div className="mt-4 rounded-lg bg-gray-900 p-4">
                  {item.sample.map((line) => (
                    <div key={line} className="font-mono text-xs text-gray-300">{line}</div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </section>

        <div className="mt-14 flex flex-wrap gap-4 text-sm">
          <Link href="/case-study" className="text-primary-400 hover:underline">Case Study</Link>
          <Link href="/methodology" className="text-primary-400 hover:underline">Methodology</Link>
          <Link href="/architecture" className="text-primary-400 hover:underline">Architecture</Link>
          <Link href="/trust" className="text-primary-400 hover:underline">Trust Center</Link>
          <Link href="/operations" className="text-primary-400 hover:underline">Operations Proof</Link>
        </div>
      </main>
    </div>
  );
}
