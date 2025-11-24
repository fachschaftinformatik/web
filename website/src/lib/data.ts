export type ForumPost = {
  id: string;
  title: string;
  excerpt: string;
  createdAt: string;
  replies: number;
  category: string;
};

export type ForumProgram = "inf-bsc" | "winf-bsc" | "med-bsc" | "inf-msc" | "winf-msc" | "med-msc";

export type ForumComment = {
  id: string;
  author: string;
  text: string;
  createdAt: string;
  parentId?: string | null;
};

export type ForumDemoPost = {
  id: string;
  title: string;
  body: string;
  tags: string[];
  author: string;
  createdAt: string;
  votes: number;
  programs: ForumProgram[];
  comments: ForumComment[];
};

type ForumProgramMeta = {
  id: ForumProgram;
  label: string;
  shortLabel: string;
  level: "Bachelor" | "Master";
};

export const forumProgramCatalog: ForumProgramMeta[] = [
  { id: "inf-bsc", label: "Informatik (B.Sc.)", shortLabel: "INF B.Sc.", level: "Bachelor" },
  { id: "winf-bsc", label: "Wirtschaftsinformatik (B.Sc.)", shortLabel: "WINF B.Sc.", level: "Bachelor" },
  { id: "med-bsc", label: "Medieninformatik (B.Sc.)", shortLabel: "MED B.Sc.", level: "Bachelor" },
  { id: "inf-msc", label: "Informatik (M.Sc.)", shortLabel: "INF M.Sc.", level: "Master" },
  { id: "winf-msc", label: "Wirtschaftsinformatik (M.Sc.)", shortLabel: "WINF M.Sc.", level: "Master" },
  { id: "med-msc", label: "Medieninformatik (M.Sc.)", shortLabel: "MED M.Sc.", level: "Master" },
];

export const forumPrograms: ForumProgram[] = forumProgramCatalog.map((meta) => meta.id);

const forumBaseSeeds: ForumDemoPost[] = [
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
      { id: "c4", author: "Noah", text: "Ich würde einen Form-Provider schreiben.", createdAt: new Date(Date.now() - 28 * 3600 * 1000).toISOString() },
      { id: "c5", author: "Sara", text: "Oder alle Inputs als controlled Components mit Zod Resolver.", createdAt: new Date(Date.now() - 27 * 3600 * 1000).toISOString() },
      { id: "c6", author: "Milan", text: "Ich mag react-hook-form, weil es performant ist.", createdAt: new Date(Date.now() - 26 * 3600 * 1000).toISOString() },
      { id: "c7", author: "Nora", text: "Yup + RHF geht auch schnell.", createdAt: new Date(Date.now() - 25 * 3600 * 1000).toISOString() },
      { id: "c8", author: "Eva", text: "Validierung nur bei Blur spart Render.", createdAt: new Date(Date.now() - 24 * 3600 * 1000).toISOString() },
      { id: "c9", author: "Timo", text: "Achte auf Accessibility, aria-describedby setzen.", createdAt: new Date(Date.now() - 23 * 3600 * 1000).toISOString() },
      { id: "c10", author: "Luca", text: "Inline Errors lieber unter das Feld statt rechts.", createdAt: new Date(Date.now() - 22 * 3600 * 1000).toISOString() },
      { id: "c11", author: "Lea", text: "Ich probiere mal einen FormContext.", createdAt: new Date(Date.now() - 21 * 3600 * 1000).toISOString() },
      { id: "c12", author: "Jonas", text: "Checkout react-aria-components für Inputs.", createdAt: new Date(Date.now() - 20 * 3600 * 1000).toISOString() },
      { id: "c13", author: "Mara", text: "Danke für die Tipps!", createdAt: new Date(Date.now() - 19 * 3600 * 1000).toISOString() },
    ],
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

const makeForumExtraSeeds = (n: number): ForumDemoPost[] => {
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

  const arr: ForumDemoPost[] = [];
  for (let i = 0; i < n; i++) {
    const id = `seed-${i}`;
    const programs: ForumProgram[] =
      i % 7 === 0
        ? forumPrograms.slice()
        : i % 3 === 0
        ? ["inf-bsc", "inf-msc", "winf-bsc"]
        : i % 2 === 0
        ? ["med-bsc", "med-msc"]
        : [forumPrograms[i % forumPrograms.length]];

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

export const forumDemoPosts: ForumDemoPost[] = [...forumBaseSeeds, ...makeForumExtraSeeds(20)];

export const mockForumPosts: ForumPost[] = forumDemoPosts.map((post) => ({
  id: post.id,
  title: post.title,
  excerpt: post.body,
  createdAt: post.createdAt,
  replies: post.comments.length,
  category: post.tags[0] ?? "Allgemein",
}));

// data.ts

export type EventItem = { id: number; title: string; src: string };
export type Bild = { id: number; title: string; thumb: string; full: string };

// --- HILFSFUNKTION FÜR BEISPIELBILDER ---
const pic = (seed: number, w: number, h: number) => `https://picsum.photos/seed/${seed}/${w}/${h}`;

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
};

// --- EVENTS 3-6 (BEISPIELBILDER) ---
// Füllt `bilderByEvent` für alle Events > 2 automatisch mit picsum-Bildern
events.forEach((event) => {
  if (event.id > 2) {
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
