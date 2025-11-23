import { useState } from "react";
import {
  Box,
  TextField,
  Button,
  Typography,
  Container,
  Chip,
  Stack,
  Paper
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import { Sidebar } from "@components/layout";
import { useAuth } from "@lib/auth";

export default function CreateNews() {
  const navigate = useNavigate();
  const { user } = useAuth();

  const [title, setTitle] = useState("");
  const [date, setDate] = useState("");
  const [imageFile, setImageFile] = useState<File | null>(null);
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
      image: imageFile ? URL.createObjectURL(imageFile) : null,
      pdf: pdf ? URL.createObjectURL(pdf) : null,
      pdfName: pdf?.name || null,
      summary,
      content,
      tags,
      links,
      isNew: true
    };

    const stored = JSON.parse(localStorage.getItem("custom-news") || "[]");
    localStorage.setItem("custom-news", JSON.stringify([...stored, newNews]));

    navigate("/news");
  }

  return (
    <Sidebar user={user} title="News erstellen">
      <Container maxWidth="md" sx={{ py: 4 }}>
        <Typography variant="h4" mb={3} fontWeight={600}>
          Neuen Beitrag erstellen
        </Typography>

        <Paper variant="outlined" sx={{ p: 4 }}>
          <Stack spacing={3}>
            <TextField
              fullWidth
              label="Titel"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />

            <TextField
              fullWidth
              label="Datum (z.B. 12. Oktober 2025)"
              value={date}
              onChange={(e) => setDate(e.target.value)}
            />

            {/* Bild Upload */}
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
                  onChange={(e) => setImageFile(e.target.files?.[0] || null)}
                />
              </Box>
              {imageFile && <Typography mt={1} variant="caption">Bild: {imageFile.name}</Typography>}
            </Box>

            {/* PDF Upload */}
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
              {pdf && <Typography mt={1} variant="caption" display="block">Datei: {pdf.name}</Typography>}
            </Box>

            <TextField
              fullWidth
              multiline
              minRows={3}
              label="Kurzbeschreibung"
              value={summary}
              onChange={(e) => setSummary(e.target.value)}
            />

            <TextField
              fullWidth
              multiline
              minRows={6}
              label="Inhalt"
              value={content}
              onChange={(e) => setContent(e.target.value)}
            />

            {/* Tags */}
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

            {/* Links */}
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

              <Stack mt={1} spacing={0.5}>
                {links.map((l, index) => (
                  <Typography key={index} variant="body2" color="primary">{l}</Typography>
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
      </Container>
    </Sidebar>
  );
}
