'use client';

import { useEffect, useState, useCallback, type ChangeEvent } from 'react';
import Link from 'next/link';
import { api, ApiRequestError } from '@/lib/api';
import type { Scope2Emission, Scope2Summary, PaginatedResponse, PageInfo } from '@/lib/types';
import { useRequireAuth } from '@/lib/session';
import { EmissionsTrendChart, ScopeBreakdownChart, EmissionsHeatmap } from '@/components/emissions';
import ErrorBoundary from '@/components/ErrorBoundary';
import { recordAuditEvent } from '@/lib/auditLog';
import type { TrendDataPoint } from '@/components/emissions/EmissionsTrendChart';
import type { ScopeData } from '@/components/emissions/ScopeBreakdownChart';

function buildTrendData(emissions: Scope2Emission[]): TrendDataPoint[] {
  const byMonth = new Map<string, TrendDataPoint>();

  for (const emission of emissions) {
    const date = new Date(emission.periodStart);
    if (Number.isNaN(date.getTime())) {
      continue;
    }

    const monthKey = new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), 1))
      .toISOString()
      .slice(0, 10);

    const existing = byMonth.get(monthKey) ?? {
      date: monthKey,
      scope1: 0,
      scope2: 0,
      scope3: 0,
      total: 0,
    };

    existing.scope2 += emission.emissionsTonsCO2e;
    existing.total += emission.emissionsTonsCO2e;
    byMonth.set(monthKey, existing);
  }

  return [...byMonth.values()].sort((left, right) => left.date.localeCompare(right.date));
}

function buildScopeBreakdown(emissions: Scope2Emission[]): { data: ScopeData[]; total: number } {
  const total = emissions.reduce((sum, emission) => sum + emission.emissionsTonsCO2e, 0);
  if (total === 0) {
    return { data: [], total: 0 };
  }

  return {
    data: [
      {
        scope: 'Scope 2',
        emissions: total,
        percentage: 100,
        activities: emissions.length,
      },
    ],
    total,
  };
}

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
  const trendData = buildTrendData(emissions);
  const scopeBreakdown = buildScopeBreakdown(emissions);

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
            <EmissionsTrendChart period="year" height={350} data={trendData} />
            <ScopeBreakdownChart
              height={350}
              startDate={startDate}
              endDate={endDate}
              data={scopeBreakdown.data}
              total={scopeBreakdown.total}
            />
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

type ParsedImportRow = {
  meterId: string;
  location: string;
  periodStart: string;
  periodEnd: string;
  kwh: number;
};

function parseImportCsv(raw: string): ParsedImportRow[] {
  const lines = raw
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean);

  if (lines.length < 2) {
    throw new Error('The CSV has no data rows. Add at least one row beneath the header and retry.');
  }

  const headers = lines[0].split(',').map((cell) => cell.trim().toLowerCase());
  const requiredHeaders = ['meter_id', 'location', 'period_start', 'period_end', 'kwh'];
  const missingHeaders = requiredHeaders.filter((header) => !headers.includes(header));

  if (missingHeaders.length > 0) {
    throw new Error(
      `The CSV is missing required columns: ${missingHeaders.join(', ')}. ` +
        'Expected columns: meter_id, location, period_start, period_end, kwh.',
    );
  }

  const columnIndex = Object.fromEntries(headers.map((header, index) => [header, index]));

  return lines.slice(1).map((line, rowIndex) => {
    const cells = line.split(',').map((cell) => cell.trim());
    const meterId = cells[columnIndex.meter_id] ?? '';
    const location = cells[columnIndex.location] ?? '';
    const periodStart = cells[columnIndex.period_start] ?? '';
    const periodEnd = cells[columnIndex.period_end] ?? '';
    const kwhRaw = cells[columnIndex.kwh] ?? '';
    const kwh = Number(kwhRaw);

    if (!meterId || !location || !periodStart || !periodEnd || !kwhRaw) {
      throw new Error(
        `Row ${rowIndex + 2} is incomplete. Every row must include meter_id, location, period_start, period_end, and kwh.`,
      );
    }

    if (!Number.isFinite(kwh) || kwh < 0) {
      throw new Error(`Row ${rowIndex + 2} has an invalid kwh value: "${kwhRaw}".`);
    }

    return {
      meterId,
      location,
      periodStart,
      periodEnd,
      kwh,
    };
  });
}

function EmptyState() {
  const [uploading, setUploading] = useState(false);
  const [uploadResult, setUploadResult] = useState<{ status: string; activitiesSaved: number } | null>(null);
  const [uploadError, setUploadError] = useState<string | null>(null);

  const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

  const handleFileUpload = async (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Client-side validation with clear, actionable errors.
    const MAX_SIZE_BYTES = 50 * 1024 * 1024; // 50MB
    const ACCEPTED_EXTENSIONS = ['.csv'];

    const hasAcceptedExtension = ACCEPTED_EXTENSIONS.some((ext) =>
      file.name.toLowerCase().endsWith(ext),
    );
    if (!hasAcceptedExtension) {
      setUploadError(
        `File "${file.name}" is not a CSV. Please upload a file ending in .csv. ` +
          `Expected columns: meter_id, location, period_start, period_end, kwh.`,
      );
      e.target.value = '';
      return;
    }
    if (file.size === 0) {
      setUploadError(
        `File "${file.name}" is empty. Please export your data and upload the CSV again.`,
      );
      e.target.value = '';
      return;
    }
    if (file.size > MAX_SIZE_BYTES) {
      setUploadError(
        `File "${file.name}" is ${(file.size / 1024 / 1024).toFixed(1)} MB — ` +
          `larger than the 50 MB limit. Split the file into smaller batches by reporting period.`,
      );
      e.target.value = '';
      return;
    }

    setUploading(true);
    setUploadError(null);
    setUploadResult(null);

    recordAuditEvent('data.import.started', {
      entityType: 'csv_upload',
      metadata: {
        file_name: file.name,
        file_size_bytes: file.size,
      },
    });

    try {
      const rows = parseImportCsv(await file.text());
      let activitiesSaved = 0;
      const failures: string[] = [];
      const REQUEST_SPACING_MS = 250;
      const MAX_ATTEMPTS = 4;

      for (let index = 0; index < rows.length; index += 1) {
        const row = rows[index];
        const rowNumber = index + 2;

        for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt += 1) {
          try {
            await api.post('/api/emissions/activities', {
              name: `${row.meterId} (${row.location})`,
              type: 'electricity',
              value: row.kwh,
              unit: 'kWh',
              period_start: row.periodStart,
              period_end: row.periodEnd,
              meter_id: row.meterId,
              location: row.location,
            });
            activitiesSaved += 1;
            break;
          } catch (err) {
            const message = err instanceof Error ? err.message : 'Unknown import error';
            const rateLimited = err instanceof ApiRequestError && err.status === 429;

            if (rateLimited && attempt < MAX_ATTEMPTS) {
              await sleep(attempt * 1000);
              continue;
            }

            failures.push(`Row ${rowNumber}: ${message}`);
            break;
          }
        }

        if (index < rows.length - 1) {
          await sleep(REQUEST_SPACING_MS);
        }
      }

      const uploadStatus = failures.length > 0 ? 'partial' : 'success';

      setUploadResult({
        status: uploadStatus,
        activitiesSaved,
      });

      if (failures.length > 0) {
        const preview = failures.slice(0, 3).join(' ');
        const remainder = failures.length > 3 ? ` ${failures.length - 3} more row(s) failed.` : '';
        setUploadError(`Imported ${activitiesSaved} row(s), but some entries failed. ${preview}${remainder}`);
      }

      recordAuditEvent('data.import.completed', {
        entityType: 'csv_upload',
        metadata: {
          file_name: file.name,
          rows_parsed: rows.length,
          activities_saved: activitiesSaved,
          status: uploadStatus,
          failures: failures.length,
        },
      });
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Upload failed';
      setUploadError(message);
      recordAuditEvent('data.import.failed', {
        entityType: 'csv_upload',
        metadata: {
          file_name: file.name,
          message,
        },
      });
    } finally {
      setUploading(false);
      // Reset the input so the same file can be re-uploaded if the user
      // fixes an issue and retries.
      e.target.value = '';
    }
  };

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
      <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>&#128202;</div>
      <h3 style={{ fontSize: '1.3rem', marginBottom: '0.5rem', color: '#fff' }}>
        {uploadResult ? `${uploadResult.activitiesSaved} Activities Imported` : 'Get Started: Import Emissions Data'}
      </h3>

      {uploadResult ? (
        <div style={{ color: '#4ade80', marginBottom: '1.5rem' }}>
          Upload successful. Refresh the page to see your data.
          <br />
          <button
            onClick={() => window.location.reload()}
            style={{
              marginTop: '1rem',
              padding: '0.75rem 1.5rem',
              background: '#16a34a',
              color: '#fff',
              border: 'none',
              borderRadius: '8px',
              fontSize: '0.95rem',
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            Refresh Page
          </button>
        </div>
      ) : (
        <>
          <p style={{ color: '#888', marginBottom: '1.5rem', maxWidth: '500px', margin: '0 auto 1.5rem' }}>
            Upload a CSV file with your utility or energy data. Expected columns:
            <br />
            <code style={{ color: '#8aa9ff', fontSize: '0.8rem' }}>
              meter_id, location, period_start, period_end, kwh
            </code>
          </p>

          {uploadError && (
            <div style={{ color: '#ff6b6b', marginBottom: '1rem', fontSize: '0.9rem' }}>
              {uploadError}
            </div>
          )}

          <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center', flexWrap: 'wrap' }}>
            <label
              style={{
                padding: '0.75rem 1.5rem',
                background: '#16a34a',
                color: '#fff',
                border: 'none',
                borderRadius: '8px',
                fontSize: '0.95rem',
                fontWeight: 600,
                cursor: uploading ? 'wait' : 'pointer',
                opacity: uploading ? 0.6 : 1,
              }}
            >
              {uploading ? 'Uploading...' : 'Upload CSV File'}
              <input
                type="file"
                accept=".csv"
                onChange={handleFileUpload}
                disabled={uploading}
                style={{ display: 'none' }}
              />
            </label>
            <Link
              href="/settings/data-sources"
              style={{
                padding: '0.75rem 1.5rem',
                background: 'transparent',
                color: '#8aa9ff',
                border: '1px solid #1d2937',
                borderRadius: '8px',
                fontSize: '0.95rem',
                textDecoration: 'none',
              }}
            >
              Configure Cloud Connectors
            </Link>
          </div>
        </>
      )}
    </div>
  );
}
