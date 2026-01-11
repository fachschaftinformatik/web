import { Typography, Paper, Box } from '@mui/material';
import { Sidebar } from '@components/layout';
import { useAuth } from '@lib/auth';

export default function LegalPage() {
    const { user } = useAuth();

    return (
        <Sidebar user={user} title="Impressum" maxWidth="lg">
            <Box>
                <Typography variant="h4" fontWeight={700} gutterBottom>
                    Impressum
                </Typography>
                <Paper elevation={0} sx={{ p: 4, borderRadius: 3, border: '1px solid', borderColor: 'divider' }}>
                    <Typography variant="body1" paragraph>
                        Angaben gemäß § 5 TMG:
                    </Typography>
                    <Typography variant="h6" gutterBottom sx={{ mt: 3 }}>
                        Kontakt
                    </Typography>
                    <Typography variant="body1">
                        Fachschaft Informatik<br />
                        Hochschule Westfälische<br />
                        Neidenburger Str. 43<br />
                        45897 Gelsenkirchen
                    </Typography>
                    <Typography variant="body1" sx={{ mt: 2 }}>
                        E-Mail: fachschaft.informatik@w-hs.de
                    </Typography>
                </Paper>
            </Box>
        </Sidebar>
    );
}
