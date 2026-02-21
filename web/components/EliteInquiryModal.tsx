'use client';

import { useState } from 'react';

interface EliteInquiryModalProps {
  isOpen: boolean;
  onClose: () => void;
  // Pre-filled from their account
  userEmail?: string;
  userName?: string;
  companyName?: string;
  companySize?: string;
  industry?: string;
  numberOfSites?: number;
  currentFrameworks?: string[];
  currentPlan?: string;
}

const FRAMEWORKS = ['CSRD', 'SEC', 'CBAM', 'IFRS S2', 'GRI', 'CDP', 'CA SB 253', 'TCFD', 'Other'];
const COMPANY_SIZES = ['100–500 employees', '500–1,000 employees', '1,000–5,000 employees', '5,000–10,000 employees', '10,000+ employees'];
const INDUSTRIES = [
  'Manufacturing', 'Energy & Utilities', 'Financial Services', 'Real Estate',
  'Technology', 'Healthcare', 'Retail & Consumer Goods', 'Transportation & Logistics',
  'Agriculture', 'Construction', 'Other',
];
const TIMELINES = ['Immediate (within 30 days)', '1–3 months', '3–6 months', 'Evaluating for next year'];

export default function EliteInquiryModal({
  isOpen,
  onClose,
  userEmail = '',
  userName = '',
  companyName = '',
  companySize = '',
  industry = '',
  numberOfSites = 1,
  currentFrameworks = [],
  currentPlan = '',
}: EliteInquiryModalProps) {
  const [fromEmail, setFromEmail] = useState(userEmail);
  const [contactName, setContactName] = useState(userName);
  const [company, setCompany] = useState(companyName);
  const [size, setSize] = useState(companySize);
  const [selectedIndustry, setSelectedIndustry] = useState(industry);
  const [sites, setSites] = useState(numberOfSites);
  const [frameworks, setFrameworks] = useState<string[]>(currentFrameworks);
  const [regions, setRegions] = useState('');
  const [timeline, setTimeline] = useState('');
  const [notes, setNotes] = useState('');
  const [submitted, setSubmitted] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const toggleFramework = (fw: string) => {
    setFrameworks(prev =>
      prev.includes(fw) ? prev.filter(f => f !== fw) : [...prev, fw]
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);

    const payload = {
      to: 'paul@offgridflow.com',
      from: fromEmail,
      contact_name: contactName,
      company_name: company,
      company_size: size,
      industry: selectedIndustry,
      number_of_sites: sites,
      compliance_frameworks: frameworks,
      regions_of_operation: regions,
      implementation_timeline: timeline,
      current_plan: currentPlan,
      additional_notes: notes,
    };

    try {
      const res = await fetch('/api/elite-inquiry', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!res.ok) throw new Error('Failed to send');
      setSubmitted(true);
    } catch {
      setError('Failed to send your inquiry. Please email paul@offgridflow.com directly.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto" aria-modal="true" role="dialog">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/60 backdrop-blur-sm transition-opacity"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="relative min-h-full flex items-center justify-center p-4">
        <div className="relative w-full max-w-2xl bg-white dark:bg-gray-900 rounded-2xl shadow-2xl">

          {/* Header */}
          <div className="bg-gradient-to-r from-gray-900 to-gray-800 rounded-t-2xl px-8 py-6">
            <div className="flex items-start justify-between">
              <div>
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-xs font-bold tracking-widest text-green-400 uppercase">Carbon Command Elite</span>
                </div>
                <h2 className="text-2xl font-bold text-white">Request Enterprise Pricing</h2>
                <p className="text-gray-400 text-sm mt-1">
                  We'll build a custom proposal tailored to your organization within 24 hours.
                </p>
              </div>
              <button
                onClick={onClose}
                className="text-gray-400 hover:text-white transition-colors mt-1 ml-4"
                aria-label="Close"
              >
                <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>

          {submitted ? (
            <div className="px-8 py-12 text-center">
              <div className="mx-auto w-16 h-16 bg-green-100 dark:bg-green-900/50 rounded-full flex items-center justify-center mb-4">
                <svg className="h-8 w-8 text-green-600 dark:text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </div>
              <h3 className="text-xl font-bold text-gray-900 dark:text-white mb-2">Inquiry Received</h3>
              <p className="text-gray-600 dark:text-gray-400 mb-6">
                Your enterprise pricing request has been sent to our team. Expect a tailored proposal within 24 business hours.
              </p>
              <button
                onClick={onClose}
                className="px-6 py-2 rounded-lg bg-green-600 text-white font-medium hover:bg-green-700 transition-colors"
              >
                Close
              </button>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="px-8 py-6 space-y-6">

              {/* Email routing */}
              <div className="bg-gray-50 dark:bg-gray-800 rounded-xl p-4 space-y-3 text-sm">
                <div className="flex items-center gap-3">
                  <span className="text-gray-500 dark:text-gray-400 w-8 shrink-0">To:</span>
                  <span className="font-medium text-gray-900 dark:text-white">paul@offgridflow.com</span>
                  <span className="ml-auto text-xs text-gray-400">OffGridFlow Customer Success</span>
                </div>
                <div className="flex items-center gap-3">
                  <label htmlFor="fromEmail" className="text-gray-500 dark:text-gray-400 w-8 shrink-0">From:</label>
                  <input
                    id="fromEmail"
                    type="email"
                    required
                    placeholder="your.name@company.com"
                    value={fromEmail}
                    onChange={e => setFromEmail(e.target.value)}
                    className="flex-1 bg-transparent border-b border-gray-300 dark:border-gray-600 focus:border-green-500 outline-none text-gray-900 dark:text-white placeholder-gray-400 py-1"
                  />
                </div>
              </div>

              {/* Contact info */}
              <div>
                <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide mb-3">Contact Information</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs text-gray-500 dark:text-gray-400 mb-1">Full Name *</label>
                    <input
                      type="text"
                      required
                      value={contactName}
                      onChange={e => setContactName(e.target.value)}
                      className="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-sm text-gray-900 dark:text-white focus:ring-2 focus:ring-green-500 focus:border-transparent outline-none"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-gray-500 dark:text-gray-400 mb-1">Company Name *</label>
                    <input
                      type="text"
                      required
                      value={company}
                      onChange={e => setCompany(e.target.value)}
                      className="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-sm text-gray-900 dark:text-white focus:ring-2 focus:ring-green-500 focus:border-transparent outline-none"
                    />
                  </div>
                </div>
              </div>

              {/* Company details */}
              <div>
                <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide mb-3">Organization Details</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs text-gray-500 dark:text-gray-400 mb-1">Company Size</label>
                    <select
                      value={size}
                      onChange={e => setSize(e.target.value)}
                      className="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-sm text-gray-900 dark:text-white focus:ring-2 focus:ring-green-500 outline-none"
                    >
                      <option value="">Select…</option>
                      {COMPANY_SIZES.map(s => <option key={s} value={s}>{s}</option>)}
                    </select>
                  </div>
                  <div>
                    <label className="block text-xs text-gray-500 dark:text-gray-400 mb-1">Industry</label>
                    <select
                      value={selectedIndustry}
                      onChange={e => setSelectedIndustry(e.target.value)}
                      className="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-sm text-gray-900 dark:text-white focus:ring-2 focus:ring-green-500 outline-none"
                    >
                      <option value="">Select…</option>
                      {INDUSTRIES.map(i => <option key={i} value={i}>{i}</option>)}
                    </select>
                  </div>
                  <div>
                    <label className="block text-xs text-gray-500 dark:text-gray-400 mb-1">Number of Facilities / Sites</label>
                    <input
                      type="number"
                      min={1}
                      value={sites}
                      onChange={e => setSites(Number(e.target.value))}
                      className="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-sm text-gray-900 dark:text-white focus:ring-2 focus:ring-green-500 outline-none"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-gray-500 dark:text-gray-400 mb-1">Regions of Operation</label>
                    <input
                      type="text"
                      placeholder="e.g. US, EU, UK, APAC"
                      value={regions}
                      onChange={e => setRegions(e.target.value)}
                      className="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-sm text-gray-900 dark:text-white focus:ring-2 focus:ring-green-500 outline-none"
                    />
                  </div>
                </div>
              </div>

              {/* Compliance frameworks */}
              <div>
                <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide mb-3">
                  Compliance Frameworks Required
                </h3>
                <div className="flex flex-wrap gap-2">
                  {FRAMEWORKS.map(fw => (
                    <button
                      key={fw}
                      type="button"
                      onClick={() => toggleFramework(fw)}
                      className={`px-3 py-1.5 rounded-full text-xs font-medium border transition-colors ${
                        frameworks.includes(fw)
                          ? 'bg-green-600 border-green-600 text-white'
                          : 'border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:border-green-400'
                      }`}
                    >
                      {fw}
                    </button>
                  ))}
                </div>
              </div>

              {/* Timeline */}
              <div>
                <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide mb-3">Implementation Timeline</h3>
                <div className="grid grid-cols-2 gap-2">
                  {TIMELINES.map(t => (
                    <button
                      key={t}
                      type="button"
                      onClick={() => setTimeline(t)}
                      className={`px-3 py-2 rounded-lg text-xs font-medium border text-left transition-colors ${
                        timeline === t
                          ? 'bg-green-600 border-green-600 text-white'
                          : 'border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:border-green-400'
                      }`}
                    >
                      {t}
                    </button>
                  ))}
                </div>
              </div>

              {/* Notes */}
              <div>
                <label className="block text-xs text-gray-500 dark:text-gray-400 mb-1 font-semibold uppercase tracking-wide">
                  Additional Context (optional)
                </label>
                <textarea
                  rows={3}
                  placeholder="Current pain points, specific requirements, data sources you use, or anything else that helps us build the right proposal…"
                  value={notes}
                  onChange={e => setNotes(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-2 text-sm text-gray-900 dark:text-white focus:ring-2 focus:ring-green-500 outline-none resize-none"
                />
              </div>

              {error && (
                <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
              )}

              <div className="flex items-center justify-between pt-2 pb-2">
                <p className="text-xs text-gray-400">
                  We respond within 24 business hours with a tailored proposal.
                </p>
                <div className="flex gap-3">
                  <button
                    type="button"
                    onClick={onClose}
                    className="px-5 py-2.5 rounded-lg text-sm font-medium text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={submitting}
                    className="px-6 py-2.5 rounded-lg text-sm font-medium bg-green-600 text-white hover:bg-green-700 disabled:opacity-50 transition-colors"
                  >
                    {submitting ? 'Sending…' : 'Send Inquiry'}
                  </button>
                </div>
              </div>

            </form>
          )}
        </div>
      </div>
    </div>
  );
}
