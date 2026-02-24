import { Suspense } from 'react';
import AuthPage from '../AuthPage';
import { PageLoader } from '@components/layout';

export default function RegisterPage() {
  return (
    <Suspense fallback={<PageLoader />}>
      <AuthPage />
    </Suspense>
  );
}
