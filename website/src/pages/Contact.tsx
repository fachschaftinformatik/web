import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import Stack from '@mui/material/Stack';
import Page from '@components/Page';

export default function Contact() {
  return (
    <Page title="Kontakt" description="Impressum und Datenschutz.">
      <Paper variant="outlined" sx={{ p: 4, borderRadius: 3 }}>
        <Typography variant="h6" gutterBottom>Impressum</Typography>
        <Divider sx={{ my: 2 }} />
        <Stack spacing={2}>
          <Box>
            <Typography variant="body2" fontWeight={600}>Fachschaftsvertretung Informatik</Typography>
            <Typography variant="body2">Neidenburgerstr. 43, 45897 Gelsenkirchen</Typography>
          </Box>
          <Box>
            <Typography variant="body2" fontWeight={600}>Vertreten durch:</Typography>
            <Typography variant="body2">Lennart Möller</Typography>
          </Box>
          <Box>
            <Typography variant="body2" fontWeight={600}>Kontakt:</Typography>
            <Typography variant="body2">Telefon: (02 09) 95 96-416</Typography>
            <Typography variant="body2">E-Mail: <a href="mailto:kontakt@fachschaftinformatik.de" style={{ color: 'inherit' }}>kontakt@fachschaftinformatik.de</a></Typography>
          </Box>
        </Stack>
        <Box sx={{ mt: 6, mb: 2 }}><Typography variant="h6" gutterBottom>Datenschutz</Typography><Divider /></Box>
        <Typography variant="body2" paragraph>
          Wir nehmen den Schutz Ihrer persönlichen Daten sehr ernst. Wir behandeln Ihre personenbezogenen Daten vertraulich und entsprechend der gesetzlichen Datenschutzvorschriften sowie dieser Datenschutzerklärung.
        </Typography>
        <Typography variant="body2">
          Verantwortliche Stelle: Fachschaftsvertretung Informatik, Westfälische Hochschule, Neidenburgerstr. 43, 45897 Gelsenkirchen.
        </Typography>
      </Paper>
    </Page>
  );
}
