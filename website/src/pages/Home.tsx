import { useEffect, useMemo, useState } from 'react';
import { useNavigate, Link as RouterLink } from 'react-router-dom';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Skeleton from '@mui/material/Skeleton';
import { useTheme, alpha } from '@mui/material/styles';
import GroupsRounded from '@mui/icons-material/GroupsRounded';
import InfoRounded from '@mui/icons-material/InfoRounded';
import Timeline from '@mui/lab/Timeline';
import TimelineItem from '@mui/lab/TimelineItem';
import TimelineSeparator from '@mui/lab/TimelineSeparator';
import TimelineConnector from '@mui/lab/TimelineConnector';
import TimelineContent from '@mui/lab/TimelineContent';
import TimelineDot from '@mui/lab/TimelineDot';
import TimelineOppositeContent from '@mui/lab/TimelineOppositeContent';

import Page from '@components/Page';
import { getDiscussions, getEvents } from '@lib/api';
import type { DtoDiscussionPostResponse as Post, DtoEventResponse as Event } from '@lib/api';
import { getSizedImageUrl, getImageSrcSet, getAvatarUrl } from '@lib/images';

const fmtDate = (d: string) => new Date(d).toLocaleDateString('de-DE', { day: '2-digit', month: 'short' });

export default function Home() {
  const navigate = useNavigate();
  const theme = useTheme();
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState(0);
  const [news, setNews] = useState<Post[]>([]);
  const [posts, setPosts] = useState<Post[]>([]);
  const [apiEvents, setApiEvents] = useState<Post[]>([]);
  const [cur, setCur] = useState(0);

  useEffect(() => {
    Promise.all([
      getDiscussions({ query: { type: 'news', limit: 3 } }),
      getDiscussions({ query: { type: 'discussion', limit: 3 } }),
      getDiscussions({ query: { type: 'event', limit: 20 } }),
      getEvents()
    ]).then(([n, p, e, evs]) => {
      if (n.data) setNews(n.data);
      if (p.data) setPosts(p.data);
      if (e.data) setApiEvents(e.data);
      if (evs.data) setEvents(evs.data);
      setLoading(false);
    });
  }, []);

  useEffect(() => {
    if (events.length > 1) {
      const itv = setInterval(() => setCur(c => (c + 1) % events.length), 5000);
      return () => clearInterval(itv);
    }
  }, [events.length]);

  const agenda = useMemo(() => {
    const now = new Date().getTime();
    return apiEvents
      .map(p => ({
        id: String(p.id),
        title: p.title || '',
        date: new Date(p.event_date || p.created_at || ''),
        location: p.location,
        cat: (p.tags as string[])?.find(t => ['Treffen', 'Workshop', 'Party', 'Info'].includes(t)) || 'Info'
      }))
      .filter(e => e.date.getTime() >= now)
      .sort((a, b) => a.date.getTime() - b.date.getTime())
      .slice(0, 5);
  }, [apiEvents]);

  const getCatColor = (c: string) => {
    switch (c) {
      case 'Treffen': return theme.palette.success.main;
      case 'Workshop': return theme.palette.info.main;
      case 'Party': return theme.palette.error.main;
      case 'Info': return theme.palette.warning.main;
      default: return theme.palette.primary.main;
    }
  };

  const PostCard = ({ p }: { p: Post }) => (
    <Box component={RouterLink} to={`/d/${p.id}`} sx={{ ...theme.custom.newsCardLink, bgcolor: alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.08 : 0.04) }}>
      <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 0.5 }}>
        <Avatar src={getAvatarUrl(p.user_avatar_url)} sx={{ width: 18, height: 18, fontSize: "0.6rem" }}>{p.user_name?.[0]}</Avatar>
        <Typography variant="caption" color="text.secondary">von <b>{p.user_name}</b> · {fmtDate(p.created_at || '')}</Typography>
      </Stack>
      <Typography variant="subtitle1" fontWeight={700}>{p.title}</Typography>
      <Typography variant="body2" color="text.secondary" noWrap>{p.body}</Typography>
    </Box>
  );

  return (
    <Page title="Startseite" maxWidth="xl">
      <Stack spacing={3}>
        <Box sx={{ position: 'relative', height: { xs: 240, md: 360 }, borderRadius: 6, overflow: 'hidden', border: '1px solid', borderColor: 'divider' }}>
          {loading ? <Skeleton variant="rectangular" sx={{ position: 'absolute', inset: 0 }} /> : events.map((ev, i) => (
            <Box key={String(ev.id)} sx={{ position: 'absolute', inset: 0, transition: 'opacity 1s', opacity: cur === i ? 1 : 0 }}>
              <Box component="img" src={getSizedImageUrl(`/api/v1/events/${ev.id}/cover`, 1600)} srcSet={getImageSrcSet(`/api/v1/events/${ev.id}/cover`)} sx={{ width: '100%', height: '100%', objectFit: 'cover' }} alt="" />
              <Box sx={{ position: 'absolute', inset: 0, background: `linear-gradient(to right, ${alpha(theme.palette.background.default, 0.95)}, transparent)` }} />
            </Box>
          ))}
          <Box sx={{ position: 'relative', zIndex: 1, height: '100%', px: { xs: 4, md: 8 }, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
            <Typography variant="overline" color="primary" fontWeight={700}>fsInformatik</Typography>
            <Typography variant="h1" sx={{ fontSize: { xs: '2rem', md: '4rem' }, mb: 2 }}>{events[cur]?.title || 'FSV Informatik'}</Typography>
            <Stack direction="row" spacing={2}>
              <Button component={RouterLink} to="/discussions" variant="contained">Beiträge</Button>
              <Button component={RouterLink} to="/events" variant="outlined">Galerie</Button>
            </Stack>
          </Box>
        </Box>

        <Box sx={{ display: 'grid', gap: 4, gridTemplateColumns: { xs: '1fr', md: '7fr 5fr' } }}>
          <Paper variant="outlined" sx={{ p: 2.5, borderRadius: 4 }}>
            <Stack direction="row" spacing={2} sx={{ mb: 2 }}>
              <Typography variant="h5" onClick={() => setTab(0)} sx={{ cursor: 'pointer', fontWeight: 700, color: tab === 0 ? 'text.primary' : 'text.secondary' }}>Ankündigungen</Typography>
              <Typography variant="h5" sx={{ color: 'divider' }}>|</Typography>
              <Typography variant="h5" onClick={() => setTab(1)} sx={{ cursor: 'pointer', fontWeight: 700, color: tab === 1 ? 'text.primary' : 'text.secondary' }}>Beiträge</Typography>
            </Stack>
            <Stack spacing={1.5}>
              {loading ? [1, 2, 3].map(i => <Skeleton key={i} variant="rectangular" height={80} sx={{ borderRadius: 3 }} />) : 
                (tab === 0 ? news : posts).map(p => <PostCard key={String(p.id)} p={p} />)}
            </Stack>
            <Button component={RouterLink} to="/discussions" variant="contained" sx={{ mt: 3 }}>Alle anzeigen</Button>
          </Paper>

          <Box>
            <Typography variant="h5" fontWeight={700} textAlign="center" sx={{ mb: 2 }}>Agenda</Typography>
            <Timeline position="alternate">
              {loading ? [1, 2, 3].map(i => <TimelineItem key={i}><TimelineSeparator><TimelineDot /><TimelineConnector /></TimelineSeparator><TimelineContent><Skeleton /></TimelineContent></TimelineItem>) :
                agenda.map(e => (
                  <TimelineItem key={e.id} onClick={() => navigate(`/d/${e.id}`)} sx={{ cursor: 'pointer' }}>
                    <TimelineOppositeContent color="text.secondary" variant="body2">{fmtDate(e.date.toISOString())}</TimelineOppositeContent>
                    <TimelineSeparator>
                      <TimelineConnector />
                      <TimelineDot sx={{ bgcolor: getCatColor(e.cat) }}>
                        {e.cat === 'Treffen' ? <GroupsRounded sx={{ fontSize: 18 }} /> : <InfoRounded sx={{ fontSize: 18 }} />}
                      </TimelineDot>
                      <TimelineConnector />
                    </TimelineSeparator>
                    <TimelineContent><Typography variant="subtitle2" fontWeight={700}>{e.title}</Typography></TimelineContent>
                  </TimelineItem>
                ))}
            </Timeline>
          </Box>
        </Box>
      </Stack>
    </Page>
  );
}
