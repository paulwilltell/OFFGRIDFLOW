'use client';

import Link from 'next/link';

export default function DemoPage() {
  return (
    <div className="max-w-5xl mx-auto py-12 px-4">
      <p className="text-sm uppercase tracking-wide text-emerald-500">Demo</p>
      <h1 className="text-3xl font-semibold mt-1">Investor Demo Mode</h1>
      <p className="mt-2 text-gray-500 dark:text-gray-400">
        This interactive demo is coming soon. Until then, explore emissions and compliance data in the main dashboard.
      </p>

      <div className="mt-6 rounded-lg border border-dashed border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/50 p-6">
        <h2 className="text-lg font-medium">What you'll get</h2>
        <ul className="mt-3 space-y-2 text-sm text-gray-600 dark:text-gray-300">
          <li>• Pre-loaded demo data for CSRD, SEC, CBAM, and California readiness.</li>
          <li>• Benchmarking against peer organizations.</li>
          <li>• Guided walkthrough of emissions trends and disclosures.</li>
        </ul>
        <p className="mt-4 text-sm text-gray-500">
          Need a live walkthrough today?{' '}
          <Link href="/" className="text-emerald-600 dark:text-emerald-400 underline">
            Contact your account team
          </Link>{' '}
          and we'll spin up a demo tenant for you.
        </p>
      </div>
    </div>
  );
}
