/**
 * @fileoverview Sentry error tracking integration for OffGridFlow
 * @description Configures Sentry for error monitoring, performance tracking,
 * and session replay in production environments.
 * 
 * @see https://docs.sentry.io/platforms/javascript/guides/nextjs/
 */

import * as Sentry from '@sentry/nextjs';

/**
 * Sentry configuration options
 */
const SENTRY_DSN = process.env.NEXT_PUBLIC_SENTRY_DSN;
const ENVIRONMENT = process.env.NODE_ENV || 'development';
const RELEASE_VERSION = process.env.NEXT_PUBLIC_APP_VERSION || '1.0.0';

/**
 * Initialize Sentry error tracking
 * 
 * Features enabled:
 * - Automatic error capture
 * - Performance monitoring (20% of transactions)
 * - Session replay for error reproduction
 * - Source maps for readable stack traces
 * - Environment tagging
 * 
 * @example
 * // Initialize in app entry point
 * import { initSentry } from '@/lib/sentry';
 * initSentry();
 */
export function initSentry(): void {
  if (!SENTRY_DSN) {
    if (ENVIRONMENT === 'development') {
      console.info('[Sentry] DSN not configured, skipping initialization');
    }
    return;
  }

  Sentry.init({
    dsn: SENTRY_DSN,
    environment: ENVIRONMENT,
    release: `offgridflow@${RELEASE_VERSION}`,
    
    // Performance Monitoring
    tracesSampleRate: ENVIRONMENT === 'production' ? 0.2 : 1.0,
    
    // Session Replay
    replaysSessionSampleRate: 0.1,
    replaysOnErrorSampleRate: 1.0,
    
    // Integrations
    integrations: [
      Sentry.browserTracingIntegration(),
      Sentry.replayIntegration({
        maskAllText: false,
        blockAllMedia: false,
      }),
    ],

    // Filter out noise
    ignoreErrors: [
      // Browser extensions
      'ResizeObserver loop limit exceeded',
      'ResizeObserver loop completed with undelivered notifications',
      // Network errors (handled by app)
      'Failed to fetch',
      'NetworkError',
      'Load failed',
      // User cancellation
      'AbortError',
      // Third-party scripts
      /^chrome-extension:\/\//,
      /^moz-extension:\/\//,
    ],

    // Custom filtering
    beforeSend(event, hint) {
      const error = hint.originalException;
      
      // Don't send errors from development
      if (ENVIRONMENT === 'development') {
        console.error('[Sentry] Error captured (not sent in dev):', error);
        return null;
      }

      // Add custom context
      event.tags = {
        ...event.tags,
        app_version: RELEASE_VERSION,
      };

      return event;
    },

    // Performance filtering
    beforeSendTransaction(event) {
      // Filter out health check transactions
      if (event.transaction?.includes('/api/health')) {
        return null;
      }
      return event;
    },
  });
}

/**
 * Capture a custom error with additional context
 * 
 * @param error - The error to capture
 * @param context - Additional context about the error
 * 
 * @example
 * try {
 *   await riskyOperation();
 * } catch (error) {
 *   captureError(error, {
 *     operation: 'fetchEmissions',
 *     userId: user.id,
 *   });
 * }
 */
export function captureError(
  error: Error | unknown,
  context?: Record<string, unknown>
): void {
  Sentry.captureException(error, {
    extra: context,
  });
}

/**
 * Capture a custom message/event
 * 
 * @param message - The message to capture
 * @param level - Severity level
 * @param context - Additional context
 * 
 * @example
 * captureMessage('User exported compliance report', 'info', {
 *   format: 'pdf',
 *   timeframe: 'yearly',
 * });
 */
export function captureMessage(
  message: string,
  level: Sentry.SeverityLevel = 'info',
  context?: Record<string, unknown>
): void {
  Sentry.captureMessage(message, {
    level,
    extra: context,
  });
}

/**
 * Set user context for error tracking
 * 
 * @param user - User information
 * 
 * @example
 * setUser({
 *   id: user.id,
 *   email: user.email,
 *   tenantId: user.tenantId,
 * });
 */
export function setUser(user: {
  id: string;
  email?: string;
  username?: string;
  tenantId?: string;
} | null): void {
  Sentry.setUser(user);
}

/**
 * Add breadcrumb for debugging
 * 
 * @param breadcrumb - Breadcrumb data
 * 
 * @example
 * addBreadcrumb({
 *   category: 'navigation',
 *   message: 'Navigated to carbon dashboard',
 *   level: 'info',
 * });
 */
export function addBreadcrumb(breadcrumb: Sentry.Breadcrumb): void {
  Sentry.addBreadcrumb(breadcrumb);
}

/**
 * Start a performance transaction
 * 
 * @param name - Transaction name
 * @param op - Operation type
 * @returns Transaction object
 * 
 * @example
 * const transaction = startTransaction('fetchEmissions', 'api.call');
 * try {
 *   await api.getEmissions();
 * } finally {
 *   transaction.finish();
 * }
 */
export function startTransaction(
  name: string,
  op: string
): Sentry.Span | undefined {
  return Sentry.startInactiveSpan({ name, op });
}

/**
 * Create a custom span for performance tracking
 * 
 * @param name - Span name
 * @param callback - Function to execute within span
 * @returns Result of callback
 * 
 * @example
 * const result = await withSpan('processEmissionData', async () => {
 *   return processData(emissions);
 * });
 */
export async function withSpan<T>(
  name: string,
  callback: () => Promise<T> | T
): Promise<T> {
  return Sentry.startSpan({ name }, callback);
}

export default {
  initSentry,
  captureError,
  captureMessage,
  setUser,
  addBreadcrumb,
  startTransaction,
  withSpan,
};
