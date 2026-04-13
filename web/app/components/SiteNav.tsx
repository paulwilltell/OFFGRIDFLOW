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
        <div className="hidden items-center gap-8 text-sm text-gray-400 md:flex">
          <a href="/#how-it-works" className="transition hover:text-white">
            How It Works
          </a>
          <a href="/#capabilities" className="transition hover:text-white">
            Capabilities
          </a>
          <a href="/#pricing" className="transition hover:text-white">
            Pricing
          </a>
          <Link href="/about" className="transition hover:text-white">
            About
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
                href="/demo"
                className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-primary-500"
              >
                Start Demo
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
}
