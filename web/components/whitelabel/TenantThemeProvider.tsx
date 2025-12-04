'use client';

import { ReactNode, useEffect, useState } from 'react';
import { api, ApiRequestError } from '../../lib/api';

interface TenantBranding {
  primaryColor: string;
  secondaryColor: string;
  logoUrl?: string;
  customCss?: string;
  fontFamily?: string;
}

interface TenantThemeProviderProps {
  children: ReactNode;
}

const DEFAULT_BRANDING: TenantBranding = {
  primaryColor: '#8aa9ff',
  secondaryColor: '#1d2940',
};

const CACHE_KEY = 'offgridflow_tenant_branding';
const CACHE_DURATION = 1000 * 60 * 60; // 1 hour

export default function TenantThemeProvider({ children }: TenantThemeProviderProps) {
  const [branding, setBranding] = useState<TenantBranding>(DEFAULT_BRANDING);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    const loadBranding = async () => {
      // Check localStorage cache first
      const cached = localStorage.getItem(CACHE_KEY);
      if (cached) {
        try {
          const { data, timestamp } = JSON.parse(cached);
          if (Date.now() - timestamp < CACHE_DURATION) {
            setBranding(data);
            applyTheme(data);
            setLoaded(true);
            return;
          }
        } catch (e) {
          console.warn('Failed to parse cached branding', e);
        }
      }

      // Fetch from API
      try {
        const data = await api.get<TenantBranding>('/api/tenant/branding');
        setBranding(data);
        applyTheme(data);
        
        // Cache the result
        localStorage.setItem(CACHE_KEY, JSON.stringify({
          data,
          timestamp: Date.now(),
        }));
      } catch (err) {
        if (err instanceof ApiRequestError) {
          console.warn('Branding API not available, using defaults');
        }
        // Use defaults
        applyTheme(DEFAULT_BRANDING);
      } finally {
        setLoaded(true);
      }
    };

    loadBranding();
  }, []);

  const applyTheme = (branding: TenantBranding) => {
    // Apply CSS variables
    const root = document.documentElement;
    root.style.setProperty('--primary-color', branding.primaryColor);
    root.style.setProperty('--secondary-color', branding.secondaryColor);
    
    if (branding.fontFamily) {
      root.style.setProperty('--font-family', branding.fontFamily);
    }

    // Inject custom CSS
    if (branding.customCss) {
      let styleElement = document.getElementById('tenant-custom-styles');
      if (!styleElement) {
        styleElement = document.createElement('style');
        styleElement.id = 'tenant-custom-styles';
        document.head.appendChild(styleElement);
      }
      styleElement.textContent = branding.customCss;
    }
  };

  if (!loaded) {
    return (
      <div style={{ padding: '2rem', textAlign: 'center' }}>
        <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>🎨</div>
        <div style={{ color: '#888' }}>Loading theme...</div>
      </div>
    );
  }

  return (
    <>
      {branding.logoUrl && (
        <style jsx global>{`
          .tenant-logo {
            background-image: url('${branding.logoUrl}');
            background-size: contain;
            background-repeat: no-repeat;
            background-position: center;
          }
        `}</style>
      )}
      {children}
    </>
  );
}

export function useTenantBranding(): TenantBranding {
  const [branding, setBranding] = useState<TenantBranding>(DEFAULT_BRANDING);

  useEffect(() => {
    const cached = localStorage.getItem(CACHE_KEY);
    if (cached) {
      try {
        const { data } = JSON.parse(cached);
        setBranding(data);
      } catch (e) {
        // Use defaults
      }
    }
  }, []);

  return branding;
}
