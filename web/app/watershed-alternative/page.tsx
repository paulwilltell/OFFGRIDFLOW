import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/watershed-alternative';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'Watershed Alternative — OffGridFlow',
  description:
    'OffGridFlow is a transparent, published-methodology alternative to Watershed. Scope 1, 2, 3 calculations with audit-ready evidence from $6,500/year.',
  keyword: 'Watershed alternative',
});

const faqs = [
  {
    question: 'How does OffGridFlow compare to Watershed?',
    answer:
      'OffGridFlow is a focused calculation and compliance reporting tool with a public methodology and transparent pricing. Watershed bundles calculation with climate program management, decarbonization strategy, and a larger consulting surface at enterprise pricing.',
  },
  {
    question: 'Which is better for audit defense?',
    answer:
      'Both tools produce calculation output. OffGridFlow\'s public audit pack (/evidence) shows an end-to-end sample: source activities → factor lookup → immutable ledger → approval trail → export reconciliation. Customers can verify defensibility before they buy.',
  },
  {
    question: 'Does OffGridFlow support decarbonization strategy?',
    answer:
      'OffGridFlow\'s focus is measurement and disclosure. We do not provide decarbonization strategy consulting or climate program management. Teams that need strategic consulting pair OffGridFlow with an external climate advisor at a fraction of the consolidated cost.',
  },
  {
    question: 'How fast is onboarding?',
    answer:
      'Self-serve customers reach a draft compliance report in under two hours. Assisted (Compliance Pro, $10,800/yr) onboards in 1-2 weeks with cloud connectors. Enterprise ($15,000/yr) includes a dedicated account manager for 2-4 week SAP integration rollouts.',
  },
];

export default function WatershedAlternativePage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'Watershed Alternative', path: PATH },
      ]}
      h1="Watershed Alternative for Carbon Accounting"
      dek="Focused, published-methodology, transparent pricing. OffGridFlow is the alternative for teams that need audit-ready Scope 1/2/3 without bundled consulting."
      slug="watershed-alternative"
      ctaUtm="watershed_alternative"
      faqs={faqs}
    >
      <p>
        Watershed is widely known for bundling carbon accounting with climate program
        consulting — a strong fit for enterprises that want a single strategic partner.
        OffGridFlow takes the opposite approach: a narrow, transparent, audit-defensible tool
        that pairs with whatever external consulting a team already uses.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">The core differences</h2>
      <ul>
        <li>
          <strong className="text-white">Methodology:</strong> OffGridFlow publishes its factor
          sources, calculation methods, and standards alignment. You can diff versions and
          reproduce any prior calculation.
        </li>
        <li>
          <strong className="text-white">Pricing:</strong> published tiers starting at $6,500.
          No per-seat or per-emission-source variable pricing.
        </li>
        <li>
          <strong className="text-white">Evidence:</strong> downloadable redacted audit packet
          on the public site.
        </li>
        <li>
          <strong className="text-white">Scope:</strong> calculation and compliance reporting
          only. Strategy and decarbonization consulting not included.
        </li>
      </ul>

      <h2 className="mt-10 text-2xl font-bold text-white">When to choose Watershed</h2>
      <p>
        If you want a single partner to handle measurement, strategy, target setting, and
        implementation planning as one consulting engagement, Watershed is purpose-built for
        that. If you want a defensible measurement layer to pair with your own strategy work,
        OffGridFlow is a tighter fit at a lower price.
      </p>

      <p className="mt-8 text-sm text-gray-500">
        <Link href="/pricing" className="text-primary-400 hover:underline">Pricing</Link>
        {' · '}
        <Link href="/evidence" className="text-primary-400 hover:underline">Sample evidence pack</Link>
        {' · '}
        <Link href="/methodology" className="text-primary-400 hover:underline">Methodology</Link>
      </p>
    </MoneyPageLayout>
  );
}
