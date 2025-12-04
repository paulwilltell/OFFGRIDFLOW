import { api, ACCESS_TOKEN_KEY, TENANT_ID_KEY } from './api';

export interface User {
  id: string;
  email: string;
  name: string;
  role: string;
  default_tenant_id?: string;
  tenants?: Tenant[];
  two_factor_enabled?: boolean;
}

export interface Tenant {
  id: string;
  name: string;
}

export interface Session {
  user: User;
  tenant: Tenant | null;
  token?: string;
}

export interface AuthResponse {
  token: string;
  user: User;
  tenant?: Tenant;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
  company_name?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

// Storage keys
const SESSION_KEY = 'offgridflow_session';

/**
 * Register a new account
 */
export async function register(data: RegisterRequest): Promise<AuthResponse> {
  const response = await api.post<AuthResponse>('/api/auth/register', data);
  if (response.token) {
    storeSession(response);
  }
  return response;
}

/**
 * Login with email and password
 */
export async function login(data: LoginRequest): Promise<AuthResponse> {
  const response = await api.post<AuthResponse>('/api/auth/login', data);
  if (response.token) {
    storeSession(response);
  }
  return response;
}

/**
 * Logout the current user
 */
export async function logout(): Promise<void> {
  try {
    await api.post('/api/auth/logout', {});
  } catch (e) {
    // Ignore errors during logout
  }
  clearSession();
}

/**
 * Get the current authenticated user
 */
export async function getCurrentUser(): Promise<AuthResponse | null> {
  try {
    const response = await api.get<AuthResponse>('/api/auth/me');
    return response;
  } catch (e) {
    return null;
  }
}

/**
 * Change password for current user
 */
export async function changePassword(currentPassword: string, newPassword: string): Promise<void> {
  await api.post('/api/auth/change-password', {
    current_password: currentPassword,
    new_password: newPassword,
  });
}

/**
 * Get session from local storage
 */
export function getSession(): Session | null {
  if (typeof window === 'undefined') return null;
  
  const stored = localStorage.getItem(SESSION_KEY);
  if (!stored) return null;
  
  try {
    return JSON.parse(stored) as Session;
  } catch {
    return null;
  }
}

/**
 * Get token from local storage
 */
export function getToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

/**
 * Store session in local storage
 */
function storeSession(auth: AuthResponse): void {
  if (typeof window === 'undefined') return;
  
  const session: Session = {
    user: auth.user,
    tenant: auth.tenant ?? null,
    token: auth.token,
  };
  
  localStorage.setItem(SESSION_KEY, JSON.stringify(session));
  localStorage.setItem(ACCESS_TOKEN_KEY, auth.token);
  if (auth.user.default_tenant_id) {
    localStorage.setItem(TENANT_ID_KEY, auth.user.default_tenant_id);
  }
}

/**
 * Clear session from local storage
 */
function clearSession(): void {
  if (typeof window === 'undefined') return;
  
  localStorage.removeItem(SESSION_KEY);
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(TENANT_ID_KEY);
}

/**
 * Check if user is authenticated
 */
export function isAuthenticated(): boolean {
  return getSession() !== null;
}

/**
 * Check if user has a specific role
 */
export function hasRole(role: string): boolean {
  const session = getSession();
  if (!session) return false;
  return session.user.role === role;
}

/**
 * React hook for auth state (use in client components)
 */
export function useAuth() {
  const session = getSession();
  
  return {
    isAuthenticated: session !== null,
    user: session?.user ?? null,
    tenant: session?.tenant ?? null,
    login,
    logout,
    register,
  };
}
