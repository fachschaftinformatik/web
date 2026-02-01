import React, { useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import Alert from '@mui/material/Alert';
import { getCsrfFromCookie } from '@lib/csrf';

export default function ResetPage() {
    const [searchParams] = useSearchParams();
    const token = searchParams.get('token') || '';
    const [password, setPassword] = useState('');
    const [confirm, setConfirm] = useState('');
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState(false);
    const navigate = useNavigate();

    const submit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        if (!token) { setError('Ungültiger Link'); return; }
        if (password.length < 8) { setError('Passwort zu kurz'); return; }
        if (password !== confirm) { setError('Passwörter stimmen nicht überein'); return; }
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
                setTimeout(() => navigate('/login'), 1500);
            } else {
                const data = await res.json();
                setError(data?.message || 'Fehler');
            }
        } catch (err) {
            console.error(err)
            setError('Netzwerkfehler');
        }
    };

    return (
        <Box component="form" onSubmit={submit} sx={{ maxWidth: 480, mx: 'auto', mt: 6 }}>
            {success && <Alert severity="success">Passwort erfolgreich zurückgesetzt. Du wirst weitergeleitet.</Alert>}
            {!success && (
                <>
                    {error && <Alert severity="error">{error}</Alert>}
                    <TextField label="Neues Passwort" type="password" fullWidth required value={password} onChange={(e) => setPassword(e.target.value)} sx={{ my: 2 }} />
                    <TextField label="Passwort bestätigen" type="password" fullWidth required value={confirm} onChange={(e) => setConfirm(e.target.value)} sx={{ my: 2 }} />
                    <Button type="submit" variant="contained">Passwort zurücksetzen</Button>
                </>
            )}
        </Box>
    );
}
