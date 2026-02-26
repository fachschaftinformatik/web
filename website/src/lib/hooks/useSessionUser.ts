import { useMemo } from 'react';
import { useAuth } from '@lib/auth';

export const useSessionUser = () => {
  const { user, isLoading, login, logout } = useAuth();

  return useMemo(() => ({
    user,
    isLoading,
    isAuthenticated: Boolean(user),
    isAdmin: user?.role === 'admin',
    isEditor: user?.role === 'editor',
    isStaff: user?.role === 'admin' || user?.role === 'editor',
    login,
    logout,
  }), [isLoading, login, logout, user]);
};
