import Link from 'next/link';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Trust Center | OffGridFlow',
  description: 'Security, privacy, compliance, and architecture documentation for enterprise procurement.',
};

export default function TrustCenterPage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <nav className="border-b border-gray-800/50 bg-dark-900/80 backdrop-blur-md">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
          <Link href="/" className="text-xl font-bold text-white">OffGridFlow</Link>
          <Link href="/" className="text-sm text-gray-400 hover:text-white">Back to Home</Link>
        </div>
      </nav>

      <div className="mx-auto max-w-5xl px-6 py-16">
        <h1 className="mb-2 text-3xl font-bold text-white">Trust Center</h1>
        <p className="mb-10 text-gray-400">
          Everything your security, legal, and procurement teams need to evaluate OffGridFlow.
        </p>

        {/* Quick answers grid */}
        <div className="mb-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[
            { title: 'SOC 2 Type I', status: 'In Preparation', target: 'Q3 2026', color: 'amber' },
            { title: 'Data Encryption', status: 'TLS 1.2+ in transit, AES-256 at rest', target: 'Active', color: 'green' },
            { title: 'Multi-Tenant Isolation', status: 'Tenant-scoped data access on every query', target: 'Active', color: 'green' },
            { title: 'MFA / 2FA', status: 'TOTP-based two-factor authentication', target: 'Available', color: 'green' },
            { title: 'SSO / SAML', status: 'Planned for Enterprise tier', target: 'Q4 2026', color: 'amber' },
            { title: 'GDPR Compliance', status: 'Data export, deletion, retention controls', target: 'Active', color: 'green' },
          ].map(item => (
            <div key={item.title} className="rounded-lg border border-gray-800 bg-gray-800/30 p-4">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-white">{item.title}</span>
                <span className={`rounded px-2 py-0.5 text-[10px] font-medium ${
                  item.color === 'green' ? 'bg-green-500/10 text-green-400' : 'bg-amber-500/10 text-amber-400'
                }`}>{item.target}</span>
              </div>
              <div className="mt-1 text-xs text-gray-500">{item.status}</div>
            </div>
          ))}
        </div>

        {/* Security Architecture */}
        <section className="mb-10">
          <h2 className="mb-4 text-xl font-semibold text-white">Security Architecture</h2>
          <div className="space-y-3">
            {[
              { area: 'Authentication', details: 'JWT-based sessions with 7-day TTL. Account lockout after 5 failed attempts (15-minute cooldown). Email verification for new accounts. CSRF protection on all state-changing operations.' },
              { area: 'Authorization', details: 'Role-based access control (Admin, User). Tenant-scoped data isolation enforced at the database query level. Subscription-tier enforcement on all premium API endpoints. Admin accounts bypass subscription checks for platform management.' },
              { area: 'Data Protection', details: 'All data encrypted in transit via TLS 1.2+. Database encryption at rest managed by infrastructure provider. API keys hashed before storage. Secrets managed via environment variables with no hardcoded credentials in source code.' },
              { area: 'Network Security', details: 'CORS restricted to explicit allowed origins (no wildcards). Security headers on every response: X-Frame-Options DENY, HSTS with preload, Content-Security-Policy, X-Content-Type-Options nosniff. Rate limiting on all HTTP methods.' },
              { area: 'Audit Logging', details: 'Immutable change log tracks all data modifications (entity, field, old/new values, user, timestamp). Calculation ledger records every emission calculation with full formula transparency. Approval workflow records preparer, reviewer, and approver with timestamps.' },
              { area: 'File Upload Security', details: 'CSV uploads validated by file extension and MIME type detection. Maximum upload size: 50MB. Content sniffing rejects non-text file types. No executable content accepted.' },
            ].map(item => (
              <div key={item.area} className="rounded-lg border border-gray-800 bg-gray-800/30 p-5">
                <div className="text-sm font-medium text-white">{item.area}</div>
                <div className="mt-2 text-xs leading-relaxed text-gray-400">{item.details}</div>
              </div>
            ))}
          </div>
        </section>

        {/* Privacy & Data Governance */}
        <section className="mb-10">
          <h2 className="mb-4 text-xl font-semibold text-white">Privacy &amp; Data Governance</h2>
          <div className="grid gap-4 sm:grid-cols-2">
            {[
              { title: 'Data Ownership', detail: 'All emission data, reports, and uploaded evidence remain the property of the customer. OffGridFlow processes data solely for service delivery.' },
              { title: 'Data Export', detail: 'Full tenant data export available via API (JSON). Includes users, activities, calculation ledger, and change log. Available at any time during subscription.' },
              { title: 'Data Deletion', detail: 'Deletion request via API or email. 30-day retention period for final export. Permanent deletion after retention window. Audit logs retained per regulatory requirement.' },
              { title: 'Retention Schedule', detail: 'Emission data: subscription + 90 days. Calculation ledger: subscription + 7 years (audit trail). User accounts: subscription + 30 days. Evidence files: subscription + 90 days.' },
              { title: 'Subprocessors', detail: 'Railway (infrastructure hosting), Stripe (payment processing), SendGrid (transactional email). No customer data shared with AI training or analytics services.' },
              { title: 'International Transfers', detail: 'Data hosted in US-West region. EU customers: Standard Contractual Clauses available. No data processing outside the hosting region without customer consent.' },
            ].map(item => (
              <div key={item.title} className="rounded-lg border border-gray-800 bg-gray-800/30 p-4">
                <div className="text-sm font-medium text-white">{item.title}</div>
                <div className="mt-1 text-xs text-gray-400">{item.detail}</div>
              </div>
            ))}
          </div>
        </section>

        {/* Incident Response */}
        <section className="mb-10">
          <h2 className="mb-4 text-xl font-semibold text-white">Incident Response</h2>
          <div className="rounded-lg border border-gray-800 bg-gray-800/30 p-5">
            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <div className="text-xs font-medium text-gray-400">P1 (Service Down)</div>
                <div className="mt-1 text-sm text-white">Response within 30 minutes</div>
                <div className="text-xs text-gray-500">Email notification to all affected customers</div>
              </div>
              <div>
                <div className="text-xs font-medium text-gray-400">P2 (Degraded Performance)</div>
                <div className="mt-1 text-sm text-white">Response within 4 hours</div>
                <div className="text-xs text-gray-500">Status page updated, email for extended outages</div>
              </div>
              <div>
                <div className="text-xs font-medium text-gray-400">Post-Incident Review</div>
                <div className="mt-1 text-sm text-white">Within 72 hours</div>
                <div className="text-xs text-gray-500">Root cause analysis shared with affected customers</div>
              </div>
              <div>
                <div className="text-xs font-medium text-gray-400">Security Vulnerability</div>
                <div className="mt-1 text-sm text-white">Report to contact@off-grid-flow.com</div>
                <div className="text-xs text-gray-500">Acknowledgment within 48 hours, remediation tracked</div>
              </div>
            </div>
          </div>
        </section>

        {/* Compliance Readiness */}
        <section className="mb-10">
          <h2 className="mb-4 text-xl font-semibold text-white">Compliance &amp; Certifications Roadmap</h2>
          <div className="space-y-3">
            {[
              { cert: 'SOC 2 Type I', timeline: 'Q3 2026', status: 'In preparation', detail: 'Point-in-time assessment of security controls. Covers: Security, Availability, Confidentiality trust service criteria.' },
              { cert: 'SOC 2 Type II', timeline: 'Q1 2027', status: 'Planned', detail: 'Operating effectiveness of controls over a 6-month observation period.' },
              { cert: 'ISO 27001', timeline: 'Q2 2027', status: 'Planned', detail: 'Information security management system certification.' },
              { cert: 'Pen Testing', timeline: 'Q3 2026', status: 'Planned', detail: 'Third-party penetration test with remediation tracking. Results available to enterprise customers under NDA.' },
            ].map(item => (
              <div key={item.cert} className="flex items-center justify-between rounded-lg border border-gray-800 bg-gray-800/30 p-4">
                <div>
                  <div className="text-sm font-medium text-white">{item.cert}</div>
                  <div className="mt-0.5 text-xs text-gray-500">{item.detail}</div>
                </div>
                <div className="ml-4 text-right">
                  <div className="text-xs font-medium text-gray-300">{item.timeline}</div>
                  <div className="text-[10px] text-gray-600">{item.status}</div>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Downloadable resources */}
        <section className="mb-10">
          <h2 className="mb-4 text-xl font-semibold text-white">Resources for Procurement</h2>
          <div className="grid gap-3 sm:grid-cols-2">
            {[
              { name: 'Security Architecture Overview', href: '/methodology', type: 'Web Page' },
              { name: 'Privacy Policy', href: '/privacy', type: 'Web Page' },
              { name: 'Terms of Service', href: '/terms', type: 'Web Page' },
              { name: 'Data Retention Policy', href: '/trust#retention', type: 'Web Page' },
              { name: 'Methodology Library', href: '/methodology', type: 'Web Page' },
              { name: 'System Status', href: '/status', type: 'Live Dashboard' },
            ].map(item => (
              <Link
                key={item.name}
                href={item.href}
                className="flex items-center justify-between rounded-lg border border-gray-800 bg-gray-800/30 p-4 transition hover:border-gray-700"
              >
                <span className="text-sm text-gray-300">{item.name}</span>
                <span className="rounded bg-gray-700/50 px-2 py-0.5 text-[10px] text-gray-400">{item.type}</span>
              </Link>
            ))}
          </div>
        </section>

        {/* Contact */}
        <section className="rounded-xl border border-primary-600/20 bg-primary-600/5 p-6 text-center">
          <h3 className="text-lg font-semibold text-white">Need more detail?</h3>
          <p className="mt-2 text-sm text-gray-400">
            For security questionnaires, DPA requests, or custom procurement requirements, contact us directly.
          </p>
          <a
            href="mailto:contact@off-grid-flow.com"
            className="mt-4 inline-block rounded-lg bg-primary-600 px-6 py-2.5 text-sm font-medium text-white hover:bg-primary-500"
          >
            Contact Security Team
          </a>
        </section>
      </div>
    </div>
  );
}
