import { config } from './config';

const CSRF_ENDPOINT = `${config.apiBaseUrl}/api/auth/csrf-token`;
export const CSRF_HEADER_NAME = 'X-CSRF-Token';

let cachedToken: string | null = null;
let cachedTokenExpiry: number = 0;
let tokenPromise: Promise<string> | null = null;

const CACHE_TTL_MS = 23 * 60 * 60 * 1000; // 23 hours

async function requestToken(): Promise<string> {
  const response = await fetch(CSRF_ENDPOINT, {
    method: 'GET',
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error('Failed to fetch CSRF token');
  }

  const payload = await response.json();
  if (!payload || typeof payload.csrf_token !== 'string') {
    throw new Error('Invalid CSRF token response');
  }

  return payload.csrf_token;
}

export async function getCSRFToken(): Promise<string> {
  if (cachedToken && Date.now() < cachedTokenExpiry) {
    return cachedToken;
  }

  // Cache expired or empty — clear stale value
  cachedToken = null;

  if (!tokenPromise) {
    tokenPromise = requestToken()
      .then((token) => {
        cachedToken = token;
        cachedTokenExpiry = Date.now() + CACHE_TTL_MS;
        return token;
      })
      .finally(() => {
        tokenPromise = null;
      });
  }

  return tokenPromise;
}

export function clearCSRFTokenCache(): void {
  cachedToken = null;
  cachedTokenExpiry = 0;
  tokenPromise = null;
}

export async function attachCSRFHeader(headers: Headers): Promise<void> {
  const token = await getCSRFToken();
  headers.set(CSRF_HEADER_NAME, token);
}
