'use client';

import { useState, type FormEvent } from 'react';
import Link from 'next/link';

interface LeadCaptureFormProps {
  /** Where the form is embedded (for tracking). */
  source: string;
  /** Pre-fill the framework interest if known. */
  framework?: string;
  /** Compact mode for sidebar/modal placement. */
  compact?: boolean;
  /** Called after successful submission. */
  onSuccess?: () => void;
}

/**
 * Lead capture form that posts to /api/elite-inquiry.
 * Used on demo, pricing, and framework pages to replace raw mailto links
 * with a real form that routes to the support inbox and can be wired into
 * a CRM pipeline.
 *
 * Routes by company size:
 *  - Enterprise ($1B+ revenue or 1000+ employees) → flags for sales follow-up
 *  - Mid-market → standard follow-up sequence
 *  - Self-serve → directs to /register
 */
export function LeadCaptureForm({
  source,
  framework,
  compact = false,
  onSuccess,
}: LeadCaptureFormProps) {
  const [contactName, setContactName] = useState('');
  const [email, setEmail] = useState('');
  const [company, setCompany] = useState('');
  const [companySize, setCompanySize] = useState('');
  const [interest, setInterest] = useState(framework || '');
  const [timeline, setTimeline] = useState('');
  const [message, setMessage] = useState('');

  const [submitting, setSubmitting] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);

    // Track the lead event via gtag if available
    try {
      const win = window as unknown as { gtag?: (...args: unknown[]) => void };
      if (typeof win.gtag === 'function') {
        win.gtag('event', 'generate_lead', {
          event_category: 'lead_capture',
          event_label: source,
          value: companySize === 'enterprise' ? 15000 : companySize === 'midmarket' ? 10800 : 6500,
          currency: 'USD',
        });
      }
    } catch { /* tracking is best-effort */ }

    try {
      const res = await fetch('/api/elite-inquiry', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          from: email,
          contact_name: contactName,
          company_name: company,
          company_size: companySize,
          compliance_frameworks: interest ? [interest] : [],
          implementation_timeline: timeline,
          additional_notes: message,
          lead_source: source,
          lead_routing: companySize === 'enterprise' ? 'sales' : 'standard',
        }),
      });

      if (!res.ok) throw new Error('Failed to submit');
      setSubmitted(true);
      onSuccess?.();
    } catch {
      setError('Failed to send. Please email contact@off-grid-flow.com directly.');
    } finally {
      setSubmitting(false);
    }
  };

  if (submitted) {
    return (
      <div className="rounded-xl border border-green-800/30 bg-green-900/10 p-6 text-center">
        <div className="text-lg font-semibold text-green-400">Received</div>
        <p className="mt-2 text-sm text-gray-400">
          We&apos;ll respond within one business day. In the meantime, you can{' '}
          <Link href="/register?plan=starter" className="text-primary-400 hover:underline">
            start a free trial
          </Link>{' '}
          to explore the platform immediately.
        </p>
      </div>
    );
  }

  const inputClass =
    'w-full rounded-lg border border-gray-700 bg-gray-900/60 px-3 py-2 text-sm text-white placeholder-gray-500 focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500';
  const labelClass = 'block text-xs font-medium text-gray-400 mb-1';

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className={compact ? 'space-y-3' : 'grid gap-4 sm:grid-cols-2'}>
        <div>
          <label htmlFor={`lead-name-${source}`} className={labelClass}>
            Name *
          </label>
          <input
            id={`lead-name-${source}`}
            type="text"
            required
            value={contactName}
            onChange={(e) => setContactName(e.target.value)}
            placeholder="Jane Smith"
            className={inputClass}
          />
        </div>
        <div>
          <label htmlFor={`lead-email-${source}`} className={labelClass}>
            Work email *
          </label>
          <input
            id={`lead-email-${source}`}
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="jane@company.com"
            className={inputClass}
          />
        </div>
        <div>
          <label htmlFor={`lead-company-${source}`} className={labelClass}>
            Company *
          </label>
          <input
            id={`lead-company-${source}`}
            type="text"
            required
            value={company}
            onChange={(e) => setCompany(e.target.value)}
            placeholder="Acme Corp"
            className={inputClass}
          />
        </div>
        <div>
          <label htmlFor={`lead-size-${source}`} className={labelClass}>
            Company size
          </label>
          <select
            id={`lead-size-${source}`}
            value={companySize}
            onChange={(e) => setCompanySize(e.target.value)}
            className={inputClass}
          >
            <option value="">Select...</option>
            <option value="startup">Under $50M revenue</option>
            <option value="midmarket">$50M - $1B revenue</option>
            <option value="enterprise">$1B+ revenue</option>
          </select>
        </div>
        <div>
          <label htmlFor={`lead-framework-${source}`} className={labelClass}>
            Primary framework
          </label>
          <select
            id={`lead-framework-${source}`}
            value={interest}
            onChange={(e) => setInterest(e.target.value)}
            className={inputClass}
          >
            <option value="">Select...</option>
            <option value="sb253">California SB 253</option>
            <option value="csrd">CSRD / ESRS E1</option>
            <option value="sec">SEC Climate Disclosure</option>
            <option value="ifrs_s2">IFRS S2</option>
            <option value="cbam">EU CBAM</option>
            <option value="multiple">Multiple frameworks</option>
          </select>
        </div>
        <div>
          <label htmlFor={`lead-timeline-${source}`} className={labelClass}>
            Timeline
          </label>
          <select
            id={`lead-timeline-${source}`}
            value={timeline}
            onChange={(e) => setTimeline(e.target.value)}
            className={inputClass}
          >
            <option value="">Select...</option>
            <option value="immediate">This month</option>
            <option value="quarter">This quarter</option>
            <option value="6months">Next 6 months</option>
            <option value="evaluating">Just evaluating</option>
          </select>
        </div>
      </div>

      {!compact && (
        <div>
          <label htmlFor={`lead-message-${source}`} className={labelClass}>
            Additional context
          </label>
          <textarea
            id={`lead-message-${source}`}
            rows={3}
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            placeholder="Tell us about your reporting needs, data sources, or questions."
            className={inputClass}
          />
        </div>
      )}

      {error && (
        <div className="rounded-lg border border-red-700 bg-red-900/20 p-3 text-xs text-red-300">
          {error}
        </div>
      )}

      <button
        type="submit"
        disabled={submitting}
        className="w-full rounded-lg bg-primary-600 px-5 py-2.5 text-sm font-semibold text-white transition hover:bg-primary-500 disabled:opacity-50"
      >
        {submitting ? 'Sending...' : 'Talk to us'}
      </button>

      <p className="text-center text-[10px] text-gray-600">
        Or{' '}
        <Link href="/register?plan=starter" className="text-primary-400 hover:underline">
          start a free trial
        </Link>{' '}
        to explore the platform now. No credit card required for trial.
      </p>
    </form>
  );
}

export default LeadCaptureForm;
