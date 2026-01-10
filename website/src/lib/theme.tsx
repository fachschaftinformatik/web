import React from 'react';
import { Fab, Tooltip } from '@mui/material';
import CssBaseline from '@mui/material/CssBaseline';
import { PaletteMode, PaletteOptions, ThemeOptions, ThemeProvider, createTheme, useTheme } from '@mui/material/styles';
import Brightness4Rounded from '@mui/icons-material/Brightness4Rounded';
import Brightness7Rounded from '@mui/icons-material/Brightness7Rounded';

// Light mode tokens (general UI palette).
const lightTokens = {
  surface: {
    app: '#f5f5f5', // page background
    section: '#fafafa', // soft section background
    card: '#ffffff', // cards, sheets, sidebar
  },
  text: {
    primary: '#1f1f1f',
    muted: '#5c5c5c',
    inverse: '#ffffff',
  },
  border: {
    subtle: '#1f1f1f1f', // 12% opacity
  },
  brand: {
    primary: {
      main: '#2f7957',
      light: '#3da474',
      dark: '#226244',
      contrastText: '#ffffff',
    },
    secondary: {
      main: '#262d36',
      light: '#404a59',
      dark: '#181e25',
      contrastText: '#ffffff',
    },
  },
  status: {
    success: {
      main: '#2f7d4a',
      light: '#38a869',
      dark: '#21633e',
      contrastText: '#ffffff',
    },
    warning: {
      main: '#d39a3f',
      light: '#e4b062',
      dark: '#b88333',
      contrastText: '#2b1d06',
    },
    info: {
      main: '#3267d8',
      light: '#4b7fe0',
      dark: '#1f4fb8',
      contrastText: '#ffffff',
    },
    error: {
      main: '#e16d48',
      light: '#ee896a',
      dark: '#b95637',
      contrastText: '#2b0c06',
    },
  },
};

// Dark mode tokens tuned to match the current UI.
const darkTokens = {
  surface: {
    app: '#0c1411',
    card: '#0f1017',
  },
  text: {
    primary: '#f6f9f6',
    muted: '#9aaea3',
  },
  border: {
    subtle: 'rgba(255, 255, 255, 0.08)',
  },
  brand: {
    primary: {
      main: '#3ee29b',
      light: '#6df2bf',
      dark: '#24b87b',
      contrastText: '#0b1511',
    },
    secondary: {
      main: '#0f172a',
      light: '#475569',
      dark: '#000000',
      contrastText: '#ffffff',
    },
  },
  status: {
    success: {
      main: '#2f7d4a',
      light: '#38a869',
      dark: '#21633e',
      contrastText: '#ffffff',
    },
    warning: {
      main: '#ffb454',
      light: '#ffc777',
      dark: '#d39a3f',
      contrastText: '#2b1d06',
    },
    info: {
      main: '#7cc7ff',
      light: '#a7dcff',
      dark: '#4aa6ea',
      contrastText: '#031424',
    },
    error: {
      main: '#e16d48',
      light: '#f08b69',
      dark: '#b95637',
      contrastText: '#2b0c06',
    },
  },
};

const appBarGreen = lightTokens.brand.primary.main;

const lightPalette: PaletteOptions = {
  mode: 'light',
  primary: lightTokens.brand.primary,
  secondary: lightTokens.brand.secondary,
  background: {
    default: lightTokens.surface.app,
    paper: lightTokens.surface.card,
  },
  text: {
    primary: lightTokens.text.primary,
    secondary: lightTokens.text.muted,
  },
  divider: lightTokens.border.subtle,
  success: lightTokens.status.success,
  warning: lightTokens.status.warning,
  info: lightTokens.status.info,
  error: lightTokens.status.error,
};

const darkPalette: PaletteOptions = {
  mode: 'dark',
  primary: darkTokens.brand.primary,
  secondary: darkTokens.brand.secondary,
  background: {
    default: darkTokens.surface.app,
    paper: darkTokens.surface.card,
  },
  text: {
    primary: darkTokens.text.primary,
    secondary: darkTokens.text.muted,
  },
  divider: darkTokens.border.subtle,
  success: darkTokens.status.success,
  warning: darkTokens.status.warning,
  info: darkTokens.status.info,
  error: darkTokens.status.error,
};

const getPalette = (mode: PaletteMode): PaletteOptions => (mode === 'dark' ? darkPalette : lightPalette);

// Typography stays global to keep visual consistency across pages.
const typography: ThemeOptions['typography'] = {
  fontFamily: 'Roboto, Arial, sans-serif',
  h5: {
    fontWeight: 700,
  },
};

// Shared component defaults.
const components: ThemeOptions['components'] = {
  MuiAppBar: {
    styleOverrides: {
      root: ({ theme }) => ({
        backgroundColor: appBarGreen,
        color: theme.palette.getContrastText(appBarGreen),
        backgroundImage: 'none',
        boxShadow: theme.shadows[4],
      }),
    },
  },
  MuiButton: {
    styleOverrides: {
      root: {
        borderRadius: 8,
        textTransform: 'none',
      },
    },
  },
};

// Theme factory: keeps palette, typography, and component overrides in one place.
export const createAppTheme = (mode: PaletteMode = 'light') =>
  createTheme({
    palette: getPalette(mode),
    typography,
    components,
  });

const theme = createAppTheme('light');

export default theme;

// Theme mode state + persistence for the UI toggle.
type ThemeModeContextValue = {
  mode: PaletteMode;
  toggleMode: () => void;
  setMode: (mode: PaletteMode) => void;
};

const ThemeModeContext = React.createContext<ThemeModeContextValue | undefined>(undefined);

const STORAGE_KEY = 'app-theme-mode';

const getStoredMode = (): PaletteMode => {
  if (typeof window === 'undefined') {
    return 'light';
  }
  const stored = window.localStorage.getItem(STORAGE_KEY);
  return stored === 'dark' ? 'dark' : 'light';
};

export const ThemeModeProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [mode, setModeState] = React.useState<PaletteMode>(() => getStoredMode());

  const setMode = React.useCallback((nextMode: PaletteMode) => {
    setModeState(nextMode);
  }, []);

  const toggleMode = React.useCallback(() => {
    setModeState((prev) => (prev === 'light' ? 'dark' : 'light'));
  }, []);

  React.useEffect(() => {
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(STORAGE_KEY, mode);
    }
  }, [mode]);

  const theme = React.useMemo(() => createAppTheme(mode), [mode]);

  return (
    <ThemeModeContext.Provider value={{ mode, toggleMode, setMode }}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </ThemeProvider>
    </ThemeModeContext.Provider>
  );
};

export const useThemeMode = () => {
  const context = React.useContext(ThemeModeContext);
  if (!context) {
    throw new Error('useThemeMode must be used within a ThemeModeProvider');
  }
  return context;
};

export const ThemeModeToggle: React.FC = () => {
  const { mode, toggleMode } = useThemeMode();
  const theme = useTheme();

  return (
    <Tooltip title={mode === 'light' ? 'Dark Mode aktivieren' : 'Light Mode aktivieren'}>
      <Fab
        color="primary"
        size="medium"
        onClick={toggleMode}
        sx={{
          position: 'fixed',
          bottom: { xs: 16, sm: 24 },
          right: { xs: 16, sm: 24 },
          zIndex: theme.zIndex.tooltip,
        }}
        aria-label="toggle color mode"
      >
        {mode === 'light' ? <Brightness4Rounded /> : <Brightness7Rounded />}
      </Fab>
    </Tooltip>
  );
};
