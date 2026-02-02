import { useEffect, useState } from 'react';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import Stack from '@mui/material/Stack';
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import Page from '@components/Page';
import { useAuth, REMEMBERED_FLAG_KEY } from '@lib/auth';
import { getPrograms, putAuthMe } from '@lib/api';
import type { DtoProgramResponse as Program, DtoUserResponse as User } from '@lib/api';
import { zDtoUpdateProfileRequest } from '@lib/api/zod.gen';
import { useThemeMode, type ThemePreference } from '@lib/theme';
import { translateError } from '@lib/errors';

export default function Settings() {
  const { user, login } = useAuth();
  const { setPreference } = useThemeMode();
  const [progs, setProgs] = useState<Program[]>([]);
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState<{ t: 's' | 'e', m: string } | null>(null);

  const [form, setForm] = useState({
    name: user?.name || '',
    program_id: user?.program_id || '',
    theme: (user?.theme as ThemePreference) || 'system',
    private: user?.private === 1
  });

  useEffect(() => {
    getPrograms().then(({ data }) => data && setProgs(data));
  }, []);

  useEffect(() => {
    if (user) setForm({
      name: user.name || '',
      program_id: user.program_id || '',
      theme: (user.theme as ThemePreference) || 'system',
      private: user.private === 1
    });
  }, [user]);

  const save = async () => {
    const v = zDtoUpdateProfileRequest.safeParse({ ...form, program_id: form.program_id || undefined });
    if (!v.success) return setMsg({ t: 'e', m: 'Eingabe prüfen' });

    setLoading(true); setMsg(null);
    try {
      const res = await putAuthMe({ body: v.data });
      if (res.error) throw res.error;
      setMsg({ t: 's', m: 'Gespeichert' });
      if (res.data) {
        login(res.data as User, localStorage.getItem(REMEMBERED_FLAG_KEY) === 'true');
        setPreference(form.theme);
      }
    } catch (err) { setMsg({ t: 'e', m: translateError(err) }); }
    finally { setLoading(false); }
  };

  const isDirty = form.name !== (user?.name || '') || form.program_id !== (user?.program_id || '') || form.theme !== (user?.theme || 'system') || form.private !== (user?.private === 1);

  return (
    <Page title="Einstellungen">
      <Stack spacing={3}>
        <Paper variant="outlined" sx={{ p: 3, borderRadius: 3 }}>
          <Typography variant="h6" sx={{ mb: 2 }}>Profil</Typography>
          <Stack spacing={2}>
            <TextField label="E-Mail" value={user?.email || ''} disabled fullWidth />
            <TextField label="Name" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} fullWidth />
            <TextField select label="Studiengang" value={form.program_id} onChange={e => setForm({ ...form, program_id: e.target.value })} fullWidth>
              <MenuItem value=""><em>Keiner</em></MenuItem>
              {progs.map(p => <MenuItem key={p.id} value={p.id}>{p.name}</MenuItem>)}
            </TextField>
          </Stack>
        </Paper>

        <Paper variant="outlined" sx={{ p: 3, borderRadius: 3 }}>
          <Typography variant="h6" sx={{ mb: 2 }}>Privatsphäre & Anzeige</Typography>
          <Stack spacing={2}>
            <FormControlLabel control={<Switch checked={form.private} onChange={e => setForm({ ...form, private: e.target.checked })} />} label="Privates Profil" />
            <TextField select label="Design" value={form.theme} onChange={e => setForm({ ...form, theme: e.target.value as ThemePreference })} fullWidth>
              <MenuItem value="light">Hell</MenuItem><MenuItem value="dark">Dunkel</MenuItem><MenuItem value="system">System</MenuItem>
            </TextField>
          </Stack>
        </Paper>

        <Button variant="contained" onClick={save} disabled={loading || !isDirty || !form.name} sx={{ py: 1.5, fontWeight: 700 }}>
          {loading ? <CircularProgress size={24} /> : 'Speichern'}
        </Button>
      </Stack>

      <Snackbar open={!!msg} autoHideDuration={4000} onClose={() => setMsg(null)}>
        <Alert severity={msg?.t === 's' ? 'success' : 'error'} variant="filled" sx={{ borderRadius: 2 }}>{msg?.m}</Alert>
      </Snackbar>
    </Page>
  );
}
