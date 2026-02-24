'use client';

import React, { useState, useEffect, useCallback } from "react";
import Typography from "@mui/material/Typography";
import Card from "@mui/material/Card";
import CardMedia from "@mui/material/CardMedia";
import Box from "@mui/material/Box";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import IconButton from "@mui/material/IconButton";
import ImageList from "@mui/material/ImageList";
import ImageListItem from "@mui/material/ImageListItem";
import ImageListItemBar from "@mui/material/ImageListItemBar";
import Pagination from "@mui/material/Pagination";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import MenuItem from "@mui/material/MenuItem";
import Snackbar from "@mui/material/Snackbar";
import Alert from "@mui/material/Alert";
import CircularProgress from "@mui/material/CircularProgress";
import Tooltip from "@mui/material/Tooltip";
import Paper from "@mui/material/Paper";
import LinearProgress from "@mui/material/LinearProgress";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import Skeleton from "@mui/material/Skeleton";
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
import { useRouter } from "next/navigation";

import { useAuth } from "@lib/auth";
import { Sidebar } from "@components/layout";
import { getEvents, getEventsByEventIdMedia, postEvents, postEventsByEventIdMedia } from "@lib/api";
import type { DtoEventResponse as EventItem, DtoMediaResponse as MediaItem } from "@lib/api";
import { getSizedImageUrl, getImageSrcSet } from "@lib/images";


const IMAGES_PER_PAGE = 10;

export default function Galerie({ params }: { params: Promise<{ slug?: string[] }> }) {
  const { slug } = React.use(params);
  const urlEventId = slug?.[0];
  const urlMediaId = slug?.[1];

  const { user } = useAuth();

  const [events, setEvents] = useState<EventItem[]>([]);
  const [selectedEventId, setSelectedEventId] = useState<string | null>(null);
  const [media, setMedia] = useState<MediaItem[]>([]);
  const [page, setPage] = useState(1);
  const [loadingMedia, setLoadingMedia] = useState(false);
  const router = useRouter();

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
      const { data } = await getEvents();
      if (data) {
        setEvents(data);

        if (!urlEventId && data.length > 0) {
          router.replace(`/events/${data[0].id}`);
        }
      }
    } catch (e) {
      console.error("Fehler beim Laden der Events", e);
    }
  }, [urlEventId, router]);

  useEffect(() => {
    void fetchEvents();
  }, [fetchEvents]);

  useEffect(() => {
    if (!selectedEventId) {
      setMedia([]);
      return;
    }

    const fetchMediaData = async () => {
      setLoadingMedia(true);
      try {
        const { data } = await getEventsByEventIdMedia({
          path: { eventId: selectedEventId }
        });
        setMedia(data || []);
      } catch (e: unknown) {
        console.error("Fehler beim Laden der Medien", e);
      } finally {
        setLoadingMedia(false);
      }
    };
    void fetchMediaData();
  }, [selectedEventId]);

  useEffect(() => {
    if (urlMediaId && media.length > 0 && selectedEventId === urlEventId) {
      const index = media.findIndex(m => String(m.id) === urlMediaId);
      if (index !== -1) {
        setPage(Math.floor(index / IMAGES_PER_PAGE) + 1);
        setLightboxIndex(index);
        setLightboxOpen(true);
      }
    } else if (!urlMediaId) {
      setLightboxOpen(false);
    }
  }, [urlMediaId, media, selectedEventId, urlEventId]);

  const getPreviewUrl = (id: string, size: number | string = 400) => getSizedImageUrl(`/api/v1/media/${id}/preview`, size as any);
  const getEventCoverUrl = (ev: EventItem, size: number | string = 400) => ev.cover_path ? getSizedImageUrl(`/api/v1/events/${ev.id}/cover`, size as any) : undefined;
  const getEventCoverSrcSet = (ev: EventItem) => ev.cover_path ? getImageSrcSet(`/api/v1/events/${ev.id}/cover`) : undefined;
  const getMediaSrcSet = (id: string) => getImageSrcSet(`/api/v1/media/${id}/preview`);

  const pageCount = Math.max(Math.ceil(media.length / IMAGES_PER_PAGE), 1);
  const displayedMedia = media.slice((page - 1) * IMAGES_PER_PAGE, page * IMAGES_PER_PAGE);
  const currentImage = media[lightboxIndex];

  const openLightbox = (idx: number) => {
    const absoluteIdx = (page - 1) * IMAGES_PER_PAGE + idx;
    setLightboxIndex(absoluteIdx);
    setLightboxOpen(true);
    const img = media[absoluteIdx];
    if (img) router.replace(`/events/${selectedEventId}/${img.id}`);
  };

  const closeLightbox = useCallback(() => {
    setLightboxOpen(false);
    router.replace(`/events/${selectedEventId}`);
  }, [selectedEventId, router]);

  const nextImage = useCallback(() => {
    if (media.length === 0) return;
    const nextIdx = (lightboxIndex + 1) % media.length;
    setLightboxIndex(nextIdx);
    setPage(Math.floor(nextIdx / IMAGES_PER_PAGE) + 1);
    const img = media[nextIdx];
    if (img) router.replace(`/events/${selectedEventId}/${img.id}`);
  }, [lightboxIndex, selectedEventId, router, media]);

  const prevImage = useCallback(() => {
    if (media.length === 0) return;
    const prevIdx = (lightboxIndex - 1 + media.length) % media.length;
    setLightboxIndex(prevIdx);
    setPage(Math.floor(prevIdx / IMAGES_PER_PAGE) + 1);
    const img = media[prevIdx];
    if (img) router.replace(`/events/${selectedEventId}/${img.id}`);
  }, [lightboxIndex, selectedEventId, router, media]);

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
    const url = window.location.origin + `/events/${selectedEventId}/${currentImage.id}`;
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
              Events & Galerie
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
                  const selected = String(ev.id) === selectedEventId;

                    return (
                      <ImageListItem
                        key={String(ev.id)}
                        onClick={() => {
                          setMedia([]); 
                          setPage(1);
                          router.push(`/events/${ev.id}`);
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
                          src={getEventCoverUrl(ev, 400)}
                          srcSet={getEventCoverSrcSet(ev)}
                          sizes="160px"
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
            ) : events.length === 0 && selectedEventId === null ? (
              <Stack direction="row" spacing={3} sx={{ p: 3, pt: 1, overflow: 'hidden' }}>
                {[1, 2, 3, 4, 5].map((i) => (
                  <Box key={i} sx={{ width: 160, flexShrink: 0 }}>
                    <Skeleton variant="rectangular" height={100} sx={{ borderRadius: 2, mb: 1 }} />
                    <Skeleton variant="text" width="80%" />
                  </Box>
                ))}
              </Stack>
            ) : (
              <Typography color="text.secondary" sx={{ py: 2, px: 2 }}>Keine Events vorhanden.</Typography>
            )}
          </Box>
        </Paper>


        {loadingMedia ? (
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: { xs: "repeat(2, 1fr)", sm: "repeat(3, 1fr)", md: "repeat(4, 1fr)", lg: "repeat(5, 1fr)" },
              gap: 2
            }}
          >
            {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
              <Skeleton
                key={i}
                variant="rectangular"
                sx={{
                  borderRadius: 2,
                  aspectRatio: "3/2",
                  width: "100%",
                  height: "auto"
                }}
              />
            ))}
          </Box>
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
                key={String(item.id)}
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
                  image={getPreviewUrl(String(item.id), 600)}
                  srcSet={getMediaSrcSet(String(item.id))}
                  sizes="(max-width: 600px) 50vw, (max-width: 900px) 33vw, (max-width: 1200px) 25vw, 200px"
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
                src={getPreviewUrl(String(currentImage.id), 1600)}
                srcSet={getMediaSrcSet(String(currentImage.id))}
                sizes="100vw"
                alt={currentImage.title}
                sx={{
                  display: "block",
                  maxWidth: "100%",
                  maxHeight: "80vh",
                  objectFit: "contain",
                }}
              />
            )}

            {media.length > 1 && (
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
  const msg = typeof error === 'string' ? error : (error as Record<string, string>)?.message || (error as Record<string, string>)?.error || "";
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
        const { error: apiError } = await postEventsByEventIdMedia({
          path: { eventId: String(eventId) },
          body: {
            file: file,
            title: title || undefined,
            description: description || undefined
          }
        });

        if (apiError) throw apiError;
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
                events.map((e) => <MenuItem key={String(e.id)} value={String(e.id)}>{e.title}</MenuItem>)
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
      const { data, error: apiError } = await postEvents({
        body: {
          title,
          file: file || undefined
        }
      });

      if (apiError) throw apiError;

      if (data && data.id) {
        onSuccess(String(data.id));
        onClose();
        setTitle("");
        setFile(null);
      }
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
