'use client';

import Link from 'next/link';
import { useMemo, useState } from 'react';
import { useSession } from '@/lib/session';

const navLinks = [
  { href: '/', label: 'Dashboard' },
  { href: '/emissions', label: 'Emissions' },
  { href: '/compliance/csrd', label: 'CSRD' },
  { href: '/settings', label: 'Settings' },
];

export function AppHeader() {
  const { user, tenants, currentTenantId, switchTenant, logout, isAuthenticated } = useSession();
  const [switching, setSwitching] = useState(false);

  const tenantOptions = useMemo(() => tenants ?? [], [tenants]);

  const handleTenantChange = async (tenantId: string) => {
    if (!tenantId || tenantId === currentTenantId) return;
    setSwitching(true);
    try {
      await switchTenant(tenantId);
    } finally {
      setSwitching(false);
    }
  };

  return (
    <header
      style={{
        padding: '1rem 1.5rem',
        borderBottom: '1px solid #1d2940',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        gap: '1rem',
      }}
    >
      <Link
        href="/"
        style={{ fontWeight: 700, letterSpacing: '0.5px', textDecoration: 'none', color: 'inherit' }}
      >
        OffGridFlow
      </Link>

      <nav style={{ display: 'flex', gap: '1.5rem', alignItems: 'center' }}>
        {isAuthenticated &&
          navLinks.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              style={{
                color: '#8aa9ff',
                textDecoration: 'none',
                fontSize: '0.9rem',
              }}
            >
              {link.label}
            </Link>
          ))}
      </nav>

      <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
        {isAuthenticated && tenantOptions.length > 1 && (
          <select
            aria-label="Select tenant"
            value={currentTenantId ?? ''}
            onChange={(e) => handleTenantChange(e.target.value)}
            disabled={switching}
            style={{
              background: '#1d2940',
              color: '#fff',
              border: '1px solid #374151',
              borderRadius: '6px',
              padding: '0.4rem 0.6rem',
              minWidth: '160px',
            }}
          >
            {tenantOptions.map((tenant) => (
              <option key={tenant.id} value={tenant.id}>
                {tenant.name}
              </option>
            ))}
          </select>
        )}

        {isAuthenticated ? (
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <span style={{ fontSize: '0.9rem', color: '#8aa9ff' }}>{user?.email}</span>
            <button
              onClick={logout}
              style={{
                padding: '0.35rem 0.8rem',
                background: '#1d2940',
                color: '#8aa9ff',
                border: '1px solid #374151',
                borderRadius: '6px',
                cursor: 'pointer',
                fontSize: '0.85rem',
              }}
            >
              Sign out
            </button>
          </div>
        ) : (
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <Link
              href="/login"
              style={{
                padding: '0.4rem 0.9rem',
                background: '#3b82f6',
                color: '#fff',
                borderRadius: '6px',
                textDecoration: 'none',
                fontSize: '0.85rem',
              }}
            >
              Sign in
            </Link>
            <Link
              href="/register"
              style={{
                padding: '0.4rem 0.9rem',
                background: '#1d2940',
                color: '#8aa9ff',
                borderRadius: '6px',
                textDecoration: 'none',
                fontSize: '0.85rem',
              }}
            >
              Register
            </Link>
          </div>
        )}
      </div>
    </header>
  );
}
