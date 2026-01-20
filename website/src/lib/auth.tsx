import React, { createContext, useState, useContext, useEffect, ReactNode, useCallback } from 'react';
import { Navigate, useLocation, Outlet } from 'react-router-dom';

import { getAuthMe, postAuthLogout } from '@lib/api';
import type { AuthUserResponse as User } from '@lib/api';
import { useThemeMode, type ThemePreference } from '@lib/theme';

const REMEMBERED_USER_KEY = 'fs_remembered_user';
export const REMEMBERED_FLAG_KEY = 'fs_remember_flag';

const readRememberedUser = (): User | null => {
  if (typeof window === 'undefined') {
    return null;
  }
  try {
    const stored = window.localStorage.getItem(REMEMBERED_USER_KEY);
    if (!stored) {
      return null;
    }
    return JSON.parse(stored) as User;
  } catch {
    window.localStorage.removeItem(REMEMBERED_USER_KEY);
    window.localStorage.removeItem(REMEMBERED_FLAG_KEY);
    return null;
  }
};

const persistRememberedUser = (userToPersist: User) => {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.setItem(REMEMBERED_USER_KEY, JSON.stringify(userToPersist));
  window.localStorage.setItem(REMEMBERED_FLAG_KEY, 'true');
};

const clearRememberedUser = () => {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.removeItem(REMEMBERED_USER_KEY);
  window.localStorage.removeItem(REMEMBERED_FLAG_KEY);
};

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  login: (user: User, rememberMe?: boolean) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(() => readRememberedUser());
  const [isLoading, setIsLoading] = useState(true);
  const { setPreference } = useThemeMode();

  const checkAuthStatus = useCallback(async () => {
    setIsLoading(true);
    try {
      const { data, error } = await getAuthMe();
      if (error || !data) {
        throw new Error('Nicht authentifiziert');
      }
      setUser(data);
      if (data.theme) {
        setPreference(data.theme as ThemePreference);
      }
      if (typeof window !== 'undefined' && window.localStorage.getItem(REMEMBERED_FLAG_KEY) === 'true') {
        persistRememberedUser(data);
      }
    } catch {
      setUser(null);
      clearRememberedUser();
    } finally {
      setIsLoading(false);
    }
  }, [setPreference]);

  useEffect(() => {
    checkAuthStatus();
  }, [checkAuthStatus]);

  useEffect(() => {
    if (user && user.theme) {
      setPreference(user.theme as ThemePreference);
    }
  }, [user, setPreference]);

  const login = (loggedInUser: User, rememberMe?: boolean) => {
    setUser(loggedInUser);
    if (loggedInUser.theme) {
      setPreference(loggedInUser.theme as ThemePreference);
    }
    if (rememberMe) {
      persistRememberedUser(loggedInUser);
    } else {
      clearRememberedUser();
    }
  };

  const logout = async () => {
    try {
      await postAuthLogout({});
      setUser(null);
      clearRememberedUser();
    } catch (error) {
      console.error('Abmeldung fehlgeschlagen:', error);
      setUser(null);
      clearRememberedUser();
    }
  };

  return (
    <AuthContext.Provider value={{ user, isLoading, login, logout }}>
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

export const ProtectedRoute: React.FC<{ children?: React.ReactNode }> = ({ children }) => {
  const { user, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) return null;

  if (!user) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return <>{children || <Outlet />}</>;
};
