'use client';

import { useEffect, useState, useCallback } from 'react';
import Link from 'next/link';
import { api } from '../../../lib/api';
import type { CSRDComplianceResponse, ValidationInfo } from '../../../lib/types';
import { useRequireAuth } from '../../../lib/session';

const severityColors: Record<string, { bg: string; text: string }> = {
  error: { bg: '#7f1d1d', text: '#fecaca' },
  warning: { bg: '#78350f', text: '#fde68a' },
  info: { bg: '#1e3a8a', text: '#bfdbfe' },
};

export default function CSRDPage() {
  const session = useRequireAuth();
  const [report, setReport] = useState<CSRDComplianceResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [year, setYear] = useState(new Date().getFullYear());

  const downloadExport = useCallback(
    (format: 'pdf' | 'xbrl') => {
      if (!report) return;
      const params = new URLSearchParams({
        format,
        year: String(year),
      });
      window.open(`/api/compliance/export?${params.toString()}`, '_blank');
    },
    [year, report]
  );

  useEffect(() => {
    if (!session.isAuthenticated) return;

    const fetchReport = async () => {
      setLoading(true);
      setError(null);
      try {
        const data = await api.get<CSRDComplianceResponse>(`/api/compliance/csrd?year=${year}`);
        setReport(data);
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : 'Failed to load CSRD report');
      } finally {
        setLoading(false);
      }
    };

    fetchReport();
  }, [year, session.isAuthenticated]);

  const validation = report?.metrics?.validation as ValidationInfo | undefined;

  if (session.loading || !session.isAuthenticated) {
    return (
      <div style={{ padding: '2rem' }}>
        <h1>CSRD / ESRS E1 Report</h1>
        <p style={{ color: '#888' }}>Checking your session...</p>
      </div>
    );
  }

  if (loading) {
    return (
      <div style={{ padding: '2rem' }}>
        <h1>CSRD / ESRS E1 Report</h1>
        <p style={{ color: '#888' }}>Loading compliance data...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: '2rem' }}>
        <h1>CSRD / ESRS E1 Report</h1>
        <div
          style={{
            color: '#ff6b6b',
            padding: '1rem',
            background: '#1a1a2e',
            borderRadius: '8px',
          }}
        >
          Error loading report: {error}
        </div>
        <Link href="/" style={{ color: '#8aa9ff', marginTop: '1rem', display: 'inline-block' }}>
          ← Back to Dashboard
        </Link>
      </div>
    );
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <h1 style={{ margin: 0 }}>CSRD / ESRS E1 Report</h1>
          <p style={{ color: '#888', margin: '0.25rem 0 0 0', fontSize: '0.9rem' }}>
            European Sustainability Reporting Standards - Climate Change
          </p>
        </div>
        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <label htmlFor="year-select" style={{ color: '#888', fontSize: '0.9rem' }}>Year:</label>
          <select
            id="year-select"
            title="Select report year"
            value={year}
            onChange={(e) => setYear(Number(e.target.value))}
            style={{
              background: '#1d2940',
              color: '#fff',
              border: '1px solid #374151',
              borderRadius: '6px',
              padding: '0.5rem',
            }}
          >
            {[2024, 2023, 2022].map((y) => (
              <option key={y} value={y}>
                {y}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Status Banner */}
      <div
        style={{
          padding: '1rem',
          background: getStatusBackground(report?.status ?? 'incomplete'),
          borderRadius: '8px',
          marginBottom: '1.5rem',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <div>
          <div style={{ fontWeight: 600, fontSize: '1.1rem' }}>
            Report Status: {formatStatus(report?.status ?? 'incomplete')}
          </div>
          <div style={{ fontSize: '0.85rem', color: '#ccc', marginTop: '0.25rem' }}>
            Organization: {report?.orgId} | Standard: {report?.standard}
          </div>
        </div>
        <div style={{ fontSize: '0.8rem', color: '#888' }}>
          Generated: {report?.timestamp ? new Date(report.timestamp).toLocaleString() : 'N/A'}
        </div>
      </div>

      {/* Validation Messages */}
      {validation && (validation.errors.length > 0 || validation.warnings.length > 0) && (
        <div style={{ marginBottom: '1.5rem' }}>
          <h2 style={{ fontSize: '1.1rem', marginBottom: '0.75rem' }}>Validation Results</h2>
          <div style={{ display: 'grid', gap: '0.5rem' }}>
            {validation.errors.map((err, i) => (
              <ValidationMessage key={`err-${i}`} type="error" message={err} />
            ))}
            {validation.warnings.map((warn, i) => (
              <ValidationMessage key={`warn-${i}`} type="warning" message={warn} />
            ))}
          </div>
        </div>
      )}

      {/* GHG Emissions Summary */}
      <h2 style={{ fontSize: '1.1rem', marginBottom: '0.75rem' }}>GHG Emissions (ESRS E1-6)</h2>
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
          gap: '1rem',
          marginBottom: '1.5rem',
        }}
      >
        <MetricCard label="Scope 1 (Direct)" value={report?.totals.scope1Tons ?? 0} unit="tCO2e" />
        <MetricCard label="Scope 2 (Energy)" value={report?.totals.scope2Tons ?? 0} unit="tCO2e" />
        <MetricCard label="Scope 3 (Value Chain)" value={report?.totals.scope3Tons ?? 0} unit="tCO2e" />
        <MetricCard
          label="Total GHG Emissions"
          value={report?.totals.totalTons ?? 0}
          unit="tCO2e"
          highlight
        />
      </div>

      {/* ESRS E1 Disclosure Requirements */}
      <h2 style={{ fontSize: '1.1rem', marginBottom: '0.75rem' }}>ESRS E1 Disclosure Requirements</h2>
      <div style={{ display: 'grid', gap: '0.75rem', marginBottom: '1.5rem' }}>
        <DisclosureItem
          code="E1-1"
          title="Transition plan for climate change mitigation"
          status={hasData(report?.metrics?.['E1-1'])}
        />
        <DisclosureItem
          code="E1-2"
          title="Policies related to climate change mitigation and adaptation"
          status={hasData(report?.metrics?.['E1-2'])}
        />
        <DisclosureItem
          code="E1-3"
          title="Actions and resources in relation to climate change policies"
          status={hasData(report?.metrics?.['E1-3'])}
        />
        <DisclosureItem
          code="E1-4"
          title="Targets related to climate change mitigation and adaptation"
          status={hasData(report?.metrics?.['E1-4'])}
        />
        <DisclosureItem
          code="E1-5"
          title="Energy consumption and mix"
          status={hasData(report?.metrics?.['E1-5'])}
        />
        <DisclosureItem
          code="E1-6"
          title="Gross Scope 1, 2, 3 and Total GHG emissions"
          status={(report?.totals.totalTons ?? 0) > 0}
        />
        <DisclosureItem
          code="E1-7"
          title="GHG removals and GHG mitigation projects financed through carbon credits"
          status={hasData(report?.metrics?.['E1-7'])}
        />
        <DisclosureItem
          code="E1-8"
          title="Internal carbon pricing"
          status={hasData(report?.metrics?.['E1-8'])}
        />
        <DisclosureItem
          code="E1-9"
          title="Anticipated financial effects from climate-related physical and transition risks"
          status={hasData(report?.metrics?.['E1-9'])}
        />
      </div>

      {/* Actions */}
      <div style={{ display: 'flex', gap: '1rem', marginTop: '2rem' }}>
        <Link
          href="/emissions"
          style={{
            padding: '0.75rem 1.5rem',
            background: '#3b82f6',
            color: '#fff',
            borderRadius: '6px',
            textDecoration: 'none',
          }}
        >
          View Emissions Data
        </Link>
        <button
          onClick={() => downloadExport('pdf')}
          disabled={loading || !report}
          style={{
            padding: '0.75rem 1.5rem',
            background: '#1d2940',
            color: '#8aa9ff',
            border: '1px solid #374151',
            borderRadius: '6px',
            cursor: loading || !report ? 'not-allowed' : 'pointer',
          }}
        >
          Export PDF
        </button>
        <button
          onClick={() => downloadExport('xbrl')}
          disabled={loading || !report}
          style={{
            padding: '0.75rem 1.5rem',
            background: '#0f172a',
            color: '#a5b4fc',
            border: '1px solid #374151',
            borderRadius: '6px',
            cursor: loading || !report ? 'not-allowed' : 'pointer',
          }}
        >
          Download XBRL
        </button>
        <Link
          href="/"
          style={{
            padding: '0.75rem 1.5rem',
            background: 'transparent',
            color: '#8aa9ff',
            borderRadius: '6px',
            textDecoration: 'none',
          }}
        >
          ← Dashboard
        </Link>
      </div>
    </div>
  );
}

// Helper functions
function getStatusBackground(status: string): string {
  switch (status) {
    case 'ok':
      return '#064e3b';
    case 'warnings':
      return '#78350f';
    case 'incomplete':
    default:
      return '#7f1d1d';
  }
}

function formatStatus(status: string): string {
  switch (status) {
    case 'ok':
      return 'Complete ✓';
    case 'warnings':
      return 'Complete with Warnings';
    case 'incomplete':
    default:
      return 'Incomplete';
  }
}

function hasData(value: unknown): boolean {
  if (value === null || value === undefined) return false;
  if (typeof value === 'object' && Object.keys(value).length === 0) return false;
  return true;
}

// Components
function ValidationMessage({ type, message }: { type: 'error' | 'warning' | 'info'; message: string }) {
  const colors = severityColors[type];
  return (
    <div
      style={{
        padding: '0.75rem 1rem',
        background: colors.bg,
        color: colors.text,
        borderRadius: '6px',
        fontSize: '0.9rem',
      }}
    >
      <span style={{ fontWeight: 600, marginRight: '0.5rem' }}>
        {type === 'error' ? '✕' : type === 'warning' ? '⚠' : 'ℹ'}
      </span>
      {message}
    </div>
  );
}

function MetricCard({
  label,
  value,
  unit,
  highlight,
}: {
  label: string;
  value: number;
  unit: string;
  highlight?: boolean;
}) {
  return (
    <div
      style={{
        padding: '1rem',
        border: `1px solid ${highlight ? '#3b82f6' : '#1d2940'}`,
        borderRadius: '12px',
        background: highlight ? 'rgba(59, 130, 246, 0.1)' : 'transparent',
      }}
    >
      <div style={{ color: '#8aa9ff', fontSize: '0.85rem', marginBottom: '0.25rem' }}>{label}</div>
      <div style={{ fontSize: '1.4rem', fontWeight: 700 }}>
        {value.toLocaleString(undefined, { maximumFractionDigits: 2 })}
        <span style={{ fontSize: '0.9rem', fontWeight: 400, marginLeft: '0.25rem' }}>{unit}</span>
      </div>
    </div>
  );
}

function DisclosureItem({
  code,
  title,
  status,
}: {
  code: string;
  title: string;
  status: boolean;
}) {
  return (
    <div
      style={{
        padding: '0.75rem 1rem',
        border: '1px solid #1d2940',
        borderRadius: '8px',
        display: 'flex',
        alignItems: 'center',
        gap: '1rem',
      }}
    >
      <span
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          justifyContent: 'center',
          width: '24px',
          height: '24px',
          borderRadius: '50%',
          background: status ? '#064e3b' : '#374151',
          color: status ? '#4ade80' : '#9ca3af',
          fontSize: '0.75rem',
        }}
      >
        {status ? '✓' : '○'}
      </span>
      <div>
        <span style={{ color: '#8aa9ff', fontWeight: 600, marginRight: '0.5rem' }}>{code}</span>
        <span style={{ color: status ? '#e5e7eb' : '#9ca3af' }}>{title}</span>
      </div>
    </div>
  );
}
