import Link from 'next/link';
import type { Metadata } from 'next';
import { SiteNav } from '../components/SiteNav';

export const metadata: Metadata = {
  title: 'Terms of Service | OffGridFlow',
  description: 'OffGridFlow terms of service governing use of our carbon accounting platform.',
};

export default function TermsPage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <SiteNav />

      <section className="px-6 py-24">
        <div className="mx-auto max-w-3xl">
          <h1 className="mb-2 text-4xl font-bold text-white">Terms of Service</h1>
          <p className="mb-4 text-sm text-gray-500">Last updated: April 2026</p>

          <div className="mb-10 rounded-xl border border-amber-500/30 bg-amber-500/5 p-4 text-sm text-amber-200">
            <strong className="block mb-1">Draft &mdash; Pending Attorney Review.</strong>
            These Terms of Service are a working draft intended to describe how the Platform is
            currently operated. They do not constitute legal advice and have not yet been finalized
            by qualified counsel. Customers with specific legal requirements should request a
            negotiated Master Subscription Agreement (MSA) from{' '}
            <a href="mailto:contact@off-grid-flow.com" className="underline">contact@off-grid-flow.com</a>.
          </div>

          <div className="space-y-8 text-gray-300 leading-relaxed">
            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">1. Acceptance of Terms</h2>
              <p>
                By accessing or using OffGridFlow (&quot;the Platform,&quot; &quot;we,&quot; or
                &quot;OffGridFlow&quot;), you (&quot;Customer&quot; or &quot;you&quot;) agree to be
                bound by these Terms of Service and the{' '}
                <Link href="/privacy" className="text-primary-400 hover:underline">Privacy Policy</Link>.
                If you do not agree to any part of these terms, you must not use the Platform.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">2. Description of Service</h2>
              <p>
                OffGridFlow provides carbon accounting software that calculates greenhouse gas
                emissions (Scope 1, 2, and 3) and generates draft compliance reports aligned with
                regulatory frameworks including CSRD / ESRS E1, SEC Climate Disclosure, California
                SB 253, EU CBAM, and IFRS S2. The Platform applies the methodology documented in
                our public{' '}
                <span className="font-medium text-gray-200">Methodology Library</span>,
                which is versioned and timestamped.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">3. Account Registration</h2>
              <p>
                You must provide accurate, complete information when creating an account. You are
                responsible for maintaining the confidentiality of your credentials and for all
                activity under your account. You must notify us immediately at
                contact@off-grid-flow.com of any unauthorized access. You must be at least 18 years
                old and authorized to bind your organization.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">4. Subscriptions, Billing, and Refunds</h2>
              <p>
                Paid subscriptions are billed annually in advance via Stripe. Prices are as
                published on our{' '}
                <Link href="/pricing" className="text-primary-400 hover:underline">pricing page</Link>{' '}
                at the time of purchase. By subscribing you authorize OffGridFlow to charge the
                payment method on file for the subscription term.
              </p>
              <p className="mt-3">
                <strong className="text-white">Refund policy.</strong> Annual subscriptions include
                a 14-day refund window from the date of initial purchase. After 14 days, annual
                subscriptions are non-refundable. Renewal charges are non-refundable. You may
                cancel at any time; access continues through the end of the current billing period.
                Fees for partial periods are not prorated.
              </p>
              <p className="mt-3">
                <strong className="text-white">Chargebacks.</strong> You agree to contact us at
                contact@off-grid-flow.com before initiating a chargeback. Chargebacks initiated
                without first contacting support may result in immediate account suspension and
                recovery of the charged amount plus processing fees.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">5. Your Data and License Grant</h2>
              <p>
                You retain all ownership of emissions data, reports, and other content you upload
                (&quot;Customer Data&quot;). You grant OffGridFlow a limited, non-exclusive,
                worldwide license to host, process, and display Customer Data solely to provide
                the Platform. OffGridFlow will not sell, rent, or use Customer Data for advertising,
                training third-party models, or any purpose beyond delivering the contracted services.
              </p>
              <p className="mt-3">
                You represent that you have all necessary rights to upload Customer Data and that
                such data does not infringe any third-party rights or violate applicable law.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">6. Acceptable Use</h2>
              <p>
                You agree not to: (a) reverse engineer, decompile, or attempt to extract source code;
                (b) interfere with or disrupt the Platform or other users; (c) introduce malware,
                viruses, or harmful code; (d) use the Platform to violate any law or regulation;
                (e) scrape, harvest, or bulk-download data outside published export APIs; (f)
                resell, sublicense, or provide access to the Platform to third parties without
                written permission.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">7. Accuracy Disclaimer &mdash; Not Legal or Audit Advice</h2>
              <p>
                OffGridFlow applies the calculation methods documented in its public{' '}
                <span className="font-medium text-gray-200">Methodology Library</span>
                {' '}and verified emission factors from published sources (EPA, IEA, DEFRA, IPCC, GHG
                Protocol). The accuracy of outputs depends on the accuracy and completeness of the
                input data you provide.
              </p>
              <p className="mt-3">
                <strong className="text-white">
                  OffGridFlow is a calculation and reporting tool only. It is not a substitute for
                  professional audit, legal, accounting, or regulatory advice.
                </strong>{' '}
                Reports generated by the Platform are drafts. You are solely responsible for
                reviewing, verifying, and independently validating all outputs before submission
                to any regulatory body, auditor, board, investor, or other stakeholder. OffGridFlow
                does not guarantee that any report will be accepted by any regulatory body or
                third party.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">8. Warranty Disclaimer &mdash; Service Provided &quot;AS IS&quot;</h2>
              <p className="uppercase text-sm">
                The Platform is provided on an &quot;AS IS&quot; and &quot;AS AVAILABLE&quot; basis.
                To the maximum extent permitted by law, OffGridFlow disclaims all warranties,
                express or implied, including merchantability, fitness for a particular purpose,
                non-infringement, accuracy, completeness, and title.
              </p>
              <p className="mt-3">
                OffGridFlow does not warrant that the Platform will be uninterrupted, error-free,
                secure against all threats, or that defects will be corrected. No oral or written
                statement obtained from OffGridFlow creates any warranty beyond those in these Terms.
              </p>
              <p className="mt-3">
                <strong className="text-white">No uptime guarantee.</strong> OffGridFlow publishes a{' '}
                <span className="font-medium text-gray-200">live system status page</span>{' '}
                and strives for high availability, but we do not guarantee specific uptime percentages
                unless expressly agreed in a signed Service Level Agreement (SLA).
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">9. Limitation of Liability</h2>
              <p>
                To the maximum extent permitted by applicable law, OffGridFlow&apos;s total
                aggregate liability arising from or related to these Terms or the Platform shall
                not exceed the greater of (a) the fees actually paid by Customer to OffGridFlow in
                the twelve (12) months immediately preceding the event giving rise to the claim,
                or (b) one hundred U.S. dollars ($100).
              </p>
              <p className="mt-3 uppercase text-sm">
                In no event shall OffGridFlow be liable for any indirect, incidental, special,
                consequential, punitive, or exemplary damages, or for any lost profits, lost
                revenue, lost data, business interruption, or regulatory fines, whether based in
                contract, tort, strict liability, or otherwise, even if advised of the possibility
                of such damages.
              </p>
              <p className="mt-3">
                These limitations apply regardless of the failure of essential purpose of any
                limited remedy. Some jurisdictions do not allow certain limitations; in those
                jurisdictions, liability is limited to the smallest amount permitted by law.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">10. Indemnification</h2>
              <p>
                You agree to defend, indemnify, and hold harmless OffGridFlow, its officers,
                employees, contractors, and affiliates from any third-party claims, damages,
                liabilities, costs, or expenses (including reasonable attorneys&apos; fees) arising
                from (a) your breach of these Terms; (b) your use of the Platform in violation of
                law; (c) the content, accuracy, or legality of Customer Data; or (d) your
                submission of Platform outputs to any regulatory body or third party.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">11. Third-Party Subprocessors</h2>
              <p>
                OffGridFlow uses the subprocessors listed in the{' '}
                <Link href="/privacy" className="text-primary-400 hover:underline">Privacy Policy</Link>{' '}
                to operate the Platform, including Stripe (payments), Railway (hosting),
                SendGrid (email), and Google (analytics and ads). These providers process data
                under their own terms and agreements; OffGridFlow contracts with each to maintain
                appropriate security and confidentiality.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">12. Service Modifications</h2>
              <p>
                OffGridFlow may update, improve, or modify the Platform from time to time, including
                adding or removing features. Material changes that adversely affect paid customers
                will be communicated with reasonable notice via email or in-app notification.
                If OffGridFlow discontinues a paid feature during a subscription term, you may
                request a prorated refund of fees attributable to the discontinued feature.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">13. Termination</h2>
              <p>
                Either party may terminate the agreement at any time. OffGridFlow may suspend or
                terminate accounts that violate these Terms, expose the Platform to legal risk, or
                fail to pay. Upon termination, you may export Customer Data for thirty (30) days
                via the governance export endpoint; after that period, data will be permanently
                deleted in accordance with the retention policy published in the{' '}
                <Link href="/trust" className="text-primary-400 hover:underline">Trust Center</Link>.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">14. Force Majeure</h2>
              <p>
                Neither party is liable for failure or delay in performance caused by events beyond
                reasonable control, including natural disasters, acts of war or terrorism, labor
                disputes, internet or telecommunications outages, government actions, or epidemics.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">15. Governing Law and Dispute Resolution</h2>
              <p>
                These Terms are governed by the laws of the State of California, excluding its
                conflict-of-laws rules. The parties agree to attempt in good faith to resolve any
                dispute informally for thirty (30) days before initiating formal proceedings. Any
                unresolved dispute shall be brought exclusively in the state or federal courts
                located in Sacramento County, California, and both parties consent to personal
                jurisdiction there.
              </p>
              <p className="mt-3 italic text-gray-400">
                <strong>Draft note:</strong> An arbitration clause and class-action waiver may be
                added subject to attorney review. Customers acquiring the Platform through an
                executed MSA may negotiate governing law and venue.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">16. Changes to These Terms</h2>
              <p>
                OffGridFlow may update these Terms from time to time. We will post the updated
                Terms with a revised &quot;Last updated&quot; date. For material changes, we will
                provide notice via email or in-app notification at least thirty (30) days before
                the changes take effect. Continued use of the Platform after the effective date
                constitutes acceptance.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">17. Severability and Entire Agreement</h2>
              <p>
                If any provision of these Terms is held unenforceable, the remaining provisions
                remain in full force and effect, and the unenforceable provision shall be modified
                to the minimum extent necessary to make it enforceable. These Terms, together with
                the Privacy Policy and any executed order form or MSA, constitute the entire
                agreement between the parties regarding the Platform.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">18. Contact</h2>
              <p>
                Questions about these Terms:{' '}
                <a href="mailto:contact@off-grid-flow.com" className="text-primary-400 hover:underline">
                  contact@off-grid-flow.com
                </a>
              </p>
            </section>
          </div>
        </div>
      </section>
    </div>
  );
}
