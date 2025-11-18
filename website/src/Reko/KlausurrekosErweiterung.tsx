import { useMemo } from "react";
import {
  Box,
  Container,
  Typography,
  Divider,
  Link,
  Paper,
  Chip,
} from "@mui/material";
import { useLocation, useSearchParams } from "react-router-dom";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";

// Datentyp für einen einzelnen Reko-Eintrag
type Entry = {
  title: string;
  date: string;        // Datum im Format "DD.MM.YYYY"
  pos: string[];       // zu welchen POs die Reko gehört, z.B. ["PO16"], ["PO23"], ["PO16","PO23"]
  href?: string;       // später Link auf PDF oder ähnliches
  verified?: boolean;  // von der Fachschaft verifiziert oder nicht
  version?: string;    // Version der Reko, z.B. v1.0, v2.0
};

// Spezielle, echte Rekos nur für "Einführung in die Programmierung"
const EPR_ENTRIES: Entry[] = [
  {
    title: "Einführung in die Programmierung",
    date: "07.07.2024",
    pos: ["PO16"],
    verified: true,
    version: "v1.0",
  },
  {
    title: "Einführung in die Programmierung",
    date: "24.09.2024",
    pos: ["PO23"],
    verified: false,
    version: "v1.1",
  },
  {
    title: "Einführung in die Programmierung",
    date: "15.02.2025",
    pos: ["PO16", "PO23"],
    verified: false,
    version: "v2.0",
  },
];

// Mapping: Modulname → Liste von Rekos
// aktuell nur EPR, weitere Module können später ergänzt werden
const REKO_MAP: Record<string, Entry[]> = {
  "Einführung in die Programmierung": EPR_ENTRIES,
  // z.B. später:
  // "Datenbanksysteme": [...],
};

// Fallback-Eintrag, falls es für ein Modul noch keine echten Rekos gibt
function createDefaultEntry(modul: string, poFromState?: string | null): Entry[] {
  const pos: string[] = [];

  // Wenn PO aus dem State kommt, übernehme ich sie in den Tag
  if (poFromState === "PO16" || poFromState === "PO23") {
    pos.push(poFromState);
  }

  // Fallback: Standard-PO, falls nichts mitgegeben wurde
  if (pos.length === 0) {
    pos.push("PO16");
  }

  return [
    {
      title: modul,
      date: "01.02.2024",
      pos,
      verified: false,
      version: "v1.0",
    },
  ];
}

// Detailseite für die Rekos eines Moduls
export default function KlausurrekosErweiterung() {
  // Location-State (kommt von der Übersicht) + Query-Parameter als Fallback
  const location = useLocation();
  const [params] = useSearchParams();

  // Studiengang wird aktuell nur zu Info/Zukunftszwecken genutzt
  const studiengang =
    (location.state as any)?.studiengang ??
    params.get("sg") ??
    "Informatik";

  // PO (PO16 oder PO23), nutze ich u.a. für Default-Einträge
  const po =
    (location.state as any)?.po ??
    params.get("po") ??
    "PO16";

  // Modulname, ist der zentrale Schlüssel für die Darstellung
  const modul =
    (location.state as any)?.modul ??
    params.get("mod") ??
    "Einführung in die Programmierung";

  // Rekos für das Modul holen und nach Datum sortieren (neueste zuerst)
  const entries: Entry[] = useMemo(() => {
    const fromMap = REKO_MAP[modul];
    const list = fromMap ?? createDefaultEntry(modul, po);

    // Hilfsfunktion: Datum "DD.MM.YYYY" in eine vergleichbare Zahl umwandeln
    const toNum = (d: string) => {
      const [day, month, year] = d.split(".");
      return Number(`${year}${month}${day}`);
    };

    return [...list].sort((a, b) => toNum(b.date) - toNum(a.date));
  }, [modul, po]);

  return (
    // Container für die Detailseite (begrenzte Breite + Abstand nach oben)
    <Container maxWidth="md" sx={{ mt: 8 }}>
      {/* Seitentitel mit Modulnamen */}
      <Typography variant="h4" color="primary" fontWeight={600} gutterBottom>
        Reko: {modul}
      </Typography>
      {/* Kurzbeschreibung oben auf der Seite */}
      <Typography variant="body1" sx={{ mb: 4 }}>
        Hier findest du die vorhandenen Rekonstruktionen für dieses Modul.
      </Typography>
      {/* Liste der Reko-Karten */}
      <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
        {entries.map((entry, i) => (
          <Paper
            key={`${entry.title}-${entry.date}-${i}`}
            sx={{
              p: 2,
              borderRadius: 2,
              transition: "0.2s",
              "&:hover": { boxShadow: 4 },
            }}
          >
            {/* Kopfbereich der Karte: Titel, PO-Tags und Verifizierungs-Haken */}
            <Box
              sx={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                mb: 0.5,
                flexWrap: "wrap",
                gap: 1,
              }}
            >
              {/* Titel + PO-Tags direkt daneben */}
              <Box
                sx={{
                  display: "flex",
                  alignItems: "center",
                  gap: 1,
                  flexWrap: "wrap",
                }}
              >
                <Typography variant="h6" fontWeight={600} color="secondary">
                  {entry.title}
                </Typography>

                <Box sx={{ display: "flex", gap: 0.5 }}>
                  {entry.pos.map((tag) => (
                    <Chip
                      key={tag}
                      label={tag}
                      size="small"
                      variant="outlined"
                      color="default" // PO16 & PO23 beide neutral (schwarz)
                    />
                  ))}
                </Box>
              </Box>

              {/* Haken zeigt, ob die Reko durch die Fachschaft verifiziert ist */}
              {entry.verified && (
                <CheckCircleIcon sx={{ color: "green", fontSize: 26 }} />
              )}
            </Box>

            {/* Datum der Klausur / Reko */}
            <Typography variant="body2" color="text.secondary">
              {entry.date}
            </Typography>

            {/* Link zur eigentlichen Reko + Anzeige der Version */}
            <Box
              sx={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                mt: 1,
              }}
            >
              <Link href={entry.href ?? "#"} color="primary">
                Reko ansehen
              </Link>

              {entry.version && (
                <Typography
                  variant="body2"
                  color="text.secondary"
                  sx={{ fontStyle: "italic" }}
                >
                  {entry.version}
                </Typography>
              )}
            </Box>
          </Paper>
        ))}
      </Box>
      <Divider sx={{ my: 4 }} />
      {/* Legende erklärt Haken, PO-Tags und Versionierung */}
      <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
        <Typography variant="body2" color="text.secondary">
          <CheckCircleIcon sx={{ color: "green", fontSize: 18, mr: 1 }} />
          = verifiziert durch die Fachschaft
        </Typography>

        <Typography
          variant="body2"
          color="text.secondary"
          sx={{ display: "flex", alignItems: "center", gap: 1 }}
        >
          <Chip label="PO16" size="small" variant="outlined" />
          <Chip label="PO23" size="small" variant="outlined" />
          = Prüfungsordnung(en)
        </Typography>

        <Typography variant="body2" color="text.secondary">
          <strong>Version (z. B. v1.0, v2.0)</strong> zeigt an, wie aktuell die Rekonstruktion ist.
        </Typography>
      </Box>
      {/* Optional: Debug-Info, falls ich später prüfen möchte, was ankommt
      <Typography variant="caption" color="text.disabled">
        Debug: {studiengang} • {po}
      </Typography>
      */}
    </Container>
  );
}