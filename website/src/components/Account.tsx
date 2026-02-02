import { useState } from 'react';
import { Link as RouterLink } from 'react-router-dom';
import Avatar from '@mui/material/Avatar';
import IconButton from '@mui/material/IconButton';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import Divider from '@mui/material/Divider';
import Typography from '@mui/material/Typography';
import Tooltip from '@mui/material/Tooltip';
import { alpha } from '@mui/material/styles';
import PersonRounded from '@mui/icons-material/PersonRounded';
import SettingsRounded from '@mui/icons-material/SettingsRounded';
import LogoutRounded from '@mui/icons-material/LogoutRounded';
import { useAuth } from '@lib/auth';
import { getAvatarUrl } from '@lib/images';

export default function Account() {
  const { user, logout } = useAuth();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);
  const handleClose = () => setAnchorEl(null);
  if (!user) return null;

  return (
    <>
      <Tooltip title="Account">
        <IconButton onClick={(e) => setAnchorEl(e.currentTarget)} size="small">
          <Avatar key={String(user.id)} src={getAvatarUrl(user.avatar_url)} sx={{ width: 32, height: 32, fontSize: '0.9rem', fontWeight: 600 }} />
        </IconButton>
      </Tooltip>
      <Menu
        anchorEl={anchorEl}
        id="account-menu"
        open={open}
        onClose={handleClose}
        onClick={handleClose}
        transformOrigin={{ horizontal: 'right', vertical: 'top' }}
        anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
        slotProps={{
          paper: {
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
          }
        }}
      >
        <MenuItem component={RouterLink} to={`/u/${user.id}`} sx={{ py: 1.2 }}>
          <ListItemIcon><PersonRounded fontSize="small" /></ListItemIcon>
          Mein Profil
        </MenuItem>
        <MenuItem component={RouterLink} to="/settings" sx={{ py: 1.2 }}>
          <ListItemIcon><SettingsRounded fontSize="small" /></ListItemIcon>
          Einstellungen
        </MenuItem>
        <Divider />
        <MenuItem onClick={() => logout()} sx={{ py: 1.2 }}>
          <ListItemIcon><LogoutRounded fontSize="small" color="error" /></ListItemIcon>
          <Typography color="error" fontWeight={600}>Abmelden</Typography>
        </MenuItem>
      </Menu>
    </>
  );
}
