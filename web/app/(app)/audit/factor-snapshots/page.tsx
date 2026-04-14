'use client';

import { useCallback, useEffect, useState } from 'react';
import { api } from '@/lib/api';
import { useRequireAuth } from '@/lib/session';

interface SnapshotFactor {
  id: string;
  scope: string;
  category: string;
  region: string;
  source: string;
  unit: string;
  value_kg_co2e: number;
}

interface Snapshot {
  id: string;
  name: string;
  period_start: string;
  period_end: string;
  status: 'draft' | 'locked' | 'superseded';
  factor_count: number;
  locked_at: string | null;
  created_at: string;
  factors?: SnapshotFactor[];
}

interface SnapshotListResponse {
  snapshots: Snapshot[];
  count: number;
}

interface SnapshotDetailResponse extends Snapshot {
  factors: SnapshotFactor[];
}

const statusConfig: Record<string, { label: string; bg: string; text: string }> = {
  draft: { label: 'Draft', bg: 'bg-yellow-500/10', text: 'text-yellow-400' },
  locked: { label: 'Locked', bg: 'bg-green-500/10', text: 'text-green-400' },
  superseded: { label: 'Superseded', bg: 'bg-gray-500/10', text: 'text-gray-400' },
};

export default function FactorSnapshotsPage() {
  useRequireAuth();

  const [snapshots, setSnapshots] = useState<Snapshot[]>([]);
  const [loading, setLoading] = useState(true);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [expandedFactors, setExpandedFactors] = useState<SnapshotFactor[]>([]);
  const [expandLoading, setExpandLoading] = useState(false);
  const [lockingId, setLockingId] = useState<string | null>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [formName, setFormName] = useState('');
  const [formPeriodStart, setFormPeriodStart] = useState('');
  const [formPeriodEnd, setFormPeriodEnd] = useState('');

  const loadSnapshots = useCallback(async () => {
    try {
      const res = await api.get<SnapshotListResponse>('/api/audit/factor-snapshots');
      setSnapshots(res.snapshots || []);
    } catch {
      setSnapshots([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadSnapshots();
  }, [loadSnapshots]);

  const handleExpand = async (id: string) => {
    if (expandedId === id) {
      setExpandedId(null);
      setExpandedFactors([]);
      return;
    }

    setExpandedId(id);
    setExpandLoading(true);
    setExpandedFactors([]);

    try {
      const res = await api.get<SnapshotDetailResponse>(`/api/audit/factor-snapshots/${id}`);
      setExpandedFactors(res.factors || []);
    } catch {
      setExpandedFactors([]);
    } finally {
      setExpandLoading(false);
    }
  };

  const handleLock = async (id: string) => {
    setLockingId(id);
    setError(null);
    try {
      await api.post<{ status: string }>(`/api/audit/factor-snapshots/${id}/lock`, {});
      await loadSnapshots();
      if (expandedId === id) {
        setExpandedId(null);
        setExpandedFactors([]);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to lock snapshot');
    } finally {
      setLockingId(null);
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formName.trim() || !formPeriodStart || !formPeriodEnd) return;

    setCreating(true);
    setError(null);
    try {
      await api.post('/api/audit/factor-snapshots', {
        name: formName.trim(),
        period_start: formPeriodStart,
        period_end: formPeriodEnd,
        factors: [],
      });
      setFormName('');
      setFormPeriodStart('');
      setFormPeriodEnd('');
      setShowCreateForm(false);
      await loadSnapshots();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create snapshot');
    } finally {
      setCreating(false);
    }
  };

  const formatDate = (dateStr: string | null): string => {
    if (!dateStr) return '\u2014';
    return new Date(dateStr).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white">Factor Snapshots</h1>
          <p className="mt-1 text-xs text-gray-500">
            Locked sets of emission factors tied to reporting periods for audit reproducibility
          </p>
        </div>
        <button
          onClick={() => setShowCreateForm(!showCreateForm)}
          className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white hover:bg-primary-500"
        >
          {showCreateForm ? 'Cancel' : 'Create Snapshot'}
        </button>
      </div>

      {error && (
        <div className="mb-4 rounded-lg border border-red-500/20 bg-red-500/10 px-4 py-3 text-sm text-red-400">
          {error}
          <button
            onClick={() => setError(null)}
            className="ml-3 text-red-300 underline hover:text-red-200"
          >
            Dismiss
          </button>
        </div>
      )}

      {showCreateForm && (
        <form
          onSubmit={handleCreate}
          className="mb-6 rounded-xl border border-gray-800 bg-gray-800/30 p-5"
        >
          <h2 className="mb-4 text-sm font-semibold text-white">New Factor Snapshot</h2>
          <div className="grid gap-4 sm:grid-cols-3">
            <label className="block">
              <span className="mb-1 block text-[11px] font-medium uppercase tracking-wider text-gray-500">
                Snapshot Name
              </span>
              <input
                type="text"
                required
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
                placeholder="e.g. FY2025 Annual"
                className="w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-white placeholder-gray-600 focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
              />
            </label>
            <label className="block">
              <span className="mb-1 block text-[11px] font-medium uppercase tracking-wider text-gray-500">
                Period Start
              </span>
              <input
                type="date"
                required
                value={formPeriodStart}
                onChange={(e) => setFormPeriodStart(e.target.value)}
                className="w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-white focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
              />
            </label>
            <label className="block">
              <span className="mb-1 block text-[11px] font-medium uppercase tracking-wider text-gray-500">
                Period End
              </span>
              <input
                type="date"
                required
                value={formPeriodEnd}
                onChange={(e) => setFormPeriodEnd(e.target.value)}
                className="w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-white focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
              />
            </label>
          </div>
          <div className="mt-4 flex justify-end gap-2">
            <button
              type="button"
              onClick={() => setShowCreateForm(false)}
              className="rounded-lg border border-gray-700 px-4 py-2 text-sm font-medium text-gray-400 hover:text-white"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={creating}
              className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white hover:bg-primary-500 disabled:opacity-50"
            >
              {creating ? 'Creating\u2026' : 'Create Snapshot'}
            </button>
          </div>
        </form>
      )}

      {loading ? (
        <div className="py-20 text-center text-gray-500">Loading snapshots...</div>
      ) : snapshots.length === 0 ? (
        <div className="rounded-xl border border-gray-800 bg-gray-800/30 p-12 text-center">
          <div className="mb-3 text-3xl">&#128247;</div>
          <h3 className="mb-2 text-base font-semibold text-white">No Factor Snapshots</h3>
          <p className="mb-4 text-sm text-gray-400">
            Create a snapshot to lock emission factors for a reporting period.
            Locked snapshots ensure audit reproducibility across reporting cycles.
          </p>
          <button
            onClick={() => setShowCreateForm(true)}
            className="inline-block rounded-lg bg-primary-600 px-5 py-2 text-sm font-medium text-white hover:bg-primary-500"
          >
            Create Your First Snapshot
          </button>
        </div>
      ) : (
        <div className="space-y-2">
          {/* Column headers */}
          <div className="grid grid-cols-12 gap-2 px-4 py-2 text-[10px] uppercase tracking-wider text-gray-600">
            <div className="col-span-3">Name</div>
            <div className="col-span-3">Period</div>
            <div className="col-span-2">Status</div>
            <div className="col-span-1 text-right">Factors</div>
            <div className="col-span-2">Locked Date</div>
            <div className="col-span-1">Actions</div>
          </div>

          {snapshots.map((snapshot) => {
            const s = statusConfig[snapshot.status] || statusConfig.draft;
            const isExpanded = expandedId === snapshot.id;

            return (
              <div key={snapshot.id}>
                <div
                  className={`grid grid-cols-12 items-center gap-2 rounded-lg border px-4 py-3 transition ${
                    isExpanded
                      ? 'border-primary-600/40 bg-gray-800/50'
                      : 'border-gray-800 bg-gray-800/30 hover:border-gray-700'
                  }`}
                >
                  <button
                    onClick={() => handleExpand(snapshot.id)}
                    className="col-span-3 flex items-center gap-2 text-left"
                  >
                    <svg
                      className={`h-3 w-3 flex-shrink-0 text-gray-500 transition-transform ${
                        isExpanded ? 'rotate-90' : ''
                      }`}
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      strokeWidth={2}
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
                    </svg>
                    <span className="truncate text-sm font-medium text-white">
                      {snapshot.name}
                    </span>
                  </button>

                  <div className="col-span-3 text-xs text-gray-400">
                    {formatDate(snapshot.period_start)} &ndash; {formatDate(snapshot.period_end)}
                  </div>

                  <div className="col-span-2">
                    <span
                      className={`inline-block rounded px-2 py-0.5 text-[10px] font-medium ${s.bg} ${s.text}`}
                    >
                      {s.label}
                    </span>
                  </div>

                  <div className="col-span-1 text-right text-xs text-gray-300">
                    {snapshot.factor_count}
                  </div>

                  <div className="col-span-2 text-xs text-gray-500">
                    {snapshot.locked_at ? formatDate(snapshot.locked_at) : '\u2014'}
                  </div>

                  <div className="col-span-1">
                    {snapshot.status === 'draft' && (
                      <button
                        onClick={() => handleLock(snapshot.id)}
                        disabled={lockingId === snapshot.id}
                        className="rounded-lg bg-green-600/20 px-2.5 py-1 text-[11px] font-medium text-green-400 hover:bg-green-600/30 disabled:opacity-50"
                      >
                        {lockingId === snapshot.id ? 'Locking\u2026' : 'Lock'}
                      </button>
                    )}
                  </div>
                </div>

                {/* Expanded factor details */}
                {isExpanded && (
                  <div className="ml-5 rounded-b-lg border border-t-0 border-gray-800 bg-gray-900/50 p-4">
                    {expandLoading ? (
                      <div className="py-6 text-center text-xs text-gray-500">
                        Loading factor details...
                      </div>
                    ) : expandedFactors.length === 0 ? (
                      <div className="py-6 text-center text-xs text-gray-500">
                        No factors attached to this snapshot.
                      </div>
                    ) : (
                      <div>
                        <div className="mb-3 flex items-center justify-between">
                          <h4 className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                            Snapshot Factors ({expandedFactors.length})
                          </h4>
                        </div>
                        <div className="overflow-x-auto">
                          <table className="w-full text-left">
                            <thead>
                              <tr className="border-b border-gray-800 text-[10px] uppercase tracking-wider text-gray-600">
                                <th className="pb-2 pr-4 font-medium">Scope</th>
                                <th className="pb-2 pr-4 font-medium">Category</th>
                                <th className="pb-2 pr-4 font-medium">Region</th>
                                <th className="pb-2 pr-4 font-medium">Source</th>
                                <th className="pb-2 pr-4 font-medium">Unit</th>
                                <th className="pb-2 pr-4 text-right font-medium">
                                  kg CO2e / unit
                                </th>
                              </tr>
                            </thead>
                            <tbody className="divide-y divide-gray-800/50">
                              {expandedFactors.map((f) => (
                                <tr key={f.id} className="text-xs">
                                  <td className="py-2 pr-4">
                                    <span
                                      className={`rounded px-1.5 py-0.5 text-[10px] font-medium ${
                                        f.scope === 'scope1'
                                          ? 'bg-red-500/10 text-red-400'
                                          : f.scope === 'scope2'
                                            ? 'bg-amber-500/10 text-amber-400'
                                            : 'bg-blue-500/10 text-blue-400'
                                      }`}
                                    >
                                      {f.scope.replace('scope', 'S')}
                                    </span>
                                  </td>
                                  <td className="py-2 pr-4 text-gray-300">
                                    {f.category || '\u2014'}
                                  </td>
                                  <td className="py-2 pr-4 text-gray-400">
                                    {f.region || 'Global'}
                                  </td>
                                  <td className="py-2 pr-4 text-gray-400">
                                    {f.source || '\u2014'}
                                  </td>
                                  <td className="py-2 pr-4 text-gray-500">{f.unit}</td>
                                  <td className="py-2 pr-4 text-right font-mono text-gray-300">
                                    {f.value_kg_co2e.toFixed(6)}
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      </div>
                    )}
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
