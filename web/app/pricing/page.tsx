import Link from 'next/link';
import type { Metadata } from 'next';
import { SiteNav } from '../components/SiteNav';
import { SiteFooter } from '../components/SiteFooter';
import { LeadCaptureForm } from '@/components/LeadCaptureForm';

export const metadata: Metadata = {
  title: 'Pricing | OffGridFlow',
  description: 'Free to upload and see your carbon footprint. Export an audit-ready report for $149 — one-time, no subscription.',
};

const tiers = [
  {
    name: 'Free',
    price: '$0',
    period: '',
    description: 'See your real footprint before you pay anything.',
    features: [
      'Upload utility & energy data (CSV)',
      'Automatic column mapping',
      'Emission factors applied (EPA eGRID, DEFRA)',
      'Full Scope 2 footprint dashboard',
      'Data quality anomaly scan',
      'No credit card required',
    ],
    cta: 'Start free',
    href: '/register',
    highlight: false,
  },
  {
    name: 'Audit-ready report',
    price: '$149',
    period: '/ report',
    description: 'One-time. Re-export free for 12 months.',
    features: [
      'Everything in Free',
      'Audit-ready PDF & CSV export',
      'GHG Protocol + CSRD / ESRS E1',
      'Full methodology & source trail',
      'Formatted to your framework',
      'No subscription, no per-seat fees',
    ],
    cta: 'Upload data to start',
    href: '/register',
    highlight: true,
  },
];

const faq = [
  {
    q: 'Is it really free to see my footprint?',
    a: 'Yes. Upload your utility data and we calculate your Scope 2 footprint, apply verified emission factors, and run a data-quality scan — all free. You only pay when you want to export the audit-ready report.',
  },
  {
    q: 'What do I get for $149?',
    a: 'A complete, audit-ready report of your inventory in PDF and CSV, formatted to your chosen framework (GHG Protocol and CSRD / ESRS E1 today), with the full methodology and a per-line source trail an auditor can follow. It is a one-time purchase and re-exports are free for 12 months.',
  },
  {
    q: 'Which scopes are supported?',
    a: 'Today the engine calculates Scope 2 (purchased electricity and energy) from your utility data. Scope 1 (fuel) and Scope 3 (travel, suppliers) are on the roadmap — the dashboard shows them clearly as not yet populated so nothing is misrepresented.',
  },
  {
    q: 'Are reports audit-ready?',
    a: 'Reports use documented GHG Protocol methodology and verified emission-factor sources, with a full calculation trail. We recommend independent verification before regulatory submission — OffGridFlow does not guarantee regulatory acceptance.',
  },
  {
    q: 'What happens to my data?',
    a: 'You can export your data at any time. If you delete your workspace, your data is removed. We never sell your data.',
  },
];

export default function PricingPage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <SiteNav />

      <section className="px-6 py-24">
        <div className="mx-auto max-w-3xl">
          <h1 className="mb-4 text-center text-4xl font-bold text-white">
            Simple, honest pricing
          </h1>
          <p className="mb-16 text-center text-gray-400">
            Free to upload and see your footprint. Pay only when you export a report.
          </p>

          <div className="grid gap-8 sm:grid-cols-2">
            {tiers.map((tier) => (
              <div
                key={tier.name}
                className={`relative rounded-xl border p-8 ${
                  tier.highlight ? 'border-primary-600 bg-dark-800' : 'border-gray-800 bg-dark-800'
                }`}
              >
                {tier.highlight && (
                  <div className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-primary-600 px-3 py-0.5 text-xs font-medium text-white">
                    Pay per report
                  </div>
                )}
                <h3 className="text-lg font-semibold text-white">{tier.name}</h3>
                <div className="mt-4 flex items-baseline gap-1">
                  <span className="text-4xl font-bold text-white">{tier.price}</span>
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
                  href={tier.href}
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

          <div className="mx-auto mt-12 max-w-3xl space-y-4 text-center text-xs text-gray-500">
            <p>
              <strong className="text-gray-300">Refund policy.</strong> Report purchases are one-time. If a report fails to generate, contact us for a full refund. See the{' '}
              <Link href="/terms" className="text-primary-400 hover:underline">Terms</Link>{' '}for detail.
            </p>
            <p>
              <strong className="text-gray-300">Calculations are drafts.</strong> OffGridFlow applies documented GHG Protocol methodology and verified emission factors. The Platform does not guarantee regulatory acceptance of any report. Customers are responsible for verification before submission.
            </p>
          </div>
        </div>
      </section>

      <section className="border-t border-gray-800/50 px-6 py-24">
        <div className="mx-auto max-w-3xl">
          <h2 className="mb-12 text-center text-3xl font-bold text-white">
            Frequently asked questions
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

      <section className="border-t border-gray-800/50 px-6 py-20">
        <div className="mx-auto max-w-2xl">
          <h2 className="mb-2 text-center text-2xl font-bold text-white">
            Need multi-site or an enterprise agreement?
          </h2>
          <p className="mb-8 text-center text-sm text-gray-400">
            Tell us what you need and we&apos;ll put together a plan.
          </p>
          <LeadCaptureForm source="pricing_page" compact />
        </div>
      </section>

      <SiteFooter />
    </div>
  );
}
