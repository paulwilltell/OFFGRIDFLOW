import Link from 'next/link';
import type { Metadata } from 'next';
import { SiteNav } from '../components/SiteNav';

export const metadata: Metadata = {
  title: 'Privacy Policy | OffGridFlow',
  description: 'OffGridFlow privacy policy — how we collect, use, and protect your data.',
};

export default function PrivacyPage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <SiteNav />

      <section className="px-6 py-24">
        <div className="mx-auto max-w-3xl">
          <h1 className="mb-2 text-4xl font-bold text-white">Privacy Policy</h1>
          <p className="mb-12 text-sm text-gray-500">Last updated: April 2026</p>

          <div className="space-y-8 text-gray-300 leading-relaxed">
            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">1. Information We Collect</h2>
              <p>
                We collect information you provide directly: name, email address, company name,
                job title, and payment information when you create an account or subscribe to our
                services. We also collect emissions data, utility records, and cloud provider
                carbon footprint data that you import into the platform for calculation and reporting.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">2. How We Use Your Information</h2>
              <ul className="list-inside list-disc space-y-1 text-gray-400">
                <li>To provide and maintain the OffGridFlow platform</li>
                <li>To calculate emissions and generate compliance reports</li>
                <li>To process payments and manage subscriptions</li>
                <li>To send service-related communications</li>
                <li>To improve our platform and develop new features</li>
              </ul>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">3. Data Protection</h2>
              <p>
                All data is encrypted in transit (TLS 1.2+) and at rest. We use multi-tenant
                isolation to ensure your emissions data is never accessible to other organizations.
                Access is controlled through role-based permissions and two-factor authentication.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">4. Data Sharing</h2>
              <p>
                We do not sell your personal information or emissions data. We share data only with:
              </p>
              <ul className="mt-2 list-inside list-disc space-y-1 text-gray-400">
                <li>Payment processors (Stripe) to handle billing</li>
                <li>Cloud infrastructure providers to host the platform</li>
                <li>As required by law or valid legal process</li>
              </ul>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">5. GDPR Rights (EU Users)</h2>
              <p>
                If you are in the European Economic Area, you have the right to access, correct,
                delete, or export your personal data. You may also object to processing or request
                restriction. Contact us at{' '}
                <a href="mailto:contact@off-grid-flow.com" className="text-primary-400 hover:underline">
                  contact@off-grid-flow.com
                </a>{' '}
                to exercise these rights.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">6. CCPA Rights (California Users)</h2>
              <p>
                California residents have the right to know what personal information we collect,
                request deletion, and opt out of the sale of personal information. We do not sell
                personal information. Contact us to exercise your rights.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">7. Data Retention</h2>
              <p>
                We retain your account data for as long as your account is active. Emissions data
                and reports are retained for the duration of your subscription plus 90 days.
                You may request deletion at any time.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">8. Contact</h2>
              <p>
                For privacy-related questions:{' '}
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
