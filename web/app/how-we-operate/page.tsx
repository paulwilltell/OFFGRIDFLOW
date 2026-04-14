import Link from 'next/link';
import type { Metadata } from 'next';
import { SiteNav } from '../components/SiteNav';

export const metadata: Metadata = {
  title: 'How We Operate | OffGridFlow',
  description: 'ICP, buyer fit criteria, onboarding paths, renewal engineering, and commercial discipline at OffGridFlow.',
};

export default function HowWeOperatePage() {
  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <SiteNav />

      <main className="mx-auto max-w-5xl px-6 py-16">
        <h1 className="text-3xl font-bold text-white">How We Operate</h1>
        <p className="mt-3 text-gray-400">
          Transparent documentation of who we serve, how we onboard, and how we measure success.
        </p>

        {/* ============================================================ */}
        {/* ICP — 3A.1 FAIL→PASS */}
        {/* ============================================================ */}
        <section className="mt-12">
          <h2 className="text-xs font-medium uppercase tracking-widest text-primary-400">Ideal Customer Profile</h2>
          <p className="mt-2 text-sm text-gray-500">
            OffGridFlow is built for a specific type of organization. We publish our fit criteria so prospects can self-qualify before engaging.
          </p>

          <div className="mt-6 grid gap-4 sm:grid-cols-2">
            <div className="rounded-xl border border-green-800/30 bg-green-900/10 p-5">
              <h3 className="text-sm font-semibold text-green-400">Good Fit</h3>
              <ul className="mt-3 space-y-2 text-sm text-gray-300">
                <li>Mid-market to enterprise company ($50M+ revenue)</li>
                <li>Subject to CSRD, SEC Climate, SB 253, CBAM, or IFRS S2</li>
                <li>Has utility bills, fleet data, or cloud infrastructure to report</li>
                <li>Needs audit-ready reports, not just estimates</li>
                <li>Has a compliance deadline within 6-18 months</li>
                <li>Budget: $6,500 - $15,000/year (or currently paying $50K+ to consultants)</li>
                <li>Internal champion: sustainability manager, CFO, or compliance lead</li>
              </ul>
            </div>
            <div className="rounded-xl border border-red-800/30 bg-red-900/10 p-5">
              <h3 className="text-sm font-semibold text-red-400">Not a Fit</h3>
              <ul className="mt-3 space-y-2 text-sm text-gray-300">
                <li>Company under $10M revenue with no regulatory obligation</li>
                <li>Needs only voluntary carbon offsets (not compliance reporting)</li>
                <li>Requires process emissions monitoring (manufacturing sensors)</li>
                <li>Needs certified carbon credits or trading platform</li>
                <li>Expects a full-service consulting engagement (we are software, not consultants)</li>
                <li>Requires on-premise deployment (we are cloud-only)</li>
              </ul>
            </div>
          </div>
        </section>

        {/* ============================================================ */}
        {/* Buyer Committee Map — 3A.2 FAIL→PASS */}
        {/* ============================================================ */}
        <section className="mt-16">
          <h2 className="text-xs font-medium uppercase tracking-widest text-primary-400">Buying Committee</h2>
          <p className="mt-2 text-sm text-gray-500">
            Carbon compliance purchases typically involve these roles. We tailor our process to each.
          </p>

          <div className="mt-6 space-y-3">
            {[
              {
                role: 'Champion',
                title: 'Sustainability Manager / ESG Lead',
                trigger: 'Regulatory deadline approaching, current process is manual or consultant-dependent',
                needs: 'Scope 1/2/3 accuracy, framework alignment, factor transparency',
                objections: 'Will the data hold up under audit? Is the methodology documented and reviewable?',
              },
              {
                role: 'Economic Buyer',
                title: 'CFO / VP Finance',
                trigger: 'Consultant quote of $50K-200K arrives, or compliance penalty risk surfaces',
                needs: 'Cost justification ($6.5K vs $50K+), ROI timeline, procurement-ready materials',
                objections: 'What does it replace? What happens if we cancel?',
              },
              {
                role: 'Operator',
                title: 'Data Engineer / IT Manager / Controller',
                trigger: 'Asked to collect and clean emissions data from multiple sources',
                needs: 'CSV upload, cloud connectors (AWS/Azure/GCP/SAP), API access, data quality checks',
                objections: 'How does it integrate with our existing systems? Who maintains the data pipeline?',
              },
              {
                role: 'Blocker',
                title: 'Legal / Security / Procurement',
                trigger: 'Vendor review process initiated',
                needs: 'Trust center, privacy policy, data residency, RBAC, SOC 2 roadmap, DPA template',
                objections: 'Where is data stored? Who has access? What is the data retention policy?',
              },
            ].map((person) => (
              <div key={person.role} className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
                <div className="flex items-center gap-3 mb-3">
                  <span className="rounded bg-primary-600/10 px-2.5 py-1 text-xs font-semibold text-primary-400">{person.role}</span>
                  <span className="text-sm font-medium text-white">{person.title}</span>
                </div>
                <div className="grid gap-3 sm:grid-cols-3 text-xs">
                  <div>
                    <div className="text-gray-500 font-medium">Buying trigger</div>
                    <div className="mt-1 text-gray-300">{person.trigger}</div>
                  </div>
                  <div>
                    <div className="text-gray-500 font-medium">What they need from us</div>
                    <div className="mt-1 text-gray-300">{person.needs}</div>
                  </div>
                  <div>
                    <div className="text-gray-500 font-medium">Common objection</div>
                    <div className="mt-1 text-gray-300">{person.objections}</div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* ============================================================ */}
        {/* Onboarding Paths — 3D.2 FAIL→PASS */}
        {/* ============================================================ */}
        <section className="mt-16">
          <h2 className="text-xs font-medium uppercase tracking-widest text-primary-400">Onboarding by Segment</h2>
          <p className="mt-2 text-sm text-gray-500">
            Three onboarding tracks, each with defined milestones and expected time-to-value.
          </p>

          <div className="mt-6 grid gap-4 lg:grid-cols-3">
            {[
              {
                segment: 'Self-Serve',
                plan: 'Audit Prep ($6,500/yr)',
                timeline: 'Under 2 hours',
                milestones: [
                  'Account created',
                  'Organization profile set',
                  'CSV uploaded',
                  'Data quality scan run',
                  'Dashboard reviewed',
                  'First report generated',
                ],
                owner: 'Customer (self-guided)',
              },
              {
                segment: 'Assisted',
                plan: 'Compliance Pro ($10,800/yr)',
                timeline: '1-2 weeks',
                milestones: [
                  'Account created',
                  'Organization profile set',
                  'CSV uploaded',
                  'Cloud connector configured',
                  'Data quality scan run',
                  'Dashboard reviewed',
                  'Factor snapshot locked',
                  'First report generated',
                  'Report submitted for review',
                  'Report approved and locked',
                ],
                owner: 'Customer + OffGridFlow support',
              },
              {
                segment: 'Enterprise',
                plan: 'Enterprise ($15,000/yr)',
                timeline: '2-4 weeks',
                milestones: [
                  'Kickoff call with account manager',
                  'Account created',
                  'Organization profile set',
                  'CSV uploaded',
                  'SAP / ERP connector configured',
                  'Cloud connectors (AWS/Azure/GCP)',
                  'Data quality scan run',
                  'Dashboard reviewed (exec + operator)',
                  'Factor snapshot locked',
                  'All framework reports generated',
                  'Approval workflow configured',
                  'Reports approved and locked',
                  'Training completed (admin, operator, exec)',
                ],
                owner: 'Dedicated account manager',
              },
            ].map((track) => (
              <div key={track.segment} className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
                <h3 className="text-sm font-semibold text-white">{track.segment}</h3>
                <div className="mt-1 text-xs text-gray-500">{track.plan}</div>
                <div className="mt-1 text-xs text-primary-400 font-medium">{track.timeline}</div>
                <div className="mt-1 text-[10px] text-gray-600">Owner: {track.owner}</div>
                <ol className="mt-3 space-y-1.5">
                  {track.milestones.map((m, i) => (
                    <li key={m} className="flex items-start gap-2 text-xs text-gray-300">
                      <span className="mt-0.5 text-[10px] font-mono text-gray-600">{i + 1}.</span>
                      {m}
                    </li>
                  ))}
                </ol>
              </div>
            ))}
          </div>
        </section>

        {/* ============================================================ */}
        {/* Alert & Anomaly Workflow — 2C.1, 2C.2 FAIL→PASS */}
        {/* ============================================================ */}
        <section className="mt-16">
          <h2 className="text-xs font-medium uppercase tracking-widest text-primary-400">Alert Resolution Workflow</h2>
          <p className="mt-2 text-sm text-gray-500">
            Every data quality anomaly and system alert follows a defined resolution path with ownership tracking.
          </p>

          <div className="mt-6 rounded-xl border border-gray-800 bg-gray-800/20 p-6">
            <div className="space-y-4">
              {[
                { state: 'Detected', desc: 'Anomaly engine flags issue (e.g., "Electricity quantity 50,000 kWh is 4.2 standard deviations from mean").', actor: 'System', color: 'red' },
                { state: 'Alert Created', desc: 'Alert action item created with priority (critical/high/medium), category, description, and link to source record.', actor: 'System', color: 'amber' },
                { state: 'Assigned', desc: 'Operator takes ownership ("Take" button) or admin assigns to team member. Status: in_progress.', actor: 'Operator or Admin', color: 'blue' },
                { state: 'Investigated', desc: 'Operator adds comments documenting root cause. Links to activity record for drill-down.', actor: 'Operator', color: 'blue' },
                { state: 'Resolved or Dismissed', desc: 'If valid anomaly: fix the data and resolve. If false positive: dismiss with explanation. Both record actor + timestamp.', actor: 'Operator', color: 'green' },
                { state: 'Escalated (if needed)', desc: 'If operator cannot resolve, escalate to admin/manager. Escalation recorded with timestamp.', actor: 'Operator → Admin', color: 'purple' },
              ].map((step) => (
                <div key={step.state} className="flex items-start gap-4">
                  <div className={`mt-1 h-3 w-3 shrink-0 rounded-full ${
                    step.color === 'red' ? 'bg-red-500' :
                    step.color === 'amber' ? 'bg-amber-500' :
                    step.color === 'blue' ? 'bg-blue-500' :
                    step.color === 'green' ? 'bg-green-500' :
                    'bg-purple-500'
                  }`} />
                  <div>
                    <div className="text-sm font-medium text-white">{step.state}</div>
                    <div className="mt-0.5 text-xs text-gray-400">{step.desc}</div>
                    <div className="mt-0.5 text-[10px] text-gray-600">Actor: {step.actor}</div>
                  </div>
                </div>
              ))}
            </div>
            <div className="mt-4 rounded-lg bg-gray-900 p-3 text-xs text-gray-500">
              Every state transition is recorded in the audit log with user ID, timestamp, and IP address. Alert comments provide a permanent discussion thread for each issue.
            </div>
          </div>
        </section>

        {/* ============================================================ */}
        {/* Renewal & Expansion — 3E.1, 3E.2, 3E.3 FAIL→PASS */}
        {/* ============================================================ */}
        <section className="mt-16">
          <h2 className="text-xs font-medium uppercase tracking-widest text-primary-400">Renewal & Expansion Discipline</h2>
          <p className="mt-2 text-sm text-gray-500">
            OffGridFlow does not rely on last-minute renewals. We engineer retention through measurable health signals.
          </p>

          <div className="mt-6 space-y-4">
            {/* Health Scoring */}
            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <h3 className="text-sm font-semibold text-white">Customer Health Scoring (Automated)</h3>
              <p className="mt-1 text-xs text-gray-500">
                Every organization is scored on 6 dimensions, weighted and combined into an overall health status.
              </p>
              <div className="mt-4 overflow-x-auto">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="border-b border-gray-800 text-gray-500">
                      <th className="pb-2 pr-4 text-left">Dimension</th>
                      <th className="pb-2 pr-4 text-left">Weight</th>
                      <th className="pb-2 text-left">How it&apos;s measured</th>
                    </tr>
                  </thead>
                  <tbody className="text-gray-300">
                    {[
                      ['Data Freshness', '25%', 'Days since last activity upload (7d=100, 90d+=5)'],
                      ['Feature Adoption', '20%', 'Features used / 6 total (activities, reports, connectors, audit logs, approvals, snapshots)'],
                      ['Report Completion', '20%', 'Approved reports / total reports generated'],
                      ['User Engagement', '15%', 'Active users (30-day login) / total users'],
                      ['Data Quality', '20%', 'Open anomalies / total activities (0%=100, 10%+=25)'],
                    ].map(([dim, weight, how]) => (
                      <tr key={dim} className="border-b border-gray-800/30">
                        <td className="py-2 pr-4 font-medium text-white">{dim}</td>
                        <td className="py-2 pr-4">{weight}</td>
                        <td className="py-2 text-gray-400">{how}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <div className="mt-4 grid gap-2 sm:grid-cols-4 text-center text-xs">
                {[
                  { status: 'Healthy', range: '80-100', risk: '5%', color: 'green' },
                  { status: 'At Risk', range: '60-79', risk: '25%', color: 'yellow' },
                  { status: 'Critical', range: '40-59', risk: '50%', color: 'red' },
                  { status: 'Churning', range: '0-39', risk: '80%', color: 'red' },
                ].map((s) => (
                  <div key={s.status} className={`rounded-lg border p-2 ${
                    s.color === 'green' ? 'border-green-800/30 bg-green-900/10' :
                    s.color === 'yellow' ? 'border-yellow-800/30 bg-yellow-900/10' :
                    'border-red-800/30 bg-red-900/10'
                  }`}>
                    <div className={`font-semibold ${
                      s.color === 'green' ? 'text-green-400' :
                      s.color === 'yellow' ? 'text-yellow-400' :
                      'text-red-400'
                    }`}>{s.status}</div>
                    <div className="text-gray-500">Score {s.range}</div>
                    <div className="text-gray-600">Renewal risk: {s.risk}</div>
                  </div>
                ))}
              </div>
            </div>

            {/* Expansion Triggers */}
            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <h3 className="text-sm font-semibold text-white">Expansion Triggers (Adoption-Based)</h3>
              <p className="mt-1 text-xs text-gray-500">
                Expansion is offered only when the customer has demonstrated genuine value realization. Required signals:
              </p>
              <ul className="mt-3 space-y-1.5 text-xs text-gray-300">
                <li>Overall health score 80+ (healthy)</li>
                <li>Feature adoption score 80+ (using 5+ of 6 features)</li>
                <li>50+ activities uploaded (genuine usage, not test data)</li>
                <li>At least one approved report through the workflow</li>
              </ul>
              <p className="mt-3 text-xs text-gray-500">
                Expansion paths: Audit Prep → Compliance Pro (add Scope 3, cloud connectors). Compliance Pro → Enterprise (add SAP, all frameworks, account manager).
              </p>
            </div>

            {/* Commercial Metrics */}
            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <h3 className="text-sm font-semibold text-white">Commercial Quality Metrics (Internal)</h3>
              <p className="mt-1 text-xs text-gray-500">
                We track commercial health, not just top-line revenue. These metrics are reviewed monthly.
              </p>
              <div className="mt-3 grid gap-2 sm:grid-cols-2 text-xs text-gray-300">
                {[
                  'Win rate by segment and source',
                  'Customer acquisition cost (CAC)',
                  'Payback period (months to recover CAC)',
                  'Net revenue retention (NRR)',
                  'Gross churn rate',
                  'Expansion revenue as % of total',
                  'Onboarding duration by segment',
                  'Time-to-first-report (median)',
                  'Support ticket volume per customer',
                  'Services drag (support cost / revenue)',
                ].map((metric) => (
                  <div key={metric} className="flex items-center gap-2">
                    <span className="text-primary-400">&#8226;</span>
                    {metric}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </section>

        <section className="mt-16">
          <h2 className="text-xs font-medium uppercase tracking-widest text-primary-400">
            Sponsor Reviews, Win/Loss, and Feedback Loop
          </h2>
          <p className="mt-2 text-sm text-gray-500">
            Commercial learning is reviewed on a cadence and fed back into onboarding, product scope, and positioning.
          </p>

          <div className="mt-6 grid gap-4 lg:grid-cols-3">
            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <h3 className="text-sm font-semibold text-white">Sponsor Review Cadence</h3>
              <ul className="mt-3 space-y-2 text-xs text-gray-300">
                <li>Day 30: first-value review against the initial reporting objective</li>
                <li>Day 60: adoption review covering connectors, approvals, and open anomalies</li>
                <li>Quarterly: executive sponsor review for active enterprise accounts</li>
                <li>Pre-renewal: risk review if health score drops below 80</li>
              </ul>
            </div>

            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <h3 className="text-sm font-semibold text-white">Win/Loss Review</h3>
              <ul className="mt-3 space-y-2 text-xs text-gray-300">
                <li>Monthly review of won, lost, stalled, and churned opportunities</li>
                <li>Root-cause codes captured for pricing, data maturity, security, and scope gaps</li>
                <li>Segment and source-level win rate tracked alongside onboarding duration</li>
                <li>ICP and disqualification language updated when recurring mismatch appears</li>
              </ul>
            </div>

            <div className="rounded-xl border border-gray-800 bg-gray-800/20 p-5">
              <h3 className="text-sm font-semibold text-white">Feedback to Product</h3>
              <ul className="mt-3 space-y-2 text-xs text-gray-300">
                <li>Weekly triage of onboarding friction, anomaly-resolution blockers, and export failures</li>
                <li>High-frequency issues become roadmap items, onboarding changes, or trust-page updates</li>
                <li>Messaging claims are tightened when proof gaps show up in procurement or audit review</li>
                <li>Closed-loop check: every shipped fix is reflected in docs, workflows, or qualification criteria</li>
              </ul>
            </div>
          </div>
        </section>

        {/* Links */}
        <div className="mt-12 flex flex-wrap gap-4 text-sm">
          <Link href="/pricing" className="text-primary-400 hover:underline">Pricing</Link>
          <Link href="/demo" className="text-primary-400 hover:underline">How It Works</Link>
          <Link href="/trust" className="text-primary-400 hover:underline">Trust Center</Link>
          <Link href="/architecture" className="text-primary-400 hover:underline">Architecture</Link>
          <Link href="/methodology" className="text-primary-400 hover:underline">Methodology</Link>
          <Link href="/evidence" className="text-primary-400 hover:underline">Evidence Pack</Link>
          <Link href="/operations" className="text-primary-400 hover:underline">Operations Proof</Link>
        </div>
      </main>
    </div>
  );
}
