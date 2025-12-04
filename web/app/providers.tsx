'use client';

import { SessionProvider } from '@/lib/session';
import { DesignSystemProvider } from './components/DesignSystemProvider';
import { ToastProvider } from './components/Toast';
import { I18nextProvider } from 'react-i18next';
import i18n from '@/lib/i18n';
import { PerformanceMonitoring } from './components/PerformanceUtils';

export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <I18nextProvider i18n={i18n}>
      <DesignSystemProvider>
        <SessionProvider>
          <PerformanceMonitoring />
          <ToastProvider />
          {children}
        </SessionProvider>
      </DesignSystemProvider>
    </I18nextProvider>
  );
}
