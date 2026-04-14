import Link from 'next/link';
import type { Metadata } from 'next';
import { SiteNav } from './components/SiteNav';
import { FadeIn } from './components/FadeIn';
import { DashboardPreview } from './components/DashboardPreview';

export const metadata: Metadata = {
  title: 'OffGridFlow | Carbon Compliance Without the Big 4 Price Tag',
  description:
    'Automate Scope 1, 2, 3 emissions tracking and generate audit-ready compliance reports for CSRD, SEC, California SB 253, CBAM, and IFRS S2. Starting at $4,800/year.',
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
              SB 253 reporting deadline approaching — 5,300+ companies must comply
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
            Big 4 firms charge $50,000&ndash;$200,000 for carbon audits. OffGridFlow delivers
            the same GHG Protocol-compliant calculations and audit-ready reports&mdash;starting
            at <strong className="text-white">$6,500/year</strong>.
          </p>

          <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Link
              href="/demo"
              className="rounded-lg bg-primary-600 px-8 py-3.5 text-base font-semibold text-white shadow-lg shadow-primary-600/20 transition hover:bg-primary-500 hover:shadow-primary-500/30"
            >
              See It In Action
            </Link>
            <Link
              href="/register?plan=starter"
              className="rounded-lg border border-gray-700 px-8 py-3.5 text-base font-medium text-gray-300 transition hover:border-gray-500 hover:text-white"
            >
              Start Free Trial
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
              GHG Protocol Compliant
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
              href="/case-study"
              className="mt-6 inline-flex items-center gap-1 text-sm font-medium text-primary-400 transition hover:text-primary-300"
            >
              Read the full case study
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" /></svg>
            </Link>
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
                solution: 'GHG Protocol-compliant engine applies verified factors from EPA eGRID, DEFRA, and IEA across Scope 1, 2, and 3. Every calculation is traceable to its source factor.',
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
          <div className="grid gap-6 md:grid-cols-3">
            {[
              {
                role: 'CFOs & Finance',
                headline: 'Close your audit without open questions',
                points: ['Audit-ready reports that pass third-party review', 'Cost control: 90% less than Big 4 engagements', 'Full calculation transparency for investor confidence'],
              },
              {
                role: 'Sustainability Managers',
                headline: 'Accurate data without the manual work',
                points: ['EPA eGRID, DEFRA, IEA factors — no manual lookups', 'Scope 1, 2, 3 in one platform', 'Track progress against reduction targets'],
              },
              {
                role: 'Operations & IT',
                headline: 'Cloud emissions tracked automatically',
                points: ['AWS, Azure, GCP carbon data via API connectors', 'SAP and utility bill integration', 'CSV bulk import for historical data'],
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
                <div className="text-xs text-gray-600">per year</div>
              </div>
              <div className="text-center">
                <div className="text-xs text-primary-400">OffGridFlow Audit Prep</div>
                <div className="mt-1 text-2xl font-bold text-white">$6,500</div>
                <div className="text-xs text-gray-500">per year</div>
              </div>
              <div className="text-center">
                <div className="text-xs text-gray-500">Your Savings</div>
                <div className="mt-1 text-2xl font-bold text-primary-400">$68,500</div>
                <div className="text-xs text-primary-400/60">91% cost reduction</div>
              </div>
            </div>
            <div className="mt-8 text-center">
              <Link
                href="/register?plan=starter"
                className="inline-block rounded-lg bg-primary-600 px-8 py-3 text-sm font-semibold text-white transition hover:bg-primary-500"
              >
                Start Saving Today
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* ================================================================== */}
      {/* PRICING                                                             */}
      {/* ================================================================== */}
      <section id="pricing" className="border-t border-gray-800/50 px-6 py-20">
        <div className="mx-auto max-w-5xl">
          <h2 className="mb-4 text-center text-3xl font-bold text-white">Transparent pricing</h2>
          <p className="mb-16 text-center text-gray-400">No hidden fees. No per-seat surprises.</p>

          <div className="grid gap-6 md:grid-cols-4">
            {[
              {
                name: 'Audit Prep',
                price: '$6,500',
                period: '/year',
                monthly: '$540/mo equivalent',
                desc: 'Scope 1 & 2 for your first audit',
                features: ['Scope 1 & 2 tracking', 'CSV & utility bill import', 'Single framework (CSRD or SB 253)', 'PDF reports', 'Email support', 'Up to 5 users'],
                cta: 'Get Started',
                href: '/register?plan=basic',
                highlight: false,
              },
              {
                name: 'Compliance Pro',
                price: '$10,800',
                period: '/year',
                monthly: '$900/mo equivalent',
                desc: 'CSRD + SEC readiness with Scope 3',
                features: ['Scope 1, 2 & basic Scope 3', 'CSRD + SEC frameworks', 'Cloud connectors (AWS, Azure, GCP)', 'PDF + XBRL exports', 'Priority support', 'Up to 15 users'],
                cta: 'Get Started',
                href: '/register?plan=pro',
                highlight: true,
              },
              {
                name: 'Enterprise',
                price: '$15,000',
                period: '/year',
                monthly: '$1,250/mo equivalent',
                desc: 'Full compliance for growing orgs',
                features: ['Full Scope 1, 2 & 3', 'All 5 frameworks', 'Cloud + SAP connectors', 'XBRL/iXBRL exports', 'Account manager', 'Up to 25 users'],
                cta: 'Get Started',
                href: '/register?plan=enterprise',
                highlight: false,
              },
              {
                name: 'Global',
                price: 'Custom',
                period: '',
                monthly: '',
                desc: 'Multi-site, multi-jurisdiction',
                features: ['Everything in Enterprise', 'Custom integrations', 'Multi-region compliance', 'On-site implementation', 'SLA guarantee', 'Unlimited users'],
                cta: 'Contact Sales',
                href: '/demo',
                highlight: false,
              },
            ].map((tier) => (
              <div
                key={tier.name}
                className={`relative rounded-xl border p-7 ${tier.highlight ? 'border-primary-600 bg-dark-800' : 'border-gray-800 bg-dark-800'}`}
              >
                {tier.highlight && (
                  <div className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-primary-600 px-3 py-0.5 text-xs font-medium text-white">
                    Most Popular
                  </div>
                )}
                <h3 className="text-lg font-semibold text-white">{tier.name}</h3>
                <div className="mt-3 flex items-baseline gap-1">
                  <span className="text-3xl font-bold text-white">{tier.price}</span>
                  {tier.period && <span className="text-sm text-gray-500">{tier.period}</span>}
                </div>
                {tier.monthly && <div className="mt-1 text-xs text-gray-600">{tier.monthly}</div>}
                <p className="mt-2 text-sm text-gray-500">{tier.desc}</p>
                <ul className="mt-5 space-y-2.5">
                  {tier.features.map((f) => (
                    <li key={f} className="flex items-start gap-2 text-sm text-gray-300">
                      <span className="mt-0.5 text-primary-400">&#10003;</span>
                      {f}
                    </li>
                  ))}
                </ul>
                <Link
                  href={tier.href}
                  className={`mt-7 block rounded-lg py-2.5 text-center text-sm font-medium transition ${
                    tier.highlight
                      ? 'bg-primary-600 text-white hover:bg-primary-500'
                      : 'border border-gray-700 text-gray-300 hover:border-gray-500 hover:text-white'
                  }`}
                >
                  {tier.cta}
                </Link>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ================================================================== */}
      {/* FINAL CTA                                                           */}
      {/* ================================================================== */}
      <section className="border-t border-gray-800/50 px-6 py-24">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="text-3xl font-bold text-white">
            Ready to cut your compliance costs by 90%?
          </h2>
          <p className="mt-4 text-gray-400">
            Join the companies replacing six-figure consulting engagements with audit-ready software.
          </p>
          <div className="mt-8 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Link
              href="/demo"
              className="rounded-lg bg-primary-600 px-8 py-3.5 text-base font-semibold text-white shadow-lg shadow-primary-600/20 transition hover:bg-primary-500"
            >
              See It In Action
            </Link>
            <Link
              href="/register?plan=starter"
              className="text-sm font-medium text-gray-400 transition hover:text-white"
            >
              Or start your free trial &rarr;
            </Link>
          </div>
        </div>
      </section>

      {/* ================================================================== */}
      {/* FOOTER                                                              */}
      {/* ================================================================== */}
      <footer className="border-t border-gray-800/50 px-6 py-12">
        <div className="mx-auto max-w-5xl">
          <div className="grid gap-8 sm:grid-cols-4">
            <div>
              <div className="text-sm font-semibold text-white">OffGridFlow</div>
              <div className="mt-1 text-xs text-gray-600">Carbon Accounting</div>
            </div>
            <div>
              <div className="mb-3 text-xs font-medium uppercase tracking-wider text-gray-500">Product</div>
              <div className="space-y-2 text-sm text-gray-400">
                <Link href="/demo" className="block hover:text-white">Demo</Link>
                <Link href="/pricing" className="block hover:text-white">Pricing</Link>
                <Link href="/security" className="block hover:text-white">Security</Link>
              </div>
            </div>
            <div>
              <div className="mb-3 text-xs font-medium uppercase tracking-wider text-gray-500">Company</div>
              <div className="space-y-2 text-sm text-gray-400">
                <Link href="/about" className="block hover:text-white">About</Link>
                <Link href="/case-study" className="block hover:text-white">Case Study</Link>
                <a href="mailto:contact@off-grid-flow.com" className="block hover:text-white">Contact</a>
              </div>
            </div>
            <div>
              <div className="mb-3 text-xs font-medium uppercase tracking-wider text-gray-500">Legal</div>
              <div className="space-y-2 text-sm text-gray-400">
                <Link href="/privacy" className="block hover:text-white">Privacy Policy</Link>
                <Link href="/terms" className="block hover:text-white">Terms of Service</Link>
              </div>
            </div>
          </div>
          <div className="mt-10 border-t border-gray-800/50 pt-6 text-center text-xs text-gray-600">
            &copy; {new Date().getFullYear()} OffGridFlow LLC. Built by Paul Timchuk.
          </div>
        </div>
      </footer>
    </div>
  );
}
