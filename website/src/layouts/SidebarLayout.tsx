import React from 'react';
import {
  AppBar,
  Avatar,
  Badge,
  Box,
  Container,
  Divider,
  Drawer,
  IconButton,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Toolbar,
  Tooltip,
  Typography,
  useMediaQuery,
} from '@mui/material';
import { alpha, useTheme } from '@mui/material/styles';
import SearchRounded from '@mui/icons-material/SearchRounded';
import NotificationsNoneRounded from '@mui/icons-material/NotificationsNoneRounded';
import Brightness4Rounded from '@mui/icons-material/Brightness4Rounded';
import Brightness7Rounded from '@mui/icons-material/Brightness7Rounded';
import MenuRounded from '@mui/icons-material/MenuRounded';
import MenuOpenRounded from '@mui/icons-material/MenuOpenRounded';
import HomeRounded from '@mui/icons-material/HomeRounded';
import ArticleRounded from '@mui/icons-material/ArticleRounded';
import PhotoLibraryRounded from '@mui/icons-material/PhotoLibraryRounded';
import ForumRounded from '@mui/icons-material/ForumRounded';
import { Link as RouterLink } from 'react-router-dom';
import { User } from '../api';
import { useThemeMode } from '../ThemeModeContext';

const drawerWidthOpen = 240;
const drawerWidthClosed = 72;
const mobileOverlayWidth = 280;

const navItems = [
  { label: 'Dashboard', href: '#dashboard', icon: <HomeRounded /> },
  { label: 'Beiträge', href: '#posts', icon: <ArticleRounded /> },
  { label: 'Galerie', href: '#gallery', icon: <PhotoLibraryRounded /> },
  { label: 'Forum', href: '/forum', icon: <ForumRounded />, isRoute: true },
];

type SidebarLayoutProps = {
  user: User;
  children: React.ReactNode;
  title?: string;
};

const SidebarLayout: React.FC<SidebarLayoutProps> = ({ user, children, title = 'Dashboard' }) => {
  const { mode, toggleMode } = useThemeMode();
  const theme = useTheme();
  const isMdUp = useMediaQuery(theme.breakpoints.up('md'));
  const [navOpen, setNavOpen] = React.useState<boolean>(isMdUp);

  React.useEffect(() => {
    setNavOpen(isMdUp);
  }, [isMdUp]);

  const toggleNav = () => setNavOpen((o) => !o);

  const NavList = ({
    labelsVisible,
    onItemClick,
  }: {
    labelsVisible: boolean;
    onItemClick?: () => void;
  }) => (
    <List sx={{ py: 0 }}>
      {navItems.map(({ label, href, icon, isRoute }) => (
        <Tooltip key={label} title={label} placement="right" disableHoverListener={labelsVisible}>
          <ListItemButton
            component={isRoute ? RouterLink : 'a'}
            to={isRoute ? href : undefined}
            href={!isRoute ? href : undefined}
            onClick={onItemClick}
            sx={[(theme) => ({
              minHeight: 48,
              borderRadius: 2,
              color: theme.palette.text.primary,
              '&:hover': {
                bgcolor:
                  alpha(theme.palette.primary.main, 0.08),
                ...theme.applyStyles("dark", {
                  bgcolor: alpha(theme.palette.primary.main, 0.15)
                })
              },
              ...theme.applyStyles("dark", {
                color: theme.palette.grey[100]
              })
            }), labelsVisible ? {
              justifyContent: 'flex-start'
            } : {
              justifyContent: 'center'
            }, labelsVisible ? {
              px: 2
            } : {
              px: 0
            }]}
          >
            <ListItemIcon
              sx={[(theme) => ({
                minWidth: 0,
                justifyContent: 'center',
                color: theme.palette.primary.main,
                ...theme.applyStyles("dark", {
                  color: theme.palette.primary.light
                })
              }), labelsVisible ? {
                mr: 1.5
              } : {
                mr: 0
              }]}
            >
              {icon}
            </ListItemIcon>
            {labelsVisible && <ListItemText primary={label} />}
          </ListItemButton>
        </Tooltip>
      ))}
    </List>
  );

  const SidebarContent = ({
    labelsVisible,
    withCloseButton,
    onClose,
  }: {
    labelsVisible: boolean;
    withCloseButton?: boolean;
    onClose?: () => void;
  }) => (
    <Box sx={{ display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0 }}>
      {withCloseButton && (
        <>
          <Toolbar sx={{ justifyContent: 'flex-end' }}>
            <Tooltip title="Menü schließen">
              <IconButton aria-label="close menu" onClick={onClose}>
                <MenuOpenRounded />
              </IconButton>
            </Tooltip>
          </Toolbar>
          <Divider />
        </>
      )}
      <Box sx={{ flex: 1, overflow: 'auto' }}>
        <NavList labelsVisible={labelsVisible} onItemClick={onClose} />
      </Box>
      <Divider />
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', py: 1 }}>
        <Tooltip title="Farbschema wechseln">
          <IconButton
            aria-label="toggle color mode"
            onClick={toggleMode}
            sx={[(theme) => ({
              border: `1px solid ${alpha(theme.palette.primary.main, mode === 'dark' ? 0.5 : 0.25)}`,
              transition: theme.transitions.create(['background-color', 'border-color'], {
                duration: theme.transitions.duration.shortest,
              }),
              '&:hover': {
                borderColor: alpha(theme.palette.primary.main, 0.6)
              }
            }), mode === 'dark' ? {
              bgcolor: alpha(theme.palette.primary.dark, 0.35)
            } : {
              bgcolor: alpha(theme.palette.primary.main, 0.1)
            }, mode === 'dark' ? {
              color: theme.palette.primary.contrastText
            } : {
              color: theme.palette.primary.main
            }, mode === 'dark' ? {
              '&:hover': {
                bgcolor: alpha(theme.palette.primary.dark, 0.55)
              }
            } : {
              '&:hover': {
                bgcolor: alpha(theme.palette.primary.main, 0.2)
              }
            }]}
          >
            {mode === 'dark' ? <Brightness7Rounded /> : <Brightness4Rounded />}
          </IconButton>
        </Tooltip>
      </Box>
    </Box>
  );

  return (
    <Box
      sx={{
        display: 'flex',
        minHeight: '100dvh',
        bgcolor: 'background.default',
        color: 'text.primary',
      }}
    >
    <AppBar
        position="fixed"
        color="primary"
        sx={(theme) => ({
          zIndex: theme.zIndex.drawer + 1,
          width: '100%',
          bgcolor: mode === 'dark'
            ? theme.palette.primary.dark
            : theme.palette.primary.main,
        })}
      >
        <Toolbar>
          <Tooltip
            title={
              isMdUp
                ? navOpen
                  ? 'Sidebar einklappen'
                  : 'Sidebar ausklappen'
                : navOpen
                  ? 'Menü schließen'
                  : 'Menü öffnen'
            }
          >
            <IconButton
              edge="start"
              color="inherit"
              aria-label="toggle sidebar"
              aria-expanded={navOpen}
              onClick={toggleNav}
              sx={{ mr: 1 }}
            >
              {navOpen ? <MenuOpenRounded /> : <MenuRounded />}
            </IconButton>
          </Tooltip>
          <Typography variant="h6" noWrap component="div">
            {title}
          </Typography>
          <Box sx={{ flexGrow: 1 }} />
          <Tooltip title="Suche">
            <IconButton aria-label="search" color="inherit">
              <SearchRounded />
            </IconButton>
          </Tooltip>
          <Tooltip title="Benachrichtigungen">
            <IconButton aria-label="notifications" sx={{ mx: 0.5 }} color="inherit">
              <Badge badgeContent={99} color="error">
                <NotificationsNoneRounded />
              </Badge>
            </IconButton>
          </Tooltip>
          <Tooltip title="Account">
            <IconButton aria-label="account" color="inherit">
              <Avatar sx={{ width: 28, height: 28 }}>
                {user.name ? user.name.charAt(0).toUpperCase() : '?'}
              </Avatar>
            </IconButton>
          </Tooltip>
        </Toolbar>
      </AppBar>
      <Drawer
        variant="permanent"
        sx={[(t) => ({
          display: { xs: 'none', md: 'block' },
          flexShrink: 0,
          whiteSpace: 'nowrap',
          '& .MuiDrawer-paper': {
            boxSizing: 'border-box',
            transition: t.transitions.create('width', {
              easing: t.transitions.easing.sharp,
              duration: navOpen ? t.transitions.duration.enteringScreen : t.transitions.duration.leavingScreen,
            }),
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden'
          }
        }), navOpen ? {
          width: {
            md: drawerWidthOpen
          }
        } : {
          width: {
            md: drawerWidthClosed
          }
        }, navOpen ? {
          '& .MuiDrawer-paper': {
            width: {
              md: drawerWidthOpen
            }
          }
        } : {
          '& .MuiDrawer-paper': {
            width: {
              md: drawerWidthClosed
            }
          }
        }]}
      >
        <Toolbar />
        <Divider />
        <SidebarContent labelsVisible={isMdUp && navOpen} />
      </Drawer>
      {!isMdUp && (
        <Drawer
          variant="temporary"
          open={navOpen}
          onClose={() => setNavOpen(false)}
          ModalProps={{ keepMounted: true }}
          slotProps={{
            paper: { sx: { width: mobileOverlayWidth } }
          }}
        >
          <SidebarContent labelsVisible withCloseButton onClose={() => setNavOpen(false)} />
        </Drawer>
      )}
      <Box component="main" sx={{ flexGrow: 1, p: { xs: 2, sm: 3 } }}>
        <Toolbar />
        <Container maxWidth="lg" disableGutters>
          {children}
        </Container>
      </Box>
    </Box>
  );
};
export default SidebarLayout;
