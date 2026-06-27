'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { toast } from '@/app/components/Toast';
import { CarbonApi } from '@/lib/api/carbon';
import { useSession } from '@/lib/session';
import {
  buildAssistantSteps,
  getAssistantProgress,
  getNextAssistantStep,
  type AssistantStep,
} from './setupAssistantModel';

const STORAGE_KEYS = {
  profileReviewed: 'offgridflow_setup_profile_reviewed',
  assistantHidden: 'offgridflow_setup_assistant_hidden',
  assistantCollapsed: 'offgridflow_setup_assistant_collapsed',
  tourCompleted: 'offgridflow_dashboard_tour_completed',
  pendingTour: 'offgridflow_dashboard_tour_pending',
} as const;

const OPEN_ASSISTANT_EVENT = 'offgridflow:open-setup-assistant';

type AuditHealth = {
  total_activities_count?: number;
  reports_generated_count?: number;
};

type TourStep = {
  selector: string;
  title: string;
  description: string;
};

const DASHBOARD_TOUR_STEPS: TourStep[] = [
  {
    selector: '.dashboard-header',
    title: 'Dashboard command bar',
    description:
      'This is your control point for the active dashboard view, methodology reference, and exports.',
  },
  {
    selector: '.dashboard-kpi-grid',
    title: 'Live carbon KPIs',
    description:
      'These cards summarize total emissions and scope splits for the current reporting period.',
  },
  {
    selector: '.dashboard-chart',
    title: 'Trend and variance view',
    description:
      'Use the trend surface to inspect movement over time and validate whether new data changed the inventory.',
  },
  {
    selector: '.dashboard-compliance-status',
    title: 'Framework readiness',
    description:
      'This panel shows how the current inventory maps into your main disclosure and compliance workflows.',
  },
  {
    selector: '.dashboard-quick-actions',
    title: 'Next operational actions',
    description:
      'Use these shortcuts to upload more data, connect systems, or move into compliance and billing flows.',
  },
];

function readStoredBoolean(key: string): boolean {
  if (typeof window === 'undefined') return false;
  return window.localStorage.getItem(key) === 'true';
}

function writeStoredBoolean(key: string, value: boolean) {
  if (typeof window === 'undefined') return;
  window.localStorage.setItem(key, value ? 'true' : 'false');
}

function clearStoredValue(key: string) {
  if (typeof window === 'undefined') return;
  window.localStorage.removeItem(key);
}

function StepBadge({ step }: { step: AssistantStep }) {
  if (step.completed) {
    return (
      <span className="rounded-full bg-emerald-500/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-emerald-300">
        Done
      </span>
    );
  }

  if (step.status === 'current') {
    return (
      <span className="rounded-full bg-primary-500/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-primary-300">
        Next
      </span>
    );
  }

  if (step.optional) {
    return (
      <span className="rounded-full bg-gray-700/70 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-gray-300">
        Optional
      </span>
    );
  }

  return (
    <span className="rounded-full bg-gray-700/70 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-gray-300">
      To do
    </span>
  );
}

function AssistantStepRow({
  step,
  pathname,
  onStartTour,
}: {
  step: AssistantStep;
  pathname: string | null;
  onStartTour: () => void;
}) {
  const isCurrentPage = step.href ? pathname === step.href : false;

  return (
    <div
      className={`rounded-xl border p-3 transition ${
        step.completed
          ? 'border-emerald-500/20 bg-emerald-500/5'
          : step.status === 'current'
            ? 'border-primary-500/30 bg-primary-500/5'
            : 'border-gray-800 bg-gray-900/40'
      }`}
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <div className="flex items-center gap-2">
            <h3 className="text-sm font-semibold text-white">{step.title}</h3>
            <StepBadge step={step} />
          </div>
          <p className="mt-1 text-xs leading-5 text-gray-400">{step.description}</p>
        </div>
      </div>

      {step.completed ? null : (
        <div className="mt-3 flex flex-wrap gap-2">
          {step.id === 'tour' ? (
            <button
              type="button"
              onClick={onStartTour}
              className="rounded-lg bg-primary-600 px-3 py-2 text-xs font-medium text-white transition hover:bg-primary-500"
            >
              {isCurrentPage ? 'Start tour' : 'Open dashboard tour'}
            </button>
          ) : step.href ? (
            <Link
              href={step.href}
              className={`rounded-lg px-3 py-2 text-xs font-medium transition ${
                step.status === 'current'
                  ? 'bg-primary-600 text-white hover:bg-primary-500'
                  : 'border border-gray-700 text-gray-300 hover:border-gray-600 hover:text-white'
              }`}
            >
              {step.ctaLabel ?? 'Open'}
            </Link>
          ) : null}
        </div>
      )}
    </div>
  );
}

export function SetupAssistant() {
  const pathname = usePathname();
  const router = useRouter();
  const { currentTenantId, user, isAuthenticated } = useSession();

  const [assistantVisible, setAssistantVisible] = useState(false);
  const [assistantCollapsed, setAssistantCollapsed] = useState(false);
  const [assistantHidden, setAssistantHidden] = useState(false);
  const [profileReviewed, setProfileReviewed] = useState(false);
  const [dashboardTourCompleted, setDashboardTourCompleted] = useState(false);
  const [health, setHealth] = useState<AuditHealth | null>(null);
  const [connectedSourceCount, setConnectedSourceCount] = useState<number>(0);
  const [loadingSummary, setLoadingSummary] = useState(true);
  const [tourIndex, setTourIndex] = useState<number | null>(null);
  const [highlightRect, setHighlightRect] = useState<DOMRect | null>(null);

  const canStartTour = Boolean(health?.total_activities_count && health.total_activities_count > 0);

  useEffect(() => {
    if (typeof window === 'undefined') return;

    const hidden = readStoredBoolean(STORAGE_KEYS.assistantHidden);
    const collapsed = readStoredBoolean(STORAGE_KEYS.assistantCollapsed);
    const profileDone = readStoredBoolean(STORAGE_KEYS.profileReviewed);
    const tourDone = readStoredBoolean(STORAGE_KEYS.tourCompleted);

    setAssistantHidden(hidden);
    setAssistantCollapsed(collapsed);
    setAssistantVisible(!hidden);
    setProfileReviewed(profileDone);
    setDashboardTourCompleted(tourDone);
  }, []);

  useEffect(() => {
    if (!pathname || typeof window === 'undefined') return;

    if (pathname.startsWith('/settings/organization')) {
      writeStoredBoolean(STORAGE_KEYS.profileReviewed, true);
      setProfileReviewed(true);
    }

    if (
      pathname.startsWith('/dashboard/carbon') &&
      readStoredBoolean(STORAGE_KEYS.pendingTour)
    ) {
      clearStoredValue(STORAGE_KEYS.pendingTour);
      setAssistantHidden(false);
      setAssistantVisible(true);
      setAssistantCollapsed(false);
      setTourIndex(0);
    }
  }, [pathname]);

  useEffect(() => {
    if (!isAuthenticated) return;

    let cancelled = false;

    async function loadSummary() {
      setLoadingSummary(true);

      const healthRequest = fetch('/api/audit/health', {
        credentials: 'include',
      })
        .then(async (response) => {
          if (!response.ok) {
            throw new Error(`Health request failed with ${response.status}`);
          }

          const payload = (await response.json()) as { health?: AuditHealth };
          return payload.health ?? null;
        })
        .catch(() => null);

      const dataSourceRequest = currentTenantId
        ? CarbonApi.getInstance()
            .getDataSources(currentTenantId)
            .then((sources) => (Array.isArray(sources) ? sources.length : 0))
            .catch(() => 0)
        : Promise.resolve(0);

      const [healthSummary, sourceCount] = await Promise.all([healthRequest, dataSourceRequest]);

      if (cancelled) return;

      setHealth(healthSummary);
      setConnectedSourceCount(sourceCount);
      setLoadingSummary(false);
    }

    void loadSummary();

    return () => {
      cancelled = true;
    };
  }, [currentTenantId, isAuthenticated]);

  const steps = useMemo(
    () =>
      buildAssistantSteps({
        profileReviewed,
        dashboardTourCompleted,
        hasUploadedData: (health?.total_activities_count ?? 0) > 0,
        hasConnectedDataSource: connectedSourceCount > 0,
        hasGeneratedReport: (health?.reports_generated_count ?? 0) > 0,
      }),
    [connectedSourceCount, dashboardTourCompleted, health, profileReviewed],
  );

  const progress = useMemo(() => getAssistantProgress(steps), [steps]);
  const nextStep = useMemo(() => getNextAssistantStep(steps), [steps]);

  const resolveTourTarget = useCallback((requestedIndex: number) => {
    for (let index = requestedIndex; index < DASHBOARD_TOUR_STEPS.length; index += 1) {
      const step = DASHBOARD_TOUR_STEPS[index];
      const element = document.querySelector(step.selector);
      if (element instanceof HTMLElement) {
        return { index, element, step };
      }
    }

    return null;
  }, []);

  const finishTour = useCallback(() => {
    setTourIndex(null);
    setHighlightRect(null);
    writeStoredBoolean(STORAGE_KEYS.tourCompleted, true);
    clearStoredValue(STORAGE_KEYS.pendingTour);
    setDashboardTourCompleted(true);
    toast.success('Dashboard tour completed.');
  }, []);

  const closeTour = useCallback((message?: string) => {
    setTourIndex(null);
    setHighlightRect(null);
    clearStoredValue(STORAGE_KEYS.pendingTour);
    if (message) {
      toast.info(message);
    }
  }, []);

  const startTour = useCallback(() => {
    if (!canStartTour) {
      toast.info('Upload your first emissions dataset to unlock the dashboard tour.');
      return;
    }

    if (!pathname?.startsWith('/dashboard/carbon')) {
      writeStoredBoolean(STORAGE_KEYS.pendingTour, true);
      router.push('/dashboard/carbon');
      toast.info('Opening the dashboard tour.');
      return;
    }

    setAssistantHidden(false);
    setAssistantVisible(true);
    setAssistantCollapsed(false);
    setTourIndex(0);
  }, [canStartTour, pathname, router]);

  useEffect(() => {
    if (tourIndex === null) return;

    const resolved = resolveTourTarget(tourIndex);
    if (!resolved) {
      finishTour();
      return;
    }

    if (resolved.index !== tourIndex) {
      setTourIndex(resolved.index);
      return;
    }

    const updateHighlight = () => {
      const nextResolved = resolveTourTarget(tourIndex);
      if (!nextResolved) return;

      nextResolved.element.scrollIntoView({
        block: 'center',
        inline: 'nearest',
        behavior: 'smooth',
      });

      const rect = nextResolved.element.getBoundingClientRect();
      setHighlightRect(rect);
    };

    updateHighlight();

    window.addEventListener('resize', updateHighlight);
    window.addEventListener('scroll', updateHighlight, true);

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        closeTour('Dashboard tour skipped. You can reopen it from the setup assistant.');
      }
    };

    window.addEventListener('keydown', onKeyDown);

    return () => {
      window.removeEventListener('resize', updateHighlight);
      window.removeEventListener('scroll', updateHighlight, true);
      window.removeEventListener('keydown', onKeyDown);
    };
  }, [closeTour, finishTour, resolveTourTarget, tourIndex]);

  useEffect(() => {
    if (typeof window === 'undefined') return;

    const onOpenAssistant = () => {
      clearStoredValue(STORAGE_KEYS.assistantHidden);
      clearStoredValue(STORAGE_KEYS.assistantCollapsed);
      setAssistantHidden(false);
      setAssistantCollapsed(false);
      setAssistantVisible(true);
      toast.info('Setup assistant reopened.');
    };

    window.addEventListener(OPEN_ASSISTANT_EVENT, onOpenAssistant);
    return () => {
      window.removeEventListener(OPEN_ASSISTANT_EVENT, onOpenAssistant);
    };
  }, []);

  const handleHideAssistant = () => {
    writeStoredBoolean(STORAGE_KEYS.assistantHidden, true);
    writeStoredBoolean(STORAGE_KEYS.assistantCollapsed, false);
    setAssistantHidden(true);
    setAssistantVisible(false);
    setAssistantCollapsed(false);
    toast.info('Setup assistant hidden. Reopen it anytime from Help.');
  };

  const handleCollapseToggle = () => {
    const nextCollapsed = !assistantCollapsed;
    writeStoredBoolean(STORAGE_KEYS.assistantCollapsed, nextCollapsed);
    setAssistantCollapsed(nextCollapsed);
  };

  const currentTourStep =
    tourIndex !== null && tourIndex < DASHBOARD_TOUR_STEPS.length
      ? DASHBOARD_TOUR_STEPS[tourIndex]
      : null;
  const activeTourIndex = tourIndex ?? 0;

  const tooltipStyle = useMemo(() => {
    if (!highlightRect || typeof window === 'undefined') {
      return {
        top: '50%',
        left: '50%',
        transform: 'translate(-50%, -50%)',
      };
    }

    const width = Math.min(360, window.innerWidth - 32);
    const preferredTop = highlightRect.bottom + 16;
    const top =
      preferredTop + 240 < window.innerHeight
        ? preferredTop
        : Math.max(16, highlightRect.top - 220);
    const left = Math.min(
      Math.max(16, highlightRect.left),
      Math.max(16, window.innerWidth - width - 16),
    );

    return {
      top,
      left,
      width,
    };
  }, [highlightRect]);

  if (!isAuthenticated || assistantHidden) {
    return currentTourStep && highlightRect ? (
      <>
        <div
          className="pointer-events-none fixed z-50 rounded-2xl border border-primary-400/70 bg-transparent shadow-[0_0_0_9999px_rgba(3,7,18,0.72)] transition-all duration-200"
          style={{
            top: Math.max(8, highlightRect.top - 8),
            left: Math.max(8, highlightRect.left - 8),
            width: highlightRect.width + 16,
            height: highlightRect.height + 16,
          }}
        />
      </>
    ) : null;
  }

  return (
    <>
      {assistantVisible && assistantCollapsed && tourIndex === null ? (
        <button
          type="button"
          onClick={handleCollapseToggle}
          className="fixed right-6 top-24 z-30 flex items-center gap-3 rounded-full border border-gray-700 bg-gray-950/95 px-4 py-2 text-left shadow-xl backdrop-blur sm:max-w-sm"
        >
          <span className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-primary-500/15 text-sm text-primary-300">
            ?
          </span>
          <span className="min-w-0">
            <span className="block text-xs font-semibold uppercase tracking-wide text-primary-300">
              Setup assistant
            </span>
            <span className="block truncate text-sm text-white">
              {progress.completeCount}/{progress.total} steps complete
            </span>
          </span>
        </button>
      ) : null}

      {assistantVisible && !assistantCollapsed ? (
        <aside className="fixed inset-x-4 bottom-20 z-30 rounded-2xl border border-gray-700 bg-gray-950/95 p-4 shadow-2xl backdrop-blur sm:inset-auto sm:right-6 sm:top-24 sm:w-[380px]">
          <div className="flex items-start justify-between gap-3">
            <div>
              <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-primary-300">
                Setup assistant
              </p>
              <h2 className="mt-1 text-lg font-semibold text-white">
                {progress.requiredPercentComplete === 100
                  ? 'Core setup complete'
                  : 'Get your Scope Calculations faster'}
              </h2>
              <p className="mt-1 text-sm leading-6 text-gray-400">
                {loadingSummary
                  ? 'Checking your workspace status.'
                  : nextStep
                    ? `Next recommended action: ${nextStep.title.toLowerCase()}.`
                    : 'Everything required for first value is in place. You can still connect live sources for automation.'}
              </p>
            </div>

            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={handleCollapseToggle}
                className="rounded-lg border border-gray-700 px-2 py-1 text-xs text-gray-300 transition hover:border-gray-600 hover:text-white"
                aria-label="Collapse setup assistant"
              >
                Hide
              </button>
              <button
                type="button"
                onClick={handleHideAssistant}
                className="rounded-lg border border-gray-700 px-2 py-1 text-xs text-gray-300 transition hover:border-gray-600 hover:text-white"
                aria-label="Skip setup assistant"
              >
                Skip
              </button>
            </div>
          </div>

          <div className="mt-4 rounded-xl border border-gray-800 bg-gray-900/60 p-3">
            <div className="flex items-center justify-between text-xs text-gray-400">
              <span>{progress.completeCount} of {progress.total} steps complete</span>
              <span>{progress.percentComplete}%</span>
            </div>
            <div className="mt-2 h-2 overflow-hidden rounded-full bg-gray-800">
              <div
                className="h-full rounded-full bg-primary-500 transition-all duration-300"
                style={{ width: `${progress.percentComplete}%` }}
              />
            </div>

            <div className="mt-3 grid grid-cols-2 gap-2 text-xs text-gray-400">
              <div className="rounded-lg border border-gray-800 bg-gray-950/80 px-3 py-2">
                <span className="block text-[10px] uppercase tracking-wide text-gray-500">
                  Activities
                </span>
                <span className="mt-1 block text-sm font-semibold text-white">
                  {health?.total_activities_count ?? 0}
                </span>
              </div>
              <div className="rounded-lg border border-gray-800 bg-gray-950/80 px-3 py-2">
                <span className="block text-[10px] uppercase tracking-wide text-gray-500">
                  Live sources
                </span>
                <span className="mt-1 block text-sm font-semibold text-white">
                  {connectedSourceCount}
                </span>
              </div>
            </div>
          </div>

          <div className="mt-4 space-y-3">
            {steps.map((step) => (
              <AssistantStepRow
                key={step.id}
                step={step}
                pathname={pathname}
                onStartTour={startTour}
              />
            ))}
          </div>

          <div className="mt-4 flex flex-wrap gap-2">
            <Link
              href="/onboarding"
              className="rounded-lg border border-gray-700 px-3 py-2 text-xs font-medium text-gray-300 transition hover:border-gray-600 hover:text-white"
            >
              Open full checklist
            </Link>
            <button
              type="button"
              onClick={startTour}
              className="rounded-lg bg-primary-600 px-3 py-2 text-xs font-medium text-white transition hover:bg-primary-500 disabled:cursor-not-allowed disabled:bg-primary-900/40 disabled:text-primary-200/60"
              disabled={!canStartTour}
            >
              {pathname?.startsWith('/dashboard/carbon') ? 'Start dashboard tour' : 'Go to dashboard tour'}
            </button>
          </div>

          <p className="mt-3 text-[11px] leading-5 text-gray-500">
            Skipping the assistant does not block anything. You can reopen it from the help menu whenever you need another pass through setup.
          </p>
        </aside>
      ) : null}

      {currentTourStep && highlightRect ? (
        <>
          <div
            className="pointer-events-none fixed z-40 rounded-2xl border border-primary-400/70 bg-transparent shadow-[0_0_0_9999px_rgba(3,7,18,0.72)] transition-all duration-200"
            style={{
              top: Math.max(8, highlightRect.top - 8),
              left: Math.max(8, highlightRect.left - 8),
              width: highlightRect.width + 16,
              height: highlightRect.height + 16,
            }}
          />

          <div
            className="fixed z-[60] rounded-2xl border border-primary-500/30 bg-gray-950/98 p-4 shadow-2xl backdrop-blur"
            style={tooltipStyle}
            role="dialog"
            aria-labelledby="dashboard-tour-title"
          >
            <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-primary-300">
              Dashboard tour
            </p>
            <h3 id="dashboard-tour-title" className="mt-1 text-base font-semibold text-white">
              {currentTourStep.title}
            </h3>
            <p className="mt-2 text-sm leading-6 text-gray-400">
              {currentTourStep.description}
            </p>
            <div className="mt-4 flex items-center justify-between">
              <span className="text-xs text-gray-500">
                Step {activeTourIndex + 1} of {DASHBOARD_TOUR_STEPS.length}
              </span>
              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={() => {
                    if (activeTourIndex === 0) {
                      closeTour('Dashboard tour skipped. You can restart it from the setup assistant.');
                      return;
                    }

                    setTourIndex(activeTourIndex - 1);
                  }}
                  className="rounded-lg border border-gray-700 px-3 py-2 text-xs font-medium text-gray-300 transition hover:border-gray-600 hover:text-white"
                >
                  {activeTourIndex === 0 ? 'Skip' : 'Back'}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    if (activeTourIndex >= DASHBOARD_TOUR_STEPS.length - 1) {
                      finishTour();
                      return;
                    }

                    setTourIndex(activeTourIndex + 1);
                  }}
                  className="rounded-lg bg-primary-600 px-3 py-2 text-xs font-medium text-white transition hover:bg-primary-500"
                >
                  {activeTourIndex >= DASHBOARD_TOUR_STEPS.length - 1 ? 'Finish' : 'Next'}
                </button>
              </div>
            </div>
          </div>
        </>
      ) : null}
    </>
  );
}

export default SetupAssistant;
