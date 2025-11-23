import React, { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Link as RouterLink } from 'react-router-dom';
import { z } from 'zod';

import Avatar from '@mui/material/Avatar';
import Button from '@mui/material/Button';
import CssBaseline from '@mui/material/CssBaseline';
import TextField from '@mui/material/TextField';
import Link from '@mui/material/Link';
import Box from '@mui/material/Box';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import Typography from '@mui/material/Typography';
import Container from '@mui/material/Container';
import Alert from '@mui/material/Alert';
import Stack from '@mui/material/Stack';
import FormControlLabel from '@mui/material/FormControlLabel';
import Checkbox from '@mui/material/Checkbox';
import IconButton from '@mui/material/IconButton';
import OutlinedInput from '@mui/material/OutlinedInput';
import InputLabel from '@mui/material/InputLabel';
import InputAdornment from '@mui/material/InputAdornment';
import FormControl from '@mui/material/FormControl';
import FormHelperText from '@mui/material/FormHelperText';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';

import { postAuthLogin } from '@lib/api';
import { useAuth, REMEMBERED_FLAG_KEY } from '@lib/auth';

const loginSchema = z.object({
  email: z.string().min(1, "Bitte gib deine E-Mail-Adresse ein.").email("Diese E-Mail-Adresse scheint nicht gültig zu sein."),
  password: z.string().min(1, "Bitte gib dein Passwort ein."),
});

type LoginFormData = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState('');
  const [errors, setErrors] = useState<Partial<Record<keyof LoginFormData, string>> & { global?: string }>({});
  const [showPassword, setShowPassword] = useState(false);
  const [rememberMe, setRememberMe] = useState(() => {
    if (typeof window === 'undefined') return false;
    return window.localStorage.getItem(REMEMBERED_FLAG_KEY) === 'true';
  });

  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { login } = useAuth();

  useEffect(() => {
    if (searchParams.get('verified') === 'true') {
      setSuccess('Deine E-Mail wurde erfolgreich bestätigt! Du kannst dich jetzt anmelden.');
    }
  }, [searchParams]);

  const handleClickShowPassword = () => setShowPassword((show) => !show);
  
  const handleMouseDownPassword = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    setErrors({});
    setSuccess('');

    const formData = new FormData(event.currentTarget);
    const rawData = {
      email: formData.get('email') as string,
      password: formData.get('password') as string,
    };

    const validationResult = loginSchema.safeParse(rawData);

    if (!validationResult.success) {
      const fieldErrors: typeof errors = {};
      validationResult.error.issues.forEach((issue) => {
        const path = issue.path[0] as keyof LoginFormData;
        fieldErrors[path] = issue.message;
      });
      setErrors(fieldErrors);
      setLoading(false);
      return;
    }

    const { email, password } = validationResult.data;

    try {
      const { data: user, error: apiError } = await postAuthLogin({
        body: { email, password }
      });

      if (apiError) {
        // @ts-ignore
        const msg = (apiError as any)?.message || 'Anmeldung fehlgeschlagen. Bitte überprüfe deine Daten.';
        setErrors({ global: msg });
        setLoading(false);
        return;
      }

      if (!user) {
        throw new Error('Ein unerwarteter Fehler ist aufgetreten.');
      }

      login(user, rememberMe);
      navigate('/dashboard');
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Anmeldung fehlgeschlagen.';
      setErrors({ global: message });
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
          Anmelden
        </Typography>

        <Box component="form" noValidate onSubmit={handleSubmit} sx={{ mt: 3, width: '100%' }}>
          
          {success && <Alert severity="success" sx={{ mb: 2 }}>{success}</Alert>}
          {errors.global && <Alert severity="error" sx={{ mb: 2 }}>{errors.global}</Alert>}

          <Stack spacing={2}>
            <TextField
              required
              fullWidth
              id="email"
              label="E-Mail Adresse"
              name="email"
              autoComplete="email"
              autoFocus
              disabled={loading}
              error={!!errors.email}
              helperText={errors.email}
            />
            
            <FormControl variant="outlined" required error={!!errors.password} fullWidth>
              <InputLabel htmlFor="outlined-adornment-password">Passwort</InputLabel>
              <OutlinedInput
                id="outlined-adornment-password"
                name="password"
                type={showPassword ? 'text' : 'password'}
                disabled={loading}
                endAdornment={
                  <InputAdornment position="end">
                    <IconButton
                      aria-label={showPassword ? 'Passwort verbergen' : 'Passwort anzeigen'}
                      onClick={handleClickShowPassword}
                      onMouseDown={handleMouseDownPassword}
                      edge="end"
                    >
                      {showPassword ? <VisibilityOff /> : <Visibility />}
                    </IconButton>
                  </InputAdornment>
                }
                label="Passwort"
                autoComplete="current-password"
              />
              {errors.password && <FormHelperText>{errors.password}</FormHelperText>}
            </FormControl>

            <FormControlLabel
              control={
                <Checkbox
                  color="primary"
                  checked={rememberMe}
                  onChange={(event) => setRememberMe(event.target.checked)}
                  name="remember"
                  disabled={loading}
                />
              }
              label="Angemeldet bleiben"
              sx={{ width: '100%', ml: -1 }} 
            />
          </Stack>

          <Button
            type="submit"
            fullWidth
            variant="contained"
            disabled={loading}
            sx={{ mt: 3, mb: 2, py: 1.2 }}
          >
            {loading ? 'Melde an...' : 'Anmelden'}
          </Button>
          
          <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Link component={RouterLink} to="/register" variant="body2">
              Noch kein Konto? Jetzt registrieren
            </Link>
          </Box>
        </Box>
      </Box>
    </Container>
  );
}
