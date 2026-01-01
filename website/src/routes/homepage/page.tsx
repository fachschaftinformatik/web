import React, { useEffect, useMemo, useState } from 'react';
import {
  Avatar,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  MenuItem,
  Paper,
  Stack,
  TextField,
  Typography,
  alpha,
  useTheme,
} from '@mui/material';
import AddRounded from '@mui/icons-material/AddRounded';
import ArrowForwardRounded from '@mui/icons-material/ArrowForwardRounded';
import CalendarMonthRounded from '@mui/icons-material/CalendarMonthRounded';
import CampaignRounded from '@mui/icons-material/CampaignRounded';
import ChevronLeftRounded from '@mui/icons-material/ChevronLeftRounded';
import ChevronRightRounded from '@mui/icons-material/ChevronRightRounded';
import EventAvailableRounded from '@mui/icons-material/EventAvailableRounded';
import ForumRounded from '@mui/icons-material/ForumRounded';
import LocationOnRounded from '@mui/icons-material/LocationOnRounded';
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
  replies: number;
};

type CalendarEvent = {
  id: string;
  title: string;
  date: string;
  time?: string;
  location?: string;
  category: string;
};

type EventDraft = {
  title: string;
  date: string;
  time: string;
  location: string;
  category: string;
};

const STORAGE_NEWS_KEY = 'custom-news';
const STORAGE_EVENTS_KEY = 'homepage-events';

const EVENT_CATEGORIES = [
  { value: 'Treffen', label: 'Treffen', color: '#2f7d4a' },
  { value: 'Workshop', label: 'Workshop', color: '#3267d8' },
  { value: 'Party', label: 'Party', color: '#e16d48' },
  { value: 'Info', label: 'Info', color: '#d39a3f' },
];

const toDateKey = (date: Date) => {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
};

const parseDateKey = (dateKey: string) => {
  const [year, month, day] = dateKey.split('-').map(Number);
  return new Date(year, (month || 1) - 1, day || 1);
};

const monthLabel = (date: Date) =>
  date.toLocaleDateString('de-DE', { month: 'long', year: 'numeric' });

const formatShortDate = (date: Date) =>
  date.toLocaleDateString('de-DE', { day: '2-digit', month: 'short' });

const formatNewsDate = (dateString: string) => {
  const parsed = new Date(dateString);
  return isNaN(parsed.getTime()) ? dateString : formatShortDate(parsed);
};

const formatForumDate = (iso: string) => {
  const d = new Date(iso);
  return isNaN(d.getTime()) ? 'k.A.' : d.toLocaleDateString('de-DE', { day: '2-digit', month: 'short' });
};

const buildCalendarDays = (date: Date) => {
  const year = date.getFullYear();
  const month = date.getMonth();
  const start = new Date(year, month, 1);
  const end = new Date(year, month + 1, 0);
  const prevEnd = new Date(year, month, 0);

  const offset = (start.getDay() + 6) % 7; // Monday = 0
  const days: Array<{ date: Date; day: number; inMonth: boolean }> = [];

  for (let i = 0; i < offset; i++) {
    const day = prevEnd.getDate() - offset + i + 1;
    days.push({ date: new Date(year, month - 1, day), day, inMonth: false });
  }

  for (let d = 1; d <= end.getDate(); d++) {
    days.push({ date: new Date(year, month, d), day: d, inMonth: true });
  }

  while (days.length % 7 !== 0) {
    const day = days.length - (offset + end.getDate()) + 1;
    days.push({ date: new Date(year, month + 1, day), day, inMonth: false });
  }

  return days;
};

const buildSeedEvents = (): CalendarEvent[] => {
  const base = new Date();
  const make = (offset: number, title: string, time: string, location: string, category: string) => {
    const date = new Date(base);
    date.setDate(base.getDate() + offset);
    return {
      id: `seed-${offset}-${title}`,
      title,
      date: toDateKey(date),
      time,
      location,
      category,
    };
  };

  return [
    make(1, 'Erstsemester-Meetup', '18:30', 'Raum A2.12', 'Treffen'),
    make(3, 'React Workshop', '16:00', 'Labor 3', 'Workshop'),
    make(8, 'Spieleabend', '19:00', 'FSV Lounge', 'Party'),
    make(12, 'Info-Session Praktika', '14:00', 'Online', 'Info'),
  ];
};

const loadEvents = (): CalendarEvent[] => {
  if (typeof window === 'undefined') return buildSeedEvents();
  try {
    const stored = JSON.parse(window.localStorage.getItem(STORAGE_EVENTS_KEY) || '[]');
    if (Array.isArray(stored) && stored.length) return stored;
  } catch {
    // ignore
  }
  return buildSeedEvents();
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
  const normalize = (post: any): ForumPostSummary => ({
    id: String(post.id),
    title: String(post.title ?? 'Neuer Beitrag'),
    author: String(post.author ?? 'Unbekannt'),
    createdAt: String(post.createdAt ?? new Date().toISOString()),
    replies: Array.isArray(post.comments) ? post.comments.length : Number(post.replies ?? 0),
  });

  if (typeof window === 'undefined') {
    return FORUM_SEED_POSTS.map(normalize);
  }
  try {
    const stored = JSON.parse(window.localStorage.getItem(FORUM_STORAGE_KEY) || '[]');
    const source = Array.isArray(stored) && stored.length ? stored : FORUM_SEED_POSTS;
    return source.map(normalize);
  } catch {
    return FORUM_SEED_POSTS.map(normalize);
  }
};

const getCategoryColor = (category: string) =>
  EVENT_CATEGORIES.find((entry) => entry.value === category)?.color || '#2f7d4a';

const toEventTimestamp = (event: CalendarEvent) => {
  const [year, month, day] = event.date.split('-').map(Number);
  const [hour, minute] = (event.time || '00:00').split(':').map(Number);
  return new Date(year, (month || 1) - 1, day || 1, hour || 0, minute || 0).getTime();
};

const toNewsTimestamp = (item: NewsItem) => {
  if (item.createdAt) {
    const parsed = Date.parse(item.createdAt);
    if (!Number.isNaN(parsed)) return parsed;
  }
  const parsedDate = Date.parse(item.date);
  if (!Number.isNaN(parsedDate)) return parsedDate;
  return typeof item.id === 'number' ? item.id : 0;
};

const NewsFeedPage: React.FC = () => {
  const { user } = useAuth();
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  const [currentMonth, setCurrentMonth] = useState(new Date());
  const [selectedDate, setSelectedDate] = useState(() => toDateKey(new Date()));
  const [newsItems, setNewsItems] = useState<NewsItem[]>([]);
  const [forumPosts, setForumPosts] = useState<ForumPostSummary[]>([]);
  const [events, setEvents] = useState<CalendarEvent[]>([]);
  const [eventDialogOpen, setEventDialogOpen] = useState(false);
  const [eventDraft, setEventDraft] = useState<EventDraft>({
    title: '',
    date: toDateKey(new Date()),
    time: '',
    location: '',
    category: EVENT_CATEGORIES[0].value,
  });

  useEffect(() => {
    const mergedNews = [...loadNews()];
    mergedNews.sort((a, b) => toNewsTimestamp(b) - toNewsTimestamp(a));
    setNewsItems(mergedNews);

    const posts = loadForumPosts();
    posts.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
    setForumPosts(posts);

    const loadedEvents = loadEvents();
    loadedEvents.sort((a, b) => toEventTimestamp(a) - toEventTimestamp(b));
    setEvents(loadedEvents);
  }, []);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    window.localStorage.setItem(STORAGE_EVENTS_KEY, JSON.stringify(events));
  }, [events]);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    const sectionTargets = Array.from(document.querySelectorAll('[data-reveal="true"]'));
    const railTargets = Array.from(document.querySelectorAll('[data-reveal="rail"]'));
    if (!sectionTargets.length && !railTargets.length) return;
    if (!('IntersectionObserver' in window)) {
      sectionTargets.forEach((element) => element.setAttribute('data-visible', 'true'));
      railTargets.forEach((element) => element.setAttribute('data-visible', 'true'));
      return;
    }
    const sectionObserver = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.setAttribute('data-visible', 'true');
          } else {
            entry.target.removeAttribute('data-visible');
          }
        });
      },
      { threshold: 0.12, rootMargin: '0px 0px -10% 0px' }
    );
    const railObserver = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.setAttribute('data-visible', 'true');
          } else {
            entry.target.removeAttribute('data-visible');
          }
        });
      },
      { threshold: 0.2, rootMargin: '0px 0px -20% 0px' }
    );
    sectionTargets.forEach((element) => sectionObserver.observe(element));
    railTargets.forEach((element) => railObserver.observe(element));
    return () => {
      sectionObserver.disconnect();
      railObserver.disconnect();
    };
  }, [newsItems.length]);

  const days = useMemo(() => buildCalendarDays(currentMonth), [currentMonth]);
  const todayKey = toDateKey(new Date());

  const eventsByDate = useMemo(() => {
    const map = new Map<string, CalendarEvent[]>();
    events.forEach((event) => {
      const list = map.get(event.date) ?? [];
      list.push(event);
      map.set(event.date, list);
    });
    return map;
  }, [events]);

  const upcomingEvents = useMemo(() => {
    const now = new Date();
    return [...events]
      .filter((event) => toEventTimestamp(event) >= now.getTime())
      .sort((a, b) => toEventTimestamp(a) - toEventTimestamp(b));
  }, [events]);

  const latestNews = newsItems.slice(0, 8);
  const leadNews = latestNews[0];
  const secondaryNews = latestNews.slice(1, 3);
  const railNews = newsItems.slice(3, 9);
  const railItems = railNews.length ? railNews : secondaryNews;
  const latestForum = forumPosts.slice(0, 3);
  const forumHighlight = latestForum[0];
  const forumRest = latestForum.slice(1);

  const upcomingPreview = upcomingEvents.slice(0, 3);
  const heroEvent = upcomingPreview[0];
  const heroImage = '/FSV-Logo.png';
  const heroEventDate = heroEvent ? parseDateKey(heroEvent.date) : null;

  const openEventDialog = (dateKey?: string) => {
    setEventDraft((prev) => ({
      ...prev,
      date: dateKey ?? selectedDate,
    }));
    setEventDialogOpen(true);
  };

  const selectDate = (dateKey: string) => {
    setSelectedDate(dateKey);
    const parsed = parseDateKey(dateKey);
    setCurrentMonth(new Date(parsed.getFullYear(), parsed.getMonth(), 1));
  };

  const handleCreateEvent = () => {
    if (!eventDraft.title.trim() || !eventDraft.date) return;
    const id = typeof crypto !== 'undefined' && crypto.randomUUID
      ? crypto.randomUUID()
      : `event-${Date.now()}`;
    const nextEventEntry: CalendarEvent = {
      id,
      title: eventDraft.title.trim(),
      date: eventDraft.date,
      time: eventDraft.time,
      location: eventDraft.location.trim(),
      category: eventDraft.category,
    };
    setEvents((prev) => [...prev, nextEventEntry].sort((a, b) => toEventTimestamp(a) - toEventTimestamp(b)));
    setEventDialogOpen(false);
    setEventDraft((prev) => ({
      ...prev,
      title: '',
      time: '',
      location: '',
    }));
  };

  const changeMonth = (offset: number) => {
    setCurrentMonth((prev) => {
      const next = new Date(prev.getFullYear(), prev.getMonth() + offset, 1);
      setSelectedDate(toDateKey(next));
      return next;
    });
  };

  const sectionShellSx = {
    position: 'relative',
    borderRadius: { xs: 4, md: 6 },
    p: { xs: 2, md: 3 },
    overflow: 'visible',
    border: '1px solid transparent',
    backdropFilter: 'blur(16px)',
    boxShadow: isDark ? '0 26px 60px rgba(0, 0, 0, 0.35)' : '0 26px 60px rgba(16, 40, 24, 0.12)',
    willChange: 'opacity, transform, filter',
    opacity: 0,
    transform: 'translateY(48px) scale(0.98)',
    filter: 'blur(8px)',
    '&[data-visible="true"]': {
      animation: 'sectionReveal 780ms cubic-bezier(0.22, 0.8, 0.2, 1) both',
    },
    '@media (prefers-reduced-motion: reduce)': {
      opacity: 1,
      transform: 'none',
      filter: 'none',
      animation: 'none',
    },
  } as const;

  const glassCardSx = {
    borderRadius: 4,
    border: '1px solid transparent',
    background: isDark
      ? 'rgba(13, 20, 17, 0.72)'
      : `linear-gradient(145deg, rgba(255, 255, 255, 0.96), ${alpha(theme.palette.primary.main, 0.04)})`,
    boxShadow: isDark ? '0 18px 36px rgba(0, 0, 0, 0.35)' : '0 18px 36px rgba(16, 40, 24, 0.12)',
    backdropFilter: 'blur(12px)',
    transition: 'transform 160ms ease, box-shadow 160ms ease, border-color 160ms ease',
    '&:hover': {
      transform: 'translateY(-3px)',
      borderColor: 'var(--accent-strong)',
      boxShadow: isDark ? '0 22px 45px rgba(0, 0, 0, 0.4)' : '0 22px 45px rgba(16, 40, 24, 0.16)',
    },
  } as const;

  const sectionTitleSx = {
    fontFamily: '"Space Grotesk", sans-serif',
    fontWeight: 700,
    fontSize: { xs: '1.2rem', md: '1.45rem' },
    letterSpacing: '-0.01em',
  } as const;

  const sectionActionSx = {
    textTransform: 'none',
    color: 'var(--accent-strong)',
    fontWeight: 600,
    '&:hover': {
      bgcolor: 'var(--accent-soft)',
    },
  } as const;

  const sectionIconSx = {
    color: 'var(--accent-strong)',
    fontSize: 22,
  } as const;

  const sectionGradients = {
    newsroom: isDark
      ? `linear-gradient(145deg, rgba(8, 12, 10, 0.96), ${alpha(theme.palette.primary.dark, 0.35)})`
      : `linear-gradient(145deg, rgba(255, 249, 236, 0.96), ${alpha(theme.palette.primary.light, 0.22)})`,
    agenda: isDark
      ? `linear-gradient(145deg, rgba(8, 12, 10, 0.96), ${alpha(theme.palette.primary.main, 0.3)})`
      : `linear-gradient(145deg, rgba(244, 252, 248, 0.96), ${alpha(theme.palette.primary.main, 0.14)})`,
    forum: isDark
      ? `linear-gradient(145deg, rgba(9, 13, 11, 0.96), ${alpha(theme.palette.primary.dark, 0.4)})`
      : `linear-gradient(145deg, rgba(247, 250, 245, 0.96), ${alpha(theme.palette.primary.light, 0.18)})`,
  } as const;

  return (
    <Sidebar user={user} title="Startseite">
      <Box
        sx={{
          position: 'relative',
          borderRadius: { xs: 4, md: 6 },
          p: { xs: 1.6, md: 2.4 },
          overflow: 'hidden',
          color: 'var(--ink)',
          fontFamily: '"Manrope", "Space Grotesk", sans-serif',
          '--ink': isDark ? '#f6f9f6' : '#0f1411',
          '--muted': isDark ? '#9aaea3' : '#5b6660',
          '--accent': isDark ? '#6df2bf' : '#1d8f63',
          '--accent-strong': isDark ? '#3ee29b' : '#156a49',
          '--accent-soft': isDark ? 'rgba(109, 242, 191, 0.16)' : 'rgba(29, 143, 99, 0.12)',
          '--accent-2': isDark ? '#ffb454' : '#f59a47',
          '--accent-3': isDark ? '#7cc7ff' : '#3a77ff',
          '--surface': isDark ? '#0b1210' : '#f2f3ef',
          '--card-bg': isDark ? 'rgba(13, 21, 18, 0.82)' : 'rgba(255, 255, 255, 0.9)',
          '--card-border': isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(15, 29, 22, 0.12)',
          background: 'var(--surface)',
          backgroundImage: isDark
            ? 'radial-gradient(circle at 18% 22%, rgba(109, 242, 191, 0.12), transparent 55%), radial-gradient(circle at 80% 12%, rgba(255, 180, 84, 0.16), transparent 60%), radial-gradient(circle at 80% 90%, rgba(124, 199, 255, 0.16), transparent 65%)'
            : 'radial-gradient(circle at 16% 18%, rgba(29, 143, 99, 0.16), transparent 55%), radial-gradient(circle at 85% 15%, rgba(245, 154, 71, 0.2), transparent 60%), radial-gradient(circle at 85% 85%, rgba(58, 119, 255, 0.12), transparent 65%)',
          '&::before': {
            content: '""',
            position: 'absolute',
            width: { xs: 260, md: 420 },
            height: { xs: 260, md: 420 },
            right: { xs: -90, md: -160 },
            top: { xs: -80, md: -140 },
            background: 'radial-gradient(circle at 30% 30%, rgba(109, 242, 191, 0.35), transparent 70%)',
            opacity: isDark ? 0.7 : 1,
            zIndex: 0,
            animation: 'float 12s ease-in-out infinite alternate',
          },
          '&::after': {
            content: '""',
            position: 'absolute',
            width: { xs: 260, md: 420 },
            height: { xs: 260, md: 420 },
            left: { xs: -110, md: -180 },
            bottom: { xs: -150, md: -220 },
            background: 'radial-gradient(circle at 70% 70%, rgba(245, 154, 71, 0.35), transparent 70%)',
            opacity: isDark ? 0.6 : 0.9,
            zIndex: 0,
            animation: 'float 10s ease-in-out infinite alternate-reverse',
          },
          '@keyframes rise': {
            from: { opacity: 0, transform: 'translateY(36px)' },
            to: { opacity: 1, transform: 'translateY(0)' },
          },
          '@keyframes sectionReveal': {
            from: {
              opacity: 0,
              transform: 'translateY(48px) scale(0.98)',
              filter: 'blur(8px)',
            },
            to: {
              opacity: 1,
              transform: 'translateY(0) scale(1)',
              filter: 'blur(0)',
            },
          },
          '@keyframes float': {
            from: { transform: 'translateY(0px)' },
            to: { transform: 'translateY(-12px)' },
          },
          '@keyframes eventFloat': {
            from: { transform: 'translateY(0px)' },
            to: { transform: 'translateY(-12px)' },
          },
          '@keyframes gradientShift': {
            '0%': { backgroundPosition: '0% 50%' },
            '100%': { backgroundPosition: '100% 50%' },
          },
        }}
      >
        <Box sx={{ position: 'relative', zIndex: 1, display: 'grid', gap: { xs: 3, md: 4 } }}>
          <Box component="section">
            <Paper
              elevation={0}
              sx={{
                position: 'relative',
                overflow: 'hidden',
                borderRadius: { xs: 5, md: 8 },
                minHeight: { xs: 320, md: 420 },
                backgroundColor: isDark ? '#0c1411' : alpha(theme.palette.primary.main, 0.06),
                backgroundImage: isDark
                  ? `linear-gradient(120deg, ${alpha(theme.palette.primary.dark, 0.72)}, rgba(8, 12, 10, 0.9))`
                  : `linear-gradient(120deg, ${alpha(theme.palette.primary.light, 0.18)}, rgba(246, 252, 248, 0.96))`,
                border: '1px solid var(--card-border)',
                boxShadow: isDark ? '0 34px 80px rgba(0, 0, 0, 0.4)' : '0 34px 80px rgba(16, 40, 24, 0.18)',
                animation: 'rise 0.6s ease both',
                '&::before': {
                  content: '""',
                  position: 'absolute',
                  right: { xs: '-6%', sm: '6%', md: '0%' },
                  top: { xs: '56%', sm: '46%', md: '10%' },
                  width: { xs: 200, sm: 280, md: 460 },
                  height: { xs: 110, sm: 150, md: 220 },
                  backgroundImage: `url(${heroImage})`,
                  backgroundRepeat: 'no-repeat',
                  backgroundPosition: 'center',
                  backgroundSize: 'contain',
                  opacity: isDark ? 0.22 : 0.28,
                  pointerEvents: 'none',
                  zIndex: 0,
                },
              }}
            >
              <Box
                sx={{
                  position: 'absolute',
                  inset: 0,
                background: isDark
                    ? 'linear-gradient(120deg, rgba(6, 10, 8, 0.85), rgba(6, 10, 8, 0.2))'
                    : `linear-gradient(120deg, ${alpha(theme.palette.primary.dark, 0.45)}, ${alpha(
                        theme.palette.primary.dark,
                        0.08
                      )})`,
                }}
              />
              <Box
                sx={{
                  position: 'absolute',
                  inset: 0,
                  background:
                    'radial-gradient(circle at 20% 20%, rgba(109, 242, 191, 0.22), transparent 55%)',
                  mixBlendMode: 'screen',
                }}
              />
              <Stack
                spacing={2}
                sx={{
                  position: 'relative',
                  zIndex: 1,
                  p: { xs: 2.5, md: 4 },
                  maxWidth: { xs: '100%', md: '56%' },
                }}
              >
                <Box>
                  <Typography
                    variant="overline"
                    sx={{
                      letterSpacing: '0.26em',
                      fontWeight: 700,
                      color: 'var(--accent-strong)',
                    }}
                  >
                    FSV Hub
                  </Typography>
                  <Typography
                    variant="caption"
                    sx={{
                      color: 'var(--muted)',
                      letterSpacing: '0.16em',
                      textTransform: 'uppercase',
                    }}
                  >
                    Fachschaft Informatik
                  </Typography>
                </Box>
                <Typography
                  variant="overline"
                  sx={{
                    letterSpacing: '0.26em',
                    color: 'var(--accent-2)',
                    fontWeight: 700,
                  }}
                >
                  Willkommen
                </Typography>
                <Typography
                  variant="h1"
                  sx={{
                    fontFamily: '"Space Grotesk", sans-serif',
                    fontWeight: 700,
                    lineHeight: 1.05,
                    fontSize: { xs: '2rem', md: '3rem' },
                    color: '#f8fbf7',
                  }}
                >
                  Willkommen bei der Fachschaft Informatik.
                </Typography>
                <Typography variant="body1" sx={{ color: '#dfe7e1', maxWidth: 560 }}>
                  Dein Hub für Termine, Unterstützung und Austausch im Studium.
                </Typography>
                <Stack
                  direction={{ xs: 'column', sm: 'row' }}
                  spacing={1}
                  alignItems={{ xs: 'stretch', sm: 'center' }}
                >
                  <Button
                    component="a"
                    href="#events"
                    variant="contained"
                    sx={{
                      borderRadius: 999,
                      textTransform: 'none',
                      bgcolor: 'var(--accent-strong)',
                      color: isDark ? '#0b1511' : '#fff',
                      boxShadow: `0 14px 30px ${alpha(theme.palette.success.main, 0.25)}`,
                      '&:hover': { bgcolor: isDark ? '#4be5ab' : '#0f5e3d' },
                    }}
                  >
                    Zu den Events
                  </Button>
                  <Button
                    component={RouterLink}
                    to="/forum"
                    variant="outlined"
                    sx={{
                      borderRadius: 999,
                      textTransform: 'none',
                      borderColor: 'rgba(255, 255, 255, 0.4)',
                      color: '#f6faf6',
                      '&:hover': { borderColor: 'var(--accent-strong)', bgcolor: 'rgba(255, 255, 255, 0.08)' },
                    }}
                  >
                    Zum Forum
                  </Button>
                </Stack>
              </Stack>

              {heroEvent && (
                <Box
                  sx={{
                    position: { xs: 'relative', md: 'absolute' },
                    right: { md: 28 },
                    bottom: { md: 24 },
                    mt: { xs: 2, md: 0 },
                    mx: { xs: 2.5, md: 0 },
                    width: { xs: 'auto', sm: 320 },
                  }}
                >
                  <Paper
                    elevation={0}
                    sx={{
                      ...glassCardSx,
                      p: 1.6,
                      borderRadius: 4,
                      transform: { md: 'rotate(-1deg)' },
                      animation: 'eventFloat 6s ease-in-out infinite alternate',
                    }}
                  >
                    <Stack spacing={1}>
                      <Typography
                        variant="overline"
                        sx={{ letterSpacing: '0.22em', color: 'var(--accent-strong)', fontWeight: 700 }}
                      >
                        Nächstes Event
                      </Typography>
                      <Typography variant="h6" fontWeight={700}>
                        {heroEvent.title}
                      </Typography>
                      <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
                        <Typography variant="caption" sx={{ color: 'var(--muted)' }}>
                          {heroEventDate ? formatShortDate(heroEventDate) : heroEvent.date}
                        </Typography>
                        {heroEvent.time && (
                          <Typography variant="caption" sx={{ color: 'var(--muted)' }}>
                            {heroEvent.time} Uhr
                          </Typography>
                        )}
                        {heroEvent.location && (
                          <Stack direction="row" spacing={0.3} alignItems="center">
                            <LocationOnRounded sx={{ fontSize: 14, color: 'var(--muted)' }} />
                            <Typography variant="caption" sx={{ color: 'var(--muted)' }}>
                              {heroEvent.location}
                            </Typography>
                          </Stack>
                        )}
                      </Stack>
                      <Button
                        component="a"
                        href="#events"
                        size="small"
                        sx={{
                          alignSelf: 'flex-start',
                          textTransform: 'none',
                          color: 'var(--accent-strong)',
                          px: 0,
                        }}
                      >
                        Zum Kalender
                      </Button>
                    </Stack>
                  </Paper>
                </Box>
              )}
            </Paper>
          </Box>

          <Box
            component="section"
            data-reveal="true"
            sx={{
              ...sectionShellSx,
              background: sectionGradients.newsroom,
            }}
          >
            <Box sx={{ position: 'relative', zIndex: 1 }}>
              <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 2 }}>
                <Stack direction="row" spacing={1} alignItems="center">
                  <CampaignRounded sx={sectionIconSx} />
                  <Typography variant="h4" sx={sectionTitleSx}>
                    Newsroom
                  </Typography>
                </Stack>
                <Button
                  component={RouterLink}
                  to="/news"
                  size="small"
                  endIcon={<ArrowForwardRounded fontSize="small" />}
                  sx={sectionActionSx}
                >
                  Alle News
                </Button>
              </Stack>

              <Box
                sx={{
                  display: 'grid',
                  gap: { xs: 2, md: 3 },
                  gridTemplateColumns: { xs: '1fr', md: 'repeat(12, minmax(0, 1fr))' },
                  alignItems: 'stretch',
                }}
              >
                <Box sx={{ gridColumn: { xs: '1 / -1', md: 'span 7' } }}>
                  {leadNews ? (
                    <Paper
                      component={RouterLink}
                      to={`/news/${leadNews.id}`}
                      elevation={0}
                      sx={{
                        ...glassCardSx,
                        display: 'flex',
                        flexDirection: 'column',
                        position: 'relative',
                        minHeight: { xs: 220, md: 280 },
                        borderRadius: 5,
                        overflow: 'hidden',
                        textDecoration: 'none',
                        color: 'inherit',
                        backgroundImage: leadNews.image
                          ? `url(${leadNews.image})`
                          : 'linear-gradient(135deg, rgba(12, 18, 15, 0.96), rgba(8, 13, 11, 0.98))',
                        backgroundSize: 'cover',
                        backgroundPosition: 'center',
                        height: '100%',
                      }}
                    >
                      <Box
                        sx={{
                          position: 'absolute',
                          inset: 0,
                          background:
                            'linear-gradient(130deg, rgba(8, 12, 10, 0.88), rgba(8, 12, 10, 0.35))',
                        }}
                      />
                      <Stack
                        spacing={0.8}
                        sx={{
                          position: 'relative',
                          zIndex: 1,
                          p: { xs: 2, md: 2.6 },
                          height: '100%',
                          justifyContent: 'flex-end',
                        }}
                      >
                        <Typography variant="overline" sx={{ letterSpacing: '0.18em', color: '#e3f2ea' }}>
                          {formatNewsDate(leadNews.date)}
                        </Typography>
                        <Typography
                          variant="h4"
                          sx={{ fontWeight: 700, color: '#f8fbf7', lineHeight: 1.15 }}
                        >
                          {leadNews.title}
                        </Typography>
                        <Typography
                          variant="body2"
                          sx={{
                            color: '#d9e3db',
                            display: '-webkit-box',
                            WebkitLineClamp: 2,
                            WebkitBoxOrient: 'vertical',
                            overflow: 'hidden',
                          }}
                        >
                          {leadNews.summary || 'Mehr lesen …'}
                        </Typography>
                      </Stack>
                    </Paper>
                  ) : (
                    <Paper
                      elevation={0}
                      sx={{
                        ...glassCardSx,
                        p: { xs: 2, md: 2.6 },
                        borderRadius: 5,
                      }}
                    >
                      <Typography variant="body1" sx={{ color: 'var(--muted)' }}>
                        Weitere News folgen in Kürze.
                      </Typography>
                    </Paper>
                  )}
                </Box>

                <Paper
                  elevation={0}
                  sx={{
                    ...glassCardSx,
                    gridColumn: { xs: '1 / -1', md: 'span 5' },
                    p: { xs: 1.8, md: 2.2 },
                    borderRadius: 4,
                    background: isDark ? 'rgba(12, 18, 16, 0.72)' : 'rgba(255, 255, 255, 0.9)',
                  }}
                >
                  <Stack spacing={1.2}>
                    <Typography
                      variant="overline"
                      sx={{ letterSpacing: '0.24em', color: 'var(--accent-strong)', fontWeight: 700 }}
                    >
                      Kurz &amp; Knapp
                    </Typography>
                    {secondaryNews.length ? (
                      secondaryNews.map((item) => (
                        <Box
                          key={item.id}
                          component={RouterLink}
                          to={`/news/${item.id}`}
                          sx={{
                            textDecoration: 'none',
                            color: 'inherit',
                            display: 'grid',
                            gap: 0.3,
                            p: 1,
                            borderRadius: 2.5,
                            border: '1px solid var(--card-border)',
                            background: isDark ? 'rgba(10, 16, 14, 0.7)' : 'rgba(255, 255, 255, 0.85)',
                            transition: 'border-color 150ms ease, transform 150ms ease',
                            '&:hover': {
                              borderColor: 'var(--accent-strong)',
                              transform: 'translateY(-2px)',
                            },
                          }}
                        >
                          <Typography variant="caption" sx={{ color: 'var(--muted)', letterSpacing: '0.12em' }}>
                            {formatNewsDate(item.date)}
                          </Typography>
                          <Typography variant="subtitle1" sx={{ fontWeight: 700, lineHeight: 1.2 }}>
                            {item.title}
                          </Typography>
                        </Box>
                      ))
                    ) : (
                      <Typography variant="body2" sx={{ color: 'var(--muted)' }}>
                        Keine weiteren News.
                      </Typography>
                    )}
                    <Button
                      component={RouterLink}
                      to="/news"
                      size="small"
                      endIcon={<ArrowForwardRounded fontSize="small" />}
                      sx={{
                        ...sectionActionSx,
                        alignSelf: 'flex-start',
                        px: 0,
                      }}
                    >
                      Alle News
                    </Button>
                  </Stack>
                </Paper>
              </Box>

              {railItems.length ? (
                <Box
                  data-reveal="rail"
                  sx={{
                    mt: { xs: 2.5, md: 3.5 },
                    opacity: 0,
                    transform: 'translateY(32px)',
                    transition: 'opacity 600ms ease, transform 600ms ease',
                    '&[data-visible="true"]': {
                      opacity: 1,
                      transform: 'translateY(0)',
                    },
                    '@media (prefers-reduced-motion: reduce)': {
                      opacity: 1,
                      transform: 'none',
                      transition: 'none',
                    },
                  }}
                >
                  <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 1.6 }}>
                    <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>
                      Mehr Stories
                    </Typography>
                  </Stack>
                  <Box
                    sx={{
                      mt: { xs: 3, md: 4 },
                      display: 'grid',
                      gridAutoFlow: 'column',
                      gridAutoColumns: { xs: '80%', sm: '55%', md: '32%' },
                      gap: 1.6,
                      overflowX: 'auto',
                      scrollSnapType: 'x mandatory',
                      pb: 1.2,
                      px: 0.4,
                      scrollbarWidth: 'thin',
                      scrollbarColor: 'rgba(0,0,0,0.2) transparent',
                      '& > *': { scrollSnapAlign: 'start', scrollSnapStop: 'always' },
                      '&::-webkit-scrollbar': {
                        height: 6,
                      },
                      '&::-webkit-scrollbar-thumb': {
                        background: 'rgba(0,0,0,0.2)',
                        borderRadius: 999,
                      },
                      '&::-webkit-scrollbar-track': {
                        background: 'transparent',
                      },
                    }}
                  >
                    {railItems.map((item) => (
                      <Paper
                        key={item.id}
                        component={RouterLink}
                        to={`/news/${item.id}`}
                        elevation={0}
                        sx={{
                          ...glassCardSx,
                          borderRadius: 4,
                          overflow: 'hidden',
                          textDecoration: 'none',
                          color: 'inherit',
                        }}
                      >
                        <Box
                          sx={{
                            height: 130,
                            backgroundImage: item.image
                              ? `url(${item.image})`
                              : 'linear-gradient(135deg, rgba(12, 18, 15, 0.96), rgba(8, 13, 11, 0.98))',
                            backgroundSize: 'cover',
                            backgroundPosition: 'center',
                          }}
                        />
                        <Stack spacing={0.6} sx={{ p: 1.6 }}>
                          <Typography variant="caption" sx={{ color: 'var(--muted)', letterSpacing: '0.12em' }}>
                            {formatNewsDate(item.date)}
                          </Typography>
                          <Typography variant="subtitle1" sx={{ fontWeight: 700, lineHeight: 1.2 }}>
                            {item.title}
                          </Typography>
                          <Typography
                            variant="body2"
                            sx={{
                              color: 'var(--muted)',
                              display: '-webkit-box',
                              WebkitLineClamp: 2,
                              WebkitBoxOrient: 'vertical',
                              overflow: 'hidden',
                            }}
                          >
                            {item.summary || 'Mehr lesen …'}
                          </Typography>
                        </Stack>
                      </Paper>
                    ))}
                  </Box>
                </Box>
              ) : null}
            </Box>
          </Box>

          <Box
            component="section"
            id="events"
            data-reveal="true"
            sx={{
              ...sectionShellSx,
              scrollMarginTop: { xs: 80, md: 100 },
              background: sectionGradients.agenda,
            }}
          >
            <Box sx={{ position: 'relative', zIndex: 1 }}>
              <Stack
                direction={{ xs: 'column', sm: 'row' }}
                spacing={1}
                alignItems={{ xs: 'flex-start', sm: 'center' }}
                justifyContent="space-between"
                sx={{ mb: 2 }}
              >
                <Stack direction="row" spacing={1} alignItems="center">
                  <EventAvailableRounded sx={sectionIconSx} />
                  <Typography variant="h4" sx={sectionTitleSx}>
                    Agenda
                  </Typography>
                </Stack>
                <Button
                  size="small"
                  variant="contained"
                  startIcon={<AddRounded fontSize="small" />}
                  onClick={() => openEventDialog()}
                  sx={{
                    borderRadius: 999,
                    textTransform: 'none',
                    bgcolor: 'var(--accent-strong)',
                    color: isDark ? '#0b1511' : '#fff',
                    '&:hover': { bgcolor: isDark ? '#4be5ab' : '#0f5e3d' },
                  }}
                >
                  Event eintragen
                </Button>
              </Stack>

              <Box
                sx={{
                  display: 'grid',
                  gap: { xs: 2, md: 3 },
                  gridTemplateColumns: { xs: '1fr', lg: '7fr 5fr' },
                  alignItems: 'start',
                }}
              >
                <Box>
                  <Typography
                    variant="caption"
                    sx={{
                      color: 'var(--muted)',
                      fontWeight: 700,
                      letterSpacing: '0.22em',
                      textTransform: 'uppercase',
                      display: 'block',
                      mb: 1.2,
                    }}
                  >
                    Kommende Termine
                  </Typography>
                  <Box
                    sx={{
                      position: 'relative',
                      pl: { xs: 2.5, md: 3.5 },
                    }}
                  >
                    <Box
                      sx={{
                        position: 'absolute',
                        left: { xs: 10, md: 14 },
                        top: 6,
                        bottom: 6,
                        width: 2,
                        bgcolor: 'var(--card-border)',
                        borderRadius: 999,
                      }}
                    />
                    <Stack spacing={1.6}>
                      {upcomingPreview.length ? (
                        upcomingPreview.map((event) => {
                          const accent = getCategoryColor(event.category);
                          const eventDate = parseDateKey(event.date);
                          const isActive = event.date === selectedDate;
                          return (
                            <Box
                              key={event.id}
                              component="button"
                              type="button"
                              onClick={() => selectDate(event.date)}
                              sx={{
                                appearance: 'none',
                                border: 'none',
                                background: 'transparent',
                                textAlign: 'left',
                                padding: 0,
                                cursor: 'pointer',
                                width: '100%',
                                fontFamily: 'inherit',
                                '&:focus-visible': {
                                  outline: '2px solid var(--accent-3)',
                                  outlineOffset: 3,
                                },
                              }}
                            >
                              <Box
                                sx={{
                                  ...glassCardSx,
                                  p: { xs: 1.6, md: 2 },
                                  borderRadius: 4,
                                  borderColor: isActive ? 'var(--accent-strong)' : 'transparent',
                                  background: isActive
                                    ? 'var(--accent-soft)'
                                    : isDark
                                    ? 'rgba(12, 18, 24, 0.72)'
                                    : 'rgba(255, 255, 255, 0.94)',
                                  position: 'relative',
                                  color: isDark ? '#f6f9f6' : 'inherit',
                                }}
                              >
                                <Box
                                  sx={{
                                    position: 'absolute',
                                    left: { xs: -16, md: -20 },
                                    top: 20,
                                    width: 12,
                                    height: 12,
                                    borderRadius: '50%',
                                    bgcolor: accent,
                                    boxShadow: `0 0 0 6px ${alpha(accent, 0.18)}`,
                                  }}
                                />
                                <Stack spacing={0.6}>
                                  <Typography variant="overline" sx={{ letterSpacing: '0.18em', color: 'var(--muted)' }}>
                                    {formatShortDate(eventDate)}
                                  </Typography>
                                  <Typography variant="h6" sx={{ fontWeight: 700 }}>
                                    {event.title}
                                  </Typography>
                                  <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
                                    {event.time && (
                                      <Typography variant="caption" sx={{ color: 'var(--muted)' }}>
                                        {event.time} Uhr
                                      </Typography>
                                    )}
                                    {event.location && (
                                      <Stack direction="row" spacing={0.3} alignItems="center">
                                        <LocationOnRounded sx={{ fontSize: 13, color: 'var(--muted)' }} />
                                        <Typography variant="caption" sx={{ color: 'var(--muted)' }}>
                                          {event.location}
                                        </Typography>
                                      </Stack>
                                    )}
                                    <Typography
                                      variant="caption"
                                      sx={{
                                        color: accent,
                                        fontWeight: 600,
                                        textTransform: 'uppercase',
                                        letterSpacing: '0.08em',
                                      }}
                                    >
                                      {event.category}
                                    </Typography>
                                  </Stack>
                                </Stack>
                              </Box>
                            </Box>
                          );
                        })
                      ) : (
                        <Typography variant="body2" sx={{ color: 'var(--muted)' }}>
                          Noch keine Events eingetragen.
                        </Typography>
                      )}
                    </Stack>
                  </Box>
                </Box>

                <Paper
                  elevation={0}
                  sx={{
                    ...glassCardSx,
                    mt: { xs: 2.6, md: 3.2 },
                    p: { xs: 1.6, md: 2 },
                    borderRadius: 4,
                    background: isDark ? 'rgba(16, 20, 18, 0.92)' : '#ffffff',
                  }}
                >
                  <Stack spacing={1}>
                    <Stack direction="row" alignItems="center" justifyContent="space-between">
                      <Stack direction="row" spacing={0.6} alignItems="center">
                        <CalendarMonthRounded sx={sectionIconSx} />
                        <Typography variant="subtitle1" fontWeight={700}>
                          Kalender
                        </Typography>
                      </Stack>
                      <Stack direction="row" spacing={0.3} alignItems="center">
                        <IconButton size="small" onClick={() => changeMonth(-1)} sx={{ p: 0.3 }}>
                          <ChevronLeftRounded fontSize="small" />
                        </IconButton>
                        <IconButton size="small" onClick={() => changeMonth(1)} sx={{ p: 0.3 }}>
                          <ChevronRightRounded fontSize="small" />
                        </IconButton>
                      </Stack>
                    </Stack>

                    <Typography variant="caption" sx={{ color: 'var(--muted)' }}>
                      {monthLabel(currentMonth)}
                    </Typography>

                    <Box
                      sx={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(7, minmax(0, 1fr))',
                        gap: 0.3,
                        textAlign: 'center',
                      }}
                    >
                      {['Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa', 'So'].map((label) => (
                        <Typography key={label} variant="caption" sx={{ fontWeight: 700, color: 'var(--muted)' }}>
                          {label}
                        </Typography>
                      ))}
                    </Box>
                    <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(7, minmax(0, 1fr))', gap: 0.3 }}>
                      {days.map(({ date, day, inMonth }, idx) => {
                        const dateKey = toDateKey(date);
                        const isToday = dateKey === todayKey;
                        const isSelected = dateKey === selectedDate;
                        const dayEvents = eventsByDate.get(dateKey) ?? [];
                        const hasEvents = dayEvents.length > 0;
                        return (
                          <Box
                            key={`${dateKey}-${idx}`}
                            component="button"
                            type="button"
                            onClick={() => selectDate(dateKey)}
                            sx={{
                              border: '1px solid',
                              borderColor: isSelected
                                ? 'var(--accent-strong)'
                                : isToday
                                ? alpha(theme.palette.success.main, 0.4)
                                : 'transparent',
                              bgcolor: isSelected ? 'var(--accent-soft)' : 'transparent',
                              borderRadius: 1.6,
                              height: { xs: 24, md: 28 },
                              cursor: 'pointer',
                              color: inMonth ? 'var(--ink)' : 'var(--muted)',
                              opacity: inMonth ? 1 : 0.5,
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              flexDirection: 'column',
                              gap: 0.1,
                              fontFamily: 'inherit',
                              transition: 'transform 120ms ease, border-color 120ms ease',
                              '&:hover': {
                                transform: 'translateY(-2px)',
                                borderColor: 'var(--accent-strong)',
                              },
                            }}
                          >
                            <Typography variant="caption" fontWeight={700} sx={{ fontSize: '0.72rem' }}>
                              {day}
                            </Typography>
                            {hasEvents && (
                              <Box
                                sx={{
                                  width: 4,
                                  height: 4,
                                  borderRadius: '50%',
                                  bgcolor: getCategoryColor(dayEvents[0].category),
                                  opacity: 0.9,
                                }}
                              />
                            )}
                          </Box>
                        );
                      })}
                    </Box>
                  </Stack>
                </Paper>
              </Box>
            </Box>
          </Box>

          <Box
            component="section"
            data-reveal="true"
            sx={{
              ...sectionShellSx,
              background: sectionGradients.forum,
            }}
          >
            <Box sx={{ position: 'relative', zIndex: 1 }}>
              <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 2 }}>
                <Stack direction="row" spacing={1} alignItems="center">
                  <ForumRounded sx={sectionIconSx} />
                  <Typography variant="h4" sx={sectionTitleSx}>
                    Forum
                  </Typography>
                </Stack>
                <Button
                  component={RouterLink}
                  to="/forum"
                  size="small"
                  endIcon={<ArrowForwardRounded fontSize="small" />}
                  sx={sectionActionSx}
                >
                  Zum Forum
                </Button>
              </Stack>

              {forumHighlight ? (
                <Box
                  sx={{
                    display: 'grid',
                    gap: { xs: 2, md: 3 },
                    gridTemplateColumns: { xs: '1fr', md: 'repeat(12, minmax(0, 1fr))' },
                  }}
                >
                  <Paper
                    component={RouterLink}
                    to="/forum"
                    elevation={0}
                    sx={{
                      ...glassCardSx,
                      gridColumn: { xs: '1 / -1', md: 'span 7' },
                      p: { xs: 2, md: 3 },
                      borderRadius: 5,
                      textDecoration: 'none',
                      color: 'inherit',
                      background: isDark
                        ? 'linear-gradient(135deg, rgba(17, 24, 30, 0.92), rgba(10, 16, 20, 0.9))'
                        : 'linear-gradient(135deg, rgba(233, 243, 255, 0.95), rgba(255, 255, 255, 0.95))',
                    }}
                  >
                    <Stack spacing={1.4}>
                      <Stack direction="row" spacing={1.2} alignItems="center">
                        <Avatar sx={{ width: 42, height: 42, bgcolor: 'var(--accent-soft)', color: 'var(--accent-strong)' }}>
                          {forumHighlight.author?.charAt(0)?.toUpperCase() || '?'}
                        </Avatar>
                        <Box>
                          <Typography variant="overline" sx={{ letterSpacing: '0.22em', color: 'var(--accent-strong)' }}>
                            Top Diskussion
                          </Typography>
                          <Typography variant="caption" sx={{ color: 'var(--muted)' }}>
                            {forumHighlight.author} · {formatForumDate(forumHighlight.createdAt)}
                          </Typography>
                        </Box>
                      </Stack>
                      <Typography variant="h4" sx={{ fontWeight: 700, lineHeight: 1.2 }}>
                        {forumHighlight.title}
                      </Typography>
                      <Typography variant="body2" sx={{ color: 'var(--muted)' }}>
                        {forumHighlight.replies} Antworten
                      </Typography>
                      <Stack direction="row" spacing={0.4} alignItems="center">
                        <Typography variant="caption" sx={{ color: 'var(--accent-strong)', fontWeight: 600 }}>
                          Zum Beitrag
                        </Typography>
                        <ArrowForwardRounded fontSize="small" sx={{ color: 'var(--accent-strong)' }} />
                      </Stack>
                    </Stack>
                  </Paper>

                  <Stack spacing={1.4} sx={{ gridColumn: { xs: '1 / -1', md: 'span 5' } }}>
                    {forumRest.length ? (
                      forumRest.map((post) => (
                        <Box
                          key={post.id}
                          component={RouterLink}
                          to="/forum"
                          sx={{
                            ...glassCardSx,
                            p: 1.6,
                            borderRadius: 3,
                            textDecoration: 'none',
                            color: 'inherit',
                            display: 'grid',
                            gap: 0.6,
                          }}
                        >
                          <Stack direction="row" spacing={1} alignItems="center">
                            <Avatar sx={{ width: 34, height: 34, fontSize: '0.8rem' }}>
                              {post.author?.charAt(0)?.toUpperCase() || '?'}
                            </Avatar>
                            <Box sx={{ minWidth: 0 }}>
                              <Typography variant="subtitle1" sx={{ fontWeight: 600 }} noWrap>
                                {post.title}
                              </Typography>
                              <Typography variant="caption" sx={{ color: 'var(--muted)' }}>
                                {post.author} · {formatForumDate(post.createdAt)}
                              </Typography>
                            </Box>
                          </Stack>
                          <Typography variant="caption" sx={{ color: 'var(--muted)' }}>
                            {post.replies} Antworten
                          </Typography>
                        </Box>
                      ))
                    ) : (
                      <Typography variant="body2" sx={{ color: 'var(--muted)' }}>
                        Keine weiteren Threads.
                      </Typography>
                    )}
                  </Stack>
                </Box>
              ) : (
                <Typography variant="body2" sx={{ color: 'var(--muted)' }}>
                  Noch keine Beiträge.
                </Typography>
              )}
            </Box>
          </Box>
        </Box>
      </Box>

      <Dialog
        open={eventDialogOpen}
        onClose={() => setEventDialogOpen(false)}
        fullWidth
        maxWidth="sm"
        PaperProps={{
          sx: {
            borderRadius: 3,
            background: isDark ? '#0f1512' : '#fff',
          },
        }}
      >
        <DialogTitle>Event eintragen</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              label="Titel"
              value={eventDraft.title}
              onChange={(event) => setEventDraft((prev) => ({ ...prev, title: event.target.value }))}
              fullWidth
            />
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
              <TextField
                label="Datum"
                type="date"
                value={eventDraft.date}
                onChange={(event) => setEventDraft((prev) => ({ ...prev, date: event.target.value }))}
                fullWidth
                InputLabelProps={{ shrink: true }}
              />
              <TextField
                label="Uhrzeit"
                type="time"
                value={eventDraft.time}
                onChange={(event) => setEventDraft((prev) => ({ ...prev, time: event.target.value }))}
                fullWidth
                InputLabelProps={{ shrink: true }}
              />
            </Stack>
            <TextField
              label="Ort"
              value={eventDraft.location}
              onChange={(event) => setEventDraft((prev) => ({ ...prev, location: event.target.value }))}
              fullWidth
            />
            <TextField
              label="Kategorie"
              select
              value={eventDraft.category}
              onChange={(event) => setEventDraft((prev) => ({ ...prev, category: event.target.value }))}
              fullWidth
            >
              {EVENT_CATEGORIES.map((category) => (
                <MenuItem key={category.value} value={category.value}>
                  {category.label}
                </MenuItem>
              ))}
            </TextField>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEventDialogOpen(false)}>Abbrechen</Button>
          <Button
            variant="contained"
            onClick={handleCreateEvent}
            disabled={!eventDraft.title.trim() || !eventDraft.date}
          >
            Speichern
          </Button>
        </DialogActions>
      </Dialog>
    </Sidebar>
  );
};

export default NewsFeedPage;
