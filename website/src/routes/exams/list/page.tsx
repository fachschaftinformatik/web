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
  Autocomplete
} from "@mui/material";

import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import SchoolRoundedIcon from '@mui/icons-material/SchoolRounded';
import DescriptionRoundedIcon from '@mui/icons-material/DescriptionRounded';
import ArrowForwardIosRoundedIcon from '@mui/icons-material/ArrowForwardIosRounded';
import AddRoundedIcon from '@mui/icons-material/AddRounded';
import CloudUploadRoundedIcon from '@mui/icons-material/CloudUploadRounded';
import InsertDriveFileRoundedIcon from '@mui/icons-material/InsertDriveFileRounded';
import CloseRoundedIcon from '@mui/icons-material/CloseRounded';
import { Link as RouterLink } from "react-router-dom";

import { useAuth } from "@lib/auth";
import { Sidebar } from "@components/layout";

const studyPrograms = [
  { id: 1, name: "Informatik (B. Sc.)", versions: ["PO2023", "PO2016"] },
  { id: 4, name: "Informatik (M. Sc.)", versions: ["PO2023", "PO2016"] },
  { id: 2, name: "Informatik und Design (B. Sc.)", versions: ["PO2023"] },
  { id: 7, name: "Internetsicherheit (M. Sc.)", versions: ["PO2023", "PO2016"] },
  { id: 5, name: "Medieninformatik (M. Sc.)", versions: ["PO2023", "PO2016"] },
  { id: 3, name: "Wirtschaftsinformatik (B. Sc.)", versions: ["PO2023", "PO2016"] },
  { id: 6, name: "Wirtschaftsinformatik (M. Sc.)", versions: ["PO2023", "PO2016"] }
];

type Semester = { title: string; items: string[] };

const getModulesForProgram = (programId: number, po: string): Semester[] => {
  if (programId === 1 && po === "PO2023") {
    return [
      { title: "1. Semester", items: ["Logik und diskrete Strukturen", "Einführung in die Programmierung", "Mathematische Grundlagen", "Technische Grundlagen der Informatik", "Technisches Englisch für Informatiker"] },
      { title: "2. Semester", items: ["Algorithmen und Datenstrukturen", "Objektorientierte Programmierung", "Statistik und Lineare Algebra", "Theoretische Informatik", "Betriebssysteme"] },
      { title: "3. Semester", items: ["Datenbanksysteme", "Softwaretechnik", "Mensch-Computer-Interaktion", "Rechnernetze", "Internetsprachen"] },
    ];
  }
  return [{ title: "Semester 1", items: ["Beispielmodul A", "Beispielmodul B"] }];
};

function UploadDialog({ open, onClose, programs }: { open: boolean; onClose: () => void; programs: typeof studyPrograms }) {
  const [uploadProgram, setUploadProgram] = useState(programs[0]);
  const [uploadPo, setUploadPo] = useState(programs[0].versions[0]);
  const [moduleName, setModuleName] = useState<string | null>(null);
  const [file, setFile] = useState<File | null>(null);
  const [dragging, setDragging] = useState(false);

  const availableModules = useMemo(() => {
    const semesters = getModulesForProgram(uploadProgram.id, uploadPo);
    return semesters.flatMap(s => s.items).sort();
  }, [uploadProgram, uploadPo]);

  const handleProgramChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const prog = programs.find(p => p.id === Number(e.target.value));
    if (prog) {
        setUploadProgram(prog);
        setUploadPo(prog.versions[0]);
        setModuleName(null); 
    }
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

  const handleSubmit = () => {
    onClose();
    setFile(null);
    setModuleName(null);
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
          
          <TextField 
            select 
            label="Studiengang" 
            fullWidth 
            value={uploadProgram.id} 
            onChange={handleProgramChange}
          >
            {programs.map((p) => (<MenuItem key={p.id} value={p.id}>{p.name}</MenuItem>))}
          </TextField>

          <TextField 
            select 
            label="Prüfungsordnung (PO)" 
            fullWidth 
            value={uploadPo} 
            onChange={(e) => {
                setUploadPo(e.target.value);
                setModuleName(null);
            }}
          >
            {uploadProgram.versions.map((v) => (<MenuItem key={v} value={v}>{v}</MenuItem>))}
          </TextField>

          <Autocomplete
            options={availableModules}
            value={moduleName}
            onChange={(_, newValue) => setModuleName(newValue)}
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
            label="Anmerkungen (Optional)"
            multiline
            minRows={3}
            fullWidth
            placeholder="Besonderheiten, Themen..."
          />
        </Stack>
      </DialogContent>
      
      <DialogActions sx={{ p: 3, pt: 0 }}>
        <Button onClick={onClose} color="inherit" sx={{ mr: 1 }}>Abbrechen</Button>
        <Button onClick={handleSubmit} variant="contained" disabled={!file || !moduleName} disableElevation size="large">
          Hochladen
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default function Exams() {
  const { user } = useAuth();
  const canUpload = user?.role === "admin" || user?.role === "editor";

  const [uploadOpen, setUploadOpen] = useState(false);
  const [selectedProgram, setSelectedProgram] = useState(studyPrograms[0]); 
  const [selectedPo, setSelectedPo] = useState(studyPrograms[0].versions[0]);
  const [search, setSearch] = useState("");

  useEffect(() => {
    if (!selectedProgram.versions.includes(selectedPo)) setSelectedPo(selectedProgram.versions[0]);
  }, [selectedProgram, selectedPo]);

  const displaySemesters = useMemo(() => {
    const rawData = getModulesForProgram(selectedProgram.id, selectedPo);
    const term = search.toLowerCase();
    if (!term) return rawData;
    return rawData.map(sem => ({
      ...sem,
      items: sem.items.filter(item => item.toLowerCase().includes(term))
    })).filter(sem => sem.items.length > 0);
  }, [selectedProgram, selectedPo, search]);

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
                        value={selectedProgram.id}
                        onChange={(e) => {
                            const prog = studyPrograms.find(p => p.id === Number(e.target.value));
                            if(prog) setSelectedProgram(prog);
                        }}
                        size="small"
                        sx={{ flex: 2, minWidth: 220 }}
                    >
                        {studyPrograms.map((prog) => (<MenuItem key={prog.id} value={prog.id}>{prog.name}</MenuItem>))}
                    </TextField>

                    <TextField
                        select
                        label="Prüfungsordnung"
                        value={selectedPo}
                        onChange={(e) => setSelectedPo(e.target.value)}
                        size="small"
                        sx={{ flex: 1, minWidth: 120 }}
                    >
                        {selectedProgram.versions.map((ver) => (<MenuItem key={ver} value={ver}>{ver}</MenuItem>))}
                    </TextField>
                </Box>
            </Stack>
        </Paper>

        <Stack spacing={4}>
          {displaySemesters.length > 0 ? (
            displaySemesters.map((sem) => (
              <Box key={sem.title}>
                <Typography variant="h6" color="primary.main" fontWeight={600} sx={{ mb: 1, px: 2 }}>{sem.title}</Typography>
                <List sx={{ p: 0 }}>
                    {sem.items.map((label, index) => (
                        <ListItem key={label} disablePadding divider={index < sem.items.length - 1}>
                            <ListItemButton 
                                component={RouterLink} 
                                to="/rekos/klausuren/modul"
                                state={{ studiengang: selectedProgram.name, po: selectedPo, modul: label }}
                                sx={{ borderRadius: 1, py: 1.5 }}
                            >
                                <ListItemText primary={label} primaryTypographyProps={{ fontWeight: 500 }} />
                                <ArrowForwardIosRoundedIcon sx={{ fontSize: 14, color: 'text.disabled' }} />
                            </ListItemButton>
                        </ListItem>
                    ))}
                </List>
              </Box>
            ))
          ) : (
             <Box textAlign="center" py={4} sx={{ opacity: 0.6 }}><Typography variant="h6">Keine Module gefunden.</Typography></Box>
          )}
        </Stack>

        <UploadDialog open={uploadOpen} onClose={() => setUploadOpen(false)} programs={studyPrograms} />

      </Container>
    </Sidebar>
  );
}
