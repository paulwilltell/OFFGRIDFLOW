'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useSession, useRequireAuth } from '@/lib/session';
import ErrorBoundary from '@/components/ErrorBoundary';
import { HelpWidget } from './components/HelpWidget';

type Step = { num: number; label: string; href: string; paths: string[] };

const STEPS: Step[] = [
  { num: 1, label: 'Upload', href: '/emissions', paths: ['/emissions'] },
  { num: 2, label: 'Review', href: '/dashboard/carbon', paths: ['/dashboard', '/audit'] },
  { num: 3, label: 'Report', href: '/reports', paths: ['/reports', '/compliance'] },
];

function getActiveStep(pathname: string | null): number {
  if (!pathname) return 1;
  for (const step of STEPS) {
    if (step.paths.some((p) => pathname === p || pathname.startsWith(p + '/'))) {
      return step.num;
    }
  }
  return 0;
}

function StepIndicator({ step, activeStep }: { step: Step; activeStep: number }) {
  const isComplete = step.num < activeStep;
  const isCurrent = step.num === activeStep;

  return (
    <Link href={step.href} className="flex items-center gap-2">
      <span
        className={`flex h-[21px] w-[21px] items-center justify-center rounded-full text-[11px] font-medium ${
          isComplete
            ? 'bg-[#2f6b50] text-white'
            : isCurrent
              ? 'border border-[#2f6b50] bg-[#2f6b50] text-white font-semibold'
              : 'border border-[#d4dbd6] text-[#9aa79f]'
        }`}
        style={{ fontFamily: "'IBM Plex Mono', monospace" }}
      >
        {isComplete ? '✓' : step.num}
      </span>
      <span
        className={`text-[13.5px] ${
          isCurrent ? 'font-semibold text-[#16201b]' : isComplete ? 'text-[#5b6b62]' : 'text-[#9aa79f]'
        }`}
      >
        {step.label}
      </span>
    </Link>
  );
}

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const session = useRequireAuth();
  const pathname = usePathname();
  const { user, logout } = useSession();

  if (!session?.isAuthenticated) {
    return null;
  }

  const activeStep = getActiveStep(pathname);
  const companyInitial = user?.name?.charAt(0)?.toUpperCase() || 'U';

  return (
    <div className="min-h-screen" style={{ background: '#f7f8f6', fontFamily: "'Schibsted Grotesk', system-ui, sans-serif", color: '#16201b' }}>
      {/* Topbar */}
      <header className="flex h-[62px] items-center justify-between border-b border-[#eef1ee] bg-white px-6">
        <Link href="/emissions" className="flex items-center gap-[10px] text-[16px] font-bold tracking-[-0.01em]">
          <span className="flex h-[23px] w-[23px] items-center justify-center rounded-[6px] bg-[#1d3b2e] text-[13px] text-[#5fbf8e]">
            ◇
          </span>
          OffGridFlow
        </Link>

        {/* 3-step stepper */}
        <nav className="flex items-center">
          {STEPS.map((step, i) => (
            <div key={step.num} className="flex items-center">
              {i > 0 && <span className="mx-[13px] h-px w-[30px] bg-[#dde3de]" />}
              <StepIndicator step={step} activeStep={activeStep} />
            </div>
          ))}
        </nav>

        {/* User */}
        <div className="flex items-center gap-3">
          <span className="text-[13.5px] text-[#5b6b62]">{user?.name || user?.email}</span>
          <button
            onClick={() => logout()}
            className="flex h-[30px] w-[30px] items-center justify-center rounded-full bg-[#dceadf] text-[13px] font-semibold text-[#2f6b50]"
            title="Sign out"
          >
            {companyInitial}
          </button>
          <Link href="/settings" className="text-[12px] text-[#9aa79f] hover:text-[#5b6b62]">
            Settings
          </Link>
        </div>
      </header>

      {/* Content */}
      <main className="mx-auto max-w-[1180px] px-6 py-8">
        <ErrorBoundary componentName="App Page" resetKeys={[pathname]}>
          {children}
        </ErrorBoundary>
      </main>

      <HelpWidget />
    </div>
  );
}
