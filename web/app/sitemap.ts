import type { MetadataRoute } from 'next';
import { SITE_URL } from '@/lib/seo';

/**
 * Dynamic sitemap for off-grid-flow.com. Covers all public marketing,
 * framework, role, comparison, dataset, trust, and legal pages.
 */

const STATIC_PATHS: Array<{ path: string; priority: number; changeFrequency: MetadataRoute.Sitemap[number]['changeFrequency'] }> = [
  { path: '/', priority: 1.0, changeFrequency: 'weekly' },
  { path: '/pricing', priority: 0.9, changeFrequency: 'monthly' },
  { path: '/demo', priority: 0.9, changeFrequency: 'monthly' },
  { path: '/case-study', priority: 0.8, changeFrequency: 'monthly' },
  { path: '/about', priority: 0.6, changeFrequency: 'monthly' },

  // Trust / evidence / operations
  { path: '/trust', priority: 0.8, changeFrequency: 'monthly' },
  { path: '/methodology', priority: 0.9, changeFrequency: 'monthly' },
  { path: '/architecture', priority: 0.8, changeFrequency: 'monthly' },
  { path: '/evidence', priority: 0.9, changeFrequency: 'monthly' },
  { path: '/operations', priority: 0.8, changeFrequency: 'monthly' },
  { path: '/how-we-operate', priority: 0.7, changeFrequency: 'monthly' },
  { path: '/security', priority: 0.7, changeFrequency: 'monthly' },
  { path: '/status', priority: 0.6, changeFrequency: 'daily' },

  // Money pages — framework and role focused
  { path: '/carbon-accounting-software', priority: 0.9, changeFrequency: 'monthly' },
  { path: '/scope-1-2-3-reporting-software', priority: 0.9, changeFrequency: 'monthly' },
  { path: '/sb-253-reporting-software', priority: 0.9, changeFrequency: 'monthly' },
  { path: '/csrd-reporting-software', priority: 0.9, changeFrequency: 'monthly' },
  { path: '/ifrs-s2-reporting-software', priority: 0.9, changeFrequency: 'monthly' },
  { path: '/cbam-reporting-software', priority: 0.9, changeFrequency: 'monthly' },
  { path: '/scope-3-supplier-emissions-software', priority: 0.9, changeFrequency: 'monthly' },
  { path: '/audit-ready-carbon-accounting', priority: 0.9, changeFrequency: 'monthly' },
  { path: '/carbon-accounting-software-for-finance-teams', priority: 0.9, changeFrequency: 'monthly' },

  // Comparison and alternative pages
  { path: '/csrd-vs-ifrs-s2-carbon-reporting', priority: 0.8, changeFrequency: 'monthly' },
  { path: '/persefoni-alternative', priority: 0.8, changeFrequency: 'monthly' },
  { path: '/watershed-alternative', priority: 0.8, changeFrequency: 'monthly' },

  // Role pages
  { path: '/for-cfos', priority: 0.8, changeFrequency: 'monthly' },
  { path: '/for-sustainability-managers', priority: 0.8, changeFrequency: 'monthly' },
  { path: '/for-procurement', priority: 0.8, changeFrequency: 'monthly' },

  // Integration pages
  { path: '/aws-carbon-data', priority: 0.7, changeFrequency: 'monthly' },
  { path: '/sap-carbon-reporting', priority: 0.7, changeFrequency: 'monthly' },
  { path: '/csv-emissions-import', priority: 0.7, changeFrequency: 'monthly' },

  // Dataset and utility pages
  { path: '/sb-253-readiness-checklist', priority: 0.8, changeFrequency: 'monthly' },
  { path: '/scope-2-factor-library', priority: 0.8, changeFrequency: 'monthly' },
  { path: '/carbon-reporting-template', priority: 0.7, changeFrequency: 'monthly' },

  // Auth entry points
  { path: '/login', priority: 0.5, changeFrequency: 'yearly' },
  { path: '/register', priority: 0.6, changeFrequency: 'yearly' },

  // Legal
  { path: '/privacy', priority: 0.4, changeFrequency: 'monthly' },
  { path: '/terms', priority: 0.4, changeFrequency: 'monthly' },
];

export default function sitemap(): MetadataRoute.Sitemap {
  const now = new Date();
  return STATIC_PATHS.map(({ path, priority, changeFrequency }) => ({
    url: `${SITE_URL}${path}`,
    lastModified: now,
    changeFrequency,
    priority,
  }));
}
