'use client';

import { useTenantBranding } from './TenantThemeProvider';

interface TenantLogoProps {
  width?: string;
  height?: string;
  fallback?: string;
}

export default function TenantLogo({ width = '200px', height = '60px', fallback = 'OffGridFlow' }: TenantLogoProps) {
  const branding = useTenantBranding();

  if (!branding.logoUrl) {
    return (
      <div
        style={{
          width,
          height,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: '1.5rem',
          fontWeight: 700,
          color: branding.primaryColor,
        }}
      >
        {fallback}
      </div>
    );
  }

  return (
    <div
      className="tenant-logo"
      style={{
        width,
        height,
      }}
      role="img"
      aria-label="Company Logo"
    />
  );
}
