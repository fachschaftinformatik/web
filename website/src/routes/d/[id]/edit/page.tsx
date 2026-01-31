import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import MenuItem from "@mui/material/MenuItem";
import Autocomplete from "@mui/material/Autocomplete";
import Chip from "@mui/material/Chip";
import Alert from "@mui/material/Alert";
import CircularProgress from "@mui/material/CircularProgress";
import IconButton from "@mui/material/IconButton";
import ArrowBackRounded from "@mui/icons-material/ArrowBackRounded";

import { useAuth } from "@lib/auth";
import { Sidebar } from "@components/layout";
import { getPrograms, getDiscussionsByPostId, putDiscussionsByPostId } from "@lib/api";
import type { DtoProgramResponse as Program } from "@lib/api";
import { FORUM_CATEGORIES, FORUM_TAGS } from "../../components";

export default function EditPost() {
    const { user } = useAuth();
    const navigate = useNavigate();
    const { postId } = useParams<{ postId: string }>();

    const [programs, setPrograms] = useState<Program[]>([]);
    const [title, setTitle] = useState("");
    const [body, setBody] = useState("");
    const [type, setType] = useState<"discussion" | "news" | "event">("discussion");
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
        if (!postId) return;
        setLoading(true);
        getDiscussionsByPostId({ path: { postId } })
            .then(({ data, error }) => {
                if (data) {
                    setTitle(data.title || "");
                    setBody(data.body || "");
                    setType(data.type as "discussion" | "news" | "event" || "discussion");

                    const loadedTags = data.tags || [];
                    const foundCategory = loadedTags.find(t => (FORUM_CATEGORIES as readonly string[]).includes(t));
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
    }, [postId]);

    useEffect(() => {
        if (!postId || programs.length === 0) return;

        getDiscussionsByPostId({ path: { postId } }).then(({ data }) => {
            if (data && data.programs) {
                const ids = data.programs.map(p => String(p.id));
                const matched = programs.filter(p => p.id && ids.includes(String(p.id)));
                setSelectedPrograms(matched);
            }
        });
    }, [postId, programs]);


    const handleSubmit = async () => {
        if (!title || !body || !postId || !category) return;

        setSaving(true);
        setError("");

        try {
            const finalTags = [category, ...tags];

            const { error: apiError } = await putDiscussionsByPostId({
                path: { postId },
                body: {
                    title,
                    body,
                    type,
                    programs: selectedPrograms.map(p => String(p.id)),
                    tags: finalTags,
                    event_date: eventDate || undefined,
                    location: location || undefined
                }
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
                        <ArrowBackRounded />
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
                                setType(val === "Ankündigung" ? "news" : "discussion");
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
                                isOptionEqualToValue={(option, value) => String(option.id) === String(value.id)}
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

                        {type === 'event' && (
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
