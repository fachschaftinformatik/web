import React, { useEffect, useMemo, useState } from 'react';
import {
  Avatar,
  Box,
  Divider,
  Grid,
  IconButton,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
  Paper,
  Stack,
  Typography,
  alpha,
  useTheme,
} from '@mui/material';
import ChevronLeftRounded from '@mui/icons-material/ChevronLeftRounded';
import ChevronRightRounded from '@mui/icons-material/ChevronRightRounded';
import MoreVertRounded from '@mui/icons-material/MoreVertRounded';
import { Link as RouterLink } from 'react-router-dom';

import { Sidebar } from '@components/layout';
import { useAuth } from '@lib/auth';
import { newsDaten, type NewsItem } from '@routes/news/page';
import { FORUM_SEED_POSTS, FORUM_STORAGE_KEY } from '@routes/forum/page';

type ForumPostSummary = {
  id: string;
  title: string;
  author: string;
  createdAt: string;
};

const STORAGE_NEWS_KEY = 'custom-news';

const monthLabel = (date: Date) =>
  date.toLocaleDateString('de-DE', { month: 'long', year: 'numeric' });

const formatDate = (date: Date) =>
  date.toLocaleDateString('de-DE', { day: '2-digit', month: 'short', year: 'numeric' });

const formatForumDate = (iso: string) => {
  const d = new Date(iso);
  return isNaN(d.getTime()) ? 'k.A.' : d.toLocaleDateString('de-DE', { day: '2-digit', month: 'short' });
};

const buildCalendarDays = (date: Date) => {
  const start = new Date(date.getFullYear(), date.getMonth(), 1);
  const end = new Date(date.getFullYear(), date.getMonth() + 1, 0);

  const days: Array<{ day: number; date: Date }> = [];
  const offset = (start.getDay() + 6) % 7; // make Monday = 0
  for (let i = 0; i < offset; i++) {
    days.push({ day: 0, date: new Date() });
  }
  for (let d = 1; d <= end.getDate(); d++) {
    days.push({ day: d, date: new Date(date.getFullYear(), date.getMonth(), d) });
  }
  return days;
};

const loadNews = (): NewsItem[] => {
  if (typeof window === 'undefined') return newsDaten;
  try {
    const custom = JSON.parse(window.localStorage.getItem(STORAGE_NEWS_KEY) || '[]');
    return [...newsDaten, ...(Array.isArray(custom) ? custom : [])];
  } catch {
    return newsDaten;
  }
};

const loadForumPosts = (): ForumPostSummary[] => {
  if (typeof window === 'undefined') {
    return FORUM_SEED_POSTS.map((p) => ({
      id: p.id,
      title: p.title,
      author: p.author,
      createdAt: p.createdAt,
    }));
  }
  try {
    const stored = JSON.parse(window.localStorage.getItem(FORUM_STORAGE_KEY) || '[]');
    const source = Array.isArray(stored) && stored.length ? stored : FORUM_SEED_POSTS;
    return source.map((p: any) => ({
      id: String(p.id),
      title: String(p.title ?? 'Neuer Beitrag'),
      author: String(p.author ?? 'Unbekannt'),
      createdAt: String(p.createdAt ?? new Date().toISOString()),
    }));
  } catch {
    return FORUM_SEED_POSTS.map((p) => ({
      id: p.id,
      title: p.title,
      author: p.author,
      createdAt: p.createdAt,
    }));
  }
};

const NewsFeedPage: React.FC = () => {
  const { user } = useAuth();
  const theme = useTheme();
  const [currentMonth, setCurrentMonth] = useState(new Date());
  const [newsItems, setNewsItems] = useState<NewsItem[]>([]);
  const [forumPosts, setForumPosts] = useState<ForumPostSummary[]>([]);

  useEffect(() => {
    setNewsItems(
      loadNews().sort(
        (a, b) => new Date(b.date).getTime() - new Date(a.date).getTime(),
      ),
    );
    setForumPosts(
      loadForumPosts().sort(
        (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
      ),
    );
  }, []);

  const days = useMemo(() => buildCalendarDays(currentMonth), [currentMonth]);
  const today = new Date();

  const latestNews = newsItems.slice(0, 4);
  const latestForum = forumPosts.slice(0, 5);

  const calendarBg = theme.palette.mode === 'light'
    ? alpha(theme.palette.success.light, 0.18)
    : alpha(theme.palette.success.dark, 0.25);

  return (
    <Sidebar user={user} title="Startseite">
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          mt: { xs: 1, md: 1.5 },
          mb: { xs: 1.5, md: 2 },
        }}
      >
        <Box
          component="img"
          src="/FSV-Logo.png"
          alt="FSV Logo"
          sx={{
            width: { xs: 500, sm: 720, md: 500 },
            height: 'auto',
            objectFit: 'contain',
          }}
        />
      </Box>
      <Box
        sx={{
          display: 'grid',
          rowGap: 2.5,
          columnGap: { xs: 2, md: 4 },
          gridTemplateColumns: {
            xs: '1fr',
            sm: '1fr 1.4fr',
            md: '1fr 1.45fr',
          },
          alignItems: 'stretch',
          gridAutoRows: '1fr',
          px: { xs: 1, sm: 1.5, md: 2.5 },
          py: { xs: 1.25, md: 2 },
          maxWidth: 1380,
          mx: 'auto',
        }}
      >
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, height: '100%' }}>
          <Stack spacing={2} sx={{ height: '100%' }}>
            <Paper
              variant="outlined"
              sx={{
                p: 2,
                borderRadius: 3,
                background: calendarBg,
              }}
            >
              <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1 }}>
                <IconButton onClick={() => setCurrentMonth((prev) => new Date(prev.getFullYear(), prev.getMonth() - 1, 1))}>
                  <ChevronLeftRounded />
                </IconButton>
                <Typography variant="subtitle1" fontWeight={700}>
                  {monthLabel(currentMonth)}
                </Typography>
                <IconButton onClick={() => setCurrentMonth((prev) => new Date(prev.getFullYear(), prev.getMonth() + 1, 1))}>
                  <ChevronRightRounded />
                </IconButton>
              </Stack>

              <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: 0.5, textAlign: 'center', mb: 1 }}>
                {['Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa', 'So'].map((d) => (
                  <Typography key={d} variant="caption" sx={{ fontWeight: 700, opacity: 0.7 }}>
                    {d}
                  </Typography>
                ))}
              </Box>

              <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: 0.75 }}>
                {days.map(({ day, date }, idx) => {
                  const isToday =
                    day > 0 &&
                    date.getDate() === today.getDate() &&
                    date.getMonth() === today.getMonth() &&
                    date.getFullYear() === today.getFullYear();
                  return (
                    <Box
                      key={`${day}-${idx}`}
                      sx={{
                        aspectRatio: '1 / 1',
                        borderRadius: 2,
                        bgcolor: isToday ? alpha(theme.palette.success.main, 0.25) : 'transparent',
                        color: isToday ? theme.palette.success.dark : 'text.primary',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        border: isToday ? `1px solid ${alpha(theme.palette.success.main, 0.6)}` : '1px dashed transparent',
                        opacity: day ? 1 : 0,
                      }}
                    >
                      <Typography variant="subtitle2" fontWeight={600}>
                        {day || ''}
                      </Typography>
                    </Box>
                  );
                })}
              </Box>
            </Paper>

            <Paper variant="outlined" sx={{ p: 2, borderRadius: 3 }}>
              <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 1 }}>
                <Typography variant="h6" fontWeight={700}>
                  Forum
                </Typography>
                <IconButton size="small" component={RouterLink} to="/forum">
                  <MoreVertRounded />
                </IconButton>
              </Stack>
              <Divider sx={{ mb: 1 }} />
              <List dense sx={{ pr: 1 }}>
                {latestForum.map((post) => (
                  <React.Fragment key={post.id}>
                    <ListItem
                      disableGutters
                      component={RouterLink}
                      to={`/forum#${post.id}`}
                      sx={{
                        alignItems: 'flex-start',
                        px: 0,
                        py: 0.5,
                        color: 'inherit',
                        textDecoration: 'none',
                        '&:hover .MuiTypography-root': { color: 'primary.main' },
                      }}
                    >
                      <ListItemAvatar>
                        <Avatar sx={{ width: 32, height: 32 }}>
                          {post.author?.charAt(0)?.toUpperCase() || '?'}
                        </Avatar>
                      </ListItemAvatar>
                      <ListItemText
                        primary={
                          <Stack direction="row" spacing={1} alignItems="center">
                            <Typography variant="body1" fontWeight={600} noWrap>
                              {post.title}
                            </Typography>
                          </Stack>
                        }
                        secondary={
                          <Typography variant="caption" color="text.secondary">
                            {post.author} · {formatForumDate(post.createdAt)}
                          </Typography>
                        }
                      />
                    </ListItem>
                    <Divider sx={{ my: 0.75 }} />
                  </React.Fragment>
                ))}
                {!latestForum.length && (
                  <Typography variant="body2" color="text.secondary">
                    Noch keine Beiträge.
                  </Typography>
                )}
              </List>
            </Paper>
          </Stack>
        </Box>

        <Box>
          <Paper
            variant="outlined"
            sx={{
              p: { xs: 2, md: 3 },
              borderRadius: 3,
              borderLeft: { md: `4px solid ${alpha(theme.palette.success.main, 0.35)}` },
              boxShadow: { md: `-6px 0 12px -8px ${alpha(theme.palette.common.black, 0.12)}` },
              height: '100%',
              display: 'flex',
              flexDirection: 'column',
            }}
          >
            <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 2 }}>
              <Typography variant="h4" fontWeight={800}>
                Neuigkeiten
              </Typography>
              <IconButton component={RouterLink} to="/news">
                <MoreVertRounded />
              </IconButton>
            </Stack>
            <Stack spacing={2.25}>
              {latestNews.map((item) => (
                <Paper
                  key={item.id}
                  variant="outlined"
                  component={RouterLink}
                  to="/news"
                  sx={{
                    display: 'flex',
                    gap: 1.75,
                    p: { xs: 1.4, md: 1.6 },
                    borderRadius: 2,
                    textDecoration: 'none',
                    color: 'inherit',
                    transition: 'transform 120ms ease, box-shadow 120ms ease',
                    '&:hover': {
                      transform: 'translateY(-2px)',
                      boxShadow: 4,
                    },
                  }}
                >
                  <Box
                    component="img"
                    src={item.image}
                    alt={item.title}
                    sx={{
                      width: { xs: 140, md: 170 },
                      height: { xs: 92, md: 110 },
                      borderRadius: 1.5,
                      objectFit: 'cover',
                      flexShrink: 0,
                    }}
                  />
                  <Stack spacing={0.5}>
                    <Typography variant="caption" color="text.secondary">
                      {formatDate(new Date(item.date))}
                    </Typography>
                    <Typography variant="h6" fontWeight={800} lineHeight={1.2}>
                      {item.title}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" noWrap>
                      {item.summary || 'Mehr lesen...'}
                    </Typography>
                  </Stack>
                </Paper>
              ))}
              {!latestNews.length && (
                <Typography variant="body2" color="text.secondary">
                  Noch keine News vorhanden.
                </Typography>
              )}
            </Stack>
          </Paper>
        </Box>
      </Box>
    </Sidebar>
  );
};

export default NewsFeedPage;
