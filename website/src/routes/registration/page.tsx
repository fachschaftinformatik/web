import React, { useState, useEffect } from 'react';
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
import MenuItem from '@mui/material/MenuItem';
import Stack from '@mui/material/Stack';
import IconButton from '@mui/material/IconButton';
import OutlinedInput from '@mui/material/OutlinedInput';
import InputLabel from '@mui/material/InputLabel';
import InputAdornment from '@mui/material/InputAdornment';
import FormControl from '@mui/material/FormControl';
import FormHelperText from '@mui/material/FormHelperText';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';

import { getPrograms, postAuthRegister } from '@lib/api';
import type { Program } from '@lib/api';

const EMAIL_SUFFIX = '@studmail.w-hs.de';

const registrationSchema = z.object({
  name: z.string().min(1, "Bitte gib deinen Namen ein."),
  emailPrefix: z.string().min(1, "Bitte gib dein E-Mail ein.").regex(/^[a-zA-Z0-9._-]+$/, "Die E-Mail enthält ungültige Zeichen."),
  password: z.string()
    .min(12, "Das Passwort muss mindestens 12 Zeichen lang sein.")
    .regex(/[A-Z]/, "Das Passwort muss mindestens einen Großbuchstaben enthalten.")
    .regex(/[a-z]/, "Das Passwort muss mindestens einen Kleinbuchstaben enthalten.")
    .regex(/[0-9]/, "Das Passwort muss mindestens eine Zahl enthalten.")
    .regex(/[^A-Za-z0-9]/, "Das Passwort muss mindestens ein Sonderzeichen enthalten."),
  confirmPassword: z.string(),
  programid: z.number({ 
    invalid_type_error: "Bitte wähle einen Studiengang aus." 
  })
  .int()
  .positive({ message: "Bitte wähle einen Studiengang aus." }),
}).superRefine((data, ctx) => {
  if (data.password !== data.confirmPassword) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: "Die Passwörter stimmen nicht überein.",
      path: ["confirmPassword"],
    });
  }
  if (data.name && data.password.toLowerCase().includes(data.name.toLowerCase())) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: "Das Passwort darf deinen Namen nicht enthalten.",
      path: ["password"],
    });
  }
  if (data.emailPrefix && data.password.toLowerCase().includes(data.emailPrefix.toLowerCase())) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: "Das Passwort darf deine E-Mail nicht enthalten.",
      path: ["password"],
    });
  }
});

type RegistrationFormData = z.infer<typeof registrationSchema>;

export default function RegistrationPage() {
  const [programs, setPrograms] = useState<Program[]>([]);
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState('');
  const [errors, setErrors] = useState<Partial<Record<keyof RegistrationFormData, string>> & { global?: string }>({});
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  useEffect(() => {
    getPrograms()
      .then(({ data }) => {
        if (data) setPrograms(data);
      })
      .catch((err) => console.error("Failed to fetch programs", err));
  }, []);

  const handleClickShowPassword = () => setShowPassword((show) => !show);
  const handleClickShowConfirmPassword = () => setShowConfirmPassword((show) => !show);
  
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
      name: formData.get('name') as string,
      emailPrefix: formData.get('emailPrefix') as string,
      password: formData.get('password') as string,
      confirmPassword: formData.get('confirmPassword') as string,
      programid: Number(formData.get('programid')),
    };

    const validationResult = registrationSchema.safeParse(rawData);

    if (!validationResult.success) {
      const fieldErrors: typeof errors = {};
      validationResult.error.issues.forEach((issue) => {
        const path = issue.path[0] as keyof RegistrationFormData;
        fieldErrors[path] = issue.message;
      });
      setErrors(fieldErrors);
      setLoading(false);
      return;
    }

    const validData = validationResult.data;
    const fullEmail = `${validData.emailPrefix}${EMAIL_SUFFIX}`;

    try {
      const { data: newUser, error: apiError } = await postAuthRegister({
        body: {
          email: fullEmail,
          name: validData.name,
          password: validData.password,
          programid: validData.programid,
        }
      });

      if (apiError) {
        // @ts-ignore
        const msg = (apiError as any)?.message || "Die Registrierung konnte nicht abgeschlossen werden.";
        setErrors({ global: msg });
        return;
      }

      if (newUser) {
        console.log('Registrierung erfolgreich:', newUser);
        setSuccess('Dein Account wurde erfolgreich erstellt! Bitte bestätige jetzt deine E-Mail-Adresse.');
      }

    } catch (err: unknown) {
      console.error('Network error:', err);
      setErrors({ global: 'Ein unerwarteter Fehler ist aufgetreten. Bitte versuche es später erneut.' });
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
          
          {errors.global && <Alert severity="error" sx={{ mb: 2 }}>{errors.global}</Alert>}
          {success && (
            <Alert severity="success" sx={{ mb: 2 }}>
              {success}
              <br />
              <Link component={RouterLink} to="/login" sx={{ mt: 1, display: 'block', fontWeight: 'bold' }}>
                Zum Login
              </Link>
            </Alert>
          )}

          <Stack spacing={2}>
            <TextField
              required
              fullWidth
              id="name"
              label="Name"
              name="name"
              autoComplete="name"
              autoFocus
              disabled={!!success || loading}
              error={!!errors.name}
              helperText={errors.name}
            />

            <TextField
              required
              fullWidth
              id="emailPrefix"
              label="E-Mail"
              name="emailPrefix"
              autoComplete="email"
              disabled={!!success || loading}
              error={!!errors.emailPrefix}
              helperText={errors.emailPrefix}
              slotProps={{
                input: {
                  endAdornment: <InputAdornment position="end">{EMAIL_SUFFIX}</InputAdornment>,
                },
              }}
            />

            <TextField
              select
              required
              fullWidth
              name="programid"
              label="Studiengang"
              id="programid"
              defaultValue=""
              disabled={!!success || loading}
              error={!!errors.programid}
              helperText={errors.programid}
            >
              {programs.map((option) => (
                <MenuItem key={option.id} value={option.id}>
                  {option.name}
                </MenuItem>
              ))}
            </TextField>

            <FormControl variant="outlined" required error={!!errors.password} fullWidth>
              <InputLabel htmlFor="outlined-adornment-password">Passwort</InputLabel>
              <OutlinedInput
                id="outlined-adornment-password"
                name="password"
                type={showPassword ? 'text' : 'password'}
                disabled={!!success || loading}
                endAdornment={
                  <InputAdornment position="end">
                    <IconButton
                      aria-label="toggle password visibility"
                      onClick={handleClickShowPassword}
                      onMouseDown={handleMouseDownPassword}
                      edge="end"
                    >
                      {showPassword ? <VisibilityOff /> : <Visibility />}
                    </IconButton>
                  </InputAdornment>
                }
                label="Passwort"
              />
              {errors.password && <FormHelperText>{errors.password}</FormHelperText>}
            </FormControl>

            <FormControl variant="outlined" required error={!!errors.confirmPassword} fullWidth>
              <InputLabel htmlFor="outlined-adornment-confirm-password">Passwort bestätigen</InputLabel>
              <OutlinedInput
                id="outlined-adornment-confirm-password"
                name="confirmPassword"
                type={showConfirmPassword ? 'text' : 'password'}
                disabled={!!success || loading}
                endAdornment={
                  <InputAdornment position="end">
                    <IconButton
                      aria-label="toggle password visibility"
                      onClick={handleClickShowConfirmPassword}
                      onMouseDown={handleMouseDownPassword}
                      edge="end"
                    >
                      {showConfirmPassword ? <VisibilityOff /> : <Visibility />}
                    </IconButton>
                  </InputAdornment>
                }
                label="Passwort bestätigen"
              />
              {errors.confirmPassword && <FormHelperText>{errors.confirmPassword}</FormHelperText>}
            </FormControl>

          </Stack>

          {!success && (
            <Button
              type="submit"
              fullWidth
              variant="contained"
              disabled={loading}
              sx={{ mt: 3, mb: 2, py: 1.2 }}
            >
              {loading ? 'Wird registriert...' : 'Registrieren'}
            </Button>
          )}

          <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Link component={RouterLink} to="/login" variant="body2">
              Bereits registriert? Anmelden
            </Link>
          </Box>
        </Box>
      </Box>
    </Container>
  );
}
