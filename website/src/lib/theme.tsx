import React from 'react';
import { PaletteMode } from '@mui/material';
import { Fab, Tooltip } from '@mui/material';
import CssBaseline from '@mui/material/CssBaseline';
import { ThemeProvider } from '@mui/material/styles';
import { useTheme } from '@mui/material/styles';
import Brightness4Rounded from '@mui/icons-material/Brightness4Rounded';
import Brightness7Rounded from '@mui/icons-material/Brightness7Rounded';
import { PaletteMode } from '@mui/material';
import { PaletteOptions, ThemeOptions, createTheme } from '@mui/material/styles';

const getPalette = (mode: PaletteMode): PaletteOptions => {
  const shared = {
    primary: {
      main: '#046709',
      light: '#4cb54f',
      dark: '#024d06',
      contrastText: '#ffffff',
    },
    secondary: {
      main: '#0f172a',
      light: '#475569',
      dark: '#000000',
      contrastText: '#ffffff',
    },
  };

  if (mode === 'dark') {
    return {
      ...shared,
      mode,
      background: {
        default: '#05060a',
        paper: '#0f1017',
      },
      text: {
        primary: '#f8fafc',
        secondary: '#cbd5f5',
      },
      divider: 'rgba(255, 255, 255, 0.12)',
    };
  }

  return {
    ...shared,
    mode,
    background: {
      default: '#f4f7fb',
      paper: '#ffffff',
    },
    text: {
      primary: '#0f172a',
      secondary: '#475569',
    },
    divider: 'rgba(15,23,42,0.12)',
  };
};

const typography: ThemeOptions['typography'] = {
  fontFamily: 'Roboto, Arial, sans-serif',
  h5: {
    fontWeight: 700,
  },
};

const components: ThemeOptions['components'] = {
  MuiButton: {
    styleOverrides: {
      root: {
        borderRadius: 8,
        textTransform: 'none',
      },
    },
  },
};

export const createAppTheme = (mode: PaletteMode = 'light') =>
  createTheme({
    palette: getPalette(mode),
    typography,
    components,
  });

const theme = createAppTheme('light');

export default theme;


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
