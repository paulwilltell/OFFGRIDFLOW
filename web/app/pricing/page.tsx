import Link from 'next/link';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Pricing | OffGridFlow',
  description: 'OffGridFlow pricing — enterprise carbon compliance starting at $4,800/year.',
};

const tiers = [
  {
    name: 'Starter',
    price: '$4,800',
    period: '/year',
    description: 'For companies beginning their compliance journey',
    features: [
      'Scope 1 & 2 emissions tracking',
      'Monthly PDF compliance reports',
      'CSV & utility bill import',
      'Single compliance framework',
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
      'All compliance frameworks (CSRD, SEC, SB 253, CBAM, IFRS S2)',
      'XBRL & iXBRL regulatory exports',
      'Supplier data portal',
      'Real-time emissions dashboard',
      'Priority support & onboarding',
      'Up to 25 users',
    ],
    cta: 'Start Demo',
    highlight: true,
  },
  {
    name: 'Global Enterprise',
    price: 'Custom',
    period: '',
    description: 'For multi-site, multi-jurisdiction enterprises',
    features: [
      'Everything in Enterprise',
      'SAP & ERP integration',
      'Custom data connectors',
      'Multi-site, multi-region support',
      'Dedicated account manager',
      'SLA & audit support',
      'Custom compliance report formats',
      'Unlimited users',
    ],
    cta: 'Contact Sales',
    highlight: false,
  },
];

const faq = [
  {
    q: 'What compliance frameworks are supported?',
    a: 'CSRD/ESRS (EU), SEC Climate Disclosure (US), California SB 253, CBAM (EU Carbon Border Adjustment), and IFRS S2 (Global ISSB). The Starter plan includes one framework; Enterprise and Global include all.',
  },
  {
    q: 'How does data import work?',
    a: 'You can import emissions data via CSV upload, utility bill upload, or automated cloud connectors. Enterprise plans include automated data pipelines from AWS, Azure, and GCP carbon footprint APIs.',
  },
  {
    q: 'Are reports audit-ready?',
    a: 'Yes. Reports are generated using GHG Protocol-compliant calculation methodologies with verified emission factors. Export formats include PDF, XBRL/iXBRL, and Excel. We recommend independent verification before regulatory submission.',
  },
  {
    q: 'Can I try before I buy?',
    a: 'Yes. Request a demo and we will walk you through the platform with your actual use case. No commitment required.',
  },
  {
    q: 'What happens to my data if I cancel?',
    a: 'You can export all data at any time. After cancellation, your data is retained for 30 days for export, then permanently deleted.',
  },
];

export default function PricingPage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <nav className="border-b border-gray-800/50 bg-dark-900/80 backdrop-blur-md">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <Link href="/" className="text-xl font-bold text-white">OffGridFlow</Link>
          <div className="flex items-center gap-3">
            <Link href="/login" className="text-sm text-gray-400 hover:text-white">Sign In</Link>
            <Link
              href="/demo"
              className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-primary-500"
            >
              Start Demo
            </Link>
          </div>
        </div>
      </nav>

      <section className="px-6 py-24">
        <div className="mx-auto max-w-6xl">
          <h1 className="mb-4 text-center text-4xl font-bold text-white">
            Transparent Pricing
          </h1>
          <p className="mb-16 text-center text-gray-400">
            Enterprise carbon compliance at a fraction of Big 4 cost. No hidden fees.
          </p>

          <div className="grid gap-8 md:grid-cols-3">
            {tiers.map((tier) => (
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
                <h3 className="text-lg font-semibold text-white">{tier.name}</h3>
                <div className="mt-4 flex items-baseline gap-1">
                  <span className="text-3xl font-bold text-white">{tier.price}</span>
                  {tier.period && <span className="text-sm text-gray-500">{tier.period}</span>}
                </div>
                <p className="mt-2 text-sm text-gray-400">{tier.description}</p>
                <ul className="mt-6 space-y-3">
                  {tier.features.map((f) => (
                    <li key={f} className="flex items-start gap-2 text-sm text-gray-300">
                      <span className="mt-0.5 text-primary-400">&#10003;</span>
                      {f}
                    </li>
                  ))}
                </ul>
                <Link
                  href={tier.name === 'Global Enterprise' ? '/demo' : `/register?plan=${tier.name.toLowerCase()}`}
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

      {/* FAQ */}
      <section className="border-t border-gray-800/50 px-6 py-24">
        <div className="mx-auto max-w-3xl">
          <h2 className="mb-12 text-center text-3xl font-bold text-white">
            Frequently Asked Questions
          </h2>
          <div className="space-y-6">
            {faq.map((item) => (
              <div key={item.q} className="rounded-xl border border-gray-800 bg-dark-800 p-6">
                <h3 className="mb-2 text-base font-semibold text-white">{item.q}</h3>
                <p className="text-sm text-gray-400">{item.a}</p>
              </div>
            ))}
          </div>
        </div>
      </section>
    </div>
  );
}
