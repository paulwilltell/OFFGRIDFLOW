import Link from 'next/link';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Demo | OffGridFlow',
  description:
    'See how OffGridFlow automates Scope 1, 2, 3 emissions tracking and generates audit-ready compliance reports.',
};

const workflow = [
  {
    step: '1',
    title: 'Ingest Emissions Data',
    desc: 'Upload a CSV of utility bills, energy consumption, or fleet fuel — or connect cloud sources (AWS, Azure, GCP, SAP) for automated ingestion.',
    detail: 'Supported formats: CSV, JSON, Excel. Cloud connectors pull data on schedule with retry and duplicate detection.',
    time: '< 5 minutes for CSV',
  },
  {
    step: '2',
    title: 'Calculate Scopes 1, 2 & 3',
    desc: 'The engine applies 184 emission factors from EPA eGRID, IEA, UK DEFRA, and IPCC to your activity data. Every calculation is recorded in an immutable ledger.',
    detail: 'Scope 2: location-based and market-based methods (GHG Protocol). Scope 3: all 15 GHG Protocol categories with spend-based and activity-based factors.',
    time: 'Instant',
  },
  {
    step: '3',
    title: 'Validate Data Quality',
    desc: 'Anomaly detection flags outliers (z-score > 3), duplicate entries, missing time periods, and sudden changes (> 50% month-over-month).',
    detail: 'Each anomaly shows expected vs. actual values, deviation %, and one-click resolve or dismiss.',
    time: '< 1 minute scan',
  },
  {
    step: '4',
    title: 'Generate Compliance Reports',
    desc: 'Select your framework — CSRD/ESRS E1, SEC Climate, California SB 253, CBAM, or IFRS S2 — and generate an audit-ready report with export to PDF or XBRL.',
    detail: 'Reports include scope totals, factor sources, methodology notes, and reconciliation checksums.',
    time: '< 30 seconds',
  },
  {
    step: '5',
    title: 'Review, Approve & Lock',
    desc: 'Submit reports through an approval workflow: preparer, reviewer, approver. Lock emission factors to the reporting period via factor snapshots for reproducibility.',
    detail: 'Every approval action is recorded with timestamp, actor, and notes. Locked calculations cannot be modified.',
    time: 'Depends on your review process',
  },
];

const frameworks = [
  { name: 'CSRD / ESRS E1', region: 'EU', desc: 'Corporate Sustainability Reporting Directive — mandatory for EU companies and large non-EU companies operating in the EU.' },
  { name: 'SEC Climate Disclosure', region: 'US', desc: 'SEC climate-related disclosure rules for US public companies — Scope 1, 2, and material Scope 3.' },
  { name: 'California SB 253', region: 'CA', desc: 'Climate Corporate Data Accountability Act — Scope 1, 2, 3 reporting for companies doing business in California with > $1B revenue.' },
  { name: 'EU CBAM', region: 'EU', desc: 'Carbon Border Adjustment Mechanism — embedded emissions reporting for imports of cement, steel, aluminum, fertilizer, electricity, hydrogen.' },
  { name: 'IFRS S2', region: 'Global', desc: 'International Sustainability Standards Board — climate-related disclosures aligned with TCFD for global capital markets.' },
];

const factorSources = [
  { name: 'EPA eGRID 2023', scope: '27 US subregions', type: 'Scope 2 grid electricity' },
  { name: 'IEA 2023', scope: '55+ countries', type: 'Scope 2 international grids' },
  { name: 'UK DEFRA 2024', scope: 'Global', type: 'Scope 1, 2, 3 fuels, transport, waste' },
  { name: 'IPCC AR6 GWP-100', scope: 'Global', type: 'Refrigerant GWP values' },
  { name: 'GHG Protocol', scope: '15 categories', type: 'Scope 3 methodology and factors' },
];

export default function DemoPage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      {/* Header */}
      <header className="sticky top-0 z-50 border-b border-gray-800/50 bg-dark-900/80 backdrop-blur-md">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
          <Link href="/" className="text-xl font-bold text-white">
            OffGridFlow
          </Link>
          <div className="flex items-center gap-4">
            <Link href="/" className="text-sm text-gray-400 hover:text-white">
              Home
            </Link>
            <Link href="/login" className="text-sm text-gray-400 hover:text-white">
              Log In
            </Link>
            <Link
              href="/register"
              className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white hover:bg-primary-500"
            >
              Start Trial
            </Link>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-5xl px-6 py-16">
        {/* Hero */}
        <section className="mb-16 text-center">
          <h1 className="text-3xl font-bold text-white sm:text-4xl">
            How OffGridFlow Works
          </h1>
          <p className="mx-auto mt-4 max-w-2xl text-lg text-gray-400">
            From raw utility data to audit-ready compliance report in under 2 hours.
            No consultants. No six-figure invoices.
          </p>
        </section>

        {/* Core Workflow */}
        <section className="mb-20">
          <h2 className="mb-8 text-xs font-medium uppercase tracking-widest text-primary-400">
            The 5-Step Workflow
          </h2>
          <div className="space-y-6">
            {workflow.map((step) => (
              <div
                key={step.step}
                className="rounded-xl border border-gray-800 bg-gray-800/30 p-6"
              >
                <div className="flex items-start gap-4">
                  <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary-600/10 text-lg font-bold text-primary-400">
                    {step.step}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <h3 className="text-lg font-semibold text-white">{step.title}</h3>
                      <span className="hidden text-xs text-gray-500 sm:inline">{step.time}</span>
                    </div>
                    <p className="mt-2 text-sm text-gray-400">{step.desc}</p>
                    <p className="mt-2 text-xs text-gray-500">{step.detail}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Compliance Frameworks */}
        <section className="mb-20">
          <h2 className="mb-2 text-xs font-medium uppercase tracking-widest text-primary-400">
            Supported Frameworks
          </h2>
          <p className="mb-6 text-sm text-gray-500">
            OffGridFlow generates reports for five major disclosure regimes.
          </p>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {frameworks.map((fw) => (
              <div
                key={fw.name}
                className="rounded-xl border border-gray-800 bg-gray-800/30 p-5"
              >
                <div className="flex items-center justify-between">
                  <h3 className="text-sm font-semibold text-white">{fw.name}</h3>
                  <span className="rounded bg-gray-700 px-2 py-0.5 text-[10px] font-medium text-gray-400">
                    {fw.region}
                  </span>
                </div>
                <p className="mt-2 text-xs text-gray-500 leading-relaxed">{fw.desc}</p>
              </div>
            ))}
          </div>
        </section>

        {/* Factor Sources */}
        <section className="mb-20">
          <h2 className="mb-2 text-xs font-medium uppercase tracking-widest text-primary-400">
            Emission Factor Sources
          </h2>
          <p className="mb-6 text-sm text-gray-500">
            184 verified emission factors with full provenance metadata.
          </p>
          <div className="overflow-hidden rounded-xl border border-gray-800">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-800 bg-gray-800/50">
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Source</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Coverage</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Used For</th>
                </tr>
              </thead>
              <tbody>
                {factorSources.map((f) => (
                  <tr key={f.name} className="border-b border-gray-800/50">
                    <td className="px-4 py-3 font-medium text-white">{f.name}</td>
                    <td className="px-4 py-3 text-gray-400">{f.scope}</td>
                    <td className="px-4 py-3 text-gray-500">{f.type}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <p className="mt-3 text-xs text-gray-600">
            Full methodology documentation:{' '}
            <Link href="/methodology" className="text-primary-400 hover:underline">
              /methodology
            </Link>
          </p>
        </section>

        {/* What You Get */}
        <section className="mb-20">
          <h2 className="mb-6 text-xs font-medium uppercase tracking-widest text-primary-400">
            What the Platform Includes
          </h2>
          <div className="grid gap-4 sm:grid-cols-2">
            {[
              { title: 'Scope 1, 2 & 3 Calculators', desc: 'Activity-based, spend-based, and supplier-specific calculation methods per GHG Protocol.' },
              { title: 'Immutable Calculation Ledger', desc: 'Every calculation recorded with formula, factor source, timestamp, and user. Locked on approval.' },
              { title: 'Factor Version Locking', desc: 'Freeze emission factors to your reporting period so auditors can reproduce any prior calculation.' },
              { title: 'Data Quality Engine', desc: 'Anomaly detection for outliers, duplicates, missing periods, and sudden changes.' },
              { title: 'Approval Workflow', desc: 'Draft, submit, review, approve, reject — with full actor attribution and notes at each stage.' },
              { title: 'Alert Action System', desc: 'Every data quality issue gets assign, comment, resolve, escalate, and dismiss actions.' },
              { title: 'PDF & XBRL Export', desc: 'Compliance reports export to PDF for stakeholders and XBRL for regulatory submission.' },
              { title: 'Cloud Connectors', desc: 'Automated data ingestion from AWS, Azure, GCP, SAP, and utility providers.' },
            ].map((item) => (
              <div key={item.title} className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
                <h3 className="text-sm font-semibold text-white">{item.title}</h3>
                <p className="mt-1 text-xs text-gray-500">{item.desc}</p>
              </div>
            ))}
          </div>
        </section>

        {/* Trust — honest, no fabrication */}
        <section className="mb-20">
          <h2 className="mb-6 text-xs font-medium uppercase tracking-widest text-primary-400">
            Trust & Security
          </h2>
          <div className="grid gap-4 sm:grid-cols-3">
            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <h3 className="text-sm font-semibold text-white">GHG Protocol Compliant</h3>
              <p className="mt-1 text-xs text-gray-500">
                Scope 1, 2 (location + market-based), and all 15 Scope 3 categories aligned with GHG Protocol Corporate Standard.
              </p>
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <h3 className="text-sm font-semibold text-white">SOC 2 Type I In Progress</h3>
              <p className="mt-1 text-xs text-gray-500">
                Targeting Q3 2026. Type II targeted Q1 2027. ISO 27001 targeted Q2 2027.
              </p>
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <h3 className="text-sm font-semibold text-white">Tenant-Isolated Architecture</h3>
              <p className="mt-1 text-xs text-gray-500">
                RBAC with admin/user/viewer roles. JWT sessions. Account lockout. CSRF protection. Immutable audit logs.
              </p>
            </div>
          </div>
          <div className="mt-4 flex flex-wrap gap-3 text-xs text-gray-600">
            <Link href="/trust" className="hover:text-primary-400">Trust Center</Link>
            <span>|</span>
            <Link href="/security" className="hover:text-primary-400">Security</Link>
            <span>|</span>
            <Link href="/privacy" className="hover:text-primary-400">Privacy Policy</Link>
            <span>|</span>
            <Link href="/methodology" className="hover:text-primary-400">Methodology</Link>
            <span>|</span>
            <Link href="/status" className="hover:text-primary-400">Status</Link>
          </div>
        </section>

        {/* CTA */}
        <section className="rounded-xl border border-primary-600/20 bg-primary-600/5 p-10 text-center">
          <h2 className="text-2xl font-bold text-white">
            Ready to see it with your data?
          </h2>
          <p className="mx-auto mt-3 max-w-lg text-sm text-gray-400">
            Upload a CSV and generate your first compliance report in under 2 hours.
            Or schedule a call to discuss your specific regulatory requirements.
          </p>
          <div className="mt-6 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Link
              href="/register"
              className="rounded-lg bg-primary-600 px-8 py-3 text-base font-semibold text-white hover:bg-primary-500"
            >
              Start Free Trial
            </Link>
            <Link
              href="mailto:contact@off-grid-flow.com?subject=OffGridFlow%20Demo%20Request"
              className="rounded-lg border border-gray-700 px-8 py-3 text-base font-medium text-gray-300 hover:border-gray-500 hover:text-white"
            >
              Schedule a Call
            </Link>
            <Link
              href="/login"
              className="text-sm text-gray-500 hover:text-white"
            >
              Already have an account? Log in
            </Link>
          </div>
        </section>
      </main>
    </div>
  );
}
