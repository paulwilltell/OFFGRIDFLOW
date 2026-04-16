import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/carbon-accounting-software';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'Carbon Accounting Software',
  description:
    'OffGridFlow is carbon accounting software for mid-market and enterprise teams. Track Scope 1, 2, and 3 emissions and generate audit-ready draft reports — starting at $6,500/year.',
  keyword: 'carbon accounting software',
});

const faqs = [
  {
    question: 'What is carbon accounting software?',
    answer:
      'Carbon accounting software ingests activity data (energy, fuel, travel, spend, supplier inputs), applies published emission factors, and produces Scope 1, 2, and 3 inventories aligned with the GHG Protocol. OffGridFlow locks factors to the reporting period so calculations can be reproduced during an audit.',
  },
  {
    question: 'How is OffGridFlow different from legacy carbon accounting tools?',
    answer:
      'OffGridFlow publishes its methodology, factor sources, and an end-to-end evidence pack publicly. Pricing starts at $6,500/year instead of the $50,000-$200,000 typical of Big 4 engagements. Self-serve customers can reach a first audit-ready draft in under two hours.',
  },
  {
    question: 'Which emission factors does OffGridFlow use?',
    answer:
      'EPA eGRID 2023 for U.S. electricity grids, IEA 2023 for international grids, UK DEFRA 2024 for global Scope 1/2/3 activities, IPCC AR6 GWP-100 for refrigerants, and GHG Protocol Scope 3 guidance for the 15 value chain categories. The active methodology version is v2026.1.0.',
  },
  {
    question: 'Is the output audit-ready?',
    answer:
      'Reports are drafts designed to survive audit scrutiny: every calculation is recorded in an immutable ledger with the activity, factor, formula, user, and timestamp. OffGridFlow does not guarantee regulatory acceptance — customers are responsible for verification before submission.',
  },
  {
    question: 'Can OffGridFlow connect to our cloud or ERP?',
    answer:
      'Yes. Built-in connectors cover AWS Cost and Usage Reports, Azure Carbon Optimization, GCP Carbon Footprint, SAP ERP, and direct utility provider APIs. CSV upload is always available as a fallback.',
  },
];

export default function CarbonAccountingSoftwarePage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'Carbon Accounting Software', path: PATH },
      ]}
      h1="Carbon Accounting Software That Survives Audit"
      dek="OffGridFlow ingests your activity data, calculates Scope 1, 2, and 3 emissions using documented GHG Protocol methodology, and produces draft reports your finance, sustainability, and audit teams can actually defend."
      slug="carbon-accounting-software"
      ctaUtm="carbon_accounting_software"
      faqs={faqs}
      showLeadForm
      relatedPages={[
        { href: '/scope-1-2-3-reporting-software', label: 'Scope 1, 2, 3 reporting', description: 'See the full multi-scope reporting workflow.' },
        { href: '/audit-ready-carbon-accounting', label: 'Audit-ready carbon accounting', description: 'Focus on assurance-grade controls and evidence.' },
        { href: '/methodology', label: 'Methodology library', description: 'Review factor sources and calculation methods.' },
      ]}
    >
      <p>
        Carbon accounting is the discipline of measuring, recording, and reporting greenhouse gas
        emissions across an organization&apos;s operations and value chain. Done properly it
        requires versioned methodology, traceable factor provenance, an immutable calculation
        ledger, and an approval workflow — the same primitives finance teams expect for any
        audited number.
      </p>

      <p>
        OffGridFlow treats emissions data the same way a general ledger treats financial
        transactions. Every activity you import is linked to the specific emission factor that
        produced the number, the method applied, the user who ran the calculation, and the
        timestamp. Factor snapshots lock the factor set to the reporting period so next
        year&apos;s recalculation cannot silently change last year&apos;s totals.
      </p>

      <p>
        <strong className="text-white">What you get on day one:</strong> CSV or cloud-connector
        data import, a calculation engine covering all three scopes, draft compliance reports for
        CSRD/ESRS E1, SEC Climate Disclosure, California SB 253, CBAM, and IFRS S2, and the same
        immutable audit trail large enterprises pay six figures for.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">Who it&apos;s for</h2>
      <p>
        Mid-market and enterprise organizations with regulatory obligations (public companies,
        companies doing business in California above SB 253 thresholds, EU entities caught by
        CSRD) who need to replace expensive consultant engagements with verifiable in-house
        software.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">How the engine works</h2>
      <ul>
        <li>
          <strong className="text-white">Scope 1 (direct):</strong> stationary and mobile
          combustion, fugitive refrigerants, process emissions. Activity-based calculations
          against EPA and DEFRA factors.
        </li>
        <li>
          <strong className="text-white">Scope 2 (energy):</strong> purchased electricity, steam,
          heating, and cooling. Location-based (regional grid) and market-based (RECs, PPAs,
          supplier-specific) methods, both reportable.
        </li>
        <li>
          <strong className="text-white">Scope 3 (value chain):</strong> all 15 GHG Protocol
          categories with activity-based, spend-based, and supplier-specific tiers. Category-level
          breakdown for SEC and CSRD materiality disclosures.
        </li>
      </ul>

      <h2 className="mt-10 text-2xl font-bold text-white">Why teams switch</h2>
      <ul>
        <li>Published methodology you can diff across versions</li>
        <li>Downloadable redacted evidence pack that mirrors a real audit packet</li>
        <li>Factor snapshots that freeze the calculation basis for each reporting period</li>
        <li>Export reconciliation with SHA-256 checksums proving the PDF matches the dashboard</li>
        <li>Self-serve pricing that does not punish healthy product usage</li>
      </ul>

    </MoneyPageLayout>
  );
}
