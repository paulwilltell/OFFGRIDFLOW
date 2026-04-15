'use client';

import { useMemo } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  downloadAbatementReport,
  evaluateAbatementRisk,
  fetchAbatementDashboard,
  selfCertifyAbatementRisk,
} from '@/lib/api/abatement';
import { toast } from '@/app/components/Toast';
import type {
  AbatementDashboardData,
  AbatementFramework,
  EvaluateAbatementResponse,
  RiskCardData,
} from '@/lib/abatement/types';
import { RiskCard } from './RiskCard';

type AbatementDashboardProps = {
  framework: AbatementFramework;
};

const frameworkLabel: Record<AbatementFramework, string> = {
  sb253: 'California SB 253',
  csrd: 'CSRD / ESRS E1',
  sec: 'SEC Climate Disclosure',
  ifrs: 'IFRS S2',
  cbam: 'EU CBAM',
};

function updateRisk(
  previous: AbatementDashboardData | undefined,
  complianceCheckId: string,
  updater: (risk: RiskCardData) => RiskCardData,
): AbatementDashboardData | undefined {
  if (!previous) {
    return previous;
  }

  return {
    ...previous,
    risks: previous.risks.map((risk) =>
      risk.complianceCheckId === complianceCheckId ? updater(risk) : risk,
    ),
  };
}

export function AbatementDashboard({ framework }: AbatementDashboardProps) {
  const queryClient = useQueryClient();

  const dashboardQuery = useQuery({
    queryKey: ['abatement', framework],
    queryFn: () => fetchAbatementDashboard(framework),
  });

  const evaluateMutation = useMutation({
    mutationFn: ({
      complianceCheckId,
      completed,
      justification,
      files,
    }: {
      complianceCheckId: string;
      completed: boolean;
      justification: string;
      files: File[];
    }) =>
      evaluateAbatementRisk(framework, {
        complianceCheckId,
        completed,
        justification,
        files,
      }),
    onMutate: async (variables) => {
      await queryClient.cancelQueries({ queryKey: ['abatement', framework] });
      const previous = queryClient.getQueryData<AbatementDashboardData>(['abatement', framework]);
      queryClient.setQueryData<AbatementDashboardData | undefined>(
        ['abatement', framework],
        updateRisk(previous, variables.complianceCheckId, (risk) => ({
          ...risk,
          completed: variables.completed,
          justification: variables.justification,
        })),
      );
      return { previous };
    },
    onError: (error, _variables, context) => {
      if (context?.previous) {
        queryClient.setQueryData(['abatement', framework], context.previous);
      }
      toast.error(error instanceof Error ? error.message : 'Failed to evaluate remediation.');
    },
    onSuccess: (response: EvaluateAbatementResponse) => {
      queryClient.setQueryData<AbatementDashboardData | undefined>(
        ['abatement', framework],
        (current) =>
          current
            ? {
                ...current,
                risks: current.risks.map((risk) =>
                  risk.complianceCheckId === response.risk.complianceCheckId ? response.risk : risk,
                ),
              }
            : current,
      );

      const statusCopy = {
        recommended: 'Assessment complete: recommendation is positive.',
        needs_clarification: 'Assessment complete: more detail is recommended.',
        insufficient: 'Assessment complete: the item still needs stronger support.',
      } as const;
      toast.success(statusCopy[response.evaluation.status]);
    },
    onSettled: async () => {
      await queryClient.invalidateQueries({ queryKey: ['abatement', framework] });
    },
  });

  const selfCertifyMutation = useMutation({
    mutationFn: ({
      actionItemId,
      complianceCheckId,
      selfCertified,
    }: {
      actionItemId: string;
      complianceCheckId: string;
      selfCertified: boolean;
    }) =>
      selfCertifyAbatementRisk(framework, {
        actionItemId,
        complianceCheckId,
        selfCertified,
      }),
    onMutate: async (variables) => {
      await queryClient.cancelQueries({ queryKey: ['abatement', framework] });
      const previous = queryClient.getQueryData<AbatementDashboardData>(['abatement', framework]);
      queryClient.setQueryData<AbatementDashboardData | undefined>(
        ['abatement', framework],
        updateRisk(previous, variables.complianceCheckId, (risk) => ({
          ...risk,
          selfCertified: variables.selfCertified,
          certifiedAt: variables.selfCertified ? new Date().toISOString() : null,
        })),
      );
      return { previous };
    },
    onError: (error, _variables, context) => {
      if (context?.previous) {
        queryClient.setQueryData(['abatement', framework], context.previous);
      }
      toast.error(error instanceof Error ? error.message : 'Failed to update self-certification.');
    },
    onSuccess: (response) => {
      queryClient.setQueryData<AbatementDashboardData | undefined>(
        ['abatement', framework],
        (current) =>
          current
            ? {
                ...current,
                risks: current.risks.map((risk) =>
                  risk.complianceCheckId === response.risk.complianceCheckId ? response.risk : risk,
                ),
              }
            : current,
      );
      toast.info(
        response.risk.selfCertified
          ? 'Item marked as resolved (self-certified).'
          : 'Self-certification removed.',
      );
    },
    onSettled: async () => {
      await queryClient.invalidateQueries({ queryKey: ['abatement', framework] });
    },
  });

  const reportMutation = useMutation({
    mutationFn: () => downloadAbatementReport(framework),
    onError: (error) => {
      toast.error(error instanceof Error ? error.message : 'Failed to generate report.');
    },
  });

  const activeEvaluateCheckId = evaluateMutation.variables?.complianceCheckId;
  const activeSelfCertifyCheckId = selfCertifyMutation.variables?.complianceCheckId;

  const addressedSummary = useMemo(() => {
    if (!dashboardQuery.data) {
      return 'Calculating current risk exposure…';
    }
    return `${dashboardQuery.data.progress.criticalAddressed} of ${dashboardQuery.data.progress.criticalTotal} critical risks addressed`;
  }, [dashboardQuery.data]);

  if (dashboardQuery.isLoading) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-semibold text-white">Risk Abatement</h1>
        <p className="text-sm text-slate-400">Loading the current risk picture…</p>
      </div>
    );
  }

  if (dashboardQuery.isError || !dashboardQuery.data) {
    return (
      <div className="rounded-3xl border border-rose-500/30 bg-rose-500/10 p-6">
        <h1 className="text-2xl font-semibold text-white">Risk Abatement</h1>
        <p className="mt-3 text-sm leading-6 text-rose-100">
          {dashboardQuery.error instanceof Error
            ? dashboardQuery.error.message
            : 'Failed to load the abatement dashboard.'}
        </p>
      </div>
    );
  }

  const dashboard = dashboardQuery.data;

  return (
    <div className="space-y-6">
      <div className="rounded-3xl border border-amber-400/40 bg-amber-400/10 p-5">
        <p className="text-xs font-semibold uppercase tracking-[0.22em] text-amber-200">
          Guidance only
        </p>
        <p className="mt-2 max-w-4xl text-sm leading-6 text-amber-50">
          {dashboard.disclaimer}
        </p>
      </div>

      <div className="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.22em] text-emerald-300">
            {frameworkLabel[framework]}
          </p>
          <h1 className="mt-2 text-3xl font-semibold text-white">Risk Abatement Dashboard</h1>
          <p className="mt-3 max-w-3xl text-sm leading-6 text-slate-300">
            OffGridFlow identifies regulatory gaps. Your team explains the remediation. The engine
            evaluates that explanation against audit-readiness checks and returns a reasoned
            recommendation.
          </p>
        </div>

        <button
          type="button"
          onClick={() => reportMutation.mutate()}
          disabled={reportMutation.isPending}
          className="rounded-2xl border border-emerald-400 bg-emerald-400 px-4 py-3 text-sm font-semibold text-slate-950 transition hover:bg-emerald-300 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {reportMutation.isPending ? 'Generating report…' : 'Generate Draft Report'}
        </button>
      </div>

      <section className="grid gap-4 xl:grid-cols-[1.35fr_0.65fr]">
        <div className="rounded-3xl border border-slate-800 bg-slate-900/70 p-5">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-400">
                Risk summary
              </p>
              <h2 className="mt-2 text-xl font-semibold text-white">Current exposure</h2>
            </div>
            <p className="text-xs uppercase tracking-[0.22em] text-slate-500">
              Reporting year {dashboard.reportingYear}
            </p>
          </div>

          <div className="mt-5 grid gap-3 md:grid-cols-3">
            <SummaryCard label="High" value={dashboard.summary.high} tone="rose" />
            <SummaryCard label="Medium" value={dashboard.summary.medium} tone="amber" />
            <SummaryCard label="Low" value={dashboard.summary.low} tone="slate" />
          </div>

          <div className="mt-6">
            <div className="flex items-center justify-between text-sm">
              <span className="font-medium text-slate-100">Critical risk progress</span>
              <span className="text-slate-300">{dashboard.progress.percentAddressed}%</span>
            </div>
            <div className="mt-3 h-3 rounded-full bg-slate-800">
              <div
                className="h-3 rounded-full bg-emerald-400 transition-all"
                style={{ width: `${dashboard.progress.percentAddressed}%` }}
              />
            </div>
            <p className="mt-3 text-sm text-slate-400">{addressedSummary}</p>
          </div>
        </div>

        <div className="rounded-3xl border border-slate-800 bg-slate-900/70 p-5">
          <p className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-400">
            {dashboard.framework.penalty_heading}
          </p>
          <h2 className="mt-2 text-xl font-semibold text-white">Penalty exposure</h2>
          <p className="mt-4 text-sm leading-6 text-slate-300">{dashboard.framework.penalty_body}</p>
          <div className="mt-6 rounded-2xl border border-slate-800 bg-slate-950/70 p-4">
            <p className="text-xs uppercase tracking-[0.18em] text-slate-500">Assessment timestamp</p>
            <p className="mt-2 text-sm text-slate-200">
              {new Date(dashboard.generatedAt).toLocaleString()}
            </p>
          </div>
        </div>
      </section>

      {dashboard.risks.length === 0 ? (
        <section className="rounded-3xl border border-emerald-500/30 bg-emerald-500/10 p-6">
          <h2 className="text-xl font-semibold text-white">No current risks identified</h2>
          <p className="mt-3 max-w-3xl text-sm leading-6 text-emerald-50">
            No triggered abatement items were found for this framework and reporting year. Generate
            a draft report to export the current workplan snapshot.
          </p>
        </section>
      ) : (
        <section className="space-y-5">
          {dashboard.risks.map((risk) => (
            <RiskCard
              key={risk.id}
              framework={framework}
              risk={risk}
              isEvaluating={evaluateMutation.isPending && activeEvaluateCheckId === risk.complianceCheckId}
              isSelfCertifying={
                selfCertifyMutation.isPending && activeSelfCertifyCheckId === risk.complianceCheckId
              }
              onEvaluate={(payload) => evaluateMutation.mutateAsync(payload)}
              onSelfCertify={(payload) => selfCertifyMutation.mutateAsync(payload)}
            />
          ))}
        </section>
      )}
    </div>
  );
}

function SummaryCard({
  label,
  value,
  tone,
}: {
  label: string;
  value: number;
  tone: 'rose' | 'amber' | 'slate';
}) {
  const tones = {
    rose: 'border-rose-500/40 bg-rose-500/10 text-rose-100',
    amber: 'border-amber-400/40 bg-amber-400/10 text-amber-100',
    slate: 'border-slate-700 bg-slate-950/70 text-slate-100',
  } as const;

  return (
    <div className={`rounded-2xl border p-4 ${tones[tone]}`}>
      <p className="text-xs font-semibold uppercase tracking-[0.22em] opacity-80">{label}</p>
      <p className="mt-3 text-3xl font-semibold">{value}</p>
    </div>
  );
}
