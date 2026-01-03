import React, { useEffect, useState } from 'react';
import {
  Avatar,
  Box,
  Button,
  Chip,
  Divider,
  Grid,
  List,
  ListItem,
  Paper,
  Stack,
  Typography,
} from '@mui/material';
import { Link as RouterLink } from 'react-router-dom';
import LogoutIcon from '@mui/icons-material/Logout';
import SchoolIcon from '@mui/icons-material/School';
import ArrowForwardIcon from '@mui/icons-material/ArrowForward';
import { FORUM_POST_SUMMARIES, ForumPostSummary } from '@lib/data';

import { useAuth } from '@lib/auth';
import { Sidebar } from '@components/layout';
import { getPrograms } from '@lib/api';

const MAX_VISIBLE_FORUM_POSTS = 3;

const DashboardPage: React.FC = () => {
  const { user, logout } = useAuth();
  const [programs, setPrograms] = useState<Record<number, string>>({});
  const [programLoading, setProgramLoading] = useState(false);

  useEffect(() => {
    let isMounted = true;
    const fetchPrograms = async () => {
      setProgramLoading(true);
      try {
        const { data } = await getPrograms();
        if (!isMounted) return;
        if (data) {
          const mapped = data.reduce<Record<number, string>>((acc, program) => {
            if (program.id !== undefined && program.name) {
              acc[program.id] = program.name;
            }
            return acc;
          }, {});
          setPrograms(mapped);
        }
      } catch (error) {
        console.error('Studiengänge konnten nicht geladen werden', error);
        if (isMounted) {
          setPrograms({});
        }
      } finally {
        if (isMounted) {
          setProgramLoading(false);
        }
      }
    };

    fetchPrograms();

    return () => {
      isMounted = false;
    };
  }, []);

  if (!user) {
    return null;
  }

  const programLabel = programLoading
    ? 'Studiengang wird geladen...'
    : (user.programid !== undefined ? programs[user.programid] : undefined)
    ?? `Studiengang #${user.programid ?? 'unbekannt'}`;

  const visibleForumPosts = FORUM_POST_SUMMARIES.slice(0, MAX_VISIBLE_FORUM_POSTS);
  const canToggleForumPosts = FORUM_POST_SUMMARIES.length > MAX_VISIBLE_FORUM_POSTS;

  const handleLogout = async () => {
    await logout();
  };

  return (
    <Sidebar user={user} title="Dashboard">
      <Stack spacing={4}>
        <Box>
          <Typography variant="h4" component="h1" fontWeight={700} gutterBottom>
            Dashboard
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Willkommen zurück, {user.name.split(' ')[0] ?? user.name}! Hier ist deine Übersicht.
          </Typography>
        </Box>

        <Paper variant="outlined" sx={{ p: 3, borderRadius: 3 }}>
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={3} alignItems="center">
            <Avatar
              sx={{
                width: 64,
                height: 64,
                bgcolor: 'primary.main',
                color: 'primary.contrastText',
                fontSize: 24,
                fontWeight: 600
              }}
            >
              {user.name ? user.name.charAt(0).toUpperCase() : '?'}
            </Avatar>
            <Box flex={1}>
              <Typography variant="h6" fontWeight={600} color="text.primary">
                {user.name}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {user.email}
              </Typography>
              <Stack direction="row" spacing={1} mt={1.5}>
                <Chip label={user.role.toUpperCase()} size="small" variant="outlined" />
                <Chip
                  label={user.verified ? 'Verifiziert' : 'Nicht verifiziert'}
                  size="small"
                  color={user.verified ? 'success' : 'default'}
                  variant={user.verified ? 'filled' : 'outlined'}
                />
              </Stack>
            </Box>
            <Button variant="outlined" color="inherit" startIcon={<LogoutIcon />} onClick={handleLogout}>
              Abmelden
            </Button>
          </Stack>
        </Paper>

        <Grid container spacing={3}>
          <Grid size={{ xs: 12, md: 8 }}>
            <Stack spacing={3}>
              <Paper variant="outlined" sx={{ p: 3, borderRadius: 3 }}>
                <Stack direction="row" justifyContent="space-between" alignItems="center" mb={1} px={1}>
                  <Typography variant="h6" color="primary.main" fontWeight={600}>Forendiskussionen</Typography>
                  {canToggleForumPosts && (
                    <Button
                      variant="text"
                      component={RouterLink}
                      to="/forum"
                      size="small"
                      color="primary"
                      endIcon={<ArrowForwardIcon />}
                    >
                      Alle anzeigen
                    </Button>
                  )}
                </Stack>

                <List disablePadding>
                  {visibleForumPosts.map((post: ForumPostSummary, index: number) => (
                    <React.Fragment key={post.id}>
                      {index > 0 && <Divider sx={{ my: 1 }} />}
                      <ListItem
                        disableGutters
                        alignItems="flex-start"
                        component={RouterLink}
                        to="/forum"
                        sx={{
                          textDecoration: 'none',
                          color: 'inherit',
                          borderRadius: 2,
                          px: 1,
                          '&:hover': {
                            bgcolor: 'action.hover',
                          },
                        }}
                      >
                        <Box flex={1}>
                          <Stack direction="row" spacing={1} alignItems="center" mb={0.5}>
                            <Typography variant="subtitle1" fontWeight={600}>
                              {post.title}
                            </Typography>
                            <Chip label={post.category} size="small" variant="outlined" />
                          </Stack>
                          <Typography variant="body2" color="text.secondary" mb={0.5}>
                            {post.excerpt}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {new Date(post.createdAt).toLocaleDateString()} · {post.replies} Antworten
                          </Typography>
                        </Box>
                      </ListItem>
                    </React.Fragment>
                  ))}
                </List>
              </Paper>

              <Paper
                variant="outlined"
                sx={{
                  p: 3,
                  borderRadius: 3,
                  bgcolor: 'background.paper'
                }}
              >
                <Stack spacing={2} alignItems="flex-start">
                  <Typography variant="h6" color="primary.main" fontWeight={600}>
                    Wissen teilen
                  </Typography>
                  <Typography variant="h5" fontWeight={600}>
                    Teile dein Prüfungswissen
                  </Typography>
                  <Typography variant="body1" color="text.secondary">
                    Lade neue Rekos hoch oder kommentiere vorhandene Skripte. Jede Ergänzung hilft und macht den Wissenspool stärker.
                  </Typography>
                  <Button
                    variant="contained"
                    startIcon={<SchoolIcon />}
                    component={RouterLink}
                    to="/exams"
                    size="large"
                  >
                    Jetzt beitragen
                  </Button>
                </Stack>
              </Paper>
            </Stack>
          </Grid>

          <Grid size={{ xs: 12, md: 4 }}>
            <Paper variant="outlined" sx={{ p: 3, borderRadius: 3 }}>
              <Typography variant="h6" color="primary.main" fontWeight={600} gutterBottom px={1}>
                Account-Details
              </Typography>
              <Stack spacing={2.5}>
                <Box>
                  <Typography variant="caption" color="text.secondary" sx={{ textTransform: 'uppercase', fontWeight: 600 }}>
                    Studiengang
                  </Typography>
                  <Typography variant="body1" fontWeight={500}>
                    {programLabel}
                  </Typography>
                </Box>
                <Divider />
                <Box>
                  <Typography variant="caption" color="text.secondary" sx={{ textTransform: 'uppercase', fontWeight: 600 }}>
                    Erstellt am
                  </Typography>
                  <Typography variant="body1" fontWeight={500}>
                    {new Date(user.created_at).toLocaleDateString()}
                  </Typography>
                </Box>
                <Divider />
                <Box>
                  <Typography variant="caption" color="text.secondary" sx={{ textTransform: 'uppercase', fontWeight: 600 }}>
                    E-Mail Adresse
                  </Typography>
                  <Typography variant="body1" fontWeight={500}>
                    {user.email}
                  </Typography>
                </Box>
                <Divider />
                <Box>
                  <Typography variant="caption" color="text.secondary" sx={{ textTransform: 'uppercase', fontWeight: 600 }}>
                    Vollständiger Name
                  </Typography>
                  <Typography variant="body1" fontWeight={500}>
                    {user.name}
                  </Typography>
                </Box>
              </Stack>
            </Paper>
          </Grid>
        </Grid>
      </Stack>
    </Sidebar>
  );
};

export default DashboardPage;
