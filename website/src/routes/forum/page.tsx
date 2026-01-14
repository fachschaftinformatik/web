import * as React from "react";
import {
  Box, Stack, Paper, Typography, TextField, Button,
  IconButton, Chip, Dialog, DialogTitle, DialogContent,
  DialogActions, InputAdornment, Menu, MenuItem,
  Select, Tooltip, Pagination, Autocomplete, Avatar
} from "@mui/material";
import { alpha, useTheme } from "@mui/material/styles";
import SearchIcon from "@mui/icons-material/Search";
import AddIcon from "@mui/icons-material/Add";
import ThumbUpOutlined from "@mui/icons-material/ThumbUpOutlined";
import ThumbDownOutlined from "@mui/icons-material/ThumbDownOutlined";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import PushPinOutlinedIcon from "@mui/icons-material/PushPinOutlined";
import ListItemIcon from "@mui/material/ListItemIcon";
import { useNavigate, Link } from "react-router-dom";
import ListItemText from "@mui/material/ListItemText";
import ReportOutlinedIcon from "@mui/icons-material/ReportOutlined";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import EditIcon from "@mui/icons-material/Edit";
import { Sidebar } from "@components/layout";
import { useAuth } from "@lib/auth";
import {
  getForumPosts,
  postForumPostsByIdVote,
  deleteForumPostsById,
  putForumPostsById,
  getAuthCsrf
} from "@lib/api";
import type {
  AuthPostResponse as ApiPost,
  AuthUserResponse as User
} from "@lib/api";

import {
  PROGRAM_CATALOG, PROGRAM_META_MAP, isoToShort,
  Post, Vote, Program, POSTS_PER_PAGE
} from "./components";
function PostItem({
  post,
  onVote,
  vote,
  onDelete,

  onTogglePin,
  user,
  isAdmin,
  onReport,
}: {
  post: Post;
  onVote: (id: string, v: Vote) => void;
  vote: Vote;
  onDelete: (id: string) => void;
  onTogglePin: (id: string) => void;
  user: User | null;
  isAdmin: boolean;
  onReport: (id: string) => void;
}) {
  const navigate = useNavigate();
  const theme = useTheme();
  const isDark = theme.palette.mode === "dark";
  const netVotes = (Number(post.votes) || 0);
  const canDelete = isAdmin || (user?.id === post.author_id);
  const [menuEl, setMenuEl] = React.useState<null | HTMLElement>(null);
  const openMenu = Boolean(menuEl);

  const handleOpenMenu = (e: React.MouseEvent<HTMLElement>) => {
    e.stopPropagation();
    setMenuEl(e.currentTarget);
  };
  const handleCloseMenu = () => setMenuEl(null);

  return (
    <Paper
      onClick={() => navigate(`/forum/${post.id}`)}
      elevation={0}
      sx={{
        p: 2,
        borderRadius: 3,
        cursor: "pointer",
        border: "1px solid",
        borderColor: theme.palette.divider,
        transition: "all 0.2s ease-in-out",
        bgcolor: "var(--card-bg)",
        "&:hover": {
          transform: "translateY(-2px)",
          boxShadow: isDark ? "0 8px 16px rgba(0,0,0,0.4)" : "0 8px 16px rgba(15,110,46,0.08)",
          borderColor: theme.palette.primary.main,
          bgcolor: isDark ? alpha(theme.palette.primary.main, 0.04) : alpha(theme.palette.primary.main, 0.02),
        },
      }}
    >
      <Stack direction="row" spacing={2} alignItems="flex-start">
        <Stack alignItems="center" spacing={0.5} sx={{ mt: 0.5 }}>
          <IconButton
            size="small"
            onClick={(e) => {
              e.stopPropagation();
              onVote(post.id!, vote === 1 ? 0 : 1);
            }}
            color={vote === 1 ? "primary" : "default"}
          >
            <ThumbUpOutlined fontSize="small" />
          </IconButton>
          <Typography variant="subtitle2" sx={{ fontWeight: 700 }}>
            {netVotes}
          </Typography>
          <IconButton
            size="small"
            onClick={(e) => {
              e.stopPropagation();
              onVote(post.id!, vote === -1 ? 0 : -1);
            }}
            color={vote === -1 ? "primary" : "default"}
          >
            <ThumbDownOutlined fontSize="small" />
          </IconButton>
        </Stack>

        <Stack spacing={1} sx={{ flex: 1 }}>
          <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
            <Stack direction="row" spacing={1} alignItems="center">
              <Typography variant="h6" sx={{ fontWeight: 700, lineHeight: 1.3 }}>
                {post.title}
              </Typography>
              {post.pinned === 1 && (
                <PushPinOutlinedIcon fontSize="small" sx={{ color: theme.palette.secondary.main }} />
              )}
            </Stack>
            <IconButton size="small" onClick={handleOpenMenu}>
              <MoreVertIcon fontSize="small" />
            </IconButton>
            <Menu
              anchorEl={menuEl}
              open={openMenu}
              onClose={handleCloseMenu}
              onClick={(e) => e.stopPropagation()}
            >

              {user?.id === post.author_id && (
                <MenuItem
                  onClick={(e) => {
                    e.stopPropagation();
                    handleCloseMenu();
                    navigate(`/forum/${post.id}/edit`);
                  }}
                >
                  <ListItemIcon>
                    <EditIcon fontSize="small" />
                  </ListItemIcon>
                  <ListItemText>Bearbeiten</ListItemText>
                </MenuItem>
              )}
              {isAdmin && (
                <MenuItem
                  disabled
                  onClick={(e) => {
                    e.stopPropagation();
                    handleCloseMenu();
                    onTogglePin(post.id!);
                  }}
                >
                  <ListItemIcon>
                    <PushPinOutlinedIcon fontSize="small" />
                  </ListItemIcon>
                  <ListItemText>{post.pinned === 1 ? "Entpinnen" : "Anpinnen"}</ListItemText>
                </MenuItem>
              )}
              <MenuItem
                disabled
                onClick={(e) => {
                  e.stopPropagation();
                  handleCloseMenu();
                  onReport(post.id!);
                }}
              >
                <ListItemIcon>
                  <ReportOutlinedIcon fontSize="small" />
                </ListItemIcon>
                <ListItemText>Melden</ListItemText>
              </MenuItem>
              {canDelete && (
                <MenuItem
                  onClick={(e) => {
                    e.stopPropagation();
                    handleCloseMenu();
                    onDelete(post.id!);
                  }}
                  sx={{ color: "error.main" }}
                >
                  <ListItemIcon>
                    <DeleteOutlineIcon fontSize="small" color="error" />
                  </ListItemIcon>
                  <ListItemText>Löschen</ListItemText>
                </MenuItem>
              )}
            </Menu>
          </Stack>

          <Typography
            variant="body2"
            sx={{
              color: theme.palette.text.secondary,
              display: "-webkit-box",
              WebkitLineClamp: 2,
              WebkitBoxOrient: "vertical",
              overflow: "hidden",
            }}
          >
            {post.body}
          </Typography>

          <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap">
            <Avatar
              src={post.author_avatar_url || undefined}
              sx={{
                width: 20,
                height: 20,
                fontSize: "0.65rem",
                bgcolor: theme.palette.primary.main,
                fontWeight: "bold"
              }}
            />
            <Typography variant="caption" sx={{ color: theme.palette.text.secondary }}>
              von <Typography component={Link} to={`/user/${post.author_id}`} onClick={(e) => e.stopPropagation()} variant="caption" sx={{ color: 'inherit', fontWeight: 600, textDecoration: 'none', '&:hover': { textDecoration: 'underline', color: 'primary.main' } }}>{post.author_name || "Anonym"}</Typography> · {isoToShort(post.created_at ?? "")}
              {post.updated_at && post.created_at && post.updated_at !== post.created_at && " (bearbeitet)"}
              {" · "}
              {post.comment_count} {post.comment_count === 1 ? "Kommentar" : "Kommentare"}
            </Typography>
          </Stack>

          <Stack direction="row" spacing={0.5} flexWrap="wrap">
            {post.programs.map((program) => (
              <Chip
                key={program}
                label={PROGRAM_META_MAP[program]?.shortLabel ?? program}
                size="small"
                variant="outlined"
              />
            ))}
            {post.tags.slice(0, 3).map((tag) => (
              <Chip
                key={tag}
                label={tag}
                size="small"
                variant="outlined"
              />
            ))}
          </Stack>
        </Stack>
      </Stack>
    </Paper>
  );
}



export default function ForumPage() {
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";

  const navigate = useNavigate();

  const [posts, setPosts] = React.useState<Post[]>([]);
  const [loading, setLoading] = React.useState(true);

  const fetchPosts = React.useCallback(async () => {
    setLoading(true);
    try {
      const { data } = await getForumPosts({ query: { limit: 100 } });
      if (data) {
        const parsed: Post[] = (data as ApiPost[]).map((p: ApiPost) => {
          let programs: Program[] = [];
          let tags: string[] = [];

          try {
            if (p.programs) {
              const res = JSON.parse(p.programs as string);
              if (Array.isArray(res)) programs = res;
            }
            if (p.tags) {
              const res = JSON.parse(p.tags as string);
              if (Array.isArray(res)) tags = res;
            }
          } catch (e) {
            console.error(p.id, e);
          }
          let links: string[] = [];
          try {
            if (p.links) {
              const res = JSON.parse(p.links as string);
              if (Array.isArray(res)) links = res;
            }
          } catch { /* empty */ }

          return {
            ...p,
            programs,
            tags,
            links,
            comments: []
          } as Post;
        });
        setPosts(parsed);
      }
    } catch (err) {
      console.error("Fetch error:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    fetchPosts();
  }, [fetchPosts]);

  React.useEffect(() => {
    if (typeof window === "undefined") return;
    const params = new URLSearchParams(window.location.search);
    const targetId = params.get("postId") || params.get("post");
    if (targetId) {
      navigate(`/forum/${targetId}`, { replace: true });
    }
  }, [navigate]);




  const [q, setQ] = React.useState("");
  const [sort, setSort] = React.useState<"new" | "votes" | "relevant">("new");
  const [activeProgramFilters, setActiveProgramFilters] = React.useState<Program[]>([]);
  const [page, setPage] = React.useState(1);

  React.useEffect(() => {
    setPage(1);
  }, [q, sort, activeProgramFilters]);

  // Filter Logic
  const filtered = React.useMemo(() => {
    const query = q.trim().toLowerCase();
    const isDefaultView = !query && !activeProgramFilters.length;
    let base = posts;

    if (activeProgramFilters.length) {
      base = base.filter((p) =>
        p.programs.some((program) => activeProgramFilters.includes(program))
      );
    }

    if (query) {
      base = base.filter(
        (p) =>
          (p.title ?? "").toLowerCase().includes(query) ||
          (p.body ?? "").toLowerCase().includes(query) ||
          p.tags.some((t) => t.toLowerCase().includes(query))
      );
    }

    const sortWithPinned = (a: Post, b: Post) => {
      if (isDefaultView && sort !== "votes") {
        const pinDiff = Number(Boolean(b.pinned)) - Number(Boolean(a.pinned));
        if (pinDiff !== 0) return pinDiff;
      }
      if (sort === "votes") {
        return (Number(b.votes)) - (Number(a.votes));
      }
      return +new Date(b.created_at ?? 0) - +new Date(a.created_at ?? 0);
    };

    return [...base].sort(sortWithPinned);
  }, [posts, q, sort, activeProgramFilters]);

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

  const handleVote = async (id: string, newVote: Vote) => {
    if (!user) return;
    try {
      const { data: csrfData } = await getAuthCsrf();
      if (!csrfData?.csrf) return;
      await postForumPostsByIdVote({
        path: { id },
        body: { vote: newVote },
        headers: { "X-CSRF-Token": csrfData.csrf }
      });
      setPosts(prev => prev.map(p => p.id === id ? { ...p, user_vote: newVote, votes: (Number(p.votes)) - (p.user_vote || 0) + newVote } : p));
    } catch (e) {
      console.error(e);
    }
  };

  const togglePin = async (id: string) => {
    if (!isAdmin) return;
    const post = posts.find(p => p.id === id);
    if (!post) return;
    try {
      const { data: csrfData } = await getAuthCsrf();
      if (!csrfData?.csrf) return;
      await putForumPostsById({
        path: { id },
        body: { pinned: post.pinned ? 0 : 1 },
        headers: { "X-CSRF-Token": csrfData.csrf }
      });
      setPosts(prev => prev.map(p => p.id === id ? { ...p, pinned: post.pinned ? 0 : 1 } : p));
    } catch (e) { console.error(e); }
  };

  const handleDelete = async (id: string) => {
    if (!isAdmin && user?.role !== "editor" && user?.id !== posts.find(p => p.id === id)?.author_id) return;
    if (!confirm("Wirklich löschen?")) return;
    try {
      const { data: csrfData } = await getAuthCsrf();
      if (!csrfData?.csrf) return;
      await deleteForumPostsById({
        path: { id },
        headers: { "X-CSRF-Token": csrfData.csrf }
      });
      setPosts(prev => prev.filter(p => p.id !== id));
    } catch (e) { console.error(e); }
  };

  const handleOpenCreate = () => navigate("/forum/create");

  const [reportFor, setReportFor] = React.useState<string | null>(null);
  const [reportReason, setReportReason] = React.useState("Spam / Werbung");
  const [reportNote, setReportNote] = React.useState("");
  const openReport = (id: string) => setReportFor(id);
  const closeReport = () => { setReportFor(null); setReportNote(""); };
  const sendReport = () => {
    console.log("Report", reportFor, reportReason, reportNote);
    // Ideally send to API
    closeReport();
  };

  return (
    <Sidebar user={user} title="Forum" maxWidth="lg">
      <Box>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 4 }}>
          <Box>
            <Typography variant="h4" fontWeight={700} gutterBottom>Forum</Typography>
            <Typography variant="body1" color="text.secondary">Diskutiere mit deinen Kommilitonen über Studium, Campus und Events.</Typography>
          </Box>
        </Box>

        {loading ? (<Box display="flex" justifyContent="center" py={8}><Typography>Lädt...</Typography></Box>) : (
          <>
            <Paper elevation={0} sx={{ p: 3, mb: 3, borderRadius: 3, bgcolor: "background.paper", border: "1px solid", borderColor: "divider" }}>
              <Stack spacing={2}>
                <Stack direction="row" spacing={2}>
                  <TextField fullWidth placeholder="Suche in Titel, Text oder Tags…" value={q} onChange={e => setQ(e.target.value)} size="small" InputProps={{ startAdornment: <InputAdornment position="start"><SearchIcon color="primary" /></InputAdornment>, sx: { borderRadius: 2 } }} />
                  <Tooltip title="Beitrag erstellen">
                    <IconButton onClick={handleOpenCreate} sx={{ bgcolor: 'primary.main', color: 'primary.contrastText', borderRadius: 2, width: 40, height: 40, '&:hover': { bgcolor: 'primary.dark' } }}><AddIcon /></IconButton>
                  </Tooltip>
                </Stack>
                <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', width: '100%', alignItems: 'center' }}>
                  <Autocomplete
                    multiple
                    options={PROGRAM_CATALOG}
                    getOptionLabel={o => o.label}
                    value={activeProgramFilters.map(id => PROGRAM_META_MAP[id]).filter(Boolean)}
                    isOptionEqualToValue={(o, v) => o.id === v.id}
                    onChange={(_, n) => setActiveProgramFilters(n.map(v => v.id))}
                    renderInput={p => <TextField {...p} label="Studiengang" size="small" placeholder="Wählen..." />}
                    size="small"
                    sx={{ flex: 2, minWidth: 300 }}
                  />
                  <TextField select label="Filter" value={sort} onChange={e => setSort(e.target.value as "new" | "votes" | "relevant")} size="small" sx={{ flex: 1, minWidth: 150 }}>
                    <MenuItem value="relevant">Relevant</MenuItem>
                    <MenuItem value="new">Neuste</MenuItem>
                    <MenuItem value="votes">Top</MenuItem>
                  </TextField>
                </Box>
              </Stack>
            </Paper>

            <Stack spacing={2}>
              {paginatedPosts.map(p => (
                <PostItem
                  key={p.id}
                  post={p}
                  vote={(p.user_vote ?? 0) as Vote}
                  onVote={handleVote}
                  onDelete={handleDelete}
                  onReport={openReport}
                  onTogglePin={togglePin}
                  user={user}
                  isAdmin={isAdmin}
                />
              ))}
              {filtered.length === 0 && <Stack spacing={1.5}><Typography color="text.secondary">Keine Treffer. Suchbegriff oder Filter anpassen.</Typography></Stack>}
            </Stack>

            {filtered.length > POSTS_PER_PAGE && (
              <Box display="flex" justifyContent="center" mt={1}>
                <Pagination count={pageCount} page={page} onChange={(_, v) => setPage(v)} color="primary" showFirstButton showLastButton />
              </Box>
            )}
          </>
        )}

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