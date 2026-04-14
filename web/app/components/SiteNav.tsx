'use client';

import Link from 'next/link';
import { useSession } from '@/lib/session';

export function SiteNav() {
  const { user, logout } = useSession();

  return (
    <nav className="sticky top-0 z-50 border-b border-gray-800/50 bg-[#0b1526]/80 backdrop-blur-md">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <Link href="/" className="flex items-center gap-2">
          <span className="text-xl font-bold text-white">OffGridFlow</span>
          <span className="hidden text-xs text-gray-500 sm:inline">
            Carbon Accounting
          </span>
        </Link>
        <div className="hidden items-center gap-6 text-sm text-gray-400 md:flex">
          <Link href="/demo" className="transition hover:text-white">
            How It Works
          </Link>
          <Link href="/pricing" className="transition hover:text-white">
            Pricing
          </Link>
          <Link href="/methodology" className="transition hover:text-white">
            Methodology
          </Link>
          <Link href="/trust" className="transition hover:text-white">
            Trust
          </Link>
        </div>
        <div className="flex items-center gap-3">
          {user ? (
            <>
              <Link
                href="/dashboard/carbon"
                className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-primary-500"
              >
                Dashboard
              </Link>
              <button
                onClick={() => logout()}
                className="text-sm text-gray-400 transition hover:text-white"
              >
                Sign Out
              </button>
            </>
          ) : (
            <>
              <Link
                href="/login"
                className="text-sm text-gray-400 transition hover:text-white"
              >
                Sign In
              </Link>
              <Link
                href="/register"
                className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-primary-500"
              >
                Start Free Trial
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
}
