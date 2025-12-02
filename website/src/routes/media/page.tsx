import React, { useState, useEffect, useCallback } from "react";
import {
  Container, Typography, Card, CardMedia, Box, Dialog, DialogTitle, DialogContent,
  DialogActions, IconButton, ImageList, ImageListItem, ImageListItemBar, Pagination,
  Button, Stack, TextField, MenuItem, Snackbar, Alert, CircularProgress, Tooltip, Paper
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";
import ArrowBackIosNewIcon from "@mui/icons-material/ArrowBackIosNew";
import ArrowForwardIosIcon from "@mui/icons-material/ArrowForwardIos";
import CloudUploadIcon from "@mui/icons-material/CloudUpload";
import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteOutlineRoundedIcon from '@mui/icons-material/DeleteOutlineRounded';
import InsertPhotoRoundedIcon from '@mui/icons-material/InsertPhotoRounded';
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import LinkRoundedIcon from "@mui/icons-material/LinkRounded";
import { alpha } from "@mui/material/styles";

import { useAuth } from "@lib/auth";
import { Sidebar } from "@components/layout";
import { client } from "@lib/api/client.gen";
import { getAuthCsrf } from "@lib/api";

type EventItem = { id: number; title: string; created_at: string; cover_path?: string };
type MediaItem = { id: string; event_id: number; title: string; description: string; mime_type: string; uploaded_at: string };

const IMAGES_PER_PAGE = 10;

const pic = (seed: number) => `https://picsum.photos/seed/${seed}/300/200`;

export default function Galerie() {
  const { user } = useAuth();
  const [events, setEvents] = useState<EventItem[]>([]);
  const [selectedEventId, setSelectedEventId] = useState<number | null>(null);
  const [media, setMedia] = useState<MediaItem[]>([]);
  const [page, setPage] = useState(1);
  const [loadingMedia, setLoadingMedia] = useState(false);

  const [uploadOpen, setUploadOpen] = useState(false);
  const [createEventOpen, setCreateEventOpen] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  
  const [lightboxOpen, setLightboxOpen] = useState(false);
  const [lightboxIndex, setLightboxIndex] = useState(0);

  const canUpload = user?.role === "admin" || user?.role === "editor";

  const fetchEvents = async () => {
    try {
      const res = await client.request({ method: 'GET', url: '/events' });
      if (res.data) {
        const list = res.data as EventItem[];
        setEvents(list);
        if (list.length > 0 && !selectedEventId) {
          setSelectedEventId(list[0].id);
        }
      }
    } catch (e) {
      console.error("Fehler beim Laden der Events", e);
    }
  };

  useEffect(() => {
    fetchEvents();
  }, []);

  useEffect(() => {
    if (selectedEventId) {
      setLoadingMedia(true);
      client.request({ method: 'GET', url: '/media', query: { event_id: selectedEventId } })
        .then(res => {
          setMedia(res.data as MediaItem[] || []);
          setPage(1);
        })
        .finally(() => setLoadingMedia(false));
    }
  }, [selectedEventId]);

  const getImageUrl = (id: string) => `/api/media/${id}/file`;
  const getEventCoverUrl = (ev: EventItem) => ev.cover_path ? `/api/events/${ev.id}/cover` : pic(ev.id);

  const pageCount = Math.max(Math.ceil(media.length / IMAGES_PER_PAGE), 1);
  const displayedMedia = media.slice((page - 1) * IMAGES_PER_PAGE, page * IMAGES_PER_PAGE);
  const currentImage = displayedMedia[lightboxIndex];

  const openLightbox = (idx: number) => {
    setLightboxIndex(idx);
    setLightboxOpen(true);
  };
  
  const closeLightbox = useCallback(() => setLightboxOpen(false), []);
  
  const nextImage = useCallback(() => {
    setLightboxIndex((i) => (i + 1) % displayedMedia.length);
  }, [displayedMedia.length]);
  
  const prevImage = useCallback(() => {
    setLightboxIndex((i) => (i - 1 + displayedMedia.length) % displayedMedia.length);
  }, [displayedMedia.length]);

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
    const url = window.location.origin + getImageUrl(currentImage.id);
    navigator.clipboard.writeText(url).then(() => {
      setSuccessMessage("Link in die Zwischenablage kopiert!");
    });
  };

  const handleEditImage = () => {
    console.log("Edit image:", currentImage);
  };

  return (
    <Sidebar user={user} title="Galerie">
      <Container maxWidth="xl" sx={{ mt: 5, mb: 10 }}>
        
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
                startIcon={<CloudUploadIcon />}
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
                  "-ms-overflow-style": "none",
                  "scrollbar-width": "none",
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
                      onClick={() => { setSelectedEventId(ev.id); setPage(1); }}
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
                  image={getImageUrl(item.id)}
                  alt={item.title || "Bild"}
                  sx={{ width: "100%", aspectRatio: "3/2", objectFit: "cover" }}
                />
              </Card>
            ))}
          </Box>
        )}

        {media.length > IMAGES_PER_PAGE && (
          <Box sx={{ mt: 5, display: "flex", justifyContent: "center" }}>
            <Pagination
              count={pageCount}
              page={page}
              onChange={(_, p) => setPage(p)}
              showFirstButton
              showLastButton
              color="primary"
              shape="rounded"
            />
          </Box>
        )}

        {canUpload && (
          <UploadDialog 
            open={uploadOpen} 
            onClose={() => setUploadOpen(false)} 
            events={events}
            onSuccess={() => { 
              fetchEvents(); 
              if(selectedEventId) {
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
            onSuccess={() => { fetchEvents(); }}
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

      </Container>
    </Sidebar>
  );
}

function getFriendlyErrorMessage(error: any): string {
  const msg = error?.message || error?.error || (typeof error === 'string' ? error : "");
  const errorMap: Record<string, string> = {
    "File too large": "Die Datei ist zu groß (maximal 50MB erlaubt).",
    "File upload failed": "Der Upload ist fehlgeschlagen. Bitte versuche es erneut.",
    "Invalid event_id": "Das ausgewählte Event ist ungültig.",
    "Missing metadata": "Es fehlen notwendige Daten für den Upload.",
    "invalid_csrf": "Deine Sitzung ist abgelaufen. Bitte lade die Seite neu.",
    "bad_request": "Ungültige Anfrage.",
    "Upload failed": "Der Upload ist fehlgeschlagen. Bitte prüfe Format und Größe der Datei."
  };
  return errorMap[msg] || `Ein Fehler ist aufgetreten: ${msg || "Unbekannter Fehler"}`;
}

function UploadDialog({ open, onClose, events, onSuccess, onCreateEvent }: any) {
  const [file, setFile] = useState<File | null>(null);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [eventId, setEventId] = useState<number | "">("");
  const [dragging, setDragging] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState("");

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault(); setDragging(false);
    if (e.dataTransfer.files?.[0]) setFile(e.dataTransfer.files[0]);
  };

  const handleSubmit = async () => {
    if (!file || !eventId) return;
    setUploading(true);
    setError("");
    try {
      const { data } = await getAuthCsrf();
      const token = data?.csrf;
      
      const formData = new FormData();
      formData.append("file", file);
      formData.append("event_id", String(eventId));
      if(title) formData.append("title", title);
      if(description) formData.append("description", description);

      const res = await client.request({
        method: 'POST',
        url: '/media',
        body: formData,
        bodySerializer: null,
        // @ts-ignore
        headers: { 
            "X-CSRF-Token": token || "",
            "Content-Type": null 
        }
      });

      if (res.error) throw res.error;
      
      onSuccess();
      onClose();
      setFile(null);
      setTitle("");
      setDescription("");
    } catch (e: any) {
      console.error("Upload error:", e);
      setError(getFriendlyErrorMessage(e));
    } finally {
      setUploading(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm" PaperProps={{ sx: { borderRadius: 3 } }}>
      <DialogTitle sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h6" fontWeight={700}>Bild hochladen</Typography>
        <IconButton onClick={onClose}><CloseIcon /></IconButton>
      </DialogTitle>
      <DialogContent sx={{ pt: 1 }}>
        <Stack spacing={3} mt={1}>
          {error && <Alert severity="error">{error}</Alert>}
          
          <Stack direction="row" spacing={1} alignItems="stretch">
            <TextField
              select
              fullWidth
              label="Event auswählen"
              value={eventId}
              onChange={(e) => setEventId(Number(e.target.value))}
              size="small"
            >
              {events.map((e: any) => <MenuItem key={e.id} value={e.id}>{e.title}</MenuItem>)}
            </TextField>
            <Tooltip title="Neues Event anlegen">
              <Button 
                variant="outlined" 
                sx={{ minWidth: 40, px: 0, borderRadius: 2 }} 
                onClick={onCreateEvent}
              >
                <AddRoundedIcon />
              </Button>
            </Tooltip>
          </Stack>

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
            <input type="file" accept="image/*" hidden id="img-upload" onChange={(e) => setFile(e.target.files?.[0] || null)} />
            {!file ? (
              <label htmlFor="img-upload" style={{ width: '100%', display: 'block', cursor: 'pointer' }}>
                <CloudUploadIcon sx={{ fontSize: 40, color: dragging ? "primary.main" : "text.secondary", mb: 1 }} />
                <Typography variant="body2" fontWeight={500}>Bild hier ablegen</Typography>
                <Typography variant="caption" color="text.secondary">oder klicken zum Auswählen</Typography>
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

          <TextField label="Titel (Optional)" size="small" value={title} onChange={(e) => setTitle(e.target.value)} />
          <TextField label="Beschreibung (Optional)" size="small" multiline rows={2} value={description} onChange={(e) => setDescription(e.target.value)} />
        </Stack>
      </DialogContent>
      <DialogActions sx={{ p: 3 }}>
        <Button onClick={onClose} color="inherit">Abbrechen</Button>
        <Button variant="contained" onClick={handleSubmit} disabled={!file || !eventId || uploading}>
          {uploading ? <CircularProgress size={24} /> : "Hochladen"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function CreateEventDialog({ open, onClose, onSuccess }: any) {
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
      const { data } = await getAuthCsrf();
      const token = data?.csrf;
      
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
        // @ts-ignore
        headers: { "X-CSRF-Token": token || "", "Content-Type": null }
      });

      if (res.error) throw res.error;

      onSuccess();
      onClose();
      setTitle("");
      setFile(null);
    } catch (e: any) {
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
