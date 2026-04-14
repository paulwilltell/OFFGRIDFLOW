'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api, ApiRequestError } from '@/lib/api';
import { useRequireAuth } from '@/lib/session';

interface LedgerEntry {
  id: string;
  scope: string;
  category: string;
  activity_id: string;
  quantity: number;
  unit: string;
  emission_factor_id: string;
  emission_factor_value: number;
  emission_factor_source: string;
  emission_factor_region: string;
  method: string;
  result_kg_co2e: number;
  result_tonnes_co2e: number;
  period_start: string;
  period_end: string;
  calculated_at: string;
  formula: string;
  notes: string;
  is_locked: boolean;
}

export default function LedgerPage() {
  useRequireAuth();
  const [entries, setEntries] = useState<LedgerEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [scopeFilter, setScopeFilter] = useState('');
  const [expanded, setExpanded] = useState<string | null>(null);
  const [apiError, setApiError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      setApiError(null);
      try {
        const params = scopeFilter ? `?scope=${scopeFilter}` : '';
        const res = await api.get<{ entries: LedgerEntry[]; count: number }>(`/api/audit/ledger${params}`);
        setEntries(res.entries || []);
      } catch (err) {
        setEntries([]);
        if (err instanceof ApiRequestError && err.status === 404) {
          setApiError('Audit ledger service is not available in this environment yet.');
        } else if (err instanceof ApiRequestError && err.status === 403) {
          setApiError('Your account does not currently have access to the audit ledger.');
        } else {
          setApiError('The audit ledger could not be loaded. Try again in a moment.');
        }
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [scopeFilter]);

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white">Calculation Ledger</h1>
          <p className="mt-1 text-xs text-gray-500">
            Immutable record of every emission calculation — traceable to source factor, method, and formula
          </p>
        </div>
        <div className="flex items-center gap-2">
          {['', 'scope1', 'scope2', 'scope3'].map(s => (
            <button
              key={s}
              onClick={() => setScopeFilter(s)}
              className={`rounded-lg px-3 py-1.5 text-xs font-medium transition ${
                scopeFilter === s ? 'bg-primary-600 text-white' : 'bg-gray-800 text-gray-400 hover:text-white'
              }`}
            >
              {s || 'All Scopes'}
            </button>
          ))}
        </div>
      </div>

      {loading ? (
        <div className="py-20 text-center text-gray-500">Loading ledger...</div>
      ) : apiError ? (
        <div className="rounded-xl border border-amber-500/30 bg-amber-500/10 p-12 text-center">
          <div className="mb-3 text-3xl">&#9888;</div>
          <h3 className="mb-2 text-base font-semibold text-white">Audit Ledger Unavailable</h3>
          <p className="mb-4 text-sm text-amber-100/80">{apiError}</p>
          <button
            onClick={() => window.location.reload()}
            className="inline-block rounded-lg bg-white/10 px-5 py-2 text-sm font-medium text-white hover:bg-white/20"
          >
            Retry
          </button>
        </div>
      ) : entries.length === 0 ? (
        <div className="rounded-xl border border-gray-800 bg-gray-800/30 p-12 text-center">
          <div className="mb-3 text-3xl">&#128220;</div>
          <h3 className="mb-2 text-base font-semibold text-white">No Calculations Recorded</h3>
          <p className="mb-4 text-sm text-gray-400">
            Upload emissions data to generate calculations. Each calculation is permanently recorded
            with its full formula, factor source, and method for audit traceability.
          </p>
          <Link
            href="/emissions"
            className="inline-block rounded-lg bg-primary-600 px-5 py-2 text-sm font-medium text-white hover:bg-primary-500"
          >
            Upload Data
          </Link>
        </div>
      ) : (
        <div className="space-y-2">
          {/* Column headers */}
          <div className="grid grid-cols-12 gap-2 px-4 py-2 text-[10px] uppercase tracking-wider text-gray-600">
            <div className="col-span-1">Scope</div>
            <div className="col-span-2">Activity</div>
            <div className="col-span-2">Quantity</div>
            <div className="col-span-2">Factor</div>
            <div className="col-span-2">Result</div>
            <div className="col-span-2">Period</div>
            <div className="col-span-1">Status</div>
          </div>

          {entries.map(entry => (
            <div key={entry.id}>
              <button
                onClick={() => setExpanded(expanded === entry.id ? null : entry.id)}
                className="w-full grid grid-cols-12 gap-2 rounded-lg border border-gray-800 bg-gray-800/30 px-4 py-3 text-left transition hover:border-gray-700"
              >
                <div className="col-span-1">
                  <span className={`rounded px-1.5 py-0.5 text-[10px] font-medium ${
                    entry.scope === 'scope1' ? 'bg-red-500/10 text-red-400' :
                    entry.scope === 'scope2' ? 'bg-amber-500/10 text-amber-400' :
                    'bg-blue-500/10 text-blue-400'
                  }`}>
                    {entry.scope.replace('scope', 'S')}
                  </span>
                </div>
                <div className="col-span-2 truncate text-xs text-gray-300">{entry.activity_id || '—'}</div>
                <div className="col-span-2 text-xs text-white">{entry.quantity.toFixed(2)} {entry.unit}</div>
                <div className="col-span-2 text-xs text-gray-400">× {entry.emission_factor_value.toFixed(6)}</div>
                <div className="col-span-2 text-xs font-medium text-white">{entry.result_tonnes_co2e.toFixed(4)} t</div>
                <div className="col-span-2 text-xs text-gray-500">
                  {entry.period_start ? `${entry.period_start.substring(0, 10)}` : '—'}
                </div>
                <div className="col-span-1">
                  {entry.is_locked ? (
                    <span className="text-[10px] text-green-400">Locked</span>
                  ) : (
                    <span className="text-[10px] text-gray-500">Draft</span>
                  )}
                </div>
              </button>

              {/* Expanded detail */}
              {expanded === entry.id && (
                <div className="ml-4 rounded-b-lg border border-t-0 border-gray-800 bg-gray-900/50 p-4">
                  {/* Formula */}
                  <div className="mb-4 rounded-lg bg-gray-800/50 p-3">
                    <div className="mb-1 text-[10px] uppercase tracking-wider text-gray-500">Formula</div>
                    <code className="text-xs text-primary-400">{entry.formula}</code>
                  </div>

                  <div className="grid gap-3 sm:grid-cols-3">
                    <div>
                      <div className="text-[10px] text-gray-500">Factor ID</div>
                      <div className="text-xs text-gray-300">{entry.emission_factor_id}</div>
                    </div>
                    <div>
                      <div className="text-[10px] text-gray-500">Factor Source</div>
                      <div className="text-xs text-gray-300">{entry.emission_factor_source}</div>
                    </div>
                    <div>
                      <div className="text-[10px] text-gray-500">Region</div>
                      <div className="text-xs text-gray-300">{entry.emission_factor_region || 'Global'}</div>
                    </div>
                    <div>
                      <div className="text-[10px] text-gray-500">Method</div>
                      <div className="text-xs text-gray-300">{entry.method}</div>
                    </div>
                    <div>
                      <div className="text-[10px] text-gray-500">Calculated At</div>
                      <div className="text-xs text-gray-300">{new Date(entry.calculated_at).toLocaleString()}</div>
                    </div>
                    <div>
                      <div className="text-[10px] text-gray-500">Result (kg CO2e)</div>
                      <div className="text-xs text-gray-300">{entry.result_kg_co2e.toFixed(4)}</div>
                    </div>
                  </div>

                  {entry.category && (
                    <div className="mt-3">
                      <div className="text-[10px] text-gray-500">Scope 3 Category</div>
                      <div className="text-xs text-gray-300">{entry.category}</div>
                    </div>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
