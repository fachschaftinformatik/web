'use client';

import { useMemo, useState, useEffect } from "react";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
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
import Button from "@mui/material/Button";
import IconButton from "@mui/material/IconButton";
import Autocomplete from "@mui/material/Autocomplete";
import Alert from "@mui/material/Alert";
import Tooltip from "@mui/material/Tooltip";
import Divider from "@mui/material/Divider";
import ListItemSecondaryAction from "@mui/material/ListItemSecondaryAction";
import CircularProgress from "@mui/material/CircularProgress";
import Skeleton from "@mui/material/Skeleton";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import AddRoundedIcon from '@mui/icons-material/AddRounded';
import CloudUploadRoundedIcon from '@mui/icons-material/CloudUploadRounded';
import InsertDriveFileRoundedIcon from '@mui/icons-material/InsertDriveFileRounded';
import CloseRoundedIcon from '@mui/icons-material/CloseRounded';
import DeleteOutlineRoundedIcon from '@mui/icons-material/DeleteOutlineRounded';
import UploadRoundedIcon from '@mui/icons-material/UploadRounded';
import Link from "next/link";

import { useAuth } from "@lib/auth";
import { Sidebar } from "@components/layout";
import { getPrograms, getProgramModules, postArchive, getArchive } from "@lib/api";
import type { DtoProgramResponse as Program, DtoModuleResponse as Module } from "@lib/api";

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
        if (currentProgram && currentProgram.id !== undefined) {
            getProgramModules({ path: { programId: String(currentProgram.id) } })
                .then(({ data }) => setProgramModules(data || []))
                .catch(() => setProgramModules([]));

            if (!currentPo || (currentProgram.versions && !currentProgram.versions.includes(currentPo))) {
                setCurrentPo(currentProgram.versions?.[0] || "");
            }
            setCurrentModule(null);
        } else {
            setProgramModules([]);
        }
    }, [currentProgram, currentPo]);

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files[0]) setFile(e.target.files[0]);
    };

    const handleDrag = (e: React.DragEvent) => { e.preventDefault(); e.stopPropagation(); };
    const handleDragIn = (e: React.DragEvent) => { e.preventDefault(); e.stopPropagation(); setDragging(true); };
    const handleDragOut = (e: React.DragEvent) => { e.preventDefault(); e.stopPropagation(); setDragging(false); };
    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault(); e.stopPropagation(); setDragging(false);
        if (e.dataTransfer.files && e.dataTransfer.files[0]) {
            if (e.dataTransfer.files[0].type === "application/pdf") setFile(e.dataTransfer.files[0]);
        }
    };

    const handleAddAssignment = () => {
        if (currentProgram && currentPo && currentModule) {
            const exists = assignments.some(a =>
                String(a.program.id) === String(currentProgram.id) &&
                a.version === currentPo &&
                String(a.module.id) === String(currentModule.id)
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

        const finalAssignments = [...assignments];
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
            for (const a of finalAssignments) {
                const { error: apiError } = await postArchive({
                    body: {
                        file: file,
                        date: examDate,
                        module_id: String(a.module.id),
                        version: a.version,
                        comment: comment || undefined
                    }
                });

                if (apiError) {
                    const msg = (apiError as { message?: string }).message || "Fehler beim Hochladen.";
                    setError(msg);
                    setLoading(false);
                    return;
                }
            }

            onSuccess();
            onClose();
            setFile(null);
            setAssignments([]);
            setComment("");
            setExamDate("");
            setCurrentModule(null);
        } catch (err) {
            console.error("Fehler beim Hochladen:", err);
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
                        <span>
                            <IconButton
                                onClick={handleSubmit}
                                disabled={loading || !isFormValid}
                                color="primary"
                                sx={{ mr: 1 }}
                            >
                                {loading ? <CircularProgress size={24} /> : <UploadRoundedIcon />}
                            </IconButton>
                        </span>
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
                        id="examDate"
                        name="examDate"
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
                            id="upload-program-select"
                            name="program"
                            select
                            label="Studiengang"
                            fullWidth
                            size="small"
                            value={String(currentProgram?.id || "")}
                            onChange={(e) => {
                                const prog = programs.find(p => String(p.id) === String(e.target.value));
                                setCurrentProgram(prog || null);
                            }}
                        >
                            {programs.length > 0 ? (
                                programs.map((p) => (<MenuItem key={String(p.id)} value={String(p.id)}>{p.name}</MenuItem>))
                            ) : (
                                <MenuItem disabled value="">Keine Studiengänge verfügbar</MenuItem>
                            )}
                        </TextField>

                        <Stack direction="row" spacing={2} alignItems="stretch">
                            <TextField
                                id="upload-po-select"
                                name="po"
                                select
                                label="PO"
                                size="small"
                                sx={{ minWidth: 130 }}
                                value={currentPo}
                                disabled={!currentProgram}
                                onChange={(e) => setCurrentPo(e.target.value)}
                            >
                                {currentProgram && currentProgram.versions && currentProgram.versions.length > 0 ? (
                                    currentProgram.versions.map((v) => (<MenuItem key={v} value={v}>{v}</MenuItem>))
                                ) : (
                                    <MenuItem disabled value="">-</MenuItem>
                                )}
                            </TextField>

                            <Autocomplete
                                options={programModules}
                                getOptionLabel={(option) => option.name || ""}
                                filterOptions={(options, state) => {
                                    const s = state.inputValue.toLowerCase();
                                    return options.filter(o =>
                                        (o.name && o.name.toLowerCase().includes(s)) ||
                                        (o.alias && o.alias.toLowerCase().includes(s))
                                    );
                                }}
                                value={currentModule}
                                onChange={(_, newValue) => setCurrentModule(newValue)}
                                disabled={!currentProgram}
                                isOptionEqualToValue={(o, v) => String(o.id) === String(v.id)}
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

export default function Archive() {
    const { user } = useAuth();
    const canUpload = user?.role === "admin" || user?.role === "editor";

    const [uploadOpen, setUploadOpen] = useState(false);
    const [programs, setPrograms] = useState<Program[]>([]);
    const [selectedPrograms, setSelectedPrograms] = useState<Program[]>([]);
    const [selectedPo, setSelectedPo] = useState("all");
    const [modules, setModules] = useState<Module[]>([]);
    const [search, setSearch] = useState("");

    const [activeModuleIds, setActiveModuleIds] = useState<Set<string>>(new Set());
    const [refreshTrigger, setRefreshTrigger] = useState(0);
    const [loading, setLoading] = useState(true);

    const sortedPos = useMemo(() => {
        if (selectedPrograms.length === 0) return [];
        const allVersions = new Set<string>();
        selectedPrograms.forEach(p => {
            p.versions?.forEach(v => allVersions.add(v));
        });
        return Array.from(allVersions).sort((a, b) => b.localeCompare(a, undefined, { numeric: true }));
    }, [selectedPrograms]);

    useEffect(() => {
        setLoading(true);
        getPrograms().then(({ data }) => {
            if (data && data.length > 0) {
                setPrograms(data);

                let defaultProg = data[0];
                if (user?.program_id) {
                    const userProg = data.find(p => String(p.id) === String(user.program_id));
                    if (userProg) defaultProg = userProg;
                }

                setSelectedPrograms([defaultProg]);

                if (defaultProg.versions && defaultProg.versions.length > 0) {
                    if (selectedPo === undefined || selectedPo === "") {
                        setSelectedPo("all");
                    }
                }
            }
        }).finally(() => {
        });
    }, [user?.program_id]);

    useEffect(() => {
        if (selectedPrograms.length > 0) {
            Promise.all(selectedPrograms.map(p => getProgramModules({ path: { programId: String(p.id) } })))
                .then(results => {
                    const allModules = results.flatMap(r => r.data || []);
                    const uniqueModules = Array.from(new Map(allModules.map(m => [String(m.id), m])).values());
                    setModules(uniqueModules);
                });

            void Promise.resolve().then(() => {
                setSelectedPo("all");
            });
        } else if (programs.length > 0) {
            Promise.all(programs.map(p => getProgramModules({ path: { programId: String(p.id) } })))
                .then(results => {
                    const allModules = results.flatMap(r => r.data || []);
                    const uniqueModules = Array.from(new Map(allModules.map(m => [String(m.id), m])).values());
                    setModules(uniqueModules);
                });
        } else {
            setModules([]);
        }
    }, [selectedPrograms, programs]);

    useEffect(() => {
        const fetchArchiveData = async () => {
            try {
                if (selectedPrograms.length > 0 && selectedPo) {
                    const fetchExams = selectedPrograms.map(p => {
                        const query: { program_id: string; version?: string } = { program_id: String(p.id) };
                        if (selectedPo !== "all") {
                            query.version = selectedPo;
                        }
                        return getArchive({ query });
                    });

                    const results = await Promise.all(fetchExams);
                    const allExams = results.flatMap(r => r.data || []);
                    const ids = new Set(allExams.map(e => String(e.module_id)).filter((id): id is string => id !== "undefined"));
                    setActiveModuleIds(ids);
                } else if (selectedPrograms.length === 0) {
                    const { data } = await getArchive({ query: {} });
                    const ids = new Set((data || []).map(e => String(e.module_id)).filter((id): id is string => id !== "undefined"));
                    setActiveModuleIds(ids);
                } else {
                    setActiveModuleIds(new Set());
                }
            } catch (err) {
                console.error(err);
                setActiveModuleIds(new Set());
            } finally {
                setLoading(false);
            }
        };

        fetchArchiveData();
    }, [selectedPrograms, selectedPo, refreshTrigger]);

    const filteredModules = useMemo(() => {
        let res = modules.filter(m => m.id !== undefined && activeModuleIds.has(String(m.id)));
        if (search) {
            const s = search.toLowerCase();
            res = res.filter(m =>
                (m.name && m.name.toLowerCase().includes(s)) ||
                (m.alias && m.alias.toLowerCase().includes(s))
            );
        }
        return res;
    }, [modules, search, activeModuleIds]);

    return (
        <Sidebar user={user} title="Archiv" maxWidth="lg">
            <Box>

                <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                    <Box>
                        <Typography variant="h4" component="h1" fontWeight={700} gutterBottom>
                            Klausurrekonstruktionen
                        </Typography>
                        <Typography variant="body1" color="text.secondary">
                            Zur Vorbereitung auf Eure Prüfungen stellen wir Euch eine Ansammlung von Klausurrekonstruktionen zur Verfügung.
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

                        <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', width: '100%', alignItems: 'center' }}>
                            <Autocomplete
                                multiple
                                id="main-program-select"
                                options={programs}
                                getOptionLabel={(option) => option.name || ""}
                                value={selectedPrograms}
                                isOptionEqualToValue={(option, value) => String(option.id) === String(value.id)}
                                onChange={(_, newValue) => setSelectedPrograms(newValue)}
                                noOptionsText="Keine Ergebnisse"
                                renderInput={(params) => (
                                    <TextField
                                        {...params}
                                        label="Studiengang"
                                        size="small"
                                        placeholder="Wählen..."
                                    />
                                )}
                                size="small"
                                sx={{ flex: 2, minWidth: 300 }}
                            />

                            <TextField
                                id="main-po-select"
                                name="po"
                                select
                                label="Prüfungsordnung"
                                value={selectedPo}
                                onChange={(e) => setSelectedPo(e.target.value)}
                                size="small"
                                sx={{ flex: 1, minWidth: 120 }}
                                disabled={selectedPrograms.length === 0}
                            >
                                <MenuItem value="all">Alle</MenuItem>
                                {sortedPos.map((ver) => (
                                    <MenuItem key={ver} value={ver}>{ver}</MenuItem>
                                ))}
                            </TextField>
                        </Box>
                    </Stack>
                </Paper>

                <Stack spacing={4}>
                    {loading ? (
                        <Box>
                            <Skeleton variant="text" width={100} sx={{ mb: 1, ml: 2 }} />
                            <List sx={{ p: 0 }}>
                                {[1, 2, 3, 4, 5].map((i) => (
                                    <ListItem key={i} divider={i < 5} sx={{ py: 1.5 }}>
                                        <Skeleton variant="text" width="60%" height={24} />
                                    </ListItem>
                                ))}
                            </List>
                        </Box>
                    ) : filteredModules.length > 0 ? (
                        <Box>
                            <Typography variant="h6" color="primary.main" fontWeight={600} sx={{ mb: 1, px: 2 }}>Module</Typography>
                            <List sx={{ p: 0 }}>
                                {filteredModules.map((mod, index) => (
                                    <ListItem key={String(mod.id)} disablePadding divider={index < filteredModules.length - 1}>
                                        <ListItemButton
                                            component={Link}
                                            href={`/archive/${mod.id}?mod=${encodeURIComponent(mod.name || '')}`}
                                            sx={{ borderRadius: 1, py: 1.5 }}
                                        >
                                            <ListItemText
                                                primary={mod.name}
                                                primaryTypographyProps={{ fontWeight: 500 }}
                                            />
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

            </Box>
        </Sidebar>
    );
}
