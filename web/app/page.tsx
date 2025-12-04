'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { api } from '../lib/api';
import type {
  ModeResponse,
  ChatResponse,
  Scope2Summary,
  ComplianceSummary,
  FrameworkStatus,
} from '../lib/types';
import { useRequireAuth } from '../lib/session';
import styles from './page.module.css';

const statusColors: Record<string, { bg: string; text: string }> = {
  ok: { bg: '#064e3b', text: '#bbf7d0' },
  partial: { bg: '#92400e', text: '#fde68a' },
  no_data: { bg: '#374151', text: '#e5e7eb' },
  not_started: { bg: '#1f2937', text: '#9ca3af' },
  not_applicable: { bg: '#111827', text: '#6b7280' },
};

export default function DashboardPage() {
  const session = useRequireAuth();

  const [mode, setMode] = useState<string>('Loading…');
  const [modeError, setModeError] = useState<string | null>(null);
  const [chatResult, setChatResult] = useState<ChatResponse | null>(null);
  const [chatLoading, setChatLoading] = useState(false);
  const [chatError, setChatError] = useState<string | null>(null);

  const [scope2Summary, setScope2Summary] = useState<Scope2Summary | null>(null);
  const [complianceSummary, setComplianceSummary] = useState<ComplianceSummary | null>(null);
  const [dataLoading, setDataLoading] = useState(true);
  const [dataError, setDataError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<string | null>(null);

  useEffect(() => {
    if (!session.isAuthenticated) return;
    api
      .get<ModeResponse>('/api/offgrid/mode')
      .then((res) => setMode(res.mode))
      .catch((err) => setModeError(err.message));
  }, [session.isAuthenticated]);

  const fetchData = useMemo(
    () => async () => {
      setDataLoading(true);
      setDataError(null);
      try {
        const [scope2, compliance] = await Promise.all([
          api.get<Scope2Summary>('/api/emissions/scope2/summary'),
          api.get<ComplianceSummary>('/api/compliance/summary'),
        ]);
        setScope2Summary(scope2);
        setComplianceSummary(compliance);
        setLastUpdated(new Date().toISOString());
      } catch (err: unknown) {
        setDataError(err instanceof Error ? err.message : 'Failed to load data');
      } finally {
        setDataLoading(false);
      }
    },
    []
  );

  useEffect(() => {
    if (!session.isAuthenticated) return;
    fetchData();
  }, [session.isAuthenticated, fetchData]);

  const handleTestChat = async () => {
    setChatLoading(true);
    setChatError(null);
    try {
      const res = await api.post<ChatResponse>('/api/ai/chat', {
        prompt: 'Summarize my current emissions status and compliance readiness.',
      });
      setChatResult(res);
    } catch (err: unknown) {
      setChatError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setChatLoading(false);
    }
  };

  const totals = complianceSummary?.totals ?? { scope1Tons: 0, scope2Tons: 0, scope3Tons: 0 };
  const totalEmissions = totals.scope1Tons + totals.scope2Tons + totals.scope3Tons;
  const activityCount = scope2Summary?.activityCount ?? 0;
  const defaultFrameworkStatus: FrameworkStatus = { name: '', status: 'not_started' };
  const complianceStatus = complianceSummary?.frameworks ?? {
    csrd: defaultFrameworkStatus,
    sec: defaultFrameworkStatus,
    cbam: defaultFrameworkStatus,
    california: defaultFrameworkStatus,
  };

  if (session.loading || !session.isAuthenticated) {
    return (
      <div className={styles.container}>
        <h1>Dashboard</h1>
        <p className={styles.loadingText}>Checking your session…</p>
      </div>
    );
  }

  return (
    <div className={styles.dashboardGrid}>
      <div className={styles.headerRow}>
        <div>
          <p className={styles.eyebrow}>Operational Overview</p>
          <h1 className={styles.title}>OffGridFlow Control Center</h1>
          <p className={styles.muted}>
            Live emissions, compliance readiness, and system status for your organization.
          </p>
        </div>
        <div className={styles.buttonGroup}>
          <button onClick={fetchData} className={styles.secondaryButton} disabled={dataLoading}>
            {dataLoading ? 'Refreshing…' : 'Refresh data'}
          </button>
          <Link href="/emissions" className={styles.navLink}>
            Emissions →
          </Link>
          <Link href="/compliance/csrd" className={styles.navLink}>
            CSRD Report →
          </Link>
        </div>
      </div>

      <div className={styles.statGrid}>
        <div className={styles.card}>
          <div className={styles.label}>OffGrid Mode Status</div>
          {modeError ? (
            <div className={styles.error}>Error: {modeError}</div>
          ) : (
            <div className={`${styles.modeValue} ${mode === 'normal' ? styles.modeNormal : styles.modeWarning}`}>
              {mode.toUpperCase()}
            </div>
          )}
          <div className={styles.subText}>
            {mode === 'normal'
              ? 'Connected to cloud services'
              : mode === 'offline'
              ? 'Running in offline mode'
              : mode === 'degraded'
              ? 'Limited connectivity'
              : 'Checking status…'}
          </div>
        </div>

        <div className={styles.card}>
          <div className={styles.label}>Total GHG Emissions</div>
          {dataLoading ? (
            <div className={styles.loading}>Loading…</div>
          ) : dataError ? (
            <div className={styles.error}>{dataError}</div>
          ) : (
            <>
              <div className={styles.value}>
                {totalEmissions.toLocaleString(undefined, { maximumFractionDigits: 1 })} tCO2e
              </div>
              <div className={styles.subText}>
                {activityCount} activities recorded · updated{' '}
                {lastUpdated ? new Date(lastUpdated).toLocaleTimeString() : '—'}
              </div>
            </>
          )}
        </div>

        <div className={styles.card}>
          <div className={styles.label}>CSRD Readiness</div>
          {dataLoading ? (
            <div className={styles.loading}>Loading…</div>
          ) : (
            <>
              <div className={styles.statusRow}>
                <StatusBadge status={complianceStatus.csrd?.status ?? 'not_started'} />
                <span className={styles.statusLabel}>
                  {formatStatus(complianceStatus.csrd?.status ?? 'not_started')}
                </span>
              </div>
              <div className={styles.subText}>
                Scope 2: {complianceStatus.csrd?.scope2Ready ? '✅' : '—'} | Scope 1:{' '}
                {complianceStatus.csrd?.scope1Ready ? '✅' : '—'} | Scope 3:{' '}
                {complianceStatus.csrd?.scope3Ready ? '✅' : '—'}
              </div>
            </>
          )}
        </div>
      </div>

      <div className={styles.sectionHeaderRow}>
        <h2 className={styles.sectionHeader}>Emissions by Scope</h2>
        <p className={styles.muted}>Location-based Scope 2 with activity count and average factor.</p>
      </div>
      <div className={styles.scopeGrid}>
        <EmissionCard
          label="Scope 1 (Direct)"
          value={totals.scope1Tons}
          loading={dataLoading}
          description="Fuel combustion, fleet, onsite generation"
        />
        <EmissionCard
          label="Scope 2 (Purchased Energy)"
          value={totals.scope2Tons}
          loading={dataLoading}
          description="Electricity, heating, cooling"
        />
        <EmissionCard
          label="Scope 3 (Value Chain)"
          value={totals.scope3Tons}
          loading={dataLoading}
          description="Suppliers, logistics, use of products"
        />
        <div className={styles.card}>
          <div className={styles.label}>Energy Consumed</div>
          <div className={styles.value}>
            {dataLoading ? '…' : `${(scope2Summary?.totalKWh ?? 0).toLocaleString()} kWh`}
          </div>
          <div className={styles.description}>
            Avg factor: {scope2Summary?.averageEmissionFactor?.toFixed(3) ?? '-'} kgCO2e/kWh
          </div>
        </div>
      </div>

      <div className={styles.sectionHeaderRow}>
        <h2 className={styles.sectionHeader}>Compliance Frameworks</h2>
        <p className={styles.muted}>Deep links into climate disclosures across jurisdictions.</p>
      </div>
      <div className={styles.frameworkGrid}>
        {complianceSummary ? (
          Object.entries(complianceSummary.frameworks).map(([key, fw]) => (
            <FrameworkCard key={key} id={key} framework={fw} />
          ))
        ) : (
          <div className={styles.loading}>{dataLoading ? 'Loading frameworks…' : 'No data'}</div>
        )}
      </div>

      <div className={styles.chatCard}>
        <div className={styles.label}>AI Sustainability Assistant</div>
        <p className={styles.muted}>One-click summary of emissions and compliance posture.</p>
        <div className={styles.chatActions}>
          <button onClick={handleTestChat} disabled={chatLoading} className={styles.chatButton}>
            {chatLoading ? 'Analyzing…' : 'Get AI Summary'}
          </button>
          {lastUpdated && <span className={styles.muted}>Data as of {new Date(lastUpdated).toLocaleString()}</span>}
        </div>
        {chatError && <div className={styles.error}>Error: {chatError}</div>}
        {chatResult && (
          <div className={styles.chatResultContainer}>
            <div className={styles.chatSource}>
              Source:{' '}
              <span className={chatResult.source === 'cloud' ? styles.sourceCloud : styles.sourceLocal}>
                {chatResult.source}
              </span>
            </div>
            <div className={styles.chatResponse}>{chatResult.output}</div>
          </div>
        )}
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors = statusColors[status] ?? statusColors.not_started;
  return (
    <span
      className={styles.statusBadge}
      style={{
        background: colors.bg,
        borderColor: colors.text,
      }}
    />
  );
}

function formatStatus(status: string): string {
  const labels: Record<string, string> = {
    ok: 'Ready',
    partial: 'In Progress',
    no_data: 'No Data',
    not_started: 'Not Started',
    not_applicable: 'N/A',
  };
  return labels[status] ?? status;
}

function EmissionCard({
  label,
  value,
  loading,
  description,
}: {
  label: string;
  value: number;
  loading: boolean;
  description: string;
}) {
  return (
    <div className={styles.card}>
      <div className={styles.label}>{label}</div>
      <div className={styles.value}>
        {loading ? '…' : `${value.toLocaleString(undefined, { maximumFractionDigits: 1 })} tCO2e`}
      </div>
      <div className={styles.description}>{description}</div>
    </div>
  );
}

function FrameworkCard({ id, framework }: { id: string; framework: FrameworkStatus }) {
  const links: Record<string, string> = {
    csrd: '/compliance/csrd',
    sec: '/compliance/sec',
    cbam: '/compliance/cbam',
    california: '/compliance/california',
    ifrs_s2: '/compliance/ifrs',
  };

  return (
    <Link href={links[id] ?? '#'} className={styles.frameworkCard}>
      <div className={styles.frameworkHeader}>
        <StatusBadge status={framework.status} />
        <span className={styles.frameworkName}>{framework.name}</span>
      </div>
      <div className={styles.frameworkStatus}>{formatStatus(framework.status)}</div>
    </Link>
  );
}
