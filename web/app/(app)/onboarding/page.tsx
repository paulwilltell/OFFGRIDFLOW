'use client';

import Link from 'next/link';
import { useSession, useRequireAuth } from '@/lib/session';

const milestones = [
  {
    step: 1,
    title: 'Account Created',
    description: 'Your account is set up with your organization profile.',
    action: null,
    href: null,
    autoComplete: true,
  },
  {
    step: 2,
    title: 'First Data Upload',
    description: 'Upload a CSV with utility bills, energy consumption, or fleet data.',
    action: 'Upload CSV',
    href: '/emissions',
    autoComplete: false,
  },
  {
    step: 3,
    title: 'Connect a Cloud Source',
    description: 'Set up automated data pipelines from AWS, Azure, GCP, SAP, or utility providers.',
    action: 'Configure Sources',
    href: '/settings/data-sources',
    autoComplete: false,
  },
  {
    step: 4,
    title: 'Review Your Inventory',
    description: 'Check Scope 1 & 2 emissions on the dashboard. Verify data quality and factor accuracy.',
    action: 'Open Dashboard',
    href: '/dashboard/carbon',
    autoComplete: false,
  },
  {
    step: 5,
    title: 'Generate First Report',
    description: 'Generate an audit-ready compliance report (CSRD, SEC, SB 253, or CBAM).',
    action: 'Generate Report',
    href: '/compliance/csrd',
    autoComplete: false,
  },
  {
    step: 6,
    title: 'Submit for Review',
    description: 'Create an approval request and submit your inventory for internal review.',
    action: 'Start Approval',
    href: '/audit/approvals',
    autoComplete: false,
  },
  {
    step: 7,
    title: 'Approved & Locked',
    description: 'Report reviewed, approved, and locked. Ready for stakeholders or auditors.',
    action: 'View Approvals',
    href: '/audit/approvals',
    autoComplete: false,
  },
];

export default function OnboardingPage() {
  useRequireAuth();
  const { user } = useSession();

  // Step 1 is always complete if the user exists
  const completedSteps = user ? 1 : 0;

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-xl font-bold text-white">Getting Started</h1>
        <p className="mt-1 text-xs text-gray-500">
          Follow these steps to reach your first audit-ready report. Typical time: under 2 hours.
        </p>
      </div>

      {/* Progress bar */}
      <div className="mb-8">
        <div className="mb-2 flex items-center justify-between text-xs text-gray-500">
          <span>{completedSteps} of {milestones.length} complete</span>
          <span>{Math.round((completedSteps / milestones.length) * 100)}%</span>
        </div>
        <div className="h-2 overflow-hidden rounded-full bg-gray-800">
          <div
            className="h-full rounded-full bg-primary-600 transition-all duration-500"
            style={{ width: `${(completedSteps / milestones.length) * 100}%` }}
          />
        </div>
      </div>

      {/* Milestones */}
      <div className="space-y-3">
        {milestones.map((m, i) => {
          const isComplete = i < completedSteps;
          const isCurrent = i === completedSteps;
          const isFuture = i > completedSteps;

          return (
            <div
              key={m.step}
              className={`rounded-xl border p-5 transition ${
                isComplete ? 'border-green-500/20 bg-green-500/5' :
                isCurrent ? 'border-primary-600/30 bg-primary-600/5' :
                'border-gray-800 bg-gray-800/20'
              }`}
            >
              <div className="flex items-start gap-4">
                <div className={`flex h-8 w-8 items-center justify-center rounded-lg text-sm font-bold ${
                  isComplete ? 'bg-green-500/10 text-green-400' :
                  isCurrent ? 'bg-primary-600/10 text-primary-400' :
                  'bg-gray-800 text-gray-600'
                }`}>
                  {isComplete ? '✓' : m.step}
                </div>
                <div className="flex-1">
                  <div className="flex items-center justify-between">
                    <h3 className={`text-sm font-semibold ${
                      isComplete ? 'text-green-400' :
                      isCurrent ? 'text-white' :
                      'text-gray-500'
                    }`}>
                      {m.title}
                    </h3>
                    {isComplete && (
                      <span className="rounded bg-green-500/10 px-2 py-0.5 text-[10px] font-medium text-green-400">Complete</span>
                    )}
                    {isCurrent && (
                      <span className="rounded bg-primary-600/10 px-2 py-0.5 text-[10px] font-medium text-primary-400">Current</span>
                    )}
                  </div>
                  <p className={`mt-1 text-xs ${isFuture ? 'text-gray-600' : 'text-gray-400'}`}>
                    {m.description}
                  </p>
                  {m.action && m.href && (isCurrent || isComplete) && (
                    <Link
                      href={m.href}
                      className={`mt-3 inline-block rounded-lg px-4 py-1.5 text-xs font-medium transition ${
                        isCurrent ? 'bg-primary-600 text-white hover:bg-primary-500' :
                        'border border-gray-700 text-gray-400 hover:text-white'
                      }`}
                    >
                      {m.action}
                    </Link>
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Help section */}
      <div className="mt-8 rounded-lg border border-gray-800 bg-gray-900/30 p-4">
        <div className="flex items-start gap-3">
          <div className="text-lg">&#128172;</div>
          <div>
            <div className="text-sm font-medium text-white">Need implementation help?</div>
            <p className="mt-1 text-xs text-gray-400">
              Enterprise customers get white-glove onboarding with a dedicated implementation manager.
              For all plans, email{' '}
              <a href="mailto:contact@off-grid-flow.com" className="text-primary-400 hover:underline">
                contact@off-grid-flow.com
              </a>
              {' '}and we&apos;ll walk you through setup.
            </p>
          </div>
        </div>
      </div>

      {/* Implementation timeline */}
      <div className="mt-6 rounded-lg border border-gray-800 bg-gray-800/30 p-5">
        <h3 className="mb-3 text-sm font-semibold text-gray-400">Typical Implementation Timeline</h3>
        <div className="grid gap-4 sm:grid-cols-3">
          <div>
            <div className="text-xs text-gray-500">Self-Serve (Audit Prep)</div>
            <div className="mt-1 text-sm font-medium text-white">Under 2 hours</div>
            <div className="mt-0.5 text-[10px] text-gray-600">CSV upload → report</div>
          </div>
          <div>
            <div className="text-xs text-gray-500">Assisted (Compliance Pro)</div>
            <div className="mt-1 text-sm font-medium text-white">1–2 weeks</div>
            <div className="mt-0.5 text-[10px] text-gray-600">Cloud connectors + Scope 3 setup</div>
          </div>
          <div>
            <div className="text-xs text-gray-500">Enterprise</div>
            <div className="mt-1 text-sm font-medium text-white">2–4 weeks</div>
            <div className="mt-0.5 text-[10px] text-gray-600">SAP integration + approval workflows</div>
          </div>
        </div>
      </div>
    </div>
  );
}
