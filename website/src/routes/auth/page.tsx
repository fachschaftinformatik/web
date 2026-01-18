import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation, useSearchParams } from 'react-router-dom';
import { z } from 'zod';

import {
    Avatar, Button, CssBaseline, TextField, Box, Typography,
    Container, Alert, Stack, FormControlLabel, Checkbox, IconButton,
    OutlinedInput, InputLabel, InputAdornment, FormControl, FormHelperText,
    Tooltip, MenuItem, Tab, Tabs, Zoom, Fade
} from '@mui/material';
import { useTheme, alpha } from '@mui/material/styles';
import {
    Visibility, VisibilityOff,
    Brightness4Rounded, Brightness7Rounded,
    PersonAddRounded as PersonAddIcon,
    LoginRounded as LoginIcon,
    ArrowBackRounded,
    ErrorOutlineRounded,
    CheckCircleOutlineRounded
} from '@mui/icons-material';

import { useThemeMode } from '@lib/theme';
import { postAuthLogin, postAuthRegister, getPrograms } from '@lib/api';
import type { AuthProgramResponse as Program } from '@lib/api';
import { useAuth, REMEMBERED_FLAG_KEY } from '@lib/auth';

const EMAIL_SUFFIX = '@studmail.w-hs.de';

const loginSchema = z.object({
    emailPrefix: z.string().min(1, "Bitte gib dein E-Mail-Kürzel ein."),
    password: z.string().min(1, "Bitte gib dein Passwort ein."),
});

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
    programid: z.coerce.number().min(1, "Bitte wähle einen Studiengang aus."),
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

const translateError = (err: unknown): string => {
    const errorObj = err as { error?: string, statusText?: string, message?: string };
    const code = errorObj?.error || errorObj?.statusText?.toLowerCase() || '';
    const message = errorObj?.message || '';

    const translations: Record<string, string> = {
        'invalid_credentials': 'Ungültige Anmeldedaten.',
        'unauthorized': 'Ungültige Anmeldedaten.',
        'email_exists': 'Diese E-Mail-Adresse ist bereits registriert.',
        'user_not_found': 'Benutzer nicht gefunden.',
        'email_not_verified': message || 'Bitte bestätige erst deine E-Mail-Adresse.',
        'invalid_request_body': 'Ungültige Anfrage. Bitte fülle alle Felder korrekt aus.',
        'server_error': 'Serverfehler. Bitte versuche es später noch einmal.',
        'database_error': 'Datenbankfehler. Bitte versuche es später noch einmal.',
        'invalid_csrf': 'Sitzung abgelaufen. Bitte lade die Seite neu.'
    };

    if (translations[code]) return translations[code];

    // Fuzzy matching for messages if code doesn't match
    const msg = message.toLowerCase();
    if (msg.includes('invalid email or password')) return 'Ungültige Anmeldedaten.';
    if (msg.includes('already exists')) return 'Diese E-Mail-Adresse ist bereits registriert.';
    if (msg.includes('failed to fetch')) return 'Verbindung zum Server fehlgeschlagen. Bitte prüfe deine Internetverbindung.';
    if (msg.includes('network error')) return 'Netzwerkfehler. Bitte prüfe deine Verbindung.';
    if (msg.includes('timeout')) return 'Zeitüberschreitung. Der Server antwortet nicht.';

    return message || 'Ein unerwarteter Fehler ist aufgetreten.';
};

export default function AuthPage() {
    const navigate = useNavigate();
    const location = useLocation();
    const [searchParams] = useSearchParams();
    const initialTab = location.pathname === '/register' ? 1 : 0;
    const [tabValue, setTabValue] = useState(initialTab);

    const [loading, setLoading] = useState(false);
    const [success, setSuccess] = useState('');
    const [errors, setErrors] = useState<Record<string, string>>({});
    const [showPassword, setShowPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);
    const [programs, setPrograms] = useState<Program[]>([]);
    const [rememberMe, setRememberMe] = useState(() => {
        if (typeof window === 'undefined') return false;
        return window.localStorage.getItem(REMEMBERED_FLAG_KEY) === 'true';
    });


    const { login } = useAuth();
    const { mode, setPreference } = useThemeMode();
    const theme = useTheme();

    useEffect(() => {
        document.title = tabValue === 0 ? "Anmelden | FSV Informatik" : "Registrieren | FSV Informatik";
        if (tabValue === 1 && programs.length === 0) {
            getPrograms()
                .then(({ data }) => { if (data) setPrograms(data); })
                .catch((err) => console.error("Fehler beim Laden der Studiengänge", err));
        }
    }, [tabValue, programs.length]);

    useEffect(() => {
        if (searchParams.get('verified') === 'true') {
            setSuccess('Deine E-Mail wurde erfolgreich bestätigt! Du kannst dich jetzt anmelden.');
            setTabValue(0);
        }
    }, [searchParams]);

    const handleTabChange = (_: React.SyntheticEvent, newValue: number) => {
        setTabValue(newValue);
        setErrors({});
        setSuccess('');
        navigate(newValue === 1 ? '/register' : '/login');
    };

    const handleThemeToggle = () => setPreference(mode === 'light' ? 'dark' : 'light');

    const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault();
        setLoading(true);
        setErrors({});
        setSuccess('');

        const formData = new FormData(event.currentTarget);
        const rawData = Object.fromEntries(formData.entries()) as Record<string, string>;

        if (tabValue === 0) {
            // Login Flow
            const validationResult = loginSchema.safeParse(rawData);
            if (!validationResult.success) {
                const fieldErrors: Record<string, string> = {};
                validationResult.error.issues.forEach((issue) => { fieldErrors[issue.path[0] as string] = issue.message; });
                setErrors(fieldErrors);
                setLoading(false);
                return;
            }

            const { emailPrefix, password } = validationResult.data;
            try {
                const { data: user, error: apiError } = await postAuthLogin({
                    body: { email: `${emailPrefix}${EMAIL_SUFFIX}`, password }
                });

                if (apiError) {
                    setErrors({ global: translateError(apiError) });
                    setLoading(false);
                    return;
                }
                if (user) {
                    login(user, rememberMe);
                    navigate(`/user/${user.id}`);
                }
            } catch (err) {
                setErrors({ global: translateError(err) });
            } finally {
                setLoading(false);
            }
        } else {
            // Register Flow
            const validationResult = registrationSchema.safeParse({
                ...rawData,
                programid: Number(rawData.programid)
            });
            if (!validationResult.success) {
                const fieldErrors: Record<string, string> = {};
                validationResult.error.issues.forEach((issue) => { fieldErrors[issue.path[0] as string] = issue.message; });
                setErrors(fieldErrors);
                setLoading(false);
                return;
            }

            const validData = validationResult.data;
            try {
                const { data: newUser, error: apiError } = await postAuthRegister({
                    body: {
                        email: `${validData.emailPrefix}${EMAIL_SUFFIX}`,
                        name: validData.name,
                        password: validData.password,
                        programid: validData.programid,
                    }
                });

                if (apiError) {
                    setErrors({ global: translateError(apiError) });
                    setLoading(false);
                    return;
                }
                if (newUser) {
                    setSuccess('Account erstellt! Bitte bestätige deine E-Mail-Adresse.');
                }
            } catch (err) {
                setErrors({ global: translateError(err) });
            } finally {
                setLoading(false);
            }
        }
    };

    const toggleBtnStyle = {
        border: '1px solid ' + alpha(theme.palette.primary.main, mode === 'dark' ? 0.5 : 0.25),
        bgcolor: mode === 'dark' ? alpha(theme.palette.primary.dark, 0.35) : alpha(theme.palette.primary.main, 0.1),
        color: mode === 'dark' ? theme.palette.primary.contrastText : theme.palette.primary.main,
        '&:hover': {
            bgcolor: mode === 'dark' ? alpha(theme.palette.primary.dark, 0.55) : alpha(theme.palette.primary.main, 0.2)
        }
    };

    return (
        <Container component="main" maxWidth="xs">
            <CssBaseline />
            <Box sx={{ marginTop: 8, display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                <Zoom in={true} style={{ transitionDelay: '100ms' }}>
                    <Avatar sx={{ m: 1, bgcolor: 'primary.main', width: 56, height: 56, boxShadow: 3 }}>
                        {tabValue === 0 ? <LoginIcon fontSize="large" /> : <PersonAddIcon fontSize="large" />}
                    </Avatar>
                </Zoom>
                <Typography component="h1" variant="h4" fontWeight={800} sx={{ mt: 2, mb: 3 }}>
                    {tabValue === 0 ? 'Willkommen zurück' : 'Account erstellen'}
                </Typography>

                <Box sx={{ width: '100%', mb: 4 }}>
                    <Tabs
                        value={tabValue}
                        onChange={handleTabChange}
                        variant="fullWidth"
                        sx={{
                            '& .MuiTabs-indicator': { height: 3, borderRadius: '3px 3px 0 0' },
                            bgcolor: alpha(theme.palette.primary.main, 0.05),
                            borderRadius: 2
                        }}
                    >
                        <Tab label="Anmelden" sx={{ fontWeight: 700 }} />
                        <Tab label="Registrieren" sx={{ fontWeight: 700 }} />
                    </Tabs>
                </Box>

                <Fade in={true} key={tabValue}>
                    <Box component="form" noValidate onSubmit={handleSubmit} sx={{ width: '100%' }}>
                        {success && (
                            <Zoom in={!!success}>
                                <Alert
                                    icon={<CheckCircleOutlineRounded fontSize="inherit" />}
                                    severity="success"
                                    sx={{
                                        mb: 3,
                                        borderRadius: 3,
                                        bgcolor: alpha(theme.palette.success.main, 0.1),
                                        backdropFilter: 'blur(10px)',
                                        border: `1px solid ${alpha(theme.palette.success.main, 0.2)}`,
                                        color: theme.palette.success.dark,
                                        fontWeight: 600,
                                        '& .MuiAlert-icon': { color: theme.palette.success.main }
                                    }}
                                >
                                    {success}
                                </Alert>
                            </Zoom>
                        )}
                        {errors.global && (
                            <Zoom in={!!errors.global}>
                                <Alert
                                    icon={<ErrorOutlineRounded fontSize="inherit" />}
                                    severity="error"
                                    sx={{
                                        mb: 3,
                                        borderRadius: 3,
                                        bgcolor: alpha(theme.palette.error.main, 0.1),
                                        backdropFilter: 'blur(10px)',
                                        border: `1px solid ${alpha(theme.palette.error.main, 0.2)}`,
                                        color: theme.palette.error.dark,
                                        fontWeight: 600,
                                        '& .MuiAlert-icon': { color: theme.palette.error.main }
                                    }}
                                >
                                    {errors.global}
                                </Alert>
                            </Zoom>
                        )}

                        <Stack spacing={2.5}>
                            {tabValue === 1 && (
                                <TextField
                                    required fullWidth id="name" label="Vollständiger Name" name="name"
                                    autoComplete="name" autoFocus disabled={loading || !!success}
                                    error={!!errors.name} helperText={errors.name}
                                />
                            )}

                            <TextField
                                required fullWidth id="emailPrefix" label="E-Mail" name="emailPrefix"
                                autoComplete="email" autoFocus={tabValue === 0} disabled={loading || !!success}
                                error={!!errors.emailPrefix} helperText={errors.emailPrefix}
                                InputProps={{
                                    endAdornment: <InputAdornment position="end">{EMAIL_SUFFIX}</InputAdornment>,
                                }}
                            />

                            {tabValue === 1 && (
                                <TextField
                                    select required fullWidth name="programid" label="Studiengang" id="programid"
                                    defaultValue="" disabled={loading || !!success} error={!!errors.programid} helperText={errors.programid}
                                >
                                    {programs.length > 0 ? (
                                        programs.map((opt) => (
                                            <MenuItem key={opt.id} value={opt.id}>{opt.name}</MenuItem>
                                        ))
                                    ) : (
                                        <MenuItem disabled value="">Lädt Studiengänge...</MenuItem>
                                    )}
                                </TextField>
                            )}

                            <FormControl variant="outlined" required error={!!errors.password} fullWidth>
                                <InputLabel htmlFor="auth-password">Passwort</InputLabel>
                                <OutlinedInput
                                    id="auth-password" name="password" type={showPassword ? 'text' : 'password'}
                                    disabled={loading || !!success}
                                    endAdornment={
                                        <InputAdornment position="end">
                                            <IconButton onClick={() => setShowPassword(!showPassword)} edge="end">
                                                {showPassword ? <VisibilityOff /> : <Visibility />}
                                            </IconButton>
                                        </InputAdornment>
                                    }
                                    label="Passwort"
                                />
                                {errors.password && <FormHelperText>{errors.password}</FormHelperText>}
                            </FormControl>

                            {tabValue === 1 && (
                                <FormControl variant="outlined" required error={!!errors.confirmPassword} fullWidth>
                                    <InputLabel htmlFor="confirm-password">Passwort bestätigen</InputLabel>
                                    <OutlinedInput
                                        id="confirm-password" name="confirmPassword" type={showConfirmPassword ? 'text' : 'password'}
                                        disabled={loading || !!success}
                                        endAdornment={
                                            <InputAdornment position="end">
                                                <IconButton onClick={() => setShowConfirmPassword(!showConfirmPassword)} edge="end">
                                                    {showConfirmPassword ? <VisibilityOff /> : <Visibility />}
                                                </IconButton>
                                            </InputAdornment>
                                        }
                                        label="Passwort bestätigen"
                                    />
                                    {errors.confirmPassword && <FormHelperText>{errors.confirmPassword}</FormHelperText>}
                                </FormControl>
                            )}

                            {tabValue === 0 && (
                                <FormControlLabel
                                    control={
                                        <Checkbox
                                            checked={rememberMe}
                                            onChange={(e) => setRememberMe(e.target.checked)}
                                            color="primary"
                                        />
                                    }
                                    label="Angemeldet bleiben"
                                    sx={{ ml: -0.5 }}
                                />
                            )}
                        </Stack>

                        <Button
                            type="submit" fullWidth variant="contained" size="large"
                            disabled={loading || !!success}
                            sx={{ mt: 4, mb: 2, height: 52, borderRadius: 2, fontWeight: 700 }}
                        >
                            {loading ? (tabValue === 0 ? 'Melde an...' : 'Erstelle Account...') : (tabValue === 0 ? 'Anmelden' : 'Registrieren')}
                        </Button>

                        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 1 }}>
                            <Button
                                startIcon={<ArrowBackRounded />}
                                onClick={() => navigate('/')}
                                sx={{
                                    color: 'text.secondary',
                                    textTransform: 'none',
                                    '&:hover': { color: 'primary.main', bgcolor: 'transparent', textDecoration: 'underline' }
                                }}
                            >
                                Zurück zur Startseite
                            </Button>
                        </Box>

                        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 3 }}>
                            <Tooltip title="Farbschema wechseln">
                                <IconButton onClick={handleThemeToggle} sx={{ ...toggleBtnStyle, p: 1.5 }}>
                                    {mode === 'dark' ? <Brightness7Rounded /> : <Brightness4Rounded />}
                                </IconButton>
                            </Tooltip>
                        </Box>
                    </Box>
                </Fade>
            </Box>
        </Container>
    );
}
