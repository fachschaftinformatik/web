import React, { useEffect, useState } from 'react';
import {
  Avatar,
  Box,
  Button,
  Chip,
  Grid,
  Paper,
  Stack,
  Typography,
  CircularProgress,
} from '@mui/material';
import { alpha, useTheme } from '@mui/material/styles';
import { Link as RouterLink, useParams } from 'react-router-dom';
import { client } from '@lib/api/client.gen';
import LogoutIcon from '@mui/icons-material/Logout';

import { useAuth } from '@lib/auth';
import { Sidebar } from '@components/layout';
import { getPrograms } from '@lib/api';


const ProfilePage: React.FC = () => {
  const { userId } = useParams<{ userId: string }>();
  const { user: currentUser, logout } = useAuth();
  const theme = useTheme();

  const [profileUser, setProfileUser] = useState<any>(null);
  const [profileLoading, setProfileLoading] = useState(true);
  const [programs, setPrograms] = useState<Record<number, string>>({});
  const [programsLoading, setProgramsLoading] = useState(false);

  const isOwnProfile = currentUser?.id === userId;

  useEffect(() => {
    if (!userId) return;

    const fetchProfile = async () => {
      setProfileLoading(true);
      try {
        const res = await client.request({
          method: 'GET',
          url: `/users/${userId}`
        });
        if (res.data) {
          setProfileUser(res.data);
        }
      } catch (err) {
        console.error("Failed to fetch profile", err);
      } finally {
        setProfileLoading(false);
      }
    };

    fetchProfile();
  }, [userId]);

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
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="50vh">
          <CircularProgress />
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


  const handleLogout = async () => {
    await logout();
  };

  return (
    <Sidebar user={currentUser} title={`Profil von ${profileUser.name}`} maxWidth="lg">
      <Box
        sx={(theme) => ({
          background: theme.palette.mode === 'light'
            ? 'linear-gradient(180deg, rgba(76, 175, 80, 0.08), rgba(255,255,255,0.95))'
            : 'linear-gradient(180deg, rgba(56, 142, 60, 0.15), rgba(11, 25, 18, 0.98))',
          borderRadius: 4,
        })}
      >
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
                sx={{
                  width: 120,
                  height: 120,
                  bgcolor: theme.palette.primary.main,
                  fontSize: 48,
                  fontWeight: 700,
                  boxShadow: `0 8px 24px ${alpha(theme.palette.primary.main, 0.25)}`
                }}
              >
                {profileUser.name ? profileUser.name.charAt(0).toUpperCase() : '?'}
              </Avatar>
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
                  {profileUser.verified === 1 && (
                    <Chip
                      label="Verfiziert"
                      size="small"
                      color="success"
                      sx={{ fontWeight: 700 }}
                    />
                  )}
                </Stack>
                <Typography variant="body1" color="text.secondary" sx={{ mt: 2, maxWidth: 600 }}>
                  Studierende/r im Bereich {programLabel}. Seit {profileUser.created_at ? new Date(profileUser.created_at).toLocaleDateString() : 'Anfang an'} dabei.
                </Typography>
              </Box>
              {isOwnProfile && (
                <Stack direction="row" spacing={2} alignSelf={{ xs: 'stretch', md: 'flex-start' }}>
                  <Button
                    variant="outlined"
                    component={RouterLink}
                    to="/settings"
                    sx={{ borderRadius: 2, fontWeight: 700 }}
                  >
                    Profil bearbeiten
                  </Button>
                  <Button
                    variant="contained"
                    color="error"
                    startIcon={<LogoutIcon />}
                    onClick={handleLogout}
                    sx={{ borderRadius: 2, fontWeight: 700, boxShadow: 'none' }}
                  >
                    Abmelden
                  </Button>
                </Stack>
              )}
            </Stack>
          </Paper>
          <Paper variant="outlined" sx={{ p: 4, borderRadius: 3, border: '1px solid', borderColor: 'divider' }}>
            <Typography variant="h6" fontWeight={700} gutterBottom sx={{ mb: 3 }}>
              Informationen
            </Typography>
            <Grid container spacing={3}>
              <Grid size={{ xs: 12, sm: 6 }}>
                <Typography variant="caption" color="text.secondary" fontWeight={700} sx={{ textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                  Studiengang
                </Typography>
                <Typography variant="body1" fontWeight={600} sx={{ mt: 0.5 }}>
                  {programLabel}
                </Typography>
              </Grid>
              <Grid size={{ xs: 12, sm: 6 }}>
                <Typography variant="caption" color="text.secondary" fontWeight={700} sx={{ textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                  Mitglied seit
                </Typography>
                <Typography variant="body1" fontWeight={600} sx={{ mt: 0.5 }}>
                  {profileUser.created_at ? new Date(profileUser.created_at).toLocaleDateString() : 'Unbekannt'}
                </Typography>
              </Grid>
              {isOwnProfile && (
                <Grid size={{ xs: 12, sm: 6 }}>
                  <Typography variant="caption" color="text.secondary" fontWeight={700} sx={{ textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                    E-Mail (Privat)
                  </Typography>
                  <Typography variant="body1" fontWeight={600} sx={{ mt: 0.5 }}>
                    {profileUser.email}
                  </Typography>
                </Grid>
              )}
              <Grid size={{ xs: 12, sm: 6 }}>
                <Typography variant="caption" color="text.secondary" fontWeight={700} sx={{ textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                  Status
                </Typography>
                <Box sx={{ mt: 0.5 }}>
                  <Chip
                    label={profileUser.active === 1 ? "Aktiv" : "Inaktiv"}
                    size="small"
                    color={profileUser.active === 1 ? "success" : "default"}
                    variant="outlined"
                    sx={{ fontWeight: 600 }}
                  />
                </Box>
              </Grid>
            </Grid>
          </Paper>


        </Stack>
      </Box>
    </Sidebar>
  );
};

export default ProfilePage;
