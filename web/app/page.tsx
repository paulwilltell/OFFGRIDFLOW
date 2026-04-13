import Link from 'next/link';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'OffGridFlow | Carbon Accounting & Compliance Platform',
  description:
    'Automate Scope 1, 2, 3 emissions tracking and generate audit-ready compliance reports for CSRD, SEC, California SB 253, CBAM, and IFRS S2.',
  openGraph: {
    title: 'OffGridFlow | Carbon Accounting & Compliance Platform',
    description:
      'Enterprise carbon compliance made accessible. Track emissions, generate reports, stay compliant.',
    type: 'website',
    url: 'https://off-grid-flow.com',
  },
};

const frameworks = [
  { name: 'CSRD / ESRS', region: 'EU' },
  { name: 'SEC Climate', region: 'US' },
  { name: 'SB 253', region: 'CA' },
  { name: 'CBAM', region: 'EU' },
  { name: 'IFRS S2', region: 'Global' },
];

const steps = [
  {
    step: '01',
    title: 'Connect Your Data',
    description:
      'Import emissions data from AWS, Azure, GCP, SAP, utility bills, or CSV. Cloud connectors pull carbon footprint data automatically.',
  },
  {
    step: '02',
    title: 'Calculate Emissions',
    description:
      'GHG Protocol-compliant calculation engine processes Scope 1, 2, and 3 emissions using 10,000+ verified emission factors.',
  },
  {
    step: '03',
    title: 'Generate Reports',
    description:
      'Export audit-ready compliance reports in PDF, XBRL, and Excel. Mapped to CSRD, SEC, SB 253, CBAM, and IFRS S2 frameworks.',
  },
];

const capabilities = [
  {
    title: 'Multi-Scope Tracking',
    description:
      'Complete Scope 1 (direct), Scope 2 (energy), and Scope 3 (value chain) emissions calculation with activity-based and spend-based methods.',
  },
  {
    title: 'Cloud Data Ingestion',
    description:
      'Automated pipelines pull carbon data from AWS Cost & Usage Reports, Azure Carbon Footprint API, and GCP Carbon Footprint API.',
  },
  {
    title: 'Regulatory Mapping',
    description:
      'Each compliance framework is embedded in the data model, validation rules, and reporting flows. No manual mapping required.',
  },
  {
    title: 'Enterprise Security',
    description:
      'Multi-tenant architecture with role-based access control, 2FA, API key management, and full audit logging. SOC 2 Type I targeted Q3 2026.',
  },
];

const pricingTiers = [
  {
    name: 'Starter',
    price: '$4,800',
    period: '/year',
    description: 'For companies beginning their compliance journey',
    features: [
      'Scope 1 & 2 emissions tracking',
      'Monthly PDF compliance reports',
      'CSV & utility bill import',
      'Email support',
      'Up to 5 users',
    ],
    cta: 'Start Demo',
    highlight: false,
  },
  {
    name: 'Enterprise',
    price: '$15,000',
    period: '/year',
    description: 'Full-scope compliance for growing organizations',
    features: [
      'Scope 1, 2, & 3 emissions tracking',
      'Cloud connectors (AWS, Azure, GCP)',
      'XBRL & iXBRL regulatory exports',
      'Supplier data portal',
      'Priority support & onboarding',
      'Up to 25 users',
    ],
    cta: 'Start Demo',
    highlight: true,
  },
  {
    name: 'Global',
    price: 'Custom',
    period: '',
    description: 'For multi-site, multi-jurisdiction enterprises',
    features: [
      'Everything in Enterprise',
      'SAP & ERP integration',
      'Custom data connectors',
      'Dedicated account manager',
      'SLA & audit support',
      'Unlimited users',
    ],
    cta: 'Contact Sales',
    highlight: false,
  },
];

export default function HomePage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      {/* Navigation */}
      <nav className="sticky top-0 z-50 border-b border-gray-800/50 bg-dark-900/80 backdrop-blur-md">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <Link href="/" className="flex items-center gap-2">
            <span className="text-xl font-bold text-white">
              OffGridFlow
            </span>
            <span className="hidden text-xs text-gray-500 sm:inline">
              Carbon Accounting
            </span>
          </Link>
          <div className="hidden items-center gap-8 text-sm text-gray-400 md:flex">
            <a href="#how-it-works" className="transition hover:text-white">
              How It Works
            </a>
            <a href="#capabilities" className="transition hover:text-white">
              Capabilities
            </a>
            <a href="#pricing" className="transition hover:text-white">
              Pricing
            </a>
            <Link href="/about" className="transition hover:text-white">
              About
            </Link>
          </div>
          <div className="flex items-center gap-3">
            <Link
              href="/login"
              className="text-sm text-gray-400 transition hover:text-white"
            >
              Sign In
            </Link>
            <Link
              href="/demo"
              className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-primary-500"
            >
              Start Demo
            </Link>
          </div>
        </div>
      </nav>

      {/* Hero */}
      <section className="relative overflow-hidden px-6 pb-20 pt-24 sm:pt-32">
        <div className="mx-auto max-w-4xl text-center">
          {/* Framework badges */}
          <div className="mb-8 flex flex-wrap items-center justify-center gap-3">
            {frameworks.map((fw) => (
              <span
                key={fw.name}
                className="rounded-full border border-gray-700 bg-dark-800 px-3 py-1 text-xs text-gray-400"
              >
                {fw.name}{' '}
                <span className="text-gray-600">({fw.region})</span>
              </span>
            ))}
          </div>

          <h1 className="text-4xl font-bold leading-tight tracking-tight text-white sm:text-5xl lg:text-6xl">
            Carbon compliance
            <br />
            <span className="text-primary-400">without the Big 4 price tag</span>
          </h1>

          <p className="mx-auto mt-6 max-w-2xl text-lg text-gray-400">
            Track Scope 1, 2, and 3 emissions. Generate audit-ready reports for
            CSRD, SEC, SB 253, CBAM, and IFRS S2. Built for companies that need
            compliance&mdash;not a six-figure consulting engagement.
          </p>

          <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Link
              href="/demo"
              className="rounded-lg bg-primary-600 px-8 py-3 text-base font-medium text-white transition hover:bg-primary-500"
            >
              Start Demo
            </Link>
            <a
              href="#how-it-works"
              className="rounded-lg border border-gray-700 px-8 py-3 text-base font-medium text-gray-300 transition hover:border-gray-500 hover:text-white"
            >
              See How It Works
            </a>
          </div>
        </div>
      </section>

      {/* How It Works */}
      <section id="how-it-works" className="border-t border-gray-800/50 px-6 py-24">
        <div className="mx-auto max-w-6xl">
          <h2 className="mb-4 text-center text-3xl font-bold text-white">
            Three steps to compliance
          </h2>
          <p className="mb-16 text-center text-gray-400">
            From raw data to audit-ready reports
          </p>

          <div className="grid gap-8 md:grid-cols-3">
            {steps.map((s) => (
              <div
                key={s.step}
                className="rounded-xl border border-gray-800 bg-dark-800 p-8"
              >
                <div className="mb-4 flex h-10 w-10 items-center justify-center rounded-lg bg-primary-600/10 text-sm font-bold text-primary-400">
                  {s.step}
                </div>
                <h3 className="mb-3 text-lg font-semibold text-white">
                  {s.title}
                </h3>
                <p className="text-sm leading-relaxed text-gray-400">
                  {s.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Capabilities */}
      <section
        id="capabilities"
        className="border-t border-gray-800/50 px-6 py-24"
      >
        <div className="mx-auto max-w-6xl">
          <h2 className="mb-4 text-center text-3xl font-bold text-white">
            Built for real compliance
          </h2>
          <p className="mb-16 text-center text-gray-400">
            Not a dashboard. A complete carbon accounting system.
          </p>

          <div className="grid gap-8 md:grid-cols-2">
            {capabilities.map((cap) => (
              <div
                key={cap.title}
                className="rounded-xl border border-gray-800 bg-dark-800 p-8"
              >
                <h3 className="mb-3 text-lg font-semibold text-white">
                  {cap.title}
                </h3>
                <p className="text-sm leading-relaxed text-gray-400">
                  {cap.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Pricing */}
      <section id="pricing" className="border-t border-gray-800/50 px-6 py-24">
        <div className="mx-auto max-w-6xl">
          <h2 className="mb-4 text-center text-3xl font-bold text-white">
            Transparent pricing
          </h2>
          <p className="mb-16 text-center text-gray-400">
            Enterprise compliance at a fraction of Big 4 cost
          </p>

          <div className="grid gap-8 md:grid-cols-3">
            {pricingTiers.map((tier) => (
              <div
                key={tier.name}
                className={`relative rounded-xl border p-8 ${
                  tier.highlight
                    ? 'border-primary-600 bg-dark-800'
                    : 'border-gray-800 bg-dark-800'
                }`}
              >
                {tier.highlight && (
                  <div className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-primary-600 px-3 py-0.5 text-xs font-medium text-white">
                    Most Popular
                  </div>
                )}
                <h3 className="text-lg font-semibold text-white">
                  {tier.name}
                </h3>
                <div className="mt-4 flex items-baseline gap-1">
                  <span className="text-3xl font-bold text-white">
                    {tier.price}
                  </span>
                  {tier.period && (
                    <span className="text-sm text-gray-500">{tier.period}</span>
                  )}
                </div>
                <p className="mt-2 text-sm text-gray-400">
                  {tier.description}
                </p>
                <ul className="mt-6 space-y-3">
                  {tier.features.map((f) => (
                    <li
                      key={f}
                      className="flex items-start gap-2 text-sm text-gray-300"
                    >
                      <span className="mt-0.5 text-primary-400">&#10003;</span>
                      {f}
                    </li>
                  ))}
                </ul>
                <Link
                  href={tier.name === 'Global' ? '/demo' : '/demo'}
                  className={`mt-8 block rounded-lg py-2.5 text-center text-sm font-medium transition ${
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

      {/* Footer */}
      <footer className="border-t border-gray-800/50 px-6 py-12">
        <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-6 sm:flex-row">
          <div className="flex items-center gap-6 text-sm text-gray-500">
            <span>&copy; {new Date().getFullYear()} OffGridFlow LLC</span>
            <Link href="/privacy" className="transition hover:text-gray-300">
              Privacy
            </Link>
            <Link href="/terms" className="transition hover:text-gray-300">
              Terms
            </Link>
            <Link href="/security" className="transition hover:text-gray-300">
              Security
            </Link>
            <Link href="/about" className="transition hover:text-gray-300">
              About
            </Link>
          </div>
          <div className="text-sm text-gray-600">
            Founded by Paul Timchuk
          </div>
        </div>
      </footer>
    </div>
  );
}
