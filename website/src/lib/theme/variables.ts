import { PaletteMode, PaletteOptions } from '@mui/material/styles';

export const lightTokens = {
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

export const darkTokens = {
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

export const appBarGreen = lightTokens.brand.primary.main;

export const getPalette = (mode: PaletteMode): PaletteOptions => {
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
