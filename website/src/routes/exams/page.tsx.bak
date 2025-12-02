import {
  Box,
  Container,
  Typography,
  Button,
  Divider,
  List,
  ListItem,
  Link,
} from "@mui/material";
import { Link as RouterLink } from "react-router-dom";
import CloudUploadIcon from "@mui/icons-material/CloudUpload";
import { useAuth } from "@lib/auth";
import { Sidebar } from "@components/layout";

export default function Exams() {
  const { user } = useAuth();
  return (
    <Sidebar user={user} title="Rekos">
      <Container maxWidth="md" sx={{ mt: 6 }}>
      <Typography variant="h4" align="center" gutterBottom>Rekos</Typography>

        <Typography variant="body1" sx={{ mb: 3 }}>
          Willkommen bei den Klausurrekonstruktionen der Fachschaft Informatik.  
          Hier findest du von Studierenden erstellte Zusammenfassungen vergangener
          Prüfungen, um dich besser auf kommende Klausuren vorzubereiten.
        </Typography>

        <Divider sx={{ mb: 4 }} />
        <Box sx={{ mb: 4 }}>
          <Typography variant="h5" color="secondary" fontWeight={600} gutterBottom>
            Was sind Rekos?
          </Typography>
          <Typography
            variant="body1"
            sx={{
              marginBottom: "16px",
            }}
          >
            Eine Klausurrekonstruktion ist ein nachträglich erstellter Bericht über
            eine geschriebene Prüfung. Sie hilft anderen Studierenden, sich auf
            kommende Klausuren vorzubereiten, indem sie einen Eindruck von
            Aufgabenstellungen und Themen gibt.
          </Typography>
          <Typography
            variant="body1"
            sx={{
              marginBottom: "16px",
            }}
          >
            Die Rekos sind keine offiziellen Prüfungsunterlagen, sondern
            freiwillige Beiträge von Studierenden für Studierende.
          </Typography>
        </Box>
        <Divider sx={{ mb: 4 }} />
        <Box sx={{ mb: 4 }}>
          <Typography variant="h5" color="secondary" fontWeight={600} gutterBottom>
            Wie kannst du mithelfen?
          </Typography>

          <Typography
            variant="body1"
          >
            Damit die Sammlung aktuell bleibt, freuen wir uns über jede neue Reko.
            Mitmachen ist ganz einfach:
          </Typography>
          <List >
            <ListItem>
              1. Vorlage herunterladen oder bei der Fachschaft abholen.
            </ListItem>
            <ListItem>
              2. Nach deiner Klausur die Aufgaben so gut wie möglich rekonstruieren.
            </ListItem>
            <ListItem>
              3. Deine Reko per E-Mail an
              &nbsp;
              <Link href="mailto:lernmaterialien@fachschaftinformatik.de">
                lernmaterialien@fachschaftinformatik.de
              </Link>
              &nbsp;
              senden oder persönlich abgeben.
            </ListItem>
          </List>
        </Box>
        <Divider sx={{ mb: 4 }} />
        <Box sx={{ textAlign: "center" }}>
          <Button
            variant="contained"
            size="large"
            component={RouterLink}
            to="/exams/list"
            sx={{ mb: 5 }}
          >
            Zu den Rekos →
          </Button>
        </Box>
      </Container>
    </Sidebar>
  );
}
