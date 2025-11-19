import React from 'react';
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
import { alpha } from '@mui/material/styles';
import { Link as RouterLink } from 'react-router-dom';
import LogoutIcon from '@mui/icons-material/Logout';
import SchoolIcon from '@mui/icons-material/School';
import { mockForumPosts } from '@lib/data';

import { useAuth } from '@lib/auth';
import { Sidebar } from '@components/layout';

const campusLabels: Record<number, string> = {
  1: 'Campus Bocholt',
  2: 'Campus Gelsenkirchen',
};

const disciplineLabels: Record<number, string> = {
  1: 'Wirtschaftsinformatik',
  2: 'Anwendungsinformatik',
  3: 'Medieninformatik',
};

const getCampusLabel = (id: number) => campusLabels[id] ?? `Campus #${id}`;
const getDisciplineLabel = (id: number) => disciplineLabels[id] ?? `Studiengang #${id}`;

const MAX_VISIBLE_FORUM_POSTS = 3;

const DashboardPage: React.FC = () => {
  const { user, logout } = useAuth();

  if (!user) {
    return null;
  }

  const visibleForumPosts = mockForumPosts.slice(0, MAX_VISIBLE_FORUM_POSTS);
  const canToggleForumPosts = mockForumPosts.length > MAX_VISIBLE_FORUM_POSTS;

  const handleLogout = async () => {
    await logout();
  };

  return (
    <Sidebar user={user} title="Dashboard">
      <Box
        sx={(theme) => ({
          background: theme.palette.mode === 'light'
            ? 'linear-gradient(180deg, rgba(76, 175, 80, 0.18), rgba(255,255,255,0.9))'
            : 'linear-gradient(180deg, rgba(56, 142, 60, 0.3), rgba(11, 25, 18, 0.95))',
          borderRadius: 3,
          px: { xs: 1, sm: 2 },
          py: 2,
        })}
      >
      <Stack spacing={3}>
        <Paper
          variant="outlined"
          sx={(theme) => ({
            p: { xs: 3, md: 4 },
            border: 'none',
            background: `linear-gradient(135deg, ${theme.palette.primary.main}, ${alpha(theme.palette.primary.dark, 0.9)})`,
            color: theme.palette.common.white,
            position: 'relative',
            overflow: 'hidden',
          })}
        >
          <Box
            sx={{
              position: 'absolute',
              inset: 0,
              background: 'radial-gradient(circle at top right, rgba(255,255,255,0.35), transparent 45%)',
              pointerEvents: 'none',
            }}
          />
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={3} alignItems="center" sx={{ position: 'relative' }}>
            <Avatar sx={{ width: 80, height: 80, bgcolor: 'rgba(255,255,255,0.18)', fontSize: 36 }}>
              {user.name ? user.name.charAt(0).toUpperCase() : '?'}
            </Avatar>
            <Box flex={1}>
              <Typography variant="h4" fontWeight={600}>
                Schön, dass du da bist, {user.name.split(' ')[0] ?? user.name}
              </Typography>
              <Typography variant="body1" sx={{ opacity: 0.9 }}>
                {user.email}
              </Typography>
              <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} mt={2}>
                <Chip label={user.role.toUpperCase()} color="default" sx={{ bgcolor: 'rgba(255,255,255,0.16)' }} />
                <Chip
                  label={user.verified ? 'Verifiziert' : 'Nicht verifiziert'}
                  sx={{ bgcolor: 'rgba(255,255,255,0.16)' }}
                  color={user.verified ? 'success' : 'default'}
                />
              </Stack>
            </Box>
            <Stack spacing={1} alignItems="flex-end">
              <Button variant="contained" color="secondary" startIcon={<LogoutIcon />} onClick={handleLogout}>
                Logout
              </Button>
            </Stack>
          </Stack>
        </Paper>
        <Paper variant="outlined" sx={{ p: 3 }}>
          <Typography variant="h6" gutterBottom>
            Account-Details
          </Typography>
          <Grid container spacing={2}>
            <Grid
              size={{
                xs: 12,
                sm: 6
              }}>
              <Typography variant="caption" color="text.secondary">
                Campus
              </Typography>
              <Typography variant="body1" fontWeight={500}>
                {getCampusLabel(user.campusid)}
              </Typography>
            </Grid>
            <Grid
              size={{
                xs: 12,
                sm: 6
              }}>
              <Typography variant="caption" color="text.secondary">
                Studiengang
              </Typography>
              <Typography variant="body1" fontWeight={500}>
                {getDisciplineLabel(user.disciplineid)}
              </Typography>
            </Grid>
            <Grid
              size={{
                xs: 12,
                sm: 6
              }}>
              <Typography variant="caption" color="text.secondary">
                Erstellt am
              </Typography>
              <Typography variant="body1" fontWeight={500}>
                {new Date(user.created_at).toLocaleString()}
              </Typography>
            </Grid>
            <Grid
              size={{
                xs: 12,
                sm: 6
              }}>
              <Typography variant="caption" color="text.secondary">
                Zuletzt aktualisiert
              </Typography>
              <Typography variant="body1" fontWeight={500}>
                {new Date(user.updated_at).toLocaleString()}
              </Typography>
            </Grid>
          </Grid>
        </Paper>

        <Paper variant="outlined" sx={{ p: 3 }}>
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} alignItems={{ xs: 'flex-start', sm: 'center' }} mb={2}>
            <Box flex={1}>
              <Typography variant="h6">Neueste Forenbeiträge</Typography>
              <Typography variant="body2" color="text.secondary">
                Vorschau auf deine letzten Aktivitäten im Forum.
              </Typography>
            </Box>
            {canToggleForumPosts && (
              <Button
                variant="outlined"
                component={RouterLink}
                to="/forum"
                color="primary"
                sx={(theme) => ({
                  borderColor:
                    theme.palette.primary.main,
                  color:
                    theme.palette.primary.main,
                  fontWeight: 600,
                  '&:hover': {
                    borderColor: theme.palette.primary.main,
                    bgcolor:
                      alpha(theme.palette.primary.main, 0.08),
                    ...theme.applyStyles("dark", {
                      bgcolor: alpha(theme.palette.primary.main, 0.2)
                    })
                  },
                  ...theme.applyStyles("dark", {
                    borderColor: alpha(theme.palette.primary.light, 0.8),
                    color: theme.palette.primary.light
                  })
                })}
              >
                Alle anzeigen
              </Button>
            )}
          </Stack>
          <List disablePadding>
            {visibleForumPosts.map((post, index) => (
              <React.Fragment key={post.id}>
                {index > 0 && (
                  <Divider
                    sx={(theme) => ({
                      my: 1.5,
                      borderColor:
                        alpha(theme.palette.grey[500], 0.3),
                      ...theme.applyStyles("dark", {
                        borderColor: alpha(theme.palette.common.white, 0.08)
                      })
                    })}
                  />
                )}
                <ListItem
                  alignItems="flex-start"
                  sx={(theme) => ({
                    px: { xs: 1.5, sm: 2 },
                    py: 1.75,
                    flexDirection: { xs: 'column', sm: 'row' },
                    gap: { xs: 1, sm: 0 },
                    borderRadius: 2,
                    border: `1px solid ${
                      alpha(theme.palette.grey[400], 0.6)
                    }`,
                    bgcolor: theme.palette.background.paper,
                    boxShadow: `0 6px 18px ${
                      alpha(theme.palette.primary.main, 0.25)
                    }`,
                    transition: theme.transitions.create(['transform', 'box-shadow', 'border-color'], {
                      duration: theme.transitions.duration.short,
                    }),
                    '&:hover': {
                      transform: 'translateY(-4px)',
                      boxShadow: `0 14px 30px ${alpha(theme.palette.primary.main, 0.35)}`,
                      borderColor: theme.palette.primary.main,
                    },
                    ...theme.applyStyles("dark", {
                      border: `1px solid ${alpha(theme.palette.common.white, 0.1)}`,
                      boxShadow: `0 6px 18px ${alpha(theme.palette.primary.dark, 0.45)}`,
                      '&:hover': {
                        boxShadow: `0 14px 30px ${alpha(theme.palette.primary.dark, 0.6)}`,
                        borderColor: theme.palette.primary.light,
                      },
                    }),
                  })}
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
                      {new Date(post.createdAt).toLocaleString()} · {post.replies} Antworten
                    </Typography>
                  </Box>
                </ListItem>
              </React.Fragment>
            ))}
          </List>
        </Paper>

        <Paper
          variant="outlined"
          sx={(theme) => ({
            p: { xs: 3, md: 4 },
            borderStyle: 'dashed',
            borderColor: alpha(theme.palette.primary.main, 0.3),
            bgcolor: alpha(theme.palette.primary.main, 0.04),
          })}
        >
          <Typography variant="subtitle2" color="text.secondary">
            Lass uns wachsen
          </Typography>
          <Typography variant="h6" gutterBottom>
            Teile dein Prüfungswissen
          </Typography>
          <Typography variant="body2" color="text.secondary" mb={2}>
            Lade neue Rekos hoch oder kommentiere vorhandene Skripte. Jede Ergänzung hilft und macht den Wissenspool stärker.
          </Typography>
          <Button
            variant="contained"
            startIcon={<SchoolIcon />}
            component={RouterLink}
            to="/exams"
          >
            Jetzt beitragen
          </Button>
        </Paper>
      </Stack>
    </Box>
    </Sidebar>
  );
};

export default DashboardPage;
