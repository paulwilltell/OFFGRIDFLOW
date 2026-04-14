import Link from 'next/link';
import type { Metadata } from 'next';
import { SiteNav } from '../components/SiteNav';

export const metadata: Metadata = {
  title: 'Security & Trust | OffGridFlow',
  description: 'How OffGridFlow protects your data — security practices, certifications roadmap, and trust center.',
};

export default function SecurityPage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <SiteNav />

      <section className="px-6 py-24">
        <div className="mx-auto max-w-3xl">
          <h1 className="mb-2 text-4xl font-bold text-white">Security &amp; Trust Center</h1>
          <p className="mb-12 text-gray-400">
            How we protect your emissions data and maintain platform integrity.
          </p>

          <div className="space-y-10 text-gray-300 leading-relaxed">
            <section>
              <h2 className="mb-4 text-2xl font-semibold text-white">Infrastructure Security</h2>
              <div className="grid gap-4 sm:grid-cols-2">
                {[
                  { label: 'Encryption in Transit', detail: 'TLS 1.2+ on all connections' },
                  { label: 'Encryption at Rest', detail: 'AES-256 database encryption' },
                  { label: 'Multi-Tenant Isolation', detail: 'Tenant-scoped data access controls' },
                  { label: 'Network Security', detail: 'Private networking, firewall rules' },
                ].map((item) => (
                  <div
                    key={item.label}
                    className="rounded-lg border border-gray-800 bg-dark-800 p-4"
                  >
                    <div className="text-sm font-medium text-white">{item.label}</div>
                    <div className="mt-1 text-xs text-gray-500">{item.detail}</div>
                  </div>
                ))}
              </div>
            </section>

            <section>
              <h2 className="mb-4 text-2xl font-semibold text-white">Authentication &amp; Access Control</h2>
              <ul className="list-inside list-disc space-y-2 text-gray-400">
                <li>JWT-based session management with 7-day token expiry</li>
                <li>Two-factor authentication (2FA) support</li>
                <li>Role-based access control (Admin, User, Read-only)</li>
                <li>API key management with scoped permissions</li>
                <li>CSRF protection on all state-changing operations</li>
                <li>Account lockout after 5 failed login attempts</li>
                <li>Email verification for new accounts</li>
              </ul>
            </section>

            <section>
              <h2 className="mb-4 text-2xl font-semibold text-white">Data Handling</h2>
              <ul className="list-inside list-disc space-y-2 text-gray-400">
                <li>Emissions data is processed and stored exclusively for your reporting needs</li>
                <li>No cross-tenant data access — strict isolation at the database level</li>
                <li>Full audit logging of data access and modifications</li>
                <li>Data export available at any time (JSON, CSV, PDF)</li>
                <li>Data deletion within 30 days of account termination</li>
              </ul>
            </section>

            <section>
              <h2 className="mb-4 text-2xl font-semibold text-white">Certification Roadmap</h2>
              <p className="mb-4 text-gray-400">
                We are pursuing industry-standard certifications on the following timeline:
              </p>
              <div className="space-y-3">
                {[
                  { cert: 'SOC 2 Type I', target: 'Q3 2026', status: 'In preparation' },
                  { cert: 'SOC 2 Type II', target: 'Q1 2027', status: 'Planned' },
                  { cert: 'ISO 27001', target: 'Q2 2027', status: 'Planned' },
                ].map((item) => (
                  <div
                    key={item.cert}
                    className="flex items-center justify-between rounded-lg border border-gray-800 bg-dark-800 p-4"
                  >
                    <div>
                      <div className="text-sm font-medium text-white">{item.cert}</div>
                      <div className="mt-0.5 text-xs text-gray-500">{item.status}</div>
                    </div>
                    <div className="text-sm text-gray-400">{item.target}</div>
                  </div>
                ))}
              </div>
            </section>

            <section>
              <h2 className="mb-4 text-2xl font-semibold text-white">Responsible Disclosure</h2>
              <p>
                If you discover a security vulnerability, please report it to{' '}
                <a href="mailto:contact@off-grid-flow.com" className="text-primary-400 hover:underline">
                  contact@off-grid-flow.com
                </a>
                . We take all reports seriously and will respond within 48 hours.
              </p>
            </section>
          </div>
        </div>
      </section>
    </div>
  );
}
