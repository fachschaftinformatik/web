import { useState, useEffect, useCallback } from "react";
import { useLocation, useNavigate, useParams, useSearchParams } from "react-router-dom";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Paper from "@mui/material/Paper";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import IconButton from "@mui/material/IconButton";
import TextField from "@mui/material/TextField";
import MenuItem from "@mui/material/MenuItem";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import ListItemButton from "@mui/material/ListItemButton";
import Skeleton from "@mui/material/Skeleton";
import useMediaQuery from "@mui/material/useMediaQuery";
import { useTheme } from "@mui/material/styles";
import CloseIcon from "@mui/icons-material/CloseRounded";
import EditIcon from "@mui/icons-material/EditRounded";
import SaveIcon from "@mui/icons-material/SaveRounded";
import DelIcon from "@mui/icons-material/DeleteRounded";
import PdfIcon from "@mui/icons-material/PictureAsPdfRounded";

import Page from "@components/Page";
import Back from "@components/Back";
import { useAuth } from "@lib/auth";
import { getArchive, getArchiveFile, getPrograms, getArchiveVersions, putArchiveId, deleteArchiveId, deleteArchiveFile } from "@lib/api";
import type { DtoArchiveEntryResponse as Exam, DtoProgramResponse as Prog } from "@lib/api";

export default function ArchiveDetail() {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const { user } = useAuth();
  const { moduleId, examId } = useParams();
  const [params] = useSearchParams();
  const loc = useLocation();
  const navigate = useNavigate();
  const canEdit = user?.role === "admin" || user?.role === "editor";

  const name = (loc.state as { modul?: string })?.modul || params.get("mod") || "Modul";
  const [exams, setExams] = useState<Exam[]>([]);
  const [sel, setSel] = useState<Exam | null>(null);
  const [url, setUrl] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [editing, setEditing] = useState(false);
  const [vers, setVers] = useState<Exam[]>([]);
  const [delOpen, setDelOpen] = useState(false);

  const [form, setForm] = useState({ date: '', comment: '', pid: '', mid: '', po: '' });
  const [progs, setProgs] = useState<Prog[]>([]);

  const fetch = useCallback(() => {
    void Promise.resolve().then(() => setLoading(true));
    getArchive({ query: { module_id: moduleId } }).then(({ data }) => {
      const sorted = (data || []).sort((a, b) => new Date(b.exam_date || 0).getTime() - new Date(a.exam_date || 0).getTime());
      setExams(sorted);
      if (examId) setSel(sorted.find(e => String(e.id) === examId) || null);
      setLoading(false);
    });
  }, [moduleId, examId]);

  useEffect(() => {
    fetch();
    getPrograms().then(({ data }) => data && setProgs(data));
  }, [fetch]);

  useEffect(() => {
    if (!sel) return;
    void Promise.resolve().then(() => {
      setForm({ date: sel.exam_date || '', comment: sel.comment || '', pid: String(sel.program_id), mid: String(sel.module_id), po: sel.version || '' });
    });
    getArchiveVersions({ path: { entryId: String(sel.id) } }).then(({ data }) => setVers(data || []));
    getArchiveFile({ path: { entryId: String(sel.id) }, query: { file_id: String(sel.file_id) } }).then(({ data }) => {
      if (data) {
        const b = data instanceof Blob ? data : new Blob([data as BlobPart], { type: 'application/pdf' });
        const u = URL.createObjectURL(b);
        setUrl(old => { if (old) URL.revokeObjectURL(old); return u; });
      }
    });
  }, [sel]);

  const save = async () => {
    if (!sel) return;
    await putArchiveId({ path: { entryId: String(sel.id) }, body: { module_id: form.mid, version: form.po, date: form.date, comment: form.comment } });
    setEditing(false); fetch();
  };

  const del = async () => {
    if (!sel) return;
    if (vers.length > 1) await deleteArchiveFile({ path: { fileId: String(sel.file_id) } });
    else await deleteArchiveId({ path: { entryId: String(sel.id) } });
    setDelOpen(false); setSel(null); fetch();
    if (vers.length <= 1) navigate(`/archive/${moduleId}`);
  };

  return (
    <Page title="Archiv" hideHeader>
      <Back to="/archive" />
      <Typography variant="h4" fontWeight={700} gutterBottom>{name}</Typography>

      {loading ? <Skeleton variant="rectangular" height={200} /> : 
        <Paper variant="outlined" sx={{ borderRadius: 3, overflow: 'hidden' }}>
          <List disablePadding>
            {exams.map((e, i) => (
              <ListItem key={String(e.id)} divider={i < exams.length - 1} disablePadding>
                <ListItemButton onClick={() => navigate(`/archive/${moduleId}/${e.id}`)} selected={String(e.id) === examId}>
                  <ListItemText primary={`${e.version} (${new Date(e.exam_date || '').toLocaleDateString('de-DE')})`} secondary={e.comment} />
                </ListItemButton>
              </ListItem>
            ))}
          </List>
        </Paper>
      }

      <Dialog open={!!sel} onClose={() => navigate(`/archive/${moduleId}`)} fullWidth maxWidth="sm">
        <DialogTitle sx={{ display: 'flex', justifyContent: 'space-between' }}>
          {editing ? 'Bearbeiten' : 'Details'}
          <Box>
            {canEdit && !editing && <IconButton onClick={() => setEditing(true)}><EditIcon /></IconButton>}
            {editing && <><IconButton onClick={save}><SaveIcon /></IconButton><IconButton onClick={() => setDelOpen(true)}><DelIcon /></IconButton></>}
            <IconButton onClick={() => navigate(`/archive/${moduleId}`)}><CloseIcon /></IconButton>
          </Box>
        </DialogTitle>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
          {editing ? <>
            <TextField type="date" label="Datum" fullWidth value={form.date} onChange={e => setForm({ ...form, date: e.target.value })} size="small" slotProps={{ inputLabel: { shrink: true } }} />
            <TextField select label="PO" fullWidth value={form.po} onChange={e => setForm({ ...form, po: e.target.value })} size="small">
              {progs.find(p => String(p.id) === form.pid)?.versions?.map(v => <MenuItem key={v} value={v}>{v}</MenuItem>)}
            </TextField>
            <TextField label="Kommentar" multiline rows={2} fullWidth value={form.comment} onChange={e => setForm({ ...form, comment: e.target.value })} size="small" />
          </> : <>
            {url ? (isMobile ? <Button fullWidth variant="outlined" startIcon={<PdfIcon />} onClick={() => window.open(url, '_blank')}>PDF öffnen</Button> : 
              <iframe src={url} width="100%" height="500px" style={{ border: 'none' }} title="Vorschau" />) : <CircularProgress />}
          </>}
        </DialogContent>
      </Dialog>

      <Dialog open={delOpen} onClose={() => setDelOpen(false)}>
        <DialogTitle>Löschen?</DialogTitle>
        <DialogActions><Button onClick={() => setDelOpen(false)}>Abbrechen</Button><Button onClick={del} color="error">Löschen</Button></DialogActions>
      </Dialog>
    </Page>
  );
}
