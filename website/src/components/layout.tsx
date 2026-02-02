import { useState } from 'react';
import { Outlet } from 'react-router-dom';
import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import useMediaQuery from '@mui/material/useMediaQuery';
import { useTheme } from '@mui/material/styles';
import { useAuth } from '@lib/auth';
import Navbar from './Navbar';
import Sidebar from './Sidebar';
import Status from './Status';
import { LayoutContextProvider } from './Page';

export default function Layout() {
  const theme = useTheme();
  const isMdUp = useMediaQuery(theme.breakpoints.up('md'));
  const { user } = useAuth();
  const [title, setTitle] = useState('Dashboard');
  const [headerActions, setHeaderActions] = useState<React.ReactNode>(null);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [desktopOpen, setDesktopOpen] = useState(() => localStorage.getItem('sidebar_desktop_open') !== 'false');

  const toggleDrawer = () => {
    if (isMdUp) {
      const next = !desktopOpen;
      setDesktopOpen(next);
      localStorage.setItem('sidebar_desktop_open', String(next));
    } else setMobileOpen(!mobileOpen);
  };

  return (
    <LayoutContextProvider value={{ setTitle, setHeaderActions }}>
      <Box sx={{ display: 'flex' }}>
        <Navbar title={title} headerActions={headerActions} navOpen={isMdUp ? desktopOpen : mobileOpen} onDrawerToggle={toggleDrawer} />
        <Sidebar open={desktopOpen} mobileOpen={mobileOpen} onMobileClose={() => setMobileOpen(false)} isMdUp={isMdUp} />
        <Box component="main" sx={{ flexGrow: 1, p: 0, width: '100%', minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
          <Toolbar />
          <Status user={user} />
          <Outlet />
        </Box>
      </Box>
    </LayoutContextProvider>
  );
}
