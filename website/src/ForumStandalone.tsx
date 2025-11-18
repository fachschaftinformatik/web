import * as React from "react";
import {
  AppBar, Toolbar, Container, Box, Stack, Paper, Typography, TextField, Button,
  Card, CardContent, IconButton, Chip, Divider, Dialog, DialogTitle, DialogContent,
  DialogActions, InputAdornment, Checkbox, FormGroup, FormControlLabel, Menu, MenuItem,
  Select, Drawer, Grid
} from "@mui/material";
import SearchIcon from "@mui/icons-material/Search";
import AddIcon from "@mui/icons-material/Add";
import ThumbUpOutlined from "@mui/icons-material/ThumbUpOutlined";
import ThumbDownOutlined from "@mui/icons-material/ThumbDownOutlined";
import RestoreIcon from "@mui/icons-material/Restore";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import FilterListIcon from "@mui/icons-material/FilterList";
import LinkIcon from "@mui/icons-material/Link";
import DarkModeOutlinedIcon from "@mui/icons-material/DarkModeOutlined";
import LightModeOutlinedIcon from "@mui/icons-material/LightModeOutlined";

const PROGRAM_CATALOG = [
  { id: "inf-bsc", label: "Informatik (B.Sc.)", shortLabel: "INF B.Sc.", level: "Bachelor" },
  { id: "winf-bsc", label: "Wirtschaftsinformatik (B.Sc.)", shortLabel: "WINF B.Sc.", level: "Bachelor" },
  { id: "med-bsc", label: "Medieninformatik (B.Sc.)", shortLabel: "MED B.Sc.", level: "Bachelor" },
  { id: "inf-msc", label: "Informatik (M.Sc.)", shortLabel: "INF M.Sc.", level: "Master" },
  { id: "winf-msc", label: "Wirtschaftsinformatik (M.Sc.)", shortLabel: "WINF M.Sc.", level: "Master" },
  { id: "med-msc", label: "Medieninformatik (M.Sc.)", shortLabel: "MED M.Sc.", level: "Master" },
] as const;

type ProgramMeta = typeof PROGRAM_CATALOG[number];
type Program = ProgramMeta["id"];
type Vote = -1 | 0 | 1;

type Comment = {
  id: string;
  author: string;
  text: string;
  createdAt: string;
  parentId?: string | null;
};

type Post = {
  id: string;
  title: string;
  body: string;
  tags: string[];
  author: string;
  createdAt: string;
  votes: number;
  programs: Program[];
  comments: Comment[];
};

type CommentAppearance = {
  surface: string;
  border: string;
  accent: string;
  textSecondary: string;
};

const PROGRAM_META_MAP: Record<Program, ProgramMeta> = PROGRAM_CATALOG.reduce((acc, meta) => {
  acc[meta.id] = meta;
  return acc;
}, {} as Record<Program, ProgramMeta>);

const PROGRAMS: Program[] = PROGRAM_CATALOG.map((meta) => meta.id);
const LEGACY_PROGRAM_MAP: Record<string, Program> = {
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

const normalizeProgramValue = (value: any): Program | null => {
  if (!value) return null;
  if (PROGRAM_META_MAP[value as Program]) return value as Program;
  const alias = LEGACY_PROGRAM_MAP[String(value)];
  if (alias) return alias;
  return null;
};

const normalizeProgramList = (list: any[]): Program[] => {
  const normalized = list
    .map((entry) => normalizeProgramValue(entry))
    .filter((entry): entry is Program => Boolean(entry));
  return normalized.length ? Array.from(new Set(normalized)) : [];
};

const createProgramFlagState = (defaults: Program[] = []): Record<Program, boolean> => {
  const flags = {} as Record<Program, boolean>;
  PROGRAMS.forEach((program) => {
    flags[program] = defaults.includes(program);
  });
  return flags;
};
const LS_KEY = "forum-demo-posts";
const LS_VOTES_KEY = "forum-demo-votes";
const currentUser = "Demo User";

const uuid = () =>
  (typeof crypto !== "undefined" && (crypto as any).randomUUID)
    ? (crypto as any).randomUUID()
    : `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;

function isoToShort(iso: string) {
  const d = new Date(iso);
  return d.toLocaleString("de-DE", {
    year: "numeric",
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

type CommentNode = Comment & { children: CommentNode[] };
function buildCommentTree(comments: Comment[]): CommentNode[] {
  const map = new Map<string, CommentNode>();
  const roots: CommentNode[] = [];
  comments.forEach((c) => map.set(c.id, { ...c, children: [] }));
  comments.forEach((c) => {
    const node = map.get(c.id)!;
    const pid = c.parentId ?? null;
    if (!pid) roots.push(node);
    else {
      const parent = map.get(pid);
      if (parent) parent.children.push(node);
      else roots.push(node);
    }
  });
  return roots;
}

function renderTextWithMentions(text: string) {
  const mentionRegex = /@[A-Za-z0-9._-]+/g;
  const nodes: React.ReactNode[] = [];
  let lastIndex = 0;
  let match: RegExpExecArray | null;
  while ((match = mentionRegex.exec(text)) !== null) {
    if (match.index > lastIndex) {
      nodes.push(text.slice(lastIndex, match.index));
    }
    const mention = match[0];
    nodes.push(
      <Box
        component="span"
        key={`${mention}-${match.index}`}
        sx={{
          color: "secondary.main",
          fontWeight: 600,
          cursor: "pointer",
          transition: "color .2s ease",
          "&:hover": { textDecoration: "underline", color: "secondary.dark" },
        }}
      >
        {mention}
      </Box>
    );
    lastIndex = match.index + mention.length;
  }
  if (lastIndex < text.length) {
    nodes.push(text.slice(lastIndex));
  }
  return nodes.length ? nodes : text;
}

/* Seeds */
const BASE_SEEDS: Post[] = [
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
      { id: "c1", author: "Jonas", text: "Ich nutze Zod + eigene Inputs.", createdAt: new Date(Date.now()-32*3600*1000).toISOString() },
      { id: "c2", author: "Mara", text: "React Hook Form ist leichtgewichtig genug.", createdAt: new Date(Date.now()-30*3600*1000).toISOString() },
      { id: "c3", author: "Lea", text: "Danke! Hast du ein Beispielrepo?", createdAt: new Date(Date.now()-29*3600*1000).toISOString(), parentId: "c1" },
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
    title: "useMemo/useCallback – Overhead vs. Nutzen?",
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

function makeExtraSeeds(n: number): Post[] {
  const authors = ["Lea","Jonas","Mara","Timo","Eva","Noah","Sara","Luca","Milan","Nora"];
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
  const tagsPool = ["react","typescript","mui","routing","state","hooks","performance","testing","vite","ui"];

  const arr: Post[] = [];
  for (let i = 0; i < n; i++) {
    const id = `seed-${i}`;
    const programs: Program[] =
      i % 7 === 0 ? PROGRAMS.slice()
      : i % 3 === 0 ? ["inf-bsc", "inf-msc", "winf-bsc"]
      : i % 2 === 0 ? ["med-bsc", "med-msc"]
      : [PROGRAMS[i % PROGRAMS.length]];
    let comments: Comment[] = [];
    if (i % 5 === 0) {
      comments = [
        {
          id: `${id}-c1`,
          author: "Eva",
          text: "Klingt spannend – hast du ein Repo?",
          createdAt: new Date(Date.now() - (i + 2) * 3600 * 1000).toISOString(),
        },
        {
          id: `${id}-c2`,
          author: "Timo",
          text: "Ich habe letzte Woche etwas Ähnliches gebaut.",
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

    const base: Post = {
      id,
      title: `Demo #${i + 1}: ${topics[i % topics.length]}`,
      body: "Dies ist ein Demo-Beitrag zum Testen von Suche, Sortierung, Filter, Votes, Erstellen und verschachtelten Kommentaren.",
      tags: [tagsPool[i % tagsPool.length], tagsPool[(i + 3) % tagsPool.length]],
      author: authors[i % authors.length],
      createdAt: new Date(Date.now() - (i + 4) * 2 * 3600 * 1000).toISOString(),
      votes: (i * 7) % 25,
      programs,
      comments,
    };
    arr.push(base);
  }
  return arr;
}

const SEED_POSTS: Post[] = [...BASE_SEEDS, ...makeExtraSeeds(20)];

/* Kommentare (rekursiv) */
function CommentThread({
  node,
  onReply,
  depth = 0,
  appearance,
}: {
  node: CommentNode;
  onReply: (parentId: string, text: string) => void;
  depth?: number;
  appearance: CommentAppearance;
}) {
  const [replyOpen, setReplyOpen] = React.useState(false);
  const [text, setText] = React.useState("");
  const replyInputBg = appearance.surface === "#111a2a" ? "rgba(255,255,255,0.08)" : "#fff";

  return (
    <Box
      sx={{
        mt: 1.25,
        pl: depth ? 3 : 0,
        position: "relative",
        "&::before": depth
          ? {
              content: '""',
              position: "absolute",
              left: 8,
              top: 0,
              bottom: 0,
              width: 2,
              bgcolor: appearance.border,
              borderRadius: 1,
            }
          : {},
      }}
    >
      <Paper
        variant="outlined"
        sx={{
          p: 1.5,
          borderRadius: 2,
          bgcolor: appearance.surface,
          borderColor: appearance.border,
        }}
      >
        <Stack spacing={0.75}>
          <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
            <Typography variant="subtitle2">{node.author}</Typography>
            <Typography component="span" variant="caption" sx={{ color: appearance.textSecondary }}>
              · {isoToShort(node.createdAt)}
            </Typography>
          </Stack>
          <Typography variant="body2">{renderTextWithMentions(node.text)}</Typography>

          <Stack direction="row" spacing={1}>
            <Button size="small" variant="text" onClick={() => setReplyOpen((v) => !v)}>
              Antworten
            </Button>
          </Stack>

          {replyOpen && (
            <Stack direction={{ xs: "column", sm: "row" }} spacing={1} alignItems="flex-start">
              <TextField
                size="small"
                fullWidth
                placeholder="Antwort schreiben…"
                value={text}
                onChange={(e) => setText(e.target.value)}
                sx={{ "& .MuiInputBase-root": { bgcolor: replyInputBg } }}
              />
              <Button
                variant="contained"
                onClick={() => {
                  if (text.trim()) {
                    onReply(node.id, text.trim());
                    setText("");
                    setReplyOpen(false);
                  }
                }}
              >
                Senden
              </Button>
            </Stack>
          )}
        </Stack>
      </Paper>

      {node.children.map((child) => (
        <CommentThread
          key={child.id}
          node={child}
          onReply={onReply}
          depth={depth + 1}
          appearance={appearance}
        />
      ))}
    </Box>
  );
}
function CommentsSection({
  comments,
  onAdd,
  appearance,
}: {
  comments: Comment[];
  onAdd: (parentId: string | null, text: string) => void;
  appearance: CommentAppearance;
}) {
  const tree = React.useMemo(() => buildCommentTree(comments), [comments]);
  const [text, setText] = React.useState("");
  const [expanded, setExpanded] = React.useState(false);
  const [fullViewOpen, setFullViewOpen] = React.useState(false);
  const shouldClamp = comments.length > 3;
  const inputBg = appearance.surface === "#111a2a" ? "rgba(255,255,255,0.08)" : "#fff";
  const renderNodes = React.useCallback(
    () =>
      tree.map((root) => (
        <CommentThread key={root.id} node={root} onReply={(pid, t) => onAdd(pid, t)} appearance={appearance} />
      )),
    [tree, onAdd, appearance]
  );

  return (
    <Box>
      <Box
        sx={{
          border: 1,
          borderColor: appearance.border,
          borderRadius: 2,
          backgroundColor: appearance.surface,
          p: 1.5,
          ...(shouldClamp && {
            maxHeight: expanded ? 420 : 220,
            overflowY: "auto",
            pr: 2,
          }),
        }}
      >
        {tree.length ? (
          <Stack spacing={1}>{renderNodes()}</Stack>
        ) : (
          <Typography variant="body2" sx={{ color: appearance.textSecondary }}>
            Noch keine Kommentare.
          </Typography>
        )}
      </Box>

      {shouldClamp && (
        <Stack direction={{ xs: "column", sm: "row" }} spacing={1} mt={1}>
          <Button size="small" variant="text" onClick={() => setExpanded((v) => !v)}>
            {expanded ? "Weniger anzeigen" : "Mehr anzeigen"}
          </Button>
          <Button size="small" variant="text" onClick={() => setFullViewOpen(true)}>
            Im Großformat öffnen
          </Button>
        </Stack>
      )}

      <Box sx={{ mt: 1.5 }}>
        <Stack direction={{ xs: "column", sm: "row" }} spacing={1} alignItems="flex-start">
          <TextField
            size="small"
            fullWidth
            placeholder="Neuen Kommentar schreiben…"
            value={text}
            onChange={(e) => setText(e.target.value)}
            sx={{ "& .MuiInputBase-root": { bgcolor: inputBg } }}
          />
          <Button
            variant="contained"
            onClick={() => {
              if (text.trim()) {
                onAdd(null, text.trim());
                setText("");
              }
            }}
          >
            Posten
          </Button>
        </Stack>
      </Box>

      <Dialog open={fullViewOpen} onClose={() => setFullViewOpen(false)} fullWidth maxWidth="md">
        <DialogTitle>Kommentare</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={1}>{renderNodes()}</Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setFullViewOpen(false)}>Schließen</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
function PostItem({
  post,
  vote,
  onVote,
  onAddComment,
  onDelete,
  onReport,
  onOpenDetail,
  commentAppearance,
  mutedColor,
}: {
  post: Post;
  vote: Vote;
  onVote: (id: string, v: Vote) => void;
  onAddComment: (postId: string, parentId: string | null, text: string) => void;
  onDelete: (id: string) => void;
  onReport: (id: string) => void;
  onOpenDetail: (post: Post) => void;
  commentAppearance: CommentAppearance;
  mutedColor: string;
}) {
  const netVotes = post.votes + (vote as number);
  const [menuEl, setMenuEl] = React.useState<null | HTMLElement>(null);
  const openMenu = Boolean(menuEl);
  const handleOpenDetail = () => onOpenDetail(post);

  return (
    <Card
      component="article"
      variant="outlined"
      sx={{
        borderRadius: 3,
        transition: "transform .2s ease, box-shadow .2s ease",
        bgcolor: "var(--card-bg)",
        borderColor: "var(--card-border)",
        "&:hover": { transform: "translateY(-2px)", boxShadow: 6 },
      }}
    >
      <CardContent sx={{ p: { xs: 2.5, md: 3 } }}>
        <Stack direction={{ xs: "column", md: "row" }} spacing={{ xs: 2, md: 3 }} alignItems="stretch">
          <Stack
            direction={{ xs: "row", md: "column" }}
            spacing={1}
            alignItems="center"
            justifyContent="center"
            sx={{
              minWidth: { xs: "100%", md: 72 },
              borderRadius: 2,
              bgcolor: "action.hover",
              p: 1.5,
            }}
          >
            <IconButton
              aria-label="upvote"
              onClick={() => onVote(post.id, vote === 1 ? 0 : 1)}
              color={vote === 1 ? "primary" : "default"}
              size="small"
            >
              <ThumbUpOutlined fontSize="small" />
            </IconButton>
            <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
              {netVotes}
            </Typography>
            <IconButton
              aria-label="downvote"
              onClick={() => onVote(post.id, vote === -1 ? 0 : -1)}
              color={vote === -1 ? "primary" : "default"}
              size="small"
            >
              <ThumbDownOutlined fontSize="small" />
            </IconButton>
          </Stack>

          <Stack spacing={1.5} sx={{ flex: 1 }}>
            <Stack direction="row" spacing={1} alignItems="flex-start">
              <Box sx={{ flex: 1, cursor: "pointer" }} onClick={handleOpenDetail}>
                <Typography variant="h6" sx={{ mb: 0.75, lineHeight: 1.2 }}>
                  {post.title}
                </Typography>
                <Typography variant="body2" sx={{ color: mutedColor }}>
                  {post.body}
                </Typography>
              </Box>
              <IconButton
                size="small"
                aria-label="Aktionen"
                onClick={(e) => {
                  e.stopPropagation();
                  setMenuEl(e.currentTarget);
                }}
              >
                <MoreVertIcon />
              </IconButton>
            </Stack>

            <Stack
              direction="row"
              spacing={1}
              alignItems="center"
              flexWrap="wrap"
              sx={{ cursor: "pointer" }}
              onClick={handleOpenDetail}
            >
              {post.tags.map((t) => (
                <Chip key={t} label={t} size="small" variant="outlined" />
              ))}
              <Divider flexItem orientation="vertical" sx={{ mx: 1 }} />
              {post.programs.map((p) => (
                <Chip
                  key={p}
                  label={PROGRAM_META_MAP[p]?.shortLabel ?? p}
                  size="small"
                  variant="outlined"
                />
              ))}
              <Divider flexItem orientation="vertical" sx={{ mx: 1 }} />
              <Typography variant="caption" sx={{ color: mutedColor }}>
                von {post.author} · {isoToShort(post.createdAt)} · {post.comments.length} Kommentare
              </Typography>
            </Stack>

            {post.comments.length > 0 && <Divider sx={{ my: 1 }} />}

            <CommentsSection
              comments={post.comments}
              onAdd={(parentId, text) => onAddComment(post.id, parentId, text)}
              appearance={commentAppearance}
            />
          </Stack>
        </Stack>
      </CardContent>

      <Menu anchorEl={menuEl} open={openMenu} onClose={() => setMenuEl(null)}>
        <MenuItem
          onClick={() => {
            setMenuEl(null);
            navigator.clipboard?.writeText(location.href + "#" + post.id);
          }}
        >
          Link kopieren
        </MenuItem>
        <MenuItem
          onClick={() => {
            setMenuEl(null);
            onReport(post.id);
          }}
        >
          Beitrag melden
        </MenuItem>
        {post.author === currentUser && (
          <MenuItem
            onClick={() => {
              setMenuEl(null);
              onDelete(post.id);
            }}
          >
            Beitrag löschen
          </MenuItem>
        )}
      </Menu>
    </Card>
  );
}
/* Hauptkomponente */
export default function ForumStandalone() {
  // Laden + Migration (program ? programs)
  const [posts, setPosts] = React.useState<Post[]>(() => {
    try {
      const raw = typeof window !== "undefined" ? localStorage.getItem(LS_KEY) : null;
      const data = (raw ? JSON.parse(raw) : SEED_POSTS) as any[];
      return data.map((p) => {
        const { program, programs: storedPrograms, ...rest } = p ?? {};
        const normalizedPrograms = Array.isArray(storedPrograms)
          ? normalizeProgramList(storedPrograms)
          : program
          ? normalizeProgramList([program])
          : [];
        const programs = normalizedPrograms.length ? normalizedPrograms : [PROGRAMS[0]];
        return { ...(rest as Omit<Post, "programs">), programs } as Post;
      });
    } catch {
      return SEED_POSTS;
    }
  });

  const [votes, setVotes] = React.useState<Record<string, Vote>>(() => {
    try {
      const raw = typeof window !== "undefined" ? localStorage.getItem(LS_VOTES_KEY) : null;
      return raw ? (JSON.parse(raw) as Record<string, Vote>) : {};
    } catch {
      return {};
    }
  });

  const [darkMode, setDarkMode] = React.useState(false);
  const palette = React.useMemo(
    () =>
      darkMode
        ? {
            background: "#05090f",
            surface: "#0f1422",
            card: "#161c2d",
            border: "rgba(255,255,255,0.12)",
            textSecondary: "rgba(255,255,255,0.6)",
          }
        : {
            background: "#f5f7fb",
            surface: "#ffffff",
            card: "#ffffff",
            border: "#e0e3eb",
            textSecondary: "rgba(0,0,0,0.6)",
          },
    [darkMode]
  );
  const commentAppearance = React.useMemo<CommentAppearance>(
    () => ({
      surface: darkMode ? "#111a2a" : "#f8fafc",
      border: darkMode ? "rgba(255,255,255,0.15)" : "#e0e3eb",
      accent: "#0f6e2e",
      textSecondary: darkMode ? "rgba(255,255,255,0.65)" : "rgba(0,0,0,0.6)",
    }),
    [darkMode]
  );
  const inputStyles = React.useMemo(
    () => ({
      "& .MuiInputBase-root": {
        bgcolor: darkMode ? "rgba(255,255,255,0.08)" : "#fff",
      },
    }),
    [darkMode]
  );

  const [detailPost, setDetailPost] = React.useState<Post | null>(null);
  const [shareCopied, setShareCopied] = React.useState(false);

  const shareUrl = React.useMemo(() => {
    if (!detailPost) return "";
    const origin =
      typeof window !== "undefined" && window.location?.origin
        ? window.location.origin
        : "https://campus-demo.local";
    return `${origin}/post/${detailPost.id}`;
  }, [detailPost]);

  React.useEffect(() => {
    if (!shareCopied) return;
    let timer: number | null = null;
    if (typeof window !== "undefined") {
      timer = window.setTimeout(() => setShareCopied(false), 2000);
    }
    return () => {
      if (timer !== null && typeof window !== "undefined") {
        window.clearTimeout(timer);
      }
    };
  }, [shareCopied]);

  React.useEffect(() => { localStorage.setItem(LS_KEY, JSON.stringify(posts)); }, [posts]);
  React.useEffect(() => { localStorage.setItem(LS_VOTES_KEY, JSON.stringify(votes)); }, [votes]);
  React.useEffect(() => {
    if (!detailPost) return;
    const updated = posts.find((p) => p.id === detailPost.id);
    if (updated && updated !== detailPost) {
      setDetailPost(updated);
    }
  }, [posts, detailPost]);

  const openDetail = (post: Post) => {
    setDetailPost(post);
    setShareCopied(false);
  };

  const closeDetail = () => {
    setDetailPost(null);
    setShareCopied(false);
  };

  const copyShareUrl = () => {
    if (!shareUrl) return;
    if (typeof navigator !== "undefined" && navigator.clipboard?.writeText) {
      navigator.clipboard
        .writeText(shareUrl)
        .then(() => setShareCopied(true))
        .catch(() => setShareCopied(true));
    } else {
      setShareCopied(true);
    }
  };

  // Suche + Sortierung
  const [q, setQ] = React.useState("");
  const [sort, setSort] = React.useState<"new" | "votes">("new");
  const [filtersOpen, setFiltersOpen] = React.useState(false);
  const [activeProgramFilters, setActiveProgramFilters] = React.useState<Program[]>([]);
  const [allProgramsOnly, setAllProgramsOnly] = React.useState(false);
  const toggleProgramFilter = (program: Program) => {
    setActiveProgramFilters((prev) =>
      prev.includes(program) ? prev.filter((entry) => entry !== program) : [...prev, program]
    );
  };
  const clearAllFilters = () => {
    setActiveProgramFilters([]);
    setAllProgramsOnly(false);
    setQ("");
    setSort("new");
  };
  const activeFiltersCount = activeProgramFilters.length + (allProgramsOnly ? 1 : 0);

  const hasAllPrograms = React.useCallback(
    (p: Post) => PROGRAMS.every((program) => p.programs.includes(program)),
    []
  );

  const filtered = React.useMemo(() => {
    const query = q.trim().toLowerCase();
    let base = posts;

    if (allProgramsOnly) {
      base = base.filter(hasAllPrograms);
    }

    if (activeProgramFilters.length) {
      base = base.filter((p) =>
        p.programs.some((program) => activeProgramFilters.includes(program))
      );
    }

    if (query) {
      base = base.filter(
        (p) =>
          p.title.toLowerCase().includes(query) ||
          p.body.toLowerCase().includes(query) ||
          p.tags.some((t) => t.toLowerCase().includes(query))
      );
    }

    if (sort === "votes") {
      return [...base].sort(
        (a, b) => (b.votes + (votes[b.id] || 0)) - (a.votes + (votes[a.id] || 0))
      );
    }
    return [...base].sort((a, b) => +new Date(b.createdAt) - +new Date(a.createdAt));
  }, [posts, q, sort, votes, hasAllPrograms, activeProgramFilters, allProgramsOnly]);

  const handleVote = (id: string, v: Vote) => setVotes((prev) => ({ ...prev, [id]: v }));

  // Kommentare
  const handleAddComment = (postId: string, parentId: string | null, text: string) => {
    const newComment: Comment = {
      id: uuid(),
      author: currentUser,
      text,
      createdAt: new Date().toISOString(),
      parentId: parentId ?? null,
    };
    setPosts((prev) =>
      prev.map((p) => (p.id === postId ? { ...p, comments: [...p.comments, newComment] } : p))
    );
  };

  // Create Dialog
  const [open, setOpen] = React.useState(false);
  const [title, setTitle] = React.useState("");
  const [body, setBody] = React.useState("");
  const [tagsInput, setTagsInput] = React.useState("");

  // "Für alle Studiengänge" + einzelne Checkboxen
  const [allPrograms, setAllPrograms] = React.useState<boolean>(false);
  const [programsInput, setProgramsInput] = React.useState<Record<Program, boolean>>(
    () => createProgramFlagState([PROGRAMS[0]])
  );
  const selectedPrograms = React.useMemo<Program[]>(
    () => (allPrograms ? PROGRAMS.slice() : PROGRAMS.filter((p) => programsInput[p])),
    [allPrograms, programsInput]
  );
  const noProgramSelected = selectedPrograms.length === 0;

  const createPost = () => {
    if (!title.trim() || noProgramSelected) return;
    const newPost: Post = {
      id: uuid(),
      title: title.trim(),
      body: body.trim(),
      tags: tagsInput.split(",").map((t) => t.trim()).filter(Boolean),
      author: currentUser,
      createdAt: new Date().toISOString(),
      votes: 0,
      programs: selectedPrograms,
      comments: [],
    };
    setPosts((prev) => [newPost, ...prev]);
    setOpen(false);
    setTitle(""); setBody(""); setTagsInput("");
    setAllPrograms(false);
    setProgramsInput(createProgramFlagState([PROGRAMS[0]]));
  };

  // Menü-Aktionen
  const handleDelete = (id: string) => {
    setPosts(prev => prev.filter(p => p.id !== id));
  };

  // Report Dialog
  const [reportFor, setReportFor] = React.useState<string | null>(null);
  const [reportReason, setReportReason] = React.useState("Spam / Werbung");
  const openReport = (id: string) => setReportFor(id);
  const sendReport = () => {
    // hier würdest du an ein Backend schicken – wir loggen nur:
    console.log("Report:", { postId: reportFor, reason: reportReason });
    setReportFor(null);
  };

  const resetDemo = () => {
    setPosts(SEED_POSTS);
    setVotes({});
    clearAllFilters();
    setFiltersOpen(false);
    closeDetail();
  };

  /* UI */
  return (
    <Box
      sx={{
        minHeight: "100vh",
        bgcolor: palette.background,
        color: darkMode ? "common.white" : "text.primary",
        "--card-bg": palette.card,
        "--card-border": palette.border,
      }}
    >
      <AppBar
        position="sticky"
        elevation={0}
        sx={{ backgroundColor: "#0f6e2e", color: "common.white", borderBottom: "1px solid rgba(255,255,255,0.2)" }}
      >
        <Toolbar sx={{ minHeight: 72 }}>
          <Container maxWidth="lg" sx={{ display: "flex", alignItems: "center", gap: 2 }}>
            <Typography variant="h6" sx={{ flexGrow: 1, fontWeight: 600 }}>
              Campus Forum
            </Typography>
            <Button
              startIcon={darkMode ? <LightModeOutlinedIcon /> : <DarkModeOutlinedIcon />}
              variant="outlined"
              color="inherit"
              onClick={() => setDarkMode((prev) => !prev)}
            >
              {darkMode ? "Light Mode" : "Dark Mode"}
            </Button>
            <Button startIcon={<AddIcon />} variant="contained" onClick={() => setOpen(true)}>
              Beitrag erstellen
            </Button>
            <Button startIcon={<RestoreIcon />} variant="outlined" onClick={resetDemo}>
              Demo zurücksetzen
            </Button>
          </Container>
        </Toolbar>
      </AppBar>

      <Container maxWidth="lg" sx={{ py: 3, color: darkMode ? "common.white" : "text.primary" }}>
        {/* Filterleiste */}
        <Paper
          elevation={0}
          sx={{
            p: 3,
            mb: 3,
            borderRadius: 4,
            border: "1px solid",
            borderColor: palette.border,
            bgcolor: palette.surface,
            boxShadow: darkMode ? "0px 8px 24px rgba(0,0,0,0.35)" : "0px 8px 24px rgba(15,110,46,0.08)",
            color: darkMode ? "common.white" : "text.primary",
          }}
        >
          <Stack spacing={2}>
            <Stack direction={{ xs: "column", md: "row" }} spacing={2} alignItems={{ xs: "stretch", md: "center" }}>
              <TextField
                placeholder="Suche in Titel, Text oder Tags…"
                value={q}
                onChange={(e) => setQ(e.target.value)}
                fullWidth
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <SearchIcon fontSize="small" />
                    </InputAdornment>
                  ),
                }}
                sx={inputStyles}
              />
              <Select
                size="small"
                value={sort}
                onChange={(e) => setSort(e.target.value as "new" | "votes")}
                sx={{
                  width: { xs: "100%", md: 220 },
                  "& .MuiOutlinedInput-root": {
                    bgcolor: darkMode ? "rgba(255,255,255,0.08)" : "#fff",
                  },
                }}
              >
                <MenuItem value="new">Neueste zuerst</MenuItem>
                <MenuItem value="votes">Beste (Votes)</MenuItem>
              </Select>
              <Button
                startIcon={<FilterListIcon />}
                variant={activeFiltersCount ? "contained" : "outlined"}
                color="secondary"
                onClick={() => setFiltersOpen(true)}
                size="small"
                sx={{
                  width: { xs: "100%", md: 220 },
                  alignSelf: { xs: "stretch", md: "center" },
                  height: 40,
                }}
              >
                Filter{activeFiltersCount ? ` (${activeFiltersCount})` : ""}
              </Button>
            </Stack>

            {(activeProgramFilters.length > 0 || allProgramsOnly) && (
              <Stack direction="row" spacing={1} flexWrap="wrap" alignItems="center">
                {allProgramsOnly && (
                  <Chip
                    label="Nur Beiträge für alle Studiengänge"
                    onDelete={() => setAllProgramsOnly(false)}
                    color="secondary"
                    variant="outlined"
                  />
                )}
                {activeProgramFilters.map((program) => (
                  <Chip
                    key={program}
                    label={PROGRAM_META_MAP[program]?.label ?? program}
                    onDelete={() => toggleProgramFilter(program)}
                    color="secondary"
                    variant="outlined"
                  />
                ))}
                <Button size="small" onClick={clearAllFilters}>
                  Alle Filter entfernen
                </Button>
              </Stack>
            )}
          </Stack>
        </Paper>

        {/* Liste */}
        <Stack spacing={2}>
          {filtered.map((p) => (
            <PostItem
              key={p.id}
              post={p}
              vote={(votes[p.id] ?? 0) as Vote}
              onVote={handleVote}
              onAddComment={handleAddComment}
              onDelete={handleDelete}
              onReport={openReport}
              onOpenDetail={openDetail}
              commentAppearance={commentAppearance}
              mutedColor={palette.textSecondary}
            />
          ))}
          {filtered.length === 0 && (
            <Typography variant="body2" sx={{ color: palette.textSecondary }}>
              Keine Treffer. Suchbegriff/Filter anpassen oder Demo zurücksetzen.
            </Typography>
          )}
        </Stack>
      </Container>

      <Drawer
        anchor="right"
        open={filtersOpen}
        onClose={() => setFiltersOpen(false)}
        PaperProps={{ sx: { bgcolor: palette.surface, color: darkMode ? "common.white" : "text.primary" } }}
      >
        <Box sx={{ width: { xs: 340, sm: 420 }, p: 3 }}>
          <Stack spacing={2}>
            <Stack direction="row" justifyContent="space-between" alignItems="center">
              <Typography variant="h6">Filter</Typography>
              <Button size="small" onClick={clearAllFilters}>
                Zurücksetzen
              </Button>
            </Stack>

            <Box>
              <Typography variant="subtitle2" sx={{ color: palette.textSecondary }}>
                Beiträge
              </Typography>
              <FormControlLabel
                sx={{ mt: 1 }}
                control={
                  <Checkbox
                    checked={allProgramsOnly}
                    onChange={(e) => setAllProgramsOnly(e.target.checked)}
                  />
                }
                label="Nur Beiträge anzeigen, die alle Studiengänge adressieren"
              />
            </Box>

            <Box>
              <Typography variant="subtitle2" sx={{ color: palette.textSecondary }}>
                Studiengänge
              </Typography>
              <Grid container spacing={2} mt={1}>
                {PROGRAM_CATALOG.map((meta) => {
                  const active = activeProgramFilters.includes(meta.id);
                  return (
                    // @ts-ignore legacy Grid props retained for old layout
                    <Grid item component="div" xs={6} key={meta.id}>
                      <Paper
                        role="button"
                        tabIndex={0}
                        onClick={() => toggleProgramFilter(meta.id)}
                        onKeyDown={(e) => {
                          if (e.key === "Enter" || e.key === " ") {
                            e.preventDefault();
                            toggleProgramFilter(meta.id);
                          }
                        }}
                        sx={{
                          p: 1.5,
                          border: 2,
                          borderColor: active ? "secondary.main" : palette.border,
                          bgcolor: active ? "rgba(15,110,46,0.12)" : palette.surface,
                          cursor: "pointer",
                          transition: "all .2s ease",
                          minHeight: 96,
                          display: "flex",
                          flexDirection: "column",
                          justifyContent: "space-between",
                          "&:hover": {
                            borderColor: "secondary.main",
                            boxShadow: 2,
                          },
                        }}
                      >
                        <Typography variant="body2" fontWeight={600}>
                          {meta.label}
                        </Typography>
                        <Typography variant="caption" color="text.secondary">
                          {meta.level}
                        </Typography>
                      </Paper>
                    </Grid>
                  );
                })}
              </Grid>
            </Box>

            <Stack direction="row" justifyContent="space-between" pt={1}>
              <Button onClick={clearAllFilters}>Alle löschen</Button>
              <Button variant="contained" onClick={() => setFiltersOpen(false)}>
                Anwenden
              </Button>
            </Stack>
          </Stack>
        </Box>
      </Drawer>

      <Dialog open={!!detailPost} onClose={closeDetail} fullWidth maxWidth="md">
        {detailPost && (
          <>
            <DialogTitle>{detailPost.title}</DialogTitle>
            <DialogContent dividers>
              <Stack spacing={2}>
                <Typography variant="body1">{detailPost.body}</Typography>

                {detailPost.tags.length > 0 && (
                  <Stack direction="row" spacing={1} flexWrap="wrap">
                    {detailPost.tags.map((tag) => (
                      <Chip key={tag} label={tag} size="small" variant="outlined" />
                    ))}
                  </Stack>
                )}

                <Stack direction="row" spacing={1} flexWrap="wrap">
                  {detailPost.programs.map((program) => (
                    <Chip
                      key={program}
                      label={PROGRAM_META_MAP[program]?.label ?? program}
                      size="small"
                      color="secondary"
                      variant="outlined"
                    />
                  ))}
                </Stack>

                <Typography variant="body2" sx={{ color: palette.textSecondary }}>
                  von {detailPost.author} · {isoToShort(detailPost.createdAt)} ·{" "}
                  {detailPost.votes + (votes[detailPost.id] || 0)} Votes
                </Typography>

                <Divider />

                <Typography variant="h6">Diskussion</Typography>
                <CommentsSection
                  comments={detailPost.comments}
                  onAdd={(parentId, text) => handleAddComment(detailPost.id, parentId, text)}
                  appearance={commentAppearance}
                />

                <Divider />

                <Stack direction={{ xs: "column", md: "row" }} spacing={1} alignItems="stretch">
                  <TextField
                    label="Link teilen"
                    value={shareUrl}
                    fullWidth
                    InputProps={{ readOnly: true }}
                    sx={inputStyles}
                  />
                  <Button
                    variant="contained"
                    startIcon={<LinkIcon />}
                    onClick={copyShareUrl}
                    sx={{ whiteSpace: "nowrap" }}
                  >
                    {shareCopied ? "Kopiert!" : "Link kopieren"}
                  </Button>
                </Stack>
                {shareCopied && (
                  <Typography variant="caption" color="success.main">
                    Link in der Zwischenablage gespeichert.
                  </Typography>
                )}
              </Stack>
            </DialogContent>
            <DialogActions>
              <Button onClick={closeDetail}>Schließen</Button>
            </DialogActions>
          </>
        )}
      </Dialog>

      {/* Create Dialog */}
      <Dialog open={open} onClose={() => setOpen(false)} fullWidth maxWidth="sm">
        <DialogTitle>Beitrag erstellen</DialogTitle>
        <DialogContent dividers>
          <Stack spacing={2} mt={1}>
            <TextField
              label="Titel"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              autoFocus
              sx={inputStyles}
            />
            <TextField
              label="Inhalt"
              value={body}
              onChange={(e) => setBody(e.target.value)}
              multiline
              minRows={5}
              sx={inputStyles}
            />
            <TextField
              label="Tags (kommagetrennt)"
              placeholder="react, typescript, mui"
              value={tagsInput}
              onChange={(e) => setTagsInput(e.target.value)}
              sx={inputStyles}
            />

            {/* Für alle + einzelne Checkboxen in einer Reihe */}
            <FormGroup row>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={allPrograms}
                    onChange={(e) => setAllPrograms(e.target.checked)}
                  />
                }
                label="Für alle Studiengänge"
                sx={{ mr: 2 }}
              />

              {PROGRAMS.map((p) => (
                <FormControlLabel
                  key={p}
                  control={
                    <Checkbox
                      checked={allPrograms ? true : programsInput[p]}
                      disabled={allPrograms}
                      onChange={(e) =>
                        setProgramsInput((prev) => ({ ...prev, [p]: e.target.checked }))
                      }
                    />
                  }
                  label={PROGRAM_META_MAP[p].label}
                />
              ))}
            </FormGroup>

            {noProgramSelected && (
              <Typography variant="caption" color="error" sx={{ mt: 0.5 }}>
                Bitte mindestens einen Studiengang auswählen (oder "Für alle Studiengänge" aktivieren).
              </Typography>
            )}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Abbrechen</Button>
          <Button variant="contained" onClick={createPost} disabled={noProgramSelected || !title.trim()}>
            Speichern
          </Button>
        </DialogActions>
      </Dialog>

      {/* Report Dialog */}
      <Dialog open={!!reportFor} onClose={() => setReportFor(null)} maxWidth="xs" fullWidth>
        <DialogTitle>Beitrag melden</DialogTitle>
        <DialogContent dividers>
          <Typography variant="body2" sx={{ mb: 1 }}>
            Wähle einen Grund:
          </Typography>
          <Select
            fullWidth
            size="small"
            value={reportReason}
            onChange={(e) => setReportReason(e.target.value as string)}
          >
            <MenuItem value="Spam / Werbung">Spam / Werbung</MenuItem>
            <MenuItem value="Beleidigung / Hate">Beleidigung / Hate</MenuItem>
            <MenuItem value="Falsche Kategorie">Falsche Kategorie</MenuItem>
            <MenuItem value="Urheberrechtsverletzung">Urheberrechtsverletzung</MenuItem>
            <MenuItem value="Sonstiges">Sonstiges</MenuItem>
          </Select>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setReportFor(null)}>Abbrechen</Button>
          <Button variant="contained" onClick={sendReport}>Melden</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}




