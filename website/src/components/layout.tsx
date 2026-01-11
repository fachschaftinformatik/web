import React, { useState, useMemo } from 'react';
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
import { useTheme, alpha, styled } from '@mui/material/styles';
import InputBase from '@mui/material/InputBase';
import SearchRounded from '@mui/icons-material/SearchRounded';
import NotificationsNoneRounded from '@mui/icons-material/NotificationsNoneRounded';
import MailRounded from '@mui/icons-material/MailRounded';
import MenuRounded from '@mui/icons-material/MenuRounded';
import MenuOpenRounded from '@mui/icons-material/MenuOpenRounded';
import DashboardRounded from '@mui/icons-material/DashboardRounded';
import CampaignRounded from '@mui/icons-material/CampaignRounded';
import SchoolRounded from '@mui/icons-material/SchoolRounded';
import QuestionAnswerRounded from '@mui/icons-material/QuestionAnswerRounded';
import CollectionsRounded from '@mui/icons-material/CollectionsRounded';
import PeopleRounded from '@mui/icons-material/PeopleRounded';
import PersonRounded from '@mui/icons-material/PersonRounded';
import SettingsRounded from '@mui/icons-material/SettingsRounded';
import LogoutRounded from '@mui/icons-material/LogoutRounded';
import Brightness4Rounded from '@mui/icons-material/Brightness4Rounded';
import Brightness7Rounded from '@mui/icons-material/Brightness7Rounded';
import GavelRounded from '@mui/icons-material/GavelRounded';
import PolicyRounded from '@mui/icons-material/PolicyRounded';
import DoneAllRounded from '@mui/icons-material/DoneAllRounded';
import { Link as RouterLink } from 'react-router-dom';

import type { AuthUserResponse as User } from '@lib/api/types.gen';
import { useThemeMode } from '@lib/theme';
import { useAuth } from '@lib/auth';
import { getAuthCsrf, putAuthMe } from '@lib/api';

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

const Search = styled('div')(({ theme }) => ({
  position: 'relative',
  borderRadius: theme.shape.borderRadius,
  backgroundColor: alpha(theme.palette.common.white, 0.15),
  '&:hover': {
    backgroundColor: alpha(theme.palette.common.white, 0.25),
  },
  marginLeft: 0,
  width: '100%',
  [theme.breakpoints.up('sm')]: {
    marginLeft: theme.spacing(1),
    width: 'auto',
  },
}));

const SearchIconWrapper = styled('div')(({ theme }) => ({
  padding: theme.spacing(0, 2),
  height: '100%',
  position: 'absolute',
  pointerEvents: 'none',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
}));

const StyledInputBase = styled(InputBase)(({ theme }) => ({
  color: 'inherit',
  width: '100%',
  '& .MuiInputBase-input': {
    padding: theme.spacing(1, 1, 1, 0),
    paddingLeft: `calc(1em + ${theme.spacing(4)})`,
    transition: theme.transitions.create('width'),
    [theme.breakpoints.up('sm')]: {
      width: '12ch',
      '&:focus': {
        width: '20ch',
      },
    },
  },
}));

const navItems = [
  { label: 'Startseite', href: '/', icon: <DashboardRounded />, isRoute: true },
  { label: 'Ankündigungen', href: '/news', icon: <CampaignRounded />, isRoute: true },
  { label: 'Rekos', href: '/exams', icon: <SchoolRounded />, isRoute: true },
  { label: 'Forum', href: '/forum', icon: <QuestionAnswerRounded />, isRoute: true },
  { label: 'Galerie', href: '/media', icon: <CollectionsRounded />, isRoute: true },
  { label: 'Team', href: '/team', icon: <PeopleRounded />, isRoute: true },
  { label: 'Impressum', href: '/legal', icon: <GavelRounded />, isRoute: true },
  { label: 'Datenschutz', href: '/privacy', icon: <PolicyRounded />, isRoute: true },
];

interface Notification {
  id: string;
  title: string;
  message: string;
  time: string;
  type: 'news' | 'forum' | 'exam';
  read: boolean;
}

const mockNotifications: Notification[] = [
  { id: '1', title: 'Neue News', message: 'Wichtige Infos zum Semesterstart wurden veröffentlicht.', time: 'Vor 5 Min', type: 'news', read: false },
  { id: '2', title: 'Forumsantwort', message: 'Jemand hat auf deinen Beitrag "Prüfungsvorbereitung" geantwortet.', time: 'Vor 1 Std', type: 'forum', read: false },
  { id: '3', title: 'Neue Reko', message: 'Eine neue Klausur-Rekonstruktion für "Lineare Algebra" ist verfügbar.', time: 'Gestern', type: 'exam', read: true },
  { id: '4', title: 'News-Update', message: 'Der Termin für das Camp-Wochenende steht fest!', time: 'Vor 2 Tagen', type: 'news', read: true },
];

interface SidebarProps {
  user?: User | null;
  children: React.ReactNode;
  title?: string;
  headerActions?: React.ReactNode;
  fullBleed?: boolean;
  maxWidth?: "xs" | "sm" | "md" | "lg" | "xl" | false;
}

const Sidebar = ({
  user,
  children,
  title = 'Dashboard',
  headerActions,
  fullBleed = false,
  maxWidth = "lg"
}: SidebarProps) => {
  const { mode } = useThemeMode();
  const theme = useTheme();
  const isMdUp = useMediaQuery(theme.breakpoints.up('md'));

  React.useEffect(() => {
    document.title = `${title} | FSV Informatik`;
  }, [title]);


  const [desktopOpen, setDesktopOpen] = useState(() => {
    const stored = localStorage.getItem('sidebar_desktop_open');
    return stored !== null ? stored === 'true' : true;
  });
  const [mobileOpen, setMobileOpen] = useState(false);

  const handleDrawerToggle = () => {
    if (isMdUp) {
      const newState = !desktopOpen;
      setDesktopOpen(newState);
      localStorage.setItem('sidebar_desktop_open', String(newState));
    } else {
      setMobileOpen(!mobileOpen);
    }
  };

  const navOpen = isMdUp ? desktopOpen : mobileOpen;


  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [notiAnchorEl, setNotiAnchorEl] = useState<null | HTMLElement>(null);
  const [notifications, setNotifications] = useState<Notification[]>(mockNotifications);

  const open = Boolean(anchorEl);
  const notiOpen = Boolean(notiAnchorEl);

  const unreadCount = notifications.filter(n => !n.read).length;

  const { logout, login } = useAuth();
  const { setPreference } = useThemeMode();

  const handleMarkAllRead = () => {
    setNotifications(prev => prev.map(n => ({ ...n, read: true })));
  };

  const handleNotiClick = (e: React.MouseEvent<HTMLElement>) => {
    setNotiAnchorEl(e.currentTarget);
  };

  const handleNotiClose = () => {
    setNotiAnchorEl(null);
  };

  const handleThemeToggle = async () => {
    const nextMode = mode === 'light' ? 'dark' : 'light';

    // Update locally first for instant feedback
    setPreference(nextMode);

    if (user) {
      try {
        const { data: csrfData } = await getAuthCsrf();
        const token = csrfData?.csrf;
        const res = await putAuthMe({
          body: {
            name: user.name,
            programid: user.programid,
            theme: nextMode
          } as any,
          headers: { "X-CSRF-Token": token || "" }
        });

        if (res.data) {
          login(res.data as any, window.localStorage.getItem('fs_remember_flag') === 'true');
        }
      } catch (err) {
        console.error("Failed to sync theme preference", err);
      }
    }
  };

  const avatarColor = useMemo(() => stringToColor(user?.name), [user?.name]);
  const avatarLetter = user?.name ? user.name.charAt(0).toUpperCase() : '?';

  const handleMenuClose = () => setAnchorEl(null);

  const toggleBtnStyle = {
    border: '1px solid ' + alpha(theme.palette.primary.main, mode === 'dark' ? 0.5 : 0.25),
    transition: theme.transitions.create(['background-color', 'border-color'], {
      duration: theme.transitions.duration.shortest,
    }),
    '&:hover': {
      borderColor: alpha(theme.palette.primary.main, 0.6),
      bgcolor: mode === 'dark'
        ? alpha(theme.palette.primary.dark, 0.55)
        : alpha(theme.palette.primary.main, 0.2)
    },
    bgcolor: mode === 'dark'
      ? alpha(theme.palette.primary.dark, 0.35)
      : alpha(theme.palette.primary.main, 0.1),
    color: mode === 'dark'
      ? theme.palette.primary.contrastText
      : theme.palette.primary.main
  };

  const drawerContent = (
    <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%', overflowY: 'auto' }}>
      {!isMdUp && (
        <>
          <Toolbar sx={{ justifyContent: 'flex-end' }}>
            <IconButton onClick={() => setMobileOpen(false)}>
              <MenuOpenRounded />
            </IconButton>
          </Toolbar>
          <Divider />
        </>
      )}

      <List sx={{ pt: 1, flexGrow: 1 }}>
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

      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', py: 1, mb: 1 }}>
        <Tooltip title="Farbschema wechseln">
          <IconButton
            aria-label="toggle color mode"
            onClick={handleThemeToggle}
            sx={toggleBtnStyle}
          >
            {mode === 'dark' ? <Brightness7Rounded /> : <Brightness4Rounded />}
          </IconButton>
        </Tooltip>
      </Box>
    </Box>
  );

  return (
    <Box sx={{ display: 'flex' }}>
      <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
        <Toolbar sx={{ position: 'relative' }}>
          <IconButton
            color="inherit"
            edge="start"
            onClick={handleDrawerToggle}
            sx={{ mr: 2, display: { xs: 'block', md: 'block' } }}
          >
            {navOpen ? <MenuOpenRounded /> : <MenuRounded />}
          </IconButton>

          <Typography variant="h6" noWrap component="div" sx={{ flexGrow: 1, display: { xs: 'none', sm: 'block' } }}>
            {title}
          </Typography>

          <Search sx={{ mr: 2 }}>
            <SearchIconWrapper>
              <SearchRounded />
            </SearchIconWrapper>
            <StyledInputBase
              placeholder="Suchen…"
              inputProps={{ 'aria-label': 'search' }}
            />
          </Search>

          {headerActions}

          {user ? (
            <>
              <Tooltip title="Posteingang">
                <span>
                  <IconButton color="inherit" size="large" disabled>
                    <MailRounded />
                  </IconButton>
                </span>
              </Tooltip>

              <Tooltip title="Benachrichtigungen">
                <IconButton
                  color="inherit"
                  size="large"
                  sx={{ mr: 1 }}
                  onClick={handleNotiClick}
                >
                  <Badge badgeContent={unreadCount} color="error">
                    <NotificationsNoneRounded />
                  </Badge>
                </IconButton>
              </Tooltip>

              <Menu
                anchorEl={notiAnchorEl}
                id="notifications-menu"
                open={notiOpen}
                onClose={handleNotiClose}
                transformOrigin={{ horizontal: 'right', vertical: 'top' }}
                anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
                PaperProps={{
                  elevation: 0,
                  sx: {
                    mt: 1.5,
                    width: 340,
                    maxHeight: 500,
                    borderRadius: '12px',
                    overflow: 'visible',
                    filter: 'drop-shadow(0px 4px 20px rgba(0,0,0,0.15))',
                    border: '1px solid',
                    borderColor: 'divider',
                    bgcolor: 'background.paper',
                    backdropFilter: 'blur(10px)',
                    overflowY: 'auto'
                  },
                }}
              >
                <Box sx={{
                  position: 'sticky',
                  top: 0,
                  bgcolor: 'background.paper',
                  zIndex: 2,
                  borderBottom: '1px solid',
                  borderColor: 'divider',
                  minHeight: 56,
                  display: 'flex',
                  alignItems: 'center'
                }}>
                  <Box sx={{ px: 2, py: 1, width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <Typography variant="subtitle1" fontWeight={700}>Benachrichtigungen</Typography>
                    <Button
                      size="small"
                      startIcon={<DoneAllRounded fontSize="small" />}
                      onClick={handleMarkAllRead}
                      sx={{ textTransform: 'none', fontSize: '0.75rem', py: 0.5 }}
                    >
                      Alle gelesen
                    </Button>
                  </Box>
                </Box>
                <List sx={{ p: 0 }}>
                  {notifications.map((noti) => (
                    <MenuItem
                      key={noti.id}
                      onClick={handleNotiClose}
                      sx={{
                        py: 2,
                        px: 2,
                        borderBottom: '1px solid',
                        borderColor: alpha(theme.palette.divider, 0.05),
                        whiteSpace: 'normal',
                        bgcolor: noti.read ? 'transparent' : alpha(theme.palette.primary.main, 0.04),
                        '&:hover': {
                          bgcolor: alpha(theme.palette.primary.main, 0.08)
                        }
                      }}
                    >
                      <ListItemIcon sx={{ minWidth: 44 }}>
                        <Avatar sx={{
                          width: 32,
                          height: 32,
                          bgcolor: noti.type === 'news'
                            ? alpha(theme.palette.info.main, 0.12)
                            : noti.type === 'forum'
                              ? alpha(theme.palette.success.main, 0.12)
                              : alpha(theme.palette.warning.main, 0.12),
                          color: noti.type === 'news'
                            ? 'info.main'
                            : noti.type === 'forum'
                              ? 'success.main'
                              : 'warning.main',
                          fontSize: '1rem'
                        }}>
                          {noti.type === 'news' ? <CampaignRounded fontSize="small" /> : noti.type === 'forum' ? <QuestionAnswerRounded fontSize="small" /> : <SchoolRounded fontSize="small" />}
                        </Avatar>
                      </ListItemIcon>
                      <Box sx={{ flexGrow: 1 }}>
                        <Typography variant="body2" sx={{ fontWeight: noti.read ? 600 : 700, mb: 0.2 }}>
                          {noti.title}
                        </Typography>
                        <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.5, lineHeight: 1.3 }}>
                          {noti.message}
                        </Typography>
                        <Typography variant="caption" sx={{ color: 'text.disabled', fontSize: '0.7rem' }}>
                          {noti.time}
                        </Typography>
                      </Box>
                      {!noti.read && (
                        <Box sx={{ width: 8, height: 8, borderRadius: '50%', bgcolor: 'primary.main', ml: 1, flexShrink: 0 }} />
                      )}
                    </MenuItem>
                  ))}
                </List>
                {notifications.length === 0 && (
                  <Box sx={{ p: 4, textAlign: 'center' }}>
                    <Typography variant="body2" color="text.secondary">Keine neuen Benachrichtigungen</Typography>
                  </Box>
                )}
              </Menu>

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
                    filter: 'drop-shadow(0px 4px 20px rgba(0,0,0,0.15))',
                    mt: 1.5,
                    minWidth: 200,
                    borderRadius: '12px',
                    border: '1px solid',
                    borderColor: 'divider',
                    bgcolor: (theme) => alpha(theme.palette.background.paper, 0.98),
                    backdropFilter: 'blur(10px)',
                  },
                }}
              >
                <MenuItem component={RouterLink} to={user ? `/user/${user.id}` : "/login"} sx={{ py: 1.2 }}>
                  <ListItemIcon>
                    <PersonRounded fontSize="small" />
                  </ListItemIcon>
                  Mein Profil
                </MenuItem>
                <MenuItem component={RouterLink} to="/settings" sx={{ py: 1.2 }}>
                  <ListItemIcon>
                    <SettingsRounded fontSize="small" />
                  </ListItemIcon>
                  Einstellungen
                </MenuItem>
                <Divider />
                <MenuItem onClick={() => { logout(); }} sx={{ py: 1.2 }}>
                  <ListItemIcon>
                    <LogoutRounded fontSize="small" color="error" />
                  </ListItemIcon>
                  <Typography color="error" fontWeight={600}>Abmelden</Typography>
                </MenuItem>
              </Menu>
            </>
          ) : (
            <Button color="inherit" component={RouterLink} to="/login">Anmelden</Button>
          )}
        </Toolbar>
      </AppBar>

      <Drawer
        variant={isMdUp ? "permanent" : "temporary"}
        open={navOpen}
        onClose={() => setMobileOpen(false)}
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

      <Box component="main" sx={{ flexGrow: 1, p: 0, width: '100%', minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
        <Toolbar />
        <Container maxWidth={fullBleed ? false : maxWidth} disableGutters sx={{ flexGrow: 1, p: fullBleed ? 0 : { xs: 2.5, md: 5 } }}>
          {children}
        </Container>
      </Box>
    </Box>
  );
};

export { Sidebar };
