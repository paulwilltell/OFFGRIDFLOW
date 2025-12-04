'use client';

import Link from 'next/link';
import { useMemo } from 'react';
import styles from './page.module.css';
import { useSession } from '@/lib/session';

export default function OrganizationPage() {
  const session = useSession();

  const tenantName = useMemo(() => {
    return (
      session.tenants.find((t) => t.id === session.currentTenantId)?.name ||
      session.user?.tenants?.[0]?.name ||
      'Current tenant'
    );
  }, [session.tenants, session.currentTenantId, session.user]);

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <div>
          <p className={styles.eyebrow}>Organization</p>
          <h1 className={styles.title}>Organization settings</h1>
          <p className={styles.muted}>View tenant details and switch tenants. User management lives under “Users”.</p>
        </div>
        <Link href="/settings" className={styles.backLink}>
          ⟵ Back to settings
        </Link>
      </div>

      <div className={styles.card}>
        <h2 className={styles.cardTitle}>Tenant</h2>
        <div className={styles.tenantRow}>
          <div>
            <div className={styles.muted}>Active tenant</div>
            <div className={styles.tenantName}>{tenantName}</div>
            <div className={styles.meta}>ID: {session.currentTenantId ?? 'not set'}</div>
          </div>
          {session.tenants.length > 1 && (
            <select
              className={styles.select}
              value={session.currentTenantId ?? ''}
              onChange={(e) => session.switchTenant(e.target.value)}
            >
              {session.tenants.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.name}
                </option>
              ))}
            </select>
          )}
        </div>
      </div>

      <div className={styles.card}>
        <h2 className={styles.cardTitle}>Profile</h2>
        <div className={styles.profileGrid}>
          <div>
            <div className={styles.muted}>Name</div>
            <div>{session.user?.name || '—'}</div>
          </div>
          <div>
            <div className={styles.muted}>Email</div>
            <div>{session.user?.email || '—'}</div>
          </div>
          <div>
            <div className={styles.muted}>Role</div>
            <div>{session.user?.role || '—'}</div>
          </div>
        </div>
      </div>
    </div>
  );
}
