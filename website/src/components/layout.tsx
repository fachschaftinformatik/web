import React, { useState, useEffect, useMemo } from 'react';
import {
  AppBar,
  Avatar,
  Badge,
  Box,
  Button,
  Container,
  Divider,
  Drawer,
  IconButton,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Menu,
  MenuItem,
  Toolbar,
  Tooltip,
  Typography,
  useMediaQuery,
} from '@mui/material';
import { useTheme } from '@mui/material/styles';
import SearchRounded from '@mui/icons-material/SearchRounded';
import NotificationsNoneRounded from '@mui/icons-material/NotificationsNoneRounded';
import MenuRounded from '@mui/icons-material/MenuRounded';
import MenuOpenRounded from '@mui/icons-material/MenuOpenRounded';
import DashboardRounded from '@mui/icons-material/DashboardRounded';
import CampaignRounded from '@mui/icons-material/CampaignRounded';
import SchoolRounded from '@mui/icons-material/SchoolRounded';
import QuestionAnswerRounded from '@mui/icons-material/QuestionAnswerRounded';
import CollectionsRounded from '@mui/icons-material/CollectionsRounded';
import PeopleRounded from '@mui/icons-material/PeopleRounded';
import PersonRounded from '@mui/icons-material/PersonRounded';
import LogoutRounded from '@mui/icons-material/LogoutRounded';
import { Link as RouterLink } from 'react-router-dom';

import type { User } from '@lib/api/types.gen';
import { useThemeMode } from '@lib/theme';
import { useAuth } from '@lib/auth';

const drawerWidthOpen = 240;
const drawerWidthClosed = 72;

const avatarPalette = [
  '#d32f2f',
  '#c2185b',
  '#7b1fa2',
  '#512da8',
  '#303f9f',
  '#1976d2',
  '#0288d1',
  '#0097a7',
  '#00796b',
  '#388e3c',
  '#e64a19',
  '#5d4037',
  '#455a64',
];

const stringToColor = (string?: string) => {
  if (!string) return avatarPalette[0];
  let hash = 0;
  for (let i = 0; i < string.length; i++) {
    hash = string.charCodeAt(i) + ((hash << 5) - hash);
  }
  const index = Math.abs(hash % avatarPalette.length);
  return avatarPalette[index];
};

const navItems = [
  { label: 'Startseite', href: '/', icon: <DashboardRounded />, isRoute: true },
  { label: 'Ankündigungen', href: '/news', icon: <CampaignRounded />, isRoute: true },
  { label: 'Rekos', href: '/exams', icon: <SchoolRounded />, isRoute: true },
  { label: 'Forum', href: '/forum', icon: <QuestionAnswerRounded />, isRoute: true },
  { label: 'Galerie', href: '/media', icon: <CollectionsRounded />, isRoute: true },
  { label: 'Team', href: '/team', icon: <PeopleRounded />, isRoute: true },
];

const Sidebar = ({ user, children, title = 'Dashboard' }: { user?: User | null; children: React.ReactNode; title?: string }) => {
  const { mode } = useThemeMode();
  const theme = useTheme();
  const isMdUp = useMediaQuery(theme.breakpoints.up('md'));
  const [navOpen, setNavOpen] = useState(isMdUp);
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);
  const { logout } = useAuth();

  const avatarColor = useMemo(() => stringToColor(user?.name), [user?.name]);
  const avatarLetter = user?.name ? user.name.charAt(0).toUpperCase() : '?';

  useEffect(() => {
    setNavOpen(isMdUp);
  }, [isMdUp]);

  const handleMenuClose = () => setAnchorEl(null);
  
  const drawerContent = (
    <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      {!isMdUp && (
        <>
          <Toolbar sx={{ justifyContent: 'flex-end' }}>
             <IconButton onClick={() => setNavOpen(false)}>
               <MenuOpenRounded />
             </IconButton>
          </Toolbar>
          <Divider />
        </>
      )}
      
      <List sx={{ pt: 1 }}>
        {navItems.map((item) => (
          <Tooltip key={item.label} title={navOpen ? '' : item.label} placement="right">
            <ListItemButton
              component={item.isRoute ? RouterLink : 'a'}
              to={item.isRoute ? item.href : undefined}
              href={!item.isRoute ? item.href : undefined}
              sx={{
                minHeight: 48,
                justifyContent: navOpen ? 'initial' : 'center',
                px: 2.5,
              }}
            >
              <ListItemIcon sx={{ minWidth: 0, mr: navOpen ? 3 : 'auto', justifyContent: 'center' }}>
                {item.icon}
              </ListItemIcon>
              <ListItemText primary={item.label} sx={{ opacity: navOpen ? 1 : 0 }} />
            </ListItemButton>
          </Tooltip>
        ))}
      </List>
    </Box>
  );

  return (
    <Box sx={{ display: 'flex' }}>
      <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
        <Toolbar>
          <IconButton
            color="inherit"
            edge="start"
            onClick={() => setNavOpen(!navOpen)}
            sx={{ mr: 2, display: { xs: 'block', md: 'block' } }}
          >
             {navOpen ? <MenuOpenRounded /> : <MenuRounded />}
          </IconButton>
          
          <Typography variant="h6" noWrap component="div" sx={{ flexGrow: 1 }}>
            {title}
          </Typography>

          {user ? (
            <>
              <Tooltip title="Suche">
                <IconButton color="inherit" size="large">
                  <SearchRounded />
                </IconButton>
              </Tooltip>

              <Tooltip title="Benachrichtigungen">
                <IconButton color="inherit" size="large" sx={{ mr: 1 }}>
                  <Badge badgeContent={4} color="error">
                    <NotificationsNoneRounded />
                  </Badge>
                </IconButton>
              </Tooltip>

              <Tooltip title="Account">
                <IconButton 
                  onClick={(e) => setAnchorEl(e.currentTarget)} 
                  size="small"
                  aria-controls={open ? 'account-menu' : undefined}
                  aria-haspopup="true"
                  aria-expanded={open ? 'true' : undefined}
                >
                  <Avatar 
                    sx={{ 
                      width: 32, 
                      height: 32,
                      bgcolor: avatarColor,
                      fontSize: '0.9rem',
                      fontWeight: 600
                    }}
                  >
                      {avatarLetter}
                  </Avatar>
                </IconButton>
              </Tooltip>
              
              <Menu
                anchorEl={anchorEl}
                id="account-menu"
                open={open}
                onClose={handleMenuClose}
                onClick={handleMenuClose}
                transformOrigin={{ horizontal: 'right', vertical: 'top' }}
                anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
                PaperProps={{
                  elevation: 0,
                  sx: {
                    overflow: 'visible',
                    filter: 'drop-shadow(0px 2px 8px rgba(0,0,0,0.32))',
                    mt: 1.5,
                    minWidth: 150,
                  },
                }}
              >
                <MenuItem component={RouterLink} to="/profile">
                  <ListItemIcon>
                    <PersonRounded fontSize="small" />
                  </ListItemIcon>
                  Mein Profil
                </MenuItem>
                <MenuItem onClick={() => { logout(); }}>
                  <ListItemIcon>
                    <LogoutRounded fontSize="small" />
                  </ListItemIcon>
                  Logout
                </MenuItem>
              </Menu>
            </>
          ) : (
            <Button color="inherit" component={RouterLink} to="/login">Login</Button>
          )}
        </Toolbar>
      </AppBar>

      <Drawer
        variant={isMdUp ? "permanent" : "temporary"}
        open={navOpen}
        onClose={() => setNavOpen(false)}
        sx={{
          width: navOpen ? drawerWidthOpen : drawerWidthClosed,
          flexShrink: 0,
          whiteSpace: 'nowrap',
          boxSizing: 'border-box',
          '& .MuiDrawer-paper': {
            width: navOpen ? drawerWidthOpen : drawerWidthClosed,
            transition: theme.transitions.create('width', {
              easing: theme.transitions.easing.sharp,
              duration: theme.transitions.duration.enteringScreen,
            }),
            overflowX: 'hidden',
          },
        }}
      >
        <Toolbar />
        {drawerContent}
      </Drawer>

      <Box component="main" sx={{ flexGrow: 1, p: 3, width: '100%' }}>
        <Toolbar />
        <Container maxWidth="xl" disableGutters>
            {children}
        </Container>
      </Box>
    </Box>
  );
};

export { Sidebar };
