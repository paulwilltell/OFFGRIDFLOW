'use client';

import { useEffect, useMemo, useState } from 'react';
import type { AbatementFramework, RiskCardData } from '@/lib/abatement/types';
import { RichTextInput } from './RichTextInput';

type EvaluatePayload = {
  complianceCheckId: string;
  completed: boolean;
  justification: string;
  files: File[];
};

type RiskCardProps = {
  framework: AbatementFramework;
  risk: RiskCardData;
  isEvaluating: boolean;
  isSelfCertifying: boolean;
  onEvaluate: (payload: EvaluatePayload) => Promise<void>;
  onSelfCertify: (payload: {
    actionItemId: string;
    complianceCheckId: string;
    selfCertified: boolean;
  }) => Promise<void>;
};

const severityTone: Record<RiskCardData['severity'], string> = {
  blocker: 'border-rose-500/50 bg-rose-500/10 text-rose-100',
  warning: 'border-amber-400/50 bg-amber-400/10 text-amber-100',
};

const priorityTone: Record<RiskCardData['priority'], string> = {
  high: 'text-rose-200',
  medium: 'text-amber-200',
  low: 'text-slate-200',
};

const assessmentTone = {
  recommended: 'border-emerald-500/50 bg-emerald-500/10 text-emerald-100',
  needs_clarification: 'border-amber-400/50 bg-amber-400/10 text-amber-100',
  insufficient: 'border-rose-500/50 bg-rose-500/10 text-rose-100',
};

function stripHtml(input: string): string {
  return input.replace(/<[^>]+>/g, ' ').replace(/\s+/g, ' ').trim();
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function formatStatusLabel(status?: string): string {
  if (!status) return 'No assessment yet';
  return status.replace(/_/g, ' ');
}

export function RiskCard({
  framework,
  risk,
  isEvaluating,
  isSelfCertifying,
  onEvaluate,
  onSelfCertify,
}: RiskCardProps) {
  const [expanded, setExpanded] = useState(false);
  const [completed, setCompleted] = useState(risk.completed);
  const [draftHtml, setDraftHtml] = useState(risk.justification);
  const [files, setFiles] = useState<File[]>([]);

  useEffect(() => {
    setCompleted(risk.completed);
    setDraftHtml(risk.justification);
  }, [risk.completed, risk.justification, risk.updatedAt]);

  const plainTextJustification = useMemo(() => stripHtml(draftHtml), [draftHtml]);

  const submitEvaluation = async () => {
    await onEvaluate({
      complianceCheckId: risk.complianceCheckId,
      completed,
      justification: plainTextJustification,
      files,
    });
    setFiles([]);
    setExpanded(true);
  };

  return (
    <article className="rounded-3xl border border-slate-800 bg-slate-900/70 p-5 shadow-[0_18px_60px_rgba(15,23,42,0.28)]">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div className="space-y-3">
          <div className="flex flex-wrap items-center gap-2">
            <span className={`rounded-full border px-3 py-1 text-xs font-semibold uppercase tracking-[0.22em] ${severityTone[risk.severity]}`}>
              {risk.severity}
            </span>
            <span className={`text-xs font-semibold uppercase tracking-[0.2em] ${priorityTone[risk.priority]}`}>
              {risk.priority} priority
            </span>
            <span className="text-xs uppercase tracking-[0.2em] text-slate-500">
              {framework.toUpperCase()}
            </span>
          </div>
          <div>
            <h2 className="text-xl font-semibold text-white">{risk.title}</h2>
            <p className="mt-2 max-w-3xl text-sm leading-6 text-slate-300">{risk.description}</p>
          </div>
        </div>

        <div className="flex flex-col items-stretch gap-2 sm:flex-row lg:flex-col lg:items-end">
          <button
            type="button"
            className="rounded-xl bg-emerald-400 px-4 py-2 text-sm font-semibold text-slate-950 transition hover:bg-emerald-300"
            onClick={() => setExpanded((current) => !current)}
          >
            {expanded ? 'Hide Risk Form' : 'Abate This Risk'}
          </button>
          <button
            type="button"
            className={`rounded-xl border px-4 py-2 text-sm font-semibold transition ${
              risk.selfCertified
                ? 'border-sky-400 bg-sky-400/10 text-sky-100 hover:bg-sky-400/20'
                : 'border-slate-700 text-slate-100 hover:border-sky-400 hover:text-white'
            } disabled:cursor-not-allowed disabled:opacity-60`}
            onClick={() =>
              onSelfCertify({
                actionItemId: risk.id,
                complianceCheckId: risk.complianceCheckId,
                selfCertified: !risk.selfCertified,
              })
            }
            disabled={isSelfCertifying}
          >
            {risk.selfCertified ? 'Resolved (Self-Certified)' : 'Mark Resolved (Self-Certified)'}
          </button>
        </div>
      </div>

      <div className="mt-4 grid gap-4 lg:grid-cols-[1.2fr_0.8fr]">
        <div className="rounded-2xl border border-slate-800/80 bg-slate-950/70 p-4">
          <h3 className="text-sm font-semibold uppercase tracking-[0.18em] text-slate-400">
            Acceptance criteria
          </h3>
          <ul className="mt-3 space-y-2 text-sm leading-6 text-slate-200">
            {risk.acceptanceCriteria.map((criterion) => (
              <li key={criterion} className="flex gap-2">
                <span className="mt-1 text-emerald-300">•</span>
                <span>{criterion}</span>
              </li>
            ))}
          </ul>
        </div>

        <div className="rounded-2xl border border-slate-800/80 bg-slate-950/70 p-4">
          <h3 className="text-sm font-semibold uppercase tracking-[0.18em] text-slate-400">
            Current assessment
          </h3>
          <div className="mt-3 flex flex-wrap gap-2">
            <span
              className={`rounded-full border px-3 py-1 text-xs font-semibold uppercase tracking-[0.2em] ${
                risk.engineStatus ? assessmentTone[risk.engineStatus] : 'border-slate-700 text-slate-300'
              }`}
            >
              {formatStatusLabel(risk.engineStatus)}
            </span>
            {risk.completed && (
              <span className="rounded-full border border-emerald-500/40 bg-emerald-500/10 px-3 py-1 text-xs font-semibold uppercase tracking-[0.2em] text-emerald-100">
                action completed
              </span>
            )}
          </div>
          <p className="mt-3 text-sm leading-6 text-slate-300">
            {risk.engineFeedback ?? 'Submit a justification to receive a reasoned recommendation.'}
          </p>
          {risk.criteriaChecked.length > 0 && (
            <div className="mt-3 flex flex-wrap gap-2">
              {risk.criteriaChecked.map((criterion) => (
                <span
                  key={criterion}
                  className="rounded-full border border-slate-700 bg-slate-900 px-3 py-1 text-xs text-slate-300"
                >
                  {criterion}
                </span>
              ))}
            </div>
          )}
        </div>
      </div>

      {risk.evidence.length > 0 && (
        <div className="mt-4 rounded-2xl border border-slate-800/80 bg-slate-950/60 p-4">
          <h3 className="text-sm font-semibold uppercase tracking-[0.18em] text-slate-400">
            Attached evidence
          </h3>
          <div className="mt-3 space-y-2">
            {risk.evidence.map((evidence) => (
              <a
                key={evidence.id}
                href={evidence.url?.startsWith('https://') || evidence.url?.startsWith('http://') ? evidence.url : '#'}
                rel="noopener noreferrer"
                className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-slate-800 bg-slate-900/70 px-3 py-2 text-sm text-slate-200 transition hover:border-emerald-400"
              >
                <span className="font-medium">{evidence.fileName}</span>
                <span className="text-xs uppercase tracking-[0.18em] text-slate-400">
                  {evidence.mimeType} · {formatBytes(evidence.sizeBytes)}
                </span>
              </a>
            ))}
          </div>
        </div>
      )}

      {expanded && (
        <div className="mt-5 rounded-3xl border border-slate-700 bg-slate-950/90 p-5">
          <div className="flex items-center gap-3">
            <input
              id={`completed-${risk.id}`}
              type="checkbox"
              checked={completed}
              onChange={(event) => setCompleted(event.target.checked)}
              className="h-4 w-4 rounded border-slate-600 bg-slate-900 text-emerald-400 focus:ring-emerald-400"
            />
            <label htmlFor={`completed-${risk.id}`} className="text-sm font-medium text-slate-100">
              I have completed the required action.
            </label>
          </div>

          <div className="mt-4">
            <RichTextInput
              value={draftHtml}
              onChange={setDraftHtml}
              placeholder="Describe what was done, attach evidence if available, and explain why this resolves the regulatory requirement. Example: 'Uploaded RECs for all three offices; certificates attached.'"
              disabled={isEvaluating}
            />
          </div>

          <div className="mt-4 rounded-2xl border border-dashed border-slate-700 bg-slate-950/80 p-4">
            <label className="block text-sm font-semibold text-slate-100">Evidence upload</label>
            <p className="mt-1 text-xs leading-5 text-slate-400">
              Attach PDF, CSV, or image evidence that supports the action you described.
            </p>
            <input
              type="file"
              accept=".pdf,.csv,image/*"
              multiple
              onChange={(event) => setFiles(Array.from(event.target.files ?? []))}
              className="mt-3 block w-full text-sm text-slate-300 file:mr-4 file:rounded-xl file:border-0 file:bg-slate-800 file:px-4 file:py-2 file:text-sm file:font-semibold file:text-white hover:file:bg-slate-700"
            />
            {files.length > 0 && (
              <ul className="mt-3 space-y-2 text-sm text-slate-200">
                {files.map((file) => (
                  <li key={`${file.name}-${file.size}`} className="flex items-center justify-between">
                    <span>{file.name}</span>
                    <span className="text-xs uppercase tracking-[0.18em] text-slate-500">
                      {formatBytes(file.size)}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </div>

          <div className="mt-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <p className="max-w-2xl text-xs leading-5 text-slate-500">
              OffGridFlow provides a reasoned recommendation only. You can still self-certify and continue your workplan even if the engine requests more detail.
            </p>
            <button
              type="button"
              onClick={submitEvaluation}
              disabled={isEvaluating}
              className="rounded-xl bg-white px-4 py-2 text-sm font-semibold text-slate-950 transition hover:bg-slate-200 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {isEvaluating ? 'Reviewing…' : 'Submit for Review'}
            </button>
          </div>
        </div>
      )}
    </article>
  );
}
