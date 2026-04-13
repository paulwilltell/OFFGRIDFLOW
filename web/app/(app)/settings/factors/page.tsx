'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import styles from './page.module.css';
import { api } from '@/lib/api';

type Factor = {
  id: string;
  scope: string;
  region: string;
  source: string;
  category: string;
  unit: string;
  valueKgCO2e: number;
  validFrom?: string;
  validTo?: string;
};

const scopes = ['scope1', 'scope2', 'scope3'];

export default function FactorsPage() {
  const [factors, setFactors] = useState<Factor[]>([]);
  const [scope, setScope] = useState('');
  const [region, setRegion] = useState('');
  const [category, setCategory] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useMemo(
    () => async () => {
      setLoading(true);
      setError(null);
      try {
        const params = new URLSearchParams();
        if (scope) params.set('scope', scope);
        if (region) params.set('region', region);
        if (category) params.set('category', category);
        const data = await api.get<Factor[]>(`/api/factors?${params.toString()}`);
        setFactors(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load factors');
      } finally {
        setLoading(false);
      }
    },
    [scope, region, category],
  );

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <div className={styles.container}>
      <div className={styles.headerRow}>
        <div>
          <p className={styles.eyebrow}>Emission Factors</p>
          <h1 className={styles.title}>Factor Library</h1>
          <p className={styles.muted}>Browse factors by scope, region, and category. Values are returned from the active registry.</p>
        </div>
        <Link href="/settings" className={styles.backLink}>
          ⟵ Back to settings
        </Link>
      </div>

      <div className={styles.filters}>
        <label className={styles.label}>
          Scope
          <select className={styles.input} value={scope} onChange={(e) => setScope(e.target.value)}>
            <option value="">All</option>
            {scopes.map((s) => (
              <option key={s} value={s}>
                {s.toUpperCase()}
              </option>
            ))}
          </select>
        </label>
        <label className={styles.label}>
          Region
          <input
            className={styles.input}
            value={region}
            onChange={(e) => setRegion(e.target.value)}
            placeholder="e.g. US-WEST"
          />
        </label>
        <label className={styles.label}>
          Category
          <input
            className={styles.input}
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            placeholder="electricity, diesel..."
          />
        </label>
        <button className={styles.primaryButton} onClick={load} disabled={loading}>
          {loading ? 'Loading…' : 'Apply'}
        </button>
      </div>

      {error && <div className={styles.error}>{error}</div>}

      <div className={styles.tableWrapper}>
        <table className={styles.table}>
          <thead>
            <tr>
              <th>ID</th>
              <th>Scope</th>
              <th>Region</th>
              <th>Category</th>
              <th>Source</th>
              <th>Unit</th>
              <th>kg CO2e/unit</th>
              <th>Valid</th>
            </tr>
          </thead>
          <tbody>
            {factors.map((f) => (
              <tr key={f.id}>
                <td className={styles.code}>{f.id}</td>
                <td>{f.scope}</td>
                <td>{f.region || '—'}</td>
                <td>{f.category || '—'}</td>
                <td>{f.source || '—'}</td>
                <td>{f.unit}</td>
                <td>{f.valueKgCO2e?.toFixed(4)}</td>
                <td>
                  {f.validFrom ? new Date(f.validFrom).toLocaleDateString() : '—'}{' '}
                  {f.validTo ? `→ ${new Date(f.validTo).toLocaleDateString()}` : ''}
                </td>
              </tr>
            ))}
            {factors.length === 0 && !loading && (
              <tr>
                <td colSpan={8} className={styles.muted}>
                  No factors match your filters.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
