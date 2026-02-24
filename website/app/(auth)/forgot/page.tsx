'use client';

import React, { useState } from 'react';
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
import { useTheme, alpha } from '@mui/material/styles';
import LockResetRoundedIcon from '@mui/icons-material/LockResetRounded';
import ArrowBackRounded from '@mui/icons-material/ArrowBackRounded';
import CheckCircleOutlineRounded from '@mui/icons-material/CheckCircleOutlineRounded';
import ErrorOutlineRounded from '@mui/icons-material/ErrorOutlineRounded';
import MenuItem from '@mui/material/MenuItem';
import Link from 'next/link';

const ALLOWED_DOMAINS = ['@studmail.w-hs.de', '@fsv-wh.de'];

export default function ForgotPage() {
    const [emailPrefix, setEmailPrefix] = useState('');
    const [emailDomain, setEmailDomain] = useState(ALLOWED_DOMAINS[0]);
    const [loading, setLoading] = useState(false);
    const [status, setStatus] = useState<'idle'|'sent'|'error'>('idle');
    const [error, setError] = useState<string | null>(null);
    const theme = useTheme();

    const submit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setStatus('idle'); setError(null);
        try {
            const res = await fetch('/api/v1/auth/forgot', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ email: `${emailPrefix}${emailDomain}` })
            });
            if (res.status === 204) {
                setStatus('sent');
            } else {
                try {
                    const data = await res.json();
                    setError(data?.message || 'Fehler beim Senden der E-Mail.');
                } catch {
                    setError(res.statusText || 'Fehler beim Senden der E-Mail.');
                }
                setStatus('error');
            }
        } catch (err) {
            console.error('Forgot password error:', err);
            setError('Netzwerkfehler');
            setStatus('error');
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
                        <LockResetRoundedIcon fontSize="large" />
                    </Avatar>
                </Zoom>
                <Typography component="h1" variant="h4" fontWeight={800} sx={{ mt: 2, mb: 3, textAlign: 'center' }}>
                    Passwort vergessen
                </Typography>

                <Fade in={true}>
                    <Box component="form" onSubmit={submit} sx={{ width: '100%' }}>
                        {status === 'sent' && (
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
                                    Falls ein Account existiert, wurde eine E-Mail mit Anweisungen versendet.
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

                        {status !== 'sent' && (
                            <Stack spacing={2.5}>
                                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                                    Gib deine E-Mail-Adresse ein, um einen Link zum Zurücksetzen deines Passworts zu erhalten.
                                </Typography>
                                <Stack direction="row" spacing={1}>
                                    <TextField
                                        label="E-Mail"
                                        fullWidth
                                        required
                                        value={emailPrefix}
                                        onChange={(e) => setEmailPrefix(e.target.value)}
                                        disabled={loading}
                                        autoFocus
                                    />
                                    <TextField
                                        select
                                        required
                                        value={emailDomain}
                                        onChange={(e) => setEmailDomain(e.target.value)}
                                        disabled={loading}
                                        sx={{ minWidth: 200 }}
                                    >
                                        {ALLOWED_DOMAINS.map((domain) => (
                                            <MenuItem key={domain} value={domain}>{domain}</MenuItem>
                                        ))}
                                    </TextField>
                                </Stack>
                                <Button
                                    type="submit"
                                    variant="contained"
                                    fullWidth
                                    size="large"
                                    disabled={loading}
                                    sx={{ mt: 2, height: 52, borderRadius: 2, fontWeight: 700 }}
                                >
                                    {loading ? 'Wird gesendet...' : 'Link anfordern'}
                                </Button>
                            </Stack>
                        )}

                        <Box sx={{ display: 'flex', justifyContent: 'center', mt: status === 'sent' ? 0 : 2 }}>
                            <Button
                                startIcon={<ArrowBackRounded />}
                                component={Link}
                                href="/login"
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
