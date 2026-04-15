'use client';

import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { usePathname, useRouter } from 'next/navigation';
import { api, ACCESS_TOKEN_KEY, REFRESH_TOKEN_KEY, TENANT_ID_KEY, ApiRequestError } from './api';

export interface Tenant {
  id: string;
  name: string;
}

export interface SessionUser {
  id: string;
  email: string;
  name: string;
  company?: string;
  firstName?: string;
  lastName?: string;
  role: string;
  default_tenant_id?: string;
  tenants?: Tenant[];
  two_factor_enabled?: boolean;
  email_verified?: boolean;
}

export interface SessionState {
  user: SessionUser | null;
  currentTenantId: string | null;
  tenants: Tenant[];
  accessToken: string | null;
  refreshToken: string | null;
  loading: boolean;
  isAuthenticated: boolean;
}

export interface LoginPayload {
  email: string;
  password: string;
  tenant_id?: string;
}

export interface LoginResponse {
  token: string;
  user: SessionUser;
  tenant?: Tenant;
}

type SessionContextValue = SessionState & {
  login: (payload: LoginPayload) => Promise<LoginResponse>;
  logout: () => Promise<void>;
  switchTenant: (tenantId: string) => Promise<void>;
  refreshSession: () => Promise<void>;
};

const SessionContext = createContext<SessionContextValue | undefined>(undefined);

const ACCESS_KEY = ACCESS_TOKEN_KEY;
const REFRESH_KEY = REFRESH_TOKEN_KEY;
const TENANT_KEY = TENANT_ID_KEY;

function persistTenant(tenantId?: string | null) {
  if (typeof window === 'undefined') return;
  if (tenantId) {
    localStorage.setItem(TENANT_KEY, tenantId);
  } else {
    localStorage.removeItem(TENANT_KEY);
  }
}

function getStoredTenant() {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(TENANT_KEY);
}

function clearLegacyTokens() {
  if (typeof window === 'undefined') return;
  localStorage.removeItem(ACCESS_KEY);
  localStorage.removeItem(REFRESH_KEY);
  localStorage.removeItem('auth_token');
  localStorage.removeItem('offgridflow_session');
}

function shouldBootstrapSession(pathname: string | null): boolean {
  if (!pathname) return false;

  return [
    '/audit',
    '/blockchain',
    '/compliance',
    '/dashboard',
    '/emissions',
    '/onboarding',
    '/settings',
    '/workflow',
  ].some((prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`));
}

export function SessionProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();

  const [state, setState] = useState<SessionState>({
    user: null,
    currentTenantId: null,
    tenants: [],
    accessToken: null,
    refreshToken: null,
    loading: true,
    isAuthenticated: false,
  });

  const setSession = useCallback((payload: LoginResponse) => {
    const { user, tenant } = payload;
    const storedTenant = getStoredTenant();
    const resolvedTenant =
      tenant?.id || user.default_tenant_id || storedTenant || user.tenants?.[0]?.id || null;

    clearLegacyTokens();
    persistTenant(resolvedTenant);

    setState({
      user,
      currentTenantId: resolvedTenant,
      tenants: user.tenants ?? [],
      accessToken: null,
      refreshToken: null,
      loading: false,
      isAuthenticated: true,
    });
  }, []);

  const clearSession = useCallback(() => {
    clearLegacyTokens();
    persistTenant(null);
    setState({
      user: null,
      currentTenantId: null,
      tenants: [],
      accessToken: null,
      refreshToken: null,
      loading: false,
      isAuthenticated: false,
    });
  }, []);

  const refreshSession = useCallback(async () => {
    try {
      const res = await api.get<LoginResponse>('/api/auth/me');
      const resolvedTenant =
        getStoredTenant() ||
        res.tenant?.id ||
        res.user?.default_tenant_id ||
        res.user?.tenants?.[0]?.id ||
        null;

      persistTenant(resolvedTenant);

      setState((prev) => ({
        ...prev,
        user: res.user ?? null,
        currentTenantId: resolvedTenant,
        tenants: res.user?.tenants ?? [],
        accessToken: null,
        refreshToken: null,
        loading: false,
        isAuthenticated: true,
      }));
    } catch (err) {
      // Only clear the session on explicit auth failures (401/403).
      // Network errors, 500s, and transient failures must NOT log the user out
      // — otherwise any flaky connection kicks paying customers back to login.
      const isAuthFailure =
        err instanceof ApiRequestError && (err.status === 401 || err.status === 403);

      if (isAuthFailure) {
        clearSession();
      } else {
        // Transient/network error: preserve token, end loading, keep current
        // authenticated state. The user keeps their session; API calls will
        // retry naturally as the user interacts.
        setState((prev) => ({
          ...prev,
          loading: false,
          isAuthenticated: Boolean(prev.user),
        }));
      }
    }
  }, [clearSession]);

  useEffect(() => {
    if (!shouldBootstrapSession(pathname)) {
      setState((prev) => (prev.loading ? { ...prev, loading: false } : prev));
      return;
    }

    void refreshSession();
  }, [pathname, refreshSession]);

  const login = useCallback(
    async (payload: LoginPayload): Promise<LoginResponse> => {
      const res = await api.post<LoginResponse>('/api/auth/login', payload);
      setSession(res);
      return res;
    },
    [setSession],
  );

  const logout = useCallback(async () => {
    try {
      await api.post('/api/auth/logout', {});
    } catch {
      // ignore logout errors
    }
    clearSession();
    router.replace('/login');
  }, [clearSession, router]);

  const switchTenant = useCallback(
    async (tenantId: string) => {
      persistTenant(tenantId);
      setState((prev) => ({
        ...prev,
        currentTenantId: tenantId,
      }));
    },
    [],
  );

  const value: SessionContextValue = useMemo(
    () => ({
      ...state,
      login,
      logout,
      switchTenant,
      refreshSession,
    }),
    [state, login, logout, switchTenant, refreshSession],
  );

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

const defaultSessionValue: SessionContextValue = {
  user: null,
  currentTenantId: null,
  tenants: [],
  accessToken: null,
  refreshToken: null,
  loading: true,
  isAuthenticated: false,
  login: async () => { throw new Error('No SessionProvider'); },
  logout: async () => {},
  switchTenant: async () => {},
  refreshSession: async () => {},
};

export function useSession(): SessionContextValue {
  const ctx = useContext(SessionContext);
  if (!ctx) {
    // Return default during SSR/prerender instead of throwing
    return defaultSessionValue;
  }
  return ctx;
}

/**
 * Simple guard helper for client components to redirect unauthenticated users.
 */
export function useRequireAuth() {
  const session = useSession();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (session.loading) return;
    if (session.isAuthenticated) return;

    const returnTo = encodeURIComponent(pathname || '/');
    router.replace(`/login?returnTo=${returnTo}`);
  }, [session.loading, session.isAuthenticated, pathname, router]);

  return session;
}
