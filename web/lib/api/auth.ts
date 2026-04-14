/**
 * Authentication API Client
 * 
 * Handles user authentication, registration, password management, and API keys
 */

import { APIError, RequestOptions } from './activities';
import { config } from '../config';

const API_BASE = config.apiBaseUrl;

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
 * Build headers
 */
const buildHeaders = (): HeadersInit => {
  return {
    'Content-Type': 'application/json',
  };
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
    credentials: 'include',
    body: JSON.stringify(data),
    signal: options.signal,
  });

  return handleResponse<AuthResponse>(response);
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
    credentials: 'include',
    body: JSON.stringify(data),
    signal: options.signal,
  });

  return handleResponse<AuthResponse>(response);
};

/**
 * Logout user
 */
export const logout = async (options: RequestOptions = {}): Promise<void> => {
  const response = await fetch(`${API_BASE}/api/auth/logout`, {
    method: 'POST',
    headers: buildHeaders(),
    credentials: 'include',
    signal: options.signal,
  });

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
    headers: buildHeaders(),
    credentials: 'include',
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
    headers: buildHeaders(),
    credentials: 'include',
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
    credentials: 'include',
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
    credentials: 'include',
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
    headers: buildHeaders(),
    credentials: 'include',
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
    headers: buildHeaders(),
    credentials: 'include',
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
    headers: buildHeaders(),
    credentials: 'include',
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Check if user is authenticated
 */
export const isAuthenticated = (): boolean => {
  if (typeof window === 'undefined') return false;
  return Boolean(localStorage.getItem('auth_token') || localStorage.getItem('offgridflow_access_token'));
};

/**
 * Refresh authentication token (if endpoint exists)
 */
export const refreshToken = async (
  options: RequestOptions = {}
): Promise<AuthResponse> => {
  const response = await fetch(`${API_BASE}/api/auth/refresh`, {
    method: 'POST',
    headers: buildHeaders(),
    credentials: 'include',
    signal: options.signal,
  });

  return handleResponse<AuthResponse>(response);
};
