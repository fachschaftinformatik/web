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

export default function StartseiteRekos() {
  return (
    <>
      {/* Hauptcontainer der Seite (begrenzte Breite und Abstand nach oben) */}
      <Container maxWidth="md" sx={{ mt: 6 }}>
        
        {/* Seitentitel */}
        <Typography variant="h4" color="primary" fontWeight={600} gutterBottom>
          Klausurrekonstruktionen
        </Typography>

        {/* Introtext, der erklärt, worum es auf der Seite geht */}
        <Typography variant="body1" sx={{ mb: 3 }}>
          Willkommen bei den Klausurrekonstruktionen der Fachschaft Informatik.  
          Hier findest du von Studierenden erstellte Zusammenfassungen vergangener
          Prüfungen, um dich besser auf kommende Klausuren vorzubereiten.
        </Typography>

        <Divider sx={{ mb: 4 }} />

        {/* Abschnitt: Erklärung, was Rekos überhaupt sind */}
        <Box sx={{ mb: 4 }}>
          {/* Überschrift des Abschnitts */}
          <Typography variant="h5" color="secondary" fontWeight={600} gutterBottom>
            Was sind Rekos?
          </Typography>

          {/* Beschreibung der Idee hinter Rekonstruktionen */}
          <Typography variant="body1" sx={{
            marginBottom: "16px"
          }}>
            Eine Klausurrekonstruktion ist ein nachträglich erstellter Bericht über
            eine geschriebene Prüfung. Sie hilft anderen Studierenden, sich auf
            kommende Klausuren vorzubereiten, indem sie einen Eindruck von
            Aufgabenstellungen und Themen gibt.
          </Typography>
          {/* Hinweis, dass Rekos keine offiziellen Unterlagen sind */}
          <Typography variant="body1" sx={{
            marginBottom: "16px"
          }}>
            Die Rekos sind keine offiziellen Prüfungsunterlagen, sondern
            freiwillige Beiträge von Studierenden für Studierende.
          </Typography>
        </Box>
        <Divider sx={{ mb: 4 }} />
        {/* Abschnitt: Wie Studierende mithelfen und eigene Rekos einreichen können */}
        <Box sx={{ mb: 4 }}>
          {/* Überschrift */}
          <Typography variant="h5" color="secondary" fontWeight={600} gutterBottom>
            Wie kannst du mithelfen?
          </Typography>

          {/* Kurze Erklärung */}
          <Typography variant="body1" sx={{
            marginBottom: "16px"
          }}>
            Damit die Sammlung aktuell bleibt, freuen wir uns über jede neue Reko.
            Mitmachen ist ganz einfach:
          </Typography>
          {/* Schritt-für-Schritt Anleitung */}
          <List sx={{ pl: 2 }}>
            <ListItem>
              1. Vorlage herunterladen oder bei der Fachschaft abholen.
            </ListItem>
            <ListItem>
              2. Nach deiner Klausur die Aufgaben so gut wie möglich rekonstruieren.
            </ListItem>
            <ListItem>
              3. Deine Reko per E-Mail an{" "}
              <Link href="mailto:lernmaterialien@fachschaftinformatik.de">
                lernmaterialien@fachschaftinformatik.de
              </Link>{" "}
              senden oder persönlich abgeben.
            </ListItem>
          </List>

          {/* Hinweis: kleine Belohnung für eingereichte Rekos */}
          <Typography variant="body2" sx={{ mt: 2 }}>
            Für eingereichte Rekos gibt es eine kleine Belohnung als Dankeschön.
          </Typography>
        </Box>

        <Divider sx={{ mb: 4 }} />

        {/* Letzter Abschnitt: Aufruf, zur Übersicht zu gehen */}
        <Box sx={{ textAlign: "center" }}>
          {/* Text über dem Button */}
          <Typography variant="h6" fontWeight={600} gutterBottom>
            Bereit, loszulegen?
          </Typography>

          {/* Kurze Beschreibung */}
          <Typography variant="body1" sx={{ mb: 3 }}>
            Hier findest du alle bisherigen Rekonstruktionen sortiert nach
            Semestern und Modulen:
          </Typography>

          {/* Button zur Klausurübersicht */}
          <Button
            variant="contained"
            size="large"
            component={RouterLink}
            to="/rekos/klausuren"
            sx={{ mb: 5 }}
          >
            Zu den Rekos →
          </Button>
        </Box>
      </Container>
    </>
  );
}