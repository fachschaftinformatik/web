import { useState, useEffect } from "react";
import {
    Box,
    Typography,
    TextField,
    Button,
    Paper,
    Stack,
    MenuItem,
    Autocomplete,
    Chip,
    Alert,
    CircularProgress,
    IconButton
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import ArrowBackRoundedIcon from "@mui/icons-material/ArrowBackRounded";

import { useAuth } from "@lib/auth";
import { Sidebar } from "@components/layout";
import { getPrograms, postForumPosts, getAuthCsrf } from "@lib/api";
import type { AuthProgramResponse as Program } from "@lib/api";
import { FORUM_CATEGORIES, FORUM_TAGS } from "@lib/config";

export default function CreatePost() {
    const { user } = useAuth();
    const navigate = useNavigate();

    const [programs, setPrograms] = useState<Program[]>([]);
    const [title, setTitle] = useState("");
    const [body, setBody] = useState("");
    const [category, setCategory] = useState<string>("");
    const isStaff = user?.role === "admin" || user?.role === "editor";
    const type: "forum" | "news" | "event" = category === "Ankündigung" ? "news" : (category === "Termin" ? "event" : "forum");
    const [selectedPrograms, setSelectedPrograms] = useState<Program[]>([]);
    const [tags, setTags] = useState<string[]>([]);

    const [eventDate, setEventDate] = useState("");
    const [location, setLocation] = useState("");

    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    useEffect(() => {
        getPrograms().then(({ data }) => {
            if (data) setPrograms(data);
        });
    }, []);

    const handleSubmit = async () => {
        if (!title || !body || !category) return;

        setLoading(true);
        setError("");

        try {
            const { data: csrfData, error: csrfError } = await getAuthCsrf();
            const token = csrfData?.csrf;
            if (csrfError || !token) throw new Error("CSRF-Token fehlt. Bitte neu laden.");

            // Combine category and tags
            const finalTags = [category, ...tags];

            const { error: apiError } = await postForumPosts({
                body: {
                    title,
                    body,
                    type,
                    programs: selectedPrograms.map(p => p.name || ""),
                    tags: finalTags,
                    event_date: eventDate || undefined,
                    location: location || undefined
                },
                headers: { "X-CSRF-Token": token }
            });

            if (apiError) throw new Error((apiError as { message?: string }).message || "Fehler beim Erstellen.");

            navigate("/discussions");
        } catch (err: unknown) {
            setError((err as Error).message);
        } finally {
            setLoading(false);
        }
    };


    return (
        <Sidebar user={user} title="Beitrag erstellen" maxWidth="md">
            <Box>
                <Box sx={{ mb: 4, display: 'flex', alignItems: 'center', gap: 2 }}>
                    <IconButton onClick={() => navigate("/discussions")} sx={{ bgcolor: 'action.hover' }}>
                        <ArrowBackRoundedIcon />
                    </IconButton>
                    <Typography variant="h4" fontWeight={700} sx={{ fontSize: { xs: '1.5rem', md: '2.125rem' } }}>
                        Neuen Beitrag erstellen
                    </Typography>
                </Box>

                {error && <Alert severity="error" sx={{ mb: 3 }}>{error}</Alert>}

                <Paper elevation={0} sx={{ p: { xs: 2.5, md: 4 }, borderRadius: 3, border: '1px solid', borderColor: 'divider' }}>
                    <Stack spacing={3}>
                        <TextField
                            label="Titel"
                            fullWidth
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                            required
                        />

                        <TextField
                            select
                            label="Kategorie"
                            value={category}
                            onChange={(e) => setCategory(e.target.value)}
                            fullWidth
                            required
                        >
                            {FORUM_CATEGORIES.filter(cat => (cat !== "Ankündigung" && cat !== "Termin") || isStaff).map((cat) => (
                                <MenuItem key={cat} value={cat}>
                                    {cat}
                                </MenuItem>
                            ))}
                        </TextField>


                        {/* Inhalt */}
                        <TextField
                            label="Inhalt"
                            multiline
                            rows={6}
                            fullWidth
                            value={body}
                            onChange={(e) => setBody(e.target.value)}
                            required
                        />

                        <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                            <Autocomplete
                                multiple
                                options={programs}
                                getOptionLabel={(option) => option.name || ""}
                                value={selectedPrograms}
                                isOptionEqualToValue={(option, value) => option.id === value.id}
                                onChange={(_, newValue) => setSelectedPrograms(newValue)}
                                renderInput={(params) => (
                                    <TextField {...params} label="Studiengänge (Optional)" placeholder="Wählen..." />
                                )}
                                fullWidth
                            />

                            <Autocomplete
                                multiple
                                options={FORUM_TAGS}
                                value={tags}
                                onChange={(_, newValue) => setTags(newValue)}
                                renderTags={(value: readonly string[], getTagProps) =>
                                    value.map((option: string, index: number) => (
                                        <Chip variant="outlined" label={option} {...getTagProps({ index })} />
                                    ))
                                }
                                renderInput={(params) => (
                                    <TextField
                                        {...params}
                                        label="Tags"
                                        placeholder="Wählen..."
                                    />
                                )}
                                fullWidth
                            />
                        </Stack>

                        {type !== 'forum' && (
                            <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                                <TextField
                                    label="Event Datum"
                                    type="datetime-local"
                                    fullWidth
                                    value={eventDate}
                                    onChange={(e) => setEventDate(e.target.value)}
                                    InputLabelProps={{ shrink: true }}
                                />
                                <TextField
                                    label="Ort"
                                    fullWidth
                                    value={location}
                                    onChange={(e) => setLocation(e.target.value)}
                                />
                            </Stack>
                        )}

                        <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 2 }}>
                            <Button
                                variant="contained"
                                onClick={handleSubmit}
                                disabled={loading || !title || !body}
                                sx={{ borderRadius: 2, px: 4, py: 1.5, fontWeight: 700 }}
                            >
                                {loading ? <CircularProgress size={20} color="inherit" /> : 'Veröffentlichen'}
                            </Button>
                        </Box>
                    </Stack>
                </Paper>
            </Box>
        </Sidebar>
    );
}
