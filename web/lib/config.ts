/**
 * Configuration helper for OffGridFlow frontend.
 * Reads API base URL from environment variables with sensible defaults.
 */

export const config = {
  /**
   * Base URL for the OffGridFlow API.
   * Set NEXT_PUBLIC_OFFGRIDFLOW_API_URL in your environment for production.
   * Defaults to http://localhost:8090 for local development.
   */
  apiBaseUrl: process.env.NEXT_PUBLIC_OFFGRIDFLOW_API_URL || 'http://localhost:8090',
};
