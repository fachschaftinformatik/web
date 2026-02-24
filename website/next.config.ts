import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'standalone',
  typescript: {
    // This disables the default Next.js type check during builds
    ignoreBuildErrors: true,
  },
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          { key: 'X-Frame-Options', value: 'SAMEORIGIN' },
          { key: 'X-Content-Type-Options', value: 'nosniff' },
          { key: 'X-XSS-Protection', value: '1; mode=block' },
          { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
          {
            key: 'Content-Security-Policy',
            value: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; font-src 'self' data:; img-src 'self' data: blob:; frame-src 'self' blob:; worker-src 'self' blob:; object-src 'none';",
          },
          { key: 'Strict-Transport-Security', value: 'max-age=31536000; includeSubDomains' },
        ],
      },
    ];
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*',
      },
    ];
  },
  async redirects() {
    return [
      {
        source: '/auth',
        destination: '/login',
        permanent: true,
      },
      {
        source: '/d',
        destination: '/discussions',
        permanent: true,
      },
      {
        source: '/forum/:path*',
        destination: '/discussions/:path*',
        permanent: true,
      },
      {
        source: '/exams/:path*',
        destination: '/archive/:path*',
        permanent: true,
      },
      {
        source: '/images',
        destination: '/events',
        permanent: true,
      },
      {
        source: '/media',
        destination: '/events',
        permanent: true,
      },
    ];
  },
  // Ensure we support the same aliases as Vite
  webpack: (config) => {
    config.resolve.alias = {
      ...config.resolve.alias,
      '@lib': './src/lib',
      '@components': './src/components',
      '@routes': './src/routes',
    };
    return config;
  },
  turbopack: {},
};

export default nextConfig;
