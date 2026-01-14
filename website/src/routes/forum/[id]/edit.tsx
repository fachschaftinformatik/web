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
import { useNavigate, useParams } from "react-router-dom";
import SaveRoundedIcon from "@mui/icons-material/SaveRounded";
import ArrowBackRoundedIcon from "@mui/icons-material/ArrowBackRounded";

import { useAuth } from "@lib/auth";
import { Sidebar } from "@components/layout";
import { getPrograms, getForumPostsById, putForumPostsById, getAuthCsrf } from "@lib/api";
import type { AuthProgramResponse as Program } from "@lib/api";
import { FORUM_CATEGORIES, FORUM_TAGS } from "../components";

export default function EditPost() {
    const { user } = useAuth();
    const navigate = useNavigate();
    const { id } = useParams();

    const [programs, setPrograms] = useState<Program[]>([]);
    const [title, setTitle] = useState("");
    const [body, setBody] = useState("");
    const [type, setType] = useState("forum");
    const [category, setCategory] = useState<string>("");
    const [selectedPrograms, setSelectedPrograms] = useState<Program[]>([]);
    const [tags, setTags] = useState<string[]>([]);
    // const [tagInput, setTagInput] = useState(""); // Removed custom tag input

    // Additional fields for News/Events
    const [eventDate, setEventDate] = useState("");
    const [location, setLocation] = useState("");
    const [imageUrl, setImageUrl] = useState("");
    const [links, setLinks] = useState<string[]>([]);
    const [linkInput, setLinkInput] = useState("");

    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState("");

    useEffect(() => {
        getPrograms().then(({ data }) => {
            if (data) setPrograms(data);
        });
    }, []);

    useEffect(() => {
        if (!id) return;
        setLoading(true);
        getForumPostsById({ path: { id } })
            .then(({ data, error }) => {
                if (data) {
                    setTitle(data.title || "");
                    setBody(data.body || "");
                    setType(data.type || "forum");

                    // Handle Category and Tags
                    let loadedTags: string[] = [];
                    try {
                        loadedTags = data.tags ? JSON.parse(data.tags as unknown as string) as string[] : [];
                    } catch { /* empty */ }

                    // Find if any tag matches a category
                    const foundCategory = loadedTags.find(t => FORUM_CATEGORIES.includes(t as typeof FORUM_CATEGORIES[number]));
                    if (foundCategory) {
                        setCategory(foundCategory);
                        // Filter out the category from the tags displayed in the tag selector
                        setTags(loadedTags.filter(t => t !== foundCategory));
                    } else {
                        setCategory(""); // Or default?
                        setTags(loadedTags);
                    }

                    if (data.event_date) setEventDate(data.event_date);
                    if (data.location) setLocation(data.location);
                    if (data.image_url) setImageUrl(data.image_url);
                    if (data.links) {
                        try {
                            const l = JSON.parse(data.links as unknown as string) as string[];
                            setLinks(l);
                        } catch { /* empty */ }
                    }
                }
                if (error) {
                    setError("Beitrag konnte nicht geladen werden.");
                }
            })
            .catch(() => setError("Fehler beim Laden."))
            .finally(() => setLoading(false));
    }, [id]);

    // Effect to match loaded program names to Program objects once programs list is available
    useEffect(() => {
        if (!id || programs.length === 0) return;

        getForumPostsById({ path: { id } }).then(({ data }) => {
            if (data && data.programs) {
                try {
                    const names = JSON.parse(data.programs as unknown as string) as string[];
                    const matched = programs.filter(p => p.name && names.includes(p.name));
                    setSelectedPrograms(matched);
                } catch { /* empty */ }
            }
        });
    }, [id, programs]);


    const handleSubmit = async () => {
        if (!title || !body || !id || !category) return;

        setSaving(true);
        setError("");

        try {
            const { data: csrfData, error: csrfError } = await getAuthCsrf();
            const token = csrfData?.csrf;
            if (csrfError || !token) throw new Error("CSRF-Token fehlt. Bitte neu laden.");

            // Combine category and tags
            const finalTags = [category, ...tags];

            const { error: apiError } = await putForumPostsById({
                path: { id },
                body: {
                    title,
                    body,
                    type,
                    programs: selectedPrograms.map(p => p.name || ""),
                    tags: finalTags,
                    event_date: eventDate || undefined,
                    location: location || undefined,
                    image_url: imageUrl || undefined,
                    links: links
                },
                headers: { "X-CSRF-Token": token }
            });

            if (apiError) throw new Error((apiError as { message?: string }).message || "Fehler beim Speichern.");

            navigate("/forum");
        } catch (err: unknown) {
            setError((err as Error).message);
        } finally {
            setSaving(false);
        }
    };

    /* const handleAddTag = (event: React.KeyboardEvent) => {
        // ... removed
    }; */

    const handleAddLink = (event: React.KeyboardEvent) => {
        if (event.key === 'Enter' && linkInput.trim()) {
            event.preventDefault();
            if (!links.includes(linkInput.trim())) {
                setLinks([...links, linkInput.trim()]);
            }
            setLinkInput("");
        }
    }

    // const canCreateNews = user?.role === "admin" || user?.role === "editor";

    if (loading) {
        return (
            <Sidebar user={user} title="Beitrag bearbeiten">
                <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>
            </Sidebar>
        );
    }

    return (
        <Sidebar user={user} title="Beitrag bearbeiten" maxWidth="md">
            <Box>
                <Box sx={{ mb: 4, display: 'flex', alignItems: 'center', gap: 2 }}>
                    <IconButton onClick={() => navigate("/forum")} sx={{ bgcolor: 'action.hover' }}>
                        <ArrowBackRoundedIcon />
                    </IconButton>
                    <Typography variant="h4" fontWeight={700}>
                        Beitrag bearbeiten
                    </Typography>
                </Box>

                {error && <Alert severity="error" sx={{ mb: 3 }}>{error}</Alert>}

                <Paper elevation={0} sx={{ p: 4, borderRadius: 3, border: '1px solid', borderColor: 'divider' }}>
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
                            {FORUM_CATEGORIES.map((cat) => (
                                <MenuItem key={cat} value={cat}>
                                    {cat}
                                </MenuItem>
                            ))}
                        </TextField>

                        {/*
                        {canCreateNews && (
                            <TextField
                                select
                                label="Typ"
                                value={type}
                                onChange={(e) => setType(e.target.value)}
                                fullWidth
                            >
                                <MenuItem value="forum">Diskussion</MenuItem>
                                <MenuItem value="news">News</MenuItem>
                                <MenuItem value="event">Event</MenuItem>
                            </TextField>
                        )}
                        */}

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

                        {(type === 'news' || type === 'event') && (
                            <Stack spacing={3}>
                                <TextField
                                    label="Bild URL"
                                    fullWidth
                                    value={imageUrl}
                                    onChange={(e) => setImageUrl(e.target.value)}
                                    placeholder="https://..."
                                />
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
                                <Autocomplete
                                    multiple
                                    freeSolo
                                    options={[]}
                                    value={links}
                                    onChange={(_, newValue) => setLinks(newValue)}
                                    renderTags={(value: readonly string[], getTagProps) =>
                                        value.map((option: string, index: number) => (
                                            <Chip variant="outlined" label={option} {...getTagProps({ index })} />
                                        ))
                                    }
                                    renderInput={(params) => (
                                        <TextField
                                            {...params}
                                            label="Links"
                                            placeholder="URL eingeben und Enter drücken..."
                                            value={linkInput}
                                            onChange={(e) => setLinkInput(e.target.value)}
                                            onKeyDown={handleAddLink}
                                        />
                                    )}
                                    fullWidth
                                />
                            </Stack>
                        )}

                        <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 2 }}>
                            <Button
                                variant="contained"
                                startIcon={saving ? <CircularProgress size={20} color="inherit" /> : <SaveRoundedIcon />}
                                onClick={handleSubmit}
                                disabled={saving || !title || !body}
                                sx={{ borderRadius: 2, px: 4, py: 1.5, fontWeight: 700 }}
                            >
                                Speichern
                            </Button>
                        </Box>
                    </Stack>
                </Paper>
            </Box>
        </Sidebar>
    );
}
