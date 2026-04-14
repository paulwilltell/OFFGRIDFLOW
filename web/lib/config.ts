/**
 * Configuration helper for OffGridFlow frontend.
 *
 * Browser requests should stay same-origin so the frontend can own the
 * authenticated session cookie. Server-side code can still call the Railway API
 * origin directly.
 */

export const DEFAULT_API_ORIGIN = 'https://offgridflow-api-v2-production.up.railway.app';

export function resolveApiBaseUrl(): string {
  if (typeof window !== 'undefined') {
    return '';
  }

  return process.env.OFFGRIDFLOW_API_ORIGIN || DEFAULT_API_ORIGIN;
}

export const config = {
  apiBaseUrl: resolveApiBaseUrl(),
  apiOrigin: process.env.OFFGRIDFLOW_API_ORIGIN || DEFAULT_API_ORIGIN,
};
