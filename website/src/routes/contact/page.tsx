import { Typography, Paper, Box, Divider } from '@mui/material';
import { Sidebar } from '@components/layout';
import { useAuth } from '@lib/auth';


export default function ContactPage() {
    const { user } = useAuth();

    return (
        <Sidebar user={user} title="Kontakt" maxWidth="lg">
            <Box>
                <Typography variant="h4" fontWeight={700} gutterBottom>
                    Kontakt
                </Typography>

                <Paper elevation={0} sx={{ p: 4, borderRadius: 3, border: '1px solid', borderColor: 'divider' }}>
                    <Typography variant="h6" gutterBottom>
                        Impressum
                    </Typography>
                    <Divider sx={{ my: 2 }} />
                    <Typography variant="body1" paragraph>
                        Angaben gemäß § 5 TMG:
                    </Typography>
                    <Typography variant="body1" fontWeight={600}>
                        Fachschaft Informatik
                    </Typography>
                    <Typography variant="body1">
                        Westfälische Hochschule<br />
                        Neidenburger Str. 43<br />
                        45897 Gelsenkirchen
                    </Typography>
                    <Typography variant="body1" sx={{ mt: 2 }}>
                        E-Mail: <a href="mailto:fachschaft.informatik@w-hs.de" style={{ color: 'inherit' }}>fachschaft.informatik@w-hs.de</a>
                    </Typography>

                    <Box sx={{ mt: 6, mb: 2 }}>
                        <Typography variant="h6" gutterBottom>
                            Datenschutz
                        </Typography>
                        <Divider />
                    </Box>

                    <Typography variant="body1" paragraph>
                        Wir nehmen den Schutz Ihrer persönlichen Daten sehr ernst. Wir behandeln Ihre personenbezogenen Daten vertraulich und entsprechend der gesetzlichen Datenschutzvorschriften sowie dieser Datenschutzerklärung.
                    </Typography>
                    <Typography variant="subtitle1" fontWeight={600} gutterBottom sx={{ mt: 3 }}>
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
