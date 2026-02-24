import React, { createContext, useContext, useEffect, ReactNode, useMemo } from 'react';
import { useSession, signIn, signOut } from 'next-auth/react';
import type { DtoUserResponse as User } from '@lib/api';
import { useThemeMode, type ThemePreference } from '@lib/theme';

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  login: (user: User, rememberMe?: boolean) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const { data: session, status } = useSession();
  const { setPreference } = useThemeMode();

  const user = (session?.user as unknown as User) || null;
  const isLoading = status === 'loading';

  useEffect(() => {
    if (user && user.theme) {
      setPreference(user.theme as ThemePreference);
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


