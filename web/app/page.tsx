import Link from 'next/link';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'OffGridFlow | Carbon accounting without the busywork',
  description:
    'Upload your utility data, see your carbon footprint calculated with GHG Protocol methodology for free, and export an audit-ready report for $149.',
  openGraph: {
    title: 'OffGridFlow | Carbon accounting without the busywork',
    description:
      'Upload your data. We calculate Scope 2 and hand you an audit-ready report. Free to review, $149 to export.',
    type: 'website',
    url: 'https://off-grid-flow.com',
  },
};

const FONT = "'Schibsted Grotesk', system-ui, sans-serif";
const MONO = "'IBM Plex Mono', monospace";

const STEPS = [
  { n: 1, title: 'Upload your data', body: 'Drop a CSV of your utility bills. We auto-map your columns — no template wrangling.' },
  { n: 2, title: 'Review your footprint', body: 'We apply verified emission factors and show your Scope 2 footprint, calculated and validated. Free.' },
  { n: 3, title: 'Download the report', body: 'Export an audit-ready PDF & CSV with full methodology and a source trail. $149, one-time.' },
];

const FREE_FEATURES = [
  'Upload utility & energy data (CSV)',
  'Automatic column mapping',
  'Emission factors applied (EPA eGRID, DEFRA)',
  'Full Scope 2 footprint dashboard',
  'Data quality anomaly scan',
];

const REPORT_FEATURES = [
  'Everything in Free',
  'Audit-ready PDF & CSV export',
  'GHG Protocol + CSRD / ESRS E1',
  'Full methodology & source trail',
  'No subscription, no per-seat fees',
];

export default function HomePage() {
  return (
    <div style={{ background: '#f7f8f6', color: '#16201b', fontFamily: FONT }} className="min-h-screen">
      {/* Nav */}
      <header className="flex h-[64px] items-center justify-between border-b border-[#eef1ee] bg-white px-6">
        <div className="flex items-center gap-[10px] text-[16px] font-bold tracking-[-0.01em]">
          <span className="flex h-[24px] w-[24px] items-center justify-center rounded-[6px] bg-[#1d3b2e] text-[13px] text-[#5fbf8e]">◇</span>
          OffGridFlow
        </div>
        <div className="flex items-center gap-5 text-[14px]">
          <Link href="/pricing" className="text-[#5b6b62] hover:text-[#16201b]">Pricing</Link>
          <Link href="/login" className="text-[#5b6b62] hover:text-[#16201b]">Sign in</Link>
          <Link href="/register" className="rounded-[9px] bg-[#1d3b2e] px-4 py-2 font-semibold text-white hover:bg-[#234e3b]">Start free</Link>
        </div>
      </header>

      {/* Hero */}
      <section className="mx-auto max-w-[900px] px-6 pb-20 pt-24 text-center">
        <div className="mx-auto mb-6 inline-flex items-center gap-2 rounded-full border border-[#dbe5de] bg-white px-3 py-1 text-[12.5px]" style={{ color: '#5b6b62' }}>
          <span className="h-[6px] w-[6px] rounded-full bg-[#5fbf8e]" /> Free to upload &amp; review · $149 per report
        </div>
        <h1 className="text-[44px] font-bold leading-[1.1] tracking-[-0.02em] sm:text-[56px]">
          Carbon accounting<br />without the busywork.
        </h1>
        <p className="mx-auto mt-6 max-w-[560px] text-[17px] leading-[1.6]" style={{ color: '#6a7a71' }}>
          Upload your utility data. We calculate your Scope&nbsp;2 footprint with GHG Protocol
          methodology and hand you an audit-ready report. Free to review, $149 to export.
        </p>
        <div className="mt-9 flex flex-col items-center justify-center gap-3 sm:flex-row">
          <Link href="/register" className="flex h-[48px] items-center gap-2 rounded-[10px] bg-[#1d3b2e] px-7 text-[15px] font-semibold text-white hover:bg-[#234e3b]">
            Start free <span className="text-[17px]">→</span>
          </Link>
          <Link href="/login" className="text-[14px] font-medium" style={{ color: '#6a7a71' }}>
            Already have an account? Sign in
          </Link>
        </div>

        {/* 3-step chips */}
        <div className="mx-auto mt-14 flex max-w-[560px] flex-col gap-3 sm:flex-row">
          {STEPS.map((s) => (
            <div key={s.n} className="flex flex-1 items-center gap-[10px] rounded-[11px] border border-[#e8ece8] bg-white px-4 py-3 text-left">
              <span className="flex h-[26px] w-[26px] shrink-0 items-center justify-center rounded-full border border-[#3f6b54] text-[12px]" style={{ fontFamily: MONO, color: '#2f6b50' }}>{s.n}</span>
              <span className="text-[13.5px] font-semibold">{s.title}</span>
            </div>
          ))}
        </div>
      </section>

      {/* How it works */}
      <section className="border-t border-[#eef1ee] bg-white px-6 py-20">
        <div className="mx-auto max-w-[960px]">
          <h2 className="text-center text-[30px] font-bold tracking-[-0.02em]">Three steps. No consultants.</h2>
          <p className="mx-auto mt-3 max-w-[520px] text-center text-[15px]" style={{ color: '#6a7a71' }}>
            You already have the data. We do the calculating, organizing, validating, and reporting.
          </p>
          <div className="mt-12 grid gap-6 md:grid-cols-3">
            {STEPS.map((s) => (
              <div key={s.n} className="rounded-[14px] border border-[#e8ece8] p-6" style={{ background: '#fbfcfb' }}>
                <span className="flex h-[34px] w-[34px] items-center justify-center rounded-[9px] text-[15px] text-white" style={{ background: '#1d3b2e', fontFamily: MONO }}>{s.n}</span>
                <div className="mt-4 text-[16px] font-semibold">{s.title}</div>
                <p className="mt-2 text-[14px] leading-[1.55]" style={{ color: '#6a7a71' }}>{s.body}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Pricing */}
      <section id="pricing" className="border-t border-[#eef1ee] px-6 py-20">
        <div className="mx-auto max-w-[760px]">
          <h2 className="text-center text-[30px] font-bold tracking-[-0.02em]">Simple, honest pricing</h2>
          <p className="mx-auto mt-3 max-w-[480px] text-center text-[15px]" style={{ color: '#6a7a71' }}>
            See your real footprint for free. Pay only when you export a report.
          </p>
          <div className="mt-12 grid gap-6 sm:grid-cols-2">
            {/* Free */}
            <div className="rounded-[14px] border border-[#e8ece8] bg-white p-8">
              <div className="text-[16px] font-semibold">Free</div>
              <div className="mt-3 text-[40px] font-bold">$0</div>
              <p className="mt-2 text-[13.5px]" style={{ color: '#6a7a71' }}>See your footprint before you pay anything.</p>
              <ul className="mt-6 flex flex-col gap-[11px]">
                {FREE_FEATURES.map((f) => (
                  <li key={f} className="flex items-start gap-2 text-[14px]" style={{ color: '#3f4f47' }}><span style={{ color: '#2f6b50' }}>✓</span> {f}</li>
                ))}
              </ul>
              <Link href="/register" className="mt-7 block rounded-[9px] border border-[#cfdcd4] py-[11px] text-center text-[14px] font-semibold" style={{ color: '#2f6b50' }}>Start free</Link>
            </div>
            {/* Report */}
            <div className="relative rounded-[14px] border-[1.5px] border-[#2f6b50] bg-white p-8">
              <div className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-[#2f6b50] px-3 py-[3px] text-[11px] font-semibold text-white">Pay per report</div>
              <div className="text-[16px] font-semibold">Audit-ready report</div>
              <div className="mt-3 flex items-baseline gap-1"><span className="text-[40px] font-bold">$149</span><span className="text-[13px]" style={{ color: '#8a978f' }}>/ report</span></div>
              <p className="mt-2 text-[13.5px]" style={{ color: '#6a7a71' }}>One-time. Re-export free for 12 months.</p>
              <ul className="mt-6 flex flex-col gap-[11px]">
                {REPORT_FEATURES.map((f) => (
                  <li key={f} className="flex items-start gap-2 text-[14px]" style={{ color: '#3f4f47' }}><span style={{ color: '#2f6b50' }}>✓</span> {f}</li>
                ))}
              </ul>
              <Link href="/register" className="mt-7 block rounded-[9px] bg-[#1d3b2e] py-[11px] text-center text-[14px] font-semibold text-white hover:bg-[#234e3b]">Upload data to start</Link>
            </div>
          </div>
          <p className="mt-8 text-center text-[12.5px]" style={{ color: '#8a978f' }}>
            Currently calculates Scope&nbsp;2 (electricity &amp; utilities). Scope&nbsp;1 &amp; 3 coming soon.
          </p>
        </div>
      </section>

      {/* Trust + CTA */}
      <section className="border-t border-[#eef1ee] bg-white px-6 py-20">
        <div className="mx-auto max-w-[620px] text-center">
          <h2 className="text-[28px] font-bold tracking-[-0.02em]">See your carbon footprint in minutes.</h2>
          <p className="mx-auto mt-3 max-w-[460px] text-[15px]" style={{ color: '#6a7a71' }}>
            Built on GHG Protocol methodology with EPA eGRID and DEFRA emission factors. Every number traceable to its source.
          </p>
          <Link href="/register" className="mt-8 inline-flex h-[48px] items-center gap-2 rounded-[10px] bg-[#1d3b2e] px-7 text-[15px] font-semibold text-white hover:bg-[#234e3b]">
            Start free <span className="text-[17px]">→</span>
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-[#eef1ee] bg-white px-6 py-10">
        <div className="mx-auto flex max-w-[960px] flex-wrap items-center justify-between gap-4 text-[13px]" style={{ color: '#8a978f' }}>
          <div className="flex items-center gap-[9px] font-semibold" style={{ color: '#16201b' }}>
            <span className="flex h-[20px] w-[20px] items-center justify-center rounded-[5px] bg-[#1d3b2e] text-[11px] text-[#5fbf8e]">◇</span>
            OffGridFlow
          </div>
          <div className="flex flex-wrap gap-x-5 gap-y-2">
            <Link href="/pricing" className="hover:text-[#16201b]">Pricing</Link>
            <Link href="/trust" className="hover:text-[#16201b]">Trust</Link>
            <Link href="/security" className="hover:text-[#16201b]">Security</Link>
            <Link href="/privacy" className="hover:text-[#16201b]">Privacy</Link>
            <Link href="/terms" className="hover:text-[#16201b]">Terms</Link>
            <a href="mailto:contact@off-grid-flow.com" className="hover:text-[#16201b]">Contact</a>
          </div>
          <div>&copy; {new Date().getFullYear()} OffGridFlow LLC</div>
        </div>
      </footer>
    </div>
  );
}
