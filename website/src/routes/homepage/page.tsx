import React, { useEffect, useMemo, useState } from 'react';
import {
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
  Grid,
  Divider,
  List,
  ListItem,
  ListItemText,
} from '@mui/material';
import AddRounded from '@mui/icons-material/AddRounded';
import ArrowForwardRounded from '@mui/icons-material/ArrowForwardRounded';
import CalendarMonthRounded from '@mui/icons-material/CalendarMonthRounded';
import ForumRounded from '@mui/icons-material/ForumRounded';
import ChevronLeftRounded from '@mui/icons-material/ChevronLeftRounded';
import ChevronRightRounded from '@mui/icons-material/ChevronRightRounded';
import CampaignRounded from '@mui/icons-material/CampaignRounded';
import { Link as RouterLink } from 'react-router-dom';

import { Sidebar } from '@components/layout';
import { useAuth } from '@lib/auth';
import {
  NEWS_DATA,
  FORUM_SEED_POSTS,
  FORUM_STORAGE_KEY,
  CALENDAR_EVENT_CATEGORIES as EVENT_CATEGORIES,
  STORAGE_KEYS
} from '@lib/data';

const STORAGE_EVENTS_KEY = STORAGE_KEYS.HOMEPAGE_EVENTS;

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

const buildCalendarDays = (date: Date) => {
  const year = date.getFullYear();
  const month = date.getMonth();
  const firstDayOfMonth = new Date(year, month, 1);
  const startDay = (firstDayOfMonth.getDay() + 6) % 7;
  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const daysInPrevMonth = new Date(year, month, 0).getDate();

  const days: { date: Date; day: number; inMonth: boolean }[] = [];

  for (let i = startDay - 1; i >= 0; i--) {
    days.push({
      date: new Date(year, month - 1, daysInPrevMonth - i),
      day: daysInPrevMonth - i,
      inMonth: false,
    });
  }

  for (let i = 1; i <= daysInMonth; i++) {
    days.push({
      date: new Date(year, month, i),
      day: i,
      inMonth: true,
    });
  }

  const remaining = (7 - (days.length % 7)) % 7;
  for (let i = 1; i <= remaining; i++) {
    days.push({
      date: new Date(year, month + 1, i),
      day: i,
      inMonth: false,
    });
  }

  return days;
};

const NewsFeedPage: React.FC = () => {
  const { user } = useAuth();
  const [events, setEvents] = useState<CalendarEvent[]>(() => {
    const savedEvents = localStorage.getItem(STORAGE_EVENTS_KEY);
    if (savedEvents) {
      try {
        return JSON.parse(savedEvents);
      } catch (e) {
        console.error('Events parse error', e);
      }
    }
    return [];
  });
  const [selectedDate, setSelectedDate] = useState<string>(toDateKey(new Date()));
  const [currentMonth, setCurrentMonth] = useState<Date>(new Date());
  const [eventDialogOpen, setEventDialogOpen] = useState(false);
  const [eventDraft, setEventDraft] = useState<EventDraft>({
    title: '',
    date: '',
    time: '',
    location: '',
    category: 'Treffen',
  });

  const [forumPosts] = useState<ForumPostSummary[]>(() => {
    const savedForum = localStorage.getItem(FORUM_STORAGE_KEY);
    return savedForum ? JSON.parse(savedForum) : FORUM_SEED_POSTS;
  });

  useEffect(() => {
  }, []);

  useEffect(() => {
    localStorage.setItem(STORAGE_EVENTS_KEY, JSON.stringify(events));
  }, [events]);

  const days = useMemo(() => buildCalendarDays(currentMonth), [currentMonth]);
  const todayKey = toDateKey(new Date());

  const eventsByDate = useMemo(() => {
    const map = new Map<string, CalendarEvent[]>();
    events.forEach((ev) => {
      const list = map.get(ev.date) || [];
      list.push(ev);
      map.set(ev.date, list);
    });
    return map;
  }, [events]);

  const selectedDateEvents = useMemo(() => eventsByDate.get(selectedDate) || [], [eventsByDate, selectedDate]);

  const latestNews = NEWS_DATA.slice(0, 5);
  const latestForum = forumPosts.slice(0, 3);

  const openEventDialog = (dateKey?: string) => {
    setEventDraft((prev) => ({
      ...prev,
      date: dateKey ?? selectedDate,
    }));
    setEventDialogOpen(true);
  };

  const handleCreateEvent = () => {
    if (!eventDraft.title.trim() || !eventDraft.date) return;
    const id = `event-${Date.now()}`;
    const nextEventEntry: CalendarEvent = {
      id,
      title: eventDraft.title.trim(),
      date: eventDraft.date,
      time: eventDraft.time,
      location: eventDraft.location.trim(),
      category: eventDraft.category,
    };
    setEvents((prev) => [...prev, nextEventEntry].sort((a, b) => parseDateKey(a.date).getTime() - parseDateKey(b.date).getTime()));
    setEventDialogOpen(false);
    setEventDraft((prev) => ({ ...prev, title: '', time: '', location: '' }));
  };

  const changeMonth = (offset: number) => {
    setCurrentMonth((prev) => new Date(prev.getFullYear(), prev.getMonth() + offset, 1));
  };

  return (
    <Sidebar user={user} title="Startseite">
      <Stack spacing={4}>
        <Box>
          <Typography variant="h4" component="h1" fontWeight={700} gutterBottom>
            Fachschaft Informatik
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Willkommen auf der offiziellen Seite der Fachschaft Informatik der Westfälischen Hochschule.
          </Typography>
        </Box>

        <Grid container spacing={4}>
          <Grid size={{ xs: 12, md: 8 }}>
            <Stack spacing={3}>
              <Stack direction="row" alignItems="center" justifyContent="space-between" px={1}>
                <Stack direction="row" spacing={1} alignItems="center">
                  <CampaignRounded color="primary" />
                  <Typography variant="h6" color="primary.main" fontWeight={600}>Newsroom</Typography>
                </Stack>
                <Button component={RouterLink} to="/news" size="small" endIcon={<ArrowForwardRounded />}> Alle News</Button>
              </Stack>

              {latestNews.map((item) => (
                <Paper key={item.id} variant="outlined" sx={{ borderRadius: 3, overflow: 'hidden' }}>
                  <Grid container>
                    <Grid size={{ xs: 12, sm: 4 }}>
                      <Box
                        component="img"
                        src={item.image}
                        alt={item.title}
                        sx={{
                          width: '100%',
                          height: '100%',
                          minHeight: 140,
                          objectFit: 'cover'
                        }}
                      />
                    </Grid>
                    <Grid size={{ xs: 12, sm: 8 }}>
                      <Box sx={{ p: 2 }}>
                        <Typography variant="caption" color="text.secondary" gutterBottom display="block">
                          {item.date}
                        </Typography>
                        <Typography variant="h6" fontWeight={700} gutterBottom component={RouterLink} to="/news" sx={{ textDecoration: 'none', color: 'inherit', '&:hover': { color: 'primary.main' } }}>
                          {item.title}
                        </Typography>
                        <Typography variant="body2" color="text.secondary" sx={{
                          display: '-webkit-box',
                          WebkitLineClamp: 2,
                          WebkitBoxOrient: 'vertical',
                          overflow: 'hidden'
                        }}>
                          {item.summary}
                        </Typography>
                      </Box>
                    </Grid>
                  </Grid>
                </Paper>
              ))}
            </Stack>
          </Grid>

          <Grid size={{ xs: 12, md: 4 }}>
            <Stack spacing={4}>
              <Paper variant="outlined" sx={{ p: 3, borderRadius: 3 }}>
                <Stack spacing={3}>
                  <Stack direction="row" alignItems="center" justifyContent="space-between">
                    <Stack direction="row" spacing={1} alignItems="center">
                      <CalendarMonthRounded color="primary" />
                      <Typography variant="h6" color="primary.main" fontWeight={600}>Agenda</Typography>
                    </Stack>
                    <IconButton size="small" onClick={() => openEventDialog()} color="primary">
                      <AddRounded />
                    </IconButton>
                  </Stack>

                  <Box>
                    <Stack direction="row" alignItems="center" justifyContent="space-between" mb={1}>
                      <Typography variant="subtitle2" fontWeight={700}>{monthLabel(currentMonth)}</Typography>
                      <Stack direction="row" spacing={0.5}>
                        <IconButton size="small" onClick={() => changeMonth(-1)}><ChevronLeftRounded fontSize="small" /></IconButton>
                        <IconButton size="small" onClick={() => changeMonth(1)}><ChevronRightRounded fontSize="small" /></IconButton>
                      </Stack>
                    </Stack>
                    <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: 0.5, textAlign: 'center', mb: 1 }}>
                      {['M', 'D', 'M', 'D', 'F', 'S', 'S'].map((d, i) => (
                        <Typography key={i} variant="caption" fontWeight={700} color="text.secondary">{d}</Typography>
                      ))}
                    </Box>
                    <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: 0.5 }}>
                      {days.map(({ date, day, inMonth }, idx) => {
                        const dateKey = toDateKey(date);
                        const isToday = dateKey === todayKey;
                        const isSelected = dateKey === selectedDate;
                        const hasEvents = eventsByDate.has(dateKey);
                        return (
                          <Box
                            key={idx}
                            component="button"
                            onClick={() => setSelectedDate(dateKey)}
                            sx={{
                              border: '1px solid',
                              borderColor: isSelected ? 'primary.main' : isToday ? 'primary.light' : 'transparent',
                              bgcolor: isSelected ? 'primary.light' : 'transparent',
                              borderRadius: 1,
                              p: 0.5,
                              cursor: 'pointer',
                              opacity: inMonth ? 1 : 0.3,
                              display: 'flex',
                              flexDirection: 'column',
                              alignItems: 'center',
                              minHeight: 32,
                              color: 'inherit',
                              transition: 'all 0.2s',
                              '&:hover': { border: '1px solid', borderColor: 'primary.main' }
                            }}
                          >
                            <Typography variant="caption" fontWeight={isSelected ? 700 : 400}>{day}</Typography>
                            {hasEvents && <Box sx={{ width: 4, height: 4, borderRadius: '50%', bgcolor: 'primary.main', mt: 0.2 }} />}
                          </Box>
                        );
                      })}
                    </Box>
                  </Box>

                  <Divider />

                  <Stack spacing={2}>
                    <Typography variant="subtitle2" fontWeight={700}>Events am {new Date(selectedDate).toLocaleDateString('de-DE')}</Typography>
                    {selectedDateEvents.length > 0 ? (
                      selectedDateEvents.map(ev => (
                        <Box key={ev.id}>
                          <Typography variant="body2" fontWeight={600}>{ev.title}</Typography>
                          <Typography variant="caption" color="text.secondary">
                            {ev.time && `${ev.time} Uhr `}{ev.location && `· ${ev.location}`}
                          </Typography>
                        </Box>
                      ))
                    ) : (
                      <Typography variant="caption" color="text.secondary">Keine Termine.</Typography>
                    )}
                  </Stack>
                </Stack>
              </Paper>

              <Paper variant="outlined" sx={{ p: 3, borderRadius: 3 }}>
                <Stack spacing={2}>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <ForumRounded color="primary" />
                    <Typography variant="h6" color="primary.main" fontWeight={600}>Forum</Typography>
                  </Stack>
                  <List disablePadding>
                    {latestForum.map((post, idx) => (
                      <React.Fragment key={post.id}>
                        {idx > 0 && <Divider sx={{ my: 1 }} />}
                        <ListItem disableGutters component={RouterLink} to="/forum" sx={{ textDecoration: 'none', color: 'inherit', '&:hover': { color: 'primary.main' } }}>
                          <ListItemText
                            primary={post.title}
                            primaryTypographyProps={{ variant: 'body2', fontWeight: 600, noWrap: true }}
                            secondary={`${post.author} · ${post.replies} Antworten`}
                            secondaryTypographyProps={{ variant: 'caption' }}
                          />
                        </ListItem>
                      </React.Fragment>
                    ))}
                  </List>
                  <Button component={RouterLink} to="/forum" fullWidth size="small" variant="outlined" sx={{ mt: 1 }}>Zum Forum</Button>
                </Stack>
              </Paper>
            </Stack>
          </Grid>
        </Grid>
      </Stack>

      <Dialog open={eventDialogOpen} onClose={() => setEventDialogOpen(false)} fullWidth maxWidth="xs">
        <DialogTitle>Event eintragen</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField label="Titel" value={eventDraft.title} onChange={(e) => setEventDraft(p => ({ ...p, title: e.target.value }))} fullWidth size="small" />
            <TextField label="Datum" type="date" value={eventDraft.date} onChange={(e) => setEventDraft(p => ({ ...p, date: e.target.value }))} fullWidth size="small" InputLabelProps={{ shrink: true }} />
            <TextField label="Uhrzeit" type="time" value={eventDraft.time} onChange={(e) => setEventDraft(p => ({ ...p, time: e.target.value }))} fullWidth size="small" InputLabelProps={{ shrink: true }} />
            <TextField label="Ort" value={eventDraft.location} onChange={(e) => setEventDraft(p => ({ ...p, location: e.target.value }))} fullWidth size="small" />
            <TextField label="Kategorie" select value={eventDraft.category} onChange={(e) => setEventDraft(p => ({ ...p, category: e.target.value }))} fullWidth size="small">
              {EVENT_CATEGORIES.map(c => <MenuItem key={c.value} value={c.value}>{c.label}</MenuItem>)}
            </TextField>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEventDialogOpen(false)}>Abbrechen</Button>
          <Button variant="contained" onClick={handleCreateEvent}>Speichern</Button>
        </DialogActions>
      </Dialog>
    </Sidebar >
  );
};

export default NewsFeedPage;
