'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useSession, useRequireAuth } from '@/lib/session';
import ErrorBoundary from '@/components/ErrorBoundary';
import { HelpWidget } from './components/HelpWidget';

const navItems = [
  {
    label: 'Dashboard',
    href: '/dashboard/carbon',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
      </svg>
    ),
  },
  {
    label: 'Emissions',
    href: '/emissions',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
      </svg>
    ),
  },
  {
    label: 'Compliance',
    href: '/compliance/csrd',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
    ),
    children: [
      { label: 'CSRD / ESRS', href: '/compliance/csrd' },
      { label: 'SEC Climate', href: '/compliance/sec' },
      { label: 'California SB 253', href: '/compliance/california' },
      { label: 'CBAM', href: '/compliance/cbam' },
      { label: 'Scope 3 Categories', href: '/compliance/scope3' },
      { label: 'Risk Abatement', href: '/compliance/abatement/sb253' },
    ],
  },
  {
    label: 'Audit',
    href: '/audit/ledger',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
      </svg>
    ),
    children: [
      { label: 'Calculation Ledger', href: '/audit/ledger' },
      { label: 'Approvals', href: '/audit/approvals' },
      { label: 'Factor Snapshots', href: '/audit/factor-snapshots' },
      { label: 'Data Quality', href: '/audit/data-quality' },
      { label: 'Alerts', href: '/audit/alerts' },
      { label: 'Getting Started', href: '/onboarding' },
    ],
  },
  {
    label: 'Settings',
    href: '/settings',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
      </svg>
    ),
    children: [
      { label: 'Billing', href: '/settings/billing' },
      { label: 'Data Sources', href: '/settings/data-sources' },
      { label: 'Data Governance', href: '/settings/data-governance' },
      { label: 'Organization', href: '/settings/organization' },
      { label: 'Security', href: '/settings/security' },
      { label: 'Users', href: '/settings/users' },
      { label: 'Emission Factors', href: '/settings/factors' },
    ],
  },
];

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const session = useRequireAuth();
  const pathname = usePathname();
  const { user, logout } = useSession();

  if (!session?.isAuthenticated) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-dark-900">
        <div className="text-gray-400">Loading...</div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen bg-dark-900">
      {/* Sidebar */}
      <aside className="fixed left-0 top-0 z-40 flex h-screen w-60 flex-col border-r border-gray-800/50 bg-dark-800">
        {/* Logo */}
        <div className="flex h-16 items-center border-b border-gray-800/50 px-5">
          <Link href="/dashboard/carbon" className="text-lg font-bold text-white">
            OffGridFlow
          </Link>
        </div>

        {/* Navigation */}
        <nav className="flex-1 overflow-y-auto px-3 py-4">
          <ul className="space-y-1">
            {navItems.map((item) => {
              const isActive = pathname === item.href || pathname?.startsWith(item.href + '/');
              const isParentActive = item.children?.some(
                (child) => pathname === child.href
              );

              return (
                <li key={item.href}>
                  <Link
                    href={item.href}
                    className={`flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition ${
                      isActive || isParentActive
                        ? 'bg-primary-600/10 text-primary-400'
                        : 'text-gray-400 hover:bg-gray-800 hover:text-white'
                    }`}
                  >
                    {item.icon}
                    {item.label}
                  </Link>
                  {/* Sub-nav */}
                  {item.children && (isActive || isParentActive) && (
                    <ul className="ml-8 mt-1 space-y-0.5">
                      {item.children.map((child) => (
                        <li key={child.href}>
                          <Link
                            href={child.href}
                            className={`block rounded-md px-3 py-1.5 text-xs transition ${
                              pathname === child.href
                                ? 'text-primary-400'
                                : 'text-gray-500 hover:text-gray-300'
                            }`}
                          >
                            {child.label}
                          </Link>
                        </li>
                      ))}
                    </ul>
                  )}
                </li>
              );
            })}
          </ul>
        </nav>

        {/* User footer */}
        <div className="border-t border-gray-800/50 p-4">
          <div className="mb-2 truncate text-xs text-gray-500">{user?.email}</div>
          <button
            onClick={() => logout()}
            className="w-full rounded-lg border border-gray-700 px-3 py-1.5 text-xs text-gray-400 transition hover:border-gray-600 hover:text-white"
          >
            Sign Out
          </button>
        </div>
      </aside>

      {/* Main content — wrapped in an ErrorBoundary so that a crash in any
          child page displays a recoverable fallback UI instead of white-screening
          the entire authenticated shell or kicking the user to login. */}
      <main className="ml-60 flex-1 p-6">
        <ErrorBoundary componentName="App Page" resetKeys={[pathname]}>
          {children}
        </ErrorBoundary>
      </main>

      {/* Floating help widget — provides every authenticated page with a
          self-service escalation path to reduce chargeback risk. */}
      <HelpWidget />
    </div>
  );
}
