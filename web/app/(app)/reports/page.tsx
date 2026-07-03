'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api';
import { useRequireAuth } from '@/lib/session';
import { createCheckoutSession, getSubscription } from '@/lib/billing';
import type { Scope2Summary } from '@/lib/types';

const REPORT_PRICE = '$149';

const FRAMEWORKS = [
  { id: 'ghg', name: 'GHG Protocol inventory', tag: '', desc: 'Scope 1/2/3 corporate standard — the universal baseline', included: true },
  { id: 'csrd', name: 'CSRD / ESRS E1', tag: 'EU', desc: 'European sustainability reporting disclosures', included: true },
  { id: 'sec', name: 'SEC Climate', tag: 'US', desc: 'US public-company climate disclosure', included: false },
  { id: 'cdp', name: 'CDP Climate', tag: 'Global', desc: 'Disclosure questionnaire response pack', included: false },
];

const WHATS_INSIDE = [
  'Executive summary with your total footprint and year-over-year context',
  'Full Scope 1 / 2 / 3 breakdown by source, facility, and activity',
  'Emission factors cited per line (EPA eGRID, DEFRA) — every number traceable',
  'Methodology appendix an auditor can follow end to end',
  'Formatted to your chosen framework — no manual mapping',
];

export default function ReportsPage() {
  const session = useRequireAuth();
  const [summary, setSummary] = useState<Scope2Summary | null>(null);
  const [selected, setSelected] = useState<string[]>(['ghg', 'csrd']);
  const [checkoutLoading, setCheckoutLoading] = useState(false);
  const [paid, setPaid] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.get<Scope2Summary>('/api/emissions/scope2/summary')
      .then(setSummary)
      .catch(() => setSummary(null));
    getSubscription()
      .then((s) => setPaid(Boolean(s.report_paid)))
      .catch(() => setPaid(false));
  }, []);

  const realTotal = summary?.totalEmissionsTonsCO2e;
  const hasData = typeof realTotal === 'number' && realTotal > 0;
  const year = new Date().getFullYear();

  const toggle = (id: string) => {
    setSelected((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));
  };

  const handleCheckout = async () => {
    setError(null);
    setCheckoutLoading(true);
    try {
      const url = await createCheckoutSession(
        'report_export',
        `${window.location.origin}/reports?paid=1`,
        `${window.location.origin}/reports`,
      );
      window.location.href = url;
    } catch {
      setError('Could not start checkout. Please try again.');
      setCheckoutLoading(false);
    }
  };

  const handleDownload = (format: 'pdf' | 'xbrl') => {
    // Export endpoint is gated server-side; only reachable once paid.
    window.open(`/api/compliance/export?format=${format}&year=${year}`, '_blank');
  };

  if (!session?.isAuthenticated) return null;

  return (
    <div className="flex flex-col gap-8 lg:flex-row">
      {/* LEFT — the sell */}
      <div className="flex-1">
        <h1 className="mb-[7px] text-[24px] font-bold tracking-[-0.02em]">Your audit-ready report</h1>
        <p className="mb-6 text-[14.5px]" style={{ color: '#6a7a71' }}>
          {hasData
            ? <>We&apos;ve calculated your footprint from your uploaded data. Choose your framework — we format your <strong>{realTotal!.toFixed(2)} tCO₂e</strong> inventory to match, with full methodology and a source trail included.</>
            : 'Upload your data first, then pick a framework and we format your inventory to match — full methodology and source trail included.'}
        </p>

        {/* framework picker */}
        <div className="mb-7 flex flex-col gap-[13px]">
          {FRAMEWORKS.map((f) => {
            const on = selected.includes(f.id);
            return (
              <button
                key={f.id}
                onClick={() => toggle(f.id)}
                className="flex items-center justify-between rounded-[12px] border bg-white p-[18px_20px] text-left transition"
                style={{ borderColor: on ? '#2f6b50' : '#e8ece8', borderWidth: on ? 1.5 : 1 }}
              >
                <div>
                  <div className="mb-[3px] text-[15px] font-semibold">
                    {f.name}{' '}
                    {f.tag && <span className="text-[11px]" style={{ fontFamily: "'IBM Plex Mono', monospace", color: '#8a978f' }}>{f.tag}</span>}
                  </div>
                  <div className="text-[13px]" style={{ color: '#8a978f' }}>{f.desc}</div>
                </div>
                <span
                  className="flex h-[22px] w-[22px] items-center justify-center rounded-[6px] text-[13px] text-white"
                  style={{ background: on ? '#2f6b50' : 'transparent', border: on ? 'none' : '1.5px solid #d4dbd6' }}
                >
                  {on ? '✓' : ''}
                </span>
              </button>
            );
          })}
        </div>

        {/* what's inside — the pitch */}
        <div className="rounded-[13px] border bg-white p-[22px]" style={{ borderColor: '#e8ece8' }}>
          <div className="mb-[14px] text-[13px] font-semibold" style={{ color: '#3f4f47' }}>WHAT&apos;S INSIDE</div>
          <div className="flex flex-col gap-[11px]">
            {WHATS_INSIDE.map((line, i) => (
              <div key={i} className="flex items-start gap-[10px] text-[14px]" style={{ color: '#3f4f47' }}>
                <span style={{ color: '#2f6b50' }}>✓</span> {line}
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* RIGHT — locked preview + checkout */}
      <div className="w-full lg:w-[380px]">
        {/* locked report preview */}
        <div className="relative overflow-hidden rounded-[13px] border bg-white" style={{ borderColor: '#e8ece8' }}>
          {/* blurred sample pages */}
          <div className="p-6" style={{ filter: 'blur(3px)', userSelect: 'none' }} aria-hidden>
            <div className="mb-2 h-[7px] w-[55%] rounded" style={{ background: '#234e3b' }} />
            <div className="mb-1 h-[4px] w-full rounded" style={{ background: '#e4e9e5' }} />
            <div className="mb-4 h-[4px] w-[80%] rounded" style={{ background: '#e4e9e5' }} />
            <div className="mb-4 flex items-end gap-1" style={{ height: 60 }}>
              <span className="w-4 rounded-sm" style={{ height: '55%', background: '#4f8f6e' }} />
              <span className="w-4 rounded-sm" style={{ height: '100%', background: '#4f8f6e' }} />
              <span className="w-4 rounded-sm" style={{ height: '42%', background: '#88bfa1' }} />
              <span className="w-4 rounded-sm" style={{ height: '73%', background: '#6aa687' }} />
              <span className="w-4 rounded-sm" style={{ height: '30%', background: '#9bccaf' }} />
            </div>
            {[92, 100, 78, 100, 64, 88].map((w, i) => (
              <div key={i} className="mb-[6px] h-[4px] rounded" style={{ width: `${w}%`, background: '#e4e9e5' }} />
            ))}
          </div>

          {/* overlay — lock when unpaid, unlocked confirmation when paid */}
          <div className="absolute inset-0 flex flex-col items-center justify-center text-center" style={{ background: paid ? 'rgba(232,240,234,0.82)' : 'rgba(247,248,246,0.72)' }}>
            <div className="mb-3 flex h-[46px] w-[46px] items-center justify-center rounded-full text-[20px]" style={{ background: '#1d3b2e', color: '#5fbf8e' }}>{paid ? '✓' : '🔒'}</div>
            <div className="text-[15px] font-semibold" style={{ color: '#16201b' }}>{paid ? 'Report unlocked' : 'Report locked'}</div>
            <div className="mt-1 max-w-[240px] text-[12.5px]" style={{ color: '#6a7a71' }}>
              {paid ? 'Download your audit-ready report below.' : 'This is a sample layout. Unlock to generate the real report from your data.'}
            </div>
          </div>
        </div>

        {/* checkout / download */}
        <div className="mt-4 rounded-[13px] border bg-white p-[22px]" style={{ borderColor: '#e8ece8' }}>
          {!paid && (
            <>
              <div className="mb-1 flex items-baseline justify-between">
                <span className="text-[14px] font-semibold">{selected.length} report{selected.length === 1 ? '' : 's'}</span>
                <span className="text-[24px] font-medium" style={{ fontFamily: "'IBM Plex Mono', monospace" }}>{REPORT_PRICE}</span>
              </div>
              <div className="mb-[18px] text-[12.5px]" style={{ color: '#8a978f' }}>One-time · re-export free for 12 months</div>
            </>
          )}

          {error && <div className="mb-3 text-[13px]" style={{ color: '#991b1b' }}>{error}</div>}

          {paid ? (
            <div className="mb-[14px] text-[13.5px] font-semibold" style={{ color: '#2f6b50' }}>Your report is ready to download.</div>
          ) : (
            <button
              onClick={handleCheckout}
              disabled={checkoutLoading || selected.length === 0 || !hasData}
              className="mb-[14px] flex h-[48px] w-full items-center justify-center gap-2 rounded-[10px] text-[15px] font-semibold text-white disabled:opacity-50"
              style={{ background: '#1d3b2e' }}
            >
              {checkoutLoading ? 'Starting checkout…' : hasData ? `Pay ${REPORT_PRICE} & generate` : 'Upload data to unlock'}
            </button>
          )}

          <div className="flex gap-[10px]">
            <button
              onClick={() => paid && handleDownload('pdf')}
              disabled={!paid}
              className="flex h-[44px] flex-1 items-center justify-center gap-[7px] rounded-[9px] border text-[13.5px] font-semibold disabled:opacity-50"
              style={{ borderColor: '#cfdcd4', color: '#3f4f47' }}
            >📄 PDF</button>
            <button
              onClick={() => paid && handleDownload('xbrl')}
              disabled={!paid}
              className="flex h-[44px] flex-1 items-center justify-center gap-[7px] rounded-[9px] border text-[13.5px] font-semibold disabled:opacity-50"
              style={{ borderColor: '#cfdcd4', color: '#3f4f47' }}
            >⊞ XBRL</button>
          </div>

          <div className="mt-4 flex items-center gap-[9px] border-t pt-4 text-[12.5px]" style={{ borderColor: '#f2f4f1', color: '#8a978f' }}>
            <span style={{ color: '#2f6b50' }}>🔒</span> Audit trail &amp; methodology embedded in every export
          </div>
        </div>
      </div>
    </div>
  );
}
