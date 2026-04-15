import type { ReactElement } from 'react';
import { SITE_URL, SITE_NAME, type BreadcrumbItem } from '@/lib/seo';

/**
 * Renders a script tag with JSON-LD structured data.
 * Use a distinct id so multiple schemas can coexist on the same page.
 */
export function JsonLd({ id, data }: { id: string; data: unknown }): ReactElement {
  return (
    <script
      id={id}
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(data) }}
    />
  );
}

// ---------- Reusable schema builders ----------

export function breadcrumbListSchema(items: BreadcrumbItem[]) {
  return {
    '@context': 'https://schema.org',
    '@type': 'BreadcrumbList',
    itemListElement: items.map((item, index) => ({
      '@type': 'ListItem',
      position: index + 1,
      name: item.name,
      item: `${SITE_URL}${item.path}`,
    })),
  };
}

export interface FaqItem {
  question: string;
  answer: string;
}

export function faqPageSchema(faqs: FaqItem[]) {
  return {
    '@context': 'https://schema.org',
    '@type': 'FAQPage',
    mainEntity: faqs.map((faq) => ({
      '@type': 'Question',
      name: faq.question,
      acceptedAnswer: {
        '@type': 'Answer',
        text: faq.answer,
      },
    })),
  };
}

export function softwareApplicationSchema() {
  return {
    '@context': 'https://schema.org',
    '@type': 'SoftwareApplication',
    name: SITE_NAME,
    applicationCategory: 'BusinessApplication',
    applicationSubCategory: 'Carbon Accounting Software',
    operatingSystem: 'Web',
    url: SITE_URL,
    offers: {
      '@type': 'Offer',
      price: '6500',
      priceCurrency: 'USD',
      priceSpecification: {
        '@type': 'UnitPriceSpecification',
        price: '6500',
        priceCurrency: 'USD',
        referenceQuantity: { '@type': 'QuantitativeValue', value: 1, unitCode: 'ANN' },
      },
    },
    provider: {
      '@type': 'Organization',
      name: SITE_NAME,
      url: SITE_URL,
    },
  };
}

export function organizationSchema() {
  return {
    '@context': 'https://schema.org',
    '@type': 'Organization',
    name: 'OffGridFlow LLC',
    url: SITE_URL,
    logo: `${SITE_URL}/offgridflow_logo_primary.png`,
    contactPoint: {
      '@type': 'ContactPoint',
      email: 'contact@off-grid-flow.com',
      contactType: 'customer support',
      areaServed: 'Worldwide',
    },
    sameAs: [],
  };
}

export function datasetSchema(params: {
  name: string;
  description: string;
  url: string;
  distributionUrl: string;
  encodingFormat: string;
  keywords?: string[];
}) {
  return {
    '@context': 'https://schema.org',
    '@type': 'Dataset',
    name: params.name,
    description: params.description,
    url: params.url,
    keywords: params.keywords,
    license: `${SITE_URL}/terms`,
    creator: {
      '@type': 'Organization',
      name: 'OffGridFlow LLC',
      url: SITE_URL,
    },
    distribution: [
      {
        '@type': 'DataDownload',
        encodingFormat: params.encodingFormat,
        contentUrl: params.distributionUrl,
      },
    ],
  };
}
