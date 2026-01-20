import React, { useState, useEffect, useCallback } from "react";
import {
  Typography, Card, CardMedia, Box, Dialog, DialogTitle, DialogContent,
  DialogActions, IconButton, ImageList, ImageListItem, ImageListItemBar, Pagination,
  Button, Stack, TextField, MenuItem, Snackbar, Alert, CircularProgress, Tooltip, Paper,
  LinearProgress, List, ListItem, ListItemText
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";
import ArrowBackIosNewIcon from "@mui/icons-material/ArrowBackIosNew";
import ArrowForwardIosIcon from "@mui/icons-material/ArrowForwardIos";
import CloudUploadIcon from "@mui/icons-material/CloudUpload";
import UploadRoundedIcon from "@mui/icons-material/UploadRounded";
import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteOutlineRoundedIcon from '@mui/icons-material/DeleteOutlineRounded';
import InsertPhotoRoundedIcon from '@mui/icons-material/InsertPhotoRounded';
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import LinkRoundedIcon from "@mui/icons-material/LinkRounded";
import { alpha } from "@mui/material/styles";
import { useNavigate, useParams } from "react-router-dom";

import { useAuth } from "@lib/auth";
import { Sidebar } from "@components/layout";
import { client } from "@lib/api/client.gen";


type EventItem = { id: string; title: string; created_at: string; cover_path?: string };
type MediaItem = { id: string; event_id: string; title: string; description: string; mime_type: string; uploaded_at: string };

const IMAGES_PER_PAGE = 10;

const pic = (seed: number) => `https://picsum.photos/seed/${seed}/300/200`;

export default function Galerie() {
  const { user } = useAuth();
  const { eventId: urlEventId, imageId: urlImageId } = useParams();

  const [events, setEvents] = useState<EventItem[]>([]);
  const [selectedEventId, setSelectedEventId] = useState<string | null>(null);
  const [media, setMedia] = useState<MediaItem[]>([]);
  const [page, setPage] = useState(1);
  const [loadingMedia, setLoadingMedia] = useState(false);
  const navigate = useNavigate();

  const [uploadOpen, setUploadOpen] = useState(false);
  const [createEventOpen, setCreateEventOpen] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const [lightboxOpen, setLightboxOpen] = useState(false);
  const [lightboxIndex, setLightboxIndex] = useState(0);

  const [preselectedEventId, setPreselectedEventId] = useState<string | null>(null);

  const canUpload = user?.role === "admin" || user?.role === "editor";

  useEffect(() => {
    if (urlEventId) {
      setSelectedEventId(urlEventId);
    }
  }, [urlEventId]);

  const fetchEvents = useCallback(async () => {
    try {
      const res = await client.request({ method: 'GET', url: '/events' });
      if (res.data) {
        const list = res.data as EventItem[];
        setEvents(list);

        if (!urlEventId && list.length > 0) {
          navigate(`/images/${list[0].id}`, { replace: true });
        }
      }
    } catch (e) {
      console.error("Fehler beim Laden der Events", e);
    }
  }, [urlEventId, navigate, setEvents]);

  useEffect(() => {
    void fetchEvents();
  }, [fetchEvents]);

  useEffect(() => {
    if (!selectedEventId) {
      setMedia([]);
      return;
    }

    const abortController = new AbortController();
    const fetchMedia = async () => {
      setLoadingMedia(true);
      try {
        const res = await client.request({
          method: 'GET',
          url: '/media',
          query: { event_id: selectedEventId },
          signal: abortController.signal as never
        });
        
        if (abortController.signal.aborted) return;

        const mediaList = res.data as MediaItem[] || [];
        setMedia(mediaList);
      } catch (e: unknown) {
        if (e instanceof Error && e.name === 'AbortError') return;
        console.error("Fehler beim Laden der Medien", e);
      } finally {
        if (!abortController.signal.aborted) {
          setLoadingMedia(false);
        }
      }
    };
    void fetchMedia();

    return () => {
      abortController.abort();
    };
  }, [selectedEventId]);

  useEffect(() => {
    if (urlImageId && media.length > 0 && selectedEventId === urlEventId) {
      const index = media.findIndex(m => m.id === urlImageId);
      if (index !== -1) {
        setPage(Math.floor(index / IMAGES_PER_PAGE) + 1);
        setLightboxIndex(index);
        setLightboxOpen(true);
      }
    } else if (!urlImageId) {
      setLightboxOpen(false);
    }
  }, [urlImageId, media, selectedEventId, urlEventId]);

  const getImageUrl = (id: string) => `/api/media/${id}/file`;
  const getPreviewUrl = (id: string) => `/api/media/${id}/preview`;
  const getEventCoverUrl = (ev: EventItem) => ev.cover_path ? `/api/events/${ev.id}/cover` : pic(Number(ev.id.toString().slice(-4)));

  const pageCount = Math.max(Math.ceil(media.length / IMAGES_PER_PAGE), 1);
  const displayedMedia = media.slice((page - 1) * IMAGES_PER_PAGE, page * IMAGES_PER_PAGE);
  const currentImage = media[lightboxIndex];

  const openLightbox = (idx: number) => {
    const absoluteIdx = (page - 1) * IMAGES_PER_PAGE + idx;
    setLightboxIndex(absoluteIdx);
    setLightboxOpen(true);
    const img = media[absoluteIdx];
    if (img) navigate(`/images/${selectedEventId}/${img.id}`, { replace: true });
  };

  const closeLightbox = useCallback(() => {
    setLightboxOpen(false);
    navigate(`/images/${selectedEventId}`, { replace: true });
  }, [selectedEventId, navigate]);

  const nextImage = useCallback(() => {
    if (media.length === 0) return;
    const nextIdx = (lightboxIndex + 1) % media.length;
    setLightboxIndex(nextIdx);
    setPage(Math.floor(nextIdx / IMAGES_PER_PAGE) + 1);
    const img = media[nextIdx];
    if (img) navigate(`/images/${selectedEventId}/${img.id}`, { replace: true });
  }, [lightboxIndex, selectedEventId, navigate, media]);

  const prevImage = useCallback(() => {
    if (media.length === 0) return;
    const prevIdx = (lightboxIndex - 1 + media.length) % media.length;
    setLightboxIndex(prevIdx);
    setPage(Math.floor(prevIdx / IMAGES_PER_PAGE) + 1);
    const img = media[prevIdx];
    if (img) navigate(`/images/${selectedEventId}/${img.id}`, { replace: true });
  }, [lightboxIndex, selectedEventId, navigate, media]);

  useEffect(() => {
    if (!lightboxOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "ArrowLeft") prevImage();
      if (e.key === "ArrowRight") nextImage();
      if (e.key === "Escape") closeLightbox();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [lightboxOpen, prevImage, nextImage, closeLightbox]);

  const handleCopyLink = () => {
    if (!currentImage) return;
    const url = window.location.origin + `/images/${selectedEventId}/${currentImage.id}`;
    navigator.clipboard.writeText(url).then(() => {
      setSuccessMessage("Link in die Zwischenablage kopiert!");
    });
  };

  const handleEditImage = () => {
  };

  return (
    <Sidebar user={user} title="Galerie" maxWidth="lg">
      <Box>

        <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <Box>
            <Typography variant="h4" component="h1" fontWeight={700} gutterBottom>
              Galerie
            </Typography>
            <Typography variant="body1" color="text.secondary">
              Fotos von Veranstaltungen, Partys und Events der Fachschaft.
            </Typography>
          </Box>
          {canUpload && (
            <Button
              variant="contained"
              startIcon={<UploadRoundedIcon />}
              onClick={() => setUploadOpen(true)}
              sx={{ boxShadow: 2 }}
            >
              Bilder hochladen
            </Button>
          )}
        </Box>

        <Paper
          elevation={0}
          sx={{
            p: 2,
            mb: 5,
            borderRadius: 3,
            bgcolor: "background.paper",
            border: "1px solid",
            borderColor: "divider"
          }}
        >

          <Box sx={{ display: "flex", justifyContent: "flex-start", minWidth: 0 }}>
            {events.length > 0 ? (
              <ImageList
                sx={(theme) => ({
                  display: "flex",
                  flexDirection: "row",
                  flexWrap: "nowrap",
                  justifyContent: "flex-start",
                  p: 3,
                  pt: 1,
                  overflowX: "auto",
                  overflowY: "visible",
                  "&::-webkit-scrollbar": { display: "none" },
                  msOverflowStyle: "none",
                  scrollbarWidth: "none",
                  width: "100%",
                  gap: 3,
                  "& > li:not(:last-of-type)": { marginRight: theme.spacing(3) },
                })}
              >
                {events.map((ev) => {
                  const selected = ev.id === selectedEventId;

                    return (
                      <ImageListItem
                        key={ev.id}
                        onClick={() => {
                          setMedia([]); // Clear media immediately
                          setPage(1); // Reset page
                          navigate(`/images/${ev.id}`);
                        }}
                        sx={(theme) => ({
                        cursor: "pointer",
                        width: 160,
                        flexShrink: 0,
                        borderRadius: 2,
                        overflow: "visible",
                        boxShadow: theme.shadows[selected ? 4 : 1],
                        outline: selected ? "2px solid" : "none",
                        outlineColor: selected ? theme.palette.primary.main : "transparent",
                        zIndex: selected ? 2 : 1,
                        transform: selected ? "scale(1.02)" : "scale(1)",
                        transition: theme.transitions.create(["transform", "box-shadow", "outline"]),
                        outlineOffset: "2px",
                        "&:hover": {
                          transform: "scale(1.05)",
                          boxShadow: theme.shadows[6],
                          zIndex: 3,
                        }
                      })}
                    >
                      <Box sx={{ overflow: "hidden", borderRadius: 2, bgcolor: "background.default" }}>
                        <Box
                          component="img"
                          src={getEventCoverUrl(ev)}
                          alt={ev.title}
                          loading="lazy"
                          sx={{ width: "100%", height: 100, objectFit: "cover", display: "block" }}
                        />
                        <ImageListItemBar
                          title={ev.title}
                          position="below"
                          sx={(theme) => ({
                            textAlign: "left",
                            px: 1,
                            py: 0.8,
                            bgcolor: selected ? alpha(theme.palette.primary.main, 0.05) : "background.paper",
                            "& .MuiImageListItemBar-title": {
                              fontSize: "0.85rem",
                              fontWeight: selected ? 600 : 500,
                              color: selected ? "primary.main" : "text.primary",
                            }
                          })}
                        />
                      </Box>
                    </ImageListItem>
                  );
                })}
              </ImageList>
            ) : (
              <Typography color="text.secondary" sx={{ py: 2, px: 2 }}>Keine Events vorhanden.</Typography>
            )}
          </Box>
        </Paper>


        {loadingMedia ? (
          <Box display="flex" justifyContent="center" py={8}><CircularProgress /></Box>
        ) : media.length === 0 ? (
          <Box
            py={8}
            textAlign="center"
            sx={(theme) => ({
              border: '1px dashed',
              borderColor: 'divider',
              borderRadius: 3,
              bgcolor: alpha(theme.palette.text.primary, 0.04)
            })}
          >
            <Typography color="text.secondary">
              {selectedEventId ? "Keine Bilder in diesem Album." : "Bitte ein Event oben auswählen."}
            </Typography>
          </Box>
        ) : (
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: { xs: "repeat(2, 1fr)", sm: "repeat(3, 1fr)", md: "repeat(4, 1fr)", lg: "repeat(5, 1fr)" },
              gap: 2
            }}
          >
            {displayedMedia.map((item, idx) => (
              <Card
                key={item.id}
                onClick={() => openLightbox(idx)}
                elevation={0}
                sx={(theme) => ({
                  borderRadius: 2,
                  overflow: "hidden",
                  cursor: "pointer",
                  border: "1px solid",
                  borderColor: "divider",
                  transition: "transform 0.2s, box-shadow 0.2s, border-color 0.2s",
                  "&:hover": {
                    transform: "translateY(-4px)",
                    boxShadow: theme.shadows[4],
                    borderColor: "primary.main",
                  },
                })}
              >
                <CardMedia
                  component="img"
                  image={getPreviewUrl(item.id)}
                  alt={item.title || "Bild"}
                  sx={{ width: "100%", aspectRatio: "3/2", objectFit: "cover" }}
                />
              </Card>
            ))}
          </Box>
        )}

        {pageCount > 1 && (
          <Box sx={{ mt: 5, display: "flex", justifyContent: "center" }}>
            <Pagination
              count={pageCount}
              page={page}
              onChange={(_, p) => setPage(p)}
              color="primary"
              shape="rounded"
              variant="outlined"
              disabled={loadingMedia}
            />
          </Box>
        )}

        {canUpload && (
          <UploadDialog
            open={uploadOpen}
            onClose={() => { setUploadOpen(false); setPreselectedEventId(null); }}
            events={events}
            preselectedId={preselectedEventId}
            onSuccess={() => {
              fetchEvents();
              if (selectedEventId) {
                const current = selectedEventId;
                setSelectedEventId(null);
                setTimeout(() => setSelectedEventId(current), 50);
              }
              setSuccessMessage("Bilder erfolgreich hochgeladen.");
            }}
            onCreateEvent={() => setCreateEventOpen(true)}
          />
        )}

        {canUpload && (
          <CreateEventDialog
            open={createEventOpen}
            onClose={() => setCreateEventOpen(false)}
            onSuccess={(id) => {
              fetchEvents();
              setPreselectedEventId(id);
            }}
          />
        )}

        <Snackbar
          open={Boolean(successMessage)}
          autoHideDuration={4000}
          onClose={() => setSuccessMessage(null)}
          anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
        >
          <Alert severity="success" variant="filled" onClose={() => setSuccessMessage(null)}>
            {successMessage}
          </Alert>
        </Snackbar>

        <Dialog
          open={lightboxOpen}
          onClose={closeLightbox}
          fullWidth
          maxWidth="lg"
          aria-labelledby="bild-dialog-title"
          slotProps={{ paper: { sx: { overflow: "hidden", bgcolor: "background.default", backgroundImage: "none" } } }}
        >
          <DialogTitle
            id="bild-dialog-title"
            sx={{ pr: 2, pl: 3, py: 1.5, position: "relative", zIndex: 1, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
          >
            <Typography variant="h6" noWrap sx={{ flex: 1 }}>
              {currentImage?.title || (currentImage && `Bild ${lightboxIndex + 1}`) || ""}
            </Typography>

            <Stack direction="row" spacing={1} alignItems="center">
              {canUpload && (
                <Tooltip title="Bearbeiten">
                  <IconButton onClick={handleEditImage} size="small">
                    <EditRoundedIcon />
                  </IconButton>
                </Tooltip>
              )}
              <Tooltip title="Link kopieren">
                <IconButton onClick={handleCopyLink} size="small">
                  <LinkRoundedIcon />
                </IconButton>
              </Tooltip>
              <Tooltip title="Schließen">
                <IconButton onClick={closeLightbox} size="small">
                  <CloseIcon />
                </IconButton>
              </Tooltip>
            </Stack>
          </DialogTitle>

          <DialogContent dividers sx={{ position: "relative", p: 0, bgcolor: "background.default", display: "flex", justifyContent: "center", alignItems: "center", minHeight: 500 }}>
            {currentImage && (
              <Box
                component="img"
                src={getImageUrl(currentImage.id)}
                alt={currentImage.title}
                sx={{
                  display: "block",
                  maxWidth: "100%",
                  maxHeight: "80vh",
                  objectFit: "contain",
                }}
              />
            )}

            {displayedMedia.length > 1 && (
              <>
                <IconButton
                  aria-label="previous"
                  onClick={prevImage}
                  sx={{
                    position: "absolute",
                    top: "50%",
                    left: 16,
                    transform: "translateY(-50%)",
                    color: "white",
                    bgcolor: "rgba(0,0,0,0.4)",
                    "&:hover": { bgcolor: "rgba(0,0,0,0.6)" },
                  }}
                >
                  <ArrowBackIosNewIcon />
                </IconButton>
                <IconButton
                  aria-label="next"
                  onClick={nextImage}
                  sx={{
                    position: "absolute",
                    top: "50%",
                    right: 16,
                    transform: "translateY(-50%)",
                    color: "white",
                    bgcolor: "rgba(0,0,0,0.4)",
                    "&:hover": { bgcolor: "rgba(0,0,0,0.6)" },
                  }}
                >
                  <ArrowForwardIosIcon />
                </IconButton>
              </>
            )}
          </DialogContent>
        </Dialog>

      </Box>
    </Sidebar>
  );
}

function getFriendlyErrorMessage(error: { message?: string; error?: string } | string | unknown): string {
  const msg = typeof error === 'string' ? error : (error as { message?: string })?.message || (error as { error?: string })?.error || "";
  const errorMap: Record<string, string> = {
    "File too large": "Die Datei ist zu groß (maximal 256MB erlaubt).",
    "File upload failed": "Der Upload ist fehlgeschlagen. Bitte versuche es erneut.",
    "Invalid event_id": "Das ausgewählte Event ist ungültig.",
    "Missing metadata": "Es fehlen notwendige Daten für den Upload.",
    "invalid_csrf": "Deine Sitzung ist abgelaufen. Bitte lade die Seite neu.",
    "bad_request": "Ungültige Anfrage.",
    "Upload failed": "Der Upload ist fehlgeschlagen. Bitte prüfe Format und Größe der Datei."
  };
  return errorMap[msg] || `Ein Fehler ist aufgetreten: ${msg || "Unbekannter Fehler"}`;
}

function UploadDialog({ open, onClose, events, onSuccess, onCreateEvent, preselectedId }: {
  open: boolean;
  onClose: () => void;
  events: EventItem[];
  onSuccess: () => void;
  onCreateEvent: () => void;
  preselectedId?: string | null;
}) {
  const [files, setFiles] = useState<File[]>([]);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [eventId, setEventId] = useState<string | "">("");

  useEffect(() => {
    if (preselectedId) {
      setEventId(preselectedId);
    }
  }, [preselectedId]);
  const [dragging, setDragging] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState("");

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault(); setDragging(false);
    if (e.dataTransfer.files) {
      const newFiles = Array.from(e.dataTransfer.files).filter(f => f.type.startsWith("image/"));
      setFiles(prev => [...prev, ...newFiles]);
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const newFiles = Array.from(e.target.files);
      setFiles(prev => [...prev, ...newFiles]);
    }
  };

  const removeFile = (idx: number) => {
    setFiles(prev => prev.filter((_, i) => i !== idx));
  };

  const handleSubmit = async () => {
    if (files.length === 0 || !eventId) return;
    setUploading(true);
    setError("");
    setProgress(0);

    let successCount = 0;
    const total = files.length;
    const failedFiles: File[] = [];

    for (let i = 0; i < total; i++) {
      const file = files[i];
      try {
        const formData = new FormData();
        formData.append("file", file);
        formData.append("event_id", String(eventId));
        if (title) formData.append("title", title);
        if (description) formData.append("description", description);

        const res = await client.request({
          method: 'POST',
          url: '/media',
          body: formData,
          bodySerializer: null,
          headers: {
            "Content-Type": null
          }
        });

        if (res.error) throw res.error;
        successCount++;
        setProgress(Math.round(((i + 1) / total) * 100));
      } catch (e: unknown) {
        console.error(`Fehler beim Upload von ${file.name}:`, e);
        setError(getFriendlyErrorMessage(e) + ` (${file.name})`);
        failedFiles.push(file);
      }
    }

    if (successCount > 0) {
      onSuccess();
    }

    if (failedFiles.length === 0 && successCount > 0) {
      onClose();
      setFiles([]);
      setTitle("");
      setDescription("");
    } else {
      setFiles(failedFiles);
    }
    setUploading(false);
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm" PaperProps={{ sx: { borderRadius: 3 } }}>
      <DialogTitle sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h6" fontWeight={700}>Bilder hochladen</Typography>
        <IconButton onClick={onClose} disabled={uploading}><CloseIcon /></IconButton>
      </DialogTitle>
      <DialogContent sx={{ pt: 1 }}>
        <Stack spacing={3} mt={1}>
          {error && <Alert severity="error">{error}</Alert>}

          {uploading && (
            <Box sx={{ width: '100%' }}>
              <LinearProgress variant="determinate" value={progress} sx={{ borderRadius: 1, height: 8 }} />
              <Typography variant="caption" sx={{ mt: 0.5, display: 'block', textAlign: 'center' }}>
                {Math.round(progress)}% hochgeladen
              </Typography>
            </Box>
          )}

          <Stack direction="row" spacing={1} alignItems="stretch">
            <TextField
              id="media-event-select"
              name="event_id"
              select
              fullWidth
              label="Event auswählen"
              value={eventId}
              onChange={(e) => setEventId(e.target.value)}
              size="small"
              disabled={uploading}
            >
              {events.length > 0 ? (
                events.map((e) => <MenuItem key={e.id} value={e.id}>{e.title}</MenuItem>)
              ) : (
                <MenuItem disabled value="">Keine Events verfügbar</MenuItem>
              )}
            </TextField>
            <Tooltip title="Neues Event anlegen">
              <span>
                <Button
                  variant="outlined"
                  sx={{ minWidth: 40, px: 0, borderRadius: 2, height: '100%' }}
                  onClick={onCreateEvent}
                  disabled={uploading}
                >
                  <AddRoundedIcon />
                </Button>
              </span>
            </Tooltip>
          </Stack>

          <Box
            onDragOver={(e) => { e.preventDefault(); if (!uploading) setDragging(true); }}
            onDragLeave={() => setDragging(false)}
            onDrop={(e) => { if (!uploading) handleDrop(e); }}
            sx={{
              border: '2px dashed',
              borderColor: dragging ? 'primary.main' : 'divider',
              bgcolor: dragging ? 'action.hover' : 'background.default',
              borderRadius: 2,
              p: 3,
              textAlign: 'center',
              cursor: uploading ? 'default' : 'pointer',
              transition: 'all 0.2s',
              '&:hover': {
                borderColor: uploading ? 'divider' : 'text.primary',
                bgcolor: uploading ? 'background.default' : 'action.hover'
              }
            }}
          >
            <input type="file" accept="image/*" multiple hidden id="img-upload" onChange={handleFileChange} disabled={uploading} />
            <label htmlFor="img-upload" style={{ width: '100%', display: 'block', cursor: uploading ? 'default' : 'pointer' }}>
              <UploadRoundedIcon sx={{ fontSize: 40, color: dragging ? "primary.main" : "text.secondary", mb: 1 }} />
              <Typography variant="body2" fontWeight={500}>Bilder hier ablegen</Typography>
              <Typography variant="caption" color="text.secondary">oder klicken zum Auswählen</Typography>
            </label>
          </Box>

          {files.length > 0 && (
            <Paper variant="outlined" sx={{ maxHeight: 200, overflow: 'auto', borderRadius: 2 }}>
              <List dense>
                {files.map((f, idx) => (
                  <ListItem key={idx} secondaryAction={
                    <IconButton edge="end" size="small" onClick={() => removeFile(idx)} disabled={uploading}>
                      <DeleteOutlineRoundedIcon fontSize="small" />
                    </IconButton>
                  }>
                    <InsertPhotoRoundedIcon sx={{ mr: 1, color: 'primary.main' }} fontSize="small" />
                    <ListItemText
                      primary={f.name}
                      secondary={`${(f.size / (1024 * 1024)).toFixed(2)} MB`}
                      primaryTypographyProps={{ variant: 'body2', noWrap: true, sx: { maxWidth: 250 } }}
                    />
                  </ListItem>
                ))}
              </List>
            </Paper>
          )}

          <TextField
            id="media-title"
            name="title"
            label="Titel für alle Bilder (Optional)"
            size="small"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            disabled={uploading}
          />
          <TextField
            id="media-description"
            name="description"
            label="Beschreibung für alle Bilder (Optional)"
            size="small"
            multiline
            rows={2}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            disabled={uploading}
          />
        </Stack>
      </DialogContent>
      <DialogActions sx={{ p: 3 }}>
        <Button onClick={onClose} color="inherit" disabled={uploading}>Abbrechen</Button>
        <Button variant="contained" onClick={handleSubmit} disabled={files.length === 0 || !eventId || uploading}>
          {uploading ? (
            <Stack direction="row" spacing={1} alignItems="center">
              <CircularProgress size={20} color="inherit" />
              <span>{Math.round(progress)}%</span>
            </Stack>
          ) : (
            `Hochladen ${files.length > 0 ? `(${files.length})` : ""}`
          )}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function CreateEventDialog({ open, onClose, onSuccess }: {
  open: boolean;
  onClose: () => void;
  onSuccess: (id: string) => void;
}) {
  const [title, setTitle] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [dragging, setDragging] = useState(false);

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault(); setDragging(false);
    if (e.dataTransfer.files?.[0]) setFile(e.dataTransfer.files[0]);
  };

  const handleSubmit = async () => {
    setLoading(true);
    setError("");
    try {
      const formData = new FormData();
      formData.append("title", title);
      if (file) {
        formData.append("file", file);
      }

      const res = await client.request({
        method: 'POST',
        url: '/events',
        body: formData,
        bodySerializer: null,
        headers: {
          "Content-Type": null
        }
      });

      if (res.error) throw res.error;

      const newEvent = res.data as { id: string };
      onSuccess(newEvent.id);
      onClose();
      setTitle("");
      setFile(null);
    } catch (e: unknown) {
      setError(getFriendlyErrorMessage(e));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth PaperProps={{ sx: { borderRadius: 3 } }}>
      <DialogTitle sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h6" fontWeight={700}>Neues Event</Typography>
        <IconButton onClick={onClose}><CloseIcon /></IconButton>
      </DialogTitle>
      <DialogContent sx={{ pt: 1 }}>
        <Stack spacing={3} mt={1}>
          {error && <Alert severity="error">{error}</Alert>}
          <TextField
            id="create-event-title"
            name="title"
            autoFocus
            label="Event Titel"
            fullWidth
            value={title}
            onChange={(e) => setTitle(e.target.value)}
          />
          <Box
            onDragOver={(e) => { e.preventDefault(); setDragging(true); }}
            onDragLeave={() => setDragging(false)}
            onDrop={handleDrop}
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
                borderColor: 'text.primary',
                bgcolor: 'action.hover'
              }
            }}
          >
            <input type="file" accept="image/*" hidden id="event-cover-upload" onChange={(e) => setFile(e.target.files?.[0] || null)} />
            {!file ? (
              <label htmlFor="event-cover-upload" style={{ width: '100%', display: 'block', cursor: 'pointer' }}>
                <CloudUploadIcon sx={{ fontSize: 40, color: dragging ? "primary.main" : "text.secondary", mb: 1 }} />
                <Typography variant="body2" fontWeight={500}>Vorschaubild hier ablegen</Typography>
                <Typography variant="caption" color="text.secondary">oder klicken (Optional)</Typography>
              </label>
            ) : (
              <Stack direction="row" alignItems="center" justifyContent="center" spacing={1}>
                <InsertPhotoRoundedIcon color="primary" />
                <Typography noWrap sx={{ maxWidth: 200, fontWeight: 500 }}>{file.name}</Typography>
                <IconButton size="small" color="error" onClick={(e) => { e.preventDefault(); setFile(null); }}>
                  <DeleteOutlineRoundedIcon />
                </IconButton>
              </Stack>
            )}
          </Box>
        </Stack>
      </DialogContent>
      <DialogActions sx={{ p: 3 }}>
        <Button onClick={onClose} color="inherit">Abbrechen</Button>
        <Button onClick={handleSubmit} variant="contained" disabled={!title || loading}>
          {loading ? <CircularProgress size={24} /> : "Erstellen"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
