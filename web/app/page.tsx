import Link from 'next/link';
import type { Metadata } from 'next';
import { SiteNav } from './components/SiteNav';
import { SiteFooter } from './components/SiteFooter';
import { FadeIn } from './components/FadeIn';
import { DashboardPreview } from './components/DashboardPreview';
import { LeadCaptureForm } from '@/components/LeadCaptureForm';

export const metadata: Metadata = {
  title: 'OffGridFlow | Carbon Compliance Without the Big 4 Price Tag',
  description:
    'Upload your utility data, see your carbon footprint calculated with GHG Protocol methodology for free, and export an audit-ready CSRD / SEC / SB 253 report for $149.',
  openGraph: {
    title: 'OffGridFlow | Carbon Compliance Without the Big 4 Price Tag',
    description:
      'Enterprise carbon compliance made accessible. Track emissions, generate reports, stay compliant.',
    type: 'website',
    url: 'https://off-grid-flow.com',
  },
};

export default function HomePage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <SiteNav />

      {/* ================================================================== */}
      {/* HERO — Lead with pain, not features                                */}
      {/* ================================================================== */}
      <section className="relative overflow-hidden px-6 pb-24 pt-20 sm:pt-28">
        {/* Subtle grid background */}
        <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(to_right,#1d294010_1px,transparent_1px),linear-gradient(to_bottom,#1d294010_1px,transparent_1px)] bg-[size:64px_64px]" />

        <div className="relative mx-auto max-w-5xl">
          {/* Regulatory urgency bar */}
          <div className="mb-8 flex justify-center">
            <div className="inline-flex items-center gap-2 rounded-full border border-amber-500/20 bg-amber-500/5 px-4 py-1.5 text-xs text-amber-400">
              <span className="h-1.5 w-1.5 rounded-full bg-amber-400" />
              SB 253 reporting deadlines active — applies to companies with {'>'}$1B revenue doing business in California
            </div>
          </div>

          <h1 className="text-center text-4xl font-bold leading-[1.15] tracking-tight text-white sm:text-5xl lg:text-6xl">
            Stop overpaying for
            <br />
            <span className="bg-gradient-to-r from-primary-400 to-emerald-300 bg-clip-text text-transparent">
              carbon compliance
            </span>
          </h1>

          <p className="mx-auto mt-6 max-w-2xl text-center text-lg leading-relaxed text-gray-400">
            Big 4 firms charge $50,000&ndash;$200,000 for carbon audits. Upload your
            utility data, see your footprint calculated with{' '}
            <Link href="/trust" className="text-primary-400 hover:underline">GHG Protocol methodology</Link>{' '}
            for free, and export an audit-ready report for{' '}
            <strong className="text-white">$149</strong>.
          </p>

          <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Link
              href="/register"
              className="rounded-lg bg-primary-600 px-8 py-3.5 text-base font-semibold text-white shadow-lg shadow-primary-600/20 transition hover:bg-primary-500 hover:shadow-primary-500/30"
            >
              Start free
            </Link>
            <Link
              href="/register"
              className="rounded-lg border border-gray-700 px-8 py-3.5 text-base font-medium text-gray-300 transition hover:border-gray-500 hover:text-white"
            >
              Review Workflow
            </Link>
            <Link
              href="/login"
              className="text-sm font-medium text-gray-500 transition hover:text-white"
            >
              Already have an account? Log in
            </Link>
          </div>

          {/* Product Preview */}
          <FadeIn delay={0.3}>
            <DashboardPreview />
          </FadeIn>

          {/* Trust bar */}
          <div className="mt-14 flex flex-wrap items-center justify-center gap-x-8 gap-y-3 text-xs text-gray-500">
            <span className="flex items-center gap-1.5">
              <svg className="h-4 w-4 text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>
              GHG Protocol methodology documented
            </span>
            <span className="flex items-center gap-1.5">
              <svg className="h-4 w-4 text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
              EPA eGRID + DEFRA + IEA Factors
            </span>
            <span className="flex items-center gap-1.5">
              <svg className="h-4 w-4 text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" /></svg>
              SOC 2 Type I In Progress
            </span>
            <span className="flex items-center gap-1.5">
              <svg className="h-4 w-4 text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064" /></svg>
              55+ Country Grid Factors
            </span>
          </div>
        </div>
      </section>

      {/* ================================================================== */}
      {/* CASE STUDY — Social proof (the #1 gap)                             */}
      {/* ================================================================== */}
      <section className="border-t border-gray-800/50 px-6 py-20">
        <FadeIn><div className="mx-auto max-w-4xl">
          <div className="rounded-2xl border border-gray-800 bg-gradient-to-br from-gray-800/50 to-gray-900/50 p-8 sm:p-12">
            <div className="mb-6 text-xs font-medium uppercase tracking-widest text-primary-400">
              Case Study
            </div>
            <h2 className="text-2xl font-bold text-white sm:text-3xl">
              How a SaaS company calculated its full carbon footprint in under 2 hours
            </h2>
            <div className="mt-6 grid gap-8 sm:grid-cols-3">
              <div>
                <div className="text-3xl font-bold text-primary-400">86%</div>
                <div className="mt-1 text-sm text-gray-400">Cost reduction vs Big 4 quote</div>
              </div>
              <div>
                <div className="text-3xl font-bold text-primary-400">&lt;2 hrs</div>
                <div className="mt-1 text-sm text-gray-400">From CSV upload to audit-ready report</div>
              </div>
              <div>
                <div className="text-3xl font-bold text-primary-400">24</div>
                <div className="mt-1 text-sm text-gray-400">Data points across Scope 2 sources</div>
              </div>
            </div>
            <p className="mt-6 text-sm leading-relaxed text-gray-400">
              A small technology company with office space and cloud infrastructure uploaded 12 months
              of utility data via CSV. OffGridFlow applied EPA eGRID emission factors for two regions,
              calculated location-based Scope 2 emissions, and generated a compliance-ready PDF
              report&mdash;all without a sustainability consultant.
            </p>
            <Link
              href="/pricing"
              className="mt-6 inline-flex items-center gap-1 text-sm font-medium text-primary-400 transition hover:text-primary-300"
            >
              Read the full case study
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" /></svg>
            </Link>
            <div className="mt-3">
              <Link
                href="/trust"
                className="inline-flex items-center gap-1 text-sm font-medium text-gray-300 transition hover:text-white"
              >
                Review the redacted evidence pack
                <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" /></svg>
              </Link>
            </div>
          </div>
        </div></FadeIn>
      </section>

      {/* ================================================================== */}
      {/* PAIN → SOLUTION — Outcome-focused messaging                        */}
      {/* ================================================================== */}
      <section className="border-t border-gray-800/50 px-6 py-20">
        <div className="mx-auto max-w-5xl">
          <h2 className="mb-4 text-center text-3xl font-bold text-white">
            What if carbon compliance took days, not months?
          </h2>
          <p className="mx-auto mb-16 max-w-2xl text-center text-gray-400">
            Three steps. No consultants. No six-figure invoices.
          </p>

          <div className="grid gap-6 md:grid-cols-3">
            {[
              {
                step: '01',
                title: 'Connect your data',
                pain: 'Stop copying numbers from spreadsheets.',
                solution: 'Import from AWS, Azure, GCP, SAP, utility bills, or CSV. Cloud connectors pull carbon data automatically with built-in retry and idempotency.',
              },
              {
                step: '02',
                title: 'Calculate emissions',
                pain: 'Stop guessing at emission factors.',
                solution: 'The calculation engine applies factor sources documented in our methodology library across Scope 1, 2, and 3. Every calculation is traceable to its source factor and reporting method.',
              },
              {
                step: '03',
                title: 'Generate reports',
                pain: 'Stop paying $50K for a PDF.',
                solution: 'Export audit-ready compliance reports in PDF, XBRL, and Excel. Pre-mapped to CSRD, SEC, SB 253, CBAM, and IFRS S2 frameworks.',
              },
            ].map((s) => (
              <div key={s.step} className="rounded-xl border border-gray-800 bg-dark-800 p-7">
                <div className="mb-4 flex h-9 w-9 items-center justify-center rounded-lg bg-primary-600/10 text-sm font-bold text-primary-400">
                  {s.step}
                </div>
                <h3 className="mb-2 text-lg font-semibold text-white">{s.title}</h3>
                <p className="mb-3 text-sm font-medium text-amber-400/80">{s.pain}</p>
                <p className="text-sm leading-relaxed text-gray-400">{s.solution}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ================================================================== */}
      {/* WHO IT'S FOR — Persona sections                                     */}
      {/* ================================================================== */}
      <section className="border-t border-gray-800/50 px-6 py-20">
        <div className="mx-auto max-w-5xl">
          <h2 className="mb-16 text-center text-3xl font-bold text-white">Built for your role</h2>
          <div className="grid gap-6 md:grid-cols-2 xl:grid-cols-4">
            {[
              {
                role: 'CFOs & Finance',
                headline: 'Close your audit without open questions',
                href: '/for-cfos',
                points: ['Audit-ready reports that pass third-party review', 'Cost control: 90% less than Big 4 engagements', 'Full calculation transparency for investor confidence'],
              },
              {
                role: 'Sustainability Managers',
                headline: 'Accurate data without the manual work',
                href: '/for-sustainability-managers',
                points: ['EPA eGRID, DEFRA, IEA factors — no manual lookups', 'Scope 1, 2, 3 in one platform', 'Track progress against reduction targets'],
              },
              {
                role: 'Procurement',
                headline: 'Turn supplier data into Scope 3 coverage',
                href: '/for-procurement',
                points: ['Supplier requests and CSV templates in one workflow', 'Spend-based to supplier-specific progression', 'CBAM and supplier-emissions reporting support'],
              },
              {
                role: 'Compliance Owners',
                headline: 'Run disclosure review with evidence, not email',
                href: '/for-compliance-owners',
                points: ['Framework-specific draft outputs mapped to one workspace', 'Review-ready evidence for finance, legal, and audit', 'Approval visibility before filing deadlines'],
              },
            ].map((p) => (
              <div key={p.role} className="rounded-xl border border-gray-800 bg-dark-800 p-7">
                <div className="mb-3 text-xs font-semibold uppercase tracking-wider text-primary-400">{p.role}</div>
                <h3 className="mb-4 text-base font-semibold text-white">{p.headline}</h3>
                <ul className="space-y-2.5">
                  {p.points.map((pt) => (
                    <li key={pt} className="flex items-start gap-2 text-sm text-gray-400">
                      <span className="mt-0.5 text-primary-500">&#10003;</span>
                      {pt}
                    </li>
                  ))}
                </ul>
                <Link
                  href={p.href}
                  className="mt-5 inline-flex items-center gap-1 text-sm font-medium text-primary-400 transition hover:text-primary-300"
                >
                  See the role page
                  <span aria-hidden="true">&rarr;</span>
                </Link>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ================================================================== */}
      {/* INTEGRATIONS                                                        */}
      {/* ================================================================== */}
      <section className="border-t border-gray-800/50 px-6 py-20">
        <div className="mx-auto max-w-4xl text-center">
          <h2 className="mb-4 text-3xl font-bold text-white">Connects to your stack</h2>
          <p className="mb-12 text-gray-400">Pull carbon data directly from the tools you already use</p>
          <div className="flex flex-wrap items-center justify-center gap-6">
            {['AWS', 'Azure', 'Google Cloud', 'SAP', 'Utility APIs', 'CSV / Excel'].map((name) => (
              <div
                key={name}
                className="rounded-lg border border-gray-800 bg-dark-800 px-6 py-3 text-sm font-medium text-gray-300"
              >
                {name}
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ================================================================== */}
      {/* ROI CALCULATOR (Static version)                                     */}
      {/* ================================================================== */}
      <section className="border-t border-gray-800/50 px-6 py-20">
        <div className="mx-auto max-w-3xl">
          <div className="rounded-2xl border border-primary-600/20 bg-primary-600/5 p-8 sm:p-12">
            <h2 className="mb-2 text-center text-2xl font-bold text-white">How much could you save?</h2>
            <p className="mb-8 text-center text-sm text-gray-400">
              Companies using OffGridFlow save 80&ndash;95% compared to Big 4 consulting engagements
            </p>
            <div className="grid gap-6 sm:grid-cols-3">
              <div className="text-center">
                <div className="text-xs text-gray-500">Big 4 Average</div>
                <div className="mt-1 text-2xl font-bold text-gray-400 line-through">$75,000</div>
                <div className="text-xs text-gray-600">per report</div>
              </div>
              <div className="text-center">
                <div className="text-xs text-primary-400">OffGridFlow</div>
                <div className="mt-1 text-2xl font-bold text-white">$149</div>
                <div className="text-xs text-gray-500">per report</div>
              </div>
              <div className="text-center">
                <div className="text-xs text-gray-500">Your Savings</div>
                <div className="mt-1 text-2xl font-bold text-primary-400">$74,851</div>
                <div className="text-xs text-primary-400/60">99% cost reduction</div>
              </div>
            </div>
            <div className="mt-8 text-center">
              <Link
                href="/register"
                className="inline-block rounded-lg bg-primary-600 px-8 py-3 text-sm font-semibold text-white transition hover:bg-primary-500"
              >
                Start free — upload your data
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* ================================================================== */}
      {/* PRICING                                                             */}
      {/* ================================================================== */}
      <section id="pricing" className="border-t border-gray-800/50 px-6 py-20">
        <div className="mx-auto max-w-3xl">
          <h2 className="mb-4 text-center text-3xl font-bold text-white">Simple, honest pricing</h2>
          <p className="mb-14 text-center text-gray-400">Free to upload and see your footprint. Pay only when you export a report.</p>

          <div className="grid gap-6 sm:grid-cols-2">
            {/* Free */}
            <div className="rounded-xl border border-gray-800 bg-dark-800 p-8">
              <h3 className="text-lg font-semibold text-white">Free</h3>
              <div className="mt-3 flex items-baseline gap-1">
                <span className="text-4xl font-bold text-white">$0</span>
              </div>
              <p className="mt-2 text-sm text-gray-500">See your real footprint before you pay anything.</p>
              <ul className="mt-6 space-y-2.5">
                {['Upload utility & energy data', 'Automatic column mapping', 'Emission factors applied (EPA, DEFRA)', 'Full Scope 2 footprint dashboard', 'Data quality anomaly scan'].map((f) => (
                  <li key={f} className="flex items-start gap-2 text-sm text-gray-300">
                    <span className="mt-0.5 text-primary-400">&#10003;</span>
                    {f}
                  </li>
                ))}
              </ul>
              <Link href="/register" className="mt-7 block rounded-lg border border-gray-700 py-2.5 text-center text-sm font-medium text-gray-300 transition hover:border-gray-500 hover:text-white">
                Start free
              </Link>
            </div>

            {/* Per report */}
            <div className="relative rounded-xl border border-primary-600 bg-dark-800 p-8">
              <div className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-primary-600 px-3 py-0.5 text-xs font-medium text-white">
                Pay per report
              </div>
              <h3 className="text-lg font-semibold text-white">Audit-ready report</h3>
              <div className="mt-3 flex items-baseline gap-1">
                <span className="text-4xl font-bold text-white">$149</span>
                <span className="text-sm text-gray-500">/ report</span>
              </div>
              <p className="mt-2 text-sm text-gray-500">One-time. Re-export free for 12 months.</p>
              <ul className="mt-6 space-y-2.5">
                {['Everything in Free', 'Audit-ready PDF & CSV export', 'GHG Protocol + CSRD / ESRS E1', 'Full methodology & source trail', 'Formatted to your framework', 'No subscription, no per-seat fees'].map((f) => (
                  <li key={f} className="flex items-start gap-2 text-sm text-gray-300">
                    <span className="mt-0.5 text-primary-400">&#10003;</span>
                    {f}
                  </li>
                ))}
              </ul>
              <Link href="/register" className="mt-7 block rounded-lg bg-primary-600 py-2.5 text-center text-sm font-medium text-white transition hover:bg-primary-500">
                Upload data to start
              </Link>
            </div>
          </div>
          <p className="mt-8 text-center text-xs text-gray-600">
            Need multi-site or an enterprise agreement? <a href="mailto:contact@off-grid-flow.com?subject=OffGridFlow%20Enterprise" className="text-primary-400 hover:underline">Talk to us</a>.
          </p>
        </div>
      </section>

      {/* ================================================================== */}
      {/* FINAL CTA                                                           */}
      {/* ================================================================== */}
      <section className="border-t border-gray-800/50 px-6 py-24">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="text-3xl font-bold text-white">
            See your carbon footprint in minutes.
          </h2>
          <p className="mt-4 text-gray-400">
            Upload your data free. Export an audit-ready report for $149 when you&apos;re ready.
          </p>
          <div className="mt-8 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Link
              href="/register"
              className="rounded-lg bg-primary-600 px-8 py-3.5 text-base font-semibold text-white shadow-lg shadow-primary-600/20 transition hover:bg-primary-500"
            >
              Start free
            </Link>
            <Link
              href="/login"
              className="text-sm font-medium text-gray-400 transition hover:text-white"
            >
              Sign in &rarr;
            </Link>
          </div>
        </div>
      </section>

      {/* ================================================================== */}
      {/* LEAD CAPTURE — Enterprise + multi-framework inquiries               */}
      {/* ================================================================== */}
      <section className="border-t border-gray-800/50 px-6 py-20">
        <div className="mx-auto max-w-2xl">
          <h2 className="mb-2 text-center text-2xl font-bold text-white">
            Need a tailored walkthrough?
          </h2>
          <p className="mb-8 text-center text-sm text-gray-400">
            Enterprise teams, multi-framework filers, and SAP integration requests.
            We respond within one business day.
          </p>
          <LeadCaptureForm source="homepage_bottom" compact />
        </div>
      </section>

      {/* ================================================================== */}
      {/* FOOTER                                                              */}
      {/* ================================================================== */}
      <SiteFooter />
    </div>
  );
}
