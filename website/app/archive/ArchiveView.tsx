'use client';

import { useState, useEffect, useCallback, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Paper from "@mui/material/Paper";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import CircularProgress from "@mui/material/CircularProgress";
import Alert from "@mui/material/Alert";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";
import DialogActions from "@mui/material/DialogActions";
import IconButton from "@mui/material/IconButton";
import TextField from "@mui/material/TextField";
import MenuItem from "@mui/material/MenuItem";
import Autocomplete from "@mui/material/Autocomplete";
import Tooltip from "@mui/material/Tooltip";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import ListItemButton from "@mui/material/ListItemButton";
import Divider from "@mui/material/Divider";
import Skeleton from "@mui/material/Skeleton";
import useMediaQuery from "@mui/material/useMediaQuery";
import { useTheme } from "@mui/material/styles";
import ArrowBackRounded from "@mui/icons-material/ArrowBackRounded";
import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import SaveRoundedIcon from "@mui/icons-material/SaveRounded";
import DeleteRoundedIcon from "@mui/icons-material/DeleteRounded";
import CancelRoundedIcon from "@mui/icons-material/CancelRounded";
import PictureAsPdfIcon from "@mui/icons-material/PictureAsPdfRounded";

import { Sidebar, PageLoader } from "@components/layout";
import { useAuth } from "@lib/auth";
import { getArchive, getArchiveFile, getPrograms, getProgramModules, getArchiveVersions, putArchiveId, deleteArchiveId, deleteArchiveFile } from "@lib/api";
import type { DtoArchiveEntryResponse as ExamListEntry, DtoProgramResponse as Program, DtoModuleResponse as Module } from "@lib/api";

function ArchiveViewContent({ moduleId, examId }: { moduleId: string; examId?: string }) {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
    const { user } = useAuth();
    const router = useRouter();
    const searchParams = useSearchParams();

    const modulName = searchParams.get("mod") || "Modul";

    const [exams, setExams] = useState<ExamListEntry[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    const [selectedExam, setSelectedExam] = useState<ExamListEntry | null>(null);
    const [previewUrl, setPreviewUrl] = useState<string | null>(null);
    const [isEditing, setIsEditing] = useState(false);
    const [isSaving, setIsSaving] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

    const [formDate, setFormDate] = useState("");
    const [formComment, setFormComment] = useState("");
    const [formProgram, setFormProgram] = useState<Program | null>(null);

    const [programs, setPrograms] = useState<Program[]>([]);
    const [programModules, setProgramModules] = useState<Module[]>([]);
    const [examVersions, setExamVersions] = useState<ExamListEntry[]>([]);
    const [fetchingVersions, setFetchingVersions] = useState(false);

    const [formPo, setFormPo] = useState("");
    const [formModule, setFormModule] = useState<Module | null>(null);

    const canEdit = user?.role === "admin" || user?.role === "editor";

    const fetchExams = useCallback(() => {
        if (!moduleId) return;
        setLoading(true);
        getArchive({ query: { module_id: String(moduleId) } })
            .then(({ data, error: apiError }) => {
                if (apiError) {
                    setError("Fehler beim Laden der Klausuren.");
                } else {
                    const sorted = (data || []).sort((a, b) => {
                        const dateA = a.exam_date ? new Date(a.exam_date).getTime() : 0;
                        const dateB = b.exam_date ? new Date(b.exam_date).getTime() : 0;
                        return dateB - dateA;
                    });
                    setExams(sorted);
                    
                    if (examId) {
                        const target = sorted.find(e => String(e.id) === String(examId));
                        if (target) setSelectedExam(target);
                    }
                }
            })
            .catch(() => setError("Netzwerkfehler."))
            .finally(() => setLoading(false));
    }, [moduleId, examId]);

    useEffect(() => {
        fetchExams();
        getPrograms().then(({ data }) => {
            if (data) setPrograms(data);
        });
    }, [fetchExams]);

    useEffect(() => {
        if (!selectedExam) {
            setPreviewUrl(prev => {
                if (prev) URL.revokeObjectURL(prev);
                return null;
            });
            return;
        }

        setFormDate(selectedExam.exam_date || "");
        setFormComment(selectedExam.comment || "");
        setFormPo(selectedExam.version || "");
        setIsEditing(false);

        const prog = programs.find(p => String(p.id) === String(selectedExam.program_id));
        setFormProgram(prog || null);

        if (prog && prog.id !== undefined) {
            getProgramModules({ path: { programId: String(prog.id) } }).then(({ data: mods }) => {
                const modules = mods || [];
                setProgramModules(modules);
                const mod = modules.find(m => String(m.id) === String(selectedExam.module_id));
                setFormModule(mod || null);
            });
        }

        if (selectedExam.id) {
            setFetchingVersions(true);
            getArchiveVersions({ path: { entryId: String(selectedExam.id) } })
                .then(({ data }) => setExamVersions(data || []))
                .finally(() => setFetchingVersions(false));
        }

        let active = true;
        getArchiveFile({ 
            path: { entryId: String(selectedExam.id) },
            query: { file_id: String(selectedExam.file_id) }
        })
            .then(({ data }) => {
                if (active && data) {
                    const blob = data instanceof Blob
                        ? data
                        : data instanceof ArrayBuffer
                            ? new Blob([data], { type: 'application/pdf' })
                            : typeof data === 'string'
                                ? new Blob([data], { type: 'application/pdf' })
                                : new Blob([JSON.stringify(data)], { type: 'application/pdf' });
                    const url = URL.createObjectURL(blob);
                    setPreviewUrl(prev => {
                        if (prev) URL.revokeObjectURL(prev);
                        return url;
                    });
                }
            })
            .catch(console.error);

        return () => {
            active = false;
        };
    }, [selectedExam, programs]);

    useEffect(() => {
        if (!isEditing || !formProgram || formProgram.id === undefined) return;

        getProgramModules({ path: { programId: String(formProgram.id) } })
            .then(({ data }) => setProgramModules(data || []));

        if (formPo && formProgram.versions && !formProgram.versions.includes(formPo)) {
            setFormPo(formProgram.versions[0] || "");
        }
    }, [formProgram, isEditing, formPo]);

    const handleSave = async () => {
        if (!selectedExam || !formProgram || !formPo || !formModule || !formDate) return;
        setIsSaving(true);

        try {
            const { data, error } = await putArchiveId({
                path: { entryId: String(selectedExam.id) },
                body: {
                    module_id: String(formModule.id),
                    version: formPo,
                    date: formDate,
                    comment: formComment
                }
            });

            if (error) {
                alert("Fehler beim Speichern.");
            } else {
                setIsEditing(false);
                fetchExams();
                if (data) setSelectedExam(data);
            }
        } catch (err) {
            console.error(err);
            alert("Ein Fehler ist aufgetreten.");
        } finally {
            setIsSaving(false);
        }
    };

    const handleDelete = async () => {
        if (!selectedExam) return;
        setIsDeleting(true);
        try {
            if (examVersions.length > 1) {
                const { error } = await deleteArchiveFile({
                    path: { fileId: String(selectedExam.file_id) },
                });
                if (error) {
                    alert("Fehler beim Löschen der Revision.");
                } else {
                    setDeleteConfirmOpen(false);
                    setIsEditing(false);
                    fetchExams();
                    const { data } = await getArchive({ query: { module_id: String(moduleId) } });
                    if (data && data.length > 0) {
                        const entry = data.find(e => String(e.id) === String(selectedExam.id));
                        if (entry) setSelectedExam(entry);
                        else setSelectedExam(null);
                    }
                }
            } else {
                const { error } = await deleteArchiveId({
                    path: { entryId: String(selectedExam.id) },
                });

                if (error) {
                    alert("Fehler beim Löschen.");
                } else {
                    setDeleteConfirmOpen(false);
                    setSelectedExam(null);
                    setIsEditing(false);
                    fetchExams();
                    router.push(`/archive/${moduleId}`);
                }
            }
        } catch (err) {
            console.error(err);
            alert("Ein Fehler ist aufgetreten.");
        } finally {
            setIsDeleting(false);
        }
    };

    const handleCloseDialog = () => {
        if (isSaving || isDeleting) return;
        router.push(`/archive/${moduleId}?mod=${encodeURIComponent(modulName)}`);
    };

    useEffect(() => {
        if (!examId) {
            setSelectedExam(null);
            return;
        }
        if (exams.length > 0) {
            const target = exams.find(e => String(e.id) === String(examId));
            if (target) setSelectedExam(target);
        }
    }, [examId, exams]);

    return (
        <Sidebar user={user} title="Archiv" maxWidth="lg">
            <Box>
                <Box>
                    <Button
                        startIcon={<ArrowBackRounded />}
                        component={Link}
                        href="/archive"
                        sx={{
                            mb: 2,
                            color: 'text.secondary',
                            textTransform: 'none',
                            '&:hover': { color: 'primary.main', bgcolor: 'transparent', textDecoration: 'underline' }
                        }}
                    >
                        Zurück
                    </Button>
                    <Typography variant="h4" component="h1" fontWeight={700} gutterBottom>
                        {modulName}
                    </Typography>
                </Box>

                {error && <Alert severity="error" sx={{ mb: 3 }}>{error}</Alert>}

                {loading ? (
                    <Paper elevation={0} sx={{ border: "1px solid", borderColor: "divider", borderRadius: 3, overflow: 'hidden' }}>
                        <List disablePadding>
                            {[1, 2, 3].map((i) => (
                                <Box key={i}>
                                    {i > 1 && <Divider />}
                                    <ListItem sx={{ py: 2, px: 3 }}>
                                        <ListItemText
                                            primary={<Skeleton variant="text" width="40%" height={24} />}
                                            secondary={<Skeleton variant="text" width="60%" height={20} />}
                                        />
                                    </ListItem>
                                </Box>
                            ))}
                        </List>
                    </Paper>
                ) : exams.length === 0 ? (
                    <Paper sx={{ p: 4, textAlign: "center", borderRadius: 2, bgcolor: "background.default", border: "1px dashed", borderColor: "divider" }}>
                        <Typography color="text.secondary">Keine Rekonstruktionen für dieses Modul gefunden.</Typography>
                    </Paper>
                ) : (
                    <Paper elevation={0} sx={{ border: "1px solid", borderColor: "divider", borderRadius: 3, overflow: 'hidden' }}>
                        <List disablePadding>
                            {exams?.map((exam, idx) => {
                                const prog = programs?.find(p => String(p.id) === String(exam.program_id));
                                const poLabel = prog ? `${prog.name} (${exam.version})` : (exam.version || "");
                                return (
                                    <Box key={String(exam.id)}>
                                        {idx > 0 && <Divider />}
                                        <ListItem disablePadding>
                                            <ListItemButton 
                                                onClick={() => router.push(`/archive/${moduleId}/${exam.id}?mod=${encodeURIComponent(modulName)}`)} 
                                                sx={{ py: 2 }}
                                                selected={String(exam.id) === String(examId)}
                                            >
                                                <ListItemText
                                                    primary={poLabel}
                                                    secondary={
                                                        <Stack direction="row" spacing={1.5} alignItems="center" sx={{ mt: 0.5 }}>
                                                            <Typography variant="caption" color="text.secondary">
                                                                {exam.exam_date ? new Date(exam.exam_date).toLocaleDateString('de-DE') : "Datum unbekannt"}
                                                            </Typography>
                                                            <Typography variant="caption" color="text.secondary">•</Typography>
                                                            <Typography variant="caption" color="text.secondary">
                                                                von {exam.uploader_name || "Anonym"}
                                                            </Typography>
                                                            {exam.comment && (
                                                                <>
                                                                    <Typography variant="caption" color="text.secondary">•</Typography>
                                                                    <Typography variant="caption" color="text.secondary" noWrap sx={{ maxWidth: 200, fontStyle: 'italic' }}>
                                                                        "{exam.comment}"
                                                                    </Typography>
                                                                </>
                                                            )}
                                                        </Stack>
                                                    }
                                                />
                                            </ListItemButton>
                                        </ListItem>
                                    </Box>
                                );
                            })}
                        </List>
                    </Paper>
                )}

                <Dialog
                    open={!!selectedExam}
                    onClose={handleCloseDialog}
                    fullWidth
                    maxWidth="sm"
                    PaperProps={{ sx: { borderRadius: 3, height: isMobile ? 'auto' : '90vh' } }}
                >
                    {selectedExam && (
                        <>
                            <DialogTitle sx={{ m: 0, p: 3, pb: 1, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                <Typography variant="h6" component="span" fontWeight={700}>
                                    {isEditing ? "Details bearbeiten" : "Details zur Reko"}
                                </Typography>
                                <Box>
                                    {canEdit && !isEditing && (
                                        <Tooltip title="Bearbeiten">
                                            <IconButton
                                                onClick={() => setIsEditing(true)}
                                                sx={{ mr: 1, color: 'text.secondary' }}
                                            >
                                                <EditRoundedIcon />
                                            </IconButton>
                                        </Tooltip>
                                    )}

                                    {isEditing && (
                                        <>
                                            <Tooltip title={examVersions.length > 1 ? "Diese Revision löschen" : "Ganze Klausur löschen"}>
                                                <IconButton
                                                    onClick={() => setDeleteConfirmOpen(true)}
                                                    disabled={isSaving || isDeleting}
                                                    sx={{ color: 'text.secondary', mr: 1 }}
                                                >
                                                    <DeleteRoundedIcon />
                                                </IconButton>
                                            </Tooltip>

                                            <Tooltip title="Speichern">
                                                <span>
                                                    <IconButton
                                                        onClick={handleSave}
                                                        disabled={isSaving}
                                                        sx={{ color: 'text.secondary', mr: 1 }}
                                                    >
                                                        {isSaving ? <CircularProgress size={24} /> : <SaveRoundedIcon />}
                                                    </IconButton>
                                                </span>
                                            </Tooltip>

                                            <Tooltip title="Abbrechen">
                                                <IconButton
                                                    onClick={() => setIsEditing(false)}
                                                    disabled={isSaving}
                                                    sx={{ mr: 1, color: 'text.secondary' }}
                                                >
                                                    <CancelRoundedIcon />
                                                </IconButton>
                                            </Tooltip>
                                        </>
                                    )}

                                    <Tooltip title="Schließen">
                                        <IconButton onClick={handleCloseDialog} disabled={isSaving} sx={{ color: 'text.secondary' }}>
                                            <CloseRoundedIcon />
                                        </IconButton>
                                    </Tooltip>
                                </Box>
                            </DialogTitle>

                            <DialogContent sx={{ p: 3, display: 'flex', flexDirection: 'column', gap: 2 }}>


                                <TextField
                                    id="edit-exam-date"
                                    name="date"
                                    label="Prüfungsdatum"
                                    type="date"
                                    fullWidth
                                    value={formDate}
                                    onChange={(e) => setFormDate(e.target.value)}
                                    disabled={!isEditing}
                                    InputLabelProps={{ shrink: true }}
                                    sx={{ mt: 1 }}
                                    size="small"
                                />

                                <Stack spacing={2}>
                                    <Stack direction="row" spacing={2}>
                                        <TextField
                                            id="edit-exam-version"
                                            name="version"
                                            select
                                            label="PO"
                                            value={formPo}
                                            disabled={!isEditing || !formProgram}
                                            onChange={(e) => setFormPo(e.target.value)}
                                            sx={{ width: 150 }}
                                            size="small"
                                        >
                                            {formProgram?.versions && formProgram.versions?.map((v) => (<MenuItem key={v} value={v}>{v}</MenuItem>))}
                                        </TextField>

                                        <TextField
                                            id="edit-exam-program"
                                            name="program"
                                            select
                                            label="Studiengang"
                                            fullWidth
                                            size="small"
                                            value={String(formProgram?.id || "")}
                                            onChange={(e) => {
                                                const prog = programs.find(p => String(p.id) === String(e.target.value));
                                                setFormProgram(prog || null);
                                            }}
                                            disabled={!isEditing}
                                            sx={{ flexGrow: 1 }}
                                        >
                                            {programs?.map((p) => (<MenuItem key={String(p.id)} value={String(p.id)}>{p.name}</MenuItem>))}
                                        </TextField>
                                    </Stack>

                                    <Stack direction="row" spacing={2}>
                                        <TextField
                                            select
                                            label="Revision"
                                            size="small"
                                            sx={{ width: 150 }}
                                            value={String(selectedExam.file_id)}
                                            disabled={isEditing || fetchingVersions}
                                            onChange={(e) => {
                                                const version = examVersions.find(v => String(v.file_id) === String(e.target.value));
                                                if (version) setSelectedExam(version);
                                            }}
                                        >
                                            {examVersions.length === 0 ? (
                                                <MenuItem value={String(selectedExam.file_id)}>
                                                    {selectedExam.edit_version}
                                                </MenuItem>
                                            ) : (
                                                examVersions?.map((v) => (
                                                    <MenuItem key={String(v.file_id)} value={String(v.file_id)}>
                                                        {v.edit_version} {v.file_id === selectedExam.file_id ? "(aktuell)" : ""}
                                                    </MenuItem>
                                                ))
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
                                            value={formModule}
                                            onChange={(_, newValue) => setFormModule(newValue)}
                                            disabled={!isEditing || !formProgram}
                                            noOptionsText="Keine Ergebnisse"
                                            fullWidth
                                            renderInput={(params) => (
                                                <TextField
                                                    {...params}
                                                    id="edit-exam-module"
                                                    name="module"
                                                    label="Modul auswählen"
                                                    size="small"
                                                />
                                            )}
                                            isOptionEqualToValue={(opt, val) => String(opt.id) === String(val.id)}
                                        />
                                    </Stack>
                                </Stack>

                                {isMobile ? (
                                    previewUrl && (
                                        <Button
                                            fullWidth
                                            variant="outlined"
                                            startIcon={<PictureAsPdfIcon />}
                                            onClick={() => window.open(previewUrl, '_blank')}
                                            sx={{ borderRadius: 2, py: 1.5 }}
                                        >
                                            PDF öffnen
                                        </Button>
                                    )
                                ) : (
                                    <Box sx={{
                                        width: '100%',
                                        flexGrow: 1,
                                        bgcolor: 'grey.100',
                                        borderRadius: 2,
                                        overflow: 'hidden',
                                        border: '1px solid',
                                        borderColor: 'divider',
                                        position: 'relative',
                                        minHeight: 400
                                    }}>
                                        {previewUrl ? (
                                            <iframe
                                                src={previewUrl}
                                                width="100%"
                                                height="100%"
                                                style={{ border: 'none' }}
                                                title="Vorschau"
                                            />
                                        ) : (
                                            <Box display="flex" flexDirection="column" alignItems="center" justifyContent="center" height="100%" sx={{ p: 4 }}>
                                                <Skeleton variant="rectangular" width="100%" height="100%" sx={{ position: 'absolute', inset: 0, borderRadius: 0 }} />
                                                <CircularProgress size={30} sx={{ zIndex: 1 }} />
                                                <Typography variant="body2" color="text.secondary" sx={{ ml: 2, mt: 1, zIndex: 1, fontWeight: 500 }}>Lade Vorschau...</Typography>
                                            </Box>
                                        )}
                                    </Box>
                                )}

                                <TextField
                                    id="edit-exam-comment"
                                    name="comment"
                                    label="Anmerkungen"
                                    multiline
                                    maxRows={4}
                                    fullWidth
                                    size="small"
                                    value={formComment}
                                    onChange={(e) => setFormComment(e.target.value)}
                                    disabled={!isEditing}
                                />
                            </DialogContent>
                        </>
                    )}
                </Dialog>

                <Dialog
                    open={deleteConfirmOpen}
                    onClose={() => setDeleteConfirmOpen(false)}
                >
                    <DialogTitle>Klausur löschen?</DialogTitle>
                    <DialogContent>
                        <DialogContentText>
                            Du bist dabei, die Reko für <strong>{modulName}</strong> vom {selectedExam?.exam_date ? new Date(selectedExam.exam_date).toLocaleDateString('de-DE') : ''} zu löschen.
                            <br />
                            Bist du sicher?
                        </DialogContentText>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={() => setDeleteConfirmOpen(false)} color="inherit">Abbrechen</Button>
                        <Button onClick={handleDelete} color="error" autoFocus disabled={isDeleting}>
                            {isDeleting ? "Lösche..." : "Löschen"}
                        </Button>
                    </DialogActions>
                </Dialog>

            </Box>
        </Sidebar>
    );
}

export default function ArchiveView({ moduleId, examId }: { moduleId: string; examId?: string }) {
  return (
    <Suspense fallback={<PageLoader />}>
      <ArchiveViewContent moduleId={moduleId} examId={examId} />
    </Suspense>
  );
}
