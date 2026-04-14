import { api } from './api';
import type { ComplianceSummary, FrameworkStatus } from './types';

type RawFrameworkStatus = {
  name?: string;
  status?: string;
  scope1Ready?: boolean;
  scope2Ready?: boolean;
  scope3Ready?: boolean;
  scope1_ready?: boolean;
  scope2_ready?: boolean;
  scope3_ready?: boolean;
};

type RawComplianceSummary = {
  frameworks?: Record<string, RawFrameworkStatus>;
  totals?: Record<string, unknown>;
  timestamp?: string;
};

const CURRENT_YEAR = new Date().getFullYear();
const SUMMARY_YEAR_CANDIDATES = [CURRENT_YEAR, CURRENT_YEAR - 1, CURRENT_YEAR - 2, CURRENT_YEAR - 3];

function numberValue(value: unknown): number {
  return typeof value === 'number' && Number.isFinite(value) ? value : 0;
}

function normalizeFramework(name: string, raw?: RawFrameworkStatus): FrameworkStatus {
  return {
    name: raw?.name ?? name,
    status: (raw?.status as FrameworkStatus['status']) ?? 'not_started',
    scope1Ready: Boolean(raw?.scope1Ready ?? raw?.scope1_ready),
    scope2Ready: Boolean(raw?.scope2Ready ?? raw?.scope2_ready),
    scope3Ready: Boolean(raw?.scope3Ready ?? raw?.scope3_ready),
  };
}

export function normalizeComplianceSummary(raw: unknown): ComplianceSummary {
  const payload = (raw ?? {}) as RawComplianceSummary;
  const totals = payload.totals ?? {};

  return {
    frameworks: {
      csrd: normalizeFramework('CSRD / ESRS E1', payload.frameworks?.csrd),
      sec: normalizeFramework('SEC Climate Disclosure', payload.frameworks?.sec),
      cbam: normalizeFramework('EU CBAM', payload.frameworks?.cbam),
      california: normalizeFramework('California Climate Disclosure', payload.frameworks?.california),
    },
    totals: {
      scope1Tons: numberValue(totals.scope1Tons ?? totals.Scope1Tons),
      scope2Tons: numberValue(totals.scope2Tons ?? totals.Scope2Tons),
      scope3Tons: numberValue(totals.scope3Tons ?? totals.Scope3Tons),
    },
    timestamp: payload.timestamp ?? new Date().toISOString(),
  };
}

function totalTons(summary: ComplianceSummary): number {
  return summary.totals.scope1Tons + summary.totals.scope2Tons + summary.totals.scope3Tons;
}

export async function fetchLatestComplianceSummary(): Promise<{ summary: ComplianceSummary; year: number }> {
  let fallbackSummary: ComplianceSummary | null = null;

  for (const year of SUMMARY_YEAR_CANDIDATES) {
    const raw = await api.get<unknown>(`/api/compliance/summary?year=${year}`);
    const summary = normalizeComplianceSummary(raw);

    if (!fallbackSummary) {
      fallbackSummary = summary;
    }

    if (totalTons(summary) > 0) {
      return { summary, year };
    }
  }

  return {
    summary: fallbackSummary ?? normalizeComplianceSummary({}),
    year: SUMMARY_YEAR_CANDIDATES[0],
  };
}
