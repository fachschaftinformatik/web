import { Typography, Paper, Box } from '@mui/material';
import { Sidebar } from '@components/layout';
import { useAuth } from '@lib/auth';

export default function PrivacyPage() {
    const { user } = useAuth();

    return (
        <Sidebar user={user} title="Datenschutz" maxWidth="lg">
            <Box>
                <Typography variant="h4" fontWeight={700} gutterBottom>
                    Datenschutz
                </Typography>
                <Paper elevation={0} sx={{ p: 4, borderRadius: 3, border: '1px solid', borderColor: 'divider' }}>
                    <Typography variant="body1" paragraph>
                        Wir nehmen den Schutz Ihrer persönlichen Daten sehr ernst. Wir behandeln Ihre personenbezogenen Daten vertraulich und entsprechend der gesetzlichen Datenschutzvorschriften sowie dieser Datenschutzerklärung.
                    </Typography>
                    <Typography variant="h6" gutterBottom sx={{ mt: 3 }}>
                        Allgemeine Hinweise
                    </Typography>
                    <Typography variant="body1">
                        Die Nutzung unserer Webseite ist in der Regel ohne Angabe personenbezogener Daten möglich. Soweit auf unseren Seiten personenbezogene Daten (beispielsweise Name, Anschrift oder E-Mail-Adressen) erhoben werden, erfolgt dies, soweit möglich, stets auf freiwilliger Basis.
                    </Typography>
                </Paper>
            </Box>
        </Sidebar>
    );
}
