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
  Button,
  IconButton,
  Autocomplete,
  Alert,
  Tooltip,
  Divider,
  ListItemSecondaryAction,
  CircularProgress
} from "@mui/material";

import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import AddRoundedIcon from '@mui/icons-material/AddRounded';
import CloudUploadRoundedIcon from '@mui/icons-material/CloudUploadRounded';
import InsertDriveFileRoundedIcon from '@mui/icons-material/InsertDriveFileRounded';
import CloseRoundedIcon from '@mui/icons-material/CloseRounded';
import DeleteOutlineRoundedIcon from '@mui/icons-material/DeleteOutlineRounded';
import UploadRoundedIcon from '@mui/icons-material/UploadRounded';
import { Link as RouterLink } from "react-router-dom";

import { useAuth } from "@lib/auth";
import { Sidebar } from "@components/layout";
import { getPrograms, getProgramModules, postExams, getExams, getAuthCsrf } from "@lib/api";
import type { Program, Module } from "@lib/api";

function UploadDialog({ open, onClose, programs, onSuccess }: { open: boolean; onClose: () => void; programs: Program[]; onSuccess: () => void }) {
  const [file, setFile] = useState<File | null>(null);
  const [examDate, setExamDate] = useState("");
  const [comment, setComment] = useState("");
  
  const [currentProgram, setCurrentProgram] = useState<Program | null>(null);
  const [currentPo, setCurrentPo] = useState("");
  const [currentModule, setCurrentModule] = useState<Module | null>(null);
  const [programModules, setProgramModules] = useState<Module[]>([]);
  
  const [assignments, setAssignments] = useState<{ program: Program; version: string; module: Module }[]>([]);

  const [dragging, setDragging] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (currentProgram) {
      getProgramModules({ path: { id: currentProgram.id } })
        .then(({ data }) => setProgramModules(data || []))
        .catch(() => setProgramModules([]));
      
      if (!currentPo || !currentProgram.versions.includes(currentPo)) {
          setCurrentPo(currentProgram.versions[0] || "");
      }
      setCurrentModule(null);
    } else {
      setProgramModules([]);
    }
  }, [currentProgram]);

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

  const handleAddAssignment = () => {
    if (currentProgram && currentPo && currentModule) {
        const exists = assignments.some(a => 
            a.program.id === currentProgram.id && 
            a.version === currentPo && 
            a.module.id === currentModule.id
        );
        
        if (!exists) {
            setAssignments([...assignments, { program: currentProgram, version: currentPo, module: currentModule }]);
            setCurrentModule(null); 
        }
    }
  };

  const handleRemoveAssignment = (index: number) => {
    setAssignments(assignments.filter((_, i) => i !== index));
  };

  const handleSubmit = async () => {
    setError("");
    
    let finalAssignments = [...assignments];
    if (finalAssignments.length === 0) {
        if (currentProgram && currentPo && currentModule) {
            finalAssignments.push({
                program: currentProgram,
                version: currentPo,
                module: currentModule
            });
        }
    }

    if (!file || !examDate || finalAssignments.length === 0) {
        return; 
    }

    setLoading(true);

    try {
      const { data: csrfData, error: csrfError } = await getAuthCsrf();
      const token = csrfData?.csrf;
      if (csrfError || !token) {
        throw new Error("Sicherheits-Token konnte nicht geladen werden. Bitte neu einloggen.");
      }

      const assignmentData = finalAssignments.map(a => ({
        programid: a.program.id,
        version: a.version,
        moduleid: a.module.id
      }));

      const formData = {
        file: file, 
        date: examDate,
        assignments: JSON.stringify(assignmentData),
        comment: comment
      };

      const { error: apiError } = await postExams({
        // @ts-ignore
        body: formData,
        headers: { "X-CSRF-Token": token }
      });

      if (apiError) {
        // @ts-ignore
        const msg = apiError.message || "Fehler beim Hochladen.";
        setError(msg);
      } else {
        onSuccess();
        onClose();
        setFile(null);
        setAssignments([]);
        setComment("");
        setExamDate("");
        setCurrentModule(null);
      }
    } catch (err) {
      console.error("Upload error:", err);
      setError(err instanceof Error ? err.message : "Netzwerkfehler.");
    } finally {
      setLoading(false);
    }
  };

  const hasValidCurrentSelection = !!(currentProgram && currentPo && currentModule);
  const hasAssignments = assignments.length > 0;
  const isFormValid = !!file && !!examDate && (hasAssignments || hasValidCurrentSelection);

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
        <Box>
            <Tooltip title="Hochladen">
                <IconButton 
                    onClick={handleSubmit} 
                    disabled={loading || !isFormValid}
                    color="primary"
                    sx={{ mr: 1 }}
                >
                    {loading ? <CircularProgress size={24} /> : <UploadRoundedIcon />}
                </IconButton>
            </Tooltip>
            <Tooltip title="Schließen">
                <IconButton onClick={onClose} disabled={loading} sx={{ color: 'text.secondary' }}>
                    <CloseRoundedIcon />
                </IconButton>
            </Tooltip>
        </Box>
      </DialogTitle>
      
      <DialogContent sx={{ p: 3 }}>
        <Stack spacing={3} sx={{ mt: 1 }}>
          {error && <Alert severity="error">{error}</Alert>}
          
          <TextField
             type="date"
             label="Prüfungsdatum"
             InputLabelProps={{ shrink: true }}
             fullWidth
             size="small"
             value={examDate}
             onChange={(e) => setExamDate(e.target.value)}
          />

          <Stack spacing={2}>
              <TextField 
                  select 
                  label="Studiengang" 
                  fullWidth 
                  size="small"
                  value={currentProgram?.id || ""} 
                  onChange={(e) => {
                      const prog = programs.find(p => p.id === Number(e.target.value));
                      setCurrentProgram(prog || null);
                  }}
              >
                  {programs.length > 0 ? (
                      programs.map((p) => (<MenuItem key={p.id} value={p.id}>{p.name}</MenuItem>))
                  ) : (
                      <MenuItem disabled value="">Keine Studiengänge verfügbar</MenuItem>
                  )}
              </TextField>

              <Stack direction="row" spacing={2} alignItems="stretch">
                  <TextField 
                      select 
                      label="PO" 
                      size="small"
                      sx={{ minWidth: 130 }}
                      value={currentPo} 
                      disabled={!currentProgram}
                      onChange={(e) => setCurrentPo(e.target.value)}
                  >
                      {currentProgram && currentProgram.versions.length > 0 ? (
                          currentProgram.versions.map((v) => (<MenuItem key={v} value={v}>{v}</MenuItem>))
                      ) : (
                          <MenuItem disabled value="">-</MenuItem>
                      )}
                  </TextField>

                  <Autocomplete
                      options={programModules}
                      getOptionLabel={(option) => option.name}
                      value={currentModule}
                      onChange={(_, newValue) => setCurrentModule(newValue)}
                      disabled={!currentProgram}
                      fullWidth
                      renderInput={(params) => (
                          <TextField 
                              {...params} 
                              label="Modul auswählen" 
                              size="small"
                              placeholder="Suche..." 
                          />
                      )}
                      noOptionsText="Keine Module gefunden"
                  />
                  
                  <Tooltip title="Zuordnung hinzufügen">
                      <span>
                          <Button 
                              onClick={handleAddAssignment} 
                              disabled={!hasValidCurrentSelection}
                              variant="outlined"
                              sx={{ 
                                  height: 40,
                                  minWidth: 40,
                                  width: 40,
                                  p: 0,
                                  borderRadius: 2,
                                  borderColor: 'rgba(0, 0, 0, 0.23)',
                                  color: hasValidCurrentSelection ? "primary.main" : "action.disabled",
                                  '&:hover': {
                                      borderColor: 'text.primary',
                                      bgcolor: 'action.hover'
                                  }
                              }}
                          >
                              <AddRoundedIcon />
                          </Button>
                      </span>
                  </Tooltip>
              </Stack>
          </Stack>

          {assignments.length > 0 && (
            <Paper variant="outlined" sx={{ maxHeight: 200, overflow: 'auto' }}>
                <List dense disablePadding>
                    {assignments.map((a, idx) => (
                        <div key={idx}>
                            {idx > 0 && <Divider />}
                            <ListItem>
                                <ListItemText 
                                    primary={a.module.name} 
                                    secondary={a.program.name + " • " + a.version}
                                    primaryTypographyProps={{ fontWeight: 500 }}
                                />
                                <ListItemSecondaryAction>
                                    <IconButton edge="end" size="small" onClick={() => handleRemoveAssignment(idx)}>
                                        <DeleteOutlineRoundedIcon fontSize="small" />
                                    </IconButton>
                                </ListItemSecondaryAction>
                            </ListItem>
                        </div>
                    ))}
                </List>
            </Paper>
          )}

          <Box 
            onDragEnter={handleDragIn} onDragLeave={handleDragOut} onDragOver={handleDrag} onDrop={handleDrop}
            sx={{ 
                border: '2px dashed', 
                borderColor: dragging ? 'primary.main' : 'divider', 
                bgcolor: dragging ? 'action.hover' : 'background.default',
                borderRadius: 2, 
                p: 3, 
                textAlign: 'center',
                cursor: 'pointer',
                transition: 'all 0.2s',
                '&:hover': {
                    borderColor: 'text.secondary',
                    bgcolor: 'action.hover'
                }
            }}
          >
            <input accept="application/pdf" style={{ display: 'none' }} id="file-upload" type="file" onChange={handleFileChange} />
            
            {!file ? (
                <label htmlFor="file-upload" style={{ cursor: 'pointer', display: 'block', width: '100%', height: '100%' }}>
                    <CloudUploadRoundedIcon sx={{ fontSize: 32, color: dragging ? "primary.main" : "text.disabled", mb: 0.5 }} />
                    <Typography variant="body2" fontWeight={500} color="text.primary">
                       Klausur-PDF hier ablegen
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                       oder klicken zum Auswählen
                    </Typography>
                </label>
            ) : (
                <Stack alignItems="center" spacing={1} direction="row" justifyContent="center">
                    <InsertDriveFileRoundedIcon color="primary" />
                    <Typography variant="body2" fontWeight={600} noWrap sx={{ maxWidth: 200 }}>{file.name}</Typography>
                    <IconButton size="small" color="error" onClick={(e) => { e.preventDefault(); setFile(null); }}>
                        <DeleteOutlineRoundedIcon fontSize="small" />
                    </IconButton>
                </Stack>
            )}
          </Box>

          <TextField
            label="Kommentar (Optional)"
            multiline
            maxRows={4}
            fullWidth
            size="small"
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            placeholder="Besonderheiten, Themen, Dozent..."
          />
        </Stack>
      </DialogContent>
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
  
  const [activeModuleIds, setActiveModuleIds] = useState<Set<number>>(new Set());
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  useEffect(() => {
    getPrograms().then(({ data }) => {
        if (data && data.length > 0) {
            setPrograms(data);
            
            let defaultProg = data[0];
            if (user?.programid) {
                const userProg = data.find(p => p.id === user.programid);
                if (userProg) defaultProg = userProg;
            }
            
            setSelectedProgram(defaultProg);
            
            if (defaultProg.versions && defaultProg.versions.length > 0) {
                const sortedVersions = [...defaultProg.versions].sort((a, b) => b.localeCompare(a, undefined, { numeric: true }));
                setSelectedPo(sortedVersions[0]);
            }
        }
    });
  }, [user?.programid]);

  useEffect(() => {
    if (selectedProgram) {
        getProgramModules({ path: { id: selectedProgram.id } })
            .then(({ data }) => setModules(data || []));
        
        if (!selectedProgram.versions.includes(selectedPo)) {
            const sorted = [...selectedProgram.versions].sort((a, b) => b.localeCompare(a, undefined, { numeric: true }));
            setSelectedPo(sorted[0] || "");
        }
    }
  }, [selectedProgram, selectedPo]);

  useEffect(() => {
    if (selectedProgram && selectedPo) {
        getExams({ query: { programid: selectedProgram.id, version: selectedPo } })
            .then(({ data }) => {
                if (data) {
                    const ids = new Set(data.map(e => e.moduleid).filter((id): id is number => id !== undefined));
                    setActiveModuleIds(ids);
                } else {
                    setActiveModuleIds(new Set());
                }
            })
            .catch(() => setActiveModuleIds(new Set()));
    }
  }, [selectedProgram, selectedPo, refreshTrigger]);

  const filteredModules = useMemo(() => {
    let res = modules.filter(m => activeModuleIds.has(m.id));
    if (search) {
        res = res.filter(m => m.name.toLowerCase().includes(search.toLowerCase()));
    }
    return res;
  }, [modules, search, activeModuleIds]);

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
        </Box>

        <Paper elevation={0} sx={{ p: 3, mb: 5, borderRadius: 3, bgcolor: "background.paper", border: "1px solid", borderColor: "divider" }}>
            <Stack spacing={2}>
                <Stack direction="row" spacing={2}>
                    <TextField
                        fullWidth
                        placeholder="Suche nach Modul..."
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                        size="small"
                        InputProps={{
                            startAdornment: <InputAdornment position="start"><SearchRoundedIcon color="primary" /></InputAdornment>,
                            sx: { borderRadius: 2 }
                        }}
                    />
                    {canUpload && (
                        <Tooltip title="Material hochladen">
                            <IconButton 
                                onClick={() => setUploadOpen(true)}
                                sx={{ 
                                    bgcolor: 'primary.main', 
                                    color: 'primary.contrastText', 
                                    borderRadius: 2,
                                    width: 40,
                                    height: 40,
                                    '&:hover': { bgcolor: 'primary.dark' }
                                }}
                            >
                                <UploadRoundedIcon />
                            </IconButton>
                        </Tooltip>
                    )}
                </Stack>
                
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
                onSuccess={() => setRefreshTrigger(prev => prev + 1)} 
            />
        )}

      </Container>
    </Sidebar>
  );
}
