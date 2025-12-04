'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { useRequireAuth } from '../../lib/session';
import { getSubscription, SubscriptionResponse, formatSubscriptionStatus, formatPeriodEnd } from '../../lib/billing';

export default function SettingsPage() {
  const session = useRequireAuth();
  const [billingStatus, setBillingStatus] = useState<SubscriptionResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadData = async () => {
      if (!session.isAuthenticated) {
        setLoading(false);
        return;
      }
      try {
        const billing = await getSubscription();
        setBillingStatus(billing);
      } catch {
        // Billing status may fail if billing is not configured yet
      } finally {
        setLoading(false);
      }
    };

    if (!session.loading) {
      void loadData();
    }
  }, [session]);

  if (loading || session.loading) {
    return (
      <div style={{ padding: '2rem' }}>
        <h1>Settings</h1>
        <p style={{ color: '#888' }}>Loading...</p>
      </div>
    );
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <h1 style={{ margin: 0 }}>Settings</h1>
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

      <div style={{ display: 'grid', gap: '1.5rem', maxWidth: '800px' }}>
        {/* Account Section */}
        <SettingsSection title="Account">
          {session.isAuthenticated && session.user ? (
            <>
              <SettingsRow label="Email" value={session.user.email} />
              <SettingsRow label="Name" value={session.user.name} />
              <SettingsRow label="Role" value={session.user.role} />
              <SettingsRow label="Current Tenant" value={session.currentTenantId ?? 'Not set'} />
              <div style={{ marginTop: '1rem', display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                <Link
                  href="/settings/security"
                  style={{
                    padding: '0.5rem 1rem',
                    background: '#1d2940',
                    color: '#8aa9ff',
                    borderRadius: '6px',
                    textDecoration: 'none',
                    fontSize: '0.85rem',
                  }}
                >
                  Security & 2FA
                </Link>
                <Link
                  href="/settings/users"
                  style={{
                    padding: '0.5rem 1rem',
                    background: '#1d2940',
                    color: '#8aa9ff',
                    borderRadius: '6px',
                    textDecoration: 'none',
                    fontSize: '0.85rem',
                  }}
                >
                  Manage Users
                </Link>
                <Link
                  href="/settings/organization"
                  style={{
                    padding: '0.5rem 1rem',
                    background: '#3b82f6',
                    color: '#fff',
                    borderRadius: '6px',
                    textDecoration: 'none',
                    fontSize: '0.85rem',
                  }}
                >
                  Organization Admin
                </Link>
                <Link
                  href="/settings/data-sources"
                  style={{
                    padding: '0.5rem 1rem',
                    background: '#1d2940',
                    color: '#8aa9ff',
                    borderRadius: '6px',
                    textDecoration: 'none',
                    fontSize: '0.85rem',
                  }}
                >
                  Connectors & Ingestion
                </Link>
              </div>
            </>
          ) : (
            <div>
              <p style={{ color: '#888', marginBottom: '1rem' }}>You are not signed in.</p>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <Link
                  href="/login"
                  style={{
                    padding: '0.5rem 1rem',
                    background: '#3b82f6',
                    color: '#fff',
                    borderRadius: '6px',
                    textDecoration: 'none',
                    fontSize: '0.85rem',
                  }}
                >
                  Sign In
                </Link>
                <Link
                  href="/register"
                  style={{
                    padding: '0.5rem 1rem',
                    background: '#1d2940',
                    color: '#8aa9ff',
                    borderRadius: '6px',
                    textDecoration: 'none',
                    fontSize: '0.85rem',
                  }}
                >
                  Create Account
                </Link>
              </div>
            </div>
          )}
        </SettingsSection>

        {/* Subscription Section */}
        <SettingsSection title="Subscription">
          {billingStatus ? (
            <>
              <SettingsRow
                label="Plan"
                value={billingStatus.plan_id ? billingStatus.plan_id : 'None'}
              />
              <SettingsRow label="Status" value={formatSubscriptionStatus(billingStatus.status)} />
              <SettingsRow label="Next billing date" value={formatPeriodEnd(billingStatus.current_period_end)} />
              <SettingsRow
                label="Seats"
                value={`${billingStatus.seats_used ?? 0} / ${billingStatus.seats_included ?? '—'}`}
              />
              <div style={{ marginTop: '1rem' }}>
                <Link
                  href="/settings/billing"
                  style={{
                    padding: '0.5rem 1rem',
                    background: '#3b82f6',
                    color: '#fff',
                    borderRadius: '6px',
                    textDecoration: 'none',
                    fontSize: '0.85rem',
                  }}
                >
                  Manage Subscription
                </Link>
              </div>
            </>
          ) : (
            <div>
              <p style={{ color: '#888', marginBottom: '1rem' }}>
                {session.isAuthenticated ? 'Loading subscription status...' : 'Sign in to view subscription.'}
              </p>
              {session.isAuthenticated && (
                <Link
                  href="/settings/billing"
                  style={{
                    padding: '0.5rem 1rem',
                    background: '#3b82f6',
                    color: '#fff',
                    borderRadius: '6px',
                    textDecoration: 'none',
                    fontSize: '0.85rem',
                  }}
                >
                  View Plans
                </Link>
              )}
            </div>
          )}
        </SettingsSection>

        {/* Data Sources Section */}
        <SettingsSection title="Data Sources">
          <p style={{ color: '#888', marginBottom: '1rem' }}>
            Configure cloud providers and data integrations for automatic emissions ingestion.
          </p>
          <div style={{ display: 'grid', gap: '0.5rem' }}>
            <DataSourceRow name="AWS CUR / Cost Explorer" status="not_configured" />
            <DataSourceRow name="Azure Emissions Impact Dashboard" status="not_configured" />
            <DataSourceRow name="GCP BigQuery Carbon Footprint" status="not_configured" />
            <DataSourceRow name="Utility Bill Import" status="available" />
          </div>
          <div style={{ marginTop: '1rem' }}>
            <Link
              href="/settings/data-sources"
              style={{
                padding: '0.5rem 1rem',
                background: '#1d2940',
                color: '#8aa9ff',
                borderRadius: '6px',
                textDecoration: 'none',
                fontSize: '0.85rem',
              }}
            >
              Configure Data Sources
            </Link>
          </div>
        </SettingsSection>

        {/* Emission Factors Section */}
        <SettingsSection title="Emission Factors">
          <p style={{ color: '#888', marginBottom: '1rem' }}>
            Customize emission factors for your organization or use defaults from EPA eGRID.
          </p>
          <SettingsRow label="Grid Factor Source" value="EPA eGRID 2022" />
          <SettingsRow label="Custom Factors" value="0 configured" />
          <div style={{ marginTop: '1rem' }}>
            <Link
              href="/settings/factors"
              style={{
                padding: '0.5rem 1rem',
                background: '#1d2940',
                color: '#8aa9ff',
                borderRadius: '6px',
                textDecoration: 'none',
                fontSize: '0.85rem',
              }}
            >
              Manage Factors
            </Link>
          </div>
        </SettingsSection>

        {/* API Access Section */}
        <SettingsSection title="API Access">
          <p style={{ color: '#888', marginBottom: '1rem' }}>
            Generate API keys for programmatic access to OffGridFlow.
          </p>
          <SettingsRow label="API Keys" value="0 active" />
          <div style={{ marginTop: '1rem', display: 'flex', gap: '0.5rem' }}>
            <button
              style={{
                padding: '0.5rem 1rem',
                background: '#1d2940',
                color: '#8aa9ff',
                border: '1px solid #374151',
                borderRadius: '6px',
                cursor: 'pointer',
                fontSize: '0.85rem',
              }}
            >
              Generate API Key
            </button>
            <a
              href="/docs/api-reference"
              target="_blank"
              style={{
                padding: '0.5rem 1rem',
                background: 'transparent',
                color: '#8aa9ff',
                borderRadius: '6px',
                textDecoration: 'none',
                fontSize: '0.85rem',
              }}
            >
              API Documentation →
            </a>
          </div>
        </SettingsSection>
      </div>
    </div>
  );
}

// Components
function SettingsSection({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div
      style={{
        padding: '1.5rem',
        border: '1px solid #1d2940',
        borderRadius: '12px',
      }}
    >
      <h2 style={{ margin: '0 0 1rem 0', fontSize: '1.1rem', color: '#8aa9ff' }}>{title}</h2>
      {children}
    </div>
  );
}

function SettingsRow({ label, value }: { label: string; value: string }) {
  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'space-between',
        padding: '0.5rem 0',
        borderBottom: '1px solid #1d2940',
      }}
    >
      <span style={{ color: '#888' }}>{label}</span>
      <span style={{ fontWeight: 500 }}>{value}</span>
    </div>
  );
}

function DataSourceRow({ name, status }: { name: string; status: 'connected' | 'available' | 'not_configured' }) {
  const statusColors: Record<string, { bg: string; text: string; label: string }> = {
    connected: { bg: '#064e3b', text: '#4ade80', label: 'Connected' },
    available: { bg: '#1e3a5f', text: '#93c5fd', label: 'Available' },
    not_configured: { bg: '#374151', text: '#9ca3af', label: 'Not Configured' },
  };

  const colors = statusColors[status];

  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '0.75rem',
        background: '#0f172a',
        borderRadius: '6px',
      }}
    >
      <span>{name}</span>
      <span
        style={{
          padding: '0.25rem 0.5rem',
          background: colors.bg,
          color: colors.text,
          borderRadius: '4px',
          fontSize: '0.75rem',
        }}
      >
        {colors.label}
      </span>
    </div>
  );
}
