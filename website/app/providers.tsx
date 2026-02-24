'use client';

import * as React from 'react';
import { AppRouterCacheProvider } from '@mui/material-nextjs/v16-appRouter';
import { ThemeModeProvider } from '@lib/theme';
import { AuthProvider } from '@lib/auth';
import { client } from '@lib/api/client.gen';

// Initialize API client
if (typeof window !== 'undefined') {
  client.setConfig({
    baseUrl: '/api/v1',
    credentials: 'include',
  });

  // CSRF Interceptor moved from App.tsx
  client.interceptors.request.use(async (request) => {
    const method = request.method.toUpperCase();
    const stateChangingMethods = ['POST', 'PUT', 'DELETE', 'PATCH'];

    if (stateChangingMethods.includes(method)) {
      const { getCsrfFromCookie, fetchCsrfToken } = await import('@lib/csrf');
      let tokenValue = getCsrfFromCookie();
      
      if (!tokenValue && !request.url.includes('/auth/csrf')) {
        const data = await fetchCsrfToken();
        tokenValue = data?.csrf || null;
      }

      if (tokenValue) {
        request.headers.set('X-CSRF-Token', tokenValue);
      }
    }
    return request;
  });
}

export default function Providers({ children }: { children: React.ReactNode }) {
  return (
    <AppRouterCacheProvider>
      <ThemeModeProvider>
        <AuthProvider>
          {children}
        </AuthProvider>
      </ThemeModeProvider>
    </AppRouterCacheProvider>
  );
}

