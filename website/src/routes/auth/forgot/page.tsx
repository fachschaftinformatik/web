import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import Alert from '@mui/material/Alert';
import { getCsrfFromCookie } from '@lib/csrf';

export default function ForgotPage() {
    const [email, setEmail] = useState('');
    const [status, setStatus] = useState<'idle'|'sent'|'error'>('idle');
    const [error, setError] = useState<string | null>(null);
    const navigate = useNavigate();

    const submit = async (e: React.FormEvent) => {
        e.preventDefault();
        setStatus('idle'); setError(null);
        try {
            const res = await fetch('/api/v1/auth/forgot', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': getCsrfFromCookie() || ''
                },
                body: JSON.stringify({ email })
            });
            if (res.status === 204) {
                setStatus('sent');
            } else {
                try {
                    const data = await res.json();
                    setError(data?.message || 'Fehler');
                } catch {
                    setError(res.statusText || 'Fehler');
                }
                setStatus('error');
            }
        } catch (err) {
            console.error('Forgot password error:', err);
            setError('Netzwerkfehler');
            setStatus('error');
        }
    };

    return (
        <Box component="form" onSubmit={submit} sx={{ maxWidth: 480, mx: 'auto', mt: 6 }}>
            {status === 'sent' && <Alert severity="success">Falls ein Account existiert, wurde eine E-Mail mit Anweisungen versendet.</Alert>}
            {status !== 'sent' && (
                <>
                    {error && <Alert severity="error">{error}</Alert>}
                    <TextField label="E-Mail" fullWidth required value={email} onChange={(e) => setEmail(e.target.value)} sx={{ my: 2 }} />
                    <Button type="submit" variant="contained">Link zum Zurücksetzen anfordern</Button>
                </>
            )}
            <Box sx={{ mt: 2 }}>
                <Button onClick={() => navigate('/login')}>Zurück zur Anmeldung</Button>
            </Box>
        </Box>
    );
}
