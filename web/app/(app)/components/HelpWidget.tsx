'use client';

import { useState, useEffect, useRef } from 'react';

/**
 * Floating help widget for the authenticated app shell.
 *
 * Provides customers with an immediate, non-blocking support path:
 *  - Quick links to key self-service pages (methodology, status, governance)
 *  - Direct email to support with a pre-populated subject
 *  - "Report a Problem" shortcut that captures the current URL
 *
 * This reduces refund and chargeback pressure by ensuring customers always
 * have a one-click escalation path before going to their bank or lawyer.
 */

export function HelpWidget() {
  const [open, setOpen] = useState(false);
  const panelRef = useRef<HTMLDivElement | null>(null);

  // Close on Escape and on outside click.
  useEffect(() => {
    if (!open) return;
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') setOpen(false);
    }
    function onClick(e: MouseEvent) {
      if (!panelRef.current) return;
      if (!panelRef.current.contains(e.target as Node)) setOpen(false);
    }
    document.addEventListener('keydown', onKey);
    document.addEventListener('mousedown', onClick);
    return () => {
      document.removeEventListener('keydown', onKey);
      document.removeEventListener('mousedown', onClick);
    };
  }, [open]);

  const reportPath =
    typeof window !== 'undefined' ? `${window.location.pathname}${window.location.search}` : '';
  const reportSubject = encodeURIComponent(
    `[OffGridFlow] Issue report: ${reportPath || 'dashboard'}`,
  );
  const reportBody = encodeURIComponent(
    `URL: ${typeof window !== 'undefined' ? window.location.href : ''}
Date: ${new Date().toISOString()}

Please describe what you were trying to do and what went wrong.

---

(Support team: check audit logs and session id for this user in the window above.)`,
  );

  return (
    <>
      {/* Floating launcher */}
      <button
        type="button"
        aria-label={open ? 'Close help' : 'Open help'}
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
        className="fixed bottom-5 right-5 z-40 flex h-11 w-11 items-center justify-center rounded-full bg-primary-600 text-white shadow-lg transition hover:bg-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-400 focus:ring-offset-2 focus:ring-offset-dark-900"
      >
        {open ? (
          <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        ) : (
          <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093M12 17h.01M12 21a9 9 0 110-18 9 9 0 010 18z" />
          </svg>
        )}
      </button>

      {/* Panel */}
      {open && (
        <div
          ref={panelRef}
          role="dialog"
          aria-labelledby="help-widget-title"
          className="fixed bottom-20 right-5 z-40 w-80 rounded-xl border border-gray-700 bg-gray-900/95 p-4 shadow-2xl backdrop-blur-md"
        >
          <div className="mb-3">
            <h3 id="help-widget-title" className="text-sm font-semibold text-white">
              Need help?
            </h3>
            <p className="mt-0.5 text-xs text-gray-400">
              Self-service resources and a direct line to support.
            </p>
          </div>

          <div className="space-y-1 text-sm">
            <a
              href="/emissions"
              className="flex items-center justify-between rounded-md px-2 py-1.5 text-gray-300 hover:bg-gray-800 hover:text-white"
            >
              <span>Upload your data (get started)</span>
            </a>
            <a
              href="/trust"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center justify-between rounded-md px-2 py-1.5 text-gray-300 hover:bg-gray-800 hover:text-white"
            >
              <span>Trust Center (security &amp; RBAC)</span>
              <span className="text-[10px] text-gray-500">&#8599;</span>
            </a>
            <a
              href="/settings"
              className="flex items-center justify-between rounded-md px-2 py-1.5 text-gray-300 hover:bg-gray-800 hover:text-white"
            >
              <span>Account &amp; data settings</span>
            </a>
          </div>

          <div className="mt-4 space-y-2 border-t border-gray-800 pt-3">
            <a
              href={`mailto:contact@off-grid-flow.com?subject=${reportSubject}&body=${reportBody}`}
              className="block w-full rounded-lg bg-primary-600 px-3 py-2 text-center text-sm font-medium text-white hover:bg-primary-500"
            >
              Report a problem
            </a>
            <a
              href="mailto:contact@off-grid-flow.com?subject=OffGridFlow%20support"
              className="block w-full rounded-lg border border-gray-700 px-3 py-2 text-center text-sm text-gray-300 hover:border-gray-500 hover:text-white"
            >
              Email support
            </a>
          </div>

          <p className="mt-3 text-[10px] text-gray-600">
            Responses within one business day. Please email before initiating a chargeback
            &mdash; most issues are resolvable within hours.
          </p>
        </div>
      )}
    </>
  );
}

export default HelpWidget;
