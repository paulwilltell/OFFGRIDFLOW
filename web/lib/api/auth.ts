/**
 * Authentication API Client
 * 
 * Handles user authentication, registration, password management, and API keys
 */

import { APIError, RequestOptions } from './activities';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8090';

/**
 * User and auth types
 */
export interface User {
  id: string;
  email: string;
  name: string;
  role: string;
  tenantId: string;
  createdAt: string;
}

export interface AuthResponse {
  user: User;
  token: string;
  expiresAt: string;
}

export interface APIKey {
  id: string;
  name: string;
  key: string; // Only shown once on creation
  keyHash?: string;
  scopes: string[];
  expiresAt?: string;
  lastUsed?: string;
  createdAt: string;
}

/**
 * Set auth token in storage
 */
export const setAuthToken = (token: string): void => {
  if (typeof window !== 'undefined') {
    localStorage.setItem('auth_token', token);
  }
};

/**
 * Get auth token from storage
 */
export const getAuthToken = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('auth_token');
};

/**
 * Remove auth token from storage
 */
export const clearAuthToken = (): void => {
  if (typeof window !== 'undefined') {
    localStorage.removeItem('auth_token');
  }
};

/**
 * Build headers
 */
const buildHeaders = (includeAuth = false): HeadersInit => {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (includeAuth) {
    const token = getAuthToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
  }

  return headers;
};

/**
 * Handle response
 */
const handleResponse = async <T>(response: Response): Promise<T> => {
  if (!response.ok) {
    const error = await response.json().catch(() => ({
      error: `HTTP ${response.status}`,
    }));

    throw new APIError(
      error.error || 'Request failed',
      response.status,
      error.details
    );
  }

  return response.json();
};

/**
 * Register new user
 */
export const register = async (
  data: {
    email: string;
    password: string;
    name: string;
  },
  options: RequestOptions = {}
): Promise<AuthResponse> => {
  const response = await fetch(`${API_BASE}/api/auth/register`, {
    method: 'POST',
    headers: buildHeaders(),
    body: JSON.stringify(data),
    signal: options.signal,
  });

  const result = await handleResponse<AuthResponse>(response);

  // Save token
  setAuthToken(result.token);

  return result;
};

/**
 * Login user
 */
export const login = async (
  data: {
    email: string;
    password: string;
  },
  options: RequestOptions = {}
): Promise<AuthResponse> => {
  const response = await fetch(`${API_BASE}/api/auth/login`, {
    method: 'POST',
    headers: buildHeaders(),
    body: JSON.stringify(data),
    signal: options.signal,
  });

  const result = await handleResponse<AuthResponse>(response);

  // Save token
  setAuthToken(result.token);

  return result;
};

/**
 * Logout user
 */
export const logout = async (options: RequestOptions = {}): Promise<void> => {
  const response = await fetch(`${API_BASE}/api/auth/logout`, {
    method: 'POST',
    headers: buildHeaders(true),
    signal: options.signal,
  });

  // Clear token regardless of response
  clearAuthToken();

  if (!response.ok) {
    // Still clear token but log error
    console.warn('Logout request failed, but token cleared locally');
  }
};

/**
 * Get current user info
 */
export const getCurrentUser = async (
  options: RequestOptions = {}
): Promise<{ user: User }> => {
  const response = await fetch(`${API_BASE}/api/auth/me`, {
    method: 'GET',
    headers: buildHeaders(true),
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Change password
 */
export const changePassword = async (
  data: {
    currentPassword: string;
    newPassword: string;
  },
  options: RequestOptions = {}
): Promise<{ message: string }> => {
  const response = await fetch(`${API_BASE}/api/auth/change-password`, {
    method: 'POST',
    headers: buildHeaders(true),
    body: JSON.stringify(data),
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Request password reset
 */
export const forgotPassword = async (
  email: string,
  options: RequestOptions = {}
): Promise<{ message: string }> => {
  const response = await fetch(`${API_BASE}/api/auth/password/forgot`, {
    method: 'POST',
    headers: buildHeaders(),
    body: JSON.stringify({ email }),
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Reset password with token
 */
export const resetPassword = async (
  data: {
    token: string;
    newPassword: string;
  },
  options: RequestOptions = {}
): Promise<{ message: string }> => {
  const response = await fetch(`${API_BASE}/api/auth/password/reset`, {
    method: 'POST',
    headers: buildHeaders(),
    body: JSON.stringify(data),
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Create API key
 */
export const createAPIKey = async (
  data: {
    name: string;
    scopes: string[];
    expiresAt?: string;
  },
  options: RequestOptions = {}
): Promise<{ key: APIKey }> => {
  const response = await fetch(`${API_BASE}/api/auth/keys`, {
    method: 'POST',
    headers: buildHeaders(true),
    body: JSON.stringify(data),
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * List API keys
 */
export const listAPIKeys = async (
  options: RequestOptions = {}
): Promise<{ keys: APIKey[] }> => {
  const response = await fetch(`${API_BASE}/api/auth/keys`, {
    method: 'GET',
    headers: buildHeaders(true),
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Revoke API key
 */
export const revokeAPIKey = async (
  keyId: string,
  options: RequestOptions = {}
): Promise<{ message: string }> => {
  const response = await fetch(`${API_BASE}/api/auth/keys/${keyId}`, {
    method: 'DELETE',
    headers: buildHeaders(true),
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Check if user is authenticated
 */
export const isAuthenticated = (): boolean => {
  return getAuthToken() !== null;
};

/**
 * Refresh authentication token (if endpoint exists)
 */
export const refreshToken = async (
  options: RequestOptions = {}
): Promise<AuthResponse> => {
  const response = await fetch(`${API_BASE}/api/auth/refresh`, {
    method: 'POST',
    headers: buildHeaders(true),
    signal: options.signal,
  });

  const result = await handleResponse<AuthResponse>(response);

  // Update token
  setAuthToken(result.token);

  return result;
};
