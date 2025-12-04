/**
 * @jest-environment jsdom
 */
import React from 'react';
import { render, screen, act, waitFor } from '@testing-library/react';
import { SessionProvider, useSession, useRequireAuth } from '@/lib/session';

// Mock next/navigation
const mockPush = jest.fn();
const mockReplace = jest.fn();
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
    replace: mockReplace,
    prefetch: jest.fn(),
    back: jest.fn(),
  }),
  usePathname: () => '/dashboard',
}));

// Mock the API client
const mockApiGet = jest.fn();
const mockApiPost = jest.fn();
jest.mock('@/lib/api', () => ({
  api: {
    get: (...args: unknown[]) => mockApiGet(...args),
    post: (...args: unknown[]) => mockApiPost(...args),
  },
  ACCESS_TOKEN_KEY: 'offgridflow_access_token',
  REFRESH_TOKEN_KEY: 'offgridflow_refresh_token',
  TENANT_ID_KEY: 'offgridflow_tenant_id',
}));

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: jest.fn((key: string) => store[key] || null),
    setItem: jest.fn((key: string, value: string) => {
      store[key] = value;
    }),
    removeItem: jest.fn((key: string) => {
      delete store[key];
    }),
    clear: jest.fn(() => {
      store = {};
    }),
  };
})();
Object.defineProperty(window, 'localStorage', { value: localStorageMock });

// Test component to access session
function TestSessionConsumer() {
  const session = useSession();
  return (
    <div>
      <span data-testid="loading">{session.loading ? 'true' : 'false'}</span>
      <span data-testid="authenticated">{session.isAuthenticated ? 'true' : 'false'}</span>
      <span data-testid="user">{session.user?.email || 'none'}</span>
    </div>
  );
}

// Test component for useRequireAuth
function ProtectedComponent() {
  const session = useRequireAuth();
  return (
    <div>
      <span data-testid="protected-loading">{session.loading ? 'true' : 'false'}</span>
      <span data-testid="protected-auth">{session.isAuthenticated ? 'true' : 'false'}</span>
    </div>
  );
}

describe('Session Module', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorageMock.clear();
    mockApiGet.mockReset();
    mockApiPost.mockReset();
  });

  describe('SessionProvider', () => {
    it('should render children', async () => {
      mockApiGet.mockRejectedValueOnce(new Error('No session'));

      await act(async () => {
        render(
          <SessionProvider>
            <div data-testid="child">Child Content</div>
          </SessionProvider>
        );
      });

      expect(screen.getByTestId('child')).toHaveTextContent('Child Content');
    });

    it('should start with loading state', async () => {
      mockApiGet.mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve({ user: null }), 100))
      );

      await act(async () => {
        render(
          <SessionProvider>
            <TestSessionConsumer />
          </SessionProvider>
        );
      });

      // After initial render, it attempts to refresh session
      await waitFor(() => {
        expect(screen.getByTestId('loading')).toHaveTextContent('false');
      });
    });

    it('should set authenticated state when user exists', async () => {
      localStorageMock.setItem('offgridflow_access_token', 'test-token');
      mockApiGet.mockResolvedValueOnce({
        user: {
          id: '1',
          email: 'test@example.com',
          name: 'Test User',
          role: 'admin',
        },
        token: 'test-token',
      });

      await act(async () => {
        render(
          <SessionProvider>
            <TestSessionConsumer />
          </SessionProvider>
        );
      });

      await waitFor(() => {
        expect(screen.getByTestId('authenticated')).toHaveTextContent('true');
        expect(screen.getByTestId('user')).toHaveTextContent('test@example.com');
      });
    });

    it('should clear session when no token exists', async () => {
      mockApiGet.mockRejectedValueOnce(new Error('Unauthorized'));

      await act(async () => {
        render(
          <SessionProvider>
            <TestSessionConsumer />
          </SessionProvider>
        );
      });

      await waitFor(() => {
        expect(screen.getByTestId('authenticated')).toHaveTextContent('false');
        expect(screen.getByTestId('user')).toHaveTextContent('none');
      });
    });
  });

  describe('useSession', () => {
    it('should throw error when used outside provider', () => {
      const consoleError = jest.spyOn(console, 'error').mockImplementation(() => {});

      expect(() => {
        render(<TestSessionConsumer />);
      }).toThrow('useSession must be used within a SessionProvider');

      consoleError.mockRestore();
    });
  });

  describe('useRequireAuth', () => {
    it('should redirect when not authenticated', async () => {
      mockApiGet.mockRejectedValueOnce(new Error('Unauthorized'));

      await act(async () => {
        render(
          <SessionProvider>
            <ProtectedComponent />
          </SessionProvider>
        );
      });

      await waitFor(() => {
        expect(mockReplace).toHaveBeenCalledWith('/login?returnTo=%2Fdashboard');
      });
    });

    it('should not redirect when authenticated', async () => {
      localStorageMock.setItem('offgridflow_access_token', 'test-token');
      mockApiGet.mockResolvedValueOnce({
        user: {
          id: '1',
          email: 'test@example.com',
          name: 'Test User',
          role: 'admin',
        },
        token: 'test-token',
      });

      await act(async () => {
        render(
          <SessionProvider>
            <ProtectedComponent />
          </SessionProvider>
        );
      });

      await waitFor(() => {
        expect(screen.getByTestId('protected-auth')).toHaveTextContent('true');
      });

      // Should not have called replace
      expect(mockReplace).not.toHaveBeenCalled();
    });
  });

  describe('Login flow', () => {
    it('should handle successful login', async () => {
    mockApiGet.mockRejectedValueOnce(new Error('No session'));
    mockApiPost.mockResolvedValueOnce({
      token: 'new-token',
      user: {
        id: '1',
        email: 'test@example.com',
        name: 'Test User',
        role: 'admin',
      },
    });

      let loginFn: Function;

      function LoginTestComponent() {
        const session = useSession();
        loginFn = session.login;
        return (
          <div>
            <span data-testid="auth">{session.isAuthenticated ? 'yes' : 'no'}</span>
          </div>
        );
      }

      await act(async () => {
        render(
          <SessionProvider>
            <LoginTestComponent />
          </SessionProvider>
        );
      });

      await act(async () => {
        await loginFn!({ email: 'test@example.com', password: 'password123' });
      });

      await waitFor(() => {
        expect(screen.getByTestId('auth')).toHaveTextContent('yes');
      });

      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        'offgridflow_access_token',
        'new-token'
      );
    });

  });

  describe('Logout flow', () => {
    it('should clear session and redirect on logout', async () => {
      localStorageMock.setItem('offgridflow_access_token', 'test-token');
      mockApiGet.mockResolvedValueOnce({
        user: {
          id: '1',
          email: 'test@example.com',
          name: 'Test User',
          role: 'admin',
        },
      });
      mockApiPost.mockResolvedValueOnce({});

      let logoutFn: Function;

      function LogoutTestComponent() {
        const session = useSession();
        logoutFn = session.logout;
        return (
          <div>
            <span data-testid="auth">{session.isAuthenticated ? 'yes' : 'no'}</span>
          </div>
        );
      }

      await act(async () => {
        render(
          <SessionProvider>
            <LogoutTestComponent />
          </SessionProvider>
        );
      });

      await waitFor(() => {
        expect(screen.getByTestId('auth')).toHaveTextContent('yes');
      });

      await act(async () => {
        await logoutFn!();
      });

      await waitFor(() => {
        expect(screen.getByTestId('auth')).toHaveTextContent('no');
      });

      expect(mockReplace).toHaveBeenCalledWith('/login');
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('offgridflow_access_token');
    });
  });
});
