import Link from 'next/link';
import type { Metadata } from 'next';
import { SiteNav } from '../components/SiteNav';

export const metadata: Metadata = {
  title: 'About | OffGridFlow',
  description: 'Meet the founder behind OffGridFlow — enterprise carbon compliance built by Paul Timchuk.',
};

export default function AboutPage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <SiteNav />

      <section className="px-6 py-24">
        <div className="mx-auto max-w-3xl">
          <h1 className="mb-8 text-4xl font-bold text-white">About OffGridFlow</h1>

          <div className="space-y-6 text-gray-300 leading-relaxed">
            <p>
              OffGridFlow is a carbon accounting and compliance platform built to make regulatory
              reporting accessible to companies that can&apos;t afford six-figure consulting engagements.
            </p>

            <p>
              Starting in 2024&ndash;2026, thousands of companies face mandatory emissions
              disclosure under CSRD (EU), SEC Climate Rules (US), California SB 253, CBAM, and
              IFRS S2. Legacy solutions from Big 4 firms run $50,000&ndash;$200,000+ per year.
              OffGridFlow calculates Scope 1, 2, and 3 emissions using GHG Protocol methodology
              (documented at <a href="/methodology" className="text-primary-400 hover:underline">/methodology</a>)
              and generates compliance reports at a fraction of that cost.
            </p>

            <h2 className="mt-12 text-2xl font-semibold text-white">Founder</h2>

            <p>
              <strong className="text-white">Paul Timchuk</strong> &mdash; Founder &amp; CEO.
              Systems architect with 12+ years of federal regulatory compliance experience and
              deep expertise in Go, Python, Kubernetes, and multi-cloud infrastructure
              (AWS/Azure/GCP). Paul built OffGridFlow from the ground up: the GHG Protocol
              calculation engine, cloud data connectors, compliance report generators, and the
              entire platform infrastructure.
            </p>

            <h2 className="mt-12 text-2xl font-semibold text-white">Technical Foundation</h2>

            <ul className="list-inside list-disc space-y-2 text-gray-400">
              <li>Scope 1, 2, 3 calculation engine using <a href="/methodology" className="text-primary-400 hover:underline">GHG Protocol methodology</a> with EPA eGRID, DEFRA, IEA, and IPCC emission factors</li>
              <li>Automated cloud connectors: AWS CUR, Azure Carbon API, GCP Carbon Footprint API</li>
              <li>Enterprise SAP &amp; utility bill integration</li>
              <li>Multi-framework reporting: CSRD/ESRS, SEC, SB 253, CBAM, IFRS S2</li>
              <li>PDF, XBRL, and Excel compliance exports</li>
              <li>Multi-tenant architecture with RBAC, 2FA, and audit logging</li>
            </ul>

            <h2 className="mt-12 text-2xl font-semibold text-white">Security Roadmap</h2>

            <ul className="list-inside list-disc space-y-2 text-gray-400">
              <li>SOC 2 Type I &mdash; targeted Q3 2026</li>
              <li>SOC 2 Type II &mdash; targeted Q1 2027</li>
              <li>ISO 27001 &mdash; targeted Q2 2027</li>
            </ul>
          </div>

          <div className="mt-16 flex gap-4">
            <Link
              href="/register?plan=starter"
              className="rounded-lg bg-primary-600 px-6 py-3 text-sm font-medium text-white transition hover:bg-primary-500"
            >
              Start Free Trial
            </Link>
            <Link
              href="/demo"
              className="rounded-lg border border-gray-700 px-6 py-3 text-sm font-medium text-gray-300 transition hover:border-gray-500 hover:text-white"
            >
              Review Workflow
            </Link>
          </div>
        </div>
      </section>
    </div>
  );
}
