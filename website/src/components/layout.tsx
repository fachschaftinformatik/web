import React, { useState, useEffect, useCallback } from 'react';
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
  Stack,
  Toolbar,
  Tooltip,
  Typography,
  useMediaQuery,
  Autocomplete,
  CircularProgress,
  ListItem,
} from '@mui/material';
import { useTheme, alpha, styled } from '@mui/material/styles';
import InputBase from '@mui/material/InputBase';
import SearchIcon from '@mui/icons-material/Search';
import NotificationsNoneRounded from '@mui/icons-material/NotificationsNoneRounded';
import MailRounded from '@mui/icons-material/MailRounded';
import MenuRounded from '@mui/icons-material/MenuRounded';
import LibraryBooksRounded from '@mui/icons-material/LibraryBooksRounded';

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
import DoneAllRounded from '@mui/icons-material/DoneAllRounded';
import CloseRounded from '@mui/icons-material/CloseRounded';
import { Link as RouterLink, useNavigate } from 'react-router-dom';

import type { DtoUserResponse as User, DtoNotificationResponse as Notification } from '@lib/api/types.gen';
import { useThemeMode } from '@lib/theme';
import { useAuth } from '@lib/auth';
import {
  putAuthMe,
  getAuthNotifications,
  putAuthNotificationsIdRead,
  putAuthNotificationsReadAll
} from '@lib/api';
import { client } from '@lib/api/client.gen';
import { getAvatarUrl } from '@lib/images';
import OfficeStatus from './OfficeStatus';

const drawerWidthOpen = 240;
const drawerWidthClosed = 72;

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
interface SearchResult {
  type: 'archive' | 'module' | 'user' | 'discussion';
  id: string;
  title: string;
  subtitle: string;
  url: string;
}

const GlobalSearch = () => {
  const theme = useTheme();
  const isSmUp = useMediaQuery(theme.breakpoints.up('sm'));
  const [mobileOpen, setMobileOpen] = useState(false);
  const [open, setOpen] = useState(false);
  const [options, setOptions] = useState<readonly SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [inputValue, setInputValue] = useState('');
  const navigate = useNavigate();
  const isSearchOpen = open && inputValue.length >= 2;

  useEffect(() => {
    if (isSearchOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [isSearchOpen]);

  useEffect(() => {
    let active = true;

    if (inputValue.length < 2) {
      Promise.resolve().then(() => {
        setOptions([]);
        setLoading(false);
      });
      return undefined;
    }

    Promise.resolve().then(() => {
      setLoading(true);
    });

    const timer = setTimeout(() => {
      client.request({
        method: 'GET',
        url: '/search',
        query: { q: inputValue }
      }).then(({ data }) => {
        if (active && data) {
          setOptions(data as SearchResult[]);
        }
      }).finally(() => {
        if (active) setLoading(false);
      });
    }, 400);

    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [inputValue]);

  if (!isSmUp && !mobileOpen) {
    return (
      <IconButton color="inherit" onClick={() => setMobileOpen(true)}>
        <SearchIcon />
      </IconButton>
    );
  }

  return (
    <Box
      sx={!isSmUp && mobileOpen ? {
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        bgcolor: alpha(theme.palette.background.paper, 0.9),
        backdropFilter: 'blur(10px)',
        display: 'flex',
        alignItems: 'center',
        px: 1.5,
        zIndex: 1200,
      } : { mr: 1 }}
    >
      <Autocomplete
        id="global-search"
        fullWidth
        open={open && inputValue.length >= 2}
        onOpen={() => setOpen(true)}
        onClose={() => setOpen(false)}
        isOptionEqualToValue={(option, value) => String(option.id) === String(value.id)}
        getOptionLabel={(option) => option.title}
        filterOptions={(x) => x}
        options={options}
        loading={loading}
        value={null}
        onInputChange={(_, newInputValue) => setInputValue(newInputValue)}
        onChange={(_, value) => {
          if (value) {
            navigate(value.url);
            setInputValue('');
            setOptions([]);
            if (!isSmUp) setMobileOpen(false);
          }
        }}
        noOptionsText="Keine Ergebnisse"
        loadingText="Wird geladen..."
        clearIcon={null}
        popupIcon={null}
        sx={{
          '& .MuiAutocomplete-inputRoot': {
            flexWrap: 'nowrap'
          }
        }}
        renderInput={(params) => (
          <Search
            ref={params.InputProps.ref}
            sx={!isSmUp ? {
              mx: 0,
              flex: 1,
              backgroundColor: alpha(theme.palette.text.primary, 0.05),
              color: theme.palette.text.primary,
              '&:hover': {
                backgroundColor: alpha(theme.palette.text.primary, 0.08),
              },
              '& .MuiInputBase-input': {
                color: theme.palette.text.primary,
                '&::placeholder': {
                  color: theme.palette.text.secondary,
                  opacity: 1
                }
              },
              '& .MuiSvgIcon-root': {
                color: theme.palette.primary.main
              }
            } : {}}
          >
            <SearchIconWrapper>
              {loading ? <CircularProgress color="inherit" size={20} /> : <SearchIcon />}
            </SearchIconWrapper>
            <StyledInputBase
              inputProps={params.inputProps}
              placeholder="Suchen…"
              autoFocus={!isSmUp}
            />
          </Search>
        )}
        renderOption={(props, option) => (
          <ListItem {...props} key={String(option.id) + option.type}>
            <ListItemIcon sx={{ minWidth: 40 }}>
              {option.type === 'module' ? <DashboardRounded fontSize="small" /> :
                option.type === 'user' ? (
                  <Avatar 
                    src={getAvatarUrl(`/api/v1/auth/avatars/${option.id}/generated_v4.svg`)}
                    sx={{ width: 24, height: 24 }}
                  />
                ) :
                  option.type === 'discussion' ? <QuestionAnswerRounded fontSize="small" /> :
                    <SchoolRounded fontSize="small" />}
            </ListItemIcon>
            <ListItemText
              primary={option.title}
              secondary={option.subtitle}
              primaryTypographyProps={{ variant: 'body2', fontWeight: 600 }}
              secondaryTypographyProps={{ variant: 'caption' }}
            />
          </ListItem>
        )}
      />
      {!isSmUp && (
        <IconButton color="inherit" onClick={() => setMobileOpen(false)} sx={{ ml: 1 }}>
          <CloseRounded />
        </IconButton>
      )}
    </Box>
  );
};

const navItems = [
  { label: 'Startseite', href: '/', icon: <DashboardRounded />, isRoute: true },
  { label: 'Diskussionen', href: '/discussions', icon: <QuestionAnswerRounded />, isRoute: true },
  { label: 'Galerie', href: '/events', icon: <CollectionsRounded />, isRoute: true },
  { label: 'Archiv', href: '/archive', icon: <LibraryBooksRounded />, isRoute: true },
  { label: 'Team', href: '/team', icon: <PeopleRounded />, isRoute: true },
  { label: 'Kontakt', href: '/contact', icon: <MailRounded />, isRoute: true },
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
  const { mode, setPreference } = useThemeMode();
  const theme = useTheme();
  const isMdUp = useMediaQuery(theme.breakpoints.up('md'));
  const navigate = useNavigate();
  const { logout, login } = useAuth();

  useEffect(() => {
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
  const [notifications, setNotifications] = useState<Notification[]>([]);

  const open = Boolean(anchorEl);
  const notiOpen = Boolean(notiAnchorEl);
  const unreadCount = notifications.filter(n => !n.read).length;

  const fetchNotifications = useCallback(async () => {
    if (!user) return;
    try {
      const { data } = await getAuthNotifications();
      if (data) setNotifications(data as Notification[]);
    } catch (err) {
      console.error("Failed to fetch notifications", err);
    }
  }, [user]);

  useEffect(() => {
    const timer = setTimeout(() => {
      void fetchNotifications();
    }, 0);
    return () => clearTimeout(timer);
  }, [fetchNotifications]);

  const handleMarkAllRead = async () => {
    try {
      await putAuthNotificationsReadAll();
      fetchNotifications();
    } catch (err) {
      console.error("Failed to mark all read", err);
    }
  };

  const handleMarkRead = async (id: string) => {
    try {
      await putAuthNotificationsIdRead({
        path: { notificationId: id }
      });
      fetchNotifications();
    } catch (err) {
      console.error("Failed to mark notification as read", err);
    }
  };

  const handleNotiClick = (e: React.MouseEvent<HTMLElement>) => {
    setNotiAnchorEl(e.currentTarget);
  };

  const handleNotiClose = () => {
    setNotiAnchorEl(null);
  };

  const handleThemeToggle = async () => {
    const nextMode = mode === 'light' ? 'dark' : 'light';
    setPreference(nextMode);
    if (user) {
      try {
        const res = await putAuthMe({
          body: {
            name: user.name || "",
            program_id: Number(user.program_id),
            theme: nextMode
          }
        });
        if (res.data) {
          login(res.data as User, window.localStorage.getItem('fs_remember_flag') === 'true');
        }
      } catch (err) {
        console.error("Failed to sync theme preference", err);
      }
    }
  };

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
    <Box sx={{
      display: 'flex',
      flexDirection: 'column',
      height: '100%',
      overflowY: 'auto',
      '&::-webkit-scrollbar': { display: 'none' },
      scrollbarWidth: 'none',
      msOverflowStyle: 'none'
    }}>
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
            sx={{ mr: 2 }}
          >
            {navOpen ? <MenuOpenRounded /> : <MenuRounded />}
          </IconButton>

          <Stack direction="row" spacing={1} alignItems="center" sx={{ flexGrow: 1 }}>
            <Typography
              variant="h6"
              noWrap
              component="div"
              sx={{
                fontWeight: 800,
                letterSpacing: '-0.02em',
                fontFamily: '"Space Grotesk", sans-serif',
                display: { xs: 'none', sm: 'block' }
              }}
            >
              fsInformatik
            </Typography>
            <Typography
              variant="h6"
              noWrap
              component="div"
              sx={{
                fontWeight: 400,
                color: alpha(theme.palette.common.white, 0.6),
                display: { xs: 'none', md: 'block' }
              }}
            >
              /
            </Typography>
            <Typography
              variant="h6"
              noWrap
              component="div"
              sx={{
                fontWeight: 400,
                display: { xs: 'none', sm: 'block' }
              }}
            >
              {title}
            </Typography>
          </Stack>

          <GlobalSearch />

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
                      disabled={unreadCount === 0}
                      sx={{ textTransform: 'none', fontSize: '0.75rem', py: 0.5 }}
                    >
                      Alle gelesen
                    </Button>
                  </Box>
                </Box>
                <List sx={{ p: 0 }}>
                  {notifications.map((noti) => (
                    <MenuItem
                      key={String(noti.id)}
                      onClick={() => {
                        if (noti.id) handleMarkRead(String(noti.id));
                        if (noti.link) navigate(noti.link);
                        handleNotiClose();
                      }}
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
                            : noti.type === 'discussion'
                              ? alpha(theme.palette.success.main, 0.12)
                              : alpha(theme.palette.warning.main, 0.12),
                          color: noti.type === 'news'
                            ? 'info.main'
                            : noti.type === 'discussion'
                              ? 'success.main'
                              : 'warning.main',
                          fontSize: '1rem'
                        }}>
                          {noti.type === 'news' ? <CampaignRounded fontSize="small" /> : noti.type === 'discussion' ? <QuestionAnswerRounded fontSize="small" /> : <SchoolRounded fontSize="small" />}
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
                          {noti.created_at ? new Date(noti.created_at).toLocaleString('de-DE', {
                            day: '2-digit',
                            month: '2-digit',
                            hour: '2-digit',
                            minute: '2-digit'
                          }) : ''}
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
                >
                  <Avatar
                    key={String(user?.id) || 'anonymous'}
                    src={getAvatarUrl(user?.avatar_url)}
                    sx={{ width: 32, height: 32, fontSize: '0.9rem', fontWeight: 600 }}
                  />
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
                <MenuItem component={RouterLink} to={user ? `/u/${user.id}` : "/login"} sx={{ py: 1.2 }}>
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
                <MenuItem onClick={() => logout()} sx={{ py: 1.2 }}>
                  <ListItemIcon>
                    <LogoutRounded fontSize="small" color="error" />
                  </ListItemIcon>
                  <Typography color="error" fontWeight={600}>Abmelden</Typography>
                </MenuItem>
              </Menu>
            </>
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
            backgroundColor: theme.palette.background.paper,
            backgroundImage: 'none',
          },
        }}
      >
        <Toolbar />
        {drawerContent}
      </Drawer>

      <Box component="main" sx={{ flexGrow: 1, p: 0, width: '100%', minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
        <Toolbar />
        <OfficeStatus user={user} />
        <Container maxWidth={fullBleed ? false : maxWidth} disableGutters sx={{ flexGrow: 1, p: fullBleed ? 0 : { xs: 2.5, md: 3 } }}>
          {children}
        </Container>
      </Box>
    </Box>
  );
};

const PageLoader = () => (
  <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', width: '100%' }}>
    <CircularProgress />
  </Box>
);

export { Sidebar, PageLoader };
