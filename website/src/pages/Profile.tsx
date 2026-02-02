import { useEffect, useState } from 'react';
import { useParams, Link as RouterLink } from 'react-router-dom';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Skeleton from '@mui/material/Skeleton';
import HistoryIcon from '@mui/icons-material/History';
import PostIcon from '@mui/icons-material/PostAdd';
import CommentIcon from '@mui/icons-material/QuestionAnswerRounded';
import ArchiveIcon from '@mui/icons-material/LibraryBooksRounded';
import MediaIcon from '@mui/icons-material/Collections';
import LockIcon from '@mui/icons-material/LockRounded';

import { useAuth } from '@lib/auth';
import Page from '@components/Page';
import Pagination from '@components/Pagination';
import { getPrograms, getUsersByUserIdActivities, getUsersByUserId } from '@lib/api';
import type { DtoPublicUserResponse as User, DtoActivityResponse as Activity } from '@lib/api';
import { getAvatarUrl } from '@lib/images';

export default function Profile() {
  const { userId } = useParams();
  const { user: me } = useAuth();

  const [u, setU] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [progs, setProgs] = useState<Record<string, string>>({});
  const [acts, setActs] = useState<Activity[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(1);

  useEffect(() => {
    getPrograms().then(({ data }) => {
      const m: Record<string, string> = {};
      data?.forEach(p => { if (p.id !== undefined) m[String(p.id)] = p.name || ''; });
      setProgs(m);
    });
  }, []);

  useEffect(() => {
    if (!userId) return;
    void Promise.resolve().then(() => setLoading(true));
    getUsersByUserId({ path: { userId } }).then(({ data }) => {
      setU(data || null);
      setLoading(false);
    });
  }, [userId]);

  useEffect(() => {
    if (!userId || (u?.private === 1 && String(me?.id) !== userId)) return;
    getUsersByUserIdActivities({ path: { userId }, query: { limit: 5, offset: (page - 1) * 5 } }).then(({ data, response }) => {
      setActs(data || []);
      setTotal(Math.ceil(parseInt(response.headers.get('X-Total-Count') || '0') / 5));
    });
  }, [userId, page, u, me]);

  if (loading) return <Page title="Profil"><Skeleton variant="rectangular" height={200} /></Page>;
  if (!u) return <Page title="Nicht gefunden"><Typography>Nutzer nicht gefunden</Typography></Page>;

  const isMe = String(me?.id) === userId;
  const priv = u.private === 1 && !isMe;

  return (
    <Page title={u.name || 'Profil'}>
      <Stack spacing={3}>
        <Paper variant="outlined" sx={{ p: 4, borderRadius: 4, display: 'flex', gap: 4, alignItems: 'center', flexWrap: 'wrap' }}>
          <Avatar src={getAvatarUrl(u.avatar_url)} sx={{ width: 100, height: 100, bgcolor: 'primary.main', fontSize: 40 }}>{u.name?.[0]}</Avatar>
          <Box sx={{ flex: 1 }}>
            <Typography variant="h3" fontWeight={700}>{u.name}</Typography>
            <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
              <Chip label={u.role?.toUpperCase()} size="small" color="primary" variant="outlined" />
              {u.active === 1 && <Chip label="Aktiv" size="small" color="success" variant="outlined" />}
            </Stack>
            {!priv && <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>{progs[String(u.program_id)] || 'Kein Studiengang'} · Dabei seit {new Date(u.created_at || '').toLocaleDateString()}</Typography>}
          </Box>
          {isMe && <Button component={RouterLink} to="/settings" variant="outlined">Bearbeiten</Button>}
        </Paper>

        <Paper variant="outlined" sx={{ p: 3, borderRadius: 4 }}>
          <Typography variant="h6" sx={{ mb: 2, display: 'flex', alignItems: 'center', gap: 1 }}><HistoryIcon /> Aktivitäten</Typography>
          {priv ? <Box sx={{ py: 4, textAlign: 'center' }}><LockIcon sx={{ fontSize: 40, opacity: 0.5 }} /><Typography>Profil ist privat</Typography></Box> : 
            <List>
              {acts.map((a, i) => (
                <ListItem key={a.id} divider={i < acts.length - 1} component={RouterLink} to={a.type?.includes('POST') || a.type?.includes('COMMENT') ? `/d/${a.target_id}` : (a.type?.includes('EXAM') ? `/archive/${a.target_id}` : `/events/${a.target_id}`)} sx={{ color: 'inherit', textDecoration: 'none' }}>
                  <ListItemIcon>
                    {a.type?.includes('POST') ? <PostIcon /> : a.type?.includes('COMMENT') ? <CommentIcon /> : a.type?.includes('EXAM') ? <ArchiveIcon /> : <MediaIcon />}
                  </ListItemIcon>
                  <ListItemText primary={a.target_name || a.type} secondary={new Date(a.created_at || '').toLocaleString()} />
                </ListItem>
              ))}
              {!acts.length && <Typography color="text.secondary" textAlign="center">Keine Aktivitäten</Typography>}
            </List>
          }
          <Pagination count={total} page={page} onChange={(_, p) => setPage(p)} />
        </Paper>
      </Stack>
    </Page>
  );
}
