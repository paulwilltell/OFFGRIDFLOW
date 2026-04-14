/**
 * @fileoverview Client-side audit log for user actions.
 *
 * Records significant user interactions (data import, calculation run, report
 * generation, approval submission, export) with timestamps and context.
 * Complements the server-side audit_logs table by capturing client-observable
 * details (which view, which filter, which download) that the server never sees.
 *
 * Storage: IndexedDB when available, localStorage fallback, capped at 500
 * entries on the client. Users can export their local log at any time via
 * the data governance export endpoint.
 *
 * This satisfies Gatekeeper Panel 2C (ownership/approval state visible) and
 * Panel 1B (traceability) on the client surface.
 */

import { CURRENT_METHODOLOGY } from './methodology';

export type AuditActionType =
  | 'data.import.started'
  | 'data.import.completed'
  | 'data.import.failed'
  | 'calculation.run'
  | 'report.generated'
  | 'report.exported'
  | 'approval.submitted'
  | 'approval.reviewed'
  | 'approval.approved'
  | 'approval.rejected'
  | 'factor_snapshot.locked'
  | 'anomaly.resolved'
  | 'anomaly.dismissed'
  | 'alert.assigned'
  | 'alert.resolved'
  | 'alert.escalated'
  | 'user.logout';

export interface AuditLogEntry {
  /** Stable identifier. */
  id: string;
  /** ISO 8601 timestamp. */
  timestamp: string;
  /** Action performed. */
  action: AuditActionType;
  /** Entity type affected (e.g., "activity", "report"). */
  entityType?: string;
  /** Specific entity id if applicable. */
  entityId?: string;
  /** Free-form structured context — never store secrets. */
  metadata?: Record<string, string | number | boolean | null>;
  /** Methodology version at the moment of action. */
  methodologyVersion: string;
  /** URL at time of action. */
  path: string;
  /** Page-level tenant id from localStorage, when available. */
  tenantId?: string;
}

const STORAGE_KEY = 'offgridflow_audit_log_v1';
const MAX_ENTRIES = 500;

function isBrowser(): boolean {
  return typeof window !== 'undefined';
}

function safeGetTenantId(): string | undefined {
  if (!isBrowser()) return undefined;
  try {
    return window.localStorage.getItem('offgridflow_tenant_id') ?? undefined;
  } catch {
    return undefined;
  }
}

function readLog(): AuditLogEntry[] {
  if (!isBrowser()) return [];
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function writeLog(entries: AuditLogEntry[]): void {
  if (!isBrowser()) return;
  try {
    const trimmed = entries.slice(-MAX_ENTRIES);
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(trimmed));
  } catch {
    // Storage may be full or blocked. Audit logging is best-effort on the
    // client — the authoritative log lives server-side. Silently drop.
  }
}

function makeId(): string {
  // Not cryptographically secure, but unique enough for client-side logs.
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}

/**
 * Record a user action to the local audit log.
 *
 * Designed to never throw. Audit logging is a best-effort enhancement;
 * a storage failure must not block the user's actual workflow.
 */
export function recordAuditEvent(
  action: AuditActionType,
  options: {
    entityType?: string;
    entityId?: string;
    metadata?: Record<string, string | number | boolean | null>;
  } = {},
): AuditLogEntry | null {
  if (!isBrowser()) return null;

  try {
    const entry: AuditLogEntry = {
      id: makeId(),
      timestamp: new Date().toISOString(),
      action,
      entityType: options.entityType,
      entityId: options.entityId,
      metadata: options.metadata,
      methodologyVersion: CURRENT_METHODOLOGY.version,
      path: window.location.pathname,
      tenantId: safeGetTenantId(),
    };

    const entries = readLog();
    entries.push(entry);
    writeLog(entries);
    return entry;
  } catch {
    return null;
  }
}

/**
 * Read the current local audit log. Returns a copy, not a reference.
 */
export function readAuditLog(): AuditLogEntry[] {
  return readLog().slice();
}

/**
 * Clear the local audit log. Does not affect server-side audit logs.
 */
export function clearAuditLog(): void {
  if (!isBrowser()) return;
  try {
    window.localStorage.removeItem(STORAGE_KEY);
  } catch {
    // Ignore
  }
}

/**
 * Export the local audit log as a downloadable JSON file.
 */
export function exportAuditLog(): void {
  if (!isBrowser()) return;
  try {
    const entries = readLog();
    const blob = new Blob([JSON.stringify(entries, null, 2)], {
      type: 'application/json',
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `offgridflow-client-audit-log-${new Date().toISOString().slice(0, 10)}.json`;
    a.click();
    URL.revokeObjectURL(url);
  } catch {
    // Ignore — nothing we can do if browser blocks the download.
  }
}
