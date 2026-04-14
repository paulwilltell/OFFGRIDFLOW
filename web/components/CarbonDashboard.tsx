'use client';

import React, { useEffect, useState, useCallback, useMemo, Suspense, memo } from 'react';
import dynamic from 'next/dynamic';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useCarbonStore } from '@/stores/carbonStore';
import { useCompliance } from '@/hooks/useCompliance';
import RealTimeProvider, { useRealTime } from '@/providers/RealTimeDataProvider';
import { LoadingSkeleton, DashboardSkeleton } from '@/components/ui/LoadingSkeleton';
import { CarbonApi, downloadFile, formatNumber } from '@/lib/api/carbon';
import { EmissionData, DataSource, Timeframe, ComplianceStatusType } from '@/types/carbon';
import ErrorBoundary from '@/components/ErrorBoundary';

const EmissionChart = dynamic(() => import('@/components/charts/EmissionChartJS'), {
  loading: () => <LoadingSkeleton type="chart" />,
  ssr: false,
});

// ============================================================================
// Metric Card
// ============================================================================

function MetricCard({
  label,
  value,
  unit,
  change,
  changeLabel,
  quality,
}: {
  label: string;
  value: string | number;
  unit?: string;
  change?: number;
  changeLabel?: string;
  quality?: 'measured' | 'estimated' | 'calculated';
}) {
  return (
    <div className="rounded-xl border border-gray-800 bg-gray-800/40 p-5">
      <div className="flex items-start justify-between">
        <span className="text-xs font-medium uppercase tracking-wider text-gray-500">{label}</span>
        {quality && (
          <span className={`rounded px-1.5 py-0.5 text-[10px] font-medium ${
            quality === 'measured' ? 'bg-green-500/10 text-green-400' :
            quality === 'estimated' ? 'bg-yellow-500/10 text-yellow-400' :
            'bg-blue-500/10 text-blue-400'
          }`}>
            {quality}
          </span>
        )}
      </div>
      <div className="mt-2 flex items-baseline gap-1.5">
        <span className="text-2xl font-bold text-white">{typeof value === 'number' ? formatNumber(value) : value}</span>
        {unit && <span className="text-sm text-gray-500">{unit}</span>}
      </div>
      {change !== undefined && (
        <div className="mt-2 flex items-center gap-1 text-xs">
          <span className={change < 0 ? 'text-green-400' : change > 0 ? 'text-red-400' : 'text-gray-500'}>
            {change > 0 ? '+' : ''}{change.toFixed(1)}%
          </span>
          {changeLabel && <span className="text-gray-600">{changeLabel}</span>}
        </div>
      )}
    </div>
  );
}

// ============================================================================
// Scope Breakdown
// ============================================================================

function ScopeBreakdown({ scope1, scope2, scope3 }: { scope1: number; scope2: number; scope3: number }) {
  const total = scope1 + scope2 + scope3;
  if (total === 0) return null;

  const scopes = [
    { name: 'Scope 1', label: 'Direct', value: scope1, pct: (scope1 / total) * 100, color: 'bg-red-500', text: 'text-red-400' },
    { name: 'Scope 2', label: 'Energy', value: scope2, pct: (scope2 / total) * 100, color: 'bg-amber-500', text: 'text-amber-400' },
    { name: 'Scope 3', label: 'Value Chain', value: scope3, pct: (scope3 / total) * 100, color: 'bg-blue-500', text: 'text-blue-400' },
  ];

  return (
    <div className="rounded-xl border border-gray-800 bg-gray-800/40 p-5">
      <h3 className="mb-4 text-xs font-medium uppercase tracking-wider text-gray-500">Emissions by Scope</h3>

      {/* Stacked bar */}
      <div className="mb-5 flex h-3 overflow-hidden rounded-full bg-gray-700">
        {scopes.map((s) => (
          <div key={s.name} className={`${s.color} transition-all duration-700`} style={{ width: `${s.pct}%` }} />
        ))}
      </div>

      {/* Scope rows */}
      <div className="space-y-3">
        {scopes.map((s) => (
          <div key={s.name} className="flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <div className={`h-2.5 w-2.5 rounded-full ${s.color}`} />
              <div>
                <span className="text-sm font-medium text-white">{s.name}</span>
                <span className="ml-1.5 text-xs text-gray-600">{s.label}</span>
              </div>
            </div>
            <div className="text-right">
              <span className="text-sm font-semibold text-white">{formatNumber(s.value)}</span>
              <span className="ml-1.5 text-xs text-gray-500">{s.pct.toFixed(1)}%</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ============================================================================
// Compliance Progress Tracker
// ============================================================================

function ComplianceTracker({ status }: { status: Record<string, ComplianceStatusType> }) {
  const frameworks = [
    { key: 'csrd', name: 'CSRD / ESRS', region: 'EU', href: '/compliance/csrd' },
    { key: 'sec', name: 'SEC Climate', region: 'US', href: '/compliance/sec' },
    { key: 'sb253', name: 'SB 253', region: 'CA', href: '/compliance/california' },
    { key: 'cbam', name: 'CBAM', region: 'EU', href: '/compliance/cbam' },
    { key: 'ifrs', name: 'IFRS S2', region: 'Global', href: '/compliance/csrd' },
  ];

  const statusConfig: Record<string, { label: string; bg: string; text: string }> = {
    complete: { label: 'Complete', bg: 'bg-green-500/10', text: 'text-green-400' },
    in_progress: { label: 'In Progress', bg: 'bg-blue-500/10', text: 'text-blue-400' },
    pending: { label: 'Pending', bg: 'bg-gray-500/10', text: 'text-gray-400' },
    at_risk: { label: 'At Risk', bg: 'bg-amber-500/10', text: 'text-amber-400' },
    overdue: { label: 'Overdue', bg: 'bg-red-500/10', text: 'text-red-400' },
  };

  return (
    <div className="rounded-xl border border-gray-800 bg-gray-800/40 p-5">
      <h3 className="mb-4 text-xs font-medium uppercase tracking-wider text-gray-500">Compliance Status</h3>
      <div className="space-y-2.5">
        {frameworks.map((fw) => {
          const s = statusConfig[status[fw.key] || 'pending'] || statusConfig.pending;
          return (
            <Link
              key={fw.key}
              href={fw.href}
              className="flex items-center justify-between rounded-lg border border-gray-800/50 bg-gray-900/30 px-4 py-2.5 transition hover:border-gray-700 hover:bg-gray-900/60"
            >
              <div className="flex items-center gap-3">
                <span className="text-sm font-medium text-white">{fw.name}</span>
                <span className="text-[10px] text-gray-600">{fw.region}</span>
              </div>
              <span className={`rounded px-2 py-0.5 text-[11px] font-medium ${s.bg} ${s.text}`}>
                {s.label}
              </span>
            </Link>
          );
        })}
      </div>
    </div>
  );
}

// ============================================================================
// Activity Ledger (Recent Emissions Data)
// ============================================================================

function ActivityLedger({ dataSources }: { dataSources: DataSource[] }) {
  if (!dataSources || dataSources.length === 0) return null;

  return (
    <div className="rounded-xl border border-gray-800 bg-gray-800/40 p-5">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-xs font-medium uppercase tracking-wider text-gray-500">Data Sources</h3>
        <Link href="/settings/data-sources" className="text-xs text-primary-400 hover:underline">
          Manage
        </Link>
      </div>
      <div className="space-y-2">
        {dataSources.slice(0, 6).map((ds) => (
          <div key={ds.id} className="flex items-center justify-between rounded-lg bg-gray-900/30 px-3 py-2">
            <div className="flex items-center gap-2.5">
              <span className={`h-2 w-2 rounded-full ${
                ds.status === 'active' ? 'bg-green-500' : ds.status === 'error' ? 'bg-red-500' : 'bg-gray-500'
              }`} />
              <span className="text-sm text-white">{ds.name}</span>
              <span className="text-[10px] uppercase text-gray-600">{ds.type}</span>
            </div>
            <div className="text-right">
              <span className="text-xs text-gray-400">{formatNumber(ds.emissions)} tCO2e</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ============================================================================
// Quick Actions
// ============================================================================

function QuickActions() {
  return (
    <div className="rounded-xl border border-gray-800 bg-gray-800/40 p-5">
      <h3 className="mb-4 text-xs font-medium uppercase tracking-wider text-gray-500">Quick Actions</h3>
      <div className="grid grid-cols-2 gap-2">
        {[
          { label: 'Upload Data', href: '/emissions', icon: '↑' },
          { label: 'Generate Report', href: '/compliance/csrd', icon: '📋' },
          { label: 'Data Sources', href: '/settings/data-sources', icon: '🔗' },
          { label: 'Billing', href: '/settings/billing', icon: '💳' },
        ].map((action) => (
          <Link
            key={action.label}
            href={action.href}
            className="flex items-center gap-2 rounded-lg border border-gray-800/50 bg-gray-900/30 px-3 py-2.5 text-xs font-medium text-gray-300 transition hover:border-gray-700 hover:text-white"
          >
            <span>{action.icon}</span>
            {action.label}
          </Link>
        ))}
      </div>
    </div>
  );
}

// ============================================================================
// Role View Switcher (Panel 2A: Role-based views)
// ============================================================================

type DashboardRole = 'executive' | 'operator' | 'auditor';

function RoleViewSwitcher({ role, onRoleChange }: { role: DashboardRole; onRoleChange: (r: DashboardRole) => void }) {
  const roles: { key: DashboardRole; label: string; desc: string }[] = [
    { key: 'executive', label: 'Executive', desc: 'KPIs, trends, compliance status' },
    { key: 'operator', label: 'Operator', desc: 'Data entry, quality, connectors' },
    { key: 'auditor', label: 'Auditor', desc: 'Ledger, approvals, evidence trail' },
  ];

  return (
    <div className="flex items-center gap-1 rounded-lg border border-gray-800 bg-gray-900/50 p-1">
      {roles.map((r) => (
        <button
          key={r.key}
          onClick={() => onRoleChange(r.key)}
          title={r.desc}
          className={`rounded-md px-3 py-1.5 text-xs font-medium transition ${
            role === r.key
              ? 'bg-primary-600/20 text-primary-400'
              : 'text-gray-500 hover:text-gray-300'
          }`}
        >
          {r.label}
        </button>
      ))}
    </div>
  );
}

// ============================================================================
// Data Quality Summary Widget (Panel 2B: Data state indicators)
// ============================================================================

function DataQualitySummary() {
  const [counts, setCounts] = useState<Record<string, number> | null>(null);

  useEffect(() => {
    fetch('/api/audit/anomalies?status=open', { credentials: 'include' })
      .then((r) => r.json())
      .then((d) => setCounts(d.summary))
      .catch(() => {});
  }, []);

  if (!counts) return null;
  const hasIssues = counts.open > 0;

  return (
    <Link
      href="/audit/data-quality"
      className={`rounded-xl border p-4 transition hover:border-gray-600 ${
        counts.critical > 0
          ? 'border-red-800/60 bg-red-900/10'
          : counts.warning > 0
          ? 'border-yellow-800/60 bg-yellow-900/10'
          : 'border-gray-800 bg-gray-800/40'
      }`}
    >
      <div className="flex items-center justify-between">
        <h3 className="text-xs font-medium uppercase tracking-wider text-gray-500">Data Quality</h3>
        {!hasIssues && <span className="rounded bg-green-500/10 px-2 py-0.5 text-[10px] font-medium text-green-400">Clean</span>}
      </div>
      {hasIssues ? (
        <div className="mt-2 flex items-center gap-3">
          {counts.critical > 0 && (
            <span className="text-sm font-semibold text-red-400">{counts.critical} critical</span>
          )}
          {counts.warning > 0 && (
            <span className="text-sm font-semibold text-yellow-400">{counts.warning} warnings</span>
          )}
          <span className="text-xs text-gray-500">{counts.open} open</span>
        </div>
      ) : (
        <p className="mt-2 text-sm font-semibold text-green-400">No anomalies detected</p>
      )}
    </Link>
  );
}

// ============================================================================
// Alert Summary Widget (Panel 2C: Alert visibility in working surface)
// ============================================================================

function AlertSummary() {
  const [counts, setCounts] = useState<Record<string, number> | null>(null);

  useEffect(() => {
    fetch('/api/audit/alerts?status=open', { credentials: 'include' })
      .then((r) => r.json())
      .then((d) => setCounts(d.summary))
      .catch(() => {});
  }, []);

  if (!counts || counts.total === 0) return null;

  return (
    <Link
      href="/audit/alerts"
      className={`rounded-xl border p-4 transition hover:border-gray-600 ${
        counts.critical_active > 0 ? 'border-red-800/60 bg-red-900/10' : 'border-gray-800 bg-gray-800/40'
      }`}
    >
      <h3 className="text-xs font-medium uppercase tracking-wider text-gray-500">Open Alerts</h3>
      <div className="mt-2 flex items-center gap-3">
        <span className="text-lg font-bold text-white">{counts.open + counts.in_progress + counts.escalated}</span>
        <span className="text-xs text-gray-500">active</span>
        {counts.critical_active > 0 && (
          <span className="rounded bg-red-500/10 px-2 py-0.5 text-[10px] font-medium text-red-400">{counts.critical_active} critical</span>
        )}
      </div>
    </Link>
  );
}

// ============================================================================
// Auditor View (Panel 2A: Role-specific view)
// ============================================================================

function AuditorView() {
  const [ledger, setLedger] = useState<any[]>([]);
  const [approvals, setApprovals] = useState<any[]>([]);

  useEffect(() => {
    Promise.all([
      fetch('/api/audit/ledger?limit=10', { credentials: 'include' }).then(r => r.json()),
      fetch('/api/audit/approvals', { credentials: 'include' }).then(r => r.json()),
    ]).then(([l, a]) => {
      setLedger(l.entries || []);
      setApprovals(a.approvals || []);
    }).catch(() => {});
  }, []);

  return (
    <div className="space-y-6">
      <div className="grid gap-4 lg:grid-cols-3">
        <MetricCard label="Ledger Records" value={ledger.length} />
        <MetricCard label="Pending Approvals" value={approvals.filter(a => a.status === 'submitted' || a.status === 'reviewed').length} />
        <MetricCard label="Approved" value={approvals.filter(a => a.status === 'approved').length} />
      </div>

      <div className="rounded-xl border border-gray-800 bg-gray-800/40 p-5">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-xs font-medium uppercase tracking-wider text-gray-500">Recent Calculations</h3>
          <Link href="/audit/ledger" className="text-xs text-primary-400 hover:underline">View All</Link>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-800 text-left text-xs text-gray-500">
                <th className="pb-2 pr-4">Scope</th>
                <th className="pb-2 pr-4">Category</th>
                <th className="pb-2 pr-4">Quantity</th>
                <th className="pb-2 pr-4">Factor</th>
                <th className="pb-2 pr-4">Result (tCO2e)</th>
                <th className="pb-2 pr-4">Status</th>
                <th className="pb-2">Date</th>
              </tr>
            </thead>
            <tbody>
              {ledger.slice(0, 8).map((entry) => (
                <tr key={entry.id} className="border-b border-gray-800/30">
                  <td className="py-2 pr-4 text-white">{entry.scope}</td>
                  <td className="py-2 pr-4 text-gray-300">{entry.category || '-'}</td>
                  <td className="py-2 pr-4 text-gray-300">{entry.quantity} {entry.unit}</td>
                  <td className="py-2 pr-4 text-gray-400">{entry.emission_factor_value}</td>
                  <td className="py-2 pr-4 font-medium text-white">{entry.result_tonnes_co2e?.toFixed(4)}</td>
                  <td className="py-2 pr-4">
                    <span className={`rounded px-1.5 py-0.5 text-[10px] font-medium ${
                      entry.is_locked ? 'bg-green-500/10 text-green-400' : 'bg-yellow-500/10 text-yellow-400'
                    }`}>
                      {entry.is_locked ? 'Locked' : 'Draft'}
                    </span>
                  </td>
                  <td className="py-2 text-gray-500 text-xs">{new Date(entry.calculated_at).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="rounded-xl border border-gray-800 bg-gray-800/40 p-5">
          <div className="mb-4 flex items-center justify-between">
            <h3 className="text-xs font-medium uppercase tracking-wider text-gray-500">Approval Workflow</h3>
            <Link href="/audit/approvals" className="text-xs text-primary-400 hover:underline">View All</Link>
          </div>
          <div className="space-y-2">
            {approvals.slice(0, 5).map((a) => (
              <div key={a.id} className="flex items-center justify-between rounded-lg bg-gray-900/30 px-3 py-2">
                <div>
                  <span className="text-sm text-white">{a.entity_type}</span>
                  <span className="ml-2 text-[10px] text-gray-500">{a.entity_id?.slice(0, 8)}</span>
                </div>
                <span className={`rounded px-2 py-0.5 text-[10px] font-medium ${
                  a.status === 'approved' ? 'bg-green-500/10 text-green-400' :
                  a.status === 'rejected' ? 'bg-red-500/10 text-red-400' :
                  a.status === 'submitted' ? 'bg-blue-500/10 text-blue-400' :
                  'bg-gray-500/10 text-gray-400'
                }`}>
                  {a.status}
                </span>
              </div>
            ))}
          </div>
        </div>

        <div className="space-y-4">
          <DataQualitySummary />
          <Link
            href="/audit/factor-snapshots"
            className="block rounded-xl border border-gray-800 bg-gray-800/40 p-4 transition hover:border-gray-600"
          >
            <h3 className="text-xs font-medium uppercase tracking-wider text-gray-500">Factor Snapshots</h3>
            <p className="mt-2 text-sm text-gray-300">Lock emission factors to reporting periods for reproducibility.</p>
          </Link>
        </div>
      </div>
    </div>
  );
}

// ============================================================================
// Operator View (Panel 2A: Role-specific view)
// ============================================================================

function OperatorView({ emissions, dataSources, timeframe, tenantId }: {
  emissions: EmissionData;
  dataSources: DataSource[];
  timeframe: Timeframe;
  tenantId: string;
}) {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 lg:grid-cols-4">
        <MetricCard label="Total Sources" value={dataSources?.length || 0} />
        <MetricCard label="Active Sources" value={dataSources?.filter(ds => ds.status === 'active').length || 0} />
        <MetricCard label="Error Sources" value={dataSources?.filter(ds => ds.status === 'error').length || 0} />
        <MetricCard label="Last Upload" value={dataSources?.[0]?.lastSync ? String(dataSources[0].lastSync) : '-'} />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <DataQualitySummary />
        <AlertSummary />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="rounded-xl border border-gray-800 bg-gray-800/40 p-5">
          <h3 className="mb-4 text-xs font-medium uppercase tracking-wider text-gray-500">Data Sources</h3>
          <div className="space-y-2">
            {(dataSources || []).map((ds) => (
              <div key={ds.id} className="flex items-center justify-between rounded-lg bg-gray-900/30 px-3 py-2">
                <div className="flex items-center gap-2.5">
                  <span className={`h-2 w-2 rounded-full ${
                    ds.status === 'active' ? 'bg-green-500' : ds.status === 'error' ? 'bg-red-500' : 'bg-gray-500'
                  }`} />
                  <span className="text-sm text-white">{ds.name}</span>
                  <span className="text-[10px] uppercase text-gray-600">{ds.type}</span>
                </div>
                <span className="text-xs text-gray-400">{formatNumber(ds.emissions)} tCO2e</span>
              </div>
            ))}
          </div>
        </div>

        <div className="space-y-3">
          <h3 className="text-xs font-medium uppercase tracking-wider text-gray-500">Data Entry</h3>
          {[
            { label: 'Upload CSV', href: '/emissions', desc: 'Bulk import emissions data' },
            { label: 'Connect Source', href: '/settings/data-sources', desc: 'Add AWS, Azure, GCP, SAP' },
            { label: 'Emission Factors', href: '/settings/factors', desc: 'Browse and verify factors' },
            { label: 'Run Quality Scan', href: '/audit/data-quality', desc: 'Check data for anomalies' },
          ].map((action) => (
            <Link
              key={action.label}
              href={action.href}
              className="block rounded-xl border border-gray-800/50 bg-gray-900/30 p-4 transition hover:border-gray-700"
            >
              <span className="text-sm font-medium text-white">{action.label}</span>
              <p className="mt-0.5 text-xs text-gray-500">{action.desc}</p>
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
}

// ============================================================================
// Dashboard Header
// ============================================================================

const DashboardHeader = memo(function DashboardHeader({
  onExport,
  lastUpdated,
  role,
  onRoleChange,
}: {
  onExport: (format: 'pdf' | 'csv' | 'excel') => void;
  lastUpdated?: string;
  role: DashboardRole;
  onRoleChange: (r: DashboardRole) => void;
}) {
  const [exportOpen, setExportOpen] = useState(false);

  const roleLabels: Record<DashboardRole, string> = {
    executive: 'Executive View',
    operator: 'Operator View',
    auditor: 'Auditor View',
  };

  return (
    <div className="mb-6 flex items-center justify-between">
      <div>
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-bold text-white">Carbon Dashboard</h1>
          <span className="rounded bg-primary-600/10 px-2 py-0.5 text-[10px] font-medium text-primary-400">
            {roleLabels[role]}
          </span>
        </div>
        <p className="mt-0.5 text-xs text-gray-500">
          {lastUpdated ? `Updated ${new Date(lastUpdated).toLocaleString()}` : 'Real-time emissions overview'}
        </p>
      </div>
      <div className="flex items-center gap-3">
        <RoleViewSwitcher role={role} onRoleChange={onRoleChange} />
        <div className="relative">
          <button
            onClick={() => setExportOpen(!exportOpen)}
            className="flex items-center gap-2 rounded-lg border border-gray-700 bg-gray-800/50 px-4 py-2 text-sm text-gray-300 transition hover:border-gray-600 hover:text-white"
          >
            Export
            <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </button>
          {exportOpen && (
            <div className="absolute right-0 z-50 mt-1 w-44 rounded-lg border border-gray-700 bg-gray-800 shadow-xl">
              {(['pdf', 'csv', 'excel'] as const).map((fmt) => (
                <button
                  key={fmt}
                  onClick={() => { onExport(fmt); setExportOpen(false); }}
                  className="block w-full px-4 py-2 text-left text-sm text-gray-300 transition first:rounded-t-lg last:rounded-b-lg hover:bg-gray-700 hover:text-white"
                >
                  Export as {fmt.toUpperCase()}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
});

// ============================================================================
// Main Dashboard Content
// ============================================================================

const DashboardContent = memo(function DashboardContent({
  tenantId = 'default',
  timeframe = 'monthly',
  onDataChange,
}: {
  tenantId?: string;
  timeframe?: Timeframe;
  onDataChange?: (data: EmissionData) => void;
}) {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [role, setRole] = useState<DashboardRole>('executive');

  const emissions = useCarbonStore((s) => s.emissions);
  const complianceStatus = useCarbonStore((s) => s.complianceStatus);
  const metrics = useCarbonStore((s) => s.metrics);
  const dataSources = useCarbonStore((s) => s.dataSources);
  const fetchEmissions = useCarbonStore((s) => s.fetchEmissions);
  const updateMetrics = useCarbonStore((s) => s.updateMetrics);

  const { deadlines, checkCompliance } = useCompliance(tenantId);
  const { subscribe } = useRealTime();

  // Load data
  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      try {
        setIsLoading(true);
        await Promise.all([fetchEmissions(tenantId, timeframe), checkCompliance()]);
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err : new Error('Failed to load dashboard'));
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    };
    load();
    return () => { cancelled = true; };
  }, [tenantId, timeframe, fetchEmissions, checkCompliance]);

  // Real-time updates
  useEffect(() => {
    const unsub = subscribe(tenantId, (update: Partial<EmissionData>) => {
      updateMetrics(update as Partial<typeof metrics>);
    });
    return unsub;
  }, [tenantId, subscribe, updateMetrics, metrics]);

  // Export handler
  const handleExport = useCallback(async (format: 'pdf' | 'csv' | 'excel') => {
    try {
      const report = await CarbonApi.generateComplianceReport(tenantId, format, ['scope1', 'scope2', 'scope3']);
      if (report?.url) {
        window.open(report.url, '_blank');
      } else {
        const data = JSON.stringify({ emissions, metrics, complianceStatus }, null, 2);
        downloadFile(data, `offgridflow-report-${Date.now()}.json`);
      }
    } catch {
      const data = JSON.stringify({ emissions, metrics, complianceStatus }, null, 2);
      downloadFile(data, `offgridflow-report-${Date.now()}.json`);
    }
  }, [tenantId, emissions, metrics, complianceStatus]);

  if (isLoading) return <DashboardSkeleton />;

  if (error) {
    return (
      <div className="mx-auto mt-20 max-w-md rounded-xl border border-red-800/50 bg-red-900/10 p-8 text-center">
        <h2 className="mb-2 text-lg font-bold text-red-400">Failed to load dashboard</h2>
        <p className="mb-4 text-sm text-red-300/60">{error.message}</p>
        <button onClick={() => fetchEmissions(tenantId, timeframe)} className="rounded-lg bg-red-600 px-5 py-2 text-sm text-white transition hover:bg-red-500">
          Retry
        </button>
      </div>
    );
  }

  // Onboarding: no data yet
  const hasData = emissions && (emissions.scope1 > 0 || emissions.scope2 > 0 || emissions.scope3 > 0);
  if (!hasData) {
    return (
      <div className="mx-auto mt-8 max-w-3xl">
        <h1 className="mb-2 text-xl font-bold text-white">Welcome to OffGridFlow</h1>
        <p className="mb-8 text-sm text-gray-400">Import your emissions data to get started.</p>
        <div className="space-y-3">
          {[
            { step: '1', title: 'Upload Emissions Data', desc: 'Upload a CSV with utility or energy data (meter_id, location, period_start, period_end, kwh).', href: '/emissions', active: true },
            { step: '2', title: 'Connect Cloud Sources', desc: 'Set up automated pipelines from AWS, Azure, GCP, SAP, or utility providers.', href: '/settings/data-sources', active: true },
            { step: '3', title: 'Generate Compliance Reports', desc: 'After importing data, generate audit-ready CSRD, SEC, SB 253, and CBAM reports.', href: '/compliance/csrd', active: false },
          ].map((item) => (
            <Link
              key={item.step}
              href={item.href}
              className={`flex items-start gap-4 rounded-xl border p-5 transition ${
                item.active ? 'border-gray-700/50 bg-gray-800/50 hover:border-primary-600/30' : 'border-gray-800/30 bg-gray-900/20'
              }`}
            >
              <div className={`flex h-8 w-8 items-center justify-center rounded-lg text-sm font-bold ${
                item.active ? 'bg-primary-600/15 text-primary-400' : 'bg-gray-800 text-gray-600'
              }`}>
                {item.step}
              </div>
              <div>
                <h3 className={`text-sm font-semibold ${item.active ? 'text-white' : 'text-gray-600'}`}>{item.title}</h3>
                <p className={`mt-0.5 text-xs ${item.active ? 'text-gray-400' : 'text-gray-700'}`}>{item.desc}</p>
              </div>
            </Link>
          ))}
        </div>
        <div className="mt-6 rounded-lg border border-gray-800/50 bg-gray-900/30 p-3 text-center">
          <p className="text-xs text-gray-600">
            Need help? <a href="mailto:paul@off-gridflow.com" className="text-primary-400 hover:underline">paul@off-gridflow.com</a>
          </p>
        </div>
      </div>
    );
  }

  const total = emissions.scope1 + emissions.scope2 + emissions.scope3;
  const intensity = metrics.revenue ? (total / metrics.revenue) * 1000000 : 0;

  return (
    <div>
      <DashboardHeader onExport={handleExport} lastUpdated={emissions.updatedAt?.toString()} role={role} onRoleChange={setRole} />

      {/* Auditor-specific view */}
      {role === 'auditor' && <AuditorView />}

      {/* Operator-specific view */}
      {role === 'operator' && (
        <OperatorView emissions={emissions} dataSources={dataSources} timeframe={timeframe} tenantId={tenantId} />
      )}

      {/* Executive view (default) */}
      {role === 'executive' && (
        <>
          {/* Top KPI Row */}
          <div className="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
            <MetricCard
              label="Total Emissions"
              value={total}
              unit="tCO2e"
              change={emissions.percentageChange}
              changeLabel="vs last period"
              quality="calculated"
            />
            <MetricCard
              label="Scope 1 — Direct"
              value={emissions.scope1}
              unit="tCO2e"
            />
            <MetricCard
              label="Scope 2 — Energy"
              value={emissions.scope2}
              unit="tCO2e"
            />
            <MetricCard
              label="Scope 3 — Value Chain"
              value={emissions.scope3}
              unit="tCO2e"
            />
          </div>

          {/* Data Quality & Alert Indicators */}
          <div className="mb-6 grid gap-4 lg:grid-cols-2">
            <DataQualitySummary />
            <AlertSummary />
          </div>

          {/* Main Grid */}
          <div className="grid gap-6 lg:grid-cols-12">
            {/* Chart — 8 cols */}
            <div className="lg:col-span-8">
              <div className="rounded-xl border border-gray-800 bg-gray-800/40 p-5">
                <div className="mb-4 flex items-center justify-between">
                  <h3 className="text-xs font-medium uppercase tracking-wider text-gray-500">Emissions Trend</h3>
                  <span className="text-[10px] text-gray-600">GHG Protocol methodology</span>
                </div>
                <Suspense fallback={<LoadingSkeleton type="chart" />}>
                  <EmissionChart data={emissions ? [emissions] : []} timeframe={timeframe} height={320} />
                </Suspense>
              </div>
            </div>

            {/* Right sidebar — 4 cols */}
            <div className="space-y-4 lg:col-span-4">
              <ScopeBreakdown scope1={emissions.scope1} scope2={emissions.scope2} scope3={emissions.scope3} />
              <ComplianceTracker status={complianceStatus as unknown as Record<string, ComplianceStatusType>} />
            </div>
          </div>

          {/* Bottom row */}
          <div className="mt-6 grid gap-6 lg:grid-cols-12">
            <div className="lg:col-span-4">
              <MetricCard
                label="Carbon Intensity"
                value={intensity > 0 ? intensity.toFixed(1) : '—'}
                unit={intensity > 0 ? 'tCO2e / $M revenue' : ''}
              />
            </div>
            <div className="lg:col-span-4">
              <ActivityLedger dataSources={dataSources} />
            </div>
            <div className="lg:col-span-4">
              <QuickActions />
            </div>
          </div>
        </>
      )}
    </div>
  );
});

// ============================================================================
// Exported Dashboard Component
// ============================================================================

interface CarbonDashboardProps {
  tenantId?: string;
  timeframe?: Timeframe;
  onDataChange?: (data: EmissionData) => void;
}

export function CarbonDashboard({ tenantId, timeframe, onDataChange }: CarbonDashboardProps) {
  const handleUpdate = useCallback((data: Partial<EmissionData>) => {
    if (onDataChange && data as EmissionData) {
      onDataChange(data as EmissionData);
    }
  }, [onDataChange]);

  return (
    <ErrorBoundary componentName="CarbonDashboard">
      <RealTimeProvider tenantId={tenantId ?? 'default'} onUpdate={handleUpdate}>
        <DashboardContent tenantId={tenantId} timeframe={timeframe} onDataChange={onDataChange} />
      </RealTimeProvider>
    </ErrorBoundary>
  );
}

export default CarbonDashboard;
