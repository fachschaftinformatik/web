import * as React from 'react';
import { useState } from 'react';
import Avatar from '@mui/material/Avatar';
import Button from '@mui/material/Button';
import CssBaseline from '@mui/material/CssBaseline';
import TextField from '@mui/material/TextField';
import Link from '@mui/material/Link';
import Box from '@mui/material/Box';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import Typography from '@mui/material/Typography';
import Container from '@mui/material/Container';

import { Alert, Stack } from '@mui/material';

// NEU: Importiere den Link-Router, um zwischen Seiten zu wechseln
import { Link as RouterLink } from 'react-router-dom';



export default function LoginPage() {
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    setError('');

    const data = new FormData(event.currentTarget);
    const email = data.get('email') as string;
    const password = data.get('password') as string;

    // Simuliere eine kurze Netzwerkverzögerung
    setTimeout(() => {
      // Hier würdest du die Login-Daten an dein Backend senden
      if (!email || !password) {
        setError('Bitte E-Mail und Passwort eingeben.');
        setLoading(false);
        return;
      }

      console.log('Login-Versuch mit:', { email });
      alert(`Login für ${email} wird versucht!`);
      setLoading(false);
    }, 500);
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
            Einloggen
          </Typography>
          <Box component="form" noValidate onSubmit={handleSubmit} sx={{ mt: 3, width: '100%' }}>
            {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

            <Stack spacing={2}>
              <TextField
                required
                fullWidth
                id="email"
                label="E-Mail Adresse"
                name="email"
                autoComplete="email"
                autoFocus
              />
              <TextField
                required
                fullWidth
                name="password"
                label="Passwort"
                type="password"
                id="password"
                autoComplete="current-password"
              />
            </Stack>

            <Button
              type="submit"
              fullWidth
              variant="contained"
              disabled={loading}
              sx={{ mt: 3, mb: 2 }}
            >
              {loading ? 'Logge ein...' : 'Einloggen'}
            </Button>
            
            <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
              {/* NEU: Dieser Link führt jetzt zur Registrierungsseite */}
              <Link component={RouterLink} to="/register" variant="body2">
                Hast du noch keinen Account? Registrieren
              </Link>
            </Box>
          </Box>
        </Box>
      </Container>

  );
}