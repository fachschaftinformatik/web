import React, { createContext, useContext, useEffect, ReactNode, useMemo } from 'react';
import { useSession, signOut } from 'next-auth/react';
import type { SessionUser } from '@lib/types/session';
import { toThemePreference, toSessionUser } from '@lib/types/guards';
import { useThemeMode } from '@lib/theme';

interface AuthContextType {
  user: SessionUser | null;
  isLoading: boolean;
  login: (user: SessionUser, rememberMe?: boolean) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const { data: session, status } = useSession();
  const { setPreference } = useThemeMode();

  const user = useMemo(() => toSessionUser(session?.user), [session?.user]);
  const isLoading = status === 'loading';

  useEffect(() => {
    if (user && user.theme) {
      setPreference(toThemePreference(user.theme));
    }
  }, [user, setPreference]);

  const value = useMemo(() => ({
    user,
    isLoading,
    login: () => {
      // The login logic is now handled by the Login page calling signIn('credentials', ...)
      // This is a no-op to maintain compatibility with existing components
    },
    logout: () => {
      signOut({ callbackUrl: '/' });
    }
  }), [user, isLoading]);

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth muss innerhalb eines AuthProviders verwendet werden');
  }
  return context;
};

