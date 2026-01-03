export type ForumProgram = "inf-bsc" | "winf-bsc" | "med-bsc" | "inf-msc" | "winf-msc" | "med-msc";

export type NewsItem = {
  id: number;
  title: string;
  date: string;
  createdAt?: string;
  image: string;
  summary: string;
  content: string;
  links: string[];
  isNew: boolean;
  tags: string[];
  pdf?: string | null;
  pdfName?: string | null;
};

export const NEWS_TAGS = ["Alle", "Events", "Jobs", "Prüfungen", "Studium", "Sonstiges"];
export const NEWS_AVAILABLE_TAGS = ["Studium", "Prüfungen", "Events", "Jobs"];

export type CalendarEventCategory = { value: string; label: string; color: string };
export const CALENDAR_EVENT_CATEGORIES: CalendarEventCategory[] = [
  { value: 'Treffen', label: 'Treffen', color: '#2f7d4a' },
  { value: 'Workshop', label: 'Workshop', color: '#3267d8' },
  { value: 'Party', label: 'Party', color: '#e16d48' },
  { value: 'Info', label: 'Info', color: '#d39a3f' },
];

export const STORAGE_KEYS = {
  FORUM_POSTS: "forum-demo-posts",
  FORUM_VOTES: "forum-demo-votes",
  HOMEPAGE_EVENTS: "homepage-events",
  NEWS_LIKED: "newsroom-liked",
  NEWS_CUSTOM: "custom-news",
  TEAM_ADMIN_PREVIEW: "team_admin_preview",
  SIDEBAR_DESKTOP_OPEN: "sidebar_desktop_open",
};

export const NAV_ITEMS_DATA = [
  { label: 'Startseite', href: '/', id: 'home' },
  { label: 'Ankündigungen', href: '/news', id: 'news' },
  { label: 'Rekos', href: '/exams', id: 'exams' },
  { label: 'Forum', href: '/forum', id: 'forum' },
  { label: 'Galerie', href: '/media', id: 'media' },
  { label: 'Team', href: '/team', id: 'team' },
];

export const AVATAR_PALETTE = [
  '#d32f2f', '#c2185b', '#7b1fa2', '#512da8', '#303f9f',
  '#1976d2', '#0288d1', '#0097a7', '#00796b', '#388e3c',
  '#e64a19', '#5d4037', '#455a64',
];

export const APP_CONSTANTS = {
  EMAIL_DOMAIN_FSV: "@fsv-whs.de",
  EMAIL_DOMAIN_STUDMAIL: "@studmail.w-hs.de",
  FORUM_POSTS_PER_PAGE: 20,
  MEDIA_IMAGES_PER_PAGE: 10,
};

export type ForumComment = {
  id: string;
  author: string;
  text: string;
  createdAt: string;
  parentId?: string | null;
};

export type ForumPost = {
  id: string;
  title: string;
  body: string;
  tags: string[];
  author: string;
  createdAt: string;
  votes: number;
  programs: ForumProgram[];
  comments: ForumComment[];
  pinned?: boolean;
};

export type ForumVote = -1 | 0 | 1;

export type ForumProgramMeta = {
  id: ForumProgram;
  label: string;
  shortLabel: string;
  level: "Bachelor" | "Master";
};

export const FORUM_PROGRAM_CATALOG: ForumProgramMeta[] = [
  { id: "inf-bsc", label: "Informatik (B.Sc.)", shortLabel: "INF B.Sc.", level: "Bachelor" },
  { id: "winf-bsc", label: "Wirtschaftsinformatik (B.Sc.)", shortLabel: "WINF B.Sc.", level: "Bachelor" },
  { id: "med-bsc", label: "Medieninformatik (B.Sc.)", shortLabel: "MED B.Sc.", level: "Bachelor" },
  { id: "inf-msc", label: "Informatik (M.Sc.)", shortLabel: "INF M.Sc.", level: "Master" },
  { id: "winf-msc", label: "Wirtschaftsinformatik (M.Sc.)", shortLabel: "WINF M.Sc.", level: "Master" },
  { id: "med-msc", label: "Medieninformatik (M.Sc.)", shortLabel: "MED M.Sc.", level: "Master" },
];

export const FORUM_PROGRAM_META_MAP: Record<ForumProgram, ForumProgramMeta> = FORUM_PROGRAM_CATALOG.reduce((acc, meta) => {
  acc[meta.id] = meta;
  return acc;
}, {} as Record<ForumProgram, ForumProgramMeta>);

export const FORUM_PROGRAMS: ForumProgram[] = FORUM_PROGRAM_CATALOG.map((meta) => meta.id);

export const FORUM_LEGACY_PROGRAM_MAP: Record<string, ForumProgram> = {
  Informatik: "inf-bsc",
  Wirtschaftsinformatik: "winf-bsc",
  Medieninformatik: "med-bsc",
  "Informatik (B.Sc.)": "inf-bsc",
  "Wirtschaftsinformatik (B.Sc.)": "winf-bsc",
  "Medieninformatik (B.Sc.)": "med-bsc",
  "Informatik (M.Sc.)": "inf-msc",
  "Wirtschaftsinformatik (M.Sc.)": "winf-msc",
  "Medieninformatik (M.Sc.)": "med-msc",
};

export const FORUM_STORAGE_KEY = STORAGE_KEYS.FORUM_POSTS;

const forumBaseSeeds: ForumPost[] = [
  {
    id: "p1",
    title: "Wie strukturiert ihr React-Formulare ohne libs?",
    body: "Ich suche einen sauberen Weg für Validierung + Fehlermeldungen ohne Formik/React Hook Form. Gibt es mit MUI Best Practices?",
    tags: ["react", "mui", "forms"],
    author: "Lea",
    createdAt: new Date(Date.now() - 36 * 3600 * 1000).toISOString(),
    votes: 7,
    programs: ["inf-bsc", "inf-msc"],
    comments: [
      { id: "c1", author: "Jonas", text: "Ich nutze Zod + eigene Inputs.", createdAt: new Date(Date.now() - 32 * 3600 * 1000).toISOString() },
      { id: "c2", author: "Mara", text: "React Hook Form ist leichtgewichtig genug.", createdAt: new Date(Date.now() - 30 * 3600 * 1000).toISOString() },
      { id: "c3", author: "Lea", text: "Danke! Hast du ein Beispielrepo?", createdAt: new Date(Date.now() - 29 * 3600 * 1000).toISOString(), parentId: "c1" },
    ],
    pinned: true,
  },
  {
    id: "p2",
    title: "TS: Unterschied zwischen type und interface?",
    body: "Wann würdet ihr type statt interface nutzen? Besonders im Kontext von Union-Types & Declaration-Merging.",
    tags: ["typescript"],
    author: "Jonas",
    createdAt: new Date(Date.now() - 5 * 3600 * 1000).toISOString(),
    votes: 12,
    programs: ["winf-bsc", "inf-msc"],
    comments: [
      {
        id: "c4",
        author: "Timo",
        text: "Ich nehme type sobald Union/Intersection im Spiel ist.",
        createdAt: new Date(Date.now() - 4.5 * 3600 * 1000).toISOString(),
      },
      {
        id: "c5",
        author: "Eva",
        text: "Interfaces fürs Structural Typing in Klassen, Rest mache ich mit type.",
        createdAt: new Date(Date.now() - 4 * 3600 * 1000).toISOString(),
      },
      {
        id: "c6",
        author: "Jonas",
        text: "Makes sense. @Timo nutzt du auch satisfies?",
        createdAt: new Date(Date.now() - 3.5 * 3600 * 1000).toISOString(),
        parentId: "c4",
      },
    ],
  },
  {
    id: "p3",
    title: "useMemo/useCallback - Overhead vs. Nutzen?",
    body: "Gibt es Richtlinien, wann der Overhead größer ist als der Nutzen? Beispiele willkommen.",
    tags: ["react", "performance"],
    author: "Mara",
    createdAt: new Date(Date.now() - 2 * 3600 * 1000).toISOString(),
    votes: 3,
    programs: ["med-bsc", "med-msc", "inf-bsc", "winf-msc"],
    comments: [
      {
        id: "c7",
        author: "Sara",
        text: "Ich nutze useMemo fast nur um schwere Berechnungen zu cachen.",
        createdAt: new Date(Date.now() - 1.8 * 3600 * 1000).toISOString(),
      },
      {
        id: "c8",
        author: "Luca",
        text: "Callbacks nur wenn ich Props in tiefe Komponenten reiche.",
        createdAt: new Date(Date.now() - 1.6 * 3600 * 1000).toISOString(),
      },
      {
        id: "c9",
        author: "Mara",
        text: "Danke euch, ich packe das mal in unser Wiki.",
        createdAt: new Date(Date.now() - 1.5 * 3600 * 1000).toISOString(),
        parentId: "c7",
      },
    ],
  },
];

const makeForumExtraSeeds = (n: number): ForumPost[] => {
  const authors = ["Lea", "Jonas", "Mara", "Timo", "Eva", "Noah", "Sara", "Luca", "Milan", "Nora"];
  const topics = [
    "State-Management mit Context",
    "Routing mit React Router",
    "Vite Build-Tipps",
    "Unit-Testing mit Vitest",
    "MUI Table vs. DataGrid",
    "Responsive Layout mit Grid",
    "Dark Mode mit MUI",
    "Form-Validation Patterns",
    "Performance messen",
    "Code-Splitting & lazy()",
  ];
  const tagsPool = ["react", "typescript", "mui", "routing", "state", "hooks", "performance", "testing", "vite", "ui"];

  const arr: ForumPost[] = [];
  for (let i = 0; i < n; i++) {
    const id = `seed-${i}`;
    const programs: ForumProgram[] =
      i % 7 === 0
        ? FORUM_PROGRAMS.slice()
        : i % 3 === 0
          ? ["inf-bsc", "inf-msc", "winf-bsc"]
          : i % 2 === 0
            ? ["med-bsc", "med-msc"]
            : [FORUM_PROGRAMS[i % FORUM_PROGRAMS.length]];

    let comments: ForumComment[] = [];
    if (i % 5 === 0) {
      comments = [
        {
          id: `${id}-c1`,
          author: "Eva",
          text: "Klingt spannend - hast du ein Repo?",
          createdAt: new Date(Date.now() - (i + 2) * 3600 * 1000).toISOString(),
        },
        {
          id: `${id}-c2`,
          author: "Timo",
          text: "Ich habe letzte Woche etwas ähnliches gebaut.",
          createdAt: new Date(Date.now() - (i + 1.8) * 3600 * 1000).toISOString(),
        },
        {
          id: `${id}-c3`,
          author: "Eva",
          text: "@Timo magst du den Link teilen?",
          createdAt: new Date(Date.now() - (i + 1.6) * 3600 * 1000).toISOString(),
          parentId: `${id}-c2`,
        },
      ];
    } else if (i % 5 === 2) {
      comments = [
        {
          id: `${id}-c1`,
          author: "Jonas",
          text: "Nutze hier unbedingt Lazy Loading.",
          createdAt: new Date(Date.now() - (i + 2.2) * 3600 * 1000).toISOString(),
        },
      ];
    }

    arr.push({
      id,
      title: `Demo #${i + 1}: ${topics[i % topics.length]}`,
      body: "Dies ist ein Demo-Beitrag zum Testen von Suche, Sortierung, Filter, Votes, Erstellen und verschachtelten Kommentaren.",
      tags: [tagsPool[i % tagsPool.length], tagsPool[(i + 3) % tagsPool.length]],
      author: authors[i % authors.length],
      createdAt: new Date(Date.now() - (i + 4) * 2 * 3600 * 1000).toISOString(),
      votes: (i * 7) % 25,
      programs,
      comments,
    });
  }
  return arr;
};

export const FORUM_SEED_POSTS: ForumPost[] = [...forumBaseSeeds, ...makeForumExtraSeeds(20)];

export type ForumPostSummary = {
  id: string;
  title: string;
  excerpt: string;
  createdAt: string;
  replies: number;
  category: string;
};

export const FORUM_POST_SUMMARIES: ForumPostSummary[] = FORUM_SEED_POSTS.map((post) => ({
  id: post.id,
  title: post.title,
  excerpt: post.body,
  createdAt: post.createdAt,
  replies: post.comments.length,
  category: post.tags[0] ?? "Allgemein",
}));

export type EventItem = { id: number; title: string; src: string };
export type Bild = { id: number; title: string; thumb: string; full: string };

export const pic = (seed: number, w: number, h: number) => `https://picsum.photos/seed/${seed}/${w}/${h}`;

export const NEWS_DATA: NewsItem[] = [
  {
    id: 1,
    title: "6-Year-Old Horse Dies at Belmont Park After Race Injury",
    date: "September 14, 2016",
    image: pic(201, 400, 200),
    summary: "A tragic accident occurred during the race, raising concerns about animal safety...",
    content: "Lorem ipsum dolor sit amet...",
    links: ["https://conference2025.com", "https://more-info.com"],
    isNew: true,
    tags: ["Events", "Jobs"]
  },
  {
    id: 2,
    title: "Tech Conference 2025 Kicks Off in Berlin",
    date: "September 12, 2025",
    image: pic(202, 400, 200),
    summary: "Industry leaders gather to discuss the future of AI, robotics, and quantum computing...",
    content: "",
    links: ["https://conference2025.com"],
    isNew: true,
    tags: ["Prüfungen", "Studium"]
  },
  {
    id: 3,
    title: "New Climate Agreement Signed",
    date: "September 10, 2025",
    image: pic(203, 400, 200),
    summary: "World leaders agreed on a historic pact aiming to reduce emissions worldwide...",
    content: "",
    links: [],
    isNew: true,
    tags: ["Events", "Studium"]
  },
  {
    id: 4,
    title: "-Year-Old Horse Dies at Belmont Park After Race Injury",
    date: "September 8, 2025",
    image: pic(204, 400, 200),
    summary: "Scientists reported a promising new therapy with encouraging early trial results...",
    content: "",
    links: [],
    isNew: false,
    tags: ["Studium"]
  },
  {
    id: 5,
    title: "Breakthrough in Cancer Research",
    date: "September 8, 2025",
    image: pic(205, 400, 200),
    summary: "Scientists reported a promising new therapy with encouraging early trial results...",
    content: "",
    links: [],
    isNew: true,
    tags: ["Events", "Studium"]
  },
];


export type TeamMember = {
  id: number;
  name: string;
  email: string;
  img: string | null;
};

export type TeamSection = {
  id: string;
  title: string;
  members: TeamMember[];
};

export const teamSections: TeamSection[] = [
  {
    id: "vorstand",
    title: "Vorstand",
    members: [
      { id: 1, name: "Eric König (Vorsitz)", email: "vorsitz.informatik@fsv-whs.de", img: null },
      { id: 2, name: "Hannah Willemsen (stellv. Vorsitz)", email: "stellv.vorsitz.informatik@fsv-whs.de", img: null },
      { id: 3, name: "Leon Pearse (Finanzen)", email: "finanzen.informatik@fsv-whs.de", img: null },
    ],
  },
  {
    id: "kultur",
    title: "Kultur",
    members: [
      { id: 4, name: "Hannah Willemsen", email: "kultur.informatik@fsv-whs.de", img: null },
      { id: 5, name: "Caner Bayram", email: "kultur.informatik@fsv-whs.de", img: null },
      { id: 6, name: "Leon Weihrauch", email: "kultur.informatik@fsv-whs.de", img: null },
      { id: 7, name: "Yannic Leismann", email: "kultur.informatik@fsv-whs.de", img: null },
      { id: 8, name: "Daniel Chudaska", email: "kultur.informatik@fsv-whs.de", img: null },
    ],
  },
  {
    id: "infrastruktur",
    title: "Infrastruktur",
    members: [
      { id: 9, name: "Francesco La Camera", email: "edv.informatik@fsv-whs.de", img: null },
    ],
  },
  {
    id: "hochschulpolitik",
    title: "Hochschulpolitik",
    members: [
      { id: 10, name: "Mike Drost", email: "hopo.informatik@fsv-whs.de", img: null },
    ],
  },
  {
    id: "oeffentlichkeit",
    title: "Öffentlichkeit",
    members: [
      { id: 11, name: "Linda Gudaqi", email: "oeffentlichkeit.informatik@fsv-whs.de", img: null },
      { id: 12, name: "Hong Nguyen", email: "oeffentlichkeit.informatik@fsv-whs.de", img: null },
    ],
  },
  {
    id: "beschaffung",
    title: "Beschaffung",
    members: [
      { id: 13, name: "Marcel Roith", email: "beschaffung.informatik@fsv-whs.de", img: null },
      { id: 14, name: "Alan Said Suleiman", email: "beschaffung.informatik@fsv-whs.de", img: null },
    ],
  },
  {
    id: "kassenwart",
    title: "Kassenwart",
    members: [
      { id: 15, name: "Eric Seidel", email: "kassenwart.informatik@fsv-whs.de", img: null },
    ],
  },
  {
    id: "sport",
    title: "Sport",
    members: [
      { id: 16, name: "Caner Bayram", email: "", img: null },
      { id: 17, name: "Yannic Leismann", email: "", img: null },
      { id: 18, name: "Daniel Chudaska", email: "", img: null },
      { id: 19, name: "Abdullah Msaedi", email: "", img: null },
    ],
  },
];

