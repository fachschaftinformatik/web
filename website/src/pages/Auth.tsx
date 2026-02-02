import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation, useSearchParams } from 'react-router-dom';
import { z } from 'zod';
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
import Fade from '@mui/material/Fade';
import { useTheme, alpha } from '@mui/material/styles';
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import Brightness4Rounded from '@mui/icons-material/Brightness4Rounded';
import Brightness7Rounded from '@mui/icons-material/Brightness7Rounded';
import PersonAddIcon from '@mui/icons-material/PersonAddRounded';
import LoginIcon from '@mui/icons-material/LoginRounded';

import { useThemeMode } from '@lib/theme';
import { postAuthLogin, postAuthRegister, getPrograms } from '@lib/api';
import type { DtoProgramResponse as Program } from '@lib/api';
import { useAuth, REMEMBERED_FLAG_KEY } from '@lib/auth';
import { translateError } from '@lib/errors';
import { zDtoLoginRequest, zDtoRegisterRequest } from '@lib/api/zod.gen';
import { fetchCsrfToken } from '@lib/csrf';
import Back from '@components/Back';

const DOMAINS = ['@studmail.w-hs.de', '@fsv-wh.de'];

const loginSchema = zDtoLoginRequest.omit({ email: true }).extend({
  prefix: z.string().min(1, "Kürzel fehlt"),
  domain: z.string(),
  password: z.string().min(1, "Passwort fehlt"),
});

const registerSchema = zDtoRegisterRequest.omit({ email: true, program_id: true }).extend({
  name: z.string().min(1, "Name fehlt"),
  prefix: z.string().min(1, "E-Mail fehlt").regex(/^[a-zA-Z0-9._-]+$/, "Ungültige Zeichen"),
  domain: z.string(),
  password: z.string().min(8, "Min. 8 Zeichen").regex(/[A-Z]/).regex(/[a-z]/).regex(/[0-9]/).regex(/[^A-Za-z0-9]/),
  confirm: z.string(),
  program_id: z.string().min(1, "Studiengang fehlt"),
}).refine(d => d.password === d.confirm, { message: "ungleich", path: ["confirm"] });

export default function Auth() {
  const navigate = useNavigate();
  const location = useLocation();
  const [params] = useSearchParams();
  const theme = useTheme();
  const { login } = useAuth();
  const { mode, setPreference } = useThemeMode();

  const [tab, setTab] = useState(location.pathname === '/register' ? 1 : 0);
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState('');
  const [errs, setErrors] = useState<Record<string, string>>({});
  const [showPwd, setShowPwd] = useState(false);
  const [programs, setPrograms] = useState<Program[]>([]);
  const [enabled, setEnabled] = useState(true);
  const [rem, setRem] = useState(() => localStorage.getItem(REMEMBERED_FLAG_KEY) === 'true');

  useEffect(() => {
    fetchCsrfToken().then(d => {
      if (d) {
        setEnabled(d.signups_enabled ?? true);
        if (!d.signups_enabled && tab === 1) { setTab(0); navigate('/login'); }
      }
    });
  }, [navigate, tab]);

  useEffect(() => {
    if (tab === 1 && !programs.length) {
      getPrograms().then(({ data }) => data && setPrograms(data));
    }
  }, [tab, programs.length]);

  useEffect(() => {
    if (params.get('verified') === 'true') {
      setSuccess('E-Mail bestätigt! Logge dich jetzt ein.');
      setTab(0);
    } else if (params.get('verified') === 'false') {
      setErrors({ global: translateError({ error: params.get('error') }) });
      setTab(0);
    }
  }, [params]);

  const submit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true); setErrors({}); setSuccess('');
    const fd = new FormData(e.currentTarget);
    const data = Object.fromEntries(fd.entries());

    try {
      if (tab === 0) {
        const v = loginSchema.parse(data);
        const { data: u, error } = await postAuthLogin({ body: { email: v.prefix + v.domain, password: v.password, remember: rem } });
        if (error) throw error;
        if (u) { login(u, rem); navigate(`/u/${u.id}`); }
      } else {
        const v = registerSchema.parse(data);
        const { data: u, error } = await postAuthRegister({ body: { email: v.prefix + v.domain, name: v.name, password: v.password, program_id: v.program_id } });
        if (error) throw error;
        if (u) setSuccess('Account erstellt! Bitte E-Mail bestätigen.');
      }
    } catch (err) {
      if (err instanceof z.ZodError) {
        const map: Record<string, string> = {};
        err.issues.forEach(i => { if (typeof i.path[0] === 'string') map[i.path[0]] = i.message; });
        setErrors(map);
      } else {
        setErrors({ global: translateError(err) });
      }
    } finally { setLoading(false); }
  };

  return (
    <Container component="main" maxWidth="xs">
      <CssBaseline />
      <Box sx={{ mt: 8, display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
        <Avatar sx={{ m: 1, bgcolor: 'primary.main', width: 56, height: 56, boxShadow: 3 }}>
          {tab === 0 ? <LoginIcon fontSize="large" /> : <PersonAddIcon fontSize="large" />}
        </Avatar>
        <Typography variant="h4" fontWeight={800} sx={{ mt: 2, mb: 3 }}>
          {tab === 0 ? 'Willkommen' : 'Registrieren'}
        </Typography>

        <Tabs value={tab} onChange={(_, v) => { setTab(v); setErrors({}); setSuccess(''); navigate(v === 1 ? '/register' : '/login'); }} variant="fullWidth" sx={{ width: '100%', mb: 4, bgcolor: alpha(theme.palette.primary.main, 0.05), borderRadius: 2 }}>
          <Tab label="Anmelden" sx={{ fontWeight: 700 }} />
          {enabled && <Tab label="Registrieren" sx={{ fontWeight: 700 }} />}
        </Tabs>

        <Fade in key={tab}>
          <Box component="form" noValidate onSubmit={submit} sx={{ width: '100%' }}>
            {success && <Alert severity="success" sx={{ mb: 3, borderRadius: 3 }}>{success}</Alert>}
            {errs.global && <Alert severity="error" sx={{ mb: 3, borderRadius: 3 }}>{errs.global}</Alert>}

            <Stack spacing={2}>
              {tab === 1 && <TextField required fullWidth label="Name" name="name" error={!!errs.name} helperText={errs.name} disabled={loading || !!success} />}
              <Stack direction="row" spacing={1}>
                <TextField required fullWidth label="E-Mail" name="prefix" error={!!errs.prefix} helperText={errs.prefix} disabled={loading || !!success} />
                <TextField select required name="domain" defaultValue={DOMAINS[0]} sx={{ minWidth: 150 }} disabled={loading || !!success}>
                  {DOMAINS.map(d => <MenuItem key={d} value={d}>{d}</MenuItem>)}
                </TextField>
              </Stack>
              {tab === 1 && (
                <TextField select required fullWidth label="Studiengang" name="program_id" defaultValue="" error={!!errs.program_id} helperText={errs.program_id} disabled={loading || !!success}>
                  {programs.map(p => <MenuItem key={String(p.id)} value={String(p.id)}>{p.name}</MenuItem>)}
                </TextField>
              )}
              <FormControl variant="outlined" required error={!!errs.password} fullWidth>
                <InputLabel>Passwort</InputLabel>
                <OutlinedInput name="password" type={showPwd ? 'text' : 'password'} disabled={loading || !!success}
                  endAdornment={<InputAdornment position="end"><IconButton onClick={() => setShowPwd(!showPwd)} edge="end">{showPwd ? <VisibilityOff /> : <Visibility />}</IconButton></InputAdornment>} label="Passwort" />
                {errs.password && <FormHelperText>{errs.password}</FormHelperText>}
              </FormControl>
              {tab === 1 && (
                <FormControl variant="outlined" required error={!!errs.confirm} fullWidth>
                  <InputLabel>Bestätigen</InputLabel>
                  <OutlinedInput name="confirm" type={showPwd ? 'text' : 'password'} disabled={loading || !!success} label="Bestätigen" />
                  {errs.confirm && <FormHelperText>{errs.confirm}</FormHelperText>}
                </FormControl>
              )}
              {tab === 0 && <FormControlLabel control={<Checkbox checked={rem} onChange={e => setRem(e.target.checked)} color="primary" />} label="Angemeldet bleiben" />}
            </Stack>

            <Button type="submit" fullWidth variant="contained" size="large" disabled={loading || !!success} sx={{ mt: 4, mb: 2, height: 52, borderRadius: 2, fontWeight: 700 }}>
              {loading ? '...' : (tab === 0 ? 'Anmelden' : 'Registrieren')}
            </Button>
            <Box sx={{ display: 'flex', justifyContent: 'center', mt: 1 }}>
              <Back to="/" label="Zurück zur Startseite" />
            </Box>
            <Box sx={{ display: 'flex', justifyContent: 'center', mt: 3 }}>
              <Tooltip title="Design">
                <IconButton onClick={() => setPreference(mode === 'light' ? 'dark' : 'light')} sx={{ border: '1px solid ' + alpha(theme.palette.primary.main, 0.2), p: 1.5 }}>
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
