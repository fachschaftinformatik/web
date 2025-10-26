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
import { Link as RouterLink } from 'react-router-dom';

// ⚠️ WICHTIG: Ersetze 'deine-uni.de' mit der E-Mail-Domain eurer Universität
const UNI_EMAIL_DOMAIN = 'deine-uni.de';



export default function RegistrationPage() {
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  // Diese Funktion wird beim Absenden des Formulars ausgeführt
  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true); // Ladezustand aktivieren
    setError(''); // Alte Fehler zurücksetzen

    const data = new FormData(event.currentTarget);
    const email = data.get('email') as string;
    const password = data.get('password') as string;
    const confirmPassword = data.get('confirmPassword') as string;

    // Simuliere eine kurze Netzwerkverzögerung (wie im Toolpad-Beispiel)
    setTimeout(() => {
      // 1. Validierung: Uni-E-Mail-Domain prüfen
      if (!email.endsWith('@' + UNI_EMAIL_DOMAIN)) {
        setError(`Registrierung nur mit einer @${UNI_EMAIL_DOMAIN} E-Mail möglich.`);
        setLoading(false); // Ladezustand beenden
        return;
      }

      // 2. Validierung: Passwörter vergleichen
      if (password !== confirmPassword) {
        setError('Die Passwörter stimmen nicht überein.');
        setLoading(false); // Ladezustand beenden
        return;
      }

      // Wenn alles erfolgreich ist
      console.log('Registrierung erfolgreich! Sende Daten an Backend:', { email });
      alert(`Account für ${email} wird erstellt!`); // Erfolgs-Feedback
      setLoading(false); // Ladezustand beenden
      // Hier würdest du die Daten tatsächlich an dein Backend senden
    }, 500); // 0.5 Sekunden Verzögerung
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
            {/* Zeigt die Fehlermeldung an, falls eine vorhanden ist */}
            {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

            <Stack spacing={2}>
              <TextField
                required
                fullWidth
                id="email"
                label="Uni-E-Mail Adresse"
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
                autoComplete="new-password"
              />
              <TextField
                required
                fullWidth
                name="confirmPassword"
                label="Passwort bestätigen"
                type="password"
                id="confirmPassword"
              />
            </Stack>

            <Button
              type="submit"
              fullWidth
              variant="contained"
              disabled={loading} // Button wird während des Ladens deaktiviert
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