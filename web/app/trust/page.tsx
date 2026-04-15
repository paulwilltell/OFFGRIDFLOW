import Link from 'next/link';
import type { Metadata } from 'next';
import { SiteNav } from '../components/SiteNav';
import { SiteFooter } from '../components/SiteFooter';

export const metadata: Metadata = {
  title: 'Trust Center | OffGridFlow',
  description: 'Security, privacy, compliance, and architecture documentation for enterprise procurement.',
};

export default function TrustCenterPage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <SiteNav />

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

        {/* RBAC Matrix */}
        <section className="mb-10">
          <h2 className="mb-4 text-xl font-semibold text-white">Role-Based Access Control (RBAC)</h2>
          <p className="mb-4 text-sm text-gray-400">
            OffGridFlow enforces role-based permissions at the API layer. Every request is validated against the authenticated user&apos;s role before data is returned.
          </p>
          <div className="overflow-x-auto rounded-xl border border-gray-800">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-800 bg-gray-800/50">
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Capability</th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500">Admin</th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500">User</th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-gray-500">Viewer</th>
                </tr>
              </thead>
              <tbody className="text-gray-300">
                {[
                  ['View dashboard and emissions data', true, true, true],
                  ['Upload CSV / connect data sources', true, true, false],
                  ['Create and edit activities', true, true, false],
                  ['Generate compliance reports', true, true, false],
                  ['Submit reports for approval', true, true, false],
                  ['Approve or reject reports', true, false, false],
                  ['Lock factor snapshots', true, false, false],
                  ['Manage users and roles', true, false, false],
                  ['Configure billing and subscription', true, false, false],
                  ['Export all organization data', true, false, false],
                  ['Request data deletion', true, false, false],
                  ['View audit logs and change history', true, true, true],
                ].map(([cap, admin, user, viewer]) => (
                  <tr key={cap as string} className="border-b border-gray-800/30">
                    <td className="px-4 py-2 text-white">{cap as string}</td>
                    <td className="px-4 py-2 text-center">{admin ? <span className="text-green-400">Yes</span> : <span className="text-gray-600">No</span>}</td>
                    <td className="px-4 py-2 text-center">{user ? <span className="text-green-400">Yes</span> : <span className="text-gray-600">No</span>}</td>
                    <td className="px-4 py-2 text-center">{viewer ? <span className="text-green-400">Yes</span> : <span className="text-gray-600">No</span>}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>

        <section className="mb-10">
          <h2 className="mb-4 text-xl font-semibold text-white">Security Control Artifacts</h2>
          <div className="grid gap-4 lg:grid-cols-2">
            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <h3 className="text-sm font-semibold text-white">Tenant Isolation Test Artifact</h3>
              <p className="mt-2 text-xs leading-relaxed text-gray-400">
                In the integration suite, a second registered tenant authenticates successfully and then calls{' '}
                <code className="rounded bg-gray-800 px-1 text-primary-400">GET /api/emissions/activities</code>.
                The response is <span className="text-white">200 OK</span> with{' '}
                <span className="text-white">0 activities</span> from the first tenant, proving no cross-tenant data leakage
                through the main emissions activity endpoint.
              </p>
            </div>

            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <h3 className="text-sm font-semibold text-white">MFA Challenge Flow</h3>
              <p className="mt-2 text-xs leading-relaxed text-gray-400">
                The login flow supports a second authentication step with a six-digit one-time code.
                After primary credential verification, users complete{' '}
                <code className="rounded bg-gray-800 px-1 text-primary-400">/api/auth/verify-2fa</code>{' '}
                using a temporary token and TOTP code before a session token is issued.
              </p>
            </div>
          </div>
        </section>

        {/* Data Governance Walkthrough */}
        <section className="mb-10">
          <h2 className="mb-4 text-xl font-semibold text-white">Data Governance Walkthrough</h2>
          <p className="mb-4 text-sm text-gray-400">
            OffGridFlow provides self-service data governance endpoints. Admins can export, request deletion, and review retention policies without contacting support.
          </p>
          <div className="space-y-4">
            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <h3 className="text-sm font-semibold text-white">Data Export (GET /api/governance/export)</h3>
              <p className="mt-1 text-xs text-gray-400">
                Admin-only. Returns a JSON package containing all organization data: users, activities, calculation ledger entries, and change log.
                Response includes <code className="rounded bg-gray-800 px-1 text-primary-400">exported_at</code> timestamp and <code className="rounded bg-gray-800 px-1 text-primary-400">tenant_name</code>.
                Download as a file via Content-Disposition header.
              </p>
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <h3 className="text-sm font-semibold text-white">Deletion Request (POST /api/governance/delete-request)</h3>
              <p className="mt-1 text-xs text-gray-400">
                Admin-only. Initiates a 30-day retention window. The request is logged in the change log with the requesting user&apos;s ID.
                Data is retained for 30 days to allow cancellation, then permanently removed.
                Response includes <code className="rounded bg-gray-800 px-1 text-primary-400">deletion_date</code> and <code className="rounded bg-gray-800 px-1 text-primary-400">retention_days</code>.
              </p>
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <h3 className="text-sm font-semibold text-white">Retention Policy (GET /api/governance/retention)</h3>
              <p className="mt-1 text-xs text-gray-400">
                Returns the organization&apos;s data retention schedule:
              </p>
              <ul className="mt-2 space-y-1 text-xs text-gray-500">
                <li><strong className="text-gray-300">Emission data:</strong> Subscription + 90 days</li>
                <li><strong className="text-gray-300">Calculation ledger:</strong> Subscription + 7 years (audit trail)</li>
                <li><strong className="text-gray-300">User accounts:</strong> Subscription + 30 days</li>
                <li><strong className="text-gray-300">Change log:</strong> Subscription + 7 years</li>
                <li><strong className="text-gray-300">Evidence files:</strong> Subscription + 90 days</li>
              </ul>
              <p className="mt-2 text-xs text-gray-500">
                Export formats: JSON (full dataset), PDF (reports), CSV (activities), XBRL (compliance).
                All emission data, reports, and uploaded evidence remain the property of the customer.
              </p>
            </div>
          </div>
        </section>

        {/* Data Classification */}
        <section className="mb-10">
          <h2 className="mb-4 text-xl font-semibold text-white">Data Classification</h2>
          <div className="overflow-x-auto rounded-xl border border-gray-800">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-800 bg-gray-800/50">
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Data Type</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Classification</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">Handling</th>
                </tr>
              </thead>
              <tbody className="text-gray-300">
                {[
                  ['User credentials (passwords)', 'Secret', 'bcrypt hashed, never stored in plaintext, never logged'],
                  ['API keys', 'Secret', 'SHA-256 hashed at rest, prefix-only display in UI'],
                  ['Cloud connector credentials', 'Confidential', 'Encrypted at rest (AES-256), tenant-scoped access only'],
                  ['Emission activity data', 'Internal', 'Tenant-isolated, soft-deleted, exportable, 90-day post-subscription retention'],
                  ['Calculation results', 'Internal', 'Immutable ledger, 7-year retention for audit compliance'],
                  ['Compliance reports', 'Internal', 'Versioned, approval-gated, export with checksum verification'],
                  ['Audit logs', 'Internal', 'Append-only, 7-year retention, includes IP and user agent'],
                  ['Email addresses', 'PII', 'Used for authentication only, exportable via governance API, deletable on request'],
                  ['Emission factors', 'Public', 'Sourced from EPA, IEA, DEFRA, IPCC — publicly available data'],
                ].map(([type, classification, handling]) => (
                  <tr key={type} className="border-b border-gray-800/30">
                    <td className="px-4 py-2 font-medium text-white">{type}</td>
                    <td className="px-4 py-2">
                      <span className={`rounded px-2 py-0.5 text-[10px] font-medium ${
                        classification === 'Secret' ? 'bg-red-500/10 text-red-400' :
                        classification === 'Confidential' ? 'bg-amber-500/10 text-amber-400' :
                        classification === 'PII' ? 'bg-purple-500/10 text-purple-400' :
                        classification === 'Internal' ? 'bg-blue-500/10 text-blue-400' :
                        'bg-gray-500/10 text-gray-400'
                      }`}>{classification}</span>
                    </td>
                    <td className="px-4 py-2 text-xs text-gray-400">{handling}</td>
                  </tr>
                ))}
              </tbody>
            </table>
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
              { name: 'Data Architecture', href: '/architecture', type: 'Web Page' },
              { name: 'Evidence Pack', href: '/evidence', type: 'Web Page' },
              { name: 'Operations Proof', href: '/operations', type: 'Web Page' },
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

        <section className="mb-10 rounded-2xl border border-gray-800 bg-gray-800/20 p-6">
          <h2 className="text-xl font-semibold text-white">Framework and Role Pages</h2>
          <p className="mt-2 text-sm text-gray-400">
            Procurement and finance reviewers often start here, then branch into the framework-specific buying pages below.
          </p>
          <div className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {[
              { href: '/sb-253-reporting-software', label: 'SB 253 reporting software' },
              { href: '/csrd-reporting-software', label: 'CSRD reporting software' },
              { href: '/ifrs-s2-reporting-software', label: 'IFRS S2 reporting software' },
              { href: '/for-cfos', label: 'For CFOs' },
              { href: '/for-sustainability-managers', label: 'For sustainability managers' },
              { href: '/for-procurement', label: 'For procurement' },
            ].map((item) => (
              <Link key={item.href} href={item.href} className="rounded-xl border border-gray-800/70 bg-gray-900/40 p-4 text-sm text-white transition hover:border-primary-600/40">
                {item.label}
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
      <SiteFooter />
    </div>
  );
}
