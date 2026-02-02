import React from 'react';
import { PaletteMode, ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { getPalette } from './variables';
import { typography, components, getCustomStyles } from './customizations';

export const createAppTheme = (mode: PaletteMode = 'light') => {
  const theme = createTheme({
    palette: getPalette(mode),
    typography,
    spacing: 8,
    shape: { borderRadius: 8 },
    components,
  });

  return {
    ...theme,
    custom: getCustomStyles(theme),
  };
};

export type ThemePreference = 'light' | 'dark' | 'system';

type ThemeModeContextValue = {
  preference: ThemePreference;
  mode: PaletteMode;
  toggleMode: () => void;
  setPreference: (pref: ThemePreference) => void;
};

const ThemeModeContext = React.createContext<ThemeModeContextValue | undefined>(undefined);

const STORAGE_KEY = 'app-theme-mode';

const getStoredPreference = (): ThemePreference => {
  if (typeof window === 'undefined') return 'system';
  const stored = window.localStorage.getItem(STORAGE_KEY);
  if (stored === 'dark' || stored === 'light' || stored === 'system') return stored;
  return 'system';
};

export const ThemeModeProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [preference, setPreferenceState] = React.useState<ThemePreference>(() => getStoredPreference());
  const [resolvedMode, setResolvedMode] = React.useState<PaletteMode>('light');

  const setPreference = React.useCallback((nextPref: ThemePreference) => {
    setPreferenceState(nextPref);
  }, []);

  const toggleMode = React.useCallback(() => {
    setPreferenceState((prev) => (prev === 'light' ? 'dark' : 'light'));
  }, []);

  React.useEffect(() => {
    if (preference === 'system') {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
      const handleChange = () => setResolvedMode(mediaQuery.matches ? 'dark' : 'light');
      setResolvedMode(mediaQuery.matches ? 'dark' : 'light');
      mediaQuery.addEventListener('change', handleChange);
      return () => mediaQuery.removeEventListener('change', handleChange);
    } else setResolvedMode(preference);
  }, [preference]);

  React.useEffect(() => {
    if (typeof window !== 'undefined') window.localStorage.setItem(STORAGE_KEY, preference);
  }, [preference]);

  const theme = React.useMemo(() => createAppTheme(resolvedMode), [resolvedMode]);

  return (
    <ThemeModeContext.Provider value={{ preference, mode: resolvedMode, toggleMode, setPreference }}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </ThemeProvider>
    </ThemeModeContext.Provider>
  );
};

export const useThemeMode = () => {
  const context = React.useContext(ThemeModeContext);
  if (!context) throw new Error('useThemeMode must be used within ThemeModeProvider');
  return context;
};
