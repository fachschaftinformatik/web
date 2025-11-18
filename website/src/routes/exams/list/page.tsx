import { useMemo, useState } from "react";
import {
  Box,
  Container,
  Typography,
  Button,
  OutlinedInput,
  InputAdornment,
  IconButton,
  Divider,
  Menu,
  MenuItem,
} from "@mui/material";

import ArrowDropDownRoundedIcon from "@mui/icons-material/ArrowDropDownRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import { Link as RouterLink } from "react-router-dom";

type Semester = { title: string; items: string[] };
const piPO16Semesters: Semester[] = [
  {
    title: "1. Semester",
    items: [
      "Technisches Englisch für Informatiker",
      "Technische Grundlagen der Informatik",
      "Grundlagen der Mathematik für Informatiker",
      "Einführung in die Programmierung",
      "Logik und diskrete Strukturen",
    ],
  },
  {
    title: "2. Semester",
    items: [
      "Mathematik für Informatiker",
      "Rechnernetze",
      "Objektorientierte Programmierung",
      "Algorithmen und Datenstrukturen",
      "Theoretische Informatik",
    ],
  },
  {
    title: "3. Semester",
    items: [
      "Betriebssysteme",
      "Internet-Sprachen",
      "Softwaretechnik",
      "Datenbanksysteme",
      "Mensch-Computer-Interaktion",
    ],
  },
  { title: "4. Semester", items: ["Softwareprojekt Informatik", "Internet-Datenbanken", "Internet-Protokolle"] },
  { title: "5. Semester", items: ["Softwareprojekt Informatik", "Prozedurale Programmierung"] },
];

// Studienverlauf Wirtschaftsinformatik, PO16
const wiPO16Semesters: Semester[] = [
  {
    title: "1. Semester",
    items: [
      "Grundlagen der Wirtschaftsinformatik",
      "Einführung in die Betriebswirtschaftslehre",
      "Grundlagen der Mathematik für Informatiker",
      "Einführung in die Programmierung",
      "Logik und diskrete Strukturen",
    ],
  },
  {
    title: "2. Semester",
    items: [
      "Mathematik für Wirtschaftsinformatiker",
      "Produktion und Materialwirtschaft",
      "Wirtschaftsenglisch für Wirtschaftsinformatiker",
      "Objektorientierte Programmierung",
      "Algorithmen und Datenstrukturen",
    ],
  },
  {
    title: "3. Semester",
    items: [
      "Projektmanagement",
      "Betriebssysteme und Netzwerke",
      "Mensch-Computer-Interaktion",
      "Softwaretechnik",
      "Datenbanksysteme",
    ],
  },
  {
    title: "4. Semester",
    items: [
      "Softwareprojekt Wirtschaftsinformatik",
      "Betriebliches Rechnungswesen",
      "Betriebliche Informationssysteme 1",
      "Geschäftsprozessmanagement",
    ],
  },
  {
    title: "5. Semester",
    items: ["Softwareprojekt Wirtschaftsinformatik", "IT-Recht", "Betriebliche Informationssysteme 2", "Digitales Marketing"],
  },
];

// Studienverlauf Informatik, PO23
const piPO23Semesters: Semester[] = [
  {
    title: "1. Semester",
    items: [
      "Technisches Englisch für Informatiker",
      "Technische Grundlagen der Informatik",
      "Mathematische Grundlagen",
      "Einführung in die Programmierung",
      "Logik und diskrete Strukturen",
    ],
  },
  {
    title: "2. Semester",
    items: [
      "Statistik und Lineare Algebra",
      "Betriebssysteme",
      "Objektorientierte Programmierung",
      "Algorithmen und Datenstrukturen",
      "Theoretische Informatik",
    ],
  },
  {
    title: "3. Semester",
    items: ["Rechnernetze", "Internetsprachen", "Softwaretechnik", "Datenbanksysteme", "Mensch-Computer-Interaktion"],
  },
];

// Studienverlauf Wirtschaftsinformatik, PO23
const wiPO23Semesters: Semester[] = [
  {
    title: "1. Semester",
    items: [
      "Logik und diskrete Strukturen",
      "Einführung in die Programmierung",
      "Mathematische Grundlagen",
      "Einführung in die Betriebswirtschaftslehre",
      "Grundlagen der Wirtschaftsinformatik",
    ],
  },
  {
    title: "2. Semester",
    items: [
      "Algorithmen und Datenstrukturen",
      "Objektorientierte Programmierung",
      "Statistik und Lineare Algebra",
      "Produktion und Materialwirtschaft",
      "Wirtschaftsenglisch",
    ],
  },
  {
    title: "3. Semester",
    items: [
      "Datenbanksysteme",
      "Softwaretechnik",
      "Mensch-Computer-Interaktion",
      "Betriebssysteme und Netzwerke",
      "Projektmanagement",
    ],
  },
  {
    title: "4. Semester",
    items: [
      "Softwareprojekt Wirtschaftsinformatik",
      "IT-Recht",
      "Grundlagen Supply Chain Management",
      "Geschäftsprozessmanagement",
    ],
  },
  {
    title: "5. Semester",
    items: [
      "Softwareprojekt Wirtschaftsinformatik",
      "Betriebliches Rechnungswesen",
      "Supply Chain Management und Digitalisierung",
      "Digitales Marketing",
    ],
  },
];

// Wahlpflichtmodule für Informatik, PO16
const electivesPI_PO16 = [
  "Betrieb komplexer verteilter Systeme",
  "Betriebswirtschaftslehre für Informatiker",
  "Bildverarbeitung",
  "Grundlagen der IT-Sicherheit",
  "IT-Recht",
  "Komponentenbasierte Softwareentwicklung",
  "Künstliche Intelligenz",
  "Mobile Computing",
  "Mobile Robotik",
  "Objektorientierte Programmierung mit C++",
  "Parallele Programmierung",
  "Practical Security Attacks and Exploitation",
  "Prozedurale Programmierung",
  "Software Design",
  "Medientechnik",
  "Modellbasierter Entwurf von Regelsystemen",
  "Systemtheorie",
  "Zeitdiskrete Regelsysteme",
];

// Wahlpflichtmodule für Wirtschaftsinformatik, PO16
const electivesWI_PO16 = [
  "Betrieb komplexer verteilter Systeme",
  "Entwicklung von Informationssystemen",
  "Grundlagen der IT-Sicherheit",
  "Internet-Datenbanken",
  "Internet-Sprachen",
  "Mobile Computing",
  "Practical Security Attacks and Exploitation",
  "Prozedurale Programmierung",
  "Software Design",
];

// Wahlpflichtmodule für Informatik, PO23
const electivesPI_PO23 = [
  "Betrieb komplexer verteilter Systeme",
  "Einführung in die Bildverarbeitung",
  "Data on the Web",
  "Data Science in Practice",
  "Einführung in die medizinische Informatik",
  "Einführung in die Robotik",
  "Internet-Protokolle",
  "IT-Recht",
  "Grundlagen der IT Sicherheit",
  "Komponentenbasierte Softwareentwicklung",
  "Künstliche Intelligenz",
  "Knowledge Graphs",
  "Mobile Application Development",
  "Mobile and Cloud Computing",
  "Mobile Robotik",
  "Parallele Programmierung",
  "Prozedurale Programmierung",
  "Practical Security Attacks and Exploitation",
  "Software Design",
  "Betriebliches Rechnungswesen",
  "Digitales Marketing",
  "Einführung in die Betriebswirtschaftslehre",
  "Geschäftsprozessmanagement",
  "Grundlagen der Wirtschaftsinformatik",
  "Angewandte Netzwerksicherheit",
  "Projektmanagement",
  "Produktion und Materialwirtschaft",
];

// Wahlpflichtmodule für Wirtschaftsinformatik, PO23
const electivesWI_PO23 = [
  "Betrieb komplexer verteilter Systeme",
  "Einführung in die Bildverarbeitung",
  "Data on the Web",
  "Einführung in die Robotik",
  "Internet-Protokolle",
  "Internet-Sprachen",
  "Grundlagen der IT Sicherheit",
  "Komponentenbasierte Softwareentwicklung",
  "Knowledge Graphs",
  "Mobile Application Development",
  "Mobile Robotik",
  "Angewandte Netzwerksicherheit",
  "Practical Security Attacks and Exploitation",
  "Software Design",
];

export default function Exams() {
  const [studiengangAnchor, setStudiengangAnchor] = useState<null | HTMLElement>(null);
  const [studiengang, setStudiengang] = useState("Studiengang");
  const [poAnchor, setPoAnchor] = useState<null | HTMLElement>(null);
  const [selectedPo, setSelectedPo] = useState("PO16");
  const [search, setSearch] = useState("");

  const semestersToShow = useMemo(() => {
    if (studiengang === "Informatik" && selectedPo === "PO16") return piPO16Semesters;
    if (studiengang === "Wirtschaftsinformatik" && selectedPo === "PO16") return wiPO16Semesters;
    if (studiengang === "Informatik" && selectedPo === "PO23") return piPO23Semesters;
    if (studiengang === "Wirtschaftsinformatik" && selectedPo === "PO23") return wiPO23Semesters;
    return null;
  }, [studiengang, selectedPo]);

  // Konfiguration, ab welchem Semesterblock die Wahlpflichtmodule angezeigt werden sollen
  const electivesConfig = useMemo(() => {
    if (selectedPo === "PO16" && studiengang === "Informatik")
      return { afterTitle: "5. Semester", list: electivesPI_PO16 };
    if (selectedPo === "PO16" && studiengang === "Wirtschaftsinformatik")
      return { afterTitle: "5. Semester", list: electivesWI_PO16 };
    if (selectedPo === "PO23" && studiengang === "Informatik")
      return { afterTitle: "3. Semester", list: electivesPI_PO23 };
    if (selectedPo === "PO23" && studiengang === "Wirtschaftsinformatik")
      return { afterTitle: "5. Semester", list: electivesWI_PO23 };
    return null;
  }, [studiengang, selectedPo]);

  return (
    // Container begrenzt die Breite und sorgt für Abstand nach oben
    <Container maxWidth="md" sx={{ mt: 6 }}>
      {/* Seitentitel */}
      <Typography variant="h4" color="primary" fontWeight={600} gutterBottom>
        Klausurrekonstruktionen
      </Typography>
      {/* Menübereich: Studiengang, PO und Suchfeld */}
      <Box sx={{ display: "flex", gap: 2, flexWrap: "wrap", mb: 4 }}>
        {/* Auswahl: Studiengang */}
        <Box>
          <Button
            variant="outlined"
            endIcon={<ArrowDropDownRoundedIcon />}
            onClick={(e) => setStudiengangAnchor(e.currentTarget)}
          >
            {studiengang}
          </Button>
          <Menu
            anchorEl={studiengangAnchor}
            open={Boolean(studiengangAnchor)}
            onClose={() => setStudiengangAnchor(null)}
          >
            <MenuItem
              onClick={() => {
                setStudiengang("Informatik");
                setStudiengangAnchor(null);
              }}
            >
              Informatik
            </MenuItem>
            <MenuItem
              onClick={() => {
                setStudiengang("Wirtschaftsinformatik");
                setStudiengangAnchor(null);
              }}
            >
              Wirtschaftsinformatik
            </MenuItem>
          </Menu>
        </Box>

        {/* Auswahl: Prüfungsordnung (PO) */}
        <Box>
          <Button
            variant="outlined"
            endIcon={<ArrowDropDownRoundedIcon />}
            onClick={(e) => setPoAnchor(e.currentTarget)}
          >
            {selectedPo}
          </Button>
          <Menu
            anchorEl={poAnchor}
            open={Boolean(poAnchor)}
            onClose={() => setPoAnchor(null)}
          >
            <MenuItem
              onClick={() => {
                setSelectedPo("PO23");
                setPoAnchor(null);
              }}
            >
              PO23
            </MenuItem>
            <MenuItem
              onClick={() => {
                setSelectedPo("PO16");
                setPoAnchor(null);
              }}
            >
              PO16
            </MenuItem>
          </Menu>
        </Box>

        {/* Suchfeld für zukünftige Filterfunktion nach Modulen */}
        <OutlinedInput
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Suche nach Modul..."
          sx={{ flexGrow: 1, maxWidth: 300 }}
          endAdornment={
            <InputAdornment position="end">
              <IconButton>
                <SearchRoundedIcon />
              </IconButton>
            </InputAdornment>
          }
        />
      </Box>
      {/* Anzeige der Semester und Module */}
      <Box>
        {semestersToShow ? (
          semestersToShow.map((sem, i) => (
            <Box key={sem.title} sx={{ mb: 3 }}>
              {/* Überschrift pro Semesterblock */}
              <Typography variant="h6" fontWeight={600} color="secondary" gutterBottom>
                {sem.title}
              </Typography>

              {/* Modulliste innerhalb des Semesters */}
              <Box sx={{ pl: 2 }}>
                {sem.items.map((label) => (
                  <Typography
                    key={label}
                    component={RouterLink}
                    to="/rekos/klausuren/modul"
                    // über state gebe ich Studiengang, PO und Modulname an die Detailseite weiter
                    state={{
                      studiengang,
                      po: selectedPo,
                      modul: label,
                    }}
                    sx={{
                      display: "block",
                      mb: 1,
                      color: "text.primary",
                      textDecoration: "none",
                      "&:hover": { textDecoration: "underline", color: "primary.main" },
                    }}
                  >
                    {label}
                  </Typography>
                ))}
              </Box>

              {/* Trennlinie zwischen den Semestern */}
              {i < semestersToShow.length - 1 && <Divider sx={{ mt: 2, mb: 2 }} />}

              {/* Wahlpflichtmodule werden nach einem bestimmten Semesterblock eingefügt */}
              {electivesConfig && sem.title === electivesConfig.afterTitle && (
                <>
                  <Divider sx={{ mt: 3, mb: 2 }} />
                  <Typography variant="h6" fontWeight={700} color="secondary" gutterBottom>
                    Wahlpflichtmodule
                  </Typography>
                  <Box sx={{ pl: 2 }}>
                    {electivesConfig.list.map((label) => (
                      <Typography
                        key={label}
                        component={RouterLink}
                        to="/rekos/klausuren/modul"
                        state={{
                          studiengang,
                          po: selectedPo,
                          modul: label,
                        }}
                        sx={{
                          display: "block",
                          mb: 1,
                          color: "text.primary",
                          textDecoration: "none",
                          "&:hover": { textDecoration: "underline", color: "primary.main" },
                        }}
                      >
                        {label}
                      </Typography>
                    ))}
                  </Box>
                </>
              )}
            </Box>
          ))
        ) : (
          // Fallback, wenn noch kein Studiengang/PO gewählt ist
          (<Typography sx={{ opacity: 0.7 }}>Bitte oben passenden <strong>Studiengang</strong>und <strong>PO</strong>auswählen.
                      </Typography>)
        )}
      </Box>
    </Container>
  );
}
