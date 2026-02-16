import React, { useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import Alert from '@mui/material/Alert';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Avatar from '@mui/material/Avatar';
import Stack from '@mui/material/Stack';
import Zoom from '@mui/material/Zoom';
import Fade from '@mui/material/Fade';
import CssBaseline from '@mui/material/CssBaseline';
import IconButton from '@mui/material/IconButton';
import OutlinedInput from '@mui/material/OutlinedInput';
import InputLabel from '@mui/material/InputLabel';
import InputAdornment from '@mui/material/InputAdornment';
import FormControl from '@mui/material/FormControl';
import { useTheme, alpha } from '@mui/material/styles';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import LockOpenRoundedIcon from '@mui/icons-material/LockOpenRounded';
import CheckCircleOutlineRounded from '@mui/icons-material/CheckCircleOutlineRounded';
import ErrorOutlineRounded from '@mui/icons-material/ErrorOutlineRounded';
import ArrowBackRounded from '@mui/icons-material/ArrowBackRounded';
import { getCsrfFromCookie } from '@lib/csrf';

export default function ResetPage() {
    const [searchParams] = useSearchParams();
    const token = searchParams.get('token') || '';
    const [password, setPassword] = useState('');
    const [confirm, setConfirm] = useState('');
    const [showPassword, setShowPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState(false);
    const navigate = useNavigate();
    const theme = useTheme();

    const submit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        if (!token) { setError('Ungültiger oder fehlender Token.'); return; }
        if (password.length < 8) { setError('Das Passwort muss mindestens 8 Zeichen lang sein.'); return; }
        if (password !== confirm) { setError('Die Passwörter stimmen nicht überein.'); return; }
        
        setLoading(true);
        try {
            const res = await fetch('/api/v1/auth/reset', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': getCsrfFromCookie() || ''
                },
                body: JSON.stringify({ token, password })
            });
            if (res.status === 200) {
                setSuccess(true);
                setTimeout(() => navigate('/login'), 2000);
            } else {
                try {
                    const data = await res.json();
                    setError(data?.message || 'Fehler beim Zurücksetzen des Passworts.');
                } catch {
                    setError(res.statusText || 'Fehler beim Zurücksetzen des Passworts.');
                }
            }
        } catch (err) {
            console.error(err)
            setError('Netzwerkfehler');
        } finally {
            setLoading(false);
        }
    };

    return (
        <Container component="main" maxWidth={false} sx={{ maxWidth: 480 }}>
            <CssBaseline />
            <Box sx={{ marginTop: 8, display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                <Zoom in={true} style={{ transitionDelay: '100ms' }}>
                    <Avatar sx={{ m: 1, bgcolor: 'primary.main', width: 56, height: 56, boxShadow: 3 }}>
                        <LockOpenRoundedIcon fontSize="large" />
                    </Avatar>
                </Zoom>
                <Typography component="h1" variant="h4" fontWeight={800} sx={{ mt: 2, mb: 3, textAlign: 'center' }}>
                    Passwort zurücksetzen
                </Typography>

                <Fade in={true}>
                    <Box component="form" onSubmit={submit} sx={{ width: '100%' }}>
                        {success && (
                            <Zoom in={true}>
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
                                    Passwort erfolgreich zurückgesetzt. Du wirst zur Anmeldung weitergeleitet.
                                </Alert>
                            </Zoom>
                        )}
                        
                        {error && (
                            <Zoom in={true}>
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
                                    {error}
                                </Alert>
                            </Zoom>
                        )}

                        {!success && (
                            <Stack spacing={2.5}>
                                <FormControl variant="outlined" required fullWidth error={!!error && error.includes('Passwort')}>
                                    <InputLabel htmlFor="reset-password">Neues Passwort</InputLabel>
                                    <OutlinedInput
                                        id="reset-password"
                                        type={showPassword ? 'text' : 'password'}
                                        value={password}
                                        onChange={(e) => setPassword(e.target.value)}
                                        disabled={loading}
                                        endAdornment={
                                            <InputAdornment position="end">
                                                <IconButton onClick={() => setShowPassword(!showPassword)} edge="end">
                                                    {showPassword ? <VisibilityOff /> : <Visibility />}
                                                </IconButton>
                                            </InputAdornment>
                                        }
                                        label="Neues Passwort"
                                    />
                                </FormControl>

                                <FormControl variant="outlined" required fullWidth error={!!error && error.includes('überein')}>
                                    <InputLabel htmlFor="confirm-password">Passwort bestätigen</InputLabel>
                                    <OutlinedInput
                                        id="confirm-password"
                                        type={showConfirmPassword ? 'text' : 'password'}
                                        value={confirm}
                                        onChange={(e) => setConfirm(e.target.value)}
                                        disabled={loading}
                                        endAdornment={
                                            <InputAdornment position="end">
                                                <IconButton onClick={() => setShowConfirmPassword(!showConfirmPassword)} edge="end">
                                                    {showConfirmPassword ? <VisibilityOff /> : <Visibility />}
                                                </IconButton>
                                            </InputAdornment>
                                        }
                                        label="Passwort bestätigen"
                                    />
                                </FormControl>

                                <Button
                                    type="submit"
                                    variant="contained"
                                    fullWidth
                                    size="large"
                                    disabled={loading}
                                    sx={{ mt: 2, height: 52, borderRadius: 2, fontWeight: 700 }}
                                >
                                    {loading ? 'Wird zurückgesetzt...' : 'Passwort speichern'}
                                </Button>
                            </Stack>
                        )}

                        <Box sx={{ display: 'flex', justifyContent: 'center', mt: success ? 0 : 2 }}>
                            <Button
                                startIcon={<ArrowBackRounded />}
                                onClick={() => navigate('/login')}
                                sx={{
                                    color: 'text.secondary',
                                    textTransform: 'none',
                                    '&:hover': { color: 'primary.main', bgcolor: 'transparent', textDecoration: 'underline' }
                                }}
                            >
                                Zurück zur Anmeldung
                            </Button>
                        </Box>
                    </Box>
                </Fade>
            </Box>
        </Container>
    );
}
