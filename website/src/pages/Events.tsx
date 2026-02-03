import { useState, useEffect, useCallback } from "react";
import { useNavigate, useParams } from "react-router-dom";
import Typography from "@mui/material/Typography";
import Card from "@mui/material/Card";
import CardMedia from "@mui/material/CardMedia";
import Box from "@mui/material/Box";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import IconButton from "@mui/material/IconButton";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import MenuItem from "@mui/material/MenuItem";
import Snackbar from "@mui/material/Snackbar";
import Alert from "@mui/material/Alert";
import CircularProgress from "@mui/material/CircularProgress";
import Paper from "@mui/material/Paper";
import Skeleton from "@mui/material/Skeleton";
import CloseIcon from "@mui/icons-material/Close";
import PrevIcon from "@mui/icons-material/ArrowBackIosNew";
import NextIcon from "@mui/icons-material/ArrowForwardIos";
import UploadBtnIcon from "@mui/icons-material/UploadRounded";
import AddIcon from "@mui/icons-material/AddRounded";

import { useAuth } from "@lib/auth";
import Page from "@components/Page";
import Pagination from "@components/Pagination";
import { getEvents, getEventsByEventIdMedia, postEvents, postEventsByEventIdMedia } from "@lib/api";
import type { DtoEventResponse as Ev, DtoMediaResponse as Med } from "@lib/api";
import { getSizedImageUrl } from "@lib/images";

const PAGE_SIZE = 10;

export default function Events() {
  const { user } = useAuth();
  const { eventId: urlEid, mediaId: urlMid } = useParams();
  const navigate = useNavigate();
  const canUp = user?.role === "admin" || user?.role === "editor";

  const [evs, setEvs] = useState<Ev[]>([]);
  const [med, setMed] = useState<Med[]>([]);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [upOpen, setUpOpen] = useState(false);
  const [crOpen, setCrOpen] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);
  const [lbOpen, setLbOpen] = useState(false);
  const [lbIdx, setLbIdx] = useState(0);

  const fetchEvs = useCallback(async () => {
    const { data } = await getEvents();
    if (data) {
      setEvs(data);
      if (!urlEid && data[0]) navigate(`/events/${data[0].id}`, { replace: true });
    }
  }, [urlEid, navigate]);

  useEffect(() => {
    void Promise.resolve().then(() => fetchEvs());
  }, [fetchEvs]);

  useEffect(() => {
    if (!urlEid) return;
    void Promise.resolve().then(() => setLoading(true));
    getEventsByEventIdMedia({ path: { eventId: urlEid } }).then(({ data }) => {
      setMed(data || []);
      setLoading(false);
    });
  }, [urlEid]);

  useEffect(() => {
    if (urlMid && med.length) {
      const idx = med.findIndex(m => String(m.id) === urlMid);
      if (idx !== -1) {
        void Promise.resolve().then(() => {
          setLbIdx(idx);
          setLbOpen(true);
        });
      }
    } else void Promise.resolve().then(() => setLbOpen(false));
  }, [urlMid, med]);

  const setLb = (i: number) => {
    const idx = (i + med.length) % med.length;
    setLbIdx(idx);
    navigate(`/events/${urlEid}/${med[idx].id}`, { replace: true });
  };

  return (
    <Page title="Galerie" description="Fotos der Fachschaft.">
      <Box sx={{ mb: 2, display: 'flex', justifyContent: 'flex-end' }}>
        {canUp && <Button variant="contained" startIcon={<UploadBtnIcon />} onClick={() => setUpOpen(true)}>Upload</Button>}
      </Box>

      <Paper variant="outlined" sx={{ p: 2, mb: 4, overflowX: 'auto' }}>
        <Stack direction="row" spacing={2}>
          {evs.length ? evs.map(e => (
            <Card key={String(e.id)} onClick={() => navigate(`/events/${e.id}`)} sx={{ 
              minWidth: 160, cursor: 'pointer', border: urlEid === String(e.id) ? '2px solid' : 'none', 
              borderColor: 'primary.main', transform: urlEid === String(e.id) ? 'scale(1.05)' : 'none' 
            }}>
              <CardMedia component="img" height="100" image={getSizedImageUrl(e.cover_path || '', 400)} />
              <Typography variant="caption" sx={{ p: 1, display: 'block' }} noWrap>{e.title}</Typography>
            </Card>
          )) : [1,2,3].map(i => <Skeleton key={i} variant="rectangular" width={160} height={120} />)}
        </Stack>
      </Paper>

      {loading ? <Skeleton variant="rectangular" height={400} /> : 
        <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: 2 }}>
          {med.slice((page-1)*PAGE_SIZE, page*PAGE_SIZE).map((m, i) => (
            <Card key={String(m.id)} onClick={() => setLb((page-1)*PAGE_SIZE + i)} sx={{ cursor: 'pointer' }}>
              <CardMedia component="img" image={getSizedImageUrl(`/api/v1/media/${m.id}/preview`, 600)} sx={{ aspectRatio: '3/2' }} />
            </Card>
          ))}
        </Box>
      }

      <Pagination count={Math.ceil(med.length / PAGE_SIZE)} page={page} onChange={(_, p) => setPage(p)} disabled={loading} />

      <UploadDialog open={upOpen} onClose={() => setUpOpen(false)} evs={evs} onDone={() => { fetchEvs(); setMsg("Upload erfolgreich"); }} onAddEv={() => setCrOpen(true)} />
      <CreateEvDialog open={crOpen} onClose={() => setCrOpen(false)} onDone={fetchEvs} />
      
      <Dialog open={lbOpen} onClose={() => navigate(`/events/${urlEid}`)} fullWidth maxWidth="lg">
        <DialogTitle sx={{ display: 'flex', justifyContent: 'space-between' }}>
          {med[lbIdx]?.title || 'Bild'}
          <IconButton onClick={() => navigate(`/events/${urlEid}`)}><CloseIcon /></IconButton>
        </DialogTitle>
        <DialogContent sx={{ textAlign: 'center', p: 0, bgcolor: '#000', position: 'relative' }}>
          {med[lbIdx] && <img src={getSizedImageUrl(`/api/v1/media/${med[lbIdx].id}/preview`, 1600)} style={{ maxWidth: '100%', maxHeight: '80vh' }} alt="" />}
          <IconButton onClick={() => setLb(lbIdx - 1)} sx={{ position: 'absolute', left: 8, top: '50%', color: '#fff' }}><PrevIcon /></IconButton>
          <IconButton onClick={() => setLb(lbIdx + 1)} sx={{ position: 'absolute', right: 8, top: '50%', color: '#fff' }}><NextIcon /></IconButton>
        </DialogContent>
      </Dialog>

      <Snackbar open={!!msg} autoHideDuration={3000} onClose={() => setMsg(null)}><Alert severity="success">{msg}</Alert></Snackbar>
    </Page>
  );
}

function UploadDialog({ open, onClose, evs, onDone, onAddEv }: { open: boolean, onClose: () => void, evs: Ev[], onDone: () => void, onAddEv: () => void }) {
  const [files, setFiles] = useState<File[]>([]);
  const [eid, setEid] = useState("");
  const [uping, setUping] = useState(false);

  const sub = async () => {
    setUping(true);
    for (const f of files) await postEventsByEventIdMedia({ path: { eventId: eid }, body: { file: f } } as never);
    setUping(false); onDone(); onClose(); setFiles([]);
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>Bilder hochladen</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ mt: 1 }}>
          <Stack direction="row" spacing={1}>
            <TextField select fullWidth label="Event" value={eid} onChange={e => setEid(e.target.value)} size="small">
              {evs.map((e) => <MenuItem key={e.id} value={String(e.id)}>{e.title}</MenuItem>)}
            </TextField>
            <Button onClick={onAddEv} variant="outlined"><AddIcon /></Button>
          </Stack>
          <input type="file" multiple onChange={e => setFiles(Array.from(e.target.files || []))} />
          {files.length > 0 && <Typography variant="caption">{files.length} {files.length === 1 ? 'Bild' : 'Bilder'} ausgewählt</Typography>}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Abbrechen</Button>
        <Button variant="contained" onClick={sub} disabled={!files.length || !eid || uping}>{uping ? <CircularProgress size={20} /> : "Upload"}</Button>
      </DialogActions>
    </Dialog>
  );
}

function CreateEvDialog({ open, onClose, onDone }: { open: boolean, onClose: () => void, onDone: () => void }) {
  const [t, setT] = useState("");
  const [f, setF] = useState<File | null>(null);
  const [l, setL] = useState(false);

  const sub = async () => {
    setL(true);
    await postEvents({ body: { title: t, file: f || undefined } } as never);
    setL(false); onDone(); onClose(); setT(""); setF(null);
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="xs">
      <DialogTitle>Neues Event</DialogTitle>
      <DialogContent><Stack spacing={2} sx={{ mt: 1 }}>
        <TextField fullWidth label="Titel" value={t} onChange={e => setT(e.target.value)} />
        <input type="file" onChange={e => setF(e.target.files?.[0] || null)} />
      </Stack></DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Abbrechen</Button>
        <Button variant="contained" onClick={sub} disabled={!t || l}>{l ? <CircularProgress size={20} /> : "Erstellen"}</Button>
      </DialogActions>
    </Dialog>
  );
}
