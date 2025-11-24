import * as React from "react";
import {
  Container, Box, Stack, Paper, Typography, TextField, Button,
  Card, CardContent, IconButton, Chip, Divider, Dialog, DialogTitle, DialogContent,
  DialogActions, InputAdornment, Checkbox, FormGroup, FormControlLabel, Menu, MenuItem,
  Select, Drawer, Grid, Tooltip, SxProps, Pagination
} from "@mui/material";
import { alpha, useTheme } from "@mui/material/styles";
import SearchIcon from "@mui/icons-material/Search";
import AddIcon from "@mui/icons-material/Add";
import ThumbUpOutlined from "@mui/icons-material/ThumbUpOutlined";
import ThumbDownOutlined from "@mui/icons-material/ThumbDownOutlined";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import FilterListIcon from "@mui/icons-material/FilterList";
import LinkIcon from "@mui/icons-material/Link";
import { Sidebar } from "@components/layout";
import { useAuth } from "@lib/auth";
import { forumDemoPosts } from "@lib/data";

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

const normalizeProgramValue = (value: unknown): Program | null => {
  if (!value) return null;
  if (PROGRAM_META_MAP[value as Program]) return value as Program;
  const alias = LEGACY_PROGRAM_MAP[String(value)];
  if (alias) return alias;
  return null;
};

const normalizeProgramList = (list: unknown[]): Program[] => {
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

const uuid = () => {
  const cryptoObj = typeof crypto !== "undefined" ? crypto : undefined;
  if (cryptoObj && "randomUUID" in cryptoObj && typeof cryptoObj.randomUUID === "function") {
    return cryptoObj.randomUUID();
  }
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
};

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

const COMMENT_COLLAPSE_LIMIT = 6;
const POSTS_PER_PAGE = 20;

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
          color: "primary.light",
          fontWeight: 600,
          cursor: "pointer",
          transition: "color .2s ease",
          "&:hover": { textDecoration: "underline", color: "primary.main" },
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
const SEED_POSTS: Post[] = forumDemoPosts;

/* Kommentare (rekursiv) */
function CommentThread({
  node,
  onReply,
  depth = 0,
  appearance,
  canReply,
}: {
  node: CommentNode;
  onReply: (parentId: string, text: string) => void;
  depth?: number;
  appearance: CommentAppearance;
  canReply: boolean;
}) {
  const [replyOpen, setReplyOpen] = React.useState(false);
  const [text, setText] = React.useState("");
  const replyInputBg = appearance.surface === "#111a2a" ? "rgba(255,255,255,0.08)" : "#fff";

  return (
    <Box
      sx={[{
        mt: 1.25,
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
          : {}
      }, depth ? {
        pl: 3
      } : {
        pl: 0
      }]}
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
            <Button
              size="small"
              variant="text"
              onClick={() => setReplyOpen((v) => !v)}
              disabled={!canReply}
            >
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
                disabled={!canReply}
                sx={{ "& .MuiInputBase-root": { bgcolor: replyInputBg } }}
              />
              <Button
                variant="contained"
                onClick={() => {
                  if (!canReply) return;
                  if (text.trim()) {
                    onReply(node.id, text.trim());
                    setText("");
                    setReplyOpen(false);
                  }
                }}
                disabled={!canReply}
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
          canReply={canReply}
        />
      ))}
    </Box>
  );
}
function CommentsSection({
  comments,
  onAdd,
  appearance,
  canComment,
  disabledHelper,
}: {
  comments: Comment[];
  onAdd: (parentId: string | null, text: string) => void;
  appearance: CommentAppearance;
  canComment: boolean;
  disabledHelper?: string;
}) {
  const totalComments = comments.length;
  const [expanded, setExpanded] = React.useState(false);
  const visibleComments = React.useMemo(
    () =>
      expanded || totalComments <= COMMENT_COLLAPSE_LIMIT
        ? comments
        : comments.slice(0, COMMENT_COLLAPSE_LIMIT),
    [comments, expanded, totalComments]
  );
  const hiddenCount = totalComments - visibleComments.length;
  const tree = React.useMemo(() => buildCommentTree(visibleComments), [visibleComments]);
  const [text, setText] = React.useState("");
  const shouldScroll = visibleComments.length > 5;
  const inputBg = appearance.surface;
  const handleAdd = React.useCallback(
    (parentId: string | null, t: string) => {
      if (!expanded && (hiddenCount > 0 || totalComments >= COMMENT_COLLAPSE_LIMIT)) {
        setExpanded(true);
      }
      onAdd(parentId, t);
    },
    [expanded, hiddenCount, totalComments, onAdd]
  );
  const renderNodes = React.useCallback(
    () =>
      tree.map((root) => (
        <CommentThread
          key={root.id}
          node={root}
          onReply={(pid, t) => handleAdd(pid, t)}
          appearance={appearance}
          canReply={canComment}
        />
      )),
    [tree, handleAdd, appearance, canComment]
  );

  return (
    <Stack spacing={1.5}>
      <Paper
        elevation={0}
        variant="outlined"
        sx={{
          borderColor: appearance.border,
          bgcolor: appearance.surface,
          p: 1.5,
          maxHeight: shouldScroll ? 440 : "none",
          overflowY: shouldScroll ? "auto" : "visible",
        }}
      >
        {tree.length ? (
          <Stack spacing={1}>{renderNodes()}</Stack>
        ) : (
          <Typography variant="body2" sx={{ color: appearance.textSecondary }}>
            Noch keine Kommentare.
          </Typography>
        )}
        {hiddenCount > 0 && !expanded && (
          <Box sx={{ mt: 1 }}>
            <Button size="small" onClick={() => setExpanded(true)}>
              Weitere Kommentare anzeigen ({hiddenCount})
            </Button>
          </Box>
        )}
      </Paper>
      <Paper
        elevation={0}
        variant="outlined"
        sx={{
          borderColor: appearance.border,
          bgcolor: appearance.surface,
          p: 1.5,
        }}
      >
        <Stack direction={{ xs: "column", sm: "row" }} spacing={1} alignItems="flex-start">
          <TextField
            size="small"
            fullWidth
            multiline
            minRows={2}
            placeholder={canComment ? "Neuen Kommentar schreiben…" : "Bitte einloggen, um zu kommentieren."}
            value={text}
            onChange={(e) => setText(e.target.value)}
            disabled={!canComment}
            sx={{
              "& .MuiInputBase-root": {
                bgcolor: inputBg,
                borderRadius: 2,
                borderColor: appearance.border,
              },
            }}
          />
          <Button
            variant="contained"
            onClick={() => {
              if (!canComment) return;
              if (text.trim()) {
                handleAdd(null, text.trim());
                setText("");
              }
            }}
            disabled={!canComment}
            sx={{
              alignSelf: { xs: "stretch", sm: "center" },
              minWidth: 96,
              height: 40,
            }}
          >
            Posten
          </Button>
        </Stack>
        {!canComment && disabledHelper && (
          <Typography variant="caption" sx={{ color: appearance.textSecondary, mt: 0.75, display: "block" }}>
            {disabledHelper}
          </Typography>
        )}
      </Paper>
    </Stack>
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
  canComment,
  commentDisabledReason,
  currentAuthor,
  programChipSx,
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
  canComment: boolean;
  commentDisabledReason?: string;
  currentAuthor: string;
  programChipSx: SxProps;
}) {
  const theme = useTheme();
  const isDark = theme.palette.mode === "dark";
  const netVotes = post.votes + (vote as number);
  const [menuEl, setMenuEl] = React.useState<null | HTMLElement>(null);
  const openMenu = Boolean(menuEl);
  const handleOpenDetail = () => onOpenDetail(post);
  const handleCardClick = (event: React.MouseEvent) => {
    const target = event.target as HTMLElement;
    if (target.closest("button, a, input, textarea, select, [role='button'], .MuiInputBase-root")) {
      return;
    }
    handleOpenDetail();
  };

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
      onClick={handleCardClick}
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
              onClick={(e) => {
                e.stopPropagation();
                onVote(post.id, vote === 1 ? 0 : 1);
              }}
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
              onClick={(e) => {
                e.stopPropagation();
                onVote(post.id, vote === -1 ? 0 : -1);
              }}
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
                <MoreVertIcon fontSize="small" />
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
                <Chip
                  key={t}
                  label={t}
                  size="small"
                  variant="outlined"
                  sx={{
                    bgcolor: isDark ? alpha(theme.palette.primary.light, 0.12) : alpha(theme.palette.primary.main, 0.08),
                    borderColor: isDark ? alpha(theme.palette.primary.light, 0.3) : alpha(theme.palette.primary.main, 0.25),
                    color: theme.palette.primary.main,
                  }}
                />
              ))}
              <Divider flexItem orientation="vertical" sx={{ mx: 1 }} />
              {post.programs.map((p) => (
                <Chip
                  key={p}
                  label={PROGRAM_META_MAP[p]?.shortLabel ?? p}
                  size="small"
                  variant="outlined"
                  sx={programChipSx}
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
              canComment={canComment}
              disabledHelper={commentDisabledReason}
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
        {currentAuthor && post.author === currentAuthor && (
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
  const { user } = useAuth();
  const theme = useTheme();
  const isDark = theme.palette.mode === "dark";
  const activeUserName = user?.name ?? user?.email ?? "";
  const canComment = Boolean(user);
  // Laden + Migration (program ? programs)
  const [posts, setPosts] = React.useState<Post[]>(() => {
    try {
      const raw = typeof window !== "undefined" ? localStorage.getItem(LS_KEY) : null;
      const parsed = raw ? (JSON.parse(raw) as unknown) : SEED_POSTS;
      const data = Array.isArray(parsed) ? parsed : SEED_POSTS;
      type StoredPost = Partial<Post> & { program?: unknown; programs?: unknown };
      return (data as StoredPost[]).map((p) => {
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
      const parsed = raw ? (JSON.parse(raw) as unknown) : null;
      return parsed && typeof parsed === "object" ? (parsed as Record<string, Vote>) : {};
    } catch {
      return {};
    }
  });

  const palette = React.useMemo(
    () => ({
      background: theme.palette.background.default,
      surface: theme.palette.background.paper,
      card: theme.palette.background.paper,
      border: theme.palette.divider,
      textSecondary: theme.palette.text.secondary,
    }),
    [theme]
  );
  const commentAppearance = React.useMemo<CommentAppearance>(
    () => ({
      surface: isDark ? alpha(theme.palette.common.white, 0.18) : theme.palette.grey[50],
      border: isDark ? alpha(theme.palette.common.white, 0.45) : alpha(theme.palette.common.black, 0.12),
      accent: theme.palette.primary.main,
      textSecondary: isDark ? alpha(theme.palette.common.white, 0.8) : theme.palette.text.secondary,
    }),
    [isDark, theme]
  );
  const programChipSx = React.useMemo(
    () =>
      isDark
        ? {
            bgcolor: alpha(theme.palette.secondary.light, 0.2),
            borderColor: alpha(theme.palette.secondary.light, 0.5),
            color: theme.palette.secondary.contrastText,
          }
        : {
            bgcolor: alpha(theme.palette.secondary.main, 0.1),
            borderColor: alpha(theme.palette.secondary.main, 0.4),
            color: theme.palette.secondary.main,
          },
    [isDark, theme]
  );
  const inputStyles = React.useMemo(
    () => ({
      "& .MuiInputBase-root": {
        bgcolor: isDark ? alpha(theme.palette.common.white, 0.08) : theme.palette.common.white,
      },
    }),
    [isDark, theme]
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
  const [page, setPage] = React.useState(1);
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

    React.useEffect(() => {
      setPage(1);
    }, [q, sort, activeProgramFilters, allProgramsOnly]);

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
  
    const pageCount = React.useMemo(
      () => Math.max(1, Math.ceil(filtered.length / POSTS_PER_PAGE)),
      [filtered.length]
    );

    React.useEffect(() => {
      setPage((prev) => Math.min(prev, pageCount));
    }, [pageCount]);

    const paginatedPosts = React.useMemo(() => {
      const startIndex = (page - 1) * POSTS_PER_PAGE;
      return filtered.slice(startIndex, startIndex + POSTS_PER_PAGE);
    }, [filtered, page]);
  
    const handleVote = (id: string, v: Vote) => setVotes((prev) => ({ ...prev, [id]: v }));

  // Kommentare
  const handleAddComment = (postId: string, parentId: string | null, text: string) => {
    if (!activeUserName) return;
    const newComment: Comment = {
      id: uuid(),
      author: activeUserName,
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

  const handleOpenCreate = () => {
    setOpen(true);
  };

  const createPost = () => {
    if (!title.trim() || noProgramSelected) return;
    const authorName = activeUserName || "Gast";
    const newPost: Post = {
      id: uuid(),
      title: title.trim(),
      body: body.trim(),
      tags: tagsInput.split(",").map((t) => t.trim()).filter(Boolean),
      author: authorName,
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
  const [reportNote, setReportNote] = React.useState("");
  const openReport = (id: string) => setReportFor(id);
  const closeReport = () => {
    setReportFor(null);
    setReportNote("");
  };
  const sendReport = () => {
    // hier würdest du an ein Backend schicken - wir loggen nur:
    console.log("Report:", { postId: reportFor, reason: reportReason, note: reportNote });
    closeReport();
  };

  const createButton = (
    <Tooltip
      title="Auch als Gast nutzbar."
      disableHoverListener={false}
    >
      <span>
        <Button
          startIcon={<AddIcon fontSize="small" />}
          variant="contained"
          size="small"
          onClick={handleOpenCreate}
          sx={{ borderRadius: 999, px: 2 }}
        >
          Beitrag erstellen
        </Button>
      </span>
    </Tooltip>
  );

  /* UI */
  return (
    <Sidebar user={user} title="Forum" headerActions={createButton}>
      <Box
        sx={{
          minHeight: "100vh",
          bgcolor: palette.background,
          "--card-bg": palette.card,
          "--card-border": palette.border,
          color: theme.palette.text.primary,
        }}
      >
        <Container maxWidth="lg" sx={{ py: 3 }}>
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
            boxShadow: isDark ? "0px 8px 24px rgba(0,0,0,0.35)" : "0px 8px 24px rgba(15,110,46,0.08)",
            color: theme.palette.text.primary,
          }}
        >
          <Stack spacing={2}>
            <Stack direction={{ xs: "column", md: "row" }} spacing={2} alignItems={{ xs: "stretch", md: "center" }}>
              <TextField
                placeholder="Suche in Titel, Text oder Tags…"
                value={q}
                onChange={(e) => setQ(e.target.value)}
                fullWidth
                sx={inputStyles}
                slotProps={{
                  input: {
                    startAdornment: (
                      <InputAdornment position="start">
                        <SearchIcon fontSize="small" />
                      </InputAdornment>
                    ),
                  }
                }}
              />
              <Select
                size="small"
                value={sort}
                onChange={(e) => setSort(e.target.value as "new" | "votes")}
                sx={{
                  width: { xs: "100%", md: 220 },
                  "& .MuiOutlinedInput-root": {
                    bgcolor: isDark ? alpha(theme.palette.common.white, 0.08) : theme.palette.common.white,
                  },
                }}
              >
                <MenuItem value="new">Neueste zuerst</MenuItem>
                <MenuItem value="votes">Beste (Votes)</MenuItem>
              </Select>
              <Button
                startIcon={<FilterListIcon />}
                variant={activeFiltersCount ? "contained" : "outlined"}
                color="primary"
                onClick={() => setFiltersOpen(true)}
                size="small"
                sx={{
                  width: { xs: "100%", md: 220 },
                  alignSelf: { xs: "stretch", md: "center" },
                  height: 40,
                  borderColor: isDark ? alpha(theme.palette.common.white, 0.35) : undefined,
                  bgcolor: !activeFiltersCount && isDark ? alpha(theme.palette.common.white, 0.06) : undefined,
                  color: !activeFiltersCount && isDark ? theme.palette.common.white : undefined,
                  "&:hover": {
                    bgcolor: activeFiltersCount
                      ? undefined
                      : isDark
                      ? alpha(theme.palette.common.white, 0.1)
                      : alpha(theme.palette.primary.main, 0.04),
                    borderColor: isDark ? alpha(theme.palette.common.white, 0.5) : undefined,
                  }
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
          {paginatedPosts.map((p) => (
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
              canComment={canComment}
              commentDisabledReason="Bitte einloggen, um zu kommentieren."
              currentAuthor={activeUserName}
              programChipSx={programChipSx}
            />
          ))}
          {filtered.length === 0 && (
            <Stack spacing={1.5}>
              <Typography variant="body2" sx={{ color: palette.textSecondary }}>
                Keine Treffer. Suchbegriff oder Filter anpassen.
              </Typography>
            </Stack>
          )}
        </Stack>
        {filtered.length > POSTS_PER_PAGE && (
          <Box sx={{ display: "flex", justifyContent: "center", mt: 1 }}>
            <Pagination
              count={pageCount}
              page={page}
              onChange={(_, value) => setPage(value)}
              color="primary"
              showFirstButton
              showLastButton
            />
          </Box>
        )}
      </Container>
      <Drawer
        anchor="right"
        open={filtersOpen}
        onClose={() => setFiltersOpen(false)}
        slotProps={{
          paper: { sx: { bgcolor: palette.surface, color: theme.palette.text.primary } }
        }}
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
                    <Grid component="div" key={meta.id} size={6}>
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
                        sx={[{
                          p: 1.5,
                          border: 2,
                          cursor: "pointer",
                          transition: "all .2s ease",
                          minHeight: 96,
                          display: "flex",
                          flexDirection: "column",
                          justifyContent: "space-between",
                          "&:hover": {
                            borderColor: "secondary.main",
                            boxShadow: 2,
                          }
                        }, active ? {
                          borderColor: "secondary.main"
                        } : {
                          borderColor: palette.border
                        }, active ? {
                          bgcolor: "rgba(15,110,46,0.12)"
                        } : {
                          bgcolor: palette.surface
                        }]}
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
                      variant="outlined"
                      sx={programChipSx}
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
                  canComment={canComment}
                  disabledHelper="Bitte einloggen, um zu kommentieren."
                />

                <Divider />

                <Stack direction={{ xs: "column", md: "row" }} spacing={1} alignItems="stretch">
                  <TextField
                    label="Link teilen"
                    value={shareUrl}
                    fullWidth
                    sx={inputStyles}
                    slotProps={{
                      input: { readOnly: true }
                    }}
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
      <Dialog open={!!reportFor} onClose={closeReport} maxWidth="xs" fullWidth>
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
          <TextField
            label="Zusätzliche Hinweise (optional)"
            multiline
            minRows={2}
            value={reportNote}
            onChange={(e) => setReportNote(e.target.value)}
            fullWidth
            sx={{ mt: 2 }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={closeReport}>Abbrechen</Button>
          <Button variant="contained" onClick={sendReport}>Melden</Button>
        </DialogActions>
      </Dialog>
    </Box>
    </Sidebar>
  );
}




