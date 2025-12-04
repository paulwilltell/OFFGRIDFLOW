'use client';

import { useEffect } from 'react';
import * as Sentry from '@sentry/nextjs';

export function PerformanceMonitoring() {
  useEffect(() => {
    // Web Vitals monitoring
    if (typeof window !== 'undefined' && 'performance' in window) {
      // Largest Contentful Paint (LCP)
      const observer = new PerformanceObserver((entryList) => {
        for (const entry of entryList.getEntries()) {
          const lcp = entry as PerformanceEntry;
          console.log('LCP:', lcp.startTime);
          
          // Send to analytics
          if (window.gtag) {
            window.gtag('event', 'web_vitals', {
              event_category: 'Web Vitals',
              event_label: 'LCP',
              value: Math.round(lcp.startTime),
              non_interaction: true,
            });
          }
        }
      });

      try {
        observer.observe({ entryTypes: ['largest-contentful-paint'] });
      } catch (e) {
        console.error('Performance observer error:', e);
      }

      // First Input Delay (FID)
      const fidObserver = new PerformanceObserver((entryList) => {
        for (const entry of entryList.getEntries()) {
          const fid = entry as PerformanceEventTiming;
          const delay = fid.processingStart - fid.startTime;
          console.log('FID:', delay);

          if (window.gtag) {
            window.gtag('event', 'web_vitals', {
              event_category: 'Web Vitals',
              event_label: 'FID',
              value: Math.round(delay),
              non_interaction: true,
            });
          }
        }
      });

      try {
        fidObserver.observe({ entryTypes: ['first-input'] });
      } catch (e) {
        console.error('FID observer error:', e);
      }

      // Cleanup
      return () => {
        observer.disconnect();
        fidObserver.disconnect();
      };
    }
  }, []);

  return null;
}

// Custom hook for lazy loading components
export function useLazyLoad<T>(
  importFunc: () => Promise<{ default: T }>,
  delay: number = 0
): T | null {
  const [component, setComponent] = useState<T | null>(null);

  useEffect(() => {
    const timer = setTimeout(() => {
      importFunc().then((mod) => setComponent(mod.default));
    }, delay);

    return () => clearTimeout(timer);
  }, [importFunc, delay]);

  return component;
}

// Virtual scrolling hook for large lists
export function useVirtualScroll(
  itemCount: number,
  itemHeight: number,
  containerHeight: number
) {
  const [scrollTop, setScrollTop] = useState(0);

  const startIndex = Math.floor(scrollTop / itemHeight);
  const endIndex = Math.min(
    itemCount - 1,
    Math.ceil((scrollTop + containerHeight) / itemHeight)
  );

  const visibleItems = [];
  for (let i = startIndex; i <= endIndex; i++) {
    visibleItems.push(i);
  }

  return {
    visibleItems,
    totalHeight: itemCount * itemHeight,
    offsetY: startIndex * itemHeight,
    onScroll: (e: React.UIEvent<HTMLElement>) => {
      setScrollTop(e.currentTarget.scrollTop);
    },
  };
}

// Image optimization component
import { useState } from 'react';
import Image from 'next/image';

interface OptimizedImageProps {
  src: string;
  alt: string;
  width: number;
  height: number;
  priority?: boolean;
}

export function OptimizedImage({
  src,
  alt,
  width,
  height,
  priority = false,
}: OptimizedImageProps) {
  const [isLoaded, setIsLoaded] = useState(false);

  return (
    <div style={{ position: 'relative', width, height }}>
      <Image
        src={src}
        alt={alt}
        width={width}
        height={height}
        priority={priority}
        loading={priority ? 'eager' : 'lazy'}
        placeholder="blur"
        blurDataURL="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mN8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
        onLoadingComplete={() => setIsLoaded(true)}
        style={{
          opacity: isLoaded ? 1 : 0,
          transition: 'opacity 0.3s ease-in-out',
        }}
      />
    </div>
  );
}

declare global {
  interface Window {
    gtag: (...args: any[]) => void;
  }
}
