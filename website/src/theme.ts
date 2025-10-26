// src/theme.ts

import { createTheme } from '@mui/material/styles';

// Das ist die zentrale Style-Datei.
// Hier kann dein Kollege alle Design-Anpassungen vornehmen.
const theme = createTheme({
  palette: {
    // Beispiel: Ändere hier die Hauptfarbe der App
    primary: {
      main: '#046709be', // Das ist das Standard-Blau von MUI
    },
    // Beispiel: Ändere hier die Sekundärfarbe
    secondary: {
      main: '#000000de', // Das ist das Standard-Pink
    },
  },
  typography: {
    // Beispiel: Ändere die Schriftart der App
    fontFamily: 'Roboto, Arial, sans-serif',
    h5: {
      fontWeight: 700, // Macht die Überschriften etwas dicker
    },
  },
  components: {
    // Beispiel: Ändere den Standard-Stil aller Buttons
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 8, // Macht alle Buttons etwas runder
          textTransform: 'none', // Verhindert, dass Text in Buttons großgeschrieben wird
        },
      },
    },
  },
});

export default theme;