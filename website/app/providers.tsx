'use client';

import * as React from 'react';
import { AppRouterCacheProvider } from '@mui/material-nextjs/v16-appRouter';
import { ThemeModeProvider } from '@lib/theme';
import { AuthProvider } from '@lib/auth';
import { SessionProvider } from 'next-auth/react';
import { client } from '@lib/api/client.gen';

// Initialize API client
if (typeof window !== 'undefined') {
  client.setConfig({
    baseUrl: '/api/v1',
    credentials: 'include',
  });
}

export default function Providers({ children }: { children: React.ReactNode }) {
  return (
    <AppRouterCacheProvider>
      <SessionProvider basePath="/api/auth">
        <ThemeModeProvider>
          <AuthProvider>
            {children}
          </AuthProvider>
        </ThemeModeProvider>
      </SessionProvider>
    </AppRouterCacheProvider>
  );
}

