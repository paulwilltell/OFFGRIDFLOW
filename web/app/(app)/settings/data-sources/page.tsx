'use client';

import { useEffect, useMemo, useState } from 'react';
import { api } from '@/lib/api';
import type { ScheduleStatus } from '@/lib/types';

// ─── Types ───────────────────────────────────────────────────────────────────

type ConnectorStatus = 'connected' | 'disconnected' | 'running' | 'error';

type Connector = {
  name: string;
  status: ConnectorStatus;
  last_run_at?: string;
  last_error?: string;
  last_error_at?: string;
};

type IngestionLog = {
  id: string;
  status: string;
  processed: number;
  succeeded: number;
  failed: number;
  started_at?: string;
  completed_at?: string;
};

type ConnectorField = { key: string; label: string; placeholder: string; secret?: boolean };

const CONNECTOR_FIELDS: Record<string, ConnectorField[]> = {
  aws: [
    { key: 'access_key_id', label: 'Access Key ID', placeholder: 'AKIA…' },
    { key: 'secret_access_key', label: 'Secret Access Key', placeholder: '••••••••', secret: true },
    { key: 'region', label: 'Region', placeholder: 'us-east-1' },
    { key: 'account_id', label: 'Account ID (optional)', placeholder: '123456789012' },
  ],
  azure: [
    { key: 'tenant_id', label: 'Tenant ID', placeholder: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' },
    { key: 'client_id', label: 'Client ID', placeholder: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' },
    { key: 'client_secret', label: 'Client Secret', placeholder: '••••••••', secret: true },
    { key: 'subscription_id', label: 'Subscription ID', placeholder: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx' },
  ],
  gcp: [
    { key: 'project_id', label: 'Project ID', placeholder: 'my-gcp-project' },
    { key: 'billing_account_id', label: 'Billing Account ID', placeholder: 'XXXXXX-XXXXXX-XXXXXX' },
    { key: 'service_account_key', label: 'Service Account JSON Key', placeholder: '{"type":"service_account",…}', secret: true },
  ],
  sap: [
    { key: 'base_url', label: 'SAP Base URL', placeholder: 'https://my-sap.example.com' },
    { key: 'client_id', label: 'Client ID', placeholder: 'client-id' },
    { key: 'client_secret', label: 'Client Secret', placeholder: '••••••••', secret: true },
    { key: 'company', label: 'Company Code', placeholder: '1000' },
  ],
  utility: [
    { key: 'provider', label: 'Utility Provider Name', placeholder: 'PG&E, Con Edison, etc.' },
    { key: 'account_number', label: 'Account Number', placeholder: '1234567890' },
    { key: 'api_key', label: 'API Key (if applicable)', placeholder: '••••••••', secret: true },
  ],
};

const CONNECTOR_LOGOS: Record<string, string> = {
  aws: '🟠',
  azure: '🔵',
  gcp: '🔴',
  sap: '🟡',
  utility: '⚡',
};

// ─── Configure Modal ──────────────────────────────────────────────────────────

function ConfigureModal({
  connector,
  onClose,
  onSaved,
}: {
  connector: Connector;
  onClose: () => void;
  onSaved: () => void;
}) {
  const name = connector.name.toLowerCase();
  const fields = CONNECTOR_FIELDS[name] ?? [];
  const [form, setForm] = useState<Record<string, string>>({});
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    try {
      await api.post('/api/connectors/configure', { name, config: form });
      onSaved();
      onClose();
    } catch (e) {
      setError((e as Error).message || 'Failed to save credentials');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="w-full max-w-lg mx-4 rounded-2xl bg-white dark:bg-gray-900 shadow-2xl border border-gray-200 dark:border-gray-700">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-5 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-3">
            <span className="text-2xl">{CONNECTOR_LOGOS[name] ?? '🔌'}</span>
            <div>
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                Configure {connector.name}
              </h2>
              <p className="text-sm text-gray-500 dark:text-gray-400">
                Credentials are stored securely per your organisation
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-xl font-bold"
          >
            ×
          </button>
        </div>

        {/* Fields */}
        <div className="px-6 py-5 space-y-4 max-h-[60vh] overflow-y-auto">
          {fields.length === 0 && (
            <p className="text-sm text-gray-500 dark:text-gray-400">
              No fields defined for this connector yet.
            </p>
          )}
          {fields.map((f) => (
            <div key={f.key}>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {f.label}
              </label>
              {f.key === 'service_account_key' ? (
                <textarea
                  rows={4}
                  placeholder={f.placeholder}
                  value={form[f.key] ?? ''}
                  onChange={(e) => setForm((prev) => ({ ...prev, [f.key]: e.target.value }))}
                  className="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-sm text-gray-900 dark:text-white font-mono focus:outline-none focus:ring-2 focus:ring-green-500 resize-none"
                />
              ) : (
                <input
                  type={f.secret ? 'password' : 'text'}
                  placeholder={f.placeholder}
                  value={form[f.key] ?? ''}
                  onChange={(e) => setForm((prev) => ({ ...prev, [f.key]: e.target.value }))}
                  className="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-sm text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-green-500"
                />
              )}
            </div>
          ))}
          {error && (
            <div className="rounded-lg bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-700 px-4 py-3 text-sm text-red-700 dark:text-red-300">
              {error}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-3 px-6 py-4 border-t border-gray-200 dark:border-gray-700">
          <button
            onClick={onClose}
            className="px-4 py-2 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className="px-5 py-2 rounded-lg text-sm font-medium text-white bg-green-600 hover:bg-green-700 disabled:opacity-50 transition-colors"
          >
            {saving ? 'Saving…' : 'Save Credentials'}
          </button>
        </div>
      </div>
    </div>
  );
}

// ─── Connector Card ───────────────────────────────────────────────────────────

function ConnectorCard({
  connector,
  loading,
  onSync,
  onTest,
  onConfigure,
}: {
  connector: Connector;
  loading: boolean;
  onSync: () => void;
  onTest: () => void;
  onConfigure: () => void;
}) {
  const name = connector.name.toLowerCase();
  const logo = CONNECTOR_LOGOS[name] ?? '🔌';
  const isConnected = connector.status === 'connected';
  const isRunning = connector.status === 'running';
  const isError = connector.status === 'error';

  return (
    <div className="rounded-2xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-6 flex flex-col gap-4 shadow-sm hover:shadow-md transition-shadow">
      {/* Header row */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="text-3xl">{logo}</span>
          <div>
            <h3 className="font-semibold text-gray-900 dark:text-white">{connector.name}</h3>
            <div className="flex items-center gap-1.5 mt-0.5">
              <span
                className={`inline-block w-2 h-2 rounded-full ${
                  isConnected
                    ? 'bg-green-500'
                    : isRunning
                    ? 'bg-blue-500 animate-pulse'
                    : isError
                    ? 'bg-red-500'
                    : 'bg-gray-300 dark:bg-gray-600'
                }`}
              />
              <span className="text-xs text-gray-500 dark:text-gray-400 capitalize">
                {connector.status}
              </span>
            </div>
          </div>
        </div>
        <button
          onClick={onConfigure}
          className="text-xs font-medium text-green-600 dark:text-green-400 hover:underline"
        >
          Configure →
        </button>
      </div>

      {/* Last sync */}
      <div className="text-xs text-gray-500 dark:text-gray-400">
        Last sync:{' '}
        {connector.last_run_at ? new Date(connector.last_run_at).toLocaleString() : 'Never'}
      </div>

      {/* Error */}
      {isError && connector.last_error && (
        <div className="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-700 px-3 py-2 text-xs text-red-600 dark:text-red-400">
          {connector.last_error}
        </div>
      )}

      {/* Actions */}
      <div className="flex gap-2 mt-auto pt-2 border-t border-gray-100 dark:border-gray-700">
        <button
          onClick={onSync}
          disabled={loading || isRunning}
          className="flex-1 py-2 rounded-lg text-xs font-medium text-white bg-green-600 hover:bg-green-700 disabled:opacity-50 transition-colors"
        >
          {isRunning ? 'Running…' : 'Sync Now'}
        </button>
        <button
          onClick={onTest}
          disabled={loading}
          className="flex-1 py-2 rounded-lg text-xs font-medium text-gray-700 dark:text-gray-200 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 disabled:opacity-50 transition-colors"
        >
          Test Connection
        </button>
      </div>
    </div>
  );
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function DataSourcesPage() {
  const [connectors, setConnectors] = useState<Connector[]>([]);
  const [logs, setLogs] = useState<IngestionLog[]>([]);
  const [schedule, setSchedule] = useState<ScheduleStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [toast, setToast] = useState<string | null>(null);
  const [configTarget, setConfigTarget] = useState<Connector | null>(null);

  const showToast = (msg: string) => {
    setToast(msg);
    setTimeout(() => setToast(null), 4000);
  };

  const fetchData = useMemo(
    () => async () => {
      setLoading(true);
      setError(null);
      try {
        const [c, l] = await Promise.all([
          api.get<Connector[]>('/api/connectors/list'),
          api.get<IngestionLog[]>('/api/ingestion/logs?limit=10'),
        ]);
        setConnectors(c);
        setLogs(l);
        try {
          const s = await api.get<ScheduleStatus>('/api/connectors/schedule');
          setSchedule(s);
        } catch {
          setSchedule(null);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load data');
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const triggerRun = async () => {
    try {
      setLoading(true);
      await api.post<{ status: string }>('/api/connectors/run', {});
      showToast('Ingestion started for all connectors');
      await fetchData();
    } catch (e) {
      setError('Failed to start ingestion');
    } finally {
      setLoading(false);
    }
  };

  const triggerSync = async (name: string) => {
    try {
      setLoading(true);
      await api.post('/api/connectors/run', {});
      showToast(`${name} sync started`);
      await fetchData();
    } catch {
      showToast('Sync failed');
    } finally {
      setLoading(false);
    }
  };

  const onTest = async (name: string) => {
    try {
      setLoading(true);
      await api.post(`/api/connectors/test?name=${encodeURIComponent(name)}`, {});
      showToast(`${name} connection test passed ✓`);
      await fetchData();
    } catch {
      showToast(`${name} connection test failed`);
    } finally {
      setLoading(false);
    }
  };

  const allConnectorNames = ['AWS', 'AZURE', 'GCP', 'SAP', 'UTILITY'];
  const merged: Connector[] = allConnectorNames.map((name) => {
    const found = connectors.find((c) => c.name.toUpperCase() === name);
    return found ?? { name, status: 'disconnected' };
  });

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 py-10 px-4 sm:px-6 lg:px-8">
      <div className="max-w-6xl mx-auto space-y-8">
        {/* Page header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <p className="text-sm font-medium text-green-600 dark:text-green-400 uppercase tracking-widest">
              Data Sources
            </p>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
              Cloud Connectors
            </h1>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              Connect your cloud accounts to automatically pull emissions data in real time.
            </p>
          </div>
          <div className="flex gap-3">
            <button
              onClick={fetchData}
              disabled={loading}
              className="px-4 py-2 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 transition-colors shadow-sm"
            >
              {loading ? 'Refreshing…' : 'Refresh'}
            </button>
            <button
              onClick={triggerRun}
              disabled={loading}
              className="px-4 py-2 rounded-lg text-sm font-medium text-white bg-green-600 hover:bg-green-700 disabled:opacity-50 transition-colors shadow-sm"
            >
              Run All Now
            </button>
          </div>
        </div>

        {/* Toast */}
        {toast && (
          <div className="fixed top-6 right-6 z-50 rounded-xl bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 text-sm font-medium px-5 py-3 shadow-xl animate-in fade-in slide-in-from-top-2">
            {toast}
          </div>
        )}

        {/* Error */}
        {error && (
          <div className="rounded-xl bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-700 px-5 py-4 text-sm text-red-700 dark:text-red-300">
            {error}
          </div>
        )}

        {/* Schedule card */}
        <div className="rounded-2xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 px-6 py-5 shadow-sm">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">
            Automated Sync Schedule
          </h2>
          {schedule ? (
            <div className="grid grid-cols-3 gap-4 text-sm">
              <div>
                <span className="text-gray-500 dark:text-gray-400">Interval</span>
                <p className="font-medium text-gray-900 dark:text-white mt-0.5">
                  {schedule.interval ?? 'Manual'}
                </p>
              </div>
              <div>
                <span className="text-gray-500 dark:text-gray-400">Last run</span>
                <p className="font-medium text-gray-900 dark:text-white mt-0.5">
                  {schedule.last_run_at ? new Date(schedule.last_run_at).toLocaleString() : '—'}
                </p>
              </div>
              <div>
                <span className="text-gray-500 dark:text-gray-400">Next run</span>
                <p className="font-medium text-gray-900 dark:text-white mt-0.5">
                  {schedule.next_run_at ? new Date(schedule.next_run_at).toLocaleString() : '—'}
                </p>
              </div>
            </div>
          ) : (
            <p className="text-sm text-gray-500 dark:text-gray-400">
              Automated schedule not active. Trigger syncs manually below or configure via Railway
              environment variables.
            </p>
          )}
        </div>

        {/* Connector grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
          {loading &&
            Array.from({ length: 5 }).map((_, i) => (
              <div
                key={i}
                className="rounded-2xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 h-48 animate-pulse"
              />
            ))}
          {!loading &&
            merged.map((c) => (
              <ConnectorCard
                key={c.name}
                connector={c}
                loading={loading}
                onSync={() => triggerSync(c.name)}
                onTest={() => onTest(c.name)}
                onConfigure={() => setConfigTarget(c)}
              />
            ))}
        </div>

        {/* Ingestion logs */}
        <div className="rounded-2xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-sm overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
            <h2 className="text-sm font-semibold text-gray-900 dark:text-white">
              Recent Ingestion Runs
            </h2>
            <span className="text-xs text-gray-500 dark:text-gray-400">Last 10 runs</span>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700/50">
                <tr>
                  {['Status', 'Processed', 'Succeeded', 'Failed', 'Started'].map((h) => (
                    <th
                      key={h}
                      className="px-5 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider"
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
                {logs.length === 0 && (
                  <tr>
                    <td
                      colSpan={5}
                      className="px-5 py-8 text-center text-sm text-gray-400 dark:text-gray-500"
                    >
                      No ingestion runs yet. Trigger a sync to get started.
                    </td>
                  </tr>
                )}
                {logs.map((log) => (
                  <tr key={log.id} className="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors">
                    <td className="px-5 py-3">
                      <StatusBadge status={log.status} />
                    </td>
                    <td className="px-5 py-3 text-gray-700 dark:text-gray-300">{log.processed}</td>
                    <td className="px-5 py-3 text-green-600 dark:text-green-400">{log.succeeded}</td>
                    <td className={`px-5 py-3 ${log.failed > 0 ? 'text-red-600 dark:text-red-400' : 'text-gray-500 dark:text-gray-400'}`}>
                      {log.failed}
                    </td>
                    <td className="px-5 py-3 text-gray-500 dark:text-gray-400">
                      {log.started_at ? new Date(log.started_at).toLocaleString() : '—'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Configure Modal */}
      {configTarget && (
        <ConfigureModal
          connector={configTarget}
          onClose={() => setConfigTarget(null)}
          onSaved={() => {
            showToast(`${configTarget.name} credentials saved ✓`);
            fetchData();
          }}
        />
      )}
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const map: Record<string, string> = {
    completed: 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-400',
    running: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400',
    error: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400',
    failed: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400',
  };
  const cls = map[status] ?? 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400';
  return (
    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium capitalize ${cls}`}>
      {status}
    </span>
  );
}
