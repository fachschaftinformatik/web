import { useMemo, useState, useEffect } from "react";
import {
  Box,
  Container,
  Typography,
  Divider,
  Paper,
  Button,
  Stack,
  CircularProgress,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  IconButton,
  TextField,
  Grid
} from "@mui/material";
import { useLocation, useSearchParams, useNavigate } from "react-router-dom";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import ArrowBackRoundedIcon from "@mui/icons-material/ArrowBackRounded";
import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import CloudDownloadRoundedIcon from "@mui/icons-material/CloudDownloadRounded";

import { Sidebar } from "@components/layout";
import { useAuth } from "@lib/auth";
import { getExams, getExamsFile } from "@lib/api";
import type { ExamListEntry } from "@lib/api";

export default function ExamDetailsPage() {
  const { user } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const [params] = useSearchParams();

  // State passed from the previous page
  const state = location.state as { studiengang?: string; po?: string; modul?: string; modulId?: number } || {};
  
  const modulName = state.modul || params.get("mod") || "Modul";
  const modulId = state.modulId ? Number(state.modulId) : (params.get("modulId") ? Number(params.get("modulId")) : null);
  const po = state.po || params.get("po") || "";
  const studiengang = state.studiengang || params.get("sg") || "Allgemein";

  const [exams, setExams] = useState<ExamListEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const [selectedExam, setSelectedExam] = useState<ExamListEntry | null>(null);
  const [downloading, setDownloading] = useState(false);

  useEffect(() => {
    if (!modulId) return;
    setLoading(true);
    // Fetch exams for this specific module
    getExams({ query: { moduleid: modulId } })
      .then(({ data, error: apiError }) => {
        if (apiError) {
            setError("Fehler beim Laden der Klausuren.");
        } else {
            // Sort: Newest exam date first
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

  const handleDownload = async () => {
    if (!selectedExam?.id) return;
    setDownloading(true);
    try {
        const { data, error } = await getExamsFile({ path: { id: selectedExam.id } });
        
        if (error || !data) {
            alert("Fehler beim Herunterladen der Datei.");
            return;
        }

        // Create blob URL and trigger download
        const blob = data instanceof Blob ? data : new Blob([data]);
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `Klausur_${modulName.replace(/\s+/g, '_')}_${selectedExam.exam_date}.pdf`;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
    } catch (e) {
        console.error(e);
        alert("Ein unerwarteter Fehler ist aufgetreten.");
    } finally {
        setDownloading(false);
    }
  };

  const handleCloseDialog = () => {
    setSelectedExam(null);
  };

  return (
    <Sidebar user={user} title={`Reko: ${modulName}`}>
      <Container maxWidth="md" sx={{ mt: 5, mb: 10 }}>
        
        <Button 
            startIcon={<ArrowBackRoundedIcon />} 
            onClick={() => navigate(-1)} 
            sx={{ mb: 3, color: "text.secondary" }}
        >
            Zurück zur Übersicht
        </Button>

        <Typography variant="h4" fontWeight={600} gutterBottom>
          Reko: {modulName}
        </Typography>

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
                                    <Typography variant="body3" color="text.secondary" fontWeight={400}>
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

        {/* Detail Dialog */}
        <Dialog 
            open={!!selectedExam} 
            onClose={handleCloseDialog}
            fullWidth
            maxWidth="sm"
            PaperProps={{ sx: { borderRadius: 3 } }}
        >
            {selectedExam && (
                <>
                    <DialogTitle sx={{ m: 0, p: 3, pb: 1, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <Typography variant="h6" fontWeight={700}>Details zur Reko</Typography>
                        <IconButton onClick={handleCloseDialog}><CloseRoundedIcon /></IconButton>
                    </DialogTitle>
                    <DialogContent sx={{ p: 3 }}>
                        <Stack spacing={3} mt={1}>
                            <TextField 
                                label="Studiengang" 
                                fullWidth 
                                value={studiengang} 
                                slotProps={{ input: { readOnly: true } }} 
                                variant="filled"
                            />
                            <Grid container spacing={2}>
                                <Grid item xs={6}>
                                    <TextField 
                                        label="Prüfungsordnung" 
                                        fullWidth 
                                        value={selectedExam.version} 
                                        slotProps={{ input: { readOnly: true } }} 
                                        variant="filled"
                                    />
                                </Grid>
                                <Grid item xs={6}>
                                    <TextField 
                                        label="Datum der Prüfung" 
                                        fullWidth 
                                        value={selectedExam.exam_date ? new Date(selectedExam.exam_date).toLocaleDateString('de-DE') : ''} 
                                        slotProps={{ input: { readOnly: true } }} 
                                        variant="filled"
                                    />
                                </Grid>
                            </Grid>
                            
                            <TextField 
                                label="Modul" 
                                fullWidth 
                                value={modulName} 
                                slotProps={{ input: { readOnly: true } }} 
                                variant="filled"
                            />

                            <TextField
                                label="Anmerkungen"
                                multiline
                                minRows={3}
                                fullWidth
                                value={selectedExam.comment || "Keine Anmerkungen vorhanden."}
                                slotProps={{ input: { readOnly: true } }} 
                                variant="filled"
                            />
                        </Stack>
                    </DialogContent>
                    <DialogActions sx={{ p: 3, pt: 0 }}>
                        <Button onClick={handleCloseDialog} color="inherit" sx={{ mr: 1 }}>Schließen</Button>
                        <Button 
                            onClick={handleDownload} 
                            variant="contained" 
                            startIcon={downloading ? <CircularProgress size={20} color="inherit" /> : <CloudDownloadRoundedIcon />}
                            disabled={downloading}
                            disableElevation
                            size="large"
                        >
                            {downloading ? "Lade..." : "Herunterladen"}
                        </Button>
                    </DialogActions>
                </>
            )}
        </Dialog>

      </Container>
    </Sidebar>
  );
}
