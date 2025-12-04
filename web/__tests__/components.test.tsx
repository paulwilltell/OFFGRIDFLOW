/**
 * @jest-environment jsdom
 */
import React from 'react';
import { render, screen } from '@testing-library/react';

// Mock next/navigation
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
    replace: jest.fn(),
    prefetch: jest.fn(),
    back: jest.fn(),
  }),
  useSearchParams: () => new URLSearchParams(),
  usePathname: () => '/',
}));

// Mock session
jest.mock('@/lib/session', () => ({
  SessionProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useSession: () => ({
    user: { id: '1', email: 'test@example.com', name: 'Test', role: 'admin' },
    currentTenantId: 'tenant-1',
    tenants: [],
    accessToken: 'token',
    refreshToken: null,
    loading: false,
    isAuthenticated: true,
    login: jest.fn(),
    logout: jest.fn(),
    switchTenant: jest.fn(),
    refreshSession: jest.fn(),
  }),
  useRequireAuth: () => ({
    user: { id: '1', email: 'test@example.com', name: 'Test', role: 'admin' },
    currentTenantId: 'tenant-1',
    tenants: [],
    accessToken: 'token',
    refreshToken: null,
    loading: false,
    isAuthenticated: true,
    login: jest.fn(),
    logout: jest.fn(),
    switchTenant: jest.fn(),
    refreshSession: jest.fn(),
  }),
}));

// Mock API
jest.mock('@/lib/api', () => ({
  api: {
    get: jest.fn().mockResolvedValue({}),
    post: jest.fn().mockResolvedValue({}),
    put: jest.fn().mockResolvedValue({}),
    patch: jest.fn().mockResolvedValue({}),
    delete: jest.fn().mockResolvedValue({}),
  },
  ApiRequestError: class ApiRequestError extends Error {
    code: string;
    status: number;
    constructor(status: number, error: { code: string; message: string }) {
      super(error.message);
      this.code = error.code;
      this.status = status;
    }
  },
  ACCESS_TOKEN_KEY: 'offgridflow_access_token',
  REFRESH_TOKEN_KEY: 'offgridflow_refresh_token',
  TENANT_ID_KEY: 'offgridflow_tenant_id',
}));

// Mock billing
jest.mock('@/lib/billing', () => ({
  getSubscription: jest.fn().mockResolvedValue({
    plan_id: 'pro',
    status: 'active',
    current_period_end: '2024-12-31',
  }),
  formatSubscriptionStatus: jest.fn((s) => s || 'None'),
  formatPeriodEnd: jest.fn((d) => d || 'N/A'),
}));

describe('Component Rendering Tests', () => {
  describe('AppHeader Component', () => {
    it('should render without crashing', async () => {
      // Dynamic import to avoid issues with module mocking
      const { AppHeader } = await import('@/app/components/AppHeader');
      
      render(<AppHeader />);
      
      expect(screen.getByText('OffGridFlow')).toBeInTheDocument();
    });

    it('should show navigation links when authenticated', async () => {
      const { AppHeader } = await import('@/app/components/AppHeader');
      
      render(<AppHeader />);
      
      expect(screen.getByText('Dashboard')).toBeInTheDocument();
      expect(screen.getByText('Emissions')).toBeInTheDocument();
      expect(screen.getByText('Settings')).toBeInTheDocument();
    });

    it('should show user email when authenticated', async () => {
      const { AppHeader } = await import('@/app/components/AppHeader');
      
      render(<AppHeader />);
      
      expect(screen.getByText('test@example.com')).toBeInTheDocument();
    });

    it('should show sign out button when authenticated', async () => {
      const { AppHeader } = await import('@/app/components/AppHeader');
      
      render(<AppHeader />);
      
      expect(screen.getByText('Sign out')).toBeInTheDocument();
    });
  });

  describe('Providers Component', () => {
    it('should render children', async () => {
      const { AppProviders } = await import('@/app/providers');
      
      render(
        <AppProviders>
          <div data-testid="child">Test Child</div>
        </AppProviders>
      );
      
      expect(screen.getByTestId('child')).toHaveTextContent('Test Child');
    });
  });
});

describe('Page Smoke Tests', () => {
  // These tests verify pages can be imported without errors
  
  it('should be able to import login page', async () => {
    const LoginPage = await import('@/app/login/page');
    expect(LoginPage.default).toBeDefined();
  });

  it('should be able to import register page', async () => {
    const RegisterPage = await import('@/app/register/page');
    expect(RegisterPage.default).toBeDefined();
  });

  it('should be able to import settings page', async () => {
    const SettingsPage = await import('@/app/settings/page');
    expect(SettingsPage.default).toBeDefined();
  });

  it('should be able to import forgot password page', async () => {
    const ForgotPage = await import('@/app/password/forgot/page');
    expect(ForgotPage.default).toBeDefined();
  });

  it('should be able to import reset password page', async () => {
    const ResetPage = await import('@/app/password/reset/page');
    expect(ResetPage.default).toBeDefined();
  });
});
