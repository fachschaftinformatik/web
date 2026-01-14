import React, { useEffect, useState } from 'react';
import {
  Avatar,
  Box,
  Button,
  Chip,
  Paper,
  Stack,
  Typography,
  CircularProgress,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Divider,
  Skeleton,
  Pagination,
} from '@mui/material';
import { alpha, useTheme } from '@mui/material/styles';
import { Link as RouterLink, useParams } from 'react-router-dom';
import HistoryIcon from '@mui/icons-material/History';
import PostAddIcon from '@mui/icons-material/PostAdd';
import CommentIcon from '@mui/icons-material/Comment';
import UploadFileIcon from '@mui/icons-material/UploadFile';
import LockRoundedIcon from '@mui/icons-material/LockRounded';

import { useAuth } from '@lib/auth';
import { Sidebar } from '@components/layout';
import { getPrograms, getUsersByIdActivities, getUsersById } from '@lib/api';
import type { AuthUserResponse as User, DatabaseActivity as Activity } from '@lib/api';


const ProfilePage: React.FC = () => {
  const { userId } = useParams<{ userId: string }>();
  const { user: currentUser } = useAuth();
  const theme = useTheme();

  const [profileUser, setProfileUser] = useState<User | null>(null);
  const [profileLoading, setProfileLoading] = useState(true);
  const [programs, setPrograms] = useState<Record<number, string>>({});
  const [programsLoading, setProgramsLoading] = useState(false);
  const [activities, setActivities] = useState<Activity[]>([]);
  const [activitiesLoading, setActivitiesLoading] = useState(false);

  const isOwnProfile = currentUser?.id === userId;

  const [page, setPage] = useState(1);
  const [pageCount, setPageCount] = useState(1);
  const LIMIT = 5;

  useEffect(() => {
    if (!userId) return;

    const fetchProfile = async () => {
      setProfileLoading(true);
      try {
        const { data } = await getUsersById({ path: { id: userId } });
        if (data) {
          setProfileUser(data as User);
        }
      } catch (err) {
        console.error("Failed to fetch profile", err);
      } finally {
        setProfileLoading(false);
      }
    };

    fetchProfile();
    // Reset activities when userId changes
    setActivities([]);
    setPage(1);
    setPageCount(1);
    // Initial fetch handled by the pagination effect below or dedicated init?
    // Let's use a separate effect for activities that depends on page
  }, [userId]);

  useEffect(() => {
    if (!userId) return;

    const fetchActivities = async () => {
      setActivitiesLoading(true);
      try {
        const offset = (page - 1) * LIMIT;
        const { data, response } = await getUsersByIdActivities({
          path: { id: userId },
          query: { limit: LIMIT, offset }
        });

        if (response.headers.has('X-Total-Count')) {
          const total = parseInt(response.headers.get('X-Total-Count') || '0', 10);
          setPageCount(Math.ceil(total / LIMIT));
        }

        if (data && Array.isArray(data)) {
          setActivities(data as Activity[]);
        }
      } catch (err) {
        console.error("Failed to fetch activities", err);
      } finally {
        setActivitiesLoading(false);
      }
    };

    fetchActivities();
  }, [userId, page]);

  useEffect(() => {
    let isMounted = true;
    const fetchPrograms = async () => {
      setProgramsLoading(true);
      try {
        const { data } = await getPrograms();
        if (!isMounted) return;
        if (data) {
          const mapped = (Array.isArray(data) ? data : []).reduce<Record<number, string>>((acc, p) => {
            if (p.id !== undefined) acc[p.id] = p.name ?? "Unbekannt";
            return acc;
          }, {});
          setPrograms(mapped);
        }
      } catch (error) {
        console.error('Studiengänge konnten nicht geladen werden', error);
      } finally {
        if (isMounted) setProgramsLoading(false);
      }
    };

    fetchPrograms();
    return () => { isMounted = false; };
  }, []);

  if (profileLoading) {
    return (
      <Sidebar user={currentUser} title="Profil">
        <Box sx={{ background: theme.palette.mode === 'light' ? 'linear-gradient(180deg, rgba(76, 175, 80, 0.04), transparent)' : 'transparent', borderRadius: 4 }}>
          <Stack spacing={4}>
            <Paper elevation={0} sx={{ p: { xs: 4, md: 5 }, borderRadius: 4, border: '1px solid', borderColor: alpha(theme.palette.divider, 0.1), backdropFilter: 'blur(20px)' }}>
              <Stack direction={{ xs: 'column', md: 'row' }} spacing={4} alignItems="center">
                <Skeleton variant="circular" width={120} height={120} />
                <Box flex={1} textAlign={{ xs: 'center', md: 'left' }}>
                  <Skeleton variant="text" width="40%" height={60} />
                  <Stack direction="row" spacing={1} justifyContent={{ xs: 'center', md: 'flex-start' }} sx={{ mt: 1 }}>
                    <Skeleton variant="rounded" width={60} height={24} />
                    <Skeleton variant="rounded" width={80} height={24} />
                  </Stack>
                  <Skeleton variant="text" width="60%" sx={{ mt: 2 }} />
                </Box>
              </Stack>
            </Paper>
            <Paper variant="outlined" sx={{ p: 4, borderRadius: 3, border: '1px solid', borderColor: 'divider' }}>
              <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 3 }}>
                <Skeleton variant="circular" width={24} height={24} />
                <Skeleton variant="text" width="200px" height={32} />
              </Stack>
              <Stack spacing={2}>
                {[1, 2, 3].map((i) => (
                  <Box key={i} sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
                    <Skeleton variant="circular" width={32} height={32} />
                    <Box sx={{ flex: 1 }}>
                      <Skeleton variant="text" width="30%" />
                      <Skeleton variant="text" width="20%" />
                    </Box>
                  </Box>
                ))}
              </Stack>
            </Paper>
          </Stack>
        </Box>
      </Sidebar>
    );
  }

  if (!profileUser) {
    return (
      <Sidebar user={currentUser} title="Profil nicht gefunden">
        <Box sx={{ mt: 5, textAlign: 'center' }}>
          <Typography variant="h5">Dieser Nutzer existiert nicht.</Typography>
          <Button component={RouterLink} to="/" sx={{ mt: 2 }}>Zurück zur Startseite</Button>
        </Box>
      </Sidebar>
    );
  }

  const programLabel = programsLoading
    ? 'Lädt...'
    : (profileUser.programid !== undefined && programs[profileUser.programid])
    ?? `Studiengang #${profileUser.programid ?? 'unbekannt'}`;



  return (
    <Sidebar user={currentUser} title={`Nutzer`} maxWidth="lg">
      <Box>
        <Stack spacing={4}>
          <Paper
            elevation={0}
            sx={(theme) => ({
              p: { xs: 4, md: 5 },
              borderRadius: 4,
              border: '1px solid',
              borderColor: alpha(theme.palette.divider, 0.1),
              background: theme.palette.mode === 'dark'
                ? alpha(theme.palette.background.paper, 0.6)
                : alpha(theme.palette.background.paper, 0.8),
              backdropFilter: 'blur(20px)',
              position: 'relative',
              overflow: 'hidden',
            })}
          >
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={4} alignItems="center">
              <Avatar
                src={profileUser.avatar_url || undefined}
                sx={{
                  width: 120,
                  height: 120,
                  bgcolor: theme.palette.primary.main,
                  fontSize: 48,
                  fontWeight: 700,
                  boxShadow: `0 8px 24px ${alpha(theme.palette.primary.main, 0.25)}`
                }}
              />
              <Box flex={1} textAlign={{ xs: 'center', md: 'left' }}>
                <Typography variant="h3" fontWeight={800} letterSpacing="-0.02em" gutterBottom>
                  {profileUser.name}
                </Typography>
                <Stack direction="row" spacing={1} justifyContent={{ xs: 'center', md: 'flex-start' }} alignItems="center">
                  <Chip
                    label={(profileUser.role ?? 'USER').toUpperCase()}
                    size="small"
                    sx={{
                      fontWeight: 700,
                      bgcolor: alpha(theme.palette.primary.main, 0.1),
                      color: 'primary.main',
                      border: '1px solid',
                      borderColor: alpha(theme.palette.primary.main, 0.2)
                    }}
                  />
                  {!(profileUser.private === 1 && !isOwnProfile) && (
                    <Chip
                      label={profileUser.active === 1 ? "Aktiv" : "Inaktiv"}
                      size="small"
                      color={profileUser.active === 1 ? "success" : "default"}
                      variant="outlined"
                      sx={{ fontWeight: 600 }}
                    />
                  )}
                </Stack>
                <Typography variant="body1" color="text.secondary" sx={{ mt: 2, maxWidth: 600 }}>
                  {!(profileUser.private === 1 && !isOwnProfile) && `Studierende/r im Bereich ${programLabel}. `}
                  Seit {profileUser.created_at ? new Date(profileUser.created_at).toLocaleDateString() : 'Anfang an'} dabei.
                </Typography>
              </Box>
              {isOwnProfile && (
                <Button
                  variant="outlined"
                  component={RouterLink}
                  to="/settings"
                  sx={{ borderRadius: 2, fontWeight: 700, alignSelf: { xs: 'stretch', md: 'flex-start' } }}
                >
                  Profil bearbeiten
                </Button>
              )}
            </Stack>
          </Paper>

          <Paper variant="outlined" sx={{ p: 4, borderRadius: 3, border: '1px solid', borderColor: 'divider' }}>
            <Stack direction="row" spacing={2} alignItems="center" sx={{ mb: 3 }}>
              <HistoryIcon color="primary" />
              <Typography variant="h6" fontWeight={700}>
                Kürzliche Aktivitäten
              </Typography>
            </Stack>

            {profileUser.private === 1 && !isOwnProfile ? (
              <Box sx={{ py: 8, textAlign: 'center' }}>
                <LockRoundedIcon sx={{ fontSize: 48, color: 'text.disabled', mb: 2, opacity: 0.5 }} />
                <Typography variant="h6" fontWeight={700} color="text.secondary">
                  Dieses Profil ist privat
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ opacity: 0.7 }}>
                  Die Aktivitäten dieses Nutzers sind nur für ihn selbst sichtbar.
                </Typography>
              </Box>
            ) : activitiesLoading ? (
              <Box display="flex" justifyContent="center" p={4}>
                <CircularProgress size={24} />
              </Box>
            ) : activities.length > 0 ? (
              <List sx={{ width: '100%', bgcolor: 'transparent' }}>
                {activities.map((activity, index) => {
                  let linkTarget = "#";
                  let primaryText = "Unbekannte Aktivität";
                  const secondaryText = activity.created_at ? new Date(activity.created_at).toLocaleString() : '';

                  if (activity.type === 'POST_CREATED') {
                    linkTarget = `/forum/${activity.target_id}`;
                    primaryText = activity.target_name || "Neuer Beitrag";
                  } else if (activity.type === 'COMMENT_ADDED') {
                    linkTarget = `/forum/${activity.target_id}`;
                    primaryText = activity.target_name ? `Kommentar zu "${activity.target_name}"` : "Neuer Kommentar";
                  } else if (activity.type === 'EXAM_UPLOADED') {
                    linkTarget = `/exams/detail?examId=${activity.target_id}`;
                    primaryText = activity.target_name ? `Klausur für ${activity.target_name}` : "Neue Altklausur";
                  } else if (activity.type === 'MEDIA_UPLOADED') {
                    linkTarget = `/media?mediaId=${activity.target_id}`;
                    primaryText = activity.target_name || "Neues Bild";
                  }

                  const isDark = theme.palette.mode === 'dark';

                  // Helper for consistent coloring across light/dark
                  const getActivityColor = (type: string, attr: 'bg' | 'fg') => {
                    const isPost = type === 'POST_CREATED';
                    const isExam = type === 'EXAM_UPLOADED';
                    // Default media/other to info

                    if (attr === 'bg') {
                      if (isPost) return alpha(theme.palette.secondary.main, isDark ? 0.2 : 0.1);
                      if (isExam) return alpha(theme.palette.success.main, isDark ? 0.2 : 0.1);
                      return alpha(theme.palette.info.main, isDark ? 0.2 : 0.1);
                    } else {
                      // Foreground
                      if (isPost) return isDark ? theme.palette.secondary.light : theme.palette.secondary.main;
                      if (isExam) return isDark ? theme.palette.success.light : theme.palette.success.main;
                      return isDark ? theme.palette.info.light : theme.palette.info.main;
                    }
                  };

                  return (
                    <React.Fragment key={activity.id}>
                      <ListItem
                        alignItems="center"
                        component={RouterLink}
                        to={linkTarget}
                        sx={{
                          px: 1,
                          py: 1.5,
                          borderRadius: 2,
                          color: 'inherit',
                          textDecoration: 'none',
                          transition: 'background-color 0.2s',
                          '&:hover': {
                            bgcolor: alpha(theme.palette.primary.main, 0.05),
                          }
                        }}
                      >
                        <ListItemIcon sx={{ minWidth: 48, mr: 2, display: 'flex', justifyContent: 'center' }}>
                          <Avatar
                            sx={{
                              width: 40,
                              height: 40,
                              bgcolor: getActivityColor(activity.type, 'bg'),
                              color: getActivityColor(activity.type, 'fg'),
                            }}
                          >
                            {activity.type === 'POST_CREATED' && <PostAddIcon fontSize="small" />}
                            {activity.type === 'COMMENT_ADDED' && <CommentIcon fontSize="small" />}
                            {activity.type === 'EXAM_UPLOADED' && <UploadFileIcon fontSize="small" />}
                            {activity.type === 'MEDIA_UPLOADED' && <UploadFileIcon fontSize="small" />}
                          </Avatar>
                        </ListItemIcon>
                        <ListItemText
                          primary={
                            <Typography variant="body1" fontWeight={600}>
                              {primaryText}
                            </Typography>
                          }
                          secondary={
                            <Typography variant="caption" color="text.secondary">
                              {secondaryText}
                            </Typography>
                          }
                        />
                      </ListItem>
                      {index < activities.length - 1 && <Divider component="li" variant="inset" />}
                    </React.Fragment>
                  );
                })}
              </List>
            ) : (
              <Typography variant="body2" color="text.secondary" textAlign="center" py={4}>
                Keine Aktivitäten gefunden.
              </Typography>
            )}

            {pageCount > 1 && (
              <Stack alignItems="center" mt={3}>
                <Pagination
                  count={pageCount}
                  page={page}
                  onChange={(_, p) => setPage(p)}
                  shape="rounded"
                  variant="outlined"
                  disabled={activitiesLoading}
                />
              </Stack>
            )}
          </Paper>
        </Stack>
      </Box>
    </Sidebar>
  );
};

export default ProfilePage;
