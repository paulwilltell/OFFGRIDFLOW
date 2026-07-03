'use client';

import { useState, useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { getSubscription, createCheckoutSession, type SubscriptionResponse } from '@/lib/billing';
import { useRequireAuth } from '@/lib/session';

const REPORT_PRICE = '$149';

function BillingContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const session = useRequireAuth();

  const paid = searchParams.get('paid') === '1';

  const [status, setStatus] = useState<SubscriptionResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [checkoutLoading, setCheckoutLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!session.loading && !session.isAuthenticated) {
      router.replace('/login?returnTo=/settings/billing');
    }
  }, [session.loading, session.isAuthenticated, router]);

  useEffect(() => {
    if (!session.loading && session.isAuthenticated) {
      getSubscription()
        .then(setStatus)
        .catch(() => setError('Failed to load billing status'))
        .finally(() => setLoading(false));
    }
  }, [session.loading, session.isAuthenticated]);

  const handleBuy = async () => {
    setCheckoutLoading(true);
    setError(null);
    try {
      const url = await createCheckoutSession(
        'report_export',
        `${window.location.origin}/settings/billing?paid=1`,
        `${window.location.origin}/settings/billing`,
      );
      window.location.href = url;
    } catch {
      setError('Could not start checkout. Please try again.');
      setCheckoutLoading(false);
    }
  };

  if (session.loading || !session.isAuthenticated || loading) {
    return (
      <div className="flex min-h-[40vh] items-center justify-center">
        <div className="h-10 w-10 animate-spin rounded-full border-b-2" style={{ borderColor: '#2f6b50' }} />
      </div>
    );
  }

  const reportPaid = Boolean(status?.report_paid);

  return (
    <div className="mx-auto max-w-[640px]" style={{ fontFamily: "'Schibsted Grotesk', system-ui, sans-serif", color: '#16201b' }}>
      <h1 className="mb-[6px] text-[24px] font-bold tracking-[-0.02em]">Billing</h1>
      <p className="mb-8 text-[14.5px]" style={{ color: '#6a7a71' }}>
        Free to upload and review your footprint. You only pay when you export a report.
      </p>

      {paid && (
        <div className="mb-5 flex items-center gap-[9px] rounded-lg border p-3 text-[13.5px]" style={{ background: '#e8f0ea', borderColor: '#bcd0c4', color: '#2f6b50' }}>
          <span>✓</span> Payment received — your reports are unlocked.
        </div>
      )}
      {error && (
        <div className="mb-5 rounded-lg border p-3 text-[13.5px]" style={{ background: '#fef2f2', borderColor: '#fecaca', color: '#991b1b' }}>
          {error}
        </div>
      )}

      {/* Report entitlement card */}
      <div className="rounded-[13px] border bg-white p-[26px]" style={{ borderColor: '#e8ece8' }}>
        <div className="flex items-start justify-between gap-4">
          <div>
            <div className="text-[15px] font-semibold">Audit-ready reports</div>
            <div className="mt-1 text-[13.5px]" style={{ color: '#6a7a71' }}>
              {reportPaid
                ? 'Unlocked — export PDF & CSV any time. Re-exports are free for 12 months.'
                : 'Locked. Unlock to export your audit-ready report as PDF & CSV.'}
            </div>
          </div>
          <span className="shrink-0 rounded-full px-[12px] py-[5px] text-[12px] font-semibold"
            style={reportPaid ? { background: '#e8f0ea', color: '#2f6b50' } : { background: '#f1f2f0', color: '#8a978f' }}>
            {reportPaid ? 'Unlocked' : 'Locked'}
          </span>
        </div>

        <div className="mt-6 flex items-center justify-between border-t pt-5" style={{ borderColor: '#f2f4f1' }}>
          <div>
            <span className="text-[26px] font-medium" style={{ fontFamily: "'IBM Plex Mono', monospace" }}>{REPORT_PRICE}</span>
            <span className="ml-2 text-[13px]" style={{ color: '#8a978f' }}>one-time per report</span>
          </div>
          {reportPaid ? (
            <button onClick={() => router.push('/reports')} className="flex h-[44px] items-center rounded-[9px] px-5 text-[14px] font-semibold text-white" style={{ background: '#1d3b2e' }}>
              Go to reports →
            </button>
          ) : (
            <button onClick={handleBuy} disabled={checkoutLoading} className="flex h-[44px] items-center rounded-[9px] px-5 text-[14px] font-semibold text-white disabled:opacity-50" style={{ background: '#1d3b2e' }}>
              {checkoutLoading ? 'Starting…' : `Unlock for ${REPORT_PRICE}`}
            </button>
          )}
        </div>
      </div>

      <div className="mt-4 flex items-center gap-[9px] rounded-[11px] border bg-white p-4 text-[12.5px]" style={{ borderColor: '#e8ece8', color: '#8a978f' }}>
        <span style={{ color: '#2f6b50' }}>🔒</span>
        Payments are processed securely by Stripe. We never store your card details.
      </div>

      <div className="mt-8">
        <Link href="/settings" className="text-[13px]" style={{ color: '#8a978f' }}>← Back to settings</Link>
      </div>
    </div>
  );
}

export default function BillingPage() {
  return (
    <Suspense fallback={
      <div className="flex min-h-[40vh] items-center justify-center">
        <div className="h-10 w-10 animate-spin rounded-full border-b-2" style={{ borderColor: '#2f6b50' }} />
      </div>
    }>
      <BillingContent />
    </Suspense>
  );
}
