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
