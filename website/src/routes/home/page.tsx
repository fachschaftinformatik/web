import React, { useEffect, useMemo, useState } from 'react';
import {
  Avatar,
  Box,
  Button,
  Paper,
  Stack,
  Typography,
  alpha,
  useTheme,
} from '@mui/material';
import ArrowForwardRounded from '@mui/icons-material/ArrowForwardRounded';
import GroupsRounded from '@mui/icons-material/GroupsRounded';
import TerminalRounded from '@mui/icons-material/TerminalRounded';
import CelebrationRounded from '@mui/icons-material/CelebrationRounded';
import InfoRounded from '@mui/icons-material/InfoRounded';
import MoreHorizRounded from '@mui/icons-material/MoreHorizRounded';
import {
  Timeline,
  TimelineItem,
  TimelineSeparator,
  TimelineConnector,
  TimelineContent,
  TimelineDot,
  TimelineOppositeContent,
} from '@mui/lab';
import { useNavigate, Link as RouterLink, Link } from 'react-router-dom';

import { Sidebar } from '@components/layout';
import { useAuth } from '@lib/auth';
import { getDiscussions, getEvents } from '@lib/api';
import type { DtoDiscussionPostResponse as ApiPost, DtoEventResponse } from '@lib/api';
import { getSizedImageUrl, getImageSrcSet, getAvatarUrl } from '@lib/images';

type NewsItem = ApiPost;

type EventItem = DtoEventResponse;
type ForumPostSummary = { id: string; title: string; body: string; author: string; authorId: string; authorAvatarUrl: string | undefined; createdAt: string; replies: number; };
type CalendarEvent = { id: string; title: string; date: string; time?: string; location?: string; category: string; };

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
const formatShortDate = (date: Date) => date.toLocaleDateString('de-DE', { day: '2-digit', month: 'short' });
const formatNewsDate = (dateString: string) => {
  const parsed = new Date(dateString);
  return isNaN(parsed.getTime()) ? dateString : formatShortDate(parsed);
};
const formatForumDate = (iso?: string) => {
  if (!iso) return 'k.A.';
  const d = new Date(iso);
  return isNaN(d.getTime()) ? 'k.A.' : d.toLocaleDateString('de-DE', { day: '2-digit', month: 'short' });
};
const loadForumPosts = (apiPosts: ApiPost[] = []): ForumPostSummary[] => {
  const normalize = (post: ApiPost): ForumPostSummary => ({
    id: String(post.id ?? ''),
    title: String(post.title ?? 'Neuer Beitrag'),
    body: String(post.body ?? ''),
    author: String(post.user_name ?? 'Anonym'),
    authorId: String(post.user_id ?? ''),
    authorAvatarUrl: post.user_avatar_url || undefined,
    createdAt: String(post.created_at ?? new Date().toISOString()),
    replies: Number(post.comment_count ?? 0),
  });
  return apiPosts.map(normalize);
};
const toEventTimestamp = (event: CalendarEvent) => {
  const [year, month, day] = event.date.split('-').map(Number);
  const [hour, minute] = event.time ? event.time.split(':').map(Number) : [23, 59];
  return new Date(year, (month || 1) - 1, day || 1, hour || 0, minute || 0).getTime();
};

const NewsFeedPage: React.FC = () => {
  const { user } = useAuth();
  const navigate = useNavigate();
  const theme = useTheme();
  const custom = theme.custom;
  const isDark = theme.palette.mode === 'dark';

  const getCategoryColor = (category: string) => {
    if (category === 'Treffen') return theme.palette.success.main;
    if (category === 'Workshop') return theme.palette.info.main;
    if (category === 'Party') return theme.palette.error.main;
    if (category === 'Info') return theme.palette.warning.main;
    return theme.palette.primary.main;
  };

  const [eventsData, setEventsData] = useState<EventItem[]>([]);
  const [carouselIndex, setCarouselIndex] = useState(0);
  const [sidebarIndex, setSidebarIndex] = useState(0);
  const [newsItems, setNewsItems] = useState<NewsItem[]>([]);
  const [forumPosts, setForumPosts] = useState<ForumPostSummary[]>([]);
  const [apiEvents, setApiEvents] = useState<ApiPost[]>([]);

  useEffect(() => {
    getDiscussions({ query: { type: 'news', limit: 10 } }).then(({ data }) => {
      if (data) setNewsItems(data as ApiPost[]);
    });
    getDiscussions({ query: { type: 'discussion', limit: 5 } }).then(({ data }) => {
      if (data) setForumPosts(loadForumPosts(data as ApiPost[]));
    });
    getDiscussions({ query: { type: 'event', limit: 20 } }).then(({ data }) => {
      if (data) setApiEvents(data as ApiPost[]);
    });

    getEvents().then(({ data }) => {
      if (data) setEventsData(data as EventItem[]);
    });
  }, []);

  useEffect(() => {
    if (eventsData.length <= 1) return;
    const interval = setInterval(() => {
      setCarouselIndex((prev) => (prev + 1) % eventsData.length);
    }, 5000);
    return () => clearInterval(interval);
  }, [eventsData.length]);

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

  const upcomingEvents = useMemo(() => {
    const now = new Date();

    const mappedApiEvents: CalendarEvent[] = apiEvents.map(post => {
      const datePart = post.event_date ? post.event_date.split('T')[0] : toDateKey(new Date(post.created_at || ''));
      const timePart = post.event_date && post.event_date.includes('T') ? post.event_date.split('T')[1].substring(0, 5) : '';
      let category = 'Info';
      if (post.tags) {
        try {
          const tagsArray = post.tags;
          if (Array.isArray(tagsArray)) {
            const found = tagsArray.find(t => ['Treffen', 'Workshop', 'Party', 'Info'].includes(t));
            if (found) category = found;
          }
        } catch (e) {
          console.error('Failed to parse tags for event category', e);
        }
      }

      return {
        id: String(post.id!),
        title: post.title!,
        date: datePart,
        time: timePart,
        location: post.location || '',
        category: category
      };
    });

    return mappedApiEvents
      .filter((event) => {
        const eventTs = toEventTimestamp(event);
        return eventTs >= now.getTime();
      })
      .sort((a, b) => toEventTimestamp(a) - toEventTimestamp(b));
  }, [apiEvents]);

  const upcomingPreview: CalendarEvent[] = upcomingEvents.slice(0, 5);

  return (
    <Sidebar user={user} title="Startseite" maxWidth="xl">
      <Stack spacing={4} sx={{ mt: 1 }}>


        <Box
          component="section"
          sx={{
            position: 'relative',
            height: { xs: 240, md: 400 },
            borderRadius: 6,
            overflow: 'hidden',
            border: '1px solid',
            borderColor: 'divider',
          }}
        >
          {eventsData.length > 0 ? (
            eventsData.map((ev, idx) => (
              <Box
                key={String(ev.id)}
                sx={{
                  position: 'absolute',
                  inset: 0,
                  transition: 'opacity 1s ease-in-out',
                  opacity: carouselIndex === idx ? 1 : 0,
                  zIndex: carouselIndex === idx ? 0 : -1,
                }}
              >
                <Box
                  component="img"
                  src={ev.cover_path ? getSizedImageUrl(`/api/v1/events/${ev.id}/cover`, 1600) : undefined}
                  srcSet={ev.cover_path ? getImageSrcSet(`/api/v1/events/${ev.id}/cover`) : undefined}
                  sizes="100vw"
                  sx={{
                    width: '100%',
                    height: '100%',
                    objectFit: 'cover',
                  }}
                />
                <Box
                  sx={{
                    position: 'absolute',
                    inset: 0,
                    background: `linear-gradient(to right, ${alpha(theme.palette.background.default, 0.95)} 0%, ${alpha(theme.palette.background.default, 0.4)} 50%, transparent 100%), linear-gradient(to top, ${alpha(theme.palette.background.default, 0.8)}, transparent)`,
                  }}
                />
              </Box>
            ))
          ) : (
            <Box
              sx={{
                position: 'absolute',
                inset: 0,
                background: isDark
                  ? `linear-gradient(120deg, ${alpha(theme.palette.background.default, 0.85)}, ${alpha(theme.palette.background.default, 0.2)})`
                  : `linear-gradient(120deg, ${alpha(theme.palette.primary.dark, 0.45)}, ${alpha(theme.palette.primary.dark, 0.08)})`,
              }}
            />
          )}

          <Box
            sx={{
              position: 'relative',
              zIndex: 1,
              height: '100%',
              px: { xs: 4, md: 8 },
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              pointerEvents: 'none',
            }}
          >
            <Stack spacing={1.2} sx={{ maxWidth: 800, pointerEvents: 'auto', pb: { xs: 4.5, md: 0 } }}>
              <Box>
                <Typography variant="overline" sx={{ ...custom.heroOverline, color: 'primary.main', fontWeight: 800, fontSize: { xs: '0.8rem', md: '1.1rem' }, letterSpacing: '0.15em' }}>
                  fsInformatik
                </Typography>
                <Typography variant="h1" sx={{
                  fontFamily: '"Space Grotesk", sans-serif',
                  fontWeight: 800,
                  fontSize: { xs: '2rem', sm: '3rem', md: '4rem' },
                  lineHeight: 1.1,
                  mb: { xs: 0.5, md: 2 },
                  textShadow: isDark ? '0 0 40px rgba(0,0,0,0.5)' : 'none',
                }}>
                  {eventsData[carouselIndex]?.title || 'Dein Hub für das Studium'}
                </Typography>
              </Box>

              <Typography variant="h6" sx={{ color: 'text.secondary', fontWeight: 400, fontSize: { xs: '0.9rem', md: '1.25rem' }, maxWidth: 600, mb: { xs: 0.5, md: 2 } }}>
                Bleib informiert über Events, Klausuren und alles Wichtige rund um dein Studium an der Westfälischen Hochschule.
              </Typography>

              <Stack direction="row" spacing={{ xs: 1, sm: 2 }} sx={{ mt: 1 }}>
                <Button
                  component={RouterLink}
                  to="/discussions"
                  variant="contained"
                  disableElevation
                  size="large"
                  sx={{
                    borderRadius: 3,
                    px: { xs: 2.5, md: 4 },
                    py: 1.2,
                    fontSize: { xs: '0.85rem', md: '1.1rem' },
                    fontWeight: 700,
                    textTransform: 'none',
                  }}
                >
                  Zum Forum
                </Button>
                <Button
                  component={RouterLink}
                  to="/events"
                  variant="outlined"
                  size="large"
                  sx={{
                    borderRadius: 3,
                    px: { xs: 2.5, md: 4 },
                    py: 1.2,
                    fontSize: { xs: '0.85rem', md: '1.1rem' },
                    fontWeight: 700,
                    backdropFilter: 'blur(10px)',
                    textTransform: 'none',
                    whiteSpace: 'nowrap'
                  }}
                >
                  Galerie
                </Button>
              </Stack>
            </Stack>
            {eventsData.length > 1 && (
              <Box sx={{ position: 'absolute', bottom: { xs: 16, md: 32 }, left: { xs: 32, md: 64 }, display: 'flex', gap: 1, pointerEvents: 'auto' }}>
                {eventsData.map((_, i) => (
                  <Box
                    key={i}
                    onClick={() => setCarouselIndex(i)}
                    sx={{
                      width: { xs: 32, md: 40 },
                      height: 4,
                      borderRadius: 2,
                      bgcolor: carouselIndex === i ? 'primary.main' : alpha(theme.palette.text.primary, 0.2),
                      cursor: 'pointer',
                      transition: 'all 0.3s',
                      '&:hover': { bgcolor: alpha(theme.palette.primary.main, 0.6) }
                    }}
                  />
                ))}
              </Box>
            )}
          </Box>
        </Box>

        <Box component="section">
          <Box sx={{ position: 'relative', zIndex: 1 }}>
            <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 2, px: 1 }}>
            </Stack>
            <Box sx={{
              display: 'grid',
              gap: { xs: 3, md: 4 },
              gridTemplateColumns: { xs: '1fr', md: 'repeat(12, minmax(0, 1fr))' },
              alignItems: 'stretch'
            }}>
              <Paper
                elevation={0}
                sx={{
                  gridColumn: { xs: '1 / -1', md: 'span 7' },
                  borderRadius: 4,
                  border: '1px solid',
                  borderColor: 'divider',
                  display: 'flex',
                  flexDirection: 'column',
                  minHeight: 380,
                  overflow: 'hidden'
                }}
              >
                <Box sx={{ p: { xs: 2.2, md: 3 }, flex: 1, display: 'flex', flexDirection: 'column' }}>
                  <Stack spacing={2.5} sx={{ flex: 1 }}>
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                      <Stack direction="row" spacing={2} alignItems="center">
                        <Typography
                          variant="h5"
                          onClick={() => setSidebarIndex(0)}
                          sx={{
                            fontWeight: 800,
                            cursor: 'pointer',
                            color: sidebarIndex === 0 ? 'text.primary' : 'text.secondary',
                            transition: 'color 0.2s',
                            '&:hover': { color: 'primary.main' }
                          }}
                        >
                          Ankündigungen
                        </Typography>
                        <Typography variant="h5" sx={{ color: 'divider', fontWeight: 300 }}>|</Typography>
                        <Typography
                          variant="h5"
                          onClick={() => setSidebarIndex(1)}
                          sx={{
                            fontWeight: 800,
                            cursor: 'pointer',
                            color: sidebarIndex === 1 ? 'text.primary' : 'text.secondary',
                            transition: 'color 0.2s',
                            '&:hover': { color: 'primary.main' }
                          }}
                        >
                          Forum
                        </Typography>
                      </Stack>
                    </Stack>

                    <Box sx={{ flex: 1 }}>
                      {sidebarIndex === 0 ? (
                        <Stack spacing={1.5}>
                          {newsItems.slice(0, 3).map((item) => (
                            <Box
                              key={String(item.id)}
                              component={RouterLink}
                              to={`/d/${item.id}`}
                              sx={{
                                ...custom.newsCardLink,
                                bgcolor: isDark ? alpha(theme.palette.primary.main, 0.08) : alpha(theme.palette.primary.main, 0.04),
                                '&:hover': {
                                  ...(custom.newsCardLink['&:hover'] || {}),
                                  bgcolor: isDark ? alpha(theme.palette.primary.main, 0.12) : alpha(theme.palette.primary.main, 0.06),
                                }
                              }}
                            >
                              <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 0.5 }}>
                                <Avatar
                                  src={getAvatarUrl(item.user_avatar_url)}
                                  sx={{
                                    width: 18,
                                    height: 18,
                                    fontSize: "0.6rem",
                                    bgcolor: theme.palette.primary.main,
                                    fontWeight: "bold"
                                  }}
                                >
                                  {item.user_name ? item.user_name[0].toUpperCase() : "A"}
                                </Avatar>
                                <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                                  von <Typography component={Link} to={`/u/${item.user_id}`} onClick={(e) => e.stopPropagation()} variant="caption" sx={{ color: 'inherit', fontWeight: 600, textDecoration: 'none', '&:hover': { textDecoration: 'underline', color: 'primary.main' } }}>{item.user_name || "Anonym"}</Typography> · {formatNewsDate(item.created_at || "")}
                                </Typography>
                              </Stack>
                              <Typography variant="subtitle1" sx={{ fontWeight: 700, lineHeight: 1.2 }}>{item.title}</Typography>
                              <Typography variant="body2" sx={{ color: 'text.secondary', display: '-webkit-box', WebkitLineClamp: 1, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}>
                                {(item.body || '').substring(0, 100)}
                              </Typography>
                            </Box>
                          ))}
                          {newsItems.length === 0 && <Typography variant="body1" sx={{ color: 'text.secondary' }}>Weitere Ankündigungen folgen in Kürze.</Typography>}
                        </Stack>
                      ) : (
                        <Stack spacing={1.5}>
                          {forumPosts.slice(0, 3).map((post) => (
                            <Box key={post.id} component={RouterLink} to={`/d/${post.id}`} sx={custom.newsCardLink}>
                              <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 0.5 }}>
                                <Avatar
                                  src={getAvatarUrl(post.authorAvatarUrl)}
                                  sx={{
                                    width: 18,
                                    height: 18,
                                    fontSize: "0.6rem",
                                    bgcolor: theme.palette.primary.main,
                                    fontWeight: "bold"
                                  }}
                                >
                                  {post.author ? post.author[0].toUpperCase() : "A"}
                                </Avatar>
                                <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                                  von <Typography component={Link} to={`/u/${post.authorId}`} onClick={(e) => e.stopPropagation()} variant="caption" sx={{ color: 'inherit', fontWeight: 600, textDecoration: 'none', '&:hover': { textDecoration: 'underline', color: 'primary.main' } }}>{post.author}</Typography> · {formatForumDate(post.createdAt)}
                                </Typography>
                              </Stack>
                              <Typography variant="subtitle1" sx={{ fontWeight: 700, lineHeight: 1.2 }}>{post.title}</Typography>
                              <Typography variant="body2" sx={{ color: 'text.secondary', display: '-webkit-box', WebkitLineClamp: 1, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}>
                                {post.body}
                              </Typography>
                            </Box>
                          ))}
                        </Stack>
                      )}
                    </Box>

                    <Button
                      component={RouterLink}
                      to={sidebarIndex === 0 ? "/discussions?type=news" : "/discussions"}
                      variant="contained"
                      disableElevation
                      endIcon={<ArrowForwardRounded fontSize="small" />}
                      sx={{
                        mt: 'auto',
                        fontWeight: 700,
                        textTransform: 'none',
                        borderRadius: 3,
                        py: 1.2,
                        width: 'fit-content'
                      }}
                    >
                      Alle anzeigen
                    </Button>
                  </Stack>
                </Box>
              </Paper>
              <Box sx={{ gridColumn: { xs: '1 / -1', md: 'span 5' }, display: 'flex', justifyContent: 'center' }}>
                <Stack spacing={2} sx={{ width: '100%', maxWidth: 400, pt: { xs: 2.2, md: 3 } }}>
                  <Stack direction="row" justifyContent="center" alignItems="center" sx={{ mb: 2 }}>
                    <Typography variant="h5" sx={{ fontWeight: 800 }}>Agenda</Typography>
                  </Stack>

                  <Timeline
                    position="alternate"
                    sx={{
                      p: 0,
                      m: 0,
                    }}
                  >
                    {upcomingPreview.length ? upcomingPreview.slice(0, 5).map((event) => {
                      const accent = getCategoryColor(event.category);
                      const eventDate = parseDateKey(event.date);

                      const Icon = event.category === 'Treffen' ? GroupsRounded :
                        event.category === 'Workshop' ? TerminalRounded :
                          event.category === 'Party' ? CelebrationRounded :
                            event.category === 'Info' ? InfoRounded : MoreHorizRounded;

                      return (
                        <TimelineItem
                          key={event.id}
                          sx={{
                            cursor: 'pointer',
                            '&:hover .MuiTimelineDot-root': {
                              transform: 'scale(1.15)',
                              boxShadow: `0 0 20px ${alpha(accent, 0.4)}`
                            }
                          }}
                          onClick={() => navigate(`/d/${event.id}`)}
                        >
                          <TimelineOppositeContent
                            sx={{ m: 'auto 0' }}
                            variant="body2"
                            color="text.secondary"
                            fontWeight={600}
                          >
                            {formatShortDate(eventDate)}
                            {event.time && (
                              <Typography variant="caption" sx={{ display: 'block', opacity: 0.7, fontWeight: 400 }}>
                                {event.time} Uhr
                              </Typography>
                            )}
                          </TimelineOppositeContent>
                          <TimelineSeparator>
                            <TimelineConnector sx={{ bgcolor: alpha(theme.palette.divider, 0.5) }} />
                            <TimelineDot
                              sx={{
                                bgcolor: accent,
                                boxShadow: 'none',
                                p: 1.2,
                                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                                border: 'none',
                              }}
                            >
                              <Icon sx={{ fontSize: 20, color: '#fff' }} />
                            </TimelineDot>
                            <TimelineConnector sx={{ bgcolor: alpha(theme.palette.divider, 0.5) }} />
                          </TimelineSeparator>
                          <TimelineContent sx={{ py: '12px', px: 2, m: 'auto 0' }}>
                            <Typography variant="h6" component="span" sx={{
                              fontWeight: 800,
                              fontSize: '0.95rem',
                              lineHeight: 1.2,
                              color: 'text.primary',
                              display: 'block'
                            }}>
                              {event.title}
                            </Typography>
                            {event.location && (
                              <Typography variant="caption" color="text.secondary" sx={{
                                fontSize: '0.8rem',
                                display: '-webkit-box',
                                WebkitLineClamp: 1,
                                WebkitBoxOrient: 'vertical',
                                overflow: 'hidden'
                              }}>
                                {event.location}
                              </Typography>
                            )}
                          </TimelineContent>
                        </TimelineItem>
                      );
                    }) : (
                      <Typography variant="body2" sx={{ color: 'text.secondary', textAlign: 'center', py: 4 }}>
                        Keine kommenden Events.
                      </Typography>
                    )}
                  </Timeline>
                </Stack>
              </Box>
            </Box>
          </Box>
        </Box>

      </Stack>
    </Sidebar >
  );
};

export default NewsFeedPage;
