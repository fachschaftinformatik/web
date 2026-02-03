import React, { useState } from 'react';
import { useNavigate, Link as RouterLink } from 'react-router-dom';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import Avatar from '@mui/material/Avatar';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import { alpha, useTheme } from '@mui/material/styles';
import ThumbUpOutlined from '@mui/icons-material/ThumbUpOutlined';
import ThumbDownOutlined from '@mui/icons-material/ThumbDownOutlined';
import MoreVertIcon from '@mui/icons-material/MoreVert';
import EditIcon from '@mui/icons-material/Edit';
import PinIcon from '@mui/icons-material/PushPinOutlined';
import DelIcon from '@mui/icons-material/DeleteOutline';

import { getAvatarUrl } from '@lib/images';
import type { DtoDiscussionPostResponse as Post, DtoUserResponse as User } from '@lib/api';

interface PostItemProps {
  p: Post;
  user: User | null;
  onVote: (id: string, vote: number) => void;
  onDelete: (id: string) => void;
}

const fmtDate = (d: string) => new Date(d).toLocaleDateString('de-DE', { day: '2-digit', month: 'short' });

export default function PostItem({ p, user, onVote, onDelete }: PostItemProps) {
  const navigate = useNavigate();
  const theme = useTheme();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  const handleOpen = (e: React.MouseEvent<HTMLElement>) => {
    e.stopPropagation();
    setAnchorEl(e.currentTarget);
  };
  const handleClose = () => setAnchorEl(null);

  return (
    <Paper 
      onClick={() => navigate(`/d/${p.id}`)} 
      variant="outlined" 
      sx={{ 
        p: 2, 
        borderRadius: 3, 
        cursor: 'pointer', 
        '&:hover': { borderColor: 'primary.main' }, 
        bgcolor: p.type === 'news' ? alpha(theme.palette.primary.main, 0.05) : 'background.paper',
        height: 140,
        display: 'flex',
        flexDirection: 'column'
      }}
    >
      <Stack direction="row" spacing={2} sx={{ height: '100%' }}>
        <Stack alignItems="center">
          <IconButton 
            size="small" 
            onClick={e => { e.stopPropagation(); onVote(String(p.id), p.user_vote === 1 ? 0 : 1); }} 
            color={p.user_vote === 1 ? "primary" : "default"}
          >
            <ThumbUpOutlined fontSize="small" />
          </IconButton>
          <Typography variant="subtitle2">{p.votes}</Typography>
          <IconButton 
            size="small" 
            onClick={e => { e.stopPropagation(); onVote(String(p.id), p.user_vote === -1 ? 0 : -1); }} 
            color={p.user_vote === -1 ? "primary" : "default"}
          >
            <ThumbDownOutlined fontSize="small" />
          </IconButton>
        </Stack>
        <Box sx={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column' }}>
          <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
            <Typography variant="h6" fontWeight={700} sx={{ 
              display: '-webkit-box',
              WebkitLineClamp: 1,
              WebkitBoxOrient: 'vertical',
              overflow: 'hidden',
              wordBreak: 'break-word'
            }}>
              {p.title} {p.pinned === 1 && <PinIcon fontSize="small" color="primary" sx={{ verticalAlign: 'middle', ml: 0.5 }} />}
            </Typography>
            {(user?.role === 'admin' || user?.id === p.user_id) && (
              <>
                <IconButton size="small" onClick={handleOpen} sx={{ ml: 1, flexShrink: 0 }}><MoreVertIcon fontSize="small" /></IconButton>
                <Menu anchorEl={anchorEl} open={open} onClose={handleClose} onClick={e => e.stopPropagation()}>
                  <MenuItem onClick={() => { handleClose(); navigate(`/d/${p.id}/edit`); }} sx={{ gap: 1.5 }}>
                    <EditIcon fontSize="small" /> Bearbeiten
                  </MenuItem>
                  <MenuItem onClick={() => { handleClose(); onDelete(String(p.id!)); }} sx={{ gap: 1.5, color: 'error.main' }}>
                    <DelIcon fontSize="small" /> Löschen
                  </MenuItem>
                </Menu>
              </>
            )}
          </Stack>
          <Typography variant="body2" color="text.secondary" sx={{ 
            mb: 1,
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
            wordBreak: 'break-word',
            flexGrow: 1,
            lineHeight: '1.4em',
            maxHeight: '2.8em'
          }}>
            {p.body}
          </Typography>
          <Stack direction="row" spacing={1} alignItems="center" sx={{ mt: 'auto' }}>
            <Avatar src={getAvatarUrl(p.user_avatar_url)} sx={{ width: 20, height: 20 }} />
            <Typography variant="caption" noWrap sx={{ overflow: 'hidden', textOverflow: 'ellipsis' }}>
              von <Typography component={RouterLink} to={`/u/${p.user_id}`} onClick={(e: React.MouseEvent) => e.stopPropagation()} variant="caption" sx={{ fontWeight: 700, color: 'inherit', textDecoration: 'none', '&:hover': { textDecoration: 'underline' } }}>{p.user_name}</Typography> · {p.comment_count} {p.comment_count === 1 ? 'Kommentar' : 'Kommentare'} {p.created_at && ` · ${fmtDate(p.created_at)}`}
            </Typography>
          </Stack>
        </Box>
      </Stack>
    </Paper>
  );
}
