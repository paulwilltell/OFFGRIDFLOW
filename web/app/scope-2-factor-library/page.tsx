import Link from 'next/link';
import type { Metadata } from 'next';
import { MoneyPageLayout } from '@/components/MoneyPageLayout';
import { JsonLd, datasetSchema } from '@/components/JsonLd';
import { buildMoneyPageMetadata, SITE_URL } from '@/lib/seo';

const PATH = '/scope-2-factor-library';
const DOWNLOAD_URL = `${SITE_URL}/downloads/scope-2-factor-library.csv`;

export const metadata: Metadata = buildMoneyPageMetadata({
  path: PATH,
  title: 'Scope 2 Emission Factor Library (Free CSV Download)',
  description:
    'Free downloadable Scope 2 grid emission factor library covering 23 regions from EPA eGRID 2023, IEA 2023, UK DEFRA 2024, and EEA 2023.',
  keyword: 'Scope 2 emission factor library',
});

const faqs = [
  {
    question: 'What factor sources are included?',
    answer:
      'EPA eGRID 2023 for US subregions, IEA 2023 for country-level international grids, UK DEFRA 2024 for the UK, and EEA 2023 for European Economic Area countries. Each row includes source attribution, vintage, and method.',
  },
  {
    question: 'Is this the same dataset OffGridFlow uses?',
    answer:
      'Yes. These are the public factors OffGridFlow seeds into its calculation engine for Scope 2 location-based calculations. Additional factor packs (Scope 1 fuels, Scope 3 categories) are documented in the methodology library.',
  },
  {
    question: 'Can I use this for my own spreadsheet-based carbon reporting?',
    answer:
      'Yes. The CSV is in a standard format with factor_id, region, unit, value, source, and vintage columns. Use it as a starting point. Be aware that spreadsheet-based reporting does not produce an immutable ledger or approval trail — essential for assurance.',
  },
  {
    question: 'How often is the library updated?',
    answer:
      'Annually, as EPA eGRID, IEA, DEFRA, and EEA publish new vintages. The OffGridFlow methodology version (currently v2026.1.0) tracks which factor vintages are active.',
  },
];

export default function Scope2FactorLibraryPage() {
  return (
    <>
      <JsonLd
        id="ld-dataset-scope2-factors"
        data={datasetSchema({
          name: 'Scope 2 Emission Factor Library (OffGridFlow v2026.1.0)',
          description:
            'Location-based Scope 2 grid emission factors for US subregions (EPA eGRID 2023), UK (DEFRA 2024), EU countries (EEA 2023), and major international grids (IEA 2023). CSV format with factor id, region, unit, kg CO2e per unit, method, source, and vintage.',
          url: `${SITE_URL}${PATH}`,
          distributionUrl: DOWNLOAD_URL,
          encodingFormat: 'text/csv',
          keywords: ['Scope 2', 'emission factors', 'EPA eGRID', 'IEA', 'DEFRA', 'grid emissions', 'location-based'],
        })}
      />
      <MoneyPageLayout
        breadcrumbs={[
          { name: 'Home', path: '/' },
          { name: 'Scope 2 Factor Library', path: PATH },
        ]}
        h1="Scope 2 Emission Factor Library"
        dek="Free downloadable CSV of 23 Scope 2 grid emission factors covering the US, UK, EU, and major international regions. Sourced from EPA eGRID 2023, IEA 2023, UK DEFRA 2024, and EEA 2023."
        slug="scope-2-factors"
        ctaUtm="scope2_factors"
        ctaText="See calculation software"
        secondaryCtaText="Download CSV"
        secondaryCtaHref="/downloads/scope-2-factor-library.csv"
        faqs={faqs}
      >
        <div className="not-prose my-8 rounded-2xl border border-primary-600/30 bg-primary-600/5 p-6">
          <h2 className="text-lg font-semibold text-white">Download the factor library</h2>
          <p className="mt-2 text-sm text-gray-400">
            23 location-based Scope 2 factors with source, region, vintage, and method. CSV format.
          </p>
          <div className="mt-4 flex flex-wrap gap-3">
            <a
              href="/downloads/scope-2-factor-library.csv"
              download
              className="rounded-lg bg-primary-600 px-5 py-2.5 text-sm font-semibold text-white hover:bg-primary-500"
            >
              Download CSV
            </a>
            <Link
              href="/methodology"
              className="rounded-lg border border-gray-700 px-5 py-2.5 text-sm text-gray-300 hover:border-gray-500 hover:text-white"
            >
              Read methodology
            </Link>
          </div>
        </div>

        <h2 className="mt-10 text-2xl font-bold text-white">What the library contains</h2>
        <ul>
          <li><strong className="text-white">US subregions:</strong> WECC, NPCC/RFC/SERC, ERCOT, MRO, US average, CAMX, NWPP</li>
          <li><strong className="text-white">Europe:</strong> UK, Germany, France, Central, North (Nordic), West, South, EU average</li>
          <li><strong className="text-white">Asia-Pacific:</strong> APAC average, Japan, Australia, India, China</li>
          <li><strong className="text-white">Americas:</strong> Canada, Brazil</li>
          <li><strong className="text-white">Thermal:</strong> generic district steam (global)</li>
        </ul>

        <h2 className="mt-10 text-2xl font-bold text-white">Source vintages</h2>
        <p>
          All factors cite source (EPA eGRID, IEA, DEFRA, EEA) and vintage (2023 or 2024). Updates
          are published as source organizations release new vintages. Check back annually or
          subscribe to product updates.
        </p>

        <p className="mt-8 text-sm text-gray-500">
          Related:{' '}
          <Link href="/methodology" className="text-primary-400 hover:underline">Methodology library</Link>
          {' · '}
          <Link href="/scope-1-2-3-reporting-software" className="text-primary-400 hover:underline">Scope 1, 2, 3 reporting</Link>
          {' · '}
          <Link href="/carbon-reporting-template" className="text-primary-400 hover:underline">Carbon reporting template</Link>
        </p>
      </MoneyPageLayout>
    </>
  );
}
