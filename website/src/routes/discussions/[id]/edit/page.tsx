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
import ArrowBackRoundedIcon from "@mui/icons-material/ArrowBackRounded";

import { useAuth } from "@lib/auth";
import { Sidebar } from "@components/layout";
import { getPrograms, getForumPostsById, putForumPostsById, getAuthCsrf } from "@lib/api";
import type { AuthProgramResponse as Program } from "@lib/api";
import { FORUM_CATEGORIES, FORUM_TAGS } from "../../components";

export default function EditPost() {
    const { user } = useAuth();
    const navigate = useNavigate();
    const { id } = useParams();

    const [programs, setPrograms] = useState<Program[]>([]);
    const [title, setTitle] = useState("");
    const [body, setBody] = useState("");
    const [type, setType] = useState("forum");
    const isStaff = user?.role === "admin" || user?.role === "editor";
    const [category, setCategory] = useState<string>("");
    const [selectedPrograms, setSelectedPrograms] = useState<Program[]>([]);
    const [tags, setTags] = useState<string[]>([]);
    const [eventDate, setEventDate] = useState("");
    const [location, setLocation] = useState("");

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

                    let loadedTags: string[] = [];
                    try {
                        loadedTags = data.tags ? JSON.parse(data.tags as unknown as string) as string[] : [];
                    } catch { /* empty */ }

                    const foundCategory = loadedTags.find(t => FORUM_CATEGORIES.includes(t as typeof FORUM_CATEGORIES[number]));
                    if (foundCategory) {
                        setCategory(foundCategory);
                        setTags(loadedTags.filter(t => t !== foundCategory));
                    } else {
                        setCategory("");
                        setTags(loadedTags);
                    }

                    if (data.event_date) setEventDate(data.event_date);
                    if (data.location) setLocation(data.location);
                }
                if (error) {
                    setError("Beitrag konnte nicht geladen werden.");
                }
            })
            .catch(() => setError("Fehler beim Laden."))
            .finally(() => setLoading(false));
    }, [id]);

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
                    location: location || undefined
                },
                headers: { "X-CSRF-Token": token }
            });

            if (apiError) throw new Error((apiError as { message?: string }).message || "Fehler beim Speichern.");

            navigate("/discussions");
        } catch (err: unknown) {
            setError((err as Error).message);
        } finally {
            setSaving(false);
        }
    };


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
                    <IconButton onClick={() => navigate("/discussions")} sx={{ bgcolor: 'action.hover' }}>
                        <ArrowBackRoundedIcon />
                    </IconButton>
                    <Typography variant="h4" fontWeight={700} sx={{ fontSize: { xs: '1.5rem', md: '2.125rem' } }}>
                        Beitrag bearbeiten
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
                            onChange={(e) => {
                                const val = e.target.value;
                                setCategory(val);
                                setType(val === "Ankündigung" ? "news" : "forum");
                            }}
                            fullWidth
                            required
                        >
                            {FORUM_CATEGORIES.filter(cat => cat !== "Ankündigung" || isStaff).map((cat) => (
                                <MenuItem key={cat} value={cat}>
                                    {cat}
                                </MenuItem>
                            ))}
                        </TextField>

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
                                disabled={saving || !title || !body}
                                sx={{ borderRadius: 2, px: 4, py: 1.5, fontWeight: 700 }}
                            >
                                {saving ? <CircularProgress size={20} color="inherit" /> : 'Speichern'}
                            </Button>
                        </Box>
                    </Stack>
                </Paper>
            </Box>
        </Sidebar>
    );
}
