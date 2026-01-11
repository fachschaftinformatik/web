import { useState } from "react";
import {
  Box,
  TextField,
  Button,
  Typography,
  Chip,
  Stack,
  Paper
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import { Sidebar } from "@components/layout";
import { useAuth } from "@lib/auth";
import DeleteIcon from "@mui/icons-material/Delete";

export default function CreateNews() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [title, setTitle] = useState("");
  const [date, setDate] = useState("");
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [summary, setSummary] = useState("");
  const [content, setContent] = useState("");
  const [pdf, setPdf] = useState<File | null>(null);
  const [tags, setTags] = useState<string[]>([]);
  const [links, setLinks] = useState<string[]>([]);
  const [currentLink, setCurrentLink] = useState("");

  const availableTags = ["Studium", "Prüfungen", "Events", "Jobs"];

  function saveToLocalStorage() {
    const newNews = {
      id: Date.now(),
      title,
      date,
      createdAt: new Date().toISOString(),
      image: imageFile ? URL.createObjectURL(imageFile) : null,
      pdf: pdf ? URL.createObjectURL(pdf) : null,
      pdfName: pdf?.name || null,
      summary,
      content,
      tags,
      links,
      isNewAt: Date.now(),
    };
    const stored = JSON.parse(localStorage.getItem("custom-news") || "[]");
    localStorage.setItem("custom-news", JSON.stringify([...stored, newNews]));
    navigate("/news");
  }
  return (
    <Sidebar user={user} title="News erstellen" maxWidth="md">
      <Box>
        <Typography variant="h4" mb={3} fontWeight={600}>
          Neuen Beitrag erstellen
        </Typography>
        <Paper variant="outlined" sx={{ p: 4 }}>
          <Stack spacing={3}>
            <TextField
              id="news-title"
              name="title"
              fullWidth
              label="Titel*"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />
            <TextField
              id="news-date"
              name="date"
              fullWidth
              label="Datum (z.B. 12.10.2025)*"
              value={date}
              onChange={(e) => setDate(e.target.value)}
            />

            <Box>
              <Typography variant="subtitle2" mb={1}>Bild auswählen:</Typography>
              <Box
                sx={{
                  border: "2px dashed",
                  borderColor: "divider",
                  borderRadius: 2,
                  p: 3,
                  textAlign: "center",
                  cursor: "pointer",
                  bgcolor: "background.default",
                  "&:hover": { borderColor: "primary.main" }
                }}
                onClick={() => document.getElementById("imageInput")?.click()}
              >
                <Typography color="text.secondary">📷 Klicken um ein Bild auszuwählen</Typography>
                <input
                  id="imageInput"
                  type="file"
                  accept="image/*"
                  style={{ display: "none" }}
                  onChange={(e) => {
                    const file = e.target.files?.[0] || null;
                    setImageFile(file);
                    if (file) setImagePreview(URL.createObjectURL(file));
                  }}
                />
              </Box>

              {imagePreview && (
                <Box sx={{ mt: 2, display: "flex", flexDirection: "column", alignItems: "flex-start" }}>


                  <Box
                    component="img"
                    src={imagePreview}
                    alt="Bildvorschau"
                    sx={{
                      width: 150,
                      height: "auto",
                      borderRadius: 2,
                      objectFit: "cover",
                      boxShadow: 2,
                      mb: 1
                    }}
                  />

                  <Button
                    variant="outlined"
                    color="error"
                    size="small"
                    endIcon={<DeleteIcon />}
                    onClick={() => {
                      setImageFile(null);
                      setImagePreview(null);
                    }}
                  >
                    Bild entfernen
                  </Button>
                </Box>
              )}
            </Box>

            <Box>
              <Typography variant="subtitle2" mb={1}>PDF / Datei auswählen:</Typography>
              <Button
                variant="outlined"
                component="label"
              >
                Datei auswählen
                <input
                  type="file"
                  hidden
                  accept=".pdf"
                  onChange={(e) => setPdf(e.target.files?.[0] || null)}
                />
              </Button>

              {pdf && (
                <Box sx={{ mt: 2, display: "flex", flexDirection: "column", alignItems: "flex-start" }}>
                  <Stack direction="row" spacing={2} alignItems="center">
                    <Typography variant="body1">📄 {pdf.name}</Typography>
                  </Stack>
                  <Button
                    variant="outlined"
                    color="error"
                    size="small"
                    endIcon={<DeleteIcon />}
                    sx={{ mt: 1 }}
                    onClick={() => setPdf(null)}
                  >
                    Datei entfernen
                  </Button>
                </Box>
              )}
            </Box>

            <TextField
              id="news-summary"
              name="summary"
              fullWidth
              multiline
              minRows={3}
              label="Kurzbeschreibung"
              value={summary}
              onChange={(e) => setSummary(e.target.value)}
            />

            <TextField
              id="news-content"
              name="content"
              fullWidth
              multiline
              minRows={6}
              label="Inhalt"
              value={content}
              onChange={(e) => setContent(e.target.value)}
            />

            <Box>
              <Typography variant="subtitle2" mb={1}>Tags:</Typography>
              <Stack direction="row" spacing={1}>
                {availableTags.map((tag) => (
                  <Chip
                    key={tag}
                    label={tag}
                    variant={tags.includes(tag) ? "filled" : "outlined"}
                    color={tags.includes(tag) ? "success" : "default"}
                    onClick={() =>
                      setTags((prev) =>
                        prev.includes(tag) ? prev.filter((t) => t !== tag) : [...prev, tag]
                      )
                    }
                  />
                ))}
              </Stack>
            </Box>

            <Box>
              <Typography variant="subtitle2" mb={1}>Links:</Typography>

              <Stack direction="row" spacing={2}>
                <TextField
                  fullWidth
                  size="small"
                  label="Neuer Link"
                  value={currentLink}
                  onChange={(e) => setCurrentLink(e.target.value)}
                />
                <Button
                  variant="outlined"
                  onClick={() => {
                    if (currentLink.trim() !== "") {
                      setLinks([...links, currentLink]);
                      setCurrentLink("");
                    }
                  }}
                >
                  Hinzufügen
                </Button>
              </Stack>

              <Stack mt={2} spacing={1}>
                {links.map((l, index) => (
                  <Box
                    key={index}
                    sx={{
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "space-between",
                      border: "1px solid",
                      borderColor: "divider",
                      p: 1,
                      borderRadius: 1
                    }}
                  >
                    <Typography variant="body2" color="primary">
                      {l}
                    </Typography>
                    <Button
                      variant="text"
                      color="error"
                      size="small"
                      onClick={() =>
                        setLinks(links.filter((_, i) => i !== index))
                      }
                    >
                      Entfernen
                    </Button>
                  </Box>
                ))}
              </Stack>
            </Box>

            <Button
              variant="contained"
              size="large"
              onClick={saveToLocalStorage}
              disabled={!title || !date}
            >
              Beitrag speichern
            </Button>
          </Stack>
        </Paper>
      </Box>
    </Sidebar>
  );
}