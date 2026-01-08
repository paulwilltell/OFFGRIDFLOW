/**
 * @fileoverview Performance optimization utilities for OffGridFlow
 * @description Provides utilities for code splitting, lazy loading, memoization,
 * debouncing, and performance monitoring.
 */

import { createElement, useCallback, useEffect, useRef, useState, useMemo } from 'react';
import dynamic from 'next/dynamic';
import type { ComponentType } from 'react';

// ============================================================================
// Lazy Loading Utilities
// ============================================================================

/**
 * Configuration options for lazy-loaded components
 */
interface LazyLoadOptions {
  /** Loading component to show while loading */
  loading?: ComponentType;
  /** Whether to disable server-side rendering */
  ssr?: boolean;
  /** Delay before showing loading state (ms) */
  delay?: number;
}

/**
 * Create a lazy-loaded component with loading state
 * 
 * @param importFn - Dynamic import function
 * @param options - Lazy load options
 * @returns Lazy-loaded component
 * 
 * @example
 * const LazyChart = lazyLoad(() => import('@/components/charts/EmissionChartJS'), {
 *   loading: ChartSkeleton,
 *   ssr: false,
 * });
 */
export function lazyLoad<P extends object>(
  importFn: () => Promise<{ default: ComponentType<P> }>,
  options: LazyLoadOptions = {}
): ComponentType<P> {
  const { loading: Loading, ssr = true, delay = 200 } = options;

  return dynamic(importFn, {
    loading: Loading
      ? () => {
          const [showLoading, setShowLoading] = useState(false);
          
          useEffect(() => {
            const timer = setTimeout(() => setShowLoading(true), delay);
            return () => clearTimeout(timer);
          }, []);

          return showLoading ? createElement(Loading) : null;
        }
      : undefined,
    ssr,
  }) as ComponentType<P>;
}

// ============================================================================
// Debounce & Throttle
// ============================================================================

/**
 * Debounce a function call
 * 
 * @param fn - Function to debounce
 * @param delay - Delay in milliseconds
 * @returns Debounced function
 * 
 * @example
 * const debouncedSearch = useDebounce((query: string) => {
 *   searchEmissions(query);
 * }, 300);
 */
export function useDebounce<T extends (...args: any[]) => any>(
  fn: T,
  delay: number
): T {
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  const debouncedFn = useCallback(
    (...args: Parameters<T>) => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
      timeoutRef.current = setTimeout(() => fn(...args), delay);
    },
    [fn, delay]
  ) as T;

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  return debouncedFn;
}

/**
 * Debounce a value change
 * 
 * @param value - Value to debounce
 * @param delay - Delay in milliseconds
 * @returns Debounced value
 * 
 * @example
 * const [search, setSearch] = useState('');
 * const debouncedSearch = useDebouncedValue(search, 300);
 * 
 * useEffect(() => {
 *   if (debouncedSearch) {
 *     fetchResults(debouncedSearch);
 *   }
 * }, [debouncedSearch]);
 */
export function useDebouncedValue<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedValue(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);

  return debouncedValue;
}

/**
 * Throttle a function call
 * 
 * @param fn - Function to throttle
 * @param limit - Minimum time between calls (ms)
 * @returns Throttled function
 * 
 * @example
 * const throttledScroll = useThrottle((e: ScrollEvent) => {
 *   handleScroll(e);
 * }, 100);
 */
export function useThrottle<T extends (...args: any[]) => any>(
  fn: T,
  limit: number
): T {
  const lastRunRef = useRef<number>(0);
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  const throttledFn = useCallback(
    (...args: Parameters<T>) => {
      const now = Date.now();
      const remaining = limit - (now - lastRunRef.current);

      if (remaining <= 0) {
        lastRunRef.current = now;
        fn(...args);
      } else if (!timeoutRef.current) {
        timeoutRef.current = setTimeout(() => {
          lastRunRef.current = Date.now();
          timeoutRef.current = null;
          fn(...args);
        }, remaining);
      }
    },
    [fn, limit]
  ) as T;

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  return throttledFn;
}

// ============================================================================
// Memoization
// ============================================================================

/**
 * Deep comparison memoization hook
 * 
 * @param value - Value to memoize
 * @returns Memoized value that only changes on deep equality change
 * 
 * @example
 * const memoizedData = useDeepMemo(complexObject);
 */
export function useDeepMemo<T>(value: T): T {
  const ref = useRef<T>(value);

  if (!deepEqual(ref.current, value)) {
    ref.current = value;
  }

  return ref.current;
}

/**
 * Deep equality check for objects/arrays
 */
function deepEqual(a: unknown, b: unknown): boolean {
  if (a === b) return true;
  if (a == null || b == null) return false;
  if (typeof a !== typeof b) return false;

  if (Array.isArray(a) && Array.isArray(b)) {
    if (a.length !== b.length) return false;
    return a.every((item, i) => deepEqual(item, b[i]));
  }

  if (typeof a === 'object' && typeof b === 'object') {
    const keysA = Object.keys(a as object);
    const keysB = Object.keys(b as object);
    if (keysA.length !== keysB.length) return false;
    return keysA.every((key) =>
      deepEqual((a as any)[key], (b as any)[key])
    );
  }

  return false;
}

// ============================================================================
// Intersection Observer (Lazy Rendering)
// ============================================================================

/**
 * Hook for intersection observer (visibility detection)
 * 
 * @param options - Intersection observer options
 * @returns [ref, isVisible] tuple
 * 
 * @example
 * const [ref, isVisible] = useIntersectionObserver({ threshold: 0.1 });
 * 
 * return (
 *   <div ref={ref}>
 *     {isVisible && <ExpensiveChart />}
 *   </div>
 * );
 */
export function useIntersectionObserver(
  options: IntersectionObserverInit = {}
): [React.RefCallback<Element>, boolean] {
  const [isVisible, setIsVisible] = useState(false);
  const observerRef = useRef<IntersectionObserver | null>(null);

  const ref = useCallback(
    (node: Element | null) => {
      if (observerRef.current) {
        observerRef.current.disconnect();
      }

      if (node) {
        observerRef.current = new IntersectionObserver(([entry]) => {
          setIsVisible(entry.isIntersecting);
        }, options);
        observerRef.current.observe(node);
      }
    },
    [options.threshold, options.root, options.rootMargin]
  );

  useEffect(() => {
    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect();
      }
    };
  }, []);

  return [ref, isVisible];
}

// ============================================================================
// Performance Monitoring
// ============================================================================

/**
 * Hook to measure component render time
 * 
 * @param componentName - Name of the component for logging
 * 
 * @example
 * function MyComponent() {
 *   useRenderTimer('MyComponent');
 *   return <div>...</div>;
 * }
 */
export function useRenderTimer(componentName: string): void {
  const renderCount = useRef(0);
  const startTime = useRef(performance.now());

  useEffect(() => {
    renderCount.current += 1;
    const renderTime = performance.now() - startTime.current;

    if (process.env.NODE_ENV === 'development') {
      console.debug(
        `[Performance] ${componentName} rendered (#${renderCount.current}): ${renderTime.toFixed(2)}ms`
      );
    }

    startTime.current = performance.now();
  });
}

/**
 * Measure async operation performance
 * 
 * @param name - Operation name
 * @param fn - Async function to measure
 * @returns Result of the function
 * 
 * @example
 * const data = await measureAsync('fetchEmissions', async () => {
 *   return await api.getEmissions();
 * });
 */
export async function measureAsync<T>(
  name: string,
  fn: () => Promise<T>
): Promise<T> {
  const start = performance.now();
  
  try {
    const result = await fn();
    const duration = performance.now() - start;
    
    if (process.env.NODE_ENV === 'development') {
      console.debug(`[Performance] ${name}: ${duration.toFixed(2)}ms`);
    }
    
    // Report to analytics if configured
    if (typeof window !== 'undefined' && (window as any).gtag) {
      (window as any).gtag('event', 'timing_complete', {
        name,
        value: Math.round(duration),
        event_category: 'Performance',
      });
    }
    
    return result;
  } catch (error) {
    const duration = performance.now() - start;
    console.error(`[Performance] ${name} failed after ${duration.toFixed(2)}ms`);
    throw error;
  }
}

// ============================================================================
// Request Animation Frame
// ============================================================================

/**
 * Hook for requestAnimationFrame-based updates
 * 
 * @param callback - Function to call on each frame
 * @param deps - Dependencies array
 * 
 * @example
 * useAnimationFrame((deltaTime) => {
 *   updateAnimation(deltaTime);
 * });
 */
export function useAnimationFrame(
  callback: (deltaTime: number) => void,
  deps: React.DependencyList = []
): void {
  const requestRef = useRef<number>();
  const previousTimeRef = useRef<number>();

  const animate = useCallback(
    (time: number) => {
      if (previousTimeRef.current !== undefined) {
        const deltaTime = time - previousTimeRef.current;
        callback(deltaTime);
      }
      previousTimeRef.current = time;
      requestRef.current = requestAnimationFrame(animate);
    },
    [callback]
  );

  useEffect(() => {
    requestRef.current = requestAnimationFrame(animate);
    return () => {
      if (requestRef.current) {
        cancelAnimationFrame(requestRef.current);
      }
    };
  }, deps);
}

// ============================================================================
// Virtual List Support
// ============================================================================

/**
 * Calculate visible items for virtual scrolling
 * 
 * @param totalItems - Total number of items
 * @param itemHeight - Height of each item
 * @param containerHeight - Height of the container
 * @param scrollTop - Current scroll position
 * @param overscan - Number of items to render outside viewport
 * @returns Start index, end index, and offset
 * 
 * @example
 * const { startIndex, endIndex, offsetY } = getVisibleRange(
 *   items.length,
 *   50,
 *   containerRef.current?.clientHeight || 0,
 *   scrollTop,
 *   5
 * );
 */
export function getVisibleRange(
  totalItems: number,
  itemHeight: number,
  containerHeight: number,
  scrollTop: number,
  overscan: number = 3
): { startIndex: number; endIndex: number; offsetY: number } {
  const startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - overscan);
  const visibleCount = Math.ceil(containerHeight / itemHeight);
  const endIndex = Math.min(
    totalItems - 1,
    startIndex + visibleCount + overscan * 2
  );
  const offsetY = startIndex * itemHeight;

  return { startIndex, endIndex, offsetY };
}

export default {
  lazyLoad,
  useDebounce,
  useDebouncedValue,
  useThrottle,
  useDeepMemo,
  useIntersectionObserver,
  useRenderTimer,
  measureAsync,
  useAnimationFrame,
  getVisibleRange,
};
