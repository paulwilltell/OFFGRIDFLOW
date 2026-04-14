'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

/**
 * Minimal, honest cookie consent banner.
 *
 * Design principles:
 *   1. Respects Do Not Track — if DNT is set, we assume non-consent.
 *   2. Denies by default until user clicks Accept. Google Ads gtag only fires
 *      after consent is given.
 *   3. Choice persists in localStorage, not a cookie — no consent-for-consent
 *      paradox.
 *   4. Reject and Accept are equally prominent; no dark patterns.
 */

const STORAGE_KEY = 'offgridflow_cookie_consent_v1';

type ConsentValue = 'accepted' | 'rejected';

function readConsent(): ConsentValue | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (raw === 'accepted' || raw === 'rejected') return raw;
    return null;
  } catch {
    return null;
  }
}

function writeConsent(value: ConsentValue): void {
  if (typeof window === 'undefined') return;
  try {
    window.localStorage.setItem(STORAGE_KEY, value);
  } catch {
    // Ignore — best-effort. If storage is blocked, the banner will reappear.
  }
}

function initGtagConsent(value: ConsentValue): void {
  if (typeof window === 'undefined') return;
  const win = window as unknown as {
    gtag?: (...args: unknown[]) => void;
    dataLayer?: unknown[];
  };
  const gtag = win.gtag;
  if (typeof gtag !== 'function') return;
  try {
    gtag('consent', 'update', {
      ad_storage: value === 'accepted' ? 'granted' : 'denied',
      ad_user_data: value === 'accepted' ? 'granted' : 'denied',
      ad_personalization: value === 'accepted' ? 'granted' : 'denied',
      analytics_storage: value === 'accepted' ? 'granted' : 'denied',
    });
  } catch {
    // Ignore — gtag wiring is enhancement, not required.
  }
}

export function CookieConsent() {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const existing = readConsent();
    if (existing) {
      // Apply prior choice to gtag and stay hidden.
      initGtagConsent(existing);
      return;
    }

    // Respect Do Not Track — treat as rejection without pestering.
    const dnt =
      (typeof navigator !== 'undefined' &&
        (navigator.doNotTrack === '1' ||
          (navigator as unknown as { msDoNotTrack?: string }).msDoNotTrack === '1')) ||
      (typeof window !== 'undefined' &&
        (window as unknown as { doNotTrack?: string }).doNotTrack === '1');
    if (dnt) {
      writeConsent('rejected');
      initGtagConsent('rejected');
      return;
    }

    setVisible(true);
  }, []);

  const handleAccept = () => {
    writeConsent('accepted');
    initGtagConsent('accepted');
    setVisible(false);
  };

  const handleReject = () => {
    writeConsent('rejected');
    initGtagConsent('rejected');
    setVisible(false);
  };

  if (!visible) return null;

  return (
    <div
      role="dialog"
      aria-labelledby="cookie-consent-title"
      aria-describedby="cookie-consent-description"
      className="fixed bottom-4 left-4 right-4 z-50 mx-auto max-w-xl rounded-xl border border-gray-700 bg-gray-900/95 p-4 shadow-2xl backdrop-blur-md sm:left-auto sm:right-4"
    >
      <div className="flex flex-col gap-3">
        <div>
          <p id="cookie-consent-title" className="text-sm font-semibold text-white">
            Cookies on this site
          </p>
          <p id="cookie-consent-description" className="mt-1 text-xs text-gray-400 leading-relaxed">
            We use strictly necessary cookies to run the site. With your consent, we also use Google
            analytics and conversion measurement cookies on marketing pages. Our in-app authenticated
            Platform does not use tracking cookies beyond session management.
            {' '}
            <Link href="/privacy#do-not-sell" className="text-primary-400 hover:underline">
              Privacy details
            </Link>
            .
          </p>
        </div>
        <div className="flex flex-col gap-2 sm:flex-row sm:justify-end">
          <button
            onClick={handleReject}
            className="rounded-lg border border-gray-700 px-4 py-2 text-xs font-medium text-gray-300 hover:border-gray-500 hover:text-white"
          >
            Reject non-essential
          </button>
          <button
            onClick={handleAccept}
            className="rounded-lg bg-primary-600 px-4 py-2 text-xs font-medium text-white hover:bg-primary-500"
          >
            Accept all
          </button>
        </div>
      </div>
    </div>
  );
}

export default CookieConsent;
