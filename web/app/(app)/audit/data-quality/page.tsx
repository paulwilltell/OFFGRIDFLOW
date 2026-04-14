'use client';

import { useCallback, useEffect, useState } from 'react';
import { api } from '@/lib/api';
import { useRequireAuth } from '@/lib/session';

interface AnomalySummary {
  open: number;
  critical: number;
  warning: number;
  info: number;
  resolved: number;
  dismissed: number;
  acknowledged: number;
}

interface Anomaly {
  id: string;
  anomaly_type: string;
  severity: 'critical' | 'warning' | 'info';
  description: string;
  expected_value: number | null;
  actual_value: number | null;
  deviation_percent: number | null;
  status: 'open' | 'acknowledged' | 'resolved' | 'dismissed';
  entity_type?: string;
  entity_id?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

interface AnomaliesResponse {
  anomalies: Anomaly[];
  count: number;
  summary: AnomalySummary;
}

interface ScanResponse {
  detected: number;
  status: string;
}

type FilterTab = 'all' | 'open' | 'critical' | 'resolved';

const severityConfig: Record<string, { label: string; bg: string; text: string; dot: string }> = {
  critical: { label: 'Critical', bg: 'bg-red-500/10', text: 'text-red-400', dot: 'bg-red-500' },
  warning: { label: 'Warning', bg: 'bg-amber-500/10', text: 'text-amber-400', dot: 'bg-amber-500' },
  info: { label: 'Info', bg: 'bg-blue-500/10', text: 'text-blue-400', dot: 'bg-blue-500' },
};

const statusConfig: Record<string, { label: string; bg: string; text: string }> = {
  open: { label: 'Open', bg: 'bg-red-500/10', text: 'text-red-400' },
  acknowledged: { label: 'Acknowledged', bg: 'bg-amber-500/10', text: 'text-amber-400' },
  resolved: { label: 'Resolved', bg: 'bg-green-500/10', text: 'text-green-400' },
  dismissed: { label: 'Dismissed', bg: 'bg-gray-500/10', text: 'text-gray-400' },
};

function formatAnomalyType(type: string): string {
  return type
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

function formatDeviation(value: number | null): string {
  if (value === null || value === undefined) return '--';
  const sign = value >= 0 ? '+' : '';
  return `${sign}${value.toFixed(1)}%`;
}

function formatValue(value: number | null): string {
  if (value === null || value === undefined) return '--';
  if (Math.abs(value) >= 1_000_000) return `${(value / 1_000_000).toFixed(2)}M`;
  if (Math.abs(value) >= 1_000) return `${(value / 1_000).toFixed(2)}K`;
  return value.toFixed(2);
}

export default function DataQualityPage() {
  useRequireAuth();

  const [anomalies, setAnomalies] = useState<Anomaly[]>([]);
  const [summary, setSummary] = useState<AnomalySummary>({
    open: 0,
    critical: 0,
    warning: 0,
    info: 0,
    resolved: 0,
    dismissed: 0,
    acknowledged: 0,
  });
  const [loading, setLoading] = useState(true);
  const [scanning, setScanning] = useState(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<FilterTab>('all');
  const [scanResult, setScanResult] = useState<ScanResponse | null>(null);

  const buildQueryParams = useCallback((tab: FilterTab): string => {
    const params = new URLSearchParams();
    switch (tab) {
      case 'open':
        params.set('status', 'open');
        break;
      case 'critical':
        params.set('severity', 'critical');
        break;
      case 'resolved':
        params.set('status', 'resolved');
        break;
    }
    const qs = params.toString();
    return qs ? `?${qs}` : '';
  }, []);

  const loadAnomalies = useCallback(async (tab: FilterTab) => {
    try {
      const params = buildQueryParams(tab);
      const res = await api.get<AnomaliesResponse>(`/api/audit/anomalies${params}`);
      setAnomalies(res.anomalies || []);
      setSummary(res.summary || {
        open: 0,
        critical: 0,
        warning: 0,
        info: 0,
        resolved: 0,
        dismissed: 0,
        acknowledged: 0,
      });
    } catch {
      setAnomalies([]);
    } finally {
      setLoading(false);
    }
  }, [buildQueryParams]);

  useEffect(() => {
    setLoading(true);
    loadAnomalies(activeTab);
  }, [activeTab, loadAnomalies]);

  const handleScan = async () => {
    setScanning(true);
    setScanResult(null);
    try {
      const res = await api.post<ScanResponse>('/api/audit/anomalies/scan', {});
      setScanResult(res);
      await loadAnomalies(activeTab);
    } catch {
      setScanResult(null);
    } finally {
      setScanning(false);
    }
  };

  const handleAction = async (id: string, action: 'resolve' | 'dismiss') => {
    setActionLoading(id);
    try {
      await api.put(`/api/audit/anomalies/${id}`, { action, notes: '' });
      await loadAnomalies(activeTab);
    } catch {
      // Error is handled by the API client
    } finally {
      setActionLoading(null);
    }
  };

  const tabs: { key: FilterTab; label: string }[] = [
    { key: 'all', label: 'All' },
    { key: 'open', label: 'Open' },
    { key: 'critical', label: 'Critical' },
    { key: 'resolved', label: 'Resolved' },
  ];

  return (
    <div>
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white">Data Quality</h1>
          <p className="mt-1 text-xs text-gray-500">
            Anomaly detection for emissions data &mdash; identify outliers, missing values, and inconsistencies
          </p>
        </div>
        <button
          onClick={handleScan}
          disabled={scanning}
          className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-primary-500 disabled:opacity-50"
        >
          {scanning ? (
            <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
          ) : (
            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          )}
          {scanning ? 'Scanning...' : 'Run Scan'}
        </button>
      </div>

      {/* Scan result banner */}
      {scanResult && (
        <div className="mb-4 flex items-center justify-between rounded-lg border border-green-500/20 bg-green-500/5 px-4 py-3">
          <div className="flex items-center gap-2">
            <svg className="h-4 w-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
            <span className="text-sm text-green-400">
              Scan complete &mdash; {scanResult.detected} anomal{scanResult.detected === 1 ? 'y' : 'ies'} detected
            </span>
          </div>
          <button
            onClick={() => setScanResult(null)}
            className="text-gray-500 hover:text-gray-300"
          >
            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      )}

      {/* Summary cards */}
      <div className="mb-6 grid grid-cols-4 gap-4">
        <div className="rounded-xl border border-gray-800 bg-gray-800/30 p-4">
          <div className="text-[10px] uppercase tracking-wider text-gray-500">Open</div>
          <div className="mt-1 text-2xl font-bold text-white">{summary.open}</div>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-800/30 p-4">
          <div className="text-[10px] uppercase tracking-wider text-gray-500">Critical</div>
          <div className="mt-1 text-2xl font-bold text-red-400">{summary.critical}</div>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-800/30 p-4">
          <div className="text-[10px] uppercase tracking-wider text-gray-500">Warning</div>
          <div className="mt-1 text-2xl font-bold text-amber-400">{summary.warning}</div>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-800/30 p-4">
          <div className="text-[10px] uppercase tracking-wider text-gray-500">Resolved</div>
          <div className="mt-1 text-2xl font-bold text-green-400">{summary.resolved}</div>
        </div>
      </div>

      {/* Filter tabs */}
      <div className="mb-4 flex items-center gap-2">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={`rounded-lg px-3 py-1.5 text-xs font-medium transition ${
              activeTab === tab.key
                ? 'bg-primary-600 text-white'
                : 'bg-gray-800 text-gray-400 hover:text-white'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Anomalies list */}
      {loading ? (
        <div className="py-20 text-center text-gray-500">Loading anomalies...</div>
      ) : anomalies.length === 0 ? (
        <div className="rounded-xl border border-gray-800 bg-gray-800/30 p-12 text-center">
          <div className="mb-3 text-3xl">&#9989;</div>
          <h3 className="mb-2 text-base font-semibold text-white">No Anomalies Found</h3>
          <p className="mb-4 text-sm text-gray-400">
            {activeTab === 'all'
              ? 'Run a scan to detect data quality issues in your emissions data.'
              : `No ${activeTab} anomalies at this time.`}
          </p>
          {activeTab === 'all' && (
            <button
              onClick={handleScan}
              disabled={scanning}
              className="inline-block rounded-lg bg-primary-600 px-5 py-2 text-sm font-medium text-white hover:bg-primary-500 disabled:opacity-50"
            >
              Run Scan
            </button>
          )}
        </div>
      ) : (
        <div className="space-y-3">
          {anomalies.map((anomaly) => {
            const sev = severityConfig[anomaly.severity] || severityConfig.info;
            const stat = statusConfig[anomaly.status] || statusConfig.open;
            const isActionable = anomaly.status === 'open' || anomaly.status === 'acknowledged';
            const isUpdating = actionLoading === anomaly.id;

            return (
              <div
                key={anomaly.id}
                className="rounded-xl border border-gray-800 bg-gray-800/30 p-5 transition hover:border-gray-700"
              >
                {/* Top row: severity, type, status, date */}
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3">
                    <span className={`inline-flex items-center gap-1.5 rounded px-2 py-0.5 text-[10px] font-medium ${sev.bg} ${sev.text}`}>
                      <span className={`h-1.5 w-1.5 rounded-full ${sev.dot}`} />
                      {sev.label}
                    </span>
                    <span className="text-sm font-medium text-white">
                      {formatAnomalyType(anomaly.anomaly_type)}
                    </span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className={`rounded px-2 py-0.5 text-[10px] font-medium ${stat.bg} ${stat.text}`}>
                      {stat.label}
                    </span>
                    <span className="text-xs text-gray-500">
                      {new Date(anomaly.created_at).toLocaleDateString(undefined, {
                        year: 'numeric',
                        month: 'short',
                        day: 'numeric',
                      })}
                    </span>
                  </div>
                </div>

                {/* Description */}
                <p className="mt-2 text-sm text-gray-400">{anomaly.description}</p>

                {/* Expected vs Actual */}
                {(anomaly.expected_value !== null || anomaly.actual_value !== null) && (
                  <div className="mt-3 flex items-center gap-6">
                    <div>
                      <div className="text-[10px] uppercase tracking-wider text-gray-600">Expected</div>
                      <div className="text-sm font-medium text-gray-300">
                        {formatValue(anomaly.expected_value)}
                      </div>
                    </div>
                    <div className="text-gray-700">
                      <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13 7l5 5m0 0l-5 5m5-5H6" />
                      </svg>
                    </div>
                    <div>
                      <div className="text-[10px] uppercase tracking-wider text-gray-600">Actual</div>
                      <div className="text-sm font-medium text-white">
                        {formatValue(anomaly.actual_value)}
                      </div>
                    </div>
                    {anomaly.deviation_percent !== null && (
                      <div className="ml-2 rounded-lg bg-gray-800/50 px-3 py-1">
                        <div className="text-[10px] uppercase tracking-wider text-gray-600">Deviation</div>
                        <div className={`text-sm font-semibold ${
                          anomaly.deviation_percent !== null && Math.abs(anomaly.deviation_percent) > 50
                            ? 'text-red-400'
                            : anomaly.deviation_percent !== null && Math.abs(anomaly.deviation_percent) > 20
                              ? 'text-amber-400'
                              : 'text-gray-300'
                        }`}>
                          {formatDeviation(anomaly.deviation_percent)}
                        </div>
                      </div>
                    )}
                  </div>
                )}

                {/* Action buttons */}
                {isActionable && (
                  <div className="mt-4 flex items-center gap-2 border-t border-gray-800 pt-4">
                    <button
                      onClick={() => handleAction(anomaly.id, 'resolve')}
                      disabled={isUpdating}
                      className="rounded-lg bg-green-600 px-3 py-1.5 text-xs font-medium text-white transition hover:bg-green-500 disabled:opacity-50"
                    >
                      {isUpdating ? 'Updating...' : 'Resolve'}
                    </button>
                    <button
                      onClick={() => handleAction(anomaly.id, 'dismiss')}
                      disabled={isUpdating}
                      className="rounded-lg bg-gray-700 px-3 py-1.5 text-xs font-medium text-gray-300 transition hover:bg-gray-600 disabled:opacity-50"
                    >
                      Dismiss
                    </button>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
