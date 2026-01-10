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

type ForumPostSummary = { id: string; title: string; author: string; createdAt: string; replies: number; };
type CalendarEvent = { id: string; title: string; date: string; time?: string; location?: string; category: string; };
type EventDraft = { title: string; date: string; time: string; location: string; category: string; };

const STORAGE_NEWS_KEY = 'custom-news';
const STORAGE_EVENTS_KEY = 'homepage-events';

const EVENT_CATEGORIES = [
  { value: 'Treffen', label: 'Treffen', tone: 'success' },
  { value: 'Workshop', label: 'Workshop', tone: 'info' },
  { value: 'Party', label: 'Party', tone: 'error' },
  { value: 'Info', label: 'Info', tone: 'warning' },
] as const;

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
const monthLabel = (date: Date) => date.toLocaleDateString('de-DE', { month: 'long', year: 'numeric' });
const formatShortDate = (date: Date) => date.toLocaleDateString('de-DE', { day: '2-digit', month: 'short' });
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
  const offset = (start.getDay() + 6) % 7; 
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
    return { id: `seed-${offset}-${title}`, title, date: toDateKey(date), time, location, category };
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
  } catch { /* ignore */ }
  return buildSeedEvents();
};
const loadNews = (): NewsItem[] => {
  if (typeof window === 'undefined') return newsDaten;
  try {
    const custom = JSON.parse(window.localStorage.getItem(STORAGE_NEWS_KEY) || '[]');
    return [...newsDaten, ...(Array.isArray(custom) ? custom : [])];
  } catch { return newsDaten; }
};
const loadForumPosts = (): ForumPostSummary[] => {
  const normalize = (post: any): ForumPostSummary => ({
    id: String(post.id),
    title: String(post.title ?? 'Neuer Beitrag'),
    author: String(post.author ?? 'Unbekannt'),
    createdAt: String(post.createdAt ?? new Date().toISOString()),
    replies: Array.isArray(post.comments) ? post.comments.length : Number(post.replies ?? 0),
  });
  if (typeof window === 'undefined') return FORUM_SEED_POSTS.map(normalize);
  try {
    const stored = JSON.parse(window.localStorage.getItem(FORUM_STORAGE_KEY) || '[]');
    const source = Array.isArray(stored) && stored.length ? stored : FORUM_SEED_POSTS;
    return source.map(normalize);
  } catch { return FORUM_SEED_POSTS.map(normalize); }
};
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
  
  const isAdmin = user?.role === 'admin';

  const theme = useTheme();
  const custom = (theme as any).custom;
  const isDark = theme.palette.mode === 'dark';

  const categoryColors = useMemo(() => {
    const paletteMap = {
      success: theme.palette.success.main,
      info: theme.palette.info.main,
      warning: theme.palette.warning.main,
      error: theme.palette.error.main,
    };
    return EVENT_CATEGORIES.reduce<Record<string, string>>((acc, entry) => {
      acc[entry.value] = paletteMap[entry.tone];
      return acc;
    }, {});
  }, [theme]);

  const getCategoryColor = (category: string) => categoryColors[category] ?? theme.palette.success.main;
  const imageOverlayBase = isDark ? theme.palette.background.default : theme.palette.text.primary;

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
    if (!('IntersectionObserver' in window)) {
        document.querySelectorAll('[data-reveal]').forEach(el => el.setAttribute('data-visible', 'true'));
        return;
    }
    const revealOnce = (entries: IntersectionObserverEntry[], observer: IntersectionObserver) => {
      entries.forEach((entry) => {
        if (!entry.isIntersecting) return;
        entry.target.setAttribute('data-visible', 'true');
        observer.unobserve(entry.target);
      });
    };
    const sectionTargets = document.querySelectorAll('[data-reveal="true"]');
    const railTargets = document.querySelectorAll('[data-reveal="rail"]');
    const sectionObserver = new IntersectionObserver(revealOnce, { threshold: 0.12, rootMargin: '0px 0px -10% 0px' });
    const railObserver = new IntersectionObserver(revealOnce, { threshold: 0.2, rootMargin: '0px 0px -20% 0px' });
    sectionTargets.forEach((el) => sectionObserver.observe(el));
    railTargets.forEach((el) => railObserver.observe(el));
    return () => { sectionObserver.disconnect(); railObserver.disconnect(); };
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
  const heroEventDate = heroEvent ? parseDateKey(heroEvent.date) : null;

  const openEventDialog = (dateKey?: string) => {
    setEventDraft((prev) => ({ ...prev, date: dateKey ?? selectedDate }));
    setEventDialogOpen(true);
  };
  const selectDate = (dateKey: string) => {
    setSelectedDate(dateKey);
    const parsed = parseDateKey(dateKey);
    setCurrentMonth(new Date(parsed.getFullYear(), parsed.getMonth(), 1));
  };
  const handleCreateEvent = () => {
    if (!isAdmin) return;

    if (!eventDraft.title.trim() || !eventDraft.date) return;
    const id = typeof crypto !== 'undefined' && crypto.randomUUID ? crypto.randomUUID() : `event-${Date.now()}`;
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
    setEventDraft((prev) => ({ ...prev, title: '', time: '', location: '' }));
  };
  const changeMonth = (offset: number) => {
    setCurrentMonth((prev) => {
      const next = new Date(prev.getFullYear(), prev.getMonth() + offset, 1);
      setSelectedDate(toDateKey(next));
      return next;
    });
  };

  return (
    <Sidebar user={user} title="Startseite" fullBleed>
      <Box sx={{
        ...custom.pageWrapper,
        backgroundColor: 'transparent',
        boxShadow: 'none',
        borderRadius: 0,
        border: 'none',
      }}>
        <Box sx={custom.gridContainer}>
          
         
          <Box
            component="section"
            sx={{
              gridColumn: '1 / -1',
              mb: { xs: 6, md: 10 } 
            }}
          >
            <Paper
              elevation={0}
              sx={custom.heroSection}
            >
              <Box
                sx={{
                  position: 'absolute',
                  inset: 0,
                  background: isDark
                    ? `linear-gradient(120deg, ${alpha(theme.palette.background.default, 0.85)}, ${alpha(theme.palette.background.default, 0.2)})`
                    : `linear-gradient(120deg, ${alpha(theme.palette.primary.dark, 0.45)}, ${alpha(theme.palette.primary.dark, 0.08)})`,
                }}
              />
              <Box
                sx={{
                  position: 'absolute',
                  inset: 0,
                  background: isDark
                    ? `radial-gradient(circle at 20% 20%, ${alpha(theme.palette.primary.light, 0.22)}, transparent 55%)`
                    : `radial-gradient(circle at 20% 20%, ${alpha(theme.palette.primary.light, 0.2)}, transparent 55%)`,
                  mixBlendMode: 'screen',
                }}
              />
              <Box
                sx={{
                  position: 'relative',
                  zIndex: 1,
                  maxWidth: { xs: '100%', md: 'calc(var(--page-max-width) + (var(--grid-gap) * 12))' },
                  mx: 'auto',
                  px: { xs: 8, md: 12 },
                  py: { xs: 3, md: 4 },
                  display: 'grid',
                  gridTemplateColumns: { xs: '1fr', md: 'repeat(12, minmax(0, 1fr))' },
                  columnGap: { xs: 2, md: 4 }, 
                  rowGap: { xs: 3, md: 4 },   
                }}
              >
                <Stack spacing={1} sx={{ gridColumn: { xs: '1 / -1', md: '1 / 9' }, justifyContent: 'center' }}>
                  <Box>
                    <Typography variant="overline" sx={custom.heroOverline}>
                      FSV Hub
                    </Typography>
                    <Typography variant="caption" sx={{ color: 'var(--muted)', letterSpacing: '0.16em', textTransform: 'uppercase', display: 'block' }}>
                      Fachschaft Informatik
                    </Typography>
                  </Box>
                  <Typography
                    variant="h1"
                    sx={{
                      fontFamily: '"Space Grotesk", sans-serif',
                      fontWeight: 700,
                      lineHeight: 1.1,
                      fontSize: { xs: '1.75rem', md: '2.5rem' },
                      color: 'var(--ink)',
                    }}
                  >
                    Willkommen bei der Fachschaft.
                  </Typography>
                  <Typography variant="body1" sx={{ color: 'var(--muted)', maxWidth: 560, mb: 0.5 }}>
                    Dein Hub für Termine, Unterstützung und Austausch im Studium.
                  </Typography>
                  <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} sx={{ mt: 1 }}>
                    <Button
                      component="a"
                      href="#events"
                      variant="contained"
                      sx={{
                        ...custom.roundButton,
                        bgcolor: 'var(--accent-strong)',
                        color: theme.palette.primary.contrastText,
                        boxShadow: `0 14px 30px ${alpha(theme.palette.success.main, 0.25)}`,
                        '&:hover': { bgcolor: isDark ? theme.palette.primary.light : theme.palette.primary.dark },
                      }}
                    >
                      Zu den Events
                    </Button>
                    <Button
                      component={RouterLink}
                      to="/forum"
                      variant="outlined"
                      sx={{
                        ...custom.roundButton,
                        borderColor: alpha(theme.palette.text.primary, 0.4),
                        color: theme.palette.text.primary,
                        '&:hover': { borderColor: 'var(--accent-strong)', bgcolor: alpha(theme.palette.text.primary, 0.08) },
                      }}
                    >
                      Zum Forum
                    </Button>
                  </Stack>
                </Stack>

                {heroEvent && (
                  <Box sx={{
                    gridColumn: { xs: '1 / -1', md: '9 / -1' },
                    justifySelf: { md: 'end' },
                    alignSelf: { md: 'flex-end' },
                    mt: { xs: 2, md: 0 },
                    width: '100%'
                  }}>
                    <Paper
                      elevation={0}
                      sx={{
                        ...custom.glassCard,
                        ...custom.eventFloat,
                        p: 1.6,
                        transform: { md: 'rotate(-1deg)' },
                      }}
                    >
                      <Stack spacing={1}>
                        <Typography variant="overline" sx={{ letterSpacing: '0.22em', color: 'var(--accent-strong)', fontWeight: 700, lineHeight: 1 }}>
                          Nächstes Event
                        </Typography>
                        <Typography variant="h6" fontWeight={700} sx={{ fontSize: '1.1rem' }}>
                          {heroEvent.title}
                        </Typography>
                        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
                          <Typography variant="caption" sx={{ color: 'var(--muted)' }}>
                            {heroEventDate ? formatShortDate(heroEventDate) : heroEvent.date}
                          </Typography>
                          {heroEvent.time && (
                            <Typography variant="caption" sx={{ color: 'var(--muted)' }}>{heroEvent.time} Uhr</Typography>
                          )}
                          {heroEvent.location && (
                            <Stack direction="row" spacing={0.3} alignItems="center">
                              <LocationOnRounded sx={{ fontSize: 14, color: 'var(--muted)' }} />
                              <Typography variant="caption" sx={{ color: 'var(--muted)' }}>{heroEvent.location}</Typography>
                            </Stack>
                          )}
                        </Stack>
                        <Button component="a" href="#events" size="small" sx={{ alignSelf: 'flex-start', textTransform: 'none', color: 'var(--accent-strong)', px: 0 }}>
                          Zum Kalender
                        </Button>
                      </Stack>
                    </Paper>
                  </Box>
                )}
              </Box>
            </Paper>
          </Box>

          <Box component="section" data-reveal="true" 
            sx={{ 
              ...custom.sectionShell, 
              gridColumn: '1 / -1', 
              background: custom.sectionGradients.newsroom,
              mb: { xs: 6, md: 10 } 
            }}>
            <Box sx={{ position: 'relative', zIndex: 1 }}>
              <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 2 }}>
                <Stack direction="row" spacing={1} alignItems="center">
                  <CampaignRounded sx={custom.sectionIcon} />
                  <Typography variant="h4" sx={custom.sectionTitle}>Newsroom</Typography>
                </Stack>
                <Button component={RouterLink} to="/news" size="small" endIcon={<ArrowForwardRounded fontSize="small" />} sx={custom.sectionAction}>
                  Alle News
                </Button>
              </Stack>
              <Box sx={{ 
                display: 'grid', 
                gap: { xs: 3, md: 4 }, 
                gridTemplateColumns: { xs: '1fr', md: 'repeat(12, minmax(0, 1fr))' }, 
                alignItems: 'stretch' 
              }}>
                <Box sx={{ gridColumn: { xs: '1 / -1', md: 'span 7' } }}>
                  {leadNews ? (
                    <Paper
                      component={RouterLink}
                      to={`/news/${leadNews.id}`}
                      elevation={0}
                      sx={{
                        ...custom.glassCard,
                        display: 'flex', flexDirection: 'column', position: 'relative', minHeight: { xs: 220, md: 280 }, borderRadius: 5, overflow: 'hidden', textDecoration: 'none', color: 'inherit', background: 'none', backgroundColor: imageOverlayBase, height: '100%',
                        '&::before': {
                          content: '""', position: 'absolute', inset: -1,
                          backgroundImage: leadNews.image
                            ? `linear-gradient(180deg, ${alpha(imageOverlayBase, 0.6)} 0%, ${alpha(imageOverlayBase, 1)} 100%), url(${leadNews.image})`
                            : `linear-gradient(135deg, ${alpha(imageOverlayBase, 0.6)}, ${alpha(imageOverlayBase, 0.98)})`,
                          backgroundSize: leadNews.image ? '100% 100%, cover' : 'cover', backgroundPosition: 'center', backgroundRepeat: 'no-repeat', zIndex: 0, pointerEvents: 'none',
                        },
                      }}
                    >
                      <Stack spacing={0.8} sx={{ position: 'relative', zIndex: 1, p: { xs: 2, md: 2.6 }, height: '100%', justifyContent: 'flex-end' }}>
                        <Typography variant="overline" sx={{ letterSpacing: '0.18em', color: alpha(theme.palette.common.white, 0.85) }}>
                          {formatNewsDate(leadNews.date)}
                        </Typography>
                        <Typography variant="h4" sx={{ fontWeight: 700, color: theme.palette.common.white, lineHeight: 1.15 }}>{leadNews.title}</Typography>
                        <Typography variant="body2" sx={{ color: alpha(theme.palette.common.white, 0.75), display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}>
                          {leadNews.summary || 'Mehr lesen …'}
                        </Typography>
                      </Stack>
                    </Paper>
                  ) : (
                    <Paper elevation={0} sx={{ ...custom.glassCard, p: { xs: 2, md: 2.6 }, borderRadius: 5 }}>
                      <Typography variant="body1" sx={{ color: 'var(--muted)' }}>Weitere News folgen in Kürze.</Typography>
                    </Paper>
                  )}
                </Box>
                <Paper elevation={0} sx={{ ...custom.glassCard, gridColumn: { xs: '1 / -1', md: 'span 5' }, p: { xs: 1.8, md: 2.2 }, borderRadius: 4, background: alpha(theme.palette.background.paper, isDark ? 0.72 : 0.9) }}>
                  <Stack spacing={2}> 
                    <Typography variant="overline" sx={{ letterSpacing: '0.24em', color: 'var(--accent-strong)', fontWeight: 700 }}>Kurz &amp; Knapp</Typography>
                    {secondaryNews.length ? secondaryNews.map((item) => (
                      <Box key={item.id} component={RouterLink} to={`/news/${item.id}`} sx={custom.newsCardLink}>
                        <Typography variant="caption" sx={{ color: 'var(--muted)', letterSpacing: '0.12em' }}>{formatNewsDate(item.date)}</Typography>
                        <Typography variant="subtitle1" sx={{ fontWeight: 700, lineHeight: 1.2 }}>{item.title}</Typography>
                      </Box>
                    )) : <Typography variant="body2" sx={{ color: 'var(--muted)' }}>Keine weiteren News.</Typography>}
                    <Button component={RouterLink} to="/news" size="small" endIcon={<ArrowForwardRounded fontSize="small" />} sx={{ ...custom.sectionAction, alignSelf: 'flex-start', px: 0 }}>
                      Alle News
                    </Button>
                  </Stack>
                </Paper>
              </Box>
              {railItems.length ? (
                <Box data-reveal="rail" sx={{ mt: { xs: 3, md: 5 }, ...custom.railReveal }}> 
                  <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 1.6 }}>
                    <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>Mehr Stories</Typography>
                  </Stack>
                  <Box sx={custom.scrollableRail}>
                    {railItems.map((item) => (
                      <Paper key={item.id} component={RouterLink} to={`/news/${item.id}`} elevation={0} sx={{ ...custom.glassCard, borderRadius: 4, overflow: 'hidden', textDecoration: 'none', color: 'inherit', borderColor: 'var(--card-border)', boxShadow: 'none', '&:hover': { transform: 'translateY(-3px)', borderColor: 'var(--accent-strong)', boxShadow: 'none' } }}>
                        <Box sx={{ height: 130, backgroundImage: item.image ? `url(${item.image})` : isDark ? `linear-gradient(135deg, ${alpha(theme.palette.background.default, 0.96)}, ${alpha(theme.palette.background.default, 0.98)})` : `linear-gradient(135deg, ${alpha(theme.palette.primary.light, 0.18)}, ${alpha(theme.palette.background.paper, 0.98)})`, backgroundSize: 'cover', backgroundPosition: 'center' }} />
                        <Stack spacing={0.6} sx={{ p: 1.6 }}>
                          <Typography variant="caption" sx={{ color: 'var(--muted)', letterSpacing: '0.12em' }}>{formatNewsDate(item.date)}</Typography>
                          <Typography variant="subtitle1" sx={{ fontWeight: 700, lineHeight: 1.2 }}>{item.title}</Typography>
                          <Typography variant="body2" sx={{ color: 'var(--muted)', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}>{item.summary || 'Mehr lesen …'}</Typography>
                        </Stack>
                      </Paper>
                    ))}
                  </Box>
                </Box>
              ) : null}
            </Box>
          </Box>

          <Box component="section" id="events" data-reveal="true" 
            sx={{ 
              ...custom.sectionShell, 
              gridColumn: '1 / -1', 
              scrollMarginTop: { xs: 80, md: 100 }, 
              background: custom.sectionGradients.agenda,
              mb: { xs: 6, md: 10 } 
            }}>
            <Box sx={{ position: 'relative', zIndex: 1 }}>
              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} alignItems={{ xs: 'flex-start', sm: 'center' }} justifyContent="space-between" sx={{ mb: 2 }}>
                <Stack direction="row" spacing={1} alignItems="center">
                  <EventAvailableRounded sx={custom.sectionIcon} />
                  <Typography variant="h4" sx={custom.sectionTitle}>Agenda</Typography>
                </Stack>

                {isAdmin && (
                  <Button
                    size="small"
                    variant="contained"
                    startIcon={<AddRounded fontSize="small" />}
                    onClick={() => openEventDialog()}
                    sx={{ ...custom.roundButton, bgcolor: 'var(--accent-strong)', color: theme.palette.primary.contrastText, '&:hover': { bgcolor: isDark ? theme.palette.primary.light : theme.palette.primary.dark } }}
                  >
                    Event eintragen
                  </Button>
                )}
              </Stack>
              <Box sx={{ 
                display: 'grid', 
                gap: { xs: 3, md: 4 }, 
                gridTemplateColumns: { xs: '1fr', lg: '7fr 5fr' }, 
                alignItems: 'start' 
              }}>
                <Box>
                  <Typography variant="caption" sx={{ color: 'var(--muted)', fontWeight: 700, letterSpacing: '0.22em', textTransform: 'uppercase', display: 'block', mb: 1.2 }}>KOMMENDE TERMINE</Typography>
                  <Box sx={{ position: 'relative', pl: { xs: 2.5, md: 3.5 } }}>
                    <Box sx={{ position: 'absolute', left: { xs: 10, md: 14 }, top: 6, bottom: 6, width: 2, bgcolor: 'var(--card-border)', borderRadius: 999 }} />
                    <Stack spacing={2.5}> 
                      {upcomingPreview.length ? upcomingPreview.map((event) => {
                        const accent = getCategoryColor(event.category);
                        const eventDate = parseDateKey(event.date);
                        const isActive = event.date === selectedDate;
                        return (
                          <Box key={event.id} component="button" type="button" onClick={() => selectDate(event.date)} sx={{ appearance: 'none', border: 'none', background: 'transparent', textAlign: 'left', padding: 0, cursor: 'pointer', width: '100%', fontFamily: 'inherit', '&:focus-visible': { outline: '2px solid var(--accent-3)', outlineOffset: 3 } }}>
                            <Box sx={{ ...custom.glassCard, p: { xs: 1.6, md: 2 }, borderRadius: 4, borderColor: isActive ? 'var(--accent-strong)' : 'transparent', background: isActive ? 'var(--accent-soft)' : isDark ? alpha(theme.palette.background.paper, 0.72) : alpha(theme.palette.background.paper, 0.94), position: 'relative', color: theme.palette.text.primary }}>
                              <Box sx={custom.timelineDot(accent)} />
                              <Stack spacing={0.6}>
                                <Typography variant="overline" sx={{ letterSpacing: '0.18em', color: 'var(--muted)' }}>{formatShortDate(eventDate)}</Typography>
                                <Typography variant="h6" sx={{ fontWeight: 700 }}>{event.title}</Typography>
                                <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
                                  {event.time && <Typography variant="caption" sx={{ color: 'var(--muted)' }}>{event.time} Uhr</Typography>}
                                  {event.location && (
                                    <Stack direction="row" spacing={0.3} alignItems="center">
                                      <LocationOnRounded sx={{ fontSize: 13, color: 'var(--muted)' }} />
                                      <Typography variant="caption" sx={{ color: 'var(--muted)' }}>{event.location}</Typography>
                                    </Stack>
                                  )}
                                  <Typography variant="caption" sx={{ color: accent, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.08em' }}>{event.category}</Typography>
                                </Stack>
                              </Stack>
                            </Box>
                          </Box>
                        );
                      }) : <Typography variant="body2" sx={{ color: 'var(--muted)' }}>Noch keine Events eingetragen.</Typography>}
                    </Stack>
                  </Box>
                </Box>
                <Paper elevation={0} sx={{ ...custom.glassCard, mt: { xs: 2.6, md: 3.2 }, p: { xs: 1.6, md: 2 }, borderRadius: 4, background: alpha(theme.palette.background.paper, 0.92) }}>
                  <Stack spacing={1}>
                    <Stack direction="row" alignItems="center" justifyContent="space-between">
                      <Stack direction="row" spacing={0.6} alignItems="center">
                        <CalendarMonthRounded sx={custom.sectionIcon} />
                        <Typography variant="subtitle1" fontWeight={700}>Kalender</Typography>
                      </Stack>
                      <Stack direction="row" spacing={0.3} alignItems="center">
                        <IconButton size="small" onClick={() => changeMonth(-1)} sx={{ p: 0.3 }}><ChevronLeftRounded fontSize="small" /></IconButton>
                        <IconButton size="small" onClick={() => changeMonth(1)} sx={{ p: 0.3 }}><ChevronRightRounded fontSize="small" /></IconButton>
                      </Stack>
                    </Stack>
                    <Typography variant="caption" sx={{ color: 'var(--muted)' }}>{monthLabel(currentMonth)}</Typography>
                    <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(7, minmax(0, 1fr))', gap: 0.3, textAlign: 'center' }}>
                      {['Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa', 'So'].map((label) => (
                        <Typography key={label} variant="caption" sx={{ fontWeight: 700, color: 'var(--muted)' }}>{label}</Typography>
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
                          <Box key={`${dateKey}-${idx}`} component="button" type="button" onClick={() => selectDate(dateKey)} sx={custom.calendarDay(isSelected, isToday, inMonth)}>
                            <Typography variant="caption" fontWeight={700} sx={{ fontSize: '0.72rem' }}>{day}</Typography>
                            {hasEvents && <Box sx={{ width: 4, height: 4, borderRadius: '50%', bgcolor: getCategoryColor(dayEvents[0].category), opacity: 0.9 }} />}
                          </Box>
                        );
                      })}
                    </Box>
                  </Stack>
                </Paper>
              </Box>
            </Box>
          </Box>

          <Box component="section" data-reveal="true" sx={{ ...custom.sectionShell, gridColumn: '1 / -1', background: custom.sectionGradients.forum }}>
            <Box sx={{ position: 'relative', zIndex: 1 }}>
              <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 2 }}>
                <Stack direction="row" spacing={1} alignItems="center">
                  <ForumRounded sx={custom.sectionIcon} />
                  <Typography variant="h4" sx={custom.sectionTitle}>Forum</Typography>
                </Stack>
                <Button component={RouterLink} to="/forum" size="small" endIcon={<ArrowForwardRounded fontSize="small" />} sx={custom.sectionAction}>
                  Zum Forum
                </Button>
              </Stack>
              {forumHighlight ? (
                <Box sx={{ 
                  display: 'grid', 
                  gap: { xs: 3, md: 4 }, 
                  gridTemplateColumns: { xs: '1fr', md: 'repeat(12, minmax(0, 1fr))' } 
                }}>
                  <Paper component={RouterLink} to="/forum" elevation={0} sx={{ ...custom.glassCard, gridColumn: { xs: '1 / -1', md: 'span 7' }, p: { xs: 1.6, md: 3 }, borderRadius: 5, textDecoration: 'none', color: 'inherit', background: isDark ? `linear-gradient(135deg, ${alpha(theme.palette.background.paper, 0.92)}, ${alpha(theme.palette.background.default, 0.9)})` : `linear-gradient(135deg, ${alpha(theme.palette.info.light, 0.2)}, ${alpha(theme.palette.background.paper, 0.98)})` }}>
                    <Stack spacing={{ xs: 1, md: 1.4 }}>
                      <Stack direction="row" spacing={1.2} alignItems="center" sx={{ flexWrap: { xs: 'wrap', sm: 'nowrap' }, rowGap: { xs: 0.4, sm: 0 } }}>
                        <Avatar sx={{ width: { xs: 36, md: 42 }, height: { xs: 36, md: 42 }, bgcolor: 'var(--accent-soft)', color: 'var(--accent-strong)' }}>
                          {forumHighlight.author?.charAt(0)?.toUpperCase() || '?'}
                        </Avatar>
                        <Box>
                          <Typography variant="overline" sx={{ letterSpacing: { xs: '0.16em', md: '0.22em' }, color: 'var(--accent-strong)' }}>Top Diskussion</Typography>
                          <Typography variant="caption" sx={{ color: 'var(--muted)' }}>{forumHighlight.author} · {formatForumDate(forumHighlight.createdAt)}</Typography>
                        </Box>
                      </Stack>
                      <Typography variant="h4" sx={{ fontWeight: 700, lineHeight: 1.2, fontSize: { xs: '1.35rem', sm: '1.6rem', md: '2rem' } }}>{forumHighlight.title}</Typography>
                      <Typography variant="body2" sx={{ color: 'var(--muted)' }}>{forumHighlight.replies} Antworten</Typography>
                      <Stack direction="row" spacing={0.4} alignItems="center" sx={{ flexWrap: 'wrap', rowGap: 0.2 }}>
                        <Typography variant="caption" sx={{ color: 'var(--accent-strong)', fontWeight: 600 }}>Zum Beitrag</Typography>
                        <ArrowForwardRounded fontSize="small" sx={{ color: 'var(--accent-strong)' }} />
                      </Stack>
                    </Stack>
                  </Paper>
                  <Stack spacing={{ xs: 2, md: 2.5 }} sx={{ gridColumn: { xs: '1 / -1', md: 'span 5' } }}> 
                    {forumRest.length ? forumRest.map((post) => (
                      <Box key={post.id} component={RouterLink} to="/forum" sx={{ ...custom.glassCard, p: { xs: 1.2, md: 1.6 }, borderRadius: 3, textDecoration: 'none', color: 'inherit', display: 'grid', gap: 0.6 }}>
                        <Stack direction="row" spacing={1} alignItems="center">
                          <Avatar sx={{ width: { xs: 30, md: 34 }, height: { xs: 30, md: 34 }, fontSize: '0.8rem' }}>{post.author?.charAt(0)?.toUpperCase() || '?'}</Avatar>
                          <Box sx={{ minWidth: 0 }}>
                            <Typography variant="subtitle1" sx={{ fontWeight: 600, fontSize: { xs: '0.95rem', md: '1.05rem' }, display: '-webkit-box', WebkitLineClamp: { xs: 2, md: 1 }, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}>{post.title}</Typography>
                            <Typography variant="caption" sx={{ color: 'var(--muted)' }}>{post.author} · {formatForumDate(post.createdAt)}</Typography>
                          </Box>
                        </Stack>
                        <Typography variant="caption" sx={{ color: 'var(--muted)' }}>{post.replies} Antworten</Typography>
                      </Box>
                    )) : <Typography variant="body2" sx={{ color: 'var(--muted)' }}>Keine weiteren Threads.</Typography>}
                  </Stack>
                </Box>
              ) : <Typography variant="body2" sx={{ color: 'var(--muted)' }}>Noch keine Beiträge.</Typography>}
            </Box>
          </Box>

        </Box>
      </Box>

      <Dialog open={eventDialogOpen} onClose={() => setEventDialogOpen(false)} fullWidth maxWidth="sm" PaperProps={{ sx: { borderRadius: 3, background: theme.palette.background.paper } }}>
        <DialogTitle>Event eintragen</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField label="Titel" value={eventDraft.title} onChange={(e) => setEventDraft((p) => ({ ...p, title: e.target.value }))} fullWidth />
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
              <TextField label="Datum" type="date" value={eventDraft.date} onChange={(e) => setEventDraft((p) => ({ ...p, date: e.target.value }))} fullWidth InputLabelProps={{ shrink: true }} />
              <TextField label="Uhrzeit" type="time" value={eventDraft.time} onChange={(e) => setEventDraft((p) => ({ ...p, time: e.target.value }))} fullWidth InputLabelProps={{ shrink: true }} />
            </Stack>
            <TextField label="Ort" value={eventDraft.location} onChange={(e) => setEventDraft((p) => ({ ...p, location: e.target.value }))} fullWidth />
            <TextField label="Kategorie" select value={eventDraft.category} onChange={(e) => setEventDraft((p) => ({ ...p, category: e.target.value }))} fullWidth>
              {EVENT_CATEGORIES.map((cat) => <MenuItem key={cat.value} value={cat.value}>{cat.label}</MenuItem>)}
            </TextField>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEventDialogOpen(false)}>Abbrechen</Button>
          <Button variant="contained" onClick={handleCreateEvent} disabled={!eventDraft.title.trim() || !eventDraft.date}>Speichern</Button>
        </DialogActions>
      </Dialog>
    </Sidebar>
  );
};

export default NewsFeedPage;