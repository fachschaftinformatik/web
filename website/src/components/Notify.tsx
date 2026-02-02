import { useState, useEffect, useCallback } from 'react';
import Box from '@mui/material/Box';
import IconButton from '@mui/material/IconButton';
import Badge from '@mui/material/Badge';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import List from '@mui/material/List';
import ListItemIcon from '@mui/material/ListItemIcon';
import Avatar from '@mui/material/Avatar';
import Tooltip from '@mui/material/Tooltip';
import { useTheme, alpha } from '@mui/material/styles';
import NotificationsNoneRounded from '@mui/icons-material/NotificationsNoneRounded';
import DoneAllRounded from '@mui/icons-material/DoneAllRounded';
import CampaignRounded from '@mui/icons-material/CampaignRounded';
import QuestionAnswerRounded from '@mui/icons-material/QuestionAnswerRounded';
import SchoolRounded from '@mui/icons-material/SchoolRounded';
import { useNavigate } from 'react-router-dom';
import type { DtoNotificationResponse as Notification } from '@lib/api/types.gen';
import { getAuthNotifications, putAuthNotificationsIdRead, putAuthNotificationsReadAll } from '@lib/api';

export default function Notify() {
  const theme = useTheme();
  const navigate = useNavigate();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const open = Boolean(anchorEl);
  const unreadCount = notifications.filter(n => !n.read).length;

  const fetchNotifications = useCallback(async () => {
    try {
      const { data } = await getAuthNotifications();
      if (data) setNotifications(data as Notification[]);
    } catch (err) {
      console.error("Failed to fetch notifications", err);
    }
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchNotifications();
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

  const handleClose = () => setAnchorEl(null);

  return (
    <>
      <Tooltip title="Benachrichtigungen">
        <IconButton color="inherit" size="large" onClick={(e) => setAnchorEl(e.currentTarget)}>
          <Badge badgeContent={unreadCount} color="error">
            <NotificationsNoneRounded />
          </Badge>
        </IconButton>
      </Tooltip>

      <Menu
        anchorEl={anchorEl}
        id="notifications-menu"
        open={open}
        onClose={handleClose}
        transformOrigin={{ horizontal: 'right', vertical: 'top' }}
        anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
        slotProps={{
          paper: {
            elevation: 0,
            sx: {
              mt: 1.5, width: 340, maxHeight: 500, borderRadius: '12px',
              filter: 'drop-shadow(0px 4px 20px rgba(0,0,0,0.15))',
              border: '1px solid', borderColor: 'divider', bgcolor: 'background.paper',
              backdropFilter: 'blur(10px)', overflowY: 'auto'
            },
          }
        }}
      >
        <Box sx={{ position: 'sticky', top: 0, bgcolor: 'background.paper', zIndex: 2, borderBottom: '1px solid', borderColor: 'divider', minHeight: 56, display: 'flex', alignItems: 'center' }}>
          <Box sx={{ px: 2, py: 1, width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <Typography variant="subtitle1" fontWeight={700}>Benachrichtigungen</Typography>
            <Button size="small" startIcon={<DoneAllRounded fontSize="small" />} onClick={handleMarkAllRead} disabled={unreadCount === 0} sx={{ textTransform: 'none', fontSize: '0.75rem', py: 0.5 }}>Alle gelesen</Button>
          </Box>
        </Box>
        <List sx={{ p: 0 }}>
          {notifications.map((noti) => (
            <MenuItem
              key={String(noti.id)}
              onClick={() => {
                if (noti.id) handleMarkRead(String(noti.id));
                if (noti.link) navigate(noti.link);
                handleClose();
              }}
              sx={{
                py: 2, px: 2, borderBottom: '1px solid', borderColor: alpha(theme.palette.divider, 0.05),
                whiteSpace: 'normal', bgcolor: noti.read ? 'transparent' : alpha(theme.palette.primary.main, 0.04),
                '&:hover': { bgcolor: alpha(theme.palette.primary.main, 0.08) }
              }}
            >
              <ListItemIcon sx={{ minWidth: 44 }}>
                <Avatar sx={{
                  width: 32, height: 32,
                  bgcolor: noti.type === 'news' ? alpha(theme.palette.info.main, 0.12) : noti.type === 'discussion' ? alpha(theme.palette.success.main, 0.12) : alpha(theme.palette.warning.main, 0.12),
                  color: noti.type === 'news' ? 'info.main' : noti.type === 'discussion' ? 'success.main' : 'warning.main',
                  fontSize: '1rem'
                }}>
                  {noti.type === 'news' ? <CampaignRounded fontSize="small" /> : noti.type === 'discussion' ? <QuestionAnswerRounded fontSize="small" /> : <SchoolRounded fontSize="small" />}
                </Avatar>
              </ListItemIcon>
              <Box sx={{ flexGrow: 1 }}>
                <Typography variant="body2" sx={{ fontWeight: noti.read ? 600 : 700, mb: 0.2 }}>{noti.title}</Typography>
                <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.5, lineHeight: 1.3 }}>{noti.message}</Typography>
                <Typography variant="caption" sx={{ color: 'text.disabled', fontSize: '0.7rem' }}>
                  {noti.created_at ? new Date(noti.created_at).toLocaleString('de-DE', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' }) : ''}
                </Typography>
              </Box>
              {!noti.read && <Box sx={{ width: 8, height: 8, borderRadius: '50%', bgcolor: 'primary.main', ml: 1, flexShrink: 0 }} />}
            </MenuItem>
          ))}
        </List>
        {notifications.length === 0 && <Box sx={{ p: 4, textAlign: 'center' }}><Typography variant="body2" color="text.secondary">Keine neuen Benachrichtigungen</Typography></Box>}
      </Menu>
    </>
  );
}
