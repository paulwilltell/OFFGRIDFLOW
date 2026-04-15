import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { buildMoneyPageMetadata } from '@/lib/seo';

const PATH = '/scope-1-2-3-reporting-software';

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'Scope 1, 2, 3 Reporting Software',
  description:
    'Calculate and report Scope 1, 2, and 3 emissions with OffGridFlow. Location-based and market-based Scope 2. All 15 Scope 3 categories. GHG Protocol aligned.',
  keyword: 'scope 1 2 3 reporting software',
});

const faqs = [
  {
    question: 'What is the difference between Scope 1, 2, and 3?',
    answer:
      'Scope 1 is direct emissions from sources you own or control (fuel combustion, company vehicles, refrigerants). Scope 2 is indirect emissions from purchased energy (electricity, steam, heating, cooling). Scope 3 is all other value chain emissions — purchased goods, business travel, employee commuting, use of sold products, and more — covering 15 GHG Protocol categories.',
  },
  {
    question: 'Do you support both location-based and market-based Scope 2?',
    answer:
      'Yes. OffGridFlow requires both methods for Scope 2 disclosure per GHG Protocol Scope 2 Guidance. Location-based uses grid average factors (EPA eGRID, IEA). Market-based uses supplier-specific instruments (RECs, PPAs, contracts) that you supply.',
  },
  {
    question: 'How do you calculate Scope 3 without supplier data?',
    answer:
      'We support activity-based, spend-based (EEIO coefficients), and supplier-specific tiers per GHG Protocol Scope 3 Standard. Start with spend-based for rapid coverage, then upgrade categories to activity-based or supplier-specific as your data matures.',
  },
  {
    question: 'Can I split Scope 3 by category for materiality disclosure?',
    answer:
      'Yes. All 15 categories are calculated and reported separately. SEC Climate rules and CSRD both require category-level transparency for material Scope 3 emissions.',
  },
  {
    question: 'How often are emission factors updated?',
    answer:
      'Factor packs are refreshed annually as source publishers release new vintages. The active methodology version (currently v2026.1.0) tracks which factor vintages are in effect. When you lock a reporting period to a factor snapshot, your calculations remain reproducible even after new factors are published.',
  },
];

export default function ScopeReportingPage() {
  return (
    <MoneyPageLayout
      breadcrumbs={[
        { name: 'Home', path: '/' },
        { name: 'Scope 1, 2, 3 Reporting', path: PATH },
      ]}
      h1="Scope 1, 2, and 3 Reporting Software"
      dek="All three scopes, documented methodology, location-based and market-based Scope 2, and every Scope 3 category. One platform, one audit trail, draft reports ready to defend."
      slug="scope-1-2-3-reporting-software"
      ctaUtm="scope_1_2_3_reporting"
      faqs={faqs}
    >
      <p>
        OffGridFlow implements the GHG Protocol Corporate Standard, the Scope 2 Guidance, and the
        Corporate Value Chain (Scope 3) Standard. Every calculation records the scope, method,
        factor, and formula in the immutable calculation ledger so a reviewer can rebuild the
        number without opening a spreadsheet.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">Scope 1 — Direct emissions</h2>
      <p>
        Stationary combustion (boilers, generators), mobile combustion (fleet), fugitive emissions
        (refrigerants with IPCC AR6 GWP-100 values), and process emissions. Activity-based
        calculations against EPA GHG Emission Factors Hub and DEFRA conversion factors.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">Scope 2 — Energy</h2>
      <p>
        Purchased electricity, steam, heating, and cooling. Both location-based (EPA eGRID
        subregions, IEA country factors) and market-based (your RECs, PPAs, green tariffs, and
        supplier-specific instruments) are calculated and reported — mandatory under GHG Protocol
        Scope 2 Guidance.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">Scope 3 — Value chain</h2>
      <p>
        All 15 categories: purchased goods &amp; services, capital goods, fuel and energy
        activities, upstream transportation, waste, business travel, employee commuting, upstream
        leased assets, downstream transportation, processing, use-phase, end-of-life, downstream
        leased assets, franchises, and investments. Calculation tiers progress from spend-based
        (EEIO) to activity-based to supplier-specific as your data matures.
      </p>

      <h2 className="mt-10 text-2xl font-bold text-white">What the output looks like</h2>
      <ul>
        <li>Dashboard with total tCO2e, scope breakdown, trend (only when real history exists)</li>
        <li>Calculation ledger with activity → factor → formula → result → user → timestamp</li>
        <li>Factor snapshot locked to the reporting period</li>
        <li>Framework exports: CSRD, SEC, SB 253, CBAM, IFRS S2 in PDF and XBRL</li>
        <li>Export reconciliation checksum proving exported numbers match the ledger</li>
      </ul>

      <p className="mt-8 text-sm text-gray-500">
        Related reading:{' '}
        <Link href="/carbon-accounting-software" className="text-primary-400 hover:underline">Carbon accounting software</Link>
        {' · '}
        <Link href="/scope-3-supplier-emissions-software" className="text-primary-400 hover:underline">Scope 3 supplier emissions</Link>
        {' · '}
        <Link href="/methodology" className="text-primary-400 hover:underline">Methodology library</Link>
      </p>
    </MoneyPageLayout>
  );
}
