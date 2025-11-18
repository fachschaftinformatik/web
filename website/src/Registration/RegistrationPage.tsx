import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Avatar from '@mui/material/Avatar';
import Button from '@mui/material/Button';
import CssBaseline from '@mui/material/CssBaseline';
import TextField from '@mui/material/TextField';
import Link from '@mui/material/Link';
import Box from '@mui/material/Box';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import Typography from '@mui/material/Typography';
import Container from '@mui/material/Container';
import { Alert, MenuItem, Stack } from '@mui/material';
import { Link as RouterLink } from 'react-router-dom';

import * as api from '../api';
import { useAuth } from '../AuthContext';

// TODO
const UNI_EMAIL_DOMAIN = 'studmail.w-hs.de';

const campusOptions = [
  { value: 1, label: 'Campus Bocholt' },
  { value: 2, label: 'Campus Gelsenkirchen' },
];

const disciplineOptions = [
  { value: 1, label: 'Wirtschaftsinformatik' },
  { value: 2, label: 'Anwendungsinformatik' },
  { value: 3, label: 'Medieninformatik' },
];

export default function RegistrationPage() {
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { login } = useAuth();

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    setError('');

    const data = new FormData(event.currentTarget);
    const email = data.get('email') as string;
    const name = data.get('name') as string;
    const password = data.get('password') as string;
    const confirmPassword = data.get('confirmPassword') as string;
    const campusidValue = data.get('campusid') as string;
    const disciplineidValue = data.get('disciplineid') as string;

    const campusid = Number.parseInt(campusidValue, 10);
    const disciplineid = Number.parseInt(disciplineidValue, 10);

    if (!email || !password || !name || !confirmPassword || !campusidValue || !disciplineidValue) {
       setError('Bitte alle mit "*" markierten Felder ausfüllen.');
       setLoading(false);
       return;
    }

    if (password !== confirmPassword) {
      setError('Die Passwörter stimmen nicht überein.');
      setLoading(false);
      return;
    }

    if (password.length < 8) {
        setError('Das Passwort muss mindestens 8 Zeichen lang sein.');
        setLoading(false);
        return;
    }

    if (!email.endsWith('@' + UNI_EMAIL_DOMAIN)) {
      setError(`Die Registrierung ist nur mit einer gültigen ${UNI_EMAIL_DOMAIN} E-Mail Adresse möglich.`);
      setLoading(false);
      return;
    }

    if (Number.isNaN(campusid) || Number.isNaN(disciplineid)) {
        setError('Campus-ID und Studiengang-ID müssen Zahlen sein.');
        setLoading(false);
        return;
    }

    if (campusid < 0 || disciplineid < 0) {
        setError('Campus-ID und Studiengang-ID müssen positive Werte haben.');
        setLoading(false);
        return;
    }

    try {
      const newUser = await api.registerUser({
          email,
          name,
          password,
          campusid,
          disciplineid
      });
      console.log('Registrierung erfolgreich:', newUser);
      try {
        const loggedInUser = await api.loginUser({ email, password });
        login(loggedInUser);
        navigate('/dashboard');
      } catch (loginError: unknown) {
        const message =
          loginError instanceof Error
            ? loginError.message
            : 'Bitte manuell einloggen.';
        setError(`Registrierung erfolgreich, aber Login fehlgeschlagen: ${message}`);
        navigate('/login');
      }
    } catch (err: unknown) {
      console.error('Registrierungsfehler:', err);
      const message =
        err instanceof Error ? err.message : 'Registrierung fehlgeschlagen. Bitte versuchen Sie es erneut.';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container component="main" maxWidth="xs">
      <CssBaseline />
      <Box
        sx={{
          marginTop: 8,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
        }}
      >
        <Avatar sx={{ m: 1, bgcolor: 'secondary.main' }}>
          <LockOutlinedIcon />
        </Avatar>
        <Typography component="h1" variant="h5">
          Account erstellen
        </Typography>
        <Box component="form" noValidate onSubmit={handleSubmit} sx={{ mt: 3, width: '100%' }}>
          {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
          <Stack spacing={2}>
            <TextField
              required
              fullWidth
              id="name"
              label="Name"
              name="name"
              autoComplete="name"
              autoFocus
            />
            <TextField
              required
              fullWidth
              id="email"
              label="E-Mail Adresse"
              name="email"
              autoComplete="email"
            />
            <TextField
              required
              fullWidth
              name="password"
              label="Passwort"
              type="password"
              id="password"
              autoComplete="new-password"
            />
            <TextField
              required
              fullWidth
              name="confirmPassword"
              label="Passwort bestätigen"
              type="password"
              id="confirmPassword"
              autoComplete="new-password"
            />
            <TextField
              select
              required
              fullWidth
              name="campusid"
              label="Campus"
              id="campusid"
              defaultValue=""
              helperText="Bitte wähle deinen Campus aus."
            >
              {campusOptions.map((option) => (
                <MenuItem key={option.value} value={option.value}>
                  {option.label}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              select
              required
              fullWidth
              name="disciplineid"
              label="Studiengang"
              id="disciplineid"
              defaultValue=""
              helperText="Bitte wähle deinen Studiengang aus."
            >
              {disciplineOptions.map((option) => (
                <MenuItem key={option.value} value={option.value}>
                  {option.label}
                </MenuItem>
              ))}
            </TextField>
          </Stack>
          <Button
            type="submit"
            fullWidth
            variant="contained"
            disabled={loading}
            sx={{ mt: 3, mb: 2 }}
          >
            {loading ? 'Registriere...' : 'Registrieren'}
          </Button>
          <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Link component={RouterLink} to="/login" variant="body2">
              Hast du schon einen Account? Einloggen
            </Link>
          </Box>
        </Box>
      </Box>
    </Container>
  );
}
