'use client';

import Link from 'next/link';
import { useRequireAuth } from '../../lib/session';
import ErrorBoundary from '../../components/ErrorBoundary';

export default function BlockchainDashboard() {
  const session = useRequireAuth();

  if (session.loading || !session.isAuthenticated) {
    return (
      <div style={{ padding: '2rem' }}>
        <h1>Blockchain</h1>
        <p style={{ color: '#888' }}>Checking your session...</p>
      </div>
    );
  }

  return (
    <ErrorBoundary>
      <div style={{ padding: '2rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '2rem' }}>
          <div>
            <h1 style={{ margin: 0 }}>Blockchain</h1>
            <p style={{ color: '#94a3b8', marginTop: '0.25rem' }}>Feature currently scheduled for a future release.</p>
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
             Dashboard
          </Link>
        </div>

        <div
          style={{
            padding: '2rem',
            borderRadius: '14px',
            border: '1px dashed #1f2937',
            background: '#0f172a',
            textAlign: 'center',
          }}
        >
          <div style={{ fontSize: '3rem', marginBottom: '0.5rem' }}>??</div>
          <h2 style={{ margin: '0.25rem 0', color: '#fff' }}>Future Release</h2>
          <p style={{ color: '#94a3b8', marginBottom: '1.5rem' }}>
            The blockchain workspace is being de-prioritized for now while we focus on compliance, ingestion, and emissions analytics.
            We will revisit capabilities like NFT minting, carbon-credit trading, and on-chain verification in a later release.
          </p>
          <div style={{ display: 'flex', justifyContent: 'center', gap: '1rem', flexWrap: 'wrap' }}>
            <span
              style={{
                padding: '0.5rem 1rem',
                borderRadius: '999px',
                border: '1px solid #374151',
                fontSize: '0.85rem',
                color: '#cbd5f5',
              }}
            >
              Roadmap: Manufacturing NFT marketplace
            </span>
            <span
              style={{
                padding: '0.5rem 1rem',
                borderRadius: '999px',
                border: '1px solid #374151',
                fontSize: '0.85rem',
                color: '#cbd5f5',
              }}
            >
              Next review: Q2 2026
            </span>
          </div>
        </div>
      </div>
    </ErrorBoundary>
  );
}
