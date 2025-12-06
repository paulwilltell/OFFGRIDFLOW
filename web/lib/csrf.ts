const CSRF_ENDPOINT = '/api/auth/csrf-token';
export const CSRF_HEADER_NAME = 'X-CSRF-Token';

let cachedToken: string | null = null;
let tokenPromise: Promise<string> | null = null;

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
  if (cachedToken) {
    return cachedToken;
  }

  if (!tokenPromise) {
    tokenPromise = requestToken()
      .then((token) => {
        cachedToken = token;
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
  tokenPromise = null;
}

export async function attachCSRFHeader(headers: Headers): Promise<void> {
  const token = await getCSRFToken();
  headers.set(CSRF_HEADER_NAME, token);
}
