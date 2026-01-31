import React, { useEffect, useState } from 'react';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import CircularProgress from '@mui/material/CircularProgress';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Divider from '@mui/material/Divider';
import Skeleton from '@mui/material/Skeleton';
import Pagination from '@mui/material/Pagination';
import { alpha, useTheme } from '@mui/material/styles';
import HistoryIcon from '@mui/icons-material/History';
import PostAddIcon from '@mui/icons-material/PostAdd';
import QuestionAnswerRounded from '@mui/icons-material/QuestionAnswerRounded';
import LibraryBooksRounded from '@mui/icons-material/LibraryBooksRounded';
import CollectionsIcon from '@mui/icons-material/Collections';
import LockRoundedIcon from '@mui/icons-material/LockRounded';
import ArrowBackRounded from '@mui/icons-material/ArrowBackRounded';

import { useAuth } from '@lib/auth';
import { Sidebar } from '@components/layout';
import { getPrograms, getUsersByUserIdActivities, getUsersByUserId } from '@lib/api';
import type { DtoPublicUserResponse as User, DtoActivityResponse as Activity } from '@lib/api';
import { getAvatarUrl } from '@lib/images';


const ProfilePage: React.FC = () => {
  const { userId } = useParams<{ userId: string }>();
  const { user: currentUser } = useAuth();
  const navigate = useNavigate();
  const theme = useTheme();

  const [profileUser, setProfileUser] = useState<User | null>(null);
  const [profileLoading, setProfileLoading] = useState(true);
  const [programs, setPrograms] = useState<Record<string, string>>({});
  const [programsLoading, setProgramsLoading] = useState(false);
  const [activities, setActivities] = useState<Activity[]>([]);
  const [activitiesLoading, setActivitiesLoading] = useState(false);

  const isOwnProfile = String(currentUser?.id) === String(userId);

  const [page, setPage] = useState(1);
  const [pageCount, setPageCount] = useState(1);
  const LIMIT = 5;

  useEffect(() => {
    if (!userId) return;

    const fetchProfile = async () => {
      setProfileLoading(true);
      try {
        const { data } = await getUsersByUserId({ path: { userId } });
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
    setActivities([]);
    setPage(1);
    setPageCount(1);
  }, [userId]);

  useEffect(() => {
    if (!userId) return;

    const fetchActivities = async () => {
      if (profileUser?.private === 1 && !isOwnProfile) {
        setActivities([]);
        setPageCount(0);
        return;
      }

      setActivitiesLoading(true);
      try {
        const offset = (page - 1) * LIMIT;
        const { data, response } = await getUsersByUserIdActivities({
          path: { userId },
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
  }, [userId, page, profileUser, isOwnProfile]);

  useEffect(() => {
    let isMounted = true;
    const fetchPrograms = async () => {
      setProgramsLoading(true);
      try {
        const { data } = await getPrograms();
        if (!isMounted) return;
        if (data) {
          const mapped = (Array.isArray(data) ? data : []).reduce<Record<string, string>>((acc, p) => {
            if (p.id !== undefined) acc[String(p.id)] = p.name ?? "Unbekannt";
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
          <Button
            startIcon={<ArrowBackRounded />}
            onClick={() => navigate('/')}
            sx={{
              mt: 2,
              color: 'text.secondary',
              textTransform: 'none',
              '&:hover': { color: 'primary.main', bgcolor: 'transparent', textDecoration: 'underline' }
            }}
          >
            Zurück zur Startseite
          </Button>
        </Box>
      </Sidebar>
    );
  }

  const programLabel = programsLoading
    ? 'Lädt...'
    : (profileUser.program_id !== undefined && programs[String(profileUser.program_id)])
    ?? `Studiengang #${profileUser.program_id ?? 'unbekannt'}`;



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
                key={profileUser.avatar_url || String(profileUser.id)}
                src={getAvatarUrl(profileUser.avatar_url)}
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
                    linkTarget = `/d/${activity.target_id}`;
                    primaryText = activity.target_name || "Neuer Beitrag";
                  } else if (activity.type === 'COMMENT_ADDED') {
                    linkTarget = `/d/${activity.target_id}`;
                    primaryText = activity.target_name ? `Kommentar zu "${activity.target_name}"` : "Neuer Kommentar";
                  } else if (activity.type === 'EXAM_UPLOADED') {
                    linkTarget = `/archive/${activity.target_id}`;
                    primaryText = activity.target_name ? `Klausur für ${activity.target_name}` : "Neue Altklausur";
                    } else if (activity.type === 'MEDIA_UPLOADED') {
                      linkTarget = `/events/${activity.target_id}`;
                      primaryText = activity.target_name || "Neues Bild";
                    } else if (activity.type === 'EVENT_CREATED') {
                      linkTarget = `/events/${activity.target_id}`;
                      primaryText = activity.target_name || "Neues Event";
                    }


                  const getActivityColor = (type: string, attr: 'bg' | 'fg') => {
                    const isDark = theme.palette.mode === 'dark';
                    let paletteColor = theme.palette.info;

                    if (type === 'POST_CREATED') paletteColor = theme.palette.error;
                    if (type === 'COMMENT_ADDED') paletteColor = theme.palette.success;
                    if (type === 'EXAM_UPLOADED') paletteColor = theme.palette.info;
                    if (type === 'MEDIA_UPLOADED') paletteColor = theme.palette.warning;
                    if (type === 'EVENT_CREATED') paletteColor = theme.palette.warning;

                    if (attr === 'bg') {
                      return alpha(paletteColor.main, isDark ? 0.2 : 0.12);
                    } else {
                      return paletteColor.main;
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
                              bgcolor: getActivityColor(activity.type || '', 'bg'),
                              color: getActivityColor(activity.type || '', 'fg'),
                            }}
                          >
                            {activity.type === 'POST_CREATED' && <PostAddIcon fontSize="small" />}
                            {activity.type === 'COMMENT_ADDED' && <QuestionAnswerRounded fontSize="small" />}
                            {activity.type === 'EXAM_UPLOADED' && <LibraryBooksRounded fontSize="small" />}
                            {(activity.type === 'MEDIA_UPLOADED' || activity.type === 'EVENT_CREATED') && <CollectionsIcon fontSize="small" />}
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
