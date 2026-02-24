import { Suspense } from 'react';
import AuthPage from '../AuthPage';
import { PageLoader } from '@components/layout';

export default function LoginPage() {
  return (
    <Suspense fallback={<PageLoader />}>
      <AuthPage />
    </Suspense>
  );
}
