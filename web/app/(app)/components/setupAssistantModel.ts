export type AssistantStepId =
  | 'profile'
  | 'tour'
  | 'upload'
  | 'connect'
  | 'report';

export type AssistantStepStatus = 'complete' | 'current' | 'pending';

export interface AssistantProgressInput {
  profileReviewed: boolean;
  dashboardTourCompleted: boolean;
  hasUploadedData: boolean;
  hasConnectedDataSource: boolean;
  hasGeneratedReport: boolean;
}

export interface AssistantStep {
  id: AssistantStepId;
  title: string;
  description: string;
  href?: string;
  ctaLabel?: string;
  optional?: boolean;
  completed: boolean;
  status: AssistantStepStatus;
}

const STEP_META: Array<Omit<AssistantStep, 'completed' | 'status'>> = [
  {
    id: 'profile',
    title: 'Review organization profile',
    description:
      'Confirm your reporting entity, fiscal year, and workspace context before you ingest data.',
    href: '/settings/organization',
    ctaLabel: 'Open profile',
  },
  {
    id: 'upload',
    title: 'Upload your first emissions dataset',
    description:
      'Import a CSV or activity file so OffGridFlow can calculate your first inventory and populate the dashboard.',
    href: '/emissions',
    ctaLabel: 'Upload data',
  },
  {
    id: 'tour',
    title: 'Take the dashboard tour',
    description:
      'Get a guided walkthrough of the live KPI, trend, compliance, and action surfaces in the dashboard.',
    href: '/dashboard/carbon',
    ctaLabel: 'Start tour',
  },
  {
    id: 'connect',
    title: 'Connect a live data source',
    description:
      'Automate refreshes from cloud, ERP, or utility systems so reporting stays current without manual uploads.',
    href: '/settings/data-sources',
    ctaLabel: 'Connect source',
    optional: true,
  },
  {
    id: 'report',
    title: 'Generate your first compliance report',
    description:
      'Produce the first audit-ready output after your data lands so your team can validate scope coverage and disclosures.',
    href: '/compliance/csrd',
    ctaLabel: 'Generate report',
  },
];

function isStepComplete(id: AssistantStepId, progress: AssistantProgressInput): boolean {
  switch (id) {
    case 'profile':
      return progress.profileReviewed;
    case 'tour':
      return progress.dashboardTourCompleted;
    case 'upload':
      return progress.hasUploadedData;
    case 'connect':
      return progress.hasConnectedDataSource;
    case 'report':
      return progress.hasGeneratedReport;
    default:
      return false;
  }
}

export function buildAssistantSteps(progress: AssistantProgressInput): AssistantStep[] {
  const completionMap = new Map<AssistantStepId, boolean>(
    STEP_META.map((step) => [step.id, isStepComplete(step.id, progress)]),
  );

  const nextRequiredId =
    STEP_META.find((step) => !step.optional && !completionMap.get(step.id))?.id ?? null;
  const nextAnyId = STEP_META.find((step) => !completionMap.get(step.id))?.id ?? null;
  const currentId = nextRequiredId ?? nextAnyId;

  return STEP_META.map((step) => {
    const completed = completionMap.get(step.id) ?? false;
    const status: AssistantStepStatus = completed
      ? 'complete'
      : step.id === currentId
        ? 'current'
        : 'pending';

    return {
      ...step,
      completed,
      status,
    };
  });
}

export function getAssistantProgress(steps: AssistantStep[]) {
  const total = steps.length;
  const completeCount = steps.filter((step) => step.completed).length;
  const requiredTotal = steps.filter((step) => !step.optional).length;
  const requiredCompleteCount = steps.filter((step) => !step.optional && step.completed).length;

  return {
    total,
    completeCount,
    requiredTotal,
    requiredCompleteCount,
    percentComplete: total === 0 ? 0 : Math.round((completeCount / total) * 100),
    requiredPercentComplete:
      requiredTotal === 0 ? 0 : Math.round((requiredCompleteCount / requiredTotal) * 100),
  };
}

export function getNextAssistantStep(steps: AssistantStep[]): AssistantStep | null {
  return steps.find((step) => !step.completed) ?? null;
}
