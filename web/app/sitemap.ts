import type { MetadataRoute } from 'next';
import { SITE_URL } from '@/lib/seo';

/**
 * Dynamic sitemap for off-grid-flow.com. Covers all public marketing,
 * framework, role, comparison, dataset, trust, and legal pages.
 */

const STATIC_PATHS: Array<{ path: string; priority: number; changeFrequency: MetadataRoute.Sitemap[number]['changeFrequency'] }> = [
  { path: '/', priority: 1.0, changeFrequency: 'weekly' },
  { path: '/pricing', priority: 0.9, changeFrequency: 'monthly' },

  // Trust / legal
  { path: '/trust', priority: 0.7, changeFrequency: 'monthly' },
  { path: '/security', priority: 0.6, changeFrequency: 'monthly' },
  { path: '/privacy', priority: 0.4, changeFrequency: 'monthly' },
  { path: '/terms', priority: 0.4, changeFrequency: 'monthly' },

  // Auth entry points
  { path: '/login', priority: 0.5, changeFrequency: 'yearly' },
  { path: '/register', priority: 0.6, changeFrequency: 'yearly' },
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
