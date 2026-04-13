import Link from 'next/link';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Terms of Service | OffGridFlow',
  description: 'OffGridFlow terms of service governing use of our carbon accounting platform.',
};

export default function TermsPage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <nav className="border-b border-gray-800/50 bg-dark-900/80 backdrop-blur-md">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <Link href="/" className="text-xl font-bold text-white">OffGridFlow</Link>
          <Link href="/" className="text-sm text-gray-400 hover:text-white">Back to Home</Link>
        </div>
      </nav>

      <section className="px-6 py-24">
        <div className="mx-auto max-w-3xl">
          <h1 className="mb-2 text-4xl font-bold text-white">Terms of Service</h1>
          <p className="mb-12 text-sm text-gray-500">Last updated: April 2026</p>

          <div className="space-y-8 text-gray-300 leading-relaxed">
            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">1. Acceptance of Terms</h2>
              <p>
                By accessing or using OffGridFlow (&quot;the Platform&quot;), you agree to be bound
                by these Terms of Service. If you do not agree, do not use the Platform.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">2. Description of Service</h2>
              <p>
                OffGridFlow provides carbon accounting software that calculates greenhouse gas
                emissions (Scope 1, 2, and 3) and generates compliance reports for regulatory
                frameworks including CSRD, SEC Climate Rules, California SB 253, CBAM, and IFRS S2.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">3. Account Registration</h2>
              <p>
                You must provide accurate, complete information when creating an account. You are
                responsible for maintaining the confidentiality of your credentials and for all
                activity under your account. Notify us immediately of any unauthorized access.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">4. Subscriptions &amp; Billing</h2>
              <p>
                Paid subscriptions are billed annually. Prices are as published on our pricing
                page at the time of purchase. You may cancel at any time; access continues through
                the end of the current billing period. Refunds are handled on a case-by-case basis.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">5. Your Data</h2>
              <p>
                You retain ownership of all emissions data, reports, and other content you upload
                to the Platform. You grant OffGridFlow a limited license to process this data
                solely for the purpose of providing the services. We will not use your data for
                any other purpose without your consent.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">6. Accuracy Disclaimer</h2>
              <p>
                OffGridFlow uses GHG Protocol-compliant calculation methodologies and verified
                emission factors. However, the accuracy of outputs depends on the accuracy of
                input data you provide. OffGridFlow is a calculation and reporting tool, not a
                substitute for professional audit or legal advice. You are responsible for
                verifying reports before submission to regulatory bodies.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">7. Limitation of Liability</h2>
              <p>
                To the maximum extent permitted by law, OffGridFlow&apos;s total liability for
                any claim arising from use of the Platform shall not exceed the amount you paid
                for the service in the 12 months preceding the claim. We are not liable for
                indirect, incidental, or consequential damages.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">8. Termination</h2>
              <p>
                Either party may terminate the agreement at any time. Upon termination, you may
                export your data for 30 days. After that period, data will be permanently deleted.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">9. Governing Law</h2>
              <p>
                These terms are governed by the laws of the State of California. Any disputes
                shall be resolved in the courts of Sacramento County, California.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">10. Contact</h2>
              <p>
                Questions about these terms:{' '}
                <a href="mailto:paul@off-gridflow.com" className="text-primary-400 hover:underline">
                  paul@off-gridflow.com
                </a>
              </p>
            </section>
          </div>
        </div>
      </section>
    </div>
  );
}
