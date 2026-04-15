import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/persefoni-alternative';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'Persefoni Alternative — OffGridFlow',
  description:
    'OffGridFlow is a transparent, self-serve alternative to Persefoni for Scope 1, 2, 3 emissions calculation and compliance reporting. Public methodology, starting at $6,500/year.',
  keyword: 'Persefoni alternative',
});

const faqs = [
  {
    question: 'Why consider OffGridFlow instead of Persefoni?',
    answer:
      'OffGridFlow publishes its methodology, factor sources, and sample audit packet before you sign a contract. Pricing starts at $6,500/year and is listed publicly. Self-serve CSV upload reaches a draft report in under two hours — no consultant engagement required.',
  },
  {
    question: 'Is OffGridFlow as comprehensive as Persefoni?',
    answer:
      'OffGridFlow covers the full Scope 1/2/3 calculation surface, all five major frameworks (CSRD, SEC, SB 253, CBAM, IFRS S2), and cloud connectors (AWS, Azure, GCP, SAP). Persefoni has a longer enterprise feature list; OffGridFlow has a narrower, more defensible core with the essentials in place.',
  },
  {
    question: 'How does support compare?',
    answer:
      'Persefoni provides white-glove enterprise support at enterprise pricing. OffGridFlow offers self-serve plus assisted onboarding at a fraction of the cost. For customers who want a hands-off consultant-led rollout, Persefoni may be the better fit.',
  },
  {
    question: 'What does OffGridFlow offer that Persefoni does not?',
    answer:
      'Public methodology library, downloadable sample evidence pack, published pricing, data architecture documentation, and a commitment to never use customer data for ML training or advertising.',
  },
];

export default function PersefoniAlternativePage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'Persefoni Alternative', path: PATH },
      ]}
      h1="Persefoni Alternative for Carbon Accounting"
      dek="Transparent pricing, public methodology, self-serve onboarding, and a published audit-evidence pack. OffGridFlow is the alternative for teams that want to verify before they buy."
      slug="persefoni-alternative"
      ctaUtm="persefoni_alternative"
      faqs={faqs}
    >
      <p>
        Persefoni is a well-known carbon accounting platform often marketed to enterprise finance
        teams. Many mid-market and lean enterprise teams evaluate it and then ask whether a more
        transparent, self-serve alternative exists. OffGridFlow is that alternative.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">What OffGridFlow does differently</h2>
      <ul>
        <li>
          <strong className="text-white">Published methodology:</strong> every factor source,
          calculation method, and standards alignment is documented at /methodology — version
          v2026.1.0 right now.
        </li>
        <li>
          <strong className="text-white">Public pricing:</strong> $6,500, $10,800, $15,000, and
          custom tiers listed on the pricing page. No per-employee add-ons.
        </li>
        <li>
          <strong className="text-white">Sample evidence pack:</strong> downloadable redacted
          audit packet at /evidence so prospects see the output quality before they sign.
        </li>
        <li>
          <strong className="text-white">Self-serve path:</strong> CSV upload to first report in
          under two hours, no mandatory consultant engagement.
        </li>
        <li>
          <strong className="text-white">Honest operations:</strong> live status page, published
          architecture, no fabricated testimonials or unverified certifications.
        </li>
      </ul>

      <h2 className="mt-10 text-2xl font-bold text-white">When to choose Persefoni instead</h2>
      <p>
        If you need a large enterprise rollout with dedicated implementation consulting, an
        extensive supplier-engagement module, and integration with a CSM team — Persefoni is the
        established choice. OffGridFlow is optimized for mid-market and lean enterprise teams
        that want defensibility without the consulting premium.
      </p>

      <p className="mt-8 text-sm text-gray-500">
        <Link href="/pricing" className="text-primary-400 hover:underline">See pricing</Link>
        {' · '}
        <Link href="/evidence" className="text-primary-400 hover:underline">Download sample evidence pack</Link>
        {' · '}
        <Link href="/methodology" className="text-primary-400 hover:underline">Read the methodology</Link>
      </p>
    </MoneyPageLayout>
  );
}
