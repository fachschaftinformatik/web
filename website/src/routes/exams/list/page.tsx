import { useMemo, useState, useEffect } from "react";
import {
  Box,
  Container,
  Typography,
  MenuItem,
  TextField,
  InputAdornment,
  Paper,
  Stack,
  List,
  ListItem,
  ListItemText,
  ListItemButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  IconButton,
  Autocomplete,
  Alert
} from "@mui/material";

import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import AddRoundedIcon from '@mui/icons-material/AddRounded';
import CloudUploadRoundedIcon from '@mui/icons-material/CloudUploadRounded';
import InsertDriveFileRoundedIcon from '@mui/icons-material/InsertDriveFileRounded';
import CloseRoundedIcon from '@mui/icons-material/CloseRounded';
import { Link as RouterLink } from "react-router-dom";

import { useAuth } from "@lib/auth";
import { Sidebar } from "@components/layout";
import { getPrograms, getProgramModules, postExams, getAuthCsrf } from "@lib/api";
import type { Program, Module } from "@lib/api";

function UploadDialog({ open, onClose, programs, onSuccess }: { open: boolean; onClose: () => void; programs: Program[]; onSuccess: () => void }) {
  const [uploadProgram, setUploadProgram] = useState<Program | null>(null);
  const [uploadPo, setUploadPo] = useState("");
  const [uploadModule, setUploadModule] = useState<Module | null>(null);
  const [programModules, setProgramModules] = useState<Module[]>([]);
  const [file, setFile] = useState<File | null>(null);
  const [examDate, setExamDate] = useState("");
  const [comment, setComment] = useState("");
  const [dragging, setDragging] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (uploadProgram) {
      getProgramModules({ path: { id: uploadProgram.id } })
        .then(({ data }) => setProgramModules(data || []))
        .catch(() => setProgramModules([]));
      setUploadPo(uploadProgram.versions[0] || "");
      setUploadModule(null);
    } else {
      setProgramModules([]);
    }
  }, [uploadProgram]);

  const handleProgramChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const prog = programs.find(p => p.id === Number(e.target.value));
    setUploadProgram(prog || null);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) setFile(e.target.files[0]);
  };

  const handleDrag = (e: React.DragEvent) => { e.preventDefault(); e.stopPropagation(); };
  const handleDragIn = (e: React.DragEvent) => { e.preventDefault(); e.stopPropagation(); setDragging(true); };
  const handleDragOut = (e: React.DragEvent) => { e.preventDefault(); e.stopPropagation(); setDragging(false); };
  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault(); e.stopPropagation(); setDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
        if(e.dataTransfer.files[0].type === "application/pdf") setFile(e.dataTransfer.files[0]);
    }
  };

  const handleSubmit = async () => {
    if (!uploadProgram || !uploadModule || !file || !examDate || !uploadPo) return;
    setLoading(true);
    setError("");

    try {
      const { data: csrfData, error: csrfError } = await getAuthCsrf();
      
      const token = csrfData?.csrf;
      if (csrfError || !token) {
        throw new Error("Sicherheits-Token konnte nicht geladen werden. Bitte neu einloggen.");
      }

      const formData = {
        file: file, 
        programid: uploadProgram.id,
        version: uploadPo,
        moduleid: uploadModule.id,
        date: examDate,
        comment: comment
      };

      const { error: apiError } = await postExams({
        // @ts-ignore: formData contains File, validation disabled via config to allow this
        body: formData,
        headers: {
            "X-CSRF-Token": token
        }
      });

      if (apiError) {
        // @ts-ignore
        const msg = apiError.message || "Fehler beim Hochladen.";
        setError(msg);
      } else {
        onSuccess();
        onClose();
        setFile(null);
        setUploadModule(null);
        setComment("");
        setExamDate("");
      }
    } catch (err) {
      console.error("Upload error:", err);
      setError(err instanceof Error ? err.message : "Netzwerkfehler.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog 
        open={open} 
        onClose={onClose} 
        fullWidth 
        maxWidth="sm" 
        PaperProps={{ sx: { borderRadius: 3 } }}
    >
      <DialogTitle sx={{ m: 0, p: 3, pb: 1, fontWeight: 700, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        Klausur hochladen
        <IconButton onClick={onClose}><CloseRoundedIcon /></IconButton>
      </DialogTitle>
      
      <DialogContent sx={{ p: 3 }}>
        <Stack spacing={3} sx={{ mt: 1 }}>
          {error && <Alert severity="error">{error}</Alert>}
          
          <TextField 
            select 
            label="Studiengang" 
            fullWidth 
            value={uploadProgram?.id || ""} 
            onChange={handleProgramChange}
          >
            {programs.length > 0 ? (
              programs.map((p) => (<MenuItem key={p.id} value={p.id}>{p.name}</MenuItem>))
            ) : (
              <MenuItem disabled value="">Keine Studiengänge verfügbar</MenuItem>
            )}
          </TextField>

          <TextField 
            select 
            label="Prüfungsordnung (PO)" 
            fullWidth 
            value={uploadPo} 
            disabled={!uploadProgram}
            onChange={(e) => setUploadPo(e.target.value)}
          >
            {uploadProgram && uploadProgram.versions.length > 0 ? (
              uploadProgram.versions.map((v) => (<MenuItem key={v} value={v}>{v}</MenuItem>))
            ) : (
              <MenuItem disabled value="">{uploadProgram ? "Keine POs verfügbar" : "Bitte Studiengang wählen"}</MenuItem>
            )}
          </TextField>

          <Autocomplete
            options={programModules}
            getOptionLabel={(option) => option.name}
            value={uploadModule}
            onChange={(_, newValue) => setUploadModule(newValue)}
            disabled={!uploadProgram}
            renderInput={(params) => (
              <TextField 
                {...params} 
                label="Modul auswählen" 
                placeholder="Suche..." 
                fullWidth 
              />
            )}
            noOptionsText="Keine Module gefunden"
          />

          <TextField
             type="date"
             label="Prüfungsdatum"
             InputLabelProps={{ shrink: true }}
             fullWidth
             value={examDate}
             onChange={(e) => setExamDate(e.target.value)}
          />

          <Box 
            onDragEnter={handleDragIn} onDragLeave={handleDragOut} onDragOver={handleDrag} onDrop={handleDrop}
            sx={{ 
                border: '2px dashed', 
                borderColor: dragging ? 'primary.main' : 'divider', 
                bgcolor: dragging ? 'action.hover' : 'background.default',
                borderRadius: 2, 
                p: 4, 
                textAlign: 'center',
                cursor: 'pointer',
                transition: 'all 0.2s'
            }}
          >
            <input accept="application/pdf" style={{ display: 'none' }} id="file-upload" type="file" onChange={handleFileChange} />
            
            {!file ? (
                <label htmlFor="file-upload" style={{ cursor: 'pointer', display: 'block' }}>
                    <CloudUploadRoundedIcon sx={{ fontSize: 40, color: dragging ? "primary.main" : "text.secondary", mb: 1 }} />
                    <Typography variant="body1" fontWeight={500} color="text.primary">
                       PDF hier ablegen
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                       oder klicken zum Auswählen
                    </Typography>
                </label>
            ) : (
                <Stack alignItems="center" spacing={1}>
                    <InsertDriveFileRoundedIcon color="primary" sx={{ fontSize: 40 }} />
                    <Typography variant="body1" fontWeight={600} noWrap sx={{ maxWidth: '100%' }}>{file.name}</Typography>
                    <Button size="small" color="error" onClick={(e) => { e.preventDefault(); setFile(null); }}>
                        Andere Datei wählen
                    </Button>
                </Stack>
            )}
          </Box>

          <TextField
            label="Kommentar (Optional)"
            multiline
            minRows={3}
            fullWidth
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            placeholder="Besonderheiten, Themen..."
          />
        </Stack>
      </DialogContent>
      
      <DialogActions sx={{ p: 3, pt: 0 }}>
        <Button onClick={onClose} color="inherit" sx={{ mr: 1 }}>Abbrechen</Button>
        <Button onClick={handleSubmit} variant="contained" disabled={loading || !file || !uploadModule} disableElevation size="large">
          {loading ? "Lade hoch..." : "Hochladen"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default function Exams() {
  const { user } = useAuth();
  const canUpload = user?.role === "admin" || user?.role === "editor";

  const [uploadOpen, setUploadOpen] = useState(false);
  const [programs, setPrograms] = useState<Program[]>([]);
  const [selectedProgram, setSelectedProgram] = useState<Program | null>(null); 
  const [selectedPo, setSelectedPo] = useState("");
  const [modules, setModules] = useState<Module[]>([]);
  const [search, setSearch] = useState("");

  useEffect(() => {
    getPrograms().then(({ data }) => {
        if (data && data.length > 0) {
            setPrograms(data);
            setSelectedProgram(data[0]);
            setSelectedPo(data[0].versions[0]);
        }
    });
  }, []);

  useEffect(() => {
    if (selectedProgram) {
        getProgramModules({ path: { id: selectedProgram.id } })
            .then(({ data }) => setModules(data || []));
        if (!selectedProgram.versions.includes(selectedPo)) {
            setSelectedPo(selectedProgram.versions[0]);
        }
    }
  }, [selectedProgram, selectedPo]);

  const filteredModules = useMemo(() => {
    if (!search) return modules;
    return modules.filter(m => m.name.toLowerCase().includes(search.toLowerCase()));
  }, [modules, search]);

  return (
    <Sidebar user={user} title="Rekos">
      <Container maxWidth="md" sx={{ mt: 5, mb: 10 }}>
        
        <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <Box>
                <Typography variant="h4" component="h1" fontWeight={700} gutterBottom>
                    Modulverzeichnis
                </Typography>
                <Typography variant="body1" color="text.secondary">
                    Verwalte deine Module und Klausuranmeldungen.
                </Typography>
            </Box>

            {canUpload && (
                <Button 
                    variant="contained" 
                    startIcon={<AddRoundedIcon />} 
                    onClick={() => setUploadOpen(true)}
                    sx={{ mt: 1, textTransform: 'none', fontWeight: 600, borderRadius: 2 }}
                    disableElevation
                >
                    Material hochladen
                </Button>
            )}
        </Box>

        <Paper elevation={0} sx={{ p: 3, mb: 5, borderRadius: 3, bgcolor: "background.paper", border: "1px solid", borderColor: "divider" }}>
            <Stack spacing={2}>
                <TextField
                    fullWidth
                    placeholder="Suche nach Modul..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    InputProps={{
                        startAdornment: <InputAdornment position="start"><SearchRoundedIcon color="primary" /></InputAdornment>,
                        sx: { borderRadius: 2 }
                    }}
                />
                <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
                    <TextField
                        select
                        label="Studiengang"
                        value={selectedProgram?.id || ""}
                        onChange={(e) => {
                            const prog = programs.find(p => p.id === Number(e.target.value));
                            if(prog) setSelectedProgram(prog);
                        }}
                        size="small"
                        sx={{ flex: 2, minWidth: 220 }}
                    >
                        {programs.length > 0 ? (
                            programs.map((prog) => (<MenuItem key={prog.id} value={prog.id}>{prog.name}</MenuItem>))
                        ) : (
                            <MenuItem disabled value="">Lädt...</MenuItem>
                        )}
                    </TextField>

                    <TextField
                        select
                        label="Prüfungsordnung"
                        value={selectedPo}
                        onChange={(e) => setSelectedPo(e.target.value)}
                        size="small"
                        sx={{ flex: 1, minWidth: 120 }}
                        disabled={!selectedProgram}
                    >
                        {selectedProgram?.versions.map((ver) => (<MenuItem key={ver} value={ver}>{ver}</MenuItem>))}
                    </TextField>
                </Box>
            </Stack>
        </Paper>

        <Stack spacing={4}>
          {filteredModules.length > 0 ? (
              <Box>
                <Typography variant="h6" color="primary.main" fontWeight={600} sx={{ mb: 1, px: 2 }}>Module</Typography>
                <List sx={{ p: 0 }}>
                    {filteredModules.map((mod, index) => (
                        <ListItem key={mod.id} disablePadding divider={index < filteredModules.length - 1}>
                            <ListItemButton 
                                component={RouterLink} 
                                to="/rekos/klausuren/modul"
                                state={{ studiengang: selectedProgram?.name, po: selectedPo, modul: mod.name, modulId: mod.id }}
                                sx={{ borderRadius: 1, py: 1.5 }}
                            >
                                <ListItemText primary={mod.name} primaryTypographyProps={{ fontWeight: 500 }} />
                            </ListItemButton>
                        </ListItem>
                    ))}
                </List>
              </Box>
          ) : (
             <Box textAlign="center" py={4} sx={{ opacity: 0.6 }}><Typography variant="h6">Keine Module gefunden.</Typography></Box>
          )}
        </Stack>

        {programs.length > 0 && (
            <UploadDialog 
                open={uploadOpen} 
                onClose={() => setUploadOpen(false)} 
                programs={programs} 
                onSuccess={() => { /* Refresh */ }} 
            />
        )}

      </Container>
    </Sidebar>
  );
}
