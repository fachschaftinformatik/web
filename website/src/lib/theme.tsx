import React from 'react';
import { Fab, Tooltip, alpha } from '@mui/material';
import CssBaseline from '@mui/material/CssBaseline';
import { PaletteMode, PaletteOptions, Theme, ThemeOptions, ThemeProvider, createTheme, useTheme } from '@mui/material/styles';

export type CustomThemeStyles = ReturnType<typeof getCustomStyles>;
import Brightness4Rounded from '@mui/icons-material/Brightness4Rounded';
import Brightness7Rounded from '@mui/icons-material/Brightness7Rounded';


const lightTokens = {
  surface: { app: '#f5f5f5', section: '#fafafa', card: '#ffffff' },
  text: { primary: '#1f1f1f', muted: '#5c5c5c', inverse: '#ffffff' },
  border: { subtle: '#1f1f1f1f' },
  brand: {
    primary: { main: '#2f7957', light: '#3da474', dark: '#226244', contrastText: '#ffffff' },
    secondary: { main: '#262d36', light: '#404a59', dark: '#181e25', contrastText: '#ffffff' },
  },
  status: {
    success: { main: '#2f7d4a', light: '#38a869', dark: '#21633e', contrastText: '#ffffff' },
    warning: { main: '#d39a3f', light: '#e4b062', dark: '#b88333', contrastText: '#2b1d06' },
    info: { main: '#3267d8', light: '#4b7fe0', dark: '#1f4fb8', contrastText: '#ffffff' },
    error: { main: '#e16d48', light: '#ee896a', dark: '#b95637', contrastText: '#2b0c06' },
  },
};

const darkTokens = {
  surface: { app: '#0c1411', card: '#0f1017' },
  text: { primary: '#f6f9f6', muted: '#9aaea3' },
  border: { subtle: 'rgba(255, 255, 255, 0.08)' },
  brand: {
    primary: { main: '#3ee29b', light: '#6df2bf', dark: '#24b87b', contrastText: '#0b1511' },
    secondary: { main: '#0f172a', light: '#475569', dark: '#000000', contrastText: '#ffffff' },
  },
  status: {
    success: { main: '#2f7d4a', light: '#38a869', dark: '#21633e', contrastText: '#ffffff' },
    warning: { main: '#ffb454', light: '#ffc777', dark: '#d39a3f', contrastText: '#2b1d06' },
    info: { main: '#7cc7ff', light: '#a7dcff', dark: '#4aa6ea', contrastText: '#031424' },
    error: { main: '#e16d48', light: '#f08b69', dark: '#b95637', contrastText: '#2b0c06' },
  },
};

const appBarGreen = lightTokens.brand.primary.main;

const getPalette = (mode: PaletteMode): PaletteOptions => {
  const tokens = mode === 'dark' ? darkTokens : lightTokens;
  return {
    mode,
    primary: tokens.brand.primary,
    secondary: tokens.brand.secondary,
    background: { default: tokens.surface.app, paper: tokens.surface.card },
    text: { primary: tokens.text.primary, secondary: tokens.text.muted },
    divider: tokens.border.subtle,
    success: tokens.status.success,
    warning: tokens.status.warning,
    info: tokens.status.info,
    error: tokens.status.error,
  };
};

const typography: ThemeOptions['typography'] = {
  fontFamily: '"Manrope", "Space Grotesk", sans-serif',
  h1: { fontFamily: '"Space Grotesk", sans-serif', fontWeight: 700, lineHeight: 1.05 },
  h4: { fontFamily: '"Space Grotesk", sans-serif', fontWeight: 700, letterSpacing: '-0.01em' },
  h5: { fontWeight: 700 },
  h6: { fontWeight: 700 },
  subtitle1: { fontWeight: 700, lineHeight: 1.2 },
  overline: { fontWeight: 700, letterSpacing: '0.26em' },
  caption: { fontWeight: 400 },
};

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
        fontWeight: 600,
      },
    },
  },
};


const getCustomStyles = (theme: Theme) => {
  const isDark = theme.palette.mode === 'dark';

  return {
    pageWrapper: {
      position: 'relative',
      borderRadius: 0,
      py: 4,
      overflow: 'hidden',
      color: 'var(--ink)',
      fontFamily: '"Manrope", "Space Grotesk", sans-serif',
      '--ink': theme.palette.text.primary,
      '--muted': theme.palette.text.secondary,
      '--accent': theme.palette.primary.light,
      '--accent-strong': theme.palette.primary.main,
      '--accent-soft': alpha(theme.palette.primary.main, isDark ? 0.16 : 0.12),
      '--accent-2': theme.palette.warning.main,
      '--accent-3': theme.palette.info.main,
      '--surface': theme.palette.background.paper,
      '--card-bg': alpha(theme.palette.background.paper, isDark ? 0.82 : 0.96),
      '--card-border': theme.palette.divider,
      background: 'var(--surface)',
      backgroundImage: 'none',
      '&::before': { content: 'none' },
      '&::after': { content: 'none' },
      '@keyframes rise': {
        from: { opacity: 0, transform: 'translateY(36px)' },
        to: { opacity: 1, transform: 'translateY(0)' },
      },
      '@keyframes sectionReveal': {
        from: { opacity: 0, transform: 'translateY(48px) scale(0.98)', filter: 'blur(8px)' },
        to: { opacity: 1, transform: 'translateY(0) scale(1)', filter: 'blur(0)' },
      },
      '@keyframes float': {
        from: { transform: 'translateY(0px)' },
        to: { transform: 'translateY(-12px)' },
      },
      '@keyframes eventFloat': {
        from: { transform: 'translateY(0px)' },
        to: { transform: 'translateY(-12px)' },
      },
      '@keyframes gradientShift': {
        '0%': { backgroundPosition: '0% 50%' },
        '100%': { backgroundPosition: '100% 50%' },
      },
    },

    gridContainer: {
      position: 'relative',
      zIndex: 1,
      width: '100%',
      maxWidth: '1440px',
      mx: 'auto',
      px: { xs: 2, md: 4 },
      display: 'grid',
      gridTemplateColumns: 'repeat(12, minmax(0, 1fr))',
      columnGap: 3,
      rowGap: 8,
    },

    sectionShell: {
      position: 'relative',
      borderRadius: { xs: 4, md: 6 },
      p: { xs: 2, md: 3 },
      overflow: 'visible',
      border: '1px solid transparent',
      backdropFilter: 'blur(16px)',
      boxShadow: isDark
        ? `0 26px 60px ${alpha(theme.palette.common.black, 0.35)}`
        : `0 26px 60px ${alpha(theme.palette.primary.dark, 0.12)}`,
      willChange: 'opacity, transform, filter',
      opacity: 0,
      transform: 'translateY(48px) scale(0.98)',
      filter: 'blur(8px)',
      '&[data-visible="true"]': {
        animation: 'sectionReveal 780ms cubic-bezier(0.22, 0.8, 0.2, 1) both',
      },
      '@media (prefers-reduced-motion: reduce)': {
        opacity: 1,
        transform: 'none',
        filter: 'none',
        animation: 'none',
      },
    },

    glassCard: {
      borderRadius: 4,
      border: '1px solid transparent',
      background: isDark
        ? alpha(theme.palette.background.paper, 0.72)
        : `linear-gradient(145deg, ${alpha(theme.palette.background.paper, 0.98)}, ${alpha(theme.palette.primary.main, 0.06)})`,
      boxShadow: isDark
        ? `0 18px 36px ${alpha(theme.palette.common.black, 0.35)}`
        : `0 18px 36px ${alpha(theme.palette.primary.dark, 0.12)}`,
      backdropFilter: 'blur(12px)',
      transition: 'transform 160ms ease, box-shadow 160ms ease, border-color 160ms ease',
      '&:hover': {
        transform: 'translateY(-3px)',
        borderColor: 'var(--accent-strong)',
        boxShadow: isDark
          ? `0 22px 45px ${alpha(theme.palette.common.black, 0.4)}`
          : `0 22px 45px ${alpha(theme.palette.primary.dark, 0.16)}`,
      },
    },

    sectionTitle: {
      fontFamily: '"Space Grotesk", sans-serif',
      fontWeight: 700,
      fontSize: { xs: '1.2rem', md: '1.45rem' },
      letterSpacing: '-0.01em',
    },
    sectionAction: {
      textTransform: 'none',
      color: 'var(--accent-strong)',
      fontWeight: 600,
      '&:hover': { bgcolor: 'var(--accent-soft)' },
    },
    sectionIcon: {
      color: 'var(--accent-strong)',
      fontSize: 22,
    },
    sectionGradients: {
      newsroom: isDark
        ? `linear-gradient(145deg, ${alpha(theme.palette.background.default, 0.96)}, ${alpha(theme.palette.primary.dark, 0.35)})`
        : `linear-gradient(145deg, ${alpha(theme.palette.background.paper, 0.98)}, ${alpha(theme.palette.primary.main, 0.08)})`,
      agenda: isDark
        ? `linear-gradient(145deg, ${alpha(theme.palette.background.default, 0.96)}, ${alpha(theme.palette.primary.main, 0.3)})`
        : `linear-gradient(145deg, ${alpha(theme.palette.background.paper, 0.98)}, ${alpha(theme.palette.primary.main, 0.06)})`,
      forum: isDark
        ? `linear-gradient(145deg, ${alpha(theme.palette.background.default, 0.96)}, ${alpha(theme.palette.primary.dark, 0.4)})`
        : `linear-gradient(145deg, ${alpha(theme.palette.background.paper, 0.98)}, ${alpha(theme.palette.primary.main, 0.07)})`,
    },

    heroSection: {
      position: 'relative',
      overflow: 'hidden',
      borderRadius: { xs: 5, md: 8 },
      minHeight: { xs: 260, md: 320 },
      backgroundColor: theme.palette.background.default,
      backgroundImage: isDark
        ? `linear-gradient(120deg, ${alpha(theme.palette.primary.dark, 0.72)}, ${alpha(theme.palette.background.default, 0.9)})`
        : `linear-gradient(120deg, ${alpha(theme.palette.primary.light, 0.16)}, ${alpha(theme.palette.background.paper, 0.98)})`,
      border: 'none',
      boxShadow: 'none',
      animation: 'rise 0.6s ease both',
    },
    heroOverline: {
      letterSpacing: '0.26em',
      fontWeight: 700,
      color: 'text.secondary',
    },
    roundButton: {
      borderRadius: 999,
      textTransform: 'none',
      fontWeight: 600,
    },
    eventFloat: {
      animation: 'eventFloat 6s ease-in-out infinite alternate',
    },

    newsCardLink: {
      textDecoration: 'none',
      color: 'inherit',
      display: 'grid',
      gap: 0.3,
      p: 1,
      borderRadius: 2.5,
      border: '1px solid var(--card-border)',
      background: alpha(theme.palette.background.paper, isDark ? 0.7 : 0.85),
      transition: 'border-color 150ms ease, transform 150ms ease',
      '&:hover': {
        borderColor: 'var(--accent-strong)',
        transform: 'translateY(-2px)',
      },
    },
    railReveal: {
      opacity: 0,
      transform: 'translateY(32px)',
      transition: 'opacity 600ms ease, transform 600ms ease',
      '&[data-visible="true"]': { opacity: 1, transform: 'translateY(0)' },
      '@media (prefers-reduced-motion: reduce)': {
        opacity: 1,
        transform: 'none',
        transition: 'none',
      },
    },
    scrollableRail: {
      display: 'grid',
      gridAutoFlow: 'column',
      gridAutoColumns: { xs: '80%', sm: '55%', md: '32%' },
      gap: 1.6,
      overflowX: 'auto',
      scrollSnapType: 'x mandatory',
      pb: 1.2,
      px: 0.4,
      scrollbarWidth: 'thin',
      scrollbarColor: `${alpha(theme.palette.text.primary, 0.2)} transparent`,
      '& > *': { scrollSnapAlign: 'start', scrollSnapStop: 'always' },
      '&::-webkit-scrollbar': { height: 6 },
      '&::-webkit-scrollbar-thumb': { background: alpha(theme.palette.text.primary, 0.2), borderRadius: 999 },
      '&::-webkit-scrollbar-track': { background: 'transparent' },
    },

    timelineDot: (color: string) => ({
      position: 'absolute',
      left: { xs: -16, md: -20 },
      top: 20,
      width: 12,
      height: 12,
      borderRadius: '50%',
      bgcolor: color,
      boxShadow: `0 0 0 6px ${alpha(color, 0.18)}`,
    }),
    calendarDay: (isSelected: boolean, isToday: boolean, inMonth: boolean) => ({
      border: '1px solid',
      borderColor: isSelected
        ? 'var(--accent-strong)'
        : isToday
          ? alpha(theme.palette.success.main, 0.4)
          : 'transparent',
      bgcolor: isSelected ? 'var(--accent-soft)' : 'transparent',
      borderRadius: 1.6,
      height: { xs: 24, md: 28 },
      cursor: 'pointer',
      color: inMonth ? 'var(--ink)' : 'var(--muted)',
      opacity: inMonth ? 1 : 0.5,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      flexDirection: 'column',
      gap: 0.1,
      fontFamily: 'inherit',
      transition: 'transform 120ms ease, border-color 120ms ease',
      '&:hover': {
        transform: 'translateY(-2px)',
        borderColor: 'var(--accent-strong)',
      },
    }),
  };
};


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