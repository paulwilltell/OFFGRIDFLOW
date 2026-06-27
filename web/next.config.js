const apiOrigin = process.env.OFFGRIDFLOW_API_ORIGIN || 'https://offgridflow-api-v2-production.up.railway.app';

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  experimental: {
    turbo: undefined,
  },
  compiler: {
    removeConsole: process.env.NODE_ENV === 'production',
  },
  env: {
    NEXT_PUBLIC_APP_VERSION: process.env.npm_package_version || '0.1.0',
  },
  eslint: {
    ignoreDuringBuilds: true,
  },
  typescript: {
    ignoreBuildErrors: true,
  },

  // Security headers
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          {
            key: 'X-Frame-Options',
            value: 'DENY',
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
          {
            key: 'Referrer-Policy',
            value: 'strict-origin-when-cross-origin',
          },
          {
            key: 'X-DNS-Prefetch-Control',
            value: 'on',
          },
          {
            key: 'Strict-Transport-Security',
            value: 'max-age=63072000; includeSubDomains; preload',
          },
          {
            key: 'Permissions-Policy',
            value: 'camera=(), microphone=(), geolocation=()',
          },
          {
            key: 'Content-Security-Policy',
            value: [
              "default-src 'self'",
              "script-src 'self' 'unsafe-inline' https://www.googletagmanager.com https://googleads.g.doubleclick.net https://www.googleadservices.com https://googleadservices.com",
              "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
              "font-src 'self' https://fonts.gstatic.com",
              "img-src 'self' data: blob: https: https://googleads.g.doubleclick.net https://www.google.com https://www.googleadservices.com",
              `connect-src 'self' ${apiOrigin} https://*.stripe.com https://googleads.g.doubleclick.net https://pagead2.googlesyndication.com https://www.google.com https://google.com https://www.googleadservices.com https://googleadservices.com`,
              "frame-src 'self' https://*.stripe.com",
              "object-src 'none'",
              "base-uri 'self'",
              "form-action 'self'",
            ].join('; '),
          },
        ],
      },
    ];
  },

  async rewrites() {
    return {
      fallback: [
        {
          source: '/api/:path*',
          destination: `${apiOrigin}/api/:path*`,
        },
      ],
    };
  },
};

module.exports = nextConfig;
