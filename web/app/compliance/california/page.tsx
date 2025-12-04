'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api } from '@/lib/api';
import { ComplianceSummary, FrameworkStatus } from '@/lib/types';
import { useRequireAuth } from '@/lib/session';
import styles from './page.module.css';

const statusCopy: Record<string, string> = {
  ok: 'Disclosure-ready for SB 253/261.',
  partial: 'Some gaps remain. Verify Scope 1/2/3 coverage and assurance.',
  no_data: 'No emissions data available. Import utility bills and activity data.',
  not_started: 'Start data collection and mapping to California requirements.',
  not_applicable: 'Not applicable to your entity.',
};

export default function CaliforniaPage() {
  const session = useRequireAuth();
  const [summary, setSummary] = useState<ComplianceSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!session.isAuthenticated) return;
    const load = async () => {
      setLoading(true);
      setError(null);
      try {
        const data = await api.get<ComplianceSummary>('/api/compliance/summary');
        setSummary(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load SB 253/261 readiness');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [session.isAuthenticated]);

  if (session.loading || !session.isAuthenticated) {
    return (
      <div className={styles.container}>
        <h1>California Climate Disclosure</h1>
        <p className={styles.muted}>Checking your session…</p>
      </div>
    );
  }

  if (loading) {
    return (
      <div className={styles.container}>
        <h1>California Climate Disclosure</h1>
        <p className={styles.muted}>Loading your SB 253/261 readiness…</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className={styles.container}>
        <h1>California Climate Disclosure</h1>
        <div className={styles.error}>{error}</div>
        <Link href="/" className={styles.backLink}>
          ← Back to Dashboard
        </Link>
      </div>
    );
  }

  const ca = summary?.frameworks.california;
  const csrd = summary?.frameworks.csrd;

  return (
    <div className={styles.container}>
      <div className={styles.headerRow}>
        <div>
          <p className={styles.eyebrow}>SB 253 / SB 261</p>
          <h1 className={styles.title}>California Climate Disclosure</h1>
          <p className={styles.muted}>
            Scope 1/2/3 disclosure readiness, assurance, and alignment to SB 253/261.
          </p>
        </div>
        <Link href="/" className={styles.navLink}>
          ← Dashboard
        </Link>
      </div>

      <div className={styles.statusCard}>
        <div className={styles.label}>Overall Status</div>
        <div className={styles.statusRow}>
          <StatusBadge status={ca?.status ?? 'not_started'} />
          <div className={styles.statusText}>{formatStatus(ca?.status ?? 'not_started')}</div>
        </div>
        <p className={styles.muted}>{statusCopy[ca?.status ?? 'not_started']}</p>
        <div className={styles.grid}>
          <ChecklistItem
            title="Scope 1"
            done={csrd?.scope1Ready ?? false}
            detail="Direct emissions captured and validated."
          />
          <ChecklistItem
            title="Scope 2"
            done={csrd?.scope2Ready ?? false}
            detail="Purchased energy with location-based factors."
          />
          <ChecklistItem
            title="Scope 3"
            done={csrd?.scope3Ready ?? false}
            detail="Value-chain emissions screened and quantified."
          />
          <ChecklistItem
            title="Assurance plan"
            done={ca?.status === 'ok'}
            detail="Plan for third-party assurance (SB 253)."
          />
        </div>
      </div>

      <div className={styles.section}>
        <h2 className={styles.sectionTitle}>Next actions</h2>
        <ul className={styles.actionList}>
          <li>Upload or connect utility and activity data to fill Scope 1/2 gaps.</li>
          <li>Screen Scope 3 categories and quantify material categories.</li>
          <li>Decide on assurance provider and timeline (limited → reasonable).</li>
          <li>Align disclosures with IFRS S2/CSRD where applicable for consistency.</li>
        </ul>
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, { bg: string; text: string }> = {
    ok: { bg: '#064e3b', text: '#4ade80' },
    partial: { bg: '#78350f', text: '#facc15' },
    no_data: { bg: '#374151', text: '#9ca3af' },
    not_started: { bg: '#1f2937', text: '#9ca3af' },
    not_applicable: { bg: '#1f2937', text: '#6b7280' },
  };
  const color = colors[status] ?? colors.not_started;
  return (
    <span
      className={styles.statusBadge}
      style={{
        background: color.bg,
        borderColor: color.text,
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

function ChecklistItem({ title, done, detail }: { title: string; done: boolean; detail: string }) {
  return (
    <div className={styles.checkItem}>
      <span className={done ? styles.checkDone : styles.checkTodo}>{done ? '✓' : '•'}</span>
      <div>
        <div className={styles.checkTitle}>{title}</div>
        <div className={styles.muted}>{detail}</div>
      </div>
    </div>
  );
}
