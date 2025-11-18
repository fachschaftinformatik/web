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
import { useAuth } from '../AuthContext';
import { mockForumPosts } from '../Forum/forumData';
import SidebarLayout from '../layouts/SidebarLayout';

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
    <SidebarLayout user={user} title="Dashboard">
      <Stack spacing={3}>
        <Paper variant="outlined" sx={{ p: 3 }}>
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={3} alignItems="center">
            <Avatar sx={{ width: 72, height: 72, bgcolor: 'primary.main', fontSize: 32 }}>
              {user.name ? user.name.charAt(0).toUpperCase() : '?'}
            </Avatar>
            <Box flex={1}>
              <Typography variant="h4" fontWeight={600}>
                Willkommen, {user.name}
              </Typography>
              <Typography variant="body1" color="text.secondary">
                {user.email}
              </Typography>
              <Stack direction="row" spacing={1} mt={2}>
                <Chip label={user.role.toUpperCase()} color="primary" variant="outlined" />
                <Chip
                  label={user.active ? 'Aktiv' : 'Inaktiv'}
                  color={user.active ? 'success' : 'default'}
                  variant="outlined"
                />
                <Chip
                  label={user.verified ? 'Verifiziert' : 'Nicht verifiziert'}
                  color={user.verified ? 'success' : 'default'}
                  variant="outlined"
                />
              </Stack>
            </Box>
            <Button variant="contained" color="secondary" startIcon={<LogoutIcon />} onClick={handleLogout}>
              Logout
            </Button>
          </Stack>
        </Paper>

        <Paper variant="outlined" sx={{ p: 3 }}>
          <Typography variant="h6" gutterBottom>
            Account-Details
          </Typography>
          <Grid container spacing={2}>
            <Grid item xs={12} sm={6}>
              <Typography variant="caption" color="text.secondary">
                Campus
              </Typography>
              <Typography variant="body1" fontWeight={500}>
                {getCampusLabel(user.campusid)}
              </Typography>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Typography variant="caption" color="text.secondary">
                Studiengang
              </Typography>
              <Typography variant="body1" fontWeight={500}>
                {getDisciplineLabel(user.disciplineid)}
              </Typography>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Typography variant="caption" color="text.secondary">
                Erstellt am
              </Typography>
              <Typography variant="body1" fontWeight={500}>
                {new Date(user.created_at).toLocaleString()}
              </Typography>
            </Grid>
            <Grid item xs={12} sm={6}>
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
                    ...theme.applyStyles("dark", {
                      border: `1px solid ${alpha(theme.palette.common.white, 0.1)}`,
                      boxShadow: `0 6px 18px ${alpha(theme.palette.primary.dark, 0.45)}`
                    })
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
          <Typography variant="caption" color="text.secondary" display="block" mt={2}>
            Dieser Bereich kann später mit den echten Profil-Beiträgen aus dem Forum verbunden werden.
          </Typography>
        </Paper>

        <Paper variant="outlined" sx={{ p: 3 }}>
          <Typography variant="h6" gutterBottom>
            Kommende Funktionen
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Hier können später News, Events oder Forenbeiträge eingebunden werden. Aktuell dient der Bereich als Platzhalter
            für weitere Module.
          </Typography>
        </Paper>
      </Stack>
    </SidebarLayout>
  );
};

export default DashboardPage;
