import Link from 'next/link';
import type { Metadata } from 'next';
import { SiteNav } from '../components/SiteNav';

export const metadata: Metadata = {
  title: 'Privacy Policy | OffGridFlow',
  description: 'OffGridFlow privacy policy — how we collect, use, and protect your data.',
};

const subprocessors = [
  {
    name: 'Stripe, Inc.',
    purpose: 'Payment processing and subscription billing',
    dataAccessed: 'Name, billing email, payment method, billing address',
    region: 'United States',
    safeguards: 'PCI DSS Level 1; SOC 1/2 Type II; signed DPA',
    url: 'https://stripe.com/privacy',
  },
  {
    name: 'Railway Corp.',
    purpose: 'Application hosting, compute, managed PostgreSQL',
    dataAccessed: 'All Customer Data at rest (encrypted) and in transit within Railway infrastructure',
    region: 'United States (us-west)',
    safeguards: 'SOC 2 Type II; encrypted volumes; network isolation; signed DPA',
    url: 'https://railway.com/legal/privacy',
  },
  {
    name: 'Twilio SendGrid, Inc.',
    purpose: 'Transactional email (verification, password reset, notifications)',
    dataAccessed: 'Recipient email address, email content, delivery metadata',
    region: 'United States',
    safeguards: 'SOC 2 Type II; ISO 27001; signed DPA',
    url: 'https://www.twilio.com/en-us/legal/privacy',
  },
  {
    name: 'Google LLC (Analytics & Ads)',
    purpose: 'Marketing analytics and conversion measurement on public pages only',
    dataAccessed: 'Anonymous page views, device/browser, referrer, conversion events on marketing pages (not in-app)',
    region: 'Global',
    safeguards: 'Google Ads Data Processing Terms; IP anonymization enabled',
    url: 'https://policies.google.com/privacy',
  },
  {
    name: 'Cloudflare, Inc. (CDN / DNS)',
    purpose: 'DNS resolution and edge delivery for marketing domain',
    dataAccessed: 'IP address, request metadata',
    region: 'Global edge',
    safeguards: 'SOC 2 Type II; ISO 27001; signed DPA',
    url: 'https://www.cloudflare.com/privacypolicy/',
  },
];

export default function PrivacyPage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <SiteNav />

      <section className="px-6 py-24">
        <div className="mx-auto max-w-3xl">
          <h1 className="mb-2 text-4xl font-bold text-white">Privacy Policy</h1>
          <p className="mb-4 text-sm text-gray-500">Last updated: April 2026</p>

          <div className="mb-10 rounded-xl border border-amber-500/30 bg-amber-500/5 p-4 text-sm text-amber-200">
            <strong className="block mb-1">Draft &mdash; Pending Attorney Review.</strong>
            This Privacy Policy is a working draft describing current data-handling practices. It
            does not constitute legal advice. Customers subject to specific regulatory regimes
            (GDPR, CCPA, HIPAA, etc.) should request a Data Processing Addendum from{' '}
            <a href="mailto:contact@off-grid-flow.com" className="underline">contact@off-grid-flow.com</a>.
          </div>

          <div className="space-y-8 text-gray-300 leading-relaxed">
            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">1. Data Controller</h2>
              <p>
                OffGridFlow LLC (&quot;OffGridFlow,&quot; &quot;we,&quot; &quot;us&quot;) is the
                data controller for personal data collected via our marketing pages and the data
                processor for Customer Data uploaded to the Platform. You (or your organization)
                remain the controller of your Customer Data.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">2. Information We Collect</h2>
              <p className="mb-3">
                <strong className="text-white">Account information:</strong> name, email address,
                company name, job title, and password hash (never the plaintext password).
              </p>
              <p className="mb-3">
                <strong className="text-white">Billing information:</strong> processed directly by
                Stripe; OffGridFlow stores only a Stripe customer identifier and subscription status.
                OffGridFlow never receives or stores full card numbers, CVVs, or bank details.
              </p>
              <p className="mb-3">
                <strong className="text-white">Customer Data:</strong> emissions activity records,
                utility bills, cloud carbon footprint data, facility identifiers, and other
                content you import or connect to the Platform.
              </p>
              <p>
                <strong className="text-white">Usage data:</strong> IP address, user agent, pages
                visited, API endpoints called, timestamps. Used for security, auditability, and
                service operations.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">3. How We Use Your Information</h2>
              <ul className="list-inside list-disc space-y-1 text-gray-400">
                <li>To provide, maintain, and improve the OffGridFlow Platform</li>
                <li>To calculate emissions and generate draft compliance reports</li>
                <li>To process payments and manage subscriptions (via Stripe)</li>
                <li>To send service-related communications (via SendGrid)</li>
                <li>To detect, prevent, and investigate fraud, abuse, and security incidents</li>
                <li>To comply with legal obligations and enforce our Terms of Service</li>
              </ul>
              <p className="mt-3">
                We do <strong className="text-white">not</strong> use Customer Data to train
                third-party machine-learning models. We do not sell or rent personal information.
                We do not use Customer Data for advertising or profiling.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">4. Legal Bases (GDPR)</h2>
              <p>
                We process personal data under the following legal bases: (a) <em>contract</em> —
                to provide the Platform you subscribed to; (b) <em>legitimate interests</em> — for
                security, fraud prevention, and service improvement; (c) <em>legal obligation</em> —
                for tax, accounting, and regulatory compliance; (d) <em>consent</em> — where
                required for marketing communications or non-essential cookies.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">5. Data Protection and Security</h2>
              <p>
                All data is encrypted in transit (TLS 1.2+) and at rest (AES-256 via managed
                Postgres volumes). We enforce multi-tenant isolation so one customer cannot access
                another&apos;s data; isolation is enforced at the database query level
                (<code className="rounded bg-gray-800 px-1 text-xs">WHERE tenant_id = $1</code>).
                Access to production data is restricted to named administrators with MFA.
                See the{' '}
                <Link href="/trust" className="text-primary-400 hover:underline">Trust Center</Link>{' '}
                for the full security architecture.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">6. Subprocessors</h2>
              <p className="mb-4">
                OffGridFlow engages the following third-party subprocessors to operate the Platform.
                Each is contractually bound to maintain appropriate security and confidentiality.
              </p>
              <div className="overflow-x-auto rounded-xl border border-gray-800">
                <table className="w-full text-sm">
                  <thead className="bg-gray-800/50 text-left text-xs text-gray-500">
                    <tr>
                      <th className="px-3 py-2">Subprocessor</th>
                      <th className="px-3 py-2">Purpose</th>
                      <th className="px-3 py-2">Data accessed</th>
                      <th className="px-3 py-2">Region</th>
                      <th className="px-3 py-2">Safeguards</th>
                    </tr>
                  </thead>
                  <tbody className="text-gray-300">
                    {subprocessors.map((sp) => (
                      <tr key={sp.name} className="border-t border-gray-800/50 align-top">
                        <td className="px-3 py-2">
                          <a href={sp.url} className="text-primary-400 hover:underline" target="_blank" rel="noopener noreferrer">
                            {sp.name}
                          </a>
                        </td>
                        <td className="px-3 py-2 text-xs">{sp.purpose}</td>
                        <td className="px-3 py-2 text-xs">{sp.dataAccessed}</td>
                        <td className="px-3 py-2 text-xs">{sp.region}</td>
                        <td className="px-3 py-2 text-xs">{sp.safeguards}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <p className="mt-3 text-xs text-gray-500">
                This list is updated when material changes occur. Customers may subscribe to
                subprocessor change notifications via a signed DPA.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">7. International Data Transfers</h2>
              <p>
                OffGridFlow primarily hosts data in the United States. Transfers of personal data
                from the EEA, UK, or Switzerland to the U.S. rely on the EU Standard Contractual
                Clauses (SCCs) and the UK International Data Transfer Addendum as applicable.
                Customers may request a copy of the SCCs from contact@off-grid-flow.com.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">8. GDPR Rights (EU / UK Users)</h2>
              <p>
                If you are in the European Economic Area, United Kingdom, or Switzerland, you have
                the right to: access your personal data; correct inaccurate data; delete your data;
                restrict or object to processing; portability (receive your data in machine-readable
                format); and to withdraw consent where processing is consent-based. Lodge
                complaints with your supervisory authority. Exercise rights at{' '}
                <a href="mailto:contact@off-grid-flow.com" className="text-primary-400 hover:underline">
                  contact@off-grid-flow.com
                </a>
                .
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">9. CCPA / CPRA Rights (California Users)</h2>
              <p>
                California residents have the rights to: know what personal information is
                collected; access and portability; delete personal information; correct inaccurate
                information; and limit use of sensitive personal information. OffGridFlow does
                not sell or share personal information for cross-context behavioral advertising.
                You may exercise these rights at contact@off-grid-flow.com or via the{' '}
                <Link href="/privacy#do-not-sell" className="text-primary-400 hover:underline">
                  Do Not Sell or Share
                </Link>{' '}
                section below. You have the right not to be discriminated against for exercising
                any of these rights.
              </p>
            </section>

            <section id="do-not-sell">
              <h2 className="mb-3 text-xl font-semibold text-white">10. Do Not Sell or Share My Personal Information</h2>
              <p>
                OffGridFlow does not sell personal information and does not share personal
                information for cross-context behavioral advertising. If this policy changes, a
                Do Not Sell mechanism will be provided here. California residents may confirm
                non-sale status in writing by emailing{' '}
                <a href="mailto:contact@off-grid-flow.com" className="text-primary-400 hover:underline">
                  contact@off-grid-flow.com
                </a>
                .
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">11. Cookies and Tracking</h2>
              <p>
                Marketing pages use strictly necessary cookies and, where consented, analytics and
                conversion measurement cookies (Google Ads). Cookies are not used within the
                authenticated Platform beyond session management. See our cookie disclosure on
                first visit.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">12. Data Retention</h2>
              <p>
                Account data is retained for the duration of your subscription plus thirty (30)
                days for export recovery. Emissions activity data and calculation ledger entries
                are retained for seven (7) years for audit and compliance purposes, consistent
                with common accounting and regulatory record-keeping expectations, unless an
                earlier deletion is requested. Audit logs are retained for seven (7) years. See
                the{' '}
                <Link href="/trust" className="text-primary-400 hover:underline">Trust Center</Link>{' '}
                for the full retention schedule.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">13. Children&apos;s Privacy</h2>
              <p>
                The Platform is not directed at children under 16 and we do not knowingly collect
                personal information from children. If we become aware of such collection, we will
                delete the information.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">14. Security Incidents</h2>
              <p>
                OffGridFlow will notify affected customers of a personal data breach without undue
                delay and in any case within seventy-two (72) hours of becoming aware, where required
                by law and feasible, with the information required by applicable regulation.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">15. Changes to This Policy</h2>
              <p>
                We may update this Privacy Policy from time to time. Material changes will be
                communicated via email or in-app notification at least thirty (30) days before
                taking effect.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">16. Contact</h2>
              <p>
                For privacy questions or to exercise your rights:{' '}
                <a href="mailto:contact@off-grid-flow.com" className="text-primary-400 hover:underline">
                  contact@off-grid-flow.com
                </a>
              </p>
              <p className="mt-2">OffGridFlow LLC</p>
            </section>
          </div>
        </div>
      </section>
    </div>
  );
}
