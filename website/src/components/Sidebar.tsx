import { Link as RouterLink } from 'react-router-dom';
import Box from '@mui/material/Box';
import Drawer from '@mui/material/Drawer';
import List from '@mui/material/List';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Tooltip from '@mui/material/Tooltip';
import IconButton from '@mui/material/IconButton';
import Toolbar from '@mui/material/Toolbar';
import { useTheme, alpha } from '@mui/material/styles';
import DashboardRounded from '@mui/icons-material/DashboardRounded';
import QuestionAnswerRounded from '@mui/icons-material/QuestionAnswerRounded';
import CollectionsRounded from '@mui/icons-material/CollectionsRounded';
import LibraryBooksRounded from '@mui/icons-material/LibraryBooksRounded';
import PeopleRounded from '@mui/icons-material/PeopleRounded';
import MailRounded from '@mui/icons-material/MailRounded';
import Brightness4Rounded from '@mui/icons-material/Brightness4Rounded';
import Brightness7Rounded from '@mui/icons-material/Brightness7Rounded';
import { useThemeMode } from '@lib/theme';
import { useAuth } from '@lib/auth';
import { putAuthMe } from '@lib/api';
import type { DtoUserResponse as User } from '@lib/api/types.gen';
import { NAV_ITEMS } from '@internals/data';

const drawerWidthOpen = 240;
const drawerWidthClosed = 72;

const iconMap: Record<string, React.ReactNode> = {
  dashboard: <DashboardRounded />,
  discussions: <QuestionAnswerRounded />,
  gallery: <CollectionsRounded />,
  archive: <LibraryBooksRounded />,
  team: <PeopleRounded />,
  contact: <MailRounded />,
};

interface SidebarProps {
  open: boolean;
  mobileOpen: boolean;
  onMobileClose: () => void;
  isMdUp: boolean;
}

export default function Sidebar({ open, mobileOpen, onMobileClose, isMdUp }: SidebarProps) {
  const theme = useTheme();
  const { mode, setPreference } = useThemeMode();
  const { user, login } = useAuth();

  const handleThemeToggle = async () => {
    const nextMode = mode === 'light' ? 'dark' : 'light';
    setPreference(nextMode);
    if (user) {
      try {
        const res = await putAuthMe({ body: { name: user.name || "", program_id: String(user.program_id), theme: nextMode } });
        if (res.data) login(res.data as User, localStorage.getItem('fs_remember_flag') === 'true');
      } catch (err) { console.error("Theme sync failed", err); }
    }
  };

  const toggleBtnStyle = {
    border: '1px solid ' + alpha(theme.palette.primary.main, mode === 'dark' ? 0.5 : 0.25),
    transition: theme.transitions.create(['background-color', 'border-color'], { duration: theme.transitions.duration.shortest }),
    '&:hover': {
      borderColor: alpha(theme.palette.primary.main, 0.6),
      bgcolor: mode === 'dark' ? alpha(theme.palette.primary.dark, 0.55) : alpha(theme.palette.primary.main, 0.2)
    },
    bgcolor: mode === 'dark' ? alpha(theme.palette.primary.dark, 0.35) : alpha(theme.palette.primary.main, 0.1),
    color: mode === 'dark' ? theme.palette.primary.contrastText : theme.palette.primary.main
  };

  const navOpen = isMdUp ? open : mobileOpen;

  const content = (
    <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%', overflowY: 'auto', '&::-webkit-scrollbar': { display: 'none' }, scrollbarWidth: 'none', msOverflowStyle: 'none' }}>
      <List sx={{ pt: 1, flexGrow: 1 }}>
        {NAV_ITEMS.map((item) => (
          <Tooltip key={item.label} title={navOpen ? '' : item.label} placement="right">
            <ListItemButton component={item.isRoute ? RouterLink : 'a'} to={item.isRoute ? item.href : undefined} href={!item.isRoute ? item.href : undefined} sx={{ minHeight: 48, justifyContent: navOpen ? 'initial' : 'center', px: 2.5 }}>
              <ListItemIcon sx={{ minWidth: 0, mr: navOpen ? 3 : 'auto', justifyContent: 'center' }}>{iconMap[item.icon]}</ListItemIcon>
              <ListItemText primary={item.label} sx={{ opacity: navOpen ? 1 : 0 }} />
            </ListItemButton>
          </Tooltip>
        ))}
      </List>
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', py: 1, mb: 1 }}>
        <Tooltip title="Farbschema wechseln">
          <IconButton aria-label="toggle color mode" onClick={handleThemeToggle} sx={toggleBtnStyle}>{mode === 'dark' ? <Brightness7Rounded /> : <Brightness4Rounded />}</IconButton>
        </Tooltip>
      </Box>
    </Box>
  );

  return (
    <Drawer
      variant={isMdUp ? "permanent" : "temporary"} open={navOpen} onClose={onMobileClose}
      sx={{
        width: navOpen ? drawerWidthOpen : drawerWidthClosed, flexShrink: 0, whiteSpace: 'nowrap', boxSizing: 'border-box',
      }}
      slotProps={{
        paper: {
          sx: {
            width: navOpen ? drawerWidthOpen : drawerWidthClosed,
            transition: theme.transitions.create('width', {
              easing: theme.transitions.easing.sharp,
              duration: theme.transitions.duration.enteringScreen,
            }),
            overflowX: 'hidden',
            backgroundColor: theme.palette.background.paper,
            backgroundImage: 'none',
          },
        }
      }}
    >
      <Toolbar />
      {content}
    </Drawer>
  );
}
