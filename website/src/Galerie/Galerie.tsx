import React from "react";
import { useAuth } from "../AuthContext";
import {
  Container,
  Typography,
  Card,
  CardMedia,
  Box,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  IconButton,
  ImageList,
  ImageListItem,
  ImageListItemBar,
  Pagination,
  Button,
  Stack,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Snackbar,
  Alert,
  SelectChangeEvent,
  TextField,
  alpha, // Import alpha für Transparenz
  Theme,
  // CircularProgress entfernt (wird nicht mehr gebraucht)
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";
import ArrowBackIosNewIcon from "@mui/icons-material/ArrowBackIosNew";
import ArrowForwardIosIcon from "@mui/icons-material/ArrowForwardIos";
import CloudUploadIcon from "@mui/icons-material/CloudUpload";

// Typ-Definitionen bleiben gleich
type EventItem = { id: number; title: string; src: string };
type Bild = { id: number; title: string; thumb: string; full: string };
type UploadEntry = { file: File; title: string };

const getImageTitle = (image?: Bild) => {
  if (!image) return "";
  const trimmed = image.title?.trim();
  if (trimmed) return trimmed;
  const candidate = image.full || image.thumb;
  return candidate.split("/").filter(Boolean).pop() ?? "Bild";
};

const defaultTitleForFile = (file: File) => {
  const name = file.name;
  const withoutExtension = name.replace(/\.[^/.]+$/, "");
  return withoutExtension || name;
};

// --- HILFSFUNKTION FÜR BEISPIELBILDER ---
const pic = (seed: number, w: number, h: number) =>
  `https://picsum.photos/seed/${seed}/${w}/${h}`;

// --- HIER SIND DEINE LOKALEN BILDER & BEISPIELBILDER ---

const events: EventItem[] = [
  // --- EVENT 1 (DEIN ORDNER 1) ---
  // Ersetze DEIN_ORDNER_1 und den Dateinamen
  { id: 1, title: "LAN-Party 2023", src: "/LAN-Party 2023/1.jpg" },

  // --- EVENT 2 (DEIN ORDNER 2) ---
  // Ersetze DEIN_ORDNER_2 und den Dateinamen
  { id: 2, title: "LAN-Party 2024", src: "/LAN-Party 2024/2024-11-29_23.07.08-_PHI1930-klaushardt.com.jpg" },

  // --- EVENTS 3-6 (BEISPIELBILDER) ---
  { id: 3, title: "Weinprobe", src: pic(103, 400, 600) },
  { id: 4, title: "Kulturabend", src: pic(104, 400, 600) },
  { id: 5, title: "Weihnachtsfeier", src: pic(105, 400, 600) },
  { id: 6, title: "Sporttag", src: pic(106, 400, 600) },
];

const IMAGES_PER_PAGE = 20;

const bilderByEvent: Record<number, Bild[]> = {
  // --- EVENT 1 (DEIN ORDNER 1) ---
  // Ersetze DEIN_ORDNER_1 und die Dateinamen
  1: [
    { id: 101, title: "", thumb: "/LAN-Party 2023/1.jpg", full: "/LAN-Party 2023/1.jpg" },
    { id: 102, title: "", thumb: "/LAN-Party 2023/1688fad1-29d5-48c3-9167-9af3553ef10f.jpg", full: "/LAN-Party 2023/1688fad1-29d5-48c3-9167-9af3553ef10f.jpg" },
    { id: 103, title: "", thumb: "/LAN-Party 2023/ad23f7da-c6cd-4a96-9fce-8b5d0ccd8cf8.jpg", full: "/LAN-Party 2023/ad23f7da-c6cd-4a96-9fce-8b5d0ccd8cf8.jpg" },
    { id: 104, title: "", thumb: "/LAN-Party 2023/b4818f1d-72e6-45ab-8f65-17e2c84b9c77.jpg", full: "/LAN-Party 2023/b4818f1d-72e6-45ab-8f65-17e2c84b9c77.jpg" },
    { id: 105, title: "", thumb: "/LAN-Party 2023/c7cdd3e2-6b75-4d06-805d-1f761b359dd7.jpg", full: "/LAN-Party 2023/c7cdd3e2-6b75-4d06-805d-1f761b359dd7.jpg" },
    { id: 106, title: "", thumb: "/LAN-Party 2023/DSC00061.jpg", full: "/LAN-Party 2023/DSC00061.jpg" },
    { id: 107, title: "", thumb: "/LAN-Party 2023/DSC00097.jpg", full: "/LAN-Party 2023/DSC00097.jpg" },
    { id: 108, title: "", thumb: "/LAN-Party 2023/DSC00114.jpg", full: "/LAN-Party 2023/DSC00114.jpg" },
    { id: 109, title: "", thumb: "/LAN-Party 2023/DSC00132.jpg", full: "/LAN-Party 2023/DSC00132.jpg" },
    { id: 110, title: "", thumb: "/LAN-Party 2023/DSC00140.jpg", full: "/LAN-Party 2023/DSC00140.jpg" },
    { id: 111, title: "", thumb: "/LAN-Party 2023/DSC00146.jpg", full: "/LAN-Party 2023/DSC00146.jpg" },
    { id: 112, title: "", thumb: "/LAN-Party 2023/DSC00170.jpg", full: "/LAN-Party 2023/DSC00170.jpg" },
    { id: 113, title: "", thumb: "/LAN-Party 2023/DSC00173.jpg", full: "/LAN-Party 2023/DSC00173.jpg" },

    // ...
  ],
  
  // --- EVENT 2 (DEIN ORDNER 2) ---
  // Ersetze DEIN_ORDNER_2 und die Dateinamen
  2: [
    { id: 201, title: "", thumb: "/LAN-Party 2024/1.jpg", full: "/LAN-Party 2024/1.jpg" },
    { id: 202, title: "", thumb: "/LAN-Party 2024/2.jpg", full: "/LAN-Party 2024/2.jpg" },
    { id: 203, title: "", thumb: "/LAN-Party 2024/3.jpg", full: "/LAN-Party 2024/3.jpg" },
    { id: 204, title: "", thumb: "/LAN-Party 2024/4.jpg", full: "/LAN-Party 2024/4.jpg" },
    { id: 205, title: "", thumb: "/LAN-Party 2024/5.jpg", full: "/LAN-Party 2024/5.jpg" },
    { id: 206, title: "", thumb: "/LAN-Party 2024/6.jpg", full: "/LAN-Party 2024/6.jpg" },
    { id: 207, title: "", thumb: "/LAN-Party 2024/7.jpg", full: "/LAN-Party 2024/7.jpg" },
    { id: 208, title: "", thumb: "/LAN-Party 2024/8.jpg", full: "/LAN-Party 2024/8.jpg" },
    { id: 209, title: "", thumb: "/LAN-Party 2024/9.jpg", full: "/LAN-Party 2024/9.jpg" },
    { id: 210, title: "", thumb: "/LAN-Party 2024/10.jpg", full: "/LAN-Party 2024/10.jpg" },
    { id: 211, title: "", thumb: "/LAN-Party 2024/11.jpg", full: "/LAN-Party 2024/11.jpg" },
    { id: 212, title: "", thumb: "/LAN-Party 2024/12.jpg", full: "/LAN-Party 2024/12.jpg" },
    { id: 213, title: "", thumb: "/LAN-Party 2024/13.jpg", full: "/LAN-Party 2024/13.jpg" },
    { id: 214, title: "", thumb: "/LAN-Party 2024/14.jpg", full: "/LAN-Party 2024/14.jpg" },
    // ...
  ],

  // Die restlichen Events werden automatisch mit Beispielbildern befüllt
};

// --- EVENTS 3-6 (BEISPIELBILDER) ---
// Füllt `bilderByEvent` für alle Events > 2 automatisch mit picsum-Bildern
events.forEach(event => {
  if (event.id > 2) { // <-- Geändert auf 2
    bilderByEvent[event.id] = Array.from({ length: 30 }).map((_, i) => {
      const seed = event.id * 100 + i;
      return {
        id: seed,
        title: `${event.title} Bild ${i + 1}`,
        thumb: pic(seed, 600, 400),
        full: pic(seed, 1600, 1066),
      };
    });
  }
});
// --- ENDE BILDER ---


export default function Galerie() {
  // --- State (vereinfacht) ---
  const [selectedEventId, setSelectedEventId] = React.useState<number | null>(
    null
  );
  const [page, setPage] = React.useState(1);
  const [customImages, setCustomImages] = React.useState<Record<number, Bild[]>>(
    {}
  );
  const [customEvents, setCustomEvents] = React.useState<EventItem[]>([]);
  const [uploadOpen, setUploadOpen] = React.useState(false);
  const [selectedUploadEventId, setSelectedUploadEventId] = React.useState<
    number | "" | "new"
  >("");
  const [uploadEntries, setUploadEntries] = React.useState<UploadEntry[]>([]);
  const [uploadError, setUploadError] = React.useState<string | null>(null);
  const [uploading, setUploading] = React.useState(false);
  const [successMessage, setSuccessMessage] = React.useState<string | null>(null);
  const [newEventTitle, setNewEventTitle] = React.useState("");
  const [newEventCover, setNewEventCover] = React.useState<string | null>(null);
  const [newEventCoverName, setNewEventCoverName] = React.useState("");
  const [newEventError, setNewEventError] = React.useState<string | null>(null);

  // --- State für Lightbox ---
  const [open, setOpen] = React.useState(false);
  const [index, setIndex] = React.useState(0);
  const { user } = useAuth();
  const canUpload = user?.role === "admin" || user?.role === "editor";
  const allEvents = React.useMemo(
    () => [...customEvents, ...events],
    [customEvents]
  );

  // Abgeleiteter State: Die Thumbnails für das ausgewählte Event
  const thumbs = selectedEventId
    ? [
        ...(bilderByEvent[selectedEventId] ?? []),
        ...(customImages[selectedEventId] ?? []),
      ]
    : [];
  const pageCount = Math.max(Math.ceil(thumbs.length / IMAGES_PER_PAGE), 1);
  const startIndex = (page - 1) * IMAGES_PER_PAGE;
  const paginatedThumbs = thumbs.slice(
    startIndex,
    startIndex + IMAGES_PER_PAGE
  );
  const currentImage = thumbs[index];
  const currentImageTitle = getImageTitle(currentImage);

  React.useEffect(() => {
    if (page > pageCount) {
      setPage(pageCount);
    }
  }, [page, pageCount]);

  const handlePageChange = (
    _: React.ChangeEvent<unknown>,
    value: number
  ) => {
    setPage(value);
  };

  const handleOpenUploadDialog = () => {
    setUploadError(null);
    setUploadEntries([]);
    setNewEventTitle("");
    setNewEventCover(null);
    setNewEventCoverName("");
    setNewEventError(null);
    setSelectedUploadEventId(
      selectedEventId ?? allEvents[0]?.id ?? ""
    );
    setUploadOpen(true);
  };

  const handleCloseUploadDialog = () => {
    if (uploading) return;
    setUploadOpen(false);
  };

  const handleFilesChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (!files) {
      setUploadEntries([]);
      return;
    }
    const entries = Array.from(files).map((file) => ({
      file,
      title: defaultTitleForFile(file),
    }));
    setUploadEntries(entries);
    setUploadError(null);
  };

  const readFileAsDataUrl = React.useCallback((file: File) => {
    return new Promise<string>((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => resolve(reader.result as string);
      reader.onerror = () =>
        reject(new Error("Datei konnte nicht gelesen werden."));
      reader.readAsDataURL(file);
    });
  }, []);

  const handleUploadImages = async () => {
    if (!selectedUploadEventId) {
      setUploadError("Bitte ein Event auswählen.");
      return;
    }
    if (uploadEntries.length === 0) {
      setUploadError("Bitte mindestens ein Bild auswählen.");
      return;
    }
    setUploading(true);
    setUploadError(null);
    try {
      let eventId: number | null = null;
      let createdTitle = "";
      if (selectedUploadEventId === "new") {
        const title = newEventTitle.trim();
        if (!title) {
          setNewEventError("Bitte einen Eventnamen angeben.");
          setUploading(false);
          return;
        }
        eventId = Date.now();
        const newEvent: EventItem = {
          id: eventId,
          title,
          src: newEventCover ?? pic(eventId, 400, 600),
        };
        setCustomEvents((prev) => [newEvent, ...prev]);
        createdTitle = title;
      } else {
        eventId = selectedUploadEventId;
      }
      const filesArray = await Promise.all(
        uploadEntries.map(async (entry, idx) => {
          const dataUrl = await readFileAsDataUrl(entry.file);
          return {
            id: Date.now() + idx,
            title: entry.title || entry.file.name,
            thumb: dataUrl,
            full: dataUrl,
          };
        })
      );
      if (eventId === null) {
        setUploadError("Es ist ein Fehler aufgetreten.");
        return;
      }
      setCustomImages((prev) => ({
        ...prev,
        [eventId]: [...(prev[eventId] ?? []), ...filesArray],
      }));
      setUploadOpen(false);
      setUploadEntries([]);
      setSuccessMessage(
        `${filesArray.length} Bild${
          filesArray.length > 1 ? "er" : ""
        } ${createdTitle ? `für "${createdTitle}" ` : ""}hinzugefügt (nur lokal sichtbar).`
      );
      setSelectedEventId(eventId);
      setSelectedUploadEventId(eventId);
      setPage(1);
      setIndex(0);
      setNewEventTitle("");
      setNewEventCover(null);
      setNewEventCoverName("");
      setNewEventError(null);
    } catch (error) {
      setUploadError(
        error instanceof Error
          ? error.message
          : "Beim Lesen der Dateien ist ein Fehler aufgetreten."
      );
    } finally {
      setUploading(false);
    }
  };

  const handleNewEventCoverChange = async (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const file = event.target.files?.[0];
    if (!file) {
      setNewEventCover(null);
      setNewEventCoverName("");
      return;
    }
    try {
      const dataUrl = await readFileAsDataUrl(file);
      setNewEventCover(dataUrl);
      setNewEventCoverName(file.name);
      setNewEventError(null);
    } catch (error) {
      setNewEventError(
        error instanceof Error
          ? error.message
          : "Das Titelbild konnte nicht gelesen werden."
      );
    }
  };

  const updateUploadEntryTitle = (index: number, value: string) => {
    setUploadEntries((prev) =>
      prev.map((entry, i) =>
        i === index ? { ...entry, title: value } : entry
      )
    );
  };


  // --- Lightbox Handlers (optimiert mit useCallback) ---
  const openLightbox = (i: number) => {
    setIndex(i);
    setOpen(true);
  };

  const closeLightbox = React.useCallback(() => {
    setOpen(false);
  }, []);

  const prev = React.useCallback(() => {
    setIndex((i) => (i - 1 + thumbs.length) % thumbs.length);
  }, [thumbs.length]);

  const next = React.useCallback(() => {
    setIndex((i) => (i + 1) % thumbs.length);
  }, [thumbs.length]);

  // Keyboard-Navigation für Lightbox
  React.useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "ArrowLeft") prev();
      if (e.key === "ArrowRight") next();
      if (e.key === "Escape") closeLightbox();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, prev, next, closeLightbox]);

  React.useEffect(() => {
    if (thumbs.length === 0) {
      setIndex(0);
      setOpen(false);
      return;
    }
    if (index >= thumbs.length) {
      setIndex(0);
    }
  }, [thumbs.length, index]);

  return (
    <Container sx={{ py: (theme: Theme) => theme.spacing(4) }}>
      <Typography variant="h4" align="center" gutterBottom>
        Galerie
      </Typography>

      {canUpload && (
        <Box
          sx={{
            display: "flex",
            justifyContent: "flex-end",
            mb: (theme: Theme) => theme.spacing(2),
          }}
        >
          <Button
            variant="contained"
            startIcon={<CloudUploadIcon />}
            onClick={handleOpenUploadDialog}
          >
            Bilder hochladen
          </Button>
        </Box>
      )}

      {/* --- Event-Auswahl als horizontale ImageList --- */}
      <Box
        sx={(theme: Theme) => ({
          mb: theme.spacing(4),
          mx: theme.spacing(-2),
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        })}
      >
        <ImageList
          sx={(theme: Theme) => ({
            display: "flex",
            flexDirection: "row",
            flexWrap: "nowrap",
            transform: "translateZ(0)",
            pb: theme.spacing(1),
            overflowX: "auto",
            "&::-webkit-scrollbar": { display: "none" },
            "-ms-overflow-style": "none",
            "scrollbar-width": "none",
            px: theme.spacing(1),
            width: "100%",
          })}
        >
          {allEvents.map((ev) => {
            const selected = ev.id === selectedEventId;
            return (
              <ImageListItem
                key={ev.id}
                onClick={() => {
                  setSelectedEventId(ev.id);
                  setIndex(0);
                  setPage(1);
                }}
                sx={(theme: Theme) => ({
                  cursor: "pointer",
                  width: theme.spacing(18.75),
                  flexShrink: 0,
                  borderRadius: theme.shape.borderRadius,
                  overflow: "hidden",
                  boxShadow: theme.shadows[selected ? 6 : 2],
                  transform: selected ? "scale(1.05)" : "scale(1)",
                  transition: theme.transitions.create([
                    "transform",
                    "box-shadow",
                    "outline",
                    "margin",
                  ]),
                  outline: selected ? "2px solid" : "none",
                  outlineOffset: "2px",
                  outlineColor: selected
                    ? theme.palette.primary.main
                    : "transparent",
                  zIndex: selected ? 2 : 1,
                  marginLeft: theme.spacing(2),
                  "&:first-of-type": {
                    marginLeft: 0,
                  },
                  "&:hover": {
                    transform: "scale(1.08)",
                    boxShadow: theme.shadows[8],
                    zIndex: 3,
                  },
                })}
              >
                <Box
                  component="img"
                  src={ev.src} // Holt sich den Pfad (entweder /DEIN_ORDNER/... oder picsum)
                  alt={ev.title}
                  loading="lazy"
                  sx={{
                    width: "100%",
                    height: (theme: Theme) => theme.spacing(20),
                    objectFit: "contain",
                    display: "block",
                  }}
                />
                <ImageListItemBar
                  title={ev.title}
                  position="below"
                  sx={(theme: Theme) => ({
                    textAlign: "center",
                    bgcolor: selected
                      ? theme.palette.primary.main
                      : theme.palette.background.paper,
                    color: selected
                      ? theme.palette.primary.contrastText
                      : theme.palette.text.primary,
                    py: theme.spacing(1),
                    "& .MuiImageListItemBar-title": {
                      fontSize: "0.8rem",
                      whiteSpace: "nowrap",
                      overflow: "hidden",
                      textOverflow: "ellipsis",
                    },
                  })}
                />
              </ImageListItem>
            );
          })}
        </ImageList>
      </Box>

      {/* --- Titel der ausgewählten Veranstaltung --- */}
      {selectedEventId && (
        <Typography
          variant="h5"
          component="h2"
          align="center"
          sx={{ mb: (theme: Theme) => theme.spacing(3) }}
        >
          {allEvents.find((ev) => ev.id === selectedEventId)?.title}
        </Typography>
      )}

      {/* --- Bilderbereich --- */}
      {!selectedEventId ? (
        <Typography
          align="center"
          color="text.secondary"
          sx={{ mt: (theme: Theme) => theme.spacing(6) }}
        >
          Bitte ein Event oben auswählen, um Bilder zu sehen.
        </Typography>
      ) : thumbs.length === 0 ? (
        <Typography
          align="center"
          color="text.secondary"
          sx={{ mt: (theme: Theme) => theme.spacing(6) }}
        >
          Für dieses Event sind noch keine Bilder hinterlegt. (Überprüfe die 'bilderByEvent'-Konstante)
        </Typography>
      ) : (
        <Box
          sx={{
            display: "grid",
            gridTemplateColumns: {
              xs: "repeat(2, 1fr)",
              sm: "repeat(3, 1fr)",
              md: "repeat(4, 1fr)",
            },
            gap: (theme: Theme) => theme.spacing(2),
          }}
        >
          {paginatedThumbs.map((b, i) => (
            <Card
              key={b.id}
              onClick={() => openLightbox(startIndex + i)}
              sx={(theme: Theme) => ({
                borderRadius: theme.shape.borderRadius,
                overflow: "hidden",
                cursor: "pointer",
                transition: theme.transitions.create([
                  "transform",
                  "box-shadow",
                ]),
                "&:hover": {
                  transform: "translateY(-3px)",
                  boxShadow: theme.shadows[6],
                },
              })}
            >
              <CardMedia
                component="img"
                image={b.thumb} // Holt sich den Pfad (entweder /DEIN_ORDNER/... oder picsum)
                alt={getImageTitle(b)}
                sx={{
                  width: "100%",
                  aspectRatio: "3/2",
                  objectFit: "contain",
                }}
              />
            </Card>
          ))}
        </Box>
      )}
      {selectedEventId && thumbs.length > IMAGES_PER_PAGE && (
        <Box
          sx={{
            mt: (theme: Theme) => theme.spacing(4),
            display: "flex",
            justifyContent: "center",
          }}
        >
          <Pagination
            count={pageCount}
            page={page}
            onChange={handlePageChange}
            showFirstButton
            showLastButton
            color="primary"
            siblingCount={0}
            variant="text"
            shape="rounded"
            sx={(theme: Theme) => ({
              "& .MuiPaginationItem-root": {
                borderRadius: "50%",
                minWidth: theme.spacing(5),
                height: theme.spacing(5),
                color: theme.palette.primary.main,
              },
              "& .MuiPaginationItem-root.Mui-selected": {
                bgcolor: theme.palette.primary.main,
                color: theme.palette.primary.contrastText,
                "&:hover": {
                  bgcolor: theme.palette.primary.dark,
                },
              },
            })}
          />
        </Box>
      )}

      {canUpload && (
        <Dialog
          open={uploadOpen}
          onClose={handleCloseUploadDialog}
          fullWidth
          maxWidth="sm"
        >
          <DialogTitle>Neue Bilder hochladen</DialogTitle>
          <DialogContent dividers>
            <Stack spacing={3}>
              <FormControl fullWidth>
                <InputLabel id="upload-event-label">Event auswählen</InputLabel>
                <Select
                  labelId="upload-event-label"
                  label="Event auswählen"
                  value={selectedUploadEventId}
                  onChange={(
                    event: SelectChangeEvent<number | "" | "new">
                  ) => {
                    const value = event.target.value;
                    if (value === "" || value === "new") {
                      setSelectedUploadEventId(value);
                      if (value !== "new") {
                        setNewEventError(null);
                      }
                    } else {
                      setSelectedUploadEventId(Number(value));
                      setNewEventError(null);
                    }
                  }}
                >
                  <MenuItem value="">
                    <em>Bitte auswählen</em>
                  </MenuItem>
                  <MenuItem value="new">+ Neues Event anlegen</MenuItem>
                  {allEvents.map((event) => (
                    <MenuItem key={event.id} value={event.id}>
                      {event.title}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>

              {selectedUploadEventId === "new" && (
                <Stack spacing={2}>
                  <TextField
                    label="Eventname"
                    value={newEventTitle}
                    onChange={(event) => {
                      setNewEventTitle(event.target.value);
                      setNewEventError(null);
                    }}
                    required
                  />
                  <Button variant="outlined" component="label">
                    Titelbild auswählen
                    <input
                      hidden
                      type="file"
                      accept="image/*"
                      onChange={handleNewEventCoverChange}
                    />
                  </Button>
                  {newEventCoverName && (
                    <Typography variant="body2" color="text.secondary">
                      Ausgewählt: {newEventCoverName}
                    </Typography>
                  )}
                  {newEventError && <Alert severity="error">{newEventError}</Alert>}
                </Stack>
              )}

              <Button
                variant="outlined"
                component="label"
                startIcon={<CloudUploadIcon />}
                disabled={uploading}
              >
                Dateien auswählen
                <input
                  hidden
                  type="file"
                  accept="image/*"
                  multiple
                  onChange={handleFilesChange}
                />
              </Button>

              {uploadEntries.length > 0 && (
                <Stack spacing={2}>
                  <Typography variant="body2" color="text.secondary">
                    {uploadEntries.length} Datei
                    {uploadEntries.length === 1 ? "" : "en"} ausgewählt.
                  </Typography>
                  {uploadEntries.map((entry, idx) => (
                    <TextField
                      key={`${entry.file.name}-${idx}`}
                      label={`Titel für ${entry.file.name}`}
                      value={entry.title}
                      onChange={(event) =>
                        updateUploadEntryTitle(idx, event.target.value)
                      }
                      helperText="Optional – leer lassen, um Dateinamen zu verwenden."
                      fullWidth
                    />
                  ))}
                </Stack>
              )}

              <Alert severity="info">
                Events und Bilder werden aktuell nur lokal simuliert und gehen
                nach einem Seiten-Reload verloren.
              </Alert>

              {uploadError && <Alert severity="error">{uploadError}</Alert>}
            </Stack>
          </DialogContent>
          <DialogActions>
            <Button onClick={handleCloseUploadDialog} disabled={uploading}>
              Abbrechen
            </Button>
            <Button
              onClick={handleUploadImages}
              variant="contained"
              disabled={uploading}
            >
              {uploading ? "Lade hoch..." : "Hochladen"}
            </Button>
          </DialogActions>
        </Dialog>
      )}

      <Snackbar
        open={Boolean(successMessage)}
        autoHideDuration={4000}
        onClose={() => setSuccessMessage(null)}
        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
      >
        <Alert
          severity="success"
          variant="filled"
          sx={{ width: "100%" }}
          onClose={() => setSuccessMessage(null)}
        >
          {successMessage}
        </Alert>
      </Snackbar>

      {/* --- Lightbox --- */}
      <Dialog
        open={open}
        onClose={closeLightbox}
        fullWidth
        maxWidth="lg" // <-- Klassische Lightbox-Größe
        aria-labelledby="bild-dialog-title"
        PaperProps={{ sx: { overflow: "hidden" } }}
      >
        <DialogTitle
          id="bild-dialog-title"
          sx={{
            pr: (theme: Theme) => theme.spacing(6),
            position: "relative",
            zIndex: 1,
          }}
        >
          {currentImageTitle && (
            <Typography
              variant="subtitle1"
              sx={{ pr: (theme: Theme) => theme.spacing(4) }}
            >
              {currentImageTitle}
            </Typography>
          )}
          <IconButton
            aria-label="close"
            onClick={closeLightbox}
            sx={{
              position: "absolute",
              right: (theme: Theme) => theme.spacing(1.5),
              top: (theme: Theme) => theme.spacing(1.5),
              zIndex: 2,
            }}
          >
            <CloseIcon />
          </IconButton>
        </DialogTitle>

        <DialogContent
          dividers
          sx={{
            position: "relative",
            p: 0,
            bgcolor: "background.default",
          }}
        >
          {currentImage && (
            <Box
              component="img"
              src={currentImage.full} // Holt sich den Pfad (entweder /DEIN_ORDNER/... oder picsum)
              alt={currentImageTitle}
              sx={{
                display: "block",
                width: "100%",
                height: "auto",
                maxHeight: "85vh", // Begrenzt die Höhe, damit es auf den Schirm passt
                objectFit: "contain",
              }}
            />
          )}
          {thumbs.length > 1 && (
            <>
              <IconButton
                aria-label="previous"
                onClick={prev}
                sx={(theme: Theme) => ({
                  position: "absolute",
                  top: "50%",
                  left: theme.spacing(1.5),
                  transform: "translateY(-50%)",
                  color: theme.palette.text.primary,
                  bgcolor: alpha(theme.palette.background.default, 0.4),
                  zIndex: 2, // <-- KORREKTUR HINZUGEFÜGT
                  "&:hover": {
                    bgcolor: alpha(theme.palette.background.default, 0.7),
                  },
                })}
              >
                <ArrowBackIosNewIcon />
              </IconButton>
              <IconButton
                aria-label="next"
                onClick={next}
                sx={(theme: Theme) => ({
                  position: "absolute",
                  top: "50%",
                  right: theme.spacing(1.5),
                  transform: "translateY(-50%)",
                  color: theme.palette.text.primary,
                  bgcolor: alpha(theme.palette.background.default, 0.4),
                  zIndex: 2, // <-- KORREKTUR HINZUGEFÜGT
                  "&:hover": {
                    bgcolor: alpha(theme.palette.background.default, 0.7),
                  },
                })}
              >
                <ArrowForwardIosIcon />
              </IconButton>
            </>
          )}
        </DialogContent>
      </Dialog>
    </Container>
  );
}
