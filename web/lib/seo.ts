/**
 * @fileoverview SEO helpers for money pages and structured data.
 * Produces consistent Next.js Metadata objects and canonical URLs.
 */

import type { Metadata } from 'next';

export const SITE_URL = 'https://off-grid-flow.com';
export const SITE_NAME = 'OffGridFlow';

export interface MoneyPageSeoConfig {
  /** Path starting with /, e.g., "/sb-253-reporting-software" */
  path: string;
  /** Full page title without the " | OffGridFlow" suffix. */
  title: string;
  /** Meta description, 140-160 chars. */
  description: string;
  /** OpenGraph image path (absolute). */
  ogImage?: string;
  /** Primary keyword for analytics tagging. */
  keyword?: string;
}

/**
 * Builds a Next.js Metadata object with canonical URL, Twitter, and OG tags.
 */
export function buildMoneyPageMetadata(config: MoneyPageSeoConfig): Metadata {
  const url = `${SITE_URL}${config.path}`;
  const fullTitle = `${config.title} | ${SITE_NAME}`;
  const ogImage = config.ogImage || `${SITE_URL}/offgridflow_logo_primary.png`;

  return {
    title: fullTitle,
    description: config.description,
    alternates: {
      canonical: url,
    },
    openGraph: {
      title: fullTitle,
      description: config.description,
      url,
      siteName: SITE_NAME,
      type: 'website',
      images: [{ url: ogImage, width: 1200, height: 630, alt: config.title }],
    },
    twitter: {
      card: 'summary_large_image',
      title: fullTitle,
      description: config.description,
      images: [ogImage],
    },
  };
}

/**
 * Builds breadcrumb items for the BreadcrumbList schema and visual breadcrumbs.
 */
export interface BreadcrumbItem {
  name: string;
  path: string;
}

export function makeBreadcrumbs(path: string, pageName: string): BreadcrumbItem[] {
  return [
    { name: 'Home', path: '/' },
    { name: pageName, path },
  ];
}
