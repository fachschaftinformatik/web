import React, { createContext, useState, useContext, useEffect, ReactNode } from 'react';
import * as api from './api';

const REMEMBERED_USER_KEY = 'fs_remembered_user';
export const REMEMBERED_FLAG_KEY = 'fs_remember_flag';

const readRememberedUser = (): api.User | null => {
  if (typeof window === 'undefined') {
    return null;
  }
  try {
    const stored = window.localStorage.getItem(REMEMBERED_USER_KEY);
    if (!stored) {
      return null;
    }
    return JSON.parse(stored) as api.User;
  } catch {
    window.localStorage.removeItem(REMEMBERED_USER_KEY);
    window.localStorage.removeItem(REMEMBERED_FLAG_KEY);
    return null;
  }
};

const persistRememberedUser = (userToPersist: api.User) => {
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
  user: api.User | null;
  isLoading: boolean;
  login: (user: api.User, rememberMe?: boolean) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<api.User | null>(() => readRememberedUser());
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const checkAuthStatus = async () => {
      setIsLoading(true);
      try {
        const currentUser = await api.getMe();
        setUser(currentUser);
        if (typeof window !== 'undefined' && window.localStorage.getItem(REMEMBERED_FLAG_KEY) === 'true') {
          persistRememberedUser(currentUser);
        }
      } catch {
        setUser(null);
        clearRememberedUser();
      } finally {
        setIsLoading(false);
      }
    };
    checkAuthStatus();
  }, []);

  const login = (loggedInUser: api.User, rememberMe?: boolean) => {
    setUser(loggedInUser);
    if (rememberMe) {
      persistRememberedUser(loggedInUser);
    } else {
      clearRememberedUser();
    }
  };

  const logout = async () => {
    try {
      await api.logoutUser();
      setUser(null);
      clearRememberedUser();
    } catch (error) {
      console.error('Logout failed:', error);
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
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
