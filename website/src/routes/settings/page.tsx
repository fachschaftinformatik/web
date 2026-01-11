import { useEffect, useState } from 'react';
import {
    Typography,
    Box,
    Paper,
    Stack,
    Switch,
    FormControlLabel,
    TextField,
    MenuItem,
    Button,
    Alert,
    CircularProgress,
    InputAdornment,
    Snackbar,
    Slide,
    type SlideProps
} from '@mui/material';
import SaveRoundedIcon from '@mui/icons-material/SaveRounded';
import { Sidebar } from '@components/layout';
import { useAuth } from '@lib/auth';
import { getPrograms, getAuthCsrf, putAuthMe } from '@lib/api';
import type { AuthProgramResponse as Program } from '@lib/api';
import { useThemeMode, type ThemePreference } from '@lib/theme';
import { alpha } from '@mui/material/styles';

export default function SettingsPage() {
    const { user, login } = useAuth();
    const [programs, setPrograms] = useState<Program[]>([]);
    const [name, setName] = useState(user?.name || "");
    const [programId, setProgramId] = useState<number>(Number(user?.programid) || 0);
    const [themePreference, setThemePreference] = useState<ThemePreference>((user as any)?.theme || 'system');
    const [loading, setLoading] = useState(false);
    const [success, setSuccess] = useState("");
    const [error, setError] = useState("");
    const { setPreference } = useThemeMode();

    function TransitionUp(props: SlideProps) {
        return <Slide {...props} direction="up" />;
    }

    const isDirty = name !== (user?.name || "") ||
        programId !== (Number(user?.programid) || 0) ||
        themePreference !== ((user as any)?.theme || 'system');

    useEffect(() => {
        document.title = "Einstellungen | FSV Informatik";
        getPrograms().then(({ data }) => {
            if (data) setPrograms(data);
        });
    }, []);

    // Sync state if user changes (e.g. after re-auth)
    useEffect(() => {
        if (user) {
            setName(user.name || "");
            setProgramId(Number(user.programid));
            setThemePreference((user as any).theme || 'system');
        }
    }, [user]);

    const handleSave = async () => {
        setLoading(true);
        setError("");
        setSuccess("");
        try {
            const { data: csrfData } = await getAuthCsrf();
            const token = csrfData?.csrf;

            const res = await putAuthMe({
                body: { name, programid: programId, theme: themePreference } as any,
                headers: { "X-CSRF-Token": token || "" }
            });

            if (res.error) {
                const msg = (res.error as { message?: string }).message || "Fehler beim Speichern der Einstellungen.";
                setError(msg);
            } else {
                setSuccess("Einstellungen erfolgreich gespeichert.");
                if (res.data) {
                    // Update the local user state
                    login(res.data as any, window.localStorage.getItem('fs_remember_flag') === 'true');
                    // Apply the theme after successful save
                    setPreference(themePreference);
                }
            }
        } catch (err) {
            setError("Netzwerkfehler beim Speichern.");
        } finally {
            setLoading(false);
        }
    };

    return (
        <Sidebar user={user} title="Einstellungen" maxWidth="md">
            <Box>
                <Box sx={{ mb: 4 }}>
                    <Typography variant="h4" fontWeight={700} gutterBottom>
                        Einstellungen
                    </Typography>
                    <Typography variant="body1" color="text.secondary" sx={{ opacity: 0.8 }}>
                        Verwalte dein Profil und deine Präferenzen für ein personalisiertes Erlebnis.
                    </Typography>
                </Box>

                <Stack spacing={4}>
                    <Snackbar
                        open={!!success}
                        autoHideDuration={4000}
                        onClose={() => setSuccess("")}
                        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
                        TransitionComponent={TransitionUp}
                    >
                        <Alert
                            severity="success"
                            variant="filled"
                            onClose={() => setSuccess("")}
                            sx={{
                                borderRadius: 3,
                                minWidth: '300px',
                                fontWeight: 600,
                                backdropFilter: 'blur(10px)',
                                backgroundColor: (theme) => alpha(theme.palette.success.main, 0.85),
                                boxShadow: '0 8px 32px rgba(0,0,0,0.12)',
                                border: '1px solid',
                                borderColor: (theme) => alpha(theme.palette.success.light, 0.5)
                            }}
                        >
                            {success}
                        </Alert>
                    </Snackbar>

                    <Snackbar
                        open={!!error}
                        autoHideDuration={6000}
                        onClose={() => setError("")}
                        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
                        TransitionComponent={TransitionUp}
                    >
                        <Alert
                            severity="error"
                            variant="filled"
                            onClose={() => setError("")}
                            sx={{
                                borderRadius: 3,
                                minWidth: '300px',
                                fontWeight: 600,
                                backdropFilter: 'blur(10px)',
                                backgroundColor: (theme) => alpha(theme.palette.error.main, 0.85),
                                boxShadow: '0 8px 32px rgba(0,0,0,0.12)',
                                border: '1px solid',
                                borderColor: (theme) => alpha(theme.palette.error.light, 0.5)
                            }}
                        >
                            {error}
                        </Alert>
                    </Snackbar>

                    <Paper elevation={0} sx={{ p: 4, borderRadius: 3, border: '1px solid', borderColor: 'divider', bgcolor: 'background.paper' }}>
                        <Typography variant="h6" fontWeight={600} gutterBottom sx={{ mb: 3 }}>
                            Profil
                        </Typography>
                        <Stack spacing={2}>
                            <TextField
                                label="E-Mail"
                                value={user?.email?.split('@')[0] || ""}
                                disabled
                                fullWidth
                                variant="outlined"
                                helperText="Die E-Mail-Adresse kann nicht geändert werden."
                                slotProps={{
                                    input: {
                                        endAdornment: <InputAdornment position="end">@studmail.w-hs.de</InputAdornment>,
                                    },
                                }}
                            />
                            <TextField
                                label="Name"
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                fullWidth
                                placeholder="Dein vollständiger Name"
                            />
                            <TextField
                                select
                                label="Studiengang"
                                value={programId}
                                onChange={(e) => setProgramId(Number(e.target.value))}
                                fullWidth
                            >
                                {programs.length > 0 ? (
                                    programs.map((p) => (
                                        <MenuItem key={p.id} value={p.id}>{p.name}</MenuItem>
                                    ))
                                ) : (
                                    <MenuItem disabled value={0}>Lädt...</MenuItem>
                                )}
                            </TextField>
                        </Stack>
                    </Paper>

                    <Paper elevation={0} sx={{ p: 4, borderRadius: 3, border: '1px solid', borderColor: 'divider' }}>
                        <Typography variant="h6" fontWeight={600} gutterBottom sx={{ mb: 2 }}>
                            Benachrichtigungen
                        </Typography>
                        <Stack spacing={1}>
                            <FormControlLabel
                                control={<Switch defaultChecked color="primary" />}
                                label={<Typography variant="body2">E-Mail-Benachrichtigungen bei neuen Rekos</Typography>}
                            />
                            <FormControlLabel
                                control={<Switch defaultChecked color="primary" />}
                                label={<Typography variant="body2">E-Mail-Benachrichtigungen bei News</Typography>}
                            />
                        </Stack>
                    </Paper>

                    <Paper elevation={0} sx={{ p: 4, borderRadius: 3, border: '1px solid', borderColor: 'divider' }}>
                        <Typography variant="h6" fontWeight={600} gutterBottom sx={{ mb: 2 }}>
                            Anzeige
                        </Typography>
                        <Stack spacing={3}>
                            <TextField
                                select
                                label="Erscheinungsbild"
                                value={themePreference}
                                onChange={(e) => {
                                    const val = e.target.value as ThemePreference;
                                    setThemePreference(val);
                                }}
                                fullWidth
                                size="small"
                                helperText="Wähle zwischen hellem, dunklem oder dem System-Design."
                            >
                                <MenuItem value="light">Hell</MenuItem>
                                <MenuItem value="dark">Dunkel</MenuItem>
                                <MenuItem value="system">System</MenuItem>
                            </TextField>
                        </Stack>
                    </Paper>

                    <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 2 }}>
                        <Button
                            variant="contained"
                            startIcon={loading ? <CircularProgress size={20} color="inherit" /> : <SaveRoundedIcon />}
                            onClick={handleSave}
                            disabled={loading || !name || !programId || !isDirty}
                            sx={{
                                borderRadius: 2,
                                px: 6,
                                py: 1.5,
                                fontWeight: 700,
                                boxShadow: 'none',
                                '&:hover': {
                                    boxShadow: '0 8px 24px rgba(0,0,0,0.12)',
                                    transform: 'translateY(-2px)'
                                },
                                '&:disabled': {
                                    opacity: 0.6
                                },
                                transition: 'all 0.2s ease'
                            }}
                        >
                            Änderungen speichern
                        </Button>
                    </Box>
                </Stack>
            </Box>
        </Sidebar>
    );
}
