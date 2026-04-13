'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api } from '@/lib/api';
import { ComplianceSummary } from '@/lib/types';
import { useRequireAuth } from '@/lib/session';

export default function SECPage() {
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
        setError(err instanceof Error ? err.message : 'Failed to load SEC climate readiness');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [session.isAuthenticated]);

  if (session.loading || !session.isAuthenticated || loading) {
    return (
      <div style={{ padding: '2rem' }}>
        <h1>SEC Climate Rule</h1>
        <p style={{ color: '#888' }}>Loading your compliance status...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: '2rem' }}>
        <h1>SEC Climate Rule</h1>
        <div style={{ color: '#ff6b6b', padding: '1rem', background: '#1a1a2e', borderRadius: '8px' }}>
          {error}
        </div>
        <Link href="/" style={{ color: '#8aa9ff', marginTop: '1rem', display: 'inline-block' }}>
          ← Back to Dashboard
        </Link>
      </div>
    );
  }

  const sec = summary?.frameworks.sec;

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <h1 style={{ margin: 0 }}>SEC Climate Disclosure</h1>
          <p style={{ color: '#888', margin: '0.25rem 0 0 0', fontSize: '0.9rem' }}>
            Readiness for SEC climate-related disclosures.
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

      <div
        style={{
          padding: '1.25rem',
          border: '1px solid #1d2940',
          borderRadius: '12px',
          display: 'grid',
          gap: '0.75rem',
          maxWidth: '720px',
        }}
      >
        <div style={{ fontSize: '0.95rem', color: '#8aa9ff', fontWeight: 600 }}>Status</div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <StatusBadge status={sec?.status ?? 'not_started'} />
          <div style={{ fontSize: '1rem', fontWeight: 600 }}>
            {formatStatus(sec?.status ?? 'not_started')}
          </div>
        </div>
        <div style={{ color: '#888', fontSize: '0.9rem' }}>
          Prepare emissions disclosures, material risk assessment, and governance controls aligned to SEC rules.
        </div>
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
      style={{
        display: 'inline-block',
        width: '12px',
        height: '12px',
        borderRadius: '50%',
        background: color.bg,
        border: `2px solid ${color.text}`,
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
