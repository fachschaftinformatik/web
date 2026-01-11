import { useState, useEffect, useCallback } from "react";
import {
    Box,
    Container,
    Typography,
    Paper,
    Button,
    Stack,
    CircularProgress,
    Alert,
    Dialog,
    DialogTitle,
    DialogContent,
    DialogContentText,
    DialogActions,
    IconButton,
    TextField,
    MenuItem,
    Autocomplete,
    Tooltip
} from "@mui/material";
import { useLocation, useSearchParams, useNavigate } from "react-router-dom";
import ArrowBackRoundedIcon from "@mui/icons-material/ArrowBackRounded";
import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import SaveRoundedIcon from "@mui/icons-material/SaveRounded";
import DeleteRoundedIcon from "@mui/icons-material/DeleteRounded";
import CancelRoundedIcon from "@mui/icons-material/CancelRounded";

import { Sidebar } from "@components/layout";
import { useAuth } from "@lib/auth";
import { getExams, getExamsFile, getPrograms, getProgramModules } from "@lib/api";
import { client } from "@lib/api/client.gen";
import type { AuthExamResponse as ExamListEntry, AuthProgramResponse as Program, AuthModuleResponse as Module } from "@lib/api";

export default function ExamDetailsPage() {
    const { user } = useAuth();
    const location = useLocation();
    const navigate = useNavigate();
    const [params] = useSearchParams();

    const state = location.state as { studiengang?: string; po?: string; modul?: string; modulId?: number } || {};

    const modulName = state.modul || params.get("mod") || "Modul";
    const modulId = state.modulId ? Number(state.modulId) : (params.get("modulId") ? Number(params.get("modulId")) : null);

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
    const [formPo, setFormPo] = useState("");
    const [formModule, setFormModule] = useState<Module | null>(null);

    const [programs, setPrograms] = useState<Program[]>([]);
    const [programModules, setProgramModules] = useState<Module[]>([]);

    const canEdit = user?.role === "admin" || user?.role === "editor";

    const fetchExams = useCallback(() => {
        if (!modulId) return;
        setLoading(true);
        getExams({ query: { moduleid: modulId } })
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
                }
            })
            .catch(() => setError("Netzwerkfehler."))
            .finally(() => setLoading(false));
    }, [modulId]);

    useEffect(() => {
        fetchExams();
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

        getPrograms().then(({ data }) => {
            if (data) {
                setPrograms(data);
                const prog = data.find(p => p.id === selectedExam.programid);
                setFormProgram(prog || null);

                if (prog && prog.id !== undefined) {
                    getProgramModules({ path: { id: prog.id as number } }).then(({ data: mods }) => {
                        const modules = mods || [];
                        setProgramModules(modules);
                        const mod = modules.find(m => m.id === selectedExam.moduleid);
                        setFormModule(mod || null);
                    });
                }
            }
        });

        let active = true;
        getExamsFile({ path: { id: String(selectedExam.id) } })
            .then(({ data }) => {
                if (active && data) {
                    const blob = data instanceof Blob ? data : new Blob([data as BlobPart]);
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
    }, [selectedExam]);

    useEffect(() => {
        if (!isEditing || !formProgram || formProgram.id === undefined) return;

        getProgramModules({ path: { id: formProgram.id as number } })
            .then(({ data }) => setProgramModules(data || []));

        if (formPo && formProgram.versions && !formProgram.versions.includes(formPo)) {
            setFormPo(formProgram.versions[0] || "");
        }
    }, [formProgram, isEditing, formPo]);

    const handleSave = async () => {
        if (!selectedExam || !formProgram || !formPo || !formModule || !formDate) return;
        setIsSaving(true);

        try {
            const res = await client.request({
                method: "PUT",
                url: "/exams/" + selectedExam.id,
                body: {
                    programid: formProgram.id,
                    version: formPo,
                    moduleid: formModule.id,
                    date: formDate,
                    comment: formComment
                }
            });

            if (res.error) {
                alert("Fehler beim Speichern.");
            } else {
                setIsEditing(false);
                fetchExams();

                if (res.data) {
                    const updated = res.data as ExamListEntry;
                    setSelectedExam(prev => prev ? ({ ...prev, ...updated }) : null);
                }
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
            const res = await client.request({
                method: "DELETE",
                url: "/exams/" + selectedExam.id,
            });

            if (res.error) {
                alert("Fehler beim Löschen.");
            } else {
                setDeleteConfirmOpen(false);
                setSelectedExam(null);
                setIsEditing(false);
                fetchExams();
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
        setSelectedExam(null);
    };

    return (
        <Sidebar user={user} title={"Reko: " + modulName}>
            <Container maxWidth="md" sx={{ mt: 5, mb: 10 }}>

                <Box sx={{ mb: 4 }}>
                    <Button
                        startIcon={<ArrowBackRoundedIcon />}
                        onClick={() => navigate(-1)}
                        sx={{ mb: 2, color: "text.secondary", px: 0, minWidth: 0, justifyContent: 'flex-start' }}
                        disableRipple
                    >
                        Zurück
                    </Button>
                    <Typography variant="h4" component="h1" fontWeight={700} gutterBottom>
                        Reko: {modulName}
                    </Typography>
                    <Typography variant="body1" color="text.secondary">
                        Alle Gedächtnisprotokolle und Altklausuren für dieses Modul.
                    </Typography>
                </Box>

                {error && <Alert severity="error" sx={{ mb: 3 }}>{error}</Alert>}

                {loading ? (
                    <Box display="flex" justifyContent="center" py={8}><CircularProgress /></Box>
                ) : exams.length === 0 ? (
                    <Paper sx={{ p: 4, textAlign: "center", borderRadius: 2, bgcolor: "background.default", border: "1px dashed", borderColor: "divider" }}>
                        <Typography color="text.secondary">Keine Rekonstruktionen für dieses Modul gefunden.</Typography>
                    </Paper>
                ) : (
                    <Stack spacing={2}>
                        {exams.map((entry) => (
                            <Paper
                                key={entry.id}
                                onClick={() => setSelectedExam(entry)}
                                sx={{
                                    p: 2.5,
                                    borderRadius: 3,
                                    transition: "all 0.2s",
                                    cursor: "pointer",
                                    border: "1px solid",
                                    borderColor: "divider",
                                    "&:hover": {
                                        borderColor: "primary.main",
                                        boxShadow: "0 4px 12px rgba(0,0,0,0.08)"
                                    },
                                }}
                            >
                                <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
                                    <Box>
                                        <Stack direction="row" alignItems="baseline" spacing={1} mb={0.5}>
                                            <Typography variant="h6" fontWeight={600}>
                                                {modulName}
                                            </Typography>
                                            {entry.version && (
                                                <Typography variant="body2" color="text.secondary" fontWeight={400}>
                                                    {entry.version}
                                                </Typography>
                                            )}
                                        </Stack>

                                        <Typography variant="body2" color="text.secondary">
                                            Prüfungsdatum: {entry.exam_date ? new Date(entry.exam_date).toLocaleDateString('de-DE') : 'Unbekannt'}
                                        </Typography>

                                        <Typography variant="caption" color="text.disabled" display="block" mt={0.5}>
                                            Hochgeladen von {entry.uploader_name || "Unbekannt"} am {entry.uploaded_at ? new Date(entry.uploaded_at).toLocaleDateString('de-DE') : '-'}
                                        </Typography>
                                    </Box>
                                </Box>
                            </Paper>
                        ))}
                    </Stack>
                )}

                <Dialog
                    open={!!selectedExam}
                    onClose={handleCloseDialog}
                    fullWidth
                    maxWidth="sm"
                    PaperProps={{ sx: { borderRadius: 3, height: '90vh' } }}
                >
                    {selectedExam && (
                        <>
                            <DialogTitle sx={{ m: 0, p: 3, pb: 1, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                <Typography variant="h6" component="span" fontWeight={700}>
                                    {isEditing ? "Reko bearbeiten" : "Details zur Reko"}
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
                                            <Tooltip title="Löschen">
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
                                    <TextField
                                        id="edit-exam-program"
                                        name="program"
                                        select
                                        label="Studiengang"
                                        fullWidth
                                        size="small"
                                        value={formProgram?.id || ""}
                                        onChange={(e) => {
                                            const prog = programs.find(p => p.id === Number(e.target.value));
                                            setFormProgram(prog || null);
                                        }}
                                        disabled={!isEditing}
                                    >
                                        {programs.length === 0 && formProgram?.id ? (
                                            <MenuItem value={formProgram.id}>{formProgram.name}</MenuItem>
                                        ) : null}
                                        {programs.map((p) => (<MenuItem key={p.id} value={p.id}>{p.name}</MenuItem>))}
                                        {programs.length === 0 && !formProgram && (
                                            <MenuItem value=""><em>Wähle einen Studiengang</em></MenuItem>
                                        )}
                                    </TextField>

                                    <Stack direction="row" spacing={2}>
                                        <TextField
                                            id="edit-exam-version"
                                            name="version"
                                            select
                                            label="PO"
                                            value={formPo}
                                            disabled={!isEditing || !formProgram}
                                            onChange={(e) => setFormPo(e.target.value)}
                                            sx={{ minWidth: 120 }}
                                            size="small"
                                        >
                                            {(!formProgram?.versions || formProgram.versions.length === 0) && formPo ? (
                                                <MenuItem value={formPo}>{formPo}</MenuItem>
                                            ) : null}
                                            {formProgram?.versions && formProgram.versions.map((v) => (<MenuItem key={v} value={v}>{v}</MenuItem>))}
                                            {(!formProgram?.versions || formProgram.versions.length === 0) && !formPo && (
                                                <MenuItem value=""><em>PO</em></MenuItem>
                                            )}
                                        </TextField>

                                        <Autocomplete
                                            options={programModules}
                                            getOptionLabel={(option) => option.name || ""}
                                            value={formModule}
                                            onChange={(_, newValue) => setFormModule(newValue)}
                                            disabled={!isEditing || !formProgram}
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
                                            isOptionEqualToValue={(opt, val) => opt.id === val.id}
                                        />
                                    </Stack>
                                </Stack>

                                <Box sx={{
                                    width: '100%',
                                    flexGrow: 1,
                                    bgcolor: 'grey.100',
                                    borderRadius: 2,
                                    overflow: 'hidden',
                                    border: '1px solid',
                                    borderColor: 'divider',
                                    position: 'relative',
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
                                        <Box display="flex" alignItems="center" justifyContent="center" height="100%">
                                            <CircularProgress size={30} />
                                            <Typography variant="body2" color="text.secondary" sx={{ ml: 2 }}>Lade Vorschau...</Typography>
                                        </Box>
                                    )}
                                </Box>

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

            </Container>
        </Sidebar>
    );
}
