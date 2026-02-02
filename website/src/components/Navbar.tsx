import { Link as RouterLink } from 'react-router-dom';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import Tooltip from '@mui/material/Tooltip';
import { useTheme, alpha } from '@mui/material/styles';
import MenuRounded from '@mui/icons-material/MenuRounded';
import MenuOpenRounded from '@mui/icons-material/MenuOpenRounded';
import MailRounded from '@mui/icons-material/MailRounded';
import SettingsRounded from '@mui/icons-material/SettingsRounded';
import { useAuth } from '@lib/auth';
import Search from './Search';
import Notify from './Notify';
import Account from './Account';

interface NavbarProps {
  title: string;
  headerActions?: React.ReactNode;
  navOpen: boolean;
  onDrawerToggle: () => void;
}

export default function Navbar({ title, headerActions, navOpen, onDrawerToggle }: NavbarProps) {
  const theme = useTheme();
  const { user } = useAuth();

  return (
    <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
      <Toolbar sx={{ position: 'relative' }}>
        <IconButton color="inherit" edge="start" onClick={onDrawerToggle} sx={{ mr: 2 }}>
          {navOpen ? <MenuOpenRounded /> : <MenuRounded />}
        </IconButton>
        <Stack direction="row" spacing={1} alignItems="center" sx={{ flexGrow: 1 }}>
          <Typography variant="h6" noWrap component="div" sx={{ fontWeight: 800, letterSpacing: '-0.02em', fontFamily: '"Space Grotesk", sans-serif', display: { xs: 'none', sm: 'block' } }}>
            fsInformatik
          </Typography>
          <Typography variant="h6" noWrap component="div" sx={{ fontWeight: 400, color: alpha(theme.palette.common.white, 0.6), display: { xs: 'none', md: 'block' } }}>
            /
          </Typography>
          <Typography variant="h6" noWrap component="div" sx={{ fontWeight: 400, display: { xs: 'none', sm: 'block' } }}>
            {title}
          </Typography>
        </Stack>
        <Search />
        {headerActions}
        {user ? (
          <Stack direction="row" spacing={0.5} alignItems="center">
            <Tooltip title="Posteingang">
              <span>
                <IconButton color="inherit" size="large" disabled>
                  <MailRounded />
                </IconButton>
              </span>
            </Tooltip>
            <Notify />
            {user.role === 'admin' && (
              <Tooltip title="Administration">
                <IconButton component={RouterLink} to="/admin" color="inherit" size="large">
                  <SettingsRounded />
                </IconButton>
              </Tooltip>
            )}
            <Account />
          </Stack>
        ) : (
          <Button
            variant="contained"
            component={RouterLink}
            to="/login"
            sx={{
              borderRadius: 2,
              textTransform: 'none',
              fontWeight: 700,
              px: 3,
              bgcolor: 'common.white',
              color: 'primary.main',
              boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
              '&:hover': {
                bgcolor: alpha(theme.palette.common.white, 0.9),
                boxShadow: '0 6px 16px rgba(0,0,0,0.15)',
                transform: 'translateY(-1px)'
              },
              transition: 'all 0.2s ease'
            }}
          >
            Anmelden
          </Button>
        )}
      </Toolbar>
    </AppBar>
  );
}
