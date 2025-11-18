import React from 'react';
import { Fab, Tooltip } from '@mui/material';
import { useTheme } from '@mui/material/styles';
import Brightness4Rounded from '@mui/icons-material/Brightness4Rounded';
import Brightness7Rounded from '@mui/icons-material/Brightness7Rounded';
import { useThemeMode } from './ThemeModeContext';

const ThemeModeToggle: React.FC = () => {
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

export default ThemeModeToggle;
