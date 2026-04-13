/** @type {import('next').NextConfig} */
// Build cache buster: env vars must be baked at build time
const nextConfig = {
  reactStrictMode: true,
  // Disable Turbopack for builds — server-external-packages.jsonc missing in 15.5.x
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
};

module.exports = nextConfig;
