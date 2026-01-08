'use client';

import { useEffect, useMemo, useState } from 'react';
import { api } from '@/lib/api';
import type { ScheduleStatus } from '@/lib/types';
import styles from './page.module.css';

type Connector = {
  name: string;
  status: string;
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
  errors?: { message: string }[];
};

export default function DataSourcesPage() {
  const [connectors, setConnectors] = useState<Connector[]>([]);
  const [logs, setLogs] = useState<IngestionLog[]>([]);
  const [schedule, setSchedule] = useState<ScheduleStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [toast, setToast] = useState<string | null>(null);

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
          const scheduleData = await api.get<ScheduleStatus>('/api/connectors/schedule');
          setSchedule(scheduleData);
        } catch {
          setSchedule(null);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load connectors or logs');
      } finally {
        setLoading(false);
      }
    },
    []
  );

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const triggerRun = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await api.post<{ status: string }>('/api/connectors/run', {});
      if (res.status !== 'started' && res.status !== 'ok') {
        throw new Error('ingestion start failed');
      }
      setToast('Ingestion started');
      await fetchData();
    } catch (e) {
      setError('Failed to start ingestion');
      setToast(`Ingestion start failed: ${(e as Error).message}`);
    } finally {
      setLoading(false);
    }
  };

  const triggerSync = async (name: string) => {
    try {
      setLoading(true);
      setError(null);
      await api.post('/api/connectors/run', {});
      setToast(`${name} sync started`);
      await fetchData();
    } catch (e) {
      setError('Failed to start sync');
      setToast(`Sync failed: ${(e as Error).message}`);
    } finally {
      setLoading(false);
    }
  };

  const onTest = async (name: string) => {
    try {
      setLoading(true);
      setError(null);
      await api.post(`/api/connectors/test?name=${encodeURIComponent(name)}`, {});
      setToast(`${name} test succeeded`);
      await fetchData();
    } catch (e) {
      setError('Failed to test connector');
      setToast(`Connector test failed: ${(e as Error).message}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.headerRow}>
        <div>
          <p className={styles.eyebrow}>Data Sources</p>
          <h1 className={styles.title}>Cloud Connectors</h1>
          <p className={styles.muted}>
            Run ingestion, test connections, and monitor recent connector activity.
          </p>
        </div>
        <div className={styles.actions}>
          <button onClick={fetchData} disabled={loading} className={styles.secondaryButton}>
            {loading ? 'Refreshing…' : 'Refresh'}
          </button>
          <button onClick={triggerRun} disabled={loading} className={styles.primaryButton}>
            Run Ingestion Now
          </button>
        </div>
      </div>

      <div className={styles.scheduleCard}>
        <div className={styles.scheduleTitle}>Automated Sync</div>
        {schedule ? (
          <>
            <div>Interval: {schedule.interval || 'Manual only'}</div>
            <div>Last run: {formatScheduleTime(schedule.last_run_at)}</div>
            <div>Next run: {formatScheduleTime(schedule.next_run_at)}</div>
          </>
        ) : (
          <div className={styles.muted}>Scheduler not configured. Trigger ingestion manually below.</div>
        )}
      </div>

      {toast && <div className={styles.toast}>{toast}</div>}
      {error && <div className={styles.error}>{error}</div>}

      <div className={styles.grid}>
        {loading &&
          Array.from({ length: 3 }).map((_, idx) => (
            <div key={idx} className={`${styles.card} ${styles.skeletonCard}`} />
          ))}
        {!loading && connectors.length === 0 && (
          <div className={styles.emptyCard}>
            <p className={styles.muted}>No connectors yet. Add a source to begin ingestion.</p>
          </div>
        )}
        {!loading &&
          connectors.map((c) => (
            <div key={c.name} className={styles.card}>
              <div className={styles.cardHeader}>
                <span className={styles.cardTitle}>{c.name}</span>
                <span className={`${styles.badge} ${statusClass(c.status)}`}>{c.status}</span>
              </div>
              <div className={styles.meta}>
                Last sync: {c.last_run_at ? new Date(c.last_run_at).toLocaleString() : 'Not yet'}
              </div>
              {c.last_error && (
                <div className={styles.metaError}>
                  Error: {c.last_error} {c.last_error_at ? `(${new Date(c.last_error_at).toLocaleString()})` : ''}
                </div>
              )}
              <div className={styles.cardActions}>
                <button onClick={() => triggerSync(c.name)} className={styles.primaryButton} disabled={loading}>
                  Sync now
                </button>
                <button onClick={() => onTest(c.name)} className={styles.secondaryButton} disabled={loading}>
                  Test connection
                </button>
              </div>
            </div>
          ))}
      </div>

      <div className={styles.logsCard}>
        <div className={styles.logsHeader}>
          <h2 className={styles.logsTitle}>Recent Ingestion Runs</h2>
          <span className={styles.muted}>Latest 10 runs</span>
        </div>
        <div className={styles.tableWrapper}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th>Status</th>
                <th>Processed</th>
                <th>Succeeded</th>
                <th>Failed</th>
                <th>Started</th>
              </tr>
            </thead>
            <tbody>
              {logs.length === 0 && (
                <tr>
                  <td colSpan={5} className={styles.muted}>
                    No ingestion runs yet.
                  </td>
                </tr>
              )}
              {logs.map((log) => (
                <tr key={log.id}>
                  <td className={styles.statusCell}>
                    <span className={`${styles.badge} ${statusClass(log.status)}`}>{log.status}</span>
                  </td>
                  <td>{log.processed}</td>
                  <td>{log.succeeded}</td>
                  <td className={log.failed > 0 ? styles.metaError : undefined}>{log.failed}</td>
                  <td>{log.started_at ? new Date(log.started_at).toLocaleString() : ''}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function statusClass(status: string) {
  switch (status) {
    case 'connected':
      return styles.badgeSuccess;
    case 'running':
      return styles.badgeInfo;
    case 'error':
      return styles.badgeError;
    default:
      return styles.badgeMuted;
  }
}

function formatScheduleTime(value?: string) {
  if (!value) {
    return 'N/A';
  }
  const parsed = Date.parse(value);
  if (Number.isNaN(parsed)) {
    return value;
  }
  return new Date(parsed).toLocaleString();
}
