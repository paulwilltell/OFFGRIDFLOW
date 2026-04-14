'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api';
import { useRequireAuth, useSession } from '@/lib/session';
import { recordAuditEvent, exportAuditLog } from '@/lib/auditLog';

interface RetentionPolicy {
  emission_data?: string;
  calculation_ledger?: string;
  user_accounts?: string;
  change_log?: string;
  evidence_files?: string;
  export_formats?: string[];
  data_ownership_statement?: string;
  [key: string]: unknown;
}

interface DeletionResponse {
  status?: string;
  retention_days?: number;
  deletion_date?: string;
  message?: string;
}

export default function DataGovernancePage() {
  const session = useRequireAuth();
  const { user } = useSession();

  const [retention, setRetention] = useState<RetentionPolicy | null>(null);
  const [retentionError, setRetentionError] = useState<string | null>(null);
  const [retentionLoading, setRetentionLoading] = useState(true);

  const [exportStatus, setExportStatus] = useState<'idle' | 'running' | 'success' | 'error'>('idle');
  const [exportError, setExportError] = useState<string | null>(null);

  const [deletionStatus, setDeletionStatus] = useState<'idle' | 'running' | 'success' | 'error'>('idle');
  const [deletionResult, setDeletionResult] = useState<DeletionResponse | null>(null);
  const [deletionError, setDeletionError] = useState<string | null>(null);
  const [confirmText, setConfirmText] = useState('');

  const isAdmin = user?.role === 'admin' || user?.role === 'superadmin';

  useEffect(() => {
    let cancelled = false;
    async function loadRetention() {
      setRetentionLoading(true);
      try {
        const res = await api.get<RetentionPolicy>('/api/governance/retention');
        if (!cancelled) setRetention(res);
      } catch (err) {
        if (!cancelled) {
          const msg = err instanceof Error ? err.message : 'Failed to load retention policy';
          setRetentionError(msg);
        }
      } finally {
        if (!cancelled) setRetentionLoading(false);
      }
    }
    loadRetention();
    return () => {
      cancelled = true;
    };
  }, []);

  const handleExport = async () => {
    setExportStatus('running');
    setExportError(null);
    try {
      const data = await api.get<unknown>('/api/governance/export');
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `offgridflow-data-export-${new Date().toISOString().slice(0, 10)}.json`;
      a.click();
      URL.revokeObjectURL(url);
      setExportStatus('success');
      recordAuditEvent('report.exported', {
        entityType: 'governance_full_export',
        metadata: { initiated_by: user?.email ?? null },
      });
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Export failed';
      setExportStatus('error');
      setExportError(msg);
    }
  };

  const handleDeletionRequest = async () => {
    if (confirmText.trim().toLowerCase() !== 'delete my organization') {
      setDeletionError('Type "DELETE MY ORGANIZATION" exactly to confirm.');
      return;
    }
    setDeletionStatus('running');
    setDeletionError(null);
    try {
      const res = await api.post<DeletionResponse>('/api/governance/delete-request', {});
      setDeletionResult(res);
      setDeletionStatus('success');
      setConfirmText('');
      recordAuditEvent('alert.resolved', {
        entityType: 'deletion_request',
        metadata: {
          initiated_by: user?.email ?? null,
          retention_days: res?.retention_days ?? null,
          deletion_date: res?.deletion_date ?? null,
        },
      });
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Deletion request failed';
      setDeletionStatus('error');
      setDeletionError(msg);
    }
  };

  if (!session?.isAuthenticated) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center text-gray-400">Loading…</div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl">
      <h1 className="text-2xl font-bold text-white">Data Governance</h1>
      <p className="mt-2 text-sm text-gray-400">
        Export your organization&apos;s data at any time, review the retention schedule, and
        request permanent deletion. Backed by the same endpoints documented in the{' '}
        <a href="/trust" className="text-primary-400 hover:underline" target="_blank" rel="noopener noreferrer">
          Trust Center
        </a>
        .
      </p>

      {/* Retention */}
      <section className="mt-10 rounded-xl border border-gray-800 bg-gray-800/30 p-6">
        <h2 className="text-sm font-semibold text-white">Retention Policy</h2>
        <p className="mt-1 text-xs text-gray-500">
          Source: <code className="rounded bg-gray-900 px-1 text-xs text-primary-400">GET /api/governance/retention</code>
        </p>
        {retentionLoading && (
          <div className="mt-4 text-sm text-gray-400">Loading current policy…</div>
        )}
        {retentionError && (
          <div className="mt-4 rounded-lg border border-red-700 bg-red-900/20 p-3 text-sm text-red-300">
            {retentionError}
          </div>
        )}
        {retention && (
          <div className="mt-4 grid gap-3 text-sm sm:grid-cols-2">
            {[
              ['Emission data', retention.emission_data],
              ['Calculation ledger', retention.calculation_ledger],
              ['User accounts', retention.user_accounts],
              ['Change log', retention.change_log],
              ['Evidence files', retention.evidence_files],
            ].map(([label, value]) =>
              value ? (
                <div key={label as string} className="rounded-lg bg-gray-900/40 p-3">
                  <div className="text-xs font-medium text-gray-500">{label}</div>
                  <div className="mt-0.5 text-sm text-gray-200">{String(value)}</div>
                </div>
              ) : null,
            )}
          </div>
        )}
      </section>

      {/* Export */}
      <section className="mt-6 rounded-xl border border-gray-800 bg-gray-800/30 p-6">
        <h2 className="text-sm font-semibold text-white">Export Organization Data</h2>
        <p className="mt-1 text-sm text-gray-400">
          Downloads a JSON archive containing users, activities, calculation ledger entries,
          and change log. Export is available at any time during your subscription.
        </p>
        <button
          type="button"
          onClick={handleExport}
          disabled={exportStatus === 'running'}
          className="mt-4 rounded-lg bg-primary-600 px-5 py-2 text-sm font-medium text-white hover:bg-primary-500 disabled:opacity-50"
        >
          {exportStatus === 'running' ? 'Exporting…' : 'Download JSON Export'}
        </button>
        {exportStatus === 'success' && (
          <p className="mt-3 text-xs text-green-400">
            Export downloaded successfully. Store the file in your records system.
          </p>
        )}
        {exportStatus === 'error' && exportError && (
          <p className="mt-3 text-xs text-red-400">Error: {exportError}</p>
        )}
      </section>

      {/* Client audit log export */}
      <section className="mt-6 rounded-xl border border-gray-800 bg-gray-800/30 p-6">
        <h2 className="text-sm font-semibold text-white">Client-Side Audit Log</h2>
        <p className="mt-1 text-sm text-gray-400">
          Download the local record of actions you&apos;ve taken in this browser (data uploads,
          calculations run, reports exported). Independent of the server-side audit_logs table.
        </p>
        <button
          type="button"
          onClick={() => {
            exportAuditLog();
            recordAuditEvent('report.exported', { entityType: 'client_audit_log' });
          }}
          className="mt-4 rounded-lg border border-gray-700 px-5 py-2 text-sm text-gray-200 hover:border-gray-500 hover:text-white"
        >
          Download Audit Log
        </button>
      </section>

      {/* Deletion */}
      <section className="mt-6 rounded-xl border border-red-900/40 bg-red-950/20 p-6">
        <h2 className="text-sm font-semibold text-red-300">Permanent Deletion</h2>
        <p className="mt-1 text-sm text-gray-400">
          Requests permanent deletion of your organization&apos;s data. A{' '}
          <strong className="text-red-300">30-day retention window</strong> begins so that export
          and recovery remain possible. After the window, data is permanently deleted.
        </p>
        {!isAdmin && (
          <div className="mt-4 rounded-lg border border-amber-700/40 bg-amber-900/20 p-3 text-xs text-amber-200">
            Only users with an admin role can initiate deletion. Contact your administrator or
            email contact@off-grid-flow.com.
          </div>
        )}
        {isAdmin && deletionStatus !== 'success' && (
          <div className="mt-4 space-y-3">
            <label className="block text-xs text-gray-400">
              Type <code className="rounded bg-gray-900 px-1 text-red-300">DELETE MY ORGANIZATION</code> to confirm:
              <input
                type="text"
                value={confirmText}
                onChange={(e) => setConfirmText(e.target.value)}
                className="mt-2 block w-full rounded-md border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-white focus:border-red-500 focus:outline-none focus:ring-1 focus:ring-red-500"
                placeholder="DELETE MY ORGANIZATION"
              />
            </label>
            <button
              type="button"
              onClick={handleDeletionRequest}
              disabled={deletionStatus === 'running'}
              className="rounded-lg bg-red-600 px-5 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-50"
            >
              {deletionStatus === 'running' ? 'Submitting…' : 'Request Permanent Deletion'}
            </button>
            {deletionError && (
              <p className="text-xs text-red-400">Error: {deletionError}</p>
            )}
          </div>
        )}
        {deletionStatus === 'success' && deletionResult && (
          <div className="mt-4 rounded-lg border border-amber-700/50 bg-amber-900/20 p-4">
            <div className="text-sm font-semibold text-amber-200">Deletion Request Recorded</div>
            <div className="mt-2 space-y-1 text-xs text-amber-100">
              {deletionResult.retention_days != null && (
                <div>
                  Retention window: <span className="font-semibold">{deletionResult.retention_days} days</span>
                </div>
              )}
              {deletionResult.deletion_date && (
                <div>
                  Final deletion on or after:{' '}
                  <span className="font-semibold">{deletionResult.deletion_date}</span>
                </div>
              )}
              {deletionResult.message && <div className="text-xs text-amber-200">{deletionResult.message}</div>}
              <div className="mt-2">
                You can still export data during the retention window using the button above.
              </div>
            </div>
          </div>
        )}
      </section>

      <p className="mt-8 text-xs text-gray-600">
        Questions about data governance? Email{' '}
        <a href="mailto:contact@off-grid-flow.com" className="text-primary-400 hover:underline">
          contact@off-grid-flow.com
        </a>
        .
      </p>
    </div>
  );
}
