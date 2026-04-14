import Link from 'next/link';
import type { Metadata } from 'next';
import { SiteNav } from '../components/SiteNav';

export const metadata: Metadata = {
  title: 'Case Study: Full Carbon Footprint in Under 2 Hours | OffGridFlow',
  description: 'How a small technology company calculated Scope 2 emissions across two data centers and office space using OffGridFlow — without hiring a consultant.',
};

export default function CaseStudyPage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <SiteNav />

      <article className="px-6 py-20">
        <div className="mx-auto max-w-3xl">
          <div className="mb-6 text-xs font-medium uppercase tracking-widest text-primary-400">Case Study</div>

          <h1 className="mb-6 text-3xl font-bold text-white sm:text-4xl">
            How a SaaS company calculated its full carbon footprint in under 2 hours
          </h1>

          <div className="mb-10 grid gap-6 rounded-xl border border-gray-800 bg-dark-800 p-6 sm:grid-cols-4">
            <div>
              <div className="text-2xl font-bold text-primary-400">86%</div>
              <div className="mt-0.5 text-xs text-gray-500">Cost savings vs consultant</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-primary-400">&lt;2 hrs</div>
              <div className="mt-0.5 text-xs text-gray-500">Time to first report</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-primary-400">24</div>
              <div className="mt-0.5 text-xs text-gray-500">Monthly data points</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-primary-400">2</div>
              <div className="mt-0.5 text-xs text-gray-500">Emission regions tracked</div>
            </div>
          </div>

          <div className="space-y-8 text-gray-300 leading-relaxed">
            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">The Challenge</h2>
              <p>
                A small technology company operating in California needed to understand its carbon
                footprint for the first time. With office space in the WECC California grid region
                and cloud infrastructure running on AWS in the Pacific Northwest, the company had
                two primary sources of Scope 2 emissions but no way to calculate them.
              </p>
              <p className="mt-3">
                A Big 4 consulting firm quoted <strong className="text-white">$35,000</strong> for
                a Scope 1 &amp; 2 assessment. For a company of this size, that was the entire
                sustainability budget for the year.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">The Solution</h2>
              <p>
                Using OffGridFlow, the sustainability lead uploaded 12 months of utility data
                (24 data points across 2 meters) via CSV in under 5 minutes. No data transformation,
                no manual factor lookups, no spreadsheets.
              </p>
              <p className="mt-3">
                OffGridFlow automatically applied the correct EPA eGRID emission factors for each
                region:
              </p>
              <ul className="mt-3 list-inside list-disc space-y-1 text-gray-400">
                <li>
                  <strong className="text-gray-300">Office (WECC California / CAMX):</strong> 0.225
                  kg CO2e/kWh &mdash; relatively clean grid due to California&apos;s renewable mandates
                </li>
                <li>
                  <strong className="text-gray-300">Cloud / AWS (WECC Northwest / NWPP):</strong> 0.252
                  kg CO2e/kWh &mdash; hydropower-heavy Pacific Northwest grid
                </li>
              </ul>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">The Results</h2>
              <div className="rounded-lg border border-gray-800 bg-gray-900/50 p-5">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-gray-800 text-left text-xs text-gray-500">
                      <th className="pb-2">Source</th>
                      <th className="pb-2">Annual kWh</th>
                      <th className="pb-2">Region Factor</th>
                      <th className="pb-2">tCO2e</th>
                    </tr>
                  </thead>
                  <tbody className="text-gray-300">
                    <tr className="border-b border-gray-800/50">
                      <td className="py-2">HQ Office</td>
                      <td>36,556</td>
                      <td>0.225</td>
                      <td>8.22</td>
                    </tr>
                    <tr className="border-b border-gray-800/50">
                      <td className="py-2">Cloud (AWS)</td>
                      <td>52,155</td>
                      <td>0.252</td>
                      <td>13.14</td>
                    </tr>
                    <tr className="font-semibold text-white">
                      <td className="py-2">Total Scope 2</td>
                      <td>88,711</td>
                      <td></td>
                      <td>21.36</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p className="mt-4">
                Total annual Scope 2 emissions: <strong className="text-white">21.36 tonnes CO2e</strong>,
                with cloud infrastructure accounting for 61% of the footprint. The PDF compliance
                report was generated in one click, ready for stakeholder review.
              </p>
            </section>

            <section>
              <h2 className="mb-3 text-xl font-semibold text-white">The Impact</h2>
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="rounded-lg border border-gray-800 bg-dark-800 p-4">
                  <div className="text-sm font-medium text-white">Cost</div>
                  <div className="mt-1 text-xs text-gray-400">
                    $6,500/year with OffGridFlow vs $35,000 quote from Big 4.
                    <strong className="text-primary-400"> 86% savings.</strong>
                  </div>
                </div>
                <div className="rounded-lg border border-gray-800 bg-dark-800 p-4">
                  <div className="text-sm font-medium text-white">Time</div>
                  <div className="mt-1 text-xs text-gray-400">
                    Under 2 hours from account creation to final report.
                    Consultant engagement would have taken 4&ndash;6 weeks.
                  </div>
                </div>
                <div className="rounded-lg border border-gray-800 bg-dark-800 p-4">
                  <div className="text-sm font-medium text-white">Accuracy</div>
                  <div className="mt-1 text-xs text-gray-400">
                    EPA eGRID subregion-level factors applied automatically.
                    No manual factor lookups or formula errors.
                  </div>
                </div>
                <div className="rounded-lg border border-gray-800 bg-dark-800 p-4">
                  <div className="text-sm font-medium text-white">Insight</div>
                  <div className="mt-1 text-xs text-gray-400">
                    Discovered cloud infrastructure is the primary emission source (61%).
                    Informed decision to evaluate carbon-efficient cloud regions.
                  </div>
                </div>
              </div>
            </section>

            <section className="rounded-xl border border-primary-600/20 bg-primary-600/5 p-6 text-center">
              <p className="text-lg font-semibold text-white">
                Want the same results for your company?
              </p>
              <p className="mt-2 text-sm text-gray-400">
                Upload your utility data and get your first report in under 2 hours. If you need the drill-down,
                review the redacted packet behind this case study first.
              </p>
              <div className="mt-5 flex justify-center gap-4">
                <Link
                  href="/evidence"
                  className="rounded-lg border border-gray-700 px-6 py-2.5 text-sm font-medium text-gray-300 transition hover:border-gray-500 hover:text-white"
                >
                  Review Evidence Pack
                </Link>
                <Link
                  href="/register?plan=starter"
                  className="rounded-lg bg-primary-600 px-6 py-2.5 text-sm font-semibold text-white transition hover:bg-primary-500"
                >
                  Start Free Trial
                </Link>
                <Link href="/demo" className="rounded-lg border border-gray-700 px-6 py-2.5 text-sm font-medium text-gray-300 transition hover:border-gray-500 hover:text-white">
                  Review Workflow
                </Link>
              </div>
            </section>
          </div>
        </div>
      </article>
    </div>
  );
}
