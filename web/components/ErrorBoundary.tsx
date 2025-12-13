'use client';

/**
 * @fileoverview React Error Boundary components for graceful error handling
 * @description Provides component-level and page-level error boundaries with
 * telemetry integration, retry capabilities, and accessible error states.
 * 
 * @example
 * // Wrap a component with error boundary
 * <ErrorBoundary fallback={<ErrorFallback />}>
 *   <MyComponent />
 * </ErrorBoundary>
 * 
 * @example
 * // Use with custom error handler
 * <ErrorBoundary onError={(error, info) => logToSentry(error, info)}>
 *   <Dashboard />
 * </ErrorBoundary>
 */

import React, { Component, ErrorInfo, ReactNode, useCallback, useState } from 'react';

// ============================================================================
// Types
// ============================================================================

/**
 * Props for the ErrorBoundary component
 */
export interface ErrorBoundaryProps {
  /** Child components to wrap */
  children: ReactNode;
  /** Custom fallback UI to render on error */
  fallback?: ReactNode | ((props: FallbackProps) => ReactNode);
  /** Callback when an error is caught */
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
  /** Callback when reset is triggered */
  onReset?: () => void;
  /** Keys that trigger a reset when changed */
  resetKeys?: unknown[];
  /** Component name for error reporting */
  componentName?: string;
}

/**
 * Props passed to fallback render functions
 */
export interface FallbackProps {
  /** The error that was caught */
  error: Error;
  /** Component stack trace */
  errorInfo: ErrorInfo | null;
  /** Function to reset the error boundary */
  resetErrorBoundary: () => void;
  /** Component name where error occurred */
  componentName?: string;
}

/**
 * State for the ErrorBoundary component
 */
interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

// ============================================================================
// Error Boundary Class Component
// ============================================================================

/**
 * React Error Boundary component for catching JavaScript errors
 * in child component tree.
 * 
 * Features:
 * - Graceful error handling with customizable fallback UI
 * - Error telemetry integration (Sentry-ready)
 * - Retry/reset capability
 * - Reset on prop changes via resetKeys
 * - Accessible error states with ARIA attributes
 * 
 * @class ErrorBoundary
 * @extends Component<ErrorBoundaryProps, ErrorBoundaryState>
 */
export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  /**
   * Update state when an error is caught
   */
  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return { hasError: true, error };
  }

  /**
   * Log error information and trigger callbacks
   */
  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    this.setState({ errorInfo });

    // Log to console in development
    if (process.env.NODE_ENV === 'development') {
      console.group('🔴 Error Boundary Caught Error');
      console.error('Error:', error);
      console.error('Component Stack:', errorInfo.componentStack);
      console.groupEnd();
    }

    // Call custom error handler
    this.props.onError?.(error, errorInfo);

    // Send to error tracking service
    this.reportError(error, errorInfo);
  }

  /**
   * Check if resetKeys have changed to trigger a reset
   */
  componentDidUpdate(prevProps: ErrorBoundaryProps): void {
    const { resetKeys } = this.props;
    const { hasError } = this.state;

    if (hasError && resetKeys) {
      const prevResetKeys = prevProps.resetKeys || [];
      const hasResetKeyChanged = resetKeys.some(
        (key, index) => key !== prevResetKeys[index]
      );

      if (hasResetKeyChanged) {
        this.resetErrorBoundary();
      }
    }
  }

  /**
   * Report error to external tracking service (Sentry, DataDog, etc.)
   * @private
   */
  private reportError(error: Error, errorInfo: ErrorInfo): void {
    if (typeof window !== 'undefined' && (window as any).Sentry) {
      (window as any).Sentry.captureException(error, {
        contexts: {
          react: {
            componentStack: errorInfo.componentStack,
            componentName: this.props.componentName,
          },
        },
      });
    }
  }

  /**
   * Reset the error boundary state
   */
  resetErrorBoundary = (): void => {
    this.props.onReset?.();
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  render(): ReactNode {
    const { hasError, error, errorInfo } = this.state;
    const { children, fallback, componentName } = this.props;

    if (hasError && error) {
      // Render custom fallback if provided as function
      if (typeof fallback === 'function') {
        return fallback({
          error,
          errorInfo,
          resetErrorBoundary: this.resetErrorBoundary,
          componentName,
        });
      }

      // Render custom fallback if provided as element
      if (fallback) {
        return fallback;
      }

      // Default fallback UI
      return (
        <DefaultErrorFallback
          error={error}
          errorInfo={errorInfo}
          resetErrorBoundary={this.resetErrorBoundary}
          componentName={componentName}
        />
      );
    }

    return children;
  }
}

// ============================================================================
// Default Fallback Components
// ============================================================================

/**
 * Default error fallback UI with retry button and expandable details
 */
export const DefaultErrorFallback: React.FC<FallbackProps> = ({
  error,
  errorInfo,
  resetErrorBoundary,
  componentName,
}) => {
  const [showDetails, setShowDetails] = useState(false);

  return (
    <div
      role="alert"
      aria-live="assertive"
      aria-atomic="true"
      className="min-h-[200px] flex items-center justify-center p-6"
    >
      <div className="bg-red-900/20 border border-red-500/30 rounded-xl p-6 max-w-lg w-full">
        {/* Error Icon & Header */}
        <div className="flex items-center gap-3 mb-4">
          <div className="p-2 bg-red-500/20 rounded-lg" aria-hidden="true">
            <svg
              className="w-6 h-6 text-red-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
          </div>
          <div>
            <h3 className="text-lg font-semibold text-red-400">
              Something went wrong
            </h3>
            {componentName && (
              <p className="text-sm text-gray-500">Error in: {componentName}</p>
            )}
          </div>
        </div>

        {/* Error Message */}
        <p className="text-gray-300 mb-4">
          {error.message || 'An unexpected error occurred'}
        </p>

        {/* Actions */}
        <div className="flex items-center gap-3">
          <button
            onClick={resetErrorBoundary}
            className="px-4 py-2 bg-red-500 hover:bg-red-600 text-white font-medium rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 focus:ring-offset-gray-900"
            aria-label="Try again to reload the component"
          >
            Try Again
          </button>
          <button
            onClick={() => setShowDetails(!showDetails)}
            className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-gray-300 font-medium rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 focus:ring-offset-gray-900"
            aria-expanded={showDetails}
            aria-controls="error-details"
          >
            {showDetails ? 'Hide Details' : 'Show Details'}
          </button>
        </div>

        {/* Error Details (Collapsible) */}
        {showDetails && errorInfo && (
          <div
            id="error-details"
            className="mt-4 p-3 bg-gray-900/50 rounded-lg overflow-auto max-h-48"
          >
            <pre className="text-xs text-gray-400 whitespace-pre-wrap font-mono">
              <strong>Stack Trace:</strong>
              {'\n'}
              {error.stack}
              {'\n\n'}
              <strong>Component Stack:</strong>
              {errorInfo.componentStack}
            </pre>
          </div>
        )}
      </div>
    </div>
  );
};

/**
 * Minimal error fallback for inline components
 */
export const InlineErrorFallback: React.FC<FallbackProps> = ({
  resetErrorBoundary,
}) => (
  <div
    role="alert"
    className="inline-flex items-center gap-2 px-3 py-1.5 bg-red-500/10 border border-red-500/20 rounded-md text-sm"
  >
    <span className="text-red-400">⚠️ Error</span>
    <button
      onClick={resetErrorBoundary}
      className="text-red-400 hover:text-red-300 underline focus:outline-none focus:ring-1 focus:ring-red-400"
      aria-label="Retry loading this component"
    >
      Retry
    </button>
  </div>
);

/**
 * Chart-specific error fallback with visualization placeholder
 */
export const ChartErrorFallback: React.FC<FallbackProps> = ({
  resetErrorBoundary,
  componentName,
}) => (
  <div
    role="alert"
    aria-label={`Error loading ${componentName || 'chart'}`}
    className="bg-gray-800/50 rounded-xl border border-gray-700/50 p-6 min-h-[300px] flex flex-col items-center justify-center"
  >
    <div className="text-center">
      <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-red-500/10 flex items-center justify-center" aria-hidden="true">
        <svg
          className="w-8 h-8 text-red-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={1.5}
            d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
          />
        </svg>
      </div>
      <h3 className="text-lg font-semibold text-gray-300 mb-2">Chart Error</h3>
      <p className="text-gray-500 mb-4 text-sm max-w-xs">
        Unable to render {componentName || 'chart'}. This may be due to invalid data.
      </p>
      <button
        onClick={resetErrorBoundary}
        className="px-4 py-2 bg-cyan-500 hover:bg-cyan-600 text-white font-medium rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-cyan-500 focus:ring-offset-2 focus:ring-offset-gray-900"
        aria-label={`Reload ${componentName || 'chart'}`}
      >
        Reload Chart
      </button>
    </div>
  </div>
);

// ============================================================================
// Hooks
// ============================================================================

/**
 * Hook for programmatic error boundary control
 * 
 * @example
 * const { showBoundary, resetBoundary } = useErrorBoundary();
 * 
 * try {
 *   await riskyOperation();
 * } catch (error) {
 *   showBoundary(error as Error);
 * }
 * 
 * @returns Object with showBoundary and resetBoundary functions
 */
export function useErrorBoundary() {
  const [error, setError] = useState<Error | null>(null);

  const showBoundary = useCallback((error: Error) => {
    setError(error);
  }, []);

  const resetBoundary = useCallback(() => {
    setError(null);
  }, []);

  if (error) {
    throw error;
  }

  return { showBoundary, resetBoundary };
}

// ============================================================================
// Higher-Order Component
// ============================================================================

/**
 * HOC to wrap a component with an error boundary
 * 
 * @example
 * const SafeChart = withErrorBoundary(EmissionChart, {
 *   fallback: ChartErrorFallback,
 *   componentName: 'EmissionChart'
 * });
 * 
 * @param WrappedComponent - Component to wrap
 * @param options - Error boundary options
 * @returns Wrapped component with error boundary
 */
export function withErrorBoundary<P extends object>(
  WrappedComponent: React.ComponentType<P>,
  options: Omit<ErrorBoundaryProps, 'children'> = {}
) {
  const displayName = WrappedComponent.displayName || WrappedComponent.name || 'Component';

  const ComponentWithErrorBoundary = (props: P) => (
    <ErrorBoundary {...options} componentName={options.componentName || displayName}>
      <WrappedComponent {...props} />
    </ErrorBoundary>
  );

  ComponentWithErrorBoundary.displayName = `withErrorBoundary(${displayName})`;

  return ComponentWithErrorBoundary;
}

export default ErrorBoundary;
