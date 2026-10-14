'use client';

import React, { useEffect, useMemo, useState } from 'react';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Skeleton from '@mui/material/Skeleton';
import { alpha, useTheme } from '@mui/material/styles';
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
import { useRouter } from 'next/navigation';
import Link from 'next/link';

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
  const router = useRouter();
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
  const [isLoading, setIsLoading] = useState(true);
  const [loadedImages, setLoadedImages] = useState<Record<string, boolean>>({});
  const [carouselIndex, setCarouselIndex] = useState(0);

  const handleImageLoad = (id: string) => {
    setLoadedImages((prev) => ({ ...prev, [id]: true }));
  };
  const [sidebarIndex, setSidebarIndex] = useState(0);
  const [newsItems, setNewsItems] = useState<NewsItem[]>([]);
  const [forumPosts, setForumPosts] = useState<ForumPostSummary[]>([]);
  const [apiEvents, setApiEvents] = useState<ApiPost[]>([]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [news, forum, apiEvs, events] = await Promise.all([
          getDiscussions({ query: { type: 'news', limit: 10 } }),
          getDiscussions({ query: { type: 'discussion', limit: 5 } }),
          getDiscussions({ query: { type: 'event', limit: 20 } }),
          getEvents()
        ]);

        if (news.data) setNewsItems(news.data);
        if (forum.data) setForumPosts(loadForumPosts(forum.data));
        if (apiEvs.data) setApiEvents(apiEvs.data);
        if (events.data) setEventsData(events.data);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
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
      <Stack spacing={3} sx={{ mt: 0 }}>


        <Box
          component="section"
          sx={{
            position: 'relative',
            minHeight: { xs: 320, md: 360 },
            display: 'flex',
            borderRadius: 6,
            overflow: 'hidden',
            border: '1px solid',
            borderColor: 'divider',
          }}
        >
          {isLoading ? (
            <Skeleton
              variant="rectangular"
              sx={{
                position: 'absolute',
                inset: 0,
                borderRadius: 6,
                bgcolor: alpha(theme.palette.action.hover, 0.1),
              }}
            />
          ) : eventsData.length > 0 ? (
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
                  onLoad={() => handleImageLoad(String(ev.id))}
                  sx={{
                    width: '100%',
                    height: '100%',
                    objectFit: 'cover',
                    transition: 'opacity 0.6s ease-in-out',
                    opacity: loadedImages[String(ev.id)] ? 1 : 0,
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
              flex: 1,
              minWidth: 0,
              px: { xs: 4, md: 8 },
              py: { xs: 3, md: 4 },
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              pointerEvents: 'none',
            }}
          >
            <Stack spacing={1.2} sx={{ maxWidth: 800, pointerEvents: 'auto' }}>
              <Box>
                <Typography variant="overline" sx={{ ...custom.heroOverline, color: 'primary.main', fontWeight: 800, fontSize: { xs: '0.8rem', md: '1.1rem' }, letterSpacing: '0.15em' }}>
                  fsInformatik
                </Typography>
                <Typography variant="h1" sx={{
                  fontFamily: 'var(--font-space-grotesk), sans-serif',
                  fontWeight: 800,
                  fontSize: { xs: '2rem', sm: '3rem', md: '4rem' },
                  lineHeight: 1.1,
                  minHeight: '2.2em',
                  mb: { xs: 0.5, md: 2 },
                  textShadow: isDark ? '0 0 40px rgba(0,0,0,0.5)' : 'none',
                }}>
                  {isLoading ? <Skeleton animation={false} width="60%" /> : eventsData[carouselIndex]?.title || 'Dein Hub für das Studium'}
                </Typography>
              </Box>

              <Typography variant="h6" sx={{ color: 'text.secondary', fontWeight: 400, fontSize: { xs: '0.9rem', md: '1.25rem' }, maxWidth: 600, mb: { xs: 0.5, md: 2 } }}>
                Bleib informiert über Events, Klausuren und alles Wichtige rund um dein Studium an der Westfälischen Hochschule.
              </Typography>

              <Stack direction="row" spacing={{ xs: 1, sm: 2 }} sx={{ mt: 1 }}>
                <Button
                  component={Link}
                  href="/discussions"
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
                  component={Link}
                  href="/events"
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
              <Box component="nav" aria-label="Veranstaltungen" sx={{ mt: 2, width: '100%', maxWidth: 600, display: 'flex', flexWrap: 'wrap', gap: 1, pointerEvents: 'auto' }}>
                {eventsData.map((event, i) => (
                  <Box
                    component="button"
                    type="button"
                    key={String(event.id)}
                    aria-label={`Veranstaltung anzeigen: ${event.title}`}
                    aria-current={carouselIndex === i ? 'true' : undefined}
                    onClick={() => setCarouselIndex(i)}
                    sx={{
                      flexShrink: 0,
                      width: { xs: 32, md: 40 },
                      height: 24,
                      border: 0,
                      p: 0,
                      borderRadius: 2,
                      bgcolor: 'transparent',
                      cursor: 'pointer',
                      '&::after': {
                        content: '""',
                        display: 'block',
                        height: 4,
                        borderRadius: 2,
                        bgcolor: carouselIndex === i ? 'primary.main' : alpha(theme.palette.text.primary, 0.2),
                        transition: 'background-color 0.3s',
                      },
                      '&:hover::after': { bgcolor: alpha(theme.palette.primary.main, 0.6) },
                      '&:focus-visible': { outline: '2px solid', outlineColor: 'primary.main', outlineOffset: 2 },
                    }}
                  />
                ))}
              </Box>
            )}
          </Box>
        </Box>

        <Box component="section">
          <Box sx={{ position: 'relative', zIndex: 1 }}>
            <Stack
              direction="row"
              sx={{
                alignItems: "center",
                justifyContent: "space-between",
                mb: 2,
                px: 1
              }}>
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
                  minHeight: 320,
                  overflow: 'hidden'
                }}
              >
                <Box sx={{ p: { xs: 2, md: 2.5 }, flex: 1, display: 'flex', flexDirection: 'column' }}>
                  <Stack spacing={2} sx={{ flex: 1 }}>
                    <Stack
                      direction="row"
                      sx={{
                        justifyContent: "space-between",
                        alignItems: "center"
                      }}>
                      <Stack direction="row" spacing={2} sx={{
                        alignItems: "center"
                      }}>
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
                      {isLoading ? (
                        <Stack spacing={1.5}>
                          {[1, 2, 3].map((i) => (
                            <Box key={i} sx={{ ...custom.newsCardLink, cursor: 'default' }}>
                              <Stack
                                direction="row"
                                spacing={1}
                                sx={{
                                  alignItems: "center",
                                  mb: 0.5
                                }}>
                                <Skeleton variant="circular" width={18} height={18} />
                                <Skeleton variant="text" width="40%" height={14} />
                              </Stack>
                              <Skeleton variant="text" width="70%" height={24} />
                              <Skeleton variant="text" width="90%" height={20} />
                            </Box>
                          ))}
                        </Stack>
                      ) : sidebarIndex === 0 ? (
                        <Stack spacing={1.5}>
                          {newsItems.slice(0, 3).map((item) => (
                            <Box
                              key={String(item.id)}
                              sx={{
                                ...custom.newsCardLink,
                                position: 'relative',
                                bgcolor: isDark ? alpha(theme.palette.primary.main, 0.08) : alpha(theme.palette.primary.main, 0.04),
                                '&:hover': {
                                  ...(custom.newsCardLink['&:hover'] || {}),
                                  bgcolor: isDark ? alpha(theme.palette.primary.main, 0.12) : alpha(theme.palette.primary.main, 0.06),
                                }
                              }}
                            >
                              <Stack
                                direction="row"
                                spacing={1}
                                sx={{
                                  alignItems: "center",
                                  mb: 0.5
                                }}>
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
                                  von <Typography component={Link} href={`/u/${item.user_id}`} variant="caption" sx={{ position: 'relative', zIndex: 1, color: 'inherit', fontWeight: 600, textDecoration: 'none', '&:hover': { textDecoration: 'underline', color: 'primary.main' } }}>{item.user_name || "Anonym"}</Typography> · {formatNewsDate(item.created_at || "")}
                                </Typography>
                              </Stack>
                              <Typography
                                component={Link}
                                href={`/d/${item.id}`}
                                variant="subtitle1"
                                sx={{
                                  color: 'inherit', textDecoration: 'none', fontWeight: 700, lineHeight: 1.2,
                                  '&::after': { content: '""', position: 'absolute', inset: 0, borderRadius: 3 },
                                  '&:focus-visible::after': { outline: '2px solid', outlineColor: 'primary.main', outlineOffset: 2 },
                                }}
                              >{item.title}</Typography>
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
                            <Box key={post.id} sx={{ ...custom.newsCardLink, position: 'relative' }}>
                              <Stack
                                direction="row"
                                spacing={1}
                                sx={{
                                  alignItems: "center",
                                  mb: 0.5
                                }}>
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
                                  von <Typography component={Link} href={`/u/${post.authorId}`} variant="caption" sx={{ position: 'relative', zIndex: 1, color: 'inherit', fontWeight: 600, textDecoration: 'none', '&:hover': { textDecoration: 'underline', color: 'primary.main' } }}>{post.author}</Typography> · {formatForumDate(post.createdAt)}
                                </Typography>
                              </Stack>
                              <Typography
                                component={Link}
                                href={`/d/${post.id}`}
                                variant="subtitle1"
                                sx={{
                                  color: 'inherit', textDecoration: 'none', fontWeight: 700, lineHeight: 1.2,
                                  '&::after': { content: '""', position: 'absolute', inset: 0, borderRadius: 3 },
                                  '&:focus-visible::after': { outline: '2px solid', outlineColor: 'primary.main', outlineOffset: 2 },
                                }}
                              >{post.title}</Typography>
                              <Typography variant="body2" sx={{ color: 'text.secondary', display: '-webkit-box', WebkitLineClamp: 1, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}>
                                {post.body}
                              </Typography>
                            </Box>
                          ))}
                        </Stack>
                      )}
                    </Box>

                    <Button
                      component={Link}
                      href={sidebarIndex === 0 ? "/discussions?type=news" : "/discussions"}
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
                <Stack spacing={1.5} sx={{ width: '100%', maxWidth: 400, pt: { xs: 2, md: 2.5 } }}>
                  <Stack
                    direction="row"
                    sx={{
                      justifyContent: "center",
                      alignItems: "center",
                      mb: 1
                    }}>
                    <Typography variant="h5" sx={{ fontWeight: 800 }}>Agenda</Typography>
                  </Stack>

                  <Timeline
                    position="alternate"
                    sx={{
                      p: 0,
                      m: 0,
                    }}
                  >
                    {isLoading ? (
                      [1, 2, 3, 4].map((i) => (
                        <TimelineItem key={i}>
                          <TimelineOppositeContent sx={{ m: 'auto 0' }}>
                            <Skeleton variant="text" width={40} />
                          </TimelineOppositeContent>
                          <TimelineSeparator>
                            <TimelineConnector />
                            <TimelineDot />
                            <TimelineConnector />
                          </TimelineSeparator>
                          <TimelineContent sx={{ py: '8px', px: 2 }}>
                            <Skeleton variant="text" width="80%" />
                            <Skeleton variant="text" width="40%" />
                          </TimelineContent>
                        </TimelineItem>
                      ))
                    ) : upcomingPreview.length ? upcomingPreview.slice(0, 5).map((event) => {
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
                            minHeight: 64,
                            cursor: 'pointer',
                            '&:hover .MuiTimelineDot-root': {
                              transform: 'scale(1.15)',
                              boxShadow: `0 0 20px ${alpha(accent, 0.4)}`
                            }
                          }}
                          onClick={() => router.push(`/d/${event.id}`)}
                        >
                          <TimelineOppositeContent
                            variant="body2"
                            sx={{
                              color: "text.secondary",
                              fontWeight: 600,
                              m: 'auto 0'
                            }}>
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
                                p: 1,
                                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                                border: 'none',
                              }}
                            >
                              <Icon sx={{ fontSize: 18, color: '#fff' }} />
                            </TimelineDot>
                            <TimelineConnector sx={{ bgcolor: alpha(theme.palette.divider, 0.5) }} />
                          </TimelineSeparator>
                          <TimelineContent sx={{ py: '8px', px: 2, m: 'auto 0' }}>
                            <Typography variant="h6" component="span" sx={{
                              fontWeight: 800,
                              fontSize: '0.875rem',
                              lineHeight: 1.2,
                              color: 'text.primary',
                              display: 'block'
                            }}>
                              {event.title}
                            </Typography>
                            {event.location && (
                              <Typography
                                variant="caption"
                                sx={{
                                  color: "text.secondary",
                                  fontSize: '0.75rem',
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
