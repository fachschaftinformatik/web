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
