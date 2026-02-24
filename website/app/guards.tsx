'use client';

import { useAuth } from '@lib/auth';
import { useRouter, usePathname } from 'next/navigation';
import { useEffect } from 'react';
import { PageLoader } from '@components/layout';

export function ProtectedGuard({ children }: { children: React.ReactNode }) {
  const { user, isLoading } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!isLoading && !user) {
      router.replace(`/login?from=${encodeURIComponent(pathname)}`);
    }
  }, [user, isLoading, router, pathname]);

  if (isLoading || !user) {
    return <PageLoader />;
  }

  return <>{children}</>;
}

export function AdminGuard({ children }: { children: React.ReactNode }) {
  const { user, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading) {
      if (!user) {
        router.replace('/login');
      } else if (user.role !== 'admin') {
        router.replace('/');
      }
    }
  }, [user, isLoading, router]);

  if (isLoading || !user || user.role !== 'admin') {
    return <PageLoader />;
  }

  return <>{children}</>;
}
