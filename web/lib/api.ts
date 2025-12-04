import { config } from './config';

export interface ApiError {
  code: string;
  message: string;
  detail?: string;
}

export class ApiRequestError extends Error {
  code: string;
  status: number;
  detail?: string;

  constructor(status: number, error: ApiError) {
    super(error.message);
    this.name = 'ApiRequestError';
    this.code = error.code;
    this.status = status;
    this.detail = error.detail;
  }
}

export type ApiClient = {
  get: <T>(path: string) => Promise<T>;
  post: <T>(path: string, body: unknown) => Promise<T>;
  put: <T>(path: string, body: unknown) => Promise<T>;
  patch: <T>(path: string, body: unknown) => Promise<T>;
  delete: <T>(path: string) => Promise<T>;
};

// Storage keys for auth/session state
export const ACCESS_TOKEN_KEY = 'offgridflow_access_token';
export const REFRESH_TOKEN_KEY = 'offgridflow_refresh_token';
export const TENANT_ID_KEY = 'offgridflow_tenant_id';

/**
 * Creates an API client configured to talk to the OffGridFlow backend.
 * @param baseUrl - Override base URL (defaults to config.apiBaseUrl)
 */
export function createClient(baseUrl: string = config.apiBaseUrl): ApiClient {
  const getAuthToken = (): string | null => {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(ACCESS_TOKEN_KEY);
  };

  const getTenantId = (): string | null => {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(TENANT_ID_KEY);
  };

  const buildHeaders = (): HeadersInit => {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };
    const token = getAuthToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    const tenantId = getTenantId();
    if (tenantId) {
      headers['X-Tenant-ID'] = tenantId;
    }
    return headers;
  };

  const handleResponse = async <T>(response: Response, path: string): Promise<T> => {
    if (!response.ok) {
      let error: ApiError = {
        code: 'unknown_error',
        message: `Request failed with status ${response.status}`,
      };
      
      try {
        const body = await response.json();
        if (body.code && body.message) {
          error = body as ApiError;
        }
      } catch {
        // Response wasn't JSON, use default error
      }

      throw new ApiRequestError(response.status, error);
    }

    // Handle empty responses
    const text = await response.text();
    if (!text) {
      return {} as T;
    }

    return JSON.parse(text) as T;
  };

  return {
    get: async <T>(path: string): Promise<T> => {
      const response = await fetch(`${baseUrl}${path}`, {
        method: 'GET',
        headers: buildHeaders(),
        credentials: 'include', // Include cookies for session auth
      });
      return handleResponse<T>(response, path);
    },

    post: async <T>(path: string, body: unknown): Promise<T> => {
      const response = await fetch(`${baseUrl}${path}`, {
        method: 'POST',
        headers: buildHeaders(),
        credentials: 'include',
        body: JSON.stringify(body),
      });
      return handleResponse<T>(response, path);
    },

    put: async <T>(path: string, body: unknown): Promise<T> => {
      const response = await fetch(`${baseUrl}${path}`, {
        method: 'PUT',
        headers: buildHeaders(),
        credentials: 'include',
        body: JSON.stringify(body),
      });
      return handleResponse<T>(response, path);
    },

    patch: async <T>(path: string, body: unknown): Promise<T> => {
      const response = await fetch(`${baseUrl}${path}`, {
        method: 'PATCH',
        headers: buildHeaders(),
        credentials: 'include',
        body: JSON.stringify(body),
      });
      return handleResponse<T>(response, path);
    },

    delete: async <T>(path: string): Promise<T> => {
      const response = await fetch(`${baseUrl}${path}`, {
        method: 'DELETE',
        headers: buildHeaders(),
        credentials: 'include',
      });
      return handleResponse<T>(response, path);
    },
  };
}

// Singleton client for convenience
export const api = createClient();
