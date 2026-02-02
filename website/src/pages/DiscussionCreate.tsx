import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import MenuItem from "@mui/material/MenuItem";
import Autocomplete from "@mui/material/Autocomplete";
import Alert from "@mui/material/Alert";
import CircularProgress from "@mui/material/CircularProgress";

import { useAuth } from "@lib/auth";
import Page from "@components/Page";
import Back from "@components/Back";
import { getPrograms, postDiscussions } from "@lib/api";
import type { DtoProgramResponse as Prog } from "@lib/api";
import { FORUM_CATEGORIES, FORUM_TAGS } from "@internals/data";

export default function DiscussionCreate() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const isStaff = user?.role === "admin" || user?.role === "editor";

  const [progs, setProgs] = useState<Prog[]>([]);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState("");
  const [form, setForm] = useState({ title: '', body: '', cat: '', progs: [] as Prog[], tags: [] as string[], date: '', loc: '' });

  useEffect(() => { getPrograms().then(({ data }) => data && setProgs(data)); }, []);

  const sub = async () => {
    setLoading(true); setErr("");
    try {
      const type = form.cat === "Ankündigung" ? "news" : (form.cat === "Termin" ? "event" : "discussion");
      await postDiscussions({ body: { 
        title: form.title, body: form.body, type, 
        programs: form.progs.map(p => String(p.id)), 
        tags: [form.cat, ...form.tags],
        event_date: form.date || undefined, location: form.loc || undefined
      }});
      navigate("/discussions");
    } catch (e: unknown) { setErr((e as Error).message || "Fehler"); }
    finally { setLoading(false); }
  };

  return (
    <Page title="Beitrag erstellen" maxWidth="md">
      <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 4 }}>
        <Back to="/discussions" icon />
        <Typography variant="h4" fontWeight={700}>Neuer Beitrag</Typography>
      </Stack>

      {err && <Alert severity="error" sx={{ mb: 3 }}>{err}</Alert>}

      <Paper variant="outlined" sx={{ p: 4, borderRadius: 3 }}>
        <Stack spacing={3}>
          <TextField label="Titel" fullWidth value={form.title} onChange={e => setForm({ ...form, title: e.target.value })} />
          <TextField select label="Kategorie" fullWidth value={form.cat} onChange={e => setForm({ ...form, cat: e.target.value })}>
            {FORUM_CATEGORIES.filter(c => (c !== "Ankündigung" && c !== "Termin") || isStaff).map(c => <MenuItem key={c} value={c}>{c}</MenuItem>)}
          </TextField>
          <TextField label="Inhalt" multiline rows={6} fullWidth value={form.body} onChange={e => setForm({ ...form, body: e.target.value })} />
          
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
            <Autocomplete multiple options={progs} getOptionLabel={o => o.name || ""} onChange={(_, v) => setForm({ ...form, progs: v })} renderInput={p => <TextField {...p} label="Studiengänge" />} fullWidth />
            <Autocomplete multiple options={FORUM_TAGS} onChange={(_, v) => setForm({ ...form, tags: v })} renderInput={p => <TextField {...p} label="Tags" />} fullWidth />
          </Stack>

          {(form.cat === "Termin") && (
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
              <TextField label="Datum" type="datetime-local" fullWidth value={form.date} onChange={e => setForm({ ...form, date: e.target.value })} slotProps={{ inputLabel: { shrink: true } }} />
              <TextField label="Ort" fullWidth value={form.loc} onChange={e => setForm({ ...form, loc: e.target.value })} />
            </Stack>
          )}

          <Button variant="contained" onClick={sub} disabled={loading || !form.title || !form.body || !form.cat} sx={{ py: 1.5, fontWeight: 700 }}>
            {loading ? <CircularProgress size={24} /> : 'Veröffentlichen'}
          </Button>
        </Stack>
      </Paper>
    </Page>
  );
}
