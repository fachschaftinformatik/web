export type ForumPost = {
  id: string;
  title: string;
  excerpt: string;
  createdAt: string;
  replies: number;
  category: string;
};

export const mockForumPosts: ForumPost[] = [
  {
    id: '1',
    title: 'Projektideen für das Sommersemester',
    excerpt:
      'Ich arbeite gerade an einem Konzept für eine Event-App. Hat jemand Interesse an einer gemeinsamen Session?',
    createdAt: '2024-02-05T09:15:00Z',
    replies: 12,
    category: 'Projekte',
  },
  {
    id: '2',
    title: 'Vorbereitung auf die Datenbanken-Klausur',
    excerpt:
      'Welche Materialien nutzt ihr? Ich habe die Zusammenfassung aus dem Wiki ergänzt und suche nach weiteren Tipps.',
    createdAt: '2024-02-03T17:42:00Z',
    replies: 7,
    category: 'Studium',
  },
  {
    id: '3',
    title: 'Aftermovie vom Campusfest',
    excerpt: 'Die ersten Bilder sind da! Feedback willkommen, bevor das Video live geht.',
    createdAt: '2024-02-01T11:02:00Z',
    replies: 4,
    category: 'Medien',
  },
  {
    id: '4',
    title: 'Neue Mentoring-Runde startet',
    excerpt:
      'Wir suchen Mentor:innen für Erstsemester. Wer hat Zeit für ein kurzes Briefing nächste Woche?',
    createdAt: '2024-01-28T19:27:00Z',
    replies: 18,
    category: 'Community',
  },
  {
    id: '5',
    title: 'Jobangebot: Werkstudent:in Backend',
    excerpt: 'Ein Partnerunternehmen in GE sucht Unterstützung für Go-Services. PN bei Interesse.',
    createdAt: '2024-01-25T08:36:00Z',
    replies: 3,
    category: 'Jobs',
  },
  {
    id: '6',
    title: 'Workshop: UX Basics',
    excerpt: 'Am Freitag findet ein kurzer Workshop zu UX-Research statt. Plätze sind begrenzt.',
    createdAt: '2024-01-22T15:10:00Z',
    replies: 9,
    category: 'Events',
  },
  {
    id: '7',
    title: 'Suche Team für Hackathon',
    excerpt: 'Wer möchte spontan beim RuhrHack teilnehmen? Frontend/Design wäre super.',
    createdAt: '2024-01-20T09:48:00Z',
    replies: 5,
    category: 'Community',
  },
  {
    id: '8',
    title: 'Hardware-Verleih aktualisiert',
    excerpt: 'Neue Kameras verfügbar. Bitte einmal Feedback geben, ob alles läuft.',
    createdAt: '2024-01-18T13:55:00Z',
    replies: 2,
    category: 'Organisation',
  },
  {
    id: '9',
    title: 'Erfahrungen mit Go-Praktikum?',
    excerpt: 'Ich starte bald im Backend-Team und freue mich über Tipps für den Einstieg.',
    createdAt: '2024-01-16T08:20:00Z',
    replies: 6,
    category: 'Jobs',
  },
  {
    id: '10',
    title: 'Semesterplanung 2024',
    excerpt: 'Welche Events stehen an? Lasst uns eine Übersicht sammeln.',
    createdAt: '2024-01-12T19:05:00Z',
    replies: 11,
    category: 'Organisation',
  },
];

// data.ts

export type EventItem = { id: number; title: string; src: string };
export type Bild = { id: number; title: string; thumb: string; full: string };

// --- HILFSFUNKTION FÜR BEISPIELBILDER ---
const pic = (seed: number, w: number, h: number) =>
  `https://picsum.photos/seed/${seed}/${w}/${h}`;

// --- HIER SIND DEINE LOKALEN BILDER & BEISPIELBILDER ---

export const events: EventItem[] = [
  // --- EVENT 1 (DEIN ORDNER 1) ---
  // Ersetze DEIN_ORDNER_1 und den Dateinamen
  { id: 1, title: "LAN-Party 2023", src: "/LAN-Party 2023/1.jpg" },

  // --- EVENT 2 (DEIN ORDNER 2) ---
  // Ersetze DEIN_ORDNER_2 und den Dateinamen
  { id: 2, title: "LAN-Party 2024", src: "/LAN-Party 2024/2024-11-29_23.07.08-_PHI1930-klaushardt.com.jpg" },

  // --- EVENTS 3-6 (BEISPIELBILDER) ---
  { id: 3, title: "Weinprobe", src: pic(103, 400, 600) },
  { id: 4, title: "Kulturabend", src: pic(104, 400, 600) },
  { id: 5, title: "Weihnachtsfeier", src: pic(105, 400, 600) },
  { id: 6, title: "Sporttag", src: pic(106, 400, 600) },
];

export const IMAGES_PER_PAGE = 20;

export const bilderByEvent: Record<number, Bild[]> = {
  // --- EVENT 1 (DEIN ORDNER 1) ---
  // Ersetze DEIN_ORDNER_1 und die Dateinamen
  1: [
    { id: 101, title: "", thumb: "/LAN-Party 2023/1.jpg", full: "/LAN-Party 2023/1.jpg" },
    { id: 102, title: "", thumb: "/LAN-Party 2023/1688fad1-29d5-48c3-9167-9af3553ef10f.jpg", full: "/LAN-Party 2023/1688fad1-29d5-48c3-9167-9af3553ef10f.jpg" },
    { id: 103, title: "", thumb: "/LAN-Party 2023/ad23f7da-c6cd-4a96-9fce-8b5d0ccd8cf8.jpg", full: "/LAN-Party 2023/ad23f7da-c6cd-4a96-9fce-8b5d0ccd8cf8.jpg" },
    { id: 104, title: "", thumb: "/LAN-Party 2023/b4818f1d-72e6-45ab-8f65-17e2c84b9c77.jpg", full: "/LAN-Party 2023/b4818f1d-72e6-45ab-8f65-17e2c84b9c77.jpg" },
    { id: 105, title: "", thumb: "/LAN-Party 2023/c7cdd3e2-6b75-4d06-805d-1f761b359dd7.jpg", full: "/LAN-Party 2023/c7cdd3e2-6b75-4d06-805d-1f761b359dd7.jpg" },
    { id: 106, title: "", thumb: "/LAN-Party 2023/DSC00061.jpg", full: "/LAN-Party 2023/DSC00061.jpg" },
    { id: 107, title: "", thumb: "/LAN-Party 2023/DSC00097.jpg", full: "/LAN-Party 2023/DSC00097.jpg" },
    { id: 108, title: "", thumb: "/LAN-Party 2023/DSC00114.jpg", full: "/LAN-Party 2023/DSC00114.jpg" },
    { id: 109, title: "", thumb: "/LAN-Party 2023/DSC00132.jpg", full: "/LAN-Party 2023/DSC00132.jpg" },
    { id: 110, title: "", thumb: "/LAN-Party 2023/DSC00140.jpg", full: "/LAN-Party 2023/DSC00140.jpg" },
    { id: 111, title: "", thumb: "/LAN-Party 2023/DSC00146.jpg", full: "/LAN-Party 2023/DSC00146.jpg" },
    { id: 112, title: "", thumb: "/LAN-Party 2023/DSC00170.jpg", full: "/LAN-Party 2023/DSC00170.jpg" },
    { id: 113, title: "", thumb: "/LAN-Party 2023/DSC00173.jpg", full: "/LAN-Party 2023/DSC00173.jpg" },

    // ...
  ],
  
  // --- EVENT 2 (DEIN ORDNER 2) ---
  // Ersetze DEIN_ORDNER_2 und die Dateinamen
  2: [
    { id: 201, title: "", thumb: "/LAN-Party 2024/1.jpg", full: "/LAN-Party 2024/1.jpg" },
    { id: 202, title: "", thumb: "/LAN-Party 2024/2.jpg", full: "/LAN-Party 2024/2.jpg" },
    { id: 203, title: "", thumb: "/LAN-Party 2024/3.jpg", full: "/LAN-Party 2024/3.jpg" },
    { id: 204, title: "", thumb: "/LAN-Party 2024/4.jpg", full: "/LAN-Party 2024/4.jpg" },
    { id: 205, title: "", thumb: "/LAN-Party 2024/5.jpg", full: "/LAN-Party 2024/5.jpg" },
    { id: 206, title: "", thumb: "/LAN-Party 2024/6.jpg", full: "/LAN-Party 2024/6.jpg" },
    { id: 207, title: "", thumb: "/LAN-Party 2024/7.jpg", full: "/LAN-Party 2024/7.jpg" },
    { id: 208, title: "", thumb: "/LAN-Party 2024/8.jpg", full: "/LAN-Party 2024/8.jpg" },
    { id: 209, title: "", thumb: "/LAN-Party 2024/9.jpg", full: "/LAN-Party 2024/9.jpg" },
    { id: 210, title: "", thumb: "/LAN-Party 2024/10.jpg", full: "/LAN-Party 2024/10.jpg" },
    { id: 211, title: "", thumb: "/LAN-Party 2024/11.jpg", full: "/LAN-Party 2024/11.jpg" },
    { id: 212, title: "", thumb: "/LAN-Party 2024/12.jpg", full: "/LAN-Party 2024/12.jpg" },
    { id: 213, title: "", thumb: "/LAN-Party 2024/13.jpg", full: "/LAN-Party 2024/13.jpg" },
    { id: 214, title: "", thumb: "/LAN-Party 2024/14.jpg", full: "/LAN-Party 2024/14.jpg" },
    // ...
  ],

  // Die restlichen Events werden automatisch mit Beispielbildern befüllt
};

// --- EVENTS 3-6 (BEISPIELBILDER) ---
// Füllt `bilderByEvent` für alle Events > 2 automatisch mit picsum-Bildern
events.forEach(event => {
  if (event.id > 2) { // <-- Geändert auf 2
    bilderByEvent[event.id] = Array.from({ length: 30 }).map((_, i) => {
      const seed = event.id * 100 + i;
      return {
        id: seed,
        title: `${event.title} Bild ${i + 1}`,
        thumb: pic(seed, 600, 400),
        full: pic(seed, 1600, 1066),
      };
    });
  }
});

// --- HELPER FUNKTIONEN ---

export const getImageTitle = (image?: Bild) => {
  if (!image) return "";
  const trimmed = image.title?.trim();
  if (trimmed) return trimmed;
  const candidate = image.full || image.thumb;
  return candidate.split("/").filter(Boolean).pop() ?? "Bild";
};

export const defaultTitleForFile = (file: File) => {
  const name = file.name;
  const withoutExtension = name.replace(/\.[^/.]+$/, "");
  return withoutExtension || name;
};

// Export der pic-Funktion für die Verwendung im Komponenten-State, falls nötig
export { pic };
