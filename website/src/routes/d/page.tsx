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
import ShareIcon from "@mui/icons-material/Share";
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
  getDiscussions,
  getPrograms,
  postDiscussionsByPostIdVote,
  deleteDiscussionsByPostId,
  putDiscussionsByPostId,
} from "@lib/api";
import type {
  DtoDiscussionPostResponse as ApiPost,
  DtoUserResponse as User,
  DtoProgramResponse
} from "@lib/api";
import { getAvatarUrl } from "@lib/images";

import {
  isoToShort,
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
  onShare,
}: {
  post: Post;
  onVote: (id: string, v: Vote) => void;
  vote: Vote;
  onDelete: (id: string) => void;
  onTogglePin: (id: string) => void;
  user: User | null;
  isAdmin: boolean;
  onReport: (id: string) => void;
  onShare: (id: string) => void;
}) {
  const navigate = useNavigate();
  const theme = useTheme();
  const isDark = theme.palette.mode === "dark";
  const netVotes = (Number(post.votes) || 0);
  const canDelete = isAdmin || (String(user?.id) === String(post.user_id));
  const [menuEl, setMenuEl] = React.useState<null | HTMLElement>(null);
  const openMenu = Boolean(menuEl);

  const handleOpenMenu = (e: React.MouseEvent<HTMLElement>) => {
    e.stopPropagation();
    setMenuEl(e.currentTarget);
  };
  const handleCloseMenu = () => setMenuEl(null);

  const handleShare = (e: React.MouseEvent) => {
    e.stopPropagation();
    handleCloseMenu();
    onShare(String(post.id!));
  };

  return (
    <Paper
      onClick={() => navigate(`/d/${post.id}`)}
      elevation={0}
      sx={{
        p: 2,
        borderRadius: 3,
        cursor: "pointer",
        border: "1px solid",
        borderColor: theme.palette.divider,
        transition: "all 0.2s ease-in-out",
        bgcolor: post.type === "news"
          ? (isDark ? alpha(theme.palette.primary.main, 0.08) : alpha(theme.palette.primary.main, 0.04))
          : "var(--card-bg)",
        "&:hover": {
          transform: "translateY(-2px)",
          boxShadow: isDark ? "0 8px 16px rgba(0,0,0,0.4)" : "0 8px 16px rgba(15,110,46,0.08)",
          borderColor: theme.palette.primary.main,
          bgcolor: post.type === "news"
            ? (isDark ? alpha(theme.palette.primary.main, 0.12) : alpha(theme.palette.primary.main, 0.06))
            : (isDark ? alpha(theme.palette.primary.main, 0.04) : alpha(theme.palette.primary.main, 0.02)),
        },
      }}
    >
      <Stack direction="row" spacing={2} alignItems="flex-start">
        <Stack alignItems="center" spacing={0.5} sx={{ mt: 0.5 }}>
          <IconButton
            size="small"
            onClick={(e) => {
              e.stopPropagation();
              onVote(String(post.id!), vote === 1 ? 0 : 1);
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
              onVote(String(post.id!), vote === -1 ? 0 : -1);
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
                 <PushPinOutlinedIcon fontSize="small" sx={{ color: 'primary.main' }} />
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
              disableScrollLock
              PaperProps={{
                elevation: 3,
                sx: {
                  borderRadius: 2,
                  mt: 0.5,
                  minWidth: 160,
                  '& .MuiMenuItem-root': {
                    px: 2,
                    py: 1,
                    display: 'flex',
                    alignItems: 'center',
                    gap: 1.5
                  }
                }
              }}
            >

              {String(user?.id) === String(post.user_id) && (
                <MenuItem
                  onClick={(e) => {
                    e.stopPropagation();
                    handleCloseMenu();
                    navigate(`/d/${post.id}/edit`);
                  }}
                >
                  <EditIcon sx={{ fontSize: 18 }} />
                  <Typography variant="body2" sx={{ fontWeight: 500 }}>Bearbeiten</Typography>
                </MenuItem>
              )}
              {(isAdmin || user?.role === "editor") && (
                <MenuItem
                  onClick={(e) => {
                    e.stopPropagation();
                    handleCloseMenu();
                    onTogglePin(String(post.id!));
                  }}
                >
                  <PushPinOutlinedIcon sx={{ fontSize: 18 }} />
                  <Typography variant="body2" sx={{ fontWeight: 500 }}>{post.pinned === 1 ? "Entpinnen" : "Anpinnen"}</Typography>
                </MenuItem>
              )}
              <MenuItem
                onClick={handleShare}
              >
                <ShareIcon sx={{ fontSize: 18 }} />
                <Typography variant="body2" sx={{ fontWeight: 500 }}>Teilen</Typography>
              </MenuItem>
              <MenuItem
                disabled
                onClick={(e) => {
                  e.stopPropagation();
                  handleCloseMenu();
                  onReport(String(post.id!));
                }}
              >
                <ReportOutlinedIcon sx={{ fontSize: 18 }} />
                <Typography variant="body2" sx={{ fontWeight: 500 }}>Melden</Typography>
              </MenuItem>
              {canDelete && (
                <MenuItem
                  onClick={(e) => {
                    e.stopPropagation();
                    handleCloseMenu();
                    onDelete(String(post.id!));
                  }}
                  sx={{ color: "error.main" }}
                >
                  <DeleteOutlineIcon sx={{ fontSize: 18, color: "inherit" }} />
                  <Typography variant="body2" sx={{ fontWeight: 500 }}>Löschen</Typography>
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
              src={getAvatarUrl(post.user_avatar_url)}
              sx={{
                width: 20,
                height: 20,
                fontSize: "0.65rem",
                bgcolor: theme.palette.primary.main,
                fontWeight: "bold"
              }}
            >
              {post.user_name ? post.user_name[0].toUpperCase() : "A"}
            </Avatar>
            <Typography variant="caption" sx={{ color: theme.palette.text.secondary }}>
              von <Typography component={Link} to={`/u/${post.user_id}`} onClick={(e) => e.stopPropagation()} variant="caption" sx={{ color: 'inherit', fontWeight: 600, textDecoration: 'none', '&:hover': { textDecoration: 'underline', color: 'primary.main' } }}>{post.user_name || "Anonym"}</Typography> · {isoToShort(post.created_at ?? "")}
              {post.updated_at && post.created_at && post.updated_at !== post.created_at && " (bearbeitet)"}
              {" · "}
              {post.comment_count} {post.comment_count === 1 ? "Kommentar" : "Kommentare"}
            </Typography>
          </Stack>

          <Stack direction="row" spacing={0.5} flexWrap="wrap">
            {(post.programs || []).map((program) => (
              <Chip
                key={String(program.id)}
                label={program.name}
                size="small"
                variant="outlined"
              />
            ))}
            {(post.tags || []).slice(0, 3).map((tag) => (
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



export default function DiscussionsPage() {
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";
  const navigate = useNavigate();

  const [posts, setPosts] = React.useState<Post[]>([]);
  const [totalCount, setTotalCount] = React.useState(0);
  const [loading, setLoading] = React.useState(true);

  const [q, setQ] = React.useState("");
  const [sort, setSort] = React.useState<"new" | "votes" | "relevant">("new");
  const [activeProgramFilters, setActiveProgramFilters] = React.useState<Program[]>([]);
  const [page, setPage] = React.useState(1);

  const [apiPrograms, setApiPrograms] = React.useState<DtoProgramResponse[]>([]);

  const fetchPosts = React.useCallback(async () => {
    setLoading(true);
    try {
      const offset = (page - 1) * POSTS_PER_PAGE;
      const { data, response } = await getDiscussions({
        query: {
          limit: POSTS_PER_PAGE,
          offset,
          query: q.trim() || undefined,
          sort: sort === "votes" ? "votes" : undefined,
          program_id: activeProgramFilters.length > 0 ? apiPrograms.find(p => p.name === activeProgramFilters[0])?.id : undefined
        }
      });

      if (data) {
        const parsed: Post[] = (data as ApiPost[]).map((p: ApiPost) => {
          return {
            ...p,
            comments: []
          } as Post;
        });
        setPosts(parsed);

        const total = parseInt(response.headers.get("X-Total-Count") || "0", 10);
        setTotalCount(total);
      }
    } catch (err) {
      console.error("Fetch error:", err);
    } finally {
      setLoading(false);
    }
  }, [page, q, sort, apiPrograms, activeProgramFilters]);

  React.useEffect(() => {
    const fetchPrograms = async () => {
      try {
        const { data } = await getPrograms();
        if (data) setApiPrograms(data);
      } catch (err) {
        console.error("Failed to fetch programs:", err);
      }
    };
    fetchPrograms();
  }, []);

  React.useEffect(() => {
    fetchPosts();
  }, [fetchPosts]);

  React.useEffect(() => {
    if (typeof window === "undefined") return;
    const params = new URLSearchParams(window.location.search);
    const targetId = params.get("postId") || params.get("post");
    if (targetId) {
      navigate(`/d/${targetId}`, { replace: true });
    }
  }, [navigate]);

  React.useEffect(() => {
    setPage(1);
  }, [q, sort, activeProgramFilters]);

  const pageCount = Math.ceil(totalCount / POSTS_PER_PAGE);

  const handleVote = async (id: string, newVote: Vote) => {
    if (!user) return;
    try {
      await postDiscussionsByPostIdVote({
        path: { postId: id },
        body: { vote: newVote },
      });
      setPosts(prev => prev.map(p => String(p.id) === id ? { ...p, user_vote: newVote, votes: (Number(p.votes)) - (p.user_vote || 0) + newVote } : p));
    } catch (e) {
      console.error(e);
    }
  };

  const togglePin = async (id: string) => {
    if (!isAdmin && user?.role !== "editor") return;
    const post = posts.find(p => String(p.id) === id);
    if (!post) return;
    try {
      await putDiscussionsByPostId({
        path: { postId: id },
        body: { pinned: !post.pinned } as unknown as { pinned: boolean }, // pinned is boolean in request but int in response
      });
      setPosts(prev => prev.map(p => String(p.id) === id ? { ...p, pinned: post.pinned ? 0 : 1 } : p));
    } catch (e) { console.error(e); }
  };

  const handleDelete = async (id: string) => {
    if (!isAdmin && user?.role !== "editor" && String(user?.id) !== String(posts.find(p => String(p.id) === id)?.user_id)) return;
    if (!confirm("Wirklich löschen?")) return;
    try {
      await deleteDiscussionsByPostId({
        path: { postId: id },
      });
      setPosts(prev => prev.filter(p => String(p.id) !== id));
    } catch (e) { console.error(e); }
  };

  const handleOpenCreate = () => navigate("/d/new");

  const handleShare = (id: string) => {
    const url = `${window.location.origin}/d/${id}`;
    navigator.clipboard.writeText(url);
  };

  const [reportFor, setReportFor] = React.useState<string | null>(null);
  const [reportReason, setReportReason] = React.useState("Spam / Werbung");
  const [reportNote, setReportNote] = React.useState("");
  const openReport = (id: string) => setReportFor(id);
  const closeReport = () => { setReportFor(null); setReportNote(""); };
  const sendReport = () => {
    closeReport();
  };

  return (
    <Sidebar user={user} title="Diskussionen" maxWidth="lg">
      <Box>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 4 }}>
          <Box>
            <Typography variant="h4" fontWeight={700} gutterBottom>Diskussionen</Typography>
            <Typography variant="body1" color="text.secondary">Diskutiere mit deinen Kommilitonen über Studium, Campus und Events.</Typography>
          </Box>
        </Box>

        {loading && posts.length === 0 ? (
          <Box display="flex" justifyContent="center" py={8}><Typography>Lädt...</Typography></Box>
        ) : (
          <>
            <Paper elevation={0} sx={{ p: 3, mb: 3, borderRadius: 3, bgcolor: "background.paper", border: "1px solid", borderColor: "divider" }}>
              <Stack spacing={2}>
                <Stack direction="row" spacing={2}>
                  <TextField
                    fullWidth
                    placeholder="Suche in Titel, Text oder Tags…"
                    value={q}
                    onChange={e => setQ(e.target.value)}
                    size="small"
                    InputProps={{
                      startAdornment: <InputAdornment position="start"><SearchIcon color="primary" /></InputAdornment>,
                      sx: { borderRadius: 2 }
                    }}
                  />
                  <Tooltip title="Beitrag erstellen">
                    <IconButton
                      onClick={handleOpenCreate}
                      sx={{
                        bgcolor: 'primary.main',
                        color: 'primary.contrastText',
                        borderRadius: 2,
                        width: 40,
                        height: 40,
                        '&:hover': { bgcolor: 'primary.dark' }
                      }}
                    >
                      <AddIcon />
                    </IconButton>
                  </Tooltip>
                </Stack>
                <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', width: '100%', alignItems: 'center' }}>
                  <Autocomplete
                    multiple
                    options={apiPrograms}
                    getOptionLabel={o => o.name || ""}
                    value={apiPrograms.filter(p => activeProgramFilters.includes(p.name || ""))}
                    isOptionEqualToValue={(o, v) => String(o.id) === String(v.id)}
                    onChange={(_, n) => setActiveProgramFilters(n.map(v => v.name || ""))}
                    renderInput={p => <TextField {...p} label="Studiengang" size="small" placeholder="Wählen..." />}
                    size="small"
                    sx={{ flex: 2, minWidth: 300 }}
                  />
                  <TextField
                    select
                    label="Filter"
                    value={sort}
                    onChange={e => setSort(e.target.value as "new" | "votes" | "relevant")}
                    size="small"
                    sx={{ flex: 1, minWidth: 150 }}
                  >
                    <MenuItem value="relevant">Relevant</MenuItem>
                    <MenuItem value="new">Neuste</MenuItem>
                    <MenuItem value="votes">Top</MenuItem>
                  </TextField>
                </Box>
              </Stack>
            </Paper>

            <Stack spacing={2} sx={{ opacity: loading ? 0.6 : 1, transition: 'opacity 0.2s' }}>
              {posts.map(p => (
                <PostItem
                  key={String(p.id)}
                  post={p}
                  vote={(p.user_vote ?? 0) as Vote}
                  onVote={handleVote}
                  onDelete={handleDelete}
                  onReport={openReport}
                   onTogglePin={togglePin}
                  user={user}
                  isAdmin={isAdmin}
                  onShare={handleShare}
                />
              ))}
              {posts.length === 0 && !loading && (
                <Stack spacing={1.5}>
                  <Typography color="text.secondary">Keine Treffer. Suchbegriff oder Filter anpassen.</Typography>
                </Stack>
              )}
            </Stack>

            {pageCount > 1 && (
              <Box display="flex" justifyContent="center" mt={4}>
                <Pagination
                  count={pageCount}
                  page={page}
                  onChange={(_, v) => setPage(v)}
                  color="primary"
                  shape="rounded"
                  variant="outlined"
                  disabled={loading}
                />
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
