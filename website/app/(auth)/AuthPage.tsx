'use client';

import React, { useState, useEffect } from 'react';
import { useRouter, useSearchParams, usePathname } from 'next/navigation';
import { z } from 'zod';
import Link from 'next/link';

import Avatar from '@mui/material/Avatar';
import Button from '@mui/material/Button';
import CssBaseline from '@mui/material/CssBaseline';
import TextField from '@mui/material/TextField';
import Box from '@mui/material/Box';
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
import Tooltip from '@mui/material/Tooltip';
import MenuItem from '@mui/material/MenuItem';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Zoom from '@mui/material/Zoom';
import Fade from '@mui/material/Fade';
import { useTheme, alpha } from '@mui/material/styles';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import Brightness4Rounded from '@mui/icons-material/Brightness4Rounded';
import Brightness7Rounded from '@mui/icons-material/Brightness7Rounded';
import PersonAddIcon from '@mui/icons-material/PersonAddRounded';
import LoginIcon from '@mui/icons-material/LoginRounded';
import ArrowBackRounded from '@mui/icons-material/ArrowBackRounded';
import ErrorOutlineRounded from '@mui/icons-material/ErrorOutlineRounded';
import CheckCircleOutlineRounded from '@mui/icons-material/CheckCircleOutlineRounded';
import HelpOutlineRounded from '@mui/icons-material/HelpOutlineRounded';

import { useThemeMode } from '@lib/theme';
import { postAuthRegister, getPrograms } from '@lib/api';
import type { DtoProgramResponse as Program } from '@lib/api';
import { useAuth } from '@lib/auth';
import { translateError } from '@lib/errors';
import { zDtoLoginRequest, zDtoRegisterRequest } from '@lib/api/zod.gen';
import { signIn } from 'next-auth/react';

const ALLOWED_DOMAINS = ['@studmail.w-hs.de', '@fsv-wh.de'];

const loginSchema = zDtoLoginRequest.omit({ email: true }).extend({
    emailPrefix: z.string().min(1, "Bitte gib dein E-Mail-Kürzel ein."),
    emailDomain: z.string().min(1),
    password: z.string().min(1, "Bitte gib dein Passwort ein."),
});

const registrationSchema = zDtoRegisterRequest.omit({ email: true, program_id: true }).extend({
    name: z.string().min(1, "Bitte gib deinen Namen ein."),
    emailPrefix: z.string().min(1, "Bitte gib dein E-Mail ein.").regex(/^[a-zA-Z0-9._-]+$/, "Die E-Mail enthält ungültige Zeichen."),
    emailDomain: z.string().min(1),
    password: z.string()
        .min(8, "Das Passwort muss mindestens 8 Zeichen lang sein.")
        .regex(/[A-Z]/, "Das Passwort muss mindestens einen Großbuchstaben enthalten.")
        .regex(/[a-z]/, "Das Passwort muss mindestens einen Kleinbuchstaben enthalten.")
        .regex(/[0-9]/, "Das Passwort muss mindestens eine Zahl enthalten.")
        .regex(/[^A-Za-z0-9]/, "Das Passwort muss mindestens ein Sonderzeichen enthalten."),
    confirmPassword: z.string(),
    program_id: z.string().min(1, "Bitte wähle einen Studiengang aus."),
}).superRefine((data, ctx) => {
    if (data.password !== data.confirmPassword) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: "Die Passwörter stimmen nicht überein.",
            path: ["confirmPassword"],
        });
    }
    if (data.name && data.password && data.password.toLowerCase().includes(data.name.toLowerCase())) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: "Das Passwort darf deinen Namen nicht enthalten.",
            path: ["password"],
        });
    }
    if (data.emailPrefix && data.password && data.password.toLowerCase().includes(data.emailPrefix.toLowerCase())) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: "Das Passwort darf deine E-Mail nicht enthalten.",
            path: ["password"],
        });
    }
});

export default function AuthPage() {
    const router = useRouter();
    const pathname = usePathname();
    const searchParams = useSearchParams();
    const initialTab = pathname === '/register' ? 1 : 0;
    const [tabValue, setTabValue] = useState(initialTab);

    const [loading, setLoading] = useState(false);
    const [success, setSuccess] = useState('');
    const [errors, setErrors] = useState<Record<string, string>>({});
    const [showPassword, setShowPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);
    const [programs, setPrograms] = useState<Program[]>([]);
    const [signupsEnabled, setSignupsEnabled] = useState(true);
    const [rememberMe, setRememberMe] = useState(false);


    const { mode, setPreference } = useThemeMode();
    const theme = useTheme();

    useEffect(() => {
        if (tabValue === 1 && programs.length === 0) {
            getPrograms()
                .then(({ data }) => { if (data) setPrograms(data); })
                .catch((err) => console.error("Fehler beim Laden der Studiengänge", err));
        }
    }, [tabValue, programs.length]);

    useEffect(() => {
        const verified = searchParams.get('verified');
        const error = searchParams.get('error');

        if (verified === 'true') {
            setSuccess('Deine E-Mail wurde erfolgreich bestätigt! Du kannst dich jetzt anmelden.');
            setTabValue(0);
        } else if (verified === 'false' && error) {
            setErrors({ global: translateError({ error }) });
            setTabValue(0);
        }
    }, [searchParams]);

    const handleTabChange = (_: React.SyntheticEvent, newValue: number) => {
        if (newValue === 1 && !signupsEnabled) return;
        setTabValue(newValue);
        setErrors({});
        setSuccess('');
        router.push(newValue === 1 ? '/register' : '/login');
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

            const { emailPrefix, emailDomain, password } = validationResult.data;
            try {
                const result = await signIn('credentials', {
                    email: `${emailPrefix}${emailDomain}`,
                    password,
                    redirect: false,
                });

                if (result?.error) {
                    setErrors({ global: "Ungültige E-Mail oder Passwort." });
                    setLoading(false);
                    return;
                }
                
                router.push('/');
            } catch (err) {
                setErrors({ global: "Anmeldung fehlgeschlagen." });
            } finally {
                setLoading(false);
            }
        } else {
            // Register Flow
            const validationResult = registrationSchema.safeParse(rawData);
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
                        email: `${validData.emailPrefix}${validData.emailDomain}`,
                        name: validData.name,
                        password: validData.password,
                        program_id: validData.program_id,
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
        <Container component="main" maxWidth={false} sx={{ maxWidth: 480 }}>
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
                        {signupsEnabled && <Tab label="Registrieren" sx={{ fontWeight: 700 }} />}
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

                        {tabValue === 0 && (
                            <Alert
                                severity="info"
                                sx={{
                                    mb: 3,
                                    borderRadius: 3,
                                    bgcolor: alpha(theme.palette.info.main, 0.1),
                                    backdropFilter: 'blur(10px)',
                                    border: `1px solid ${alpha(theme.palette.info.main, 0.2)}`,
                                    color: theme.palette.info.dark,
                                    fontWeight: 500,
                                    alignItems: 'center',
                                    '& .MuiAlert-icon': { color: theme.palette.info.main }
                                }}
                            >
                                Die ZA Kennung kann nicht verwendet werden. Um dich anzumelden, musst du dich zuerst registrieren.
                            </Alert>
                        )}

                        {tabValue === 1 && (
                            <Alert
                                severity="info"
                                sx={{
                                    mb: 3,
                                    borderRadius: 3,
                                    bgcolor: alpha(theme.palette.info.main, 0.1),
                                    backdropFilter: 'blur(10px)',
                                    border: `1px solid ${alpha(theme.palette.info.main, 0.2)}`,
                                    color: theme.palette.info.dark,
                                    fontWeight: 500,
                                    alignItems: 'center',
                                    '& .MuiAlert-icon': { color: theme.palette.info.main }
                                }}
                            >
                                Zur Registrierung ist die @studmail.w-hs.de E-Mail erforderlich.
                            </Alert>
                        )}

                        <Stack spacing={2.5}>
                            {tabValue === 1 && (
                                <TextField
                                    required fullWidth id="name" label="Vollständiger Name" name="name"
                                    autoComplete="name" autoFocus disabled={loading || (tabValue === 1 && !!success)}
                                    error={!!errors.name} helperText={errors.name}
                                />
                            )}

                                <Stack direction="row" spacing={1} sx={{ width: '100%', position: 'relative', alignItems: 'flex-start' }}>
                                    <TextField
                                        required fullWidth id="emailPrefix" label="E-Mail" name="emailPrefix"
                                        autoComplete="email" autoFocus={tabValue === 0} disabled={loading || (tabValue === 1 && !!success)}
                                        error={!!errors.emailPrefix} helperText={errors.emailPrefix}
                                        sx={{ flexGrow: 1 }}
                                    />
                                    <TextField
                                        select required name="emailDomain" id="emailDomain"
                                        defaultValue={ALLOWED_DOMAINS[0]} disabled={loading || (tabValue === 1 && !!success)}
                                        sx={{ minWidth: 200 }}
                                    >
                                        {ALLOWED_DOMAINS.map((domain) => (
                                            <MenuItem key={domain} value={domain}>{domain}</MenuItem>
                                        ))}
                                    </TextField>
                                    <Box sx={{ 
                                        position: 'absolute', 
                                        left: '100%', 
                                        ml: 1, 
                                        height: 56, // Height of the input field
                                        display: 'flex', 
                                        alignItems: 'center' 
                                    }}>
                                        <Tooltip 
                                            title="Wir benötigen deine studentische E-Mail-Adresse, um deinen Immatrikulationsstatus zu verifizieren. Dies ist notwendig für den Zugriff auf Klausurprotokolle." 
                                            arrow 
                                            placement="right"
                                            slotProps={{
                                                tooltip: {
                                                    sx: {
                                                        fontSize: '0.875rem',
                                                        fontWeight: 500,
                                                        p: 1.5,
                                                        lineHeight: 1.4,
                                                        maxWidth: 300,
                                                        bgcolor: mode === 'dark' ? alpha(theme.palette.background.paper, 0.95) : alpha(theme.palette.text.primary, 0.9),
                                                        color: mode === 'dark' ? theme.palette.text.primary : theme.palette.background.paper,
                                                        border: mode === 'dark' ? `1px solid ${alpha(theme.palette.divider, 0.1)}` : 'none',
                                                        boxShadow: theme.shadows[4],
                                                        backdropFilter: 'blur(10px)'
                                                    }
                                                },
                                                arrow: {
                                                    sx: {
                                                        color: mode === 'dark' ? alpha(theme.palette.background.paper, 0.95) : alpha(theme.palette.text.primary, 0.9),
                                                        '&::before': {
                                                            border: mode === 'dark' ? `1px solid ${alpha(theme.palette.divider, 0.1)}` : 'none',
                                                        }
                                                    }
                                                }
                                            }}
                                        >
                                            <IconButton size="small" sx={{ color: 'text.secondary' }}>
                                                <HelpOutlineRounded fontSize="small" />
                                            </IconButton>
                                        </Tooltip>
                                    </Box>
                                </Stack>

                            {tabValue === 1 && (
                                <TextField
                                    select required fullWidth name="program_id" label="Studiengang" id="program_id"
                                    defaultValue="" disabled={loading || (tabValue === 1 && !!success)} error={!!errors.program_id} helperText={errors.program_id}
                                >
                                    {programs.length > 0 ? (
                                        programs.map((opt) => (
                                            <MenuItem key={String(opt.id)} value={String(opt.id)}>{opt.name}</MenuItem>
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
                                    disabled={loading || (tabValue === 1 && !!success)}
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
                                        disabled={loading || (tabValue === 1 && !!success)}
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
                                <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mt: -1 }}>
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
                                    <Button
                                        type="button"
                                        component={Link}
                                        href="/forgot"
                                        size="small"
                                        sx={{
                                            color: 'text.secondary',
                                            textTransform: 'none',
                                            fontWeight: 500,
                                            '&:hover': { color: 'primary.main', bgcolor: 'transparent', textDecoration: 'underline' }
                                        }}
                                    >
                                        Passwort vergessen?
                                    </Button>
                                </Stack>
                            )}
                        </Stack>

                        <Button
                            type="submit" fullWidth variant="contained" size="large"
                            disabled={loading || (tabValue === 1 && !!success)}
                            sx={{ mt: 4, mb: 2, height: 52, borderRadius: 2, fontWeight: 700 }}
                        >
                            {loading ? (tabValue === 0 ? 'Melde an...' : 'Erstelle Account...') : (tabValue === 0 ? 'Anmelden' : 'Registrieren')}
                        </Button>

                        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 1 }}>
                            <Button
                                startIcon={<ArrowBackRounded />}
                                component={Link}
                                href="/"
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
