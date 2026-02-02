import { useState, useEffect, useMemo } from "react";
import { Link as RouterLink } from "react-router-dom";
import MenuItem from "@mui/material/MenuItem";
import TextField from "@mui/material/TextField";
import InputAdornment from "@mui/material/InputAdornment";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import ListItemButton from "@mui/material/ListItemButton";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import Button from "@mui/material/Button";
import IconButton from "@mui/material/IconButton";
import Autocomplete from "@mui/material/Autocomplete";
import CircularProgress from "@mui/material/CircularProgress";
import Skeleton from "@mui/material/Skeleton";
import SearchIcon from "@mui/icons-material/SearchRounded";
import UpBtnIcon from "@mui/icons-material/UploadRounded";

import { useAuth } from "@lib/auth";
import Page from "@components/Page";
import { getPrograms, getProgramModules, postArchive, getArchive } from "@lib/api";
import type { DtoProgramResponse as Prog, DtoModuleResponse as Mod } from "@lib/api";

export default function Archive() {
  const { user } = useAuth();
  const canUp = user?.role === "admin" || user?.role === "editor";

  const [upOpen, setUpOpen] = useState(false);
  const [progs, setProgs] = useState<Prog[]>([]);
  const [selProgs, setSelProgs] = useState<Prog[]>([]);
  const [selPo, setSelPo] = useState("all");
  const [mods, setMods] = useState<Mod[]>([]);
  const [search, setSearch] = useState("");
  const [activeIds, setActiveIds] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);
  const [refresh, setRefresh] = useState(0);

  useEffect(() => {
    getPrograms().then(({ data }) => {
      if (data) {
        setProgs(data);
        const def = data.find(p => String(p.id) === String(user?.program_id)) || data[0];
        if (def) setSelProgs([def]);
      }
    });
  }, [user?.program_id]);

  useEffect(() => {
    const p = selProgs.length ? selProgs : progs;
    if (!p.length) return;
    Promise.all(p.map(x => getProgramModules({ path: { programId: String(x.id) } }))).then(res => {
      const all = res.flatMap(r => r.data || []);
      setMods(Array.from(new Map(all.map(m => [String(m.id), m])).values()));
    });
  }, [selProgs, progs]);

  useEffect(() => {
    void Promise.resolve().then(() => setLoading(true));
    const query = selProgs.length === 1 ? { program_id: String(selProgs[0].id), version: selPo === 'all' ? undefined : selPo } : {};
    getArchive({ query }).then(({ data }) => {
      setActiveIds(new Set((data || []).map(e => String(e.module_id))));
      setLoading(false);
    });
  }, [selProgs, selPo, refresh]);

  const filtered = useMemo(() => mods.filter(m => activeIds.has(String(m.id)) && (!search || m.name?.toLowerCase().includes(search.toLowerCase()))), [mods, activeIds, search]);
  const pos = useMemo(() => Array.from(new Set(selProgs.flatMap(p => p.versions || []))).sort().reverse(), [selProgs]);

  return (
    <Page title="Klausuren" description="Zur Vorbereitung auf Eure Prüfungen stellen wir Euch eine Ansammlung von Klausurrekonstruktionen zur Verfügung.">
      <Paper variant="outlined" sx={{ p: 3, mb: 4 }}>
        <Stack spacing={2}>
          <Stack direction="row" spacing={2}>
            <TextField fullWidth placeholder="Modul suchen..." value={search} onChange={e => setSearch(e.target.value)} size="small" slotProps={{ input: { startAdornment: <InputAdornment position="start"><SearchIcon color="primary" /></InputAdornment> } }} />
            {canUp && <IconButton onClick={() => setUpOpen(true)} sx={{ bgcolor: 'primary.main', color: '#fff', borderRadius: 2, '&:hover': { bgcolor: 'primary.dark' } }}><UpBtnIcon /></IconButton>}
          </Stack>
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
            <Autocomplete multiple options={progs} getOptionLabel={o => o.name || ''} value={selProgs} onChange={(_, v) => setSelProgs(v)} renderInput={p => <TextField {...p} label="Studiengang" size="small" />} sx={{ flex: 2 }} />
            <TextField select label="PO" value={selPo} onChange={e => setSelPo(e.target.value)} size="small" sx={{ flex: 1 }}>
              <MenuItem value="all">Alle</MenuItem>
              {pos.map(v => <MenuItem key={v} value={v}>{v}</MenuItem>)}
            </TextField>
          </Stack>
        </Stack>
      </Paper>

      {loading ? <Skeleton variant="rectangular" height={300} /> : 
        <List sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 2, p: 0 }}>
          {filtered.map((m, i) => (
            <ListItem key={String(m.id)} divider={i < filtered.length - 1} disablePadding>
              <ListItemButton component={RouterLink} to={`/archive/${m.id}`} state={{ modul: m.name }}>
                <ListItemText primary={m.name} />
              </ListItemButton>
            </ListItem>
          ))}
        </List>
      }
      <UploadDialog open={upOpen} onClose={() => setUpOpen(false)} progs={progs} onDone={() => setRefresh(r => r + 1)} />
    </Page>
  );
}

function UploadDialog({ open, onClose, progs, onDone }: { open: boolean, onClose: () => void, progs: Prog[], onDone: () => void }) {
  const [f, setF] = useState<File | null>(null);
  const [d, setD] = useState("");
  const [pid, setPid] = useState("");
  const [mid, setMid] = useState("");
  const [ver, setVer] = useState("");
  const [ms, setMs] = useState<Mod[]>([]);
  const [l, setL] = useState(false);

  useEffect(() => {
    if (pid) getProgramModules({ path: { programId: pid } }).then(({ data }) => setMs(data || []));
  }, [pid]);

  const sub = async () => {
    if (!f || !d || !mid || !ver) return;
    setL(true);
    const { error } = await postArchive({ body: { file: f, date: d, module_id: mid, version: ver } } as never);
    setL(false);
    if (!error) { onDone(); onClose(); setF(null); } else alert("Fehler");
  };

  const selP = progs.find((p) => String(p.id) === pid);

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>Hochladen</DialogTitle>
      <DialogContent><Stack spacing={2} sx={{ mt: 1 }}>
        <TextField type="date" label="Datum" slotProps={{ inputLabel: { shrink: true } }} fullWidth value={d} onChange={e => setD(e.target.value)} />
        <TextField select label="Studiengang" fullWidth value={pid} onChange={e => setPid(e.target.value)}>
          {progs.map((p) => <MenuItem key={p.id} value={String(p.id)}>{p.name}</MenuItem>)}
        </TextField>
        <TextField select label="PO" fullWidth value={ver} onChange={e => setVer(e.target.value)} disabled={!pid}>
          {selP?.versions?.map((v: string) => <MenuItem key={v} value={v}>{v}</MenuItem>)}
        </TextField>
        <Autocomplete options={ms} getOptionLabel={o => o.name || ''} onChange={(_, v) => setMid(String(v?.id || ''))} renderInput={p => <TextField {...p} label="Modul" />} disabled={!pid} />
        <input type="file" onChange={e => setF(e.target.files?.[0] || null)} />
      </Stack></DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Abbrechen</Button>
        <Button variant="contained" onClick={sub} disabled={l}>{l ? <CircularProgress size={20} /> : "Upload"}</Button>
      </DialogActions>
    </Dialog>
  );
}
