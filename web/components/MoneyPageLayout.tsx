import Link from 'next/link';
import type { ReactNode } from 'react';
import { SiteNav } from '../app/components/SiteNav';
import {
  JsonLd,
  breadcrumbListSchema,
  faqPageSchema,
  softwareApplicationSchema,
  type FaqItem,
} from './JsonLd';
import type { BreadcrumbItem } from '@/lib/seo';

interface MoneyPageLayoutProps {
  breadcrumbs: BreadcrumbItem[];
  /** The primary H1 headline for the page. */
  h1: string;
  /** Short subheadline / dek under the H1. */
  dek?: string;
  /** FAQ items rendered and emitted as FAQPage schema. */
  faqs?: FaqItem[];
  /** Slug id suffix used to make script ids unique across pages. */
  slug: string;
  /** UTM content tag for the primary CTA. */
  ctaUtm?: string;
  /** Primary CTA text. Defaults to "Start Free Trial". */
  ctaText?: string;
  /** Secondary CTA shown alongside the primary button. */
  secondaryCtaText?: string;
  secondaryCtaHref?: string;
  /** Page body content. */
  children: ReactNode;
}

/**
 * Shared marketing layout for "money pages" — framework- or role-targeted
 * landing pages that drive organic and paid traffic to a clear CTA.
 *
 * Renders:
 *  - SiteNav (unified marketing navigation)
 *  - Breadcrumb trail (visible) + BreadcrumbList JSON-LD
 *  - H1 + dek
 *  - Content passed as children
 *  - Proof block linking to methodology, evidence, security, status
 *  - FAQ list (rendered) + FAQPage JSON-LD
 *  - Closing CTA
 *  - Trust footer
 */
export function MoneyPageLayout({
  breadcrumbs,
  h1,
  dek,
  faqs,
  slug,
  ctaUtm,
  ctaText = 'Start Free Trial',
  secondaryCtaText = 'Review the workflow',
  secondaryCtaHref = '/demo',
  children,
}: MoneyPageLayoutProps) {
  const utmParam = ctaUtm ? `&utm_source=organic&utm_medium=money_page&utm_campaign=${ctaUtm}` : '';
  const primaryCtaHref = `/register?plan=starter${utmParam}`;

  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <SiteNav />

      <JsonLd id={`ld-breadcrumb-${slug}`} data={breadcrumbListSchema(breadcrumbs)} />
      {faqs && faqs.length > 0 && (
        <JsonLd id={`ld-faq-${slug}`} data={faqPageSchema(faqs)} />
      )}
      <JsonLd id={`ld-software-${slug}`} data={softwareApplicationSchema()} />

      <main className="mx-auto max-w-4xl px-6 py-16">
        {/* Breadcrumbs (visible) */}
        <nav aria-label="Breadcrumb" className="mb-8 text-xs text-gray-500">
          <ol className="flex items-center gap-2">
            {breadcrumbs.map((item, index) => (
              <li key={item.path} className="flex items-center gap-2">
                {index > 0 && <span className="text-gray-700">/</span>}
                {index < breadcrumbs.length - 1 ? (
                  <Link href={item.path} className="hover:text-primary-400">
                    {item.name}
                  </Link>
                ) : (
                  <span className="text-gray-300">{item.name}</span>
                )}
              </li>
            ))}
          </ol>
        </nav>

        {/* Hero */}
        <header className="mb-12">
          <h1 className="text-3xl font-bold leading-tight text-white sm:text-4xl lg:text-5xl">
            {h1}
          </h1>
          {dek && (
            <p className="mt-4 max-w-3xl text-lg leading-relaxed text-gray-400">{dek}</p>
          )}
          <div className="mt-8 flex flex-wrap items-center gap-4">
            <Link
              href={primaryCtaHref}
              className="rounded-lg bg-primary-600 px-7 py-3 text-sm font-semibold text-white shadow-lg shadow-primary-600/20 transition hover:bg-primary-500"
            >
              {ctaText}
            </Link>
            <Link
              href={secondaryCtaHref}
              className="rounded-lg border border-gray-700 px-7 py-3 text-sm font-medium text-gray-300 transition hover:border-gray-500 hover:text-white"
            >
              {secondaryCtaText}
            </Link>
          </div>
        </header>

        {/* Body content */}
        <article className="prose prose-invert max-w-none space-y-6 text-gray-300 leading-relaxed">
          {children}
        </article>

        {/* Proof block */}
        <aside className="mt-16 rounded-2xl border border-gray-800 bg-gray-800/30 p-6">
          <h2 className="text-xs font-medium uppercase tracking-widest text-primary-400">
            Evidence you can verify
          </h2>
          <p className="mt-2 text-sm text-gray-400">
            OffGridFlow publishes its methodology, architecture, and live health so customers can
            inspect how numbers are produced before they ever buy.
          </p>
          <div className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {[
              { label: 'Methodology', href: '/methodology', desc: '5 factor sources, 5 methods, versioned' },
              { label: 'Evidence Pack', href: '/evidence', desc: 'Redacted sample audit packet' },
              { label: 'Architecture', href: '/architecture', desc: 'Data model + traceability chain' },
              { label: 'Live Status', href: '/status', desc: 'Real-time health endpoint' },
            ].map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className="block rounded-lg border border-gray-800/60 bg-gray-900/40 p-3 transition hover:border-primary-600/40"
              >
                <div className="text-sm font-medium text-white">{item.label}</div>
                <div className="mt-1 text-xs text-gray-500">{item.desc}</div>
              </Link>
            ))}
          </div>
        </aside>

        {/* FAQ */}
        {faqs && faqs.length > 0 && (
          <section className="mt-16">
            <h2 className="mb-6 text-2xl font-bold text-white">Frequently Asked Questions</h2>
            <div className="space-y-4">
              {faqs.map((faq) => (
                <details
                  key={faq.question}
                  className="group rounded-xl border border-gray-800 bg-gray-800/30 p-5"
                >
                  <summary className="cursor-pointer text-sm font-semibold text-white marker:hidden">
                    {faq.question}
                  </summary>
                  <p className="mt-3 text-sm text-gray-400">{faq.answer}</p>
                </details>
              ))}
            </div>
          </section>
        )}

        {/* Closing CTA */}
        <section className="mt-16 rounded-2xl border border-primary-600/20 bg-primary-600/5 p-8 text-center">
          <h2 className="text-2xl font-bold text-white">Ready to see it with your data?</h2>
          <p className="mx-auto mt-3 max-w-lg text-sm text-gray-400">
            Upload a CSV and generate your first compliance report in under two hours. No
            consultant engagement required.
          </p>
          <div className="mt-6 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <Link
              href={primaryCtaHref}
              className="rounded-lg bg-primary-600 px-7 py-3 text-sm font-semibold text-white hover:bg-primary-500"
            >
              {ctaText}
            </Link>
            <Link
              href="mailto:contact@off-grid-flow.com?subject=OffGridFlow%20Inquiry"
              className="rounded-lg border border-gray-700 px-7 py-3 text-sm font-medium text-gray-300 hover:border-gray-500 hover:text-white"
            >
              Talk to us
            </Link>
          </div>
          <p className="mt-4 text-[11px] text-gray-500">
            OffGridFlow calculates emissions using documented{' '}
            <Link href="/methodology" className="text-primary-400 hover:underline">
              GHG Protocol methodology
            </Link>
            . Reports are drafts; customers are responsible for verification before regulatory submission.
          </p>
        </section>
      </main>
    </div>
  );
}

export default MoneyPageLayout;
