'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useSession, useRequireAuth } from '@/lib/session';

type OnboardingSegment = 'self_serve' | 'assisted' | 'enterprise';

const segmentConfig: Record<OnboardingSegment, { label: string; desc: string; timeline: string }> = {
  self_serve: { label: 'Self-Serve', desc: 'CSV upload to first report', timeline: 'Under 2 hours' },
  assisted: { label: 'Assisted', desc: 'Cloud connectors + Scope 3', timeline: '1-2 weeks' },
  enterprise: { label: 'Enterprise', desc: 'Full integration + approval workflows', timeline: '2-4 weeks' },
};

const milestones = [
  {
    step: 1,
    title: 'Account Created',
    description: 'Your account is set up with your organization profile.',
    action: null,
    href: null,
    autoComplete: true,
    segments: ['self_serve', 'assisted', 'enterprise'] as OnboardingSegment[],
    training: null,
  },
  {
    step: 2,
    title: 'Organization Profile',
    description: 'Set your industry, country, fiscal year, and reporting preferences.',
    action: 'Set Up Profile',
    href: '/settings/organization',
    autoComplete: false,
    segments: ['self_serve', 'assisted', 'enterprise'] as OnboardingSegment[],
    training: { role: 'admin', topic: 'Organization settings and user management' },
  },
  {
    step: 3,
    title: 'First Data Upload',
    description: 'Upload a CSV with utility bills, energy consumption, or fleet data.',
    action: 'Upload CSV',
    href: '/emissions',
    autoComplete: false,
    segments: ['self_serve', 'assisted', 'enterprise'] as OnboardingSegment[],
    training: { role: 'operator', topic: 'CSV format requirements and data validation' },
  },
  {
    step: 4,
    title: 'Connect a Cloud Source',
    description: 'Set up automated data pipelines from AWS, Azure, GCP, SAP, or utility providers.',
    action: 'Configure Sources',
    href: '/settings/data-sources',
    autoComplete: false,
    segments: ['assisted', 'enterprise'] as OnboardingSegment[],
    training: { role: 'operator', topic: 'Cloud connector setup and credential management' },
  },
  {
    step: 5,
    title: 'Run Data Quality Scan',
    description: 'Verify data integrity — detect outliers, duplicates, and missing periods before reporting.',
    action: 'Run Scan',
    href: '/audit/data-quality',
    autoComplete: false,
    segments: ['self_serve', 'assisted', 'enterprise'] as OnboardingSegment[],
    training: { role: 'operator', topic: 'Data quality checks and anomaly resolution' },
  },
  {
    step: 6,
    title: 'Review Emissions Dashboard',
    description: 'Check Scope 1, 2 & 3 emissions on the dashboard. Verify factor accuracy.',
    action: 'Open Dashboard',
    href: '/dashboard/carbon',
    autoComplete: false,
    segments: ['self_serve', 'assisted', 'enterprise'] as OnboardingSegment[],
    training: { role: 'executive', topic: 'Dashboard views and KPI interpretation' },
  },
  {
    step: 7,
    title: 'Lock Factor Snapshot',
    description: 'Freeze emission factors to your reporting period for audit reproducibility.',
    action: 'Factor Snapshots',
    href: '/audit/factor-snapshots',
    autoComplete: false,
    segments: ['assisted', 'enterprise'] as OnboardingSegment[],
    training: { role: 'auditor', topic: 'Factor version locking and calculation ledger' },
  },
  {
    step: 8,
    title: 'Generate First Report',
    description: 'Generate an audit-ready compliance report (CSRD, SEC, SB 253, or CBAM).',
    action: 'Generate Report',
    href: '/compliance/csrd',
    autoComplete: false,
    segments: ['self_serve', 'assisted', 'enterprise'] as OnboardingSegment[],
    training: { role: 'executive', topic: 'Compliance framework selection and report generation' },
  },
  {
    step: 9,
    title: 'Submit for Review',
    description: 'Create an approval request and submit your inventory for internal review.',
    action: 'Start Approval',
    href: '/audit/approvals',
    autoComplete: false,
    segments: ['assisted', 'enterprise'] as OnboardingSegment[],
    training: { role: 'auditor', topic: 'Approval workflow and audit trail review' },
  },
  {
    step: 10,
    title: 'Approved & Locked',
    description: 'Report reviewed, approved, and locked. Ready for stakeholders or auditors.',
    action: 'View Approvals',
    href: '/audit/approvals',
    autoComplete: false,
    segments: ['assisted', 'enterprise'] as OnboardingSegment[],
    training: null,
  },
];

export default function OnboardingPage() {
  useRequireAuth();
  const { user } = useSession();
  const [segment, setSegment] = useState<OnboardingSegment>('self_serve');
  const [milestoneData, setMilestoneData] = useState<any>(null);

  // Load milestone data from API
  useEffect(() => {
    fetch('/api/audit/health', { credentials: 'include' })
      .then((r) => r.json())
      .then((d) => setMilestoneData(d.health))
      .catch(() => {});
  }, []);

  // Filter milestones by segment
  const filteredMilestones = milestones.filter((m) => m.segments.includes(segment));

  // Calculate completion based on real data
  let completedSteps = 0;
  if (user) completedSteps = 1; // Account created
  if (milestoneData?.total_activities_count > 0) completedSteps = Math.max(completedSteps, 3);
  if (milestoneData?.reports_generated_count > 0) completedSteps = Math.max(completedSteps, 6);

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-xl font-bold text-white">Getting Started</h1>
        <p className="mt-1 text-xs text-gray-500">
          Follow these steps to reach your first audit-ready report.
        </p>
      </div>

      {/* Segment Selector */}
      <div className="mb-6 flex items-center gap-1 rounded-lg border border-gray-800 bg-gray-900/50 p-1">
        {(Object.entries(segmentConfig) as [OnboardingSegment, typeof segmentConfig[OnboardingSegment]][]).map(([key, cfg]) => (
          <button
            key={key}
            onClick={() => setSegment(key)}
            className={`flex-1 rounded-md px-4 py-2 text-center transition ${
              segment === key
                ? 'bg-primary-600/20 text-primary-400'
                : 'text-gray-500 hover:text-gray-300'
            }`}
          >
            <div className="text-xs font-medium">{cfg.label}</div>
            <div className="text-[10px] text-gray-600">{cfg.timeline}</div>
          </button>
        ))}
      </div>

      {/* Progress bar */}
      <div className="mb-8">
        <div className="mb-2 flex items-center justify-between text-xs text-gray-500">
          <span>{Math.min(completedSteps, filteredMilestones.length)} of {filteredMilestones.length} complete</span>
          <span>{Math.round((Math.min(completedSteps, filteredMilestones.length) / filteredMilestones.length) * 100)}%</span>
        </div>
        <div className="h-2 overflow-hidden rounded-full bg-gray-800">
          <div
            className="h-full rounded-full bg-primary-600 transition-all duration-500"
            style={{ width: `${(Math.min(completedSteps, filteredMilestones.length) / filteredMilestones.length) * 100}%` }}
          />
        </div>
      </div>

      {/* Milestones */}
      <div className="space-y-3">
        {filteredMilestones.map((m, i) => {
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
                  <div className="mt-3 flex items-center gap-2">
                    {m.action && m.href && (isCurrent || isComplete) && (
                      <Link
                        href={m.href}
                        className={`inline-block rounded-lg px-4 py-1.5 text-xs font-medium transition ${
                          isCurrent ? 'bg-primary-600 text-white hover:bg-primary-500' :
                          'border border-gray-700 text-gray-400 hover:text-white'
                        }`}
                      >
                        {m.action}
                      </Link>
                    )}
                    {m.training && (isCurrent || isComplete) && (
                      <span className="rounded border border-gray-800 bg-gray-900/50 px-2.5 py-1 text-[10px] text-gray-500">
                        {m.training.role === 'admin' ? 'Admin' :
                         m.training.role === 'operator' ? 'Operator' :
                         m.training.role === 'executive' ? 'Executive' : 'Auditor'} — {m.training.topic}
                      </span>
                    )}
                  </div>
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
