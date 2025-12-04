'use client';

import { useEffect, useState, useCallback } from 'react';
import Link from 'next/link';
import { api } from '../../lib/api';
import type { Scope2Emission, Scope2Summary, PaginatedResponse, PageInfo } from '../../lib/types';
import { useRequireAuth } from '../../lib/session';
import { EmissionsTrendChart, ScopeBreakdownChart, EmissionsHeatmap } from '../../components/emissions';
import ErrorBoundary from '../../components/ErrorBoundary';

export default function EmissionsPage() {
  const session = useRequireAuth();
  const [emissions, setEmissions] = useState<Scope2Emission[]>([]);
  const [summary, setSummary] = useState<Scope2Summary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pageInfo, setPageInfo] = useState<PageInfo | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);

  // Filters
  const [region, setRegion] = useState<string>('');
  const [startDate, setStartDate] = useState<string>('');
  const [endDate, setEndDate] = useState<string>('');

  const fetchEmissions = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      // Build query params
      const params = new URLSearchParams();
      params.set('page', String(currentPage));
      params.set('per_page', String(pageSize));
      if (region) params.set('region', region);
      if (startDate) params.set('start_date', startDate);
      if (endDate) params.set('end_date', endDate);

      const [emissionsRes, summaryRes] = await Promise.all([
        api.get<PaginatedResponse<Scope2Emission> | Scope2Emission[]>(
          `/api/emissions/scope2?${params.toString()}`
        ),
        api.get<Scope2Summary>('/api/emissions/scope2/summary'),
      ]);

      // Handle both paginated and array responses
      if (Array.isArray(emissionsRes)) {
        setEmissions(emissionsRes);
        setPageInfo(null);
      } else {
        setEmissions(emissionsRes.data);
        setPageInfo(emissionsRes.pageInfo);
      }

      setSummary(summaryRes);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to load emissions');
    } finally {
      setLoading(false);
    }
  }, [currentPage, pageSize, region, startDate, endDate]);

  useEffect(() => {
    if (!session.isAuthenticated) return;
    fetchEmissions();
  }, [fetchEmissions, session.isAuthenticated]);

  if (session.loading || !session.isAuthenticated) {
    return (
      <div style={{ padding: '2rem' }}>
        <h1>Emissions Explorer</h1>
        <p style={{ color: '#888' }}>Checking your session...</p>
      </div>
    );
  }

  // Calculate totals from loaded data (fallback if summary fails)
  const displayTotals = {
    emissions: summary?.totalEmissionsTonsCO2e ?? emissions.reduce((sum, e) => sum + e.emissionsTonsCO2e, 0),
    energy: summary?.totalKWh ?? emissions.reduce((sum, e) => sum + e.quantityKWh, 0),
    avgFactor: summary?.averageEmissionFactor ?? 0,
    count: summary?.activityCount ?? emissions.length,
  };

  // Get unique regions for filter dropdown
  const uniqueRegions = [...new Set(emissions.map((e) => e.region).filter(Boolean))];

  return (
    <ErrorBoundary>
      <div>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
          <div>
            <h1 style={{ margin: 0 }}>Emissions Explorer</h1>
            <p style={{ color: '#888', margin: '0.25rem 0 0 0', fontSize: '0.9rem' }}>
              Comprehensive emissions tracking and analytics
            </p>
          </div>
          <Link
            href="/"
            style={{
              padding: '0.5rem 1rem',
              background: '#1d2940',
              color: '#8aa9ff',
              borderRadius: '6px',
              textDecoration: 'none',
              fontSize: '0.85rem',
            }}
          >
            ← Dashboard
          </Link>
        </div>

      {/* Error Banner */}
      {error && (
        <div
          style={{
            color: '#fecaca',
            padding: '1rem',
            background: '#7f1d1d',
            borderRadius: '8px',
            marginBottom: '1rem',
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
          }}
        >
          <span style={{ fontSize: '1.5rem' }}>⚠️</span>
          <div>
            <div style={{ fontWeight: 600, marginBottom: '0.25rem' }}>Error Loading Data</div>
            <div style={{ fontSize: '0.9rem' }}>{error}</div>
          </div>
        </div>
      )}

      {/* Visualizations Section */}
      {!loading && emissions.length > 0 && (
        <div style={{ marginBottom: '2rem' }}>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(500px, 1fr))', gap: '1.5rem', marginBottom: '1.5rem' }}>
            <EmissionsTrendChart period="year" height={350} />
            <ScopeBreakdownChart height={350} startDate={startDate} endDate={endDate} />
          </div>
          <EmissionsHeatmap period="week" height={350} />
        </div>
      )}

      {/* Summary Cards */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
          gap: '1rem',
          marginBottom: '1.5rem',
        }}
      >
        <SummaryCard
          label="Total Scope 2 Emissions"
          value={`${displayTotals.emissions.toFixed(2)} tCO2e`}
          loading={loading}
        />
        <SummaryCard
          label="Total Energy Consumed"
          value={`${displayTotals.energy.toLocaleString()} kWh`}
          loading={loading}
        />
        <SummaryCard
          label="Avg Emission Factor"
          value={`${displayTotals.avgFactor.toFixed(4)} kg/kWh`}
          loading={loading}
        />
        <SummaryCard
          label="Activity Records"
          value={String(displayTotals.count)}
          loading={loading}
        />
      </div>

      {/* Region Breakdown */}
      {summary?.regionBreakdown && Object.keys(summary.regionBreakdown).length > 0 && (
        <div style={{ marginBottom: '1.5rem' }}>
          <h2 style={{ fontSize: '1rem', marginBottom: '0.75rem' }}>Emissions by Region</h2>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
            {Object.entries(summary.regionBreakdown).map(([regionName, tons]) => (
              <div
                key={regionName}
                style={{
                  padding: '0.5rem 0.75rem',
                  background: '#1d2940',
                  borderRadius: '6px',
                  fontSize: '0.85rem',
                }}
              >
                <span style={{ color: '#8aa9ff' }}>{regionName}:</span>{' '}
                <span style={{ fontWeight: 600 }}>{tons.toFixed(2)} tCO2e</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Filters */}
      <div className="flex flex-wrap gap-4 mb-4 p-4 rounded-lg" style={{ background: '#0f172a' }}>
        <FilterSelect
          label="Region"
          value={region}
          onChange={setRegion}
          options={[{ value: '', label: 'All Regions' }, ...uniqueRegions.map((r) => ({ value: r, label: r }))]}
        />
        <FilterInput label="Start Date" type="date" value={startDate} onChange={setStartDate} />
        <FilterInput label="End Date" type="date" value={endDate} onChange={setEndDate} />
        <FilterSelect
          label="Rows"
          value={String(pageSize)}
          onChange={(v) => {
            setCurrentPage(1);
            setPageSize(Number(v));
          }}
          options={[
            { value: '10', label: '10 / page' },
            { value: '25', label: '25 / page' },
            { value: '50', label: '50 / page' },
          ]}
        />
        <button
          onClick={() => {
            setRegion('');
            setStartDate('');
            setEndDate('');
            setCurrentPage(1);
          }}
          style={{
            alignSelf: 'flex-end',
            padding: '0.5rem 1rem',
            background: 'transparent',
            color: '#8aa9ff',
            border: '1px solid #374151',
            borderRadius: '6px',
            cursor: 'pointer',
            fontSize: '0.85rem',
          }}
        >
          Clear Filters
        </button>
      </div>

      {/* Data Table */}
      <h2 style={{ fontSize: '1rem', marginBottom: '0.75rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
        <span>Scope 2 Activities</span>
        {emissions.length > 0 && (
          <span style={{ fontSize: '0.85rem', color: '#888', fontWeight: 400 }}>
            ({emissions.length} records)
          </span>
        )}
      </h2>
      {loading ? (
        <div style={{ color: '#888' }}>
          <SkeletonTable rows={pageSize >= 25 ? 6 : 4} />
        </div>
      ) : emissions.length === 0 ? (
        <EmptyState />
      ) : (
        <>
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', minWidth: '800px' }}>
              <thead>
                <tr style={{ borderBottom: '2px solid #1d2940' }}>
                  <th style={{ padding: '0.75rem 0.5rem', textAlign: 'left' }}>Meter ID</th>
                  <th style={{ padding: '0.75rem 0.5rem', textAlign: 'left' }}>Region</th>
                  <th style={{ padding: '0.75rem 0.5rem', textAlign: 'right' }}>kWh</th>
                  <th style={{ padding: '0.75rem 0.5rem', textAlign: 'right' }}>tCO2e</th>
                  <th style={{ padding: '0.75rem 0.5rem', textAlign: 'right' }}>Factor</th>
                  <th style={{ padding: '0.75rem 0.5rem', textAlign: 'left' }}>Method</th>
                  <th style={{ padding: '0.75rem 0.5rem', textAlign: 'left' }}>Period</th>
                </tr>
              </thead>
              <tbody>
                {emissions.map((row, idx) => (
                  <tr key={row.id || idx} style={{ borderTop: '1px solid #1d2940' }}>
                    <td style={{ padding: '0.75rem 0.5rem' }}>{row.meterId}</td>
                    <td style={{ padding: '0.75rem 0.5rem' }}>
                      <span
                        style={{
                          padding: '0.25rem 0.5rem',
                          background: '#1d2940',
                          borderRadius: '4px',
                          fontSize: '0.85rem',
                        }}
                      >
                        {row.region || row.location}
                      </span>
                    </td>
                    <td style={{ padding: '0.75rem 0.5rem', textAlign: 'right' }}>
                      {row.quantityKWh.toLocaleString()}
                    </td>
                    <td style={{ padding: '0.75rem 0.5rem', textAlign: 'right', fontWeight: 600 }}>
                      {row.emissionsTonsCO2e.toFixed(4)}
                    </td>
                    <td style={{ padding: '0.75rem 0.5rem', textAlign: 'right', color: '#888' }}>
                      {row.emissionFactor?.toFixed(4) ?? '-'}
                    </td>
                    <td style={{ padding: '0.75rem 0.5rem' }}>
                      <span
                        style={{
                          padding: '0.2rem 0.4rem',
                          background: row.methodology === 'market-based' ? '#1e3a5f' : '#0f3a2d',
                          borderRadius: '3px',
                          fontSize: '0.75rem',
                          color: row.methodology === 'market-based' ? '#93c5fd' : '#86efac',
                        }}
                      >
                        {row.methodology}
                      </span>
                    </td>
                    <td style={{ padding: '0.75rem 0.5rem', fontSize: '0.85rem', color: '#888' }}>
                      {formatDate(row.periodStart)} - {formatDate(row.periodEnd)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {pageInfo && pageInfo.totalPages > 1 && (
            <div style={{ display: 'flex', justifyContent: 'center', gap: '0.5rem', marginTop: '1rem' }}>
              <button
                onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                disabled={!pageInfo.hasPrev}
                style={{
                  padding: '0.5rem 1rem',
                  background: pageInfo.hasPrev ? '#1d2940' : '#0f172a',
                  color: pageInfo.hasPrev ? '#fff' : '#666',
                  border: 'none',
                  borderRadius: '6px',
                  cursor: pageInfo.hasPrev ? 'pointer' : 'not-allowed',
                }}
              >
                Previous
              </button>
              <span style={{ padding: '0.5rem 1rem', color: '#888' }}>
                Page {pageInfo.page} of {pageInfo.totalPages}
              </span>
              <button
                onClick={() => setCurrentPage((p) => p + 1)}
                disabled={!pageInfo.hasNext}
                style={{
                  padding: '0.5rem 1rem',
                  background: pageInfo.hasNext ? '#1d2940' : '#0f172a',
                  color: pageInfo.hasNext ? '#fff' : '#666',
                  border: 'none',
                  borderRadius: '6px',
                  cursor: pageInfo.hasNext ? 'pointer' : 'not-allowed',
                }}
              >
                Next
              </button>
            </div>
          )}
        </>
      )}
    </div>
    </ErrorBoundary>
  );
}

// Helper functions
function formatDate(isoString: string): string {
  if (!isoString) return '—';
  return new Date(isoString).toLocaleDateString();
}

// Components
function SummaryCard({ label, value, loading }: { label: string; value: string; loading: boolean }) {
  return (
    <div style={{ padding: '1rem', border: '1px solid #1d2940', borderRadius: '12px' }}>
      <div style={{ color: '#8aa9ff', fontSize: '0.9rem' }}>{label}</div>
      <div style={{ fontSize: '1.4rem', fontWeight: 700 }}>{loading ? '...' : value}</div>
    </div>
  );
}

function FilterSelect({
  label,
  value,
  onChange,
  options,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  options: { value: string; label: string }[];
}) {
  const inputId = `filter-select-${label.toLowerCase().replace(/\s+/g, '-')}`;
  return (
    <div>
      <label htmlFor={inputId} style={{ display: 'block', fontSize: '0.8rem', color: '#888', marginBottom: '0.25rem' }}>
        {label}
      </label>
      <select
        id={inputId}
        title={`Select ${label}`}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        style={{
          background: '#1d2940',
          color: '#fff',
          border: '1px solid #374151',
          borderRadius: '6px',
          padding: '0.5rem',
          minWidth: '150px',
        }}
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    </div>
  );
}

function FilterInput({
  label,
  type,
  value,
  onChange,
}: {
  label: string;
  type: string;
  value: string;
  onChange: (v: string) => void;
}) {
  const inputId = `filter-input-${label.toLowerCase().replace(/\s+/g, '-')}`;
  return (
    <div>
      <label htmlFor={inputId} style={{ display: 'block', fontSize: '0.8rem', color: '#888', marginBottom: '0.25rem' }}>
        {label}
      </label>
      <input
        id={inputId}
        type={type}
        title={`Enter ${label}`}
        placeholder={`Enter ${label.toLowerCase()}`}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        style={{
          background: '#1d2940',
          color: '#fff',
          border: '1px solid #374151',
          borderRadius: '6px',
          padding: '0.5rem',
        }}
      />
    </div>
  );
}

function SkeletonTable({ rows }: { rows: number }) {
  return (
    <div style={{ overflowX: 'auto' }}>
      <table style={{ width: '100%', borderCollapse: 'collapse', minWidth: '800px' }}>
        <thead>
          <tr style={{ borderBottom: '2px solid #1d2940' }}>
            {Array.from({ length: 7 }).map((_, idx) => (
              <th key={idx} style={{ padding: '0.75rem 0.5rem' }}>
                &nbsp;
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {Array.from({ length: rows }).map((_, idx) => (
            <tr key={idx} style={{ borderTop: '1px solid #1d2940' }}>
              {Array.from({ length: 7 }).map((__, cellIdx) => (
                <td key={cellIdx} style={{ padding: '0.75rem 0.5rem' }}>
                  <div
                    style={{
                      height: '12px',
                      width: '100%',
                      maxWidth: cellIdx === 0 ? '140px' : cellIdx === 3 ? '80px' : '120px',
                      background: 'linear-gradient(90deg, #111827 25%, #1f2937 37%, #111827 63%)',
                      backgroundSize: '400% 100%',
                      animation: 'shimmer 1.8s ease infinite',
                      borderRadius: '4px',
                    }}
                  />
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      <style jsx>{`
        @keyframes shimmer {
          0% {
            background-position: -200% 0;
          }
          100% {
            background-position: 200% 0;
          }
        }
      `}</style>
    </div>
  );
}

function EmptyState() {
  return (
    <div
      style={{
        padding: '4rem 2rem',
        textAlign: 'center',
        background: '#0f172a',
        borderRadius: '12px',
        border: '2px dashed #1d2937',
      }}
    >
      <div style={{ fontSize: '4rem', marginBottom: '1rem' }}>??</div>
      <h3 style={{ fontSize: '1.3rem', marginBottom: '0.5rem', color: '#fff' }}>No Emissions Data Available</h3>
      <p
        style={{
          color: '#888',
          marginBottom: '1.5rem',
          maxWidth: '500px',
          margin: '0 auto 1.5rem',
        }}
      >
        Connect a cloud connector or upload your utility bills to begin tracking energy usage. Once ingestion runs,
        emissions totals automatically surface on this dashboard.
      </p>
      <div
        style={{
          display: 'flex',
          gap: '1rem',
          justifyContent: 'center',
          flexWrap: 'wrap',
        }}
      >
        <button
          style={{
            padding: '0.75rem 1.5rem',
            background: '#8aa9ff',
            color: '#0a0f1e',
            border: 'none',
            borderRadius: '8px',
            fontSize: '0.95rem',
            fontWeight: 600,
            cursor: 'not-allowed',
            opacity: 0.6,
          }}
          disabled
        >
          ?? Upload Data (Coming Soon)
        </button>
        <button
          style={{
            padding: '0.75rem 1.5rem',
            background: 'transparent',
            color: '#8aa9ff',
            border: '1px solid #1d2937',
            borderRadius: '8px',
            fontSize: '0.95rem',
            cursor: 'not-allowed',
            opacity: 0.6,
          }}
          disabled
        >
          ?? View Documentation (Coming Soon)
        </button>
      </div>
    </div>
  );
}
