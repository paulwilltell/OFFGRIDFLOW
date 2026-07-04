'use client';

import { useEffect, useMemo, useState } from 'react';
import { useRouter } from 'next/navigation';
import { api } from '@/lib/api';
import { useRequireAuth } from '@/lib/session';
import type { Scope2Summary, Scope2Emission, PaginatedResponse } from '@/lib/types';

type AnomalySummary = { open: number; critical: number; warning: number };

// Aggregate Scope 1/2/3 totals from GET /api/emissions (free review endpoint).
type EmissionsTotals = { scope1Tons: number; scope2Tons: number; scope3Tons: number; totalTons: number };

function fmt(n: number): string {
  return n.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function pctOf(part: number, whole: number): string {
  if (whole <= 0) return '0%';
  return `${Math.round((part / whole) * 100)}% of total`;
}

export default function ReviewDashboard() {
  const session = useRequireAuth();
  const router = useRouter();
  const [summary, setSummary] = useState<Scope2Summary | null>(null);
  const [emissions, setEmissions] = useState<Scope2Emission[]>([]);
  const [totals, setTotals] = useState<EmissionsTotals | null>(null);
  const [anomalies, setAnomalies] = useState<AnomalySummary | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    (async () => {
      try {
        const [sum, list, agg] = await Promise.all([
          api.get<Scope2Summary>('/api/emissions/scope2/summary').catch(() => null),
          api.get<PaginatedResponse<Scope2Emission>>('/api/emissions/scope2?limit=100').catch(() => null),
          api.get<EmissionsTotals>('/api/emissions').catch(() => null),
        ]);
        if (!active) return;
        setSummary(sum);
        setEmissions(list?.data ?? []);
        setTotals(agg);
      } finally {
        if (active) setLoading(false);
      }
    })();
    // Anomaly summary is best-effort — drives the data-quality badge.
    fetch('/api/audit/anomalies?status=open', { credentials: 'include' })
      .then((r) => (r.ok ? r.json() : null))
      .then((d) => active && d?.summary && setAnomalies({ open: d.summary.open ?? 0, critical: d.summary.critical ?? 0, warning: d.summary.warning ?? 0 }))
      .catch(() => {});
    return () => { active = false; };
  }, []);

  const scope1 = totals?.scope1Tons ?? 0;
  const scope2 = totals?.scope2Tons ?? summary?.totalEmissionsTonsCO2e ?? 0;
  const scope3 = totals?.scope3Tons ?? 0;
  const total = totals?.totalTons ?? (summary?.totalEmissionsTonsCO2e ?? 0);
  const hasData = total > 0 || emissions.length > 0;

  // Aggregate emissions by meter for the "by source" bar chart.
  const bySource = useMemo(() => {
    const map = new Map<string, number>();
    for (const e of emissions) {
      const key = e.meterId || e.location || 'unknown';
      map.set(key, (map.get(key) ?? 0) + e.emissionsTonsCO2e);
    }
    const rows = [...map.entries()].map(([name, tonnes]) => ({ name, tonnes })).sort((a, b) => b.tonnes - a.tonnes);
    const max = rows.length ? rows[0].tonnes : 1;
    return { rows: rows.slice(0, 6), max };
  }, [emissions]);

  const anomaliesClean = !anomalies || anomalies.open === 0;

  if (!session?.isAuthenticated) return null;

  if (!loading && !hasData) {
    return (
      <div className="mx-auto max-w-[520px] py-16 text-center">
        <div className="mx-auto mb-5 flex h-[58px] w-[58px] items-center justify-center rounded-[14px] text-[26px]" style={{ background: '#e8f0ea', color: '#2f6b50' }}>📊</div>
        <h1 className="mb-2 text-[22px] font-bold tracking-[-0.02em]">No data yet</h1>
        <p className="mb-6 text-[15px]" style={{ color: '#6a7a71' }}>Upload your emissions data and we&apos;ll calculate your footprint here.</p>
        <button onClick={() => router.push('/emissions')} className="inline-flex h-[44px] items-center rounded-[9px] px-6 text-[14.5px] font-semibold text-white" style={{ background: '#1d3b2e' }}>
          Upload data →
        </button>
      </div>
    );
  }

  return (
    <div>
      {/* Header row */}
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <div className="mb-[6px] text-[13px]" style={{ color: '#8a978f' }}>Total footprint · FY{new Date().getFullYear()}</div>
          <div className="flex items-baseline gap-[10px]">
            <span className="text-[46px] font-medium tracking-[-0.02em]" style={{ fontFamily: "'IBM Plex Mono', monospace" }}>{loading ? '—' : fmt(total)}</span>
            <span className="text-[16px]" style={{ color: '#6a7a71' }}>tCO₂e</span>
          </div>
        </div>
        <div className="flex items-center gap-[10px]">
          <span className="flex items-center gap-[7px] rounded-[8px] px-[12px] py-[7px] text-[13px] font-semibold"
            style={anomaliesClean ? { background: '#e8f0ea', color: '#2f6b50' } : { background: '#fbf3e8', color: '#a86a24' }}>
            <span className="h-[7px] w-[7px] rounded-full" style={{ background: anomaliesClean ? '#2f6b50' : '#d9922e' }} />
            {anomaliesClean ? 'Data clean — no anomalies' : `${anomalies!.open} anomal${anomalies!.open === 1 ? 'y' : 'ies'} to review`}
          </span>
          <button onClick={() => router.push('/reports')} className="flex h-[38px] items-center gap-[7px] rounded-[9px] px-[18px] text-[14px] font-semibold text-white" style={{ background: '#1d3b2e' }}>
            Generate report →
          </button>
        </div>
      </div>

      {/* Scope cards */}
      <div className="mb-[18px] grid grid-cols-1 gap-4 md:grid-cols-3">
        <ScopeCard dot="#234e3b" label="Scope 1 · Direct" value={loading ? '—' : fmt(scope1)} muted={scope1 <= 0}
          note={scope1 > 0 ? pctOf(scope1, total) : 'No fuel data yet'} highlight={scope1 > 0} />
        <ScopeCard dot="#4f8f6e" label="Scope 2 · Energy" value={loading ? '—' : fmt(scope2)} muted={scope2 <= 0}
          note={scope2 > 0 ? `${pctOf(scope2, total)} · ${summary?.activityCount ?? emissions.length} sources` : 'No energy data yet'} highlight={scope2 > 0} />
        <ScopeCard dot="#cdab6e" label="Scope 3 · Value chain" value={loading ? '—' : fmt(scope3)} muted={scope3 <= 0}
          note={scope3 > 0 ? pctOf(scope3, total) : 'No travel/supplier data yet'} highlight={scope3 > 0} />
      </div>

      {/* Breakdown */}
      <div className="flex flex-col gap-4 lg:flex-row">
        <div className="flex-[1.5] rounded-[13px] border bg-white p-[22px]" style={{ borderColor: '#e8ece8' }}>
          <div className="mb-[18px] text-[13px] font-semibold" style={{ color: '#3f4f47' }}>Emissions by source</div>
          <div className="flex flex-col gap-[15px]">
            {bySource.rows.length === 0 && <div className="text-[13px]" style={{ color: '#9aa79f' }}>No sources yet.</div>}
            {bySource.rows.map((r) => (
              <div key={r.name} className="flex items-center gap-[14px]">
                <span className="w-[140px] truncate text-[11.5px]" style={{ fontFamily: "'IBM Plex Mono', monospace", color: '#5b6b62' }}>{r.name}</span>
                <span className="h-[14px] flex-1 overflow-hidden rounded-[7px]" style={{ background: '#f0f3f0' }}>
                  <span className="block h-full rounded-[7px]" style={{ width: `${Math.max(4, (r.tonnes / bySource.max) * 100)}%`, background: '#4f8f6e' }} />
                </span>
                <span className="w-[74px] text-right text-[12.5px]" style={{ fontFamily: "'IBM Plex Mono', monospace" }}>{fmt(r.tonnes)}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="flex-1 rounded-[13px] border bg-white p-[22px]" style={{ borderColor: '#e8ece8' }}>
          <div className="mb-4 flex items-center justify-between">
            <span className="text-[13px] font-semibold" style={{ color: '#3f4f47' }}>Data sources</span>
            <button onClick={() => router.push('/emissions')} className="text-[12.5px] font-semibold" style={{ color: '#2f6b50' }}>Add more</button>
          </div>
          <SourceRow active label={`${summary?.activityCount ?? emissions.length} utility meters`} status="Active" />
          <SourceRow active={scope1 > 0} label="Fuel & fleet (Scope 1)" status={scope1 > 0 ? 'Active' : 'No data yet'} />
          <SourceRow active={scope3 > 0} label="Travel & suppliers (Scope 3)" status={scope3 > 0 ? 'Active' : 'No data yet'} last />
        </div>
      </div>
    </div>
  );
}

function ScopeCard({ dot, label, value, note, muted, highlight }: { dot: string; label: string; value: string; note: string; muted?: boolean; highlight?: boolean }) {
  return (
    <div className="rounded-[13px] border bg-white p-5" style={{ borderColor: '#e8ece8' }}>
      <div className="mb-[14px] flex items-center gap-2">
        <span className="h-[9px] w-[9px] rounded-full" style={{ background: dot }} />
        <span className="text-[13px] font-semibold" style={{ color: '#3f4f47' }}>{label}</span>
      </div>
      <div className="text-[26px]" style={{ fontFamily: "'IBM Plex Mono', monospace", color: muted ? '#9aa79f' : '#16201b' }}>
        {value}<span className="ml-[5px] text-[13px]" style={{ color: muted ? '#9aa79f' : '#6a7a71' }}>tCO₂e</span>
      </div>
      <div className="mt-[6px] text-[12.5px]" style={{ color: highlight ? '#4f8f6e' : muted ? '#b3bdb6' : '#8a978f', fontWeight: highlight ? 500 : 400 }}>{note}</div>
    </div>
  );
}

function SourceRow({ label, status, active, last }: { label: string; status: string; active?: boolean; last?: boolean }) {
  return (
    <div className="flex items-center justify-between py-[9px]" style={{ borderBottom: last ? 'none' : '1px solid #f2f4f1' }}>
      <span className="flex items-center gap-[9px] text-[13px]" style={{ color: active ? '#16201b' : '#9aa79f' }}>
        <span className="h-[7px] w-[7px] rounded-full" style={{ background: active ? '#4f8f6e' : '#d4dbd6' }} />
        {label}
      </span>
      <span className="text-[12px]" style={{ color: active ? '#6a7a71' : '#b3bdb6', fontFamily: active ? "'IBM Plex Mono', monospace" : 'inherit' }}>{status}</span>
    </div>
  );
}
